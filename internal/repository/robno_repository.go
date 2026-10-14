package repository

import (
	"context"
	"database/sql"
	"fmt"

	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// RobnoRepository holds the statements of the robno (goods) module that are not plain CRUD of one
// entity: the entry of the documents and their stavke, the stanja and the lookups of the forms.
// The methods are grouped by table in the robno_*_repository.go files.
//
// Every statement on a table with god/kar is restricted to the period of the session in ctx.
// Methods that take a db.Transaction run in it; a nil transaction runs on the database.
type RobnoRepository struct {
	db db.Database
}

// NewRobnoRepository returns the repository of the robno module.
func NewRobnoRepository(database db.Database) *RobnoRepository {
	return &RobnoRepository{db: database}
}

// BeginTx starts a transaction for the methods that take one.
func (r *RobnoRepository) BeginTx() (db.Transaction, error) {
	return r.db.Beginx()
}

// runner is what a statement needs: the database or a transaction.
type runner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *RobnoRepository) on(tx db.Transaction) runner {
	if tx != nil {
		return tx
	}
	return r.db
}

// inPeriod adds the period of the session (prefix is "" or an alias with a dot, e.g. "d.").
func inPeriod(ctx context.Context, b *querybuilder.Builder, prefix string) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b.Where(prefix+"god = ?", s.God).Where(prefix+"kar = ?", s.Kar)
	return nil
}

// row runs a single-row query and scans it. A missing row is returned as an error that wraps
// sql.ErrNoRows, so callers can test it with errors.Is.
func (r *RobnoRepository) row(ctx context.Context, tx db.Transaction, what string, b *querybuilder.Builder, dest ...any) error {
	q, args, err := b.Build()
	if err != nil {
		return fmt.Errorf("build %s: %w", what, err)
	}
	if err := r.on(tx).QueryRowContext(ctx, q, args...).Scan(dest...); err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	return nil
}

// exec runs a statement and returns the number of affected rows.
func (r *RobnoRepository) exec(ctx context.Context, tx db.Transaction, what string, b *querybuilder.Builder) (int64, error) {
	q, args, err := b.Build()
	if err != nil {
		return 0, fmt.Errorf("build %s: %w", what, err)
	}
	res, err := r.on(tx).ExecContext(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", what, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", what, err)
	}
	return n, nil
}

// execOne runs a statement that must change exactly one row: ErrNotFound when it matched none
// (e.g. an id of another period).
func (r *RobnoRepository) execOne(ctx context.Context, tx db.Transaction, what string, b *querybuilder.Builder) error {
	n, err := r.exec(ctx, tx, what, b)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s: %w", what, ErrNotFound)
	}
	return nil
}

// selectRows runs a query into dest (a pointer to a slice); not in a transaction.
func (r *RobnoRepository) selectRows(ctx context.Context, what string, dest any, b *querybuilder.Builder) error {
	q, args, err := b.Build()
	if err != nil {
		return fmt.Errorf("build %s: %w", what, err)
	}
	if err := r.db.SelectContext(ctx, dest, q, args...); err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	return nil
}

// poreskaStopa is the PDV stopa (rpor.pp) of a poreska oznaka on a datum: the last one from that
// datum, 0 without one. po and datum are static SQL expressions written in code (a column or "?").
func poreskaStopa(po, datum string) string {
	return `coalesce((select x.pp from rpor x where x.po = ` + po + ` and x.datum <= ` + datum + ` order by x.datum desc limit 1), 0)::float8`
}

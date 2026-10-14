package repository

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

const (
	// ActionTypeInsert is used for insert operations
	ActionTypeInsert = "insert"
	// ActionTypeUpdate is used for update operations
	ActionTypeUpdate = "update"
	// ActionTypeDelete is used for delete operations
	ActionTypeDelete = "delete"
)

// BaseRepository is the generic repository. All SQL is generated through
// internal/infrastructure/db/querybuilder.
//
// For entities with god/kar columns, GetByID, GetAll, GetTotalRecords, Update and Delete are
// restricted to the tenant of the session in the context. Use WithoutTenantScope() for the
// few repositories that must work across tenants (and review every such use).
//
// A BaseRepository may be copied by value: it holds no lock, and the entity metadata is a
// read-only pointer shared by all copies.
type BaseRepository[T any] struct {
	DB         db.Database
	TableName  string
	fieldCache map[string]reflect.StructField

	table         *entityTable // set by NewBaseRepository; nil for struct literals
	noTenantScope bool
}

// NewBaseRepository creates a repository. It panics at startup when the table name or the
// entity's db tags are not valid SQL identifiers, so mistakes show up immediately.
func NewBaseRepository[T any](database db.Database, tableName string) *BaseRepository[T] {
	br := &BaseRepository[T]{
		DB:         database,
		TableName:  tableName,
		fieldCache: make(map[string]reflect.StructField),
	}
	var entity T
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		br.fieldCache[strings.ToLower(field.Name)] = field
	}
	var zero T
	table, err := newEntityTable(tableName, zero)
	if err != nil {
		panic(fmt.Sprintf("repository: %v", err))
	}
	br.table = table
	return br
}

// entity returns the entity metadata. A repository created as a struct literal has none
// cached, so it is built on every call; nothing is written to r, so there is no data race.
func (r *BaseRepository[T]) entity() (*entityTable, error) {
	if r.table != nil {
		return r.table, nil
	}
	var zero T
	return newEntityTable(r.TableName, zero)
}

// WithoutTenantScope returns a copy of the repository without the god/kar restriction.
// The receiver stays tenant-scoped, so other users of a shared repository are not affected.
func (r *BaseRepository[T]) WithoutTenantScope() *BaseRepository[T] {
	c := *r
	c.noTenantScope = true
	return &c
}

func (r *BaseRepository[T]) GetFieldCache() map[string]reflect.StructField {
	return r.fieldCache
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, idField string, idValue any) (*T, error) {
	var entity T
	query, args, err := r.CreateGetByIDStatement(ctx, idField, idValue)
	if err != nil {
		return nil, err
	}
	if err := r.DB.GetContext(ctx, &entity, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching entity by ID: %w", err)
	}
	return &entity, nil
}

// GetAll returns one page of records. sortBy must be a column of the entity; anything else is
// ignored and the list is sorted by idField. sortOrder is "desc" or anything else for ascending.
func (r *BaseRepository[T]) GetAll(ctx context.Context, pageSize int, offset int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]T, error) {
	var entities []T

	tb, err := r.entity()
	if err != nil {
		return nil, err
	}
	qb, err := r.CreateGetAllStatement(ctx, tableFields, idField, searchText)
	if err != nil {
		return nil, err
	}
	allowed := tb.sortColumns()
	key := strings.ToLower(strings.TrimSpace(sortBy))
	if _, ok := allowed[key]; !ok {
		key = strings.ToLower(idField)
		allowed[key] = key
	}
	qb.OrderByKey(key, sortOrder, allowed)
	qb.Limit(pageSize)
	qb.Offset(offset)

	query, args, err := qb.Build()
	if err != nil {
		return nil, err
	}
	if err := r.DB.SelectContext(ctx, &entities, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching entity records: %w", err)
	}
	return &entities, nil
}

func (r *BaseRepository[T]) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	countRec := 0
	query, args, err := r.CreateGetCountRecordsStatement(ctx, tableFields, searchText)
	if err != nil {
		return 0, err
	}
	if err := r.DB.GetContext(ctx, &countRec, query, args...); err != nil {
		return 0, fmt.Errorf("error fetching count of records: %w", err)
	}
	return countRec, nil
}

// Query is a statement built with querybuilder (a *querybuilder.Builder or *querybuilder.Union).
type Query interface {
	Build() (string, []any, error)
}

// Select runs a custom query (joins, CTEs, unions) and scans every row into T. Use
// ApplyFilters on the builder first when the query must be tenant-scoped.
func (r *BaseRepository[T]) Select(ctx context.Context, q Query) (*[]T, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, fmt.Errorf("build select on %s: %w", r.TableName, err)
	}
	var entities []T
	if err := r.DB.SelectContext(ctx, &entities, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching entity records: %w", err)
	}
	return &entities, nil
}

// Count returns the number of rows of b's query, ignoring its ORDER BY, LIMIT and OFFSET, so
// the builder of a page query can be reused for its total.
func (r *BaseRepository[T]) Count(ctx context.Context, b *querybuilder.Builder) (int, error) {
	query, args, err := b.BuildCount()
	if err != nil {
		return 0, fmt.Errorf("build count on %s: %w", r.TableName, err)
	}
	countRec := 0
	if err := r.DB.GetContext(ctx, &countRec, query, args...); err != nil {
		return 0, fmt.Errorf("error fetching count of records: %w", err)
	}
	return countRec, nil
}

// Exec runs a custom INSERT, UPDATE or DELETE. Unlike Update and Delete it does not require a
// matched row.
func (r *BaseRepository[T]) Exec(ctx context.Context, q Query) error {
	query, args, err := q.Build()
	if err != nil {
		return fmt.Errorf("build statement on %s: %w", r.TableName, err)
	}
	if _, err := r.DB.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	return nil
}

// ExecWithTx is Exec within an existing transaction.
func (r *BaseRepository[T]) ExecWithTx(ctx context.Context, tx db.Transaction, q Query) error {
	query, args, err := q.Build()
	if err != nil {
		return fmt.Errorf("build statement on %s: %w", r.TableName, err)
	}
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	return nil
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T, idField string, tableFields []domain.Fields) (int64, error) {
	query, values, err := r.CreateInsertStatement(ctx, entity, tableFields, idField)
	if err != nil {
		return 0, err
	}
	return doTransaction(ctx, r.DB, ActionTypeInsert, query, values...)
}

// Update changes one record. It returns ErrNotFound when no row matched (for example because
// the id belongs to another tenant).
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T, idField string, idValue any, tableFields []domain.Fields) error {
	query, values, err := r.CreateUpdateStatement(ctx, entity, idField, idValue, tableFields)
	if err != nil {
		return err
	}
	_, err = doTransaction(ctx, r.DB, ActionTypeUpdate, query, values...)
	return err
}

// Delete removes one record. It returns ErrNotFound when no row matched.
func (r *BaseRepository[T]) Delete(ctx context.Context, idField string, id int64) error {
	query, values, err := r.CreateDeleteStatement(ctx, idField, id)
	if err != nil {
		return err
	}
	_, err = doTransaction(ctx, r.DB, ActionTypeDelete, query, values...)
	return err
}

// ============ TRANSACTION-AWARE METHODS ============

// BeginTx starts a new database transaction
func (r *BaseRepository[T]) BeginTx() (db.Transaction, error) {
	return r.DB.Beginx()
}

// CreateWithTx executes an INSERT within an existing transaction
func (r *BaseRepository[T]) CreateWithTx(ctx context.Context, tx db.Transaction, entity *T, idField string, tableFields []domain.Fields) (int64, error) {
	query, values, err := r.CreateInsertStatement(ctx, entity, tableFields, idField)
	if err != nil {
		return 0, err
	}
	lastInsertedID := int64(0)
	if err := tx.QueryRowContext(ctx, query, values...).Scan(&lastInsertedID); err != nil {
		return 0, fmt.Errorf("query execution failed: %w", err)
	}
	return lastInsertedID, nil
}

// UpdateWithTx executes an UPDATE within an existing transaction (ErrNotFound if no row matched).
func (r *BaseRepository[T]) UpdateWithTx(ctx context.Context, tx db.Transaction, entity *T, idField string, idValue any, tableFields []domain.Fields) error {
	query, values, err := r.CreateUpdateStatement(ctx, entity, idField, idValue, tableFields)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	return requireAffected(res)
}

// DeleteWithTx executes a DELETE within an existing transaction (ErrNotFound if no row matched).
func (r *BaseRepository[T]) DeleteWithTx(ctx context.Context, tx db.Transaction, idField string, idValue int64) error {
	query, values, err := r.CreateDeleteStatement(ctx, idField, idValue)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	return requireAffected(res)
}

func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// doTransaction executes a query within a transaction.
func doTransaction(ctx context.Context, database db.Database, actionType, query string, values ...any) (int64, error) {
	tx, err := database.Beginx()
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful commit

	if actionType == ActionTypeDelete || actionType == ActionTypeUpdate {
		res, err := tx.ExecContext(ctx, query, values...)
		if err != nil {
			return 0, fmt.Errorf("query execution failed: %w", err)
		}
		if err := requireAffected(res); err != nil {
			return 0, err
		}
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("commit failed: %w", err)
		}
		return 0, nil
	}

	lastInsertedID := int64(0)
	if err := tx.QueryRowContext(ctx, query, values...).Scan(&lastInsertedID); err != nil {
		return 0, fmt.Errorf("query execution failed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit failed: %w", err)
	}
	return lastInsertedID, nil
}

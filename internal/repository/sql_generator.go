package repository

import (
	"context"
	"fmt"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
)

// SqlGenerator is implemented by BaseRepository. Every method returns an error instead of an
// empty query when something is invalid (missing session, bad column, ...).
type SqlGenerator[T any] interface {
	CreateInsertStatement(ctx context.Context, entity *T, tableFields []domain.Fields, idField string) (string, []any, error)
	CreateUpdateStatement(ctx context.Context, entity *T, idField string, idValue any, tableFields []domain.Fields) (string, []any, error)
	CreateDeleteStatement(ctx context.Context, idField string, idValue any) (string, []any, error)
	CreateGetByIDStatement(ctx context.Context, idField string, idValue any) (string, []any, error)
	CreateGetAllStatement(ctx context.Context, tableFields []domain.Fields, idField string, searchText string) (*querybuilder.Builder, error)
	CreateGetCountRecordsStatement(ctx context.Context, tableFields []domain.Fields, searchText string) (string, []any, error)
}

var _ SqlGenerator[struct{}] = (*BaseRepository[struct{}])(nil)

// applyScope restricts b to the session's tenant (god/kar). It does nothing for entities
// without those columns or after WithoutTenantScope().
func (r *BaseRepository[T]) applyScope(ctx context.Context, tb *entityTable, b *querybuilder.Builder, qual string) error {
	if r.noTenantScope || (!tb.hasGod && !tb.hasKar) {
		return nil
	}
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	tb.scopeTenant(b, qual, s)
	return nil
}

// CreateInsertStatement generates INSERT ... RETURNING. The entity argument is unused and
// kept for compatibility; the values come from tableFields.
func (r *BaseRepository[T]) CreateInsertStatement(ctx context.Context, _ *T, tableFields []domain.Fields, idField string) (string, []any, error) {
	tb, err := r.entity()
	if err != nil {
		return "", nil, err
	}
	s, err := sessionFromContext(ctx)
	if err != nil {
		return "", nil, err
	}
	return tb.insert(s, tableFields, idField)
}

// CreateUpdateStatement generates UPDATE ... WHERE id = $n (and the tenant scope).
func (r *BaseRepository[T]) CreateUpdateStatement(ctx context.Context, _ *T, idField string, idValue any, tableFields []domain.Fields) (string, []any, error) {
	tb, err := r.entity()
	if err != nil {
		return "", nil, err
	}
	s, err := sessionFromContext(ctx)
	if err != nil {
		return "", nil, err
	}
	b, err := tb.updateBuilder(s, tableFields, idField, idValue)
	if err != nil {
		return "", nil, err
	}
	if err := r.applyScope(ctx, tb, b, ""); err != nil {
		return "", nil, err
	}
	return b.Build()
}

// CreateDeleteStatement generates DELETE ... WHERE id = $n (and the tenant scope).
func (r *BaseRepository[T]) CreateDeleteStatement(ctx context.Context, idField string, idValue any) (string, []any, error) {
	tb, err := r.entity()
	if err != nil {
		return "", nil, err
	}
	b, err := tb.deleteBuilder(idField, idValue)
	if err != nil {
		return "", nil, err
	}
	if err := r.applyScope(ctx, tb, b, ""); err != nil {
		return "", nil, err
	}
	return b.Build()
}

// CreateGetByIDStatement generates SELECT ... WHERE id = $n (and the tenant scope).
func (r *BaseRepository[T]) CreateGetByIDStatement(ctx context.Context, idField string, idValue any) (string, []any, error) {
	tb, err := r.entity()
	if err != nil {
		return "", nil, err
	}
	b, err := tb.selectByIDBuilder(idField, idValue)
	if err != nil {
		return "", nil, err
	}
	if err := r.applyScope(ctx, tb, b, ""); err != nil {
		return "", nil, err
	}
	return b.Build()
}

// CreateGetAllStatement returns a builder with SELECT, tenant scope and the free-text search
// already applied. Callers add joins, filters, sorting and paging, then Build().
func (r *BaseRepository[T]) CreateGetAllStatement(ctx context.Context, tableFields []domain.Fields, idField string, searchText string) (*querybuilder.Builder, error) {
	tb, err := r.entity()
	if err != nil {
		return nil, err
	}
	b, err := tb.selectAllBuilder(tableFields, idField)
	if err != nil {
		return nil, err
	}
	if err := r.applyScope(ctx, tb, b, querybuilder.QuoteIdent(tb.name)+"."); err != nil {
		return nil, err
	}
	b.AddSearch(tb.searchColumns(tableFields), searchText)
	return b, nil
}

// CreateGetCountRecordsStatement generates SELECT count(*) with tenant scope and search.
func (r *BaseRepository[T]) CreateGetCountRecordsStatement(ctx context.Context, tableFields []domain.Fields, searchText string) (string, []any, error) {
	tb, err := r.entity()
	if err != nil {
		return "", nil, err
	}
	b := tb.countBuilder()
	if err := r.applyScope(ctx, tb, b, querybuilder.QuoteIdent(tb.name)+"."); err != nil {
		return "", nil, err
	}
	b.AddSearch(tb.searchColumns(tableFields), searchText)
	return b.Build()
}

// GetHasGodHasKar reports whether the entity has the god and kar columns.
func (r *BaseRepository[T]) GetHasGodHasKar() (bool, bool) {
	tb, err := r.entity()
	if err != nil {
		return false, false
	}
	return tb.hasGod, tb.hasKar
}

// ApplyFilters adds the tenant scope and the free-text search to any builder, for custom
// queries with joins. alias is the table alias used in that query ("" = none; must be a plain
// lower-case identifier).
func (r *BaseRepository[T]) ApplyFilters(ctx context.Context, b *querybuilder.Builder, alias string, tableFields []domain.Fields, searchText string) error {
	tb, err := r.entity()
	if err != nil {
		return err
	}
	qual := ""
	if alias != "" {
		if !querybuilder.IsIdent(alias) {
			return fmt.Errorf("invalid table alias %q", alias)
		}
		qual = alias + "."
	}
	if err := r.applyScope(ctx, tb, b, qual); err != nil {
		return err
	}
	b.AddSearch(tb.searchColumns(tableFields), searchText)
	return nil
}

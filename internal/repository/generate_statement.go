package repository

import (
	"context"
	"helia/internal/common"
	"helia/internal/domain"
	"reflect"
)

type SqlGenerator[T any] interface {
	CreateInsertStatement(ctx context.Context, entity *T, tableFields []domain.Fields, idField string) (string, []interface{})
	CreateUpdateStatement(ctx context.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{})
	CreateGetByIDStatement(idField string, idValue interface{}) string
	CreateGetAllStatement(ctx context.Context, tableFields []domain.Fields, idField string, searchParams ...string) (string, []interface{})
	CreateGetCountRecordsStatement(ctx context.Context, tableFields []domain.Fields, searchParams ...string) (string, []interface{})
}

// CreateInsertStatement generates an INSERT query using QueryBuilder
func (r *BaseRepository[T]) CreateInsertStatement(ctx context.Context, entity *T, tableFields []domain.Fields, idField string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(*entity),
	}
	qb := common.NewRepositoryQueryBuilder(config)

	sqlQuery, args := qb.BuildInsert(ctx, tableFields, idField)
	return sqlQuery, args
}

// CreateUpdateStatement generates an UPDATE query using QueryBuilder
func (r *BaseRepository[T]) CreateUpdateStatement(ctx context.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(*entity),
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.BuildUpdate(ctx, tableFields, idField, idValue)
}

// CreateGetByID generates a SELECT query for a single record by ID
func (r *BaseRepository[T]) CreateGetByIDStatement(idField string, idValue interface{}) string {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(new(T)).Elem(),
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.BuildSelectByID(idField)
}

// CreateGetAll generates a SELECT query for multiple records with filtering
func (r *BaseRepository[T]) CreateGetAllStatement(ctx context.Context, tableFields []domain.Fields, idField string, searchParams ...string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}
	qbRepo := common.NewRepositoryQueryBuilder(config)
	qb := common.NewQueryBuilder(qbRepo.BuildSelectAll(tableFields, idField, searchParams...), true)

	return qb.Build()

}

// CreateGetCountRecords generates a COUNT query
func (r *BaseRepository[T]) CreateGetCountRecordsStatement(ctx context.Context, tableFields []domain.Fields, searchParams ...string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	query := qb.BuildCount(tableFields, searchParams...)

	return query, qb.GetArgs()
}

// GetHasGodHasKar checks if entity has god and kar fields (alias for CheckGogKar)
func (r *BaseRepository[T]) GetHasGodHasKar() (bool, bool) {
	entityType := reflect.TypeOf(new(T)).Elem()
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: entityType,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.CheckGodKarFields()
}

// CreateBasicWhere generates a WHERE clause with god/kar and search conditions
func (r *BaseRepository[T]) CreateBasicWhere(tableFields []domain.Fields, args *[]interface{}, hasGod, hasKar bool, god, kar int, searchParams ...string) string {
	entityType := reflect.TypeOf(new(T)).Elem()
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  entityType,
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	qb.AddGodKarConditions(hasGod, hasKar, god, kar)
	qb.AddSearchConditions(tableFields, searchParams...)
	*args = qb.GetArgs()
	return qb.GetWhereClause()
}

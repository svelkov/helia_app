package repository

import (
	"helia/internal/common"
	"helia/internal/domain"
	"reflect"
)

type SqlGenerator[T any] interface {
	CreateInsertStatement(entity *T, tableFields []domain.Fields) (string, []interface{})
	CreateUpdateStatement(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{})
	CreateGetByID(idField string, idValue interface{}) string
	CreateGetAll(tableFields []domain.Fields, idField string, searchParams ...string) (string, []interface{})
	CreateGetCountRecords(tableFields []domain.Fields, searchParams ...string) (string, []interface{})
}

// CreateInsertStatement generates an INSERT query using QueryBuilder
func (r *BaseRepository[T]) CreateInsertStatement(entity *T, tableFields []domain.Fields) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(*entity),
	}
	qb := common.NewRepositoryQueryBuilder(config)

	// Extract ID field name from the entity
	idField := "id"
	for i := 0; i < reflect.TypeOf(*entity).NumField(); i++ {
		field := reflect.TypeOf(*entity).Field(i)
		if dbTag := field.Tag.Get("db"); dbTag == "id" {
			idField = dbTag
			break
		}
	}

	return qb.BuildInsert(tableFields, idField)
}

// CreateUpdateStatement generates an UPDATE query using QueryBuilder
func (r *BaseRepository[T]) CreateUpdateStatement(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(*entity),
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.BuildUpdate(tableFields, idField, idValue)
}

// CreateGetByID generates a SELECT query for a single record by ID
func (r *BaseRepository[T]) CreateGetByID(idField string, idValue interface{}) string {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(new(T)).Elem(),
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.BuildSelectByID(idField)
}

// CreateGetAll generates a SELECT query for multiple records with filtering
func (r *BaseRepository[T]) CreateGetAll(tableFields []domain.Fields, idField string, searchParams ...string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	query := qb.BuildSelectAll(tableFields, idField, searchParams...)
	return query, qb.GetArgs()
}

// CreateGetCountRecords generates a COUNT query
func (r *BaseRepository[T]) CreateGetCountRecords(tableFields []domain.Fields, searchParams ...string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	query := qb.BuildCount(tableFields, searchParams...)
	return query, qb.GetArgs()
}

// CheckGogKar checks if entity has god and kar fields
func (r *BaseRepository[T]) CheckGogKar() (bool, bool) {
	entityType := reflect.TypeOf(new(T)).Elem()
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: entityType,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.CheckGodKarFields()
}

// GetHasGodHasKar checks if entity has god and kar fields (alias for CheckGogKar)
func (r *BaseRepository[T]) GetHasGodHasKar() (bool, bool) {
	return r.CheckGogKar()
}

// CreateBasicWhere generates a WHERE clause with god/kar and search conditions
func (r *BaseRepository[T]) CreateBasicWhere(tableFields []domain.Fields, args *[]interface{}, hasGod, hasKar bool, searchParams ...string) string {
	entityType := reflect.TypeOf(new(T)).Elem()
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  entityType,
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	qb.AddGodKarConditions(hasGod, hasKar)
	qb.AddSearchConditions(tableFields, searchParams...)
	*args = qb.GetArgs()
	return qb.GetWhereClause()
}

package repository

import (
	"helia/internal/common"
	"helia/internal/domain"
	"reflect"

	"github.com/gin-gonic/gin"
)

type SqlGenerator[T any] interface {
	CreateInsertStatement(entity *T, tableFields []domain.Fields) (string, []interface{})
	CreateUpdateStatement(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{})
	CreateGetByIDStatement(idField string, idValue interface{}) string
	CreateGetAllStatement(c *gin.Context, tableFields []domain.Fields, idField string, searchParams ...string) (string, []interface{})
	CreateGetCountRecordsStatement(c *gin.Context, tableFields []domain.Fields, searchParams ...string) (string, []interface{})
}

// CreateInsertStatement generates an INSERT query using QueryBuilder
func (r *BaseRepository[T]) CreateInsertStatement(c *gin.Context, entity *T, tableFields []domain.Fields, idField string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(*entity),
	}
	qb := common.NewRepositoryQueryBuilder(config)

	sqlQuery, args := qb.BuildInsert(c, tableFields, idField)
	return sqlQuery, args
}

// CreateUpdateStatement generates an UPDATE query using QueryBuilder
func (r *BaseRepository[T]) CreateUpdateStatement(c *gin.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:  r.TableName,
		EntityType: reflect.TypeOf(*entity),
	}
	qb := common.NewRepositoryQueryBuilder(config)
	return qb.BuildUpdate(c, tableFields, idField, idValue)
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
func (r *BaseRepository[T]) CreateGetAllStatement(c *gin.Context, tableFields []domain.Fields, idField string, searchParams ...string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	userSession := domain.GetSessionFromContext(c)
	hasGod, hasKar := r.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	query := qb.BuildSelectAll(tableFields, idField, searchParams...)

	return query, qb.GetArgs()
}

// CreateGetCountRecords generates a COUNT query
func (r *BaseRepository[T]) CreateGetCountRecordsStatement(c *gin.Context, tableFields []domain.Fields, searchParams ...string) (string, []interface{}) {
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	userSession := domain.GetSessionFromContext(c)
	hasGod, hasKar := r.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
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

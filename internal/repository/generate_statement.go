package repository

import (
	"context"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"reflect"
	"strings"
)

type SqlGenerator[T any] interface {
	CreateInsertStatement(ctx context.Context, entity *T, tableFields []domain.Fields, idField string) (string, []interface{})
	CreateUpdateStatement(ctx context.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{})
	CreateGetByIDStatement(idField string, idValue interface{}) string
	CreateGetAllStatement(ctx context.Context, tableFields []domain.Fields, idField string, searchText string) (string, []interface{})
	CreateGetCountRecordsStatement(ctx context.Context, tableFields []domain.Fields, searchText string) (string, []interface{})
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
func (r *BaseRepository[T]) CreateGetAllStatement(ctx context.Context, tableFields []domain.Fields, idField string, searchText string) *common.QueryBuilder {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil
	}
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  reflect.TypeOf(new(T)).Elem(),
		TableFields: tableFields,
	}

	qbRepo := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qbRepo.CheckGodKarFields()
	qb := common.NewQueryBuilder(qbRepo.BuildSelectAll(tableFields, idField, searchText), true)
	qb.SetEntityType(config.EntityType)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)

	return qb

}

// CreateGetCountRecords generates a COUNT query
func (r *BaseRepository[T]) CreateGetCountRecordsStatement(ctx context.Context, tableFields []domain.Fields, searchText string) (string, []interface{}) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return "", nil
	}
	qb := common.NewQueryBuilder(fmt.Sprintf(" select count(*) from %s", r.TableName), true)
	qb.SetEntityType(reflect.TypeOf(new(T)).Elem())
	hasGod, hasKar := qb.CheckGodKarFields()
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	// Note: god/kar filtering should be added explicitly in service layer
	qb.AddSearchConditions(tableFields, searchText)

	query, args := qb.Build()
	return query, args
}

// GetHasGodHasKar checks if entity has god and kar fields (alias for CheckGodKar)
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
func (r *BaseRepository[T]) CreateBasicWhere(tableFields []domain.Fields, args *[]interface{}, hasGod, hasKar bool, god, kar int, searchText string) string {
	entityType := reflect.TypeOf(new(T)).Elem()
	config := common.RepositoryConfig{
		TableName:   r.TableName,
		EntityType:  entityType,
		TableFields: tableFields,
	}
	qb := common.NewRepositoryQueryBuilder(config)
	qb.AddGodKarConditions(hasGod, hasKar, god, kar)
	qb.AddSearchConditions(tableFields, searchText)
	*args = qb.GetArgs()
	return qb.GetWhereClause()
}

func getFieldsForEntity[T any](entity *T, idField string, fields []domain.Fields) []string {
	entityType := reflect.TypeOf(*entity)
	columns := make([]string, 0, entityType.NumField())

	seen := make(map[string]struct{})
	addCol := func(col string) {
		if _, exists := seen[col]; !exists {
			seen[col] = struct{}{}
			columns = append(columns, col)
		}
	}

	addCol(strings.ToLower(idField))

	// Collect all columns
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		column := field.Tag.Get("db")
		if column == "" || column == "-" {
			continue
		}

		lowerCol := strings.ToLower(column)
		if lowerCol == "god" || lowerCol == "kar" {
			addCol(lowerCol)
			continue
		}

		// Check if field is in tableFields
		for _, f := range fields {
			if strings.EqualFold(f.Name, column) {
				addCol(lowerCol)
				break
			}
		}
	}
	return columns
}

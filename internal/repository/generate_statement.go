package repository

import (
	"fmt"
	"helia/internal/domain"
	"reflect"
	"strings"
	"time"
)

const (
	god        = "god"
	kar        = "kar"
	xopunos    = "xopunos"
	xdatunosa  = "xdatunosa"
	xopizmene  = "xopizmene"
	xdatizmene = "xdatizmene"
)

type SqlGenerator[T any] interface {
	CreateInsertStatement(entity *T, tableFields []domain.Fields) (string, []interface{})
	CreateUpdateStatement(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{})
	CreateGetByID(idField string, idValue interface{}) string
	CreateGetAll(args *[]interface{}, tableFields []domain.Fields, idField string, searchParams ...string) string
	CheckGogKar() (bool, bool)
	CreateGetCountRecords(tableFields []domain.Fields, args *[]interface{}, searchParams ...string) string
	GetFieldType(val reflect.Value, fieldName string) (interface{}, bool)
	CreateBasicWhere(tableFields []domain.Fields, args *[]interface{}, hasGod, hasKar bool, searchParams ...string) string
	GetHasGodHasKar() (bool, bool)
}

func (r *BaseRepository[T]) CreateInsertStatement(entity *T, tableFields []domain.Fields) (string, []interface{}) {
	// Get the type and value of the entity
	entityType := reflect.TypeOf(*entity)

	// Prepare slices for columns and placeholders
	var columns []string
	var placeholders []string
	var values []interface{}

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		column := field.Tag.Get("db") // Get the `db` tag value
		if column == "" || column == "-" {
			continue // Skip fields without `db` tags or explicitly ignored
		}
		if strings.ToLower(column) == "god" {
			columns = append(columns, strings.ToLower(column))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
			values = append(values, r.cfg.God)
			continue
		}
		if strings.ToLower(column) == "kar" {
			columns = append(columns, strings.ToLower(column))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
			values = append(values, r.cfg.Kar)
			continue
		}
		if strings.ToLower(column) == xdatunosa {
			columns = append(columns, strings.ToLower(column))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
			values = append(values, time.Now())
			continue
		}
	}
	// Extract specified fields dynamically
	for _, fieldName := range tableFields {
		columns = append(columns, fieldName.Name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
		values = append(values, fieldName.Value)
	}

	// Construct the SQL query
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		r.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, values
}

func (r *BaseRepository[T]) CreateUpdateStatement(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) (string, []interface{}) {

	// Prepare slices for columns and values
	var columns []string
	var values []interface{}

	for _, field := range tableFields {
		columns = append(columns, fmt.Sprintf(` %s = $%d`, field.Name, len(values)+1))
		values = append(values, field.Value)
	}
	//add xupdatedate
	columns = append(columns, fmt.Sprintf(` %s = $%d`, "xdatizmene", len(values)+1))
	values = append(values, time.Now())

	// Append the ID field for the WHERE clause
	values = append(values, idValue)
	query := fmt.Sprintf(
		`UPDATE "%s" SET %s WHERE "%s" = $%d`,
		r.TableName,
		strings.Join(columns, ", "),
		idField,
		len(values),
	)

	// return the query
	return query, values
}

func (r *BaseRepository[T]) CreateGetByID(idField string, idValue interface{}) string {
	// Get the type of the entity
	entityType := reflect.TypeOf(new(T)).Elem()

	// Prepare slices for columns and pointers
	var columns []string

	// Iterate over fields of the struct to build the SELECT clause
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		column := field.Tag.Get("db") // Get the `db` tag value
		if column == "" || column == "-" {
			continue // Skip fields without `db` tags or explicitly ignored
		}
		columns = append(columns, fmt.Sprintf(`"%s"`, column))
	}

	// Construct the SELECT query
	query := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE "%s" = $1`,
		strings.Join(columns, ", "),
		r.TableName,
		idField,
	)

	return query
}
func (r *BaseRepository[T]) CreateGetAll(args *[]interface{}, tableFields []domain.Fields, idField string, searchParams ...string) string {
	hasGod := false
	hasKar := false
	// Get the type of the entity
	entityType := reflect.TypeOf(new(T)).Elem()

	// Prepare slices for columns and pointers
	var columns []string
	columns = append(columns, idField)
	// Iterate over fields of the struct to build the SELECT clause
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		column := field.Tag.Get("db") // Get the `db` tag value
		if column == "" || column == "-" {
			continue // Skip fields without `db` tags or explicitly ignored
		}
		if strings.ToLower(column) == "god" {
			hasGod = true
			columns = append(columns, column)
		}
		if strings.ToLower(column) == "kar" {
			hasKar = true
			columns = append(columns, strings.ToLower(column))
		}
		for _, field := range tableFields {
			if strings.EqualFold(field.Name, column) {
				columns = append(columns, strings.ToLower(column))
				break
			}
		}

	}

	sqlWhereBasic := r.CreateBasicWhere(tableFields, args, hasGod, hasKar, searchParams...)
	// Construct the SELECT query
	query := fmt.Sprintf(`SELECT %s FROM %s  %s ORDER BY %s`, strings.Join(columns, ", "), r.TableName, sqlWhereBasic, idField)

	return query
}

func (r *BaseRepository[T]) CheckGogKar() (bool, bool) {
	hasGod := false
	hasKar := false
	// Get the type of the entity
	entityType := reflect.TypeOf(new(T)).Elem()

	// Iterate over fields of the struct to build the SELECT clause
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		column := field.Tag.Get("db") // Get the `db` tag value
		if column == "" || column == "-" {
			continue // Skip fields without `db` tags or explicitly ignored
		}
		if strings.ToLower(column) == "god" {
			hasGod = true
			continue
		}
		if strings.ToLower(column) == "kar" {
			hasKar = true
			continue
		}
	}
	return hasGod, hasKar

}

func (r *BaseRepository[T]) CreateGetCountRecords(tableFields []domain.Fields, args *[]interface{}, searchParams ...string) string {
	hasGod, hasKar := r.GetHasGodHasKar()
	sqlWhereBasic := r.CreateBasicWhere(tableFields, args, hasGod, hasKar, searchParams...)
	// Construct the SELECT query
	query := fmt.Sprintf(
		`SELECT count(*) FROM %s  %s `,
		r.TableName, sqlWhereBasic,
	)

	return query
}

func (r *BaseRepository[T]) GetFieldType(val reflect.Value, fieldName string) (interface{}, bool) {

	fieldNameLower := strings.ToLower(fieldName)

	// Iterate over struct fields
	for i := 0; i < val.NumField(); i++ {
		structField := val.Type().Field(i)
		if strings.ToLower(structField.Name) == fieldNameLower {
			return val.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func (r *BaseRepository[T]) CreateBasicWhere(tableFields []domain.Fields, args *[]interface{}, hasGod, hasKar bool, searchParams ...string) string {
	sqlWhereBasic := " WHERE 1=1 "
	paramNr := 1
	if hasGod {
		sqlWhereBasic = fmt.Sprintf("%s AND god = $1 ", sqlWhereBasic)
		*args = append(*args, r.cfg.God)
		paramNr = 2
	}
	if hasKar {
		sqlWhereBasic = fmt.Sprintf("%s AND kar = $2 ", sqlWhereBasic)
		*args = append(*args, r.cfg.Kar)
		paramNr = 3
	}
	likeString := ""
	if len(searchParams) >= 1 && len(searchParams[0]) > 0 {
		// Get the type of the entity
		entityType := reflect.TypeOf(new(T)).Elem()
		likeString += ` AND ( `
		// Iterate over fields of the struct to build the SELECT clause
		for i := 0; i < entityType.NumField(); i++ {
			field := entityType.Field(i)
			//no search on god and kar
			for _, tblField := range tableFields {
				if !strings.EqualFold(tblField.Name, field.Name) || strings.ToLower(field.Name) == "god" || strings.ToLower(field.Name) == "kar" {
					continue
				}
				switch field.Type.Name() {
				case "int", "int64", "int32", "float32", "float64":
					likeString += fmt.Sprintf(" OR (%s::TEXT ILIKE $%d)", strings.ToLower(field.Name), paramNr)
				case "string":
					likeString += fmt.Sprintf(" OR (%s::TEXT ILIKE $%d)", strings.ToLower(field.Name), paramNr)
				case "bool":
					likeString += fmt.Sprintf(" OR (%s::TEXT ILIKE $%d)", strings.ToLower(field.Name), paramNr)
				default:
					continue
				}
				*args = append(*args, fmt.Sprintf("%%%s%%", searchParams[0]))
				paramNr++
			}
		}
		likeString += ` )`
		likeString = strings.ReplaceAll(likeString, "AND (  OR", "AND ( ")
	}
	sqlWhereBasic = fmt.Sprintf("%s %s ", sqlWhereBasic, likeString)

	return sqlWhereBasic
}

func (r *BaseRepository[T]) GetHasGodHasKar() (bool, bool) {
	hasGod := false
	hasKar := false
	// Get the type of the entity
	entityType := reflect.TypeOf(new(T)).Elem()

	// Iterate over fields of the struct to build the SELECT clause
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		column := field.Tag.Get("db") // Get the `db` tag value
		if column == "" || column == "-" {
			continue // Skip fields without `db` tags or explicitly ignored
		}
		if strings.ToLower(column) == "god" {
			hasGod = true
			continue
		}
		if strings.ToLower(column) == "kar" {
			hasKar = true
			continue
		}
	}
	return hasGod, hasKar
}

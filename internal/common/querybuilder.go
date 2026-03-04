package common

import (
	"fmt"
	"helia/internal/domain"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type QueryBuilder struct {
	baseQuery   string
	whereClause strings.Builder
	args        []interface{}
	paramCount  int
	joins       []string
	orderBy     string
	sortOrder   string
	groupBy     string
	having      string
	limit       string
	offset      string
	// Repository-specific fields
	tableName    string
	entityType   reflect.Type
	selectFields []string
}

// RepositoryConfig holds configuration for repository operations
type RepositoryConfig struct {
	TableName     string
	EntityType    reflect.Type
	TableFields   []domain.Fields
	IncludeGodKar bool
}

type QueryType string

const (
	SelectQuery QueryType = "select"
	InsertQuery QueryType = "insert"
	UpdateQuery QueryType = "update"
	DeleteQuery QueryType = "delete"
	CountQuery  QueryType = "count"
)

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	qb := &QueryBuilder{
		baseQuery:  baseQuery,
		paramCount: 1,
	}
	qb.whereClause.WriteString(" WHERE 1 = 1 ")
	return qb
}

// NewRepositoryQueryBuilder creates a query builder for repository operations
func NewRepositoryQueryBuilder(config RepositoryConfig) *QueryBuilder {
	qb := &QueryBuilder{
		tableName:  config.TableName,
		entityType: config.EntityType,
		paramCount: 1,
	}
	qb.whereClause.WriteString(" WHERE 1 = 1 ")
	return qb
}

func (qb *QueryBuilder) SetEntityType(entityType reflect.Type) {
	qb.entityType = entityType
}

// AddCondition adds a simple condition with operator
func (qb *QueryBuilder) AddCondition(field string, value interface{}, operator string) *QueryBuilder {
	if value != nil && value != "" {
		qb.whereClause.WriteString(fmt.Sprintf(" AND %s %s $%d", field, operator, qb.paramCount))
		qb.args = append(qb.args, value)
		qb.paramCount++
	}
	return qb
}

// AddEqual condition (most common)
func (qb *QueryBuilder) AddEqual(field string, value interface{}) *QueryBuilder {
	return qb.AddCondition(field, value, "=")
}

// AddLike condition for partial matches
func (qb *QueryBuilder) AddLike(field string, value interface{}) *QueryBuilder {
	if value != nil && value != "" {
		qb.whereClause.WriteString(fmt.Sprintf(" AND %s ILIKE $%d", field, qb.paramCount))
		qb.args = append(qb.args, "%"+value.(string)+"%")
		qb.paramCount++
	}
	return qb
}

// AddLike condition for partial matches
func (qb *QueryBuilder) AddLikeBegin(field string, value interface{}) *QueryBuilder {
	if value != nil && value != "" {
		qb.whereClause.WriteString(fmt.Sprintf(" AND %s ILIKE $%d", field, qb.paramCount))
		qb.args = append(qb.args, value.(string)+"%")
		qb.paramCount++
	}
	return qb
}

// AddIn condition for multiple values
func (qb *QueryBuilder) AddIn(field string, values []interface{}) *QueryBuilder {
	if len(values) > 0 {
		placeholders := make([]string, len(values))
		for i, val := range values {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount)
			qb.args = append(qb.args, val)
			qb.paramCount++
		}
		qb.whereClause.WriteString(fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", ")))
	}
	return qb
}

// AddJoin adds JOIN clauses
func (qb *QueryBuilder) AddJoin(joinClause string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause)
	return qb
}

// AddOrderBy adds ORDER BY clause
func (qb *QueryBuilder) AddOrderBy(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// SetSortOrder sets the sort order (ASC or DESC)
func (qb *QueryBuilder) AddSortOrder(sortOrder string) *QueryBuilder {
	qb.sortOrder = sortOrder
	return qb
}

// AddGroupBy adds GROUP BY clause
func (qb *QueryBuilder) AddGroupBy(groupBy string) *QueryBuilder {
	qb.groupBy = groupBy
	return qb
}

// AddHaving adds HAVING clause
func (qb *QueryBuilder) AddHaving(having string) *QueryBuilder {
	qb.having = having
	return qb
}

// SetLimit sets LIMIT
func (qb *QueryBuilder) SetLimit(limit int) *QueryBuilder {
	if limit > 0 {
		qb.limit = fmt.Sprintf("LIMIT %d", limit)
	}
	return qb
}

// SetOffset sets OFFSET
func (qb *QueryBuilder) SetOffset(offset int) *QueryBuilder {
	if offset > 0 {
		qb.offset = fmt.Sprintf("OFFSET %d", offset)
	}
	return qb
}

// AddCustomCondition for complex conditions (joined with AND)
func (qb *QueryBuilder) AddCustomCondition(condition string, values ...interface{}) *QueryBuilder {
	qb.whereClause.WriteString(" AND " + condition)
	for _, val := range values {
		if val != nil && val != "" {
			qb.args = append(qb.args, val)
		}
	}
	return qb
}

// AddOrCondition for OR conditions
func (qb *QueryBuilder) AddOrCondition(condition string, values ...interface{}) *QueryBuilder {
	qb.whereClause.WriteString(" OR " + condition)
	for _, val := range values {
		if val != nil && val != "" {
			qb.args = append(qb.args, val)
		}
	}
	return qb
}

// Build constructs the final SQL query and arguments
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	query.WriteString(qb.baseQuery)

	// Add JOINs
	for _, join := range qb.joins {
		query.WriteString(" " + join)
	}

	// Add WHERE clause
	query.WriteString(" " + qb.whereClause.String())

	// Add GROUP BY
	if qb.groupBy != "" {
		query.WriteString(" GROUP BY " + qb.groupBy)
	}

	// Add HAVING
	if qb.having != "" {
		query.WriteString(" HAVING " + qb.having)
	}

	// Add ORDER BY
	if qb.orderBy != "" {
		query.WriteString(" ORDER BY " + qb.orderBy)
	}
	// Add SORT ORDER
	if qb.sortOrder != "" {
		query.WriteString(" " + qb.sortOrder)
	}

	// Add LIMIT and OFFSET
	if qb.limit != "" {
		query.WriteString(" " + qb.limit)
	}
	if qb.offset != "" {
		query.WriteString(" " + qb.offset)
	}

	return query.String(), qb.args
}

// GetArgs returns current arguments (useful for debugging)
func (qb *QueryBuilder) GetArgs() []interface{} {
	return qb.args
}

// GetArgs returns current arguments (useful for debugging)
func (qb *QueryBuilder) GetArgsCount() int {
	return len(qb.args)
}
func (qb *QueryBuilder) AddArgs(args ...interface{}) {
	qb.args = append(qb.args, args...)
	qb.paramCount += len(args)
}

// ==================== Repository-Specific Methods ====================

// BuildInsert constructs an INSERT query with RETURNING clause
func (qb *QueryBuilder) BuildInsert(c *gin.Context, fields []domain.Fields, idField string) (string, []interface{}) {
	if qb.tableName == "" {
		return "", nil
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	// Check if entity has god/kar fields
	hasGod, hasKar := qb.CheckGodKarFields()

	// Build a map of field names from provided fields for quick lookup
	fieldMap := make(map[string]domain.Fields)
	for _, field := range fields {
		fieldMap[strings.ToLower(field.Name)] = field
	}
	userSession := domain.GetSessionFromContext(c)
	// Add god/kar first if entity has them
	if hasGod {
		columns = append(columns, "god")
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
		values = append(values, userSession.SelectedGod)
	}
	if hasKar {
		columns = append(columns, "kar")
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
		values = append(values, userSession.SelectedKar)
	}

	// Add xopunos
	columns = append(columns, "xopunos")
	placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
	values = append(values, userSession.UserName)
	// Add xdatunosa
	columns = append(columns, "xdatunosa")
	placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
	values = append(values, time.Now())

	// Add remaining provided fields (skip god, kar, xdatunosa)
	for _, field := range fields {
		lowerName := strings.ToLower(field.Name)
		if lowerName == "god" || lowerName == "kar" || lowerName == "xdatunosa" || lowerName == "xopunos" {
			continue
		}
		columns = append(columns, strings.ToLower(field.Name))
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
		values = append(values, field.Value)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING %s`,
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		idField,
	)

	return query, values
}

// BuildUpdate constructs an UPDATE query
func (qb *QueryBuilder) BuildUpdate(c *gin.Context, fields []domain.Fields, idField string, idValue interface{}) (string, []interface{}) {
	if qb.tableName == "" {
		return "", nil
	}

	var columns []string
	var values []interface{}
	userSession := domain.GetSessionFromContext(c)
	// Add provided fields
	for _, field := range fields {
		if strings.ToLower(field.Name) == "xdatizmene" || strings.ToLower(field.Name) == "xopizmene" {
			continue // Skip update timestamp and user if provided in fields
		}
		columns = append(columns, fmt.Sprintf(` %s = $%d`, strings.ToLower(field.Name), len(values)+1))
		values = append(values, field.Value)
	}

	// Add update user
	columns = append(columns, fmt.Sprintf(` %s = $%d`, "xopizmene", len(values)+1))
	values = append(values, userSession.UserName)

	// Add update timestamp
	columns = append(columns, fmt.Sprintf(` %s = $%d`, "xdatizmene", len(values)+1))
	values = append(values, time.Now())

	// Add ID for WHERE clause
	values = append(values, idValue)

	query := fmt.Sprintf(
		`UPDATE "%s" SET %s WHERE "%s" = $%d`,
		qb.tableName,
		strings.Join(columns, ", "),
		idField,
		len(values),
	)

	return query, values
}

// BuildSelectByID constructs a SELECT query for single record by ID
func (qb *QueryBuilder) BuildSelectByID(idField string) string {
	if qb.tableName == "" || qb.entityType == nil {
		return ""
	}

	var columns []string
	for i := 0; i < qb.entityType.NumField(); i++ {
		field := qb.entityType.Field(i)
		column := field.Tag.Get("db")
		shouldGetField := field.Tag.Get("addupdate")
		if shouldGetField == "false" {
			continue
		}
		if column == "" || column == "-" {
			continue
		}
		columns = append(columns, fmt.Sprintf(`"%s"`, column))
	}

	query := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE "%s" = $1`,
		strings.Join(columns, ", "),
		qb.tableName,
		idField,
	)

	return query
}

// BuildSelectAll constructs a SELECT query for multiple records with filtering
func (qb *QueryBuilder) BuildSelectAll(fields []domain.Fields, idField string, searchParams ...string) string {
	if qb.tableName == "" || qb.entityType == nil {
		return ""
	}

	var columns []string
	columns = append(columns, idField)

	// Collect all columns
	for i := 0; i < qb.entityType.NumField(); i++ {
		field := qb.entityType.Field(i)
		column := field.Tag.Get("db")
		if column == "" || column == "-" {
			continue
		}

		lowerCol := strings.ToLower(column)
		if lowerCol == "god" || lowerCol == "kar" {
			columns = append(columns, lowerCol)
			continue
		}

		// Check if field is in tableFields
		for _, f := range fields {
			if strings.EqualFold(f.Name, column) {
				columns = append(columns, lowerCol)
				break
			}
		}
	}
	qb.baseQuery = fmt.Sprintf(`SELECT %s FROM "%s" `, strings.Join(columns, ", "), qb.tableName)
	// Build WHERE clause
	// Note: god/kar filtering should be added explicitly in service layer
	//qb.AddSearchConditions(fields, searchParams...)

	return qb.baseQuery
}

// BuildCount constructs a COUNT query
func (qb *QueryBuilder) BuildCount(fields []domain.Fields, searchParams ...string) string {
	if qb.tableName == "" || qb.entityType == nil {
		return ""
	}

	// Note: god/kar filtering should be added explicitly in service layer
	qb.AddSearchConditions(fields, searchParams...)

	whereClause := qb.whereClause.String()
	query := fmt.Sprintf(`SELECT count(*) FROM %s %s`, qb.tableName, whereClause)

	return query
}

// BuildDelete constructs a DELETE query
func (qb *QueryBuilder) BuildDelete(idField string, idValue interface{}) (string, []interface{}) {
	if qb.tableName == "" {
		return "", nil
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, qb.tableName, idField)
	return query, []interface{}{idValue}
}

// Helper methods

// CheckGodKarFields checks if entity has god and kar fields
func (qb *QueryBuilder) CheckGodKarFields() (bool, bool) {
	if qb.entityType == nil {
		return false, false
	}

	hasGod, hasKar := false, false
	for i := 0; i < qb.entityType.NumField(); i++ {
		field := qb.entityType.Field(i)
		column := field.Tag.Get("db")

		// Check if field is a struct and recursively check its fields
		if field.Type.Kind() == reflect.Struct {
			hasGod2, hasKar2 := checkNestedStructFields(field.Type)
			hasGod = hasGod || hasGod2
			hasKar = hasKar || hasKar2
			break
		}
		if column == "" || column == "-" {
			continue
		}

		lowerCol := strings.ToLower(column)
		switch lowerCol {
		case "god":
			hasGod = true
		case "kar":
			hasKar = true
		}
	}
	return hasGod, hasKar
}

// checkNestedStructFields recursively checks nested struct fields for god/kar columns
func checkNestedStructFields(structType reflect.Type) (bool, bool) {
	hasGod, hasKar := false, false

	if structType == nil || structType.Kind() != reflect.Struct {
		return false, false
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		column := field.Tag.Get("db")
		if column == "" || column == "-" {
			continue
		}

		// Recursively check if this field is also a struct
		if field.Type.Kind() == reflect.Struct {
			hasGod2, hasKar2 := checkNestedStructFields(field.Type)
			hasGod = hasGod || hasGod2
			hasKar = hasKar || hasKar2
			continue
		}

		lowerCol := strings.ToLower(column)
		if lowerCol == "god" {
			hasGod = true
		} else if lowerCol == "kar" {
			hasKar = true
		}
	}
	return hasGod, hasKar
}

// AddGodKarConditions adds god and kar conditions to WHERE clause
func (qb *QueryBuilder) AddGodKarConditions(hasGod, hasKar bool, god, kar int) {
	entityName := ""
	if qb.entityType != nil {
		entityName = fmt.Sprintf("%s.", strings.ToLower(qb.entityType.Name()))
	}

	if hasGod {
		qb.whereClause.WriteString(fmt.Sprintf(" AND %sgod = $%d ", entityName, qb.paramCount))
		qb.args = append(qb.args, god)
		qb.paramCount++
	}

	if hasKar {
		qb.whereClause.WriteString(fmt.Sprintf(" AND %skar = $%d ", entityName, qb.paramCount))
		qb.args = append(qb.args, kar)
		qb.paramCount++
	}
}

// AddSearchConditions adds search/filter conditions to WHERE clause
// Supports table-qualified column names when Field contains "table.column" format
func (qb *QueryBuilder) AddSearchConditions(fields []domain.Fields, searchParams ...string) {
	if len(searchParams) == 0 || len(searchParams[0]) == 0 {
		return
	}

	if qb.entityType == nil {
		return
	}

	likeString := " AND ( "
	hasCondition := false

	for i := 0; i < qb.entityType.NumField(); i++ {
		field := qb.entityType.Field(i)
		dbCol := field.Tag.Get("db")

		// Skip fields not in tableFields or marked to skip in search
		found := false
		qualifiedCol := ""
		for _, tblField := range fields {
			if tblField.SkipInSearch {
				continue
			}
			if strings.EqualFold(tblField.Name, field.Name) && strings.ToLower(field.Name) != "god" && strings.ToLower(field.Name) != "kar" {
				found = true
				// in table definition fields
				// Use Field property if it contains table qualification (e.g., "partneri.sifra")
				if tblField.Field != "" && strings.Contains(tblField.Field, ".") {
					qualifiedCol = tblField.Field
				} else {
					qualifiedCol = strings.ToLower(dbCol)
				}
				break
			}
		}

		if !found {
			continue
		}

		switch field.Type.Name() {
		case "int", "int64", "int32", "float32", "float64", "string", "bool":
			likeString += fmt.Sprintf(" OR (%s::TEXT ILIKE $%d)", qualifiedCol, qb.paramCount)
			qb.args = append(qb.args, fmt.Sprintf("%%%s%%", searchParams[0]))
			qb.paramCount++
			hasCondition = true
		}
	}

	if hasCondition {
		likeString += " )"
		likeString = strings.ReplaceAll(likeString, "AND (  OR", "AND ( ")
		qb.whereClause.WriteString(likeString)
	}
}

// GetWhereClause returns the WHERE clause as a string
func (qb *QueryBuilder) GetWhereClause() string {
	return qb.whereClause.String()
}

// ==================== UNION Query Support ====================

// UnionQueryBuilder handles UNION queries combining multiple QueryBuilders
type UnionQueryBuilder struct {
	queries   []*QueryBuilder
	unionType string // "UNION" or "UNION ALL"
	orderBy   string
	limit     string
	offset    string
}

// NewUnionQueryBuilder creates a new UNION query builder
func NewUnionQueryBuilder(unionType string) *UnionQueryBuilder {
	if unionType != "UNION" {
		unionType = "UNION ALL"
	}
	return &UnionQueryBuilder{
		queries:   make([]*QueryBuilder, 0),
		unionType: unionType,
	}
}

// AddQuery adds a query builder to the UNION
func (uqb *UnionQueryBuilder) AddQuery(qb *QueryBuilder) *UnionQueryBuilder {
	uqb.queries = append(uqb.queries, qb)
	return uqb
}

// AddOrderBy adds ORDER BY clause to the UNION result
func (uqb *UnionQueryBuilder) AddOrderBy(orderBy string) *UnionQueryBuilder {
	uqb.orderBy = orderBy
	return uqb
}

// SetLimit sets LIMIT for the UNION result
func (uqb *UnionQueryBuilder) SetLimit(limit int) *UnionQueryBuilder {
	if limit > 0 {
		uqb.limit = fmt.Sprintf("LIMIT %d", limit)
	}
	return uqb
}

// SetOffset sets OFFSET for the UNION result
func (uqb *UnionQueryBuilder) SetOffset(offset int) *UnionQueryBuilder {
	if offset > 0 {
		uqb.offset = fmt.Sprintf("OFFSET %d", offset)
	}
	return uqb
}

// adjustParameterPlaceholders replaces all $N placeholders with offset applied
// Manually parses and replaces parameters without regex
func adjustParameterPlaceholders(query string, offset int) string {
	if offset == 0 {
		return query
	}

	var result strings.Builder
	i := 0
	for i < len(query) {
		if query[i] == '$' && i+1 < len(query) && query[i+1] >= '0' && query[i+1] <= '9' {
			// Found a $ followed by a digit
			j := i + 1
			for j < len(query) && query[j] >= '0' && query[j] <= '9' {
				j++
			}
			// Extract and convert number
			numStr := query[i+1 : j]
			oldNum := 0
			fmt.Sscanf(numStr, "%d", &oldNum)
			newNum := oldNum + offset
			result.WriteString(fmt.Sprintf("$%d", newNum))
			i = j
		} else {
			result.WriteByte(query[i])
			i++
		}
	}
	return result.String()
}

// Build constructs the final UNION query with proper parameter placeholders
func (uqb *UnionQueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder
	var allArgs []interface{}
	var paramOffset int

	for i, qb := range uqb.queries {
		if i > 0 {
			query.WriteString(" " + uqb.unionType + " ")
		}

		// Build individual query
		q, args := qb.Build()

		// Adjust parameter placeholders for union
		// Use a single regex pass to replace all $N placeholders
		adjustedQ := adjustParameterPlaceholders(q, paramOffset)

		query.WriteString("(" + adjustedQ + ")")
		allArgs = append(allArgs, args...)
		paramOffset += len(args)
	}

	// Add final clauses
	if uqb.orderBy != "" {
		query.WriteString(" ORDER BY " + uqb.orderBy)
	}
	if uqb.limit != "" {
		query.WriteString(" " + uqb.limit)
	}
	if uqb.offset != "" {
		query.WriteString(" " + uqb.offset)
	}

	return query.String(), allArgs
}

// GetArgs returns all arguments from the union
func (uqb *UnionQueryBuilder) GetArgs() []interface{} {
	return nil // Args are included in Build() result
}

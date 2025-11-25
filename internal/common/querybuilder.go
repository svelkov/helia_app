package common

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	baseQuery   string
	whereClause strings.Builder
	args        []interface{}
	paramCount  int
	joins       []string
	orderBy     string
	groupBy     string
	limit       string
	offset      string
}

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	qb := &QueryBuilder{
		baseQuery:  baseQuery,
		paramCount: 1,
	}
	qb.whereClause.WriteString(" WHERE 1 = 1 ")
	return qb
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

// AddGroupBy adds GROUP BY clause
func (qb *QueryBuilder) AddGroupBy(groupBy string) *QueryBuilder {
	qb.groupBy = groupBy
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

// AddCustomCondition for complex conditions
func (qb *QueryBuilder) AddCustomCondition(condition string, values ...interface{}) *QueryBuilder {
	qb.whereClause.WriteString(" AND " + condition)
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

	// Add ORDER BY
	if qb.orderBy != "" {
		query.WriteString(" ORDER BY " + qb.orderBy)
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

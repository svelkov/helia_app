package querybuilder

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Union combines several Builders with UNION or UNION ALL.
// Placeholders are numbered once over the whole statement, so no renumbering is needed.
type Union struct {
	keyword string
	queries []*Builder
	orderBy []string
	limit   int
	offset  int
	errs    []error
}

// NewUnion creates a UNION (duplicates removed).
func NewUnion() *Union { return &Union{keyword: "UNION"} }

// NewUnionAll creates a UNION ALL.
func NewUnionAll() *Union { return &Union{keyword: "UNION ALL"} }

// Add appends a query to the union.
func (u *Union) Add(b *Builder) *Union {
	u.queries = append(u.queries, b)
	return u
}

// OrderBy sets ORDER BY for the combined result (validated items, or column positions).
func (u *Union) OrderBy(expr string) *Union {
	if strings.TrimSpace(expr) == "" {
		return u
	}
	items, err := parseOrder(expr)
	if err != nil {
		u.errs = append(u.errs, err)
		return u
	}
	u.orderBy = append(u.orderBy, items...)
	return u
}

// Limit sets LIMIT for the combined result.
func (u *Union) Limit(n int) *Union {
	if n > 0 {
		u.limit = n
	}
	return u
}

// Offset sets OFFSET for the combined result.
func (u *Union) Offset(n int) *Union {
	if n > 0 {
		u.offset = n
	}
	return u
}

// Build returns the final SQL and arguments.
func (u *Union) Build() (string, []any, error) {
	if len(u.errs) > 0 {
		return "", nil, errors.Join(u.errs...)
	}
	if len(u.queries) == 0 {
		return "", nil, errors.New("querybuilder: union has no queries")
	}
	var sb strings.Builder
	var args []any
	for i, q := range u.queries {
		s, a, err := q.render()
		if err != nil {
			return "", nil, fmt.Errorf("union query %d: %w", i, err)
		}
		if i > 0 {
			sb.WriteString(" " + u.keyword + " ")
		}
		sb.WriteString("(" + s + ")")
		args = append(args, a...)
	}
	if len(u.orderBy) > 0 {
		sb.WriteString(" ORDER BY " + strings.Join(u.orderBy, ", "))
	}
	if u.limit > 0 {
		sb.WriteString(" LIMIT " + strconv.Itoa(u.limit))
	}
	if u.offset > 0 {
		sb.WriteString(" OFFSET " + strconv.Itoa(u.offset))
	}
	return finalize(sb.String(), args)
}

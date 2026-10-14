// Package querybuilder builds parameterized PostgreSQL statements.
//
// Rules the builder enforces:
//   - Values are always passed as arguments, never concatenated into SQL.
//   - Raw SQL fragments (Where, AddJoin, Having, Set, From, WithCTE, OrderByExpr, GroupByExpr)
//     use "?" as the placeholder ("??" is a literal question mark). Placeholders are numbered
//     ($1, $2, ...) once, in Build(), so the order in which you call methods never matters.
//   - Identifiers (columns, operators, OrderBy / GroupBy items) are validated; anything
//     that is not a plain identifier makes Build() return an error. OrderByExpr and
//     GroupByExpr are the raw-fragment escape hatch for static expressions.
//   - Raw fragments must be static SQL written by the developer, never request input.
package querybuilder

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const identPattern = `[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*`

var (
	identRe     = regexp.MustCompile(`^` + identPattern + `$`)
	orderItemRe = regexp.MustCompile(`(?i)^(` + identPattern + `|\d+)( (ASC|DESC))?( NULLS (FIRST|LAST))?$`)

	operators = map[string]string{
		"=": "=", "<>": "<>", "!=": "<>", "<": "<", "<=": "<=", ">": ">", ">=": ">=",
		"like": "LIKE", "ilike": "ILIKE",
	}
)

// IsIdent reports whether s is a plain, optionally dot-qualified SQL identifier.
func IsIdent(s string) bool { return identRe.MatchString(s) }

// QuoteIdent quotes an identifier, splitting on dots: a.b -> "a"."b".
func QuoteIdent(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		parts[i] = `"` + strings.ReplaceAll(p, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

// Args converts a typed slice to []any, e.g. for AddIn.
func Args[T any](v []T) []any {
	out := make([]any, len(v))
	for i := range v {
		out[i] = v[i]
	}
	return out
}

type part struct {
	sql  string
	args []any
}

type cte struct {
	name string
	q    part
}

// Cond is a raw condition with "?" placeholders, used by WhereAny.
type Cond struct {
	SQL  string
	Args []any
}

// C creates a Cond.
func C(sql string, args ...any) Cond { return Cond{SQL: sql, Args: args} }

// Builder builds a single SELECT / UPDATE statement. Not safe for concurrent use.
type Builder struct {
	base      part
	ctes      []cte
	sets      []part
	from      part
	joins     []part
	where     []part
	groupBy   []part
	having    []part
	orderBy   []part
	returning []string
	limit     int
	forUpdate bool
	offset    int
	errs      []error
}

// New starts a statement from a static base such as "SELECT a, b FROM t".
// Do not include WHERE; it is added automatically when conditions exist.
func New(base string, args ...any) *Builder {
	return &Builder{base: part{sql: base, args: args}}
}

// NewUpdate starts an UPDATE of the given table.
func NewUpdate(table string) *Builder {
	return &Builder{base: part{sql: "UPDATE " + QuoteIdent(table)}}
}

// Clone returns an independent copy.
func (b *Builder) Clone() *Builder {
	c := *b
	c.ctes = append([]cte(nil), b.ctes...)
	c.sets = append([]part(nil), b.sets...)
	c.joins = append([]part(nil), b.joins...)
	c.where = append([]part(nil), b.where...)
	c.groupBy = append([]part(nil), b.groupBy...)
	c.having = append([]part(nil), b.having...)
	c.orderBy = append([]part(nil), b.orderBy...)
	c.returning = append([]string(nil), b.returning...)
	c.errs = append([]error(nil), b.errs...)
	return &c
}

func (b *Builder) fail(format string, a ...any) *Builder {
	b.errs = append(b.errs, fmt.Errorf("querybuilder: "+format, a...))
	return b
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

// ---------------------------------------------------------------- conditions

// Where adds a raw static condition with ? placeholders. It is wrapped in parentheses,
// so an OR inside it cannot escape other conditions (e.g. a tenant filter).
func (b *Builder) Where(cond string, args ...any) *Builder {
	if strings.TrimSpace(cond) == "" {
		return b
	}
	b.where = append(b.where, part{sql: "(" + cond + ")", args: args})
	return b
}

// WhereAny adds (c1 OR c2 OR ...), combined with the other conditions by AND.
func (b *Builder) WhereAny(conds ...Cond) *Builder {
	var sqls []string
	var args []any
	for _, c := range conds {
		if strings.TrimSpace(c.SQL) == "" {
			continue
		}
		sqls = append(sqls, c.SQL)
		args = append(args, c.Args...)
	}
	if len(sqls) == 0 {
		return b
	}
	b.where = append(b.where, part{sql: "(" + strings.Join(sqls, " OR ") + ")", args: args})
	return b
}

// AddCondition adds "field <operator> ?". The condition is skipped when value is nil or "".
// Allowed operators: = <> != < <= > >= like ilike.
func (b *Builder) AddCondition(field string, value any, operator string) *Builder {
	if !IsIdent(field) {
		return b.fail("invalid column %q", field)
	}
	op, ok := operators[strings.ToLower(strings.TrimSpace(operator))]
	if !ok {
		return b.fail("operator %q is not allowed", operator)
	}
	if isEmpty(value) {
		return b
	}
	b.where = append(b.where, part{sql: field + " " + op + " ?", args: []any{value}})
	return b
}

// AddEqual adds "field = ?" (skipped for nil or "").
func (b *Builder) AddEqual(field string, value any) *Builder {
	return b.AddCondition(field, value, "=")
}

func (b *Builder) addLike(field string, value any, prefix, suffix string) *Builder {
	if !IsIdent(field) {
		return b.fail("invalid column %q", field)
	}
	if isEmpty(value) {
		return b
	}
	arg := prefix + escapeLike(fmt.Sprint(value)) + suffix
	b.where = append(b.where, part{sql: field + " ILIKE ?", args: []any{arg}})
	return b
}

// AddLike adds a contains match: field ILIKE '%value%' (wildcards in value are escaped).
func (b *Builder) AddLike(field string, value any) *Builder { return b.addLike(field, value, "%", "%") }

// AddLikeBegin adds a prefix match: field ILIKE 'value%'.
func (b *Builder) AddLikeBegin(field string, value any) *Builder {
	return b.addLike(field, value, "", "%")
}

// AddIn adds "field IN (?, ?, ...)". An empty list adds no condition.
func (b *Builder) AddIn(field string, values []any) *Builder {
	if !IsIdent(field) {
		return b.fail("invalid column %q", field)
	}
	if len(values) == 0 {
		return b
	}
	ph := strings.TrimSuffix(strings.Repeat("?, ", len(values)), ", ")
	b.where = append(b.where, part{sql: field + " IN (" + ph + ")", args: values})
	return b
}

// AddSearch adds (c1::text ILIKE ? OR c2::text ILIKE ? ...) for a free-text search.
func (b *Builder) AddSearch(columns []string, text string) *Builder {
	for _, c := range columns {
		if !IsIdent(c) {
			return b.fail("invalid search column %q", c)
		}
	}
	if text == "" || len(columns) == 0 {
		return b
	}
	pattern := "%" + escapeLike(text) + "%"
	sqls := make([]string, len(columns))
	args := make([]any, len(columns))
	for i, c := range columns {
		sqls[i] = c + "::text ILIKE ?"
		args[i] = pattern
	}
	b.where = append(b.where, part{sql: "(" + strings.Join(sqls, " OR ") + ")", args: args})
	return b
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func escapeLike(s string) string { return likeEscaper.Replace(s) }

// ---------------------------------------------------------------- clauses

// WithCTE adds a WITH name AS (query). The query is static SQL with ? placeholders.
func (b *Builder) WithCTE(name, query string, args ...any) *Builder {
	if !IsIdent(name) || strings.Contains(name, ".") {
		return b.fail("invalid CTE name %q", name)
	}
	b.ctes = append(b.ctes, cte{name: name, q: part{sql: query, args: args}})
	return b
}

// WithCTEFrom adds WITH name AS (sub), where sub is another builder with its own conditions and
// arguments. The SQL and the arguments of sub are taken together, so they can not be bound to the
// wrong placeholders; they are numbered with the outer statement. sub is read when it is added:
// later changes to sub are not seen. An error of sub is returned by Build.
func (b *Builder) WithCTEFrom(name string, sub *Builder) *Builder {
	q, args, err := sub.render()
	if err != nil {
		return b.fail("CTE %q: %v", name, err)
	}
	return b.WithCTE(name, q, args...)
}

// AddJoin adds a static JOIN fragment, e.g. "JOIN partneri p ON p.id = x.partner_id".
func (b *Builder) AddJoin(join string, args ...any) *Builder {
	b.joins = append(b.joins, part{sql: join, args: args})
	return b
}

// GroupBy sets GROUP BY columns (validated identifiers).
func (b *Builder) GroupBy(cols ...string) *Builder {
	for _, c := range cols {
		if !IsIdent(c) {
			return b.fail("invalid GROUP BY column %q", c)
		}
	}
	for _, c := range cols {
		b.groupBy = append(b.groupBy, part{sql: c})
	}
	return b
}

// GroupByExpr adds a static GROUP BY expression such as "EXTRACT(MONTH FROM danal)" or
// "LEFT(konto, ?)". Like Where it is raw SQL: write it in code, never pass request input.
func (b *Builder) GroupByExpr(expr string, args ...any) *Builder {
	if strings.TrimSpace(expr) != "" {
		b.groupBy = append(b.groupBy, part{sql: expr, args: args})
	}
	return b
}

// Having adds a static HAVING condition (multiple calls are joined with AND).
func (b *Builder) Having(cond string, args ...any) *Builder {
	if strings.TrimSpace(cond) != "" {
		b.having = append(b.having, part{sql: "(" + cond + ")", args: args})
	}
	return b
}

func parseOrder(expr string) ([]string, error) {
	var out []string
	for _, item := range strings.Split(expr, ",") {
		item = strings.Join(strings.Fields(item), " ")
		if !orderItemRe.MatchString(item) {
			return nil, fmt.Errorf("querybuilder: invalid ORDER BY item %q", item)
		}
		out = append(out, item)
	}
	return out, nil
}

// OrderBy adds static ORDER BY items such as "name DESC, id". Each item is validated.
func (b *Builder) OrderBy(expr string) *Builder {
	if strings.TrimSpace(expr) == "" {
		return b
	}
	items, err := parseOrder(expr)
	if err != nil {
		b.errs = append(b.errs, err)
		return b
	}
	for _, it := range items {
		b.orderBy = append(b.orderBy, part{sql: it})
	}
	return b
}

// OrderByExpr adds a static ORDER BY expression such as "fpro.sifra::numeric" or
// `sort_key COLLATE "C" NULLS LAST`. Like Where it is raw SQL: write it in code, never pass
// request input (use OrderByKey for that). Items from OrderBy, OrderByExpr and OrderByKey
// keep their call order.
func (b *Builder) OrderByExpr(expr string, args ...any) *Builder {
	if strings.TrimSpace(expr) != "" {
		b.orderBy = append(b.orderBy, part{sql: expr, args: args})
	}
	return b
}

// OrderByKey sorts by a column looked up in a whitelist. Use it for request input:
// key is the sort key from the client, allowed maps keys to real columns. An unknown key
// is ignored; dir is normalized to ASC or DESC.
func (b *Builder) OrderByKey(key, dir string, allowed map[string]string) *Builder {
	col, ok := allowed[key]
	if !ok {
		return b
	}
	if !IsIdent(col) {
		return b.fail("invalid whitelisted column %q", col)
	}
	d := "ASC"
	if strings.EqualFold(strings.TrimSpace(dir), "desc") {
		d = "DESC"
	}
	b.orderBy = append(b.orderBy, part{sql: col + " " + d})
	return b
}

// ForUpdate adds FOR UPDATE at the end of a SELECT (row locks for the rest of the transaction).
func (b *Builder) ForUpdate() *Builder {
	b.forUpdate = true
	return b
}

// Limit sets LIMIT (ignored when n <= 0).
func (b *Builder) Limit(n int) *Builder {
	if n > 0 {
		b.limit = n
	}
	return b
}

// Offset sets OFFSET (ignored when n <= 0).
func (b *Builder) Offset(n int) *Builder {
	if n > 0 {
		b.offset = n
	}
	return b
}

// ---------------------------------------------------------------- UPDATE helpers

// Set adds a static SET fragment such as "name = ?".
func (b *Builder) Set(clause string, args ...any) *Builder {
	b.sets = append(b.sets, part{sql: clause, args: args})
	return b
}

// SetValue adds "column = ?".
func (b *Builder) SetValue(column string, value any) *Builder {
	if !IsIdent(column) {
		return b.fail("invalid column %q", column)
	}
	return b.Set(column+" = ?", value)
}

// From adds a static FROM fragment for UPDATE ... FROM.
func (b *Builder) From(clause string, args ...any) *Builder {
	b.from = part{sql: clause, args: args}
	return b
}

// Returning sets RETURNING columns ("*" allowed).
func (b *Builder) Returning(cols ...string) *Builder {
	for _, c := range cols {
		if c != "*" && !IsIdent(c) {
			return b.fail("invalid RETURNING column %q", c)
		}
	}
	b.returning = cols
	return b
}

// ---------------------------------------------------------------- build

func (b *Builder) render() (string, []any, error) {
	if len(b.errs) > 0 {
		return "", nil, errors.Join(b.errs...)
	}
	if strings.TrimSpace(b.base.sql) == "" {
		return "", nil, errors.New("querybuilder: empty base query")
	}
	var sb strings.Builder
	var args []any
	add := func(p part) {
		sb.WriteString(p.sql)
		args = append(args, p.args...)
	}

	if len(b.ctes) > 0 {
		sb.WriteString("WITH ")
		for i, c := range b.ctes {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(c.name + " AS (")
			add(c.q)
			sb.WriteString(")")
		}
		sb.WriteString(" ")
	}
	add(b.base)
	if len(b.sets) > 0 {
		sb.WriteString(" SET ")
		for i, s := range b.sets {
			if i > 0 {
				sb.WriteString(", ")
			}
			add(s)
		}
	}
	if b.from.sql != "" {
		sb.WriteString(" FROM ")
		add(b.from)
	}
	for _, j := range b.joins {
		sb.WriteString(" ")
		add(j)
	}
	if len(b.where) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range b.where {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			add(w)
		}
	}
	if len(b.groupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		for i, g := range b.groupBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			add(g)
		}
	}
	if len(b.having) > 0 {
		sb.WriteString(" HAVING ")
		for i, h := range b.having {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			add(h)
		}
	}
	if len(b.orderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		for i, o := range b.orderBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			add(o)
		}
	}
	if b.limit > 0 {
		sb.WriteString(" LIMIT " + strconv.Itoa(b.limit))
	}
	if b.offset > 0 {
		sb.WriteString(" OFFSET " + strconv.Itoa(b.offset))
	}
	if b.forUpdate {
		sb.WriteString(" FOR UPDATE")
	}
	if len(b.returning) > 0 {
		sb.WriteString(" RETURNING " + strings.Join(b.returning, ", "))
	}
	return sb.String(), args, nil
}

// Build returns the final SQL with $1, $2, ... placeholders and its arguments.
// It fails if any builder call was invalid or placeholders and arguments do not match.
func (b *Builder) Build() (string, []any, error) {
	q, args, err := b.render()
	if err != nil {
		return "", nil, err
	}
	return finalize(q, args)
}

// BuildCount returns SELECT count(*) over this query, ignoring ORDER BY, LIMIT and OFFSET.
// It does not modify the builder, so the same filters can be used for the page query.
func (b *Builder) BuildCount() (string, []any, error) {
	c := b.Clone()
	c.orderBy, c.limit, c.offset, c.forUpdate = nil, 0, 0, false
	q, args, err := c.render()
	if err != nil {
		return "", nil, err
	}
	return finalize("SELECT count(*) FROM ("+q+") AS cnt", args)
}

func finalize(q string, args []any) (string, []any, error) {
	out, n := numberPlaceholders(q)
	if n != len(args) {
		return "", nil, fmt.Errorf("querybuilder: %d placeholders but %d arguments", n, len(args))
	}
	return out, args, nil
}

// numberPlaceholders turns ? into $1..$n, "??" into a literal ?, and ignores quoted text.
func numberPlaceholders(q string) (string, int) {
	var sb strings.Builder
	n := 0
	inSingle, inDouble := false, false
	for i := 0; i < len(q); i++ {
		c := q[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
			sb.WriteByte(c)
		case c == '"' && !inSingle:
			inDouble = !inDouble
			sb.WriteByte(c)
		case c == '?' && !inSingle && !inDouble:
			if i+1 < len(q) && q[i+1] == '?' {
				sb.WriteByte('?')
				i++
			} else {
				n++
				sb.WriteString("$" + strconv.Itoa(n))
			}
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String(), n
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
)

// Errors returned by the repository.
var (
	// ErrNoSession means the context carries no user session, but the operation needs one
	// (audit columns on insert/update, or the god/kar tenant scope).
	ErrNoSession = errors.New("repository: no user session in context")
	// ErrNotFound is returned by Update and Delete when no row matched. It wraps sql.ErrNoRows,
	// so errors.Is(err, sql.ErrNoRows) also works.
	ErrNotFound = fmt.Errorf("repository: record not found: %w", sql.ErrNoRows)
)

// now is replaced in tests.
var now = time.Now

// entitySession holds what the audit and tenant columns need.
type entitySession struct {
	UserName string
	God      int
	Kar      int
}

// sessionFromContext reads the session from the context. It is a variable so tests can
// replace it.
var sessionFromContext = func(ctx context.Context) (entitySession, error) {
	s := domain.GetSessionFromStdContext(ctx)
	if s == nil {
		return entitySession{}, ErrNoSession
	}
	return entitySession{UserName: s.UserName, God: s.SelectedGod, Kar: s.SelectedKar}, nil
}

type column struct {
	goName   string
	db       string
	typ      reflect.Type
	noSelect bool // tag addupdate:"false"
}

// entityTable describes one table and its entity type. It is built once per repository and is
// read-only afterwards, so it is safe for concurrent use.
type entityTable struct {
	name    string
	columns []column
	hasGod  bool
	hasKar  bool
}

var timeType = reflect.TypeOf(time.Time{})

// newEntityTable validates the table name and reads the `db` tags of entity (struct or pointer
// to struct) once. Embedded and nested structs are flattened.
func newEntityTable(name string, entity any) (*entityTable, error) {
	if !querybuilder.IsIdent(name) {
		return nil, fmt.Errorf("invalid table name %q", name)
	}
	t := reflect.TypeOf(entity)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("entity for table %q must be a struct or pointer to struct", name)
	}
	tb := &entityTable{name: name}
	tb.collect(t)
	for _, c := range tb.columns {
		if !querybuilder.IsIdent(c.db) || strings.Contains(c.db, ".") {
			return nil, fmt.Errorf("table %q: invalid db tag %q on field %s", name, c.db, c.goName)
		}
		switch strings.ToLower(c.db) {
		case "god":
			tb.hasGod = true
		case "kar":
			tb.hasKar = true
		}
	}
	return tb, nil
}

func (t *entityTable) collect(typ reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		ft := f.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && ft != timeType && f.Tag.Get("db") == "" {
			t.collect(ft)
			continue
		}
		db := f.Tag.Get("db")
		if db == "" || db == "-" {
			continue
		}
		t.columns = append(t.columns, column{
			goName: f.Name, db: db, typ: f.Type, noSelect: f.Tag.Get("addupdate") == "false",
		})
	}
}

func reserved(name string) bool {
	switch strings.ToLower(name) {
	case "god", "kar", "xopunos", "xdatunosa", "xopizmene", "xdatizmene":
		return true
	}
	return false
}

func fieldValue(f domain.Fields) any {
	if f.IsNull {
		return nil
	}
	return f.Value
}

func colName(f domain.Fields) (string, error) {
	n := strings.ToLower(f.Name)
	if !querybuilder.IsIdent(n) || strings.Contains(n, ".") {
		return "", fmt.Errorf("invalid column name %q", f.Name)
	}
	return n, nil
}

func checkID(idField string) error {
	if !querybuilder.IsIdent(idField) || strings.Contains(idField, ".") {
		return fmt.Errorf("invalid id column %q", idField)
	}
	return nil
}

// insert builds INSERT ... RETURNING idField. god, kar and the audit columns come from the
// session; the same names in fields are ignored.
func (t *entityTable) insert(s entitySession, fields []domain.Fields, idField string) (string, []any, error) {
	if err := checkID(idField); err != nil {
		return "", nil, err
	}
	var cols []string
	var vals []any
	add := func(c string, v any) {
		cols = append(cols, querybuilder.QuoteIdent(c))
		vals = append(vals, v)
	}
	if t.hasGod {
		add("god", s.God)
	}
	if t.hasKar {
		add("kar", s.Kar)
	}
	add("xopunos", s.UserName)
	add("xdatunosa", now())
	for _, f := range fields {
		if reserved(f.Name) {
			continue
		}
		n, err := colName(f)
		if err != nil {
			return "", nil, err
		}
		add(n, fieldValue(f))
	}
	ph := make([]string, len(vals))
	for i := range ph {
		ph[i] = fmt.Sprintf("$%d", i+1)
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
		querybuilder.QuoteIdent(t.name), strings.Join(cols, ", "), strings.Join(ph, ", "),
		querybuilder.QuoteIdent(idField))
	return q, vals, nil
}

// updateBuilder builds UPDATE ... WHERE id = ? and fills the audit columns. The id condition
// uses Where (not AddEqual), so an empty id can never silently remove the condition.
func (t *entityTable) updateBuilder(s entitySession, fields []domain.Fields, idField string, idValue any) (*querybuilder.Builder, error) {
	if err := checkID(idField); err != nil {
		return nil, err
	}
	b := querybuilder.NewUpdate(t.name)
	for _, f := range fields {
		if reserved(f.Name) {
			continue
		}
		n, err := colName(f)
		if err != nil {
			return nil, err
		}
		b.Set(querybuilder.QuoteIdent(n)+" = ?", fieldValue(f))
	}
	b.Set(`"xopizmene" = ?`, s.UserName)
	b.Set(`"xdatizmene" = ?`, now())
	return b.Where(querybuilder.QuoteIdent(idField)+" = ?", idValue), nil
}

// selectByIDBuilder builds SELECT of all selectable columns WHERE id = ?. Non-pointer columns
// are wrapped in COALESCE so NULLs scan into plain Go values; pointer columns stay bare.
func (t *entityTable) selectByIDBuilder(idField string, idValue any) (*querybuilder.Builder, error) {
	if err := checkID(idField); err != nil {
		return nil, err
	}
	var cols []string
	for _, c := range t.columns {
		if !c.noSelect {
			cols = append(cols, coalesceExpr(c.db, c.typ))
		}
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q: entity has no db columns", t.name)
	}
	b := querybuilder.New(fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "), querybuilder.QuoteIdent(t.name)))
	return b.Where(querybuilder.QuoteIdent(idField)+" = ?", idValue), nil
}

func coalesceExpr(col string, typ reflect.Type) string {
	q := querybuilder.QuoteIdent(col)
	var zero string
	switch typ.Kind() {
	case reflect.String:
		zero = "''"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		zero = "0"
	case reflect.Bool:
		zero = "false"
	default: // pointers, time.Time, everything else
		return q
	}
	return fmt.Sprintf("COALESCE(%s, %s) AS %s", q, zero, q)
}

// selectAllBuilder starts a SELECT of idField, god/kar (if present) and the entity columns
// named in fields. Add filters, sorting and paging on the returned builder.
func (t *entityTable) selectAllBuilder(fields []domain.Fields, idField string) (*querybuilder.Builder, error) {
	if err := checkID(idField); err != nil {
		return nil, err
	}
	cols := []string{querybuilder.QuoteIdent(idField)}
	seen := map[string]bool{strings.ToLower(idField): true}
	for _, c := range t.columns {
		l := strings.ToLower(c.db)
		if seen[l] {
			continue
		}
		if l == "god" || l == "kar" {
			cols = append(cols, querybuilder.QuoteIdent(l))
			seen[l] = true
			continue
		}
		for _, f := range fields {
			if strings.EqualFold(f.Name, c.db) {
				cols = append(cols, querybuilder.QuoteIdent(l))
				seen[l] = true
				break
			}
		}
	}
	return querybuilder.New(fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "),
		querybuilder.QuoteIdent(t.name))), nil
}

func (t *entityTable) countBuilder() *querybuilder.Builder {
	return querybuilder.New("SELECT count(*) FROM " + querybuilder.QuoteIdent(t.name))
}

func (t *entityTable) deleteBuilder(idField string, idValue any) (*querybuilder.Builder, error) {
	if err := checkID(idField); err != nil {
		return nil, err
	}
	b := querybuilder.New("DELETE FROM " + querybuilder.QuoteIdent(t.name))
	return b.Where(querybuilder.QuoteIdent(idField)+" = ?", idValue), nil
}

// scopeTenant restricts b to the session's god/kar. It uses Where, never AddEqual, so the
// condition can not be skipped for an empty value. qual is a ready-made prefix such as
// `"partneri".` or `p.`, or "".
func (t *entityTable) scopeTenant(b *querybuilder.Builder, qual string, s entitySession) {
	if t.hasGod {
		b.Where(qual+"god = ?", s.God)
	}
	if t.hasKar {
		b.Where(qual+"kar = ?", s.Kar)
	}
}

// searchColumns returns the columns for a free-text search over the entity fields listed in
// fields (matched by Go field name). god/kar and SkipInSearch fields are skipped; pointer
// columns are included; time.Time is not searchable.
func (t *entityTable) searchColumns(fields []domain.Fields) []string {
	var out []string
	for _, c := range t.columns {
		l := strings.ToLower(c.db)
		if l == "god" || l == "kar" {
			continue
		}
		typ := c.typ
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		switch typ.Kind() {
		case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		default:
			continue
		}
		for _, f := range fields {
			if f.SkipInSearch || !strings.EqualFold(f.Name, c.goName) {
				continue
			}
			if strings.Contains(f.Field, ".") {
				out = append(out, f.Field)
			} else {
				out = append(out, l)
			}
			break
		}
	}
	return out
}

// sortColumns is the whitelist for ORDER BY: lower-cased db column names of the entity.
func (t *entityTable) sortColumns() map[string]string {
	m := make(map[string]string, len(t.columns)+1)
	for _, c := range t.columns {
		l := strings.ToLower(c.db)
		m[l] = l
	}
	return m
}

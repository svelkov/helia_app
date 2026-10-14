package querybuilder

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	allowed := map[string]string{"name": "c.name", "email": "c.email"}
	tests := []struct {
		name     string
		b        *Builder
		wantSQL  string
		wantArgs []any
	}{
		{"plain select", New("SELECT id FROM t"), "SELECT id FROM t", nil},
		{"equal skips empty", New("SELECT id FROM t").AddEqual("a", 1).AddEqual("b", "").AddEqual("c", nil),
			"SELECT id FROM t WHERE a = $1", []any{1}},
		{"placeholders follow call order across methods",
			New("SELECT id FROM t").AddEqual("a", 1).Where("b > ? AND c < ?", 2, 3).AddEqual("d", 4),
			"SELECT id FROM t WHERE a = $1 AND (b > $2 AND c < $3) AND d = $4", []any{1, 2, 3, 4}},
		{"base args come first", New("SELECT id, ? AS k FROM t", 9).AddEqual("a", 1),
			"SELECT id, $1 AS k FROM t WHERE a = $2", []any{9, 1}},
		{"or is grouped and cannot escape the tenant filter",
			New("SELECT id FROM t").AddEqual("god", 1).AddEqual("kar", 2).WhereAny(C("x = ?", 3), C("y = ?", 4)),
			"SELECT id FROM t WHERE god = $1 AND kar = $2 AND (x = $3 OR y = $4)", []any{1, 2, 3, 4}},
		{"where with inner OR is parenthesized",
			New("SELECT id FROM t").AddEqual("god", 1).Where("a = ? OR b = ?", 2, 3),
			"SELECT id FROM t WHERE god = $1 AND (a = $2 OR b = $3)", []any{1, 2, 3}},
		{"like escapes wildcards", New("SELECT id FROM t").AddLike("name", "50%_x"),
			"SELECT id FROM t WHERE name ILIKE $1", []any{`%50\%\_x%`}},
		{"like begin accepts non strings", New("SELECT id FROM t").AddLikeBegin("code", 42),
			"SELECT id FROM t WHERE code ILIKE $1", []any{"42%"}},
		{"in", New("SELECT id FROM t").AddIn("id", Args([]int{1, 2, 3})),
			"SELECT id FROM t WHERE id IN ($1, $2, $3)", []any{1, 2, 3}},
		{"empty in adds nothing", New("SELECT id FROM t").AddIn("id", nil), "SELECT id FROM t", nil},
		{"search", New("SELECT id FROM t").AddSearch([]string{"a", "p.b"}, "x"),
			"SELECT id FROM t WHERE (a::text ILIKE $1 OR p.b::text ILIKE $2)", []any{"%x%", "%x%"}},
		{"search with no columns or text adds nothing", New("SELECT id FROM t").AddSearch(nil, "x").AddSearch([]string{"a"}, ""),
			"SELECT id FROM t", nil},
		{"order by whitelist asc default", New("SELECT id FROM t").OrderByKey("name", "bogus", allowed),
			"SELECT id FROM t ORDER BY c.name ASC", nil},
		{"order by whitelist desc", New("SELECT id FROM t").OrderByKey("email", "DESC", allowed),
			"SELECT id FROM t ORDER BY c.email DESC", nil},
		{"order by unknown key ignored", New("SELECT id FROM t").OrderByKey("1; DROP TABLE t", "asc", allowed),
			"SELECT id FROM t", nil},
		{"order by nulls last", New("SELECT id FROM t").OrderBy("name desc nulls last, id"),
			"SELECT id FROM t ORDER BY name desc nulls last, id", nil},
		{"static order by multi", New("SELECT id FROM t").OrderBy("name DESC, id"),
			"SELECT id FROM t ORDER BY name DESC, id", nil},
		{"group by having", New("SELECT a, count(*) FROM t").GroupBy("a").Having("count(*) > ?", 1),
			"SELECT a, count(*) FROM t GROUP BY a HAVING (count(*) > $1)", []any{1}},
		{"order by expression", New("SELECT id FROM t").OrderByExpr(`sort_key COLLATE "C" NULLS LAST`),
			`SELECT id FROM t ORDER BY sort_key COLLATE "C" NULLS LAST`, nil},
		{"order by expression with arg after where",
			New("SELECT id FROM t").Where("a = ?", 1).OrderByExpr("left(fpro.konto, ?)::numeric", 3),
			"SELECT id FROM t WHERE (a = $1) ORDER BY left(fpro.konto, $2)::numeric", []any{1, 3}},
		{"order by column, expression and key keep call order",
			New("SELECT id FROM t").OrderBy("a").OrderByExpr("COALESCE(NULLIF(b, '')::numeric, 0)").OrderByKey("name", "desc", allowed),
			"SELECT id FROM t ORDER BY a, COALESCE(NULLIF(b, '')::numeric, 0), c.name DESC", nil},
		{"empty order by expression adds nothing", New("SELECT id FROM t").OrderByExpr("  "), "SELECT id FROM t", nil},
		{"group by expression", New("SELECT EXTRACT(MONTH FROM danal) AS mesec FROM t").GroupByExpr("EXTRACT(MONTH FROM danal)"),
			"SELECT EXTRACT(MONTH FROM danal) AS mesec FROM t GROUP BY EXTRACT(MONTH FROM danal)", nil},
		{"group by and order by args are numbered in SQL order",
			New("SELECT LEFT(konto, ?) FROM t", 3).Where("a = ?", 1).GroupBy("b").GroupByExpr("LEFT(konto, ?)", 3).
				Having("count(*) > ?", 2).OrderByExpr("LEFT(konto, ?)", 3),
			"SELECT LEFT(konto, $1) FROM t WHERE (a = $2) GROUP BY b, LEFT(konto, $3) HAVING (count(*) > $4) ORDER BY LEFT(konto, $5)",
			[]any{3, 1, 3, 2, 3}},
		{"empty group by expression adds nothing", New("SELECT id FROM t").GroupByExpr(""), "SELECT id FROM t", nil},
		{"limit offset", New("SELECT id FROM t").Limit(10).Offset(20),
			"SELECT id FROM t LIMIT 10 OFFSET 20", nil},
		{"for update comes last", New("SELECT id FROM t").Where("id = ?", 4).OrderBy("id").Limit(1).ForUpdate(),
			"SELECT id FROM t WHERE (id = $1) ORDER BY id LIMIT 1 FOR UPDATE", []any{4}},
		{"join args numbered in order",
			New("SELECT t.id FROM t").AddJoin("JOIN u ON u.id = t.u AND u.k = ?", "k").AddEqual("t.a", 1),
			"SELECT t.id FROM t JOIN u ON u.id = t.u AND u.k = $1 WHERE t.a = $2", []any{"k", 1}},
		{"cte args come first",
			New("SELECT id FROM x").WithCTE("x", "SELECT id FROM t WHERE a = ?", 1).AddEqual("id", 2),
			"WITH x AS (SELECT id FROM t WHERE a = $1) SELECT id FROM x WHERE id = $2", []any{1, 2}},
		{"update", NewUpdate("t").Set("a = ?", 1).SetValue("b", 2).Where("id = ?", 3).Returning("id"),
			`UPDATE "t" SET a = $1, b = $2 WHERE (id = $3) RETURNING id`, []any{1, 2, 3}},
		{"update from", NewUpdate("t").Set("a = u.a").From("u").Where("t.id = u.id"),
			`UPDATE "t" SET a = u.a FROM u WHERE (t.id = u.id)`, nil},
		{"literal question mark and quoted text are not placeholders",
			New("SELECT id FROM t WHERE j ?? 'a?b' AND n = ?", 1),
			"SELECT id FROM t WHERE j ? 'a?b' AND n = $1", []any{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := tt.b.Build()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) || (len(args) > 0 && !reflect.DeepEqual(args, tt.wantArgs)) {
				t.Errorf("args got %v want %v", args, tt.wantArgs)
			}
		})
	}
}

func TestBuildErrors(t *testing.T) {
	tests := []struct {
		name string
		b    *Builder
		want string
	}{
		{"bad column", New("SELECT 1").AddEqual("a; DROP TABLE t", 1), "invalid column"},
		{"bad operator", New("SELECT 1").AddCondition("a", 1, "= 1 OR 1="), "not allowed"},
		{"bad order by", New("SELECT 1").OrderBy("name; DROP TABLE t"), "invalid ORDER BY"},
		{"bad order by subselect", New("SELECT 1").OrderBy("(SELECT 1)"), "invalid ORDER BY"},
		{"bad group by", New("SELECT 1").GroupBy("a, (select 1)"), "invalid GROUP BY"},
		{"bad cte name", New("SELECT 1").WithCTE("x as (select 1); --", "SELECT 1"), "invalid CTE"},
		{"bad whitelist value", New("SELECT 1").OrderByKey("k", "asc", map[string]string{"k": "a;b"}), "invalid whitelisted"},
		{"bad search column", New("SELECT 1").AddSearch([]string{"a)--"}, "x"), "invalid search column"},
		{"bad returning", New("SELECT 1").Returning("id; x"), "invalid RETURNING"},
		{"too few args", New("SELECT 1").Where("a = ? AND b = ?", 1), "2 placeholders but 1 arguments"},
		{"too many args", New("SELECT 1").Where("a = ?", 1, 2), "1 placeholders but 2 arguments"},
		{"order by expression missing arg", New("SELECT 1").OrderByExpr("left(a, ?)"), "1 placeholders but 0 arguments"},
		{"group by expression extra arg", New("SELECT 1").GroupByExpr("a", 1), "0 placeholders but 1 arguments"},
		{"empty base", New(""), "empty base"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tt.b.Build()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got error %v, want one containing %q", err, tt.want)
			}
		})
	}
}

func TestBuildCountDoesNotMutate(t *testing.T) {
	b := New("SELECT id FROM t").AddEqual("a", 1).AddSearch([]string{"n"}, "x").OrderBy("id").Limit(10).Offset(5)

	// FOR UPDATE is not allowed in the count's subquery: BuildCount drops it.
	if cq, _, _ := New("SELECT id FROM t").ForUpdate().BuildCount(); cq != "SELECT count(*) FROM (SELECT id FROM t) AS cnt" {
		t.Errorf("count of a FOR UPDATE query: %s", cq)
	}

	cq, cargs, err := b.BuildCount()
	if err != nil {
		t.Fatal(err)
	}
	wantC := "SELECT count(*) FROM (SELECT id FROM t WHERE a = $1 AND (n::text ILIKE $2)) AS cnt"
	if cq != wantC || len(cargs) != 2 {
		t.Errorf("count\n got: %s %v\nwant: %s", cq, cargs, wantC)
	}

	// Calling BuildCount first must not change the page query, and calling it twice is stable.
	q1, a1, _ := b.Build()
	q2, a2, _ := b.Build()
	wantQ := "SELECT id FROM t WHERE a = $1 AND (n::text ILIKE $2) ORDER BY id LIMIT 10 OFFSET 5"
	if q1 != wantQ || q2 != wantQ || len(a1) != 2 || len(a2) != 2 {
		t.Errorf("page query\n got: %s (%v)\nwant: %s", q1, a1, wantQ)
	}
}

func TestBuildCountDropsOrderByExprArgs(t *testing.T) {
	b := New("SELECT id FROM t").Where("a = ?", 1).OrderByExpr("left(k, ?)", 2)
	tests := []struct {
		name     string
		build    func() (string, []any, error)
		wantSQL  string
		wantArgs []any
	}{
		{"count", b.BuildCount, "SELECT count(*) FROM (SELECT id FROM t WHERE (a = $1)) AS cnt", []any{1}},
		{"page", b.Build, "SELECT id FROM t WHERE (a = $1) ORDER BY left(k, $2)", []any{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := tt.build()
			if err != nil {
				t.Fatal(err)
			}
			if sql != tt.wantSQL || !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("got %s %v\nwant %s %v", sql, args, tt.wantSQL, tt.wantArgs)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	a := New("SELECT id, 'a' AS src FROM ta").AddEqual("x", 1).Where("y = ?", 2)
	b := New("SELECT id, 'b' AS src FROM tb").AddEqual("x", 3)
	sql, args, err := NewUnionAll().Add(a).Add(b).OrderBy("2 DESC").Limit(5).Offset(10).Build()
	if err != nil {
		t.Fatal(err)
	}
	want := "(SELECT id, 'a' AS src FROM ta WHERE x = $1 AND (y = $2)) UNION ALL " +
		"(SELECT id, 'b' AS src FROM tb WHERE x = $3) ORDER BY 2 DESC LIMIT 5 OFFSET 10"
	if sql != want {
		t.Errorf("sql\n got: %s\nwant: %s", sql, want)
	}
	if !reflect.DeepEqual(args, []any{1, 2, 3}) {
		t.Errorf("args %v", args)
	}
	if _, _, err := NewUnion().Build(); err == nil {
		t.Error("empty union must fail")
	}
	if _, _, err := NewUnion().Add(New("SELECT 1")).OrderBy("a;b").Build(); err == nil {
		t.Error("bad union order by must fail")
	}
	if _, _, err := NewUnion().Add(New("SELECT 1").AddEqual("a b", 1)).Build(); err == nil {
		t.Error("error in a member query must surface")
	}
}

func TestWithCTEFrom(t *testing.T) {
	sub := func() *Builder {
		return New("SELECT konto, iznos FROM fpro").Where("god = ?", 2025).AddEqual("vkonta", "1").OrderBy("konto")
	}
	tests := []struct {
		name     string
		b        func() *Builder
		wantSQL  string
		wantArgs []any
	}{
		{"the CTE arguments are numbered before the outer ones",
			func() *Builder {
				return New("SELECT konto FROM filtered").WithCTEFrom("filtered", sub()).Where("konto ILIKE ?", "4%")
			},
			"WITH filtered AS (SELECT konto, iznos FROM fpro WHERE (god = $1) AND vkonta = $2 ORDER BY konto) " +
				"SELECT konto FROM filtered WHERE (konto ILIKE $3)",
			[]any{2025, "1", "4%"}},
		{"a static CTE after it continues the numbering",
			func() *Builder {
				return New("SELECT * FROM agg").WithCTEFrom("filtered", sub()).
					WithCTE("agg", "SELECT konto, sum(iznos) AS s FROM filtered WHERE konto <> ? GROUP BY konto", "9").
					Where("s > ?", 0)
			},
			"WITH filtered AS (SELECT konto, iznos FROM fpro WHERE (god = $1) AND vkonta = $2 ORDER BY konto), " +
				"agg AS (SELECT konto, sum(iznos) AS s FROM filtered WHERE konto <> $3 GROUP BY konto) SELECT * FROM agg WHERE (s > $4)",
			[]any{2025, "1", "9", 0}},
		{"a JSON ?? operator in the CTE survives as a single ?",
			func() *Builder {
				return New("SELECT id FROM f").WithCTEFrom("f", New("SELECT id FROM t").Where("j ?? 'a'").Where("n = ?", 3))
			},
			"WITH f AS (SELECT id FROM t WHERE (j ? 'a') AND (n = $1)) SELECT id FROM f",
			[]any{3}},
		{"later changes to the sub builder are not seen",
			func() *Builder {
				s := sub()
				b := New("SELECT konto FROM filtered").WithCTEFrom("filtered", s)
				s.Where("kar = ?", 7)
				return b
			},
			"WITH filtered AS (SELECT konto, iznos FROM fpro WHERE (god = $1) AND vkonta = $2 ORDER BY konto) SELECT konto FROM filtered",
			[]any{2025, "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := tt.b().Build()
			if err != nil {
				t.Fatal(err)
			}
			if sql != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", sql, tt.wantSQL)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args\n got: %#v\nwant: %#v", args, tt.wantArgs)
			}
		})
	}
	if _, _, err := New("SELECT 1").WithCTEFrom("x", New("SELECT 1").AddEqual("a b", 1)).Build(); err == nil {
		t.Error("an invalid sub builder must fail the outer Build")
	}
	if _, _, err := New("SELECT 1").WithCTEFrom("bad name", New("SELECT 1")).Build(); err == nil {
		t.Error("an invalid CTE name must fail")
	}
}

func TestQuoteIdent(t *testing.T) {
	for in, want := range map[string]string{"t": `"t"`, "a.b": `"a"."b"`, `we"ird`: `"we""ird"`} {
		if got := QuoteIdent(in); got != want {
			t.Errorf("QuoteIdent(%q)=%s want %s", in, got, want)
		}
	}
}

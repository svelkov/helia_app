package finansijsko

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
)

type sqlCall struct {
	query string
	args  []any
}

func recordingDB(calls *[]sqlCall) *db.MockDatabase {
	rec := func(q string, a []any) { *calls = append(*calls, sqlCall{q, a}) }
	return &db.MockDatabase{
		SelectContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			rec(q, a)
			return nil
		},
		GetContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			rec(q, a)
			return nil
		},
	}
}

func sessionCtx() context.Context {
	return domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7})
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// TestFinansijskoSQLScopes checks the conditions the migration added or kept on purpose: the
// tenant of ids that come from the request, the sort whitelist and the searches that were broken.
func TestFinansijskoSQLScopes(t *testing.T) {
	tests := []struct {
		name     string
		run      func(d db.Database) error
		want     string // a part of the first (single-line) query
		wantArgs []any
	}{
		{
			name: "stavke izvoda by the header id from the request are limited to the tenant",
			run: func(d db.Database) error {
				s := &IzvodiResource{izvdetRepo: repository.NewBaseRepository[domain.Fizvdet](d, "fizvdet")}
				return s.GetIzvodiDetail(sessionCtx(), &domain.TableData{}, true, 10, 1, domain.IzvodiParams{IDfizvzag: "9"})
			},
			want:     "FROM fizvdet d WHERE (d.idfizvzag = $1) AND (d.god = $2) AND (d.kar = $3)) AS cnt",
			wantArgs: []any{"9", 2026, 7},
		},
		{
			name: "stavke naloga for print preview: tenant and a whitelisted sort",
			run: func(d db.Database) error {
				s := &NalogResource{fproRepo: *repository.NewBaseRepository[domain.Fpro](d, "fpro")}
				return s.GetNalogStampanjeDataDetalji(sessionCtx(), &domain.TableData{}, false, 2, 10, 42, "", "konto", "desc")
			},
			want:     "WHERE (fpro.idfnal = $1) AND (fpro.god = $2) AND (fpro.kar = $3) ORDER BY fpro.konto DESC LIMIT 10 OFFSET 10",
			wantArgs: []any{int64(42), 2026, 7},
		},
		{
			name: "stavke naloga: a sort key outside the whitelist is ignored",
			run: func(d db.Database) error {
				s := &NalogResource{fproRepo: *repository.NewBaseRepository[domain.Fpro](d, "fpro")}
				return s.GetNalogStampanjeDataDetalji(sessionCtx(), &domain.TableData{}, false, 1, 10, 42, "", "konto; DROP TABLE fpro", "asc")
			},
			want:     "WHERE (fpro.idfnal = $1) AND (fpro.god = $2) AND (fpro.kar = $3) LIMIT 10",
			wantArgs: []any{int64(42), 2026, 7},
		},
		{
			name: "kontni plan: the sort column comes from the whitelist",
			run: func(d db.Database) error {
				s := &FkplResource{fkplRepo: repository.NewBaseRepository[domain.Fkpl](d, "fkpl")}
				return s.GetAllByVkonta(sessionCtx(), "2", &domain.TableData{}, 1, 10, false, "Naziv", "DESC", "", common.TipStampePreview)
			},
			want:     "WHERE fkpl.vkonta = $1 AND (fkpl.god = $2) AND (fkpl.kar = $3) ORDER BY fkpl.naziv DESC LIMIT 10",
			wantArgs: []any{"2", 2026, 7},
		},
		{
			name: "salda po komercijalistima: the search uses its own arguments (was $1/$2 = god/kar)",
			run: func(d db.Database) error {
				s := &SaldaResource{
					fproRepo:     repository.NewBaseRepository[domain.Fpro](d, "fpro"),
					saldaKomRepo: repository.NewBaseRepository[domain.SaldaKomercijalistiDto](d, "fpro"),
				}
				return s.SaldaPoKomercijalistima(sessionCtx(), &domain.TableData{}, true, 1, 10, "ab", "", "")
			},
			want: "WHERE fpro.vkonta = $1 AND ((fpro.vrd = 10 OR fpro.vrd = 30)) AND (fpro.god = $2) AND (fpro.kar = $3) " +
				"AND (kom.sifkom::TEXT ILIKE $4 OR kom.imeprezime::TEXT ILIKE $5) GROUP BY fpro.komid, kom.sifkom, kom.imeprezime) AS cnt",
			wantArgs: []any{1, 2026, 7, "%ab%", "%ab%"},
		},
		{
			name: "IOS detalji by the konto id from the request are limited to the tenant",
			run: func(d db.Database) error {
				s := &OtvoreneStavkeResource{fproRepo: *repository.NewBaseRepository[domain.Fpro](d, "fpro")}
				return s.GetIOSDetalji(sessionCtx(), 5, &domain.TableData{}, true, 10, 1, "", common.TipStampePreview, nil)
			},
			want:     "FROM fpro WHERE (fpro.idfkpl = $1) AND (fpro.god = $2) AND (fpro.kar = $3)) AS cnt",
			wantArgs: []any{int64(5), 2026, 7},
		},
		{
			name: "zakljucni list sintetika: filtered CTE in the tenant, prefix length as literal, search as argument",
			run: func(d db.Database) error {
				s := &BilansiResource{fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro"), cfg: config.Config{NDuzSint: 3}}
				return s.GetZakljucniListSintetika(sessionCtx(), &domain.TableData{}, domain.ZakljucniParams{OdKonta: "100", OdDatuma: "2026-01-01", SearchText: "20"}, false, 10, 1)
			},
			want: "WITH filtered AS (SELECT konto, fnal.tipdok, kat, iznos FROM fpro inner join fnal on fnal.idfnal = fpro.idfnal " +
				"WHERE (left(fpro.konto, 3) >= $1) AND fnal.danal >= $2 AND (fpro.god = $3) AND (fpro.kar = $4)), agg AS ( SELECT LEFT(konto, 1) AS k1, " +
				"LEFT(konto, 2) AS k2, LEFT(konto, 3) AS k3,",
			wantArgs: []any{"100", "2026-01-01", 2026, 7, "20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			if err := tt.run(recordingDB(&calls)); err != nil {
				t.Fatal(err)
			}
			if len(calls) == 0 {
				t.Fatal("no query was run")
			}
			got := oneLine(calls[0].query)
			if !strings.Contains(got, tt.want) {
				t.Errorf("sql\n got: %s\nwant to contain: %s", got, tt.want)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %#v want %#v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

// TestFinansijskoSQLHelpers checks the shared SQL fragments with their exact SQL and arguments.
func TestFinansijskoSQLHelpers(t *testing.T) {
	tests := []struct {
		name     string
		b        func() *querybuilder.Builder
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "bilans search: bils columns, aop as text, one argument per column",
			b: func() *querybuilder.Builder {
				b := querybuilder.New("SELECT bilsid FROM bils")
				whereBilansSearch(b, "bils", "7")
				return b
			},
			wantSQL:  "SELECT bilsid FROM bils WHERE (bils.nazp ILIKE $1 OR bils.grac ILIKE $2 OR bils.konta ILIKE $3 OR bils.aop::text ILIKE $4)",
			wantArgs: []any{"%7%", "%7%", "%7%", "%7%"},
		},
		{
			name: "bilans search: no text, no condition",
			b: func() *querybuilder.Builder {
				b := querybuilder.New("SELECT bilsid FROM bils")
				whereBilansSearch(b, "bils", "")
				return b
			},
			wantSQL: "SELECT bilsid FROM bils",
		},
		{
			name: "range: each bound only when set",
			b: func() *querybuilder.Builder {
				b := querybuilder.New("SELECT 1 FROM fpro")
				whereRange(b, "sifra::numeric", "", "99")
				return b
			},
			wantSQL:  "SELECT 1 FROM fpro WHERE (sifra::numeric <= $1)",
			wantArgs: []any{"99"},
		},
		{
			name: "dospece columns: valuta buckets in days from the stanje date",
			b: func() *querybuilder.Builder {
				cols, args := dospeceKolone(false, []any{"10"}, []any{"30"}, "(fpro.dadok + fpro.rok)::date", "2026-03-31", [5]int{15, 30, 60, 90, 120})
				return querybuilder.New("SELECT "+cols+" FROM fpro", args...)
			},
			wantSQL: "SELECT CASE WHEN fpro.vrd IN ($1) THEN fpro.iznos ELSE 0 END as realizacija, " +
				"CASE WHEN fpro.vrd IN ($2) THEN fpro.iznos ELSE 0 END as placeno, " +
				"CASE WHEN fpro.vrd IN ($3) AND fpro.dadok IS NOT NULL AND (fpro.dadok + fpro.rok)::date <= $4::date THEN fpro.iznos ELSE 0 END as dospelarealizacija, " +
				"CASE WHEN fpro.vrd IN ($5) AND (fpro.dadok + fpro.rok)::date > ($6::date + $7::int) AND (fpro.dadok + fpro.rok)::date <= ($8::date + $9::int) THEN fpro.iznos ELSE 0 END as dospece15, " +
				"CASE WHEN fpro.vrd IN ($10) AND (fpro.dadok + fpro.rok)::date > ($11::date + $12::int) AND (fpro.dadok + fpro.rok)::date <= ($13::date + $14::int) THEN fpro.iznos ELSE 0 END as dospece30, " +
				"CASE WHEN fpro.vrd IN ($15) AND (fpro.dadok + fpro.rok)::date > ($16::date + $17::int) AND (fpro.dadok + fpro.rok)::date <= ($18::date + $19::int) THEN fpro.iznos ELSE 0 END as dospece60, " +
				"CASE WHEN fpro.vrd IN ($20) AND (fpro.dadok + fpro.rok)::date > ($21::date + $22::int) AND (fpro.dadok + fpro.rok)::date <= ($23::date + $24::int) THEN fpro.iznos ELSE 0 END as dospece90, " +
				"CASE WHEN fpro.vrd IN ($25) AND (fpro.dadok + fpro.rok)::date > ($26::date + $27::int) AND (fpro.dadok + fpro.rok)::date <= ($28::date + $29::int) THEN fpro.iznos ELSE 0 END as dospece120, " +
				"CASE WHEN fpro.vrd IN ($30) AND (fpro.dadok + fpro.rok)::date > ($31::date + $32::int) THEN fpro.iznos ELSE 0 END as dospece120plus FROM fpro",
			wantArgs: []any{"10", "30", "10", "2026-03-31",
				"10", "2026-03-31", 0, "2026-03-31", 15,
				"10", "2026-03-31", 15, "2026-03-31", 30,
				"10", "2026-03-31", 30, "2026-03-31", 60,
				"10", "2026-03-31", 60, "2026-03-31", 90,
				"10", "2026-03-31", 90, "2026-03-31", 120,
				"10", "2026-03-31", 120},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := tt.b().Build()
			if err != nil {
				t.Fatal(err)
			}
			if got := oneLine(sql); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args\n got: %#v\nwant: %#v", args, tt.wantArgs)
			}
		})
	}
}

func TestDospeceDani(t *testing.T) {
	ok := domain.OtvStavkeParam{Dospece15: "15", Dospece30: "30", Dospece60: "60", Dospece90: "90", Dospece120: "120"}
	got, err := dospeceDani(ok)
	if err != nil || got != [5]int{15, 30, 60, 90, 120} {
		t.Fatalf("got %v, %v", got, err)
	}
	for _, bad := range []string{"", "1 day", "-5", "15'; DROP TABLE fpro; --"} {
		p := ok
		p.Dospece60 = bad
		if _, err := dospeceDani(p); err == nil || !strings.Contains(err.Error(), "dospece60") {
			t.Errorf("%q: want the dospece60 error, got %v", bad, err)
		}
	}
}

// TestFinansijskoFixedQueries checks the queries that were broken before the migration and are
// fixed now: each call is listed with the part of its SQL and its arguments.
func TestFinansijskoFixedQueries(t *testing.T) {
	type want struct {
		sql  string // a part of the (single-line) query
		args []any
	}
	tests := []struct {
		name  string
		run   func(d db.Database) error
		wants []want // one per query, in order
	}{
		{
			name: "bilans stanja: opening balance and monthly movements read tipdok/konto of fpro",
			run: func(d db.Database) error {
				s := &BilansiResource{fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro")}
				_, _, _, _, err := s.queryFproBilsAggregate(sessionCtx(), true, true, 2025, 7, 1, 6, "20")
				return err
			},
			wants: []want{
				{"from fpro WHERE (fpro.god = $1) AND (fpro.kar = $2) AND (fpro.tipdok = '00') AND fpro.konto ILIKE $3", []any{2025, 7, "20%"}},
				{"from fpro WHERE (fpro.god = $1) AND (fpro.kar = $2) AND (fpro.tipdok != '00') AND fpro.konto ILIKE $3 " +
					"AND (extract(month from danal)::int >= $4) AND (extract(month from danal)::int <= $5)", []any{2025, 7, "20%", 1, 6}},
			},
		},
		{
			name: "bilans uspeha: movements read tipdok of fpro",
			run: func(d db.Database) error {
				s := &BilansiResource{fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro")}
				_, _, err := s.queryFproAggregate(sessionCtx(), true, true, 2025, 7, 0, 0, "59")
				return err
			},
			wants: []want{
				{"from fpro WHERE (fpro.god = $1) AND (fpro.kar = $2) AND (fpro.tipdok != '00') AND fpro.konto ILIKE $3 " +
					"AND (fpro.konto != '599') AND (fpro.konto != '699') AND (left(fpro.konto, 4) != '5999') AND (left(fpro.konto, 4) != '6999')",
					[]any{2025, 7, "59%"}},
			},
		},
		{
			name: "salda klase 5 i 6 totals: an unset bound is skipped, no klasa means 5 and 6",
			run: func(d db.Database) error {
				s := &SaldaResource{fproRepo: repository.NewBaseRepository[domain.Fpro](d, "fpro")}
				_, err := s.GetSaldaKlase5i6TotalValues(sessionCtx(), domain.SaldaParam{OdSifre: "undefined", DoSifre: "100", OdKonta: "", DoKonta: "undefined"})
				return err
			},
			wants: []want{
				{"WHERE (COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0) <= $1) AND ((f.konto like '5%' OR f.konto like '6%')) " +
					"AND (f.god = $2) AND (f.kar = $3) AND (f.tipdok <> $4)", []any{"100", 2026, 7, "00"}},
				{"WHERE (COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0) <= $1) AND ((f.konto like '5%' OR f.konto like '6%')) " +
					"AND (f.god = $2) AND (f.kar = $3) AND (f.tipdok = $4)", []any{"100", 2026, 7, "00"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			if err := tt.run(recordingDB(&calls)); err != nil {
				t.Fatal(err)
			}
			if len(calls) != len(tt.wants) {
				t.Fatalf("%d queries, want %d", len(calls), len(tt.wants))
			}
			for i, w := range tt.wants {
				if got := oneLine(calls[i].query); !strings.Contains(got, w.sql) {
					t.Errorf("query %d\n got: %s\nwant to contain: %s", i, got, w.sql)
				}
				if !reflect.DeepEqual(calls[i].args, w.args) {
					t.Errorf("query %d args got %#v want %#v", i, calls[i].args, w.args)
				}
			}
		})
	}
}

func TestBilansAggregateReturnsError(t *testing.T) {
	d := &db.MockDatabase{SelectContextFunc: func(context.Context, interface{}, string, ...interface{}) error {
		return errors.New("column fnal.tipdok does not exist")
	}}
	s := &BilansiResource{fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro")}
	if _, _, err := s.queryFproAggregate(sessionCtx(), true, true, 2025, 7, 0, 0, "59"); err == nil {
		t.Error("a failed aggregate must return its error (it used to be ignored and gave 0, 0)")
	}
	if _, _, _, _, err := s.queryFproBilsAggregate(sessionCtx(), true, true, 2025, 7, 1, 12, "20"); err == nil {
		t.Error("a failed aggregate must return its error")
	}
}

// TestAzuriranjeKontaLookups runs two stavke izvoda through AzuriranjeKonta: every row gets its
// own lookups (they used to pile up), and the konto of the partner is looked up for that partner.
func TestAzuriranjeKontaLookups(t *testing.T) {
	var calls []sqlCall
	rec := func(q string, a []any) { calls = append(calls, sqlCall{oneLine(q), a}) }
	d := &db.MockDatabase{
		SelectContextFunc: func(_ context.Context, dest interface{}, q string, a ...interface{}) error {
			rec(q, a)
			switch p := dest.(type) {
			case *[]domain.Fizvdet:
				*p = []domain.Fizvdet{
					{IDFizvdet: 1, Osnplac: "101", Brracup: "160-1", Duguje: 5},
					{IDFizvdet: 2, Osnplac: "102", Potrazuje: 3},
				}
			case *[]domain.TekRacuni:
				*p = []domain.TekRacuni{{IDPartneri: 9}}
			case *[]domain.Fkpl:
				*p = []domain.Fkpl{{Konto: "4350", Sifra: "77"}}
			}
			return nil
		},
		ExecContextFunc: func(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
			rec(q, a)
			return nil, nil
		},
	}
	s := &IzvodiResource{
		izvdetRepo:    repository.NewBaseRepository[domain.Fizvdet](d, "fizvdet"),
		sifplizvRepo:  repository.NewBaseRepository[domain.Sifplizv](d, "sifplizv"),
		tekracuniRepo: repository.NewBaseRepository[domain.TekRacuni](d, "tekracuni"),
		partneriRepo:  repository.NewBaseRepository[domain.Partneri](d, "partneri"),
		fkplRepo:      repository.NewBaseRepository[domain.Fkpl](d, "fkpl"),
		cfg:           config.Config{Konta: config.KontaConfig{KontoDobavljaca: "4350", KontoKupca: "2040"}},
	}
	if err := s.AzuriranjeKonta(sessionCtx(), 55, false, true); err != nil {
		t.Fatal(err)
	}
	wants := []sqlCall{
		{"osnplac, brracup, pnabrodob FROM fizvdet WHERE (idfizvzag = $1) AND (god = $2) AND (kar = $3)", []any{int64(55), 2026, 7}},
		// row 1
		{"FROM sifplizv WHERE (sifplac = $1)", []any{"101"}},
		{"SELECT idpartneri, tekrac FROM tekracuni WHERE (tekrac = $1) AND (god = $2) AND (kar = $3)", []any{"160-1", 2026, 7}},
		{"FROM fkpl WHERE (konto = $1) AND (idpartneri = $2) AND (god = $3) AND (kar = $4)", []any{"4350", int64(9), 2026, 7}},
		{`UPDATE "fizvdet" SET konto = $1, sifra = $2, kat = $3, vrd = $4 WHERE (idfizvdet = $5)`, []any{"4350", "77", int16(1), 40, int64(1)}},
		// row 2: only its own payment code, no partner (no account number)
		{"FROM sifplizv WHERE (sifplac = $1)", []any{"102"}},
		{`UPDATE "fizvdet" SET konto = $1, sifra = $2, kat = $3, vrd = $4 WHERE (idfizvdet = $5)`, []any{"", "", int16(0), 0, int64(2)}},
	}
	if len(calls) != len(wants) {
		for i, c := range calls {
			t.Logf("%d: %s %v", i, c.query, c.args)
		}
		t.Fatalf("%d queries, want %d", len(calls), len(wants))
	}
	for i, w := range wants {
		if !strings.Contains(calls[i].query, w.query) {
			t.Errorf("query %d\n got: %s\nwant to contain: %s", i, calls[i].query, w.query)
		}
		if !reflect.DeepEqual(calls[i].args, w.args) {
			t.Errorf("query %d args got %#v want %#v", i, calls[i].args, w.args)
		}
	}
}

func TestIzvodiHeaderSearch(t *testing.T) {
	var calls []sqlCall
	s := &IzvodiResource{izvhdrRepo: repository.NewBaseRepository[domain.Fizvzag](recordingDB(&calls), "fizvzag")}
	s.SetIzvodiFields()
	if err := s.GetIzvodiHeader(sessionCtx(), &domain.TableData{}, true, 10, 1, domain.IzvodiParams{SearchText: "abc"}); err != nil {
		t.Fatal(err)
	}
	// Every searched column is qualified with the izvod (the joined banke has konto, sifra and
	// brrac too); banka (NullString) and datizv (NullTime) are not searched.
	want := "WHERE (i.god = $1) AND (i.kar = $2) AND (i.konto::text ILIKE $3 OR i.sifra::text ILIKE $4 OR i.izvbr::text ILIKE $5 " +
		"OR i.nalog::text ILIKE $6 OR i.tipdok::text ILIKE $7 OR i.izvsts::text ILIKE $8 OR i.brrac::text ILIKE $9)) AS cnt"
	if got := oneLine(calls[0].query); !strings.Contains(got, want) {
		t.Errorf("sql\n got: %s\nwant to contain: %s", got, want)
	}
	wantArgs := []any{2026, 7, "%abc%", "%abc%", "%abc%", "%abc%", "%abc%", "%abc%", "%abc%"}
	if !reflect.DeepEqual(calls[0].args, wantArgs) {
		t.Errorf("args got %#v want %#v", calls[0].args, wantArgs)
	}
}

func TestFkplFieldValues(t *testing.T) {
	id := int64(4)
	in := []domain.Fields{{Name: "konto"}, {Name: "tipanalitikeid"}, {Name: "vkonta"}, {Name: "naziv"}, {Name: "devizni"}, {Name: "kolicinski"}}
	got := fkplFieldValues(&domain.Fkpl{Konto: "2040", Naziv: "Kupci", Vkonta: 1, Devizni: true, TipanalitikeID: &id}, in)
	want := []domain.Fields{{Name: "konto", Value: "2040"}, {Name: "tipanalitikeid", Value: "4"}, {Name: "vkonta", Value: "1"},
		{Name: "naziv", Value: "Kupci"}, {Name: "devizni", Value: "true"}, {Name: "kolicinski", Value: "false"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	if in[0].Value != "" {
		t.Error("the fields of the caller must not be changed")
	}
	got = fkplFieldValues(&domain.Fkpl{}, []domain.Fields{{Name: "tipanalitikeid"}})
	if !got[0].IsNull {
		t.Error("a missing tip analitike must be NULL, not an empty string")
	}
}

// TestRequiredKeysAreConditions checks keys that used to go through AddEqual (skipped when empty).
func TestRequiredKeysAreConditions(t *testing.T) {
	id := int64(4)
	tests := []struct {
		name     string
		run      func(d db.Database) error
		want     string
		wantArgs []any
	}{
		{
			name: "fkpl duplicate check: konto without analitika is compared with IS NULL",
			run: func(d db.Database) error {
				s := &FkplResource{fkplRepo: repository.NewBaseRepository[domain.Fkpl](d, "fkpl"), cfg: config.Config{NDuzSint: 3}}
				s.ValidateEntity(sessionCtx(), &domain.Fkpl{Konto: "2040", Naziv: "Kupci", Vkonta: 2})
				return nil
			},
			want:     "SELECT idfkpl FROM fkpl WHERE (konto = $1) AND (tipanalitikeid IS NULL) AND (god = $2) AND (kar = $3)",
			wantArgs: []any{"2040", 2026, 7},
		},
		{
			name: "fkpl duplicate check: konto with analitika",
			run: func(d db.Database) error {
				s := &FkplResource{fkplRepo: repository.NewBaseRepository[domain.Fkpl](d, "fkpl"), cfg: config.Config{NDuzSint: 3}}
				s.ValidateEntity(sessionCtx(), &domain.Fkpl{Konto: "2040", Naziv: "Kupci", Vkonta: 1, TipanalitikeID: &id})
				return nil
			},
			want:     "SELECT idfkpl FROM fkpl WHERE (konto = $1) AND (tipanalitikeid = $2) AND (god = $3) AND (kar = $4)",
			wantArgs: []any{"2040", int64(4), 2026, 7},
		},
		{
			name: "saldo pojedinacnog konta: an empty konto stays a condition",
			run: func(d db.Database) error {
				s := &SaldaResource{fproRepo: repository.NewBaseRepository[domain.Fpro](d, "fpro")}
				return s.GetSaldaPojedinacnihKonta(sessionCtx(), &domain.TableData{}, false, 10, 1, "", "", "1")
			},
			want:     "FROM fpro f WHERE (f.konto = $1) AND f.vkonta = $2 AND (f.tipdok = $3) AND (f.god = $4) AND (f.kar = $5)",
			wantArgs: []any{"", "1", "00", 2026, 7, "", "1", "00", 2026, 7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			if err := tt.run(recordingDB(&calls)); err != nil {
				t.Fatal(err)
			}
			if len(calls) == 0 {
				t.Fatal("no query was run")
			}
			if got := oneLine(calls[0].query); !strings.Contains(got, tt.want) {
				t.Errorf("sql\n got: %s\nwant to contain: %s", got, tt.want)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %#v want %#v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

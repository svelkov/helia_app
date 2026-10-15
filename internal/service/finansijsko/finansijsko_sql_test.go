package finansijsko

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"helia/config"
	"helia/i18n"
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
			name: "bilans stanja: the salda of all the konto prefixes in one query (opening balance and monthly movements of fpro)",
			run: func(d db.Database) error {
				s := &BilansiResource{fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro")}
				_, err := s.fproBilsSalda(sessionCtx(), true, true, 2025, 7, 1, 6, []string{"20", "204"})
				return err
			},
			wants: []want{
				{"WITH k AS (select fpro.konto, " +
					"sum(case when fpro.tipdok = '00' and fpro.kat in (1,2) then fpro.iznos else 0 end) as pdug, " +
					"sum(case when fpro.tipdok = '00' and fpro.kat in (3,4) then fpro.iznos else 0 end) as ppot, " +
					"sum(case when fpro.tipdok != '00' and fpro.kat in (1,2) and true and extract(month from fpro.danal)::int >= $1 " +
					"and extract(month from fpro.danal)::int <= $2 then fpro.iznos else 0 end) as mdug, " +
					"sum(case when fpro.tipdok != '00' and fpro.kat in (3,4) and true and extract(month from fpro.danal)::int >= $3 " +
					"and extract(month from fpro.danal)::int <= $4 then fpro.iznos else 0 end) as mpot " +
					"from fpro WHERE (fpro.god = $5) AND (fpro.kar = $6) AND (fpro.konto is not null) GROUP BY fpro.konto) " +
					"select p.prefiks as konto, coalesce(sum(k.pdug), 0) as pocstanjedug, coalesce(sum(k.ppot), 0) as pocstanjepot, " +
					"coalesce(sum(k.mdug), 0) as prometdug, coalesce(sum(k.mpot), 0) as prometpot " +
					"from (values ($7::text), ($8::text)) p(prefiks) join k on left(lower(k.konto), length(p.prefiks)) = lower(p.prefiks) " +
					"GROUP BY p.prefiks", []any{1, 6, 1, 6, 2025, 7, "20", "204"}},
			},
		},
		{
			name: "bilans stanja: no konto prefix, no query",
			run: func(d db.Database) error {
				s := &BilansiResource{fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro")}
				_, err := s.fproBilsSalda(sessionCtx(), true, true, 2025, 7, 1, 6, nil)
				return err
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
	if _, err := s.fproBilsSalda(sessionCtx(), true, true, 2025, 7, 1, 12, []string{"20"}); err == nil {
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

// TestBilansStanjaObrada runs the obrada of the bilans stanja over two levels of rows: the leaf rows get their iznosi
// from the salda of their konto prefixes (all of them read from fpro in one query), the summary row adds and
// subtracts the rows of its pozicije, and every row is written by one UPDATE (they used to be one statement a row).
func TestBilansStanjaObrada(t *testing.T) {
	bils := []domain.Bils{
		{BilsID: 1, AOP: 10, Konta: "+S20;-D204", NiPo: 1},
		{BilsID: 2, AOP: 11, Konta: "+P43", NiPo: 1},
		{BilsID: 3, AOP: 20, NiPo: 2, Pozic1: 10, Pozic2: -11},
	}
	salda := []domain.FproDto{
		{Fpro: domain.Fpro{Konto: "20"}, PocStanjeDug: 100000, PocStanjePot: 40000, PrometDug: 500000, PrometPot: 200000},
		{Fpro: domain.Fpro{Konto: "204"}, PocStanjeDug: 10000, PrometDug: 50000},
		{Fpro: domain.Fpro{Konto: "43"}, PocStanjePot: 30000, PrometPot: 70000},
	}
	var fproUpiti int
	var izmene []sqlCall
	d := &db.MockDatabase{
		SelectContextFunc: func(_ context.Context, dest interface{}, q string, a ...interface{}) error {
			switch dest := dest.(type) {
			case *[]domain.Bils:
				switch {
				case strings.Contains(q, "max(nipo)"):
					*dest = []domain.Bils{{NiPo: 2}}
				case strings.Contains(q, "pozic_1"):
					*dest = append([]domain.Bils(nil), bils...)
				}
			case *[]domain.FproDto:
				fproUpiti++
				if got := a[len(a)-3:]; !reflect.DeepEqual(got, []any{"20", "204", "43"}) {
					t.Errorf("konto prefixes %v, want 20, 204, 43", got)
				}
				*dest = salda
			}
			return nil
		},
		ExecContextFunc: func(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
			izmene = append(izmene, sqlCall{q, a})
			return driverRowsAffected(1), nil
		},
	}
	s := &BilansiResource{
		bilsRepo: repository.NewBaseRepository[domain.Bils](d, "bils"),
		fproRepo: repository.NewBaseRepository[domain.FproDto](d, "fpro"),
		fin:      repository.NewFinansijskoRepository(d),
	}
	var tbl domain.TableData
	if err := s.GetBilansStanjaObrada(sessionCtx(), &tbl, &domain.BilansiTotals{}, "2026-06-30", false, true); err != nil {
		t.Fatal(err)
	}
	if fproUpiti != 1 {
		t.Errorf("%d queries of fpro, want 1", fproUpiti)
	}
	// The resets of the totals (3 with lPGODizPS), then the iznosi of all the rows in one statement.
	if len(izmene) != 4 {
		t.Fatalf("%d statements, want 4: %v", len(izmene), izmene)
	}
	upis := izmene[3]
	if !strings.Contains(upis.query, "FROM (values") {
		t.Fatalf("the last statement is not the write of the iznosi: %s", upis.query)
	}
	want := []any{"ana",
		// 20: S (500000 - 200000) + (100000 - 40000); 204: -D 50000 and 10000.
		1, 300000.0, int64(300), 50000.0, int64(50), 0.0, int64(0),
		// 43: P 70000 + 30000.
		2, 100000.0, int64(100), 30000.0, int64(30), 0.0, int64(0),
		// The summary row: AOP 10 - AOP 11.
		3, 200000.0, int64(200), 20000.0, int64(20), 0.0, int64(0),
		2026, 7}
	if !reflect.DeepEqual(upis.args, want) {
		t.Errorf("iznosi\n got %#v\nwant %#v", upis.args, want)
	}
}

// driverRowsAffected is the sql.Result of a statement that changed n rows.
type driverRowsAffected int64

func (n driverRowsAffected) LastInsertId() (int64, error) { return 0, nil }
func (n driverRowsAffected) RowsAffected() (int64, error) { return int64(n), nil }

// TestBilansiStampaPoAOP checks that the prints of the bilans stanja and of the bilans uspeha list their rows by AOP.
func TestBilansiStampaPoAOP(t *testing.T) {
	for _, tt := range []struct {
		name string
		run  func(s *BilansiResource) error
		want string
	}{
		{name: "bilans stanja", want: "ORDER BY bils.aop ASC",
			run: func(s *BilansiResource) error {
				return s.GetBilansStanjaZaStampu(sessionCtx(), &domain.TableData{}, common.TipStampePrint, false)
			}},
		{name: "bilans uspeha", want: "ORDER BY bilu.aop ASC",
			run: func(s *BilansiResource) error {
				return s.GetBilansUspehaZaStampu(sessionCtx(), &domain.TableData{}, common.TipStampePrint)
			}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			d := recordingDB(&calls)
			s := &BilansiResource{
				bilsRepo: repository.NewBaseRepository[domain.Bils](d, "bils"),
				biluRepo: repository.NewBaseRepository[domain.Bilu](d, "bilu"),
			}
			if err := tt.run(s); err != nil {
				t.Fatal(err)
			}
			if len(calls) == 0 || !strings.HasSuffix(oneLine(calls[len(calls)-1].query), tt.want) {
				t.Errorf("the rows of the print are not ordered by AOP: %v", calls)
			}
		})
	}
}

// TestBilansiStampaUkupno checks the totals row at the end of the tables of the prints of the bilans stanja and of
// the bilans uspeha: "Ukupno" under the first column and the sums of the shown values under their own columns.
func TestBilansiStampaUkupno(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	d := &db.MockDatabase{SelectContextFunc: func(_ context.Context, dest interface{}, _ string, _ ...interface{}) error {
		switch dest := dest.(type) {
		case *[]domain.Bils:
			*dest = []domain.Bils{
				{BilsID: 1, AOP: 1, TGod: 1500, PGod: 700, PGodPS: 300, TGodH: 2, PGodH: 1, PGodHPS: 0},
				// A negative value is shown (and summed) as 0.
				{BilsID: 2, AOP: 2, TGod: -100, PGod: 300, PGodPS: 200, TGodH: 3, PGodH: 4, PGodHPS: 5},
			}
		case *[]domain.Bilu:
			*dest = []domain.Bilu{
				{BiluID: 1, AOP: 1001, TGod: 2500, PGod: 1000, TGodH: 3, PGodH: 1},
				{BiluID: 2, AOP: 1002, TGod: 500, PGod: 2000, TGodH: 1, PGodH: 2},
			}
		}
		return nil
	}}
	s := &BilansiResource{
		bilsRepo: repository.NewBaseRepository[domain.Bils](d, "bils"),
		biluRepo: repository.NewBaseRepository[domain.Bilu](d, "bilu"),
	}
	s.setServiceFieldValues()
	f := func(v float64, dec int) string { return common.FormatNumberWithSystemLocale(v, dec) }
	ukupno := i18n.GetInstance().Label("Ukupno")
	for _, tt := range []struct {
		name    string
		headers []domain.Fields
		run     func(tbl *domain.TableData) error
		want    map[int]string
	}{
		{name: "bilans stanja", headers: s.GetBilansStanjaStampaTableFields(),
			run: func(tbl *domain.TableData) error {
				return s.GetBilansStanjaZaStampu(sessionCtx(), tbl, common.TipStampePreview, false)
			},
			want: map[int]string{0: ukupno, 4: f(1500, 2), 5: f(1000, 2), 6: f(500, 2), 7: f(5, 2), 8: f(5, 2), 9: f(5, 2)}},
		{name: "bilans uspeha", headers: s.GetBilansUspehaStampaTableFields(),
			run: func(tbl *domain.TableData) error {
				return s.GetBilansUspehaZaStampu(sessionCtx(), tbl, common.TipStampePreview)
			},
			want: map[int]string{0: ukupno, 4: f(4, 2), 5: f(3, 2), 6: f(3000, 2), 7: f(3000, 2)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tbl := domain.TableData{Headers: tt.headers}
			if err := tt.run(&tbl); err != nil {
				t.Fatal(err)
			}
			if !tbl.HasTotals || len(tbl.Totals) != len(tbl.Headers) {
				t.Fatalf("no totals row: HasTotals %v, %d totals for %d columns", tbl.HasTotals, len(tbl.Totals), len(tbl.Headers))
			}
			for i, got := range tbl.Totals {
				if got != tt.want[i] {
					t.Errorf("total of column %d (%s) = %q, want %q", i, tbl.Headers[i].Name, got, tt.want[i])
				}
				if tt.want[i] != "" && !tbl.Headers[i].IncludeInTotals {
					t.Errorf("column %d (%s) does not show its total", i, tbl.Headers[i].Name)
				}
			}
		})
	}
}

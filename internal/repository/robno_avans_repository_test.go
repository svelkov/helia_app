package repository

import (
	"context"
	"strings"
	"testing"

	"helia/internal/infrastructure/db"
)

// rsifKoloneOdgovor answers the catalog query of the columns of rsif with cols and the insert of
// the artikal of the šifra 0 with its rsifid.
func rsifKoloneOdgovor(cols [][]any) func(string, []any) drvResult {
	return func(q string, _ []any) drvResult {
		if strings.Contains(q, "information_schema") {
			return drvResult{cols: []string{"key", "c1", "c2", "c3", "c4"}, rows: cols}
		}
		return drvResult{cols: []string{"rsifid"}, rows: [][]any{{int64(99)}}}
	}
}

var rsifKolone = [][]any{
	{"rsifid", "integer", "NO", "", "true"},
	{"god", "integer", "NO", "", "false"},
	{"kar", "integer", "NO", "", "false"},
	{"sifra", "integer", "NO", "", "false"},
	{"naziv", "character varying", "NO", "", "false"},
	{"tip", "character", "NO", "", "false"},
	{"grupa", "integer", "NO", "grupe.idgrupe", "true"},
	{"opis", "text", "NO", "", "false"},
	{"aktivan", "boolean", "NO", "", "false"},
	{"datum", "date", "NO", "", "false"},
	{"napomena", "text", "YES", "", "false"},
	{"cena", "numeric", "NO", "", "true"},
	{"po", "integer", "NO", "", "false"},
	{"xopunos", "character varying", "NO", "", "false"},
	{"xdatunosa", "timestamp without time zone", "NO", "", "false"},
}

var robnoAvansCases = []goldenCase{
	// avansfakt
	{name: "avansfakt add", respond: answer(1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			return r.AvansfaktDodaj(ctx, 41)
		}},
	{name: "avansfakt add iznosi", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AvansfaktDodajIznose(ctx, tx, 41, 100, 20, 50, 5, 10)
		}},
	{name: "avansfakt delete", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AvansfaktBrisi(ctx, tx, 41)
		}},
	{name: "fvknjrac konta of the kupac", respond: answer(0, "2040,2041", "31"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.FvknjracKontaKupca(ctx, "2040", "15")
			return err
		}},
	{name: "rdok profak", respond: answer(0, "4, 7"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RdokProfak(ctx, 41)
			return err
		}},
	{name: "rdok stopa and napomena", respond: answer(0, 20.0, ""), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RdokStopaNapomena(ctx, tx, 41, 3)
			return err
		}},
	// rpro, rsif, rcene
	{name: "rpro saved stavka for update", respond: answer(0, int64(41), "1320", "S20", 1.0, 100.0, int64(3), int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproStaraStavka(ctx, tx, 77)
			return err
		}},
	{name: "rpro renumber", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RproPrenumerisi(ctx, tx, 41, 2)
		}},
	{name: "rsif of the šifra 0", respond: answer(0, int64(5), "kom"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RsifNula(ctx, tx)
			return err
		}},
	{name: "rcene of the šifra 0", respond: answer(0, int64(0)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RceneNulaPostoji(ctx, tx)
			return err
		}},
	{name: "rsif create the šifra 0", respond: rsifKoloneOdgovor(rsifKolone), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RsifKreirajNulu(ctx, tx, "Avans")
			return err
		}},
	// brisanje
	{name: "rdok to delete for update", respond: answer(0, int64(9), "", "", 0.0, "", 0.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RdokZaBrisanje(ctx, tx, 41)
			return err
		}},
	{name: "rpro ids of the document", respond: func(string, []any) drvResult {
		return drvResult{cols: []string{"rproid"}, rows: [][]any{{int64(77)}, {int64(78)}}}
	},
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			ids, err := r.RproIDsDokumenta(ctx, 41)
			if err == nil && len(ids) != 2 {
				return errFound
			}
			return err
		}},
	{name: "avansfakt of the faktura delete", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AvansfaktFaktureBrisi(ctx, tx, 41)
		}},
	{name: "faktrate delete", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.FaktrateBrisi(ctx, tx, 41)
		}},
	{name: "dokusldet delete", respond: answer(0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DokusldetBrisi(ctx, tx, 41)
		}},
	{name: "rnal one document less", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RnalOduzmiDokument(ctx, tx, 9)
		}},
	{name: "rnal of another period", respond: answer(0), inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RnalOduzmiDokument(ctx, tx, 9)
		}},
}

func TestRobnoAvansGolden(t *testing.T) {
	runGolden(t, "robno_avans.golden", robnoAvansCases)
}

func TestRsifKreirajNuluRejectsNames(t *testing.T) {
	tests := []struct {
		name string
		col  []any
	}{
		{"column name", []any{"naziv; drop table rsif", "text", "NO", "", "false"}},
		{"qualified column name", []any{"x.naziv", "text", "NO", "", "false"}},
		{"referenced table", []any{"grupa", "integer", "NO", "grupe x.idgrupe", "false"}},
		{"referenced column", []any{"grupa", "integer", "NO", "grupe.id grupe", "false"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, f := newDrvDB(t, rsifKoloneOdgovor([][]any{tt.col}))
			if _, err := NewRobnoRepository(d).RsifKreirajNulu(robnoCtx(), nil, "Avans"); err == nil {
				t.Fatal("want an error")
			}
			if len(f.calls) != 1 {
				t.Fatalf("%d statements, want only the catalog query", len(f.calls))
			}
		})
	}
}

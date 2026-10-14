package repository

import (
	"context"
	"testing"
	"time"

	"helia/internal/infrastructure/db"
)

var datumUnosa = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

var robnoUnosCases = []goldenCase{
	{name: "konto of the magacin or of the stavke", respond: answer(0, "1320"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.KontoMagacinaIliStavke(ctx, 3, 41)
			return err
		}},
	{name: "šifre of the šifarnici", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.SifreSifarnika(ctx, 4, 6, 2, 8, 3)
			return err
		}},
	{name: "avans by broj", respond: answer(0, int64(40)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.AvansIDPoBroju(ctx, 4, " 2040 ", " 15 ")
			return err
		}},
	{name: "rdok by broj with fana for update", respond: answer(0, int64(41), "15"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RdokPoBrojuFana(ctx, tx, 3, 10, 12)
			return err
		}},
	{name: "rpro set the partner", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RproPostaviPartnera(ctx, tx, 41, "pkto", "2040")
		}},
	{name: "rnal of a new document for update", respond: answer(0, "01", int64(1), nil, nil, int64(1), int64(3), int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RnalZaDokument(ctx, tx, 9)
			return err
		}},
	{name: "avansfakt delete the others", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AvansfaktBrisiOstale(ctx, tx, 41, []int64{38, 39})
		}},
	{name: "avansfakt delete all", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AvansfaktBrisiOstale(ctx, tx, 41, nil)
		}},
	{name: "avansfakt of the faktura", respond: answer(0, int64(1)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.AvansfaktPostoji(ctx, tx, 38, 41)
			return err
		}},
	{name: "avansfakt iznosi", respond: answer(0, 100.0, 0.0, 0.0, 20.0, 0.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.AvansfaktIznosi(ctx, tx, 38)
			return err
		}},
	{name: "avansi of the obračun", respond: func(string, []any) drvResult { return drvResult{cols: []string{"avansid"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.AvansiObracuna(ctx, 41)
			return err
		}},
	{name: "faktrate of the document", respond: func(string, []any) drvResult { return drvResult{cols: []string{"faktrateid"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.FaktrateDokumenta(ctx, 41)
			return err
		}},
	{name: "stavke of the obračun", respond: func(string, []any) drvResult { return drvResult{cols: []string{"ugrabat"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.StavkeObracuna(ctx, 41)
			return err
		}},
	{name: "saldo of the kupac", respond: func(string, []any) drvResult { return drvResult{cols: []string{"fakture"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.SaldoKupca(ctx, " 2040 ", " 15 ", datumUnosa)
			return err
		}},
	{name: "avansi of the document", respond: func(string, []any) drvResult { return drvResult{cols: []string{"avansid"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.AvansiDokumenta(ctx, 41)
			return err
		}},
	{name: "banke žiro računi", respond: func(string, []any) drvResult {
		return drvResult{cols: []string{"r"}, rows: [][]any{{"160-1  Banka"}}}
	},
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.BankeZiroRacuni(ctx)
			return err
		}},
	{name: "defaults of a new document", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.NoviDokumentDefaults(ctx, "Firma d.o.o.")
			return err
		}},
	{name: "rpro stavke of the document", respond: func(string, []any) drvResult { return drvResult{cols: []string{"rproid"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproStavkeDokumenta(ctx, 41)
			return err
		}},
	{name: "rdok kontekst", respond: func(string, []any) drvResult { return drvResult{cols: []string{"rdokid"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RdokKontekst(ctx, 41)
			return err
		}},
	{name: "rsif artikal with its stanje", respond: func(string, []any) drvResult { return drvResult{cols: []string{"rsifid"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalStanje(ctx, 3, " 1320 ", 15)
			return err
		}},
	{name: "magkonto konta", respond: func(string, []any) drvResult { return drvResult{cols: []string{"k"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagkontoKonta(ctx, 3)
			return err
		}},
	{name: "rpro last cena to the kupac", respond: func(string, []any) drvResult { return drvResult{cols: []string{"fcena"}} },
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproZadnjaCenaKupca(ctx, 41, 15)
			return err
		}},
	{name: "rsif id, jm and dani for update", respond: answer(0, int64(77), "kom", int64(30)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, err := r.RsifIDJmDani(ctx, tx, 15)
			return err
		}},
	{name: "rpro count of the nalog", respond: answer(0, int64(4)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproBrojNaloga(ctx, tx, 9)
			return err
		}},
	{name: "rnal dug and pot for update", respond: answer(0, 10.0, 0.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RnalDugPot(ctx, tx, 9)
			return err
		}},
	{name: "rsif artikal of a stavka by barkod", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalStavke(ctx, true, "8600123456789", " 1320 ", datumUnosa)
			return err
		}},
	{name: "rsta stanje of a stavka", respond: answer(0, 3.0, 9.5, 9.0, 8.0, 8.5),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RstaStanjeStavke(ctx, 3, "1320", 15)
			return err
		}},
	{name: "rpro form of a stavka", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproStavkaForma(ctx, 77, 41)
			return err
		}},
	{name: "rpro taxexreco", respond: answer(0, "PDV-RS-24"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproTaxexreco(ctx, 77)
			return err
		}},
}

func TestRobnoUnosGolden(t *testing.T) {
	runGolden(t, "robno_unos.golden", robnoUnosCases)
}

func TestRproPostaviPartneraRejectsColumns(t *testing.T) {
	d, f := newDrvDB(t, answer(1))
	if err := NewRobnoRepository(d).RproPostaviPartnera(robnoCtx(), nil, 41, "iznos", "0"); err == nil {
		t.Fatal("want an error for a column that is not of the partner")
	}
	if len(f.calls) != 0 {
		t.Fatalf("%d statements ran, want none", len(f.calls))
	}
}

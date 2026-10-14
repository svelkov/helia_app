package repository

import (
	"context"
	"testing"

	"helia/internal/infrastructure/db"
)

var robnoUslugeCases = []goldenCase{
	{name: "magacin zalihe", respond: answer(0, int64(2), int64(3)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.MagacinZalihe(ctx, 3)
			return err
		}},
	{name: "magkonto first konto", respond: answer(0, "1320"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagkontoPrviKonto(ctx, 3)
			return err
		}},
	{name: "rdok izvorni of another year", respond: answer(0, nil, 0.0, ""),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RdokIzvorniGodine(ctx, 2025, 1, 172, 4)
			return err
		}},
	{name: "rsif artikal of a usluga by šifra", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalUsluge(ctx, false, "15", nil)
			return err
		}},
	{name: "rsta stanje of a usluga", respond: answer(0, 3.0, 8.0, 9.5, 9.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RstaStanjeUsluge(ctx, 3, "1320", 15)
			return err
		}},
	{name: "rdok stopa of a tarifa", respond: answer(0, 20.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RdokStopaUsluge(ctx, 41, 3)
			return err
		}},
	{name: "rsif tip and stanje", respond: answer(0, "R", 3.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RsifTipStanje(ctx, 15, 3, " 1320 ")
			return err
		}},
	{name: "rsif id and tip", respond: answer(0, int64(77), "U"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RsifIDTip(ctx, tx, 15)
			return err
		}},
	{name: "dokusldet changed", respond: func(q string, _ []any) drvResult {
		if q[:6] == "select" {
			return drvResult{cols: []string{"count"}, rows: [][]any{{int64(1)}}}
		}
		return drvResult{affected: 1}
	}, inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DokusldetSnimi(ctx, tx, 41, "Opis")
		}},
	{name: "dokusldet added", respond: func(q string, _ []any) drvResult {
		if q[:6] == "select" {
			return drvResult{cols: []string{"count"}, rows: [][]any{{int64(0)}}}
		}
		return drvResult{affected: 1}
	}, inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DokusldetSnimi(ctx, tx, 41, "Opis")
		}},
	{name: "rpro stavka of a usluga to delete", respond: answer(0, int64(41), int64(15), "1320", 1.0, 10.0, "", "", 0.0, int64(2), "U"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproUslugaBrisanje(ctx, tx, 77)
			return err
		}},
	{name: "rdok take iznos", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokOduzmiIznos(ctx, tx, 41, 10)
		}},
	{name: "rpro form of a usluga", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproUslugaForma(ctx, 77, 41)
			return err
		}},
	{name: "rdok vrd of a stavka", respond: answer(0, int64(172)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RdokVrdStavke(ctx, 77)
			return err
		}},
	{name: "dokvrsta vrste of the group", respond: func(string, []any) drvResult {
		return drvResult{cols: []string{"vrd"}, rows: [][]any{{int64(170)}, {int64(171)}}}
	},
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.DokvrstaVrsteGrupe(ctx, 170, 1, 199)
			return err
		}},
}

func TestRobnoUslugeGolden(t *testing.T) {
	runGolden(t, "robno_usluge.golden", robnoUslugeCases)
}

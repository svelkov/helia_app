package repository

import (
	"context"
	"testing"
	"time"

	"helia/internal/infrastructure/db"
)

var robnoPrenosnicaCases = []goldenCase{
	{name: "initable value of the kar", respond: answer(0, "NE"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.IniVrednost(ctx, "DOZMENJAJCENU", "PRENOSNICA")
			return err
		}},
	{name: "rdok magacin1 of a prenosnica", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RdokPrenosnicaMagacin1(ctx, 41)
			return err
		}},
	{name: "magacin of a prenosnica", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagacinPrenosnice(ctx, 3)
			return err
		}},
	{name: "magkonto naziv konta", respond: answer(0, " Roba "),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagkontoNazivKonta(ctx, 3, " 1320 ")
			return err
		}},
	{name: "rdok by broj for update", respond: answer(0, int64(41), int64(42)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RdokPoBrojuZaIzmenu(ctx, tx, 3, 112, 12)
			return err
		}},
	{name: "magacini mag", respond: answer(0, int64(1), int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.MagaciniMag(ctx, tx, 3, 4)
			return err
		}},
	{name: "rdok next broj", respond: answer(0, int64(5)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RdokSledeciBroj(ctx, tx, 4, 113)
			return err
		}},
	{name: "rdok set kontra", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokPostaviKontru(ctx, tx, 41, 42)
		}},
	{name: "rdok kontra by vrdokid", respond: answer(0, int64(42)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RdokKontraPoVeziID(ctx, tx, 41)
			return err
		}},
	{name: "rdok kontra with its konto", respond: answer(0, int64(42), "1320"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RdokKontra(ctx, 41)
			return err
		}},
	{name: "rpro set dadok", respond: answer(3), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RproPostaviDadok(ctx, tx, 41, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
		}},
	{name: "rnal to copy", respond: answer(0, "P", int64(1), int64(7), nil, nil), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RnalKopija(ctx, tx, 9)
			return err
		}},
	{name: "rnal by broj", respond: answer(0, int64(10)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RnalPoBroju(ctx, tx, 1, "P", 4)
			return err
		}},
	{name: "drsta stanje of a lot", respond: answer(0, 5.0, 2.0, 9.5, 8.0, 8.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.DrstaLotStanje(ctx, "1320", 15, 1, " A1 ", " S1 ", 81234)
			return err
		}},
	{name: "rsta ulaz, izlaz and prosnc", respond: answer(0, 5.0, 2.0, 8.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, err := r.RstaUlazIzlazProsnc(ctx, tx, 4, "1320", 15)
			return err
		}},
	{name: "rsta set prosnc", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RstaPostaviProsnc(ctx, tx, 4, "1320", 15, 8.25)
		}},
	{name: "drsta set prosnc", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DrstaPostaviProsnc(ctx, tx, 2, 77, 8.25)
		}},
	{name: "rpro stavka with its lot to delete", respond: answer(0, int64(15), int64(2), "1320", "", "", 2.0, 19.0, 0.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproStavkaLotBrisanje(ctx, tx, 77)
			return err
		}},
	{name: "rpro stavka of the kontra", respond: answer(0, int64(78)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproKontraStavka(ctx, tx, 42, 77)
			return err
		}},
}

func TestRobnoPrenosnicaGolden(t *testing.T) {
	runGolden(t, "robno_prenosnica.golden", robnoPrenosnicaCases)
}

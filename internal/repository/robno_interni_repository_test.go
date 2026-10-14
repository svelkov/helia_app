package repository

import (
	"context"
	"testing"
	"time"

	"helia/internal/infrastructure/db"
)

var robnoInterniCases = []goldenCase{
	{name: "rvr fprod", respond: answer(0, "2040"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RvrFprod(ctx)
			return err
		}},
	{name: "rdok vrd", respond: answer(0, int64(172)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RdokVrd(ctx, 41)
			return err
		}},
	{name: "rdok iznos and brst for update", respond: answer(0, 100.0, int64(3)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RdokIznosBrst(ctx, tx, 41)
			return err
		}},
	{name: "rsta stanje and cene", respond: answer(0, 3.0, 9.5, 8.0, 10.0, 20.0, 12.0, 14.4),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RstaStanjeCene(ctx, 3, " 1320 ", 15)
			return err
		}},
	{name: "rsif artikal with stopa by šifra", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalSaStopom(ctx, false, "15", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
			return err
		}},
	{name: "rsif artikal with stopa by barkod", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalSaStopom(ctx, true, "8600123456789", nil)
			return err
		}},
	{name: "rsif id, naziv and jm", respond: answer(0, int64(77), "Vijak", "kom"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, err := r.RsifIDNazivJm(ctx, tx, 15)
			return err
		}},
	{name: "rcene of the artikal", respond: answer(0, int64(1)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RcenePostoji(ctx, tx, 15)
			return err
		}},
	{name: "rpro količina of the stavka", respond: answer(0, 2.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RproKolicina(ctx, 77, 15, "1320")
			return err
		}},
	{name: "rpro old promena for update", respond: answer(0, int64(15), "1320", 2.0, 19.0, "", "", 0.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproStaraPromena(ctx, tx, 77, 41)
			return err
		}},
	{name: "rpro rbr", respond: answer(0, int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproRbr(ctx, tx, 77)
			return err
		}},
	{name: "rpro renumber with brst", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RproPrenumerisiSaBrst(ctx, tx, 41, 2)
		}},
	{name: "rpro last cena of the prodavnica", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproZadnjaCenaProdavnice(ctx, 41, 15)
			return err
		}},
	{name: "rpro form of an interni stavka", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproInterniForma(ctx, 77, 41)
			return err
		}},
	// porudžbenica
	{name: "komercijalista sifkom", respond: answer(0, "12"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.KomercijalistaSifra(ctx, 6)
			return err
		}},
	{name: "rsta stanje and ncena", respond: answer(0, 3.0, 8.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RstaStanjeNcena(ctx, 3, " 1320 ", 15)
			return err
		}},
	{name: "rsif artikal with ncena by barkod", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalSaNcenom(ctx, true, "8600123456789", nil)
			return err
		}},
	{name: "rpro old količina for update", respond: answer(0, int64(15), "1320", 2.0, 16.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, _, err := r.RproStaraKolicina(ctx, tx, 77, 41)
			return err
		}},
	{name: "rpro rdokid", respond: answer(0, int64(41)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproRdokID(ctx, tx, 77)
			return err
		}},
	{name: "rpro form of a porudžbenica stavka", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproPorudzbenicaForma(ctx, 77, 41)
			return err
		}},
}

func TestRobnoInterniGolden(t *testing.T) {
	runGolden(t, "robno_interni.golden", robnoInterniCases)
}

package repository

import (
	"context"
	"testing"

	"helia/internal/infrastructure/db"
)

var robnoOpstiCases = []goldenCase{
	{name: "dokvrsta opis", respond: answer(0, "PROSECNA"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.DokvrstaOpis(ctx, 89)
			return err
		}},
	{name: "rdok protivstav", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RdokProtivstav(ctx, 41)
			return err
		}},
	{name: "orgjed first with its mesto troška", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.OrgjedPrvaSaMestom(ctx)
			return err
		}},
	{name: "magacin with the opis of the nalog", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagacinOpsti(ctx, 3, 9)
			return err
		}},
	{name: "fkpl naziv of a sintetički konto", respond: answer(0, "Roba"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.FkplNaziv(ctx, " 1320 ", " ")
			return err
		}},
	{name: "fkpl naziv of an analitika", respond: answer(0, "Kupac"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.FkplNaziv(ctx, "2040", " 15 ")
			return err
		}},
	{name: "rsta stanje of an opšti dokument", respond: answer(0, 5.0, 2.0, 50.0, 20.0, 9.5, 8.0, 8.5),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RstaStanjeOpsti(ctx, 3, " 1320 ", 15)
			return err
		}},
	{name: "drsta količina of a lot", respond: answer(0, 3.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.DrstaLotKolicina(ctx, 1, " 1320 ", 15, " A1 ", " S1 ", 81234)
			return err
		}},
	{name: "rsif artikal of an opšti dokument", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalOpsti(ctx, false, "15")
			return err
		}},
	{name: "rsif id, po and naziv", respond: answer(0, int64(77), int64(3), "Vijak"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, err := r.RsifIDPoNaziv(ctx, tx, 15)
			return err
		}},
	{name: "drsta delete an empty lot", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DrstaBrisiPrazanLot(ctx, tx, " 1320 ", 15, 1, " A1 ", " S1 ", 81234)
		}},
	{name: "drsta set prosnc of the šifra", respond: answer(2),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			return r.DrstaPostaviProsncSifre(ctx, 1, 15, 8.25)
		}},
	{name: "rdok one stavka more", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokDodajStavku(ctx, tx, 41)
		}},
}

func TestRobnoOpstiGolden(t *testing.T) {
	runGolden(t, "robno_opsti.golden", robnoOpstiCases)
}

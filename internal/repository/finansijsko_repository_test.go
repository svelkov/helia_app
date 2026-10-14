package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// fin is the repository of the finansijsko module over the database of a golden case.
func fin(r *RobnoRepository) *FinansijskoRepository { return NewFinansijskoRepository(r.db) }

var kopija = KopijaNaloga{Nalog: 12, Tipdok: "10", Danal: "2026-03-01", Datob: "2026-03-02", Opis: "Kopija", God: 2026, Kar: 7, Operater: "ana"}

var finansijskoCases = []goldenCase{
	{name: "fpro last rbr of the nalog", respond: answer(0, int64(4)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := fin(r).FproZadnjiRbr(ctx, 5)
			return err
		}},
	{name: "fpro sums of the nalog", respond: answer(0, 10.0, 10.0, 0.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := fin(r).FproSumeNaloga(ctx, 5)
			return err
		}},
	{name: "function exists", respond: answer(0, true),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := fin(r).FunkcijaPostoji(ctx, "fn_kopiraj_stavke_u_postojeci_nalog(bigint,bigint,text)")
			return err
		}},
	{name: "copy a nalog", respond: answer(0, int64(6)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := fin(r).FnKopirajNalog(ctx, 5, kopija, 3)
			return err
		}},
	{name: "copy the stavke into a nalog", respond: answer(0, "ok"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := fin(r).FnKopirajStavke(ctx, 5, 6, "ana")
			return err
		}},
	{name: "reverse a nalog", respond: answer(0, int64(7)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := fin(r).FnStornirajNalog(ctx, 5, kopija)
			return err
		}},
	{name: "fizvzag add", respond: answer(0, int64(11)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := fin(r).FizvzagDodaj(ctx, tx, &domain.Fizvzag{God: 2026, Kar: 7, Brrac: "160-1", Izvbr: 3, Konto: "2410"})
			return err
		}},
	{name: "fizvdet delete of the period of the header", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return fin(r).FizvdetBrisi(ctx, tx, 11, 2026, 7)
		}},
	{name: "fizvdet add two", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			d := sql.NullTime{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Valid: true}
			op := sql.NullString{String: "ana", Valid: true}
			return fin(r).FizvdetDodaj(ctx, tx, []domain.Fizvdet{
				{God: 2026, Kar: 7, Brrac: "160-1", Izvbr: 3, Rbr: 1, Konto: "2040", Sifra: "15", Iznos: 10, Xdatunosa: d, Xopunos: op},
				{God: 2026, Kar: 7, Brrac: "160-1", Izvbr: 3, Rbr: 2, Konto: "4350", Sifra: "8", Iznos: 5, Xdatunosa: d, Xopunos: op},
			}, 11)
		}},
	{name: "fizvdet add none", inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return fin(r).FizvdetDodaj(ctx, tx, nil, 11)
		}},
}

func TestFinansijskoGolden(t *testing.T) {
	runGolden(t, "finansijsko.golden", finansijskoCases)
}

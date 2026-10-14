package repository

import (
	"context"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var robnoIzvestajiCases = func() []goldenCase {
	p := domain.RobnoPrometParams{OdMagacina: "1", DoMagacina: "9", OdDatuma: "2026-03-01", DoDatuma: "2026-03-31", UlazIzlaz: PrometUlaz}
	d1 := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	return []goldenCase{
		{name: "promet ulaz per magacin and artikal", run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.PrometUlazIzlaz(ctx, p)
			return err
		}},
		{name: "promet ulaz i izlaz, sum of the magacini", run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			q := p
			q.ZbirMagacina, q.UlazIzlaz = true, ""
			_, err := r.PrometUlazIzlaz(ctx, q)
			return err
		}},
		{name: "promet izlaz", run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			q := p
			q.UlazIzlaz = PrometIzlaz
			_, err := r.PrometUlazIzlaz(ctx, q)
			return err
		}},
		{name: "vrste dokumenata of a vrsta naloga", run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.VrdGrupeNaloga(ctx, " 70 ")
			return err
		}},
		{name: "ncena: promet of the artikal", run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.NcenaPromet(ctx, 15, 3, d1, d1)
			return err
		}},
		{name: "ncena: first ulaz", run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.NcenaPrviUlaz(ctx, 15, 3, d1)
			return err
		}},
	}
}()

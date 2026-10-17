package repository

import (
	"context"
	"testing"
	"time"

	"helia/internal/infrastructure/db"
)

// robnoValuteCases are the statements of the valute and of the kursna lista (no period).
var robnoValuteCases = []goldenCase{
	{name: "valuta of the šifra", respond: answer(0, "EUR", 1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, found, err := r.Valuta(ctx, 978)
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "valuta not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, found, err := r.Valuta(ctx, 978)
			if found {
				return errFound
			}
			return err
		}},
	{name: "srednji kurs on the datum", respond: answer(0, 117.25),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := r.SrednjiKurs(ctx, 978, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC))
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "no kursna lista on the datum", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := r.SrednjiKurs(ctx, 978, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC))
			if found {
				return errFound
			}
			return err
		}},
}

func TestRobnoValuteGolden(t *testing.T) {
	runGolden(t, "robno_valute.golden", robnoValuteCases)
}

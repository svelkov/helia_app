package repository

import (
	"context"
	"testing"

	"helia/internal/infrastructure/db"
)

var robnoKnjiznoPismoCases = []goldenCase{
	// No rows: the struct of the row is scanned by its column names.
	{name: "rdok knjižno pismo",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RdokKnjiznoPismo(ctx, 41)
			return err
		}},
	{name: "rdok preracun of the knjižno pismo", respond: answer(0, 1000.0, 800.0, 50.0, int64(3), 2.0, 1.0, 750.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RdokKnpPreracun(ctx, tx, 41)
			return err
		}},
}

func TestRobnoKnjiznoPismoGolden(t *testing.T) {
	runGolden(t, "robno_knjiznopismo.golden", robnoKnjiznoPismoCases)
}

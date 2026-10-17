package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"helia/internal/infrastructure/db/querybuilder"
)

// The valute and the kursna lista (valute, klista). Both are common to all the periods: they have no god/kar.

// Valuta returns the oznaka and the "važi za" (the units the kurs is for) of the valuta of the šifra; found is false
// without it.
func (r *RobnoRepository) Valuta(ctx context.Context, sifval int) (oznval string, vaziza int, found bool, err error) {
	b := querybuilder.New(`select coalesce(v.oznval, ''), coalesce(v.vaziza, 0)::int from valute v`).
		Where("v.sifval = ?", sifval)
	b.OrderBy("v.idvalute").Limit(1)
	err = r.row(ctx, nil, "valuta", b, &oznval, &vaziza)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, false, nil
	}
	if err != nil {
		return "", 0, false, err
	}
	return oznval, vaziza, true, nil
}

// SrednjiKurs returns the srednji kurs of the valuta of the šifra on the datum (the kursna lista of that day, the legacy
// HReadSeekFirst(KLISTA,PKEY,[sifval,datum])); found is false without it.
func (r *RobnoRepository) SrednjiKurs(ctx context.Context, sifval int, datum time.Time) (float64, bool, error) {
	b := querybuilder.New(`select coalesce(k.srednji, 0)::float8 from klista k`).
		Where("k.sifval = ?", sifval).
		Where("k.datklista = ?::date", datum.Format("2006-01-02"))
	b.OrderBy("k.idklista").Limit(1)
	var kurs float64
	err := r.row(ctx, nil, "srednji kurs", b, &kurs)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return kurs, true, nil
}

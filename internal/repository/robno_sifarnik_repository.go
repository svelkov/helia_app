package repository

import (
	"context"
	"database/sql"
	"errors"

	"helia/internal/infrastructure/db/querybuilder"
)

// Lookups in the šifarnici of the robno module (magacini, kontni plan, konta of the magacini),
// always in the period of the session.

// MagacinMag returns the number (mag) of a magacin; the error wraps sql.ErrNoRows when it is not
// a magacin of the period.
func (r *RobnoRepository) MagacinMag(ctx context.Context, magaciniID int) (int, error) {
	b := querybuilder.New(`select mag from magacini`).Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var mag int
	err := r.row(ctx, nil, "magacin", b, &mag)
	return mag, err
}

// KontoZaliha returns the konto and the vrsta of a konto or an analitika (vkonta 1 or 2) of the
// kontni plan; the error wraps sql.ErrNoRows when there is none in the period.
func (r *RobnoRepository) KontoZaliha(ctx context.Context, idfkpl int) (konto string, vkonta int, err error) {
	b := querybuilder.New(`select trim(konto), vkonta from fkpl`).
		Where("idfkpl = ?", idfkpl).
		Where("vkonta in (1, 2)")
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", 0, err
	}
	err = r.row(ctx, nil, "konto zaliha", b, &konto, &vkonta)
	return konto, vkonta, err
}

// MagkontoDrugi counts the other rows (not magkontoID) of the magacin with the same konto zaliha.
func (r *RobnoRepository) MagkontoDrugi(ctx context.Context, magaciniID, idfkpl int, magkontoID int64) (int, error) {
	b := querybuilder.New(`select count(*) from magkonto`).
		Where("magaciniid = ?", magaciniID).
		Where("idfkpl = ?", idfkpl).
		Where("magkontoid <> ?", magkontoID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var n int
	err := r.row(ctx, nil, "konta magacina", b, &n)
	return n, err
}

// SintetickiKontoPostoji counts the konta (vkonta 2) of the kontni plan with this konto.
func (r *RobnoRepository) SintetickiKontoPostoji(ctx context.Context, konto string) (int, error) {
	b := querybuilder.New(`select count(*) from fkpl`).
		Where("vkonta = 2").
		Where("konto = ?", konto)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var n int
	err := r.row(ctx, nil, "konto", b, &n)
	return n, err
}

// MagacinTipmag returns the tip of a magacin (upper case, trimmed); "" when it is not a magacin of
// the period.
func (r *RobnoRepository) MagacinTipmag(ctx context.Context, magaciniID int) (string, error) {
	b := querybuilder.New(`select upper(trim(coalesce(tipmag, ''))) from magacini`).Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	var tip string
	err := r.row(ctx, nil, "tip magacina", b, &tip)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return tip, err
}

// RvrFdob returns the konto of the dobavljači of the vrste robnih dokumenata (rvr.fdob) of the period; "" without
// it.
func (r *RobnoRepository) RvrFdob(ctx context.Context) (string, error) {
	b := querybuilder.New(`select coalesce(fdob, '') from rvr`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	b.Limit(1)
	var fdob string
	err := r.row(ctx, nil, "konto dobavljača", b, &fdob)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return fdob, err
}

// RvrFprod returns the prodavnica of the vrste robnih dokumenata (rvr.fprod) of the period; "" without it.
func (r *RobnoRepository) RvrFprod(ctx context.Context) (string, error) {
	b := querybuilder.New(`select coalesce(fprod, '') from rvr`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	b.Limit(1)
	var fprod string
	err := r.row(ctx, nil, "prodavnica", b, &fprod)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return fprod, err
}

// KomercijalistaSifra returns the šifra (sifkom, as text) of a komercijalista; "" when it is not one
// of the period.
func (r *RobnoRepository) KomercijalistaSifra(ctx context.Context, komID int64) (string, error) {
	b := querybuilder.New(`select coalesce(sifkom, 0)::text from komercijalisti`).Where("komid = ?", komID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	var sifkom string
	err := r.row(ctx, nil, "komercijalista", b, &sifkom)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return sifkom, err
}

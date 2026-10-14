package repository

import (
	"context"
	"database/sql"
	"errors"

	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The robni dokumenti (rdok) of the period of the session.

// RdokPoBroju returns the document of a magacin by its vrsta and broj, with its number of stavke;
// found is false when there is none.
func (r *RobnoRepository) RdokPoBroju(ctx context.Context, magaciniID, vrd, dokum int) (rdokID int64, brojStavki int, found bool, err error) {
	b := querybuilder.New(`select r.rdokid, (select count(*) from rpro p where p.rdokid = r.rdokid) from rdok r`)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return 0, 0, false, err
	}
	b.Where("r.magaciniid = ?", magaciniID).Where("r.vrd = ?", vrd).Where("r.dokum = ?", dokum).Limit(1)
	err = r.row(ctx, nil, "dokument", b, &rdokID, &brojStavki)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return rdokID, brojStavki, err == nil, err
}

// RdokDodajIznos adds iznos (rounded to 2 decimals) and brst stavke to a document; ErrNotFound when
// it is not a document of the period.
func (r *RobnoRepository) RdokDodajIznos(ctx context.Context, tx db.Transaction, rdokID int64, iznos float64, brst int) error {
	b := querybuilder.NewUpdate("rdok").
		Set("iznos = round((coalesce(iznos, 0) + ?)::numeric, 2), brst = coalesce(brst, 0) + ?", iznos, brst).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "iznos dokumenta", b)
}

// RdokOduzmiStavku takes the iznos of a deleted stavka and one stavka (not below 0) from a
// document; ErrNotFound when it is not a document of the period.
func (r *RobnoRepository) RdokOduzmiStavku(ctx context.Context, tx db.Transaction, rdokID int64, iznos float64) error {
	b := querybuilder.NewUpdate("rdok").
		Set("iznos = round((coalesce(iznos, 0) - ?)::numeric, 2), brst = greatest(coalesce(brst, 0) - 1, 0)", iznos).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "iznos dokumenta", b)
}

// RdokProfakBi returns the profaktura (profak, the brojevi of the avansni računi) and the broj fiskalnog
// isečka (bi, as text) of a document; both are "" when it is not a document of the period.
func (r *RobnoRepository) RdokProfakBi(ctx context.Context, rdokID int64) (profak, bi string, err error) {
	b := querybuilder.New(`select coalesce(trim(profak), ''), coalesce(bi::text, '') from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", "", err
	}
	err = r.row(ctx, nil, "profaktura dokumenta", b, &profak, &bi)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	return profak, bi, err
}

// RdokPktoPana returns the konto and the šifra of the partner (pkto, pana, trimmed) of a document;
// both are "" when it is not a document of the period.
func (r *RobnoRepository) RdokPktoPana(ctx context.Context, rdokID int64) (pkto, pana string, err error) {
	b := querybuilder.New(`select coalesce(trim(pkto), ''), coalesce(trim(pana), '') from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", "", err
	}
	err = r.row(ctx, nil, "partner dokumenta", b, &pkto, &pana)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	return pkto, pana, err
}

// RdokPktoPanaStopa returns the konto and the šifra of the partner of a document and the PDV stopa
// of the poreska oznaka po on the datum of the document, for a new stavka.
func (r *RobnoRepository) RdokPktoPanaStopa(ctx context.Context, tx db.Transaction, rdokID int64, po int) (pkto, pana string, stopa float64, err error) {
	b := querybuilder.New(`select coalesce(pkto, ''), coalesce(pana, ''), `+poreskaStopa("?", "dadok")+` from rdok`, po).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", "", 0, err
	}
	err = r.row(ctx, tx, "dokument stavke", b, &pkto, &pana, &stopa)
	return pkto, pana, stopa, err
}

// RdokStopa returns the PDV stopa of the poreska oznaka po on the datum of a document.
func (r *RobnoRepository) RdokStopa(ctx context.Context, tx db.Transaction, rdokID int64, po int) (float64, error) {
	b := querybuilder.New(`select `+poreskaStopa("?", "dadok")+` from rdok`, po).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var stopa float64
	err := r.row(ctx, tx, "stopa dokumenta", b, &stopa)
	return stopa, err
}

// RdokDodajVrednosti adds a stavka to a document of a popis (or an opšti dokument): its prodajna
// and nabavna vrednost, its porez, its iznos (rounded to 2 decimals) and one stavka.
func (r *RobnoRepository) RdokDodajVrednosti(ctx context.Context, tx db.Transaction, rdokID int64, prodvred, nabvred, vporez, iznos float64) error {
	b := querybuilder.NewUpdate("rdok").
		Set(`prodvred = coalesce(prodvred, 0) + ?, nabvred = coalesce(nabvred, 0) + ?, vporez = coalesce(vporez, 0) + ?,
			iznos = round((coalesce(iznos, 0) + ?)::numeric, 2), brst = coalesce(brst, 0) + 1`, prodvred, nabvred, vporez, iznos).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "vrednosti dokumenta", b)
}

// RdokOduzmiVrednosti takes a deleted stavka from a document of a popis: its prodajna and nabavna
// vrednost, its porez, its iznos and one stavka (not below 0).
func (r *RobnoRepository) RdokOduzmiVrednosti(ctx context.Context, tx db.Transaction, rdokID int64, prodvred, nabvred, vporez, iznos float64) error {
	b := querybuilder.NewUpdate("rdok").
		Set(`prodvred = coalesce(prodvred, 0) - ?, nabvred = coalesce(nabvred, 0) - ?, vporez = coalesce(vporez, 0) - ?,
			iznos = round((coalesce(iznos, 0) - ?)::numeric, 2), brst = greatest(coalesce(brst, 0) - 1, 0)`, prodvred, nabvred, vporez, iznos).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "vrednosti dokumenta", b)
}

// RdokPktoPanaTx returns the konto and the šifra of the partner (pkto, pana) of a document in a
// transaction; the error wraps sql.ErrNoRows when it is not a document of the period.
func (r *RobnoRepository) RdokPktoPanaTx(ctx context.Context, tx db.Transaction, rdokID int64) (pkto, pana string, err error) {
	b := querybuilder.New(`select coalesce(pkto, ''), coalesce(pana, '') from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", "", err
	}
	err = r.row(ctx, tx, "partner dokumenta", b, &pkto, &pana)
	return pkto, pana, err
}

// RdokVrd returns the vrsta of a document; found is false when it is not a document of the period.
func (r *RobnoRepository) RdokVrd(ctx context.Context, rdokID int64) (vrd int, found bool, err error) {
	b := querybuilder.New(`select coalesce(vrd, 0) from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	err = r.row(ctx, nil, "vrsta dokumenta", b, &vrd)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return vrd, err == nil, err
}

// RdokIznosBrst reads (and locks) the iznos and the number of stavke of a document; the error wraps
// sql.ErrNoRows when it is not a document of the period.
func (r *RobnoRepository) RdokIznosBrst(ctx context.Context, tx db.Transaction, rdokID int64) (iznos float64, brst int, err error) {
	b := querybuilder.New(`select coalesce(iznos, 0)::float8, coalesce(brst, 0) from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, err
	}
	b.ForUpdate()
	err = r.row(ctx, tx, "iznos dokumenta", b, &iznos, &brst)
	return iznos, brst, err
}

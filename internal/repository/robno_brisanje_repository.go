package repository

import (
	"context"

	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The deletion of a robni dokument with what refers to it, in the period of the session.

// RdokBrisanje is what decides whether a document may be deleted.
type RdokBrisanje struct {
	RnalID                           int64
	Knjizen, Fiskalizovan, StatusSef string
	SefID, PepdvID                   float64
}

// RdokZaBrisanje reads (and locks) a document to delete: its nalog, whether it is posted (knjige_1)
// or fiscalized (knjige_4), its SEF id and status and its id of the pojedinačna evidencija PDV. The
// error wraps sql.ErrNoRows when it is not a document of the period.
func (r *RobnoRepository) RdokZaBrisanje(ctx context.Context, tx db.Transaction, rdokID int64) (RdokBrisanje, error) {
	b := querybuilder.New(`select coalesce(rnalid, 0)::bigint, coalesce(knjige_1, ''), coalesce(knjige_4, ''),
		coalesce(salesinvoiceid, 0)::float8, coalesce(status_salinv, ''), coalesce(individualvatid, 0)::float8 from rdok`).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RdokBrisanje{}, err
	}
	b.ForUpdate()
	var d RdokBrisanje
	err := r.row(ctx, tx, "dokument za brisanje", b, &d.RnalID, &d.Knjizen, &d.Fiskalizovan, &d.SefID, &d.StatusSef, &d.PepdvID)
	return d, err
}

// RproIDsDokumenta returns the ids of the stavke of a document in their order.
func (r *RobnoRepository) RproIDsDokumenta(ctx context.Context, rdokID int64) ([]int64, error) {
	b := querybuilder.New(`select rproid from rpro`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return nil, err
	}
	b.OrderBy("rproid")
	var ids []int64
	if err := r.selectRows(ctx, "stavke dokumenta", &ids, b); err != nil {
		return nil, err
	}
	return ids, nil
}

// brisiPoDokumentu deletes the rows of a table (with god/kar) that refer to a document.
func (r *RobnoRepository) brisiPoDokumentu(ctx context.Context, tx db.Transaction, what, table string, rdokID int64) error {
	b := querybuilder.New(`delete from `+table).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, what, b)
	return err
}

// AvansfaktFaktureBrisi deletes the avansi closed by a faktura.
func (r *RobnoRepository) AvansfaktFaktureBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.brisiPoDokumentu(ctx, tx, "avansi fakture", "avansfakt", rdokID)
}

// FaktrateBrisi deletes the rate of a faktura.
func (r *RobnoRepository) FaktrateBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.brisiPoDokumentu(ctx, tx, "rate fakture", "faktrate", rdokID)
}

// DokusldetBrisi deletes the dodatni opis usluge of a document.
func (r *RobnoRepository) DokusldetBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.brisiPoDokumentu(ctx, tx, "opis usluge", "dokusldet", rdokID)
}

// RnalOduzmiDokument counts one document less in a nalog (not below 0); ErrNotFound when it is not
// a nalog of the period.
func (r *RobnoRepository) RnalOduzmiDokument(ctx context.Context, tx db.Transaction, rnalID int64) error {
	b := querybuilder.NewUpdate("rnal").Set("brdo = greatest(coalesce(brdo, 0) - 1, 0)").Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "dokumenti naloga", b)
}

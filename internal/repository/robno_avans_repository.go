package repository

import (
	"context"
	"database/sql"
	"errors"

	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The avansi (avansfakt) of the avansni računi of the period of the session and the lookups of
// their forms.

// AvansfaktDodaj creates the avans of an avansni račun (no faktura yet, the osnovice and the PDV 0)
// when it has none.
func (r *RobnoRepository) AvansfaktDodaj(ctx context.Context, avansID int64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into avansfakt (god, kar, avansid, osnfakt1, pdv1fakt, osnfakt2, pdv2fakt, osnfakt3,
			osnavans1, pdv1avans, osnavans2, pdv2avans, osnavans3, xopunos, xdatunosa)
		select ?, ?, ?, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ?, now()`, s.God, s.Kar, avansID, s.UserName).
		Where("not exists (select 1 from avansfakt where avansid = ?)", avansID)
	_, err = r.exec(ctx, nil, "avans", b)
	return err
}

// AvansfaktDodajIznose adds the osnovice and the PDV of a stavka to the (first) avans of an avansni
// račun; nothing changes when it has none.
func (r *RobnoRepository) AvansfaktDodajIznose(ctx context.Context, tx db.Transaction, avansID int64, osn1, pdv1, osn2, pdv2, osn3 float64) error {
	b := querybuilder.NewUpdate("avansfakt").
		Set(`osnavans1 = coalesce(osnavans1, 0) + ?, pdv1avans = coalesce(pdv1avans, 0) + ?,
			osnavans2 = coalesce(osnavans2, 0) + ?, pdv2avans = coalesce(pdv2avans, 0) + ?, osnavans3 = coalesce(osnavans3, 0) + ?`,
			osn1, pdv1, osn2, pdv2, osn3).
		Where("avansfaktid = (select min(a.avansfaktid) from avansfakt a where a.avansid = ?)", avansID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "iznosi avansa", b)
	return err
}

// AvansfaktBrisi deletes the avans of an avansni račun.
func (r *RobnoRepository) AvansfaktBrisi(ctx context.Context, tx db.Transaction, avansID int64) error {
	b := querybuilder.New(`delete from avansfakt`).Where("avansid = ?", avansID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "brisanje avansa", b)
	return err
}

// FvknjracKontaKupca returns the analitička konta the kupac of an avansni račun must have (the konta
// of the izlazni računi, fvknjrac I 3, separated by commas) and the partner (idpartneri as text) of
// the konto and šifra of the kupac; idpartneri is "" when the kupac is not in the kontni plan.
func (r *RobnoRepository) FvknjracKontaKupca(ctx context.Context, konto, sifra string) (konta, idpartneri string, err error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return "", "", err
	}
	b := querybuilder.New(`select coalesce((select f.konta from fvknjrac f where f.god = ? and f.kar = ? and f.vktip = 'I'
			and f.vkrbr = 3 order by f.vkrbr limit 1), ''),
			coalesce((select p.idpartneri::text from fkpl p where p.god = ? and p.kar = ? and p.konto = ? and p.sifra = ? limit 1), '')`,
		s.God, s.Kar, s.God, s.Kar, konto, sifra)
	err = r.row(ctx, nil, "konta kupca", b, &konta, &idpartneri)
	return konta, idpartneri, err
}

// RdokProfak returns the profaktura of a document; "" when it is not a document of the period.
func (r *RobnoRepository) RdokProfak(ctx context.Context, rdokID int64) (string, error) {
	b := querybuilder.New(`select coalesce(profak, '') from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	var profak string
	err := r.row(ctx, nil, "profaktura", b, &profak)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return profak, err
}

// RdokStopaNapomena returns the PDV stopa of the poreska oznaka po on the datum of a document and
// its napomena o poreskom oslobođenju.
func (r *RobnoRepository) RdokStopaNapomena(ctx context.Context, tx db.Transaction, rdokID int64, po int) (stopa float64, napomena string, err error) {
	b := querybuilder.New(`select `+poreskaStopa("?", "dadok")+`, coalesce(pornapomena, '') from rdok`, po).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", err
	}
	err = r.row(ctx, tx, "stopa dokumenta", b, &stopa, &napomena)
	return stopa, napomena, err
}

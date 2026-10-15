package repository

import (
	"context"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The knjižna pisma (rdok of the groups KNO and KNZ): the columns of the screen and the legacy Preracun.

// RdokKnjiznoPismo returns the ugovoreni rabat and the kasa (% and vrednosti), the iznos, the porez, the vezni
// dokument and the tip of the knjižno pismo with its text ("šifra razloga - opis") of a document; no row when it
// is not a document of the period.
func (r *RobnoRepository) RdokKnjiznoPismo(ctx context.Context, rdokID int64) ([]domain.RobKnpDokumentDto, error) {
	b := querybuilder.New(`select coalesce(r.ugrabat, 0)::float8 as ugrabat, coalesce(r.vrugrabat, 0)::float8 as vrugrabat,
			coalesce(r.pkase, 0)::float8 as pkase, coalesce(r.vkase, 0)::float8 as vkase, coalesce(r.iznos, 0)::float8 as iznos,
			coalesce(r.vporez, 0)::float8 as vporez, coalesce(r.vrdokid, 0)::bigint as vrdokid, t.tipknjid::bigint as tipknjid,
			coalesce(concat_ws(' - ', t.sifrazlog::text, nullif(t.opis, '')), '') as tiptext
		from rdok r`).
		AddJoin("left join tipknpisma t on t.tipknjid = r.tipknjid").
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.RobKnpDokumentDto
	if err := r.selectRows(ctx, "knjižno pismo", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// RdokKnpPreracun reads in the transaction what the legacy Preracun of a knjižno pismo needs: the sums of its
// stavke (količina by the fakturna cena, by the cena, and the rabat of the fakturna vrednost), their number and
// the ugovoreni rabat, the kasa and the iznos of the document; the error wraps sql.ErrNoRows when it is not a
// document of the period.
func (r *RobnoRepository) RdokKnpPreracun(ctx context.Context, tx db.Transaction, rdokID int64) (domain.RobKnpPreracunDto, error) {
	b := querybuilder.New(`select
			coalesce((select sum(coalesce(p.kolic, 0) * coalesce(p.fcena, 0)) from rpro p where p.rdokid = r.rdokid), 0)::float8,
			coalesce((select sum(coalesce(p.kolic, 0) * coalesce(p.cena, 0)) from rpro p where p.rdokid = r.rdokid), 0)::float8,
			coalesce((select sum(coalesce(p.kolic, 0) * coalesce(p.fcena, 0) * coalesce(p.rab, 0) / 100) from rpro p
				where p.rdokid = r.rdokid), 0)::float8,
			(select count(*) from rpro p where p.rdokid = r.rdokid)::int,
			coalesce(r.ugrabat, 0)::float8, coalesce(r.pkase, 0)::float8, coalesce(r.iznos, 0)::float8
		from rdok r`).
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return domain.RobKnpPreracunDto{}, err
	}
	var p domain.RobKnpPreracunDto
	err := r.row(ctx, tx, "obračun knjižnog pisma", b, &p.Fakturna, &p.Cena, &p.Rabat, &p.Stavke, &p.Ugrabat, &p.Pkase, &p.Iznos)
	return p, err
}

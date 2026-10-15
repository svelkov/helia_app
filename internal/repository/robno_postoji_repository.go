package repository

import (
	"context"
	"errors"
	"strings"

	"helia/internal/infrastructure/db/querybuilder"
)

// The checks of the forms of the robno module: does a row exist. Tables with god/kar are restricted
// to the period of the session; porkat and valute are catalogs without a period.

// postoji tells whether the count query b finds at least one row.
func (r *RobnoRepository) postoji(ctx context.Context, what string, b *querybuilder.Builder) (bool, error) {
	var n int
	if err := r.row(ctx, nil, what, b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// postojiUPeriodu is postoji for a count query on a table with god/kar.
func (r *RobnoRepository) postojiUPeriodu(ctx context.Context, what string, b *querybuilder.Builder) (bool, error) {
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	return r.postoji(ctx, what, b)
}

// RsifPostoji tells whether the artikal with the šifra exists.
func (r *RobnoRepository) RsifPostoji(ctx context.Context, sifra int) (bool, error) {
	return r.postojiUPeriodu(ctx, "artikal", querybuilder.New(`select count(*) from rsif`).Where("sifra = ?", sifra))
}

// FkplAnalitikaPostoji tells whether the kontni plan has the konto with the šifra.
func (r *RobnoRepository) FkplAnalitikaPostoji(ctx context.Context, konto, sifra string) (bool, error) {
	return r.postojiUPeriodu(ctx, "analitika", querybuilder.New(`select count(*) from fkpl`).
		Where("konto = ?", konto).Where("sifra = ?", sifra))
}

// FkplAnalitickiKontoPostoji tells whether the kontni plan has the analitički konto (vkonta 1) with
// the šifra.
func (r *RobnoRepository) FkplAnalitickiKontoPostoji(ctx context.Context, konto, sifra string) (bool, error) {
	return r.postojiUPeriodu(ctx, "analitika", querybuilder.New(`select count(*) from fkpl`).
		Where("vkonta = 1").Where("konto = ?", konto).Where("sifra = ?", sifra))
}

// FkplAnalitikaTrimPostoji tells whether the kontni plan has the konto with the šifra, both compared
// trimmed (the šifra "" for a konto without one).
func (r *RobnoRepository) FkplAnalitikaTrimPostoji(ctx context.Context, konto, sifra string) (bool, error) {
	return r.postojiUPeriodu(ctx, "analitika", querybuilder.New(`select count(*) from fkpl`).
		Where("trim(konto) = ?", konto).Where("coalesce(trim(sifra), '') = ?", sifra))
}

// FkplKontoPartneraPostoji tells whether the partner has the analitički konto (vkonta 1).
func (r *RobnoRepository) FkplKontoPartneraPostoji(ctx context.Context, konto string, idpartneri int) (bool, error) {
	return r.postojiUPeriodu(ctx, "konto partnera", querybuilder.New(`select count(*) from fkpl`).
		Where("konto = ?", konto).Where("idpartneri = ?", idpartneri).Where("vkonta = 1"))
}

// AvansZatvoren tells whether the avans of an avansni račun is closed by a faktura.
func (r *RobnoRepository) AvansZatvoren(ctx context.Context, avansID int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "zatvoren avans", querybuilder.New(`select count(*) from avansfakt`).
		Where("avansid = ?", avansID).Where("coalesce(rdokid, 0) <> 0"))
}

// RdokBrojPostoji tells whether the magacin has a document of the vrsta with the broj.
func (r *RobnoRepository) RdokBrojPostoji(ctx context.Context, magaciniID, vrd, dokum int) (bool, error) {
	return r.postojiUPeriodu(ctx, "broj dokumenta", querybuilder.New(`select count(*) from rdok`).
		Where("magaciniid = ?", magaciniID).Where("vrd = ?", vrd).Where("dokum = ?", dokum))
}

// RdokBrojDrugogNalogaPostoji tells whether the magacin has another document (not rdokID, not of
// the nalog rnalID) of the vrsta with the broj.
func (r *RobnoRepository) RdokBrojDrugogNalogaPostoji(ctx context.Context, magaciniID, vrd, dokum int, rnalID, rdokID int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "broj dokumenta", querybuilder.New(`select count(*) from rdok`).
		Where("magaciniid = ?", magaciniID).Where("vrd = ?", vrd).Where("dokum = ?", dokum).
		Where("coalesce(rnalid, 0) <> ?", rnalID).Where("rdokid <> ?", rdokID))
}

// RdokAvansniDrugogKupca tells whether one of the avansni računi (vrsta 172) with the brojevi is of
// another kupac than the konto and šifra.
func (r *RobnoRepository) RdokAvansniDrugogKupca(ctx context.Context, brojevi []int, konto, sifra string) (bool, error) {
	if len(brojevi) == 0 {
		return false, errors.New("avansni računi: no brojevi")
	}
	ph := strings.TrimSuffix(strings.Repeat("?, ", len(brojevi)), ", ")
	return r.postojiUPeriodu(ctx, "avansni računi kupca", querybuilder.New(`select count(*) from rdok`).
		Where("vrd = 172").Where("dokum in ("+ph+")", querybuilder.Args(brojevi)...).
		Where("(coalesce(fkto, '') <> ? or coalesce(fana, '') <> ?)", konto, sifra))
}

// OrgjedPostoji tells whether the org. jedinica exists.
func (r *RobnoRepository) OrgjedPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "org. jedinica", querybuilder.New(`select count(*) from orgjed`).Where("idorgjed = ?", id))
}

// TipknpismaPostoji tells whether the tip of the knjižno pismo exists.
func (r *RobnoRepository) TipknpismaPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "tip knjižnog pisma", querybuilder.New(`select count(*) from tipknpisma`).Where("tipknjid = ?", id))
}

// MestotrPostoji tells whether the mesto troška exists.
func (r *RobnoRepository) MestotrPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "mesto troška", querybuilder.New(`select count(*) from mestotr`).Where("mestotrid = ?", id))
}

// MestotrUOrgjed tells whether the mesto troška is of the org. jedinica.
func (r *RobnoRepository) MestotrUOrgjed(ctx context.Context, mestotrID, orgjedID int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "mesto troška", querybuilder.New(`select count(*) from mestotr`).
		Where("mestotrid = ?", mestotrID).Where("idorgjed = ?", orgjedID))
}

// KomercijalistaPostoji tells whether the komercijalista exists.
func (r *RobnoRepository) KomercijalistaPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "komercijalista", querybuilder.New(`select count(*) from komercijalisti`).Where("komid = ?", id))
}

// BnkizvPostoji tells whether the bankovni izvod exists.
func (r *RobnoRepository) BnkizvPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "bankovni izvod", querybuilder.New(`select count(*) from bnkizv`).Where("bnkizvid = ?", id))
}

// FispPostoji tells whether the mesto isporuke exists and is of the partner (konto, šifra).
func (r *RobnoRepository) FispPostoji(ctx context.Context, id int64, konto, sifra string) (bool, error) {
	return r.postojiUPeriodu(ctx, "mesto isporuke", querybuilder.New(`select count(*) from fisp`).
		Where("fispid = ?", id).Where("konto = ?", konto).Where("sifra = ?", sifra))
}

// PorkatPostoji tells whether the poreska kategorija exists (a catalog without a period).
func (r *RobnoRepository) PorkatPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postoji(ctx, "poreska kategorija", querybuilder.New(`select count(*) from porkat`).Where("porkatid = ?", id))
}

// RporPostoji tells whether the poreska oznaka (tarifa) has a stopa (rpor, a catalog without a period).
func (r *RobnoRepository) RporPostoji(ctx context.Context, po int) (bool, error) {
	return r.postoji(ctx, "poreska tarifa", querybuilder.New(`select count(*) from rpor`).Where("po = ?", po))
}

// MagacinPostoji tells whether the magacin exists.
func (r *RobnoRepository) MagacinPostoji(ctx context.Context, id int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "magacin", querybuilder.New(`select count(*) from magacini`).Where("magaciniid = ?", id))
}

// MagacinPoNabavnoj tells whether the magacin keeps its zalihe by the nabavna cena (nacvodzal 1).
func (r *RobnoRepository) MagacinPoNabavnoj(ctx context.Context, magaciniID int) (bool, error) {
	return r.postojiUPeriodu(ctx, "način vođenja zaliha", querybuilder.New(`select count(*) from magacini`).
		Where("magaciniid = ?", magaciniID).Where("coalesce(nacvodzal, 0) = 1"))
}

// MagkontoPostoji tells whether the magacin has the konto (compared trimmed).
func (r *RobnoRepository) MagkontoPostoji(ctx context.Context, magaciniID int, konto string) (bool, error) {
	return r.postojiUPeriodu(ctx, "konto magacina", querybuilder.New(`select count(*) from magkonto`).
		Where("magaciniid = ?", magaciniID).Where("trim(konto) = ?", konto))
}

// DokvrstaProsecna tells whether the vrsta of the documents is one of the prosečne cene (its opis
// has PROS).
func (r *RobnoRepository) DokvrstaProsecna(ctx context.Context, vrd int) (bool, error) {
	return r.postojiUPeriodu(ctx, "vrsta dokumenta", querybuilder.New(`select count(*) from dokvrsta`).
		Where("vrd = ?", vrd).Where("opis like '%PROS%'"))
}

// DokvrstaAvansnaGodine tells whether the vrsta of the documents of the year god (of the kar of the
// session) is one of the avansni računi (grupa ARA or ARU). The year is the one of an izvorni
// dokument, which may be another than the one of the session.
func (r *RobnoRepository) DokvrstaAvansnaGodine(ctx context.Context, god, vrd int) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	return r.postoji(ctx, "vrsta dokumenta", querybuilder.New(`select count(*) from dokvrsta`).
		Where("god = ?", god).Where("kar = ?", s.Kar).Where("vrd = ?", vrd).
		Where("upper(trim(coalesce(grpdok, ''))) in ('ARA', 'ARU')"))
}

// FvrObveznikPdv tells whether the firma (by its naziv) is a PDV obveznik in the period.
func (r *RobnoRepository) FvrObveznikPdv(ctx context.Context, firma string) (bool, error) {
	return r.postojiUPeriodu(ctx, "obveznik PDV", querybuilder.New(`select count(*) from fvr`).
		Where("naziv = ?", firma).Where("coalesce(obv, false)"))
}

// PopdvPoljePostoji tells whether the polje of the POPDV exists.
func (r *RobnoRepository) PopdvPoljePostoji(ctx context.Context, polje string) (bool, error) {
	return r.postojiUPeriodu(ctx, "polje POPDV", querybuilder.New(`select count(*) from popdv`).Where("polje = ?", polje))
}

// PopdvPoljeNeodgovarajuce tells whether the polje of the POPDV is not one of the izlazni računi:
// not of the tip 3, or of an evidencija PDV that is not of the vktip I.
func (r *RobnoRepository) PopdvPoljeNeodgovarajuce(ctx context.Context, polje string) (bool, error) {
	b := querybuilder.New(`select count(*) from popdv p`).AddJoin("left join fvepdv e on e.fvepdvid = p.fvepdvid")
	if err := inPeriod(ctx, b, "p."); err != nil {
		return false, err
	}
	b.Where("p.polje = ?", polje).
		Where("(coalesce(p.tip, 0) <> 3 or (e.fvepdvid is not null and coalesce(e.vktip, '') <> 'I'))")
	return r.postoji(ctx, "polje POPDV", b)
}

// ValutaStrana tells whether the valuta is another one than the valuta of the firma in the period
// (valute is a catalog without a period).
func (r *RobnoRepository) ValutaStrana(ctx context.Context, id int64) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	return r.postoji(ctx, "strana valuta", querybuilder.New(`select count(*) from valute v`).
		Where("v.idvalute = ?", id).
		Where("coalesce(v.sifval, 0) <> coalesce((select f.sifval from fvr f where f.god = ? and f.kar = ? limit 1), 0)", s.God, s.Kar))
}

// FkplSintetikaPostoji tells whether the kontni plan has the sintetički konto (vkonta 2).
func (r *RobnoRepository) FkplSintetikaPostoji(ctx context.Context, konto string) (bool, error) {
	n, err := r.SintetickiKontoPostoji(ctx, konto)
	return n > 0, err
}

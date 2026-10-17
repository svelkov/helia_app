package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The porudžbenice dobavljačima of the komercijalno (the legacy pr_PorudzbenicaDOBAVLJ): edok of the vrste of the modul
// PRO and the group NRD for one dobavljač, with their stavke in epro, taken by hand or from the stavke of a potvrđen
// zahtev za ponudu (vrdokid).

// komPorGrupa is the group of the vrste dokumenata of the porudžbenice dobavljačima (the legacy PR_QRY_DOKVRSTA with
// "NRD").
const komPorGrupa = "NRD"

// KomPorVrste returns the vrste dokumenata of the porudžbenice dobavljačima of the period (dokvrsta of the modul PRO and
// the group NRD) as "vrd - opis" keyed by the vrd, in the order of the vrd.
func (r *KomercijalnoRepository) KomPorVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return r.komVrste(ctx, komPorGrupa)
}

// KomPorNoviBroj returns the next broj of a porudžbenica of the vrsta in the magacin in the period (the legacy
// NoviBrojEDOK of pr_PorudzbenicaDOBAVLJ: the largest dokum of the vrsta and the magacin + 1).
func (r *KomercijalnoRepository) KomPorNoviBroj(ctx context.Context, tx db.Transaction, vrd, mag int) (int, error) {
	b := querybuilder.New(`select coalesce(max(dokum), 0)::int + 1 from edok`).Where("vrd = ?", vrd).Where("mag = ?", mag)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var broj int
	if err := r.row(ctx, tx, "novi broj porudžbenice", b, &broj); err != nil {
		return 0, err
	}
	return broj, nil
}

// KomPorudzbenica returns a porudžbenica of the period with the texts of its magacin, komercijalista, org. jedinica,
// mesto troška and valuta and the broj and datum of its zahtev za ponudu; found is false without it.
func (r *KomercijalnoRepository) KomPorudzbenica(ctx context.Context, rdokID int64) (domain.KomPorudzbenicaRow, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.vrd, 0)::int, coalesce(e.dokum, 0)::int, e.dadok, coalesce(e.rok, 0)::int,
			coalesce(e.magaciniid, 0)::int, coalesce(e.mag, 0)::int, coalesce(e.mag::text || '-' || coalesce(m.opis, ''), ''), coalesce(m.adresa, ''),
			coalesce(e.fkto, ''), coalesce(e.fana, ''), e.komid::bigint, coalesce(k.sifkom, 0)::int, coalesce(k.imeprezime, ''),
			e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''),
			coalesce(e.foot, ''), coalesce(e.doksts, ''), coalesce(e.knjige_1, ''), coalesce(e.knjige_4, ''),
			coalesce(e.sifval, 0)::int, coalesce((select v.oznval from valute v where v.sifval = e.sifval order by v.idvalute limit 1), ''),
			coalesce(e.kurs, 0)::float8, coalesce(e.vaziza, 0)::int, coalesce(e.vrdokid, 0)::bigint, coalesce(z.dokum, 0)::int, z.dadok,
			coalesce(e.otp, ''), coalesce(e.pla, ''), coalesce(e.brugov, ''), coalesce(e.nacpak, ''), coalesce(e.posuslnab, ''),
			coalesce(e.pratdok, '')
		from edok e`).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		AddJoin("left join edok z on z.rdokid = e.vrdokid and z.god = e.god and z.kar = e.kar").
		Where("e.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomPorudzbenicaRow{}, false, err
	}
	var p domain.KomPorudzbenicaRow
	err := r.row(ctx, nil, "porudžbenica", b, &p.RdokID, &p.Vrd, &p.Dokum, &p.Dadok, &p.Rok, &p.MagaciniID, &p.Mag, &p.MagacinText,
		&p.MagacinAdresa, &p.Fkto, &p.Fana, &p.KomID, &p.KomSifra, &p.KomIme, &p.Orgjed, &p.OrgjedText, &p.MestoTroska, &p.MestoTroskaText,
		&p.Foot, &p.Doksts, &p.Knjige1, &p.Knjige4, &p.SifVal, &p.OznVal, &p.Kurs, &p.VaziZa, &p.VrdokID, &p.VezaDokum, &p.VezaDatum,
		&p.Otp, &p.Pla, &p.Brugov, &p.Nacpak, &p.Posuslnab, &p.Pratdok)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPorudzbenicaRow{}, false, nil
	}
	if err != nil {
		return domain.KomPorudzbenicaRow{}, false, err
	}
	return p, true, nil
}

// KomPorDodaj adds a porudžbenica (the legacy SnimiEDOK of a new document): not of a nalog, nepotvrđena (doksts 10),
// not posted (knjige N), its opis "Porudzbenica br: <broj>", the dobavljač, the uslovi nabavke, the valuta with its
// kurs and the zahtev za ponudu (vrdokid), the JM of the ambalaža empty (jm1 and jm2 are not null without a default),
// with the audit columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomPorDodaj(ctx context.Context, tx db.Transaction, p domain.KomPorudzbenicaUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana,
			pla, otp, foot, brst, vrdokid, idorgjed, mestotrid, komid, sifval, kurs, vaziza, brugov, nacpak, posuslnab, pratdok, doksts,
			knjige_1, knjige_2, jm1, jm2, xdatunosa, xopunos)
		values (?, ?, 0, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'N', 'N', '', '', now(), ?)`,
		s.God, s.Kar, p.Vrd, p.Dokum, p.Dadok, p.Mag, p.MagaciniID, fmt.Sprintf("Porudzbenica br: %d", p.Dokum), p.Rok, p.Fkto, p.Fana,
		p.Pla, p.Otp, p.Foot, p.VrdokID, p.Orgjed, p.MestoTroska, p.KomID, p.SifVal, p.Kurs, p.VaziZa, p.Brugov, p.Nacpak, p.Posuslnab,
		p.Pratdok, domain.KomZahtevNepotvrdjen, s.UserName).
		Returning("rdokid")
	var rdokID int64
	if err := r.row(ctx, tx, "nova porudžbenica", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomPorIzmeni saves the header of a porudžbenica of the period (the legacy SnimiEDOK of a saved document): its datum,
// rok, dobavljač, komercijalista, org. jedinica, mesto troška, tekst, valuta with its kurs and uslovi nabavke, with the
// audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomPorIzmeni(ctx context.Context, tx db.Transaction, p domain.KomPorudzbenicaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("dadok = ?", p.Dadok).
		Set("rok = ?", p.Rok).
		Set("fkto = ?", p.Fkto).
		Set("fana = ?", p.Fana).
		Set("komid = ?", p.KomID).
		Set("idorgjed = ?", p.Orgjed).
		Set("mestotrid = ?", p.MestoTroska).
		Set("foot = ?", p.Foot).
		Set("sifval = ?", p.SifVal).
		Set("kurs = ?", p.Kurs).
		Set("vaziza = ?", p.VaziZa).
		Set("otp = ?", p.Otp).
		Set("pla = ?", p.Pla).
		Set("brugov = ?", p.Brugov).
		Set("nacpak = ?", p.Nacpak).
		Set("posuslnab = ?", p.Posuslnab).
		Set("pratdok = ?", p.Pratdok).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", p.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "porudžbenica", b)
}

// KomPorPotvrdi confirms a nepotvrđena porudžbenica of the period (the legacy BTN_NoviDOK1 "Potvrda dok.": doksts 40).
// It returns false when the porudžbenica is not nepotvrđena any more.
func (r *KomercijalnoRepository) KomPorPotvrdi(ctx context.Context, rdokID int64) (bool, error) {
	return r.KomZahtevPotvrdi(ctx, rdokID)
}

// KomPorBrisi deletes a porudžbenica of the period with its stavke (the legacy BTN_Brisanje): the stavke of the zahtev
// za ponudu they were taken from are slobodne again; ErrNotFound without the porudžbenica.
func (r *KomercijalnoRepository) KomPorBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.KomPonudaBrisi(ctx, tx, rdokID)
}

// KomPorVeza returns the potvrđen zahtev za ponudu (doksts 40, the vrste of the group ZPN) of the period a porudžbenica
// is linked to (the legacy BTN_TraziDOK); found is false without it.
func (r *KomercijalnoRepository) KomPorVeza(ctx context.Context, rdokID int64) (domain.KomPorVeza, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.dokum, 0)::int, e.dadok, coalesce(e.rok, 0)::int, coalesce(e.otp, ''),
			coalesce(e.magaciniid, 0)::int, e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''), coalesce(e.fkto, ''), coalesce(e.fana, '')
		from edok e`).
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		Where("e.rdokid = ?", rdokID).
		Where("coalesce(e.doksts, '') = ?", domain.KomZahtevPotvrdjen).
		Where(`e.vrd in (select v.vrd from dokvrsta v where v.god = e.god and v.kar = e.kar
			and upper(trim(coalesce(v.modul, ''))) = ? and upper(trim(coalesce(v.grpdok, ''))) = ?)`, komZahtevModul, komPonudaGrupa)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomPorVeza{}, false, err
	}
	var v domain.KomPorVeza
	err := r.row(ctx, nil, "zahtev za ponudu porudžbenice", b, &v.RdokID, &v.Dokum, &v.Dadok, &v.Rok, &v.Otp, &v.MagaciniID, &v.Orgjed,
		&v.OrgjedText, &v.MestoTroska, &v.MtText, &v.Fkto, &v.Fana)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPorVeza{}, false, nil
	}
	if err != nil {
		return domain.KomPorVeza{}, false, err
	}
	return v, true, nil
}

// KomPorDobavljac returns the dobavljač of the konto and šifra of the period (the analitika of fkpl with its partner;
// the naziv of the analitika without a partner); found is false without it.
func (r *KomercijalnoRepository) KomPorDobavljac(ctx context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error) {
	b := querybuilder.New(`select coalesce(p.naziv, f.naziv, ''), coalesce(p.adresa, ''), coalesce(nullif(p.pobro, 0)::text, ''),
			coalesce(p.mesto, ''), coalesce(p.pib, ''), coalesce(p.tippdv, 0)::int
		from fkpl f`).
		AddJoin("left join partneri p on p.idpartneri = f.idpartneri").
		Where("f.vkonta = 1").
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra))
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomDobavljacPodaci{}, false, err
	}
	b.OrderBy("f.idfkpl").Limit(1)
	var d domain.KomDobavljacPodaci
	err := r.row(ctx, nil, "dobavljač", b, &d.Naziv, &d.Adresa, &d.Pobro, &d.Mesto, &d.Pib, &d.TipPdv)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomDobavljacPodaci{}, false, nil
	}
	if err != nil {
		return domain.KomDobavljacPodaci{}, false, err
	}
	return d, true, nil
}

// KomPorIDFkpl returns the id of the analitika of the konto and šifra of the period (the dobavljač of the print); 0
// without it.
func (r *KomercijalnoRepository) KomPorIDFkpl(ctx context.Context, konto, sifra string) (int64, error) {
	b := querybuilder.New(`select f.idfkpl::bigint from fkpl f`).
		Where("f.vkonta = 1").
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra))
	if err := inPeriod(ctx, b, "f."); err != nil {
		return 0, err
	}
	b.OrderBy("f.idfkpl").Limit(1)
	var id int64
	err := r.row(ctx, nil, "analitika dobavljača", b, &id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// KomPorStavke returns the stavke of a porudžbenica of the period in the order of their redni broj, with the naziv,
// the JM and the komercijalni opis of their artikal and the decimals of the JM (jedmere.brdecimala, 2 without it; the
// legacy PuniDETALJ1 and PR_QRY_EPRO).
func (r *KomercijalnoRepository) KomPorStavke(ctx context.Context, rdokID int64) ([]domain.KomPorStavkaRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rbr, 0)::int as rbr, coalesce(p.sifra, 0)::int as sifra,
			coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm, coalesce(s.komercopis, '') as komercopis,
			coalesce(p.trazkolic, 0)::float8 as kolicina, coalesce(p.cena, 0)::float8 as cena, coalesce(p.cenaval, 0)::float8 as cenaval,
			coalesce(p.rab, 0)::float8 as rabat, coalesce(p.rproid1, 0)::bigint as rproid1,
			coalesce((select j.brdecimala from jedmere j where j.jm = s.jm and j.god = p.god and j.kar = p.kar limit 1), 2)::int as decimale
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []komPorStavkaDto
	if err := r.selectRows(ctx, "stavke porudžbenice", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPorStavkaRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPorStavkaRow{RproID: row.RproID, Rbr: row.Rbr, Sifra: row.Sifra, Naziv: row.Naziv, JM: row.JM, Komercopis: row.Komercopis,
			Kolicina: row.Kolicina, Cena: row.Cena, CenaVal: row.CenaVal, Rabat: row.Rabat, RproID1: row.RproID1, Decimale: row.Decimale}
	}
	return out, nil
}

// komPorStavkaDto is a row of KomPorStavke.
type komPorStavkaDto struct {
	RproID     int64   `db:"rproid"`
	Rbr        int     `db:"rbr"`
	Sifra      int     `db:"sifra"`
	Naziv      string  `db:"naziv"`
	JM         string  `db:"jm"`
	Komercopis string  `db:"komercopis"`
	Kolicina   float64 `db:"kolicina"`
	Cena       float64 `db:"cena"`
	CenaVal    float64 `db:"cenaval"`
	Rabat      float64 `db:"rabat"`
	RproID1    int64   `db:"rproid1"`
	Decimale   int     `db:"decimale"`
}

// KomPorStavkaDodaj adds a stavka of a porudžbenica (the legacy AddEPRO for a stavka entered by hand, AddEPRONAR for a
// stavka of the zahtev za ponudu with its rproid in rproid1): the artikal, the količina, the cena in RSD (cena and
// ncena) and in the valuta (cenaval), the rabat, the magacin and the vrsta, broj and datum of the document, with the
// audit columns.
func (r *KomercijalnoRepository) KomPorStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomPorStavkaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, mag, magaciniid, trazkolic, cena, ncena, cenaval, rab,
			rproid1, vrd, dokum, dadok, xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, st.RdokID, st.Sifra, st.Po, st.RsifID, st.Rbr, st.Mag, st.MagaciniID, st.Kolicina, st.Cena, st.Cena, st.CenaVal,
		st.Rabat, st.RproID1, st.Vrd, st.Dokum, st.Dadok, s.UserName)
	_, err = r.exec(ctx, tx, "nova stavka porudžbenice", b)
	return err
}

// KomPorStavkaIzmeni changes the količina, the cena (cena and ncena in RSD, cenaval in the valuta) and the rabat of a
// stavka of a porudžbenica of the period (the legacy exit of a changed row of TABLE_PR_QRY_EPRO1), with the audit
// columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomPorStavkaIzmeni(ctx context.Context, tx db.Transaction, rdokID, rproID int64, kolicina, cena, cenaVal,
	rabat float64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("trazkolic = ?", kolicina).
		Set("cena = ?", cena).
		Set("ncena = ?", cena).
		Set("cenaval = ?", cenaVal).
		Set("rab = ?", rabat).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", rproID).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavka porudžbenice", b)
}

// KomPorVratiSve deletes all the stavke of a porudžbenica of the period (the legacy BTN_VratiSve): the stavke of the
// zahtev za ponudu they were taken from are slobodne again.
func (r *KomercijalnoRepository) KomPorVratiSve(ctx context.Context, tx db.Transaction, rdokID int64) error {
	oslobodi := querybuilder.NewUpdate("epro").
		Set("ststatus = ?", komStavkaSlobodna).
		Where("rproid in (select p.rproid1 from epro p where p.rdokid = ? and p.god = epro.god and p.kar = epro.kar)", rdokID)
	if err := inPeriod(ctx, oslobodi, ""); err != nil {
		return err
	}
	if _, err := r.exec(ctx, tx, "oslobađanje stavki zahteva za ponudu", oslobodi); err != nil {
		return err
	}
	b := querybuilder.New(`delete from epro`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "stavke porudžbenice", b)
	return err
}

// KomPorIzvori returns the slobodne stavke of a zahtev za ponudu of the period (not rezervisane and not taken by a
// stavka, the legacy "Količine po ponudi" of PuniDETALJ) in the order of their redni broj, with the naziv and the JM of
// their artikal.
func (r *KomercijalnoRepository) KomPorIzvori(ctx context.Context, ponudaID int64) ([]domain.KomPorIzvorRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rbr, 0)::int as rbr, coalesce(p.sifra, 0)::int as sifra,
			coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm, coalesce(p.trazkolic, 0)::float8 as kolicina
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rdokid = ?", ponudaID).
		Where("coalesce(p.ststatus, '') <> ?", komStavkaRezervisana).
		Where("not exists (select 1 from epro x where x.rproid1 = p.rproid and x.god = p.god and x.kar = p.kar)")
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []komPorIzvorDto
	if err := r.selectRows(ctx, "stavke zahteva za ponudu za porudžbenicu", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPorIzvorRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPorIzvorRow{RproID: row.RproID, Rbr: row.Rbr, Sifra: row.Sifra, Naziv: row.Naziv, JM: row.JM, Kolicina: row.Kolicina}
	}
	return out, nil
}

// komPorIzvorDto is a row of KomPorIzvori.
type komPorIzvorDto struct {
	RproID   int64   `db:"rproid"`
	Rbr      int     `db:"rbr"`
	Sifra    int     `db:"sifra"`
	Naziv    string  `db:"naziv"`
	JM       string  `db:"jm"`
	Kolicina float64 `db:"kolicina"`
}

// KomPorIzvor returns a slobodna stavka of the zahtev za ponudu of the period a porudžbenica takes (the legacy
// AddEPRONAR): not rezervisana and not taken by a stavka, with the poreska stopa of its artikal and its cena in the
// valuta; found is false otherwise.
func (r *KomercijalnoRepository) KomPorIzvor(ctx context.Context, tx db.Transaction, ponudaID, rproID int64) (domain.KomPorIzvor, bool, error) {
	b := querybuilder.New(`select p.rproid::bigint, coalesce(p.sifra, 0)::int, coalesce(s.po, 0)::int, coalesce(p.rsifid, 0)::bigint,
			coalesce(p.trazkolic, 0)::float8, coalesce(p.cenaval, 0)::float8
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rproid = ?", rproID).
		Where("p.rdokid = ?", ponudaID).
		Where("coalesce(p.ststatus, '') <> ?", komStavkaRezervisana).
		Where("not exists (select 1 from epro x where x.rproid1 = p.rproid and x.god = p.god and x.kar = p.kar)")
	if err := inPeriod(ctx, b, "p."); err != nil {
		return domain.KomPorIzvor{}, false, err
	}
	var st domain.KomPorIzvor
	err := r.row(ctx, tx, "stavka zahteva za ponudu", b, &st.RproID, &st.Sifra, &st.Po, &st.RsifID, &st.Kolicina, &st.CenaVal)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPorIzvor{}, false, nil
	}
	if err != nil {
		return domain.KomPorIzvor{}, false, err
	}
	return st, true, nil
}

// KomPorPrikaz returns the porudžbenice of the tab "Pregled porudžbenice" (the legacy PR_QRY_EDOK): the ones of the
// vrsta and the status in the magacini of the user of the session (maguser), of one magacin or of all of them
// (SviMagacini), the last ones first, with the "mag-opis" of the magacin, the naziv of the dobavljač and the
// "šifra-ime" of the komercijalista.
func (r *KomercijalnoRepository) KomPorPrikaz(ctx context.Context, p domain.KomPonudaPrikazParams) ([]domain.KomPorPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid,
			coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin, coalesce(e.vrd, 0)::int as vrd,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, coalesce(e.fkto, '') as fkto, coalesce(e.fana, '') as fana,
			coalesce(f.naziv, '') as dobavljac, coalesce(e.rok, 0)::int as rok, coalesce(e.otp, '') as otp,
			coalesce(k.sifkom::text || '-' || coalesce(k.imeprezime, ''), '') as referent
		from edok e`).
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana").
		Where("e.vrd = ?", p.Vrd).
		Where("coalesce(e.doksts, '') = ?", p.Doksts)
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.foot", "f.naziv", "e.fana", "k.imeprezime")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok desc nulls last, e.vrd desc, e.dokum desc")
	var rows []komPorPrikazDto
	if err := r.selectRows(ctx, "pregled porudžbenica", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPorPrikazRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPorPrikazRow{RdokID: row.RdokID, Magacin: row.Magacin, Vrd: row.Vrd, Dokum: row.Dokum, Dadok: row.Dadok, Fkto: row.Fkto,
			Fana: row.Fana, Dobavljac: row.Dobavljac, Rok: row.Rok, Otp: row.Otp, Referent: row.Referent}
	}
	return out, nil
}

// komPorPrikazDto is a row of KomPorPrikaz.
type komPorPrikazDto struct {
	RdokID    int64        `db:"rdokid"`
	Magacin   string       `db:"magacin"`
	Vrd       int          `db:"vrd"`
	Dokum     int          `db:"dokum"`
	Dadok     sql.NullTime `db:"dadok"`
	Fkto      string       `db:"fkto"`
	Fana      string       `db:"fana"`
	Dobavljac string       `db:"dobavljac"`
	Rok       int          `db:"rok"`
	Otp       string       `db:"otp"`
	Referent  string       `db:"referent"`
}

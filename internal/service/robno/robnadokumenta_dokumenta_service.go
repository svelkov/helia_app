package robno

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
)

//
// Štampa fakture: the report RobnoStampaFaktura (the legacy ROB_RPT_STAMPA_FAKTURA)
//

// robnoStampaFakturaUmanjenjeClan is the član of the Zakon o PDV the obaveštenje o umanjenju odbitka
// prethodnog PDV-a refers to.
const robnoStampaFakturaUmanjenjeClan = "21. st. 3"

// robnoStampaFakturaGrupaAvans is the group of the vrste dokumenta of the avansni računi
// (dokvrsta.grpdok): their print shows the "Ukupno sa PDV-om" of the avans (ITERATION_AVANSZBIR).
const robnoStampaFakturaGrupaAvans = "ARA"

// robnoStampaFakturaVrdAvans is the vrsta dokumenta of the avansni računi the fakture close (the
// brojevi of rdok.profak are dokumenti of this vrsta, like the legacy report reads them).
const robnoStampaFakturaVrdAvans = 172

// robnoStampaFakturaPibTekRac is the PIB of the firm (Lav Kompjuteri) whose fakture always print the
// tekući račun of the firm (fvr.tekrac) instead of the računi of the document or of its banke.
const robnoStampaFakturaPibTekRac = "100111568"

// robnoStampaFakturaVrsta is the print of the fakture: the faktura (veleprodaja, the legacy
// ROB_RPT_STAMPA_FAKTURA), the maloprodajni račun (ROB_RPT_STAMPA_FAKTURA_MP), the faktura usluga
// (ROB_RPT_STAMPA_FAKTURA_USLUGE2) or the avansni račun (ROB_RPT_STAMPA_ARA). All of them print the
// same source query (ROB_QRY_FAKTURESTAMPA).
type robnoStampaFakturaVrsta int

const (
	robnoStampaFakturaVeleprodaja robnoStampaFakturaVrsta = iota
	robnoStampaFakturaMaloprodaja
	robnoStampaFakturaUsluge
	robnoStampaFakturaAvansni
)

// robnoStampaFakturaUslugeAvansi turns on the avansni računi on the faktura usluga. The legacy
// ROB_RPT_STAMPA_FAKTURA_USLUGE2 has its procedure Avans commented out, so the faktura usluga prints no
// avansi; its translation is the one of the faktura (stampaFakturaAvansi and robnoStampaFakturaAvansi:
// the stavke of the avansni računi, vrsta 172, whose brojevi are in rdok.profak, with the PDV of their
// poreska stopa, summed per stopa), ready to be turned on here.
const robnoStampaFakturaUslugeAvansi = false

// robnoStampaFakturaPibUgovornaStrana is the PIB of the firm (Eparhija Karlovci) whose fakture usluga
// print the mesto isporuke as the "Ugovorna strana".
const robnoStampaFakturaPibUgovornaStrana = "102689206"

// robnoStampaFakturaGrupaMaloprodaja is the group of the vrste dokumenta of the maloprodajni računi
// (dokvrsta.grpdok), the documents the legacy ROB_RPT_STAMPA_FAKTURA_MP prints.
const robnoStampaFakturaGrupaMaloprodaja = "DIR"

// robnoStampaFakturaMagacin is the magacin of the print: its mesto is the mesto izdavanja računa and
// its adresa is added to the adresa of the firm; the maloprodajni račun prints its opis, telefon and
// e-mail (Skladište).
type robnoStampaFakturaMagacin struct {
	found  bool
	mesto  string
	adresa string
	opis   string
	tel    string
	email  string
}

// GetStampaFaktura returns the fakture of the selection, ready to print: one view per robni dokument,
// in the order of the source query of the legacy report (rdokid, rbr of the stavke), and the
// izdavalac (fvr) of the print. Like the legacy report, the group of the vrste dokumenta of the print
// is the group of the vrsta "od" (OdVrd) when the selection does not give the groups.
func (s *RobnoDokumentaResource) GetStampaFaktura(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaVeleprodaja)
}

// GetStampaFakturaMP returns the maloprodajni računi of the selection, ready to print (the legacy
// ROB_RPT_STAMPA_FAKTURA_MP, the same source query as the faktura with the group DIR): the stavke at the
// maloprodajna cena with the PDV (rpro.mcenap) and the PDV per stopa calculated out of the iznos (see
// robnoStampaFakturaView). Without a selected document and without groups of the selection the
// documents of the group DIR are printed, like the legacy report.
func (s *RobnoDokumentaResource) GetStampaFakturaMP(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	if params.GrupeDokumenata == "" && params.RdokID == 0 {
		params.GrupeDokumenata = robnoStampaFakturaGrupaMaloprodaja
	}
	return s.stampaFakture(ctx, params, robnoStampaFakturaMaloprodaja)
}

// GetStampaFakturaUsluge returns the fakture usluga of the selection, ready to print (the legacy
// ROB_RPT_STAMPA_FAKTURA_USLUGE2, the same source query as the faktura with the group of the vrsta
// dokumenta of the selection): the stavke and the totals like the faktura, without the avansni računi
// (see robnoStampaFakturaUslugeAvansi).
func (s *RobnoDokumentaResource) GetStampaFakturaUsluge(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaUsluge)
}

// GetStampaFakturaAvansni returns the avansni računi of the selection, ready to print (the legacy
// ROB_RPT_STAMPA_ARA, the same source query as the faktura with the group of the vrsta dokumenta of the
// selection, ARA): the stavke and the totals like the faktura (the iznos of the avans is the osnovica
// and the PDV of its stopa is added), with the datum of the avansna uplata and the vezni dokument.
func (s *RobnoDokumentaResource) GetStampaFakturaAvansni(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaAvansni)
}

// stampaFakture returns the fakture of the selection in the print vrsta (the faktura, the maloprodajni
// račun or the faktura usluga), ready to print, and the izdavalac of the print (see GetStampaFaktura,
// GetStampaFakturaMP and GetStampaFakturaUsluge).
func (s *RobnoDokumentaResource) stampaFakture(ctx context.Context, params domain.RobnoStampaFakturaParams, vrsta robnoStampaFakturaVrsta) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	magacin, err := s.stampaFakturaMagacin(ctx, userSession, params.MagaciniID)
	if err != nil {
		return nil, firma, err
	}
	// The selected document is printed whatever its group; otherwise the group is the one of the
	// vrsta dokumenta of the selection.
	if params.GrupeDokumenata == "" && params.RdokID == 0 {
		if params.GrupeDokumenata, err = s.stampaFakturaGrupa(ctx, userSession, firstNonEmpty(params.OdVrd, params.Vrd)); err != nil {
			return nil, firma, err
		}
	}
	rows, err := s.stampaFakturaRows(ctx, userSession, params)
	if err != nil {
		return nil, firma, err
	}
	if len(rows) == 0 {
		return []domain.RobnoStampaFakturaView{}, firma, nil
	}

	// The documents of the selection in the order of the query, with their stavke, and the brojevi of
	// the avansni računi they close.
	var rdokIDs []any
	stavkeOf := map[int64][]domain.RobnoStampaFakturaRowDto{}
	avansniBrojevi := map[int64]bool{}
	for _, row := range rows {
		if _, found := stavkeOf[row.RdokID]; !found {
			rdokIDs = append(rdokIDs, row.RdokID)
			for _, broj := range robnoStampaFakturaProfak(row.Profak) {
				avansniBrojevi[broj] = true
			}
		}
		stavkeOf[row.RdokID] = append(stavkeOf[row.RdokID], row)
	}
	avansi, err := s.stampaFakturaAvansi(ctx, userSession, avansniBrojevi)
	if err != nil {
		return nil, firma, err
	}
	ambalazaOf, err := s.stampaFakturaAmbalaza(ctx, userSession, rdokIDs)
	if err != nil {
		return nil, firma, err
	}
	rateOf, err := s.stampaFakturaRate(ctx, userSession, rdokIDs)
	if err != nil {
		return nil, firma, err
	}

	fakture := make([]domain.RobnoStampaFakturaView, 0, len(rdokIDs))
	for _, id := range rdokIDs {
		rdokID := id.(int64)
		stavke := stavkeOf[rdokID]
		// The stavke of the avansni računi of the document (its PROFAK).
		var avansiDok []domain.RobnoStampaFakturaAvansDto
		brojevi := map[int64]bool{}
		for _, broj := range robnoStampaFakturaProfak(stavke[0].Profak) {
			brojevi[broj] = true
		}
		for _, av := range avansi {
			if vrsta == robnoStampaFakturaUsluge && !robnoStampaFakturaUslugeAvansi {
				break
			}
			if brojevi[av.Dokum] {
				avansiDok = append(avansiDok, av)
			}
		}
		fakture = append(fakture, robnoStampaFakturaView(firma, magacin, stavke, avansiDok, ambalazaOf[rdokID], rateOf[rdokID], vrsta))
	}
	return fakture, firma, nil
}

// stampaFakturaRows runs the source query of the legacy report (ROB_QRY_FAKTURESTAMPA): the stavke
// (rpro) of the robni dokumenti (rdok) of the vrsta naloga and of the groups of the vrste dokumenta of
// the selection, without the ambalaža (the artikli with a šifra 99xxxx), with the data of the header
// of the print. The price of a stavka is its fakturna (VP) cena, rpro.fcena (rpro.cena is the
// nabavna cena and rpro.iznos the nabavna value), and its iznos the količina times that cena.
func (s *RobnoDokumentaResource) stampaFakturaRows(ctx context.Context, userSession *domain.UserSession, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaRowDto, error) {
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			rdok.datiz,
			coalesce(rdok.rok, 0) as rok,
			coalesce(rdok.pla, '') as pla,
			coalesce(rpro.mcenap, 0) as mcenap,
			coalesce(misp.naziv, '') as mispnaziv,
			coalesce(misp.adresa, '') as mispadresa,
			coalesce(misp.pobro, 0) as misppobro,
			coalesce(misp.mesto, '') as mispmesto,
			coalesce(nullif(misp.gln, 0)::text, '') as mispgln,
			coalesce(rdok.narudzb, '') as narudzb,
			coalesce(rdok.vozac, '') as vozac,
			coalesce(rdok.brvozila, '') as brvozila,
			coalesce(rdok.foot, '') as foot,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(rdok.opis, '') as opis,
			coalesce(rdok.zirorac, '') as zirorac,
			coalesce(rdok.profak, '') as profak,
			coalesce(rdok.posuslnab, '') as posuslnab,
			coalesce(rdok.bttowght, 0) as bttowght,
			coalesce(rdok.netwght, 0) as netwght,
			coalesce(rdok.totboxs, 0) as totboxs,
			coalesce(rdok.brpaleta, 0) as brpaleta,
			coalesce(dokvrsta.opis, '') as doknaslov,
			coalesce(dokvrsta.grpdok, '') as grpdok,
			coalesce(rdok.fkto, '') as kupackonto,
			coalesce(rdok.fana, '') as kupacsifra,
			coalesce(p.naziv, fkpl.naziv, '') as kupacnaziv,
			coalesce(p.adresa, '') as kupacadresa,
			trim(coalesce(nullif(p.pobro, 0)::text, '') || ' ' || coalesce(p.mesto, '')) as kupacmesto,
			coalesce(p.pak, 0)::bigint::text as kupacpak,
			coalesce(p.pib, '') as kupacpib,
			coalesce(p.matbr, '') as kupacmbr,
			coalesce(p.telefon, '') as kupactelefon,
			coalesce(rdok.kom, 0)::text as komsifra,
			coalesce(kom.imeprezime, '') as komnaziv,
			trim(coalesce(banke.banka, '') || ' ' || coalesce(banke.brrac, '')) as banka,
			coalesce(concat_ws(' ', porkat.article,
				'st. ' || nullif(porkat.paragraph, ''),
				'tač. ' || nullif(porkat.point, '')), '') as oslobodjenjeclan,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(nullif(rpro.naz1, ''), rsif.naziv, '') as naz1,
			coalesce(nullif(rpro.jm, ''), rsif.jm, '') as jm,
			coalesce(rpro.po, 0) as po,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(nullif(rpro.fcena, 0), rpro.cena, 0) as cena,
			coalesce(rpro.rab, 0) as rab,
			round(coalesce(rpro.kolic, 0) * coalesce(nullif(rpro.fcena, 0), rpro.cena, 0), 2) as iznos,
			coalesce(ps.pp, rpro.pdvpct, 0) as stopa,
			coalesce(rpro.dani, 0) as dani
		from rpro`, true)
	qb.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin("inner join dokvrsta on dokvrsta.god = rpro.god and dokvrsta.kar = rpro.kar and dokvrsta.vrd = rpro.vrd")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	// The naziv and the jedinica mere of the artikal when the stavka has none of its own (rpro.naz1 is
	// filled only for the stavke whose naziv was changed on the document).
	qb.AddJoin("left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin("left join komercijalisti kom on kom.komid = rdok.komid")
	// The mesto isporuke of the document: the fisp of the kupac (konto and šifra) and of rdok.mi, like
	// the legacy HReadSeekFirst(FISP, GODKARKONTOSIFRASIFRAMI, ...).
	qb.AddJoin(`left join lateral (select f.naziv, f.adresa, f.pobro, f.mesto, f.gln from fisp f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
			and f.mi = coalesce(rdok.mi, 0)
		order by f.fispid limit 1) misp on true`)
	qb.AddJoin("left join banke on banke.god = rdok.god and banke.kar = rdok.kar and banke.sifra = rdok.sifbank::text and coalesce(rdok.sifbank, 0) <> 0")
	qb.AddJoin("left join porkat on porkat.keykat = rdok.keykat and coalesce(rdok.keykat, '') <> ''")
	qb.AddJoin(robnoStampaFakturaStopaJoin)
	robnoStampaDokumentaUslovi(qb, userSession, params)
	qb.AddIn("dokvrsta.grpdok", grupeDokumenata(params.GrupeDokumenata))
	// The legacy NOT WL.NumToString(RPRO.SIFRA,'06d') LIKE '99____': the ambalaža (990000-999999) is
	// printed in its own block.
	qb.AddCustomCondition("not " + robnoStampaFakturaAmbalazaCondition)
	qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	sqlQuery, args := qb.Build()
	entities, err := s.stampaFakturaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	return *entities, nil
}

// robnoStampaFakturaStopaJoin joins the poreska stopa (rpor.pp) of the poreska oznaka of the stavka
// (rpro.po) in force on the date of the stavka, like the legacy VratiPorez2 (rpro.pdvpct is never
// filled; it is read only when the oznaka has no stopa).
const robnoStampaFakturaStopaJoin = `left join lateral (select r.pp from rpor r
	where r.po = rpro.po and r.datum <= rpro.dadok
	order by r.datum desc limit 1) ps on true`

// robnoStampaFakturaAmbalazaCondition selects the stavke of the ambalaža: the artikli with a 6 digit
// šifra 99xxxx (the legacy WL.NumToString(RPRO.SIFRA,'06d') LIKE '99____').
const robnoStampaFakturaAmbalazaCondition = "coalesce(rpro.sifra, 0) between 990000 and 999999"

// stampaFakturaAmbalaza returns the evidentna ambalaža of the fakture of the print (the stavke with a
// šifra 99xxxx, the legacy query of the ambalaža), keyed by the rdokid of the faktura and ordered by
// the šifra.
func (s *RobnoDokumentaResource) stampaFakturaAmbalaza(ctx context.Context, userSession *domain.UserSession, rdokIDs []any) (map[int64][]domain.RobnoStampaFakturaRowDto, error) {
	qb := common.NewQueryBuilder(`
		select
			rpro.rdokid,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(nullif(rpro.naz1, ''), rsif.naziv, '') as naz1,
			coalesce(nullif(rpro.jm, ''), rsif.jm, '') as jm,
			coalesce(rpro.po, 0) as po,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.cena, 0) as cena,
			coalesce(rpro.rab, 0) as rab,
			coalesce(rpro.iznos, 0) as iznos,
			coalesce(ps.pp, rpro.pdvpct, 0) as stopa
		from rpro`, true)
	qb.AddJoin("left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(robnoStampaFakturaStopaJoin)
	qb.AddEqual("rpro.god", userSession.SelectedGod)
	qb.AddEqual("rpro.kar", userSession.SelectedKar)
	qb.AddIn("rpro.rdokid", rdokIDs)
	qb.AddCustomCondition(robnoStampaFakturaAmbalazaCondition)
	qb.AddOrderBy("rpro.rdokid, rpro.sifra")
	sqlQuery, args := qb.Build()
	entities, err := s.stampaFakturaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	result := map[int64][]domain.RobnoStampaFakturaRowDto{}
	for _, entity := range *entities {
		result[entity.RdokID] = append(result[entity.RdokID], entity)
	}
	return result, nil
}

// stampaFakturaAvansi returns the stavke of the avansni računi (vrd 172) with the given brojevi, like
// the legacy procedure Avans reads them (the brojevi are the PROFAK of the fakture of the print).
func (s *RobnoDokumentaResource) stampaFakturaAvansi(ctx context.Context, userSession *domain.UserSession, brojevi map[int64]bool) ([]domain.RobnoStampaFakturaAvansDto, error) {
	if len(brojevi) == 0 {
		return nil, nil
	}
	var dokumenti []any
	for broj := range brojevi {
		dokumenti = append(dokumenti, broj)
	}
	qb := common.NewQueryBuilder(`
		select
			coalesce(rpro.dokum, 0) as dokum,
			rpro.dadok,
			coalesce(rpro.iznos, 0) as iznos,
			coalesce(ps.pp, rpro.pdvpct, 0) as stopa
		from rpro`, true)
	qb.AddJoin(robnoStampaFakturaStopaJoin)
	qb.AddEqual("rpro.god", userSession.SelectedGod)
	qb.AddEqual("rpro.kar", userSession.SelectedKar)
	qb.AddEqual("rpro.vrd", robnoStampaFakturaVrdAvans)
	qb.AddIn("rpro.dokum", dokumenti)
	qb.AddOrderBy("rpro.dokum, rpro.rbr")
	sqlQuery, args := qb.Build()
	entities, err := s.stampaFakturaAvansRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	return *entities, nil
}

// stampaFakturaRate returns the rate of the fakture of the print (faktrate), keyed by the rdokid of
// the faktura.
func (s *RobnoDokumentaResource) stampaFakturaRate(ctx context.Context, userSession *domain.UserSession, rdokIDs []any) (map[int64][]domain.RobnoStampaFakturaRataDto, error) {
	qb := common.NewQueryBuilder(`
		select
			fr.rdokid,
			coalesce(fr.rbrrate, 0) as rbrrate,
			fr.datrate,
			coalesce(fr.procenat, 0) as procenat,
			coalesce(fr.osnovica, 0) as osnovica,
			coalesce(fr.pdv, 0) as pdv
		from faktrate fr`, true)
	qb.AddEqual("fr.god", userSession.SelectedGod)
	qb.AddEqual("fr.kar", userSession.SelectedKar)
	qb.AddIn("fr.rdokid", rdokIDs)
	qb.AddOrderBy("fr.rdokid, fr.rbrrate")
	sqlQuery, args := qb.Build()
	entities, err := s.stampaFakturaRateRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	result := map[int64][]domain.RobnoStampaFakturaRataDto{}
	for _, entity := range *entities {
		result[entity.RdokID] = append(result[entity.RdokID], entity)
	}
	return result, nil
}

// stampaFirma returns the izdavalac of the printed robni dokumenti (common to the prints of every
// vrsta dokumenta): the firm (fvr) of the session with its tekući računi (the banke not flagged
// nafakne) and its logo.
func (s *RobnoDokumentaResource) stampaFirma(ctx context.Context, userSession *domain.UserSession) (domain.RobnoStampaFakturaFirmaDto, error) {
	qb := common.NewQueryBuilder(`
		select
			coalesce(fvr.naziv, '') as naziv,
			coalesce(fvr.adresa, '') as adresa,
			coalesce(nullif(fvr.pobro, 0)::text, '') as pobro,
			coalesce(fvr.mesto, '') as mesto,
			coalesce(fvr.pib, '') as pib,
			coalesce(fvr.matbr, '') as matbr,
			coalesce(fvr.sifdel, '') as sifdel,
			coalesce(fvr.tel, '') as tel,
			coalesce(fvr.email, '') as email,
			coalesce(fvr.apr, '') as apr,
			coalesce(fvr.bpg, '') as bpg,
			coalesce(fvr.tekrac, '') as tekrac,
			coalesce(fvr.obv, false) as obv,
			coalesce(fvr.brobvpdv, '') as brobvpdv,
			coalesce((select string_agg(trim(b.brrac) || ' - ' || trim(b.banka), ';  ' order by b.idbanke)
				from banke b
				where b.god = fvr.god and b.kar = fvr.kar and not coalesce(b.nafakne, false)), '') as banke,
			fvr.logofirme as logo
		from fvr`, true)
	qb.AddEqual("fvr.god", userSession.SelectedGod)
	qb.AddEqual("fvr.kar", userSession.SelectedKar)
	qb.AddEqual("fvr.naziv", userSession.Firma)
	sqlQuery, args := qb.Build()
	entities, err := s.stampaFakturaFirmaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.RobnoStampaFakturaFirmaDto{}, err
	}
	if len(*entities) == 0 {
		return domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("firma %q not found", userSession.Firma)
	}
	return (*entities)[0], nil
}

// stampaFakturaMagacin returns the magacin of the print (not found when the selection has none).
func (s *RobnoDokumentaResource) stampaFakturaMagacin(ctx context.Context, userSession *domain.UserSession, magaciniID int) (robnoStampaFakturaMagacin, error) {
	if magaciniID == 0 {
		return robnoStampaFakturaMagacin{}, nil
	}
	qb := common.NewQueryBuilder(`select magaciniid, mag, coalesce(mesto, '') as mesto, coalesce(adresa, '') as adresa,
		coalesce(opis, '') as opis, coalesce(tel, '') as tel, coalesce(email, '') as email from magacini`, true)
	qb.AddEqual("god", userSession.SelectedGod)
	qb.AddEqual("kar", userSession.SelectedKar)
	qb.AddEqual("magaciniid", magaciniID)
	sqlQuery, args := qb.Build()
	entities, err := s.magRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return robnoStampaFakturaMagacin{}, err
	}
	if len(*entities) == 0 {
		return robnoStampaFakturaMagacin{}, nil
	}
	mag := (*entities)[0]
	return robnoStampaFakturaMagacin{found: true, mesto: mag.Mesto, adresa: mag.Adresa, opis: mag.Opis, tel: mag.Tel, email: mag.Email}, nil
}

// stampaFakturaGrupa returns the group (dokvrsta.grpdok) of the vrsta dokumenta "od" of the print,
// like the legacy report reads it ("" when the vrsta is not given or does not exist).
func (s *RobnoDokumentaResource) stampaFakturaGrupa(ctx context.Context, userSession *domain.UserSession, odVrd string) (string, error) {
	vrd, err := strconv.Atoi(strings.TrimSpace(odVrd))
	if err != nil {
		return "", nil
	}
	qb := common.NewQueryBuilder("select coalesce(grpdok, '') as grpdok from dokvrsta", true)
	qb.AddEqual("god", userSession.SelectedGod)
	qb.AddEqual("kar", userSession.SelectedKar)
	qb.AddEqual("vrd", vrd)
	sqlQuery, args := qb.Build()
	entities, err := s.dokvrstaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return "", err
	}
	if len(*entities) == 0 {
		return "", nil
	}
	return (*entities)[0].GrpDok, nil
}

// RobnoStampaLogo returns the logo of the firm as a data URI for the <img> of the print of a robni
// dokument ("" when the firm has no logo).
func RobnoStampaLogo(logo []byte) string {
	if len(logo) == 0 {
		return ""
	}
	return "data:" + http.DetectContentType(logo) + ";base64," + base64.StdEncoding.EncodeToString(logo)
}

// robnoStampaFakturaView builds the printed faktura of one robni dokument from its stavke (the rows of
// the source query, all of the same document), the stavke of its avansni računi, its ambalaža and its
// rate, in the print vrsta (the maloprodajni račun, the legacy ROB_RPT_STAMPA_FAKTURA_MP, has its own
// cene and PDV, see below).
func robnoStampaFakturaView(firma domain.RobnoStampaFakturaFirmaDto, magacin robnoStampaFakturaMagacin, stavke []domain.RobnoStampaFakturaRowDto, avansi []domain.RobnoStampaFakturaAvansDto, ambalaza []domain.RobnoStampaFakturaRowDto, rate []domain.RobnoStampaFakturaRataDto, vrsta robnoStampaFakturaVrsta) domain.RobnoStampaFakturaView {
	maloprodaja := vrsta == robnoStampaFakturaMaloprodaja
	dok := stavke[0]
	fak := domain.RobnoStampaFakturaView{
		KupacKonto:   dok.KupacKonto,
		KupacSifra:   dok.KupacSifra,
		KupacNaziv:   dok.KupacNaziv,
		KupacAdresa:  dok.KupacAdresa,
		KupacMesto:   dok.KupacMesto,
		KupacPak:     dok.KupacPak,
		KupacPib:     dok.KupacPib,
		KupacMbr:     dok.KupacMbr,
		KupacTelefon: dok.KupacTelefon,

		Otpremnica:         dok.Dokiz,
		Narudzbenica:       dok.Narudzb,
		UsloviPlacanja:     dok.Pla,
		Komercijalista:     dok.KomercijalistaSif,
		KomercijalistaOpis: dok.KomercijalistaNaz,
		Vozac:              dok.Vozac,
		RegBrojVozila:      dok.Brvozila,

		// "Račun: 130-1156": the vrsta and the broj of the document, like the legacy report.
		Naslov:         i18n.GetInstance().Label("Račun naslov"),
		BrojDokumenta:  fmt.Sprintf("%d-%d", dok.Vrd, dok.Dokum),
		DatumPrometa:   common.FormatNullTime(dok.Dadok, common.DateLayout),
		Valuta:         common.AddDaysToNullTime(dok.Dadok, int(dok.Rok), common.DateLayout),
		DatumDokumenta: common.FormatNullTime(dok.Dadok, common.DateLayout),
		Nalog:          fmt.Sprintf("%s-%d", dok.Tipdok, dok.Nalog),

		OslobodjenjeClan: dok.OslobodjenjeClan,
		PoreskaNapomena:  dok.Pornapomena,

		Napomena:    dok.Foot,
		BrutoTezina: robnoStampaFakturaNonZero(dok.Bttowght, 2),
		NetoTezina:  robnoStampaFakturaNonZero(dok.Netwght, 2),
		BrojKutija:  robnoStampaFakturaNonZero(float64(dok.Totboxs), 0),
		BrojPaleta:  robnoStampaFakturaNonZero(float64(dok.Brpaleta), 0),
	}
	robnoStampaFakturaIzdavalac(&fak, firma, magacin, dok)
	robnoStampaFakturaMestoIsporuke(&fak, firma, dok, vrsta)
	if maloprodaja {
		// "MP-Račun br.: 196-12" and the Skladište: the magacin of the print.
		fak.Naslov = i18n.GetInstance().Label("MP-Račun br.")
		fak.MagacinOpis = magacin.opis
		fak.MagacinTel = magacin.tel
		fak.MagacinEmail = magacin.email
	}
	if vrsta == robnoStampaFakturaAvansni {
		// "AVANSNI RAČUN Br.: 172-8", the datum of the avansna uplata and the vezni dokument (its opis
		// only when it is not the broj of the vezni dokument again).
		fak.Naslov = i18n.GetInstance().Label("Avansni račun Br.")
		fak.DatumAvansneUplate = common.FormatNullTime(dok.Datiz, common.DateLayout)
		fak.VezniDokument = dok.Dokiz
		if dok.Opis != dok.Dokiz {
			fak.VezniDokumentOpis = dok.Opis
		}
	}
	if dok.Rok != 0 {
		fak.RokPlacanja = fmt.Sprintf("%d", dok.Rok)
	}

	// The stavke and the totals per poreska stopa, like the legacy report: the iznos of a stavka is
	// the količina times the fakturna cena (before the rabat), its rabat is rounded to the para, its
	// osnovica is the iznos less the rabat and its PDV the osnovica times the stopa. A stavka with its
	// own rok (rpro.dani) falls due on the date of the document plus those days.
	//
	// The maloprodajni račun prints the stavke at the maloprodajna cena with the PDV (rpro.mcenap; the
	// fakturna cena with the PDV of its stopa when the stavka has none): the iznos of a stavka is the
	// količina times that cena (before the rabat), the rabat is rounded to the para, the iznos less the
	// rabat holds the PDV, its osnovica is that iznos without the PDV of the stopa and the PDV the rest.
	// SVEGA and RABAT are with the PDV and ZA NAPLATU is SVEGA less RABAT.
	type stopaSums struct{ osnovica, pdv float64 }
	perStopa := map[float64]*stopaSums{}
	var stope []float64
	perDanu := map[time.Time]float64{}
	imaDane := false
	var ukIznos, ukRabat, ukPdv, ukOslobodjeno float64
	for _, st := range stavke {
		cena, iznos := st.Cena, st.Iznos
		if maloprodaja {
			cena = st.Mcenap
			if cena == 0 {
				cena = robnoStampaFakturaRound(st.Cena * (100 + st.Stopa) / 100)
			}
			iznos = robnoStampaFakturaRound(st.Kolic * cena)
		}
		rabat := robnoStampaFakturaRound(iznos * st.Rab / 100)
		osnovica := iznos - rabat
		pdv := robnoStampaFakturaRound(osnovica * st.Stopa / 100)
		if maloprodaja {
			saPdv := iznos - rabat
			osnovica = robnoStampaFakturaRound(saPdv * 100 / (100 + st.Stopa))
			pdv = robnoStampaFakturaRound(saPdv - osnovica)
		}
		fak.Stavke = append(fak.Stavke, domain.RobnoStampaFakturaStavka{
			Rbr:        fmt.Sprintf("%d", st.Rbr),
			Sifra:      robnoStampaFakturaSifra(st.Sifra),
			Naziv:      st.Naz1,
			Jm:         st.Jm,
			Kolicina:   robnoStampaFakturaKolicina(st.Kolic),
			Cena:       common.FormatNumberWithSystemLocale(cena, 2),
			ProcRabata: common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			Rabat:      common.FormatNumberWithSystemLocale(rabat, 2),
			ProcPdv:    robnoStampaFakturaStopa(st.Stopa),
			Pdv:        common.FormatNumberWithSystemLocale(pdv, 2),
			Iznos:      common.FormatNumberWithSystemLocale(iznos, 2),
		})
		if _, found := perStopa[st.Stopa]; !found {
			perStopa[st.Stopa] = &stopaSums{}
			stope = append(stope, st.Stopa)
		}
		perStopa[st.Stopa].osnovica += osnovica
		perStopa[st.Stopa].pdv += pdv
		ukIznos += iznos
		ukRabat += rabat
		ukPdv += pdv
		if st.Stopa == 0 {
			ukOslobodjeno += osnovica
		}
		if dok.Dadok.Valid {
			perDanu[dok.Dadok.Time.AddDate(0, 0, int(st.Dani))] += osnovica + pdv
		}
		if st.Dani != 0 {
			imaDane = true
		}
	}
	// The lowest stopa first, like the legacy report ("PDV PO STOPI 10,00%" before 20,00%).
	sort.Float64s(stope)
	for _, stopa := range stope {
		fak.PdvPoStopama = append(fak.PdvPoStopama, domain.RobnoStampaFakturaPdv{
			Stopa:    common.FormatNumberWithSystemLocale(stopa, 2) + "%",
			Osnovica: common.FormatNumberWithSystemLocale(perStopa[stopa].osnovica, 2),
			Pdv:      common.FormatNumberWithSystemLocale(perStopa[stopa].pdv, 2),
			Iznos:    common.FormatNumberWithSystemLocale(perStopa[stopa].osnovica+perStopa[stopa].pdv, 2),
		})
	}
	// SVEGA is the iznos of the stavke (before the rabat, without the PDV), RABAT the rabat of the
	// stavke and ZA NAPLATU the osnovica (SVEGA less RABAT) with the PDV, like the legacy report.
	zaNaplatu := ukIznos - ukRabat + ukPdv
	if maloprodaja {
		zaNaplatu = ukIznos - ukRabat
	}
	fak.Svega = common.FormatNumberWithSystemLocale(ukIznos, 2)
	fak.Rabat = common.FormatNumberWithSystemLocale(ukRabat, 2)
	fak.ZaNaplatu = common.FormatNumberWithSystemLocale(zaNaplatu, 2)
	// Za naplatu po danima: only when some stavka has its own rok.
	if imaDane {
		var dani []time.Time
		for dan := range perDanu {
			dani = append(dani, dan)
		}
		sort.Slice(dani, func(i, j int) bool { return dani[i].Before(dani[j]) })
		for _, dan := range dani {
			fak.ZaNaplatuPoDanima = append(fak.ZaNaplatuPoDanima, domain.RobnoStampaFakturaZaNaplatuDan{
				Datum: dan.Format(common.DateLayout),
				Iznos: common.FormatNumberWithSystemLocale(perDanu[dan], 2),
			})
		}
	}
	if fak.OslobodjenjeClan != "" {
		fak.OslobodjenjeIznos = common.FormatNumberWithSystemLocale(ukOslobodjeno, 2)
	}
	if dok.Grpdok == robnoStampaFakturaGrupaAvans {
		fak.UkupnoSaPdv = fak.ZaNaplatu
	}
	// A document with a negative total (storno, smanjenje) reduces the odbitak prethodnog PDV-a of the
	// kupac: the print adds the obaveštenje o umanjenju.
	if zaNaplatu < 0 && ukPdv != 0 {
		fak.Umanjenje = true
		fak.UmanjenjeClan = robnoStampaFakturaUmanjenjeClan
	}

	robnoStampaFakturaAvansi(&fak, avansi, zaNaplatu)

	// The evidentna ambalaža: numbered in the order of the šifra; it is not part of the totals.
	for i, a := range ambalaza {
		fak.Ambalaza = append(fak.Ambalaza, domain.RobnoStampaFakturaAmbalaza{
			Rbr:      fmt.Sprintf("%d", i+1),
			Sifra:    robnoStampaFakturaSifra(a.Sifra),
			Naziv:    a.Naz1,
			Jm:       a.Jm,
			Kolicina: common.FormatNumberWithSystemLocale(a.Kolic, 2),
			Cena:     common.FormatNumberWithSystemLocale(a.Cena, 2),
			Iznos:    common.FormatNumberWithSystemLocale(a.Iznos, 2),
			Stopa:    robnoStampaFakturaStopa(a.Stopa),
		})
	}

	// The rate of the faktura and their totals.
	var rateOsnovica, ratePdv float64
	for _, r := range rate {
		fak.Rate = append(fak.Rate, domain.RobnoStampaFakturaRata{
			Rbr:      fmt.Sprintf("%d", r.Rbrrate),
			Datum:    common.FormatNullTime(r.Datrate, common.DateLayout),
			Procenat: robnoStampaFakturaStopa(r.Procenat),
			Osnovica: common.FormatNumberWithSystemLocale(r.Osnovica, 2),
			Pdv:      common.FormatNumberWithSystemLocale(r.Pdv, 2),
			UkSaPdv:  common.FormatNumberWithSystemLocale(r.Osnovica+r.Pdv, 2),
		})
		rateOsnovica += r.Osnovica
		ratePdv += r.Pdv
	}
	if len(rate) > 0 {
		fak.RateUkOsnovica = common.FormatNumberWithSystemLocale(rateOsnovica, 2)
		fak.RateUkPdv = common.FormatNumberWithSystemLocale(ratePdv, 2)
		fak.RateUkSaPdv = common.FormatNumberWithSystemLocale(rateOsnovica+ratePdv, 2)
	}
	return fak
}

// robnoStampaFakturaIzdavalac fills the izdavalac of the faktura the way the legacy report composes
// its items: every value carries its caption.
func robnoStampaFakturaIzdavalac(fak *domain.RobnoStampaFakturaView, firma domain.RobnoStampaFakturaFirmaDto, magacin robnoStampaFakturaMagacin, dok domain.RobnoStampaFakturaRowDto) {
	lbl := i18n.GetInstance().Label
	// ITEM_Adresa: poštanski broj, mesto, adresa of the firm and the adresa of the magacin.
	fak.FirmaAdresa = strings.TrimSpace(firma.Pobro+" "+firma.Mesto) + " , " + firma.Adresa
	if magacin.found && magacin.adresa != "" {
		fak.FirmaAdresa += " , " + magacin.adresa
	}
	// ITEM_MIR: the mesto of the magacin, otherwise the mesto of the firm.
	fak.MestoIzdavanja = firma.Mesto
	if magacin.found {
		fak.MestoIzdavanja = magacin.mesto
	}
	if firma.Obv {
		fak.FirmaObveznik = lbl("Obv. PDV-a br") + ": " + firma.Brobvpdv
	} else {
		fak.FirmaObveznik = lbl("Nije obveznik PDV-a")
	}
	if firma.Apr != "" {
		fak.FirmaApr = lbl("Rešenje APR") + " " + firma.Apr
	}
	if firma.Email != "" {
		fak.FirmaEmail = lbl("E-mail") + ": " + firma.Email
	}
	fak.FirmaPib = lbl("PIB") + " : " + firma.Pib
	fak.FirmaMbr = lbl("Matični broj") + " : " + firma.Matbr
	fak.FirmaSifdel = lbl("Šifra delatnosti") + " : " + firma.Sifdel
	fak.FirmaTel = lbl("Telefon/faks") + " : " + firma.Tel
	if firma.Bpg != "" {
		fak.FirmaBpg = lbl("BPG") + ": " + firma.Bpg
	}
	// ITEM_ZIRO: the tekući račun of the document (rdok.posuslnab); when it has none ("", "-1" or "-")
	// the tekući računi of the firm (its banke), and for one firm always fvr.tekrac.
	ziro := dok.Posuslnab
	switch {
	case firma.Pib == robnoStampaFakturaPibTekRac:
		ziro = firma.Tekrac
	case ziro == "" || ziro == "-1" || ziro == "-":
		ziro = firma.Banke
	}
	if ziro != "" {
		fak.FirmaZiro = lbl("Tekući račun") + " : " + ziro
	}
	fak.FirmaBanka = dok.Banka
}

// robnoStampaFakturaMestoIsporuke fills the mesto isporuke of the faktura (ITEM_MISP and ITEM_MISP1)
// when the document has one: "M. Isporuke: naziv" (on the faktura usluga of one firm "Ugovorna strana:
// naziv") and "adresa, poštanski broj mesto" with the GLN when it is set.
func robnoStampaFakturaMestoIsporuke(fak *domain.RobnoStampaFakturaView, firma domain.RobnoStampaFakturaFirmaDto, dok domain.RobnoStampaFakturaRowDto, vrsta robnoStampaFakturaVrsta) {
	if dok.MispNaziv == "" {
		return
	}
	lbl := i18n.GetInstance().Label
	caption := lbl("M. Isporuke")
	if vrsta == robnoStampaFakturaUsluge && firma.Pib == robnoStampaFakturaPibUgovornaStrana {
		caption = lbl("Ugovorna strana")
	}
	fak.MestoIsporuke = caption + ": " + dok.MispNaziv
	fak.MestoIsporuke2 = dok.MispAdresa + ", " + strings.TrimSpace(fmt.Sprintf("%d %s", dok.MispPobro, dok.MispMesto))
	if dok.MispGln != "" {
		fak.MestoIsporuke2 += " " + lbl("GLN") + ":" + dok.MispGln
	}
}

// robnoStampaFakturaAvansi fills the avansni računi closed with the faktura (the legacy procedure
// Avans): one row per avansni račun and poreska stopa, the avansni PDV per stopa, the total of the
// avansi with their PDV and the rest to pay.
func robnoStampaFakturaAvansi(fak *domain.RobnoStampaFakturaView, avansi []domain.RobnoStampaFakturaAvansDto, zaNaplatu float64) {
	if len(avansi) == 0 {
		return
	}
	type sums struct {
		broj     int64
		datum    string
		stopa    float64
		osnovica float64
		pdv      float64
	}
	var redovi []*sums
	poRacunu := map[string]*sums{}
	poStopi := map[float64]*sums{}
	var stope []float64
	var uAvans, uPor float64
	for _, av := range avansi {
		pdv := robnoStampaFakturaRound(av.Iznos * av.Stopa / 100)
		uAvans += av.Iznos
		uPor += pdv
		key := fmt.Sprintf("%d/%v", av.Dokum, av.Stopa)
		if poRacunu[key] == nil {
			poRacunu[key] = &sums{broj: av.Dokum, datum: common.FormatNullTime(av.Dadok, common.DateLayout), stopa: av.Stopa}
			redovi = append(redovi, poRacunu[key])
		}
		poRacunu[key].osnovica += av.Iznos
		poRacunu[key].pdv += pdv
		if poStopi[av.Stopa] == nil {
			poStopi[av.Stopa] = &sums{stopa: av.Stopa}
			stope = append(stope, av.Stopa)
		}
		poStopi[av.Stopa].osnovica += av.Iznos
		poStopi[av.Stopa].pdv += pdv
	}
	for _, r := range redovi {
		fak.Avansi = append(fak.Avansi, domain.RobnoStampaFakturaAvans{
			Broj:     fmt.Sprintf("%d", r.broj),
			Datum:    r.datum,
			Osnovica: common.FormatNumberWithSystemLocale(r.osnovica, 2),
			Stopa:    robnoStampaFakturaStopa(r.stopa),
			Pdv:      common.FormatNumberWithSystemLocale(r.pdv, 2),
		})
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(stope)))
	for _, stopa := range stope {
		fak.AvansniPdv = append(fak.AvansniPdv, domain.RobnoStampaFakturaPdv{
			Stopa:    robnoStampaFakturaStopa(stopa),
			Osnovica: common.FormatNumberWithSystemLocale(poStopi[stopa].osnovica, 2),
			Pdv:      common.FormatNumberWithSystemLocale(poStopi[stopa].pdv, 2),
		})
	}
	fak.UkupnoAvans = common.FormatNumberWithSystemLocale(uAvans+uPor, 2)
	fak.OstaloZaUplatu = common.FormatNumberWithSystemLocale(zaNaplatu-uAvans-uPor, 2)
}

// robnoStampaFakturaProfak returns the brojevi of the avansni računi of a PROFAK (a comma separated
// list of brojevi; the parts that are not numbers are skipped).
func robnoStampaFakturaProfak(profak string) []int64 {
	var brojevi []int64
	for _, part := range strings.Split(profak, ",") {
		if broj, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
			brojevi = append(brojevi, broj)
		}
	}
	return brojevi
}

// firstNonEmpty returns the first of the values that is not empty.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// robnoStampaFakturaKolicina renders the količina of a stavka: without decimals when it is a whole
// number ("125"), otherwise with its decimals, like the legacy report.
func robnoStampaFakturaKolicina(value float64) string {
	if value == math.Trunc(value) {
		return common.FormatNumberWithSystemLocale(value, 0)
	}
	return common.FormatNumberWithSystemLocale(value, 3)
}

// robnoStampaFakturaRound rounds an amount to the para.
func robnoStampaFakturaRound(value float64) float64 {
	return math.Round(value*100) / 100
}

// robnoStampaFakturaStopa renders a percentage of the print ("20%", "7,5%").
func robnoStampaFakturaStopa(value float64) string {
	precision := 0
	if value != math.Trunc(value) {
		precision = 2
	}
	return common.FormatNumberWithSystemLocale(value, precision) + "%"
}

// robnoStampaFakturaNonZero renders a number of the print, "" when it is 0 (an empty field of the
// legacy report).
func robnoStampaFakturaNonZero(value float64, precision int) string {
	if value == 0 {
		return ""
	}
	return common.FormatNumberWithSystemLocale(value, precision)
}

// robnoStampaFakturaSifra renders the šifra of the artikal of a stavka ("" for the stavke without an
// artikal, e.g. the usluge).
func robnoStampaFakturaSifra(sifra int64) string {
	if sifra == 0 {
		return ""
	}
	return fmt.Sprintf("%d", sifra)
}

//
// Common parts of the prints of the robni dokumenti
//

//
// Štampa izvozne fakture: the report RobnoStampaFakturaIzvoz (the faktura in a foreign valuta, the
// legacy PR_RPT_FAKTURA_OTPIZV with its query PR_QRY_FAKTURE)
//

// robnoStampaFakturaIzvozIzjave are the izjave of the exporter of the izvozna faktura by rdok.izjizv.
var robnoStampaFakturaIzvozIzjave = map[int64]string{
	1: "The exporter of the products covered by this document (customs autorization No. ____________) declares that, except where otherwise clearly indicated, these products are of Serbian preferential origin",
	2: "The exporter of the products covered by this document (customs autorization No. ____________) declares that, except where otherwise clearly indicated, these products are of Serbian preferential origin - no cumulation applied",
}

// GetStampaFakturaIzvoz returns the izvozne fakture of the selection, ready to print (the legacy
// PR_RPT_FAKTURA_OTPIZV): one view per robni dokument in the order of the documents and of the rbr of
// their stavke, without the ambalaža (šifra 990000 and over), and the izdavalac (fvr) of the print. The
// selection is the common one of the prints (robnoStampaDokumentaUslovi). The stavke are priced in the
// valuta of the document (rpro.cenaval): the total value of a stavka is the količina times that cena,
// its discount the total value times the % rabata. TOTAL is the sum of the total values, DISCOUNT the
// sum of the discounts, SUBTOTAL the total less the discount, the guarantee discount (rdok.ugrabat %)
// and the discount for payment in advance (rdok.pkase %) are taken from it in turn, and the TOTAL FOR
// PAYMENT is what remains; the total in RSD is the total for payment at the kurs of the document.
// withKomercOpis adds the komercijalni opis of the artikal to its naziv (the legacy ipCBOX_KOMERCOPIS).
func (s *RobnoDokumentaResource) GetStampaFakturaIzvoz(ctx context.Context, params domain.RobnoStampaFakturaParams, withKomercOpis bool) ([]domain.RobnoStampaFakturaIzvozView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	magacin, err := s.stampaFakturaMagacin(ctx, userSession, params.MagaciniID)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(rdok.pla, '') as pla,
			coalesce(rdok.vozac, '') as vozac,
			coalesce(rdok.brvozila, '') as brvozila,
			coalesce(rdok.foot, '') as foot,
			coalesce(rdok.kurs, 0) as kurs,
			coalesce(rdok.pkase, 0) as pkase,
			coalesce(rdok.ugrabat, 0) as ugrabat,
			coalesce(rdok.zirorac, '') as zirorac,
			coalesce(rdok.posuslnab, '') as posuslnab,
			coalesce(rdok.bttowght, 0) as bttowght,
			coalesce(rdok.netwght, 0) as netwght,
			coalesce(rdok.paritet, '') as paritet,
			coalesce(rdok.izjizv, 0) as izjizv,
			coalesce(valute.oznval, '') as valuta,
			coalesce(edok.brotp, 0) as brotp,
			edok.datotp,
			coalesce(p.naziv, '') as kupacnaziv,
			coalesce(p.adresa, '') as kupacadresa,
			coalesce(p.pobro, 0) as kupacpobro,
			coalesce(p.mesto, '') as kupacmesto,
			coalesce(p.tippdv, 0) as kupactippdv,
			coalesce(p.pib, '') as kupacpib,
			coalesce(p.budzetski, false) as kupacbudzetski,
			coalesce(p.jbkjs, '') as kupacjbkjs,
			coalesce(p.jmbg, '') as kupacjmbg,
			coalesce(p.bpg, '') as kupacbpg,
			coalesce(p.index, '') as kupacindex,
			coalesce(p.matbr, '') as kupacmbr,
			coalesce(p.telefon, '') as kupactelefon,
			coalesce(misp.naziv, '') as mispnaziv,
			coalesce(misp.adresa, '') as mispadresa,
			coalesce(misp.pobro, 0) as misppobro,
			coalesce(misp.mesto, '') as mispmesto,
			coalesce(nullif(misp.gln, 0)::text, '') as mispgln,
			coalesce(bnk.bnkdes, '') as bankanaziv,
			coalesce(bnk.swiftadr, '') as bankaswift,
			coalesce(bnk.brojrac, '') as bankaracun,
			coalesce(bnk.beneficiary, '') as bankakorisnik,
			coalesce(bnk.corrbank, '') as bankakorespondent,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(nullif(rpro.naz1, ''), rsif.naziv, '') as naziv,
			coalesce(rsif.komercopis, '') as komercopis,
			coalesce(rsif.zemljaproizv, '') as zemlja,
			coalesce(nullif(rsif.tarifnaozn, 0)::text, '') as pcn,
			coalesce(nullif(rpro.jm, ''), rsif.jm, '') as jm,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.cenaval, 0) as cenaval,
			coalesce(rpro.rab, 0) as rab
		from rpro`, true)
	qb.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin("left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin("left join valute on valute.sifval = rdok.sifval")
	// The despatch note: the otpremnica (edok) of the document.
	qb.AddJoin("left join edok on edok.rdokid = rdok.edokid and coalesce(rdok.edokid, 0) <> 0")
	// The kupac: the partner of the first fkpl of the konto and šifra of the document.
	qb.AddJoin(`left join lateral (select f.idpartneri from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
		order by f.vkonta desc, f.idfkpl limit 1) kup on true`)
	qb.AddJoin("left join partneri p on p.idpartneri = kup.idpartneri")
	// The destination: the fisp of the kupac and of rdok.mi.
	qb.AddJoin(`left join lateral (select f.naziv, f.adresa, f.pobro, f.mesto, f.gln from fisp f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
			and f.mi = coalesce(rdok.mi, 0)
		order by f.fispid limit 1) misp on true`)
	// The bank instructions of the document (bnkizv of rdok.sifbank).
	qb.AddJoin(`left join lateral (select b.bnkdes, b.swiftadr, b.brojrac, b.beneficiary, b.corrbank from bnkizv b
		where b.god = rdok.god and b.kar = rdok.kar and b.sifbank = rdok.sifbank
		order by b.bnkizvid limit 1) bnk on true`)
	robnoStampaDokumentaUslovi(qb, userSession, params)
	qb.AddCustomCondition("coalesce(rpro.sifra, 0) < 990000")
	qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	sqlQuery, args := qb.Build()
	rows, err := s.stampaFakturaIzvozRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	fakture := []domain.RobnoStampaFakturaIzvozView{}
	for i := 0; i < len(*rows); {
		j := i
		for j < len(*rows) && (*rows)[j].RdokID == (*rows)[i].RdokID {
			j++
		}
		fakture = append(fakture, robnoStampaFakturaIzvozView(firma, magacin, (*rows)[i:j], withKomercOpis))
		i = j
	}
	return fakture, firma, nil
}

// robnoStampaFakturaIzvozView builds the printed izvozna faktura of the stavke of one robni dokument
// (see GetStampaFakturaIzvoz).
func robnoStampaFakturaIzvozView(firma domain.RobnoStampaFakturaFirmaDto, magacin robnoStampaFakturaMagacin, rows []domain.RobnoStampaFakturaIzvozRowDto, withKomercOpis bool) domain.RobnoStampaFakturaIzvozView {
	dok := rows[0]
	fak := domain.RobnoStampaFakturaIzvozView{
		KupacNaziv:    dok.KupacNaziv,
		KupacAdresa:   dok.KupacAdresa,
		KupacMesto:    robnoStampaFakturaIzvozMesto(dok.KupacPobro, dok.KupacMesto),
		KupacPib:      robnoStampaFakturaIzvozPib(dok),
		KupacMbr:      "Matični broj : " + dok.KupacMbr,
		KupacTelefon:  "Tel/fax : " + dok.KupacTelefon,
		UslovPlacanja: dok.Pla,
		Vozac:         dok.Vozac,
		BrojVozila:    dok.Brvozila,
		BrojDokumenta: fmt.Sprintf("%d-%d", dok.Vrd, dok.Dokum),
		DatumFakture:  common.FormatNullTime(dok.Dadok, common.DateLayout),
		Valuta:        dok.Valuta,
		Napomena:      dok.Foot,
		Paritet:       dok.Paritet,
		BrutoTezina:   common.FormatNumberWithSystemLocale(dok.Bttowght, 3),
		NetoTezina:    common.FormatNumberWithSystemLocale(dok.Netwght, 3),
		Izjava:        robnoStampaFakturaIzvozIzjave[dok.Izjizv],

		BankaNaziv:        dok.BankaNaziv,
		BankaSwift:        dok.BankaSwift,
		BankaRacun:        dok.BankaRacun,
		BankaKorisnik:     dok.BankaKorisnik,
		BankaKorespondent: dok.BankaKorespondent,
	}
	if dok.KupacNaziv != "" {
		fak.KupacSifra = strings.TrimSpace(dok.Fkto + " " + dok.Fana)
	}
	// The izdavalac: the adresa with the adresa of the magacin, the mesto of the magacin as the place of
	// issue, the telefon with the international prefix and the account of the document (rdok.zirorac;
	// rdok.posuslnab when the document has none).
	fak.FirmaAdresa = strings.TrimSpace(firma.Pobro+" "+firma.Mesto) + " , " + firma.Adresa
	fak.MestoIzdavanja = firma.Mesto
	if magacin.found {
		fak.MestoIzdavanja = magacin.mesto
		if magacin.adresa != "" {
			fak.FirmaAdresa += " , " + magacin.adresa
		}
	}
	if firma.Apr != "" {
		fak.FirmaApr = "Rešenje APR " + firma.Apr
	}
	fak.FirmaPib = "PIB : " + firma.Pib
	fak.FirmaMbr = "Matični broj : " + firma.Matbr
	fak.FirmaSifdel = "Šifra delatnosti : " + firma.Sifdel
	fak.FirmaTel = "Telefon/faks : " + robnoStampaFakturaIzvozTelefon(firma.Tel)
	fak.FirmaZiro = dok.Zirorac
	if fak.FirmaZiro == "" && dok.Posuslnab != "-" && dok.Posuslnab != "-1" {
		fak.FirmaZiro = dok.Posuslnab
	}
	if dok.MispNaziv != "" {
		fak.Destinacija = "Destination:" + dok.MispNaziv
		fak.Destinacija2 = dok.MispAdresa + ", " + robnoStampaFakturaIzvozMesto(dok.MispPobro, dok.MispMesto)
		if dok.MispGln != "" {
			fak.Destinacija2 += " GLN:" + dok.MispGln
		}
	}
	if dok.Brotp != 0 {
		fak.Otpremnica = fmt.Sprintf("%d / %s", dok.Brotp, common.FormatNullTime(dok.Datotp, common.DateLayout))
	}

	var svega, rabat float64
	for _, st := range rows {
		iznos := robnoStampaFakturaRound(st.Kolic * st.Cenaval)
		popust := robnoStampaFakturaRound(iznos * st.Rab / 100)
		svega += iznos
		rabat += popust
		naziv := st.Naziv
		if withKomercOpis && st.Komercopis != "" {
			naziv = strings.TrimSpace(naziv + " " + st.Komercopis)
		}
		fak.Stavke = append(fak.Stavke, domain.RobnoStampaFakturaIzvozStavka{
			Rbr:      fmt.Sprintf("%d", st.Rbr),
			Sifra:    robnoStampaFakturaSifra(st.Sifra),
			Naziv:    naziv,
			Zemlja:   st.Zemlja,
			Pcn:      st.Pcn,
			Jm:       st.Jm,
			Kolicina: common.FormatNumberWithSystemLocale(st.Kolic, 3),
			Cena:     common.FormatNumberWithSystemLocale(st.Cenaval, 4),
			Rabat:    common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			Iznos:    common.FormatNumberWithSystemLocale(iznos, 2),
		})
	}
	neto := robnoStampaFakturaRound(svega - rabat)
	ugRabat := robnoStampaFakturaRound(neto * dok.Ugrabat / 100)
	kasa := robnoStampaFakturaRound((neto - ugRabat) * dok.Pkase / 100)
	zaNaplatu := robnoStampaFakturaRound(neto - ugRabat - kasa)
	fak.Svega = common.FormatNumberWithSystemLocale(svega, 2)
	fak.Rabat = common.FormatNumberWithSystemLocale(rabat, 2)
	fak.ImaPopuste = dok.Ugrabat != 0 || dok.Pkase != 0
	fak.Neto = common.FormatNumberWithSystemLocale(neto, 2)
	fak.UgRabatProc = common.FormatNumberWithSystemLocale(dok.Ugrabat, 2) + "%"
	fak.UgRabat = common.FormatNumberWithSystemLocale(ugRabat, 2)
	fak.KasaProc = common.FormatNumberWithSystemLocale(dok.Pkase, 2) + "%"
	fak.Kasa = common.FormatNumberWithSystemLocale(kasa, 2)
	fak.ZaNaplatu = common.FormatNumberWithSystemLocale(zaNaplatu, 2)
	fak.UkupnoRsd = common.FormatNumberWithSystemLocale(robnoStampaFakturaRound(zaNaplatu*dok.Kurs), 2)
	return fak
}

// robnoStampaFakturaIzvozTelefon returns the telefon of the firm in the international format for the
// izvozna faktura: a number with the international prefix (+ or 00) is printed as it is (00 becomes
// +), a domestic number (0xx...) gets the prefix of Serbia (+381) instead of its leading 0. The legacy
// report always put "++381 " before the number without its first character, which doubles the prefix
// of a number that already has it.
func robnoStampaFakturaIzvozTelefon(tel string) string {
	tel = strings.TrimSpace(tel)
	switch {
	case tel == "" || strings.HasPrefix(tel, "+"):
		return tel
	case strings.HasPrefix(tel, "00"):
		return "+" + tel[2:]
	case strings.HasPrefix(tel, "0"):
		return "+381 " + tel[1:]
	}
	return tel
}

// robnoStampaFakturaIzvozMesto returns the poštanski broj and the mesto of an adresa (without the
// poštanski broj when it is not set, e.g. a foreign partner).
func robnoStampaFakturaIzvozMesto(pobro int64, mesto string) string {
	if pobro == 0 {
		return strings.TrimSpace(mesto)
	}
	return strings.TrimSpace(fmt.Sprintf("%d %s", pobro, mesto))
}

// robnoStampaFakturaIzvozPib returns the PIB line of the kupac of the izvozna faktura by the PDV
// status of the partner (partneri.tippdv), like the legacy report (Srbija): PIB (with the JBKJS of a
// budžetski korisnik) for the PDV obveznici and the others with a PIB (1, 2), JMBG and BPG for the
// poljoprivrednici (3), JMBG for the fizička lica (4) and JMBG and INDEX for 5.
func robnoStampaFakturaIzvozPib(dok domain.RobnoStampaFakturaIzvozRowDto) string {
	switch {
	case dok.KupacTipPdv <= 2:
		pib := "PIB :" + dok.KupacPib
		if dok.KupacBudzetski {
			pib += "    JBKJS : " + dok.KupacJbkjs
		}
		return pib
	case dok.KupacTipPdv == 3:
		var pib string
		if dok.KupacJmbg != "" {
			pib = "JMBG :" + dok.KupacJmbg + " ; "
		}
		if dok.KupacBpg != "" {
			pib += "BPG :" + dok.KupacBpg
		}
		return pib
	case dok.KupacTipPdv == 4 && dok.KupacJmbg != "":
		return "JMBG :" + dok.KupacJmbg
	case dok.KupacTipPdv == 5:
		var pib string
		if dok.KupacJmbg != "" {
			pib = "JMBG :" + dok.KupacJmbg + " ; "
		}
		if dok.KupacIndex != "" {
			pib += "INDEX :" + dok.KupacIndex
		}
		return pib
	}
	return ""
}

// robnoStampaDokumentaUslovi adds to a query of the stavke (rpro) of the robni dokumenti (rdok) the
// selection common to the prints of every vrsta dokumenta: the period of the session, the vrsta naloga,
// the ranges of the broj naloga, of the broj dokumenta and of the vrsta dokumenta, the magacin, the
// selected document (RdokID), the vrsta dokumenta (Vrd) and the range of the datum naloga.
func robnoStampaDokumentaUslovi(qb *common.QueryBuilder, userSession *domain.UserSession, params domain.RobnoStampaFakturaParams) {
	qb.AddEqual("rpro.god", userSession.SelectedGod)
	qb.AddEqual("rpro.kar", userSession.SelectedKar)
	qb.AddEqual("rdok.tipdok", params.Tipdok)
	addNumberCondition(qb, "rpro.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rpro.nalog", params.DoNaloga, "<=")
	addNumberCondition(qb, "rpro.dokum", params.OdDokum, ">=")
	addNumberCondition(qb, "rpro.dokum", params.DoDokum, "<=")
	addNumberCondition(qb, "rpro.vrd", params.OdVrd, ">=")
	addNumberCondition(qb, "rpro.vrd", params.DoVrd, "<=")
	if params.MagaciniID != 0 {
		qb.AddEqual("rpro.magaciniid", params.MagaciniID)
	}
	if params.RdokID != 0 {
		qb.AddEqual("rdok.rdokid", params.RdokID)
	}
	addNumberCondition(qb, "rpro.vrd", params.Vrd, "=")
	qb.AddCondition("rdok.danal", params.OdDanal, ">=")
	qb.AddCondition("rdok.danal", params.DoDanal, "<=")
}

// GetStampaDokument returns what the print of the "Štampa" sub-tab needs to choose the print of a robni
// dokument (domain.RobnoStampaDokumentDto): of the selected document (rdokID) of the period of the
// session, else of the vrsta dokumenta of the filter (vrd, without the data of a document). It is the
// empty dto when neither is given or found.
func (s *RobnoDokumentaResource) GetStampaDokument(ctx context.Context, rdokID int64, vrd string) (domain.RobnoStampaDokumentDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.RobnoStampaDokumentDto{}, fmt.Errorf("no user session found")
	}
	var qb *common.QueryBuilder
	if rdokID != 0 {
		qb = common.NewQueryBuilder(`
			select
				rdok.rdokid,
				coalesce(rdok.vrd, 0) as vrd,
				coalesce(dokvrsta.grpdok, '') as grpdok,
				coalesce(dokvrsta.dokozn, '') as dokozn,
				coalesce(dokvrsta.dodoznfak, '') as dodoznfak,
				coalesce(rdok.sifval, 0) as sifval,
				coalesce((select fvr.sifval from fvr where fvr.god = rdok.god and fvr.kar = rdok.kar
					limit 1), 0) as domacavaluta,
				coalesce(rdok.knjige_1, '') as knjige,
				exists (select 1 from rpro where rpro.rdokid = rdok.rdokid) as imastavke
			from rdok`, true)
		qb.AddJoin("left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
		qb.AddEqual("rdok.god", userSession.SelectedGod)
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
		qb.AddEqual("rdok.rdokid", rdokID)
	} else {
		vrsta, err := strconv.Atoi(strings.TrimSpace(vrd))
		if err != nil {
			return domain.RobnoStampaDokumentDto{}, nil
		}
		qb = common.NewQueryBuilder(`
			select
				0::bigint as rdokid,
				coalesce(dokvrsta.vrd, 0) as vrd,
				coalesce(dokvrsta.grpdok, '') as grpdok,
				coalesce(dokvrsta.dokozn, '') as dokozn,
				coalesce(dokvrsta.dodoznfak, '') as dodoznfak,
				0::bigint as sifval,
				0::bigint as domacavaluta,
				'' as knjige,
				false as imastavke
			from dokvrsta`, true)
		qb.AddEqual("dokvrsta.god", userSession.SelectedGod)
		qb.AddEqual("dokvrsta.kar", userSession.SelectedKar)
		qb.AddEqual("dokvrsta.vrd", vrsta)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.stampaDokumentRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.RobnoStampaDokumentDto{}, err
	}
	if len(*entities) == 0 {
		return domain.RobnoStampaDokumentDto{}, nil
	}
	return (*entities)[0], nil
}

//
// Štampa popisa: the report RobnoStampaPopis (vrsta dokumenta 101)
//

// GetStampaPopis returns the popisi of the selection, ready to print: one view per robni dokument in
// the order of the documents and of the rbr of their stavke, and the izdavalac (fvr) of the print. The
// selection is the common one of the prints (robnoStampaDokumentaUslovi); the stavke show the nabavna
// cena (rpro.ncena) with the nabavna vrednost (količina * nabavna cena) and the cena with the iznos of
// the stavka, the document the totals of both values.
func (s *RobnoDokumentaResource) GetStampaPopis(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rbr")
}

// GetStampaOpstiDokument returns the opšti dokumenti of the selection (the groups OPD, POT, KOL and FIN
// of the vrste dokumenta, the legacy RPT_OPDSTAMPA with its query ROB_QRY_OPDSTAMPA), ready to print:
// one view per robni dokument with its stavke in the order they were entered (rpro.rproid), the totals
// of the nabavna vrednost, of the iznos and of the količina, and the izdavalac (fvr) of the print. The
// template prints the stavke by the group of the document (Grupa).
func (s *RobnoDokumentaResource) GetStampaOpstiDokument(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rproid")
}

// stampaOpstiDokumenti returns the popisi or the opšti dokumenti of the selection with their stavke in
// the order orderBy (see GetStampaPopis and GetStampaOpstiDokument).
func (s *RobnoDokumentaResource) stampaOpstiDokumenti(ctx context.Context, params domain.RobnoStampaFakturaParams, orderBy string) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`select
			rdok.rdokid,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(dokvrsta.opis, '') as vrdopis,
			coalesce(dokvrsta.grpdok, '') as grpdok,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.mag, 0) as mag,
			coalesce(mg.opis, '') as magnaziv,
			coalesce(oj.ojozn || '-' || oj.naziv, '') as oj,
			coalesce(mt.mtroska || '-' || mt.opis, '') as mt,
			coalesce(oj1.ojozn || '-' || oj1.naziv, '') as ojprijem,
			coalesce(mt1.mtroska || '-' || mt1.opis, '') as mtprijem,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.konto, '') as konto,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(nullif(rpro.naz1, ''), rsif.naziv, '') as naz1,
			coalesce(nullif(rpro.jm, ''), rsif.jm, '') as jm,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.ncena, 0) as ncena,
			coalesce(rpro.cena, 0) as cena,
			coalesce(rpro.iznos, 0) as iznos
		from rpro`, true)
	qb.AddJoin(" inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin(" left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin(" left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(" left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	qb.AddJoin(" left join orgjed oj on oj.idorgjed = rdok.idorgjed")
	qb.AddJoin(" left join mestotr mt on mt.mestotrid = rdok.mestotrid")
	qb.AddJoin(" left join orgjed oj1 on oj1.idorgjed = rdok.idorgjed1")
	qb.AddJoin(" left join mestotr mt1 on mt1.mestotrid = rdok.mestotrid1")
	robnoStampaDokumentaUslovi(qb, userSession, params)
	qb.AddOrderBy(orderBy)
	sqlQuery, args := qb.Build()
	rows, err := s.stampaPopisRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}

	popisi := []domain.RobnoStampaPopisView{}
	var nabavna, iznos, kolicina float64
	zatvori := func() {
		if len(popisi) == 0 {
			return
		}
		last := &popisi[len(popisi)-1]
		last.UkupnoNabavnaVrednost = common.FormatNumberWithSystemLocale(nabavna, 2)
		last.UkupnoIznos = common.FormatNumberWithSystemLocale(iznos, 2)
		last.UkupnoKolicina = common.FormatNumberWithSystemLocale(kolicina, 3)
	}
	for i, r := range *rows {
		if i == 0 || (*rows)[i-1].RdokID != r.RdokID {
			zatvori()
			nabavna, iznos, kolicina = 0, 0, 0
			popisi = append(popisi, domain.RobnoStampaPopisView{
				BrojDokumenta:     fmt.Sprintf("%d-%d", r.Vrd, r.Dokum),
				DatumDokumenta:    common.FormatNullTime(r.Dadok, common.DateLayout),
				VrstaDokumenta:    fmt.Sprintf("%d-%s", r.Vrd, r.VrdOpis),
				Nalog:             fmt.Sprintf("%s-%d", r.Tipdok, r.Nalog),
				Magacin:           fmt.Sprintf("%d-%s", r.Mag, r.MagNaziv),
				Oj:                r.Oj,
				MestoTroska:       r.MestoTroska,
				OjPrijem:          r.OjPrijem,
				MestoTroskaPrijem: r.MestoTroskaPrijem,
				Grupa:             r.Grpdok,
				IzvorniDokument:   r.Dokiz,
			})
		}
		nabavnaVrednost := robnoStampaFakturaRound(r.Kolic * r.Ncena)
		nabavna += nabavnaVrednost
		iznos += r.Iznos
		kolicina += r.Kolic
		pop := &popisi[len(popisi)-1]
		pop.Stavke = append(pop.Stavke, domain.RobnoStampaPopisStavka{
			Rbr:             fmt.Sprintf("%d", r.Rbr),
			Konto:           r.Konto,
			Sifra:           robnoStampaFakturaSifra(r.Sifra),
			Naziv:           r.Naz1,
			Jm:              r.Jm,
			Kolicina:        common.FormatNumberWithSystemLocale(r.Kolic, 3),
			NabavnaCena:     common.FormatNumberWithSystemLocale(r.Ncena, 3),
			NabavnaVrednost: common.FormatNumberWithSystemLocale(nabavnaVrednost, 2),
			Cena:            common.FormatNumberWithSystemLocale(r.Cena, 2),
			Iznos:           common.FormatNumberWithSystemLocale(r.Iznos, 2),
		})
	}
	zatvori()
	return popisi, firma, nil
}

// GetStampaKalkulacija returns the kalkulacije veleprodaje (prijemni listovi, the legacy
// RPT_ROB_KALKULACIJA) of the selection, ready to print: one view per robni dokument in the order of the
// documents and of the rbr of their stavke, and the izdavalac (fvr) of the print. The selection is the
// common one of the prints (robnoStampaDokumentaUslovi). Per stavka:
//   - fakturna vrednost = količina * fakturna cena (rpro.fcena), vrednost rabata = fakturna vrednost *
//     % rabata, neto fakturna vrednost = fakturna vrednost - vrednost rabata, the neto fakturna cena the
//     fakturna cena less its rabat rounded to 2 decimals;
//   - the zavisni trošak (rpro.ztro) and the interni zavisni trošak (rpro.ztrin) with the oznaka "%"
//     (rpro.tz, rpro.tzi) are percentages of the neto fakturna vrednost, the zavisni trošak of other
//     dobavljači (rpro.ztrof) an amount per unit;
//   - nabavna vrednost = neto fakturna vrednost + the troškovi, nabavna cena = rpro.ncena;
//   - veleprodajna vrednost = količina * veleprodajna cena (rpro.vpcena), vrednost marže = veleprodajna
//   - nabavna vrednost, % marže = vrednost marže / nabavna vrednost.
//
// Every vrednost is rounded to 2 decimals and the totals are the sums of the rounded vrednosti. A
// domestic document has the porezi of the document of the dobavljač by poreska tarifa (osnovica = the
// neto fakturna vrednost of the stavke of the tarifa, porez = osnovica * stopa); a document in a foreign
// valuta (Devizni) has the valuta and the kurs instead.
func (s *RobnoDokumentaResource) GetStampaKalkulacija(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaKalkulacijaView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`select
			rdok.rdokid,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			rdok.datiz,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(rdok.otp, '') as otp,
			coalesce(rdok.pla, '') as pla,
			coalesce(rdok.rok, 0) as rok,
			coalesce(rdok.foot, '') as foot,
			coalesce(rdok.mag, 0) as mag,
			coalesce(mg.opis, '') as magnaziv,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(p.naziv, dob.naziv, '') as partnernaziv,
			coalesce(p.adresa, '') as partneradresa,
			trim(coalesce(nullif(p.pobro, 0)::text, '') || ' ' || coalesce(p.mesto, '')) as partnermesto,
			coalesce(p.pib, '') as partnerpib,
			coalesce(rdok.sifval, 0) as sifval,
			coalesce((select fvr.sifval from fvr where fvr.god = rdok.god and fvr.kar = rdok.kar limit 1), 0) as domacavaluta,
			coalesce(valute.oznval, '') as valutaoznaka,
			coalesce(rdok.kurs, 0) as kurs,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(nullif(rpro.naz1, ''), rsif.naziv, '') as naz1,
			coalesce(nullif(rpro.jm, ''), rsif.jm, '') as jm,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.fcena, 0) as fcena,
			coalesce(rpro.rab, 0) as rab,
			coalesce(rpro.tz, '') as tz,
			coalesce(rpro.ztro, 0) as ztro,
			coalesce(rpro.ztrof, 0) as ztrof,
			coalesce(rpro.tzi, '') as tzi,
			coalesce(rpro.ztrin, 0) as ztrin,
			coalesce(rpro.ncena, 0) as ncena,
			coalesce(nullif(rpro.vpcena, 0), rpro.mcena, 0) as vpcena,
			coalesce(rpro.po, 0) as po,
			coalesce(tar.pt, '') as tarifa,
			coalesce(tar.pp, rpro.pdvpct, 0) as stopa
		from rpro`, true)
	qb.AddJoin(" inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin(" left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(" left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	// The dobavljač: the first fkpl of the konto and šifra of the document (the analitički konto first),
	// so that a konto and šifra with more fkpl rows does not repeat the stavke.
	qb.AddJoin(` left join lateral (select f.naziv, f.idpartneri from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
		order by f.vkonta desc, f.idfkpl limit 1) dob on true`)
	qb.AddJoin(" left join partneri p on p.idpartneri = dob.idpartneri")
	qb.AddJoin(" left join valute on valute.sifval = rdok.sifval")
	// The poreska tarifa (rpor.pt) and stopa (rpor.pp) of the poreska oznaka of the stavka in force on
	// the date of the stavka.
	qb.AddJoin(` left join lateral (select r.pt, r.pp from rpor r
		where r.po = rpro.po and r.datum <= rpro.dadok
		order by r.datum desc limit 1) tar on true`)
	robnoStampaDokumentaUslovi(qb, userSession, params)
	qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	sqlQuery, args := qb.Build()
	rows, err := s.stampaKalkulacijaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}

	kalkulacije := []domain.RobnoStampaKalkulacijaView{}
	for i := 0; i < len(*rows); {
		j := i
		for j < len(*rows) && (*rows)[j].RdokID == (*rows)[i].RdokID {
			j++
		}
		kalkulacije = append(kalkulacije, robnoStampaKalkulacijaView((*rows)[i:j]))
		i = j
	}
	return kalkulacije, firma, nil
}

// robnoStampaKalkulacijaVrednosti are the vrednosti of a stavka (or the totals) of the kalkulacija.
type robnoStampaKalkulacijaVrednosti struct {
	fakturna, rabat, neto, trosak, trosakDobavljaca, interni, nabavna, marza, vp float64
}

func (v *robnoStampaKalkulacijaVrednosti) add(o robnoStampaKalkulacijaVrednosti) {
	v.fakturna += o.fakturna
	v.rabat += o.rabat
	v.neto += o.neto
	v.trosak += o.trosak
	v.trosakDobavljaca += o.trosakDobavljaca
	v.interni += o.interni
	v.nabavna += o.nabavna
	v.marza += o.marza
	v.vp += o.vp
}

// robnoStampaKalkulacijaTrosak returns the vrednost of a zavisni trošak of a stavka: a percentage of the
// neto fakturna vrednost with the oznaka "%", nothing without an oznaka ("-").
func robnoStampaKalkulacijaTrosak(oznaka string, trosak, neto float64) float64 {
	if strings.TrimSpace(oznaka) == "%" {
		return robnoStampaFakturaRound(neto * trosak / 100)
	}
	return 0
}

// robnoStampaKalkulacijaOznaka returns the oznaka of a zavisni trošak as printed ("-" without one).
func robnoStampaKalkulacijaOznaka(oznaka string) string {
	if oznaka = strings.TrimSpace(oznaka); oznaka != "" {
		return oznaka
	}
	return "-"
}

// robnoStampaKalkulacijaPorez is a poreska tarifa of the document of the dobavljač while it is summed.
type robnoStampaKalkulacijaPorez struct {
	tarifa          string
	stopa, osnovica float64
}

// robnoStampaKalkulacijaView builds the printed kalkulacija of the stavke of one robni dokument (see
// GetStampaKalkulacija).
func robnoStampaKalkulacijaView(rows []domain.RobnoStampaKalkulacijaRowDto) domain.RobnoStampaKalkulacijaView {
	r := rows[0]
	kal := domain.RobnoStampaKalkulacijaView{
		BrojDokumenta:            fmt.Sprintf("%d-%d", r.Vrd, r.Dokum),
		Devizni:                  r.Sifval > 0 && r.Sifval != r.DomacaValuta,
		KontoDobavljaca:          r.Fkto,
		SifraDobavljaca:          r.Fana,
		DobavljacNaziv:           r.PartnerNaziv,
		DobavljacAdresa:          r.PartnerAdresa,
		DobavljacMesto:           r.PartnerMesto,
		DobavljacPib:             r.PartnerPib,
		Magacin:                  fmt.Sprintf("%d-%s", r.Mag, r.MagNaziv),
		Nalog:                    fmt.Sprintf("%s/%d", r.Tipdok, r.Nalog),
		DatumDokumenta:           common.FormatNullTime(r.Dadok, common.DateLayout),
		NacinDopremanja:          r.Otp,
		DokumentDobavljaca:       r.Dokiz,
		DatumDokumentaDobavljaca: common.FormatNullTime(r.Datiz, common.DateLayout),
		NacinPlacanja:            r.Pla,
		RokPlacanja:              fmt.Sprintf("%d", r.Rok),
		Napomena:                 r.Foot,
		Valuta:                   strings.TrimSpace(fmt.Sprintf("%d %s", r.Sifval, r.ValutaOznaka)),
		Oznaka:                   r.ValutaOznaka,
		Kurs:                     strconv.FormatFloat(r.Kurs, 'f', -1, 64),
	}
	var ukupno robnoStampaKalkulacijaVrednosti
	porezi := []*robnoStampaKalkulacijaPorez{}
	poreziPo := map[int64]*robnoStampaKalkulacijaPorez{}
	for _, st := range rows {
		var v robnoStampaKalkulacijaVrednosti
		v.fakturna = robnoStampaFakturaRound(st.Kolic * st.Fcena)
		v.rabat = robnoStampaFakturaRound(v.fakturna * st.Rab / 100)
		v.neto = robnoStampaFakturaRound(v.fakturna - v.rabat)
		v.trosak = robnoStampaKalkulacijaTrosak(st.Tz, st.Ztro, v.neto)
		v.trosakDobavljaca = robnoStampaFakturaRound(st.Kolic * st.Ztrof)
		v.interni = robnoStampaKalkulacijaTrosak(st.Tzi, st.Ztrin, v.neto)
		v.nabavna = robnoStampaFakturaRound(v.neto + v.trosak + v.trosakDobavljaca + v.interni)
		v.vp = robnoStampaFakturaRound(st.Kolic * st.Vpcena)
		v.marza = robnoStampaFakturaRound(v.vp - v.nabavna)
		ukupno.add(v)
		marzaPct := 0.0
		if v.nabavna != 0 {
			marzaPct = v.marza / v.nabavna * 100
		}
		kal.Stavke = append(kal.Stavke, domain.RobnoStampaKalkulacijaStavka{
			Rbr:                      fmt.Sprintf("%d", st.Rbr),
			Sifra:                    robnoStampaFakturaSifra(st.Sifra),
			Naziv:                    st.Naz1,
			Jm:                       st.Jm,
			Kolicina:                 common.FormatNumberWithSystemLocale(st.Kolic, 3),
			FakturnaCena:             common.FormatNumberWithSystemLocale(st.Fcena, 4),
			Rabat:                    common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			NetoCena:                 common.FormatNumberWithSystemLocale(st.Fcena-robnoStampaFakturaRound(st.Fcena*st.Rab/100), 4),
			OznakaTroska:             robnoStampaKalkulacijaOznaka(st.Tz),
			ZavisniTrosak:            common.FormatNumberWithSystemLocale(st.Ztro, 6),
			ZavisniTrosakDobavljaca:  common.FormatNumberWithSystemLocale(st.Ztrof, 8),
			OznakaInternog:           robnoStampaKalkulacijaOznaka(st.Tzi),
			InterniTrosak:            common.FormatNumberWithSystemLocale(st.Ztrin, 2),
			NabavnaCena:              common.FormatNumberWithSystemLocale(st.Ncena, 2),
			Marza:                    common.FormatNumberWithSystemLocale(marzaPct, 2) + "%",
			VpCena:                   common.FormatNumberWithSystemLocale(st.Vpcena, 3),
			FakturnaVrednost:         common.FormatNumberWithSystemLocale(v.fakturna, 2),
			VrednostRabata:           common.FormatNumberWithSystemLocale(v.rabat, 2),
			NetoVrednost:             common.FormatNumberWithSystemLocale(v.neto, 2),
			VrednostTroska:           common.FormatNumberWithSystemLocale(v.trosak, 2),
			VrednostTroskaDobavljaca: common.FormatNumberWithSystemLocale(v.trosakDobavljaca, 2),
			VrednostInternog:         common.FormatNumberWithSystemLocale(v.interni, 2),
			NabavnaVrednost:          common.FormatNumberWithSystemLocale(v.nabavna, 2),
			VrednostMarze:            common.FormatNumberWithSystemLocale(v.marza, 2),
			VpVrednost:               common.FormatNumberWithSystemLocale(v.vp, 2),
		})
		if pz, found := poreziPo[st.Po]; found {
			pz.osnovica += v.neto
		} else {
			pz = &robnoStampaKalkulacijaPorez{tarifa: st.Tarifa, stopa: st.Stopa, osnovica: v.neto}
			poreziPo[st.Po] = pz
			porezi = append(porezi, pz)
		}
	}
	kal.Ukupno = domain.RobnoStampaKalkulacijaStavka{
		FakturnaVrednost:         common.FormatNumberWithSystemLocale(ukupno.fakturna, 2),
		VrednostRabata:           common.FormatNumberWithSystemLocale(ukupno.rabat, 2),
		NetoVrednost:             common.FormatNumberWithSystemLocale(ukupno.neto, 2),
		VrednostTroska:           common.FormatNumberWithSystemLocale(ukupno.trosak, 2),
		VrednostTroskaDobavljaca: common.FormatNumberWithSystemLocale(ukupno.trosakDobavljaca, 2),
		VrednostInternog:         common.FormatNumberWithSystemLocale(ukupno.interni, 2),
		NabavnaVrednost:          common.FormatNumberWithSystemLocale(ukupno.nabavna, 2),
		VrednostMarze:            common.FormatNumberWithSystemLocale(ukupno.marza, 2),
		VpVrednost:               common.FormatNumberWithSystemLocale(ukupno.vp, 2),
	}
	if kal.Devizni {
		return kal
	}
	var iznos, osnovica, obracunat float64
	for _, pz := range porezi {
		osn := robnoStampaFakturaRound(pz.osnovica)
		por := robnoStampaFakturaRound(osn * pz.stopa / 100)
		iznos += osn + por
		osnovica += osn
		obracunat += por
		kal.Porezi = append(kal.Porezi, domain.RobnoStampaKalkulacijaPorez{
			IznosDokumenta: common.FormatNumberWithSystemLocale(osn+por, 2),
			Osnovica:       common.FormatNumberWithSystemLocale(osn, 2),
			Porez:          common.FormatNumberWithSystemLocale(por, 2),
			Tarifa:         pz.tarifa,
			Stopa:          common.FormatNumberWithSystemLocale(pz.stopa, 2),
		})
	}
	kal.UkupnoPorezi = domain.RobnoStampaKalkulacijaPorez{
		IznosDokumenta: common.FormatNumberWithSystemLocale(iznos, 2),
		Osnovica:       common.FormatNumberWithSystemLocale(osnovica, 2),
		Porez:          common.FormatNumberWithSystemLocale(obracunat, 2),
	}
	return kal
}

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

// robnoStampaFakturaMagacin is the magacin of the print: its mesto is the mesto izdavanja računa and
// its adresa is added to the adresa of the firm.
type robnoStampaFakturaMagacin struct {
	found  bool
	mesto  string
	adresa string
}

// GetStampaFaktura returns the fakture of the selection, ready to print: one view per robni dokument,
// in the order of the source query of the legacy report (rdokid, rbr of the stavke), and the
// izdavalac (fvr) of the print. Like the legacy report, the group of the vrste dokumenta of the print
// is the group of the vrsta "od" (OdVrd) when the selection does not give the groups.
func (s *RobnoDokumentaResource) GetStampaFaktura(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFakturaFirma(ctx, userSession)
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
			if brojevi[av.Dokum] {
				avansiDok = append(avansiDok, av)
			}
		}
		fakture = append(fakture, robnoStampaFakturaView(firma, magacin, stavke, avansiDok, ambalazaOf[rdokID], rateOf[rdokID]))
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
			coalesce(rdok.narudzb, '') as narudzb,
			coalesce(rdok.vozac, '') as vozac,
			coalesce(rdok.brvozila, '') as brvozila,
			coalesce(rdok.foot, '') as foot,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.dokiz, '') as dokiz,
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
	qb.AddJoin("left join banke on banke.god = rdok.god and banke.kar = rdok.kar and banke.sifra = rdok.sifbank::text and coalesce(rdok.sifbank, 0) <> 0")
	qb.AddJoin("left join porkat on porkat.keykat = rdok.keykat and coalesce(rdok.keykat, '') <> ''")
	qb.AddJoin(robnoStampaFakturaStopaJoin)
	qb.AddEqual("rpro.god", userSession.SelectedGod)
	qb.AddEqual("rpro.kar", userSession.SelectedKar)
	qb.AddEqual("rdok.tipdok", params.Tipdok)
	qb.AddIn("dokvrsta.grpdok", grupeDokumenata(params.GrupeDokumenata))
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

// stampaFakturaFirma returns the izdavalac of the fakture: the firm (fvr) of the session with its
// tekući računi (the banke not flagged nafakne).
func (s *RobnoDokumentaResource) stampaFakturaFirma(ctx context.Context, userSession *domain.UserSession) (domain.RobnoStampaFakturaFirmaDto, error) {
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
	qb := common.NewQueryBuilder("select magaciniid, mag, coalesce(mesto, '') as mesto, coalesce(adresa, '') as adresa from magacini", true)
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
	return robnoStampaFakturaMagacin{found: true, mesto: mag.Mesto, adresa: mag.Adresa}, nil
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

// RobnoStampaFakturaLogo returns the logo of the firm as a data URI for the <img> of the print ("" when
// the firm has no logo).
func RobnoStampaFakturaLogo(logo []byte) string {
	if len(logo) == 0 {
		return ""
	}
	return "data:" + http.DetectContentType(logo) + ";base64," + base64.StdEncoding.EncodeToString(logo)
}

// robnoStampaFakturaView builds the printed faktura of one robni dokument from its stavke (the rows of
// the source query, all of the same document), the stavke of its avansni računi, its ambalaža and its
// rate.
func robnoStampaFakturaView(firma domain.RobnoStampaFakturaFirmaDto, magacin robnoStampaFakturaMagacin, stavke []domain.RobnoStampaFakturaRowDto, avansi []domain.RobnoStampaFakturaAvansDto, ambalaza []domain.RobnoStampaFakturaRowDto, rate []domain.RobnoStampaFakturaRataDto) domain.RobnoStampaFakturaView {
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
	if dok.Rok != 0 {
		fak.RokPlacanja = fmt.Sprintf("%d", dok.Rok)
	}

	// The stavke and the totals per poreska stopa, like the legacy report: the iznos of a stavka is
	// the količina times the fakturna cena (before the rabat), its rabat is rounded to the para, its
	// osnovica is the iznos less the rabat and its PDV the osnovica times the stopa. A stavka with its
	// own rok (rpro.dani) falls due on the date of the document plus those days.
	type stopaSums struct{ osnovica, pdv float64 }
	perStopa := map[float64]*stopaSums{}
	var stope []float64
	perDanu := map[time.Time]float64{}
	imaDane := false
	var ukIznos, ukRabat, ukPdv, ukOslobodjeno float64
	for _, st := range stavke {
		rabat := robnoStampaFakturaRound(st.Iznos * st.Rab / 100)
		osnovica := st.Iznos - rabat
		pdv := robnoStampaFakturaRound(osnovica * st.Stopa / 100)
		fak.Stavke = append(fak.Stavke, domain.RobnoStampaFakturaStavka{
			Rbr:        fmt.Sprintf("%d", st.Rbr),
			Sifra:      robnoStampaFakturaSifra(st.Sifra),
			Naziv:      st.Naz1,
			Jm:         st.Jm,
			Kolicina:   robnoStampaFakturaKolicina(st.Kolic),
			Cena:       common.FormatNumberWithSystemLocale(st.Cena, 2),
			ProcRabata: common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			Rabat:      common.FormatNumberWithSystemLocale(rabat, 2),
			ProcPdv:    robnoStampaFakturaStopa(st.Stopa),
			Pdv:        common.FormatNumberWithSystemLocale(pdv, 2),
			Iznos:      common.FormatNumberWithSystemLocale(st.Iznos, 2),
		})
		if _, found := perStopa[st.Stopa]; !found {
			perStopa[st.Stopa] = &stopaSums{}
			stope = append(stope, st.Stopa)
		}
		perStopa[st.Stopa].osnovica += osnovica
		perStopa[st.Stopa].pdv += pdv
		ukIznos += st.Iznos
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
		})
	}
	// SVEGA is the iznos of the stavke (before the rabat, without the PDV), RABAT the rabat of the
	// stavke and ZA NAPLATU the osnovica (SVEGA less RABAT) with the PDV, like the legacy report.
	zaNaplatu := ukIznos - ukRabat + ukPdv
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

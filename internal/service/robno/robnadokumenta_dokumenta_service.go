package robno

import (
	"context"
	"database/sql"
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
	robnoStampaFakturaKnjizno
	robnoStampaFakturaOtpremnica
	robnoStampaFakturaProfaktura
	robnoStampaFakturaZaduzenjeCO
	robnoStampaFakturaRazduzenjeCO
)

// The titles of the faktura-otpremnica (the legacy ipTIPFAKT, the option of the print): the račun -
// otpremnica, the račun and the otpremnica.
const (
	RobnoStampaTipFakturaRacunOtpremnica = 1
	RobnoStampaTipFakturaRacun           = 2
	RobnoStampaTipFakturaOtpremnica      = 3
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
	return s.stampaFakture(ctx, params, robnoStampaFakturaVeleprodaja, 0)
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
	return s.stampaFakture(ctx, params, robnoStampaFakturaMaloprodaja, 0)
}

// GetStampaFakturaUsluge returns the fakture usluga of the selection, ready to print (the legacy
// ROB_RPT_STAMPA_FAKTURA_USLUGE2, the same source query as the faktura with the group of the vrsta
// dokumenta of the selection): the stavke and the totals like the faktura, without the avansni računi
// (see robnoStampaFakturaUslugeAvansi).
func (s *RobnoDokumentaResource) GetStampaFakturaUsluge(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaUsluge, 0)
}

// GetStampaZaduzenjeCO returns the zaduženja za obaveze (zaduženje CO, the group FCO) of the selection,
// ready to print (the legacy ROB_RPT_STAMPA_ZADUZENJA_CO, the same source query as the faktura with the
// group of the vrsta dokumenta of the selection): the stavke with their konto and analitička šifra
// (rpro.fkto, rpro.fana; the poziv na broj of the payment), the naziv of the artikal (rpro.naz1 for a
// stavka without an artikal) and the iznos, and the totals, the avansi, the ambalaža and the rate like
// the faktura. A vrsta with the predznak "-" is the storno: "Storno Račun-otpremnica" ZA POVRAĆAJ.
func (s *RobnoDokumentaResource) GetStampaZaduzenjeCO(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaZaduzenjeCO, 0)
}

// GetStampaRazduzenjeCO returns the razduženja za obaveze (razduženje CO, the group RCO) of the selection,
// ready to print (the legacy ROB_RPT_STAMPA_RAZDUZENJA_CO): like the zaduženje za obaveze (see
// GetStampaZaduzenjeCO) with the title "RAZDUŽENJE ZA OBAVEZE", and the stanje of the account of the
// kupac (rdok.fkto, rdok.fana) after the razduženje: the sum of its fakture (fpro.kat 1 and 2), of its
// uplate (fpro.kat 3 and 4) and their saldo.
func (s *RobnoDokumentaResource) GetStampaRazduzenjeCO(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	razduzenja, firma, err := s.stampaFakture(ctx, params, robnoStampaFakturaRazduzenjeCO, 0)
	if err != nil || len(razduzenja) == 0 {
		return razduzenja, firma, err
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	qb := common.NewQueryBuilder(`select
			coalesce(fpro.konto, '') as kupackonto,
			coalesce(fpro.sifra, '') as kupacsifra,
			coalesce(sum(case when fpro.kat in (1, 2) then fpro.iznos end), 0) as stanjefakture,
			coalesce(sum(case when fpro.kat in (3, 4) then fpro.iznos end), 0) as stanjeuplate
		from fpro`, true)
	qb.AddEqual("fpro.god", userSession.SelectedGod)
	qb.AddEqual("fpro.kar", userSession.SelectedKar)
	var konta []any
	for _, r := range razduzenja {
		konta = append(konta, r.KupacKonto)
	}
	qb.AddIn("fpro.konto", konta)
	qb.AddGroupBy("fpro.konto, fpro.sifra")
	sqlQuery, args := qb.Build()
	stanja, err := s.stampaFakturaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	f2 := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 2) }
	for i := range razduzenja {
		var fakture, uplate float64
		for _, st := range *stanja {
			if st.KupacKonto == razduzenja[i].KupacKonto && st.KupacSifra == razduzenja[i].KupacSifra {
				fakture, uplate = st.StanjeFakture, st.StanjeUplate
			}
		}
		razduzenja[i].StanjeFakture = f2(fakture)
		razduzenja[i].StanjeUplate = f2(uplate)
		razduzenja[i].StanjeSaldo = f2(fakture - uplate)
	}
	return razduzenja, firma, nil
}

// GetStampaFakturaOtpremnica returns the fakture-otpremnice of the selection, ready to print (the legacy
// PR_RPT_FAKTURA_OTP, printed for the fakture of the tip proizvodnje 2 and for the group FZR; its query
// PR_QRY_FAKTURE is covered by the source query of the faktura): the stavke and the totals like the
// faktura, with the PDV of one unit of the stavke and the magacin of the document (Skladište). The
// title is by tipFakture (RobnoStampaTipFaktura*: 1 and 3 "Otpremnica", 2 "Račun"; "Storno ..." for a
// vrsta with the predznak "-") and "Konačni račun" for a document that closes avansi (avansfakt); the
// broj of the document is magacin-vrsta-broj.
func (s *RobnoDokumentaResource) GetStampaFakturaOtpremnica(ctx context.Context, params domain.RobnoStampaFakturaParams, tipFakture int) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaOtpremnica, tipFakture)
}

// robnoStampaFakturaOtpremnicaNaslov returns the title of a faktura-otpremnica (see
// GetStampaFakturaOtpremnica): "Konačni račun" for a document that closes avansi, else by tipFakture
// and the storno (the predznak "-" of the vrsta dokumenta).
func robnoStampaFakturaOtpremnicaNaslov(dok domain.RobnoStampaFakturaRowDto, tipFakture int) string {
	lbl := i18n.GetInstance().Label
	if dok.Konacni {
		return lbl("Konačni račun")
	}
	storno := strings.TrimSpace(dok.Predznak) == "-"
	if tipFakture == RobnoStampaTipFakturaRacun {
		if storno {
			return lbl("Storno račun")
		}
		return lbl("Račun naslov")
	}
	if storno {
		return lbl("Storno otpremnica")
	}
	return lbl("Otpremnica")
}

// The titles of the profaktura (the legacy ipTip, the option of the print): the profaktura and the
// ponuda.
const (
	RobnoStampaTipProfakturaProfaktura = 1
	RobnoStampaTipProfakturaPonuda     = 2
)

// robnoStampaFakturaGrupaProfaktura is the group of the vrste dokumenta of the profakture
// (dokvrsta.grpdok), the documents the legacy ROB_RPT_STAMPA_PROFAKTURA prints.
const robnoStampaFakturaGrupaProfaktura = "PRO"

// GetStampaProfaktura returns the profakture (ponude) of the selection, ready to print (the legacy
// ROB_RPT_STAMPA_PROFAKTURA, the source query of the faktura with the group PRO): the stavke at the
// cena of the stavka (rpro.cena) with the rabat (per unit for the artikli of the model "D"), the
// osnovica without the taksa of the artikal (for the model "D" the osnovica of the akciza, the akciza,
// the marža less the rabat), the PDV of every stavka on its osnovica, the totals, the avansi and the za
// naplatu po danima (see robnoStampaProfakturaStavke). tipProfakture is the title
// (RobnoStampaTipProfaktura*: "PROFAKTURA Br." or "PONUDA").
func (s *RobnoDokumentaResource) GetStampaProfaktura(ctx context.Context, params domain.RobnoStampaFakturaParams, tipProfakture int) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	if params.GrupeDokumenata == "" && params.RdokID == 0 {
		params.GrupeDokumenata = robnoStampaFakturaGrupaProfaktura
	}
	return s.stampaFakture(ctx, params, robnoStampaFakturaProfaktura, tipProfakture)
}

// GetStampaFakturaKnjizno returns the knjižna odobrenja and zaduženja of the selection, ready to print
// (the legacy PR_RPT_KNJODOBRENJE, the groups KPK and KPF; its query PR_QRY_FAKTURE is covered by the
// source query of the faktura): the stavke and the totals like the faktura, with the barkod and the
// PDV of one unit of the stavke. A document of a vrsta with the predznak "-" is a knjižno odobrenje
// (the valuta is the datum of the smanjenje, rdok.datpro, the total is ZA POVRAĆAJ and the napomena on
// the izmena of the poreska osnovica is printed), the others are knjižna zaduženja.
func (s *RobnoDokumentaResource) GetStampaFakturaKnjizno(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaKnjizno, 0)
}

// GetStampaFakturaAvansni returns the avansni računi of the selection, ready to print (the legacy
// ROB_RPT_STAMPA_ARA, the same source query as the faktura with the group of the vrsta dokumenta of the
// selection, ARA): the stavke and the totals like the faktura (the iznos of the avans is the osnovica
// and the PDV of its stopa is added), with the datum of the avansna uplata and the vezni dokument.
func (s *RobnoDokumentaResource) GetStampaFakturaAvansni(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakture(ctx, params, robnoStampaFakturaAvansni, 0)
}

// stampaFakture returns the fakture of the selection in the print vrsta (the faktura, the maloprodajni
// račun, the faktura usluga, the avansni račun, the knjižno odobrenje or the faktura-otpremnica), ready
// to print, and the izdavalac of the print (see GetStampaFaktura,
// GetStampaFakturaMP, GetStampaFakturaUsluge, GetStampaFakturaAvansni, GetStampaFakturaKnjizno and
// GetStampaFakturaOtpremnica; tipFakture is the type of the faktura-otpremnica, 0 for the others).
func (s *RobnoDokumentaResource) stampaFakture(ctx context.Context, params domain.RobnoStampaFakturaParams, vrsta robnoStampaFakturaVrsta, tipFakture int) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
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
		fakture = append(fakture, robnoStampaFakturaView(firma, magacin, stavke, avansiDok, ambalazaOf[rdokID], rateOf[rdokID], vrsta, tipFakture))
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
			coalesce(rsif.barkod, '') as barkod,
			rdok.datpro,
			coalesce(dokvrsta.predznak, '') as predznak,
			coalesce(rdok.mag, 0) as mag,
			coalesce((select m.opis from magacini m where m.god = rdok.god and m.kar = rdok.kar and m.mag = rdok.mag limit 1), '') as magopis,
			exists (select 1 from avansfakt af where af.rdokid = rdok.rdokid) as konacni,
			coalesce(rpro.cena, 0) as rcena,
			coalesce(rpro.taxcat, '') as taxcat,
			coalesce(rsif.model, '') as model,
			coalesce(rc.itaksa, 0) as itaksa,
			coalesce(rc.pakc, 0) as pakc,
			coalesce(rc.iakc, 0) as iakc,
			coalesce(rc.vma, 0) as vma,
			coalesce(jm.brdecimala, 2) as brdecimala,
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
			coalesce(rpro.dani, 0) as dani,
			coalesce(rpro.fkto, '') as stavkafkto,
			coalesce(rpro.fana, '') as stavkafana,
			coalesce(rsif.naziv, '') as rsifnaziv
		from rpro`, true)
	qb.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin("inner join dokvrsta on dokvrsta.god = rpro.god and dokvrsta.kar = rpro.kar and dokvrsta.vrd = rpro.vrd")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	// The naziv and the jedinica mere of the artikal when the stavka has none of its own (rpro.naz1 is
	// filled only for the stavke whose naziv was changed on the document).
	qb.AddJoin("left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin("left join komercijalisti kom on kom.komid = rdok.komid")
	// The cene of the artikal (the taksa, the akciza and the marža of the profaktura) and the decimals of
	// the jedinica mere of the stavka.
	qb.AddJoin(`left join lateral (select c.itaksa, c.pakc, c.iakc, c.vma from rcene c
		where c.rsifid = rpro.rsifid order by c.rceneid limit 1) rc on true`)
	qb.AddJoin(`left join lateral (select j.brdecimala from jedmere j
		where j.god = rpro.god and j.kar = rpro.kar and j.jm = coalesce(nullif(rpro.jm, ''), rsif.jm)
		order by j.jedmereid limit 1) jm on true`)
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
			coalesce(fvr.regbr, '') as regbr,
			coalesce(fvr.obv, false) as obv,
			coalesce(fvr.brobvpdv, '') as brobvpdv,
			coalesce((select string_agg(trim(b.brrac) || ' - ' || trim(b.banka), ';  ' order by b.idbanke)
				from banke b
				where b.god = fvr.god and b.kar = fvr.kar and not coalesce(b.nafakne, false)), '') as banke,
			coalesce((select string_agg(trim(b.brrac) || ';', ' ' order by b.idbanke)
				from banke b
				where b.god = fvr.god and b.kar = fvr.kar and not coalesce(b.nafakne, false)), '') as bankeracuni,
			coalesce((select string_agg(trim(b.banka) || ';', '  ' order by b.idbanke)
				from banke b
				where b.god = fvr.god and b.kar = fvr.kar and not coalesce(b.nafakne, false)), '') as bankenazivi,
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
func robnoStampaFakturaView(firma domain.RobnoStampaFakturaFirmaDto, magacin robnoStampaFakturaMagacin, stavke []domain.RobnoStampaFakturaRowDto, avansi []domain.RobnoStampaFakturaAvansDto, ambalaza []domain.RobnoStampaFakturaRowDto, rate []domain.RobnoStampaFakturaRataDto, vrsta robnoStampaFakturaVrsta, tipFakture int) domain.RobnoStampaFakturaView {
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
	odobrenje := vrsta == robnoStampaFakturaKnjizno && strings.TrimSpace(dok.Predznak) == "-"
	if vrsta == robnoStampaFakturaKnjizno {
		// "Knjižno odobrenje: 401-5" (the predznak "-") or "Knjižno zaduženje: 400-5"; the odobrenje
		// falls due on the datum of the smanjenje, pays back (ZA POVRAĆAJ) and asks for the izmena of
		// the poreska osnovica.
		lbl := i18n.GetInstance().Label
		fak.Naslov = lbl("Knjižno zaduženje")
		if odobrenje {
			fak.Naslov = lbl("Knjižno odobrenje")
			fak.ValutaNaslov = lbl("Datum smanjenja")
			fak.ZaNaplatuNaslov = lbl("Za povraćaj")
			fak.Valuta = common.FormatNullTime(dok.Datpro, common.DateLayout)
			fak.NapomenaOsnovica = true
		}
	}
	if vrsta == robnoStampaFakturaProfaktura {
		// "PROFAKTURA Br.: 900-5" or "PONUDA: 900-5", the računi and the banks of the firm apart.
		lbl := i18n.GetInstance().Label
		fak.Naslov = lbl("Profaktura Br.")
		if tipFakture == RobnoStampaTipProfakturaPonuda {
			fak.Naslov = lbl("Ponuda")
		}
		if firma.BankeRacuni != "" {
			fak.FirmaZiro = lbl("Tekući račun") + " : " + firma.BankeRacuni
		}
		fak.FirmaBankeNazivi = firma.BankeNazivi
		fak.FirmaBanka = ""
		robnoStampaProfakturaStavke(&fak, stavke, avansi)
		robnoStampaFakturaAmbalazaView(&fak, ambalaza)
		return fak
	}
	if vrsta == robnoStampaFakturaOtpremnica {
		// The broj "magacin-vrsta-broj", the Skladište and the title by the type of faktura.
		fak.BrojDokumenta = fmt.Sprintf("%d-%d-%d", dok.Mag, dok.Vrd, dok.Dokum)
		if dok.MagOpis != "" {
			fak.MagacinOpis = fmt.Sprintf("%d-%s", dok.Mag, dok.MagOpis)
		}
		fak.Naslov = robnoStampaFakturaOtpremnicaNaslov(dok, tipFakture)
	}
	if vrsta == robnoStampaFakturaZaduzenjeCO || vrsta == robnoStampaFakturaRazduzenjeCO {
		// "ZADUŽENJE ZA OBAVEZE: 210-5" ("RAZDUŽENJE ZA OBAVEZE"); the storno (the predznak "-") is paid
		// back.
		lbl := i18n.GetInstance().Label
		fak.Naslov = lbl("Zaduženje za obaveze")
		if vrsta == robnoStampaFakturaRazduzenjeCO {
			fak.Naslov = lbl("Razduženje za obaveze")
			fak.Razduzenje = true
		}
		if strings.TrimSpace(dok.Predznak) == "-" {
			fak.Naslov = lbl("Storno Račun-otpremnica")
			fak.ZaNaplatuNaslov = lbl("Za povraćaj")
		}
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
			Barkod:     st.Barkod,
			PdvPoJm:    common.FormatNumberWithSystemLocale(robnoStampaFakturaRound(cena*(100-st.Rab)/100*st.Stopa/100), 2),
		})
		if vrsta == robnoStampaFakturaZaduzenjeCO || vrsta == robnoStampaFakturaRazduzenjeCO {
			stavka := &fak.Stavke[len(fak.Stavke)-1]
			stavka.Konto, stavka.Analitika = st.StavkaFkto, st.StavkaFana
			if st.Sifra != 0 && st.RsifNaziv != "" {
				stavka.Naziv = st.RsifNaziv
			}
		}
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
	if zaNaplatu < 0 && ukPdv != 0 && vrsta != robnoStampaFakturaKnjizno {
		fak.Umanjenje = true
		fak.UmanjenjeClan = robnoStampaFakturaUmanjenjeClan
	}

	robnoStampaFakturaAvansi(&fak, avansi, zaNaplatu)

	robnoStampaFakturaAmbalazaView(&fak, ambalaza)

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

// robnoStampaFakturaAmbalazaView fills the evidentna ambalaža of a faktura: numbered in the order of
// the šifra; it is not part of the totals.
func robnoStampaFakturaAmbalazaView(fak *domain.RobnoStampaFakturaView, ambalaza []domain.RobnoStampaFakturaRowDto) {
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
}

// robnoStampaProfakturaStavke fills the stavke and the totals of a profaktura like the body of the
// legacy ROB_RPT_STAMPA_PROFAKTURA: per stavka the vrednost (količina * cena of the stavka), the rabat
// (vrednost * % rabata; for the artikli of the model "D" the rabat per unit * količina), the osnovica
// (vrednost - rabat - the taksa of the artikal * količina; for the model "D" the osnovica of the
// akciza (pakc * količina) with the akciza (iakc %) and the marža (vma * količina) less the rabat) and
// the PDV of the osnovica at the stopa of the stavka (0 for a document with a poreska napomena and a
// stavka of another tax category than S10 and S20, because of the SEF). SVEGA is the sum of the
// vrednosti, RABAT of the rabati and ZA NAPLATU SVEGA - RABAT + PDV; the PDV is summed per stopa
// (osnovica and PDV), the za naplatu po danima per dospeće (the datum of the document plus the dani of
// the stavka, or plus the rok of the document) and the avansi like the faktura. The količina has the
// decimals of its jedinica mere.
func robnoStampaProfakturaStavke(fak *domain.RobnoStampaFakturaView, stavke []domain.RobnoStampaFakturaRowDto, avansi []domain.RobnoStampaFakturaAvansDto) {
	dok := stavke[0]
	type stopaSums struct{ osnovica, pdv float64 }
	perStopa := map[float64]*stopaSums{}
	var stope []float64
	perDanu := map[time.Time]float64{}
	var ukVrednost, ukRabat, ukPdv float64
	for i, st := range stavke {
		stopa := st.Stopa
		taxcat := strings.TrimSpace(st.Taxcat)
		if strings.TrimSpace(st.Pornapomena) != "" && taxcat != "" && taxcat != "-" && taxcat != "S10" && taxcat != "S20" {
			stopa = 0
		}
		vrednost := robnoStampaFakturaRound(st.Kolic * st.Rcena)
		modelD := strings.EqualFold(strings.TrimSpace(st.Model), "D")
		var rabat, osnovica float64
		if modelD {
			rabat = robnoStampaFakturaRound(st.Rab * st.Kolic)
			akcOsnovica := robnoStampaFakturaRound(st.Pakc * st.Kolic)
			akciza := robnoStampaFakturaRound(akcOsnovica * st.Iakc / 100)
			osnovica = robnoStampaFakturaRound(akciza + akcOsnovica + st.Vma*st.Kolic - rabat)
		} else {
			rabat = robnoStampaFakturaRound(vrednost * st.Rab / 100)
			taksa := robnoStampaFakturaRound(st.Itaksa * st.Kolic)
			osnovica = robnoStampaFakturaRound(vrednost - rabat - taksa)
		}
		pdv := robnoStampaFakturaRound(osnovica * stopa / 100)
		decimale := int(st.Brdecimala)
		if decimale > 3 {
			decimale = 3
		}
		fak.Stavke = append(fak.Stavke, domain.RobnoStampaFakturaStavka{
			Rbr:        fmt.Sprintf("%d", i+1),
			Sifra:      robnoStampaFakturaSifra(st.Sifra),
			Naziv:      st.Naz1,
			Jm:         st.Jm,
			Kolicina:   common.FormatNumberWithSystemLocale(st.Kolic, decimale),
			Cena:       common.FormatNumberWithSystemLocale(st.Rcena, 2),
			ProcRabata: common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			Rabat:      common.FormatNumberWithSystemLocale(rabat, 2),
			ProcPdv:    robnoStampaFakturaStopa(stopa),
			Pdv:        common.FormatNumberWithSystemLocale(pdv, 2),
			Iznos:      common.FormatNumberWithSystemLocale(vrednost, 2),
		})
		if _, found := perStopa[stopa]; !found {
			perStopa[stopa] = &stopaSums{}
			stope = append(stope, stopa)
		}
		perStopa[stopa].osnovica += osnovica
		perStopa[stopa].pdv += pdv
		ukVrednost += vrednost
		ukRabat += rabat
		ukPdv += pdv
		if dok.Dadok.Valid {
			dani := int(dok.Rok)
			if st.Dani > 0 {
				dani = int(st.Dani)
			}
			perDanu[dok.Dadok.Time.AddDate(0, 0, dani)] += vrednost - rabat + pdv
		}
	}
	sort.Float64s(stope)
	for _, stopa := range stope {
		fak.PdvPoStopama = append(fak.PdvPoStopama, domain.RobnoStampaFakturaPdv{
			Stopa:    common.FormatNumberWithSystemLocale(stopa, 2) + "%",
			Osnovica: common.FormatNumberWithSystemLocale(perStopa[stopa].osnovica, 2),
			Pdv:      common.FormatNumberWithSystemLocale(perStopa[stopa].pdv, 2),
		})
	}
	zaNaplatu := ukVrednost - ukRabat + ukPdv
	fak.Svega = common.FormatNumberWithSystemLocale(ukVrednost, 2)
	fak.Rabat = common.FormatNumberWithSystemLocale(ukRabat, 2)
	fak.ZaNaplatu = common.FormatNumberWithSystemLocale(zaNaplatu, 2)
	// Za naplatu po danima: only when the stavke fall due on more than one day.
	if len(perDanu) > 1 {
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
	robnoStampaFakturaAvansi(fak, avansi, zaNaplatu)
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

// robnoStampaProfakturaIzvozIzjave are the izjave of the exporter of the izvozna profaktura by
// rdok.izjizv (the profaktura also has the European Community preferential origin).
var robnoStampaProfakturaIzvozIzjave = map[int64]string{
	1: "The exporter of the products covered by this document (customs autorization No. _________________) declares that, except where otherwise clearly indicated, these products are of Serbian preferential origin.",
	2: "The exporter of the products covered by this document (customs autorization No. _________________) declares that, except where otherwise clearly indicated, these products are of Serbian preferential origin - no cumulation applied.",
	3: "The exporter of the products covered by this document (customs autorization No. _________________) declares that, except where otherwise clearly indicated, these products are of European Community preferential origin.",
	4: "The exporter of the products covered by this document (customs autorization No. _________________) declares that, except where otherwise clearly indicated, these products are of European Community preferential origin - no cumulation applied.",
}

// GetStampaProfakturaIzvoz returns the izvozne profakture of the selection, ready to print (the legacy
// ROB_RPT_PROFAKTURAIZV with its query ROB_QRY_RPRO, the profakture in a foreign valuta): the data and
// the totals of the izvozna faktura (see GetStampaFakturaIzvoz) with the naziv and the komercijalni
// opis of the artikli (the legacy prints both), the discount of every stavka, the due date, the sales
// person, the status of the PDV obveznik and the four izjave of the exporter; the PIB of a foreign
// kupac (partneri.ter 3) is not printed and every stavka is printed (also the ambalaža).
func (s *RobnoDokumentaResource) GetStampaProfakturaIzvoz(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaIzvozView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaFakturaIzvoz(ctx, params, true, true)
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
	return s.stampaFakturaIzvoz(ctx, params, withKomercOpis, false)
}

// stampaFakturaIzvoz returns the izvozne fakture (proforma: the izvozne profakture) of the selection,
// ready to print (see GetStampaFakturaIzvoz and GetStampaProfakturaIzvoz).
func (s *RobnoDokumentaResource) stampaFakturaIzvoz(ctx context.Context, params domain.RobnoStampaFakturaParams, withKomercOpis, proforma bool) ([]domain.RobnoStampaFakturaIzvozView, domain.RobnoStampaFakturaFirmaDto, error) {
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
			coalesce(rdok.rok, 0) as rok,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(kom.sifkom, 0) as komsifra,
			coalesce(kom.imeprezime, '') as komnaziv,
			coalesce(valute.oznval, '') as valuta,
			coalesce(edok.brotp, 0) as brotp,
			edok.datotp,
			coalesce(p.naziv, '') as kupacnaziv,
			coalesce(p.adresa, '') as kupacadresa,
			coalesce(p.pobro, 0) as kupacpobro,
			coalesce(p.mesto, '') as kupacmesto,
			coalesce(p.tippdv, 0) as kupactippdv,
			coalesce(p.ter, 0) as kupacter,
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
	qb.AddJoin("left join komercijalisti kom on kom.komid = rdok.komid")
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
	// The izvozna faktura without the ambalaža (the profaktura prints every stavka, by rproid).
	if proforma {
		qb.AddOrderBy("rdok.rdokid, rpro.rproid")
	} else {
		qb.AddCustomCondition("coalesce(rpro.sifra, 0) < 990000")
		qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	}
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
		fakture = append(fakture, robnoStampaFakturaIzvozView(firma, magacin, (*rows)[i:j], withKomercOpis, proforma))
		i = j
	}
	return fakture, firma, nil
}

// robnoStampaFakturaIzvozView builds the printed izvozna faktura (proforma: profaktura) of the stavke
// of one robni dokument (see GetStampaFakturaIzvoz and GetStampaProfakturaIzvoz).
func robnoStampaFakturaIzvozView(firma domain.RobnoStampaFakturaFirmaDto, magacin robnoStampaFakturaMagacin, rows []domain.RobnoStampaFakturaIzvozRowDto, withKomercOpis, proforma bool) domain.RobnoStampaFakturaIzvozView {
	dok := rows[0]
	fak := domain.RobnoStampaFakturaIzvozView{
		KupacNaziv:    dok.KupacNaziv,
		KupacAdresa:   dok.KupacAdresa,
		KupacMesto:    robnoStampaFakturaIzvozMesto(dok.KupacPobro, dok.KupacMesto),
		KupacPib:      robnoStampaFakturaIzvozPib(dok, proforma),
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
	if proforma {
		robnoStampaProfakturaIzvozView(&fak, firma, dok)
	}

	// The količina with 3 decimals (the profaktura with 2).
	kolicinaDecimale := 3
	if proforma {
		kolicinaDecimale = 2
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
			Kolicina: common.FormatNumberWithSystemLocale(st.Kolic, kolicinaDecimale),
			Cena:     common.FormatNumberWithSystemLocale(st.Cenaval, 4),
			Rabat:    common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			Iznos:    common.FormatNumberWithSystemLocale(iznos, 2),
			Popust:   common.FormatNumberWithSystemLocale(popust, 2),
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

// robnoStampaProfakturaIzvozView fills what the izvozna profaktura prints differently from the izvozna
// faktura: the status of the PDV obveznik, the telefon as it is, the tekući račun (rdok.posuslnab),
// the delivery place, the despatch note (rdok.dokiz), the due date, the sales person and the izjava
// of the exporter.
func robnoStampaProfakturaIzvozView(fak *domain.RobnoStampaFakturaIzvozView, firma domain.RobnoStampaFakturaFirmaDto, dok domain.RobnoStampaFakturaIzvozRowDto) {
	if firma.Obv {
		fak.FirmaObveznik = "Obv. PDV-a br: " + firma.Brobvpdv
	} else {
		fak.FirmaObveznik = "Nije obveznik PDV-a"
	}
	fak.FirmaTel = "Telefon/faks :" + firma.Tel
	// The tekući račun of the document ("", "-" and "-1" are none, like on the faktura).
	fak.FirmaZiro = ""
	if dok.Posuslnab != "" && dok.Posuslnab != "-" && dok.Posuslnab != "-1" {
		fak.FirmaZiro = "Tekući račun : " + dok.Posuslnab
	}
	if dok.MispNaziv != "" {
		fak.Destinacija = "Delivery. Place:" + dok.MispNaziv
	}
	fak.Otpremnica = dok.Dokiz
	fak.DatumDospeca = common.AddDaysToNullTime(dok.Dadok, int(dok.Rok), common.DateLayout)
	if dok.KomSifra != 0 {
		fak.Komercijalista = strings.TrimSpace(fmt.Sprintf("%d %s", dok.KomSifra, dok.KomNaziv))
	}
	fak.Izjava = robnoStampaProfakturaIzvozIzjave[dok.Izjizv]
}

// robnoStampaFakturaIzvozMesto returns the poštanski broj and the mesto of an adresa (without the
// poštanski broj when it is not set, e.g. a foreign partner).
func robnoStampaFakturaIzvozMesto(pobro int64, mesto string) string {
	if pobro == 0 {
		return strings.TrimSpace(mesto)
	}
	return strings.TrimSpace(fmt.Sprintf("%d %s", pobro, mesto))
}

// robnoStampaFakturaIzvozPib returns the PIB line of the kupac of the izvozna faktura (see
// robnoStampaPartnerPib); the profaktura prints no PIB for a foreign kupac (partneri.ter 3) and a space
// after the colons ("PIB : ").
func robnoStampaFakturaIzvozPib(dok domain.RobnoStampaFakturaIzvozRowDto, proforma bool) string {
	sep := ":"
	if proforma {
		if dok.KupacTer == 3 {
			return ""
		}
		sep = ": "
	}
	return robnoStampaPartnerPib(robnoStampaPartner{
		TipPdv: dok.KupacTipPdv, Pib: dok.KupacPib, Budzetski: dok.KupacBudzetski, Jbkjs: dok.KupacJbkjs,
		Jmbg: dok.KupacJmbg, Bpg: dok.KupacBpg, Index: dok.KupacIndex,
	}, sep)
}

// robnoStampaPrenosnicaStavka returns the printed stavka of a prenosnica (see GetStampaPrenosnica): the
// naziv of the artikal (the naziv of the stavka when the artikal is not found) with, in a magacin with
// serije, the serija and the rok trajanja on a new line (when the stavka has them), and with the
// komercijalni opis, the proizvođač and its šifra when the option is set; the jedinica mere of the
// artikal, the količina with its decimals, the cena and the vrednost (rpro.iznos).
func robnoStampaPrenosnicaStavka(r domain.RobnoStampaPopisRowDto, opcije RobnoStampaPrenosnicaOpcije) domain.RobnoStampaPopisStavka {
	naziv, jm := r.Naz1, r.Jm
	if r.RsifNaziv != "" {
		naziv, jm = r.RsifNaziv, r.RsifJm
		if r.Tipzal != 1 && (r.Serija != "" || r.Roktr != 0) {
			naziv += "\nSerija: " + r.Serija + " Rok trajanja: " + robnoStampaWinDevDatum(r.Roktr)
		}
		if opcije.KomercOpis {
			naziv = strings.TrimSpace(strings.Join([]string{naziv, r.Komercopis, r.Pro, r.Proizsifra}, " "))
		}
	}
	decimale := int(r.Brdecimala)
	if decimale > 3 {
		decimale = 3
	}
	return domain.RobnoStampaPopisStavka{
		Rbr:      fmt.Sprintf("%d", r.Rbr),
		Konto:    r.Konto,
		Sifra:    robnoStampaFakturaSifra(r.Sifra),
		Naziv:    naziv,
		Jm:       jm,
		Kolicina: common.FormatNumberWithSystemLocale(r.Kolic, decimale),
		Cena:     common.FormatNumberWithSystemLocale(r.Cena, 2),
		Iznos:    common.FormatNumberWithSystemLocale(r.Iznos, 2),
	}
}

// robnoStampaPrenosnicaUlazVrednost returns the vrednost of a stavka of the prenosnica ulaz: the
// količina times the cena, rounded to the para (the legacy ITEM_IZNOS = ITEM_KOLIC * ITEM_CENA).
func robnoStampaPrenosnicaUlazVrednost(r domain.RobnoStampaPopisRowDto) float64 {
	return robnoStampaFakturaRound(r.Kolic * r.Cena)
}

// robnoStampaNivelacijaStavka returns the printed stavka of a nivelacija cena (see GetStampaNivelacija)
// with its stara and nova vrednost: the naziv of the artikal (in a magacin with serije, rdok's magacin
// with tipzal not 1, with the otk, the serija and the rok trajanja), the stara cena (rpro.fcena) and
// vrednost, the procenat (rpro.vma), the nova cena (rpro.cena) and vrednost and the vrednost of the
// nivelacija (the legacy ITEM_STVRED, ITEM_NOVRED and ITEM_RAZL).
func robnoStampaNivelacijaStavka(r domain.RobnoStampaPopisRowDto) (domain.RobnoStampaPopisStavka, float64, float64) {
	naziv, jm := r.Naz1, r.Jm
	if r.RsifNaziv != "" {
		naziv, jm = r.RsifNaziv, r.RsifJm
	}
	if r.Tipzal != 1 {
		naziv = strings.TrimSpace(naziv + " " + r.Otk + " " + r.Serija)
		if r.Roktr != 0 {
			naziv += " " + robnoStampaWinDevDatum(r.Roktr)
		}
	}
	stara := robnoStampaFakturaRound(r.Kolic * r.Fcena)
	nova := robnoStampaFakturaRound(r.Kolic * r.Cena)
	return domain.RobnoStampaPopisStavka{
		Rbr:           fmt.Sprintf("%d", r.Rbr),
		Sifra:         robnoStampaFakturaSifra(r.Sifra),
		Naziv:         naziv,
		Jm:            jm,
		Kolicina:      common.FormatNumberWithSystemLocale(r.Kolic, 3),
		StaraCena:     common.FormatNumberWithSystemLocale(r.Fcena, 2),
		StaraVrednost: common.FormatNumberWithSystemLocale(stara, 2),
		Procenat:      common.FormatNumberWithSystemLocale(r.Vma, 2) + "%",
		Cena:          common.FormatNumberWithSystemLocale(r.Cena, 2),
		Iznos:         common.FormatNumberWithSystemLocale(nova, 2),
		Razlika:       common.FormatNumberWithSystemLocale(nova-stara, 2),
	}, stara, nova
}

// robnoStampaZaduzenjeSIStavka returns the printed stavka rbr of a zaduženje sitnog inventara (see
// GetStampaZaduzenjeSI) or of a zaduženje gradilišta (gradiliste): the naziv and the jedinica mere of the
// artikal (of the stavka when the artikal is not found), the količina, the cena and the vrednost. On the
// zaduženje gradilišta in a magacin with the tip zaliha 2 the naziv has a second line with the otk, the
// serija and the rok trajanja (when the stavka has one).
func robnoStampaZaduzenjeSIStavka(r domain.RobnoStampaPopisRowDto, rbr int, gradiliste bool) domain.RobnoStampaPopisStavka {
	naziv, jm := r.Naz1, r.Jm
	if r.RsifNaziv != "" {
		naziv, jm = r.RsifNaziv, r.RsifJm
	}
	if gradiliste && r.Tipzal == 2 {
		naziv += "\n" + r.Otk + "-" + r.Serija
		if r.Roktr != 0 {
			naziv += "-" + robnoStampaWinDevDatum(r.Roktr)
		}
	}
	return domain.RobnoStampaPopisStavka{
		Rbr:      fmt.Sprintf("%d", rbr),
		Konto:    r.Konto,
		Sifra:    robnoStampaFakturaSifra(r.Sifra),
		Naziv:    naziv,
		Jm:       jm,
		Kolicina: common.FormatNumberWithSystemLocale(r.Kolic, 3),
		Cena:     common.FormatNumberWithSystemLocale(r.Cena, 2),
		Iznos:    common.FormatNumberWithSystemLocale(r.Iznos, 2),
	}
}

// robnoStampaPopisTekGodStavka returns the printed stavka of a popis tekuće godine (see
// GetStampaPopisTekGod): the naziv of the stavka (of the artikal when the stavka has none), the otk, the
// serija, the rok trajanja, the količina with 2 decimals, the cena with 3 and the iznos.
func robnoStampaPopisTekGodStavka(r domain.RobnoStampaPopisRowDto) domain.RobnoStampaPopisStavka {
	return domain.RobnoStampaPopisStavka{
		Rbr:      fmt.Sprintf("%d", r.Rbr),
		Konto:    r.Konto,
		Sifra:    robnoStampaFakturaSifra(r.Sifra),
		Naziv:    r.Naz1,
		Jm:       r.Jm,
		Otk:      r.Otk,
		Serija:   r.Serija,
		Rok:      robnoStampaWinDevDatumFormat(r.Roktr, "02/01/2006"),
		Kolicina: common.FormatNumberWithSystemLocale(r.Kolic, 2),
		Cena:     common.FormatNumberWithSystemLocale(r.Cena, 3),
		Iznos:    common.FormatNumberWithSystemLocale(r.Iznos, 2),
	}
}

// robnoStampaWinDevDatum returns a WinDev integer date (the number of days since 01.01.1800, like
// IntegerToDate) as a date ("" for 0).
func robnoStampaWinDevDatum(days int64) string {
	return robnoStampaWinDevDatumFormat(days, common.DateLayout)
}

// robnoStampaWinDevDatumFormat returns a WinDev integer date (see robnoStampaWinDevDatum) in the layout.
func robnoStampaWinDevDatumFormat(days int64, layout string) string {
	if days == 0 {
		return ""
	}
	return time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(days)).Format(layout)
}

// robnoStampaDokumentaUslovi adds to a query of the stavke (rpro) of the robni dokumenti (rdok) the
// selection common to the prints of every vrsta dokumenta: the period of the session, the vrsta naloga,
// the ranges of the broj naloga, of the broj dokumenta and of the vrsta dokumenta, the magacin, the
// selected document (RdokID), the vrsta dokumenta (Vrd) and the range of the datum naloga.
func robnoStampaDokumentaUslovi(qb *common.QueryBuilder, userSession *domain.UserSession, params domain.RobnoStampaFakturaParams) {
	robnoStampaDokumentaUsloviTabele(qb, userSession, params, "rpro")
}

// robnoStampaDokumentaUsloviTabele adds the common selection of the prints (see
// robnoStampaDokumentaUslovi) on the columns of the table t: rpro for the prints of the stavke, rdok for
// the prints of the documents without stavke.
func robnoStampaDokumentaUsloviTabele(qb *common.QueryBuilder, userSession *domain.UserSession, params domain.RobnoStampaFakturaParams, t string) {
	qb.AddEqual(t+".god", userSession.SelectedGod)
	qb.AddEqual(t+".kar", userSession.SelectedKar)
	qb.AddEqual("rdok.tipdok", params.Tipdok)
	addNumberCondition(qb, t+".nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, t+".nalog", params.DoNaloga, "<=")
	addNumberCondition(qb, t+".dokum", params.OdDokum, ">=")
	addNumberCondition(qb, t+".dokum", params.DoDokum, "<=")
	addNumberCondition(qb, t+".vrd", params.OdVrd, ">=")
	addNumberCondition(qb, t+".vrd", params.DoVrd, "<=")
	if params.MagaciniID != 0 {
		qb.AddEqual(t+".magaciniid", params.MagaciniID)
	}
	if params.RdokID != 0 {
		qb.AddEqual("rdok.rdokid", params.RdokID)
	}
	addNumberCondition(qb, t+".vrd", params.Vrd, "=")
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
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rbr", robnoStampaOpstiDokumentOpcije{})
}

// GetStampaPopisTekGod returns the popisi tekuće godine of the selection (the group PTG of the vrste
// dokumenta, the legacy RPT_POPIS_TEKGOD with its query QRY_RPRO_ZADOK), ready to print: like the popis,
// with the otk, the serija and the rok trajanja of every stavka (the količina with 2 decimals, the cena
// with 3) and the total of the iznos.
func (s *RobnoDokumentaResource) GetStampaPopisTekGod(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rbr", robnoStampaOpstiDokumentOpcije{popisTekGod: true})
}

// GetStampaZaduzenjeSI returns the zaduženja sitnog inventara of the selection (the group SIV of the
// vrste dokumenta, the legacy RPT_ROB_ZADUZ_SI with the query of the opšti dokumenti ROB_QRY_OPDSTAMPA),
// ready to print: the broj of the document is nalog-vrsta-broj, the header has the konto and the šifra
// of the radnik (rdok.pkto and rdok.pana with the nazivi of their sintetički and analitički konto); the
// stavke are numbered per document in the order they were entered (rpro.rproid), with the naziv and the
// jedinica mere of the artikal, the količina, the cena and the vrednost (rpro.iznos), and their totals.
func (s *RobnoDokumentaResource) GetStampaZaduzenjeSI(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rproid", robnoStampaOpstiDokumentOpcije{zaduzenjeSI: true})
}

// GetStampaZaduzenjeGradilista returns the zaduženja gradilišta of the selection (the group GRD of the
// vrste dokumenta, the legacy RPT_ROB_ZADGRADILISTA with the query of the opšti dokumenti
// ROB_QRY_OPDSTAMPA), ready to print: like the zaduženje sitnog inventara (see GetStampaZaduzenjeSI),
// with the gradilište (rdok.pkto rdok.pana and its naziv) in the header, the vrednost of a stavka the
// količina times the cena and, in a magacin with the tip zaliha 2, the otk, the serija and the rok
// trajanja of the stavka under its naziv.
func (s *RobnoDokumentaResource) GetStampaZaduzenjeGradilista(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rproid", robnoStampaOpstiDokumentOpcije{zaduzenjeSI: true, gradiliste: true})
}

// GetStampaNivelacija returns the nivelacije cena of the selection (the group NIV of the vrste
// dokumenta, the legacy RPT_ROB_NIVELACIJA; its query QRY_ROB_NIV is covered by the query of the opšti
// dokumenti), ready to print, in the order of the nalozi, of the documents and of the stavke as entered:
// per stavka the stara cena (rpro.fcena) and vrednost, the procenat (rpro.vma), the nova cena
// (rpro.cena) and vrednost and the vrednost of the nivelacija (količina * (nova - stara cena)), with
// their totals per document; in a magacin with serije the naziv of the artikal has the otk, the serija
// and the rok trajanja. When the selection has more nalozi, every document shows its nalog and the last
// document of a nalog the total of the nivelacija of the nalog (the legacy BREAK1, printed for a range of
// nalozi).
func (s *RobnoDokumentaResource) GetStampaNivelacija(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	dokumenti, firma, err := s.stampaOpstiDokumenti(ctx, params, "rdok.rnalid, rdok.rdokid, rpro.rproid", robnoStampaOpstiDokumentOpcije{nivelacija: true})
	if err != nil {
		return nil, firma, err
	}
	robnoStampaNivelacijaNalozi(dokumenti)
	return dokumenti, firma, nil
}

// robnoStampaNivelacijaNalozi fills the nalog and the total of the nivelacija per nalog of the printed
// nivelacije when they belong to more nalozi (see GetStampaNivelacija).
func robnoStampaNivelacijaNalozi(dokumenti []domain.RobnoStampaPopisView) {
	nalozi := map[string]bool{}
	for _, dok := range dokumenti {
		nalozi[dok.Nalog] = true
	}
	if len(nalozi) < 2 {
		return
	}
	var ukupno float64
	for i := range dokumenti {
		dokumenti[i].NalogNivelacije = dokumenti[i].Nalog
		ukupno += dokumenti[i].RazlikaIznos
		if i == len(dokumenti)-1 || dokumenti[i+1].Nalog != dokumenti[i].Nalog {
			dokumenti[i].UkupnoZaNalog = common.FormatNumberWithSystemLocale(ukupno, 2)
			ukupno = 0
		}
	}
}

// robnoStampaOpstiDokumentOpcije are the prints of stampaOpstiDokumenti beside the popis and the opšti
// dokument: the prenosnica (with its options), the popis tekuće godine, the zaduženje sitnog inventara
// (with gradiliste: the zaduženje gradilišta) and the nivelacija cena.
type robnoStampaOpstiDokumentOpcije struct {
	prenosnica  *RobnoStampaPrenosnicaOpcije
	popisTekGod bool
	zaduzenjeSI bool
	gradiliste  bool
	nivelacija  bool
}

// RobnoStampaPrenosnicaOpcije are the options of the print of the prenosnica (the legacy ipCBOX_NAZIV
// and ipCBOX_KOMERCOPIS): the naziv of the artikal and its komercijalni opis with the proizvođač and
// its šifra.
type RobnoStampaPrenosnicaOpcije struct {
	Naziv, KomercOpis bool
	// Ulaz is the prenosnica ulaz (the legacy RPT_ROB_PRENOSNICEULAZ): the vrednost of a stavka is the
	// količina times the cena (the izlaz prints rpro.iznos).
	Ulaz bool
}

// GetStampaPrenosnica returns the prenosnice of the selection (the group PRI of the vrste dokumenta, the
// legacy RPT_ROB_PRENOSNICE, and with opcije.Ulaz the group PRU, RPT_ROB_PRENOSNICEULAZ, both with the
// query of the opšti dokumenti ROB_QRY_OPDSTAMPA), ready to print:
// one view per robni dokument with its stavke in the order they were entered (rpro.rproid), the
// magacin the goods come from and the one they go to, the totals of the vrednost (rpro.iznos) and the
// izdavalac (fvr) of the print. The naziv of a stavka is the one of the artikal (with the serija and
// the rok trajanja in a magacin with serije, rdok's magacin with tipzal not 1, and with the komercijalni
// opis by the options); the jedinica mere is the one of the artikal and the količina has its decimals.
func (s *RobnoDokumentaResource) GetStampaPrenosnica(ctx context.Context, params domain.RobnoStampaFakturaParams, opcije RobnoStampaPrenosnicaOpcije) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rproid", robnoStampaOpstiDokumentOpcije{prenosnica: &opcije})
}

// GetStampaOpstiDokument returns the opšti dokumenti of the selection (the groups OPD, POT, KOL and FIN
// of the vrste dokumenta, the legacy RPT_OPDSTAMPA with its query ROB_QRY_OPDSTAMPA), ready to print:
// one view per robni dokument with its stavke in the order they were entered (rpro.rproid), the totals
// of the nabavna vrednost, of the iznos and of the količina, and the izdavalac (fvr) of the print. The
// template prints the stavke by the group of the document (Grupa).
func (s *RobnoDokumentaResource) GetStampaOpstiDokument(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaOpstiDokumenti(ctx, params, "rdok.rdokid, rpro.rproid", robnoStampaOpstiDokumentOpcije{})
}

// stampaOpstiDokumenti returns the popisi, the opšti dokumenti, the prenosnice, the popisi tekuće godine,
// the zaduženja sitnog inventara, the zaduženja gradilišta or the nivelacije cena (by opcije) of the
// selection with their stavke in the order orderBy (see GetStampaPopis, GetStampaOpstiDokument,
// GetStampaPrenosnica, GetStampaPopisTekGod, GetStampaZaduzenjeSI, GetStampaZaduzenjeGradilista and
// GetStampaNivelacija).
func (s *RobnoDokumentaResource) stampaOpstiDokumenti(ctx context.Context, params domain.RobnoStampaFakturaParams, orderBy string, opcije robnoStampaOpstiDokumentOpcije) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error) {
	prenosnica := opcije.prenosnica
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
			coalesce(mg1.mag || '-' || mg1.opis, '') as magprijem,
			coalesce(mg.tipzal, 0) as tipzal,
			coalesce(rpro.serija, '') as serija,
			coalesce(rpro.otk, '') as otk,
			coalesce(rdok.nalog, 0) as nalog1,
			coalesce(nullif(rdok.pkto, '') || '-' || ktr.naziv, '') as kontoradnik,
			coalesce(nullif(rdok.pana, '') || '-' || rad.naziv, '') as sifraradnik,
			trim(coalesce(rdok.pkto, '') || ' ' || coalesce(rdok.pana, '')) as gradiliste,
			coalesce(grd.naziv, '') as gradilistenaziv,
			coalesce(rdok.rnalid, 0) as rnalid,
			coalesce(rpro.fcena, 0) as fcena,
			coalesce(rpro.vma, 0) as vma,
			coalesce(rdok.opis, '') as opis,
			coalesce(rpro.roktr, 0)::bigint as roktr,
			coalesce(rsif.naziv, '') as rsifnaziv,
			coalesce(rsif.komercopis, '') as komercopis,
			coalesce(rsif.pro, '') as pro,
			coalesce(rsif.proizsifra, '') as proizsifra,
			coalesce(rsif.jm, '') as rsifjm,
			coalesce(jm.brdecimala, 2) as brdecimala,
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
	// The radnik of the zaduženje sitnog inventara: the sintetički konto (vkonta 2) of rdok.pkto and the
	// analitički konto (vkonta 1) of rdok.pkto and rdok.pana.
	qb.AddJoin(` left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.pkto and f.vkonta = 2
		order by f.idfkpl limit 1) ktr on true`)
	qb.AddJoin(` left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.vkonta = 1 and f.konto = rdok.pkto and f.sifra = rdok.pana
		order by f.idfkpl limit 1) rad on true`)
	// The gradilište of the zaduženje gradilišta: the fkpl of rdok.pkto and rdok.pana.
	qb.AddJoin(` left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.pkto and f.sifra = rdok.pana
		order by f.vkonta desc, f.idfkpl limit 1) grd on true`)
	// The magacin the goods go to (the prenosnica) and the decimals of the jedinica mere of the artikal.
	qb.AddJoin(" left join magacini mg1 on mg1.god = rdok.god and mg1.kar = rdok.kar and mg1.mag = rdok.magid1 and coalesce(rdok.magid1, 0) <> 0")
	qb.AddJoin(` left join lateral (select j.brdecimala from jedmere j
		where j.god = rpro.god and j.kar = rpro.kar and j.jm = coalesce(nullif(rsif.jm, ''), rpro.jm)
		order by j.jedmereid limit 1) jm on true`)
	robnoStampaDokumentaUslovi(qb, userSession, params)
	qb.AddOrderBy(orderBy)
	sqlQuery, args := qb.Build()
	rows, err := s.stampaPopisRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}

	popisi := []domain.RobnoStampaPopisView{}
	var nabavna, iznos, kolicina, cena, staraVrednost, novaVrednost, razlika float64
	zatvori := func() {
		if len(popisi) == 0 {
			return
		}
		last := &popisi[len(popisi)-1]
		last.UkupnoNabavnaVrednost = common.FormatNumberWithSystemLocale(nabavna, 2)
		last.UkupnoIznos = common.FormatNumberWithSystemLocale(iznos, 2)
		last.UkupnoKolicina = common.FormatNumberWithSystemLocale(kolicina, 3)
		last.UkupnoCena = common.FormatNumberWithSystemLocale(cena, 2)
		last.UkupnoStaraVrednost = common.FormatNumberWithSystemLocale(staraVrednost, 2)
		last.UkupnoNovaVrednost = common.FormatNumberWithSystemLocale(novaVrednost, 2)
		last.UkupnoRazlika = common.FormatNumberWithSystemLocale(razlika, 2)
		last.RazlikaIznos = razlika
	}
	for i, r := range *rows {
		if i == 0 || (*rows)[i-1].RdokID != r.RdokID {
			zatvori()
			nabavna, iznos, kolicina, cena, staraVrednost, novaVrednost, razlika = 0, 0, 0, 0, 0, 0, 0
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
				MagacinPrijem:     r.MagPrijem,
				KontoRadnika:      r.KontoRadnik,
				SifraRadnika:      r.SifraRadnik,
				Gradiliste:        r.Gradiliste,
				GradilisteNaziv:   r.GradilisteNaziv,
				Napomena:          r.Opis,
			})
			if opcije.zaduzenjeSI {
				popisi[len(popisi)-1].BrojDokumenta = fmt.Sprintf("%d-%d-%d", r.Nalog1, r.Vrd, r.Dokum)
			}
		}
		nabavnaVrednost := robnoStampaFakturaRound(r.Kolic * r.Ncena)
		nabavna += nabavnaVrednost
		// The prenosnica ulaz and the zaduženje gradilišta: the vrednost is the količina times the cena.
		if (prenosnica != nil && prenosnica.Ulaz) || opcije.gradiliste {
			r.Iznos = robnoStampaPrenosnicaUlazVrednost(r)
		}
		iznos += r.Iznos
		kolicina += r.Kolic
		cena += r.Cena
		pop := &popisi[len(popisi)-1]
		if prenosnica != nil {
			pop.Stavke = append(pop.Stavke, robnoStampaPrenosnicaStavka(r, *prenosnica))
			continue
		}
		if opcije.zaduzenjeSI {
			pop.Stavke = append(pop.Stavke, robnoStampaZaduzenjeSIStavka(r, len(pop.Stavke)+1, opcije.gradiliste))
			continue
		}
		if opcije.nivelacija {
			stavka, stara, nova := robnoStampaNivelacijaStavka(r)
			staraVrednost += stara
			novaVrednost += nova
			razlika += nova - stara
			pop.Stavke = append(pop.Stavke, stavka)
			continue
		}
		if opcije.popisTekGod {
			pop.Stavke = append(pop.Stavke, robnoStampaPopisTekGodStavka(r))
			continue
		}
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
	return s.stampaKalkulacije(ctx, params, false)
}

// GetStampaKalkulacijaMaloprodaje returns the kalkulacije maloprodaje of the selection (the group KAL,
// the legacy RPT_ROB_KALKULACIJA_MP with its query ROB_QRY_RPRO), ready to print: like the kalkulacija
// veleprodaje (see GetStampaKalkulacija) with the stavke in the order they were entered (rpro.rproid),
// the prodavnica (rdok.pkto and rdok.pana) instead of the magacin, the data of the firm in the header and
// the stavke at the maloprodajne cene (see robnoStampaKalkulacijaMPStavka); the porezi of the document of
// the dobavljač are the ones entered with it (rppo) per poreska oznaka, the porezi of the kalkulacija the
// osnovica and the ukalkulisani PDV of the stavke per poreska oznaka.
func (s *RobnoDokumentaResource) GetStampaKalkulacijaMaloprodaje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaKalkulacijaView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaKalkulacije(ctx, params, true)
}

// stampaKalkulacije returns the kalkulacije veleprodaje or (maloprodaja) maloprodaje of the selection
// (see GetStampaKalkulacija and GetStampaKalkulacijaMaloprodaje).
func (s *RobnoDokumentaResource) stampaKalkulacije(ctx context.Context, params domain.RobnoStampaFakturaParams, maloprodaja bool) ([]domain.RobnoStampaKalkulacijaView, domain.RobnoStampaFakturaFirmaDto, error) {
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
			coalesce(tar.pp, rpro.pdvpct, 0) as stopa,
			coalesce(rpro.mcenap, 0) as mcenap,
			coalesce(rdok.pkto, '') as pkto,
			coalesce(rdok.pana, '') as pana,
			coalesce(prod.naziv, '') as prodavnicanaziv,
			prod.naziv is not null as prodavnicanadjena
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
	// The prodavnica of the kalkulacija maloprodaje: the fkpl of rdok.pkto and rdok.pana.
	qb.AddJoin(` left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.pkto and f.sifra = rdok.pana
		order by f.vkonta desc, f.idfkpl limit 1) prod on true`)
	robnoStampaDokumentaUslovi(qb, userSession, params)
	if maloprodaja {
		qb.AddOrderBy("rdok.rdokid, rpro.rproid")
	} else {
		qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	}
	sqlQuery, args := qb.Build()
	rows, err := s.stampaKalkulacijaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	// The porezi of the documents of the dobavljači (rppo) per document and poreska oznaka.
	rppoOf := map[int64][]domain.RobnoStampaKalkulacijaRowDto{}
	if maloprodaja && len(*rows) > 0 {
		var rdokIDs []any
		for i, r := range *rows {
			if i == 0 || (*rows)[i-1].RdokID != r.RdokID {
				rdokIDs = append(rdokIDs, r.RdokID)
			}
		}
		pq := common.NewQueryBuilder(`select
				rppo.rdokid,
				coalesce(rppo.po, 0) as po,
				coalesce(sum(rppo.osn), 0) as osn,
				coalesce(sum(rppo.ppor), 0) as ppor,
				coalesce(min(rppo.pdv), 0) as pdv
			from rppo`, true)
		pq.AddIn("rppo.rdokid", rdokIDs)
		pq.AddGroupBy("rppo.rdokid, rppo.po")
		pq.AddOrderBy("rppo.rdokid, rppo.po")
		pSQL, pArgs := pq.Build()
		porezi, err := s.stampaKalkulacijaRepo.GetAllCustom(ctx, pSQL, "", pArgs, "", "")
		if err != nil {
			return nil, firma, err
		}
		for _, pz := range *porezi {
			rppoOf[pz.RdokID] = append(rppoOf[pz.RdokID], pz)
		}
	}

	kalkulacije := []domain.RobnoStampaKalkulacijaView{}
	for i := 0; i < len(*rows); {
		j := i
		for j < len(*rows) && (*rows)[j].RdokID == (*rows)[i].RdokID {
			j++
		}
		if maloprodaja {
			kalkulacije = append(kalkulacije, robnoStampaKalkulacijaMPView(firma, (*rows)[i:j], rppoOf[(*rows)[i].RdokID]))
		} else {
			kalkulacije = append(kalkulacije, robnoStampaKalkulacijaView((*rows)[i:j]))
		}
		i = j
	}
	return kalkulacije, firma, nil
}

// robnoStampaKalkulacijaMPView builds the printed kalkulacija maloprodaje of the stavke of one robni
// dokument and of the porezi of the document of its dobavljač (rppo, see
// GetStampaKalkulacijaMaloprodaje). Per stavka the fakturna, the rabat, the neto fakturna vrednost, the
// zavisni troškovi and the nabavna vrednost like the kalkulacija veleprodaje (the zavisni troškovi
// together); the maloprodajna vrednost is the količina times the maloprodajna cena with the PDV
// (rpro.mcenap), the PDV the part of the stopa in it (vrednost * stopa / (100 + stopa)), the vrednost
// marže the maloprodajna vrednost less the PDV and the nabavna vrednost and the % MP marže the vrednost
// marže of the nabavna vrednost. Every vrednost is rounded to 2 decimals.
func robnoStampaKalkulacijaMPView(firma domain.RobnoStampaFakturaFirmaDto, rows []domain.RobnoStampaKalkulacijaRowDto, rppo []domain.RobnoStampaKalkulacijaRowDto) domain.RobnoStampaKalkulacijaView {
	lbl := i18n.GetInstance().Label
	f2 := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 2) }
	r2 := robnoStampaFakturaRound
	r := rows[0]
	kal := robnoStampaKalkulacijaView(rows)
	kal.Maloprodaja = true
	kal.Porezi, kal.UkupnoPorezi = nil, domain.RobnoStampaKalkulacijaPorez{}
	kal.FirmaMesto = firma.Adresa + "," + firma.Mesto
	kal.FirmaPib = lbl("PIB") + ": " + firma.Pib + ", " + lbl("Šifra delatnosti") + ": " + firma.Sifdel + ", " + lbl("Matični broj") + ": " + firma.Matbr
	kal.FirmaTel = lbl("Telefon") + ":" + firma.Tel
	kal.VremeStampe = time.Now().Format("15:04:05")
	kal.Prodavnica = lbl("Prodavnica ne postoji u kontnom planu!")
	if r.ProdavnicaNadjena {
		kal.Prodavnica = r.Pkto + "-" + r.Pana + " " + r.ProdavnicaNaziv
	}

	type porezKalk struct {
		po              int64
		stopa           float64
		osnovica, porez float64
	}
	poreziKalk := map[int64]*porezKalk{}
	var redKalk []int64
	var ukTrosak, ukMarza, ukPdv, ukMp float64
	for i, st := range rows {
		var v robnoStampaKalkulacijaVrednosti
		v.fakturna = r2(st.Kolic * st.Fcena)
		v.rabat = r2(v.fakturna * st.Rab / 100)
		v.neto = r2(v.fakturna - v.rabat)
		v.trosak = robnoStampaKalkulacijaTrosak(st.Tz, st.Ztro, v.neto)
		v.trosakDobavljaca = r2(st.Kolic * st.Ztrof)
		v.interni = robnoStampaKalkulacijaTrosak(st.Tzi, st.Ztrin, v.neto)
		v.nabavna = r2(v.neto + v.trosak + v.trosakDobavljaca + v.interni)
		trosak := r2(v.trosak + v.trosakDobavljaca + v.interni)
		mp := r2(st.Kolic * st.Mcenap)
		pdv := r2(mp * st.Stopa / (100 + st.Stopa))
		marza := r2(mp - pdv - v.nabavna)
		marzaPct := 0.0
		if v.nabavna != 0 {
			marzaPct = marza / v.nabavna * 100
		}
		stavka := &kal.Stavke[i]
		stavka.Marza = f2(marzaPct) + "%"
		stavka.VrednostTroska = f2(trosak)
		stavka.VrednostMarze = f2(marza)
		stavka.StopaPdv = f2(st.Stopa) + "%"
		stavka.MpCena = f2(st.Mcenap)
		stavka.VrednostPdv = f2(pdv)
		stavka.MpVrednost = f2(mp)
		ukTrosak += trosak
		ukMarza += marza
		ukPdv += pdv
		ukMp += mp

		pz := poreziKalk[st.Po]
		if pz == nil {
			pz = &porezKalk{po: st.Po, stopa: st.Stopa}
			poreziKalk[st.Po] = pz
			redKalk = append(redKalk, st.Po)
		}
		pz.osnovica += mp - pdv
		pz.porez += pdv
	}
	kal.Ukupno.VrednostTroska = f2(ukTrosak)
	kal.Ukupno.VrednostMarze = f2(ukMarza)
	kal.Ukupno.VrednostPdv = f2(ukPdv)
	kal.Ukupno.MpVrednost = f2(ukMp)

	// Po dokumentu dobavljača: the porezi entered with the document (rppo) per poreska oznaka.
	var iznos, osnovica, porez float64
	for _, pz := range rppo {
		kal.Porezi = append(kal.Porezi, domain.RobnoStampaKalkulacijaPorez{
			IznosDokumenta: f2(pz.Osn + pz.Ppor),
			Osnovica:       f2(pz.Osn),
			Porez:          f2(pz.Ppor),
			Tarifa:         fmt.Sprintf("%d", pz.Po),
			Stopa:          f2(pz.Pdv),
		})
		iznos += pz.Osn + pz.Ppor
		osnovica += pz.Osn
		porez += pz.Ppor
	}
	if len(rppo) > 0 {
		kal.UkupnoPorezi = domain.RobnoStampaKalkulacijaPorez{IznosDokumenta: f2(iznos), Osnovica: f2(osnovica), Porez: f2(porez)}
	}
	// Po kalkulaciji: the osnovica and the ukalkulisani PDV of the stavke per poreska oznaka.
	sort.Slice(redKalk, func(i, j int) bool { return redKalk[i] < redKalk[j] })
	var osnKal, porKal float64
	for _, po := range redKalk {
		pz := poreziKalk[po]
		kal.PoreziKalkulacije = append(kal.PoreziKalkulacije, domain.RobnoStampaKalkulacijaPorez{
			Osnovica: f2(pz.osnovica),
			Porez:    f2(pz.porez),
			Tarifa:   fmt.Sprintf("%d", pz.po),
			Stopa:    f2(pz.stopa),
		})
		osnKal += pz.osnovica
		porKal += pz.porez
	}
	kal.UkupnoPoreziKalk = domain.RobnoStampaKalkulacijaPorez{Osnovica: f2(osnKal), Porez: f2(porKal)}
	return kal
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

//
// Štampa knjižnog pisma: the report RobnoStampaKnjiznoPismo (the knjižna pisma with stavke, the groups
// KNO and KNZ, the legacy ROB_RPT_KNJPISMO with its query QRY_RPRO_ZADOK)
//

// robnoStampaKnjiznoPismoSuma is the osnovica and the PDV of one poreska stopa of a knjižno pismo: of the
// faktura it corrects and of the knjižno pismo itself (the legacy STPdv).
type robnoStampaKnjiznoPismoSuma struct {
	osnFakt, pdvFakt, osnKnp, pdvKnp float64
}

// robnoStampaKnjiznoPismoVrednosti returns the bruto vrednost, the rabat, the ugovoreni rabat, the kasa
// and the osnovica of a stavka like the legacy Preracun1/Preracun2: the bruto vrednost is the količina
// times the fakturna cena, the rabat its % rabata, the ugovoreni rabat (ugrabat %) is taken from the
// vrednost less the rabat and the kasa (pkase %) from what remains; the osnovica is the rest.
func robnoStampaKnjiznoPismoVrednosti(kolic, fcena, rab, ugrabat, pkase float64) (bruto, rabat, ugRabat, kasa, osnovica float64) {
	bruto = robnoStampaFakturaRound(kolic * fcena)
	rabat = robnoStampaFakturaRound(bruto * rab / 100)
	ugRabat = robnoStampaFakturaRound((bruto - rabat) * ugrabat / 100)
	kasa = robnoStampaFakturaRound((bruto - ugRabat - rabat) * pkase / 100)
	osnovica = robnoStampaFakturaRound(bruto - ugRabat - kasa - rabat)
	return
}

// GetStampaKnjiznoPismo returns the knjižna pisma with stavke of the selection, ready to print (the
// legacy ROB_RPT_KNJPISMO): one view per robni dokument with its stavke in the order of the rbr, and the
// izdavalac (fvr) of the print. Like the legacy procedures Preracun1 and Preracun2, the osnovica and the
// PDV per poreska stopa are calculated for the faktura the knjižno pismo corrects (rdok.vrdokid) and for
// the knjižno pismo (a stavka takes the količina, the cena and the rabat of the stavka that corrects it,
// rpro.rproid1, when there is one); the print shows their difference per stopa (PORESKA OSNOVICA, PDV
// PO STOPI) and the UKUPNA NAKNADA (the difference of the osnovica with the difference of the PDV; only
// the osnovica for a document with a poreska napomena). A negative naknada is a knjižno odobrenje
// ("Odobravamo Vam kako sledi"), a positive one a knjižno zaduženje; without a poreska napomena the
// napomena on the izmena of the poreska osnovica gives the difference of the PDV. BRUTO, RABAT, the
// ugovoreni rabat and the kasa are the ones of the knjižno pismo; the naknada is also given in words.
func (s *RobnoDokumentaResource) GetStampaKnjiznoPismo(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(rdok.kom, 0) as kom,
			coalesce(kom.imeprezime, '') as komnaziv,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.foot, '') as foot,
			coalesce(rdok.ugrabat, 0) as ugrabat,
			coalesce(rdok.pkase, 0) as pkase,
			coalesce(rdok.vrdokid, 0) as vrdokid,
			coalesce(tkp.sifrazlog || '-' || tkp.opis, '') as tipknjpisma,
			coalesce(vez.vrd, 0) as veznivrd,
			coalesce(vez.dokum, 0) as veznidokum,
			vez.dadok as veznidadok,
			coalesce(mg.mesto, '') as magmesto,
			coalesce(mg.adresa, '') as magadresa,
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
			coalesce(p.telefon, '') as kupactelefon,
			coalesce(misp.naziv, '') as mispnaziv,
			coalesce(misp.adresa, '') as mispadresa,
			coalesce(misp.pobro, 0) as misppobro,
			coalesce(misp.mesto, '') as mispmesto,
			coalesce(nullif(misp.gln, 0)::text, '') as mispgln,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(rsif.naziv, rpro.naz1, '') as naziv,
			coalesce(rsif.jm, rpro.jm, '') as jm,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.fcena, 0) as fcena,
			coalesce(rpro.rab, 0) as rab,
			coalesce(ps.pp, rpro.pdvpct, 0) as stopa,
			isp.rproid is not null as ispravkafound,
			coalesce(isp.kolic, 0) as ispravkakolic,
			coalesce(isp.fcena, 0) as ispravkafcena,
			coalesce(isp.rab, 0) as ispravkarab
		from rpro`, true)
	qb.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin("left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin("left join komercijalisti kom on kom.god = rdok.god and kom.kar = rdok.kar and kom.sifkom = rdok.kom")
	qb.AddJoin("left join tipknpisma tkp on tkp.tipknjid = rdok.tipknjid")
	qb.AddJoin("left join rdok vez on vez.rdokid = rdok.vrdokid and coalesce(rdok.vrdokid, 0) <> 0")
	qb.AddJoin("left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	// The kupac: the partner of the analitički konto (vkonta 1) of the konto and šifra of the document.
	qb.AddJoin(`left join lateral (select f.idpartneri from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.vkonta = 1 and f.konto = rdok.fkto and f.sifra = rdok.fana
		order by f.idfkpl limit 1) kup on true`)
	qb.AddJoin("left join partneri p on p.idpartneri = kup.idpartneri")
	qb.AddJoin(`left join lateral (select f.naziv, f.adresa, f.pobro, f.mesto, f.gln from fisp f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
			and f.mi = coalesce(rdok.mi, 0)
		order by f.fispid limit 1) misp on true`)
	// The stavka that corrects this stavka (the legacy HReadSeekFirst(RPRO, RPROID1, ...)).
	qb.AddJoin(`left join lateral (select r.rproid, r.kolic, r.fcena, r.rab from rpro r
		where r.rproid1 = rpro.rproid order by r.rproid limit 1) isp on true`)
	qb.AddJoin(robnoStampaFakturaStopaJoin)
	robnoStampaDokumentaUslovi(qb, userSession, params)
	qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	sqlQuery, args := qb.Build()
	rows, err := s.stampaKnjiznoPismoRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	if len(*rows) == 0 {
		return []domain.RobnoStampaFakturaView{}, firma, nil
	}

	// The stavke of the fakture the knjižna pisma correct (the legacy Preracun1).
	var fakturaIDs []any
	vidjena := map[int64]bool{}
	for _, r := range *rows {
		if r.VrdokID != 0 && !vidjena[r.VrdokID] {
			vidjena[r.VrdokID] = true
			fakturaIDs = append(fakturaIDs, r.VrdokID)
		}
	}
	fakturaOf := map[int64][]domain.RobnoStampaKnjiznoPismoFakturaDto{}
	if len(fakturaIDs) > 0 {
		fq := common.NewQueryBuilder(`
			select
				rdok.rdokid,
				coalesce(rpro.kolic, 0) as kolic,
				coalesce(rpro.fcena, 0) as fcena,
				coalesce(rpro.rab, 0) as rab,
				coalesce(ps.pp, rpro.pdvpct, 0) as stopa,
				coalesce(rdok.ugrabat, 0) as ugrabat,
				coalesce(rdok.pkase, 0) as pkase
			from rpro`, true)
		fq.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
		fq.AddJoin(robnoStampaFakturaStopaJoin)
		fq.AddIn("rdok.rdokid", fakturaIDs)
		fq.AddOrderBy("rpro.rproid")
		fSQL, fArgs := fq.Build()
		stavke, err := s.stampaKnjiznoPismoFakturaRepo.GetAllCustom(ctx, fSQL, "", fArgs, "", "")
		if err != nil {
			return nil, firma, err
		}
		for _, st := range *stavke {
			fakturaOf[st.RdokID] = append(fakturaOf[st.RdokID], st)
		}
	}

	pisma := []domain.RobnoStampaFakturaView{}
	for i := 0; i < len(*rows); {
		j := i
		for j < len(*rows) && (*rows)[j].RdokID == (*rows)[i].RdokID {
			j++
		}
		dok := (*rows)[i]
		pisma = append(pisma, robnoStampaKnjiznoPismoView(firma, (*rows)[i:j], fakturaOf[dok.VrdokID]))
		i = j
	}
	return pisma, firma, nil
}

// robnoStampaKnjiznoPismoView builds the printed knjižno pismo of the stavke of one robni dokument and
// of the stavke of the faktura it corrects (see GetStampaKnjiznoPismo).
func robnoStampaKnjiznoPismoView(firma domain.RobnoStampaFakturaFirmaDto, rows []domain.RobnoStampaKnjiznoPismoRowDto, faktura []domain.RobnoStampaKnjiznoPismoFakturaDto) domain.RobnoStampaFakturaView {
	dok := rows[0]
	fak := robnoStampaKnjiznoPismoZaglavlje(firma, dok)
	robnoStampaKnjiznoPismoStavke(&fak, dok, rows, faktura)
	return fak
}

// robnoStampaKnjiznoPismoZaglavlje returns the header of a printed knjižno pismo (with or without
// stavke): the kupac with its PIB line, the komercijalista, the broj, the date and the nalog of the
// document, the napomene, the tip of the knjižno pismo, the vezni dokument, the izdavalac with the adresa
// and the mesto of the magacin and the mesto isporuke.
func robnoStampaKnjiznoPismoZaglavlje(firma domain.RobnoStampaFakturaFirmaDto, dok domain.RobnoStampaKnjiznoPismoRowDto) domain.RobnoStampaFakturaView {
	lbl := i18n.GetInstance().Label
	fak := domain.RobnoStampaFakturaView{
		KupacNaziv:   dok.KupacNaziv,
		KupacAdresa:  dok.KupacAdresa,
		KupacMesto:   robnoStampaFakturaIzvozMesto(dok.KupacPobro, dok.KupacMesto),
		KupacTelefon: dok.KupacTelefon,
		KupacPib: robnoStampaPartnerPib(robnoStampaPartner{
			TipPdv: dok.KupacTipPdv, Pib: dok.KupacPib, Budzetski: dok.KupacBudzetski, Jbkjs: dok.KupacJbkjs,
			Jmbg: dok.KupacJmbg, Bpg: dok.KupacBpg, Index: dok.KupacIndex,
		}, ":"),
		Komercijalista:     fmt.Sprintf("%d", dok.Kom),
		KomercijalistaOpis: dok.KomNaziv,
		BrojDokumenta:      fmt.Sprintf("%d-%d", dok.Vrd, dok.Dokum),
		DatumDokumenta:     common.FormatNullTime(dok.Dadok, common.DateLayout),
		Nalog:              fmt.Sprintf("%s-%d", dok.Tipdok, dok.Nalog),
		Napomena:           dok.Foot,
		PoreskaNapomena:    dok.Pornapomena,
		TipKnjiznogPisma:   dok.TipKnjPisma,
	}
	if dok.KupacNaziv != "" {
		fak.KupacKonto, fak.KupacSifra = dok.Fkto, dok.Fana
	}
	if dok.VezniVrd != 0 || dok.VezniDokum != 0 {
		datum := ""
		if dok.VezniDadok.Valid {
			datum = dok.VezniDadok.Time.Format("02/01/2006")
		}
		fak.VezniDokument = fmt.Sprintf("%s: %d-%d %s : %s", lbl("Vezni dokum."), dok.VezniVrd, dok.VezniDokum, lbl("Od datuma"), datum)
	}
	// The izdavalac: the adresa with the adresa of the magacin of the document and its mesto as the
	// mesto izdavanja dokumenta.
	fak.FirmaAdresa = strings.TrimSpace(firma.Pobro+" "+firma.Mesto) + " , " + firma.Adresa
	fak.MestoIzdavanja = firma.Mesto
	if dok.MagMesto != "" || dok.MagAdresa != "" {
		fak.MestoIzdavanja = dok.MagMesto
		if dok.MagAdresa != "" {
			fak.FirmaAdresa += " , " + dok.MagAdresa
		}
	}
	if firma.Obv {
		fak.FirmaObveznik = lbl("Obv. PDV-a br") + ": " + firma.Brobvpdv
	} else {
		fak.FirmaObveznik = lbl("Nije obveznik PDV-a")
	}
	if firma.Apr != "" {
		fak.FirmaApr = lbl("Rešenje APR") + " " + firma.Apr
	}
	fak.FirmaPib = lbl("PIB") + " : " + firma.Pib
	fak.FirmaMbr = lbl("Matični broj") + " : " + firma.Matbr
	fak.FirmaSifdel = lbl("Šifra delatnosti") + " : " + firma.Sifdel
	fak.FirmaTel = lbl("Telefon/faks") + " :" + firma.Tel
	if dok.MispNaziv != "" {
		fak.MestoIsporuke = lbl("M. Isporuke") + ":" + dok.MispNaziv
		fak.MestoIsporuke2 = dok.MispAdresa + ", " + robnoStampaFakturaIzvozMesto(dok.MispPobro, dok.MispMesto)
		if dok.MispGln != "" {
			fak.MestoIsporuke2 += " " + lbl("GLN") + ":" + dok.MispGln
		}
	}
	return fak
}

// robnoStampaKnjiznoPismoStavke fills the stavke and the totals of a knjižno pismo with stavke: the
// legacy Preracun1 and Preracun2 (see GetStampaKnjiznoPismo).
func robnoStampaKnjiznoPismoStavke(fak *domain.RobnoStampaFakturaView, dok domain.RobnoStampaKnjiznoPismoRowDto, rows []domain.RobnoStampaKnjiznoPismoRowDto, faktura []domain.RobnoStampaKnjiznoPismoFakturaDto) {
	lbl := i18n.GetInstance().Label
	// Preracun1: the osnovica and the PDV per stopa of the faktura.
	sume := map[float64]*robnoStampaKnjiznoPismoSuma{}
	suma := func(stopa float64) *robnoStampaKnjiznoPismoSuma {
		if sume[stopa] == nil {
			sume[stopa] = &robnoStampaKnjiznoPismoSuma{}
		}
		return sume[stopa]
	}
	for _, st := range faktura {
		_, _, _, _, osn := robnoStampaKnjiznoPismoVrednosti(st.Kolic, st.Fcena, st.Rab, st.Ugrabat, st.Pkase)
		sm := suma(st.Stopa)
		sm.osnFakt += osn
		sm.pdvFakt += robnoStampaFakturaRound(osn * st.Stopa / 100)
	}
	// Preracun2: the osnovica and the PDV per stopa of the knjižno pismo, its totals and its stavke.
	var bruto, rabat, ugRabat, kasa float64
	for _, st := range rows {
		kolic, fcena, rab := st.Kolic, st.Fcena, st.Rab
		if st.IspravkaFound {
			kolic, fcena, rab = st.IspravkaKolic, st.IspravkaFcena, st.IspravkaRab
		}
		b, r, ug, ks, osn := robnoStampaKnjiznoPismoVrednosti(kolic, fcena, rab, dok.Ugrabat, dok.Pkase)
		bruto += b
		rabat += r
		ugRabat += ug
		kasa += ks
		sm := suma(st.Stopa)
		sm.osnKnp += osn
		sm.pdvKnp += robnoStampaFakturaRound(osn * st.Stopa / 100)

		// The printed stavka: the ispravka of the cena of the stavka, its rabat, its PDV of one unit and
		// its iznos.
		iznos := robnoStampaFakturaRound(st.Kolic * st.Fcena)
		fak.Stavke = append(fak.Stavke, domain.RobnoStampaFakturaStavka{
			Rbr:        fmt.Sprintf("%d", st.Rbr),
			Sifra:      robnoStampaFakturaSifra(st.Sifra),
			Naziv:      st.Naziv,
			Jm:         st.Jm,
			Kolicina:   common.FormatNumberWithSystemLocale(st.Kolic, 3),
			Cena:       common.FormatNumberWithSystemLocale(st.Fcena, 2),
			ProcRabata: common.FormatNumberWithSystemLocale(st.Rab, 2) + "%",
			Rabat:      common.FormatNumberWithSystemLocale(robnoStampaFakturaRound(iznos*st.Rab/100), 2),
			ProcPdv:    robnoStampaFakturaStopa(st.Stopa),
			PdvPoJm:    common.FormatNumberWithSystemLocale(robnoStampaFakturaRound(st.Fcena*(100-st.Rab)/100*st.Stopa/100), 2),
			Iznos:      common.FormatNumberWithSystemLocale(iznos, 2),
		})
	}
	// The difference per stopa (the lowest stopa first) and the naknada.
	var stope []float64
	for stopa := range sume {
		stope = append(stope, stopa)
	}
	sort.Float64s(stope)
	var totOsn, totPdv float64
	for _, stopa := range stope {
		sm := sume[stopa]
		osn := sm.osnKnp - sm.osnFakt
		pdv := sm.pdvKnp - sm.pdvFakt
		totOsn += osn
		totPdv += pdv
		fak.PdvPoStopama = append(fak.PdvPoStopama, domain.RobnoStampaFakturaPdv{
			Stopa:    common.FormatNumberWithSystemLocale(stopa, 2) + "%",
			Osnovica: common.FormatNumberWithSystemLocale(osn, 2),
			Pdv:      common.FormatNumberWithSystemLocale(pdv, 2),
		})
	}
	naknada := totOsn + totPdv
	if strings.TrimSpace(dok.Pornapomena) != "" {
		naknada = totOsn
	}
	naknada = robnoStampaFakturaRound(naknada)
	fak.Svega = common.FormatNumberWithSystemLocale(bruto, 2)
	fak.Rabat = common.FormatNumberWithSystemLocale(rabat, 2)
	if dok.Ugrabat != 0 || dok.Pkase != 0 {
		fak.Neto = common.FormatNumberWithSystemLocale(bruto-rabat, 2)
		fak.UgovoreniRabat = common.FormatNumberWithSystemLocale(ugRabat, 2)
		fak.Kasa = common.FormatNumberWithSystemLocale(kasa, 2)
	}
	fak.ZaNaplatu = common.FormatNumberWithSystemLocale(naknada, 2)
	fak.Slovima = BrojSlovima(math.Abs(naknada))
	napomenaPorez := strings.TrimSpace(dok.Pornapomena) == ""
	pdvIznos := common.FormatNumberWithSystemLocale(math.Abs(robnoStampaFakturaRound(totPdv)), 2)
	switch {
	case naknada < 0:
		fak.Naslov = lbl("Knjižno odobrenje Br.")
		fak.OdobravaZaduzuje = lbl("Odobravamo Vam kako sledi") + ":"
		if napomenaPorez {
			fak.NapomenaPorez = i18n.GetInstance().TextWithParams("knjizno_pismo_odobrenje_napomena", map[string]string{"iznos": pdvIznos})
		}
	case naknada > 0:
		fak.Naslov = lbl("Knjižno zaduženje Br.")
		fak.OdobravaZaduzuje = lbl("Zadužujemo Vas kako sledi") + ":"
		if napomenaPorez {
			fak.NapomenaPorez = i18n.GetInstance().TextWithParams("knjizno_pismo_zaduzenje_napomena", map[string]string{"iznos": pdvIznos})
		}
	}
}

// robnoStampaPartner are the data of a partner (partneri) its PIB line is made of.
type robnoStampaPartner struct {
	TipPdv                       int64
	Pib, Jbkjs, Jmbg, Bpg, Index string
	Budzetski                    bool
}

// robnoStampaPartnerPib returns the PIB line of a kupac by the PDV status of the partner
// (partneri.tippdv), like the legacy reports (Srbija): PIB (with the JBKJS of a budžetski korisnik) for
// the PDV obveznici and the others with a PIB (1, 2), JMBG and BPG for the poljoprivrednici (3), JMBG for
// the fizička lica (4) and JMBG and INDEX for 5; sep is what follows the captions (":" or ": ").
func robnoStampaPartnerPib(p robnoStampaPartner, sep string) string {
	switch {
	case p.TipPdv <= 2:
		pib := "PIB " + sep + p.Pib
		if p.Budzetski {
			pib += "    JBKJS : " + p.Jbkjs
		}
		return pib
	case p.TipPdv == 3:
		var pib string
		if p.Jmbg != "" {
			pib = "JMBG " + sep + p.Jmbg + " ; "
		}
		if p.Bpg != "" {
			pib += "BPG " + sep + p.Bpg
		}
		return pib
	case p.TipPdv == 4 && p.Jmbg != "":
		return "JMBG " + sep + p.Jmbg
	case p.TipPdv == 5:
		var pib string
		if p.Jmbg != "" {
			pib = "JMBG " + sep + p.Jmbg + " ; "
		}
		if p.Index != "" {
			pib += "INDEX " + sep + p.Index
		}
		return pib
	}
	return ""
}

// robnoStampaDinar is the šifra of the valuta of the dinar (valute.sifval).
const robnoStampaDinar = 941

// GetStampaKnjiznoPismoFin returns the finansijska knjižna pisma (without stavke) of the selection,
// ready to print (the legacy ROB_RPT_KNJPISMOFIN, a report without a source query: the data of the
// robni dokument, rdok): the header of the knjižno pismo (see GetStampaKnjiznoPismo), the poreska
// osnovica, the PDV with its stopa and the UKUPNA NAKNADA in words (see robnoStampaKnjiznoPismoFinView),
// the napomena and the napomena on the izmena of the poreska osnovica with the PDV of the document. The vrsta dokumenta with the oznaka KNO is a knjižno odobrenje, KNZ a knjižno zaduženje.
func (s *RobnoDokumentaResource) GetStampaKnjiznoPismoFin(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(rdok.kom, 0) as kom,
			coalesce(kom.imeprezime, '') as komnaziv,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.foot, '') as foot,
			coalesce(tkp.sifrazlog || '-' || tkp.opis, '') as tipknjpisma,
			coalesce(dokvrsta.dokozn, '') as dokozn,
			coalesce(rdok.iznos, 0) as iznos,
			coalesce(rdok.vporez, 0) as vporez,
			coalesce(rdok.pkase, 0) as pkase,
			coalesce((select fvr.sifval from fvr where fvr.god = rdok.god and fvr.kar = rdok.kar limit 1), 0) as domacavaluta,
			coalesce(mg.mesto, '') as magmesto,
			coalesce(mg.adresa, '') as magadresa,
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
			coalesce(p.telefon, '') as kupactelefon,
			coalesce(misp.naziv, '') as mispnaziv,
			coalesce(misp.adresa, '') as mispadresa,
			coalesce(misp.pobro, 0) as misppobro,
			coalesce(misp.mesto, '') as mispmesto,
			coalesce(nullif(misp.gln, 0)::text, '') as mispgln
		from rdok`, true)
	qb.AddJoin("left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin("left join komercijalisti kom on kom.god = rdok.god and kom.kar = rdok.kar and kom.sifkom = rdok.kom")
	qb.AddJoin("left join tipknpisma tkp on tkp.tipknjid = rdok.tipknjid")
	qb.AddJoin("left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	qb.AddJoin(`left join lateral (select f.idpartneri from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.vkonta = 1 and f.konto = rdok.fkto and f.sifra = rdok.fana
		order by f.idfkpl limit 1) kup on true`)
	qb.AddJoin("left join partneri p on p.idpartneri = kup.idpartneri")
	qb.AddJoin(`left join lateral (select f.naziv, f.adresa, f.pobro, f.mesto, f.gln from fisp f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
			and f.mi = coalesce(rdok.mi, 0)
		order by f.fispid limit 1) misp on true`)
	robnoStampaDokumentaUsloviTabele(qb, userSession, params, "rdok")
	qb.AddOrderBy("rdok.rdokid")
	sqlQuery, args := qb.Build()
	rows, err := s.stampaKnjiznoPismoRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	pisma := make([]domain.RobnoStampaFakturaView, 0, len(*rows))
	for _, dok := range *rows {
		pisma = append(pisma, robnoStampaKnjiznoPismoFinView(firma, dok))
	}
	return pisma, firma, nil
}

// robnoStampaKnjiznoPismoFinView builds the printed finansijsko knjižno pismo of a robni dokument (see
// GetStampaKnjiznoPismoFin) like the legacy BREAK_FOTRDOKID: every amount without its sign, the poreska
// osnovica (rdok.iznos), the PDV (rdok.vporez) with its stopa (held in rdok.pkase) printed only for a
// PDV obveznik and a document without a poreska napomena, the UKUPNA NAKNADA (osnovica with the PDV) and
// the naknada in words only when the domestic valuta of the firm is the dinar (941).
func robnoStampaKnjiznoPismoFinView(firma domain.RobnoStampaFakturaFirmaDto, dok domain.RobnoStampaKnjiznoPismoRowDto) domain.RobnoStampaFakturaView {
	lbl := i18n.GetInstance().Label
	fak := robnoStampaKnjiznoPismoZaglavlje(firma, dok)
	if firma.Obv && strings.TrimSpace(dok.Pornapomena) == "" {
		fak.PdvPoStopama = []domain.RobnoStampaFakturaPdv{{
			Stopa:    common.FormatNumberWithSystemLocale(math.Abs(dok.Pkase), 2) + "%",
			Osnovica: common.FormatNumberWithSystemLocale(math.Abs(dok.Iznos), 2),
			Pdv:      common.FormatNumberWithSystemLocale(math.Abs(dok.Vporez), 2),
		}}
	}
	naknada := math.Abs(robnoStampaFakturaRound(dok.Iznos + dok.Vporez))
	fak.ZaNaplatu = common.FormatNumberWithSystemLocale(naknada, 2)
	if dok.DomacaValuta == robnoStampaDinar {
		fak.Slovima = BrojSlovima(naknada)
	}
	pdvIznos := common.FormatNumberWithSystemLocale(dok.Vporez, 2)
	switch strings.ToUpper(strings.TrimSpace(dok.Dokozn)) {
	case "KNO":
		fak.Naslov = lbl("Knjižno odobrenje Br.")
		fak.OdobravaZaduzuje = lbl("Odobravamo Vam kako sledi") + ":"
		fak.NapomenaPorez = i18n.GetInstance().TextWithParams("knjizno_pismo_odobrenje_napomena", map[string]string{"iznos": pdvIznos})
	case "KNZ":
		fak.Naslov = lbl("Knjižno zaduženje Br.")
		fak.OdobravaZaduzuje = lbl("Zadužujemo Vas kako sledi") + ":"
		fak.NapomenaPorez = i18n.GetInstance().TextWithParams("knjizno_pismo_zaduzenje_napomena", map[string]string{"iznos": pdvIznos})
	}
	return fak
}

//
// Štampa internog prenosa proizvodnje: the report RobnoStampaInterniPrenos (the group PPR of the vrste
// dokumenta, the legacy ROB_RPT_INTRAC_PROIZV with its query ROB_QRY_IRTSTAMPA)
//

// GetStampaInterniPrenosProizvodnje returns the interni prenosi proizvodnje of the selection, ready to
// print (the legacy ROB_RPT_INTRAC_PROIZV): one view per robni dokument with its stavke in the order of
// the rbr (without the ambalaža, šifra 990000 and over), the totals and the porezi, and the izdavalac
// (fvr) of the print. The stavke are calculated like the legacy body (see robnoStampaInterniPrenosView).
// The objekat the goods go to is the fkpl of rdok.pkto and rdok.pana (for a konto starting with 9, of
// the karton of the pogonsko knjigovodstvo, fvrcmp.kar1).
func (s *RobnoDokumentaResource) GetStampaInterniPrenosProizvodnje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaInterniPrenosView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaInterniPrenosi(ctx, params, false)
}

// GetStampaInternaZakljucnica returns the interne zaključnice of the selection, ready to print (the
// group IRT of the vrste dokumenta, the legacy ROB_RPT_STAMPA_INTRAC with the same query
// ROB_QRY_IRTSTAMPA): like the interni prenos proizvodnje (see GetStampaInterniPrenosProizvodnje), with
// the barkod of the artikal and, in a magacin with the tip zaliha 2, its otk, serija and rok trajanja
// (with the captions of the firm, rvr) under the naziv; the nabavna vrednost of an artikal of the grupa
// 777 does not deduct the taksa and the stopa of the porez out of the iznos has 4 decimals.
func (s *RobnoDokumentaResource) GetStampaInternaZakljucnica(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaInterniPrenosView, domain.RobnoStampaFakturaFirmaDto, error) {
	return s.stampaInterniPrenosi(ctx, params, true)
}

// GetStampaInternaZakljucnicaRacun returns the računi - otpremnice of the interne zaključnice of the
// selection, ready to print (the group IRT with the option irrn, the legacy ROB_RPT_STAMPA_INTRACFKT with
// the same query ROB_QRY_IRTSTAMPA): the stavke and the porezi of the interna zaključnica (see
// GetStampaInternaZakljucnica) with the header of a račun (the kupac, the mesto isporuke, the izdavalac
// with the adresa and the mesto of the magacin; see robnoStampaInternaZakljucnicaRacunView).
func (s *RobnoDokumentaResource) GetStampaInternaZakljucnicaRacun(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaInternaZakljucnicaRacunView, domain.RobnoStampaFakturaFirmaDto, error) {
	prenosi, firma, err := s.stampaInterniPrenosi(ctx, params, true)
	if err != nil || len(prenosi) == 0 {
		return []domain.RobnoStampaInternaZakljucnicaRacunView{}, firma, err
	}
	var rdokIDs []any
	for _, p := range prenosi {
		rdokIDs = append(rdokIDs, p.RdokID)
	}
	// The header of the račun: the document, the kupac (the partner of the analitički konto of rdok.fkto
	// and rdok.fana), the mesto isporuke (fisp of rdok.mi) and the magacin.
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.rok, 0) as rok,
			coalesce(rdok.pla, '') as pla,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.foot, '') as foot,
			coalesce(mg.mesto, '') as magmesto,
			coalesce(mg.adresa, '') as magadresa,
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
			coalesce(p.telefon, '') as kupactelefon,
			coalesce(misp.naziv, '') as mispnaziv,
			coalesce(misp.adresa, '') as mispadresa,
			coalesce(misp.pobro, 0) as misppobro,
			coalesce(misp.mesto, '') as mispmesto
		from rdok`, true)
	qb.AddJoin("left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	qb.AddJoin(`left join lateral (select f.idpartneri from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.vkonta = 1 and f.konto = rdok.fkto and f.sifra = rdok.fana
		order by f.idfkpl limit 1) kup on true`)
	qb.AddJoin("left join partneri p on p.idpartneri = kup.idpartneri")
	qb.AddJoin(`left join lateral (select f.naziv, f.adresa, f.pobro, f.mesto from fisp f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
			and f.mi = coalesce(rdok.mi, 0)
		order by f.fispid limit 1) misp on true`)
	qb.AddIn("rdok.rdokid", rdokIDs)
	sqlQuery, args := qb.Build()
	rows, err := s.stampaKnjiznoPismoRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	zaglavlja := map[int64]domain.RobnoStampaKnjiznoPismoRowDto{}
	for _, r := range *rows {
		zaglavlja[r.RdokID] = r
	}
	racuni := make([]domain.RobnoStampaInternaZakljucnicaRacunView, 0, len(prenosi))
	for _, p := range prenosi {
		racuni = append(racuni, robnoStampaInternaZakljucnicaRacunView(firma, zaglavlja[p.RdokID], p))
	}
	return racuni, firma, nil
}

// robnoStampaInternaZakljucnicaRacunView builds the printed račun - otpremnica of an interna zaključnica
// like the legacy report: the kupac ("Šifra kupca" konto šifra, the naziv, poštanski broj, mesto and
// adresa in one line, the PIB line by the PDV status and the telefon) with the mesto isporuke; the
// izdavalac with its adresa (and the adresa of the magacin), the žiro računi and the banks, the PIB, the
// matični broj and the telefon; "Račun - otpremnica: nalog-vrsta-broj" with the datum prometa dobara
// (the date of the document), the valuta (plus the rok), the datum dokumenta (the day of the print, the
// legacy Today()), the nalog and the mesto izdavanja (the mesto of the magacin); the otpremnica, the
// uslovi and the rok plaćanja. The PDV per stopa are the porezi of the interna zaključnica, SVEGA and ZA
// NAPLATU its maloprodajna vrednost.
func robnoStampaInternaZakljucnicaRacunView(firma domain.RobnoStampaFakturaFirmaDto, dok domain.RobnoStampaKnjiznoPismoRowDto, prenos domain.RobnoStampaInterniPrenosView) domain.RobnoStampaInternaZakljucnicaRacunView {
	lbl := i18n.GetInstance().Label
	fak := robnoStampaKnjiznoPismoZaglavlje(firma, dok)
	fak.KupacAdresa = strings.TrimSpace(robnoStampaFakturaIzvozMesto(dok.KupacPobro, dok.KupacMesto) + " " + dok.KupacAdresa)
	fak.KupacMesto = ""
	if dok.KupacNaziv == "" {
		// The kupac is printed only when the partner is found.
		fak.KupacPib = ""
	}
	fak.Komercijalista, fak.KomercijalistaOpis = "", ""
	fak.Naslov = lbl("Račun - otpremnica")
	fak.BrojDokumenta = fmt.Sprintf("%d-%d-%d", dok.Nalog, dok.Vrd, dok.Dokum)
	fak.DatumPrometa = common.FormatNullTime(dok.Dadok, common.DateLayout)
	fak.Valuta = common.AddDaysToNullTime(dok.Dadok, int(dok.Rok), common.DateLayout)
	fak.DatumDokumenta = time.Now().Format(common.DateLayout)
	fak.Otpremnica = dok.Dokiz
	fak.UsloviPlacanja = dok.Pla
	fak.RokPlacanja = fmt.Sprintf("%d", dok.Rok)
	fak.FirmaPib = lbl("PIB") + " : " + firma.Pib
	fak.FirmaMbr = lbl("Mat. broj") + " : " + firma.Matbr
	fak.FirmaTel = lbl("Telefon/faks") + " : " + firma.Tel
	fak.FirmaZiro = lbl("Žiro račun") + " : " + firma.BankeRacuni
	fak.FirmaBankeNazivi = firma.BankeNazivi
	if dok.MispNaziv != "" {
		fak.MestoIsporuke2 = dok.MispAdresa + ", " + robnoStampaFakturaIzvozMesto(dok.MispPobro, dok.MispMesto)
	}
	for _, pz := range prenos.Porezi {
		fak.PdvPoStopama = append(fak.PdvPoStopama, domain.RobnoStampaFakturaPdv{
			Stopa:    pz.Stopa,
			Osnovica: pz.Osnovica,
			Pdv:      pz.Pdv,
			Iznos:    pz.MaloprodajniIznos,
		})
	}
	fak.Svega = prenos.Ukupno.McVrednost
	fak.Rabat = common.FormatNumberWithSystemLocale(float64(0), 2)
	fak.ZaNaplatu = prenos.Ukupno.McVrednost
	return domain.RobnoStampaInternaZakljucnicaRacunView{Faktura: fak, Prenos: prenos}
}

// stampaInterniPrenosi returns the interni prenosi proizvodnje or (interna) the interne zaključnice of
// the selection, ready to print (see GetStampaInterniPrenosProizvodnje and GetStampaInternaZakljucnica).
func (s *RobnoDokumentaResource) stampaInterniPrenosi(ctx context.Context, params domain.RobnoStampaFakturaParams, interna bool) ([]domain.RobnoStampaInterniPrenosView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.tipdok, '') as tipdok,
			coalesce(rdok.nalog, 0) as nalog,
			coalesce(dokvrsta.opis, '') as vrdopis,
			coalesce(mg.mag || '-' || mg.opis, '') as magoznaka,
			concat_ws(' ,', nullif(trim(mg.adresa), ''), nullif(trim(mg.mesto), '')) as magadresa,
			coalesce(rdok.pkto, '') as pkto,
			coalesce(rdok.pana, '') as pana,
			coalesce(prijem.naziv, '') as prijemnaziv,
			prijem.naziv is not null as prijemnadjen,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(rpro.sifra, 0) as sifra,
			coalesce(rsif.naziv, rpro.naz1, '') as naziv,
			coalesce(rsif.jm, rpro.jm, '') as jm,
			coalesce(rsif.model, '') as model,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.cena, 0) as cena,
			coalesce(rpro.mcenap, 0) as mcenap,
			coalesce(rpro.vra, 0) as vra,
			coalesce(rpro.po, 0) as po,
			rpro.dadok as stavkadadok,
			coalesce(rc.itaksa, 0) as itaksa,
			coalesce(rc.pakc, 0) as pakc,
			coalesce(rc.iakc, 0) as iakc,
			coalesce(rc.vma, 0) as vma,
			coalesce(rc.mma, 0) as mma,
			coalesce(rsif.barkod, '') as barkod,
			coalesce(rsif.gru, 0) as gru,
			coalesce(rpro.otk, '') as otk,
			coalesce(rpro.serija, '') as serija,
			coalesce(rpro.roktr, 0)::bigint as roktr,
			coalesce(mg.tipzal, 0) as tipzal,
			coalesce(rvr.otklbl, '') as otklbl,
			coalesce(rvr.serijalbl, '') as serijalbl,
			coalesce(rvr.roklbl, '') as roklbl
		from rpro`, true)
	qb.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin("left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin("left join rsif on rsif.god = rpro.god and rsif.kar = rpro.kar and rsif.sifra = rpro.sifra")
	qb.AddJoin(`left join lateral (select c.itaksa, c.pakc, c.iakc, c.vma, c.mma from rcene c
		where c.rsifid = rsif.rsifid order by c.rceneid limit 1) rc on true`)
	qb.AddJoin("left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	// The captions of the otk, the serija and the rok trajanja of the firm.
	qb.AddJoin(`left join lateral (select v.otklbl, v.serijalbl, v.roklbl from rvr v
		where v.god = rdok.god and v.kar = rdok.kar limit 1) rvr on true`)
	// The objekat the goods go to: the fkpl of rdok.pkto and rdok.pana; a konto starting with 9 is in
	// the karton of the pogonsko knjigovodstvo of the firm (fvrcmp.kar1).
	qb.AddJoin(`left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.konto = rdok.pkto and f.sifra = rdok.pana
			and f.kar = case when left(coalesce(rdok.pkto, ''), 1) = '9' then coalesce((select c.kar1::int from fvrcmp c
				where c.god = rdok.god and c.kar = rdok.kar and c.knjigovod = 'Pogonsko' and c.kar1 ~ '^[0-9]+$' limit 1), rdok.kar)
				else rdok.kar end
		order by f.vkonta desc, f.idfkpl limit 1) prijem on true`)
	robnoStampaDokumentaUsloviTabele(qb, userSession, params, "rdok")
	qb.AddCustomCondition("coalesce(rpro.sifra, 0) < 990000")
	qb.AddOrderBy("rdok.rdokid, rpro.rbr")
	sqlQuery, args := qb.Build()
	rows, err := s.stampaInterniPrenosRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	if len(*rows) == 0 {
		return []domain.RobnoStampaInterniPrenosView{}, firma, nil
	}
	// The poreske stope of the poreske oznake of the stavke (the legacy VratiPorez2 and
	// Rob_QRY_AKTPORSTOPA).
	var oznake []int64
	for _, r := range *rows {
		oznake = append(oznake, r.Po)
	}
	stope, err := s.stampaPoreskeStope(ctx, oznake)
	if err != nil {
		return nil, firma, err
	}

	prenosi := []domain.RobnoStampaInterniPrenosView{}
	for i := 0; i < len(*rows); {
		j := i
		for j < len(*rows) && (*rows)[j].RdokID == (*rows)[i].RdokID {
			j++
		}
		prenosi = append(prenosi, robnoStampaInterniPrenosView((*rows)[i:j], stope, interna))
		i = j
	}
	return prenosi, firma, nil
}

// stampaPoreskeStope returns the stope (rpor) of the poreske oznake, in the order of the oznaka, the tip
// and the datum.
func (s *RobnoDokumentaResource) stampaPoreskeStope(ctx context.Context, oznake []int64) ([]domain.RobnoStampaPoreskaStopaDto, error) {
	var in []any
	vidjena := map[int64]bool{}
	for _, po := range oznake {
		if !vidjena[po] {
			vidjena[po] = true
			in = append(in, po)
		}
	}
	pq := common.NewQueryBuilder(`select coalesce(po, 0) as po, coalesce(tip, 0) as tip, coalesce(pt, '') as pt,
		coalesce(pp, 0) as pp, datum from rpor`, true)
	pq.AddIn("po", in)
	pq.AddOrderBy("po, tip, datum")
	pSQL, pArgs := pq.Build()
	stope, err := s.stampaPoreskaStopaRepo.GetAllCustom(ctx, pSQL, "", pArgs, "", "")
	if err != nil {
		return nil, err
	}
	return *stope, nil
}

// robnoStampaInternaZakljucnicaNaziv returns the naziv of a stavka of the interna zaključnica: the naziv
// of the artikal and, in a magacin with the tip zaliha 2 and when the stavka has them, a second line with
// the otk, the serija and the rok trajanja, each with the caption of the firm (rvr; "OTK", "Serija" and
// "Rok trajanja" when the firm has none).
func robnoStampaInternaZakljucnicaNaziv(st domain.RobnoStampaInterniPrenosRowDto) string {
	naziv := st.Naziv
	if st.Tipzal != 2 || (st.Otk == "" && st.Serija == "" && st.Roktr == 0) {
		return naziv
	}
	caption := func(firme, osnovni string) string {
		if firme != "" {
			return firme
		}
		return osnovni
	}
	naziv += "\n"
	if st.Otk != "" {
		naziv += caption(st.OtkLbl, "OTK") + ": " + st.Otk + " "
	}
	if st.Serija != "" {
		naziv += caption(st.SerijaLbl, "Serija") + ": " + st.Serija + " "
	}
	if st.Roktr != 0 {
		naziv += caption(st.RokLbl, "Rok trajanja") + ": " + robnoStampaWinDevDatum(st.Roktr) + " "
	}
	return strings.TrimRight(naziv, " ")
}

// robnoStampaPoreskaStopa returns the stopa of the poreska oznaka po on the date (the legacy
// VratiPorez2: the stopa of the latest rpor of the oznaka from before the date).
func robnoStampaPoreskaStopa(stope []domain.RobnoStampaPoreskaStopaDto, po int64, datum sql.NullTime) float64 {
	var nadjena *domain.RobnoStampaPoreskaStopaDto
	for i := range stope {
		st := &stope[i]
		if st.Po != po || !st.Datum.Valid || (datum.Valid && st.Datum.Time.After(datum.Time)) {
			continue
		}
		if nadjena == nil || !st.Datum.Time.Before(nadjena.Datum.Time) {
			nadjena = st
		}
	}
	if nadjena == nil {
		return 0
	}
	return nadjena.Pp
}

// robnoStampaAktivnePoreskeStope returns the porezi of the poreska oznaka po in force on the date (the
// legacy Rob_QRY_AKTPORSTOPA): per tip of the porez the latest rpor from before the date, ordered by
// the tip.
func robnoStampaAktivnePoreskeStope(stope []domain.RobnoStampaPoreskaStopaDto, po int64, datum sql.NullTime) []domain.RobnoStampaPoreskaStopaDto {
	poTipu := map[int64]domain.RobnoStampaPoreskaStopaDto{}
	var tipovi []int64
	for _, st := range stope {
		if st.Po != po || !st.Datum.Valid || (datum.Valid && st.Datum.Time.After(datum.Time)) {
			continue
		}
		prev, found := poTipu[st.Tip]
		if !found {
			tipovi = append(tipovi, st.Tip)
		}
		if !found || !st.Datum.Time.Before(prev.Datum.Time) {
			poTipu[st.Tip] = st
		}
	}
	sort.Slice(tipovi, func(i, j int) bool { return tipovi[i] < tipovi[j] })
	aktivne := make([]domain.RobnoStampaPoreskaStopaDto, 0, len(tipovi))
	for _, tip := range tipovi {
		aktivne = append(aktivne, poTipu[tip])
	}
	return aktivne
}

// robnoStampaInterniPrenosPorez is one porez (tip of the porez and poreska oznaka) of a printed interni
// prenos while it is summed (the legacy STPorezi).
type robnoStampaInterniPrenosPorez struct {
	tip, po                           int64
	tarifa                            string
	pp, stopa, osnovica, porez, iznos float64
}

// robnoStampaInterniPrenosView builds the printed interni prenos proizvodnje of the stavke of one robni
// dokument like the legacy body. Per stavka:
//   - vrednost = količina * cena, taksa = količina * the taksa of the artikal (rcene.itaksa);
//   - rabat = (vrednost - taksa) * rpro.vra % (for an artikal of the model "D", duvan: rpro.vra *
//     količina), nabavna vrednost = vrednost - taksa - rabat, nabavna cena = nabavna vrednost / količina;
//   - maloprodajna vrednost sa porezom = količina * rpro.mcenap; the porez (legacy porezi_obracun): for
//     every porez in force of the poreska oznaka of the stavka on the date of the document, of the
//     osnovica maloprodajna vrednost - taksa at the stopa calculated out of the iznos (stopa / (sum of
//     the stope + 100) * 100); for duvan of the osnovica of the akciza (rcene.pakc * količina) with the
//     akciza (rcene.iakc %), the marža (rcene.vma + rcene.mma per unit) less the rabat, at the stopa; no
//     porez for the poreska oznaka 0;
//   - maloprodajna vrednost bez poreza = maloprodajna vrednost - porez - taksa, the cene per unit, the
//     marža = maloprodajna vrednost bez poreza - nabavna vrednost and the % marže of the nabavna vrednost
//     (for duvan the marža per unit).
//
// The totals add the vrednosti of the stavke; the porezi are summed per porez and printed with the
// stopa of the porez, the porez, the maloprodajni iznos and the poreska osnovica (the maloprodajni iznos
// without the porez). The interna zaključnica (interna) differs as described at
// GetStampaInternaZakljucnica.
func robnoStampaInterniPrenosView(rows []domain.RobnoStampaInterniPrenosRowDto, stope []domain.RobnoStampaPoreskaStopaDto, interna bool) domain.RobnoStampaInterniPrenosView {
	dok := rows[0]
	prenos := domain.RobnoStampaInterniPrenosView{
		RdokID:             dok.RdokID,
		Naslov:             dok.VrdOpis + ":",
		BrojDokumenta:      fmt.Sprintf("%d-%d", dok.Vrd, dok.Dokum),
		DatumDokumenta:     common.FormatNullTime(dok.Dadok, common.DateLayout),
		Nalog:              fmt.Sprintf("%s/%d", dok.Tipdok, dok.Nalog),
		MagacinIzlaz:       dok.MagOznaka,
		MagacinIzlazAdresa: dok.MagAdresa,
		ObjekatPrijem:      strings.TrimSpace(dok.Pkto + " " + dok.Pana),
		ObjekatPrijemNaziv: dok.PrijemNaziv,
		Interna:            interna,
	}
	// The stopa of a porez out of the iznos is rounded to 8 decimals and the cena has 3 (the interna
	// zaključnica: 4 and 2).
	decimaleStope, decimaleCene := 1e8, 3
	if interna {
		decimaleStope, decimaleCene = 1e4, 2
	}
	if !dok.PrijemNadjen {
		prenos.ObjekatPrijemNaziv = i18n.GetInstance().Label("Ne postoji naziv!")
	}
	f2 := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 2) }
	porezi := map[[2]int64]*robnoStampaInterniPrenosPorez{}
	var redPoreza [][2]int64
	var uVre, uMrab, uMvre, uMar, uMpibp, uTak, uPor, uVmc float64
	for i, st := range rows {
		duvan := strings.EqualFold(strings.TrimSpace(st.Model), "D")
		taksa := robnoStampaFakturaRound(st.Kolic * st.Itaksa)
		vrednost := robnoStampaFakturaRound(st.Kolic * st.Cena)
		var rabat float64
		if duvan {
			rabat = robnoStampaFakturaRound(st.Vra * st.Kolic)
		} else {
			rabat = robnoStampaFakturaRound(st.Vra * (vrednost - taksa) / 100)
		}
		nabavna := robnoStampaFakturaRound(vrednost - taksa - rabat)
		if interna && st.Gru == 777 {
			nabavna = robnoStampaFakturaRound(vrednost - rabat)
		}
		mcVrednost := robnoStampaFakturaRound(st.Kolic * st.Mcenap)
		// porezi_obracun
		var osnovica, miz float64
		if duvan {
			akcOsnovica := robnoStampaFakturaRound(st.Pakc * st.Kolic)
			akciza := robnoStampaFakturaRound(akcOsnovica * st.Iakc / 100)
			osnovica = akciza + akcOsnovica + (st.Vma+st.Mma)*st.Kolic - rabat
		} else {
			osnovica = robnoStampaFakturaRound(mcVrednost - taksa)
			miz = osnovica
		}
		aktivne := robnoStampaAktivnePoreskeStope(stope, st.Po, dok.Dadok)
		var zbir float64
		for _, a := range aktivne {
			zbir += a.Pp
		}
		var porez float64
		tarifa := ""
		for _, a := range aktivne {
			kljuc := [2]int64{a.Tip, st.Po}
			pz := porezi[kljuc]
			if pz == nil {
				pz = &robnoStampaInterniPrenosPorez{tip: a.Tip, po: st.Po, tarifa: a.Pt, pp: a.Pp}
				porezi[kljuc] = pz
				redPoreza = append(redPoreza, kljuc)
			}
			pz.stopa = a.Pp
			if miz != 0 {
				pz.stopa = math.Round(a.Pp/(zbir+100)*100*decimaleStope) / decimaleStope
			}
			por := robnoStampaFakturaRound(osnovica * pz.stopa / 100)
			pz.porez = robnoStampaFakturaRound(pz.porez + por)
			pz.osnovica = robnoStampaFakturaRound(pz.osnovica + osnovica)
			pz.iznos = robnoStampaFakturaRound(pz.iznos + miz)
			tarifa = pz.tarifa
			porez = robnoStampaFakturaRound(porez + por)
		}
		if st.Po == 0 {
			porez = 0
		}
		mcBezPoreza := robnoStampaFakturaRound(mcVrednost - (porez + taksa))
		marza := robnoStampaFakturaRound(mcBezPoreza - nabavna)
		var marzaProc, nabavnaCena, mcCena, mcBezPorezaCena float64
		if st.Kolic != 0 {
			nabavnaCena = robnoStampaFakturaRound(nabavna / st.Kolic)
			mcCena = robnoStampaFakturaRound(mcVrednost / st.Kolic)
			mcBezPorezaCena = robnoStampaFakturaRound(mcBezPoreza / st.Kolic)
		}
		marzaTekst := ""
		switch {
		case duvan && st.Kolic != 0:
			marzaProc = marza / st.Kolic
			marzaTekst = f2(marzaProc)
		case nabavna != 0:
			marzaProc = robnoStampaFakturaRound(marza / nabavna * 100)
			marzaTekst = f2(marzaProc) + "%"
		default:
			marzaTekst = f2(0) + "%"
		}
		naziv, barkod := st.Naziv, ""
		if interna {
			naziv = robnoStampaInternaZakljucnicaNaziv(st)
			barkod = st.Barkod
		}
		prenos.Stavke = append(prenos.Stavke, domain.RobnoStampaInterniPrenosStavka{
			Rbr:         fmt.Sprintf("%d", i+1),
			Sifra:       robnoStampaFakturaSifra(st.Sifra),
			Naziv:       naziv,
			Barkod:      barkod,
			Jm:          st.Jm,
			Kolicina:    common.FormatNumberWithSystemLocale(st.Kolic, 3),
			Cena:        common.FormatNumberWithSystemLocale(st.Cena, decimaleCene),
			Rabat:       f2(st.Vra) + "%",
			NabavnaCena: f2(nabavnaCena),
			Marza:       marzaTekst,
			McBezPoreza: f2(mcBezPorezaCena),
			// The legacy ITEM_taksa (xRtak) is never set: the taksa per unit prints 0.
			Taksa:  f2(0),
			Tarifa: tarifa,
			Stopa:  f2(robnoStampaPoreskaStopa(stope, st.Po, st.StavkaDadok)) + "%",
			Mc:     f2(mcCena),

			Vrednost:            f2(vrednost),
			RabatIznos:          f2(rabat),
			NabavnaVrednost:     f2(nabavna),
			MarzaVrednost:       f2(marza),
			McBezPorezaVrednost: f2(mcBezPoreza),
			TaksaVrednost:       f2(taksa),
			PorezVrednost:       f2(porez),
			McVrednost:          f2(mcVrednost),
		})
		uVre += vrednost
		uMrab += rabat
		uMvre += nabavna
		uMar += marza
		uMpibp += mcBezPoreza
		uTak += taksa
		uPor += porez
		uVmc += mcVrednost
	}
	prenos.Ukupno = domain.RobnoStampaInterniPrenosStavka{
		Vrednost:            f2(uVre),
		RabatIznos:          f2(uMrab),
		NabavnaVrednost:     f2(uMvre),
		MarzaVrednost:       f2(uMar),
		McBezPorezaVrednost: f2(uMpibp),
		TaksaVrednost:       f2(uTak),
		PorezVrednost:       f2(uPor),
		McVrednost:          f2(uVmc),
	}
	// The porezi in the order of the legacy key (tip + poreska oznaka).
	sort.SliceStable(redPoreza, func(i, j int) bool {
		return redPoreza[i][0]+redPoreza[i][1] < redPoreza[j][0]+redPoreza[j][1]
	})
	var uOsn, uPdv, uIzn float64
	for _, kljuc := range redPoreza {
		pz := porezi[kljuc]
		// The poreska osnovica is the maloprodajni iznos without the porez (for duvan, without a
		// maloprodajni iznos, the osnovica of the porez).
		osnovica := robnoStampaFakturaRound(pz.iznos - pz.porez)
		if pz.iznos == 0 {
			osnovica = pz.osnovica
		}
		prenos.Porezi = append(prenos.Porezi, domain.RobnoStampaInterniPrenosPorez{
			Sp:                fmt.Sprintf("%d", pz.po),
			Tarifa:            pz.tarifa,
			Osnovica:          f2(osnovica),
			Stopa:             f2(pz.pp) + "%",
			Pdv:               f2(pz.porez),
			MaloprodajniIznos: f2(pz.iznos),
		})
		uOsn += osnovica
		uPdv += pz.porez
		uIzn += pz.iznos
	}
	prenos.UkupnoPorezi = domain.RobnoStampaInterniPrenosPorez{
		Osnovica:          f2(uOsn),
		Pdv:               f2(uPdv),
		MaloprodajniIznos: f2(uIzn),
	}
	return prenos
}

//
// Štampa nivelacije maloprodaje: the report RobnoStampaNivelacijaMaloprodaje (the group MNI of the
// vrste dokumenta, the legacy RPT_NIVELACIJA_MALOPRODAJE with its query QRY_RPRO_ZADOK)
//

// GetStampaNivelacijaMaloprodaje returns the nivelacije maloprodaje of the selection, ready to print (the
// legacy RPT_NIVELACIJA_MALOPRODAJE): one view per robni dokument with its stavke in the order of the
// nalog, the document and the rbr, calculated like the legacy report (see
// robnoStampaNivelacijaMPView), the totals and the porezi per nova poreska oznaka, and the izdavalac
// (fvr) of the print. The stara MP cena of a stavka is rpro.cena, the nova rpro.mcenap; the stara
// poreska oznaka is rpro.po, the nova rpro.dani. The prodavnica is the fkpl of rdok.pkto and rdok.pana.
func (s *RobnoDokumentaResource) GetStampaNivelacijaMaloprodaje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaNivelacijaMPView, domain.RobnoStampaFakturaFirmaDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, domain.RobnoStampaFakturaFirmaDto{}, fmt.Errorf("no user session found")
	}
	firma, err := s.stampaFirma(ctx, userSession)
	if err != nil {
		return nil, firma, err
	}
	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0) as dokum,
			rdok.dadok,
			coalesce(rdok.pkto, '') as pkto,
			coalesce(rdok.pana, '') as pana,
			coalesce(prod.naziv, '') as prijemnaziv,
			prod.naziv is not null as prijemnadjen,
			coalesce(rpro.rbr, 0) as rbr,
			coalesce(nullif(rpro.naz1, ''), rsif.naziv, '') as naziv,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.cena, 0) as cena,
			coalesce(rpro.mcenap, 0) as mcenap,
			coalesce(rpro.po, 0) as po,
			coalesce(rpro.dani, 0) as dani
		from rpro`, true)
	qb.AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid")
	qb.AddJoin("left join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(`left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.pkto and f.sifra = rdok.pana
		order by f.vkonta desc, f.idfkpl limit 1) prod on true`)
	robnoStampaDokumentaUsloviTabele(qb, userSession, params, "rdok")
	qb.AddOrderBy("rdok.rnalid, rdok.rdokid, rpro.rbr")
	sqlQuery, args := qb.Build()
	rows, err := s.stampaInterniPrenosRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, firma, err
	}
	if len(*rows) == 0 {
		return []domain.RobnoStampaNivelacijaMPView{}, firma, nil
	}
	var oznake []int64
	for _, r := range *rows {
		oznake = append(oznake, r.Po, r.Dani)
	}
	stope, err := s.stampaPoreskeStope(ctx, oznake)
	if err != nil {
		return nil, firma, err
	}
	nivelacije := []domain.RobnoStampaNivelacijaMPView{}
	for i := 0; i < len(*rows); {
		j := i
		for j < len(*rows) && (*rows)[j].RdokID == (*rows)[i].RdokID {
			j++
		}
		nivelacije = append(nivelacije, robnoStampaNivelacijaMPView(firma, (*rows)[i:j], stope))
		i = j
	}
	return nivelacije, firma, nil
}

// robnoStampaNivelacijaMPView builds the printed nivelacija maloprodaje of the stavke of one robni
// dokument like the legacy report: the stopa of the stara (rpro.po) and of the nova poreska oznaka
// (rpro.dani) on the date of the document (VratiPorez2) with the tarifa of the oznaka, and the stopa of
// the porez out of the MP cena (stopa * 100 / (stopa + 100), 2 decimals). Per stavka: the vrednosti are
// the količina times the cene (rounded to the para), the razlika is the nova less the stara MP cena, the
// stari and the novi porez per unit are the MP cena times its preračunata stopa (rounded to the para) and
// their vrednosti the količina times them, the RUC per unit is the razlika less the razlika poreza and its
// vrednost the količina times it. The porezi are summed per nova poreska oznaka with the preračunata
// stopa of its first stavka.
func robnoStampaNivelacijaMPView(firma domain.RobnoStampaFakturaFirmaDto, rows []domain.RobnoStampaInterniPrenosRowDto, stope []domain.RobnoStampaPoreskaStopaDto) domain.RobnoStampaNivelacijaMPView {
	lbl := i18n.GetInstance().Label
	dok := rows[0]
	niv := domain.RobnoStampaNivelacijaMPView{
		FirmaNaziv:      firma.Naziv,
		FirmaAdresa:     strings.TrimSpace(firma.Pobro + " " + firma.Mesto + " " + firma.Adresa),
		FirmaPib:        lbl("PIB") + " : " + firma.Pib,
		FirmaMbr:        lbl("Mat. broj") + " : " + firma.Matbr,
		FirmaRegbr:      lbl("Registarski broj") + " : " + firma.Regbr,
		BrojDokumenta:   fmt.Sprintf("%d-%d", dok.Vrd, dok.Dokum),
		DatumNivelacije: common.FormatNullTime(dok.Dadok, common.DateLayout),
		DatumStampe:     time.Now().Format(common.DateLayout + " 15:04"),
	}
	if dok.PrijemNadjen {
		niv.Prodavnica = dok.Pkto + "-" + dok.Pana
		niv.ProdavnicaNaziv = dok.PrijemNaziv
	}
	f2 := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 2) }
	r2 := robnoStampaFakturaRound
	// tarifa returns the tarifa of a poreska oznaka (the legacy HReadSeekFirst(RPOR, PO, ...)).
	tarifa := func(po int64) string {
		for _, st := range stope {
			if st.Po == po {
				return st.Pt
			}
		}
		return ""
	}
	type porez struct {
		po                                              int64
		tarifa                                          string
		staraOsn, novaOsn, stopa, stariPorez, noviPorez float64
	}
	porezi := map[int64]*porez{}
	var redPoreza []int64
	var uStara, uNova, uRazlika, uStariPorez, uNoviPorez, uRazlikaPoreza, uRuc float64
	for _, st := range rows {
		staraStopa := robnoStampaPoreskaStopa(stope, st.Po, dok.Dadok)
		staraPrStopa := r2(staraStopa * 100 / (staraStopa + 100))
		novaStopa := robnoStampaPoreskaStopa(stope, st.Dani, dok.Dadok)
		novaPrStopa := r2(novaStopa * 100 / (novaStopa + 100))
		staraTarifa, novaTarifa := tarifa(st.Po), tarifa(st.Dani)

		staraVrednost := r2(st.Kolic * st.Cena)
		novaVrednost := r2(st.Kolic * st.Mcenap)
		razlika := st.Mcenap - st.Cena
		razlikaVrednost := r2(st.Mcenap*st.Kolic - st.Cena*st.Kolic)
		stariPorez := r2(st.Cena * staraPrStopa / 100)
		noviPorez := r2(st.Mcenap * novaPrStopa / 100)
		razlikaPoreza := noviPorez - stariPorez
		stariPorezVrednost := r2(st.Kolic * stariPorez)
		noviPorezVrednost := r2(st.Kolic * noviPorez)
		razlikaPorezaVrednost := noviPorezVrednost - stariPorezVrednost
		ruc := razlika - razlikaPoreza
		rucVrednost := st.Kolic * ruc

		niv.Stavke = append(niv.Stavke, domain.RobnoStampaNivelacijaMPStavka{
			Rbr:      fmt.Sprintf("%d", st.Rbr),
			Naziv:    st.Naziv,
			Kolicina: common.FormatNumberWithSystemLocale(st.Kolic, 3),

			StaraCena:       f2(st.Cena),
			StaraVrednost:   f2(staraVrednost),
			NovaCena:        f2(st.Mcenap),
			NovaVrednost:    f2(novaVrednost),
			Razlika:         f2(razlika),
			RazlikaVrednost: f2(razlikaVrednost),

			StaraOznaka:        fmt.Sprintf("%d", st.Po),
			StaraStopa:         strings.TrimSpace(robnoStampaFakturaStopa(staraStopa) + " " + staraTarifa),
			StariPorez:         f2(stariPorez),
			StariPorezVrednost: f2(stariPorezVrednost),
			NovaOznaka:         fmt.Sprintf("%d", st.Dani),
			NovaStopa:          strings.TrimSpace(robnoStampaFakturaStopa(novaStopa) + " " + novaTarifa),
			NoviPorez:          f2(noviPorez),
			NoviPorezVrednost:  f2(noviPorezVrednost),

			RazlikaPoreza:         f2(razlikaPoreza),
			RazlikaPorezaVrednost: f2(razlikaPorezaVrednost),
			Ruc:                   f2(ruc),
			RucVrednost:           f2(rucVrednost),
		})
		uStara += staraVrednost
		uNova += novaVrednost
		uRazlika += razlikaVrednost
		uStariPorez += stariPorezVrednost
		uNoviPorez += noviPorezVrednost
		uRazlikaPoreza += razlikaPorezaVrednost
		uRuc += rucVrednost

		pz := porezi[st.Dani]
		if pz == nil {
			pz = &porez{po: st.Dani, tarifa: novaTarifa, stopa: novaPrStopa}
			porezi[st.Dani] = pz
			redPoreza = append(redPoreza, st.Dani)
		}
		pz.staraOsn += staraVrednost
		pz.novaOsn += novaVrednost
		pz.stariPorez += stariPorezVrednost
		pz.noviPorez += noviPorezVrednost
	}
	niv.Ukupno = domain.RobnoStampaNivelacijaMPStavka{
		StaraVrednost:         f2(uStara),
		NovaVrednost:          f2(uNova),
		RazlikaVrednost:       f2(uRazlika),
		StariPorezVrednost:    f2(uStariPorez),
		NoviPorezVrednost:     f2(uNoviPorez),
		RazlikaPorezaVrednost: f2(uRazlikaPoreza),
		RucVrednost:           f2(uRuc),
	}
	// The porezi in the order of the poreska oznaka (the key of the legacy associative array).
	sort.Slice(redPoreza, func(i, j int) bool { return redPoreza[i] < redPoreza[j] })
	for _, po := range redPoreza {
		pz := porezi[po]
		niv.Porezi = append(niv.Porezi, domain.RobnoStampaNivelacijaMPPorez{
			Sp:            fmt.Sprintf("%d", pz.po),
			Tarifa:        pz.tarifa,
			StaraOsnovica: f2(pz.staraOsn),
			NovaOsnovica:  f2(pz.novaOsn),
			Stopa:         f2(pz.stopa) + "%",
			StariPorez:    f2(pz.stariPorez),
			NoviPorez:     f2(pz.noviPorez),
			RazlikaPoreza: f2(pz.noviPorez - pz.stariPorez),
		})
	}
	return niv
}

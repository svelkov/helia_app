package robno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

// ---------------------------------------------------------------------------------------------------
// Kalkulacije veleprodaje (the group PLT, the legacy rob_Kalkulacije_Veleprodaja with its stavke
// rob_Kalkulacije_Veleprodaja_Stavke and its zavisni troškovi rob_Zavisni_Troskovi): a robni dokument of a dobavljač
// whose stavke enter the roba at its nabavna cena (the fakturna cena less the rabat with the zavisni troškovi) and set
// its veleprodajna and maloprodajna cena; with the porez and the avans paid to the dobavljač (rppo, appo) and the
// zavisni troškovi of other dobavljači (rztr with their porez in rppo). The automatic nivelacija of the legacy is not
// ported.
// ---------------------------------------------------------------------------------------------------

// The oznake of a zavisni trošak of a stavka (rpro.tz, rpro.tzi; the legacy RADIO_ZT and RADIO_ZTI): none, a procenat
// of the neto fakturna cena or a vrednost.
const (
	robKalBezTroska = "-"
	robKalProcenat  = "%"
	robKalVrednost  = "d"
)

// robKalOznaka is the oznaka of a zavisni trošak of the form: "%", "d", else "-".
func robKalOznaka(value string) string {
	switch strings.TrimSpace(value) {
	case robKalProcenat:
		return robKalProcenat
	case robKalVrednost:
		return robKalVrednost
	}
	return robKalBezTroska
}

// RobKalObracun are the values the nabavna cena of a stavka of a kalkulacija is computed from: the fakturna cena (in
// the domaća valuta) and the rabat, the zavisni trošak and the interni zavisni trošak with their oznake and the
// procenat of the external zavisni troškovi of the document.
type RobKalObracun struct {
	Fcena, Rab       float64
	Tz, Tzi          string
	Ztro, Ztrin      float64
	ProcenatTroskova float64
}

// robKalTrosak is a zavisni trošak of the neto fakturna cena: a procenat of it (rounded to 2 decimals), a vrednost,
// or none.
func robKalTrosak(neto float64, oznaka string, vrednost float64) float64 {
	switch oznaka {
	case robKalProcenat:
		return robZaokruzi(neto*vrednost/100, 2)
	case robKalVrednost:
		return vrednost
	}
	return 0
}

// RobKalNabavnaCena is the nabavna cena of a stavka of a kalkulacija (the legacy IzracunajVREDNOSTI and the exits of
// the fields of the cene): the fakturna cena less the rabat (the neto fakturna cena), with its zavisni trošak, the
// procenat of the external zavisni troškovi of the document on both (rounded to 2 decimals) and the interni zavisni
// trošak; rounded to 4 decimals.
func RobKalNabavnaCena(o RobKalObracun) float64 {
	neto := o.Fcena - o.Fcena*o.Rab/100
	ztr := robKalTrosak(neto, o.Tz, o.Ztro)
	ncena := neto + ztr + robZaokruzi((neto+ztr)*o.ProcenatTroskova/100, 2)
	ncena += robKalTrosak(neto, o.Tzi, o.Ztrin)
	return robZaokruzi(ncena, 4)
}

// RobKalVpCena is the veleprodajna cena: the nabavna cena with the veleprodajna marža, rounded to 2 decimals.
func RobKalVpCena(ncena, vma float64) float64 {
	return robZaokruzi(ncena+ncena*vma/100, 2)
}

// RobKalMcena is the maloprodajna cena without the porez: the nabavna cena with the maloprodajna marža, rounded to 2
// decimals (the legacy xMcena).
func RobKalMcena(ncena, mma float64) float64 {
	return robZaokruzi(ncena+ncena*mma/100, 2)
}

// RobKalMpCena is the maloprodajna cena with the PDV: the maloprodajna cena with the porez of its stopa (rounded to 2
// decimals).
func RobKalMpCena(ncena, mma, stopa float64) float64 {
	mcena := RobKalMcena(ncena, mma)
	return mcena + robZaokruzi(mcena*stopa/100, 2)
}

// robKalProcenatTroskova is the procenat of the external zavisni troškovi of a document (the legacy
// IzracunajEKSTZAVTROSAK): their sum by the fakturna vrednost robe (of the last trošak, like the legacy), rounded to 8
// decimals; 0 without it.
func robKalProcenatTroskova(ukupno, iznfak float64) float64 {
	if iznfak == 0 {
		return 0
	}
	return robZaokruzi(ukupno/iznfak*100, 8)
}

// RobKalStavkeTableRows returns the stavke of a kalkulacija as the rows of the grid of the stavke (the columns of
// GetKalStavkeTableFields, the legacy QRY_RRPO_DET: redni broj, konto, šifra, naziv, JM, količina, fakturna cena,
// rabat, OTK broj, serija, rok, nabavna cena, VP marža, VP cena, MP marža, MP cena sa PDV), by the redni broj.
func RobKalStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rok := ""
		if st.Roktr > 0 {
			rok = robnoStampaWinDevDatumFormat(int64(st.Roktr), common.DateLayout)
		}
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Fcena), common.FormatNumberWithSystemLocale(st.Rab, 2),
				st.Otk, st.Serija, rok, robFormatIznos(st.Ncena), common.FormatNumberWithSystemLocale(st.Vma, 2), robFormatIznos(st.Cena),
				common.FormatNumberWithSystemLocale(st.Mma, 2), robFormatIznos(st.Mcenap)},
		})
	}
	return rows
}

// RobKalPrikazStavkeTableRows returns the stavke of the selected kalkulacija of the tab "Pregled kalkulacija" (the
// columns of GetKalPrikazStavkeTableFields, the legacy TABLE_QRY_RPRO: šifra, naziv, JM, količina, fakturna cena,
// rabat, VP cena, VP marža, interni zavisni trošak, zavisni trošak, zavisni trošak drugih dobavljača), by the redni
// broj.
func RobKalPrikazStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Sifra), st.Naziv, st.Jm, common.FormatNumberWithSystemLocale(st.Kolic, 3),
				robFormatIznos(st.Fcena), common.FormatNumberWithSystemLocale(st.Rab, 2), robFormatIznos(st.Cena),
				common.FormatNumberWithSystemLocale(st.Vma, 2), common.FormatNumberWithSystemLocale(st.Ztrin, 6),
				common.FormatNumberWithSystemLocale(st.Ztro, 6), common.FormatNumberWithSystemLocale(st.Ztrof, 8)},
		})
	}
	return rows
}

// RobKalDokumentaPrikazTableRows returns the kalkulacije as the rows of the grid of the documents of the tab "Pregled
// kalkulacija" (the columns of GetKalDokumentaTableFields, the legacy TABLE_QRY_RDOK: vrsta and broj naloga, vrsta,
// broj and datum dokumenta, iznos, konto and šifra of the dobavljač, naziv partnera, opis, rok plaćanja, datum
// izvornog dokumenta, datum obrade, magacin, broj izvornog dokumenta).
func RobKalDokumentaPrikazTableRows(dokumenta []domain.RobDokumentPrikazRow) []domain.TableRow {
	broj := func(n int) string {
		if n == 0 {
			return ""
		}
		return strconv.Itoa(n)
	}
	rows := make([]domain.TableRow, 0, len(dokumenta))
	for _, d := range dokumenta {
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(d.RdokID, 10),
			Fields: []string{d.Tipdok, broj(d.Nalog), broj(d.Vrd), broj(d.Dokum), robFormatDatum(d.Dadok), robFormatIznos(d.Iznos), d.Fkto,
				d.Fana, d.Naziv, d.Opis, strconv.Itoa(d.Rok), robFormatDatum(d.Datiz), robFormatDatum(d.Datob), broj(d.Mag), d.Dokiz},
		})
	}
	return rows
}

// RobKalTroskoviTableRows returns the zavisni troškovi of a kalkulacija as the rows of their grid (the columns of
// GetKalTroskoviTableFields, the legacy TABLE_ZTROSAK: konto and šifra of the dobavljač, its naziv, broj, datum and rok
// of its document, the fakturna vrednost robe and the iznos of the trošak).
func RobKalTroskoviTableRows(troskovi []domain.RobKalTrosakRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(troskovi))
	for _, t := range troskovi {
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(t.RztrID, 10),
			Fields: []string{t.Fkto, t.Fana, t.Naziv, t.Dokdob, robFormatDatum(t.Dadok), strconv.Itoa(t.Rok), robFormatIznos(t.Iznfak),
				robFormatIznos(t.Iznos)},
		})
	}
	return rows
}

// ---------------------------------------------------------------------------------------------------
// The header (the legacy rob_Kalkulacije_Veleprodaja)
// ---------------------------------------------------------------------------------------------------

// GetRobKalDokument returns the document of the "Kalkulacije veleprodaje" screen: the document of the params like the
// one of the faktura (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one with the
// next broj, NoviBrojRDOK; the dobavljač is the kupac of the faktura) with the opis of its vrsta and
//   - a saved one: its EPP polje, "PDV se može odbiti", the narudžbenica, the plaćen porez and avans and whether it
//     has zavisni troškovi;
//   - a new one: the konto of the dobavljači of the robno (rvr.fdob), the EPP and "PDV se može odbiti" checked (the
//     legacy initialization of CBOX_EPP and CBOX_ODBITNI), no izvorni dokument, no org. jedinica and no mesto troška
//     ("-": the user chooses them; the legacy COMBO_ORGJED preselected the first ones).
//
// The entry of a zavisni trošak and of a stavka repeat the data of the document.
func (s *RobnoDokumentaResource) GetRobKalDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoKalDokument, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return domain.RobnoKalDokument{}, fmt.Errorf("no user session found")
	}
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoKalDokument{}, err
	}
	res := domain.RobnoKalDokument{Header: domain.RobnoKalHeaderView{RobnoFaktureHeaderView: dok.Header}, Poruka: dok.Poruka}
	h := &res.Header
	if dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(h.Vrd)); err != nil {
		return res, err
	} else if found {
		h.VrstaOpis = dokvrsta.Opis
	}
	fdob, err := s.robno.RvrFdob(ctx)
	if err != nil {
		return res, err
	}
	if h.RdokID > 0 {
		if err := s.robKalPodaciDokumenta(ctx, h); err != nil {
			return res, err
		}
	} else {
		h.KupacKonto, h.KupacSifra = strings.TrimSpace(fdob), ""
		h.Epp, h.Odbitni = true, true
		h.IzvorniDokDatum = time.Time{}
	}
	res.Trosak = domain.RobnoKalTrosak{RdokID: h.RdokID, KontoDobavljaca: strings.TrimSpace(fdob), DatumDokumenta: h.DatumDokum}
	if err := s.robKalStavka(ctx, &res, dok.Stavka); err != nil {
		return res, err
	}
	return res, nil
}

// robKalPodaciDokumenta fills the columns of a saved kalkulacija the faktura has not: the EPP polje, "PDV se može
// odbiti", the narudžbenica, the plaćen porez and avans and whether it has zavisni troškovi.
func (s *RobnoDokumentaResource) robKalPodaciDokumenta(ctx context.Context, h *domain.RobnoKalHeaderView) error {
	k, err := s.robno.RdokKalkulacijaPodaci(ctx, h.RdokID)
	if err != nil {
		return err
	}
	h.Epp = k.FseppID.Valid || strings.TrimSpace(k.EppPolje) != ""
	h.FseppID, h.FseppText, h.Odbitni = k.FseppID, k.FseppText, k.Odbitni
	if k.EdokID > 0 {
		h.VezaID, h.VezaBroj, h.VezaDatum = k.EdokID, strconv.Itoa(k.EdokDokum), robNullTime(k.EdokDadok)
	}
	if h.Porezi, err = s.robno.RppoDokumenta(ctx, h.RdokID); err != nil {
		return err
	}
	if h.Avansi, err = s.robno.AppoDokumenta(ctx, h.RdokID); err != nil {
		return err
	}
	troskovi, err := s.robno.RztrDokumenta(ctx, h.RdokID)
	if err != nil {
		return err
	}
	h.ImaTroskove = len(troskovi) > 0
	return nil
}

// robKalStavka fills the entry of a stavka: the first konto of the magacin, its tip zaliha, the podaci o dokumentu with
// its valuta and the procenat of the external zavisni troškovi of the document.
func (s *RobnoDokumentaResource) robKalStavka(ctx context.Context, res *domain.RobnoKalDokument, st domain.RobnoFaktureStavka) error {
	h := &res.Header
	stavka := &res.Stavka
	stavka.RdokID, stavka.Konto = h.RdokID, st.Konto
	stavka.Vrd, stavka.Brdok, stavka.DatumDokument = st.Vrd, st.Brdok, st.DatumDokumenta
	stavka.Valuta, stavka.Kurs, stavka.VaziZa = h.ValutaText, h.Kurs, h.VaziZa
	rows, err := s.robno.MagacinOpsti(ctx, h.MagaciniID, h.RnalID)
	if err != nil {
		return err
	}
	if rows != nil && len(*rows) > 0 {
		stavka.Tipzal = common.StringToInt((*rows)[0].Key)
	}
	if h.RdokID <= 0 {
		return nil
	}
	if k, found, err := s.GetRobStavkeKontekst(ctx, h.RdokID); err != nil {
		return err
	} else if found {
		stavka.Devizni = k.Devizni()
	}
	stavka.ProcenatTroskova, err = s.robKalProcenatDokumenta(ctx, h.RdokID)
	return err
}

// robKalProcenatDokumenta is the procenat of the external zavisni troškovi of a document (robKalProcenatTroskova).
func (s *RobnoDokumentaResource) robKalProcenatDokumenta(ctx context.Context, rdokID int64) (float64, error) {
	ukupno, iznfak, err := s.robno.RztrProcenat(ctx, rdokID)
	if err != nil {
		return 0, err
	}
	return robKalProcenatTroskova(ukupno, iznfak), nil
}

// robKalPorezRed is a row of a grid of the porezi as the form sends it.
type robKalPorezRed struct {
	osn, ppor, po, rac, datum string
}

// robKalRedovi returns the rows of a grid of the porezi of the form (the fields of a row by their index; a missing
// field of a row is empty).
func robKalRedovi(osn, ppor, po, rac, datum []string) []robKalPorezRed {
	at := func(values []string, i int) string {
		if i < len(values) {
			return strings.TrimSpace(values[i])
		}
		return ""
	}
	n := max(len(osn), len(ppor), len(po), len(rac), len(datum))
	redovi := make([]robKalPorezRed, 0, n)
	for i := range n {
		redovi = append(redovi, robKalPorezRed{osn: at(osn, i), ppor: at(ppor, i), po: at(po, i), rac: at(rac, i), datum: at(datum, i)})
	}
	return redovi
}

// robKalProveriPoreze checks the rows of a grid of the porezi (the field of the grid): a row with an osnovica (the
// legacy saves only those) has a poreska oznaka that exists; with avans also the broj and the datum of the avansni
// račun (the legacy "Morate uneti broj i datum avansnog racuna!").
func (s *RobnoDokumentaResource) robKalProveriPoreze(ctx context.Context, field string, redovi []robKalPorezRed, avans bool,
	add func(field, message string)) {
	for i, red := range redovi {
		if common.StringToFloat64(red.osn) == 0 {
			continue
		}
		po, err := strconv.Atoi(red.po)
		if err != nil || po <= 0 {
			add(field, fmt.Sprintf("Red %d: morate uneti poresku oznaku.", i+1))
			continue
		}
		if exists, err := s.robno.RporPostoji(ctx, po); err != nil {
			add(field, common.ErrMsgReadData)
		} else if !exists {
			add(field, fmt.Sprintf("Red %d: poreska oznaka ne postoji.", i+1))
		}
		if avans {
			if _, ok := robnoUnosDatum(red.datum); red.rac == "" || !ok {
				add(field, "Morate uneti broj i datum avansnog računa.")
			}
		}
	}
}

// robKalPorezi returns the rows of a grid of the porezi to save: the rows with an osnovica, with the stopa of their
// poreska oznaka and its rpor row on the datum (the legacy VratiPorez2 and the view of RPOR) and, for a storno, the
// osnovica and the porez negative.
func (s *RobnoDokumentaResource) robKalPorezi(ctx context.Context, redovi []robKalPorezRed, datum time.Time, storno bool) ([]domain.RobKalPorezUnos, error) {
	porezi := []domain.RobKalPorezUnos{}
	for _, red := range redovi {
		osn := common.StringToFloat64(red.osn)
		if osn == 0 {
			continue
		}
		p := domain.RobKalPorezUnos{Osnovica: osn, Porez: common.StringToFloat64(red.ppor), Po: common.StringToInt(red.po), Rac: red.rac}
		if storno {
			p.Osnovica, p.Porez = -math.Abs(p.Osnovica), -math.Abs(p.Porez)
		}
		if d, ok := robnoUnosDatum(red.datum); ok {
			p.Datum = sql.NullTime{Time: d, Valid: true}
		}
		var err error
		if p.Stopa, err = s.robno.RporStopa(ctx, nil, p.Po, datum); err != nil {
			return nil, err
		}
		if p.RporID, err = s.robno.RporIDNaDan(ctx, p.Po, datum); err != nil {
			return nil, err
		}
		porezi = append(porezi, p)
	}
	return porezi, nil
}

// robKalPopdv checks the POPDV polje of a kalkulacija or of a zavisni trošak (the field) of a PDV obveznik whose vrsta
// requires it (dokvrsta.flgpopdv): required, it exists, of an evidencija of the ulazni računi ("U") and of the tip 3;
// with tippdv (the tip PDV of the dobavljač, > 0) the polje is one of its tipovi PDV (only the header checks it).
func (s *RobnoDokumentaResource) robKalPopdv(ctx context.Context, field, polje string, tippdv int, add func(field, message string)) {
	polje = strings.ReplaceAll(strings.TrimSpace(polje), ",", ".")
	if polje == "" {
		add(field, "Morate uneti POPDV polje.")
		return
	}
	p, found, err := s.robno.PopdvPoljeUlaz(ctx, polje)
	switch {
	case err != nil:
		add(field, common.ErrMsgReadData)
	case !found:
		add(field, "POPDV polje ne postoji.")
	case p.Vktip != "" && p.Vktip != "U":
		add(field, "Uneto polje nije u skladu sa vrstom dokumenta.")
	case p.Tip != 3:
		add(field, "Neodgovarajuće POPDV polje.")
	case tippdv > 0 && !strings.Contains(p.TippdvPartneri, strconv.Itoa(tippdv)):
		add(field, "Neodgovarajuće POPDV polje.")
	}
}

// robKalTraziPopdv tells whether the POPDV polje is checked: the firm is a PDV obveznik and the vrsta requires it.
func (s *RobnoDokumentaResource) robKalTraziPopdv(ctx context.Context, vrd int) (bool, error) {
	obveznik, err := s.robnoUnosPdvObveznik(ctx)
	if err != nil || !obveznik {
		return false, err
	}
	dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, vrd)
	return found && dokvrsta.FlgPopDV, err
}

// ValidateRobKal validates the header of a kalkulacija veleprodaje before it is saved, like the legacy ValidacijaRDOK:
//   - the vrsta exists; the datum dokumenta is valid and not after the datum naloga; the broj dokumenta is a whole
//     number; the broj and the datum of the izvorni dokument are required;
//   - the konto and the šifra of the dobavljač are required and it exists (fkpl); a given mesto nabavke is one of
//     the dobavljač;
//   - a given org. jedinica and mesto troška exist (the mesto troška of the org. jedinica; not required);
//   - the POPDV polje of a PDV obveznik whose vrsta requires it (robKalPopdv, with the tip PDV of the dobavljač);
//   - the rows of the plaćen porez and avans (robKalProveriPoreze).
//
// Without errors and without Potvrda, an izvorni dokument the dobavljač already has in another document of the vrsta
// gives the question of the legacy ("Izvorni dokument ... postoji! Nastavak unosa?"), pitanje.
func (s *RobnoDokumentaResource) ValidateRobKal(ctx context.Context, params domain.RobnoKalParams) (fieldErrors []domain.FieldError, pitanje string) {
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors, ""
	}
	vrd := common.StringToInt(params.Vrd)
	if _, found, err := s.robnoUnosDokvrsta(ctx, vrd); err != nil {
		add("vrd", common.ErrMsgReadData)
	} else if !found {
		add("vrd", robnoUnosMsgNePostoji)
	}
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Datum dokumenta ne može biti veći od datuma naloga.")
	}
	s.robOpstiBrojDokumenta(params.BrojDokum, add)
	if strings.TrimSpace(params.IzvorniDokum) == "" {
		add("brdokorg", "Morate uneti broj izvornog dokumenta.")
	}
	if _, ok := robnoUnosDatum(params.IzvorniDokDatum); !ok {
		add("datumizvora", "Morate uneti datum izvornog dokumenta.")
	}

	// The dobavljač and its mesto nabavke.
	konto, sifra := strings.TrimSpace(params.DobavljacKonto), strings.TrimSpace(params.DobavljacSifra)
	tippdv, naziv := 0, ""
	switch {
	case konto == "":
		add("kupackonto", common.ErrMsgObavezanPodatak)
	case sifra == "":
		add("kupacsifra", "Dobavljač sa takvom šifrom ne postoji.")
	default:
		if exists, err := s.robno.FkplAnalitikaPostoji(ctx, konto, sifra); err != nil {
			add("kupacsifra", common.ErrMsgReadData)
		} else if !exists {
			add("kupacsifra", "Dobavljač sa takvom šifrom ne postoji.")
		} else if dob, found, err := s.GetRobKupac(ctx, konto, sifra); err == nil && found {
			tippdv, naziv = dob.Tippdv, dob.Naziv
		}
	}
	if id, given, ok := robnoUnosID(params.MestoNabavke); !ok {
		add("fispid", robnoUnosMsgNePostoji)
	} else if given {
		if exists, err := s.robno.FispPostoji(ctx, id, konto, sifra); err != nil {
			add("fispid", common.ErrMsgReadData)
		} else if !exists {
			add("fispid", "Mesto nabavke sa unetom šifrom ne postoji.")
		}
	}

	// The org. jedinica and the mesto troška are not required (empty or "-": NULL); given, the checks of the opšti
	// dokument.
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)

	if trazi, err := s.robKalTraziPopdv(ctx, vrd); err != nil {
		add("popdv", common.ErrMsgReadData)
	} else if trazi {
		s.robKalPopdv(ctx, "popdv", params.Popdv, tippdv, add)
	}
	s.robKalProveriPoreze(ctx, "porezi", robKalRedovi(params.PorezOsn, params.PorezPpor, params.PorezPo, nil, nil), false, add)
	s.robKalProveriPoreze(ctx, "avansi", robKalRedovi(params.AvansOsn, params.AvansPpor, params.AvansPo, params.AvansRac, params.AvansDatum),
		true, add)
	if len(fieldErrors) > 0 || params.Potvrda {
		return fieldErrors, ""
	}
	if dokiz := strings.TrimSpace(params.IzvorniDokum); dokiz != "" {
		if postoji, err := s.robno.RdokIzvorniDobavljacaPostoji(ctx, konto, sifra, vrd, dokiz, params.RdokID); err != nil {
			add("brdokorg", common.ErrMsgReadData)
		} else if postoji {
			pitanje = fmt.Sprintf("Izvorni dokument %s za %s postoji! Nastavak unosa?", dokiz, strings.TrimSpace(naziv))
		}
	}
	return fieldErrors, pitanje
}

// SaveRobKal saves the header of a kalkulacija veleprodaje (the legacy "Sačuvaj" after ValidacijaRDOK: NoviSLOGRDOK)
// and returns its rdokid, in one transaction:
//   - the robni dokument with the fields of the header (SaveRobniDokument, a new one counted in the nalog): the
//     datumi, the opis, the izvorni dokument, the dobavljač and its mesto nabavke (fispid, mi), the valuta, the uslovi
//     nabavke, the org. jedinica and the mesto troška, the POPDV polje, the EPP polje (fseppid and its polje; without
//     the EPP 0 and ""), "PDV se može odbiti" and the narudžbenica (edokid);
//   - the plaćen porez (rppo of the document) and the plaćen avans (appo) are the rows of their grids (robKalPorezi:
//     with the vrsta, the broj and the datum of the document, the dobavljač, and of the porez the broj of the izvorni
//     dokument as the račun).
func (s *RobnoDokumentaResource) SaveRobKal(ctx context.Context, params domain.RobnoKalParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	fispID, valutaID, fseppID := id(params.MestoNabavke), id(params.Valuta), id(params.Fsepp)
	sif, err := s.robno.SifreSifarnika(ctx, fispID, 0, valutaID, 0, 0)
	if err != nil {
		return 0, err
	}
	eppPolje := ""
	if !params.Epp {
		fseppID = 0
	} else if fseppID > 0 {
		if eppPolje, _, err = s.robno.FseppPolje(ctx, fseppID); err != nil {
			return 0, err
		}
	}
	vrd, dokum := common.StringToInt(params.Vrd), common.StringToInt(params.BrojDokum)
	dokvrsta, _, err := s.robnoUnosDokvrsta(ctx, vrd)
	if err != nil {
		return 0, err
	}
	storno := strings.TrimSpace(dokvrsta.Predznak) == "-"
	datum, _ := robnoUnosDatum(params.DatumDokum)
	konto, sifra := strings.TrimSpace(params.DobavljacKonto), strings.TrimSpace(params.DobavljacSifra)
	porezi, err := s.robKalPorezi(ctx, robKalRedovi(params.PorezOsn, params.PorezPpor, params.PorezPo, nil, nil), datum, storno)
	if err != nil {
		return 0, err
	}
	avansi, err := s.robKalPorezi(ctx, robKalRedovi(params.AvansOsn, params.AvansPpor, params.AvansPo, params.AvansRac, params.AvansDatum),
		datum, storno)
	if err != nil {
		return 0, err
	}
	dok := domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        vrd,
		Dokum:      dokum,
		Polja: []domain.Fields{
			robPoljeDatum("dadok", params.DatumDokum),
			robPoljeText("opis", params.Opis),
			robPoljeInt("rok", params.Rok),
			robPoljeDatum("datiz", params.IzvorniDokDatum),
			robPoljeText("dokiz", params.IzvorniDokum),
			robPoljeText("fkto", konto),
			robPoljeText("fana", sifra),
			robPoljeText("otp", params.NacinDopremanja),
			robPoljeText("kamt", params.UsloviKamate),
			robPoljeText("pla", params.UsloviPlacanja),
			robPoljeText("foot", params.Tekst),
			robPoljeText("pornapomena", params.NapomenaPoresko),
			robPoljeID("fispid", fispID), robPoljeInt("mi", sif.Key),
			robPoljeID("idvalute", valutaID), robPoljeInt("sifval", sif.C2),
			robPoljeBroj("kurs", params.Kurs),
			robPoljeInt("vaziza", params.VaziZa),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
			robPoljeText("polje", strings.ReplaceAll(strings.TrimSpace(params.Popdv), ",", ".")),
			robPoljeText("epp_polje", eppPolje),
			{Name: "fseppid", Value: strconv.FormatInt(fseppID, 10)},
			{Name: "odbitni", Value: strconv.FormatBool(params.Odbitni)},
			{Name: "edokid", Value: strconv.FormatInt(id(params.Veza), 10)},
		},
		NoviPolja: []domain.Fields{{Name: "knjige_3", Value: "N"}},
	}
	return s.saveRobniDokumentSa(ctx, dok, func(ctx context.Context, tx db.Transaction, rdokID int64) error {
		d := repository.KalPorezDokument{Vrd: vrd, Dokum: dokum, Dadok: sql.NullTime{Time: datum, Valid: !datum.IsZero()},
			Konto: konto, Sifra: sifra, Rac: strings.TrimSpace(params.IzvorniDokum), RdokID: rdokID}
		if err := s.robno.RppoDokumentaBrisi(ctx, tx, rdokID); err != nil {
			return err
		}
		for _, p := range porezi {
			if err := s.robno.RppoDodaj(ctx, tx, d, p); err != nil {
				return err
			}
		}
		if err := s.robno.AppoDokumentaBrisi(ctx, tx, rdokID); err != nil {
			return err
		}
		for _, p := range avansi {
			if err := s.robno.AppoDodaj(ctx, tx, d, p); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetRobKalStopa returns the stopa of a poreska oznaka on a datum (the legacy VratiPorez2 of the exit of COL_PO of the
// grids of the porezi: the datum of the document or of the document of the dobavljač); 0 without it.
func (s *RobnoDokumentaResource) GetRobKalStopa(ctx context.Context, po int, datum string) (float64, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return 0, fmt.Errorf("no user session found")
	}
	d, ok := robnoUnosDatum(datum)
	if !ok || po <= 0 {
		return 0, nil
	}
	return s.robno.RporStopa(ctx, nil, po, d)
}

// GetRobKalZaliheBrisanja returns the artikli whose stanje would be negative after the deletion of a kalkulacija, as
// the lines of the warning of the legacy ("šifra-naziv ; Zaliha posle brisanja: stanje").
func (s *RobnoDokumentaResource) GetRobKalZaliheBrisanja(ctx context.Context, rdokID int64) ([]string, error) {
	zalihe, err := s.robno.RproZaliheBrisanja(ctx, rdokID)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(zalihe))
	for _, z := range zalihe {
		lines = append(lines, fmt.Sprintf("%d-%s ; Zaliha posle brisanja: %s", z.Sifra, strings.TrimSpace(z.Naziv),
			common.FormatNumberWithSystemLocale(z.Stanje, 3)))
	}
	return lines, nil
}

// DeleteRobKal deletes a kalkulacija veleprodaje with its stavke (the legacy BTN_Brisanje): a document printed on the
// fiskalni printer or posted is not deleted (poruka says why); else every stavka is taken back from its nalog, from
// the stanje and from the zalihe by lot (robKalBrisiStavkuTx) and deleted, the porez of its zavisni troškovi, its
// zavisni troškovi, its plaćen porez and avans and its rows of the knjiga evidencije prometa (kepu) are deleted, the
// document is deleted and the nalog counts one document less (robBrisiDokumentSa). After the deletion the stanje of
// every artikal gets its prosečna nabavna cena computed again (IzracunajNCENA).
func (s *RobnoDokumentaResource) DeleteRobKal(ctx context.Context, rdokID int64) (string, error) {
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return "Niste selektovali dokument za brisanje.", nil
	}
	stavke, err := s.GetRobStavkeDokumenta(ctx, rdokID)
	if err != nil {
		return "", err
	}
	poruka, err := s.robBrisiDokumentSa(ctx, rdokID, false, s.robKalBrisiStavkuTx,
		s.robno.RppoTroskovaDokumentaBrisi, s.robno.RztrDokumentaBrisi, s.robno.RppoDokumentaBrisi, s.robno.AppoDokumentaBrisi,
		s.robno.KepuDokumentaBrisi)
	if err != nil || poruka != "" {
		return poruka, err
	}
	return "", s.robKalProsecneCene(ctx, k, stavke)
}

// robKalProsecneCene gives the stanje of the artikli of the stavke (each artikal in its konto once) the prosečna nabavna
// cena computed again (robOpstiProsecnaCenaPonovo).
func (s *RobnoDokumentaResource) robKalProsecneCene(ctx context.Context, k domain.RobStavkeKontekst, stavke []domain.RobStavkaRow) error {
	done := map[string]bool{}
	for _, st := range stavke {
		key := strings.TrimSpace(st.Konto) + "|" + strconv.Itoa(st.Sifra)
		if done[key] {
			continue
		}
		done[key] = true
		if err := s.robOpstiProsecnaCenaPonovo(ctx, k, st.Konto, st.Sifra); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------
// The stavke (the legacy rob_Kalkulacije_Veleprodaja_Stavke)
// ---------------------------------------------------------------------------------------------------

// GetRobKalArtikal returns the artikal of a stavka of a kalkulacija with what the form of the stavke shows of it, like
// the legacy exit of EDT_SIFRA: the artikal by its šifra (or by its barkod for a value longer than 6 characters) with
// its naziv, JM, its konto (when it has one, else the konto of the form) and its tarifa with its stopa on the datum of
// the document; from its stanje in the konto of the magacin the stanje, the cene of the last kalkulacija (fakturna
// cena, rabat, nabavna cena, the marže, the VP and the MP cena) and the stara VP and MP cena. found is false when the
// artikal does not exist.
func (s *RobnoDokumentaResource) GetRobKalArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobKalArtikal, bool, error) {
	a := domain.RobKalArtikal{}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	if vrednost == "" {
		return a, false, nil
	}
	sif, found, err := s.robno.RsifArtikalKalkulacije(ctx, len(vrednost) > 6, vrednost, robNullTime(k.Dadok))
	if err != nil || !found {
		return a, false, err
	}
	a.Sifra, a.Naziv, a.Jm, a.Tarifa, a.Stopa = sif.Sifra, sif.Naziv, sif.Jm, sif.Po, sif.Stopa
	a.Konto = strings.TrimSpace(konto)
	if sif.Konto != "" {
		a.Konto = sif.Konto
	}
	c, stFound, err := s.robno.RstaKalkulacijaCene(ctx, k.MagaciniID, a.Konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if stFound {
		a.Stanje, a.Fcena, a.Rab, a.Ncena = c.Stanje, c.Fcena, c.Rab, c.Ncena
		a.Vma, a.Cena, a.Mma, a.Mcenap = c.Vma, c.Cena, c.Mma, c.Mcenap
		a.StaraCena, a.StaraMcenap = c.Cena, c.Mcenap
	}
	return a, true, nil
}

// robKalCeneStavke are the cene of a stavka of a kalkulacija computed from its form (robKalCene).
type robKalCeneStavke struct {
	kolicina, cenaval, fcena, rab   float64
	tz, tzi                         string
	ztro, ztrin                     float64
	ncena, vma, cena, mma, mcena    float64
	mcenap, procenatTroskova, stopa float64
}

// robKalCene computes the cene of a stavka from the form: the fakturna cena (of a devizni document in its valuta, the
// legacy exit of EDT_FCENA: converted by the kurs, 4 decimals), the zavisni troškovi without their oznaka 0 (the legacy
// BTN_Snimi1), the nabavna cena (RobKalNabavnaCena with the procenat of the external zavisni troškovi of the
// document), the VP and the MP cena as entered (computed from the marže when not entered) and the maloprodajna cena
// without the porez (RobKalMcena).
func robKalCene(k domain.RobStavkeKontekst, params domain.RobKalStavkaParams, procenatTroskova, stopa float64) robKalCeneStavke {
	c := robKalCeneStavke{
		kolicina: common.StringToFloat64(params.Kolic), cenaval: common.StringToFloat64(params.Fcena), rab: common.StringToFloat64(params.Rabat),
		tz: robKalOznaka(params.Tz), tzi: robKalOznaka(params.Tzi), ztro: common.StringToFloat64(params.Ztro),
		ztrin: common.StringToFloat64(params.Ztrin), vma: common.StringToFloat64(params.Vma), mma: common.StringToFloat64(params.Mma),
		procenatTroskova: procenatTroskova, stopa: stopa,
	}
	c.fcena = c.cenaval
	if k.Devizni() && k.Kurs != 0 && k.Vaziza > 0 {
		c.fcena = robZaokruzi(c.cenaval*k.Kurs/k.Vaziza, 4)
	}
	if c.tz == robKalBezTroska || c.ztro == 0 {
		c.tz, c.ztro = robKalBezTroska, 0
	}
	if c.tzi == robKalBezTroska || c.ztrin == 0 {
		c.tzi, c.ztrin = robKalBezTroska, 0
	}
	c.ncena = RobKalNabavnaCena(RobKalObracun{Fcena: c.fcena, Rab: c.rab, Tz: c.tz, Tzi: c.tzi, Ztro: c.ztro, Ztrin: c.ztrin,
		ProcenatTroskova: procenatTroskova})
	if c.cena = common.StringToFloat64(params.Cena); c.cena == 0 {
		c.cena = RobKalVpCena(c.ncena, c.vma)
	}
	if c.mcenap = common.StringToFloat64(params.Mcenap); c.mcenap == 0 {
		c.mcenap = RobKalMpCena(c.ncena, c.mma, stopa)
	}
	c.mcena = RobKalMcena(c.ncena, c.mma)
	return c
}

// ValidateRobKalStavka validates a stavka of a kalkulacija before it is saved, like the legacy Validacija of
// rob_Kalkulacije_Veleprodaja_Stavke: the document exists; the konto is a konto of the magacin; the artikal exists; the
// količina and the fakturna cena are greater than 0, the rabat is not negative, the nabavna cena (computed) is greater
// than 0; a zavisni trošak needs its oznaka. Without errors and without Potvrda, a stavka without the OTK broj in a
// magacin of the tip zaliha 2 gives the question of the legacy ("Niste izabrali OTK broj! Unos bez OTK broja?"),
// pitanje.
func (s *RobnoDokumentaResource) ValidateRobKalStavka(ctx context.Context, params domain.RobKalStavkaParams) (fieldErrors []domain.FieldError, pitanje string) {
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("sifraartikla", common.ErrMsgSessionNotFound)
		return fieldErrors, ""
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil {
		add("sifraartikla", common.ErrMsgReadData)
		return fieldErrors, ""
	}
	if !found {
		add("sifraartikla", "Dokument ne postoji.")
		return fieldErrors, ""
	}
	if konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", ""); konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	} else if exists, err := s.robno.MagkontoPostoji(ctx, k.MagaciniID, konto); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if !exists {
		add("kontostavke", "Izabrani konto nije definisan za magacin.")
	}
	sifra, err := strconv.Atoi(strings.TrimSpace(params.Sifra))
	naziv := ""
	if err != nil {
		add("sifraartikla", "Artikal sa takvom šifrom ne postoji.")
	} else if a, found, err := s.robno.RsifArtikalKalkulacije(ctx, false, strconv.Itoa(sifra), robNullTime(k.Dadok)); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", "Artikal sa takvom šifrom ne postoji.")
	} else {
		naziv = a.Naziv
	}
	if common.StringToFloat64(params.Kolic) <= 0 {
		add("kolicina", "Količina mora biti pozitivna vrednost.")
	}
	if common.StringToFloat64(params.Fcena) <= 0 {
		add("fakturnacena", "Fakturna cena mora biti pozitivna vrednost.")
	}
	if common.StringToFloat64(params.Rabat) < 0 {
		add("rabatstavke", "Rabat ne može biti negativan.")
	}
	if robKalOznaka(params.Tz) == robKalBezTroska && common.StringToFloat64(params.Ztro) > 0 {
		add("ztro", "Morate uneti tip zavisnog troška.")
	}
	if robKalOznaka(params.Tzi) == robKalBezTroska && common.StringToFloat64(params.Ztrin) > 0 {
		add("ztrin", "Morate uneti tip internog zavisnog troška.")
	}
	if len(fieldErrors) == 0 {
		procenat, err := s.robKalProcenatDokumenta(ctx, k.RdokID)
		if err != nil {
			add("fakturnacena", common.ErrMsgReadData)
		} else if robKalCene(k, params, procenat, 0).ncena <= 0 {
			add("fakturnacena", "Nabavna cena mora biti veća od 0.")
		}
	}
	if len(fieldErrors) > 0 || params.Potvrda {
		return fieldErrors, ""
	}
	if k.Tipzal == 2 && strings.TrimSpace(params.Otk) == "" {
		pitanje = fmt.Sprintf("Šifra: %d %s\nNiste izabrali OTK broj!\nUnos bez OTK broja?", sifra, strings.TrimSpace(naziv))
	}
	return fieldErrors, pitanje
}

// SaveRobKalStavka saves a stavka of a kalkulacija veleprodaje, a new one (RproID 0) or a changed one ("Izmeni"), like
// the legacy NoviSLOGRPRO, in one transaction (robKalSnimiStavku). After a change the stanje of the artikal of the old
// and of the new stavka gets its prosečna nabavna cena computed again (IzracunajNCENA).
func (s *RobnoDokumentaResource) SaveRobKalStavka(ctx context.Context, params domain.RobKalStavkaParams) (rproID int64, err error) {
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("dokument koji obrađujete ne postoji")
	}
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	rproID, stara, err := s.robKalSnimiStavku(ctx, tx, k, params)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	if params.RproID > 0 {
		err = s.robKalProsecneCene(ctx, k, []domain.RobStavkaRow{{Konto: stara.Konto, Sifra: stara.Sifra},
			{Konto: strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", ""), Sifra: common.StringToInt(params.Sifra)}})
	}
	return rproID, err
}

// robKalSnimiStavku saves a stavka of a kalkulacija in the transaction (see SaveRobKalStavka) and returns its id and,
// of a changed one, its old values:
//   - the cene of the stavka (robKalCene with the procenat of the external zavisni troškovi of the document and the
//     stopa of the tarifa of the artikal on the datum of the document);
//   - the iznos is the količina by the VP cena (in a magacin of the nabavna cena by the nabavna cena, which is then
//     the cena of the stavka), rounded to 2 decimals; for a storno the količina and the iznos are negative;
//   - a changed stavka is taken back first (robKalVratiStavku: the nalog, the stanje and the zalihe by lot);
//   - the stavka with the columns of its document, the artikal, the konto, the cene (fakturna cena and the cena in the
//     valuta, rabat, nabavna cena, marže, VP cena, maloprodajna cena with and without the porez, the zavisni troškovi
//     with their oznake and the procenat of the external ones, ztrof), the lot and, a new one, the next redni broj;
//   - the cene of the artikal (rcene: a new artikal of the cene is added);
//   - the nalog, the stanje (with the cene of the kalkulacija) and, in a magacin of the tip zaliha 1 or 2, the zalihe
//     by the lot get the iznos and the količina (AzurirajRNAL, AzurirajRSTA, AzurirajDRSTA); the stanje and the lot
//     get the VP cena, a new stavka the prosečna nabavna cena of the stanje with it (the legacy PROSNC), also of the
//     zalihe by lot of the artikal;
//   - the document gets the iznos and a new stavka counts in it.
func (s *RobnoDokumentaResource) robKalSnimiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst,
	params domain.RobKalStavkaParams) (int64, domain.RobKalStavkaDto, error) {
	stara := domain.RobKalStavkaDto{}
	if domain.GetSessionFromStdContext(ctx) == nil {
		return 0, stara, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	rsifID, po, found, err := s.robno.RsifIDPo(ctx, tx, sifra)
	if err != nil {
		return 0, stara, err
	}
	if !found {
		return 0, stara, fmt.Errorf("artikal ne postoji")
	}
	stopa := 0.0
	if k.Dadok.Valid {
		if stopa, err = s.robno.RporStopa(ctx, tx, po, k.Dadok.Time); err != nil {
			return 0, stara, err
		}
	}
	procenat, err := s.robKalProcenatDokumenta(ctx, k.RdokID)
	if err != nil {
		return 0, stara, err
	}
	c := robKalCene(k, params, procenat, stopa)
	cenaStavke := c.cena
	if k.NabavnaCena() {
		cenaStavke = c.ncena
	}
	kolicina := c.kolicina
	iznos := robZaokruzi(kolicina*cenaStavke, 2)
	if k.Storno() {
		kolicina, iznos = -math.Abs(kolicina), -math.Abs(iznos)
	}
	rok, _ := robnoUnosDatum(params.Rok)
	roktr := RobWinDevDatum(rok)
	otk, serija := strings.TrimSpace(params.Otk), strings.TrimSpace(params.Serija)

	// A changed stavka: its old values are taken back.
	if params.RproID > 0 {
		if stara, err = s.robno.RproKalkulacija(ctx, tx, params.RproID, k.RdokID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, stara, fmt.Errorf("stavka ne postoji")
			}
			return 0, stara, err
		}
		if err = s.robKalVratiStavku(ctx, tx, k, stara); err != nil {
			return 0, stara, err
		}
	}
	// The stanje before the new stavka, for its prosečna nabavna cena.
	ulaz, izlaz, prosnc, err := s.robno.RstaUlazIzlazProsnc(ctx, tx, k.MagaciniID, konto, sifra)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, stara, err
	}

	polja := []domain.Fields{
		{Name: "konto", Value: konto},
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		{Name: "po", Value: strconv.Itoa(po)},
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("cena", cenaStavke),
		robPoljeFloat("iznos", iznos),
		robPoljeFloat("fcena", c.fcena),
		robPoljeFloat("cenaval", c.cenaval),
		robPoljeFloat("rab", c.rab),
		robPoljeFloat("ncena", c.ncena),
		robPoljeFloat("vma", c.vma),
		robPoljeFloat("vpcena", c.cena),
		robPoljeFloat("mma", c.mma),
		robPoljeFloat("mcena", c.mcena),
		robPoljeFloat("mcenap", c.mcenap),
		robPoljeFloat("ztro", c.ztro),
		robPoljeFloat("ztrin", c.ztrin),
		robPoljeFloat("ztrof", c.procenatTroskova),
		{Name: "tz", Value: c.tz},
		{Name: "tzi", Value: c.tzi},
		robPoljeText("otk", otk),
		robPoljeText("serija", serija),
		{Name: "roktr", Value: strconv.FormatInt(roktr, 10)},
		{Name: "rproid1", Value: "0"},
	}
	rproID := params.RproID
	if rproID > 0 {
		if err := s.unos.Rpro.UpdateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", rproID, polja); err != nil {
			return 0, stara, err
		}
	} else {
		rbr, err := s.robno.RproSledeciRbr(ctx, tx, k.RdokID)
		if err != nil {
			return 0, stara, err
		}
		polja = append(polja,
			domain.Fields{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
			domain.Fields{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
			domain.Fields{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
			domain.Fields{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
			robPoljeNullTime("dadok", k.Dadok),
			domain.Fields{Name: "mag", Value: strconv.Itoa(k.Mag)},
			domain.Fields{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""}, domain.Fields{Name: "taxcat", Value: ""}, domain.Fields{Name: "taxexreco", Value: ""},
		)
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja); err != nil {
			return 0, stara, err
		}
	}

	if err := s.robKalCeneArtikla(ctx, tx, k, sifra, rsifID, c); err != nil {
		return 0, stara, err
	}

	// The nalog, the stanje and the zalihe by lot.
	if err := s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos-stara.Iznos); err != nil {
		return 0, stara, err
	}
	promena := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra, Kolicina: kolicina, Iznos: iznos,
		Cena: cenaStavke, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Vma: c.vma, Mma: c.mma, Ztro: c.ztro,
		Fcena: c.fcena, Mcena: c.mcena, Ncena: c.ncena, Rab: c.rab, Ztrin: c.ztrin, Mcenap: c.mcenap, Tz: c.tz, Tzi: c.tzi,
		Otk: otk, Serija: serija, Roktr: roktr, Brst: 1}
	if err := s.AzurirajRSTA(ctx, tx, promena); err != nil {
		return 0, stara, err
	}
	if k.Tipzal == 1 || k.Tipzal == 2 {
		if err := s.AzurirajDRSTA(ctx, tx, promena); err != nil {
			return 0, stara, err
		}
		if err := s.robno.DrstaLotPostaviVpcena(ctx, tx, konto, sifra, k.Mag, otk, serija, roktr, c.cena); err != nil {
			return 0, stara, err
		}
	}
	if err := s.robno.RstaPostaviVpcena(ctx, tx, k.MagaciniID, konto, sifra, c.cena); err != nil {
		return 0, stara, err
	}
	if params.RproID == 0 {
		// The prosečna nabavna cena with the new stavka (the legacy PROSNC: the količina as entered).
		stanje := ulaz - izlaz
		novaProsnc := c.ncena
		if c.kolicina+stanje != 0 {
			novaProsnc = robZaokruzi((c.kolicina*c.ncena+stanje*prosnc)/(c.kolicina+stanje), 2)
		}
		if err := s.robno.RstaPostaviProsnc(ctx, tx, k.MagaciniID, konto, sifra, novaProsnc); err != nil {
			return 0, stara, err
		}
		if err := s.robno.DrstaPostaviProsnc(ctx, tx, k.Mag, rsifID, novaProsnc); err != nil {
			return 0, stara, err
		}
	}

	// The document: its iznos, a new stavka counts in it.
	noviBrst := 0
	if params.RproID == 0 {
		noviBrst = 1
	}
	if err := s.robno.RdokDodajIznos(ctx, tx, k.RdokID, iznos-stara.Iznos, noviBrst); err != nil {
		return 0, stara, err
	}
	return rproID, stara, nil
}

// robKalCeneArtikla sets the cene of the artikal (rcene) from a stavka of a kalkulacija (the legacy RCENE of
// NoviSLOGRPRO): the marže, the zavisni troškovi with their oznake, the rabat, the nabavna cena, the VP cena and the
// maloprodajna cena without and with the porez; an artikal without its cene gets them (with the fakturna cena: the VP
// cena, of a devizni document in its valuta, and the planska cena 0).
func (s *RobnoDokumentaResource) robKalCeneArtikla(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, sifra int, rsifID int64,
	c robKalCeneStavke) error {
	cene, err := s.robno.RcenePostoji(ctx, tx, sifra)
	if err != nil {
		return err
	}
	if cene {
		return s.robno.RcenePostaviCene(ctx, tx, sifra, []repository.Izmena{
			{Kolona: "mma", Vrednost: c.mma}, {Kolona: "vma", Vrednost: c.vma}, {Kolona: "ztro", Vrednost: c.ztro},
			{Kolona: "mcena", Vrednost: c.mcena}, {Kolona: "mcenap", Vrednost: c.mcenap}, {Kolona: "ncena", Vrednost: c.ncena},
			{Kolona: "ztrin", Vrednost: c.ztrin}, {Kolona: "rab", Vrednost: c.rab}, {Kolona: "tz", Vrednost: c.tz},
			{Kolona: "tzi", Vrednost: c.tzi}, {Kolona: "cena", Vrednost: c.cena},
		})
	}
	fcena := c.cena
	if k.Devizni() && k.Vaziza > 0 {
		fcena = robZaokruzi(c.cena*k.Kurs/k.Vaziza, 2)
	}
	_, err = s.unos.Rcene.CreateWithTx(ctx, tx, &domain.RobRceneDto{}, "rceneid", []domain.Fields{
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		robPoljeFloat("mma", c.mma), robPoljeFloat("vma", c.vma), robPoljeFloat("ztro", c.ztro),
		robPoljeFloat("mcena", c.mcena), robPoljeFloat("mcenap", c.mcenap), robPoljeFloat("ncena", c.ncena),
		robPoljeFloat("ztrin", c.ztrin), robPoljeFloat("rab", c.rab), {Name: "tz", Value: c.tz}, {Name: "tzi", Value: c.tzi},
		robPoljeFloat("cena", c.cena), robPoljeFloat("fcena", fcena), {Name: "pcena", Value: "0"},
		{Name: "vra", Value: "0"}, {Name: "mra", Value: "0"},
	})
	return err
}

// robKalVratiStavku takes a stavka of a kalkulacija back in the transaction (the legacy BrisiRPRO): its iznos from the
// nalog (AzurirajRNAL), its količina and iznos from the stanje with the cene of the stavka (AzurirajRSTA, a stavka
// less) and, in a magacin of the tip zaliha 1 or 2, from the zalihe by its lot (AzurirajDRSTA).
func (s *RobnoDokumentaResource) robKalVratiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, st domain.RobKalStavkaDto) error {
	if err := s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -st.Iznos); err != nil {
		return err
	}
	promena := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: strings.TrimSpace(st.Konto), Sifra: st.Sifra, Kolicina: -st.Kolicina,
		Iznos: -st.Iznos, Cena: st.Cena, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Vma: st.Vma, Mma: st.Mma,
		Ztro: st.Ztro, Fcena: st.Fcena, Mcena: st.Mcena, Ncena: st.Ncena, Rab: st.Rab, Ztrin: st.Ztrin, Mcenap: st.Mcenap,
		Tz: st.Tz, Tzi: st.Tzi, Otk: strings.TrimSpace(st.Otk), Serija: strings.TrimSpace(st.Serija), Roktr: int64(st.Roktr), Brst: -1}
	if err := s.AzurirajRSTA(ctx, tx, promena); err != nil {
		return err
	}
	if k.Tipzal == 1 || k.Tipzal == 2 {
		return s.AzurirajDRSTA(ctx, tx, promena)
	}
	return nil
}

// robKalBrisiStavkuTx deletes a stavka of a kalkulacija in the transaction of the caller and returns its document (the
// legacy BrisiRPRO and the stavke of BTN_Brisanje): the stavka is deleted and taken back (robKalVratiStavku), the
// document loses its iznos and counts one stavka less.
func (s *RobnoDokumentaResource) robKalBrisiStavkuTx(ctx context.Context, tx db.Transaction, rproID int64) (int64, error) {
	rdokID, err := s.robno.RproRdokID(ctx, tx, rproID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("stavka ne postoji")
		}
		return 0, err
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("dokument ne postoji")
	}
	st, err := s.robno.RproKalkulacija(ctx, tx, rproID, rdokID)
	if err != nil {
		return 0, err
	}
	if err := s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return 0, err
	}
	if err := s.robKalVratiStavku(ctx, tx, k, st); err != nil {
		return 0, err
	}
	return rdokID, s.robno.RdokOduzmiStavku(ctx, tx, rdokID, st.Iznos)
}

// DeleteRobKalStavka deletes a stavka of a kalkulacija (the legacy BrisiRPRO) in one transaction and returns its
// document: the stavka is taken back and deleted (robKalBrisiStavkuTx), the stavke after it get the redni brojevi
// again (rbr and brst). After the deletion the stanje of the artikal gets its prosečna nabavna cena computed again
// (IzracunajNCENA).
func (s *RobnoDokumentaResource) DeleteRobKalStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	rdokID, err = s.robno.RproRdokID(ctx, tx, rproID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("stavka ne postoji")
		}
		return 0, err
	}
	st, err := s.robno.RproKalkulacija(ctx, tx, rproID, rdokID)
	if err != nil {
		return 0, err
	}
	if _, err = s.robKalBrisiStavkuTx(ctx, tx, rproID); err != nil {
		return 0, err
	}
	if err = s.robno.RproPrenumerisiSaBrst(ctx, tx, rdokID, st.Rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return rdokID, err
	}
	return rdokID, s.robKalProsecneCene(ctx, k, []domain.RobStavkaRow{{Konto: st.Konto, Sifra: st.Sifra}})
}

// GetRobKalStavkaForma returns a saved stavka of a kalkulacija for the form of the stavke (the legacy BTN_Izmeni): its
// konto, artikal (naziv, JM, tarifa with its stopa, stanje), količina (as entered: a storno has it saved negative),
// the fakturna cena as entered (of a devizni document in its valuta), the rabat, the zavisni troškovi with their
// oznake, the marže, the cene and the lot. found is false without it.
func (s *RobnoDokumentaResource) GetRobKalStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobKalStavkaForma, bool, error) {
	f := domain.RobKalStavkaForma{}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return f, false, err
	}
	st, err := s.robno.RproKalkulacija(ctx, nil, rproID, rdokID)
	if errors.Is(err, sql.ErrNoRows) {
		return f, false, nil
	}
	if err != nil {
		return f, false, err
	}
	f.RproID, f.Sifra, f.Konto = st.RproID, st.Sifra, strings.TrimSpace(st.Konto)
	f.Kolicina, f.Fcena, f.Cenaval, f.Rab, f.Ncena = math.Abs(st.Kolicina), st.Fcena, st.Cenaval, st.Rab, st.Ncena
	if !k.Devizni() || f.Cenaval == 0 {
		f.Cenaval = st.Fcena
	}
	f.Tz, f.Ztro, f.Tzi, f.Ztrin = robKalOznaka(st.Tz), st.Ztro, robKalOznaka(st.Tzi), st.Ztrin
	f.Vma, f.Mma, f.Mcenap = st.Vma, st.Mma, st.Mcenap
	f.Cena = st.Cena
	if k.NabavnaCena() {
		// In a magacin of the nabavna cena the cena of the stavka is the nabavna cena: the VP cena is the one of the
		// stanje of the stavka.
		f.Cena = RobKalVpCena(st.Ncena, st.Vma)
	}
	f.Otk, f.Serija = strings.TrimSpace(st.Otk), strings.TrimSpace(st.Serija)
	if st.Roktr > 0 {
		f.Rok = robnoStampaWinDevDatumFormat(int64(st.Roktr), common.HtmlLayout)
	}
	a, aFound, err := s.GetRobKalArtikal(ctx, rdokID, strconv.Itoa(st.Sifra), f.Konto)
	if err != nil {
		return f, true, err
	}
	if aFound {
		f.Naziv, f.Jm, f.Tarifa, f.Stopa, f.Stanje = a.Naziv, a.Jm, a.Tarifa, a.Stopa, a.Stanje
	}
	return f, true, nil
}

// ---------------------------------------------------------------------------------------------------
// The zavisni troškovi (the legacy rob_Zavisni_Troskovi)
// ---------------------------------------------------------------------------------------------------

// GetRobKalTroskovi returns the zavisni troškovi of a kalkulacija (the legacy PrikaziTABELU of rob_Zavisni_Troskovi).
func (s *RobnoDokumentaResource) GetRobKalTroskovi(ctx context.Context, rdokID int64) ([]domain.RobKalTrosakRow, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.robno.RztrDokumenta(ctx, rdokID)
}

// GetRobKalTrosakForma returns a saved zavisni trošak for its form (the legacy BTN_Izmeni of rob_Zavisni_Troskovi): the
// trošak with the data of its dobavljač and its plaćen porez. found is false without it.
func (s *RobnoDokumentaResource) GetRobKalTrosakForma(ctx context.Context, rdokID, rztrID int64) (domain.RobKalTrosakForma, bool, error) {
	f := domain.RobKalTrosakForma{}
	t, found, err := s.robno.RztrTrosak(ctx, rztrID, rdokID)
	if err != nil || !found {
		return f, false, err
	}
	f.RobKalTrosakRow = t
	if t.Dadok.Valid {
		f.Datum = t.Dadok.Time.Format(common.HtmlLayout)
	}
	if dob, found, err := s.GetRobKupac(ctx, t.Fkto, t.Fana); err != nil {
		return f, true, err
	} else if found {
		f.KupacAdresa, f.KupacPostanskiBroj, f.KupacMesto, f.KupacPib = dob.Adresa, dob.Pobro, dob.Mesto, dob.Pib
		f.SistemPdv = strconv.Itoa(dob.Tippdv)
	}
	if f.Porezi, err = s.robno.RppoTroska(ctx, rztrID); err != nil {
		return f, true, err
	}
	return f, true, nil
}

// ValidateRobKalTrosak validates a zavisni trošak before it is saved, like the legacy Validacija of
// rob_Zavisni_Troskovi: the document exists; the konto and the šifra of the dobavljač are required and it exists; the
// datum and the broj of the document of the dobavljač are required; the fakturna vrednost robe is greater than 0; a
// given konto troška is a sintetički konto of the kontni plan; without the konto troška the POPDV polje of a PDV
// obveznik whose vrsta requires it (robKalPopdv: the messages of the legacy, which here block the save); the rok is a
// whole number; the rows of the plaćen porez (robKalProveriPoreze).
func (s *RobnoDokumentaResource) ValidateRobKalTrosak(ctx context.Context, params domain.RobKalTrosakParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("trkonto", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil {
		add("trkonto", common.ErrMsgReadData)
		return fieldErrors
	}
	if !found {
		add("trkonto", "Dokument ne postoji.")
		return fieldErrors
	}
	konto, sifra := strings.TrimSpace(params.Konto), strings.TrimSpace(params.Sifra)
	switch {
	case konto == "":
		add("trkonto", common.ErrMsgObavezanPodatak)
	case sifra == "":
		add("trsifra", "Dobavljač sa takvom šifrom ne postoji.")
	default:
		if exists, err := s.robno.FkplAnalitikaPostoji(ctx, konto, sifra); err != nil {
			add("trsifra", common.ErrMsgReadData)
		} else if !exists {
			add("trsifra", "Dobavljač sa takvom šifrom ne postoji.")
		}
	}
	if _, ok := robnoUnosDatum(params.Dadok); !ok {
		add("trdadok", "Morate uneti datum dokumenta od dobavljača.")
	}
	if strings.TrimSpace(params.Dokdob) == "" {
		add("trdokdob", "Morate uneti broj računa od dobavljača.")
	}
	if common.StringToFloat64(params.Iznfak) <= 0 {
		add("triznfak", "Morate uneti iznos dokumenta dobavljača.")
	}
	if rok := strings.TrimSpace(params.Rok); rok != "" {
		if _, err := strconv.Atoi(rok); err != nil {
			add("trrok", robnoUnosMsgCeoBroj)
		}
	}
	tkonto := strings.TrimSpace(params.Tkonto)
	if tkonto != "" {
		if exists, err := s.robno.FkplSintetikaPostoji(ctx, tkonto); err != nil {
			add("trtkonto", common.ErrMsgReadData)
		} else if !exists {
			add("trtkonto", "Nepostojeći konto.")
		}
	}
	if tkonto == "" {
		if trazi, err := s.robKalTraziPopdv(ctx, k.Vrd); err != nil {
			add("trpopdv", common.ErrMsgReadData)
		} else if trazi {
			s.robKalPopdv(ctx, "trpopdv", params.Polje, 0, add)
		}
	}
	s.robKalProveriPoreze(ctx, "trporezi", robKalRedovi(params.PorezOsn, params.PorezPpor, params.PorezPo, nil, nil), false, add)
	return fieldErrors
}

// SaveRobKalTrosak saves a zavisni trošak of a kalkulacija (the legacy NoviSLOGRZTR), a new one (RztrID 0, with the
// vrsta and the broj of the document) or a changed one, and its plaćen porez (the rows of its grid with an osnovica,
// rppo of the trošak: with the vrsta, the broj of the document as the račun, the datum of the document of the
// dobavljač and the dobavljač; the legacy deletes the old ones and adds the rows of the grid) in one transaction. It
// returns the id of the trošak.
func (s *RobnoDokumentaResource) SaveRobKalTrosak(ctx context.Context, params domain.RobKalTrosakParams) (rztrID int64, err error) {
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("dokument koji obrađujete ne postoji")
	}
	datum, _ := robnoUnosDatum(params.Dadok)
	porezi, err := s.robKalPorezi(ctx, robKalRedovi(params.PorezOsn, params.PorezPpor, params.PorezPo, nil, nil), datum, false)
	if err != nil {
		return 0, err
	}
	z := repository.RztrUnos{
		Fkto: strings.TrimSpace(params.Konto), Fana: strings.TrimSpace(params.Sifra), Dokdob: strings.TrimSpace(params.Dokdob),
		Opis: strings.TrimSpace(params.Opis), Pornapomena: strings.TrimSpace(params.Pornapomena),
		Polje: strings.ReplaceAll(strings.TrimSpace(params.Polje), ",", "."), Tkonto: strings.TrimSpace(params.Tkonto),
		Dadok: sql.NullTime{Time: datum, Valid: !datum.IsZero()}, Rok: common.StringToInt(params.Rok), Flguvozpdv: params.Flguvozpdv,
		Iznfak: common.StringToFloat64(params.Iznfak), Iznos: common.StringToFloat64(params.Iznos),
	}
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if rztrID = params.RztrID; rztrID > 0 {
		if err = s.robno.RztrIzmeni(ctx, tx, rztrID, k.RdokID, z); err != nil {
			return 0, err
		}
	} else if rztrID, err = s.robno.RztrDodaj(ctx, tx, k.RdokID, k.Vrd, k.Dokum, z); err != nil {
		return 0, err
	}
	if err = s.robno.RppoTroskaBrisi(ctx, tx, rztrID); err != nil {
		return 0, err
	}
	d := repository.KalPorezDokument{Vrd: k.Vrd, Dokum: k.Dokum, Dadok: z.Dadok, Konto: z.Fkto, Sifra: z.Fana, Rac: strconv.Itoa(k.Dokum),
		RztrID: rztrID}
	for _, p := range porezi {
		if err = s.robno.RppoDodaj(ctx, tx, d, p); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rztrID, nil
}

// DeleteRobKalTrosak deletes a zavisni trošak of a kalkulacija with its plaćen porez (the legacy BTN_Brisanje of
// rob_Zavisni_Troskovi) in one transaction.
func (s *RobnoDokumentaResource) DeleteRobKalTrosak(ctx context.Context, rdokID, rztrID int64) (err error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return fmt.Errorf("no user session found")
	}
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = s.robno.RppoTroskaBrisi(ctx, tx, rztrID); err != nil {
		return err
	}
	if err = s.robno.RztrBrisi(ctx, tx, rztrID, rdokID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = fmt.Errorf("zavisni trošak ne postoji")
		}
		return err
	}
	return tx.Commit()
}

package robno

// The entry of the robni dokumenti of a robni nalog: the screens the legacy "Snimi nalog" opens by the
// group of the vrsta dokumenta (fakture, interni računi, kalkulacije, ...). The functions of every
// screen (the validation of its document, its queries and its saving) are collected here, the screen
// of each group in its own section.

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
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
)

// RobnoDokumentaUnosRepos are the repositories of the entry of the documents: the document, the kupac,
// the avansi and the rate, the writes of the stavke, of the cene and of the šifarnik, and the documents
// of the "Prikaz faktura" tab. The other statements of the entry are in repository.RobnoRepository.
type RobnoDokumentaUnosRepos struct {
	Dokument repository.BaseRepository[domain.RobniDokumentRow]
	Kupac    repository.BaseRepository[domain.RobKupacDto]
	Avansi   repository.BaseRepository[domain.RobAvansDto]
	// The rate of the payment of a faktura.
	Rate  repository.BaseRepository[domain.RobRataDto]
	Rpro  repository.BaseRepository[domain.RobRproDto]
	Rcene repository.BaseRepository[domain.RobRceneDto]
	Rsif  repository.BaseRepository[domain.RobRsifDto]
	// The documents of the "Prikaz faktura" tab.
	Prikaz repository.BaseRepository[domain.RobDokumentPrikazRow]
}

// Short messages of the field errors of the entry: the error is shown under its field, so it says
// only what is wrong with the value.
const (
	robnoUnosMsgNeispravanDatum = "Neispravan datum."
	robnoUnosMsgNePostoji       = "Ne postoji."
	robnoUnosMsgCeoBroj         = "Mora biti ceo broj."
)

// ---------------------------------------------------------------------------------------------------
// Fakture veleprodaje (the groups FAK, FRP and FZR, the legacy rob_Fakture; the knjižna odobrenja
// and zaduženja, KPK, get their own screen and validation)
// ---------------------------------------------------------------------------------------------------

// ValidateRobFakture validates the header of a faktura veleprodaje before it is saved, like the legacy
// procedure ValidacijaRDOK of rob_Fakture. Only the checks that end in an error of a field are
// done; the questions of the legacy (the datum dokumenta out of the business year, the izvorni
// dokument of another kupac or already used) do not stop the saving and are not part of it:
//   - the datum dokumenta is required; the datum izvornog dokumenta is required and not after the
//     datum naloga;
//   - the konto and the šifra of the kupac are required and the kupac must exist (fkpl); the mesto
//     isporuke must be one of the kupac;
//   - the org. jedinica, the mesto troška and the komercijalista, when given, must exist, and so must the banka
//     za uplatu - izvoz;
//   - the avansni računi (profak and the grid of the avansi, vrsta 172) must be of the kupac;
//   - a PDV obveznik (fvr.obv) enters the POPDV polje for a vrsta dokumenta with the POPDV
//     (dokvrsta.flgpopdv): it must exist, be of the type 3 and of an evidencija of the izlazni računi
//     (fvepdv.vktip "I"); a firm out of the PDV enters the poreska kategorija;
//   - the datum prve rate is required with the rate;
//   - a budžetski kupac has the broj ugovora;
//   - the nastanak PDV obaveze is required; a poreska kategorija has the napomena o poreskom
//     oslobođenju;
//   - the broj dokumenta is required and no other nalog has the document of the same vrsta and broj
//     in the magacin.
func (s *RobnoDokumentaResource) ValidateRobFakture(ctx context.Context, params domain.RobnoFaktureParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}

	// The vrsta dokumenta (its POPDV).
	vrd, err := strconv.Atoi(strings.TrimSpace(params.Vrd))
	if err != nil {
		add("vrd", robnoUnosMsgNePostoji)
		return fieldErrors
	}
	dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, vrd)
	if err != nil {
		add("vrd", common.ErrMsgReadData)
		return fieldErrors
	}
	if !found {
		add("vrd", robnoUnosMsgNePostoji)
		return fieldErrors
	}

	// The datum dokumenta and the datum izvornog dokumenta (not after the datum naloga).
	if _, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	}
	if datiz, ok := robnoUnosDatum(params.IzvorniDokDatum); !ok {
		add("datumizvora", robnoUnosDatumMsg(params.IzvorniDokDatum))
	} else {
		rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID)
		if err != nil {
			add("datumizvora", common.ErrMsgReadData)
		} else if found && rnal.Danal.Valid && datiz.After(rnal.Danal.Time) {
			add("datumizvora", "Posle datuma naloga.")
		}
	}

	// The kupac and its mesto isporuke.
	konto, sifra := strings.TrimSpace(params.KupacKonto), strings.TrimSpace(params.KupacSifra)
	if konto == "" {
		add("kupackonto", common.ErrMsgObavezanPodatak)
	}
	if sifra == "" {
		add("kupacsifra", common.ErrMsgObavezanPodatak)
	}
	kupacOK := false
	if konto != "" && sifra != "" {
		exists, err := s.robno.FkplAnalitikaPostoji(ctx, konto, sifra)
		if err != nil {
			add("kupacsifra", common.ErrMsgReadData)
		} else if !exists {
			add("kupacsifra", robnoUnosMsgNePostoji)
		} else {
			kupacOK = true
		}
	}
	if id, given, ok := robnoUnosID(params.MestoIsporuke); !ok {
		add("fispid", robnoUnosMsgNePostoji)
	} else if given && kupacOK {
		exists, err := s.robno.FispPostoji(ctx, id, konto, sifra)
		if err != nil {
			add("fispid", common.ErrMsgReadData)
		} else if !exists {
			add("fispid", "Nije mesto isporuke kupca.")
		}
	}

	// The org. jedinica and the mesto troška, the komercijalista (each when given).
	s.robnoUnosPostojiID(ctx, params.Orgjed, "idorgjed", s.robno.OrgjedPostoji, add)
	s.robnoUnosPostojiID(ctx, params.MestoTroska, "mestotrid", s.robno.MestotrPostoji, add)
	if id, given, ok := robnoUnosID(params.Komercijalista); !ok {
		add("komid", robnoUnosMsgNePostoji)
	} else if given {
		if exists, err := s.robno.KomercijalistaPostoji(ctx, id); err != nil {
			add("komid", common.ErrMsgReadData)
		} else if !exists {
			add("komid", robnoUnosMsgNePostoji)
		}
	}

	// The banka za uplatu - izvoz (when given).
	if id, given, ok := robnoUnosID(params.BankaIzvoz); !ok {
		add("bnkizvid", robnoUnosMsgNePostoji)
	} else if given {
		if exists, err := s.robno.BnkizvPostoji(ctx, id); err != nil {
			add("bnkizvid", common.ErrMsgReadData)
		} else if !exists {
			add("bnkizvid", robnoUnosMsgNePostoji)
		}
	}

	// The avansi of the grid: the broj of an avansni račun of the kupac with the data of the avans.
	if kupacOK {
		if _, nepostoje, err := s.robAvansiIDs(ctx, params, konto, sifra); err != nil {
			add("avansbrdok", common.ErrMsgReadData)
		} else if len(nepostoje) > 0 {
			add("avansbrdok", "Nije avans kupca: "+strings.Join(nepostoje, ", ")+".")
		}
	}

	// The avansni računi (vrsta 172) of the faktura must be of the kupac (an avansni račun has its profaktura
	// there).
	if profak := strings.TrimSpace(params.Profaktura); profak != "" && !params.Avansni {
		brojevi, ok := robnoUnosBrojevi(profak)
		if !ok {
			add("profak", "Brojevi odvojeni zarezom.")
		} else if kupacOK {
			exists, err := s.robno.RdokAvansniDrugogKupca(ctx, brojevi, konto, sifra)
			if err != nil {
				add("profak", common.ErrMsgReadData)
			} else if exists {
				add("profak", "Avansni račun drugog kupca.")
			}
		}
	}

	// The POPDV polje of a PDV obveznik, the poreska kategorija of a firm out of the PDV; a given
	// poreska kategorija must exist.
	porkatID, porkatGiven, porkatOK := robnoUnosID(params.PorKategorija)
	obveznik, err := s.robnoUnosPdvObveznik(ctx)
	if err != nil {
		add("popdv", common.ErrMsgReadData)
	} else if obveznik {
		if dokvrsta.FlgPopDV {
			s.robnoUnosPopdv(ctx, userSession, strings.TrimSpace(params.Popdv), add)
		}
	} else if porkatOK && !porkatGiven {
		add("porkatid", common.ErrMsgObavezanPodatak)
	}
	if !porkatOK {
		add("porkatid", robnoUnosMsgNePostoji)
	} else if porkatGiven {
		if exists, err := s.robno.PorkatPostoji(ctx, porkatID); err != nil {
			add("porkatid", common.ErrMsgReadData)
		} else if !exists {
			add("porkatid", robnoUnosMsgNePostoji)
		}
	}

	// The datum prve rate of the rate.
	if brojRata, err := strconv.Atoi(strings.TrimSpace(params.BrojRata)); err == nil && brojRata > 0 {
		if _, ok := robnoUnosDatum(params.DatumPrveRate); !ok {
			add("datumprverate", robnoUnosDatumMsg(params.DatumPrveRate))
		}
	}

	// The kupac: the broj ugovora of a budžetski kupac, the kredit limiti (the legacy SaldoKUPCA; not of an
	// avansni račun).
	if kupacOK && !params.Avansni {
		kupac, _, err := s.GetRobKupac(ctx, konto, sifra)
		if err != nil {
			add("kupacsifra", common.ErrMsgReadData)
		} else {
			if kupac.Budzetski && strings.TrimSpace(params.BrojUgovora) == "" {
				add("brojugovora", common.ErrMsgObavezanPodatak)
			}
			if kupac.Krlimit > 0 || kupac.Krlimitdosp > 0 {
				datum, ok := robnoUnosDatum(params.DatumDokum)
				if !ok {
					datum = time.Now()
				}
				if saldo, err := s.GetRobSaldoKupca(ctx, konto, sifra, datum); err != nil {
					add("kupacsifra", common.ErrMsgReadData)
				} else if msg := RobKreditLimitPoruka(kupac, saldo); msg != "" {
					add("kupacsifra", msg)
				}
			}
		}
	}

	// The nastanak PDV obaveze; the napomena o poreskom oslobođenju of a poreska kategorija.
	if nastanak := strings.TrimSpace(params.NastanakPDV); nastanak == "" || nastanak == "-1" || nastanak == RobNastanakNeIzabran {
		add("nastanakpdv", common.ErrMsgObavezanPodatak)
	}
	if porkatGiven && strings.TrimSpace(params.NapomenaPoresko) == "" && !params.Avansni {
		add("napomporeskoosl", common.ErrMsgObavezanPodatak)
	}

	// The broj dokumenta: no other nalog has the document of the vrsta and broj in the magacin.
	dokum, err := strconv.Atoi(strings.TrimSpace(params.BrojDokum))
	switch {
	case strings.TrimSpace(params.BrojDokum) == "":
		add("brdok", common.ErrMsgObavezanPodatak)
	case err != nil:
		add("brdok", robnoUnosMsgCeoBroj)
	default:
		exists, err := s.robno.RdokBrojDrugogNalogaPostoji(ctx, params.MagaciniID, vrd, dokum, params.RnalID, params.RdokID)
		if err != nil {
			add("brdok", common.ErrMsgReadData)
		} else if exists {
			add("brdok", "Postoji u drugom nalogu.")
		}
	}
	return fieldErrors
}

// GetRobFaktureDokument returns the document of the "Fakture veleprodaje" screen with the broj the user
// selected or typed, like the legacy exit of EDT_DOKUM of rob_Fakture:
//   - an existing document of the nalog that can be changed (RobProveraIzmeneDokumenta) is read with
//     its kupac, its finansijsko stanje (dospeće, saldo) and its avansi (the legacy FileToScreen,
//     PrikaziFISP, Dospece, SaldoKUPCA and PuniTableAVANS) into the header, which stays editable: the
//     "Sačuvaj" of the header updates the document (RdokID), and only then the stavke are entered;
//   - a document that cannot be changed gives the message (Poruka) and a new document with the next
//     broj (NoviBrojRDOK);
//   - a broj without a document gives a new document with that broj, an empty broj one with the next
//     broj, both with the defaults of the legacy Reset (the domaća valuta, kurs and važi za 1, the
//     POPDV polje and the oznaka of the poreska kategorija, the konto of the kupci, the datumi of the
//     nalog; PuniPOLJE, PuniFKTO, PuniCOMBOS).
//
// The nastanak PDV obaveze and the žiro računi are the values of the combos of both.
func (s *RobnoDokumentaResource) GetRobFaktureDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoFaktureDokument, error) {
	res := domain.RobnoFaktureDokument{}
	// The document of the id gives its vrsta (and its magacin without the one of the nalog: a document
	// of another nalog gives a new one in the magacin of the nalog).
	var byID domain.RobniDokumentRow
	byIDFound := false
	if params.RdokID > 0 {
		var err error
		if byID, byIDFound, err = s.GetRobniDokumentByID(ctx, params.RdokID); err != nil {
			return res, err
		}
		if !byIDFound {
			return res, fmt.Errorf("dokument ne postoji")
		}
		params.Vrd = byID.Vrd
		if params.MagaciniID <= 0 {
			params.MagaciniID = byID.MagaciniID
		}
	}
	rnal, _, err := s.robnoUnosNalog(ctx, params.RnalID)
	if err != nil {
		return res, err
	}
	if params.MagaciniID <= 0 && rnal.MagaciniID.Valid {
		params.MagaciniID = int(rnal.MagaciniID.Int64)
	}
	dokvrsta, _, err := s.robnoUnosDokvrsta(ctx, params.Vrd)
	if err != nil {
		return res, err
	}
	obveznik, err := s.robnoUnosPdvObveznik(ctx)
	if err != nil {
		return res, err
	}
	if res.ZiroRacuni, err = s.GetRobZiroRacuni(ctx); err != nil {
		return res, err
	}
	h := &res.Header
	h.RnalID, h.MagaciniID, h.Vrd = params.RnalID, params.MagaciniID, strconv.Itoa(params.Vrd)
	h.PdvObveznik, h.Storno = obveznik, strings.TrimSpace(dokvrsta.Predznak) == "-"
	var nastanakDefault string
	h.CbxNastanakPDV, nastanakDefault = RobFaktureNastanakPDV(h.Storno, obveznik, dokvrsta.GrpDok)

	// The document of the id, or of the broj: an existing one that can be changed is opened.
	dokum, dokumErr := strconv.Atoi(strings.TrimSpace(params.Dokum))
	if byIDFound {
		dokum, dokumErr = byID.Dokum, nil
	}
	if dokumErr == nil && dokum > 0 {
		row, found := byID, byIDFound
		if !found {
			if row, found, err = s.GetRobniDokument(ctx, params.MagaciniID, params.Vrd, dokum); err != nil {
				return res, err
			}
		}
		if found {
			if res.Poruka = RobProveraIzmeneDokumenta(row, params.RnalID); res.Poruka == "" {
				err := s.robFaktureIzDokumenta(ctx, &res, row)
				s.robFaktureStavkaDokumenta(ctx, &res)
				return res, err
			}
			dokum = 0
		}
	}
	if dokumErr != nil || dokum <= 0 {
		if dokum, err = s.GetNoviBrojDokumenta(ctx, params.MagaciniID, params.Vrd); err != nil {
			return res, err
		}
	}

	// A new document with the defaults.
	h.BrojDokum = strconv.Itoa(dokum)
	if rnal.Danal.Valid {
		h.DatumDokum, h.IzvorniDokDatum, h.DatumSmanjenja = rnal.Danal.Time, rnal.Danal.Time, rnal.Danal.Time
	}
	h.Kurs, h.VaziZa = 1, 1
	h.NastanakPDV = nastanakDefault
	h.ZirpRacun = robZiroNijeIzabran
	if obveznik {
		h.OznakaKat = "S"
		if dokvrsta.FlgPopDV {
			h.Popdv = dokvrsta.Polje
		}
	} else {
		h.OznakaKat = "SS"
	}
	defaults, err := s.robnoUnosNoviDokumentDefaults(ctx)
	if err != nil {
		return res, err
	}
	h.Valuta = common.StringToInt(defaults.Key)
	h.IDValute = robNullID(defaults.C1)
	h.ValutaText = defaults.C2
	h.KupacKonto = defaults.C3
	if h.Popdv != "" {
		h.PopdvText = h.Popdv
	}
	s.robFaktureStavkaDokumenta(ctx, &res)
	return res, nil
}

// robFaktureStavkaDokumenta fills the data of the document the entry of a stavka repeats (read only): the
// konto magacina (the konto of the magacin in magkonto, else the konto of the stavke of the document),
// the konto, the šifra and the naziv of the kupac, the vrsta, the broj and the datum of the document (its
// iznos is filled from the document, robFaktureIzDokumenta).
func (s *RobnoDokumentaResource) robFaktureStavkaDokumenta(ctx context.Context, res *domain.RobnoFaktureDokument) {
	h := &res.Header
	st := &res.Stavka
	st.KontoKupca = strings.TrimSpace(h.KupacKonto + " " + h.KupacSifra)
	st.KupacNaziv = h.KupacNaziv
	st.Vrd, st.Brdok = h.Vrd, h.BrojDokum
	if !h.DatumDokum.IsZero() {
		st.DatumDokumenta = h.DatumDokum.Format(common.DateLayout)
	}
	st.RdokID = h.RdokID
	st.Taxexreco = "-"
	st.CbxTaxcat, st.Taxcat = RobTaxCatCombo(h.PdvObveznik)
	st.CbxTaxexreco, _ = RobTaxExemCombo()
	if k, found, err := s.GetRobStavkeKontekst(ctx, h.RdokID); err == nil && found {
		st.Devizni, st.Kurs, st.VaziZa = k.Devizni(), k.Kurs, k.Vaziza
		if st.Devizni {
			st.ValutaPoruka = strings.TrimSpace(fmt.Sprintf("Unesite cenu u valuti: %d %s", k.Sifval, k.Valuta))
		}
		st.NabavnaCenaDokumenta = k.NabavnaCena()
	}
	// The konto of a new stavka: the first konto of the magacin.
	if konta, err := s.GetRobKontaMagacina(ctx, h.MagaciniID); err == nil && len(konta) > 0 {
		st.Konto = konta[0]
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return
	}
	if konto, err := s.robno.KontoMagacinaIliStavke(ctx, h.MagaciniID, h.RdokID); err == nil {
		st.KontoMagacina = konto
	}
}

// robFaktureIzDokumenta fills the document of the "Fakture veleprodaje" screen from an existing robni
// dokument: its header, its kupac and finansijsko stanje, its avansi.
func (s *RobnoDokumentaResource) robFaktureIzDokumenta(ctx context.Context, res *domain.RobnoFaktureDokument, row domain.RobniDokumentRow) error {
	h := &res.Header
	res.Stavka.IznosDokumenta = robFormatIznos(row.Iznos)
	// The header is filled, not saved: "Sačuvaj" updates the document (rdokid) and opens the stavke.
	h.RdokID = row.RdokID
	h.BrojDokum = strconv.Itoa(row.Dokum)
	h.DatumDokum = robNullTime(row.Dadok)
	h.IzvorniDokum, h.IzvorniDokDatum = row.Dokiz, robNullTime(row.Datiz)
	h.Opis = row.Opis
	h.ZaPeriod, h.PeriodOd, h.PeriodDo = row.Kpperiod, robNullTime(row.Oddat), robNullTime(row.Dodat)
	h.KupacKonto, h.KupacSifra = row.Fkto, row.Fana
	h.Valuta, h.IDValute, h.ValutaText = row.Sifval, row.IDValute, row.ValutaText
	h.Kurs, h.VaziZa = row.Kurs, row.Vaziza
	h.BrojRata, h.DatumPrveRate = row.BrojRata, robNullTime(row.DatumPrveRate)
	h.UsloviPlacanja, h.UsloviKamate, h.NacinOtpreme = row.Pla, row.Kamt, row.Otp
	h.Rok = row.Rok
	h.Dospeva = RobDospece(row.Dadok, row.Rok)
	h.BrojUgovora, h.Narudzbenica = row.Brugov, row.Narudzb
	h.ZirpRacun = row.Posuslnab
	if strings.TrimSpace(h.ZirpRacun) == "" {
		h.ZirpRacun = robZiroNijeIzabran
	}
	h.Vozac, h.RegBr = row.Vozac, row.Brvozila
	h.BrutoTezina, h.NetoTezina = row.Bttowght, row.Netwght
	h.BrojKutija, h.BrojPaleta = row.Totboxs, row.Brpaleta
	h.Paritet, h.IzvozIzjava = row.Paritet, row.Izjizv
	h.IDBankaIzvoz, h.BankaIzvozText = row.BnkizvID, row.BnkizvText
	h.Popdv, h.PopdvText = row.Polje, row.PopdvText
	h.PorKategorija, h.OznakaKat, h.NapomenaPoresko = row.Keykat, row.Category, row.Pornapomena
	h.IDPorKat, h.PorKatText = row.PorkatID, row.PorkatText
	if row.Descriptioncode.Valid {
		h.NastanakPDV = strconv.FormatInt(row.Descriptioncode.Int64, 10)
	}
	h.Orgejed, h.OrgjedText = row.IDOrgjed, row.OrgjedText
	h.MestoTroska, h.MestoTroskaText = row.MestoTrID, row.MestoTrText
	h.MestoIsporuke, h.MestoIsporukeText = row.FispID, row.FispText
	h.MispPostanskiBroj, h.MispMesto = row.FispPobro, row.FispMesto
	h.Komercijalista, h.KomercijalistaText = row.KomID, row.KomText
	h.Tekst = row.Foot
	h.JavnaNabavka, h.JBKJNosioc, h.BrojTendera = row.Javnab, row.Jbkjsnjn, row.Brtend

	// The kupac and its finansijsko stanje on the datum of the document.
	kupac, found, err := s.GetRobKupac(ctx, row.Fkto, row.Fana)
	if err != nil {
		return err
	}
	if found {
		h.KupacNaziv, h.KupacAdresa = kupac.Naziv, kupac.Adresa
		h.KupacPostanskiBroj, h.KupacMesto = kupac.Pobro, kupac.Mesto
		h.KupacPib = kupac.Pib
		h.SistemPdv = strconv.Itoa(kupac.Tippdv)
		h.KupacInfo = kupac.Napomena
		datum := robNullTime(row.Dadok)
		if datum.IsZero() {
			datum = time.Now()
		}
		if h.RobnoFaktureFinStanjeKupca, err = s.GetRobSaldoKupca(ctx, row.Fkto, row.Fana, datum); err != nil {
			return err
		}
	}

	// The avansi closed on the document, by the columns of the grid of the avansi (traži, broj avansa,
	// datum, iznos avansa, ostatak, iznos zatvoren na fakturi; the key of the row is the avans).
	avansi, err := s.GetRobAvansiDokumenta(ctx, row.RdokID)
	if err != nil {
		return err
	}
	for _, a := range avansi {
		res.Avansi = append(res.Avansi, domain.TableRow{
			ID: strconv.FormatInt(a.AvansID, 10),
			Fields: []string{"", strconv.Itoa(a.Dokum), robFormatDatum(a.Dadok),
				robFormatIznos(a.Iznos), robFormatIznos(a.Ostatak), robFormatIznos(a.Zatvoreno)},
		})
	}
	return nil
}

// GetRobFaktureFlags returns the flags of the "Fakture veleprodaje" screen: the firm a PDV obveznik
// (fvr.obv, it selects the poreske kategorije of the obveznici) and the vrsta dokumenta a storno
// (dokvrsta.predznak "-").
func (s *RobnoDokumentaResource) GetRobFaktureFlags(ctx context.Context, vrd int) (pdvObveznik, storno bool, err error) {
	if pdvObveznik, err = s.robnoUnosPdvObveznik(ctx); err != nil {
		return false, false, err
	}
	dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, vrd)
	if err != nil {
		return false, false, err
	}
	return pdvObveznik, found && strings.TrimSpace(dokvrsta.Predznak) == "-", nil
}

// The values of the nastanak PDV obaveze (rdok.descriptioncode, the codes of the e-faktura).
const (
	RobNastanakNeIzabran        = "999" // "-": not chosen (the validation requires a choice)
	RobNastanakNeNastaje        = "0"
	RobNastanakDatumPrometa     = "35"
	RobNastanakClan16Tacka2a    = "3"
	RobNastanakClan36a          = "432"
	robNastanakTextNeNastaje    = "Ne nastaje obaveza obračuna PDV"
	robNastanakTextNeIzabran    = "-"
	robNastanakTextDatumPrometa = "Datum prometa"
)

// RobFaktureNastanakPDV returns the values of the combo of the nastanak PDV obaveze of a robni dokument
// and its default value, like the legacy rob_Fakture fills COMBO_NASTANAK:
//   - a storno: only "Ne nastaje obaveza obračuna PDV" (0);
//   - a PDV obveznik: "-" (999, not chosen) and the moments of the obaveza, by default the datum
//     prometa (35); the knjižna odobrenja and zaduženja (the groups KPK and KPF) and the avansni računi
//     (ARA and ARU: the datum avansa, 432) have their own ones;
//   - a firm out of the PDV: only "Ne nastaje obaveza obračuna PDV" (0).
func RobFaktureNastanakPDV(storno, pdvObveznik bool, grpdok string) ([]domain.ComboItem, string) {
	neNastaje := domain.ComboItem{Key: RobNastanakNeNastaje, Value: robNastanakTextNeNastaje}
	switch {
	case storno || !pdvObveznik:
		return []domain.ComboItem{neNastaje}, RobNastanakNeNastaje
	case strings.EqualFold(strings.TrimSpace(grpdok), "ARA"), strings.EqualFold(strings.TrimSpace(grpdok), "ARU"):
		// An avansni račun (the legacy PuniCOMBOS of rob_AvansniRacuni): by default the datum avansa.
		return []domain.ComboItem{
			{Key: RobNastanakNeIzabran, Value: robNastanakTextNeIzabran},
			{Key: RobNastanakClan36a, Value: "Datum avansa"},
			neNastaje,
		}, RobNastanakClan36a
	case strings.EqualFold(strings.TrimSpace(grpdok), "KPK"), strings.EqualFold(strings.TrimSpace(grpdok), "KPF"):
		return []domain.ComboItem{
			{Key: RobNastanakNeIzabran, Value: robNastanakTextNeIzabran},
			{Key: RobNastanakClan16Tacka2a, Value: "Datum izdavanja-zaračunavanje troškova"},
			{Key: RobNastanakDatumPrometa, Value: "Datum povećanja-ugovor"},
			neNastaje,
		}, RobNastanakDatumPrometa
	default:
		return []domain.ComboItem{
			{Key: RobNastanakNeIzabran, Value: robNastanakTextNeIzabran},
			{Key: RobNastanakDatumPrometa, Value: robNastanakTextDatumPrometa},
			{Key: RobNastanakClan16Tacka2a, Value: "Član 16 tačka 2a) ZPDV"},
			{Key: RobNastanakClan36a, Value: "Član 36a ZPDV"},
			neNastaje,
		}, RobNastanakDatumPrometa
	}
}

// robnoUnosPdvObveznik tells whether the firm of the session is a PDV obveznik (fvr.obv).
func (s *RobnoDokumentaResource) robnoUnosPdvObveznik(ctx context.Context) (bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return false, fmt.Errorf("no user session found")
	}
	return s.robno.FvrObveznikPdv(ctx, userSession.Firma)
}

// robnoUnosPopdv checks the POPDV polje of a faktura of a PDV obveznik: it is required, must exist in
// the period, be of the type 3 and of an evidencija of the izlazni računi (fvepdv.vktip "I").
func (s *RobnoDokumentaResource) robnoUnosPopdv(ctx context.Context, userSession *domain.UserSession, polje string, add func(field, message string)) {
	if polje == "" {
		add("popdv", common.ErrMsgObavezanPodatak)
		return
	}
	exists, err := s.robno.PopdvPoljePostoji(ctx, polje)
	if err != nil {
		add("popdv", common.ErrMsgReadData)
		return
	}
	if !exists {
		add("popdv", robnoUnosMsgNePostoji)
		return
	}
	neodgovarajuce, err := s.robno.PopdvPoljeNeodgovarajuce(ctx, polje)
	if err != nil {
		add("popdv", common.ErrMsgReadData)
	} else if neodgovarajuce {
		add("popdv", "Neodgovarajuće polje.")
	}
}

// SaveRobFaktura saves the header of the faktura veleprodaje (the legacy "Sačuvaj" of rob_Fakture, after
// ValidacijaRDOK): the robni dokument with the fields of the header and the avansi of its grid, a new
// one when the nalog has no document of the vrsta and broj in the magacin (SnimiRDOK), the iznosi of its
// rate (PreracunRATA, the end of SnimiRDOK), then its obračun (ObracunFAKT). It returns the rdokid of
// the document.
func (s *RobnoDokumentaResource) SaveRobFaktura(ctx context.Context, params domain.RobnoFaktureParams) (int64, error) {
	dok, err := s.robFaktureSnimanje(ctx, params)
	if err != nil {
		return 0, err
	}
	rdokID, err := s.SaveRobniDokument(ctx, dok)
	if err != nil {
		return 0, err
	}
	if err := s.PreracunRataDokumenta(ctx, rdokID); err != nil {
		return rdokID, err
	}
	return rdokID, s.ObracunFakture(ctx, rdokID)
}

// robFaktureSnimanje returns the document to save from the header of the faktura: the columns of rdok
// of the header (the legacy ScreenToFile), with the old columns of the šifarnici (mi, kom, sifval,
// sifbank, keykat and category) by their keys, and the avansi of the grid.
func (s *RobnoDokumentaResource) robFaktureSnimanje(ctx context.Context, params domain.RobnoFaktureParams) (domain.RobniDokumentSnimanje, error) {
	dok := domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        common.StringToInt(params.Vrd),
		Dokum:      common.StringToInt(params.BrojDokum),
		SaAvansima: true,
	}
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	fispID, komID, valutaID, bankaID, porkatID := id(params.MestoIsporuke), id(params.Komercijalista), id(params.IDValute),
		id(params.BankaIzvoz), id(params.PorKategorija)

	// The old columns of the šifarnici by their keys, in one query.
	sif, err := s.robno.SifreSifarnika(ctx, fispID, komID, valutaID, bankaID, porkatID)
	if err != nil {
		return dok, err
	}
	nastanak := robPoljeInt("descriptioncode", params.NastanakPDV)
	if strings.TrimSpace(params.NastanakPDV) == RobNastanakNeIzabran {
		nastanak = domain.Fields{Name: "descriptioncode", IsNull: true}
	}
	dok.Polja = []domain.Fields{
		robPoljeDatum("dadok", params.DatumDokum),
		robPoljeText("dokiz", params.IzvorniDokum),
		robPoljeDatum("datiz", params.IzvorniDokDatum),
		robPoljeText("opis", params.Opis),
		robPoljeInt("rok", params.Rok),
		robPoljeText("fkto", params.KupacKonto),
		robPoljeText("fana", params.KupacSifra),
		robPoljeID("fispid", fispID), robPoljeInt("mi", sif.Key),
		robPoljeID("komid", komID), robPoljeInt("kom", sif.C1),
		robPoljeID("idvalute", valutaID), robPoljeInt("sifval", sif.C2),
		robPoljeBroj("kurs", params.Kurs),
		robPoljeInt("vaziza", params.VaziZa),
		robPoljeInt("brojrata", params.BrojRata),
		robPoljeDatum("datumprrate", params.DatumPrveRate),
		robPoljeText("pla", params.UsloviPlacanja),
		robPoljeText("kamt", params.UsloviKamate),
		robPoljeText("otp", params.NacinOtpreme),
		robPoljeText("posuslnab", params.ZiroRacun),
		robPoljeText("brugov", params.BrojUgovora),
		robPoljeText("narudzb", params.Narudzbenica),
		robPoljeText("paritet", params.Paritet),
		robPoljeInt("izjizv", params.IzvozIzjava),
		robPoljeID("bnkizvid", bankaID), robPoljeInt("sifbank", sif.C3),
		robPoljeText("polje", params.Popdv),
		robPoljeText("keykat", sif.C4),
		robPoljeText("category", sif.C5),
		robPoljeText("pornapomena", params.NapomenaPoresko),
		nastanak,
		robPoljeID("idorgjed", id(params.Orgjed)),
		robPoljeID("mestotrid", id(params.MestoTroska)),
		robPoljeText("vozac", params.Vozac),
		robPoljeText("brvozila", params.RegBr),
		robPoljeBroj("bttowght", params.BrutoTezina),
		robPoljeBroj("netwght", params.NetoTezina),
		robPoljeInt("totboxs", params.BrojKutija),
		robPoljeInt("brpaleta", params.BrojPaleta),
		{Name: "javnab", Value: strconv.FormatBool(robnoUnosChecked(params.JavnaNabavka))},
		robPoljeText("jbkjsnjn", params.JBKJSNosilac),
		robPoljeText("brtend", params.BrojTendera),
		robPoljeText("foot", params.Tekst),
	}
	if dok.Avansi, _, err = s.robAvansiIDs(ctx, params, params.KupacKonto, params.KupacSifra); err != nil {
		return dok, err
	}
	return dok, nil
}

// robAvansiIDs returns the avansi of the grid of the avansi of a document (the key of a saved row, or
// the avansni račun, vrsta 172, of the kupac with the broj typed in a new row) and the brojevi that are
// not an avans of the kupac. An avans is an avansni račun with its data (a row of avansfakt).
func (s *RobnoDokumentaResource) robAvansiIDs(ctx context.Context, params domain.RobnoFaktureParams, konto, sifra string) ([]int64, []string, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, nil, fmt.Errorf("no user session found")
	}
	var ids []int64
	var nepostoje []string
	vidjen := map[int64]bool{}
	for i, broj := range params.AvansBrdok {
		broj = strings.TrimSpace(broj)
		var avansID int64
		if i < len(params.AvansID) {
			avansID, _, _ = robnoUnosID(params.AvansID[i])
		}
		if avansID == 0 && broj != "" {
			dokum, err := strconv.Atoi(broj)
			if err == nil {
				id, err := s.robno.AvansIDPoBroju(ctx, dokum, konto, sifra)
				if err != nil {
					return nil, nil, err
				}
				avansID = id
			}
			if avansID == 0 {
				nepostoje = append(nepostoje, broj)
			}
		}
		if avansID > 0 && !vidjen[avansID] {
			vidjen[avansID] = true
			ids = append(ids, avansID)
		}
	}
	return ids, nepostoje, nil
}

// ---------------------------------------------------------------------------------------------------
// The functions of the entry every document can use (the legacy procedures of the windows)
// ---------------------------------------------------------------------------------------------------

// SaveRobniDokument saves a robni dokument of a nalog in one transaction (the legacy SnimiRDOK, shared by
// the documents of the robno, which all are rdok):
//   - a new document (no document of the vrsta and broj in the magacin) is added with the columns of the
//     screen, the columns of its nalog (vrsta, broj and datum of the nalog, datum obrade, magacin), the
//     counters at 0 and knjige "N" (not posted); the nalog counts one document more (rnal.brdo);
//   - an existing one is changed with the columns of the screen, and its stavke get its datum dokumenta
//     and, when the kupac changed, its šifra (rpro.dadok, rpro.fana), and the prodavnica of the document
//     when it has one (rpro.pkto, rpro.pana: the interni prenosi);
//   - with SaAvansima the avansi closed on the document follow the list: a missing avans is added with
//     the osnovice and the PDV of the avans (the faktura parts are computed by ObracunFakture), an avans
//     no more in the list is removed.
//
// The xdat/xop columns are written by the repository. The iznosi of the rate of the document are the
// next step (PreracunRataDokumenta).
func (s *RobnoDokumentaResource) SaveRobniDokument(ctx context.Context, dok domain.RobniDokumentSnimanje) (rdokID int64, err error) {
	return s.saveRobniDokumentSa(ctx, dok, nil)
}

// saveRobniDokumentSa is SaveRobniDokument with dodatno, the save of the other data of the document (e.g. the plaćen
// porez of a kalkulacija) in the same transaction, after the document (nil: none).
func (s *RobnoDokumentaResource) saveRobniDokumentSa(ctx context.Context, dok domain.RobniDokumentSnimanje,
	dodatno func(ctx context.Context, tx db.Transaction, rdokID int64) error) (rdokID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	tx, err := s.unos.Dokument.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// The document of the vrsta and broj in the magacin (locked for the change).
	var staraFana string
	rdokID, staraFana, err = s.robno.RdokPoBrojuFana(ctx, tx, dok.MagaciniID, dok.Vrd, dok.Dokum)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		if rdokID, err = s.robDodajDokument(ctx, tx, dok); err != nil {
			return 0, err
		}
	case err != nil:
		return 0, err
	default:
		if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", rdokID, dok.Polja); err != nil {
			return 0, err
		}
		// The stavke get the datum dokumenta and the new kupac.
		if dadok, ok := robPolje(dok.Polja, "dadok"); ok {
			if err = s.robno.RproPostaviDadok(ctx, tx, rdokID, robPoljeVrednost(dadok)); err != nil {
				return 0, err
			}
		}
		if fana, ok := robPolje(dok.Polja, "fana"); ok && staraFana != "" && staraFana != fana.Value {
			if err = s.robno.RproPostaviPartnera(ctx, tx, rdokID, "fana", fana.Value); err != nil {
				return 0, err
			}
		}
		// The stavke get the prodavnica of the document (the interni prenosi), when it has one.
		for _, kolona := range []string{"pkto", "pana"} {
			if p, ok := robPolje(dok.Polja, kolona); ok && strings.TrimSpace(p.Value) != "" {
				if err = s.robno.RproPostaviPartnera(ctx, tx, rdokID, kolona, strings.TrimSpace(p.Value)); err != nil {
					return 0, err
				}
			}
		}
	}

	if dok.SaAvansima {
		if err = s.robSnimiAvanse(ctx, tx, rdokID, dok.Avansi); err != nil {
			return 0, err
		}
	}
	if dodatno != nil {
		if err = dodatno(ctx, tx, rdokID); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// robDodajDokument adds a new robni dokument of the nalog (SaveRobniDokument) and counts it in the nalog.
func (s *RobnoDokumentaResource) robDodajDokument(ctx context.Context, tx db.Transaction, dok domain.RobniDokumentSnimanje) (int64, error) {
	n, err := s.robno.RnalZaDokument(ctx, tx, dok.RnalID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("robni nalog %d ne postoji", dok.RnalID)
	}
	if err != nil {
		return 0, err
	}
	tipdok, nalog, danal, datob, mag, magaciniID, brdo := n.Tipdok, n.Nalog, n.Danal, n.Datob, n.Mag, n.MagaciniID, n.Brdo
	polja := append(append([]domain.Fields{}, dok.Polja...), dok.NoviPolja...)
	// The columns of the nalog and the defaults of a new document; a default the screen writes itself (e.g. the
	// broj fiskalnog isečka, bi, of the profaktura, the iznos of a knjižno pismo without stavke) is the one of the
	// screen (a column may be written only once).
	for _, p := range []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(dok.RnalID, 10)},
		{Name: "tipdok", Value: tipdok},
		{Name: "nalog", Value: strconv.Itoa(nalog)},
		robPoljeNullTime("danal", danal),
		robPoljeNullTime("datob", datob),
		{Name: "mag", Value: strconv.Itoa(mag)},
		{Name: "magaciniid", Value: strconv.Itoa(magaciniID)},
		{Name: "vrd", Value: strconv.Itoa(dok.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(dok.Dokum)},
		{Name: "iznos", Value: "0"},
		{Name: "rbr", Value: "0"},
		{Name: "brst", Value: "0"},
		{Name: "bi", Value: "0"},
		{Name: "avansid", Value: "0"},
		{Name: "opiz", Value: ""},
		{Name: "knjige_1", Value: "N"},
		{Name: "knjige_2", Value: "N"},
		// NOT NULL without a default (the polje of the EPP; empty, the legacy does not write it).
		{Name: "epp_polje", Value: ""},
	} {
		if _, ima := robPolje(polja, p.Name); !ima {
			polja = append(polja, p)
		}
	}
	rdokID, err := s.unos.Dokument.CreateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", polja)
	if err != nil {
		return 0, err
	}
	// The nalog counts one document more.
	if err := s.rnalHeaderRepo.UpdateWithTx(ctx, tx, &domain.Rnal{}, "rnalid", dok.RnalID,
		[]domain.Fields{{Name: "brdo", Value: strconv.Itoa(brdo + 1)}}); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// robSnimiAvanse makes the avansi closed on the document the ones of the list (SaveRobniDokument): a
// missing avans is added with the osnovice and the PDV of the avans (its first row of avansfakt) and
// the faktura parts 0, an avans no more in the list is removed.
func (s *RobnoDokumentaResource) robSnimiAvanse(ctx context.Context, tx db.Transaction, rdokID int64, avansi []int64) error {
	if err := s.robno.AvansfaktBrisiOstale(ctx, tx, rdokID, avansi); err != nil {
		return err
	}
	for _, avansID := range avansi {
		postoji, err := s.robno.AvansfaktPostoji(ctx, tx, avansID, rdokID)
		if err != nil {
			return err
		}
		if postoji {
			continue
		}
		a, err := s.robno.AvansfaktIznosi(ctx, tx, avansID)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if _, err := s.unos.Avansi.CreateWithTx(ctx, tx, &domain.RobAvansDto{}, "avansfaktid", []domain.Fields{
			{Name: "rdokid", Value: strconv.FormatInt(rdokID, 10)},
			{Name: "avansid", Value: strconv.FormatInt(avansID, 10)},
			robPoljeFloat("osnavans1", a.Osn1), robPoljeFloat("osnavans2", a.Osn2), robPoljeFloat("osnavans3", a.Osn3),
			robPoljeFloat("pdv1avans", a.Pdv1), robPoljeFloat("pdv2avans", a.Pdv2),
			{Name: "osnfakt1", Value: "0"}, {Name: "osnfakt2", Value: "0"}, {Name: "osnfakt3", Value: "0"},
			{Name: "pdv1fakt", Value: "0"}, {Name: "pdv2fakt", Value: "0"},
		}); err != nil {
			return err
		}
	}
	return nil
}

// ObracunFakture computes the iznosi of a faktura from its stavke and closes its avansi (the legacy
// ObracunFAKT): see RobObracunFakture. The document gets its iznos (the magacinska vrednost), the
// vrednost of the ugovoreni rabat and of the kasa skonto; every avans of the document gets the part of
// the osnovice and of the PDV of the faktura it closes.
func (s *RobnoDokumentaResource) ObracunFakture(ctx context.Context, rdokID int64) error {
	return s.robObracunFakture(ctx, rdokID, false)
}

// ObracunFaktureUsluge computes the iznosi of a faktura usluga from its stavke and closes its avansi like
// ObracunFakture, with the PDV the avansi close of the obračun of the stavke of the legacy
// rob_FAKTURE_Usluga_DET2 (its ObracunFAKT): the osnovica of the faktura by the stopa of the avans (its PDV
// on its osnovica, in whole percents), at most what is left of the PDV of the avans.
func (s *RobnoDokumentaResource) ObracunFaktureUsluge(ctx context.Context, rdokID int64) error {
	return s.robObracunFakture(ctx, rdokID, true)
}

// robObracunFakture is ObracunFakture (pdvPoStopiAvansa false) and ObracunFaktureUsluge (true).
func (s *RobnoDokumentaResource) robObracunFakture(ctx context.Context, rdokID int64, pdvPoStopiAvansa bool) (err error) {
	stavke, err := s.robStavkeObracuna(ctx, rdokID)
	if err != nil || len(stavke) == 0 {
		return err
	}
	avansiDok, err := s.robno.AvansiObracuna(ctx, rdokID)
	if err != nil {
		return err
	}
	obracun, zatvoreno := robObracunFakture(stavke, avansiDok, pdvPoStopiAvansa)

	tx, err := s.unos.Dokument.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", rdokID, []domain.Fields{
		robPoljeFloat("iznos", obracun.Iznos), robPoljeFloat("vkase", obracun.Kasa), robPoljeFloat("vrugrabat", obracun.UgovoreniRabat),
	}); err != nil {
		return err
	}
	// The avansi: the faktura parts (the key of a row is its avansfaktid, read as avansid).
	for _, a := range zatvoreno {
		if err = s.unos.Avansi.UpdateWithTx(ctx, tx, &domain.RobAvansDto{}, "avansfaktid", a.AvansID, []domain.Fields{
			robPoljeFloat("osnfakt1", a.OsnFakt1), robPoljeFloat("osnfakt2", a.OsnFakt2), robPoljeFloat("osnfakt3", a.OsnFakt3),
			robPoljeFloat("pdv1fakt", a.PdvFakt1), robPoljeFloat("pdv2fakt", a.PdvFakt2),
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// PreracunRataDokumenta computes the iznosi of the rate of the payment of a document (faktrate of the
// rdokid; the legacy PreracunRATA of the RDOK): the first rata keeps its osnovica and procenat and gets
// the PDV of its procenat; the other rate share equally what is left of the procenat, of the osnovica
// (the iznos of the faktura without the PDV) and of the PDV (see RobPreracunFakture). A document without
// rate has nothing to compute.
func (s *RobnoDokumentaResource) PreracunRataDokumenta(ctx context.Context, rdokID int64) (err error) {
	rate, err := s.robno.FaktrateDokumenta(ctx, rdokID)
	if err != nil || len(rate) == 0 {
		return err
	}
	stavke, err := s.robStavkeObracuna(ctx, rdokID)
	if err != nil {
		return err
	}
	faktura := RobPreracunFakture(stavke)
	osnovica := faktura.Iznos - faktura.Rabat - faktura.Kasa - faktura.UgovoreniRabat

	tx, err := s.unos.Rate.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	r2 := func(v float64) float64 { return math.Round(v*100) / 100 }
	// The first rata: its PDV by its procenat.
	var prvaOsnovica, prviProcenat, prviPdv float64
	for _, r := range rate {
		if r.Rbrrate == 1 {
			prvaOsnovica, prviProcenat = r.Osnovica, r.Procenat
			prviPdv = r2(faktura.Pdv * r.Procenat / 100)
			if err = s.unos.Rate.UpdateWithTx(ctx, tx, &domain.RobRataDto{}, "faktrateid", r.FaktrateID,
				[]domain.Fields{robPoljeFloat("pdv", prviPdv)}); err != nil {
				return err
			}
			break
		}
	}
	// The other rate: equal parts of the rest.
	ostale := len(rate) - 1
	if ostale > 0 {
		for _, r := range rate {
			if r.Rbrrate <= 1 {
				continue
			}
			if err = s.unos.Rate.UpdateWithTx(ctx, tx, &domain.RobRataDto{}, "faktrateid", r.FaktrateID, []domain.Fields{
				robPoljeFloat("procenat", (100-prviProcenat)/float64(ostale)),
				robPoljeFloat("osnovica", r2((osnovica-prvaOsnovica)/float64(ostale))),
				robPoljeFloat("pdv", r2((faktura.Pdv-prviPdv)/float64(ostale))),
			}); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// robStavkeObracuna returns the stavke of a document for its obračun (RobObracunFakture and
// RobPreracunFakture): with the stopa of their oznaka on the datum of the document; a document without
// stavke gives one row with zeros (its ugovoreni rabat and kasa skonto).
func (s *RobnoDokumentaResource) robStavkeObracuna(ctx context.Context, rdokID int64) ([]domain.RobObracunStavkaDto, error) {
	return s.robno.StavkeObracuna(ctx, rdokID)
}

// RobPreracun are the iznosi of a faktura computed by RobPreracunFakture.
type RobPreracun struct {
	Iznos              float64 // the prodajna vrednost of the stavke
	Rabat              float64 // the rabat of the stavke
	UgovoreniRabat     float64 // the ugovoreni rabat of the document
	Kasa               float64 // the kasa skonto of the document
	MagacinskaVrednost float64 // the magacinska vrednost of the stavke
	Pdv                float64 // the PDV (0 for a faktura with the napomena o poreskom oslobođenju)
}

// Ukupno is the iznos of the faktura to pay: the prodajna vrednost without the rabat, the ugovoreni
// rabat and the kasa skonto, with the PDV.
func (p RobPreracun) Ukupno() float64 {
	return p.Iznos - p.Rabat - p.Kasa - p.UgovoreniRabat + p.Pdv
}

// RobPreracunFakture computes the iznosi of a faktura from its stavke like the legacy
// PreracunFAKTURAPROIZVGLOB (the RDOK with its stavke rpro): the prodajna and the magacinska vrednost,
// the rabat, the ugovoreni rabat and the kasa skonto of the document, and the PDV of every stavka on its
// osnovica after the rabat, the ugovoreni rabat and the kasa skonto (none with the napomena o poreskom
// oslobođenju).
func RobPreracunFakture(stavke []domain.RobObracunStavkaDto) RobPreracun {
	r2 := func(v float64) float64 { return math.Round(v*100) / 100 }
	var p RobPreracun
	var ugrabat, pkase float64
	oslobodjen := false
	for _, st := range stavke {
		ugrabat, pkase = st.Ugrabat, st.Pkase
		oslobodjen = strings.TrimSpace(st.Pornapomena) != ""
		p.Iznos += st.Kolic * st.Fcena
		p.MagacinskaVrednost += st.Kolic * st.Cena
		p.Rabat += st.Kolic * st.Fcena * st.Rab / 100
		bruto := r2(st.Kolic * st.Fcena)
		rab := r2(bruto * st.Rab / 100)
		ugovoreni := r2((bruto - rab) * ugrabat / 100)
		kasa := r2((bruto - rab - ugovoreni) * pkase / 100)
		osnovica := r2(bruto - rab - ugovoreni - kasa)
		p.Pdv += r2(osnovica * st.Stopa / 100)
	}
	if oslobodjen {
		p.Pdv = 0
	}
	p.Iznos, p.Rabat, p.MagacinskaVrednost = r2(p.Iznos), r2(p.Rabat), r2(p.MagacinskaVrednost)
	p.UgovoreniRabat = r2((p.Iznos - p.Rabat) * ugrabat / 100)
	p.Kasa = r2((p.Iznos - p.Rabat - p.UgovoreniRabat) * pkase / 100)
	return p
}

// RobObracun are the iznosi of a faktura computed by RobObracunFakture.
type RobObracun struct {
	Iznos          float64 // the magacinska vrednost of the stavke (rdok.iznos)
	UgovoreniRabat float64 // the vrednost of the ugovoreni rabat (rdok.vrugrabat)
	Kasa           float64 // the vrednost of the kasa skonto (rdok.vkase)
}

// RobObracunFakture computes the iznosi of a faktura from its stavke like the legacy ObracunFAKT, and
// the parts of the faktura its avansi close:
//   - per stavka the prodajna vrednost and its rabat (rounded to 2 decimals), the magacinska vrednost,
//     and the osnovica by the stopa of the stavke (the viša 18/20, the niža 8/10 and 0; a stavka of a
//     faktura with the napomena o poreskom oslobođenju and a poreska kategorija other than S10/S20 has
//     the stopa 0);
//   - the ugovoreni rabat and the kasa skonto on the vrednost after the rabat, the osnovice after both,
//     the PDV of the osnovice;
//   - the avansi (by their datum) close the osnovice and the PDV of the faktura with what the other
//     fakture did not close of them yet (the avansi of the document with the sums of the other
//     fakture: OsnFakt and PdvFakt of the input). The result has, per avans, the parts of this faktura.
func RobObracunFakture(stavke []domain.RobObracunStavkaDto, avansi []domain.RobObracunAvansDto) (RobObracun, []domain.RobObracunAvansDto) {
	return robObracunFakture(stavke, avansi, false)
}

// robObracunFakture is RobObracunFakture; with pdvPoStopiAvansa the PDV an avans closes is the one of the
// legacy ObracunFAKT of rob_FAKTURE_Usluga_DET2: the osnovica of the faktura still open by the stopa of the
// avans (round(pdvavans / osnavans * 100) percents), rounded to 2 decimals, at most what is left of the PDV
// of the avans.
func robObracunFakture(stavke []domain.RobObracunStavkaDto, avansi []domain.RobObracunAvansDto, pdvPoStopiAvansa bool) (RobObracun, []domain.RobObracunAvansDto) {
	r2 := func(v float64) float64 { return math.Round(v*100) / 100 }
	var iznos, iznmc, vrrab, osnNiza, osnVisa, osn0, stopaNiza, stopaVisa float64
	var ugrabat, pkase float64
	for _, st := range stavke {
		ugrabat, pkase = st.Ugrabat, st.Pkase
		rvre := r2(st.Kolic * st.Fcena)
		rrab := r2(rvre * st.Rab / 100)
		iznos += rvre
		iznmc += r2(st.Kolic * st.Cena)
		vrrab += rrab
		stopa := st.Stopa
		taxcat := strings.TrimSpace(st.Taxcat)
		if strings.TrimSpace(st.Pornapomena) != "" && taxcat != "" && taxcat != "-" && taxcat != "S10" && taxcat != "S20" {
			stopa = 0
		}
		switch stopa {
		case 8, 10:
			stopaNiza = stopa
			osnNiza += r2(rvre - rrab)
		case 18, 20:
			stopaVisa = stopa
			osnVisa += r2(rvre - rrab)
		case 0:
			osn0 += r2(rvre - rrab)
		}
	}
	iznos, vrrab, iznmc = r2(iznos), r2(vrrab), r2(iznmc)
	obracun := RobObracun{Iznos: iznmc}
	obracun.UgovoreniRabat = r2((iznos - vrrab) * ugrabat / 100)
	obracun.Kasa = r2((iznos - vrrab - obracun.UgovoreniRabat) * pkase / 100)
	// The osnovice after the ugovoreni rabat and the kasa skonto, and their PDV.
	umanji := func(osn float64) float64 {
		if osn == 0 {
			return 0
		}
		osn = r2(osn - osn*ugrabat/100)
		return r2(osn - osn*pkase/100)
	}
	osnNiza, osnVisa, osn0 = umanji(osnNiza), umanji(osnVisa), umanji(osn0)
	pdvNiza, pdvVisa := r2(osnNiza*stopaNiza/100), r2(osnVisa*stopaVisa/100)

	// The avansi close what is left of the faktura, avans by avans.
	zatvori := func(faktura *float64, ostatak float64) float64 {
		ostatak = math.Max(ostatak, 0)
		if *faktura <= ostatak {
			deo := *faktura
			*faktura = 0
			return r2(deo)
		}
		*faktura -= ostatak
		return r2(ostatak)
	}
	zatvoreno := make([]domain.RobObracunAvansDto, 0, len(avansi))
	// The PDV of the osnovica still open by the stopa of the avans, at most what is left of its PDV.
	pdvPoStopi := func(osnovica, osnAvansa, pdvAvansa, ostatak float64) float64 {
		pdv := 0.0
		if osnAvansa > 0 {
			pdv = r2(osnovica * math.Round(pdvAvansa/osnAvansa*100) / 100)
		}
		if pdv <= ostatak {
			return pdv
		}
		return ostatak
	}
	for _, a := range avansi {
		z := domain.RobObracunAvansDto{AvansID: a.AvansID}
		if pdvPoStopiAvansa {
			z.PdvFakt1 = pdvPoStopi(osnVisa, a.OsnAvans1, a.PdvAvans1, a.PdvAvans1-a.PdvFakt1)
			z.PdvFakt2 = pdvPoStopi(osnNiza, a.OsnAvans2, a.PdvAvans2, a.PdvAvans2-a.PdvFakt2)
		}
		z.OsnFakt1 = zatvori(&osnVisa, a.OsnAvans1-a.OsnFakt1)
		if !pdvPoStopiAvansa {
			z.PdvFakt1 = zatvori(&pdvVisa, a.PdvAvans1-a.PdvFakt1)
		}
		z.OsnFakt2 = zatvori(&osnNiza, a.OsnAvans2-a.OsnFakt2)
		if !pdvPoStopiAvansa {
			z.PdvFakt2 = zatvori(&pdvNiza, a.PdvAvans2-a.PdvFakt2)
		}
		z.OsnFakt3 = zatvori(&osn0, a.OsnAvans3-a.OsnFakt3)
		zatvoreno = append(zatvoreno, z)
	}
	return obracun, zatvoreno
}

// robPolje returns the field of the columns (found false without it).
func robPolje(polja []domain.Fields, name string) (domain.Fields, bool) {
	for _, p := range polja {
		if p.Name == name {
			return p, true
		}
	}
	return domain.Fields{}, false
}

// robPoljeVrednost returns the value of a field for a query (nil for NULL).
func robPoljeVrednost(p domain.Fields) any {
	if p.IsNull {
		return nil
	}
	return p.Value
}

// robPoljeText is a column of text (trimmed).
func robPoljeText(name, value string) domain.Fields {
	return domain.Fields{Name: name, Value: strings.TrimSpace(value)}
}

// robPoljeDatum is a column of a date of the form (NULL when it is empty or not a date).
func robPoljeDatum(name, value string) domain.Fields {
	t, ok := robnoUnosDatum(value)
	if !ok {
		return domain.Fields{Name: name, IsNull: true}
	}
	return domain.Fields{Name: name, Value: t.Format(common.HtmlLayout)}
}

// robPoljeNullTime is a column of a nullable date.
func robPoljeNullTime(name string, t sql.NullTime) domain.Fields {
	if !t.Valid {
		return domain.Fields{Name: name, IsNull: true}
	}
	return domain.Fields{Name: name, Value: t.Time.Format(common.HtmlLayout)}
}

// robPoljeInt is a column of a whole number of the form (0 when it is empty).
func robPoljeInt(name, value string) domain.Fields {
	return domain.Fields{Name: name, Value: strconv.Itoa(int(math.Round(common.StringToFloat64(value))))}
}

// robPoljeBroj is a column of a decimal number of the form (in the format of the locale; 0 when empty).
func robPoljeBroj(name, value string) domain.Fields {
	return robPoljeFloat(name, common.StringToFloat64(value))
}

// robPoljeFloat is a column of a decimal number.
func robPoljeFloat(name string, value float64) domain.Fields {
	return domain.Fields{Name: name, Value: strconv.FormatFloat(value, 'f', -1, 64)}
}

// robnoUnosChecked tells whether a checkbox of the form is checked ("on", "true" or "1").
func robnoUnosChecked(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "on", "true", "1":
		return true
	}
	return false
}

// robPoljeID is a column of the key of a šifarnik (NULL for no key).
func robPoljeID(name string, id int64) domain.Fields {
	if id <= 0 {
		return domain.Fields{Name: name, IsNull: true}
	}
	return domain.Fields{Name: name, Value: strconv.FormatInt(id, 10)}
}

// ---------------------------------------------------------------------------------------------------
// The functions of the entry every document can use (the legacy procedures of the windows)
// ---------------------------------------------------------------------------------------------------

// robZiroNijeIzabran is the value "-" of the žiro račun of a document (none selected).
const robZiroNijeIzabran = "-"

// robniDokumentQuery reads a robni dokument (rdok) with the texts of its keys for the screen
// (GetRobniDokument by its broj, GetRobniDokumentByID by its id).
const robniDokumentQuery = `select r.rdokid, coalesce(r.rnalid, 0) as rnalid, coalesce(r.magaciniid, 0) as magaciniid, coalesce(r.tipdok, '') as tipdok,
			coalesce(r.nalog, 0) as nalog, r.danal, r.vrd, r.dokum, r.dadok, coalesce(r.dokiz, '') as dokiz, r.datiz,
			coalesce(r.opis, '') as opis, coalesce(r.fkto, '') as fkto, coalesce(r.fana, '') as fana, coalesce(r.rok, 0) as rok,
			coalesce(r.iznos, 0)::float8 as iznos, coalesce(r.sifval, 0) as sifval, coalesce(r.kurs, 0)::float8 as kurs,
			coalesce(r.vaziza, 0) as vaziza, coalesce(r.brojrata, 0) as brojrata, r.datumprrate,
			coalesce(r.pla, '') as pla, coalesce(r.kamt, '') as kamt, coalesce(r.otp, '') as otp, coalesce(r.foot, '') as foot,
			coalesce(r.brugov, '') as brugov, coalesce(r.narudzb, '') as narudzb, coalesce(r.posuslnab, '') as posuslnab,
			coalesce(r.vozac, '') as vozac, coalesce(r.brvozila, '') as brvozila, coalesce(r.bttowght, 0)::float8 as bttowght,
			coalesce(r.netwght, 0)::float8 as netwght, coalesce(r.totboxs, 0) as totboxs, coalesce(r.brpaleta, 0) as brpaleta,
			coalesce(r.paritet, '') as paritet, coalesce(r.izjizv::text, '') as izjizv, coalesce(r.kpperiod, false) as kpperiod,
			r.oddat, r.dodat, coalesce(r.javnab, false) as javnab, coalesce(r.jbkjsnjn, '') as jbkjsnjn,
			coalesce(r.brtend, '') as brtend, coalesce(r.polje, '') as polje, coalesce(r.keykat, '') as keykat,
			coalesce(r.category, '') as category, coalesce(r.pornapomena, '') as pornapomena,
			r.descriptioncode::bigint as descriptioncode, coalesce(r.knjige_1, '') as knjizen,
			coalesce(r.knjige_4, '') as fiskalizovan, r.salesinvoiceid::float8 as salesinvoiceid,
			r.individualvatid::float8 as individualvatid,
			r.idorgjed::bigint as idorgjed, coalesce(concat_ws(' - ', nullif(o.ojozn, ''), nullif(o.naziv, '')), '') as orgjedtext,
			r.mestotrid::bigint as mestotrid, coalesce(concat_ws(' - ', nullif(m.mtroska, ''), nullif(m.opis, '')), '') as mestotrtext,
			fi.fispid, coalesce(concat_ws(' - ', fi.mi::bigint::text, nullif(fi.naziv, '')), '') as fisptext,
			coalesce(nullif(fi.pobro, 0)::text, '') as fisppobro, coalesce(fi.mesto, '') as fispmesto,
			ko.komid::bigint as komid, coalesce(concat_ws(' - ', ko.sifkom::text, nullif(ko.imeprezime, '')), '') as komtext,
			va.idvalute, coalesce(concat_ws(' - ', va.sifval::text, nullif(va.naziv, '')), '') as valutatext,
			bi.bnkizvid, coalesce(concat_ws(' - ', bi.sifbank::text, nullif(bi.bnkdes, '')), '') as bnkizvtext,
			coalesce(concat_ws(' - ', nullif(r.polje, ''), nullif(pd.opis, '')), '') as popdvtext,
			pk.porkatid, case when pk.porkatid is null then '' else concat_ws(' - ', nullif(r.keykat, ''), nullif(r.category, '')) end as porkattext
		from rdok r
		left join orgjed o on o.idorgjed = r.idorgjed
		left join mestotr m on m.mestotrid = r.mestotrid
		left join lateral (select f.fispid, f.mi, f.naziv, f.pobro, f.mesto from fisp f
			where f.fispid = r.fispid or (r.fispid is null and coalesce(r.mi, 0) > 0 and f.god = r.god and f.kar = r.kar
				and f.konto = r.fkto and f.sifra = r.fana and f.mi = r.mi)
			order by f.fispid limit 1) fi on true
		left join lateral (select k.komid, k.sifkom, k.imeprezime from komercijalisti k
			where k.komid = r.komid or (r.komid is null and coalesce(r.kom, 0) > 0 and k.god = r.god and k.kar = r.kar
				and k.sifkom = r.kom)
			order by k.komid limit 1) ko on true
		left join lateral (select v.idvalute, v.sifval, v.naziv from valute v
			where v.idvalute = r.idvalute or (r.idvalute is null and v.sifval = r.sifval)
			order by v.idvalute limit 1) va on true
		left join lateral (select b.bnkizvid, b.sifbank, b.bnkdes from bnkizv b
			where b.bnkizvid = r.bnkizvid or (r.bnkizvid is null and coalesce(r.sifbank, 0) > 0 and b.god = r.god
				and b.kar = r.kar and b.sifbank = r.sifbank)
			order by b.bnkizvid limit 1) bi on true
		left join lateral (select p.opis from popdv p where p.god = r.god and p.kar = r.kar and p.polje = r.polje
			limit 1) pd on true
		left join lateral (select x.porkatid from porkat x where x.keykat = r.keykat and x.category = r.category
			order by x.porkatid limit 1) pk on true`

// GetRobniDokument returns the robni dokument of the magacin with the vrsta and broj (the legacy seek
// GODKARMAGVRDDOKUM) with the texts of its šifarnici in one query: the mesto isporuke, the
// komercijalista, the valuta and the banka za uplatu - izvoz by their keys or, for a document saved
// before them, by the old columns (mi, kom, sifval, sifbank), the POPDV polje and the poreska kategorija
// by their šifre. found is false when the document does not exist.
func (s *RobnoDokumentaResource) GetRobniDokument(ctx context.Context, magaciniID, vrd, dokum int) (domain.RobniDokumentRow, bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.RobniDokumentRow{}, false, fmt.Errorf("no user session found")
	}
	b := querybuilder.New(robniDokumentQuery).
		Where("r.god = ?", userSession.SelectedGod).
		Where("r.kar = ?", userSession.SelectedKar).
		Where("r.magaciniid = ?", magaciniID).
		Where("r.vrd = ?", vrd).
		Where("r.dokum = ?", dokum)
	entities, err := s.unos.Dokument.Select(ctx, b)
	if err != nil {
		return domain.RobniDokumentRow{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.RobniDokumentRow{}, false, nil
	}
	return (*entities)[0], true, nil
}

// GetRobniDokumentByID reads the robni dokument of the id (rdokid) of the selected year and firm, like
// GetRobniDokument.
func (s *RobnoDokumentaResource) GetRobniDokumentByID(ctx context.Context, rdokID int64) (domain.RobniDokumentRow, bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.RobniDokumentRow{}, false, fmt.Errorf("no user session found")
	}
	if rdokID <= 0 {
		return domain.RobniDokumentRow{}, false, nil
	}
	b := querybuilder.New(robniDokumentQuery).
		Where("r.god = ?", userSession.SelectedGod).
		Where("r.kar = ?", userSession.SelectedKar).
		Where("r.rdokid = ?", rdokID)
	entities, err := s.unos.Dokument.Select(ctx, b)
	if err != nil {
		return domain.RobniDokumentRow{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.RobniDokumentRow{}, false, nil
	}
	return (*entities)[0], true, nil
}

// RobProveraIzmeneDokumenta returns why an existing robni dokument cannot be opened for a change in
// the nalog rnalID ("" when it can), like the legacy exit of EDT_DOKUM: the document of another nalog,
// printed on the fiskalni printer (knjige[4] "S"), posted (knjige[1] "D") or sent to the SEF. (The
// legacy check of the broj naloga is the one of the nalog: the document of another nalog.)
func RobProveraIzmeneDokumenta(dok domain.RobniDokumentRow, rnalID int64) string {
	switch {
	case dok.RnalID != rnalID:
		return fmt.Sprintf("U magacinu već postoji dokument %d-%d sa datumom %s. Nalog: %s-%d od %s",
			dok.Vrd, dok.Dokum, robFormatDatum(dok.Dadok), strings.TrimSpace(dok.Tipdok), dok.Nalog, robFormatDatum(dok.Danal))
	case strings.TrimSpace(dok.Fiskalizovan) == "S":
		return "Ovaj dokument je odštampan na fiskalni printer, izmene nisu dozvoljene!"
	case strings.TrimSpace(dok.Knjizen) == "D":
		return "Ovaj dokument je proknjižen, izmene nisu dozvoljene!"
	case dok.SalesInvoiceID.Float64 > 0 || dok.IndividualVatID.Float64 > 0:
		return "Ovaj dokument je poslat na SEF, izmene nisu dozvoljene!"
	}
	return ""
}

// GetNoviBrojDokumenta returns the next broj of the documents of the vrsta in the magacin (the legacy
// NoviBrojRDOK: the largest broj + 1, 1 for the first document).
func (s *RobnoDokumentaResource) GetNoviBrojDokumenta(ctx context.Context, magaciniID, vrd int) (int, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	return s.robno.RdokSledeciBroj(ctx, nil, magaciniID, vrd)
}

// RobDospece returns the datum dospeća of a document (the legacy Dospece: the datum dokumenta plus the
// rok in days; zero without the datum).
func RobDospece(dadok sql.NullTime, rok int) time.Time {
	if !dadok.Valid {
		return time.Time{}
	}
	return dadok.Time.AddDate(0, 0, rok)
}

// GetRobKupac returns the kupac of the konto and šifra (fkpl with its partner; the legacy reads of
// FKPL and PARTNERI): the naziv of the analitika when it has no partner, the napomena as plain text.
// found is false when the kupac does not exist.
func (s *RobnoDokumentaResource) GetRobKupac(ctx context.Context, konto, sifra string) (domain.RobKupacDto, bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.RobKupacDto{}, false, fmt.Errorf("no user session found")
	}
	// konto and sifra are the key of the kupac (Where, not AddEqual): empty ones find no kupac.
	b := querybuilder.New(`select coalesce(p.naziv, f.naziv, '') as naziv, coalesce(p.adresa, '') as adresa,
			coalesce(nullif(p.pobro, 0)::text, '') as pobro, coalesce(p.mesto, '') as mesto, coalesce(p.pib, '') as pib,
			coalesce(p.tippdv, 0) as tippdv, coalesce(p.napomena, '') as napomena,
			coalesce(p.budzetski, false) as budzetski, coalesce(p.krlimit, 0)::float8 as krlimit,
			coalesce(p.krlimitdosp, 0)::float8 as krlimitdosp
		from fkpl f left join partneri p on p.idpartneri = f.idpartneri`).
		Where("f.god = ?", userSession.SelectedGod).
		Where("f.kar = ?", userSession.SelectedKar).
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra)).
		Limit(1)
	entities, err := s.unos.Kupac.Select(ctx, b)
	if err != nil {
		return domain.RobKupacDto{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.RobKupacDto{}, false, nil
	}
	kupac := (*entities)[0]
	// The napomena of the partner is the RTF of the legacy rich edit control.
	kupac.Napomena = common.RtfToText(kupac.Napomena)
	return kupac, true, nil
}

// RobSistemPdvCombo returns the values of the sistem PDV of the partner (partneri.tippdv).
func RobSistemPdvCombo() []domain.ComboItem {
	return []domain.ComboItem{
		{Key: "1", Value: "U sistemu PDV-a"},
		{Key: "2", Value: "Van sistema PDV-a"},
		{Key: "3", Value: "Registrovani poljoprivrednik"},
		{Key: "4", Value: "Fizičko lice"},
		{Key: "5", Value: "Student"},
	}
}

// GetRobSaldoKupca returns the finansijsko stanje of the kupac on the datum (the legacy SaldoKUPCA) in
// one query: the fakture (dospele / nedospele by the datum dokumenta plus the rok) and the uplate of
// the finansijsko knjigovodstvo (fpro, the vrste 10, 30, 80 and 90), and the fakture of the robno
// knjigovodstvo not posted yet (the vrste of the groups FAK, FPR, FRP and KPK without knjige[1] "D"),
// whose iznos is the one of the stavke (the rabat, the ugovoreni rabat, the kasa skonto and the porez
// on the stopa of the datum of the stavka).
func (s *RobnoDokumentaResource) GetRobSaldoKupca(ctx context.Context, konto, sifra string, datum time.Time) (domain.RobnoFaktureFinStanjeKupca, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.RobnoFaktureFinStanjeKupca{}, fmt.Errorf("no user session found")
	}
	entities, err := s.robno.SaldoKupca(ctx, konto, sifra, datum)
	if err != nil {
		return domain.RobnoFaktureFinStanjeKupca{}, err
	}
	if len(entities) == 0 {
		return domain.RobnoFaktureFinStanjeKupca{}, nil
	}
	return RobFinStanjeKupca(entities[0]), nil
}

// RobFinStanjeKupca computes the finansijsko stanje of the kupac from the sums of its saldo like the
// legacy SaldoKUPCA: the uplate close the dospele fakture first, then the nedospele ones.
func RobFinStanjeKupca(sume domain.RobSaldoKupcaDto) domain.RobnoFaktureFinStanjeKupca {
	st := domain.RobnoFaktureFinStanjeKupca{
		UkRealiz:     sume.Fakture,
		Placeno:      sume.Uplate,
		Saldo:        sume.Fakture - sume.Uplate,
		Dospelo:      math.Max(sume.Dospelo-sume.Uplate, 0),
		Nedospelo:    sume.Nedospelo,
		UkRealizRob:  sume.DospeloRob + sume.NedospeloRob,
		DospeloRob:   sume.DospeloRob,
		NedospeloRob: sume.NedospeloRob,
	}
	if sume.Dospelo-sume.Uplate < 0 {
		st.Nedospelo = sume.Nedospelo - (sume.Uplate - sume.Dospelo)
	}
	st.UkDospelo = st.Dospelo + st.DospeloRob
	st.UkNedospelo = st.Nedospelo + st.NedospeloRob
	st.UkSaldo = st.UkDospelo + st.UkNedospelo
	return st
}

// RobKreditLimitPoruka returns why a faktura cannot be issued to the kupac ("" when it can), like the
// end of the legacy SaldoKUPCA: the ukupni saldo over the kredit limit, the dospeli dug over the limit
// of the dospeli dug.
func RobKreditLimitPoruka(kupac domain.RobKupacDto, st domain.RobnoFaktureFinStanjeKupca) string {
	switch {
	case kupac.Krlimit > 0 && kupac.Krlimit < st.UkSaldo:
		return "Prekoračen kredit limit."
	case kupac.Krlimitdosp > 0 && kupac.Krlimitdosp < st.UkDospelo:
		return "Prekoračen dospeli dug."
	}
	return ""
}

// GetRobAvansiDokumenta returns the avansi closed on the document (the legacy PuniTableAVANS) in one
// query: per avans its avansni račun, its iznos, what of it is still open after all the fakture it is
// closed on, and the iznos closed on this document.
func (s *RobnoDokumentaResource) GetRobAvansiDokumenta(ctx context.Context, rdokID int64) ([]domain.RobAvansDto, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.robno.AvansiDokumenta(ctx, rdokID)
}

// GetRobZiroRacuni returns the values of the žiro računi of the firm the documents show (the legacy
// PuniCOMBOS: "-" and the banke without nafakne, "broj računa  banka").
func (s *RobnoDokumentaResource) GetRobZiroRacuni(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	racuni, err := s.robno.BankeZiroRacuni(ctx)
	if err != nil {
		return nil, err
	}
	items := []domain.ComboItem{{Key: robZiroNijeIzabran, Value: robZiroNijeIzabran}}
	for _, r := range racuni {
		items = append(items, domain.ComboItem{Key: r, Value: r})
	}
	return items, nil
}

// robnoUnosNoviDokumentDefaults returns the defaults of a new document of the firm in one row: the
// domaća valuta (fvr.sifval; Key its šifra, C1 its idvalute, C2 its text) and the konto of the kupci
// (rvr.fkup, C3).
func (s *RobnoDokumentaResource) robnoUnosNoviDokumentDefaults(ctx context.Context) (domain.SearchRowDto, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.SearchRowDto{}, fmt.Errorf("no user session found")
	}
	entities, err := s.robno.NoviDokumentDefaults(ctx, userSession.Firma)
	if err != nil {
		return domain.SearchRowDto{}, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.SearchRowDto{}, nil
	}
	return (*entities)[0], nil
}

// robNullTime returns the time of a nullable date (zero when it is NULL).
func robNullTime(t sql.NullTime) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// robNullID returns the key of a šifarnik as a nullable id (NULL for "" and 0).
func robNullID(value string) sql.NullInt64 {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

// robFormatDatum formats a nullable date dd.mm.yyyy ("" when it is NULL).
func robFormatDatum(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(common.DateLayout)
}

// robFormatIznos formats an iznos of a grid with the locale of the system and 2 decimals.
func robFormatIznos(value float64) string {
	return common.FormatNumberWithSystemLocale(value, 2)
}

// ---------------------------------------------------------------------------------------------------
// Helpers of the entry of the documents
// ---------------------------------------------------------------------------------------------------

// robnoUnosDokvrsta returns the vrsta dokumenta of the period (its group, predznak, the flag of the
// POPDV and its POPDV polje); found is false when it does not exist.
func (s *RobnoDokumentaResource) robnoUnosDokvrsta(ctx context.Context, vrd int) (domain.Dokvrsta, bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Dokvrsta{}, false, fmt.Errorf("no user session found")
	}
	b := querybuilder.New(`select vrd, coalesce(opis, '') as opis, coalesce(grpdok, '') as grpdok, coalesce(predznak, '') as predznak,
		coalesce(flgpopdv, false) as flgpopdv, coalesce(polje, '') as polje from dokvrsta`).
		Where("god = ?", userSession.SelectedGod).
		Where("kar = ?", userSession.SelectedKar).
		Where("vrd = ?", vrd)
	entities, err := s.dokvrstaRepo.Select(ctx, b)
	if err != nil {
		return domain.Dokvrsta{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Dokvrsta{}, false, nil
	}
	return (*entities)[0], true, nil
}

// GetRobnoUnosNalog returns the robni nalog of the entry screens by its id (its vrsta, broj, magacin and
// totals; found false without it): the screens take the nalog from its id only.
func (s *RobnoDokumentaResource) GetRobnoUnosNalog(ctx context.Context, rnalID int64) (domain.Rnal, bool, error) {
	return s.robnoUnosNalog(ctx, rnalID)
}

// robnoUnosNalog returns the robni nalog (its vrsta, broj, datum, magacin and totals; found false without it).
func (s *RobnoDokumentaResource) robnoUnosNalog(ctx context.Context, rnalID int64) (domain.Rnal, bool, error) {
	if rnalID <= 0 {
		return domain.Rnal{}, false, nil
	}
	// The id comes from the request: the nalog must also be of the period of the session.
	b := querybuilder.New(`select rnalid, coalesce(tipdok, '') as tipdok, idtipdok, coalesce(nalog, 0) as nalog, danal, datob,
		opis, magaciniid, mag, dug, pot, brdo, brst, oper from rnal`).
		Where("rnal.rnalid = ?", rnalID)
	if err := s.rnalHeaderRepo.ApplyFilters(ctx, b, "rnal", nil, ""); err != nil {
		return domain.Rnal{}, false, err
	}
	entities, err := s.rnalHeaderRepo.Select(ctx, b)
	if err != nil {
		return domain.Rnal{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Rnal{}, false, nil
	}
	return (*entities)[0], true, nil
}

// robnoUnosPostojiID checks an optional key of a šifarnik (e.g. the org. jedinica): empty or "-" is no key (NULL),
// a given key must exist (postoji tells whether the key exists).
func (s *RobnoDokumentaResource) robnoUnosPostojiID(ctx context.Context, value, field string, postoji func(context.Context, int64) (bool, error),
	add func(field, message string)) {
	id, given, ok := robnoUnosID(value)
	switch {
	case !ok:
		add(field, robnoUnosMsgNePostoji)
	case !given:
		// No key: saved as NULL.
	default:
		if exists, err := postoji(ctx, id); err != nil {
			add(field, common.ErrMsgReadData)
		} else if !exists {
			add(field, robnoUnosMsgNePostoji)
		}
	}
}

// robnoUnosDatum parses a date of the form (the HTML date input, or dd.mm.yyyy).
func robnoUnosDatum(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{common.HtmlLayout, common.DateLayout} {
		if t, err := common.StringToTimeWithLayout(value, layout); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// robnoUnosDatumMsg is the error of a date that is not valid: required when it is empty.
func robnoUnosDatumMsg(value string) string {
	if strings.TrimSpace(value) == "" {
		return common.ErrMsgObavezanPodatak
	}
	return robnoUnosMsgNeispravanDatum
}

// robnoUnosID parses the key of a table combo: given is false for no value ("" or "-", saved as
// NULL), ok is false for a value that is not a key.
func robnoUnosID(value string) (id int64, given, ok bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return 0, false, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, true, false
	}
	return id, true, true
}

// robnoUnosBrojevi returns the list of the brojevi separated by commas (e.g. the avansni računi of a
// faktura) as the list of an sql "in", ok false when one of them is not a whole number.
func robnoUnosBrojevi(value string) ([]int, bool) {
	parts := strings.Split(value, ",")
	brojevi := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		brojevi = append(brojevi, n)
	}
	if len(brojevi) == 0 {
		return nil, false
	}
	return brojevi, true
}

// ---------------------------------------------------------------------------------------------------
// The stavke of the documents (the legacy window of the stavke, rob_FAKTURE_DET and the like)
// ---------------------------------------------------------------------------------------------------

// GetRobStavkeDokumenta returns the stavke of a robni dokument, the last one first (the legacy query of
// the grid of the stavke): the stavka with the naziv of its artikal (rsif).
func (s *RobnoDokumentaResource) GetRobStavkeDokumenta(ctx context.Context, rdokID int64) ([]domain.RobStavkaRow, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.robno.RproStavkeDokumenta(ctx, rdokID)
}

// RobStavkeTableRows returns the stavke as the rows of the grid of the stavke of the fakture (the columns
// redni broj, konto, šifra, naziv, JM, količina, magacinska cena, iznos, prodajna cena, rabat).
func RobStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Cena), robFormatIznos(st.Iznos),
				robFormatIznos(st.Fcena), robFormatIznos(st.Rab)},
		})
	}
	return rows
}

// robDokumentaPrikazFrom is the from of the query of the documents of the "Prikaz faktura" tab: the
// document with the naziv of its partner (fkpl) and the tip of a knjižno pismo (tipknpisma, by its key).
const robDokumentaPrikazFrom = ` from rdok r
		left join lateral (select f.naziv from fkpl f
			where f.god = r.god and f.kar = r.kar and f.konto = r.fkto and f.sifra = r.fana
			order by f.vkonta desc, f.idfkpl limit 1) kup on true
		left join tipknpisma tkp on tkp.tipknjid = r.tipknjid`

// robDokumentaPrikazSort are the columns of the grid of the documents the grid can be sorted by.
var robDokumentaPrikazSort = map[string]string{"dokum": "r.dokum", "dadok": "r.dadok"}

// GetRobDokumentaPrikaz returns a page of the robni dokumenti of the vrste in the magacin of the selected
// year and firm with the naziv of their partner, the last one first (the legacy query of the grid of the
// documents of the "Prikaz faktura" tab, with the columns of its grid only), and the number of all of
// them. The search text is looked for in the broj, the partner (konto, šifra, naziv), the opis and the
// izvorni dokument; sortBy (dokum or dadok) and sortOrder (ASC or DESC) sort the grid.
func (s *RobnoDokumentaResource) GetRobDokumentaPrikaz(ctx context.Context, magaciniID int, vrste []int, params domain.RobDokumentaPrikazParams) ([]domain.RobDokumentPrikazRow, int, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, fmt.Errorf("no user session found")
	}
	vrd := make([]any, 0, len(vrste))
	for _, v := range vrste {
		vrd = append(vrd, v)
	}
	b := querybuilder.New(`select r.rdokid, coalesce(r.rbr, 0)::int as rbr, coalesce(r.tipdok, '') as tipdok,
			coalesce(r.nalog, 0) as nalog, r.danal, coalesce(r.vrd, 0) as vrd, coalesce(r.dokum, 0) as dokum, r.dadok,
			coalesce(r.iznos, 0)::float8 as iznos, coalesce(r.fkto, '') as fkto, coalesce(r.fana, '') as fana,
			coalesce(kup.naziv, '') as naziv, coalesce(r.opis, '') as opis, coalesce(r.rok, 0)::int as rok, r.datiz, r.datob,
			coalesce(r.mag, 0)::int as mag, coalesce(r.dokiz, '') as dokiz, coalesce(r.mi, 0)::int as mi,
			coalesce(r.kamt, '') as kamt, coalesce(r.otp, '') as otp, coalesce(r.kom, 0)::int as kom,
			coalesce(r.oper, '') as oper,
			coalesce(concat_ws(' - ', tkp.sifrazlog::text, nullif(tkp.opis, '')), '') as tipknjpisma`+robDokumentaPrikazFrom).
		Where("r.god = ?", userSession.SelectedGod).
		Where("r.kar = ?", userSession.SelectedKar).
		AddIn("r.vrd", vrd).
		Where("r.magaciniid = ?", magaciniID).
		AddSearch([]string{"r.dokum", "r.fkto", "r.fana", "kup.naziv", "r.opis", "r.dokiz"}, strings.TrimSpace(params.Search))

	total, err := s.unos.Prikaz.Count(ctx, b)
	if err != nil {
		return nil, 0, err
	}

	if _, ok := robDokumentaPrikazSort[params.SortBy]; ok {
		order := "ASC"
		if strings.EqualFold(params.SortOrder, "DESC") {
			order = "DESC"
		}
		b.OrderByKey(params.SortBy, order, robDokumentaPrikazSort).OrderBy("r.rdokid " + order)
	} else {
		b.OrderBy("r.danal desc, r.tipdok desc, r.vrd desc, r.dadok desc, r.dokum desc, r.rdokid desc")
	}
	if params.PageSize > 0 {
		b.Limit(params.PageSize).Offset((max(params.Page, 1) - 1) * params.PageSize)
	}
	entities, err := s.unos.Prikaz.Select(ctx, b)
	if err != nil || entities == nil {
		return nil, total, err
	}
	return *entities, total, nil
}

// RobDokumentaPrikazTableRows returns the documents as the rows of the grid of the documents of the
// "Prikaz faktura" tab (the key of a row is its rdokid; the columns of GetFaktureDokumentaTableFields).
func RobDokumentaPrikazTableRows(dokumenta []domain.RobDokumentPrikazRow) []domain.TableRow {
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
			Fields: []string{d.Tipdok, broj(d.Nalog), robFormatDatum(d.Danal), broj(d.Vrd), broj(d.Dokum), robFormatDatum(d.Dadok),
				robFormatIznos(d.Iznos), d.Fkto, d.Fana, d.Naziv, d.Opis, broj(d.Rok), robFormatDatum(d.Datiz), robFormatDatum(d.Datob),
				broj(d.Mag), d.Dokiz, broj(d.Mi), d.Kamt, d.Otp, broj(d.Kom), d.Oper},
		})
	}
	return rows
}

// RobStavkePrikazTableRows returns the stavke as the rows of the grid of the stavke of the selected
// document of the "Prikaz faktura" tab (the columns of GetFaktureDokumentStavkeTableFields), by the redni
// broj like the legacy grid (GetRobStavkeDokumenta returns the last one first).
func RobStavkePrikazTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Cena), robFormatIznos(st.Rab), robFormatIznos(st.Iznos)},
		})
	}
	return rows
}

// GetRobStavkeKontekst returns what the stavke of a robni dokument need of their document, its nalog, its
// magacin and its vrsta, in one query (found is false without the document).
func (s *RobnoDokumentaResource) GetRobStavkeKontekst(ctx context.Context, rdokID int64) (domain.RobStavkeKontekst, bool, error) {
	entities, err := s.robno.RdokKontekst(ctx, rdokID)
	if err != nil || len(entities) == 0 {
		return domain.RobStavkeKontekst{}, false, err
	}
	return entities[0], true, nil
}

// GetRobArtikalStanje returns the artikal of the šifra (rsif) with its stanje in the konto of the magacin
// (rsta: ulaz - izlaz) and its veleprodajna cena; found is false when the artikal does not exist.
func (s *RobnoDokumentaResource) GetRobArtikalStanje(ctx context.Context, magaciniID int, konto string, sifra int) (domain.RobArtikalStanjeDto, bool, error) {
	entities, err := s.robno.RsifArtikalStanje(ctx, magaciniID, konto, sifra)
	if err != nil || len(entities) == 0 {
		return domain.RobArtikalStanjeDto{}, false, err
	}
	return entities[0], true, nil
}

// GetRobKontaMagacina returns the konta of the magacin (magkonto; the legacy sKONTOWHS and the seek of
// MAGKONTO): a stavka is in one of them.
func (s *RobnoDokumentaResource) GetRobKontaMagacina(ctx context.Context, magaciniID int) ([]string, error) {
	return s.robno.MagkontoKonta(ctx, magaciniID)
}

// GetRobZadnjaCena returns the last cena of the artikal to the kupac of the document (the legacy
// ZadnjaCena): the prodajna cena, the rabat and the datum of the last stavka of the kupac with the šifra
// in a document of the same vrsta up to the datum of the document. found is false when the kupac has not
// bought the artikal ("Kupac nije kupovao ovaj artikal").
func (s *RobnoDokumentaResource) GetRobZadnjaCena(ctx context.Context, rdokID int64, sifra int) (domain.RobZadnjaCenaDto, bool, error) {
	entities, err := s.robno.RproZadnjaCenaKupca(ctx, rdokID, sifra)
	if err != nil || len(entities) == 0 {
		return domain.RobZadnjaCenaDto{}, false, err
	}
	return entities[0], true, nil
}

// RobTaxCatCombo returns the poreske kategorije of a stavka and the default one (the legacy PuniTAXCAT):
// for a PDV obveznik the kategorije of the PDV (S20 by default), else only SS.
func RobTaxCatCombo(pdvObveznik bool) ([]domain.ComboItem, string) {
	if !pdvObveznik {
		return []domain.ComboItem{{Key: "SS", Value: "SS"}}, "SS"
	}
	kategorije := []string{"S10", "S20", "Z", "E", "AE10", "AE20", "R", "O", "OE", "SS", "N"}
	items := make([]domain.ComboItem, 0, len(kategorije))
	for _, k := range kategorije {
		items = append(items, domain.ComboItem{Key: k, Value: k})
	}
	return items, "S20"
}

// RobTaxExemCombo returns the osnovi oslobođenja of a stavka and the default one (the legacy
// PuniTAXEXEM): "-" (none) and PDV-RS-OBJ.
func RobTaxExemCombo() ([]domain.ComboItem, string) {
	return []domain.ComboItem{{Key: "-", Value: "-"}, {Key: "PDV-RS-OBJ", Value: "PDV-RS-OBJ"}}, "-"
}

// RobObracunStavke returns the iznos of a stavka, its neto prodajna cena and the iznos of its rabat (the
// legacy obracunaj): the količina by the cena, the cena less the rabat, the rabat of the iznos.
func RobObracunStavke(kolicina, cena, rabat float64) (iznos, netoCena, iznosRabata float64) {
	iznos = kolicina * cena
	netoCena = cena - rabat*cena/100
	iznosRabata = rabat * iznos / 100
	return iznos, netoCena, iznosRabata
}

// RobWinDevDatum returns the WinDev integer date of a date (the days from 1.1.1800, the legacy
// ConvertDateToInteger; 0 for no date).
func RobWinDevDatum(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return int64(t.Sub(time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)
}

// ValidateRobStavka validates a stavka of a robni dokument before it is saved (the legacy Validacija of
// the window of the stavke). Only the checks that end in an error of a field are done (the questions of
// the legacy - the zalihe not enough for a document not at the nabavna cena, the cena under the
// kalkulisana or under the nabavna, the stavka without the OTK - do not stop the saving):
//   - the konto is required and is a konto of the magacin (magkonto);
//   - the šifra is required and the artikal exists;
//   - the količina is required; without enough zalihe (a document not storno, an artikal that is not a
//     usluga) a document at the nabavna cena is not saved;
//   - the cena is more than 0, the nabavna cena too for a document at the nabavna cena;
//   - with the rabat the neto prodajna cena is not more than the cena.
func (s *RobnoDokumentaResource) ValidateRobStavka(ctx context.Context, params domain.RobStavkaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil {
		add("sifraartikla", common.ErrMsgReadData)
		return fieldErrors
	}
	if !found {
		add("sifraartikla", "Dokument ne postoji.")
		return fieldErrors
	}

	// The konto: a konto of the magacin.
	konto := strings.TrimSpace(params.Konto)
	if konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	} else if konta, err := s.GetRobKontaMagacina(ctx, k.MagaciniID); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if !robSadrzi(konta, konto) {
		add("kontostavke", "Konto nije otvoren za magacin.")
	}

	// The artikal.
	var artikal domain.RobArtikalStanjeDto
	artikalOK := false
	sifraText := strings.TrimSpace(params.Sifra)
	if sifraText == "" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if sifra, err := strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if artikal, artikalOK, err = s.GetRobArtikalStanje(ctx, k.MagaciniID, konto, sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !artikalOK {
		add("sifraartikla", robnoUnosMsgNePostoji)
	}

	// The količina and the zalihe.
	kolicina := common.StringToFloat64(params.Kolicina)
	if kolicina == 0 {
		add("kolicina", common.ErrMsgObavezanPodatak)
	} else if artikalOK && sifraText != "0" && strings.TrimSpace(artikal.Tip) != "U" && !k.Storno() && k.NabavnaCena() {
		if (kolicina > 0 && artikal.Stanje <= 0) || math.Round((artikal.Stanje-kolicina)*1000)/1000 < 0 {
			add("kolicina", "Nema dovoljno zaliha.")
		}
	}

	// The cene and the rabat.
	cena := common.StringToFloat64(params.Cena)
	if cena <= 0 {
		add("prodajnacena", "Mora biti veća od 0.")
	}
	if k.NabavnaCena() && common.StringToFloat64(params.NabavnaCena) <= 0 {
		add("nabavnacena", "Mora biti veća od 0.")
	}
	rabat := common.StringToFloat64(params.Rabat)
	if _, neto, _ := RobObracunStavke(kolicina, cena, rabat); rabat != 0 && neto > cena {
		add("rabatstavke", "Neto cena veća od prodajne.")
	}
	return fieldErrors
}

// SaveRobStavka saves a stavka of a robni dokument, a new one (RproID 0) or a changed one (the legacy
// CreateRpro of the window of the stavke, SaveRPRO), in one transaction:
//   - the artikal gets the dani of the stavka (rsif.dani);
//   - the stavka has the columns of its document (nalog, vrsta, broj, datum, kupac, magacin), the
//     artikal (rsifid, JM), the količina and the iznos (negative for a storno; the cena is the nabavna one
//     for a document at the nabavna cena), the prodajna cena (fcena) and the cena in the valuta of the
//     document (by its kurs), the poreska kategorija and the osnov oslobođenja; a new stavka gets the next
//     redni broj;
//   - a new artikal of the cene (rcene) is added with the cena, the nabavna cena and the rabat;
//   - the nalog, the stanje of the artikal and the zalihe by OTK, serija and rok (the magacini with the
//     tip zaliha 1 and 2) get the iznos and the količina (AzurirajRNAL, AzurirajRSTA, AzurirajDRSTA); a
//     changed stavka is taken back from the stanje first;
//   - the document gets the iznos of the stavka and a new stavka counts in it (rdok.iznos, rdok.brst).
//
// Then the rate and the obračun of the document are computed again (PreracunRATA, ObracunFAKT). The
// xdat/xop columns are written by the repository; the kontrolna datoteka SR is not in this database.
func (s *RobnoDokumentaResource) SaveRobStavka(ctx context.Context, params domain.RobStavkaParams) (rproID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if rproID, err = s.SaveRobStavkaTx(ctx, tx, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	// The rate and the obračun of the document (its avansi).
	if err := s.PreracunRataDokumenta(ctx, params.RdokID); err != nil {
		return rproID, err
	}
	return rproID, s.ObracunFakture(ctx, params.RdokID)
}

// SaveRobStavkaTx saves a stavka of a robni dokument in the transaction of the caller (see SaveRobStavka,
// which commits it and then computes the rate and the obračun of the document).
func (s *RobnoDokumentaResource) SaveRobStavkaTx(ctx context.Context, tx db.Transaction, params domain.RobStavkaParams) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("dokument koji obrađujete ne postoji")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.TrimSpace(params.Konto)
	kolicina := common.StringToFloat64(params.Kolicina)
	fcena := common.StringToFloat64(params.Cena)
	nabavna := common.StringToFloat64(params.NabavnaCena)
	cena := fcena
	if k.NabavnaCena() {
		cena = nabavna
	}
	iznos := math.Round(math.Abs(kolicina*cena)*100) / 100
	kolicina = math.Abs(kolicina)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}
	cenaVal := fcena
	if k.Kurs > 0 && k.Vaziza > 0 {
		cenaVal = math.Round(fcena/k.Kurs/k.Vaziza*10000) / 10000
	}
	taxexreco := strings.TrimSpace(params.Taxexreco)
	if taxexreco == "-" {
		taxexreco = ""
	}

	// The artikal: its dani (written when they change); its rsifid and JM for the stavka.
	rsifID, jm, dani, err := s.robno.RsifIDJmDani(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if noviDani := common.StringToInt(params.Dani); noviDani != dani {
		if err = s.unos.Rsif.UpdateWithTx(ctx, tx, &domain.RobRsifDto{}, "rsifid", rsifID,
			[]domain.Fields{{Name: "dani", Value: strconv.Itoa(noviDani)}}); err != nil {
			return 0, err
		}
	}

	// A changed stavka: its old values are taken back from the stanje (the nalog and the document get
	// the difference of the iznos).
	stara := domain.RobPromenaStanja{}
	if params.RproID > 0 {
		if stara, err = s.robno.RproStaraPromena(ctx, tx, params.RproID, params.RdokID); err != nil {
			return 0, err
		}
	}

	rok, _ := robnoUnosDatum(params.Rok)
	polja := []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
		{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
		{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
		robPoljeNullTime("dadok", k.Dadok),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		robPoljeText("konto", konto),
		{Name: "jm", Value: jm},
		robPoljeFloat("cena", cena),
		robPoljeFloat("ncena", nabavna),
		robPoljeFloat("fcena", fcena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		robPoljeBroj("rab", params.Rabat),
		robPoljeInt("po", params.Tarifa),
		robPoljeInt("dani", params.Dani),
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)},
		robPoljeFloat("cenaval", cenaVal),
		robPoljeText("taxcat", params.Taxcat),
		{Name: "taxexreco", Value: taxexreco},
	}
	if params.RproID > 0 {
		rproID = params.RproID
		if err = s.unos.Rpro.UpdateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", rproID, polja); err != nil {
			return 0, err
		}
	} else {
		var rbr int
		if rbr, err = s.robno.RproSledeciRbr(ctx, tx, params.RdokID); err != nil {
			return 0, err
		}
		polja = append(polja,
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""},
		)
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja); err != nil {
			return 0, err
		}
	}

	// A new artikal of the cene.
	cene, err := s.robno.RcenePostoji(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if !cene {
		if _, err = s.unos.Rcene.CreateWithTx(ctx, tx, &domain.RobRceneDto{}, "rceneid", []domain.Fields{
			{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
			{Name: "sifra", Value: strconv.Itoa(sifra)},
			{Name: "mma", Value: "0"}, {Name: "vra", Value: "0"}, {Name: "mra", Value: "0"}, {Name: "ztro", Value: "0"},
			robPoljeFloat("cena", fcena), robPoljeFloat("fcena", fcena),
			{Name: "mcena", Value: "0"}, {Name: "mcenap", Value: "0"},
			robPoljeFloat("ncena", nabavna), {Name: "ztrin", Value: "0"},
			robPoljeBroj("rab", params.Rabat), {Name: "tz", Value: ""}, {Name: "tzi", Value: ""},
		}); err != nil {
			return 0, err
		}
	}

	// The nalog gets the difference of the iznos; the stanje of the artikal and the zalihe by OTK, serija
	// and rok (the magacini with the tip zaliha 1 and 2): the old stavka taken back, the new one added (the
	// legacy AzurirajRNAL, AzurirajRSTA and AzurirajDRSTA; the cene 0, TZ and TZI "-").
	razlikaIznosa := iznos - stara.Iznos
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), razlikaIznosa); err != nil {
		return 0, err
	}
	promena := func(sifra int, konto string, kolicina, iznos float64, otk, serija string, roktr int64, brst int) domain.RobPromenaStanja {
		return domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra, Kolicina: kolicina, Iznos: iznos,
			Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
			Otk: strings.TrimSpace(otk), Serija: strings.TrimSpace(serija), Roktr: roktr, Brst: brst}
	}
	promene := []domain.RobPromenaStanja{}
	if params.RproID > 0 {
		promene = append(promene, promena(stara.Sifra, stara.Konto, -stara.Kolicina, -stara.Iznos, stara.Otk, stara.Serija, stara.Roktr, -1))
	}
	promene = append(promene, promena(sifra, konto, kolicina, iznos, params.Otk, params.Serija, RobWinDevDatum(rok), 1))
	for _, p := range promene {
		if err = s.robAzurirajStanjeStavke(ctx, tx, k, p); err != nil {
			return 0, err
		}
	}

	// The document: its iznos, a new stavka counts in it.
	dokIznos, dokBrst, err := s.robno.RdokIznosBrst(ctx, tx, params.RdokID)
	if err != nil {
		return 0, err
	}
	if params.RproID == 0 {
		dokBrst++
	}
	if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", params.RdokID, []domain.Fields{
		robPoljeFloat("iznos", math.Round((dokIznos+razlikaIznosa)*100)/100), {Name: "brst", Value: strconv.Itoa(dokBrst)},
	}); err != nil {
		return 0, err
	}
	return rproID, nil
}

// AzurirajRNAL updates the nalog of the stavke of a robni dokument (the legacy
// GlobalProcROBNO.AzurirajRNAL, shared by the documents): the iznos of a stavka on its duguje (dugovno)
// and / or its potražuje (potražno), the broj stavki as the count of the stavke of the nalog. It runs in
// the transaction of the save of the stavka (which sees the stavka already).
func (s *RobnoDokumentaResource) AzurirajRNAL(ctx context.Context, tx db.Transaction, rnalID int64, dugovno, potrazno bool, iznos float64) error {
	brst, err := s.robno.RproBrojNaloga(ctx, tx, rnalID)
	if err != nil {
		return err
	}
	dug, pot, err := s.robno.RnalDugPot(ctx, tx, rnalID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if dugovno {
		dug += iznos
	}
	if potrazno {
		pot += iznos
	}
	return s.rnalHeaderRepo.UpdateWithTx(ctx, tx, &domain.Rnal{}, "rnalid", rnalID, []domain.Fields{
		robPoljeFloat("dug", math.Round(dug*100)/100), robPoljeFloat("pot", math.Round(pot*100)/100),
		{Name: "brst", Value: strconv.Itoa(brst)},
	})
}

// robKnjizenje is what the change of the stanje needs of the vrsta dokumenta: its knjiženje (dokvrsta.kodknj,
// "D" dugovno / ulaz, "P" potražno / izlaz), its group and its predznak.
type robKnjizenje struct {
	kodknj, grpdok, predznak string
}

// robDokvrstaKnjizenje returns the knjiženje of the vrsta dokumenta (dokvrstaid; empty without it).
func (s *RobnoDokumentaResource) robDokvrstaKnjizenje(ctx context.Context, tx db.Transaction, dokvrstaID int64) (robKnjizenje, error) {
	k, err := s.robno.DokvrstaKnjizenjePoID(ctx, tx, dokvrstaID)
	if err != nil {
		return robKnjizenje{}, err
	}
	return robKnjizenje{kodknj: strings.ToUpper(strings.TrimSpace(k.Kodknj)), grpdok: strings.TrimSpace(k.Grpdok),
		predznak: strings.TrimSpace(k.Predznak)}, nil
}

// robGrupeAzuriranjaCene returns the groups of the vrste dokumenta that update the cene of the stanje: the
// basic ones and the ones of the firm (initable: odeljak GRUPADOK, ključ AZURIRAJUCENU, the legacy
// INITABLE.DEFVREDNOST separated by ";").
func (s *RobnoDokumentaResource) robGrupeAzuriranjaCene(ctx context.Context, tx db.Transaction, osnovne ...string) ([]string, error) {
	dodatne, err := s.robno.IniVrednostTx(ctx, tx, "GRUPADOK", "AZURIRAJUCENU")
	if err != nil {
		return nil, err
	}
	grupe := append([]string{}, osnovne...)
	for _, g := range strings.Split(dodatne, ";") {
		if g = strings.TrimSpace(g); g != "" {
			grupe = append(grupe, g)
		}
	}
	return grupe, nil
}

// robStanjeMesec returns the month of the date of the change of the stanje (1..12; January without a
// date).
func robStanjeMesec(t time.Time) int {
	if t.IsZero() {
		return 1
	}
	return int(t.Month())
}

// robStanjeKolone are the columns and the values of a new row of the stanje (rsta, drsta): the kretanje of
// the month by the knjiženje (D: dug, ulaz; P: pot, izlaz; the početno stanje for the vrsta naloga "00")
// and the NOT NULL columns of the months of the potražuje at 0.
func robStanjeKolone(p domain.RobPromenaStanja, k robKnjizenje, m int) map[string]any {
	kolone := map[string]any{}
	for i := 1; i <= 12; i++ {
		kolone[fmt.Sprintf("mpot_%d", i)] = 0
	}
	kolone["dug"], kolone["pot"], kolone["ulaz"], kolone["izlaz"] = 0, 0, 0, 0
	switch k.kodknj {
	case "D":
		kolone["dug"], kolone[fmt.Sprintf("mdug_%d", m)] = p.Iznos, p.Iznos
		kolone["ulaz"], kolone[fmt.Sprintf("mul_%d", m)] = p.Kolicina, p.Kolicina
		if strings.TrimSpace(p.Tipdok) == "00" {
			kolone["pov"], kolone["pos"] = p.Iznos, p.Kolicina
		}
	case "P":
		kolone["pot"], kolone[fmt.Sprintf("mpot_%d", m)] = p.Iznos, p.Iznos
		kolone["izlaz"], kolone[fmt.Sprintf("miz_%d", m)] = p.Kolicina, p.Kolicina
	}
	return kolone
}

// robStanjeKretanje returns the changes of the kretanje of the month of an existing row of the stanje
// (see robStanjeKolone): each column gets the value added.
func robStanjeKretanje(p domain.RobPromenaStanja, k robKnjizenje, m int) []repository.Izmena {
	var izmene []repository.Izmena
	add := func(kolona string, vrednost float64) {
		izmene = append(izmene, repository.Izmena{Kolona: kolona, Dodaj: true, Vrednost: vrednost})
	}
	switch k.kodknj {
	case "D":
		add("dug", p.Iznos)
		add(fmt.Sprintf("mdug_%d", m), p.Iznos)
		add("ulaz", p.Kolicina)
		add(fmt.Sprintf("mul_%d", m), p.Kolicina)
		if strings.TrimSpace(p.Tipdok) == "00" {
			add("pov", p.Iznos)
			add("pos", p.Kolicina)
		}
	case "P":
		add("pot", p.Iznos)
		add(fmt.Sprintf("mpot_%d", m), p.Iznos)
		add("izlaz", p.Kolicina)
		add(fmt.Sprintf("miz_%d", m), p.Kolicina)
	}
	return izmene
}

// AzurirajRSTA changes the stanje of an artikal in a konto of a magacin (rsta) by a stavka (the legacy
// GlobalProcROBNO.AzurirajRSTA, shared by the documents), in the transaction of the save:
//   - a new stanje gets the kretanje of the month by the knjiženje of the vrsta (see robStanjeKolone), the
//     cena (when more than 0) and the other cene of the stavka (the prosečna nabavna cena for the popis),
//     the datum of the first kretanje;
//   - an existing one adds the kretanje and the broj stavki (Brst); the vrste of the groups that update the
//     cene (POP, PLT, NIV, ULS, KAL and the ones of the firm) with the predznak "+" set its cene that are
//     not 0 (for the popis only the ones that are not 0 of VMA, MMA and the rabat, else all three).
func (s *RobnoDokumentaResource) AzurirajRSTA(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	k, err := s.robDokvrstaKnjizenje(ctx, tx, p.DokvrstaID)
	if err != nil {
		return err
	}
	m := robStanjeMesec(p.Danal)
	rstaID, err := s.robno.RstaIDZaIzmenu(ctx, tx, p.MagaciniID, p.Konto, p.Sifra)
	if errors.Is(err, sql.ErrNoRows) {
		rsifID, mag, err := s.robno.RsifIDMag(ctx, tx, p.Sifra, p.MagaciniID)
		if err != nil {
			return err
		}
		kolone := robStanjeKolone(p, k, m)
		kolone["rsifid"], kolone["sifra"], kolone["konto"] = rsifID, p.Sifra, p.Konto
		kolone["mag"], kolone["magaciniid"] = mag, p.MagaciniID
		if !p.Danal.IsZero() {
			kolone["zdat"] = p.Danal.Format(common.HtmlLayout)
		}
		if p.Cena > 0 {
			kolone["cena"] = p.Cena
		}
		kolone["brst"] = 1
		kolone["vma"], kolone["mma"], kolone["ztro"] = p.Vma, p.Mma, p.Ztro
		kolone["fcena"], kolone["mcena"], kolone["ncena"] = p.Fcena, p.Mcena, p.Ncena
		if k.grpdok == "POP" {
			kolone["prosnc"] = p.Ncena
		}
		kolone["rab"], kolone["ztrin"], kolone["mcenap"] = p.Rab, p.Ztrin, p.Mcenap
		kolone["tz"], kolone["tzi"] = p.Tz, p.Tzi
		return s.robno.StanjeDodaj(ctx, tx, repository.TabelaRsta, kolone)
	}
	if err != nil {
		return err
	}

	izmene := robStanjeKretanje(p, k, m)
	setCol := func(kolona string, vrednost any) {
		izmene = append(izmene, repository.Izmena{Kolona: kolona, Vrednost: vrednost})
	}
	izmene = append(izmene, repository.Izmena{Kolona: "brst", Dodaj: true, Vrednost: p.Brst})
	grupe, err := s.robGrupeAzuriranjaCene(ctx, tx, "POP", "PLT", "NIV", "ULS", "KAL")
	if err != nil {
		return err
	}
	if robSadrzi(grupe, k.grpdok) && k.predznak == "+" {
		if p.Cena != 0 {
			setCol("cena", p.Cena)
		}
		if p.Fcena != 0 {
			setCol("fcena", p.Fcena)
		}
		if p.Mcena != 0 {
			setCol("mcena", p.Mcena)
		}
		if p.Ncena != 0 {
			setCol("ncena", p.Ncena)
		}
		if k.grpdok == "POP" {
			if p.Ncena != 0 {
				setCol("prosnc", p.Ncena)
			}
			if p.Vma != 0 {
				setCol("vma", p.Vma)
			}
			if p.Mma != 0 {
				setCol("mma", p.Mma)
			}
			if p.Rab != 0 {
				setCol("rab", p.Rab)
			}
		} else {
			setCol("vma", p.Vma)
			setCol("mma", p.Mma)
			setCol("rab", p.Rab)
		}
		if p.Mcenap != 0 {
			setCol("mcenap", p.Mcenap)
		}
		setCol("ztro", p.Ztro)
		setCol("ztrin", p.Ztrin)
		setCol("tz", p.Tz)
		setCol("tzi", p.Tzi)
	}
	return s.robno.StanjeIzmeni(ctx, tx, repository.TabelaRsta, rstaID, izmene)
}

// AzurirajDRSTA changes the zalihe of an artikal by lot (OTK, serija, rok trajanja) in a konto of a
// magacin (drsta) by a stavka (the legacy GlobalProcROBNO.AzurirajDRSTA, shared by the documents, for the
// magacini with the tip zaliha 1 and 2), in the transaction of the save:
//   - a new lot gets the kretanje of the month by the knjiženje of the vrsta (see robStanjeKolone) and the
//     cena as its cena and nabavna cena (when it is not 0);
//   - an existing one adds the kretanje; the vrste of the groups that update the cene (POP, PLT, NIV, ULS
//     and the ones of the firm) set its cena and nabavna cena (and the prosečna nabavna cena for the
//     popis) when the cena is not 0.
func (s *RobnoDokumentaResource) AzurirajDRSTA(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	k, err := s.robDokvrstaKnjizenje(ctx, tx, p.DokvrstaID)
	if err != nil {
		return err
	}
	m := robStanjeMesec(p.Danal)
	rsifID, mag, err := s.robno.RsifIDMag(ctx, tx, p.Sifra, p.MagaciniID)
	if err != nil {
		return err
	}
	rstaID, err := s.robno.DrstaIDZaIzmenu(ctx, tx, p.Konto, p.Sifra, mag, p.Otk, p.Serija, p.Roktr)
	if errors.Is(err, sql.ErrNoRows) {
		kolone := robStanjeKolone(p, k, m)
		kolone["rsifid"], kolone["sifra"], kolone["konto"] = rsifID, p.Sifra, p.Konto
		kolone["otk"], kolone["serija"], kolone["roktr"] = p.Otk, p.Serija, p.Roktr
		kolone["mag"], kolone["magaciniid"] = mag, p.MagaciniID
		kolone["zstatus"] = ""
		if p.Cena != 0 {
			kolone["cena"], kolone["ncena"] = p.Cena, p.Cena
		}
		return s.robno.StanjeDodaj(ctx, tx, repository.TabelaDrsta, kolone)
	}
	if err != nil {
		return err
	}

	izmene := robStanjeKretanje(p, k, m)
	grupe, err := s.robGrupeAzuriranjaCene(ctx, tx, "POP", "PLT", "NIV", "ULS")
	if err != nil {
		return err
	}
	if robSadrzi(grupe, k.grpdok) && p.Cena != 0 {
		izmene = append(izmene, repository.Izmena{Kolona: "cena", Vrednost: p.Cena}, repository.Izmena{Kolona: "ncena", Vrednost: p.Cena})
		if k.grpdok == "POP" {
			izmene = append(izmene, repository.Izmena{Kolona: "prosnc", Vrednost: p.Cena})
		}
	}
	return s.robno.StanjeIzmeni(ctx, tx, repository.TabelaDrsta, rstaID, izmene)
}

// robAzurirajStanjeStavke changes the stanje of the artikal of a stavka (AzurirajRSTA) and, in the magacini
// with the tip zaliha 1 and 2, its zalihe by lot (AzurirajDRSTA): the tip zaliha 1 without the lot (the
// legacy sOTK, sSERIJA and nROKTR are empty), the tip zaliha 2 by the OTK, the serija and the rok of the
// stavka.
func (s *RobnoDokumentaResource) robAzurirajStanjeStavke(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, p domain.RobPromenaStanja) error {
	if err := s.AzurirajRSTA(ctx, tx, p); err != nil {
		return err
	}
	switch k.Tipzal {
	case 1:
		p.Otk, p.Serija, p.Roktr = "", "", 0
		return s.AzurirajDRSTA(ctx, tx, p)
	case 2:
		return s.AzurirajDRSTA(ctx, tx, p)
	}
	return nil
}

// GetRobStavkaArtikal returns the artikal of a stavka of the document with what the form of the stavka
// shows of it (the legacy exit of EDT_SIFRA and the choice of the artikal):
//   - the artikal by its šifra, or by its barkod for a value longer than 6 characters; its naziv, JM,
//     tarifa with its stopa on the datum of the document, dani, and its konto (else the konto given);
//   - its stanje and cene in the konto of the magacin (rsta): the veleprodajna cena, the magacinska cena
//     and the nabavna cena of the way of the zalihe (nacvodzal 1 the nabavna, 3 the prosečna nabavna);
//     a proizvod (tip P) at the cena of the proizvoda (rcene.pcena); the nabavna cena 0 is entered;
//   - the poreska kategorija of its tarifa (1: S20, 2: S10) and the last cena to the kupac.
//
// found is false when the artikal does not exist.
func (s *RobnoDokumentaResource) GetRobStavkaArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobStavkaArtikal, bool, error) {
	a := domain.RobStavkaArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	rows, err := s.robno.RsifArtikalStavke(ctx, len(vrednost) > 6, vrednost, konto, robNullTime(k.Dadok))
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm = common.StringToInt(r.Key), r.C1, r.C2
	a.Tarifa, a.Dani, a.Konto = common.StringToInt(r.C3), common.StringToInt(r.C4), r.C5
	a.Stopa = common.StringToFloat64(r.C7)
	switch a.Tarifa {
	case 1:
		a.Taxcat = "S20"
	case 2:
		a.Taxcat = "S10"
	}

	// The stanje and the cene in the konto of the magacin.
	st, stFound, err := s.robno.RstaStanjeStavke(ctx, k.MagaciniID, a.Konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if stFound {
		a.Stanje, a.Cena, a.MagCena = st.Stanje, st.Vpcena, st.Cena
		switch {
		case k.VpCena():
			a.NabavnaCena = st.Ncena
		case k.NabavnaCena():
			a.NabavnaCena = st.Prosnc
		}
	}
	if strings.TrimSpace(r.C6) == "P" {
		a.Cena = common.StringToFloat64(r.C8)
		a.NabavnaCena = a.Cena
	}
	a.NabavnaUnos = a.NabavnaCena == 0

	// The last cena to the kupac.
	zadnja, ima, err := s.GetRobZadnjaCena(ctx, rdokID, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if ima {
		a.ZadnjaCena, a.ZadnjiRabat = zadnja.Fcena, zadnja.Rab
		if zadnja.Dadok.Valid {
			a.ZadnjiDatum = zadnja.Dadok.Time.Format(common.HtmlLayout)
		}
	} else {
		a.Poruka = "Kupac nije kupovao ovaj artikal..."
	}
	return a, true, nil
}

// GetRobStavkaForma returns a saved stavka of the document as the form of the stavke shows it (the legacy
// "Izmene"): its values, the naziv and the stopa of its artikal, its stanje in the konto of the magacin.
// found is false when the stavka does not exist.
func (s *RobnoDokumentaResource) GetRobStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobStavkaForma, bool, error) {
	f := domain.RobStavkaForma{}
	rows, err := s.robno.RproStavkaForma(ctx, rproID, rdokID)
	if err != nil || rows == nil || len(*rows) == 0 {
		return f, false, err
	}
	r := (*rows)[0]
	f.RproID, f.Sifra, f.Naziv, f.Konto, f.Jm = int64(common.StringToInt(r.Key)), common.StringToInt(r.C1), r.C2, r.C3, r.C4
	brojevi := strings.Split(r.C5, "|")
	broj := func(i int) float64 {
		if i < len(brojevi) {
			return common.StringToFloat64(brojevi[i])
		}
		return 0
	}
	f.Kolicina, f.Cena, f.Rabat = broj(0), broj(1), broj(2)
	f.Tarifa, f.Dani = int(broj(3)), int(broj(4))
	f.NabavnaCena, f.MagCena = broj(5), broj(6)
	if rok := int64(broj(7)); rok != 0 {
		f.Rok = robnoStampaWinDevDatumFormat(rok, common.HtmlLayout)
	}
	f.Stopa, f.Stanje = broj(8), broj(9)
	f.Otk, f.Serija, f.Taxcat = r.C6, r.C7, r.C8
	taxexreco, err := s.robno.RproTaxexreco(ctx, rproID)
	if err != nil {
		return f, true, err
	}
	f.Taxexreco = taxexreco
	if f.Taxexreco == "" {
		f.Taxexreco = "-"
	}
	return f, true, nil
}

// DeleteRobStavka deletes a stavka of a robni dokument (by its rproid; its document is the one of the
// stavka, returned for the screen) in one transaction (the legacy "Brisanje", BrisiRPRO;
// TODO: confirm with the legacy procedure): the stavka is taken back from its nalog, from the stanje of
// the artikal and from the zalihe by lot (like the legacy "Izmene"), the document loses its iznos and
// counts one stavka less; then the rate and the obračun of the document are computed again.
func (s *RobnoDokumentaResource) DeleteRobStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if rdokID, err = s.DeleteRobStavkaTx(ctx, tx, rproID); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	if err := s.PreracunRataDokumenta(ctx, rdokID); err != nil {
		return rdokID, err
	}
	return rdokID, s.ObracunFakture(ctx, rdokID)
}

// DeleteRobStavkaTx deletes a stavka of a robni dokument (by its rproid) in the transaction of the caller
// and returns its document, rpro.rdokid (see DeleteRobStavka, which commits it and then computes the rate
// and the obračun of the document).
func (s *RobnoDokumentaResource) DeleteRobStavkaTx(ctx context.Context, tx db.Transaction, rproID int64) (int64, error) {
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
	stara, err := s.robno.RproStaraPromena(ctx, tx, rproID, rdokID)
	if err != nil {
		return 0, err
	}
	sifra, konto, kolicina, iznos, otk, serija, roktr := stara.Sifra, stara.Konto, stara.Kolicina, stara.Iznos, stara.Otk, stara.Serija, stara.Roktr
	if err = s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos); err != nil {
		return 0, err
	}
	if err = s.robAzurirajStanjeStavke(ctx, tx, k, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra,
		Kolicina: -kolicina, Iznos: -iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Otk: otk, Serija: serija, Roktr: roktr, Brst: -1}); err != nil {
		return 0, err
	}
	dokIznos, dokBrst, err := s.robno.RdokIznosBrst(ctx, tx, rdokID)
	if err != nil {
		return 0, err
	}
	if dokBrst > 0 {
		dokBrst--
	}
	if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", rdokID, []domain.Fields{
		robPoljeFloat("iznos", math.Round((dokIznos-iznos)*100)/100), {Name: "brst", Value: strconv.Itoa(dokBrst)},
	}); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// robSadrzi tells whether the list has the value.
func robSadrzi(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

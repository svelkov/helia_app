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
)

// ---------------------------------------------------------------------------------------------------
// Profakture (the group PRO, the legacy rob_ProFakture with its stavke rob_PROFAKTURE_DET): a robni dokument
// (rdok) of a kupac like the faktura veleprodaje, without the avansi, the rate and the obračun of the faktura
// ---------------------------------------------------------------------------------------------------

// robProUsloviPlacanja are the uslovi plaćanja of a new profaktura (the content of the legacy EDT_PLA).
const robProUsloviPlacanja = "VIRMANOM"

// RobProStavkeTableRows returns the stavke of a profaktura as the rows of the grid of the stavke (the columns of
// GetProStavkeTableFields, the legacy QRY_RRPO_DET: redni broj, konto, šifra, naziv, JM, količina, cena,
// rabat, iznos), by the redni broj. The naziv is the one of the stavka (rpro.naz1, typed for the šifra 0),
// else the one of the artikal (see RproStavkeDokumenta).
func RobProStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
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

// GetRobProDokument returns the document of the "Profakture" screen: the document of the params like the one of
// the faktura (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one with the
// next broj, NoviBrojRDOK; a new one gets the konto of the kupci, rvr.fkup, and the domaća valuta with the kurs
// and the važi za 1) with the opis of its vrsta and the domaća valuta of the firm. A saved document has its
// profaktura and broj fiskalnog isečka; a new one is of today (the legacy EDT_DADOK = Today()) without the
// datum izvornog dokumenta (Reset) and with the uslovi plaćanja "VIRMANOM". The stavke repeat the kupac (the
// legacy EDT_KUPAC).
func (s *RobnoDokumentaResource) GetRobProDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoProDokument, error) {
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoProDokument{}, err
	}
	res := domain.RobnoProDokument{
		Header: domain.RobnoProHeaderView{RobnoFaktureHeaderView: dok.Header},
		Stavka: domain.RobnoProStavka{RobnoFaktureStavka: dok.Stavka},
		Poruka: dok.Poruka,
	}
	h := &res.Header
	if dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(h.Vrd)); err != nil {
		return res, err
	} else if found {
		h.VrstaOpis = dokvrsta.Opis
	}
	defaults, err := s.robnoUnosNoviDokumentDefaults(ctx)
	if err != nil {
		return res, err
	}
	h.DomacaValuta = common.StringToInt(defaults.Key)
	if h.RdokID > 0 {
		if h.Profaktura, h.BrojIsecka, err = s.robno.RdokProfakBi(ctx, h.RdokID); err != nil {
			return res, err
		}
	} else {
		h.DatumDokum = time.Now()
		h.Dospeva = h.DatumDokum
		h.IzvorniDokDatum = time.Time{}
		h.UsloviPlacanja = robProUsloviPlacanja
		res.Stavka.DatumDokumenta = h.DatumDokum.Format(common.DateLayout)
	}
	if strings.TrimSpace(h.KupacNaziv) != "" {
		res.Stavka.Kupac = strings.TrimSpace(fmt.Sprintf("%s %s %s ;%s ;%s %s", h.KupacKonto, h.KupacSifra, h.KupacNaziv,
			h.KupacAdresa, h.KupacPostanskiBroj, h.KupacMesto))
	}
	return res, nil
}

// ValidateRobPro validates the header of a profaktura before it is saved, like the legacy ValidacijaRDOK of
// rob_ProFakture:
//   - the datum dokumenta is a valid date, not after the datum naloga;
//   - the konto and the šifra of the kupac are required and the kupac exists (fkpl); the mesto isporuke is one
//     of the kupac;
//   - the org. jedinica is required; a given komercijalista exists;
//   - the profaktura (the brojevi of the avansni računi, vrsta 172) are avansni računi of the kupac;
//   - the šifra valute is required.
//
// The broj dokumenta is a whole number and no other nalog has the document of the vrsta and broj in the magacin
// (the legacy exit of EDT_DOKUM); the rok is a whole number.
func (s *RobnoDokumentaResource) ValidateRobPro(ctx context.Context, params domain.RobnoProParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	vrd := common.StringToInt(params.Vrd)
	if _, found, err := s.robnoUnosDokvrsta(ctx, vrd); err != nil {
		add("vrd", common.ErrMsgReadData)
	} else if !found {
		add("vrd", robnoUnosMsgNePostoji)
	}

	// The datum dokumenta, not after the datum naloga.
	if datum, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && datum.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}

	// The broj dokumenta: no other nalog has the document of the vrsta and broj in the magacin.
	if broj := strings.TrimSpace(params.BrojDokum); broj == "" {
		add("brdok", common.ErrMsgObavezanPodatak)
	} else if dokum, err := strconv.Atoi(broj); err != nil {
		add("brdok", robnoUnosMsgCeoBroj)
	} else if exists, err := s.robno.RdokBrojDrugogNalogaPostoji(ctx, params.MagaciniID, vrd, dokum, params.RnalID, params.RdokID); err != nil {
		add("brdok", common.ErrMsgReadData)
	} else if exists {
		add("brdok", "Postoji u drugom nalogu.")
	}
	if rok := strings.TrimSpace(params.Rok); rok != "" {
		if _, err := strconv.Atoi(rok); err != nil {
			add("rok", robnoUnosMsgCeoBroj)
		}
	}

	// The kupac and its mesto isporuke.
	konto, sifra := strings.TrimSpace(params.KupacKonto), strings.TrimSpace(params.KupacSifra)
	kupacOK := false
	switch {
	case konto == "":
		add("kupackonto", common.ErrMsgObavezanPodatak)
	case sifra == "":
		add("kupacsifra", common.ErrMsgObavezanPodatak)
	default:
		if exists, err := s.robno.FkplAnalitikaPostoji(ctx, konto, sifra); err != nil {
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
		if exists, err := s.robno.FispPostoji(ctx, id, konto, sifra); err != nil {
			add("fispid", common.ErrMsgReadData)
		} else if !exists {
			add("fispid", "Nije mesto isporuke kupca.")
		}
	}

	// The org. jedinica (required), the komercijalista (when given).
	s.robnoUnosObavezanID(ctx, params.Orgjed, "idorgjed", s.robno.OrgjedPostoji, add)
	if id, given, ok := robnoUnosID(params.Komercijalista); !ok {
		add("komid", robnoUnosMsgNePostoji)
	} else if given {
		if exists, err := s.robno.KomercijalistaPostoji(ctx, id); err != nil {
			add("komid", common.ErrMsgReadData)
		} else if !exists {
			add("komid", robnoUnosMsgNePostoji)
		}
	}

	// The avansni računi (vrsta 172) of the profaktura must be of the kupac.
	if profak := strings.TrimSpace(params.Profaktura); profak != "" {
		if brojevi, ok := robnoUnosBrojevi(profak); !ok {
			add("profak", "Brojevi odvojeni zarezom.")
		} else if kupacOK {
			if exists, err := s.robno.RdokAvansniDrugogKupca(ctx, brojevi, konto, sifra); err != nil {
				add("profak", common.ErrMsgReadData)
			} else if exists {
				add("profak", "Avansni račun drugog kupca.")
			}
		}
	}

	// The šifra valute.
	if _, given, ok := robnoUnosID(params.IDValute); !ok || !given {
		add("idvalute", common.ErrMsgObavezanPodatak)
	}
	return fieldErrors
}

// SaveRobPro saves the header of a profaktura (the legacy "Sačuvaj" of rob_ProFakture, after ValidacijaRDOK:
// SnimiRDOK): the robni dokument with the fields of the header, with the old columns of the šifarnici (mi, kom,
// sifval and sifbank) by their keys; a new one when the nalog has no document of the vrsta and broj in the
// magacin (SaveRobniDokument). It returns the rdokid of the document.
func (s *RobnoDokumentaResource) SaveRobPro(ctx context.Context, params domain.RobnoProParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	fispID, komID, valutaID, bankaID := id(params.MestoIsporuke), id(params.Komercijalista), id(params.IDValute), id(params.BankaIzvoz)
	sif, err := s.robno.SifreSifarnika(ctx, fispID, komID, valutaID, bankaID, 0)
	if err != nil {
		return 0, err
	}
	return s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        common.StringToInt(params.Vrd),
		Dokum:      common.StringToInt(params.BrojDokum),
		Polja: []domain.Fields{
			robPoljeDatum("dadok", params.DatumDokum),
			robPoljeText("dokiz", params.IzvorniDokum),
			robPoljeDatum("datiz", params.IzvorniDokDatum),
			robPoljeText("opis", params.Opis),
			robPoljeInt("bi", params.BrojIsecka),
			robPoljeText("pornapomena", params.NapomenaPoresko),
			robPoljeInt("brojrata", params.BrojRata),
			robPoljeDatum("datumprrate", params.DatumPrveRate),
			robPoljeID("idvalute", valutaID), robPoljeInt("sifval", sif.C2),
			robPoljeBroj("kurs", params.Kurs),
			robPoljeInt("vaziza", params.VaziZa),
			robPoljeText("profak", params.Profaktura),
			robPoljeText("otp", params.NacinOtpreme),
			robPoljeText("pla", params.UsloviPlacanja),
			robPoljeInt("rok", params.Rok),
			robPoljeText("kamt", params.UsloviKamate),
			robPoljeText("vozac", params.Vozac),
			robPoljeText("brvozila", params.RegBr),
			robPoljeText("foot", params.Tekst),
			robPoljeText("fkto", params.KupacKonto),
			robPoljeText("fana", params.KupacSifra),
			robPoljeID("fispid", fispID), robPoljeInt("mi", sif.Key),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
			robPoljeID("komid", komID), robPoljeInt("kom", sif.C1),
			robPoljeText("paritet", params.Paritet),
			robPoljeInt("izjizv", params.IzvozIzjava),
			robPoljeID("bnkizvid", bankaID), robPoljeInt("sifbank", sif.C3),
		},
	})
}

// DeleteRobPro deletes a profaktura with its stavke (the legacy "Brisanje" of rob_ProFakture): a document
// printed on the fiskalni printer (knjige[4] "S") or posted (knjige[1] "D") is not deleted (poruka says why);
// else every stavka is taken back from the nalog and the stanje and deleted like the stavke of the porudžbenica
// (robPorBrisiStavkuTx: the same legacy procedure), the document is deleted and the nalog counts one document
// less (robBrisiDokumentSa).
func (s *RobnoDokumentaResource) DeleteRobPro(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokumentSa(ctx, rdokID, false, s.robPorBrisiStavkuTx)
}

// ---------------------------------------------------------------------------------------------------
// The stavke of a profaktura (the legacy rob_PROFAKTURE_DET)
// ---------------------------------------------------------------------------------------------------

// GetRobProArtikal returns the artikal of a stavka of a profaktura with what the form of the stavke shows of it,
// like the legacy exit of EDT_SIFRA of rob_PROFAKTURE_DET: the artikal by its šifra (or by its barkod for a
// value longer than 6 characters) with its naziv, JM, tarifa with its stopa on the datum of the document and
// dani; its stanje in the konto of the stavka (rsta: ulaz - izlaz) with its cena (rsta.cena) and nabavna cena
// (rsta.ncena, shown only); the last cena and rabat to the kupac (ZadnjaCena). found is false when the artikal
// does not exist.
func (s *RobnoDokumentaResource) GetRobProArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobStavkaArtikal, bool, error) {
	a := domain.RobStavkaArtikal{}
	if domain.GetSessionFromStdContext(ctx) == nil {
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
	a.Tarifa, a.Dani, a.Stopa = common.StringToInt(r.C3), common.StringToInt(r.C4), common.StringToFloat64(r.C7)
	// The konto of the stavka stays (the legacy keeps EDT_KONTO); without it the konto of the artikal.
	if a.Konto = strings.TrimSpace(konto); a.Konto == "" {
		a.Konto = strings.TrimSpace(r.C5)
	}
	st, stFound, err := s.robno.RstaStanjeStavke(ctx, k.MagaciniID, a.Konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if stFound {
		a.Stanje, a.Cena, a.NabavnaCena = st.Stanje, st.Cena, st.Ncena
	}
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

// ValidateRobProStavka validates a stavka of a profaktura before it is saved, like the legacy Validacija of
// rob_PROFAKTURE_DET: the konto is required and a konto of the magacin (magkonto), the šifra is required and
// the artikal exists (the šifra 0 too: its naziv and JM are typed), the količina is required and not 0, the cena
// is more than 0 and with the rabat the neto cena is not more than the cena.
func (s *RobnoDokumentaResource) ValidateRobProStavka(ctx context.Context, params domain.RobProStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("sifraartikla", common.ErrMsgSessionNotFound)
		return fieldErrors
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
	if konto := strings.TrimSpace(params.Konto); konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	} else if exists, err := s.robno.MagkontoPostoji(ctx, k.MagaciniID, konto); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if !exists {
		add("kontostavke", "Konto nije otvoren za magacin.")
	}
	if sifraText := strings.TrimSpace(params.Sifra); sifraText == "" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if sifra, err := strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.RsifPostoji(ctx, sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	}
	if strings.TrimSpace(params.Kolicina) == "" || common.StringToFloat64(params.Kolicina) == 0 {
		add("kolicina", common.ErrMsgObavezanPodatak)
	}
	cena := common.StringToFloat64(params.Cena)
	if cena <= 0 {
		add("prodajnacena", "Mora biti veća od 0.")
	}
	rabat := common.StringToFloat64(params.Rabat)
	if _, neto, _ := RobObracunStavke(0, cena, rabat); rabat != 0 && neto > cena {
		add("rabatstavke", "Neto cena veća od prodajne.")
	}
	return fieldErrors
}

// SaveRobProStavka saves a stavka of a profaktura, a new one (RproID 0, the legacy CreateRpro) or a changed one
// (EditRpro), in one transaction (robProSnimiStavku).
func (s *RobnoDokumentaResource) SaveRobProStavka(ctx context.Context, params domain.RobProStavkaParams) (rproID int64, err error) {
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
	if rproID, err = s.robProSnimiStavku(ctx, tx, k, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robProSnimiStavku saves a stavka of a profaktura in the transaction (see SaveRobProStavka), like the legacy
// CreateRpro and EditRpro of rob_PROFAKTURE_DET:
//   - the iznos is the količina by the cena (without the rabat), rounded to 2 decimals, negative for a storno;
//   - the stavka has the columns of its document (nalog, vrsta, broj, datum, magacin, kupac), the artikal
//     (rsifid; the naziv and the JM of the artikal, the typed ones for the šifra 0), the cena (cena and fcena),
//     the cena in the valuta of the document (cena / kurs / važi za, 3 decimals), the količina, the iznos, the
//     rabat, the tarifa and the dani; the nabavna cena is 0 (the legacy xNabCena is never set); a new stavka
//     gets the next redni broj (NadjiZadnjRBR);
//   - the nalog and the stanje of the artikal (AzurirajRNAL, AzurirajRSTA; the legacy does not update the zalihe
//     by lot) get the iznos and the količina: a changed stavka is taken back first;
//   - the document gets the iznos of the stavke and a new stavka counts in it (rdok.iznos, rdok.brst).
func (s *RobnoDokumentaResource) robProSnimiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, params domain.RobProStavkaParams) (int64, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.TrimSpace(params.Konto)
	cena := common.StringToFloat64(params.Cena)
	kolicina := math.Abs(common.StringToFloat64(params.Kolicina))
	iznos := robZaokruzi(math.Abs(kolicina*cena), 2)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}
	cenaVal := cena
	if k.Kurs > 0 && k.Vaziza > 0 {
		cenaVal = robZaokruzi(cena/k.Kurs/k.Vaziza, 3)
	}
	rsifID, naziv, jm, err := s.robno.RsifIDNazivJm(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if sifra == 0 {
		naziv, jm = strings.TrimSpace(params.Naziv), strings.TrimSpace(params.Jm)
	}

	// A changed stavka: its old values are taken back from the stanje.
	stara := domain.RobPromenaStanja{}
	if params.RproID > 0 {
		if stara.Sifra, stara.Konto, stara.Kolicina, stara.Iznos, err = s.robno.RproStaraKolicina(ctx, tx, params.RproID, k.RdokID); err != nil {
			return 0, err
		}
	}
	polja := []domain.Fields{
		robPoljeText("konto", konto),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		{Name: "naz1", Value: naziv},
		{Name: "jm", Value: jm},
		robPoljeFloat("cena", cena),
		robPoljeFloat("fcena", cena),
		robPoljeFloat("ncena", 0),
		robPoljeFloat("cenaval", cenaVal),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		robPoljeBroj("rab", params.Rabat),
		robPoljeInt("po", params.Tarifa),
		robPoljeInt("dani", params.Dani),
	}
	rproID := params.RproID
	if rproID > 0 {
		if err := s.unos.Rpro.UpdateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", rproID, polja); err != nil {
			return 0, err
		}
	} else {
		rbr, err := s.robno.RproSledeciRbr(ctx, tx, k.RdokID)
		if err != nil {
			return 0, err
		}
		polja = append(polja,
			domain.Fields{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
			domain.Fields{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
			domain.Fields{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
			domain.Fields{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
			domain.Fields{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
			robPoljeNullTime("dadok", k.Dadok),
			domain.Fields{Name: "fkto", Value: k.Fkto},
			domain.Fields{Name: "fana", Value: k.Fana},
			domain.Fields{Name: "mag", Value: strconv.Itoa(k.Mag)},
			domain.Fields{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""}, domain.Fields{Name: "taxcat", Value: ""}, domain.Fields{Name: "taxexreco", Value: ""},
		)
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja); err != nil {
			return 0, err
		}
	}

	// The nalog and the stanje: the old stavka taken back, the new one added.
	razlikaIznosa := iznos - stara.Iznos
	if err := s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), razlikaIznosa); err != nil {
		return 0, err
	}
	promena := func(sifra int, konto string, kolicina, iznos float64, brst int) domain.RobPromenaStanja {
		return domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra, Kolicina: kolicina, Iznos: iznos,
			Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Brst: brst}
	}
	if params.RproID > 0 {
		if err := s.AzurirajRSTA(ctx, tx, promena(stara.Sifra, stara.Konto, -stara.Kolicina, -stara.Iznos, -1)); err != nil {
			return 0, err
		}
	}
	if err := s.AzurirajRSTA(ctx, tx, promena(sifra, konto, kolicina, iznos, 1)); err != nil {
		return 0, err
	}

	// The document: its iznos, a new stavka counts in it.
	noviBrst := 0
	if params.RproID == 0 {
		noviBrst = 1
	}
	if err := s.robno.RdokDodajIznos(ctx, tx, k.RdokID, razlikaIznosa, noviBrst); err != nil {
		return 0, err
	}
	return rproID, nil
}

// DeleteRobProStavka deletes a stavka of a profaktura (the legacy BrisiRPRO) in one transaction and returns its
// document: the stavka is taken back from its nalog and from the stanje like a stavka of the porudžbenica
// (robPorBrisiStavkuTx: the same legacy procedure) and the stavke after it get the redni brojevi again (rbr
// only, from the redni broj of the deleted one).
func (s *RobnoDokumentaResource) DeleteRobProStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	rbr, err := s.robno.RproRbr(ctx, tx, rproID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("stavka ne postoji")
		}
		return 0, err
	}
	if rdokID, err = s.robPorBrisiStavkuTx(ctx, tx, rproID); err != nil {
		return 0, err
	}
	if err = s.robno.RproPrenumerisi(ctx, tx, rdokID, rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

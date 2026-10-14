package robno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Interni prenosi - dostavnice (the groups IRT and IRP, the legacy rob_Interni_racuni): the document is a
// robni dokument (rdok) like the faktura, with the prodavnica (rdok.pkto, rdok.pana)
// ---------------------------------------------------------------------------------------------------

// RobInterniVrste are the vrste dokumenata of the grid of the documents of the "Pregled internih prenosa" tab
// (the legacy ipVRD "120;121;128;129" of rob_Interni_racuni).
var RobInterniVrste = []int{120, 121, 128, 129}

// RobInterniStavkeTableRows returns the stavke of an interni prenos as the rows of the grid of the stavke of
// the screen (the columns of GetInterniStavkeTableFields: redni broj, konto, šifra, naziv, količina, the
// velepr. cena and rabat, the nabavna cena, the malopr. marža, the malopr. cena and the one with the porez),
// by the redni broj.
func RobInterniStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Fcena), robFormatIznos(st.Rab),
				robFormatIznos(st.Ncena), robFormatIznos(st.Mma), robFormatIznos(st.Mcena), robFormatIznos(st.Mcenap)},
		})
	}
	return rows
}

// RobInterniDokumentStavkeTableRows returns the stavke of the selected document of the "Pregled internih
// prenosa" tab (DETALJI, the columns of GetInterniDokumentStavkeTableFields: konto, šifra, naziv, JM,
// količina, the planska, the malopr. and the velepr. cena, the velepr. rabat, the malopr. marža and the
// malopr. cena with the porez), by the redni broj.
func RobInterniDokumentStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm, common.FormatNumberWithSystemLocale(st.Kolic, 3),
				robFormatIznos(st.Pcena), robFormatIznos(st.Mcena), robFormatIznos(st.Fcena), robFormatIznos(st.Rab),
				robFormatIznos(st.Mma), robFormatIznos(st.Mcenap)},
		})
	}
	return rows
}

// RobInterniDokumentaTableRows returns the documents of the "Pregled internih prenosa" tab as the rows of
// its grid (the key of a row is its rdokid; the columns of GetInterniDokumentaTableFields).
func RobInterniDokumentaTableRows(dokumenta []domain.RobDokumentPrikazRow) []domain.TableRow {
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
			Fields: []string{broj(d.Rbr), d.Tipdok, broj(d.Nalog), broj(d.Vrd), broj(d.Dokum), robFormatDatum(d.Dadok),
				robFormatIznos(d.Iznos), d.Fkto, d.Fana, d.Naziv, d.Opis},
		})
	}
	return rows
}

// GetRobInterniDokument returns the document of the "Interni prenosi - dostavnice" screen: the document of
// the params like the one of the faktura (GetRobFaktureDokument: an existing one of the nalog, else a new
// one with the next broj, NoviBrojRDOK) with its prodavnica, the konto and the šifra of the document (a new
// one: the konto of the prodavnice of the robno knjigovodstvo, rvr.fprod) with its naziv, adresa,
// poštanski broj and mesto. The prodajni objekat of the stavke is the prodavnica.
func (s *RobnoDokumentaResource) GetRobInterniDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoInterniDokument, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.RobnoInterniDokument{}, fmt.Errorf("no user session found")
	}
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoInterniDokument{}, err
	}
	res := domain.RobnoInterniDokument{
		Header: domain.RobnoInterniHeaderView{RobnoFaktureHeaderView: dok.Header},
		Stavka: domain.RobnoInterniStavka{RobnoFaktureStavka: dok.Stavka},
		Poruka: dok.Poruka,
	}
	h := &res.Header
	var pkto, pana string
	if h.RdokID > 0 {
		pkto, pana, err = s.robno.RdokPktoPana(ctx, h.RdokID)
	} else {
		pkto, err = s.robno.RvrFprod(ctx)
	}
	if err != nil {
		return res, err
	}
	h.ProdavnicaKonto, h.ProdavnicaSifra = strings.TrimSpace(pkto), strings.TrimSpace(pana)
	if h.ProdavnicaKonto != "" && h.ProdavnicaSifra != "" {
		prodavnica, found, err := s.GetRobKupac(ctx, h.ProdavnicaKonto, h.ProdavnicaSifra)
		if err != nil {
			return res, err
		}
		if found {
			h.ProdavnicaNaziv, h.ProdavnicaAdresa = prodavnica.Naziv, prodavnica.Adresa
			h.ProdavnicaPostanskiBroj, h.ProdavnicaMesto = prodavnica.Pobro, prodavnica.Mesto
		}
	}
	res.Stavka.ProdajniObjekat = strings.TrimSpace(strings.Join([]string{h.ProdavnicaKonto, h.ProdavnicaSifra, h.ProdavnicaNaziv}, " "))
	vrd := common.StringToInt(h.Vrd)
	if vrd == 0 {
		vrd = params.Vrd
	}
	if res.Stavka.Grupa, err = s.robInterniGrupa(ctx, vrd); err != nil {
		return res, err
	}
	return res, nil
}

// ValidateRobInterni validates the header of an interni prenos before it is saved, like the legacy
// ValidacijaRDOK of rob_Interni_racuni (the profaktura is not part of the screen):
//   - the datum dokumenta is required, in the business year and not after the datum naloga;
//   - the broj dokumenta is a whole number (the key of the document);
//   - the prodavnica (konto and šifra) is an analitika of the kontni plan (fkpl.vkonta 1);
//   - for the group IRT a given kupac is an analitika of the kontni plan;
//   - the org. jedinica is required.
func (s *RobnoDokumentaResource) ValidateRobInterni(ctx context.Context, params domain.RobnoInterniParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	god := userSession.SelectedGod

	// The vrsta dokumenta (its group: IRT, IRP).
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

	// The datum dokumenta: in the business year, not after the datum naloga.
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if dadok.Year() != god {
		add("datumfakturisanja", "Nije u poslovnoj godini.")
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}

	// The broj dokumenta: a whole number (the key of the document, rdok.dokum).
	if _, err := strconv.Atoi(strings.TrimSpace(params.BrojDokum)); strings.TrimSpace(params.BrojDokum) == "" {
		add("brdok", common.ErrMsgObavezanPodatak)
	} else if err != nil {
		add("brdok", robnoUnosMsgCeoBroj)
	}

	// The prodavnica: an analitika of the kontni plan.
	pkto, pana := strings.TrimSpace(params.ProdavnicaKonto), strings.TrimSpace(params.ProdavnicaSifra)
	switch {
	case pkto == "":
		add("pkto", common.ErrMsgObavezanPodatak)
	case pana == "":
		add("pana", common.ErrMsgObavezanPodatak)
	default:
		if exists, err := s.robno.FkplAnalitickiKontoPostoji(ctx, pkto, pana); err != nil {
			add("pana", common.ErrMsgReadData)
		} else if !exists {
			add("pana", robnoUnosMsgNePostoji)
		}
	}

	// The kupac of an IRT: when given, an analitika of the kontni plan.
	if strings.EqualFold(strings.TrimSpace(dokvrsta.GrpDok), "IRT") && strings.TrimSpace(params.KupacKonto) != "" {
		if exists, err := s.robno.FkplAnalitickiKontoPostoji(ctx, strings.TrimSpace(params.KupacKonto), strings.TrimSpace(params.KupacSifra)); err != nil {
			add("kupacsifra", common.ErrMsgReadData)
		} else if !exists {
			add("kupacsifra", robnoUnosMsgNePostoji)
		}
	}

	// The org. jedinica: required.
	if id, given, ok := robnoUnosID(params.Orgjed); !given {
		add("idorgjed", common.ErrMsgObavezanPodatak)
	} else if !ok {
		add("idorgjed", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.OrgjedPostoji(ctx, id); err != nil {
		add("idorgjed", common.ErrMsgReadData)
	} else if !exists {
		add("idorgjed", robnoUnosMsgNePostoji)
	}
	return fieldErrors
}

// SaveRobInterni saves the header of an interni prenos (the legacy "Sačuvaj" of rob_Interni_racuni, after
// ValidacijaRDOK: SnimiRDOK): the robni dokument with the fields of the header, a new one when the nalog has
// no document of the vrsta and broj in the magacin (SaveRobniDokument; its stavke get the datum dokumenta and
// the prodavnica). It returns the rdokid of the document.
func (s *RobnoDokumentaResource) SaveRobInterni(ctx context.Context, params domain.RobnoInterniParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
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
			robPoljeInt("rok", params.Rok),
			robPoljeText("fkto", params.KupacKonto),
			robPoljeText("fana", params.KupacSifra),
			robPoljeText("pkto", params.ProdavnicaKonto),
			robPoljeText("pana", params.ProdavnicaSifra),
			robPoljeText("pornapomena", params.NapomenaPoresko),
			robPoljeText("otp", params.NacinOtpreme),
			robPoljeText("pla", params.UsloviPlacanja),
			robPoljeText("kamt", params.UsloviKamate),
			robPoljeText("foot", params.Tekst),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
		},
	})
}

// DeleteRobFaktura deletes a faktura with its stavke (the legacy "Briši" of rob_Fakture): a document sent to
// the SEF (salesinvoiceid, the status not "Cancelled" or "Otkazana"), sent to the Pojedinačna evidencija PDV
// (individualvatid), printed on the fiskalni printer (knjige[4] "S") or posted (knjige[1] "D") is not deleted
// (poruka says why); else the stavke and the document are deleted (robBrisiDokument) with the avansi and the
// rate of the document (avansfakt, faktrate).
func (s *RobnoDokumentaResource) DeleteRobFaktura(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokument(ctx, rdokID, true)
}

// DeleteRobFakturaUsluge deletes a faktura usluga with its stavke (the legacy "Briši" of rob_Fakture_Usluga):
// like DeleteRobFaktura, with the dodatni opis usluge of the document (dokusldet).
func (s *RobnoDokumentaResource) DeleteRobFakturaUsluge(ctx context.Context, rdokID int64) (string, error) {
	// An avansni račun (the legacy "Briši" of rob_AvansniRacuni).
	if vrd, found, err := s.robno.RdokVrd(ctx, rdokID); err == nil && found {
		if grupa, _ := s.robInterniGrupa(ctx, vrd); robAvansni(grupa) {
			return s.DeleteRobAvansni(ctx, rdokID)
		}
	}
	return s.robBrisiDokument(ctx, rdokID, true, s.robno.DokusldetBrisi)
}

// DeleteRobInterni deletes an interni prenos with its stavke (the legacy "Briši" of rob_Interni_racuni): a
// document printed on the fiskalni printer (knjige[4] "S") or posted (knjige[1] "D") is not deleted (poruka
// says why); else the stavke and the document are deleted (robBrisiDokument).
func (s *RobnoDokumentaResource) DeleteRobInterni(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokument(ctx, rdokID, false)
}

// robBrisiDokument deletes a robni dokument with its stavke in one transaction, like the "Briši" of the legacy
// windows: the checks of the window (faktura: also the SEF and the Pojedinačna evidencija PDV) end with poruka
// (the document is not deleted); else every stavka is taken back from its nalog and from the stanje of the
// artikal (DeleteRobStavkaTx: the legacy AzurirajRNAL and AzurirajRSTA; the SR is not updated) and deleted,
// for a faktura its avansi and rate are deleted (before the document: they refer to it), the document is
// deleted and the nalog counts one document less (rnal.brdo; its stavke, rnal.brst, are counted again by the
// deletion of the stavke). The functions of dodatno delete the other data of the document before it.
func (s *RobnoDokumentaResource) robBrisiDokument(ctx context.Context, rdokID int64, faktura bool, dodatno ...robBrisanjeDokumenta) (poruka string, err error) {
	return s.robBrisiDokumentSa(ctx, rdokID, faktura, s.DeleteRobStavkaTx, dodatno...)
}

// robBrisanjeDokumenta deletes, in the transaction tx, data that refers to the document rdokID.
type robBrisanjeDokumenta func(ctx context.Context, tx db.Transaction, rdokID int64) error

// robBrisiDokumentSa is robBrisiDokument with the deletion of a stavka of the window (brisiStavku, in the
// transaction: e.g. the porudžbenica takes its stavke back from the stanje only, without the zalihe by lot).
func (s *RobnoDokumentaResource) robBrisiDokumentSa(ctx context.Context, rdokID int64, faktura bool,
	brisiStavku func(ctx context.Context, tx db.Transaction, rproID int64) (int64, error), dodatno ...robBrisanjeDokumenta) (poruka string, err error) {
	tx, err := s.unos.Dokument.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil || poruka != "" {
			_ = tx.Rollback()
		}
	}()
	d, err := s.robno.RdokZaBrisanje(ctx, tx, rdokID)
	if errors.Is(err, sql.ErrNoRows) {
		return "Niste selektovali dokument za brisanje.", nil
	}
	if err != nil {
		return "", err
	}
	statusSef := strings.TrimSpace(d.StatusSef)
	switch {
	case faktura && d.SefID > 0 && statusSef != "Cancelled" && statusSef != "Otkazana":
		return "Dokument je poslat na SEF, brisanje nije dozvoljeno.", nil
	case faktura && d.PepdvID > 0:
		return "Dokument je poslat u Pojedinačnu evidenciju PDV, brisanje nije dozvoljeno.", nil
	case strings.TrimSpace(d.Fiskalizovan) == "S":
		return "Ovaj dokument je odštampan na fiskalnom printeru, brisanje nije dozvoljeno.", nil
	case strings.TrimSpace(d.Knjizen) == "D":
		return "Ovaj dokument je proknjižen, brisanje nije dozvoljeno.", nil
	}

	// The stavke, every one taken back from the nalog and the stanje.
	stavke, err := s.robno.RproIDsDokumenta(ctx, rdokID)
	if err != nil {
		return "", err
	}
	for _, rproID := range stavke {
		if _, err = brisiStavku(ctx, tx, rproID); err != nil {
			return "", err
		}
	}
	brisanja := dodatno
	if faktura {
		brisanja = append(brisanja, s.robno.AvansfaktFaktureBrisi, s.robno.FaktrateBrisi)
	}
	for _, brisi := range brisanja {
		if err = brisi(ctx, tx, rdokID); err != nil {
			return "", err
		}
	}
	if err = s.unos.Dokument.DeleteWithTx(ctx, tx, "rdokid", rdokID); err != nil {
		return "", err
	}
	if d.RnalID > 0 {
		if err = s.robno.RnalOduzmiDokument(ctx, tx, d.RnalID); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return "", nil
}

// ---------------------------------------------------------------------------------------------------
// The stavke of an interni prenos (the legacy rob_InterniRacuni_DET)
// ---------------------------------------------------------------------------------------------------

// robInterniGrupa returns the group of the vrsta of the document (dokvrsta.grpdok: IRT, IRP, IRS), upper case.
func (s *RobnoDokumentaResource) robInterniGrupa(ctx context.Context, vrd int) (string, error) {
	dokvrsta, _, err := s.robnoUnosDokvrsta(ctx, vrd)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(strings.TrimSpace(dokvrsta.GrpDok)), nil
}

// robInterniRsta is the stanje and the cene of an artikal in the konto of the magacin (rsta).
type robInterniRsta struct {
	stanje, cena, prosnc, vma, mma, mcena, mcenap float64
}

// robInterniStanje reads the stanje and the cene of the artikal in the konto of the magacin (rsta); found is
// false without it.
func (s *RobnoDokumentaResource) robInterniStanje(ctx context.Context, magaciniID int, konto string, sifra int) (robInterniRsta, bool, error) {
	c, found, err := s.robno.RstaStanjeCene(ctx, magaciniID, konto, sifra)
	if err != nil || !found {
		return robInterniRsta{}, false, err
	}
	return robInterniRsta{stanje: c.Stanje, cena: c.Cena, prosnc: c.Prosnc, vma: c.Vma, mma: c.Mma, mcena: c.Mcena, mcenap: c.Mcenap}, true, nil
}

// robZaokruzi rounds the value to the decimals.
func robZaokruzi(value float64, decimale int) float64 {
	p := math.Pow(10, float64(decimale))
	return math.Round(value*p) / p
}

// GetRobInterniArtikal returns the artikal of a stavka of an interni prenos with what the form of the stavke
// shows of it, like the legacy exit of EDT_SIFRA and EDT_KONTO of rob_InterniRacuni_DET:
//   - the artikal by its šifra, or by its barkod for a value longer than 6 characters; its naziv, JM, tarifa
//     with its stopa on the datum of the document (VratiPorez2), model and taksa (rcene.itaksa);
//   - its stanje in the konto of the magacin (rsta: ulaz - izlaz); for an IRT its cene: in a magacin of the
//     nabavna cena (nacvodzal 3) the cena and the nabavna cena are rsta.prosnc and the rabat 0, else the cena
//     is rsta.cena, the rabat vma * 100 / (vma + 100) and the nabavna cena the cena without the rabat; the
//     maloprodajna cena and the one with the porez of rsta, the marža from the maloprodajna and the nabavna
//     cena; the maloprodajna cena with the porez of rsta for the model "D";
//   - the last cena of the prodavnica (ZadnjaCena).
//
// found is false when the artikal does not exist.
func (s *RobnoDokumentaResource) GetRobInterniArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobInterniArtikal, bool, error) {
	a := domain.RobInterniArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	grupa, err := s.robInterniGrupa(ctx, k.Vrd)
	if err != nil {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	rows, err := s.robno.RsifArtikalSaStopom(ctx, len(vrednost) > 6, vrednost, robNullTime(k.Dadok))
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm = common.StringToInt(r.Key), r.C1, r.C2
	a.Tarifa, a.Model = common.StringToInt(r.C3), strings.TrimSpace(r.C4)
	a.Stopa, a.Taksa = common.StringToFloat64(r.C5), common.StringToFloat64(r.C6)

	st, stFound, err := s.robInterniStanje(ctx, k.MagaciniID, konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	a.Grupa, a.Rsta = grupa, stFound
	if stFound {
		a.Stanje = st.stanje
		a.RstaCena, a.RstaMma, a.RstaMcenap = st.cena, st.mma, st.mcenap
		if st.vma+100 != 0 {
			a.RstaVra = st.vma * 100 / (st.vma + 100)
		}
		if grupa == "IRT" {
			a.Cene = true
			if k.Nacvodzal == 3 {
				a.Cena, a.Vra, a.Ncena = st.prosnc, 0, st.prosnc
			} else {
				a.Cena, a.Vra = st.cena, a.RstaVra
				a.Ncena = robZaokruzi(a.Cena-robZaokruzi(a.Cena*a.Vra/100, 2), 2)
			}
			a.Mcena = st.mcena
			if a.Ncena > 0 {
				a.Mma = (st.mcena - a.Ncena) / a.Ncena * 100
			}
			a.Mcenap = st.mcenap
		}
		if strings.EqualFold(a.Model, "D") {
			a.McenapModel, a.Mcenap = true, st.mcenap
		}
	}

	// ZadnjaCena: the last stavka of the artikal to the prodavnica in a document of the vrsta up to the datum
	// of the document; else the maloprodajne cene of the stanje; else the message.
	zadnja, err := s.robno.RproZadnjaCenaProdavnice(ctx, rdokID, a.Sifra)
	if err != nil {
		return a, true, err
	}
	switch {
	case zadnja != nil && len(*zadnja) > 0:
		z := (*zadnja)[0]
		a.ZadnjaMcena, a.ZadnjiVra = common.StringToFloat64(z.Key), common.StringToFloat64(z.C1)
		a.ZadnjaMcenap, a.ZadnjaMma, a.ZadnjiDatum = common.StringToFloat64(z.C2), common.StringToFloat64(z.C3), z.C4
	case stFound:
		a.ZadnjaMcena, a.ZadnjaMcenap = st.mcena, st.mcenap
	default:
		a.Poruka = "Prodajni objekat nije uzimao ovaj artikal..."
	}
	return a, true, nil
}

// ValidateRobInterniStavka validates a stavka of an interni prenos before it is saved, like the legacy
// Validacija of rob_InterniRacuni_DET: the konto and the šifra are required (the artikal exists), the količina
// is greater than 0, the cena, the maloprodajna cena and the one with the porez are greater than 0, and for an
// IRT (not a storno) the stanje has the količina (the legacy question "Nema dovoljno zaliha. Da li ste sigurni
// da izdajete robu?" is the error of the field, like the faktura); the stanje of a changed stavka has its own
// količina. The question of the OTK broj of the magacini of the tip zaliha 2 ("Unos bez OTK broja?") is not
// asked: the stavka is saved without it.
func (s *RobnoDokumentaResource) ValidateRobInterniStavka(ctx context.Context, params domain.RobInterniStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("sifraartikla", common.ErrMsgReadData)
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
	grupa, err := s.robInterniGrupa(ctx, k.Vrd)
	if err != nil {
		add("sifraartikla", common.ErrMsgReadData)
		return fieldErrors
	}
	konto := strings.TrimSpace(params.Konto)
	if konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	}
	sifra := 0
	if sifraText := strings.TrimSpace(params.Sifra); sifraText == "" || sifraText == "0" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if sifra, err = strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.RsifPostoji(ctx, sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	}
	kolicina := common.StringToFloat64(params.Kolic)
	switch {
	case strings.TrimSpace(params.Kolic) == "":
		add("kolicina", common.ErrMsgObavezanPodatak)
	case kolicina <= 0:
		add("kolicina", robInterniMsgVeceOdNule)
	}
	if common.StringToFloat64(params.Cena) <= 0 {
		add("magacinskacena", robInterniMsgVeceOdNule)
	}
	if common.StringToFloat64(params.Mcena) <= 0 {
		add("maloprcena", robInterniMsgVeceOdNule)
	}
	if common.StringToFloat64(params.Mcenap) <= 0 {
		add("mpcenasaporezom", robInterniMsgVeceOdNule)
	}
	// The zalihe of an IRT.
	if grupa == "IRT" && !k.Storno() && kolicina > 0 && sifra > 0 && konto != "" {
		st, _, err := s.robInterniStanje(ctx, k.MagaciniID, konto, sifra)
		if err != nil {
			add("kolicina", common.ErrMsgReadData)
			return fieldErrors
		}
		stanje := st.stanje
		if params.RproID > 0 {
			if stara, found, err := s.robno.RproKolicina(ctx, params.RproID, sifra, konto); err == nil && found {
				stanje += math.Abs(stara)
			}
		}
		if stanje <= 0 || robZaokruzi(stanje-kolicina, 3) < 0 {
			add("kolicina", "Nema dovoljno zaliha.")
		}
	}
	return fieldErrors
}

// robInterniMsgVeceOdNule is the error of a količina or a cena that is not greater than 0.
const robInterniMsgVeceOdNule = "Mora biti veće od 0."

// SaveRobInterniStavka saves a stavka of an interni prenos, a new one (RproID 0, the legacy CreateRPRO) or a
// changed one (UpdateRPRO), in one transaction:
//   - the iznos is the količina by the cena (an IRS: by the nabavna cena of the form), negative for a storno,
//     rounded to 2 decimals;
//   - the stavka has the columns of its document (nalog, vrsta, broj, datum, magacin, kupac, prodavnica), the
//     artikal (naziv, JM, rsifid), the cena, the rabat (vra), the nabavna cena (ncena; an IRS: pcena), the marža
//     (mma), the maloprodajne cene (mcena, mcenap), the tarifa, the OTK broj, the serija and the rok; rab, vma
//     and mra are 0; a new stavka gets the next redni broj (NadjiZadnjRBR);
//   - a new artikal of the cene (rcene) is added with the cene of the stavka;
//   - the nalog, the stanje of the artikal and the zalihe by lot (AzurirajRNAL, AzurirajRSTA, AzurirajDRSTA)
//     get the iznos and the količina: a changed stavka is taken back first;
//   - the document gets the iznos of the stavka and a new stavka counts in it (rdok.iznos, rdok.brst).
func (s *RobnoDokumentaResource) SaveRobInterniStavka(ctx context.Context, params domain.RobInterniStavkaParams) (rproID int64, err error) {
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
	grupa, err := s.robInterniGrupa(ctx, k.Vrd)
	if err != nil {
		return 0, err
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

	sifra := common.StringToInt(params.Sifra)
	konto := strings.TrimSpace(params.Konto)
	kolicina := math.Abs(common.StringToFloat64(params.Kolic))
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	vra, mma := common.StringToFloat64(params.Vra), common.StringToFloat64(params.Mma)
	mcena, mcenap := common.StringToFloat64(params.Mcena), common.StringToFloat64(params.Mcenap)
	iznos := kolicina * cena
	if grupa == "IRS" {
		iznos = ncena * kolicina
	}
	iznos = robZaokruzi(math.Abs(iznos), 2)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}

	// The artikal: its rsifid, naziv and JM for the stavka.
	rsifID, naziv, jm, err := s.robno.RsifIDNazivJm(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	// The prodavnica of the document.
	pkto, pana, err := s.robno.RdokPktoPanaTx(ctx, tx, params.RdokID)
	if err != nil {
		return 0, err
	}

	// A changed stavka: its old values are taken back from the stanje.
	stara := domain.RobPromenaStanja{}
	if params.RproID > 0 {
		if stara, err = s.robno.RproStaraPromena(ctx, tx, params.RproID, params.RdokID); err != nil {
			return 0, err
		}
	}

	rok, _ := robnoUnosDatum(params.Rok)
	nabavnaKolona := "ncena"
	if grupa == "IRS" {
		nabavnaKolona = "pcena"
	}
	polja := []domain.Fields{
		robPoljeText("konto", konto),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		{Name: "naz1", Value: naziv},
		{Name: "jm", Value: jm},
		robPoljeFloat("cena", robZaokruzi(cena, 4)),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "rab", Value: "0"},
		robPoljeInt("po", params.Tarifa),
		{Name: "vma", Value: "0"},
		robPoljeFloat("mma", robZaokruzi(mma, 2)),
		robPoljeFloat("mcenap", robZaokruzi(mcenap, 4)),
		robPoljeFloat("vra", robZaokruzi(vra, 2)),
		{Name: "mra", Value: "0"},
		robPoljeFloat("mcena", robZaokruzi(mcena, 4)),
		robPoljeFloat(nabavnaKolona, robZaokruzi(ncena, 4)),
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)},
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
			domain.Fields{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
			domain.Fields{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
			domain.Fields{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
			domain.Fields{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
			domain.Fields{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
			domain.Fields{Name: "mag", Value: strconv.Itoa(k.Mag)},
			domain.Fields{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
			robPoljeNullTime("dadok", k.Dadok),
			domain.Fields{Name: "fkto", Value: k.Fkto},
			domain.Fields{Name: "fana", Value: k.Fana},
			robPoljeText("pkto", pkto),
			robPoljeText("pana", pana),
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""}, domain.Fields{Name: "taxcat", Value: ""}, domain.Fields{Name: "taxexreco", Value: ""},
		)
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja); err != nil {
			return 0, err
		}
	}

	// A new artikal of the cene (CreateRPRO): the cene of the stavka.
	if params.RproID == 0 {
		var cene bool
		if cene, err = s.robno.RcenePostoji(ctx, tx, sifra); err != nil {
			return 0, err
		}
		if !cene {
			if _, err = s.unos.Rcene.CreateWithTx(ctx, tx, &domain.RobRceneDto{}, "rceneid", []domain.Fields{
				{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
				{Name: "sifra", Value: strconv.Itoa(sifra)},
				robPoljeFloat("mma", mma), robPoljeFloat("vra", vra), {Name: "mra", Value: "0"}, {Name: "ztro", Value: "0"},
				robPoljeFloat("cena", cena), {Name: "fcena", Value: "0"},
				robPoljeFloat("mcena", mcena), robPoljeFloat("mcenap", mcenap),
				robPoljeFloat("ncena", ncena), {Name: "ztrin", Value: "0"},
				robPoljeFloat("rab", mma), {Name: "tz", Value: ""}, {Name: "tzi", Value: ""},
			}); err != nil {
				return 0, err
			}
		}
	}

	// The nalog, the stanje of the artikal and the zalihe by lot: the old stavka taken back, the new one added.
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
		robPoljeFloat("iznos", robZaokruzi(dokIznos+razlikaIznosa, 2)), {Name: "brst", Value: strconv.Itoa(dokBrst)},
	}); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// DeleteRobInterniStavka deletes a stavka of an interni prenos (the legacy BrisiRPRO) in one transaction and
// returns its document: the stavka is taken back from its nalog, from the stanje and from the zalihe by lot,
// the document loses its iznos and counts one stavka less (DeleteRobStavkaTx), and the stavke after it get
// the redni brojevi again (rbr and brst, from the redni broj of the deleted one).
func (s *RobnoDokumentaResource) DeleteRobInterniStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
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
	if rdokID, err = s.DeleteRobStavkaTx(ctx, tx, rproID); err != nil {
		return 0, err
	}
	if err = s.robno.RproPrenumerisiSaBrst(ctx, tx, rdokID, rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// GetRobInterniStavkaForma returns a saved stavka of an interni prenos for the form of the stavke (the legacy
// "Izmene" of rob_InterniRacuni_DET): its konto, šifra, količina, cene, rabat, marža, OTK broj, serija and rok,
// the artikal with its stopa on the datum of the document, model and taksa, and the stanje with the količina
// of the stavka (the legacy "Izmene" takes the stavka back from the stanje). found is false without it.
func (s *RobnoDokumentaResource) GetRobInterniStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobInterniStavkaForma, bool, error) {
	f := domain.RobInterniStavkaForma{}
	rows, err := s.robno.RproInterniForma(ctx, rproID, rdokID)
	if err != nil || rows == nil || len(*rows) == 0 {
		return f, false, err
	}
	r := (*rows)[0]
	num := common.StringToFloat64
	f.RproID, f.Sifra, f.Konto = rproID, common.StringToInt(r.Key), strings.TrimSpace(r.C1)
	f.Kolic, f.Cena, f.Vra, f.Ncena, f.Mma = math.Abs(num(r.C2)), num(r.C3), num(r.C4), num(r.C5), num(r.C6)
	if cene := strings.Split(r.C7, "|"); len(cene) == 3 {
		f.Mcena, f.Mcenap = num(cene[0]), num(cene[1])
		if roktr, _ := strconv.ParseInt(cene[2], 10, 64); roktr > 0 {
			f.Rok = robnoStampaWinDevDatumFormat(roktr, common.HtmlLayout)
		}
	}
	if lot := strings.SplitN(r.C8, "|", 2); len(lot) == 2 {
		f.Otk, f.Serija = strings.TrimSpace(lot[0]), strings.TrimSpace(lot[1])
	}
	artikal, aFound, err := s.GetRobInterniArtikal(ctx, rdokID, strconv.Itoa(f.Sifra), f.Konto)
	if err != nil {
		return f, true, err
	}
	if aFound {
		f.Naziv, f.Jm, f.Tarifa, f.Stopa, f.Model, f.Taksa = artikal.Naziv, artikal.Jm, artikal.Tarifa, artikal.Stopa, artikal.Model, artikal.Taksa
		f.Stanje = artikal.Stanje + f.Kolic
	}
	return f, true, nil
}

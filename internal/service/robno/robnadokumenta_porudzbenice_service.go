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
// Porudžbenice (the group POR, the legacy rob_Porudzbenica with its stavke rob_PORUDZBENICE_DET): a robni
// dokument (rdok) of a dobavljač (rdok.fkto, rdok.fana) whose stavke are at the nabavna cena (rpro.ncena)
// ---------------------------------------------------------------------------------------------------

// robPorUsloviPlacanja are the uslovi plaćanja of a new porudžbenica (the content of the legacy EDT_PLA).
const robPorUsloviPlacanja = "VIRMANOM"

// RobPorStavkeTableRows returns the stavke of a porudžbenica as the rows of the grid of the stavke (the
// columns of GetPorStavkeTableFields, the legacy QRY_RRPO_DET: redni broj, konto, šifra, naziv, JM,
// količina, nabavna cena, iznos), by the redni broj. The grid of the stavke of the selected document of the
// tab "Pregled porudžbenica" has the same columns (the legacy QRY_RPRO_ZADOK).
func RobPorStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Ncena), robFormatIznos(st.Iznos)},
		})
	}
	return rows
}

// GetRobPorDokument returns the document of the "Porudžbenice" screen: the document of the params like the one
// of the faktura (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one with
// the next broj, NoviBrojRDOK; the dobavljač is the kupac of the faktura, a new one gets the konto of rvr.fkup)
// with the opis of its vrsta. A new document is of today (the legacy EDT_DADOK = Today()) with the uslovi
// plaćanja "VIRMANOM". The stavke repeat the dobavljač (the legacy EDT_KUPAC).
func (s *RobnoDokumentaResource) GetRobPorDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoPorDokument, error) {
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoPorDokument{}, err
	}
	res := domain.RobnoPorDokument{
		Header: domain.RobnoPorHeaderView{RobnoFaktureHeaderView: dok.Header},
		Stavka: domain.RobnoPorStavka{RobnoFaktureStavka: dok.Stavka, MagaciniID: dok.Header.MagaciniID},
		Poruka: dok.Poruka,
	}
	h := &res.Header
	if dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(h.Vrd)); err != nil {
		return res, err
	} else if found {
		h.VrstaOpis = dokvrsta.Opis
	}
	if h.RdokID == 0 {
		h.DatumDokum = time.Now()
		h.Dospeva = h.DatumDokum
		h.UsloviPlacanja = robPorUsloviPlacanja
		res.Stavka.DatumDokumenta = h.DatumDokum.Format(common.DateLayout)
	}
	if strings.TrimSpace(h.KupacNaziv) != "" {
		res.Stavka.Dobavljac = strings.TrimSpace(fmt.Sprintf("%s %s %s ;%s ;%s %s", h.KupacKonto, h.KupacSifra, h.KupacNaziv,
			h.KupacAdresa, h.KupacPostanskiBroj, h.KupacMesto))
	}
	return res, nil
}

// ValidateRobPor validates the header of a porudžbenica before it is saved, like the legacy ValidacijaRDOK
// of rob_Porudzbenica: the datum dokumenta is a valid date, the konto and the šifra of the dobavljač are
// required and the dobavljač exists (fkpl), a given org. jedinica and komercijalista exist;
// the broj dokumenta is a whole number (the key of the document) and the rok a whole number.
func (s *RobnoDokumentaResource) ValidateRobPor(ctx context.Context, params domain.RobnoPorParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	if _, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(params.Vrd)); err != nil {
		add("vrd", common.ErrMsgReadData)
	} else if !found {
		add("vrd", robnoUnosMsgNePostoji)
	}
	if _, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	}
	if broj := strings.TrimSpace(params.BrojDokum); broj == "" {
		add("brdok", common.ErrMsgObavezanPodatak)
	} else if _, err := strconv.Atoi(broj); err != nil {
		add("brdok", robnoUnosMsgCeoBroj)
	}
	if rok := strings.TrimSpace(params.Rok); rok != "" {
		if _, err := strconv.Atoi(rok); err != nil {
			add("rok", robnoUnosMsgCeoBroj)
		}
	}

	// The dobavljač: the konto and the šifra of the kontni plan.
	konto, sifra := strings.TrimSpace(params.DobavljacKonto), strings.TrimSpace(params.DobavljacSifra)
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
		}
	}

	// The org. jedinica and the komercijalista (each when given).
	s.robnoUnosPostojiID(ctx, params.Orgjed, "idorgjed", s.robno.OrgjedPostoji, add)
	if id, given, ok := robnoUnosID(params.Komercijalista); !ok {
		add("komid", robnoUnosMsgNePostoji)
	} else if given {
		if exists, err := s.robno.KomercijalistaPostoji(ctx, id); err != nil {
			add("komid", common.ErrMsgReadData)
		} else if !exists {
			add("komid", robnoUnosMsgNePostoji)
		}
	}
	return fieldErrors
}

// SaveRobPor saves the header of a porudžbenica (the legacy "Sačuvaj" of rob_Porudzbenica, after
// ValidacijaRDOK: SnimiRDOK): the robni dokument with the fields of the header (a new one gets the opis
// "Porudzbenica br: N"), a new one when the nalog has no document of the vrsta and broj in the magacin
// (SaveRobniDokument). It returns the rdokid of the document.
func (s *RobnoDokumentaResource) SaveRobPor(ctx context.Context, params domain.RobnoPorParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	komID := id(params.Komercijalista)
	kom := ""
	if komID > 0 {
		var err error
		if kom, err = s.robno.KomercijalistaSifra(ctx, komID); err != nil {
			return 0, err
		}
	}
	dokum := common.StringToInt(params.BrojDokum)
	return s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        common.StringToInt(params.Vrd),
		Dokum:      dokum,
		Polja: []domain.Fields{
			robPoljeDatum("dadok", params.DatumDokum),
			robPoljeInt("rok", params.Rok),
			robPoljeText("fkto", params.DobavljacKonto),
			robPoljeText("fana", params.DobavljacSifra),
			robPoljeText("otp", params.NacinDopremanja),
			robPoljeText("pla", params.UsloviPlacanja),
			robPoljeText("foot", params.Tekst),
			robPoljeID("komid", komID), robPoljeInt("kom", kom),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
		},
		NoviPolja: []domain.Fields{robPoljeText("opis", fmt.Sprintf("Porudzbenica br: %d", dokum))},
	})
}

// DeleteRobPor deletes a porudžbenica with its stavke (the legacy "Brisanje" of rob_Porudzbenica): a document
// printed on the fiskalni printer (knjige[4] "S") or posted (knjige[1] "D") is not deleted (poruka says why);
// else every stavka is taken back from the nalog and the stanje (robPorBrisiStavkuTx) and deleted, the
// document is deleted and the nalog counts one document less (robBrisiDokumentSa).
func (s *RobnoDokumentaResource) DeleteRobPor(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokumentSa(ctx, rdokID, false, s.robPorBrisiStavkuTx)
}

// ---------------------------------------------------------------------------------------------------
// The stavke of a porudžbenica (the legacy rob_PORUDZBENICE_DET)
// ---------------------------------------------------------------------------------------------------

// robPorStanje reads the stanje (ulaz - izlaz) and the nabavna cena of the artikal in the konto of the
// magacin (rsta); found is false without it.
func (s *RobnoDokumentaResource) robPorStanje(ctx context.Context, magaciniID int, konto string, sifra int) (stanje, ncena float64, found bool, err error) {
	return s.robno.RstaStanjeNcena(ctx, magaciniID, konto, sifra)
}

// GetRobPorArtikal returns the artikal of a stavka of a porudžbenica with what the form of the stavke shows of
// it, like the legacy exit of EDT_SIFRA and EDT_KONTO of rob_PORUDZBENICE_DET: the artikal by its šifra (or by
// its barkod for a value longer than 6 characters) with its naziv, JM and tarifa with its stopa on the datum
// of the document (VratiPorez2); its stanje in the konto of the magacin (rsta: ulaz - izlaz) and its nabavna
// cena, the one of the stanje (rsta.ncena), without the stanje the one of the cene (rcene.ncena). found is
// false when the artikal does not exist.
func (s *RobnoDokumentaResource) GetRobPorArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobPorArtikal, bool, error) {
	a := domain.RobPorArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	rows, err := s.robno.RsifArtikalSaNcenom(ctx, len(vrednost) > 6, vrednost, robNullTime(k.Dadok))
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm = common.StringToInt(r.Key), r.C1, r.C2
	a.Tarifa, a.Stopa, a.Ncena = common.StringToInt(r.C3), common.StringToFloat64(r.C4), common.StringToFloat64(r.C5)
	stanje, ncena, stFound, err := s.robPorStanje(ctx, k.MagaciniID, konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if a.Rsta = stFound; stFound {
		a.Stanje, a.Ncena = stanje, ncena
	}
	return a, true, nil
}

// ValidateRobPorStavka validates a stavka of a porudžbenica before it is saved, like the legacy Validacija of
// rob_PORUDZBENICE_DET: the konto is required and a konto of the magacin (magkonto), the šifra is required and
// the artikal exists, the količina is required and not 0.
func (s *RobnoDokumentaResource) ValidateRobPorStavka(ctx context.Context, params domain.RobPorStavkaParams) []domain.FieldError {
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
	if konto := strings.TrimSpace(params.Konto); konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	} else if exists, err := s.robno.MagkontoPostoji(ctx, k.MagaciniID, konto); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if !exists {
		add("kontostavke", "Konto nije otvoren za magacin.")
	}
	if sifraText := strings.TrimSpace(params.Sifra); sifraText == "" || sifraText == "0" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if sifra, err := strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.RsifPostoji(ctx, sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	}
	if strings.TrimSpace(params.Kolic) == "" || common.StringToFloat64(params.Kolic) == 0 {
		add("kolicina", common.ErrMsgObavezanPodatak)
	}
	return fieldErrors
}

// SaveRobPorStavka saves a stavka of a porudžbenica, a new one (RproID 0, the legacy CreateRpro) or a changed
// one (EditRpro), in one transaction (robPorSnimiStavku).
func (s *RobnoDokumentaResource) SaveRobPorStavka(ctx context.Context, params domain.RobPorStavkaParams) (rproID int64, err error) {
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
	if rproID, err = s.robPorSnimiStavku(ctx, tx, k, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robPorSnimiStavku saves a stavka of a porudžbenica in the transaction (see SaveRobPorStavka):
//   - the iznos is the količina by the nabavna cena, rounded to 2 decimals, negative for a storno;
//   - the stavka has the columns of its document (nalog, vrsta, broj, datum, magacin, dobavljač), the artikal
//     (naziv, JM, rsifid), the nabavna cena (ncena), the količina, the iznos and the tarifa; a new stavka gets
//     the next redni broj (NadjiZadnjRBR);
//   - the nalog and the stanje of the artikal (AzurirajRNAL, AzurirajRSTA: the legacy does not update the
//     zalihe by lot) get the iznos and the količina: a changed stavka is taken back first;
//   - the document gets the iznos of the stavke and a new stavka counts in it (rdok.iznos, rdok.brst).
func (s *RobnoDokumentaResource) robPorSnimiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, params domain.RobPorStavkaParams) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.TrimSpace(params.Konto)
	ncena := common.StringToFloat64(params.Ncena)
	kolicina := math.Abs(common.StringToFloat64(params.Kolic))
	iznos := robZaokruzi(math.Abs(kolicina*ncena), 2)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}
	rsifID, naziv, jm, err := s.robno.RsifIDNazivJm(ctx, tx, sifra)
	if err != nil {
		return 0, err
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
		robPoljeFloat("ncena", ncena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		robPoljeInt("po", params.Tarifa),
	}
	rproID := params.RproID
	if rproID > 0 {
		if err := s.unos.Rpro.UpdateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", rproID, polja); err != nil {
			return 0, err
		}
	} else {
		rbr, rbrErr := s.robno.RproSledeciRbr(ctx, tx, k.RdokID)
		if rbrErr != nil {
			return 0, rbrErr
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
			domain.Fields{Name: "cena", Value: "0"},
			domain.Fields{Name: "rab", Value: "0"},
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""}, domain.Fields{Name: "taxcat", Value: ""}, domain.Fields{Name: "taxexreco", Value: ""},
		)
		var err error
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

// DeleteRobPorStavka deletes a stavka of a porudžbenica (the legacy BrisiRPRO) in one transaction and returns
// its document: the stavka is taken back from its nalog and from the stanje (robPorBrisiStavkuTx) and the
// stavke after it get the redni brojevi again (rbr and brst, from the redni broj of the deleted one).
func (s *RobnoDokumentaResource) DeleteRobPorStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
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
	if err = s.robno.RproPrenumerisiSaBrst(ctx, tx, rdokID, rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// robPorBrisiStavkuTx deletes a stavka of a porudžbenica in the transaction of the caller and returns its
// document (the legacy BrisiRPRO and the stavke of "Brisanje" of the document): the stavka is deleted, its
// iznos is taken back from the nalog (AzurirajRNAL, which also counts the stavke of the nalog again), its
// količina and iznos from the stanje (AzurirajRSTA; the legacy does not update the zalihe by lot), and the
// document loses its iznos and counts one stavka less.
func (s *RobnoDokumentaResource) robPorBrisiStavkuTx(ctx context.Context, tx db.Transaction, rproID int64) (int64, error) {
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
	return rdokID, s.robPorBrisiStavku(ctx, tx, k, rproID)
}

// robPorBrisiStavku deletes the stavka of the document of k in the transaction (see robPorBrisiStavkuTx).
func (s *RobnoDokumentaResource) robPorBrisiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64) error {
	sifra, konto, kolicina, iznos, err := s.robno.RproStaraKolicina(ctx, tx, rproID, k.RdokID)
	if err != nil {
		return err
	}
	if err := s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return err
	}
	if err := s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos); err != nil {
		return err
	}
	if err := s.AzurirajRSTA(ctx, tx, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra,
		Kolicina: -kolicina, Iznos: -iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Brst: -1}); err != nil {
		return err
	}
	return s.robno.RdokOduzmiStavku(ctx, tx, k.RdokID, iznos)
}

// GetRobPorStavkaForma returns a saved stavka of a porudžbenica for the form of the stavke (the legacy
// "Izmene" of rob_PORUDZBENICE_DET): its konto, šifra, količina, nabavna cena and tarifa, the artikal with
// its stopa on the datum of the document and its stanje. found is false without it.
func (s *RobnoDokumentaResource) GetRobPorStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobPorStavkaForma, bool, error) {
	f := domain.RobPorStavkaForma{}
	rows, err := s.robno.RproPorudzbenicaForma(ctx, rproID, rdokID)
	if err != nil || rows == nil || len(*rows) == 0 {
		return f, false, err
	}
	r := (*rows)[0]
	f.RproID, f.Sifra, f.Konto = rproID, common.StringToInt(r.Key), strings.TrimSpace(r.C1)
	f.Kolic, f.Ncena = math.Abs(common.StringToFloat64(r.C2)), common.StringToFloat64(r.C3)
	f.Tarifa, f.Stopa = common.StringToInt(r.C4), common.StringToFloat64(r.C5)
	artikal, aFound, err := s.GetRobPorArtikal(ctx, rdokID, strconv.Itoa(f.Sifra), f.Konto)
	if err != nil {
		return f, true, err
	}
	if aFound {
		f.Naziv, f.Jm, f.Stanje = artikal.Naziv, artikal.Jm, artikal.Stanje
	}
	return f, true, nil
}

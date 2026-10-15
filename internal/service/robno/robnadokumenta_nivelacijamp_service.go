package robno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Nivelacije maloprodaje (the group MNI, the legacy rob_Nivelacija_maloprodaje with its stavke
// rob_Nivelacija_Maloprodaje_DET): a robni dokument (rdok) of a prodavnica (rdok.pkto, rdok.pana, also fkto
// and fana) whose stavke change the MP cena with PDV of an artikal given by its naziv only (the šifra 0, the
// konto 999999): the stara cena (rpro.cena), the nova cena (rpro.mcenap), the stara (rpro.po) and the nova
// poreska tarifa (rpro.dani); the iznos of a stavka is its količina by the nova less the stara cena
// ---------------------------------------------------------------------------------------------------

// robMniKonto is the konto of the stavke of a nivelacija maloprodaje (the legacy RPRO.KONTO = "999999").
const robMniKonto = "999999"

// robMniMsgSifra0 is the message of a firm without the artikal of the šifra 0 (the legacy CreateRpro).
const robMniMsgSifra0 = "Za unos šifre samo sa opisom potrebno je u šifarniku artikala otvoriti artikal sa šifrom 0."

// robMniIznos is the iznos of a stavka of a nivelacija maloprodaje (the legacy CreateRpro and EditRpro): the
// količina by the nova less the stara cena, rounded to 2 decimals.
func robMniIznos(kolicina, staraCena, novaCena float64) float64 {
	return robZaokruzi(kolicina*(novaCena-staraCena), 2)
}

// robMniNovaTarifa is the nova poreska tarifa of a stavka: the one of the form with "Promeni tarifu", else the
// stara one (the legacy exit of EDT_po: EDT_po1 = EDT_po).
func robMniNovaTarifa(params domain.RobMniStavkaParams) string {
	if params.PromeniTarifu {
		return strings.TrimSpace(params.NovaTarifa)
	}
	return strings.TrimSpace(params.StaraTarifa)
}

// RobMniStavkeTableRows returns the stavke of a nivelacija maloprodaje as the rows of the grid of the stavke
// (the columns of GetMniStavkeTableFields, the legacy TABLE_rpro: redni broj, naziv, JM, količina, stara cena,
// nova cena, razlika (the nova less the stara cena), stara tarifa, nova tarifa), by the redni broj. The grid of
// the stavke of the selected document of the tab "Pregled nivelacija" has the same columns.
func RobMniStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Naziv, st.Jm, common.FormatNumberWithSystemLocale(st.Kolic, 3),
				robFormatIznos(st.Cena), robFormatIznos(st.Mcenap), robFormatIznos(st.Mcenap - st.Cena), strconv.Itoa(st.Po), strconv.Itoa(st.Dani)},
		})
	}
	return rows
}

// GetRobMniDokument returns the document of the "Nivelacije maloprodaje" screen: the document of the params like
// the one of the faktura (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one
// with the next broj, NoviBrojRDOK, and the message why; the prodavnica is the kupac of the faktura) with the opis
// of its vrsta. A new document has the konto of the prodavnice of the robno knjigovodstvo (rvr.fprod, the legacy
// entry of EDT_PKTO), no izvorni dokument and the first org. jedinica with its first mesto troška (the legacy
// combos). The stavke repeat the vrsta, the broj and the datum of the document.
func (s *RobnoDokumentaResource) GetRobMniDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoMniDokument, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return domain.RobnoMniDokument{}, fmt.Errorf("no user session found")
	}
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoMniDokument{}, err
	}
	res := domain.RobnoMniDokument{
		Header: domain.RobnoMniHeaderView{RobnoFaktureHeaderView: dok.Header},
		Stavka: domain.RobnoMniStavka{RdokID: dok.Header.RdokID, Vrd: dok.Stavka.Vrd, Brdok: dok.Stavka.Brdok, DatumDokument: dok.Stavka.DatumDokumenta},
		Poruka: dok.Poruka,
	}
	h := &res.Header
	if dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(h.Vrd)); err != nil {
		return res, err
	} else if found {
		h.VrstaOpis = dokvrsta.Opis
	}
	if h.RdokID > 0 {
		return res, nil
	}
	fprod, err := s.robno.RvrFprod(ctx)
	if err != nil {
		return res, err
	}
	h.KupacKonto, h.KupacSifra = strings.TrimSpace(fprod), ""
	h.IzvorniDokDatum = time.Time{}
	return res, nil
}

// ValidateRobMni validates the header of a nivelacija maloprodaje before it is saved, like the legacy
// ValidacijaRDOK of rob_Nivelacija_maloprodaje: the vrsta exists; the datum dokumenta is required and not after
// the datum naloga; the konto and the analitička šifra of the prodavnica are required and exist in the kontni
// plan; a given org. jedinica and mesto troška exist (the mesto troška of the org. jedinica; not required);
// the broj dokumenta is a whole number.
func (s *RobnoDokumentaResource) ValidateRobMni(ctx context.Context, params domain.RobnoMniParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	if _, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(params.Vrd)); err != nil {
		add("vrd", common.ErrMsgReadData)
	} else if !found {
		add("vrd", robnoUnosMsgNePostoji)
	}
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}
	s.robOpstiBrojDokumenta(params.BrojDokum, add)

	// The prodavnica: the konto and the šifra of the kontni plan.
	konto, sifra := strings.TrimSpace(params.ProdavnicaKonto), strings.TrimSpace(params.ProdavnicaSifra)
	switch {
	case konto == "":
		add("kupackonto", common.ErrMsgObavezanPodatak)
	case sifra == "":
		add("kupacsifra", common.ErrMsgObavezanPodatak)
	default:
		if exists, err := s.robno.FkplAnalitikaPostoji(ctx, konto, sifra); err != nil {
			add("kupacsifra", common.ErrMsgReadData)
		} else if !exists {
			add("kupacsifra", "Prodavnica sa takvom šifrom ne postoji.")
		}
	}

	// The org. jedinica and the mesto troška are not required (empty or "-": NULL); given, the checks of the opšti
	// dokument.
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)
	return fieldErrors
}

// SaveRobMni saves the header of a nivelacija maloprodaje (the legacy "Sačuvaj" of rob_Nivelacija_maloprodaje,
// after ValidacijaRDOK: SnimiRDOK) and returns its rdokid: the robni dokument with the datum, the izvorni
// dokument, the opis, the prodavnica (as pkto, pana and as fkto, fana), the org. jedinica and the mesto troška,
// a new one when the nalog has no document of the vrsta and broj in the magacin (SaveRobniDokument, counted in
// the nalog) with no rok, otpremnica, kamata, plaćanje, footer or napomena o poreskom oslobođenju.
func (s *RobnoDokumentaResource) SaveRobMni(ctx context.Context, params domain.RobnoMniParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	konto, sifra := strings.TrimSpace(params.ProdavnicaKonto), strings.TrimSpace(params.ProdavnicaSifra)
	return s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        common.StringToInt(params.Vrd),
		Dokum:      common.StringToInt(params.BrojDokum),
		Polja: []domain.Fields{
			robPoljeDatum("dadok", params.DatumDokum),
			robPoljeText("opis", params.Opis),
			robPoljeDatum("datiz", params.IzvorniDokDatum),
			robPoljeText("dokiz", params.IzvorniDokum),
			robPoljeText("fkto", konto),
			robPoljeText("fana", sifra),
			robPoljeText("pkto", konto),
			robPoljeText("pana", sifra),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
		},
		NoviPolja: []domain.Fields{
			{Name: "rok", Value: "0"},
			{Name: "otp", Value: ""},
			{Name: "kamt", Value: ""},
			{Name: "pla", Value: ""},
			{Name: "foot", Value: ""},
			{Name: "pornapomena", Value: ""},
		},
	})
}

// DeleteRobMni deletes a nivelacija maloprodaje with its stavke (the legacy BTN_Brisanje of
// rob_Nivelacija_maloprodaje): a document printed on the fiskalni printer (knjige[4] "S") or posted (knjige[1]
// "D") is not deleted (poruka says why); else every stavka is taken back from the nalog and from the stanje of
// the konto 999999 (robPorBrisiStavkuTx: its količina and iznos) and deleted, the document is deleted and the
// nalog counts one document less (robBrisiDokumentSa).
func (s *RobnoDokumentaResource) DeleteRobMni(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokumentSa(ctx, rdokID, false, s.robPorBrisiStavkuTx)
}

// ---------------------------------------------------------------------------------------------------
// The stavke of a nivelacija maloprodaje (the legacy rob_Nivelacija_Maloprodaje_DET)
// ---------------------------------------------------------------------------------------------------

// robMniProveriStavku checks the values of the form of a stavka of a nivelacija maloprodaje (the legacy
// Validacija of rob_Nivelacija_Maloprodaje_DET): the količina is not 0, the stara and the nova cena are greater
// than 0, the stara and the nova tarifa (the stara one without "Promeni tarifu") are given and not 0.
func robMniProveriStavku(params domain.RobMniStavkaParams, add func(field, message string)) {
	if strings.TrimSpace(params.Kolic) == "" || common.StringToFloat64(params.Kolic) == 0 {
		add("kolicina", "Morate uneti količinu artikla.")
	}
	if common.StringToFloat64(params.StaraCena) <= 0 {
		add("staracena", "Morate uneti staru cenu artikla.")
	}
	if common.StringToFloat64(params.NovaCena) <= 0 {
		add("novacena", "Morate uneti novu cenu artikla.")
	}
	tarifa := func(field, value string) {
		if po, err := strconv.Atoi(strings.TrimSpace(value)); err != nil || po == 0 {
			add(field, "Morate uneti šifru poreske stope.")
		}
	}
	tarifa("staratarifa", params.StaraTarifa)
	if params.PromeniTarifu {
		tarifa("novatarifa", params.NovaTarifa)
	}
}

// ValidateRobMniStavka validates a stavka of a nivelacija maloprodaje before it is saved: the document exists,
// the values of the form (robMniProveriStavku), the tarife exist (rpor) and the firm has the artikal of the
// šifra 0 (the legacy CreateRpro: robMniMsgSifra0).
func (s *RobnoDokumentaResource) ValidateRobMniStavka(ctx context.Context, params domain.RobMniStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("nazivmni", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	if _, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID); err != nil {
		add("nazivmni", common.ErrMsgReadData)
		return fieldErrors
	} else if !found {
		add("nazivmni", "Dokument ne postoji.")
		return fieldErrors
	}
	robMniProveriStavku(params, add)
	if len(fieldErrors) > 0 {
		return fieldErrors
	}
	tarife := [][2]string{{"staratarifa", params.StaraTarifa}}
	if params.PromeniTarifu {
		tarife = append(tarife, [2]string{"novatarifa", params.NovaTarifa})
	}
	for _, t := range tarife {
		if exists, err := s.robno.RporPostoji(ctx, common.StringToInt(t[1])); err != nil {
			add(t[0], common.ErrMsgReadData)
		} else if !exists {
			add(t[0], "Poreska tarifa ne postoji.")
		}
	}
	if exists, err := s.robno.RsifPostoji(ctx, 0); err != nil {
		add("nazivmni", common.ErrMsgReadData)
	} else if !exists {
		add("nazivmni", robMniMsgSifra0)
	}
	return fieldErrors
}

// SaveRobMniStavka saves a stavka of a nivelacija maloprodaje, a new one (RproID 0, the legacy CreateRpro) or a
// changed one ("Izmeni" of its row, the legacy EditRpro), in one transaction (robMniSnimiStavku).
func (s *RobnoDokumentaResource) SaveRobMniStavka(ctx context.Context, params domain.RobMniStavkaParams) (rproID int64, err error) {
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
	if rproID, err = s.robMniSnimiStavku(ctx, tx, k, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robMniSnimiStavku saves a stavka of a nivelacija maloprodaje in the transaction (see SaveRobMniStavka):
//   - the iznos is the količina by the nova less the stara cena (robMniIznos); for a storno the iznos and the
//     količina are negative;
//   - the stavka is the artikal of the šifra 0 in the konto 999999 with the naziv and the JM of the form, the
//     stara cena (cena), the nova cena (mcenap), the stara tarifa (po) and the nova one (dani), the columns of
//     its document and, a new one, the next redni broj;
//   - the nalog and the stanje of the šifra 0 in the konto 999999 (AzurirajRNAL, AzurirajRSTA; the šifra 0 has
//     no zalihe by lot, the SR, RSAL and RSIN are not kept) get the iznos and the količina: a changed stavka is
//     taken back first;
//   - the document gets the iznos of the stavke and a new stavka counts in it (rdok.iznos, rdok.brst).
func (s *RobnoDokumentaResource) robMniSnimiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, params domain.RobMniStavkaParams) (int64, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return 0, fmt.Errorf("no user session found")
	}
	staraCena, novaCena := common.StringToFloat64(params.StaraCena), common.StringToFloat64(params.NovaCena)
	kolicina := common.StringToFloat64(params.Kolic)
	iznos := robMniIznos(kolicina, staraCena, novaCena)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}
	rsifID, _, found, err := s.robno.RsifIDPo(ctx, tx, 0)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, errors.New(robMniMsgSifra0)
	}

	// A changed stavka: its old values are taken back from the stanje.
	stara := domain.RobPromenaStanja{}
	if params.RproID > 0 {
		if stara.Sifra, stara.Konto, stara.Kolicina, stara.Iznos, err = s.robno.RproStaraKolicina(ctx, tx, params.RproID, k.RdokID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("stavka ne postoji")
			}
			return 0, err
		}
	}
	polja := []domain.Fields{
		{Name: "sifra", Value: "0"},
		{Name: "konto", Value: robMniKonto},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		robPoljeText("naz1", params.Naziv),
		robPoljeText("jm", params.Jm),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("cena", staraCena),
		robPoljeFloat("mcenap", novaCena),
		robPoljeFloat("iznos", iznos),
		robPoljeInt("po", params.StaraTarifa),
		robPoljeInt("dani", robMniNovaTarifa(params)),
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
			domain.Fields{Name: "rab", Value: "0"},
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
		if err := s.AzurirajRSTA(ctx, tx, promena(stara.Sifra, strings.TrimSpace(stara.Konto), -stara.Kolicina, -stara.Iznos, -1)); err != nil {
			return 0, err
		}
	}
	if err := s.AzurirajRSTA(ctx, tx, promena(0, robMniKonto, kolicina, iznos, 1)); err != nil {
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

// DeleteRobMniStavka deletes a stavka of a nivelacija maloprodaje (the legacy BrisiRPRO of
// rob_Nivelacija_Maloprodaje_DET) and returns its document: the deletion of a stavka of the porudžbenica
// (DeleteRobPorStavka: its količina and iznos taken back from the nalog and from the stanje of its konto, the
// document with a stavka less, the stavke after it renumbered).
func (s *RobnoDokumentaResource) DeleteRobMniStavka(ctx context.Context, rproID int64) (int64, error) {
	return s.DeleteRobPorStavka(ctx, rproID)
}

// GetRobMniStavkaForma returns a saved stavka of a nivelacija maloprodaje for the form of the stavke ("Izmeni" of
// its row): its naziv, JM, količina (as entered: a storno has it saved negative), cene and tarife with their
// stope. found is false without it.
func (s *RobnoDokumentaResource) GetRobMniStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobMniStavkaForma, bool, error) {
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return domain.RobMniStavkaForma{}, false, err
	}
	f, err := s.robno.RproNivelacijaMPForma(ctx, rproID, rdokID)
	if errors.Is(err, sql.ErrNoRows) {
		return f, false, nil
	}
	if err != nil {
		return f, false, err
	}
	if k.Storno() {
		f.Kolicina = -f.Kolicina
	}
	return f, true, nil
}

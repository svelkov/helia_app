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
// Zaduženja sitnog inventara (the group SIV, the legacy rob_ZaduzSitanInventar): the sitan inventar of the magacin of
// the nalog charged to a radnik (a konto and its analitička šifra of the kontni plan, rdok.pkto and rdok.pana), with
// the org. jedinica and the mesto troška of the document and of the roba (where the roba goes). The screen is the
// one of the opšti dokumenti (Siv) with the stavke of the opšti dokument (the konto, the artikal, the količina, the
// cena, the nabavna cena and the lot in a magacin of the tip zaliha 2).
// ---------------------------------------------------------------------------------------------------

// robSivDokument fills the radnik of a saved zaduženje sitnog inventara (robOpstiRadnik) and the cena of its stavke:
// the cena and the nabavna cena are entered (the legacy EDT_CENA, EDT_NCENA).
func (s *RobnoDokumentaResource) robSivDokument(ctx context.Context, res *domain.RobnoOpstiDokument) error {
	if err := s.robOpstiRadnik(ctx, &res.Header); err != nil {
		return err
	}
	st := &res.Stavka
	st.CenaNaziv, st.CenaZakljucana, st.NabCena = "Cena", false, false
	return nil
}

// robSivValidate validates the header of a zaduženje sitnog inventara (the legacy ValidacijaRDOK of
// rob_ZaduzSitanInventar): the datum dokumenta is required and not after the datum naloga; the org. jedinica and the
// mesto troška of the roba are required, the ones of the document are not; given, they exist (the mesto troška of its
// org. jedinica); the
// konto and the analitička šifra of the radnik are required and exist in the kontni plan (a saved document keeps its
// radnik: the legacy ModifyRDOK does not change it); the broj dokumenta is a whole number.
func (s *RobnoDokumentaResource) robSivValidate(ctx context.Context, params domain.RobnoOpstiParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if domain.GetSessionFromStdContext(ctx) == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}
	// The org. jedinica and the mesto troška of the document are not required (empty or "-": NULL), the ones of the
	// roba are; given, the checks of the opšti dokument.
	for _, polje := range [][2]string{{"idorgjed1", params.Orgjed1}, {"mestotrid1", params.MestoTroska1}} {
		if _, given, ok := robnoUnosID(polje[1]); ok && !given {
			add(polje[0], common.ErrMsgObavezanPodatak)
		}
	}
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)
	s.robOpstiOrgjedMt(ctx, params.Orgjed1, params.MestoTroska1, "idorgjed1", "mestotrid1", add)
	if params.RdokID <= 0 {
		pkto, pana := strings.TrimSpace(params.Pkto), robGrdSifra(params.Pana)
		switch {
		case pkto == "" || pkto == "-":
			add("pkto", common.ErrMsgObavezanPodatak)
		case pana == "":
			add("pana", common.ErrMsgObavezanPodatak)
		default:
			if exists, err := s.robno.FkplAnalitikaTrimPostoji(ctx, pkto, pana); err != nil {
				add("pana", common.ErrMsgReadData)
			} else if !exists {
				add("pana", "Ne postoji u kontnom planu.")
			}
		}
	}
	s.robOpstiBrojDokumenta(params.BrojDokum, add)
	return fieldErrors
}

// robSivSaveHeader saves the header of a zaduženje sitnog inventara and returns its rdokid: a new document of the
// nalog (the legacy SnimiRDOK: SaveRobniDokument, counted in the nalog) with the datumi, the opis, the izvorni
// dokument, the org. jedinice and mesta troška, the radnik, no rok, magacin where the roba goes, kupac or napomena o
// poreskom oslobođenju; an existing one (the legacy ModifyRDOK) gets the datumi, the opis, the izvorni dokument and
// the org. jedinice and mesta troška (not the radnik).
func (s *RobnoDokumentaResource) robSivSaveHeader(ctx context.Context, params domain.RobnoOpstiParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	vrd, dokum := common.StringToInt(params.Vrd), common.StringToInt(params.BrojDokum)
	postoji, err := s.robno.RdokBrojPostoji(ctx, params.MagaciniID, vrd, dokum)
	if err != nil {
		return 0, err
	}
	polja := []domain.Fields{
		robPoljeDatum("dadok", params.DatumDokum),
		robPoljeText("opis", params.Opis),
		robPoljeDatum("datiz", params.IzvorniDokDatum),
		robPoljeText("dokiz", params.IzvorniDokum),
		{Name: "rok", Value: "0"},
		robPoljeID("idorgjed", id(params.Orgjed)),
		robPoljeID("mestotrid", id(params.MestoTroska)),
		robPoljeID("idorgjed1", id(params.Orgjed1)),
		robPoljeID("mestotrid1", id(params.MestoTroska1)),
	}
	if !postoji {
		polja = append(polja,
			domain.Fields{Name: "magid1", Value: "0"},
			domain.Fields{Name: "fkto", Value: ""},
			domain.Fields{Name: "fana", Value: ""},
			domain.Fields{Name: "pornapomena", Value: ""},
			robPoljeText("pkto", strings.TrimSpace(params.Pkto)),
			robPoljeText("pana", robGrdSifra(params.Pana)),
		)
	}
	return s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{RnalID: params.RnalID, MagaciniID: params.MagaciniID, Vrd: vrd, Dokum: dokum, Polja: polja})
}

// robSivArtikal returns the artikal of a stavka of a zaduženje sitnog inventara (the legacy exit of EDT_SIFRA of
// rob_ZaduzSitanInventar): the artikal by its šifra with its naziv and JM (the konto of the form), and
//   - in a magacin of the tip zaliha 2 the stanje, the cena and the nabavna cena of the lot of the form (drsta) when
//     it has a stanje and the document is not a storno, else 0;
//   - else the stanje, the cena and the nabavna cena in the konto of the magacin (rsta), 0 without it;
//
// a nabavna cena 0 of the stanje is the one of the cene of the artikal (rcene), and then a cena 0 too. found is false
// for an artikal that does not exist.
func (s *RobnoDokumentaResource) robSivArtikal(ctx context.Context, k domain.RobStavkeKontekst, sifra, konto, otk, serija, rok string) (domain.RobOpstiArtikal, bool, error) {
	a := domain.RobOpstiArtikal{}
	if domain.GetSessionFromStdContext(ctx) == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	rows, err := s.robno.RsifArtikalSaCenama(ctx, sifra)
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm, a.Konto = common.StringToInt(r.Key), r.C1, r.C2, strings.TrimSpace(konto)
	found := false
	if k.Tipzal == 2 {
		datum, _ := robnoUnosDatum(rok)
		stanje, cena, ncena, lot, err := s.robno.DrstaLotCene(ctx, k.Mag, a.Konto, a.Sifra, otk, serija, RobWinDevDatum(datum))
		if err != nil {
			return a, true, err
		}
		if found = lot && stanje > 0 && !k.Storno(); found {
			a.Stanje, a.Cena, a.Ncena = stanje, cena, ncena
		}
	} else {
		st, err := s.robOpstiStanje(ctx, k.MagaciniID, a.Konto, a.Sifra)
		if err != nil {
			return a, true, err
		}
		if found = st.found; found {
			a.Stanje, a.Cena, a.Ncena = st.ulaz-st.izlaz, st.cena, st.ncena
		}
	}
	if found && a.Ncena == 0 {
		a.Ncena = common.StringToFloat64(r.C5)
		if a.Cena == 0 {
			a.Cena = common.StringToFloat64(r.C4)
		}
	}
	return a, true, nil
}

// robSivStavka adds a stavka to a zaduženje sitnog inventara in one transaction (the legacy CreateRpro of
// rob_ZaduzSitanInventar):
//   - the iznos is the količina by the cena (in a magacin of the sirovine, tipmag "S", by the nabavna cena), rounded
//     to 2 decimals, negative with the količina for a storno;
//   - the stavka has the columns of its document (the radnik as pkto and pana), the artikal, the konto, the cena, the
//     nabavna cena, the lot (the OTK broj, the serija and the rok), the next redni broj of the document;
//   - the nalog, the stanje (without a cena) and, in a magacin of the tip zaliha 2, the zalihe by the lot get it
//     (AzurirajRNAL, AzurirajRSTA, AzurirajDRSTA), the document its iznos and a stavka more. The SR, RSAL and RSIN are
//     not kept.
func (s *RobnoDokumentaResource) robSivStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	jedinicna := cena
	if strings.EqualFold(strings.TrimSpace(k.Tipmag), "S") {
		jedinicna = ncena
	}
	kolicina := math.Abs(common.StringToFloat64(params.Kolic))
	iznos := math.Abs(robZaokruzi(kolicina*jedinicna, 2))
	if k.Storno() {
		kolicina, iznos = -kolicina, -iznos
	}
	rok, _ := robnoUnosDatum(params.Rok)
	roktr := RobWinDevDatum(rok)

	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	rsifID, po, found, err := s.robno.RsifIDPo(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("artikal ne postoji")
	}
	pkto, pana, err := s.robno.RdokPktoPanaTx(ctx, tx, k.RdokID)
	if err != nil {
		return 0, err
	}
	rbr, err := s.robno.RproSledeciRbrNaloga(ctx, tx, k.RnalID, k.RdokID)
	if err != nil {
		return 0, err
	}
	if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
		{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
		{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
		robPoljeNullTime("dadok", k.Dadok),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "konto", Value: konto},
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		robPoljeFloat("cena", cena),
		robPoljeFloat("ncena", ncena),
		{Name: "rab", Value: "0"},
		{Name: "po", Value: strconv.Itoa(po)},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		// The legacy RPRO.BRST + 1 of the new stavka.
		{Name: "brst", Value: "1"},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(roktr, 10)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	}); err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos); err != nil {
		return 0, err
	}
	if err = s.robSivStanje(ctx, tx, k, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra, Kolicina: kolicina,
		Iznos: iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Otk: strings.TrimSpace(params.Otk), Serija: strings.TrimSpace(params.Serija), Roktr: roktr, Brst: 1}); err != nil {
		return 0, err
	}
	if err = s.robno.RdokDodajIznos(ctx, tx, k.RdokID, iznos, 1); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robSivStanje changes the stanje of the artikal of a stavka of a zaduženje sitnog inventara (AzurirajRSTA) and, in a
// magacin of the tip zaliha 2, its zalihe by the lot of the stavka (AzurirajDRSTA; the legacy does not change the
// zalihe of a magacin of the tip zaliha 1).
func (s *RobnoDokumentaResource) robSivStanje(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, p domain.RobPromenaStanja) error {
	if err := s.AzurirajRSTA(ctx, tx, p); err != nil {
		return err
	}
	if k.Tipzal != 2 {
		return nil
	}
	return s.AzurirajDRSTA(ctx, tx, p)
}

// robSivBrisi deletes a stavka of a zaduženje sitnog inventara in the transaction tx (the legacy BrisiRPRO of
// rob_ZaduzSitanInventar): it is taken back from the stanje and, in a magacin of the tip zaliha 2, from the zalihe by
// its lot (robSivStanje), from its document (the iznos, a stavka less) and from its nalog (AzurirajRNAL, after the
// deletion), it is deleted and the stavke after it get the redni broj less by one.
func (s *RobnoDokumentaResource) robSivBrisi(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64) error {
	st, err := s.robno.RproUslugaBrisanje(ctx, tx, rproID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("stavka ne postoji")
	}
	if err != nil {
		return err
	}
	if err = s.robSivStanje(ctx, tx, k, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: strings.TrimSpace(st.Konto), Sifra: st.Sifra,
		Kolicina: -st.Kolicina, Iznos: -st.Iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Otk: strings.TrimSpace(st.Otk), Serija: strings.TrimSpace(st.Serija), Roktr: int64(st.Roktr), Brst: -1}); err != nil {
		return err
	}
	if err = s.robno.RdokOduzmiStavku(ctx, tx, k.RdokID, st.Iznos); err != nil {
		return err
	}
	if err = s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -st.Iznos); err != nil {
		return err
	}
	return s.robno.RproSmanjiRbr(ctx, tx, k.RdokID, st.Rbr)
}

// RobSivDokumentaPrikazTableRows returns the zaduženja sitnog inventara as the rows of the grid of the documents of the
// tab "Pregled zaduž. inventara radnika" (the columns of GetSivDokumentaTableFields, the legacy TABLE_QRY_RDOK of
// rob_ZaduzSitanInventar: vrsta and broj naloga, vrsta, broj and datum dokumenta, iznos, opis, datum obrade, magacin,
// broj and datum izvornog dokumenta, operater).
func RobSivDokumentaPrikazTableRows(dokumenta []domain.RobDokumentPrikazRow) []domain.TableRow {
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
			Fields: []string{d.Tipdok, broj(d.Nalog), broj(d.Vrd), broj(d.Dokum), robFormatDatum(d.Dadok), robFormatIznos(d.Iznos), d.Opis,
				robFormatDatum(d.Datob), broj(d.Mag), d.Dokiz, robFormatDatum(d.Datiz), d.Oper},
		})
	}
	return rows
}

// RobSivPrikazStavkeTableRows returns the stavke of the selected zaduženje sitnog inventara of the tab "Pregled zaduž.
// inventara radnika" (the columns of GetSivPrikazStavkeTableFields, the legacy TABLE_QRY_RPRO: redni broj, konto,
// šifra, naziv, JM, količina, VP cena, VP iznos, nabavna cena and the nabavni iznos, the količina by the nabavna
// cena), by the redni broj.
func RobSivPrikazStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm, common.FormatNumberWithSystemLocale(st.Kolic, 3),
				robFormatIznos(st.Cena), robFormatIznos(st.Iznos), robFormatIznos(st.Ncena), robFormatIznos(robZaokruzi(st.Kolic*st.Ncena, 2))},
		})
	}
	return rows
}

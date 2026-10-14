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
// Zaduženja gradilišta (the group GRD, the legacy rob_ZadauzGRADILISTA): the roba of the magacin of the nalog
// charged to a gradilište, a konto and its analitička šifra of the kontni plan (rdok.pkto, rdok.pana). The
// screen is the one of the opšti dokumenti (Grd): the stavke have the konto, the artikal, the količina and the
// cena, their nabavna cena is the one of the last prenosnica ulaz of the artikal. The trebovanja (the group OTR,
// the legacy rob_Trebovanja) are the same with the konto and the šifra of the trošak (otr): a saved one gets all
// the fields of the header, its stavke have no check of the cena and of the zalihe, keep their lot and change
// the stanje only (not the zalihe by lot, the cena of the stanje not).
// ---------------------------------------------------------------------------------------------------

// robGrdDokument fills the gradilište of a zaduženje gradilišta (rdok.pkto and rdok.pana with the naziv of the
// kontni plan: the one of the šifra, else of the konto, the legacy exits of EDT_PKTO and EDT_PANA); a new one has
// no gradilište and the datum izvornog dokumenta of the nalog (the legacy exit of EDT_DOKUM). The cena of the
// stavke is entered.
func (s *RobnoDokumentaResource) robGrdDokument(ctx context.Context, res *domain.RobnoOpstiDokument) error {
	h := &res.Header
	if h.RdokID > 0 {
		var err error
		if h.Pkto, h.Pana, err = s.robno.RdokPktoPana(ctx, h.RdokID); err != nil {
			return err
		}
		if h.Pkto != "" {
			if h.GradilisteNaziv, err = s.robOpstiNazivKonta(ctx, h.Pkto, h.Pana); err != nil {
				return err
			}
		}
	} else if rnal, found, err := s.robnoUnosNalog(ctx, h.RnalID); err != nil {
		return err
	} else if found && rnal.Danal.Valid {
		h.IzvorniDokDatum = rnal.Danal.Time
	}
	res.Stavka.CenaNaziv, res.Stavka.CenaZakljucana, res.Stavka.NabCena = "Cena", false, false
	return nil
}

// robGrdValidate validates the header of a zaduženje gradilišta (the legacy ValidacijaRDOK of
// rob_ZadauzGRADILISTA): the datum dokumenta is required and not after the datum naloga; the konto of the
// gradilište is required and the konto with the šifra of the gradilište exist in the kontni plan; the org.
// jedinica and the mesto troška are not required (empty or "-" are saved as NULL), a given one exists; the broj
// dokumenta is a whole number.
func (s *RobnoDokumentaResource) robGrdValidate(ctx context.Context, params domain.RobnoOpstiParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
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
	if pkto := strings.TrimSpace(params.Pkto); pkto == "" || pkto == "-" {
		add("pkto", common.ErrMsgObavezanPodatak)
	} else if exists, err := s.robno.FkplAnalitikaTrimPostoji(ctx, pkto, robGrdSifra(params.Pana)); err != nil {
		add("pkto", common.ErrMsgReadData)
	} else if !exists {
		add("pana", "Ne postoji u kontnom planu.")
	}
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)
	s.robOpstiBrojDokumenta(params.BrojDokum, add)
	return fieldErrors
}

// robGrdSifra is the šifra of the gradilište of the form ("" for none: the empty combo has "-").
func robGrdSifra(pana string) string {
	if pana = strings.TrimSpace(pana); pana == "-" {
		return ""
	}
	return pana
}

// robGrdSaveHeader saves the header of a zaduženje gradilišta (the legacy SnimiRDOK of rob_ZadauzGRADILISTA): a
// new document of the nalog (SaveRobniDokument: counted in the nalog) with the datumi, the opis, the izvorni
// dokument, the gradilište, the org. jedinica and the mesto troška, no rok, magacin where the roba goes, kupac or
// napomena o poreskom oslobođenju; an existing one gets the izvorni dokument, the opis and the gradilište only, of
// a trebovanje (otr, the legacy SnimiRDOK of rob_Trebovanja: ScreenToFile) also the datum dokumenta, the org.
// jedinica and the mesto troška. It returns the rdokid of the document.
func (s *RobnoDokumentaResource) robGrdSaveHeader(ctx context.Context, params domain.RobnoOpstiParams, otr bool) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
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
		robPoljeText("dokiz", params.IzvorniDokum),
		robPoljeDatum("datiz", params.IzvorniDokDatum),
		robPoljeText("opis", params.Opis),
		robPoljeText("pkto", strings.TrimSpace(params.Pkto)),
		robPoljeText("pana", robGrdSifra(params.Pana)),
	}
	if postoji && otr {
		polja = append(polja,
			robPoljeDatum("dadok", params.DatumDokum),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
		)
	}
	if !postoji {
		polja = append(polja,
			robPoljeDatum("dadok", params.DatumDokum),
			domain.Fields{Name: "rok", Value: "0"},
			domain.Fields{Name: "magid1", Value: "0"},
			domain.Fields{Name: "fkto", Value: ""},
			domain.Fields{Name: "fana", Value: ""},
			domain.Fields{Name: "pornapomena", Value: ""},
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
		)
	}
	return s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{RnalID: params.RnalID, MagaciniID: params.MagaciniID, Vrd: vrd, Dokum: dokum, Polja: polja})
}

// robGrdZadnjiUlaz returns the cena and the nabavna cena of the artikal of its last prenosnica ulaz (the vrsta
// 113, its last datum dokumenta in the period; the legacy NadjiZadnjuVPC: the nabavna cena is the veleprodajna
// cena of the magacin the roba came from); 0 without it.
func (s *RobnoDokumentaResource) robGrdZadnjiUlaz(ctx context.Context, sifra int) (cena, ncena float64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, 0, fmt.Errorf("no user session found")
	}
	return s.robno.RproZadnjiUlaz(ctx, robPreKontraVrd, sifra)
}

// robGrdArtikal returns the artikal of a stavka of a zaduženje gradilišta (the legacy exit of EDT_SIFRA of
// rob_ZadauzGRADILISTA): the artikal by its šifra with its naziv and JM (the konto of the form), its stanje in the
// konto of the magacin (rsta) and the cena of its last prenosnica ulaz (robGrdZadnjiUlaz: the legacy
// NadjiZadnjuVPC sets the cena after the one of the stanje). found is false for an artikal that does not exist.
func (s *RobnoDokumentaResource) robGrdArtikal(ctx context.Context, k domain.RobStavkeKontekst, sifra, konto string) (domain.RobOpstiArtikal, bool, error) {
	a := domain.RobOpstiArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	sif, naziv, jm, found, err := s.robno.RsifArtikalPoTekstu(ctx, sifra)
	if err != nil || !found {
		return a, false, err
	}
	a.Sifra, a.Naziv, a.Jm, a.Konto = sif, naziv, jm, strings.TrimSpace(konto)
	st, err := s.robOpstiStanje(ctx, k.MagaciniID, a.Konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if st.found {
		a.Stanje = st.ulaz - st.izlaz
	}
	if a.Cena, a.Ncena, err = s.robGrdZadnjiUlaz(ctx, a.Sifra); err != nil {
		return a, true, err
	}
	return a, true, nil
}

// robGrdValidateStavka validates a stavka of a zaduženje gradilišta (the legacy Validacija of
// rob_ZadauzGRADILISTA): the konto is required and a konto of the magacin; the šifra is required and the artikal
// exists; the količina and the cena are greater than 0; a document that is not a storno has the količina in the
// stanje of the konto (the legacy "Nema dovoljno kolicina na zalihama!!!"). A trebovanje (otr, the legacy
// Validacija of rob_Trebovanja) checks the količina only (its checks of the cena and of the zalihe are commented
// out).
func (s *RobnoDokumentaResource) robGrdValidateStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams, otr bool) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("sifraartikla", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	if konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	} else if konta, err := s.GetRobKontaMagacina(ctx, k.MagaciniID); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if !robSadrzi(konta, konto) {
		add("kontostavke", "Nije konto magacina.")
	}
	sifra, sifraOK := 0, false
	if sifraText := strings.TrimSpace(params.Sifra); sifraText == "" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if v, err := strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.RsifPostoji(ctx, v); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else {
		sifra, sifraOK = v, true
	}
	kolicina := common.StringToFloat64(params.Kolic)
	if strings.TrimSpace(params.Kolic) == "" && !otr {
		add("kolicina", common.ErrMsgObavezanPodatak)
	} else if kolicina <= 0 {
		add("kolicina", robInterniMsgVeceOdNule)
	}
	if otr {
		return fieldErrors
	}
	if common.StringToFloat64(params.Cena) <= 0 {
		add("cenaopd", robInterniMsgVeceOdNule)
	}
	if sifraOK && konto != "" && kolicina > 0 && !k.Storno() {
		st, err := s.robOpstiStanje(ctx, k.MagaciniID, konto, sifra)
		if err != nil {
			add("kolicina", common.ErrMsgReadData)
		} else if stanje := st.ulaz - st.izlaz; stanje <= 0 || robZaokruzi(stanje, 3)-robZaokruzi(kolicina, 3) < 0 {
			add("kolicina", robPreMsgZalihe)
		}
	}
	return fieldErrors
}

// robGrdStavka adds a stavka to a zaduženje gradilišta in one transaction (the legacy CreateRpro of
// rob_ZadauzGRADILISTA): the iznos is the količina by the cena (negative with the količina for a storno); the
// stavka has the columns of its document (the gradilište as pkto and pana), the artikal, the konto, the cena, the
// nabavna cena of the last prenosnica ulaz of the artikal and the next redni broj; the nalog, the stanje (with the
// cena) and the zalihe (drsta, without the lot) of the magacin get it (AzurirajRNAL, AzurirajRSTA and
// AzurirajDRSTA), the document its iznos and a stavka more. A trebovanje (otr, the legacy CreateRpro of
// rob_Trebovanja) keeps the lot of the stavka (the OTK broj, the serija and the rok of the form) and changes the
// stanje without the cena, not the zalihe by lot. The SR, RSAL and RSIN are not kept.
func (s *RobnoDokumentaResource) robGrdStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams, otr bool) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	_, ncena, err := s.robGrdZadnjiUlaz(ctx, sifra)
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
	if rproID, err = s.robGrdDodajStavku(ctx, tx, k, params, ncena, otr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robGrdDodajStavku adds a stavka with the nabavna cena ncena to the zaduženje gradilišta of k in the transaction
// tx (see robGrdStavka).
func (s *RobnoDokumentaResource) robGrdDodajStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams,
	ncena float64, otr bool) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	cena := common.StringToFloat64(params.Cena)
	kolicina := math.Abs(common.StringToFloat64(params.Kolic))
	iznos := math.Abs(robZaokruzi(kolicina*cena, 2))
	if k.Storno() {
		kolicina, iznos = -kolicina, -iznos
	}
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
	polja := []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
		{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
		{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
		robPoljeNullTime("dadok", k.Dadok),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "konto", Value: konto},
		robPoljeFloat("cena", cena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		robPoljeFloat("ncena", ncena),
		{Name: "rab", Value: "0"},
		{Name: "po", Value: strconv.Itoa(po)},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		{Name: "brst", Value: strconv.Itoa(rbr)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	}
	if otr {
		rok, _ := robnoUnosDatum(params.Rok)
		polja = append(polja, robPoljeText("otk", params.Otk), robPoljeText("serija", params.Serija),
			domain.Fields{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)})
	}
	rproID, err := s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja)
	if err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos); err != nil {
		return 0, err
	}
	p := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra, Kolicina: kolicina, Iznos: iznos, Cena: cena,
		Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Brst: 1}
	if otr {
		p.Cena = 0
	}
	if err = s.AzurirajRSTA(ctx, tx, p); err != nil {
		return 0, err
	}
	if !otr {
		p.Cena = 0
		if err = s.AzurirajDRSTA(ctx, tx, p); err != nil {
			return 0, err
		}
	}
	if err = s.robno.RdokDodajIznos(ctx, tx, k.RdokID, iznos, 1); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robGrdBrisi deletes a stavka of a zaduženje gradilišta in the transaction tx (the legacy BrisiRPRO of
// rob_ZadauzGRADILISTA): it is taken back from the stanje and the zalihe of the magacin (AzurirajRSTA,
// AzurirajDRSTA; a trebovanje, otr, from the stanje only), from its document (the iznos, a stavka less) and from
// its nalog (AzurirajRNAL, after the deletion), it is deleted and the stavke after it get the redni broj less by
// one.
func (s *RobnoDokumentaResource) robGrdBrisi(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64, otr bool) error {
	sifra, rbr, konto, kolicina, iznos, err := s.robno.RproStavkaBrisanje(ctx, tx, rproID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("stavka ne postoji")
	}
	if err != nil {
		return err
	}
	p := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: strings.TrimSpace(konto), Sifra: sifra, Kolicina: -kolicina, Iznos: -iznos,
		Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Brst: -1}
	if err = s.AzurirajRSTA(ctx, tx, p); err != nil {
		return err
	}
	if !otr {
		if err = s.AzurirajDRSTA(ctx, tx, p); err != nil {
			return err
		}
	}
	if err = s.robno.RdokOduzmiStavku(ctx, tx, k.RdokID, iznos); err != nil {
		return err
	}
	if err = s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos); err != nil {
		return err
	}
	return s.robno.RproSmanjiRbr(ctx, tx, k.RdokID, rbr)
}

// RobGrdStavkeTableRows returns the stavke of a zaduženje gradilišta as the rows of the grid of the stavke (the
// columns of GetOpstiGrdStavkeTableFields: redni broj, konto, šifra, naziv, JM, količina, cena, vrednost).
func RobGrdStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := RobOpstiStavkeTableRows(stavke)
	for i := range rows {
		rows[i].Fields = rows[i].Fields[:8]
	}
	return rows
}

package robno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Opšti dokumenti (the groups OPD, POP and POT, the legacy rob_opsti_dokumenti_opd): a robni dokument with
// the protivstav konto and the org. jedinice and mesta troška of the document and of the robe; the document
// of the vrsta, its stavke are added and deleted (the legacy has no change of a stavka). The groups FIN (the
// legacy rob_opsti_dokumenti_FIN: the stavke with the iznos) and KOL (rob_opsti_dokumenti_KOL: the stavke
// with the količina) have the org. jedinica and the mesto troška of the document only; MNV
// (rob_opsti_dokumenti_MNV: the stavke of a nivelacija by the naziv, without the artikal) has none.
// ---------------------------------------------------------------------------------------------------

// robOpstiServis tells whether the opis of a nalog has the word "servis" (the legacy bServis: the cena of a
// stavka of a magacin of the nabavna cena is entered).
var robOpstiServis = regexp.MustCompile(`(?i)\bservis\b`)

// robOpstiTipArtikla are the tipovi of the artikli of a magacin by its tip (the legacy tipART of
// rob_opsti_dokumenti_opd) with the opis of the magacin (sMAGOP).
func robOpstiTipArtikla(tipmag string) (tipovi []string, opis string) {
	switch strings.ToUpper(strings.TrimSpace(tipmag)) {
	case "V":
		return []string{"R"}, "magacinu veleprodaje"
	case "L":
		return []string{"L"}, "magacinu poluproizvoda"
	case "S":
		return []string{"S", "A"}, "magacinu sirovina"
	case "G":
		return []string{"P"}, "magacinu gotovih proizvoda"
	case "M":
		return []string{"R"}, "magacinu maloprodaje"
	case "D":
		return []string{"D"}, "magacinu rezervnih delova"
	}
	return nil, ""
}

// robOpstiCenaNaziv is the label of the cena of the stavke by the tip of the magacin (the legacy caption of
// EDT_CENA): the magacini of the poluproizvodi, the sirovine, the gotovi proizvodi and the rezervni delovi
// have the magacinska cena, the others the veleprodajna one.
func robOpstiCenaNaziv(tipmag string) string {
	switch strings.ToUpper(strings.TrimSpace(tipmag)) {
	case "L", "S", "G", "D":
		return "Magacinska cena"
	}
	return "Veleprodajna cena"
}

// GetRobOpstiDokument returns the document of the "Opšti dokumenti" screen with the broj the user selected or
// typed, like the legacy exit of EDT_DOKUM of rob_opsti_dokumenti_opd: the document of the faktura
// (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one with the message
// why; a new one with the datum of the nalog) with the protivstav konto and šifra with its naziv, the org.
// jedinica and the mesto troška where the roba goes and the opis of the vrsta; a new document has the first
// org. jedinica with its first mesto troška for both (the legacy PuniCOMBO) and no datum izvornog dokumenta.
// The entry of a stavka has the first konto of the magacin, its tip zaliha and the label and the state of
// the cena.
func (s *RobnoDokumentaResource) GetRobOpstiDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoOpstiDokument, error) {
	res := domain.RobnoOpstiDokument{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return res, fmt.Errorf("no user session found")
	}
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return res, err
	}
	res.Header.RobnoFaktureHeaderView, res.Poruka = dok.Header, dok.Poruka
	h := &res.Header
	vrd := common.StringToInt(h.Vrd)
	grupa, err := s.robInterniGrupa(ctx, vrd)
	if err != nil {
		return res, err
	}
	h.Fin, res.Stavka.Fin = grupa == "FIN", grupa == "FIN"
	h.Kol, res.Stavka.Kol = grupa == "KOL", grupa == "KOL"
	h.Mnv, res.Stavka.Mnv = grupa == "MNV", grupa == "MNV"
	h.Niv, res.Stavka.Niv = grupa == "NIV", grupa == "NIV"
	// A prenosnica (the group PRI, the legacy rob_Prenosnice); its kontra dokument (PRU) is only viewed.
	h.Pre, res.Stavka.Pre = grupa == "PRI", grupa == "PRI"
	h.SamoPregled = grupa == "PRU"
	// A zaduženje gradilišta (the group GRD, the legacy rob_ZadauzGRADILISTA).
	// A trebovanje (the group OTR, the legacy rob_Trebovanja) is a zaduženje gradilišta with the konto of the trošak.
	h.Grd, res.Stavka.Grd = grupa == "GRD" || grupa == "OTR", grupa == "GRD" || grupa == "OTR"
	h.Otr, res.Stavka.Otr = grupa == "OTR", grupa == "OTR"
	// A popis u toku godine (the group PTG, the legacy rob_opsti_dokPOPIS).
	h.Ptg, res.Stavka.Ptg = grupa == "PTG", grupa == "PTG"
	// A zaduženje sitnog inventara (the group SIV, the legacy rob_ZaduzSitanInventar).
	h.Siv, res.Stavka.Siv = grupa == "SIV", grupa == "SIV"
	if opis, err := s.robno.DokvrstaOpis(ctx, vrd); err == nil {
		h.OpisVrste = strings.TrimSpace(opis)
	}

	if h.RdokID > 0 {
		rows, err := s.robno.RdokProtivstav(ctx, h.RdokID)
		if err != nil {
			return res, err
		}
		if rows != nil && len(*rows) > 0 {
			r := (*rows)[0]
			h.Tkonto, h.Tsifra = r.Key, r.C1
			h.Orgjed1, h.Orgjed1Text = robNullID(r.C2), r.C3
			h.MestoTroska1, h.MestoTroska1Text = robNullID(r.C4), r.C5
		}
		if h.Tkonto != "" {
			if h.TkontoNaziv, err = s.robOpstiNazivKonta(ctx, h.Tkonto, h.Tsifra); err != nil {
				return res, err
			}
		}
	} else {
		// A new document has no datum izvornog dokumenta (the legacy NoviDOK).
		h.IzvorniDokDatum = robNullTime(sql.NullTime{})
		// The org. jedinica and the mesto troška of the document: none ("-", the user chooses them); the ones of the
		// roba: the first org. jedinica with its first mesto troška.
		rows, err := s.robno.OrgjedPrvaSaMestom(ctx)
		if err == nil && rows != nil && len(*rows) > 0 {
			r := (*rows)[0]
			h.Orgjed1, h.Orgjed1Text, h.MestoTroska1, h.MestoTroska1Text = robNullID(r.Key), r.C1, robNullID(r.C2), r.C3
		}
	}

	// The entry of a stavka: the first konto of the magacin, the tip zaliha, the cena.
	st := &res.Stavka
	st.RdokID = h.RdokID
	if st.Konto, err = s.robUslugeKontoMagacina(ctx, h.MagaciniID); err != nil {
		return res, err
	}
	rows, err := s.robno.MagacinOpsti(ctx, h.MagaciniID, h.RnalID)
	if err != nil {
		return res, err
	}
	if rows != nil && len(*rows) > 0 {
		r := (*rows)[0]
		st.Tipzal = common.StringToInt(r.Key)
		st.NabCena = common.StringToInt(r.C1) == 3
		st.CenaZakljucana = st.NabCena && !robOpstiServis.MatchString(r.C3)
		st.CenaNaziv = robOpstiCenaNaziv(r.C2)
		// A nivelacija veleprodaje: the magacin keeps its zalihe at the prodajna cena (the legacy PodaciZAGLAVLJE).
		if h.Niv && common.StringToInt(r.C1) != 1 {
			res.Poruka = strings.TrimSpace(res.Poruka + " " + robNivMsgProdajnaCena)
		}
	}
	if h.Pre {
		if err := s.robPreDokument(ctx, &res); err != nil {
			return res, err
		}
	}
	if h.Grd {
		if err := s.robGrdDokument(ctx, &res); err != nil {
			return res, err
		}
	}
	if h.Ptg {
		if err := s.robPtgDokument(ctx, &res); err != nil {
			return res, err
		}
	}
	if h.Siv {
		if err := s.robSivDokument(ctx, &res); err != nil {
			return res, err
		}
	}
	return res, nil
}

// robOpstiNazivKonta is the naziv of the protivstav konto (the konto of the kontni plan, fkpl.vkonta 2) or of
// its analitička šifra (vkonta 1), "Nepostojeći konto!" without it (the legacy exit of EDT_TKONTO and
// EDT_TSIFRA).
func (s *RobnoDokumentaResource) robOpstiNazivKonta(ctx context.Context, konto, sifra string) (string, error) {
	naziv, found, err := s.robno.FkplNaziv(ctx, konto, sifra)
	if err != nil {
		return "", err
	}
	if !found {
		return "Nepostojeći konto!", nil
	}
	return naziv, nil
}

// ValidateRobOpsti validates the header of an opšti dokument before it is saved, like the legacy
// ValidacijaRDOK of rob_opsti_dokumenti_opd: the datum dokumenta is required and not after the datum naloga;
// the org. jedinica and the mesto troška of the document and of the roba are not required (empty or "-" are
// saved as NULL), a given one exists (robOpstiOrgjedMt); a given protivstav konto exists in the kontni plan
// (fkpl.vkonta 2); the broj dokumenta is a whole number.
func (s *RobnoDokumentaResource) ValidateRobOpsti(ctx context.Context, params domain.RobnoOpstiParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("brdok", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	grupa := s.robOpstiGrupa(ctx, params.Vrd)
	if grupa == "NIV" {
		return s.robNivValidate(ctx, params)
	}
	if grupa == "PRI" || grupa == "PRU" {
		return s.robPreValidate(ctx, params, grupa)
	}
	if grupa == "GRD" || grupa == "OTR" {
		return s.robGrdValidate(ctx, params)
	}
	if grupa == "SIV" {
		return s.robSivValidate(ctx, params)
	}
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}
	// The nivelacija (the legacy ValidacijaRDOK of rob_opsti_dokumenti_MNV): the datum only.
	if grupa == "MNV" {
		s.robOpstiBrojDokumenta(params.BrojDokum, add)
		return fieldErrors
	}
	// The org. jedinica and the mesto troška are not required (empty or "-": NULL).
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)
	if grupa == "FIN" || grupa == "KOL" {
		// A finansijski or a količinski opšti dokument (the legacy ValidacijaRDOK of rob_opsti_dokumenti_FIN and
		// rob_opsti_dokumenti_KOL): no protivstav konto, org. jedinica and mesto troška of the roba.
		s.robOpstiBrojDokumenta(params.BrojDokum, add)
		return fieldErrors
	}
	s.robOpstiOrgjedMt(ctx, params.Orgjed1, params.MestoTroska1, "idorgjed1", "mestotrid1", add)
	if tkonto := strings.TrimSpace(params.Tkonto); tkonto != "" {
		if exists, err := s.robno.FkplSintetikaPostoji(ctx, tkonto); err != nil {
			add("tkonto", common.ErrMsgReadData)
		} else if !exists {
			add("tkonto", robnoUnosMsgNePostoji)
		}
	}
	s.robOpstiBrojDokumenta(params.BrojDokum, add)
	return fieldErrors
}

// robOpstiOrgjedMt checks an org. jedinica and its mesto troška of an opšti dokument (the fields ojField and
// mtField): both are not required (empty or "-": saved as NULL); a given one exists and a given mesto troška is
// one of the given org. jedinica (the search of the mesta troška gives only the ones of the org. jedinica).
func (s *RobnoDokumentaResource) robOpstiOrgjedMt(ctx context.Context, oj, mt, ojField, mtField string, add func(field, message string)) {
	ojID, ojGiven, ojOK := robnoUnosID(oj)
	if !ojOK {
		add(ojField, robnoUnosMsgNePostoji)
	} else if ojGiven {
		if exists, err := s.robno.OrgjedPostoji(ctx, ojID); err != nil {
			add(ojField, common.ErrMsgReadData)
			ojOK = false
		} else if !exists {
			add(ojField, robnoUnosMsgNePostoji)
			ojOK = false
		}
	}
	mtID, mtGiven, mtOK := robnoUnosID(mt)
	switch {
	case !mtOK:
		add(mtField, robnoUnosMsgNePostoji)
	case !mtGiven:
	default:
		if exists, err := s.robno.MestotrPostoji(ctx, mtID); err != nil {
			add(mtField, common.ErrMsgReadData)
		} else if !exists {
			add(mtField, robnoUnosMsgNePostoji)
		} else if ojOK && ojGiven {
			if svoj, err := s.robno.MestotrUOrgjed(ctx, mtID, ojID); err != nil {
				add(mtField, common.ErrMsgReadData)
			} else if !svoj {
				add(mtField, "Nije mesto troška izabrane OJ.")
			}
		}
	}
}

// robOpstiBrojDokumenta checks the broj of an opšti dokument: required, a whole number.
func (s *RobnoDokumentaResource) robOpstiBrojDokumenta(brojDokum string, add func(field, message string)) {
	switch broj := strings.TrimSpace(brojDokum); {
	case broj == "":
		add("brdok", common.ErrMsgObavezanPodatak)
	default:
		if _, err := strconv.Atoi(broj); err != nil {
			add("brdok", robnoUnosMsgCeoBroj)
		}
	}
}

// robOpstiGrupa is the group of the vrsta of an opšti dokument (OPD, POP, POT, FIN, KOL, MNV), "" when it
// cannot be read.
func (s *RobnoDokumentaResource) robOpstiGrupa(ctx context.Context, vrd string) string {
	grupa, err := s.robInterniGrupa(ctx, common.StringToInt(vrd))
	if err != nil {
		return ""
	}
	return grupa
}

// SaveRobOpsti saves the header of an opšti dokument (the legacy SnimiRDOK of rob_opsti_dokumenti_opd): a new
// document of the nalog (SaveRobniDokument: with the columns of the nalog, counted in it) or the existing one
// of the vrsta and broj, with the datum, the opis, the izvorni dokument, the protivstav konto and šifra and the
// org. jedinice and mesta troška of the document and of the roba; a new one has no rok, kupac or napomena o
// poreskom oslobođenju. It returns the rdokid of the document.
func (s *RobnoDokumentaResource) SaveRobOpsti(ctx context.Context, params domain.RobnoOpstiParams) (int64, error) {
	switch s.robOpstiGrupa(ctx, params.Vrd) {
	case "NIV":
		return s.robNivSaveHeader(ctx, params)
	case "PRI":
		return s.robPreSaveHeader(ctx, params)
	case "GRD", "OTR":
		return s.robGrdSaveHeader(ctx, params, s.robOpstiGrupa(ctx, params.Vrd) == "OTR")
	case "PTG":
		// A popis u toku godine (the legacy SnimiRDOK of rob_opsti_dokPOPIS): the header of the trebovanje (a saved
		// one gets all its fields) with the radnik as the konto and the šifra.
		return s.robGrdSaveHeader(ctx, params, true)
	case "SIV":
		return s.robSivSaveHeader(ctx, params)
	}
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	dok := domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        common.StringToInt(params.Vrd),
		Dokum:      common.StringToInt(params.BrojDokum),
		Polja: []domain.Fields{
			robPoljeDatum("dadok", params.DatumDokum),
			robPoljeText("opis", params.Opis),
			robPoljeDatum("datiz", params.IzvorniDokDatum),
			robPoljeText("dokiz", params.IzvorniDokum),
			{Name: "rok", Value: "0"},
			{Name: "fkto", Value: ""},
			{Name: "fana", Value: ""},
			{Name: "pornapomena", Value: ""},
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
			robPoljeID("idorgjed1", id(params.Orgjed1)),
			robPoljeID("mestotrid1", id(params.MestoTroska1)),
			robPoljeText("tkonto", params.Tkonto),
			robPoljeText("tsifra", params.Tsifra),
		},
	}
	// A finansijski or a količinski opšti dokument (the legacy SnimiRDOK of rob_opsti_dokumenti_FIN and
	// rob_opsti_dokumenti_KOL) has the org. jedinica and the mesto troška of the document only, a nivelacija
	// (rob_opsti_dokumenti_MNV) none of them.
	var bez []string
	switch s.robOpstiGrupa(ctx, params.Vrd) {
	case "FIN", "KOL":
		bez = []string{"idorgjed1", "mestotrid1", "tkonto", "tsifra"}
	case "MNV":
		bez = []string{"idorgjed", "mestotrid", "idorgjed1", "mestotrid1", "tkonto", "tsifra"}
	}
	if len(bez) > 0 {
		polja := dok.Polja[:0]
		for _, p := range dok.Polja {
			if !robSadrzi(bez, p.Name) {
				polja = append(polja, p)
			}
		}
		dok.Polja = polja
	}
	return s.SaveRobniDokument(ctx, dok)
}

// robOpstiRsta are the stanje and the cene of an artikal in a konto of the magacin (rsta).
type robOpstiRsta struct {
	found                                      bool
	ulaz, izlaz, dug, pot, cena, ncena, prosnc float64
}

// robOpstiStanje reads the stanje and the cene of the artikal in the konto of the magacin (rsta).
func (s *RobnoDokumentaResource) robOpstiStanje(ctx context.Context, magaciniID int, konto string, sifra int) (robOpstiRsta, error) {
	st, found, err := s.robno.RstaStanjeOpsti(ctx, magaciniID, konto, sifra)
	if err != nil || !found {
		return robOpstiRsta{}, err
	}
	return robOpstiRsta{found: true, ulaz: st.Ulaz, izlaz: st.Izlaz, dug: st.Dug, pot: st.Pot, cena: st.Cena, ncena: st.Ncena, prosnc: st.Prosnc}, nil
}

// cene are the cena and the nabavna cena of a stavka of an opšti dokument from the stanje (the legacy exit of
// EDT_SIFRA): in a magacin of the nabavna cena both the prosečna nabavna cena; else, for a vrsta of the
// prosečne cene (its opis has "PROS"), the vrednost of the stanje by its količina (when both are positive),
// else the cena of the stanje, and the nabavna cena of the stanje.
func (r robOpstiRsta) cene(nabavnaCena, prosecna bool) (cena, ncena float64) {
	if !r.found {
		return 0, 0
	}
	if nabavnaCena {
		return r.prosnc, r.prosnc
	}
	cena = r.cena
	if prosecna && r.ulaz-r.izlaz > 0 && r.dug-r.pot > 0 {
		cena = robZaokruzi((r.dug-r.pot)/(r.ulaz-r.izlaz), 2)
	}
	return cena, r.ncena
}

// robOpstiDrsta reads the stanje of the artikal by lot (drsta: the konto, the šifra, the magacin, the OTK broj,
// the serija and the rok); found is false without it.
func (s *RobnoDokumentaResource) robOpstiDrsta(ctx context.Context, mag int, konto string, sifra int, otk, serija string, roktr int64) (float64, bool, error) {
	return s.robno.DrstaLotKolicina(ctx, mag, konto, sifra, otk, serija, roktr)
}

// robOpstiProsecna tells whether the vrsta of the document is one of the prosečne cene (the legacy ProsSTA: its
// opis has "PROS").
func (s *RobnoDokumentaResource) robOpstiProsecna(ctx context.Context, vrd int) bool {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return false
	}
	ok, err := s.robno.DokvrstaProsecna(ctx, vrd)
	return err == nil && ok
}

// GetRobOpstiArtikal returns the artikal of a stavka of an opšti dokument with what the form of the stavke
// shows of it, like the legacy exit of EDT_SIFRA of rob_opsti_dokumenti_opd: the artikal by its šifra, or by
// its barkod for a value longer than 6 characters, with its naziv, JM and konto (the konto of the artikal,
// else the konto of the form); its cene (see cene) and its stanje in the konto. In a magacin of the tip zaliha
// 2 the artikal must exist and be of a tip of the magacin (else Poruka), and its stanje is the one of the lot
// of the form (the OTK broj, the serija and the rok), without it the cene are 0. found is false for the šifra
// 0 and an artikal that does not exist.
func (s *RobnoDokumentaResource) GetRobOpstiArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto, otk, serija, rok string) (domain.RobOpstiArtikal, bool, error) {
	a := domain.RobOpstiArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	if vrednost == "" || vrednost == "0" {
		return a, false, nil
	}
	// A nivelacija veleprodaje (the legacy exit of EDT_SIFRA of rob_Nivelacije_Veleprodaja).
	switch grupa, _ := s.robInterniGrupa(ctx, k.Vrd); grupa {
	case "NIV":
		return s.robNivArtikal(ctx, k, vrednost, konto)
	case "PRI":
		// A prenosnica (the legacy exit of EDT_SIFRA of rob_Prenosnice).
		return s.robPreArtikal(ctx, k, vrednost, konto, otk, serija, rok)
	case "GRD", "OTR":
		// A zaduženje gradilišta, a trebovanje (the legacy exits of EDT_SIFRA of rob_ZadauzGRADILISTA and rob_Trebovanja).
		return s.robGrdArtikal(ctx, k, vrednost, konto)
	case "PTG":
		// A popis u toku godine (the legacy exit of EDT_SIFRA of rob_opsti_dokPOPIS).
		return s.robPtgArtikal(ctx, k, vrednost, konto, otk, serija, rok)
	case "SIV":
		// A zaduženje sitnog inventara (the legacy exit of EDT_SIFRA of rob_ZaduzSitanInventar).
		return s.robSivArtikal(ctx, k, vrednost, konto, otk, serija, rok)
	}
	rows, err := s.robno.RsifArtikalOpsti(ctx, len(vrednost) > 6, vrednost)
	if err != nil {
		return a, false, err
	}
	if rows == nil || len(*rows) == 0 {
		if k.Tipzal == 2 {
			a.Poruka = "Artikal sa izabranom šifrom ne postoji."
		}
		return a, false, nil
	}
	r := (*rows)[0]
	grupa, _ := s.robInterniGrupa(ctx, k.Vrd)
	if grupa == "FIN" {
		a.Sifra, a.Naziv, a.Jm, a.Konto = common.StringToInt(r.Key), r.C1, r.C2, strings.TrimSpace(konto)
		st, err := s.robOpstiStanje(ctx, k.MagaciniID, a.Konto, a.Sifra)
		if err != nil {
			return a, true, err
		}
		a.Stanje = st.ulaz - st.izlaz
		return a, true, nil
	}
	if k.Tipzal == 2 {
		tipovi, opis := robOpstiTipArtikla(k.Tipmag)
		if !robSadrzi(tipovi, r.C4) {
			a.Poruka = "Izabrani tip artikla ne pripada " + opis + "."
			return a, false, nil
		}
	}
	a.Sifra, a.Naziv, a.Jm = common.StringToInt(r.Key), r.C1, r.C2
	a.Konto = strings.TrimSpace(konto)
	if r.C3 != "" {
		a.Konto = r.C3
	}
	st, err := s.robOpstiStanje(ctx, k.MagaciniID, a.Konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	switch k.Tipzal {
	case 2:
		datum, _ := robnoUnosDatum(rok)
		stanje, lot, err := s.robOpstiDrsta(ctx, k.Mag, a.Konto, a.Sifra, otk, serija, RobWinDevDatum(datum))
		if err != nil {
			return a, true, err
		}
		if lot {
			a.Cena, a.Ncena = st.cene(k.NabavnaCena(), s.robOpstiProsecna(ctx, k.Vrd))
			a.Stanje = stanje
		}
	default:
		a.Cena, a.Ncena = st.cene(k.NabavnaCena(), s.robOpstiProsecna(ctx, k.Vrd))
		a.Stanje = st.ulaz - st.izlaz
	}
	return a, true, nil
}

// ValidateRobOpstiStavka validates a stavka of an opšti dokument before it is saved, like the legacy Validacija
// of rob_opsti_dokumenti_opd: the konto is required and a konto of the magacin; the šifra is required and the
// artikal exists; the količina and the cena are greater than 0; a document posted on the potražuje (an izlaz)
// that is not a storno has the količina in the stanje (of the lot in a magacin of the tip zaliha 2): the
// legacy question "Nema dovoljno zaliha. Da li ste sigurni da izdajete robu?" is the error of the field. The
// question of the OTK broj ("Unos bez OTK broja?") is not asked.
func (s *RobnoDokumentaResource) ValidateRobOpstiStavka(ctx context.Context, params domain.RobOpstiStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("sifraartikla", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, params.RdokID)
	if err != nil || !found {
		add("sifraartikla", common.ErrMsgReadData)
		return fieldErrors
	}
	// A nivelacija (the legacy Validacija of rob_opsti_dokumenti_MNV): the naziv is required, the količina and
	// the nova cena are greater than 0.
	grupa, _ := s.robInterniGrupa(ctx, k.Vrd)
	if grupa == "NIV" {
		return s.robNivValidateStavka(ctx, k, params)
	}
	if grupa == "PRI" {
		return s.robPreValidateStavka(ctx, k, params)
	}
	if grupa == "GRD" || grupa == "OTR" {
		return s.robGrdValidateStavka(ctx, k, params, grupa == "OTR")
	}
	if grupa == "PTG" || grupa == "SIV" {
		// A popis u toku godine and a zaduženje sitnog inventara (the legacy Validacija of rob_opsti_dokPOPIS and of
		// rob_ZaduzSitanInventar: the same checks).
		return s.robPtgValidateStavka(ctx, k, params)
	}
	if grupa == "MNV" {
		if strings.TrimSpace(params.Naziv) == "" {
			add("nazivmnv", common.ErrMsgObavezanPodatak)
		}
		if strings.TrimSpace(params.Kolic) == "" {
			add("kolicina", common.ErrMsgObavezanPodatak)
		} else if common.StringToFloat64(params.Kolic) <= 0 {
			add("kolicina", robInterniMsgVeceOdNule)
		}
		if common.StringToFloat64(params.NovaCena) <= 0 {
			add("novacena", robInterniMsgVeceOdNule)
		}
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
	} else if sifra, err = strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.RsifPostoji(ctx, sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else {
		sifraOK = true
	}
	// A finansijski opšti dokument (the legacy Validacija of rob_opsti_dokumenti_FIN): the iznos is greater than 0.
	if grupa == "FIN" {
		if common.StringToFloat64(params.Iznos) <= 0 {
			add("iznosfin", robInterniMsgVeceOdNule)
		}
		return fieldErrors
	}
	kolicina := common.StringToFloat64(params.Kolic)
	if strings.TrimSpace(params.Kolic) == "" {
		add("kolicina", common.ErrMsgObavezanPodatak)
	} else if kolicina <= 0 {
		add("kolicina", robInterniMsgVeceOdNule)
	}
	// A količinski opšti dokument (the legacy Validacija of rob_opsti_dokumenti_KOL): the količina only (its
	// check of the zalihe is commented out).
	if grupa == "KOL" {
		return fieldErrors
	}
	if common.StringToFloat64(params.Cena) <= 0 {
		add("cenaopd", robInterniMsgVeceOdNule)
	}
	// The zalihe of an izlaz.
	if k.Potrazno() && !k.Storno() && sifraOK && konto != "" && kolicina > 0 {
		var stanje float64
		if k.Tipzal == 2 {
			datum, _ := robnoUnosDatum(params.Rok)
			if stanje, _, err = s.robOpstiDrsta(ctx, k.Mag, konto, sifra, params.Otk, params.Serija, RobWinDevDatum(datum)); err != nil {
				add("kolicina", common.ErrMsgReadData)
				return fieldErrors
			}
		} else {
			st, err := s.robOpstiStanje(ctx, k.MagaciniID, konto, sifra)
			if err != nil {
				add("kolicina", common.ErrMsgReadData)
				return fieldErrors
			}
			stanje = st.ulaz - st.izlaz
		}
		if stanje <= 0 || robZaokruzi(stanje-kolicina, 3) < 0 {
			add("kolicina", "Nema dovoljno zaliha.")
		}
	}
	return fieldErrors
}

// SaveRobOpstiStavka adds a stavka to an opšti dokument (the legacy CreateRpro of rob_opsti_dokumenti_opd) in
// one transaction:
//   - the iznos is the količina by the cena, rounded to 2 decimals, negative (with the količina) for a storno;
//   - the stavka has the columns of its document (nalog, vrsta, broj, datum, magacin, kupac and prodavnica),
//     the artikal (šifra, naziv, tarifa, rsifid), the konto, the cena, the nabavna cena, the OTK broj, the serija
//     and the rok, the prosečna nabavna cena of the stanje (konprosnc), the next redni broj; rab and mma 0;
//   - a new artikal of the cene (rcene) gets the cena and the nabavna cena of the stavka;
//   - the nalog, the stanje of the artikal and the zalihe by lot get the iznos and the količina (AzurirajRNAL,
//     AzurirajRSTA and AzurirajDRSTA with the cena of the stavka; the SR, RSAL and RSIN are not updated);
//   - a document posted on the duguje (an ulaz) gives the stanje its new prosečna nabavna cena (the one of the
//     stanje before and of the stavke by their količine), also the zalihe by lot of the artikal in the magacin;
//   - the document gets the prodajna and the nabavna vrednost, the porez of the stavka (by the stopa of its
//     tarifa on the datum of the document), its iznos and a stavka more.
func (s *RobnoDokumentaResource) SaveRobOpstiStavka(ctx context.Context, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
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
	if grupa, err := s.robInterniGrupa(ctx, k.Vrd); err != nil {
		return 0, err
	} else if grupa == "FIN" {
		return s.robOpstiFinStavka(ctx, k, params)
	} else if grupa == "KOL" {
		return s.robOpstiKolStavka(ctx, k, params)
	} else if grupa == "MNV" {
		return s.robOpstiMnvStavka(ctx, k, params)
	} else if grupa == "NIV" {
		return s.robNivStavka(ctx, k, params)
	} else if grupa == "PRI" {
		return s.robPreStavka(ctx, k, params)
	} else if grupa == "GRD" || grupa == "OTR" {
		return s.robGrdStavka(ctx, k, params, grupa == "OTR")
	} else if grupa == "PTG" {
		return s.robPtgStavka(ctx, k, params)
	} else if grupa == "SIV" {
		return s.robSivStavka(ctx, k, params)
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	kolicinaForme := common.StringToFloat64(params.Kolic)
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	kolicina, iznos := math.Abs(kolicinaForme), math.Abs(robZaokruzi(kolicinaForme*cena, 2))
	if k.Storno() {
		kolicina, iznos = -kolicina, -iznos
	}
	rok, _ := robnoUnosDatum(params.Rok)
	st, err := s.robOpstiStanje(ctx, k.MagaciniID, konto, sifra)
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
	rsifID, po, naziv, err := s.robno.RsifIDPoNaziv(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	pkto, pana, pdvpct, err := s.robno.RdokPktoPanaStopa(ctx, tx, params.RdokID, po)
	if err != nil {
		return 0, err
	}
	rbr, err := s.robno.RproSledeciRbrNaloga(ctx, tx, k.RnalID, params.RdokID)
	if err != nil {
		return 0, err
	}
	konprosnc := 0.0
	if st.found {
		konprosnc = st.prosnc
	}
	if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
		{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
		{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
		robPoljeNullTime("dadok", k.Dadok),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "naz1", Value: naziv},
		{Name: "konto", Value: konto},
		robPoljeFloat("cena", cena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		robPoljeFloat("ncena", ncena),
		{Name: "rab", Value: "0"},
		{Name: "po", Value: strconv.Itoa(po)},
		{Name: "fana", Value: k.Fana},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		{Name: "brst", Value: strconv.Itoa(rbr)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)},
		robPoljeFloat("konprosnc", konprosnc),
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	}); err != nil {
		return 0, err
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
			robPoljeFloat("cena", cena), {Name: "fcena", Value: "0"}, {Name: "mcena", Value: "0"}, {Name: "mcenap", Value: "0"},
			robPoljeFloat("ncena", ncena), {Name: "ztrin", Value: "0"}, {Name: "rab", Value: "0"}, {Name: "tz", Value: ""}, {Name: "tzi", Value: ""},
		}); err != nil {
			return 0, err
		}
	}

	// The nalog, the stanje and the zalihe by lot.
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos); err != nil {
		return 0, err
	}
	if err = s.robAzurirajStanjeStavke(ctx, tx, k, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra,
		Kolicina: kolicina, Iznos: iznos, Cena: cena, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Otk: strings.TrimSpace(params.Otk), Serija: strings.TrimSpace(params.Serija), Roktr: RobWinDevDatum(rok), Brst: 1}); err != nil {
		return 0, err
	}
	// An ulaz: the prosečna nabavna cena of the stanje (and of the zalihe by lot of the artikal in the magacin).
	if k.Dugovno() {
		stanje := st.ulaz - st.izlaz
		prosnc := ncena
		if kolicina+stanje != 0 {
			prosnc = robZaokruzi((kolicina*ncena+stanje*st.prosnc)/(kolicina+stanje), 2)
		}
		if err = s.robno.RstaPostaviProsnc(ctx, tx, k.MagaciniID, konto, sifra, prosnc); err != nil {
			return 0, err
		}
		if err = s.robno.DrstaPostaviProsnc(ctx, tx, k.Mag, rsifID, prosnc); err != nil {
			return 0, err
		}
	}

	// The document: the prodajna and the nabavna vrednost, the porez, the iznos, a stavka more.
	if err = s.robno.RdokDodajVrednosti(ctx, tx, params.RdokID, kolicinaForme*cena, kolicinaForme*ncena, kolicina*cena*pdvpct/100, iznos); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// DeleteRobOpstiStavka deletes a stavka of an opšti dokument (the legacy BrisiRPRO of rob_opsti_dokumenti_opd)
// in one transaction and returns its document: the stavka is taken back from its nalog, from the stanje and
// from the zalihe by lot, it is deleted, the document loses its iznos and a stavka, and the zalihe by lot of
// the stavka without the stanje and without another stavka are deleted. The redni brojevi of the other stavke
// stay (the legacy does not number them again). After the deletion of a stavka of an ulaz the prosečna nabavna
// cena is computed again (IzracunajNCENA; not of a KOL: the legacy BrisiRPRO of rob_opsti_dokumenti_KOL does
// not). A FIN changes the stanje only (no zalihe by lot).
func (s *RobnoDokumentaResource) DeleteRobOpstiStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
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
	st, err := s.robno.RproUslugaBrisanje(ctx, tx, rproID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("stavka ne postoji")
		}
		return 0, err
	}
	rdokID = st.RdokID
	sifra, konto, otk, serija, kolicina, iznos, roktr := st.Sifra, st.Konto, st.Otk, st.Serija, st.Kolicina, st.Iznos, st.Roktr
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("dokument ne postoji")
	}
	grupa, err := s.robInterniGrupa(ctx, k.Vrd)
	if err != nil {
		return 0, err
	}
	fin, kol := grupa == "FIN", grupa == "KOL"
	if grupa == "NIV" {
		if err = s.robNivBrisi(ctx, tx, k, rproID, konto, sifra, iznos, otk, serija, int64(roktr)); err != nil {
			return 0, err
		}
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return rdokID, nil
	}
	if grupa == "PRI" || grupa == "GRD" || grupa == "OTR" || grupa == "PTG" || grupa == "SIV" {
		switch grupa {
		case "PRI":
			err = s.robPreBrisi(ctx, tx, k, rproID)
		case "PTG":
			err = s.robPtgBrisi(ctx, tx, k, rproID)
		case "SIV":
			err = s.robSivBrisi(ctx, tx, k, rproID)
		default:
			err = s.robGrdBrisi(ctx, tx, k, rproID, grupa == "OTR")
		}
		if err != nil {
			return 0, err
		}
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return rdokID, nil
	}
	if grupa == "MNV" {
		if err = s.robOpstiMnvBrisi(ctx, tx, k, rproID, kolicina, iznos); err != nil {
			return 0, err
		}
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return rdokID, nil
	}
	promena := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra,
		Kolicina: -kolicina, Iznos: -iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Otk: strings.TrimSpace(otk), Serija: strings.TrimSpace(serija), Roktr: int64(roktr), Brst: -1}
	if fin {
		// A finansijski opšti dokument (the legacy BrisiRPRO of rob_opsti_dokumenti_FIN): the stanje only.
		err = s.AzurirajRSTA(ctx, tx, promena)
	} else {
		err = s.robAzurirajStanjeStavke(ctx, tx, k, promena)
	}
	if err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return 0, err
	}
	// The nalog after the deletion (AzurirajRNAL counts the stavke of the nalog).
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos); err != nil {
		return 0, err
	}
	if err = s.robno.RdokOduzmiStavku(ctx, tx, rdokID, iznos); err != nil {
		return 0, err
	}
	if fin {
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return rdokID, nil
	}
	// The zalihe by lot without the stanje and without another stavka.
	if err = s.robno.DrstaBrisiPrazanLot(ctx, tx, konto, sifra, k.Mag, otk, serija, int64(roktr)); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	// An ulaz: the stanje of the artikal and its zalihe by lot in the magacin get the prosečna nabavna cena again
	// (the legacy IzracunajNCENA on the datum of the nalog and of the document; after the deletion of the stavka).
	if k.Dugovno() && !kol {
		if err = s.robOpstiProsecnaCenaPonovo(ctx, k, konto, sifra); err != nil {
			return rdokID, err
		}
	}
	return rdokID, nil
}

// robOpstiProsecnaCenaPonovo gives the stanje of the artikal in the konto of the magacin (rsta.prosnc, rounded
// to 2 decimals) and its zalihe by lot in the magacin (drsta.prosnc) the prosečna nabavna cena computed again
// (IzracunajNCENA on the datum of the nalog and of the document of the kontekst).
func (s *RobnoDokumentaResource) robOpstiProsecnaCenaPonovo(ctx context.Context, k domain.RobStavkeKontekst, konto string, sifra int) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	ncena, err := s.IzracunajNCENA(ctx, robNullTime(k.Danal), robNullTime(k.Dadok), k.MagaciniID, sifra)
	if err != nil {
		return err
	}
	if err := s.robno.RstaPostaviProsnc(ctx, nil, k.MagaciniID, strings.TrimSpace(konto), sifra, robZaokruzi(ncena, 2)); err != nil {
		return err
	}
	return s.robno.DrstaPostaviProsncSifre(ctx, k.Mag, sifra, ncena)
}

// robOpstiFinStavka adds a stavka to a finansijski opšti dokument (the legacy CreateRpro of
// rob_opsti_dokumenti_FIN) in one transaction: the stavka has the iznos (negative for a storno) without the
// količina and the cena, the columns of its document, the artikal, the konto and the next redni broj; the
// nalog and the stanje of the artikal get the iznos (AzurirajRNAL, AzurirajRSTA; no zalihe by lot), the
// document gets the iznos and a stavka more.
func (s *RobnoDokumentaResource) robOpstiFinStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	iznos := math.Abs(common.StringToFloat64(params.Iznos))
	if k.Storno() {
		iznos = -iznos
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
	rsifID, po, found, err := s.robno.RsifIDPo(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("artikal ne postoji")
	}
	pkto, pana, err := s.robno.RdokPktoPanaTx(ctx, tx, params.RdokID)
	if err != nil {
		return 0, err
	}
	rbr, err := s.robno.RproSledeciRbr(ctx, tx, params.RdokID)
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
		{Name: "cena", Value: "0"},
		{Name: "kolic", Value: "0"},
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		{Name: "rab", Value: "0"},
		{Name: "po", Value: strconv.Itoa(po)},
		{Name: "fana", Value: k.Fana},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		{Name: "brst", Value: strconv.Itoa(rbr)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	}); err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos); err != nil {
		return 0, err
	}
	if err = s.AzurirajRSTA(ctx, tx, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra, Iznos: iznos,
		Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Brst: 1}); err != nil {
		return 0, err
	}
	if err = s.robno.RdokDodajIznos(ctx, tx, params.RdokID, iznos, 1); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robOpstiKolStavka adds a stavka to a količinski opšti dokument (the legacy CreateRpro of
// rob_opsti_dokumenti_KOL) in one transaction: the stavka has the količina (negative for a storno) without the
// cena and the iznos, the columns of its document, the artikal, the konto, the OTK broj, the serija and the rok
// and the next redni broj; the stanje of the artikal and its zalihe by lot get the količina (AzurirajRSTA and
// AzurirajDRSTA: the tip zaliha 1 without the lot, the tip zaliha 2 by the lot of the stavka), the nalog and the
// document a stavka more (the iznos 0).
func (s *RobnoDokumentaResource) robOpstiKolStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	kolicina := math.Abs(common.StringToFloat64(params.Kolic))
	if k.Storno() {
		kolicina = -kolicina
	}
	rok, _ := robnoUnosDatum(params.Rok)
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
	pkto, pana, err := s.robno.RdokPktoPanaTx(ctx, tx, params.RdokID)
	if err != nil {
		return 0, err
	}
	rbr, err := s.robno.RproSledeciRbr(ctx, tx, params.RdokID)
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
		{Name: "cena", Value: "0"},
		robPoljeFloat("kolic", kolicina),
		{Name: "iznos", Value: "0"},
		{Name: "fkto", Value: k.Fkto},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		{Name: "rab", Value: "0"},
		{Name: "po", Value: strconv.Itoa(po)},
		{Name: "fana", Value: k.Fana},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		{Name: "brst", Value: strconv.Itoa(rbr)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	}); err != nil {
		return 0, err
	}
	if err = s.robAzurirajStanjeStavke(ctx, tx, k, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: konto, Sifra: sifra,
		Kolicina: kolicina, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-",
		Otk: strings.TrimSpace(params.Otk), Serija: strings.TrimSpace(params.Serija), Roktr: RobWinDevDatum(rok), Brst: 1}); err != nil {
		return 0, err
	}
	// The nalog: a stavka more (the iznos of the stavka is 0).
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), 0); err != nil {
		return 0, err
	}
	if err = s.robno.RdokDodajStavku(ctx, tx, params.RdokID); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robOpstiMnvIznos is the efekat of the nivelacija of a stavka (the legacy exit of EDT_Cena of
// rob_opsti_dokumenti_MNV): the količina by the nova cena less the količina by the stara cena, rounded to 2
// decimals.
func robOpstiMnvIznos(kolicina, staraCena, novaCena float64) float64 {
	return robZaokruzi(kolicina*novaCena-kolicina*staraCena, 2)
}

// robOpstiMnvPredznak is the iznos of a stavka of a nivelacija as it is posted: negative for a storno (the
// legacy CreateRpro of rob_opsti_dokumenti_MNV keeps the stavka as entered and posts it with the predznak of
// the vrsta).
func robOpstiMnvPredznak(k domain.RobStavkeKontekst, iznos float64) float64 {
	if k.Storno() {
		return -iznos
	}
	return iznos
}

// robOpstiMnvStavka adds a stavka to a nivelacija (the legacy CreateRpro of rob_opsti_dokumenti_MNV) in one
// transaction: the stavka without the artikal (the šifra 0, no konto) with the naziv, the količina, the stara
// cena (fcena), the nova cena (cena) and the efekat of the nivelacija (iznos) as entered, the columns of its
// document and the next redni broj; rab and mma 0. It is posted by AzurirajStanja (the legacy
// GlobalProcROBNO.AZurirajStanja) with the šifra 0, no konto and the količina and the iznos negative for a storno.
func (s *RobnoDokumentaResource) robOpstiMnvStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	kolicina := common.StringToFloat64(params.Kolic)
	staraCena, novaCena := common.StringToFloat64(params.StaraCena), common.StringToFloat64(params.NovaCena)
	iznos := robOpstiMnvIznos(kolicina, staraCena, novaCena)
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	pkto, pana, err := s.robno.RdokPktoPanaTx(ctx, tx, params.RdokID)
	if err != nil {
		return 0, err
	}
	rbr, err := s.robno.RproSledeciRbr(ctx, tx, params.RdokID)
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
		{Name: "sifra", Value: "0"},
		{Name: "konto", Value: ""},
		robPoljeText("naz1", params.Naziv),
		robPoljeFloat("cena", novaCena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("fcena", staraCena),
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		{Name: "rab", Value: "0"},
		{Name: "fana", Value: k.Fana},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "brst", Value: strconv.Itoa(rbr)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	}); err != nil {
		return 0, err
	}
	if err = s.AzurirajStanja(ctx, tx, k.RnalID, k.RdokID, rproID, 0, "", robOpstiMnvPredznak(k, kolicina), robOpstiMnvPredznak(k, iznos)); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robOpstiMnvBrisi deletes a stavka of a nivelacija (the legacy BrisiRPRO of rob_opsti_dokumenti_MNV) in the
// transaction tx: its količina and iznos (as posted: negative for a storno) are taken back (BrisiStanja, the
// legacy GlobalProcROBNO.BrisiStanja), then the stavka is deleted.
func (s *RobnoDokumentaResource) robOpstiMnvBrisi(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64, kolicina, iznos float64) error {
	if err := s.BrisiStanja(ctx, tx, k.RnalID, k.RdokID, rproID, 0, "", robOpstiMnvPredznak(k, kolicina), robOpstiMnvPredznak(k, iznos)); err != nil {
		return err
	}
	return s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID)
}

// RobOpstiMnvStavkeTableRows returns the stavke of a nivelacija as the rows of the grid of the stavke (the
// columns of GetOpstiMnvStavkeTableFields: redni broj, naziv, količina, stara cena, nova cena, iznos).
func RobOpstiMnvStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Naziv, common.FormatNumberWithSystemLocale(st.Kolic, 3),
				robFormatIznos(st.Fcena), robFormatIznos(st.Cena), robFormatIznos(st.Iznos)},
		})
	}
	return rows
}

// RobOpstiKolStavkeTableRows returns the stavke of a količinski opšti dokument as the rows of the grid of the
// stavke (the columns of GetOpstiKolStavkeTableFields: redni broj, konto, šifra, naziv, JM, količina, OTK broj,
// serija, rok).
func RobOpstiKolStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rok := ""
		if st.Roktr > 0 {
			rok = robnoStampaWinDevDatumFormat(int64(st.Roktr), common.DateLayout)
		}
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), st.Otk, st.Serija, rok},
		})
	}
	return rows
}

// RobOpstiFinStavkeTableRows returns the stavke of a finansijski opšti dokument as the rows of the grid of the
// stavke (the columns of GetOpstiFinStavkeTableFields: redni broj, konto, šifra, naziv, JM, iznos).
func RobOpstiFinStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{
			ID:     strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm, robFormatIznos(st.Iznos)},
		})
	}
	return rows
}

// RobOpstiStavkeTableRows returns the stavke of an opšti dokument as the rows of the grid of the stavke of the
// screen (the columns of GetOpstiStavkeTableFields: redni broj, konto, šifra, naziv, JM, količina, cena,
// iznos, OTK broj, serija, rok).
func RobOpstiStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rok := ""
		if st.Roktr > 0 {
			rok = robnoStampaWinDevDatumFormat(int64(st.Roktr), common.DateLayout)
		}
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Cena), robFormatIznos(st.Iznos), st.Otk, st.Serija, rok},
		})
	}
	return rows
}

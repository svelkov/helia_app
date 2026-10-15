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
// Kalkulacije maloprodaje (the group KAL, the legacy rob_Kalkulacije_Maloprodaja with its stavke
// rob_Kalkulacije_Maloprodaja_Stavke): a robni dokument of a dobavljač for a prodavnica (rdok.pkto, rdok.pana, the
// konto of the stanje of its stavke) whose stavke enter the roba at the maloprodajna cena with the PDV; the porez paid
// to the dobavljač (rppo). The header, its validation and the plaćen porez are the ones of the kalkulacija
// veleprodaje without the valuta, the mesto nabavke, the avans and the zavisni troškovi of other dobavljači.
// ---------------------------------------------------------------------------------------------------

// robKmpMsgMaloprodaja is the message of a stavka with an artikal in a magacin that is not of the maloprodaja (the
// legacy exit of EDT_SIFRA).
const robKmpMsgMaloprodaja = "Maloprodajne kalkulacije sa šifrom morate raditi u magacinu maloprodaje."

// robKmpMaloprodaja tells whether the magacin is a magacin of the maloprodaja (magacini.tipmag "M").
func robKmpMaloprodaja(tipmag string) bool {
	return strings.EqualFold(strings.TrimSpace(tipmag), "M")
}

// RobKmpNabavnaCena is the nabavna cena of a stavka of a kalkulacija maloprodaje (the legacy exits of EDT_RAB and
// EDT_ZTRO): the fakturna cena less the rabat with its zavisni trošak (a procenat of it rounded to 2 decimals, or a
// vrednost); no external and no interni zavisni troškovi.
func RobKmpNabavnaCena(fcena, rab float64, tz string, ztro float64) float64 {
	return RobKalNabavnaCena(RobKalObracun{Fcena: fcena, Rab: rab, Tz: robKalOznaka(tz), Ztro: ztro, Tzi: robKalBezTroska})
}

// RobKmpStavkeTableRows returns the stavke of a kalkulacija maloprodaje as the rows of the grid of the stavke (the
// columns of GetKmpStavkeTableFields, the legacy TABLE_RPRO: redni broj, konto, šifra, opis artikla, JM, količina,
// fakturna cena, rabat, nabavna cena, zavisni trošak with its oznaka, MP marža, MP cena sa PDV), by the redni broj.
func RobKmpStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		trosak := ""
		if oznaka := robKalOznaka(st.Tz); oznaka != robKalBezTroska {
			trosak = common.FormatNumberWithSystemLocale(st.Ztro, 2) + " " + oznaka
		}
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Fcena), common.FormatNumberWithSystemLocale(st.Rab, 2),
				robFormatIznos(st.Ncena), trosak, common.FormatNumberWithSystemLocale(st.Mma, 2), robFormatIznos(st.Mcenap)},
		})
	}
	return rows
}

// ---------------------------------------------------------------------------------------------------
// The header (the legacy rob_Kalkulacije_Maloprodaja)
// ---------------------------------------------------------------------------------------------------

// GetRobKmpDokument returns the document of the "Kalkulacije maloprodaje" screen: the document of the params like the
// one of the faktura (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one with the
// next broj, NoviBrojRDOK, with the datum of the nalog and, for a PDV obveznik whose vrsta asks it, the POPDV polje of
// the vrsta) with the opis of its vrsta and whether the vrsta asks the POPDV polje (the legacy GR_POPDV), and
//   - a saved one: its prodavnica with its opis, its EPP polje, "PDV se može odbiti" and its plaćen porez;
//   - a new one: no dobavljač (the legacy has no default konto), the EPP and "PDV se može odbiti" checked (the legacy
//     initialization of CBOX_EPP and CBOX_ODBITNI), no izvorni dokument, the first org. jedinica with its first
//     mesto troška (the legacy COMBO_ORGJED).
//
// The entry of a stavka has the prodavnica of the document as its konto and whether the magacin is of the maloprodaja.
func (s *RobnoDokumentaResource) GetRobKmpDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoKmpDokument, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return domain.RobnoKmpDokument{}, fmt.Errorf("no user session found")
	}
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoKmpDokument{}, err
	}
	res := domain.RobnoKmpDokument{Header: domain.RobnoKmpHeaderView{RobnoFaktureHeaderView: dok.Header}, Poruka: dok.Poruka}
	h := &res.Header
	vrd := common.StringToInt(h.Vrd)
	if dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, vrd); err != nil {
		return res, err
	} else if found {
		h.VrstaOpis = dokvrsta.Opis
	}
	if h.TraziPopdv, err = s.robKalTraziPopdv(ctx, vrd); err != nil {
		return res, err
	}
	if h.RdokID > 0 {
		if err := s.robKmpPodaciDokumenta(ctx, h); err != nil {
			return res, err
		}
	} else {
		h.KupacKonto, h.KupacSifra = "", ""
		h.Epp, h.Odbitni = true, true
		h.IzvorniDokDatum = time.Time{}
	}
	st := &res.Stavka
	st.RdokID, st.Konto, st.Pana = h.RdokID, h.ProdavnicaKonto, h.ProdavnicaSifra
	st.Vrd, st.Brdok, st.DatumDokument = dok.Stavka.Vrd, dok.Stavka.Brdok, dok.Stavka.DatumDokumenta
	rows, err := s.robno.MagacinOpsti(ctx, h.MagaciniID, h.RnalID)
	if err != nil {
		return res, err
	}
	if rows != nil && len(*rows) > 0 {
		st.Maloprodaja = robKmpMaloprodaja((*rows)[0].C2)
	}
	return res, nil
}

// robKmpPodaciDokumenta fills the columns of a saved kalkulacija maloprodaje the faktura has not: the prodavnica with
// its opis (the naziv of its konto and šifra in the kontni plan), the EPP polje, "PDV se može odbiti" and the plaćen
// porez.
func (s *RobnoDokumentaResource) robKmpPodaciDokumenta(ctx context.Context, h *domain.RobnoKmpHeaderView) error {
	pkto, pana, err := s.robno.RdokPktoPana(ctx, h.RdokID)
	if err != nil {
		return err
	}
	h.ProdavnicaKonto, h.ProdavnicaSifra = pkto, pana
	if pkto != "" && pana != "" {
		if h.ProdavnicaOpis, _, err = s.robno.FkplNaziv(ctx, pkto, pana); err != nil {
			return err
		}
	}
	k, err := s.robno.RdokKalkulacijaPodaci(ctx, h.RdokID)
	if err != nil {
		return err
	}
	h.Epp = k.FseppID.Valid || strings.TrimSpace(k.EppPolje) != ""
	h.FseppID, h.FseppText, h.Odbitni = k.FseppID, k.FseppText, k.Odbitni
	h.Porezi, err = s.robno.RppoDokumenta(ctx, h.RdokID)
	return err
}

// ValidateRobKmp validates the header of a kalkulacija maloprodaje before it is saved, like the legacy ValidacijaRDOK:
//   - the vrsta exists; the datum dokumenta is valid and not after the datum naloga; the broj dokumenta is a whole
//     number; the broj and the datum of the izvorni dokument are required;
//   - a given org. jedinica and mesto troška exist (the mesto troška of the org. jedinica; not required);
//   - the konto and the šifra of the dobavljač are required and it exists (fkpl);
//   - the šifra of the prodavnica has at least 2 characters and the konto and the šifra of the prodavnica exist in the
//     kontni plan;
//   - the POPDV polje of a PDV obveznik whose vrsta requires it (robKalPopdv, with the tip PDV of the dobavljač);
//   - the rows of the plaćen porez (robKalProveriPoreze).
//
// Without errors and without Potvrda, an izvorni dokument the dobavljač already has in another document of the vrsta
// gives the question of the legacy ("Izvorni dokument ... postoji! Nastavak unosa?"), pitanje.
func (s *RobnoDokumentaResource) ValidateRobKmp(ctx context.Context, params domain.RobnoKmpParams) (fieldErrors []domain.FieldError, pitanje string) {
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

	// The org. jedinica and the mesto troška are not required (empty or "-": NULL); given, the checks of the opšti
	// dokument.
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)

	// The dobavljač.
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

	// The prodavnica.
	pkto, pana := strings.TrimSpace(params.ProdavnicaKonto), strings.TrimSpace(params.ProdavnicaSifra)
	switch {
	case len([]rune(pana)) < 2:
		add("pana", "Šifra mora biti najmanje dvocifrena.")
	case pkto == "":
		add("pkto", common.ErrMsgObavezanPodatak)
	default:
		if exists, err := s.robno.FkplAnalitikaPostoji(ctx, pkto, pana); err != nil {
			add("pana", common.ErrMsgReadData)
		} else if !exists {
			add("pana", "Konto i šifra prodavnice ne postoje u kontnom planu.")
		}
	}

	if trazi, err := s.robKalTraziPopdv(ctx, vrd); err != nil {
		add("popdv", common.ErrMsgReadData)
	} else if trazi {
		s.robKalPopdv(ctx, "popdv", params.Popdv, tippdv, add)
	}
	s.robKalProveriPoreze(ctx, "porezi", robKalRedovi(params.PorezOsn, params.PorezPpor, params.PorezPo, nil, nil), false, add)
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

// SaveRobKmp saves the header of a kalkulacija maloprodaje (the legacy "Sačuvaj" after ValidacijaRDOK: NoviSLOGRDOK)
// and returns its rdokid, in one transaction:
//   - the robni dokument with the fields of the header (SaveRobniDokument, a new one counted in the nalog): the datumi,
//     the opis, the izvorni dokument, the dobavljač, the prodavnica, the uslovi nabavke (the način plaćanja also of a
//     saved document, which the legacy keeps only for a new one), the org. jedinica and the mesto troška, the POPDV
//     polje, the EPP polje (fseppid and its polje; without the EPP 0 and "") and "PDV se može odbiti";
//   - the plaćen porez (rppo of the document) is the rows of its grid (robKalPorezi: with the vrsta, the broj and the
//     datum of the document, the dobavljač and the broj of the izvorni dokument as the račun).
func (s *RobnoDokumentaResource) SaveRobKmp(ctx context.Context, params domain.RobnoKmpParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	fseppID := id(params.Fsepp)
	eppPolje := ""
	var err error
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
			robPoljeText("pkto", strings.TrimSpace(params.ProdavnicaKonto)),
			robPoljeText("pana", strings.TrimSpace(params.ProdavnicaSifra)),
			robPoljeText("pla", params.UsloviPlacanja),
			robPoljeText("foot", params.Tekst),
			robPoljeText("pornapomena", params.NapomenaPoresko),
			robPoljeID("idorgjed", id(params.Orgjed)),
			robPoljeID("mestotrid", id(params.MestoTroska)),
			robPoljeText("polje", strings.ReplaceAll(strings.TrimSpace(params.Popdv), ",", ".")),
			robPoljeText("epp_polje", eppPolje),
			{Name: "fseppid", Value: strconv.FormatInt(fseppID, 10)},
			{Name: "odbitni", Value: strconv.FormatBool(params.Odbitni)},
		},
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
		return nil
	})
}

// DeleteRobKmp deletes a kalkulacija maloprodaje with its stavke (the legacy BTN_Brisanje): a document printed on the
// fiskalni printer or posted is not deleted (poruka says why); else every stavka is taken back from its nalog and from
// the stanje (robKmpBrisiStavkuTx) and deleted, its plaćen porez and avans are deleted, the document is deleted and the
// nalog counts one document less (robBrisiDokumentSa). A kalkulacija maloprodaje has no zavisni troškovi of other
// dobavljači and no rows of the knjiga evidencije prometa (the legacy comment).
func (s *RobnoDokumentaResource) DeleteRobKmp(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokumentSa(ctx, rdokID, false, s.robKmpBrisiStavkuTx, s.robno.RppoDokumentaBrisi, s.robno.AppoDokumentaBrisi)
}

// ---------------------------------------------------------------------------------------------------
// The stavke (the legacy rob_Kalkulacije_Maloprodaja_Stavke)
// ---------------------------------------------------------------------------------------------------

// GetRobKmpArtikal returns the artikal of a stavka of a kalkulacija maloprodaje with what the form of the stavke shows
// of it, like the legacy exit of EDT_SIFRA: the šifra 0 is an artikal entered by its naziv (found, nothing else); an
// other šifra (or a barkod, a value longer than 6 characters) only in a magacin of the maloprodaja (else Poruka, the
// legacy message): the artikal with its naziv, JM, tarifa with its stopa on the datum of the document, its stanje in the
// konto of the prodavnica (konto) and its poslednja cena (the last maloprodajna cena with the PDV of its cene, the
// legacy RCENE.MCENAP). found is false when the artikal does not exist.
func (s *RobnoDokumentaResource) GetRobKmpArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobKmpArtikal, bool, error) {
	a := domain.RobKmpArtikal{}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	if vrednost == "" {
		return a, false, nil
	}
	if vrednost == "0" {
		return a, true, nil
	}
	if !robKmpMaloprodaja(k.Tipmag) {
		a.Poruka = robKmpMsgMaloprodaja
		return a, true, nil
	}
	sif, found, err := s.robno.RsifArtikalKalkulacije(ctx, len(vrednost) > 6, vrednost, robNullTime(k.Dadok))
	if err != nil || !found {
		return a, false, err
	}
	a.Sifra, a.Naziv, a.Jm, a.Tarifa, a.Stopa = sif.Sifra, sif.Naziv, sif.Jm, sif.Po, sif.Stopa
	if _, a.PoslCena, _, err = s.robno.RceneMaloprodajne(ctx, a.Sifra); err != nil {
		return a, true, err
	}
	c, stFound, err := s.robno.RstaKalkulacijaCene(ctx, k.MagaciniID, konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if stFound {
		a.Stanje = c.Stanje
	}
	return a, true, nil
}

// robKmpCeneStavke are the cene of a stavka of a kalkulacija maloprodaje computed from its form (robKmpCene).
type robKmpCeneStavke struct {
	kolicina, fcena, rab, ztro, ncena, mma, mcena, mcenap float64
	tz                                                    string
}

// robKmpCene computes the cene of a stavka from the form: the zavisni trošak without its oznaka 0, the nabavna cena
// (RobKmpNabavnaCena), the MP cena with the PDV as entered (computed from the marža and the stopa when not entered) and
// the maloprodajna cena without the porez (RobKalMcena).
func robKmpCene(params domain.RobKmpStavkaParams, stopa float64) robKmpCeneStavke {
	c := robKmpCeneStavke{
		kolicina: common.StringToFloat64(params.Kolic), fcena: common.StringToFloat64(params.Fcena), rab: common.StringToFloat64(params.Rabat),
		tz: robKalOznaka(params.Tz), ztro: common.StringToFloat64(params.Ztro), mma: common.StringToFloat64(params.Mma),
	}
	if c.tz == robKalBezTroska || c.ztro == 0 {
		c.tz, c.ztro = robKalBezTroska, 0
	}
	c.ncena = RobKmpNabavnaCena(c.fcena, c.rab, c.tz, c.ztro)
	if c.mcenap = common.StringToFloat64(params.Mcenap); c.mcenap == 0 {
		c.mcenap = RobKalMpCena(c.ncena, c.mma, stopa)
	}
	c.mcena = RobKalMcena(c.ncena, c.mma)
	return c
}

// ValidateRobKmpStavka validates a stavka of a kalkulacija maloprodaje before it is saved, like the legacy Validacija
// of rob_Kalkulacije_Maloprodaja_Stavke: the document exists; the artikal exists (the šifra 0 too: the artikal of the
// stavke entered by their naziv) and an artikal other than the šifra 0 only in a magacin of the maloprodaja; the
// tarifa of the šifra 0 is given and exists (the legacy exit of EDT_PO); the količina and the fakturna cena are
// greater than 0, the rabat is not negative, the nabavna cena (computed) is greater than 0, the MP cena is not
// negative.
func (s *RobnoDokumentaResource) ValidateRobKmpStavka(ctx context.Context, params domain.RobKmpStavkaParams) []domain.FieldError {
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
	sifra, err := strconv.Atoi(strings.TrimSpace(params.Sifra))
	switch {
	case err != nil:
		add("sifraartikla", "Artikal sa takvom šifrom ne postoji.")
	case sifra != 0 && !robKmpMaloprodaja(k.Tipmag):
		add("sifraartikla", robKmpMsgMaloprodaja)
	default:
		if exists, err := s.robno.RsifPostoji(ctx, sifra); err != nil {
			add("sifraartikla", common.ErrMsgReadData)
		} else if !exists {
			add("sifraartikla", "Artikal sa takvom šifrom ne postoji.")
		}
	}
	if err == nil && sifra == 0 {
		if po, err := strconv.Atoi(strings.TrimSpace(params.Tarifa)); err != nil || po <= 0 {
			add("tarifa", "Morate uneti simbol poreza.")
		} else if exists, err := s.robno.RporPostoji(ctx, po); err != nil {
			add("tarifa", common.ErrMsgReadData)
		} else if !exists {
			add("tarifa", "Poreska tarifa ne postoji.")
		}
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
	if common.StringToFloat64(params.Fcena) > 0 && robKmpCene(params, 0).ncena <= 0 {
		add("fakturnacena", "Nabavna cena mora biti pozitivna vrednost.")
	}
	if common.StringToFloat64(params.Mcenap) < 0 {
		add("mpcena", "Maloprodajna cena ne može biti negativna.")
	}
	return fieldErrors
}

// SaveRobKmpStavka saves a stavka of a kalkulacija maloprodaje, a new one (RproID 0, the legacy NoviSLOGRPRO) or a
// changed one ("Izmeni" of its row: the old stavka taken back, the new one posted), in one transaction
// (robKmpSnimiStavku).
func (s *RobnoDokumentaResource) SaveRobKmpStavka(ctx context.Context, params domain.RobKmpStavkaParams) (rproID int64, err error) {
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
	if rproID, err = s.robKmpSnimiStavku(ctx, tx, k, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robKmpSnimiStavku saves a stavka of a kalkulacija maloprodaje in the transaction (see SaveRobKmpStavka):
//   - the artikal: of the šifra 0 the naziv, the JM and the tarifa of the form, else the ones of the artikal;
//   - the cene of the stavka (robKmpCene with the stopa of the tarifa on the datum of the document);
//   - the iznos is the količina by the MP cena with the PDV, rounded to 2 decimals; for a storno the količina and the
//     iznos are negative;
//   - a changed stavka is taken back first (robKmpVratiStavku);
//   - the stavka in the konto of the prodavnica (rdok.pkto, rdok.pana) with the columns of its document, the artikal,
//     the cene (fakturna cena, rabat, nabavna cena, zavisni trošak with its oznaka, maloprodajna marža, maloprodajna cena
//     without and with the porez; no VP cena) and, a new one, the next redni broj;
//   - the nalog and the stanje get the iznos and the količina (AzurirajRNAL, AzurirajRSTA with the cene of the stavka:
//     its cena the MP cena of an artikal of a magacin of the maloprodaja, else 0; the legacy SR, RSAL and RSIN are not
//     kept);
//   - the document gets the iznos and a new stavka counts in it.
func (s *RobnoDokumentaResource) robKmpSnimiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, params domain.RobKmpStavkaParams) (int64, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	rsifID, po, found, err := s.robno.RsifIDPo(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("artikal ne postoji")
	}
	naziv, jm := strings.TrimSpace(params.Naziv), strings.TrimSpace(params.Jm)
	if sifra == 0 {
		po = common.StringToInt(params.Tarifa)
	} else if _, naziv, jm, err = s.robno.RsifIDNazivJm(ctx, tx, sifra); err != nil {
		return 0, err
	}
	stopa := 0.0
	if k.Dadok.Valid {
		if stopa, err = s.robno.RporStopa(ctx, tx, po, k.Dadok.Time); err != nil {
			return 0, err
		}
	}
	c := robKmpCene(params, stopa)
	kolicina := c.kolicina
	iznos := robZaokruzi(kolicina*c.mcenap, 2)
	if k.Storno() {
		kolicina, iznos = -math.Abs(kolicina), -math.Abs(iznos)
	}
	pkto, pana, err := s.robno.RdokPktoPanaTx(ctx, tx, k.RdokID)
	if err != nil {
		return 0, err
	}
	pkto, pana = strings.TrimSpace(pkto), strings.TrimSpace(pana)

	// A changed stavka: its old values are taken back.
	stara := domain.RobKalStavkaDto{}
	if params.RproID > 0 {
		if stara, err = s.robno.RproKalkulacija(ctx, tx, params.RproID, k.RdokID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("stavka ne postoji")
			}
			return 0, err
		}
		if err = s.robKmpVratiStavku(ctx, tx, k, stara); err != nil {
			return 0, err
		}
	}

	polja := []domain.Fields{
		{Name: "konto", Value: pkto},
		robPoljeText("pana", pana),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		{Name: "po", Value: strconv.Itoa(po)},
		robPoljeText("naz1", naziv),
		robPoljeText("jm", jm),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		robPoljeFloat("fcena", c.fcena),
		robPoljeFloat("rab", c.rab),
		robPoljeFloat("ncena", c.ncena),
		robPoljeFloat("ztro", c.ztro),
		{Name: "tz", Value: c.tz},
		robPoljeFloat("mma", c.mma),
		robPoljeFloat("mcena", c.mcena),
		robPoljeFloat("mcenap", c.mcenap),
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
			return 0, err
		}
	}

	// The nalog and the stanje.
	if err := s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos-stara.Iznos); err != nil {
		return 0, err
	}
	cenaStanja := 0.0
	if sifra != 0 && robKmpMaloprodaja(k.Tipmag) {
		cenaStanja = c.mcenap
	}
	if err := s.AzurirajRSTA(ctx, tx, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: pkto, Sifra: sifra, Kolicina: kolicina,
		Iznos: iznos, Cena: cenaStanja, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Mma: c.mma, Ztro: c.ztro,
		Fcena: c.fcena, Mcena: c.mcena, Ncena: c.ncena, Rab: c.rab, Mcenap: c.mcenap, Tz: c.tz, Tzi: robKalBezTroska, Brst: 1}); err != nil {
		return 0, err
	}

	// The document: its iznos, a new stavka counts in it.
	noviBrst := 0
	if params.RproID == 0 {
		noviBrst = 1
	}
	if err := s.robno.RdokDodajIznos(ctx, tx, k.RdokID, iznos-stara.Iznos, noviBrst); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robKmpVratiStavku takes a stavka of a kalkulacija maloprodaje back in the transaction (the legacy BrisiSLOGRPRO): its
// iznos from the nalog (AzurirajRNAL) and its količina and iznos from the stanje with the cene of the stavka
// (AzurirajRSTA, a stavka less).
func (s *RobnoDokumentaResource) robKmpVratiStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, st domain.RobKalStavkaDto) error {
	if err := s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -st.Iznos); err != nil {
		return err
	}
	return s.AzurirajRSTA(ctx, tx, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: strings.TrimSpace(st.Konto), Sifra: st.Sifra,
		Kolicina: -st.Kolicina, Iznos: -st.Iznos, Cena: st.Cena, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok,
		Vma: st.Vma, Mma: st.Mma, Ztro: st.Ztro, Fcena: st.Fcena, Mcena: st.Mcena, Ncena: st.Ncena, Rab: st.Rab, Ztrin: st.Ztrin,
		Mcenap: st.Mcenap, Tz: robKalOznaka(st.Tz), Tzi: robKalBezTroska, Brst: -1})
}

// robKmpBrisiStavkuTx deletes a stavka of a kalkulacija maloprodaje in the transaction of the caller and returns its
// document (the legacy BrisiSLOGRPRO and the stavke of BTN_Brisanje): the stavka is deleted and taken back
// (robKmpVratiStavku), the document loses its iznos and counts one stavka less (the legacy forgets rdok.brst).
func (s *RobnoDokumentaResource) robKmpBrisiStavkuTx(ctx context.Context, tx db.Transaction, rproID int64) (int64, error) {
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
	if err := s.robKmpVratiStavku(ctx, tx, k, st); err != nil {
		return 0, err
	}
	return rdokID, s.robno.RdokOduzmiStavku(ctx, tx, rdokID, st.Iznos)
}

// DeleteRobKmpStavka deletes a stavka of a kalkulacija maloprodaje (the legacy BrisiSLOGRPRO) in one transaction and
// returns its document: the stavka is taken back and deleted (robKmpBrisiStavkuTx), the stavke after it get the redni
// brojevi again (rbr and brst).
func (s *RobnoDokumentaResource) DeleteRobKmpStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
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
	if _, err = s.robKmpBrisiStavkuTx(ctx, tx, rproID); err != nil {
		return 0, err
	}
	if err = s.robno.RproPrenumerisiSaBrst(ctx, tx, rdokID, st.Rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// GetRobKmpStavkaForma returns a saved stavka of a kalkulacija maloprodaje for the form of the stavke ("Izmeni" of its
// row): its artikal (of the šifra 0 its naziv, JM and tarifa), its količina (as entered: a storno has it saved
// negative), cene and zavisni trošak, with the stopa of its tarifa on the datum of the document and, of an artikal, its
// stanje and poslednja cena. found is false without it.
func (s *RobnoDokumentaResource) GetRobKmpStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobKmpStavkaForma, bool, error) {
	f := domain.RobKmpStavkaForma{}
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
	f.RproID, f.Sifra, f.Naziv, f.Jm, f.Tarifa = st.RproID, st.Sifra, strings.TrimSpace(st.Naz1), strings.TrimSpace(st.Jm), st.Po
	f.Kolicina, f.Fcena, f.Rab, f.Ncena = math.Abs(st.Kolicina), st.Fcena, st.Rab, st.Ncena
	f.Tz, f.Ztro, f.Mma, f.Mcenap = robKalOznaka(st.Tz), st.Ztro, st.Mma, st.Mcenap
	if k.Dadok.Valid {
		if f.Stopa, err = s.robno.RporStopa(ctx, nil, f.Tarifa, k.Dadok.Time); err != nil {
			return f, true, err
		}
	}
	if st.Sifra != 0 {
		a, aFound, err := s.GetRobKmpArtikal(ctx, rdokID, strconv.Itoa(st.Sifra), st.Konto)
		if err != nil {
			return f, true, err
		}
		if aFound && a.Poruka == "" {
			f.Stanje, f.PoslCena = a.Stanje, a.PoslCena
			if f.Naziv == "" {
				f.Naziv, f.Jm = a.Naziv, a.Jm
			}
		}
	}
	return f, true, nil
}

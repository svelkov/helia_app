package robno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Knjižna pisma (the groups KNO and KNZ, the legacy rob_KnjiznoPismoFIN with its stavke rob_KnjiznoPismoFINDET
// and the window rob_KnjPismoPR_Obracun): a robni dokument (rdok) of a kupac that corrects a faktura (the vezni
// dokument, rdok.vrdokid: its stavke are copied, rpro.rproid1, and corrected) or, without it, gives only the
// osnovica and the porez (rdok.iznos, rdok.vporez, the % poreza in rdok.pkase)
// ---------------------------------------------------------------------------------------------------

// The messages of the knjižna pisma.
const (
	robKnpMsgDupla      = "Ova stavka je već uneta, nedozvoljen dupli unos."
	robKnpMsgNijeVezan  = "Dokument nije vezan za fakturu, pregled fakture nije dostupan."
	robKnpMsgNemaStavki = "Dokument nema stavki."
)

// RobKnpStavkeTableRows returns the stavke of a knjižno pismo or of the faktura it corrects as the rows of a grid
// (the columns of GetKnpStavkeTableFields, the legacy TABLE_rpro and TABLE_rpro1: redni broj, konto, šifra,
// naziv, JM, OTK, serija, rok, količina, fakturna cena, rabat, iznos). poRbr gives them by the redni broj (the
// stavke of the faktura), else the last one first (the corrected stavke, the legacy ORDER BY RBR DESC).
func RobKnpStavkeTableRows(stavke []domain.RobStavkaRow, poRbr bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := range stavke {
		st := stavke[i]
		if poRbr {
			st = stavke[len(stavke)-1-i]
		}
		rok := ""
		if st.Roktr != 0 {
			rok = robnoStampaWinDevDatumFormat(int64(st.Roktr), common.DateLayout)
		}
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm, st.Otk, st.Serija, rok,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Fcena), robFormatIznos(st.Rab), robFormatIznos(st.Iznos)},
		})
	}
	return rows
}

// robKnpTipKolona is the position of the tip of the knjižno pismo in the grid of the documents of the "Prikaz
// knjižnih pisama" tab: after the naziv partnera (the columns of the fakture, RobDokumentaPrikazTableRows).
const robKnpTipKolona = 10

// RobKnpDokumentaPrikazTableRows returns the knjižna pisma as the rows of the grid of the documents of the "Prikaz
// knjižnih pisama" tab: the columns of the fakture (RobDokumentaPrikazTableRows) with the tip of the knjižno pismo
// after the naziv partnera (the legacy TABLE_QRY_RDOK of rob_KnjiznoPismoFIN).
func RobKnpDokumentaPrikazTableRows(dokumenta []domain.RobDokumentPrikazRow) []domain.TableRow {
	rows := RobDokumentaPrikazTableRows(dokumenta)
	for i := range rows {
		f := rows[i].Fields
		rows[i].Fields = append(append(append([]string{}, f[:robKnpTipKolona]...), dokumenta[i].TipKnjPisma), f[robKnpTipKolona:]...)
	}
	return rows
}

// robKnpPodaci reads the columns of the knjižno pismo (or of the faktura) of the id the screen needs besides
// RobniDokumentRow (see RobnoRepository.RdokKnjiznoPismo); found is false without it.
func (s *RobnoDokumentaResource) robKnpPodaci(ctx context.Context, rdokID int64) (domain.RobKnpDokumentDto, bool, error) {
	rows, err := s.robno.RdokKnjiznoPismo(ctx, rdokID)
	if err != nil || len(rows) == 0 {
		return domain.RobKnpDokumentDto{}, false, err
	}
	return rows[0], true, nil
}

// GetRobKnpDokument returns the document of the "Knjižna pisma" screen: the document of the params like the one of
// the faktura (GetRobFaktureDokument: an existing one of the nalog that can be changed, else a new one with the
// next broj, NoviBrojRDOK, the konto of the kupci, rvr.fkup, and the domaća valuta with the kurs and the važi za
// 1) with the opis of its vrsta. A saved document has its tip, and its vezni dokument with the new ugovoreni rabat
// and kasa, or without it the osnovica and the porez (the legacy FileToScreen of BTN_TraziDOK); a new one is of
// today (EDT_DADOK = Today()). The stavke repeat the kupac (the legacy EDT_KUPAC).
func (s *RobnoDokumentaResource) GetRobKnpDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoKnpDokument, error) {
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoKnpDokument{}, err
	}
	res := domain.RobnoKnpDokument{Header: domain.RobnoKnpHeaderView{RobnoFaktureHeaderView: dok.Header}, Poruka: dok.Poruka}
	h := &res.Header
	if dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(h.Vrd)); err != nil {
		return res, err
	} else if found {
		h.VrstaOpis = dokvrsta.Opis
	}
	if h.RdokID > 0 {
		knp, found, err := s.robKnpPodaci(ctx, h.RdokID)
		if err != nil {
			return res, err
		}
		if found {
			h.TipKnjID, h.TipKnjText = knp.TipknjID, knp.TipText
			if knp.VrdokID > 0 {
				// The vezni dokument (its data when it is a document of the period; the link stays anyway).
				if h.Vezni, _, err = s.GetRobKnpVezni(ctx, knp.VrdokID); err != nil {
					return res, err
				}
				h.Vezni.RdokID = knp.VrdokID
				h.Ugrabat, h.Vrugrabat, h.Pkase, h.Vkase = knp.Ugrabat, knp.Vrugrabat, knp.Pkase, knp.Vkase
			} else {
				h.Osnovica, h.VrednostPoreza, h.ProcPoreza = knp.Iznos, knp.Vporez, knp.Pkase
				h.IznosDokumenta = robZaokruzi(knp.Iznos+knp.Vporez, 2)
			}
		}
	} else {
		h.DatumDokum = time.Now()
		h.IzvorniDokDatum = time.Time{}
	}
	res.Stavke = domain.RobnoKnpStavke{RdokID: h.RdokID, Vrd: h.Vrd, Brdok: h.BrojDokum, IznosDokumenta: dok.Stavka.IznosDokumenta}
	if !h.DatumDokum.IsZero() {
		res.Stavke.DatumDokumenta = h.DatumDokum.Format(common.DateLayout)
	}
	if strings.TrimSpace(h.KupacNaziv) != "" {
		res.Stavke.Kupac = strings.TrimSpace(fmt.Sprintf("%s %s %s ;%s ;%s %s", h.KupacKonto, h.KupacSifra, h.KupacNaziv,
			h.KupacAdresa, h.KupacPostanskiBroj, h.KupacMesto))
	}
	return res, nil
}

// GetRobKnpVezni returns the faktura of the id a knjižno pismo corrects with what the header takes of it (the
// legacy BTN_TraziOTP): its broj and datum, its kupac with its data, its org. jedinica and mesto troška, its
// ugovoreni rabat and kasa and its valuta; found is false when it is not a document of the period.
func (s *RobnoDokumentaResource) GetRobKnpVezni(ctx context.Context, rdokID int64) (domain.RobnoKnpVezni, bool, error) {
	v := domain.RobnoKnpVezni{}
	row, found, err := s.GetRobniDokumentByID(ctx, rdokID)
	if err != nil || !found {
		return v, false, err
	}
	v.RdokID, v.Dokum = row.RdokID, row.Dokum
	if row.Dadok.Valid {
		v.Datum = row.Dadok.Time.Format(common.HtmlLayout)
	}
	v.KupacKonto, v.KupacSifra = row.Fkto, row.Fana
	v.IDOrgjed, v.OrgjedText = row.IDOrgjed.Int64, row.OrgjedText
	v.MestoTrID, v.MestoTrText = row.MestoTrID.Int64, row.MestoTrText
	v.IDValute, v.ValutaText, v.Kurs, v.Vaziza = row.IDValute.Int64, row.ValutaText, row.Kurs, row.Vaziza
	kupac, kFound, err := s.GetRobKupac(ctx, row.Fkto, row.Fana)
	if err != nil {
		return v, true, err
	}
	if kFound {
		v.KupacNaziv, v.KupacAdresa, v.KupacPostanskiBroj, v.KupacMesto = kupac.Naziv, kupac.Adresa, kupac.Pobro, kupac.Mesto
		v.KupacPib, v.SistemPdv = kupac.Pib, strconv.Itoa(kupac.Tippdv)
	}
	knp, kFound, err := s.robKnpPodaci(ctx, rdokID)
	if err != nil {
		return v, true, err
	}
	if kFound {
		v.Ugrabat, v.Vrugrabat, v.Pkase, v.Vkase = knp.Ugrabat, knp.Vrugrabat, knp.Pkase, knp.Vkase
	}
	return v, true, nil
}

// ValidateRobKnp validates the header of a knjižno pismo before it is saved, like the legacy ValidacijaRDOK of
// rob_KnjiznoPismoFIN: the datum dokumenta is a valid date, not after the datum naloga; the konto and the šifra of
// the kupac are required and the kupac exists (fkpl); a given komercijalista exists; the šifra valute is
// required; a given org. jedinica and mesto troška exist (not required); a given tip of the knjižno pismo exists. The
// broj dokumenta is a whole number and no other nalog has the document of the vrsta and broj in the magacin (the
// legacy BTN_TraziDOK); a given vezni dokument is a document of the period.
func (s *RobnoDokumentaResource) ValidateRobKnp(ctx context.Context, params domain.RobnoKnpParams) []domain.FieldError {
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
	if datum, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && datum.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}
	if broj := strings.TrimSpace(params.BrojDokum); broj == "" {
		add("brdok", common.ErrMsgObavezanPodatak)
	} else if dokum, err := strconv.Atoi(broj); err != nil {
		add("brdok", robnoUnosMsgCeoBroj)
	} else if exists, err := s.robno.RdokBrojDrugogNalogaPostoji(ctx, params.MagaciniID, vrd, dokum, params.RnalID, params.RdokID); err != nil {
		add("brdok", common.ErrMsgReadData)
	} else if exists {
		add("brdok", "Postoji u drugom nalogu.")
	}

	// The kupac.
	konto, sifra := strings.TrimSpace(params.KupacKonto), strings.TrimSpace(params.KupacSifra)
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

	// The komercijalista (when given), the valuta, the org. jedinica and the mesto troška, the tip (when given).
	robnoUnosNeobavezanID := func(value, field string, postoji func(context.Context, int64) (bool, error)) {
		if id, given, ok := robnoUnosID(value); !ok {
			add(field, robnoUnosMsgNePostoji)
		} else if given {
			if exists, err := postoji(ctx, id); err != nil {
				add(field, common.ErrMsgReadData)
			} else if !exists {
				add(field, robnoUnosMsgNePostoji)
			}
		}
	}
	robnoUnosNeobavezanID(params.Komercijalista, "komid", s.robno.KomercijalistaPostoji)
	if _, given, ok := robnoUnosID(params.IDValute); !ok || !given {
		add("idvalute", common.ErrMsgObavezanPodatak)
	}
	s.robnoUnosPostojiID(ctx, params.Orgjed, "idorgjed", s.robno.OrgjedPostoji, add)
	s.robnoUnosPostojiID(ctx, params.MestoTroska, "mestotrid", s.robno.MestotrPostoji, add)
	robnoUnosNeobavezanID(params.TipKnjID, "tipknjid", s.robno.TipknpismaPostoji)

	// The vezni dokument of a new knjižno pismo (a saved one keeps its own, the legacy BTN_TraziOTP grayed).
	if id, given, ok := robnoUnosID(params.VezniRdokID); !ok {
		add("vrdokid", robnoUnosMsgNePostoji)
	} else if given && params.RdokID <= 0 {
		if _, found, err := s.GetRobniDokumentByID(ctx, id); err != nil {
			add("vrdokid", common.ErrMsgReadData)
		} else if !found {
			add("vrdokid", robnoUnosMsgNePostoji)
		}
	}
	return fieldErrors
}

// SaveRobKnp saves the header of a knjižno pismo (the legacy "Sačuvaj" of rob_KnjiznoPismoFIN, after
// ValidacijaRDOK: SnimiRDOK) and returns its rdokid: the robni dokument with the fields of the header, its tip,
// the poreska kategorija (keykat, category) and the old columns of the šifarnici (kom, sifval) by their keys, a
// new one when the nalog has no document of the vrsta and broj in the magacin (SaveRobniDokument):
//   - with a vezni dokument (rdok.vrdokid; a saved knjižno pismo keeps its own) its broj and datum (dokiz, datiz)
//     and the new ugovoreni rabat and kasa (ugrabat, pkase); then the document is computed again (Preracun);
//   - without it the osnovica (iznos), the vrednost poreza (vporez) and the % poreza (pkase).
func (s *RobnoDokumentaResource) SaveRobKnp(ctx context.Context, params domain.RobnoKnpParams) (int64, error) {
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	komID, valutaID, porkatID := id(params.Komercijalista), id(params.IDValute), id(params.PorKategorija)
	vezniID := id(params.VezniRdokID)
	if params.RdokID > 0 {
		knp, _, err := s.robKnpPodaci(ctx, params.RdokID)
		if err != nil {
			return 0, err
		}
		vezniID = knp.VrdokID
	}
	sif, err := s.robno.SifreSifarnika(ctx, 0, komID, valutaID, 0, porkatID)
	if err != nil {
		return 0, err
	}
	polja := []domain.Fields{
		robPoljeDatum("dadok", params.DatumDokum),
		robPoljeText("opis", params.Opis),
		robPoljeText("fkto", params.KupacKonto),
		robPoljeText("fana", params.KupacSifra),
		robPoljeID("komid", komID), robPoljeInt("kom", sif.C1),
		robPoljeID("idorgjed", id(params.Orgjed)),
		robPoljeID("mestotrid", id(params.MestoTroska)),
		robPoljeText("foot", params.Tekst),
		robPoljeText("keykat", sif.C4),
		robPoljeText("category", sif.C5),
		robPoljeText("pornapomena", params.NapomenaPoresko),
		robPoljeID("tipknjid", id(params.TipKnjID)),
		robPoljeID("idvalute", valutaID), robPoljeInt("sifval", sif.C2),
		robPoljeBroj("kurs", params.Kurs),
		robPoljeInt("vaziza", params.VaziZa),
		{Name: "vrdokid", Value: strconv.FormatInt(vezniID, 10)},
	}
	if vezniID > 0 {
		polja = append(polja, robPoljeBroj("ugrabat", params.Ugrabat), robPoljeBroj("pkase", params.Pkase))
		if vezni, found, err := s.GetRobniDokumentByID(ctx, vezniID); err != nil {
			return 0, err
		} else if found {
			polja = append(polja, robPoljeText("dokiz", strconv.Itoa(vezni.Dokum)), robPoljeNullTime("datiz", vezni.Dadok))
		}
	} else {
		polja = append(polja,
			robPoljeFloat("iznos", robZaokruzi(common.StringToFloat64(params.Osnovica), 2)),
			robPoljeFloat("vporez", robZaokruzi(common.StringToFloat64(params.VrednostPoreza), 2)),
			robPoljeBroj("pkase", params.ProcPoreza),
			robPoljeText("dokiz", ""), domain.Fields{Name: "datiz", IsNull: true})
	}
	rdokID, err := s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{
		RnalID:     params.RnalID,
		MagaciniID: params.MagaciniID,
		Vrd:        common.StringToInt(params.Vrd),
		Dokum:      common.StringToInt(params.BrojDokum),
		Polja:      polja,
	})
	if err != nil || vezniID <= 0 {
		return rdokID, err
	}
	return rdokID, s.robKnpPreracunDokumenta(ctx, rdokID)
}

// DeleteRobKnp deletes a knjižno pismo with its stavke (the legacy "Brisanje" of rob_KnjiznoPismoFIN): a document
// printed on the fiskalni printer (knjige[4] "S") or posted (knjige[1] "D") is not deleted (poruka says why); else
// its stavke are deleted (the legacy does not take them back from the stanje), the nalog loses the iznos of the
// document and counts its stavke again (robKnpOdknjizi), the document is deleted and the nalog counts one document
// less (robBrisiDokumentSa).
func (s *RobnoDokumentaResource) DeleteRobKnp(ctx context.Context, rdokID int64) (string, error) {
	return s.robBrisiDokumentSa(ctx, rdokID, false, s.robKnpBrisiStavkuTx, s.robKnpOdknjizi)
}

// robKnpBrisiStavkuTx deletes a stavka of a knjižno pismo in the transaction of the caller and returns its document.
func (s *RobnoDokumentaResource) robKnpBrisiStavkuTx(ctx context.Context, tx db.Transaction, rproID int64) (int64, error) {
	rdokID, err := s.robno.RproRdokID(ctx, tx, rproID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("stavka ne postoji")
		}
		return 0, err
	}
	return rdokID, s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID)
}

// robKnpOdknjizi takes the iznos of a knjižno pismo (the one its Preracun posted) from its nalog and counts the
// stavke of the nalog again (AzurirajRNAL), in the transaction of its deletion (after its stavke are deleted).
func (s *RobnoDokumentaResource) robKnpOdknjizi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return err
	}
	p, err := s.robno.RdokKnpPreracun(ctx, tx, rdokID)
	if err != nil {
		return err
	}
	iznos := 0.0
	if knp, found, err := s.robKnpPodaci(ctx, rdokID); err != nil {
		return err
	} else if found && knp.VrdokID > 0 {
		iznos = p.Iznos
	}
	return s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos)
}

// robKnpPreracun computes a knjižno pismo again in the transaction (the legacy Preracun of rob_KnjiznoPismoFINDET
// and of rob_KnjiznoPismoFIN): the fakturna vrednost (količina by the fakturna cena), its rabat and the vrednost
// by the cena of its stavke, rounded to 2 decimals; the ugovoreni rabat is taken from the fakturna vrednost less
// the rabat and the kasa from what remains (rdok.vrugrabat, rdok.vkase); the iznos of the document is the vrednost
// by the cena (rdok.iznos) and brst the number of its stavke. The nalog gets the change of the iznos and counts its
// stavke again (AzurirajRNAL, the legacy UpdateRNAL).
func (s *RobnoDokumentaResource) robKnpPreracun(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst) error {
	p, err := s.robno.RdokKnpPreracun(ctx, tx, k.RdokID)
	if err != nil {
		return err
	}
	fakturna, rabat, iznos := robZaokruzi(p.Fakturna, 2), robZaokruzi(p.Rabat, 2), robZaokruzi(p.Cena, 2)
	ugovoreni := robZaokruzi((fakturna-rabat)*p.Ugrabat/100, 2)
	kasa := robZaokruzi((fakturna-rabat-ugovoreni)*p.Pkase/100, 2)
	if err := s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", k.RdokID, []domain.Fields{
		robPoljeFloat("vkase", kasa), robPoljeFloat("vrugrabat", ugovoreni), robPoljeFloat("iznos", iznos),
		{Name: "brst", Value: strconv.Itoa(p.Stavke)},
	}); err != nil {
		return err
	}
	return s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos-p.Iznos)
}

// robKnpTx runs fn in one transaction with the context of the stavke of the knjižno pismo and computes the document
// again (robKnpPreracun) before the commit.
func (s *RobnoDokumentaResource) robKnpTx(ctx context.Context, rdokID int64, fn func(tx db.Transaction, k domain.RobStavkeKontekst) error) (err error) {
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("dokument ne postoji")
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
	if fn != nil {
		if err = fn(tx, k); err != nil {
			return err
		}
	}
	if err = s.robKnpPreracun(ctx, tx, k); err != nil {
		return err
	}
	return tx.Commit()
}

// robKnpPreracunDokumenta computes the knjižno pismo of the id again in its own transaction (the Preracun of the
// legacy SnimiRDOK of a knjižno pismo with a vezni dokument).
func (s *RobnoDokumentaResource) robKnpPreracunDokumenta(ctx context.Context, rdokID int64) error {
	return s.robKnpTx(ctx, rdokID, nil)
}

// ---------------------------------------------------------------------------------------------------
// The stavke of a knjižno pismo (the legacy rob_KnjiznoPismoFINDET)
// ---------------------------------------------------------------------------------------------------

// GetRobKnpStavke returns the stavke of the faktura the knjižno pismo corrects (the legacy TABLE_rpro,
// QRY_RRPO_DET) and the stavke of the knjižno pismo (TABLE_rpro1, PuniDETALJ1).
func (s *RobnoDokumentaResource) GetRobKnpStavke(ctx context.Context, rdokID int64) (faktura, korigovane []domain.RobStavkaRow, err error) {
	knp, found, err := s.robKnpPodaci(ctx, rdokID)
	if err != nil || !found {
		return nil, nil, err
	}
	if knp.VrdokID > 0 {
		if faktura, err = s.GetRobStavkeDokumenta(ctx, knp.VrdokID); err != nil {
			return nil, nil, err
		}
	}
	korigovane, err = s.GetRobStavkeDokumenta(ctx, rdokID)
	return faktura, korigovane, err
}

// DodajRobKnpStavke copies stavke of the faktura into the knjižno pismo (the legacy "Dodaj" and "Dodaj sve",
// AddRPRO) in one transaction and computes the document again: a copy has the artikal, the konto, the količina, the
// cene, the iznos, the rabat, the OTK, the serija and the rok of the stavka of the faktura (rproid1 is that stavka)
// and the next redni broj. A stavka already copied into the knjižno pismo is not copied again: "Dodaj" (one stavka)
// answers with the message why (poruka), "Dodaj sve" leaves it out. rproIDs empty copies all the stavke of the
// faktura.
func (s *RobnoDokumentaResource) DodajRobKnpStavke(ctx context.Context, rdokID int64, rproIDs []int64) (poruka string, err error) {
	knp, found, err := s.robKnpPodaci(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found || knp.VrdokID <= 0 {
		return robKnpMsgNijeVezan, nil
	}
	fakture, err := s.GetRobStavkeDokumenta(ctx, knp.VrdokID)
	if err != nil {
		return "", err
	}
	stavke := map[int64]domain.RobStavkaRow{}
	sve := len(rproIDs) == 0
	for i := len(fakture) - 1; i >= 0; i-- {
		stavke[fakture[i].RproID] = fakture[i]
		if sve {
			rproIDs = append(rproIDs, fakture[i].RproID)
		}
	}
	if len(rproIDs) == 0 {
		return "Ne postoji nijedna stavka za dodavanje.", nil
	}
	err = s.robKnpTx(ctx, rdokID, func(tx db.Transaction, k domain.RobStavkeKontekst) error {
		for _, id := range rproIDs {
			st, ok := stavke[id]
			if !ok {
				return fmt.Errorf("stavka %d nije stavka fakture", id)
			}
			kopija, err := s.robno.RproKontraStavka(ctx, tx, rdokID, id)
			if err != nil {
				return err
			}
			if kopija > 0 {
				if !sve {
					poruka = robKnpMsgDupla
					return nil
				}
				continue
			}
			if err := s.robKnpKopirajStavku(ctx, tx, k, st); err != nil {
				return err
			}
		}
		return nil
	})
	return poruka, err
}

// robKnpKopirajStavku adds the copy of the stavka st of the faktura to the knjižno pismo of k (the legacy AddRPRO).
func (s *RobnoDokumentaResource) robKnpKopirajStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, st domain.RobStavkaRow) error {
	rbr, err := s.robno.RproSledeciRbr(ctx, tx, k.RdokID)
	if err != nil {
		return err
	}
	_, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
		{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
		{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
		robPoljeNullTime("dadok", k.Dadok),
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		{Name: "sifra", Value: strconv.Itoa(st.Sifra)},
		{Name: "po", Value: strconv.Itoa(st.Po)},
		{Name: "rsifid", Value: strconv.FormatInt(st.RsifID, 10)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		{Name: "mag", Value: strconv.Itoa(st.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		robPoljeFloat("kolic", st.Kolic),
		robPoljeFloat("cena", st.Cena),
		robPoljeFloat("fcena", st.Fcena),
		robPoljeFloat("ncena", st.Ncena),
		robPoljeFloat("iznos", st.Iznos),
		robPoljeText("otk", st.Otk),
		robPoljeText("konto", st.Konto),
		robPoljeText("serija", st.Serija),
		{Name: "roktr", Value: strconv.FormatInt(int64(st.Roktr), 10)},
		robPoljeFloat("rab", st.Rab),
		{Name: "rproid1", Value: strconv.FormatInt(st.RproID, 10)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	})
	return err
}

// GetRobKnpStavkaForma returns a stavka of the knjižno pismo for its form (the legacy row of TABLE_rpro1): the
// artikal, the količina, the fakturna cena and the rabat; found is false without it.
func (s *RobnoDokumentaResource) GetRobKnpStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobKnpStavkaForma, bool, error) {
	stavke, err := s.GetRobStavkeDokumenta(ctx, rdokID)
	if err != nil {
		return domain.RobKnpStavkaForma{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return domain.RobKnpStavkaForma{RproID: st.RproID, Sifra: st.Sifra, Naziv: st.Naziv, Kolicina: st.Kolic, Cena: st.Fcena, Rabat: st.Rab}, true, nil
		}
	}
	return domain.RobKnpStavkaForma{}, false, nil
}

// SaveRobKnpStavka saves the corrected values of a stavka of the knjižno pismo (the legacy exit of a row of
// TABLE_rpro1): the fakturna cena, the rabat, the količina and the iznos (the količina by the fakturna cena,
// rounded to 2 decimals), then the document is computed again.
func (s *RobnoDokumentaResource) SaveRobKnpStavka(ctx context.Context, params domain.RobKnpStavkaParams) error {
	return s.robKnpTx(ctx, params.RdokID, func(tx db.Transaction, k domain.RobStavkeKontekst) error {
		rdokID, err := s.robno.RproRdokID(ctx, tx, params.RproID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && rdokID != k.RdokID) {
			return fmt.Errorf("stavka ne postoji")
		}
		if err != nil {
			return err
		}
		kolicina, cena := common.StringToFloat64(params.Kolicina), common.StringToFloat64(params.Cena)
		return s.unos.Rpro.UpdateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", params.RproID, []domain.Fields{
			robPoljeFloat("fcena", cena),
			robPoljeBroj("rab", params.Rabat),
			robPoljeFloat("kolic", kolicina),
			robPoljeFloat("iznos", robZaokruzi(kolicina*cena, 2)),
		})
	})
}

// DeleteRobKnpStavka deletes a stavka of the knjižno pismo (the legacy "Vrati jedan") and returns its document: the
// stavke after it get the redni brojevi again (rbr) and the document is computed again.
func (s *RobnoDokumentaResource) DeleteRobKnpStavka(ctx context.Context, rproID int64) (int64, error) {
	rdokID, err := s.robno.RproRdokID(ctx, nil, rproID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("stavka ne postoji")
		}
		return 0, err
	}
	return rdokID, s.robKnpTx(ctx, rdokID, func(tx db.Transaction, k domain.RobStavkeKontekst) error {
		rbr, err := s.robno.RproRbr(ctx, tx, rproID)
		if err != nil {
			return err
		}
		if _, err := s.robKnpBrisiStavkuTx(ctx, tx, rproID); err != nil {
			return err
		}
		return s.robno.RproPrenumerisi(ctx, tx, k.RdokID, rbr)
	})
}

// DeleteRobKnpStavke deletes all the stavke of the knjižno pismo (the legacy "Vrati sve") and computes the document
// again.
func (s *RobnoDokumentaResource) DeleteRobKnpStavke(ctx context.Context, rdokID int64) error {
	ids, err := s.robno.RproIDsDokumenta(ctx, rdokID)
	if err != nil {
		return err
	}
	return s.robKnpTx(ctx, rdokID, func(tx db.Transaction, _ domain.RobStavkeKontekst) error {
		for _, id := range ids {
			if _, err := s.robKnpBrisiStavkuTx(ctx, tx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------------------------------------
// The window "Pregled vrednosti fakture po starim i novim uslovima prodaje" (the legacy rob_KnjPismoPR_Obracun)
// ---------------------------------------------------------------------------------------------------

// robKnpVrednosti are the values of a document by its stavke and conditions (the legacy aaFAKT and aaPDV): the
// bruto vrednost, the rabat, the ugovoreni rabat and the kasa with their %, and the osnovica and the PDV by stopa.
type robKnpVrednosti struct {
	bruto, rabat, ugrabat, pkase, vrUgrabat, vrKase float64
	osnovica, pdv                                   map[float64]float64
}

// robKnpObracunaj returns the values of the stavke with the ugovoreni rabat and the kasa (the legacy Preracun1 and
// Preracun2): every stavka its bruto vrednost, rabat, ugovoreni rabat, kasa and osnovica (see
// robnoStampaKnjiznoPismoVrednosti) and the PDV of its osnovica by its stopa, rounded to 2 decimals.
func robKnpObracunaj(stavke []domain.RobStavkaRow, ugrabat, pkase float64) robKnpVrednosti {
	v := robKnpVrednosti{ugrabat: ugrabat, pkase: pkase, osnovica: map[float64]float64{}, pdv: map[float64]float64{}}
	for _, st := range stavke {
		bruto, rabat, ugRabat, kasa, osn := robnoStampaKnjiznoPismoVrednosti(st.Kolic, st.Fcena, st.Rab, ugrabat, pkase)
		v.bruto += bruto
		v.rabat += rabat
		v.vrUgrabat += ugRabat
		v.vrKase += kasa
		v.osnovica[st.Stopa] += osn
		v.pdv[st.Stopa] += robZaokruzi(osn*st.Stopa/100, 2)
	}
	return v
}

// GetRobKnpObracun returns the rows of the window "Pregled vrednosti fakture po starim i novim uslovima prodaje"
// (the legacy rob_KnjPismoPR_Obracun) of a knjižno pismo: the values by the new conditions (its stavke with its
// ugovoreni rabat and kasa, Preracun2), by the conditions of the faktura (the stavke of the faktura with its
// ugovoreni rabat and kasa, Preracun1) and their difference: BRUTO, RABAT, NETO, the ugovoreni rabat and the kasa
// with their %, the osnovica and the PDV of every stopa (of both, the lowest first; the legacy pairs the rows by
// their position) and the IZNOS FAKTURE. poruka says why there are no rows (no vezni dokument, no stavke).
func (s *RobnoDokumentaResource) GetRobKnpObracun(ctx context.Context, rdokID int64) (redovi []domain.RobnoKnpObracunRed, poruka string, err error) {
	knp, found, err := s.robKnpPodaci(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if !found || knp.VrdokID <= 0 {
		return nil, robKnpMsgNijeVezan, nil
	}
	faktura, korigovane, err := s.GetRobKnpStavke(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if len(korigovane) == 0 {
		return nil, robKnpMsgNemaStavki, nil
	}
	fakt, _, err := s.robKnpPodaci(ctx, knp.VrdokID)
	if err != nil {
		return nil, "", err
	}
	return robKnpObracunRedovi(robKnpObracunaj(korigovane, knp.Ugrabat, knp.Pkase), robKnpObracunaj(faktura, fakt.Ugrabat, fakt.Pkase)), "", nil
}

// robKnpObracunRedovi pairs the values by the new conditions and by the ones of the faktura in the rows of the
// window (see GetRobKnpObracun).
func robKnpObracunRedovi(novo, faktura robKnpVrednosti) []domain.RobnoKnpObracunRed {
	procenat := func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) + "%" }
	red := func(opis string, n, f float64) domain.RobnoKnpObracunRed {
		return domain.RobnoKnpObracunRed{OpisNovo: opis, OpisFaktura: opis, OpisRazlika: opis, Novo: n, Faktura: f, Razlika: robZaokruzi(n-f, 2)}
	}
	redovi := []domain.RobnoKnpObracunRed{
		red("BRUTO", novo.bruto, faktura.bruto),
		red("RABAT", novo.rabat, faktura.rabat),
		red("NETO", novo.bruto-novo.rabat, faktura.bruto-faktura.rabat),
	}
	ugovoreni := red("UGOVORENI RABAT", novo.vrUgrabat, faktura.vrUgrabat)
	ugovoreni.OpisNovo, ugovoreni.OpisFaktura = "UGOVORENI RABAT "+procenat(novo.ugrabat), "UGOVORENI RABAT "+procenat(faktura.ugrabat)
	kasa := red("KASA", novo.vrKase, faktura.vrKase)
	kasa.OpisNovo, kasa.OpisFaktura = "KASA "+procenat(novo.pkase), "KASA "+procenat(faktura.pkase)
	redovi = append(redovi, ugovoreni, kasa)

	stope := map[float64]bool{}
	for stopa := range novo.osnovica {
		stope[stopa] = true
	}
	for stopa := range faktura.osnovica {
		stope[stopa] = true
	}
	sortirane := make([]float64, 0, len(stope))
	for stopa := range stope {
		sortirane = append(sortirane, stopa)
	}
	sort.Float64s(sortirane)
	var pdvNovo, pdvFaktura float64
	for _, stopa := range sortirane {
		redovi = append(redovi,
			red("Osnovica "+procenat(stopa), novo.osnovica[stopa], faktura.osnovica[stopa]),
			red("PDV "+procenat(stopa), novo.pdv[stopa], faktura.pdv[stopa]))
		pdvNovo += novo.pdv[stopa]
		pdvFaktura += faktura.pdv[stopa]
	}
	iznos := func(v robKnpVrednosti, pdv float64) float64 {
		return robZaokruzi(v.bruto-v.rabat-v.vrUgrabat-v.vrKase+pdv, 2)
	}
	redovi = append(redovi, red("IZNOS FAKTURE", iznos(novo, pdvNovo), iznos(faktura, pdvFaktura)))
	for i := range redovi {
		redovi[i].Novo, redovi[i].Faktura = robZaokruzi(redovi[i].Novo, 2), robZaokruzi(redovi[i].Faktura, 2)
		if math.Abs(redovi[i].Razlika) < 0.005 {
			redovi[i].Razlika = 0
		}
	}
	return redovi
}

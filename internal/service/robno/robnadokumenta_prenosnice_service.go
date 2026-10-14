package robno

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Prenosnice (the legacy rob_Prenosnice): the prenosnica izlaz (the group PRI, vrsta 114) moves the roba of
// the magacin of the nalog to another magacin. Its header is saved with the kontra dokument, the prenosnica
// ulaz (the group PRU, vrsta 113) in the nalog of the same broj of the magacin where the roba goes (made when it
// does not exist); every stavka of the izlaz is also a stavka of the ulaz. The prenosnica ulaz is not entered
// (only viewed and printed). The screen is the one of the opšti dokumenti (Pre).
// ---------------------------------------------------------------------------------------------------

const (
	// robPreKontraVrd is the vrsta of the kontra dokument (the legacy nKontraVRD).
	robPreKontraVrd = 113
	// robPreMsgNaZalihama is the error of an artikal without the stanje (the legacy "Selektovani artikal ne
	// postoji na zalihama...").
	robPreMsgNaZalihama = "Nema ga na zalihama."
	robPreMsgZalihe     = "Nema dovoljno zaliha."
)

// robPreTipoviMagacina are the tipovi of the magacini where the roba of a prenosnica goes (the legacy
// ROB_QRY_MAG with "G;V;D;S;L"; the search "magacini-prenosa" has the same list).
const robPreTipoviMagacina = "('G', 'V', 'D', 'S', 'L')"

// robIniVrednost is the value of a setting of the firm (initable of the kar of the session, the odeljak and the
// ključ, without a username; the legacy seek KARODELJAKKLJUCUSERNAME with ""), "" without it.
func (s *RobnoDokumentaResource) robIniVrednost(ctx context.Context, odeljak, kljuc string) (string, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return "", fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(trim(defvrednost), '') as key from initable
		where kar = $1 and odeljak = $2 and kljuc = $3 and coalesce(username, '') = '' order by initableid limit 1`, "",
		[]any{userSession.SelectedKar, odeljak, kljuc}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return "", err
	}
	return (*rows)[0].Key, nil
}

// robPrePana is the analitika of the magacin on the document (the legacy sPANA: the magacin with two digits).
func robPrePana(mag int) string {
	return fmt.Sprintf("%02d", mag)
}

// robPreDokument fills the header and the entry of a stavka of a prenosnica (the legacy PuniCOMBO,
// PuniMAGKONTO and the exit of the broj of rob_Prenosnice): of a saved document the magacin where the roba goes
// (rdok.magid1, not changed any more) and the konto zaliha of its kontra dokument; of a new one the first
// magacin of the tipovi of the prenosnica other than the magacin of the nalog, with its first konto. The cena of
// a stavka may be changed unless initable DOZMENJAJCENU / PRENOSNICA is "NE".
func (s *RobnoDokumentaResource) robPreDokument(ctx context.Context, res *domain.RobnoOpstiDokument) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	h := &res.Header
	if h.RdokID > 0 {
		rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(r.magid1, 0)::text as key,
				coalesce(concat_ws(' - ', m.mag::text, nullif(m.opis, '')), '') as c1,
				coalesce(trim(coalesce(k.pkto, k2.pkto)), '') as c2
			from rdok r
			left join magacini m on m.magaciniid = r.magid1
			left join rdok k on k.rdokid = r.vrdokid
			left join lateral (select x.pkto from rdok x where x.vrdokid = r.rdokid order by x.rdokid limit 1) k2 on true
			where r.rdokid = $1`, "", []any{h.RdokID}, "", "")
		if err != nil {
			return err
		}
		if rows != nil && len(*rows) > 0 {
			r := (*rows)[0]
			h.Magacin1, h.Magacin1Text, h.Konto1 = robNullID(r.Key), r.C1, r.C2
		}
		h.Magacin1Zakljucan = true
	} else {
		rows, err := s.unos.Tekst.GetAllCustom(ctx, `select m.magaciniid::text as key, coalesce(concat_ws(' - ', m.mag::text, nullif(m.opis, '')), '') as c1
			from magacini m where m.god = $1 and m.kar = $2 and m.magaciniid <> $3 and upper(trim(coalesce(m.tipmag, ''))) in `+robPreTipoviMagacina+`
			order by m.mag limit 1`, "", []any{userSession.SelectedGod, userSession.SelectedKar, h.MagaciniID}, "", "")
		if err != nil {
			return err
		}
		if rows != nil && len(*rows) > 0 {
			h.Magacin1, h.Magacin1Text = robNullID((*rows)[0].Key), (*rows)[0].C1
			if h.Konto1, err = s.robUslugeKontoMagacina(ctx, int(h.Magacin1.Int64)); err != nil {
				return err
			}
		}
	}
	if h.Konto1 != "" && h.Magacin1.Valid {
		naziv, err := s.robPreNazivKonta(ctx, int(h.Magacin1.Int64), h.Konto1)
		if err != nil {
			return err
		}
		h.Konto1Text = strings.Join(robNeprazni(h.Konto1, naziv), " - ")
	}

	st := &res.Stavka
	st.CenaNaziv, st.CenaZakljucana = "Cena", false
	doz, err := s.robIniVrednost(ctx, "DOZMENJAJCENU", "PRENOSNICA")
	if err != nil {
		return err
	}
	st.DozMenjajCenu = !strings.EqualFold(doz, "NE")
	return nil
}

// robPreNazivKonta is the naziv of a konto of the magacin (the fkpl of its magkonto).
func (s *RobnoDokumentaResource) robPreNazivKonta(ctx context.Context, magaciniID int, konto string) (string, error) {
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce((select f.naziv from fkpl f where f.idfkpl = k.idfkpl), '') as key
		from magkonto k where k.magaciniid = $1 and trim(k.konto) = $2 order by k.magkontoid limit 1`, "",
		[]any{magaciniID, strings.TrimSpace(konto)}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return "", err
	}
	return strings.TrimSpace((*rows)[0].Key), nil
}

// RobPreKontoMagacina returns the first konto of the magacin with its naziv (the konto zaliha where the roba of
// a prenosnica goes when its magacin is selected: the legacy ListSelectPlus of Combo_MAGKONTO).
func (s *RobnoDokumentaResource) RobPreKontoMagacina(ctx context.Context, magaciniID int) (konto, naziv string, err error) {
	if konto, err = s.robUslugeKontoMagacina(ctx, magaciniID); err != nil || konto == "" {
		return konto, "", err
	}
	naziv, err = s.robPreNazivKonta(ctx, magaciniID, konto)
	return konto, naziv, err
}

// robNeprazni are the values that are not empty.
func robNeprazni(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// robPreValidate validates the header of a prenosnica (the legacy ValidacijaRDOK of rob_Prenosnice): the
// datum dokumenta is required and not after the datum naloga; the magacin where the roba goes is required,
// exists and is not the magacin of the nalog; its konto zaliha is required and a konto of that magacin; the
// org. jedinica and the mesto troška of the document and of the roba are required (and exist, a mesto troška of
// its org. jedinica); the prenosnica ulaz is not entered; the broj dokumenta is a whole number.
func (s *RobnoDokumentaResource) robPreValidate(ctx context.Context, params domain.RobnoOpstiParams, grupa string) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if grupa == "PRU" {
		add("brdok", "Ova vrsta se formira automatski.")
		return fieldErrors
	}
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.After(rnal.Danal.Time) {
		add("datumfakturisanja", "Posle datuma naloga.")
	}
	// The magacin and the konto zaliha where the roba goes.
	magacin1, magacinOK := 0, false
	switch id, given, ok := robnoUnosID(params.Magacin1); {
	case !ok:
		add("magacin1", robnoUnosMsgNePostoji)
	case !given:
		add("magacin1", common.ErrMsgObavezanPodatak)
	default:
		if exists, err := s.robnoUnosPostoji(ctx, "select count(*) from magacini where magaciniid = $1", id); err != nil {
			add("magacin1", common.ErrMsgReadData)
		} else if !exists {
			add("magacin1", robnoUnosMsgNePostoji)
		} else if int(id) == params.MagaciniID {
			add("magacin1", "Isti magacin kao magacin naloga.")
		} else {
			magacin1, magacinOK = int(id), true
		}
	}
	if konto := strings.TrimSpace(params.Konto1); konto == "" || konto == "-" {
		add("konto1", common.ErrMsgObavezanPodatak)
	} else if magacinOK {
		if konta, err := s.GetRobKontaMagacina(ctx, magacin1); err != nil {
			add("konto1", common.ErrMsgReadData)
		} else if len(konta) == 0 {
			add("konto1", "Magacin nema konto.")
		} else if !robSadrzi(konta, konto) {
			add("konto1", "Nije konto magacina.")
		}
	}
	// The org. jedinice and the mesta troška are required.
	for _, f := range []struct{ value, field string }{
		{params.Orgjed, "idorgjed"}, {params.MestoTroska, "mestotrid"}, {params.Orgjed1, "idorgjed1"}, {params.MestoTroska1, "mestotrid1"},
	} {
		if _, given, ok := robnoUnosID(f.value); ok && !given {
			add(f.field, common.ErrMsgObavezanPodatak)
		}
	}
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)
	s.robOpstiOrgjedMt(ctx, params.Orgjed1, params.MestoTroska1, "idorgjed1", "mestotrid1", add)
	s.robOpstiBrojDokumenta(params.BrojDokum, add)
	return fieldErrors
}

// robPreSaveHeader saves the header of a prenosnica with its kontra dokument in one transaction (the legacy
// SnimiRDOK and ModifyRDOK of rob_Prenosnice):
//   - a new prenosnica (no document of the vrsta and broj in the magacin) gets the nalog of the same vrsta and
//     broj of the magacin where the roba goes (NoviNALOG: made when it does not exist), the document of the
//     nalog with the datumi, the opis, the izvorni dokument, the magacin where the roba goes (magid1), the
//     konto and the analitika of its magacin (pkto, pana) and the org. jedinice and mesta troška, and the kontra
//     dokument: the next broj of the vrsta 113 in the magacin where the roba goes, the nalog of that magacin,
//     the same datumi, opis and izvorni dokument, the magacin of the prenosnica (magid1) and its id (vrdokid),
//     the konto zaliha where the roba goes and its analitika (pkto, pana), the org. jedinice and mesta troška
//     swapped; the prenosnica gets the id of the kontra dokument (vrdokid);
//   - a saved one is changed with the datumi, the opis, the izvorni dokument and the org. jedinice and mesta
//     troška (not posted: knjige "N"), its kontra dokument the same (swapped) and the konto zaliha where the roba
//     goes; the stavke of both get the datum dokumenta.
//
// It returns the rdokid of the prenosnica.
func (s *RobnoDokumentaResource) robPreSaveHeader(ctx context.Context, params domain.RobnoOpstiParams) (rdokID int64, err error) {
	tx, err := s.unos.Dokument.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if rdokID, err = s.robPreSnimiZaglavlje(ctx, tx, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// robPreSnimiZaglavlje saves the header of a prenosnica with its kontra dokument in the transaction tx (see
// robPreSaveHeader).
func (s *RobnoDokumentaResource) robPreSnimiZaglavlje(ctx context.Context, tx db.Transaction, params domain.RobnoOpstiParams) (rdokID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	vrd, dokum := common.StringToInt(params.Vrd), common.StringToInt(params.BrojDokum)
	magacin1, konto1 := int(id(params.Magacin1)), strings.TrimSpace(params.Konto1)
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
	kontraPolja := []domain.Fields{
		robPoljeDatum("dadok", params.DatumDokum),
		robPoljeText("opis", params.Opis),
		robPoljeDatum("datiz", params.IzvorniDokDatum),
		robPoljeText("dokiz", params.IzvorniDokum),
		{Name: "rok", Value: "0"},
		robPoljeID("idorgjed", id(params.Orgjed1)),
		robPoljeID("mestotrid", id(params.MestoTroska1)),
		robPoljeID("idorgjed1", id(params.Orgjed)),
		robPoljeID("mestotrid1", id(params.MestoTroska)),
		robPoljeText("pkto", konto1),
	}
	nisuKnjizeni := []domain.Fields{{Name: "knjige_1", Value: "N"}, {Name: "knjige_2", Value: "N"}}
	konto, err := s.robUslugeKontoMagacina(ctx, params.MagaciniID)
	if err != nil {
		return 0, err
	}
	var vrdokID int64
	err = tx.QueryRowContext(ctx, `select rdokid, coalesce(vrdokid, 0)::bigint from rdok
		where god = $1 and kar = $2 and magaciniid = $3 and vrd = $4 and dokum = $5 for update`,
		userSession.SelectedGod, userSession.SelectedKar, params.MagaciniID, vrd, dokum).Scan(&rdokID, &vrdokID)
	switch {
	case err == sql.ErrNoRows:
		var mag, mag1 int
		if err = tx.QueryRowContext(ctx, `select coalesce((select mag from magacini where magaciniid = $1), 0),
				coalesce((select mag from magacini where magaciniid = $2), 0)`, params.MagaciniID, magacin1).Scan(&mag, &mag1); err != nil {
			return 0, err
		}
		var rnal1ID int64
		if rnal1ID, err = s.robPreNalog(ctx, tx, params.RnalID, magacin1, mag1, params.Opis); err != nil {
			return 0, err
		}
		if rdokID, err = s.robDodajDokument(ctx, tx, domain.RobniDokumentSnimanje{RnalID: params.RnalID, MagaciniID: params.MagaciniID,
			Vrd: vrd, Dokum: dokum, Polja: append(append([]domain.Fields{}, polja...),
				domain.Fields{Name: "magid1", Value: strconv.Itoa(magacin1)},
				domain.Fields{Name: "fkto", Value: ""}, domain.Fields{Name: "fana", Value: ""}, domain.Fields{Name: "pornapomena", Value: ""},
				domain.Fields{Name: "pkto", Value: konto}, domain.Fields{Name: "pana", Value: robPrePana(mag)})}); err != nil {
			return 0, err
		}
		var dokum1 int
		if err = tx.QueryRowContext(ctx, `select coalesce(max(dokum), 0) + 1 from rdok where god = $1 and kar = $2 and magaciniid = $3 and vrd = $4`,
			userSession.SelectedGod, userSession.SelectedKar, magacin1, robPreKontraVrd).Scan(&dokum1); err != nil {
			return 0, err
		}
		var kontraID int64
		if kontraID, err = s.robDodajDokument(ctx, tx, domain.RobniDokumentSnimanje{RnalID: rnal1ID, MagaciniID: magacin1,
			Vrd: robPreKontraVrd, Dokum: dokum1, Polja: append(append([]domain.Fields{}, kontraPolja...),
				domain.Fields{Name: "magid1", Value: strconv.Itoa(params.MagaciniID)},
				domain.Fields{Name: "vrdokid", Value: strconv.FormatInt(rdokID, 10)},
				domain.Fields{Name: "fkto", Value: ""}, domain.Fields{Name: "fana", Value: ""}, domain.Fields{Name: "pornapomena", Value: ""},
				domain.Fields{Name: "pana", Value: robPrePana(mag1)})}); err != nil {
			return 0, err
		}
		if err = s.unos.Dokument.CreateUpdateCustomWithTx(ctx, tx, `update rdok set vrdokid = $2 where rdokid = $1`, rdokID, kontraID); err != nil {
			return 0, err
		}
	case err != nil:
		return 0, err
	default:
		if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", rdokID, append(polja, nisuKnjizeni...)); err != nil {
			return 0, err
		}
		if vrdokID == 0 {
			if err = tx.QueryRowContext(ctx, `select coalesce((select rdokid from rdok where vrdokid = $1 order by rdokid limit 1), 0)::bigint`,
				rdokID).Scan(&vrdokID); err != nil {
				return 0, err
			}
		}
		ids := []int64{rdokID}
		if vrdokID > 0 {
			if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", vrdokID, append(kontraPolja, nisuKnjizeni...)); err != nil {
				return 0, err
			}
			ids = append(ids, vrdokID)
		}
		for _, dokID := range ids {
			if err = s.unos.Dokument.CreateUpdateCustomWithTx(ctx, tx, `update rpro set dadok = $1 where rdokid = $2`,
				robPoljeVrednost(polja[0]), dokID); err != nil {
				return 0, err
			}
		}
	}
	return rdokID, nil
}

// robPreNalog returns the nalog of the vrsta and broj of the nalog rnalID in the magacin where the roba goes,
// adding it when it does not exist (the legacy NoviNALOG): its vrsta, broj and datumi, the opis of the
// document, the magacin, the counters at 0 and the operater.
func (s *RobnoDokumentaResource) robPreNalog(ctx context.Context, tx db.Transaction, rnalID int64, magacin1, mag1 int, opis string) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	var tipdok string
	var nalog int
	var idtipdok sql.NullInt64
	var danal, datob sql.NullTime
	if err := tx.QueryRowContext(ctx, `select coalesce(tipdok, ''), coalesce(nalog, 0), idtipdok, danal, datob from rnal where rnalid = $1`,
		rnalID).Scan(&tipdok, &nalog, &idtipdok, &danal, &datob); err != nil {
		return 0, err
	}
	var rnal1ID int64
	err := tx.QueryRowContext(ctx, `select rnalid from rnal where god = $1 and kar = $2 and nalog = $3 and tipdok = $4 and magaciniid = $5
		order by rnalid limit 1`, userSession.SelectedGod, userSession.SelectedKar, nalog, tipdok, magacin1).Scan(&rnal1ID)
	if err == nil {
		return rnal1ID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	idtip := domain.Fields{Name: "idtipdok", IsNull: true}
	if idtipdok.Valid {
		idtip = domain.Fields{Name: "idtipdok", Value: strconv.FormatInt(idtipdok.Int64, 10)}
	}
	return s.rnalHeaderRepo.CreateWithTx(ctx, tx, &domain.Rnal{}, common.IDrnal, []domain.Fields{
		{Name: "tipdok", Value: tipdok},
		idtip,
		{Name: "nalog", Value: strconv.Itoa(nalog)},
		robPoljeNullTime("danal", danal),
		robPoljeNullTime("datob", datob),
		robPoljeText("opis", opis),
		{Name: "magaciniid", Value: strconv.Itoa(magacin1)},
		{Name: "mag", Value: strconv.Itoa(mag1)},
		{Name: "rbr", Value: "0"},
		{Name: "dug", Value: "0"},
		{Name: "pot", Value: "0"},
		{Name: "brdo", Value: "0"},
		{Name: "brst", Value: "0"},
		{Name: "oper", Value: userSession.UserName},
		{Name: "nalsts", Value: ""},
		{Name: "pinalid", IsNull: true},
	})
}

// robPreKontra returns the kontra dokument of a prenosnica (its vrdokid, else the document whose vrdokid is the
// prenosnica) with the konto zaliha where the roba goes (its pkto); 0 without it.
func (s *RobnoDokumentaResource) robPreKontra(ctx context.Context, rdokID int64) (kontraID int64, konto1 string, err error) {
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select k.rdokid::text as key, coalesce(trim(k.pkto), '') as c1 from rdok r
		join rdok k on k.rdokid = coalesce(nullif(r.vrdokid, 0), (select x.rdokid from rdok x where x.vrdokid = r.rdokid order by x.rdokid limit 1))
		where r.rdokid = $1`, "", []any{rdokID}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return 0, "", err
	}
	return int64(common.StringToInt((*rows)[0].Key)), (*rows)[0].C1, nil
}

// robPreArtikal returns the artikal of a stavka of a prenosnica (the legacy exit of EDT_SIFRA of rob_Prenosnice):
// the artikal by its šifra with its naziv and JM (the konto of the form); its stanje, cena and nabavna cena in
// the magacin of the nalog: in a magacin of the tip zaliha 2 the ones of the lot of the form (drsta), else of the
// konto (rsta). The cena is the prosečna nabavna cena in a magacin of the nabavna cena, else the cena of the
// stanje; the nabavna cena the prosečna nabavna cena, its nabavna cena when that is 0; all 0 without the stanje.
// found is false for an artikal that does not exist.
func (s *RobnoDokumentaResource) robPreArtikal(ctx context.Context, k domain.RobStavkeKontekst, sifra, konto, otk, serija, rok string) (domain.RobOpstiArtikal, bool, error) {
	a := domain.RobOpstiArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select a.sifra::text as key, coalesce(a.naziv, '') as c1, coalesce(a.jm, '') as c2
		from rsif a where a.god = $1 and a.kar = $2 and a.sifra::text = $3 limit 1`, "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, strings.TrimSpace(sifra)}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm, a.Konto = common.StringToInt(r.Key), r.C1, r.C2, strings.TrimSpace(konto)
	var ulaz, izlaz, cena, prosnc, ncena float64
	found := false
	if k.Tipzal == 2 {
		datum, _ := robnoUnosDatum(rok)
		lot, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(ulaz, 0)::float8::text as key, coalesce(izlaz, 0)::float8::text as c1,
				coalesce(cena, 0)::float8::text as c2, coalesce(prosnc, 0)::float8::text as c3, coalesce(ncena, 0)::float8::text as c4
			from drsta where god = $1 and kar = $2 and konto = $3 and sifra = $4 and mag = $5 and otk = $6 and serija = $7 and roktr = $8
			order by rstaid limit 1`, "", []any{userSession.SelectedGod, userSession.SelectedKar, a.Konto, a.Sifra, k.Mag,
			strings.TrimSpace(otk), strings.TrimSpace(serija), RobWinDevDatum(datum)}, "", "")
		if err != nil {
			return a, true, err
		}
		if lot != nil && len(*lot) > 0 {
			l := (*lot)[0]
			f := common.StringToFloat64
			ulaz, izlaz, cena, prosnc, ncena, found = f(l.Key), f(l.C1), f(l.C2), f(l.C3), f(l.C4), true
		}
	} else {
		st, err := s.robOpstiStanje(ctx, k.MagaciniID, a.Konto, a.Sifra)
		if err != nil {
			return a, true, err
		}
		ulaz, izlaz, cena, prosnc, ncena, found = st.ulaz, st.izlaz, st.cena, st.prosnc, st.ncena, st.found
	}
	if found {
		a.Stanje = ulaz - izlaz
		a.Cena = cena
		if k.NabavnaCena() {
			a.Cena = prosnc
		}
		a.Ncena = prosnc
		if a.Ncena == 0 {
			a.Ncena = ncena
		}
	}
	return a, true, nil
}

// robPreValidateStavka validates a stavka of a prenosnica before it is saved (the legacy Validacija of
// rob_Prenosnice): the konto is required and a konto of the magacin; the kontra dokument has the konto zaliha
// where the roba goes; the šifra is required and the artikal exists; the količina, the cena and the nabavna cena
// are greater than 0, in a magacin of the nabavna cena the cena is the nabavna cena; the artikal is on the
// zalihe of the magacin (rsta in a magacin of the tip zaliha 1, the lot of the stavka in drsta). A document that
// is not a storno has the količina in the stanje (the legacy questions "Nema dovoljno zaliha. Da li ste sigurni
// da izdajete robu?" are the error of the field; the one of the lot only when initable DOZZALMINUS /
// PRENOSNICA is empty or "DA").
func (s *RobnoDokumentaResource) robPreValidateStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		for _, fe := range fieldErrors {
			if fe.Field == field {
				return
			}
		}
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
	if _, konto1, err := s.robPreKontra(ctx, k.RdokID); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if konto1 == "" {
		add("kontostavke", "Nema konto magacina gde ide roba.")
	}
	sifra, sifraOK := 0, false
	if sifraText := strings.TrimSpace(params.Sifra); sifraText == "" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if v, err := strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robnoUnosPostoji(ctx, `select count(*) from rsif where god = $1 and kar = $2 and sifra = $3`,
		userSession.SelectedGod, userSession.SelectedKar, v); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else {
		sifra, sifraOK = v, true
	}
	kolicina := common.StringToFloat64(params.Kolic)
	if strings.TrimSpace(params.Kolic) == "" {
		add("kolicina", common.ErrMsgObavezanPodatak)
	} else if kolicina <= 0 {
		add("kolicina", robInterniMsgVeceOdNule)
	}
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	if cena <= 0 {
		add("cenaopd", robInterniMsgVeceOdNule)
	}
	if ncena <= 0 {
		add("nabavnacena", robInterniMsgVeceOdNule)
	}
	if k.NabavnaCena() && ncena != cena {
		add("cenaopd", "Mora biti jednaka nabavnoj ceni.")
	}
	if !sifraOK || konto == "" {
		return fieldErrors
	}
	// The zalihe: the stanje of the konto (tip zaliha 1) and the lot (drsta).
	if k.Tipzal == 1 {
		st, err := s.robOpstiStanje(ctx, k.MagaciniID, konto, sifra)
		if err != nil {
			add("sifraartikla", common.ErrMsgReadData)
			return fieldErrors
		}
		if !st.found {
			add("sifraartikla", robPreMsgNaZalihama)
		} else if stanje := robZaokruzi(st.ulaz-st.izlaz, 3); !k.Storno() && ((kolicina > 0 && stanje <= 0) || robZaokruzi(stanje-kolicina, 3) < 0) {
			add("kolicina", robPreMsgZalihe)
		}
	}
	datum, _ := robnoUnosDatum(params.Rok)
	stanje, lot, err := s.robOpstiDrsta(ctx, k.Mag, konto, sifra, params.Otk, params.Serija, RobWinDevDatum(datum))
	if err != nil {
		add("sifraartikla", common.ErrMsgReadData)
		return fieldErrors
	}
	if !lot {
		add("sifraartikla", robPreMsgNaZalihama)
		return fieldErrors
	}
	doz, err := s.robIniVrednost(ctx, "DOZZALMINUS", "PRENOSNICA")
	if err != nil {
		add("kolicina", common.ErrMsgReadData)
		return fieldErrors
	}
	if (doz == "" || strings.EqualFold(doz, "DA")) && !k.Storno() && robZaokruzi(stanje, 3) < kolicina {
		add("kolicina", robPreMsgZalihe)
	}
	return fieldErrors
}

// robPreIznos is the iznos and the cene of a stavka of a prenosnica in a magacin (the legacy CreateRpro): in a
// magacin of the sirovine (tipmag "S") the iznos is by the nabavna cena and the cena of the stavka is the
// nabavna cena (its nabavna cena the cena), else by the cena; the količina and the iznos are negative for a
// storno.
func robPreIznos(tipmag string, storno bool, kolicina, cena, ncena float64) (kol, iznos, cenaStavke, ncenaStavke, xcena float64) {
	kol, iznos, cenaStavke, ncenaStavke, xcena = kolicina, kolicina*cena, cena, ncena, cena
	if strings.EqualFold(strings.TrimSpace(tipmag), "S") {
		iznos, cenaStavke, ncenaStavke, xcena = kolicina*ncena, ncena, cena, ncena
	}
	kol, iznos = math.Abs(kol), math.Abs(robZaokruzi(iznos, 2))
	if storno {
		kol, iznos = -kol, -iznos
	}
	return kol, iznos, cenaStavke, ncenaStavke, xcena
}

// robPreStavka adds a stavka to a prenosnica and to its kontra dokument in one transaction (the legacy CreateRpro
// of rob_Prenosnice):
//   - the stavka of the prenosnica (robPreIznos by the tip of its magacin, the predznak of its vrsta) with the
//     columns of the document, the artikal, the konto, the cene, the lot and the next redni broj; the nalog, the
//     stanje and the zalihe by lot of the magacin get it (AzurirajRNAL, AzurirajRSTA and AzurirajDRSTA, without
//     the cena), the document its iznos and a stavka more;
//   - the stavka of the kontra dokument (by the tip of the magacin where the roba goes and the predznak of the
//     vrsta 113) with the same artikal, redni broj and lot, the konto zaliha where the roba goes and the id of the
//     stavka of the prenosnica (rproid1); the nalog, the stanje (with the cena) and the zalihe by lot of that
//     magacin get it, the kontra dokument its iznos and a stavka more; the stanje of the artikal and its zalihe
//     by lot in that magacin get the new prosečna nabavna cena (of the stanje before and of the stavka by the
//     cena of the way the magacin keeps its zalihe: 1 the cena, 3 the nabavna cena).
//
// The SR, RSAL and RSIN are not kept.
func (s *RobnoDokumentaResource) robPreStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	kontraID, konto1, err := s.robPreKontra(ctx, k.RdokID)
	if err != nil {
		return 0, err
	}
	if kontraID == 0 {
		return 0, fmt.Errorf("kontra dokument prenosnice ne postoji")
	}
	kk, found, err := s.GetRobStavkeKontekst(ctx, kontraID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("kontra dokument prenosnice ne postoji")
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
	if rproID, err = s.robPreDodajStavku(ctx, tx, k, kk, konto1, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robPreDodajStavku adds a stavka to the prenosnica of k and to its kontra dokument kk (with the konto zaliha
// konto1) in the transaction tx (see robPreStavka).
func (s *RobnoDokumentaResource) robPreDodajStavku(ctx context.Context, tx db.Transaction, k, kk domain.RobStavkeKontekst, konto1 string,
	params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	god, kar := userSession.SelectedGod, userSession.SelectedKar
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	kolicina := common.StringToFloat64(params.Kolic)
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	rok, _ := robnoUnosDatum(params.Rok)
	roktr := RobWinDevDatum(rok)
	otk, serija := strings.TrimSpace(params.Otk), strings.TrimSpace(params.Serija)
	var rsifID int64
	var po int
	if err = tx.QueryRowContext(ctx, `select rsifid, coalesce(po, 0) from rsif where god = $1 and kar = $2 and sifra = $3 limit 1`,
		god, kar, sifra).Scan(&rsifID, &po); err != nil {
		return 0, err
	}
	var rbr int
	if err = tx.QueryRowContext(ctx, `select coalesce(max(rbr), 0) + 1 from rpro where rnalid = $1 and rdokid = $2`, k.RnalID, k.RdokID).
		Scan(&rbr); err != nil {
		return 0, err
	}
	// The stavka of a document of the prenosnica.
	stavka := func(d domain.RobStavkeKontekst, konto string, kol, iznos, cenaStavke, ncenaStavke float64, rproID1 int64) (int64, error) {
		var pkto, pana string
		if err := tx.QueryRowContext(ctx, `select coalesce(pkto, ''), coalesce(pana, '') from rdok where rdokid = $1`, d.RdokID).
			Scan(&pkto, &pana); err != nil {
			return 0, err
		}
		polja := []domain.Fields{
			{Name: "rnalid", Value: strconv.FormatInt(d.RnalID, 10)},
			{Name: "nalog", Value: strconv.Itoa(d.Nalog)},
			{Name: "rdokid", Value: strconv.FormatInt(d.RdokID, 10)},
			{Name: "vrd", Value: strconv.Itoa(d.Vrd)},
			{Name: "dokum", Value: strconv.Itoa(d.Dokum)},
			robPoljeNullTime("dadok", d.Dadok),
			{Name: "sifra", Value: strconv.Itoa(sifra)},
			{Name: "konto", Value: konto},
			robPoljeFloat("kolic", kol),
			robPoljeFloat("iznos", iznos),
			robPoljeFloat("cena", cenaStavke),
			robPoljeFloat("ncena", ncenaStavke),
			{Name: "fkto", Value: d.Fkto},
			{Name: "fana", Value: d.Fana},
			robPoljeText("pkto", pkto),
			robPoljeText("pana", pana),
			{Name: "rab", Value: "0"},
			{Name: "po", Value: strconv.Itoa(po)},
			{Name: "mma", Value: "0"},
			{Name: "mag", Value: strconv.Itoa(d.Mag)},
			{Name: "magaciniid", Value: strconv.Itoa(d.MagaciniID)},
			{Name: "brst", Value: "1"},
			{Name: "rbr", Value: strconv.Itoa(rbr)},
			{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
			{Name: "otk", Value: otk},
			{Name: "serija", Value: serija},
			{Name: "roktr", Value: strconv.FormatInt(roktr, 10)},
			// NOT NULL without a default.
			{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
		}
		if rproID1 > 0 {
			polja = append(polja, domain.Fields{Name: "rproid1", Value: strconv.FormatInt(rproID1, 10)})
		}
		return s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja)
	}
	// Its posting: the nalog, the stanje and the zalihe by lot of the magacin, the document.
	knjizi := func(d domain.RobStavkeKontekst, konto string, kol, iznos, cenaStanja float64) error {
		if err := s.AzurirajRNAL(ctx, tx, d.RnalID, d.Dugovno(), d.Potrazno(), iznos); err != nil {
			return err
		}
		p := domain.RobPromenaStanja{MagaciniID: d.MagaciniID, Konto: konto, Sifra: sifra, Kolicina: kol, Iznos: iznos, Cena: cenaStanja,
			Danal: d.Danal.Time, DokvrstaID: d.DokvrstaID, Tipdok: d.Tipdok, Tz: "-", Tzi: "-", Brst: 1}
		if err := s.AzurirajRSTA(ctx, tx, p); err != nil {
			return err
		}
		p.Otk, p.Serija, p.Roktr = otk, serija, roktr
		if err := s.AzurirajDRSTA(ctx, tx, p); err != nil {
			return err
		}
		return s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) + $2)::numeric, 2),
			brst = coalesce(brst, 0) + 1 where rdokid = $1`, d.RdokID, iznos)
	}

	// The prenosnica.
	kol, iznos, cenaStavke, ncenaStavke, _ := robPreIznos(k.Tipmag, k.Storno(), kolicina, cena, ncena)
	if rproID, err = stavka(k, konto, kol, iznos, cenaStavke, ncenaStavke, 0); err != nil {
		return 0, err
	}
	if err = knjizi(k, konto, kol, iznos, 0); err != nil {
		return 0, err
	}

	// The kontra dokument, with the stanje of the magacin where the roba goes before it.
	kol1, iznos1, cena1, ncena1, xcena := robPreIznos(kk.Tipmag, kk.Storno(), kolicina, cena, ncena)
	if _, err = stavka(kk, konto1, kol1, iznos1, cena1, ncena1, rproID); err != nil {
		return 0, err
	}
	var rstaUlaz, rstaIzlaz, rstaProsnc float64
	err = tx.QueryRowContext(ctx, `select round(coalesce(ulaz, 0), 3)::float8, round(coalesce(izlaz, 0), 3)::float8, coalesce(prosnc, 0)::float8
		from rsta where god = $1 and kar = $2 and magaciniid = $3 and konto = $4 and sifra = $5 order by rstaid limit 1`,
		god, kar, kk.MagaciniID, konto1, sifra).Scan(&rstaUlaz, &rstaIzlaz, &rstaProsnc)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	err = nil
	if err = knjizi(kk, konto1, kol1, iznos1, xcena); err != nil {
		return 0, err
	}
	// The prosečna nabavna cena of the stanje of the magacin where the roba goes.
	switch kk.Nacvodzal {
	case 1:
		xcena = cena
	case 3:
		xcena = ncena
	}
	stanje := rstaUlaz - rstaIzlaz
	prosnc := xcena
	if kol1+stanje != 0 {
		prosnc = robZaokruzi((kol1*xcena+stanje*rstaProsnc)/(kol1+stanje), 2)
	}
	if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rsta set prosnc = $6
		where god = $1 and kar = $2 and magaciniid = $3 and konto = $4 and sifra = $5`, god, kar, kk.MagaciniID, konto1, sifra, prosnc); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update drsta set prosnc = $5 where god = $1 and kar = $2 and mag = $3 and rsifid = $4`,
		god, kar, kk.Mag, rsifID, prosnc); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robPreBrisi deletes a stavka of a prenosnica and its stavka of the kontra dokument (rproid1) in the transaction
// tx (the legacy BrisiRPRO of rob_Prenosnice): each is taken back from the stanje and the zalihe by lot of its
// magacin (AzurirajRSTA and AzurirajDRSTA), from its document (the iznos, a stavka less) and from its nalog
// (AzurirajRNAL, after the deletion), it is deleted and the stavke after it in its document get the redni broj
// less by one.
func (s *RobnoDokumentaResource) robPreBrisi(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64) error {
	kontraID, _, err := s.robPreKontra(ctx, k.RdokID)
	if err != nil {
		return err
	}
	kk := domain.RobStavkeKontekst{}
	if kontraID > 0 {
		var found bool
		if kk, found, err = s.GetRobStavkeKontekst(ctx, kontraID); err != nil {
			return err
		} else if !found {
			kontraID = 0
		}
	}
	return s.robPreObrisiStavku(ctx, tx, k, kk, kontraID, rproID)
}

// robPreObrisiStavku deletes the stavka rproID of the prenosnica of k and its stavka of the kontra dokument
// kontraID (kk; 0 without it) in the transaction tx (see robPreBrisi).
func (s *RobnoDokumentaResource) robPreObrisiStavku(ctx context.Context, tx db.Transaction, k, kk domain.RobStavkeKontekst, kontraID, rproID int64) error {
	brisi := func(d domain.RobStavkeKontekst, rproID int64) error {
		var sifra, rbr int
		var konto, otk, serija string
		var kolicina, iznos, roktr float64
		if err := tx.QueryRowContext(ctx, `select coalesce(sifra, 0), coalesce(rbr, 0), coalesce(konto, ''), coalesce(otk, ''), coalesce(serija, ''),
				coalesce(kolic, 0)::float8, coalesce(iznos, 0)::float8, coalesce(roktr, 0)::float8
			from rpro where rproid = $1 for update`, rproID).Scan(&sifra, &rbr, &konto, &otk, &serija, &kolicina, &iznos, &roktr); err != nil {
			return err
		}
		p := domain.RobPromenaStanja{MagaciniID: d.MagaciniID, Konto: strings.TrimSpace(konto), Sifra: sifra, Kolicina: -kolicina, Iznos: -iznos,
			Danal: d.Danal.Time, DokvrstaID: d.DokvrstaID, Tipdok: d.Tipdok, Tz: "-", Tzi: "-", Brst: -1}
		if err := s.AzurirajRSTA(ctx, tx, p); err != nil {
			return err
		}
		p.Otk, p.Serija, p.Roktr = strings.TrimSpace(otk), strings.TrimSpace(serija), int64(roktr)
		if err := s.AzurirajDRSTA(ctx, tx, p); err != nil {
			return err
		}
		if err := s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) - $2)::numeric, 2),
			brst = greatest(coalesce(brst, 0) - 1, 0) where rdokid = $1`, d.RdokID, iznos); err != nil {
			return err
		}
		if err := s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
			return err
		}
		if err := s.AzurirajRNAL(ctx, tx, d.RnalID, d.Dugovno(), d.Potrazno(), -iznos); err != nil {
			return err
		}
		return s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rpro set rbr = rbr - 1 where rdokid = $1 and rbr > $2`, d.RdokID, rbr)
	}
	if err := brisi(k, rproID); err != nil {
		return err
	}
	// The stavka of the kontra dokument.
	if kontraID == 0 {
		return nil
	}
	var kontraRproID int64
	if err := tx.QueryRowContext(ctx, `select coalesce((select rproid from rpro where rdokid = $1 and rproid1 = $2 order by rproid limit 1), 0)::bigint`,
		kontraID, rproID).Scan(&kontraRproID); err != nil {
		return err
	}
	if kontraRproID == 0 {
		return nil
	}
	return brisi(kk, kontraRproID)
}

// RobPreStavkeTableRows returns the stavke of a prenosnica as the rows of the grid of the stavke (the columns of
// GetOpstiPreStavkeTableFields: the ones of the opšti dokument with the nabavna cena after the iznos).
func RobPreStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := RobOpstiStavkeTableRows(stavke)
	for i := range rows {
		f := rows[i].Fields
		rows[i].Fields = append(append(append([]string{}, f[:8]...), robFormatIznos(stavke[i].Ncena)), f[8:]...)
	}
	return rows
}

package robno

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Nivelacije veleprodaje (the group NIV, the legacy rob_Nivelacije_Veleprodaja): an opšti dokument whose
// stavke change the veleprodajna cena of the artikli of a magacin whose zalihe are kept at the prodajna
// cena; the vrednost of a stavka is its količina by the difference of the nova and the stara cena
// ---------------------------------------------------------------------------------------------------

// robNivMsgProdajnaCena is the message of a magacin whose zalihe are not kept at the prodajna cena (the legacy
// PodaciZAGLAVLJE of rob_Nivelacije_Veleprodaja).
const robNivMsgProdajnaCena = "Zalihe se ne vode po prodajnoj ceni!"

// robNivMsgBrisanje is the warning of the deletion of a stavka of a nivelacija (the legacy BTN_Brisanje).
const robNivMsgBrisanje = "Ukoliko želite da cena bude kao pre nivelacije, pre brisanja nivelišite artikal na staru cenu pa brišite obe stavke!"

// RobNivMsgBrisanje returns the warning of the deletion of a stavka of a nivelacija veleprodaje.
func RobNivMsgBrisanje() string {
	return robNivMsgBrisanje
}

// robNivProdajnaCena tells whether the zalihe of the magacin are kept at the prodajna cena (magacini.nacvodzal 1).
func (s *RobnoDokumentaResource) robNivProdajnaCena(ctx context.Context, magaciniID int) (bool, error) {
	return s.robnoUnosPostoji(ctx, `select count(*) from magacini where magaciniid = $1 and coalesce(nacvodzal, 0) = 1`, magaciniID)
}

// robNivValidate validates the header of a nivelacija veleprodaje (the legacy ValidacijaRDOK of
// rob_Nivelacije_Veleprodaja): the datum dokumenta is required and not before the datum naloga, the broj and
// the datum of the izvorni dokument are required, a given org. jedinica and mesto troška exist (not required:
// empty or "-" is saved as NULL); the magacin keeps its zalihe at the prodajna cena (the legacy message of
// PodaciZAGLAVLJE); the broj dokumenta is a whole number.
func (s *RobnoDokumentaResource) robNivValidate(ctx context.Context, params domain.RobnoOpstiParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if dadok, ok := robnoUnosDatum(params.DatumDokum); !ok {
		add("datumfakturisanja", robnoUnosDatumMsg(params.DatumDokum))
	} else if rnal, found, err := s.robnoUnosNalog(ctx, params.RnalID); err != nil {
		add("datumfakturisanja", common.ErrMsgReadData)
	} else if found && rnal.Danal.Valid && dadok.Before(rnal.Danal.Time) {
		add("datumfakturisanja", "Pre datuma naloga.")
	}
	if strings.TrimSpace(params.IzvorniDokum) == "" {
		add("brdokorg", common.ErrMsgObavezanPodatak)
	}
	if _, ok := robnoUnosDatum(params.IzvorniDokDatum); !ok {
		add("datumizvora", robnoUnosDatumMsg(params.IzvorniDokDatum))
	}
	// The org. jedinica and the mesto troška are not required (empty or "-": NULL).
	s.robOpstiOrgjedMt(ctx, params.Orgjed, params.MestoTroska, "idorgjed", "mestotrid", add)
	if ok, err := s.robNivProdajnaCena(ctx, params.MagaciniID); err != nil {
		add("brdok", common.ErrMsgReadData)
	} else if !ok {
		add("brdok", robNivMsgProdajnaCena)
	}
	s.robOpstiBrojDokumenta(params.BrojDokum, add)
	return fieldErrors
}

// robNivSaveHeader saves the header of a nivelacija veleprodaje (the legacy NoviSLOGRDOK of
// rob_Nivelacije_Veleprodaja): a new document of the nalog (SaveRobniDokument) or the existing one of the vrsta
// and broj with the izvorni dokument, the opis, the org. jedinica and the mesto troška, no fakturni konto,
// otpremnica, kamata, plaćanje, footer or rok; the datum dokumenta is set only while the document has no stavke
// (an existing one without stavke has its iznos, redni broj and broj stavki set to 0).
func (s *RobnoDokumentaResource) robNivSaveHeader(ctx context.Context, params domain.RobnoOpstiParams) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	id := func(value string) int64 {
		v, _, _ := robnoUnosID(value)
		return v
	}
	vrd, dokum := common.StringToInt(params.Vrd), common.StringToInt(params.BrojDokum)
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select r.rdokid::text as key, (select count(*) from rpro p where p.rdokid = r.rdokid)::text as c1
		from rdok r where r.god = $1 and r.kar = $2 and r.magaciniid = $3 and r.vrd = $4 and r.dokum = $5 limit 1`, "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, params.MagaciniID, vrd, dokum}, "", "")
	if err != nil {
		return 0, err
	}
	postoji := rows != nil && len(*rows) > 0
	saStavkama := postoji && common.StringToInt((*rows)[0].C1) > 0
	polja := []domain.Fields{
		robPoljeText("opis", params.Opis),
		robPoljeDatum("datiz", params.IzvorniDokDatum),
		robPoljeText("dokiz", params.IzvorniDokum),
		{Name: "rok", Value: "0"},
		{Name: "fkto", Value: ""},
		{Name: "fana", Value: ""},
		{Name: "otp", Value: ""},
		{Name: "kamt", Value: ""},
		{Name: "pla", Value: ""},
		{Name: "foot", Value: ""},
		robPoljeID("idorgjed", id(params.Orgjed)),
		robPoljeID("mestotrid", id(params.MestoTroska)),
	}
	if !saStavkama {
		polja = append(polja, robPoljeDatum("dadok", params.DatumDokum))
		if postoji {
			polja = append(polja, domain.Fields{Name: "iznos", Value: "0"}, domain.Fields{Name: "rbr", Value: "0"}, domain.Fields{Name: "brst", Value: "0"})
		}
	}
	return s.SaveRobniDokument(ctx, domain.RobniDokumentSnimanje{RnalID: params.RnalID, MagaciniID: params.MagaciniID, Vrd: vrd, Dokum: dokum, Polja: polja})
}

// robNivRsta reads the stanje and the veleprodajna cena of the artikal in the konto of the magacin (rsta);
// found is false without it.
func (s *RobnoDokumentaResource) robNivRsta(ctx context.Context, magaciniID int, konto string, sifra int) (stanje, vpcena float64, found bool, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, 0, false, fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8::text as key, coalesce(vpcena, 0)::float8::text as c1
		from rsta where god = $1 and kar = $2 and magaciniid = $3 and konto = $4 and sifra = $5 order by rstaid limit 1`, "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, magaciniID, strings.TrimSpace(konto), sifra}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return 0, 0, false, err
	}
	return common.StringToFloat64((*rows)[0].Key), common.StringToFloat64((*rows)[0].C1), true, nil
}

// robNivArtikal returns the artikal of a stavka of a nivelacija veleprodaje (the legacy exit of EDT_SIFRA of
// rob_Nivelacije_Veleprodaja): its naziv and JM, and from the stanje in the konto of the form the količina and
// the stara cena (the veleprodajna cena of the stanje): in a magacin of the tip zaliha 1 the stanje of the
// konto; of the tip zaliha 2 the sum of the zalihe by lot of the artikal, when it equals the stanje of the
// konto (else the stara cena is 0 and Poruka asks to check the zalihe). found is false for an artikal that
// does not exist.
func (s *RobnoDokumentaResource) robNivArtikal(ctx context.Context, k domain.RobStavkeKontekst, sifraText, konto string) (domain.RobOpstiArtikal, bool, error) {
	a := domain.RobOpstiArtikal{Konto: strings.TrimSpace(konto)}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	sifra, err := strconv.Atoi(strings.TrimSpace(sifraText))
	if err != nil {
		return a, false, nil
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select a.rsifid::text as key, coalesce(a.naziv, '') as c1, coalesce(a.jm, '') as c2
		from rsif a where a.god = $1 and a.kar = $2 and a.sifra = $3 limit 1`, "", []any{userSession.SelectedGod, userSession.SelectedKar, sifra}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm = sifra, r.C1, r.C2
	stanje, vpcena, found, err := s.robNivRsta(ctx, k.MagaciniID, a.Konto, sifra)
	if err != nil {
		return a, true, err
	}
	switch k.Tipzal {
	case 1:
		if found {
			a.Kolicina, a.StaraCena = stanje, vpcena
		}
	case 2:
		// The zalihe by lot of the artikal (the legacy FOR EACH DRSTA by its rsifid).
		lotovi, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(sum(coalesce(ulaz, 0) - coalesce(izlaz, 0)), 0)::float8::text as key
			from drsta where rsifid = $1 and coalesce(ulaz, 0) - coalesce(izlaz, 0) > 0`, "", []any{common.StringToInt(r.Key)}, "", "")
		if err != nil {
			return a, true, err
		}
		dstanje := 0.0
		if lotovi != nil && len(*lotovi) > 0 {
			dstanje = common.StringToFloat64((*lotovi)[0].Key)
		}
		if found {
			if robZaokruzi(stanje-dstanje, 3) == 0 {
				a.Kolicina, a.StaraCena = dstanje, vpcena
			} else {
				a.Poruka = fmt.Sprintf("Proverite ukupne i pojedinačne zalihe za šifru: %d %s", sifra, a.Naziv)
			}
		}
	}
	return a, true, nil
}

// robNivValidateStavka validates a stavka of a nivelacija veleprodaje (the legacy ValidacijaRPRO of
// rob_Nivelacije_Veleprodaja): the artikal exists, the količina, the nova and the stara cena are greater than 0,
// the artikal has a stanje in the konto of the magacin and it is greater than 0.
func (s *RobnoDokumentaResource) robNivValidateStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("sifraartikla", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	sifra, err := strconv.Atoi(strings.TrimSpace(params.Sifra))
	if err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robnoUnosPostoji(ctx, `select count(*) from rsif where god = $1 and kar = $2 and sifra = $3`,
		userSession.SelectedGod, userSession.SelectedKar, sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	}
	if common.StringToFloat64(params.Kolic) <= 0 {
		add("kolicina", robInterniMsgVeceOdNule)
	}
	if common.StringToFloat64(params.NovaCena) <= 0 {
		add("novacena", robInterniMsgVeceOdNule)
	}
	if common.StringToFloat64(params.StaraCena) <= 0 {
		add("staracena", robInterniMsgVeceOdNule)
	}
	if stanje, _, found, err := s.robNivRsta(ctx, k.MagaciniID, strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", ""), sifra); err != nil {
		add("kolicina", common.ErrMsgReadData)
	} else if !found {
		add("kolicina", "Artikal ne postoji na zalihama.")
	} else if stanje <= 0 {
		add("kolicina", "Količina na zalihama je 0 ili negativna.")
	}
	return fieldErrors
}

// robNivLot is the količina and the lot of a stavka of a nivelacija veleprodaje.
type robNivLot struct {
	kolicina    float64
	otk, serija string
	roktr       int64
}

// robNivStavka saves a stavka of a nivelacija veleprodaje (the legacy "Sačuvaj" of the stavke of
// rob_Nivelacije_Veleprodaja) in one transaction: in a magacin of the tip zaliha 2 one stavka per lot of the
// artikal in the magacin (with its stanje as the količina; the legacy FOR EACH DRSTA), else one with the
// količina and the lot of the form; see robNivSlog. It returns the id of the last stavka added (0 without one).
func (s *RobnoDokumentaResource) robNivStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select rsifid::text as key, coalesce(po, 0)::text as c1 from rsif where god = $1 and kar = $2 and sifra = $3 limit 1`,
		"", []any{userSession.SelectedGod, userSession.SelectedKar, sifra}, "", "")
	if err != nil {
		return 0, err
	}
	if rows == nil || len(*rows) == 0 {
		return 0, fmt.Errorf("artikal ne postoji")
	}
	rsifID, po := int64(common.StringToInt((*rows)[0].Key)), common.StringToInt((*rows)[0].C1)
	rok, _ := robnoUnosDatum(params.Rok)
	lotovi := []robNivLot{{kolicina: common.StringToFloat64(params.Kolic), otk: strings.TrimSpace(params.Otk), serija: strings.TrimSpace(params.Serija),
		roktr: RobWinDevDatum(rok)}}
	if k.Tipzal == 2 {
		dr, err := s.unos.Tekst.GetAllCustom(ctx, `select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8::text as key, coalesce(otk, '') as c1,
				coalesce(serija, '') as c2, coalesce(roktr, 0)::bigint::text as c3
			from drsta where god = $1 and kar = $2 and mag = $3 and rsifid = $4 and coalesce(ulaz, 0) - coalesce(izlaz, 0) >= 0 order by rstaid`, "",
			[]any{userSession.SelectedGod, userSession.SelectedKar, k.Mag, rsifID}, "", "")
		if err != nil {
			return 0, err
		}
		lotovi = lotovi[:0]
		if dr != nil {
			for _, r := range *dr {
				lotovi = append(lotovi, robNivLot{kolicina: common.StringToFloat64(r.Key), otk: r.C1, serija: r.C2, roktr: int64(common.StringToInt(r.C3))})
			}
		}
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
	stavka := robNivSlogParams{
		sifra: sifra, rsifID: rsifID, po: po, konto: strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", ""),
		staraCena: common.StringToFloat64(params.StaraCena), novaCena: common.StringToFloat64(params.NovaCena), procenat: common.StringToFloat64(params.Procenat),
	}
	for _, lot := range lotovi {
		stavka.robNivLot = lot
		id, err := s.robNivSlog(ctx, tx, k, stavka)
		if err != nil {
			return 0, err
		}
		if id > 0 {
			rproID = id
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robNivSlogParams are the values of one stavka of a nivelacija veleprodaje.
type robNivSlogParams struct {
	robNivLot
	sifra                         int
	rsifID                        int64
	po                            int
	konto                         string
	staraCena, novaCena, procenat float64
}

// robNivSlog adds one stavka of a nivelacija veleprodaje (the legacy NoviSLOGRPRO) in the transaction tx: when
// its iznos (the količina by the nova less the stara cena, rounded to 2 decimals, negative for a storno) is not
// 0 the stavka (the artikal, the konto, the količina, the stara cena as fcena, the nova cena, the procenat as
// vma, the lot and the next redni broj) is added and posted: the nalog (AzurirajRNAL), the stanje and the zalihe
// by lot (AzurirajRSTA and AzurirajDRSTA with the iznos and the nova cena, no količina) and the veleprodajna cena
// of the stanje of the konto, the document gets the iznos and a stavka more. The zalihe by lot of the artikal in
// the magacin get the nova cena as their cena and veleprodajna cena in every case. It returns the id of the
// stavka (0 without one).
func (s *RobnoDokumentaResource) robNivSlog(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, p robNivSlogParams) (rproID int64, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	iznos := robZaokruzi(p.kolicina*(p.novaCena-p.staraCena), 2)
	if iznos != 0 {
		if k.Storno() {
			iznos = -iznos
		}
		var rbr int
		if err = tx.QueryRowContext(ctx, `select coalesce(max(rbr), 0) + 1 from rpro where rdokid = $1`, k.RdokID).Scan(&rbr); err != nil {
			return 0, err
		}
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", []domain.Fields{
			{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
			{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
			{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
			{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
			{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
			robPoljeNullTime("dadok", k.Dadok),
			{Name: "mag", Value: strconv.Itoa(k.Mag)},
			{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
			{Name: "sifra", Value: strconv.Itoa(p.sifra)},
			{Name: "konto", Value: p.konto},
			robPoljeFloat("kolic", p.kolicina),
			robPoljeFloat("fcena", p.staraCena),
			robPoljeFloat("cena", p.novaCena),
			robPoljeFloat("iznos", iznos),
			robPoljeFloat("vma", p.procenat),
			{Name: "po", Value: strconv.Itoa(p.po)},
			{Name: "rsifid", Value: strconv.FormatInt(p.rsifID, 10)},
			{Name: "rbr", Value: strconv.Itoa(rbr)},
			{Name: "brst", Value: strconv.Itoa(rbr)},
			robPoljeText("otk", p.otk),
			robPoljeText("serija", p.serija),
			{Name: "roktr", Value: strconv.FormatInt(p.roktr, 10)},
			// NOT NULL without a default.
			{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
		}); err != nil {
			return 0, err
		}
		if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos); err != nil {
			return 0, err
		}
		promena := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: p.konto, Sifra: p.sifra, Iznos: iznos, Cena: p.novaCena,
			Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Otk: p.otk, Serija: p.serija, Roktr: p.roktr, Brst: 1}
		if err = s.AzurirajRSTA(ctx, tx, promena); err != nil {
			return 0, err
		}
		if err = s.AzurirajDRSTA(ctx, tx, promena); err != nil {
			return 0, err
		}
		if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rsta set vpcena = $6
			where god = $1 and kar = $2 and magaciniid = $3 and konto = $4 and sifra = $5`,
			userSession.SelectedGod, userSession.SelectedKar, k.MagaciniID, p.konto, p.sifra, p.novaCena); err != nil {
			return 0, err
		}
		if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) + $2)::numeric, 2),
			brst = coalesce(brst, 0) + 1 where rdokid = $1`, k.RdokID, iznos); err != nil {
			return 0, err
		}
	}
	err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update drsta set cena = $5, vpcena = $5
		where god = $1 and kar = $2 and magaciniid = $3 and sifra = $4`,
		userSession.SelectedGod, userSession.SelectedKar, k.MagaciniID, p.sifra, p.novaCena)
	return rproID, err
}

// robNivBrisi deletes a stavka of a nivelacija veleprodaje (the legacy BrisiRPRO of rob_Nivelacije_Veleprodaja)
// in the transaction tx: its iznos is taken back from the stanje and the zalihe by lot of the artikal
// (AzurirajRSTA and AzurirajDRSTA without the količina and the cena) and from the document (with a stavka less),
// the stavka is deleted and the nalog loses its iznos (AzurirajRNAL). The cene of the artikal stay (the legacy
// warns to make a nivelacija to the stara cena first).
func (s *RobnoDokumentaResource) robNivBrisi(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64, konto string, sifra int,
	iznos float64, otk, serija string, roktr int64) error {
	promena := domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: strings.TrimSpace(konto), Sifra: sifra, Iznos: -iznos,
		Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Otk: strings.TrimSpace(otk), Serija: strings.TrimSpace(serija),
		Roktr: roktr, Brst: -1}
	if err := s.AzurirajRSTA(ctx, tx, promena); err != nil {
		return err
	}
	if err := s.AzurirajDRSTA(ctx, tx, promena); err != nil {
		return err
	}
	if err := s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) - $2)::numeric, 2),
		brst = greatest(coalesce(brst, 0) - 1, 0) where rdokid = $1`, k.RdokID, iznos); err != nil {
		return err
	}
	if err := s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return err
	}
	return s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos)
}

// RobNivStavkeTableRows returns the stavke of a nivelacija veleprodaje as the rows of the grid of the stavke (the
// columns of GetOpstiNivStavkeTableFields: redni broj, šifra, naziv, JM, količina, stara cena, procenat, nova
// cena, razlika (the količina by the nova less the stara cena) and, with lot, the OTK broj, the serija and the
// rok).
func RobNivStavkeTableRows(lot bool) func([]domain.RobStavkaRow) []domain.TableRow {
	return func(stavke []domain.RobStavkaRow) []domain.TableRow {
		rows := make([]domain.TableRow, 0, len(stavke))
		for _, st := range stavke {
			fields := []string{strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.Jm, common.FormatNumberWithSystemLocale(st.Kolic, 3),
				robFormatIznos(st.Fcena), common.FormatNumberWithSystemLocale(st.Vma, 2), robFormatIznos(st.Cena),
				robFormatIznos(st.Kolic * (st.Cena - st.Fcena))}
			if lot {
				rok := ""
				if st.Roktr > 0 {
					rok = robnoStampaWinDevDatumFormat(int64(st.Roktr), common.DateLayout)
				}
				fields = append(fields, st.Otk, st.Serija, rok)
			}
			rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), Fields: fields})
		}
		return rows
	}
}

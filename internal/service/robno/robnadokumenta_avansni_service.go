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
// Avansni računi (the groups ARA and ARU, the legacy rob_AvansniRacuni and rob_AvansniRacuni_DET): the screen
// of the fakture usluga with the header of the faktura without what an avansni račun has not, and the stavke
// of the šifra 0 (the opis of the robe or usluge, the količina 1, the tarifa, the cena and the iznos); the
// avans (avansfakt) gets the osnovice and the PDV of its stavke
// ---------------------------------------------------------------------------------------------------

// RobAvansniGrupe are the groups of the vrste dokumenata of the avansni računi.
var RobAvansniGrupe = []string{"ARA", "ARU"}

// robAvansniNaslov is the title of the screen of the avansni računi.
const robAvansniNaslov = "Avansni računi"

// robAvansniPolja are the columns of the header of the faktura an avansni račun saves (the legacy SnimiRDOK of
// rob_AvansniRacuni); the others of the faktura are not on its screen.
var robAvansniPolja = []string{"dadok", "dokiz", "datiz", "opis", "fkto", "fana", "komid", "kom", "pla", "otp", "posuslnab", "brugov",
	"polje", "keykat", "category", "pornapomena", "descriptioncode", "idorgjed", "mestotrid", "foot"}

// robAvansni tells whether the group of the vrsta is the one of an avansni račun.
func robAvansni(grupa string) bool {
	return robSadrzi(RobAvansniGrupe, strings.ToUpper(strings.TrimSpace(grupa)))
}

// robAvansniDokument makes the document of the fakture usluga the one of an avansni račun (the legacy
// rob_AvansniRacuni): the title, the header of the avansni račun with its profaktura and, for a new one, the
// uslovi plaćanja "AVANS"; the stavka of the šifra 0 with the količina 1.
func (s *RobnoDokumentaResource) robAvansniDokument(ctx context.Context, res *domain.RobnoUslugeDokument) error {
	h, st := &res.Header, &res.Stavka
	h.Avansni, h.Knjizno = true, false
	st.Avansni = true
	res.Naslov = robAvansniNaslov
	st.Naslov = robAvansniNaslov + " - stavke"
	st.SifraArtikla, st.Kolicina = "0", "1"
	res.Avansi = nil
	if h.RdokID > 0 {
		rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(profak, '') as key from rdok where rdokid = $1`, "", []any{h.RdokID}, "", "")
		if err != nil {
			return err
		}
		if rows != nil && len(*rows) > 0 {
			h.Profaktura = strings.TrimSpace((*rows)[0].Key)
		}
	} else if strings.TrimSpace(h.UsloviPlacanja) == "" {
		h.UsloviPlacanja = "AVANS"
	}
	return nil
}

// ValidateRobAvansni validates the header of an avansni račun (the legacy ValidacijaRDOK of rob_AvansniRacuni):
// the checks of the faktura it has (ValidateRobFakture with Avansni: the datumi, the kupac, the org. jedinica,
// the mesto troška, the komercijalista, the POPDV polje or the poreska kategorija, the nastanak PDV obaveze,
// the broj dokumenta), the broj dokumenta not 0 and the analitička konta of the izlazni računi (fvknjrac I 3)
// the partner of the kupac must have (the legacy "Za partnera ... morate sledeća konta"). The question of the
// datum dokumenta out of the business year is not asked.
func (s *RobnoDokumentaResource) ValidateRobAvansni(ctx context.Context, params domain.RobnoFaktureParams) []domain.FieldError {
	params.Avansni = true
	fieldErrors := s.ValidateRobFakture(ctx, params)
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fieldErrors
	}
	if dokum, err := strconv.Atoi(strings.TrimSpace(params.BrojDokum)); err == nil && dokum == 0 {
		add("brdok", "Ne može biti 0.")
	}
	konto, sifra := strings.TrimSpace(params.KupacKonto), strings.TrimSpace(params.KupacSifra)
	if konto == "" || sifra == "" {
		return fieldErrors
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce((select f.konta from fvknjrac f where f.god = $1 and f.kar = $2 and f.vktip = 'I'
			and f.vkrbr = 3 order by f.vkrbr limit 1), '') as key,
			coalesce((select p.idpartneri::text from fkpl p where p.god = $1 and p.kar = $2 and p.konto = $3 and p.sifra = $4 limit 1), '') as c1`, "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, konto, sifra}, "", "")
	if err != nil {
		add("kupacsifra", common.ErrMsgReadData)
		return fieldErrors
	}
	if rows == nil || len(*rows) == 0 || (*rows)[0].C1 == "" {
		return fieldErrors
	}
	var nema []string
	for _, k := range strings.Split((*rows)[0].Key, ",") {
		if k = strings.TrimSpace(k); k == "" {
			continue
		}
		postoji, err := s.robnoUnosPostoji(ctx, `select count(*) from fkpl where god = $1 and kar = $2 and konto = $3 and idpartneri = $4 and vkonta = 1`,
			userSession.SelectedGod, userSession.SelectedKar, k, common.StringToInt((*rows)[0].C1))
		if err != nil {
			add("kupacsifra", common.ErrMsgReadData)
			return fieldErrors
		}
		if !postoji {
			nema = append(nema, k)
		}
	}
	if len(nema) > 0 {
		add("kupacsifra", "Partner nema konta: "+strings.Join(nema, ", ")+".")
	}
	return fieldErrors
}

// SaveRobAvansni saves the header of an avansni račun (the legacy SnimiRDOK of rob_AvansniRacuni): the columns
// of the faktura it has (robAvansniPolja) with its profaktura, no rok and mesto isporuke (SaveRobniDokument,
// without the avansi of a faktura; a new one gets bi 0 and no opis izmene there); a new one gets its avans
// (avansfakt: the avansid the document, no faktura yet, the osnovice and the PDV 0). It returns the rdokid of
// the document.
func (s *RobnoDokumentaResource) SaveRobAvansni(ctx context.Context, params domain.RobnoFaktureParams) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	dok, err := s.robFaktureSnimanje(ctx, params)
	if err != nil {
		return 0, err
	}
	dok.SaAvansima, dok.Avansi = false, nil
	polja := make([]domain.Fields, 0, len(robAvansniPolja)+5)
	for _, p := range dok.Polja {
		if robSadrzi(robAvansniPolja, p.Name) {
			polja = append(polja, p)
		}
	}
	// A new document gets bi 0 and no opis izmene with the columns of the nalog (robDodajDokument).
	dok.Polja = append(polja, robPoljeText("profak", params.Profaktura), domain.Fields{Name: "rok", Value: "0"},
		domain.Fields{Name: "mi", Value: "0"})
	rdokID, err := s.SaveRobniDokument(ctx, dok)
	if err != nil {
		return 0, err
	}
	// The avans of the document (avansfakt.rdokid, the faktura that closes it, is a key of rdok: NULL for none).
	return rdokID, s.unos.Rpro.CreateUpdateCustom(ctx, `insert into avansfakt (god, kar, avansid, osnfakt1, pdv1fakt, osnfakt2, pdv2fakt, osnfakt3,
			osnavans1, pdv1avans, osnavans2, pdv2avans, osnavans3, xopunos, xdatunosa)
		select $1, $2, $3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, $4, now() where not exists (select 1 from avansfakt where avansid = $3)`,
		userSession.SelectedGod, userSession.SelectedKar, rdokID, userSession.UserName)
}

// DeleteRobAvansni deletes an avansni račun with its stavke (the legacy "Briši" of rob_AvansniRacuni) like a
// faktura (robBrisiDokument: the SEF, the fiskalni printer and the posting are checked), with its avans; an
// avans closed by a faktura is not deleted (poruka says why).
func (s *RobnoDokumentaResource) DeleteRobAvansni(ctx context.Context, rdokID int64) (string, error) {
	zatvoren, err := s.robnoUnosPostoji(ctx, `select count(*) from avansfakt where avansid = $1 and coalesce(rdokid, 0) <> 0`, rdokID)
	if err != nil {
		return "", err
	}
	if zatvoren {
		return "Avans je zatvoren fakturom, brisanje nije dozvoljeno.", nil
	}
	return s.robBrisiDokument(ctx, rdokID, true, `delete from avansfakt where avansid = $1`)
}

// ValidateRobAvansniStavka validates a stavka of an avansni račun (the legacy Validacija of
// rob_AvansniRacuni_DET): the opis is required, the iznos (the količina by the cena) is greater than 0 and the
// tarifa is required.
func (s *RobnoDokumentaResource) ValidateRobAvansniStavka(params domain.RobUslugaStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if strings.TrimSpace(params.Naziv) == "" {
		add("opisusluge", common.ErrMsgObavezanPodatak)
	}
	kolicina := common.StringToFloat64(params.Kolic)
	if kolicina == 0 {
		kolicina = 1
	}
	if kolicina*common.StringToFloat64(params.Cena) <= 0 {
		add("cenausluge", robInterniMsgVeceOdNule)
	}
	if common.StringToInt(params.Tarifa) <= 0 {
		add("tarifa", common.ErrMsgObavezanPodatak)
	}
	return fieldErrors
}

// robAvansniStara is a saved stavka of an avansni račun (before its change or deletion).
type robAvansniStara struct {
	rdokID          int64
	konto, taxcat   string
	kolicina, iznos float64
	po, rbr         int
}

// robAvansniStavkaStara reads a saved stavka of an avansni račun in the transaction tx (locked).
func robAvansniStavkaStara(ctx context.Context, tx db.Transaction, rproID int64) (robAvansniStara, error) {
	st := robAvansniStara{}
	err := tx.QueryRowContext(ctx, `select coalesce(rdokid, 0)::bigint, coalesce(konto, ''), coalesce(taxcat, ''), coalesce(kolic, 0)::float8,
			coalesce(iznos, 0)::float8, coalesce(po, 0), coalesce(rbr, 0)
		from rpro where rproid = $1 for update`, rproID).Scan(&st.rdokID, &st.konto, &st.taxcat, &st.kolicina, &st.iznos, &st.po, &st.rbr)
	if err == sql.ErrNoRows {
		err = fmt.Errorf("stavka ne postoji")
	}
	return st, err
}

// SaveRobAvansniStavka saves a stavka of an avansni račun, a new one (RproID 0, the legacy CreateRpro of
// rob_AvansniRacuni_DET) or a changed one (EditRpro), in one transaction:
//   - the stavka of the šifra 0 (the artikal of the šifra 0 is created when it does not exist, the legacy
//     KreirajRSIF) with the opis, the cena (also the fakturna cena), the količina (1 for a new one; a changed one
//     keeps its own), the iznos (the količina by the cena, rounded to 2 decimals, negative with the količina
//     for a storno), the tarifa and the poreska kategorija with its osnov oslobođenja; a new one gets the
//     columns of its document, the konto of the form and the next redni broj, and the artikal of the šifra 0
//     gets its cena (rcene) when it has none;
//   - the nalog and the stanje of the artikal (AzurirajRNAL, AzurirajRSTA) get the iznos and the količina (a
//     changed stavka the difference), the document its iznos and a new stavka counts in it;
//   - the avans gets the osnovica and the PDV of the stavka by the stopa of its tarifa (a changed stavka: the
//     old ones taken back by its old tarifa, the legacy took the difference by the new tarifa).
func (s *RobnoDokumentaResource) SaveRobAvansniStavka(ctx context.Context, params domain.RobUslugaStavkaParams) (rproID int64, err error) {
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
	naziv, cena, tarifa := strings.TrimSpace(params.Naziv), common.StringToFloat64(params.Cena), common.StringToInt(params.Tarifa)
	taxcat, taxexreco := strings.TrimSpace(params.Taxcat), strings.TrimSpace(params.Taxexreco)
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// The artikal of the šifra 0.
	var rsifID int64
	var jm string
	err = tx.QueryRowContext(ctx, `select rsifid, coalesce(jm, '') from rsif where god = $1 and kar = $2 and sifra = 0 order by rsifid limit 1`,
		userSession.SelectedGod, userSession.SelectedKar).Scan(&rsifID, &jm)
	if err == sql.ErrNoRows {
		if rsifID, err = s.robKreirajRsifNula(ctx, tx, naziv); err != nil {
			err = fmt.Errorf("artikal šifre 0 ne postoji i ne može se kreirati (proverite šifarnike artikala): %w", err)
		}
	}
	if err != nil {
		return 0, err
	}

	stara := robAvansniStara{konto: strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")}
	kolicina := 1.0
	if params.RproID > 0 {
		if stara, err = robAvansniStavkaStara(ctx, tx, params.RproID); err != nil {
			return 0, err
		}
		if stara.rdokID != params.RdokID {
			return 0, fmt.Errorf("stavka ne pripada dokumentu")
		}
		if stara.kolicina != 0 {
			kolicina = math.Abs(stara.kolicina)
		}
	}
	iznos := robZaokruzi(math.Abs(kolicina*cena), 2)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}
	polja := []domain.Fields{
		{Name: "naz1", Value: naziv},
		robPoljeFloat("cena", cena),
		robPoljeFloat("fcena", cena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "po", Value: strconv.Itoa(tarifa)},
		{Name: "taxcat", Value: taxcat},
		{Name: "taxexreco", Value: taxexreco},
	}
	if params.RproID > 0 {
		rproID = params.RproID
		if err = s.unos.Rpro.UpdateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", rproID, polja); err != nil {
			return 0, err
		}
	} else {
		var rbr int
		if err = tx.QueryRowContext(ctx, `select coalesce(max(rbr), 0) + 1 from rpro where rdokid = $1`, params.RdokID).Scan(&rbr); err != nil {
			return 0, err
		}
		polja = append(polja,
			domain.Fields{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
			domain.Fields{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
			domain.Fields{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
			domain.Fields{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
			domain.Fields{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
			robPoljeNullTime("dadok", k.Dadok),
			domain.Fields{Name: "sifra", Value: "0"},
			domain.Fields{Name: "konto", Value: stara.konto},
			robPoljeText("jm", jm),
			domain.Fields{Name: "ncena", Value: "0"},
			domain.Fields{Name: "fkto", Value: k.Fkto},
			domain.Fields{Name: "fana", Value: k.Fana},
			domain.Fields{Name: "rab", Value: "0"},
			domain.Fields{Name: "dani", Value: "0"},
			domain.Fields{Name: "mag", Value: strconv.Itoa(k.Mag)},
			domain.Fields{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""},
		)
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja); err != nil {
			return 0, err
		}
		// The cena of the artikal of the šifra 0.
		var cene int
		if err = tx.QueryRowContext(ctx, `select count(*) from rcene where god = $1 and kar = $2 and sifra = 0`,
			userSession.SelectedGod, userSession.SelectedKar).Scan(&cene); err != nil {
			return 0, err
		}
		if cene == 0 {
			if _, err = s.unos.Rcene.CreateWithTx(ctx, tx, &domain.RobRceneDto{}, "rceneid", []domain.Fields{
				{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)}, {Name: "sifra", Value: "0"},
				{Name: "mma", Value: "0"}, {Name: "vra", Value: "0"}, {Name: "mra", Value: "0"}, {Name: "ztro", Value: "0"},
				robPoljeFloat("cena", cena), {Name: "fcena", Value: "0"}, {Name: "mcena", Value: "0"}, {Name: "mcenap", Value: "0"},
				{Name: "ncena", Value: "0"}, {Name: "ztrin", Value: "0"}, {Name: "rab", Value: "0"}, {Name: "tz", Value: ""}, {Name: "tzi", Value: ""},
			}); err != nil {
				return 0, err
			}
		}
	}

	// The nalog, the stanje and the document: the difference to the old stavka.
	brst := 0
	if params.RproID == 0 {
		brst = 1
	}
	if err = s.AzurirajRSTA(ctx, tx, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: stara.konto, Sifra: 0, Kolicina: kolicina - stara.kolicina,
		Iznos: iznos - stara.iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Brst: brst}); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) + $2)::numeric, 2),
		brst = coalesce(brst, 0) + $3 where rdokid = $1`, params.RdokID, iznos-stara.iznos, brst); err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos-stara.iznos); err != nil {
		return 0, err
	}
	// The avans: the old stavka taken back, the new one added.
	if params.RproID > 0 {
		if err = s.robAvansniAvansfakt(ctx, tx, params.RdokID, stara.po, stara.taxcat, -stara.iznos); err != nil {
			return 0, err
		}
	}
	if err = s.robAvansniAvansfakt(ctx, tx, params.RdokID, tarifa, taxcat, iznos); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// DeleteRobAvansniStavka deletes a stavka of an avansni račun (the legacy BrisiRPRO of rob_AvansniRacuni_DET) in
// one transaction and returns its document: the stavka is taken back from the stanje, the document (its iznos
// and a stavka less) and the avans, it is deleted, the nalog loses its iznos (AzurirajRNAL) and the stavke
// after it get the redni brojevi again.
func (s *RobnoDokumentaResource) DeleteRobAvansniStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	stara, err := robAvansniStavkaStara(ctx, tx, rproID)
	if err != nil {
		return 0, err
	}
	rdokID = stara.rdokID
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("dokument ne postoji")
	}
	if err = s.AzurirajRSTA(ctx, tx, domain.RobPromenaStanja{MagaciniID: k.MagaciniID, Konto: stara.konto, Sifra: 0, Kolicina: -stara.kolicina,
		Iznos: -stara.iznos, Danal: k.Danal.Time, DokvrstaID: k.DokvrstaID, Tipdok: k.Tipdok, Tz: "-", Tzi: "-", Brst: -1}); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) - $2)::numeric, 2),
		brst = greatest(coalesce(brst, 0) - 1, 0) where rdokid = $1`, rdokID, stara.iznos); err != nil {
		return 0, err
	}
	if err = s.robAvansniAvansfakt(ctx, tx, rdokID, stara.po, stara.taxcat, -stara.iznos); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -stara.iznos); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rpro p set rbr = x.n
		from (select rproid, ($2 - 1 + row_number() over (order by rbr, rproid))::int as n from rpro where rdokid = $1 and rbr > $2) x
		where p.rproid = x.rproid`, rdokID, stara.rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// robAvansniAvansfakt adds the osnovica and the PDV of an iznos of a stavka of an avansni račun to its avans
// (the first avansfakt of the avansid; the legacy CreateRpro, EditRpro and BrisiRPRO), by the stopa of the
// tarifa on the datum of the document: the stopa 18 or 20 the osnovica and the PDV 1, 8 or 10 the osnovica and
// the PDV 2, the stopa 0 the osnovica 3; with the napomena o poreskom oslobođenju a stavka of a poreska
// kategorija other than S10 and S20 (and "", "-") is without the PDV (in the osnovica 3: for the SEF).
func (s *RobnoDokumentaResource) robAvansniAvansfakt(ctx context.Context, tx db.Transaction, rdokID int64, tarifa int, taxcat string, iznos float64) error {
	var stopa float64
	var napomena string
	if err := tx.QueryRowContext(ctx, `select `+fmt.Sprintf(robPoreskaStopaSQL, "$2", "dadok")+`, coalesce(pornapomena, '') from rdok where rdokid = $1`,
		rdokID, tarifa).Scan(&stopa, &napomena); err != nil {
		return err
	}
	oslobodjeno := strings.TrimSpace(napomena) != "" && !robSadrzi([]string{"", "-", "S10", "S20"}, strings.TrimSpace(taxcat))
	var osn1, pdv1, osn2, pdv2, osn3 float64
	if stopa == 18 || stopa == 20 {
		if oslobodjeno {
			stopa = 0
		} else {
			osn1 = robZaokruzi(iznos, 2)
		}
		pdv1 = robZaokruzi(iznos*stopa/100, 2)
	}
	if stopa == 8 || stopa == 10 {
		if oslobodjeno {
			stopa = 0
		} else {
			osn2 = robZaokruzi(iznos, 2)
		}
		pdv2 = robZaokruzi(iznos*stopa/100, 2)
	}
	if stopa == 0 {
		osn3 = robZaokruzi(iznos, 2)
	}
	return s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update avansfakt set osnavans1 = coalesce(osnavans1, 0) + $2, pdv1avans = coalesce(pdv1avans, 0) + $3,
			osnavans2 = coalesce(osnavans2, 0) + $4, pdv2avans = coalesce(pdv2avans, 0) + $5, osnavans3 = coalesce(osnavans3, 0) + $6
		where avansfaktid = (select min(a.avansfaktid) from avansfakt a where a.avansid = $1)`, rdokID, osn1, pdv1, osn2, pdv2, osn3)
}

// robKreirajRsifNula creates the artikal of the šifra 0 (a usluga, tip U, with the naziv; the legacy
// KreirajRSIF) in the transaction tx and returns its rsifid. rsif has many columns without a default: the
// artikal copies the keys of the šifarnici (the foreign keys: grupa, podgrupa...) of the first artikal of the
// godina and the kartica, its other columns have their default or are empty ("", 0, false; NULL when allowed).
func (s *RobnoDokumentaResource) robKreirajRsifNula(ctx context.Context, tx db.Transaction, naziv string) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	kolone, err := s.unos.Tekst.GetAllCustom(ctx, `select c.column_name as key, c.data_type as c1, c.is_nullable as c2,
			coalesce((select f.confrelid::regclass::text || '.' || r.attname from pg_constraint f
				join pg_attribute a on a.attrelid = f.conrelid and a.attnum = f.conkey[1]
				join pg_attribute r on r.attrelid = f.confrelid and r.attnum = f.confkey[1]
				where f.conrelid = 'rsif'::regclass and f.contype = 'f' and a.attname = c.column_name limit 1), '') as c3,
			(c.column_default is not null)::text as c4
		from information_schema.columns c
		where c.table_schema = current_schema() and c.table_name = 'rsif' and c.is_identity = 'NO'
		order by c.ordinal_position`, "", nil, "", "")
	if err != nil {
		return 0, err
	}
	if kolone == nil {
		return 0, fmt.Errorf("rsif nema kolone")
	}
	names, izrazi := []string{}, []string{}
	for _, k := range *kolone {
		izraz := ""
		switch k.Key {
		case "rsifid":
			continue
		case "god":
			izraz = "$1"
		case "kar":
			izraz = "$2"
		case "sifra":
			izraz = "0"
		case "tip":
			izraz = "'U'"
		case "naziv":
			izraz = "$3"
		case "xopunos":
			izraz = "$4"
		case "xdatunosa":
			izraz = "now()"
		default:
			switch {
			case k.C3 != "":
				// A key of a šifarnik (also with a default: its default may not exist).
				izraz = k.Key
			case k.C4 == "true":
				// The default of the column.
				continue
			case k.C2 == "YES":
				izraz = "null"
			case strings.Contains(k.C1, "char") || k.C1 == "text":
				izraz = "''"
			case k.C1 == "boolean":
				izraz = "false"
			case strings.Contains(k.C1, "date") || strings.Contains(k.C1, "time"):
				izraz = k.Key
			default:
				izraz = "0"
			}
		}
		// A key of a šifarnik: the one of the copied artikal when it exists in the šifarnik, else the first one
		// of the šifarnik (the keys of the old artikli are not always checked).
		if ref := strings.SplitN(k.C3, ".", 2); k.C3 != "" && len(ref) == 2 {
			izraz = fmt.Sprintf("coalesce((select x.%[3]s from %[2]s x where x.%[3]s = rsif.%[1]s limit 1), (select min(x.%[3]s) from %[2]s x))",
				k.Key, ref[0], ref[1])
		}
		names, izrazi = append(names, k.Key), append(izrazi, izraz)
	}
	var rsifID int64
	err = tx.QueryRowContext(ctx, `insert into rsif (`+strings.Join(names, ", ")+`) select `+strings.Join(izrazi, ", ")+`
		from rsif where rsif.god = $1 and rsif.kar = $2 order by rsif.rsifid limit 1 returning rsifid`,
		userSession.SelectedGod, userSession.SelectedKar, naziv, userSession.UserName).Scan(&rsifID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("nema artikala za šifru 0")
	}
	return rsifID, err
}

// RobAvansniStavkeTableRows returns the stavke of an avansni račun as the rows of the grid of the stavke (the
// columns of GetAvansniStavkeTableFields: redni broj, opis, količina, cena, iznos, tarifa, poreska kategorija,
// osnov oslobođenja).
func RobAvansniStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Naziv, common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Cena),
				robFormatIznos(st.Iznos), strconv.Itoa(st.Po), st.Taxcat, st.Taxexreco},
		})
	}
	return rows
}

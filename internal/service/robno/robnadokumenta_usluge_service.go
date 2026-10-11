package robno

import (
	"context"
	"database/sql"
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
// Fakture usluga (the groups PRE, FUR, FCO, RCO and KPF, the legacy rob_Fakture_Usluga): the document is the
// faktura (its header, validation, save, obračun and deletion are the ones of the faktura) with the stavke of
// the legacy rob_FAKTURE_Usluga_DET2 (a stavka of the šifra 0 has its own naziv, JM and tarifa)
// ---------------------------------------------------------------------------------------------------

// RobUslugeGrupe are the groups of the vrste dokumenata of the "Fakture usluga" screen.
var RobUslugeGrupe = []string{"PRE", "FUR", "FCO", "RCO", "KPF"}

// robUslugeOtpremaPolja are the columns of the faktura the faktura usluga has not on its screen (the
// paritet, the izvozna izjava, the banka za uplatu - izvoz, the težine, the kutije and the palete): its save
// does not write them.
var robUslugeOtpremaPolja = []string{"paritet", "izjizv", "bnkizvid", "sifbank", "bttowght", "netwght", "totboxs", "brpaleta"}

// robUslugeNaslov is the title of the screen by the group of the vrsta (the legacy STC_Static1 of
// rob_Fakture_Usluga and Naslov): "Zaduženje CO", "Razduženje CO", the knjižno odobrenje (a storno of the
// KPF) or zaduženje, else "Fakture tekst".
func robUslugeNaslov(grupa string, storno bool) string {
	switch grupa {
	case "FCO":
		return "Zaduženje CO"
	case "RCO":
		return "Razduženje CO"
	case "KPF":
		if storno {
			return "Knjižno odobrenje"
		}
		return "Knjižno zaduženje"
	}
	return "Fakture tekst"
}

// GetRobUslugeDokument returns the document of the "Fakture usluga" screen with the broj the user selected or
// typed, like the legacy exit of EDT_DOKUM of rob_Fakture_Usluga: the document of the faktura
// (GetRobFaktureDokument) with the header of the usluge and the entry of a stavka of the usluge (the konto of
// the magacin, the tip zaliha and the cene of the magacin; a KPF has the šifra 0) and the title of the screen.
func (s *RobnoDokumentaResource) GetRobUslugeDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoUslugeDokument, error) {
	dok, err := s.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return domain.RobnoUslugeDokument{}, err
	}
	res := domain.RobnoUslugeDokument{Header: dok.Header, Stavka: domain.RobnoUslugeStavka{RobnoFaktureStavka: dok.Stavka},
		Avansi: dok.Avansi, ZiroRacuni: dok.ZiroRacuni, Poruka: dok.Poruka}
	grupa, err := s.robInterniGrupa(ctx, common.StringToInt(res.Header.Vrd))
	if err != nil {
		return res, err
	}
	h, st := &res.Header, &res.Stavka
	h.Usluge, h.Knjizno = true, grupa == "KPF"
	st.Grupa = grupa
	res.Naslov = robUslugeNaslov(grupa, h.Storno)
	st.Naslov = res.Naslov + " - stavke"
	if grupa == "KPF" {
		st.SifraArtikla = "0"
	}
	// The magacin: its tip zaliha (the OTK broj, the serija and the rok), how it keeps its zalihe.
	if magacin, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(tipzal, 0)::text as key, coalesce(nacvodzal, 0)::text as c1
		from magacini where magaciniid = $1`, "", []any{h.MagaciniID}, "", ""); err == nil && magacin != nil && len(*magacin) > 0 {
		st.Tipzal = common.StringToInt((*magacin)[0].Key)
		nacvodzal := common.StringToInt((*magacin)[0].C1)
		st.VpCena, st.NabCena = nacvodzal == 1, nacvodzal == 3
	}
	// The konto of a new stavka: the first konto of the magacin (the legacy sMagKonto).
	if konto, err := s.robUslugeKontoMagacina(ctx, h.MagaciniID); err == nil && konto != "" {
		st.Konto = konto
	}
	return res, nil
}

// robUslugeKontoMagacina is the first konto of the magacin (magkonto, the legacy sMagKonto).
func (s *RobnoDokumentaResource) robUslugeKontoMagacina(ctx context.Context, magaciniID int) (string, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return "", fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select trim(konto) as key from magkonto where god = $1 and kar = $2 and magaciniid = $3
		and coalesce(trim(konto), '') <> '' order by magkontoid limit 1`, "", []any{userSession.SelectedGod, userSession.SelectedKar, magaciniID}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return "", err
	}
	return (*rows)[0].Key, nil
}

// robUslugeDokiz is the izvorni dokument of a knjižno pismo finansijsko (GOD-MAG-VRD-DOKUM, the legacy
// ProveriDOKIZ).
var robUslugeDokiz = regexp.MustCompile(`^(\d+)-(\d+)-(\d+)-(\d+)$`)

// robUslugeIzvorni reads the izvorni dokument of a knjižno pismo finansijsko (the legacy ProveriDOKIZ): the
// broj GOD-MAG-VRD-DOKUM (the godina of two digits), its datum, its SEF id and status. msg is the error of the
// field: a wrong broj, an avansni račun (ARA, ARU) for a knjižno zaduženje (not a storno), a document that
// does not exist, a document of a knjižno odobrenje (a storno) sent to the SEF and not approved there.
func (s *RobnoDokumentaResource) robUslugeIzvorni(ctx context.Context, dokiz string, storno bool) (datiz sql.NullTime, msg string, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return datiz, "", fmt.Errorf("no user session found")
	}
	m := robUslugeDokiz.FindStringSubmatch(strings.ReplaceAll(strings.TrimSpace(dokiz), " ", ""))
	if m == nil {
		return datiz, "Format: GOD-MAG-VRD-DOKUM.", nil
	}
	god, mag, vrd, dokum := common.StringToInt(m[1])+2000, common.StringToInt(m[2]), common.StringToInt(m[3]), common.StringToInt(m[4])
	if !storno {
		avansni, err := s.robnoUnosPostoji(ctx, `select count(*) from dokvrsta where god = $1 and kar = $2 and vrd = $3
			and upper(trim(coalesce(grpdok, ''))) in ('ARA', 'ARU')`, god, userSession.SelectedKar, vrd)
		if err != nil {
			return datiz, "", err
		}
		if avansni {
			return datiz, "Ne može avansni račun.", nil
		}
	}
	var sefID float64
	var status string
	err = s.unos.Rpro.DB.QueryRowContext(ctx, `select dadok, coalesce(salesinvoiceid, 0)::float8, coalesce(status_salinv, '')
		from rdok where god = $1 and kar = $2 and mag = $3 and vrd = $4 and dokum = $5 order by rdokid limit 1`,
		god, userSession.SelectedKar, mag, vrd, dokum).Scan(&datiz, &sefID, &status)
	if err == sql.ErrNoRows {
		return datiz, "Izvorni dokument ne postoji.", nil
	}
	if err != nil {
		return datiz, "", err
	}
	status = strings.TrimSpace(status)
	if storno && sefID > 0 && status != "Approved" && status != "Odobrena" {
		return datiz, "Nije odobren na SEF-u.", nil
	}
	return datiz, "", nil
}

// ValidateRobUsluge validates the header of a faktura usluga before it is saved, like the legacy
// ValidacijaRDOK of rob_Fakture_Usluga: the checks of the faktura (ValidateRobFakture) and the ones of the
// usluge (the questions of the legacy do not stop the saving and are not part of it):
//   - a valuta other than the domaća one has the kurs;
//   - a knjižno pismo finansijsko (KPF) "Za period" has the datum od and do (od not after do), else the
//     izvorni dokument (GOD-MAG-VRD-DOKUM, ProveriDOKIZ); a knjižno odobrenje (a storno) has the datum
//     smanjenja, not after the datum dokumenta.
func (s *RobnoDokumentaResource) ValidateRobUsluge(ctx context.Context, params domain.RobnoFaktureParams) []domain.FieldError {
	fieldErrors := s.ValidateRobFakture(ctx, params)
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fieldErrors
	}
	vrd := common.StringToInt(params.Vrd)
	dokvrsta, found, err := s.robnoUnosDokvrsta(ctx, vrd)
	if err != nil || !found {
		return fieldErrors
	}
	grupa := strings.ToUpper(strings.TrimSpace(dokvrsta.GrpDok))
	storno := strings.TrimSpace(dokvrsta.Predznak) == "-"

	// The kurs of a foreign valuta.
	if id, given, ok := robnoUnosID(params.IDValute); ok && given {
		strana, err := s.robnoUnosPostoji(ctx, `select count(*) from valute v where v.idvalute = $1 and coalesce(v.sifval, 0) <>
			coalesce((select f.sifval from fvr f where f.god = $2 and f.kar = $3 limit 1), 0)`, id, userSession.SelectedGod, userSession.SelectedKar)
		if err != nil {
			add("kurs", common.ErrMsgReadData)
		} else if strana && common.StringToFloat64(params.Kurs) <= 0 {
			add("kurs", robInterniMsgVeceOdNule)
		}
	}

	if grupa != "KPF" {
		return fieldErrors
	}
	if robnoUnosChecked(params.ZaPeriod) {
		od, odOK := robnoUnosDatum(params.PeriodOd)
		do, doOK := robnoUnosDatum(params.PeriodDo)
		if !odOK {
			add("periodod", robnoUnosDatumMsg(params.PeriodOd))
		}
		if !doOK {
			add("perioddo", robnoUnosDatumMsg(params.PeriodDo))
		}
		if odOK && doOK && od.After(do) {
			add("periodod", "Posle datuma do.")
		}
	} else if strings.TrimSpace(params.IzvorniDokum) == "" {
		add("brdokorg", common.ErrMsgObavezanPodatak)
	}
	if strings.TrimSpace(params.IzvorniDokum) != "" {
		if _, msg, err := s.robUslugeIzvorni(ctx, params.IzvorniDokum, storno); err != nil {
			add("brdokorg", common.ErrMsgReadData)
		} else if msg != "" {
			add("brdokorg", msg)
		}
	}
	if storno {
		if smanjenje, ok := robnoUnosDatum(params.DatumSmanjenja); !ok {
			add("datumsmanjenja", robnoUnosDatumMsg(params.DatumSmanjenja))
		} else if dadok, ok := robnoUnosDatum(params.DatumDokum); ok && smanjenje.After(dadok) {
			add("datumsmanjenja", "Posle datuma dokumenta.")
		}
	}
	return fieldErrors
}

// SaveRobUsluge saves the header of a faktura usluga (the legacy "Sačuvaj" of rob_Fakture_Usluga, after
// ValidacijaRDOK) like the faktura (SaveRobFaktura: SnimiRDOK, PreracunRATA, ObracunFAKT), without the
// columns the faktura usluga has not on its screen, with "Za period" and its datumi of a knjižno pismo
// finansijsko (KPF) and the datum of its izvorni dokument (the legacy ProveriDOKIZ). The datum smanjenja is
// not saved (rdok.datpro). It returns the rdokid of the document.
func (s *RobnoDokumentaResource) SaveRobUsluge(ctx context.Context, params domain.RobnoFaktureParams) (int64, error) {
	dokvrsta, _, err := s.robnoUnosDokvrsta(ctx, common.StringToInt(params.Vrd))
	if err != nil {
		return 0, err
	}
	knjizno := strings.EqualFold(strings.TrimSpace(dokvrsta.GrpDok), "KPF")
	if knjizno && strings.TrimSpace(params.IzvorniDokum) != "" {
		datiz, msg, err := s.robUslugeIzvorni(ctx, params.IzvorniDokum, strings.TrimSpace(dokvrsta.Predznak) == "-")
		if err != nil {
			return 0, err
		}
		if msg == "" && datiz.Valid {
			params.IzvorniDokDatum = datiz.Time.Format(common.HtmlLayout)
		}
	}
	dok, err := s.robFaktureSnimanje(ctx, params)
	if err != nil {
		return 0, err
	}
	polja := make([]domain.Fields, 0, len(dok.Polja)+3)
	for _, p := range dok.Polja {
		if !robSadrzi(robUslugeOtpremaPolja, p.Name) {
			polja = append(polja, p)
		}
	}
	if knjizno {
		zaPeriod := robnoUnosChecked(params.ZaPeriod)
		polja = append(polja, domain.Fields{Name: "kpperiod", Value: strconv.FormatBool(zaPeriod)})
		if zaPeriod {
			polja = append(polja, robPoljeDatum("oddat", params.PeriodOd), robPoljeDatum("dodat", params.PeriodDo))
		} else {
			polja = append(polja, domain.Fields{Name: "oddat", IsNull: true}, domain.Fields{Name: "dodat", IsNull: true})
		}
	}
	dok.Polja = polja
	rdokID, err := s.SaveRobniDokument(ctx, dok)
	if err != nil {
		return 0, err
	}
	if err := s.PreracunRataDokumenta(ctx, rdokID); err != nil {
		return rdokID, err
	}
	return rdokID, s.ObracunFakture(ctx, rdokID)
}

// ---------------------------------------------------------------------------------------------------
// The stavke of a faktura usluga (the legacy rob_FAKTURE_Usluga_DET2)
// ---------------------------------------------------------------------------------------------------

// robUslugaTaxcat is the poreska kategorija of a tarifa (the legacy exit of EDT_po: 1 S20, 2 S10).
func robUslugaTaxcat(tarifa int) string {
	switch tarifa {
	case 1:
		return "S20"
	case 2:
		return "S10"
	}
	return ""
}

// GetRobUslugaArtikal returns the artikal of a stavka of a faktura usluga with what the form of the stavke
// shows of it, like the legacy exit of EDT_SIFRA of rob_FAKTURE_Usluga_DET2:
//   - the artikal by its šifra, or by its barkod for a value longer than 6 characters: its naziv, JM, tarifa
//     with its stopa on the datum of the document (VratiPorez2) and its poreska kategorija, its tip; the
//     konto of the stavka is the konto of the artikal, else the first konto of the magacin;
//   - its stanje and cene in the konto (rsta): the nabavna cena (prosnc), the cena (vpcena, else the cena of
//     rcene) and the magacinska cena (cena); without the stanje all of them 0;
//   - for the vrsta 280 the cena is the last cena to the kupac (ZadnjaCena270), 0 without it.
//
// found is false when the artikal does not exist.
func (s *RobnoDokumentaResource) GetRobUslugaArtikal(ctx context.Context, rdokID int64, sifraIliBarkod string) (domain.RobUslugaArtikal, bool, error) {
	a := domain.RobUslugaArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	k, found, err := s.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		return a, false, err
	}
	vrednost := strings.TrimSpace(sifraIliBarkod)
	kolona := "a.sifra::text"
	if len(vrednost) > 6 {
		kolona = "a.barkod"
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, fmt.Sprintf(`select a.sifra::text as key, coalesce(a.naziv, '') as c1,
			coalesce(a.jm, '') as c2, coalesce(a.po, 0)::text as c3, coalesce(a.tip, '') as c4, trim(coalesce(a.konto, '')) as c5,
			`+fmt.Sprintf(robPoreskaStopaSQL, "a.po", "$4::date")+`::text as c6,
			coalesce((select c.cena from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra limit 1), 0)::float8::text as c7
		from rsif a where a.god = $1 and a.kar = $2 and %s = $3 limit 1`, kolona), "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, vrednost, robNullTime(k.Dadok)}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return a, false, err
	}
	r := (*rows)[0]
	a.Sifra, a.Naziv, a.Jm = common.StringToInt(r.Key), r.C1, r.C2
	a.Tarifa, a.Tip, a.Konto = common.StringToInt(r.C3), strings.TrimSpace(r.C4), r.C5
	a.Stopa, a.Taxcat = common.StringToFloat64(r.C6), robUslugaTaxcat(a.Tarifa)
	if a.Konto == "" {
		if a.Konto, err = s.robUslugeKontoMagacina(ctx, k.MagaciniID); err != nil {
			return a, true, err
		}
	}
	stanje, err := s.unos.Tekst.GetAllCustom(ctx, `select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8::text as key,
			coalesce(prosnc, 0)::float8::text as c1, coalesce(vpcena, 0)::float8::text as c2, coalesce(cena, 0)::float8::text as c3
		from rsta where god = $1 and kar = $2 and magaciniid = $3 and konto = $4 and sifra = $5
		order by rstaid limit 1`, "", []any{userSession.SelectedGod, userSession.SelectedKar, k.MagaciniID, a.Konto, a.Sifra}, "", "")
	if err != nil {
		return a, true, err
	}
	if stanje != nil && len(*stanje) > 0 {
		st := (*stanje)[0]
		a.Stanje, a.Ncena, a.Cena, a.MagCena = common.StringToFloat64(st.Key), common.StringToFloat64(st.C1),
			common.StringToFloat64(st.C2), common.StringToFloat64(st.C3)
		if a.Cena == 0 {
			a.Cena = common.StringToFloat64(r.C7)
		}
	}
	if k.Vrd == 280 {
		a.Cena = 0
		if zadnja, found, err := s.GetRobZadnjaCena(ctx, rdokID, a.Sifra); err != nil {
			return a, true, err
		} else if found {
			a.Cena = zadnja.Fcena
		}
	}
	return a, true, nil
}

// GetRobUslugaStopa returns the poreska stopa of the tarifa of a stavka of a faktura usluga (the legacy exit
// of EDT_po): the stopa on the datum of the document (VratiPorez2), for a document with the stari PDV
// (rdok.flgstaripdv) the stopa of the day before the last change of the stopa of the tarifa; with the
// poreska kategorija of the tarifa.
func (s *RobnoDokumentaResource) GetRobUslugaStopa(ctx context.Context, rdokID int64, tarifa int) (domain.RobUslugaStopa, error) {
	res := domain.RobUslugaStopa{Taxcat: robUslugaTaxcat(tarifa)}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select (case when coalesce(d.flgstaripdv, false) then
			coalesce((select x.pp from rpor x where x.po = $2 and x.datum <= (select max(y.datum) from rpor y where y.po = $2) - 1
				order by x.datum desc limit 1), 0)
		else `+fmt.Sprintf(robPoreskaStopaSQL, "$2", "d.dadok")+` end)::float8::text as key
		from rdok d where d.rdokid = $1`, "", []any{rdokID, tarifa}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return res, err
	}
	res.Stopa = common.StringToFloat64((*rows)[0].Key)
	return res, nil
}

// robUslugaStanjeStavke reads the stanje of the artikal of a stavka of a faktura usluga in the konto (rsta:
// ulaz - izlaz) and its tip.
func (s *RobnoDokumentaResource) robUslugaStanjeStavke(ctx context.Context, magaciniID int, konto string, sifra int) (stanje float64, tip string, found bool, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, "", false, fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(a.tip, '') as key,
			coalesce((select (coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0))::float8 from rsta r where r.god = a.god and r.kar = a.kar
				and r.magaciniid = $4 and r.konto = $5 and r.sifra = a.sifra order by r.rstaid limit 1), 0)::float8::text as c1
		from rsif a where a.god = $1 and a.kar = $2 and a.sifra = $3 limit 1`, "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, sifra, magaciniID, strings.TrimSpace(konto)}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return 0, "", false, err
	}
	return common.StringToFloat64((*rows)[0].C1), strings.TrimSpace((*rows)[0].Key), true, nil
}

// ValidateRobUslugaStavka validates a stavka of a faktura usluga before it is saved, like the legacy
// Validacija of rob_FAKTURE_Usluga_DET2:
//   - the šifra is required (the šifra 0, a usluga with its own naziv, not for the zaduženje and the
//     razduženje CO) and the artikal must exist;
//   - the količina is required (not 0);
//   - an artikal of the zalihe (a šifra other than 0, not of the tip U) of a document that is not a storno
//     has the količina in the stanje of its konto (the stanje of a changed stavka has its own količina):
//     "Nema dovoljno zaliha." (in a magacin of the nabavna cena the error of the legacy, else its question
//     "Da li ste sigurni da izdajete robu?" as the error of the field, like the faktura);
//   - the cena is required and greater than 0.
func (s *RobnoDokumentaResource) ValidateRobUslugaStavka(ctx context.Context, params domain.RobUslugaStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
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
	grupa, err := s.robInterniGrupa(ctx, k.Vrd)
	if err != nil {
		add("sifraartikla", common.ErrMsgReadData)
		return fieldErrors
	}
	sifraText := strings.TrimSpace(params.Sifra)
	sifra, sifraErr := strconv.Atoi(sifraText)
	tip, stanje, artikalOK := "", 0.0, false
	switch {
	case sifraText == "" || (sifra == 0 && (grupa == "FCO" || grupa == "RCO")):
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	case sifraErr != nil:
		add("sifraartikla", robnoUnosMsgNePostoji)
	default:
		var postoji bool
		if stanje, tip, postoji, err = s.robUslugaStanjeStavke(ctx, k.MagaciniID, params.Konto, sifra); err != nil {
			add("sifraartikla", common.ErrMsgReadData)
		} else if !postoji {
			add("sifraartikla", robnoUnosMsgNePostoji)
		} else {
			artikalOK = true
		}
	}
	kolicina := common.StringToFloat64(params.Kolic)
	if strings.TrimSpace(params.Kolic) == "" || kolicina == 0 {
		add("kolicina", common.ErrMsgObavezanPodatak)
	} else if artikalOK && sifra != 0 && tip != "U" && !k.Storno() {
		// The stanje of a changed stavka has its own količina.
		if params.RproID > 0 {
			stara, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(kolic, 0)::float8::text as key from rpro
				where rproid = $1 and sifra = $2 and trim(konto) = $3`, "", []any{params.RproID, sifra, strings.TrimSpace(params.Konto)}, "", "")
			if err == nil && stara != nil && len(*stara) > 0 {
				stanje += math.Abs(common.StringToFloat64((*stara)[0].Key))
			}
		}
		if (kolicina > 0 && stanje <= 0) || robZaokruzi(stanje-kolicina, 3) < 0 {
			add("kolicina", "Nema dovoljno zaliha.")
		}
	}
	if strings.TrimSpace(params.Cena) == "" {
		add("cenausluge", common.ErrMsgObavezanPodatak)
	} else if common.StringToFloat64(params.Cena) <= 0 {
		add("cenausluge", robInterniMsgVeceOdNule)
	}
	return fieldErrors
}

// SaveRobUslugaStavka saves a stavka of a faktura usluga, a new one (RproID 0, the legacy CreateRpro) or a
// changed one (UpdateRPRO), in one transaction:
//   - the cena of the stavka is the cena of the form, in a magacin of the nabavna cena the nabavna cena for an
//     artikal of the zalihe (a šifra other than 0, not of the tip U); the iznos is the količina by it (the cena
//     for the količina 0), negative for a storno, rounded to 2 decimals;
//   - the stavka has the columns of its document, the šifra, konto, naziv (opis usluge) and JM of the form,
//     the cena, the nabavna cena (0 for a usluga), the fakturna cena (the cena of the form), the rabat, the
//     tarifa with its stopa (pdvpct), the cena in the valuta of the document, the poreska kategorija with its
//     osnov oslobođenja, the OTK broj, the serija and the rok; a new one gets the next redni broj;
//   - the nalog, the stanje of the artikal (AzurirajRNAL, AzurirajRSTA) and, for an artikal of the zalihe, the
//     zalihe by lot (AzurirajDRSTA) get the iznos and the količina (a changed stavka the difference); the
//     document gets the iznos and a new stavka counts in it (rdok.brst);
//   - the dodatni opis usluge, when given, is the one of the document (dokusldet).
//
// Then the rate and the obračun of the document (PreracunRATA, ObracunFAKT of the stavke).
func (s *RobnoDokumentaResource) SaveRobUslugaStavka(ctx context.Context, params domain.RobUslugaStavkaParams) (rproID int64, err error) {
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
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	var rsifID int64
	var tip string
	if err = tx.QueryRowContext(ctx, `select rsifid, coalesce(tip, '') from rsif where god = $1 and kar = $2 and sifra = $3 limit 1`,
		userSession.SelectedGod, userSession.SelectedKar, sifra).Scan(&rsifID, &tip); err != nil {
		return 0, err
	}
	usluga := sifra == 0 || strings.TrimSpace(tip) == "U"
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	xCena := cena
	if k.NabavnaCena() && !usluga {
		xCena = ncena
	}
	kolicina := common.StringToFloat64(params.Kolic)
	iznos := cena
	if kolicina != 0 {
		iznos = robZaokruzi(kolicina*xCena, 2)
	}
	iznos, kolicina = robZaokruzi(math.Abs(iznos), 2), math.Abs(kolicina)
	if k.Storno() {
		iznos, kolicina = -iznos, -kolicina
	}
	stopa, err := s.GetRobUslugaStopa(ctx, params.RdokID, common.StringToInt(params.Tarifa))
	if err != nil {
		return 0, err
	}
	cenaVal := common.StringToFloat64(params.CenaVal)
	if k.Kurs > 0 && k.Vaziza > 0 {
		cenaVal = robZaokruzi(cena/k.Kurs/k.Vaziza, 4)
	}
	rok, _ := robnoUnosDatum(params.Rok)

	// A changed stavka: its old values are taken back.
	stara := domain.RobPromenaStanja{}
	if params.RproID > 0 {
		var stariRok float64
		if err = tx.QueryRowContext(ctx, `select coalesce(sifra, 0), coalesce(konto, ''), coalesce(kolic, 0)::float8,
				coalesce(iznos, 0)::float8, coalesce(otk, ''), coalesce(serija, ''), coalesce(roktr, 0)::float8
			from rpro where rproid = $1 and rdokid = $2 for update`, params.RproID, params.RdokID).
			Scan(&stara.Sifra, &stara.Konto, &stara.Kolicina, &stara.Iznos, &stara.Otk, &stara.Serija, &stariRok); err != nil {
			return 0, err
		}
		stara.Roktr = int64(stariRok)
	}

	rproCena, rproNcena := xCena, ncena
	if usluga {
		rproCena, rproNcena = cena, 0
	}
	polja := []domain.Fields{
		{Name: "naz1", Value: strings.TrimSpace(params.Naziv)},
		robPoljeText("jm", params.Jm),
		robPoljeFloat("cena", rproCena),
		robPoljeFloat("ncena", rproNcena),
		robPoljeFloat("fcena", cena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		robPoljeBroj("rab", params.Rab),
		robPoljeInt("po", params.Tarifa),
		robPoljeFloat("pdvpct", stopa.Stopa),
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)},
		robPoljeFloat("cenaval", cenaVal),
		robPoljeText("taxcat", params.Taxcat),
		{Name: "taxexreco", Value: strings.TrimSpace(params.Taxexreco)},
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
			domain.Fields{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
			domain.Fields{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
			robPoljeNullTime("dadok", k.Dadok),
			domain.Fields{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
			domain.Fields{Name: "mag", Value: strconv.Itoa(k.Mag)},
			domain.Fields{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
			domain.Fields{Name: "sifra", Value: strconv.Itoa(sifra)},
			domain.Fields{Name: "konto", Value: konto},
			domain.Fields{Name: "fkto", Value: k.Fkto},
			domain.Fields{Name: "fana", Value: k.Fana},
			domain.Fields{Name: "rbr", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "brst", Value: strconv.Itoa(rbr)},
			domain.Fields{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
			// NOT NULL without a default.
			domain.Fields{Name: "sifartkup", Value: ""},
		)
		if rproID, err = s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", polja); err != nil {
			return 0, err
		}
	}

	// The nalog and the stanje: the old stavka taken back, the new one added.
	razlikaIznosa := iznos - stara.Iznos
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), razlikaIznosa); err != nil {
		return 0, err
	}
	if params.RproID > 0 {
		if err = s.robUslugaAzurirajStanje(ctx, tx, k, domain.RobPromenaStanja{Sifra: stara.Sifra, Konto: stara.Konto, Kolicina: -stara.Kolicina,
			Iznos: -stara.Iznos, Otk: stara.Otk, Serija: stara.Serija, Roktr: stara.Roktr, Brst: -1}, usluga); err != nil {
			return 0, err
		}
	}
	if err = s.robUslugaAzurirajStanje(ctx, tx, k, domain.RobPromenaStanja{Sifra: sifra, Konto: konto, Kolicina: kolicina, Iznos: iznos,
		Otk: strings.TrimSpace(params.Otk), Serija: strings.TrimSpace(params.Serija), Roktr: RobWinDevDatum(rok), Brst: 1}, usluga); err != nil {
		return 0, err
	}

	// The document: its iznos, a new stavka counts in it.
	var dokIznos float64
	var dokBrst int
	if err = tx.QueryRowContext(ctx, `select coalesce(iznos, 0)::float8, coalesce(brst, 0) from rdok where rdokid = $1 for update`,
		params.RdokID).Scan(&dokIznos, &dokBrst); err != nil {
		return 0, err
	}
	if params.RproID == 0 {
		dokBrst++
	}
	if err = s.unos.Dokument.UpdateWithTx(ctx, tx, &domain.RobniDokumentRow{}, "rdokid", params.RdokID, []domain.Fields{
		robPoljeFloat("iznos", robZaokruzi(dokIznos+razlikaIznosa, 2)), {Name: "brst", Value: strconv.Itoa(dokBrst)},
	}); err != nil {
		return 0, err
	}

	// The dodatni opis usluge of the document.
	if opis := strings.TrimSpace(params.OpisDet); opis != "" {
		var postoji int
		if err = tx.QueryRowContext(ctx, `select count(*) from dokusldet where rdokid = $1`, params.RdokID).Scan(&postoji); err != nil {
			return 0, err
		}
		query := `update dokusldet set opisdet = $2 where rdokid = $1`
		if postoji == 0 {
			query = `insert into dokusldet (god, kar, edokid, rdokid, opisdet) values ($3, $4, 0, $1, $2)`
		}
		args := []any{params.RdokID, opis}
		if postoji == 0 {
			args = append(args, userSession.SelectedGod, userSession.SelectedKar)
		}
		if err = s.unos.Dokument.CreateUpdateCustomWithTx(ctx, tx, query, args...); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	if err := s.PreracunRataDokumenta(ctx, params.RdokID); err != nil {
		return rproID, err
	}
	return rproID, s.ObracunFaktureUsluge(ctx, params.RdokID)
}

// robUslugaAzurirajStanje updates the stanje of the artikal of a stavka of a faktura usluga (AzurirajRSTA,
// the cene 0, TZ and TZI "-") and, for an artikal of the zalihe (not a usluga: the šifra 0 or the tip U), the
// zalihe by lot (AzurirajDRSTA, by the tip zaliha of the magacin).
func (s *RobnoDokumentaResource) robUslugaAzurirajStanje(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, p domain.RobPromenaStanja, usluga bool) error {
	p.MagaciniID, p.Danal, p.DokvrstaID, p.Tipdok, p.Tz, p.Tzi = k.MagaciniID, k.Danal.Time, k.DokvrstaID, k.Tipdok, "-", "-"
	if usluga || p.Sifra == 0 {
		return s.AzurirajRSTA(ctx, tx, p)
	}
	return s.robAzurirajStanjeStavke(ctx, tx, k, p)
}

// DeleteRobUslugaStavka deletes a stavka of a faktura usluga (the legacy BrisiDOKUSLDET and BrisiRPRO) in one
// transaction and returns its document: the dodatni opis usluge of the document is deleted, the stavka is
// taken back from its nalog, from the stanje and (an artikal of the zalihe) from the zalihe by lot, the
// document loses its iznos, the stavka is deleted and the stavke after it get the redni brojevi again (rbr).
// Then the rate and the obračun of the document (PreracunRATA, ObracunFAKT of the stavke).
func (s *RobnoDokumentaResource) DeleteRobUslugaStavka(ctx context.Context, rproID int64) (rdokID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var sifra, rbr int
	var konto, otk, serija, tip string
	var kolicina, iznos, roktr float64
	if err = tx.QueryRowContext(ctx, `select coalesce(p.rdokid, 0)::bigint, coalesce(p.sifra, 0), coalesce(p.konto, ''),
			coalesce(p.kolic, 0)::float8, coalesce(p.iznos, 0)::float8, coalesce(p.otk, ''), coalesce(p.serija, ''),
			coalesce(p.roktr, 0)::float8, coalesce(p.rbr, 0), coalesce((select a.tip from rsif a where a.rsifid = p.rsifid), '')
		from rpro p where p.rproid = $1 for update`, rproID).
		Scan(&rdokID, &sifra, &konto, &kolicina, &iznos, &otk, &serija, &roktr, &rbr, &tip); err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("stavka ne postoji")
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
	if err = s.unos.Dokument.CreateUpdateCustomWithTx(ctx, tx, `delete from dokusldet where rdokid = $1`, rdokID); err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos); err != nil {
		return 0, err
	}
	if err = s.robUslugaAzurirajStanje(ctx, tx, k, domain.RobPromenaStanja{Sifra: sifra, Konto: konto, Kolicina: -kolicina, Iznos: -iznos,
		Otk: otk, Serija: serija, Roktr: int64(roktr), Brst: -1}, sifra == 0 || strings.TrimSpace(tip) == "U"); err != nil {
		return 0, err
	}
	if err = s.unos.Dokument.CreateUpdateCustomWithTx(ctx, tx, `update rdok set iznos = round((coalesce(iznos, 0) - $2)::numeric, 2)
		where rdokid = $1`, rdokID, iznos); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return 0, err
	}
	if err = s.unos.Rpro.CreateUpdateCustomWithTx(ctx, tx, `update rpro p set rbr = x.n
		from (select rproid, ($2 - 1 + row_number() over (order by rbr, rproid))::int as n from rpro where rdokid = $1 and rbr > $2) x
		where p.rproid = x.rproid`, rdokID, rbr); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	if err := s.PreracunRataDokumenta(ctx, rdokID); err != nil {
		return rdokID, err
	}
	return rdokID, s.ObracunFaktureUsluge(ctx, rdokID)
}

// GetRobUslugaStavkaForma returns a saved stavka of a faktura usluga for the form of the stavke (the legacy
// "Izmene" of rob_FAKTURE_Usluga_DET2): its naziv (opis usluge), JM, šifra, konto, količina, cena (the
// fakturna cena), rabat, tarifa with its stopa on the datum of the document, poreska kategorija with its osnov
// oslobođenja, nabavna cena, OTK broj, serija and rok, and the dodatni opis usluge of the document (its text:
// the legacy kept it as RTF, common.RtfToText). found is false without it.
func (s *RobnoDokumentaResource) GetRobUslugaStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobUslugaStavkaForma, bool, error) {
	f := domain.RobUslugaStavkaForma{}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(r.sifra, 0)::text as key, coalesce(r.naz1, '') as c1,
			(coalesce(r.konto, '') || '|' || coalesce(r.jm, '') || '|' || coalesce(r.taxcat, '') || '|' || coalesce(r.taxexreco, '') || '|' ||
				coalesce(r.otk, '') || '|' || coalesce(r.serija, '')) as c2,
			coalesce(r.po, 0)::text as c3, `+fmt.Sprintf(robPoreskaStopaSQL, "r.po", "d.dadok")+`::text as c4,
			(abs(coalesce(r.kolic, 0))::float8::text || '|' || coalesce(r.fcena, 0)::float8::text || '|' || coalesce(r.rab, 0)::float8::text || '|' ||
				coalesce(r.ncena, 0)::float8::text) as c5,
			coalesce(r.roktr, 0)::bigint::text as c6,
			coalesce((select u.opisdet from dokusldet u where u.rdokid = r.rdokid order by u.dokusldetid limit 1), '') as c7
		from rpro r join rdok d on d.rdokid = r.rdokid where r.rproid = $1 and r.rdokid = $2`, "", []any{rproID, rdokID}, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return f, false, err
	}
	r := (*rows)[0]
	f.RproID, f.Sifra, f.Naziv = rproID, common.StringToInt(r.Key), r.C1
	if t := strings.Split(r.C2, "|"); len(t) == 6 {
		f.Konto, f.Jm, f.Taxcat, f.Taxexreco = strings.TrimSpace(t[0]), strings.TrimSpace(t[1]), strings.TrimSpace(t[2]), strings.TrimSpace(t[3])
		f.Otk, f.Serija = strings.TrimSpace(t[4]), strings.TrimSpace(t[5])
	}
	f.Tarifa, f.Stopa = common.StringToInt(r.C3), common.StringToFloat64(r.C4)
	if n := strings.Split(r.C5, "|"); len(n) == 4 {
		f.Kolicina, f.Cena, f.Rabat, f.Ncena = common.StringToFloat64(n[0]), common.StringToFloat64(n[1]), common.StringToFloat64(n[2]),
			common.StringToFloat64(n[3])
	}
	if roktr, _ := strconv.ParseInt(r.C6, 10, 64); roktr > 0 {
		f.Rok = robnoStampaWinDevDatumFormat(roktr, common.HtmlLayout)
	}
	f.OpisDet = common.RtfToText(r.C7)
	return f, true, nil
}

// RobUslugeStavkeTableRows returns the stavke of a faktura usluga as the rows of the grid of the stavke of the
// screen (the columns of GetUslugeStavkeTableFields: redni broj, konto, šifra, opis usluge, JM, količina,
// cena, rabat, iznos, porez, tarifa, poreska kategorija, osnov oslobođenja).
func RobUslugeStavkeTableRows(stavke []domain.RobStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{
			ID: strconv.FormatInt(st.RproID, 10),
			Fields: []string{strconv.Itoa(st.Rbr), st.Konto, strconv.Itoa(st.Sifra), st.Naziv, st.Jm,
				common.FormatNumberWithSystemLocale(st.Kolic, 3), robFormatIznos(st.Cena), robFormatIznos(st.Rab), robFormatIznos(st.Iznos),
				robFormatIznos(st.Stopa), strconv.Itoa(st.Po), st.Taxcat, st.Taxexreco},
		})
	}
	return rows
}

// GetVrsteGrupeDokumenta returns the vrste dokumenata of the group of the vrsta vrd (dokvrsta.grpdok) between
// vrdOd and vrdDo, like the legacy QRY_RDOK_TIP of rob_Fakture_Usluga (the grid of the documents of the tab
// "Pregled faktura").
func (s *RobnoDokumentaResource) GetVrsteGrupeDokumenta(ctx context.Context, vrd, vrdOd, vrdDo int) ([]int, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select v.vrd::text as key from dokvrsta v
		where v.god = $1 and v.kar = $2 and v.vrd between $4 and $5
			and upper(trim(coalesce(v.grpdok, ''))) = (select upper(trim(coalesce(g.grpdok, ''))) from dokvrsta g
				where g.god = $1 and g.kar = $2 and g.vrd = $3 limit 1)
		order by v.vrd`, "", []any{userSession.SelectedGod, userSession.SelectedKar, vrd, vrdOd, vrdDo}, "", "")
	if err != nil || rows == nil {
		return nil, err
	}
	vrste := make([]int, 0, len(*rows))
	for _, r := range *rows {
		vrste = append(vrste, common.StringToInt(r.Key))
	}
	return vrste, nil
}

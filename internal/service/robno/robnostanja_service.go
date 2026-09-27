package robno

import (
	"context"
	"fmt"
	"slices"

	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

type RobnoStanjaService interface {
	GetStanjePojedinacnogArtikla(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams, string) error
	GetUkupnaObrada(context.Context, *domain.RobnoStanjaTotal, domain.RobnoStanjaParams) error
	GetStanjaViseArtikala(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaViseArtikalaSifra(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams, string) error
	GetStanjaViseArtikalaGrupa(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams, string) error
	GetStanjaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, totalValues *domain.RobnoStanjaTotal, getTotals bool, params domain.RobnoStanjaParams, printType string) error
	GetSvodjenjeZaliha(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetFvrData(context.Context) (domain.Fvr, error)
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	GetTipDokComboValues(context.Context) ([]domain.ComboItem, error)
	GetOjComboValues(context.Context) ([]domain.ComboItem, error)
	GetMestoTroskaComboValues(context.Context, int) ([]domain.ComboItem, error)
	SetDefaultTableData(tbl *domain.TableData)
	GetPojedinacnogArtiklaTableFields() []domain.Fields
	GetViseArtikalaTableFields() []domain.Fields
	GetSubsintetiskogKontaTableFields() []domain.Fields
	GetSvodjenjeZalihaTableFields() []domain.Fields
}

type RobnoStanjaResource struct {
	robnostanjeRepo                  repository.BaseRepository[domain.RobnoStanjeDto]
	rproRepo                         repository.BaseRepository[domain.Rpro]
	drstaRepo                        repository.BaseRepository[domain.Drsta]
	magRepo                          repository.BaseRepository[domain.Magacini]
	tipdokRepo                       repository.BaseRepository[domain.Tipdok]
	ojRepo                           repository.BaseRepository[domain.Orgjed]
	mestoTroskaRepo                  repository.BaseRepository[domain.Mestotr]
	fvrRepo                          repository.BaseRepository[domain.Fvr]
	commonSvc                        commonsvc.CommonService
	pojedinacniArtikalTableFields    []domain.Fields
	viseArtikalaTableFields          []domain.Fields
	viseArtikalaStampaTableFields    []domain.Fields
	viseArtikalaStampaFinTableFields []domain.Fields
	subsintetickogKontaTableFields   []domain.Fields
	svodjenjeZalihaTableFields       []domain.Fields
}

func NewRobnoStanjaService(robnostanjeRepo repository.BaseRepository[domain.RobnoStanjeDto], rproRepo repository.BaseRepository[domain.Rpro], drstaRepo repository.BaseRepository[domain.Drsta], magRepo repository.BaseRepository[domain.Magacini], tipdokRepo repository.BaseRepository[domain.Tipdok], ojRepo repository.BaseRepository[domain.Orgjed], mestoTroskaRepo repository.BaseRepository[domain.Mestotr], fvrRepo repository.BaseRepository[domain.Fvr], commonSvc commonsvc.CommonService) *RobnoStanjaResource {
	rs := &RobnoStanjaResource{
		robnostanjeRepo: robnostanjeRepo,
		rproRepo:        rproRepo,
		drstaRepo:       drstaRepo,
		magRepo:         magRepo,
		tipdokRepo:      tipdokRepo,
		ojRepo:          ojRepo,
		mestoTroskaRepo: mestoTroskaRepo,
		fvrRepo:         fvrRepo,
		commonSvc:       commonSvc,
	}
	rs.setTableFileds()
	return rs
}

func (s *RobnoStanjaResource) GetStanjePojedinacnogArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoStanjaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pojedinacniArtikalTableFields

	hasGod, hasKar := s.drstaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select rsif.sifra, rsif.jm, drsta.cena,
		mdug_1, mdug_2, mdug_3, mdug_4, mdug_5, mdug_6, mdug_7, mdug_8, mdug_9, mdug_10, mdug_11, mdug_12,
		mpot_1, mpot_2, mpot_3, mpot_4, mpot_5, mpot_6, mpot_7, mpot_8, mpot_9, mpot_10, mpot_11, mpot_12, 
		mul_1, mul_2, mul_3, mul_4, mul_5, mul_6, mul_7, mul_8, mul_9, mul_10, mul_11, mul_12, 
		miz_1, miz_2, miz_3, miz_4, miz_5, miz_6, miz_7, miz_8, miz_9, miz_10, miz_11, miz_12 from drsta`, true)
	qb.AddJoin(" inner join rsif on rsif.rsifid = drsta.rsifid")
	if hasGod {
		qb.AddEqual("drsta.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("drsta.kar", userSession.SelectedKar)
	}
	if params.Magacin != 0 {
		qb.AddEqual("drsta.magaciniid", params.MagaciniID)
	}
	if params.Konto != "" {
		qb.AddEqual("drsta.konto", params.Konto)
	}
	if params.SifraArtikla != "" {
		qb.AddEqual("rsif.sifra", params.SifraArtikla)
	}
	qb.AddOrderBy("rsif.sifra")
	if !getTotalRecords && printType == common.TipStampePreview {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	months := common.GetMontshName()
	rowsPerArticle := len(months)
	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities)*rowsPerArticle, pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		ulaz := []float64{entity.MUl1, entity.MUl2, entity.MUl3, entity.MUl4, entity.MUl5, entity.MUl6, entity.MUl7, entity.MUl8, entity.MUl9, entity.MUl10, entity.MUl11, entity.MUl12}
		izlaz := []float64{entity.MIz1, entity.MIz2, entity.MIz3, entity.MIz4, entity.MIz5, entity.MIz6, entity.MIz7, entity.MIz8, entity.MIz9, entity.MIz10, entity.MIz11, entity.MIz12}
		duguje := []float64{entity.MDug1, entity.MDug2, entity.MDug3, entity.MDug4, entity.MDug5, entity.MDug6, entity.MDug7, entity.MDug8, entity.MDug9, entity.MDug10, entity.MDug11, entity.MDug12}
		potrazuje := []float64{entity.MPot1, entity.MPot2, entity.MPot3, entity.MPot4, entity.MPot5, entity.MPot6, entity.MPot7, entity.MPot8, entity.MPot9, entity.MPot10, entity.MPot11, entity.MPot12}
		for i, month := range months {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					month,
					common.FormatNumberWithSystemLocale(ulaz[i], 3),
					common.FormatNumberWithSystemLocale(izlaz[i], 3),
					common.FormatNumberWithSystemLocale(duguje[i], 2),
					common.FormatNumberWithSystemLocale(potrazuje[i], 2),
				},
				HasUpdate: false,
				HasDelete: false,
			})
		}
	}
	return nil
}
func (s *RobnoStanjaResource) GetUkupnaObrada(ctx context.Context, total *domain.RobnoStanjaTotal, params domain.RobnoStanjaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select drsta.ulaz, drsta.izlaz, drsta.dug as duguje, drsta.pot as potrazuje,
		rsif.sifra, coalesce(rsif.naziv, '') as naziv_artikla, coalesce(rsif.jm, '') as jm,
		coalesce(drsta.cena, 0) as cena, coalesce(fkpl.naziv, '') as kontonaziv, coalesce(magacini.opis, '') as magacinnaziv
		from drsta`, true)
	qb.AddJoin("inner join rsif on rsif.rsifid = drsta.rsifid")
	qb.AddJoin("left join magacini on magacini.magaciniid = drsta.magaciniid")
	qb.AddJoin("left join fkpl on fkpl.god = drsta.god and fkpl.kar = drsta.kar and fkpl.vkonta = 2 and fkpl.konto = drsta.konto")
	if hasGod {
		qb.AddEqual("drsta.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("drsta.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("drsta.magaciniid", params.MagaciniID)
	}
	if params.Konto != "" {
		qb.AddEqual("drsta.konto", params.Konto)
	}
	if params.SifraArtikla != "" {
		qb.AddEqual("rsif.sifra", params.SifraArtikla)
	}
	qb.AddOrderBy("rsif.sifra")
	sqlQuery, args := qb.Build()
	// Execute the query and populate the total object
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if entities == nil || len(*entities) == 0 {
		return nil
	}
	entity := (*entities)[0]
	total.Ulaz = entity.Ulaz
	total.Izlaz = entity.Izlaz
	total.Stanje = entity.Ulaz - entity.Izlaz
	total.Duguje = entity.Duguje
	total.Potrazuje = entity.Potrazuje
	total.Saldo = entity.Duguje - entity.Potrazuje
	total.Sifra = entity.Sifra
	total.NazivArtikla = entity.NazivArtikla
	total.Jm = entity.Jm
	total.Cena = entity.Cena
	total.KontoNaziv = entity.KontoNaziv
	total.MagacinNaziv = entity.MagacinNaziv
	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikala(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikala

	return nil
}
func (s *RobnoStanjaResource) GetStanjaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, totalValues *domain.RobnoStanjaTotal, getTotals bool, params domain.RobnoStanjaParams, printType string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	// First query: the opening balance (documents with tipdok = '00') summed into ONE row with
	// mesec = 0, so the report shows it as the single "Početno stanje" line. There is no GROUP BY,
	// the aggregate collapses the whole year into a single row.
	qb1 := common.NewQueryBuilder(`SELECT 0 as mesec,
			COALESCE(SUM(CASE WHEN upper(dokvrsta.kodknj) = 'D' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END), 0) AS duguje,
			COALESCE(SUM(CASE WHEN upper(dokvrsta.kodknj) = 'P' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END), 0) AS potrazuje
    FROM rpro `, true)
	qb1.AddJoin(" inner join rnal on rnal.rnalid = rpro.rnalid")
	qb1.AddJoin(" inner join dokvrsta on dokvrsta.god = rpro.god and dokvrsta.kar = rpro.kar and dokvrsta.vrd = rpro.vrd")
	// Add conditions to first query
	hasGod, hasKar := s.robnostanjeRepo.GetHasGodHasKar()
	if hasGod {
		qb1.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb1.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	qb1.AddEqual("rnal.tipdok", "00")
	qb1.AddLike("rpro.konto", params.Konto)
	qb1.AddEqual("rnal.magaciniid", params.MagaciniID)
	// Second query: the monthly data (tipdok != '00') summed per month - one row per month with
	// duguje/potrazuje, which SUM() together with the GROUP BY collapses the documents of that month.
	qb2 := common.NewQueryBuilder(`SELECT EXTRACT(MONTH FROM rnal.danal) as mesec,
			COALESCE(SUM(CASE WHEN upper(dokvrsta.kodknj) = 'D' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END), 0) AS duguje,
			COALESCE(SUM(CASE WHEN upper(dokvrsta.kodknj) = 'P' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END), 0) AS potrazuje
    FROM rpro `, true)
	qb2.AddJoin(" inner join rnal on rnal.rnalid = rpro.rnalid")
	qb2.AddJoin(" inner join dokvrsta on dokvrsta.god = rpro.god and dokvrsta.kar = rpro.kar and dokvrsta.vrd = rpro.vrd")
	// Add same base conditions
	if hasGod {
		qb2.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb2.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	qb2.AddEqual("rnal.magaciniid", params.MagaciniID)
	qb2.AddEqual("rpro.konto", params.Konto)
	qb2.AddCondition("rnal.tipdok", "00", "!=")
	qb2.AddGroupBy("EXTRACT(MONTH FROM rnal.danal)")
	// Create UNION
	uqb := common.NewUnionQueryBuilder("UNION ALL")
	uqb.AddQuery(qb1)
	uqb.AddQuery(qb2)
	uqb.AddOrderBy("mesec")

	sqlQuery, args := uqb.Build()
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotals {
		for i, entity := range *entities {
			if i == 0 {
				totalValues.PocStanjeDug = entity.Duguje
				totalValues.PocStanjePot = entity.Potrazuje
				totalValues.PocStanjeSaldo = entity.Duguje - entity.Potrazuje
			} else {
				totalValues.TekuciPromDug += entity.Duguje
				totalValues.TekuciPromPot += entity.Potrazuje
			}
		}
		totalValues.TekuciPromSaldo = totalValues.TekuciPromDug - totalValues.TekuciPromPot
		totalValues.UkPromDug = totalValues.PocStanjeDug + totalValues.TekuciPromDug
		totalValues.UkPromPot = totalValues.PocStanjePot + totalValues.TekuciPromPot
		totalValues.UkPromSaldo = totalValues.UkPromDug - totalValues.UkPromPot
		return nil
	}
	// Create template data with opening balance and all 12 months initialized to 0
	templateData := make([]domain.SaldaDto, 13) // 0 (opening balance) + 12 months

	// Merge actual data into template - update the corresponding month rows with real values
	kumulSaldo := 0.0
	if entities != nil {
		for _, entity := range *entities {
			if entity.Mesec >= 0 && entity.Mesec <= 12 {
				// Monthly data
				saldoDto := domain.SaldaDto{
					Mesec:     entity.Mesec,
					Duguje:    entity.Duguje,
					Potrazuje: entity.Potrazuje,
					Saldo:     entity.Duguje - entity.Potrazuje,
				}
				kumulSaldo = kumulSaldo + (entity.Duguje - entity.Potrazuje)
				templateData[entity.Mesec] = saldoDto
				templateData[entity.Mesec].SaldoKumul = kumulSaldo
			}
		}
	}
	monthNames := []string{}
	monthNames = append(monthNames, "Pocetno stanje")
	monthNames = append(monthNames, common.GetMontshName()...)
	for i, salda := range templateData {
		tbl.Rows[i] = domain.TableRow{
			Fields: []string{
				monthNames[i],
				common.FormatNumberWithSystemLocale(salda.Duguje, 2),
				common.FormatNumberWithSystemLocale(salda.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(salda.Saldo, 2),
				common.FormatNumberWithSystemLocale(salda.SaldoKumul, 2),
			},
		}
	}
	return nil
}

// legacyOnlyWithMovementPIB is the PIB of the company for which the legacy report listed only
// the articles that had some movement in the period (WinDev: "privremeno zbog Triangl papira").
const legacyOnlyWithMovementPIB = "100110650"

// mesecRange turns the "od meseca" / "do meseca" form values into a valid 1..12 range.
// The form ships 1..99, where 99 means "through the whole year".
func mesecRange(odMeseca, doMeseca string) (int, int) {
	od := common.StringToInt(odMeseca)
	do := common.StringToInt(doMeseca)
	if od < 1 || od > 12 {
		od = 1
	}
	if do < 1 || do > 12 {
		do = 12
	}
	if do < od {
		// A reversed range would sum nothing, so sum the months the user did select.
		od, do = do, od
	}
	return od, do
}

// zbirMeseca sums the monthly quantities of one article for the months od..do.
// It mirrors the WinDev loop "FOR i = odMeseca TO doMeseca" over MUL/MIZ/MDUG/MPOT.
func zbirMeseca(ent domain.RobnoStanjeDto, odMesec, doMesec int) (ulaz, izlaz, dug, pot float64) {
	mul := []float64{ent.MUl1, ent.MUl2, ent.MUl3, ent.MUl4, ent.MUl5, ent.MUl6, ent.MUl7, ent.MUl8, ent.MUl9, ent.MUl10, ent.MUl11, ent.MUl12}
	miz := []float64{ent.MIz1, ent.MIz2, ent.MIz3, ent.MIz4, ent.MIz5, ent.MIz6, ent.MIz7, ent.MIz8, ent.MIz9, ent.MIz10, ent.MIz11, ent.MIz12}
	mdug := []float64{ent.MDug1, ent.MDug2, ent.MDug3, ent.MDug4, ent.MDug5, ent.MDug6, ent.MDug7, ent.MDug8, ent.MDug9, ent.MDug10, ent.MDug11, ent.MDug12}
	mpot := []float64{ent.MPot1, ent.MPot2, ent.MPot3, ent.MPot4, ent.MPot5, ent.MPot6, ent.MPot7, ent.MPot8, ent.MPot9, ent.MPot10, ent.MPot11, ent.MPot12}
	for m := odMesec; m <= doMesec; m++ {
		ulaz += mul[m-1]
		izlaz += miz[m-1]
		dug += mdug[m-1]
		pot += mpot[m-1]
	}
	return ulaz, izlaz, dug, pot
}

// cenaArtikla reproduces the WinDev price selection for one report row:
//   - CENA, or PROSNC when the magazine keeps stock at purchase price,
//   - "Prosecna cena - stanje": (dug - pot) / (ulaz - izlaz), falling back to dug / ulaz
//     when there is no stock, and to 0 when there is nothing at all,
//   - "Prosecna cena - ulaz": dug / ulaz (applied last, exactly like the legacy code).
func cenaArtikla(ent domain.RobnoStanjeDto, nabavnaCena bool, ulaz, izlaz, dug, pot float64, params domain.RobnoStanjaParams) float64 {
	cena := ent.Cena
	if nabavnaCena {
		cena = ent.Prosnc
	}
	stanje := ulaz - izlaz
	if params.ProsecnaCenaStanje {
		switch {
		case stanje != 0:
			cena = (dug - pot) / stanje
		case ulaz != 0:
			cena = dug / ulaz
		default:
			cena = 0
		}
	}
	if params.ProsecnaCenaUlaz && ulaz != 0 {
		cena = dug / ulaz
	}
	return cena
}

// magacinJeNabavnaCena reports whether the selected magazine keeps stock at purchase price
// (MAGACINI.NACVODZAL = 3). The legacy code seeks MAGACINI before running the report.
func (s *RobnoStanjaResource) magacinJeNabavnaCena(ctx context.Context, session *domain.UserSession, magaciniID int) bool {
	if magaciniID == 0 {
		return false
	}
	hasGod, hasKar := s.magRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select nacvodzal from magacini", true)
	qb.AddEqual("magaciniid", magaciniID)
	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.magRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil || entities == nil || len(*entities) == 0 {
		return false
	}
	return (*entities)[0].Nacvodzal == 3
}

// samoSaPrometom reports whether the report must hide articles without any movement.
func (s *RobnoStanjaResource) samoSaPrometom(ctx context.Context) bool {
	fvr, err := common.GetFvrData(ctx, &s.fvrRepo)
	if err != nil {
		return false
	}
	return fvr.PIB == legacyOnlyWithMovementPIB
}

// legacyArticleMovementCondition keeps only the articles whose stock record has some movement.
// The legacy reports used it with the comment "privremeno zbog Triangl papira".
const legacyArticleMovementCondition = "(COALESCE(drsta.ulaz, 0) <> 0 OR COALESCE(drsta.izlaz, 0) <> 0 OR COALESCE(drsta.dug, 0) <> 0 OR COALESCE(drsta.pot, 0) <> 0)"

// robnoStanjeArtikliColumns are the drsta/rsif columns both "Prikaz stanja više artikala"
// reports read: the article identification, the two prices and the 12 monthly quantities/amounts.
const robnoStanjeArtikliColumns = `drsta.konto, drsta.sifra, drsta.cena, drsta.prosnc,
		coalesce(rsif.gru, 0) as gru,
		coalesce(rsif.naziv, '') as naziv_artikla, coalesce(rsif.jm, '') as jm,
		drsta.mul_1, drsta.mul_2, drsta.mul_3, drsta.mul_4, drsta.mul_5, drsta.mul_6,
		drsta.mul_7, drsta.mul_8, drsta.mul_9, drsta.mul_10, drsta.mul_11, drsta.mul_12,
		drsta.miz_1, drsta.miz_2, drsta.miz_3, drsta.miz_4, drsta.miz_5, drsta.miz_6,
		drsta.miz_7, drsta.miz_8, drsta.miz_9, drsta.miz_10, drsta.miz_11, drsta.miz_12,
		drsta.mdug_1, drsta.mdug_2, drsta.mdug_3, drsta.mdug_4, drsta.mdug_5, drsta.mdug_6,
		drsta.mdug_7, drsta.mdug_8, drsta.mdug_9, drsta.mdug_10, drsta.mdug_11, drsta.mdug_12,
		drsta.mpot_1, drsta.mpot_2, drsta.mpot_3, drsta.mpot_4, drsta.mpot_5, drsta.mpot_6,
		drsta.mpot_7, drsta.mpot_8, drsta.mpot_9, drsta.mpot_10, drsta.mpot_11, drsta.mpot_12`

// robnoStanjeArtikliBaseQuery is the FROM part shared by both reports.
const robnoStanjeArtikliBaseQuery = "select " + robnoStanjeArtikliColumns + " from drsta"

// robnoStanjeArtikalRow builds one grid row of the "više artikala" reports.
func robnoStanjeArtikalRow(ent domain.RobnoStanjeDto, cena, ulaz, izlaz float64) domain.TableRow {
	return domain.TableRow{
		Fields: []string{
			ent.Konto,
			ent.Sifra,
			ent.NazivArtikla,
			ent.Jm,
			common.FormatNumberWithSystemLocale(cena, 2),
			common.FormatNumberWithSystemLocale(ulaz, 3),
			common.FormatNumberWithSystemLocale(izlaz, 3),
			common.FormatNumberWithSystemLocale(ulaz-izlaz, 3),
		},
		HasUpdate: false,
		HasDelete: false,
	}
}

// robnoStanjeArtikalStampaRow builds one line of the printed "Prikaz stanja više artikala" report:
// the grid columns plus the running number; the financial columns are appended when the
// "Finansijski iznos" option is on.
func robnoStanjeArtikalStampaRow(rbr int, ent domain.RobnoStanjeDto, cena, ulaz, izlaz, dug, pot float64, finansijski bool) domain.TableRow {
	fields := []string{
		fmt.Sprintf("%d", rbr),
		ent.Konto,
		ent.Sifra,
		ent.NazivArtikla,
		ent.Jm,
		common.FormatNumberWithSystemLocale(cena, 2),
		common.FormatNumberWithSystemLocale(ulaz, 3),
		common.FormatNumberWithSystemLocale(izlaz, 3),
		common.FormatNumberWithSystemLocale(ulaz-izlaz, 3),
	}
	if finansijski {
		fields = append(fields,
			common.FormatNumberWithSystemLocale(dug, 2),
			common.FormatNumberWithSystemLocale(pot, 2),
			common.FormatNumberWithSystemLocale(dug-pot, 2))
	}
	return domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
}

// robnoStanjeGrupaHeaderRow builds the full-width "Grupa: N" band of the grouped print.
// classRow is either "group-header" or "group-header-new-page" ("Nova strana po grupi").
func robnoStanjeGrupaHeaderRow(grupa int, classRow string) domain.TableRow {
	return domain.TableRow{
		ClassRow:  classRow,
		Fields:    []string{fmt.Sprintf("%s: %d", i18n.GetInstance().Label("Grupa"), grupa)},
		HasUpdate: false,
		HasDelete: false,
	}
}

// robnoStanjeArtikalStampaTotalRow builds the "Ukupno" line at the end of the printed report.
func robnoStanjeArtikalStampaTotalRow(ulaz, izlaz, dug, pot float64, finansijski bool) domain.TableRow {
	fields := []string{
		"",
		"",
		"",
		i18n.GetInstance().Label("Ukupno"),
		"",
		"",
		common.FormatNumberWithSystemLocale(ulaz, 3),
		common.FormatNumberWithSystemLocale(izlaz, 3),
		common.FormatNumberWithSystemLocale(ulaz-izlaz, 3),
	}
	if finansijski {
		fields = append(fields,
			common.FormatNumberWithSystemLocale(dug, 2),
			common.FormatNumberWithSystemLocale(pot, 2),
			common.FormatNumberWithSystemLocale(dug-pot, 2))
	}
	return domain.TableRow{ClassRow: "group-total", Fields: fields, HasUpdate: false, HasDelete: false}
}

// setRobnoStanjeArtikliTotals fills the table footer with the sum of the loaded rows; only the
// columns flagged with IncludeInTotals (ulaz, izlaz, stanje) are filled.
func setRobnoStanjeArtikliTotals(tbl *domain.TableData, ulaz, izlaz float64) {
	ukupno := map[string]string{
		"ulaz":   common.FormatNumberWithSystemLocale(ulaz, 3),
		"izlaz":  common.FormatNumberWithSystemLocale(izlaz, 3),
		"stanje": common.FormatNumberWithSystemLocale(ulaz-izlaz, 3),
	}
	tbl.Totals = make([]string, len(tbl.Headers))
	if len(tbl.Totals) > 0 {
		tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
	}
	for i, header := range tbl.Headers {
		if value, ok := ukupno[header.Name]; ok {
			tbl.Totals[i] = value
		}
	}
}

// GetStanjaViseArtikalaSifra renders "Prikaz stanja više artikala" (variant "po šifri"): one row
// per article with the summed quantities and the price for the selected period.
//
//  1. reads MAGACINI for the selected magazine to know whether the stock is kept at purchase
//     price (NACVODZAL = 3) and therefore PROSNC has to be shown instead of CENA,
//  2. selects the articles from RSTA (drsta) joined with RSIF (rsif) for the current
//     god/kar/magacin and the "od/do konta" and "od/do šifre" ranges,
//  3. sums the monthly columns MUL/MIZ/MDUG/MPOT from the "od meseca" to the
//     "zaključno sa mesecom" month,
//  4. prices the row with CENA/PROSNC or with one of the two "prosečna cena" averages.
func (s *RobnoStanjaResource) GetStanjaViseArtikalaSifra(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, page, size)
	// The printed report uses its own columns (running number, konto, šifra, naziv, JM, cena, ulaz,
	// izlaz, stanje and, when "Finansijski iznos" is on, duguje, potražuje, saldo) and is never paged.
	stampa := printType == common.TipStampePrint
	finansijski := stampa && params.FinansijskiIznos
	switch {
	case !stampa:
		tbl.Headers = s.viseArtikalaTableFields
	case finansijski:
		tbl.Headers = s.viseArtikalaStampaFinTableFields
	default:
		tbl.Headers = s.viseArtikalaStampaTableFields
	}

	nabavnaCena := s.magacinJeNabavnaCena(ctx, userSession, params.Magacin)
	odMesec, doMesec := mesecRange(params.OdMeseca, params.DoMeseca)

	hasGod, hasKar := s.drstaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(robnoStanjeArtikliBaseQuery, true)
	qb.AddJoin(" inner join rsif on rsif.rsifid = drsta.rsifid")
	if hasGod {
		qb.AddEqual("drsta.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("drsta.kar", userSession.SelectedKar)
	}
	qb.AddEqual("drsta.magaciniid", params.Magacin)
	qb.AddCondition("drsta.konto", params.OdKonta, ">=")
	qb.AddCondition("drsta.konto", params.DoKonta, "<=")
	qb.AddCondition("drsta.sifra", params.OdSifre, ">=")
	qb.AddCondition("drsta.sifra", params.DoSifre, "<=")
	// "Za dobavljaca": only articles linked to a supplier (legacy: RSIF.KDOB = <sifra dobavljaca>).
	if params.ZaDobavljaca {
		qb.AddCustomCondition("COALESCE(rsif.kdob, '') <> ''")
	}
	// "Artikli sa stanjem" / "Artikli bez stanja" (legacy: RSTA.ULAZ - RSTA.IZLAZ).
	if params.ArtikliSaStanjem && !params.ArtikliBezStanja {
		qb.AddCustomCondition("COALESCE(drsta.ulaz, 0) - COALESCE(drsta.izlaz, 0) <> 0")
	}
	if !params.ArtikliSaStanjem && params.ArtikliBezStanja {
		qb.AddCustomCondition("COALESCE(drsta.ulaz, 0) - COALESCE(drsta.izlaz, 0) = 0")
	}
	// Legacy: for one company (PIB) the report listed only articles with a movement.
	if s.samoSaPrometom(ctx) {
		qb.AddCustomCondition(legacyArticleMovementCondition)
	}
	qb.AddOrderBy("drsta.sifra")
	if !total && !stampa && size > 0 {
		qb.SetLimit(size)
		qb.SetOffset((page - 1) * size)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if total && !stampa {
		common.SetTableTotalRecords(tbl, len(*entities), size)
		return nil
	}

	tbl.Rows = []domain.TableRow{}
	var ukupnoUlaz, ukupnoIzlaz, ukupnoDug, ukupnoPot float64
	rbr := 0
	for _, entity := range *entities {
		ulaz, izlaz, dug, pot := zbirMeseca(entity, odMesec, doMesec)
		cena := cenaArtikla(entity, nabavnaCena, ulaz, izlaz, dug, pot, params)
		ukupnoUlaz += ulaz
		ukupnoIzlaz += izlaz
		ukupnoDug += dug
		ukupnoPot += pot
		rbr++
		if stampa {
			tbl.Rows = append(tbl.Rows, robnoStanjeArtikalStampaRow(rbr, entity, cena, ulaz, izlaz, dug, pot, finansijski))
		} else {
			tbl.Rows = append(tbl.Rows, robnoStanjeArtikalRow(entity, cena, ulaz, izlaz))
		}
	}

	if stampa {
		if rbr > 0 {
			tbl.Rows = append(tbl.Rows, robnoStanjeArtikalStampaTotalRow(ukupnoUlaz, ukupnoIzlaz, ukupnoDug, ukupnoPot, finansijski))
		}
		return nil
	}

	setRobnoStanjeArtikliTotals(tbl, ukupnoUlaz, ukupnoIzlaz)
	return nil
}

// GetStanjaViseArtikalaGrupa renders "Prikaz stanja više artikala" (variant "po grupi"): one row
// per article of the selected groups, ordered by group and article code.
//
// It is the Go/PostgreSQL translation of the WinDev procedure, which
//  1. reads MAGACINI to know whether the stock is kept at purchase price (NACVODZAL = 3) and
//     therefore PROSNC has to be shown instead of CENA,
//  2. runs the stored query "QRY_rob_STANJE_ARTIKLA_PO_GRUPI" for god/kar, the group range and
//     the article-code range, i.e. the article list of those groups,
//  3. for every article reads RSTA (drsta) for the selected magazine and keeps the row only when
//     the stock record exists and has a movement (legacy: "privremeno zbog Triangl papira"),
//  4. takes the article data from RSIF (rsif) and the konto from RSTA (drsta), and sums the
//     monthly columns MUL/MIZ/MDUG/MPOT from the "od meseca" to the "zaključno sa mesecom" month.
func (s *RobnoStanjaResource) GetStanjaViseArtikalaGrupa(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, page, size)
	// Like the "po šifri" report, the printed group report uses its own columns and is never paged.
	stampa := printType == common.TipStampePrint
	finansijski := stampa && params.FinansijskiIznos
	switch {
	case !stampa:
		tbl.Headers = s.viseArtikalaTableFields
	case finansijski:
		tbl.Headers = s.viseArtikalaStampaFinTableFields
	default:
		tbl.Headers = s.viseArtikalaStampaTableFields
	}

	nabavnaCena := s.magacinJeNabavnaCena(ctx, userSession, params.Magacin)
	odMesec, doMesec := mesecRange(params.OdMeseca, params.DoMeseca)

	hasGod, hasKar := s.drstaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(robnoStanjeArtikliBaseQuery, true)
	qb.AddJoin(" inner join rsif on rsif.rsifid = drsta.rsifid")
	if hasGod {
		qb.AddEqual("drsta.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("drsta.kar", userSession.SelectedKar)
	}
	qb.AddEqual("drsta.magaciniid", params.Magacin)
	// Parameters of the legacy query: the group range and the article-code range.
	qb.AddCondition("rsif.gru", params.OdGrupe, ">=")
	qb.AddCondition("rsif.gru", params.DoGrupe, "<=")
	qb.AddCondition("rsif.sifra", params.OdSifre, ">=")
	qb.AddCondition("rsif.sifra", params.DoSifre, "<=")
	// The legacy query returned only articles with a stock record that has a movement.
	qb.AddCustomCondition(legacyArticleMovementCondition)
	qb.AddOrderBy("rsif.gru, rsif.sifra")
	if !total && !stampa && size > 0 {
		qb.SetLimit(size)
		qb.SetOffset((page - 1) * size)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if total && !stampa {
		common.SetTableTotalRecords(tbl, len(*entities), size)
		return nil
	}

	// The group form has no "prosečna cena" options, so cenaArtikla returns CENA/PROSNC here.
	tbl.Rows = []domain.TableRow{}
	var ukupnoUlaz, ukupnoIzlaz, ukupnoDug, ukupnoPot float64
	rbr := 0
	lastGrupa := -1
	for _, entity := range *entities {
		ulaz, izlaz, dug, pot := zbirMeseca(entity, odMesec, doMesec)
		cena := cenaArtikla(entity, nabavnaCena, ulaz, izlaz, dug, pot, params)
		ukupnoUlaz += ulaz
		ukupnoIzlaz += izlaz
		ukupnoDug += dug
		ukupnoPot += pot
		if stampa {
			// The printed report is broken per group; "Nova strana po grupi" starts every group
			// on a new page (except the first one, which would produce an empty first page).
			if entity.Gru != lastGrupa {
				classRow := "group-header"
				if params.NovaStranaPoGrupi && lastGrupa != -1 {
					classRow = "group-header-new-page"
				}
				tbl.Rows = append(tbl.Rows, robnoStanjeGrupaHeaderRow(entity.Gru, classRow))
				lastGrupa = entity.Gru
				rbr = 0
			}
			rbr++
			tbl.Rows = append(tbl.Rows, robnoStanjeArtikalStampaRow(rbr, entity, cena, ulaz, izlaz, dug, pot, finansijski))
			continue
		}
		tbl.Rows = append(tbl.Rows, robnoStanjeArtikalRow(entity, cena, ulaz, izlaz))
	}

	if stampa {
		if lastGrupa != -1 {
			tbl.Rows = append(tbl.Rows, robnoStanjeArtikalStampaTotalRow(ukupnoUlaz, ukupnoIzlaz, ukupnoDug, ukupnoPot, finansijski))
		}
		return nil
	}

	setRobnoStanjeArtikliTotals(tbl, ukupnoUlaz, ukupnoIzlaz)
	return nil
}

func (s *RobnoStanjaResource) GetSvodjenjeZaliha(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetSvodjenjeZaliha

	return nil
}

// GetMagacinComboValues returns the magacini of the current period (CommonService).
func (s *RobnoStanjaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinComboValues(ctx)
}

// GetTipDokComboValues returns the vrste naloga of the current period keyed by idtipdok
// (CommonService).
func (s *RobnoStanjaResource) GetTipDokComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetTipdokIDComboValues(ctx)
}

// GetOjComboValues returns the organizacione jedinice of the current period (CommonService).
func (s *RobnoStanjaResource) GetOjComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetOrgJedComboValues(ctx, commonsvc.WithEmptyOption())
}

// GetMestoTroskaComboValues returns the mesta troška of one organizaciona jedinica (CommonService).
func (s *RobnoStanjaResource) GetMestoTroskaComboValues(ctx context.Context, idOrgjed int) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMestoTroskaComboValues(ctx, int64(idOrgjed), commonsvc.WithEmptyOption())
}

func (s *RobnoStanjaResource) GetPojedinacnogArtiklaTableFields() []domain.Fields {
	return s.pojedinacniArtikalTableFields
}
func (s *RobnoStanjaResource) GetViseArtikalaTableFields() []domain.Fields {
	return s.viseArtikalaTableFields
}

func (s *RobnoStanjaResource) GetSubsintetiskogKontaTableFields() []domain.Fields {
	return s.subsintetickogKontaTableFields
}
func (s *RobnoStanjaResource) GetSvodjenjeZalihaTableFields() []domain.Fields {
	return s.svodjenjeZalihaTableFields
}

// GetFvrData retrieves company (fvr) data for the current session.
func (s *RobnoStanjaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, &s.fvrRepo)
}

func (s *RobnoStanjaResource) SetDefaultTableData(tbl *domain.TableData) {
	tbl.SearchEnabled = false
	tbl.ShowPagination = false

	// Opening balance row
	fields := []string{"Početno stanje", "", "", "", ""}
	tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
	tbl.Rows = append(tbl.Rows, tblRow)

	months := common.GetMontshName()
	// Add month rows
	for i := 0; i <= 11; i++ {
		// Create a NEW slice for each row instead of reusing the same slice
		monthFields := []string{months[i], "", "", "", ""}
		tblRow := domain.TableRow{Fields: monthFields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
}

func (s *RobnoStanjaResource) setTableFileds() {
	s.pojedinacniArtikalTableFields = []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "12", TextAlign: "left"},
		{Name: "ulaz", Label: "Ulaz", Width: "14", TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "14", TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "14", TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "14", TextAlign: "right"},
	}
	// Stanje više artikala po sifri: jedan red po artiklu (konto, identifikacija artikla,
	// cena i zbir ulaza/izlaza/stanja za izabrani period).
	s.viseArtikalaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10", Field: "drsta.konto", SkipInSearch: false, Sortable: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "10", Field: "rsif.sifra", SkipInSearch: false, Sortable: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "38", Field: "rsif.naziv", SkipInSearch: false},
		{Name: "jm", Label: "JM", Width: "6", Field: "rsif.jm", SkipInSearch: false, TextAlign: "center", Sortable: true},
		{Name: "cena", Label: "Cena", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "ulaz", Label: "Ulaz", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "izlaz", Label: "Izlaz", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "stanje", Label: "Stanje", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true, Sortable: true},
	}
	// Stampanje "Prikaz stanja više artikala": grid kolone + redni broj. Kada je uključen
	// "Finansijski iznos" dodaju se i kolone duguje/potražuje/saldo (viseArtikalaStampaFinTableFields).
	s.viseArtikalaStampaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. Br.", Width: "5", SkipInSearch: true, TextAlign: "center"},
		{Name: "konto", Label: "Konto", Width: "8", SkipInSearch: true, TextAlign: "left"},
		{Name: "sifra", Label: "Šifra", Width: "7", SkipInSearch: true, TextAlign: "right"},
		{Name: "naziv", Label: "Naziv", Width: "30", SkipInSearch: true, TextAlign: "left"},
		{Name: "jm", Label: "JM", Width: "5", SkipInSearch: true, TextAlign: "center"},
		{Name: "cena", Label: "Cena", Width: "9", SkipInSearch: true, TextAlign: "right"},
		{Name: "ulaz", Label: "Ulaz", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "stanje", Label: "Stanje", Width: "10", SkipInSearch: true, TextAlign: "right"},
	}
	s.viseArtikalaStampaFinTableFields = append(slices.Clone(s.viseArtikalaStampaTableFields),
		domain.Fields{Name: "duguje", Label: "Duguje", Width: "11", SkipInSearch: true, TextAlign: "right"},
		domain.Fields{Name: "potrazuje", Label: "Potražuje", Width: "11", SkipInSearch: true, TextAlign: "right"},
		domain.Fields{Name: "saldo", Label: "Saldo", Width: "11", SkipInSearch: true, TextAlign: "right"},
	)

	s.subsintetickogKontaTableFields = []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "10"},
		{Name: "duguje", Label: "Duguje", Width: "12", TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12", TextAlign: "right"},
		{Name: "saldo", Label: "Saldo u mesecu", Width: "12", TextAlign: "right"},
		{Name: "saldokumul", Label: "Saldo na kraju meseca", Width: "15", TextAlign: "right"},
	}
	s.svodjenjeZalihaTableFields = []domain.Fields{}
}

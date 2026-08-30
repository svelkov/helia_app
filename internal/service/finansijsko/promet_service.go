package finansijsko

import (
	"context"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// FproViewData encapsulates all data needed for the Nalog display page.
type PrometViewData struct {
	FproEntities []domain.Fpro
	prometRepo   repository.BaseRepository[domain.PrometDto]
	TableData    domain.TableData
}

// NalogService defines the interface for operations related to Fpro (Nalogs).
type PrometService interface {
	GetFieldCache() map[string]reflect.StructField
	GetPrometAnalitickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, isMI bool, params domain.PrometParam) error
	GetPrometSubsintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometKarticaSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometKontaAnaliticki(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometSubsintetikaVrd(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam, tipStampe string) error
	GetPrometDeviznaAnalitickaKonta(ctx context.Context, tbl, tblDevizni *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometTotals(ctx context.Context, params domain.PrometParam) (domain.PrometResponse, error)
	GetAnkontaTableFields() []domain.Fields
	GetAnKontaMiTableFields() []domain.Fields
	GetAnDeviznaKontaTableFields() []domain.Fields
	GetAnDeviznaKontaRekapTableFields() []domain.Fields
	GetSubsintetickihKontaTableFields() []domain.Fields
	GetSintetickihKontaTableFields() []domain.Fields
	GetKarticaSintetikaTableFields() []domain.Fields
	GetSubsintetikaVrdTableFields() []domain.Fields
	GetKontaAnalitickiTableFields() []domain.Fields
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetPrometAnalitickaKarticaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometAnalitickaKarticaStampaTableFields() []domain.Fields
	GetPrometSubsintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometSubsintetickihKontaStampaTableFields() []domain.Fields
	GetPrometSintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometSintetickihKontaStampaTableFields() []domain.Fields
	GetPrometKarticaSintKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometKarticaSintKontaStampaTableFields() []domain.Fields
	GetPrometSubsintetikaVrdStampaTableFields() []domain.Fields
	GetPrometAnalitickaKarticaPoMIStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometAnalitickaKarticaPoMIStampaTableFields() []domain.Fields
}

// PrometResource implements the PrometService interface.
type PrometResource struct {
	service                                      *service.BaseService[domain.PrometDto]
	prometRepo                                   *repository.BaseRepository[domain.PrometDto]
	fkplRepo                                     *repository.BaseRepository[domain.Fkpl]
	fvrRepo                                      *repository.BaseRepository[domain.Fvr]
	prometAnKontaTableFields                     []domain.Fields
	prometAnKontaMiTableFields                   []domain.Fields
	prometAnDeviznaKontaTableFields              []domain.Fields
	prometAnDeviznaKontaRekapTableFields         []domain.Fields
	prometSubsintetickihKontaTablefields         []domain.Fields
	prometSintetickihKontaTableFields            []domain.Fields
	prometKarticaSintetickihKontaTableFields     []domain.Fields
	prometSubsintetikaVrdTableFields             []domain.Fields
	prometKontaAnalitickiTableFields             []domain.Fields
	prometAnalitickaKarticaStampaTableFields     []domain.Fields
	prometSubsintetickihKontaStampaTableFields   []domain.Fields
	prometSintetickihKontaStampaTableFields      []domain.Fields
	prometKarticaSintKontaStampaTableFields      []domain.Fields
	prometSubsintetikaVrdStampaTableFields       []domain.Fields
	prometAnalitickaKarticaPoMIStampaTableFields []domain.Fields
}

func NewPrometService(service *service.BaseService[domain.PrometDto], prometRepo *repository.BaseRepository[domain.PrometDto], fkplRepo *repository.BaseRepository[domain.Fkpl], fvrRepo *repository.BaseRepository[domain.Fvr]) *PrometResource {
	rs := &PrometResource{
		service:    service,
		prometRepo: prometRepo,
		fkplRepo:   fkplRepo,
		fvrRepo:    fvrRepo,
	}
	rs.setServiceFieldValues()
	return rs
}
func (s *PrometResource) GetPrometTotals(ctx context.Context, params domain.PrometParam) (domain.PrometResponse, error) {
	var response domain.PrometResponse
	reportTip := params.ReportTip
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		if params.Konto == "" || params.Sifra == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "prometankontami": // Analiticka Konta po MI
		if params.Konto == "" || params.Sifra == "" || params.OdDatuma == "" || params.DoDatuma == "" || params.OdMI == "" || params.DoMI == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "deviznihanalitickihkonta": // Devizna Analiticka Konta
		if params.Konto == "" || params.Sifra == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "subsintetickakonta": // Subsinteticka Konta
		if params.OdKonta == "" || params.DoKonta == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		if params.Konto == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "kontaanaliticki": // Konta Analiticki
		if params.Konto == "" || params.OdSifre == "" || params.DoSifre == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	default:
		return response, fmt.Errorf("invalid report tip")
	}

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return response, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	// Get "promet do" totals (up to start date)
	qbDo := common.NewQueryBuilder(`
		select 
			coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
		from fpro`, true)

	if hasGod {
		qbDo.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qbDo.AddEqual("kar", userSession.SelectedKar)
	}
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		qbDo.AddEqual("konto", params.Konto)
		qbDo.AddEqual("sifra", params.Sifra)
		qbDo.AddEqual("vkonta", 1)
		qbDo.AddCondition("danal", params.OdDatuma, "<")
	case "prometankontami": // Analiticka Konta po MI
		qbDo.AddEqual("konto", params.Konto)
		qbDo.AddEqual("sifra", params.Sifra)
		qbDo.AddEqual("vkonta", 1)
		qbDo.AddCondition("danal", params.OdDatuma, "<")
		qbDo.AddCondition("mi", params.OdMI, ">=")
		qbDo.AddCondition("mi", params.DoMI, "<=")
	case "deviznihanalitickihkonta": // Devizna Analiticka Konta
		qbDo.AddEqual("konto", params.Konto)
		qbDo.AddEqual("sifra", params.Sifra)
		qbDo.AddEqual("vkonta", 1)
		qbDo.AddCondition("danal", params.OdDatuma, "<")
	case "subsintetickakonta": // Subsinteticka Konta
		qbDo.AddCondition("konto", params.OdKonta, ">=")
		qbDo.AddCondition("konto", params.DoKonta, "<=")
		qbDo.AddCondition("danal", params.OdDatuma, "<")

	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		qbDo.AddLikeBegin("konto", params.Konto)
		qbDo.AddCondition("danal", params.OdDatuma, "<")
	case "kontaanaliticki": // Konta Analiticki
		qbDo.AddEqual("konto", params.Konto)
		qbDo.AddCondition("sifra::numeric", params.OdSifre, ">=")
		qbDo.AddCondition("sifra::numeric", params.DoSifre, "<=")
		qbDo.AddCondition("danal", params.OdDatuma, "<")
		qbDo.AddEqual("vkonta", 1)
	default:
	}

	prometDoQuery, prometDoArgs := qbDo.Build()
	prometDoResults, err := s.prometRepo.GetAllCustom(ctx, prometDoQuery, "", prometDoArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet do totals: %v", err)
	}

	var prometDoDuguje, prometDoPotrazuje float64
	if len(*prometDoResults) > 0 {
		prometDoDuguje = (*prometDoResults)[0].Duguje
		prometDoPotrazuje = (*prometDoResults)[0].Potrazuje
	}

	response.Totals = domain.TotalValues{
		DugDo:   common.FormatFloatNumber64WithSystemLocale(prometDoDuguje, 2),
		PotDo:   common.FormatFloatNumber64WithSystemLocale(prometDoPotrazuje, 2),
		SaldoDo: common.FormatFloatNumber64WithSystemLocale(prometDoDuguje-prometDoPotrazuje, 2),
	}

	// Get "promet za period" totals (for the specified period)
	qbPeriod := common.NewQueryBuilder(`
		select 
			coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
		from fpro`, true)

	if hasGod {
		qbPeriod.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qbPeriod.AddEqual("kar", userSession.SelectedKar)
	}
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		qbPeriod.AddEqual("konto", params.Konto)
		qbPeriod.AddEqual("sifra", params.Sifra)
		qbPeriod.AddEqual("vkonta", 1)
		qbPeriod.AddCondition("danal", params.OdDatuma, ">=")
		qbPeriod.AddCondition("danal", params.DoDatuma, "<=")
	case "prometankontami": // Analiticka Konta po MI
		qbPeriod.AddEqual("konto", params.Konto)
		qbPeriod.AddEqual("sifra", params.Sifra)
		qbPeriod.AddEqual("vkonta", 1)
		qbPeriod.AddCondition("danal", params.OdDatuma, ">=")
		qbPeriod.AddCondition("danal", params.DoDatuma, "<=")
		qbPeriod.AddCondition("mi", params.OdMI, ">=")
		qbPeriod.AddCondition("mi", params.DoMI, "<=")
	case "deviznihanalitickihkonta": // Devizna Analiticka Konta
		qbPeriod.AddEqual("konto", params.Konto)
		qbPeriod.AddEqual("sifra", params.Sifra)
		qbPeriod.AddEqual("vkonta", 1)
		qbPeriod.AddCondition("danal", params.OdDatuma, ">=")
		qbPeriod.AddCondition("danal", params.DoDatuma, "<=")
	case "subsintetickakonta": // Subsinteticka Konta
		qbPeriod.AddCondition("konto", params.OdKonta, ">=")
		qbPeriod.AddCondition("konto", params.DoKonta, "<=")
		qbPeriod.AddCondition("danal", params.OdDatuma, ">=")
		qbPeriod.AddCondition("danal", params.DoDatuma, "<=")
	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		qbPeriod.AddLikeBegin("konto", params.Konto)
		qbPeriod.AddCondition("danal", params.OdDatuma, ">=")
		qbPeriod.AddCondition("danal", params.DoDatuma, "<=")
	case "kontaanaliticki": // Konta Analiticki
		qbPeriod.AddEqual("konto", params.Konto)
		qbPeriod.AddCondition("sifra::numeric", params.OdSifre, ">=")
		qbPeriod.AddCondition("sifra::numeric", params.DoSifre, "<=")
		qbPeriod.AddCondition("danal", params.OdDatuma, ">=")
		qbPeriod.AddCondition("danal", params.DoDatuma, "<=")
		qbPeriod.AddEqual("vkonta", 1)
	default:
	}

	prometPeriodQuery, prometPeriodArgs := qbPeriod.Build()
	prometPeriodResults, err := s.prometRepo.GetAllCustom(ctx, prometPeriodQuery, "", prometPeriodArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet period totals: %v", err)
	}

	var prometPeriodDuguje, prometPeriodPotrazuje float64
	if len(*prometPeriodResults) > 0 {
		prometPeriodDuguje = (*prometPeriodResults)[0].Duguje
		prometPeriodPotrazuje = (*prometPeriodResults)[0].Potrazuje
	}

	response.Totals.DugPer = common.FormatFloatNumber64WithSystemLocale(prometPeriodDuguje, 2)
	response.Totals.PotPer = common.FormatFloatNumber64WithSystemLocale(prometPeriodPotrazuje, 2)
	response.Totals.SaldoPer = common.FormatFloatNumber64WithSystemLocale(prometPeriodDuguje-prometPeriodPotrazuje, 2)
	response.Totals.DugTot = common.FormatFloatNumber64WithSystemLocale(response.Totals.DugDo+prometPeriodDuguje, 2)
	response.Totals.PotTot = common.FormatFloatNumber64WithSystemLocale(response.Totals.PotDo+prometPeriodPotrazuje, 2)
	response.Totals.SaldoTot = common.FormatFloatNumber64WithSystemLocale(response.Totals.SaldoDo+(prometPeriodDuguje-prometPeriodPotrazuje), 2)
	return response, err
}

func (s *PrometResource) CheckPrometParameters(ctx context.Context, requiredFields []string, params domain.PrometParam) (fieldsError []domain.FieldError) {
	// Build query dynamically
	qb := common.NewQueryBuilder(`SELECT f.konto, f.sifra FROM baza.fkpl as f`, true)

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "User session not found"}}
	}

	// Add system conditions
	hasGod, hasKAr := s.service.Repo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKAr {
		qb.AddEqual("f.kar", session.SelectedKar)
	}

	// Add user conditions
	qb.AddEqual("f.konto", params.Konto)
	qb.AddEqual("f.sifra", params.Sifra)
	qb.AddEqual("f.vkonta", params.Vkonta)

	sqlQuery, args := qb.Build()

	entities, err := s.service.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetData}}
	}
	if len(*entities) == 0 {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetKontoSifra}}
	}

	return nil
}

func (s *PrometResource) GetPrometAnalitickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, isMI bool, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT danal, tipdok, concat(tipdok,'-',nalog) as nalog, idfpro, kat, iznos, kolic, 
		       	vrd, dokum, dadok, rok, tra, coalesce(ojozn, '') as oj, coalesce(opis, '') as opis, coalesce(sifval, 0) as sifval, kurs, 
		       	deviznos, cena, konto, idfnal, idfkpl, 
				coalesce(dokumv, '') as dokumv, dadokv, coalesce(travez, 0) as travez,
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				CASE WHEN kat = 1 OR kat = 2 THEN kolic ELSE 0 END as kolduguje,
				CASE WHEN kat = 3 OR kat = 4 THEN kolic ELSE 0 END as kolpotrazuje	
		FROM fpro`, true)
	qbTotal := common.NewQueryBuilder(`SELECT 
			   	coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			   	coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje,
				coalesce(sum(case when kat = 1 or kat = 2 then kolic else 0 end), 0) as kolduguje,
				coalesce(sum(case when kat = 3 or kat = 4 then kolic else 0 end), 0) as kolpotrazuje,	
				coalesce(sum(deviznos), 0) as deviznos
		FROM fpro`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
		qbTotal.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
		qbTotal.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("konto", params.Konto)
	qb.AddEqual("sifra", params.Sifra)
	qb.AddEqual("vkonta", 1)
	qb.AddCondition("danal", params.OdDatuma, ">=")
	qb.AddCondition("danal", params.DoDatuma, "<=")

	qbTotal.AddEqual("konto", params.Konto)
	qbTotal.AddEqual("sifra", params.Sifra)
	qbTotal.AddEqual("vkonta", 1)
	qbTotal.AddCondition("danal", params.OdDatuma, ">=")
	qbTotal.AddCondition("danal", params.DoDatuma, "<=")
	if isMI {
		qb.AddCondition("mi", params.OdMI, ">=")
		qb.AddCondition("mi", params.DoMI, "<=")
		qbTotal.AddCondition("mi", params.OdMI, ">=")
		qbTotal.AddCondition("mi", params.DoMI, "<=")
	}
	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		//qbTotal.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		if isMI {
			qb.AddSearchConditions(s.GetAnKontaMiTableFields(), params.SearchText)
			//qbTotal.AddSearchConditions(s.GetAnKontaMiTableFields(), params.SearchText)
		} else {
			qb.AddSearchConditions(s.GetAnkontaTableFields(), params.SearchText)
			//qbTotal.AddSearchConditions(s.GetAnkontaTableFields(), params.SearchText)
		}
	}

	qb.AddOrderBy("god, kar, danal, tipdok, nalog, idfpro")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot, totKolDug, totKolPot, totDevIznos float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{}

			// Add MI as first field only if isMI is true
			if isMI {
				fields = append(fields, entity.MI)
			}

			// Add common fields
			fields = append(fields,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				entity.Tra,
				entity.Ojozn,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Kurs, 2),
				common.FormatNumberWithSystemLocale(entity.Deviznos, 2),
				common.FormatNumberWithSystemLocale(entity.Kolduguje, 2),
				common.FormatNumberWithSystemLocale(entity.Kolpotrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Cena, 2),
				common.FormatNumberWithSystemLocale(entity.Stanje, 2),
				entity.Dokumv,
				entity.Dadokv.Time.Format(common.DateLayout),
				entity.Travez,
			)
			if !isMI {
				fields = append(fields, entity.MI)
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
			totKolDug += entity.Kolduguje
			totKolPot += entity.Kolpotrazuje
			totDevIznos += entity.Deviznos
		}
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		case "kolduguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolDug, 2)
		case "kolpotrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolPot, 2)
		case "deviznos":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDevIznos, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometDeviznaAnalitickaKonta(ctx context.Context, tbl, tblDevizni *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT danal, tipdok, concat(tipdok,'-',nalog) as nalog, idfpro, kat, iznos, kolic, 
		       	vrd, dokum, dadok, rok, tra, coalesce(ojozn, '') as oj, coalesce(opis, '') as opis, coalesce(sifval, 0) as sifval, kurs, 
		       	deviznos, cena, konto, idfnal, idfkpl, 
				coalesce(dokumv, '') as dokumv, dadokv, coalesce(travez, 0) as travez,
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				CASE WHEN kat = 1 OR kat = 2 THEN kolic ELSE 0 END as kolduguje,
				CASE WHEN kat = 3 OR kat = 4 THEN kolic ELSE 0 END as kolpotrazuje	
		FROM fpro`, true)
	qbTotal := common.NewQueryBuilder(`SELECT 
			   	coalesce(sifval, 0) as sifval,
			   	coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			   	coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje,
				coalesce(sum(case when kat = 1 or kat = 2 then iznos * kurs else 0 end), 0) as devduguje,
			   	coalesce(sum(case when kat = 3 or kat = 4 then iznos * kurs else 0 end), 0) as devpotrazuje,
				coalesce(sum(deviznos), 0) as deviznos
		FROM fpro`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
		qbTotal.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
		qbTotal.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("konto", params.Konto)
	qb.AddEqual("sifra", params.Sifra)
	qb.AddEqual("vkonta", 1)
	qb.AddCondition("danal", params.OdDatuma, ">=")
	qb.AddCondition("danal", params.DoDatuma, "<=")

	qbTotal.AddEqual("konto", params.Konto)
	qbTotal.AddEqual("sifra", params.Sifra)
	qbTotal.AddEqual("vkonta", 1)
	qbTotal.AddCondition("danal", params.OdDatuma, ">=")
	qbTotal.AddCondition("danal", params.DoDatuma, "<=")
	qbTotal.AddGroupBy("sifval")

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetAnkontaTableFields(), params.SearchText)
		//qbTotal.AddSearchConditions(s.GetAnkontaTableFields(), params.SearchText)
	}

	qb.AddOrderBy("god, kar, danal, tipdok, nalog")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot, totKolDug, totKolPot, totDevIznos float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Opis,
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Kurs, 4),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				entity.Tra,
				entity.Ojozn,
				entity.Konto,
			)
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
			totKolDug += entity.Kolduguje
			totKolPot += entity.Kolpotrazuje
			totDevIznos += entity.Deviznos
		}
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		case "kolduguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolDug, 2)
		case "kolpotrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolPot, 2)
		case "deviznos":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDevIznos, 2)
		}
	}
	sqlQueryTotal, argsTotal := qbTotal.Build()
	entitiesDevizni, err := s.prometRepo.GetAllCustom(ctx, sqlQueryTotal, "", argsTotal, "", "")
	if err != nil {
		return err
	}
	if len(*entitiesDevizni) > 0 {
		for _, entity := range *entitiesDevizni {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Devduguje, 2),
				common.FormatNumberWithSystemLocale(entity.Devpotrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Devduguje-entity.Devpotrazuje, 2),
			)
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tblDevizni.Rows = append(tblDevizni.Rows, tblRow)
		}
	}

	return nil
}
func (s *PrometResource) GetPrometAnalitickaKarticaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`SELECT
		f.tipdok, f.nalog, f.danal, f.vrd,
		COALESCE(f.opis, '') as opis,
		COALESCE(f.ojozn, '') as oj,
		f.dokum,
		f.tra,
		f.dadok,
		f.rok,
		f.brst,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(f.sifra, '') as sifra,
		COALESCE(fk.naziv, '') as naziv
	FROM fpro f`, true)
	qb.AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl")

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddEqual("f.vkonta", 1)
	qb.AddCondition("f.konto", params.OdKonta, ">=")
	qb.AddCondition("f.konto", params.DoKonta, "<=")
	qb.AddCondition("f.sifra::numeric", params.OdSifre, ">=")
	qb.AddCondition("f.sifra::numeric", params.DoSifre, "<=")
	qb.AddCondition("f.danal", params.OdDatuma, ">=")
	qb.AddCondition("f.danal", params.DoDatuma, "<=")
	qb.AddOrderBy("f.konto, f.sifra::numeric, f.danal, f.nalog::numeric, f.brst")

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	lastGroupKey := ""
	var groupDug, groupPot, runningSaldo float64
	var lastKonto, lastSifra, lastNaziv string

	for _, entity := range *entities {
		groupKey := entity.Konto + "|" + entity.Sifra

		if groupKey != lastGroupKey {
			if lastGroupKey != "" {
				pSaldo := groupDug - groupPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "konto-total",
					Fields: []string{
						translator.Label("Ukupno za") + ": " + lastKonto + " | " + lastSifra + " | " + lastNaziv,
						common.FormatNumberWithSystemLocale(groupDug, 2),
						common.FormatNumberWithSystemLocale(groupPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			groupDug, groupPot = 0, 0
			runningSaldo = 0
			lastGroupKey = groupKey
			lastKonto = entity.Konto
			lastSifra = entity.Sifra
			lastNaziv = entity.Naziv
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, entity.Sifra, entity.Naziv},
			})
		}

		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Opis,
				entity.OJ,
				entity.Dokum,
				entity.Tra,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	if lastGroupKey != "" {
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto + " | " + lastSifra + " | " + lastNaziv,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	nHeaders := len(tbl.Headers)
	if nHeaders >= 3 {
		grandSaldo := grandDug - grandPot
		tbl.Totals[nHeaders-3] = common.FormatNumberWithSystemLocale(grandDug, 2)
		tbl.Totals[nHeaders-2] = common.FormatNumberWithSystemLocale(grandPot, 2)
		tbl.Totals[nHeaders-1] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
	}

	return nil
}

// GetPrometSubsintetickihKonta retrieves data for subsinteticka konta based on the provided parameters and populates the table data structure.
func (s *PrometResource) GetPrometSubsintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT  konto, tipdok, nalog, danal, coalesce(opis, '') as opis, 
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				idfpro
		FROM fpro`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)

	}
	qb.AddCondition("konto::numeric", params.OdKonta, ">=")
	qb.AddCondition("konto::numeric", params.DoKonta, "<=")
	qb.AddCondition("danal", params.OdDatuma, ">=")
	qb.AddCondition("danal", params.DoDatuma, "<=")

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetSubsintetickihKontaTableFields(), params.SearchText)
	}
	qb.AddOrderBy("konto, danal, tipdok, nalog")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Konto,
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT LEFT(konto, 3) as konto, tipdok, nalog, 
			   	MAX(danal) as danal, 
			   	MAX(coalesce(opis, '')) as opis,
			   	SUM(CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END) as potrazuje
		FROM fpro`, true)

	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddLikeBegin("konto", params.OdKonta)
	qb.AddCondition("danal", params.OdDatuma, ">=")
	qb.AddCondition("danal", params.DoDatuma, "<=")

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetSintetickihKontaTableFields(), params.SearchText)
	}
	qb.AddGroupBy("LEFT(konto, 3), tipdok, nalog")
	qb.AddOrderBy("tipdok, nalog")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}
	return nil
}

func (s *PrometResource) GetPrometKarticaSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	sqlText := ""
	groupByClause := ""
	if params.Analitika == "true" {
		sqlText = `SELECT fpro.konto, fpro.sifra,
			   	MAX(COALESCE(fkpl.naziv, '')) as opis,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro
				LEFT JOIN fkpl ON fkpl.konto = fpro.konto AND fkpl.sifra = fpro.sifra AND fkpl.god = fpro.god AND fkpl.kar = fpro.kar`
		groupByClause = "fpro.konto, fpro.sifra"
	} else {
		sqlText = `SELECT fpro.konto, 
			   	MAX(COALESCE(fkpl.naziv, '')) as opis,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro
				LEFT JOIN fkpl ON fkpl.konto = fpro.konto AND fkpl.god = fpro.god AND fkpl.kar = fpro.kar AND fkpl.vkonta = 2`
		groupByClause = "fpro.konto"
	}

	qb := common.NewQueryBuilder(sqlText, true)
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddLikeBegin("fpro.konto", params.OdKonta)
	qb.AddCondition("danal", params.OdDatuma, ">=")
	qb.AddCondition("danal", params.DoDatuma, "<=")

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))

		qb.AddSearchConditions(s.GetKarticaSintetikaTableFields(), params.SearchText)
	}
	qb.AddGroupBy(groupByClause)
	qb.AddOrderBy("fpro.konto::numeric")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Konto,
				"",
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}
			if params.Analitika == "true" {
				fields[1] = entity.Sifra
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}
	return nil
}

func (s *PrometResource) GetPrometSubsintetikaVrd(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam, tipStampe string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	// Look up konto name from fkpl
	if tipStampe == common.TipStampePrint {
		qbFkpl := common.NewQueryBuilder("SELECT konto, naziv FROM fkpl", true)
		if hasGod {
			if hasGod {
				qbFkpl.AddEqual("god", userSession.SelectedGod)
			}
			if hasKar {
				qbFkpl.AddEqual("kar", userSession.SelectedKar)
			}
		}
		qbFkpl.AddEqual("konto", params.OdKonta)
		qbFkpl.AddEqual("vkonta", "2")
		sqlFkpl, argsFkpl := qbFkpl.Build()

		fkplEntities, _ := s.fkplRepo.GetAllCustom(ctx, sqlFkpl, "", argsFkpl, "", "")
		if fkplEntities != nil && len(*fkplEntities) > 0 {
			tbl.ContentTitle = (*fkplEntities)[0].Naziv
		}
	}
	qb := common.NewQueryBuilder(`SELECT 
				CASE fpro.vkonta
				WHEN '1' THEN 'Analitika'
				WHEN '2' THEN 'Subsintetika'
				ELSE ''
				END as vkonta,
				fpro.vrd,
				SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro`, true)

	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddLikeBegin("fpro.konto", params.Konto)
	qb.AddCondition("danal", params.OdDatuma, ">=")
	qb.AddCondition("danal", params.DoDatuma, "<=")

	// Add search conditions if search text is provided

	if tipStampe == common.TipStampePreview && params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetSubsintetikaVrdTableFields(), params.SearchText)
	}
	qb.AddGroupBy(`fpro.vrd, fpro.vkonta`)
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	// fmt.Println(sqlQuery, args)
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if tipStampe == common.TipStampePreview && getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Vkonta,
				entity.Vrd,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}

			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometKontaAnaliticki(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT f.danal, f.tipdok, concat(f.tipdok,'-',nalog) as nalog, f.idfpro, f.kat, f.iznos, f.kolic, 
		       	f.vrd, coalesce(f.dokum, '') as dokum, f.dadok, 
				coalesce(f.ojozn,'') as oj, coalesce(f.opis,'') as opis, 
				f.rok, f.tra, f.sifval, f.kurs,
		       	f.deviznos, f.cena, f.konto, f.sifra, fk.naziv, f.idfkpl, 
				coalesce(f.dokumv, '') as dokumv, f.dadokv, f.travez, coalesce(f.rdokid, 0) as rdokid,
			   	CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
			   	CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
				CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.kolic ELSE 0 END as kolduguje,
				CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.kolic ELSE 0 END as kolpotrazuje	
		FROM fpro as f`, true)
	qb.AddJoin(` LEFT JOIN fkpl fk on fk.idfkpl = f.idfkpl`)

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddEqual("f.konto", params.Konto)
	qb.AddCondition("f.sifra::numeric", params.OdSifre, ">=")
	qb.AddCondition("f.sifra::numeric", params.DoSifre, "<=")
	qb.AddEqual("f.vkonta", 1)
	qb.AddCondition("f.danal", params.OdDatuma, ">=")
	qb.AddCondition("f.danal", params.DoDatuma, "<=")

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetKontaAnalitickiTableFields(), params.SearchText)

	}

	qb.AddOrderBy("f.sifra::numeric, f.danal, f.tipdok, f.nalog")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	var totDug, totPot, totKolDug, totKolPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Sifra,
				entity.Naziv,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				entity.Tra,
				entity.Ojozn,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Kurs, 2),
				common.FormatNumberWithSystemLocale(entity.Deviznos, 2),
				common.FormatNumberWithSystemLocale(entity.Kolduguje, 2),
				common.FormatNumberWithSystemLocale(entity.Kolpotrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Cena, 2),
				common.FormatNumberWithSystemLocale(entity.Stanje, 2),
				entity.Dokumv,
				entity.Dadokv.Time.Format(common.DateLayout),
				entity.Travez,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
			totKolDug += entity.Kolduguje
			totKolPot += entity.Kolpotrazuje

		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		case "kolduguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolDug, 2)
		case "kolpotrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolPot, 2)
		}
	}
	return nil
}
func (s *PrometResource) GetPrometAnalitickaKarticaPoMIStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`SELECT
		f.tipdok, f.nalog, f.danal, f.vrd,
		COALESCE(f.opis, '') as opis,
		COALESCE(f.ojozn, '') as oj,
		f.dokum,
		f.tra,
		f.dadok,
		f.rok,
		f.brst,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(f.sifra, '') as sifra,
		COALESCE(fk.naziv, '') as naziv,
		COALESCE(f.mi::text, '0') as mi
	FROM fpro f`, true)
	qb.AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl")
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddEqual("f.vkonta", 1)
	qb.AddCondition("f.konto", params.OdKonta, ">=")
	qb.AddCondition("f.konto", params.DoKonta, "<=")
	qb.AddCondition("f.sifra::numeric", params.OdSifre, ">=")
	qb.AddCondition("f.sifra::numeric", params.DoSifre, "<=")
	qb.AddCondition("f.danal", params.OdDatuma, ">=")
	qb.AddCondition("f.danal", params.DoDatuma, "<=")
	if params.OdMI != "" {
		qb.AddCondition("f.mi::numeric", params.OdMI, ">=")
	}
	if params.DoMI != "" {
		qb.AddCondition("f.mi::numeric", params.DoMI, "<=")
	}
	qb.AddOrderBy("f.konto, f.sifra::numeric, f.mi::numeric, f.danal, f.nalog::numeric, f.brst")

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))
	if entities == nil || len(*entities) == 0 {
		return nil
	}

	monthNames := []string{"", "Januar", "Februar", "Mart", "April", "Maj", "Jun", "Jul", "Avgust", "Septembar", "Oktobar", "Novembar", "Decembar"}

	var grandDug, grandPot float64
	var groupDug, groupPot, runningSaldo float64
	var miDug, miPot float64
	var monthDug, monthPot float64
	lastGroupKey := ""
	lastMIKey := ""
	lastMonthKey := ""
	lastMonthLabel := ""
	var lastKonto, lastSifra, lastNaziv, lastMI string

	appendMonthTotal := func() {
		mSaldo := monthDug - monthPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "month-total",
			Fields: []string{
				lastMonthLabel,
				common.FormatNumberWithSystemLocale(monthDug, 2),
				common.FormatNumberWithSystemLocale(monthPot, 2),
				common.FormatNumberWithSystemLocale(mSaldo, 2),
			},
		})
		monthDug, monthPot = 0, 0
		lastMonthKey = ""
		lastMonthLabel = ""
	}

	appendMiTotal := func() {
		if params.SaldaPoMesecima && lastMonthKey != "" {
			appendMonthTotal()
		}
		miSaldo := miDug - miPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "mi-total",
			Fields: []string{
				translator.Label("Ukupno za mesto isporuke") + ": " + lastMI,
				common.FormatNumberWithSystemLocale(miDug, 2),
				common.FormatNumberWithSystemLocale(miPot, 2),
				common.FormatNumberWithSystemLocale(miSaldo, 2),
			},
		})
		miDug, miPot = 0, 0
		lastMIKey = ""
	}

	appendKontoTotal := func() {
		if lastMIKey != "" {
			appendMiTotal()
		}
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto + " | " + lastSifra + " | " + lastNaziv,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
		groupDug, groupPot = 0, 0
		runningSaldo = 0
		lastGroupKey = ""
	}

	for _, entity := range *entities {
		groupKey := entity.Konto + "|" + entity.Sifra
		miKey := entity.MI
		monthKey := ""
		monthLabel := ""
		if params.SaldaPoMesecima && entity.Danal.Valid {
			month := int(entity.Danal.Time.Month())
			year := entity.Danal.Time.Year()
			monthKey = fmt.Sprintf("%04d-%02d", year, month)
			monthLabel = fmt.Sprintf("%s %d", monthNames[month], year)
		}

		// Close month subtotal if month changed (same MI)
		if params.SaldaPoMesecima && monthKey != lastMonthKey && lastMonthKey != "" {
			appendMonthTotal()
		}

		// Close MI group if MI changed
		if miKey != lastMIKey && lastMIKey != "" {
			appendMiTotal()
		}

		// Close konto group if konto+sifra changed
		if groupKey != lastGroupKey && lastGroupKey != "" {
			appendKontoTotal()
		}

		// Open new konto group
		if groupKey != lastGroupKey {
			lastGroupKey = groupKey
			lastKonto = entity.Konto
			lastSifra = entity.Sifra
			lastNaziv = entity.Naziv
			runningSaldo = 0
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, entity.Sifra, entity.Naziv},
			})
		}

		// Open new MI group
		if miKey != lastMIKey {
			lastMIKey = miKey
			lastMI = miKey
			miDug, miPot = 0, 0
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "mi-header",
				Fields:   []string{miKey},
			})
		}

		// Open new month (track only)
		if params.SaldaPoMesecima && monthKey != lastMonthKey {
			lastMonthKey = monthKey
			lastMonthLabel = monthLabel
			monthDug, monthPot = 0, 0
		}

		// Data row
		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		miDug += dug
		miPot += pot
		monthDug += dug
		monthPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Opis,
				entity.OJ,
				entity.Dokum,
				entity.Tra,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	// Close remaining groups
	if lastGroupKey != "" {
		appendKontoTotal()
	}

	// Grand totals
	nHeaders := len(tbl.Headers)
	if nHeaders >= 3 {
		grandSaldo := grandDug - grandPot
		tbl.Totals[nHeaders-3] = common.FormatNumberWithSystemLocale(grandDug, 2)
		tbl.Totals[nHeaders-2] = common.FormatNumberWithSystemLocale(grandPot, 2)
		tbl.Totals[nHeaders-1] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
	}
	return nil
}

// GetPrometKarticaSintKontaStampa generates data for the "Promet Kartica Sintetickog Konta" report based on the provided parameters.
func (s *PrometResource) GetPrometKarticaSintKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	// Look up konto name from fkpl (sintetički konto vkonta=3)
	fkplQb := common.NewQueryBuilder(`SELECT konto, naziv FROM fkpl`, true)
	if hasGod {
		fkplQb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		fkplQb.AddEqual("kar", userSession.SelectedKar)
	}
	fkplQb.AddEqual("konto", params.OdKonta)
	fkplQb.AddEqual("vkonta", 3)
	fkplSql, fkplArgs := fkplQb.Build()
	fkplEntities, _ := s.fkplRepo.GetAllCustom(ctx, fkplSql, "", fkplArgs, "", "")
	if fkplEntities != nil && len(*fkplEntities) > 0 {
		tbl.ContentTitle = (*fkplEntities)[0].Naziv
	}

	var sqlQuery string
	var args []interface{}

	if !params.Analitika {
		// AGGREGATED MODE: one row per konto+sifra combination
		qb := common.NewQueryBuilder(`SELECT
			f.konto,
			COALESCE(f.sifra, '') as sifra,
			COALESCE(MAX(fk.naziv), '') as opis,
			SUM(CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END) as duguje,
			SUM(CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END) as potrazuje
		FROM fpro f`, true)
		qb.AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl")
		if hasGod {
			qb.AddEqual("f.god", userSession.SelectedGod)
		}
		if hasKar {
			qb.AddEqual("f.kar", userSession.SelectedKar)
		}
		qb.AddLikeBegin("f.konto", params.OdKonta)
		qb.AddCondition("f.danal", params.OdDatuma, ">=")
		qb.AddCondition("f.danal", params.DoDatuma, "<=")
		qb.AddGroupBy("f.konto, f.sifra")
		qb.AddOrderBy("f.konto, coalesce(NULLIF(f.sifra, ''), '0')::numeric")
		sqlQuery, args = qb.Build()
	} else {
		// DETAILED MODE: individual journal entries
		qb := common.NewQueryBuilder(`SELECT
			f.konto,
			COALESCE(f.sifra, '') as sifra,
			COALESCE(fk.naziv, '') as opis,
			CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
			CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje
		FROM fpro f`, true)
		qb.AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl")
		if hasGod {
			qb.AddEqual("f.god", userSession.SelectedGod)
		}
		if hasKar {
			qb.AddEqual("f.kar", userSession.SelectedKar)
		}
		qb.AddLikeBegin("f.konto", params.OdKonta)
		qb.AddCondition("f.danal", params.OdDatuma, ">=")
		qb.AddCondition("f.danal", params.DoDatuma, "<=")
		qb.AddOrderBy("f.konto, COALESCE(NULLIF(f.sifra, ''), '0')::numeric, f.danal, f.nalog::numeric")
		sqlQuery, args = qb.Build()
	}

	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	for i, entity := range *entities {
		dug := entity.Duguje
		pot := entity.Potrazuje
		saldo := dug - pot
		grandDug += dug
		grandPot += pot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", i+1),
				entity.Konto,
				entity.Sifra,
				entity.Opis,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(saldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	for i, header := range tbl.Headers {
		switch header.Name {
		case "opis":
			tbl.Totals[i] = translator.Label("Ukupno za listu")
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
		}
	}
	return nil
}

func (s *PrometResource) GetPrometSintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`SELECT
		f.tipdok, f.nalog, f.danal,
		COALESCE(f.opis, '') as opis,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(fkpl.naziv, '') as naziv
	FROM fpro f`, true)
	qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = f.idfkpl")

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddLikeBegin("f.konto", params.OdKonta)
	qb.AddCondition("f.danal", params.OdDatuma, ">=")
	qb.AddCondition("f.danal", params.DoDatuma, "<=")
	qb.AddOrderBy("f.konto, f.danal, f.tipdok, f.nalog::numeric")

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	var groupDug, groupPot, runningSaldo float64
	var monthDug, monthPot float64
	lastKonto := ""
	var lastNaziv, lastMonth string

	monthKey := func(entity domain.PrometDto) string {
		if !entity.Danal.Valid {
			return ""
		}
		return fmt.Sprintf("%02d/%d", int(entity.Danal.Time.Month()), entity.Danal.Time.Year())
	}

	for _, entity := range *entities {
		currentMonth := monthKey(entity)

		if entity.Konto != lastKonto {
			if lastKonto != "" {
				if params.SaldaPoMesecima && lastMonth != "" {
					tbl.Rows = append(tbl.Rows, domain.TableRow{
						ClassRow: "month-total",
						Fields: []string{
							translator.Label("Ukupno za mesec") + " " + lastMonth,
							common.FormatNumberWithSystemLocale(monthDug, 2),
							common.FormatNumberWithSystemLocale(monthPot, 2),
							common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
						},
					})
				}
				pSaldo := groupDug - groupPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "konto-total",
					Fields: []string{
						translator.Label("Ukupno za") + ": " + lastKonto,
						common.FormatNumberWithSystemLocale(groupDug, 2),
						common.FormatNumberWithSystemLocale(groupPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			groupDug, groupPot, runningSaldo = 0, 0, 0
			monthDug, monthPot = 0, 0
			lastKonto = entity.Konto
			lastNaziv = entity.Naziv
			lastMonth = currentMonth
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, lastNaziv},
			})
		} else if params.SaldaPoMesecima && currentMonth != lastMonth {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
			monthDug, monthPot = 0, 0
			lastMonth = currentMonth
		}

		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		monthDug += dug
		monthPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	if lastKonto != "" {
		if params.SaldaPoMesecima && lastMonth != "" {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
		}
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSubsintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`SELECT
		f.tipdok, f.nalog, f.danal,
		COALESCE(f.opis, '') as opis,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(fkpl.naziv, '') as naziv
	FROM fpro f`, true)
	qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = f.idfkpl")

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddCondition("f.konto::numeric", params.OdKonta, ">=")
	qb.AddCondition("f.konto::numeric", params.DoKonta, "<=")
	qb.AddCondition("f.danal", params.OdDatuma, ">=")
	qb.AddCondition("f.danal", params.DoDatuma, "<=")
	qb.AddOrderBy("f.konto, f.danal, f.tipdok, f.nalog::numeric")

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	var groupDug, groupPot, runningSaldo float64
	var monthDug, monthPot float64
	lastKonto := ""
	var lastNaziv, lastMonth string

	monthKey := func(entity domain.PrometDto) string {
		if !entity.Danal.Valid {
			return ""
		}
		return fmt.Sprintf("%02d/%d", int(entity.Danal.Time.Month()), entity.Danal.Time.Year())
	}

	for _, entity := range *entities {
		currentMonth := monthKey(entity)

		if entity.Konto != lastKonto {
			if lastKonto != "" {
				// Close out last month of previous konto
				if params.SaldaPoMesecima && lastMonth != "" {
					tbl.Rows = append(tbl.Rows, domain.TableRow{
						ClassRow: "month-total",
						Fields: []string{
							translator.Label("Ukupno za mesec") + " " + lastMonth,
							common.FormatNumberWithSystemLocale(monthDug, 2),
							common.FormatNumberWithSystemLocale(monthPot, 2),
							common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
						},
					})
				}
				// Emit konto-total
				pSaldo := groupDug - groupPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "konto-total",
					Fields: []string{
						translator.Label("Ukupno za") + ": " + lastKonto,
						common.FormatNumberWithSystemLocale(groupDug, 2),
						common.FormatNumberWithSystemLocale(groupPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			groupDug, groupPot, runningSaldo = 0, 0, 0
			monthDug, monthPot = 0, 0
			lastKonto = entity.Konto
			lastNaziv = entity.Naziv
			lastMonth = currentMonth
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, lastNaziv},
			})
		} else if params.SaldaPoMesecima && currentMonth != lastMonth {
			// Month changed within same konto
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
			monthDug, monthPot = 0, 0
			lastMonth = currentMonth
		}

		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		monthDug += dug
		monthPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	if lastKonto != "" {
		if params.SaldaPoMesecima && lastMonth != "" {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
		}
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetFieldCache() map[string]reflect.StructField {
	return s.service.GetFieldCache()
}

func (s *PrometResource) GetAnkontaTableFields() []domain.Fields {
	return s.prometAnKontaTableFields
}

func (s *PrometResource) GetAnKontaMiTableFields() []domain.Fields {
	return s.prometAnKontaMiTableFields
}
func (s *PrometResource) GetAnDeviznaKontaTableFields() []domain.Fields {
	return s.prometAnDeviznaKontaTableFields
}
func (s *PrometResource) GetAnDeviznaKontaRekapTableFields() []domain.Fields {
	return s.prometAnDeviznaKontaRekapTableFields
}
func (s *PrometResource) GetSubsintetickihKontaTableFields() []domain.Fields {
	return s.prometSubsintetickihKontaTablefields
}
func (s *PrometResource) GetSintetickihKontaTableFields() []domain.Fields {
	return s.prometSintetickihKontaTableFields
}

func (s *PrometResource) GetKarticaSintetikaTableFields() []domain.Fields {
	return s.prometKarticaSintetickihKontaTableFields
}
func (s *PrometResource) GetSubsintetikaVrdTableFields() []domain.Fields {
	return s.prometSubsintetikaVrdTableFields
}
func (s *PrometResource) GetKontaAnalitickiTableFields() []domain.Fields {
	return s.prometKontaAnalitickiTableFields
}

func (s *PrometResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

func (s *PrometResource) GetPrometAnalitickaKarticaStampaTableFields() []domain.Fields {
	return s.prometAnalitickaKarticaStampaTableFields
}

func (s *PrometResource) GetPrometSubsintetickihKontaStampaTableFields() []domain.Fields {
	return s.prometSubsintetickihKontaStampaTableFields
}

func (s *PrometResource) GetPrometSintetickihKontaStampaTableFields() []domain.Fields {
	return s.prometSintetickihKontaStampaTableFields
}

func (s *PrometResource) GetPrometKarticaSintKontaStampaTableFields() []domain.Fields {
	return s.prometKarticaSintKontaStampaTableFields
}

func (s *PrometResource) GetPrometSubsintetikaVrdStampaTableFields() []domain.Fields {
	return s.prometSubsintetikaVrdStampaTableFields
}

func (s *PrometResource) GetPrometAnalitickaKarticaPoMIStampaTableFields() []domain.Fields {
	return s.prometAnalitickaKarticaPoMIStampaTableFields
}

func (s *PrometResource) setServiceFieldValues() {
	s.prometAnKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrd", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina Dokumenta", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "Org. Jedinica", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolduguje", Label: "Količina Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolpotrazuje", Label: "Količina Potražuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezni Dokument", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Datum Veznog Dokumenta", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.travez", SkipInSearch: false},
		{Name: "mi", Label: "Mesto Isporuke", Width: "10", Field: "fpro.mi", SkipInSearch: false},
	}

	s.prometAnKontaMiTableFields = []domain.Fields{
		{Name: "mi", Label: "Mesto Isporuke", Width: "10", Field: "fpro.mi", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "VD", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "Dokum", Label: "Dokument", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina", Width: "4", Field: "fpro.tra", SkipInSearch: false},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolduguje", Label: "Količina Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolpotrazuje", Label: "Količina Potražuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezni Dokument", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Datum Veznog Dokumenta", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.travez", SkipInSearch: false},
	}

	s.prometAnDeviznaKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "fpro.kurs", SkipInSearch: true},
		{Name: "duguje", Label: "Devizno Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Devizno Potražuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "Dokum", Label: "Broj Dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.oj", SkipInSearch: true},
		{Name: "konto", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.konto", SkipInSearch: true},
	}
	s.prometAnDeviznaKontaRekapTableFields = []domain.Fields{
		{Name: "valuta", Label: "Valuta", Width: "30", Field: "fpro.sifval", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devduguje", Label: "Devizno  Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devpotrazuje", Label: "Devizno Potražuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devsaldo", Label: "Devizni Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetickihKontaTablefields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "3", Field: "fpro.konto", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "3", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "6", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "6", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "40", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSintetickihKontaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "4", Field: "fpro.tipdok", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKarticaSintetickihKontaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10", Field: "fpro.konto", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifra", Label: "Sifra", Width: "10", Field: "fpro.sifra", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetikaVrdTableFields = []domain.Fields{
		{Name: "vkonta", Label: "Vrsta Konta", Width: "10", Field: "fpro.vkonta", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "vrd", Label: "Vrsta knjizenja", Width: "10", Field: "fpro.vrd", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetikaVrdStampaTableFields = []domain.Fields{
		{Name: "vkonta", Label: "Vrsta konta", Width: "15", Field: "vkonta", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "15", Field: "fpro.vrd", SkipInSearch: true, TextAlign: "center"},
		{Name: "duguje", Label: "Duguje", Width: "20", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "20", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "20", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometAnalitickaKarticaPoMIStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "4", Field: "f.vrd", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "30", Field: "f.opis", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "f.ojozn", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "f.dokum", SkipInSearch: true},
		{Name: "tra", Label: "Poslovna godina", Width: "5", Field: "f.tra", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "f.dadok", SkipInSearch: true},
		{Name: "rok", Label: "Rok", Width: "4", Field: "f.rok", SkipInSearch: true},
		{Name: "brst", Label: "Redni broj", Width: "5", Field: "f.brst", SkipInSearch: true, TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKontaAnalitickiTableFields = []domain.Fields{
		{Name: "sifra", Label: "Sifra", Width: "10", Field: "fpro.sifra", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "naziv", Label: "Naziv", Width: "30", Field: "fpro.naziv", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "dokum", Label: "Broj Dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: false},
		{Name: "tra", Label: "Godina Dokumenta", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolduguje", Label: "Količina Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolpotrazuje", Label: "Količina Potražuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezni Dokument", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Datum Veznog Dokumenta", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.travez", SkipInSearch: false},
	}
	s.prometAnalitickaKarticaStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "4", Field: "f.vrd", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "30", Field: "f.opis", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "f.ojozn", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "f.dokum", SkipInSearch: true},
		{Name: "tra", Label: "Poslovna godina", Width: "5", Field: "f.tra", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "f.dadok", SkipInSearch: true},
		{Name: "rok", Label: "Rok", Width: "4", Field: "f.rok", SkipInSearch: true},
		{Name: "brst", Label: "Redni broj", Width: "5", Field: "f.brst", SkipInSearch: true, TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetickihKontaStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "40", Field: "f.opis", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSintetickihKontaStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "40", Field: "f.opis", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKarticaSintKontaStampaTableFields = []domain.Fields{
		{Name: "rednibr", Label: "Red. Br.", Width: "5", Field: "rednibr", SkipInSearch: true, TextAlign: "center"},
		{Name: "konto", Label: "Konto", Width: "8", Field: "f.konto", SkipInSearch: true},
		{Name: "sifra", Label: "Sifra", Width: "8", Field: "f.sifra", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "40", Field: "opis", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
}

package finansijsko

import (
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"

	"github.com/gin-gonic/gin"
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
	GetPrometAnalitickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, isMI bool) error
	GetPrometSubsintetickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPrometSintetickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPrometKarticaSintetickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPrometKontaAnaliticki(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPrometSubsintetikaVrd(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPrometTotals(c *gin.Context) (domain.PrometResponse, error)
	GetAnkontaTableFields() []domain.Fields
	GetAnKontaMiTableFields() []domain.Fields
	GetAnDeviznaKontaTableFields() []domain.Fields
	GetSubsintetickihKontaTableFields() []domain.Fields
	GetSintetickihKontaTableFields() []domain.Fields
	GetKarticaSintetikaTableFields() []domain.Fields
	GetSubsintetikaVrdTableFields() []domain.Fields
	GetKontaAnalitickiTableFields() []domain.Fields
}

// PrometResource implements the PrometService interface.
type PrometResource struct {
	service                                  *service.BaseService[domain.PrometDto]
	prometRepo                               *repository.BaseRepository[domain.PrometDto]
	fkplRepo                                 *repository.BaseRepository[domain.Fkpl]
	prometAnKontaTableFields                 []domain.Fields
	prometAnKontaMiTableFields               []domain.Fields
	prometAnDeviznaKontaTableFields          []domain.Fields
	prometSubsintetickihKontaTablefields     []domain.Fields
	prometSintetickihKontaTableFields        []domain.Fields
	prometKarticaSintetickihKontaTableFields []domain.Fields
	prometSubsintetikaVrdTableFields         []domain.Fields
	prometKontaAnalitickiTableFields         []domain.Fields
}

func NewPrometService(service *service.BaseService[domain.PrometDto], prometRepo *repository.BaseRepository[domain.PrometDto], fkplRepo *repository.BaseRepository[domain.Fkpl]) *PrometResource {
	rs := &PrometResource{
		service:    service,
		prometRepo: prometRepo,
		fkplRepo:   fkplRepo,
	}
	rs.setServiceFieldValues()
	return rs
}
func (s *PrometResource) GetPrometTotals(c *gin.Context) (domain.PrometResponse, error) {
	var response domain.PrometResponse
	var konto, sifra, odDatuma, doDatuma, odMI, doMI, odkonta, dokonta, odsifre, dosifre string
	reportTip := c.Query("tabname")
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		konto = c.Query("konto")
		sifra = c.Query("sifra")
		odDatuma = c.Query("oddatuma")
		doDatuma = c.Query("dodatuma")
		if konto == "" || sifra == "" || odDatuma == "" || doDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "prometankontami": // Analiticka Konta po MI
		konto = c.Query("konto")
		sifra = c.Query("sifra")
		odDatuma = c.Query("oddatuma")
		doDatuma = c.Query("dodatuma")
		odMI = c.Query("odmi")
		doMI = c.Query("domi")
		if konto == "" || sifra == "" || odDatuma == "" || doDatuma == "" || doMI == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "deviznahanalitickihkonta": // Devizna Analiticka Konta
	case "subsintetickakonta": // Subsinteticka Konta
		odkonta = c.Query("odkonta")
		dokonta = c.Query("dokonta")
		odDatuma = c.Query("oddatuma")
		doDatuma = c.Query("dodatuma")
		if odkonta == "" || dokonta == "" || odDatuma == "" || doDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		konto = c.Query("konto")
		odDatuma = c.Query("oddatuma")
		doDatuma = c.Query("dodatuma")
		if konto == "" || odDatuma == "" || doDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "kontaanaliticki": // Konta Analiticki
		konto = c.Query("konto")
		odsifre = c.Query("odsifre")
		dosifre = c.Query("dosifre")
		odDatuma = c.Query("oddatuma")
		doDatuma = c.Query("dodatuma")
		if konto == "" || odsifre == "" || dosifre == "" || odDatuma == "" || doDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	default:
		return response, fmt.Errorf("invalid report tip")
	}

	session := domain.GetSessionFromContext(c)
	if session == nil {
		return response, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	// Get "promet do" totals (up to start date)
	qbDo := common.NewQueryBuilder(`
		select 
			coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
		from fpro`)

	if hasGod {
		qbDo.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qbDo.AddEqual("kar", session.SelectedKar)
	}
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		qbDo.AddEqual("konto", konto)
		qbDo.AddEqual("sifra", sifra)
		qbDo.AddEqual("vkonta", 1)
		qbDo.AddCondition("danal", odDatuma, "<")
	case "prometankontami": // Analiticka Konta po MI
		qbDo.AddEqual("konto", konto)
		qbDo.AddEqual("sifra", sifra)
		qbDo.AddEqual("vkonta", 1)
		qbDo.AddCondition("danal", odDatuma, "<")
		qbDo.AddCondition("mi", odMI, ">=")
		qbDo.AddCondition("mi", doMI, "<=")
	case "deviznahanalitickihkonta": // Devizna Analiticka Konta
	case "subsintetickakonta": // Subsinteticka Konta
		qbDo.AddCondition("konto", odkonta, ">=")
		qbDo.AddCondition("konto", dokonta, "<=")
		qbDo.AddCondition("danal", odDatuma, "<")

	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		qbDo.AddLikeBegin("konto", konto)
		qbDo.AddCondition("danal", odDatuma, "<")
	case "kontaanaliticki": // Konta Analiticki
		qbDo.AddEqual("konto", konto)
		qbDo.AddCondition("sifra::numeric", odsifre, ">=")
		qbDo.AddCondition("sifra::numeric", dosifre, "<=")
		qbDo.AddCondition("danal", odDatuma, "<")
		qbDo.AddEqual("vkonta", 1)
	default:
	}

	prometDoQuery, prometDoArgs := qbDo.Build()
	prometDoResults, err := s.prometRepo.GetAllCustom(c, prometDoQuery, "", prometDoArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet do totals: %v", err)
	}

	var prometDoDuguje, prometDoPotrazuje float64
	if len(*prometDoResults) > 0 {
		prometDoDuguje = (*prometDoResults)[0].Duguje
		prometDoPotrazuje = (*prometDoResults)[0].Potrazuje
	}

	response.Totals = domain.TotalValues{
		DugDo:   prometDoDuguje,
		PotDo:   prometDoPotrazuje,
		SaldoDo: prometDoDuguje - prometDoPotrazuje,
	}

	// Get "promet za period" totals (for the specified period)
	qbPeriod := common.NewQueryBuilder(`
		select 
			coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
		from fpro`)

	if hasGod {
		qbPeriod.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qbPeriod.AddEqual("kar", session.SelectedKar)
	}
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		qbPeriod.AddEqual("konto", konto)
		qbPeriod.AddEqual("sifra", sifra)
		qbPeriod.AddEqual("vkonta", 1)
		qbPeriod.AddCondition("danal", odDatuma, ">=")
		qbPeriod.AddCondition("danal", doDatuma, "<=")
	case "prometankontami": // Analiticka Konta po MI
		qbPeriod.AddEqual("konto", konto)
		qbPeriod.AddEqual("sifra", sifra)
		qbPeriod.AddEqual("vkonta", 1)
		qbPeriod.AddCondition("danal", odDatuma, ">=")
		qbPeriod.AddCondition("danal", doDatuma, "<=")
		qbPeriod.AddCondition("mi", odMI, ">=")
		qbPeriod.AddCondition("mi", doMI, "<=")
	case "deviznahanalitickihkonta": // Devizna Analiticka Konta
	case "subsintetickakonta": // Subsinteticka Konta
		qbPeriod.AddCondition("konto", odkonta, ">=")
		qbPeriod.AddCondition("konto", dokonta, "<=")
		qbPeriod.AddCondition("danal", odDatuma, ">=")
		qbPeriod.AddCondition("danal", doDatuma, "<=")
	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		qbPeriod.AddLikeBegin("konto", konto)
		qbPeriod.AddCondition("danal", odDatuma, ">=")
		qbPeriod.AddCondition("danal", doDatuma, "<=")
	case "kontaanaliticki": // Konta Analiticki
		qbPeriod.AddEqual("konto", konto)
		qbPeriod.AddCondition("sifra::numeric", odsifre, ">=")
		qbPeriod.AddCondition("sifra::numeric", dosifre, "<=")
		qbPeriod.AddCondition("danal", odDatuma, ">=")
		qbPeriod.AddCondition("danal", doDatuma, "<=")
		qbPeriod.AddEqual("vkonta", 1)
	default:
	}

	prometPeriodQuery, prometPeriodArgs := qbPeriod.Build()
	prometPeriodResults, err := s.prometRepo.GetAllCustom(c, prometPeriodQuery, "", prometPeriodArgs, "", "")
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

func (s *PrometResource) CheckPrometParameters(c *gin.Context, requiredFields []string) (fieldsError []domain.FieldError) {

	fieldsError = common.ValidateRequiredParams(c, requiredFields)
	if len(fieldsError) > 0 {
		return
	}
	// Build query dynamically
	qb := common.NewQueryBuilder(`SELECT f.konto, f.sifra FROM baza.fkpl as f`)

	session := domain.GetSessionFromContext(c)
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
	qb.AddEqual("f.konto", c.Query("konto"))
	qb.AddEqual("f.sifra", c.Query("sifra"))
	qb.AddEqual("f.vkonta", c.Query("vkonta"))

	sqlQuery, args := qb.Build()

	entities, err := s.service.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetData}}
	}
	if len(*entities) == 0 {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetKontoSifra}}
	}

	return nil
}

func (s *PrometResource) GetPrometAnalitickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, isMI bool) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	odMI := c.Query("odmi")
	doMI := c.Query("domi")
	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT danal, tipdok, concat(tipdok,'-',nalog) as nalog, idfpro, kat, iznos, kolic, 
		       	vrd, dokum, dadok, rok, tra, ojozn as oj, opis, sifval, kurs, 
		       	deviznos, cena, konto, idfnal, idfkpl, dokumv, dadokv, travez, rdokid,
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				CASE WHEN kat = 1 OR kat = 2 THEN kolic ELSE 0 END as kolduguje,
				CASE WHEN kat = 3 OR kat = 4 THEN kolic ELSE 0 END as kolpotrazuje	
		FROM fpro`)

	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	qb.AddEqual("konto", konto)
	qb.AddEqual("sifra", sifra)
	qb.AddEqual("vkonta", 1)
	qb.AddCondition("danal", odDatuma, ">=")
	qb.AddCondition("danal", doDatuma, "<=")
	if isMI {
		qb.AddCondition("mi", odMI, ">=")
		qb.AddCondition("mi", doMI, "<=")
	}
	// Add search conditions if search text is provided
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		if isMI {
			qb.AddSearchConditions(s.GetAnKontaMiTableFields(), searchText)
		} else {
			qb.AddSearchConditions(s.GetAnkontaTableFields(), searchText)
		}
	}

	qb.AddOrderBy("god, kar, danal, tipdok, nalog, idfpro")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
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
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSubsintetickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	odkonta := c.Query("odkonta")
	dokonta := c.Query("dokonta")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT  konto, tipdok, nalog, danal, opis, 
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				idfpro
		FROM fpro`)

	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	qb.AddCondition("konto::numeric", odkonta, ">=")
	qb.AddCondition("konto::numeric", dokonta, "<=")
	qb.AddCondition("danal", odDatuma, ">=")
	qb.AddCondition("danal", doDatuma, "<=")

	// Add search conditions if search text is provided
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetSubsintetickihKontaTableFields(), searchText)
	}
	qb.AddOrderBy("konto, danal, tipdok, nalog")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
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
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSintetickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	odkonta := c.Query("konto")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT LEFT(konto, 3) as konto, tipdok, nalog, 
			   	MAX(danal) as danal, 
			   	MAX(opis) as opis,
			   	SUM(CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END) as potrazuje
		FROM fpro`)

	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	qb.AddLikeBegin("konto", odkonta)
	qb.AddCondition("danal", odDatuma, ">=")
	qb.AddCondition("danal", doDatuma, "<=")

	// Add search conditions if search text is provided
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetSintetickihKontaTableFields(), searchText)
	}
	qb.AddGroupBy("LEFT(konto, 3), tipdok, nalog")
	qb.AddOrderBy("tipdok, nalog")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
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
		}
	}

	return nil
}

func (s *PrometResource) GetPrometKarticaSintetickihKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	odkonta := c.Query("konto")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	analitika := c.Query("analitika")

	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	sqlText := ""
	groupByClause := ""
	if analitika == "true" {
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

	qb := common.NewQueryBuilder(sqlText)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddLikeBegin("fpro.konto", odkonta)
	qb.AddCondition("danal", odDatuma, ">=")
	qb.AddCondition("danal", doDatuma, "<=")

	// Add search conditions if search text is provided
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetKarticaSintetikaTableFields(), searchText)
	}
	qb.AddGroupBy(groupByClause)
	qb.AddOrderBy("fpro.konto::numeric")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
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
			if analitika == "true" {
				fields[1] = entity.Sifra
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSubsintetikaVrd(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	odkonta := c.Query("konto")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table

	qb := common.NewQueryBuilder(`SELECT 
				CASE fpro.vkonta
				WHEN '1' THEN 'Analitika'
				WHEN '2' THEN 'Subsintetika'
				ELSE ''
				END as vkonta,
				fpro.vrd,
				SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro`)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddLikeBegin("fpro.konto", odkonta)
	qb.AddCondition("danal", odDatuma, ">=")
	qb.AddCondition("danal", doDatuma, "<=")

	// Add search conditions if search text is provided
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetSubsintetikaVrdTableFields(), searchText)
	}
	qb.AddGroupBy(`fpro.vrd, fpro.konto, fpro.vkonta`)
	qb.AddOrderBy("fpro.konto::numeric")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
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
		}
	}

	return nil
}

func (s *PrometResource) GetPrometKontaAnaliticki(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	konto := c.Query("konto")
	odSifre := c.Query("odsifre")
	doSifre := c.Query("dosifre")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.prometRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT f.danal, f.tipdok, concat(f.tipdok,'-',nalog) as nalog, f.idfpro, f.kat, f.iznos, f.kolic, 
		       	f.vrd, f.dokum, f.dadok, f.rok, f.tra, f.ojozn as oj, f.opis, f.sifval, f.kurs,
		       	f.deviznos, f.cena, f.konto, f.sifra, fk.naziv, f.idfkpl, f.dokumv, f.dadokv, f.travez, f.rdokid,
			   	CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
			   	CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
				CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.kolic ELSE 0 END as kolduguje,
				CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.kolic ELSE 0 END as kolpotrazuje	
		FROM fpro as f`)
	qb.AddJoin(` LEFT JOIN fkpl fk on fk.idfkpl = f.idfkpl`)
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", session.SelectedKar)
	}
	qb.AddEqual("f.konto", konto)
	qb.AddCondition("f.sifra::numeric", odSifre, ">=")
	qb.AddCondition("f.sifra::numeric", doSifre, "<=")
	qb.AddEqual("f.vkonta", 1)
	qb.AddCondition("f.danal", odDatuma, ">=")
	qb.AddCondition("f.danal", doDatuma, "<=")
	// Add search conditions if search text is provided
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.PrometDto{}))
		qb.AddSearchConditions(s.GetKontaAnalitickiTableFields(), searchText)

	}

	qb.AddOrderBy("f.sifra::numeric, f.danal, f.tipdok, f.nalog")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
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
func (s *PrometResource) setServiceFieldValues() {
	s.prometAnKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrd", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "Dokum", Label: "Broj dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina Dokumenta", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "Org. Jedinica", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true},
		{Name: "kolduguje", Label: "Kolic. Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true},
		{Name: "kolpotrazuje", Label: "Kolic. Potrazuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezn. Dokum.", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Dat. Vezn. Dokum.", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God. Vezn. Dokum. ", Width: "4", Field: "fpro.travez", SkipInSearch: false},
		{Name: "mi", Label: "Mesto Isporuke", Width: "10", Field: "fpro.mi", SkipInSearch: false},
	}

	s.prometAnKontaMiTableFields = []domain.Fields{
		{Name: "mi", Label: "Mesto Isporuke", Width: "10", Field: "fpro.mi", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "VD", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "Dokum", Label: "Dokument", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Dat. Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina", Width: "4", Field: "fpro.tra", SkipInSearch: false},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true},
		{Name: "kolduguje", Label: "Kolic. Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true},
		{Name: "kolpotrazuje", Label: "Kolic. Potrazuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezn. Dokum.", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Dat. Vezn. Dokum.", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God. Vezn. Dokum. ", Width: "4", Field: "fpro.travez", SkipInSearch: false},
	}

	s.prometAnDeviznaKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10"},
		{Name: "danal", Label: "Datum Naloga", Width: "10"},
		{Name: "vrd", Label: "VD", Width: "4"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "sifval", Label: "Sifval", Width: "4"},
		{Name: "kurs", Label: "Kurs", Width: "6"},
		{Name: "duguje", Label: "Dev. Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Dev. Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
		{Name: "Dokum", Label: "Dokument", Width: "6"},
		{Name: "dadok", Label: "Dat. Dokumenta", Width: "6"},
		{Name: "rok", Label: "Rok", Width: "4"},
		{Name: "tra", Label: "Godina", Width: "4"},
		{Name: "oj", Label: "OJ", Width: "4"},
		{Name: "konto", Label: "God. Vezn. Dokum. ", Width: "4"},
	}
	s.prometSubsintetickihKontaTablefields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "3", Field: "fpro.konto", SkipInSearch: false},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "3", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "6", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "6", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "40", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
	}
	s.prometSintetickihKontaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "4", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
	}
	s.prometKarticaSintetickihKontaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Sifra", Width: "10", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
	}
	s.prometSubsintetikaVrdTableFields = []domain.Fields{
		{Name: "vkonta", Label: "Vrsta Konta", Width: "10", Field: "fpro.vkonta", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta knjizenja", Width: "10", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
	}
	s.prometKontaAnalitickiTableFields = []domain.Fields{
		{Name: "sifra", Label: "Sifra", Width: "10", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv", Width: "30", Field: "fpro.naziv", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "VD", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "dokum", Label: "Dokument", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Dat. Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: false},
		{Name: "tra", Label: "Godina", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true},
		{Name: "kolduguje", Label: "Kolic. Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true},
		{Name: "kolpotrazuje", Label: "Kolic. Potrazuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezn. Dokum.", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Dat. Vezn. Dokum.", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God. Vezn. Dokum. ", Width: "4", Field: "fpro.travez", SkipInSearch: false},
		{Name: "rdokid", Label: "Rdokid", Width: "6", Field: "fpro.rdokid", SkipInSearch: true},
	}
}

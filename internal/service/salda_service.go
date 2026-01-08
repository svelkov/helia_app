package service

import (
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// SaldaViewData encapsulates all data needed for the Salda display page.
type SaldaViewData struct {
	FproEntities []domain.Fpro
	TableData    domain.TableData
}

// SaldaService defines the interface for operations related to Salda.
type SaldaService interface {
	GetPojedKontaTableFields() []domain.Fields
	GetFieldCache() map[string]reflect.StructField
	GetSaldaPojedinacnihKonta(c *gin.Context, getTotalRecords bool, calculatedPageSize, currentPage int) (domain.SaldaResponse, error)
	GetSaldaGrupeKonta(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error
	GetTotalRecordsCustom(queryText, whereText string, args []interface{}, limitOffset, orderBy string) (int, error)
	CheckSaldaParameters(c *gin.Context) []domain.FieldError
	SetDefaultTableData(tbl *domain.TableData)
	GetSaldaTotalValues(c *gin.Context) (domain.SaldaDto, error)
	GetSaldaPartneriList(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error
	ProcessSaldaPartneriDetails(c *gin.Context, idPartneri int64, tblKonta, tblDetalji *domain.TableData) error
	GetSaldaPartneriPrelomljeno(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) (domain.TableData, error)
	SaldaPoKomercijalistima(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error
	RealizacijaKomercijalisti(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error
	GetSaldaPartneriTableFields() []domain.Fields
	GetGrupeKontaTableFields() []domain.Fields
	GetSaldaPartneriHeaderTableFields() []domain.Fields
	GetSaldaPartneriDetailTableFields() []domain.Fields
	GetSaldaPartneriPrelomljenoTableFields() []domain.Fields
	GetKomercijalistiTableFields() []domain.Fields
	GetRealizacijaKomercijalistiTableFields() []domain.Fields
}

// SaldaResource implements the SaldaService interface.
type SaldaResource struct {
	service                                   *BaseService[domain.SaldaDto]
	saldaRepo                                 *repository.BaseRepository[domain.SaldaDto]
	fkplRepo                                  *repository.BaseRepository[domain.Fkpl]
	fproRepo                                  *repository.BaseRepository[domain.Fpro]
	partneriRepo                              *repository.BaseRepository[domain.Partneri]
	saldaPartneriRepo                         *repository.BaseRepository[domain.SaldaPartnerDto]
	saldaKomRepo                              *repository.BaseRepository[domain.SaldaKomercijalistiDto]
	saldaPojedinacniTableFields               []domain.Fields
	saldaGrupeKontaTableFields                []domain.Fields
	saldaPartneriTableFields                  []domain.Fields
	saldaPartneriHeaderTableFields            []domain.Fields
	saldaPartneriDetailTableFields            []domain.Fields
	saldaPartneriPrelomljenoTableFields       []domain.Fields
	saldaKomercijalistiTableFields            []domain.Fields
	saldaRealizacijakomercijalistiTableFields []domain.Fields
}

func NewSaldaService(service *BaseService[domain.SaldaDto],
	saldaRepo *repository.BaseRepository[domain.SaldaDto],
	fkplRepo *repository.BaseRepository[domain.Fkpl],
	fproRepo *repository.BaseRepository[domain.Fpro],
	partneriRepo *repository.BaseRepository[domain.Partneri],
	saldaPartneriRepo *repository.BaseRepository[domain.SaldaPartnerDto],
	saldaKomRepo *repository.BaseRepository[domain.SaldaKomercijalistiDto]) *SaldaResource {
	rs := &SaldaResource{
		service:           service,
		saldaRepo:         saldaRepo,
		fkplRepo:          fkplRepo,
		fproRepo:          fproRepo,
		partneriRepo:      partneriRepo,
		saldaPartneriRepo: saldaPartneriRepo,
		saldaKomRepo:      saldaKomRepo,
	}
	rs.setServiceFieldValues()
	rs.setServiceSaldaPartneriFieldValues()
	return rs
}

func (s *SaldaResource) CheckSaldaParameters(c *gin.Context, requiredFields []string) (fieldsError []domain.FieldError) {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "user session not found"}}
	}

	fieldsError = common.ValidateRequiredParams(c, requiredFields)
	if len(fieldsError) > 0 {
		return
	}
	// Build query dynamically
	qb := common.NewQueryBuilder(`SELECT f.konto, f.sifra FROM fkpl as f`)

	// Add system conditions
	hasGod, hasKAr := s.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}

	// Add user conditions
	qb.AddEqual("f.konto", c.Query("konto"))
	if c.Query("sifra") != "" {
		qb.AddEqual("f.sifra", c.Query("sifra"))
	}
	qb.AddEqual("f.vkonta", c.Query("tipkonta"))

	sqlQuery, args := qb.Build()

	entities, err := s.fkplRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetData}}
	}
	if len(*entities) == 0 {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetKontoSifra}}
	}

	return nil
}
func (s *SaldaResource) CheckSaldaGrupeParameters(c *gin.Context, requiredFields []string) (fieldsError []domain.FieldError) {
	fieldsError = []domain.FieldError{}

	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "user session not found"}}
	}

	nduzsin := 3 // Default length of synthetic account TODO should be get from config
	comboTipIzv := c.Query("cbx_tipizvestaja")
	od_konta := c.Query("od_konta")
	do_konta := c.Query("do_konta")
	od_sifre := c.Query("od_sifre")
	do_sifre := c.Query("do_sifre")
	cbx_klasa := c.Query("cbx_klasa")

	switch comboTipIzv {
	case "analitika":
		fieldsError = append(fieldsError, validateKontoFields(od_konta, do_konta, nduzsin, "<")...)
		fieldsError = append(fieldsError, validateSifraFields(od_sifre, do_sifre)...)

	case "subsintetika", "klasa_subsintetika":
		fieldsError = append(fieldsError, validateKontoFields(od_konta, do_konta, nduzsin, "<")...)
	case "sintetika":
		fieldsError = append(fieldsError, validateKontoFields(od_konta, do_konta, nduzsin, "=")...)

	case "klasa_sifra":
		fieldsError = append(fieldsError, validateSifraFields(od_sifre, do_sifre)...)
		if len(cbx_klasa) <= 0 {
			fieldsError = append(fieldsError, domain.FieldError{Field: "cbx_klasa", ErrorMessage: "Morate izabrati klasu konta!!!"})
		}
	}

	return fieldsError
}

func getKontoErrorMessage(nduzsin int, operator string) string {
	if operator == "=" {
		if nduzsin == 3 {
			return "Konto mora biti trocifreni!!!"
		}
		return "Konto mora biti cetvorocifreni!!!"
	}
	// operator == "<"
	if nduzsin == 3 {
		return "Konto mora biti bar cetvorocifren!!!"
	}
	return "Konto mora biti bar petocifren!!!"
}

func validateKontoFields(od_konta, do_konta string, nduzsin int, operator string) []domain.FieldError {
	var errors []domain.FieldError
	msg := getKontoErrorMessage(nduzsin, operator)

	if operator == "<" {
		if len(od_konta) < nduzsin+1 {
			errors = append(errors, domain.FieldError{Field: "od_konta", ErrorMessage: msg})
		}
		if len(do_konta) < nduzsin+1 {
			errors = append(errors, domain.FieldError{Field: "do_konta", ErrorMessage: msg})
		}
	} else {
		if len(od_konta) != nduzsin {
			errors = append(errors, domain.FieldError{Field: "od_konta", ErrorMessage: msg})
		}
		if len(do_konta) != nduzsin {
			errors = append(errors, domain.FieldError{Field: "do_konta", ErrorMessage: msg})
		}
	}

	return errors
}

func validateSifraFields(od_sifre, do_sifre string) []domain.FieldError {
	var errors []domain.FieldError
	msg := "Sifra mora biti bar dvocifrena!!!"

	if len(od_sifre) < 2 {
		errors = append(errors, domain.FieldError{Field: "od_sifre", ErrorMessage: msg})
	}
	if len(do_sifre) < 2 {
		errors = append(errors, domain.FieldError{Field: "do_sifre", ErrorMessage: msg})
	}

	return errors
}
func (s *SaldaResource) GetSaldaPojedinacnihKonta(c *gin.Context) (domain.SaldaResponse, error) {
	var response domain.SaldaResponse

	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return response, fmt.Errorf("user session not found")
	}

	// Create first query (opening balance - tipdok = '00')
	qb1 := common.NewQueryBuilder(`SELECT 0 as mesec,
    COALESCE(SUM(CASE WHEN kat = '1' OR kat = '2' THEN iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN kat = '3' OR kat = '4' THEN iznos ELSE 0 END), 0) as pot
    FROM fpro f`)

	// Add conditions to first query
	hasGod, hasKAr := s.fproRepo.GetHasGodHasKar()
	if hasGod {
		qb1.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb1.AddEqual("f.kar", userSession.SelectedKar)
	}
	vkonta := c.Query("tipkonta")
	if vkonta != "3" {
		qb1.AddEqual("f.konto", c.Query("konto"))
		if c.Query("sifra") != "" {
			qb1.AddEqual("f.sifra", c.Query("sifra"))
		}
	}
	qb1.AddEqual("f.tipdok", "00")

	// Add vkonta conditions.
	switch vkonta {
	case "1":
		qb1.AddEqual("f.vkonta", vkonta)
	case "2":
		qb1.AddIn("f.vkonta", []interface{}{"1", "2"})
	case "3":
		qb1.AddLike("f.konto", c.Query("konto"))
	}

	qb1.AddGroupBy("EXTRACT(MONTH FROM danal)")
	qb1.AddOrderBy("mesec ASC")
	// Create second query (monthly data - tipdok != '00')
	qb2 := common.NewQueryBuilder(`SELECT EXTRACT(MONTH FROM danal) as mesec,
    COALESCE(SUM(CASE WHEN kat = '1' OR kat = '2' THEN iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN kat = '3' OR kat = '4' THEN iznos ELSE 0 END), 0) as pot
    FROM fpro f`)

	// Add same base conditions
	if hasGod {
		qb2.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb2.AddEqual("f.kar", userSession.SelectedKar)
	}
	if vkonta != "3" {
		qb2.AddEqual("f.konto", c.Query("konto"))
		if c.Query("sifra") != "" {
			qb2.AddEqual("f.sifra", c.Query("sifra"))
		}
	}
	// Add vkonta conditions.
	switch vkonta {
	case "1":
		qb2.AddEqual("f.vkonta", vkonta)
	case "2":
		qb2.AddIn("f.vkonta", []interface{}{"1", "2"})
	case "3":
		qb2.AddLike("f.konto", c.Query("konto"))
	}

	qb2.AddCondition("f.tipdok", "00", "!=")
	qb2.AddGroupBy("EXTRACT(MONTH FROM danal)")
	// Create UNION
	uqb := common.NewUnionQueryBuilder("UNION ALL")
	uqb.AddQuery(qb1)
	uqb.AddQuery(qb2)
	uqb.AddOrderBy("mesec")

	sqlQuery, args := uqb.Build()

	entities, err := s.fproRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return response, err
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
					Duguje:    entity.Dug,
					Potrazuje: entity.Pot,
					Saldo:     entity.Dug - entity.Pot,
				}
				kumulSaldo = kumulSaldo + (entity.Dug - entity.Pot)
				templateData[entity.Mesec] = saldoDto
				templateData[entity.Mesec].SaldoKumul = kumulSaldo
			}
		}
	}

	response.Data = templateData
	return response, nil
}
func (s *SaldaResource) GetSaldaGrupeKonta(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	gnGod := userSession.SelectedGod
	gnKar := userSession.SelectedKar
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Get filter parameters with error handling
	odKonta := c.Query("od_konta")
	doKonta := c.Query("do_konta")
	odSifra := c.Query("od_sifre")
	doSifra := c.Query("do_sifre")
	//chkSaldaValute := utils.ParseStringToBool(c.Query("chk_salda_valute"))
	cbxTipIzvestaja := c.Query("cbx_tipizvestaja")
	cbxKlasa := c.Query("cbx_klasa")

	cbx_odmeseca := c.Query("cbx_odmeseca")
	cbx_domeseca := c.Query("cbx_domeseca")

	chkSaldoRazl0 := common.StringToBool(c.Query("chk_saldo_razl_nula"))
	chkSaldoVece0 := common.StringToBool(c.Query("chk_saldo_vece_nula"))
	chkSaldoManje0 := common.StringToBool(c.Query("chk_saldo_manje_nula"))
	chkSaldoJednako0 := common.StringToBool(c.Query("chk_saldo_nula"))

	odSalda := common.StringToFloat64(c.Query("od_salda"))
	doSalda := common.StringToFloat64(c.Query("do_salda"))
	searchText := c.Query("query")

	tbl.SearchEnabled = false
	tbl.ShowPagination = true
	tbl.ShowActions = false
	tbl.BtnAdd.IsVisible = false
	tbl.BtnPrint.IsVisible = false
	tbl.Pagination.CurrentPage = currentPage
	tbl.Pagination.StartRecord = (currentPage-1)*pageSize + 1
	tbl.Pagination.EndRecord = tbl.Pagination.StartRecord + pageSize - 1
	tbl.Pagination.PageSize = pageSize
	// Build combined query with partner join (eliminates N+1)
	// Separate sums for početno stanje (tipdok='00') and other promet (tipdok!='00')
	qb := common.NewQueryBuilder(`
		SELECT 
			fkpl.konto, fkpl.sifra, fkpl.naziv as kontonaziv,
			COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) as pstdug,
			COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pstpot,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) - COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pstsaldo,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) as dug,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pot,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) - COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as saldo,
			fkpl.idpartneri,
			COALESCE(p.naziv, '') as naziv,
			COALESCE(p.adresa, '') as adresa,
			COALESCE(p.mesto, '') as mesto,
			COALESCE(p.pib, '') as pib,
			COALESCE(p.jmbg, '') as jmbg,
			COALESCE(p.bpg, '') as bpg,
			COALESCE(p.index, '') as brindex		
		FROM fpro`)

	qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("LEFT JOIN partneri p ON p.idpartneri = fkpl.idpartneri")

	if hasGod {
		qb.AddEqual("fpro.god", gnGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", gnKar)
	}
	qb.AddCondition("fkpl.konto", odKonta, ">=")
	qb.AddCondition("fkpl.konto", doKonta, "<=")
	qb.AddCondition("fkpl.sifra", odSifra, ">=")
	qb.AddCondition("fkpl.sifra", doSifra, "<=")
	if cbxTipIzvestaja == "analitika" {
		qb.AddEqual("fkpl.vkonta", "1")
	}
	if cbxTipIzvestaja == "subsintetika" {
		qb.AddEqual("fkpl.vkonta", "2")
	}
	if cbxTipIzvestaja == "sintetika" {
		qb.AddEqual("fkpl.vkonta", "3")
	}
	if cbx_odmeseca != "" {
		qb.AddCondition("EXTRACT(MONTH FROM fpro.danal)", cbx_odmeseca, ">=")
	}
	if cbx_domeseca != "" {
		qb.AddCondition("EXTRACT(MONTH FROM fpro.danal)", cbx_domeseca, "<=")
	}

	if cbxKlasa != "" && cbxTipIzvestaja == "klasa_sifra" {
		klasaKonta := s.extractGrupa(odKonta, 2)
		qb.AddLike("fkpl.konto", klasaKonta+"%")
	}
	if chkSaldoRazl0 || chkSaldoVece0 || chkSaldoManje0 || chkSaldoJednako0 {
		if chkSaldoRazl0 {
			qb.AddCustomCondition(" (pstsaldo + saldo) <> 0 ", "AND")
		}
		if chkSaldoVece0 || chkSaldoManje0 {
			qb.AddCustomCondition(fmt.Sprintf(" (pstsaldo + saldo) > %.2f ", odSalda), "AND")
			qb.AddCustomCondition(fmt.Sprintf(" (pstsaldo + saldo) < %.2f ", doSalda), "AND")
		}
		if chkSaldoJednako0 {
			qb.AddCustomCondition(" (pstsaldo + saldo) = 0 ", "AND")
		}
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.SaldaPartnerDto{}))
		qb.AddSearchConditions(s.GetSaldaPartneriPrelomljenoTableFields(), searchText)
	}
	qb.AddGroupBy("fkpl.konto, fkpl.sifra, fkpl.naziv, fkpl.idpartneri, p.naziv, p.adresa, p.mesto, p.pib, p.jmbg, p.bpg, p.index")
	qb.AddOrderBy("fkpl.konto, fkpl.sifra")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.saldaPartneriRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		totRecords := len(*entities)
		tbl.Pagination.TotalRecords = totRecords
		tbl.Pagination.TotalPages = (totRecords + pageSize - 1) / pageSize
		if tbl.Pagination.EndRecord > totRecords {
			tbl.Pagination.EndRecord = totRecords
		}
		if tbl.Pagination.StartRecord > totRecords {
			tbl.Pagination.StartRecord = totRecords
		}
		return nil
	}

	if entities != nil {
		for i, entity := range *entities {
			// Build naziv with partner info
			if entity.IDPartneri == 0 {
				entity.Naziv = entity.NazivKonta
			} else {
				entity.Naziv = buildNaziv(entity.Naziv, entity.PIB, entity.JMBG, entity.BPG, entity.BrIndex)
			}
			fields := []string{
				"+",
				fmt.Sprintf("%d", i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.PstSaldo, 2),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Saldo, 2),
				entity.Adresa,
				entity.Mesto,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

func buildNaziv(naziv, pib, jmbg, bpg, index string) string {
	if pib == "" && jmbg == "" && bpg == "" && index == "" {
		return naziv
	}

	var sb strings.Builder
	sb.WriteString(naziv)
	sb.WriteString("\n   ")

	if pib != "" {
		sb.WriteString(" PIB:" + pib + "   ")
	}
	if jmbg != "" {
		sb.WriteString(" JMBG:" + jmbg + "   ")
	}
	if bpg != "" {
		sb.WriteString(" BPG:" + bpg + "   ")
	}
	if index != "" {
		sb.WriteString(" INDEX:" + index)
	}

	return sb.String()
}

func (s *SaldaResource) extractGrupa(konto string, nDuzSin int) string {
	if len(konto) > nDuzSin {
		return konto[:nDuzSin]
	}
	return konto
}
func (s *SaldaResource) GetSaldaTotalValues(c *gin.Context) (domain.SaldaDto, error) {
	var totals domain.SaldaDto

	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return totals, fmt.Errorf("user session not found")
	}

	// Create first query (opening balance - tipdok = '00')
	qb := common.NewQueryBuilder(`SELECT 
    COALESCE(SUM(CASE WHEN kat = '1' OR kat = '2' THEN iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN kat = '3' OR kat = '4' THEN iznos ELSE 0 END), 0) as pot
    FROM fpro f`)

	// Add conditions to first query
	hasGod, hasKAr := s.fproRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}

	vkonta := c.Query("tipkonta")
	if vkonta != "3" {
		qb.AddEqual("f.konto", c.Query("konto"))
		if c.Query("sifra") != "" {
			qb.AddEqual("f.sifra", c.Query("sifra"))
		}
	}
	// Add vkonta conditions.
	switch vkonta {
	case "1":
		qb.AddEqual("f.vkonta", vkonta)
	case "2":
		qb.AddIn("f.vkonta", []interface{}{"1", "2"})
	case "3":
		qb.AddLike("f.konto", c.Query("konto"))
	}
	qb.AddEqual("f.tipdok", "00")

	sqlQuery, args := qb.Build()

	entities, err := s.fproRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return totals, err
	}

	for _, entity := range *entities {
		totals.PocStanjeDug += entity.Dug
		totals.PocStanjePot += entity.Pot
		totals.PocStanjeSaldo += entity.Dug - entity.Pot
	}
	qb = common.NewQueryBuilder(`SELECT 
    COALESCE(SUM(CASE WHEN kat = '1' OR kat = '2' THEN iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN kat = '3' OR kat = '4' THEN iznos ELSE 0 END), 0) as pot
    FROM fpro f`)

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	if vkonta != "3" {
		qb.AddEqual("f.konto", c.Query("konto"))
		if c.Query("sifra") != "" {
			qb.AddEqual("f.sifra", c.Query("sifra"))
		}
	}
	// Add vkonta conditions.
	switch vkonta {
	case "1":
		qb.AddEqual("f.vkonta", vkonta)
	case "2":
		qb.AddIn("f.vkonta", []interface{}{"1", "2"})
	case "3":
		qb.AddLike("f.konto", c.Query("konto"))
	}
	qb.AddCondition("f.tipdok", "00", "!=")
	sqlQuery, args = qb.Build()

	entities, err = s.fproRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return totals, err
	}

	for _, entity := range *entities {
		totals.TekuciPromDug += entity.Dug
		totals.TekuciPromPot += entity.Pot
		totals.TekuciPromSaldo += entity.Dug - entity.Pot
	}
	totals.UkPromDug = totals.PocStanjeDug + totals.TekuciPromDug
	totals.UkPromPot = totals.PocStanjePot + totals.TekuciPromPot
	totals.UkPromSaldo = totals.PocStanjeSaldo + totals.TekuciPromSaldo
	return totals, nil
}
func (s *SaldaResource) SetDefaultTableData(tbl *domain.TableData) {
	tbl.TableID = "saldapojedinacnihkonta-table"
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
func (s *SaldaResource) GetSaldaPartneriList(c *gin.Context, tblPartneri *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	// Build query dynamically
	var qb *common.QueryBuilder
	// if search text is not epmty, add search conditions
	searchText := c.Query("query")
	hasGod, hasKAr := s.fkplRepo.GetHasGodHasKar()
	if getTotRecords {
		qb = common.NewQueryBuilder(`SELECT COUNT(*) as total FROM partneri `)

		// Add system conditions
		if hasGod {
			qb.AddEqual("god", userSession.SelectedGod)
		}
		if hasKAr {
			qb.AddEqual("kar", userSession.SelectedKar)
		}
		if searchText != "" {
			qb.SetEntityType(reflect.TypeOf(domain.Partneri{}))
			qb.AddSearchConditions(s.GetSaldaPartneriTableFields(), searchText)
		}
		sqlQuery, args := qb.Build()

		totRecords, err := s.partneriRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
		tblPartneri.SearchEnabled = true
		tblPartneri.ShowPagination = true
		tblPartneri.ShowActions = false
		tblPartneri.BtnAdd.IsVisible = false
		tblPartneri.BtnPrint.IsVisible = false
		tblPartneri.Pagination.CurrentPage = currentPage
		tblPartneri.Pagination.StartRecord = (currentPage-1)*pageSize + 1
		tblPartneri.Pagination.EndRecord = tblPartneri.Pagination.StartRecord + pageSize - 1
		tblPartneri.Pagination.PageSize = pageSize
		tblPartneri.Pagination.TotalRecords = totRecords
		tblPartneri.Pagination.TotalPages = (totRecords + pageSize - 1) / pageSize
		if tblPartneri.Pagination.EndRecord > totRecords {
			tblPartneri.Pagination.EndRecord = totRecords
		}
		if tblPartneri.Pagination.StartRecord > totRecords {
			tblPartneri.Pagination.StartRecord = totRecords
		}
		return err
	}

	qb = common.NewQueryBuilder(`SELECT idpartneri, sifra, naziv, pib, pobro, adresa, mesto FROM partneri `)

	// Add system conditions using userSession instead of globals
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKAr {
		qb.AddEqual("kar", userSession.SelectedKar)
	}

	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Partneri{}))
		qb.AddSearchConditions(s.GetSaldaPartneriTableFields(), searchText)
	}
	qb.AddOrderBy("sifra::NUMERIC ASC")

	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)

	sqlQuery, args := qb.Build()
	entities, err := s.partneriRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	common.SetTableRows(tblPartneri, *entities, s.GetSaldaPartneriTableFields(), "idpartneri", "", s.partneriRepo.GetFieldCache())

	return nil
}
func (s *SaldaResource) ProcessSaldaPartneriDetails(c *gin.Context, idPartneri int64, tblKonta, tblDetalji *domain.TableData) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	bBudzetsko := false

	hasGod, hasKar := s.partneriRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT budzetski FROM partneri`)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("idpartneri", idPartneri)
	sqlQuery, args := qb.Build()
	partneriEntities, err := s.partneriRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if partneriEntities != nil && len(*partneriEntities) > 0 {
		bBudzetsko = (*partneriEntities)[0].Budzetski == true //budzetski
	}

	// Get user session from context
	userSession = domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	gnGod := userSession.SelectedGod
	gnKar := userSession.SelectedKar

	// Query 1: Get aggregated Konta/Salda data per FKPL
	qbKonta := common.NewQueryBuilder(`
        SELECT 
            f.konto,
            f.sifra,
            f.naziv,
            COALESCE(SUM(CASE WHEN fp.kat = 1 OR fp.kat = 2 THEN fp.iznos ELSE 0 END), 0) as dug,
            COALESCE(SUM(CASE WHEN fp.kat = 3 OR fp.kat = 4 THEN fp.iznos ELSE 0 END), 0) as pot
        FROM fpro fp `)
	hasGod, hasKar = s.fproRepo.GetHasGodHasKar()
	qbKonta.AddJoin("LEFT JOIN fkpl f ON f.idfkpl = fp.idfkpl")
	if hasGod {
		qbKonta.AddEqual("fp.god", gnGod)
	}
	if hasKar {
		qbKonta.AddEqual("fp.kar", gnKar)
	}
	qbKonta.AddEqual("fp.vkonta", 1)
	qbKonta.AddEqual("f.idpartneri", idPartneri)
	qbKonta.AddGroupBy("f.konto, f.sifra, f.naziv")
	qbKonta.AddOrderBy("f.konto")

	sqlQueryKonta, argsKonta := qbKonta.Build()
	kontaEntities, err := s.fproRepo.GetAllCustom(c, sqlQueryKonta, "", argsKonta, "", "")
	if err != nil {
		return err
	}

	// Populate Konta table
	if kontaEntities != nil && len(*kontaEntities) > 0 {
		for _, konta := range *kontaEntities {
			fields := []string{
				fmt.Sprintf("%s %s-%s", konta.Konto, konta.Sifra, konta.Naziv),
				common.FormatNumberWithSystemLocale(konta.Dug, 2),
				common.FormatNumberWithSystemLocale(konta.Pot, 2),
				common.FormatNumberWithSystemLocale(konta.Dug-konta.Pot, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tblKonta.Rows = append(tblKonta.Rows, tblRow)
		}
	}

	// Query 2: Get all detail transactions
	qbDetalji := common.NewQueryBuilder(`
        SELECT fp.konto, fp.sifra, fp.vrd, fp.kat,
            fp.tipdok, fp.nalog, fp.danal, fp.dokum,
            fp.dadok, fp.tra, fp.iznos
        FROM fpro fp `)
	qbDetalji.AddJoin(" JOIN fkpl f ON fp.idfkpl = f.idfkpl ")
	qbDetalji.AddEqual("f.idpartneri", idPartneri)
	qbDetalji.AddEqual("fp.vkonta", 1)
	if hasGod {
		qbDetalji.AddEqual("fp.god", gnGod)
	}
	if hasKar {
		qbDetalji.AddEqual("fp.kar", gnKar)
	}
	qbDetalji.AddOrderBy("fp.danal ASC")

	sqlQueryDetalji, argsDetalji := qbDetalji.Build()
	detaljiEntities, err := s.fproRepo.GetAllCustom(c, sqlQueryDetalji, "", argsDetalji, "", "")
	if err != nil {
		return err
	}

	// Populate Detalji table with VRD mapping and TIP calculation
	if detaljiEntities != nil && len(*detaljiEntities) > 0 {
		for _, detail := range *detaljiEntities {
			// Map VRD
			var sVrd string
			switch detail.Vrd {
			case 10:
				sVrd = "10-Izdat racun"
			case 20:
				sVrd = "20-Primljen racun"
			case 30:
				sVrd = "30-Primljena uplata"
			case 40:
				sVrd = "40-Izvrsena uplata"
			case 80:
				sVrd = "80-Opsti dok."
			case 90:
				sVrd = "90-Aut. knj. dokum."
			default:
				sVrd = ""
			}

			// Determine TIP
			var sTip string
			switch detail.Vrd {
			case 10, 20:
				sTip = "F"
			case 30, 40:
				sTip = "U"
			case 80, 90:
				kontoPrefix := string(detail.Konto[0])
				if bBudzetsko {
					if kontoPrefix == "1" && (detail.Kat == 1 || detail.Kat == 2) {
						sTip = "F"
					} else if kontoPrefix == "1" && (detail.Kat == 3 || detail.Kat == 4) {
						sTip = "U"
					} else if kontoPrefix == "2" && (detail.Kat == 1 || detail.Kat == 2) {
						sTip = "U"
					} else if kontoPrefix == "2" && (detail.Kat == 3 || detail.Kat == 4) {
						sTip = "F"
					}
				} else {
					if kontoPrefix == "2" && (detail.Kat == 1 || detail.Kat == 2) {
						sTip = "F"
					} else if kontoPrefix == "2" && (detail.Kat == 3 || detail.Kat == 4) {
						sTip = "U"
					} else if kontoPrefix == "4" && (detail.Kat == 1 || detail.Kat == 2) {
						sTip = "U"
					} else if kontoPrefix == "4" && (detail.Kat == 3 || detail.Kat == 4) {
						sTip = "F"
					}
				}
			default:
				sTip = "-"
			}

			fields := []string{
				sTip,
				fmt.Sprintf("%s-%d", detail.Tipdok, detail.Nalog),
				detail.Danal.Format("02.01.2006"),
				sVrd,
				detail.Dokum,
				detail.Dadok.Time.Format("02.01.2006"),
				fmt.Sprintf("%d", detail.Tra),
				common.FormatNumberWithSystemLocale(detail.Iznos, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tblDetalji.Rows = append(tblDetalji.Rows, tblRow)
		}
	}

	return nil
}
func (s *SaldaResource) GetSaldaPartneriPrelomljeno(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	gnGod := userSession.SelectedGod
	gnKar := userSession.SelectedKar
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	odSifra := c.Query("sifra_od")
	doSifra := c.Query("sifra_do")
	searchText := c.Query("query")

	tbl.SearchEnabled = false
	tbl.ShowPagination = true
	tbl.ShowActions = false
	tbl.BtnAdd.IsVisible = false
	tbl.BtnPrint.IsVisible = false
	tbl.Pagination.CurrentPage = currentPage
	tbl.Pagination.StartRecord = (currentPage-1)*pageSize + 1
	tbl.Pagination.EndRecord = tbl.Pagination.StartRecord + pageSize - 1
	tbl.Pagination.PageSize = pageSize

	// Build partner balance query
	qb := common.NewQueryBuilder(`
        select 
            partneri.sifra,
            partneri.naziv,
            partneri.pib,
            coalesce(sum(case when fpro.konto like '200%' or fpro.konto like '201%' or fpro.konto like '202%' or 
                fpro.konto like '203%' or fpro.konto like '204%' or fpro.konto like '205%' or fpro.konto like '206%' 
                then fpro.iznos else 0 end), 0) as kupac,
            coalesce(sum(case when fpro.konto like '431%' or fpro.konto like '432%' or fpro.konto like '433%' or 
                fpro.konto like '434%' or fpro.konto like '435%' or fpro.konto like '436%' 
                then fpro.iznos else 0 end), 0) as dobavljac,
            coalesce(sum(case when fpro.konto like '430%' then fpro.iznos else 0 end), 0) as primljenavans,
            coalesce(sum(case when fpro.konto like '150%' or fpro.konto like '151%' or fpro.konto like '152%' or 
                fpro.konto like '153%' or fpro.konto like '154%' or fpro.konto like '155%' 
                then fpro.iznos else 0 end), 0) as datavans
        from fpro`)

	qb.AddJoin("inner join fkpl on fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("inner join partneri on partneri.idpartneri = fkpl.idpartneri")
	if hasGod {
		qb.AddEqual("fpro.god", gnGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", gnKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.konto ilike '15%' OR fpro.konto ilike '20%' OR fpro.konto ilike '43%')")
	if odSifra != "" {
		qb.AddCondition("fpro.sifra", odSifra, ">=")
	}
	if doSifra != "" {
		qb.AddCondition("fpro.sifra", doSifra, "<=")
	}
	// if search text is not epmty, add search conditions
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.SaldaPartnerDto{}))
		qb.AddSearchConditions(s.GetSaldaPartneriPrelomljenoTableFields(), searchText)
	}

	qb.AddGroupBy("fkpl.idpartneri, partneri.sifra, partneri.naziv, partneri.pib")
	qb.AddOrderBy("partneri.sifra::numeric asc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.saldaPartneriRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		totRecords := len(*entities)
		tbl.Pagination.TotalRecords = totRecords
		tbl.Pagination.TotalPages = (totRecords + pageSize - 1) / pageSize
		if tbl.Pagination.EndRecord > totRecords {
			tbl.Pagination.EndRecord = totRecords
		}
		if tbl.Pagination.StartRecord > totRecords {
			tbl.Pagination.StartRecord = totRecords
		}
		return nil
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			// Determine Duguje/Potrazuje
			stanje := entity.Kupac + entity.DatAvans - entity.Dobavljac - entity.PrimljenAvanas
			dugPot := "-"
			if stanje > 0 {
				dugPot = "Duguje"
			} else if stanje < 0 {
				dugPot = "Potražuje"
			}

			fields := []string{
				entity.Sifra,
				entity.Naziv,
				entity.PIB,
				common.FormatNumberWithSystemLocale(entity.Kupac, 2),
				common.FormatNumberWithSystemLocale(entity.Dobavljac, 2),
				common.FormatNumberWithSystemLocale(entity.PrimljenAvanas, 2),
				common.FormatNumberWithSystemLocale(entity.DatAvans, 2),
				common.FormatNumberWithSystemLocale(stanje, 2),
				dugPot,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}
func (s *SaldaResource) SaldaPoKomercijalistima(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	gnGod := userSession.SelectedGod
	gnKar := userSession.SelectedKar
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	podDatumom := c.Query("pod_datumom")
	odKomercijaliste := c.Query("od_komercijaliste")
	doKomercijaliste := c.Query("do_komercijaliste")
	searchText := c.Query("query")
	// Get total records if requested
	if getTotRecords {
		qbCount := common.NewQueryBuilder(`select count(distinct fpro.komid) as total from fpro`)
		qbCount.AddJoin("left join komercijalisti as kom on kom.komid = fpro.komid")

		if hasGod {
			qbCount.AddEqual("fpro.god", gnGod)
		}
		if hasKar {
			qbCount.AddEqual("fpro.kar", gnKar)
		}
		qbCount.AddEqual("fpro.vkonta", 1)
		qbCount.AddCustomCondition("(fpro.vrd = 10 OR fpro.vrd = 30)")
		qbCount.AddCondition("fpro.danal", podDatumom, "<=")
		qbCount.AddCondition("kom.sifkom", odKomercijaliste, ">=")
		qbCount.AddCondition("kom.sifkom", doKomercijaliste, "<=")
		// if search text is not empty, add search conditions for commercialist fields
		if searchText != "" {
			likePattern := fmt.Sprintf("%%%s%%", searchText)
			qbCount.AddCustomCondition("(kom.sifkom::TEXT ILIKE $1 OR kom.imeprezime::TEXT ILIKE $2)", likePattern, likePattern)
		}

		sqlQuery, args := qbCount.Build()
		totRecords, err := s.fproRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
		if err != nil {
			return err
		}

		tbl.SearchEnabled = true
		tbl.ShowPagination = true
		tbl.ShowActions = false
		tbl.BtnAdd.IsVisible = false
		tbl.BtnPrint.IsVisible = false
		tbl.Pagination.CurrentPage = currentPage
		tbl.Pagination.StartRecord = (currentPage-1)*pageSize + 1
		tbl.Pagination.EndRecord = tbl.Pagination.StartRecord + pageSize - 1
		tbl.Pagination.PageSize = pageSize
		tbl.Pagination.TotalRecords = totRecords
		tbl.Pagination.TotalPages = (totRecords + pageSize - 1) / pageSize
		if tbl.Pagination.EndRecord > totRecords {
			tbl.Pagination.EndRecord = totRecords
		}
		if tbl.Pagination.StartRecord > totRecords {
			tbl.Pagination.StartRecord = totRecords
		}
		return nil
	}

	// Build commercialist balance query
	qb := common.NewQueryBuilder(
		fmt.Sprintf(`select 
            fpro.komid,
            kom.sifkom,
            kom.imeprezime,
            coalesce(sum(case when fpro.kat in (1,2) then fpro.iznos else 0 end), 0) as dug,
    		coalesce(sum(case when fpro.kat in (3,4) then fpro.iznos else 0 end), 0) as pot,
    		coalesce(sum(case when (fpro.dadok::date + (fpro.rok || ' days')::interval)::date <= '%s'::date
           	then case when fpro.kat in (1,2) then fpro.iznos 
                    when fpro.kat in (3,4) then -fpro.iznos 
                                 else 0 end
                  else 0 end), 0) as dospelo
        from fpro`, podDatumom))

	qb.AddJoin("left join komercijalisti as kom on kom.komid = fpro.komid")
	if hasGod {
		qb.AddEqual("fpro.god", gnGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", gnKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.vrd = 10 OR fpro.vrd = 30)")
	qb.AddCondition("fpro.danal", podDatumom, "<=")
	qb.AddCondition("kom.sifkom", odKomercijaliste, ">=")
	qb.AddCondition("kom.sifkom", doKomercijaliste, "<=")
	// if search text is not empty, add search conditions for commercialist fields
	if searchText != "" {
		likePattern := fmt.Sprintf("%%%s%%", searchText)
		qb.AddCustomCondition("(kom.sifkom::TEXT ILIKE $1 OR kom.imeprezime::TEXT ILIKE $2)", likePattern, likePattern)
	}
	qb.AddGroupBy("fpro.komid, kom.sifkom, kom.imeprezime")
	qb.AddOrderBy("kom.sifkom asc")
	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)

	sqlQuery, args := qb.Build()
	entities, err := s.saldaKomRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Populate table rows and filter by sifra range
	if entities != nil && len(*entities) > 0 {
		for i, entity := range *entities {

			fields := []string{
				"+", // Detalji placeholder
				common.FormatNumberWithSystemLocale(i, 0),
				common.FormatNumberWithSystemLocale(entity.Sifkom, 0),
				entity.Imeprezime,
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Dospelo, 2),
				common.FormatNumberWithSystemLocale((entity.Duguje-entity.Potrazuje)-entity.Dospelo, 2),
				common.FormatNumberWithSystemLocale(entity.KomID, 0),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}
func (s *SaldaResource) RealizacijaKomercijalisti(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	gnGod := userSession.SelectedGod
	gnKar := userSession.SelectedKar
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	odDatuma := c.Query("od_datuma")
	doDatuma := c.Query("do_datuma")
	odKomercijaliste := c.Query("od_komercijaliste")
	doKomercijaliste := c.Query("do_komercijaliste")
	searchText := c.Query("query")
	// Get total records if requested
	if getTotRecords {
		qbCount := common.NewQueryBuilder(`select count(distinct fpro.komid) as total from fpro`)
		qbCount.AddJoin("left join komercijalisti as kom on kom.komid = fpro.komid")

		if hasGod {
			qbCount.AddEqual("fpro.god", gnGod)
		}
		if hasKar {
			qbCount.AddEqual("fpro.kar", gnKar)
		}
		qbCount.AddEqual("fpro.vkonta", 1)
		qbCount.AddCustomCondition("(fpro.vrd = 10 OR fpro.vrd = 30)")
		qbCount.AddCondition("fpro.danal", odDatuma, ">=")
		qbCount.AddCondition("fpro.danal", doDatuma, "<=")
		qbCount.AddCondition("kom.sifkom", odKomercijaliste, ">=")
		qbCount.AddCondition("kom.sifkom", doKomercijaliste, "<=")
		// if search text is not empty, add search conditions for commercialist fields
		if searchText != "" {
			likePattern := fmt.Sprintf("%%%s%%", searchText)
			qbCount.AddCustomCondition("(kom.sifkom::TEXT ILIKE $1 OR kom.imeprezime::TEXT ILIKE $2)", likePattern, likePattern)
		}

		sqlQuery, args := qbCount.Build()
		totRecords, err := s.fproRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
		if err != nil {
			return err
		}

		tbl.SearchEnabled = true
		tbl.ShowPagination = true
		tbl.ShowActions = false
		tbl.BtnAdd.IsVisible = false
		tbl.BtnPrint.IsVisible = false
		tbl.Pagination.CurrentPage = currentPage
		tbl.Pagination.StartRecord = (currentPage-1)*pageSize + 1
		tbl.Pagination.EndRecord = tbl.Pagination.StartRecord + pageSize - 1
		tbl.Pagination.PageSize = pageSize
		tbl.Pagination.TotalRecords = totRecords
		tbl.Pagination.TotalPages = (totRecords + pageSize - 1) / pageSize
		if tbl.Pagination.EndRecord > totRecords {
			tbl.Pagination.EndRecord = totRecords
		}
		if tbl.Pagination.StartRecord > totRecords {
			tbl.Pagination.StartRecord = totRecords
		}
		return nil
	}

	// Build commercialist balance query
	qb := common.NewQueryBuilder(`select 
            fpro.komid,
            kom.sifkom,
            kom.imeprezime,
            coalesce(sum(case when fpro.kat in (1,2) then fpro.iznos else 0 end), 0) as dug,
    		coalesce(sum(case when fpro.kat in (3,4) then fpro.iznos else 0 end), 0) as pot,
    		coalesce(sum(case when (fpro.dadok::date + (fpro.rok || ' days')::interval)::date <= fpro.danal::date
           	then case when fpro.kat in (3,4) then fpro.iznos 
                                 else 0 end
                  else 0 end), 0) as vanperioda
        from fpro`)

	qb.AddJoin("left join komercijalisti as kom on kom.komid = fpro.komid")
	if hasGod {
		qb.AddEqual("fpro.god", gnGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", gnKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.vrd = 10 OR fpro.vrd = 30)")
	qb.AddCondition("fpro.danal", odDatuma, ">=")
	qb.AddCondition("fpro.danal", doDatuma, "<=")
	qb.AddCondition("kom.sifkom", odKomercijaliste, ">=")
	qb.AddCondition("kom.sifkom", doKomercijaliste, "<=")
	// if search text is not empty, add search conditions for commercialist fields
	if searchText != "" {
		likePattern := fmt.Sprintf("%%%s%%", searchText)
		qb.AddCustomCondition("(kom.sifkom::TEXT ILIKE $1 OR kom.imeprezime::TEXT ILIKE $2)", likePattern, likePattern)
	}
	qb.AddGroupBy("fpro.komid, kom.sifkom, kom.imeprezime")
	qb.AddOrderBy("kom.sifkom::NUMERIC asc")
	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)

	sqlQuery, args := qb.Build()
	entities, err := s.saldaKomRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Populate table rows and filter by sifra range
	if entities != nil && len(*entities) > 0 {
		for i, entity := range *entities {

			fields := []string{
				"+", // Detalji placeholder
				common.FormatNumberWithSystemLocale(i, 0),
				common.FormatNumberWithSystemLocale(entity.Sifkom, 0),
				entity.Imeprezime,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale((entity.Duguje - entity.Potrazuje), 2),
				common.FormatNumberWithSystemLocale((entity.Vanperioda), 2),
				common.FormatNumberWithSystemLocale(entity.KomID, 0),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

func (s *SaldaResource) GetFieldCache() map[string]reflect.StructField {
	return s.service.fieldCache
}

func (s *SaldaResource) GetPojedKontaTableFields() []domain.Fields {
	return s.saldaPojedinacniTableFields
}
func (s *SaldaResource) GetGrupeKontaTableFields() []domain.Fields {
	return s.saldaGrupeKontaTableFields
}

func (s *SaldaResource) GetSaldaPartneriTableFields() []domain.Fields {
	return s.saldaPartneriTableFields
}
func (s *SaldaResource) GetSaldaPartneriHeaderTableFields() []domain.Fields {
	return s.saldaPartneriHeaderTableFields
}
func (s *SaldaResource) GetSaldaPartneriDetailTableFields() []domain.Fields {
	return s.saldaPartneriDetailTableFields
}
func (s *SaldaResource) GetSaldaPartneriPrelomljenoTableFields() []domain.Fields {
	return s.saldaPartneriPrelomljenoTableFields
}
func (s *SaldaResource) GetKomercijalistiTableFields() []domain.Fields {
	return s.saldaKomercijalistiTableFields
}
func (s *SaldaResource) GetRealizacijaKomercijalistiTableFields() []domain.Fields {
	return s.saldaRealizacijakomercijalistiTableFields
}

func (s *SaldaResource) setServiceFieldValues() {
	s.saldaPojedinacniTableFields = []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "10"},
		{Name: "duguje", Label: "Duguje", Width: "12"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12"},
		{Name: "saldo", Label: "Saldo u mesecu", Width: "12"},
		{Name: "saldokumul", Label: "Saldo na kraju meseca", Width: "15"},
	}
}

func (s *SaldaResource) setServiceSaldaPartneriFieldValues() {
	s.saldaPartneriTableFields = []domain.Fields{
		{Name: "sifra", Label: "Šifra partnera", Width: "10"},
		{Name: "naziv", Label: "Naziv partnera", Width: "60"},
		{Name: "pib", Label: "PIB", Width: "12"},
		{Name: "adresa", Label: "Adresa", Width: "30"},
		{Name: "pobro", Label: "Poštanski broj", Width: "12"},
		{Name: "mesto", Label: "Mesto", Width: "30"},
		{Name: "idpartneri", Label: "ID Partnera", Width: "30"},
	}
	s.saldaGrupeKontaTableFields = []domain.Fields{
		{Name: "detail", Label: "Detail", Width: "10"},
		{Name: "rbr", Label: "Redni Broj", Width: "10"},
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Šifra", Width: "10"},
		{Name: "naziv", Label: "Naziv konta", Width: "60"},
		{Name: "pst", Label: "PST", Width: "12", Field: "pst", SkipInSearch: true},
		{Name: "dudguje", Label: "Duguje", Width: "30", Field: "doduguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "30", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "30", Field: "saldo", SkipInSearch: true},
		{Name: "adresa", Label: "Adresa", Width: "30"},
		{Name: "mesto", Label: "Mesto", Width: "30"},
	}
	s.saldaPartneriHeaderTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "15"},
		{Name: "duguje", Label: "Duguje", Width: "15"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "15"},
		{Name: "saldo", Label: "Saldo", Width: "15"},
	}
	s.saldaPartneriDetailTableFields = []domain.Fields{
		{Name: "f_u", Label: "F/U", Width: "8"},
		{Name: "nalog", Label: "Nalog", Width: "6"},
		{Name: "danal", Label: "Datum Naloga", Width: "12"},
		{Name: "tipdok", Label: "Vrsta Dokumenta", Width: "20"},
		{Name: "brdok", Label: "Broj Dokumenta", Width: "10"},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "12"},
		{Name: "goddok", Label: "Godina Dokumenta", Width: "6"},
		{Name: "iznos", Label: "Iznos", Width: "15"},
	}
	s.saldaPartneriPrelomljenoTableFields = []domain.Fields{
		{Name: "sifra", Label: "Šifra partnera", Width: "10", Field: "partneri.sifra"},
		{Name: "naziv", Label: "Naziv partnera", Width: "60", Field: "partneri.naziv"},
		{Name: "pib", Label: "PIB", Width: "12", Field: "partneri.pib"},
		{Name: "kupac", Label: "Kupac", Width: "30", Field: "kupac", SkipInSearch: true},
		{Name: "dobavljac", Label: "Dobavljač", Width: "12", Field: "dobavljac", SkipInSearch: true},
		{Name: "primljenavans", Label: "Primljeni Avans", Width: "30", Field: "primljenavans", SkipInSearch: true},
		{Name: "datavans", Label: "Dat Avans", Width: "30", Field: "datavans", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "30", Field: "stanje", SkipInSearch: true},
		{Name: "dugpot", Label: "Duguje/Potrazuje", Width: "30", Field: "dugpot", SkipInSearch: true},
	}
	s.saldaKomercijalistiTableFields = []domain.Fields{
		{Name: "detalji", Label: "Detalji", Width: "10"},
		{Name: "rbr", Label: "Redni Broj", Width: "60"},
		{Name: "sifra", Label: "Sifra", Width: "12", Field: "komercijalisti.sifra"},
		{Name: "komercijalista", Label: "Komercijalista", Width: "30", Field: "naziv", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "saldo", SkipInSearch: true},
		{Name: "dospelo", Label: "Dospelo", Width: "30", Field: "dospelo", SkipInSearch: true},
		{Name: "nedospelo", Label: "Nedospelo", Width: "30", Field: "nedospelo", SkipInSearch: true},
	}
	s.saldaRealizacijakomercijalistiTableFields = []domain.Fields{
		{Name: "detalji", Label: "Detalji", Width: "10"},
		{Name: "rbr", Label: "Redni Broj", Width: "60"},
		{Name: "sifra", Label: "Sifra", Width: "12", Field: "komercijalisti.sifra"},
		{Name: "komercijalista", Label: "Komercijalista", Width: "30", Field: "naziv", SkipInSearch: true},
		{Name: "fakturisano", Label: "Fakturisano", Width: "12", Field: "fakturisano", SkipInSearch: true},
		{Name: "naplaceno", Label: "Naplaceno", Width: "30", Field: "naplaceno", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "30", Field: "saldo", SkipInSearch: true},
		{Name: "naplacenovanperioda", Label: "Naplaceno van perioda", Width: "30", Field: "naplacenovanperioda", SkipInSearch: true},
	}
}

func validateKontoLength(konto string, nDuzSin int) (bool, string) {
	if len(konto) < nDuzSin+1 {
		msg := "Konto mora biti bar cetvorocifren!!!"
		if nDuzSin == 4 {
			msg = "Konto mora biti bar petocifren!!!"
		}
		return false, msg
	}
	return true, ""
}

func validateSifraLength(sifra string) (bool, string) {
	if len(sifra) < 2 {
		return false, "Sifra mora biti bar dvocifrena!!!"
	}
	return true, ""
}

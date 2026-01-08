package service

import (
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
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
	GetTableStavkeFields() []domain.Fields
	GetTableNalogFields() []domain.Fields
	GetNaloziStavke(c *gin.Context, nalogID int64, searchQuery string, page int, offset int, tableFields []domain.Fields) (domain.TableData, error)
	GetFieldCache() map[string]reflect.StructField
	GetPrometAnalitickihKonta(c *gin.Context) (domain.TableData, error)
	GetPrometAnalitickihKontaMi(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetPrometDeviznihAnalitickihKonta(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetPrometSubsintetickhKonta(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetPrometSintetickihKonta(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetPrometKarticaSintetickihKonta(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetPrometAnKontaVrd(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetPrometKontaAnaliticki(c *gin.Context, searchQuery string, page int, offset int) (domain.TableData, error)
	GetTotalRecordsCustom(queryText, whereText string, args []interface{}, limitOffset, orderBy string) (int, error)
	CheckParameters(c *gin.Context) []domain.FieldError
}

// NalogResource implements the NalogService interface.
type PrometResource struct {
	service                                  *BaseService[domain.PrometDto]
	prometRepo                               *repository.BaseRepository[domain.PrometDto]
	fkplRepo                                 *repository.BaseRepository[domain.Fkpl]
	prometAnKontaTableFields                 []domain.Fields
	prometAnKontaMiTableFields               []domain.Fields
	prometAnDeviznaKontaTableFields          []domain.Fields
	prometSubsintetickihKontaTablefields     []domain.Fields
	prometSintetickihKontaTableFields        []domain.Fields
	prometKarticaSintetickihKontaTableFields []domain.Fields
	prometAnKontaVrdTableFields              []domain.Fields
	prometKontaAnalitickiTableFields         []domain.Fields
}

func NewPrometService(service *BaseService[domain.PrometDto], prometRepo *repository.BaseRepository[domain.PrometDto], fkplRepo *repository.BaseRepository[domain.Fkpl]) *PrometResource {
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

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	if konto == "" || sifra == "" || odDatuma == "" || doDatuma == "" {

		return response, fmt.Errorf("missing required parameters")
	}

	session := domain.GetSessionFromContext(c)
	if session == nil {
		return response, fmt.Errorf("user session not found")
	}

	baseArgs := []interface{}{session.SelectedGod, session.SelectedKar, konto, sifra}

	//Get totals values
	// Get "promet do" totals (up to start date)
	prometDoArgs := append(baseArgs, odDatuma) // danal < odDatuma
	prometDoQuery := `select 
        coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
        coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
        from fpro  
        where god = $1 and kar = $2 and konto = $3 and sifra = $4 
              and vkonta = 1 and danal < $5`

	prometDoResults, err := s.prometRepo.GetAllCustom(c, prometDoQuery, "", prometDoArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet do totals: %v", err)
	}

	var prometDoDuguje, prometDoPotrazuje float64
	if len(*prometDoResults) > 0 {
		// Assuming the result struct has Duguje and Potrazuje fields
		prometDoDuguje = (*prometDoResults)[0].Duguje
		prometDoPotrazuje = (*prometDoResults)[0].Potrazuje
	}

	response.Totals = domain.PrometTotalValues{
		DugDo:   prometDoDuguje,
		PotDo:   prometDoPotrazuje,
		SaldoDo: prometDoDuguje - prometDoPotrazuje,
	}

	// Get "promet za period" totals (for the specified period)
	prometPeriodArgs := append(baseArgs, odDatuma, doDatuma)
	prometPeriodQuery := `select 
        coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
        coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
        from fpro  
        where god = $1 and kar = $2 and konto = $3 and sifra = $4 
              and vkonta = 1 and danal >= $5 and danal <= $6`

	prometPeriodResults, err := s.prometRepo.GetAllCustom(c, prometPeriodQuery, "", prometPeriodArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet period totals: %v", err)
	}

	var prometPeriodDuguje, prometPeriodPotrazuje float64
	if len(*prometPeriodResults) > 0 {
		// Assuming the result struct has Duguje and Potrazuje fields
		prometPeriodDuguje = (*prometPeriodResults)[0].Duguje
		prometPeriodPotrazuje = (*prometPeriodResults)[0].Potrazuje
	}

	response.Totals.DugPer = prometPeriodDuguje
	response.Totals.PotPer = prometPeriodPotrazuje
	response.Totals.SaldoPer = prometPeriodDuguje - prometPeriodPotrazuje
	response.Totals.DugTot = response.Totals.DugDo + response.Totals.DugPer
	response.Totals.PotTot = response.Totals.PotDo + response.Totals.PotPer
	response.Totals.SaldoTot = response.Totals.SaldoDo + response.Totals.SaldoPer
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

func (s *PrometResource) GetPrometAnalitickihKonta(c *gin.Context, getTotalRecords bool, calculatedPageSize, currentPage int) (domain.PrometResponse, error) {
	var response domain.PrometResponse
	args := []interface{}{}
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return response, fmt.Errorf("user session not found")
	}

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	args = append(args, session.SelectedGod, session.SelectedKar, konto, sifra, odDatuma, doDatuma)
	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", calculatedPageSize, (currentPage-1)*calculatedPageSize)

	baseArgs := []interface{}{session.SelectedGod, session.SelectedKar, konto, sifra}
	//if we need to get only total records we chec the bool gettotalrecords
	if getTotalRecords {
		sqlQuery := `SELECT count(*)
		FROM fpro 
		WHERE 1=1 AND
		god = $1 AND kar = $2 AND konto = $3 AND sifra = $4 
		      AND vkonta = 1 AND danal >= $5 AND danal <= $6`
		totalRecords, err := s.prometRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
		response.TotalRecords = totalRecords
		return response, err
	}

	//Get totals values
	// Get "promet do" totals (up to start date)
	prometDoArgs := append(baseArgs, odDatuma) // danal < odDatuma
	prometDoQuery := `select 
        coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
        coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
        from fpro  
        where god = $1 and kar = $2 and konto = $3 and sifra = $4 
              and vkonta = 1 and danal < $5`

	prometDoResults, err := s.prometRepo.GetAllCustom(c, prometDoQuery, "", prometDoArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet do totals: %v", err)
	}

	var prometDoDuguje, prometDoPotrazuje float64
	if len(*prometDoResults) > 0 {
		// Assuming the result struct has Duguje and Potrazuje fields
		prometDoDuguje = (*prometDoResults)[0].Duguje
		prometDoPotrazuje = (*prometDoResults)[0].Potrazuje
	}

	response.Totals = domain.PrometTotalValues{
		DugDo:   prometDoDuguje,
		PotDo:   prometDoPotrazuje,
		SaldoDo: prometDoDuguje - prometDoPotrazuje,
	}

	// Get "promet za period" totals (for the specified period)
	prometPeriodArgs := append(baseArgs, odDatuma, doDatuma)
	prometPeriodQuery := `select 
        coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
        coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
        from fpro  
        where god = $1 and kar = $2 and konto = $3 and sifra = $4 
              and vkonta = 1 and danal >= $5 and danal <= $6`

	prometPeriodResults, err := s.prometRepo.GetAllCustom(c, prometPeriodQuery, "", prometPeriodArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet period totals: %v", err)
	}

	var prometPeriodDuguje, prometPeriodPotrazuje float64
	if len(*prometPeriodResults) > 0 {
		// Assuming the result struct has Duguje and Potrazuje fields
		prometPeriodDuguje = (*prometPeriodResults)[0].Duguje
		prometPeriodPotrazuje = (*prometPeriodResults)[0].Potrazuje
	}

	response.Totals.DugPer = prometPeriodDuguje
	response.Totals.PotPer = prometPeriodPotrazuje
	response.Totals.SaldoPer = prometPeriodDuguje - prometPeriodPotrazuje
	response.Totals.DugTot = response.Totals.DugDo + response.Totals.DugPer
	response.Totals.PotTot = response.Totals.PotDo + response.Totals.PotPer
	response.Totals.SaldoTot = response.Totals.SaldoDo + response.Totals.SaldoPer

	//Get data for the table
	sqlQuery := `SELECT god, kar, danal, tipdok, concat(tipdok,'-',nalog) as nalog, idfpro, kat, iznos, kolic, 
		       vrd, dokum, dadok, rok, tra, ojozn, opis, sifval, kurs, 
		       deviznos, cena, konto, idfnal, idfkpl, dokumv, dadokv, travez, rdokid,
			   CASE 
					WHEN kat = 1 OR kat = 2 THEN iznos 
					ELSE 0 
				END as duguje,
				CASE
					WHEN kat = 3 OR kat = 4 THEN iznos 
					ELSE 0 
				END as potrazuje,
				CASE 
					WHEN kat = 1 OR kat = 2 THEN kolic 
					ELSE 0 
				END as kolduguje,
				CASE
					WHEN kat = 3 OR kat = 4 THEN kolic 
					ELSE 0 
				END as kolpotrazuje
		FROM fpro 
		WHERE 1=1 AND
		god = $1 AND kar = $2 AND konto = $3 AND sifra = $4 
		      AND vkonta = 1 AND danal >= $5 AND danal <= $6
		ORDER BY god, kar, danal, tipdok, nalog, idfpro`

	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, limitOffset, "")
	if err != nil {
		return response, err
	}
	response.Data = *entities
	return response, err
}
func (s *PrometResource) GetPrometAnalitickihKontaMi(c *gin.Context, getTotalRecords bool, calculatedPageSize, currentPage int) (domain.PrometResponse, error) {
	var response domain.PrometResponse
	args := []interface{}{}
	konto := c.Query("konto")
	sifra := c.Query("sifra")
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return response, fmt.Errorf("user session not found")
	}

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	odMI := c.Query("odmi")
	doMI := c.Query("domi")

	args = append(args, session.SelectedGod, session.SelectedKar, konto, sifra, odDatuma, doDatuma, odMI, doMI)
	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", calculatedPageSize, (currentPage-1)*calculatedPageSize)

	baseArgs := []interface{}{session.SelectedGod, session.SelectedKar, konto, sifra}
	//if we need to get only total records we chec the bool gettotalrecords
	if getTotalRecords {
		sqlQuery := `SELECT count(*)
		FROM fpro 
		WHERE 1=1 AND
		god = $1 AND kar = $2 AND konto = $3 AND sifra = $4 
		      AND vkonta = 1 AND danal >= $5 AND danal <= $6 AND mi >= $7 AND mi <= $8`
		totalRecords, err := s.prometRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
		response.TotalRecords = totalRecords
		return response, err
	}

	//Get totals values
	// Get "promet do" totals (up to start date)
	prometDoArgs := append(baseArgs, odDatuma, odMI, doMI) // danal < odDatuma
	prometDoQuery := `select 
        coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
        coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
        from fpro  
        where god = $1 and kar = $2 and konto = $3 and sifra = $4 
              and vkonta = 1 and danal < $5 and mi >= $6 and mi <= $7`

	prometDoResults, err := s.prometRepo.GetAllCustom(c, prometDoQuery, "", prometDoArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet do totals: %v", err)
	}

	var prometDoDuguje, prometDoPotrazuje float64
	if len(*prometDoResults) > 0 {
		// Assuming the result struct has Duguje and Potrazuje fields
		prometDoDuguje = (*prometDoResults)[0].Duguje
		prometDoPotrazuje = (*prometDoResults)[0].Potrazuje
	}

	response.Totals = domain.PrometTotalValues{
		DugDo:   prometDoDuguje,
		PotDo:   prometDoPotrazuje,
		SaldoDo: prometDoDuguje - prometDoPotrazuje,
	}

	// Get "promet za period" totals (for the specified period)
	prometPeriodArgs := append(baseArgs, odDatuma, doDatuma, odMI, doMI)
	prometPeriodQuery := `select 
        coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
        coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
        from fpro  
        where god = $1 and kar = $2 and konto = $3 and sifra = $4 
              and vkonta = 1 and danal >= $5 and danal <= $6 and mi >= $7 and mi <= $8`

	prometPeriodResults, err := s.prometRepo.GetAllCustom(c, prometPeriodQuery, "", prometPeriodArgs, "", "")
	if err != nil {
		return response, fmt.Errorf("error getting promet period totals: %v", err)
	}

	var prometPeriodDuguje, prometPeriodPotrazuje float64
	if len(*prometPeriodResults) > 0 {
		// Assuming the result struct has Duguje and Potrazuje fields
		prometPeriodDuguje = (*prometPeriodResults)[0].Duguje
		prometPeriodPotrazuje = (*prometPeriodResults)[0].Potrazuje
	}

	response.Totals.DugPer = prometPeriodDuguje
	response.Totals.PotPer = prometPeriodPotrazuje
	response.Totals.SaldoPer = prometPeriodDuguje - prometPeriodPotrazuje
	response.Totals.DugTot = response.Totals.DugDo + response.Totals.DugPer
	response.Totals.PotTot = response.Totals.PotDo + response.Totals.PotPer
	response.Totals.SaldoTot = response.Totals.SaldoDo + response.Totals.SaldoPer

	//Get data for the table
	sqlQuery := `SELECT god, kar, danal, tipdok, concat(tipdok,'-',nalog) as nalog, idfpro, kat, iznos, kolic, 
		       vrd, dokum, dadok, rok, tra, ojozn, opis, sifval, kurs, 
		       deviznos, cena, konto, idfnal, idfkpl, dokumv, dadokv, travez, rdokid,
			   CASE 
					WHEN kat = 1 OR kat = 2 THEN iznos 
					ELSE 0 
				END as duguje,
				CASE
					WHEN kat = 3 OR kat = 4 THEN iznos 
					ELSE 0 
				END as potrazuje,
				CASE 
					WHEN kat = 1 OR kat = 2 THEN kolic 
					ELSE 0 
				END as kolduguje,
				CASE
					WHEN kat = 3 OR kat = 4 THEN kolic 
					ELSE 0 
				END as kolpotrazuje
		FROM fpro 
		WHERE 1=1 AND
		god = $1 AND kar = $2 AND konto = $3 AND sifra = $4 
		      AND vkonta = 1 AND danal >= $5 AND danal <= $6 
			  AND mi >= $7 AND mi <= $8
		ORDER BY god, kar, danal, tipdok, nalog`

	entities, err := s.prometRepo.GetAllCustom(c, sqlQuery, "", args, limitOffset, "")
	if err != nil {
		return response, err
	}
	response.Data = *entities
	return response, err
}
func (s *PrometResource) GetFieldCache() map[string]reflect.StructField {
	return s.service.fieldCache
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

func (s *PrometResource) GetKarticaSintetickihKontaTableFields() []domain.Fields {
	return s.prometKarticaSintetickihKontaTableFields
}
func (s *PrometResource) GetAnKontaVrdTableFields() []domain.Fields {
	return s.prometAnKontaVrdTableFields
}
func (s *PrometResource) GetKontaAnalitickiTableFields() []domain.Fields {
	return s.prometKontaAnalitickiTableFields
}
func (s *PrometResource) setServiceFieldValues() {
	s.prometAnKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10"},
		{Name: "danal", Label: "Datum Naloga", Width: "10"},
		{Name: "vrd", Label: "VD", Width: "4"},
		{Name: "Dokum", Label: "Dokument", Width: "6"},
		{Name: "dadok", Label: "Dat. Dokumenta", Width: "6"},
		{Name: "rok", Label: "Rok", Width: "4"},
		{Name: "tra", Label: "Godina", Width: "4"},
		{Name: "oj", Label: "OJ", Width: "4"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "duguje", Label: "Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
		{Name: "sifval", Label: "Sifval", Width: "4"},
		{Name: "kurs", Label: "Kurs", Width: "6"},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8"},
		{Name: "kolduguje", Label: "Kolic. Duguje", Width: "8"},
		{Name: "kolpotrazuje", Label: "Kolic. Potrazuje", Width: "8"},
		{Name: "cena", Label: "Cena", Width: "8"},
		{Name: "stanje", Label: "Stanje", Width: "8"},
		{Name: "dokumv", Label: "Vezn. Dokum.", Width: "6"},
		{Name: "dadokv", Label: "Dat. Vezn. Dokum.", Width: "6"},
		{Name: "travez", Label: "God. Vezn. Dokum. ", Width: "4"},
		{Name: "rdokid", Label: "Rdokid", Width: "6"},
	}

	s.prometAnKontaMiTableFields = []domain.Fields{
		{Name: "mi", Label: "Mesto Isporuke", Width: "10"},
		{Name: "nalog", Label: "Nalog", Width: "10"},
		{Name: "danal", Label: "Datum Naloga", Width: "10"},
		{Name: "vrd", Label: "VD", Width: "4"},
		{Name: "Dokum", Label: "Dokument", Width: "6"},
		{Name: "dadok", Label: "Dat. Dokumenta", Width: "6"},
		{Name: "rok", Label: "Rok", Width: "4"},
		{Name: "tra", Label: "Godina", Width: "4"},
		{Name: "oj", Label: "OJ", Width: "4"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "duguje", Label: "Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
		{Name: "sifval", Label: "Sifval", Width: "4"},
		{Name: "kurs", Label: "Kurs", Width: "6"},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8"},
		{Name: "kolduguje", Label: "Kolic. Duguje", Width: "8"},
		{Name: "kolpotrazuje", Label: "Kolic. Potrazuje", Width: "8"},
		{Name: "cena", Label: "Cena", Width: "8"},
		{Name: "stanje", Label: "Stanje", Width: "8"},
		{Name: "dokumv", Label: "Vezn. Dokum.", Width: "6"},
		{Name: "dadokv", Label: "Dat. Vezn. Dokum.", Width: "6"},
		{Name: "travez", Label: "God. Vezn. Dokum. ", Width: "4"},
		{Name: "rdokid", Label: "Rdokid", Width: "6"},
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
		{Name: "vrd", Label: "VD", Width: "4"},
		{Name: "nalog", Label: "Nalog", Width: "10"},
		{Name: "danal", Label: "Datum Naloga", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "duguje", Label: "Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
	}
	s.prometSintetickihKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10"},
		{Name: "danal", Label: "Datum Naloga", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "duguje", Label: "Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
	}
	s.prometKarticaSintetickihKontaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "duguje", Label: "Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
	}
	s.prometKontaAnalitickiTableFields = []domain.Fields{
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "30"},
		{Name: "nalog", Label: "Nalog", Width: "10"},
		{Name: "danal", Label: "Datum Naloga", Width: "10"},
		{Name: "vrd", Label: "VD", Width: "4"},
		{Name: "dokum", Label: "Dokument", Width: "6"},
		{Name: "dadok", Label: "Dat. Dokumenta", Width: "6"},
		{Name: "rok", Label: "Rok", Width: "4"},
		{Name: "tra", Label: "Godina", Width: "4"},
		{Name: "oj", Label: "OJ", Width: "4"},
		{Name: "opis", Label: "Opis", Width: "30"},
		{Name: "duguje", Label: "Duguje", Width: "8"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8"},
		{Name: "saldo", Label: "Saldo", Width: "84"},
		{Name: "sifval", Label: "Sifval", Width: "4"},
		{Name: "kurs", Label: "Kurs", Width: "6"},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8"},
		{Name: "kolduguje", Label: "Kolic. Duguje", Width: "8"},
		{Name: "kolpotrazuje", Label: "Kolic. Potrazuje", Width: "8"},
		{Name: "cena", Label: "Cena", Width: "8"},
		{Name: "stanje", Label: "Stanje", Width: "8"},
		{Name: "dokumv", Label: "Vezn. Dokum.", Width: "6"},
		{Name: "dadokv", Label: "Dat. Vezn. Dokum.", Width: "6"},
		{Name: "travez", Label: "God. Vezn. Dokum. ", Width: "4"},
		{Name: "rdokid", Label: "Rdokid", Width: "6"},
	}
}

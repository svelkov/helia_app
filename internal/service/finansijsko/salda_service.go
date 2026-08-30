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
	"strings"
)

// SaldaViewData encapsulates all data needed for the Salda display page.
type SaldaViewData struct {
	FproEntities []domain.Fpro
	TableData    domain.TableData
}

// SaldaService defines the interface for operations related to Salda.
type SaldaService interface {
	GetSaldaPojedinacnihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, page int, konto, sifra, tipkonta string) error
	GetSaldaGrupeKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, page int, params domain.SaldaParam, duzSintetika int) error
	GetSaldaPartneriList(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error
	GetSaldaTotalValues(ctx context.Context, konto, sifra, tipKonta string) (domain.SaldaDto, error)
	GetSaldaKlase5i6TotalValues(ctx context.Context, saldaParam domain.SaldaParam) (domain.SaldaDto, error)
	ProcessSaldaPartneriDetails(ctx context.Context, idPartneri int64, tblKonta, tblDetalji *domain.TableData, searchText, sortBy, sortOrder string) error
	GetSaldaPartneriPrelomljeno(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error
	GetSaldaPartneriPrelomljenoStampa(ctx context.Context, tbl *domain.TableData, sifrOd, sifraDo string) error
	GetSaldaKlase5i6Analitika(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, saldaParam domain.SaldaParam, searchText string) error
	GetSaldaKlase5i6AnalitikaStampa(ctx context.Context, tbl *domain.TableData, saldaParam domain.SaldaParam) error
	SaldaKlase5i6MT(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, saldaParam domain.SaldaParam, searchText string) error
	GetSaldaKlase5i6MTStampa(ctx context.Context, tbl *domain.TableData, saldaParam domain.SaldaParam) error
	SaldaPoKomercijalistima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error
	RealizacijaKomercijalisti(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error

	CheckSaldaGrupeParameters(ctx context.Context, requiredFields []string, params domain.SaldaParam) (fieldsError []domain.FieldError)
	CheckSaldaParameters(ctx context.Context, requiredFields []string, konto, sifra, tipkonta string) []domain.FieldError

	SetDefaultTableData(tbl *domain.TableData)
	GetFieldCache() map[string]reflect.StructField
	GetPojedKontaTableFields() []domain.Fields
	GetSaldaPartneriTableFields() []domain.Fields
	GetGrupeKontaTableFields() []domain.Fields
	GetSaldaPartneriHeaderTableFields() []domain.Fields
	GetSaldaPartneriDetailTableFields() []domain.Fields
	GetSaldaPartneriPrelomljenoTableFields() []domain.Fields
	GetSaldaPartneriPrelomljenoStampaFields() []domain.Fields
	GetSaldaKlase5i6AnalitikaTableFields() []domain.Fields
	GetSaldaKlase5i6AnalitikaStampaFields() []domain.Fields
	GetSaldaKlase5i6MTTableFields() []domain.Fields
	GetSaldaKlase5i6MTStampaFields() []domain.Fields
	GetKomercijalistiTableFields() []domain.Fields
	GetRealizacijaKomercijalistiTableFields() []domain.Fields
	GetSaldaPartneraPoKontimaStampaFields() []domain.Fields
	GetSaldaPartneraPoKontimaStampa(ctx context.Context, tbl *domain.TableData, sifrOd, sifraDo string, stampajDetalje bool) error
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// SaldaResource implements the SaldaService interface.
type SaldaResource struct {
	service                                   *service.BaseService[domain.SaldaDto]
	saldaRepo                                 *repository.BaseRepository[domain.SaldaDto]
	fkplRepo                                  *repository.BaseRepository[domain.Fkpl]
	fproRepo                                  *repository.BaseRepository[domain.Fpro]
	partneriRepo                              *repository.BaseRepository[domain.Partneri]
	saldaPartneriRepo                         *repository.BaseRepository[domain.SaldaPartnerDto]
	saldaKomRepo                              *repository.BaseRepository[domain.SaldaKomercijalistiDto]
	fvrRepo                                   *repository.BaseRepository[domain.Fvr]
	saldaPojedinacniTableFields               []domain.Fields
	saldaGrupeKontaTableFields                []domain.Fields
	saldaPartneriTableFields                  []domain.Fields
	saldaPartneriHeaderTableFields            []domain.Fields
	saldaPartneriDetailTableFields            []domain.Fields
	saldaPartneriPrelomljenoTableFields       []domain.Fields
	saldaPartneriPrelomljenoStampaTableFields []domain.Fields
	saldaKlase5i6AnalitikaTableFields         []domain.Fields
	saldaKlase5i6AnalitikaStampaTableFields   []domain.Fields
	saldaKlase5i6MTTableFields                []domain.Fields
	saldaKlase5i6MTStampaTableFields          []domain.Fields
	saldaKomercijalistiTableFields            []domain.Fields
	saldaRealizacijakomercijalistiTableFields []domain.Fields
	saldaPartneraPoKontimaStampaTableFields   []domain.Fields
}

func NewSaldaService(service *service.BaseService[domain.SaldaDto],
	saldaRepo *repository.BaseRepository[domain.SaldaDto],
	fkplRepo *repository.BaseRepository[domain.Fkpl],
	fproRepo *repository.BaseRepository[domain.Fpro],
	partneriRepo *repository.BaseRepository[domain.Partneri],
	saldaPartneriRepo *repository.BaseRepository[domain.SaldaPartnerDto],
	saldaKomRepo *repository.BaseRepository[domain.SaldaKomercijalistiDto],
	fvrRepo *repository.BaseRepository[domain.Fvr]) *SaldaResource {
	rs := &SaldaResource{
		service:           service,
		saldaRepo:         saldaRepo,
		fkplRepo:          fkplRepo,
		fproRepo:          fproRepo,
		partneriRepo:      partneriRepo,
		saldaPartneriRepo: saldaPartneriRepo,
		saldaKomRepo:      saldaKomRepo,
		fvrRepo:           fvrRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *SaldaResource) CheckSaldaParameters(ctx context.Context, requiredFields []string, konto, sifra, tipkonta string) (fieldsError []domain.FieldError) {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "user session not found"}}
	}

	// Validate required fields are not empty
	for _, field := range requiredFields {
		switch field {
		case "konto":
			if konto == "" {
				fieldsError = append(fieldsError, domain.FieldError{Field: "konto", ErrorMessage: "obavezan podatak..."})
			}
		case "tipkonta":
			if tipkonta == "" {
				fieldsError = append(fieldsError, domain.FieldError{Field: "tipkonta", ErrorMessage: "obavezan podatak..."})
			}
		}
	}

	if len(fieldsError) > 0 {
		return
	}
	// Build query dynamically
	qb := common.NewQueryBuilder(`SELECT f.konto, f.sifra FROM fkpl as f`, true)

	// Add system conditions
	hasGod, hasKAr := s.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}

	// Add user conditions
	qb.AddEqual("f.konto", konto)
	if sifra != "" {
		qb.AddEqual("f.sifra", sifra)
	}
	qb.AddEqual("f.vkonta", tipkonta)

	sqlQuery, args := qb.Build()

	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetData}}
	}
	if len(*entities) == 0 {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetKontoSifra}}
	}

	return nil
}

func (s *SaldaResource) CheckSaldaGrupeParameters(ctx context.Context, requiredFields []string, params domain.SaldaParam) (fieldsError []domain.FieldError) {
	fieldsError = []domain.FieldError{}

	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "user session not found"}}
	}

	nduzsin := 3 // Default length of synthetic account TODO should be get from config
	if params.OdKonta == "" {
		fieldsError = append(fieldsError, domain.FieldError{Field: "odkonta", ErrorMessage: "obavezan podatak..."})
	}
	if params.DoKonta == "" {
		fieldsError = append(fieldsError, domain.FieldError{Field: "dokonta", ErrorMessage: "obavezan podatak..."})
	}
	if params.OdSifre == "" {
		fieldsError = append(fieldsError, domain.FieldError{Field: "odsifre", ErrorMessage: "obavezan podatak..."})
	}
	if params.DoSifre == "" {
		fieldsError = append(fieldsError, domain.FieldError{Field: "dosifre", ErrorMessage: "obavezan podatak..."})
	}

	switch params.CbxTipIzvestaja {
	case "analitika":
		fieldsError = append(fieldsError, validateKontoFields(params.OdKonta, params.DoKonta, nduzsin, "<")...)
		fieldsError = append(fieldsError, validateSifraFields(params.OdSifre, params.DoSifre)...)

	case "subsintetika", "klasa_subsintetika":
		fieldsError = append(fieldsError, validateKontoFields(params.OdKonta, params.DoKonta, nduzsin, "<")...)
	case "sintetika":
		fieldsError = append(fieldsError, validateKontoFields(params.OdKonta, params.DoKonta, nduzsin, "=")...)

	case "klasa_sifra":
		fieldsError = append(fieldsError, validateSifraFields(params.OdSifre, params.DoSifre)...)
		if len(params.CbxKlasa) <= 0 {
			fieldsError = append(fieldsError, domain.FieldError{Field: "cbx_klasa", ErrorMessage: "Morate izabrati klasu konta!!!"})
		}
	}

	return fieldsError
}
func (s *SaldaResource) GetSaldaPojedinacnihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, page int, konto, sifra, tipkonta string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	// Create first query (opening balance - tipdok = '00')
	qb1 := common.NewQueryBuilder(`SELECT 0 as mesec,
    COALESCE(SUM(CASE WHEN kat = '1' OR kat = '2' THEN iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN kat = '3' OR kat = '4' THEN iznos ELSE 0 END), 0) as pot
    FROM fpro f`, true)

	// Add conditions to first query
	hasGod, hasKAr := s.fproRepo.GetHasGodHasKar()
	if hasGod {
		qb1.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb1.AddEqual("f.kar", userSession.SelectedKar)
	}
	if tipkonta != "3" {
		qb1.AddEqual("f.konto", konto)
		if sifra != "" {
			qb1.AddEqual("f.sifra", sifra)
		}
	}
	qb1.AddEqual("f.tipdok", "00")

	// Add tipkonta conditions.
	switch tipkonta {
	case "1":
		qb1.AddEqual("f.vkonta", tipkonta)
	case "2":
		qb1.AddIn("f.vkonta", []interface{}{"1", "2"})
	case "3":
		qb1.AddLike("f.konto", konto)
	}

	qb1.AddGroupBy("EXTRACT(MONTH FROM danal)")
	qb1.AddOrderBy("mesec ASC")
	// Create second query (monthly data - tipdok != '00')
	qb2 := common.NewQueryBuilder(`SELECT EXTRACT(MONTH FROM danal) as mesec,
    COALESCE(SUM(CASE WHEN kat = '1' OR kat = '2' THEN iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN kat = '3' OR kat = '4' THEN iznos ELSE 0 END), 0) as pot
    FROM fpro f`, true)

	// Add same base conditions
	if hasGod {
		qb2.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKAr {
		qb2.AddEqual("f.kar", userSession.SelectedKar)
	}
	if tipkonta != "3" {
		qb2.AddEqual("f.konto", konto)
		if sifra != "" {
			qb2.AddEqual("f.sifra", sifra)
		}
	}
	// Add tipkonta conditions.
	switch tipkonta {
	case "1":
		qb2.AddEqual("f.vkonta", tipkonta)
	case "2":
		qb2.AddIn("f.vkonta", []interface{}{"1", "2"})
	case "3":
		qb2.AddLike("f.konto", konto)
	}

	qb2.AddCondition("f.tipdok", "00", "!=")
	qb2.AddGroupBy("EXTRACT(MONTH FROM danal)")
	// Create UNION
	uqb := common.NewUnionQueryBuilder("UNION ALL")
	uqb.AddQuery(qb1)
	uqb.AddQuery(qb2)
	uqb.AddOrderBy("mesec")

	sqlQuery, args := uqb.Build()

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
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
					Duguje:    entity.Dug.Float64,
					Potrazuje: entity.Pot.Float64,
					Saldo:     entity.Dug.Float64 - entity.Pot.Float64,
				}
				kumulSaldo = kumulSaldo + (entity.Dug.Float64 - entity.Pot.Float64)
				templateData[entity.Mesec] = saldoDto
				templateData[entity.Mesec].SaldoKumul = kumulSaldo
			}
		}
	}
	tbl.Rows = saldaDtoToTableRows(templateData)
	return nil
}
func (s *SaldaResource) GetSaldaGrupeKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, params domain.SaldaParam, duzSintetika int) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	gnGod := userSession.SelectedGod
	gnKar := userSession.SelectedKar
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Parse filter parameters
	const saldoExpr = "(SUM(CASE WHEN fpro.kat IN ('1','2') THEN fpro.iznos ELSE 0 END) - SUM(CASE WHEN fpro.kat IN ('3','4') THEN fpro.iznos ELSE 0 END))"

	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Build combined query with partner join (eliminates N+1)
	selectFields := ``
	if params.CbxTipIzvestaja == "analitika" {
		selectFields = `fpro.konto,
			COALESCE(fpro.sifra, '') as sifra, 
			COALESCE(MIN(p.naziv), '') as kontonaziv,
			COALESCE(MIN(p.adresa), '') as adresa,
			COALESCE(MIN(p.mesto), '') as mesto,
			COALESCE(MIN(p.pib), '') as pib,`
	}
	if params.CbxTipIzvestaja == "subsintetika" {
		selectFields = `fpro.konto,
		'' as sifra,
		COALESCE(MIN(grp_fkpl.naziv), '') as kontonaziv,
		'' as adresa,
		'' as mesto,
		'' as pib,`
	}
	if params.CbxTipIzvestaja == "sintetika" {
		selectFields = fmt.Sprintf(` left(fpro.konto, %d) as konto,
		'' as sifra,
		COALESCE(MIN(grp_fkpl.naziv), '') as kontonaziv,
		'' as adresa,
		'' as mesto,
		'' as pib,`, duzSintetika)
	}

	qb := common.NewQueryBuilder(fmt.Sprintf(`
		SELECT
			%s
			COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) as pstdug,
			COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pstpot,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) as dug,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pot
		FROM fpro`, selectFields), true)

	if params.CbxTipIzvestaja == "analitika" {
		qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl")
		qb.AddJoin("LEFT JOIN tipanalitike ta ON ta.tipanalitikeid = fkpl.tipanalitikeid")
		qb.AddJoin("LEFT JOIN partneri p ON p.tipanalitikeid = ta.tipanalitikeid AND p.sifra = fpro.sifra")
	}
	if params.CbxTipIzvestaja == "subsintetika" {
		joinStr := "LEFT JOIN fkpl grp_fkpl ON grp_fkpl.konto = fpro.konto AND grp_fkpl.vkonta = 2"
		if hasGod {
			joinStr += fmt.Sprintf(" AND grp_fkpl.god = %d", gnGod)
		}
		if hasKar {
			joinStr += fmt.Sprintf(" AND grp_fkpl.kar = %d", gnKar)
		}
		qb.AddJoin(joinStr)
	}
	if params.CbxTipIzvestaja == "sintetika" {
		joinStr := fmt.Sprintf("LEFT JOIN fkpl grp_fkpl ON grp_fkpl.konto = left(fpro.konto, %d) AND grp_fkpl.vkonta = 3", duzSintetika)
		if hasGod {
			joinStr += fmt.Sprintf(" AND grp_fkpl.god = %d", gnGod)
		}
		if hasKar {
			joinStr += fmt.Sprintf(" AND grp_fkpl.kar = %d", gnKar)
		}
		qb.AddJoin(joinStr)
	}
	if hasGod {
		qb.AddEqual("fpro.god", gnGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", gnKar)
	}
	if params.CbxTipIzvestaja == "analitika" {
		qb.AddCondition("fpro.konto::numeric", params.OdKonta, ">=")
		qb.AddCondition("fpro.konto::numeric", params.DoKonta, "<=")
		qb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.OdSifre, ">=")
		qb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.DoSifre, "<=")
	}
	if params.CbxTipIzvestaja == "subsintetika" {
		qb.AddCondition("fpro.konto::numeric", params.OdKonta, ">=")
		qb.AddCondition("fpro.konto::numeric", params.DoKonta, "<=")
	}
	if params.CbxTipIzvestaja == "sintetika" {
		qb.AddCondition(fmt.Sprintf("left(fpro.konto, %d)::numeric", duzSintetika), params.OdKonta, ">=")
		qb.AddCondition(fmt.Sprintf("left(fpro.konto, %d)::numeric", duzSintetika), params.DoKonta, "<=")
	}
	qb.AddCondition("EXTRACT(MONTH FROM fpro.danal)", params.CbxOdMeseca, ">=")
	qb.AddCondition("EXTRACT(MONTH FROM fpro.danal)", params.CbxDoMeseca, "<=")

	if params.CbxTipIzvestaja == "klasa_sifra" {
		klasaKonta := s.extractGrupa(params.OdKonta, 2)
		qb.AddLike("fpro.konto", klasaKonta+"%")
	}

	switch params.CbxTipIzvestaja {
	case "analitika":
		qb.AddGroupBy("fpro.konto, fpro.sifra")
		qb.AddOrderBy("fpro.konto::numeric, COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)")
	case "subsintetika":
		qb.AddGroupBy("fpro.konto")
		qb.AddOrderBy("fpro.konto::numeric")
	case "sintetika":
		qb.AddGroupBy(fmt.Sprintf("left(fpro.konto, %d)", duzSintetika))
		qb.AddOrderBy(fmt.Sprintf("left(fpro.konto, %d)::numeric", duzSintetika))
	}
	switch params.SaldoFilter {
	case "razl_nula":
		qb.AddHaving(saldoExpr + " <> 0")
	case "vece_nula":
		qb.AddHaving(saldoExpr + " > 0")
	case "manje_nula":
		qb.AddHaving(saldoExpr + " < 0")
	case "nula":
		qb.AddHaving(saldoExpr + " = 0")
	}
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.saldaPartneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	if entities != nil {
		for i, entity := range *entities {
			// Build naziv with partner info
			if params.CbxTipIzvestaja == "analitika" {
				entity.Naziv = buildNaziv(entity.NazivKonta, entity.PIB, entity.JMBG, entity.BPG, entity.BrIndex)
			} else {
				entity.Naziv = entity.NazivKonta
			}
			fields := []string{
				"+",
				fmt.Sprintf("%d", i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.PstDug-entity.PstPot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.PstDug-entity.PstPot+entity.Dug-entity.Pot, 2),
				entity.Adresa,
				entity.Mesto,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}
func (s *SaldaResource) GetSaldaPartneriList(ctx context.Context, tblPartneri *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tblPartneri, currentPage, pageSize)
	// Add system conditions using userSession instead of globals
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT p.idpartneri, p.sifra, p.naziv, p.pib, p.pobro, p.adresa, p.mesto FROM partneri p`, true)
	qb.AddCustomCondition("EXISTS (SELECT 1 FROM fkpl f INNER JOIN fpro fp ON fp.idfkpl = f.idfkpl WHERE f.idpartneri = p.idpartneri)")
	// Add system conditions
	if hasGod {
		qb.AddEqual("p.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("p.kar", userSession.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Partneri{}))
		qb.AddSearchConditions(s.GetSaldaPartneriTableFields(), searchText)
	}
	qb.AddOrderBy("p.sifra::NUMERIC ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.partneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tblPartneri, len(*entities), pageSize)
		return nil
	}
	common.SetTableRows(tblPartneri, *entities, s.GetSaldaPartneriTableFields(), "idpartneri", "", s.partneriRepo.GetFieldCache())
	return nil
}
func (s *SaldaResource) GetSaldaPartneriPrelomljeno(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, searchText, sortBy, sortOrder string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

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
        from fpro`, true)

	qb.AddJoin("inner join fkpl on fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("inner join partneri on partneri.idpartneri = fkpl.idpartneri")
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.konto ilike '15%' OR fpro.konto ilike '20%' OR fpro.konto ilike '43%')")
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
	entities, err := s.saldaPartneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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

func (s *SaldaResource) GetSaldaPartneriPrelomljenoStampa(ctx context.Context, tbl *domain.TableData, sifrOd, sifraDo string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
        select
            partneri.sifra,
            partneri.naziv,
            partneri.adresa,
            partneri.pobro,
            partneri.mesto,
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
        from fpro`, true)
	qb.AddJoin("inner join fkpl on fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("inner join partneri on partneri.idpartneri = fkpl.idpartneri")
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.konto ilike '15%' OR fpro.konto ilike '20%' OR fpro.konto ilike '43%')")
	qb.AddCondition("partneri.sifra", sifrOd, ">=")
	qb.AddCondition("partneri.sifra", sifraDo, "<=")
	qb.AddGroupBy("fkpl.idpartneri, partneri.sifra, partneri.naziv, partneri.adresa, partneri.pobro, partneri.mesto, partneri.pib")
	qb.AddOrderBy("partneri.sifra::numeric asc")

	sqlQuery, args := qb.Build()
	entities, err := s.saldaPartneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
	tbl.HasTotals = true

	var totKupac, totDobavljac, totPrimljenAvans, totDatAvans, totStanje float64
	if entities != nil {
		for _, entity := range *entities {
			stanje := entity.Kupac + entity.DatAvans - entity.Dobavljac - entity.PrimljenAvanas
			dugPot := "-"
			if stanje > 0 {
				dugPot = "Duguje"
			} else if stanje < 0 {
				dugPot = "Potražuje"
			}
			totKupac += entity.Kupac
			totDobavljac += entity.Dobavljac
			totPrimljenAvans += entity.PrimljenAvanas
			totDatAvans += entity.DatAvans
			totStanje += stanje
			fields := []string{
				entity.Sifra,
				entity.Naziv,
				entity.Adresa,
				fmt.Sprintf("%d", entity.PostanskiBroj),
				entity.Mesto,
				entity.PIB,
				common.FormatNumberWithSystemLocale(entity.Kupac, 2),
				common.FormatNumberWithSystemLocale(entity.Dobavljac, 2),
				common.FormatNumberWithSystemLocale(entity.PrimljenAvanas, 2),
				common.FormatNumberWithSystemLocale(entity.DatAvans, 2),
				common.FormatNumberWithSystemLocale(stanje, 2),
				dugPot,
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: fields})
		}
	}

	tbl.Totals[6] = common.FormatNumberWithSystemLocale(totKupac, 2)
	tbl.Totals[7] = common.FormatNumberWithSystemLocale(totDobavljac, 2)
	tbl.Totals[8] = common.FormatNumberWithSystemLocale(totPrimljenAvans, 2)
	tbl.Totals[9] = common.FormatNumberWithSystemLocale(totDatAvans, 2)
	tbl.Totals[10] = common.FormatNumberWithSystemLocale(totStanje, 2)

	return nil
}
func (s *SaldaResource) GetSaldaPartneraPoKontimaStampa(ctx context.Context, tbl *domain.TableData, sifrOd, sifraDo string, stampajDetalje bool) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	translator := i18n.GetInstance()

	// Query 1: konta aggregates per partner per konto
	qb := common.NewQueryBuilder(`
		SELECT
			p.idpartneri,
			p.sifra as sifra,
			p.naziv as naziv,
			COALESCE(p.adresa, '') as adresa,
			p.pobro,
			COALESCE(p.pib, '') as pib,
			COALESCE(p.mesto, '') as mesto,
			fkpl.konto,
			fkpl.naziv as kontonaziv,
			COALESCE(SUM(CASE WHEN fpro.kat IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as dug,
			COALESCE(SUM(CASE WHEN fpro.kat IN (3,4) THEN fpro.iznos ELSE 0 END), 0) as pot
		FROM fpro`, true)
	qb.AddJoin("JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("JOIN partneri p ON p.idpartneri = fkpl.idpartneri")
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCondition("p.sifra", sifrOd, ">=")
	qb.AddCondition("p.sifra", sifraDo, "<=")
	qb.AddGroupBy("p.idpartneri, p.sifra, p.naziv, p.adresa, p.pobro, p.pib, p.mesto, fkpl.konto, fkpl.naziv")
	qb.AddOrderBy("p.sifra::numeric, fkpl.konto")

	sqlQuery, args := qb.Build()
	kontaEntities, err := s.saldaPartneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Query 2: details (if stampajDetalje)
	type detailKey struct {
		idPartneri int64
		konto      string
	}
	detaljiMap := map[detailKey][]domain.Fpro{}
	if stampajDetalje && kontaEntities != nil && len(*kontaEntities) > 0 {
		qbDet := common.NewQueryBuilder(`
			SELECT
				fkpl.idpartneri as idfkpl,
				fp.konto, fp.vrd, fp.kat, fp.tipdok, fp.nalog, fp.danal, fp.dokum, fp.dadok, fp.tra, fp.iznos, fp.god
			FROM fpro fp`, true)
		qbDet.AddJoin("JOIN fkpl ON fkpl.idfkpl = fp.idfkpl")
		qbDet.AddJoin("JOIN partneri p ON p.idpartneri = fkpl.idpartneri")
		if hasGod {
			qbDet.AddEqual("fp.god", userSession.SelectedGod)
		}
		if hasKar {
			qbDet.AddEqual("fp.kar", userSession.SelectedKar)
		}
		qbDet.AddEqual("fp.vkonta", 1)
		qbDet.AddCondition("p.sifra", sifrOd, ">=")
		qbDet.AddCondition("p.sifra", sifraDo, "<=")
		qbDet.AddOrderBy("p.sifra::numeric, fp.konto, fp.danal")

		sqlDet, argsDet := qbDet.Build()
		detEntities, derr := s.fproRepo.GetAllCustom(ctx, sqlDet, "", argsDet, "", "")
		if derr != nil {
			return derr
		}
		if detEntities != nil {
			for _, d := range *detEntities {
				k := detailKey{idPartneri: d.IDFkpl, konto: d.Konto}
				detaljiMap[k] = append(detaljiMap[k], d)
			}
		}
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.HasTotals = true
	var grandDug, grandPot float64
	lastPartnerID := int64(-1)
	var partnerDug, partnerPot float64
	var lastPartnerNaziv string

	if kontaEntities == nil {
		return nil
	}

	for _, entity := range *kontaEntities {
		if entity.IDPartneri != lastPartnerID {
			// Close previous partner
			if lastPartnerID != -1 {
				pSaldo := partnerDug - partnerPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "partner-total",
					Fields: []string{
						translator.Label("Ukupno za partnera") + ": " + lastPartnerNaziv,
						common.FormatNumberWithSystemLocale(partnerDug, 2),
						common.FormatNumberWithSystemLocale(partnerPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			// Open new partner
			partnerDug, partnerPot = 0, 0
			lastPartnerID = entity.IDPartneri
			lastPartnerNaziv = entity.Naziv
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "partner-header",
				Fields: []string{
					entity.Sifra + " - " + entity.Naziv,
					entity.Adresa,
					fmt.Sprintf("%d", entity.PostanskiBroj),
					entity.Mesto,
					entity.PIB,
				},
			})
		}

		dug := entity.Dug
		pot := entity.Pot
		saldo := dug - pot

		// Konto header
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-header",
			Fields: []string{
				entity.Konto,
				entity.NazivKonta,
			},
		})

		// Detail rows (only if stampajDetalje)
		if stampajDetalje {
			k := detailKey{idPartneri: entity.IDPartneri, konto: entity.Konto}
			for _, detail := range detaljiMap[k] {
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
				var sTip string
				switch detail.Vrd {
				case 10, 20:
					sTip = "F"
				case 30, 40:
					sTip = "U"
				case 80, 90:
					if len(detail.Konto) > 0 {
						kontoPrefix := string(detail.Konto[0])
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
				danalStr := ""
				if detail.Danal != nil {
					danalStr = detail.Danal.Format(common.DateLayout)
				}
				dadokStr := ""
				if detail.Dadok.Valid {
					dadokStr = detail.Dadok.Time.Format(common.DateLayout)
				}
				godStr := fmt.Sprintf("%d", detail.God)
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					Fields: []string{
						sTip,
						fmt.Sprintf("%s-%d", detail.Tipdok, detail.Nalog),
						danalStr,
						sVrd,
						detail.Dokum,
						dadokStr,
						godStr,
						common.FormatNumberWithSystemLocale(detail.Iznos, 2),
					},
				})
			}
		}

		// Konto total
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za konto") + ": " + entity.Konto + " - " + entity.NazivKonta,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(saldo, 2),
			},
		})
		partnerDug += dug
		partnerPot += pot
		grandDug += dug
		grandPot += pot
	}

	// Close last partner
	if lastPartnerID != -1 {
		pSaldo := partnerDug - partnerPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "partner-total",
			Fields: []string{
				translator.Label("Ukupno za partnera") + ": " + lastPartnerNaziv,
				common.FormatNumberWithSystemLocale(partnerDug, 2),
				common.FormatNumberWithSystemLocale(partnerPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	tbl.Totals[0] = translator.Label("Ukupno za izveštaj")
	tbl.Totals[5] = common.FormatNumberWithSystemLocale(grandDug, 2)
	tbl.Totals[6] = common.FormatNumberWithSystemLocale(grandPot, 2)
	tbl.Totals[7] = common.FormatNumberWithSystemLocale(grandSaldo, 2)

	return nil
}

func (s *SaldaResource) GetSaldaKlase5i6Analitika(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, saldaParam domain.SaldaParam, searchText string) error {
	common.SetupTablePagination(tbl, currentPage, pageSize)
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Create first query (opening balance - tipdok = '00')
	qb := common.NewQueryBuilder(`SELECT f.konto, coalesce(f.sifra, '') as sifra, fkpl.naziv,
    COALESCE(SUM(CASE WHEN f.kat = '1' OR f.kat = '2' THEN f.iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN f.kat = '3' OR f.kat = '4' THEN f.iznos ELSE 0 END), 0) as pot
    FROM fpro f`, true)
	qb.AddJoin(" left join fkpl on fkpl.idfkpl = f.idfkpl")
	// Add conditions to first query
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.OdSifre, ">=")
	qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.DoSifre, "<=")
	qb.AddCondition("f.danal::date", saldaParam.OdDatuma, ">=")
	qb.AddCondition("f.danal::date", saldaParam.DoDatuma, "<=")
	// Check account range between 5 and 6 (class 5 and 6)
	switch saldaParam.Klasa {
	case "5":
		qb.AddCustomCondition("f.konto like '5%'")
	case "6":
		qb.AddCustomCondition("f.konto like '6%'")
	default:
		qb.AddCustomCondition("(f.konto like '5%' OR f.konto like '6%')")
	}

	qb.AddGroupBy("f.konto, coalesce(f.sifra, ''), fkpl.naziv")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	if entities != nil {
		for _, entity := range *entities {
			// Monthly data
			fields := []string{
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil

}

func (s *SaldaResource) GetSaldaKlase5i6AnalitikaStampa(ctx context.Context, tbl *domain.TableData, saldaParam domain.SaldaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		SELECT f.konto, COALESCE(f.sifra, '') as sifra, fkpl.naziv,
			COALESCE(oj.naziv, '') as ojozn,
			COALESCE(m.opis, '') as mtroska,
			COALESCE(SUM(CASE WHEN f.kat IN (1,2) THEN f.iznos ELSE 0 END), 0) as dug,
			COALESCE(SUM(CASE WHEN f.kat IN (3,4) THEN f.iznos ELSE 0 END), 0) as pot
		FROM fpro f`, true)
	qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = f.idfkpl")
	qb.AddJoin("LEFT JOIN orgjed oj ON oj.idorgjed = f.idorgjed")
	qb.AddJoin("LEFT JOIN mestotr m ON m.mestotrid = f.mestotrid")

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	switch saldaParam.Klasa {
	case "5":
		qb.AddCustomCondition("f.konto like '5%'")
	case "6":
		qb.AddCustomCondition("f.konto like '6%'")
	default:
		qb.AddCustomCondition("(f.konto like '5%' OR f.konto like '6%')")
	}

	qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.OdSifre, ">=")
	qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.DoSifre, "<=")
	qb.AddCondition("f.danal::date", saldaParam.OdDatuma, ">=")
	qb.AddCondition("f.danal::date", saldaParam.DoDatuma, "<=")

	qb.AddGroupBy("f.konto, COALESCE(f.sifra, ''), fkpl.naziv, COALESCE(oj.naziv, ''), COALESCE(m.opis, '')")
	qb.AddOrderBy("COALESCE(f.sifra, ''), COALESCE(oj.naziv, ''), f.konto")

	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	translator := i18n.GetInstance()
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = translator.Label("Ukupno")
	tbl.HasTotals = true

	var grandDug, grandPot float64
	lastSifra := "\x01"
	lastOj := "\x01"
	var ojDug, ojPot float64
	var sifraDug, sifraPot float64

	emitOjSubtotal := func() {
		ojSaldo := ojDug - ojPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "group-total",
			Fields: []string{
				translator.Label("Ukupno za OJ") + ": " + lastOj, "", "",
				common.FormatNumberWithSystemLocale(ojDug, 2),
				common.FormatNumberWithSystemLocale(ojPot, 2),
				common.FormatNumberWithSystemLocale(ojSaldo, 2),
			},
		})
		ojDug, ojPot = 0, 0
	}

	emitSifraSubtotal := func() {
		sifraSaldo := sifraDug - sifraPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "sifra-total",
			Fields: []string{
				translator.Label("Ukupno za šifru") + " " + lastSifra + ":", "", "",
				common.FormatNumberWithSystemLocale(sifraDug, 2),
				common.FormatNumberWithSystemLocale(sifraPot, 2),
				common.FormatNumberWithSystemLocale(sifraSaldo, 2),
			},
		})
		sifraDug, sifraPot = 0, 0
	}

	if entities != nil {
		for _, entity := range *entities {
			sifra := entity.Sifra
			ojozn := entity.Ojozn.String

			if sifra != lastSifra {
				// New sifra group — close previous OJ and sifra subtotals
				if lastSifra != "\x01" {
					emitOjSubtotal()
					emitSifraSubtotal()
				}
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "group-header",
					Fields:   []string{sifra, "", "", "", "", ""},
				})
				lastSifra = sifra
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "subgroup-header",
					Fields:   []string{ojozn, "", "", "", "", ""},
				})
				lastOj = ojozn
			} else if ojozn != lastOj {
				// Same sifra, new OJ
				emitOjSubtotal()
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "subgroup-header",
					Fields:   []string{ojozn, "", "", "", "", ""},
				})
				lastOj = ojozn
			}

			dug := entity.Dug.Float64
			pot := entity.Pot.Float64
			saldo := dug - pot
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					entity.Konto,
					entity.Naziv,
					entity.Mtroska.String,
					common.FormatNumberWithSystemLocale(dug, 2),
					common.FormatNumberWithSystemLocale(pot, 2),
					common.FormatNumberWithSystemLocale(saldo, 2),
				},
			})
			ojDug += dug
			ojPot += pot
			sifraDug += dug
			sifraPot += pot
			grandDug += dug
			grandPot += pot
		}
		if lastSifra != "\x01" {
			emitOjSubtotal()
			emitSifraSubtotal()
		}
	}

	grandSaldo := grandDug - grandPot
	tbl.Totals[3] = common.FormatNumberWithSystemLocale(grandDug, 2)
	tbl.Totals[4] = common.FormatNumberWithSystemLocale(grandPot, 2)
	tbl.Totals[5] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
	return nil
}

func (s *SaldaResource) SaldaKlase5i6MT(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, saldaParam domain.SaldaParam, searchText string) error {
	common.SetupTablePagination(tbl, currentPage, pageSize)
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Create first query (opening balance - tipdok = '00')
	qb := common.NewQueryBuilder(`SELECT f.konto, fkpl.naziv, max(coalesce(m.opis, '')) as mtroska,
    COALESCE(SUM(CASE WHEN f.kat = '1' OR f.kat = '2' THEN f.iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN f.kat = '3' OR f.kat = '4' THEN f.iznos ELSE 0 END), 0) as pot
    FROM fpro f`, true)
	qb.AddJoin(" left join fkpl on fkpl.idfkpl = f.idfkpl")
	qb.AddJoin(" left join mestotr m on m.mestotrid = f.mestotrid")
	// Add conditions to first query
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.OdSifre, ">=")
	qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.DoSifre, "<=")
	qb.AddCondition("f.danal::date", saldaParam.OdDatuma, ">=")
	qb.AddCondition("f.danal::date", saldaParam.DoDatuma, "<=")
	// Check account range between 5 and 6 (class 5 and 6)
	switch saldaParam.Klasa {
	case "5":
		qb.AddCustomCondition("f.konto like '5%'")
	case "6":
		qb.AddCustomCondition("f.konto like '6%'")
	default:
		qb.AddCustomCondition("(f.konto like '5%' OR f.konto like '6%')")
	}
	qb.AddGroupBy("f.konto, f.sifra, fkpl.naziv, m.mtroska")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	if entities != nil {
		for _, entity := range *entities {
			// Monthly data
			fields := []string{
				entity.Konto,
				entity.Naziv,
				entity.Mtroska.String,
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil

}

func (s *SaldaResource) GetSaldaKlase5i6MTStampa(ctx context.Context, tbl *domain.TableData, saldaParam domain.SaldaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		SELECT f.konto, fkpl.naziv,
			COALESCE(m.mtroska, '') as mtroska,
			COALESCE(SUM(CASE WHEN f.kat IN ('1','2') THEN f.iznos ELSE 0 END), 0) as dug,
			COALESCE(SUM(CASE WHEN f.kat IN ('3','4') THEN f.iznos ELSE 0 END), 0) as pot
		FROM fpro f`, true)
	qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = f.idfkpl")
	qb.AddJoin("LEFT JOIN mestotr m ON m.mestotrid = f.mestotrid")

	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}

	switch saldaParam.Klasa {
	case "5":
		qb.AddCustomCondition("f.konto like '5%'")
	case "6":
		qb.AddCustomCondition("f.konto like '6%'")
	default:
		qb.AddCustomCondition("(f.konto like '5%' OR f.konto like '6%')")
	}

	qb.AddCondition("f.konto", saldaParam.OdKonta, ">=")
	qb.AddCondition("f.konto", saldaParam.DoKonta, "<=")
	qb.AddCondition("f.danal::date", saldaParam.OdDatuma, ">=")
	qb.AddCondition("f.danal::date", saldaParam.DoDatuma, "<=")

	qb.AddGroupBy("COALESCE(m.mtroska, ''), f.konto, fkpl.naziv")
	qb.AddOrderBy("COALESCE(m.mtroska, ''), f.konto")

	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	translator := i18n.GetInstance()
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = translator.Label("Ukupno za izveštaj")
	tbl.HasTotals = true

	var grandDug, grandPot float64
	lastMT := "\x01"
	var mtDug, mtPot float64

	emitMTSubtotal := func() {
		mtSaldo := mtDug - mtPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "group-total",
			Fields: []string{
				translator.Label("Ukupno za MT") + ": " + lastMT,
				"",
				common.FormatNumberWithSystemLocale(mtDug, 2),
				common.FormatNumberWithSystemLocale(mtPot, 2),
				common.FormatNumberWithSystemLocale(mtSaldo, 2),
			},
		})
		mtDug, mtPot = 0, 0
	}

	if entities != nil {
		for _, entity := range *entities {
			mt := entity.Mtroska.String

			if mt != lastMT {
				if lastMT != "\x01" {
					emitMTSubtotal()
				}
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "group-header",
					Fields:   []string{"MT", mt, "", "", ""},
				})
				lastMT = mt
			}

			dug := entity.Dug.Float64
			pot := entity.Pot.Float64
			saldo := dug - pot
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					entity.Konto,
					entity.Naziv,
					common.FormatNumberWithSystemLocale(dug, 2),
					common.FormatNumberWithSystemLocale(pot, 2),
					common.FormatNumberWithSystemLocale(saldo, 2),
				},
			})
			mtDug += dug
			mtPot += pot
			grandDug += dug
			grandPot += pot
		}
		if lastMT != "\x01" {
			emitMTSubtotal()
		}
	}

	tbl.Totals[2] = common.FormatNumberWithSystemLocale(grandDug, 2)
	tbl.Totals[3] = common.FormatNumberWithSystemLocale(grandPot, 2)
	tbl.Totals[4] = common.FormatNumberWithSystemLocale(grandDug-grandPot, 2)
	return nil
}

func (s *SaldaResource) SaldaPoKomercijalistima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build commercialist balance query
	qb := common.NewQueryBuilder(`select 
            coalesce(fpro.komid, 0) as komid,
            coalesce(kom.sifkom, 0) as sifkom,
            coalesce(kom.imeprezime, '') as imeprezime,
            coalesce(sum(case when fpro.kat in (1,2) then fpro.iznos else 0 end), 0) as dug,
    		coalesce(sum(case when fpro.kat in (3,4) then fpro.iznos else 0 end), 0) as pot,
    		coalesce(sum(case when fpro.kat in (1,2) then fpro.iznos 
                    when fpro.kat in (3,4) then 0-fpro.iznos 
                                 else 0 end), 0) as dospelo
        from fpro`, true)

	qb.AddJoin("left join komercijalisti as kom on kom.komid = fpro.komid")
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.vrd = 10 OR fpro.vrd = 30)")
	// if search text is not empty, add search conditions for commercialist fields
	if searchText != "" {
		likePattern := fmt.Sprintf("%%%s%%", searchText)
		qb.AddCustomCondition("(kom.sifkom::TEXT ILIKE $1 OR kom.imeprezime::TEXT ILIKE $2)", likePattern, likePattern)
	}
	qb.AddGroupBy("fpro.komid, kom.sifkom, kom.imeprezime")
	qb.AddOrderBy("kom.sifkom asc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.saldaKomRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
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
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}
func (s *SaldaResource) RealizacijaKomercijalisti(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage int, pageSize int, searchText, sortBy, sortOrder string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build commercialist balance query
	qb := common.NewQueryBuilder(`select 
            coalesce(fpro.komid, 0) as komid,
            coalesce(kom.sifkom, 0) as sifkom,
            coalesce(kom.imeprezime, '') as imeprezime,
            coalesce(sum(case when fpro.kat in (1,2) then fpro.iznos else 0 end), 0) as dug,
    		coalesce(sum(case when fpro.kat in (3,4) then fpro.iznos else 0 end), 0) as pot,
    		coalesce(sum(case when (fpro.dadok::date + (fpro.rok || ' days')::interval)::date <= fpro.danal::date
           	then case when fpro.kat in (3,4) then fpro.iznos 
                                 else 0 end
                  else 0 end), 0) as vanperioda
        from fpro`, true)

	qb.AddJoin("left join komercijalisti as kom on kom.komid = fpro.komid")
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddEqual("fpro.vkonta", 1)
	qb.AddCustomCondition("(fpro.vrd = 10 OR fpro.vrd = 30)")
	// if search text is not empty, add search conditions for commercialist fields
	if searchText != "" {
		likePattern := fmt.Sprintf("%%%s%%", searchText)
		qb.AddCustomCondition("(kom.sifkom::TEXT ILIKE $1 OR kom.imeprezime::TEXT ILIKE $2)", likePattern, likePattern)
	}
	qb.AddGroupBy("fpro.komid, kom.sifkom, kom.imeprezime")
	qb.AddOrderBy("kom.sifkom::NUMERIC asc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.saldaKomRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
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
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
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
func buildNaziv(naziv, pib, jmbg, bpg, index string) string {
	if pib == "" && jmbg == "" && bpg == "" && index == "" {
		return naziv
	}

	var sb strings.Builder
	sb.WriteString(naziv)
	sb.WriteString("\n   ")

	if pib != "" {
		sb.WriteString(" PIB:")
		sb.WriteString(pib)
		sb.WriteString("   ")
	}
	if jmbg != "" {
		sb.WriteString(" JMBG:")
		sb.WriteString(jmbg)
		sb.WriteString("   ")
	}
	if bpg != "" {
		sb.WriteString(" BPG:")
		sb.WriteString(bpg)
		sb.WriteString("   ")
	}
	if index != "" {
		sb.WriteString(" INDEX:")
		sb.WriteString(index)
	}

	return sb.String()
}
func (s *SaldaResource) extractGrupa(konto string, nDuzSin int) string {
	if len(konto) > nDuzSin {
		return konto[:nDuzSin]
	}
	return konto
}
func (s *SaldaResource) GetSaldaTotalValues(ctx context.Context, konto, sifra, tipKonta string) (domain.SaldaDto, error) {
	var totals domain.SaldaDto

	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return totals, fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`
		SELECT 
			COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) as pstdug,
			COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pstpot,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('1', '2') THEN fpro.iznos ELSE 0 END), 0) as dug,
			COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat IN ('3', '4') THEN fpro.iznos ELSE 0 END), 0) as pot
			FROM fpro`, true)
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	if tipKonta == "3" {
		qb.AddLikeBegin("fpro.konto", konto)
	}
	if tipKonta == "2" {
		qb.AddEqual("fpro.konto", konto)
	}
	if tipKonta == "1" {
		qb.AddEqual("fpro.konto", konto)
		qb.AddEqual("fpro.sifra", sifra)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return totals, err
	}
	if entities != nil && len(*entities) > 0 {
		totals.PocStanjeDug = (*entities)[0].PSTDug
		totals.PocStanjePot = (*entities)[0].PSTPot
		totals.PocStanjeSaldo = (*entities)[0].PSTDug - (*entities)[0].PSTPot
		totals.TekuciPromDug = (*entities)[0].Dug.Float64
		totals.TekuciPromPot = (*entities)[0].Pot.Float64
		totals.TekuciPromSaldo = (*entities)[0].Dug.Float64 - (*entities)[0].Pot.Float64
		totals.UkPromDug = totals.PocStanjeDug + totals.TekuciPromDug
		totals.UkPromPot = totals.PocStanjePot + totals.TekuciPromPot
		totals.UkPromSaldo = totals.PocStanjeSaldo + totals.TekuciPromSaldo
	}

	return totals, nil
}

func (s *SaldaResource) GetSaldaKlase5i6TotalValues(ctx context.Context, saldaParam domain.SaldaParam) (domain.SaldaDto, error) {
	var totals domain.SaldaDto

	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return totals, fmt.Errorf("user session not found")
	}

	// Create first query (opening balance - tipdok = '00')
	qbPst := common.NewQueryBuilder(`SELECT 
    COALESCE(SUM(CASE WHEN f.kat = '1' OR f.kat = '2' THEN f.iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN f.kat = '3' OR f.kat = '4' THEN f.iznos ELSE 0 END), 0) as pot
    FROM fpro f`, true)
	qb := common.NewQueryBuilder(`SELECT 
    COALESCE(SUM(CASE WHEN f.kat = '1' OR f.kat = '2' THEN f.iznos ELSE 0 END), 0) as dug,
    COALESCE(SUM(CASE WHEN f.kat = '3' OR f.kat = '4' THEN f.iznos ELSE 0 END), 0) as pot
    FROM fpro f`, true)
	// Add conditions to first query
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
		qbPst.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
		qbPst.AddEqual("f.kar", userSession.SelectedKar)
	}
	if saldaParam.OdSifre != "undefined" && saldaParam.DoSifre == "undefined" {
		qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.OdSifre, ">=")
		qb.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.DoSifre, "<=")
		qbPst.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.OdSifre, ">=")
		qbPst.AddCondition("COALESCE(NULLIF(COALESCE(f.sifra, ''), '')::numeric, 0)", saldaParam.DoSifre, "<=")
	}
	if saldaParam.OdKonta != "undefined" && saldaParam.DoKonta == "undefined" {
		qb.AddCondition("COALESCE(NULLIF(COALESCE(f.konto, ''), '')::numeric, 0)", saldaParam.OdKonta, ">=")
		qb.AddCondition("COALESCE(NULLIF(COALESCE(f.konto, ''), '')::numeric, 0)", saldaParam.DoKonta, "<=")
		qbPst.AddCondition("COALESCE(NULLIF(COALESCE(f.konto, ''), '')::numeric, 0)", saldaParam.OdKonta, ">=")
		qbPst.AddCondition("COALESCE(NULLIF(COALESCE(f.konto, ''), '')::numeric, 0)", saldaParam.DoKonta, "<=")
	}

	qb.AddCondition("f.danal::date", saldaParam.OdDatuma, ">=")
	qb.AddCondition("f.danal::date", saldaParam.DoDatuma, "<=")
	qb.AddCondition("f.tipdok", "00", "<>")

	qbPst.AddEqual("f.tipdok", "00")
	// Check account range between 5 and 6 (class 5 and 6)
	switch saldaParam.Klasa {
	case "5":
		qb.AddCustomCondition("f.konto like '5%'")
		qbPst.AddCustomCondition("f.konto like '5%'")
	case "6":
		qb.AddCustomCondition("f.konto like '6%'")
		qbPst.AddCustomCondition("f.konto like '6%'")
	default:

	}
	sqlQuery, args := qb.Build()
	sqlQueryPst, argsPst := qbPst.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return totals, err
	}
	if entities != nil && len(*entities) > 0 {
		totals.TekuciPromDug = (*entities)[0].Dug.Float64
		totals.TekuciPromPot = (*entities)[0].Pot.Float64
		totals.TekuciPromSaldo = (*entities)[0].Dug.Float64 - (*entities)[0].Pot.Float64
	}
	entitiesPst, err := s.fproRepo.GetAllCustom(ctx, sqlQueryPst, "", argsPst, "", "")
	if err != nil {
		return totals, err
	}
	if entitiesPst != nil && len(*entitiesPst) > 0 {
		totals.PocStanjeDug = (*entitiesPst)[0].Dug.Float64
		totals.PocStanjePot = (*entitiesPst)[0].Pot.Float64
		totals.PocStanjeSaldo = (*entitiesPst)[0].Dug.Float64 - (*entitiesPst)[0].Pot.Float64
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

func (s *SaldaResource) ProcessSaldaPartneriDetails(ctx context.Context, idPartneri int64, tblKonta, tblDetalji *domain.TableData, searchText, sortBy, sortOrder string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	bBudzetsko := false

	hasGod, hasKar := s.partneriRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT budzetski FROM partneri`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("idpartneri", idPartneri)
	sqlQuery, args := qb.Build()
	partneriEntities, err := s.partneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if partneriEntities != nil && len(*partneriEntities) > 0 {
		bBudzetsko = (*partneriEntities)[0].Budzetski == true //budzetski
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
        FROM fpro fp `, true)
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
	kontaEntities, err := s.fproRepo.GetAllCustom(ctx, sqlQueryKonta, "", argsKonta, "", "")
	if err != nil {
		return err
	}

	// Populate Konta table
	if kontaEntities != nil && len(*kontaEntities) > 0 {
		for _, konta := range *kontaEntities {
			fields := []string{
				fmt.Sprintf("%s %s-%s", konta.Konto, konta.Sifra, konta.Naziv),
				common.FormatNumberWithSystemLocale(konta.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(konta.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(konta.Dug.Float64-konta.Pot.Float64, 2),
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
        FROM fpro fp `, true)
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
	detaljiEntities, err := s.fproRepo.GetAllCustom(ctx, sqlQueryDetalji, "", argsDetalji, "", "")
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
				detail.Danal.Format(common.DateLayout),
				sVrd,
				detail.Dokum,
				detail.Dadok.Time.Format(common.DateLayout),
				fmt.Sprintf("%d", detail.Tra),
				common.FormatNumberWithSystemLocale(detail.Iznos, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tblDetalji.Rows = append(tblDetalji.Rows, tblRow)
		}
	}

	return nil
}
func (s *SaldaResource) GetFieldCache() map[string]reflect.StructField {
	return s.service.GetFieldCache()
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
func (s *SaldaResource) GetSaldaPartneriPrelomljenoStampaFields() []domain.Fields {
	return s.saldaPartneriPrelomljenoStampaTableFields
}
func (s *SaldaResource) GetSaldaKlase5i6AnalitikaTableFields() []domain.Fields {
	return s.saldaKlase5i6AnalitikaTableFields
}
func (s *SaldaResource) GetSaldaKlase5i6AnalitikaStampaFields() []domain.Fields {
	return s.saldaKlase5i6AnalitikaStampaTableFields
}
func (s *SaldaResource) GetSaldaKlase5i6MTTableFields() []domain.Fields {
	return s.saldaKlase5i6MTTableFields
}

func (s *SaldaResource) GetSaldaKlase5i6MTStampaFields() []domain.Fields {
	return s.saldaKlase5i6MTStampaTableFields
}

func (s *SaldaResource) GetKomercijalistiTableFields() []domain.Fields {
	return s.saldaKomercijalistiTableFields
}
func (s *SaldaResource) GetRealizacijaKomercijalistiTableFields() []domain.Fields {
	return s.saldaRealizacijakomercijalistiTableFields
}
func (s *SaldaResource) GetSaldaPartneraPoKontimaStampaFields() []domain.Fields {
	return s.saldaPartneraPoKontimaStampaTableFields
}

func (s *SaldaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

func (s *SaldaResource) setServiceFieldValues() {
	s.saldaPojedinacniTableFields = []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "10"},
		{Name: "duguje", Label: "Duguje", Width: "12", TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12", TextAlign: "right"},
		{Name: "saldo", Label: "Saldo u mesecu", Width: "12", TextAlign: "right"},
		{Name: "saldokumul", Label: "Saldo na kraju meseca", Width: "15", TextAlign: "right"},
	}
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
		{Name: "detalji", Label: "Detalji", Width: "10"},
		{Name: "rbr", Label: "Redni Broj", Width: "10"},
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Šifra", Width: "10"},
		{Name: "naziv", Label: "Naziv konta", Width: "60"},
		{Name: "pst", Label: "Početno stanje", Width: "12", Field: "pst", SkipInSearch: true},
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
	s.saldaPartneriPrelomljenoStampaTableFields = []domain.Fields{
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "partneri.sifra"},
		{Name: "naziv", Label: "Naziv partnera", Width: "30", Field: "partneri.naziv"},
		{Name: "adresa", Label: "Adresa", Width: "20", Field: "partneri.adresa"},
		{Name: "pobro", Label: "Poštanski broj", Width: "8", Field: "partneri.pobro", TextAlign: "center"},
		{Name: "mesto", Label: "Mesto", Width: "12", Field: "partneri.mesto"},
		{Name: "pib", Label: "PIB", Width: "10", Field: "partneri.pib", TextAlign: "center"},
		{Name: "kupac", Label: "Saldo Kupac", Width: "10", Field: "kupac", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "dobavljac", Label: "Saldo Dobavljač", Width: "10", Field: "dobavljac", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "primljenavans", Label: "Primljeni Avansi", Width: "10", Field: "primljenavans", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "datavans", Label: "Dati Avansi", Width: "10", Field: "datavans", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "stanje", Label: "Stanje", Width: "10", Field: "stanje", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "dugpot", Label: "Duguje/Potražuje", Width: "10", Field: "dugpot", SkipInSearch: true, TextAlign: "center"},
	}

	s.saldaKlase5i6AnalitikaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "60"},
		{Name: "duguje", Label: "Duguje", Width: "30", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "30", Field: "saldo", SkipInSearch: true},
	}
	s.saldaKlase5i6AnalitikaStampaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "40"},
		{Name: "mtroska", Label: "Mesto troška", Width: "12", TextAlign: "center"},
		{Name: "dug", Label: "Duguje", Width: "13", TextAlign: "right", IncludeInTotals: true},
		{Name: "pot", Label: "Potražuje", Width: "13", TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "12", TextAlign: "right", IncludeInTotals: true},
	}
	s.saldaKlase5i6MTTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "60"},
		{Name: "mtroska", Label: "Mesto troska", Width: "12", Field: "konto"},
		{Name: "duguje", Label: "Duguje", Width: "30", Field: "duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12", Field: "potrazuje", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "30", Field: "saldo", SkipInSearch: true},
	}
	s.saldaKlase5i6MTStampaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "55"},
		{Name: "dug", Label: "Duguje", Width: "15", TextAlign: "right", IncludeInTotals: true},
		{Name: "pot", Label: "Potražuje", Width: "15", TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "15", TextAlign: "right", IncludeInTotals: true},
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
	s.saldaPartneraPoKontimaStampaTableFields = []domain.Fields{
		{Name: "fu", Label: "F/U", Width: "5", TextAlign: "center"},
		{Name: "brnal", Label: "Broj naloga", Width: "12"},
		{Name: "danal", Label: "Datum naloga", Width: "12", TextAlign: "center"},
		{Name: "vrstadok", Label: "Vrsta dokum.", Width: "20"},
		{Name: "brmdok", Label: "Broj dokum.", Width: "12"},
		{Name: "dadok", Label: "Datum dokum.", Width: "12", TextAlign: "center"},
		{Name: "god", Label: "Godina dokum.", Width: "8", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos", Width: "15", TextAlign: "right"},
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

// Helper function to convert SaldaDto to TableRow with month names
func saldaDtoToTableRows(saldaData []domain.SaldaDto) []domain.TableRow {
	monthNames := []string{
		"Početno stanje",
		"Januar",
		"Februar",
		"Mart",
		"April",
		"Maj",
		"Jun",
		"Jul",
		"Avgust",
		"Septembar",
		"Oktobar",
		"Novembar",
		"Decembar",
	}

	rows := make([]domain.TableRow, len(saldaData))
	for i, salda := range saldaData {
		monthName := monthNames[i]
		if monthName == "" {
			monthName = fmt.Sprintf("Mesec %d", salda.Mesec)
		}

		rows[i] = domain.TableRow{
			Fields: []string{
				monthName,
				common.FormatNumberWithSystemLocale(salda.Duguje, 2),
				common.FormatNumberWithSystemLocale(salda.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(salda.Saldo, 2),
				common.FormatNumberWithSystemLocale(salda.SaldoKumul, 2),
			},
		}
	}
	return rows
}

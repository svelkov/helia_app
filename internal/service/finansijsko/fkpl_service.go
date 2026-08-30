package finansijsko

import (
	"context"
	"errors"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// FkplService defines the interface for operations related to Fkpl.
type FkplService interface {
	service.Service[domain.Fkpl]
	TraziKonto(ctx context.Context, konto, sifra, vkonta string) (entities *[]domain.Fkpl, err error)
	KontoSearchForTable(ctx context.Context, tbl *domain.TableData, searchValue, konto, vkonta, fieldName string, fieldColIndex int) error
	GetAllByVkonta(ctx context.Context, vkonta string, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, printType string) error
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetFkplTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Fkpl) []domain.FieldError
	GetAnalitikaForSelect(ctx context.Context) ([]domain.ComboItem, error)
}

// FkplResource implements the FkplService interface.
type FkplResource struct {
	service          *service.BaseService[domain.Fkpl]
	fkplRepo         *repository.BaseRepository[domain.Fkpl]
	fvrRepo          *repository.BaseRepository[domain.Fvr]
	tipAnalitikeRepo *repository.BaseRepository[domain.Tipanalitike]
	partneriRepo     *repository.BaseRepository[domain.Partneri]
	fkplTableFields  []domain.Fields
	cfg              config.Config
}

func NewFkplResource(service *service.BaseService[domain.Fkpl], fkplRepo *repository.BaseRepository[domain.Fkpl], fvrRepo *repository.BaseRepository[domain.Fvr], tipAnalitikeRepo *repository.BaseRepository[domain.Tipanalitike], partneriRepo *repository.BaseRepository[domain.Partneri], cfg config.Config) *FkplResource {
	rs := &FkplResource{
		service:          service,
		fkplRepo:         fkplRepo,
		fvrRepo:          fvrRepo,
		tipAnalitikeRepo: tipAnalitikeRepo,
		partneriRepo:     partneriRepo,
		cfg:              cfg,
	}	
	rs.setKontniPlanTableFields()
	return rs
}

func (s *FkplResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return common.GetFvrData(ctx, s.fvrRepo)
}
func (s *FkplResource) GetAnalitikaForSelect(ctx context.Context) ([]domain.ComboItem, error) {
	if s.tipAnalitikeRepo == nil {
		return nil, fmt.Errorf("tipAnalitikeRepo not initialized")
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	hasGod, hasKar := s.tipAnalitikeRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT tipanalitikeid, naziv FROM tipanalitike", true)
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.Build()
	entiteis, err := s.tipAnalitikeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, fmt.Errorf("error fetching tip analitike: %w", err)
	}
	comboItems := []domain.ComboItem{}
	for entity := range *entiteis {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", (*entiteis)[entity].TipanalitikeID),
			Value: (*entiteis)[entity].Naziv,
		})
	}

	return comboItems, nil
}

func (s *FkplResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}
func (s *FkplResource) GetFkplTableFields() []domain.Fields {
	return s.fkplTableFields
}

func (s *FkplResource) Create(ctx context.Context, entity *domain.Fkpl, idFkpl string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "fkpl",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   s.GetFkplTableFields(),
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildInsert(ctx, fields, common.IDfkpl)
	// Start a transaction
	tx, err := s.fkplRepo.BeginTx()
	if err != nil {
		return nil, 0, fmt.Errorf("error beginning transaction: %w", err)
	}
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	err = tx.QueryRowContext(ctx, sqlQuery, args...).Err()
	if err != nil {
		return nil, 0, fmt.Errorf("insert fkpl failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, 0, nil
}
func (s *FkplResource) Update(ctx context.Context, entity *domain.Fkpl, idField string, idFiledValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "fkpl",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   fields,
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildUpdate(ctx, fields, idField, idFiledValue)
	// Start a transaction
	tx, err := s.fkplRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	err = tx.QueryRowContext(ctx, sqlQuery, args...).Err()
	if err != nil {
		return nil, fmt.Errorf("insert fkpl failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, nil

}
func (s *FkplResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}
func (s *FkplResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Fkpl, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

// GetAllCustom implements FkplService.
func (s *FkplResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fkpl, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}
func (s *FkplResource) GetAllByVkonta(ctx context.Context, vkonta string, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT idfkpl, konto, coalesce(a.naziv, '') as annaziv, fkpl.naziv, devizni, kursirati, vkonta, kolicinski FROM fkpl", true)
	qb.AddJoin(" left join tipanalitike a on a.tipanalitikeid = fkpl.tipanalitikeid")
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if vkonta != "" {
		qb.AddEqual("vkonta", vkonta)
	}
	if searchText != "" && printType == common.TipStampePreview {
		qb.AddCustomSearchCondition([]string{"konto", "a.naziv", "fkpl.naziv"}, searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("fkpl.%s", sortBy))
		}
		if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
			qb.AddSortOrder(sortOrder)
		}
		if printType == common.TipStampePreview {
			qb.SetLimit(pageSize)
			qb.SetOffset((currentPage - 1) * pageSize)
		}
	}

	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	for _, entity := range *entities {
		fields := []string{
			entity.Konto,
			entity.AnalitikaNaziv,
			entity.Naziv,
			fmt.Sprintf("%t", entity.Devizni),
			fmt.Sprintf("%t", entity.Kursirati),
			fmt.Sprintf("%t", entity.Kolicinski),
			fmt.Sprintf("%d", entity.Vkonta),
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

// GetByID implements FkplService.
func (s *FkplResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Fkpl, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT idfkpl, konto, coalesce(a.naziv, '') as annaziv, fkpl.naziv, devizni, kolicinski, vkonta, fkpl.tipanalitikeid FROM fkpl", true)
	qb.AddJoin(" left join tipanalitike a on a.tipanalitikeid = fkpl.tipanalitikeid")
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual(idField, idValue)

	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	if len(*entities) == 0 {
		return nil, errors.New("Entity not found")
	}
	return &(*entities)[0], nil
}

// GetTotalRecords implements FkplService.
func (s *FkplResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

// GetTotalRecordsCustom implements FkplService.
func (s *FkplResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements FkplService.
func (s *FkplResource) MapEntityToValues(entity *domain.Fkpl, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}
func (s *FkplResource) ValidateEntity(ctx context.Context, entity *domain.Fkpl) []domain.FieldError {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "Korisnička sesija nije pronađena"}}
	}

	var fieldErrors []domain.FieldError
	if entity.Konto == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Obavezan podatak..."})
	}
	if entity.Naziv == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "naziv", ErrorMessage: "Obavezan podatak..."})
	}
	if entity.Vkonta == 1 && entity.TipanalitikeID == nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "tipanalitikeid", ErrorMessage: "Obavezan podatak..."})
	}
	if entity.Vkonta == 2 && len(entity.Konto) <= s.cfg.NDuzSint {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: fmt.Sprintf("Subsintetički konto mora imati minimum %d karaktera", s.cfg.NDuzSint+1)})
	}
	if entity.Vkonta == 3 && len(entity.Konto) != s.cfg.NDuzSint {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: fmt.Sprintf("Sintetički konto mora imati tačno %d karaktera", s.cfg.NDuzSint)})
	}
	if entity.Vkonta == 4 && len(entity.Konto) != 2 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Grupa konto mora imati tačno 2 karaktera"})
	}
	if entity.Vkonta == 5 && len(entity.Konto) != 1 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Klasa konto mora imati tačno 1 karakter"})
	}
	qb := common.NewQueryBuilder("SELECT idfkpl FROM fkpl", true)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}
	qb.AddEqual("konto", entity.Konto)
	qb.AddEqual("tipanalitikeid", entity.TipanalitikeID)
	sqlQuery, args := qb.Build()
	existingEntity, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Greška prilikom provere jedinstvenosti konta"})
	} else if len(*existingEntity) > 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Konto sa ovom šifrom i vrstom analitike već postoji"})
	}
	return fieldErrors
}

func (s *FkplResource) TraziKonto(ctx context.Context, konto, sifra, vkonta string) (entities *[]domain.Fkpl, err error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	var qb *common.QueryBuilder
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
	// if vkonta == "1"  then search ibn table partneri for the tipanalitke for that konto and use that tipanalitikeid to search in fkpl, if vkonta == "2" then search directly in fkpl with vkonta = 2 and konto
	if vkonta == "1" {
		qb = common.NewQueryBuilder(`SELECT ta.tipanalitikeid, p.naziv as naziv, p.sifra as sifra FROM partneri as p `, true)
		qb.AddJoin("inner join tipanalitike ta on ta.tipanalitikeid = p.tipanalitikeid ")
		qb.AddJoin(" inner join fkpl f on f.tipanalitikeid = ta.tipanalitikeid ")
	}
	if vkonta == "2" {
		qb = common.NewQueryBuilder(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM fkpl as f`, true)
	}
	if qb == nil {
		return nil, errors.New("Invalid vkonta value")
	}
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", session.SelectedKar)
	}

	if konto != "" {
		qb.AddEqual("f.konto", konto)
	}
	// only search in fkpl if vkonta is 2, if vkonta is 1 then search in partneri and join with fkpl through tipanalitikeid
	if vkonta == "2" {
		qb.AddEqual("f.vkonta", vkonta)
	}
	if vkonta == "1" && sifra != "" {
		qb.AddEqual("p.sifra", sifra)
	}

	sqlQuery, args := qb.Build()
	entities, err = s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, errors.New(common.ErrMsgSearchKonto)
	}
	if len(*entities) == 0 {
		return nil, errors.New(common.ErrMsgNotFound)
	}
	return entities, nil
}
func (s *FkplResource) KontoSearchForTable(ctx context.Context, tbl *domain.TableData, searchValue, konto, vkonta, fieldName string, fieldColIndex int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}
	var qb *common.QueryBuilder

	// if vkonta == "1"  then search ibn table partneri for the tipanalitke for that konto and use that tipanalitikeid to search in fkpl, if vkonta == "2" then search directly in fkpl with vkonta = 2 and konto
	if vkonta == "1" {
		qb = common.NewQueryBuilder(`SELECT distinct ta.tipanalitikeid, p.naziv as naziv, p.sifra as sifra, f.konto FROM partneri as p `, true)
		qb.AddJoin("left join tipanalitike ta on ta.tipanalitikeid = p.tipanalitikeid ")

		hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
		fkplJoin := " left join fkpl f on f.tipanalitikeid = ta.tipanalitikeid "
		if hasGod {
			fkplJoin += fmt.Sprintf(" and f.god = %d ", session.SelectedGod)
		}
		if hasKar {
			fkplJoin += fmt.Sprintf(" and f.kar = %d ", session.SelectedKar)
		}
		if searchValue != "" {
			fkplJoin += fmt.Sprintf(" and (f.sifra ILIKE '%%' || $%d || '%%' OR f.naziv ILIKE '%%' || $%d || '%%')", 1, 1)
			qb.AddArgs(searchValue)
		}
		qb.AddJoin(fkplJoin)
		hasGod, hasKar = s.partneriRepo.GetHasGodHasKar()
		if hasGod {
			qb.AddEqual("p.god", session.SelectedGod)
		}
		if hasKar {
			qb.AddEqual("p.kar", session.SelectedKar)
		}
		if searchValue != "" {
			nbrArgs := qb.GetArgsCount()
			qb.AddCustomCondition(fmt.Sprintf(`(p.sifra ILIKE '%%' || $%d || '%%' OR p.naziv ILIKE '%%' || $%d || '%%')`, nbrArgs+1, nbrArgs+1))
			qb.AddArgs(searchValue)
		}
		qb.AddOrderBy("p.sifra")
	}
	if vkonta == "2" {
		hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
		qb = common.NewQueryBuilder(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM fkpl as f`, true)
		if hasGod {
			qb.AddEqual("f.god", session.SelectedGod)
		}
		if hasKar {
			qb.AddEqual("f.kar", session.SelectedKar)
		}
		qb.AddEqual("f.vkonta", "2")
		if searchValue != "" {
			nbrArgs := qb.GetArgsCount()
			qb.AddCustomCondition(fmt.Sprintf(`(f.konto ILIKE '%%' || $%d || '%%' OR f.sifra ILIKE '%%' || $%d || '%%' OR f.naziv ILIKE '%%' || $%d || '%%')`, nbrArgs+1, nbrArgs+1, nbrArgs+1))
			qb.AddArgs(searchValue)
		}
		qb.AddOrderBy("f.konto")
	}

	qb.SetLimit(20)

	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return errors.New(common.ErrMsgReadData)
	}

	// Prepare table data
	tbl.ShowActions = false
	tbl.ShowPagination = false
	tbl.FuncClick = "selectRow"
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)"

	tbl.DestField = fieldName
	tbl.DestFieldColIndex = fieldColIndex
	for _, entity := range *entities {
		fields := []string{}
		// Add common fields
		fields = append(fields,
			entity.Konto,
			entity.Sifra,
			entity.Naziv)
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
	tbl.BtnAdd = domain.Button{IsVisible: false}
	tbl.BtnPrint = domain.Button{IsVisible: false}

	return nil
}

func (s *FkplResource) GetTableFileds() []domain.Fields {
	return s.fkplTableFields
}

func (s *FkplResource) setKontniPlanTableFields() {
	s.fkplTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10", Sortable: true},
		{Name: "annnaziv", Label: "Naziv Analitike", Width: "10"},
		{Name: "naziv", Label: "Naziv konta", Width: "80", Sortable: true},
		{Name: "devizni", Label: "Devizno konto", Type: "checkbox", Width: "5", TextAlign: "center"},
		{Name: "kursirati", Label: "Kursirati konto", Type: "checkbox", Width: "5", TextAlign: "center"},
		{Name: "kolicinski", Label: "Kolicinski", Type: "checkbox", Width: "5", TextAlign: "center"},
		{Name: "vkonta", Label: "Vrsta konta", Width: "4", TextAlign: "center"},
	}

}

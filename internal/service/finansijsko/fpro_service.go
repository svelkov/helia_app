package finansijsko

import (
	"context"
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
	"helia/internal/service"
	"math"
	"reflect"
	"strings"
	"time"
)

const (
	urlFproStavka      = "/api/fpro/stavka"
	urlFproNalogStavke = "/api/fpro/nalog/:%d"
)

// FproViewData encapsulates all data needed for the Nalog display page.
type FproViewData struct {
	FproEntities []domain.Fpro
	TableData    domain.TableData
}

// NalogService defines the interface for operations related to Fpro (Nalogs).
type FproService interface {
	service.Service[domain.Fpro]
	GetAllFproByFnalID(ctx context.Context, fproPayload *domain.FproPayload, tbl *domain.TableData, fnalID int64, currentPage, pageSize int, searchText string) error
	SaveNalogStavke(ctx context.Context, fproStavke *domain.FproPayload) error
	DeleteFpro(ctx context.Context, id int64, tblStavke *domain.TableData, currentPage, pageSize int) error
	FproValidate(ctx context.Context, entity *domain.FproPayload) []domain.FieldError
	GetNalogTotalValues(ctx context.Context, nalogTotal *domain.NalogTotalValues, idFnal int64) error
	GetOrgJedinice(ctx context.Context) ([]domain.ComboItem, error)
	GetMestoTroska(ctx context.Context, idorgjed int64) ([]domain.ComboItem, error)
	GetValute(ctx context.Context) ([]domain.ComboItem, error)
	GetKomercijalisti(ctx context.Context) ([]domain.ComboItem, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	GetMI(ctx context.Context) ([]domain.ComboItem, error)
	GetFieldCache() map[string]reflect.StructField
	SetNalogIDFieldName(string)
	GetTableStavkeFields() []domain.Fields
	GetTableNalogFields() []domain.Fields
}

// FproResource implements the FproService interface.
type FproResource struct {
	service                 *service.BaseService[domain.Fpro]
	fproRepo                repository.BaseRepository[domain.Fpro]
	fnalRepo                repository.BaseRepository[domain.Fnal]
	fkplRepo                repository.BaseRepository[domain.Fkpl]
	ojRepo                  repository.BaseRepository[domain.Orgjed]
	mtroskaRepo             repository.BaseRepository[domain.Mestotr]
	fvrRepo                 repository.BaseRepository[domain.Fvr]
	fvknjracRepo            repository.BaseRepository[domain.Fvknjrac]
	valuteRepo              repository.BaseRepository[domain.Valute]
	komercijalistiRepo      repository.BaseRepository[domain.Komercijalisti]
	miRepo                  repository.BaseRepository[domain.Fisp]
	magaciniRepo            repository.BaseRepository[domain.Magacini]
	fproIDFieldName         string
	naloziTableFields       []domain.Fields
	naloziStavkeTableFields []domain.Fields
	cfg                     config.Config
}

func NewFproService(
	Service *service.BaseService[domain.Fpro],
	fproRepo repository.BaseRepository[domain.Fpro],
	fnalRepo repository.BaseRepository[domain.Fnal],
	fkplRepo repository.BaseRepository[domain.Fkpl],
	ojRepo repository.BaseRepository[domain.Orgjed],
	mtroskaRepo repository.BaseRepository[domain.Mestotr],
	fvrRepo repository.BaseRepository[domain.Fvr],
	fvknjracRepo repository.BaseRepository[domain.Fvknjrac],
	valuteRepo repository.BaseRepository[domain.Valute],
	komercijalistiRepo repository.BaseRepository[domain.Komercijalisti],
	miRepo repository.BaseRepository[domain.Fisp],
	magaciniRepo repository.BaseRepository[domain.Magacini],
	fproIDFieldName string,
	naloziTableFields []domain.Fields,
	naloziStavkeTableFields []domain.Fields,
	cfg config.Config,
) *FproResource {
	rs := &FproResource{
		service:                 Service,
		fproRepo:                fproRepo,
		fnalRepo:                fnalRepo,
		fkplRepo:                fkplRepo,
		ojRepo:                  ojRepo,
		mtroskaRepo:             mtroskaRepo,
		fvrRepo:                 fvrRepo,
		fvknjracRepo:            fvknjracRepo,
		valuteRepo:              valuteRepo,
		komercijalistiRepo:      komercijalistiRepo,
		miRepo:                  miRepo,
		magaciniRepo:            magaciniRepo,
		fproIDFieldName:         fproIDFieldName,
		naloziTableFields:       naloziTableFields,
		naloziStavkeTableFields: naloziStavkeTableFields,
		cfg:                     cfg,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *FproResource) SetNalogIDFieldName(fproIDFieldName string) {
	s.fproIDFieldName = fproIDFieldName
}
func (s *FproResource) SetNaloziTableFields(naloziTableFields []domain.Fields) {
	s.naloziTableFields = naloziTableFields
}

func (s *FproResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}

// Create implements NalogService.
func (s *FproResource) Create(ctx context.Context, Fpro *domain.Fpro, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	return s.service.Create(ctx, Fpro, idField, fields)
}

// Delete implements NalogService.
func (s *FproResource) Delete(ctx context.Context, idField string, id int64) error {
	return s.service.Delete(ctx, idField, id)
}

// GetAll implements NalogService.
func (s *FproResource) GetAll(ctx context.Context, page int, offset int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Fpro, error) {
	return s.service.GetAll(ctx, page, offset, tableFields, idField, searchText, sortBy, sortOrder)
}

// GetAllCustom implements NalogService.
func (s *FproResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fpro, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements NalogService.
func (s *FproResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Fpro, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found in context")
	}
	qb := common.NewQueryBuilder(`select fp.idfpro, fp.rbr, fp.nalog, fp.tipdok, fp.danal, fp.iznos, fp.kat, coalesce(fp.opis, '') as opis, fp.dadok,
	fp.rok, fp.vrd, fp.vkonta, fp.konto, coalesce(fp.sifra, '') as sifra, fp.tra, fp.deviznos, fp.kurs,
	fp.sifval, fp.mi, coalesce(fp.dokum, '') as dokum, fp.idfnal, 
	coalesce(fp.idorgjed, 0) as idorgjed, coalesce(fp.komid, 0) as komid, coalesce(fp.mestotrid, 0) as mestotrid,  fp.idfkpl, 
	coalesce(fp.dokumv, '') as dokumv, fp.dadokv, fp.travez,
	coalesce(fk.naziv, '') as nazivkonta,
	coalesce(p.naziv, '') as NazivAnalitike,
	case when fp.kat in (1,2) then fp.iznos
		 else 0 end as dug,
	case when fp.kat in (3,4) then fp.iznos
		 else 0 end as pot
	from fpro as fp`, true)
	qb.AddJoin(` left join fkpl fk on fk.idfkpl = fp.idfkpl`)
	qb.AddJoin(` left join partneri p on p.sifra = fp.sifra and p.tipanalitikeid = fk.tipanalitikeid `)
	qb.AddEqual(fmt.Sprintf("fp.%s", idField), idValue)
	sqlQuery, args := qb.Build()
	entities, err := s.service.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	if entities == nil || len(*entities) == 0 {
		return nil, fmt.Errorf("Fpro not found for %s = %d", idField, idValue)
	}
	return &(*entities)[0], nil

}

// GetTotalRecords implements NalogService.
func (s *FproResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

// GetTotalRecordsCustom implements NalogService.
func (s *FproResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements NalogService.
func (s *FproResource) MapEntityToValues(entity *domain.Fpro, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

// Update implements NalogService.
func (s *FproResource) Update(ctx context.Context, entity *domain.Fpro, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	return s.service.Update(ctx, entity, idField, idValue, tableFields)
}

func (s *FproResource) GetAllFproByFnalID(ctx context.Context, fproPayload *domain.FproPayload, tbl *domain.TableData, fnalID int64, currentPage, pageSize int, searchText string) error {
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.ShowActions = true
	tbl.HasTotals = true
	qbCount := common.NewQueryBuilder(`select count(*) from fpro fp `, true)
	qbCount.AddJoin(` left join fkpl fk on fk.idfkpl = fp.idfkpl`)
	qbCount.AddEqual("fp.idfnal", fnalID)
	if searchText != "" {
		qbCount.AddLike("CONCAT( fp.nalog, ' ', fp.tipdok, ' ', fp.danal, ' ', fp.opis, ' ', fp.dadok, ' ', fp.rok, ' ', fp.vrd, ' ', fp.vkonta, ' ', fp.konto, ' ', fp.sifra, ' ', fk.naziv)", searchText)
	}
	sqlCountQuery, args := qbCount.Build()
	totalRecords, err := s.fproRepo.GetTotalRecordsCustom(ctx, sqlCountQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get total records for Fpro by fnalID: %w", err)
	}
	common.SetTableTotalRecords(tbl, totalRecords, pageSize)

	qb := common.NewQueryBuilder(`select fp.idfpro, fp.rbr, fp.nalog, fp.tipdok, fp.danal, fp.iznos, fp.kat, coalesce(fp.opis, '') as opis, fp.dadok,
	fp.rok, fp.vrd, fp.vkonta, fp.konto, coalesce(fp.sifra, '') as sifra, fp.tra, fp.deviznos, fp.kurs,
	fp.sifval, fp.mi, coalesce(fp.dokum, '') as dokum, fp.idfnal, 
	fp.idorgjed, fp.idfkpl, fp.komid, fp.mestotrid, 
	coalesce(fp.dokumv, '') as dokumv, fp.dadokv,	fp.travez, fk.naziv as naziv,
	case when fp.kat in (1,2) then fp.iznos
		 else 0 end as dug,
	case when fp.kat in (3,4) then fp.iznos
		 else 0 end as pot
	from fpro as fp`, true)
	qb.AddJoin(` left join fkpl fk on fk.idfkpl = fp.idfkpl`)
	qb.AddEqual("fp.idfnal", fnalID)
	if searchText != "" {
		qb.AddLike("CONCAT( fp.nalog, ' ', fp.tipdok, ' ', fp.danal, ' ', fp.opis, ' ', fp.dadok, ' ', fp.rok, ' ', fp.vrd, ' ', fp.vkonta, ' ', fp.konto, ' ', fp.sifra, ' ', fk.naziv)", searchText)
	}

	qb.AddOrderBy("fp.rbr")
	qb.AddSortOrder(" DESC")
	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get Fpro by fnalID: %w", err)
	}
	common.SetTableTotalRecords(tbl, totalRecords, pageSize)
	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		tbl.Totals = make([]string, len(tbl.Headers))
		tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
		var dugTotal, potTotal float64
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				fmt.Sprintf("%d", entity.Rbr),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				fmt.Sprintf("%d", entity.Vrd),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
			)
			dugTotal += entity.Dug.Float64
			potTotal += entity.Pot.Float64
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFpro), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
		tbl.Totals[6] = common.FormatNumberWithSystemLocale(dugTotal, 2)
		tbl.Totals[7] = common.FormatNumberWithSystemLocale(potTotal, 2)
	}

	mtroska := []domain.ComboItem{{Key: "-", Value: "-"}} // Default option when no records are found
	orgjed, err := s.GetOrgJedinice(ctx)
	if err != nil {
		return err
	}
	magacini, err := s.GetMagacini(ctx)
	if err != nil {
		return err
	}
	komercijalisti, err := s.GetKomercijalisti(ctx)
	if err != nil {
		return err
	}
	mi, err := s.GetMI(ctx)
	if err != nil {
		return err
	}
	valute, err := s.GetValute(ctx)
	if err != nil {
		return err
	}
	fproPayload.CbxOrgjed = orgjed
	fproPayload.CbxMtroska = mtroska
	fproPayload.CbxValute = valute
	fproPayload.CbxMi = mi
	fproPayload.CbxMagacin = magacini
	fproPayload.CbxKomercijalista = komercijalisti
	return err
}

// SaveNalogStavke implements the logic to save the nalog stavke (items) for a given nalog.
func (s *FproResource) SaveNalogStavke(ctx context.Context, fproStavke *domain.FproPayload) error {
	// Build the insert query using QueryBuilder
	var insertedID int64
	qb := common.NewRepositoryQueryBuilder(common.RepositoryConfig{
		TableName:     "fpro",
		EntityType:    reflect.TypeOf(domain.Fpro{}),
		TableFields:   s.GetTableStavkeFields(),
		IncludeGodKar: true,
	})
	fnal, err := s.fnalRepo.GetByID(ctx, common.IDfnal, fproStavke.IDFnal)
	if err != nil {
		return fmt.Errorf("failed to get fnal by ID: %w", err)
	}
	if fnal == nil {
		return fmt.Errorf("fnal not found for ID: %d", fproStavke.IDFnal)
	}
	fproStavke.Nalog = fmt.Sprintf("%d", fnal.Nalog)
	fproStavke.Tipdok = fnal.Tipdok
	fproStavke.Danal = fnal.Danal.Format(common.HtmlLayout)
	fproStavke.Datob = fnal.Datob.Format(common.HtmlLayout)
	if fproStavke.Opisknj == "" {
		fproStavke.Opisknj = "-"
	}
	if fproStavke.IDFpro == 0 {
		lastRbr, err := s.getLastRbr(ctx, fproStavke.IDFnal)
		if err != nil {
			return fmt.Errorf("failed to get last Rbr: %w", err)
		}
		fproStavke.Rbr = lastRbr + 1
	}
	fields := s.mapFieldsToValues(fproStavke)
	sqlQuery := ""
	args := []interface{}{}
	if fproStavke.IDFpro != 0 {
		sqlQuery, args = qb.BuildUpdate(ctx, fields, common.IDfpro, fproStavke.IDFpro)
	} else {
		sqlQuery, args = qb.BuildInsert(ctx, fields, common.IDfpro)
	}
	// Start a transaction
	tx, err := s.fproRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// check only we have insert record
	if fproStavke.IDFpro == 0 {
		err = tx.QueryRowContext(ctx, sqlQuery, args...).Scan(&insertedID)
		if err != nil {
			return fmt.Errorf("insert fpro failed: %w", err)
		}
	} else {
		_, err = tx.ExecContext(ctx, sqlQuery, args...)
		if err != nil {
			return fmt.Errorf("update fpro failed: %w", err)
		}
	}
	err = s.recalculateFnal(ctx, tx, fproStavke.IDFnal)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *FproResource) DeleteFpro(ctx context.Context, idFpro int64, tblStavke *domain.TableData, currentPage, pageSize int) error {
	// Get the Fpro record to be deleted
	fpro, err := s.fproRepo.GetByID(ctx, common.IDfpro, idFpro)
	if err != nil {
		return fmt.Errorf("failed to get Fpro by ID: %w", err)
	}
	if fpro == nil {
		return fmt.Errorf("Fpro not found for ID: %d", idFpro)
	}
	// Delete the Fpro record
	qb := common.NewQueryBuilder(`delete from fpro `, true)
	qb.AddEqual("idfpro", idFpro)
	sqlQuery, args := qb.Build()

	// Start a transaction
	tx, err := s.fproRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	_, err = tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Fpro: %w", err)
	}
	// delete kir
	if fpro.Vrd == 10 {
		qbDeleteKir := common.NewQueryBuilder(`delete from kir `, true)
		qbDeleteKir.AddEqual("idfpro", fpro.IDFpro)
		sqlUpdate, argsUpdate := qbDeleteKir.Build()
		_, err = tx.ExecContext(ctx, sqlUpdate, argsUpdate...)
		if err != nil {
			return fmt.Errorf("failed to delete kir: %w", err)
		}
	}
	// delete kpr
	if fpro.Vrd == 20 {
		qbDeleteKpr := common.NewQueryBuilder(`delete from kpr `, true)
		qbDeleteKpr.AddEqual("idfpro", fpro.IDFpro)

		sqlUpdate, argsUpdate := qbDeleteKpr.Build()
		_, err = tx.ExecContext(ctx, sqlUpdate, argsUpdate...)
		if err != nil {
			return fmt.Errorf("failed to delete kpr: %w", err)
		}
	}
	err = s.recalculateFnal(ctx, tx, fpro.IDFnal)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	// Refresh the table data after deletion
	err = s.GetAllFproByFnalID(ctx, &domain.FproPayload{}, tblStavke, fpro.IDFnal, currentPage, pageSize, "")
	if err != nil {
		return fmt.Errorf("failed to refresh table data: %w", err)
	}
	tblStavke.URLGetAll = fmt.Sprintf("/api/fpro/nalog/%d", fpro.IDFnal)
	tblStavke.URLPrefix = fmt.Sprintf("/api/fpro/nalog/%d", fpro.IDFnal)
	tblStavke.BtnDelete.HxActionURL = "/api/fpro/confirm-delete"
	tblStavke.BtnUpdate.HxActionURL = "/api/fpro/stavka/update"
	tblStavke.BtnUpdate.HxOnAfterRequest = "populateFproUpdateFormFromEvent(event)"
	tblStavke.BtnUpdate.HxSwap = "none"
	tblStavke.BtnUpdate.HxRequestType = "GET"
	tblStavke.DetailURL = fmt.Sprintf("/api/fpro/nalog/%d", fpro.IDFnal)
	tblStavke.SearchEnabled = true
	tblStavke.BtnAdd.IsVisible = false
	tblStavke.BtnPrint.IsVisible = false
	tblStavke.ShowActions = true
	return nil
}

// recalculateFnal recalculates dug, pot, brst and nalsts on fnal from the
// remaining fpro rows. It must be called within an open transaction.
func (s *FproResource) recalculateFnal(ctx context.Context, tx db.Transaction, idFnal int64) error {
	qb := common.NewQueryBuilder(`update fnal set
		dug    = (select coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) from fpro where idfnal = $1),
		pot    = (select coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) from fpro where idfnal = $1),
		brst   = (select count(*) from fpro where idfnal = $1),
		nalsts = case when
			(select coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) from fpro where idfnal = $1) =
			(select coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) from fpro where idfnal = $1)
			then 'Slozen' else 'Neslozen' end`, true)
	qb.AddArgs(idFnal)
	qb.AddEqual("idfnal", idFnal)
	sql, args := qb.Build()
	_, err := tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to recalculate fnal (id=%d): %w", idFnal, err)
	}
	return nil
}

func (s *FproResource) getLastRbr(ctx context.Context, fnlID int64) (int64, error) {
	// Implementation for getting the last Rbr
	qb := common.NewQueryBuilder(`select coalesce(max(rbr), 0) from fpro `, true)
	qb.AddEqual("idfnal", fnlID)
	sqlQuery, args := qb.Build()
	var lastRbr int64
	err := s.fproRepo.DB.QueryRowContext(ctx, sqlQuery, args...).Scan(&lastRbr)
	if err != nil {
		return 0, fmt.Errorf("failed to get last Rbr: %w", err)
	}
	return lastRbr, nil
}

func (s *FproResource) mapFieldsToValues(fproStavke *domain.FproPayload) []domain.Fields {
	fields := []domain.Fields{}
	add := func(name, value string) {
		value = strings.TrimSpace(value)
		if value == "" || value == "-" {
			return
		}
		fields = append(fields, domain.Fields{Name: name, Value: value})
	}

	add(common.IDfnal, fmt.Sprintf("%d", fproStavke.IDFnal))
	add("idfkpl", fmt.Sprintf("%d", fproStavke.IDFkpl))
	add("rbr", fmt.Sprintf("%d", fproStavke.Rbr))
	add("tipdok", fproStavke.Tipdok)
	add("nalog", fproStavke.Nalog)
	add("danal", fproStavke.Danal)
	add("opis", fproStavke.Opisknj)
	add("brst", fproStavke.Brst)
	add("konto", fproStavke.Konto)
	add("sifra", fproStavke.Sifra)
	add("vkonta", fproStavke.Vkonta)
	add("vrd", fproStavke.Vrd)

	duguje := common.StringToFloat64(fproStavke.Duguje)
	potrazuje := common.StringToFloat64(fproStavke.Potrazuje)
	if duguje > 0 {
		add("kat", "1")
		add("iznos", fmt.Sprintf("%.2f", duguje))
	} else if duguje < 0 {
		add("kat", "2")
		add("iznos", fmt.Sprintf("%.2f", -math.Abs(duguje)))
	}
	if potrazuje > 0 {
		add("kat", "3")
		add("iznos", fmt.Sprintf("%.2f", potrazuje))
	} else if potrazuje < 0 {
		add("kat", "4")
		add("iznos", fmt.Sprintf("%.2f", -math.Abs(potrazuje)))
	}

	add("dokum", fproStavke.Dokum)
	add("dadok", fproStavke.Dadok)
	add("rok", fproStavke.Rok)
	add("tra", fproStavke.Tra)
	add("dokumv", fproStavke.Dokvezni)
	add("dadokv", fproStavke.Dadokv)
	add("travez", fproStavke.Travez)
	add("idorgjed", fproStavke.IDorgjed)
	add("mestotrid", fproStavke.Mestotrid)
	add("magaciniid", fproStavke.Magaciniid)
	add("komid", fproStavke.Komid)
	add("fispid", fproStavke.Fispid)
	add("idvalute", fproStavke.IDValute)
	add("kurs", fproStavke.Kurs)
	add("deviznos", fproStavke.Deviznos)

	return fields
}

// GetOrgJedinice fetches the list of orgjed options for filtering.
func (s *FproResource) GetOrgJedinice(ctx context.Context) ([]domain.ComboItem, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found
	hasGod, hasKar := s.ojRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select idorgjed, ojozn, naziv from orgjed`, true)
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("ojozn ASC")
	sqlQuery, args := qb.Build()
	ojEntites, err := s.ojRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, err
	}
	for _, oj := range *ojEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", oj.IDOrgjed),
			Value: fmt.Sprintf("%s - %s", oj.OjOzn, oj.Naziv),
		})
	}
	return comboItems, nil
}

// GetMestoTroska fetches the list of Mesto Troska options based on the provided idorgjed.
func (s *FproResource) GetMestoTroska(ctx context.Context, idorgjed int64) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found
	qb := common.NewQueryBuilder(`select mestotrid, mtroska, opis, idorgjed from mestotr`, true)
	hasGod, hasKar := s.mtroskaRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("idorgjed", idorgjed)
	qb.AddOrderBy("mtroska ASC")
	sqlQuery, args := qb.Build()
	mtroskaEntites, err := s.mtroskaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, err
	}
	for _, mtroska := range *mtroskaEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", mtroska.MestoTrID),
			Value: fmt.Sprintf("%s - %s", mtroska.Mtroska, mtroska.Opis),
		})
	}
	return comboItems, nil
}

// GetValute fetches the list of valute options for filtering.
func (s *FproResource) GetValute(ctx context.Context) ([]domain.ComboItem, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found
	hasGod, hasKar := s.valuteRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select idvalute, sifval, naziv from valute`, true)
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("sifval ASC")
	sqlQuery, args := qb.Build()
	valuteEntites, err := s.valuteRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, err
	}
	for _, valute := range *valuteEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", valute.IDValute),
			Value: fmt.Sprintf("%d - %s", valute.Sifval, valute.Naziv.String),
		})
	}
	return comboItems, nil
}

// GetKomercijalisti fetches the list of komercijalisti options for filtering.
func (s *FproResource) GetKomercijalisti(ctx context.Context) ([]domain.ComboItem, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found
	hasGod, hasKar := s.komercijalistiRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select komid, sifkom, imeprezime from komercijalisti`, true)
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("sifkom ASC")
	sqlQuery, args := qb.Build()
	komercijalistiEntites, err := s.komercijalistiRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, err
	}
	for _, komercijalista := range *komercijalistiEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", komercijalista.KomID),
			Value: fmt.Sprintf("%d - %s", komercijalista.Sifkom, komercijalista.ImePrezime),
		})
	}
	return comboItems, nil
}

// GetMagacini fetches the list of magacini
func (s *FproResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found
	hasGod, hasKar := s.magaciniRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select magaciniid, mag, opis from magacini`, true)
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("mag ASC")
	sqlQuery, args := qb.Build()
	magaciniEntites, err := s.magaciniRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, err
	}
	for _, magacin := range *magaciniEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", magacin.MagaciniID),
			Value: fmt.Sprintf("%d - %s", magacin.Mag, magacin.Opis),
		})
	}
	return comboItems, nil
}

// GetMI fetches the list of MI options for filtering.
func (s *FproResource) GetMI(ctx context.Context) ([]domain.ComboItem, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found
	hasGod, hasKar := s.komercijalistiRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select fispid, mi, naziv from fisp`, true)
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("mi ASC")
	sqlQuery, args := qb.Build()
	miEntites, err := s.miRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, err
	}
	for _, mi := range *miEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", mi.FispID),
			Value: fmt.Sprintf("%d - %s", mi.MI, mi.Naziv),
		})
	}
	return comboItems, nil
}

func (s *FproResource) GetNalogTotalValues(ctx context.Context, nalogTotal *domain.NalogTotalValues, idFnal int64) error {
	// Implementation for fetching nalog total values
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	var totDuguje, totPotrazuje, totSaldo float64
	qb := common.NewQueryBuilder(`select 
		coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) as duguje,
		coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as potrazuje,
		coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) - coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as saldo
		
		from fpro`, true)
	qb.AddEqual("idfnal", idFnal)
	sqlQuery, args := qb.Build()
	err := s.fproRepo.DB.QueryRowContext(ctx, sqlQuery, args...).Scan(&totDuguje, &totPotrazuje, &totSaldo)
	if err != nil {
		return err
	}
	// Convert float64 to string
	nalogTotal.Duguje = common.FormatNumberWithSystemLocale(totDuguje, 2)
	nalogTotal.Potrazuje = common.FormatNumberWithSystemLocale(totPotrazuje, 2)
	nalogTotal.Saldo = common.FormatNumberWithSystemLocale(totSaldo, 2)

	return nil
}

// FproValidate implements validation for Fpro entities.
func (s *FproResource) FproValidate(ctx context.Context, entity *domain.FproPayload) []domain.FieldError {
	var fieldErrors []domain.FieldError
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return append(fieldErrors, domain.FieldError{Field: "session", ErrorMessage: common.ErrMsgUserSessionNotFound})
	}

	if s.isGodinaZatvorena(ctx, userSession.SelectedGod, userSession.SelectedKar) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "session", ErrorMessage: "godina je zatvorena, nije moguće izvršiti izmene na stavci naloga"})
		return fieldErrors
	}
	if entity.Konto == "" || len(entity.Konto) <= 3 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: common.ErrMsgObavezanPodatak})
	}
	fkplEntity, exist := s.findFkpl(ctx, userSession.SelectedGod, userSession.SelectedKar, entity.Konto, entity.Sifra)
	if !exist {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "nepostojeći konto"})
	} else {
		entity.Vkonta = fmt.Sprintf("%d", fkplEntity.Vkonta)
		entity.IDFkpl = fkplEntity.IDFkpl
	}
	//TODO check how to implemet with tipovi analitike
	// if entity.Sifra != "" && !sifraExists(entity.Sifra) {
	// 	fieldErrors = append(fieldErrors, domain.FieldError{Field: "sifra", ErrorMessage: "nepostojeći analitički sifra"})
	// }
	if entity.Vrd != "10" && entity.Vrd != "20" && entity.Vrd != "30" && entity.Vrd != "40" && entity.Vrd != "80" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "vrd", ErrorMessage: "neispravna vrsta dokumenta"})
	}
	if entity.Vrd == "90" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "vrd", ErrorMessage: "ne možete izabrati ovu vrstu dokumenta"})
	}
	fkpl, foundFkpl := s.findFkpl(ctx, userSession.SelectedGod, userSession.SelectedKar, entity.Konto, entity.Sifra)
	if !foundFkpl {
		if entity.Sifra == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "nepostojeći subsintetički konto"})
		} else {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "sifra", ErrorMessage: "nepostojeća analitička sifra"})
		}
		return fieldErrors
	}

	if fkpl.Vkonta == 3 || common.StringToInt(entity.Vkonta) == 3 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "knjiženje na sintetičkim kontima nije dozvoljeno"})
	}
	entity.Vkonta = fmt.Sprintf("%d", fkpl.Vkonta)
	entity.IDFkpl = fkpl.IDFkpl
	if entity.Vrd != "80" && entity.Vrd != "90" {
		if len(entity.Sifra) < 2 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "sifra", ErrorMessage: "šifra mora biti bar dvocifrena"})
		}
	}
	duguje := common.StringToFloat64(entity.Duguje)
	potrazuje := common.StringToFloat64(entity.Potrazuje)

	if duguje != 0 && potrazuje != 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "duguje", ErrorMessage: "ne možete uneti iznos u polje duguje i potražuje u isto vreme"})
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "potrazuje", ErrorMessage: "ne možete uneti iznos u polje duguje i potražuje u isto vreme"})
	}
	if duguje == 0 && potrazuje == 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "duguje", ErrorMessage: "morate uneti vrednost u polje duguje ili polje potražuje"})
	}

	if fkpl.Devizni {
		if common.StringToFloat64(entity.Deviznos) == 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "devizniiznos", ErrorMessage: "devizni iznos ne može biti 0 za devizni konto"})
		}
		if !s.existsValuta(ctx, entity.IDValute) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "idvalute", ErrorMessage: "ne postoji valuta sa takvom šifrom"})
		}
		if common.StringToFloat64(entity.Kurs) <= 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "kurs", ErrorMessage: common.ErrMsgObavezanPodatak})
		}
	}

	if entity.Vrd == "10" || entity.Vrd == "20" || entity.Vrd == "30" || entity.Vrd == "40" {
		if entity.IDorgjed != "-" && !s.existsOrgjed(ctx, entity.IDorgjed) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "idorgjed", ErrorMessage: "niste izabrali organizacionu jedinicu"})
		}
		if entity.Dokum == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dokum", ErrorMessage: common.ErrMsgObavezanPodatak})
		}
		dadok, err := time.Parse(common.HtmlLayout, entity.Dadok)
		if err != nil || dadok.IsZero() {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dadok", ErrorMessage: common.ErrMsgObavezanPodatak})
		}
		if common.StringToInt(entity.Tra) == 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "tra", ErrorMessage: common.ErrMsgObavezanPodatak})
		}
		if (entity.Vrd == "10" && duguje < 0) || (entity.Vrd == "20" && potrazuje < 0) {
			if entity.Dokvezni == "" {
				fieldErrors = append(fieldErrors, domain.FieldError{Field: "dokvezni", ErrorMessage: common.ErrMsgObavezanPodatak})
			}
			if common.StringToInt(entity.Travez) == 0 {
				fieldErrors = append(fieldErrors, domain.FieldError{Field: "travez", ErrorMessage: common.ErrMsgObavezanPodatak})
			}
			if common.StringToDate(entity.Dadokv, "02.01.2006").IsZero() {
				fieldErrors = append(fieldErrors, domain.FieldError{Field: "dadokv", ErrorMessage: common.ErrMsgObavezanPodatak})
			}
		}
	}
	if entity.Magaciniid != "-" && !s.existsMagacin(ctx, entity.Magaciniid) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "magaciniid", ErrorMessage: "ne postoji magacin sa takvom šifrom"})
	}
	if entity.Komid != "-" && !s.existsKomercijalista(ctx, entity.Komid) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "komid", ErrorMessage: "ne postoji komercijalista sa takvom šifrom"})
	}
	if entity.Mestotrid != "-" && !s.existsMtroska(ctx, entity.Mestotrid) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "mestotrid", ErrorMessage: "ne postoji mesto troška sa takvom šifrom"})
	}
	if entity.Fispid != "-" && !s.existsMI(ctx, entity.Fispid) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "fispid", ErrorMessage: "ne postoji mi sa takvom šifrom"})
	}
	if entity.IDValute != "-" && !s.existsValuta(ctx, entity.IDValute) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "idvalute", ErrorMessage: "ne postoji valuta sa takvom šifrom"})
	}
	// if common.dadok, ok := parseHTMLDate(dadokStr); ok {
	// 	if userSession.SelectedGod-dadok.Year() > 10 {
	// 		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dadok", ErrorMessage: fmt.Sprintf("datum dokumenta ne može biti manji od 01.01.%d", userSession.SelectedGod-10)})
	// 	}
	// 	if dadok.Year() != userSession.SelectedGod {
	// 		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dadok", ErrorMessage: "datum dokumenta je različit od poslovne godine"})
	// 	}
	// }

	// if strings.Contains(dokum, ",") {
	// 	fieldErrors = append(fieldErrors, domain.FieldError{Field: "dokum", ErrorMessage: "nedozvoljen karakter u broju dokumenta: ,"})
	// }

	return fieldErrors
}

func (s *FproResource) isGodinaZatvorena(ctx context.Context, god, kar int) bool {
	qb := common.NewQueryBuilder("select god, kar, godzatv from fvr", true)
	qb.AddEqual("god", god)
	qb.AddEqual("kar", kar)
	sqlQuery, args := qb.Build()
	entities, err := s.fvrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil || entities == nil || len(*entities) == 0 {
		return false
	}
	return (*entities)[0].GodZatv
}

func (s *FproResource) findFkpl(ctx context.Context, god, kar int, konto, sifra string) (*domain.Fkpl, bool) {
	vkonta := 2
	var qb *common.QueryBuilder
	if sifra != "" {
		qb = common.NewQueryBuilder("select idfkpl, vkonta, konto, p.sifra, f.naziv, devizni, p.idpartneri, p.naziv as partnernaziv from partneri p ", true)
		qb.AddJoin(" inner join fkpl f on p.tipanalitikeid = f.tipanalitikeid and f.god = p.god and f.kar = p.kar and f.konto = $1 and f.vkonta = 2")
		qb.AddArgs(konto)
		qb.AddEqual("p.god", god)
		qb.AddEqual("p.kar", kar)
		qb.AddEqual("p.sifra", sifra)
	} else {
		qb = common.NewQueryBuilder("select idfkpl, god, kar, vkonta, konto, sifra, naziv, devizni from fkpl f", true)
		qb.AddEqual("f.god", god)
		qb.AddEqual("f.kar", kar)
		qb.AddEqual("f.konto", konto)
		qb.AddEqual("f.vkonta", vkonta)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil || entities == nil || len(*entities) == 0 {
		return nil, false
	}
	return &(*entities)[0], true
}

// existsmagacin checks if a magacin with the given ID exists in the database.
func (s *FproResource) existsMagacin(ctx context.Context, idMagacin string) bool {
	entity, err := s.magaciniRepo.GetByID(ctx, common.IDmagacin, common.StringToInt64(idMagacin))
	return err == nil && entity != nil && entity.MagaciniID != 0
}

// existsValuta checks if a valuta with the given ID exists in the database.
func (s *FproResource) existsValuta(ctx context.Context, idValute string) bool {
	entity, err := s.valuteRepo.GetByID(ctx, common.IDvalute, common.StringToInt64(idValute))
	return err == nil && entity != nil && entity.IDValute != 0
}

// existsOrgjed checks if an orgjed with the given ID exists in the database.
func (s *FproResource) existsOrgjed(ctx context.Context, idOrgjed string) bool {
	entity, err := s.ojRepo.GetByID(ctx, common.IDorgjed, common.StringToInt64(idOrgjed))
	return err == nil && entity != nil && entity.IDOrgjed != 0
}

// existsMestoTroska checks if a Mesto Troska with the given ID exists in the database.
func (s *FproResource) existsMtroska(ctx context.Context, idMestotr string) bool {
	entity, err := s.mtroskaRepo.GetByID(ctx, common.IDmestotr, common.StringToInt64(idMestotr))
	return err == nil && entity != nil && entity.MestoTrID != 0
}

// existsKomercijalista checks if a Komercijalista with the given ID exists in the database.
func (s *FproResource) existsKomercijalista(ctx context.Context, komID string) bool {
	entity, err := s.komercijalistiRepo.GetByID(ctx, common.IDkomercijalista, common.StringToInt64(komID))
	return err == nil && entity != nil && entity.KomID != 0
}

// existsMI checks if a MI with the given ID exists in the database.
func (s *FproResource) existsMI(ctx context.Context, miID string) bool {
	entity, err := s.miRepo.GetByID(ctx, common.IDmi, common.StringToInt64(miID))
	return err == nil && entity != nil && entity.FispID != 0
}

// GetTableStavkeFields  fetches all data required to render the Nalog Stavke list page.
func (s *FproResource) GetTableStavkeFields() []domain.Fields {
	if len(s.naloziStavkeTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStavkeTableFields
}

// GetTableNalogFields fetches all data required to render the Nalog list page.
func (s *FproResource) GetTableNalogFields() []domain.Fields {
	if len(s.naloziTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziTableFields
}

func (s *FproResource) setServiceFieldValues() {
	s.naloziTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni Broj", Width: "4"},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "6"},
		{Name: "nalog", Label: "Br. Naloga", Width: "12"},
		{Name: "danal", Label: "Datum naloga", Width: "12"},
		{Name: "opis", Label: "Opis", Width: "60"},
		{Name: "dug", Label: "Duguje", Width: "14"},
		{Name: "pot", Label: "Potrazuje", Width: "14"},
		{Name: "datob", Label: "Datum obrade", Width: "12"},
		{Name: "brst", Label: "Br.Stavki", Width: "5"},
		{Name: "nalsts", Label: "Status naloga", Width: "10"},
	}
	s.naloziStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni Broj", Width: "4", IncludeInTotals: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "6"},
		{Name: "sifra", Label: "Sifra", Width: "6"},
		{Name: "nazivkonta", Label: "Naziv Konta", Width: "60"},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "60"},
		{Name: "dug", Label: "Duguje", Width: "14", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pot", Label: "Potrazuje", Width: "14", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "dokum", Label: "Broj Dokumenta", Width: "12"},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "12"},
	}
}

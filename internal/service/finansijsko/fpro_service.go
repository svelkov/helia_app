package finansijsko

import (
	"context"
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"helia/internal/service"
	commonsvc "helia/internal/service/common"
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
	fin                     *repository.FinansijskoRepository
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
	commonSvc               commonsvc.CommonService
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
	commonSvc commonsvc.CommonService,
	fproIDFieldName string,
	naloziTableFields []domain.Fields,
	naloziStavkeTableFields []domain.Fields,
	cfg config.Config,
) *FproResource {
	rs := &FproResource{
		service:                 Service,
		fin:                     repository.NewFinansijskoRepository(fproRepo.DB),
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
		commonSvc:               commonSvc,
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

// GetByID implements NalogService.
func (s *FproResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Fpro, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found in context")
	}
	if !querybuilder.IsIdent(idField) {
		return nil, fmt.Errorf("invalid id column %q", idField)
	}
	b := querybuilder.New(`select fp.idfpro, fp.rbr, fp.nalog, fp.tipdok, fp.danal, fp.iznos, fp.kat, coalesce(fp.opis, '') as opis, fp.dadok,
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
	from fpro as fp`).
		AddJoin("left join fkpl fk on fk.idfkpl = fp.idfkpl").
		// The partner of the period of the stavka (joined by its šifra, not by idpartneri).
		AddJoin("left join partneri p on p.sifra = fp.sifra and p.tipanalitikeid = fk.tipanalitikeid and p.god = fp.god and p.kar = fp.kar").
		Where("fp."+idField+" = ?", idValue)
	// The id comes from the request: the stavka must also be of the period of the session.
	if err := s.service.Repo.ApplyFilters(ctx, b, "fp", nil, ""); err != nil {
		return nil, err
	}
	entities, err := s.service.Repo.Select(ctx, b)
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
	b := querybuilder.New(`select fp.idfpro, fp.rbr, fp.nalog, fp.tipdok, fp.danal, fp.iznos, fp.kat, coalesce(fp.opis, '') as opis, fp.dadok,
	fp.rok, fp.vrd, fp.vkonta, fp.konto, coalesce(fp.sifra, '') as sifra, fp.tra, fp.deviznos, fp.kurs,
	fp.sifval, fp.mi, coalesce(fp.dokum, '') as dokum, fp.idfnal,
	fp.idorgjed, fp.idfkpl, fp.komid, fp.mestotrid,
	coalesce(fp.dokumv, '') as dokumv, fp.dadokv,	fp.travez, fk.naziv as naziv,
	case when fp.kat in (1,2) then fp.iznos
		 else 0 end as dug,
	case when fp.kat in (3,4) then fp.iznos
		 else 0 end as pot
	from fpro as fp`).
		AddJoin("left join fkpl fk on fk.idfkpl = fp.idfkpl").
		Where("fp.idfnal = ?", fnalID)
	if searchText != "" {
		b.Where("CONCAT( fp.nalog, ' ', fp.tipdok, ' ', fp.danal, ' ', fp.opis, ' ', fp.dadok, ' ', fp.rok, ' ', fp.vrd, ' ', fp.vkonta, ' ', fp.konto, ' ', fp.sifra, ' ', fk.naziv) ILIKE ?",
			"%"+searchText+"%")
	}
	// The nalog id comes from the request: the stavke must also be of the period of the session.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fp", nil, ""); err != nil {
		return err
	}
	totalRecords, err := s.fproRepo.Count(ctx, b)
	if err != nil {
		return fmt.Errorf("failed to get total records for Fpro by fnalID: %w", err)
	}
	common.SetTableTotalRecords(tbl, totalRecords, pageSize)

	b.OrderBy("fp.rbr DESC").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
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
	// The repository fills god/kar from the session and the audit columns; the update is limited to
	// the period of the session (ErrNotFound for a stavka of another period).
	if fproStavke.IDFpro == 0 {
		if _, err = s.fproRepo.CreateWithTx(ctx, tx, &domain.Fpro{}, common.IDfpro, fields); err != nil {
			return fmt.Errorf("insert fpro failed: %w", err)
		}
	} else {
		if err = s.fproRepo.UpdateWithTx(ctx, tx, &domain.Fpro{}, common.IDfpro, fproStavke.IDFpro, fields); err != nil {
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
	// Delete the Fpro record (limited to the period of the session, like the GetByID above).
	if err = s.fproRepo.DeleteWithTx(ctx, tx, common.IDfpro, idFpro); err != nil {
		return fmt.Errorf("failed to delete Fpro: %w", err)
	}
	// delete kir / kpr of the stavka (there may be none, so a plain statement, not Delete)
	if fpro.Vrd == 10 {
		if err = s.fproRepo.ExecWithTx(ctx, tx, querybuilder.New("delete from kir").Where("idfpro = ?", fpro.IDFpro)); err != nil {
			return fmt.Errorf("failed to delete kir: %w", err)
		}
	}
	if fpro.Vrd == 20 {
		if err = s.fproRepo.ExecWithTx(ctx, tx, querybuilder.New("delete from kpr").Where("idfpro = ?", fpro.IDFpro)); err != nil {
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
	// Every ? of the SET is the same nalog: each placeholder takes its own argument.
	b := querybuilder.New(`update fnal set
		dug    = (select coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) from fpro where idfnal = ?),
		pot    = (select coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) from fpro where idfnal = ?),
		brst   = (select count(*) from fpro where idfnal = ?),
		nalsts = case when
			(select coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) from fpro where idfnal = ?) =
			(select coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) from fpro where idfnal = ?)
			then 'Slozen' else 'Neslozen' end`, idFnal, idFnal, idFnal, idFnal, idFnal).
		Where("idfnal = ?", idFnal)
	if err := s.fproRepo.ExecWithTx(ctx, tx, b); err != nil {
		return fmt.Errorf("failed to recalculate fnal (id=%d): %w", idFnal, err)
	}
	return nil
}

func (s *FproResource) getLastRbr(ctx context.Context, fnlID int64) (int64, error) {
	lastRbr, err := s.fin.FproZadnjiRbr(ctx, fnlID)
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
// GetOrgJedinice returns the organizacione jedinice of the current period (CommonService).
func (s *FproResource) GetOrgJedinice(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetOrgJedComboValues(ctx, commonsvc.WithEmptyOption())
}

// GetMestoTroska returns the mesta troška of one organizaciona jedinica (CommonService).
func (s *FproResource) GetMestoTroska(ctx context.Context, idorgjed int64) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMestoTroskaComboValues(ctx, idorgjed, commonsvc.WithEmptyOption())
}

// GetValute returns the valute of the current period (CommonService).
func (s *FproResource) GetValute(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetValuteComboValues(ctx, commonsvc.WithEmptyOption())
}

// GetKomercijalisti returns the komercijalisti of the current period (CommonService).
func (s *FproResource) GetKomercijalisti(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetKomercijalistiComboValues(ctx, commonsvc.WithEmptyOption())
}

// GetMagacini returns the magacini of the current period (CommonService).
func (s *FproResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinComboValues(ctx, commonsvc.WithEmptyOption())
}

// GetMI returns the mesta isporuke of the current period (CommonService).
func (s *FproResource) GetMI(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMestoIsporukeComboValues(ctx, commonsvc.WithEmptyOption())
}

func (s *FproResource) GetNalogTotalValues(ctx context.Context, nalogTotal *domain.NalogTotalValues, idFnal int64) error {
	// Implementation for fetching nalog total values
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	// The nalog id comes from the request: only the stavke of the period of the session.
	totDuguje, totPotrazuje, totSaldo, err := s.fin.FproSumeNaloga(ctx, idFnal)
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
	b := querybuilder.New("select god, kar, godzatv from fvr").
		Where("god = ?", god).
		Where("kar = ?", kar)
	entities, err := s.fvrRepo.Select(ctx, b)
	if err != nil || entities == nil || len(*entities) == 0 {
		return false
	}
	return (*entities)[0].GodZatv
}

func (s *FproResource) findFkpl(ctx context.Context, god, kar int, konto, sifra string) (*domain.Fkpl, bool) {
	vkonta := 2
	// konto and sifra are the key of the analitika (Where, not AddEqual: never skipped).
	var b *querybuilder.Builder
	if sifra != "" {
		b = querybuilder.New("select idfkpl, vkonta, konto, p.sifra, f.naziv, devizni, p.idpartneri, p.naziv as partnernaziv from partneri p").
			AddJoin("inner join fkpl f on p.tipanalitikeid = f.tipanalitikeid and f.god = p.god and f.kar = p.kar and f.konto = ? and f.vkonta = 2", konto).
			Where("p.god = ?", god).
			Where("p.kar = ?", kar).
			Where("p.sifra = ?", sifra)
	} else {
		b = querybuilder.New("select idfkpl, god, kar, vkonta, konto, sifra, naziv, devizni from fkpl f").
			Where("f.god = ?", god).
			Where("f.kar = ?", kar).
			Where("f.konto = ?", konto).
			Where("f.vkonta = ?", vkonta)
	}
	entities, err := s.fkplRepo.Select(ctx, b)
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

package finansijsko

import (
	"context"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"reflect"
	"time"
)

// KamateService defines the interface for operations related to Kamate (Interest Rates and Calculations).
type KamateService interface {
	GetTipoviKamate(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetKamatneStope(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetFormiranjeKamatnihListova(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.KamateParameters, searchText string) error
	GetObracunKamate(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odBrojaListe, doBrojaListe, podDatumom, searchText string) error
	GetTipKamateOptions(ctx context.Context, model *domain.Tkam) error
	ValidateKamatneStope(ctx context.Context, tkam *domain.Tkam, cAction string) ([]domain.FieldError, error)
	SaveKamatneStope(ctx context.Context, tkam *domain.Tkam, idValue int64, cAction string) error
	DeleteKamatneStope(ctx context.Context, id int64) error
	ValidateTipoviKamate(ctx context.Context, tkam *domain.Kam, cAction string) ([]domain.FieldError, error)
	SaveTipoviKamate(ctx context.Context, tkam *domain.Kam, idValue int64, cAction string) error
	GetTipoviKamateByID(ctx context.Context, id int64) (*domain.Kam, error)
	GetKamatneStopeByID(ctx context.Context, id int64) (*domain.Tkam, error)
	DeleteTipoviKamate(ctx context.Context, id int64) error
	GetKamatneStopeTableFields() []domain.Fields
	GetTipoviKamateTableFields() []domain.Fields
	GetFormiranjeListovaTableFields() []domain.Fields
	GetFormiranjeListovaPartneriTableFields() []domain.Fields
	GetObracunTableFields() []domain.Fields
}

// KamateResource implements the KamateService interface.
type KamateResource struct {
	fkplRepo                        *repository.BaseRepository[domain.Fkpl]
	fproRepo                        *repository.BaseRepository[domain.Fpro]
	kamRepo                         *repository.BaseRepository[domain.Kam]
	tkamRepo                        *repository.BaseRepository[domain.Tkam]
	tipoviKamateTableFields         []domain.Fields
	kamateStopeTableFields          []domain.Fields
	formiranjeListovaFields         []domain.Fields
	formiranjeListovaPartneriFields []domain.Fields
	obracunTableFields              []domain.Fields
}

func NewKamateService(
	fkplRepo *repository.BaseRepository[domain.Fkpl],
	kamRepo *repository.BaseRepository[domain.Kam],
	tkamRepo *repository.BaseRepository[domain.Tkam],
	fproRepo *repository.BaseRepository[domain.Fpro],
) *KamateResource {
	rs := &KamateResource{
		fkplRepo: fkplRepo,
		kamRepo:  kamRepo,
		tkamRepo: tkamRepo,
		fproRepo: fproRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetTipoviKamateTableFields returns the table field definitions for Tipovi kamate (Types of Interest)
func (s *KamateResource) GetTipoviKamateTableFields() []domain.Fields {
	return s.tipoviKamateTableFields
}

// GetKamatneStopeTableFields returns the table field definitions for Kamate (default Kamatne stope)
func (s *KamateResource) GetKamatneStopeTableFields() []domain.Fields {
	return s.kamateStopeTableFields
}

// GetFormiranjeListovaTableFields returns the table field definitions for Formiranje listova
func (s *KamateResource) GetFormiranjeListovaTableFields() []domain.Fields {
	return s.formiranjeListovaFields
}

func (s *KamateResource) GetFormiranjeListovaPartneriTableFields() []domain.Fields {
	return s.formiranjeListovaPartneriFields
}

// GetObracunTableFields returns the table field definitions for Obracun kamate
func (s *KamateResource) GetObracunTableFields() []domain.Fields {
	return s.obracunTableFields
}

func (s *KamateResource) GetKamatneStopeByID(ctx context.Context, id int64) (*domain.Tkam, error) {

	qb := common.NewQueryBuilder(`SELECT tkam.idtkam, tkam.tipkam, tkam.odd, tkam.dod, tkam.kst, kam.opis, kam.model FROM tkam `, true)
	qb.AddJoin(" LEFT JOIN kam on kam.idkam = tkam.idkam ")
	qb.AddEqual("idtkam", id)
	sqlQuery, args := qb.Build()
	entities, err := s.tkamRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	if entities != nil && len(*entities) > 0 {
		return &(*entities)[0], nil
	}
	return &domain.Tkam{}, nil
}

// GetKamatneStope retrieves data for Kamatne stope (Interest Rates)
func (s *KamateResource) GetKamatneStope(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.tkamRepo.GetHasGodHasKar()

	// Build query for Kamatne stope
	qb := common.NewQueryBuilder(`SELECT tkam.idtkam, tkam.tipkam, tkam.odd, tkam.dod, tkam.kst, kam.opis, kam.model FROM tkam `, true)
	qb.AddJoin(" LEFT JOIN kam on tkam.idkam = kam.idkam")
	if hasGod {
		qb.AddEqual("tkam.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("tkam.kar", session.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Tkam{}))
		qb.AddSearchConditions(s.GetKamatneStopeTableFields(), searchText)
	}

	qb.AddOrderBy("tipkam")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.tkamRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			opis, model := "", ""
			if entity.Opis != nil {
				opis = *entity.Opis
			}
			if entity.Model != nil {
				model = *entity.Model
			}
			fields := []string{
				fmt.Sprintf("%d", entity.Tipkam),
				opis,
				model,
				entity.Odd.Format(common.HtmlLayout),
				entity.Dod.Format(common.HtmlLayout),
				common.FormatNumberWithSystemLocale(entity.Kst, 2),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.Idtkam), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetKamatneStopeTableFields()

	return nil
}

// ValidateKamatneStope validates a Tkam (Kamatne Stope) record
func (s *KamateResource) ValidateKamatneStope(ctx context.Context, tkam *domain.Tkam, cAction string) ([]domain.FieldError, error) {
	fieldErrors := []domain.FieldError{}

	if cAction == common.ActionAdd {
		// Check if tipkam value already exists
		if tkam.Idkam == nil || *tkam.Idkam <= 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "idkam", ErrorMessage: "Morate uneti tip kamate!!!"})
		}
	}
	if tkam.Odd == nil || tkam.Odd.IsZero() {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "odd", ErrorMessage: "Obavezan podatak!"})
	}
	if tkam.Dod == nil || tkam.Dod.IsZero() {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dod", ErrorMessage: "Obavezan podatak!"})
	}
	if tkam.Kst <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "kst", ErrorMessage: "Kamatna stopa ne može biti negativna!!!"})
	}

	return fieldErrors, nil
}

// SaveKamatneStope saves a Tkam (Kamatne Stope) record - handles both create and update
func (s *KamateResource) SaveKamatneStope(ctx context.Context, tkam *domain.Tkam, idValue int64, cAction string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	if cAction == common.ActionAdd {
		// Create new record
		qb := common.NewQueryBuilder(`INSERT INTO tkam (god, kar, tipkam, odd, dod, kst, xopunos, xdatunosa) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING idtkam `, false)
		qb.AddArgs(userSession.SelectedGod, userSession.SelectedKar, tkam.Tipkam, tkam.Odd, tkam.Dod, tkam.Kst, userSession.UserName, time.Now())
		sqlQuery, args := qb.Build()

		err := s.tkamRepo.CreateUpdateCustom(ctx, sqlQuery, args...)
		if err != nil {
			return fmt.Errorf("error creating kamatne stope record: %w", err)
		}
	}
	if cAction == common.ActionUpdate {
		// Update existing record
		qb := common.NewQueryBuilder(`UPDATE tkam SET idkam = $1, tipkam = $2, odd = $3, 
		dod = $4, kst = $5, xopizmene = $6, xdatizmene = $7`, true)
		qb.AddArgs(tkam.Idkam, tkam.Tipkam, tkam.Odd, tkam.Dod, tkam.Kst, userSession.UserName, time.Now())
		qb.AddEqual("idtkam", idValue)
		sqlQuery, args := qb.Build()
		tx, err := s.tkamRepo.BeginTx()
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}
		_, err = tx.ExecContext(ctx, sqlQuery, args...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating kamatne stope record: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
		}
	}

	return nil
}

// DeleteKamatneStope deletes a Kam (Kamatne Stope) record
func (s *KamateResource) DeleteKamatneStope(ctx context.Context, id int64) error {

	err := s.tkamRepo.Delete(ctx, common.IDtkam, id)
	if err != nil {
		return fmt.Errorf("error deleting kamatne stope record: %w", err)
	}

	return nil
}

func (s *KamateResource) GetTipKamateOptions(ctx context.Context, model *domain.Tkam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	// Build query for Kamatne stope
	hasGod, hasKar := s.kamRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idkam, tipkam, opis, model FROM kam `, true)
	if hasGod {
		qb.AddEqual("kam.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kam.kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("tipkam")
	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.kamRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	for _, entity := range *entities {
		model.TipKamateOptions = append(model.TipKamateOptions, domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.Idkam),
			Value: fmt.Sprintf("%d - %s", entity.Tipkam, *entity.Opis),
		})
	}
	return nil
}

func (s *KamateResource) GetTipoviKamateByID(ctx context.Context, id int64) (*domain.Kam, error) {
	return s.kamRepo.GetByID(ctx, common.IDkam, id)
}

func (s *KamateResource) GetTipoviKamate(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.kamRepo.GetHasGodHasKar()

	// Build query for Kamatne stope
	qb := common.NewQueryBuilder(`SELECT idkam, tipkam, opis, model FROM kam `, true)
	if hasGod {
		qb.AddEqual("kam.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kam.kar", session.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Kam{}))
		qb.AddSearchConditions(s.GetKamatneStopeTableFields(), searchText)
	}

	qb.AddOrderBy("tipkam")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.kamRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%d", entity.Tipkam),
				*(entity.Opis),
				*(entity.Model),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.Idkam), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetTipoviKamateTableFields()

	return nil

}

// ValidateTipoviKamate validates a Kam (Tipovi Kamate) record
func (s *KamateResource) ValidateTipoviKamate(ctx context.Context, kam *domain.Kam, cAction string) ([]domain.FieldError, error) {
	fieldErrors := []domain.FieldError{}
	if cAction == common.ActionAdd {
		// Check if tipkam value already exists
		if kam.Tipkam <= 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "tipkam", ErrorMessage: "Morate uneti tip kamate!!!"})
		}
	}
	if kam.Opis == nil || *kam.Opis == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "opis", ErrorMessage: "Opis je obavezan!!!"})
	}
	if kam.Model == nil || (*kam.Model != "G" && *kam.Model != "M") {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "model", ErrorMessage: "Model kamate je obavezan!!!"})
	}

	return fieldErrors, nil
}

// SaveTipoviKamate saves a Tkam (Tipovi Kamate) record - handles both create and update
func (s *KamateResource) SaveTipoviKamate(ctx context.Context, kam *domain.Kam, idValue int64, cAction string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	if cAction == common.ActionAdd {
		// Create new record
		qb := common.NewQueryBuilder(`INSERT INTO kam (god, kar, tipkam, opis, model, xopunos, xdatunosa) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING idkam `, false)
		qb.AddArgs(userSession.SelectedGod, userSession.SelectedKar, kam.Tipkam, kam.Opis, kam.Model, userSession.UserName, time.Now())
		sqlQuery, args := qb.Build()

		err := s.kamRepo.CreateUpdateCustom(ctx, sqlQuery, args...)
		if err != nil {
			return fmt.Errorf("error creating tipovi kamate record: %w", err)
		}
	}
	if cAction == common.ActionUpdate {
		// Update existing record
		qb := common.NewQueryBuilder(`UPDATE kam SET tipkam = $1, opis = $2, 
		model = $3, xopizmene = $4, xdatizmene = $5`, true)
		qb.AddArgs(kam.Tipkam, kam.Opis, kam.Model, userSession.UserName, time.Now())
		qb.AddEqual("idkam", idValue)

		sqlQuery, args := qb.Build()
		tx, err := s.kamRepo.BeginTx()
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}
		_, err = tx.ExecContext(ctx, sqlQuery, args...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating tipovi kamate record: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
		}
	}
	return nil
}

// DeleteTipoviKamate deletes a Tkam (Tipovi Kamate) record
func (s *KamateResource) DeleteTipoviKamate(ctx context.Context, id int64) error {

	err := s.kamRepo.Delete(ctx, common.IDkam, id)
	if err != nil {
		return fmt.Errorf("error deleting tipovi kamate record: %w", err)
	}

	return nil
}

// GetFormiranjeLista retrieves data for Formiranje kamatnih listova (Forming Interest Lists)
func (s *KamateResource) GetFormiranjeKamatnihListova(ctx context.Context, tblPartneri *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.KamateParameters, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tblPartneri, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT f.konto, f.sifra, p.naziv, p.adresa, p.mesto
		FROM partneri p `, true)
	qb.AddJoin(" LEFT JOIN fpro f ON f.idpartneri = p.idpartneri ")
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", session.SelectedKar)
	}
	qb.AddIn("f.vrd", []any{"10", "30"})
	qb.AddEqual("f.konto", params.Konto)
	nbrArgs := qb.GetArgsCount()
	qb.AddCustomCondition(` (
        f.dadok IS NULL 
        OR f.rok IS NULL 
        OR (f.dadok + f.rok)::date >= $` + fmt.Sprintf("%d", nbrArgs+1) + `::date
    )
    AND (
        f.dadok IS NULL 
        OR f.rok IS NULL 
        OR (f.dadok + f.rok)::date <= $` + fmt.Sprintf("%d", nbrArgs+2) + `::date
    ) `)
	qb.AddCustomCondition(` (
        f.sifra ~ '^[0-9]+\.?[0-9]*$' 
        AND f.sifra::numeric BETWEEN $` + fmt.Sprintf("%d", nbrArgs+3) + ` AND $` + fmt.Sprintf("%d", nbrArgs+4) + `	
    )`)
	qb.AddArgs(params.OdDatuma, params.DoDatuma, params.OdSifre, params.DoSifre)
	if searchText != "" {
		qb.AddCustomSearchCondition([]string{"p.naziv", "p.mesto", "p.adresa"}, searchText)
	}
	qb.AddOrderBy("f.konto, f.sifra")
	qb.AddGroupBy("f.konto, f.sifra, p.naziv, p.adresa, p.mesto")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tblPartneri, len(*entities), pageSize)
		return nil
	}
	// Set table headers
	tblPartneri.Headers = s.GetFormiranjeListovaPartneriTableFields()

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				entity.Adresa,
				entity.Mesto,
				
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tblPartneri.Rows = append(tblPartneri.Rows, tblRow)
		}
	}

	return nil
}

// GetObracunKamate retrieves data for Obracun kamate (Interest Calculation)
func (s *KamateResource) GetObracunKamate(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odBrojaListe, doBrojaListe, podDatumom, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Obracun kamate", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()

	// Build query for Obracun kamate
	qb := common.NewQueryBuilder(`
		SELECT 
			fkpl.idfkpl, fkpl.konto, fkpl.sifra, fkpl.naziv,
			fkpl.god, fkpl.kar,
			0 as br_kam_lista,
			0 as broj_dokumenta,
			'1900-01-01'::date as datum_dokuma,
			0 as rok,
			'1900-01-01'::date as datum_rospesca,
			'1900-01-01'::date as od_datuma,
			'1900-01-01'::date as do_datuma,
			0 as osnova,
			0 as duguje,
			0 as potrazuje,
			0 as kamatna_stopa,
			0 as model_kamate,
			0 as koeficijent,
			0 as iznos_dana,
			0 as broj,
			0 as iznos
		FROM fkpl `, true)

	if hasGod {
		qb.AddEqual("fkpl.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fkpl.kar", session.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fkpl{}))
		qb.AddSearchConditions(s.GetObracunTableFields(), searchText)
	}
	qb.AddCondition("br_kam_lista", odBrojaListe, ">=")
	qb.AddCondition("br_kam_lista", doBrojaListe, "<=")
	qb.AddCondition("datum_dokuma", podDatumom, "<=")

	qb.AddOrderBy("fkpl.konto ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				"",
				"",
				"01/01/1900",
				"0",
				"01/01/1900",
				"01/01/1900",
				"01/01/1900",
				common.FormatNumberWithSystemLocale(entity.Kolicinski, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				"0",
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				"0",
				common.FormatNumberWithSystemLocale(0, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetObracunTableFields()

	return nil
}

// setServiceFieldValues initializes table field definitions for Kamate
func (s *KamateResource) setServiceFieldValues() {
	// Fields for Tipovi kamate (Types of Interest)
	s.tipoviKamateTableFields = []domain.Fields{
		{Name: "tipkam", Label: "Tipkam", Width: "12", Field: "", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "12", Field: "", SkipInSearch: false},
		{Name: "model", Label: "Model", Width: "12", Field: "", SkipInSearch: false},
	}
	// Fields for Kamatne stope (Interest Rates)
	s.kamateStopeTableFields = []domain.Fields{
		{Name: "tipkam", Label: "Tip kamate", Width: "12", Field: "", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "12", Field: "", SkipInSearch: false},
		{Name: "model", Label: "Model", Width: "12", Field: "", SkipInSearch: false},
		{Name: "odd", Label: "Stopa važi od datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "dod", Label: "Stopa važi do datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "kst", Label: "Kamatna stopa", Width: "12", Field: "", SkipInSearch: true},
	}

	// Fields for Formiranje listova (Forming Lists)
	s.formiranjeListovaFields = []domain.Fields{
		{Name: "dokum", Label: "Broj Dokumenta", Width: "12", Field: "", SkipInSearch: true},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "12", Field: "", SkipInSearch: true},
		{Name: "rok", Label: "Rok", Width: "10", Field: "", SkipInSearch: true},
		{Name: "dospece", Label: "Datum dospeca", Width: "12", Field: "", SkipInSearch: true},
		{Name: "zaduzenje", Label: "Zaduzenje", Width: "12", Field: "", SkipInSearch: true},
		{Name: "uplata", Label: "Uplata", Width: "12", Field: "", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "nalog", Label: "Broj Naloga", Width: "12", Field: "", SkipInSearch: true},
		{Name: "danal", Label: "Datum Naloga", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tzatv", Label: "TZATV", Width: "12", Field: "", SkipInSearch: true},
		{Name: "totv", Label: "TOTV", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tbrsi", Label: "TBRSI", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tipdok", Label: "Tip Dokumenta", Width: "12", Field: "", SkipInSearch: true},
	}
	s.formiranjeListovaPartneriFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "12", Field: "", SkipInSearch: false},
		{Name: "sifra", Label: "Sifra", Width: "12", Field: "", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv", Width: "12", Field: "", SkipInSearch: false},
		{Name: "adresa", Label: "Adresa", Width: "12", Field: "", SkipInSearch: false},
		{Name: "mesto", Label: "Mesto", Width: "12", Field: "", SkipInSearch: false},
	}
	// Fields for Obracun kamate (Interest Calculation)
	s.obracunTableFields = []domain.Fields{
		{Name: "br_kam_lista", Label: "Br. kam. lista", Width: "12", Field: "", SkipInSearch: true},
		{Name: "dokum", Label: "Broj Dokumenta", Width: "12", Field: "", SkipInSearch: true},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "12", Field: "", SkipInSearch: true},
		{Name: "rok", Label: "Rok.", Width: "10", Field: "", SkipInSearch: true},
		{Name: "datum_dospeca", Label: "Datum dospeca", Width: "12", Field: "", SkipInSearch: true},
		{Name: "oddat", Label: "Od Datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "doddat", Label: "Do Datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "osnova", Label: "Osnova", Width: "12", Field: "", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12", Field: "", SkipInSearch: true},
		{Name: "kamatna_stopa", Label: "Kamatna Stopa", Width: "12", Field: "", SkipInSearch: true},
		{Name: "model_kamate", Label: "Model Kamate", Width: "12", Field: "", SkipInSearch: true},
		{Name: "koeficijent", Label: "Koeficijent", Width: "12", Field: "", SkipInSearch: true},
		{Name: "iznos_dana", Label: "Iznos Dana", Width: "12", Field: "", SkipInSearch: true},
		{Name: "broj", Label: "Broj", Width: "12", Field: "", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "", SkipInSearch: true},
	}
}

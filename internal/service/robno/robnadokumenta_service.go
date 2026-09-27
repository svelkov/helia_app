package robno

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

// RobnoDokumentaService defines the operations of the "Robna dokumenta" option
// (unos / pregled / specifikacije / kontiranje / prepis robnih dokumenata i prikazi naloga).
//
// Implemented so far: the shared helpers, the combos and the tab "Unos dokumenta".
// The queries of the remaining tabs are added as the single tabs are implemented, e.g.:
//
//	GetPregledDokumenata(context.Context, *domain.TableData, domain.RobnoDokumentaParams) error
type RobnoDokumentaService interface {
	GetFvrData(context.Context) (domain.Fvr, error)
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	GetTipdokComboValues(context.Context) ([]domain.ComboItem, error)
	GetVrstaDokumentaComboValues(context.Context) ([]domain.ComboItem, error)

	// Tab 1 - Unos dokumenta
	GetUnosDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error
	GetUnosDokumentaTotal(ctx context.Context, total *domain.RobnoDokumentaTotal) error
	GetNextNalog(ctx context.Context, tipdok string) (int, error)
	GetByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Rnal, error)
	ValidateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) []domain.FieldError
	CreateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) (int64, error)
	UpdateUnosDokumenta(ctx context.Context, rnalID int64, params domain.RobnoDokumentaParams) error

	// Tab 2 - Pregled dokumenta (sub-tabs "Štampa" and "eFaktura")
	GetPregledStampa(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error
	GetPregledEFaktura(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error

	// Tab 4 - Kontiranje dokumenata (sub-tabs "Knjiženje dokumenata", "Pregled proknjiženih /
	// neproknjiženih dokumenata" and "Pregled proknjiženih / neproknjiženih dokumenata po
	// magacinima").
	GetKontiranjeKnjizenje(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error
	GetKontiranjePregled(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error
	GetKontiranjePoMagacinima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error

	// Tab 6 - Prikaz ukupne obrade
	GetPrikazUkupneObrade(ctx context.Context, tbl *domain.TableData, pageSize int) error

	// Tab 7 - Prikaz naloga
	GetPrikazNaloga(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error

	// Tab 8 - Prikaz dokumenata u nalogu (the parameters of the tab are the same selection of nalozi
	// as the "Prikaz naloga" tab, its rows are the robni dokumenti of those nalozi).
	GetPrikazDokumenataUNalogu(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error

	// Tab 9 - Prikaz dokumenata po operateru (the same robni dokumenti as tab 8, grouped by the
	// operater of the document).
	GetPrikazDokumenataPooperateru(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error

	GetUnosDokumentaTableFields() []domain.Fields
	GetPregledDokumentaTableFields() []domain.Fields
	GetPregledStampaTableFields() []domain.Fields
	GetPregledEFakturaTableFields() []domain.Fields
	GetSpecifikacijeDokumentaTableFields() []domain.Fields
	GetKontiranjeKnjizenjeTableFields() []domain.Fields
	GetKontiranjePregledTableFields() []domain.Fields
	GetKontiranjePoMagacinimaTableFields() []domain.Fields
	GetPrepisDokumentaTableFields() []domain.Fields
	GetPrikazUkupneObradeTableFields() []domain.Fields
	GetPrikazNalogaTableFields() []domain.Fields
	GetPrikazDokumenataUNaloguTableFields() []domain.Fields
	GetPrikazDokumenataPooperateruTableFields() []domain.Fields
}

type RobnoDokumentaResource struct {
	// rnalRepo, rnalTotalsRepo and rnalHeaderRepo are the repositories of the robni nalozi (rnal):
	// the grid of the "Unos dokumenta" tab, the totals behind its "Prikaz ukupne obrade" panel and
	// the header of a nalog used by the save of that tab.
	rnalRepo       repository.BaseRepository[domain.RobnoDokumentaDto]
	rnalTotalsRepo repository.BaseRepository[domain.RobnoDokumentaTotalsDto]
	rnalHeaderRepo repository.BaseRepository[domain.Rnal]
	// pregledStampaRepo and pregledEFakturaRepo are the repositories of the grids of the "Pregled
	// dokumenta" tab (they only carry the row type and the table of the query: rnal joined with rdok).
	pregledStampaRepo   repository.BaseRepository[domain.RobnoDokumentaDto]
	pregledEFakturaRepo repository.BaseRepository[domain.RobnoDokumentaDto]
	// kontiranjeRepo is the repository of the grids of the "Kontiranje dokumenata" tab (they all
	// read the robni dokumenti of the current period from rdok).
	kontiranjeRepo repository.BaseRepository[domain.RobnoDokumentaDto]
	// prikazUkupneObradeRepo is the repository of the grid of the "Prikaz ukupne obrade" tab (the
	// magacini of the current period with the totals of their robni nalozi).
	prikazUkupneObradeRepo repository.BaseRepository[domain.PrikazUkupneObradeDto]
	// prikazNalogaRepo is the repository of the grid of the "Prikaz naloga" tab (the robni nalozi
	// of the current period).
	prikazNalogaRepo repository.BaseRepository[domain.RobnoDokumentaDto]
	// prikazDokumenataUNaloguRepo is the repository of the grid of the "Prikaz dokumenata u nalogu"
	// tab (the robni dokumenti of the selected nalozi).
	prikazDokumenataUNaloguRepo repository.BaseRepository[domain.RobnoDokumentaDto]
	// prikazDokumenataPooperateruRepo is the repository of the grid of the "Prikaz dokumenata po
	// operateru" tab (the same robni dokumenti, grouped by their operater).
	prikazDokumenataPooperateruRepo repository.BaseRepository[domain.RobnoDokumentaDto]
	tipdokRepo                      repository.BaseRepository[domain.Tipdok]
	dokvrstaRepo                    repository.BaseRepository[domain.Dokvrsta]
	magRepo                         repository.BaseRepository[domain.Magacini]
	fvrRepo                         repository.BaseRepository[domain.Fvr]
	commonSvc                       commonsvc.CommonService

	// TODO: add the repositories needed by the remaining tabs (rdok - robni dokument,
	// rpro - robni promet, rsif - artikli, fkpl, ...). All the grids of the option share the row type
	// domain.RobnoDokumentaDto, so the repositories differ in the table of their queries only.

	unosDokumentaTableFields               []domain.Fields
	pregledDokumentaTableFields            []domain.Fields
	pregledStampaTableFields               []domain.Fields
	pregledEFakturaTableFields             []domain.Fields
	specifikacijeDokumentaTableFields      []domain.Fields
	kontiranjeKnjizenjeTableFields         []domain.Fields
	kontiranjePregledTableFields           []domain.Fields
	kontiranjePoMagacinimaTableFields      []domain.Fields
	prepisDokumentaTableFields             []domain.Fields
	prikazUkupneObradeTableFields          []domain.Fields
	prikazNalogaTableFields                []domain.Fields
	prikazDokumenataUNaloguTableFields     []domain.Fields
	prikazDokumenataPooperateruTableFields []domain.Fields
}

// NewRobnoDokumentaService creates the service of the "Robna dokumenta" option.
func NewRobnoDokumentaService(
	rnalRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	rnalTotalsRepo repository.BaseRepository[domain.RobnoDokumentaTotalsDto],
	rnalHeaderRepo repository.BaseRepository[domain.Rnal],
	pregledStampaRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	pregledEFakturaRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	kontiranjeRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	prikazUkupneObradeRepo repository.BaseRepository[domain.PrikazUkupneObradeDto],
	prikazNalogaRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	prikazDokumenataUNaloguRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	prikazDokumenataPooperateruRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	tipdokRepo repository.BaseRepository[domain.Tipdok],
	dokvrstaRepo repository.BaseRepository[domain.Dokvrsta],
	magRepo repository.BaseRepository[domain.Magacini],
	fvrRepo repository.BaseRepository[domain.Fvr],
	commonSvc commonsvc.CommonService,
) *RobnoDokumentaResource {
	s := &RobnoDokumentaResource{
		rnalRepo:                        rnalRepo,
		rnalTotalsRepo:                  rnalTotalsRepo,
		rnalHeaderRepo:                  rnalHeaderRepo,
		pregledStampaRepo:               pregledStampaRepo,
		pregledEFakturaRepo:             pregledEFakturaRepo,
		kontiranjeRepo:                  kontiranjeRepo,
		prikazUkupneObradeRepo:          prikazUkupneObradeRepo,
		prikazNalogaRepo:                prikazNalogaRepo,
		prikazDokumenataUNaloguRepo:     prikazDokumenataUNaloguRepo,
		prikazDokumenataPooperateruRepo: prikazDokumenataPooperateruRepo,
		tipdokRepo:                      tipdokRepo,
		dokvrstaRepo:                    dokvrstaRepo,
		magRepo:                         magRepo,
		fvrRepo:                         fvrRepo,
		commonSvc:                       commonSvc,
	}
	s.setTableFields()
	return s
}

// GetFvrData retrieves company (fvr) data of the current session (used for the report headers).
func (s *RobnoDokumentaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, &s.fvrRepo)
}

// GetMagacinComboValues returns the magacini of the current period (CommonService).
func (s *RobnoDokumentaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinComboValues(ctx)
}

// GetTipdokComboValues returns the vrste naloga of the current period keyed by the tipdok code
// (CommonService).
func (s *RobnoDokumentaResource) GetTipdokComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetTipdokComboValues(ctx)
}

// GetVrstaDokumentaComboValues returns the vrste dokumenata of the current period (CommonService).
func (s *RobnoDokumentaResource) GetVrstaDokumentaComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetVrstaDokumentaComboValues(ctx)
}

// GetUnosDokumenta returns the robni nalozi (rnal) of the selected vrsta naloga for the grid of
// the "Unos dokumenta" tab. When getTotalRecords is true only the number of records is set on the
// table (the query runs without LIMIT/OFFSET), otherwise the rows of the requested page are set.
func (s *RobnoDokumentaResource) GetUnosDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.unosDokumentaTableFields

	qb := common.NewQueryBuilder(`select rnalid, tipdok, nalog, magaciniid, mag, danal, opis, datob, brdo, brst, dug, pot from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	qb.AddCondition("rnal.tipdok", params.Tipdok, "=")
	if params.MagaciniID != 0 {
		qb.AddEqual("rnal.magaciniid", params.MagaciniID)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rnal.tipdok", "rnal.nalog", "rnal.opis", "rnal.oper"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.nalog desc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.rnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				strings.TrimSpace(entity.Tipdok),
				fmt.Sprintf("%d", entity.Nalog),
				magacinLabel(entity.Mag, entity.MagaciniID),
				nullDateLabel(entity.Danal),
				entity.Opis,
				nullDateLabel(entity.Datob),
				fmt.Sprintf("%d", entity.Brdo),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetUnosDokumentaTotal fills the "Prikaz ukupne obrade" panel: the totals of all robni nalozi
// of the current period (god/kar), like the legacy screen shows them.
func (s *RobnoDokumentaResource) GetUnosDokumentaTotal(ctx context.Context, total *domain.RobnoDokumentaTotal) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder(`select count(*) as uknaloga,
		coalesce(sum(brdo), 0)::int as ukdokumenata,
		coalesce(sum(brst), 0)::int as ukstavki,
		coalesce(sum(dug), 0) as ukduguje,
		coalesce(sum(pot), 0) as ukpotrazuje
		from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	sqlQuery, args := qb.Build()
	entities, err := s.rnalTotalsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	var row domain.RobnoDokumentaTotalsDto
	if len(*entities) > 0 {
		row = (*entities)[0]
	}
	total.UkNaloga = fmt.Sprintf("%d", row.UkNaloga)
	total.UkDokumenata = fmt.Sprintf("%d", row.UkDokumenata)
	total.UkStavki = fmt.Sprintf("%d", row.UkStavki)
	total.Duguje = common.FormatNumberWithSystemLocale(row.UkDuguje, 2)
	total.Potrazuje = common.FormatNumberWithSystemLocale(row.UkPotrazuje, 2)
	total.Saldo = common.FormatNumberWithSystemLocale(row.UkDuguje-row.UkPotrazuje, 2)
	return nil
}

// GetNextNalog returns the next broj naloga of the given vrsta naloga (CommonService).
func (s *RobnoDokumentaResource) GetNextNalog(ctx context.Context, tipdok string) (int, error) {
	nextNalog, err := s.commonSvc.GetNextRnalNalog(ctx, tipdok)
	if err != nil {
		return 0, err
	}
	return int(nextNalog), nil
}

// GetByTipdokNalog returns the header of the robni nalog (rnal) of the given vrsta naloga and broj
// naloga of the current period. A zero value is returned when the nalog does not exist yet (the
// caller then saves a new one).
func (s *RobnoDokumentaResource) GetByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Rnal, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Rnal{}, fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder(`select rnalid, tipdok, idtipdok, nalog, danal, datob, opis, magaciniid, mag, dug, pot, brdo, brst, oper from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	qb.AddCondition("rnal.tipdok", tipdok, "=")
	qb.AddEqual("rnal.nalog", nalog)
	sqlQuery, args := qb.Build()
	entities, err := s.rnalHeaderRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Rnal{}, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Rnal{}, nil
	}
	return (*entities)[0], nil
}

// ValidateUnosDokumenta validates the header of the nalog of the "Unos dokumenta" tab: the vrsta
// naloga, the broj naloga and both dates are obligatory and the vrsta naloga and the magacin have
// to exist in the current period.
func (s *RobnoDokumentaResource) ValidateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	if strings.TrimSpace(params.Tipdok) == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "tipdok", ErrorMessage: common.ErrMsgObavezanPodatak})
	} else if _, err := s.commonSvc.GetTipdokIDByCode(ctx, strings.TrimSpace(params.Tipdok)); err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "tipdok", ErrorMessage: common.ErrMsgNotFound})
	}
	if strings.TrimSpace(params.Nalog) == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "nalog", ErrorMessage: common.ErrMsgObavezanPodatak})
	} else if _, err := strconv.Atoi(strings.TrimSpace(params.Nalog)); err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "nalog", ErrorMessage: "broj naloga mora biti ceo broj"})
	}
	if _, err := common.StringToTimeWithLayout(params.Danal, common.HtmlLayout); err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "danal", ErrorMessage: common.ErrMsgObavezanPodatak})
	}
	if _, err := common.StringToTimeWithLayout(params.Datob, common.HtmlLayout); err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "datob", ErrorMessage: common.ErrMsgObavezanPodatak})
	}
	if params.MagaciniID > 0 {
		if _, err := s.magacinByID(ctx, params.MagaciniID); err != nil {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "magaciniid", ErrorMessage: common.ErrMsgNotFound})
		}
	}
	return fieldErrors
}

// CreateUnosDokumenta inserts the header of a new robni nalog (rnal). Only the columns of the
// header are written: the totals (dug, pot, brdo, brst) stay 0 until the documents of the nalog are
// saved, and god, kar, xopunos and xdatunosa are filled by the repository builder from the session.
func (s *RobnoDokumentaResource) CreateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) (int64, error) {
	fields, err := s.unosDokumentaFields(ctx, params, true)
	if err != nil {
		return 0, err
	}
	// The entity is needed for the detection of the god/kar columns of rnal; the values are taken
	// from the field list above.
	return s.rnalHeaderRepo.Create(ctx, &domain.Rnal{}, common.IDrnal, fields)
}

// UpdateUnosDokumenta saves the header of an existing robni nalog (rnal): the vrsta naloga and the
// broj naloga can be corrected together with the dates, the opis and the magacin. The totals are
// left untouched (they belong to the documents of the nalog) and xopizmene/xdatizmene are filled
// by the repository builder.
func (s *RobnoDokumentaResource) UpdateUnosDokumenta(ctx context.Context, rnalID int64, params domain.RobnoDokumentaParams) error {
	fields, err := s.unosDokumentaFields(ctx, params, false)
	if err != nil {
		return err
	}
	return s.rnalHeaderRepo.Update(ctx, &domain.Rnal{}, common.IDrnal, rnalID, fields)
}

// unosDokumentaFields returns the columns of the header of a robni nalog with the values of the
// form of the "Unos dokumenta" tab. The vrsta naloga is resolved to idtipdok and the magacin to its
// mag (both are what the documents of the nalog store) and god, kar, xopunos and xdatunosa are
// added by the repository builder from the session.
//
// When insert is true the columns of a new nalog are added: it starts without documents, so rbr and
// the totals are 0 and nalsts stays empty until the documents of the nalog are entered.
//
// TODO: the legacy save also fills rbr/nalsts of the header; they are set when the documents of the
// nalog (rdok/rpro) are implemented. The vrsta dokumenta (vrd) is not a column of the header - it
// selects the documents of the nalog.
func (s *RobnoDokumentaResource) unosDokumentaFields(ctx context.Context, params domain.RobnoDokumentaParams, insert bool) ([]domain.Fields, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	tipdok := strings.TrimSpace(params.Tipdok)
	idTipdok, err := s.commonSvc.GetTipdokIDByCode(ctx, tipdok)
	if err != nil {
		return nil, err
	}
	mag := 0
	if params.MagaciniID > 0 {
		magacin, err := s.magacinByID(ctx, params.MagaciniID)
		if err != nil {
			return nil, err
		}
		mag = magacin.Mag
	}
	fields := []domain.Fields{
		{Name: "tipdok", Value: tipdok},
		{Name: "idtipdok", Value: fmt.Sprintf("%d", idTipdok)},
		{Name: "nalog", Value: strconv.Itoa(common.StringToInt(params.Nalog))},
		{Name: "danal", Value: params.Danal},
		{Name: "datob", Value: params.Datob},
		{Name: "opis", Value: params.Opis},
		{Name: "magaciniid", Value: strconv.Itoa(params.MagaciniID)},
		{Name: "mag", Value: strconv.Itoa(mag)},
	}
	if !insert {
		return fields, nil
	}
	return append(fields,
		domain.Fields{Name: "rbr", Value: "0"},
		domain.Fields{Name: "dug", Value: "0"},
		domain.Fields{Name: "pot", Value: "0"},
		domain.Fields{Name: "brdo", Value: "0"},
		domain.Fields{Name: "brst", Value: "0"},
		domain.Fields{Name: "oper", Value: userSession.UserName},
		domain.Fields{Name: "nalsts", Value: ""},
		// pinalid links a robni nalog to the financial nalog it was transferred to; a new nalog has
		// no link yet, so NULL is written (the column default 0 does not exist in table pinal).
		domain.Fields{Name: "pinalid", IsNull: true},
	), nil
}

// magacinByID returns the magacin of the current period with the given magaciniid.
func (s *RobnoDokumentaResource) magacinByID(ctx context.Context, magaciniID int) (domain.Magacini, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Magacini{}, fmt.Errorf("no user session found")
	}
	hasGod, hasKar := s.magRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("select magaciniid, mag, opis from magacini", true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("magaciniid", magaciniID)
	sqlQuery, args := qb.Build()
	entities, err := s.magRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Magacini{}, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Magacini{}, fmt.Errorf("%s: magaciniid=%d", common.ErrMsgNotFound, magaciniID)
	}
	return (*entities)[0], nil
}

//
// Tab 2 - Pregled dokumenta (sub-tabs "Štampa" and "eFaktura")
//

// GetPregledStampa fills the grid of the "Štampa" sub-tab: the robni dokumenti (rdok) with the
// header of their nalog (rnal), the partner of the document (fkpl/partneri) and the data of the
// document, filtered by magacin, vrsta dokumenta and the range of the dates of the nalog.
//
// The rows are ordered by the broj naloga and the broj dokumenta (the order in which the legacy
// report prints them).
func (s *RobnoDokumentaResource) GetPregledStampa(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pregledStampaTableFields
	hasGod, hasKar := s.pregledStampaRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rnal.nalog,
			rnal.danal,
			rnal.datob,
			rdok.dokum,
			rdok.dadok,
			rdok.dop,
			rdok.datiz,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(rdok.iznos, 0) as iznos,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(fkpl.naziv, '') as naziv,
			coalesce(p.pib, '') as pib,
			coalesce(p.jbkjs, '') as jbkjs,
			coalesce(p.adresa, '') as adresa,
			coalesce(p.mesto, '') as mesto,
			av.dokum as avansdokum,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.tkonto, '') as tkonto
		from rnal`, true)
	qb.AddJoin("inner join rdok on rdok.rnalid = rnal.rnalid")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	// The broj avansnog racuna of a document is the broj dokumenta of the avans it was created from
	// (rdok.avansid points to that rdok row; it is 0 when the document has no avans).
	qb.AddJoin("left join rdok av on av.rdokid = rdok.avansid")
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.Vrd != "" {
		qb.AddCondition("rdok.vrd", params.Vrd, "=")
	}
	if params.OdDanal != "" {
		qb.AddCondition("rnal.danal", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rnal.danal", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rdok.dokiz", "rdok.fkto", "rdok.fana", "fkpl.naziv", "p.pib", "p.mesto"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.nalog, rdok.dokum")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.pregledStampaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullDateLabel(entity.Datob),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				nullDateLabel(entity.Dop),
				nullDateLabel(entity.Datiz),
				entity.Dokiz,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				entity.Pib,
				entity.Jbkjs,
				entity.Adresa,
				entity.Mesto,
				nullInt64Label(entity.AvansDokum),
				entity.Pornapomena,
				entity.Tkonto,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPregledEFaktura fills the grid of the "eFaktura" sub-tab: the robni dokumenti (rdok) of the
// given groups of documents (dokvrsta.grpdok) with their status in the eFaktura (SEF) system,
// filtered by the range of the dates of the documents. The list is a working list, so the newest
// documents come first.
func (s *RobnoDokumentaResource) GetPregledEFaktura(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pregledEFakturaTableFields
	hasGod, hasKar := s.pregledEFakturaRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			coalesce(rdok.status_salinv, '') as statussalinv,
			rdok.datum_stat_salinv as datumstatsalinv,
			coalesce(rdok.komentar, '') as komentar,
			coalesce(rdok.cirinvoiceid, '') as cirinvoiceid,
			coalesce(rdok.vatrecordingstatus, '') as vatrecordingstatus,
			rdok.datum_stat_indvat as datumstatindvat,
			rdok.magaciniid,
			rnal.nalog,
			rnal.danal,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(fkpl.naziv, '') as naziv,
			coalesce(rdok.iznos, 0) as iznos,
			rdok.dop,
			coalesce(rdok.dokiz, '') as dokiz,
			case when coalesce(rdok.sifval, 0) = 0 then ''
				else coalesce(rdok.sifval::text || ' - ' || valute.naziv, rdok.sifval::text) end as valuta,
			coalesce(rdok.kurs, 0) as kurs,
			rdok.rdokid,
			rdok.salesinvoiceid
		from rnal`, true)
	qb.AddJoin("inner join rdok on rdok.rnalid = rnal.rnalid")
	qb.AddJoin("inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	qb.AddJoin("left join valute on valute.sifval = rdok.sifval")
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if groups := grupeDokumenata(params.GrupeDokumenata); len(groups) > 0 {
		qb.AddIn("dokvrsta.grpdok", groups)
	}
	if params.OdDanal != "" {
		qb.AddCondition("rdok.dadok", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.dadok", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rdok.dokiz", "rdok.fkto", "rdok.fana", "fkpl.naziv", "rdok.status_salinv", "rdok.cirinvoiceid"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.nalog desc, rdok.dokum desc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.pregledEFakturaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.StatusSalinv,
				nullDateLabel(entity.DatumStatSalinv),
				entity.Komentar,
				entity.CirInvoiceID,
				entity.VatRecordingStatus,
				nullDateLabel(entity.DatumStatIndVat),
				nullInt64Label(entity.MagaciniID),
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				nullDateLabel(entity.Dop),
				entity.Dokiz,
				entity.Valuta,
				common.FormatNumberWithSystemLocale(entity.Kurs, 2),
				fmt.Sprintf("%d", entity.RdokID),
				nullFloat64Label(entity.SalesInvoiceID),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

//
// Tab 4 - Kontiranje dokumenata (sub-tabs "Knjiženje dokumenata", "Pregled proknjiženih /
// neproknjiženih dokumenata" and "Pregled proknjiženih / neproknjiženih dokumenata po magacinima")
//

// GetKontiranjeKnjizenje fills the grid of the "Knjiženje dokumenata" sub-tab: the robni dokumenti
// (rdok) selected with the parameters of the tab (vrsta naloga za knjiženje, vrsta dokumenta,
// magacin and the ranges of the broj naloga, the broj dokumenta and the datum naloga).
//
// The list is ordered by the broj naloga and the broj dokumenta, the order in which the documents
// are posted.
//
// TODO: the posting ("Knjiži") and the check of the balance ("Pr. ravnotežu") of the legacy screen
// work on this list; whether the legacy screen also hides the documents that are already posted
// (rdok.knjige_1 = 'D') is verified together with those actions.
func (s *RobnoDokumentaResource) GetKontiranjeKnjizenje(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.kontiranjeKnjizenjeTableFields
	hasGod, hasKar := s.kontiranjeRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(kontiranjeSelectQuery, true)
	addKontiranjeConditions(qb, params, hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddOrderBy("rdok.nalog, rdok.dokum")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.kontiranjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	return s.setKontiranjeRows(tbl, entities, getTotalRecords, pageSize)
}

// GetKontiranjePregled fills the grid of the "Pregled proknjiženih / neproknjiženih dokumenata"
// sub-tab: the robni dokumenti (rdok) of the selection, either the posted (rdok.knjige_1 = 'D') or
// the not posted ones, according to the radio buttons of the sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePregled(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	tbl.Headers = s.kontiranjePregledTableFields
	return s.getKontiranjePregledList(ctx, tbl, getTotalRecords, pageSize, currentPage, params, "rdok.nalog, rdok.dokum")
}

// GetKontiranjePoMagacinima fills the grid of the "Pregled proknjiženih / neproknjiženih dokumenata
// po magacinima" sub-tab: the same rows as GetKontiranjePregled, grouped by magacin (the legacy
// screen differs from the previous one by the magacin of the row and shows the vrsta naloga).
func (s *RobnoDokumentaResource) GetKontiranjePoMagacinima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	tbl.Headers = s.kontiranjePoMagacinimaTableFields
	return s.getKontiranjePregledList(ctx, tbl, getTotalRecords, pageSize, currentPage, params, "rdok.magaciniid, rdok.nalog, rdok.dokum")
}

// getKontiranjePregledList runs the query of the two "Pregled ..." sub-tabs of the "Kontiranje
// dokumenata" tab (they share the filters and the date of the rows, they differ in the order and in
// the columns).
func (s *RobnoDokumentaResource) getKontiranjePregledList(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams, orderBy string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.kontiranjeRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(kontiranjeSelectQuery, true)
	addKontiranjeConditions(qb, params, hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	addKontiranjeStatusCondition(qb, params.Proknjizen)
	qb.AddOrderBy(orderBy)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.kontiranjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	return s.setKontiranjeRows(tbl, entities, getTotalRecords, pageSize)
}

// knjigeProknjizen is the value of rdok.knjige_1 of a robni dokument that is already posted (the
// legacy screens keep 'D' = da / proknjižen and 'N' = ne / neproknjižen in the first of the four
// knjige flags of the document).
const knjigeProknjizen = "D"

// kontiranjeSelectQuery is the row of the grids of the "Kontiranje dokumenata" tab: the data of a
// robni dokument (rdok) shown by the three sub-tabs.
const kontiranjeSelectQuery = `
	select
		rdok.tipdok,
		rdok.nalog,
		rdok.danal,
		rdok.vrd,
		rdok.dokum,
		rdok.dadok,
		coalesce(rdok.iznos, 0) as iznos,
		coalesce(rdok.opis, '') as opis,
		coalesce(rdok.polje, '') as polje,
		rdok.rok,
		rdok.magaciniid,
		coalesce(rdok.knjige_1, '') as knjige1
	from rdok`

// addKontiranjeConditions adds the filters of the "Kontiranje dokumenata" tab: the current period,
// the magacin, the vrsta naloga za knjiženje, the vrsta dokumenta and the ranges of the broj
// naloga, the broj dokumenta and the datum naloga. An empty range is not filtered; a range that is
// not a number (or not a date) is ignored so that a partially typed form never breaks the query.
func addKontiranjeConditions(qb *common.QueryBuilder, params domain.RobnoDokumentaParams, hasGod, hasKar bool, god, kar int) {
	if hasGod {
		qb.AddEqual("rdok.god", god)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", kar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.Tipdok != "" {
		qb.AddEqual("rdok.tipdok", params.Tipdok)
	}
	if params.Vrd != "" {
		qb.AddCondition("rdok.vrd", params.Vrd, "=")
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	addNumberCondition(qb, "rdok.dokum", params.OdDokum, ">=")
	addNumberCondition(qb, "rdok.dokum", params.DoDokum, "<=")
	if params.OdDanal != "" {
		qb.AddCondition("rdok.danal", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.danal", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.tipdok", "rdok.dokum", "rdok.opis", "rdok.dokiz", "rdok.fkto", "rdok.fana"}, params.SearchText)
	}
}

// addKontiranjeStatusCondition adds the state of the posting selected with the radio buttons of the
// "Pregled proknjiženih / neproknjiženih dokumenata" sub-tabs. The legacy screens keep the state of
// every robni dokument in rdok.knjige_1 ('D' = proknjižen, 'N' = neproknjižen).
func addKontiranjeStatusCondition(qb *common.QueryBuilder, proknjizen string) {
	if proknjizen == knjigeProknjizen {
		qb.AddCustomCondition("rdok.knjige_1 = 'D'")
		return
	}
	qb.AddCustomCondition("coalesce(rdok.knjige_1, '') <> 'D'")
}

// addNumberCondition adds a condition of a whole number (the ranges of the grid are typed as text
// and are only applied when they are a valid number).
func addNumberCondition(qb *common.QueryBuilder, field, value, operator string) {
	if value == "" {
		return
	}
	if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
		return
	}
	qb.AddCondition(field, value, operator)
}

// setKontiranjeRows sets the total number of records or the rows of the page (the cells are built
// from the headers of the grid, so every sub-tab shows its own columns in its own order).
func (s *RobnoDokumentaResource) setKontiranjeRows(tbl *domain.TableData, entities *[]domain.RobnoDokumentaDto, getTotalRecords bool, pageSize int) error {
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		cells := make([]string, 0, len(tbl.Headers))
		for _, header := range tbl.Headers {
			cells = append(cells, kontiranjeCell(entity, header.Name))
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: cells, HasUpdate: false, HasDelete: false})
	}
	return nil
}

// kontiranjeCell renders one cell of the grids of the "Kontiranje dokumenata" tab.
func kontiranjeCell(entity domain.RobnoDokumentaDto, field string) string {
	switch field {
	case "tipdok":
		return entity.Tipdok
	case "nalog":
		return fmt.Sprintf("%d", entity.Nalog)
	case "danal":
		return nullDateLabel(entity.Danal)
	case "vrd":
		return nullInt64Label(entity.Vrd)
	case "dokum":
		return nullInt64Label(entity.Dokum)
	case "dadok":
		return nullDateLabel(entity.Dadok)
	case "iznos":
		return common.FormatNumberWithSystemLocale(entity.Iznos, 2)
	case "opis":
		return entity.Opis
	case "polje":
		return entity.Polje
	case "rok":
		return nullInt64Label(entity.Rok)
	case "magaciniid":
		return nullInt64Label(entity.MagaciniID)
	}
	return ""
}

// grupeDokumenata splits the "Grupe dokumenata" filter of the eFaktura sub-tab (a comma separated
// list of dokvrsta.grpdok codes) into the values of an IN condition.
func grupeDokumenata(value string) []any {
	groups := []any{}
	for _, part := range strings.Split(value, ",") {
		group := strings.ToUpper(strings.TrimSpace(part))
		if group == "" {
			continue
		}
		groups = append(groups, group)
	}
	return groups
}

// nullInt64Label renders a nullable number of the grid ("" when there is none).
func nullInt64Label(value sql.NullInt64) string {
	if !value.Valid {
		return ""
	}
	return fmt.Sprintf("%d", value.Int64)
}

// nullFloat64Label renders a nullable decimal number of the grid ("" when there is none).
func nullFloat64Label(value sql.NullFloat64) string {
	if !value.Valid {
		return ""
	}
	return common.FormatNumberWithSystemLocale(value.Float64, 0)
}

// magacinLabel renders the "Magacin" cell of the grid.
func magacinLabel(mag, magaciniID sql.NullInt64) string {
	if mag.Valid {
		return fmt.Sprintf("%d", mag.Int64)
	}
	if magaciniID.Valid {
		return fmt.Sprintf("%d", magaciniID.Int64)
	}
	return ""
}

// nullDateLabel renders a nullable date of the grid (dd.MM.yyyy).
func nullDateLabel(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(common.DateLayout)
}

//
// Tab 6 - Prikaz ukupne obrade
//

// prikazUkupneObradeQuery is the row of the grid of the "Prikaz ukupne obrade" tab: every magacin
// of the current period (god/kar) with the number of the robni nalozi (rnal) of that magacin and
// their totals. The totals of a magacin without a nalog are 0 (left join), like the legacy screen
// lists all magacini.
//
// The god/kar of the magacini are the first two parameters of the query; the sub query uses the
// same values, so it reads them once more as $1/$2.
const prikazUkupneObradeQuery = `
	select
		m.mag,
		coalesce(m.opis, '') as opis,
		coalesce(m.mesto, '') as mesto,
		coalesce(t.brojnaloga, 0) as brojnaloga,
		coalesce(t.ukdokumenata, 0) as ukdokumenata,
		coalesce(t.brstavki, 0) as brstavki,
		coalesce(t.duguje, 0) as duguje,
		coalesce(t.potrazuje, 0) as potrazuje
	from magacini m
	left join (
		select
			rnal.magaciniid,
			count(*) as brojnaloga,
			coalesce(sum(rnal.brdo), 0)::int as ukdokumenata,
			coalesce(sum(rnal.brst), 0)::int as brstavki,
			coalesce(sum(rnal.dug), 0) as duguje,
			coalesce(sum(rnal.pot), 0) as potrazuje
		from rnal
		where rnal.god = $1 and rnal.kar = $2
		group by rnal.magaciniid
	) t on t.magaciniid = m.magaciniid
	where m.god = $1 and m.kar = $2
	order by m.mag`

// GetPrikazUkupneObrade fills the grid of the "Prikaz ukupne obrade" tab: the totals of the robni
// nalozi of every magacin of the current period (the tab has no parameters of the selection and
// the whole list fits on one page).
func (s *RobnoDokumentaResource) GetPrikazUkupneObrade(ctx context.Context, tbl *domain.TableData, pageSize int) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, 1, pageSize)
	tbl.Headers = s.prikazUkupneObradeTableFields

	entities, err := s.prikazUkupneObradeRepo.GetAllCustom(ctx, prikazUkupneObradeQuery, "", []any{userSession.SelectedGod, userSession.SelectedKar}, "", "")
	if err != nil {
		return err
	}
	common.SetTableTotalRecords(tbl, len(*entities), pageSize)
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", entity.Mag),
				entity.Opis,
				entity.Mesto,
				fmt.Sprintf("%d", entity.BrojNaloga),
				fmt.Sprintf("%d", entity.UkDokumenata),
				fmt.Sprintf("%d", entity.BrStavki),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

//
// Tab 7 - Prikaz naloga
//

// GetPrikazNaloga fills the grid of the "Prikaz naloga" tab: the robni nalozi (rnal) of the
// current period selected with the parameters of the tab (magacin, range of the vrste naloga and
// range of the broj naloga). The filters "Po datumu naloga", "Po datumu obrade" and "Po
// operateru" are applied only when their checkbox is checked, like the legacy screen does.
//
// The rows are ordered by the vrsta naloga and the broj naloga.
func (s *RobnoDokumentaResource) GetPrikazNaloga(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazNalogaTableFields
	hasGod, hasKar := s.prikazNalogaRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rnal.rnalid,
			rnal.tipdok,
			rnal.nalog,
			rnal.danal,
			rnal.datob,
			coalesce(rnal.brdo, 0) as brdo,
			coalesce(rnal.brst, 0) as brst,
			coalesce(rnal.dug, 0) as dug,
			coalesce(rnal.pot, 0) as pot,
			coalesce(rnal.oper, '') as oper,
			rnal.magaciniid
		from rnal`, true)
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rnal.magaciniid", params.MagaciniID)
	}
	if params.OdVrd != "" {
		qb.AddCondition("rnal.tipdok", params.OdVrd, ">=")
	}
	if params.DoVrd != "" {
		qb.AddCondition("rnal.tipdok", params.DoVrd, "<=")
	}
	addNumberCondition(qb, "rnal.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rnal.nalog", params.DoNaloga, "<=")
	if params.ChkDatumNaloga {
		if params.OdDanal != "" {
			qb.AddCondition("rnal.danal", params.OdDanal, ">=")
		}
		if params.DoDanal != "" {
			qb.AddCondition("rnal.danal", params.DoDanal, "<=")
		}
	}
	if params.ChkDatumObrade {
		if params.OdDatob != "" {
			qb.AddCondition("rnal.datob", params.OdDatob, ">=")
		}
		if params.DoDatob != "" {
			qb.AddCondition("rnal.datob", params.DoDatob, "<=")
		}
	}
	if params.ChkOperator && params.Oper != "" {
		qb.AddLike("rnal.oper", params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rnal.tipdok", "rnal.nalog", "rnal.oper", "rnal.opis"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.tipdok, rnal.nalog")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.prikazNalogaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				fmt.Sprintf("%d", entity.RnalID),
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullDateLabel(entity.Datob),
				fmt.Sprintf("%d", entity.Brdo),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				entity.Oper,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrikazDokumenataUNalogu fills the grid of the "Prikaz dokumenata u nalogu" tab: the robni
// dokumenti (rdok) of the nalozi selected with the parameters of the tab (magacin, range of the
// vrste naloga and range of the broj naloga). The filters "Po datumu naloga", "Po datumu obrade"
// and "Po operateru" are applied only when their checkbox is checked, like the legacy screen.
//
// The rows are ordered by the vrsta naloga, the broj naloga and the broj dokumenta.
func (s *RobnoDokumentaResource) GetPrikazDokumenataUNalogu(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazDokumenataUNaloguTableFields
	hasGod, hasKar := s.prikazDokumenataUNaloguRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.vrd,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.brst, 0) as brst,
			coalesce(rdok.iznos, 0) as iznos,
			rdok.datob,
			coalesce(rdok.oper, '') as oper,
			rdok.magaciniid
		from rdok`, true)
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.OdVrd != "" {
		qb.AddCondition("rdok.tipdok", params.OdVrd, ">=")
	}
	if params.DoVrd != "" {
		qb.AddCondition("rdok.tipdok", params.DoVrd, "<=")
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	if params.ChkDatumNaloga {
		if params.OdDanal != "" {
			qb.AddCondition("rdok.danal", params.OdDanal, ">=")
		}
		if params.DoDanal != "" {
			qb.AddCondition("rdok.danal", params.DoDanal, "<=")
		}
	}
	if params.ChkDatumObrade {
		if params.OdDatob != "" {
			qb.AddCondition("rdok.datob", params.OdDatob, ">=")
		}
		if params.DoDatob != "" {
			qb.AddCondition("rdok.datob", params.DoDatob, "<=")
		}
	}
	if params.ChkOperator && params.Oper != "" {
		qb.AddLike("rdok.oper", params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.tipdok", "rdok.nalog", "rdok.dokum", "rdok.opis", "rdok.oper"}, params.SearchText)
	}
	qb.AddOrderBy("rdok.tipdok, rdok.nalog, rdok.dokum")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.prikazDokumenataUNaloguRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Vrd),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				nullDateLabel(entity.Datob),
				entity.Oper,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrikazDokumenataPooperateru fills the grid of the "Prikaz dokumenata po operateru" tab: the
// same robni dokumenti (rdok) of the current period as the "Prikaz dokumenata u nalogu" tab, but
// grouped by the operater of the document (the legacy screen prints them per operator). The panel of
// the tab has no magacin and no vrsta naloga selection, so those filters are not applied; the filters
// "Po datumu naloga", "Po datumu obrade" and "Po operateru" are applied only when their checkbox is
// checked, like the legacy screen does.
//
// The rows are ordered by the operater, the vrsta naloga, the broj naloga and the broj dokumenta.
func (s *RobnoDokumentaResource) GetPrikazDokumenataPooperateru(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazDokumenataPooperateruTableFields
	hasGod, hasKar := s.prikazDokumenataPooperateruRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.vrd,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.brst, 0) as brst,
			coalesce(rdok.iznos, 0) as iznos,
			rdok.datob,
			coalesce(rdok.oper, '') as oper,
			rdok.magaciniid
		from rdok`, true)
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	if params.ChkDatumNaloga {
		if params.OdDanal != "" {
			qb.AddCondition("rdok.danal", params.OdDanal, ">=")
		}
		if params.DoDanal != "" {
			qb.AddCondition("rdok.danal", params.DoDanal, "<=")
		}
	}
	if params.ChkDatumObrade {
		if params.OdDatob != "" {
			qb.AddCondition("rdok.datob", params.OdDatob, ">=")
		}
		if params.DoDatob != "" {
			qb.AddCondition("rdok.datob", params.DoDatob, "<=")
		}
	}
	if params.ChkOperator && params.Oper != "" {
		qb.AddLike("rdok.oper", params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.tipdok", "rdok.nalog", "rdok.dokum", "rdok.opis", "rdok.oper"}, params.SearchText)
	}
	qb.AddOrderBy("rdok.oper, rdok.tipdok, rdok.nalog, rdok.dokum")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.prikazDokumenataPooperateruRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Vrd),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				nullDateLabel(entity.Datob),
				entity.Oper,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// Table fields - one getter per tab.
func (s *RobnoDokumentaResource) GetUnosDokumentaTableFields() []domain.Fields {
	return s.unosDokumentaTableFields
}

func (s *RobnoDokumentaResource) GetPregledDokumentaTableFields() []domain.Fields {
	return s.pregledDokumentaTableFields
}

func (s *RobnoDokumentaResource) GetPregledStampaTableFields() []domain.Fields {
	return s.pregledStampaTableFields
}

func (s *RobnoDokumentaResource) GetPregledEFakturaTableFields() []domain.Fields {
	return s.pregledEFakturaTableFields
}

func (s *RobnoDokumentaResource) GetSpecifikacijeDokumentaTableFields() []domain.Fields {
	return s.specifikacijeDokumentaTableFields
}

func (s *RobnoDokumentaResource) GetKontiranjeKnjizenjeTableFields() []domain.Fields {
	return s.kontiranjeKnjizenjeTableFields
}

func (s *RobnoDokumentaResource) GetKontiranjePregledTableFields() []domain.Fields {
	return s.kontiranjePregledTableFields
}

func (s *RobnoDokumentaResource) GetKontiranjePoMagacinimaTableFields() []domain.Fields {
	return s.kontiranjePoMagacinimaTableFields
}

func (s *RobnoDokumentaResource) GetPrepisDokumentaTableFields() []domain.Fields {
	return s.prepisDokumentaTableFields
}

func (s *RobnoDokumentaResource) GetPrikazUkupneObradeTableFields() []domain.Fields {
	return s.prikazUkupneObradeTableFields
}

func (s *RobnoDokumentaResource) GetPrikazNalogaTableFields() []domain.Fields {
	return s.prikazNalogaTableFields
}

func (s *RobnoDokumentaResource) GetPrikazDokumenataUNaloguTableFields() []domain.Fields {
	return s.prikazDokumenataUNaloguTableFields
}

func (s *RobnoDokumentaResource) GetPrikazDokumenataPooperateruTableFields() []domain.Fields {
	return s.prikazDokumenataPooperateruTableFields
}

// setTableFields defines the grid columns of every tab of the "Robna dokumenta" option.
func (s *RobnoDokumentaResource) setTableFields() {
	// Tab 1 - "Unos dokumenta": one row per robni nalog (rnal), the same columns as the legacy grid.
	s.unosDokumentaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "8", Field: "rnal.tipdok", TextAlign: "center", Sortable: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "rnal.nalog", TextAlign: "right", Sortable: true},
		{Name: "mag", Label: "Magacin", Width: "7", Field: "rnal.mag", TextAlign: "center"},
		{Name: "danal", Label: "Datum naloga", Width: "9", Field: "rnal.danal", TextAlign: "center", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "22", Field: "rnal.opis", Sortable: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", Field: "rnal.datob", TextAlign: "center", Sortable: true},
		{Name: "brdo", Label: "Uk. broj dokumenata", Width: "10", Field: "rnal.brdo", TextAlign: "right", SkipInSearch: true},
		{Name: "brst", Label: "Br. stavki", Width: "7", Field: "rnal.brst", TextAlign: "right", SkipInSearch: true},
		{Name: "dug", Label: "Duguje", Width: "10", Field: "rnal.dug", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "pot", Label: "Potražuje", Width: "10", Field: "rnal.pot", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}

	// Tab 2 - "Pregled dokumenta": the grid is the same for the tab and for its first sub-tab
	// ("Štampa"), the eFaktura sub-tab has its own columns.
	s.pregledStampaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj naloga", Width: "7", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", TextAlign: "center"},
		{Name: "datob", Label: "Datum obrade", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dop", Label: "Datum dospeća", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "datiz", Label: "Datum izv. dokumenta", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "9"},
		{Name: "iznos", Label: "Iznos", Width: "9", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "fkto", Label: "Konto", Width: "6"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "naziv", Label: "Naziv", Width: "20"},
		{Name: "pib", Label: "PIB", Width: "8"},
		{Name: "jbkjs", Label: "JBKJS", Width: "7", SkipInSearch: true},
		{Name: "adresa", Label: "Adresa", Width: "16", SkipInSearch: true},
		{Name: "mesto", Label: "Mesto", Width: "10"},
		{Name: "avansdokum", Label: "Broj avansnog računa", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "pornapomena", Label: "Poreska napomena", Width: "14", SkipInSearch: true},
		{Name: "tkonto", Label: "Konto prodavnice", Width: "8", SkipInSearch: true},
	}
	s.pregledDokumentaTableFields = s.pregledStampaTableFields

	s.pregledEFakturaTableFields = []domain.Fields{
		{Name: "statussalinv", Label: "Status", Width: "9"},
		{Name: "datumstatsalinv", Label: "Datum statusa", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "komentar", Label: "Komentar", Width: "12", SkipInSearch: true},
		{Name: "cirinvoiceid", Label: "CIR", Width: "9"},
		{Name: "vatrecordingstatus", Label: "Status IndVAT", Width: "9", SkipInSearch: true},
		{Name: "datumstatindvat", Label: "Datum statusa PE", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "magaciniid", Label: "Magacin", Width: "6", TextAlign: "center", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "7", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", TextAlign: "center"},
		{Name: "fkto", Label: "Konto", Width: "6"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "naziv", Label: "Naziv", Width: "20"},
		{Name: "iznos", Label: "Iznos", Width: "9", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "dop", Label: "Datum dospeća", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "9"},
		{Name: "valuta", Label: "Valuta", Width: "8", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "rdokid", Label: "RDOKID", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "salesinvoiceid", Label: "Sales", Width: "10", TextAlign: "right", SkipInSearch: true},
	}

	// Tab 3 - "Specifikacije dokumenta"
	// TODO: columns of the document specifications.
	s.specifikacijeDokumentaTableFields = []domain.Fields{}

	// Tab 4 - "Kontiranje dokumenata": one row per robni dokument (rdok) with the data the legacy
	// screen shows before the posting, with the vrsta naloga and the datum naloga of the document.
	s.kontiranjeKnjizenjeTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "8", TextAlign: "center", Sortable: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", TextAlign: "right", Sortable: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos dokumenta", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "20"},
		{Name: "polje", Label: "POPDV polje", Width: "8", SkipInSearch: true},
		{Name: "rok", Label: "Rok plaćanja", Width: "7", TextAlign: "center", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", SkipInSearch: true},
	}

	// The two "Pregled ..." sub-tabs of the tab show the datum naloga next to the broj naloga and
	// the vrsta dokumenta; the "po magacinima" one adds the vrsta naloga (its rows are grouped by
	// the magacin of the document).
	s.kontiranjePregledTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj naloga", Width: "8", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos dokumenta", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "20"},
		{Name: "polje", Label: "POPDV polje", Width: "8", SkipInSearch: true},
		{Name: "rok", Label: "Rok plaćanja", Width: "7", TextAlign: "center", SkipInSearch: true},
	}
	s.kontiranjePoMagacinimaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "8", TextAlign: "center", Sortable: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos dokumenta", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "20"},
		{Name: "polje", Label: "POPDV polje", Width: "8", SkipInSearch: true},
		{Name: "rok", Label: "Rok plaćanja", Width: "7", TextAlign: "center", SkipInSearch: true},
	}

	// Tab 5 - "Prepis dokumenta"
	// TODO: columns of the document rewrite (prepis).
	s.prepisDokumentaTableFields = []domain.Fields{}

	// Tab 6 - "Prikaz ukupne obrade": one row per magacin of the current period with the totals of
	// its robni nalozi (rnal), like the legacy screen shows them.
	s.prikazUkupneObradeTableFields = []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8", TextAlign: "center", Sortable: true},
		{Name: "opis", Label: "Opis magacina", Width: "20", Sortable: true},
		{Name: "mesto", Label: "Mesto", Width: "12", Sortable: true},
		{Name: "brojnaloga", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true, IncludeInTotals: true},
		{Name: "ukdokumenata", Label: "Uk. broj dokumenata", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "brstavki", Label: "Br. stavki", Width: "8", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "duguje", Label: "Duguje", Width: "12", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}

	// Tab 7 - "Prikaz naloga": one row per robni nalog (rnal) with the totals it was saved with.
	s.prikazNalogaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "rnalid", Label: "ID", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "brdo", Label: "Broj dokumenata", Width: "9", TextAlign: "right", SkipInSearch: true},
		{Name: "brst", Label: "Broj stavki", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "dug", Label: "Duguje", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "pot", Label: "Potražuje", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "10"},
	}

	// Tab 8 - "Prikaz dokumenata u nalogu": one row per robni dokument (rdok) of the selected
	// nalozi, with the data of the document and of its nalog (the columns of the legacy screen).
	s.prikazDokumenataUNaloguTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Dokument", Width: "9", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "brst", Label: "Broj stavki", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "10"},
	}

	// Tab 9 - "Prikaz dokumenata po operateru": the same robni dokumenti as tab 8 (the legacy screen
	// prints them per operator, so the operater is the first column of its grid).
	s.prikazDokumenataPooperateruTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Dokument", Width: "9", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "brst", Label: "Broj stavki", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "10"},
	}
}

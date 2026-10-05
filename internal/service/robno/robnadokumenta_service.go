package robno

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"helia/i18n"
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
	GetUnosDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error
	GetUnosDokumentaTotal(ctx context.Context, total *domain.RobnoDokumentaTotal) error
	GetNextNalog(ctx context.Context, tipdok string) (int, error)
	GetByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Rnal, error)
	ValidateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) []domain.FieldError
	CreateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) (int64, error)
	UpdateUnosDokumenta(ctx context.Context, rnalID int64, params domain.RobnoDokumentaParams) error

	// Tab 2 - Pregled dokumenta (sub-tabs "Štampa" and "eFaktura")
	GetPregledStampa(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error
	GetPregledEFaktura(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error

	// Tab 3 - Specifikacije dokumenta (the grid is the one of the "Štampa" sub-tab of "Pregled
	// dokumenta"; the tab adds the ranges of the vrste naloga, of the broj naloga and of the broj
	// dokumenta and the state of the print).
	GetSpecifikacijeDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Tab 4 - Kontiranje dokumenata (sub-tabs "Knjiženje dokumenata", "Pregled proknjiženih /
	// neproknjiženih dokumenata" and "Pregled proknjiženih / neproknjiženih dokumenata po
	// magacinima").
	GetKontiranjeKnjizenje(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error
	GetKontiranjePregled(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error
	GetKontiranjePoMagacinima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error

	// Tab 6 - Prikaz ukupne obrade
	GetPrikazUkupneObrade(ctx context.Context, tbl *domain.TableData, getTotRecords bool, currentPage, pageSize int, printType string) error

	// Tab 7 - Prikaz naloga
	GetPrikazNaloga(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Tab 8 - Prikaz dokumenata u nalogu (the parameters of the tab are the same selection of nalozi
	// as the "Prikaz naloga" tab, its rows are the robni dokumenti of those nalozi).
	GetPrikazDokumenataUNalogu(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Tab 9 - Prikaz dokumenata po operateru (the robni dokumenti of tab 8 grouped by the operater of
	// the document, with the number of his documents and of their stavke and his duguje/potražuje).
	GetPrikazDokumenataPoOperateru(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Štampa fakture: the fakture of the selection ready to print (the report RobnoStampaFaktura, the
	// legacy ROB_RPT_STAMPA_FAKTURA) and the izdavalac (fvr) of the print.
	GetStampaFaktura(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)

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
	GetFaktureStavkeTableFields() []domain.Fields
	GetFaktureAvansiTableFields() []domain.Fields
}

type RobnoDokumentaResource struct {
	robnaDokRepo           repository.BaseRepository[domain.RobnoDokumentaDto]
	rnalTotalsRepo         repository.BaseRepository[domain.RobnoDokumentaTotalsDto]
	rnalHeaderRepo         repository.BaseRepository[domain.Rnal]
	prikazUkupneObradeRepo repository.BaseRepository[domain.PrikazUkupneObradeDto]
	tipdokRepo             repository.BaseRepository[domain.Tipdok]
	dokvrstaRepo           repository.BaseRepository[domain.Dokvrsta]
	magRepo                repository.BaseRepository[domain.Magacini]
	fvrRepo                repository.BaseRepository[domain.Fvr]
	commonSvc              commonsvc.CommonService

	// The queries of the štampa fakture (GetStampaFaktura): the stavke with the header of their
	// document, the avansi closed on the fakture, their rate and the izdavalac.
	stampaFakturaRepo      repository.BaseRepository[domain.RobnoStampaFakturaRowDto]
	stampaFakturaAvansRepo repository.BaseRepository[domain.RobnoStampaFakturaAvansDto]
	stampaFakturaRateRepo  repository.BaseRepository[domain.RobnoStampaFakturaRataDto]
	stampaFakturaFirmaRepo repository.BaseRepository[domain.RobnoStampaFakturaFirmaDto]

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
	prikazNalogaPrintTableFields           []domain.Fields
	prikazDokumenataUNaloguTableFields     []domain.Fields
	prikazDokumenataPooperateruTableFields []domain.Fields
	faktureStavkeTableFields               []domain.Fields
	faktureAvansiTableFields               []domain.Fields
}

// NewRobnoDokumentaService creates the service of the "Robna dokumenta" option.
func NewRobnoDokumentaService(
	robnaDokRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	rnalTotalsRepo repository.BaseRepository[domain.RobnoDokumentaTotalsDto],
	rnalHeaderRepo repository.BaseRepository[domain.Rnal],
	prikazUkupneObradeRepo repository.BaseRepository[domain.PrikazUkupneObradeDto],
	tipdokRepo repository.BaseRepository[domain.Tipdok],
	dokvrstaRepo repository.BaseRepository[domain.Dokvrsta],
	magRepo repository.BaseRepository[domain.Magacini],
	fvrRepo repository.BaseRepository[domain.Fvr],
	commonSvc commonsvc.CommonService,
	stampaFakturaRepo repository.BaseRepository[domain.RobnoStampaFakturaRowDto],
	stampaFakturaAvansRepo repository.BaseRepository[domain.RobnoStampaFakturaAvansDto],
	stampaFakturaRateRepo repository.BaseRepository[domain.RobnoStampaFakturaRataDto],
	stampaFakturaFirmaRepo repository.BaseRepository[domain.RobnoStampaFakturaFirmaDto],
) *RobnoDokumentaResource {
	s := &RobnoDokumentaResource{
		robnaDokRepo:           robnaDokRepo,
		rnalTotalsRepo:         rnalTotalsRepo,
		rnalHeaderRepo:         rnalHeaderRepo,
		prikazUkupneObradeRepo: prikazUkupneObradeRepo,
		tipdokRepo:             tipdokRepo,
		dokvrstaRepo:           dokvrstaRepo,
		magRepo:                magRepo,
		fvrRepo:                fvrRepo,
		commonSvc:              commonSvc,
		stampaFakturaRepo:      stampaFakturaRepo,
		stampaFakturaAvansRepo: stampaFakturaAvansRepo,
		stampaFakturaRateRepo:  stampaFakturaRateRepo,
		stampaFakturaFirmaRepo: stampaFakturaFirmaRepo,
	}
	s.setTableFields()
	return s
}

// GetFvrData retrieves company (fvr) data of the current session (report headers).
func (s *RobnoDokumentaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, &s.fvrRepo)
}

// GetMagacinComboValues returns the magacini of the current period (CommonService).
func (s *RobnoDokumentaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinComboValues(ctx)
}

// GetTipdokComboValues returns the vrste naloga of the current period keyed by the tipdok code.
func (s *RobnoDokumentaResource) GetTipdokComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetTipdokComboValues(ctx)
}

// GetVrstaDokumentaComboValues returns the vrste dokumenata of the current period (CommonService).
func (s *RobnoDokumentaResource) GetVrstaDokumentaComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetVrstaDokumentaComboValues(ctx)
}

// GetUnosDokumenta returns the robni nalozi (rnal) of the selected vrsta naloga for the grid.
func (s *RobnoDokumentaResource) GetUnosDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
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
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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

// GetUnosDokumentaTotal fills the "Prikaz ukupne obrade" panel of the tab.
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
	total.UkNaloga = common.FormatNumberWithSystemLocale(row.UkNaloga, 0)
	total.UkDokumenata = common.FormatNumberWithSystemLocale(row.UkDokumenata, 0)
	total.UkStavki = common.FormatNumberWithSystemLocale(row.UkStavki, 0)
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

// GetByTipdokNalog returns the header of the robni nalog of a vrsta naloga and broj naloga.
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

// ValidateUnosDokumenta validates the header of the "Unos dokumenta" form.
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

// CreateUnosDokumenta inserts the header of a new robni nalog (rnal).
func (s *RobnoDokumentaResource) CreateUnosDokumenta(ctx context.Context, params domain.RobnoDokumentaParams) (int64, error) {
	fields, err := s.unosDokumentaFields(ctx, params, true)
	if err != nil {
		return 0, err
	}
	// The entity is needed for the detection of the god/kar columns of rnal; the values are taken
	// from the field list above.
	return s.rnalHeaderRepo.Create(ctx, &domain.Rnal{}, common.IDrnal, fields)
}

// UpdateUnosDokumenta saves the header of an existing robni nalog (rnal).
func (s *RobnoDokumentaResource) UpdateUnosDokumenta(ctx context.Context, rnalID int64, params domain.RobnoDokumentaParams) error {
	fields, err := s.unosDokumentaFields(ctx, params, false)
	if err != nil {
		return err
	}
	return s.rnalHeaderRepo.Update(ctx, &domain.Rnal{}, common.IDrnal, rnalID, fields)
}

// unosDokumentaFields returns the columns of the header of a robni nalog (TODO: rbr/nalsts).
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

// Tab 2 - Pregled dokumenta (sub-tabs "Štampa" and "eFaktura")
// GetPregledStampa fills the grid of the "Štampa" sub-tab of "Pregled dokumenta": the robni
func (s *RobnoDokumentaResource) GetPregledStampa(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pregledStampaTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.datob,
			rdok.dokum,
			rdok.dadok,
			rdok.rok,
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
		from rdok`, true)
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	// The broj avansnog racuna of a document is the broj dokumenta of the avans it was created from
	// (rdok.avansid points to that rdok row; it is 0 when the document has no avans).
	qb.AddJoin("left join rdok av on av.rdokid = rdok.avansid")
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	if params.Vrd != "" {
		qb.AddCondition("rdok.vrd", params.Vrd, "=")
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.OdDanal != "" {
		qb.AddCondition("rdok.danal", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.danal", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rdok.dokiz", "rdok.fkto", "rdok.fana", "fkpl.naziv", "p.pib", "p.mesto"}, params.SearchText)
	}
	qb.AddOrderBy("rdok.danal desc, rdok.tipdok desc, rdok.nalog desc, rdok.dadok desc, rdok.dokum desc")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			// The id of the row is the rdokid of the document: the print of the sub-tab prints the
			// faktura of the selected row.
			ID: fmt.Sprintf("%d", entity.RdokID),
			Fields: []string{
				fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
				common.FormatNullTime(entity.Danal, common.DateLayout),
				common.FormatNullTime(entity.Datob, common.DateLayout),
				fmt.Sprintf("%d", entity.Dokum.Int64),
				common.FormatNullTime(entity.Dadok, common.DateLayout),
				common.AddDaysToNullTime(entity.Dadok, int(entity.Rok.Int64), common.DateLayout),
				common.FormatNullTime(entity.Datiz, common.DateLayout),
				entity.Dokiz,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				entity.Pib,
				entity.Jbkjs,
				entity.Adresa,
				entity.Mesto,
				fmt.Sprintf("%d", entity.AvansDokum.Int64),
				entity.Pornapomena,
				entity.Tkonto,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// Tab 3 - Specifikacije dokumenta
// GetSpecifikacijeDokumenta fills the rows of the "Specifikacije dokumenta" tab: one row per stavka
func (s *RobnoDokumentaResource) GetSpecifikacijeDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	isPrint := printType == common.TipStampePrint
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.specifikacijeDokumentaTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`
		select
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.vrd,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.rok, 0) as rok,
			case when coalesce(rdok.sifval, 0) = 0 then ''
				else coalesce(rdok.sifval::text || ' - ' || valute.naziv, rdok.sifval::text) end as valuta,
			coalesce(rpro.fkto, '') as fkto,
			coalesce(rpro.fana, '') as fana,
			coalesce(fkpl.naziv, '') as naziv,
			coalesce(rpro.iznos, 0) as iznos,
			round(coalesce(rpro.iznos, 0) * coalesce(rpro.rab, 0) / 100, 2) as rabat,
			round((coalesce(rpro.iznos, 0) - round(coalesce(rpro.iznos, 0) * coalesce(rpro.rab, 0) / 100, 2))
				* coalesce(ps.pp, 0) / 100, 2) as porez,
			coalesce(ps.pp, 0) as stopa,
			rdok.rnalid
		from rdok`, true)
	qb.AddJoin("inner join rpro on rpro.rdokid = rdok.rdokid")
	qb.AddJoin("inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rpro.fkto and fkpl.sifra = rpro.fana")
	qb.AddJoin("left join valute on valute.sifval = rdok.sifval")
	// The poreska stopa of the stavka (rpor.pp): the stopa of its "po" that was in force on the
	// datum of the document, like the other robno reports read it.
	qb.AddJoin(`left join lateral (select r.pp from rpor r
		where r.po = rpro.po and r.datum <= rdok.dadok
		order by r.datum desc limit 1) ps on true`)
	// The groups of the vrste dokumenta the report covers (the legacy FAK, PRE, DIR, ARU and FZR).
	qb.AddIn("dokvrsta.grpdok", []any{"FAK", "PRE", "DIR", "ARU", "FZR"})
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.Vrd != "" {
		qb.AddCondition("rpro.vrd", params.Vrd, "=")
	}
	if params.OdVrd != "" {
		qb.AddCondition("rdok.tipdok", params.OdVrd, ">=")
	}
	if params.DoVrd != "" {
		qb.AddCondition("rdok.tipdok", params.DoVrd, "<=")
	}
	if params.OdDanal != "" {
		qb.AddCondition("rdok.dadok", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.dadok", params.DoDanal, "<=")
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	addNumberCondition(qb, "rdok.dokum", params.OdDokum, ">=")
	addNumberCondition(qb, "rdok.dokum", params.DoDokum, "<=")
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rpro.fkto", "rpro.fana", "fkpl.naziv"}, params.SearchText)
	}
	// The rows of a nalog and of a document stay together (the bands of the report are the nalog and
	// the document), like the legacy ORDER BY RNALID, RDOKID.
	qb.AddOrderBy("rdok.rnalid, rdok.rdokid, rpro.rproid")
	if !getTotalRecords && !isPrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && !isPrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// The columns of the amounts of the report: the label of the bands is printed over the columns
	// before them and every amount in its own column.
	amountColumn := 0
	for i, header := range tbl.Headers {
		if header.Name == "iznos" {
			amountColumn = i
			break
		}
	}
	type amounts struct {
		iznos     float64
		rabat     float64
		porez     float64
		zanaplatu float64
	}
	type stopaAmounts struct {
		osnovica float64
		pdv      float64
	}
	bandRow := func(classRow, label string, sums amounts) domain.TableRow {
		cells := make([]string, len(tbl.Headers))
		cells[amountColumn] = label
		for i, value := range []float64{sums.iznos, sums.rabat, sums.porez, sums.zanaplatu} {
			if column := amountColumn + i; column < len(cells) {
				cells[column] = common.FormatNumberWithSystemLocale(value, 2)
			}
		}
		return domain.TableRow{ClassRow: classRow, Fields: cells}
	}

	rbr := 0
	currentNalog := int64(-1)
	nalogSums := amounts{}
	reportSums := amounts{}
	stopaSums := map[float64]*stopaAmounts{}
	var stopaOrder []float64
	flushNalog := func() {
		if currentNalog == -1 {
			return
		}
		tbl.Rows = append(tbl.Rows, bandRow("nalog-total", i18n.GetInstance().Label("Ukupno za nalog")+":", nalogSums))
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		if isPrint && entity.RnalID != currentNalog {
			flushNalog()
			currentNalog = entity.RnalID
			nalogSums = amounts{}
		}
		rbr++
		amountsOfStavka := amounts{
			iznos:     entity.Iznos,
			rabat:     entity.Rabat,
			porez:     entity.Porez,
			zanaplatu: entity.Iznos - entity.Rabat + entity.Porez,
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", rbr),
				entity.Tipdok,
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Vrd),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				nullInt64Label(entity.Rok),
				entity.Valuta,
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(amountsOfStavka.iznos, 2),
				common.FormatNumberWithSystemLocale(amountsOfStavka.rabat, 2),
				common.FormatNumberWithSystemLocale(amountsOfStavka.porez, 2),
				common.FormatNumberWithSystemLocale(amountsOfStavka.zanaplatu, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
		if !isPrint {
			continue
		}
		nalogSums.iznos += amountsOfStavka.iznos
		nalogSums.rabat += amountsOfStavka.rabat
		nalogSums.porez += amountsOfStavka.porez
		nalogSums.zanaplatu += amountsOfStavka.zanaplatu
		reportSums.iznos += amountsOfStavka.iznos
		reportSums.rabat += amountsOfStavka.rabat
		reportSums.porez += amountsOfStavka.porez
		reportSums.zanaplatu += amountsOfStavka.zanaplatu
		if _, found := stopaSums[entity.Stopa]; !found {
			stopaSums[entity.Stopa] = &stopaAmounts{}
			stopaOrder = append(stopaOrder, entity.Stopa)
		}
		stopaSums[entity.Stopa].osnovica += amountsOfStavka.iznos - amountsOfStavka.rabat
		stopaSums[entity.Stopa].pdv += amountsOfStavka.porez
	}
	if !isPrint {
		return nil
	}
	flushNalog()
	tbl.Rows = append(tbl.Rows, bandRow("report-total", i18n.GetInstance().Label("Ukupno za izvestaj")+":", reportSums))

	// The "Ambalaza" summary: the PDV of the selection per poreska stopa (the columns of the legacy
	// report are Stopa, Osnovica and PDV) and the total of the two amount columns.
	sort.Float64s(stopaOrder)
	totalOsnovica, totalPdv := 0.0, 0.0
	for _, stopa := range stopaOrder {
		sums := stopaSums[stopa]
		totalOsnovica += sums.osnovica
		totalPdv += sums.pdv
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "ambalaza",
			Fields: []string{
				fmt.Sprintf("%s%%", common.FormatNumberWithSystemLocale(stopa, 2)),
				common.FormatNumberWithSystemLocale(sums.osnovica, 2),
				common.FormatNumberWithSystemLocale(sums.pdv, 2),
			},
		})
	}
	tbl.Rows = append(tbl.Rows, domain.TableRow{
		ClassRow: "ambalaza-total",
		Fields: []string{
			i18n.GetInstance().Label("Ukupno") + ":",
			common.FormatNumberWithSystemLocale(totalOsnovica, 2),
			common.FormatNumberWithSystemLocale(totalPdv, 2),
		},
	})
	return nil
}

// GetPregledEFaktura fills the grid of the "eFaktura" sub-tab (status of the documents).
func (s *RobnoDokumentaResource) GetPregledEFaktura(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pregledEFakturaTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

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
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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

// GetKontiranjeKnjizenje fills the grid of the "Knjiženje dokumenata" sub-tab (TODO).
func (s *RobnoDokumentaResource) GetKontiranjeKnjizenje(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.kontiranjeKnjizenjeTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(kontiranjeSelectQuery, true)
	addKontiranjeConditions(qb, params, hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddOrderBy("rdok.nalog, rdok.dokum")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	return s.setKontiranjeRows(tbl, entities, getTotalRecords, pageSize)
}

// GetKontiranjePregled fills the grid of the "Pregled proknjiženih / neproknjiženih" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePregled(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	tbl.Headers = s.kontiranjePregledTableFields
	return s.getKontiranjePregledList(ctx, tbl, getTotalRecords, currentPage, pageSize, params, "rdok.nalog, rdok.dokum")
}

// GetKontiranjePoMagacinima fills the same grid grouped by the magacin of the document.
func (s *RobnoDokumentaResource) GetKontiranjePoMagacinima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	tbl.Headers = s.kontiranjePoMagacinimaTableFields
	return s.getKontiranjePregledList(ctx, tbl, getTotalRecords, currentPage, pageSize, params, "rdok.magaciniid, rdok.nalog, rdok.dokum")
}

// getKontiranjePregledList runs the shared query of the two "Pregled ..." sub-tabs.
func (s *RobnoDokumentaResource) getKontiranjePregledList(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, orderBy string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(kontiranjeSelectQuery, true)
	addKontiranjeConditions(qb, params, hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	addKontiranjeStatusCondition(qb, params.Proknjizen)
	qb.AddOrderBy(orderBy)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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

// addKontiranjeConditions adds the filters of the "Kontiranje dokumenata" tab.
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

// addKontiranjeStatusCondition adds the state of the posting (rdok.knjige_1).
func addKontiranjeStatusCondition(qb *common.QueryBuilder, proknjizen string) {
	if proknjizen == knjigeProknjizen {
		qb.AddCustomCondition("rdok.knjige_1 = 'D'")
		return
	}
	qb.AddCustomCondition("coalesce(rdok.knjige_1, '') <> 'D'")
}

// addNumberCondition adds a condition of a whole number (validated ranges only).
func addNumberCondition(qb *common.QueryBuilder, field, value, operator string) {
	if value == "" {
		return
	}
	if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
		return
	}
	qb.AddCondition(field, value, operator)
}

// setKontiranjeRows sets the total records or the rows of the page of the tab.
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

// grupeDokumenata splits the "Grupe dokumenata" filter into the values of an IN condition.
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

// nullFloat64Label renders a nullable decimal number of the grid ("" when none).
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

// GetPrikazUkupneObrade fills the grid of the "Prikaz ukupne obrade" tab (totals per magacin).
func (s *RobnoDokumentaResource) GetPrikazUkupneObrade(ctx context.Context, tbl *domain.TableData, getTotRecords bool, currentPage, pageSize int, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazUkupneObradeTableFields
	qb := common.NewQueryBuilder(`SELECT
			m.mag,
				COALESCE(m.opis, '') AS opis,
				COALESCE(m.mesto, '') AS mesto,
				COALESCE(n.brojnaloga, 0)::int AS brojnaloga,
				COALESCE(d.ukdokumenata, 0)::int AS ukdokumenata,
				COALESCE(s.brstavki, 0)::int AS brstavki,
				COALESCE(n.duguje, 0) AS duguje,
				COALESCE(n.potrazuje, 0) AS potrazuje
			FROM magacini m
			LEFT JOIN (
				SELECT
					r.magaciniid,
					COUNT(*) AS brojnaloga,
					SUM(r.dug) AS duguje,
					SUM(r.pot) AS potrazuje
				FROM rnal r
				GROUP BY r.magaciniid
			) n
				ON n.magaciniid = m.magaciniid
			LEFT JOIN (
				SELECT
					r.magaciniid,
					COUNT(DISTINCT d.rdokid) AS ukdokumenata
				FROM rnal r
				JOIN rdok d
					ON d.rnalid = r.rnalid
				GROUP BY r.magaciniid
			) d
				ON d.magaciniid = m.magaciniid
			LEFT JOIN (
				SELECT
					r.magaciniid,
					COUNT(*) AS brstavki
				FROM rnal r
				JOIN rdok d
					ON d.rnalid = r.rnalid
				JOIN rpro p
					ON p.rdokid = d.rdokid
				GROUP BY r.magaciniid
			) s
				ON s.magaciniid = m.magaciniid `, true)
	qb.AddEqual("m.god", userSession.SelectedGod)
	qb.AddEqual("m.kar", userSession.SelectedKar)
	if !getTotRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("m.mag")
	sqlQuery, args := qb.Build()
	entities, err := s.prikazUkupneObradeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	var ukupnoBrojNaloga, ukupnoDokumenata, ukupnoStavki int
	var ukupnoDuguje, ukupnoPotrazuje float64
	for _, entity := range *entities {
		ukupnoBrojNaloga += entity.BrojNaloga
		ukupnoDokumenata += entity.UkDokumenata
		ukupnoStavki += entity.BrStavki
		ukupnoDuguje += entity.Duguje
		ukupnoPotrazuje += entity.Potrazuje
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
	// The printed report closes with the "Ukupno" line; the print preview of the tab (TipStampePreview)
	// shows the rows of the magacini only.
	if printType == common.TipStampePrint && len(tbl.Rows) > 0 {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			IsGroupTotal: true,
			Fields: []string{
				i18n.GetInstance().Label("Ukupno"),
				"",
				"",
				fmt.Sprintf("%d", ukupnoBrojNaloga),
				fmt.Sprintf("%d", ukupnoDokumenata),
				fmt.Sprintf("%d", ukupnoStavki),
				common.FormatNumberWithSystemLocale(ukupnoDuguje, 2),
				common.FormatNumberWithSystemLocale(ukupnoPotrazuje, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// Tab 7 - Prikaz naloga
// GetPrikazNaloga fills the grid of the "Prikaz naloga" tab (robni nalozi).
func (s *RobnoDokumentaResource) GetPrikazNaloga(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazNalogaTableFields
	if printType == common.TipStampePrint {
		tbl.Headers = s.prikazNalogaPrintTableFields
	}
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

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
			coalesce(rnal.xopunos, '') as oper,
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
	// The three optional filters of the tab, each one applied only when its checkbox is on.
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
		qb.AddCustomSearchCondition([]string{"rnal.xopunos", "rnal.xopizmene"}, params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rnal.tipdok", "rnal.nalog", "rnal.xopunos", "rnal.opis"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.tipdok, rnal.nalog desc")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		if printType == common.TipStampePrint {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
					fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
					nullDateLabel(entity.Danal),
					nullDateLabel(entity.Datob),
					fmt.Sprintf("%d", entity.Brdo),
					fmt.Sprintf("%d", entity.Brst),
					common.FormatNumberWithSystemLocale(entity.Dug, 2),
					common.FormatNumberWithSystemLocale(entity.Pot, 2),
					entity.Oper,
				},
				HasUpdate: false,
				HasDelete: false})
		} else {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
					"🔽",
					fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
					nullDateLabel(entity.Danal),
					nullDateLabel(entity.Datob),
					fmt.Sprintf("%d", entity.Brdo),
					fmt.Sprintf("%d", entity.Brst),
					common.FormatNumberWithSystemLocale(entity.Dug, 2),
					common.FormatNumberWithSystemLocale(entity.Pot, 2),
					entity.Oper,
				},
				HasUpdate: false,
				HasDelete: false})
		}
	}
	return nil
}

// Tab 8 - Prikaz dokumenata u nalogu
// GetPrikazDokumenataUNalogu fills the grid of the "Prikaz dokumenata u nalogu" tab: the robni
func (s *RobnoDokumentaResource) GetPrikazDokumenataUNalogu(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazDokumenataUNaloguTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

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
			coalesce(nullif(rdok.xopunos, ''), rdok.xopunos, '') as oper,
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
	// The operater filter of the legacy screen looks in both the user of the insert (xopunos) and
	// the user of the last change (xopizmene) of the document.
	if params.ChkOperator && params.Oper != "" {
		qb.AddCustomSearchCondition([]string{"rdok.xopunos", "rdok.xopizmene"}, params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.tipdok", "rdok.nalog", "rdok.dokum", "rdok.opis", "rdok.xopunos", "rdok.xopizmene"}, params.SearchText)
	}
	qb.AddOrderBy("rdok.nalog, rdok.vrd, rdok.dokum")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
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

// Tab 9 - Prikaz dokumenata po operateru
// GetPrikazDokumenataPoOperateru fills the grid of the "Prikaz dokumenata po operateru" tab: the
// robni dokumenti (rdok) of the selection grouped by the operater of the document (its xopizmene
// when it was changed, else its xopunos), with the number of his documents and of their stavke and
// the sums of duguje/potražuje of his documents by the konta 'D'/'P' (dokvrsta.kodknj), like the
// legacy procedure builds its per-operater table. With printType = TipStampePrint the whole result
// is returned unpaginated (the print of the tab).
func (s *RobnoDokumentaResource) GetPrikazDokumenataPoOperateru(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazDokumenataPooperateruTableFields

	qb := common.NewQueryBuilder(`
		select
			coalesce(nullif(rdok.xopunos, ''), rdok.xopunos, '') as oper,
			count(*)::int as brdo,
			coalesce(sum(rdok.brst), 0)::int as brst,
			coalesce(sum(case when upper(coalesce(dokvrsta.kodknj, '')) = 'D' then coalesce(rdok.iznos, 0) else 0 end), 0) as dug,
			coalesce(sum(case when upper(coalesce(dokvrsta.kodknj, '')) = 'P' then coalesce(rdok.iznos, 0) else 0 end), 0) as pot
		from rdok
		left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd`, true)
	// The period of the session must scope the documents. The row type of this option has no
	// god/kar db tags (GetHasGodHasKar reports false), so the two conditions are added explicitly.
	qb.AddEqual("rdok.god", userSession.SelectedGod)
	qb.AddEqual("rdok.kar", userSession.SelectedKar)
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
	// The operater filter of the legacy screen looks in both the user of the insert (xopunos) and
	// the user of the last change (xopizmene) of the document.
	if params.ChkOperator && params.Oper != "" {
		qb.AddCustomSearchCondition([]string{"rdok.xopunos", "rdok.xopizmene"}, params.Oper)
	}
	qb.AddGroupBy("coalesce(nullif(rdok.xopunos, ''), rdok.xopunos, '')")
	qb.AddOrderBy("oper")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Oper,
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

// GetUnosDokumentaTableFields returns the grid columns of the "Unos dokumenta" tab.
func (s *RobnoDokumentaResource) GetUnosDokumentaTableFields() []domain.Fields {
	return s.unosDokumentaTableFields
}

// GetPregledDokumentaTableFields returns the grid columns of the "Pregled dokumenta" tab.
func (s *RobnoDokumentaResource) GetPregledDokumentaTableFields() []domain.Fields {
	return s.pregledDokumentaTableFields
}

// GetPregledStampaTableFields returns the grid columns of the "Štampa" sub-tab.
func (s *RobnoDokumentaResource) GetPregledStampaTableFields() []domain.Fields {
	return s.pregledStampaTableFields
}

// GetPregledEFakturaTableFields returns the grid columns of the "eFaktura" sub-tab.
func (s *RobnoDokumentaResource) GetPregledEFakturaTableFields() []domain.Fields {
	return s.pregledEFakturaTableFields
}

// GetSpecifikacijeDokumentaTableFields returns the (still empty) grid columns of the tab.
func (s *RobnoDokumentaResource) GetSpecifikacijeDokumentaTableFields() []domain.Fields {
	return s.specifikacijeDokumentaTableFields
}

// GetKontiranjeKnjizenjeTableFields returns the grid columns of the "Knjiženje" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjeKnjizenjeTableFields() []domain.Fields {
	return s.kontiranjeKnjizenjeTableFields
}

// GetKontiranjePregledTableFields returns the grid columns of the "Pregled" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePregledTableFields() []domain.Fields {
	return s.kontiranjePregledTableFields
}

// GetKontiranjePoMagacinimaTableFields returns the columns of the "po magacinima" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePoMagacinimaTableFields() []domain.Fields {
	return s.kontiranjePoMagacinimaTableFields
}

// GetPrepisDokumentaTableFields returns the (still empty) grid columns of the tab.
func (s *RobnoDokumentaResource) GetPrepisDokumentaTableFields() []domain.Fields {
	return s.prepisDokumentaTableFields
}

// GetPrikazUkupneObradeTableFields returns the columns of the "Prikaz ukupne obrade" tab.
func (s *RobnoDokumentaResource) GetPrikazUkupneObradeTableFields() []domain.Fields {
	return s.prikazUkupneObradeTableFields
}

// GetPrikazNalogaTableFields returns the grid columns of the "Prikaz naloga" tab.
func (s *RobnoDokumentaResource) GetPrikazNalogaTableFields() []domain.Fields {
	return s.prikazNalogaTableFields
}

// GetPrikazDokumenataUNaloguTableFields returns the columns of the "u nalogu" tab.
func (s *RobnoDokumentaResource) GetPrikazDokumenataUNaloguTableFields() []domain.Fields {
	return s.prikazDokumenataUNaloguTableFields
}

// GetPrikazDokumenataPooperateruTableFields returns the columns of the "po operateru" tab.
func (s *RobnoDokumentaResource) GetPrikazDokumenataPooperateruTableFields() []domain.Fields {
	return s.prikazDokumenataPooperateruTableFields
}

// GetFaktureStavkeTableFields returns the columns of the grid of the stavke of the "Fakture
// veleprodaje" screen.
func (s *RobnoDokumentaResource) GetFaktureStavkeTableFields() []domain.Fields {
	return s.faktureStavkeTableFields
}

// GetFaktureAvansiTableFields returns the columns of the grid of the avansi of the "Fakture
// veleprodaje" screen.
func (s *RobnoDokumentaResource) GetFaktureAvansiTableFields() []domain.Fields {
	return s.faktureAvansiTableFields
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

	// Tab 3 - "Specifikacije dokumenta": one row per stavka (rpro) of the robni dokumenti (rdok) of
	// the selection, with the columns of the legacy report RobSpecifikacijaFakture ("SPECIFIKACIJA
	// EKSTERNIH RACUNA") - the header of the document, its partner and the amounts of the stavka.
	s.specifikacijeDokumentaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "tipdok", Label: "Vrsta naloga", Width: "5", TextAlign: "center"},
		{Name: "nalog", Label: "Broj naloga", Width: "5", TextAlign: "right"},
		{Name: "danal", Label: "Datum naloga", Width: "7", TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "5", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokum.", Width: "5", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "7", TextAlign: "center"},
		{Name: "rok", Label: "Rok", Width: "4", TextAlign: "center", SkipInSearch: true},
		{Name: "valuta", Label: "Valuta", Width: "5", SkipInSearch: true},
		{Name: "fkto", Label: "Kupac", Width: "5"},
		{Name: "fana", Label: "Šifra", Width: "5"},
		{Name: "naziv", Label: "Naziv kupca", Width: "16"},
		{Name: "iznos", Label: "Iznos", Width: "8", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "7", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "porez", Label: "Porez", Width: "7", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "zanaplatu", Label: "Za naplatu", Width: "9", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}

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
		{Name: "detalji", Label: "Detalji", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "brdo", Label: "Broj dokumenata", Width: "9", TextAlign: "right", SkipInSearch: true},
		{Name: "brst", Label: "Broj stavki", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "dug", Label: "Duguje", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "pot", Label: "Potražuje", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "10"},
	}

	// The printed "Prikaz naloga" report has no "Detalji" column (the expand arrow is a screen-only
	// affordance): the same columns without it.
	s.prikazNalogaPrintTableFields = make([]domain.Fields, 0, len(s.prikazNalogaTableFields))
	for _, field := range s.prikazNalogaTableFields {
		if field.Name == "detalji" {
			continue
		}
		s.prikazNalogaPrintTableFields = append(s.prikazNalogaPrintTableFields, field)
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

	// Tab 9 - "Prikaz dokumenata po operateru": one row per operater (the operater of a document is
	// its xopizmene when it was changed, else its xopunos) with the number of his robni dokumenti and
	// of their stavke and the sums of duguje/potražuje of his documents (the legacy procedure groups
	// the documents by the operater and builds this table).
	s.prikazDokumenataPooperateruTableFields = []domain.Fields{
		{Name: "oper", Label: "Operater", Width: "16"},
		{Name: "nbrd", Label: "Broj dokumenata", Width: "10", TextAlign: "right"},
		{Name: "nbrst", Label: "Broj stavki", Width: "10", TextAlign: "right"},
		{Name: "xdug", Label: "Duguje", Width: "12", TextAlign: "right"},
		{Name: "xpot", Label: "Potražuje", Width: "12", TextAlign: "right"},
	}

	// "Fakture veleprodaje" screen: the grid of the stavke of the faktura and the grid of the avansi
	// that can be closed with it.
	s.faktureStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "24"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "magacinskacena", Label: "Magacinska cena", Width: "9", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", IncludeInTotals: true},
		{Name: "prodajnacena", Label: "Prodajna cena", Width: "9", TextAlign: "right"},
		{Name: "rabat", Label: "Rabat", Width: "6", TextAlign: "right"},
	}
	s.faktureAvansiTableFields = []domain.Fields{
		{Name: "trazi", Label: "Traži", Width: "5", SkipInSearch: true},
		{Name: "brdok", Label: "Broj dok.", Width: "8", SkipInSearch: true},
		{Name: "avansa", Label: "Avansa", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "datumavansa", Label: "Datum avansa", Width: "10", SkipInSearch: true},
		{Name: "iznosavansa", Label: "Iznos avansa", Width: "10", TextAlign: "right", SkipInSearch: true},
		{Name: "ostatakavansa", Label: "Ostatak avansa", Width: "10", TextAlign: "right", SkipInSearch: true},
		{Name: "zatvorenona", Label: "Iznos koji je zatv. na fakt.", Width: "12", TextAlign: "right", SkipInSearch: true},
	}
}

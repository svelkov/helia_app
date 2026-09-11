package robno

import (
	"context"
	"fmt"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
)

type RobnoStanjaService interface {
	GetStanjePojedinacnogArtikla(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaViseArtikala(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaViseArtikalaSifra(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaViseArtikalaGrupa(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaSubsintetickogKonta(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetSvodjenjeZaliha(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	GetTipDokComboValues(context.Context) ([]domain.ComboItem, error)
	GetOjComboValues(context.Context) ([]domain.ComboItem, error)
	GetMestoTroskaComboValues(context.Context, int) ([]domain.ComboItem, error)
	SetDefaultTableData(tbl *domain.TableData)
	GetPojedinacnogArtiklaTableFields() []domain.Fields
	GetViseArtikalaTableFields() []domain.Fields
	GetSubsintetiskogKontaTableFields() []domain.Fields
	GetSvodjenjeZalihaTableFields() []domain.Fields
}

type RobnoStanjaResource struct {
	rproRepo                       repository.BaseRepository[domain.RobnoStanjeDto]
	magRepo                        repository.BaseRepository[domain.Magacini]
	tipdokRepo                     repository.BaseRepository[domain.Tipdok]
	ojRepo                         repository.BaseRepository[domain.Orgjed]
	mestoTroskaRepo                repository.BaseRepository[domain.Mestotr]
	fvrRepo                        repository.BaseRepository[domain.Fvr]
	pojedinacniArtikalTableFields  []domain.Fields
	viseArtikalaTableFields        []domain.Fields
	subsintetickogKontaTableFields []domain.Fields
	svodjenjeZalihaTableFields     []domain.Fields
}

func NewRobnoStanjaService(rproRepo repository.BaseRepository[domain.RobnoStanjeDto], magRepo repository.BaseRepository[domain.Magacini], tipdokRepo repository.BaseRepository[domain.Tipdok], ojRepo repository.BaseRepository[domain.Orgjed], mestoTroskaRepo repository.BaseRepository[domain.Mestotr], fvrRepo repository.BaseRepository[domain.Fvr]) *RobnoStanjaResource {
	rs := &RobnoStanjaResource{
		rproRepo:        rproRepo,
		magRepo:         magRepo,
		tipdokRepo:      tipdokRepo,
		ojRepo:          ojRepo,
		mestoTroskaRepo: mestoTroskaRepo,
		fvrRepo:         fvrRepo,
	}
	rs.setTableFileds()
	return rs
}

func (s *RobnoStanjaResource) GetStanjePojedinacnogArtikla(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjePojedinacnogArtikla

	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikala(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikala

	return nil
}
func (s *RobnoStanjaResource) GetStanjaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaSubsintetickogKonta

	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikalaSifra(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikalaSifra

	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikalaGrupa(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikalaGrupa

	return nil
}

func (s *RobnoStanjaResource) GetSvodjenjeZaliha(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetSvodjenjeZaliha

	return nil
}
func (s *RobnoStanjaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.magRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select magaciniid, mag, opis from magacini", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("mag")
	sqlQuery, args := qb.Build()
	entites, err := s.magRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := make([]domain.ComboItem, len(*entites))
	for i, entity := range *entites {
		comboItems[i] = domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.MagaciniID),
			Value: fmt.Sprintf("%d - %s", entity.Mag, entity.Opis),
		}
	}
	return comboItems, nil
}
func (s *RobnoStanjaResource) GetTipDokComboValues(ctx context.Context) ([]domain.ComboItem, error) {

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select idtipdok, tipdok, opis from tipdok", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("tipdok")
	sqlQuery, args := qb.Build()
	entites, err := s.tipdokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := make([]domain.ComboItem, len(*entites))
	for i, entity := range *entites {
		comboItems[i] = domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.IDTipDok),
			Value: fmt.Sprintf("%s - %s", entity.TipDok, entity.Opis),
		}
	}
	return comboItems, nil
}

func (s *RobnoStanjaResource) GetOjComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.ojRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select idorgjed, ojozn, naziv from orgjed", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("ojozn")
	sqlQuery, args := qb.Build()
	entites, err := s.ojRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found

	for _, entity := range *entites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.IDOrgjed),
			Value: fmt.Sprintf("%s - %s", entity.OjOzn, entity.Naziv),
		})
	}
	return comboItems, nil
}

func (s *RobnoStanjaResource) GetMestoTroskaComboValues(ctx context.Context, idOrgjed int) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.mestoTroskaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select mestotrid, mtroska, opis from mestotr", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("idorgjed", idOrgjed)
	qb.AddOrderBy("mtroska")
	sqlQuery, args := qb.Build()
	entites, err := s.mestoTroskaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found

	for _, entity := range *entites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.MestoTrID),
			Value: fmt.Sprintf("%s - %s", entity.Mtroska, entity.Opis),
		})
	}
	return comboItems, nil
}

func (s *RobnoStanjaResource) GetPojedinacnogArtiklaTableFields() []domain.Fields {
	return s.pojedinacniArtikalTableFields
}
func (s *RobnoStanjaResource) GetViseArtikalaTableFields() []domain.Fields {
	return s.viseArtikalaTableFields
}
func (s *RobnoStanjaResource) GetSubsintetiskogKontaTableFields() []domain.Fields {
	return s.subsintetickogKontaTableFields
}
func (s *RobnoStanjaResource) GetSvodjenjeZalihaTableFields() []domain.Fields {
	return s.svodjenjeZalihaTableFields
}
func (s *RobnoStanjaResource) SetDefaultTableData(tbl *domain.TableData) {
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

func (s *RobnoStanjaResource) setTableFileds() {
	s.pojedinacniArtikalTableFields = []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "12"},
		{Name: "ulaz", Label: "Ulaz", Width: "14", TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "14", TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "14", TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "14", TextAlign: "right"},
	}
	s.viseArtikalaTableFields = []domain.Fields{}
	s.subsintetickogKontaTableFields = []domain.Fields{}
	s.svodjenjeZalihaTableFields = []domain.Fields{}
}

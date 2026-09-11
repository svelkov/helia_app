package robno

import (
	"context"
	"fmt"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
)

// RobnoKarticaService exposes the two inventory-card views used by the Robno module.
type RobnoKarticaService interface {
	GetKarticaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams) error
	GetKarticaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams) error
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	ValidacijaSubsintetickogKonta(params domain.RobnoKarticaParams) []domain.FieldError
	ValidacijaKarticaArtikla(params domain.RobnoKarticaParams) []domain.FieldError
	GetKarticaArtiklaTableFields() []domain.Fields
	GetKarticaSubsintetickogKontaTableFields() []domain.Fields
	GetKarticaSubsintetickogKontaStampaTableFields() []domain.Fields
	GetKarticaSubsintetickogKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.RobnoKarticaParams) error
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// RobnoKarticaResource reuses the established Promet queries while keeping the
// Robno handler independent from the larger Finance service interface.
type RobnoKarticaResource struct {
	robnoKarticaRepo                      *repository.BaseRepository[domain.RobnoStanjeDto]
	magRepo                               *repository.BaseRepository[domain.Magacini]
	fvrRepo                               *repository.BaseRepository[domain.Fvr]
	karticaArtiklaTableFields             []domain.Fields
	subsintetickaKarticaKontaTableFields  []domain.Fields
	subsintetickaKarticaStampaTableFields []domain.Fields
}

func NewRobnoKarticaService(robnoKarticaRepo *repository.BaseRepository[domain.RobnoStanjeDto], magRepo *repository.BaseRepository[domain.Magacini], fvrRepo *repository.BaseRepository[domain.Fvr]) *RobnoKarticaResource {
	rs := &RobnoKarticaResource{
		robnoKarticaRepo: robnoKarticaRepo,
		magRepo:          magRepo,
		fvrRepo:          fvrRepo,
	}
	rs.setTableFields()
	return rs
}

func (s *RobnoKarticaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}
func (s *RobnoKarticaResource) GetKarticaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams) error {
	// TODO: Implement GetArticleCard using s.service and s.robnoKarticaRepo
	return nil
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams) error {
	// TODO: Implement GetSubsyntheticCard using s.service and s.robnoKarticaRepo
	return nil
}

// GetKarticaSubsintetickogKontaStampa builds the print-friendly (unpaginated) data set for the
// Kartica subsintetičkog konta report.
func (s *RobnoKarticaResource) GetKarticaSubsintetickogKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.RobnoKarticaParams) error {
	tbl.Headers = s.subsintetickaKarticaStampaTableFields
	// TODO: Implement report data retrieval using s.robnoKarticaRepo and params
	return nil
}
func (s *RobnoKarticaResource) ValidacijaKarticaArtikla(params domain.RobnoKarticaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	//TODO Implement validation logic for Kartica Artikla fields
	return fieldErrors
}
func (s *RobnoKarticaResource) ValidacijaSubsintetickogKonta(params domain.RobnoKarticaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	if params.Konto == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "obavezan podatak"})
	}
	if params.Magacin <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "magacin", ErrorMessage: "obavezan podatak"})
	}
	if params.CbxDatum {
		if params.OdDanal == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatuma", ErrorMessage: "obavezan podatak"})
		}
		if params.DoDanal == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatuma", ErrorMessage: "obavezan podatak"})
		}
	}
	if params.OdDanal != "" && params.DoDanal != "" {
		odDanal, errOd := time.Parse("2006-01-02", params.OdDanal)
		doDanal, errDo := time.Parse("2006-01-02", params.DoDanal)
		if errOd != nil || errDo != nil {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatuma", ErrorMessage: "Neispravan format datuma"})
		} else if odDanal.After(doDanal) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatuma", ErrorMessage: "Neispravan opseg datuma naloga"})
		}
		if odDanal.After(doDanal) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatuma", ErrorMessage: "Neispravan opseg datuma naloga"})
		}
	}
	if params.CbxBrojNaloga && params.Nalozi == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "nalozi", ErrorMessage: "obavezan podatak"})
	}
	if params.CbxIznos {
		if params.OdIznosa > params.DoIznosa {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "odiznosa", ErrorMessage: "Neispravan opseg iznosa"})
		}
	}
	return fieldErrors
}

func (s *RobnoKarticaResource) GetKarticaArtiklaTableFields() []domain.Fields {
	return s.karticaArtiklaTableFields
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKontaTableFields() []domain.Fields {
	return s.subsintetickaKarticaKontaTableFields
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKontaStampaTableFields() []domain.Fields {
	return s.subsintetickaKarticaStampaTableFields
}
func (s *RobnoKarticaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.magRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select mag, opis from magacini", true)
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
			Key:   fmt.Sprintf("%d", entity.Mag),
			Value: fmt.Sprintf("%d - %s", entity.Mag, entity.Opis),
		}
	}
	return comboItems, nil
}
func (s *RobnoKarticaResource) setTableFields() {
	// Initialize the table fields for the RobnoKarticaResource
	s.karticaArtiklaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "5", Field: "nalog", SkipInSearch: true, TextAlign: "center"},
		{Name: "danal", Label: "Datum naloga", Width: "5", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "5", Field: "vrd", SkipInSearch: true, TextAlign: "center"},
		{Name: "brdo", Label: "Broj Dokumenta", Width: "5", Field: "brdo", SkipInSearch: true, TextAlign: "center"},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "5", Field: "dadok", SkipInSearch: true, TextAlign: "center"},
		{Name: "konto", Label: "Konto", Width: "5", Field: "konto", SkipInSearch: true, TextAlign: "center"},
		{Name: "sifra", Label: "Sifra", Width: "5", Field: "sifra", SkipInSearch: true, TextAlign: "center"},
		{Name: "naziv", Label: "Naziv", Width: "5", Field: "naziv", SkipInSearch: true, TextAlign: "center"},
		{Name: "cena", Label: "Kolicina", Width: "5", Field: "kolicina", SkipInSearch: true, TextAlign: "center"},
		{Name: "ulaz", Label: "Ulaz", Width: "5", Field: "ulaz", SkipInSearch: true, TextAlign: "center"},
		{Name: "izlaz", Label: "Izlaz", Width: "5", Field: "izlaz", SkipInSearch: true, TextAlign: "center"},
		{Name: "stanje", Label: "Stanje", Width: "5", Field: "stanje", SkipInSearch: true, TextAlign: "center"},
		{Name: "iznos", Label: "Iznos", Width: "5", Field: "iznos", SkipInSearch: true, TextAlign: "center"},
		{Name: "duguje", Label: "Duguje", Width: "5", Field: "duguje", SkipInSearch: true, TextAlign: "center"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "5", Field: "potrazuje", SkipInSearch: true, TextAlign: "center"},
		{Name: "saldo", Label: "Saldo", Width: "5", Field: "saldo", SkipInSearch: true, TextAlign: "center"},
		{Name: "fkto", Label: "Fkto", Width: "5", Field: "fkto", SkipInSearch: true, TextAlign: "center"},
		{Name: "fana", Label: "Napomena", Width: "5", Field: "fana", SkipInSearch: true, TextAlign: "center"},
		{Name: "fkplnaz", Label: "Naziv konta", Width: "5", Field: "fkplnaz", SkipInSearch: true, TextAlign: "center"},
		{Name: "valuta", Label: "Valuta", Width: "5", Field: "valuta", SkipInSearch: true, TextAlign: "center"},
		{Name: "kurs", Label: "Kurs", Width: "5", Field: "kurs", SkipInSearch: true, TextAlign: "center"},
		{Name: "cenaval", Label: "Cena u valuti", Width: "5", Field: "cenaval", SkipInSearch: true, TextAlign: "center"},
		{Name: "mesec", Label: "Mesec", Width: "5", Field: "mesec", SkipInSearch: true, TextAlign: "center"},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "5", Field: "dokiz", SkipInSearch: true, TextAlign: "center"},
		{Name: "dadokiz", Label: "Datum izvornog dokumenta", Width: "5", Field: "dadokiz", SkipInSearch: true, TextAlign: "center"},
		{Name: "otk", Label: "OTK", Width: "5", Field: "otk", SkipInSearch: true, TextAlign: "center"},
		{Name: "serija", Label: "Serija", Width: "5", Field: "serija", SkipInSearch: true, TextAlign: "center"},
		{Name: "rok", Label: "Rok", Width: "5", Field: "rok", SkipInSearch: true, TextAlign: "center"},
	}
	s.subsintetickaKarticaKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "5", Field: "nalog", SkipInSearch: false, TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "5", Field: "danal", SkipInSearch: false, TextAlign: "center"},
		{Name: "opis", Label: "Opis", Width: "5", Field: "opis", SkipInSearch: false, TextAlign: "left"},
		{Name: "duguje", Label: "Duguje", Width: "5", Field: "duguje", SkipInSearch: false, TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "5", Field: "potrazuje", SkipInSearch: false, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "5", Field: "saldo", SkipInSearch: false, TextAlign: "right"},
	}
	s.subsintetickaKarticaStampaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj Naloga", Width: "8", Field: "nalog", SkipInSearch: true, TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "opis", Label: "Opis", Width: "30", Field: "opis", SkipInSearch: true, TextAlign: "left"},
		{Name: "duguje", Label: "Iznos Duguje", Width: "12", Field: "duguje", SkipInSearch: true, TextAlign: "right"},
		{Name: "potrazuje", Label: "Iznos Potrazuje", Width: "12", Field: "potrazuje", SkipInSearch: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "saldo", SkipInSearch: true, TextAlign: "right"},
		{Name: "znak", Label: "D/P", Width: "3", Field: "znak", SkipInSearch: true, TextAlign: "center"},
	}
}

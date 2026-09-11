package robno

import (
	"context"
	"fmt"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
)

type RobnoPrometService interface {
	GetPrometArtikala(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetPrometPoKupcima(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetNabavkeOdDobavljaca(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetPrometRucLagerLista(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetPrometGradilista(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetPrometGradilisteVpcNc(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	GetRobneGrupeComboValues(context.Context) ([]domain.ComboItem, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetPrometArtiklaTableFields() []domain.Fields
	GetPrometKupcaTableFields() []domain.Fields
	GetPrometRucTableFields() []domain.Fields
	GetPrometOdDobavljacaTableFields() []domain.Fields
	GetPrometGradilistaTableFields() []domain.Fields
	GetPrometGradilisteVpcNcTableFields() []domain.Fields
	GetPrometRucLagerListaTableFields() []domain.Fields
}

type RobnoPrometResource struct {
	rproRepo                         *repository.BaseRepository[domain.Rpro]
	magRepo                          *repository.BaseRepository[domain.Magacini]
	grupaRepo                        *repository.BaseRepository[domain.Rgru]
	fvrRepo                          *repository.BaseRepository[domain.Fvr]
	prometGrupeArtikalaTableFields   []domain.Fields
	prometKupciTableFields           []domain.Fields
	prometDobavljaciTableFields      []domain.Fields
	prometRucTableFields             []domain.Fields
	prometGradilisteTableFields      []domain.Fields
	prometRucLagerListaTableFields   []domain.Fields
	prometGradilisteVpcNcTableFields []domain.Fields
}

func NewRobnoPrometService(rproRepo *repository.BaseRepository[domain.Rpro], magRepo *repository.BaseRepository[domain.Magacini], grupaRepo *repository.BaseRepository[domain.Rgru], fvrRepo *repository.BaseRepository[domain.Fvr]) *RobnoPrometResource {
	rs := &RobnoPrometResource{
		rproRepo:                         rproRepo,
		magRepo:                          magRepo,
		grupaRepo:                        grupaRepo,
		fvrRepo:                          fvrRepo,
		prometGrupeArtikalaTableFields:   []domain.Fields{},
		prometKupciTableFields:           []domain.Fields{},
		prometDobavljaciTableFields:      []domain.Fields{},
		prometRucTableFields:             []domain.Fields{},
		prometGradilisteTableFields:      []domain.Fields{},
		prometRucLagerListaTableFields:   []domain.Fields{},
		prometGradilisteVpcNcTableFields: []domain.Fields{},
	}
	rs.setTableFields()
	return rs
}
func (s *RobnoPrometResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

func (s *RobnoPrometResource) GetPrometArtiklaTableFields() []domain.Fields {
	return s.prometGrupeArtikalaTableFields
}
func (s *RobnoPrometResource) GetPrometKupcaTableFields() []domain.Fields {
	return s.prometKupciTableFields
}
func (s *RobnoPrometResource) GetPrometOdDobavljacaTableFields() []domain.Fields {
	return s.prometDobavljaciTableFields
}
func (s *RobnoPrometResource) GetPrometRucTableFields() []domain.Fields {
	return s.prometRucTableFields
}
func (s *RobnoPrometResource) GetPrometGradilistaTableFields() []domain.Fields {
	return s.prometGradilisteTableFields
}
func (s *RobnoPrometResource) GetPrometGradilisteVpcNcTableFields() []domain.Fields {
	return s.prometGradilisteVpcNcTableFields
}
func (s *RobnoPrometResource) GetPrometRucLagerListaTableFields() []domain.Fields {
	return s.prometRucLagerListaTableFields
}

func (s *RobnoPrometResource) GetPrometArtikala(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error {
	//TODO implememt
	return nil
}
func (s *RobnoPrometResource) GetPrometPoKupcima(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error {
	//TODO implememt
	return nil
}
func (s *RobnoPrometResource) GetNabavkeOdDobavljaca(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error {
	//TODO implememt
	return nil
}
func (s *RobnoPrometResource) GetPrometRucLagerLista(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error {
	//TODO implememt
	return nil
}

func (s *RobnoPrometResource) GetPrometGradilista(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error {
	//TODO implememt
	return nil
}
func (s *RobnoPrometResource) GetPrometGradilisteVpcNc(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error {
	//TODO implememt
	return nil
}
func (s *RobnoPrometResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
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
func (s *RobnoPrometResource) GetRobneGrupeComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.grupaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select gru, naziv from rgru", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("gru")
	sqlQuery, args := qb.Build()
	entites, err := s.grupaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := make([]domain.ComboItem, len(*entites))
	for i, entity := range *entites {
		comboItems[i] = domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.Gru),
			Value: fmt.Sprintf("%d - %s", entity.Gru, entity.Naziv),
		}
	}
	return comboItems, nil
}

func (s *RobnoPrometResource) setTableFields() {
	// Initialize table fields here
	s.prometGrupeArtikalaTableFields = []domain.Fields{
		{Name: "detalj", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "grupa", Label: "Grupa", Width: "6", Field: "gru", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", Field: "sifra", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "8", Field: "konto", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "28", Field: "naz1"},
		{Name: "jm", Label: "JM", Width: "4", Field: "jm", SkipInSearch: true, TextAlign: "center"},
		{Name: "ulaz", Label: "Količina ulaz", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "izlaz", Label: "Količina izlaz", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "finulaz", Label: "Finansijski ulaz", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "finizlaz", Label: "Finansijski izlaz", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKupciTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7", Field: "konto", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra kupca", Width: "8", Field: "sifra", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv kupca", Width: "20", Field: "naziv"},
		{Name: "adresa", Label: "Adresa", Width: "16", Field: "adresa"},
		{Name: "mesto", Label: "Mesto", Width: "9", Field: "mesto"},
		{Name: "iznos", Label: "Iznos", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ugrabat", Label: "Ugovoreni rabat", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kasa", Label: "Kasa skonto", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "netrezpdv", Label: "Neto (bez PDV)", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "neto", Label: "Neto", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometDobavljaciTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7", Field: "konto", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra dobavljača", Width: "8", Field: "sifra", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv dobavljača", Width: "18", Field: "naziv"},
		{Name: "adresa", Label: "Adresa", Width: "14", Field: "adresa"},
		{Name: "mesto", Label: "Mesto", Width: "7", Field: "mesto"},
		{Name: "fakvred", Label: "Fakturna vrednost", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "netofakvred", Label: "Neto fakturna vrednost", Width: "11", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ztrovred", Label: "Zavisni troškovi nabavke", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "nabvred", Label: "Nabavna vrednost", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ruc", Label: "RUC (razlika u ceni)", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procruc", Label: "Procenat RUC", Width: "7", SkipInSearch: true, TextAlign: "right"},
		{Name: "vpvred", Label: "VPC vrednost", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometRucTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "2", SkipInSearch: true, TextAlign: "center"},
		{Name: "mag", Label: "Magacin", Width: "5", Field: "mag", SkipInSearch: true, TextAlign: "right"},
		{Name: "brnal", Label: "Broj naloga", Width: "6", Field: "brnal", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "yrd", Label: "Vrsta dokumenta", Width: "5", Field: "yrd", SkipInSearch: true, TextAlign: "right"},
		{Name: "brdok", Label: "Broj dokumenta", Width: "6", Field: "brdok", SkipInSearch: true},
		{Name: "doniz", Label: "Izvorni dokument", Width: "6", Field: "doniz", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "dadok", SkipInSearch: true, TextAlign: "center"},
		{Name: "dospece", Label: "Dospeće", Width: "8", Field: "dospece", SkipInSearch: true, TextAlign: "center"},
		{Name: "netopyrd", Label: "Neto vrednost", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "navyrd", Label: "Nabavna vrednost", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ruc", Label: "RUC (razlika u ceni)", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procnuc", Label: "Procenat RUC", Width: "6", SkipInSearch: true, TextAlign: "right"},
		{Name: "fito", Label: "Konto", Width: "7", Field: "fito", SkipInSearch: true},
		{Name: "fana", Label: "Napomena", Width: "8", Field: "fana", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "18", Field: "naziv"},
		{Name: "sifona", Label: "Šifra artikla", Width: "7", Field: "sifona", SkipInSearch: true, TextAlign: "right"},
		{Name: "komerc", Label: "Komercijalista", Width: "10", Field: "komerc", SkipInSearch: true},
	}
	s.prometGradilisteTableFields = []domain.Fields{}
	s.prometGradilisteVpcNcTableFields = []domain.Fields{}
}

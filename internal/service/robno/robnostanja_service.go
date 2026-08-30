package robno

import (
	"context"

	"helia/internal/domain"
	"helia/internal/repository"
)

type RobnoStanjaService interface {
	GetStanjePojedinacnogArtikla(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetStanjaViseArtikala(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetStanjaSubsintetickogKonta(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetSvodjenjeZaliha(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetPojedinacnogArtiklaTableFields() []domain.Fields
	GetViseArtikalaTableFields() []domain.Fields
	GetSubsintetiskogKontaTableFields() []domain.Fields
	GetSvodjenjeZalihaTableFields() []domain.Fields
}

type RobnoStanjaResource struct {
	rproRepo                       repository.BaseRepository[domain.RobnoStanjeDto]
	pojedinacniArtikalTableFields  []domain.Fields
	viseArtikalaTableFields        []domain.Fields
	subsintetickogKontaTableFields []domain.Fields
	svodjenjeZalihaTableFields     []domain.Fields
}

func NewRobnoStanjaService(rproRepo repository.BaseRepository[domain.RobnoStanjeDto]) *RobnoStanjaResource {
	rs := &RobnoStanjaResource{
		rproRepo: rproRepo,
	}
	rs.setTableFileds()
	return rs
}

func (s *RobnoStanjaResource) GetStanjePojedinacnogArtikla(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.PrometParam) error {
	// TODO: Implement GetStanjePojedinacnogArtikla

	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikala(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.PrometParam) error {
	// TODO: Implement GetStanjaViseArtikala

	return nil
}
func (s *RobnoStanjaResource) GetStanjaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.PrometParam) error {
	// TODO: Implement GetStanjaSubsintetickogKonta

	return nil
}
func (s *RobnoStanjaResource) GetSvodjenjeZaliha(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.PrometParam) error {
	// TODO: Implement GetSvodjenjeZaliha

	return nil
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

func (s *RobnoStanjaResource) setTableFileds() {
	s.pojedinacniArtikalTableFields = []domain.Fields{}
	s.viseArtikalaTableFields = []domain.Fields{}
	s.subsintetickogKontaTableFields = []domain.Fields{}
	s.svodjenjeZalihaTableFields = []domain.Fields{}
}

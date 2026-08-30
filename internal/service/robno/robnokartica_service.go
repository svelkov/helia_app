package robno

import (
	"context"

	"helia/internal/domain"
	"helia/internal/repository"
)

// RobnoKarticaService exposes the two inventory-card views used by the Robno module.
type RobnoKarticaService interface {
	GetKarticaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetKarticaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetKarticaArtiklaTableFields() []domain.Fields
	GetKarticaSubsintetickogKontaTableFields() []domain.Fields
}

// RobnoKarticaResource reuses the established Promet queries while keeping the
// Robno handler independent from the larger Finance service interface.
type RobnoKarticaResource struct {
	robnoKarticaRepo                     *repository.BaseRepository[domain.RobnoStanjeDto]
	karticaArtiklaTableFields            []domain.Fields
	subsintetickaKarticaKontaTableFields []domain.Fields
}

func NewRobnoKarticaService(robnoKarticaRepo *repository.BaseRepository[domain.RobnoStanjeDto]) *RobnoKarticaResource {
	rs := &RobnoKarticaResource{
		robnoKarticaRepo: robnoKarticaRepo,
	}
	rs.setTableFields()
	return rs
}

func (s *RobnoKarticaResource) GetKarticaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	// TODO: Implement GetArticleCard using s.service and s.robnoKarticaRepo
	return nil
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	// TODO: Implement GetSubsyntheticCard using s.service and s.robnoKarticaRepo
	return nil
}

func (s *RobnoKarticaResource) GetKarticaArtiklaTableFields() []domain.Fields {
	return s.karticaArtiklaTableFields
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKontaTableFields() []domain.Fields {
	return s.subsintetickaKarticaKontaTableFields
}

func (s *RobnoKarticaResource) setTableFields() {
	// Initialize the table fields for the RobnoKarticaResource
	s.karticaArtiklaTableFields = []domain.Fields{
		// Add the fields for the article card table here
	}
	s.subsintetickaKarticaKontaTableFields = []domain.Fields{
		// Add the fields for the subsynthetic card table here
	}
}

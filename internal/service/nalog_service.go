package service

import "helia/internal/domain"

// AuthService defines the interface for authentication-related operations.
type NalogService interface {
	GetAll(page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Fnal, error)
	GetByID(idField string, idValue int) (*domain.Fnal, error)
}

type NalogServiceImpl struct {
	baseService    BaseService[domain.Fnal]
	nalogService   NalogService
	tipdok_service BaseService[domain.Tipdok]
}

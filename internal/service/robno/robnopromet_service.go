package robno

import (
	"context"

	"helia/internal/domain"
	finservice "helia/internal/service/finansijsko"
)

type RobnoPrometService interface {
	GetGrupe1(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetGrupe2(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetDobavljaci(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetRuc(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetGradiliste(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetGradilisteVpcNc(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
	GetGrupeTableFields() []domain.Fields
	GetDobavljaciTableFields() []domain.Fields
	GetRucTableFields() []domain.Fields
	GetGradilisteTableFields() []domain.Fields
}

type RobnoPrometResource struct{ promet finservice.PrometService }

func NewRobnoPrometService(promet finservice.PrometService) *RobnoPrometResource {
	return &RobnoPrometResource{promet: promet}
}
func (s *RobnoPrometResource) GetGrupe1(c context.Context, t *domain.TableData, b bool, z, p int, x domain.PrometParam) error {
	return s.promet.GetPrometKontaAnaliticki(c, t, b, z, p, x)
}
func (s *RobnoPrometResource) GetGrupe2(c context.Context, t *domain.TableData, b bool, z, p int, x domain.PrometParam) error {
	return s.promet.GetPrometKontaAnaliticki(c, t, b, z, p, x)
}
func (s *RobnoPrometResource) GetDobavljaci(c context.Context, t *domain.TableData, b bool, z, p int, x domain.PrometParam) error {
	return s.promet.GetPrometKontaAnaliticki(c, t, b, z, p, x)
}
func (s *RobnoPrometResource) GetRuc(c context.Context, t *domain.TableData, b bool, z, p int, x domain.PrometParam) error {
	return s.promet.GetPrometKontaAnaliticki(c, t, b, z, p, x)
}
func (s *RobnoPrometResource) GetGradiliste(c context.Context, t *domain.TableData, b bool, z, p int, x domain.PrometParam) error {
	return s.promet.GetPrometKontaAnaliticki(c, t, b, z, p, x)
}
func (s *RobnoPrometResource) GetGradilisteVpcNc(c context.Context, t *domain.TableData, b bool, z, p int, x domain.PrometParam) error {
	return s.promet.GetPrometKontaAnaliticki(c, t, b, z, p, x)
}
func (s *RobnoPrometResource) GetGrupeTableFields() []domain.Fields {
	return s.promet.GetKontaAnalitickiTableFields()
}
func (s *RobnoPrometResource) GetDobavljaciTableFields() []domain.Fields {
	return s.promet.GetKontaAnalitickiTableFields()
}
func (s *RobnoPrometResource) GetRucTableFields() []domain.Fields {
	return s.promet.GetKontaAnalitickiTableFields()
}
func (s *RobnoPrometResource) GetGradilisteTableFields() []domain.Fields {
	return s.promet.GetKontaAnalitickiTableFields()
}

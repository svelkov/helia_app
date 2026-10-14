package finansijsko

import (
	"context"
	"errors"
	"helia/internal/domain"
	"helia/internal/repository"
)

const (
	fvrContentTitle string = "FVR -OPSTI PODACI"
	fvrTableID      string = "fvr-table"
	fvrURLPrefix    string = "/api/fvr/"
	fvrURLGetAll    string = "/api/fvr/all"
)

var fvrTableFields = []domain.Fields{
	{Name: "god", Label: "Poslovna Godina", Type: "text", Width: "5"},
	{Name: "kar", Label: "Komintent", Type: "text", Width: "40"},
	{Name: "naziv", Label: "Naziv komintenta", Type: "text", Width: "4"},
}

type FvrService interface {
	GetAllFvr(ctx context.Context) (*domain.Firma, error)
	GetAllGod(ctx context.Context, nazivFirme string) ([]int, error)
	GetAllKar(ctx context.Context, nazivFirme string, god int) ([]int, error)
}
type FvrResource struct {
	fvrRepo *repository.BaseRepository[domain.Fvr]
}

func NewFvrService(fvrRepo *repository.BaseRepository[domain.Fvr]) *FvrResource {
	return &FvrResource{
		fvrRepo: fvrRepo,
	}
}

func (s *FvrResource) GetAllFvr(ctx context.Context) (*domain.Firma, error) {
	firma := &domain.Firma{}
	// The firme of every period: the list of the choice of the period (repository.FvrFirme).
	entities, err := repository.FvrFirme(ctx, s.fvrRepo)
	if err != nil {
		return firma, err
	}
	if len(entities) == 0 {
		return nil, errors.New("no available config data")
	}
	for _, item := range entities {
		poslGodina, err := s.GetAllGod(ctx, item.Naziv)
		if err != nil {
			return nil, errors.New("no available config data")
		}
		poslGodine := []domain.Godina{}
		for _, god := range poslGodina {
			knjigovodstvo, err := s.GetAllKar(ctx, item.Naziv, god)
			if err != nil {
				return nil, errors.New("no available config data")
			}
			poslGodine = append(poslGodine, domain.Godina{
				God: god,
				Kar: knjigovodstvo,
			})
		}
		firma.Firme = append(firma.Firme, domain.FvrFirma{
			Naziv:     item.Naziv,
			Knjigovod: item.Knjigovod,
			Godine:    poslGodine,
		},
		)
	}

	return firma, nil
}

func (s *FvrResource) GetAllGod(ctx context.Context, nazivFirme string) ([]int, error) {
	godine, err := repository.FvrGodine(ctx, s.fvrRepo, nazivFirme)
	if err != nil {
		return []int{}, err
	}
	return godine, nil
}

func (s *FvrResource) GetAllKar(ctx context.Context, nazivFirme string, god int) ([]int, error) {
	karovi, err := repository.FvrKarovi(ctx, s.fvrRepo, nazivFirme, god)
	if err != nil {
		return []int{}, err
	}
	return karovi, nil
}

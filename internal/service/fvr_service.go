package service

import (
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
	GetAllFvr() ([]domain.Fvr, []int, error)
	GetAllKar(god int) ([]int, error)
	GetAllGod(kar int) ([]int, error)
}
type FvrResource struct {
	fvrRepo *repository.BaseRepository[domain.Fvr]
}

func NewFvrService(fvrRepo *repository.BaseRepository[domain.Fvr]) *FvrResource {
	return &FvrResource{
		fvrRepo: fvrRepo,
	}
}

func (s *FvrResource) GetAllFvr() ([]domain.Fvr, []int, error) {
	args := []interface{}{}

	selectQuery := `SELECT DISTINCT ON (naziv) god, kar, naziv FROM fvr
					WHERE god > 0
					order by fvr.naziv, fvr.god `

	entities, err := s.fvrRepo.GetAllCustom(selectQuery, "", args, "", "")
	if err != nil {
		return *entities, []int{}, err
	}
	if len(*entities) == 0 {
		return nil, []int{}, errors.New("no available config data")
	}
	poslGodina, err := s.GetAllGod((*entities)[0].God)
	if err != nil {
		return nil, []int{}, errors.New("no available config data")
	}
	return *entities, poslGodina, nil
}
func (s *FvrResource) GetAllKar(god int) ([]int, error) {
	args := []interface{}{}
	args = append(args, god)
	result := []int{}
	selectQuery := `SELECT DISTINCT ON (kar) god, kar, naziv FROM fvr
					WHERE god = $1
					order by fvr.kar `

	entities, err := s.fvrRepo.GetAllCustom(selectQuery, "", args, "", "")
	if err != nil {
		return []int{}, err
	}
	for _, item := range *entities {
		result = append(result, item.Kar)
	}

	return result, nil
}

func (s *FvrResource) GetAllGod(kar int) ([]int, error) {
	args := []interface{}{}
	args = append(args, kar)
	result := []int{}
	selectQuery := `SELECT DISTINCT ON (god) god, kar, naziv FROM fvr
					WHERE god > 0 AND kar = $1
					order by fvr.god desc`

	entities, err := s.fvrRepo.GetAllCustom(selectQuery, "", args, "", "")
	if err != nil {
		return []int{}, err
	}
	for _, item := range *entities {
		result = append(result, item.God)
	}

	return result, nil
}

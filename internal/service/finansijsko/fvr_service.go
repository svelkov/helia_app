package finansijsko

import (
	"errors"
	"helia/internal/domain"
	"helia/internal/repository"

	"github.com/gin-gonic/gin"
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
	GetAllFvr(c *gin.Context) (*domain.Firma, error)
	GetAllGod(c *gin.Context, nazivFirme string) ([]int, error)
	GetAllKar(c *gin.Context, nazivFirme string, god int) ([]int, error)
}
type FvrResource struct {
	fvrRepo *repository.BaseRepository[domain.Fvr]
}

func NewFvrService(fvrRepo *repository.BaseRepository[domain.Fvr]) *FvrResource {
	return &FvrResource{
		fvrRepo: fvrRepo,
	}
}

func (s *FvrResource) GetAllFvr(c *gin.Context) (*domain.Firma, error) {
	args := []interface{}{}
	firma := &domain.Firma{}

	selectQuery := `SELECT DISTINCT ON (naziv) naziv FROM fvr
					WHERE god > 0
					order by fvr.naziv`

	entities, err := s.fvrRepo.GetAllCustom(c, selectQuery, "", args, "", "")
	if err != nil {
		return firma, err
	}
	if len(*entities) == 0 {
		return nil, errors.New("no available config data")
	}
	for _, item := range *entities {
		poslGodina, err := s.GetAllGod(c, item.Naziv)
		if err != nil {
			return nil, errors.New("no available config data")
		}
		poslGodine := []domain.Godina{}
		for _, god := range poslGodina {
			knjigovodstvo, err := s.GetAllKar(c, item.Naziv, god)
			if err != nil {
				return nil, errors.New("no available config data")
			}
			poslGodine = append(poslGodine, domain.Godina{
				God: god,
				Kar: knjigovodstvo,
			})
		}
		firma.Firme = append(firma.Firme, domain.FvrFirma{
			Naziv:  item.Naziv,
			Godine: poslGodine,
		},
		)
	}

	return firma, nil
}

func (s *FvrResource) GetAllGod(c *gin.Context, nazivFirme string) ([]int, error) {
	args := []interface{}{}
	args = append(args, nazivFirme)
	result := []int{}
	selectQuery := `SELECT DISTINCT ON (god) god FROM fvr
					WHERE fvr.naziv = $1
					order by fvr.god desc`

	entities, err := s.fvrRepo.GetAllCustom(c, selectQuery, "", args, "", "")
	if err != nil {
		return []int{}, err
	}
	for _, item := range *entities {
		result = append(result, item.God)
	}

	return result, nil
}

func (s *FvrResource) GetAllKar(c *gin.Context, nazivFirme string, god int) ([]int, error) {
	args := []interface{}{}
	args = append(args, nazivFirme, god)
	result := []int{}
	selectQuery := `SELECT kar FROM fvr
					WHERE fvr.naziv = $1 AND god = $2
					order by fvr.kar`

	entities, err := s.fvrRepo.GetAllCustom(c, selectQuery, "", args, "", "")
	if err != nil {
		return []int{}, err
	}
	for _, item := range *entities {
		result = append(result, item.Kar)
	}

	return result, nil
}

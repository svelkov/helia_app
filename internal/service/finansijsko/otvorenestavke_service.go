package finansijsko

import (
	"fmt"
	"helia/internal/domain"
	"helia/internal/repository"

	"github.com/gin-gonic/gin"
)

// OtvoreneStavkeService defines the service interface for Otvorene Stavke operations
type OtvoreneStavkeService interface {
	GetOtvoreneStavke(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetZatvoreneStavke(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetIOS(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetOpomene(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetDospelaPotraživanja(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPregledDugovanjaPredStarosti(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPregledDospelogDuga(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPovezivanjRacunaIUplata(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetOtvoreneStavkeFields() []domain.Fields
}

// OtvoreneStavkeResource implements OtvoreneStavkeService
type OtvoreneStavkeResource struct {
	repo                     repository.BaseRepository[domain.Fkpl]
	fieldOtvoreneStavke      []domain.Fields
	fieldZatvoreneStavke     []domain.Fields
	fieldDospelaPotraživanja []domain.Fields
}

// NewOtvoreneStavkeService creates a new service instance
func NewOtvoreneStavkeService(repo repository.BaseRepository[domain.Fkpl]) *OtvoreneStavkeResource {
	service := &OtvoreneStavkeResource{
		repo: repo,
	}
	service.setServiceFieldValues()
	return service
}

// GetOtvoreneStavke retrieves open items data
func (s *OtvoreneStavkeResource) GetOtvoreneStavke(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// if err := common.ValidateRequiredParams(c, "konto"); err != nil {
	// 	return err
	// }

	// common.SetTableConfig(tbl, "Otvorene stavke", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldOtvoreneStavke

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetZatvoreneStavke retrieves closed items data
func (s *OtvoreneStavkeResource) GetZatvoreneStavke(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// common.SetTableConfig(tbl, "Zatvorene stavke", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldZatvoreneStavke

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetIOS retrieves IOS (Izvod otvorenih stavki) data
func (s *OtvoreneStavkeResource) GetIOS(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query

	// common.SetTableConfig(tbl, "IOS", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldOtvoreneStavke

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetOpomene retrieves reminders data
func (s *OtvoreneStavkeResource) GetOpomene(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// if err := common.ValidateRequiredParams(c, "konto"); err != nil {
	// 	return err
	// }

	// common.SetTableConfig(tbl, "Opomene", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldOtvoreneStavke

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetDospelaPotraživanja retrieves due receivables/payables data
func (s *OtvoreneStavkeResource) GetDospelaPotraživanja(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// if err := common.ValidateRequiredParams(c, "konto"); err != nil {
	// 	return err
	// }

	// common.SetTableConfig(tbl, "Dospela potraživanja/dugovanja", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldDospelaPotraživanja

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetPregledDugovanjaPredStarosti retrieves payables overview by age
func (s *OtvoreneStavkeResource) GetPregledDugovanjaPredStarosti(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// common.SetTableConfig(tbl, "Pregled Dugovanja/Obaveze po starosti", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldDospelaPotraživanja

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetPregledDospelogDuga retrieves due debt overview
func (s *OtvoreneStavkeResource) GetPregledDospelogDuga(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// if err := common.ValidateRequiredParams(c, "konto"); err != nil {
	// 	return err
	// }
	// common.SetTableConfig(tbl, "Pregled dospelog duga", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldDospelaPotraživanja

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}

// GetPovezivanjRacunaIUplata retrieves account and payment linking data
func (s *OtvoreneStavkeResource) GetPovezivanjRacunaIUplata(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// if err := common.ValidateRequiredParams(c, "konto"); err != nil {
	// 	return err
	// }

	// common.SetTableConfig(tbl, "Povezivanje računa i uplata", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldOtvoreneStavke

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

	return nil
}
func (s *OtvoreneStavkeResource) GetOtvoreneStavkeFields() []domain.Fields {
	return s.fieldOtvoreneStavke
}

// setServiceFieldValues initializes table field definitions for Otvorene Stavke
func (s *OtvoreneStavkeResource) setServiceFieldValues() {
	// Otvorene Stavke fields (Tab 1, 2, 3, 4, 8)
	s.fieldOtvoreneStavke = []domain.Fields{
		{Name: "detalj", Label: "Detalj", Width: "5", Field: "fnal.detalj", SkipInSearch: true},
		{Name: "fu", Label: "F/U", Width: "5", Field: "fnal.fu", SkipInSearch: false},
		{Name: "broj_dokum", Label: "Broj dokum.", Width: "10", Field: "fnal.nalog", SkipInSearch: false},
		{Name: "vrsta_dokumenta", Label: "Vrsta dokumenta", Width: "12", Field: "fnal.tipdok", SkipInSearch: false},
		{Name: "dat_dokum", Label: "Dat. dokum.", Width: "10", Field: "fnal.danal", SkipInSearch: false},
		{Name: "oj", Label: "OJ", Width: "5", Field: "fnal.brst", SkipInSearch: true},
		{Name: "broj_naloga", Label: "Broj naloga", Width: "10", Field: "fnal.nalog", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fnal.dug", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fnal.pot", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "opis_knjizenja", Label: "Opis knjiženja", Width: "15", Field: "fnal.opis", SkipInSearch: false},
	}

	// Zatvorene Stavke fields (Tab 2)
	s.fieldZatvoreneStavke = []domain.Fields{
		{Name: "detalj", Label: "Detalj", Width: "5", Field: "fnal.detalj", SkipInSearch: true},
		{Name: "fu", Label: "F/U", Width: "5", Field: "fnal.fu", SkipInSearch: false},
		{Name: "broj_dokum", Label: "Broj dokum.", Width: "10", Field: "fnal.nalog", SkipInSearch: false},
		{Name: "vrsta_dokumenta", Label: "Vrsta dokumenta", Width: "12", Field: "fnal.tipdok", SkipInSearch: false},
		{Name: "dat_dokum", Label: "Dat. dokum.", Width: "10", Field: "fnal.danal", SkipInSearch: false},
		{Name: "broj_naloga", Label: "Broj naloga", Width: "10", Field: "fnal.nalog", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fnal.dug", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fnal.pot", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "opis_knjizenja", Label: "Opis knjiženja", Width: "12", Field: "fnal.opis", SkipInSearch: false},
	}

	// Dospela Potraživanja fields (Tab 5, 6, 7)
	s.fieldDospelaPotraživanja = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fkpl.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fkpl.sifra", SkipInSearch: false},
		{Name: "naziv_partnera", Label: "Naziv partnera", Width: "15", Field: "partneri.naziv", SkipInSearch: false},
		{Name: "mesto", Label: "Mesto", Width: "12", Field: "partneri.mesto", SkipInSearch: false},
		{Name: "ukupna_realizacija", Label: "Ukupna realizacija", Width: "10", Field: "fnal.iznos", SkipInSearch: true},
		{Name: "raceno", Label: "Račeno", Width: "10", Field: "", SkipInSearch: true},
		{Name: "ukupan_dug", Label: "Ukupan DUG", Width: "10", Field: "", SkipInSearch: true},
		{Name: "dospeli_0_15", Label: "0-15 dana", Width: "10", Field: "", SkipInSearch: true},
		{Name: "dospeli_16_30", Label: "16-30 dana", Width: "10", Field: "", SkipInSearch: true},
		{Name: "dospeli_31_60", Label: "31-60 dana", Width: "10", Field: "", SkipInSearch: true},
		{Name: "dospeli_61_90", Label: "61-90 dana", Width: "10", Field: "", SkipInSearch: true},
		{Name: "dospeli_90plus", Label: ">90 dana", Width: "10", Field: "", SkipInSearch: true},
	}
}

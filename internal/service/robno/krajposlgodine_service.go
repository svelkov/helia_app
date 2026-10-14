package robno

import (
	"context"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

type KrajPoslovneGodineService interface {
	GetMagacini(context.Context) ([]domain.ComboItem, error)
	GetPopis(context.Context, domain.KrajPoslovneGodineParams) (domain.KrajPopisData, error)
	GetVisakManjak(context.Context, domain.KrajPoslovneGodineParams) ([]domain.KrajVisakManjakRow, error)
	PrepisStanja(context.Context, domain.KrajPoslovneGodineParams, int) error
	ValidateRSTA(domain.KrajPoslovneGodineParams) []domain.FieldError
	ValidateObrada(domain.KrajPoslovneGodineParams) []domain.FieldError
	GetType1Fields() []domain.Fields
	GetType2Fields() []domain.Fields
	GetVisakManjakFields() []domain.Fields
}

type KrajPoslovneGodineResource struct {
	magRepo     *repository.BaseRepository[domain.Magacini]
	drstaRepo   *repository.BaseRepository[domain.Drsta]
	commonSvc   commonsvc.CommonService
	robno       *repository.RobnoRepository
	type1Fields []domain.Fields
	type2Fields []domain.Fields
	visakFields []domain.Fields
}

func NewKrajPoslovneGodineService(magRepo *repository.BaseRepository[domain.Magacini], drstaRepo *repository.BaseRepository[domain.Drsta], commonSvc commonsvc.CommonService) *KrajPoslovneGodineResource {
	s := &KrajPoslovneGodineResource{magRepo: magRepo, drstaRepo: drstaRepo, commonSvc: commonSvc,
		robno: repository.NewRobnoRepository(drstaRepo.DB)}
	s.setTableFields()
	return s
}

func (s *KrajPoslovneGodineResource) GetType1Fields() []domain.Fields       { return s.type1Fields }
func (s *KrajPoslovneGodineResource) GetType2Fields() []domain.Fields       { return s.type2Fields }
func (s *KrajPoslovneGodineResource) GetVisakManjakFields() []domain.Fields { return s.visakFields }

func (s *KrajPoslovneGodineResource) ValidateRSTA(p domain.KrajPoslovneGodineParams) []domain.FieldError {
	var errors []domain.FieldError
	if p.Magacin <= 0 {
		errors = append(errors, domain.FieldError{Field: "magacin", ErrorMessage: "obavezan podatak"})
	}
	if p.TipZal != 1 && p.TipZal != 2 {
		errors = append(errors, domain.FieldError{Field: "tipzal", ErrorMessage: "dozvoljene vrednosti su 1 ili 2"})
	}
	if p.Cena < 1 || p.Cena > 4 {
		errors = append(errors, domain.FieldError{Field: "cena", ErrorMessage: "dozvoljene vrednosti su 1-4"})
	}
	if p.NovaGod <= 0 {
		errors = append(errors, domain.FieldError{Field: "novagod", ErrorMessage: "obavezan podatak"})
	}
	if p.OdSifre != 0 && p.DoSifre != 0 && p.OdSifre > p.DoSifre {
		errors = append(errors, domain.FieldError{Field: "odsifre", ErrorMessage: "neispravan opseg"})
	}
	if p.OdKonta != "" && p.DoKonta != "" && p.OdKonta > p.DoKonta {
		errors = append(errors, domain.FieldError{Field: "odkonta", ErrorMessage: "neispravan opseg"})
	}
	return errors
}

func (s *KrajPoslovneGodineResource) ValidateObrada(p domain.KrajPoslovneGodineParams) []domain.FieldError {
	var errors []domain.FieldError
	if p.Magacin <= 0 {
		errors = append(errors, domain.FieldError{Field: "magacin", ErrorMessage: "obavezan podatak"})
	}
	if p.NovaGod <= 0 {
		errors = append(errors, domain.FieldError{Field: "novagod", ErrorMessage: "obavezan podatak"})
	}
	if p.Nalog <= 0 {
		errors = append(errors, domain.FieldError{Field: "nalog", ErrorMessage: "obavezan podatak"})
	}
	if p.Dokum <= 0 {
		errors = append(errors, domain.FieldError{Field: "dokum", ErrorMessage: "obavezan podatak"})
	}
	if p.Vrd <= 0 {
		errors = append(errors, domain.FieldError{Field: "vrd", ErrorMessage: "obavezan podatak"})
	}
	return errors
}

// GetMagacini returns the magacini of the current period keyed by mag (CommonService).
func (s *KrajPoslovneGodineResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinByMagComboValues(ctx)
}

func (s *KrajPoslovneGodineResource) GetPopis(ctx context.Context, params domain.KrajPoslovneGodineParams) (domain.KrajPopisData, error) {
	popisData := domain.KrajPopisData{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.KrajPopisData{}, fmt.Errorf("no user session found")
	}
	if params.TipZal == 0 {
		b := querybuilder.New(`SELECT tipzal FROM magacini`).Where("mag = ?", params.Magacin)
		if err := s.magRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
			return domain.KrajPopisData{}, err
		}
		entities, err := s.magRepo.Select(ctx, b)
		if err != nil {
			return domain.KrajPopisData{}, fmt.Errorf("get TIPZAL: %w", err)
		}
		if entities == nil || len(*entities) == 0 {
			return domain.KrajPopisData{}, fmt.Errorf("no TIPZAL found")
		}
		popisData.TipZal = (*entities)[0].Tipzal
	}
	if params.TipZal == 1 {
		return s.getPopisType1(ctx, params)
	}
	if params.TipZal == 2 {
		return s.getPopisType2(ctx, params)
	}
	return domain.KrajPopisData{}, fmt.Errorf("unsupported TIPZAL %d; expected 1 or 2", params.TipZal)
}

func (s *KrajPoslovneGodineResource) getPopisType1(ctx context.Context, p domain.KrajPoslovneGodineParams) (domain.KrajPopisData, error) {
	if err := s.requireTables(ctx, "rsta", "rcene", "rsif", "rgru"); err != nil {
		return domain.KrajPopisData{}, err
	}
	rows, err := s.robno.PopisZaliha1(ctx, p)
	return domain.KrajPopisData{TipZal: 1, Type1: rows}, err
}

func (s *KrajPoslovneGodineResource) getPopisType2(ctx context.Context, p domain.KrajPoslovneGodineParams) (domain.KrajPopisData, error) {
	if err := s.requireTables(ctx, "drsta", "rsta", "rsif", "rgru"); err != nil {
		return domain.KrajPopisData{}, err
	}
	rows, err := s.robno.PopisZaliha2(ctx, p)
	return domain.KrajPopisData{TipZal: 2, Type2: rows}, err
}

func (s *KrajPoslovneGodineResource) GetVisakManjak(ctx context.Context, p domain.KrajPoslovneGodineParams) ([]domain.KrajVisakManjakRow, error) {
	user := domain.GetSessionFromStdContext(ctx)
	if user == nil {
		return nil, fmt.Errorf("no user session found")
	}
	if err := s.requireTables(ctx, "rsta", "rpro", "rdok", "rsif", "rgru"); err != nil {
		return nil, err
	}
	return s.robno.VisakManjak(ctx, p)
}

func (s *KrajPoslovneGodineResource) PrepisStanja(ctx context.Context, p domain.KrajPoslovneGodineParams, targetYear int) error {
	if targetYear <= 0 {
		return fmt.Errorf("invalid target year")
	}
	price, err := prepisPriceColumn(p.Cena)
	if err != nil {
		return err
	}
	tx, err := s.robno.BeginTx()
	if err != nil {
		return fmt.Errorf("begin prepis transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, table := range []string{"rsta", "drsta", "rcene", "rdok", "rnal", "sr", "rpro", "rsal", "rsin", "dokvrsta"} {
		exists, err := s.robno.TableExists(ctx, tx, table)
		if err != nil {
			return fmt.Errorf("verify prepis table %s: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("prepis schema unavailable: required table %s does not exist", strings.ToUpper(table))
		}
	}
	return fmt.Errorf("prepis schema unavailable for price %s: WinDev column mapping is not present; no rows were written", price)
}
func prepisPriceColumn(choice int) (string, error) {
	switch choice {
	case 1:
		return "CENA", nil
	case 2:
		return "PROSNC", nil
	case 3:
		return "NCENA", nil
	case 4:
		return "VPCENA", nil
	default:
		return "", fmt.Errorf("invalid prepis price option %d; expected 1-4", choice)
	}
}
func (s *KrajPoslovneGodineResource) requireTables(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		exists, err := s.robno.TableExists(ctx, nil, table)
		if err != nil {
			return fmt.Errorf("verify %s schema: %w", strings.ToUpper(table), err)
		}
		if !exists {
			return fmt.Errorf("schema unavailable: required table %s does not exist", strings.ToUpper(table))
		}
	}
	return nil
}

func (s *KrajPoslovneGodineResource) setTableFields() {
	s.type1Fields = []domain.Fields{
		{Name: "redbr", Label: "Red.br.", Field: "redbr"},
		{Name: "konto", Label: "Konto", Field: "konto"},
		{Name: "sifra", Label: "Šifra", Field: "sifra"},
		{Name: "naziv", Label: "Naziv", Field: "naziv"},
		{Name: "jm", Label: "JM", Field: "jm"},
		{Name: "grupa", Label: "Grupa", Field: "grupa"},
		{Name: "nazivgrupe", Label: "Naziv grupe", Field: "nazivgrupe"},
		{Name: "cena", Label: "Cena", Field: "cena", TextAlign: "right"},
		{Name: "kolicina1", Label: "Količina 1", Field: "kolicina1", TextAlign: "right"},
		{Name: "kolicina2", Label: "Količina 2", Field: "kolicina2", TextAlign: "right"},
		{Name: "kolicina3", Label: "Količina 3", Field: "kolicina3", TextAlign: "right"},
		{Name: "ukupnakolicina", Label: "Ukupna količina", Field: "ukupnakolicina", TextAlign: "right"},
		{Name: "stanjezaliha", Label: "Stanje zaliha", Field: "stanjezaliha", TextAlign: "right"},
	}
	s.type2Fields = []domain.Fields{
		{Name: "redbr", Label: "Red.br.", Field: "redbr"},
		{Name: "konto", Label: "Konto", Field: "konto"},
		{Name: "sifra", Label: "Šifra", Field: "sifra"},
		{Name: "naziv", Label: "Naziv", Field: "naziv"},
		{Name: "jm", Label: "JM", Field: "jm"},
		{Name: "grupa", Label: "Grupa", Field: "grupa"},
		{Name: "nazivgrupe", Label: "Naziv grupe", Field: "nazivgrupe"},
		{Name: "otk", Label: "OTK", Field: "otk"},
		{Name: "serija", Label: "Serija", Field: "serija"},
		{Name: "rok", Label: "Rok", Field: "rok"},
		{Name: "kolicina1", Label: "Količina 1", Field: "kolicina1", TextAlign: "right"},
		{Name: "kolicina2", Label: "Količina 2", Field: "kolicina2", TextAlign: "right"},
		{Name: "kolicina3", Label: "Količina 3", Field: "kolicina3", TextAlign: "right"},
		{Name: "ukupnakolicina", Label: "Ukupna količina", Field: "ukupnakolicina", TextAlign: "right"},
		{Name: "stanjezaliha", Label: "Stanje zaliha", Field: "stanjezaliha", TextAlign: "right"},
	}
	s.visakFields = []domain.Fields{
		{Name: "redbr", Label: "Red.br.", Field: "redbr"},
		{Name: "sifra", Label: "Šifra", Field: "sifra"},
		{Name: "konto", Label: "Konto", Field: "konto"},
		{Name: "naziv", Label: "Naziv", Field: "naziv"},
		{Name: "jm", Label: "JM", Field: "jm"},
		{Name: "cena", Label: "Cena", Field: "cena", TextAlign: "right"},
		{Name: "kolicinapopisa", Label: "Popisano stanje", Field: "kolicinapopisa", TextAlign: "right"},
		{Name: "iznospopisa", Label: "Iznos popisa", Field: "iznospopisa", TextAlign: "right"},
		{Name: "stanjeknjigovodstveno", Label: "Knjigovodstveno stanje", Field: "stanjeknjigovodstveno", TextAlign: "right"},
		{Name: "iznosknjigovodstveno", Label: "Knjigovodstveni iznos", Field: "iznosknjigovodstveno", TextAlign: "right"},
		{Name: "visak", Label: "Višak", Field: "visak", TextAlign: "right"},
		{Name: "iznosviska", Label: "Iznos viška", Field: "iznosviska", TextAlign: "right"},
		{Name: "manjak", Label: "Manjak", Field: "manjak", TextAlign: "right"},
		{Name: "iznosmanjka", Label: "Iznos manjka", Field: "iznosmanjka", TextAlign: "right"},
		{Name: "finansijskivisak", Label: "Finansijski višak", Field: "finansijskivisak", TextAlign: "right"},
		{Name: "finansijskimanjak", Label: "Finansijski manjak", Field: "finansijskimanjak", TextAlign: "right"},
	}
}

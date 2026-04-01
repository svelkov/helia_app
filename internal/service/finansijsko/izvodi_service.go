package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"

	"github.com/gin-gonic/gin"
)

// IzvodiService defines the interface for operations related to bank statements (Izvodi).
type IzvodiService interface {
	GetMasterTableFields() []domain.Fields
	GetDetailTableFields() []domain.Fields
	GetIzvodiHeader(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetIzvodiDetail(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	ImportIzvod(c *gin.Context, filePath string, software string, options map[string]bool) error
	AzurirajKonta(c *gin.Context, izvodID string) error
	BrisiIzvod(c *gin.Context, izvodID string) error
	ProveriRaznotezu(c *gin.Context, izvodID string) error
	KnjiziIzvod(c *gin.Context, izvodID string, nalogParams map[string]interface{}) error
	OznaciKaoNeproknjizene(c *gin.Context, izvodIDs []string) error
}

// IzvodiResource implements the IzvodiService interface.
type IzvodiResource struct {
	service                 *service.BaseService[domain.Fizvzag]
	izvhdrrepo              *repository.BaseRepository[domain.Fizvzag]
	izvdetrepo              *repository.BaseRepository[domain.Fizvdet]
	cfg                     config.Config
	izvodiHeaderTableFields []domain.Fields
	izvodiDetailTableFields []domain.Fields
}

func NewIzvodiResource(izvhdrrepo *repository.BaseRepository[domain.Fizvzag], izvdetrepo *repository.BaseRepository[domain.Fizvdet], cfg config.Config) *IzvodiResource {
	rs := &IzvodiResource{
		izvhdrrepo: izvhdrrepo,
		izvdetrepo: izvdetrepo,
		cfg:        cfg,
	}
	rs.SetIzvodiFields()
	return rs
}

func (s *IzvodiResource) GetIzvodiHeader(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// Get parameters from form
	idbanke := c.Query("idbanke")
	odBrojaIzvoda := c.Query("odBrojaIzvoda")
	doBrojaIzvoda := c.Query("doBrojaIzvoda")
	odDatumaIzvoda := c.Query("odDatumaIzvoda")
	doDatumaIzvoda := c.Query("doDatumaIzvoda")
	statusIzvoda := c.Query("statusIzvoda")
	brojRacuna := c.Query("brojRacuna")
	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.izvhdrrepo.GetHasGodHasKar()

	// Build query with QueryBuilder matching WinDev SQL structure
	qb := common.NewQueryBuilder(`
		SELECT 
			i.idfizvzag,
			i.god,
			i.kar,
			i.brrac,
			i.izvbr,
			i.datizv,
			i.konto,
			i.sifra,
			i.prstanje,
			i.ukdug,
			i.ukpot,
			i.nstanje,
			i.ukbrst,
			i.nalog,
			i.tipdok,
			i.izvsts,
			i.idbanke,
			b.banka,
			b.konto as konto_ba,
			b.sifra as sifra_ba
		FROM fizvzag i
		LEFT JOIN banke b ON b.idbanke = i.idbanke`, true)

	if hasGod {
		qb.AddEqual("i.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("i.kar", session.SelectedKar)
	}

	if idbanke != "" && idbanke != "0" {
		qb.AddEqual("i.idbanke", idbanke)
	}

	if brojRacuna != "" {
		qb.AddEqual("i.brrac", brojRacuna)
	}

	if odBrojaIzvoda != "" && odBrojaIzvoda != "0" {
		qb.AddCondition("i.izvbr", odBrojaIzvoda, ">=")
	}

	if doBrojaIzvoda != "" && doBrojaIzvoda != "999999" {
		qb.AddCondition("i.izvbr", doBrojaIzvoda, "<=")
	}

	if odDatumaIzvoda != "" {
		qb.AddCondition("i.datizv", odDatumaIzvoda, ">=")
	}

	if doDatumaIzvoda != "" {
		qb.AddCondition("i.datizv", doDatumaIzvoda, "<=")
	}

	if statusIzvoda != "" {
		qb.AddEqual("i.izvsts", statusIzvoda)
	}

	// Add search conditions if search text is provided
	if searchText != "" {
		qb.AddSearchConditions(s.GetMasterTableFields(), searchText)
	}

	qb.AddOrderBy("i.datizv DESC, i.izvbr DESC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.izvhdrrepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%v", entity.Izvbr),
				entity.Datizv.Time.Format(common.DateLayout),
				entity.Brrac,
				entity.Konto,
				entity.Sifra,
				entity.Banka.String,
				common.FormatNumberWithSystemLocale(entity.Prstanje, 2),
				common.FormatNumberWithSystemLocale(entity.Ukdug, 2),
				common.FormatNumberWithSystemLocale(entity.Ukpot, 2),
				common.FormatNumberWithSystemLocale(entity.Nstanje, 2),
				fmt.Sprintf("%d", entity.Ukbrst),
				fmt.Sprintf("%d", entity.Nalog),
				entity.Tipdok,
				entity.Izvsts,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

func (s *IzvodiResource) GetIzvodiDetail(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// Get izvodID from query parameter
	izvodID := c.Query("idfizvzag")
	if izvodID == "" {
		return fmt.Errorf("izvodID parameter required")
	}

	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)

	qb := common.NewQueryBuilder(`
		SELECT 
			d.idfizvdet,
			d.rbr,
			d.konto,
			d.sifra,
			d.kat,
			d.vrd,
			d.konto1,
			d.sifra1,
			d.iznos,
			d.duguje,
			d.potrazuje,
			d.nsedprim,
			d.brracup,
			d.osnplac,
			d.sdozn,
			d.sdozn1,
			d.modelzad,
			d.pnabrzad,
			d.mododob,
			d.pnabrodob,
			d.prekl
		FROM fizvdet d`, true)

	qb.AddEqual("d.idfizvzag", izvodID)

	// Add search conditions if search text is provided
	if searchText != "" {
		qb.AddSearchConditions(s.GetDetailTableFields(), searchText)
	}

	qb.AddOrderBy("d.rbr")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.izvdetrepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%d", entity.Rbr),
				entity.Konto,
				entity.Sifra,
				fmt.Sprintf("%d", entity.Kat),
				fmt.Sprintf("%d", entity.Vrd),
				entity.Konto1,
				entity.Sifra1,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				entity.Nsedprim,
				entity.Brracup,
				entity.Osnplac,
				entity.Sdozn,
				entity.Sdozn1,
				entity.Modelzad,
				entity.Pnabrzad,
				entity.Mododob,
				entity.Pnabrodob,
				entity.Prekl,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

func (s *IzvodiResource) ImportIzvod(c *gin.Context, filePath string, software string, options map[string]bool) error {
	// TODO: Implement import logic based on software type
	// - Parse CSV/XML/XLSX file based on software selection
	// - Validate data
	// - Insert into izvodi and izvodi_detalji tables
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) AzurirajKonta(c *gin.Context, izvodID string) error {
	// TODO: Implement account update logic
	// - Update accounts in izvodi_detalji based on parameters
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) BrisiIzvod(c *gin.Context, izvodID string) error {
	// TODO: Implement delete logic
	// - Delete from izvodi_detalji first
	// - Delete from izvodi
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) ProveriRaznotezu(c *gin.Context, izvodID string) error {
	// TODO: Implement balance check logic
	// - Calculate total debits and credits
	// - Check if they match
	// - Return error if unbalanced
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) KnjiziIzvod(c *gin.Context, izvodID string, nalogParams map[string]interface{}) error {
	// TODO: Implement posting logic
	// - Create nalog entries
	// - Update izvod status to "proknjižen"
	// - Insert into appropriate ledger tables
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) OznaciKaoNeproknjizene(c *gin.Context, izvodIDs []string) error {
	// TODO: Implement status update logic
	// - Update status_izvoda to "neproknjižen" for selected izvodi
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) SetIzvodiFields() {
	s.izvodiHeaderTableFields = []domain.Fields{
		{Name: "izvbr", Label: "Broj izvoda", Width: "10", Field: "i.izvbr", SkipInSearch: false},
		{Name: "datizv", Label: "Datum izvoda", Width: "10", Field: "i.datizv", SkipInSearch: false},
		{Name: "brrac", Label: "Broj računa", Width: "15", Field: "i.brrac", SkipInSearch: false},
		{Name: "konto", Label: "Konto", Width: "10", Field: "i.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "10", Field: "i.sifra", SkipInSearch: false},
		{Name: "banka", Label: "Naziv banke", Width: "20", Field: "b.banka", SkipInSearch: false},
		{Name: "prstanje", Label: "Prethodno stanje", Width: "12", Field: "i.prstanje", SkipInSearch: true},
		{Name: "ukdug", Label: "Ukupno duguje", Width: "12", Field: "i.ukdug", SkipInSearch: true},
		{Name: "ukpot", Label: "Ukupno potražuje", Width: "12", Field: "i.ukpot", SkipInSearch: true},
		{Name: "nstanje", Label: "Novo stanje", Width: "12", Field: "i.nstanje", SkipInSearch: true},
		{Name: "ukbrst", Label: "Broj stavki", Width: "8", Field: "i.ukbrst", SkipInSearch: true},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "i.nalog", SkipInSearch: false},
		{Name: "tipdok", Label: "Tip dokumenta", Width: "6", Field: "i.tipdok", SkipInSearch: false},
		{Name: "izvsts", Label: "Status izvoda", Width: "12", Field: "i.izvsts", SkipInSearch: false},
		{Name: "idfizvzag", Label: "ID", Width: "8", Field: "i.idfizvzag", SkipInSearch: true},
	}

	s.izvodiDetailTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "6", Field: "d.rbr", SkipInSearch: false},
		{Name: "konto", Label: "Konto", Width: "10", Field: "d.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "10", Field: "d.sifra", SkipInSearch: false},
		{Name: "kat", Label: "Kategorija", Width: "4", Field: "d.kat", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta", Width: "4", Field: "d.vrd", SkipInSearch: false},
		{Name: "konto1", Label: "Protivkonto", Width: "10", Field: "d.konto1", SkipInSearch: false},
		{Name: "sifra1", Label: "Protivšifra", Width: "10", Field: "d.sifra1", SkipInSearch: false},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "d.iznos", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "d.duguje", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "d.potrazuje", SkipInSearch: true},
		{Name: "nsedprim", Label: "Naziv i sedište primaoca", Width: "20", Field: "d.nsedprim", SkipInSearch: false},
		{Name: "brracup", Label: "Broj računa uplatioca", Width: "15", Field: "d.brracup", SkipInSearch: false},
		{Name: "osnplac", Label: "Osnov plaćanja", Width: "10", Field: "d.osnplac", SkipInSearch: false},
		{Name: "sdozn", Label: "Svrha doznake", Width: "20", Field: "d.sdozn", SkipInSearch: false},
		{Name: "sdozn1", Label: "Svrha doznake 1", Width: "20", Field: "d.sdozn1", SkipInSearch: false},
		{Name: "modelzad", Label: "Model zaduženja", Width: "8", Field: "d.modelzad", SkipInSearch: false},
		{Name: "pnabrzad", Label: "Poziv na broj zaduženja", Width: "18", Field: "d.pnabrzad", SkipInSearch: false},
		{Name: "mododob", Label: "Model odobrenja", Width: "8", Field: "d.mododob", SkipInSearch: false},
		{Name: "pnabrodob", Label: "Poziv na broj odobrenja", Width: "18", Field: "d.pnabrodob", SkipInSearch: false},
		{Name: "prekl", Label: "Podaci sa reklame", Width: "18", Field: "d.prekl", SkipInSearch: false},
		{Name: "idfizvdet", Label: "ID", Width: "8", Field: "d.idfizvdet", SkipInSearch: true},
	}
}
func (s *IzvodiResource) GetMasterTableFields() []domain.Fields {
	return s.izvodiHeaderTableFields
}

func (s *IzvodiResource) GetDetailTableFields() []domain.Fields {
	return s.izvodiDetailTableFields
}

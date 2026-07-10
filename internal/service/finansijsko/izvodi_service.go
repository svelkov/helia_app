package finansijsko

import (
	"context"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"log"
	"strings"
	"time"
)

// IzvodiService defines the interface for operations related to bank statements (Izvodi).
type IzvodiService interface {
	GetMasterTableFields() []domain.Fields
	GetDetailTableFields() []domain.Fields
	GetIzvodiHeader(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.IzvodiParams) error
	GetIzvodiDetail(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.IzvodiParams) error
	ImportIzvod(ctx context.Context, fileData []string, software string) error
	AzurirajKonta(ctx context.Context, izvodID string) error
	BrisiIzvod(ctx context.Context, izvodID string) error
	ProveriRaznotezu(ctx context.Context, izvodID string) error
	KnjiziIzvod(ctx context.Context, izvodID string, nalogParams map[string]interface{}) error
	OznaciKaoNeproknjizene(ctx context.Context, izvodIDs []string) error
	GetBanke(ctx context.Context) ([]domain.ComboItem, error)
	GetTipdokOptions(ctx context.Context) ([]domain.ComboItem, error)
	GetNextNalog(ctx context.Context, tipdok string) (int64, error)
}

// IzvodiResource implements the IzvodiService interface.
type IzvodiResource struct {
	service                 *service.BaseService[domain.Fizvzag]
	izvhdrRepo              *repository.BaseRepository[domain.Fizvzag]
	izvdetRepo              *repository.BaseRepository[domain.Fizvdet]
	bankeRepo               *repository.BaseRepository[domain.Banke]
	tipdokRepo              *repository.BaseRepository[domain.Tipdok]
	fnalRepo                *repository.BaseRepository[domain.Fnal]
	cfg                     config.Config
	izvodiHeaderTableFields []domain.Fields
	izvodiDetailTableFields []domain.Fields
}

func NewIzvodiResource(izvhdrRepo *repository.BaseRepository[domain.Fizvzag], izvdetRepo *repository.BaseRepository[domain.Fizvdet],
	bankeRepo *repository.BaseRepository[domain.Banke], tipdokRepo *repository.BaseRepository[domain.Tipdok], fnalRepo *repository.BaseRepository[domain.Fnal], cfg config.Config) *IzvodiResource {
	rs := &IzvodiResource{
		izvhdrRepo: izvhdrRepo,
		izvdetRepo: izvdetRepo,
		bankeRepo:  bankeRepo,
		tipdokRepo: tipdokRepo,
		fnalRepo:   fnalRepo,
		cfg:        cfg,
	}
	rs.SetIzvodiFields()
	return rs
}

func (s *IzvodiResource) GetIzvodiHeader(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.IzvodiParams) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// Get parameters from form

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.izvhdrRepo.GetHasGodHasKar()

	// Build query with QueryBuilder matching WinDev SQL structure
	qb := common.NewQueryBuilder(`
		SELECT 
			i.idfizvzag, i.god, i.kar, i.brrac,
			i.izvbr, i.datizv, i.konto, i.sifra, i.prstanje,
			i.ukdug, i.ukpot, i.nstanje, i.ukbrst,
			i.nalog, i.tipdok, i.izvsts, i.idbanke, coalesce(b.banka,'') as banka, coalesce(b.brrac,'') as brrac,
			i.konto, i.sifra
		FROM fizvzag i`, true)

	if hasGod {
		qb.AddEqual("i.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("i.kar", session.SelectedKar)
	}
	qb.AddJoin(` LEFT JOIN banke b ON b.idbanke = i.idbanke `)
	if params.IDbanke != "" && params.IDbanke != "0" {
		qb.AddEqual("i.idbanke", params.IDbanke)
	}

	if params.BrojRacuna != "" {
		qb.AddEqual("i.brrac", params.BrojRacuna)
	}

	if params.OdBrojaIzvoda != "" {
		qb.AddCondition("i.izvbr", params.OdBrojaIzvoda, ">=")
	}

	if params.DoBrojaIzvoda != "" {
		qb.AddCondition("i.izvbr", params.DoBrojaIzvoda, "<=")
	}

	if params.OdDatumaIzvoda != "" {
		qb.AddCondition("i.datizv", params.OdDatumaIzvoda, ">=")
	}

	if params.DoDatumaIzvoda != "" {
		qb.AddCondition("i.datizv", params.DoDatumaIzvoda, "<=")
	}

	if params.StatusIzvoda != "" {
		qb.AddEqual("i.izvsts", params.StatusIzvoda)
	}

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.AddSearchConditions(s.GetMasterTableFields(), params.SearchText)
	}

	qb.AddOrderBy("i.datizv DESC, i.izvbr DESC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.izvhdrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%d", entity.IDFizvzag),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFizvzag), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

func (s *IzvodiResource) GetIzvodiDetail(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.IzvodiParams) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	qb := common.NewQueryBuilder(`
		SELECT 
			d.idfizvdet, d.rbr, d.konto, d.sifra, d.kat,
			d.vrd, d.konto1, d.sifra1, d.iznos, d.duguje,
			d.potrazuje, d.nsedprim, d.brracup, d.osnplac,
			d.sdozn, d.sdozn1, d.modelzad, d.pnabrzad,
			d.mododob, d.pnabrodob, d.prekl
		FROM fizvdet d`, true)

	qb.AddEqual("d.idfizvzag", params.IDfizvzag)

	// Add search conditions if search text is provided
	if params.SearchText != "" {
		qb.AddSearchConditions(s.GetDetailTableFields(), params.SearchText)
	}

	qb.AddOrderBy("d.rbr")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.izvdetRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFizvdet), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

func (s *IzvodiResource) GetBanke(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.bankeRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idbanke, banka, bnkcod FROM banke`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	sqlQuery, args := qb.Build()
	banke, err := s.bankeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	var comboItems []domain.ComboItem
	for _, banka := range *banke {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", banka.IDBanKe),
			Value: fmt.Sprintf("%s - %s", banka.BnkCod, banka.Banka),
		})
	}
	return comboItems, nil
}

// GetTipdokOptions fetches the list of tipdok options for filtering. This method stays the same.
func (s *IzvodiResource) GetTipdokOptions(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idtipdok, tipdok, opis FROM tipdok`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddCustomCondition("(grpdok = 'FIN' OR grpdok = 'SVI')")
	qb.AddOrderBy("tipdok::NUMERIC ASC")
	sqlQuery, args := qb.Build()
	entites, err := s.tipdokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tipdok options: %w", err)
	}
	var comboItems []domain.ComboItem
	for _, entity := range *entites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   entity.TipDok,
			Value: fmt.Sprintf("%s - %s", entity.TipDok, entity.Opis),
		})
	}
	return comboItems, nil
}

// Update implements NalogService.
func (s *IzvodiResource) GetNextNalog(ctx context.Context, tipdok string) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT COALESCE(MAX(nalog), 0) + 1 as nalog FROM fnal`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) == 0 {
		return 1, nil
	}

	return (*entities)[0].Nalog, nil

}

func (s *IzvodiResource) ImportIzvod(ctx context.Context, fileData []string, software string) error {
	//var izvodiData any
	switch software {
	case "halcom":
		// Call the import function for Halcom software
		err := s.ImportDATAHALCOM(ctx, fileData)
		if err != nil {
			return fmt.Errorf("failed to import Halcom data: %w", err)
		}
	}

	return fmt.Errorf("unsupported software: %s", software)
}

func (s *IzvodiResource) ImportDATAHALCOM(ctx context.Context, fileData []string) error {
	izvodHalcom := []domain.IzvodHalcom{}
	usrSession := domain.GetSessionFromStdContext(ctx)
	if usrSession == nil {
		return fmt.Errorf("user session not found")
	}
	god, kar := 0, 0
	hasGod, hasKar := s.izvhdrRepo.GetHasGodHasKar()
	if hasGod {
		god = usrSession.SelectedGod
	}
	if hasKar {
		kar = usrSession.SelectedKar
	}
	for i, line := range fileData {
		// Skip the header line if present
		if i == 0 {
			continue
		}
		// Parse each line according to Halcom format
		izvod, err := ParseIzvodHalcomLine(line)
		if err != nil {
			log.Printf("Error parsing line %d: %v", i+1, err)
			continue
		}
		izvodHalcom = append(izvodHalcom, *izvod)
	}
	// create header and detail of izvodi
	tx, err := s.izvhdrRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer handles both panic and normal errors
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			// Convert panic to error or re-panic
			err = fmt.Errorf("panic recovered: %v", r)
			// Or re-panic if you want to crash:
			// panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	for _, izvod := range izvodHalcom {
		var newIdfizvzag int64
		izvodHdrID, err := s.getIzvodHeader(ctx, god, kar, izvod.BrRacuna, izvod.BrIzvoda, izvod.DatumObrada)
		if err != nil {
			return fmt.Errorf("failed to get izvod header: %w", err)
		}
		// If the header does not exist, insert it
		if izvodHdrID == 0 {
			qb := common.NewQueryBuilder(`INSERT INTO fizvzag (god, kar, brrac, datizv, izvbr, valuta, prstanje, ukdug, ukpot,
			 nstanje, ukbrst, nalog, tipdok, izvsts)
										VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING idfizvzag`, false)
			qb.AddArgs(god, kar, izvod.BrRacuna, izvod.DatumObrada, izvod.BrIzvoda, 0, 0, 0, 0, 0, 0, 0, "", 1)
			sqlQuery, args := qb.Build()
			err = tx.QueryRowContext(ctx, sqlQuery, args...).Scan(&newIdfizvzag)
			if err != nil {
				return fmt.Errorf("failed to insert izvod: %w", err)
			}
		}
	}
	return nil
}
func (s *IzvodiResource) ImportDATAOTP(ctx context.Context, fileData []string) error {

	return nil
}

func (s *IzvodiResource) getIzvodHeader(ctx context.Context, god, kar int, brRacuna string, brIzvoda string, datumObrade time.Time) (int64, error) {
	qb := common.NewQueryBuilder(`SELECT fiizvzagid FROM fizvzag `, true)
	if god > 0 {
		qb.AddEqual("god", god)
	}
	if kar > 0 {
		qb.AddEqual("kar", kar)
	}
	qb.AddEqual("brrac", brRacuna)
	qb.AddEqual("izvbr", brIzvoda)
	qb.AddEqual("datizv", datumObrade)

	sqlQuery, args := qb.Build()
	entities, err := s.izvhdrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to insert izvod: %w", err)
	}
	if entities != nil && len(*entities) > 0 {
		return int64((*entities)[0].IDFizvzag), nil
	}
	return 0, nil
}

// ParseIzvodHalcomLine parses a single line and returns IzvodHalcom
func ParseIzvodHalcomLine(line string) (*domain.IzvodHalcom, error) {
	fields := strings.Split(line, "#")

	// Expected number of fields (28 in your case)
	const expectedFields = 28
	if len(fields) < expectedFields {
		return nil, fmt.Errorf("expected %d fields, got %d", expectedFields, len(fields))
	}

	izvod := &domain.IzvodHalcom{}
	// Map fields by index - this is fast and clear
	izvod.BrRacuna = strings.TrimSpace(fields[0])
	izvod.DatumObrada = common.StringToDate(strings.TrimSpace(fields[1]))
	izvod.BrIzvoda = strings.TrimSpace(fields[2])
	izvod.Valuta = strings.TrimSpace(fields[3])
	izvod.DatumValute = common.StringToDate(strings.TrimSpace(fields[4]))
	izvod.IznosZaduzenja = common.FormatFloatNumber64WithSystemLocale(common.StringToFloat64(strings.TrimSpace(fields[5])), 2)
	izvod.IznosOdobrenja = common.FormatFloatNumber64WithSystemLocale(common.StringToFloat64(strings.TrimSpace(fields[6])), 2)
	izvod.OznakaKnjizenja = strings.TrimSpace(fields[7])
	izvod.Opis = strings.TrimSpace(fields[8])
	izvod.DatumKnjizenja = common.StringToDate(strings.TrimSpace(fields[9]))
	izvod.RacunPartnera = strings.TrimSpace(fields[10])
	izvod.NazivPartnera = strings.TrimSpace(fields[11])
	izvod.Svrha = strings.TrimSpace(fields[12])
	izvod.OznakaVrstePosla = strings.TrimSpace(fields[13])
	izvod.PozivNaBrojOdobrenja = strings.TrimSpace(fields[14])
	izvod.PozivNaBrojZaduzenja = strings.TrimSpace(fields[15])
	izvod.ModelOdobrenja = strings.TrimSpace(fields[16])
	izvod.ModelZaduzenja = strings.TrimSpace(fields[17])
	izvod.IDNaloga = strings.TrimSpace(fields[18])
	izvod.VremeNastanka = strings.TrimSpace(fields[19])
	izvod.VremePrijema = strings.TrimSpace(fields[20])
	izvod.LeviPotpisnik = strings.TrimSpace(fields[21])
	izvod.DesniPotpisnik = strings.TrimSpace(fields[22])
	izvod.DatumValuteNaloga = strings.TrimSpace(fields[23])
	izvod.DatumPripreme = strings.TrimSpace(fields[24])
	izvod.TipNaloga = strings.TrimSpace(fields[25])
	izvod.Hitni = strings.TrimSpace(fields[26])
	izvod.ReferencaBanke = strings.TrimSpace(fields[27])

	return izvod, nil
}

func (s *IzvodiResource) AzurirajKonta(ctx context.Context, izvodID string) error {
	// TODO: Implement account update logic
	// - Update accounts in izvodi_detalji based on parameters
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) BrisiIzvod(ctx context.Context, izvodID string) error {
	// TODO: Implement delete logic
	// - Delete from izvodi_detalji first
	// - Delete from izvodi
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) ProveriRaznotezu(ctx context.Context, izvodID string) error {
	// TODO: Implement balance check logic
	// - Calculate total debits and credits
	// - Check if they match
	// - Return error if unbalanced
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) KnjiziIzvod(ctx context.Context, izvodID string, nalogParams map[string]interface{}) error {
	// TODO: Implement posting logic
	// - Create nalog entries
	// - Update izvod status to "proknjižen"
	// - Insert into appropriate ledger tables
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) OznaciKaoNeproknjizene(ctx context.Context, izvodIDs []string) error {
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

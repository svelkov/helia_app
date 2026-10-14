package finansijsko

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"helia/internal/service"
	commonsvc "helia/internal/service/common"
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
	ImportIzvod(ctx context.Context, fileData string, software string, prekoJMBG bool) error
	AzuriranjeKonta(ctx context.Context, fizvhdrID int64, povezivanjeJMBG bool, azurirajKonta bool) error
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
	fin                     *repository.FinansijskoRepository
	izvhdrRepo              *repository.BaseRepository[domain.Fizvzag]
	izvdetRepo              *repository.BaseRepository[domain.Fizvdet]
	bankeRepo               *repository.BaseRepository[domain.Banke]
	tipdokRepo              *repository.BaseRepository[domain.Tipdok]
	fnalRepo                *repository.BaseRepository[domain.Fnal]
	partneriRepo            *repository.BaseRepository[domain.Partneri]
	tekracuniRepo           *repository.BaseRepository[domain.TekRacuni]
	sifplizvRepo            *repository.BaseRepository[domain.Sifplizv]
	fkplRepo                *repository.BaseRepository[domain.Fkpl]
	fvrRepo                 *repository.BaseRepository[domain.Fvr]
	cfg                     config.Config
	commonSvc               commonsvc.CommonService
	izvodiHeaderTableFields []domain.Fields
	izvodiDetailTableFields []domain.Fields
}

func NewIzvodiResource(izvhdrRepo *repository.BaseRepository[domain.Fizvzag], izvdetRepo *repository.BaseRepository[domain.Fizvdet],
	bankeRepo *repository.BaseRepository[domain.Banke], tipdokRepo *repository.BaseRepository[domain.Tipdok], fnalRepo *repository.BaseRepository[domain.Fnal], partneriRepo *repository.BaseRepository[domain.Partneri], tekracuniRepo *repository.BaseRepository[domain.TekRacuni], sifplizvRepo *repository.BaseRepository[domain.Sifplizv], fkplRepo *repository.BaseRepository[domain.Fkpl], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config, commonSvc commonsvc.CommonService) *IzvodiResource {
	rs := &IzvodiResource{
		fin:           repository.NewFinansijskoRepository(izvhdrRepo.DB),
		izvhdrRepo:    izvhdrRepo,
		izvdetRepo:    izvdetRepo,
		bankeRepo:     bankeRepo,
		tipdokRepo:    tipdokRepo,
		fnalRepo:      fnalRepo,
		partneriRepo:  partneriRepo,
		tekracuniRepo: tekracuniRepo,
		sifplizvRepo:  sifplizvRepo,
		fkplRepo:      fkplRepo,
		fvrRepo:       fvrRepo,
		cfg:           cfg,
		commonSvc:     commonSvc,
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

	// Build query with QueryBuilder matching WinDev SQL structure
	b := querybuilder.New(`
		SELECT
			i.idfizvzag, i.god, i.kar, i.brrac,
			i.izvbr, i.datizv, i.konto, i.sifra, i.prstanje,
			i.ukdug, i.ukpot, i.nstanje, i.ukbrst,
			i.nalog, coalesce(i.tipdok,'') as tipdok, i.izvsts,
			i.idbanke, coalesce(b.banka,'') as banka, coalesce(b.brrac,'') as brrac,
			i.konto, i.sifra
		FROM fizvzag i`).
		AddJoin(`LEFT JOIN banke b ON b.idbanke = i.idbanke`)
	if params.IDbanke != "0" {
		b.AddEqual("i.idbanke", params.IDbanke)
	}
	// The optional filters of the form skip empty values by themselves.
	b.AddEqual("i.brrac", params.BrojRacuna).
		AddCondition("i.izvbr", params.OdBrojaIzvoda, ">=").
		AddCondition("i.izvbr", params.DoBrojaIzvoda, "<=").
		AddCondition("i.datizv", params.OdDatumaIzvoda, ">=").
		AddCondition("i.datizv", params.DoDatumaIzvoda, "<=").
		AddEqual("i.izvsts", params.StatusIzvoda)
	// The period of the session (i.god/kar) and the search over the columns of the grid (all
	// qualified, because the joined banke has konto/sifra/brrac too). The old builder had no
	// entity type here, so its search never filtered.
	if err := s.izvhdrRepo.ApplyFilters(ctx, b, "i", s.GetMasterTableFields(), params.SearchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.izvhdrRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("i.datizv DESC, i.izvbr DESC").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.izvhdrRepo.Select(ctx, b)
	if err != nil {
		return err
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
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFizvzag), Fields: fields, HasUpdate: false, HasDelete: true}
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
	// The header id comes from the request: a key (Where), restricted to the period of the session,
	// plus the search over the columns of the grid (the old builder had no entity type here, so
	// its search never filtered).
	b := querybuilder.New(`
		SELECT
			d.idfizvdet, d.rbr, d.konto, d.sifra, d.kat,
			d.vrd, coalesce(d.konto1, '') as konto1, coalesce(d.sifra1, '') as sifra1, d.iznos, d.duguje,
			d.potrazuje, d.nsedprim, d.brracup, d.osnplac,
			coalesce(d.sdozn, '') as sdozn, coalesce(d.sdozn1, '') as sdozn1, coalesce(d.modelzad, '') as modelzad, coalesce(d.pnabrzad, '') as pnabrzad,
			coalesce(d.mododob, '') as mododob, coalesce(d.pnabrodob, '') as pnabrodob, coalesce(d.prekl, '') as prekl
		FROM fizvdet d`).
		Where("d.idfizvzag = ?", params.IDfizvzag)
	if err := s.izvdetRepo.ApplyFilters(ctx, b, "d", s.GetDetailTableFields(), params.SearchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.izvdetRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("d.rbr").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.izvdetRepo.Select(ctx, b)
	if err != nil {
		return err
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

// GetBanke returns the banke of the current period (CommonService).
func (s *IzvodiResource) GetBanke(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetBankeComboValues(ctx)
}

// GetTipdokOptions returns the vrste naloga that can be used for knjiženje izvoda
// (grpdok FIN/SVI) keyed by the tipdok code (CommonService).
func (s *IzvodiResource) GetTipdokOptions(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetTipdokFinComboValues(ctx)
}

// GetNextNalog returns the next broj naloga of the given vrsta naloga (CommonService).
func (s *IzvodiResource) GetNextNalog(ctx context.Context, tipdok string) (int64, error) {
	return s.commonSvc.GetNextFnalNalog(ctx, tipdok)
}

func (s *IzvodiResource) ImportIzvod(ctx context.Context, fileData string, software string, prekoJMBG bool) error {
	//var izvodiData any
	switch software {
	case "halcom":
		// Call the import function for Halcom software
		err := s.ImportDATAHALCOM(ctx, fileData, prekoJMBG)
		if err != nil {
			return fmt.Errorf("failed to import Halcom data: %w", err)
		}
	case "trezor":
		// Call the import function for Trezor software
		err := s.ImportDATATREZOR(ctx, fileData, prekoJMBG)
		if err != nil {
			return fmt.Errorf("failed to import Trezor data: %w", err)
		}
	case "otp", "societe":
		// Call the import function for OTP or Societe software
		err := s.ImportDATAOTP(ctx, fileData, prekoJMBG)
		if err != nil {
			return fmt.Errorf("failed to import OTP or Societe data: %w", err)
		}
	default:
		return fmt.Errorf("unsupported software: %s", software)
	}

	return nil
}

func (s *IzvodiResource) AzuriranjeKonta(ctx context.Context, fizvhdrID int64, povezivanjeJMBG bool, azurirajKonta bool) error {

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	// The header id comes from the request: a key (Where), restricted to the period of the session.
	// osnplac, brracup and pnabrodob are the inputs of the lookups below.
	b := querybuilder.New(`SELECT idfizvdet, rbr, konto, sifra, kat, vrd, iznos, duguje, potrazuje,
		osnplac, brracup, pnabrodob FROM fizvdet`).
		Where("idfizvzag = ?", fizvhdrID)
	if err := s.izvdetRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return err
	}
	entities, err := s.izvdetRepo.Select(ctx, b)
	if err != nil {
		return fmt.Errorf("failed to get Fizvdet entities: %w", err)
	}

	// setKontoSifra applies a found konto/sifra: always when azurirajKonta, otherwise only the
	// empty ones.
	setKontoSifra := func(fizvdet *domain.Fizvdet, konto, sifra string) {
		if azurirajKonta || fizvdet.Konto == "" {
			fizvdet.Konto = konto
		}
		if azurirajKonta || fizvdet.Sifra == "" {
			fizvdet.Sifra = sifra
		}
	}

	// Every lookup is built for its own row.
	for _, fizvdet := range *entities {
		// Step 1: Look up SIFPLIZV by OSNPLAC
		if osnplac := strings.TrimSpace(fizvdet.Osnplac); osnplac != "" {
			bOsn := querybuilder.New(`SELECT sifplac, konto, sifra FROM sifplizv`).Where("sifplac = ?", osnplac)
			sifplEntites, err := s.sifplizvRepo.Select(ctx, bOsn)
			if err != nil {
				return fmt.Errorf("osnov placanja %s: %w", osnplac, err)
			}
			if len(*sifplEntites) > 0 {
				setKontoSifra(&fizvdet, (*sifplEntites)[0].Konto, (*sifplEntites)[0].Sifra)
			}
		}

		// Step 2: Determine partner ID based on JMBG or account number
		nIDPARTNERI, err := s.azuriranjePartnerID(ctx, fizvdet, povezivanjeJMBG)
		if err != nil {
			return err
		}

		// Step 3: konto dobavljaca for DUGUJE, Step 4: konto kupca for POTRAZUJE - the analiticki
		// konto of that partner in the kontni plan.
		if nIDPARTNERI > 0 && fizvdet.Duguje != 0 {
			fkpl, err := s.azuriranjeFkplPartnera(ctx, s.cfg.Konta.KontoDobavljaca, nIDPARTNERI)
			if err != nil {
				return err
			}
			if fkpl != nil {
				setKontoSifra(&fizvdet, fkpl.Konto, fkpl.Sifra)
				fizvdet.Vrd = 40
				fizvdet.Kat = 1
			}
		}
		if nIDPARTNERI > 0 && fizvdet.Potrazuje != 0 {
			fkpl, err := s.azuriranjeFkplPartnera(ctx, s.cfg.Konta.KontoKupca, nIDPARTNERI)
			if err != nil {
				return err
			}
			if fkpl != nil {
				setKontoSifra(&fizvdet, fkpl.Konto, fkpl.Sifra)
				fizvdet.Vrd = 30
				fizvdet.Kat = 3
			}
		}
		// At the end of loop update fizvdet record in the database
		// The idfizvdet comes from the rows of the period read above.
		bUpdate := querybuilder.NewUpdate("fizvdet").
			Set("konto = ?, sifra = ?, kat = ?, vrd = ?", fizvdet.Konto, fizvdet.Sifra, fizvdet.Kat, fizvdet.Vrd).
			Where("idfizvdet = ?", fizvdet.IDFizvdet)
		if err := s.izvdetRepo.Exec(ctx, bUpdate); err != nil {
			return fmt.Errorf("failed to update Fizvdet: %w", err)
		}
	}

	return nil
}

// azuriranjePartnerID finds the partner of a stavka izvoda in the tenant: by the JMBG in the
// poziv na broj odobrenja, or by the account number of the uplatilac. 0 when not found.
func (s *IzvodiResource) azuriranjePartnerID(ctx context.Context, fizvdet domain.Fizvdet, povezivanjeJMBG bool) (int64, error) {
	if povezivanjeJMBG {
		// Extract JMBG from PNABRODOB field: up to 13 characters as is, otherwise the part before
		// the first "-" (the part after it is the payment code 01 or 02).
		pnabrodob := strings.TrimSpace(fizvdet.Pnabrodob)
		sJMBG := pnabrodob
		if len(pnabrodob) > 13 {
			sJMBG = strings.TrimSpace(strings.Split(pnabrodob, "-")[0])
		}
		if sJMBG == "" {
			return 0, nil
		}
		b := querybuilder.New(`SELECT idpartneri, god, kar, jmbg FROM partneri`).Where("jmbg = ?", sJMBG)
		if err := s.partneriRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
			return 0, err
		}
		partneri, err := s.partneriRepo.Select(ctx, b)
		if err != nil {
			return 0, fmt.Errorf("partner po JMBG: %w", err)
		}
		if len(*partneri) > 0 {
			return int64((*partneri)[0].IDPartneri), nil
		}
		return 0, nil
	}
	// Look up by account number from TEKRACUNI
	if fizvdet.Brracup == "" {
		return 0, nil
	}
	// (The account column of tekracuni is tekrac; the old lookup used brrac, which is not a column of
	// tekracuni, so it always failed and the error was ignored.)
	b := querybuilder.New(`SELECT idpartneri, tekrac FROM tekracuni`).Where("tekrac = ?", fizvdet.Brracup)
	if err := s.tekracuniRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return 0, err
	}
	tekracuni, err := s.tekracuniRepo.Select(ctx, b)
	if err != nil {
		return 0, fmt.Errorf("partner po tekucem racunu %s: %w", fizvdet.Brracup, err)
	}
	if len(*tekracuni) > 0 {
		return int64((*tekracuni)[0].IDPartneri), nil
	}
	return 0, nil
}

// azuriranjeFkplPartnera returns the analiticki konto of the partner on the given konto in the
// kontni plan of the period (nil when there is none).
func (s *IzvodiResource) azuriranjeFkplPartnera(ctx context.Context, konto string, idPartneri int64) (*domain.Fkpl, error) {
	// (vrd is not a column of domain.Fkpl; selecting it made the old lookup fail.)
	b := querybuilder.New(`SELECT idfkpl, god, kar, konto, sifra, idpartneri FROM fkpl`).
		Where("konto = ?", konto).
		Where("idpartneri = ?", idPartneri)
	if err := s.fkplRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return nil, err
	}
	fkplEntities, err := s.fkplRepo.Select(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("konto %s partnera %d: %w", konto, idPartneri, err)
	}
	if len(*fkplEntities) == 0 {
		return nil, nil
	}
	return &(*fkplEntities)[0], nil
}

// UpisIzvoda inserts the header and detail records into the database.
func (s *IzvodiResource) UpisIzvoda(ctx context.Context, izvHdr *domain.Fizvzag, izvDet []domain.Fizvdet) error {
	//check ig the izvod exist and is already booked, if so return error
	// The period of the header being imported (izvHdr.God/Kar); the bank account and the number
	// of the izvod are keys.
	hasGod, hasKar := s.izvhdrRepo.GetHasGodHasKar()
	b := querybuilder.New(`SELECT idfizvzag, izvsts FROM fizvzag`).
		Where("brrac = ?", izvHdr.Brrac).
		Where("izvbr = ?", izvHdr.Izvbr)
	if hasGod {
		b.Where("god = ?", izvHdr.God)
	}
	if hasKar {
		b.Where("kar = ?", izvHdr.Kar)
	}
	entities, err := s.izvhdrRepo.Select(ctx, b)
	IdFizvzag := int64(0)
	if err != nil {
		return fmt.Errorf("failed to get Fizvzag entities: %w", err)
	}
	if len(*entities) > 0 {
		IdFizvzag = (*entities)[0].IDFizvzag
		if (*entities)[0].Izvsts == "40" {
			return fmt.Errorf("izvod sa brojem %d vec postoji i proknjizen je", izvHdr.Izvbr)
		}
	}

	tx, err := s.izvhdrRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Izvod postoji uradi update, inace insert
	if IdFizvzag > 0 {
		// The idfizvzag was found above in the period of the header.
		bUpdate := querybuilder.NewUpdate("fizvzag").
			Set(`datizv = ?, konto = ?, sifra = ?, prstanje = ?, ukdug = ?,
		ukpot = ?, nstanje = ?, ukbrst = ?, nalog = ?, tipdok = ?, izvsts = ?,
		 idbanke = ?, xdatizmene = NOW(), xopizmene = ?`, izvHdr.Datizv, izvHdr.Konto, izvHdr.Sifra,
				izvHdr.Prstanje, izvHdr.Ukdug, izvHdr.Ukpot, izvHdr.Nstanje,
				izvHdr.Ukbrst, izvHdr.Nalog, izvHdr.Tipdok, izvHdr.Izvsts, izvHdr.IDbanke, izvHdr.Xopunos).
			Where("idfizvzag = ?", IdFizvzag)
		err = s.izvhdrRepo.ExecWithTx(ctx, tx, bUpdate)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update Fizvzag: %w", err)
		}
	} else {
		IdFizvzag, err = s.fin.FizvzagDodaj(ctx, tx, izvHdr)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert Fizvzag: %w", err)
		}
	}
	// Delete existing detail records for the header (in its period) before inserting new ones
	err = s.fin.FizvdetBrisi(ctx, tx, IdFizvzag, izvHdr.God, izvHdr.Kar)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing detail records: %w", err)
	}
	err = s.fin.FizvdetDodaj(ctx, tx, izvDet, IdFizvzag)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert detail records: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// ImportDATATREZOR imports bank statement data from Trezor software.
func (s *IzvodiResource) ImportDATATREZOR(ctx context.Context, fileData string, prekoJMBG bool) error {
	izvodTrezor := domain.TrezorDokument{}
	err := xml.Unmarshal([]byte(fileData), &izvodTrezor)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Trezor data: %w", err)
	}
	izvodHdr, izvodDet, err := s.MapTrezorToDomain(ctx, izvodTrezor, prekoJMBG)
	if err != nil {
		return fmt.Errorf("failed to map Trezor data to domain: %w", err)
	}
	return s.UpisIzvoda(ctx, izvodHdr, izvodDet)
}

// ImportDATAHALCOM imports bank statement data from Halcom software.
func (s *IzvodiResource) ImportDATAHALCOM(ctx context.Context, fileData string, prekoJMBG bool) error {
	izvodHalcom := []domain.IzvodHalcom{}
	usrSession := domain.GetSessionFromStdContext(ctx)
	if usrSession == nil {
		return fmt.Errorf("user session not found")
	}
	fileLines := strings.Split(fileData, "\n")
	for i, line := range fileLines {
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
	if len(izvodHalcom) == 0 {
		return fmt.Errorf("no valid records found in the file")
	}
	izvHdr, izvDet, err := s.MapHalcomToDomain(ctx, izvodHalcom, prekoJMBG)
	if err != nil {
		return fmt.Errorf("failed to map Halcom data to domain: %w", err)
	}
	return s.UpisIzvoda(ctx, izvHdr, izvDet)
}

// ImportDATAOTP imports bank statement data from OTP or Societe software.
func (s *IzvodiResource) ImportDATAOTP(ctx context.Context, fileData string, prekoJMBG bool) error {
	izvod := domain.StmtRsList{}
	err := xml.Unmarshal([]byte(fileData), &izvod)
	if err != nil {
		return fmt.Errorf("failed to unmarshal OTP data: %w", err)
	}
	izvHdr, izvDet, err := s.MapIntesaToDomain(ctx, izvod, prekoJMBG)
	if err != nil {
		return fmt.Errorf("failed to map OTP ili Societe data to domain: %w", err)
	}
	return s.UpisIzvoda(ctx, izvHdr, izvDet)
}

// MapTrezorToDomain maps the Trezor data to domain.Fizvzag and domain.Fizvdet.
func (s *IzvodiResource) MapTrezorToDomain(ctx context.Context, izvodTrezor domain.TrezorDokument, prekoJMBG bool) (*domain.Fizvzag, []domain.Fizvdet, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, nil, fmt.Errorf("user session not found")
	}
	// Map the TrezorDokument to domain.Fizvzag and domain.Fizvdet
	izvodHdr := &domain.Fizvzag{
		God:       userSession.SelectedGod,
		Kar:       userSession.SelectedKar,
		Brrac:     izvodTrezor.Zbirni.RacunIzvoda,
		Izvbr:     common.StringToInt(izvodTrezor.Zbirni.BrojIzvoda),
		Datizv:    sql.NullTime{Time: common.StringToDate(izvodTrezor.Zaglavlje.DatumIzvoda, common.DateLayout), Valid: true},
		Prstanje:  izvodTrezor.Zbirni.PrethodniSaldo,
		Ukdug:     izvodTrezor.Zbirni.KumulativnoDuguje,
		Ukpot:     izvodTrezor.Zbirni.KumulativnoPotrazuje,
		Ukbrst:    len(izvodTrezor.Stavke),
		Konto:     "",
		Sifra:     "",
		Izvsts:    "1", // status 1 -importovan, status 40 - proknjizen
		Xopunos:   sql.NullString{String: userSession.UserName, Valid: true},
		Xdatunosa: sql.NullTime{Time: time.Now(), Valid: true},
	}
	// Update bank info
	bank, err := s.getBankInfo(ctx, userSession.SelectedGod, userSession.SelectedKar, izvodTrezor.Zbirni.RacunIzvoda)
	if err == nil && bank != nil {
		izvodHdr.IDbanke = sql.NullInt64{Int64: int64(bank.IDBanke), Valid: true}
		izvodHdr.Konto = bank.Konto
		izvodHdr.Sifra = bank.Sifra
	}

	i := int64(0)
	izvDet := []domain.Fizvdet{}
	ukDug, ukPot := float64(0), float64(0)
	for _, stavka := range izvodTrezor.Stavke {
		i++
		konto, sifra := "", ""
		kat := int16(0)
		dug, pot := float64(0), float64(0)
		brRacDet := ""
		firstChar := string([]rune(stavka.IzvorInformacije)[0])

		izvodDet := domain.Fizvdet{
			God:       userSession.SelectedGod,
			Kar:       userSession.SelectedKar,
			Brrac:     izvodHdr.Brrac,
			Izvbr:     common.StringToInt(izvodTrezor.Zbirni.BrojIzvoda),
			Datizv:    sql.NullTime{Time: common.StringToDate(izvodTrezor.Zaglavlje.DatumIzvoda, common.DateLayout), Valid: true},
			Rbr:       i,
			Konto:     konto,
			Sifra:     sifra,
			Iznos:     stavka.Iznos,
			Kat:       kat,
			Vrd:       90,
			Duguje:    dug,
			Potrazuje: pot,
			Nsedprim:  stavka.MestoZaduzenja,
			Brracup:   brRacDet,
			Osnplac:   stavka.SifraPlacanja,
			Sdozn:     "",
			Sdozn1:    stavka.SvrhaDoznake,
			Modelzad:  stavka.ModelPozivaZaduzenja,
			Pnabrzad:  stavka.PozivZaduzenja,
			Mododob:   stavka.ModelPozivaOdobrenja,
			Pnabrodob: stavka.PozivOdobrenja,
			Prekl:     stavka.PodatakZaReklamaciju,
		}
		if firstChar == "1" {
			izvodDet.Nsedprim = stavka.NazivOdobrenja + " " + stavka.MestoOdobrenja
			izvodDet.Duguje = stavka.Iznos
			izvodDet.Brracup = stavka.RacunOdobrenja
			izvodDet.Potrazuje = 0
			izvodDet.Kat = 1
			ukDug += stavka.Iznos
		} else {
			izvodDet.Nsedprim = stavka.NazivZaduzenja + " " + stavka.MestoZaduzenja
			izvodDet.Duguje = 0
			izvodDet.Potrazuje = stavka.Iznos
			izvodDet.Kat = 3
			izvodDet.Brracup = stavka.RacunZaduzenja
			ukPot += stavka.Iznos
		}
		//get konto and sifra from tekracuni table based on brRacDet
		konto, sifra, err := s.getSifplIzvodiInfo(ctx, common.StringToInt(stavka.SifraPlacanja))
		if err != nil {
			log.Printf("failed to get konto and sifra: %v", err)
		}
		if konto != "" {
			izvodDet.Konto = konto
		}
		if sifra != "" {
			izvodDet.Sifra = sifra
		}
		partnerID := s.getPartneriID(ctx, prekoJMBG, "", brRacDet)
		konto, sifra = s.getPartnerKontoSifra(ctx, partnerID, izvodDet.Duguje, izvodDet.Potrazuje)
		if konto != "" {
			izvodDet.Konto = konto
		}
		if sifra != "" {
			izvodDet.Sifra = sifra
		}
		izvodDet.Xopunos = sql.NullString{String: userSession.UserName, Valid: true}
		izvodDet.Xdatunosa = sql.NullTime{Time: time.Now(), Valid: true}
		// add the izvodDet to the list
		izvDet = append(izvDet, izvodDet)
	}
	izvodHdr.Ukdug = ukDug
	izvodHdr.Ukpot = ukPot
	izvodHdr.Nstanje = izvodHdr.Prstanje + ukDug - ukPot
	izvodHdr.Ukbrst = len(izvDet)
	return izvodHdr, izvDet, nil
}

// MapHalcomToDomain maps the Halcom data to domain.Fizvzag and domain.Fizvdet.
func (s *IzvodiResource) MapHalcomToDomain(ctx context.Context, izvodHalcom []domain.IzvodHalcom, prekoJMBG bool) (*domain.Fizvzag, []domain.Fizvdet, error) {
	if len(izvodHalcom) == 0 {
		return nil, nil, fmt.Errorf("no records to map")
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, nil, fmt.Errorf("user session not found")
	}
	izvodFirst := izvodHalcom[0]
	// Map the HalcomData to domain.Fizvzag and domain.Fizvdet
	izvodHdr := &domain.Fizvzag{
		God:       userSession.SelectedGod,
		Kar:       userSession.SelectedKar,
		Brrac:     izvodFirst.BrRacuna,
		Izvbr:     common.StringToInt(strings.TrimSpace(izvodFirst.BrIzvoda)),
		Datizv:    sql.NullTime{Time: common.StringToDate(izvodFirst.DatumKnjizenja, common.HtmlLayout), Valid: true},
		Konto:     "",
		Sifra:     "",
		Prstanje:  0,
		Izvsts:    "1",
		Xdatunosa: sql.NullTime{Time: time.Now(), Valid: true},
		Xopunos:   sql.NullString{String: userSession.UserName, Valid: true},
	}
	// Update bank info
	bank, err := s.getBankInfo(ctx, userSession.SelectedGod, userSession.SelectedKar, izvodFirst.BrRacuna)
	if err == nil && bank != nil {
		izvodHdr.IDbanke = sql.NullInt64{Int64: int64(bank.IDBanke), Valid: true}
		izvodHdr.Konto = bank.Konto
		izvodHdr.Sifra = bank.Sifra
	}
	// Map the HalcomData to domain.Fizvdet
	izvDet := []domain.Fizvdet{}
	totDug, totPot := 0.0, 0.0
	i := 0
	for _, row := range izvodHalcom {
		i++
		dug, pot, iznos := 0.0, 0.0, 0.0
		konto, sifra := "", ""
		vrd := 90
		kat := 0
		if row.IznosZaduzenja > 0 {
			dug = row.IznosZaduzenja
			iznos = row.IznosZaduzenja
			kat = 1
			totDug += dug
		}
		if row.IznosOdobrenja > 0 {
			pot = row.IznosOdobrenja
			iznos = row.IznosOdobrenja
			kat = 3
			totPot += pot
		}
		konto, sifra, err := s.getSifplIzvodiInfo(ctx, common.StringToInt(row.OznakaVrstePosla))
		if err != nil {
			log.Printf("Error fetching sifplizv info for OznakaVrstePosla %s: %v", row.OznakaVrstePosla, err)
		}
		partnerID := s.getPartneriID(ctx, prekoJMBG, row.PozivNaBrojOdobrenja, row.RacunPartnera)
		if partnerID != 0 {
			konto, sifra = s.getPartnerKontoSifra(ctx, partnerID, dug, pot)
		}
		if konto != "" && sifra != "" {
			if dug > 0 {
				vrd = 40
				kat = 1
			}
			if pot > 0 {
				vrd = 30
				kat = 3
			}
		}
		fizvdet := domain.Fizvdet{
			God:       userSession.SelectedGod,
			Kar:       userSession.SelectedKar,
			Brrac:     row.BrRacuna,
			Izvbr:     common.StringToInt(row.BrIzvoda),
			Rbr:       int64(i),
			Datizv:    sql.NullTime{Time: common.StringToDate(row.DatumObrada, common.HtmlLayout), Valid: true},
			Konto:     konto,
			Sifra:     sifra,
			Iznos:     iznos,
			Kat:       int16(kat),
			Vrd:       vrd,
			Nsedprim:  row.NazivPartnera,
			Brracup:   row.RacunPartnera,
			Osnplac:   row.OznakaVrstePosla,
			Sdozn:     row.OznakaVrstePosla,
			Duguje:    dug,
			Potrazuje: pot,
			Modelzad:  row.ModelZaduzenja,
			Pnabrzad:  row.PozivNaBrojZaduzenja,
			Mododob:   row.ModelOdobrenja,
			Pnabrodob: row.PozivNaBrojOdobrenja,
			Prekl:     row.ReferencaBanke,
			Xdatunosa: sql.NullTime{Time: time.Now(), Valid: true},
			Xopunos:   sql.NullString{String: userSession.UserName, Valid: true},
		}
		izvDet = append(izvDet, fizvdet)
	}
	izvodHdr.Ukdug += totDug
	izvodHdr.Ukpot += totPot
	izvodHdr.Nstanje = izvodHdr.Prstanje + totDug - totPot
	izvodHdr.Ukbrst = len(izvDet)

	return izvodHdr, izvDet, nil
}

// MapIntesaToDomain maps the Intesa data to domain.Fizvzag and domain.Fizvdet.
func (s *IzvodiResource) MapIntesaToDomain(ctx context.Context, izvodIntesa domain.StmtRsList, prekoJMBG bool) (*domain.Fizvzag, []domain.Fizvdet, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, nil, fmt.Errorf("user session not found")
	}
	// Map the Intesa data to domain.Fizvzag and domain.Fizvdet
	izvodHdr := &domain.Fizvzag{
		God:       userSession.SelectedGod,
		Kar:       userSession.SelectedKar,
		Brrac:     izvodIntesa.StmtRs.AcctID,
		Izvbr:     common.StringToInt(izvodIntesa.StmtRs.StmtNumber),
		Datizv:    sql.NullTime{Time: time.Time(izvodIntesa.StmtRs.LedgerBal.DtAsOf), Valid: true},
		Prstanje:  izvodIntesa.StmtRs.LedgerBal.BalAmt,
		Ukdug:     izvodIntesa.StmtRs.LedgerBal.BalAmt,
		Ukpot:     izvodIntesa.StmtRs.AvailBal.BalAmt,
		Ukbrst:    izvodIntesa.StmtRs.TrnList.Count,
		Nstanje:   izvodIntesa.StmtRs.AvailBal.BalAmt,
		Izvsts:    "1", // status 1 -importovan, status 40 - proknjizen
		Xopunos:   sql.NullString{String: userSession.UserName, Valid: true},
		Xdatunosa: sql.NullTime{Time: time.Now(), Valid: true},
	}
	// Update bank info
	bank, err := s.getBankInfo(ctx, userSession.SelectedGod, userSession.SelectedKar, izvodIntesa.StmtRs.AcctID)
	if err == nil && bank != nil {
		izvodHdr.IDbanke = sql.NullInt64{Int64: int64(bank.IDBanke), Valid: true}
		izvodHdr.Konto = bank.Konto
		izvodHdr.Sifra = bank.Sifra
	}
	ukDug, ukPot := float64(0), float64(0)
	i := int64(0)
	izvDet := []domain.Fizvdet{}
	for _, stavka := range izvodIntesa.StmtRs.TrnList.StmtTrns {
		i++
		konto, sifra := "", ""
		kat := int16(0)
		modZaduzenja, modOdobrenja := "", ""
		dug, pot := float64(0), float64(0)
		if stavka.Benefit == "debit" {
			kat = 1
			dug = stavka.TrnAmt
			modZaduzenja = stavka.Purpose
			ukDug += stavka.TrnAmt
		}
		if stavka.Benefit == "credit" {
			kat = 3
			pot = stavka.TrnAmt
			modOdobrenja = stavka.Purpose
			ukPot += stavka.TrnAmt
		}
		izvodDet := domain.Fizvdet{
			God:       userSession.SelectedGod,
			Kar:       userSession.SelectedKar,
			Brrac:     izvodIntesa.StmtRs.AcctID,
			Izvbr:     common.StringToInt(izvodIntesa.StmtRs.StmtNumber),
			Datizv:    sql.NullTime{Time: time.Time(izvodIntesa.StmtRs.LedgerBal.DtAsOf), Valid: true},
			Rbr:       i,
			Konto:     konto,
			Sifra:     sifra,
			Iznos:     stavka.TrnAmt,
			Kat:       kat,
			Vrd:       90,
			Duguje:    dug,
			Potrazuje: pot,
			Nsedprim:  stavka.PayeeInfo.City,
			Brracup:   stavka.PayeeAccountInfo.AcctID,
			Osnplac:   stavka.PurposeCode,
			Sdozn:     "",
			Sdozn1:    "",
			Modelzad:  modZaduzenja,
			Pnabrzad:  "",
			Mododob:   modOdobrenja,
			Pnabrodob: "",
			Prekl:     "",
		}
		//get konto and sifra from tekracuni table based on brRacDet
		konto, sifra, err := s.getSifplIzvodiInfo(ctx, common.StringToInt(stavka.PurposeCode))
		if err != nil {
			log.Printf("failed to get konto and sifra: %v", err)
		}
		if konto != "" {
			izvodDet.Konto = konto
		}
		if sifra != "" {
			izvodDet.Sifra = sifra
		}
		partnerID := s.getPartneriID(ctx, prekoJMBG, "", stavka.PayeeAccountInfo.AcctID)
		konto, sifra = s.getPartnerKontoSifra(ctx, partnerID, izvodDet.Duguje, izvodDet.Potrazuje)
		if konto != "" {
			izvodDet.Konto = konto
		}
		if sifra != "" {
			izvodDet.Sifra = sifra
		}
		izvodDet.Xopunos = sql.NullString{String: userSession.UserName, Valid: true}
		izvodDet.Xdatunosa = sql.NullTime{Time: time.Now(), Valid: true}
		// add the izvodDet to the list
		izvDet = append(izvDet, izvodDet)
	}
	izvodHdr.Ukdug = ukDug
	izvodHdr.Ukpot = ukPot
	izvodHdr.Nstanje = izvodHdr.Prstanje + ukDug - ukPot
	izvodHdr.Ukbrst = len(izvDet)
	return izvodHdr, izvDet, nil
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
	izvod.BrRacuna = formatAccountNumber(strings.TrimSpace(fields[0]))
	izvod.DatumObrada = strings.TrimSpace(fields[1])
	izvod.BrIzvoda = strings.TrimSpace(fields[2])
	izvod.Valuta = strings.TrimSpace(fields[3])
	izvod.DatumValute = strings.TrimSpace(fields[4])
	izvod.IznosZaduzenja = common.FormatFloatNumber64WithSystemLocale(common.StringToFloat64(strings.TrimSpace(fields[5])), 2)
	izvod.IznosOdobrenja = common.FormatFloatNumber64WithSystemLocale(common.StringToFloat64(strings.TrimSpace(fields[6])), 2)
	izvod.OznakaKnjizenja = strings.TrimSpace(fields[7])
	izvod.Opis = strings.TrimSpace(fields[8])
	izvod.DatumKnjizenja = strings.TrimSpace(fields[9])
	izvod.RacunPartnera = formatAccountNumber(strings.TrimSpace(fields[10]))
	izvod.NazivPartnera = common.TruncateString(strings.TrimSpace(fields[11]), 72)
	izvod.Svrha = common.TruncateString(strings.TrimSpace(fields[12]), 72)
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

// // processHeader processes the header record
// func (s *IzvodiResource) processHeader(ctx context.Context, tx db.Transaction, izvRecord domain.IzvodHalcom, brRac string) (int64, bool, error) {
// 	// Check if header exists
// 	userSession := domain.GetSessionFromStdContext(ctx)
// 	if userSession == nil {
// 		return 0, false, fmt.Errorf("user session not found")
// 	}

// 	hasGod, hasKar := s.izvhdrRepo.GetHasGodHasKar()
// 	qb := common.NewQueryBuilder(`SELECT idfizvzag, brrac, izvbr, datizv, prstanje, ukdug, ukpot, nstanje, ukbrst,
// 	coalesce(nalog, 0) as nalog, coalesce(tipdok, '') as tipdok, izvsts, idbanke FROM fizvzag`, true)
// 	if hasGod {
// 		qb.AddEqual("god", userSession.SelectedGod)
// 	}
// 	if hasKar {
// 		qb.AddEqual("kar", userSession.SelectedKar)
// 	}
// 	qb.AddEqual("brrac", izvRecord.BrRacuna)
// 	qb.AddEqual("izvbr", izvRecord.BrIzvoda)
// 	qb.AddEqual("datizv", izvRecord.DatumObrada)
// 	sqlQuery, args := qb.Build()
// 	existingHeaders, err := s.izvhdrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
// 	if err != nil {
// 		return 0, false, fmt.Errorf("failed to query existing header: %w", err)
// 	}
// 	var existingHeader *domain.Fizvzag
// 	headerFields := []domain.Fields{}
// 	newID := int64(0)
// 	if len(*existingHeaders) > 0 {
// 		existingHeader = &(*existingHeaders)[0]
// 	}
// 	if existingHeader == nil {
// 		// Create new header
// 		headerFields = append(headerFields, []domain.Fields{
// 			{Name: "brrac", Value: brRac},
// 			{Name: "izvbr", Value: strings.TrimSpace(izvRecord.BrIzvoda)},
// 			{Name: "datizv", Value: izvRecord.DatumObrada},
// 			{Name: "konto", Value: ""},
// 			{Name: "sifra", Value: ""},
// 			{Name: "prstanje", Value: "0"},
// 			{Name: "izvsts", Value: "1"},
// 		}...)
// 		datumObrade, err := time.Parse(common.DateLayout, izvRecord.DatumObrada)
// 		if err != nil {
// 			return 0, false, fmt.Errorf("failed to parse date: %w", err)
// 		}
// 		header := &domain.Fizvzag{
// 			God:      userSession.SelectedGod,
// 			Kar:      userSession.SelectedKar,
// 			Brrac:    brRac,
// 			Izvbr:    common.StringToInt(strings.TrimSpace(izvRecord.BrIzvoda)),
// 			Datizv:   sql.NullTime{Time: datumObrade, Valid: true},
// 			Konto:    "",
// 			Sifra:    "",
// 			Prstanje: 0,
// 			Izvsts:   "1",
// 		}
// 		// Update bank info
// 		bank, err := s.getBankInfo(ctx, userSession.SelectedGod, userSession.SelectedKar, brRac)
// 		if err == nil && bank != nil {
// 			header.IDbanke = sql.NullInt64{Int64: int64(bank.IDBanke), Valid: true}
// 			header.Konto = bank.Konto
// 			header.Sifra = bank.Sifra
// 			headerFields = append(headerFields, domain.Fields{Name: "idbanke", Value: fmt.Sprintf("%d", bank.IDBanke)})
// 			headerFields = append(headerFields, domain.Fields{Name: "konto", Value: bank.Konto})
// 			headerFields = append(headerFields, domain.Fields{Name: "sifra", Value: bank.Sifra})
// 		}

// 		// Insert new header
// 		newID, err = s.izvhdrRepo.CreateWithTx(ctx, tx, header, common.IDfizvzag, headerFields)
// 		if err != nil {
// 			return 0, false, fmt.Errorf("failed to insert new header: %w", err)
// 		}
// 		return int64(newID), true, nil
// 	}

// 	// Check if header is in valid state for modification
// 	if common.StringToInt(existingHeader.Izvsts) > 1 {
// 		return 0, false, fmt.Errorf("selected statement is already processed")
// 	}

// 	if err := s.izvhdrRepo.UpdateWithTx(ctx, tx, existingHeader, common.IDfizvzag, newID, headerFields); err != nil {
// 		return 0, false, fmt.Errorf("failed to update header: %w", err)
// 	}

// 	return int64(existingHeader.IDFizvzag), false, nil
// }

// formatAccountNumber formats account number: xxx-xxxxxxxxxxxxx-xx
func formatAccountNumber(account string) string {
	// Remove existing dashes
	clean := strings.ReplaceAll(account, "-", "")

	// Check if we have enough characters
	if len(clean) < 18 {
		return account
	}

	// Format: xxx-xxxxxxxxxxxxx-xx
	// Positions: 0-2 = first 3 digits, 3-15 = next 13 digits, 16-17 = last 2 digits
	return clean[:3] + "-" + clean[3:16] + "-" + clean[16:18]
}
func (s *IzvodiResource) getPartneriID(ctx context.Context, searchJbmg bool, jmbg, brrac string) int {
	// Get partner ID based on JMBG and account number
	if searchJbmg {
		// Search by JMBG
		// The partner of the period of the session (the JMBG is not unique across periods).
		b := querybuilder.New(`SELECT idpartneri FROM partneri`).Where("jmbg = ?", jmbg)
		if err := s.partneriRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
			log.Printf("Error fetching partner by JMBG %s: %v", jmbg, err)
			return 0
		}
		partner, err := s.partneriRepo.Select(ctx, b)
		if err != nil {
			log.Printf("Error fetching partner by JMBG %s: %v", jmbg, err)
			return 0
		}
		if partner != nil && len(*partner) > 0 {
			return (*partner)[0].IDPartneri
		}
	} else {
		// Search by account number
		// The account number is a key: an empty one no longer matches the first tekuci racun.
		b := querybuilder.New(`SELECT IDpartneri FROM tekracuni`).Where("tekrac = ?", brrac)
		if err := s.tekracuniRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
			log.Printf("Error fetching partner by account number %s: %v", brrac, err)
			return 0
		}
		tekracuni, err := s.tekracuniRepo.Select(ctx, b)
		if err != nil {
			log.Printf("Error fetching partner by account number %s: %v", brrac, err)
			return 0
		}
		if tekracuni != nil && len(*tekracuni) > 0 {
			return (*tekracuni)[0].IDPartneri
		}

	}
	return 0
}
func (s *IzvodiResource) getPartnerKontoSifra(ctx context.Context, partnerID int, duguje, potrazuje float64) (string, string) {
	// Get partner konto and sifra
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		log.Printf("User session not found")
		return "", ""
	}
	konto, sifra := "", ""
	// The partner id is a key, restricted to the tenant of the session.
	b := querybuilder.New(`SELECT sifra FROM partneri`).Where("idpartneri = ?", partnerID)
	if err := s.partneriRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		log.Printf("Error fetching partner konto and sifra for partnerID %d: %v", partnerID, err)
		return "", ""
	}
	partnerEntities, err := s.partneriRepo.Select(ctx, b)
	if err != nil {
		log.Printf("Error fetching partner konto and sifra for partnerID %d: %v", partnerID, err)
		return "", ""
	}
	if partnerEntities != nil && len(*partnerEntities) > 0 {
		sifra = (*partnerEntities)[0].Sifra
	}
	bFkpl := querybuilder.New(`SELECT konto FROM fkpl`).Where("vkonta = ?", 2)
	if err := s.fkplRepo.ApplyFilters(ctx, bFkpl, "", nil, ""); err != nil {
		log.Printf("Error fetching fkpl konto for duguje: %v", err)
		return "", ""
	}
	fvrData, err := repository.GetFvrData(ctx, s.fvrRepo)
	if err != nil {
		log.Printf("Error fetching FVR data: %v", err)
		return "", ""
	}
	if duguje > 0 {
		bFkpl.AddLikeBegin("konto", fvrData.KontaDob1)
	}
	if potrazuje > 0 {
		bFkpl.AddLikeBegin("konto", fvrData.KontaKup1)
	}
	fkplEntities, err := s.fkplRepo.Select(ctx, bFkpl)
	if err != nil {
		log.Printf("Error fetching fkpl konto for duguje: %v", err)
		return "", ""
	}
	if fkplEntities != nil && len(*fkplEntities) > 0 {
		konto = (*fkplEntities)[0].Konto
	}

	return konto, sifra
}

func (s *IzvodiResource) getBankInfo(ctx context.Context, god, kar int, brRac string) (*domain.Banke, error) {
	// Get bank info
	// The period comes from the caller (the header being imported); the account number is a key.
	hasGod, hasKar := s.bankeRepo.GetHasGodHasKar()
	bBank := querybuilder.New(`SELECT idbanke, banka, brrac, konto, sifra FROM banke`).
		Where("brrac = ?", brRac)
	if hasGod {
		bBank.Where("god = ?", god)
	}
	if hasKar {
		bBank.Where("kar = ?", kar)
	}
	bankEntities, err := s.bankeRepo.Select(ctx, bBank)
	if err != nil {
		return nil, fmt.Errorf("failed to query bank info: %w", err)
	}
	if len(*bankEntities) > 0 {
		return &(*bankEntities)[0], nil
	}
	return nil, nil
}
func (s *IzvodiResource) getSifplIzvodiInfo(ctx context.Context, sifPlacanja int) (string, string, error) {
	// Get bank info
	b := querybuilder.New(`SELECT sifplac, opis, konto, sifra FROM sifplizv`).
		Where("sifplac = ?", sifPlacanja)
	entities, err := s.sifplizvRepo.Select(ctx, b)
	if err != nil {
		return "", "", fmt.Errorf("failed to query bank info: %w", err)
	}
	if len(*entities) > 0 {
		return (*entities)[0].Konto, (*entities)[0].Sifra, nil
	}
	return "", "", nil
}

func (s *IzvodiResource) AzurirajKonta(ctx context.Context, izvodID string) error {
	// TODO: Implement account update logic
	// - Update accounts in izvodi_detalji based on parameters
	return fmt.Errorf("not implemented")
}

func (s *IzvodiResource) BrisiIzvod(ctx context.Context, fizvzagID string) error {
	// The id comes from the request: a key (Where, so an empty id can not delete every row), and
	// both deletes are restricted to the period of the session.
	bDet := querybuilder.New(`DELETE FROM fizvdet`).Where("idfizvzag = ?", fizvzagID)
	if err := s.izvdetRepo.ApplyFilters(ctx, bDet, "", nil, ""); err != nil {
		return err
	}
	bHdr := querybuilder.New(`DELETE FROM fizvzag`).Where("idfizvzag = ?", fizvzagID)
	if err := s.izvhdrRepo.ApplyFilters(ctx, bHdr, "", nil, ""); err != nil {
		return err
	}
	// - Delete from izvodi_detalji first
	tx, err := s.izvhdrRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("panic recovered: %v", r)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = s.izvdetRepo.ExecWithTx(ctx, tx, bDet)
	if err != nil {
		return fmt.Errorf("failed to delete detail records: %w", err)
	}

	// - Delete from izvodi header
	err = s.izvhdrRepo.ExecWithTx(ctx, tx, bHdr)
	if err != nil {
		return fmt.Errorf("failed to delete header record: %w", err)
	}
	return nil
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
		//{Name: "idfizvdet", Label: "ID", Width: "8", Field: "d.idfizvdet", SkipInSearch: true},
	}
}
func (s *IzvodiResource) GetMasterTableFields() []domain.Fields {
	return s.izvodiHeaderTableFields
}

func (s *IzvodiResource) GetDetailTableFields() []domain.Fields {
	return s.izvodiDetailTableFields
}

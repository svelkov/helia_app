package finansijsko

import (
	"context"
	"database/sql"
	"encoding/xml"
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
	izvodiHeaderTableFields []domain.Fields
	izvodiDetailTableFields []domain.Fields
}

func NewIzvodiResource(izvhdrRepo *repository.BaseRepository[domain.Fizvzag], izvdetRepo *repository.BaseRepository[domain.Fizvdet],
	bankeRepo *repository.BaseRepository[domain.Banke], tipdokRepo *repository.BaseRepository[domain.Tipdok], fnalRepo *repository.BaseRepository[domain.Fnal], partneriRepo *repository.BaseRepository[domain.Partneri], tekracuniRepo *repository.BaseRepository[domain.TekRacuni], sifplizvRepo *repository.BaseRepository[domain.Sifplizv], fkplRepo *repository.BaseRepository[domain.Fkpl], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config) *IzvodiResource {
	rs := &IzvodiResource{
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
			i.nalog, coalesce(i.tipdok,'') as tipdok, i.izvsts, 
			i.idbanke, coalesce(b.banka,'') as banka, coalesce(b.brrac,'') as brrac,
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
	qb := common.NewQueryBuilder(`
		SELECT 
			d.idfizvdet, d.rbr, d.konto, d.sifra, d.kat,
			d.vrd, coalesce(d.konto1, '') as konto1, coalesce(d.sifra1, '') as sifra1, d.iznos, d.duguje,
			d.potrazuje, d.nsedprim, d.brracup, d.osnplac,
			coalesce(d.sdozn, '') as sdozn, coalesce(d.sdozn1, '') as sdozn1, coalesce(d.modelzad, '') as modelzad, coalesce(d.pnabrzad, '') as pnabrzad,
			coalesce(d.mododob, '') as mododob, coalesce(d.pnabrodob, '') as pnabrodob, coalesce(d.prekl, '') as prekl
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
			Key:   fmt.Sprintf("%d", banka.IDBanke),
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
	qb := common.NewQueryBuilder(`SELECT idfizvdet, rbr, konto, sifra, kat, vrd, iznos, duguje, potrazuje FROM fizvdet `, true)
	qb.AddEqual("idfizvzag", fizvhdrID)
	sqlQuery, args := qb.Build()
	entities, err := s.izvdetRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get Fizvdet entities: %w", err)
	}

	qbOsnPlacanja := common.NewQueryBuilder(`SELECT sifplac, konto, sifra FROM sifplizv `, true)
	tekRacuni := common.NewQueryBuilder(`SELECT idpartneri, brrac FROM tekracuni `, true)
	hasGod, hasKar := s.tekracuniRepo.GetHasGodHasKar()
	if hasGod {
		tekRacuni.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		tekRacuni.AddEqual("kar", userSession.SelectedKar)
	}
	qbPartneri := common.NewQueryBuilder(`SELECT idpartneri, god, kar, jmbg FROM partneri `, true)
	hasGod, hasKar = s.partneriRepo.GetHasGodHasKar()
	if hasGod {
		qbPartneri.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qbPartneri.AddEqual("kar", userSession.SelectedKar)
	}
	qbFkpl := common.NewQueryBuilder(`SELECT idfkpl, god, kar, vrd, konto, sifra, idpartneri FROM fkpl `, true)
	hasGod, hasKar = s.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qbFkpl.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qbFkpl.AddEqual("kar", userSession.SelectedKar)
	}
	for _, fizvdet := range *entities {
		// Step 1: Look up SIFPLIZV by OSNPLAC
		qbOsnPlacanja.AddEqual("sifplac", fizvdet.Osnplac)
		qryOsnPlacanja, argsOsnPlacanja := qbOsnPlacanja.Build()
		sifplEntites, err := s.sifplizvRepo.GetAllCustom(ctx, qryOsnPlacanja, "", argsOsnPlacanja, "", "")
		if err == nil && sifplEntites != nil && len(*sifplEntites) > 0 {
			sifplizv := (*sifplEntites)[0]
			// Update KONTO and SIFRA based on checkbox condition
			if azurirajKonta {
				fizvdet.Konto = sifplizv.Konto
				fizvdet.Sifra = sifplizv.Sifra
			} else {
				if fizvdet.Konto == "" {
					fizvdet.Konto = sifplizv.Konto
				}
				if fizvdet.Sifra == "" {
					fizvdet.Sifra = sifplizv.Sifra
				}
			}
		}

		// Step 2: Determine partner ID based on JMBG or account number
		var nIDPARTNERI int64 = 0

		if povezivanjeJMBG {
			// Extract JMBG from PNABRODOB field
			pnabrodob := strings.TrimSpace(fizvdet.Pnabrodob)
			var sJMBG string

			if len(pnabrodob) <= 13 {
				sJMBG = pnabrodob
			} else {
				// Split by "-" and take first part as JMBG
				parts := strings.Split(pnabrodob, "-")
				if len(parts) >= 1 {
					sJMBG = strings.TrimSpace(parts[0])
				}
				// Note: parts[1] would be sCode (payment code 01 or 02)
			}

			// Look up partner by JMBG
			if sJMBG != "" {
				qbPartneri.AddEqual("jmbg", sJMBG)
				qryPartneri, argsPartneri := qbPartneri.Build()
				partneri, err := s.partneriRepo.GetAllCustom(ctx, qryPartneri, "", argsPartneri, "", "")
				if err == nil && partneri != nil && len(*partneri) > 0 {
					nIDPARTNERI = int64((*partneri)[0].IDPartneri)
				}
			}
		} else {
			// Look up by account number from TEKRACUNI
			if fizvdet.Brracup != "" {
				tekRacuni.AddEqual("brrac", fizvdet.Brracup)
				qryTekRacuni, argsTekRacuni := tekRacuni.Build()
				tekracuni, err := s.tekracuniRepo.GetAllCustom(ctx, qryTekRacuni, "", argsTekRacuni, "", "")
				if err == nil && tekracuni != nil && len(*tekracuni) > 0 {
					nIDPARTNERI = int64((*tekracuni)[0].IDPartneri)
				}
			}
		}

		// Step 3: Update FKPL based on DUGUJE (debit) amount
		if nIDPARTNERI > 0 && fizvdet.Duguje != 0 {
			qbFkpl.AddEqual("konto", s.cfg.Konta.KontoDobavljaca)
			fkplQuery, fkplArgs := qbFkpl.Build()
			fkplEntities, err := s.fkplRepo.GetAllCustom(ctx, fkplQuery, "", fkplArgs, "", "")
			if err == nil && fkplEntities != nil && len(*fkplEntities) > 0 {
				fkpl := (*fkplEntities)[0]
				if azurirajKonta {
					fizvdet.Konto = fkpl.Konto
					fizvdet.Sifra = fkpl.Sifra
				} else {
					if fizvdet.Konto == "" {
						fizvdet.Konto = fkpl.Konto
					}
					if fizvdet.Sifra == "" {
						fizvdet.Sifra = fkpl.Sifra
					}
				}
				fizvdet.Vrd = 40
				fizvdet.Kat = 1
			}
		}

		// Step 4: Update FKPL based on POTRAZUJE (credit) amount
		if nIDPARTNERI > 0 && fizvdet.Potrazuje != 0 {
			qbFkpl.AddEqual("konto", s.cfg.Konta.KontoKupca)
			fkplQuery, fkplArgs := qbFkpl.Build()
			fkplEntities, err := s.fkplRepo.GetAllCustom(ctx, fkplQuery, "", fkplArgs, "", "")
			if err == nil && fkplEntities != nil && len(*fkplEntities) > 0 {
				fkpl := (*fkplEntities)[0]
				if azurirajKonta {
					fizvdet.Konto = fkpl.Konto
					fizvdet.Sifra = fkpl.Sifra
				} else {
					if fizvdet.Konto == "" {
						fizvdet.Konto = fkpl.Konto
					}
					if fizvdet.Sifra == "" {
						fizvdet.Sifra = fkpl.Sifra
					}
				}
				fizvdet.Vrd = 30
				fizvdet.Kat = 3
			}
		}
		// At the end of loop update fizvdet record in the database
		qbUpdate := common.NewQueryBuilder(`UPDATE fizvdet SET konto = $1, sifra = $2, kat = $3, vrd = $4 `, true)
		qbUpdate.AddArgs(fizvdet.Konto, fizvdet.Sifra, fizvdet.Kat, fizvdet.Vrd)
		qbUpdate.AddEqual("idfizvdet", fizvdet.IDFizvdet)
		sqlUpdate, argsUpdate := qbUpdate.Build()
		tx, err := s.izvdetRepo.BeginTx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		_, err = tx.ExecContext(ctx, sqlUpdate, argsUpdate...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update Fizvdet: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// UpisIzvoda inserts the header and detail records into the database.
func (s *IzvodiResource) UpisIzvoda(ctx context.Context, izvHdr *domain.Fizvzag, izvDet []domain.Fizvdet) error {
	//check ig the izvod exist and is already booked, if so return error
	hasGod, hasKar := s.izvhdrRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idfizvzag, izvsts FROM fizvzag `, true)
	if hasGod {
		qb.AddEqual("god", izvHdr.God)
	}
	if hasKar {
		qb.AddEqual("kar", izvHdr.Kar)
	}
	qb.AddEqual("brrac", izvHdr.Brrac)
	qb.AddEqual("izvbr", izvHdr.Izvbr)
	sqlQuery, args := qb.Build()
	entities, err := s.izvhdrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
		qbUpdate := common.NewQueryBuilder(`UPDATE fizvzag SET 
		datizv = $1, konto = $2, sifra = $3, prstanje = $4, ukdug = $5, 
		ukpot = $6, nstanje = $7, ukbrst = $8, nalog = $9, tipdok = $10, izvsts = $11,
		 idbanke = $12, xdatizmene = NOW(), xopizmene = $13 WHERE idfizvzag = $14`, false)
		sqlUpdate, argsUpdate := qbUpdate.Build()
		argsUpdate = append(argsUpdate, izvHdr.Datizv, izvHdr.Konto, izvHdr.Sifra,
			izvHdr.Prstanje, izvHdr.Ukdug, izvHdr.Ukpot, izvHdr.Nstanje,
			izvHdr.Ukbrst, izvHdr.Nalog, izvHdr.Tipdok, izvHdr.Izvsts, izvHdr.IDbanke, izvHdr.Xopunos, IdFizvzag)
		_, err = tx.ExecContext(ctx, sqlUpdate, argsUpdate...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update Fizvzag: %w", err)
		}
	} else {
		qbInsert := common.NewQueryBuilder(`INSERT INTO fizvzag (god, kar, brrac, izvbr, datizv, konto, sifra, 
		prstanje, ukdug, ukpot, nstanje, ukbrst, nalog, tipdok, izvsts, idbanke, xdatunosa, xopunos) 
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING idfizvzag`, false)
		sqlInsert, argsInsert := qbInsert.Build()
		argsInsert = append(argsInsert, izvHdr.God, izvHdr.Kar, izvHdr.Brrac, izvHdr.Izvbr,
			izvHdr.Datizv, izvHdr.Konto, izvHdr.Sifra, izvHdr.Prstanje, izvHdr.Ukdug,
			izvHdr.Ukpot, izvHdr.Nstanje, izvHdr.Ukbrst, izvHdr.Nalog,
			izvHdr.Tipdok, izvHdr.Izvsts, izvHdr.IDbanke, izvHdr.Xdatunosa, izvHdr.Xopunos)
		err := tx.QueryRowContext(ctx, sqlInsert, argsInsert...).Scan(&IdFizvzag)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert Fizvzag: %w", err)
		}
	}
	// Delete existing detail records for the header before inserting new ones
	sqlDelete := `DELETE FROM fizvdet WHERE idfizvzag = $1`
	_, err = tx.ExecContext(ctx, sqlDelete, IdFizvzag)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing detail records: %w", err)
	}
	sqlQuery, args = s.createBulkInsertDetailQuery(ctx, izvDet, IdFizvzag)
	_, err = tx.ExecContext(ctx, sqlQuery, args...)
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

// getIzvodHeader retrieves the ID of the Izvod header based on the provided parameters.
func (s *IzvodiResource) getIzvodHeader(ctx context.Context, god, kar int, brRacuna string, brIzvoda string, datumObrade time.Time) (int64, error) {
	qb := common.NewQueryBuilder(`SELECT fizvzagid FROM fizvzag `, true)
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
		return 0, fmt.Errorf("failed to query izvod header: %w", err)
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

func (s *IzvodiResource) createBulkInsertDetailQuery(ctx context.Context, izvDet []domain.Fizvdet, idFizvzag int64) (string, []interface{}) {
	if len(izvDet) == 0 {
		return "", nil
	}

	// Column list - excluding auto-generated idfizvdet
	columns := []string{
		"god", "kar", "brrac", "izvbr", "datizv", "rbr",
		"konto", "sifra", "iznos", "kat", "vrd", "nsedprim",
		"brracup", "osnplac", "sdozn", "duguje", "potrazuje",
		"modelzad", "pnabrzad", "mododob", "pnabrodob", "prekl",
		"xdatunosa", "xopunos", "idfizvzag",
	}

	// Build query with your exact column list
	sqlQuery := "INSERT INTO fizvdet (" + strings.Join(columns, ", ") + ") VALUES "

	var args []interface{}
	paramCount := 1
	i := 0
	for _, row := range izvDet {
		if i > 0 {
			sqlQuery += ", "
		}

		placeholders := make([]string, len(columns))
		for j := range columns {
			placeholders[j] = fmt.Sprintf("$%d", paramCount)
			paramCount++
		}
		sqlQuery += "(" + strings.Join(placeholders, ", ") + ")"
		i++

		args = append(args, row.God, row.Kar, row.Brrac, row.Izvbr, row.Datizv, row.Rbr, row.Konto, row.Sifra,
			row.Iznos, row.Kat, row.Vrd, row.Nsedprim, row.Brracup,
			row.Osnplac, row.Sdozn, row.Duguje, row.Potrazuje, row.Modelzad, row.Pnabrzad,
			row.Mododob, row.Pnabrodob, row.Prekl, row.Xdatunosa, row.Xopunos, idFizvzag)
	}

	sqlQuery = strings.TrimSuffix(sqlQuery, ", \n") // Remove the trailing comma and newline
	return sqlQuery, args
}

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
		partner, err := s.partneriRepo.GetAllCustom(ctx, "SELECT IDpartneri FROM partneri WHERE jmbg = $1", "", []interface{}{jmbg}, "", "")
		if err != nil {
			log.Printf("Error fetching partner by JMBG %s: %v", jmbg, err)
			return 0
		}
		if partner != nil && len(*partner) > 0 {
			return (*partner)[0].IDPartneri
		}
	} else {
		// Search by account number
		qb := common.NewQueryBuilder(`SELECT IDpartneri FROM tekracuni`, true)
		hasGod, hasKar := s.tekracuniRepo.GetHasGodHasKar()
		if hasGod {
			qb.AddEqual("god", domain.GetSessionFromStdContext(ctx).SelectedGod)
		}
		if hasKar {
			qb.AddEqual("kar", domain.GetSessionFromStdContext(ctx).SelectedKar)
		}
		qb.AddEqual("tekrac", brrac)
		sqlQuery, args := qb.Build()
		tekracuni, err := s.tekracuniRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	qb := common.NewQueryBuilder(`SELECT sifra FROM partneri `, true)
	qb.AddEqual("idpartneri", partnerID)
	sqlQuery, args := qb.Build()
	partnerEntities, err := s.partneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		log.Printf("Error fetching partner konto and sifra for partnerID %d: %v", partnerID, err)
		return "", ""
	}
	if partnerEntities != nil && len(*partnerEntities) > 0 {
		sifra = (*partnerEntities)[0].Sifra
	}
	qb = common.NewQueryBuilder(`SELECT konto FROM fkpl `, true)
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod) // Replace with actual god value if needed
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar) // Replace with actual kar value if needed
	}
	qb.AddEqual("vkonta", 2)
	fvrData, err := common.GetFvrData(ctx, s.fvrRepo)
	if err != nil {
		log.Printf("Error fetching FVR data: %v", err)
		return "", ""
	}
	if duguje > 0 {
		qb.AddLikeBegin("konto", fvrData.KontaDob1)
	}
	if potrazuje > 0 {
		qb.AddLikeBegin("konto", fvrData.KontaKup1)

	}
	sqlQuery, args = qb.Build()
	fkplEntities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	qbBank := common.NewQueryBuilder(`SELECT idbanke, banka, brrac, konto, sifra FROM banke`, true)
	hasGod, hasKar := s.bankeRepo.GetHasGodHasKar()
	if hasGod {
		qbBank.AddEqual("god", god)
	}
	if hasKar {
		qbBank.AddEqual("kar", kar)
	}
	qbBank.AddEqual("brrac", brRac)
	sqlQueryBank, argsBank := qbBank.Build()
	bankEntities, err := s.bankeRepo.GetAllCustom(ctx, sqlQueryBank, "", argsBank, "", "")
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
	qb := common.NewQueryBuilder(`SELECT sifplac, opis, konto, sifra FROM sifplizv`, true)

	qb.AddEqual("sifplac", sifPlacanja)
	sqlQuery, args := qb.Build()
	entities, err := s.sifplizvRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	// - Delete from izvodi_detalji first
	qb := common.NewQueryBuilder(`DELETE FROM fizvdet `, true)
	qb.AddEqual("idfizvzag", fizvzagID)
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
	sqlQuery, args := qb.Build()
	_, err = tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete detail records: %w", err)
	}

	// - Delete from izvodi header
	qb = common.NewQueryBuilder(`DELETE FROM fizvzag `, true)
	qb.AddEqual("idfizvzag", fizvzagID)
	sqlQuery, args = qb.Build()
	_, err = tx.ExecContext(ctx, sqlQuery, args...)
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

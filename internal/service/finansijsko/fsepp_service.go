package finansijsko

import (
	"context"
	"encoding/csv"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"helia/internal/service"
	"log"
	"reflect"
	"strings"
)

// FseppService defines the interface for operations related to FSEPP (Fiksna Evidencija Prethodnog Poreza).
type FseppService interface {
	GetSekcijeIzvori(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetEvidencija(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odDatuma, doDatuma, searchText string) error
	GetSefKpr(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odDatuma, doDatuma, searchText string) error
	FseppSefKprImport(ctx context.Context, tbl *domain.TableData, fileData string, filterType, odDatuma, doDatuma string, getTotalRecords bool, pageSize, currentPage int) error
	GetSekcijeIzvoriTableFields() []domain.Fields
	GetEvidencijaTableFields() []domain.Fields
	GetSefKprTableFields() []domain.Fields
	GetFieldCache() map[string]reflect.StructField
}

// FseppResource implements the FseppService interface.
type FseppResource struct {
	fseppService             *service.BaseService[domain.Fsepp]
	fseppSefKprService       *service.BaseService[domain.FseppSefKpr]
	fseppRepo                *repository.BaseRepository[domain.Fsepp]
	fseppSefKprRepo          *repository.BaseRepository[domain.FseppSefKpr]
	kprRepo                  *repository.BaseRepository[domain.Kpr]
	sekcijeIzvoriTableFields []domain.Fields
	evidencijaTableFields    []domain.Fields
	sefKprTableFields        []domain.Fields
}

func NewFseppService(
	fseppService *service.BaseService[domain.Fsepp],
	fseppSefKprService *service.BaseService[domain.FseppSefKpr],
	fseppRepo *repository.BaseRepository[domain.Fsepp],
	fseppSefKprRepo *repository.BaseRepository[domain.FseppSefKpr],
	kprRepo *repository.BaseRepository[domain.Kpr],
) *FseppResource {
	rs := &FseppResource{
		fseppService:       fseppService,
		fseppSefKprService: fseppSefKprService,
		fseppRepo:          fseppRepo,
		fseppSefKprRepo:    fseppSefKprRepo,
		kprRepo:            kprRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetSekcijeIzvoriTableFields returns the table field definitions for Sekcije i izvori
func (s *FseppResource) GetSekcijeIzvoriTableFields() []domain.Fields {
	return s.sekcijeIzvoriTableFields
}

// GetEvidencijaTableFields returns the table field definitions for Evidencija PP
func (s *FseppResource) GetEvidencijaTableFields() []domain.Fields {
	return s.evidencijaTableFields
}

// GetSefKprTableFields returns the table field definitions for SEF-KPR
func (s *FseppResource) GetSefKprTableFields() []domain.Fields {
	return s.sefKprTableFields
}

// GetSekcije retrieves data for EPP Sekcije i izvori
func (s *FseppResource) GetSekcijeIzvori(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Build query for Sekcije i izvori
	b := querybuilder.New(`
    SELECT
        -- Row 1: Basic info
        fsepp.fseppid, fsepp.god, fsepp.kar, fsepp.tip, fsepp.deo, 
        fsepp.sekcija, fsepp.izvor, fsepp.polje, fsepp.opis,
        -- Row 2: Tax/Amount fields
        fsepp.osn1, fsepp.pdv1, fsepp.osn2, fsepp.pdv2, fsepp.nipo,
        -- Row 3: Position array fields
        fsepp.pozic_1, fsepp.pozic_2, 
        fsepp.pozic_3, fsepp.pozic_4, 
        fsepp.pozic_5,
        -- Row 4: Actual amounts
        fsepp.aktosn1, fsepp.aktpdv1, fsepp.aktosn2, fsepp.aktpdv2,
        -- Row 5: KPR fields
        fsepp.kpr_polje1o, fsepp.kpr_polje1p, 
        fsepp.kpr_polje2o, fsepp.kpr_polje2p, 
        -- Row 6: Document and dates
        fsepp.grpdok, fsepp.autosef, fsepp.oddat, fsepp.dodat
    FROM fsepp`)
	// The period of the session (fsepp.god/kar) and the search over the columns of the grid.
	if err := s.fseppRepo.ApplyFilters(ctx, b, "fsepp", s.GetSekcijeIzvoriTableFields(), searchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fseppRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fsepp.nipo ASC, fsepp.sekcija ASC").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	// Execute query and populate table
	entities, err := s.fseppRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%d", entity.Deo),
				fmt.Sprintf("%d", entity.Sekcija),
				fmt.Sprintf("%d", entity.Izvor),
				entity.Polje,
				entity.Opis,
				fmt.Sprintf("%d", entity.Nipo),
				common.FormatNumberWithSystemLocale(entity.AktOsn1, 2),
				common.FormatNumberWithSystemLocale(entity.AktPdv1, 2),
				common.FormatNumberWithSystemLocale(entity.AktOsn2, 2),
				common.FormatNumberWithSystemLocale(entity.AktPdv2, 2),
				common.FormatNumberWithSystemLocale(entity.KprPolje1O, 2),
				common.FormatNumberWithSystemLocale(entity.KprPolje1P, 2),
				common.FormatNumberWithSystemLocale(entity.KprPolje2O, 2),
				common.FormatNumberWithSystemLocale(entity.KprPolje2P, 2),
				entity.Pozic1,
				entity.Pozic2,
				entity.Pozic3,
				entity.Pozic4,
				entity.Pozic5,
				entity.OdDat.Time.Format(common.DateLayout),
				entity.DoDat.Time.Format(common.DateLayout),
				entity.GrpDok,
				fmt.Sprintf("%d", entity.Tip),
				fmt.Sprintf("%v", entity.AutoSef),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.FseppID), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}

	}
	return nil
}

// GetEvidencija retrieves data for FSEPP Evidencija PP
func (s *FseppResource) GetEvidencija(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odDatuma, doDatuma, searchText string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Evidencija PP", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Step 2: Get MAX of nipo
	bMaxNipo := querybuilder.New(`SELECT MAX(nipo) as nipo FROM fsepp`)
	if err := s.fseppRepo.ApplyFilters(ctx, bMaxNipo, "", nil, ""); err != nil {
		return err
	}
	var maxNipo int
	fseppRecords, err := s.fseppRepo.Select(ctx, bMaxNipo)
	if err != nil {
		return fmt.Errorf("error fetching max nipo: %w", err)
	}
	if len(*fseppRecords) > 0 {
		maxNipo = (*fseppRecords)[0].Nipo
	}

	// Step 3: Update FSEPP records based on KPR data
	bKpr := querybuilder.New(`SELECT kpr.fseppid, kpr.god, kpr.kar, kpr.dknjiz,
									kpr.nisuobvpdv, kpr.uvozbezpdv, kpr.prethpdv, kpr.pretpdv1, kpr.pretpdv2,
					kpr.uvozpdv, kpr.poljvred, kpr.poljpdv, kpr.uvozosnpdv, kpr.osnbezpdv, kpr.osnovicavt, kpr.osnovicant, kpr.prethpdvvt, kpr.prethpdvnt,
					kpr.pretpdv1vt, kpr.pretpdv1nt, kpr.pretpdv2vt, kpr.pretpdv2nt, kpr.epp_polje FROM kpr`).
		AddCondition("kpr.dknjiz", odDatuma, ">=").
		AddCondition("kpr.dknjiz", doDatuma, "<=").
		Where("kpr.fseppid > ?", 0)
	if err := s.kprRepo.ApplyFilters(ctx, bKpr, "kpr", nil, ""); err != nil {
		return err
	}
	kprRecords, err := s.kprRepo.Select(ctx, bKpr)
	if err != nil {
		return fmt.Errorf("error fetching kpr records: %w", err)
	}

	// Step 1: Reset all FSEPP records for given god and kar
	bReset := querybuilder.NewUpdate("fsepp").
		Set("osn1 = 0, pdv1 = 0, osn2 = 0, pdv2 = 0, oddat = ?, dodat = ?", odDatuma, doDatuma)
	if err := s.fseppRepo.ApplyFilters(ctx, bReset, "", nil, ""); err != nil {
		return err
	}
	// Start a transaction
	tx, err := s.fseppRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	// Roll back on every early return (also the ones that return a shadowed err); after the
	// Commit below the rollback does nothing.
	defer func() { _ = tx.Rollback() }()
	err = s.fseppRepo.ExecWithTx(ctx, tx, bReset)
	if err != nil {
		return fmt.Errorf("error updating fsepp records: %w", err)
	}

	for _, kpr := range *kprRecords {
		sPoljeDA := ""
		sPoljeNE := ""
		fsepp, err := s.fseppRepo.GetByID(ctx, common.IDfsepp, kpr.FseppID)
		if err != nil {
			log.Printf("error fetching fsepp record: %v", err)
			continue //should continue to next record if error occurs
		}
		if fsepp != nil && fsepp.Nipo == 1 {
			if fsepp.AktOsn1 && fsepp.KprPolje1O != "" {
				fsepp.Osn1 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje1O))
			}
			if fsepp.AktPdv1 && fsepp.KprPolje1P != "" {
				fsepp.Pdv1 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje1P))
			}
			if fsepp.AktOsn2 && fsepp.KprPolje2O != "" {
				fsepp.Osn2 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje2O))
			}
			if fsepp.AktPdv2 && fsepp.KprPolje2P != "" {
				fsepp.Pdv2 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje2P))
			}
			if err := s.updateFseppIznosi(ctx, tx, fsepp); err != nil {
				return fmt.Errorf("error updating fsepp record %d: %w", fsepp.FseppID, err)
			}
			sPoljeDA = fsepp.PoljePdvDa
			sPoljeNE = fsepp.PoljePdvNe
			if sPoljeDA != "" {
				// fseppid is read so that updateFseppIznosi changes this row (it matched fseppid 0 before).
				bFsepp := querybuilder.New(`SELECT
					fsepp.fseppid, fsepp.polje, fsepp.opis, fsepp.osn1, fsepp.pdv1, fsepp.osn2, fsepp.pdv2, fsepp.nipo,
					fsepp.oddat, fsepp.dodat, fsepp.kpr_polje1o, fsepp.kpr_polje1p, fsepp.kpr_polje2o, fsepp.kpr_polje2p FROM fsepp`).
					Where("fsepp.polje_pdv_da = ?", sPoljeDA)
				if err := s.fseppRepo.ApplyFilters(ctx, bFsepp, "fsepp", nil, ""); err != nil {
					return err
				}
				entites, err := s.fseppRepo.Select(ctx, bFsepp)
				if err != nil {
					log.Printf("error fetching fsepp record for polje_pdv_da: %v", err)
					continue
				}
				if entites != nil && len(*entites) > 0 {
					fsepp := &(*entites)[0]
					if fsepp.AktOsn1 && fsepp.KprPolje1O != "" {
						fsepp.Osn1 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje1O))
					}
					if fsepp.AktPdv1 && fsepp.KprPolje1P != "" {
						fsepp.Pdv1 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje1P))
					}
					if fsepp.AktOsn2 && fsepp.KprPolje2O != "" {
						fsepp.Osn2 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje2O))
					}
					if fsepp.AktPdv2 && fsepp.KprPolje2P != "" {
						fsepp.Pdv2 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje2P))
					}
					if err := s.updateFseppIznosi(ctx, tx, fsepp); err != nil {
						return fmt.Errorf("error updating fsepp record %d: %w", fsepp.FseppID, err)
					}
				}
			}
			if sPoljeNE != "" {
				bFsepp := querybuilder.New(`SELECT
					fsepp.fseppid, fsepp.polje, fsepp.opis, fsepp.osn1, fsepp.pdv1, fsepp.osn2, fsepp.pdv2, fsepp.nipo,
					fsepp.oddat, fsepp.dodat, fsepp.kpr_polje1o, fsepp.kpr_polje1p, fsepp.kpr_polje2o, fsepp.kpr_polje2p FROM fsepp`).
					Where("fsepp.polje_pdv_ne = ?", sPoljeNE)
				if err := s.fseppRepo.ApplyFilters(ctx, bFsepp, "fsepp", nil, ""); err != nil {
					return err
				}
				entites, err := s.fseppRepo.Select(ctx, bFsepp)
				if err != nil {
					log.Printf("error fetching fsepp record for polje_pdv_ne: %v", err)
					continue
				}
				if entites != nil && len(*entites) > 0 {
					fsepp := &(*entites)[0]
					if fsepp.AktOsn1 && fsepp.KprPolje1O != "" {
						fsepp.Osn1 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje1O))
					}
					if fsepp.AktPdv1 && fsepp.KprPolje1P != "" {
						fsepp.Pdv1 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje1P))
					}
					if fsepp.AktOsn2 && fsepp.KprPolje2O != "" {
						fsepp.Osn2 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje2O))
					}
					if fsepp.AktPdv2 && fsepp.KprPolje2P != "" {
						fsepp.Pdv2 += common.AnyToFloat64(GetFieldValue(kpr, fsepp.KprPolje2P))
					}
					if err := s.updateFseppIznosi(ctx, tx, fsepp); err != nil {
						return fmt.Errorf("error updating fsepp record %d: %w", fsepp.FseppID, err)
					}
				}
			}
		}
	}

	// Step 4: do the calculation for the positions
	for k := 2; k <= maxNipo; k++ {
		bFsepp := querybuilder.New(`SELECT pozic_0, pozic_1, pozic_2, pozic_3, pozic_4, pozic_5 FROM fsepp`).
			Where("fsepp.nipo = ?", k)
		if err := s.fseppRepo.ApplyFilters(ctx, bFsepp, "fsepp", nil, ""); err != nil {
			return err
		}
		entities, err := s.fseppRepo.Select(ctx, bFsepp)
		if err != nil {
			log.Printf("error fetching fsepp records for nipo %d: %v", k, err)
			continue
		}
		if entities != nil && len(*entities) > 0 {
			fsepp := &(*entities)[0]
			for _, pozicija := range []string{fsepp.Pozic1, fsepp.Pozic2, fsepp.Pozic3, fsepp.Pozic4, fsepp.Pozic5} {
				if err := s.ObradaPozicije(ctx, fsepp, pozicija); err != nil {
					return fmt.Errorf("pozicija %q of fsepp %d: %w", pozicija, fsepp.FseppID, err)
				}
			}
			err = s.UpdateFsepp(ctx, tx, fsepp) // Update the FSEPP record after processing positions
			if err != nil {
				log.Printf("error updating fsepp record: %v", err)
				return fmt.Errorf("error updating fsepp record: %w", err)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("error committing transaction: %v", err)
		return fmt.Errorf("error committing transaction: %w", err)
	}
	// Step 3: get updated data for Evidencija PP
	// Build query for Evidencija PP
	b := querybuilder.New(`
    SELECT
         fsepp.polje, fsepp.opis, fsepp.osn1, fsepp.pdv1, fsepp.osn2, fsepp.pdv2, fsepp.nipo,
         fsepp.oddat, fsepp.dodat
    FROM fsepp`).
		AddCondition("fsepp.oddat", odDatuma, ">=").
		AddCondition("fsepp.dodat", doDatuma, "<=")
	// The period of the session (fsepp.god/kar) and the search over the columns of the grid.
	if err := s.fseppRepo.ApplyFilters(ctx, b, "fsepp", s.GetEvidencijaTableFields(), searchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fseppRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fsepp.nipo ASC, fsepp.sekcija ASC").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	// Execute query and populate table
	entities, err := s.fseppRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	tbl.Headers = s.GetEvidencijaTableFields()
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Polje,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Osn1, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv1, 2),
				common.FormatNumberWithSystemLocale(entity.Osn2, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv2, 2),
				entity.OdDat.Time.Format(common.DateLayout),
				entity.DoDat.Time.Format(common.DateLayout),
				fmt.Sprintf("%d", entity.Nipo),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.FseppID), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

func (s *FseppResource) UpdateFsepp(ctx context.Context, tx db.Transaction, fsepp *domain.Fsepp) error {
	err := s.updateFseppIznosi(ctx, tx, fsepp)
	if err != nil {
		log.Printf("error updating fsepp record: %v", err)
	}
	return err
}

// updateFseppIznosi writes the osnovice and the PDV of a fsepp row (by its fseppid, which comes from
// a row of the period read before).
func (s *FseppResource) updateFseppIznosi(ctx context.Context, tx db.Transaction, fsepp *domain.Fsepp) error {
	b := querybuilder.NewUpdate("fsepp").
		Set("osn1 = ?, pdv1 = ?, osn2 = ?, pdv2 = ?", fsepp.Osn1, fsepp.Pdv1, fsepp.Osn2, fsepp.Pdv2).
		Where("fseppid = ?", fsepp.FseppID)
	return s.fseppRepo.ExecWithTx(ctx, tx, b)
}

func (s *FseppResource) ObradaPozicije(ctx context.Context, fsepp *domain.Fsepp, pozicija string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	bMinus, sPozic := false, ""
	if getMiddle(pozicija, 1, 1) == "+" || getMiddle(pozicija, 1, 1) == "-" {
		if getMiddle(pozicija, 1, 1) == "-" {
			bMinus = true
		} else {
			bMinus = false
		}
		sPozic = getMiddle(pozicija, 2, len(pozicija)-1)
		//sOppozic = getMiddle(pozicija, 1, 3)
	}
	if getMiddle(pozicija, 4, 1) == "+" || getMiddle(pozicija, 4, 1) == "-" {
		if getMiddle(pozicija, 4, 1) == "-" {
			bMinus = true
		} else {
			bMinus = false
		}
		sPozic = getMiddle(pozicija, 5, len(pozicija)-4)
		//sOppozic = getMiddle(pozicija, 1, 3)
	}
	sPozic = ""
	//sOppozic = ""
	bMinus = false
	if sPozic != "" {
		bPozic := querybuilder.New(`SELECT * FROM fsepp`).Where("fsepp.polje = ?", sPozic)
		if err := s.fseppRepo.ApplyFilters(ctx, bPozic, "fsepp", nil, ""); err != nil {
			return err
		}
		entitiesPozic, err := s.fseppRepo.Select(ctx, bPozic)
		if err != nil {
			log.Printf("error fetching fsepp records for polje %s: %v", sPozic, err)
			return err
		}
		if entitiesPozic != nil && len(*entitiesPozic) > 0 {
			fseppPozic := &(*entitiesPozic)[0]
			if bMinus {
				if fsepp.AktOsn1 == true {
					fsepp.Osn1 -= fseppPozic.Osn1
				}
				if fsepp.AktPdv1 == true {
					fsepp.Pdv1 -= fseppPozic.Pdv1
				}
				if fsepp.AktOsn2 == true {
					fsepp.Osn2 -= fseppPozic.Osn2
				}
				if fsepp.AktPdv2 == true {
					fsepp.Pdv2 -= fseppPozic.Pdv2
				}
			} else {
				if fsepp.AktOsn1 == true {
					fsepp.Osn1 += fseppPozic.Osn1
				}
				if fsepp.AktPdv1 == true {
					fsepp.Pdv1 += fseppPozic.Pdv1
				}
				if fsepp.AktOsn2 == true {
					fsepp.Osn2 += fseppPozic.Osn2
				}
				if fsepp.AktPdv2 == true {
					fsepp.Pdv2 += fseppPozic.Pdv2
				}
			}
		}
	}
	return nil
}

func processPozic(pozic string) (bMinus bool, sPozic, sOppozic string) {

	if len(pozic) < 4 {
		// Skip if string is too short
		return false, "", ""
	}

	// Get 4th character (1-indexed) = index 3 (0-indexed)
	signChar := pozic[3:4]

	if signChar == "+" || signChar == "-" {
		// Determine if minus
		bMinus = signChar == "-"

		// Get from position 5 to end (1-indexed) = index 4 to end (0-indexed)
		sPozic = pozic[4:]

		// Get first 3 characters (1-indexed positions 1-3) = indices 0-2
		sOppozic = pozic[:3]

	}
	return bMinus, sPozic, sOppozic
}

// Helper function that mimics the behavior of the Middle function from other languages
func getMiddle(s string, start, length int) string {
	// Convert from 1-indexed to 0-indexed
	startIdx := start - 1

	// Check boundaries
	if startIdx < 0 || startIdx >= len(s) {
		return ""
	}

	endIdx := startIdx + length
	if endIdx > len(s) {
		endIdx = len(s)
	}

	return s[startIdx:endIdx]
}

// GetSefKpr retrieves data for FSEPP SEF-KPR
func (s *FseppResource) GetSefKpr(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odDatuma, doDatuma, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "SEF-KPR", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Build query for SEF-KPR
	b := querybuilder.New(`
		SELECT
			feppsef.redbroj, feppsef.dokumenttip, feppsef.brdokumenta,
			feppsef.datumdokumenta, feppsef.datumlicnog, feppsef.iznos,
			feppsef.pdv, feppsef.konto, feppsef.status,
			feppsef.idfeppsef
		FROM feppsef`).
		AddCondition("feppsef.datumdokumenta", odDatuma, ">=").
		AddCondition("feppsef.datumdokumenta", doDatuma, "<=")
	// The period of the session (feppsef.god/kar) and the search over the columns of the grid.
	if err := s.fseppSefKprRepo.ApplyFilters(ctx, b, "feppsef", s.GetSefKprTableFields(), searchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fseppSefKprRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("feppsef.redbroj ASC").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	// Execute query and populate table
	entities, err := s.fseppSefKprRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%v", entity.RedBroj),
				entity.DokumentTip,
				entity.BrDokumenta,
				entity.DatumDokumenta.Format(common.DateLayout),
				entity.DatumLicnog.Format(common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				common.FormatNumberWithSystemLocale(entity.PDV, 2),
				entity.Konto,
				entity.Status,
				fmt.Sprintf("%v", entity.IDFeppSef),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFeppSef), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetSefKprTableFields()

	return nil
}

func (s *FseppResource) FseppSefKprImport(ctx context.Context, tbl *domain.TableData, fileData string, filterType, odDatuma, doDatuma string, getTotalRecords bool, pageSize, currentPage int) error {
	// Setup pagination
	common.SetupTablePagination(tbl, currentPage, pageSize)

	reader := csv.NewReader(strings.NewReader(fileData))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Create the map to store elements
	aaEPP := make(map[string]domain.FseppElement)
	for i, record := range records {
		// Assuming the CSV has the same structure as the FseppSefKpr entity
		if i > 1 { // Skip header line
			pkey := fmt.Sprintf("%s_%s", record[5], record[1]) // primary key based on POLJE_PIB and DOKUM
			// Check if key already exists
			if _, exists := aaEPP[pkey]; !exists {
				stEL := domain.FseppElement{
					Polje_SEF: getField(record, 0),
					Polje_KPR: "",
					Dokum:     getField(record, 1),
					Izvor:     getField(record, 3),
					PIB:       getField(record, 5),
					OSN1_SEF:  common.StringToFloat64(getField(record, 8)),
					OSN1_KPR:  0,
					PDV1_SEF:  common.StringToFloat64(getField(record, 9)),
					PDV11_KPR: 0,
					PDV12_KPR: 0,
					OSN2_SEF:  common.StringToFloat64(getField(record, 10)),
					OSN2_KPR:  0,
					PDV2_SEF:  common.StringToFloat64(getField(record, 11)),
					PDV21_KPR: 0,
					PDV22_KPR: 0,
					SistemID:  getField(record, 2),
					Status:    getField(record, 4),
					DateV:     common.StringToDate(getField(record, 6), "02.01.2006"),
					DateOb:    common.StringToDate(getField(record, 7), "02.01.2006"),
					OdDat:     "",
					DoDat:     "",
				}
				aaEPP[pkey] = stEL
			}
		}
	}

	err = s.GetKprDataForSefData(ctx, &aaEPP, odDatuma, doDatuma)
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(aaEPP), pageSize)
		return nil
	}

	// Convert map to slice for pagination
	elements := make([]domain.FseppElement, 0, len(aaEPP))
	for _, element := range aaEPP {
		elements = append(elements, element)
	}

	// Get total records (before pagination)
	totalRecords := len(elements)
	common.SetTableTotalRecords(tbl, totalRecords, pageSize)

	// Apply pagination
	startIdx := (currentPage - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalRecords {
		startIdx = totalRecords - 1
		if startIdx < 0 {
			startIdx = 0
		}
	}
	if endIdx > totalRecords {
		endIdx = totalRecords
	}

	// Get paginated slice
	paginatedElements := elements[startIdx:endIdx]

	// Set table headers
	tbl.Headers = s.GetSefKprTableFields()

	// Populate table rows with paginated data
	for i, element := range paginatedElements {
		fields := []string{
			element.Polje_SEF,
			element.Polje_KPR,
			element.Dokum,
			element.Izvor,
			element.PIB,
			common.FormatNumberWithSystemLocale(element.OSN1_SEF, 2),
			common.FormatNumberWithSystemLocale(element.OSN1_KPR, 2),
			common.FormatNumberWithSystemLocale(element.PDV1_SEF, 2),
			common.FormatNumberWithSystemLocale(element.PDV11_KPR, 2),
			common.FormatNumberWithSystemLocale(element.PDV12_KPR, 2),
			common.FormatNumberWithSystemLocale(element.OSN2_SEF, 2),
			common.FormatNumberWithSystemLocale(element.OSN2_KPR, 2),
			common.FormatNumberWithSystemLocale(element.PDV2_SEF, 2),
			common.FormatNumberWithSystemLocale(element.PDV21_KPR, 2),
			common.FormatNumberWithSystemLocale(element.PDV22_KPR, 2),
			element.SistemID,
			element.DateV.Format(common.HtmlLayout),
			element.DateOb.Format(common.HtmlLayout),
			odDatuma,
			doDatuma,
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d_%s", i, element.Dokum), Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

func (s *FseppResource) GetKprDataForSefData(ctx context.Context, aaEpp *map[string]domain.FseppElement, odDatuma, doDatuma string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	b := querybuilder.New(`select dokum, pib, epp_polje, osnovicavt, pretpdv1vt, pretpdv2vt, osnovicant, pretpdv1nt, pretpdv2nt
	    	from kpr`).
		AddCondition("dknjiz", odDatuma, ">=").
		AddCondition("dknjiz", doDatuma, "<=")
	if err := s.kprRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return err
	}
	entities, err := s.kprRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			pkey := fmt.Sprintf("%s_%s", entity.PIB, entity.Dokum)
			if _, exists := (*aaEpp)[pkey]; exists {
				stEL := domain.FseppElement{
					Polje_SEF: "",
					Polje_KPR: entity.EppPolje,
					Dokum:     entity.Dokum,
					Izvor:     "",
					PIB:       entity.PIB,
					OSN1_SEF:  0,
					OSN1_KPR:  entity.OsnovicaVT,
					PDV1_SEF:  0,
					PDV11_KPR: entity.PretPDV1VT,
					PDV12_KPR: entity.PretPDV2VT,
					OSN2_SEF:  0,
					OSN2_KPR:  entity.OsnovicaNT,
					PDV2_SEF:  0,
					PDV21_KPR: entity.PretPDV1NT,
					PDV22_KPR: entity.PretPDV2NT,
					SistemID:  "0",
					Status:    "",
					OdDat:     odDatuma,
					DoDat:     doDatuma,
				}
				(*aaEpp)[pkey] = stEL
			} else {
				stEL := (*aaEpp)[pkey]
				stEL.Polje_KPR = entity.EppPolje
				stEL.OSN1_KPR = entity.OsnovicaVT
				stEL.PDV11_KPR = entity.PretPDV1VT
				stEL.PDV12_KPR = entity.PretPDV2VT
				stEL.OSN2_KPR = entity.OsnovicaNT
				stEL.PDV21_KPR = entity.PretPDV1NT
				stEL.PDV22_KPR = entity.PretPDV2NT
				stEL.OdDat = odDatuma
				stEL.DoDat = doDatuma
				(*aaEpp)[pkey] = stEL
			}
		}
	}
	return nil
}

// Generic helper to get value from struct by field name (indirection)
// Returns any that can be cast to the appropriate type
func GetFieldValue(obj any, fieldName string) any {
	v := reflect.ValueOf(obj)

	// If pointer, get the element
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// If it's a map
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if strings.EqualFold(key.String(), fieldName) {
				val := v.MapIndex(key)
				if val.IsValid() {
					return val.Interface()
				}
			}
		}
		return nil
	}

	// If it's a struct
	if v.Kind() == reflect.Struct {
		t := v.Type()

		// Try by struct tag first
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("db")
			if strings.EqualFold(tag, fieldName) {
				val := v.Field(i)
				if val.IsValid() {
					return val.Interface()
				}
			}
		}

		// Try by field name (case insensitive)
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if strings.EqualFold(field.Name, fieldName) {
				val := v.Field(i)
				if val.IsValid() {
					return val.Interface()
				}
			}
		}
	}

	return nil
}

// Helper to safely get CSV field
func getField(record []string, index int) string {
	if index < len(record) {
		return strings.TrimSpace(record[index])
	}
	return ""
}

// GetFieldCache returns the cached field structure
func (s *FseppResource) GetFieldCache() map[string]reflect.StructField {
	if s.fseppService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.fseppService.GetFieldCache()
}

// setServiceFieldValues initializes table field definitions for FSEPP
func (s *FseppResource) setServiceFieldValues() {
	// Fields for Sekcije i izvori
	s.sekcijeIzvoriTableFields = []domain.Fields{
		{Name: "deo", Label: "Deo", Width: "8", Field: "fsepp.deo", SkipInSearch: false},
		{Name: "sekcija", Label: "Sekcija", Width: "15", Field: "fsepp.sekcija", SkipInSearch: false},
		{Name: "izvor", Label: "Izvor", Width: "15", Field: "fsepp.izvor", SkipInSearch: false},
		{Name: "polje", Label: "Polje", Width: "15", Field: "fsepp.polje", SkipInSearch: false},
		{Name: "opis", Label: "Naziv polja", Width: "25", Field: "fsepp.opis", SkipInSearch: false},
		{Name: "nipo", Label: "Nivo", Width: "10", Field: "fsepp.nipo", SkipInSearch: false},
		{Name: "aktosn1", Label: "Aktivan Osnovica 1", Width: "12", Field: "fsepp.aktosn1", Type: "checkbox", SkipInSearch: true},
		{Name: "aktpdv1", Label: "Aktivan PDV 1", Width: "12", Field: "fsepp.aktpdv1", Type: "checkbox", SkipInSearch: true},
		{Name: "aktosn2", Label: "Aktivan Osnovica 2", Width: "12", Field: "fsepp.aktosn2", Type: "checkbox", SkipInSearch: true},
		{Name: "aktpdv2", Label: "Aktivan PDV 2", Width: "12", Field: "fsepp.aktpdv2", Type: "checkbox", SkipInSearch: true},
		{Name: "kpr_polje1o", Label: "KPR Polje Osnovica viša", Width: "15", Field: "fsepp.kpr_polje1o", SkipInSearch: true},
		{Name: "kpr_polje1p", Label: "KPR Polje PDV viša", Width: "15", Field: "fsepp.kpr_polje1p", SkipInSearch: true},
		{Name: "kpr_polje2o", Label: "KPR Polje Osnovica niža", Width: "15", Field: "fsepp.kpr_polje2o", SkipInSearch: true},
		{Name: "kpr_polje2p", Label: "KPR Polje PDV niža", Width: "15", Field: "fsepp.kpr_polje2p", SkipInSearch: true},
		{Name: "pozic_1", Label: "Pozicija 1", Width: "10", Field: "fsepp.pozic_1", SkipInSearch: true},
		{Name: "pozic_2", Label: "Pozicija 2", Width: "10", Field: "fsepp.pozic_2", SkipInSearch: true},
		{Name: "pozic_3", Label: "Pozicija 3", Width: "10", Field: "fsepp.pozic_3", SkipInSearch: true},
		{Name: "pozic_4", Label: "Pozicija 4", Width: "10", Field: "fsepp.pozic_4", SkipInSearch: true},
		{Name: "pozic_5", Label: "Pozicija 5", Width: "10", Field: "fsepp.pozic_5", SkipInSearch: true},
		{Name: "oddat", Label: "Od datuma", Width: "10", Field: "fsepp.oddat", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "10", Field: "fsepp.dodat", SkipInSearch: false},
		{Name: "gprdok", Label: "Tipdok", Width: "10", Field: "fsepp.gprdok", SkipInSearch: true},
		{Name: "tip", Label: "Tip", Width: "10", Field: "fsepp.tip", SkipInSearch: true},
		{Name: "autosef", Label: "Auto", Width: "10", Field: "fsepp.autosef", Type: "checkbox", SkipInSearch: true},
	}
	// Fields for Evidencija PP
	s.evidencijaTableFields = []domain.Fields{
		{Name: "polje", Label: "Polje", Width: "15", Field: "fsepp.polje", SkipInSearch: false},
		{Name: "opis", Label: "Naziv polja", Width: "25", Field: "fsepp.opis", SkipInSearch: false},
		{Name: "osn1", Label: "Osnovica 1", Width: "12", Field: "fsepp.osn1", SkipInSearch: false},
		{Name: "pdv1", Label: "PDV 1", Width: "12", Field: "fsepp.pdv1", SkipInSearch: false},
		{Name: "osn2", Label: "Osnovica 2", Width: "12", Field: "fsepp.osn2", SkipInSearch: false},
		{Name: "pdv2", Label: "PDV 2", Width: "12", Field: "fsepp.pdv2", SkipInSearch: false},
		{Name: "oddat", Label: "Od datuma", Width: "10", Field: "fsepp.oddat", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "10", Field: "fsepp.dodat", SkipInSearch: false},
		{Name: "nipo", Label: "Nivo podatka", Width: "10", Field: "fsepp.nipo", SkipInSearch: false},
	}

	// Fields for SEF-KPR
	s.sefKprTableFields = []domain.Fields{
		{Name: "polje_sef", Label: "Polje SEF", Width: "8", Field: "fseppkpr.popis_sef", SkipInSearch: false},
		{Name: "polje_kpr", Label: "Polje KPR", Width: "8", Field: "fseppkpr.popis_kpr", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "12", Field: "fseppkpr.brdokumenta", SkipInSearch: false},
		{Name: "izvor", Label: "Izvor", Width: "10", Field: "fseppkpr.izvor", SkipInSearch: false},
		{Name: "pib", Label: "PIB", Width: "12", Field: "fseppkpr.pib", SkipInSearch: false},
		{Name: "osn1_sef", Label: "Osnovica opšta stopa SEF", Width: "15", Field: "fseppkpr.osn_opsta_stopa_sef", SkipInSearch: false},
		{Name: "osn1_kpr", Label: "Osnovica opšta stopa KPR", Width: "15", Field: "fseppkpr.osn_opsta_stopa_kpr", SkipInSearch: false},
		{Name: "pdv1_sef", Label: "PDV opšta stopa SEF", Width: "15", Field: "fseppkpr.pdv_opsta_stopa_sef", SkipInSearch: false},
		{Name: "pdv11_kpr", Label: "PDV opsta stopa KPR sa pravom", Width: "15", Field: "fseppkpr.pdv_opsta_stopa_kpr", SkipInSearch: false},
		{Name: "pdv12_kpr", Label: "PDV opsta stopa KPR bez prava", Width: "12", Field: "fseppkpr.pdv_opsta_stopa", SkipInSearch: false},
		{Name: "osn2_sef", Label: "Osnovica posebna stopa SEF", Width: "15", Field: "fseppkpr.osn_posebna_stopa_sef", SkipInSearch: false},
		{Name: "osn2_kpr", Label: "Osnovica posebna stopa KPR", Width: "15", Field: "fseppkpr.osn_posebna_stopa_kpr", SkipInSearch: false},
		{Name: "pdv2_sef", Label: "PDV posebna stopa SEF", Width: "15", Field: "fseppkpr.pdv_posebna_stopa_sef", SkipInSearch: false},
		{Name: "pdv21_kpr", Label: "PDV posebna stopa KPR sa pravom", Width: "15", Field: "fseppkpr.pdv_posebna_stopa_kpr", SkipInSearch: false},
		{Name: "pdv22_kpr", Label: "PDV posebna stopa KPR bez prava", Width: "15", Field: "fseppkpr.pdv_posebna_stopa_kpr_bez", SkipInSearch: false},
		{Name: "sistemid", Label: "Sistem ID", Width: "10", Field: "fseppkpr.sistem_id", SkipInSearch: true},
		{Name: "datev", Label: "Datum evidencije", Width: "12", Field: "fseppkpr.datum_evidencije", SkipInSearch: false},
		{Name: "datob", Label: "Datum obrade", Width: "12", Field: "fseppkpr.datum_obrade", SkipInSearch: false},
		{Name: "oddat", Label: "Od datuma", Width: "10", Field: "fseppkpr.od_datuma", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "10", Field: "fseppkpr.do_datuma", SkipInSearch: false},
	}
}

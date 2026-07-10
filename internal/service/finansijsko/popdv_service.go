package finansijsko

import (
	"context"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// PopdvService defines the interface for operations related to POPDV (Prijava PDV od nabavke dobara i usluga).
type PopdvService interface {
	GetPolja(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetPrijava(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odDatuma, doDatuma, searchText string) error
	GetStampa(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetPoljaTableFields() []domain.Fields
	GetPrijavaTableFields() []domain.Fields
	GetStampaTableFields() []domain.Fields
	GetFieldCache() map[string]reflect.StructField
}

// PopdvResource implements the PopdvService interface.
type PopdvResource struct {
	popdvService       *service.BaseService[domain.Popdv]
	popdvRepo          *repository.BaseRepository[domain.Popdv]
	poljaTableFields   []domain.Fields
	prijavaTableFields []domain.Fields
	stampaTableFields  []domain.Fields
}

func NewPopdvService(
	popdvService *service.BaseService[domain.Popdv],
	popdvRepo *repository.BaseRepository[domain.Popdv],
) *PopdvResource {
	rs := &PopdvResource{
		popdvService: popdvService,
		popdvRepo:    popdvRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetPoljaTableFields returns the table field definitions for Polja
func (s *PopdvResource) GetPoljaTableFields() []domain.Fields {
	return s.poljaTableFields
}

// GetPrijavaTableFields returns the table field definitions for Prijava
func (s *PopdvResource) GetPrijavaTableFields() []domain.Fields {
	return s.prijavaTableFields
}

// GetStampaTableFields returns the table field definitions for Stampa
func (s *PopdvResource) GetStampaTableFields() []domain.Fields {
	return s.stampaTableFields
}

// GetFieldCache returns the field cache for the service
func (s *PopdvResource) GetFieldCache() map[string]reflect.StructField {
	return s.popdvRepo.GetFieldCache()
}

// GetPolja retrieves data for POPDV Polja
func (s *PopdvResource) GetPolja(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.popdvRepo.GetHasGodHasKar()

	// Build query for Polja
	qb := common.NewQueryBuilder(`SELECT
				popdv.popdvid, popdv.god, popdv.kar, popdv.tip,
				popdv.deo, popdv.polje, popdv.opis, popdv.naknada,
				popdv.osn1, popdv.pdv1, popdv.osn2, popdv.pdv2,
				popdv.iznos, popdv.poljvred, popdv.poljpdv,
				popdv.nipo, popdv.pppdv_polje,
				popdv.pozic_1, popdv.pozic_2, popdv.pozic_3,
				popdv.pozic_4, popdv.pozic_5, popdv.pozic_6,
				popdv.pozic_7, popdv.pozic_8, popdv.pozic_9,
				popdv.pozic_10, popdv.pozic_11, popdv.pozic_12,
				popdv.npredznak, popdv.aktosn1, popdv.aktpdv1,
				popdv.aktosn2, popdv.aktpdv2,
				CASE popdv.npredznak 
					WHEN 1 THEN '1 - Povecanje i smanjenje' 
					WHEN 2 THEN '2 - Povecanje' 
					WHEN 3 THEN '3 - Smanjenje' 
					ELSE '' 
				END AS predznaktxt,
				popdv.oddat, popdv.dodat, popdv.fvepdvid,
				popdv.vkrbr_fvknjrac, popdv.ter_partneri,
				popdv.tippdv_partneri, popdv.prioritet,
				popdv.labela_aktosn1, popdv.labela_aktpdv1,
				popdv.labela_aktosn2, popdv.labela_aktpdv2,
				popdv.povpolje1, popdv.povpolje2
			FROM 
				popdv`, true)

	if hasGod {
		qb.AddEqual("popdv.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("popdv.kar", session.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fsepp{}))
		qb.AddSearchConditions(s.GetPoljaTableFields(), searchText)
	}
	qb.AddOrderBy(` ORDER BY deo ASC, popdvid ASC`)

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.popdvRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%d", entity.Deo),
				fmt.Sprintf("%d", entity.Tip),
				entity.Polje,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Osn1, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv1, 2),
				common.FormatNumberWithSystemLocale(entity.Osn2, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv2, 2),
				common.FormatNumberWithSystemLocale(entity.Naknada, 2),
				common.FormatNumberWithSystemLocale(entity.Poljvred, 2),
				common.FormatNumberWithSystemLocale(entity.Poljpdv, 2),
				fmt.Sprintf("%d", entity.Nipo),
				entity.Oddat.Format(common.HtmlLayout),
				entity.Dodat.Format(common.HtmlLayout),
				common.FormatNumberWithSystemLocale(entity.Aktosn1, 2),
				common.FormatNumberWithSystemLocale(entity.Aktpdv1, 2),
				common.FormatNumberWithSystemLocale(entity.Aktosn2, 2),
				common.FormatNumberWithSystemLocale(entity.Aktpdv2, 2),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.Popdvid), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

// GetPrijava retrieves data for POPDV Prijava report
func (s *PopdvResource) GetPrijava(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odDatuma, doDatuma, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.popdvRepo.GetHasGodHasKar()

	// Build query for Prijava
	qb := common.NewQueryBuilder(`select popdv.popdvid, popdv.deo, popdv.tip, popdv.polje, popdv.opis, 
		popdv.osn1, popdv.pdv1, popdv.osn2, popdv.pdv2, popdv.poljvred, popdv.poljpdv, popdv.nipo, 
		popdv.oddat, popdv.dodat from popdv`, true)

	if hasGod {
		qb.AddEqual("popdv.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("popdv.kar", session.SelectedKar)
	}

	// Add date range filters if provided
	if odDatuma != "" {
		qb.AddCondition("popdv.oddat", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("popdv.dodat", doDatuma, "<=")
	}

	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fsepp{}))
		qb.AddSearchConditions(s.GetPrijavaTableFields(), searchText)
	}
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.popdvRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%d", entity.Deo),
				fmt.Sprintf("%d", entity.Tip),
				entity.Polje,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Osn1, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv1, 2),
				common.FormatNumberWithSystemLocale(entity.Osn2, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv2, 2),
				common.FormatNumberWithSystemLocale(entity.Poljvred, 2),
				common.FormatNumberWithSystemLocale(entity.Poljpdv, 2),
				fmt.Sprintf("%d", entity.Nipo),
				entity.Oddat.Format(common.HtmlLayout),
				entity.Dodat.Format(common.HtmlLayout),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.Popdvid), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

// GetStampa retrieves data for POPDV Stampa
func (s *PopdvResource) GetStampa(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.popdvRepo.GetHasGodHasKar()

	// Build query for Stampa
	qb := common.NewQueryBuilder(`select popdv.popdvid, popdv.deo, popdv.tip, popdv.polje, popdv.opis, 
		popdv.osn1, popdv.pdv1, popdv.osn2, popdv.pdv2, popdv.naknada, popdv.nipo, popdv.oddat, popdv.dodat, 
		popdv.aktosn1, popdv.aktpdv1, popdv.aktosn2, popdv.aktpdv2 from popdv`, true)

	if hasGod {
		qb.AddEqual("popdv.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("popdv.kar", session.SelectedKar)
	}

	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fsepp{}))
		qb.AddSearchConditions(s.GetPoljaTableFields(), searchText)
	}
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.popdvRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%d", entity.Deo),
				fmt.Sprintf("%d", entity.Tip),
				entity.Polje,
				entity.Opis,
				fmt.Sprintf("%.2f", entity.Osn1),
				fmt.Sprintf("%.2f", entity.Pdv1),
				fmt.Sprintf("%.2f", entity.Osn2),
				fmt.Sprintf("%.2f", entity.Pdv2),
				fmt.Sprintf("%.2f", entity.Naknada),
				fmt.Sprintf("%d", entity.Nipo),
				entity.Oddat.Format(common.HtmlLayout),
				entity.Dodat.Format(common.HtmlLayout),
				common.FormatNumberWithSystemLocale(entity.Aktosn1, 2),
				common.FormatNumberWithSystemLocale(entity.Aktpdv1, 2),
				common.FormatNumberWithSystemLocale(entity.Aktosn2, 2),
				common.FormatNumberWithSystemLocale(entity.Aktpdv2, 2),
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.Popdvid), Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

// setServiceFieldValues initializes table field definitions for POPDV
func (s *PopdvResource) setServiceFieldValues() {
	// Fields for Polja POPDV
	s.poljaTableFields = []domain.Fields{
		{Name: "tip", Label: "Tip podatka", Width: "15", Field: "popdv.tip", SkipInSearch: false},
		{Name: "deo", Label: "Deo", Width: "8", Field: "popdv.deo", SkipInSearch: false},
		{Name: "polje", Label: "Polje", Width: "15", Field: "popdv.polje", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "popdv.opis", SkipInSearch: false},
		{Name: "povpolje1", Label: "Povezano polje 1", Width: "12", Field: "popdv.osn1", SkipInSearch: false},
		{Name: "povpolje2", Label: "Povezano polje 2", Width: "12", Field: "popdv.pdv1", SkipInSearch: false},
		{Name: "nipo", Label: "Nivo podatka", Width: "8", Field: "popdv.nipo", SkipInSearch: false},
		{Name: "pppdvpolje", Label: "Polje iz PPPDV prijave", Width: "12", Field: "popdv.osn2", SkipInSearch: false},
		{Name: "predznaktxt", Label: "Predznak", Width: "12", Field: "popdv.pdv2", SkipInSearch: false},
		{Name: "aktosn1", Label: "Aktivan OSN1", Width: "12", Field: "popdv.osn2", SkipInSearch: false},
		{Name: "labelaaktosn1", Label: "Labela za OSN1", Width: "12", Field: "popdv.pdv2", SkipInSearch: false},
		{Name: "aktpdv1", Label: "Aktivan PDV1", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "labelaaktpdv1", Label: "Labela za PDV1", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "aktonsn2", Label: "Aktivan OSN2", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "labelaaktosn2", Label: "Labela za OSN2", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "aktpdv2", Label: "Aktivan PDV2", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "labelaaktpdv2", Label: "Labela za PDV2", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "naknada", Label: "Naknada", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "osn1", Label: "OSN1", Width: "12", Field: "popdv.osn1", SkipInSearch: false},
		{Name: "pdv1", Label: "PDV1", Width: "12", Field: "popdv.pdv1", SkipInSearch: false},
		{Name: "osn2", Label: "OSN2", Width: "12", Field: "popdv.osn2", SkipInSearch: false},
		{Name: "pdv2", Label: "PDV2", Width: "12", Field: "popdv.pdv2", SkipInSearch: false},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "popdv.iznos", SkipInSearch: false},
		{Name: "poljvred", Label: "Poljvred", Width: "12", Field: "popdv.poljvred", SkipInSearch: false},
		{Name: "pozic1", Label: "Pozicija 1", Width: "12", Field: "popdv.pozic1", SkipInSearch: false},
		{Name: "pozic2", Label: "Pozicija 2", Width: "12", Field: "popdv.pozic2", SkipInSearch: false},
		{Name: "pozic3", Label: "Pozicija 3", Width: "12", Field: "popdv.pozic3", SkipInSearch: false},
		{Name: "pozic4", Label: "Pozicija 4", Width: "12", Field: "popdv.pozic4", SkipInSearch: false},
		{Name: "pozic5", Label: "Pozicija 5", Width: "12", Field: "popdv.pozic5", SkipInSearch: false},
		{Name: "pozic6", Label: "Pozicija 6", Width: "12", Field: "popdv.pozic6", SkipInSearch: false},
		{Name: "pozic7", Label: "Pozicija 7", Width: "12", Field: "popdv.pozic7", SkipInSearch: false},
		{Name: "pozic8", Label: "Pozicija 8", Width: "12", Field: "popdv.pozic8", SkipInSearch: false},
		{Name: "pozic9", Label: "Pozicija 9", Width: "12", Field: "popdv.pozic9", SkipInSearch: false},
		{Name: "pozic10", Label: "Pozicija 10", Width: "12", Field: "popdv.pozic10", SkipInSearch: false},
		{Name: "pozic11", Label: "Pozicija 11", Width: "12", Field: "popdv.pozic11", SkipInSearch: false},
		{Name: "pozic12", Label: "Pozicija 12", Width: "12", Field: "popdv.pozic12", SkipInSearch: false},
		{Name: "oddat", Label: "Od datuma", Width: "12", Field: "popdv.oddat", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "12", Field: "popdv.dodat", SkipInSearch: false},
		{Name: "prioritet", Label: "Prioritet", Width: "12", Field: "popdv.prioritet", SkipInSearch: false},
	}

	// Fields for POPDV Prijava
	s.prijavaTableFields = []domain.Fields{
		{Name: "deo", Label: "Deo", Width: "8", Field: "popdv.deo", SkipInSearch: false},
		{Name: "tip", Label: "Tip podatka", Width: "15", Field: "popdv.tip", SkipInSearch: false},
		{Name: "polje", Label: "Polje", Width: "15", Field: "popdv.polje", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "popdv.opis", SkipInSearch: false},
		{Name: "osn1", Label: "Osn. viša stopa", Width: "12", Field: "popdv.osn1", SkipInSearch: false},
		{Name: "pdv1", Label: "PDV viša stopa", Width: "12", Field: "popdv.pdv1", SkipInSearch: false},
		{Name: "osn2", Label: "Osn. niža stopa", Width: "12", Field: "popdv.osn2", SkipInSearch: false},
		{Name: "pdv2", Label: "PDV niža stopa", Width: "12", Field: "popdv.pdv2", SkipInSearch: false},
		{Name: "poljvred", Label: "Poljvred", Width: "12", Field: "popdv.poljvred", SkipInSearch: false},
		{Name: "poljpdv", Label: "Poljpdv", Width: "12", Field: "popdv.poljpdv", SkipInSearch: false},
		{Name: "nipo", Label: "Nivo", Width: "8", Field: "popdv.nipo", SkipInSearch: false},
		{Name: "oddat", Label: "Od datuma", Width: "12", Field: "popdv.oddat", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "12", Field: "popdv.dodat", SkipInSearch: false},
	}

	// Fields for POPDV Stampa
	s.stampaTableFields = []domain.Fields{
		{Name: "deo", Label: "Deo", Width: "8", Field: "popdv.deo", SkipInSearch: false},
		{Name: "tip", Label: "Tip podatka", Width: "15", Field: "popdv.tip", SkipInSearch: false},
		{Name: "polje", Label: "Polje", Width: "15", Field: "popdv.polje", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "popdv.opis", SkipInSearch: false},
		{Name: "osn1", Label: "Osn. viša stopa", Width: "12", Field: "popdv.osn1", SkipInSearch: false},
		{Name: "pdv1", Label: "PDV viša stopa", Width: "12", Field: "popdv.pdv1", SkipInSearch: false},
		{Name: "osn2", Label: "Osn. niža stopa", Width: "12", Field: "popdv.osn2", SkipInSearch: false},
		{Name: "pdv2", Label: "PDV niža stopa", Width: "12", Field: "popdv.pdv2", SkipInSearch: false},
		{Name: "naknada", Label: "Naknada", Width: "12", Field: "popdv.naknada", SkipInSearch: false},
		{Name: "nipo", Label: "Nivo", Width: "8", Field: "popdv.nipo", SkipInSearch: false},
		{Name: "oddat", Label: "Od datuma", Width: "12", Field: "popdv.oddat", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "12", Field: "popdv.dodat", SkipInSearch: false},
		{Name: "aktosn1", Label: "Akt. Osn. 1", Width: "10", Field: "popdv.aktosn1", Type: "checkbox", SkipInSearch: true},
		{Name: "aktpdv1", Label: "Akt. PDV 1", Width: "10", Field: "popdv.aktpdv1", Type: "checkbox", SkipInSearch: true},
		{Name: "aktosn2", Label: "Akt. Osn. 2", Width: "10", Field: "popdv.aktosn2", Type: "checkbox", SkipInSearch: true},
		{Name: "aktpdv2", Label: "Akt. PDV 2", Width: "10", Field: "popdv.aktpdv2", Type: "checkbox", SkipInSearch: true},
	}
}

package finansijsko

import (
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"math"
	"reflect"

	"github.com/gin-gonic/gin"
)

// PoreskeKnjigeService defines the interface for operations related to Poreske Knjige (Tax Books).
type PoreskeKnjigeService interface {
	GetTableFields() []domain.Fields
	GetTipoveKnjigaValues(c *gin.Context, comboValues *[]domain.ComboItem, ipVkTip string) error
	GetKnjigaIzdatihRacuna(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetKnjigaPrimljenihRacuna(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPoreskaPrijava(c *gin.Context, poreskaPrijavaData *domain.PoreskaPrijavaData) error
	GetFieldCache() map[string]reflect.StructField
}

// PoreskeKnjigeResource implements the PoreskeKnjigeService interface.
type PoreskeKnjigeResource struct {
	kirService     *service.BaseService[domain.KirPayload]
	kprService     *service.BaseService[domain.KprPayload]
	kirRepo        *repository.BaseRepository[domain.KirPayload]
	kprRepo        *repository.BaseRepository[domain.KprPayload]
	fvknjracRepo   *repository.BaseRepository[domain.Fvknjrac]
	kirTableFields []domain.Fields
	kprTableFields []domain.Fields
}

func NewPoreskeKnjigeService(kirService *service.BaseService[domain.KirPayload], kprService *service.BaseService[domain.KprPayload], kirRepo *repository.BaseRepository[domain.KirPayload], kprRepo *repository.BaseRepository[domain.KprPayload], fvknjracRepo *repository.BaseRepository[domain.Fvknjrac]) *PoreskeKnjigeResource {
	rs := &PoreskeKnjigeResource{
		kirService:   kirService,
		kprService:   kprService,
		kirRepo:      kirRepo,
		kprRepo:      kprRepo,
		fvknjracRepo: fvknjracRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetTableFields returns the table field definitions for Poreske Knjige (KIR - Issued Invoices by default)
func (s *PoreskeKnjigeResource) GetTableFields() []domain.Fields {
	return s.kirTableFields
}

// GetKirTableFields returns the table field definitions for Knjiga izdatih racuna (KIR)
func (s *PoreskeKnjigeResource) GetKirTableFields() []domain.Fields {
	return s.kirTableFields
}

// GetKprTableFields returns the table field definitions for Knjiga primljenih racuna (KPR)
func (s *PoreskeKnjigeResource) GetKprTableFields() []domain.Fields {
	return s.kprTableFields
}

// GetKnjigaValues returns the available knjiga (book type) options
func (s *PoreskeKnjigeResource) GetTipoveKnjigaValues(c *gin.Context, comboValues *[]domain.ComboItem, ipVkTip string) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	qb := common.NewQueryBuilder(`
		SELECT vkrbr, opis FROM fvknjrac `)

	hasGod, hasKar := s.fvknjracRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	qb.AddCondition("vktip", ipVkTip, "=")

	qb.AddOrderBy("vktip ASC, vkrbr ASC")
	sqlQuery, args := qb.Build()
	entities, err := s.fvknjracRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get knjiga values: %w", err)
	}

	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			*comboValues = append(*comboValues, domain.ComboItem{
				Key:   fmt.Sprintf("%v", entity.VkRbr),
				Value: entity.Opis,
			})
		}
	}
	return nil
}

// GetKnjigaIzdatih retrieves data for Knjiga izdatih racuna
func (s *PoreskeKnjigeResource) GetKnjigaIzdatihRacuna(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetTableConfig(tbl, "Knjiga izdatih računa", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.kirRepo.GetHasGodHasKar()
	odKnjige := ""
	doKnjige := ""
	// Get parameters from query
	vKnjige := c.Query("knjiga")
	if vKnjige == "99999" {
		odKnjige = "1"
		doKnjige = "99998"
	} else {
		odKnjige = c.Query("knjiga")
		doKnjige = c.Query("knjiga")
	}

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")
	// Build query for Knjiga izdatih racuna
	qb := common.NewQueryBuilder(`
		SELECT 
		   -- ROW_NUMBER() OVER (ORDER BY kir.dknjiz ASC, kir.vrd ASC, kir.numdok ASC, kir.dokum ASC, kir.idkir ASC) AS krbr,
			kir.krbr, kir.dknjiz, kir.dokum, kir.dizd, kir.naziv,
			kir.pib, kir.iznsapdv, kir.oslobcl24, kir.oslobcl25,
			kir.izvozsapr, kir.izvozbezpr, kir.osn1, kir.pdv1, kir.osn2,
			kir.pdv2, kir.prom1, kir.prom2, kir.vpr, kir.vktip, kir.vkrbr,
			kir.kracun, kir.idkir,
			kir.idfpro, kir.idfvknjrac, kir.numdok,
			partneri.naziv AS naziv_pa,
			partneri.mesto AS mesto_pa,
			partneri.pib AS pib_pa,
			partneri.adresa AS adresa_pa
		FROM kir `)

	qb.AddJoin(`LEFT JOIN partneri ON partneri.idpartneri = kir.idpartneri`)
	if hasGod {
		qb.AddEqual("kir.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kir.kar", session.SelectedKar)
	}
	if odKnjige == doKnjige {
		qb.AddCondition("kir.vkrbr", odKnjige, "=")
	} else {
		qb.AddCondition("kir.vkrbr", odKnjige, ">=")
		qb.AddCondition("kir.vkrbr", doKnjige, "<=")
	}
	if odDatuma != "" {
		qb.AddCondition("kir.dknjiz", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("kir.dknjiz", doDatuma, "<=")
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Kir{}))
		qb.AddSearchConditions(s.GetKirTableFields(), searchText)
	}

	qb.AddOrderBy("kir.krbr ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.kirRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%v", entity.Krbr),
				entity.Dknjiz.Format(common.DateLayout),
				entity.Dizd.Format(common.DateLayout),
				entity.Dokum,
				entity.Naziv,
				entity.PIB,
				common.FormatNumberWithSystemLocale(entity.IznsaPDV, 2),
				common.FormatNumberWithSystemLocale(entity.OslobCL24, 2),
				common.FormatNumberWithSystemLocale(entity.OslobCL25, 2),
				common.FormatNumberWithSystemLocale(entity.IzvozSaPr, 2),
				common.FormatNumberWithSystemLocale(entity.IzvozBezPr, 2),
				common.FormatNumberWithSystemLocale(entity.Osn1, 2),
				common.FormatNumberWithSystemLocale(entity.PDV1, 2),
				common.FormatNumberWithSystemLocale(entity.Osn2, 2),
				common.FormatNumberWithSystemLocale(entity.PDV2, 2),
				fmt.Sprintf("%v", entity.IDKir),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetKirTableFields()

	return nil
}

// GetKnjigaUlaznih retrieves data for Knjiga ulaznih racuna
func (s *PoreskeKnjigeResource) GetKnjigaPrimljenihRacuna(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.kprRepo.GetHasGodHasKar()
	odKnjige := ""
	doKnjige := ""

	common.SetTableConfig(tbl, "Knjiga primljenih računa", "", true, true, false)
	tbl.SearchEnabled = true

	// Get parameters from query
	vKnjige := c.Query("knjiga")
	if vKnjige == "99999" {
		odKnjige = "1"
		doKnjige = "99998"
	} else {
		odKnjige = c.Query("knjiga")
		doKnjige = c.Query("knjiga")
	}

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")
	// Build query for Knjiga ulaznih racuna - similar to GetKnjigaIzdatih
	// but querying incoming invoices from suppliers
	qb := common.NewQueryBuilder(`
		SELECT 
			kpr.drbr, kpr.dknjiz, kpr.dizd, kpr.duvoz,
			kpr.dokum, partneri.naziv as naziv_pa, partneri.pib as pib_pa, kpr.iznsapdv,
			kpr.iznoslob, kpr.nisuobvpdv, kpr.uvozbezpdv, kpr.prethpdv,
			kpr.pretpdv1, kpr.pretpdv2, kpr.uvozpdv, kpr.poljvred, kpr.idkpr
		FROM kpr`)
	qb.AddJoin(`LEFT JOIN partneri ON partneri.idpartneri = kpr.idpartneri`)
	if hasGod {
		qb.AddEqual("kpr.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kpr.kar", session.SelectedKar)
	}
	if odKnjige == odKnjige {
		qb.AddCondition("kpr.vkrbr", odKnjige, "=")
	} else {
		qb.AddCondition("kpr.vkrbr", odKnjige, ">=")
		qb.AddCondition("kpr.vkrbr", doKnjige, "<=")
	}
	if odDatuma != "" {
		qb.AddCondition("kpr.dknjiz", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("kpr.dknjiz", doDatuma, "<=")
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Kpr{}))
		qb.AddSearchConditions(s.GetKprTableFields(), searchText)
	}

	qb.AddOrderBy("kpr.drbr ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.kprRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%v", entity.DRbr),
				entity.DKnjiz.Time.Format(common.DateLayout),
				entity.DIzd.Time.Format(common.DateLayout),
				entity.DUvoz.Time.Format(common.DateLayout),
				entity.Dokum,
				entity.Naziv,
				entity.PIB,
				common.FormatNumberWithSystemLocale(entity.IznsAPDV, 2),
				common.FormatNumberWithSystemLocale(entity.IznosLob, 2),
				common.FormatNumberWithSystemLocale(entity.NisuObvPDV, 2),
				common.FormatNumberWithSystemLocale(entity.UvozBezPDV, 2),
				common.FormatNumberWithSystemLocale(entity.PrethodPDV, 2),
				common.FormatNumberWithSystemLocale(entity.PretPDV1, 2),
				common.FormatNumberWithSystemLocale(entity.PretPDV2, 2),
				common.FormatNumberWithSystemLocale(entity.UvozPDV, 2),
				common.FormatNumberWithSystemLocale(entity.PoljVred, 2),
				fmt.Sprintf("%v", entity.IDKpr),
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true})
		}
	}

	tbl.Headers = s.GetKprTableFields()
	return nil
}

// GetPoreskaPrijava retrieves data for Poreska prijava
func (s *PoreskeKnjigeResource) GetPoreskaPrijava(c *gin.Context, poreskaPrijavaData *domain.PoreskaPrijavaData) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	hasGod, hasKar := s.kirRepo.GetHasGodHasKar()
	// Build query for KIR records - fetch all records, not aggregates
	kirQB := common.NewQueryBuilder(`
		SELECT 	kir.dknjiz, kir.vkrbr, kir.izvozsapr, kir.oslobcl24,
			kir.izvozbezpr, kir.oslobcl25, kir.osn1, kir.osn2,
			kir.pdv1, kir.pdv2, kir.idkir
		FROM kir`)

	if hasGod {
		kirQB.AddEqual("kir.god", session.SelectedGod)
	}
	if hasKar {
		kirQB.AddEqual("kir.kar", session.SelectedKar)
	}
	kirQB.AddCondition("kir.dknjiz", odDatuma, ">=")
	kirQB.AddCondition("kir.dknjiz", doDatuma, "<=")

	// Execute query and process KIR records
	sqlKir, argsKir := kirQB.Build()
	kirEntities, err := s.kirRepo.GetAllCustom(c, sqlKir, "", argsKir, "", "")
	if err != nil {
		return err
	}
	if err == nil && kirEntities != nil && len(*kirEntities) > 0 {
		for _, kir := range *kirEntities {
			poreskaPrijavaData.Edt001 += kir.IzvozSaPr + kir.OslobCL24
			poreskaPrijavaData.Edt002 += kir.IzvozBezPr + kir.OslobCL25
			if kir.Vkrbr != 3 {
				poreskaPrijavaData.Edt003 += kir.Osn1
			}
			if kir.Vkrbr != 3 {
				poreskaPrijavaData.Edt004 += kir.Osn2
			}
			poreskaPrijavaData.Edt103 += kir.PDV1
			poreskaPrijavaData.Edt104 += kir.PDV2
		}
	}

	// Build query for KPR aggregates
	kprQB := common.NewQueryBuilder(`
		SELECT 
			uvozosnpdv, uvozpdv, poljvred, poljpdv, osnbezpdv, iznoslob,
			nisuobvpdv, uvozbezpdv, osnbezpod, pretpdv1
		FROM kpr`)
	hasGod, hasKar = s.kprRepo.GetHasGodHasKar()
	if hasGod {
		kprQB.AddEqual("kpr.god", session.SelectedGod)
	}
	if hasKar {
		kprQB.AddEqual("kpr.kar", session.SelectedKar)
	}
	kprQB.AddCondition("kpr.dknjiz", odDatuma, ">=")
	kprQB.AddCondition("kpr.dknjiz", doDatuma, "<=")

	// Execute query and process KPR records
	sqlKpr, argsKpr := kprQB.Build()
	kprEntities, err := s.kprRepo.GetAllCustom(c, sqlKpr, "", argsKpr, "", "")
	if err != nil {
		return err
	}
	if err == nil && kprEntities != nil && len(*kprEntities) > 0 {
		for _, kpr := range *kprEntities {
			poreskaPrijavaData.Edt006 += kpr.UvozOsnPDV
			poreskaPrijavaData.Edt106 += kpr.UvozPDV
			poreskaPrijavaData.Edt007 += kpr.PoljVred
			poreskaPrijavaData.Edt107 += kpr.PoljPDV
			poreskaPrijavaData.Edt008 += kpr.OsnBezPDV + kpr.IznosLob + kpr.NisuObvPDV + kpr.UvozBezPDV + kpr.OsnBezPod
			poreskaPrijavaData.Edt108 += kpr.PretPDV1
		}
	}

	// Round all individual values
	poreskaPrijavaData.Edt001 = math.Round(poreskaPrijavaData.Edt001)
	poreskaPrijavaData.Edt002 = math.Round(poreskaPrijavaData.Edt002)
	poreskaPrijavaData.Edt003 = math.Round(poreskaPrijavaData.Edt003)
	poreskaPrijavaData.Edt004 = math.Round(poreskaPrijavaData.Edt004)
	poreskaPrijavaData.Edt006 = math.Round(poreskaPrijavaData.Edt006)
	poreskaPrijavaData.Edt007 = math.Round(poreskaPrijavaData.Edt007)
	poreskaPrijavaData.Edt008 = math.Round(poreskaPrijavaData.Edt008)
	poreskaPrijavaData.Edt103 = math.Round(poreskaPrijavaData.Edt103)
	poreskaPrijavaData.Edt104 = math.Round(poreskaPrijavaData.Edt104)
	poreskaPrijavaData.Edt106 = math.Round(poreskaPrijavaData.Edt106)
	poreskaPrijavaData.Edt107 = math.Round(poreskaPrijavaData.Edt107)
	poreskaPrijavaData.Edt108 = math.Round(poreskaPrijavaData.Edt108)

	// Calculate totals
	poreskaPrijavaData.Edt005 = poreskaPrijavaData.Edt001 + poreskaPrijavaData.Edt002 + poreskaPrijavaData.Edt003 + poreskaPrijavaData.Edt004
	poreskaPrijavaData.Edt105 = poreskaPrijavaData.Edt103 + poreskaPrijavaData.Edt104
	poreskaPrijavaData.Edt009 = poreskaPrijavaData.Edt006 + poreskaPrijavaData.Edt007 + poreskaPrijavaData.Edt008
	poreskaPrijavaData.Edt109 = poreskaPrijavaData.Edt106 + poreskaPrijavaData.Edt107 + poreskaPrijavaData.Edt108
	poreskaPrijavaData.Edt110 = math.Round(poreskaPrijavaData.Edt105 - poreskaPrijavaData.Edt109)
	return nil
}

// GetFieldCache returns the cached field structure
func (s *PoreskeKnjigeResource) GetFieldCache() map[string]reflect.StructField {
	if s.kirService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.kirService.GetFieldCache()
}

// setServiceFieldValues initializes table field definitions for Poreske Knjige
func (s *PoreskeKnjigeResource) setServiceFieldValues() {
	// Fields for Knjiga izdatih racuna (KIR - Issued Invoices)
	s.kirTableFields = []domain.Fields{
		{Name: "krbr", Label: "Redni broj", Width: "8", Field: "kir.krbr", SkipInSearch: false},
		{Name: "dknjiz", Label: "Datum knjiženja", Width: "10", Field: "kir.dknjiz", SkipInSearch: false},
		{Name: "dizd", Label: "Datum izdavanja", Width: "10", Field: "kir.dizd", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "kir.dokum", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv kupca", Width: "25", Field: "partneri.naziv", SkipInSearch: false},
		{Name: "pib", Label: "PIB kupca", Width: "12", Field: "partneri.pib", SkipInSearch: false},
		{Name: "iznsapdv", Label: "Iznos sa PDV", Width: "12", Field: "kir.iznsapdv", SkipInSearch: false},
		{Name: "oslobcl24", Label: "Oslobođen CL 24", Width: "12", Field: "kir.oslobcl24", SkipInSearch: false},
		{Name: "oslobcl25", Label: "Oslobođen CL 25", Width: "12", Field: "kir.oslobcl25", SkipInSearch: false},
		{Name: "izvozsapr", Label: "Izvoz sa pravom", Width: "12", Field: "kir.izvozsapr", SkipInSearch: true},
		{Name: "izvozbezpr", Label: "Izvoz bez prava", Width: "12", Field: "kir.izvozbezpr", SkipInSearch: true},
		{Name: "osn1", Label: "Osnova 1", Width: "12", Field: "kir.osn1", SkipInSearch: true},
		{Name: "pdv1", Label: "PDV 1", Width: "12", Field: "kir.pdv1", SkipInSearch: true},
		{Name: "osn2", Label: "Osnova 2", Width: "12", Field: "kir.osn2", SkipInSearch: true},
		{Name: "pdv2", Label: "PDV 2", Width: "12", Field: "kir.pdv2", SkipInSearch: true},
		{Name: "idkir", Label: "ID", Width: "8", Field: "kir.idkir", SkipInSearch: true},
	}

	// Fields for Knjiga primljenih racuna (KPR - Received Invoices)
	s.kprTableFields = []domain.Fields{
		{Name: "drbr", Label: "Redni broj", Width: "8", Field: "kpr.drbr", SkipInSearch: false},
		{Name: "dknjiz", Label: "Datum knjiženja", Width: "10", Field: "kpr.dknjiz", SkipInSearch: false},
		{Name: "dizd", Label: "Datum izdavanja", Width: "10", Field: "kpr.dizd", SkipInSearch: false},
		{Name: "duvoz", Label: "Datum uvoza", Width: "10", Field: "kpr.duvoz", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "kpr.dokum", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv dobavljača", Width: "25", Field: "partneri.naziv", SkipInSearch: false},
		{Name: "pib", Label: "PIB dobavljača", Width: "12", Field: "partneri.pib", SkipInSearch: false},
		{Name: "iznsapdv", Label: "Iznos sa PDV", Width: "12", Field: "kpr.iznsapdv", SkipInSearch: true},
		{Name: "iznoslob", Label: "Iznos oslobođen", Width: "12", Field: "kpr.iznoslob", SkipInSearch: true},
		{Name: "nisuobvpdv", Label: "Nisu obavezni PDV", Width: "12", Field: "kpr.nisuobvpdv", SkipInSearch: true},
		{Name: "uvozbezpdv", Label: "Uvoz bez PDV", Width: "12", Field: "kpr.uvozbezpdv", SkipInSearch: true},
		{Name: "prethpdv", Label: "Prethodnog PDV", Width: "12", Field: "kpr.prethpdv", SkipInSearch: true},
		{Name: "pretpdv1", Label: "Prethodni PDV 1", Width: "12", Field: "kpr.pretpdv1", SkipInSearch: true},
		{Name: "pretpdv2", Label: "Prethodni PDV 2", Width: "12", Field: "kpr.pretpdv2", SkipInSearch: true},
		{Name: "uvozpdv", Label: "Uvoz PDV", Width: "12", Field: "kpr.uvozpdv", SkipInSearch: true},
		{Name: "poljvred", Label: "Polje vrednost", Width: "12", Field: "kpr.poljvred", SkipInSearch: true},
		{Name: "idkpr", Label: "ID", Width: "8", Field: "kpr.idkpr", SkipInSearch: true},
	}
}

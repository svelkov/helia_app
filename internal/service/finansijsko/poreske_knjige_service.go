package finansijsko

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"helia/pkg/utils"
	"math"
	"reflect"
)

// PoreskeKnjigeService defines the interface for operations related to Poreske Knjige (Tax Books).
type PoreskeKnjigeService interface {
	GetTableFields() []domain.Fields
	GetKirTableFields() []domain.Fields
	GetKprTableFields() []domain.Fields
	GetKirStampaTableFields() []domain.Fields
	GetKprStampaTableFields() []domain.Fields
	GetTipoveKnjigaValues(ctx context.Context, comboValues *[]domain.ComboItem, ipVkTip string) error
	GetTipdokValues(ctx context.Context, comboValues *[]domain.ComboItem) error
	GetKnjigaIzdatihRacuna(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.KnjigeParameters) error
	GetKnjigaIzdatihRacunaStampa(ctx context.Context, tbl *domain.TableData, params domain.KnjigeParameters) error
	GetKnjigaPrimljenihRacuna(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.KnjigeParameters) error
	GetKnjigaPrimljenihRacunaStampa(ctx context.Context, tbl *domain.TableData, params domain.KnjigeParameters) error
	GetPoreskaPrijava(ctx context.Context, poreskaPrijavaData *domain.PoreskaPrijavaData) error
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetFieldCache() map[string]reflect.StructField
	KirValidate(ctx context.Context, kir *domain.Kir, cAction string) ([]domain.FieldError, error)
	KprVaidate(ctx context.Context, kpr *domain.Kpr, cAction string) ([]domain.FieldError, error)
	SaveKnjigaIzdatihRacuna(ctx context.Context, kir *domain.Kir, cAction string) error
	SaveKnjigaPrimljenihRacuna(ctx context.Context, kpr *domain.Kpr, cAction string) error
}

// PoreskeKnjigeResource implements the PoreskeKnjigeService interface.
type PoreskeKnjigeResource struct {
	kirService           *service.BaseService[domain.Kir]
	kprService           *service.BaseService[domain.Kpr]
	kirRepo              *repository.BaseRepository[domain.Kir]
	kprRepo              *repository.BaseRepository[domain.Kpr]
	fvknjracRepo         *repository.BaseRepository[domain.Fvknjrac]
	tipdokRepo           *repository.BaseRepository[domain.Tipdok]
	fvrRepo              *repository.BaseRepository[domain.Fvr]
	kirTableFields       []domain.Fields
	kprTableFields       []domain.Fields
	kirStampaTableFields []domain.Fields
	kprStampaTableFields []domain.Fields
}

func NewPoreskeKnjigeService(kirService *service.BaseService[domain.Kir], kprService *service.BaseService[domain.Kpr], kirRepo *repository.BaseRepository[domain.Kir], kprRepo *repository.BaseRepository[domain.Kpr], fvknjracRepo *repository.BaseRepository[domain.Fvknjrac], tipdokRepo *repository.BaseRepository[domain.Tipdok], fvrRepo *repository.BaseRepository[domain.Fvr]) *PoreskeKnjigeResource {
	rs := &PoreskeKnjigeResource{
		kirService:   kirService,
		kprService:   kprService,
		kirRepo:      kirRepo,
		kprRepo:      kprRepo,
		fvknjracRepo: fvknjracRepo,
		tipdokRepo:   tipdokRepo,
		fvrRepo:      fvrRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *PoreskeKnjigeResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return utils.GetFvrData(ctx, s.fvrRepo)
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

// GetKirStampaTableFields returns the print table field definitions for KIR
func (s *PoreskeKnjigeResource) GetKirStampaTableFields() []domain.Fields {
	return s.kirStampaTableFields
}

// GetKprStampaTableFields returns the print table field definitions for KPR
func (s *PoreskeKnjigeResource) GetKprStampaTableFields() []domain.Fields {
	return s.kprStampaTableFields
}

// GetKnjigaValues returns the available knjiga (book type) options
func (s *PoreskeKnjigeResource) GetTipoveKnjigaValues(ctx context.Context, comboValues *[]domain.ComboItem, ipVkTip string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	qb := common.NewQueryBuilder(`SELECT vkrbr, opis FROM fvknjrac `, true)

	hasGod, hasKar := s.fvknjracRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	qb.AddEqual("vktip", ipVkTip)

	qb.AddOrderBy("vktip ASC, vkrbr ASC")
	sqlQuery, args := qb.Build()
	entities, err := s.fvknjracRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get knjiga values: %w", err)
	}

	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			*comboValues = append(*comboValues, domain.ComboItem{
				Key:   fmt.Sprintf("%v", entity.VkRbr),
				Value: fmt.Sprintf("%d - %s", entity.VkRbr, entity.Opis),
			})
		}
		*comboValues = append(*comboValues, domain.ComboItem{Key: "999", Value: "999 - Sve knjige"})
	}
	return nil
}

// GetTipdokOptions fetches the list of tipdok options for filtering. This method stays the same.
func (s *PoreskeKnjigeResource) GetTipdokValues(ctx context.Context, comboValues *[]domain.ComboItem) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idtipdok, tipdok, opis FROM tipdok`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddCustomCondition("(grpdok = 'FIN' OR grpdok = 'SVI')")
	qb.AddOrderBy("tipdok::NUMERIC ASC")
	sqlQuery, args := qb.Build()
	tipdokValues, err := s.tipdokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get tipdok options: %w", err)
	}
	if tipdokValues != nil && len(*tipdokValues) > 0 {
		for _, entity := range *tipdokValues {
			*comboValues = append(*comboValues, domain.ComboItem{
				Key:   fmt.Sprintf("%s", entity.TipDok),
				Value: fmt.Sprintf("%s - %s", entity.TipDok, entity.Opis),
			})
		}
	}
	return nil
}

// GetKnjigaIzdatih retrieves data for Knjiga izdatih racuna
func (s *PoreskeKnjigeResource) GetKnjigaIzdatihRacuna(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.KnjigeParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.kirRepo.GetHasGodHasKar()
	odKnjige := ""
	doKnjige := ""
	// Get parameters from query
	vKnjige := params.Knjiga
	if vKnjige == "999" {
		odKnjige = "1"
		doKnjige = "999"
	} else {
		odKnjige = params.Knjiga
		doKnjige = params.Knjiga
	}
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
			coalesce(partneri.naziv, '') AS naziv_pa,
			coalesce(partneri.mesto, '') AS mesto_pa,
			coalesce(partneri.pib, '') AS pib,
			coalesce(partneri.adresa, '') AS adresa_pa
		FROM kir `, true)

	qb.AddJoin(`LEFT JOIN partneri ON partneri.idpartneri = kir.idpartneri`)
	if hasGod {
		qb.AddEqual("kir.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kir.kar", session.SelectedKar)
	}
	if odKnjige == doKnjige {
		qb.AddCondition("kir.vkrbr", common.StringToInt(odKnjige), "=")
	} else {
		qb.AddCondition("kir.vkrbr", common.StringToInt(odKnjige), ">=")
		qb.AddCondition("kir.vkrbr", common.StringToInt(doKnjige), "<=")
	}
	if params.OdDatuma != "" {
		qb.AddCondition("kir.dknjiz", params.OdDatuma, ">=")
	}
	if params.DoDatuma != "" {
		qb.AddCondition("kir.dknjiz", params.DoDatuma, "<=")
	}
	if params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Kir{}))
		qb.AddSearchConditions(s.GetKirTableFields(), params.SearchText)
	}

	qb.AddOrderBy("kir.krbr ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.kirRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
				common.FormatNullTime(entity.Dknjiz, common.DateLayout),
				common.FormatNullTime(entity.Dizd, common.DateLayout),
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

// GetKnjigaIzdatihRacunaStampa fetches all KIR records (no pagination) and computes totals for printing.
func (s *PoreskeKnjigeResource) GetKnjigaIzdatihRacunaStampa(ctx context.Context, tbl *domain.TableData, params domain.KnjigeParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.kirRepo.GetHasGodHasKar()
	odKnjige := ""
	doKnjige := ""
	vKnjige := params.Knjiga
	if vKnjige == "999" {
		odKnjige = "1"
		doKnjige = "999"
	} else {
		odKnjige = params.Knjiga
		doKnjige = params.Knjiga
	}

	qb := common.NewQueryBuilder(`
		SELECT
			kir.krbr, kir.nalog, kir.dknjiz, kir.dokum, kir.dizd,
			kir.tipdok,
			coalesce(partneri.naziv, kir.naziv, '') AS naziv_pa,
			coalesce(partneri.pib, kir.pib, '') AS pib,
			kir.iznsapdv, kir.oslobcl24, kir.oslobcl25,
			kir.izvozsapr, kir.izvozbezpr, kir.osn1, kir.pdv1, kir.osn2, kir.pdv2
		FROM kir`, true)
	qb.AddJoin(`LEFT JOIN partneri ON partneri.idpartneri = kir.idpartneri`)
	if hasGod {
		qb.AddEqual("kir.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kir.kar", session.SelectedKar)
	}
	if odKnjige == doKnjige {
		qb.AddCondition("kir.vkrbr", common.StringToInt(odKnjige), "=")
	} else {
		qb.AddCondition("kir.vkrbr", common.StringToInt(odKnjige), ">=")
		qb.AddCondition("kir.vkrbr", common.StringToInt(doKnjige), "<=")
	}
	if params.OdDatuma != "" {
		qb.AddCondition("kir.dknjiz", params.OdDatuma, ">=")
	}
	if params.DoDatuma != "" {
		qb.AddCondition("kir.dknjiz", params.DoDatuma, "<=")
	}
	groupByNalog := params.StampajPoNalozima || params.StampajPoNalozimaZbirno
	summaryByNalog := params.StampajPoNalozimaZbirno
	if groupByNalog {
		qb.AddOrderBy("kir.tipdok ASC, kir.nalog ASC, kir.krbr ASC")
	} else {
		qb.AddOrderBy("kir.krbr ASC")
	}
	sqlQuery, args := qb.Build()
	entities, err := s.kirRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	type kirPrintTotals struct {
		iznsaPDV   float64
		oslobCL24  float64
		oslobCL25  float64
		izvozSaPr  float64
		izvozBezPr float64
		osn1       float64
		pdv1       float64
		osn2       float64
		pdv2       float64
	}

	addToTotals := func(t *kirPrintTotals, entity domain.Kir) {
		t.iznsaPDV += entity.IznsaPDV
		t.oslobCL24 += entity.OslobCL24
		t.oslobCL25 += entity.OslobCL25
		t.izvozSaPr += entity.IzvozSaPr
		t.izvozBezPr += entity.IzvozBezPr
		t.osn1 += entity.Osn1
		t.pdv1 += entity.PDV1
		t.osn2 += entity.Osn2
		t.pdv2 += entity.PDV2
	}

	formatDetailRow := func(entity domain.Kir) []string {
		ukPromet := entity.OslobCL24 + entity.OslobCL25 + entity.IzvozSaPr + entity.IzvozBezPr + entity.Osn1 + entity.Osn2
		promSpravom := entity.OslobCL24 + entity.IzvozSaPr + entity.Osn1 + entity.Osn2
		return []string{
			fmt.Sprintf("%v", entity.Krbr),
			common.FormatNullTime(entity.Dknjiz, common.DateLayout),
			entity.Dokum,
			common.FormatNullTime(entity.Dizd, common.DateLayout),
			entity.NativPartnera,
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
			common.FormatNumberWithSystemLocale(ukPromet, 2),
			common.FormatNumberWithSystemLocale(promSpravom, 2),
		}
	}

	formatTotalRow := func(label string, totals kirPrintTotals) []string {
		ukPromet := totals.oslobCL24 + totals.oslobCL25 + totals.izvozSaPr + totals.izvozBezPr + totals.osn1 + totals.osn2
		promSpravom := totals.oslobCL24 + totals.izvozSaPr + totals.osn1 + totals.osn2
		return []string{
			label,
			"",
			"",
			"",
			"",
			"",
			common.FormatNumberWithSystemLocale(totals.iznsaPDV, 2),
			common.FormatNumberWithSystemLocale(totals.oslobCL24, 2),
			common.FormatNumberWithSystemLocale(totals.oslobCL25, 2),
			common.FormatNumberWithSystemLocale(totals.izvozSaPr, 2),
			common.FormatNumberWithSystemLocale(totals.izvozBezPr, 2),
			common.FormatNumberWithSystemLocale(totals.osn1, 2),
			common.FormatNumberWithSystemLocale(totals.pdv1, 2),
			common.FormatNumberWithSystemLocale(totals.osn2, 2),
			common.FormatNumberWithSystemLocale(totals.pdv2, 2),
			common.FormatNumberWithSystemLocale(ukPromet, 2),
			common.FormatNumberWithSystemLocale(promSpravom, 2),
		}
	}

	var (
		totIznsaPDV    float64
		totOslobCL24   float64
		totOslobCL25   float64
		totIzvozSaPr   float64
		totIzvozBezPr  float64
		totOsn1        float64
		totPDV1        float64
		totOsn2        float64
		totPDV2        float64
		totUkPromet    float64
		totPromSpravom float64
	)

	if entities != nil {
		var (
			currentGroup  string
			currentTotals kirPrintTotals
			groupOpen     bool
		)

		flushGroup := func() {
			if !groupOpen {
				return
			}
			if groupByNalog {
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					Fields:   formatTotalRow(fmt.Sprintf("Ukupno za nalog: %s", currentGroup), currentTotals),
					ClassRow: "nalog-total",
				})
			}
			currentTotals = kirPrintTotals{}
			groupOpen = false
		}

		for _, entity := range *entities {
			if groupByNalog {
				groupKey := fmt.Sprintf("%s-%v", entity.TipDok, entity.Nalog)
				if !groupOpen || groupKey != currentGroup {
					flushGroup()
					currentGroup = groupKey
					groupOpen = true
					if !summaryByNalog {
						tbl.Rows = append(tbl.Rows, domain.TableRow{
							Fields:   []string{fmt.Sprintf("Nalog: %s", currentGroup)},
							ClassRow: "nalog-header",
						})
					}
				}
				addToTotals(&currentTotals, entity)
				if !summaryByNalog {
					tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: formatDetailRow(entity)})
				}
			} else {
				tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: formatDetailRow(entity)})
			}

			totIznsaPDV += entity.IznsaPDV
			totOslobCL24 += entity.OslobCL24
			totOslobCL25 += entity.OslobCL25
			totIzvozSaPr += entity.IzvozSaPr
			totIzvozBezPr += entity.IzvozBezPr
			totOsn1 += entity.Osn1
			totPDV1 += entity.PDV1
			totOsn2 += entity.Osn2
			totPDV2 += entity.PDV2
			totUkPromet += entity.OslobCL24 + entity.OslobCL25 + entity.IzvozSaPr + entity.IzvozBezPr + entity.Osn1 + entity.Osn2
			totPromSpravom += entity.OslobCL24 + entity.IzvozSaPr + entity.Osn1 + entity.Osn2
		}
		flushGroup()
	}

	tbl.Headers = s.GetKirStampaTableFields()
	tbl.HasTotals = true
	tbl.Totals = []string{
		"", "UKUPNO:", "", "", "", "",
		common.FormatNumberWithSystemLocale(totIznsaPDV, 2),
		common.FormatNumberWithSystemLocale(totOslobCL24, 2),
		common.FormatNumberWithSystemLocale(totOslobCL25, 2),
		common.FormatNumberWithSystemLocale(totIzvozSaPr, 2),
		common.FormatNumberWithSystemLocale(totIzvozBezPr, 2),
		common.FormatNumberWithSystemLocale(totOsn1, 2),
		common.FormatNumberWithSystemLocale(totPDV1, 2),
		common.FormatNumberWithSystemLocale(totOsn2, 2),
		common.FormatNumberWithSystemLocale(totPDV2, 2),
		common.FormatNumberWithSystemLocale(totUkPromet, 2),
		common.FormatNumberWithSystemLocale(totPromSpravom, 2),
	}
	return nil
}

func (s *PoreskeKnjigeResource) KirValidate(ctx context.Context, kir *domain.Kir, cAction string) ([]domain.FieldError, error) {
	fieldErrors := []domain.FieldError{}
	if cAction == common.ActionAdd {
		entity, err := s.kirService.GetByID(ctx, common.IDKir, int64(kir.IDKir))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error fetching existing record for validation: %w", err)
		}
		if entity != nil {
			return nil, fmt.Errorf("Ovakvi podaci vec postoje... Nedozvoljen dupli unos!!!")
		}
	}

	if kir.Dokum == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dokum", ErrorMessage: "Morate uneti broj dokumenta!!!"})
	}
	if kir.Dizd.Valid == false {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dizd", ErrorMessage: "Morate uneti datum dokumenta!!!"})
	}
	if kir.Dknjiz.Valid == false {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dknjiz", ErrorMessage: "Morate uneti datum knjizenja!!!"})
	}
	if kir.Dizd.Valid && kir.Dknjiz.Valid && kir.Dizd.Time.After(kir.Dknjiz.Time) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dizd", ErrorMessage: "Datum dokumenta ne moze biti veci od datuma knjizenja!!!"})
	}

	// IF EDT_IZNSAPDV <= 0 AND (PomVRD = 1 OR PomVRD = 3)  THEN
	// 	gbOK = False
	// 	Info("Morate uneti ukupan iznos sa PDV-om ili iznos mora biti veci od 0 !!!")
	// 	ReturnToCapture(EDT_IZNSAPDV)
	// END
	// IF EDT_IZNSAPDV >= 0 AND (PomVRD = 2 OR PomVRD = 4) THEN
	// 	gbOK = False
	// 	Info("Morate uneti ukupan iznos sa PDV-om ili iznos mora biti manji od 0 !!!")
	// 	ReturnToCapture(EDT_IZNSAPDV)
	// END
	if kir.IznsaPDV != kir.Osn1+kir.PDV1+kir.Osn2+kir.PDV2 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "iznsaPDV", ErrorMessage: "Morate proveriti ukupan iznos, osnovicu i PDV !!!"})
	}

	return fieldErrors, nil

}
func (s *PoreskeKnjigeResource) SaveKnjigaIzdatihRacuna(ctx context.Context, kir *domain.Kir, cAction string) error {

	return nil
}

// GetKnjigaPrimljenihRacunaStampa fetches all KPR records (no pagination) and computes totals for printing.
func (s *PoreskeKnjigeResource) GetKnjigaPrimljenihRacunaStampa(ctx context.Context, tbl *domain.TableData, params domain.KnjigeParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.kprRepo.GetHasGodHasKar()
	odKnjige := ""
	doKnjige := ""
	vKnjige := params.Knjiga
	if vKnjige == "999" {
		odKnjige = "1"
		doKnjige = "999"
	} else {
		odKnjige = params.Knjiga
		doKnjige = params.Knjiga
	}

	qb := common.NewQueryBuilder(`
		SELECT
			kpr.drbr, kpr.nalog, kpr.tipdok, kpr.dknjiz, kpr.duvoz, kpr.dokum, kpr.dizd,
			coalesce(partneri.naziv, kpr.naziv, '') AS naziv,
			coalesce(partneri.pib, kpr.pib, '') AS pib,
			kpr.iznsapdv, kpr.prethpdv, kpr.iznoslob, kpr.nisuobvpdv, kpr.uvozbezpdv,
			kpr.prethpdv, kpr.pretpdv1, kpr.pretpdv2, kpr.uvozpdv, kpr.poljvred, kpr.poljpdv
		FROM kpr`, true)
	qb.AddJoin(`LEFT JOIN partneri ON partneri.idpartneri = kpr.idpartneri`)
	if hasGod {
		qb.AddEqual("kpr.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kpr.kar", session.SelectedKar)
	}
	if odKnjige == doKnjige {
		qb.AddCondition("kpr.vkrbr", common.StringToInt(odKnjige), "=")
	} else {
		qb.AddCondition("kpr.vkrbr", common.StringToInt(odKnjige), ">=")
		qb.AddCondition("kpr.vkrbr", common.StringToInt(doKnjige), "<=")
	}
	if params.OdDatuma != "" {
		qb.AddCondition("kpr.dknjiz", params.OdDatuma, ">=")
	}
	if params.DoDatuma != "" {
		qb.AddCondition("kpr.dknjiz", params.DoDatuma, "<=")
	}

	groupByNalog := params.StampajPoNalozima || params.StampajPoNalozimaZbirno
	summaryByNalog := params.StampajPoNalozimaZbirno
	if groupByNalog {
		qb.AddOrderBy("kpr.tipdok ASC, kpr.nalog ASC, kpr.drbr ASC")
	} else {
		qb.AddOrderBy("kpr.drbr ASC")
	}
	sqlQuery, args := qb.Build()
	entities, err := s.kprRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	type kprPrintTotals struct {
		iznsAPDV    float64
		naknadaPDV  float64
		oslobodjeno float64
		nabavkeBez  float64
		uvozBez     float64
		ukupno      float64
		iznos14     float64
		iznos15     float64
		uvoz21      float64
		uvoz22      float64
		polj25      float64
		polj24      float64
	}

	addToTotals := func(t *kprPrintTotals, entity domain.Kpr) {
		t.iznsAPDV += entity.IznsAPDV
		t.naknadaPDV += entity.PrethodPDV
		t.oslobodjeno += entity.IznosLob
		t.nabavkeBez += entity.NisuObvPDV
		t.uvozBez += entity.UvozBezPDV
		t.ukupno += entity.PrethodPDV + entity.PretPDV1 + entity.PretPDV2 + entity.UvozPDV
		t.iznos14 += entity.PrethodPDV
		t.iznos15 += entity.PretPDV1
		t.uvoz21 += entity.PretPDV2
		t.uvoz22 += entity.UvozPDV
		t.polj25 += entity.PoljVred
		t.polj24 += entity.PoljPDV
	}

	formatDetailRow := func(entity domain.Kpr) []string {
		ukupno := entity.PrethodPDV + entity.PretPDV1 + entity.PretPDV2 + entity.UvozPDV
		return []string{
			fmt.Sprintf("%v", entity.DRbr),
			common.FormatNullTime(entity.DKnjiz, common.DateLayout),
			common.FormatNullTime(entity.DUvoz, common.DateLayout),
			entity.Dokum,
			common.FormatNullTime(entity.DIzd, common.DateLayout),
			entity.Naziv,
			entity.PIB,
			common.FormatNumberWithSystemLocale(entity.IznsAPDV, 2),
			common.FormatNumberWithSystemLocale(entity.PrethodPDV, 2),
			common.FormatNumberWithSystemLocale(entity.IznosLob, 2),
			common.FormatNumberWithSystemLocale(entity.NisuObvPDV, 2),
			common.FormatNumberWithSystemLocale(entity.UvozBezPDV, 2),
			common.FormatNumberWithSystemLocale(ukupno, 2),
			common.FormatNumberWithSystemLocale(entity.PrethodPDV, 2),
			common.FormatNumberWithSystemLocale(entity.PretPDV1, 2),
			common.FormatNumberWithSystemLocale(entity.PretPDV2, 2),
			common.FormatNumberWithSystemLocale(entity.UvozPDV, 2),
			common.FormatNumberWithSystemLocale(entity.PoljVred, 2),
			common.FormatNumberWithSystemLocale(entity.PoljPDV, 2),
		}
	}

	formatTotalRow := func(label string, totals kprPrintTotals) []string {
		return []string{
			label,
			"",
			"",
			"",
			"",
			"",
			"",
			common.FormatNumberWithSystemLocale(totals.iznsAPDV, 2),
			common.FormatNumberWithSystemLocale(totals.naknadaPDV, 2),
			common.FormatNumberWithSystemLocale(totals.oslobodjeno, 2),
			common.FormatNumberWithSystemLocale(totals.nabavkeBez, 2),
			common.FormatNumberWithSystemLocale(totals.uvozBez, 2),
			common.FormatNumberWithSystemLocale(totals.ukupno, 2),
			common.FormatNumberWithSystemLocale(totals.iznos14, 2),
			common.FormatNumberWithSystemLocale(totals.iznos15, 2),
			common.FormatNumberWithSystemLocale(totals.uvoz21, 2),
			common.FormatNumberWithSystemLocale(totals.uvoz22, 2),
			common.FormatNumberWithSystemLocale(totals.polj25, 2),
			common.FormatNumberWithSystemLocale(totals.polj24, 2),
		}
	}

	var grandTotals kprPrintTotals

	if entities != nil {
		var (
			currentGroup string
			groupTotals  kprPrintTotals
			groupOpen    bool
		)

		flushGroup := func() {
			if !groupOpen {
				return
			}
			if groupByNalog {
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					Fields:   formatTotalRow(fmt.Sprintf("Ukupno za nalog: %s", currentGroup), groupTotals),
					ClassRow: "nalog-total",
				})
			}
			groupTotals = kprPrintTotals{}
			groupOpen = false
		}

		for _, entity := range *entities {
			if groupByNalog {
				groupKey := fmt.Sprintf("%s-%v", entity.TipDok, entity.Nalog)
				if !groupOpen || groupKey != currentGroup {
					flushGroup()
					currentGroup = groupKey
					groupOpen = true
					if !summaryByNalog {
						tbl.Rows = append(tbl.Rows, domain.TableRow{
							Fields:   []string{fmt.Sprintf("Nalog: %s", currentGroup)},
							ClassRow: "nalog-header",
						})
					}
				}
				addToTotals(&groupTotals, entity)
				if !summaryByNalog {
					tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: formatDetailRow(entity)})
				}
			} else {
				tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: formatDetailRow(entity)})
			}
			addToTotals(&grandTotals, entity)
		}
		flushGroup()
	}

	tbl.Headers = s.GetKprStampaTableFields()
	tbl.HasTotals = true
	tbl.Totals = formatTotalRow("", grandTotals)
	tbl.Totals[6] = "Ukupno:"
	return nil
}

// GetKnjigaUlaznih retrieves data for Knjiga ulaznih racuna
func (s *PoreskeKnjigeResource) GetKnjigaPrimljenihRacuna(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.KnjigeParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.kprRepo.GetHasGodHasKar()
	odKnjige := ""
	doKnjige := ""

	tbl.SearchEnabled = true

	// Get parameters from query
	vKnjige := params.Knjiga
	if vKnjige == "999" {
		odKnjige = "1"
		doKnjige = "999"
	} else {
		odKnjige = params.Knjiga
		doKnjige = params.Knjiga
	}

	odDatuma := params.OdDatuma
	doDatuma := params.DoDatuma
	searchText := params.SearchText
	// Build query for Knjiga ulaznih racuna - similar to GetKnjigaIzdatih
	// but querying incoming invoices from suppliers
	qb := common.NewQueryBuilder(`
		SELECT 
			kpr.drbr, kpr.dknjiz, kpr.dizd, kpr.duvoz,
			kpr.dokum, kpr.iznsapdv,
			kpr.iznoslob, kpr.nisuobvpdv, kpr.uvozbezpdv, kpr.prethpdv,
			kpr.pretpdv1, kpr.pretpdv2, kpr.uvozpdv, kpr.poljvred, kpr.idkpr,
			coalesce(partneri.naziv, '') AS naziv_pa,
			coalesce(partneri.mesto, '') AS mesto_pa,
			coalesce(partneri.pib, '') AS pib,
			coalesce(partneri.adresa, '') AS adresa_pa
		FROM kpr`, true)
	qb.AddJoin(`LEFT JOIN partneri ON partneri.idpartneri = kpr.idpartneri`)
	if hasGod {
		qb.AddEqual("kpr.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kpr.kar", session.SelectedKar)
	}
	if odKnjige == doKnjige {
		qb.AddCondition("kpr.vkrbr", common.StringToInt(odKnjige), "=")
	} else {
		qb.AddCondition("kpr.vkrbr", common.StringToInt(odKnjige), ">=")
		qb.AddCondition("kpr.vkrbr", common.StringToInt(doKnjige), "<=")
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
	entities, err := s.kprRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
func (s *PoreskeKnjigeResource) KprVaidate(ctx context.Context, kpr *domain.Kpr, cAction string) ([]domain.FieldError, error) {

	return nil, nil
}
func (s *PoreskeKnjigeResource) SaveKnjigaPrimljenihRacuna(ctx context.Context, kpr *domain.Kpr, cAction string) error {

	return nil
}

// GetPoreskaPrijava retrieves data for Poreska prijava
func (s *PoreskeKnjigeResource) GetPoreskaPrijava(ctx context.Context, poreskaPrijavaData *domain.PoreskaPrijavaData) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	odDatuma := ctx.Value("oddatuma").(string)
	doDatuma := ctx.Value("dodatuma").(string)

	hasGod, hasKar := s.kirRepo.GetHasGodHasKar()
	// Build query for KIR records - fetch all records, not aggregates
	kirQB := common.NewQueryBuilder(`
		SELECT 	kir.dknjiz, kir.vkrbr, kir.izvozsapr, kir.oslobcl24,
			kir.izvozbezpr, kir.oslobcl25, kir.osn1, kir.osn2,
			kir.pdv1, kir.pdv2, kir.idkir
		FROM kir`, true)

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
	kirEntities, err := s.kirRepo.GetAllCustom(ctx, sqlKir, "", argsKir, "", "")
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
		FROM kpr`, true)
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
	kprEntities, err := s.kprRepo.GetAllCustom(ctx, sqlKpr, "", argsKpr, "", "")
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
		{Name: "krbr", Label: "Redni broj", Width: "8", Field: "kir.krbr", SkipInSearch: false, TextAlign: "right"},
		{Name: "dknjiz", Label: "Datum knjiženja", Width: "10", Field: "kir.dknjiz", SkipInSearch: false},
		{Name: "dizd", Label: "Datum izdavanja", Width: "10", Field: "kir.dizd", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "kir.dokum", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv kupca", Width: "25", Field: "partneri.naziv", SkipInSearch: false},
		{Name: "pib", Label: "PIB kupca", Width: "12", Field: "partneri.pib", SkipInSearch: false},
		{Name: "iznsapdv", Label: "Iznos sa PDV", Width: "12", Field: "kir.iznsapdv", SkipInSearch: false, TextAlign: "right"},
		{Name: "oslobcl24", Label: "Oslobođen CL 24", Width: "12", Field: "kir.oslobcl24", SkipInSearch: false, TextAlign: "right"},
		{Name: "oslobcl25", Label: "Oslobođen CL 25", Width: "12", Field: "kir.oslobcl25", SkipInSearch: false, TextAlign: "right"},
		{Name: "izvozsapr", Label: "Izvoz sa pravom", Width: "12", Field: "kir.izvozsapr", SkipInSearch: true, TextAlign: "right"},
		{Name: "izvozbezpr", Label: "Izvoz bez prava", Width: "12", Field: "kir.izvozbezpr", SkipInSearch: true, TextAlign: "right"},
		{Name: "osn1", Label: "Osnova 1", Width: "12", Field: "kir.osn1", SkipInSearch: true, TextAlign: "right"},
		{Name: "pdv1", Label: "PDV 1", Width: "12", Field: "kir.pdv1", SkipInSearch: true, TextAlign: "right"},
		{Name: "osn2", Label: "Osnova 2", Width: "12", Field: "kir.osn2", SkipInSearch: true, TextAlign: "right"},
		{Name: "pdv2", Label: "PDV 2", Width: "12", Field: "kir.pdv2", SkipInSearch: true, TextAlign: "right"},
		{Name: "idkir", Label: "ID", Width: "8", Field: "kir.idkir", SkipInSearch: true},
	}

	// Fields for Knjiga primljenih racuna (KPR - Received Invoices)
	s.kprTableFields = []domain.Fields{
		{Name: "drbr", Label: "Redni broj", Width: "8", Field: "kpr.drbr", SkipInSearch: false, TextAlign: "right"},
		{Name: "dknjiz", Label: "Datum knjiženja", Width: "10", Field: "kpr.dknjiz", SkipInSearch: false},
		{Name: "dizd", Label: "Datum izdavanja", Width: "10", Field: "kpr.dizd", SkipInSearch: false},
		{Name: "duvoz", Label: "Datum uvoza", Width: "10", Field: "kpr.duvoz", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "kpr.dokum", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv dobavljača", Width: "25", Field: "partneri.naziv", SkipInSearch: false},
		{Name: "pib", Label: "PIB dobavljača", Width: "12", Field: "partneri.pib", SkipInSearch: false},
		{Name: "iznsapdv", Label: "Iznos sa PDV", Width: "12", Field: "kpr.iznsapdv", SkipInSearch: true, TextAlign: "right"},
		{Name: "iznoslob", Label: "Iznos oslobođen", Width: "12", Field: "kpr.iznoslob", SkipInSearch: true, TextAlign: "right"},
		{Name: "nisuobvpdv", Label: "Nisu obavezni PDV", Width: "12", Field: "kpr.nisuobvpdv", SkipInSearch: true, TextAlign: "right"},
		{Name: "uvozbezpdv", Label: "Uvoz bez PDV", Width: "12", Field: "kpr.uvozbezpdv", SkipInSearch: true, TextAlign: "right"},
		{Name: "prethpdv", Label: "Prethodnog PDV", Width: "12", Field: "kpr.prethpdv", SkipInSearch: true, TextAlign: "right"},
		{Name: "pretpdv1", Label: "Prethodni PDV 1", Width: "12", Field: "kpr.pretpdv1", SkipInSearch: true, TextAlign: "right"},
		{Name: "pretpdv2", Label: "Prethodni PDV 2", Width: "12", Field: "kpr.pretpdv2", SkipInSearch: true, TextAlign: "right"},
		{Name: "uvozpdv", Label: "Uvoz PDV", Width: "12", Field: "kpr.uvozpdv", SkipInSearch: true, TextAlign: "right"},
		{Name: "poljvred", Label: "Polje vrednost", Width: "12", Field: "kpr.poljvred", SkipInSearch: true, TextAlign: "right"},
		{Name: "idkpr", Label: "ID", Width: "8", Field: "kpr.idkpr", SkipInSearch: true},
	}

	// Fields for Knjiga izdatih racuna - print (stamp) layout (17 columns, no ID, computed cols 16-17)
	s.kirStampaTableFields = []domain.Fields{
		{Name: "krbr", Label: "Red. br.", Width: "4", TextAlign: "right"},
		{Name: "dknjiz", Label: "Datum knjiženja", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "left"},
		{Name: "dizd", Label: "Datum izd.", Width: "8", TextAlign: "center"},
		{Name: "naziv", Label: "Naziv kupca", Width: "18", TextAlign: "left"},
		{Name: "pib", Label: "PIB", Width: "10", TextAlign: "left"},
		{Name: "iznsapdv", Label: "Ukupna naknada sa PDV", Width: "10", TextAlign: "right"},
		{Name: "oslobcl24", Label: "Oslobođen cl.24", Width: "10", TextAlign: "right"},
		{Name: "oslobcl25", Label: "Oslobođen cl.25", Width: "10", TextAlign: "right"},
		{Name: "izvozsapr", Label: "Izvoz sa pravom", Width: "10", TextAlign: "right"},
		{Name: "izvozbezpr", Label: "Izvoz bez prava", Width: "10", TextAlign: "right"},
		{Name: "osn1", Label: "Osnov (opšta)", Width: "10", TextAlign: "right"},
		{Name: "pdv1", Label: "PDV (opšta)", Width: "10", TextAlign: "right"},
		{Name: "osn2", Label: "Osnov (posebna)", Width: "10", TextAlign: "right"},
		{Name: "pdv2", Label: "PDV (posebna)", Width: "10", TextAlign: "right"},
		{Name: "ukpromet", Label: "Uk. promet bez PDV", Width: "10", TextAlign: "right"},
		{Name: "prometspravom", Label: "Promet sa pravom odb.", Width: "10", TextAlign: "right"},
	}

	// Fields for Knjiga primljenih racuna - print (19 columns)
	s.kprStampaTableFields = []domain.Fields{
		{Name: "drbr", Label: "Red. br.", Width: "4", TextAlign: "right"},
		{Name: "dknjiz", Label: "Knjiženja isprave", Width: "8", TextAlign: "center"},
		{Name: "duvoz", Label: "Prij. carin. isprave", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj računa", Width: "8", TextAlign: "left"},
		{Name: "dizd", Label: "Datum izdavanja", Width: "8", TextAlign: "center"},
		{Name: "naziv", Label: "Naziv dobavljača", Width: "18", TextAlign: "left"},
		{Name: "pib", Label: "PIB ili JMBG", Width: "10", TextAlign: "left"},
		{Name: "iznsapdv", Label: "Ukupna naknada sa PDV", Width: "10", TextAlign: "right"},
		{Name: "naknadapdv", Label: "Naknada bez PDV", Width: "10", TextAlign: "right"},
		{Name: "oslobodj", Label: "Oslobođ. nabavke", Width: "10", TextAlign: "right"},
		{Name: "nabavbez", Label: "Nabavka bez PDV", Width: "10", TextAlign: "right"},
		{Name: "uvozbez", Label: "Naknada uvoza bez PDV", Width: "10", TextAlign: "right"},
		{Name: "ukupno", Label: "Ukupan iznos", Width: "10", TextAlign: "right"},
		{Name: "iznos14", Label: "Iznos preth. PDV", Width: "10", TextAlign: "right"},
		{Name: "iznos15", Label: "Iznos preth. PDV ne može odbiti", Width: "10", TextAlign: "right"},
		{Name: "uvoz21", Label: "Vrednost dobara bez PDV", Width: "10", TextAlign: "right"},
		{Name: "uvoz22", Label: "Iznos PDV", Width: "10", TextAlign: "right"},
		{Name: "polj25", Label: "Vrednost primlj. dobara", Width: "10", TextAlign: "right"},
		{Name: "polj24", Label: "Iznos naknade", Width: "10", TextAlign: "right"},
	}
}

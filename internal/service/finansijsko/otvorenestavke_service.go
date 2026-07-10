package finansijsko

import (
	"context"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/pkg/utils"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

// OtvoreneStavkeService defines the service interface for Otvorene Stavke operations
type OtvoreneStavkeService interface {
	GetOtvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetOtvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string, tipStampe string, tipObrade string, repParams *domain.ReportParameters) error
	GetZatvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetZatvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string, tipStampe string, repParams *domain.ReportParameters) error
	GetIOSPartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetIOSDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string, tipStampe string, repParams *domain.ReportParameters) error
	GetIOSStampaFields() []domain.Fields
	GetDospelaPotrazivanjaPartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetDospelaPotrazivanjaDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetPregledPotrazivanjaObaveze(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetPregledDugovanjaPoStarosti(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetPovezivanjeRacunaUplata(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPartneriFields() []domain.Fields
	GetPartneriFieldsSintetika() []domain.Fields
	GetOtvoreneStavkeDetaljiFields() []domain.Fields
	GetZatvoreneStavkeDetaljiFields() []domain.Fields
	GetIOSDetaljiFields() []domain.Fields
	GetDospelaDetaljiFields() []domain.Fields
	GetDugovanjaFields() []domain.Fields
	GetPovezivanjePartneriFields() []domain.Fields
	GetPovezivanjeUplateFields() []domain.Fields
	GetPovezivanjeFaktureFields() []domain.Fields
	GetOpomenaStampaFields() []domain.Fields
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetZatvoreneStavkeStampaFields() []domain.Fields
	GetDospelaPotrazivanjaStampa(ctx context.Context, params domain.OtvStavkeParam) ([]domain.DospelaPartnerReport, error)
	GetPregledPotrazivanjaObavezeStampa(ctx context.Context, params domain.OtvStavkeParam) ([]domain.PregledPotrazivanjaPartnerReport, error)
	GetPotrazivanjaDugovanjaStampaFields() []domain.Fields
	GetPregledDugovanjaPoStarostiStampa(ctx context.Context, tbl *domain.TableData, params domain.OtvStavkeParam) error
	GetDugovanjaStarostiStampaFields() []domain.Fields
}

// OtvoreneStavkeResource implements OtvoreneStavkeService
type OtvoreneStavkeResource struct {
	fproRepo                            repository.BaseRepository[domain.Fpro]
	fvrRepo                             *repository.BaseRepository[domain.Fvr]
	partneriRepo                        *repository.BaseRepository[domain.Partneri]
	fieldsPartneri                      []domain.Fields
	fieldsPartneriSintetika             []domain.Fields
	fieldsOtvoreneStavkeDetalji         []domain.Fields
	fieldsZatvoreneStavke               []domain.Fields
	fieldsDospelaPotrazivanja           []domain.Fields
	fieldsZatvoreneStavkeDetalji        []domain.Fields
	fieldsIOSDetalji                    []domain.Fields
	fieldsDospelaDetalji                []domain.Fields
	fieldsPotrazivanjaObavezePartneri   []domain.Fields
	fieldsPotrazivanjaObaveze           []domain.Fields
	fieldsPovezivanjePartneri           []domain.Fields
	fieldsPovezivanjeUplate             []domain.Fields
	fieldsPovezivanjeFakture            []domain.Fields
	fieldsZatvoreneStavkeStampa         []domain.Fields
	fieldsOpomenaStampa                 []domain.Fields
	fieldsIOSStampa                     []domain.Fields
	fieldsDospelaStampa                 []domain.Fields
	fieldsPotrazivanjaDugovanjaStampa   []domain.Fields
	fieldsDugovanjaStarostiStampaFields []domain.Fields
}

// NewOtvoreneStavkeService creates a new service instance
func NewOtvoreneStavkeService(repo repository.BaseRepository[domain.Fpro], fvrRepo *repository.BaseRepository[domain.Fvr], partneriRepo *repository.BaseRepository[domain.Partneri]) *OtvoreneStavkeResource {
	service := &OtvoreneStavkeResource{
		fproRepo:     repo,
		fvrRepo:      fvrRepo,
		partneriRepo: partneriRepo,
	}
	service.setServiceFieldValues()
	return service
}

// GetOtvoreneStavkePartneri retrieves open items data for partners
func (s *OtvoreneStavkeResource) GetOtvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike, 
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`, true)
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")

	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddEqual("fpro.konto", params.Konto)
	qb.AddCondition("fpro.sifra", params.OdSifre, ">=")
	qb.AddCondition("fpro.sifra", params.DoSifre, "<=")
	qb.AddCondition("fpro.dadok + fpro.rok", params.PodDatumom, "<=")
	//qb.AddCondition("fpro.otvstavkedana", otvstavkedana, "<=")
	// Add search conditions if search text is provided
	if params.SearchText != "" && params.SearchText != "undefined" && len(params.SearchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), params.SearchText)
	}
	// Add GROUP BY
	qb.AddGroupBy("fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv")
	// Add HAVING clause to filter only partners with both debtor and creditor accounts
	qb.AddHaving(` 
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.sifra")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		tbl.Totals = make([]string, len(tbl.Headers))
		var totalDug, totalPot float64
		for _, entity := range *entities {
			totalDug += entity.Dug.Float64
			totalPot += entity.Pot.Float64
		}
		if len(tbl.Headers) >= 6 {
			tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
			tbl.Totals[3] = common.FormatNumberWithSystemLocale(totalDug, 2)
			tbl.Totals[4] = common.FormatNumberWithSystemLocale(totalPot, 2)
			tbl.Totals[5] = common.FormatNumberWithSystemLocale(totalDug-totalPot, 2)
		}
		return nil
	}
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Konto,
				entity.Sifra,
				entity.NazivAnalitike,
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetOtvoreneStavkeDetalji retrieves open items details for partners
func (s *OtvoreneStavkeResource) GetOtvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string, tipStampe string, tipObrade string, repParams *domain.ReportParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F' 
								  			  when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
								coalesce(fpro.dokum, '') as dokum, fpro.tra, coalesce(fpro.ojozn, '') as ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								coalesce(fpro.opis, '') as opis, fpro.vrd,
								fpro.konto, coalesce(fpro.sifra, '') as sifra
								FROM fpro`, true)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddCustomCondition(` EXISTS (
				SELECT 1
				FROM fkpl
				JOIN partneri p
					ON fkpl.tipanalitikeid = p.tipanalitikeid
				WHERE fkpl.idfkpl = fpro.idfkpl
					AND p.sifra = fpro.sifra
					AND fkpl.idfkpl = $3
			)`)
	qb.AddArgs(id)
	// Add search conditions if search text is provided
	if searchText != "" && searchText != "undefined" && len(searchText) >= 1 && tipStampe != common.TipStampePrint {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetOtvoreneStavkeDetaljiFields(), searchText)
	}
	if !getTotalRecords && tipStampe != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.dokum, fpro.dadok")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords && tipStampe != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	if tipStampe == common.TipStampePrint {
		translator := i18n.GetInstance()
		partner, err := s.getPartneriData(ctx, id)
		if err != nil {
			return err
		}
		repParams.ParameterItems["Naziv"] = domain.ParameterItem{Name: translator.Label("Naziv"), Value: partner.Naziv}
		repParams.ParameterItems["Adresa"] = domain.ParameterItem{Name: translator.Label("Adresa"), Value: partner.Adresa}
		repParams.ParameterItems["Postcode"] = domain.ParameterItem{Name: translator.Label("Postcode"), Value: fmt.Sprintf("%d", partner.PoBro)}
		repParams.ParameterItems["Mesto"] = domain.ParameterItem{Name: translator.Label("Mesto"), Value: partner.Mesto}
		repParams.ParameterItems["ObvPDV"] = domain.ParameterItem{Name: translator.Label("Obveznik PDV Br."), Value: partner.MatBr}
		repParams.ParameterItems["PIB"] = domain.ParameterItem{Name: translator.Label("PIB"), Value: partner.PIB}
		repParams.ParameterItems["Telefon"] = domain.ParameterItem{Name: translator.Label("Telefon"), Value: partner.Telefon}
		repParams.ParameterItems["Sifra"] = domain.ParameterItem{Name: translator.Label("Sifra"), Value: partner.Sifra}
	}
	// Populate table rows
	rbr := 0
	totDug, totPot := 0.0, 0.0
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			rbr++
			fields := []string{}
			// Add common fields
			if tipStampe == common.TipStampePrint {
				repParams.ParameterItems["Konto"] = domain.ParameterItem{Name: i18n.GetInstance().Label("Konto"), Value: entity.Konto}
				if tipObrade == "opomena" && entity.DocType == "F" {
					fields = append(fields,
						fmt.Sprintf("%d", rbr),
						entity.Opis,
						entity.Ojozn.String,
						entity.Dokum,
						common.FormatNullTime(entity.Dadok, common.DateLayout),
						fmt.Sprintf("%d", entity.Tra),
						fmt.Sprintf("%d", entity.Rok),
						common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
						common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
						common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
						common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
					)
					totDug += entity.Dug.Float64
					totPot += entity.Pot.Float64
				}
				if tipObrade == "otvstavke" {
					fields = append(fields,
						fmt.Sprintf("%d", rbr),
						entity.Opis,
						entity.Ojozn.String,
						entity.Dokum,
						common.FormatNullTime(entity.Dadok, common.DateLayout),
						fmt.Sprintf("%d", entity.Tra),
						fmt.Sprintf("%d", entity.Rok),
						common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
						common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
						common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
						common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
					)
					totDug += entity.Dug.Float64
					totPot += entity.Pot.Float64
				}
			} else {
				fields = append(fields,
					"+",
					entity.DocType,
					entity.Dokum,
					fmt.Sprintf("%d", entity.Tra),
					entity.Ojozn.String,
					fmt.Sprintf("%d", entity.Nalog),
					entity.Danal.Format(common.DateLayout),
					common.FormatNullTime(entity.Dadok, common.DateLayout),
					fmt.Sprintf("%d", entity.Rok),
					common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
					entity.Opis)
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	tbl.Totals[8] = common.FormatNumberWithSystemLocale(totDug, 2)
	tbl.Totals[9] = common.FormatNumberWithSystemLocale(totPot, 2)
	tbl.Totals[10] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
	return nil
}

// GetZatvoreneStavkePartneri retrieves closed items data for partners
func (s *OtvoreneStavkeResource) GetZatvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike, 
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`, true)
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")

	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddIn("fpro.vrd", []interface{}{10, 20, 30, 40})
	qb.AddEqual("fpro.vkonta", "1")
	qb.AddEqual("fpro.konto", params.Konto)
	qb.AddCondition("fpro.sifra", params.OdSifre, ">=")
	qb.AddCondition("fpro.sifra", params.DoSifre, "<=")
	qb.AddCondition("fpro.dadok", params.OdDatuma, ">=")
	qb.AddCondition("fpro.dadok", params.DoDatuma, "<=")
	// Add search conditions if search text is provided
	if params.SearchText != "" && params.SearchText != "undefined" && len(params.SearchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), params.SearchText)
	}
	// Add GROUP BY
	qb.AddGroupBy("fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv")
	// Add HAVING clause to filter only partners with both debtor and creditor accounts
	qb.AddHaving(` 
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) = 0`)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.sifra")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Konto,
				entity.Sifra,
				entity.NazivAnalitike,
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetZatvoreneStavkeDetalji retrieves closed items details for partners
func (s *OtvoreneStavkeResource) GetZatvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string, tipStampe string, repParams *domain.ReportParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT
								fpro.rbr, fpro.konto, coalesce(fpro.sifra, '') as sifra, fpro.dokum, fpro.tra, fpro.ojozn, fpro.tipdok, fpro.nalog, fpro.danal, 
								fpro.dadok, fpro.rok, 
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								fpro.opis, fpro.vrd, fpro.vkonta, fpro.kat
								FROM fpro`, true)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddCustomCondition(` EXISTS (
			SELECT 1
				FROM fkpl
				JOIN partneri p
					ON fkpl.tipanalitikeid = p.tipanalitikeid
				WHERE fkpl.idfkpl = fpro.idfkpl
					AND p.sifra = fpro.sifra
					AND fkpl.idfkpl = $3
			)`)
	qb.AddArgs(id)
	// Add search conditions if search text is provided
	if searchText != "" && searchText != "undefined" && len(searchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), searchText)
	}
	if !getTotalRecords && tipStampe != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.dokum, fpro.dadok")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	if tipStampe == common.TipStampePrint {
		partner, err := s.getPartneriData(ctx, id)
		if err != nil {
			return err
		}
		konto := ""
		if entities != nil && len(*entities) > 0 {
			konto = (*entities)[0].Konto
			repParams.ParameterItems["Konto"] = domain.ParameterItem{Name: i18n.GetInstance().Label("Konto"), Value: konto}
		}
		repParams.ParameterItems["Partner"] = domain.ParameterItem{Name: "Partner", Value: fmt.Sprintf("%s %s - %s, %s %d  %s", konto, partner.Sifra, partner.Naziv, partner.Adresa, partner.PoBro, partner.Mesto)}
	}
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		totDug, totPot := 0.0, 0.0
		rbr := 0
		for _, entity := range *entities {
			totDug += entity.Dug.Float64
			totPot += entity.Pot.Float64
			rbr++
			fields := []string{}
			// Add common fields
			if tipStampe == common.TipStampePrint {
				fields = append(fields,
					fmt.Sprintf("%d", rbr),
					entity.Tipdok,
					fmt.Sprintf("%d", entity.Nalog),
					entity.Danal.Format(common.DateLayout),
					entity.Dokum,
					common.FormatNullTime(entity.Dadok, common.DateLayout),
					fmt.Sprintf("%d", entity.Rok),
					common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				)
			} else {
				fields = append(fields,
					entity.Dokum,
					fmt.Sprintf("%d", entity.Tra),
					entity.Ojozn.String,
					entity.Tipdok,
					fmt.Sprintf("%d", entity.Nalog),
					entity.Danal.Format(common.DateLayout),
					common.FormatNullTime(entity.Dadok, common.DateLayout),
					fmt.Sprintf("%d", entity.Rok),
					common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
					entity.Opis,
					fmt.Sprintf("%d", entity.Vrd),
					fmt.Sprintf("%d", entity.Vkonta),
					fmt.Sprintf("%d", entity.Kat),
				)
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
		if tipStampe == common.TipStampePrint {
			tbl.Totals = make([]string, len(tbl.Headers))
			tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

			for i, header := range tbl.Headers {
				if header.IncludeInTotals {
					switch header.Field {
					case "duguje":
						tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
					case "potrazuje":
						tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
					}
				}
			}
		}
	}

	return nil
}
func (s *OtvoreneStavkeResource) getPartneriData(ctx context.Context, id int64) (partner domain.Partneri, err error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return partner, fmt.Errorf("user session not found")
	}
	// query for getting partner data for report parameters
	partnerQuery := common.NewQueryBuilder(`SELECT coalesce(p.naziv, '') as naziv,
								coalesce(p.sifra, '') as sifra,
								coalesce(p.adresa, '') as adresa,
								coalesce(p.mesto, '') as mesto,
								coalesce(p.pib, '') as pib,
								coalesce(p.matbr, '') as matbr,
								coalesce(p.pobro, 0) as pobro,
								coalesce(p.telefon, '') as telefon
								FROM partneri p`, true)
	hasGod, hasKar := s.partneriRepo.GetHasGodHasKar()
	if hasGod {
		partnerQuery.AddEqual("p.god", userSession.SelectedGod)
	}
	if hasKar {
		partnerQuery.AddEqual("p.kar", userSession.SelectedKar)
	}
	partnerQuery.AddCustomCondition(` EXISTS (
				SELECT 1
				FROM fkpl
				JOIN fpro
					ON fpro.idfkpl = fkpl.idfkpl
				WHERE fkpl.tipanalitikeid = p.tipanalitikeid
					AND p.sifra = fpro.sifra
					AND fkpl.idfkpl = $3
			)`)
	partnerQuery.AddArgs(id)
	partnerSQL, partnerArgs := partnerQuery.Build()
	partnerEntity, err := s.partneriRepo.GetAllCustom(ctx, partnerSQL, "", partnerArgs, "", "")
	if err != nil {
		return partner, err
	}
	entity := domain.Partneri{}
	if len(*partnerEntity) > 0 {
		entity = (*partnerEntity)[0]
	}
	return entity, nil
}

// GetIOS retrieves IOS (Izvod otvorenih stavki) data
func (s *OtvoreneStavkeResource) GetIOSPartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike, 
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`, true)
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")

	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddEqual("fpro.konto", params.Konto)
	qb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.OdSifre, ">=")
	qb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.DoSifre, "<=")
	qb.AddCondition("fpro.dadok", params.PodDatumom, "<=")
	//qb.AddCondition("fpro.otvstavkedana", params.otvstavkedana, "<=")
	// Add search conditions if search text is provided
	if params.SearchText != "" && params.SearchText != "undefined" && len(params.SearchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), params.SearchText)
	}
	// Add GROUP BY
	qb.AddGroupBy("fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv")
	// Add HAVING clause to filter only partners with both debtor and creditor accounts
	qb.AddHaving(` 
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Konto,
				entity.Sifra,
				entity.NazivAnalitike,
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetIOSDetalji retrieves IOS details for partners
func (s *OtvoreneStavkeResource) GetIOSDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string, tipStampe string, repParams *domain.ReportParameters) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F' 
								  			  when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
								fpro.dokum, fpro.tra, fpro.ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								coalesce(fpro.opis, '') as opis,
								fpro.vrd
								FROM fpro`, true)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddEqual("fpro.idfkpl", id)
	// Add search conditions if search text is provided
	if searchText != "" && searchText != "undefined" && len(searchText) >= 1 && tipStampe != common.TipStampePrint {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetOtvoreneStavkeDetaljiFields(), searchText)
	}
	if !getTotalRecords && tipStampe != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.dokum, fpro.dadok")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	if tipStampe == common.TipStampePrint {
		partner, err := s.getPartneriData(ctx, id)
		if err != nil {
			return err
		}
		translator := i18n.GetInstance()
		repParams.ParameterItems["Naziv"] = domain.ParameterItem{Name: translator.Label("Naziv"), Value: partner.Naziv}
		repParams.ParameterItems["Adresa"] = domain.ParameterItem{Name: translator.Label("Adresa"), Value: partner.Adresa}
		repParams.ParameterItems["Postcode"] = domain.ParameterItem{Name: translator.Label("Postcode"), Value: fmt.Sprintf("%d", partner.PoBro)}
		repParams.ParameterItems["Mesto"] = domain.ParameterItem{Name: translator.Label("Mesto"), Value: partner.Mesto}
		repParams.ParameterItems["PIB"] = domain.ParameterItem{Name: translator.Label("PIB"), Value: partner.PIB}
		repParams.ParameterItems["Telefon"] = domain.ParameterItem{Name: translator.Label("Telefon"), Value: partner.Telefon}
		repParams.ParameterItems["Sifra"] = domain.ParameterItem{Name: translator.Label("Sifra"), Value: partner.Sifra}
	}
	// Populate table rows
	totDug, totPot := 0.0, 0.0
	rbr := 0
	tbl.Totals = make([]string, len(tbl.Headers))
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			rbr++
			fields := []string{}
			if tipStampe == common.TipStampePrint {
				totDug += entity.Dug.Float64
				totPot += entity.Pot.Float64
				fields = append(fields,
					fmt.Sprintf("%d", rbr),
					entity.Opis,
					entity.Ojozn.String,
					entity.Dokum,
					common.FormatNullTime(entity.Dadok, common.DateLayout),
					fmt.Sprintf("%d", entity.Rok),
					common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
				)
			} else {
				fields = append(fields,
					entity.DocType,
					entity.Dokum,
					fmt.Sprintf("%d", entity.Tra),
					entity.Ojozn.String,
					fmt.Sprintf("%d", entity.Nalog),
					entity.Danal.Format(common.DateLayout),
					common.FormatNullTime(entity.Dadok, common.DateLayout),
					fmt.Sprintf("%d", entity.Rok),
					common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
					entity.Opis,
					fmt.Sprintf("%d", entity.Vrd))
			}
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
		if tipStampe == common.TipStampePrint {
			tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
			for i, header := range tbl.Headers {
				if header.IncludeInTotals {
					switch header.Field {
					case "duguje":
						tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
					case "potrazuje":
						tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
					case "iznos":
						tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
					}
				}
			}
		}
	}

	return nil
}

// GetIOSStampaFields returns fields for the IOS print report
func (s *OtvoreneStavkeResource) GetIOSStampaFields() []domain.Fields {
	return s.fieldsIOSStampa
}

// GetDospelaPotrazivanjaPartneri retrieves due receivables data for partners
func (s *OtvoreneStavkeResource) GetDospelaPotrazivanjaPartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	brojDanaInt, err := strconv.Atoi(params.BrojDana)
	if err != nil {
		return fmt.Errorf("invalid BrojDana value: %v", err)
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike, 
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`, true)
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")

	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddEqual("fpro.konto", params.Konto)
	qb.AddCondition("fpro.sifra::numeric", params.OdSifre, ">=")
	qb.AddCondition("fpro.sifra::numeric", params.DoSifre, "<=")
	qb.AddCondition("fpro.dadok + fpro.rok", params.PodDatumom, "<=")
	if params.TipPotrazivanja == "D" {
		qb.AddCondition("fpro.dadok + fpro.rok", time.Now().AddDate(0, 0, brojDanaInt).Format(common.HtmlLayout), "<=")
	}
	if params.TipPregleda == "P" {
		qb.AddCondition("fpro.dadok + fpro.rok", time.Now().AddDate(0, 0, -brojDanaInt).Format(common.HtmlLayout), ">=")
	}

	// Add search conditions if search text is provided
	if params.SearchText != "" && params.SearchText != "undefined" && len(params.SearchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), params.SearchText)
	}
	// Add GROUP BY
	qb.AddGroupBy("fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv")
	// Add HAVING clause to filter only partners with both debtor and creditor accounts
	qb.AddHaving(` 
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.sifra::numeric")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fields := []string{}
			// Add common fields
			if params.TipPregleda == "A" {
				fields = append(fields,
					entity.Konto,
					entity.Sifra,
					entity.NazivAnalitike,
					common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
					common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
				)
			}
			if params.TipPregleda == "S" {
				fields = append(fields,
					entity.Konto,
					entity.Sifra,
					entity.NazivAnalitike,
				)
			}

			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetDospelaPotrazivanjaDetalji retrieves overdue receivables details for partners
func (s *OtvoreneStavkeResource) GetDospelaPotrazivanjaDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F' 
								  			  when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
								coalesce(fpro.dokum, '') as dokum, fpro.tra, fpro.ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								coalesce(fpro.opis, '') as opis, fpro.vrd
								FROM fpro`, true)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddEqual("fpro.idfkpl", id)
	// Add search conditions if search text is provided
	if searchText != "" && searchText != "undefined" && len(searchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetOtvoreneStavkeDetaljiFields(), searchText)
	}
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.dokum, fpro.dadok")
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.DocType,
				entity.Dokum,
				fmt.Sprintf("%d", entity.Rok),
				common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Pot.Float64, 2),
				common.FormatNumberWithSystemLocale(entity.Dug.Float64-entity.Pot.Float64, 2),
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				entity.Opis)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetDospelaPotrazivanjaStampa retrieves all dospela potrazivanja grouped by partner for analytical print report.
func (s *OtvoreneStavkeResource) GetDospelaPotrazivanjaStampa(ctx context.Context, params domain.OtvStavkeParam) ([]domain.DospelaPartnerReport, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	brojDanaInt, _ := strconv.Atoi(params.BrojDana)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Fetch all partners with due items (no pagination) – same filter as GetDospelaPotrazivanjaPartneri
	qbP := common.NewQueryBuilder(`SELECT fkpl.idfkpl, fpro.konto, coalesce(fpro.sifra, '') as sifra, fkpl.naziv as nazivanalitike,
		SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`, true)
	qbP.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")
	if hasGod {
		qbP.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qbP.AddEqual("fpro.kar", session.SelectedKar)
	}
	qbP.AddEqual("fpro.konto", params.Konto)
	qbP.AddCondition("fpro.sifra::numeric", params.OdSifre, ">=")
	qbP.AddCondition("fpro.sifra::numeric", params.DoSifre, "<=")
	qbP.AddCondition("fpro.dadok + fpro.rok", params.PodDatumom, "<=")
	if params.TipPotrazivanja == "D" {
		qbP.AddCondition("fpro.dadok + fpro.rok", time.Now().AddDate(0, 0, brojDanaInt).Format(common.HtmlLayout), "<=")
	}
	if params.TipPregleda == "P" {
		qbP.AddCondition("fpro.dadok + fpro.rok", time.Now().AddDate(0, 0, -brojDanaInt).Format(common.HtmlLayout), ">=")
	}
	qbP.AddGroupBy("fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv")
	qbP.AddHaving(`SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`)
	qbP.AddOrderBy("fpro.sifra::numeric")
	sqlP, argsP := qbP.Build()

	partneriRows, err := s.fproRepo.GetAllCustom(ctx, sqlP, "", argsP, "", "")
	if err != nil {
		return nil, err
	}

	var result []domain.DospelaPartnerReport
	for _, pr := range *partneriRows {
		report := domain.DospelaPartnerReport{
			Konto: pr.Konto,
			Sifra: pr.Sifra,
			Naziv: pr.NazivAnalitike,
		}

		// Fetch detail rows for this partner filtered by dospece <= PodDatumom
		qbD := common.NewQueryBuilder(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F'
			when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
			fpro.dokum, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
			case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
			case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
			fpro.vrd
			FROM fpro`, true)
		if hasGod {
			qbD.AddEqual("fpro.god", session.SelectedGod)
		}
		if hasKar {
			qbD.AddEqual("fpro.kar", session.SelectedKar)
		}
		qbD.AddEqual("fpro.idfkpl", pr.IDFkpl)
		qbD.AddCondition("fpro.dadok + fpro.rok", params.PodDatumom, "<=")
		qbD.AddOrderBy("fpro.nalog, fpro.dadok")
		sqlD, argsD := qbD.Build()

		detalji, err := s.fproRepo.GetAllCustom(ctx, sqlD, "", argsD, "", "")
		if err != nil {
			return nil, err
		}

		var totalDug, totalPot float64
		for i, d := range *detalji {
			dospece := common.AddDaysToNullTime(d.Dadok, d.Rok, common.DateLayout)
			item := domain.DospelaPartnerItem{
				RBr:        fmt.Sprintf("%d", i+1),
				Nalog:      fmt.Sprintf("%d", d.Nalog),
				DatNalog:   d.Danal.Format(common.DateLayout),
				TipDok:     d.DocType,
				BrDok:      d.Dokum,
				DatDok:     common.FormatNullTime(d.Dadok, common.DateLayout),
				Rok:        fmt.Sprintf("%d", d.Rok),
				Dospece:    dospece,
				DospeliDug: common.FormatNumberWithSystemLocale(d.Dug.Float64, 2),
				Placeno:    common.FormatNumberWithSystemLocale(d.Pot.Float64, 2),
				Saldo:      common.FormatNumberWithSystemLocale(d.Dug.Float64-d.Pot.Float64, 2),
			}
			totalDug += d.Dug.Float64
			totalPot += d.Pot.Float64
			report.Items = append(report.Items, item)
		}
		report.TotalDug = common.FormatNumberWithSystemLocale(totalDug, 2)
		report.TotalPot = common.FormatNumberWithSystemLocale(totalPot, 2)
		report.TotalSaldo = common.FormatNumberWithSystemLocale(totalDug-totalPot, 2)
		result = append(result, report)
	}
	return result, nil
}

// GetPregledPotrazivanjaObaveze retrieves due receivables and obligations data for partners
func (s *OtvoreneStavkeResource) GetPregledPotrazivanjaObaveze(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// Validate interval parameters - must be numeric only to prevent SQL injection
	intervals := map[string]string{
		"dospece15":  params.Dospece15,
		"dospece30":  params.Dospece30,
		"dospece60":  params.Dospece60,
		"dospece90":  params.Dospece90,
		"dospece120": params.Dospece120,
	}

	for name, value := range intervals {
		if !regexp.MustCompile(`^\d+$`).MatchString(value) {
			return fmt.Errorf("invalid %s parameter: must be numeric", name)
		}
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build age bucket intervals
	interval15 := params.Dospece15 + " days"
	interval30 := params.Dospece30 + " days"
	interval60 := params.Dospece60 + " days"
	interval90 := params.Dospece90 + " days"
	interval120 := params.Dospece120 + " days"

	// Determine vkonta based on tipPregleda (K=kupci, D=dobavljaci)
	vrd1 := ""
	vrd2 := ""
	if params.TipPregleda == "K" { //kupci
		vrd1 = "10"
		vrd2 = "30"
	}
	if params.TipPregleda == "D" { //dobavljaci
		vrd1 = "20"
		vrd2 = "40"
	}

	// Get data for the table with age bucket categorization
	qb := common.NewQueryBuilder(fmt.Sprintf(`
		SELECT
			fpro.konto, 
			coalesce(fpro.sifra, '') as sifra, 
			fkpl.naziv as nazivanalitike,
			-- Total realized amount (invoices only)
			SUM(CASE WHEN fpro.vrd IN ($1) THEN fpro.iznos ELSE 0 END) as realizacija,
			-- Total payments
			SUM(CASE WHEN fpro.vrd IN ($2) THEN fpro.iznos ELSE 0 END) as placeno,
			-- Already due (past state date)
			SUM(CASE WHEN fpro.vrd IN ($1) AND fpro.dadok IS NOT NULL AND (fpro.dadok + fpro.rok)::date <= '%s'::date
					THEN fpro.iznos ELSE 0 END) as dospelarealizacija,
			-- Due 0-N days from state date (using dospece15)
			SUM(CASE WHEN fpro.vrd IN ($1) 
					AND (fpro.dadok + fpro.rok)::date > '%s'::date 
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval15+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece15,
			-- Due (dospece15+1)-N days from state date (using dospece30)
			SUM(CASE WHEN fpro.vrd IN ($1) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval15+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval30+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece30,
			-- Due (dospece30+1)-N days from state date (using dospece60)
			SUM(CASE WHEN fpro.vrd IN ($1) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval30+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval60+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece60,
			-- Due (dospece60+1)-N days from state date (using dospece90)
			SUM(CASE WHEN fpro.vrd IN ($1) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval60+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval90+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece90,
			-- Due (dospece90+1)-N days from state date (using dospece120)
			SUM(CASE WHEN fpro.vrd IN ($1) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval90+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval120+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece120,
			-- Due more than dospece120 days from state date
			SUM(CASE WHEN fpro.vrd IN ($1) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval120+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece120plus
		FROM fpro
	`, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan,
		params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan,
		params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan), true)
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")

	qb.AddArgs(vrd1, vrd2)
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}

	qb.AddCondition("fpro.konto::numeric", params.OdKonta, ">=")
	qb.AddCondition("fpro.konto::numeric", params.DoKonta, "<=")
	qb.AddCondition("NULLIF(fpro.sifra, '')::numeric", params.OdSifre, ">=")
	qb.AddCondition("NULLIF(fpro.sifra, '')::numeric", params.DoSifre, "<=")
	qb.AddCondition("fpro.dadok", params.StanjeNaDan, "<=")
	qb.AddIn("fpro.vrd", []interface{}{vrd1, vrd2})

	// Add search conditions if search text is provided
	if params.SearchText != "" && params.SearchText != "undefined" && len(params.SearchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), params.SearchText)
	}

	// Add GROUP BY
	qb.AddGroupBy("fpro.konto, fpro.sifra, fkpl.naziv")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("NULLIF(fpro.sifra, '')::numeric")

	sqlQuery, args := qb.Build()

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fields := []string{}
			// Detailed view with age buckets
			fields = append(fields,
				entity.Konto,
				entity.Sifra,
				entity.NazivAnalitike,
				common.FormatNumberWithSystemLocale(entity.Realizacija, 2),
				common.FormatNumberWithSystemLocale(entity.Placeno, 2),
				common.FormatNumberWithSystemLocale(entity.Realizacija-entity.Placeno, 2),
				common.FormatNumberWithSystemLocale(entity.DospelaRealizacija-entity.Placeno, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece15, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece30, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece60, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece90, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece120, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece120Plus, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetPregledPotrazivanjaObavezeStampa retrieves all pregled potrazivanja/obaveze data grouped by partner for print report.
func (s *OtvoreneStavkeResource) GetPregledPotrazivanjaObavezeStampa(ctx context.Context, params domain.OtvStavkeParam) ([]domain.PregledPotrazivanjaPartnerReport, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	// Validate interval parameters - must be numeric only to prevent SQL injection
	for name, value := range map[string]string{
		"dospece15": params.Dospece15, "dospece30": params.Dospece30,
		"dospece60": params.Dospece60, "dospece90": params.Dospece90,
		"dospece120": params.Dospece120,
	} {
		if !regexp.MustCompile(`^\d+$`).MatchString(value) {
			return nil, fmt.Errorf("invalid %s parameter: must be numeric", name)
		}
	}

	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	sd := params.StanjeNaDan
	i15 := params.Dospece15 + " days"
	i30 := params.Dospece30 + " days"
	i60 := params.Dospece60 + " days"
	i90 := params.Dospece90 + " days"
	i120 := params.Dospece120 + " days"

	vrd1, vrd2 := "", ""
	switch params.TipPregleda {
	case "K":
		vrd1, vrd2 = "10", "30"
	case "D":
		vrd1, vrd2 = "20", "40"
	}

	qbP := common.NewQueryBuilder(fmt.Sprintf(`
		SELECT fpro.konto, coalesce(fpro.sifra, '') as sifra, fkpl.idfkpl,
			fkpl.naziv as nazivanalitike,
			COALESCE(p.mesto, '') as mesto,
			SUM(CASE WHEN fpro.vrd IN ($1) THEN fpro.iznos ELSE 0 END) as realizacija,
			SUM(CASE WHEN fpro.vrd IN ($2) THEN fpro.iznos ELSE 0 END) as placeno,
			SUM(CASE WHEN fpro.vrd IN ($1) AND fpro.dadok IS NOT NULL
				AND (fpro.dadok + fpro.rok)::date <= '%s'::date
				THEN fpro.iznos ELSE 0 END) as dospelarealizacija,
			SUM(CASE WHEN fpro.vrd IN ($1)
				AND (fpro.dadok + fpro.rok)::date > '%s'::date
				AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i15+`')::date
				THEN fpro.iznos ELSE 0 END) as dospece15,
			SUM(CASE WHEN fpro.vrd IN ($1)
				AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i15+`')::date
				AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i30+`')::date
				THEN fpro.iznos ELSE 0 END) as dospece30,
			SUM(CASE WHEN fpro.vrd IN ($1)
				AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i30+`')::date
				AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i60+`')::date
				THEN fpro.iznos ELSE 0 END) as dospece60,
			SUM(CASE WHEN fpro.vrd IN ($1)
				AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i60+`')::date
				AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i90+`')::date
				THEN fpro.iznos ELSE 0 END) as dospece90,
			SUM(CASE WHEN fpro.vrd IN ($1)
				AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i90+`')::date
				AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i120+`')::date
				THEN fpro.iznos ELSE 0 END) as dospece120,
			SUM(CASE WHEN fpro.vrd IN ($1)
				AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i120+`')::date
				THEN fpro.iznos ELSE 0 END) as dospece120plus
		FROM fpro`,
		sd, sd, sd, sd, sd, sd, sd, sd, sd, sd, sd, sd), true)
	qbP.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl")
	qbP.AddJoin("LEFT JOIN partneri p ON p.idpartneri = fkpl.idpartneri")
	qbP.AddArgs(vrd1, vrd2)
	if hasGod {
		qbP.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qbP.AddEqual("fpro.kar", session.SelectedKar)
	}
	qbP.AddCondition("fpro.konto::numeric", params.OdKonta, ">=")
	qbP.AddCondition("fpro.konto::numeric", params.DoKonta, "<=")
	qbP.AddCondition("NULLIF(fpro.sifra, '')::numeric", params.OdSifre, ">=")
	qbP.AddCondition("NULLIF(fpro.sifra, '')::numeric", params.DoSifre, "<=")
	qbP.AddCondition("fpro.dadok", params.StanjeNaDan, "<=")
	qbP.AddIn("fpro.vrd", []interface{}{vrd1, vrd2})
	qbP.AddGroupBy("fpro.konto, fpro.sifra, fkpl.idfkpl, fkpl.naziv, p.mesto")
	qbP.AddOrderBy("NULLIF(fpro.sifra, '')::numeric")
	sqlP, argsP := qbP.Build()

	partneriRows, err := s.fproRepo.GetAllCustom(ctx, sqlP, "", argsP, "", "")
	if err != nil {
		return nil, err
	}

	var result []domain.PregledPotrazivanjaPartnerReport
	for _, pr := range *partneriRows {
		report := domain.PregledPotrazivanjaPartnerReport{
			Konto:             pr.Konto,
			Sifra:             pr.Sifra,
			Naziv:             pr.NazivAnalitike,
			Mesto:             pr.Mesto,
			UkupnaRealizacija: common.FormatNumberWithSystemLocale(pr.Realizacija, 2),
			Placeno:           common.FormatNumberWithSystemLocale(pr.Placeno, 2),
			UkupanDug:         common.FormatNumberWithSystemLocale(pr.Realizacija-pr.Placeno, 2),
			DospeliDug:        common.FormatNumberWithSystemLocale(pr.DospelaRealizacija-pr.Placeno, 2),
			Dospece15:         common.FormatNumberWithSystemLocale(pr.Dospece15, 2),
			Dospece30:         common.FormatNumberWithSystemLocale(pr.Dospece30, 2),
			Dospece60:         common.FormatNumberWithSystemLocale(pr.Dospece60, 2),
			Dospece90:         common.FormatNumberWithSystemLocale(pr.Dospece90, 2),
			Dospece120:        common.FormatNumberWithSystemLocale(pr.Dospece120, 2),
			Dospece120Plus:    common.FormatNumberWithSystemLocale(pr.Dospece120Plus, 2),
		}

		// Detail rows per document for this partner
		qbD := common.NewQueryBuilder(fmt.Sprintf(`
			SELECT fpro.dokum, fpro.dadok, fpro.rok,
				CASE WHEN fpro.vrd IN ($1) THEN fpro.iznos ELSE 0 END as realizacija,
				CASE WHEN fpro.vrd IN ($2) THEN fpro.iznos ELSE 0 END as placeno,
				CASE WHEN fpro.vrd IN ($1) AND fpro.dadok IS NOT NULL
					AND (fpro.dadok + fpro.rok)::date <= '%s'::date
					THEN fpro.iznos ELSE 0 END as dospelarealizacija,
				CASE WHEN fpro.vrd IN ($1)
					AND (fpro.dadok + fpro.rok)::date > '%s'::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i15+`')::date
					THEN fpro.iznos ELSE 0 END as dospece15,
				CASE WHEN fpro.vrd IN ($1)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i15+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i30+`')::date
					THEN fpro.iznos ELSE 0 END as dospece30,
				CASE WHEN fpro.vrd IN ($1)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i30+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i60+`')::date
					THEN fpro.iznos ELSE 0 END as dospece60,
				CASE WHEN fpro.vrd IN ($1)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i60+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i90+`')::date
					THEN fpro.iznos ELSE 0 END as dospece90,
				CASE WHEN fpro.vrd IN ($1)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i90+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i120+`')::date
					THEN fpro.iznos ELSE 0 END as dospece120,
				CASE WHEN fpro.vrd IN ($1)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i120+`')::date
					THEN fpro.iznos ELSE 0 END as dospece120plus
			FROM fpro`,
			sd, sd, sd, sd, sd, sd, sd, sd, sd, sd, sd, sd), true)
		qbD.AddArgs(vrd1, vrd2)
		if hasGod {
			qbD.AddEqual("fpro.god", session.SelectedGod)
		}
		if hasKar {
			qbD.AddEqual("fpro.kar", session.SelectedKar)
		}
		qbD.AddEqual("fpro.idfkpl", pr.IDFkpl)
		qbD.AddCondition("fpro.dadok", params.StanjeNaDan, "<=")
		qbD.AddIn("fpro.vrd", []interface{}{vrd1, vrd2})
		qbD.AddOrderBy("fpro.dadok, fpro.dokum")
		sqlD, argsD := qbD.Build()

		detalji, err := s.fproRepo.GetAllCustom(ctx, sqlD, "", argsD, "", "")
		if err != nil {
			return nil, err
		}
		for _, d := range *detalji {
			dospece := common.AddDaysToNullTime(d.Dadok, d.Rok, common.DateLayout)
			item := domain.PregledPotrazivanjaItem{
				BrDok:          d.Dokum,
				DatDok:         common.FormatNullTime(d.Dadok, common.DateLayout),
				Rok:            fmt.Sprintf("%d", d.Rok),
				DatDospeca:     dospece,
				Iznos:          common.FormatNumberWithSystemLocale(d.Realizacija, 2),
				Placeno:        common.FormatNumberWithSystemLocale(d.Placeno, 2),
				DospeliIznos:   common.FormatNumberWithSystemLocale(d.DospelaRealizacija, 2),
				Dospece15:      common.FormatNumberWithSystemLocale(d.Dospece15, 2),
				Dospece30:      common.FormatNumberWithSystemLocale(d.Dospece30, 2),
				Dospece60:      common.FormatNumberWithSystemLocale(d.Dospece60, 2),
				Dospece90:      common.FormatNumberWithSystemLocale(d.Dospece90, 2),
				Dospece120:     common.FormatNumberWithSystemLocale(d.Dospece120, 2),
				Dospece120Plus: common.FormatNumberWithSystemLocale(d.Dospece120Plus, 2),
			}
			report.Items = append(report.Items, item)
		}
		result = append(result, report)
	}
	return result, nil
}

// GetPregledDugovanjaPoStarosti retrieves payables overview by age
func (s *OtvoreneStavkeResource) GetPregledDugovanjaPoStarosti(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// Validate interval parameters - must be numeric only to prevent SQL injection
	intervals := map[string]string{
		"dospece15":  params.Dospece15,
		"dospece30":  params.Dospece30,
		"dospece60":  params.Dospece60,
		"dospece90":  params.Dospece90,
		"dospece120": params.Dospece120,
	}

	for name, value := range intervals {
		if !regexp.MustCompile(`^\d+$`).MatchString(value) {
			return fmt.Errorf("invalid %s parameter: must be numeric", name)
		}
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build age bucket intervals
	interval15 := params.Dospece15 + " days"
	interval30 := params.Dospece30 + " days"
	interval60 := params.Dospece60 + " days"
	interval90 := params.Dospece90 + " days"
	interval120 := params.Dospece120 + " days"

	// Get data for the table with age bucket categorization
	qb := common.NewQueryBuilder(fmt.Sprintf(`
		SELECT
			fpro.konto, 
			fpro.sifra, 
			fkpl.naziv as nazivanalitike,
			-- Total realized amount (invoices only)
			SUM(CASE WHEN fpro.vrd IN (10, 20) THEN fpro.iznos ELSE 0 END) as realizacija,
			-- Total payments
			SUM(CASE WHEN fpro.vrd IN (30, 40) THEN fpro.iznos ELSE 0 END) as placeno,
			-- Already due (past state date)
			SUM(CASE WHEN fpro.vrd IN (10, 20) AND fpro.dadok IS NOT NULL AND fpro.dadok::date <= '%s'::date
					THEN fpro.iznos ELSE 0 END) as dospelarealizacija,
			-- Due 0-N days from state date (using dospece15)
			SUM(CASE WHEN fpro.vrd IN (10, 20) 
					AND (fpro.dadok + fpro.rok)::date > '%s'::date 
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval15+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece15,
			-- Due (dospece15+1)-N days from state date (using dospece30)
			SUM(CASE WHEN fpro.vrd IN (10, 20) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval15+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval30+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece30,
			-- Due (dospece30+1)-N days from state date (using dospece60)
			SUM(CASE WHEN fpro.vrd IN (10, 20) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval30+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval60+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece60,
			-- Due (dospece60+1)-N days from state date (using dospece90)
			SUM(CASE WHEN fpro.vrd IN (10, 20) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval60+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval90+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece90,
			-- Due (dospece90+1)-N days from state date (using dospece120)
			SUM(CASE WHEN fpro.vrd IN (10, 20) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval90+`')::date 
			 		AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+interval120+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece120,
			-- Due more than dospece120 days from state date
			SUM(CASE WHEN fpro.vrd IN (10, 20) 
			 		AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+interval120+`')::date
			 		THEN fpro.iznos ELSE 0 END) as dospece120plus
		FROM fpro
	`, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan,
		params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan,
		params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan, params.StanjeNaDan), true)
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")

	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}

	// Determine vkonta based on tipPregleda (K=kupci, D=dobavljaci)
	if params.TipPregleda == "K" {
		//qb.AddEqual("fpro.vkonta", "2")
	} else if params.TipPregleda == "D" {
		//qb.AddEqual("fpro.vkonta", "1")
	}

	qb.AddCondition("fpro.konto", params.OdKonta, ">=")
	qb.AddCondition("fpro.konto", params.DoKonta, "<=")
	qb.AddCondition("fpro.sifra", params.OdSifre, ">=")
	qb.AddCondition("fpro.sifra", params.DoSifre, "<=")

	// Add search conditions if search text is provided
	if params.SearchText != "" && params.SearchText != "undefined" && len(params.SearchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), params.SearchText)
	}

	// Add GROUP BY
	qb.AddGroupBy("fpro.konto, fpro.sifra, fkpl.naziv")

	// Add HAVING clause to filter partners with non-zero totals
	// qb.AddHaving(`SUM(CASE WHEN fpro.vrd IN (10, 20) THEN fpro.iznos ELSE 0 END) <> 0`)

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("fpro.sifra::numeric")

	sqlQuery, args := qb.Build()

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fields := []string{}
			// Detailed view with age buckets
			fields = append(fields,
				entity.Konto,
				entity.Sifra,
				entity.NazivAnalitike,
				common.FormatNumberWithSystemLocale(entity.Realizacija, 2),
				common.FormatNumberWithSystemLocale(entity.Placeno, 2),
				common.FormatNumberWithSystemLocale(entity.Realizacija-entity.Placeno, 2),
				common.FormatNumberWithSystemLocale(entity.DospelaRealizacija-entity.Placeno, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece15, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece30, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece60, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece90, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece120, 2),
				common.FormatNumberWithSystemLocale(entity.Dospece120Plus, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetPregledDugovanjaPoStarostiStampa retrieves all dospeli dug po starosti data grouped by partner for print report.
func (s *OtvoreneStavkeResource) GetPregledDugovanjaPoStarostiStampa(ctx context.Context, tbl *domain.TableData, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	// Validate interval parameters - must be numeric only to prevent SQL injection
	for name, value := range map[string]string{
		"dospece15": params.Dospece15, "dospece30": params.Dospece30,
		"dospece60": params.Dospece60, "dospece90": params.Dospece90,
		"dospece120": params.Dospece120,
	} {
		if !regexp.MustCompile(`^\d+$`).MatchString(value) {
			return fmt.Errorf("invalid %s parameter: must be numeric", name)
		}
	}
	// val15, _ := strconv.Atoi(params.Dospece15)
	// val30, _ := strconv.Atoi(params.Dospece30)
	// val60, _ := strconv.Atoi(params.Dospece60)
	// val90, _ := strconv.Atoi(params.Dospece90)
	// dana := translator.Label("Dana")
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	sd := params.StanjeNaDan
	i15 := params.Dospece15 + " days"
	i30 := params.Dospece30 + " days"
	i60 := params.Dospece60 + " days"
	i90 := params.Dospece90 + " days"
	i120 := params.Dospece120 + " days"

	qb := common.NewQueryBuilder(fmt.Sprintf(`
		SELECT fpro.konto, fpro.sifra, fkpl.idfkpl,
			fkpl.naziv as nazivanalitike,
			COALESCE(p.mesto, '') as mesto,
			SUM(CASE WHEN fpro.vrd IN (10, 20) THEN fpro.iznos ELSE 0 END) as realizacija,
			SUM(CASE WHEN fpro.vrd IN (30, 40) THEN fpro.iznos ELSE 0 END) as placeno,
			SUM(CASE WHEN fpro.vrd IN (10, 20) AND fpro.dadok IS NOT NULL AND fpro.dadok::date <= '%s'::date
					THEN fpro.iznos ELSE 0 END) as dospelarealizacija,
			SUM(CASE WHEN fpro.vrd IN (10, 20)
					AND (fpro.dadok + fpro.rok)::date > '%s'::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i15+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece15,
			SUM(CASE WHEN fpro.vrd IN (10, 20)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i15+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i30+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece30,
			SUM(CASE WHEN fpro.vrd IN (10, 20)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i30+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i60+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece60,
			SUM(CASE WHEN fpro.vrd IN (10, 20)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i60+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i90+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece90,
			SUM(CASE WHEN fpro.vrd IN (10, 20)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i90+`')::date
					AND (fpro.dadok + fpro.rok)::date <= ('%s'::date + INTERVAL '`+i120+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece120,
			SUM(CASE WHEN fpro.vrd IN (10, 20)
					AND (fpro.dadok + fpro.rok)::date > ('%s'::date + INTERVAL '`+i120+`')::date
					THEN fpro.iznos ELSE 0 END) as dospece120plus
		FROM fpro`,
		sd, sd, sd, sd, sd, sd, sd, sd, sd, sd, sd, sd), true)
	qb.AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("LEFT JOIN partneri p ON p.idpartneri = fkpl.idpartneri")
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}
	qb.AddCondition("fpro.konto", params.OdKonta, ">=")
	qb.AddCondition("fpro.konto", params.DoKonta, "<=")
	qb.AddCondition("fpro.sifra", params.OdSifre, ">=")
	qb.AddCondition("fpro.sifra", params.DoSifre, "<=")
	qb.AddGroupBy("fpro.konto, fpro.sifra, fkpl.idfkpl, fkpl.naziv, p.mesto")
	qb.AddOrderBy("fpro.sifra::numeric")
	sqlQuery, argsQuery := qb.Build()

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", argsQuery, "", "")
	if err != nil {
		return err
	}
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
	totRealiz, totPlaceno, totDug, totDospeliDug, totDosp15, totDosp30, totDosp60, totDosp90, totDosp120, totDosp120Plus := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	for _, entity := range *entities {
		fields := []string{
			entity.Konto,
			entity.Sifra,
			entity.NazivAnalitike,
			entity.Mesto,
			common.FormatNumberWithSystemLocale(entity.Realizacija, 2),
			common.FormatNumberWithSystemLocale(entity.Placeno, 2),
			common.FormatNumberWithSystemLocale(entity.Realizacija-entity.Placeno, 2),
			common.FormatNumberWithSystemLocale(entity.DospelaRealizacija-entity.Placeno, 2),
			common.FormatNumberWithSystemLocale(entity.Dospece15, 2),
			common.FormatNumberWithSystemLocale(entity.Dospece30, 2),
			common.FormatNumberWithSystemLocale(entity.Dospece60, 2),
			common.FormatNumberWithSystemLocale(entity.Dospece90, 2),
			common.FormatNumberWithSystemLocale(entity.Dospece120, 2),
			common.FormatNumberWithSystemLocale(entity.Dospece120Plus, 2),
		}
		totRealiz += entity.Realizacija
		totPlaceno += entity.Placeno
		totDug += entity.Realizacija - entity.Placeno
		totDospeliDug += entity.DospelaRealizacija - entity.Placeno
		totDosp15 += entity.Dospece15
		totDosp30 += entity.Dospece30
		totDosp60 += entity.Dospece60
		totDosp90 += entity.Dospece90
		totDosp120 += entity.Dospece120
		totDosp120Plus += entity.Dospece120Plus

		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFpro), Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
	tbl.Totals[4] = common.FormatNumberWithSystemLocale(totRealiz, 2)
	tbl.Totals[5] = common.FormatNumberWithSystemLocale(totPlaceno, 2)
	tbl.Totals[6] = common.FormatNumberWithSystemLocale(totDug, 2)
	tbl.Totals[7] = common.FormatNumberWithSystemLocale(totDospeliDug, 2)
	tbl.Totals[8] = common.FormatNumberWithSystemLocale(totDosp15, 2)
	tbl.Totals[9] = common.FormatNumberWithSystemLocale(totDosp30, 2)
	tbl.Totals[10] = common.FormatNumberWithSystemLocale(totDosp60, 2)
	tbl.Totals[11] = common.FormatNumberWithSystemLocale(totDosp90, 2)
	tbl.Totals[12] = common.FormatNumberWithSystemLocale(totDosp120, 2)
	tbl.Totals[13] = common.FormatNumberWithSystemLocale(totDosp120Plus, 2)
	return nil
}

// GetPovezivanjeRacunaUplata retrieves account and payment linking data
func (s *OtvoreneStavkeResource) GetPovezivanjeRacunaUplata(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	return nil
}
func (s *OtvoreneStavkeResource) GetOtvoreneStavkeDetaljiFields() []domain.Fields {
	return s.fieldsOtvoreneStavkeDetalji
}
func (s *OtvoreneStavkeResource) GetPartneriFields() []domain.Fields {
	return s.fieldsPartneri
}
func (s *OtvoreneStavkeResource) GetPartneriFieldsSintetika() []domain.Fields {
	return s.fieldsPartneriSintetika
}
func (s *OtvoreneStavkeResource) GetZatvoreneStavkeDetaljiFields() []domain.Fields {
	return s.fieldsZatvoreneStavkeDetalji
}
func (s *OtvoreneStavkeResource) GetIOSDetaljiFields() []domain.Fields {
	return s.fieldsIOSDetalji
}
func (s *OtvoreneStavkeResource) GetDospelaDetaljiFields() []domain.Fields {
	return s.fieldsDospelaDetalji
}
func (s *OtvoreneStavkeResource) GetDugovanjaFields() []domain.Fields {
	return s.fieldsPotrazivanjaObaveze
}
func (s *OtvoreneStavkeResource) GetPovezivanjePartneriFields() []domain.Fields {
	return s.fieldsPovezivanjePartneri
}
func (s *OtvoreneStavkeResource) GetPovezivanjeUplateFields() []domain.Fields {
	return s.fieldsPovezivanjeUplate
}
func (s *OtvoreneStavkeResource) GetPovezivanjeFaktureFields() []domain.Fields {
	return s.fieldsPovezivanjeFakture
}
func (s *OtvoreneStavkeResource) GetZatvoreneStavkeStampaFields() []domain.Fields {
	return s.fieldsZatvoreneStavkeStampa
}

// GetOpomenaStampaFields retrieves the fields configuration for the "Opomena Stampa" report.
func (s *OtvoreneStavkeResource) GetOpomenaStampaFields() []domain.Fields {
	return s.fieldsOpomenaStampa
}

// GetFvrData retrieves company (fvr) data for the current session.
func (s *OtvoreneStavkeResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return utils.GetFvrData(ctx, s.fvrRepo)
}
func (s *OtvoreneStavkeResource) GetPotrazivanjaDugovanjaStampaFields() []domain.Fields {
	return s.fieldsPotrazivanjaDugovanjaStampa
}
func (s *OtvoreneStavkeResource) GetDugovanjaStarostiStampaFields() []domain.Fields {
	return s.fieldsDugovanjaStarostiStampaFields
}

// setServiceFieldValues initializes table field definitions for Otvorene Stavke
func (s *OtvoreneStavkeResource) setServiceFieldValues() {
	s.fieldsPartneri = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false, IncludeInTotals: true},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivanalitike", Label: "Naziv analitike", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fpro.dug", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fpro.pot", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
	}
	s.fieldsPartneriSintetika = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivanalitike", Label: "Naziv analitike", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
	}
	s.fieldsOtvoreneStavkeDetalji = []domain.Fields{
		{Name: "det", Label: "Detalji", Width: "5", Field: "fpro.detalj", SkipInSearch: true},
		{Name: "doktip", Label: "F/U", Width: "5", Field: "fpro.fu", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokum", Width: "10", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "tra", Label: "Godina dokum", Width: "5", Field: "fpro.tra", SkipInSearch: false},
		{Name: "oj", Label: "Org. jedinica", Width: "12", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "nalog", Label: "Broj naloga", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "dadok", Label: "Datum dokumenta", Width: "10", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "10", Field: "fpro.rok", SkipInSearch: false},
		{Name: "dospece", Label: "Datum dospeća", Width: "10", Field: "fpro.dospece", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fpro.dug", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fpro.pot", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "15", Field: "fpro.opis", SkipInSearch: false},
	}

	// Zatvorene Stavke fields (Tab 2)
	s.fieldsZatvoreneStavkeDetalji = []domain.Fields{
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "tra", Label: "Godina dokum", Width: "5", Field: "fpro.tra", SkipInSearch: false},
		{Name: "oj", Label: "Org. jedinica", Width: "12", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "5", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Broj naloga", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "dadok", Label: "Datum dokumenta", Width: "10", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "10", Field: "fpro.rok", SkipInSearch: false},
		{Name: "dospece", Label: "Datum dospeća", Width: "10", Field: "fpro.dospece", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fpro.dug", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fpro.pot", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "12", Field: "fpro.opis", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "5", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "vkonta", Label: "Vrsta konta", Width: "5", Field: "fpro.vkonta", SkipInSearch: false},
		{Name: "kat", Label: "Način knjizenja", Width: "5", Field: "fpro.kat", SkipInSearch: false},
	}
	s.fieldsZatvoreneStavkeStampa = []domain.Fields{
		{Name: "redbr", Label: "Redni broj", Width: "5", Field: "rbr", TextAlign: "right", IncludeInTotals: false},
		{Name: "vrd", Label: "Vrsta Naloga za knjizenje", Width: "8", Field: "vrd", TextAlign: "left"},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "nalog", TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "danal", TextAlign: "left"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "12", Field: "dokum", TextAlign: "left"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "10", Field: "dadok", TextAlign: "left"},
		{Name: "rok", Label: "Rok", Width: "6", Field: "rok", TextAlign: "left"},
		{Name: "dospece", Label: "Dospece", Width: "10", Field: "dospece", TextAlign: "left"},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", TextAlign: "right", IncludeInTotals: true},
	}
	s.fieldsOpomenaStampa = []domain.Fields{
		{Name: "redbr", Label: "Redni Broj", Width: "5", Field: "redbr", TextAlign: "center", IncludeInTotals: true},
		{Name: "opis", Label: "Opis knjiženja", Width: "20", Field: "opis", TextAlign: "left"},
		{Name: "oj", Label: "OJ", Width: "8", Field: "oj", TextAlign: "center"},
		{Name: "brdok", Label: "Broj dokumenta", Width: "10", Field: "brdok", TextAlign: "center"},
		{Name: "datdok", Label: "Datum dokumenta", Width: "10", Field: "datdok", TextAlign: "center"},
		{Name: "goddok", Label: "Godina dokumenta", Width: "8", Field: "goddok", TextAlign: "center"},
		{Name: "rok", Label: "Rok", Width: "6", Field: "rok", TextAlign: "center"},
		{Name: "dospece", Label: "Dospece", Width: "10", Field: "dospece", TextAlign: "center"},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", TextAlign: "right", IncludeInTotals: true},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "iznos", TextAlign: "right", IncludeInTotals: true},
	}
	s.fieldsIOSStampa = []domain.Fields{
		{Name: "redbr", Label: "Redni Broj", Width: "5", Field: "redbr", TextAlign: "center"},
		{Name: "opis", Label: "Opis knjiženja", Width: "20", Field: "opis", TextAlign: "left"},
		{Name: "oj", Label: "OJ", Width: "8", Field: "oj", TextAlign: "center"},
		{Name: "brdok", Label: "Broj dokumenta", Width: "10", Field: "brdok", TextAlign: "center"},
		{Name: "datdok", Label: "Datum dokumenta", Width: "10", Field: "datdok", TextAlign: "center"},
		{Name: "rok", Label: "Rok", Width: "6", Field: "rok", TextAlign: "center"},
		{Name: "dospece", Label: "Dospeće", Width: "10", Field: "dospece", TextAlign: "center"},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", TextAlign: "right", IncludeInTotals: true},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "iznos", TextAlign: "right", IncludeInTotals: true},
	}
	s.fieldsIOSDetalji = []domain.Fields{
		{Name: "doktip", Label: "F/U", Width: "5", Field: "fpro.fu", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokum", Width: "10", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "tra", Label: "Godina dokum", Width: "5", Field: "fpro.tra", SkipInSearch: false},
		{Name: "oj", Label: "Org. jedinica", Width: "12", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "nalog", Label: "Broj naloga", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "dadok", Label: "Datum dokumenta", Width: "10", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "10", Field: "fpro.rok", SkipInSearch: false},
		{Name: "dospece", Label: "Datum dospeća", Width: "10", Field: "fpro.dospece", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fpro.dug", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fpro.pot", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "15", Field: "fpro.opis", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "5", Field: "fpro.vrd", SkipInSearch: false},
	}
	s.fieldsDospelaDetalji = []domain.Fields{
		{Name: "doktip", Label: "F/U", Width: "5", Field: "fpro.fu", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokum", Width: "10", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "10", Field: "fpro.rok", SkipInSearch: false},
		{Name: "dospece", Label: "Datum dospeća", Width: "10", Field: "fpro.dospece", SkipInSearch: false},
		{Name: "dug", Label: "Dospeli dug", Width: "12", Field: "fpro.dug", SkipInSearch: true},
		{Name: "pot", Label: "Plaćeno", Width: "12", Field: "fpro.pot", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "15", Field: "fpro.opis", SkipInSearch: false},
	}

	s.fieldsPotrazivanjaObavezePartneri = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivanalitike", Label: "Naziv analitike", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
	}
	// Dospela Potraživanja fields (Tab 5, 6, 7)
	s.fieldsPotrazivanjaObaveze = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivanalitike", Label: "Naziv analitike", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "realizacija", Label: "Ukupna realizacija", Width: "12", Field: "fpro.realizacija", SkipInSearch: true},
		{Name: "placeno", Label: "Plaćeno", Width: "12", Field: "fpro.placeno", SkipInSearch: true},
		{Name: "ukdug", Label: "Ukupno dugovanje/obaveza", Width: "12", Field: "fpro.ukdug", SkipInSearch: true},
		{Name: "dospelidug", Label: "Dospeli dug", Width: "12", Field: "fpro.dospelarealizacija", SkipInSearch: true},
		{Name: "dospece15", Label: "Dospeva 15 dana", Width: "12", Field: "fpro.dospece15", SkipInSearch: true, Params: map[string]string{"days": "0-15"}},
		{Name: "dospece30", Label: "Dospeva 30 dana", Width: "12", Field: "fpro.dospece30", SkipInSearch: true, Params: map[string]string{"days": "16-30"}},
		{Name: "dospece60", Label: "Dospeva 60 dana", Width: "12", Field: "fpro.dospece60", SkipInSearch: true, Params: map[string]string{"days": "31-60"}},
		{Name: "dospece90", Label: "Dospeva 90 dana", Width: "12", Field: "fpro.dospece90", SkipInSearch: true, Params: map[string]string{"days": "61-90"}},
		{Name: "dospece120", Label: "Dospeva 120 dana", Width: "12", Field: "fpro.dospece120", SkipInSearch: true, Params: map[string]string{"days": "91-120"}},
		{Name: "dospece120plus", Label: "Dospeva >120 dana", Width: "12", Field: "fpro.dospece120plus", SkipInSearch: true, Params: map[string]string{"days": ">120"}},
	}
	s.fieldsPovezivanjePartneri = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "15", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "15", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv", Width: "70", SkipInSearch: false},
	}
	s.fieldsPovezivanjeUplate = []domain.Fields{
		{Name: "brojnaloga", Label: "Broj naloga", Width: "12", SkipInSearch: false},
		{Name: "datumnaloga", Label: "Datum naloga", Width: "13", SkipInSearch: false},
		{Name: "dokument", Label: "Dokument", Width: "12", SkipInSearch: false},
		{Name: "datumdok", Label: "Datum dokumenta", Width: "14", SkipInSearch: false},
		{Name: "dospece", Label: "Dospeće", Width: "13", SkipInSearch: false},
		{Name: "ostatakuplate", Label: "Ostatak Uplate", Width: "13", SkipInSearch: false, IncludeInTotals: true, TextAlign: "right"},
	}
	s.fieldsPovezivanjeFakture = []domain.Fields{
		{Name: "brojnaloga", Label: "Broj naloga", Width: "9", SkipInSearch: false},
		{Name: "brojdokumenta", Label: "Broj dokumenta", Width: "9", SkipInSearch: false},
		{Name: "poslovgod", Label: "Poslov. godina", Width: "8", SkipInSearch: false},
		{Name: "oj", Label: "OJ", Width: "5", SkipInSearch: false},
		{Name: "datumdok", Label: "Datum dokumenta", Width: "9", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "7", SkipInSearch: false},
		{Name: "dospece", Label: "Dospeće", Width: "8", SkipInSearch: false},
		{Name: "iznos", Label: "Iznos računa", Width: "9", SkipInSearch: false, IncludeInTotals: true, TextAlign: "right"},
		{Name: "nezatvoreniiznos", Label: "Nezatvoreni iznos", Width: "9", SkipInSearch: false, IncludeInTotals: true, TextAlign: "right"},
		{Name: "kompenzacije", Label: "Kompenzacije u toku", Width: "9", SkipInSearch: false, IncludeInTotals: true, TextAlign: "right"},
		{Name: "iznoszatvaranje", Label: "Iznos za zatvaranje", Width: "9", SkipInSearch: false, IncludeInTotals: true, TextAlign: "right"},
	}
	s.fieldsPotrazivanjaDugovanjaStampa = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv partnera", Width: "20", SkipInSearch: false},
		{Name: "mesto", Label: "Mesto", Width: "12", SkipInSearch: false},
		{Name: "realizacija", Label: "Ukupna realizacija", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "placeno", Label: "Plaćeno", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "ukdug", Label: "Ukupan DUG", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospelidug", Label: "Dospeli DUG", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece15", Label: "Dospeva za 0-15 dana", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece30", Label: "Dospeva za 16-30 dana", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece60", Label: "Dospeva za 31-60 dana", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece90", Label: "Dospeva za 61-90 dana", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece120", Label: "Dospeva za 91-120 dana", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece120plus", Label: "Dospeva za >120 dana", Width: "12", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
	}
	s.fieldsDugovanjaStarostiStampaFields = []domain.Fields{
		{Name: "konto", Label: "Konto", TextAlign: "center", IncludeInTotals: true},
		{Name: "sifra", Label: "Sifra", TextAlign: "center"},
		{Name: "naziv", Label: "Naziv Partnera", TextAlign: "left"},
		{Name: "mesto", Label: "Mesto", TextAlign: "center"},
		{Name: "realizacija", Label: "Ukupna realizacija", TextAlign: "right"},
		{Name: "placeno", Label: "Placeno", TextAlign: "right"},
		{Name: "ukdug", Label: "Ukupan DUG", TextAlign: "right"},
		{Name: "dospelidug", Label: "Dospeli DUG", TextAlign: "right"},
		{Name: "dospece15", Label: "Dugov staro", Params: map[string]string{"days": "0-15"}, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece30", Label: "Dugov staro", Params: map[string]string{"days": "16-30"}, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece60", Label: "Dugov staro", Params: map[string]string{"days": "31-60"}, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece90", Label: "Dugov staro", Params: map[string]string{"days": "61-90"}, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece120", Label: "Dugov staro", Params: map[string]string{"days": "91-120"}, TextAlign: "right", IncludeInTotals: true},
		{Name: "dospece120plus", Label: "Dugov staro", Params: map[string]string{"days": ">120"}, TextAlign: "right", IncludeInTotals: true},
	}

}

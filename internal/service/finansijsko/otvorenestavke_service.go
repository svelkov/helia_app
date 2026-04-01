package finansijsko

import (
	"context"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

// OtvoreneStavkeService defines the service interface for Otvorene Stavke operations
type OtvoreneStavkeService interface {
	GetOtvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetOtvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetZatvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetZatvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetIOSPartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetIOSDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetDospelaPotrazivanjaPartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetDospelaPotrazivanjaDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetPregledPotrazivanjaObaveze(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetPregledDugovanjaPoStarosti(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error
	GetPovezivanjRacunaIUplata(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetPartneriFields() []domain.Fields
	GetPartneriFieldsSintetika() []domain.Fields
	GetOtvoreneStavkeDetaljiFields() []domain.Fields
	GetZatvoreneStavkeDetaljiFields() []domain.Fields
	GetIOSDetaljiFields() []domain.Fields
	GetDospelaDetaljiFields() []domain.Fields
	GetDugovanjaFields() []domain.Fields
}

// OtvoreneStavkeResource implements OtvoreneStavkeService
type OtvoreneStavkeResource struct {
	fproRepo                          repository.BaseRepository[domain.Fpro]
	fieldsPartneri                    []domain.Fields
	fieldsPartneriSintetika           []domain.Fields
	fieldsOtvoreneStavkeDetalji       []domain.Fields
	fieldsZatvoreneStavke             []domain.Fields
	fieldsDospelaPotrazivanja         []domain.Fields
	fieldsZatvoreneStavkeDetalji      []domain.Fields
	fieldsIOSDetalji                  []domain.Fields
	fieldsDospelaDetalji              []domain.Fields
	fieldsPotrazivanjaObavezePartneri []domain.Fields
	fieldsPotrazivanjaObaveze         []domain.Fields
}

// NewOtvoreneStavkeService creates a new service instance
func NewOtvoreneStavkeService(repo repository.BaseRepository[domain.Fpro]) *OtvoreneStavkeResource {
	service := &OtvoreneStavkeResource{
		fproRepo: repo,
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
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivpartnera, 
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
			totalDug += entity.Dug
			totalPot += entity.Pot
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
				entity.NazivPartnera,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetOtvoreneStavkeDetalji retrieves open items details for partners
func (s *OtvoreneStavkeResource) GetOtvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
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
								fpro.opis, fpro.vrd
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
				"+",
				entity.DocType,
				entity.Dokum,
				fmt.Sprintf("%d", entity.Tra),
				entity.Ojozn,
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				common.FormatNullTime(entity.Dadok, common.DateLayout),
				fmt.Sprintf("%d", entity.Rok),
				common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
				entity.Opis)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

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
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivpartnera, 
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
				entity.NazivPartnera,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetZatvoreneStavkeDetalji retrieves closed items details for partners
func (s *OtvoreneStavkeResource) GetZatvoreneStavkeDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	qb := common.NewQueryBuilder(`SELECT
								fpro.dokum, fpro.tra, fpro.ojozn, fpro.tipdok, fpro.nalog, fpro.danal, 
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
	qb.AddEqual("fpro.idfkpl", id)
	// Add search conditions if search text is provided
	if searchText != "" && searchText != "undefined" && len(searchText) >= 1 {
		qb.SetEntityType(reflect.TypeOf(domain.Fpro{}))
		qb.AddSearchConditions(s.GetPartneriFields(), searchText)
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
				entity.Dokum,
				fmt.Sprintf("%d", entity.Tra),
				entity.Ojozn,
				entity.Tipdok,
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				common.FormatNullTime(entity.Dadok, common.DateLayout),
				fmt.Sprintf("%d", entity.Rok),
				common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
				entity.Opis,
				fmt.Sprintf("%d", entity.Vrd),
				fmt.Sprintf("%d", entity.Vkonta),
				fmt.Sprintf("%d", entity.Kat),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
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
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv as nazivpartnera, 
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
				entity.NazivPartnera,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
			)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetIOSDetalji retrieves IOS details for partners
func (s *OtvoreneStavkeResource) GetIOSDetalji(ctx context.Context, id int64, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
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
								fpro.opis, fpro.vrd
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
				fmt.Sprintf("%d", entity.Tra),
				entity.Ojozn,
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				common.FormatNullTime(entity.Dadok, common.DateLayout),
				fmt.Sprintf("%d", entity.Rok),
				common.AddDaysToNullTime(entity.Dadok, entity.Rok, common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
				entity.Opis,
				fmt.Sprintf("%d", entity.Vrd))
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
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
	qb := common.NewQueryBuilder(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivpartnera, 
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
					entity.NazivPartnera,
					common.FormatNumberWithSystemLocale(entity.Dug, 2),
					common.FormatNumberWithSystemLocale(entity.Pot, 2),
					common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
				)
			}
			if params.TipPregleda == "S" {
				fields = append(fields,
					entity.Konto,
					entity.Sifra,
					entity.NazivPartnera,
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
								fpro.dokum, fpro.tra, fpro.ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								fpro.opis, fpro.vrd
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
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2),
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				entity.Opis)
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
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
			fpro.sifra, 
			fkpl.naziv as nazivpartnera,
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

	qb.AddCondition("fpro.konto", params.OdKonta, ">=")
	qb.AddCondition("fpro.konto", params.DoKonta, "<=")
	qb.AddCondition("fpro.sifra", params.OdSifre, ">=")
	qb.AddCondition("fpro.sifra", params.DoSifre, "<=")
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
				entity.NazivPartnera,
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
			fkpl.naziv as nazivpartnera,
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
				entity.NazivPartnera,
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

// GetPovezivanjRacunaIUplata retrieves account and payment linking data
func (s *OtvoreneStavkeResource) GetPovezivanjRacunaIUplata(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	//TODO should be implemented with actual database query
	// if err := common.ValidateRequiredParams(c, "konto"); err != nil {
	// 	return err
	// }

	// common.SetTableConfig(tbl, "Povezivanje računa i uplata", "", false, false, false)
	// common.SetupTablePagination(tbl, currentPage, pageSize)
	// tbl.Fields = s.fieldOtvoreneStavke

	// // Mock data - replace with actual database query
	// tbl.Rows = []map[string]interface{}{}
	// tbl.TotalRecords = 0

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

// setServiceFieldValues initializes table field definitions for Otvorene Stavke
func (s *OtvoreneStavkeResource) setServiceFieldValues() {
	s.fieldsPartneri = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false, IncludeInTotals: true},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivpartnera", Label: "Naziv partnera", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "fpro.dug", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "fpro.pot", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
	}
	s.fieldsPartneriSintetika = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivpartnera", Label: "Naziv partnera", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
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
		{Name: "dokum", Label: "Broj dokum", Width: "10", Field: "fpro.dokum", SkipInSearch: false},
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
		{Name: "nazivpartnera", Label: "Naziv partnera", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
	}
	// Dospela Potraživanja fields (Tab 5, 6, 7)
	s.fieldsPotrazivanjaObaveze = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "nazivpartnera", Label: "Naziv partnera", Width: "15", Field: "fkpl.naziv", SkipInSearch: false},
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
}

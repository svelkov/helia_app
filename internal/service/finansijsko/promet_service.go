package finansijsko

import (
	"context"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// FproViewData encapsulates all data needed for the Nalog display page.
type PrometViewData struct {
	FproEntities []domain.Fpro
	prometRepo   repository.BaseRepository[domain.PrometDto]
	TableData    domain.TableData
}

// NalogService defines the interface for operations related to Fpro (Nalogs).
type PrometService interface {
	GetFieldCache() map[string]reflect.StructField
	GetPrometAnalitickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, isMI bool, params domain.PrometParam) error
	GetPrometSubsintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometKarticaSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometKontaAnaliticki(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometSubsintetikaVrd(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam, tipStampe string) error
	GetPrometDeviznaAnalitickaKonta(ctx context.Context, tbl, tblDevizni *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error
	GetPrometTotals(ctx context.Context, params domain.PrometParam) (domain.PrometResponse, error)
	GetAnkontaTableFields() []domain.Fields
	GetAnKontaMiTableFields() []domain.Fields
	GetAnDeviznaKontaTableFields() []domain.Fields
	GetAnDeviznaKontaRekapTableFields() []domain.Fields
	GetSubsintetickihKontaTableFields() []domain.Fields
	GetSintetickihKontaTableFields() []domain.Fields
	GetKarticaSintetikaTableFields() []domain.Fields
	GetSubsintetikaVrdTableFields() []domain.Fields
	GetKontaAnalitickiTableFields() []domain.Fields
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetPrometAnalitickaKarticaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometAnalitickaKarticaStampaTableFields() []domain.Fields
	GetPrometSubsintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometSubsintetickihKontaStampaTableFields() []domain.Fields
	GetPrometSintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometSintetickihKontaStampaTableFields() []domain.Fields
	GetPrometKarticaSintKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometKarticaSintKontaStampaTableFields() []domain.Fields
	GetPrometSubsintetikaVrdStampaTableFields() []domain.Fields
	GetPrometAnalitickaKarticaPoMIStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error
	GetPrometAnalitickaKarticaPoMIStampaTableFields() []domain.Fields
}

// PrometResource implements the PrometService interface.
type PrometResource struct {
	service                                      *service.BaseService[domain.PrometDto]
	prometRepo                                   *repository.BaseRepository[domain.PrometDto]
	fkplRepo                                     *repository.BaseRepository[domain.Fkpl]
	fvrRepo                                      *repository.BaseRepository[domain.Fvr]
	prometAnKontaTableFields                     []domain.Fields
	prometAnKontaMiTableFields                   []domain.Fields
	prometAnDeviznaKontaTableFields              []domain.Fields
	prometAnDeviznaKontaRekapTableFields         []domain.Fields
	prometSubsintetickihKontaTablefields         []domain.Fields
	prometSintetickihKontaTableFields            []domain.Fields
	prometKarticaSintetickihKontaTableFields     []domain.Fields
	prometSubsintetikaVrdTableFields             []domain.Fields
	prometKontaAnalitickiTableFields             []domain.Fields
	prometAnalitickaKarticaStampaTableFields     []domain.Fields
	prometSubsintetickihKontaStampaTableFields   []domain.Fields
	prometSintetickihKontaStampaTableFields      []domain.Fields
	prometKarticaSintKontaStampaTableFields      []domain.Fields
	prometSubsintetikaVrdStampaTableFields       []domain.Fields
	prometAnalitickaKarticaPoMIStampaTableFields []domain.Fields
}

func NewPrometService(service *service.BaseService[domain.PrometDto], prometRepo *repository.BaseRepository[domain.PrometDto], fkplRepo *repository.BaseRepository[domain.Fkpl], fvrRepo *repository.BaseRepository[domain.Fvr]) *PrometResource {
	rs := &PrometResource{
		service:    service,
		prometRepo: prometRepo,
		fkplRepo:   fkplRepo,
		fvrRepo:    fvrRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// prometTotalsSelect sums the duguje and potrazuje of fpro.
const prometTotalsSelect = `
		select
			coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje
		from fpro`

// prometKarticaSelect is the stavke of the analiticka kartica (and its devizna variant).
const prometKarticaSelect = `SELECT danal, tipdok, concat(tipdok,'-',nalog) as nalog, idfpro, kat, iznos, kolic,
		       	vrd, dokum, dadok, rok, tra, coalesce(ojozn, '') as oj, coalesce(opis, '') as opis, coalesce(sifval, 0) as sifval, kurs,
		       	deviznos, cena, konto, idfnal, idfkpl,
				coalesce(dokumv, '') as dokumv, dadokv, coalesce(travez, 0) as travez,
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				CASE WHEN kat = 1 OR kat = 2 THEN kolic ELSE 0 END as kolduguje,
				CASE WHEN kat = 3 OR kat = 4 THEN kolic ELSE 0 END as kolpotrazuje
		FROM fpro`

// wherePrometAnKonto adds the filters of the analiticka kartica: the analiticki konto (konto,
// sifra, vkonta 1) and the period od-do datuma.
func wherePrometAnKonto(b *querybuilder.Builder, params domain.PrometParam) {
	b.AddEqual("konto", params.Konto).
		AddEqual("sifra", params.Sifra).
		AddEqual("vkonta", 1).
		AddCondition("danal", params.OdDatuma, ">=").
		AddCondition("danal", params.DoDatuma, "<=")
}

// whereRange adds "expr >= od" and "expr <= do" for a static expression such as
// "sifra::numeric" or "left(fpro.konto, 3)"; each is skipped when empty, like the old
// AddCondition.
func whereRange(b *querybuilder.Builder, expr, od, do string) {
	if od != "" {
		b.Where(expr+" >= ?", od)
	}
	if do != "" {
		b.Where(expr+" <= ?", do)
	}
}

// prometTotalsUslovi adds the filters of the report tip to the totals query: with period the
// promet od-do datuma, without it the promet before odDatuma ("promet do").
func prometTotalsUslovi(b *querybuilder.Builder, reportTip string, params domain.PrometParam, period bool) {
	datum := func() {
		if period {
			b.AddCondition("danal", params.OdDatuma, ">=").AddCondition("danal", params.DoDatuma, "<=")
		} else {
			b.AddCondition("danal", params.OdDatuma, "<")
		}
	}
	switch reportTip {
	case "prometankonta", "deviznihanalitickihkonta": // Analiticka Konta, Devizna Analiticka Konta
		b.AddEqual("konto", params.Konto).AddEqual("sifra", params.Sifra).AddEqual("vkonta", 1)
		datum()
	case "prometankontami": // Analiticka Konta po MI
		b.AddEqual("konto", params.Konto).AddEqual("sifra", params.Sifra).AddEqual("vkonta", 1)
		datum()
		b.AddCondition("mi", params.OdMI, ">=").AddCondition("mi", params.DoMI, "<=")
	case "subsintetickakonta": // Subsinteticka Konta
		b.AddCondition("konto", params.OdKonta, ">=").AddCondition("konto", params.DoKonta, "<=")
		datum()
	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		b.AddLikeBegin("konto", params.Konto)
		datum()
	case "kontaanaliticki": // Konta Analiticki
		b.AddEqual("konto", params.Konto)
		whereRange(b, "sifra::numeric", params.OdSifre, params.DoSifre)
		datum()
		b.AddEqual("vkonta", 1)
	}
}

func (s *PrometResource) GetPrometTotals(ctx context.Context, params domain.PrometParam) (domain.PrometResponse, error) {
	var response domain.PrometResponse
	reportTip := params.ReportTip
	switch reportTip {
	case "prometankonta": // Analiticka Konta
		if params.Konto == "" || params.Sifra == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "prometankontami": // Analiticka Konta po MI
		if params.Konto == "" || params.Sifra == "" || params.OdDatuma == "" || params.DoDatuma == "" || params.OdMI == "" || params.DoMI == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "deviznihanalitickihkonta": // Devizna Analiticka Konta
		if params.Konto == "" || params.Sifra == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "subsintetickakonta": // Subsinteticka Konta
		if params.OdKonta == "" || params.DoKonta == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "sintetickakonta", "karticasintetickihkonta", "subsintetickakontapovrd": // Sinteticka Konta
		if params.Konto == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	case "kontaanaliticki": // Konta Analiticki
		if params.Konto == "" || params.OdSifre == "" || params.DoSifre == "" || params.OdDatuma == "" || params.DoDatuma == "" {
			return response, fmt.Errorf("missing required parameters")
		}
	default:
		return response, fmt.Errorf("invalid report tip")
	}

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return response, fmt.Errorf("user session not found")
	}

	// Get "promet do" totals (up to start date)
	bDo := querybuilder.New(prometTotalsSelect)
	prometTotalsUslovi(bDo, reportTip, params, false)
	if err := s.prometRepo.ApplyFilters(ctx, bDo, "", nil, ""); err != nil {
		return response, err
	}
	prometDoResults, err := s.prometRepo.Select(ctx, bDo)
	if err != nil {
		return response, fmt.Errorf("error getting promet do totals: %v", err)
	}

	var prometDoDuguje, prometDoPotrazuje float64
	if len(*prometDoResults) > 0 {
		prometDoDuguje = (*prometDoResults)[0].Duguje
		prometDoPotrazuje = (*prometDoResults)[0].Potrazuje
	}

	response.Totals = domain.TotalValues{
		DugDo:   common.FormatFloatNumber64WithSystemLocale(prometDoDuguje, 2),
		PotDo:   common.FormatFloatNumber64WithSystemLocale(prometDoPotrazuje, 2),
		SaldoDo: common.FormatFloatNumber64WithSystemLocale(prometDoDuguje-prometDoPotrazuje, 2),
	}

	// Get "promet za period" totals (for the specified period)
	bPeriod := querybuilder.New(prometTotalsSelect)
	prometTotalsUslovi(bPeriod, reportTip, params, true)
	if err := s.prometRepo.ApplyFilters(ctx, bPeriod, "", nil, ""); err != nil {
		return response, err
	}
	prometPeriodResults, err := s.prometRepo.Select(ctx, bPeriod)
	if err != nil {
		return response, fmt.Errorf("error getting promet period totals: %v", err)
	}

	var prometPeriodDuguje, prometPeriodPotrazuje float64
	if len(*prometPeriodResults) > 0 {
		prometPeriodDuguje = (*prometPeriodResults)[0].Duguje
		prometPeriodPotrazuje = (*prometPeriodResults)[0].Potrazuje
	}

	response.Totals.DugPer = common.FormatFloatNumber64WithSystemLocale(prometPeriodDuguje, 2)
	response.Totals.PotPer = common.FormatFloatNumber64WithSystemLocale(prometPeriodPotrazuje, 2)
	response.Totals.SaldoPer = common.FormatFloatNumber64WithSystemLocale(prometPeriodDuguje-prometPeriodPotrazuje, 2)
	response.Totals.DugTot = common.FormatFloatNumber64WithSystemLocale(response.Totals.DugDo+prometPeriodDuguje, 2)
	response.Totals.PotTot = common.FormatFloatNumber64WithSystemLocale(response.Totals.PotDo+prometPeriodPotrazuje, 2)
	response.Totals.SaldoTot = common.FormatFloatNumber64WithSystemLocale(response.Totals.SaldoDo+(prometPeriodDuguje-prometPeriodPotrazuje), 2)
	return response, err
}

func (s *PrometResource) CheckPrometParameters(ctx context.Context, requiredFields []string, params domain.PrometParam) (fieldsError []domain.FieldError) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "User session not found"}}
	}

	// Build query dynamically; the user conditions are optional (skipped when empty), as before.
	b := querybuilder.New(`SELECT f.konto, f.sifra FROM baza.fkpl as f`).
		AddEqual("f.konto", params.Konto).
		AddEqual("f.sifra", params.Sifra).
		AddEqual("f.vkonta", params.Vkonta)
	// Add system conditions
	var entities *[]domain.PrometDto
	err := s.service.Repo.ApplyFilters(ctx, b, "f", nil, "")
	if err == nil {
		entities, err = s.service.Repo.Select(ctx, b)
	}
	if err != nil {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetData}}
	}
	if len(*entities) == 0 {
		return []domain.FieldError{{Field: "konto", ErrorMessage: common.ErrMsgGetKontoSifra}}
	}

	return nil
}

func (s *PrometResource) GetPrometAnalitickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, isMI bool, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	// (The old code also built a totals query here that was never executed; it is dropped.)
	b := querybuilder.New(prometKarticaSelect)
	wherePrometAnKonto(b, params)
	if isMI {
		b.AddCondition("mi", params.OdMI, ">=").AddCondition("mi", params.DoMI, "<=")
	}
	// The period of the session and the search text, if provided.
	fields := s.GetAnkontaTableFields()
	if isMI {
		fields = s.GetAnKontaMiTableFields()
	}
	if err := s.prometRepo.ApplyFilters(ctx, b, "", fields, params.SearchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("god, kar, danal, tipdok, nalog, idfpro").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot, totKolDug, totKolPot, totDevIznos float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{}

			// Add MI as first field only if isMI is true
			if isMI {
				fields = append(fields, entity.MI)
			}

			// Add common fields
			fields = append(fields,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				entity.Tra,
				entity.Ojozn,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Kurs, 2),
				common.FormatNumberWithSystemLocale(entity.Deviznos, 2),
				common.FormatNumberWithSystemLocale(entity.Kolduguje, 2),
				common.FormatNumberWithSystemLocale(entity.Kolpotrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Cena, 2),
				common.FormatNumberWithSystemLocale(entity.Stanje, 2),
				entity.Dokumv,
				entity.Dadokv.Time.Format(common.DateLayout),
				entity.Travez,
			)
			if !isMI {
				fields = append(fields, entity.MI)
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
			totKolDug += entity.Kolduguje
			totKolPot += entity.Kolpotrazuje
			totDevIznos += entity.Deviznos
		}
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		case "kolduguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolDug, 2)
		case "kolpotrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolPot, 2)
		case "deviznos":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDevIznos, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometDeviznaAnalitickaKonta(ctx context.Context, tbl, tblDevizni *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(prometKarticaSelect)
	wherePrometAnKonto(b, params)
	bTotal := querybuilder.New(`SELECT
			   	coalesce(sifval, 0) as sifval,
			   	coalesce(sum(case when kat = 1 or kat = 2 then iznos else 0 end), 0) as duguje,
			   	coalesce(sum(case when kat = 3 or kat = 4 then iznos else 0 end), 0) as potrazuje,
				coalesce(sum(case when kat = 1 or kat = 2 then iznos * kurs else 0 end), 0) as devduguje,
			   	coalesce(sum(case when kat = 3 or kat = 4 then iznos * kurs else 0 end), 0) as devpotrazuje,
				coalesce(sum(deviznos), 0) as deviznos
		FROM fpro`)
	wherePrometAnKonto(bTotal, params)
	// The period of the session; the search text only on the stavke (not on the totals), as before.
	if err := s.prometRepo.ApplyFilters(ctx, b, "", s.GetAnkontaTableFields(), params.SearchText); err != nil {
		return err
	}
	if err := s.prometRepo.ApplyFilters(ctx, bTotal, "", nil, ""); err != nil {
		return err
	}
	bTotal.GroupBy("sifval")

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("god, kar, danal, tipdok, nalog").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot, totKolDug, totKolPot, totDevIznos float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Opis,
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Kurs, 4),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				entity.Tra,
				entity.Ojozn,
				entity.Konto,
			)
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
			totKolDug += entity.Kolduguje
			totKolPot += entity.Kolpotrazuje
			totDevIznos += entity.Deviznos
		}
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		case "kolduguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolDug, 2)
		case "kolpotrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolPot, 2)
		case "deviznos":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDevIznos, 2)
		}
	}
	entitiesDevizni, err := s.prometRepo.Select(ctx, bTotal)
	if err != nil {
		return err
	}
	if len(*entitiesDevizni) > 0 {
		for _, entity := range *entitiesDevizni {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Devduguje, 2),
				common.FormatNumberWithSystemLocale(entity.Devpotrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Devduguje-entity.Devpotrazuje, 2),
			)
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tblDevizni.Rows = append(tblDevizni.Rows, tblRow)
		}
	}

	return nil
}
func (s *PrometResource) GetPrometAnalitickaKarticaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()
	b := querybuilder.New(`SELECT
		f.tipdok, f.nalog, f.danal, f.vrd,
		COALESCE(f.opis, '') as opis,
		COALESCE(f.ojozn, '') as oj,
		f.dokum,
		f.tra,
		f.dadok,
		f.rok,
		f.brst,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(f.sifra, '') as sifra,
		COALESCE(fk.naziv, '') as naziv
	FROM fpro f`).
		AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl").
		AddEqual("f.vkonta", 1).
		AddCondition("f.konto", params.OdKonta, ">=").
		AddCondition("f.konto", params.DoKonta, "<=")
	whereRange(b, "f.sifra::numeric", params.OdSifre, params.DoSifre)
	b.AddCondition("f.danal", params.OdDatuma, ">=").
		AddCondition("f.danal", params.DoDatuma, "<=").
		OrderByExpr("f.konto, f.sifra::numeric, f.danal, f.nalog::numeric, f.brst")
	if err := s.prometRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
		return err
	}

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	lastGroupKey := ""
	var groupDug, groupPot, runningSaldo float64
	var lastKonto, lastSifra, lastNaziv string

	for _, entity := range *entities {
		groupKey := entity.Konto + "|" + entity.Sifra

		if groupKey != lastGroupKey {
			if lastGroupKey != "" {
				pSaldo := groupDug - groupPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "konto-total",
					Fields: []string{
						translator.Label("Ukupno za") + ": " + lastKonto + " | " + lastSifra + " | " + lastNaziv,
						common.FormatNumberWithSystemLocale(groupDug, 2),
						common.FormatNumberWithSystemLocale(groupPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			groupDug, groupPot = 0, 0
			runningSaldo = 0
			lastGroupKey = groupKey
			lastKonto = entity.Konto
			lastSifra = entity.Sifra
			lastNaziv = entity.Naziv
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, entity.Sifra, entity.Naziv},
			})
		}

		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Opis,
				entity.OJ,
				entity.Dokum,
				entity.Tra,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	if lastGroupKey != "" {
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto + " | " + lastSifra + " | " + lastNaziv,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	nHeaders := len(tbl.Headers)
	if nHeaders >= 3 {
		grandSaldo := grandDug - grandPot
		tbl.Totals[nHeaders-3] = common.FormatNumberWithSystemLocale(grandDug, 2)
		tbl.Totals[nHeaders-2] = common.FormatNumberWithSystemLocale(grandPot, 2)
		tbl.Totals[nHeaders-1] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
	}

	return nil
}

// GetPrometSubsintetickihKonta retrieves data for subsinteticka konta based on the provided parameters and populates the table data structure.
func (s *PrometResource) GetPrometSubsintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT  konto, tipdok, nalog, danal, coalesce(opis, '') as opis,
			   	CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END as duguje,
			   	CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END as potrazuje,
				idfpro
		FROM fpro`)
	whereRange(b, "konto::numeric", params.OdKonta, params.DoKonta)
	b.AddCondition("danal", params.OdDatuma, ">=").
		AddCondition("danal", params.DoDatuma, "<=")
	// The period of the session and the search text, if provided.
	if err := s.prometRepo.ApplyFilters(ctx, b, "", s.GetSubsintetickihKontaTableFields(), params.SearchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("konto, danal, tipdok, nalog").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Konto,
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT LEFT(konto, 3) as konto, tipdok, nalog,
			   	MAX(danal) as danal,
			   	MAX(coalesce(opis, '')) as opis,
			   	SUM(CASE WHEN kat = 1 OR kat = 2 THEN iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN kat = 3 OR kat = 4 THEN iznos ELSE 0 END) as potrazuje
		FROM fpro`).
		AddLikeBegin("konto", params.OdKonta).
		AddCondition("danal", params.OdDatuma, ">=").
		AddCondition("danal", params.DoDatuma, "<=")
	// The period of the session and the search text, if provided.
	if err := s.prometRepo.ApplyFilters(ctx, b, "", s.GetSintetickihKontaTableFields(), params.SearchText); err != nil {
		return err
	}
	b.GroupByExpr("LEFT(konto, 3)").GroupBy("tipdok", "nalog")

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("tipdok, nalog").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}
	return nil
}

func (s *PrometResource) GetPrometKarticaSintetickihKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	sqlText := ""
	var groupBy []string
	if params.Analitika == "true" {
		sqlText = `SELECT fpro.konto, fpro.sifra,
			   	MAX(COALESCE(fkpl.naziv, '')) as opis,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro
				LEFT JOIN fkpl ON fkpl.konto = fpro.konto AND fkpl.sifra = fpro.sifra AND fkpl.god = fpro.god AND fkpl.kar = fpro.kar`
		groupBy = []string{"fpro.konto", "fpro.sifra"}
	} else {
		sqlText = `SELECT fpro.konto,
			   	MAX(COALESCE(fkpl.naziv, '')) as opis,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro
				LEFT JOIN fkpl ON fkpl.konto = fpro.konto AND fkpl.god = fpro.god AND fkpl.kar = fpro.kar AND fkpl.vkonta = 2`
		groupBy = []string{"fpro.konto"}
	}

	b := querybuilder.New(sqlText).
		AddLikeBegin("fpro.konto", params.OdKonta).
		AddCondition("danal", params.OdDatuma, ">=").
		AddCondition("danal", params.DoDatuma, "<=")
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.prometRepo.ApplyFilters(ctx, b, "fpro", s.GetKarticaSintetikaTableFields(), params.SearchText); err != nil {
		return err
	}
	b.GroupBy(groupBy...)

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderByExpr("fpro.konto::numeric").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Konto,
				"",
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}
			if params.Analitika == "true" {
				fields[1] = entity.Sifra
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}
	return nil
}

func (s *PrometResource) GetPrometSubsintetikaVrd(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam, tipStampe string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Look up konto name from fkpl (the konto is a key of the lookup)
	if tipStampe == common.TipStampePrint {
		bFkpl := querybuilder.New("SELECT konto, naziv FROM fkpl").
			Where("konto = ?", params.OdKonta).
			Where("vkonta = ?", "2")
		if err := s.fkplRepo.ApplyFilters(ctx, bFkpl, "", nil, ""); err != nil {
			return err
		}
		fkplEntities, err := s.fkplRepo.Select(ctx, bFkpl)
		if err != nil {
			return fmt.Errorf("naziv konta: %w", err)
		}
		if fkplEntities != nil && len(*fkplEntities) > 0 {
			tbl.ContentTitle = (*fkplEntities)[0].Naziv
		}
	}
	b := querybuilder.New(`SELECT
				CASE fpro.vkonta
				WHEN '1' THEN 'Analitika'
				WHEN '2' THEN 'Subsintetika'
				ELSE ''
				END as vkonta,
				fpro.vrd,
				SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as duguje,
			   	SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as potrazuje
				FROM fpro`).
		AddLikeBegin("fpro.konto", params.Konto).
		AddCondition("danal", params.OdDatuma, ">=").
		AddCondition("danal", params.DoDatuma, "<=")

	// The period of the session (fpro.god/kar) and, in the preview, the search text.
	search := ""
	if tipStampe == common.TipStampePreview {
		search = params.SearchText
	}
	if err := s.prometRepo.ApplyFilters(ctx, b, "fpro", s.GetSubsintetikaVrdTableFields(), search); err != nil {
		return err
	}
	b.GroupBy("fpro.vrd", "fpro.vkonta")

	// Set total records and pagination
	if tipStampe == common.TipStampePreview && getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Vkonta,
				entity.Vrd,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
			}

			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometKontaAnaliticki(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.PrometParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT f.danal, f.tipdok, concat(f.tipdok,'-',nalog) as nalog, f.idfpro, f.kat, f.iznos, f.kolic,
		       	f.vrd, coalesce(f.dokum, '') as dokum, f.dadok,
				coalesce(f.ojozn,'') as oj, coalesce(f.opis,'') as opis,
				f.rok, f.tra, f.sifval, f.kurs,
		       	f.deviznos, f.cena, f.konto, f.sifra, fk.naziv, f.idfkpl,
				coalesce(f.dokumv, '') as dokumv, f.dadokv, f.travez, coalesce(f.rdokid, 0) as rdokid,
			   	CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
			   	CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
				CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.kolic ELSE 0 END as kolduguje,
				CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.kolic ELSE 0 END as kolpotrazuje
		FROM fpro as f`).
		AddJoin(`LEFT JOIN fkpl fk on fk.idfkpl = f.idfkpl`).
		AddEqual("f.konto", params.Konto)
	whereRange(b, "f.sifra::numeric", params.OdSifre, params.DoSifre)
	b.AddEqual("f.vkonta", 1).
		AddCondition("f.danal", params.OdDatuma, ">=").
		AddCondition("f.danal", params.DoDatuma, "<=")
	// The period of the session (f.god/kar) and the search text, if provided.
	if err := s.prometRepo.ApplyFilters(ctx, b, "f", s.GetKontaAnalitickiTableFields(), params.SearchText); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.prometRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderByExpr("f.sifra::numeric, f.danal, f.tipdok, f.nalog").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	var totDug, totPot, totKolDug, totKolPot float64
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Sifra,
				entity.Naziv,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Dokum,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				entity.Tra,
				entity.Ojozn,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
				entity.Sifval,
				common.FormatNumberWithSystemLocale(entity.Kurs, 2),
				common.FormatNumberWithSystemLocale(entity.Deviznos, 2),
				common.FormatNumberWithSystemLocale(entity.Kolduguje, 2),
				common.FormatNumberWithSystemLocale(entity.Kolpotrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Cena, 2),
				common.FormatNumberWithSystemLocale(entity.Stanje, 2),
				entity.Dokumv,
				entity.Dadokv.Time.Format(common.DateLayout),
				entity.Travez,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
			totDug += entity.Duguje
			totPot += entity.Potrazuje
			totKolDug += entity.Kolduguje
			totKolPot += entity.Kolpotrazuje

		}
	}
	// fill totals
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDug-totPot, 2)
		case "kolduguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolDug, 2)
		case "kolpotrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKolPot, 2)
		}
	}
	return nil
}
func (s *PrometResource) GetPrometAnalitickaKarticaPoMIStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()

	b := querybuilder.New(`SELECT
		f.tipdok, f.nalog, f.danal, f.vrd,
		COALESCE(f.opis, '') as opis,
		COALESCE(f.ojozn, '') as oj,
		f.dokum,
		f.tra,
		f.dadok,
		f.rok,
		f.brst,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(f.sifra, '') as sifra,
		COALESCE(fk.naziv, '') as naziv,
		COALESCE(f.mi::text, '0') as mi
	FROM fpro f`).
		AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl").
		AddEqual("f.vkonta", 1).
		AddCondition("f.konto", params.OdKonta, ">=").
		AddCondition("f.konto", params.DoKonta, "<=")
	whereRange(b, "f.sifra::numeric", params.OdSifre, params.DoSifre)
	b.AddCondition("f.danal", params.OdDatuma, ">=").
		AddCondition("f.danal", params.DoDatuma, "<=")
	whereRange(b, "f.mi::numeric", params.OdMI, params.DoMI)
	b.OrderByExpr("f.konto, f.sifra::numeric, f.mi::numeric, f.danal, f.nalog::numeric, f.brst")
	if err := s.prometRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
		return err
	}

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))
	if entities == nil || len(*entities) == 0 {
		return nil
	}

	monthNames := []string{"", "Januar", "Februar", "Mart", "April", "Maj", "Jun", "Jul", "Avgust", "Septembar", "Oktobar", "Novembar", "Decembar"}

	var grandDug, grandPot float64
	var groupDug, groupPot, runningSaldo float64
	var miDug, miPot float64
	var monthDug, monthPot float64
	lastGroupKey := ""
	lastMIKey := ""
	lastMonthKey := ""
	lastMonthLabel := ""
	var lastKonto, lastSifra, lastNaziv, lastMI string

	appendMonthTotal := func() {
		mSaldo := monthDug - monthPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "month-total",
			Fields: []string{
				lastMonthLabel,
				common.FormatNumberWithSystemLocale(monthDug, 2),
				common.FormatNumberWithSystemLocale(monthPot, 2),
				common.FormatNumberWithSystemLocale(mSaldo, 2),
			},
		})
		monthDug, monthPot = 0, 0
		lastMonthKey = ""
		lastMonthLabel = ""
	}

	appendMiTotal := func() {
		if params.SaldaPoMesecima && lastMonthKey != "" {
			appendMonthTotal()
		}
		miSaldo := miDug - miPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "mi-total",
			Fields: []string{
				translator.Label("Ukupno za mesto isporuke") + ": " + lastMI,
				common.FormatNumberWithSystemLocale(miDug, 2),
				common.FormatNumberWithSystemLocale(miPot, 2),
				common.FormatNumberWithSystemLocale(miSaldo, 2),
			},
		})
		miDug, miPot = 0, 0
		lastMIKey = ""
	}

	appendKontoTotal := func() {
		if lastMIKey != "" {
			appendMiTotal()
		}
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto + " | " + lastSifra + " | " + lastNaziv,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
		groupDug, groupPot = 0, 0
		runningSaldo = 0
		lastGroupKey = ""
	}

	for _, entity := range *entities {
		groupKey := entity.Konto + "|" + entity.Sifra
		miKey := entity.MI
		monthKey := ""
		monthLabel := ""
		if params.SaldaPoMesecima && entity.Danal.Valid {
			month := int(entity.Danal.Time.Month())
			year := entity.Danal.Time.Year()
			monthKey = fmt.Sprintf("%04d-%02d", year, month)
			monthLabel = fmt.Sprintf("%s %d", monthNames[month], year)
		}

		// Close month subtotal if month changed (same MI)
		if params.SaldaPoMesecima && monthKey != lastMonthKey && lastMonthKey != "" {
			appendMonthTotal()
		}

		// Close MI group if MI changed
		if miKey != lastMIKey && lastMIKey != "" {
			appendMiTotal()
		}

		// Close konto group if konto+sifra changed
		if groupKey != lastGroupKey && lastGroupKey != "" {
			appendKontoTotal()
		}

		// Open new konto group
		if groupKey != lastGroupKey {
			lastGroupKey = groupKey
			lastKonto = entity.Konto
			lastSifra = entity.Sifra
			lastNaziv = entity.Naziv
			runningSaldo = 0
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, entity.Sifra, entity.Naziv},
			})
		}

		// Open new MI group
		if miKey != lastMIKey {
			lastMIKey = miKey
			lastMI = miKey
			miDug, miPot = 0, 0
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "mi-header",
				Fields:   []string{miKey},
			})
		}

		// Open new month (track only)
		if params.SaldaPoMesecima && monthKey != lastMonthKey {
			lastMonthKey = monthKey
			lastMonthLabel = monthLabel
			monthDug, monthPot = 0, 0
		}

		// Data row
		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		miDug += dug
		miPot += pot
		monthDug += dug
		monthPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Vrd,
				entity.Opis,
				entity.OJ,
				entity.Dokum,
				entity.Tra,
				entity.Dadok.Time.Format(common.DateLayout),
				entity.Rok,
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	// Close remaining groups
	if lastGroupKey != "" {
		appendKontoTotal()
	}

	// Grand totals
	nHeaders := len(tbl.Headers)
	if nHeaders >= 3 {
		grandSaldo := grandDug - grandPot
		tbl.Totals[nHeaders-3] = common.FormatNumberWithSystemLocale(grandDug, 2)
		tbl.Totals[nHeaders-2] = common.FormatNumberWithSystemLocale(grandPot, 2)
		tbl.Totals[nHeaders-1] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
	}
	return nil
}

// GetPrometKarticaSintKontaStampa generates data for the "Promet Kartica Sintetickog Konta" report based on the provided parameters.
func (s *PrometResource) GetPrometKarticaSintKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()

	// Look up konto name from fkpl (sintetički konto vkonta=3; the konto is a key of the lookup)
	bFkpl := querybuilder.New(`SELECT konto, naziv FROM fkpl`).
		Where("konto = ?", params.OdKonta).
		Where("vkonta = ?", 3)
	if err := s.fkplRepo.ApplyFilters(ctx, bFkpl, "", nil, ""); err != nil {
		return err
	}
	fkplEntities, err := s.fkplRepo.Select(ctx, bFkpl)
	if err != nil {
		return fmt.Errorf("naziv konta: %w", err)
	}
	if fkplEntities != nil && len(*fkplEntities) > 0 {
		tbl.ContentTitle = (*fkplEntities)[0].Naziv
	}

	var b *querybuilder.Builder
	if !params.Analitika {
		// AGGREGATED MODE: one row per konto+sifra combination
		b = querybuilder.New(`SELECT
			f.konto,
			COALESCE(f.sifra, '') as sifra,
			COALESCE(MAX(fk.naziv), '') as opis,
			SUM(CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END) as duguje,
			SUM(CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END) as potrazuje
		FROM fpro f`).
			GroupBy("f.konto", "f.sifra").
			OrderByExpr("f.konto, coalesce(NULLIF(f.sifra, ''), '0')::numeric")
	} else {
		// DETAILED MODE: individual journal entries
		b = querybuilder.New(`SELECT
			f.konto,
			COALESCE(f.sifra, '') as sifra,
			COALESCE(fk.naziv, '') as opis,
			CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
			CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje
		FROM fpro f`).
			OrderByExpr("f.konto, COALESCE(NULLIF(f.sifra, ''), '0')::numeric, f.danal, f.nalog::numeric")
	}
	b.AddJoin("LEFT JOIN fkpl fk ON fk.idfkpl = f.idfkpl").
		AddLikeBegin("f.konto", params.OdKonta).
		AddCondition("f.danal", params.OdDatuma, ">=").
		AddCondition("f.danal", params.DoDatuma, "<=")
	if err := s.prometRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
		return err
	}

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	for i, entity := range *entities {
		dug := entity.Duguje
		pot := entity.Potrazuje
		saldo := dug - pot
		grandDug += dug
		grandPot += pot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", i+1),
				entity.Konto,
				entity.Sifra,
				entity.Opis,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(saldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	for i, header := range tbl.Headers {
		switch header.Name {
		case "opis":
			tbl.Totals[i] = translator.Label("Ukupno za listu")
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
		}
	}
	return nil
}

func (s *PrometResource) GetPrometSintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()

	b := querybuilder.New(`SELECT
		f.tipdok, f.nalog, f.danal,
		COALESCE(f.opis, '') as opis,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(fkpl.naziv, '') as naziv
	FROM fpro f`).
		AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = f.idfkpl").
		AddLikeBegin("f.konto", params.OdKonta).
		AddCondition("f.danal", params.OdDatuma, ">=").
		AddCondition("f.danal", params.DoDatuma, "<=").
		OrderByExpr("f.konto, f.danal, f.tipdok, f.nalog::numeric")
	if err := s.prometRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
		return err
	}

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	var groupDug, groupPot, runningSaldo float64
	var monthDug, monthPot float64
	lastKonto := ""
	var lastNaziv, lastMonth string

	monthKey := func(entity domain.PrometDto) string {
		if !entity.Danal.Valid {
			return ""
		}
		return fmt.Sprintf("%02d/%d", int(entity.Danal.Time.Month()), entity.Danal.Time.Year())
	}

	for _, entity := range *entities {
		currentMonth := monthKey(entity)

		if entity.Konto != lastKonto {
			if lastKonto != "" {
				if params.SaldaPoMesecima && lastMonth != "" {
					tbl.Rows = append(tbl.Rows, domain.TableRow{
						ClassRow: "month-total",
						Fields: []string{
							translator.Label("Ukupno za mesec") + " " + lastMonth,
							common.FormatNumberWithSystemLocale(monthDug, 2),
							common.FormatNumberWithSystemLocale(monthPot, 2),
							common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
						},
					})
				}
				pSaldo := groupDug - groupPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "konto-total",
					Fields: []string{
						translator.Label("Ukupno za") + ": " + lastKonto,
						common.FormatNumberWithSystemLocale(groupDug, 2),
						common.FormatNumberWithSystemLocale(groupPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			groupDug, groupPot, runningSaldo = 0, 0, 0
			monthDug, monthPot = 0, 0
			lastKonto = entity.Konto
			lastNaziv = entity.Naziv
			lastMonth = currentMonth
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, lastNaziv},
			})
		} else if params.SaldaPoMesecima && currentMonth != lastMonth {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
			monthDug, monthPot = 0, 0
			lastMonth = currentMonth
		}

		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		monthDug += dug
		monthPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	if lastKonto != "" {
		if params.SaldaPoMesecima && lastMonth != "" {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
		}
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetPrometSubsintetickihKontaStampa(ctx context.Context, tbl *domain.TableData, params domain.PrometStampaParam) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	translator := i18n.GetInstance()

	b := querybuilder.New(`SELECT
		f.tipdok, f.nalog, f.danal,
		COALESCE(f.opis, '') as opis,
		CASE WHEN f.kat = 1 OR f.kat = 2 THEN f.iznos ELSE 0 END as duguje,
		CASE WHEN f.kat = 3 OR f.kat = 4 THEN f.iznos ELSE 0 END as potrazuje,
		f.konto,
		COALESCE(fkpl.naziv, '') as naziv
	FROM fpro f`).
		AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = f.idfkpl")
	whereRange(b, "f.konto::numeric", params.OdKonta, params.DoKonta)
	b.AddCondition("f.danal", params.OdDatuma, ">=").
		AddCondition("f.danal", params.DoDatuma, "<=").
		OrderByExpr("f.konto, f.danal, f.tipdok, f.nalog::numeric")
	if err := s.prometRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
		return err
	}

	entities, err := s.prometRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))

	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var grandDug, grandPot float64
	var groupDug, groupPot, runningSaldo float64
	var monthDug, monthPot float64
	lastKonto := ""
	var lastNaziv, lastMonth string

	monthKey := func(entity domain.PrometDto) string {
		if !entity.Danal.Valid {
			return ""
		}
		return fmt.Sprintf("%02d/%d", int(entity.Danal.Time.Month()), entity.Danal.Time.Year())
	}

	for _, entity := range *entities {
		currentMonth := monthKey(entity)

		if entity.Konto != lastKonto {
			if lastKonto != "" {
				// Close out last month of previous konto
				if params.SaldaPoMesecima && lastMonth != "" {
					tbl.Rows = append(tbl.Rows, domain.TableRow{
						ClassRow: "month-total",
						Fields: []string{
							translator.Label("Ukupno za mesec") + " " + lastMonth,
							common.FormatNumberWithSystemLocale(monthDug, 2),
							common.FormatNumberWithSystemLocale(monthPot, 2),
							common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
						},
					})
				}
				// Emit konto-total
				pSaldo := groupDug - groupPot
				tbl.Rows = append(tbl.Rows, domain.TableRow{
					ClassRow: "konto-total",
					Fields: []string{
						translator.Label("Ukupno za") + ": " + lastKonto,
						common.FormatNumberWithSystemLocale(groupDug, 2),
						common.FormatNumberWithSystemLocale(groupPot, 2),
						common.FormatNumberWithSystemLocale(pSaldo, 2),
					},
				})
			}
			groupDug, groupPot, runningSaldo = 0, 0, 0
			monthDug, monthPot = 0, 0
			lastKonto = entity.Konto
			lastNaziv = entity.Naziv
			lastMonth = currentMonth
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "konto-header",
				Fields:   []string{entity.Konto, lastNaziv},
			})
		} else if params.SaldaPoMesecima && currentMonth != lastMonth {
			// Month changed within same konto
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
			monthDug, monthPot = 0, 0
			lastMonth = currentMonth
		}

		dug := entity.Duguje
		pot := entity.Potrazuje
		runningSaldo += dug - pot
		groupDug += dug
		groupPot += pot
		monthDug += dug
		monthPot += pot
		grandDug += dug
		grandPot += pot

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Tipdok,
				entity.Nalog,
				entity.Danal.Time.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(dug, 2),
				common.FormatNumberWithSystemLocale(pot, 2),
				common.FormatNumberWithSystemLocale(runningSaldo, 2),
			},
		})
	}

	if lastKonto != "" {
		if params.SaldaPoMesecima && lastMonth != "" {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "month-total",
				Fields: []string{
					translator.Label("Ukupno za mesec") + " " + lastMonth,
					common.FormatNumberWithSystemLocale(monthDug, 2),
					common.FormatNumberWithSystemLocale(monthPot, 2),
					common.FormatNumberWithSystemLocale(monthDug-monthPot, 2),
				},
			})
		}
		pSaldo := groupDug - groupPot
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "konto-total",
			Fields: []string{
				translator.Label("Ukupno za") + ": " + lastKonto,
				common.FormatNumberWithSystemLocale(groupDug, 2),
				common.FormatNumberWithSystemLocale(groupPot, 2),
				common.FormatNumberWithSystemLocale(pSaldo, 2),
			},
		})
	}

	grandSaldo := grandDug - grandPot
	for i, header := range tbl.Headers {
		switch header.Name {
		case "duguje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandDug, 2)
		case "potrazuje":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandPot, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(grandSaldo, 2)
		}
	}

	return nil
}

func (s *PrometResource) GetFieldCache() map[string]reflect.StructField {
	return s.service.GetFieldCache()
}

func (s *PrometResource) GetAnkontaTableFields() []domain.Fields {
	return s.prometAnKontaTableFields
}

func (s *PrometResource) GetAnKontaMiTableFields() []domain.Fields {
	return s.prometAnKontaMiTableFields
}
func (s *PrometResource) GetAnDeviznaKontaTableFields() []domain.Fields {
	return s.prometAnDeviznaKontaTableFields
}
func (s *PrometResource) GetAnDeviznaKontaRekapTableFields() []domain.Fields {
	return s.prometAnDeviznaKontaRekapTableFields
}
func (s *PrometResource) GetSubsintetickihKontaTableFields() []domain.Fields {
	return s.prometSubsintetickihKontaTablefields
}
func (s *PrometResource) GetSintetickihKontaTableFields() []domain.Fields {
	return s.prometSintetickihKontaTableFields
}

func (s *PrometResource) GetKarticaSintetikaTableFields() []domain.Fields {
	return s.prometKarticaSintetickihKontaTableFields
}
func (s *PrometResource) GetSubsintetikaVrdTableFields() []domain.Fields {
	return s.prometSubsintetikaVrdTableFields
}
func (s *PrometResource) GetKontaAnalitickiTableFields() []domain.Fields {
	return s.prometKontaAnalitickiTableFields
}

func (s *PrometResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return repository.GetFvrData(ctx, s.fvrRepo)
}

func (s *PrometResource) GetPrometAnalitickaKarticaStampaTableFields() []domain.Fields {
	return s.prometAnalitickaKarticaStampaTableFields
}

func (s *PrometResource) GetPrometSubsintetickihKontaStampaTableFields() []domain.Fields {
	return s.prometSubsintetickihKontaStampaTableFields
}

func (s *PrometResource) GetPrometSintetickihKontaStampaTableFields() []domain.Fields {
	return s.prometSintetickihKontaStampaTableFields
}

func (s *PrometResource) GetPrometKarticaSintKontaStampaTableFields() []domain.Fields {
	return s.prometKarticaSintKontaStampaTableFields
}

func (s *PrometResource) GetPrometSubsintetikaVrdStampaTableFields() []domain.Fields {
	return s.prometSubsintetikaVrdStampaTableFields
}

func (s *PrometResource) GetPrometAnalitickaKarticaPoMIStampaTableFields() []domain.Fields {
	return s.prometAnalitickaKarticaPoMIStampaTableFields
}

func (s *PrometResource) setServiceFieldValues() {
	s.prometAnKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrd", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "dokum", Label: "Broj dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina Dokumenta", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "Org. Jedinica", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolduguje", Label: "Količina Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolpotrazuje", Label: "Količina Potražuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezni Dokument", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Datum Veznog Dokumenta", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.travez", SkipInSearch: false},
		{Name: "mi", Label: "Mesto Isporuke", Width: "10", Field: "fpro.mi", SkipInSearch: false},
	}

	s.prometAnKontaMiTableFields = []domain.Fields{
		{Name: "mi", Label: "Mesto Isporuke", Width: "10", Field: "fpro.mi", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "VD", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "Dokum", Label: "Dokument", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina", Width: "4", Field: "fpro.tra", SkipInSearch: false},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolduguje", Label: "Količina Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolpotrazuje", Label: "Količina Potražuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezni Dokument", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Datum Veznog Dokumenta", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.travez", SkipInSearch: false},
	}

	s.prometAnDeviznaKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "fpro.kurs", SkipInSearch: true},
		{Name: "duguje", Label: "Devizno Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Devizno Potražuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "Dokum", Label: "Broj Dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: true},
		{Name: "tra", Label: "Godina", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.oj", SkipInSearch: true},
		{Name: "konto", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.konto", SkipInSearch: true},
	}
	s.prometAnDeviznaKontaRekapTableFields = []domain.Fields{
		{Name: "valuta", Label: "Valuta", Width: "30", Field: "fpro.sifval", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devduguje", Label: "Devizno  Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devpotrazuje", Label: "Devizno Potražuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devsaldo", Label: "Devizni Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetickihKontaTablefields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "3", Field: "fpro.konto", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "3", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "6", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "6", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "40", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSintetickihKontaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "4", Field: "fpro.tipdok", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKarticaSintetickihKontaTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10", Field: "fpro.konto", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifra", Label: "Sifra", Width: "10", Field: "fpro.sifra", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetikaVrdTableFields = []domain.Fields{
		{Name: "vkonta", Label: "Vrsta Konta", Width: "10", Field: "fpro.vkonta", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "vrd", Label: "Vrsta knjizenja", Width: "10", Field: "fpro.vrd", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetikaVrdStampaTableFields = []domain.Fields{
		{Name: "vkonta", Label: "Vrsta konta", Width: "15", Field: "vkonta", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "15", Field: "fpro.vrd", SkipInSearch: true, TextAlign: "center"},
		{Name: "duguje", Label: "Duguje", Width: "20", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "20", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "20", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometAnalitickaKarticaPoMIStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "4", Field: "f.vrd", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "30", Field: "f.opis", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "f.ojozn", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "f.dokum", SkipInSearch: true},
		{Name: "tra", Label: "Poslovna godina", Width: "5", Field: "f.tra", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "f.dadok", SkipInSearch: true},
		{Name: "rok", Label: "Rok", Width: "4", Field: "f.rok", SkipInSearch: true},
		{Name: "brst", Label: "Redni broj", Width: "5", Field: "f.brst", SkipInSearch: true, TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKontaAnalitickiTableFields = []domain.Fields{
		{Name: "sifra", Label: "Sifra", Width: "10", Field: "fpro.sifra", SkipInSearch: false, TextAlign: "right", IncludeInTotals: true},
		{Name: "naziv", Label: "Naziv", Width: "30", Field: "fpro.naziv", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "10", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "danal", Label: "Datum Naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "4", Field: "fpro.vrd", SkipInSearch: false},
		{Name: "dokum", Label: "Broj Dokumenta", Width: "6", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "6", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "rok", Label: "Rok", Width: "4", Field: "fpro.rok", SkipInSearch: false},
		{Name: "tra", Label: "Godina Dokumenta", Width: "4", Field: "fpro.tra", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "30", Field: "fpro.opis", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "8", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "8", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "84", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "sifval", Label: "Sifval", Width: "4", Field: "fpro.sifval", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", Field: "kurs", SkipInSearch: true},
		{Name: "deviznos", Label: "Devizni Iznos", Width: "8", Field: "deviznos", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolduguje", Label: "Količina Duguje", Width: "8", Field: "kolduguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolpotrazuje", Label: "Količina Potražuje", Width: "8", Field: "kolpotrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: "Cena", Width: "8", Field: "cena", SkipInSearch: true},
		{Name: "stanje", Label: "Stanje", Width: "8", Field: "stanje", SkipInSearch: true},
		{Name: "dokumv", Label: "Vezni Dokument", Width: "6", Field: "fpro.dokumv", SkipInSearch: false},
		{Name: "dadokv", Label: "Datum Veznog Dokumenta", Width: "6", Field: "fpro.dadokv", SkipInSearch: false},
		{Name: "travez", Label: "God Veznog Dokumenta", Width: "4", Field: "fpro.travez", SkipInSearch: false},
	}
	s.prometAnalitickaKarticaStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "4", Field: "f.vrd", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "30", Field: "f.opis", SkipInSearch: true},
		{Name: "oj", Label: "OJ", Width: "4", Field: "f.ojozn", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "10", Field: "f.dokum", SkipInSearch: true},
		{Name: "tra", Label: "Poslovna godina", Width: "5", Field: "f.tra", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "f.dadok", SkipInSearch: true},
		{Name: "rok", Label: "Rok", Width: "4", Field: "f.rok", SkipInSearch: true},
		{Name: "brst", Label: "Redni broj", Width: "5", Field: "f.brst", SkipInSearch: true, TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSubsintetickihKontaStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "40", Field: "f.opis", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometSintetickihKontaStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "4", Field: "f.tipdok", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "f.nalog", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "f.danal", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "40", Field: "f.opis", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "10", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "10", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "10", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKarticaSintKontaStampaTableFields = []domain.Fields{
		{Name: "rednibr", Label: "Red. Br.", Width: "5", Field: "rednibr", SkipInSearch: true, TextAlign: "center"},
		{Name: "konto", Label: "Konto", Width: "8", Field: "f.konto", SkipInSearch: true},
		{Name: "sifra", Label: "Sifra", Width: "8", Field: "f.sifra", SkipInSearch: true},
		{Name: "opis", Label: "Opis", Width: "40", Field: "opis", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
}

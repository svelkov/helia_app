package finansijsko

import (
	"context"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"regexp"
	"strconv"
	"strings"
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
// fproPartnerKontaExists restricts fpro to the stavke of one partner on one analiticki konto
// (fkpl.idfkpl = ?, the id of the row of the partner grid).
const fproPartnerKontaExists = `EXISTS (
				SELECT 1
				FROM fkpl
				JOIN partneri p
					ON fkpl.tipanalitikeid = p.tipanalitikeid
				WHERE fkpl.idfkpl = fpro.idfkpl
					AND p.sifra = fpro.sifra
					AND p.god = fpro.god AND p.kar = fpro.kar
					AND fkpl.idfkpl = ?
			)`

// formValue returns a value sent by the form, "" when the client sent none or the JavaScript
// "undefined" (an unset field).
func formValue(text string) string {
	if text == "undefined" {
		return ""
	}
	return text
}

// whereDospelaUslovi adds the filters of the dospela potrazivanja: the sifra range (numeric), the
// valuta (dadok + rok) up to the date of the form and the window of brojDana days. The optional
// ones are skipped when empty, like the old AddCondition.
func whereDospelaUslovi(b *querybuilder.Builder, params domain.OtvStavkeParam, brojDana int) {
	if params.OdSifre != "" {
		b.Where("fpro.sifra::numeric >= ?", params.OdSifre)
	}
	if params.DoSifre != "" {
		b.Where("fpro.sifra::numeric <= ?", params.DoSifre)
	}
	if params.PodDatumom != "" {
		b.Where("fpro.dadok + fpro.rok <= ?", params.PodDatumom)
	}
	if params.TipPotrazivanja == "D" {
		b.Where("fpro.dadok + fpro.rok <= ?", time.Now().AddDate(0, 0, brojDana).Format(common.HtmlLayout))
	}
	if params.TipPregleda == "P" {
		b.Where("fpro.dadok + fpro.rok >= ?", time.Now().AddDate(0, 0, -brojDana).Format(common.HtmlLayout))
	}
}

// dospeceBuckets are the aliases of the age buckets, in the order of the dospece parameters.
var dospeceBuckets = [5]string{"dospece15", "dospece30", "dospece60", "dospece90", "dospece120"}

var dospeceDaniRe = regexp.MustCompile(`^\d+$`)

// dospeceDani validates and parses the five dospece parameters (number of days).
func dospeceDani(p domain.OtvStavkeParam) ([5]int, error) {
	var dani [5]int
	for i, v := range [5]string{p.Dospece15, p.Dospece30, p.Dospece60, p.Dospece90, p.Dospece120} {
		if !dospeceDaniRe.MatchString(v) {
			return dani, fmt.Errorf("invalid %s parameter: must be numeric", dospeceBuckets[i])
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return dani, fmt.Errorf("invalid %s parameter: %w", dospeceBuckets[i], err)
		}
		dani[i] = n
	}
	return dani, nil
}

// dospeceKolone builds the columns of the pregled potrazivanja/dugovanja by age: realizacija
// (vrd in fakture), placeno (vrd in placanja), dospelarealizacija (dospeloExpr <= stanjeNaDan)
// and the buckets dospece15 .. dospece120plus by the valuta (dadok + rok) counted in days from
// stanjeNaDan. With sum every column is SUM(...). The SQL is static; every value is an argument
// (date + integer days in PostgreSQL is the same date as the old date + INTERVAL 'N days').
func dospeceKolone(sum bool, fakture, placanja []any, dospeloExpr, stanjeNaDan string, dani [5]int) (string, []any) {
	var cols []string
	var args []any
	add := func(expr, alias string, a ...any) {
		if sum {
			expr = "SUM(" + expr + ")"
		}
		cols = append(cols, expr+" as "+alias)
		args = append(args, a...)
	}
	in := func(n int) string { return strings.TrimSuffix(strings.Repeat("?, ", n), ", ") }
	inFakture := "fpro.vrd IN (" + in(len(fakture)) + ")"
	withFakture := func(a ...any) []any { return append(append([]any{}, fakture...), a...) }
	const valuta = "(fpro.dadok + fpro.rok)::date"

	add("CASE WHEN "+inFakture+" THEN fpro.iznos ELSE 0 END", "realizacija", fakture...)
	add("CASE WHEN fpro.vrd IN ("+in(len(placanja))+") THEN fpro.iznos ELSE 0 END", "placeno", placanja...)
	add("CASE WHEN "+inFakture+" AND fpro.dadok IS NOT NULL AND "+dospeloExpr+" <= ?::date THEN fpro.iznos ELSE 0 END",
		"dospelarealizacija", withFakture(stanjeNaDan)...)
	od := 0
	for i, alias := range dospeceBuckets {
		add("CASE WHEN "+inFakture+" AND "+valuta+" > (?::date + ?::int) AND "+valuta+" <= (?::date + ?::int) THEN fpro.iznos ELSE 0 END",
			alias, withFakture(stanjeNaDan, od, stanjeNaDan, dani[i])...)
		od = dani[i]
	}
	add("CASE WHEN "+inFakture+" AND "+valuta+" > (?::date + ?::int) THEN fpro.iznos ELSE 0 END",
		"dospece120plus", withFakture(stanjeNaDan, od)...)
	return strings.Join(cols, ",\n\t\t\t"), args
}

// wherePregledKontaSifre adds the konto and sifra ranges of the pregled potrazivanja (numeric
// comparison; each is skipped when empty, like the old AddCondition).
func wherePregledKontaSifre(b *querybuilder.Builder, p domain.OtvStavkeParam) {
	if p.OdKonta != "" {
		b.Where("fpro.konto::numeric >= ?", p.OdKonta)
	}
	if p.DoKonta != "" {
		b.Where("fpro.konto::numeric <= ?", p.DoKonta)
	}
	if p.OdSifre != "" {
		b.Where("NULLIF(fpro.sifra, '')::numeric >= ?", p.OdSifre)
	}
	if p.DoSifre != "" {
		b.Where("NULLIF(fpro.sifra, '')::numeric <= ?", p.DoSifre)
	}
}

func (s *OtvoreneStavkeResource) GetOtvoreneStavkePartneri(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.OtvStavkeParam) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl").
		AddEqual("fpro.konto", params.Konto).
		AddCondition("fpro.sifra", params.OdSifre, ">=").
		AddCondition("fpro.sifra", params.DoSifre, "<=")
	if params.PodDatumom != "" {
		b.Where("fpro.dadok + fpro.rok <= ?", params.PodDatumom)
	}
	//b.AddCondition("fpro.otvstavkedana", otvstavkedana, "<=")
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(params.SearchText)); err != nil {
		return err
	}
	// Add GROUP BY
	b.GroupBy("fkpl.idfkpl", "fpro.konto", "fpro.sifra", "fkpl.naziv").
		// Add HAVING clause to filter only partners with both debtor and creditor accounts
		Having(`
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`).
		OrderBy("fpro.sifra")
	if !getTotalRecords {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	// The total-records call reads every row: the totals of the grid are summed from them.
	entities, err := s.fproRepo.Select(ctx, b)
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F'
								  			  when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
								coalesce(fpro.dokum, '') as dokum, fpro.tra, coalesce(fpro.ojozn, '') as ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								coalesce(fpro.opis, '') as opis, fpro.vrd,
								fpro.konto, coalesce(fpro.sifra, '') as sifra
								FROM fpro`).
		Where(fproPartnerKontaExists, id)
	search := ""
	if tipStampe != common.TipStampePrint {
		search = formValue(searchText)
	}
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetOtvoreneStavkeDetaljiFields(), search); err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords && tipStampe != common.TipStampePrint {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fpro.dokum, fpro.dadok")
	if tipStampe != common.TipStampePrint {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl").
		AddIn("fpro.vrd", []any{10, 20, 30, 40}).
		Where("fpro.vkonta = ?", "1").
		AddEqual("fpro.konto", params.Konto).
		AddCondition("fpro.sifra", params.OdSifre, ">=").
		AddCondition("fpro.sifra", params.DoSifre, "<=").
		AddCondition("fpro.dadok", params.OdDatuma, ">=").
		AddCondition("fpro.dadok", params.DoDatuma, "<=")
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(params.SearchText)); err != nil {
		return err
	}
	// Add GROUP BY
	b.GroupBy("fkpl.idfkpl", "fpro.konto", "fpro.sifra", "fkpl.naziv").
		// Add HAVING clause to filter only partners with both debtor and creditor accounts
		Having(`
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) = 0`)
	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fpro.sifra").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT
								fpro.rbr, fpro.konto, coalesce(fpro.sifra, '') as sifra, fpro.dokum, fpro.tra, fpro.ojozn, fpro.tipdok, fpro.nalog, fpro.danal,
								fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								fpro.opis, fpro.vrd, fpro.vkonta, fpro.kat
								FROM fpro`).
		Where(fproPartnerKontaExists, id)
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(searchText)); err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fpro.dokum, fpro.dadok")
	if tipStampe != common.TipStampePrint {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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
	b := querybuilder.New(`SELECT coalesce(p.naziv, '') as naziv,
								coalesce(p.sifra, '') as sifra,
								coalesce(p.adresa, '') as adresa,
								coalesce(p.mesto, '') as mesto,
								coalesce(p.pib, '') as pib,
								coalesce(p.matbr, '') as matbr,
								coalesce(p.pobro, 0) as pobro,
								coalesce(p.telefon, '') as telefon
								FROM partneri p`).
		Where(` EXISTS (
				SELECT 1
				FROM fkpl
				JOIN fpro
					ON fpro.idfkpl = fkpl.idfkpl
				WHERE fkpl.tipanalitikeid = p.tipanalitikeid
					AND p.sifra = fpro.sifra
					AND fpro.god = p.god AND fpro.kar = p.kar
					AND fkpl.god = p.god AND fkpl.kar = p.kar
					AND fkpl.idfkpl = ?
			)`, id)
	if err := s.partneriRepo.ApplyFilters(ctx, b, "p", nil, ""); err != nil {
		return partner, err
	}
	partnerEntity, err := s.partneriRepo.Select(ctx, b)
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT fkpl.idfkpl, fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl").
		AddEqual("fpro.konto", params.Konto)
	// The sifra range compares the numeric value of the sifra (optional, like the other filters).
	if params.OdSifre != "" {
		b.Where("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0) >= ?", params.OdSifre)
	}
	if params.DoSifre != "" {
		b.Where("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0) <= ?", params.DoSifre)
	}
	b.AddCondition("fpro.dadok", params.PodDatumom, "<=")
	//b.AddCondition("fpro.otvstavkedana", params.otvstavkedana, "<=")
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(params.SearchText)); err != nil {
		return err
	}
	// Add GROUP BY
	b.GroupBy("fkpl.idfkpl", "fpro.konto", "fpro.sifra", "fkpl.naziv").
		// Add HAVING clause to filter only partners with both debtor and creditor accounts
		Having(`
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`)
	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderByExpr("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F'
								  			  when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
								fpro.dokum, fpro.tra, fpro.ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								coalesce(fpro.opis, '') as opis,
								fpro.vrd
								FROM fpro`).
		Where("fpro.idfkpl = ?", id)
	search := ""
	if tipStampe != common.TipStampePrint {
		search = formValue(searchText)
	}
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetOtvoreneStavkeDetaljiFields(), search); err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fpro.dokum, fpro.dadok")
	if tipStampe != common.TipStampePrint {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT fkpl.idfkpl,fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike,
			   	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
				SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl").
		AddEqual("fpro.konto", params.Konto)
	whereDospelaUslovi(b, params, brojDanaInt)
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(params.SearchText)); err != nil {
		return err
	}
	// Add GROUP BY
	b.GroupBy("fkpl.idfkpl", "fpro.konto", "fpro.sifra", "fkpl.naziv").
		// Add HAVING clause to filter only partners with both debtor and creditor accounts
		Having(`
       	SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`)
	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderByExpr("fpro.sifra::numeric").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	//if we need to get only total records we check the bool gettotalrecords
	//Get data for the table
	b := querybuilder.New(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F'
								  			  when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
								coalesce(fpro.dokum, '') as dokum, fpro.tra, fpro.ojozn, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
								case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
								case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
								coalesce(fpro.opis, '') as opis, fpro.vrd
								FROM fpro`).
		Where("fpro.idfkpl = ?", id)
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetOtvoreneStavkeDetaljiFields(), formValue(searchText)); err != nil {
		return err
	}
	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy("fpro.dokum, fpro.dadok").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	// Fetch all partners with due items (no pagination) – same filter as GetDospelaPotrazivanjaPartneri
	bP := querybuilder.New(`SELECT fkpl.idfkpl, fpro.konto, coalesce(fpro.sifra, '') as sifra, fkpl.naziv as nazivanalitike,
		SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) as dug,
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) as pot
		FROM fpro`).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl").
		AddEqual("fpro.konto", params.Konto)
	whereDospelaUslovi(bP, params, brojDanaInt)
	if err := s.fproRepo.ApplyFilters(ctx, bP, "fpro", nil, ""); err != nil {
		return nil, err
	}
	bP.GroupBy("fkpl.idfkpl", "fpro.konto", "fpro.sifra", "fkpl.naziv").
		Having(`SUM(CASE WHEN fpro.kat = 1 OR fpro.kat = 2 THEN fpro.iznos ELSE 0 END) -
		SUM(CASE WHEN fpro.kat = 3 OR fpro.kat = 4 THEN fpro.iznos ELSE 0 END) <> 0`).
		OrderByExpr("fpro.sifra::numeric")

	partneriRows, err := s.fproRepo.Select(ctx, bP)
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
		bD := querybuilder.New(`SELECT coalesce(case when fpro.vrd = 10 or fpro.vrd = 20 then 'F'
			when fpro.vrd = 30 or fpro.vrd = 40 then 'U' end, '') as doctype,
			fpro.dokum, fpro.nalog, fpro.danal, fpro.dadok, fpro.rok,
			case when fpro.kat = 1 or fpro.kat = 2 then fpro.iznos else 0 end as dug,
			case when fpro.kat = 3 or fpro.kat = 4 then fpro.iznos else 0 end as pot,
			fpro.vrd
			FROM fpro`).
			Where("fpro.idfkpl = ?", pr.IDFkpl)
		if params.PodDatumom != "" {
			bD.Where("fpro.dadok + fpro.rok <= ?", params.PodDatumom)
		}
		if err := s.fproRepo.ApplyFilters(ctx, bD, "fpro", nil, ""); err != nil {
			return nil, err
		}
		bD.OrderBy("fpro.nalog, fpro.dadok")

		detalji, err := s.fproRepo.Select(ctx, bD)
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

	dani, err := dospeceDani(params)
	if err != nil {
		return err
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

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
	kolone, koloneArgs := dospeceKolone(true, []any{vrd1}, []any{vrd2}, "(fpro.dadok + fpro.rok)::date", params.StanjeNaDan, dani)
	b := querybuilder.New(`SELECT fpro.konto, coalesce(fpro.sifra, '') as sifra, fkpl.naziv as nazivanalitike,
			`+kolone+`
		FROM fpro`, koloneArgs...).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")
	wherePregledKontaSifre(b, params)
	b.AddCondition("fpro.dadok", params.StanjeNaDan, "<=").
		AddIn("fpro.vrd", []any{vrd1, vrd2})
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(params.SearchText)); err != nil {
		return err
	}
	// Add GROUP BY
	b.GroupBy("fpro.konto", "fpro.sifra", "fkpl.naziv")

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderByExpr("NULLIF(fpro.sifra, '')::numeric").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	dani, err := dospeceDani(params)
	if err != nil {
		return nil, err
	}

	vrd1, vrd2 := "", ""
	switch params.TipPregleda {
	case "K":
		vrd1, vrd2 = "10", "30"
	case "D":
		vrd1, vrd2 = "20", "40"
	}

	kolone, koloneArgs := dospeceKolone(true, []any{vrd1}, []any{vrd2}, "(fpro.dadok + fpro.rok)::date", params.StanjeNaDan, dani)
	bP := querybuilder.New(`SELECT fpro.konto, coalesce(fpro.sifra, '') as sifra, fkpl.idfkpl,
			fkpl.naziv as nazivanalitike,
			COALESCE(p.mesto, '') as mesto,
			`+kolone+`
		FROM fpro`, koloneArgs...).
		AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl").
		AddJoin("LEFT JOIN partneri p ON p.idpartneri = fkpl.idpartneri")
	wherePregledKontaSifre(bP, params)
	bP.AddCondition("fpro.dadok", params.StanjeNaDan, "<=").
		AddIn("fpro.vrd", []any{vrd1, vrd2})
	if err := s.fproRepo.ApplyFilters(ctx, bP, "fpro", nil, ""); err != nil {
		return nil, err
	}
	bP.GroupBy("fpro.konto", "fpro.sifra", "fkpl.idfkpl", "fkpl.naziv", "p.mesto").
		OrderByExpr("NULLIF(fpro.sifra, '')::numeric")

	partneriRows, err := s.fproRepo.Select(ctx, bP)
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
		koloneD, koloneDArgs := dospeceKolone(false, []any{vrd1}, []any{vrd2}, "(fpro.dadok + fpro.rok)::date", params.StanjeNaDan, dani)
		bD := querybuilder.New(`SELECT fpro.dokum, fpro.dadok, fpro.rok,
				`+koloneD+`
			FROM fpro`, koloneDArgs...).
			Where("fpro.idfkpl = ?", pr.IDFkpl).
			AddCondition("fpro.dadok", params.StanjeNaDan, "<=").
			AddIn("fpro.vrd", []any{vrd1, vrd2})
		if err := s.fproRepo.ApplyFilters(ctx, bD, "fpro", nil, ""); err != nil {
			return nil, err
		}
		bD.OrderBy("fpro.dadok, fpro.dokum")

		detalji, err := s.fproRepo.Select(ctx, bD)
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

	dani, err := dospeceDani(params)
	if err != nil {
		return err
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)

	// Get data for the table with age bucket categorization
	kolone, koloneArgs := dospeceKolone(true, []any{10, 20}, []any{30, 40}, "fpro.dadok::date", params.StanjeNaDan, dani)
	b := querybuilder.New(`SELECT fpro.konto, fpro.sifra, fkpl.naziv as nazivanalitike,
			`+kolone+`
		FROM fpro`, koloneArgs...).
		AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl").
		// (The vkonta by tipPregleda, K=2 / D=1, is not applied: it was commented out.)
		AddCondition("fpro.konto", params.OdKonta, ">=").
		AddCondition("fpro.konto", params.DoKonta, "<=").
		AddCondition("fpro.sifra", params.OdSifre, ">=").
		AddCondition("fpro.sifra", params.DoSifre, "<=")
	// The period of the session (fpro.god/kar) and the search text, if provided.
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", s.GetPartneriFields(), formValue(params.SearchText)); err != nil {
		return err
	}

	// Add GROUP BY
	b.GroupBy("fpro.konto", "fpro.sifra", "fkpl.naziv")

	// Add HAVING clause to filter partners with non-zero totals
	// b.Having(`SUM(CASE WHEN fpro.vrd IN (10, 20) THEN fpro.iznos ELSE 0 END) <> 0`)

	// Set total records and pagination
	if getTotalRecords {
		total, err := s.fproRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderByExpr("fpro.sifra::numeric").
		Limit(pageSize).
		Offset((currentPage - 1) * pageSize)
	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
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

	dani, err := dospeceDani(params)
	if err != nil {
		return err
	}

	kolone, koloneArgs := dospeceKolone(true, []any{10, 20}, []any{30, 40}, "fpro.dadok::date", params.StanjeNaDan, dani)
	b := querybuilder.New(`SELECT fpro.konto, fpro.sifra, fkpl.idfkpl,
			fkpl.naziv as nazivanalitike,
			COALESCE(p.mesto, '') as mesto,
			`+kolone+`
		FROM fpro`, koloneArgs...).
		AddJoin("LEFT JOIN fkpl ON fkpl.idfkpl = fpro.idfkpl").
		AddJoin("LEFT JOIN partneri p ON p.idpartneri = fkpl.idpartneri").
		AddCondition("fpro.konto", params.OdKonta, ">=").
		AddCondition("fpro.konto", params.DoKonta, "<=").
		AddCondition("fpro.sifra", params.OdSifre, ">=").
		AddCondition("fpro.sifra", params.DoSifre, "<=")
	if err := s.fproRepo.ApplyFilters(ctx, b, "fpro", nil, ""); err != nil {
		return err
	}
	b.GroupBy("fpro.konto", "fpro.sifra", "fkpl.idfkpl", "fkpl.naziv", "p.mesto").
		OrderByExpr("fpro.sifra::numeric")

	entities, err := s.fproRepo.Select(ctx, b)
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
	return repository.GetFvrData(ctx, s.fvrRepo)
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

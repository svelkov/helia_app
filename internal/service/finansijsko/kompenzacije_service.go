package finansijsko

import (
	"context"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// KompenzacijeService defines the interface for operations related to Kompenzacije.
type KompenzacijeService interface {
	GetPregledPartneraTableFields() []domain.Fields
	GetFormiranjeTableFields() []domain.Fields
	GetKompenzacijeTableFields() []domain.Fields
	GetDokumentaTableFields() []domain.Fields
	// Observer pattern: userSession comes from context.Context
	ObradaPredlogKompenzacije(ctx context.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int, searchText string) error
	FormiranjeKompenzacije(ctx context.Context, tbl *domain.TableData, tipObrade int, getTotalRecords bool, currentPage, pageSize int, konto, sifra, stanjeNaDan string, checkDospece bool) error
	GetPregledKompenzacijaList(ctx context.Context, tbl *domain.TableData, odDat, doDat, odDok, doDok, statusFilter string) error
}

// KompenzacijeResource implements the KompenzacijeService interface.
type KompenzacijeResource struct {
	service                    *service.BaseService[domain.KompenzacijeDto]
	fproRepo                   *repository.BaseRepository[domain.Fpro]
	pregledPartneraTableFields []domain.Fields
	formiranjeTableFields      []domain.Fields
	kompenzacijeHdrTableFields []domain.Fields
	kompenzacijeDetTableFields []domain.Fields
	searchFiledsTab1           []string
}

// NewKompenzacijeService creates a new instance of KompenzacijeResource
func NewKompenzacijeService(service *service.BaseService[domain.KompenzacijeDto],
	fproRepo *repository.BaseRepository[domain.Fpro]) *KompenzacijeResource {
	rs := &KompenzacijeResource{
		service:  service,
		fproRepo: fproRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetPregledPartneraTableFields returns the table fields for pregled partnera tab
func (s *KompenzacijeResource) GetPregledPartneraTableFields() []domain.Fields {
	return s.pregledPartneraTableFields
}

// GetFormiranjeTableFields returns the table fields for formiranje tab (dužnik/poverilac tables)
func (s *KompenzacijeResource) GetFormiranjeTableFields() []domain.Fields {
	return s.formiranjeTableFields
}

// GetKompenzacijeTableFields returns the table fields for kompenzacije list
func (s *KompenzacijeResource) GetKompenzacijeTableFields() []domain.Fields {
	return s.kompenzacijeHdrTableFields
}

// GetDokumentaTableFields returns the table fields for dokumenta in kompenzacija
func (s *KompenzacijeResource) GetDokumentaTableFields() []domain.Fields {
	return s.kompenzacijeDetTableFields
}

func (s *KompenzacijeResource) ObradaPredlogKompenzacije(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, searchText string) error {
	// UserSession is now retrieved from context.Context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	// Build single optimized query - group by partner only, one row per partner with both debtor/creditor data
	qb := common.NewQueryBuilder(`
			select 
				coalesce(max(case when fpro.konto like '204%' then fpro.konto else null end), '') as konto_duznika,
				coalesce(max(case when fpro.konto like '204%' then fpro.sifra else null end), '') as sifra_duznika,
				coalesce(max(case when fpro.konto like '435%' then fpro.konto else null end), '') as konto_poverioca,
				coalesce(max(case when fpro.konto like '435%' then fpro.sifra else null end), '') as sifra_poverioca,
				sum(case when fpro.konto like '204%' and fpro.vrd in (10, 20) then fpro.iznos else 0 end) -
				sum(case when fpro.konto like '204%' and fpro.vrd in (30, 40) then fpro.iznos else 0 end) as iznos_dokum_duznika,
				sum(case when fpro.konto like '435%' and fpro.vrd in (10, 20) then fpro.iznos else 0 end) -
				sum(case when fpro.konto like '435%' and fpro.vrd in (30, 40) then fpro.iznos else 0 end) as iznos_dokum_poverioca,
				p.idpartneri,
				p.naziv,
				p.mesto,
				p.adresa,
				coalesce(komp.total_kompd, 0) as kompenzacije_duznik, 
				coalesce(komp.total_kompp, 0) as kompenzacije_poverilac
			from fpro`, true)
	// query totals
	qbTotals := common.NewQueryBuilder(`
	select	
	sum(case when fpro.konto like '204%' and fpro.vrd in (10, 20) then fpro.iznos else 0 end) -
	sum(case when fpro.konto like '204%' and fpro.vrd in (30, 40) then fpro.iznos else 0 end) as iznos_dokum_duznika,
	sum(case when fpro.konto like '435%' and fpro.vrd in (10, 20) then fpro.iznos else 0 end) -
	sum(case when fpro.konto like '435%' and fpro.vrd in (30, 40) then fpro.iznos else 0 end) as iznos_dokum_poverioca,
	sum(coalesce(komp.total_kompd, 0)) as kompenzacije_duznik, 
	sum(coalesce(komp.total_kompp, 0)) as kompenzacije_poverilac
	from fpro`, true)

	qb.AddJoin(` left join fkpl on fpro.idfkpl = fkpl.idfkpl `)
	qb.AddJoin(` left join partneri p on p.idpartneri = fkpl.idpartneri`)
	qb.AddJoin(fmt.Sprintf(` left join (
				select 
					sum(case when kompenzh.konto like '204%%' then iznosdok else 0 end) as total_kompd,
					sum(case when kompenzh.konto like '435%%' then iznosdok else 0 end) as total_kompp
				from kompenzh
				where god = %d and kar = %d and stsdok <= 10
			) komp on 1=1`, userSession.SelectedGod, userSession.SelectedKar))
	qbTotals.AddJoin(` left join fkpl on fpro.idfkpl = fkpl.idfkpl `)
	qbTotals.AddJoin(` left join partneri p on p.idpartneri = fkpl.idpartneri`)
	qbTotals.AddJoin(fmt.Sprintf(` left join (
				select
					sum(case when kompenzh.konto like '204%%' then iznosdok else 0 end) as total_kompd,
					sum(case when kompenzh.konto like '435%%' then iznosdok else 0 end) as total_kompp
				from kompenzh
				where god = %d and kar = %d and stsdok <= 10
			) komp on 1=1`, userSession.SelectedGod, userSession.SelectedKar))

	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
		qbTotals.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
		qbTotals.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	qb.AddCustomCondition("(fpro.konto like '204%' OR fpro.konto like '435%')")
	qb.AddGroupBy("p.idpartneri, p.naziv, p.mesto, p.adresa, total_kompd, total_kompp")
	qb.AddHaving("max(case when fpro.konto like '204%' then fpro.konto end) IS NOT NULL AND max(case when fpro.konto like '435%' then fpro.konto end) IS NOT NULL")
	qb.AddOrderBy("p.naziv")

	qbTotals.AddCustomCondition("(fpro.konto like '204%' OR fpro.konto like '435%')")
	qbTotals.AddHaving("max(case when fpro.konto like '204%' then fpro.konto end) IS NOT NULL AND max(case when fpro.konto like '435%' then fpro.konto end) IS NOT NULL")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	if searchText != "" {
		qb.AddCustomSearchCondition(s.searchFiledsTab1, searchText)
		qbTotals.AddCustomSearchCondition(s.searchFiledsTab1, searchText)
	}
	// Build and execute query
	sqlQuery, args := qb.Build()
	entities, err := s.service.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to query fpro: %s", err.Error())
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	// Set table rows - one row per partner with both debtor and creditor data
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.KontoDuznika,
				entity.SifraDuznika,
				entity.KontoPoverioca,
				entity.SifraPoverioca,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.IznosDokumDuznika, 2),
				common.FormatNumberWithSystemLocale(entity.IznosDokumPoverioca, 2),
				common.FormatNumberWithSystemLocale(entity.KompenzacijeDuznik, 2),
				common.FormatNumberWithSystemLocale(entity.KompenzacijePoverilac, 2),
				entity.Mesto,
				entity.Adresa,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	// Get totals and set in table footer
	sqlTotalsQuery, totalsArgs := qbTotals.Build()
	totalsEntities, err := s.service.GetAllCustom(ctx, sqlTotalsQuery, "", totalsArgs, "", "")
	if err != nil {
		return fmt.Errorf("failed to query totals: %s", err.Error())
	}
	var totDuznik, totPoverilac, totKompDuznik, totKompPoverilac float64
	for _, tot := range *totalsEntities {
		totDuznik += tot.IznosDokumDuznika
		totPoverilac += tot.IznosDokumPoverioca
		totKompDuznik += tot.KompenzacijeDuznik
		totKompPoverilac += tot.KompenzacijePoverilac
	}
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	for i, header := range tbl.Headers {
		switch header.Name {
		case "iznos_dokum_duznika":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDuznik, 2)
		case "iznos_dokum_poverioca":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totPoverilac, 2)
		case "saldo":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totDuznik-totPoverilac, 2)
		case "kompenzacije_duznik":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKompDuznik, 2)
		case "kompenzacije_poverilac":
			tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKompPoverilac, 2)
		}
	}

	return nil
}

// FormiranjeKompenzacije fetches debtor or creditor obligations for kompenzacije formiranje
// ipTip: 1 = debtor obligations, 2 = creditor obligations
func (s *KompenzacijeResource) FormiranjeKompenzacije(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, konto, sifra, stanjeNaDan string, checkDospece bool) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build inner query builder for WITH clause
	innerQb := common.NewQueryBuilder(`
		select 
			fpro.konto,
			fpro.sifra,
			fpro.ojozn,
			fpro.idorgjed,
			fpro.nalog,
			fpro.danal,
			fpro.dadok,
			fpro.rok,
			fpro.vrd,
			fpro.iznos,
			fpro.opis,
			case when fpro.iznos > 0 then fpro.tra else fpro.travez end as norm_tra,
			case when fpro.iznos > 0 then fpro.dokum else fpro.dokumv end as norm_dokum,
			kompd.total_komp
		from fpro
		left join (
			select idfpro, sum(iznosdok) as total_komp 
			from kompenzd
			where exists (
				select 1 from kompenzh 
				where kompenzh.kompenzhid = kompenzd.kompenzhid 
				and kompenzh.stsdok <= 10
			)
			group by idfpro
		) kompd on fpro.idfpro = kompd.idfpro
	`, true)

	// Add WHERE conditions to inner query builder
	innerQb.AddEqual("fpro.vkonta", 1)
	innerQb.AddEqual("fpro.konto", konto)
	innerQb.AddEqual("fpro.sifra", sifra)

	if hasGod {
		innerQb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		innerQb.AddEqual("fpro.kar", userSession.SelectedKar)
	}

	if checkDospece && stanjeNaDan != "" {
		innerQb.AddCustomCondition(fmt.Sprintf("fpro.dadok + (fpro.rok || ' days')::interval <= '%s'::date", stanjeNaDan))
	}
	if stanjeNaDan != "" {
		innerQb.AddCustomCondition(fmt.Sprintf("fpro.dadok <= '%s'::date", stanjeNaDan))
	}
	// Build inner query and get SQL + args
	innerSQL, innerArgs := innerQb.Build()

	// Build outer query using WITH clause
	qb := common.NewQueryBuilder(fmt.Sprintf(`
		with inner_data as (%s)
		select 
			t.konto as konto_duznika,
			t.sifra as sifra_duznika,
			t.norm_dokum as dokum,
			max(t.idorgjed) as idorgjed,
			t.ojozn as oj,
			t.norm_tra as tra,
			max(t.nalog) as nalog,
			max(t.danal) as danal,
			max(t.dadok) as dadok,
			max(t.dadok) + max(t.rok) as dospece,
			max(t.rok) as rok,
			min(t.vrd) as vrd,
			sum(case when t.vrd in (10,20) then t.iznos else 0 end) as faktura,
			sum(case when t.vrd in (30,40) then t.iznos else 0 end) as uplata,
			sum(case when t.vrd in (10,20) then t.iznos else 0 end) - 
			sum(case when t.vrd in (30,40) then t.iznos else 0 end) as saldo,
			max(t.opis) as opis,
			coalesce(max(t.total_komp), 0) as kompenzacije_duznik
		from inner_data t
	`, innerSQL), true)

	// Add inner query parameters and configure outer aggregation
	qb.AddArgs(innerArgs...)
	qb.AddGroupBy("t.konto, t.sifra, t.ojozn, t.norm_tra, t.norm_dokum")
	qb.AddHaving("sum(case when t.vrd in (10,20) then t.iznos else 0 end) - sum(case when t.vrd in (30,40) then t.iznos else 0 end) > 0")
	qb.AddOrderBy("t.norm_dokum")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Build and execute query
	sqlQuery, args := qb.Build()
	fmt.Println(sqlQuery, args)
	entities, err := s.service.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to query fpro: %s", err.Error())
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	// Set table rows - one row per partner with both debtor and creditor data
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%d", entity.Nalog),
				common.FormatNullTime(entity.Dadok, common.HtmlLayout),
				common.FormatNumberWithSystemLocale(entity.Faktura, 2),
				common.FormatNumberWithSystemLocale(entity.Uplata, 2),
				common.FormatNumberWithSystemLocale(entity.Saldo, 2),
				common.FormatNumberWithSystemLocale(entity.KompenzacijeDuznik, 2),
				common.FormatNumberWithSystemLocale(entity.Saldo-entity.KompenzacijeDuznik, 2),
				entity.Dokum,
				entity.Vrd,
				entity.Tra,
				entity.Danal.Format(common.HtmlLayout),
				entity.Rok,
				entity.Dospece,
				entity.Opis,
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: true, ID: fmt.Sprintf("%d", entity.IDPartneri)}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	// // Get totals and set in table footer
	// sqlTotalsQuery, totalsArgs := qbTotals.Build()
	// totalsEntities, err := s.service.GetAllCustom(c, sqlTotalsQuery, "", totalsArgs, "", "")
	// if err != nil {
	// 	return fmt.Errorf("failed to query totals: %s", err.Error())
	// }
	// var totFaktura, totUplata, totKomp, totKompUToku float64
	// for _, tot := range *totalsEntities {
	// 	totFaktura += tot.Faktura
	// 	totUplata += tot.Uplata
	// 	totKomp += tot.KompenzacijeDuznik
	// 	totKompUToku += tot.KompenzacijePoverilac
	// }
	// tbl.Totals = make([]string, len(tbl.Headers))
	// tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
	// for i, header := range tbl.Headers {
	// 	switch header.Name {
	// 	case "faktura":
	// 		tbl.Totals[i] = common.FormatNumberWithSystemLocale(totFaktura, 2)
	// 	case "uplata":
	// 		tbl.Totals[i] = common.FormatNumberWithSystemLocale(totUplata, 2)
	// 	case "saldo":
	// 		tbl.Totals[i] = common.FormatNumberWithSystemLocale(totFaktura-totUplata, 2)
	// 	case "kompenzacije_duznik":
	// 		tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKomp, 2)
	// 	case "kompenzacije_poverilac":
	// 		tbl.Totals[i] = common.FormatNumberWithSystemLocale(totKompUToku, 2)
	// 	}
	// }
	return nil
}

// GetPregledKompenzacijaList fetches the list of kompenzacije for Tab 3 (Pregled)
func (s *KompenzacijeResource) PregledKompenzacije(ctx context.Context, tblHdr, tblDet *domain.TableData, getTotalRecords bool, currentPage, pageSize int, statusDok string, searchText string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tblHdr, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build query
	qb := common.NewQueryBuilder(`
		select 
			kompenzh.god,
			kompenzh.kar,
			kompenzh.konto as konto_poverioca,
			kompenzh.sifra as sifra_poverioca,
			kompenzh.kontod as konto_duznika,
			kompenzh.sifrad as sifra_duznika,
			kompenzh.dokum,
			kompenzh.dadok,
			kompenzh.odglicp as odglicep,
			kompenzh.odglicd as odgliced,
			kompenzh.stsdok,
			fkpl.naziv,
			kompenzh.kompenzhid,
			kompenzh.iznosdok as iznos
		from kompenzh`, true)

	// Add JOIN
	qb.AddJoin("inner join fkpl on fkpl.idfkpl = kompenzh.idfkpl")

	// Add WHERE conditions
	if hasGod {
		qb.AddEqual("kompenzh.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kompenzh.kar", userSession.SelectedKar)
	}

	// Status filter
	if statusDok != "" && statusDok != "svi" {
		var statusValue int
		switch statusDok {
		case "nepotvrdjen":
			statusValue = 0
		case "potvrdjen":
			statusValue = 5
		case "proknjizen":
			statusValue = 10
		default:
			statusValue = 0
		}
		qb.AddEqual("kompenzh.stsdok", statusValue)
	}
	// if search text is not epmty, add search conditions
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.KompenzacijeDto{}))
		qb.AddSearchConditions(s.GetPregledPartneraTableFields(), searchText)
	}
	// Order by
	qb.AddOrderBy("kompenzh.dokum asc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	// Build and execute query
	sqlQuery, args := qb.Build()
	entities, err := s.service.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to query kompenzh: %s", err.Error())
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tblHdr, len(*entities), pageSize)
		return nil
	}
	// Process results
	for _, entity := range *entities {

		// Determine status text
		statusText := ""
		switch entity.Stsdok {
		case 0:
			statusText = "Nepotvrđen"
		case 5:
			statusText = "Potvrđen"
		case 10:
			statusText = "Proknjižen"
		default:
			statusText = fmt.Sprintf("Status %d", entity.Stsdok)
		}

		// Add row to table based on kompenzacijeTableFields structure
		fields := []string{
			entity.KontoPoverioca,
			entity.SifraPoverioca,
			entity.Naziv,
			entity.KontoDuznika,
			entity.SifraDuznika,
			entity.Dokum,
			common.FormatNullTime(entity.Dadok, common.DateLayout),
			common.FormatNumberWithSystemLocale(entity.Iznos, 2), // iznos
			entity.Odglicep,
			entity.Odgliced,
			statusText,
		}
		tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
		tblHdr.Rows = append(tblHdr.Rows, tblRow)
	}

	return nil
}

// setServiceFieldValues initializes all table field definitions
func (s *KompenzacijeResource) setServiceFieldValues() {
	s.searchFiledsTab1 = []string{"p.naziv", "p.mesto", "p.adresa", "fpro.konto", "fpro.sifra"}
	// Tab 1: Pregled partnera za kompenzacije (based on screenshot)
	s.pregledPartneraTableFields = []domain.Fields{
		{Name: "konto_duznika", Label: "Konto dužnika", Width: "10", Field: "konto_duznika", IncludeInTotals: true, TextAlign: "right"},
		{Name: "sifra_duznika", Label: "Šifra Dužnika", Width: "10", Field: "sifra_duznika"},
		{Name: "konto_poverioca", Label: "Konto poverioca", Width: "10", Field: "konto_poverioca"},
		{Name: "sifra_poverioca", Label: "Šifra poverioca", Width: "10", Field: "sifra_poverioca"},
		{Name: "naziv", Label: "Naziv", Width: "40", Field: "p.naziv"},
		{Name: "iznos_dokum_duznika", Label: "Iznos dokum dužnika", Width: "15", Field: "kd.komp_iznos", IncludeInTotals: true, TextAlign: "right"},
		{Name: "iznos_dokum_poverioca", Label: "Iznos dokum poverioca", Width: "15", Field: "kp.komp_iznos", IncludeInTotals: true, TextAlign: "right"},
		{Name: "kompenzacije_duznik", Label: "Kompenzacije dužnik", Width: "15", Field: "kompenzacije_duznik", IncludeInTotals: true, TextAlign: "right"},
		{Name: "kompenzacije_poverilac", Label: "Kompenzacije poverilac", Width: "15", Field: "kompenzacije_poverilac", IncludeInTotals: true, TextAlign: "right"},
		{Name: "mesto", Label: "Mesto", Width: "20", Field: "p.mesto"},
		{Name: "adresa", Label: "Adresa", Width: "30", Field: "p.adresa"},
	}

	// Tab 2: Formiranje kompenzacije (used for both dužnik and poverilac tables)
	s.formiranjeTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj naloga", Width: "8", IncludeInTotals: true, TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokum", Width: "10"},
		{Name: "faktura", Label: "Iznos Faktura", Width: "12", IncludeInTotals: true, TextAlign: "right"},
		{Name: "uplata", Label: "Iznos Uplata", Width: "12", IncludeInTotals: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "12", IncludeInTotals: true, TextAlign: "right"},
		{Name: "kompenz", Label: "Kompenzacije u toku", Width: "12", IncludeInTotals: true, TextAlign: "right"},
		{Name: "iznkomp", Label: "Iznos za kompenziranje", Width: "12", IncludeInTotals: true, TextAlign: "right"},
		{Name: "dokum", Label: "Broj dokum", Width: "10"},
		{Name: "vrd", Label: "Vrsta dokum", Width: "8"},
		{Name: "tra", Label: "Godina Dokum", Width: "10"},
		{Name: "danal", Label: "Datum naloga", Width: "5"},
		{Name: "rok", Label: "Rok", Width: "5"},
		{Name: "dospece", Label: "Datum dospeca", Width: "10"},
		{Name: "opis", Label: "Opis knjizenja", Width: "20"},
	}

	// Tab 3 & 4: Kompenzacije table (used in pregled and knjiženje tabs)
	s.kompenzacijeHdrTableFields = []domain.Fields{
		{Name: "konto_poverioca", Label: "Konto poverioca", Width: "10"},
		{Name: "sifra_poverioca", Label: "Šifra poverioca", Width: "8"},
		{Name: "naziv_duznika", Label: "Naziv poverioca", Width: "25"},
		{Name: "konto_poverioca2", Label: "Konto dužnika", Width: "10"},
		{Name: "sifra_poverioca2", Label: "Šifra dužnika", Width: "8"},
		{Name: "broj_kompenzacije", Label: "Broj kompenzacije", Width: "10"},
		{Name: "datum_kompenzacije", Label: "Datum kompenzacije", Width: "10"},
		{Name: "iznos", Label: "Iznos", Width: "12"},
		{Name: "odgovorno_lice_poverioca", Label: "Odgovorno lice poverioca", Width: "15"},
		{Name: "odgovorno_lice_duznika", Label: "Odgovorno lice dužnika", Width: "15"},
		{Name: "status_dokumenta", Label: "Status dokumenta", Width: "12"},
	}

	// Dokumenta table (used in pregled and knjiženje tabs)
	s.kompenzacijeDetTableFields = []domain.Fields{
		{Name: "broj_dokumenta", Label: "Broj Dokumenta", Width: "15"},
		{Name: "datum_dokumenta", Label: "Datum dokumenta", Width: "12"},
		{Name: "datum_dospeca", Label: "Datum dospeca", Width: "12"},
		{Name: "iznos_dokumenta", Label: "Iznos dokumenta", Width: "15"},
		{Name: "tip_stavke", Label: "Tip stavke dužnik/poverilac", Width: "20"},
	}
}

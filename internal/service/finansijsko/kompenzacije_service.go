package finansijsko

import (
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// KompenzacijeService defines the interface for operations related to Kompenzacije.
type KompenzacijeService interface {
	GetPregledPartneraTableFields() []domain.Fields
	GetFormiranjeTableFields() []domain.Fields
	GetKompenzacijeTableFields() []domain.Fields
	GetDokumentaTableFields() []domain.Fields
	ObradaPredlogKompenzacije(c *gin.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int) error
	FormiranjeKompenzacije(c *gin.Context, tblDuznik, tblPoverilac *domain.TableData, tipObrade int) error
	GetPregledKompenzacijaList(c *gin.Context, tbl *domain.TableData, odDat, doDat, odDok, doDok, statusFilter string) error
}

// KompenzacijeResource implements the KompenzacijeService interface.
type KompenzacijeResource struct {
	service                    *service.BaseService[domain.KompenzacijeDto]
	fproRepo                   *repository.BaseRepository[domain.Fpro]
	pregledPartneraTableFields []domain.Fields
	formiranjeTableFields      []domain.Fields
	kompenzacijeHdrTableFields []domain.Fields
	kompenzacijeDetTableFields []domain.Fields
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

func (s *KompenzacijeResource) ObradaPredlogKompenzacije(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	searchText := c.Query("query")

	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build single optimized query with all JOINs and aggregations
	qb := common.NewQueryBuilder(`
        select 
            max(case when fpro.konto like '204%' then fpro.konto else '' end) as konto_duznika,
            max(case when fpro.konto like '204%' then fpro.sifra else '' end) as sifra_duznika,
            max(case when fpro.konto like '435%' then fpro.konto else '' end) as konto_poverioca,
            max(case when fpro.konto like '435%' then fpro.sifra else '' end) as sifra_poverioca,
            p.naziv as naziv,
            p.mesto as mesto,
			p.adresa as adresa,
            sum(case when fpro.konto like '204%' then 
                (case fpro.vrd when 10 then fpro.iznos when 20 then fpro.iznos else 0 end) - 
                (case fpro.vrd when 30 then fpro.iznos when 40 then fpro.iznos else 0 end)
            else 0 end) as iznos_dokum_duznika,
            sum(case when fpro.konto like '435%' then 
                (case fpro.vrd when 10 then fpro.iznos when 20 then fpro.iznos else 0 end) - 
                (case fpro.vrd when 30 then fpro.iznos when 40 then fpro.iznos else 0 end)
            else 0 end) as iznos_dokum_poverioca,
            coalesce(kd.komp_iznos, 0) as kompenzacije_duznik,
            coalesce(kp.komp_iznos, 0) as kompenzacije_poverilac
        from fpro`)

	// Build JOIN conditions for fkpl and partneri
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")

	// Build kompenzh subquery conditions based on god/kar
	var kompWhereConditions []string
	if hasGod {
		kompWhereConditions = append(kompWhereConditions, fmt.Sprintf("god = %d", userSession.SelectedGod))
	}
	if hasKar {
		kompWhereConditions = append(kompWhereConditions, fmt.Sprintf("kar = %d", userSession.SelectedKar))
	}
	kompWhereConditions = append(kompWhereConditions, "stsdok <= 10")
	kompWhereClause := strings.Join(kompWhereConditions, " and ")

	// LEFT JOIN for debtor compensations (konto like '204%')
	qb.AddJoin(fmt.Sprintf(`left join (
        select konto, sifra, sum(iznosdok) as komp_iznos 
        from kompenzh 
        where %s
        group by konto, sifra
    ) kd on kd.konto = fpro.konto and kd.sifra = fpro.sifra and fpro.konto like '204%%'`, kompWhereClause))

	// LEFT JOIN for creditor compensations (kontod like '435%')
	qb.AddJoin(fmt.Sprintf(`left join (
        select kontod, sifrad, sum(iznosdok) as komp_iznos 
        from kompenzh 
        where %s
        group by kontod, sifrad
    ) kp on kp.kontod = fpro.konto and kp.sifrad = fpro.sifra and fpro.konto like '435%%'`, kompWhereClause))

	// Add WHERE conditions for fpro
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}
	// Add custom condition to filter only accounts starting with 204 or 435
	qb.AddCustomCondition("(fpro.konto like '204%' or fpro.konto like '435%')")
	// if search text is not epmty, add search conditions
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.KompenzacijeDto{}))
		qb.AddSearchConditions(s.GetPregledPartneraTableFields(), searchText)
	}
	// Add GROUP BY
	qb.AddGroupBy("p.naziv, p.mesto, p.adresa, kd.komp_iznos, kp.komp_iznos")

	// Add HAVING clause to filter only partners with both debtor and creditor accounts
	qb.AddHaving(` 
        max(case when fpro.konto like '204%' then fpro.konto else '' end) <> '' and
        max(case when fpro.konto like '435%' then fpro.konto else '' end) <> '' and
        sum(case when fpro.konto like '204%' then 
            (case fpro.vrd when 10 then fpro.iznos when 20 then fpro.iznos else 0 end) - 
            (case fpro.vrd when 30 then fpro.iznos when 40 then fpro.iznos else 0 end)
        else 0 end) > 0 and
        sum(case when fpro.konto like '435%' then 
            (case fpro.vrd when 10 then fpro.iznos when 20 then fpro.iznos else 0 end) - 
            (case fpro.vrd when 30 then fpro.iznos when 40 then fpro.iznos else 0 end)
        else 0 end) > 0`)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	// Build and execute query
	sqlQuery, args := qb.Build()
	entities, err := s.service.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to query fpro: %s", err.Error())
	}
	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	// Set table rows
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

	return nil
}

// FormiranjeKompenzacije fetches debtor or creditor obligations for kompenzacije formiranje
// ipTip: 1 = debtor obligations, 2 = creditor obligations
func (s *KompenzacijeResource) FormiranjeKompenzacije(c *gin.Context, tblDuznik, tblPoverilac *domain.TableData, tipObrade int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	// Get query parameters
	konto := c.Query("konto")
	sifra := c.Query("sifra")
	datumFilter := c.Query("datum_filter")
	checkDueDate := c.Query("check_due_date") == "true"

	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build single optimized query with conditional aggregations and subquery JOIN
	qb := common.NewQueryBuilder(`
		select 
			f.idfpro,
			f.konto,
			f.sifra,
			case 
				when f.iznos > 0 then f.dokum
				else f.dokumv
			end as dokum,
			f.idorgjed as idorgjed,
			f.ojozn,
			case 
				when f.iznos > 0 then f.tra
				else f.travez
			end as tra,
			f.nalog as nalog,
			f.danal as danal,
			f.dadok as dadok,
			f.dadok + f.rok as dospece,
			f.rok as rok,
			case when f.vrd in (10, 20) then f.vrd else 0 end as vrd,
			sum(case when f.vrd in (10, 20) then f.iznos else 0 end) as xfaktura,
			sum(case when f.vrd in (30, 40) then f.iznos else 0 end) as xuplata,
			sum(case when f.vrd in (10, 20) then f.iznos else 0 end) - 
			sum(case when f.vrd in (30, 40) then f.iznos else 0 end) as xsaldo,
			f.brst as brst,
			f.opis as opis,
			coalesce(komp.iznos_kompenz, 0) as xiznkompenz
		from fpro f`)

	// LEFT JOIN subquery for existing compensations
	var kompWhereConditions []string
	if hasGod {
		kompWhereConditions = append(kompWhereConditions, fmt.Sprintf("kh.god = %d", userSession.SelectedGod))
	}
	if hasKar {
		kompWhereConditions = append(kompWhereConditions, fmt.Sprintf("kh.kar = %d", userSession.SelectedKar))
	}
	kompWhereConditions = append(kompWhereConditions, "kh.stsdok <= 10")
	kompWhereClause := strings.Join(kompWhereConditions, " and ")

	qb.AddJoin(fmt.Sprintf(`
		left join (
			select kd.idfpro, sum(kd.iznosdok) as iznos_kompenz
			from kompenzd kd
			join kompenzh kh on kh.kompenzhid = kd.kompenzhid
			where %s
			group by kd.idfpro
		) komp on komp.idfpro = f.idfpro`, kompWhereClause))

	// Add WHERE conditions
	if hasGod {
		qb.AddEqual("f.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", userSession.SelectedKar)
	}
	qb.AddEqual("f.vkonta", 1)
	qb.AddEqual("f.konto", konto)
	qb.AddEqual("f.sifra", sifra)

	// Group by document key components
	qb.AddGroupBy("f.konto, f.sifra, f.ojozn, f.tra, f.dokum, f.travez, f.dokumv, f.idfpro, f.idorgjed, f.nalog, f.danal, f.dadok, f.rok, f.brst, f.opis, f.iznos")

	// Add HAVING to filter out zero balance entries
	havingClause := `
		(sum(case when f.vrd in (10, 20) then f.iznos else 0 end) - 
		 sum(case when f.vrd in (30, 40) then f.iznos else 0 end)) > 0`

	// Add due date filter if enabled
	if checkDueDate && datumFilter != "" {
		havingClause += fmt.Sprintf(" and f.dadok + f.rok <= '%s'", datumFilter)
	}

	qb.AddHaving(havingClause)
	qb.AddOrderBy("f.dadok")

	// Build and execute query
	sqlQuery, args := qb.Build()

	rows, err := s.fproRepo.DB.QueryContext(c, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to query fpro: %s", err.Error())
	}
	defer rows.Close()

	// Process results and populate table
	xUkupno := 0.0

	for rows.Next() {
		var idfpro, idorgjed, tra, rok, vrd, brst int
		var konto, sifra, dokum, ojozn, nalog, danal, dadok, dospece, opis string
		var xfaktura, xuplata, xsaldo, xiznkompenz float64

		err := rows.Scan(&idfpro, &konto, &sifra, &dokum, &idorgjed, &ojozn, &tra,
			&nalog, &danal, &dadok, &dospece, &rok, &vrd, &xfaktura, &xuplata,
			&xsaldo, &brst, &opis, &xiznkompenz)
		if err != nil {
			return fmt.Errorf("failed to scan row: %s", err.Error())
		}

		// Calculate final amount for compensation
		iznZaKomp := xfaktura - xuplata - xiznkompenz
		xUkupno += iznZaKomp

		// Add row to table
		fields := []string{
			nalog,
			dadok,
			common.FormatNumberWithSystemLocale(xfaktura, 2),
			common.FormatNumberWithSystemLocale(xuplata, 2),
			common.FormatNumberWithSystemLocale(xsaldo, 2),
			common.FormatNumberWithSystemLocale(xiznkompenz, 2),
			common.FormatNumberWithSystemLocale(iznZaKomp, 2),
			dokum,
			fmt.Sprintf("%d", vrd),
			fmt.Sprintf("%d", tra),
			danal,
			fmt.Sprintf("%d", rok),
			dospece,
			opis,
		}
		tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: true}
		switch tipObrade {
		case 1:
			tblDuznik.Rows = append(tblDuznik.Rows, tblRow)
		case 2:
			tblPoverilac.Rows = append(tblPoverilac.Rows, tblRow)
		}
	}

	return nil
}

// GetPregledKompenzacijaList fetches the list of kompenzacije for Tab 3 (Pregled)
func (s *KompenzacijeResource) PregledKompenzacije(c *gin.Context, tblHdr, tblDet *domain.TableData, getTotalRecords bool, currentPage, pageSize int) error {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	statusDok := c.Query("status_filter")
	searchText := c.Query("query")

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
		from kompenzh`)

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
	entities, err := s.service.GetAllCustom(c, sqlQuery, "", args, "", "")
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
			entity.Dadok.Format(common.DateLayout),
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
	// Tab 1: Pregled partnera za kompenzacije (based on screenshot)
	s.pregledPartneraTableFields = []domain.Fields{
		{Name: "konto_duznika", Label: "Konto dužnika", Width: "10", Field: "konto_duznika"},
		{Name: "sifra_duznika", Label: "Šifra Dužnika", Width: "10", Field: "sifra_duznika"},
		{Name: "konto_poverioca", Label: "Konto poverioca", Width: "10", Field: "konto_poverioca"},
		{Name: "sifra_poverioca", Label: "Šifra poverioca", Width: "10", Field: "sifra_poverioca"},
		{Name: "naziv", Label: "Naziv", Width: "40", Field: "p.naziv"},
		{Name: "iznos_dokum_duznika", Label: "Iznos dokum. dužnika", Width: "15", Field: "kd.komp_iznos"},
		{Name: "iznos_dokum_poverioca", Label: "Iznos dokum. poverioca", Width: "15", Field: "kp.komp_iznos"},
		{Name: "kompenzacije_duznik", Label: "Kompenzacije dužnik", Width: "15", Field: "kompenzacije_duznik"},
		{Name: "kompenzacije_poverilac", Label: "Kompenzacije poverilac", Width: "15", Field: "kompenzacije_poverilac"},
		{Name: "mesto", Label: "Mesto", Width: "20", Field: "p.mesto"},
		{Name: "adresa", Label: "Adresa", Width: "30", Field: "p.adresa"},
	}

	// Tab 2: Formiranje kompenzacije (used for both dužnik and poverilac tables)
	s.formiranjeTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj naloga", Width: "8"},
		{Name: "dadok", Label: "Datum dokum.", Width: "10"},
		{Name: "faktura", Label: "Iznos Faktura", Width: "12"},
		{Name: "uplata", Label: "Iznos Uplata", Width: "12"},
		{Name: "saldo", Label: "Saldo", Width: "12"},
		{Name: "kompenz", Label: "Kompenzacije u toku", Width: "12"},
		{Name: "iznkomp", Label: "Iznos za kompenz.", Width: "12"},
		{Name: "dokum", Label: "Broj dokum.", Width: "10"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "8"},
		{Name: "tra", Label: "Godina Dokum.", Width: "10"},
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

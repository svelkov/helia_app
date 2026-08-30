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

// DnevnikService defines the interface for operations related to Dnevnik knjizenja.
type DnevnikService interface {
	GetDnevnikKnjizenja(ctx context.Context, tbl *domain.TableData, getTotRecords bool, currentPage int, pageSize int, odDatuma, doDatuma, searchText string, printType string) error
	GetDnevnikTableFields() []domain.Fields
	GetDnevnikStampaTableFields() []domain.Fields
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// DnevnikResource implements the DnevnikService interface.
type DnevnikResource struct {
	service                  *service.BaseService[domain.DnevnikDto]
	fproRepo                 *repository.BaseRepository[domain.Fpro]
	fvrRepo                  *repository.BaseRepository[domain.Fvr]
	dnevnikTableFields       []domain.Fields
	dnevnikStampaTableFields []domain.Fields
}

// NewDnevnikService creates a new instance of DnevnikResource
func NewDnevnikService(service *service.BaseService[domain.DnevnikDto],
	fproRepo *repository.BaseRepository[domain.Fpro], fvrRepo *repository.BaseRepository[domain.Fvr]) *DnevnikResource {
	rs := &DnevnikResource{
		service:  service,
		fproRepo: fproRepo,
		fvrRepo:  fvrRepo,
	}
	rs.setServiceFieldValues()
	return rs
}
func (s *DnevnikResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

// GetDnevnikTableFields returns the table fields for dnevnik knjizenja
func (s *DnevnikResource) GetDnevnikTableFields() []domain.Fields {
	return s.dnevnikTableFields
}

func (s *DnevnikResource) GetDnevnikStampaTableFields() []domain.Fields {
	return s.dnevnikStampaTableFields
}

// GetDnevnikKnjizenja fetches the dnevnik knjizenja records
func (s *DnevnikResource) GetDnevnikKnjizenja(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, odDatuma, doDatuma, searchText string, printType string) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Build query
	qb := common.NewQueryBuilder(`
		select 
			row_number() over (order by fpro.danal, fpro.nalog) as redbr,
			fpro.rbr,
			fpro.danal,
			fpro.tipdok,
			fpro.nalog,
			fpro.konto,
			coalesce(fpro.sifra, '') as sifra,
			case when fpro.sifra <> '' then coalesce(p.naziv, '') else coalesce(fkpl.naziv, '') end as naziv,
			case when fpro.kat in (1, 2) then fpro.iznos else 0 end as duguje,
			case when fpro.kat in (3, 4) then fpro.iznos else 0 end as potrazuje,
			case when fpro.kat in (1, 2) then fpro.iznos else 0 end - 
			case when fpro.kat in (3, 4) then fpro.iznos else 0 end as saldo,
			coalesce(fpro.opis, '') as opis,
			coalesce(fpro.dokum, '') as dokum,
			fpro.dadok,
			concat(coalesce(o.ojozn,''), '-', coalesce(o.naziv,'')) as ojozn,
			concat(coalesce(v.sifval::text,''), '-', coalesce(v.oznval,'')) as sifval,
			case when fpro.kat in (1, 2) then fpro.deviznos else 0 end as devdug,
			case when fpro.kat in (3, 4) then fpro.deviznos else 0 end as devpot
		from fpro`, true)

	// Add JOIN for fkpl
	qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")
	qb.AddJoin("left join partneri p on p.sifra = fpro.sifra and p.tipanalitikeid = fkpl.tipanalitikeid")
	qb.AddJoin("left join orgjed o on o.idorgjed = fpro.idorgjed")
	qb.AddJoin("left join valute v on v.idvalute = fpro.idvalute")

	// Add WHERE conditions
	if hasGod {
		qb.AddEqual("fpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", userSession.SelectedKar)
	}

	// Date range filter
	if odDatuma != "" {
		qb.AddCondition("fpro.danal", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("fpro.danal", doDatuma, "<=")
	}
	// if search text is not epmty, add search conditions
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.DnevnikDto{}))
		qb.AddSearchConditions(s.GetDnevnikTableFields(), searchText)
	}

	// Order by
	qb.AddOrderBy("fpro.danal, fpro.nalog, fpro.rbr")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
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
		// set totals for duguje, potrazuje, saldo, devdug, devpot
		var totalDuguje, totalPotrazuje, totalSaldo, totalDevDug, totalDevPot float64
		for _, entity := range *entities {
			totalDuguje += entity.Duguje
			totalPotrazuje += entity.Potrazuje
			totalSaldo += entity.Saldo
			totalDevDug += entity.Devdug
			totalDevPot += entity.Devpot
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

		for i, header := range tbl.Headers {
			if header.IncludeInTotals {
				switch header.Field {
				case "duguje":
					tbl.Totals[i] = common.FormatNumberWithSystemLocale(totalDuguje, 2)
				case "potrazuje":
					tbl.Totals[i] = common.FormatNumberWithSystemLocale(totalPotrazuje, 2)
				case "saldo":
					tbl.Totals[i] = common.FormatNumberWithSystemLocale(totalSaldo, 2)
				case "devdug":
					tbl.Totals[i] = common.FormatNumberWithSystemLocale(totalDevDug, 2)
				case "devpot":
					tbl.Totals[i] = common.FormatNumberWithSystemLocale(totalDevPot, 2)
				}
			}
		}

		return nil
	}
	rbr := 0
	// Process results and populate table
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			rbr++
			fields := []string{}
			if printType == "O" { //obrada - samo polja potrebna za prikaz u obradi
				fields = []string{
					fmt.Sprintf("%d", rbr),
					entity.Danal.Time.Format(common.DateLayout),
					entity.Tipdok,
					fmt.Sprintf("%d", entity.Nalog),
					entity.Konto,
					entity.Sifra,
					entity.Naziv,
					common.FormatNumberWithSystemLocale(entity.Duguje, 2),
					common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
					common.FormatNumberWithSystemLocale(entity.Saldo, 2),
					entity.Opis,
					entity.Dokum,
					entity.Dadok.Time.Format(common.DateLayout),
					entity.Ojozn,
					entity.Sifval,
					common.FormatNumberWithSystemLocale(entity.Devdug, 2),
					common.FormatNumberWithSystemLocale(entity.Devpot, 2),
				}
			}
			if printType == "S" { //stampa - polja potrebna za štampu
				fields = []string{
					fmt.Sprintf("%d", rbr),
					fmt.Sprintf("%d", entity.Rbr),
					entity.Danal.Time.Format(common.DateLayout),
					entity.Tipdok,
					fmt.Sprintf("%d", entity.Nalog),
					entity.Konto,
					entity.Sifra,
					entity.Naziv,
					common.FormatNumberWithSystemLocale(entity.Duguje, 2),
					common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				}
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// setServiceFieldValues initializes all table field definitions
func (s *DnevnikResource) setServiceFieldValues() {
	s.dnevnikTableFields = []domain.Fields{
		{Name: "rbr", Label: "R. broj", Width: "5", Field: "rbr", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "tipdok", Label: "Tip dok", Width: "5", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "8", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv konta", Width: "25", Field: "fkpl.naziv", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "saldo", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "opis", Label: "OPIS", Width: "20", Field: "fpro.opis", SkipInSearch: false},
		{Name: "dokum", Label: "Dokument", Width: "12", Field: "fpro.dokum", SkipInSearch: false},
		{Name: "dadok", Label: "Datum dokumenta", Width: "10", Field: "fpro.dadok", SkipInSearch: false},
		{Name: "ojozn", Label: "OJ", Width: "5", Field: "fpro.ojozn", SkipInSearch: false},
		{Name: "sifval", Label: "Šifra valute", Width: "5", Field: "fpro.sifval", SkipInSearch: false},
		{Name: "devdug", Label: "Devizno duguje", Width: "12", Field: "devdug", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "devpot", Label: "Devizno potražuje", Width: "12", Field: "devpot", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.dnevnikStampaTableFields = []domain.Fields{
		{Name: "redbr", Label: "Redni broj", Width: "5", Field: "rbr", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rbr", Label: "Red broj stavke u nalogu", Width: "5", Field: "fpro.rbr", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "fpro.danal", SkipInSearch: false},
		{Name: "tipdok", Label: "Vrsta naloga", Width: "5", Field: "fpro.tipdok", SkipInSearch: false},
		{Name: "nalog", Label: "Nalog", Width: "8", Field: "fpro.nalog", SkipInSearch: false},
		{Name: "konto", Label: "Konto", Width: "8", Field: "fpro.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "8", Field: "fpro.sifra", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv konta", Width: "25", SkipInSearch: false},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "duguje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", Field: "potrazuje", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
}

package robno

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"

	"github.com/lib/pq"
)

// RobnoKompodaciService defines the commercial-data reports shown in the
// Robno module. Query implementations can be added independently per tab.
type RobnoKompodaciService interface {
	GetPrikazKarticeKupcaDobavljaca(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams) error
	GetPrikazSaldaKupcaDobavljaca(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams) error
	GetPrikazProdajePoMI(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams) error
	GetPregledRealizacijePoKupcimaArtiklima(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams, string) error
	GetPregledRealizacijePoKupcimaGrupama(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams) error
	GetPregledRealizacijePoArtiklima(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams, string) error
	GetPregledUcescaArtikla(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams, string) error
	GetPregledUcescaGrupeArtikala(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams, string) error
	GetRobneGrupeComboValues(context.Context) ([]domain.ComboItem, error)
	GetPrikazKarticeKupcaDobavljacaTableFields() []domain.Fields
	GetPrikazSaldaKupcaDobavljacaTableFields() []domain.Fields
	GetPrikazProdajePoMITableFields() []domain.Fields
	GetPregledRealizacijePoKupcimaArtiklimaTableFields() []domain.Fields
	GetPregledRealizacijePoKupcimaGrupamaTableFields() []domain.Fields
	GetPregledRealizacijePoArtiklimaTableFields() []domain.Fields
	GetPregledUcescaArtiklaTableFields() []domain.Fields
	GetPregledUcescaGrupeArtikalaTableFields() []domain.Fields
	GetFvrData(context.Context) (domain.Fvr, error)
}

type RobnoKompodaciResource struct {
	partnerRepo                         *repository.BaseRepository[domain.Partneri]
	fproRepo                            *repository.BaseRepository[domain.Fpro]
	rproRepo                            *repository.BaseRepository[domain.Rpro]
	grupaRepo                           *repository.BaseRepository[domain.Rgru]
	fvrRepo                             *repository.BaseRepository[domain.Fvr]
	komPodaciRepo                       *repository.BaseRepository[domain.RobnoKomPodaciDto]
	karticaKupcaDobavljacaFields        []domain.Fields
	saldoKupcaDobavljacaFields          []domain.Fields
	prodajaPoMesecuProdajeFields        []domain.Fields
	realizacijaPoKupcimaArtiklimaFields []domain.Fields
	realizacijaPoKupcimaGrupamaFields   []domain.Fields
	realizacijaPoArtiklimaFields        []domain.Fields
	ucesceArtiklaFields                 []domain.Fields
	ucesceGrupeArtikalaFields           []domain.Fields
}

func NewRobnoKompodaciService(partnerRepo *repository.BaseRepository[domain.Partneri], fproRepo *repository.BaseRepository[domain.Fpro], rproRepo *repository.BaseRepository[domain.Rpro], grupaRepo *repository.BaseRepository[domain.Rgru], fvrRepo *repository.BaseRepository[domain.Fvr], komPodaciRepo *repository.BaseRepository[domain.RobnoKomPodaciDto]) *RobnoKompodaciResource {
	s := &RobnoKompodaciResource{partnerRepo: partnerRepo, fproRepo: fproRepo, rproRepo: rproRepo, grupaRepo: grupaRepo, fvrRepo: fvrRepo, komPodaciRepo: komPodaciRepo}
	s.setTableFields()
	return s
}

func (s *RobnoKompodaciResource) GetPrikazKarticeKupcaDobavljaca(ctx context.Context, tbl *domain.TableData, totalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams) error {
	return nil
}
func (s *RobnoKompodaciResource) GetPrikazSaldaKupcaDobavljaca(ctx context.Context, tbl *domain.TableData, totalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams) error {
	return nil
}
func (s *RobnoKompodaciResource) GetPrikazProdajePoMI(ctx context.Context, tbl *domain.TableData, totalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams) error {
	return nil
}
func (s *RobnoKompodaciResource) GetPregledRealizacijePoKupcimaArtiklima(ctx context.Context, tbl *domain.TableData, totalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams, tipStampe string) error {
	return nil
}
func (s *RobnoKompodaciResource) GetPregledRealizacijePoKupcimaGrupama(ctx context.Context, tbl *domain.TableData, totalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams) error {
	return nil
}

// GetPregledRealizacijePoArtiklima ports the ROB_QRY_REALIZ loop: documents are aggregated
// per article (WinDev aaGRP) and per article/customer/MI (WinDev aaFAKT), so the rows emitted
// are an article header, one row per customer and an article TOTAL row.
func (s *RobnoKompodaciResource) GetPregledRealizacijePoArtiklima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKomPodaciParams, tipStampe string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New("user session not found")
	}
	vrdParam := []any{130, 131, 188, 189}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()

	// WinDev resolved the customer through FKPL -> PARTNERI and the delivery place through FISP
	fkplGodKar, fispGodKar := "", ""
	if hasGod {
		fkplGodKar += " and fkpl.god = rdok.god"
		fispGodKar += " and fisp.god = rdok.god"
	}
	if hasKar {
		fkplGodKar += " and fkpl.kar = rdok.kar"
		fispGodKar += " and fisp.kar = rdok.kar"
	}

	qb := common.NewQueryBuilder(`
		select 
			rsif.sifra,
			max(rsif.naziv) as naziv,
			max(rsif.jm) as jm,
			concat(rdok.fkto, ' ', rdok.fana) as kupac,
			coalesce(rdok.mi, 0) as mi,
			max(coalesce(p.naziv, '') || ', ' || coalesce(p.adresa, '') || ', ' || coalesce(p.mesto, '')) as nazivkupca,
			max(coalesce(
				nullif(trim(coalesce(fisp.naziv, '') || ' ' || coalesce(fisp.mesto, '')), ''),
				coalesce(p.naziv, '') || ', ' || coalesce(p.adresa, '') || ', ' || coalesce(p.mesto, '')
			)) as nazivmesta,
			sum(rpro.kolic) as kolic,
			sum(round(rpro.kolic * rpro.fcena, 2)) as xiznos,
			sum(round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) as xrab,
			sum(round(
				(round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) 
				* rdok.ugrabat / 100, 
				2
			)) as xugrabat,
			sum(round(
				(round(rpro.kolic * rpro.fcena, 2) 
				 - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
				 - round(
					 (round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) 
					 * rdok.ugrabat / 100, 
					 2
				   )) * rdok.pkase / 100,
				2
			)) as xkasa
		from rdok
		inner join rpro on rdok.rdokid = rpro.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid`, true)
	qb.AddJoin("left join fkpl on fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana" + fkplGodKar)
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	qb.AddJoin("left join fisp on fisp.konto = rdok.fkto and fisp.sifra = rdok.fana and fisp.mi = rdok.mi" + fispGodKar)
	if hasGod {
		qb.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	qb.AddIn("rdok.vrd", vrdParam)
	qb.AddCondition("rsif.sifra", params.OdArtikla, ">=")
	qb.AddCondition("rsif.sifra", params.DoArtikla, "<=")
	qb.AddCondition("rdok.dadok", params.OdDatuma, ">=")
	qb.AddCondition("rdok.dadok", params.DoDatuma, "<=")

	if params.OdGrupe == "-" {
		params.OdGrupe = "0"
	}
	if params.DoGrupe == "-" {
		params.DoGrupe = "99999"
	}
	qb.AddCondition("rsif.gru", params.OdGrupe, ">=")
	qb.AddCondition("rsif.gru", params.DoGrupe, "<=")

	// RADIO_TRZISTE: domaće = bez devizne valute, izvoz = sa deviznom valutom
	switch params.Trziste {
	case "1":
		qb.AddCustomCondition("(coalesce(rdok.sifval, 0) = 0)")
	case "2":
		qb.AddCustomCondition("(coalesce(rdok.sifval, 0) > 0)")
	}

	// Add search conditions if search text is provided
	if tipStampe == common.TipStampePreview && params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.RobnoKomPodaciDto{}))
		qb.AddSearchConditions(s.GetPregledRealizacijePoArtiklimaTableFields(), params.SearchText)
	}
	qb.AddGroupBy("rsif.sifra, rdok.fkto, rdok.fana, coalesce(rdok.mi, 0)")
	qb.AddOrderBy("rsif.sifra, rdok.fkto, rdok.fana, coalesce(rdok.mi, 0)")
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.komPodaciRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if len(*entities) == 0 {
		return errors.New("no data found")
	}
	// Set total records and pagination
	if tipStampe == common.TipStampePreview && getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	var gKolic, gBruto, gRabat, gUgrab, gKasa, gNeto float64
	emitTotal := func() {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "group-total",
			Fields: []string{
				"TOTAL:",
				"",
				"",
				"",
				common.FormatNumberWithSystemLocale(gKolic, 2),
				common.FormatNumberWithSystemLocale(gBruto, 2),
				common.FormatNumberWithSystemLocale(gRabat, 2),
				common.FormatNumberWithSystemLocale(gUgrab, 2),
				common.FormatNumberWithSystemLocale(gKasa, 2),
				common.FormatNumberWithSystemLocale(gNeto, 2),
			},
		})
		gKolic, gBruto, gRabat, gUgrab, gKasa, gNeto = 0, 0, 0, 0, 0, 0
	}

	lastSifra := ""
	for _, entity := range *entities {
		if entity.Sifra != lastSifra {
			if lastSifra != "" {
				emitTotal()
			}
			lastSifra = entity.Sifra
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				ClassRow: "group-header",
				Fields:   []string{fmt.Sprintf("%s - %s  JM: %s", entity.Sifra, entity.Naziv, entity.Jm)},
			})
		}

		bruto := entity.XIznos
		neto := entity.XIznos - entity.XRab - entity.XUgrabat - entity.XKasa
		gKolic += entity.Kolic
		gBruto += bruto
		gRabat += entity.XRab
		gUgrab += entity.XUgrabat
		gKasa += entity.XKasa
		gNeto += neto

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Kupac,
				entity.NazivKupca,
				fmt.Sprintf("%d", entity.Mi),
				entity.NazivMesta,
				common.FormatNumberWithSystemLocale(entity.Kolic, 2),
				common.FormatNumberWithSystemLocale(bruto, 2),
				common.FormatNumberWithSystemLocale(entity.XRab, 2),
				common.FormatNumberWithSystemLocale(entity.XUgrabat, 2),
				common.FormatNumberWithSystemLocale(entity.XKasa, 2),
				common.FormatNumberWithSystemLocale(neto, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	if lastSifra != "" {
		emitTotal()
	}
	return nil
}
func (s *RobnoKompodaciResource) GetPregledUcescaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKomPodaciParams, tipStampe string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New("user session not found")
	}
	vrdParam := []any{130, 131, 188, 189}
	var total float64 = 0
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	// Build and execute the total net value query to calculate the total for percentage calculations.
	qbTotal := common.NewQueryBuilder(`select 
	    coalesce(round(sum(round(rpro.kolic * rpro.fcena, 2) 
        - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)) * rdok.pkase / 100, 2)), 2), 0) as total_netvalue
    from rpro `, true)
	qbTotal.AddJoin(` inner join rdok on rdok.rdokid = rpro.rdokid`)
	qbTotal.AddJoin(` inner join rsif on rsif.rsifid = rpro.rsifid`)
	if hasGod {
		qbTotal.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qbTotal.AddEqual("rpro.kar", userSession.SelectedKar)
	}

	qbTotal.AddIn("rdok.vrd", vrdParam)
	qbTotal.AddCondition("rsif.sifra", params.OdArtikla, ">=")
	qbTotal.AddCondition("rsif.sifra", params.DoArtikla, "<=")
	qbTotal.AddCondition("rdok.dadok", params.OdDatuma, ">=")
	qbTotal.AddCondition("rdok.dadok", params.DoDatuma, "<=")

	sqlQuery, args := qbTotal.Build()
	entities, err := s.komPodaciRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if len(*entities) > 0 {
		total = (*entities)[0].TotalNetvalue
	}

	qb := common.NewQueryBuilder(`
		select 
			rsif.sifra,
			max(rsif.naziv) as naziv,
			max(rsif.jm) as jm,
			sum(rpro.kolic) as kolic,
			sum(round(rpro.kolic * rpro.fcena, 2)) as xiznos,
			sum(round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) as xrab,
			sum(round(
				(round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) 
				* rdok.ugrabat / 100, 
				2
			)) as xugrabat,
			sum(round(
				(round(rpro.kolic * rpro.fcena, 2) 
				 - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
				 - round(
					 (round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) 
					 * rdok.ugrabat / 100, 
					 2
				   )) * rdok.pkase / 100,
				2
			)) as xkasa
		from rdok
		inner join rpro on rdok.rdokid = rpro.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid`, true)
	if hasGod {
		qb.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	qb.AddIn("rdok.vrd", vrdParam)
	qb.AddCondition("rsif.sifra", params.OdArtikla, ">=")
	qb.AddCondition("rsif.sifra", params.DoArtikla, "<=")
	qb.AddCondition("rdok.dadok", params.OdDatuma, ">=")
	qb.AddCondition("rdok.dadok", params.DoDatuma, "<=")

	// Add search conditions if search text is provided
	if tipStampe == common.TipStampePreview && params.SearchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.RobnoKomPodaciDto{}))
		qb.AddSearchConditions(s.GetPregledUcescaArtiklaTableFields(), params.SearchText)
	}
	qb.AddGroupBy("rsif.sifra")
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args = qb.Build()
	entities, err = s.komPodaciRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if len(*entities) == 0 {
		return errors.New("no data found")
	}
	// Set total records and pagination
	if tipStampe == common.TipStampePreview && getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	for _, entity := range *entities {
		var procenat float64
		entity.TotalNetvalue = entity.XIznos - entity.XRab - entity.XUgrabat - entity.XKasa
		if total == 0 {
			procenat = 0
		} else {
			procenat = entity.TotalNetvalue / total * 100
		}
		fields := []string{
			entity.Sifra,
			entity.Naziv,
			entity.Jm,
			common.FormatNumberWithSystemLocale(entity.Kolic, 2),
			common.FormatNumberWithSystemLocale(entity.TotalNetvalue, 2),
			common.FormatNumberWithSystemLocale(procenat, 2),
		}

		tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
	return nil

}
func (s *RobnoKompodaciResource) GetPregledUcescaGrupeArtikala(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams, tipStampe string) error {
	vrdParam := []any{130, 131, 188, 189}
	total := 0.0
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	qbTotal := common.NewQueryBuilder(`select 
    round(sum(round(rpro.kolic * rpro.fcena, 2) 
        - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)) * rdok.pkase / 100, 2)), 2) as total_netvalue
    from rpro `, true)
	qbTotal.AddJoin(` inner join rdok on rdok.rdokid = rpro.rdokid`)
	qbTotal.AddJoin(` inner join rsif on rsif.rsifid = rpro.rsifid`)
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	if hasGod {
		qbTotal.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qbTotal.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	if params.OdGrupe == "-" {
		params.OdGrupe = "0"
	}
	if params.DoGrupe == "-" {
		params.DoGrupe = "99999"
	}
	qbTotal.AddIn("rdok.vrd", vrdParam)
	qbTotal.AddCondition("rsif.gru", params.OdGrupe, ">=")
	qbTotal.AddCondition("rsif.gru", params.DoGrupe, "<=")

	sqlQuery, args := qbTotal.Build()
	entities, err := s.komPodaciRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if len(*entities) > 0 {
		total = (*entities)[0].TotalNetvalue
	}
	sqlQuery, args = s.BuildRobnoKomPodaciQueryWithCTE(params, userSession, getTotalRecords, pageSize, page, tipStampe)
	entities, err = s.komPodaciRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			var procenat float64
			if total == 0 {
				procenat = 0
			} else {
				procenat = entity.TotalNetvalue / total * 100
			}
			fields := []string{
				entity.Grupa,
				entity.NazivGrupe,
				common.FormatNumberWithSystemLocale(entity.TotKolic, 2),
				common.FormatNumberWithSystemLocale(entity.TotalNetvalue, 2),
				common.FormatNumberWithSystemLocale(procenat, 2),
			}

			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	return nil
}

// BuildRobnoKomPodaciQueryWithCTE builds a complex CTE query for robno kompodaci reports
func (s *RobnoKompodaciResource) BuildRobnoKomPodaciQueryWithCTE(params domain.RobnoKomPodaciParams, userSession *domain.UserSession, getTotalRecords bool, pageSize, currentPage int, tipStampe string) (string, []any) {
	// Build the CTE query with explicit INNER JOINs
	cteQuery := `
		select 
			rsif.gru,
			coalesce(rgru.naziv, '') as nazivgrupe,
			rpro.kolic,
			round(rpro.kolic * rpro.fcena, 2) as xiznos,
			round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2) as xrab,
			round(
				(round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) 
				* rdok.ugrabat / 100, 
				2
			) as xugrabat,
			round(
				(round(rpro.kolic * rpro.fcena, 2) 
				 - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
				 - round(
					 (round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) 
					 * rdok.ugrabat / 100, 
					 2
				   )) * rdok.pkase / 100,
				2
			) as xkasa
		from rdok
		inner join rpro on rdok.rdokid = rpro.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid
		left join rgru on rgru.rgruid = rsif.rgruid
		where 
			rdok.god = $1
			and rdok.kar = $2
			and rdok.dadok between $3 and $4
			and rdok.vrd = any($5)
			and rsif.gru >= $6
			and rsif.gru <= $7`

	// Create query builder for the main SELECT from CTE
	qb := common.NewQueryBuilder("select gru, nazivgrupe, sum(kolic) as totkolic, round(sum(xiznos - xrab - xugrabat - xkasa), 2) as total_netvalue from row_calculations", false)

	// Add CTE to query builder
	qb.WithCTE("row_calculations", cteQuery)

	// Set the parameters for the CTE
	vrdArray := []int{130, 131, 188, 189} // Default VRD values

	qb.AddArgs(
		userSession.SelectedGod, // $1
		userSession.SelectedKar, // $2
		params.OdDatuma,         // $3
		params.DoDatuma,         // $4
		pq.Array(vrdArray),      // $5
		params.OdGrupe,          // $6
		params.DoGrupe,          // $7
	)
	// Add GROUP BY and ORDER BY
	qb.AddGroupBy("gru, nazivgrupe")
	qb.AddOrderBy("gru")
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	return qb.Build()
}

// Table fields getters for various reports
func (s *RobnoKompodaciResource) GetPrikazKarticeKupcaDobavljacaTableFields() []domain.Fields {
	return s.karticaKupcaDobavljacaFields
}
func (s *RobnoKompodaciResource) GetPrikazSaldaKupcaDobavljacaTableFields() []domain.Fields {
	return s.saldoKupcaDobavljacaFields
}
func (s *RobnoKompodaciResource) GetPrikazProdajePoMITableFields() []domain.Fields {
	return s.prodajaPoMesecuProdajeFields
}
func (s *RobnoKompodaciResource) GetPregledRealizacijePoKupcimaArtiklimaTableFields() []domain.Fields {
	return s.realizacijaPoKupcimaArtiklimaFields
}
func (s *RobnoKompodaciResource) GetPregledRealizacijePoKupcimaGrupamaTableFields() []domain.Fields {
	return s.realizacijaPoKupcimaGrupamaFields
}
func (s *RobnoKompodaciResource) GetPregledRealizacijePoArtiklimaTableFields() []domain.Fields {
	return s.realizacijaPoArtiklimaFields
}
func (s *RobnoKompodaciResource) GetPregledUcescaArtiklaTableFields() []domain.Fields {
	return s.ucesceArtiklaFields
}
func (s *RobnoKompodaciResource) GetPregledUcescaGrupeArtikalaTableFields() []domain.Fields {
	return s.ucesceGrupeArtikalaFields
}

func (s *RobnoKompodaciResource) GetRobneGrupeComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.grupaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select gru, naziv from rgru", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("gru")
	sqlQuery, args := qb.Build()
	entites, err := s.grupaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := []domain.ComboItem{{Key: "-", Value: "-"}}
	for _, entity := range *entites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.Gru),
			Value: fmt.Sprintf("%d - %s", entity.Gru, entity.Naziv),
		})
	}
	return comboItems, nil
}

func (s *RobnoKompodaciResource) setTableFields() {
	text := func(name, label, width string) domain.Fields {
		return domain.Fields{Name: name, Label: label, Width: width}
	}
	numeric := func(name, label, width string) domain.Fields {
		return domain.Fields{Name: name, Label: label, Width: width, TextAlign: "right", IncludeInTotals: true, SkipInSearch: true}
	}

	s.karticaKupcaDobavljacaFields = []domain.Fields{
		text("nalog", "Broj naloga", "8"), text("danal", "Datum naloga", "8"), text("vrd", "VD", "4"), text("dokum", "Broj dokumenta", "8"), text("dadok", "Datum dokumenta", "8"), text("rok", "Rok", "5"), text("tra", "Poslovna godina", "7"), text("oj", "OJ", "5"), text("opis", "Opis", "24"), numeric("duguje", "Duguje", "10"), numeric("potrazuje", "Potražuje", "10"), numeric("saldo", "Saldo", "10"),
	}
	s.saldoKupcaDobavljacaFields = []domain.Fields{
		text("sifra", "Šifra", "8"),
		text("naziv", "Naziv partnera", "24"),
		text("pib", "PIB", "10"),
		text("adresa", "Adresa", "22"),
		text("posta", "Poštanski broj", "8"),
		text("mesto", "Mesto", "16"),
		text("konto", "Konto", "10"), numeric("duguje", "Duguje", "12"), numeric("potrazuje", "Potražuje", "12"), numeric("saldo", "Saldo", "12")}
	s.realizacijaPoKupcimaArtiklimaFields = []domain.Fields{text("kupac", "Kupac", "10"), text("naziv", "Naziv kupca", "26"), text("sifra", "Šifra artikla", "10"), text("opis", "Opis artikla", "30"), text("jm", "JM", "6"), numeric("kolicina", "Količina", "12"), numeric("bruto", "Bruto", "12"), numeric("rabat", "Rabat", "10"), numeric("ugrab", "Ugovoreni rabat", "12"), numeric("kasa", "Kasa", "10"), numeric("neto", "Neto iznos bez PDV", "14")}
	s.realizacijaPoKupcimaGrupamaFields = []domain.Fields{
		{Name: "kupac", Label: "Kupac", Width: "10", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv kupca", Width: "26", SkipInSearch: false},
		{Name: "grupa", Label: "Grupa", Width: "8", SkipInSearch: false},
		{Name: "opis", Label: "Naziv grupe", Width: "26", SkipInSearch: false},
		{Name: "kolicina", Label: "Količina", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "bruto", Label: "Bruto", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "rabat", Label: "Rabat", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "ugrab", Label: "Ugovoreni rabat", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "kasa", Label: "Kasa", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "neto", Label: "Neto iznos bez PDV", Width: "14", SkipInSearch: true, TextAlign: "right"},
	}
	s.realizacijaPoArtiklimaFields = []domain.Fields{
		{Name: "kupac", Label: "Kupac", Field: "concat(rdok.fkto, ' ', rdok.fana)", Width: "12", SkipInSearch: false},
		{Name: "nazivkupca", Label: "Naziv kupca", Field: "p.naziv", Width: "40", SkipInSearch: false},
		{Name: "mi", Label: "Mesto isporuke", Width: "8", SkipInSearch: true, TextAlign: "center"},
		{Name: "nazivmesta", Label: "Naziv mesta isporuke", Width: "40", SkipInSearch: true},
		{Name: "kolicina", Label: "Količina", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "bruto", Label: "Bruto", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "rabat", Label: "Rabat", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "ugrab", Label: "Ugovoreni rabat", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "kasa", Label: "Kasa", Width: "10", SkipInSearch: true, TextAlign: "right"},
		{Name: "neto", Label: "Neto iznos bez PDV", Width: "14", SkipInSearch: true, TextAlign: "right"},
	}
	s.ucesceArtiklaFields = []domain.Fields{
		{Name: "sifra", Label: "Šifra artikla", Field: "rsif.sifra", Width: "10", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv artikla", Field: "rsif.naziv", Width: "38", SkipInSearch: false},
		{Name: "jm", Label: "JM", Field: "rsif.jm", Width: "6", SkipInSearch: false},
		{Name: "kolicina", Label: "Količina", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "neto", Label: "Neto iznos bez PDV", Width: "16", SkipInSearch: true, TextAlign: "right"},
		{Name: "procenat", Label: "Procenat učešća", Width: "12", SkipInSearch: true, TextAlign: "right"},
	}
	s.ucesceGrupeArtikalaFields = []domain.Fields{
		{Name: "grupa", Label: "Šifra grupe", Width: "10", SkipInSearch: false},
		{Name: "opis", Label: "Naziv grupe", Width: "42", SkipInSearch: false},
		{Name: "kolicina", Label: "Količina", Width: "12", SkipInSearch: true, TextAlign: "right"},
		{Name: "neto", Label: "Neto iznos bez PDV", Width: "18", SkipInSearch: true, TextAlign: "right"},
		{Name: "procenat", Label: "Procenat učešća", Width: "14", SkipInSearch: true, TextAlign: "right"},
	}
}

// GetFvrData retrieves company (fvr) data for the current session.
func (s *RobnoKompodaciResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

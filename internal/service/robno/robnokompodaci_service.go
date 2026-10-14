package robno

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"

	"github.com/lib/pq"
)

// RobnoKompodaciService defines the commercial-data reports shown in the
// Robno module. Query implementations can be added independently per tab.
type RobnoKompodaciService interface {
	GetPregledRealizacijePoKupcimaArtiklima(context.Context, *domain.TableData, bool, int, int, domain.RobnoKomPodaciParams, string) error
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
	commonSvc                           commonsvc.CommonService
	karticaKupcaDobavljacaFields        []domain.Fields
	saldoKupcaDobavljacaFields          []domain.Fields
	prodajaPoMesecuProdajeFields        []domain.Fields
	realizacijaPoKupcimaArtiklimaFields []domain.Fields
	realizacijaPoKupcimaGrupamaFields   []domain.Fields
	realizacijaPoArtiklimaFields        []domain.Fields
	ucesceArtiklaFields                 []domain.Fields
	ucesceGrupeArtikalaFields           []domain.Fields
}

func NewRobnoKompodaciService(partnerRepo *repository.BaseRepository[domain.Partneri], fproRepo *repository.BaseRepository[domain.Fpro], rproRepo *repository.BaseRepository[domain.Rpro], grupaRepo *repository.BaseRepository[domain.Rgru], fvrRepo *repository.BaseRepository[domain.Fvr], komPodaciRepo *repository.BaseRepository[domain.RobnoKomPodaciDto], commonSvc commonsvc.CommonService) *RobnoKompodaciResource {
	s := &RobnoKompodaciResource{partnerRepo: partnerRepo, fproRepo: fproRepo, rproRepo: rproRepo, grupaRepo: grupaRepo, fvrRepo: fvrRepo, komPodaciRepo: komPodaciRepo, commonSvc: commonSvc}
	s.setTableFields()
	return s
}

func (s *RobnoKompodaciResource) GetPregledRealizacijePoKupcimaArtiklima(ctx context.Context, tbl *domain.TableData, totalRecords bool, pageSize, page int, params domain.RobnoKomPodaciParams, tipStampe string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New("user session not found")
	}
	vrdParam := []any{130, 131, 188, 189}
	common.SetupTablePagination(tbl, page, pageSize)
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	fkplGodKar := ""
	if hasGod {
		fkplGodKar = " and fkpl.god = rdok.god"
	}
	if hasKar {
		fkplGodKar += " and fkpl.kar = rdok.kar"
	}
	customerSelect := "concat(rdok.fkto, ' ', rdok.fana) as kupac"
	customerNameSelect := "max(coalesce(p.naziv, '')) as nazivkupca"
	groupBy := []string{"rdok.fkto", "rdok.fana", "rsif.sifra"}
	if params.TipIzvestaja == "1" {
		// ObradaPOARTIKLU aggregates aaFAKT by article, not by customer.
		customerSelect = "'' as kupac"
		customerNameSelect = "'' as nazivkupca"
		groupBy = []string{"rsif.sifra"}
	}

	b := querybuilder.New(`
		select
			` + customerSelect + `,
			` + customerNameSelect + `,
			rsif.sifra,
			max(rsif.naziv) as naziv,
			max(rsif.jm) as jm,
			sum(rpro.kolic) as kolic,
			sum(round(rpro.kolic * rpro.fcena, 2)) as xiznos,
			sum(round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) as xrab,
			sum(round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)) as xugrabat,
			sum(round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2) - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)) * rdok.pkase / 100, 2)) as xkasa
			,sum(round(case when rpro.ncena > 0 then rpro.kolic * rpro.ncena else rpro.kolic * rpro.cena end, 2)) as nabiznos
		from rdok
		inner join rpro on rdok.rdokid = rpro.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid`).
		AddJoin("left join fkpl on fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana" + fkplGodKar).
		AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	if hasGod {
		b.Where("rpro.god = ?", userSession.SelectedGod)
	}
	if hasKar {
		b.Where("rpro.kar = ?", userSession.SelectedKar)
	}
	b.AddIn("rdok.vrd", vrdParam).
		AddCondition("rsif.sifra", params.OdArtikla, ">=").
		AddCondition("rsif.sifra", params.DoArtikla, "<=").
		AddCondition("rdok.dadok", params.OdDatuma, ">=").
		AddCondition("rdok.dadok", params.DoDatuma, "<=")

	if params.OdGrupe == "-" {
		params.OdGrupe = "0"
	}
	if params.DoGrupe == "-" {
		params.DoGrupe = "99999"
	}
	b.AddCondition("rsif.gru", params.OdGrupe, ">=").
		AddCondition("rsif.gru", params.DoGrupe, "<=")
	switch params.Trziste {
	case "1":
		b.Where("coalesce(rdok.sifval, 0) = 0")
	case "2":
		b.Where("coalesce(rdok.sifval, 0) > 0")
	}
	if tipStampe == common.TipStampePreview {
		// RobnoKomPodaciDto has no god/kar, so ApplyFilters only adds the search.
		if err := s.komPodaciRepo.ApplyFilters(ctx, b, "", s.GetPregledRealizacijePoKupcimaArtiklimaTableFields(), params.SearchText); err != nil {
			return err
		}
	}
	b.GroupBy(groupBy...)

	if tipStampe == common.TipStampePreview && totalRecords {
		total, err := s.komPodaciRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		if total == 0 {
			return errors.New("no data found")
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	b.OrderBy(strings.Join(groupBy, ", "))
	if tipStampe == common.TipStampePreview && !totalRecords {
		b.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	entities, err := s.komPodaciRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	if entities == nil || len(*entities) == 0 {
		return errors.New("no data found")
	}

	for _, entity := range *entities {
		neto := entity.XIznos - entity.XRab - entity.XUgrabat - entity.XKasa
		tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: []string{
			entity.Kupac,
			entity.NazivKupca,
			entity.Sifra,
			entity.Naziv,
			entity.Jm,
			common.FormatNumberWithSystemLocale(entity.Kolic, 2),
			common.FormatNumberWithSystemLocale(entity.XIznos, 2),
			common.FormatNumberWithSystemLocale(entity.XRab, 2),
			common.FormatNumberWithSystemLocale(entity.XUgrabat, 2),
			common.FormatNumberWithSystemLocale(entity.XKasa, 2),
			common.FormatNumberWithSystemLocale(neto, 2),
		}})
	}
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

	fkplGodKar, fispGodKar := "", ""
	if hasGod {
		fkplGodKar += " and fkpl.god = rdok.god"
		fispGodKar += " and fisp.god = rdok.god"
	}
	if hasKar {
		fkplGodKar += " and fkpl.kar = rdok.kar"
		fispGodKar += " and fisp.kar = rdok.kar"
	}

	b := querybuilder.New(`
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
		inner join rsif on rsif.rsifid = rpro.rsifid`).
		AddJoin("left join fkpl on fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana" + fkplGodKar).
		AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri").
		AddJoin("left join fisp on fisp.konto = rdok.fkto and fisp.sifra = rdok.fana and fisp.mi = rdok.mi" + fispGodKar)
	if hasGod {
		b.Where("rpro.god = ?", userSession.SelectedGod)
	}
	if hasKar {
		b.Where("rpro.kar = ?", userSession.SelectedKar)
	}
	b.AddIn("rdok.vrd", vrdParam).
		AddCondition("rsif.sifra", params.OdArtikla, ">=").
		AddCondition("rsif.sifra", params.DoArtikla, "<=").
		AddCondition("rdok.dadok", params.OdDatuma, ">=").
		AddCondition("rdok.dadok", params.DoDatuma, "<=")

	if params.OdGrupe == "-" {
		params.OdGrupe = "0"
	}
	if params.DoGrupe == "-" {
		params.DoGrupe = "99999"
	}
	b.AddCondition("rsif.gru", params.OdGrupe, ">=").
		AddCondition("rsif.gru", params.DoGrupe, "<=")

	// RADIO_TRZISTE: domaće = bez devizne valute, izvoz = sa deviznom valutom
	switch params.Trziste {
	case "1":
		b.Where("coalesce(rdok.sifval, 0) = 0")
	case "2":
		b.Where("coalesce(rdok.sifval, 0) > 0")
	}

	// Add search conditions if search text is provided (RobnoKomPodaciDto has no god/kar, so
	// ApplyFilters only adds the search).
	if tipStampe == common.TipStampePreview {
		if err := s.komPodaciRepo.ApplyFilters(ctx, b, "", s.GetPregledRealizacijePoArtiklimaTableFields(), params.SearchText); err != nil {
			return err
		}
	}
	b.GroupBy("rsif.sifra", "rdok.fkto", "rdok.fana").GroupByExpr("coalesce(rdok.mi, 0)").
		OrderBy("rsif.sifra, rdok.fkto, rdok.fana").OrderByExpr("coalesce(rdok.mi, 0)")
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	entities, err := s.komPodaciRepo.Select(ctx, b)
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
	bTotal := querybuilder.New(`select
	    coalesce(round(sum(round(rpro.kolic * rpro.fcena, 2)
        - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)) * rdok.pkase / 100, 2)), 2), 0) as total_netvalue
    from rpro`).
		AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid").
		AddJoin("inner join rsif on rsif.rsifid = rpro.rsifid")
	if hasGod {
		bTotal.Where("rpro.god = ?", userSession.SelectedGod)
	}
	if hasKar {
		bTotal.Where("rpro.kar = ?", userSession.SelectedKar)
	}
	bTotal.AddIn("rdok.vrd", vrdParam).
		AddCondition("rsif.sifra", params.OdArtikla, ">=").
		AddCondition("rsif.sifra", params.DoArtikla, "<=").
		AddCondition("rdok.dadok", params.OdDatuma, ">=").
		AddCondition("rdok.dadok", params.DoDatuma, "<=")

	entities, err := s.komPodaciRepo.Select(ctx, bTotal)
	if err != nil {
		return err
	}
	if len(*entities) > 0 {
		total = (*entities)[0].TotalNetvalue
	}

	b := querybuilder.New(`
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
		inner join rsif on rsif.rsifid = rpro.rsifid`)
	if hasGod {
		b.Where("rpro.god = ?", userSession.SelectedGod)
	}
	if hasKar {
		b.Where("rpro.kar = ?", userSession.SelectedKar)
	}
	b.AddIn("rdok.vrd", vrdParam).
		AddCondition("rsif.sifra", params.OdArtikla, ">=").
		AddCondition("rsif.sifra", params.DoArtikla, "<=").
		AddCondition("rdok.dadok", params.OdDatuma, ">=").
		AddCondition("rdok.dadok", params.DoDatuma, "<=")

	// Add search conditions if search text is provided (RobnoKomPodaciDto has no god/kar, so
	// ApplyFilters only adds the search).
	if tipStampe == common.TipStampePreview {
		if err := s.komPodaciRepo.ApplyFilters(ctx, b, "", s.GetPregledUcescaArtiklaTableFields(), params.SearchText); err != nil {
			return err
		}
	}
	b.GroupBy("rsif.sifra")
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	entities, err = s.komPodaciRepo.Select(ctx, b)
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
	bTotal := querybuilder.New(`select
    round(sum(round(rpro.kolic * rpro.fcena, 2)
        - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)
        - round((round(rpro.kolic * rpro.fcena, 2) - round(rpro.kolic * rpro.fcena * rpro.rab / 100, 2)) * rdok.ugrabat / 100, 2)) * rdok.pkase / 100, 2)), 2) as total_netvalue
    from rpro`).
		AddJoin("inner join rdok on rdok.rdokid = rpro.rdokid").
		AddJoin("inner join rsif on rsif.rsifid = rpro.rsifid")
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	if hasGod {
		bTotal.Where("rpro.god = ?", userSession.SelectedGod)
	}
	if hasKar {
		bTotal.Where("rpro.kar = ?", userSession.SelectedKar)
	}
	if params.OdGrupe == "-" {
		params.OdGrupe = "0"
	}
	if params.DoGrupe == "-" {
		params.DoGrupe = "99999"
	}
	bTotal.AddIn("rdok.vrd", vrdParam).
		AddCondition("rsif.gru", params.OdGrupe, ">=").
		AddCondition("rsif.gru", params.DoGrupe, "<=")

	entities, err := s.komPodaciRepo.Select(ctx, bTotal)
	if err != nil {
		return err
	}
	if len(*entities) > 0 {
		total = (*entities)[0].TotalNetvalue
	}
	entities, err = s.komPodaciRepo.Select(ctx, s.BuildRobnoKomPodaciQueryWithCTE(params, userSession, getTotalRecords, pageSize, page, tipStampe))
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
func (s *RobnoKompodaciResource) BuildRobnoKomPodaciQueryWithCTE(params domain.RobnoKomPodaciParams, userSession *domain.UserSession, getTotalRecords bool, pageSize, currentPage int, tipStampe string) *querybuilder.Builder {
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
			rdok.god = ?
			and rdok.kar = ?
			and rdok.dadok between ? and ?
			and rdok.vrd = any(?)
			and rsif.gru >= ?
			and rsif.gru <= ?`

	vrdArray := []int{130, 131, 188, 189} // Default VRD values

	// The main SELECT from the CTE; the arguments of the CTE are numbered first.
	b := querybuilder.New("select gru, nazivgrupe, sum(kolic) as totkolic, round(sum(xiznos - xrab - xugrabat - xkasa), 2) as total_netvalue from row_calculations").
		WithCTE("row_calculations", cteQuery,
			userSession.SelectedGod,
			userSession.SelectedKar,
			params.OdDatuma,
			params.DoDatuma,
			pq.Array(vrdArray),
			params.OdGrupe,
			params.DoGrupe,
		).
		GroupBy("gru", "nazivgrupe").
		OrderBy("gru")
	if tipStampe == common.TipStampePreview && !getTotalRecords {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	return b
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

// GetRobneGrupeComboValues returns the robne grupe of the current period (CommonService).
func (s *RobnoKompodaciResource) GetRobneGrupeComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetRobneGrupeComboValues(ctx, commonsvc.WithEmptyOption())
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
	return repository.GetFvrData(ctx, s.fvrRepo)
}

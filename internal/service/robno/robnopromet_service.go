package robno

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

type RobnoPrometService interface {
	GetPrometArtikala(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometPoKupcima(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetNabavkeOdDobavljaca(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometRucLagerLista(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometRucUlazIzlaz(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometRucMagacini(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometRucFakture(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometGradilista(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetPrometGradilisteVpcNc(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	GetRobneGrupeComboValues(context.Context) ([]domain.ComboItem, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetPrometArtiklaTableFields() []domain.Fields
	GetPrometKupcaTableFields() []domain.Fields
	GetPrometRucTableFields() []domain.Fields
	GetPrometOdDobavljacaTableFields() []domain.Fields
	GetPrometGradilistaTableFields() []domain.Fields
	GetPrometGradilisteVpcNcTableFields() []domain.Fields
	GetPrometRucLagerListaTableFields() []domain.Fields
	GetPrometRucUlazIzlazTableFields() []domain.Fields
	GetPrometRucMagaciniTableFields() []domain.Fields
	GetPrometRucFaktureTableFields() []domain.Fields
}

type RobnoPrometResource struct {
	rproRepo                         *repository.BaseRepository[domain.Rpro]
	magRepo                          *repository.BaseRepository[domain.Magacini]
	grupaRepo                        *repository.BaseRepository[domain.Rgru]
	fvrRepo                          *repository.BaseRepository[domain.Fvr]
	robnoPrometRepo                  *repository.BaseRepository[domain.RobnoPrometDto]
	commonSvc                        commonsvc.CommonService
	prometGrupeArtikalaTableFields   []domain.Fields
	prometKupciTableFields           []domain.Fields
	prometDobavljaciTableFields      []domain.Fields
	prometRucTableFields             []domain.Fields
	prometGradilisteTableFields      []domain.Fields
	prometRucLagerListaTableFields   []domain.Fields
	prometRucUlazIzlazTableFields    []domain.Fields
	prometRucMagaciniTableFields     []domain.Fields
	prometRucFaktureTableFields      []domain.Fields
	prometGradilisteVpcNcTableFields []domain.Fields
}

func NewRobnoPrometService(rproRepo *repository.BaseRepository[domain.Rpro], magRepo *repository.BaseRepository[domain.Magacini], grupaRepo *repository.BaseRepository[domain.Rgru], fvrRepo *repository.BaseRepository[domain.Fvr], robnoPrometRepo *repository.BaseRepository[domain.RobnoPrometDto], commonSvc commonsvc.CommonService) *RobnoPrometResource {
	rs := &RobnoPrometResource{
		rproRepo:                         rproRepo,
		magRepo:                          magRepo,
		grupaRepo:                        grupaRepo,
		fvrRepo:                          fvrRepo,
		robnoPrometRepo:                  robnoPrometRepo,
		commonSvc:                        commonSvc,
		prometGrupeArtikalaTableFields:   []domain.Fields{},
		prometKupciTableFields:           []domain.Fields{},
		prometDobavljaciTableFields:      []domain.Fields{},
		prometRucTableFields:             []domain.Fields{},
		prometGradilisteTableFields:      []domain.Fields{},
		prometRucLagerListaTableFields:   []domain.Fields{},
		prometRucUlazIzlazTableFields:    []domain.Fields{},
		prometRucMagaciniTableFields:     []domain.Fields{},
		prometRucFaktureTableFields:      []domain.Fields{},
		prometGradilisteVpcNcTableFields: []domain.Fields{},
	}
	rs.setTableFields()
	return rs
}

// GetPrometArtikala renders "Promet artikala po grupama za period" (robno promet, tab 1): one row
func (s *RobnoPrometResource) GetPrometArtikala(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	isPrint := printType == common.TipStampePrint
	tbl.Headers = robnoPrometHeaders(s.prometGrupeArtikalaTableFields, isPrint)
	tbl.HasTotals = true

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select rpro.sifra as sifra_art,
		coalesce(min(rsif.gru), 0) as gru,
		coalesce(min(rpro.konto), '') as konto,
		coalesce(min(rsif.naziv), '') as naziv,
		coalesce(min(rsif.jm), '') as jm,
		coalesce(sum(case when upper(dokvrsta.kodknj) = 'D' and coalesce(dokvrsta.grpdok, '') <> 'NIV' then rpro.kolic else 0 end), 0) as ulaz,
		coalesce(sum(case when upper(dokvrsta.kodknj) = 'P' and coalesce(dokvrsta.grpdok, '') <> 'NIV' then rpro.kolic else 0 end), 0) as izlaz,
		coalesce(sum(case when upper(dokvrsta.kodknj) = 'D' then case when dokvrsta.grpdok = 'NIV' then rpro.kolic * (rpro.fcena - rpro.ncena) else rpro.kolic * rpro.cena end else 0 end), 0) as finulaz,
		coalesce(sum(case when upper(dokvrsta.kodknj) = 'P' then case when dokvrsta.grpdok = 'NIV' then rpro.kolic * (rpro.fcena - rpro.ncena) else rpro.kolic * rpro.cena end else 0 end), 0) as finizlaz
	from rdok`, true)
	qb.AddJoin(" inner join rpro on rpro.rdokid = rdok.rdokid")
	qb.AddJoin(" inner join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(" left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	qb.AddCondition("rdok.mag", params.OdMagacina, ">=")
	qb.AddCondition("rdok.mag", params.DoMagacina, "<=")
	qb.AddCondition("rpro.sifra", params.OdSifreArtikla, ">=")
	qb.AddCondition("rpro.sifra", params.DoSifreArtikla, "<=")
	qb.AddCondition("rdok.dadok", params.OdDatuma, ">=")
	qb.AddCondition("rdok.dadok", params.DoDatuma, "<=")
	qb.AddCondition("rsif.gru", params.OdGrupe, ">=")
	qb.AddCondition("rsif.gru", params.DoGrupe, "<=")
	qb.AddGroupBy("rpro.sifra")
	qb.AddOrderBy("rpro.sifra")
	if !getTotalRecords && !isPrint && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// The print loads the whole result in one call: its totals and its rows.
	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		}
		// The first pass runs without LIMIT, so the totals cover the whole result; only the
		// columns flagged with IncludeInTotals (ulaz, izlaz, finulaz, finizlaz) are filled, the
		// others stay empty.
		var ulaz, izlaz, finulaz, finizlaz float64
		for _, entity := range *entities {
			ulaz += entity.Ulaz
			izlaz += entity.Izlaz
			finulaz += entity.Finulaz
			finizlaz += entity.Finizlaz
		}
		ukupno := map[string]string{
			"ulaz":     common.FormatNumberWithSystemLocale(ulaz, 3),
			"izlaz":    common.FormatNumberWithSystemLocale(izlaz, 3),
			"finulaz":  common.FormatNumberWithSystemLocale(finulaz, 2),
			"finizlaz": common.FormatNumberWithSystemLocale(finizlaz, 2),
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok {
				tbl.Totals[i] = value
			}
		}
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: robnoPrometRowFields(isPrint,
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				fmt.Sprintf("%d", entity.Gru),
				fmt.Sprintf("%d", entity.SifraArt),
				entity.Konto,
				entity.Naziv,
				entity.JM,
				common.FormatNumberWithSystemLocale(entity.Ulaz, 3),
				common.FormatNumberWithSystemLocale(entity.Izlaz, 3),
				common.FormatNumberWithSystemLocale(entity.Finulaz, 2),
				common.FormatNumberWithSystemLocale(entity.Finizlaz, 2),
			),
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometPoKupcima renders "Promet artikala po kupcima za period" (robno promet, tab 2): one row
func (s *RobnoPrometResource) GetPrometPoKupcima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	isPrint := printType == common.TipStampePrint
	tbl.Headers = robnoPrometHeaders(s.prometKupciTableFields, isPrint)
	tbl.HasTotals = true

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`with xiz as (
		select rdok.fkto as konto, rdok.fana as sifra,
			coalesce(cp.naziv, 'NEPOZNAT PARTNER!') as naziv,
			coalesce(cp.adresa, '') as adresa,
			coalesce(cp.mesto, '') as mesto,
			round(coalesce(rpro.kolic, 0) * coalesce(rpro.fcena, 0), 2) as xiznos,
			coalesce(rpro.rab, 0) as rab,
			coalesce(rdok.ugrabat, 0) as ugrabat,
			coalesce(rdok.pkase, 0) as pkase,
			coalesce(rp.pp, 0) as pstopa,
			rdok.god as god, rdok.kar as kar, rdok.mag as mag, rdok.dadok as dadok,
			rpro.sifra as sifra_art, coalesce(rsif.gru, 0) as gru
		from rdok
		inner join rpro on rpro.rdokid = rdok.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid
		inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd and upper(dokvrsta.kodknj) = 'P'
		left join lateral (select r.pp from rpor r
			where r.po = rpro.po and r.datum <= rdok.dadok
			order by r.datum desc limit 1) rp on true
		left join (select distinct on (f.god, f.kar, f.konto, f.sifra)
				f.god, f.kar, f.konto, f.sifra, p.naziv, p.adresa, p.mesto
			from fkpl f
			left join partneri p on p.idpartneri = f.idpartneri
			order by f.god, f.kar, f.konto, f.sifra, f.idfkpl) cp
			on cp.god = rdok.god and cp.kar = rdok.kar and cp.konto = rdok.fkto and cp.sifra = rdok.fana
	), rab as (
		select xiz.*, round(xiz.xiznos * xiz.rab / 100, 2) as vrabat from xiz
	), ugr as (
		select rab.*, round((rab.xiznos - rab.vrabat) * rab.ugrabat / 100, 2) as vrugrabat from rab
	), kas as (
		select ugr.*, round((ugr.xiznos - ugr.vrabat - ugr.vrugrabat) * ugr.pkase / 100, 2) as vrkasa from ugr
	), net as (
		select kas.*, (kas.xiznos - kas.vrabat - kas.vrugrabat - kas.vrkasa) as netbezpdv from kas
	)
	select konto, sifra, naziv, adresa, mesto,
		sum(xiznos) as iznos,
		sum(vrugrabat) as ugrabat,
		sum(vrabat) as rabat,
		sum(vrkasa) as kasa,
		sum(netbezpdv) as netrezpdv,
		sum(netbezpdv + round(netbezpdv * pstopa / 100, 2)) as neto
	from net`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddCondition("mag", params.OdMagacina, ">=")
	qb.AddCondition("mag", params.DoMagacina, "<=")
	qb.AddCondition("sifra_art", params.OdSifreArtikla, ">=")
	qb.AddCondition("sifra_art", params.DoSifreArtikla, "<=")
	qb.AddCondition("dadok", params.OdDatuma, ">=")
	qb.AddCondition("dadok", params.DoDatuma, "<=")
	qb.AddCondition("gru", params.OdGrupe, ">=")
	qb.AddCondition("gru", params.DoGrupe, "<=")
	qb.AddGroupBy("konto, sifra, naziv, adresa, mesto")
	qb.AddOrderBy("konto, sifra")
	if !getTotalRecords && !isPrint && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// The print loads the whole result in one call: its totals and its rows.
	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		}
		// The first pass runs without LIMIT, so the totals cover the whole result; only the
		// columns flagged with IncludeInTotals are filled, the others stay empty.
		var iznos, ugrabat, rabat, kasa, netrezpdv, neto float64
		for _, entity := range *entities {
			iznos += entity.Iznos
			ugrabat += entity.Ugrabat
			rabat += entity.Rabat
			kasa += entity.Kasa
			netrezpdv += entity.Netrezpdv
			neto += entity.Neto
		}
		ukupno := map[string]string{
			"iznos":     common.FormatNumberWithSystemLocale(iznos, 2),
			"ugrabat":   common.FormatNumberWithSystemLocale(ugrabat, 2),
			"rabat":     common.FormatNumberWithSystemLocale(rabat, 2),
			"kasa":      common.FormatNumberWithSystemLocale(kasa, 2),
			"netrezpdv": common.FormatNumberWithSystemLocale(netrezpdv, 2),
			"neto":      common.FormatNumberWithSystemLocale(neto, 2),
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok {
				tbl.Totals[i] = value
			}
		}
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: robnoPrometRowFields(isPrint,
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				entity.Adresa,
				entity.Mesto,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				common.FormatNumberWithSystemLocale(entity.Ugrabat, 2),
				common.FormatNumberWithSystemLocale(entity.Rabat, 2),
				common.FormatNumberWithSystemLocale(entity.Kasa, 2),
				common.FormatNumberWithSystemLocale(entity.Netrezpdv, 2),
				common.FormatNumberWithSystemLocale(entity.Neto, 2),
			),
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetNabavkeOdDobavljaca renders "Nabavka po dobavljačima" (robno promet, tab 3):
func (s *RobnoPrometResource) GetNabavkeOdDobavljaca(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	isPrint := printType == common.TipStampePrint
	tbl.Headers = robnoPrometHeaders(s.prometDobavljaciTableFields, isPrint)
	tbl.HasTotals = true

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`with xiz as (
		select rdok.fkto as konto, rdok.fana as sifra,
			coalesce(cp.naziv, 'NEPOZNAT PARTNER!') as naziv,
			coalesce(cp.adresa, '') as adresa,
			coalesce(cp.mesto, '') as mesto,
			round(coalesce(rpro.kolic, 0) * coalesce(rpro.fcena, 0), 2) as fakvred,
			coalesce(rpro.rab, 0) as rab,
			round(coalesce(rpro.kolic, 0) * coalesce(rpro.ncena, 0), 2) as nabvred,
			round(coalesce(rpro.kolic, 0) * coalesce(rpro.vpcena, 0), 2) as vpvred,
			rdok.god as god, rdok.kar as kar, rdok.mag as mag, rdok.dadok as dadok,
			rpro.sifra as sifra_art, coalesce(rsif.gru, 0) as gru
		from rdok
		inner join rpro on rpro.rdokid = rdok.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid
		inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd and upper(dokvrsta.kodknj) = 'D'
		left join (select distinct on (f.god, f.kar, f.konto, f.sifra)
				f.god, f.kar, f.konto, f.sifra, p.naziv, p.adresa, p.mesto
			from fkpl f
			left join partneri p on p.idpartneri = f.idpartneri
			order by f.god, f.kar, f.konto, f.sifra, f.idfkpl) cp
			on cp.god = rdok.god and cp.kar = rdok.kar and cp.konto = rdok.fkto and cp.sifra = rdok.fana
	), rab as (
		select xiz.*, round(xiz.fakvred * xiz.rab / 100, 2) as rabat from xiz
	)
	select konto, sifra, naziv, adresa, mesto,
		sum(fakvred) as fakvred,
		sum(rabat) as rabat,
		sum(fakvred - rabat) as netofakvred,
		sum(nabvred) as nabvred,
		sum(nabvred - (fakvred - rabat)) as ztrovred,
		sum(vpvred - nabvred) as ruc,
		case when sum(nabvred) <> 0 then round(sum(vpvred - nabvred) / sum(nabvred) * 100, 2) else 0 end as procruc,
		sum(vpvred) as vpvred
	from rab`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddCondition("mag", params.OdMagacina, ">=")
	qb.AddCondition("mag", params.DoMagacina, "<=")
	qb.AddCondition("sifra_art", params.OdSifreArtikla, ">=")
	qb.AddCondition("sifra_art", params.DoSifreArtikla, "<=")
	qb.AddCondition("dadok", params.OdDatuma, ">=")
	qb.AddCondition("dadok", params.DoDatuma, "<=")
	qb.AddCondition("gru", params.OdGrupe, ">=")
	qb.AddCondition("gru", params.DoGrupe, "<=")
	qb.AddGroupBy("konto, sifra, naziv, adresa, mesto")
	qb.AddOrderBy("konto, sifra")
	if !getTotalRecords && !isPrint && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// The print loads the whole result in one call: its totals and its rows.
	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		}
		// The first pass runs without LIMIT, so the totals cover the whole result; only the
		// columns flagged with IncludeInTotals are filled (procruc is a percentage, not a sum), the
		// others stay empty.
		var fakvred, rabat, netofakvred, ztrovred, nabvred, ruc, vpvred float64
		for _, entity := range *entities {
			fakvred += entity.Fakvred
			rabat += entity.Rabat
			netofakvred += entity.Netofakvred
			ztrovred += entity.Ztrovred
			nabvred += entity.Nabvred
			ruc += entity.Ruc
			vpvred += entity.Vpvred
		}
		ukupno := map[string]string{
			"fakvred":     common.FormatNumberWithSystemLocale(fakvred, 2),
			"rabat":       common.FormatNumberWithSystemLocale(rabat, 2),
			"netofakvred": common.FormatNumberWithSystemLocale(netofakvred, 2),
			"ztrovred":    common.FormatNumberWithSystemLocale(ztrovred, 2),
			"nabvred":     common.FormatNumberWithSystemLocale(nabvred, 2),
			"ruc":         common.FormatNumberWithSystemLocale(ruc, 2),
			"vpvred":      common.FormatNumberWithSystemLocale(vpvred, 2),
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok {
				tbl.Totals[i] = value
			}
		}
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: robnoPrometRowFields(isPrint,
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				entity.Adresa,
				entity.Mesto,
				common.FormatNumberWithSystemLocale(entity.Fakvred, 2),
				common.FormatNumberWithSystemLocale(entity.Rabat, 2),
				common.FormatNumberWithSystemLocale(entity.Netofakvred, 2),
				common.FormatNumberWithSystemLocale(entity.Ztrovred, 2),
				common.FormatNumberWithSystemLocale(entity.Nabvred, 2),
				common.FormatNumberWithSystemLocale(entity.Ruc, 2),
				common.FormatNumberWithSystemLocale(entity.Procruc, 2),
				common.FormatNumberWithSystemLocale(entity.Vpvred, 2),
			),
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometGradilista renders "Izveštaj zaduženja gradilišta" (robno promet, tab 4):
func (s *RobnoPrometResource) GetPrometGradilista(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	isPrint := printType == common.TipStampePrint
	tbl.Headers = robnoPrometHeaders(s.prometGradilisteTableFields, isPrint)
	tbl.HasTotals = true

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`with xiz as (
		select coalesce(rdok.pkto, '') as konto, coalesce(rdok.pana, '') as sifra,
			coalesce(cp.naziv, '') as naziv,
			coalesce(rpro.iznos, 0) as iznos,
			rdok.god as god, rdok.kar as kar, rdok.mag as mag, rdok.dadok as dadok,
			rpro.sifra as sifra_art, coalesce(rsif.gru, 0) as gru
		from rdok
		inner join rpro on rpro.rdokid = rdok.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid
		inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd and upper(dokvrsta.grpdok) = 'GRD'
		left join (select distinct on (f.god, f.kar, f.konto, f.sifra)
				f.god, f.kar, f.konto, f.sifra, f.naziv
			from fkpl f
			order by f.god, f.kar, f.konto, f.sifra, f.idfkpl) cp
			on cp.god = rdok.god and cp.kar = rdok.kar and cp.konto = rdok.pkto and cp.sifra = rdok.pana
	)
	select konto, sifra, naziv, sum(iznos) as iznos
	from xiz`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddCondition("mag", params.OdMagacina, ">=")
	qb.AddCondition("mag", params.DoMagacina, "<=")
	qb.AddCondition("sifra_art", params.OdSifreArtikla, ">=")
	qb.AddCondition("sifra_art", params.DoSifreArtikla, "<=")
	qb.AddCondition("dadok", params.OdDatuma, ">=")
	qb.AddCondition("dadok", params.DoDatuma, "<=")
	qb.AddCondition("gru", params.OdGrupe, ">=")
	qb.AddCondition("gru", params.DoGrupe, "<=")
	qb.AddGroupBy("konto, sifra, naziv")
	qb.AddOrderBy("konto, sifra")
	if !getTotalRecords && !isPrint && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// The print loads the whole result in one call: its totals and its rows.
	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		}
		// The first pass runs without LIMIT, so the total covers the whole result; only the column
		// flagged with IncludeInTotals (iznos) is filled, the others stay empty.
		var iznos float64
		for _, entity := range *entities {
			iznos += entity.Iznos
		}
		ukupno := map[string]string{
			"iznos": common.FormatNumberWithSystemLocale(iznos, 2),
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok {
				tbl.Totals[i] = value
			}
		}
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: robnoPrometRowFields(isPrint,
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
			),
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometGradilisteVpcNc renders "Izveštaj zaduženja gradilišta VPC-NC" (robno promet, tab 5):
func (s *RobnoPrometResource) GetPrometGradilisteVpcNc(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	isPrint := printType == common.TipStampePrint
	tbl.Headers = robnoPrometHeaders(s.prometGradilisteVpcNcTableFields, isPrint)
	tbl.HasTotals = true

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`with xiz as (
		select coalesce(rdok.pkto, '') as konto, coalesce(rdok.pana, '') as sifra,
			coalesce(cp.naziv, '') as naziv,
			coalesce(rpro.kolic, 0) * coalesce(rpro.cena, 0) as vpiznos,
			coalesce(rpro.kolic, 0) * coalesce(rpro.ncena, 0) as nciznos,
			rdok.god as god, rdok.kar as kar, rdok.mag as mag, rdok.dadok as dadok,
			rpro.sifra as sifra_art, coalesce(rsif.gru, 0) as gru
		from rdok
		inner join rpro on rpro.rdokid = rdok.rdokid
		inner join rsif on rsif.rsifid = rpro.rsifid
		inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd and upper(dokvrsta.grpdok) = 'GRD'
		left join (select distinct on (f.god, f.kar, f.konto, f.sifra)
				f.god, f.kar, f.konto, f.sifra, f.naziv
			from fkpl f
			order by f.god, f.kar, f.konto, f.sifra, f.idfkpl) cp
			on cp.god = rdok.god and cp.kar = rdok.kar and cp.konto = rdok.pkto and cp.sifra = rdok.pana
	)
	select konto, sifra, naziv,
		sum(vpiznos) as vpiznos,
		sum(nciznos) as nciznos,
		sum(vpiznos) - sum(nciznos) as razlika
	from xiz`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddCondition("mag", params.OdMagacina, ">=")
	qb.AddCondition("mag", params.DoMagacina, "<=")
	qb.AddCondition("sifra_art", params.OdSifreArtikla, ">=")
	qb.AddCondition("sifra_art", params.DoSifreArtikla, "<=")
	qb.AddCondition("dadok", params.OdDatuma, ">=")
	qb.AddCondition("dadok", params.DoDatuma, "<=")
	qb.AddCondition("gru", params.OdGrupe, ">=")
	qb.AddCondition("gru", params.DoGrupe, "<=")
	qb.AddGroupBy("konto, sifra, naziv")
	qb.AddOrderBy("konto, sifra")
	if !getTotalRecords && !isPrint && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// The print loads the whole result in one call: its totals and its rows.
	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		}
		// The first pass runs without LIMIT, so the totals cover the whole result; the three sum
		// columns flagged with IncludeInTotals (vpiznos, nciznos, razlika) are filled, the others
		// stay empty.
		var vpiznos, nciznos, razlika float64
		for _, entity := range *entities {
			vpiznos += entity.Vpiznos
			nciznos += entity.Nciznos
			razlika += entity.Razlika
		}
		ukupno := map[string]string{
			"vpiznos": common.FormatNumberWithSystemLocale(vpiznos, 2),
			"nciznos": common.FormatNumberWithSystemLocale(nciznos, 2),
			"razlika": common.FormatNumberWithSystemLocale(razlika, 2),
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok {
				tbl.Totals[i] = value
			}
		}
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: robnoPrometRowFields(isPrint,
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Vpiznos, 2),
				common.FormatNumberWithSystemLocale(entity.Nciznos, 2),
				common.FormatNumberWithSystemLocale(entity.Razlika, 2),
			),
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// The prices of the "Lager lista" (the radio "tip cene" of the sub-tab).
const (
	robnoPrometLagerNetoFakturna    = "netofakturna"
	robnoPrometLagerProsecnaNabavna = "prosecnanabavna"
	robnoPrometLagerProdajna        = "prodajna"
)

// robnoPrometLagerNaziviJoins joins the names of the magacin, of the robna grupa and of the robna
// podgrupa of the article of the lager list (the table is rpro or rsta, both with god, kar and mag;
// the grupa and the podgrupa are the ones of the article, rsif).
func robnoPrometLagerNaziviJoins(table string) string {
	return fmt.Sprintf(` left join magacini mg on mg.god = %[1]s.god and mg.kar = %[1]s.kar and mg.mag = %[1]s.mag
	left join rgru rg on rg.god = rsif.god and rg.kar = rsif.kar and rg.gru = rsif.gru
	left join rpgru rpg on rpg.god = rsif.god and rpg.kar = rsif.kar and rpg.gru = rsif.gru and rpg.pgru = rsif.pgru`, table)
}

// robnoPrometLagerStampaRow returns one article of the print of the lager list: the columns of the
// legacy print, šifra, naziv, JM, stanje, cena and vrednost.
func robnoPrometLagerStampaRow(item *robnoPrometLagerStavka) domain.TableRow {
	return domain.TableRow{
		Fields: []string{
			fmt.Sprintf("%d", item.sifra),
			strings.TrimSpace(item.naziv + " " + item.serbr),
			item.jm,
			common.FormatNumberWithSystemLocale(item.stanje, 3),
			common.FormatNumberWithSystemLocale(item.cena, 2),
			common.FormatNumberWithSystemLocale(item.stanje*item.cena, 2),
		},
	}
}

// robnoPrometLagerGrupisano returns the rows of the lager list printed by grupe and podgrupe, like the
// legacy print: the articles (ordered by magacin, grupa and podgrupa) under the header of their
// magacin, robna grupa and robna podgrupa, and the total of the stanje and of the vrednost at the end
// of every podgrupa, grupa and magacin.
func robnoPrometLagerGrupisano(items []*robnoPrometLagerStavka) []domain.TableRow {
	type sums struct{ stanje, vrednost float64 }
	var uMagacin, uGrupa, uPodgrupa sums
	header := func(class string, broj int, naziv string) domain.TableRow {
		return domain.TableRow{ClassRow: class, Fields: []string{fmt.Sprintf("%d", broj), naziv}}
	}
	total := func(class string, broj int, naziv string, sum sums) domain.TableRow {
		return domain.TableRow{ClassRow: class, Fields: []string{
			fmt.Sprintf("%d", broj),
			naziv,
			common.FormatNumberWithSystemLocale(sum.stanje, 2),
			common.FormatNumberWithSystemLocale(sum.vrednost, 2),
		}}
	}
	rows := []domain.TableRow{}
	closeGroups := func(prev *robnoPrometLagerStavka, podgrupa, grupa, magacin bool) {
		if podgrupa {
			rows = append(rows, total(domain.RobnoPrometLagerPodgrupaUkupno, prev.pgru, prev.pgruNaziv, uPodgrupa))
			uPodgrupa = sums{}
		}
		if grupa {
			rows = append(rows, total(domain.RobnoPrometLagerGrupaUkupno, prev.gru, prev.gruNaziv, uGrupa))
			uGrupa = sums{}
		}
		if magacin {
			rows = append(rows, total(domain.RobnoPrometLagerMagacinUkupno, prev.mag, prev.magNaziv, uMagacin))
			uMagacin = sums{}
		}
	}
	for i, item := range items {
		newMagacin, newGrupa, newPodgrupa := true, true, true
		if i > 0 {
			prev := items[i-1]
			newMagacin = prev.mag != item.mag
			newGrupa = newMagacin || prev.gru != item.gru
			newPodgrupa = newGrupa || prev.pgru != item.pgru
			closeGroups(prev, newPodgrupa, newGrupa, newMagacin)
		}
		if newMagacin {
			rows = append(rows, header(domain.RobnoPrometLagerMagacin, item.mag, item.magNaziv))
		}
		if newGrupa {
			rows = append(rows, header(domain.RobnoPrometLagerGrupa, item.gru, item.gruNaziv))
		}
		if newPodgrupa {
			rows = append(rows, header(domain.RobnoPrometLagerPodgrupa, item.pgru, item.pgruNaziv))
		}
		rows = append(rows, robnoPrometLagerStampaRow(item))
		vrednost := item.stanje * item.cena
		for _, sum := range []*sums{&uMagacin, &uGrupa, &uPodgrupa} {
			sum.stanje += item.stanje
			sum.vrednost += vrednost
		}
	}
	if len(items) > 0 {
		closeGroups(items[len(items)-1], true, true, true)
	}
	return rows
}

// TODO Stole check the round of the values
// robnoPrometWinDevNumeric keeps 6 decimals of a value, truncating the others, like the assignment to a
// WinDev numeric variable (the legacy procedures hold the running price in a numeric, so every step of
// the average loses the decimals after the 6th; the lager list matches the legacy print only with the
// same truncation). The value is first rounded to 9 decimals so that the float error of an exact value
// (e.g. 292.9056 stored as 292.905599999...) does not truncate it to the decimal below.
func robnoPrometWinDevNumeric(value float64) float64 {
	return math.Trunc(math.Round(value*1e9)/1e3) / 1e6
}

// robnoPrometLagerStampaFields returns the columns of the print of the "Lager lista", the ones of the
// legacy print: šifra, naziv, JM, stanje (with its total), the price of the chosen tip cene and the
// vrednost at that price (with its total).
func robnoPrometLagerStampaFields(tipCene string) []domain.Fields {
	cena, vrednost := "Neto fakturna cena", "Vrednost po neto fakt. ceni"
	switch tipCene {
	case robnoPrometLagerProsecnaNabavna:
		cena, vrednost = "Prosečna nab. cena", "Vrednost po pros. nab. ceni"
	case robnoPrometLagerProdajna:
		cena, vrednost = "Prodajna cena", "Vrednost po prodajnoj ceni"
	}
	return []domain.Fields{
		{Name: "sifra", Label: "Šifra artikla", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "40"},
		{Name: "jm", Label: "J.M.", Width: "5"},
		{Name: "stanje", Label: "Stanje", Width: "13", TextAlign: "right", IncludeInTotals: true},
		{Name: "cena", Label: cena, Width: "13", TextAlign: "right"},
		{Name: "vrednost", Label: vrednost, Width: "16", TextAlign: "right", IncludeInTotals: true},
	}
}

// robnoPrometLagerStavka is one row of the "Lager lista": the stanje of an article in a magacin and
// its price.
type robnoPrometLagerStavka struct {
	mag, sifra, magaciniID, gru, pgru int
	naziv, jm, serbr                  string
	magNaziv, gruNaziv, pgruNaziv     string
	stanje, cena                      float64
}

// GetPrometRucLagerLista renders the "Lager lista" (robno promet, tab 4, sub-tab 1): the stanje of every
func (s *RobnoPrometResource) GetPrometRucLagerLista(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	isPrint := printType == common.TipStampePrint
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometRucLagerListaTableFields
	if isPrint {
		tbl.Headers = robnoPrometLagerStampaFields(params.TipCene)
	}
	tbl.HasTotals = true

	odPodgrupe, doPodgrupe := params.OdPodgrupe, params.DoPodgrupe
	if params.OdGrupe != params.DoGrupe {
		odPodgrupe, doPodgrupe = "0", "999"
	}

	// The stavke of the articles up to the date of the stanje (the legacy ROB_QRY_LAGER).
	qb := common.NewQueryBuilder(`select
			coalesce(rpro.mag, 0) as mag,
			coalesce(rpro.sifra, 0) as sifra_art,
			coalesce(rdok.magaciniid, 0) as magaciniid,
			coalesce(rsif.naziv, '') as naziv,
			coalesce(rsif.jm, '') as jm,
			coalesce(rsif.gru, 0) as gru,
			coalesce(rsif.pgru, 0) as pgru,
			coalesce(rsif.serbr, '') as serbr,
			coalesce(mg.opis, '') as magnaziv,
			coalesce(rg.naziv, '') as grunaziv,
			coalesce(rpg.naziv, '') as pgrunaziv,
			upper(coalesce(dokvrsta.kodknj, '')) as kodknj,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.fcena, 0) as fcena,
			coalesce(rpro.ncena, 0) as ncena
		from rdok`, true)
	qb.AddJoin(" inner join rpro on rpro.rdokid = rdok.rdokid")
	qb.AddJoin(" inner join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(" inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin(robnoPrometLagerNaziviJoins("rpro"))
	qb.AddEqual("rpro.god", userSession.SelectedGod)
	qb.AddEqual("rpro.kar", userSession.SelectedKar)
	qb.AddCondition("rpro.mag", params.OdMagacina, ">=")
	qb.AddCondition("rpro.mag", params.DoMagacina, "<=")
	qb.AddCondition("rsif.gru", params.OdGrupe, ">=")
	qb.AddCondition("rsif.gru", params.DoGrupe, "<=")
	qb.AddCondition("rsif.pgru", odPodgrupe, ">=")
	qb.AddCondition("rsif.pgru", doPodgrupe, "<=")
	qb.AddCondition("rdok.danal", params.StanjeNaDan, "<=")
	qb.AddCustomCondition("coalesce(dokvrsta.grpdok, '') <> 'NIV'")
	qb.AddOrderBy("rpro.mag, rpro.sifra, rdok.danal, rdok.rdokid")
	sqlQuery, args := qb.Build()
	stavke, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// The stanje of the articles of the magacini (rsta): the articles without a stavka and, for the
	// prodajna cena, the cena of every article in its magacin.
	qb = common.NewQueryBuilder(`select
			coalesce(rsta.mag, 0) as mag,
			coalesce(rsta.sifra, 0) as sifra_art,
			coalesce(rsta.magaciniid, 0) as magaciniid,
			coalesce(rsif.naziv, '') as naziv,
			coalesce(rsif.jm, '') as jm,
			coalesce(rsif.gru, 0) as gru,
			coalesce(rsif.pgru, 0) as pgru,
			coalesce(rsif.serbr, '') as serbr,
			coalesce(mg.opis, '') as magnaziv,
			coalesce(rg.naziv, '') as grunaziv,
			coalesce(rpg.naziv, '') as pgrunaziv,
			coalesce(rsta.ulaz, 0) as ulaz,
			coalesce(rsta.izlaz, 0) as izlaz,
			coalesce(rsta.cena, 0) as cena
		from rsta`, true)
	qb.AddJoin(" inner join rsif on rsif.rsifid = rsta.rsifid")
	qb.AddJoin(robnoPrometLagerNaziviJoins("rsta"))
	qb.AddEqual("rsta.god", userSession.SelectedGod)
	qb.AddEqual("rsta.kar", userSession.SelectedKar)
	qb.AddCondition("rsta.mag", params.OdMagacina, ">=")
	qb.AddCondition("rsta.mag", params.DoMagacina, "<=")
	qb.AddCondition("rsif.gru", params.OdGrupe, ">=")
	qb.AddCondition("rsif.gru", params.DoGrupe, "<=")
	qb.AddCondition("rsif.pgru", odPodgrupe, ">=")
	qb.AddCondition("rsif.pgru", doPodgrupe, "<=")
	qb.AddOrderBy("rsta.mag, rsta.sifra")
	sqlQuery, args = qb.Build()
	stanja, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// The stanje and the price of every article of the stavke: the running stanje and the running
	// average price start again with every article of a magacin.
	type lagerKey struct{ mag, sifra int }
	lager := map[lagerKey]*robnoPrometLagerStavka{}
	var stavkeLagera []*robnoPrometLagerStavka
	var stanje, cena float64
	fakturna := params.TipCene != robnoPrometLagerProsecnaNabavna && params.TipCene != robnoPrometLagerProdajna
	prevSifra, started := 0, false
	for _, st := range *stavke {
		key := lagerKey{st.Mag, st.SifraArt}
		if !started || st.SifraArt != prevSifra {
			stanje = 0
			if !fakturna {
				cena = 0
			}
			prevSifra, started = st.SifraArt, true
		}
		switch st.Kodknj {
		case "D":
			switch params.TipCene {
			case robnoPrometLagerProsecnaNabavna:
				if (st.Vrd == 101 || st.Vrd == 110 || st.Vrd == 111) && stanje+st.Kolic != 0 {
					cena = robnoPrometWinDevNumeric((stanje*cena + st.Kolic*st.Ncena) / (stanje + st.Kolic))
				}
			case robnoPrometLagerProdajna:
				// The prodajna cena is the cena of the stanje (rsta), set below.
			default:
				if stanje+st.Kolic != 0 {
					cena = robnoPrometWinDevNumeric((stanje*cena + st.Kolic*st.Fcena) / (stanje + st.Kolic))
				}
			}
			stanje += st.Kolic
		case "P":
			if params.TipCene == robnoPrometLagerProsecnaNabavna && st.Vrd == 121 && stanje+st.Kolic != 0 {
				cena = robnoPrometWinDevNumeric((stanje*cena + math.Abs(st.Kolic)*st.Ncena) / (stanje + math.Abs(st.Kolic)))
			}
			stanje -= st.Kolic
		}
		item, found := lager[key]
		if !found {
			item = &robnoPrometLagerStavka{mag: st.Mag, sifra: st.SifraArt, magaciniID: st.MagaciniID, gru: st.Gru, pgru: st.Pgru, naziv: st.Naziv, jm: st.JM, serbr: st.Serbr,
				magNaziv: st.MagNaziv, gruNaziv: st.GruNaziv, pgruNaziv: st.PgruNaziv, cena: cena}
			lager[key] = item
			stavkeLagera = append(stavkeLagera, item)
		} else if !fakturna {
			item.cena = cena
		}
		item.stanje = stanje
	}

	type stanjeKey struct{ magaciniID, sifra int }
	prodajneCene := map[stanjeKey]float64{}
	for _, sta := range *stanja {
		prodajneCene[stanjeKey{sta.MagaciniID, sta.SifraArt}] = sta.Cena
		key := lagerKey{sta.Mag, sta.SifraArt}
		if _, found := lager[key]; found {
			continue
		}
		item := &robnoPrometLagerStavka{mag: sta.Mag, sifra: sta.SifraArt, magaciniID: sta.MagaciniID, gru: sta.Gru, pgru: sta.Pgru, naziv: sta.Naziv, jm: sta.JM, serbr: sta.Serbr,
			magNaziv: sta.MagNaziv, gruNaziv: sta.GruNaziv, pgruNaziv: sta.PgruNaziv, stanje: sta.Ulaz - sta.Izlaz, cena: sta.Cena}
		lager[key] = item
		stavkeLagera = append(stavkeLagera, item)
	}
	if params.TipCene == robnoPrometLagerProdajna {
		for _, item := range stavkeLagera {
			item.cena = prodajneCene[stanjeKey{item.magaciniID, item.sifra}]
		}
	}

	// The rows of the list: only the stanja <> 0 when "Zalihe <> 0" is checked, ordered by magacin and
	// šifra; by magacin, grupa and podgrupa first when the list is printed by grupe (the groups of the
	// print) and by the naziv when it is printed in the azbučni red.
	rows := make([]*robnoPrometLagerStavka, 0, len(stavkeLagera))
	for _, item := range stavkeLagera {
		if params.ZaliheOdNule && item.stanje == 0 {
			continue
		}
		rows = append(rows, item)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if params.StampajGrupaPodgrupa {
			if a.mag != b.mag {
				return a.mag < b.mag
			}
			if a.gru != b.gru {
				return a.gru < b.gru
			}
			if a.pgru != b.pgru {
				return a.pgru < b.pgru
			}
		}
		if params.AzbucniRed {
			if na, nb := strings.ToLower(a.naziv), strings.ToLower(b.naziv); na != nb {
				return na < nb
			}
		}
		if a.mag != b.mag {
			return a.mag < b.mag
		}
		return a.sifra < b.sifra
	})

	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(rows), pageSize)
		}
		var stanjeUkupno, vrednost float64
		for _, item := range rows {
			stanjeUkupno += item.stanje
			vrednost += item.stanje * item.cena
		}
		ukupno := map[string]string{
			"stanje":   common.FormatNumberWithSystemLocale(stanjeUkupno, 2),
			"vrednost": common.FormatNumberWithSystemLocale(vrednost, 2),
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok {
				tbl.Totals[i] = value
			}
		}
		if !isPrint {
			return nil
		}
	}

	// The requested page of the grid (the print shows every row).
	page := rows
	if !isPrint && pageSize > 0 {
		start := (currentPage - 1) * pageSize
		if start < 0 {
			start = 0
		}
		if start > len(rows) {
			start = len(rows)
		}
		end := start + pageSize
		if end > len(rows) {
			end = len(rows)
		}
		page = rows[start:end]
	}
	tbl.Rows = []domain.TableRow{}
	if isPrint && params.StampajGrupaPodgrupa {
		tbl.Rows = robnoPrometLagerGrupisano(page)
		return nil
	}
	for _, item := range page {
		if isPrint {
			tbl.Rows = append(tbl.Rows, robnoPrometLagerStampaRow(item))
			continue
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: robnoPrometRowFields(isPrint,
				fmt.Sprintf("%d", item.mag),
				fmt.Sprintf("%d", item.sifra),
				strings.TrimSpace(item.naziv+" "+item.serbr),
				item.jm,
				common.FormatNumberWithSystemLocale(item.stanje, 3),
				common.FormatNumberWithSystemLocale(item.cena, 2),
				common.FormatNumberWithSystemLocale(item.stanje*item.cena, 2),
				fmt.Sprintf("%d", item.gru),
				fmt.Sprintf("%d", item.pgru),
				item.serbr,
			),
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometRucMagacini renders the "RUC po magacinima" (robno promet, tab 4, sub-tab 3): for every
func (s *RobnoPrometResource) GetPrometRucMagacini(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	isPrint := printType == common.TipStampePrint
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometRucMagaciniTableFields
	if isPrint {
		tbl.Headers = robnoPrometHiddenFields(s.prometRucMagaciniTableFields, "detail", "mag")
	}
	tbl.HasTotals = true

	// The stavke of the fakture of the period, ordered by magacin and article (the legacy dsDATA).
	qb := common.NewQueryBuilder(`select
			coalesce(rpro.mag, 0) as mag,
			coalesce(rpro.sifra, 0) as sifra_art,
			rsif.rsifid,
			coalesce(rsif.naziv, '') as naziv,
			coalesce(rsif.jm, '') as jm,
			coalesce(mg.opis, '') as magnaziv,
			coalesce(mg.nacvodzal, 0) as nacvodzal,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.cena, 0) as cena,
			coalesce(rpro.ncena, 0) as ncena,
			coalesce(rpro.fcena, 0) as fcena,
			coalesce(rpro.rab, 0) as rab
		from rdok`, true)
	qb.AddJoin(" inner join rpro on rpro.rdokid = rdok.rdokid")
	qb.AddJoin(" inner join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(" inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin(" left join magacini mg on mg.god = rpro.god and mg.kar = rpro.kar and mg.mag = rpro.mag")
	qb.AddEqual("rdok.god", userSession.SelectedGod)
	qb.AddEqual("rdok.kar", userSession.SelectedKar)
	qb.AddCondition("rdok.mag", params.OdMagacina, ">=")
	qb.AddCondition("rdok.mag", params.DoMagacina, "<=")
	qb.AddCondition("rdok.danal", params.OdDatuma, ">=")
	qb.AddCondition("rdok.danal", params.DoDatuma, "<=")
	qb.AddCustomCondition("upper(coalesce(dokvrsta.grpdok, '')) = 'FAK'")
	qb.AddCustomCondition("rpro.sifra > 0")
	qb.AddOrderBy("rpro.mag, rpro.sifra, rdok.danal, rdok.rdokid, rpro.rproid")
	sqlQuery, args := qb.Build()
	stavke, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// The articles of the magacini: the sums of their stavke with 6 decimals, like the legacy.
	type rucStavka struct {
		mag, sifra                     int
		naziv, jm, magNaziv            string
		kolic, netofvred, nabvred, ruc float64
	}
	type rucKey struct{ mag, rsifID int }
	artikli := map[rucKey]*rucStavka{}
	var redosled []*rucStavka
	for _, st := range *stavke {
		key := rucKey{st.Mag, st.RsifID}
		item, found := artikli[key]
		if !found {
			item = &rucStavka{mag: st.Mag, sifra: st.SifraArt, naziv: st.Naziv, jm: st.JM, magNaziv: st.MagNaziv}
			artikli[key] = item
			redosled = append(redosled, item)
		}
		cena, nabavna := st.Cena, st.Ncena
		if st.Nacvodzal == 3 {
			cena, nabavna = st.Fcena, st.Cena
		}
		netofvred := st.Kolic * cena * (1 - st.Rab/100)
		nabvred := st.Kolic * nabavna
		item.kolic = robnoPrometWinDevNumeric(item.kolic + st.Kolic)
		item.netofvred = robnoPrometWinDevNumeric(item.netofvred + netofvred)
		item.nabvred = robnoPrometWinDevNumeric(item.nabvred + nabvred)
		item.ruc = robnoPrometWinDevNumeric(item.ruc + netofvred - nabvred)
	}
	var ukupnoRuc float64
	for _, item := range redosled {
		ukupnoRuc = robnoPrometWinDevNumeric(ukupnoRuc + item.ruc)
	}

	// The values of the columns of an article (or of a total), by the name of the column.
	type sums struct{ kolic, netofvred, nabvred, ruc float64 }
	procenatRuc := func(ruc, nabvred float64) float64 {
		switch {
		case nabvred != 0:
			return ruc / nabvred * 100
		case ruc != 0:
			return 100
		}
		return 0
	}
	ucesce := func(ruc float64) float64 {
		if ukupnoRuc == 0 {
			return 0
		}
		return ruc / ukupnoRuc * 100
	}
	prosek := func(vrednost, kolic float64) string {
		if kolic == 0 {
			return common.FormatNumberWithSystemLocale(0, 3)
		}
		return common.FormatNumberWithSystemLocale(math.Round(vrednost/kolic*100)/100, 3)
	}
	values := func(sum sums, withAverages bool) map[string]string {
		v := map[string]string{
			"kolic":     common.FormatNumberWithSystemLocale(sum.kolic, 3),
			"netofvred": common.FormatNumberWithSystemLocale(sum.netofvred, 2),
			"nabvred":   common.FormatNumberWithSystemLocale(sum.nabvred, 2),
			"ruc":       common.FormatNumberWithSystemLocale(sum.ruc, 2),
			"procruc":   common.FormatNumberWithSystemLocale(procenatRuc(sum.ruc, sum.nabvred), 3) + " %",
			"procu":     common.FormatNumberWithSystemLocale(ucesce(sum.ruc), 3) + " %",
		}
		if withAverages {
			v["prosfc"] = prosek(sum.netofvred, sum.kolic)
			v["prosnc"] = prosek(sum.nabvred, sum.kolic)
		}
		return v
	}
	row := func(texts map[string]string) []string {
		fields := make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			fields[i] = texts[header.Name]
		}
		return fields
	}
	artikal := func(item *rucStavka) []string {
		texts := values(sums{item.kolic, item.netofvred, item.nabvred, item.ruc}, true)
		texts["detail"] = "🔽"
		texts["mag"] = fmt.Sprintf("%d", item.mag)
		texts["sifra"] = fmt.Sprintf("%d", item.sifra)
		texts["naziv"] = item.naziv
		texts["jm"] = item.jm
		return row(texts)
	}

	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(redosled), pageSize)
		}
		var ukupno sums
		for _, item := range redosled {
			ukupno.kolic += item.kolic
			ukupno.netofvred += item.netofvred
			ukupno.nabvred += item.nabvred
			ukupno.ruc += item.ruc
		}
		tbl.Totals = row(values(ukupno, false))
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	if isPrint {
		// The articles under the header of their magacin and the total of every magacin.
		var magacin sums
		for i, item := range redosled {
			if i == 0 || redosled[i-1].mag != item.mag {
				if !params.StampajSamoZbir {
					tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: domain.RobnoPrometRucMagacin, Fields: []string{fmt.Sprintf("%d", item.mag), item.magNaziv}})
				}
				magacin = sums{}
			}
			if !params.StampajSamoZbir {
				tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: artikal(item)})
			}
			magacin.kolic += item.kolic
			magacin.netofvred += item.netofvred
			magacin.nabvred += item.nabvred
			magacin.ruc += item.ruc
			if i == len(redosled)-1 || redosled[i+1].mag != item.mag {
				fields := row(values(magacin, false))
				fields[0], fields[1] = fmt.Sprintf("%d", item.mag), item.magNaziv
				tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: domain.RobnoPrometRucMagacinUkupno, Fields: fields})
			}
		}
		return nil
	}

	// The requested page of the grid.
	start, end := 0, len(redosled)
	if pageSize > 0 {
		start = (currentPage - 1) * pageSize
		if start < 0 {
			start = 0
		}
		if start > len(redosled) {
			start = len(redosled)
		}
		if start+pageSize < end {
			end = start + pageSize
		}
	}
	for _, item := range redosled[start:end] {
		tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: artikal(item), HasUpdate: false, HasDelete: false})
	}
	return nil
}

// robnoPrometHiddenFields returns the columns without the ones with the given names.
func robnoPrometHiddenFields(fields []domain.Fields, names ...string) []domain.Fields {
	hidden := map[string]bool{}
	for _, name := range names {
		hidden[name] = true
	}
	headers := make([]domain.Fields, 0, len(fields))
	for _, field := range fields {
		if !hidden[field.Name] {
			headers = append(headers, field)
		}
	}
	return headers
}

// GetPrometRucFakture renders the "Izlazne fakture" (robno promet, tab 4, sub-tab 4): 
func (s *RobnoPrometResource) GetPrometRucFakture(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	isPrint := printType == common.TipStampePrint
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometRucFaktureTableFields
	if isPrint {
		tbl.Headers = robnoPrometRucFaktureStampaFields
	}
	tbl.HasTotals = true

	// The stavke of the fakture of the period (the legacy dsDATA).
	qb := common.NewQueryBuilder(`select
			rdok.rdokid,
			coalesce(rdok.mag, 0) as mag,
			coalesce(mg.opis, '') as magnaziv,
			coalesce(mg.nacvodzal, 0) as nacvodzal,
			coalesce(rdok.tipdok, '') || '-' || coalesce(rdok.nalog, 0) as brnal,
			rdok.danal,
			coalesce(rdok.vrd, 0) as vrd,
			coalesce(rdok.dokum, 0)::text as dokum,
			coalesce(rdok.dokiz, '') as dokiz,
			rdok.dadok,
			coalesce(rdok.rok, 0) as rok,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(cp.naziv, 'Naziv ne postoji!') as naziv,
			coalesce(rdok.kom, 0) as kom,
			coalesce(kom.imeprezime, 'Ne postoji komercijalista') as komerc,
			coalesce(rpro.kolic, 0) as kolic,
			coalesce(rpro.cena, 0) as cena,
			coalesce(rpro.ncena, 0) as ncena,
			coalesce(rpro.fcena, 0) as fcena,
			coalesce(rpro.rab, 0) as rab
		from rdok`, true)
	qb.AddJoin(" inner join rpro on rpro.rdokid = rdok.rdokid")
	qb.AddJoin(" inner join rsif on rsif.rsifid = rpro.rsifid")
	qb.AddJoin(" inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin(" left join magacini mg on mg.god = rdok.god and mg.kar = rdok.kar and mg.mag = rdok.mag")
	qb.AddJoin(` left join (select distinct on (f.god, f.kar, f.konto, f.sifra) f.god, f.kar, f.konto, f.sifra, f.naziv
			from fkpl f
			order by f.god, f.kar, f.konto, f.sifra, f.idfkpl) cp
			on cp.god = rdok.god and cp.kar = rdok.kar and cp.konto = rdok.fkto and cp.sifra = rdok.fana`)
	qb.AddJoin(" left join komercijalisti kom on kom.god = rdok.god and kom.kar = rdok.kar and kom.sifkom = rdok.kom")
	qb.AddEqual("rdok.god", userSession.SelectedGod)
	qb.AddEqual("rdok.kar", userSession.SelectedKar)
	qb.AddCondition("rdok.mag", params.OdMagacina, ">=")
	qb.AddCondition("rdok.mag", params.DoMagacina, "<=")
	qb.AddCondition("rdok.danal", params.OdDatuma, ">=")
	qb.AddCondition("rdok.danal", params.DoDatuma, "<=")
	qb.AddCustomCondition("upper(coalesce(dokvrsta.grpdok, '')) = 'FAK'")
	if !params.UkljuceneUsluge {
		qb.AddCustomCondition("upper(coalesce(rsif.tip, '')) <> 'U'")
	}
	qb.AddOrderBy("rdok.mag, rdok.danal, rdok.tipdok, rdok.nalog, rdok.vrd, rdok.dokum, rdok.dadok, rdok.rdokid, rpro.rproid")
	sqlQuery, args := qb.Build()
	stavke, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// The documents: the sums of their stavke with 6 decimals, like the legacy.
	type rucDokument struct {
		st                      domain.RobnoPrometDto
		netofvred, nabvred, ruc float64
	}
	dokumenti := map[int64]*rucDokument{}
	var redosled []*rucDokument
	for _, st := range *stavke {
		dok, found := dokumenti[st.RdokID]
		if !found {
			dok = &rucDokument{st: st}
			dokumenti[st.RdokID] = dok
			redosled = append(redosled, dok)
		}
		cena, nabavna := st.Cena, st.Ncena
		if st.Nacvodzal == 3 {
			cena, nabavna = st.Fcena, st.Cena
		}
		netofvred := st.Kolic * cena * (1 - st.Rab/100)
		nabvred := st.Kolic * nabavna
		dok.netofvred = robnoPrometWinDevNumeric(dok.netofvred + netofvred)
		dok.nabvred = robnoPrometWinDevNumeric(dok.nabvred + nabvred)
		dok.ruc = robnoPrometWinDevNumeric(dok.ruc + netofvred - nabvred)
	}

	type sums struct{ netofvred, nabvred, ruc float64 }
	values := func(sum sums) map[string]string {
		procruc := 0.0
		switch {
		case sum.nabvred != 0:
			procruc = sum.ruc / sum.nabvred * 100
		case sum.ruc != 0:
			procruc = 100
		}
		return map[string]string{
			"netofvred": common.FormatNumberWithSystemLocale(sum.netofvred, 2),
			"nabvred":   common.FormatNumberWithSystemLocale(sum.nabvred, 2),
			"ruc":       common.FormatNumberWithSystemLocale(sum.ruc, 2),
			"procruc":   common.FormatNumberWithSystemLocale(procruc, 3) + " %",
		}
	}
	row := func(texts map[string]string) []string {
		fields := make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			fields[i] = texts[header.Name]
		}
		return fields
	}
	dokument := func(dok *rucDokument) []string {
		st := dok.st
		texts := values(sums{dok.netofvred, dok.nabvred, dok.ruc})
		texts["detail"] = "🔽"
		texts["mag"] = fmt.Sprintf("%d", st.Mag)
		texts["brnal"] = st.Brnal
		texts["danal"] = common.FormatNullTime(st.Danal, common.DateLayout)
		texts["vrd"] = fmt.Sprintf("%d", st.Vrd)
		texts["brdok"] = st.Dokum
		texts["dokiz"] = st.Dokiz
		texts["dadok"] = common.FormatNullTime(st.Dadok, common.DateLayout)
		texts["dospece"] = common.AddDaysToNullTime(st.Dadok, st.Rok, common.DateLayout)
		texts["fkto"] = st.Fkto
		texts["fana"] = st.Fana
		texts["naziv"] = st.Naziv
		texts["sifkom"] = fmt.Sprintf("%d", st.Kom)
		texts["komerc"] = st.Komerc
		texts["dokument"] = fmt.Sprintf("%s-%d-%s", st.Brnal, st.Vrd, st.Dokum)
		return row(texts)
	}

	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(redosled), pageSize)
		}
		var ukupno sums
		for _, dok := range redosled {
			ukupno.netofvred += dok.netofvred
			ukupno.nabvred += dok.nabvred
			ukupno.ruc += dok.ruc
		}
		tbl.Totals = row(values(ukupno))
		if !isPrint {
			return nil
		}
	}

	tbl.Rows = []domain.TableRow{}
	if isPrint {
		// The documents under the header of their magacin and the total of every magacin.
		var magacin sums
		for i, dok := range redosled {
			if i == 0 || redosled[i-1].st.Mag != dok.st.Mag {
				if !params.StampajSamoZbir {
					tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: domain.RobnoPrometRucMagacin, Fields: []string{fmt.Sprintf("%d", dok.st.Mag), dok.st.MagNaziv}})
				}
				magacin = sums{}
			}
			if !params.StampajSamoZbir {
				tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: dokument(dok)})
			}
			magacin.netofvred += dok.netofvred
			magacin.nabvred += dok.nabvred
			magacin.ruc += dok.ruc
			if i == len(redosled)-1 || redosled[i+1].st.Mag != dok.st.Mag {
				fields := row(values(magacin))
				fields[0], fields[1] = fmt.Sprintf("%d", dok.st.Mag), dok.st.MagNaziv
				tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: domain.RobnoPrometRucMagacinUkupno, Fields: fields})
			}
		}
		return nil
	}

	// The requested page of the grid.
	start, end := 0, len(redosled)
	if pageSize > 0 {
		start = (currentPage - 1) * pageSize
		if start < 0 {
			start = 0
		}
		if start > len(redosled) {
			start = len(redosled)
		}
		if start+pageSize < end {
			end = start + pageSize
		}
	}
	for _, dok := range redosled[start:end] {
		tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", dok.st.RdokID), Fields: dokument(dok), HasUpdate: false, HasDelete: false})
	}
	return nil
}

// robnoPrometRucFaktureStampaFields are the columns of the print of the "Izlazne fakture", the ones of
// the legacy "PREGLED FAKTURA": the document (nalog-vrsta-broj), its dates, the kupac, the values of the
// RUC and the komercijalista.
var robnoPrometRucFaktureStampaFields = []domain.Fields{
	{Name: "dokument", Label: "Dokument", Width: "10"},
	{Name: "danal", Label: "Datum naloga", Width: "6", TextAlign: "center"},
	{Name: "dadok", Label: "Datum dokumenta", Width: "6", TextAlign: "center"},
	{Name: "dospece", Label: "Dospeće", Width: "6", TextAlign: "center"},
	{Name: "fkto", Label: "Konto", Width: "4"},
	{Name: "fana", Label: "Šifra", Width: "4"},
	{Name: "naziv", Label: "Naziv kupca", Width: "22"},
	{Name: "netofvred", Label: "Neto fakt. vrednost", Width: "8", TextAlign: "right", IncludeInTotals: true},
	{Name: "nabvred", Label: "Nabavna vrednost", Width: "8", TextAlign: "right", IncludeInTotals: true},
	{Name: "ruc", Label: "RUC", Width: "7", TextAlign: "right", IncludeInTotals: true},
	{Name: "procruc", Label: "% RUC-a", Width: "5", TextAlign: "right"},
	{Name: "sifkom", Label: "Šifra kom.", Width: "3", TextAlign: "right"},
	{Name: "komerc", Label: "Komercijalista", Width: "11"},
}

// The promet of the "Ulaz/izlaz za period" (the radio of the sub-tab).
const (
	robnoPrometUlaz      = "ulaz"
	robnoPrometIzlaz     = "izlaz"
	robnoPrometUlazIzlaz = "ulazizlaz"
)

// GetPrometRucUlazIzlaz renders the "Ulaz/izlaz za period" (robno promet, tab 4, sub-tab 2): for every
func (s *RobnoPrometResource) GetPrometRucUlazIzlaz(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	isPrint := printType == common.TipStampePrint
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = robnoPrometUlazIzlazHeaders(s.prometRucUlazIzlazTableFields, params.UlazIzlaz, isPrint)
	tbl.HasTotals = true

	// The kodknj of the stavke of the chosen promet and the sign of their količina and magacinska
	// vrednost: for the ulaz i izlaz the izlazi subtract; the fakturna vrednost and the rabat are summed
	// only for the ulaz or the izlaz.
	kodknj, sign, fakturna := "('D', 'P')", "case when upper(dokvrsta.kodknj) = 'P' then -1 else 1 end", "0"
	switch params.UlazIzlaz {
	case robnoPrometUlaz:
		kodknj, sign, fakturna = "('D')", "1", "1"
	case robnoPrometIzlaz:
		kodknj, sign, fakturna = "('P')", "1", "1"
	}
	// The key of the rows: the magacin and the article, or the article for the sum of the magacini.
	key, magacin, magaciniID, stanje := "s.mag, s.sifra", "s.mag", "min(s.magaciniid)",
		`(select coalesce(sum(coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0)), 0) from rsta r
			where r.god = $1 and r.kar = $2 and r.magaciniid = agg.magaciniid and r.sifra = agg.sifra_art)`
	if params.ZbirMagacina {
		key, magacin, magaciniID = "s.sifra", "0", "0"
		stanje = `(select coalesce(sum(coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0)), 0) from rsta r
			where r.god = $1 and r.kar = $2 and r.sifra = agg.sifra_art and r.mag >= $3 and r.mag <= $4)`
	}
	// $5 is the first day of the year, $6 OdDatuma (the first day of the period), $7 DoDatuma: the
	// stavke before the period and the stavke of the period.
	before, period := "s.danal < $6", "s.danal >= $6"
	sqlQuery := fmt.Sprintf(`with stavke as (
			select rdok.mag, rdok.magaciniid, rdok.danal, rpro.sifra,
				rsif.naziv, rsif.jm, rsif.pro, rsif.proizsifra, rsif.komercopis,
				upper(dokvrsta.kodknj) as kodknj,
				%[1]s as znak,
				coalesce(rpro.kolic, 0) as kolic,
				round(coalesce(rpro.kolic, 0) * coalesce(rpro.cena, 0), 2) as magvred,
				%[2]s * round(coalesce(rpro.kolic, 0) * coalesce(rpro.fcena, 0), 2) as fakvred,
				%[2]s * round(coalesce(rpro.kolic, 0) * coalesce(rpro.fcena, 0) * coalesce(rpro.rab, 0) / 100, 2) as rabat
			from rdok
			inner join rpro on rpro.rdokid = rdok.rdokid
			inner join rsif on rsif.rsifid = rpro.rsifid
			inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd
			where rpro.god = $1 and rpro.kar = $2
				and rdok.mag >= $3 and rdok.mag <= $4
				and rdok.danal >= $5 and rdok.danal <= $7
				and upper(dokvrsta.kodknj) in %[3]s
				and rpro.sifra > 0
		), agg as (
			select %[4]s as mag, %[5]s as magaciniid, s.sifra as sifra_art,
				coalesce(min(s.naziv), '') as naziv,
				coalesce(min(s.jm), '') as jm,
				coalesce(min(s.pro), '') as pro,
				coalesce(min(s.proizsifra), '') as proizsifra,
				coalesce(min(s.komercopis), '') as komercopis,
				coalesce(sum(case when %[6]s then s.znak * s.kolic end), 0) as koldodat,
				coalesce(sum(case when %[7]s and s.kodknj = 'D' then s.kolic end), 0) as ulaz,
				coalesce(sum(case when %[7]s and s.kodknj = 'P' then s.kolic end), 0) as izlaz,
				coalesce(sum(s.znak * s.kolic), 0) as kolnadan,
				coalesce(sum(case when %[6]s then s.fakvred end), 0) as fakvreddo,
				coalesce(sum(case when %[6]s then s.rabat end), 0) as rabatdo,
				coalesce(sum(case when %[6]s then s.znak * s.magvred end), 0) as magvreddo,
				coalesce(sum(case when %[7]s then s.fakvred end), 0) as fakvrednost,
				coalesce(sum(case when %[7]s then s.rabat end), 0) as rabat,
				coalesce(sum(case when %[7]s then s.znak * s.magvred end), 0) as magvrednost,
				coalesce(sum(s.fakvred), 0) as fakvrednostnadan,
				coalesce(sum(s.rabat), 0) as rabatnadan,
				coalesce(sum(s.znak * s.magvred), 0) as magvrednostnadan
			from stavke s
			group by %[8]s
		)
		select agg.*, %[9]s as stanje
		from agg
		order by agg.sifra_art, agg.mag`,
		sign, fakturna, kodknj, magacin, magaciniID, before, period, key, stanje)
	odDatuma, doDatuma := params.OdDatuma, params.DoDatuma
	args := []any{userSession.SelectedGod, userSession.SelectedKar, params.OdMagacina, params.DoMagacina,
		fmt.Sprintf("%d-01-01", userSession.SelectedGod), odDatuma, doDatuma}
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	rows := *entities

	// The values of the columns of a row, by the name of the column.
	vrednosti := func(e domain.RobnoPrometDto) map[string]float64 {
		return map[string]float64{
			"koldodat": e.Koldodat, "ulaz": e.Ulaz, "izlaz": e.Izlaz, "kolnadan": e.Kolnadan, "stanje": e.Stanje,
			"fakvreddo": e.Fakvreddo, "rabatdo": e.Rabatdo, "magvreddo": e.Magvreddo,
			"fakvrednost": e.Fakvrednost, "rabat": e.Rabat, "magvrednost": e.Magvrednost,
			"fakvrednostnadan": e.Fakvrednostnadan, "rabatnadan": e.Rabatnadan, "magvrednostnadan": e.Magvrednostnadan,
		}
	}
	kolicine := map[string]bool{"koldodat": true, "ulaz": true, "izlaz": true, "kolnadan": true, "stanje": true}
	format := func(name string, value float64) string {
		if kolicine[name] {
			return common.FormatNumberWithSystemLocale(value, 3)
		}
		return common.FormatNumberWithSystemLocale(value, 2)
	}

	if getTotalRecords || isPrint {
		if !isPrint {
			common.SetTableTotalRecords(tbl, len(rows), pageSize)
		}
		ukupno := map[string]float64{}
		for _, e := range rows {
			for name, value := range vrednosti(e) {
				ukupno[name] += value
			}
		}
		tbl.Totals = make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if value, ok := ukupno[header.Name]; ok && header.IncludeInTotals {
				tbl.Totals[i] = format(header.Name, value)
			}
		}
		if !isPrint {
			return nil
		}
	}

	// The requested page of the grid (the print shows every row).
	page := rows
	if !isPrint && pageSize > 0 {
		start := (currentPage - 1) * pageSize
		if start < 0 {
			start = 0
		}
		if start > len(rows) {
			start = len(rows)
		}
		end := start + pageSize
		if end > len(rows) {
			end = len(rows)
		}
		page = rows[start:end]
	}
	tbl.Rows = []domain.TableRow{}
	for _, e := range page {
		values := vrednosti(e)
		texts := map[string]string{
			"detail":     "🔽",
			"mag":        fmt.Sprintf("%d", e.Mag),
			"sifra":      fmt.Sprintf("%d", e.SifraArt),
			"naziv":      e.Naziv,
			"jm":         e.JM,
			"pro":        e.Pro,
			"proizsifra": e.Proizsifra,
			"komercopis": e.Komercopis,
		}
		fields := make([]string, len(tbl.Headers))
		for i, header := range tbl.Headers {
			if text, ok := texts[header.Name]; ok {
				fields[i] = text
			} else if value, ok := values[header.Name]; ok {
				fields[i] = format(header.Name, value)
			}
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false})
	}
	return nil
}

// robnoPrometUlazIzlazHeaders returns the columns of the "Ulaz/izlaz za period" for the chosen promet,
func robnoPrometUlazIzlazHeaders(fields []domain.Fields, ulazIzlaz string, isPrint bool) []domain.Fields {
	koldodat, kolnadan := "Stanje do početka perioda", "Stanje na kraju perioda"
	hidden := map[string]bool{"fakvreddo": true, "rabatdo": true, "fakvrednost": true, "rabat": true, "fakvrednostnadan": true, "rabatnadan": true}
	switch ulazIzlaz {
	case robnoPrometUlaz:
		koldodat, kolnadan = "Količina ulaz do početka perioda", "Količina ulaz na kraju perioda"
		hidden = map[string]bool{"izlaz": true}
	case robnoPrometIzlaz:
		koldodat, kolnadan = "Količina izlaz do početka perioda", "Količina izlaz na kraju perioda"
		hidden = map[string]bool{"ulaz": true}
	}
	if isPrint {
		for _, name := range []string{"detail", "pro", "proizsifra", "komercopis"} {
			hidden[name] = true
		}
	}
	headers := make([]domain.Fields, 0, len(fields))
	for _, field := range fields {
		if hidden[field.Name] {
			continue
		}
		switch field.Name {
		case "koldodat":
			field.Label = koldodat
		case "kolnadan":
			field.Label = kolnadan
		}
		headers = append(headers, field)
	}
	return headers
}

// GetMagacinComboValues returns the magacini of the current period keyed by mag (CommonService).
func (s *RobnoPrometResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinByMagComboValues(ctx)
}

// GetRobneGrupeComboValues returns the robne grupe of the current period (CommonService).
func (s *RobnoPrometResource) GetRobneGrupeComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetRobneGrupeComboValues(ctx)
}

func (s *RobnoPrometResource) setTableFields() {
	// Initialize table fields here
	s.prometGrupeArtikalaTableFields = []domain.Fields{
		{Name: "detalj", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "grupa", Label: "Grupa", Width: "6", Field: "gru", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", Field: "sifra", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "8", Field: "konto", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "28", Field: "naz1"},
		{Name: "jm", Label: "JM", Width: "4", Field: "jm", SkipInSearch: true, TextAlign: "center"},
		{Name: "ulaz", Label: "Količina ulaz", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "izlaz", Label: "Količina izlaz", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "finulaz", Label: "Finansijski ulaz", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "finizlaz", Label: "Finansijski izlaz", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometKupciTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7", Field: "konto", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra kupca", Width: "8", Field: "sifra", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv kupca", Width: "20", Field: "naziv"},
		{Name: "adresa", Label: "Adresa", Width: "16", Field: "adresa"},
		{Name: "mesto", Label: "Mesto", Width: "9", Field: "mesto"},
		{Name: "iznos", Label: "Iznos", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ugrabat", Label: "Ugovoreni rabat", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kasa", Label: "Kasa skonto", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "netrezpdv", Label: "Neto (bez PDV)", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "neto", Label: "Neto", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometDobavljaciTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7", Field: "konto", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra dobavljača", Width: "8", Field: "sifra", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv dobavljača", Width: "18", Field: "naziv"},
		{Name: "adresa", Label: "Adresa", Width: "14", Field: "adresa"},
		{Name: "mesto", Label: "Mesto", Width: "7", Field: "mesto"},
		{Name: "fakvred", Label: "Fakturna vrednost", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "netofakvred", Label: "Neto fakturna vrednost", Width: "11", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ztrovred", Label: "Zavisni troškovi nabavke", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "nabvred", Label: "Nabavna vrednost", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ruc", Label: "RUC (razlika u ceni)", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procruc", Label: "Procenat RUC", Width: "7", SkipInSearch: true, TextAlign: "right"},
		{Name: "vpvred", Label: "VPC vrednost", Width: "10", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometRucTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "2", SkipInSearch: true, TextAlign: "center"},
		{Name: "mag", Label: "Magacin", Width: "5", Field: "mag", SkipInSearch: true, TextAlign: "right"},
		{Name: "brnal", Label: "Broj naloga", Width: "6", Field: "brnal", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "5", Field: "vrd", SkipInSearch: true, TextAlign: "right"},
		{Name: "brdok", Label: "Broj dokumenta", Width: "6", Field: "brdok", SkipInSearch: true},
		{Name: "doniz", Label: "Izvorni dokument", Width: "6", Field: "doniz", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "dadok", SkipInSearch: true, TextAlign: "center"},
		{Name: "dospece", Label: "Dospeće", Width: "8", Field: "dospece", SkipInSearch: true, TextAlign: "center"},
		{Name: "netovrd", Label: "Neto vrednost", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "nabvrd", Label: "Nabavna vrednost", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ruc", Label: "RUC (razlika u ceni)", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procnuc", Label: "Procenat RUC", Width: "6", SkipInSearch: true, TextAlign: "right"},
		{Name: "fkto", Label: "Konto", Width: "7", Field: "fkto", SkipInSearch: true},
		{Name: "fana", Label: "Napomena", Width: "8", Field: "fana", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "18", Field: "naziv"},
		{Name: "sifart", Label: "Šifra artikla", Width: "7", Field: "sifart", SkipInSearch: true, TextAlign: "right"},
		{Name: "komerc", Label: "Komercijalista", Width: "10", Field: "komerc", SkipInSearch: true},
	}
	// Tab 4 - Lager lista: the columns of the legacy TABLE_ROB_QRY_LAGER.
	s.prometRucLagerListaTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "mag", Label: "Magacin", Width: "5", SkipInSearch: true, TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "7", SkipInSearch: true, TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "28"},
		{Name: "jm", Label: "JM", Width: "4", SkipInSearch: true, TextAlign: "center"},
		{Name: "stanje", Label: "Stanje", Width: "9", SkipInSearch: true, TextAlign: "right"},
		{Name: "cena", Label: "Cena", Width: "9", SkipInSearch: true, TextAlign: "right"},
		{Name: "vrednost", Label: "Vrednost", Width: "11", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "grupa", Label: "Grupa", Width: "5", SkipInSearch: true, TextAlign: "right"},
		{Name: "podgrupa", Label: "Podgrupa", Width: "6", SkipInSearch: true, TextAlign: "right"},
		{Name: "serbr", Label: "Serijski broj", Width: "9", SkipInSearch: true},
	}
	// Tab 4, sub-tab 2 - Ulaz/izlaz za period: the columns of the legacy table of the sub-tab (the
	// quantities and the fakturna, rabat and magacinska values before the period, in the period and on
	// the last day; the columns with a total in the legacy table carry IncludeInTotals).
	s.prometRucUlazIzlazTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "mag", Label: "Magacin", Width: "4", SkipInSearch: true, TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "5", SkipInSearch: true, TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "16"},
		{Name: "jm", Label: "JM", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "koldodat", Label: "Količina do datuma", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ulaz", Label: "Ulaz", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "izlaz", Label: "Izlaz", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "kolnadan", Label: "Količina na dan", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "stanje", Label: "Stanje", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "fakvreddo", Label: "Fakturna vrednost do datuma", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabatdo", Label: "Rabat do datuma", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "magvreddo", Label: "Magacinska vrednost do datuma", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "fakvrednost", Label: "Fakturna vrednost", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "magvrednost", Label: "Magacinska vrednost", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "fakvrednostnadan", Label: "Fakturna vrednost na dan", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "rabatnadan", Label: "Rabat na dan", Width: "6", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "magvrednostnadan", Label: "Magacinska vrednost na dan", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "pro", Label: "Proizvođač", Width: "7"},
		{Name: "proizsifra", Label: "Šifra proizvođača", Width: "6", SkipInSearch: true},
		{Name: "komercopis", Label: "Komercijalni opis", Width: "10"},
	}
	// Tab 4, sub-tab 3 - RUC po magacinima: the columns of the legacy TABLE_ROB_QRY_RUC (the columns with a
	// total in the legacy table carry IncludeInTotals).
	s.prometRucMagaciniTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "mag", Label: "Magacin", Width: "4", SkipInSearch: true, TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "5", SkipInSearch: true, TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "20"},
		{Name: "jm", Label: "J.M.", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "kolic", Label: "Količina", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "netofvred", Label: "Neto fakturna vrednost", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "prosfc", Label: "Prosečna neto fakturna cena", Width: "6", SkipInSearch: true, TextAlign: "right"},
		{Name: "nabvred", Label: "Nabavna vrednost", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "prosnc", Label: "Prosečna neto nabavna cena", Width: "6", SkipInSearch: true, TextAlign: "right"},
		{Name: "ruc", Label: "Vrednost RUC", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procruc", Label: "Procenat RUC", Width: "5", SkipInSearch: true, TextAlign: "right"},
		{Name: "procu", Label: "Procenat učešća u RUC", Width: "5", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	// Tab 4, sub-tab 4 - Izlazne fakture: the columns of the legacy TABLE_ROB_QRY_RUCDOK (the columns with
	// a total in the legacy table carry IncludeInTotals).
	s.prometRucFaktureTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "mag", Label: "Magacin", Width: "4", SkipInSearch: true, TextAlign: "right"},
		{Name: "brnal", Label: "Broj naloga", Width: "6", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "6", SkipInSearch: true, TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "4", SkipInSearch: true, TextAlign: "right"},
		{Name: "brdok", Label: "Broj dokumenta", Width: "5", SkipInSearch: true},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "6", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "6", SkipInSearch: true, TextAlign: "center"},
		{Name: "dospece", Label: "Dospeće", Width: "6", SkipInSearch: true, TextAlign: "center"},
		{Name: "netofvred", Label: "Neto fakturna vrednost", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "nabvred", Label: "Nabavna vrednost", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ruc", Label: "Vrednost RUC", Width: "7", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procruc", Label: "Procenat RUC", Width: "5", SkipInSearch: true, TextAlign: "right"},
		{Name: "fkto", Label: "Konto", Width: "4", SkipInSearch: true},
		{Name: "fana", Label: "Šifra", Width: "4", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv kupca", Width: "16"},
		{Name: "sifkom", Label: "Šifra kom.", Width: "3", SkipInSearch: true, TextAlign: "right"},
		{Name: "komerc", Label: "Komercijalista", Width: "10"},
	}
	s.prometGradilisteTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7", Field: "konto", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra gradilišta", Width: "10", Field: "sifra", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv gradilišta", Width: "30", Field: "naziv"},
		{Name: "iznos", Label: "Iznos", Width: "12", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
	s.prometGradilisteVpcNcTableFields = []domain.Fields{
		{Name: "detail", Label: "Detalj", Width: "3", SkipInSearch: true, TextAlign: "center"},
		{Name: "rbr", Label: "Redni broj", Width: "4", Field: "rbr", SkipInSearch: true, TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7", Field: "konto", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra gradilišta", Width: "8", Field: "sifra", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv gradilišta", Width: "24", Field: "naziv"},
		{Name: "vpiznos", Label: "VP iznos", Width: "14", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "nciznos", Label: "NC iznos", Width: "14", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "razlika", Label: "Razlika (VP-NC)", Width: "14", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
	}
}

// robnoPrometHeaders returns the columns of a robno promet report: the columns of the grid and, in the
// print, the same columns without the detail column (the drill-down button of the grid rows).
func robnoPrometHeaders(fields []domain.Fields, isPrint bool) []domain.Fields {
	if !isPrint {
		return fields
	}
	headers := make([]domain.Fields, 0, len(fields))
	for _, field := range fields {
		if field.Name == "detail" || field.Name == "detalj" {
			continue
		}
		headers = append(headers, field)
	}
	return headers
}

// robnoPrometRowFields returns the cells of one row of a robno promet report: in the grid the row
// starts with the cell of the detail column, the print has no such column.
func robnoPrometRowFields(isPrint bool, cells ...string) []string {
	if isPrint {
		return cells
	}
	return append([]string{"🔽"}, cells...)
}

func (s *RobnoPrometResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

func (s *RobnoPrometResource) GetPrometArtiklaTableFields() []domain.Fields {
	return s.prometGrupeArtikalaTableFields
}
func (s *RobnoPrometResource) GetPrometKupcaTableFields() []domain.Fields {
	return s.prometKupciTableFields
}
func (s *RobnoPrometResource) GetPrometOdDobavljacaTableFields() []domain.Fields {
	return s.prometDobavljaciTableFields
}
func (s *RobnoPrometResource) GetPrometRucTableFields() []domain.Fields {
	return s.prometRucTableFields
}
func (s *RobnoPrometResource) GetPrometGradilistaTableFields() []domain.Fields {
	return s.prometGradilisteTableFields
}
func (s *RobnoPrometResource) GetPrometGradilisteVpcNcTableFields() []domain.Fields {
	return s.prometGradilisteVpcNcTableFields
}
func (s *RobnoPrometResource) GetPrometRucLagerListaTableFields() []domain.Fields {
	return s.prometRucLagerListaTableFields
}

// GetPrometRucUlazIzlazTableFields returns the columns of the "Ulaz/izlaz za period" sub-tab (tab 4,
// sub-tab 2).
func (s *RobnoPrometResource) GetPrometRucUlazIzlazTableFields() []domain.Fields {
	return s.prometRucUlazIzlazTableFields
}

// GetPrometRucMagaciniTableFields returns the columns of the "RUC po magacinima" sub-tab (tab 4, sub-tab
// 3).
func (s *RobnoPrometResource) GetPrometRucMagaciniTableFields() []domain.Fields {
	return s.prometRucMagaciniTableFields
}

// GetPrometRucFaktureTableFields returns the columns of the "Izlazne fakture" sub-tab (tab 4, sub-tab 4).
func (s *RobnoPrometResource) GetPrometRucFaktureTableFields() []domain.Fields {
	return s.prometRucFaktureTableFields
}

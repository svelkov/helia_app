package robno

import (
	"context"
	"fmt"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

type RobnoPrometService interface {
	GetPrometArtikala(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error
	GetPrometPoKupcima(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error
	GetNabavkeOdDobavljaca(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error
	GetPrometRucLagerLista(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error
	GetPrometGradilista(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error
	GetPrometGradilisteVpcNc(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error
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
		prometGradilisteVpcNcTableFields: []domain.Fields{},
	}
	rs.setTableFields()
	return rs
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

// GetPrometArtikala renders "Promet artikala po grupama za period" (robno promet, tab 1): one row
// per article with the quantities and the financial values summed for the period. It is the
// Go/PostgreSQL translation of the WinDev procedure Obrada().
//
// The first pass (getTotalRecords = true) runs the whole grouped query to publish the number of
// articles and the footer totals; the second pass loads only the requested page.
func (s *RobnoPrometResource) GetPrometArtikala(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometGrupeArtikalaTableFields
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
	if !getTotalRecords && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
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
		return nil
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				"plus.gif",
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
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometPoKupcima renders "Promet artikala po kupcima za period" (robno promet, tab 2): one row
// per partner (fkto + fana) with the iznos, the rabat, the ugovoreni rabat, the kasa skonto, the
// neto bez PDV and the neto with PDV of its sales staves.
//
// The query already returns the aggregated partners, so, like GetPrometArtikala: the first pass
// (getTotalRecords = true) runs it without LIMIT to publish the number of partners and the footer
// totals, the second pass loads only the requested page.
func (s *RobnoPrometResource) GetPrometPoKupcima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometKupciTableFields
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
	if !getTotalRecords && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
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
		return nil
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				"plus.gif",
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
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetNabavkeOdDobavljaca renders "Nabavka po dobavljačima" (robno promet, tab 3): one row per
// supplier (fkto + fana) with the fakturna vrednost, the rabat, the neto fakturna vrednost, the
// zavisni troškovi nabavke, the nabavna vrednost, the RUC and the VPC vrednost of its purchase
// staves. It is the Go/PostgreSQL translation of the WinDev procedure Obrada().
//
// The query already returns the aggregated suppliers, so, like the two other reports of the
// module: the first pass (getTotalRecords = true) runs it without LIMIT to publish the number of
// suppliers and the footer totals, the second pass loads only the requested page.
func (s *RobnoPrometResource) GetNabavkeOdDobavljaca(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometDobavljaciTableFields
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
	if !getTotalRecords && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
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
		return nil
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				"plus.gif",
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
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometGradilista renders "Izveštaj zaduženja gradilišta" (robno promet, tab 4): one row per
// gradilište (pkto + pana of the document) with the sum of the iznos of the staves of its
// trebovanje documents. It is the Go/PostgreSQL translation of the WinDev procedure Obrada().
//
// The legacy selected the documents with `(RDOK.VRD = nVRD OR RDOK.VRD = nVRD1)`, i.e. the two
// document types "TREBOVANJE ZADUŽENJE" (126 and 127); both are the only types whose DOKVRSTA.GRPDOK
// is 'GRD' (gradilište) in every period, so the report filters on that group instead of on the two
// hard-coded numbers (the Go form has no VRD/U1-U8 input, the same decision as in GetPrometPoKupcima).
// The displayed naziv is FKPL.NAZIV of the gradilište (the legacy HReadSeekFirst(FKPL,...)), NOT
// RSIF.NAZIV, and the legacy does not round: it simply accumulates RPRO.IZNOS, so the sum is done in
// SQL numeric.
//
// The query already returns the aggregated gradilišta, so, like the other reports of the module: the
// first pass (getTotalRecords = true) runs it without LIMIT to publish the number of gradilišta and
// the footer total, the second pass loads only the requested page.
func (s *RobnoPrometResource) GetPrometGradilista(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometGradilisteTableFields
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
	if !getTotalRecords && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
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
		return nil
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				"plus.gif",
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetPrometGradilisteVpcNc renders "Izveštaj zaduženja gradilišta VPC-NC" (robno promet, tab 5): one
// row per gradilište (pkto + pana of the document) with, for the staves of its trebovanje documents,
// the VP amount (kolic * cena), the NC amount (kolic * ncena) and their difference. It is the
// Go/PostgreSQL translation of the WinDev procedure Obrada().
//
// It differs from GetPrometGradilista only in the accumulated amounts: the legacy sums
// RPRO.KOLIC * RPRO.CENA and RPRO.KOLIC * RPRO.NCENA without rounding each stavka and displays
// "VP iznos - NC iznos" of the two accumulated totals, so both sums and the difference are computed
// in SQL numeric over the same gradilište documents (grpdok = 'GRD', the two document types the
// legacy selected with `(RDOK.VRD = nVRD OR RDOK.VRD = nVRD1)`) and the naziv is FKPL.NAZIV with the
// ” fallback.
//
// The query already returns the aggregated gradilišta, so, like the other reports of the module: the
// first pass (getTotalRecords = true) runs it without LIMIT to publish the number of gradilišta and
// the footer totals, the second pass loads only the requested page.
func (s *RobnoPrometResource) GetPrometGradilisteVpcNc(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoPrometParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prometGradilisteVpcNcTableFields
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
	if !getTotalRecords && pageSize > 0 {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoPrometRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
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
		return nil
	}

	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				"plus.gif",
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Vpiznos, 2),
				common.FormatNumberWithSystemLocale(entity.Nciznos, 2),
				common.FormatNumberWithSystemLocale(entity.Razlika, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

func (s *RobnoPrometResource) GetPrometRucLagerLista(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams) error {
	//TODO implememt
	return nil
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
		{Name: "yrd", Label: "Vrsta dokumenta", Width: "5", Field: "yrd", SkipInSearch: true, TextAlign: "right"},
		{Name: "brdok", Label: "Broj dokumenta", Width: "6", Field: "brdok", SkipInSearch: true},
		{Name: "doniz", Label: "Izvorni dokument", Width: "6", Field: "doniz", SkipInSearch: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", Field: "dadok", SkipInSearch: true, TextAlign: "center"},
		{Name: "dospece", Label: "Dospeće", Width: "8", Field: "dospece", SkipInSearch: true, TextAlign: "center"},
		{Name: "netopyrd", Label: "Neto vrednost", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "navyrd", Label: "Nabavna vrednost", Width: "9", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "ruc", Label: "RUC (razlika u ceni)", Width: "8", SkipInSearch: true, TextAlign: "right", IncludeInTotals: true},
		{Name: "procnuc", Label: "Procenat RUC", Width: "6", SkipInSearch: true, TextAlign: "right"},
		{Name: "fito", Label: "Konto", Width: "7", Field: "fito", SkipInSearch: true},
		{Name: "fana", Label: "Napomena", Width: "8", Field: "fana", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "18", Field: "naziv"},
		{Name: "sifona", Label: "Šifra artikla", Width: "7", Field: "sifona", SkipInSearch: true, TextAlign: "right"},
		{Name: "komerc", Label: "Komercijalista", Width: "10", Field: "komerc", SkipInSearch: true},
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

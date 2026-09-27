// Package common holds the shared services of the project. CommonService collects every combo
// (lookup) that is used by more than one screen, so the queries are written once and injected
// into the domain services instead of being duplicated in each of them.
package common

import (
	"context"
	"fmt"

	"helia/internal/domain"
	"helia/internal/repository"

	helcommon "helia/internal/common"
)

// CommonService is the contract of the shared lookups. It is injected into the domain services
// (and can be injected into the handlers as well) so they do not repeat the same queries.
type CommonService interface {
	// Magacini
	GetMagacinComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetMagacinByMagComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)

	// Vrste naloga (tipdok)
	GetTipdokComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipdokIDComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipdokFinComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipdokIDByCode(ctx context.Context, tipdok string) (int64, error)

	// Vrste dokumenata (dokvrsta)
	GetVrstaDokumentaComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)

	// Organizacione jedinice i mesta troška
	GetOrgJedComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetMestoTroskaComboValues(ctx context.Context, idOrgjed int64, opts ...ComboOption) ([]domain.ComboItem, error)

	// Robne grupe i podgrupe
	GetRobneGrupeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetRobnePodgrupeComboValues(ctx context.Context, gru int, opts ...ComboOption) ([]domain.ComboItem, error)

	// Ostali šifarnici
	GetValuteComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetKomercijalistiComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetMestoIsporukeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetBankeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipoviAnalitikeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetJediniceMereComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)

	// Poreske knjige
	GetTipovePoreskihKnjigaComboValues(ctx context.Context, vkTip string, opts ...ComboOption) ([]domain.ComboItem, error)

	// Next broj naloga (finansijski nalog - fnal, robni nalog - rnal)
	GetNextFnalNalog(ctx context.Context, tipdok string) (int64, error)
	GetNextRnalNalog(ctx context.Context, tipdok string) (int64, error)
}

// CommonResource implements CommonService.
type CommonResource struct {
	magaciniRepo       repository.BaseRepository[domain.Magacini]
	tipdokRepo         repository.BaseRepository[domain.Tipdok]
	dokvrstaRepo       repository.BaseRepository[domain.Dokvrsta]
	orgjedRepo         repository.BaseRepository[domain.Orgjed]
	mestotrRepo        repository.BaseRepository[domain.Mestotr]
	rgruRepo           repository.BaseRepository[domain.Rgru]
	rpgruRepo          repository.BaseRepository[domain.Rpgru]
	valuteRepo         repository.BaseRepository[domain.Valute]
	komercijalistiRepo repository.BaseRepository[domain.Komercijalisti]
	fispRepo           repository.BaseRepository[domain.Fisp]
	bankeRepo          repository.BaseRepository[domain.Banke]
	tipanalitikeRepo   repository.BaseRepository[domain.Tipanalitike]
	jedmereRepo        repository.BaseRepository[domain.Jedmere]
	fvknjracRepo       repository.BaseRepository[domain.Fvknjrac]
	fnalRepo           repository.BaseRepository[domain.Fnal]
	rnalRepo           repository.BaseRepository[domain.RobnoDokumentaDto]
}

// NewCommonService creates the shared service with the repositories of the šifarnici.
func NewCommonService(
	magaciniRepo repository.BaseRepository[domain.Magacini],
	tipdokRepo repository.BaseRepository[domain.Tipdok],
	dokvrstaRepo repository.BaseRepository[domain.Dokvrsta],
	orgjedRepo repository.BaseRepository[domain.Orgjed],
	mestotrRepo repository.BaseRepository[domain.Mestotr],
	rgruRepo repository.BaseRepository[domain.Rgru],
	rpgruRepo repository.BaseRepository[domain.Rpgru],
	valuteRepo repository.BaseRepository[domain.Valute],
	komercijalistiRepo repository.BaseRepository[domain.Komercijalisti],
	fispRepo repository.BaseRepository[domain.Fisp],
	bankeRepo repository.BaseRepository[domain.Banke],
	tipanalitikeRepo repository.BaseRepository[domain.Tipanalitike],
	jedmereRepo repository.BaseRepository[domain.Jedmere],
	fvknjracRepo repository.BaseRepository[domain.Fvknjrac],
	fnalRepo repository.BaseRepository[domain.Fnal],
	rnalRepo repository.BaseRepository[domain.RobnoDokumentaDto],
) *CommonResource {
	return &CommonResource{
		magaciniRepo:       magaciniRepo,
		tipdokRepo:         tipdokRepo,
		dokvrstaRepo:       dokvrstaRepo,
		orgjedRepo:         orgjedRepo,
		mestotrRepo:        mestotrRepo,
		rgruRepo:           rgruRepo,
		rpgruRepo:          rpgruRepo,
		valuteRepo:         valuteRepo,
		komercijalistiRepo: komercijalistiRepo,
		fispRepo:           fispRepo,
		bankeRepo:          bankeRepo,
		tipanalitikeRepo:   tipanalitikeRepo,
		jedmereRepo:        jedmereRepo,
		fvknjracRepo:       fvknjracRepo,
		fnalRepo:           fnalRepo,
		rnalRepo:           rnalRepo,
	}
}

//
// Magacini
//

// GetMagacinComboValues returns "mag - opis" of every magacin of the current period, keyed by
// magaciniid (the key most screens store in their magaciniid column).
func (s *CommonResource) GetMagacinComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select magaciniid, mag, opis from magacini", true)
	hasGod, hasKar := s.magaciniRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("mag")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.magaciniRepo, query, args,
		func(m domain.Magacini) string { return fmt.Sprintf("%d", m.MagaciniID) },
		func(m domain.Magacini) string { return fmt.Sprintf("%d - %s", m.Mag, m.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetMagacinByMagComboValues returns "mag - opis" of every magacin of the current period, keyed by
// mag (the screens that filter by mag, e.g. robno stanje/promet, need this flavour).
func (s *CommonResource) GetMagacinByMagComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select mag, opis from magacini", true)
	hasGod, hasKar := s.magaciniRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("mag")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.magaciniRepo, query, args,
		func(m domain.Magacini) string { return fmt.Sprintf("%d", m.Mag) },
		func(m domain.Magacini) string { return fmt.Sprintf("%d - %s", m.Mag, m.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

//
// Vrste naloga (tipdok)
//

// GetTipdokComboValues returns "tipdok - opis" of the current period keyed by the tipdok code
// (what documents store in their tipdok column).
func (s *CommonResource) GetTipdokComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select tipdok, opis from tipdok", true)
	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("tipdok::numeric")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.tipdokRepo, query, args,
		func(t domain.Tipdok) string { return t.TipDok },
		func(t domain.Tipdok) string { return fmt.Sprintf("%s - %s", t.TipDok, t.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetTipdokIDComboValues is the same list keyed by idtipdok (used by the screens that filter by
// the tipdok id).
func (s *CommonResource) GetTipdokIDComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select idtipdok, tipdok, opis from tipdok", true)
	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("tipdok::numeric")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.tipdokRepo, query, args,
		func(t domain.Tipdok) string { return fmt.Sprintf("%d", t.IDTipDok) },
		func(t domain.Tipdok) string { return fmt.Sprintf("%s - %s", t.TipDok, t.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetTipdokFinComboValues returns the vrste naloga that can be used for knjiženje
// (grpdok FIN or SVI), keyed by the tipdok code.
func (s *CommonResource) GetTipdokFinComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select tipdok, opis from tipdok", true)
	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddCustomCondition("(grpdok = 'FIN' OR grpdok = 'SVI')")
	qb.AddOrderBy("tipdok::numeric")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.tipdokRepo, query, args,
		func(t domain.Tipdok) string { return t.TipDok },
		func(t domain.Tipdok) string { return fmt.Sprintf("%s - %s", t.TipDok, t.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetTipdokIDByCode resolves the idtipdok of a vrsta naloga from its tipdok code in the current
// period (the screens that store idtipdok but receive the code need this conversion).
func (s *CommonResource) GetTipdokIDByCode(ctx context.Context, tipdok string) (int64, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return 0, err
	}
	qb := helcommon.NewQueryBuilder("select idtipdok, tipdok, opis from tipdok", true)
	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	query, args := qb.Build()
	entities, err := s.tipdokRepo.GetAllCustom(ctx, query, "", args, "", "")
	if err != nil {
		return 0, err
	}
	if entities == nil || len(*entities) == 0 {
		return 0, fmt.Errorf("%s: %s", helcommon.ErrNoDataFound, tipdok)
	}
	return int64((*entities)[0].IDTipDok), nil
}

//
// Vrste dokumenata (dokvrsta)
//

// GetVrstaDokumentaComboValues returns "vrd - opis" of every vrsta dokumenta of the current period.
func (s *CommonResource) GetVrstaDokumentaComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select vrd, opis from dokvrsta", true)
	hasGod, hasKar := s.dokvrstaRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("vrd")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.dokvrstaRepo, query, args,
		func(d domain.Dokvrsta) string { return fmt.Sprintf("%d", d.Vrd) },
		func(d domain.Dokvrsta) string { return fmt.Sprintf("%d - %s", d.Vrd, d.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

//
// Organizacione jedinice i mesta troška
//

// GetOrgJedComboValues returns "ojozn - naziv" of every organizaciona jedinica of the current period.
func (s *CommonResource) GetOrgJedComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select idorgjed, ojozn, naziv from orgjed", true)
	hasGod, hasKar := s.orgjedRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("ojozn")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.orgjedRepo, query, args,
		func(o domain.Orgjed) string { return fmt.Sprintf("%d", o.IDOrgjed) },
		func(o domain.Orgjed) string { return fmt.Sprintf("%s - %s", o.OjOzn, o.Naziv) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetMestoTroskaComboValues returns "mtroska - opis" of the mesta troška of one organizaciona
// jedinica (idOrgjed = 0 returns the mesta troška of every organizaciona jedinica).
func (s *CommonResource) GetMestoTroskaComboValues(ctx context.Context, idOrgjed int64, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select mestotrid, mtroska, opis from mestotr", true)
	hasGod, hasKar := s.mestotrRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	if idOrgjed != 0 {
		qb.AddEqual("idorgjed", idOrgjed)
	}
	qb.AddOrderBy("mtroska")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.mestotrRepo, query, args,
		func(m domain.Mestotr) string { return fmt.Sprintf("%d", m.MestoTrID) },
		func(m domain.Mestotr) string { return fmt.Sprintf("%s - %s", m.Mtroska, m.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

//
// Robne grupe i podgrupe
//

// GetRobneGrupeComboValues returns "gru - naziv" of every robna grupa of the current period.
func (s *CommonResource) GetRobneGrupeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select gru, naziv from rgru", true)
	hasGod, hasKar := s.rgruRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("gru")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.rgruRepo, query, args,
		func(g domain.Rgru) string { return fmt.Sprintf("%d", g.Gru) },
		func(g domain.Rgru) string { return fmt.Sprintf("%d - %s", g.Gru, g.Naziv) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetRobnePodgrupeComboValues returns "pgru - naziv" of the robne podgrupe of one robna grupa
// (gru = 0 returns the podgrupe of every grupa).
func (s *CommonResource) GetRobnePodgrupeComboValues(ctx context.Context, gru int, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select gru, pgru, naziv from rpgru", true)
	hasGod, hasKar := s.rpgruRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	if gru != 0 {
		qb.AddEqual("gru", gru)
	}
	qb.AddOrderBy("gru, pgru")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.rpgruRepo, query, args,
		func(p domain.Rpgru) string { return fmt.Sprintf("%d", p.Pgru) },
		func(p domain.Rpgru) string { return fmt.Sprintf("%d - %s", p.Pgru, p.Naziv) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

//
// Ostali šifarnici
//

// GetValuteComboValues returns "sifval - naziv" of every valuta of the current period.
func (s *CommonResource) GetValuteComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select idvalute, sifval, naziv from valute", true)
	hasGod, hasKar := s.valuteRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("sifval")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.valuteRepo, query, args,
		func(v domain.Valute) string { return fmt.Sprintf("%d", v.IDValute) },
		func(v domain.Valute) string { return fmt.Sprintf("%d - %s", v.Sifval, v.Naziv.String) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetKomercijalistiComboValues returns "sifkom - imeprezime" of every komercijalista of the
// current period.
func (s *CommonResource) GetKomercijalistiComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select komid, sifkom, imeprezime from komercijalisti", true)
	hasGod, hasKar := s.komercijalistiRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("sifkom")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.komercijalistiRepo, query, args,
		func(k domain.Komercijalisti) string { return fmt.Sprintf("%d", k.KomID) },
		func(k domain.Komercijalisti) string { return fmt.Sprintf("%d - %s", k.Sifkom, k.ImePrezime) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetMestoIsporukeComboValues returns "mi - naziv" of every mesto isporuke (fisp) of the current
// period.
func (s *CommonResource) GetMestoIsporukeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select fispid, mi, naziv from fisp", true)
	hasGod, hasKar := s.fispRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("mi")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.fispRepo, query, args,
		func(m domain.Fisp) string { return fmt.Sprintf("%d", m.FispID) },
		func(m domain.Fisp) string { return fmt.Sprintf("%d - %s", m.MI, m.Naziv) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetBankeComboValues returns "bnkcod - banka" of every banka of the current period.
func (s *CommonResource) GetBankeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select idbanke, banka, bnkcod from banke", true)
	hasGod, hasKar := s.bankeRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("banka")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.bankeRepo, query, args,
		func(b domain.Banke) string { return fmt.Sprintf("%d", b.IDBanke) },
		func(b domain.Banke) string { return fmt.Sprintf("%s - %s", b.BnkCod, b.Banka) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetTipoviAnalitikeComboValues returns the tipovi analitike (partners, komercijalisti, ...) that
// can be attached to a konto.
func (s *CommonResource) GetTipoviAnalitikeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	qb := helcommon.NewQueryBuilder("select tipanalitikeid, naziv from tipanalitike", true)
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		hasGod, hasKar := s.tipanalitikeRepo.GetHasGodHasKar()
		qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	}
	qb.AddOrderBy("tipanalitikeid")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.tipanalitikeRepo, query, args,
		func(t domain.Tipanalitike) string { return fmt.Sprintf("%d", t.TipanalitikeID) },
		func(t domain.Tipanalitike) string { return t.Naziv })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetJediniceMereComboValues returns "jm - opis" of every jedinica mere of the current period.
func (s *CommonResource) GetJediniceMereComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select jm, opis from jedmere", true)
	hasGod, hasKar := s.jedmereRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("jm")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.jedmereRepo, query, args,
		func(j domain.Jedmere) string { return j.JM },
		func(j domain.Jedmere) string { return fmt.Sprintf("%s - %s", j.JM, j.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetTipovePoreskihKnjigaComboValues returns "vkrbr - opis" of the vrste poreskih knjiga
// (fvknjrac) for the given vktip (KIR/KPR) of the current period. The caller can append its own
// "999 - Sve knjige" option.
func (s *CommonResource) GetTipovePoreskihKnjigaComboValues(ctx context.Context, vkTip string, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	qb := helcommon.NewQueryBuilder("select vkrbr, opis from fvknjrac", true)
	hasGod, hasKar := s.fvknjracRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddEqual("vktip", vkTip)
	qb.AddOrderBy("vktip asc, vkrbr asc")
	query, args := qb.Build()
	items, err := comboQuery(ctx, s.fvknjracRepo, query, args,
		func(f domain.Fvknjrac) string { return fmt.Sprintf("%d", f.VkRbr) },
		func(f domain.Fvknjrac) string { return fmt.Sprintf("%d - %s", f.VkRbr, f.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

//
// Next broj naloga
//

// GetNextFnalNalog returns the next broj naloga (max(nalog) + 1) of the given vrsta naloga in the
// current period for the financial nalozi (fnal).
func (s *CommonResource) GetNextFnalNalog(ctx context.Context, tipdok string) (int64, error) {
	return nextNalogQuery(ctx, s.fnalRepo, "fnal", tipdok, func(f domain.Fnal) int64 { return f.Nalog })
}

// GetNextRnalNalog returns the next broj naloga (max(nalog) + 1) of the given vrsta naloga in the
// current period for the robni nalozi (rnal).
func (s *CommonResource) GetNextRnalNalog(ctx context.Context, tipdok string) (int64, error) {
	return nextNalogQuery(ctx, s.rnalRepo, "rnal", tipdok, func(r domain.RobnoDokumentaDto) int64 { return int64(r.Nalog) })
}

//
// Helpers
//

// ComboOption tweaks the produced combo list (e.g. the legacy "-" ("nothing selected") option the
// financial screens prepend).
type ComboOption func(*comboConfig)

type comboConfig struct {
	emptyFirst bool
	emptyKey   string
	emptyValue string
}

// WithEmptyOption prepends the legacy "-" option used when nothing is selected.
func WithEmptyOption() ComboOption {
	return func(c *comboConfig) {
		c.emptyFirst = true
		c.emptyKey = "-"
		c.emptyValue = "-"
	}
}

// WithEmptyOptionValues prepends a custom "nothing selected" option.
func WithEmptyOptionValues(key, value string) ComboOption {
	return func(c *comboConfig) {
		c.emptyFirst = true
		c.emptyKey = key
		c.emptyValue = value
	}
}

// withOptions applies the options to the produced list.
func withOptions(items []domain.ComboItem, opts []ComboOption) []domain.ComboItem {
	cfg := comboConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if !cfg.emptyFirst {
		return items
	}
	result := make([]domain.ComboItem, 0, len(items)+1)
	result = append(result, domain.ComboItem{Key: cfg.emptyKey, Value: cfg.emptyValue})
	result = append(result, items...)
	return result
}

// comboQuery runs a combo query and maps every row to a domain.ComboItem.
func comboQuery[T any](ctx context.Context, repo repository.BaseRepository[T], query string, args []any, key func(T) string, value func(T) string) ([]domain.ComboItem, error) {
	entities, err := repo.GetAllCustom(ctx, query, "", args, "", "")
	if err != nil {
		return nil, err
	}
	items := make([]domain.ComboItem, 0, len(*entities))
	for _, entity := range *entities {
		items = append(items, domain.ComboItem{Key: key(entity), Value: value(entity)})
	}
	return items, nil
}

// nextNalogQuery runs the "next broj naloga" query (max(nalog) + 1 of the given tipdok in the
// current period) against any nalog table (fnal, rnal).
func nextNalogQuery[T any](ctx context.Context, repo repository.BaseRepository[T], table, tipdok string, nalog func(T) int64) (int64, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return 0, err
	}
	qb := helcommon.NewQueryBuilder(fmt.Sprintf("select coalesce(max(nalog), 0) + 1 as nalog from %s", table), true)
	hasGod, hasKar := repo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	query, args := qb.Build()
	entities, err := repo.GetAllCustom(ctx, query, "", args, "", "")
	if err != nil {
		return 0, err
	}
	if entities == nil || len(*entities) == 0 {
		return 1, nil
	}
	return nalog((*entities)[0]), nil
}

// sessionFrom returns the current user session (god/kar/firma are needed by almost every combo).
func sessionFrom(ctx context.Context) (*domain.UserSession, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("no user session found")
	}
	return session, nil
}

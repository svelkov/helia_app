// Package common holds the shared services of the project. CommonService collects every combo
// (lookup) that is used by more than one screen, so the queries are written once and injected
// into the domain services instead of being duplicated in each of them.
package common

import (
	"context"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"

	helcommon "helia/internal/common"
)

// CommonService is the contract of the shared lookups. It is injected into the domain services
// (and can be injected into the handlers as well) so they do not repeat the same queries.
type CommonService interface {
	// Magacini
	GetMagacinComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetMagacinByMagComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetMagacinKorisnikaComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)

	// Vrste naloga (tipdok)
	GetTipdokComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipdokIDComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipdokFinComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error)
	GetTipdokIDByCode(ctx context.Context, tipdok string) (int64, error)
	GetTipdokByCode(ctx context.Context, tipdok string) (domain.Tipdok, error)

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
	GetFnalByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Fnal, bool, error)
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
	b := querybuilder.New("select magaciniid, mag, opis from magacini").OrderBy("mag")
	items, err := comboQuery(ctx, &s.magaciniRepo, b, "", opts,
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
	b := querybuilder.New("select mag, opis from magacini").OrderBy("mag")
	items, err := comboQuery(ctx, &s.magaciniRepo, b, "", opts,
		func(m domain.Magacini) string { return fmt.Sprintf("%d", m.Mag) },
		func(m domain.Magacini) string { return fmt.Sprintf("%d - %s", m.Mag, m.Opis) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetMagacinKorisnikaComboValues returns "mag - opis" of the magacini of the current period the user
// of the session may work with (maguser.iduser, the login of the user, like the legacy
// ROB_QRY_MAGUSER), keyed by magaciniid. The filters of the options are on the joined tables, so
// their columns are qualified, e.g. WithIn("magacini.tipmag", "V", "M", "D"),
// WithIn("magacini.mag", 1, 2).
func (s *CommonResource) GetMagacinKorisnikaComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	session, err := sessionFrom(ctx)
	if err != nil {
		return nil, err
	}
	// Where (not AddEqual): a session without a user name must not see every magacin.
	b := querybuilder.New(`select distinct magacini.magaciniid, magacini.mag, magacini.opis
		from magacini inner join maguser on maguser.mag = magacini.mag`).
		Where("upper(maguser.iduser) = ?", strings.ToUpper(strings.TrimSpace(session.UserName))).
		OrderBy("magacini.mag")
	items, err := comboQuery(ctx, &s.magaciniRepo, b, "magacini", opts,
		func(m domain.Magacini) string { return fmt.Sprintf("%d", m.MagaciniID) },
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
	b := querybuilder.New("select tipdok, opis from tipdok").OrderByExpr("tipdok::numeric")
	items, err := comboQuery(ctx, &s.tipdokRepo, b, "", opts,
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
	b := querybuilder.New("select idtipdok, tipdok, opis from tipdok").OrderByExpr("tipdok::numeric")
	items, err := comboQuery(ctx, &s.tipdokRepo, b, "", opts,
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
	return s.GetTipdokComboValues(ctx, append([]ComboOption{WithGrupeDokumenata("FIN", "SVI")}, opts...)...)
}

// GetTipdokIDByCode resolves the idtipdok of a vrsta naloga from its tipdok code in the current
// period (the screens that store idtipdok but receive the code need this conversion).
func (s *CommonResource) GetTipdokIDByCode(ctx context.Context, tipdok string) (int64, error) {
	t, err := s.GetTipdokByCode(ctx, tipdok)
	if err != nil {
		return 0, err
	}
	return int64(t.IDTipDok), nil
}

// GetTipdokByCode returns the vrsta naloga of the tipdok code in the current period with its groups
// (grpdok), the vrste dokumenata (grpvrd) and the magacini (magacin) it allows.
func (s *CommonResource) GetTipdokByCode(ctx context.Context, tipdok string) (domain.Tipdok, error) {
	// Where (not AddEqual): an empty code must not return an arbitrary tipdok.
	b := querybuilder.New(`select idtipdok, tipdok, opis, coalesce(grpdok, '') as grpdok,
		coalesce(grpvrd, '') as grpvrd, coalesce(magacin, '') as magacin from tipdok`).
		Where("tipdok = ?", tipdok)
	if err := s.tipdokRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return domain.Tipdok{}, err
	}
	entities, err := s.tipdokRepo.Select(ctx, b)
	if err != nil {
		return domain.Tipdok{}, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Tipdok{}, fmt.Errorf("%s: %s", helcommon.ErrNoDataFound, tipdok)
	}
	return (*entities)[0], nil
}

//
// Vrste dokumenata (dokvrsta)
//

// GetVrstaDokumentaComboValues returns "vrd - opis" of every vrsta dokumenta of the current period.
func (s *CommonResource) GetVrstaDokumentaComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	b := querybuilder.New("select vrd, opis from dokvrsta").OrderBy("vrd")
	items, err := comboQuery(ctx, &s.dokvrstaRepo, b, "", opts,
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
	b := querybuilder.New("select idorgjed, ojozn, naziv from orgjed").OrderBy("ojozn")
	items, err := comboQuery(ctx, &s.orgjedRepo, b, "", opts,
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
	b := querybuilder.New("select mestotrid, mtroska, opis from mestotr").OrderBy("mtroska")
	if idOrgjed != 0 {
		b.Where("idorgjed = ?", idOrgjed)
	}
	items, err := comboQuery(ctx, &s.mestotrRepo, b, "", opts,
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
	b := querybuilder.New("select gru, naziv from rgru").OrderBy("gru")
	items, err := comboQuery(ctx, &s.rgruRepo, b, "", opts,
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
	b := querybuilder.New("select gru, pgru, naziv from rpgru").OrderBy("gru, pgru")
	if gru != 0 {
		b.Where("gru = ?", gru)
	}
	items, err := comboQuery(ctx, &s.rpgruRepo, b, "", opts,
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
	b := querybuilder.New("select idvalute, sifval, naziv from valute").OrderBy("sifval")
	items, err := comboQuery(ctx, &s.valuteRepo, b, "", opts,
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
	b := querybuilder.New("select komid, sifkom, imeprezime from komercijalisti").OrderBy("sifkom")
	items, err := comboQuery(ctx, &s.komercijalistiRepo, b, "", opts,
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
	b := querybuilder.New("select fispid, mi, naziv from fisp").OrderBy("mi")
	items, err := comboQuery(ctx, &s.fispRepo, b, "", opts,
		func(m domain.Fisp) string { return fmt.Sprintf("%d", m.FispID) },
		func(m domain.Fisp) string { return fmt.Sprintf("%d - %s", m.MI, m.Naziv) })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetBankeComboValues returns "bnkcod - banka" of every banka of the current period.
func (s *CommonResource) GetBankeComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	b := querybuilder.New("select idbanke, banka, bnkcod from banke").OrderBy("banka")
	items, err := comboQuery(ctx, &s.bankeRepo, b, "", opts,
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
	// tipanalitike has no god/kar, so the list is the same for every period and also works before
	// login (the repository asks for the session only for tables with god/kar).
	b := querybuilder.New("select tipanalitikeid, naziv from tipanalitike").OrderBy("tipanalitikeid")
	items, err := comboQuery(ctx, &s.tipanalitikeRepo, b, "", opts,
		func(t domain.Tipanalitike) string { return fmt.Sprintf("%d", t.TipanalitikeID) },
		func(t domain.Tipanalitike) string { return t.Naziv })
	if err != nil {
		return nil, err
	}
	return withOptions(items, opts), nil
}

// GetJediniceMereComboValues returns "jm - opis" of every jedinica mere of the current period.
func (s *CommonResource) GetJediniceMereComboValues(ctx context.Context, opts ...ComboOption) ([]domain.ComboItem, error) {
	b := querybuilder.New("select jm, opis from jedmere").OrderBy("jm")
	items, err := comboQuery(ctx, &s.jedmereRepo, b, "", opts,
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
	b := querybuilder.New("select vkrbr, opis from fvknjrac").
		AddEqual("vktip", vkTip). // optional: an empty vkTip lists the knjige of every vktip
		OrderBy("vktip asc, vkrbr asc")
	items, err := comboQuery(ctx, &s.fvknjracRepo, b, "", opts,
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
	return nextNalogQuery(ctx, &s.fnalRepo, "fnal", tipdok, func(f domain.Fnal) int64 { return f.Nalog })
}

// GetFnalByTipdokNalog returns the financial nalog (fnal) of the vrsta naloga and broj naloga in the
// current period (found false when there is none).
func (s *CommonResource) GetFnalByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Fnal, bool, error) {
	// Where (not AddEqual): both are the key of the nalog; an empty tipdok must not match any nalog.
	b := querybuilder.New("select idfnal, tipdok, nalog, danal from fnal").
		Where("tipdok = ?", tipdok).
		Where("nalog = ?", nalog)
	if err := s.fnalRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return domain.Fnal{}, false, err
	}
	entities, err := s.fnalRepo.Select(ctx, b)
	if err != nil {
		return domain.Fnal{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Fnal{}, false, nil
	}
	return (*entities)[0], true, nil
}

// GetNextRnalNalog returns the next broj naloga (max(nalog) + 1) of the given vrsta naloga in the
// current period for the robni nalozi (rnal).
func (s *CommonResource) GetNextRnalNalog(ctx context.Context, tipdok string) (int64, error) {
	return nextNalogQuery(ctx, &s.rnalRepo, "rnal", tipdok, func(r domain.RobnoDokumentaDto) int64 { return int64(r.Nalog) })
}

//
// Helpers
//

// ComboOption tweaks a combo: it filters the rows of its query (WithEqual, WithIn,
// WithGrupeDokumenata) or changes the produced list (e.g. the legacy "-" ("nothing selected") option
// the financial screens prepend). Every combo of CommonService accepts them, so a screen that needs
// a filtered list passes the filter instead of getting a new function:
//
//	commonSvc.GetTipdokComboValues(ctx, commonsvc.WithGrupeDokumenata("ROB", "SVI"))
//	commonSvc.GetMagacinComboValues(ctx, commonsvc.WithEqual("tipzal", 2), commonsvc.WithEmptyOption())
type ComboOption func(*comboConfig)

type comboConfig struct {
	emptyFirst bool
	emptyKey   string
	emptyValue string
	// filters are added to the where of the query of the combo (with its parameters).
	filters []func(*querybuilder.Builder)
}

// WithEqual keeps the rows whose column field equals value (the condition is skipped for a nil or
// empty value). field is a column of the table of the combo, written by the developer (never user
// input); the value is passed as a parameter of the query.
func WithEqual(field string, value any) ComboOption {
	return func(c *comboConfig) {
		c.filters = append(c.filters, func(b *querybuilder.Builder) { b.AddEqual(field, value) })
	}
}

// WithIn keeps the rows whose column field is one of values (no condition without values), like the
// legacy "FIELD IN ({ipGRP})". field is a column of the table of the combo, written by the developer.
func WithIn[T any](field string, values ...T) ComboOption {
	return func(c *comboConfig) {
		in := make([]any, 0, len(values))
		for _, v := range values {
			in = append(in, v)
		}
		c.filters = append(c.filters, func(b *querybuilder.Builder) { b.AddIn(field, in) })
	}
}

// WithGrupeDokumenata keeps the rows of the groups of the documents (the column grpdok of tipdok and
// dokvrsta), e.g. WithGrupeDokumenata("ROB", "SVI") for the vrste naloga of the robno knjigovodstvo.
func WithGrupeDokumenata(grupe ...string) ComboOption {
	return WithIn("grpdok", grupe...)
}

// applyComboFilters adds the filters of the options to the query of a combo.
func applyComboFilters(b *querybuilder.Builder, opts []ComboOption) {
	cfg := comboConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	for _, filter := range cfg.filters {
		filter(b)
	}
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

// comboQuery limits the combo query b to the period of the session (columns qualified with alias,
// "" for none), adds the filters of the options, runs it and maps every row to a domain.ComboItem.
func comboQuery[T any](ctx context.Context, repo *repository.BaseRepository[T], b *querybuilder.Builder, alias string, opts []ComboOption, key func(T) string, value func(T) string) ([]domain.ComboItem, error) {
	if err := repo.ApplyFilters(ctx, b, alias, nil, ""); err != nil {
		return nil, err
	}
	applyComboFilters(b, opts)
	entities, err := repo.Select(ctx, b)
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
func nextNalogQuery[T any](ctx context.Context, repo *repository.BaseRepository[T], table, tipdok string, nalog func(T) int64) (int64, error) {
	// Where (not AddEqual): the next broj is per vrsta naloga, also for an empty tipdok.
	b := querybuilder.New("select coalesce(max(nalog), 0) + 1 as nalog from "+querybuilder.QuoteIdent(table)).
		Where("tipdok = ?", tipdok)
	if err := repo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return 0, err
	}
	entities, err := repo.Select(ctx, b)
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

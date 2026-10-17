package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

type sqlCall struct {
	query string
	args  []any
}

func recordingDB(calls *[]sqlCall) *db.MockDatabase {
	return &db.MockDatabase{
		SelectContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			*calls = append(*calls, sqlCall{q, a})
			return nil
		},
	}
}

func sessionCtx() context.Context {
	return domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7})
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func newCommon(d db.Database) *CommonResource {
	return NewCommonService(
		*repository.NewBaseRepository[domain.Magacini](d, "magacini"),
		*repository.NewBaseRepository[domain.Tipdok](d, "tipdok"),
		*repository.NewBaseRepository[domain.Dokvrsta](d, "dokvrsta"),
		*repository.NewBaseRepository[domain.Orgjed](d, "orgjed"),
		*repository.NewBaseRepository[domain.Mestotr](d, "mestotr"),
		*repository.NewBaseRepository[domain.Rgru](d, "rgru"),
		*repository.NewBaseRepository[domain.Rpgru](d, "rpgru"),
		*repository.NewBaseRepository[domain.Valute](d, "valute"),
		*repository.NewBaseRepository[domain.Komercijalisti](d, "komercijalisti"),
		*repository.NewBaseRepository[domain.Fisp](d, "fisp"),
		*repository.NewBaseRepository[domain.Banke](d, "banke"),
		*repository.NewBaseRepository[domain.Tipanalitike](d, "tipanalitike"),
		*repository.NewBaseRepository[domain.Jedmere](d, "jedmere"),
		*repository.NewBaseRepository[domain.Fvknjrac](d, "fvknjrac"),
		*repository.NewBaseRepository[domain.Fnal](d, "fnal"),
		*repository.NewBaseRepository[domain.RobnoDokumentaDto](d, "rnal"),
	)
}

func TestCommonServiceSQL(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		run      func(ctx context.Context, s *CommonResource) error
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "combo with tenant and options",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, err := s.GetMagacinComboValues(ctx, WithEqual("tipzal", 2), WithIn("tipmag", "V", "M"))
				return err
			},
			wantSQL:  "select magaciniid, mag, opis from magacini WHERE (god = $1) AND (kar = $2) AND tipzal = $3 AND tipmag IN ($4, $5) ORDER BY mag",
			wantArgs: []any{2026, 7, 2, "V", "M"},
		},
		{
			name: "magacini of the user: qualified tenant and the user is always a condition",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, err := s.GetMagacinKorisnikaComboValues(ctx)
				return err
			},
			wantSQL: "select distinct magacini.magaciniid, magacini.mag, magacini.opis from magacini inner join maguser on maguser.mag = magacini.mag " +
				"WHERE (upper(maguser.iduser) = $1) AND (magacini.god = $2) AND (magacini.kar = $3) ORDER BY magacini.mag",
			wantArgs: []any{"ANA", 2026, 7},
		},
		{
			name: "tipdok combo ordered by expression",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, err := s.GetTipdokFinComboValues(ctx)
				return err
			},
			wantSQL:  "select tipdok, opis from tipdok WHERE (god = $1) AND (kar = $2) AND grpdok IN ($3, $4) ORDER BY tipdok::numeric",
			wantArgs: []any{2026, 7, "FIN", "SVI"},
		},
		{
			name: "tipdok by code keeps an empty code as a condition",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				// The fake has no rows, so "not found" is the expected result.
				if _, err := s.GetTipdokByCode(ctx, ""); err == nil || !strings.Contains(err.Error(), "nema pronadjenih podataka") {
					return fmt.Errorf("want the not-found error, got %v", err)
				}
				return nil
			},
			wantSQL: "select idtipdok, tipdok, opis, coalesce(grpdok, '') as grpdok, coalesce(grpvrd, '') as grpvrd, " +
				"coalesce(magacin, '') as magacin from tipdok WHERE (tipdok = $1) AND (god = $2) AND (kar = $3)",
			wantArgs: []any{"", 2026, 7},
		},
		{
			name: "fnal by tipdok and nalog",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, _, err := s.GetFnalByTipdokNalog(ctx, "10", 5)
				return err
			},
			wantSQL:  "select idfnal, tipdok, nalog, danal from fnal WHERE (tipdok = $1) AND (nalog = $2) AND (god = $3) AND (kar = $4)",
			wantArgs: []any{"10", 5, 2026, 7},
		},
		{
			name: "next rnal nalog",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, err := s.GetNextRnalNalog(ctx, "20")
				return err
			},
			wantSQL:  `select coalesce(max(nalog), 0) + 1 as nalog from "rnal" WHERE (tipdok = $1) AND (god = $2) AND (kar = $3)`,
			wantArgs: []any{"20", 2026, 7},
		},
		{
			name: "mesta troska of one orgjed",
			ctx:  sessionCtx(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, err := s.GetMestoTroskaComboValues(ctx, 4)
				return err
			},
			wantSQL:  "select mestotrid, mtroska, opis from mestotr WHERE (idorgjed = $1) AND (god = $2) AND (kar = $3) ORDER BY mtroska",
			wantArgs: []any{int64(4), 2026, 7},
		},
		{
			name: "tipovi analitike (no god/kar) work without a session",
			ctx:  context.Background(),
			run: func(ctx context.Context, s *CommonResource) error {
				_, err := s.GetTipoviAnalitikeComboValues(ctx)
				return err
			},
			wantSQL: "select tipanalitikeid, naziv from tipanalitike ORDER BY tipanalitikeid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			if err := tt.run(tt.ctx, newCommon(recordingDB(&calls))); err != nil {
				t.Fatal(err)
			}
			if len(calls) != 1 {
				t.Fatalf("got %d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			if len(calls[0].args) != len(tt.wantArgs) || (len(tt.wantArgs) > 0 && !reflect.DeepEqual(calls[0].args, tt.wantArgs)) {
				t.Errorf("args got %v want %v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

func TestSearchSQL(t *testing.T) {
	var calls []sqlCall
	s := NewSearchService(*repository.NewBaseRepository[domain.SearchRowDto](recordingDB(&calls), "partneri"))
	if _, err := s.Search(sessionCtx(), "partneri", "ab", nil); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("got %d queries, want 1", len(calls))
	}
	got := oneLine(calls[0].query)
	wantTail := "from partneri p WHERE (p.god = $1) AND (p.kar = $2) AND (p.sifra ILIKE '%' || $3 || '%' OR " +
		"p.naziv ILIKE '%' || $4 || '%' OR p.pib ILIKE '%' || $5 || '%' OR p.matbr ILIKE '%' || $6 || '%') " +
		"ORDER BY p.sifra LIMIT 50"
	if !strings.HasSuffix(got, wantTail) {
		t.Errorf("sql\n got: %s\nwant suffix: %s", got, wantTail)
	}
	if want := []any{2026, 7, "ab", "ab", "ab", "ab"}; !reflect.DeepEqual(calls[0].args, want) {
		t.Errorf("args got %v want %v", calls[0].args, want)
	}
}

// TestSearchKalkulacijaSQL checks the searches of the header of a kalkulacija veleprodaje: the polja of the EPP of the
// period and the narudžbenice of the magacin (required: without it there is no query).
func TestSearchKalkulacijaSQL(t *testing.T) {
	tests := []struct {
		name, entity, text string
		filters            map[string]string
		wantTail           string
		wantArgs           []any
		wantNoSQL          bool
	}{
		{
			name: "epp polja", entity: "epp-polja", text: "8a",
			wantTail: "from fsepp f WHERE (f.god = $1) AND (f.kar = $2) AND (f.polje ILIKE '%' || $3 || '%' OR " +
				"f.opis ILIKE '%' || $4 || '%') ORDER BY f.polje LIMIT 50",
			wantArgs: []any{2026, 7, "8a", "8a"},
		},
		{
			name: "narudžbenice of the magacin", entity: "narudzbenice", filters: map[string]string{"magaciniid": "3"},
			wantTail: "from edok e WHERE (e.god = $1) AND (e.kar = $2) AND (e.magaciniid::text = $3) " +
				"ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3"},
		},
		{name: "narudžbenice without the magacin", entity: "narudzbenice", wantNoSQL: true},
		{
			name: "zahtevi za nabavku of the vrsta and the magacin", entity: "zahtevi-nabavka", filters: map[string]string{"vrd": "410", "magaciniid": "3"},
			wantTail: "from edok e WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '10') AND (e.magaciniid::text = $3) " +
				"AND (e.vrd::text = $4) ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3", "410"},
		},
		{name: "zahtevi za nabavku without the magacin", entity: "zahtevi-nabavka", filters: map[string]string{"vrd": "410"}, wantNoSQL: true},
		{
			name: "artikli of the proizvodnja", entity: "artikli-proizvodnja", text: "vij",
			wantTail: "from rsif r WHERE (r.god = $1) AND (r.kar = $2) AND (upper(trim(coalesce(r.tip, ''))) in ('S', 'A', 'R', 'D', 'K')) " +
				"AND (r.sifra::text ILIKE '%' || $3 || '%' OR r.naziv ILIKE '%' || $4 || '%' OR r.barkod ILIKE '%' || $5 || '%') ORDER BY r.sifra LIMIT 50",
			wantArgs: []any{2026, 7, "vij", "vij", "vij"},
		},
		{
			name: "zahtevi za ponudu of the vrsta and the magacin", entity: "zahtevi-ponuda", filters: map[string]string{"vrd": "310", "magaciniid": "3"},
			wantTail: "from edok e WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '10') AND (e.magaciniid::text = $3) " +
				"AND (e.vrd::text = $4) ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3", "310"},
		},
		{name: "zahtevi za ponudu without the vrsta", entity: "zahtevi-ponuda", filters: map[string]string{"magaciniid": "3"}, wantNoSQL: true},
		{
			name: "porudžbenice of the vrsta and the magacin", entity: "porudzbenice-dobavljaca", filters: map[string]string{"vrd": "320", "magaciniid": "3"},
			wantTail: "from edok e WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '10') AND (e.magaciniid::text = $3) " +
				"AND (e.vrd::text = $4) ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3", "320"},
		},
		{name: "porudžbenice without the magacin", entity: "porudzbenice-dobavljaca", filters: map[string]string{"vrd": "320"}, wantNoSQL: true},
		{
			name: "potvrđeni zahtevi za ponudu of the magacin", entity: "ponude-potvrdjene", filters: map[string]string{"magaciniid": "3"},
			wantTail: "from edok e left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana " +
				"WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '40') AND (e.vrd in (select v.vrd from dokvrsta v " +
				"where v.god = e.god and v.kar = e.kar and upper(trim(coalesce(v.modul, ''))) = 'PRO' and upper(trim(coalesce(v.grpdok, ''))) = 'ZPN')) " +
				"AND (e.magaciniid::text = $3) ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3"},
		},
		{name: "potvrđeni zahtevi za ponudu without the magacin", entity: "ponude-potvrdjene", wantNoSQL: true},
		{
			name: "potvrđene porudžbenice of the magacin", entity: "porudzbenice-potvrdjene", filters: map[string]string{"magaciniid": "3"},
			wantTail: "from edok e left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana " +
				"WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '40') AND (e.vrd in (select v.vrd from dokvrsta v " +
				"where v.god = e.god and v.kar = e.kar and upper(trim(coalesce(v.modul, ''))) = 'PRO' and upper(trim(coalesce(v.grpdok, ''))) = 'NRD')) " +
				"AND (e.magaciniid::text = $3) ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3"},
		},
		{name: "potvrđene porudžbenice without the magacin", entity: "porudzbenice-potvrdjene", wantNoSQL: true},
		{
			name: "potvrđene profakture of the magacin", entity: "profakture-potvrdjene", text: "12", filters: map[string]string{"magaciniid": "3"},
			wantTail: "from edok e left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana " +
				"WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '20') AND (e.vrd in (select v.vrd from dokvrsta v " +
				"where v.god = e.god and v.kar = e.kar and upper(trim(coalesce(v.modul, ''))) = 'PRO' and upper(trim(coalesce(v.grpdok, ''))) = 'PRF')) " +
				"AND (e.magaciniid::text = $3) AND (e.dokum::text ILIKE '%' || $4 || '%' OR e.fana ILIKE '%' || $5 || '%' OR f.naziv ILIKE '%' || $6 || '%') " +
				"ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3", "12", "12", "12"},
		},
		{name: "potvrđene profakture without the magacin", entity: "profakture-potvrdjene", wantNoSQL: true},
		{
			name: "potvrđeni nalozi of the magacin", entity: "nalozi-potvrdjeni", filters: map[string]string{"magaciniid": "3"},
			wantTail: "from edok e left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana " +
				"WHERE (e.god = $1) AND (e.kar = $2) AND (coalesce(e.doksts, '') = '20') AND (e.vrd in (select v.vrd from dokvrsta v " +
				"where v.god = e.god and v.kar = e.kar and upper(trim(coalesce(v.modul, ''))) = 'PRO' and upper(trim(coalesce(v.grpdok, ''))) = 'NRK')) " +
				"AND (e.magaciniid::text = $3) ORDER BY e.dadok desc, e.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "3"},
		},
		{name: "potvrđeni nalozi without the magacin", entity: "nalozi-potvrdjeni", wantNoSQL: true},
		{
			name: "artikli of the zalihe of the magacin with stanje", entity: "artikli-zalihe", text: "bra",
			filters: map[string]string{"magaciniid": "3", "sastanjem": "true"},
			wantTail: "on z.rsifid = r.rsifid and z.god = r.god and z.kar = r.kar WHERE (r.god = $1) AND (r.kar = $2) AND ((z.stanje > 0)::text = $3) " +
				"AND (z.magaciniid::text = $4) AND (r.sifra::text ILIKE '%' || $5 || '%' OR r.naziv ILIKE '%' || $6 || '%' OR r.barkod ILIKE '%' || $7 || '%') " +
				"ORDER BY r.sifra LIMIT 50",
			wantArgs: []any{2026, 7, "true", "3", "bra", "bra", "bra"},
		},
		{name: "artikli of the zalihe without the magacin", entity: "artikli-zalihe", wantNoSQL: true},
		{
			name: "cenovnici", entity: "cenovnici", text: "vp",
			wantTail: "from rcenovnikhdr h left join klasacen k on k.klasacenid = h.klasacenid WHERE (h.broj::text ILIKE '%' || $1 || '%' " +
				"OR k.naziv ILIKE '%' || $2 || '%') ORDER BY h.datcen desc, h.broj desc LIMIT 50",
			wantArgs: []any{"vp", "vp"},
		},
		{
			name: "artikli sirovine", entity: "artikli-sirovine", text: "bra",
			wantTail: "from rsif r WHERE (r.god = $1) AND (r.kar = $2) AND (upper(trim(coalesce(r.tip, ''))) in ('S', 'A')) " +
				"AND (r.sifra::text ILIKE '%' || $3 || '%' OR r.naziv ILIKE '%' || $4 || '%' OR r.barkod ILIKE '%' || $5 || '%') ORDER BY r.sifra LIMIT 50",
			wantArgs: []any{2026, 7, "bra", "bra", "bra"},
		},
		{
			name: "dobavljači of the zahtev za ponudu", entity: "dobavljaci-ponude", text: "met",
			wantTail: "from fkpl f left join partneri p on p.idpartneri = f.idpartneri WHERE (f.god = $1) AND (f.kar = $2) AND (f.vkonta = 1) " +
				"AND (f.konto like '43%') AND (f.sifra ILIKE '%' || $3 || '%' OR f.naziv ILIKE '%' || $4 || '%' OR p.naziv ILIKE '%' || $5 || '%' " +
				"OR p.pib ILIKE '%' || $6 || '%') ORDER BY f.konto, f.sifra LIMIT 50",
			wantArgs: []any{2026, 7, "met", "met", "met", "met"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			s := NewSearchService(*repository.NewBaseRepository[domain.SearchRowDto](recordingDB(&calls), "rdok"))
			if _, err := s.Search(sessionCtx(), tt.entity, tt.text, tt.filters); err != nil {
				t.Fatal(err)
			}
			if tt.wantNoSQL {
				if len(calls) != 0 {
					t.Fatalf("got %d queries, want none", len(calls))
				}
				return
			}
			if len(calls) != 1 {
				t.Fatalf("got %d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); !strings.HasSuffix(got, tt.wantTail) {
				t.Errorf("sql\n got: %s\nwant suffix: %s", got, tt.wantTail)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %v want %v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

// TestSearchFaktureKnjiznogPismaSQL checks the search of the vezni dokument of a knjižno pismo: the fakture of the
// period in the magacin (required: without it there is no query), of the kupac when it is given.
func TestSearchFaktureKnjiznogPismaSQL(t *testing.T) {
	tests := []struct {
		name      string
		filters   map[string]string
		wantTail  string
		wantArgs  []any
		wantNoSQL bool
	}{
		{
			name:    "magacin and kupac",
			filters: map[string]string{"magaciniid": "3", "kupackonto": "2040", "kupacsifra": "15"},
			wantTail: "WHERE (r.god = $1) AND (r.kar = $2) AND (upper(trim(v.grpdok)) in ('FAK', 'FRP', 'FZR')) AND (r.fana = $3) " +
				"AND (r.fkto = $4) AND (r.magaciniid::text = $5) ORDER BY r.dadok desc, r.dokum desc LIMIT 50",
			wantArgs: []any{2026, 7, "15", "2040", "3"},
		},
		{name: "without the magacin", filters: map[string]string{"kupackonto": "2040"}, wantNoSQL: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			s := NewSearchService(*repository.NewBaseRepository[domain.SearchRowDto](recordingDB(&calls), "rdok"))
			if _, err := s.Search(sessionCtx(), "fakture-knjiznog-pisma", "", tt.filters); err != nil {
				t.Fatal(err)
			}
			if tt.wantNoSQL {
				if len(calls) != 0 {
					t.Fatalf("got %d queries, want none", len(calls))
				}
				return
			}
			if len(calls) != 1 {
				t.Fatalf("got %d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); !strings.HasSuffix(got, tt.wantTail) {
				t.Errorf("sql\n got: %s\nwant suffix: %s", got, tt.wantTail)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %v want %v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

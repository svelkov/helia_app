package robno

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

type sqlCall struct {
	query string
	args  []any
}

func recordingDB(calls *[]sqlCall) *db.MockDatabase {
	rec := func(q string, a []any) { *calls = append(*calls, sqlCall{q, a}) }
	return &db.MockDatabase{
		SelectContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			rec(q, a)
			return nil
		},
		GetContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			rec(q, a)
			return nil
		},
	}
}

func sessionCtx() context.Context {
	return domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7})
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// TestRobnoSQLScopes checks the conditions the migration added or kept on purpose: the tenant of the
// report and of the lookups, the keys that may not be skipped and the sort whitelist.
func TestRobnoSQLScopes(t *testing.T) {
	tests := []struct {
		name       string
		run        func(d db.Database) error
		wantSuffix string // the tail of the (single-line) query from WHERE on
		wantArgs   []any
	}{
		{
			name: "stanje subsintetickog konta: both union parts are limited to the tenant",
			run: func(d db.Database) error {
				s := &RobnoStanjaResource{
					robnostanjeRepo: *repository.NewBaseRepository[domain.RobnoStanjeDto](d, "drsta"),
					rproRepo:        *repository.NewBaseRepository[domain.Rpro](d, "rpro"),
				}
				// getTotals: only the totals are computed, so no row formatting (i18n) is needed.
				return s.GetStanjaSubsintetickogKonta(sessionCtx(), &domain.TableData{}, &domain.RobnoStanjaTotal{}, true,
					domain.RobnoStanjaParams{MagaciniID: 3, Konto: "1320"}, common.TipStampePreview)
			},
			wantSuffix: "WHERE (rpro.god = $1) AND (rpro.kar = $2) AND (rnal.tipdok = $3) AND rpro.konto ILIKE $4 AND (rnal.magaciniid = $5)) " +
				"UNION ALL (SELECT EXTRACT(MONTH FROM rnal.danal) as mesec,",
			wantArgs: []any{2026, 7, "00", "%1320%", 3, 2026, 7, 3, "1320", "00"},
		},
		{
			name: "mesto isporuke: the partner of the sifra in the tenant",
			run: func(d db.Database) error {
				s := &FispResource{partneriRepo: repository.NewBaseRepository[domain.Partneri](d, "partneri")}
				err := s.setPartnerID(sessionCtx(), &domain.Fisp{Sifra: "77"})
				if err == nil || !strings.Contains(err.Error(), "ne postoji") {
					t.Errorf("want the not-found error of the empty result, got %v", err)
				}
				return nil
			},
			wantSuffix: "WHERE (sifra = $1) AND (god = $2) AND (kar = $3) ORDER BY idpartneri LIMIT 1",
			wantArgs:   []any{"77", 2026, 7},
		},
		{
			name: "robni nalog by id from the request is limited to the tenant",
			run: func(d db.Database) error {
				s := &RobnoDokumentaResource{rnalHeaderRepo: *repository.NewBaseRepository[domain.Rnal](d, "rnal")}
				_, _, err := s.GetRobnoUnosNalog(sessionCtx(), 55)
				return err
			},
			wantSuffix: "from rnal WHERE (rnal.rnalid = $1) AND (rnal.god = $2) AND (rnal.kar = $3)",
			wantArgs:   []any{int64(55), 2026, 7},
		},
		{
			name: "kupac: empty konto and sifra stay conditions",
			run: func(d db.Database) error {
				s := &RobnoDokumentaResource{unos: RobnoDokumentaUnosRepos{Kupac: *repository.NewBaseRepository[domain.RobKupacDto](d, "fkpl")}}
				_, _, err := s.GetRobKupac(sessionCtx(), "", "")
				return err
			},
			wantSuffix: "WHERE (f.god = $1) AND (f.kar = $2) AND (f.konto = $3) AND (f.sifra = $4) LIMIT 1",
			wantArgs:   []any{2026, 7, "", ""},
		},
		{
			name: "artikli: a whitelisted sort column",
			run: func(d db.Database) error {
				s := &ArtikliResource{artikliRepo: repository.NewBaseRepository[domain.Rsif](d, "rsif")}
				return s.GetAllArtikli(sessionCtx(), &domain.TableData{}, 2, 10, false, "Naziv", "DESC", "", common.TipStampePreview)
			},
			wantSuffix: "FROM rsif WHERE (god = $1) AND (kar = $2) ORDER BY rsif.naziv DESC LIMIT 10 OFFSET 10",
			wantArgs:   []any{2026, 7},
		},
		{
			name: "artikli: a sort key outside the whitelist is ignored",
			run: func(d db.Database) error {
				s := &ArtikliResource{artikliRepo: repository.NewBaseRepository[domain.Rsif](d, "rsif")}
				return s.GetAllArtikli(sessionCtx(), &domain.TableData{}, 1, 10, false, "naziv; DROP TABLE rsif", "asc", "", common.TipStampePreview)
			},
			wantSuffix: "FROM rsif WHERE (god = $1) AND (kar = $2) LIMIT 10",
			wantArgs:   []any{2026, 7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			if err := tt.run(recordingDB(&calls)); err != nil {
				t.Fatal(err)
			}
			if len(calls) == 0 {
				t.Fatal("no query was run")
			}
			got := oneLine(calls[0].query)
			if !strings.Contains(got, tt.wantSuffix) {
				t.Errorf("sql\n got: %s\nwant to contain: %s", got, tt.wantSuffix)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %v want %v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

// TestStampaFakturaIDLists checks the helpers that read the rows of a list of documents: an empty
// list reads nothing (AddIn would drop the condition), a list is read in the tenant.
func TestStampaFakturaIDLists(t *testing.T) {
	session := &domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7}
	newResource := func(d db.Database) *RobnoDokumentaResource {
		return &RobnoDokumentaResource{
			stampaFakturaRepo:     *repository.NewBaseRepository[domain.RobnoStampaFakturaRowDto](d, "rpro"),
			stampaFakturaRateRepo: *repository.NewBaseRepository[domain.RobnoStampaFakturaRataDto](d, "faktrate"),
		}
	}
	tests := []struct {
		name     string
		run      func(s *RobnoDokumentaResource, ids []any) error
		ids      []any
		want     string // a part of the (single-line) query; "" = no query
		wantArgs []any
	}{
		{"rate: empty list", func(s *RobnoDokumentaResource, ids []any) error {
			_, err := s.stampaFakturaRate(sessionCtx(), session, ids)
			return err
		}, nil, "", nil},
		{"rate: in the tenant", func(s *RobnoDokumentaResource, ids []any) error {
			_, err := s.stampaFakturaRate(sessionCtx(), session, ids)
			return err
		}, []any{int64(3), int64(4)}, "WHERE (fr.god = $1) AND (fr.kar = $2) AND fr.rdokid IN ($3, $4)", []any{2026, 7, int64(3), int64(4)}},
		{"ambalaza: empty list", func(s *RobnoDokumentaResource, ids []any) error {
			_, err := s.stampaFakturaAmbalaza(sessionCtx(), session, ids)
			return err
		}, []any{}, "", nil},
		{"ambalaza: in the tenant", func(s *RobnoDokumentaResource, ids []any) error {
			_, err := s.stampaFakturaAmbalaza(sessionCtx(), session, ids)
			return err
		}, []any{int64(3)}, "(rpro.god = $1) AND (rpro.kar = $2) AND rpro.rdokid IN ($3)", []any{2026, 7, int64(3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			if err := tt.run(newResource(recordingDB(&calls)), tt.ids); err != nil {
				t.Fatal(err)
			}
			if tt.want == "" {
				if len(calls) != 0 {
					t.Fatalf("an empty list must not run a query, got %s", calls[0].query)
				}
				return
			}
			if len(calls) != 1 {
				t.Fatalf("%d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); !strings.Contains(got, tt.want) {
				t.Errorf("sql\n got: %s\nwant to contain: %s", got, tt.want)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %#v want %#v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

package repository

import (
	"context"
	"reflect"
	"testing"

	"helia/internal/domain"
)

// fvrListCases are the lists of the firme, their years and their kar for the choice of the period:
// no period of the session (they are read before one is chosen).
var fvrListCases = []struct {
	name     string
	respond  drvResult
	run      func(r *BaseRepository[domain.Fvr]) (any, error)
	want     any
	wantSQL  string
	wantArgs []any
}{
	{
		name:    "firme",
		respond: drvResult{cols: []string{"naziv", "knjigovod"}, rows: [][]any{{"A", "1"}, {"B", "2"}}},
		run: func(r *BaseRepository[domain.Fvr]) (any, error) {
			firme, err := FvrFirme(context.Background(), r)
			names := []string{}
			for _, f := range firme {
				names = append(names, f.Naziv)
			}
			return names, err
		},
		want:     []string{"A", "B"},
		wantSQL:  "SELECT DISTINCT ON (naziv) naziv, knjigovod FROM fvr WHERE (god > 0) ORDER BY fvr.naziv",
		wantArgs: []any{},
	},
	{
		name:    "years of a firma",
		respond: drvResult{cols: []string{"god"}, rows: [][]any{{int64(2026)}, {int64(2025)}}},
		run: func(r *BaseRepository[domain.Fvr]) (any, error) {
			return FvrGodine(context.Background(), r, "A")
		},
		want:     []int{2026, 2025},
		wantSQL:  "SELECT DISTINCT ON (god) god FROM fvr WHERE (fvr.naziv = $1) ORDER BY fvr.god desc",
		wantArgs: []any{"A"},
	},
	{
		name:    "kar of a firma in a year",
		respond: drvResult{cols: []string{"kar"}, rows: [][]any{{int64(7)}}},
		run: func(r *BaseRepository[domain.Fvr]) (any, error) {
			return FvrKarovi(context.Background(), r, "A", 2026)
		},
		want:     []int{7},
		wantSQL:  "SELECT kar FROM fvr WHERE (fvr.naziv = $1) AND (god = $2) ORDER BY fvr.kar",
		wantArgs: []any{"A", 2026},
	},
}

func TestFvrLists(t *testing.T) {
	for _, tt := range fvrListCases {
		t.Run(tt.name, func(t *testing.T) {
			d, f := newDrvDB(t, func(string, []any) drvResult { return tt.respond })
			// Without a session: the lists are read before a period is chosen.
			got, err := tt.run(NewBaseRepository[domain.Fvr](d, "fvr"))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
			if len(f.calls) != 1 {
				t.Fatalf("%d statements, want 1", len(f.calls))
			}
			if got := squash(f.calls[0].query); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			if !reflect.DeepEqual(f.calls[0].args, tt.wantArgs) {
				t.Errorf("args\n got: %#v\nwant: %#v", f.calls[0].args, tt.wantArgs)
			}
		})
	}
}

func TestGetFvrData(t *testing.T) {
	d, f := newDrvDB(t, func(string, []any) drvResult {
		return drvResult{cols: []string{"naziv", "adresa"}, rows: [][]any{{"Firma d.o.o.", "Ulica 1"}}}
	})
	ctx := domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7, Firma: "Firma d.o.o."})
	fvr, err := GetFvrData(ctx, NewBaseRepository[domain.Fvr](d, "fvr"))
	if err != nil {
		t.Fatal(err)
	}
	if fvr.Adresa != "Ulica 1" {
		t.Errorf("adresa %q", fvr.Adresa)
	}
	wantSQL := "SELECT naziv, adresa, pobro, mesto, pib, matbr, sifdel FROM fvr WHERE (fvr.naziv = $1) AND (god = $2) AND (kar = $3)"
	if got := squash(f.calls[0].query); got != wantSQL {
		t.Errorf("sql\n got: %s\nwant: %s", got, wantSQL)
	}
	if want := []any{"Firma d.o.o.", 2026, 7}; !reflect.DeepEqual(f.calls[0].args, want) {
		t.Errorf("args\n got: %#v\nwant: %#v", f.calls[0].args, want)
	}
	if _, err := GetFvrData(context.Background(), NewBaseRepository[domain.Fvr](d, "fvr")); err == nil {
		t.Error("want an error without a session")
	}
}

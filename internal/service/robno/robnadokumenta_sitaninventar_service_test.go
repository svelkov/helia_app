package robno

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
)

// TestRobSivPrikaz checks the grids of the tab "Pregled zaduž. inventara radnika": every row has a cell per column,
// the stavke come by the redni broj and the nabavni iznos is the količina by the nabavna cena.
func TestRobSivPrikaz(t *testing.T) {
	s := &RobnoDokumentaResource{}
	s.setTableFields()
	dadok := sql.NullTime{Time: time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC), Valid: true}
	dokumenta := RobSivDokumentaPrikazTableRows([]domain.RobDokumentPrikazRow{{RdokID: 7, Tipdok: "28", Nalog: 1, Vrd: 470, Dokum: 3, Dadok: dadok, Iznos: 120, Opis: "Alat"}})
	if len(dokumenta) != 1 || len(dokumenta[0].Fields) != len(s.GetSivDokumentaTableFields()) {
		t.Fatalf("documents %#v; want one row with %d cells", dokumenta, len(s.GetSivDokumentaTableFields()))
	}
	// The last one first, as the query gives them.
	stavke := RobSivPrikazStavkeTableRows([]domain.RobStavkaRow{
		{RproID: 2, Rbr: 2, Sifra: 16, Naziv: "Bušilica", Kolic: 1, Cena: 100, Iznos: 100, Ncena: 80},
		{RproID: 1, Rbr: 1, Sifra: 15, Naziv: "Čekić", Kolic: 2.5, Cena: 10, Iznos: 25, Ncena: 8.333},
	})
	tests := []struct {
		name      string
		row       int
		id        string
		nabIznos  string
		wantCells int
	}{
		{name: "the first redni broj first", row: 0, id: "1", nabIznos: robFormatIznos(20.83), wantCells: len(s.GetSivPrikazStavkeTableFields())},
		{name: "then the second", row: 1, id: "2", nabIznos: robFormatIznos(80), wantCells: len(s.GetSivPrikazStavkeTableFields())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := stavke[tt.row]
			if r.ID != tt.id || len(r.Fields) != tt.wantCells {
				t.Fatalf("row %#v; want the id %s with %d cells", r, tt.id, tt.wantCells)
			}
			if got := r.Fields[len(r.Fields)-1]; got != tt.nabIznos {
				t.Errorf("nabavni iznos %q, want %q", got, tt.nabIznos)
			}
		})
	}
	if got := stavke[0].Fields[5]; got != common.FormatNumberWithSystemLocale(2.5, 3) {
		t.Errorf("količina %q", got)
	}
}

// TestRobSivValidateBezSesije checks that the header of a zaduženje sitnog inventara is not validated without a
// session (no repository is reached).
func TestRobSivValidateBezSesije(t *testing.T) {
	got := (&RobnoDokumentaResource{}).robSivValidate(context.Background(), domain.RobnoOpstiParams{})
	if len(got) != 1 || got[0].Field != "brdok" {
		t.Errorf("errors = %#v, want one error of brdok", got)
	}
}

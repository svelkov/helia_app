package robno

import (
	"reflect"
	"strings"
	"testing"

	"helia/internal/domain"
	"helia/internal/repository"
)

// TestRobKnpObracunRedovi checks the rows of the window of the obračun of a knjižno pismo (the legacy
// rob_KnjPismoPR_Obracun): the values by the new conditions, by the ones of the faktura and their difference,
// with the stope of both.
func TestRobKnpObracunRedovi(t *testing.T) {
	faktura := []domain.RobStavkaRow{
		{Kolic: 10, Fcena: 100, Rab: 10, Stopa: 20}, // bruto 1000, rabat 100
		{Kolic: 1, Fcena: 50, Rab: 0, Stopa: 10},    // bruto 50
	}
	novo := []domain.RobStavkaRow{
		{Kolic: 8, Fcena: 100, Rab: 10, Stopa: 20}, // bruto 800, rabat 80
	}
	got := robKnpObracunRedovi(robKnpObracunaj(novo, 0, 0), robKnpObracunaj(faktura, 5, 0))

	type red struct {
		opisNovo, opisFaktura, opisRazlika string
		novo, faktura, razlika             float64
	}
	want := []red{
		{"BRUTO", "BRUTO", "BRUTO", 800, 1050, -250},
		{"RABAT", "RABAT", "RABAT", 80, 100, -20},
		{"NETO", "NETO", "NETO", 720, 950, -230},
		// the ugovoreni rabat of the faktura: 5% of 900 and of 50
		{"UGOVORENI RABAT 0%", "UGOVORENI RABAT 5%", "UGOVORENI RABAT", 0, 47.5, -47.5},
		{"KASA 0%", "KASA 0%", "KASA", 0, 0, 0},
		// the stopa 10 only of the faktura: 50 less 5% = 47.5, PDV 4.75
		{"Osnovica 10%", "Osnovica 10%", "Osnovica 10%", 0, 47.5, -47.5},
		{"PDV 10%", "PDV 10%", "PDV 10%", 0, 4.75, -4.75},
		// the stopa 20: 720 and 900 less 5% = 855
		{"Osnovica 20%", "Osnovica 20%", "Osnovica 20%", 720, 855, -135},
		{"PDV 20%", "PDV 20%", "PDV 20%", 144, 171, -27},
		// 720 + 144 and 1050 - 100 - 47.5 + 4.75 + 171
		{"IZNOS FAKTURE", "IZNOS FAKTURE", "IZNOS FAKTURE", 864, 1078.25, -214.25},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d: %#v", len(got), len(want), got)
	}
	for i, w := range want {
		t.Run(w.opisRazlika, func(t *testing.T) {
			g := red{got[i].OpisNovo, got[i].OpisFaktura, got[i].OpisRazlika, got[i].Novo, got[i].Faktura, got[i].Razlika}
			if !reflect.DeepEqual(g, w) {
				t.Errorf("row %d = %+v, want %+v", i, g, w)
			}
		})
	}
}

// TestRobKnpStavkeTableRows checks the order of the rows of the stavke of a knjižno pismo: by the redni broj (the
// stavke of the faktura) or the last one first (the corrected stavke, as they come from the query).
func TestRobKnpStavkeTableRows(t *testing.T) {
	stavke := []domain.RobStavkaRow{{RproID: 2, Rbr: 2}, {RproID: 1, Rbr: 1}} // the last one first
	tests := []struct {
		name  string
		poRbr bool
		want  []string
	}{
		{name: "by the redni broj", poRbr: true, want: []string{"1", "2"}},
		{name: "the last one first", poRbr: false, want: []string{"2", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := RobKnpStavkeTableRows(stavke, tt.poRbr)
			var ids []string
			for _, r := range rows {
				ids = append(ids, r.ID)
				if len(r.Fields) != 12 {
					t.Errorf("row %s has %d cells, want 12", r.ID, len(r.Fields))
				}
			}
			if !reflect.DeepEqual(ids, tt.want) {
				t.Errorf("ids = %v, want %v", ids, tt.want)
			}
		})
	}
}

// TestRobKnpDokumentaPrikaz checks the grid of the documents of the "Prikaz knjižnih pisama" tab: the tip of the
// knjižno pismo after the naziv partnera, in the cells and in the columns.
func TestRobKnpDokumentaPrikaz(t *testing.T) {
	s := &RobnoDokumentaResource{}
	s.setTableFields()
	fields := s.GetKnpDokumentaTableFields()
	if len(fields) != len(s.GetFaktureDokumentaTableFields())+1 {
		t.Fatalf("got %d columns, want the ones of the fakture and one more", len(fields))
	}
	if fields[robKnpTipKolona-1].Name != "naziv" || fields[robKnpTipKolona].Name != "tipknjpisma" {
		t.Errorf("columns %q, %q; want naziv, tipknjpisma", fields[robKnpTipKolona-1].Name, fields[robKnpTipKolona].Name)
	}
	rows := RobKnpDokumentaPrikazTableRows([]domain.RobDokumentPrikazRow{{RdokID: 7, Naziv: "Kupac", Opis: "Opis", TipKnjPisma: "2 - Povraćaj"}})
	if len(rows) != 1 || len(rows[0].Fields) != len(fields) {
		t.Fatalf("rows %#v; want one row with %d cells", rows, len(fields))
	}
	if got := rows[0].Fields[robKnpTipKolona-1 : robKnpTipKolona+2]; !reflect.DeepEqual(got, []string{"Kupac", "2 - Povraćaj", "Opis"}) {
		t.Errorf("cells %v; want the naziv, the tip and the opis", got)
	}
}

// TestRobDokumentaPrikazSQL checks that the query of the grid of the documents reads the tip of the knjižno pismo
// by its key and keeps the period of the session and the magacin.
func TestRobDokumentaPrikazSQL(t *testing.T) {
	var calls []sqlCall
	d := recordingDB(&calls)
	s := &RobnoDokumentaResource{unos: RobnoDokumentaUnosRepos{Prikaz: *repository.NewBaseRepository[domain.RobDokumentPrikazRow](d, "rdok")}}
	if _, _, err := s.GetRobDokumentaPrikaz(sessionCtx(), 3, []int{410}, domain.RobDokumentaPrikazParams{}); err != nil {
		t.Fatal(err)
	}
	if len(calls) == 0 {
		t.Fatal("no query")
	}
	got := oneLine(calls[len(calls)-1].query)
	for _, want := range []string{
		"coalesce(concat_ws(' - ', tkp.sifrazlog::text, nullif(tkp.opis, '')), '') as tipknjpisma",
		"left join tipknpisma tkp on tkp.tipknjid = r.tipknjid",
		"WHERE (r.god = $1) AND (r.kar = $2) AND r.vrd IN ($3) AND (r.magaciniid = $4)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("sql\n got: %s\nwant to contain: %s", got, want)
		}
	}
}

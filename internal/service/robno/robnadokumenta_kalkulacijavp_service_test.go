package robno

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
)

func TestRobKalOznaka(t *testing.T) {
	tests := []struct{ in, want string }{
		{"%", "%"}, {" d ", "d"}, {"-", "-"}, {"", "-"}, {"x", "-"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := robKalOznaka(tt.in); got != tt.want {
				t.Errorf("robKalOznaka(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRobKalNabavnaCena(t *testing.T) {
	tests := []struct {
		name string
		o    RobKalObracun
		want float64
	}{
		{name: "fakturna cena only", o: RobKalObracun{Fcena: 100, Tz: "-", Tzi: "-"}, want: 100},
		{name: "rabat", o: RobKalObracun{Fcena: 100, Rab: 10, Tz: "-", Tzi: "-"}, want: 90},
		{name: "zavisni trošak in %", o: RobKalObracun{Fcena: 100, Rab: 10, Tz: "%", Ztro: 5, Tzi: "-"}, want: 94.5},
		{name: "zavisni trošak as vrednost", o: RobKalObracun{Fcena: 100, Tz: "d", Ztro: 3.25, Tzi: "-"}, want: 103.25},
		// neto 90, ztr 4.5, external 2% of 94.5 = 1.89, interni 1 = 97.39
		{name: "external and interni troškovi", o: RobKalObracun{Fcena: 100, Rab: 10, Tz: "%", Ztro: 5, Tzi: "d", Ztrin: 1, ProcenatTroskova: 2},
			want: 97.39},
		{name: "interni trošak in %", o: RobKalObracun{Fcena: 200, Tz: "-", Tzi: "%", Ztrin: 2.5}, want: 205},
		{name: "oznaka without a trošak", o: RobKalObracun{Fcena: 100, Tz: "-", Ztro: 7, Tzi: "-"}, want: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RobKalNabavnaCena(tt.o); got != tt.want {
				t.Errorf("RobKalNabavnaCena() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRobKalCeneMarze(t *testing.T) {
	if got := RobKalVpCena(97.39, 20); got != 116.87 {
		t.Errorf("RobKalVpCena() = %v, want 116.87", got)
	}
	if got := RobKalMcena(100, 25); got != 125 {
		t.Errorf("RobKalMcena() = %v, want 125", got)
	}
	// 100 with 25% = 125, porez 20% = 25 -> 150
	if got := RobKalMpCena(100, 25, 20); got != 150 {
		t.Errorf("RobKalMpCena() = %v, want 150", got)
	}
}

func TestRobKalProcenatTroskova(t *testing.T) {
	if got := robKalProcenatTroskova(150, 1000); got != 15 {
		t.Errorf("robKalProcenatTroskova() = %v, want 15", got)
	}
	if got := robKalProcenatTroskova(150, 0); got != 0 {
		t.Errorf("robKalProcenatTroskova() without the fakturna vrednost = %v, want 0", got)
	}
}

func TestRobKalCene(t *testing.T) {
	domaci := domain.RobStavkeKontekst{Sifval: 941, DomacaValuta: 941, Kurs: 1, Vaziza: 1}
	devizni := domain.RobStavkeKontekst{Sifval: 978, DomacaValuta: 941, Kurs: 117.2, Vaziza: 1}
	tests := []struct {
		name                           string
		k                              domain.RobStavkeKontekst
		params                         domain.RobKalStavkaParams
		wantFcena, wantNcena, wantCena float64
		wantMcenap                     float64
		wantTz, wantTzi                string
	}{
		{name: "the cene computed from the marže", k: domaci,
			params:    domain.RobKalStavkaParams{Kolic: "2", Fcena: "100", Rabat: "10", Tz: "%", Ztro: "5", Tzi: "-", Vma: "20", Mma: "25"},
			wantFcena: 100, wantNcena: 94.5, wantCena: 113.4, wantMcenap: RobKalMpCena(94.5, 25, 20), wantTz: "%", wantTzi: "-"},
		{name: "the VP and the MP cena as entered", k: domaci,
			params:    domain.RobKalStavkaParams{Kolic: "1", Fcena: "100", Tz: "-", Tzi: "-", Vma: "20", Cena: "125", Mma: "25", Mcenap: "160"},
			wantFcena: 100, wantNcena: 100, wantCena: 125, wantMcenap: 160, wantTz: "-", wantTzi: "-"},
		{name: "a trošak 0 has no oznaka", k: domaci,
			params:    domain.RobKalStavkaParams{Kolic: "1", Fcena: "100", Tz: "%", Ztro: "0", Tzi: "d", Ztrin: "0", Cena: "100", Mcenap: "100"},
			wantFcena: 100, wantNcena: 100, wantCena: 100, wantMcenap: 100, wantTz: "-", wantTzi: "-"},
		{name: "devizni: the fakturna cena converted by the kurs", k: devizni,
			params:    domain.RobKalStavkaParams{Kolic: "1", Fcena: "10", Tz: "-", Tzi: "-", Cena: "1500", Mcenap: "1800"},
			wantFcena: 1172, wantNcena: 1172, wantCena: 1500, wantMcenap: 1800, wantTz: "-", wantTzi: "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := robKalCene(tt.k, tt.params, 0, 20)
			if c.fcena != tt.wantFcena || c.ncena != tt.wantNcena || c.cena != tt.wantCena || c.mcenap != tt.wantMcenap ||
				c.tz != tt.wantTz || c.tzi != tt.wantTzi {
				t.Errorf("robKalCene() = %+v, want fcena %v ncena %v cena %v mcenap %v tz %q tzi %q", c, tt.wantFcena, tt.wantNcena,
					tt.wantCena, tt.wantMcenap, tt.wantTz, tt.wantTzi)
			}
		})
	}
}

func TestRobKalRedovi(t *testing.T) {
	got := robKalRedovi([]string{" 100 ", "50"}, []string{"20"}, []string{"2", "1"}, nil, []string{"", "2026-03-01"})
	want := []robKalPorezRed{
		{osn: "100", ppor: "20", po: "2"},
		{osn: "50", po: "1", datum: "2026-03-01"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("robKalRedovi() = %#v, want %#v", got, want)
	}
	if got := robKalRedovi(nil, nil, nil, nil, nil); len(got) != 0 {
		t.Errorf("robKalRedovi() of no rows = %#v", got)
	}
}

func TestRobKalTableRows(t *testing.T) {
	kol := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 3) }
	dva := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 2) }
	stavke := []domain.RobStavkaRow{
		{RproID: 12, Rbr: 2, Konto: "1320", Sifra: 16, Naziv: "Matica", Jm: "kom", Kolic: 1, Fcena: 10, Ncena: 10, Vma: 20, Cena: 12, Mma: 25,
			Mcenap: 15},
		{RproID: 11, Rbr: 1, Konto: "1320", Sifra: 15, Naziv: "Vijak", Jm: "kom", Kolic: 2, Fcena: 100, Rab: 10, Otk: "O1", Serija: "S1",
			Ncena: 94.5, Vma: 20, Cena: 113.4, Mma: 25, Mcenap: 141.75},
	}
	want := []domain.TableRow{
		{ID: "11", Fields: []string{"1", "1320", "15", "Vijak", "kom", kol(2), robFormatIznos(100), dva(10), "O1", "S1", "", robFormatIznos(94.5),
			dva(20), robFormatIznos(113.4), dva(25), robFormatIznos(141.75)}},
		{ID: "12", Fields: []string{"2", "1320", "16", "Matica", "kom", kol(1), robFormatIznos(10), dva(0), "", "", "", robFormatIznos(10),
			dva(20), robFormatIznos(12), dva(25), robFormatIznos(15)}},
	}
	if got := RobKalStavkeTableRows(stavke); !reflect.DeepEqual(got, want) {
		t.Errorf("RobKalStavkeTableRows() = %#v, want %#v", got, want)
	}

	datum := sql.NullTime{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	troskovi := []domain.RobKalTrosakRow{{RztrID: 8, Fkto: "4350", Fana: "16", Naziv: "Prevoznik", Dokdob: "T-5", Dadok: datum, Rok: 15,
		Iznfak: 1000, Iznos: 50}}
	wantTroskovi := []domain.TableRow{{ID: "8", Fields: []string{"4350", "16", "Prevoznik", "T-5", robFormatDatum(datum), "15",
		robFormatIznos(1000), robFormatIznos(50)}}}
	if got := RobKalTroskoviTableRows(troskovi); !reflect.DeepEqual(got, wantTroskovi) {
		t.Errorf("RobKalTroskoviTableRows() = %#v, want %#v", got, wantTroskovi)
	}
}

// TestRobKalValidacijaBezSesije checks that the validations of the kalkulacija stop with an error of the first field when
// the request has no session (no repository is reached).
func TestRobKalValidacijaBezSesije(t *testing.T) {
	s := &RobnoDokumentaResource{}
	tests := []struct {
		name  string
		run   func() []domain.FieldError
		field string
	}{
		{name: "header", run: func() []domain.FieldError {
			errs, _ := s.ValidateRobKal(context.Background(), domain.RobnoKalParams{})
			return errs
		}, field: "brdok"},
		{name: "stavka", run: func() []domain.FieldError {
			errs, _ := s.ValidateRobKalStavka(context.Background(), domain.RobKalStavkaParams{})
			return errs
		}, field: "sifraartikla"},
		{name: "zavisni trošak", run: func() []domain.FieldError {
			return s.ValidateRobKalTrosak(context.Background(), domain.RobKalTrosakParams{})
		}, field: "trkonto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.run()
			if len(got) != 1 || got[0].Field != tt.field {
				t.Errorf("errors = %#v, want one error of %q", got, tt.field)
			}
		})
	}
}

package robno

import (
	"context"
	"reflect"
	"testing"

	"helia/internal/common"
	"helia/internal/domain"
)

func TestRobKmpNabavnaCena(t *testing.T) {
	tests := []struct {
		name             string
		fcena, rab, ztro float64
		tz               string
		want             float64
	}{
		{name: "fakturna cena less the rabat", fcena: 100, rab: 10, tz: "-", want: 90},
		{name: "zavisni trošak in %", fcena: 100, rab: 10, tz: "%", ztro: 5, want: 94.5},
		{name: "zavisni trošak as vrednost", fcena: 100, tz: "d", ztro: 2.5, want: 102.5},
		{name: "oznaka without a trošak", fcena: 100, tz: "-", ztro: 7, want: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RobKmpNabavnaCena(tt.fcena, tt.rab, tt.tz, tt.ztro); got != tt.want {
				t.Errorf("RobKmpNabavnaCena() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRobKmpCene(t *testing.T) {
	tests := []struct {
		name                         string
		params                       domain.RobKmpStavkaParams
		wantNcena, wantMcena, wantMp float64
		wantTz                       string
	}{
		{name: "the MP cena from the marža and the stopa",
			params:    domain.RobKmpStavkaParams{Kolic: "2", Fcena: "100", Tz: "-", Mma: "25"},
			wantNcena: 100, wantMcena: 125, wantMp: 150, wantTz: "-"},
		{name: "the MP cena as entered",
			params:    domain.RobKmpStavkaParams{Kolic: "1", Fcena: "100", Rabat: "10", Tz: "%", Ztro: "5", Mma: "20", Mcenap: "140"},
			wantNcena: 94.5, wantMcena: 113.4, wantMp: 140, wantTz: "%"},
		{name: "a trošak 0 has no oznaka",
			params:    domain.RobKmpStavkaParams{Kolic: "1", Fcena: "100", Tz: "d", Ztro: "0", Mcenap: "120"},
			wantNcena: 100, wantMcena: 100, wantMp: 120, wantTz: "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := robKmpCene(tt.params, 20)
			if c.ncena != tt.wantNcena || c.mcena != tt.wantMcena || c.mcenap != tt.wantMp || c.tz != tt.wantTz {
				t.Errorf("robKmpCene() = %+v, want ncena %v mcena %v mcenap %v tz %q", c, tt.wantNcena, tt.wantMcena, tt.wantMp, tt.wantTz)
			}
		})
	}
}

func TestRobKmpMaloprodaja(t *testing.T) {
	for _, tt := range []struct {
		tipmag string
		want   bool
	}{{"M", true}, {" m ", true}, {"V", false}, {"", false}} {
		if got := robKmpMaloprodaja(tt.tipmag); got != tt.want {
			t.Errorf("robKmpMaloprodaja(%q) = %v, want %v", tt.tipmag, got, tt.want)
		}
	}
}

func TestRobKmpStavkeTableRows(t *testing.T) {
	kol := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 3) }
	dva := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 2) }
	stavke := []domain.RobStavkaRow{
		{RproID: 12, Rbr: 2, Konto: "1340", Sifra: 0, Naziv: "Usluga", Jm: "kom", Kolic: 1, Fcena: 50, Ncena: 50, Tz: "-", Mma: 20, Mcenap: 72},
		{RproID: 11, Rbr: 1, Konto: "1340", Sifra: 15, Naziv: "Vijak", Jm: "kom", Kolic: 2, Fcena: 100, Rab: 10, Ncena: 94.5, Tz: "%", Ztro: 5,
			Mma: 25, Mcenap: 141.75},
	}
	want := []domain.TableRow{
		{ID: "11", Fields: []string{"1", "1340", "15", "Vijak", "kom", kol(2), robFormatIznos(100), dva(10), robFormatIznos(94.5), dva(5) + " %",
			dva(25), robFormatIznos(141.75)}},
		{ID: "12", Fields: []string{"2", "1340", "0", "Usluga", "kom", kol(1), robFormatIznos(50), dva(0), robFormatIznos(50), "", dva(20),
			robFormatIznos(72)}},
	}
	if got := RobKmpStavkeTableRows(stavke); !reflect.DeepEqual(got, want) {
		t.Errorf("RobKmpStavkeTableRows() = %#v, want %#v", got, want)
	}
}

// TestRobKmpValidacijaBezSesije checks that the validations of the kalkulacija maloprodaje stop with an error of the
// first field when the request has no session (no repository is reached).
func TestRobKmpValidacijaBezSesije(t *testing.T) {
	s := &RobnoDokumentaResource{}
	if errs, _ := s.ValidateRobKmp(context.Background(), domain.RobnoKmpParams{}); len(errs) != 1 || errs[0].Field != "brdok" {
		t.Errorf("header errors = %#v, want one error of brdok", errs)
	}
	if errs := s.ValidateRobKmpStavka(context.Background(), domain.RobKmpStavkaParams{}); len(errs) != 1 || errs[0].Field != "sifraartikla" {
		t.Errorf("stavka errors = %#v, want one error of sifraartikla", errs)
	}
}

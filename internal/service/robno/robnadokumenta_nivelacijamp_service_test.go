package robno

import (
	"context"
	"reflect"
	"testing"

	"helia/internal/common"
	"helia/internal/domain"
)

func TestRobMniIznos(t *testing.T) {
	tests := []struct {
		name                string
		kolicina, sta, nova float64
		want                float64
	}{
		{name: "higher nova cena", kolicina: 2, sta: 100, nova: 110, want: 20},
		{name: "lower nova cena", kolicina: 3, sta: 100, nova: 90, want: -30},
		{name: "rounded to the para", kolicina: 1.5, sta: 10.333, nova: 10.999, want: 1},
		{name: "same cena", kolicina: 5, sta: 50, nova: 50, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := robMniIznos(tt.kolicina, tt.sta, tt.nova); got != tt.want {
				t.Errorf("robMniIznos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRobMniNovaTarifa(t *testing.T) {
	tests := []struct {
		name   string
		params domain.RobMniStavkaParams
		want   string
	}{
		{name: "without Promeni tarifu: the stara tarifa", params: domain.RobMniStavkaParams{StaraTarifa: " 1 ", NovaTarifa: "2"}, want: "1"},
		{name: "with Promeni tarifu: the nova tarifa", params: domain.RobMniStavkaParams{StaraTarifa: "1", NovaTarifa: " 2", PromeniTarifu: true}, want: "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := robMniNovaTarifa(tt.params); got != tt.want {
				t.Errorf("robMniNovaTarifa() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRobMniProveriStavku(t *testing.T) {
	ok := domain.RobMniStavkaParams{Kolic: "2", StaraCena: "100", NovaCena: "110", StaraTarifa: "1", NovaTarifa: "0"}
	with := func(change func(p *domain.RobMniStavkaParams)) domain.RobMniStavkaParams {
		p := ok
		change(&p)
		return p
	}
	tests := []struct {
		name   string
		params domain.RobMniStavkaParams
		want   []string
	}{
		{name: "valid, the nova tarifa is the stara one", params: ok, want: nil},
		{name: "negative količina is allowed", params: with(func(p *domain.RobMniStavkaParams) { p.Kolic = "-1" }), want: nil},
		{name: "no količina", params: with(func(p *domain.RobMniStavkaParams) { p.Kolic = "0" }), want: []string{"kolicina"}},
		{name: "no cene", params: with(func(p *domain.RobMniStavkaParams) { p.StaraCena, p.NovaCena = "", "-5" }), want: []string{"staracena", "novacena"}},
		{name: "no stara tarifa", params: with(func(p *domain.RobMniStavkaParams) { p.StaraTarifa = "" }), want: []string{"staratarifa"}},
		{name: "Promeni tarifu without the nova tarifa", params: with(func(p *domain.RobMniStavkaParams) { p.PromeniTarifu = true }), want: []string{"novatarifa"}},
		{name: "Promeni tarifu with the nova tarifa", params: with(func(p *domain.RobMniStavkaParams) { p.PromeniTarifu, p.NovaTarifa = true, "2" }), want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			robMniProveriStavku(tt.params, func(field, _ string) { got = append(got, field) })
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRobMniStavkeTableRows(t *testing.T) {
	kol := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 3) }
	stavke := []domain.RobStavkaRow{
		{RproID: 12, Rbr: 2, Naziv: "Hleb", Jm: "kom", Kolic: 1, Cena: 50, Mcenap: 45, Po: 2, Dani: 2},
		{RproID: 11, Rbr: 1, Naziv: "Mleko", Jm: "l", Kolic: 2, Cena: 100, Mcenap: 110, Po: 1, Dani: 3},
	}
	want := []domain.TableRow{
		{ID: "11", Fields: []string{"1", "Mleko", "l", kol(2), robFormatIznos(100), robFormatIznos(110), robFormatIznos(10), "1", "3"}},
		{ID: "12", Fields: []string{"2", "Hleb", "kom", kol(1), robFormatIznos(50), robFormatIznos(45), robFormatIznos(-5), "2", "2"}},
	}
	if got := RobMniStavkeTableRows(stavke); !reflect.DeepEqual(got, want) {
		t.Errorf("RobMniStavkeTableRows() = %#v, want %#v", got, want)
	}
	if got := RobMniStavkeTableRows(nil); len(got) != 0 {
		t.Errorf("RobMniStavkeTableRows(nil) = %#v, want no rows", got)
	}
}

// TestRobMniValidacijaBezSesije checks that the validations of the nivelacija maloprodaje stop with an error of
// the first field when the request has no session (no repository is reached).
func TestRobMniValidacijaBezSesije(t *testing.T) {
	s := &RobnoDokumentaResource{}
	tests := []struct {
		name  string
		run   func() []domain.FieldError
		field string
	}{
		{name: "header", run: func() []domain.FieldError {
			return s.ValidateRobMni(context.Background(), domain.RobnoMniParams{})
		}, field: "brdok"},
		{name: "stavka", run: func() []domain.FieldError {
			return s.ValidateRobMniStavka(context.Background(), domain.RobMniStavkaParams{})
		}, field: "nazivmni"},
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

package robno

import (
	"context"
	"reflect"
	"testing"

	"helia/internal/common"
	"helia/internal/domain"
)

func TestRobProStavkeTableRows(t *testing.T) {
	kol := func(v float64) string { return common.FormatNumberWithSystemLocale(v, 3) }
	tests := []struct {
		name   string
		stavke []domain.RobStavkaRow
		want   []domain.TableRow
	}{
		{name: "no stavke", stavke: nil, want: []domain.TableRow{}},
		{
			name: "by the redni broj, with the cena, the rabat and the iznos",
			stavke: []domain.RobStavkaRow{
				{RproID: 12, Rbr: 2, Konto: "1320", Sifra: 0, Naziv: "Usluga montaže", Jm: "kom", Kolic: 1, Cena: 50, Rab: 0, Iznos: 50},
				{RproID: 11, Rbr: 1, Konto: "1320", Sifra: 15, Naziv: "Vijak", Jm: "kom", Kolic: 2, Cena: 100, Rab: 10, Iznos: 200},
			},
			// The numbers are in the format of the system locale.
			want: []domain.TableRow{
				{ID: "11", Fields: []string{"1", "1320", "15", "Vijak", "kom", kol(2), robFormatIznos(100), robFormatIznos(10), robFormatIznos(200)}},
				{ID: "12", Fields: []string{"2", "1320", "0", "Usluga montaže", "kom", kol(1), robFormatIznos(50), robFormatIznos(0), robFormatIznos(50)}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RobProStavkeTableRows(tt.stavke)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RobProStavkeTableRows() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

// TestRobProValidacijaBezSesije checks that the validations of the profaktura stop with an error of the first
// field when the request has no session (no repository is reached).
func TestRobProValidacijaBezSesije(t *testing.T) {
	s := &RobnoDokumentaResource{}
	tests := []struct {
		name  string
		run   func() []domain.FieldError
		field string
	}{
		{name: "header", run: func() []domain.FieldError {
			return s.ValidateRobPro(context.Background(), domain.RobnoProParams{})
		}, field: "brdok"},
		{name: "stavka", run: func() []domain.FieldError {
			return s.ValidateRobProStavka(context.Background(), domain.RobProStavkaParams{})
		}, field: "sifraartikla"},
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

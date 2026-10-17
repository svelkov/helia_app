package robno

import (
	"reflect"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/repository"
)

// TestRobZnNeraspolozivo checks the change of the neraspoložive zalihe of the legacy AzurRSTA_ZN and AzurDRSTA_ZN by
// the knjiženje and the kvalitet (zaliha "Z", else "N").
func TestRobZnNeraspolozivo(t *testing.T) {
	ulaz, izlaz, ostalo := robKnjizenje{kodknj: "D"}, robKnjizenje{kodknj: "P"}, robKnjizenje{kodknj: "X"}
	tests := []struct {
		name       string
		k          robKnjizenje
		zaliha     bool
		rsta, drst float64
	}{
		{name: "ulaz of the zalihe lowers them", k: ulaz, zaliha: true, rsta: -5, drst: -5},
		{name: "ulaz of the neraspoložive zalihe raises them", k: ulaz, rsta: 5, drst: 5},
		{name: "izlaz of the zalihe: rsta not, drsta raised", k: izlaz, zaliha: true, rsta: 0, drst: 5},
		{name: "izlaz of the neraspoložive zalihe: rsta not, drsta lowered", k: izlaz, rsta: 0, drst: -5},
		{name: "another knjiženje changes nothing", k: ostalo, zaliha: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := robZnRstaNeraspolozivo(tt.k, 5, tt.zaliha); got != tt.rsta {
				t.Errorf("rsta = %v, want %v", got, tt.rsta)
			}
			if got := robZnDrstaNeraspolozivo(tt.k, 5, tt.zaliha); got != tt.drst {
				t.Errorf("drsta = %v, want %v", got, tt.drst)
			}
		})
	}
}

// TestRobZnPredznak checks the sign of the količina and the iznos by the predznak of the vrsta.
func TestRobZnPredznak(t *testing.T) {
	if got := robZnPredznak(-3, robKnjizenje{predznak: "+"}); got != 3 {
		t.Errorf("+ = %v, want 3", got)
	}
	if got := robZnPredznak(3, robKnjizenje{predznak: "-"}); got != -3 {
		t.Errorf("- = %v, want -3", got)
	}
}

// TestRobZnKolone checks the new rows of the stanje (rsta) and of the lot (drsta) of a stavka of a prijemnica.
func TestRobZnKolone(t *testing.T) {
	p := domain.RobPromenaStanja{MagaciniID: 3, Konto: "1010", Sifra: 15, Kolicina: 4, Iznos: 40, Cena: 10,
		Danal: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Otk: "002", Serija: "S1", Roktr: 0}
	ulaz := robKnjizenje{kodknj: "D", predznak: "+"}
	storno := robKnjizenje{kodknj: "D", predznak: "-"}
	tests := []struct {
		name   string
		k      robKnjizenje
		zaliha bool
		p      domain.RobPromenaStanja
		drsta  bool
		want   map[string]any
	}{
		{name: "rsta of the zalihe: the ulaz, the neraspoložive at 0", k: ulaz, zaliha: true, p: p,
			want: map[string]any{"ulaz": 4.0, "mul_3": 4.0, "dug": 40.0, "mdug_3": 40.0, "kolicner": 0.0, "brst": 1, "cena": 10.0, "zdat": "2026-03-09"}},
		{name: "rsta of the neraspoložive zalihe: the ulaz too", k: ulaz, p: p,
			want: map[string]any{"ulaz": 4.0, "mul_3": 4.0, "kolicner": 4.0}},
		{name: "rsta of a storno: the negative ulaz raises the neraspoložive", k: storno, zaliha: true,
			p:    domain.RobPromenaStanja{Konto: "1010", Sifra: 15, Kolicina: -4, Iznos: -40, Danal: p.Danal},
			want: map[string]any{"ulaz": -4.0, "kolicner": 4.0}},
		{name: "drsta of the zalihe: the ulaz and the lot, the neraspoložive at 0", k: ulaz, zaliha: true, p: p, drsta: true,
			want: map[string]any{"ulaz": 4.0, "mul_3": 4.0, "kolicner": 0.0, "otk": "002", "serija": "S1", "zstatus": "", "cena": 10.0}},
		{name: "drsta of the neraspoložive zalihe: no ulaz", k: ulaz, p: p, drsta: true,
			want: map[string]any{"ulaz": 0, "dug": 0, "kolicner": 4.0}},
		{name: "drsta of a storno: the negative neraspoložive stay", k: storno, p: domain.RobPromenaStanja{Kolicina: -4, Danal: p.Danal}, drsta: true,
			want: map[string]any{"kolicner": -4.0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			if tt.drsta {
				got = robZnDrstaKolone(tt.p, tt.k, 3, 77, 1, tt.zaliha)
			} else {
				got = robZnRstaKolone(tt.p, tt.k, 3, 77, 1, tt.zaliha)
			}
			for k, v := range tt.want {
				if !reflect.DeepEqual(got[k], v) {
					t.Errorf("%s = %#v, want %#v", k, got[k], v)
				}
			}
			if _, ok := got["mul_3"]; ok && tt.drsta && !tt.zaliha {
				t.Errorf("the lot of the neraspoložive zalihe has the ulaz of the month")
			}
		})
	}
}

// TestRobZnIzmene checks the changes of an existing stanje (rsta) and lot (drsta).
func TestRobZnIzmene(t *testing.T) {
	p := domain.RobPromenaStanja{Kolicina: 4, Iznos: 40, Danal: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)}
	ulaz := robKnjizenje{kodknj: "D", predznak: "+"}
	kretanje := []repository.Izmena{
		{Kolona: "dug", Dodaj: true, Vrednost: 40.0}, {Kolona: "mdug_3", Dodaj: true, Vrednost: 40.0},
		{Kolona: "ulaz", Dodaj: true, Vrednost: 4.0}, {Kolona: "mul_3", Dodaj: true, Vrednost: 4.0},
	}
	tests := []struct {
		name string
		got  []repository.Izmena
		want []repository.Izmena
	}{
		{name: "rsta of the zalihe", got: robZnRstaIzmene(p, ulaz, 3, true),
			want: append(append([]repository.Izmena{}, kretanje...), repository.Izmena{Kolona: "kolicner", Dodaj: true, Vrednost: -4.0})},
		{name: "rsta of the neraspoložive zalihe", got: robZnRstaIzmene(p, ulaz, 3, false),
			want: append(append([]repository.Izmena{}, kretanje...), repository.Izmena{Kolona: "kolicner", Dodaj: true, Vrednost: 4.0})},
		{name: "drsta of the zalihe, not below 0", got: robZnDrstaIzmene(p, ulaz, 3, true),
			want: append(append([]repository.Izmena{}, kretanje...), repository.Izmena{Kolona: "kolicner", Dodaj: true, NeNegativno: true, Vrednost: -4.0})},
		{name: "drsta of the neraspoložive zalihe: no kretanje", got: robZnDrstaIzmene(p, ulaz, 3, false),
			want: []repository.Izmena{{Kolona: "kolicner", Dodaj: true, NeNegativno: true, Vrednost: 4.0}}},
		{name: "drsta of a storno may go below 0", got: robZnDrstaIzmene(domain.RobPromenaStanja{Kolicina: -4}, robKnjizenje{kodknj: "D", predznak: "-"}, 1, false),
			want: []repository.Izmena{{Kolona: "kolicner", Dodaj: true, Vrednost: -4.0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("izmene = %#v, want %#v", tt.got, tt.want)
			}
		})
	}
}

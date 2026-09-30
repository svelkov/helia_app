package config

import "testing"

func TestNDuzSintForKnjigovod(t *testing.T) {
	cfg := Config{NDuzSint: 3}

	tests := []struct {
		knjigovod string
		want      int
	}{
		{"Finansijsko", 3},
		{"Pogonsko", 3},
		{"finansijsko", 3},
		{"POGONSKO", 3},
		{"  Finansijsko  ", 3},
		{"Robno", 4},
		{"Blagajna", 4},
		{"", 4},
	}

	for _, tt := range tests {
		if got := cfg.NDuzSintForKnjigovod(tt.knjigovod); got != tt.want {
			t.Errorf("NDuzSintForKnjigovod(%q) = %d, want %d", tt.knjigovod, got, tt.want)
		}
	}
}

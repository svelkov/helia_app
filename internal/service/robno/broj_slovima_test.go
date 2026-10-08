package robno

import "testing"

// TestBrojSlovima checks the amounts in words, also the forms of the nouns of the thousands and the
// millions after one, two to four and the others.
func TestBrojSlovima(t *testing.T) {
	for in, want := range map[float64]string{
		0:          "nula i 00/100",
		1:          "jedan i 00/100",
		12.5:       "dvanaest i 50/100",
		21:         "dvadesetjedan i 00/100",
		102:        "stodva i 00/100",
		1000:       "hiljadu i 00/100",
		1234.56:    "hiljadudvestatridesetčetiri i 56/100",
		2000:       "dvehiljade i 00/100",
		5000:       "pethiljada i 00/100",
		11000:      "jedanaesthiljada i 00/100",
		21000:      "dvadesetjednahiljada i 00/100",
		22000:      "dvadesetdvehiljade i 00/100",
		1000000:    "milion i 00/100",
		2500000:    "dvamilionapetstohiljada i 00/100",
		31000000:   "tridesetjedanmilion i 00/100",
		1000000000: "milijardu i 00/100",
		704241:     "sedamstočetirihiljadedvestačetrdesetjedan i 00/100",
	} {
		if got := BrojSlovima(in); got != want {
			t.Errorf("%v: got %q want %q", in, got, want)
		}
	}
}

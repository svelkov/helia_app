package robno

import (
	"fmt"
	"math"
	"strings"
)

// The words of the Serbian numbers (latinica) for BrojSlovima.
var (
	brojSlovimaJedinice = []string{"", "jedan", "dva", "tri", "četiri", "pet", "šest", "sedam", "osam", "devet"}
	brojSlovimaNaest    = []string{"deset", "jedanaest", "dvanaest", "trinaest", "četrnaest", "petnaest", "šesnaest", "sedamnaest", "osamnaest", "devetnaest"}
	brojSlovimaDesetice = []string{"", "", "dvadeset", "trideset", "četrdeset", "pedeset", "šezdeset", "sedamdeset", "osamdeset", "devedeset"}
	brojSlovimaStotine  = []string{"", "sto", "dvesta", "trista", "četiristo", "petsto", "šeststo", "sedamsto", "osamsto", "devetsto"}
)

// brojSlovimaGrupa is a group of three digits (thousands, millions, billions) with its noun alone (one
// thousand: "hiljadu"), in the singular after a number ending in one, in the paucal (2-4) and in the
// plural, and whether the noun is feminine (jedna, dve).
type brojSlovimaGrupa struct {
	sama, jednina, paukal, mnozina string
	zenski                         bool
}

var brojSlovimaGrupe = []brojSlovimaGrupa{
	{},
	{"hiljadu", "hiljada", "hiljade", "hiljada", true},
	{"milion", "milion", "miliona", "miliona", false},
	{"milijardu", "milijarda", "milijarde", "milijardi", true},
}

// BrojSlovima returns an amount in words in Serbian (latinica), the dinari in words written together
// and the pare as a fraction, e.g. 1234.56 -> "hiljadudvestatridesetčetiri i 56/100" ("nula i 00/100"
// for 0). The sign is not given: the caller passes the absolute value.
func BrojSlovima(iznos float64) string {
	pare := int64(math.Round(iznos * 100))
	dinari, ostatak := pare/100, pare%100
	reci := brojSlovimaCeo(dinari)
	if reci == "" {
		reci = "nula"
	}
	return fmt.Sprintf("%s i %02d/100", reci, ostatak)
}

// brojSlovimaCeo returns a whole number in words ("" for 0).
func brojSlovimaCeo(broj int64) string {
	var delovi []string
	for grupa := len(brojSlovimaGrupe) - 1; grupa >= 0; grupa-- {
		delilac := int64(math.Pow(1000, float64(grupa)))
		trojka := broj / delilac % 1000
		if trojka == 0 {
			continue
		}
		g := brojSlovimaGrupe[grupa]
		if grupa == 0 {
			delovi = append(delovi, brojSlovimaTrojka(trojka, false))
			continue
		}
		// "hiljadu", "milion": one alone is the noun only.
		if trojka == 1 {
			delovi = append(delovi, g.sama)
			continue
		}
		delovi = append(delovi, brojSlovimaTrojka(trojka, g.zenski)+brojSlovimaImenica(trojka, g))
	}
	return strings.Join(delovi, "")
}

// brojSlovimaImenica returns the noun of a group after a number: the singular after 1 (but not 11),
// the paucal after 2-4 (but not 12-14), the plural otherwise.
func brojSlovimaImenica(trojka int64, g brojSlovimaGrupa) string {
	jedinica, desetica := trojka%10, trojka/10%10
	switch {
	case desetica == 1:
		return g.mnozina
	case jedinica == 1:
		return g.jednina
	case jedinica >= 2 && jedinica <= 4:
		return g.paukal
	}
	return g.mnozina
}

// brojSlovimaTrojka returns a number of three digits in words; zenski gives the feminine forms of one
// and two (jedna, dve) before a feminine noun.
func brojSlovimaTrojka(trojka int64, zenski bool) string {
	stotina, desetica, jedinica := trojka/100, trojka/10%10, trojka%10
	reci := brojSlovimaStotine[stotina]
	switch {
	case desetica == 1:
		return reci + brojSlovimaNaest[jedinica]
	default:
		reci += brojSlovimaDesetice[desetica]
	}
	switch {
	case zenski && jedinica == 1:
		reci += "jedna"
	case zenski && jedinica == 2:
		reci += "dve"
	default:
		reci += brojSlovimaJedinice[jedinica]
	}
	return reci
}

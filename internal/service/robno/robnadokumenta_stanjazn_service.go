package robno

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

// The stanje of the prijemnice materijala of the komercijalno (the legacy GlobalProcROBNO.AzurRSTA_ZN and
// AzurDRSTA_ZN): a stavka that meets the kvalitet (zaliha, the legacy DaNe "Z") goes into the zalihe and lowers the
// neraspoložive zalihe (kolicner), one that does not (the legacy "N", a magacin with the analiza) goes into the
// neraspoložive zalihe. Only the vrste with the knjiženje D (ulaz) or P (izlaz) change the stanje; the količina and the
// iznos get the sign of the predznak of the vrsta (a storno is negative).

// robZnPredznak is the value with the sign of the predznak of the vrsta: negative for a storno ("-"), else positive.
func robZnPredznak(v float64, k robKnjizenje) float64 {
	if k.predznak == "-" {
		return -math.Abs(v)
	}
	return math.Abs(v)
}

// robZnRstaNeraspolozivo is the change of the neraspoložive zalihe of the stanje by konto (rsta) by a stavka: an ulaz
// (D) of the zalihe lowers them, one of the neraspoložive zalihe raises them; an izlaz (P) does not change them.
func robZnRstaNeraspolozivo(k robKnjizenje, kolicina float64, zaliha bool) float64 {
	if k.kodknj != "D" {
		return 0
	}
	if zaliha {
		return -kolicina
	}
	return kolicina
}

// robZnDrstaNeraspolozivo is the change of the neraspoložive zalihe of the zalihe by lot (drsta) by a stavka: an ulaz
// (D) lowers them for the zalihe and raises them for the neraspoložive zalihe, an izlaz (P) the other way round.
func robZnDrstaNeraspolozivo(k robKnjizenje, kolicina float64, zaliha bool) float64 {
	switch {
	case k.kodknj == "D" && zaliha, k.kodknj == "P" && !zaliha:
		return -kolicina
	case k.kodknj == "D", k.kodknj == "P":
		return kolicina
	}
	return 0
}

// robZnRstaKolone are the columns of a new stanje by konto (rsta) of a stavka (AzurRSTA_ZN): the kretanje of the month
// by the knjiženje (for the zalihe and the neraspoložive zalihe alike, see robStanjeKolone), the artikal, the konto,
// the magacin, the datum of the first kretanje, the cena (when not 0), the broj stavki 1 and the neraspoložive zalihe
// (not below 0).
func robZnRstaKolone(p domain.RobPromenaStanja, k robKnjizenje, m int, rsifID int64, mag int, zaliha bool) map[string]any {
	kolone := robStanjeKolone(p, k, m)
	kolone["rsifid"], kolone["sifra"], kolone["konto"] = rsifID, p.Sifra, p.Konto
	kolone["mag"], kolone["magaciniid"], kolone["brst"] = mag, p.MagaciniID, 1
	if !p.Danal.IsZero() {
		kolone["zdat"] = p.Danal.Format(common.HtmlLayout)
	}
	if p.Cena != 0 {
		kolone["cena"] = p.Cena
	}
	kolone["kolicner"] = math.Max(robZnRstaNeraspolozivo(k, p.Kolicina, zaliha), 0)
	return kolone
}

// robZnRstaIzmene are the changes of an existing stanje by konto (rsta) by a stavka (AzurRSTA_ZN): the kretanje of the
// month and the neraspoložive zalihe of an ulaz (the legacy does not keep them at 0 here).
func robZnRstaIzmene(p domain.RobPromenaStanja, k robKnjizenje, m int, zaliha bool) []repository.Izmena {
	izmene := robStanjeKretanje(p, k, m)
	if k.kodknj == "D" {
		izmene = append(izmene, repository.Izmena{Kolona: "kolicner", Dodaj: true, Vrednost: robZnRstaNeraspolozivo(k, p.Kolicina, zaliha)})
	}
	return izmene
}

// robZnDrstaKolone are the columns of a new lot (drsta) of a stavka (AzurDRSTA_ZN): the kretanje of the month only for
// the zalihe, the artikal, the konto, the lot (OTK, serija, rok trajanja), the magacin, the cena (when not 0) and the
// neraspoložive zalihe (not below 0, except for a storno).
func robZnDrstaKolone(p domain.RobPromenaStanja, k robKnjizenje, m int, rsifID int64, mag int, zaliha bool) map[string]any {
	kretanje := robKnjizenje{}
	if zaliha {
		kretanje = k
	}
	kolone := robStanjeKolone(p, kretanje, m)
	kolone["rsifid"], kolone["sifra"], kolone["konto"] = rsifID, p.Sifra, p.Konto
	kolone["otk"], kolone["serija"], kolone["roktr"] = p.Otk, p.Serija, p.Roktr
	kolone["mag"], kolone["magaciniid"], kolone["zstatus"] = mag, p.MagaciniID, ""
	if p.Cena != 0 {
		kolone["cena"] = p.Cena
	}
	neraspolozivo := robZnDrstaNeraspolozivo(k, p.Kolicina, zaliha)
	if neraspolozivo < 0 && k.predznak != "-" {
		neraspolozivo = 0
	}
	kolone["kolicner"] = neraspolozivo
	return kolone
}

// robZnDrstaIzmene are the changes of an existing lot (drsta) by a stavka (AzurDRSTA_ZN): the kretanje of the month
// only for the zalihe and the neraspoložive zalihe (not below 0, except for a storno).
func robZnDrstaIzmene(p domain.RobPromenaStanja, k robKnjizenje, m int, zaliha bool) []repository.Izmena {
	var izmene []repository.Izmena
	if zaliha {
		izmene = robStanjeKretanje(p, k, m)
	}
	return append(izmene, repository.Izmena{Kolona: "kolicner", Dodaj: true, NeNegativno: k.predznak != "-",
		Vrednost: robZnDrstaNeraspolozivo(k, p.Kolicina, zaliha)})
}

// robZnKnjizenje returns the knjiženje of the vrsta of a stavka with its količina and iznos signed by the predznak;
// ok is false for a vrsta that does not change the stanje (not D or P).
func (s *RobnoDokumentaResource) robZnKnjizenje(ctx context.Context, tx db.Transaction, p *domain.RobPromenaStanja) (robKnjizenje, bool, error) {
	k, err := s.robDokvrstaKnjizenje(ctx, tx, p.DokvrstaID)
	if err != nil {
		return k, false, err
	}
	if k.kodknj != "D" && k.kodknj != "P" {
		return k, false, nil
	}
	p.Kolicina, p.Iznos = robZnPredznak(p.Kolicina, k), robZnPredznak(p.Iznos, k)
	return k, true, nil
}

// AzurirajRSTAZN changes the stanje of an artikal in a konto of a magacin (rsta) by a stavka of a prijemnica (the legacy
// GlobalProcROBNO.AzurRSTA_ZN), in the transaction of the potvrda: a new one with robZnRstaKolone, an existing one with
// robZnRstaIzmene. zaliha is the legacy DaNe "Z" (else "N").
func (s *RobnoDokumentaResource) AzurirajRSTAZN(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja, zaliha bool) error {
	k, ok, err := s.robZnKnjizenje(ctx, tx, &p)
	if err != nil || !ok {
		return err
	}
	m := robStanjeMesec(p.Danal)
	rstaID, err := s.robno.RstaIDZaIzmenu(ctx, tx, p.MagaciniID, p.Konto, p.Sifra)
	if errors.Is(err, sql.ErrNoRows) {
		rsifID, mag, err := s.robno.RsifIDMag(ctx, tx, p.Sifra, p.MagaciniID)
		if err != nil {
			return err
		}
		return s.robno.StanjeDodaj(ctx, tx, repository.TabelaRsta, robZnRstaKolone(p, k, m, rsifID, mag, zaliha))
	}
	if err != nil {
		return err
	}
	return s.robno.StanjeIzmeni(ctx, tx, repository.TabelaRsta, rstaID, robZnRstaIzmene(p, k, m, zaliha))
}

// AzurirajDRSTAZN changes the zalihe of an artikal by lot (OTK, serija, rok trajanja) in a konto of a magacin (drsta) by
// a stavka of a prijemnica (the legacy GlobalProcROBNO.AzurDRSTA_ZN; always by the lot of the stavka), in the
// transaction of the potvrda: a new lot with robZnDrstaKolone, an existing one with robZnDrstaIzmene. zaliha is the
// legacy DaNe "Z" (else "N").
func (s *RobnoDokumentaResource) AzurirajDRSTAZN(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja, zaliha bool) error {
	k, ok, err := s.robZnKnjizenje(ctx, tx, &p)
	if err != nil || !ok {
		return err
	}
	m := robStanjeMesec(p.Danal)
	rsifID, mag, err := s.robno.RsifIDMag(ctx, tx, p.Sifra, p.MagaciniID)
	if err != nil {
		return err
	}
	rstaID, err := s.robno.DrstaIDZaIzmenu(ctx, tx, p.Konto, p.Sifra, mag, p.Otk, p.Serija, p.Roktr)
	if errors.Is(err, sql.ErrNoRows) {
		return s.robno.StanjeDodaj(ctx, tx, repository.TabelaDrsta, robZnDrstaKolone(p, k, m, rsifID, mag, zaliha))
	}
	if err != nil {
		return err
	}
	return s.robno.StanjeIzmeni(ctx, tx, repository.TabelaDrsta, rstaID, robZnDrstaIzmene(p, k, m, zaliha))
}

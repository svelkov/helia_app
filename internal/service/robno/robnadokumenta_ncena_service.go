package robno

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
)

// IzracunajNCENA computes the prosečna nabavna cena of the artikal (the šifra) in the magacin from the stavke of
// the robni dokumenti of the selected year and firm up to the datum of the nalog and of the document (the
// legacy GlobalProcROBNO.IzracunajNCENA), the stavke by the datum of the nalog, of the document and by their
// order (the vrste 119 and 133 are not counted):
//   - the popis (vrsta 101): the cena of the stanje and of the stavka by their količine (with a positive stanje),
//     else the cena of the stavka; the stanje gets the količina, the saldo its vrednost;
//   - the ulazi of the nabavna cena (the vrste 110, 111, 113, 117, 198, 199): the same with the nabavna cena of the
//     stavka;
//   - the vrste 121 and 129: the same with the količina without its sign and the nabavna cena; the stanje loses the
//     količina, the saldo is the vrednost of the stavka;
//   - the other vrste: an ulaz (dokvrsta.kodknj "D") adds the količina and the vrednost to the stanje and the saldo,
//     an izlaz ("P") takes them away.
//
// Without a cena it is the nabavna cena of the first ulaz of the artikal in the magacin (the vrste 101, 110, 113,
// 117, 121, 129, 198, 131 and 189) from the datum of the nalog.
func (s *RobnoDokumentaResource) IzracunajNCENA(ctx context.Context, danal, dadok time.Time, magaciniID, sifra int) (float64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	rows, err := s.unos.Tekst.GetAllCustom(ctx, `select d.vrd::text as key, coalesce(p.kolic, 0)::float8::text as c1,
			coalesce(p.cena, 0)::float8::text as c2, coalesce(p.ncena, 0)::float8::text as c3,
			upper(trim(coalesce(v.kodknj, ''))) as c4
		from rdok d
		join rpro p on p.rdokid = d.rdokid
		left join dokvrsta v on v.god = d.god and v.kar = d.kar and v.vrd = d.vrd
		where d.vrd <> 119 and d.vrd <> 133 and d.god = $1 and d.kar = $2 and p.sifra = $3 and d.magaciniid = $4
			and d.danal <= $5 and d.dadok <= $6
		order by d.danal asc, d.dadok asc, p.rproid`, "",
		[]any{userSession.SelectedGod, userSession.SelectedKar, sifra, magaciniID, danal, dadok}, "", "")
	if err != nil {
		return 0, err
	}
	var cena, stanje, saldo float64
	if rows != nil {
		for _, r := range *rows {
			vrd := common.StringToInt(r.Key)
			kolic, rcena, ncena := common.StringToFloat64(r.C1), common.StringToFloat64(r.C2), common.StringToFloat64(r.C3)
			switch vrd {
			case 101:
				if stanje+kolic != 0 && stanje > 0 {
					cena = (stanje*cena + kolic*rcena) / (stanje + kolic)
				} else {
					cena = rcena
				}
				stanje += kolic
				saldo += kolic * rcena
			case 110, 111, 113, 117, 198, 199:
				if stanje+kolic != 0 && stanje > 0 {
					cena = (stanje*cena + kolic*ncena) / (stanje + kolic)
				} else {
					cena = ncena
				}
				stanje += kolic
				saldo += kolic * rcena
			case 121, 129:
				if stanje+math.Abs(kolic) != 0 && stanje > 0 {
					cena = (stanje*cena + math.Abs(kolic)*ncena) / (stanje + math.Abs(kolic))
				} else {
					cena = ncena
				}
				stanje -= kolic
				// The legacy "xSaldo -= xSaldo - (KOLIC * CENA)".
				saldo = kolic * rcena
			default:
				switch strings.TrimSpace(r.C4) {
				case "D":
					stanje += kolic
					saldo += kolic * rcena
				case "P":
					stanje -= kolic
					saldo -= kolic * rcena
				}
			}
		}
	}
	if cena != 0 {
		return cena, nil
	}
	prvi, err := s.unos.Tekst.GetAllCustom(ctx, `select coalesce(p.ncena, 0)::float8::text as key
		from rdok d join rpro p on p.rdokid = d.rdokid
		where d.god = $1 and d.kar = $2 and d.danal >= $3 and d.vrd in (101, 110, 113, 117, 121, 129, 198, 131, 189)
			and p.sifra = $4 and d.magaciniid = $5
		order by d.danal asc limit 1`, "", []any{userSession.SelectedGod, userSession.SelectedKar, danal, sifra, magaciniID}, "", "")
	if err != nil || prvi == nil || len(*prvi) == 0 {
		return 0, err
	}
	return common.StringToFloat64((*prvi)[0].Key), nil
}

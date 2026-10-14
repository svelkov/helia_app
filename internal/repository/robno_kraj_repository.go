package repository

import (
	"context"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The statements of the kraj poslovne godine (popis lists, višak/manjak, prepis).

// TableExists reports whether a table exists (to_regclass). tx may be nil.
func (r *RobnoRepository) TableExists(ctx context.Context, tx db.Transaction, table string) (bool, error) {
	var exists bool
	err := r.row(ctx, tx, "table "+table, querybuilder.New(`SELECT to_regclass(?) IS NOT NULL`, table), &exists)
	return exists, err
}

// popisOrder maps the sort of the popis list (a request value) to its ORDER BY expression; the
// expressions use the alias s of rsta.
func popisOrder(tip string) (string, error) {
	switch tip {
	case "", "sifra":
		return "s.sifra", nil
	case "naziv":
		return "coalesce(a.naziv,''),s.sifra", nil
	case "grupa":
		return "coalesce(a.gru,0),s.sifra", nil
	case "grupa-naziv":
		return "coalesce(a.gru,0),coalesce(a.naziv,''),s.sifra", nil
	default:
		return "", fmt.Errorf("unknown popis sort %q", tip)
	}
}

// popisRange adds the optional magacin, konto and sifra ranges of the popis (0 / "" = all).
func popisRange(b *querybuilder.Builder, alias string, p domain.KrajPoslovneGodineParams) {
	b.Where("(? = 0 OR "+alias+".mag = ?)", p.Magacin, p.Magacin).
		Where("(? = '' OR "+alias+".konto >= ?)", p.OdKonta, p.OdKonta).
		Where("(? = '' OR "+alias+".konto <= ?)", p.DoKonta, p.DoKonta).
		Where("(? = 0 OR "+alias+".sifra >= ?)", p.OdSifre, p.OdSifre).
		Where("(? = 0 OR "+alias+".sifra <= ?)", p.DoSifre, p.DoSifre)
}

// PopisZaliha1 returns the popis list of a magacin with TIPZAL 1 (stanje per konto and sifra, the
// cena from rcene or the prosečna cena), in the period of the session. tip is the sort.
func (r *RobnoRepository) PopisZaliha1(ctx context.Context, p domain.KrajPoslovneGodineParams) ([]domain.KrajPopisType1Row, error) {
	order, err := popisOrder(p.Tip)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`SELECT row_number() over (ORDER BY ` + order + `)::int redbr,s.konto,s.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,coalesce(a.gru,0) grupa,coalesce(g.naziv,'') nazivgrupe,
	 CASE WHEN coalesce(m.nacvodzal,0)=3 THEN coalesce(max(s.prosnc),0) ELSE coalesce(max(c.cena),0) END cena,
	 0::numeric kolicina1,0::numeric kolicina2,0::numeric kolicina3,0::numeric ukupnakolicina,
	 coalesce(sum(s.ulaz-s.izlaz),0) stanjezaliha
	 FROM rsta s`).
		// rcene has no mag: one row per god, kar and šifra (the old join on c.mag failed).
		AddJoin("LEFT JOIN rcene c ON c.sifra=s.sifra AND c.god=s.god AND c.kar=s.kar").
		AddJoin("LEFT JOIN rsif a ON a.sifra=s.sifra AND a.god=s.god AND a.kar=s.kar").
		AddJoin("LEFT JOIN rgru g ON g.gru=a.gru AND g.god=s.god AND g.kar=s.kar").
		AddJoin("LEFT JOIN magacini m ON m.mag=s.mag AND m.god=s.god AND m.kar=s.kar")
	if err := inPeriod(ctx, b, "s."); err != nil {
		return nil, err
	}
	popisRange(b, "s", p)
	b.GroupBy("s.konto", "s.sifra", "a.naziv", "a.jm", "a.gru", "g.naziv", "m.nacvodzal").
		Having("? OR sum(s.ulaz-s.izlaz)<>0", p.ObradiNule).
		OrderByExpr(order)
	var rows []domain.KrajPopisType1Row
	err = r.selectRows(ctx, "popis zaliha", &rows, b)
	return rows, err
}

// PopisZaliha2 returns the popis list of a magacin with TIPZAL 2 (per otk, serija and rok from
// drsta, with the stanje of rsta), in the period of the session. tip is the sort.
// The rok is the date of drsta.roktr (the days from 1800-01-01, like RobWinDevDatum) as DD.MM.YYYY,
// empty for 0 (the old coalesce of roktr with an empty text failed on the numeric).
func (r *RobnoRepository) PopisZaliha2(ctx context.Context, p domain.KrajPoslovneGodineParams) ([]domain.KrajPopisType2Row, error) {
	order, err := popisOrder(p.Tip)
	if err != nil {
		return nil, err
	}
	order = strings.ReplaceAll(order, "s.", "d.")
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`SELECT row_number() over (ORDER BY `+order+`)::int redbr,d.konto,d.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,coalesce(a.gru,0) grupa,coalesce(g.naziv,'') nazivgrupe,
	 coalesce(d.otk,'') otk,coalesce(d.serija,'') serija,coalesce(to_char(date '1800-01-01' + nullif(d.roktr, 0)::int, 'DD.MM.YYYY'),'') rok,0::numeric kolicina1,0::numeric kolicina2,0::numeric kolicina3,0::numeric ukupnakolicina,
	 coalesce((SELECT sum(s.ulaz-s.izlaz) FROM rsta s WHERE s.god=? AND s.kar=? AND s.mag=d.mag AND s.sifra=d.sifra AND s.konto=d.konto),0) stanjezaliha
	 FROM drsta d`, s.God, s.Kar).
		AddJoin("LEFT JOIN rsif a ON a.sifra=d.sifra AND a.god=d.god AND a.kar=d.kar").
		AddJoin("LEFT JOIN rgru g ON g.gru=a.gru AND g.god=d.god AND g.kar=d.kar").
		Where("d.god = ?", s.God).
		Where("d.kar = ?", s.Kar)
	popisRange(b, "d", p)
	b.OrderByExpr(order)
	var rows []domain.KrajPopisType2Row
	err = r.selectRows(ctx, "popis zaliha", &rows, b)
	return rows, err
}

// VisakManjak compares the knjigovodstveno stanje of the period of the session with the popis
// document (nalog, dokum, vrd) of the new year p.NovaGod in the same kar, and returns the rows
// with a difference.
func (r *RobnoRepository) VisakManjak(ctx context.Context, p domain.KrajPoslovneGodineParams) ([]domain.KrajVisakManjakRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`SELECT row_number() over (ORDER BY sifra)::int redbr,sifra,konto,naziv,jm,cena,kolicina kolicinapopisa,popis_iznos iznospopisa,
		stanje stanjeknjigovodstveno,iznos iznosknjigovodstveno,
		case when razlika>0 then razlika else 0 end visak,
		case when razlika>0 then (popis_iznos-iznos) else 0 end iznosviska,
		case when razlika<0 then -razlika else 0 end manjak,
		case when razlika<0 then (iznos-popis_iznos) else 0 end iznosmanjka,
		case when razlika>0 and popis_iznos-iznos>0 then popis_iznos-iznos else 0 end finansijskivisak,
		case when razlika<0 and iznos-popis_iznos>0 then iznos-popis_iznos else 0 end finansijskimanjak
		FROM data_rows`).
		WithCTE("book", `SELECT s.konto,s.sifra,sum(s.ulaz-s.izlaz) stanje,sum(s.dug-s.pot) iznos,max(s.cena) cena FROM rsta s
			WHERE s.god=? AND s.kar=? AND (?=0 OR s.mag=?) GROUP BY s.konto,s.sifra`, s.God, s.Kar, p.Magacin, p.Magacin).
		WithCTE("popis", `SELECT p.konto,p.sifra,sum(p.kolic) kolicina,sum(p.iznos) iznos FROM rpro p JOIN rdok d ON d.rdokid=p.rdokid
			WHERE p.god=? AND p.kar=? AND p.nalog=? AND p.dokum=? AND p.vrd=? AND (?=0 OR p.mag=?) GROUP BY p.konto,p.sifra`,
			p.NovaGod, s.Kar, p.Nalog, p.Dokum, p.Vrd, p.Magacin, p.Magacin).
		WithCTE("data_rows", `SELECT b.konto,b.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,b.cena,b.iznos,b.stanje,
			coalesce(p.kolicina,0) kolicina,coalesce(p.iznos,0) popis_iznos,coalesce(p.kolicina,0)-b.stanje razlika
			FROM book b LEFT JOIN popis p ON p.konto=b.konto AND p.sifra=b.sifra LEFT JOIN rsif a ON a.sifra=b.sifra AND a.god=? AND a.kar=?`,
			s.God, s.Kar).
		Where("razlika<>0").
		OrderBy("sifra")
	var rows []domain.KrajVisakManjakRow
	err = r.selectRows(ctx, "visak manjak", &rows, b)
	return rows, err
}

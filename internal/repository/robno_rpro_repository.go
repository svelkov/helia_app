package repository

import (
	"context"
	"database/sql"
	"errors"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The stavke of the robni dokumenti (rpro) of the period of the session.

// RproSledeciRbr returns the next redni broj of the stavke of a document.
func (r *RobnoRepository) RproSledeciRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error) {
	b := querybuilder.New(`select coalesce(max(rbr), 0) + 1 from rpro`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var rbr int
	err := r.row(ctx, tx, "redni broj stavke", b, &rbr)
	return rbr, err
}

// RproSledeciRbrNaloga returns the next redni broj of the stavke of a document of a nalog.
func (r *RobnoRepository) RproSledeciRbrNaloga(ctx context.Context, tx db.Transaction, rnalID, rdokID int64) (int, error) {
	b := querybuilder.New(`select coalesce(max(rbr), 0) + 1 from rpro`).Where("rnalid = ?", rnalID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var rbr int
	err := r.row(ctx, tx, "redni broj stavke", b, &rbr)
	return rbr, err
}

// RproPopisStavka reads (and locks) a stavka of a popis: količina, cena, nabavna cena, iznos and
// poreska oznaka.
func (r *RobnoRepository) RproPopisStavka(ctx context.Context, tx db.Transaction, rproID int64) (kolicina, cena, ncena, iznos float64, po int, err error) {
	b := querybuilder.New(`select coalesce(kolic, 0)::float8, coalesce(cena, 0)::float8, coalesce(ncena, 0)::float8, coalesce(iznos, 0)::float8,
		coalesce(po, 0) from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, 0, 0, 0, err
	}
	b.ForUpdate()
	err = r.row(ctx, tx, "stavka popisa", b, &kolicina, &cena, &ncena, &iznos, &po)
	return kolicina, cena, ncena, iznos, po, err
}

// RproZadnjiUlaz returns the cena and the nabavna cena of the last stavka of an artikal in a
// document of the vrsta vrd (the last datum dokumenta, then the last stavka); 0, 0 without one.
func (r *RobnoRepository) RproZadnjiUlaz(ctx context.Context, vrd, sifra int) (cena, ncena float64, err error) {
	b := querybuilder.New(`select coalesce(p.cena, 0)::float8, coalesce(p.ncena, 0)::float8 from rpro p`).
		AddJoin("join rdok r on r.rdokid = p.rdokid")
	if err := inPeriod(ctx, b, "r."); err != nil {
		return 0, 0, err
	}
	b.Where("r.vrd = ?", vrd).Where("p.sifra = ?", sifra).OrderByExpr("r.dadok desc nulls last, p.rproid desc").Limit(1)
	err = r.row(ctx, nil, "zadnji ulaz artikla", b, &cena, &ncena)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, nil
	}
	return cena, ncena, err
}

// RproStavkaBrisanje reads (and locks) a stavka to delete: šifra, redni broj, konto, količina and
// iznos; the error wraps sql.ErrNoRows when it is not a stavka of the period.
func (r *RobnoRepository) RproStavkaBrisanje(ctx context.Context, tx db.Transaction, rproID int64) (sifra, rbr int, konto string, kolicina, iznos float64, err error) {
	b := querybuilder.New(`select coalesce(sifra, 0), coalesce(rbr, 0), coalesce(konto, ''), coalesce(kolic, 0)::float8, coalesce(iznos, 0)::float8
		from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, "", 0, 0, err
	}
	b.ForUpdate()
	err = r.row(ctx, tx, "stavka", b, &sifra, &rbr, &konto, &kolicina, &iznos)
	return sifra, rbr, konto, kolicina, iznos, err
}

// RproSmanjiRbr gives the stavke of a document after the redni broj rbr the redni broj less by one
// (after a stavka is deleted).
func (r *RobnoRepository) RproSmanjiRbr(ctx context.Context, tx db.Transaction, rdokID int64, rbr int) error {
	b := querybuilder.NewUpdate("rpro").Set("rbr = rbr - 1").Where("rdokid = ?", rdokID).Where("rbr > ?", rbr)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "redni brojevi stavki", b)
	return err
}

// RproStara is a saved stavka before its change or deletion.
type RproStara struct {
	RdokID          int64
	Konto, Taxcat   string
	Kolicina, Iznos float64
	Po, Rbr         int
}

// RproStaraStavka reads (and locks) a saved stavka; the error wraps sql.ErrNoRows when it is not a
// stavka of the period.
func (r *RobnoRepository) RproStaraStavka(ctx context.Context, tx db.Transaction, rproID int64) (RproStara, error) {
	b := querybuilder.New(`select coalesce(rdokid, 0)::bigint, coalesce(konto, ''), coalesce(taxcat, ''), coalesce(kolic, 0)::float8,
		coalesce(iznos, 0)::float8, coalesce(po, 0), coalesce(rbr, 0) from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RproStara{}, err
	}
	b.ForUpdate()
	var st RproStara
	err := r.row(ctx, tx, "stavka", b, &st.RdokID, &st.Konto, &st.Taxcat, &st.Kolicina, &st.Iznos, &st.Po, &st.Rbr)
	return st, err
}

// RproPrenumerisi gives the stavke of a document after the redni broj rbr the redni brojevi rbr,
// rbr+1, ... in their order (after a stavka is deleted).
func (r *RobnoRepository) RproPrenumerisi(ctx context.Context, tx db.Transaction, rdokID int64, rbr int) error {
	return r.prenumerisi(ctx, tx, "rbr = x.n", rdokID, rbr)
}

// RproPrenumerisiSaBrst is RproPrenumerisi that also sets brst to the new redni broj.
func (r *RobnoRepository) RproPrenumerisiSaBrst(ctx context.Context, tx db.Transaction, rdokID int64, rbr int) error {
	return r.prenumerisi(ctx, tx, "rbr = x.n, brst = x.n", rdokID, rbr)
}

// prenumerisi numbers the stavke of a document after rbr again; set is the static SET clause.
func (r *RobnoRepository) prenumerisi(ctx context.Context, tx db.Transaction, set string, rdokID int64, rbr int) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("rpro").Set(set).
		From(`(select rproid, (? - 1 + row_number() over (order by rbr, rproid))::int as n from rpro
			where rdokid = ? and rbr > ? and god = ? and kar = ?) x`, rbr, rdokID, rbr, s.God, s.Kar).
		Where("rpro.rproid = x.rproid")
	_, err = r.exec(ctx, tx, "redni brojevi stavki", b)
	return err
}

// RproKolicina returns the količina of a stavka of the artikal in the konto (trimmed); found is false
// without it.
func (r *RobnoRepository) RproKolicina(ctx context.Context, rproID int64, sifra int, konto string) (kolicina float64, found bool, err error) {
	b := querybuilder.New(`select coalesce(kolic, 0)::float8 from rpro`).
		Where("rproid = ?", rproID).Where("sifra = ?", sifra).Where("trim(konto) = ?", konto)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	err = r.row(ctx, nil, "količina stavke", b, &kolicina)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return kolicina, err == nil, err
}

// RproStaraPromena reads (and locks) a saved stavka of a document as the change of the stanje it
// made: šifra, konto, količina, iznos, OTK broj, serija and rok trajanja. The error wraps
// sql.ErrNoRows when it is not a stavka of the document in the period.
func (r *RobnoRepository) RproStaraPromena(ctx context.Context, tx db.Transaction, rproID, rdokID int64) (domain.RobPromenaStanja, error) {
	b := querybuilder.New(`select coalesce(sifra, 0), coalesce(konto, ''), coalesce(kolic, 0)::float8,
		coalesce(iznos, 0)::float8, coalesce(otk, ''), coalesce(serija, ''), coalesce(roktr, 0)::float8 from rpro`).
		Where("rproid = ?", rproID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return domain.RobPromenaStanja{}, err
	}
	b.ForUpdate()
	var p domain.RobPromenaStanja
	var rok float64
	if err := r.row(ctx, tx, "stavka", b, &p.Sifra, &p.Konto, &p.Kolicina, &p.Iznos, &p.Otk, &p.Serija, &rok); err != nil {
		return domain.RobPromenaStanja{}, err
	}
	p.Roktr = int64(rok)
	return p, nil
}

// RproRbr returns the redni broj of a stavka; the error wraps sql.ErrNoRows when it is not a stavka
// of the period.
func (r *RobnoRepository) RproRbr(ctx context.Context, tx db.Transaction, rproID int64) (int, error) {
	b := querybuilder.New(`select coalesce(rbr, 0) from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var rbr int
	err := r.row(ctx, tx, "redni broj stavke", b, &rbr)
	return rbr, err
}

// RproZadnjaCenaProdavnice returns the cene of the last stavka of the artikal to the prodavnica of a
// document (the same pkto, pana and vrsta, up to its datum): key = maloprodajna cena, c1 = rabat,
// c2 = maloprodajna cena with the porez, c3 = marža, c4 = datum (DD.MM.YYYY).
func (r *RobnoRepository) RproZadnjaCenaProdavnice(ctx context.Context, rdokID int64, sifra int) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(r.mcena, 0)::float8::text as key,
			coalesce(r.vra, 0)::float8::text as c1, coalesce(r.mcenap, 0)::float8::text as c2, coalesce(r.mma, 0)::float8::text as c3,
			coalesce(to_char(r.dadok, 'DD.MM.YYYY'), '') as c4
		from rdok d`).
		AddJoin(`join rpro r on r.god = d.god and r.kar = d.kar and r.pkto = d.pkto and r.pana = d.pana and r.vrd = d.vrd
			and r.dadok <= d.dadok`).
		Where("d.rdokid = ?", rdokID).Where("r.sifra = ?", sifra)
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.OrderBy("r.dadok desc, r.rdokid desc, r.rproid desc").Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "zadnja cena prodavnice", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// RproInterniForma returns a saved stavka of an interni prenos for its form: key = šifra, c1 = konto,
// c2 = količina, c3 = cena, c4 = rabat, c5 = nabavna cena (ncena, else pcena), c6 = marža,
// c7 = maloprodajna cena|with the porez|rok trajanja, c8 = OTK broj|serija.
func (r *RobnoRepository) RproInterniForma(ctx context.Context, rproID, rdokID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select r.sifra::text as key, coalesce(r.konto, '') as c1,
			coalesce(r.kolic, 0)::float8::text as c2, coalesce(r.cena, 0)::float8::text as c3, coalesce(r.vra, 0)::float8::text as c4,
			coalesce(nullif(r.ncena, 0), r.pcena, 0)::float8::text as c5, coalesce(r.mma, 0)::float8::text as c6,
			(coalesce(r.mcena, 0)::float8::text || '|' || coalesce(r.mcenap, 0)::float8::text || '|' || coalesce(r.roktr, 0)::bigint::text) as c7,
			(coalesce(r.otk, '') || '|' || coalesce(r.serija, '')) as c8
		from rpro r`).Where("r.rproid = ?", rproID).Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "stavka", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// RproStaraKolicina reads (and locks) a saved stavka of a document: šifra, konto, količina and
// iznos. The error wraps sql.ErrNoRows when it is not a stavka of the document in the period.
func (r *RobnoRepository) RproStaraKolicina(ctx context.Context, tx db.Transaction, rproID, rdokID int64) (sifra int, konto string, kolicina, iznos float64, err error) {
	b := querybuilder.New(`select coalesce(sifra, 0), coalesce(konto, ''), coalesce(kolic, 0)::float8, coalesce(iznos, 0)::float8
		from rpro`).Where("rproid = ?", rproID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", 0, 0, err
	}
	b.ForUpdate()
	err = r.row(ctx, tx, "stavka", b, &sifra, &konto, &kolicina, &iznos)
	return sifra, konto, kolicina, iznos, err
}

// RproRdokID returns the document of a stavka; the error wraps sql.ErrNoRows when it is not a
// stavka of the period.
func (r *RobnoRepository) RproRdokID(ctx context.Context, tx db.Transaction, rproID int64) (int64, error) {
	b := querybuilder.New(`select coalesce(rdokid, 0)::bigint from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var rdokID int64
	err := r.row(ctx, tx, "dokument stavke", b, &rdokID)
	return rdokID, err
}

// RproPorudzbenicaForma returns a saved stavka of a porudžbenica for its form: key = šifra,
// c1 = konto, c2 = količina, c3 = nabavna cena, c4 = poreska oznaka, c5 = its stopa on the datum.
func (r *RobnoRepository) RproPorudzbenicaForma(ctx context.Context, rproID, rdokID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select r.sifra::text as key, coalesce(r.konto, '') as c1,
			coalesce(r.kolic, 0)::float8::text as c2, coalesce(r.ncena, 0)::float8::text as c3, coalesce(r.po, 0)::text as c4,
			`+poreskaStopa("r.po", "r.dadok")+`::text as c5
		from rpro r`).Where("r.rproid = ?", rproID).Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "stavka", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// FinansijskoRepository holds the statements of the finansijsko module that are not plain CRUD of
// one entity: the stavke of a nalog, the bank statements (izvodi) and the database functions that
// copy and reverse a nalog.
type FinansijskoRepository struct {
	db db.Database
}

// NewFinansijskoRepository returns the repository of the finansijsko module.
func NewFinansijskoRepository(database db.Database) *FinansijskoRepository {
	return &FinansijskoRepository{db: database}
}

// row runs a single-row query on the transaction (nil: on the database) and scans it; a missing row
// is an error that wraps sql.ErrNoRows.
func (r *FinansijskoRepository) row(ctx context.Context, tx db.Transaction, what string, b *querybuilder.Builder, dest ...any) error {
	q, args, err := b.Build()
	if err != nil {
		return fmt.Errorf("build %s: %w", what, err)
	}
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, q, args...)
	} else {
		row = r.db.QueryRowContext(ctx, q, args...)
	}
	if err := row.Scan(dest...); err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	return nil
}

// exec runs a statement in the transaction (nil: on the database).
func (r *FinansijskoRepository) exec(ctx context.Context, tx db.Transaction, what string, b *querybuilder.Builder) error {
	q, args, err := b.Build()
	if err != nil {
		return fmt.Errorf("build %s: %w", what, err)
	}
	if tx != nil {
		_, err = tx.ExecContext(ctx, q, args...)
	} else {
		_, err = r.db.ExecContext(ctx, q, args...)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	return nil
}

// bilsIznosiPoNaredbi is the number of rows of the bilans stanja one statement of BilsUpisiIznose writes (7
// arguments a row: far under the limit of the arguments of a statement).
const bilsIznosiPoNaredbi = 1000

// BilsUpisiIznose writes the iznosi of the rows of the bilans stanja of the period computed by its obrada (the legacy
// HModify of every BILS row of ObradaBILS) with the audit columns, as one UPDATE ... FROM (VALUES ...) for up to
// bilsIznosiPoNaredbi rows instead of one statement a row; nothing without rows.
func (r *FinansijskoRepository) BilsUpisiIznose(ctx context.Context, tx db.Transaction, redovi []domain.BilsIznosi) error {
	if len(redovi) == 0 {
		return nil
	}
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	for start := 0; start < len(redovi); start += bilsIznosiPoNaredbi {
		deo := redovi[start:min(start+bilsIznosiPoNaredbi, len(redovi))]
		vrednosti := make([]string, len(deo))
		args := make([]any, 0, len(deo)*7)
		for i, red := range deo {
			vrednosti[i] = "(?::bigint, ?::numeric, ?::integer, ?::numeric, ?::integer, ?::numeric, ?::integer)"
			args = append(args, red.BilsID, red.TGod, red.TGodH, red.PGod, red.PGodH, red.PGodPS, red.PGodHPS)
		}
		b := querybuilder.NewUpdate("bils").
			Set("tgod = v.tgod, tgodh = v.tgodh, pgod = v.pgod, pgodh = v.pgodh, pgodps = v.pgodps, pgodhps = v.pgodhps").
			Set("xopizmene = ?", s.UserName).
			Set("xdatizmene = now()").
			From("(values "+strings.Join(vrednosti, ", ")+") v(bilsid, tgod, tgodh, pgod, pgodh, pgodps, pgodhps)", args...).
			Where("bils.bilsid = v.bilsid")
		if err := inPeriod(ctx, b, "bils."); err != nil {
			return err
		}
		if err := r.exec(ctx, tx, "iznosi bilansa stanja", b); err != nil {
			return err
		}
	}
	return nil
}

// FproZadnjiRbr returns the largest redni broj of the stavke of a nalog (0 without stavke).
func (r *FinansijskoRepository) FproZadnjiRbr(ctx context.Context, idFnal int64) (int64, error) {
	b := querybuilder.New(`select coalesce(max(rbr), 0) from fpro`).Where("idfnal = ?", idFnal)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var rbr int64
	err := r.row(ctx, nil, "zadnji redni broj", b, &rbr)
	return rbr, err
}

// FproSumeNaloga returns the duguje, the potražuje and the saldo of the stavke of a nalog.
func (r *FinansijskoRepository) FproSumeNaloga(ctx context.Context, idFnal int64) (duguje, potrazuje, saldo float64, err error) {
	b := querybuilder.New(`select
		coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) as duguje,
		coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as potrazuje,
		coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) - coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as saldo
		from fpro`).Where("idfnal = ?", idFnal)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, 0, err
	}
	err = r.row(ctx, nil, "sume naloga", b, &duguje, &potrazuje, &saldo)
	return duguje, potrazuje, saldo, err
}

// FunkcijaPostoji tells whether the database has the function with the signature (e.g.
// "fn_x(bigint,text)").
func (r *FinansijskoRepository) FunkcijaPostoji(ctx context.Context, potpis string) (bool, error) {
	var ok bool
	err := r.row(ctx, nil, "funkcija "+potpis, querybuilder.New(`select to_regprocedure(?) is not null`, potpis), &ok)
	return ok, err
}

// KopijaNaloga is the new nalog a copy or a storno of a nalog makes.
type KopijaNaloga struct {
	Nalog        int64
	Tipdok       string
	Danal, Datob string // the datumi as text (yyyy-mm-dd)
	Opis         string
	God, Kar     int
	Operater     string
}

// FnKopirajNalog copies a nalog with its stavke into a new nalog (fn_kopiraj_nalog_sa_stavkama) and
// returns the new idfnal. The ids are passed to the function as they are (the user's decision: the
// function works on primary keys, no tenant check here).
func (r *FinansijskoRepository) FnKopirajNalog(ctx context.Context, idFnal int64, k KopijaNaloga, idTipdok int64) (int64, error) {
	b := querybuilder.New(`select fn_kopiraj_nalog_sa_stavkama(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		idFnal, k.Nalog, k.Tipdok, k.Danal, k.Datob, k.Opis, idTipdok, k.God, k.Kar, k.Operater)
	var id int64
	err := r.row(ctx, nil, "kopiranje naloga", b, &id)
	return id, err
}

// FnKopirajStavke copies the stavke of a nalog into an existing nalog
// (fn_kopiraj_stavke_u_postojeci_nalog); no tenant check (see FnKopirajNalog).
func (r *FinansijskoRepository) FnKopirajStavke(ctx context.Context, izvorID, ciljID int64, operater string) (sql.NullString, error) {
	b := querybuilder.New(`select fn_kopiraj_stavke_u_postojeci_nalog(?, ?, ?)`, izvorID, ciljID, operater)
	var res sql.NullString
	err := r.row(ctx, nil, "kopiranje stavki", b, &res)
	return res, err
}

// FnStornirajNalog reverses a nalog with its stavke into a new nalog
// (fn_storniraj_nalog_sa_stavkama) and returns the new idfnal; no tenant check (see FnKopirajNalog).
func (r *FinansijskoRepository) FnStornirajNalog(ctx context.Context, idFnal int64, k KopijaNaloga) (int64, error) {
	b := querybuilder.New(`select fn_storniraj_nalog_sa_stavkama(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		idFnal, k.Nalog, k.Tipdok, k.Danal, k.Datob, k.Opis, k.God, k.Kar, k.Operater)
	var id int64
	err := r.row(ctx, nil, "storniranje naloga", b, &id)
	return id, err
}

// FizvzagDodaj adds the header of an izvod (with its god and kar) and returns its idfizvzag.
func (r *FinansijskoRepository) FizvzagDodaj(ctx context.Context, tx db.Transaction, h *domain.Fizvzag) (int64, error) {
	b := querybuilder.New(`insert into fizvzag (god, kar, brrac, izvbr, datizv, konto, sifra,
		prstanje, ukdug, ukpot, nstanje, ukbrst, nalog, tipdok, izvsts, idbanke, xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.God, h.Kar, h.Brrac, h.Izvbr, h.Datizv, h.Konto, h.Sifra, h.Prstanje, h.Ukdug,
		h.Ukpot, h.Nstanje, h.Ukbrst, h.Nalog, h.Tipdok, h.Izvsts, h.IDbanke, h.Xdatunosa, h.Xopunos).
		Returning("idfizvzag")
	var id int64
	err := r.row(ctx, tx, "zaglavlje izvoda", b, &id)
	return id, err
}

// FizvdetBrisi deletes the stavke of an izvod of the period god/kar (the one of its header).
func (r *FinansijskoRepository) FizvdetBrisi(ctx context.Context, tx db.Transaction, idFizvzag int64, god, kar int) error {
	b := querybuilder.New(`delete from fizvdet`).Where("idfizvzag = ?", idFizvzag).Where("god = ?", god).Where("kar = ?", kar)
	return r.exec(ctx, tx, "brisanje stavki izvoda", b)
}

// fizvdetKolone are the columns of a stavka of an izvod that are written (not idfizvdet).
var fizvdetKolone = []string{
	"god", "kar", "brrac", "izvbr", "datizv", "rbr",
	"konto", "sifra", "iznos", "kat", "vrd", "nsedprim",
	"brracup", "osnplac", "sdozn", "duguje", "potrazuje",
	"modelzad", "pnabrzad", "mododob", "pnabrodob", "prekl",
	"xdatunosa", "xopunos", "idfizvzag",
}

// FizvdetDodaj adds the stavke of an izvod in one statement; nothing without stavke.
func (r *FinansijskoRepository) FizvdetDodaj(ctx context.Context, tx db.Transaction, stavke []domain.Fizvdet, idFizvzag int64) error {
	if len(stavke) == 0 {
		return nil
	}
	red := "(" + strings.TrimSuffix(strings.Repeat("?, ", len(fizvdetKolone)), ", ") + ")"
	redovi := make([]string, len(stavke))
	args := make([]any, 0, len(stavke)*len(fizvdetKolone))
	for i, s := range stavke {
		redovi[i] = red
		args = append(args, s.God, s.Kar, s.Brrac, s.Izvbr, s.Datizv, s.Rbr, s.Konto, s.Sifra,
			s.Iznos, s.Kat, s.Vrd, s.Nsedprim, s.Brracup,
			s.Osnplac, s.Sdozn, s.Duguje, s.Potrazuje, s.Modelzad, s.Pnabrzad,
			s.Mododob, s.Pnabrodob, s.Prekl, s.Xdatunosa, s.Xopunos, idFizvzag)
	}
	b := querybuilder.New(`insert into fizvdet (`+strings.Join(fizvdetKolone, ", ")+`) values `+strings.Join(redovi, ", "), args...)
	return r.exec(ctx, tx, "stavke izvoda", b)
}

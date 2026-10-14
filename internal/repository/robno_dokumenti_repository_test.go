package repository

import (
	"context"
	"testing"

	"helia/internal/infrastructure/db"
)

// answer returns a fake answer for every statement: one row of the given values for a query, and
// `affected` rows for an exec.
func answer(affected int64, vals ...any) func(string, []any) drvResult {
	cols := make([]string, len(vals))
	for i := range cols {
		cols[i] = "c"
	}
	return func(string, []any) drvResult {
		r := drvResult{cols: cols, affected: affected}
		if len(vals) > 0 {
			r.rows = [][]any{vals}
		}
		return r
	}
}

func noRows(string, []any) drvResult { return drvResult{cols: []string{"key"}} }

// robnoDokumentiCases are the statements of the entry of the robni dokumenti and their stavke.
var robnoDokumentiCases = []goldenCase{
	// rdok
	{name: "rdok by broj with the number of stavke", respond: answer(0, int64(5), int64(2)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RdokPoBroju(ctx, 3, 89, 12)
			return err
		}},
	{name: "rdok by broj not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, found, err := r.RdokPoBroju(ctx, 3, 89, 12)
			if found {
				return errFound
			}
			return err
		}},
	{name: "rdok add iznos and stavke", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokDodajIznos(ctx, tx, 41, 12.5, 1)
		}},
	{name: "rdok add iznos of a document of another period", respond: answer(0), inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokDodajIznos(ctx, tx, 41, 12.5, 1)
		}},
	{name: "rdok take a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokOduzmiStavku(ctx, tx, 41, 12.5)
		}},
	// rpro
	{name: "rpro next rbr", respond: answer(0, int64(4)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproSledeciRbr(ctx, tx, 41)
			return err
		}},
	// rsif, rsta, drsta
	{name: "rsif artikal", respond: answer(0, int64(77), "Vijak", "kom"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, _, err := r.RsifArtikal(ctx, 15)
			return err
		}},
	{name: "rsif id and po in a transaction", respond: answer(0, int64(77), int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, err := r.RsifIDPo(ctx, tx, 15)
			return err
		}},
	{name: "rsta stanje and vpcena", respond: answer(0, 3.0, 9.5),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RstaStanjeVpcena(ctx, 3, " 1320 ", 15)
			return err
		}},
	{name: "rsta set vpcena", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RstaPostaviVpcena(ctx, tx, 3, "1320", 15, 9.5)
		}},
	{name: "drsta stanje of the artikal", respond: answer(0, 3.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.DrstaStanjeArtikla(ctx, 77)
			return err
		}},
	{name: "drsta lotovi", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.DrstaLotovi(ctx, 12, 77)
			return err
		}},
	{name: "drsta set cene", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DrstaPostaviCene(ctx, tx, 3, 15, 9.5)
		}},
	// popis
	{name: "rdok pkto and pana", respond: answer(0, "2040", "15"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RdokPktoPana(ctx, 41)
			return err
		}},
	{name: "rdok pkto, pana and stopa", respond: answer(0, "2040", "15", 20.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, err := r.RdokPktoPanaStopa(ctx, tx, 41, 3)
			return err
		}},
	{name: "rdok stopa", respond: answer(0, 20.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RdokStopa(ctx, tx, 41, 3)
			return err
		}},
	{name: "rdok add vrednosti", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokDodajVrednosti(ctx, tx, 41, 100, 80, 20, 100)
		}},
	{name: "rdok take vrednosti", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokOduzmiVrednosti(ctx, tx, 41, 100, 80, 20, 100)
		}},
	{name: "rpro next rbr of the nalog", respond: answer(0, int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproSledeciRbrNaloga(ctx, tx, 9, 41)
			return err
		}},
	{name: "rpro stavka of a popis for update", respond: answer(0, 2.0, 5.0, 4.0, 10.0, int64(3)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, _, _, err := r.RproPopisStavka(ctx, tx, 77)
			return err
		}},
	{name: "magacin tipmag", respond: answer(0, "V"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagacinTipmag(ctx, 3)
			return err
		}},
	{name: "rsif artikal with cene", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RsifArtikalSaCenama(ctx, " 15 ")
			return err
		}},
	// gradilište
	{name: "rdok pkto and pana in a transaction", respond: answer(0, "2040", "15"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RdokPktoPanaTx(ctx, tx, 41)
			return err
		}},
	{name: "rpro last ulaz of the artikal", respond: answer(0, 5.0, 4.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RproZadnjiUlaz(ctx, 113, 15)
			return err
		}},
	{name: "rpro stavka to delete for update", respond: answer(0, int64(15), int64(2), "1320", 2.0, 10.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, _, _, _, err := r.RproStavkaBrisanje(ctx, tx, 77)
			return err
		}},
	{name: "rpro renumber after a deleted stavka", respond: answer(3), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RproSmanjiRbr(ctx, tx, 41, 2)
		}},
	{name: "rsif artikal by text", respond: answer(0, int64(15), "Vijak", "kom"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, _, err := r.RsifArtikalPoTekstu(ctx, " 15 ")
			return err
		}},
	// profakture
	{name: "rdok profak and bi", respond: answer(0, "4, 7", "123"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RdokProfakBi(ctx, 41)
			return err
		}},
}

type foundError struct{}

func (foundError) Error() string { return "found, want not found" }

var errFound error = foundError{}

func TestRobnoDokumentiGolden(t *testing.T) {
	runGolden(t, "robno_dokumenti.golden", robnoDokumentiCases)
}

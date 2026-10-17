package repository

import (
	"context"
	"testing"
	"time"

	"helia/internal/infrastructure/db"
)

var robnoKnjizenjeCases = []goldenCase{
	{name: "rsta add a row", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.StanjeDodaj(ctx, tx, TabelaRsta, map[string]any{
				"rsifid": int64(77), "sifra": 15, "konto": "1320", "mag": 1, "magaciniid": 3,
				"dug": 10.0, "mdug_3": 10.0, "ulaz": 1.0, "mul_3": 1.0, "brst": 1,
				"god": 1999, // the session decides
			})
		}},
	{name: "drsta add a lot", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.StanjeDodaj(ctx, tx, TabelaDrsta, map[string]any{"sifra": 15, "otk": "A1", "serija": "S1", "roktr": int64(81234), "zstatus": ""})
		}},
	{name: "rsta change a row", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.StanjeIzmeni(ctx, tx, TabelaRsta, 5, []Izmena{
				{Kolona: "dug", Dodaj: true, Vrednost: 10.0},
				{Kolona: "mdug_3", Dodaj: true, Vrednost: 10.0},
				{Kolona: "brst", Dodaj: true, Vrednost: 1},
				{Kolona: "cena", Vrednost: 9.5},
			})
		}},
	{name: "drsta change the neraspoložive zalihe, not below 0", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.StanjeIzmeni(ctx, tx, TabelaDrsta, 5, []Izmena{
				{Kolona: "ulaz", Dodaj: true, Vrednost: 2.0},
				{Kolona: "kolicner", Dodaj: true, NeNegativno: true, Vrednost: -2.0},
			})
		}},
	{name: "drsta change a row of another period", respond: answer(0), inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.StanjeIzmeni(ctx, tx, TabelaDrsta, 5, []Izmena{{Kolona: "cena", Vrednost: 9.5}, {Kolona: "ncena", Vrednost: 9.5}})
		}},
	{name: "rsta without changes", inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.StanjeIzmeni(ctx, tx, TabelaRsta, 5, nil)
		}},
	{name: "rsta id for update", respond: answer(0, int64(5)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RstaIDZaIzmenu(ctx, tx, 3, "1320", 15)
			return err
		}},
	{name: "drsta id for update", respond: answer(0, int64(6)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.DrstaIDZaIzmenu(ctx, tx, "1320", 15, 1, "A1", "S1", 81234)
			return err
		}},
	{name: "rsif id or 0", respond: answer(0, int64(0)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RsifIDIliNula(ctx, tx, 15)
			return err
		}},
	{name: "rsif id and mag", respond: answer(0, int64(77), int64(1)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RsifIDMag(ctx, tx, 15, 3)
			return err
		}},
	{name: "dokvrsta knjiženje by id", respond: answer(0, "d", "PLT ", "+"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.DokvrstaKnjizenjePoID(ctx, tx, 8)
			return err
		}},
	{name: "dokvrsta knjiženje of the vrsta", respond: answer(0, "D", "PLT", "+"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.DokvrstaKnjizenjeVrste(ctx, tx, 89)
			return err
		}},
	{name: "initable in a transaction", respond: answer(0, "KAL;ULS"), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.IniVrednostTx(ctx, tx, "GRUPADOK", "AZURIRAJUCENU")
			return err
		}},
	{name: "rpro cene", respond: answer(0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, int64(3)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproCeneStavke(ctx, tx, 77)
			return err
		}},
	{name: "rdok vrd and dadok", respond: answer(0, int64(89), nil), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := r.RdokVrdDadok(ctx, tx, 41)
			return err
		}},
	{name: "rnal magacin", respond: answer(0, int64(3), int64(1), "01", nil), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RnalMagacinNaloga(ctx, tx, 9)
			return err
		}},
	{name: "rcene set cene", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RcenePostaviCene(ctx, tx, 15, []Izmena{{Kolona: "cena", Vrednost: 9.5}, {Kolona: "mcenap", Vrednost: 12.0}})
		}},
	{name: "rnal post", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RnalKnjizi(ctx, tx, 9, -1, 0, 10)
		}},
	{name: "rdok post a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokKnjiziStavku(ctx, tx, 41, 10, 2)
		}},
	{name: "rdok take a stavka back", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RdokVratiStavku(ctx, tx, 41, -10)
		}},
	{name: "rpor stopa", respond: answer(0, 20.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RporStopa(ctx, tx, 3, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
			return err
		}},
}

func TestRobnoKnjizenjeGolden(t *testing.T) {
	runGolden(t, "robno_knjizenje.golden", robnoKnjizenjeCases)
}

func TestKnjizenjeRejectsColumns(t *testing.T) {
	tests := []struct {
		name string
		run  func(ctx context.Context, r *RobnoRepository) error
	}{
		{"insert of an unknown column", func(ctx context.Context, r *RobnoRepository) error {
			return r.StanjeDodaj(ctx, nil, TabelaRsta, map[string]any{"cena; drop table rsta": 1})
		}},
		{"insert of a column of the lot into rsta", func(ctx context.Context, r *RobnoRepository) error {
			return r.StanjeDodaj(ctx, nil, TabelaRsta, map[string]any{"otk": "A1"})
		}},
		{"update of an unknown column", func(ctx context.Context, r *RobnoRepository) error {
			return r.StanjeIzmeni(ctx, nil, TabelaDrsta, 5, []Izmena{{Kolona: "mdug_13", Vrednost: 1}})
		}},
		{"unknown table", func(ctx context.Context, r *RobnoRepository) error {
			return r.StanjeDodaj(ctx, nil, StanjeTabela("rsif"), map[string]any{"sifra": 1})
		}},
		{"cene of a column of the stanje", func(ctx context.Context, r *RobnoRepository) error {
			return r.RcenePostaviCene(ctx, nil, 15, []Izmena{{Kolona: "brst", Vrednost: 1}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, f := newDrvDB(t, answer(1))
			if err := tt.run(robnoCtx(), NewRobnoRepository(d)); err == nil {
				t.Fatal("want an error")
			}
			if len(f.calls) != 0 {
				t.Fatalf("%d statements ran, want none", len(f.calls))
			}
		})
	}
}

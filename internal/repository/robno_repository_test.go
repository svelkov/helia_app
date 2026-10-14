package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// robnoCase is one statement of the RobnoRepository: the scripted answer of the fake driver, the
// call, and the exact SQL (whitespace squashed) and arguments it must send.
type robnoCase struct {
	name     string
	respond  drvResult
	run      func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error
	inTx     bool
	wantSQL  string
	wantArgs []any
	wantErr  error
}

// robnoCtx is the session of the tests: god 2026, kar 7.
func robnoCtx() context.Context {
	return domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7})
}

func runRobnoCases(t *testing.T, tests []robnoCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, f := newDrvDB(t, func(string, []any) drvResult { return tt.respond })
			r := NewRobnoRepository(d)
			var tx db.Transaction
			if tt.inTx {
				var err error
				if tx, err = d.Beginx(); err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback() }()
			}
			err := tt.run(robnoCtx(), r, tx)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err %v, want %v", err, tt.wantErr)
			}
			if len(f.calls) != 1 {
				t.Fatalf("%d statements, want 1", len(f.calls))
			}
			if f.calls[0].inTx != tt.inTx {
				t.Errorf("in transaction = %v, want %v", f.calls[0].inTx, tt.inTx)
			}
			if got := squash(f.calls[0].query); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			want := tt.wantArgs
			if want == nil {
				want = []any{}
			}
			if !reflect.DeepEqual(f.calls[0].args, want) {
				t.Errorf("args\n got: %#v\nwant: %#v", f.calls[0].args, want)
			}
		})
	}
}

func TestRobnoRepositoryWithoutSession(t *testing.T) {
	d, f := newDrvDB(t, nil)
	if _, err := NewRobnoRepository(d).MagacinMag(context.Background(), 3); !errors.Is(err, ErrNoSession) {
		t.Fatalf("err %v, want ErrNoSession", err)
	}
	if len(f.calls) != 0 {
		t.Error("no statement may run without the period of the session")
	}
}

// robnoSifarnikCases are the lookups in the šifarnici of the robno module.
var robnoSifarnikCases = []robnoCase{
	{
		name:    "magacin mag",
		respond: drvRow([]string{"mag"}, int64(12)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			mag, err := r.MagacinMag(ctx, 3)
			if err == nil && mag != 12 {
				return fmt.Errorf("mag %d", mag)
			}
			return err
		},
		wantSQL:  "select mag from magacini WHERE (magaciniid = $1) AND (god = $2) AND (kar = $3)",
		wantArgs: []any{3, 2026, 7},
	},
	{
		name:    "konto zaliha",
		respond: drvRow([]string{"konto", "vkonta"}, "1320", int64(1)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			konto, vkonta, err := r.KontoZaliha(ctx, 44)
			if err == nil && (konto != "1320" || vkonta != 1) {
				return fmt.Errorf("got %q %d", konto, vkonta)
			}
			return err
		},
		wantSQL:  "select trim(konto), vkonta from fkpl WHERE (idfkpl = $1) AND (vkonta in (1, 2)) AND (god = $2) AND (kar = $3)",
		wantArgs: []any{44, 2026, 7},
	},
	{
		name:    "other konta of the magacin",
		respond: drvRow([]string{"count"}, int64(0)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.MagkontoDrugi(ctx, 3, 44, 9)
			return err
		},
		wantSQL:  "select count(*) from magkonto WHERE (magaciniid = $1) AND (idfkpl = $2) AND (magkontoid <> $3) AND (god = $4) AND (kar = $5)",
		wantArgs: []any{3, 44, int64(9), 2026, 7},
	},
	{
		name:    "konto of the kontni plan",
		respond: drvRow([]string{"count"}, int64(1)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.SintetickiKontoPostoji(ctx, "6040")
			return err
		},
		wantSQL:  "select count(*) from fkpl WHERE (vkonta = 2) AND (konto = $1) AND (god = $2) AND (kar = $3)",
		wantArgs: []any{"6040", 2026, 7},
	},
}

func TestRobnoSifarnikSQL(t *testing.T) {
	runRobnoCases(t, robnoSifarnikCases)
}

// krajParams are the parameters of the kraj poslovne godine of robnoKrajCases.
var krajParams = domain.KrajPoslovneGodineParams{Magacin: 3, OdKonta: "1320", OdSifre: 0, DoSifre: 99, ObradiNule: true, Tip: "naziv", Nalog: 5, Dokum: 6, Vrd: 89, NovaGod: 2027}

// robnoKrajCases are the statements of the kraj poslovne godine.
var robnoKrajCases = []robnoCase{
	{
		name:    "popis TIPZAL 1 sorted by naziv",
		respond: drvResult{cols: []string{"redbr"}},
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.PopisZaliha1(ctx, krajParams)
			return err
		},
		wantSQL: "SELECT row_number() over (ORDER BY coalesce(a.naziv,''),s.sifra)::int redbr,s.konto,s.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,coalesce(a.gru,0) grupa,coalesce(g.naziv,'') nazivgrupe, " +
			"CASE WHEN coalesce(m.nacvodzal,0)=3 THEN coalesce(max(s.prosnc),0) ELSE coalesce(max(c.cena),0) END cena, 0::numeric kolicina1,0::numeric kolicina2,0::numeric kolicina3,0::numeric ukupnakolicina, " +
			"coalesce(sum(s.ulaz-s.izlaz),0) stanjezaliha FROM rsta s LEFT JOIN rcene c ON c.sifra=s.sifra AND c.god=s.god AND c.kar=s.kar " +
			"LEFT JOIN rsif a ON a.sifra=s.sifra AND a.god=s.god AND a.kar=s.kar LEFT JOIN rgru g ON g.gru=a.gru AND g.god=s.god AND g.kar=s.kar " +
			"LEFT JOIN magacini m ON m.mag=s.mag AND m.god=s.god AND m.kar=s.kar WHERE (s.god = $1) AND (s.kar = $2) AND (($3 = 0 OR s.mag = $4)) " +
			"AND (($5 = '' OR s.konto >= $6)) AND (($7 = '' OR s.konto <= $8)) AND (($9 = 0 OR s.sifra >= $10)) AND (($11 = 0 OR s.sifra <= $12)) " +
			"GROUP BY s.konto, s.sifra, a.naziv, a.jm, a.gru, g.naziv, m.nacvodzal HAVING ($13 OR sum(s.ulaz-s.izlaz)<>0) ORDER BY coalesce(a.naziv,''),s.sifra",
		wantArgs: []any{2026, 7, 3, 3, "1320", "1320", "", "", 0, 0, 99, 99, true},
	},
	{
		name:    "popis TIPZAL 2",
		respond: drvResult{cols: []string{"redbr"}},
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.PopisZaliha2(ctx, krajParams)
			return err
		},
		wantSQL: "SELECT row_number() over (ORDER BY coalesce(a.naziv,''),d.sifra)::int redbr,d.konto,d.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,coalesce(a.gru,0) grupa,coalesce(g.naziv,'') nazivgrupe, " +
			"coalesce(d.otk,'') otk,coalesce(d.serija,'') serija,coalesce(to_char(date '1800-01-01' + nullif(d.roktr, 0)::int, 'DD.MM.YYYY'),'') rok,0::numeric kolicina1,0::numeric kolicina2,0::numeric kolicina3,0::numeric ukupnakolicina, " +
			"coalesce((SELECT sum(s.ulaz-s.izlaz) FROM rsta s WHERE s.god=$1 AND s.kar=$2 AND s.mag=d.mag AND s.sifra=d.sifra AND s.konto=d.konto),0) stanjezaliha " +
			"FROM drsta d LEFT JOIN rsif a ON a.sifra=d.sifra AND a.god=d.god AND a.kar=d.kar LEFT JOIN rgru g ON g.gru=a.gru AND g.god=d.god AND g.kar=d.kar " +
			"WHERE (d.god = $3) AND (d.kar = $4) AND (($5 = 0 OR d.mag = $6)) AND (($7 = '' OR d.konto >= $8)) AND (($9 = '' OR d.konto <= $10)) " +
			"AND (($11 = 0 OR d.sifra >= $12)) AND (($13 = 0 OR d.sifra <= $14)) ORDER BY coalesce(a.naziv,''),d.sifra",
		wantArgs: []any{2026, 7, 2026, 7, 3, 3, "1320", "1320", "", "", 0, 0, 99, 99},
	},
	{
		name: "popis sort outside the whitelist",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			q := krajParams
			q.Tip = "s.sifra; drop table rsta"
			if _, err := r.PopisZaliha1(ctx, q); err == nil {
				return errors.New("an unknown sort must be an error")
			}
			// one statement is expected by the harness: run a harmless one
			_, err := r.TableExists(ctx, nil, "rsta")
			return err
		},
		respond:  drvRow([]string{"exists"}, true),
		wantSQL:  "SELECT to_regclass($1) IS NOT NULL",
		wantArgs: []any{"rsta"},
	},
	{
		name:    "visak manjak: the popis of the new year in the same kar",
		respond: drvResult{cols: []string{"redbr"}},
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.VisakManjak(ctx, krajParams)
			return err
		},
		wantSQL: "WITH book AS (SELECT s.konto,s.sifra,sum(s.ulaz-s.izlaz) stanje,sum(s.dug-s.pot) iznos,max(s.cena) cena FROM rsta s WHERE s.god=$1 AND s.kar=$2 AND ($3=0 OR s.mag=$4) GROUP BY s.konto,s.sifra), " +
			"popis AS (SELECT p.konto,p.sifra,sum(p.kolic) kolicina,sum(p.iznos) iznos FROM rpro p JOIN rdok d ON d.rdokid=p.rdokid WHERE p.god=$5 AND p.kar=$6 AND p.nalog=$7 AND p.dokum=$8 AND p.vrd=$9 AND ($10=0 OR p.mag=$11) GROUP BY p.konto,p.sifra), " +
			"data_rows AS (SELECT b.konto,b.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,b.cena,b.iznos,b.stanje, coalesce(p.kolicina,0) kolicina,coalesce(p.iznos,0) popis_iznos,coalesce(p.kolicina,0)-b.stanje razlika " +
			"FROM book b LEFT JOIN popis p ON p.konto=b.konto AND p.sifra=b.sifra LEFT JOIN rsif a ON a.sifra=b.sifra AND a.god=$12 AND a.kar=$13) " +
			"SELECT row_number() over (ORDER BY sifra)::int redbr,sifra,konto,naziv,jm,cena,kolicina kolicinapopisa,popis_iznos iznospopisa, stanje stanjeknjigovodstveno,iznos iznosknjigovodstveno, " +
			"case when razlika>0 then razlika else 0 end visak, case when razlika>0 then (popis_iznos-iznos) else 0 end iznosviska, case when razlika<0 then -razlika else 0 end manjak, " +
			"case when razlika<0 then (iznos-popis_iznos) else 0 end iznosmanjka, case when razlika>0 and popis_iznos-iznos>0 then popis_iznos-iznos else 0 end finansijskivisak, " +
			"case when razlika<0 and iznos-popis_iznos>0 then iznos-popis_iznos else 0 end finansijskimanjak FROM data_rows WHERE (razlika<>0) ORDER BY sifra",
		wantArgs: []any{2026, 7, 3, 3, 2027, 7, 5, 6, 89, 3, 3, 2026, 7},
	},
	{
		name:    "table exists in a transaction",
		respond: drvRow([]string{"exists"}, true),
		inTx:    true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.TableExists(ctx, tx, "rpro")
			return err
		},
		wantSQL:  "SELECT to_regclass($1) IS NOT NULL",
		wantArgs: []any{"rpro"},
	},
}

func TestRobnoKrajSQL(t *testing.T) {
	runRobnoCases(t, robnoKrajCases)
}

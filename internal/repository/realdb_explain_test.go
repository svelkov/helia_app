package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"helia/config"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestRealDBExplain checks every statement of the golden tests against the database of
// config/config.json: it captures the SQL and the arguments of each case with the fake driver and
// runs EXPLAIN of it (the plan only, nothing is executed) in a read-only transaction that is
// always rolled back. The plain SELECTs are also run (read only). The calls of the database
// functions that copy or reverse a nalog are not sent. It runs only with HELIA_EXPLAIN_DB=1.
func TestRealDBExplain(t *testing.T) {
	if os.Getenv("HELIA_EXPLAIN_DB") != "1" {
		t.Skip("set HELIA_EXPLAIN_DB=1 to check the statements against the database of config/config.json (read only)")
	}
	data, err := os.ReadFile("../../config/config.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	c := cfg.DBConfig
	real, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSearchPath))
	if err != nil {
		t.Fatal(err)
	}
	defer real.Close()

	var stmts []drvCall
	names := map[int]string{}
	add := func(name string, calls []drvCall) {
		for _, call := range calls {
			names[len(stmts)] = name
			stmts = append(stmts, call)
		}
	}
	for file, cases := range map[string][]goldenCase{
		"izvestaji": robnoIzvestajiCases, "dokumenti": robnoDokumentiCases, "postoji": robnoPostojiCases, "avans": robnoAvansCases,
		"interni": robnoInterniCases, "prenosnica": robnoPrenosnicaCases, "usluge": robnoUslugeCases, "knjizenje": robnoKnjizenjeCases,
		"opsti": robnoOpstiCases, "unos": robnoUnosCases, "finansijsko": finansijskoCases, "valute": robnoValuteCases,
		"komercijalno": komercijalnoCases, "ponuda": komercijalnoPonudaCases, "porudzbenica": komercijalnoPorCases,
		"prijemnica": komercijalnoPrijCases, "profaktura": komercijalnoProCases, "nalog": komercijalnoNalogCases, "otpremnica": komercijalnoOtpCases,
		// The e-otpremnice need migrations/003 (their columns of edok and epro, the tables enot and eopma).
		"eotpremnica": komercijalnoEotpCases,
	} {
		for _, gc := range cases {
			if gc.name == "rsif create the šifra 0" {
				continue // its columns are a fixture: checked below with the columns of the real catalog
			}
			d, f := newDrvDB(t, gc.respond)
			var tx db.Transaction
			if gc.inTx {
				tx, _ = d.Beginx()
			}
			_ = gc.run(robnoCtx(), NewRobnoRepository(d), tx)
			if tx != nil {
				_ = tx.Rollback()
			}
			add(file+"/"+gc.name, f.calls)
		}
	}
	for _, rc := range append(append([]robnoCase{}, robnoSifarnikCases...), robnoKrajCases...) {
		resp := rc.respond
		d, f := newDrvDB(t, func(string, []any) drvResult { return resp })
		var tx db.Transaction
		if rc.inTx {
			tx, _ = d.Beginx()
		}
		_ = rc.run(robnoCtx(), NewRobnoRepository(d), tx)
		if tx != nil {
			_ = tx.Rollback()
		}
		add("robno/"+rc.name, f.calls)
	}
	for _, fc := range fvrListCases {
		resp := fc.respond
		d, f := newDrvDB(t, func(string, []any) drvResult { return resp })
		_, _ = fc.run(NewBaseRepository[domain.Fvr](d, "fvr"))
		add("fvr/"+fc.name, f.calls)
	}
	{
		// The artikal of the šifra 0 with the columns of rsif of the real catalog (one read-only
		// select); the insert it builds is only explained.
		kolone, err := NewRobnoRepository(db.NewPostgresDB(sqlx.NewDb(real, "postgres"))).rsifKolone(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		rows := make([][]any, len(kolone))
		for i, k := range kolone {
			rows[i] = []any{k.Key, k.C1, k.C2, k.C3, k.C4}
		}
		d, f := newDrvDB(t, func(q string, _ []any) drvResult {
			if strings.Contains(q, "information_schema") {
				return drvResult{cols: []string{"key", "c1", "c2", "c3", "c4"}, rows: rows}
			}
			return drvResult{cols: []string{"rsifid"}, rows: [][]any{{int64(1)}}}
		})
		if _, err := NewRobnoRepository(d).RsifKreirajNulu(robnoCtx(), nil, "Avans"); err != nil {
			t.Errorf("rsif of the šifra 0 with the real columns: %v", err)
		}
		add("avans/rsif create the šifra 0 (real columns)", f.calls)
	}
	{
		d, f := newDrvDB(t, func(string, []any) drvResult {
			return drvResult{cols: []string{"lockedby", "expiresat"}, affected: 1}
		})
		lr, ctx, exp := NewLockRepository(d), context.Background(), time.Now()
		tx, _ := lr.BeginTx()
		_, _ = lr.InsertTx(ctx, tx, "fnal", 7, "ana", exp)
		_, _, _, _ = lr.HolderTx(ctx, tx, "fnal", 7, true)
		_ = lr.DeleteTx(ctx, tx, "fnal", 7, true)
		_ = lr.RefreshTx(ctx, tx, "fnal", 7, "ana", exp)
		_ = tx.Rollback()
		_ = lr.Release(ctx, "fnal", 7, "ana")
		_ = lr.DeleteExpired(ctx)
		_, _ = lr.Active(ctx, "fnal", 7, exp)
		add("locks", f.calls)
	}

	explained, ran, skipped := 0, 0, 0
	for i, s := range stmts {
		q := strings.TrimSpace(s.query)
		low := strings.ToLower(q)
		if strings.HasPrefix(low, "select fn_") {
			skipped++
			continue
		}
		if err := explain(real, q, s.args); err != nil {
			t.Errorf("%s\n%s\n%v", names[i], squash(q), err)
			continue
		}
		explained++
		if (strings.HasPrefix(low, "select") || strings.HasPrefix(low, "with")) && !strings.Contains(q, "FOR UPDATE") {
			if err := runSelect(real, q, s.args); err != nil {
				t.Errorf("run %s\n%s\n%v", names[i], squash(q), err)
				continue
			}
			ran++
		}
	}
	t.Logf("%d statements: %d explained, %d plain selects run, %d database function calls not sent", len(stmts), explained, ran, skipped)
}

// readOnly runs f in a read-only transaction with a statement timeout and always rolls it back.
func readOnly(real *sql.DB, f func(tx *sql.Tx) error) error {
	ctx := context.Background()
	tx, err := real.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "set local statement_timeout = '10s'"); err != nil {
		return err
	}
	return f(tx)
}

func explain(real *sql.DB, q string, args []any) error {
	return readOnly(real, func(tx *sql.Tx) error {
		rows, err := tx.Query("EXPLAIN "+q, args...)
		if err != nil {
			return err
		}
		return rows.Close()
	})
}

func runSelect(real *sql.DB, q string, args []any) error {
	return readOnly(real, func(tx *sql.Tx) error {
		rows, err := tx.Query(q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
		}
		return rows.Err()
	})
}

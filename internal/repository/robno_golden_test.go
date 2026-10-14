package repository

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"helia/internal/infrastructure/db"
)

// The statements of the RobnoRepository are long, so their exact SQL and arguments are kept in
// golden files (testdata/*.golden): one block per case with the statements in the order they run.
// After a change, run `go test ./internal/repository -run Golden -update`, review the diff of the
// golden file against the intended SQL, and commit both.
var updateGolden = flag.Bool("update", false, "rewrite the golden files of the repository tests")

type goldenCase struct {
	name    string
	respond func(query string, args []any) drvResult // nil: no rows, nothing affected
	inTx    bool
	run     func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error
	wantErr error
}

func runGolden(t *testing.T, file string, cases []goldenCase) {
	t.Helper()
	var out strings.Builder
	for _, c := range cases {
		d, f := newDrvDB(t, c.respond)
		r := NewRobnoRepository(d)
		var tx db.Transaction
		if c.inTx {
			var err error
			if tx, err = d.Beginx(); err != nil {
				t.Fatal(err)
			}
		}
		err := c.run(robnoCtx(), r, tx)
		if tx != nil {
			_ = tx.Rollback()
		}
		if !errors.Is(err, c.wantErr) {
			t.Errorf("%s: err %v, want %v", c.name, err, c.wantErr)
		}
		fmt.Fprintf(&out, "### %s\n", c.name)
		if len(f.calls) == 0 {
			out.WriteString("(no statement)\n")
		}
		for _, call := range f.calls {
			fmt.Fprintf(&out, "%s\n%#v\nin transaction: %v\n", squash(call.query), call.args, call.inTx)
		}
		out.WriteString("\n")
	}
	path := filepath.Join("testdata", file)
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	got := out.String()
	w := strings.ReplaceAll(string(want), "\r\n", "\n")
	if got == w {
		return
	}
	gb, wb := strings.Split(got, "\n### "), strings.Split(w, "\n### ")
	for i := 0; i < len(gb) || i < len(wb); i++ {
		var g, x string
		if i < len(gb) {
			g = gb[i]
		}
		if i < len(wb) {
			x = wb[i]
		}
		if g != x {
			t.Errorf("%s: case %d differs\n got:\n%s\nwant:\n%s", file, i+1, g, x)
		}
	}
}

func TestRobnoIzvestajiGolden(t *testing.T) {
	runGolden(t, "robno_izvestaji.golden", robnoIzvestajiCases)
}

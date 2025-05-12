package domain

import "database/sql"

// Oamgrp represents the 'oamgrp' table in the 'baza' schema.
type Oamgrp struct {
	Oamgrpid   int            `db:"oamgrpid"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Agrupa     int            `db:"agrupa"`
	Naziv      string         `db:"naziv"`
	Sllist     int16          `db:"sllist"`
	Datsllist  sql.NullTime   `db:"datsllist"`
	Xdatunosa  sql.NullTime   `db:"xdatunosa"`
	Xdatizmene sql.NullTime   `db:"xdatizmene"`
	Xopunos    sql.NullString `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
}

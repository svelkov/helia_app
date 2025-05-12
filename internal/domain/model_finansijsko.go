package domain

import (
	"database/sql"
	"time"
)

// FKPL represents the "baza.fkpl" table.
type Fkpl struct {
	IDFKPL      int            `db:"idfkpl"`
	God         int            `db:"god"`
	Kar         int            `db:"kar"`
	Vkonta      int16          `db:"vkonta"`
	Konto       string         `db:"konto"`
	Sifra       string         `db:"sifra"`
	Naziv       string         `db:"naziv"`
	Devizni     bool           `db:"devizni"`
	Xdatunosa   time.Time      `db:"xdatunosa"`
	Xopunos     string         `db:"xopunos"`
	Xdatizmene  sql.NullTime   `db:"xdatizmene"`
	Xopizmene   sql.NullString `db:"xopizmene"`
	IDPartneri  *int           `db:"idpartneri"`
	Novikonto   string         `db:"novikonto"`
	Starikonto  string         `db:"starikonto"`
	Stariidpart *float64       `db:"stariidpart"`
	Kolicinski  bool           `db:"kolicinski"`
	Spcojid     *int           `db:"spcojid"`
	Stariidspc  *int           `db:"stariidspc"`
	NazivEng    string         `db:"naziv_eng"`
	Kursirati   bool           `db:"kursirati"`
}

// FNAL represents the "baza.fnal" table.
type Fnal struct {
	IDFnal     int            `db:"idfnal"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Nalog      int64          `db:"nalog"`
	Tipdok     string         `db:"tipdok"`
	Danal      time.Time      `db:"danal"`
	Opis       string         `db:"opis"`
	Dug        float64        `db:"dug"`
	Pot        float64        `db:"pot"`
	Rbr        int64          `db:"rbr"`
	Datob      time.Time      `db:"datob"`
	Oper       string         `db:"oper"`
	Brst       int            `db:"brst"`
	Abr        int            `db:"abr"`
	Nalsts     string         `db:"nalsts"`
	Xdatunosa  time.Time      `db:"xdatunosa"`
	Xdatizmene sql.NullTime   `db:"xdatizmene"`
	Xopunos    string         `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
	IDTipdok   *int           `db:"idtipdok"` // Nullable foreign key
}

// Sf represents the 'sf' table in the 'baza' schema.
type Sf struct {
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Brst       int            `db:"brst"`
	Brna       int            `db:"brna"`
	Dug        float64        `db:"dug"`
	Pot        float64        `db:"pot"`
	Rbr        int64          `db:"rbr"`
	Xdatunosa  time.Time      `db:"xdatunosa"`
	Xdatizmene sql.NullTime   `db:"xdatizmene"`
	Xopunos    string         `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
}
type UkupnaObrada struct {
	UkNaloga  int64
	UkStavki  int64
	Duguje    float64
	Potrazuje float64
}

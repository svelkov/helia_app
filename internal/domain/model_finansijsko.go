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
	Nalog      int64          `db:"nalog" schema:"nalog" `
	Tipdok     string         `db:"tipdok" schema:"tipdok" `
	Danal      time.Time      `db:"danal" schema:"danal"`
	Opis       string         `db:"opis" schema:"opis"`
	Dug        float64        `db:"dug"`
	Pot        float64        `db:"pot"`
	Rbr        int64          `db:"rbr"`
	Datob      time.Time      `db:"datob" schema:"datob"`
	Oper       string         `db:"oper"`
	Brst       int            `db:"brst"`
	Abr        int            `db:"abr"`
	Nalsts     string         `db:"nalsts"`
	Xdatunosa  sql.NullTime   `db:"xdatunosa"`
	Xdatizmene sql.NullTime   `db:"xdatizmene"`
	Xopunos    sql.NullString `db:"xopunos"`
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

type Fpro struct {
	IDFpro        int64         `db:"idfpro"`
	God           int           `db:"god"`
	Kar           int           `db:"kar"`
	Nalog         float64       `db:"nalog"`
	Tipdok        string        `db:"tipdok"`
	Rbr           int64         `db:"rbr"`
	Danal         *time.Time    `db:"danal"`
	Brst          int           `db:"brst"`
	Iznos         float64       `db:"iznos"`
	Kat           int16         `db:"kat"`
	Opis          string        `db:"opis"`
	Dadok         *time.Time    `db:"dadok"`
	Rok           int           `db:"rok"`
	Vrd           int           `db:"vrd"`
	Vkonta        int16         `db:"vkonta"`
	Konto         string        `db:"konto"`
	Sifra         string        `db:"sifra"`
	Tra           int           `db:"tra"`
	Deviznos      float64       `db:"deviznos"`
	Kurs          float64       `db:"kurs"`
	Sifval        int           `db:"sifval"`
	Mi            *int          `db:"mi"`
	Ostatak       float64       `db:"ostatak"`
	Dokum         string        `db:"dokum"`
	Vkrbr         int           `db:"vkrbr"`
	Vktip         string        `db:"vktip"`
	Xdatunosa     *sql.NullTime `db:"xdatunosa"`
	Xdatizmene    *sql.NullTime `db:"xdatizmene"`
	Xopunos       string        `db:"xopunos"`
	Xopizmene     string        `db:"xopizmene"`
	IDFnal        *int          `db:"idfnal"`
	IDFvknjrac    *int          `db:"idfvknjrac"`
	IDOrgjed      *int          `db:"idorgjed"`
	IDFkpl        *int          `db:"idfkpl"`
	Mag           int16         `db:"mag"`
	Komid         *int          `db:"komid"`
	Rdokid        *int          `db:"rdokid"`
	Mtroska       string        `db:"mtroska"`
	Mestotrid     *int          `db:"mestotrid"`
	Flegkomp      int16         `db:"flegkomp"`
	Dokumv        string        `db:"dokumv"`
	Dadokv        *time.Time    `db:"dadokv"`
	Travez        int           `db:"travez"`
	Starikonto    string        `db:"starikonto"`
	Ojozn         string        `db:"ojozn"`
	Sifkom        int           `db:"sifkom"`
	Kolic         float64       `db:"kolic"`
	Cena          float64       `db:"cena"`
	IDKir         *int          `db:"idkir"`
	IDKpr         int           `db:"idkpr"`
	Dug           float64       `db:"dug"`
	Pot           float64       `db:"pot"`
	NazivKonta    string        `db:"nazivkonta"`
	NazivPartnera string        `db:"nazivpartnera"`
}

type KopirajNalog struct {
	IDFnal       int64     `db:"idfnal"`
	TipdokOld    string    `db:"tipdok"`
	NalogOld     string    `db:"nalodold"`
	DanalOld     string    `db:"danalold"`
	DatKnjOld    string    `db:"datknjold"`
	OpisOld      string    `db:"opisold"`
	TipdokNew    string    `db:"tipdoknew"`
	NalogNew     string    `db:"nalognew"`
	DanalNew     time.Time `db:"danalnew"`
	DatknjNew    time.Time `db:"datknjnew"`
	OpisNew      string    `db:"opisnew"`
	TipdokValues []ComboItem
}

type UkupnaObrada struct {
	UkNaloga  string
	UkStavki  string
	Duguje    string
	Potrazuje string
}

// PrometHDR represents the structure for PROMETHDR records
type PrometTotalValues struct {
	DugDo    float64 `json:"dugDo" db:"dugdo"`
	PotDo    float64 `json:"potDo" db:"potdo"`
	SaldoDo  float64 `json:"saldoDo" db:"saldodo"`
	DugPer   float64 `json:"dugPer" db:"dugper"`
	PotPer   float64 `json:"potPer" db:"potper"`
	SaldoPer float64 `json:"saldoPer" db:"saldoper"`
	DugTot   float64 `json:"dugTot" db:"dugtot"`
	PotTot   float64 `json:"potTot" db:"pottot"`
	SaldoTot float64 `json:"saldoTot" db:"saldotot"`
}
type PrometResponse struct {
	TotalRecords int               `json:"totalRecords"`
	Data         []PrometDto       `json:"data"`
	Totals       PrometTotalValues `json:"totals"`
}
type PrometDto struct {
	God          int          `db:"god"`
	Kar          int          `db:"kar"`
	Tipdok       string       `db:"tipdok"`
	Nalog        string       `db:"nalog"`
	Danal        sql.NullTime `db:"danal"`
	Vrd          string       `db:"vrd"`
	Dokum        string       `db:"dokum"`
	Dadok        sql.NullTime `db:"dadok"`
	Rok          string       `db:"rok"`
	Tra          string       `db:"tra"`
	OJ           string       `db:"oj"`
	Opis         string       `db:"opis"`
	Duguje       float64      `db:"duguje"`
	Potrazuje    float64      `db:"potrazuje"`
	Saldo        float64      `db:"saldo"`
	Sifval       string       `db:"sifval"`
	Kurs         float64      `db:"kurs"`
	Deviznos     float64      `db:"deviznos"`
	Kolduguje    float64      `db:"kolduguje"`
	Kolpotrazuje float64      `db:"kolpotrazuje"`
	Cena         float64      `db:"cena"`
	Stanje       float64      `db:"stanje"`
	Dokumv       string       `db:"dokumv"`
	Dadokv       sql.NullTime `db:"dadokv"`
	Travez       string       `db:"travez"`
	Rdokid       int          `db:"rdokid"`
	MI           string       `db:"mi"`
	Devduguje    float64      `db:"devduguje"`
	Devpotrazuje float64      `db:"devpotrazuje"`
	Iznos        float64      `db:"iznos"`
	Kolic        float64      `db:"kolic"`
	Ojozn        string       `db:"ojozn"`
	Kat          string       `db:"kat"`
	Konto        string       `db:"konto"`
	Sifra        string       `db:"sifra"`
	Idfpro       int          `db:"idfpro"`
	Idfnal       int          `db:"idfnal"`
	Idfkpl       int          `db:"idfkpl"`
}

type FnalPayload struct {
	// IDFnal    int64  `schema:"idfnal"`
	Nalog  string `form:"nalog" binding:"required"`
	Danal  string `form:"danal" binding:"required"`
	Datob  string `form:"datob" binding:"required"`
	Tipdok string `form:"tipdok" binding:"required"`
	Opis   string `form:"opis"`
}

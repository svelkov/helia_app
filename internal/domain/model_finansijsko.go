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
	Xdatunosa   time.Time      `db:"xdatunosa" format:"datetime"`
	Xdatizmene  sql.NullTime   `db:"xdatizmene" format:"datetime"`
	Xopunos     string         `db:"xopunos"`
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
	IDFnal     int64          `db:"idfnal"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Nalog      int64          `db:"nalog" form:"nalog"`
	Tipdok     string         `db:"tipdok" form:"tipdok"`
	Danal      time.Time      `db:"danal" form:"danal" format:"date"`
	Opis       string         `db:"opis" form:"opis"`
	Dug        float64        `db:"dug"`
	Pot        float64        `db:"pot"`
	Rbr        int64          `db:"rbr"`
	Datob      time.Time      `db:"datob" form:"datob" format:"date"`
	Oper       string         `db:"oper"`
	Brst       int            `db:"brst"`
	Abr        int            `db:"abr"`
	Nalsts     string         `db:"nalsts"`
	Xdatunosa  sql.NullTime   `db:"xdatunosa" format:"datetime"`
	Xdatizmene sql.NullTime   `db:"xdatizmene" format:"datetime"`
	Xopunos    sql.NullString `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
	IDTipdok   int64          `db:"idtipdok"`
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
	Xdatunosa  time.Time      `db:"xdatunosa" format:"datetime"`
	Xdatizmene sql.NullTime   `db:"xdatizmene" format:"datetime"`
	Xopunos    string         `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
}

type Fpro struct {
	IDFpro        int64         `db:"idfpro" addupdate:"true"`
	God           int           `db:"god" addupdate:"true"`
	Kar           int           `db:"kar" addupdate:"true"`
	Nalog         int64         `db:"nalog" addupdate:"true"`
	Tipdok        string        `db:"tipdok" addupdate:"true"`
	Rbr           int64         `db:"rbr" addupdate:"true"`
	Danal         *time.Time    `db:"danal" addupdate:"true"`
	Brst          int           `db:"brst" addupdate:"true"`
	Iznos         float64       `db:"iznos" addupdate:"true"`
	Kat           int16         `db:"kat" addupdate:"true"`
	Opis          string        `db:"opis" addupdate:"true"`
	Dadok         sql.NullTime  `db:"dadok" addupdate:"true"`
	Rok           int           `db:"rok" addupdate:"true"`
	Vrd           int           `db:"vrd" addupdate:"true"`
	Vkonta        int16         `db:"vkonta" addupdate:"true"`
	Konto         string        `db:"konto" addupdate:"true"`
	Sifra         string        `db:"sifra" addupdate:"true"`
	Naziv         string        `db:"naziv" addupdate:"false"`
	Tra           int           `db:"tra" addupdate:"true"`
	Deviznos      float64       `db:"deviznos" addupdate:"true"`
	Kurs          float64       `db:"kurs" addupdate:"true"`
	Sifval        int           `db:"sifval" addupdate:"true"`
	Mi            sql.NullInt16 `db:"mi" addupdate:"true"`
	Ostatak       float64       `db:"ostatak" addupdate:"true"`
	Dokum         string        `db:"dokum" addupdate:"true"`
	Vkrbr         int           `db:"vkrbr" addupdate:"true"`
	Vktip         string        `db:"vktip" addupdate:"true"`
	Xdatunosa     sql.NullTime  `db:"xdatunosa" format:"datetime" addupdate:"true"`
	Xdatizmene    sql.NullTime  `db:"xdatizmene" format:"datetime" addupdate:"true"`
	Xopunos       string        `db:"xopunos" addupdate:"true"`
	Xopizmene     string        `db:"xopizmene" addupdate:"true"`
	IDFnal        int64         `db:"idfnal" addupdate:"true"`
	IDFvknjrac    sql.NullInt64 `db:"idfvknjrac" addupdate:"true"`
	IDOrgjed      sql.NullInt64 `db:"idorgjed" addupdate:"true"`
	IDFkpl        int64         `db:"idfkpl" addupdate:"true"`
	Mag           sql.NullInt16 `db:"mag" addupdate:"true"`
	Komid         sql.NullInt16 `db:"komid" addupdate:"true"`
	Rdokid        sql.NullInt64 `db:"rdokid" addupdate:"true"`
	Mtroska       string        `db:"mtroska" addupdate:"true"`
	Mestotrid     sql.NullInt16 `db:"mestotrid" addupdate:"true"`
	Flegkomp      int16         `db:"flegkomp" addupdate:"true"`
	Dokumv        string        `db:"dokumv" addupdate:"true"`
	Dadokv        sql.NullTime  `db:"dadokv" addupdate:"true"`
	Travez        int           `db:"travez" addupdate:"true"`
	Starikonto    string        `db:"starikonto" addupdate:"true"`
	Ojozn         string        `db:"ojozn" addupdate:"true"`
	Sifkom        sql.NullInt16 `db:"sifkom" addupdate:"true"`
	Kolic         float64       `db:"kolic" addupdate:"true"`
	Cena          float64       `db:"cena" addupdate:"true"`
	IDKir         sql.NullInt64 `db:"idkir" addupdate:"true"`
	IDKpr         sql.NullInt64 `db:"idkpr" addupdate:"true"`
	Dug           float64       `db:"dug" addupdate:"false"`
	Pot           float64       `db:"pot" addupdate:"false"`
	NazivKonta    string        `db:"nazivkonta" addupdate:"false"`
	NazivPartnera string        `db:"nazivpartnera" addupdate:"false"`
	Mesec         int           `db:"mesec" addupdate:"false"`
}

type KopirajNalog struct {
	IDFnal       int64  `db:"idfnal"`
	TipdokOld    string `db:"tipdok"`
	NalogOld     string `db:"nalodold"`
	DanalOld     string `db:"danalold"`
	DatKnjOld    string `db:"datknjold"`
	OpisOld      string `db:"opisold"`
	TipdokNew    string `db:"tipdoknew"`
	NalogNew     string `db:"nalognew"`
	DanalNew     string `db:"danalnew"`
	DatknjNew    string `db:"datknjnew"`
	OpisNew      string `db:"opisnew"`
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

type SaldaDto struct {
	Mesec           int     `db:"mesec"`
	MesecNaziv      string  `db:"mesecnaziv"`
	Duguje          float64 `db:"duguje"`
	Potrazuje       float64 `db:"potrazuje"`
	Saldo           float64 `db:"saldo"`
	SaldoKumul      float64 `db:"saldokumul"`
	PocStanjeDug    float64 `db:"pocstanjedugu"`
	PocStanjePot    float64 `db:"pocstanjepot"`
	PocStanjeSaldo  float64 `db:"pocstanjesaldo"`
	TekuciPromDug   float64 `db:"tekucpromdug"`
	TekuciPromPot   float64 `db:"tekucprompot"`
	TekuciPromSaldo float64 `db:"tekucpromsaldo"`
	UkPromDug       float64 `db:"ukpromdug"`
	UkPromPot       float64 `db:"ukprompot"`
	UkPromSaldo     float64 `db:"ukpromsaldo"`
}

type SaldaPartnerDto struct {
	IDPartneri     int64   `db:"idpartneri"`
	Konto          string  `db:"konto"`
	Sifra          string  `db:"sifra"`
	NazivKonta     string  `db:"kontonaziv"`
	Naziv          string  `db:"naziv"`
	PIB            string  `db:"pib"`
	Adresa         string  `db:"adresa"`
	PostanskiBroj  int     `db:"pobro"`
	Mesto          string  `db:"mesto"`
	Kupac          float64 `db:"kupac"`
	Dobavljac      float64 `db:"dobavljac"`
	PrimljenAvanas float64 `db:"primljenavans"`
	DatAvans       float64 `db:"datavans"`
	Saldo          float64 `db:"saldo"`
	Stanje         float64 `db:"stanje"`
	DugPot         string  `db:"dugpot"`
	JMBG           string  `db:"jmbg"`
	BPG            string  `db:"bpg"`
	BrIndex        string  `db:"brindex"`
	PstDug         float64 `db:"pstdug"`
	PstPot         float64 `db:"pstpot"`
	PstSaldo       float64 `db:"pstsaldo"`
	Dug            float64 `db:"dug"`
	Pot            float64 `db:"pot"`
}

type SaldaKomercijalistiDto struct {
	KomID      int16   `db:"komid"`
	Sifkom     int16   `db:"sifkom"`
	Imeprezime string  `db:"imeprezime"`
	Duguje     float64 `db:"dug"`
	Potrazuje  float64 `db:"pot"`
	Saldo      float64 `db:"saldo"`
	Dospelo    float64 `db:"dospelo"`
	Nedospelo  float64 `db:"nedospelo"`
	Vanperioda float64 `db:"vanperioda"`
}

// SaldaHeader represents header/summary data for Salda po kontima in partneri view
type SaldaHeaderDto struct {
	Konto     string  `db:"konto" json:"konto"`
	Duguje    float64 `db:"duguje" json:"duguje"`
	Potrazuje float64 `db:"potrazuje" json:"potrazuje"`
	Saldo     float64 `db:"saldo" json:"saldo"`
}

// SaldaDetail represents transaction details for Salda detalji
type SaldaDetailDto struct {
	FU     string  `db:"f_u" json:"f_u"`
	Nalog  string  `db:"nalog" json:"nalog"`
	Danal  string  `db:"danal" json:"danal"`
	TipDok string  `db:"tipdok" json:"tipdok"`
	BrDok  string  `db:"brdok" json:"brdok"`
	GodDok int     `db:"goddok" json:"goddok"`
	Iznos  float64 `db:"iznos" json:"iznos"`
}

type SaldaResponse struct {
	TotalRecords int        `json:"totalRecords"`
	Data         []SaldaDto `json:"data"`
	Totals       SaldaDto   `json:"totals"`
}

type FnalPayload struct {
	// IDFnal    int64  `schema:"idfnal"`
	Nalog     string `form:"nalog"`
	Danal     string `form:"danal"`
	Datob     string `form:"datob"`
	Tipdok    string `form:"tipdok"`
	Opis      string `form:"opis"`
	Duguje    string `form:"duguje"`
	Potrazuje string `form:"potrazuje"`
	Saldo     string `form:"saldo"`
}
type FproPayload struct {
	IDFnal         int64       `schema:"idfnal"`
	Tipdok         string      `form:"tipdok"`
	Nalog          string      `form:"nalog"`
	Danal          string      `form:"danal"`
	Datob          string      `form:"datob"`
	Opis           string      `form:"opis"`
	Brst           string      `form:"brst"`
	Konto          string      `form:"konto"`
	Sifra          string      `form:"sifra"`
	Vrd            string      `form:"vrd"`
	Opisknj        string      `form:"opisknj"`
	Mag            string      `form:"mag"`
	Mi             string      `form:"mi"`
	Komercijalista string      `form:"komercijalista"`
	Iznos          string      `form:"iznos"`
	Kat            string      `form:"kat"`
	Orgjed         []ComboItem `form:"orgjed"`
	Mtroska        []ComboItem `form:"mtroska"`
	Dokum          string      `form:"dokum"`
	Dadok          string      `form:"dadok"`
	GodDok         string      `form:"god_dok"`
	Rok            string      `form:"rok"`
	Tra            string      `form:"tra"`
	Dokvezni       string      `form:"dokvezni"`
	Dadokv         string      `form:"dadokv"`
	Godvezni       string      `form:"godvezni"`
	Travez         string      `form:"travez"`
	Napomena       string      `form:"napomena"`
	Vkonta         string      `form:"vkonta"`
	Sifval         string      `form:"sifval"`
	Kurs           string      `form:"kurs"`
	Deviznos       string      `form:"deviznos"`
	Duguje         string
	Potrazuje      string
	Saldo          string
}

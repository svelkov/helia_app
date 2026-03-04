package domain

import (
	"database/sql"
	"time"
)

// FKPL represents the "baza.fkpl" table.
type Fkpl struct {
	IDFkpl      int            `db:"idfkpl"`
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
	Oper       sql.NullString `db:"oper"`
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
	DocType       string        `db:"doctype" addupdate:"false"`
	// Age bucket fields for PregledPotrazivanjaObaveze
	Realizacija        float64 `db:"realizacija" addupdate:"false"`
	Placeno            float64 `db:"placeno" addupdate:"false"`
	DospelaRealizacija float64 `db:"dospelarealizacija" addupdate:"false"`
	Dospece15          float64 `db:"dospece15" addupdate:"false"`
	Dospece30          float64 `db:"dospece30" addupdate:"false"`
	Dospece60          float64 `db:"dospece60" addupdate:"false"`
	Dospece90          float64 `db:"dospece90" addupdate:"false"`
	Dospece120         float64 `db:"dospece120" addupdate:"false"`
	Dospece120Plus     float64 `db:"dospece120plus" addupdate:"false"`
}
type FproDto struct {
	Fpro
	SubsinNaziv  string  `db:"subsin_naziv"`
	SinNaziv     string  `db:"sin_naziv"`
	PocStanjeDug float64 `db:"pocstanjedug"`
	PocStanjePot float64 `db:"pocstanjepot"`
	PrometDug    float64 `db:"prometdug"`
	PrometPot    float64 `db:"prometpot"`
}

type KopirajNalog struct {
	IDFnal       int64  `db:"idfnal"`
	TipdokOld    string `db:"tipdok"`
	NalogOld     string `db:"nalogold"`
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
type TotalValues struct {
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
	TotalRecords int         `json:"totalRecords"`
	Data         []PrometDto `json:"data"`
	Totals       TotalValues `json:"totals"`
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
	Naziv        string       `db:"naziv"`
	Vkonta       string       `db:"vkonta"`
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

type KompenzacijeDto struct {
	God                   int       `db:"god" json:"god"`
	Kar                   int       `db:"kar" json:"kar"`
	KompBr                int64     `db:"kompbr" json:"kompbr"`
	Dokum                 string    `db:"dokum" json:"dokum"`
	Dadok                 time.Time `db:"dadok" json:"dadok" format:"date"`
	DatKomp               time.Time `db:"datkomp" json:"datkomp" format:"date"`
	Iznos                 float64   `db:"iznos" json:"iznos"`
	Status                string    `db:"status" json:"status"`
	Partner               string    `db:"partner" json:"partner"`
	KontoDuznika          string    `db:"konto_duznika" json:"konto_duznika"`
	SifraDuznika          string    `db:"sifra_duznika" json:"sifra_duznika"`
	KontoPoverioca        string    `db:"konto_poverioca" json:"konto_poverioca"`
	SifraPoverioca        string    `db:"sifra_poverioca" json:"sifra_poverioca"`
	Naziv                 string    `db:"naziv" json:"naziv"`
	IznosDokumDuznika     float64   `db:"iznos_dokum_duznika" json:"iznos_dokum_duznika"`
	IznosDokumPoverioca   float64   `db:"iznos_dokum_poverioca" json:"iznos_dokum_poverioca"`
	KompenzacijeDuznik    float64   `db:"kompenzacije_duznik" json:"kompenzacije_duznik"`
	KompenzacijePoverilac float64   `db:"kompenzacije_poverilac" json:"kompenzacije_poverilac"`
	Mesto                 string    `db:"mesto" json:"mesto"`
	Adresa                string    `db:"adresa" json:"adresa"`
	Stsdok                int       `db:"stsdok" json:"stsdok"`
	Odglicep              string    `db:"odglicep" json:"odglicep"`
	Odgliced              string    `db:"odgliced" json:"odgliced"`
	Kompenzhid            int64     `db:"kompenzhid" json:"kompenzacijhid"`
}

// DnevnikDto represents dnevnik knjizenja record with calculated fields
type DnevnikDto struct {
	Rbr       int          `db:"rbr" json:"rbr"`
	Danal     sql.NullTime `db:"danal" json:"danal" format:"date"`
	Tipdok    string       `db:"tipdok" json:"tipdok"`
	Nalog     int          `db:"nalog" json:"nalog"`
	Konto     string       `db:"konto" json:"konto"`
	Sifra     string       `db:"sifra" json:"sifra"`
	Naziv     string       `db:"naziv" json:"naziv"`
	Duguje    float64      `db:"duguje" json:"duguje"`
	Potrazuje float64      `db:"potrazuje" json:"potrazuje"`
	Saldo     float64      `db:"saldo" json:"saldo"`
	Opis      string       `db:"opis" json:"opis"`
	Dokum     string       `db:"dokum" json:"dokum"`
	Dadok     sql.NullTime `db:"dadok" json:"dadok" format:"date"`
	Ojozn     string       `db:"ojozn" json:"ojozn"`
	Sifval    string       `db:"sifval" json:"sifval"`
	Devdug    float64      `db:"devdug" json:"devdug"`
	Devpot    float64      `db:"devpot" json:"devpot"`
}

// Fizvzag represents the "baza.fizvzag" table (bank statement header).
type Fizvzag struct {
	IDFizvzag  int            `db:"idfizvzag"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Konto      string         `db:"konto"`
	Sifra      string         `db:"sifra"`
	Izvbr      int            `db:"izvbr"`
	Datizv     sql.NullTime   `db:"datizv" format:"date"`
	Prstanje   float64        `db:"prstanje"`
	Ukdug      float64        `db:"ukdug"`
	Ukpot      float64        `db:"ukpot"`
	Nstanje    float64        `db:"nstanje"`
	Ukbrst     int            `db:"ukbrst"`
	Nalog      int64          `db:"nalog"`
	Tipdok     string         `db:"tipdok"`
	Izvsts     string         `db:"izvsts"`
	Brrac      string         `db:"brrac"`
	Xdatunosa  sql.NullTime   `db:"xdatunosa" format:"datetime"`
	Xdatizmene sql.NullTime   `db:"xdatizmene" format:"datetime"`
	Xopunos    sql.NullString `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
	IDbanke    sql.NullInt64  `db:"idbanke"`
	Banka      sql.NullString `db:"banka"`
}

// Fizvdet represents the "baza.fizvdet" table (bank statement details).
type Fizvdet struct {
	IDFizvdet  int            `db:"idfizvdet"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Brrac      string         `db:"brrac"`
	Izvbr      int            `db:"izvbr"`
	Datizv     sql.NullTime   `db:"datizv" format:"date"`
	Rbr        int64          `db:"rbr"`
	Konto      string         `db:"konto"`
	Sifra      string         `db:"sifra"`
	Iznos      float64        `db:"iznos"`
	Kat        int16          `db:"kat"`
	IDOrgjed   int            `db:"idorgjed"`
	Vrd        int            `db:"vrd"`
	Konto1     string         `db:"konto1"`
	Sifra1     string         `db:"sifra1"`
	Nsedprim   string         `db:"nsedprim"`
	Brracup    string         `db:"brracup"`
	Osnplac    string         `db:"osnplac"`
	Sdozn      string         `db:"sdozn"`
	Sdozn1     string         `db:"sdozn1"`
	Duguje     float64        `db:"duguje"`
	Potrazuje  float64        `db:"potrazuje"`
	Modelzad   string         `db:"modelzad"`
	Pnabrzad   string         `db:"pnabrzad"`
	Mododob    string         `db:"mododob"`
	Pnabrodob  string         `db:"pnabrodob"`
	Prekl      string         `db:"prekl"`
	Xdatunosa  sql.NullTime   `db:"xdatunosa" format:"datetime"`
	Xdatizmene sql.NullTime   `db:"xdatizmene" format:"datetime"`
	Xopunos    sql.NullString `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
	IDFizvzag  sql.NullInt64  `db:"idfizvzag"`
}

// Kir represents the "baza.kir" table (issued invoices register).
type Kir struct {
	IDKir      int            `db:"idkir" gorm:"primaryKey"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Vktip      string         `db:"vktip"`
	Vkrbr      int            `db:"vkrbr" form:"vkrbr"`
	Krbr       int64          `db:"krbr" form:"krbr"`
	IDPartneri int            `db:"idpartneri"`
	Dknjiz     time.Time      `db:"dknjiz" format:"date" form:"datknjiz"`
	Danal      time.Time      `db:"danal" format:"date" form:"danal"`
	Dizd       time.Time      `db:"dizd" format:"date" form:"datizd"`
	Kracun     string         `db:"kracun" form:"kracun"`
	IznsaPDV   float64        `db:"iznsapdv" form:"iznosapdv"`
	OslobCL24  float64        `db:"oslobcl24" form:"oslobcl24"`
	OslobCL25  float64        `db:"oslobcl25" form:"oslobcl25"`
	IzvozSaPr  float64        `db:"izvozsapr" form:"izvozsapr"`
	IzvozBezPr float64        `db:"izvozbezpr" form:"izvozbezpr"`
	Osn1       float64        `db:"osn1" form:"osn1"`
	PDV1       float64        `db:"pdv1" form:"pdv1"`
	Osn2       float64        `db:"osn2" form:"osn2"`
	PDV2       float64        `db:"pdv2" form:"pdv2"`
	Prom1      float64        `db:"prom1" form:"prom1"`
	Prom2      float64        `db:"prom2" form:"prom2"`
	Nalog      int64          `db:"nalog" form:"nalog"`
	TipDok     string         `db:"tipdok" form:"tipdok"`
	Vrd        int            `db:"vrd" `
	Dokum      string         `db:"dokum" form:"dokum"`
	Vpr        int            `db:"vpr"`
	Rkar       int            `db:"rkar"`
	Brst       int            `db:"brst"`
	Konto      string         `db:"konto" form:"konto"`
	Sifra      string         `db:"sifra" form:"sifra"`
	PIB        string         `db:"pib" form:"pib"`
	Naziv      string         `db:"naziv" form:"naziv"`
	Xdatunosa  sql.NullTime   `db:"xdatunosa"`
	Xdatizmene sql.NullTime   `db:"xdatizmene"`
	Xopunos    sql.NullString `db:"xopunos"`
	Xopizmene  sql.NullString `db:"xopizmene"`
	IDFVknjrac *int           `db:"idfvknjrac"`
	IDFpro     *int64         `db:"idfpro"`
	Numdok     int            `db:"numdok"`
	Datprometa sql.NullTime   `db:"datprometa" format:"date"`
}
type KirPayload struct {
	Kir
	NazivPartnera string `db:"naziv_pa"`
	Mesto         string `db:"mesto_pa"`
	Pib           string `db:"pib_pa"`
	Adresa        string `db:"adresa_pa"`
}
type KprPayload struct {
	Kpr
	NazivPartnera string `db:"naziv_pa"`
	Mesto         string `db:"mesto_pa"`
	Pib           string `db:"pib_pa"`
	Adresa        string `db:"adresa_pa"`
}

// Kpr represents the "baza.kpr" table (tax invoices register - received invoices).
type Kpr struct {
	IDKpr        int            `db:"idkpr" gorm:"primaryKey"`
	God          int            `db:"god"`
	Kar          int            `db:"kar"`
	VKTip        string         `db:"vktip"`
	VKRbr        int            `db:"vkrbr"`
	DRbr         int            `db:"drbr"`
	DKnjiz       sql.NullTime   `db:"dknjiz" format:"date"`
	DUvoz        sql.NullTime   `db:"duvoz" format:"date"`
	DIzd         sql.NullTime   `db:"dizd" format:"date"`
	IznsAPDV     float64        `db:"iznsapdv"`
	IznosLob     float64        `db:"iznoslob"`
	NisuObvPDV   float64        `db:"nisuobvpdv"`
	UvozBezPDV   float64        `db:"uvozbezpdv"`
	PrethodPDV   float64        `db:"prethpdv"`
	PretPDV1     float64        `db:"pretpdv1"`
	PretPDV2     float64        `db:"pretpdv2"`
	UvozPDV      float64        `db:"uvozpdv"`
	PoljVred     float64        `db:"poljvred"`
	PoljPDV      float64        `db:"poljpdv"`
	Vrd          int            `db:"vrd"`
	Konto        string         `db:"konto"`
	Sifra        string         `db:"sifra"`
	Ter          int            `db:"ter"`
	UvozOsnPDV   float64        `db:"uvozosnpdv"`
	Vpr          int            `db:"vpr"`
	OsnBezPDV    float64        `db:"osnbezpdv"`
	Brst         int            `db:"brst"`
	RKar         int            `db:"rkar"`
	Nalog        int64          `db:"nalog"`
	Dokum        string         `db:"dokum"`
	PIB          string         `db:"pib"`
	Naziv        string         `db:"naziv"`
	IDPartneri   int            `db:"idpartneri"`
	DAnal        sql.NullTime   `db:"danal" format:"date"`
	TipDok       string         `db:"tipdok"`
	XDatUnosa    sql.NullTime   `db:"xdatunosa" format:"datetime"`
	XDatIzmene   sql.NullTime   `db:"xdatizmene" format:"datetime"`
	XOpUnos      sql.NullString `db:"xopunos"`
	XOpIzmene    sql.NullString `db:"xopizmene"`
	IDFPro       int64          `db:"idfpro"`
	IDFVknJrac   *int           `db:"idfvknjrac"`
	OsnBezPod    float64        `db:"osnbezpod"`
	DatPrometa   sql.NullTime   `db:"datprometa" format:"date"`
	OsnovicaVT   float64        `db:"osnovicavt"`
	OsnovicaNT   float64        `db:"osnovicant"`
	PrethodPDVVT float64        `db:"prethpdvvt"`
	PrethodPDVNT float64        `db:"prethpdvnt"`
	TKonto       string         `db:"tkonto"`
	TSifra       string         `db:"tsifra"`
}

// PoreskaPrijavaData holds all tax form field values (EDT_001 through EDT_110)
type PoreskaPrijavaData struct {
	Edt001 float64 `json:"edt_001"`
	Edt002 float64 `json:"edt_002"`
	Edt003 float64 `json:"edt_003"`
	Edt004 float64 `json:"edt_004"`
	Edt005 float64 `json:"edt_005"`
	Edt006 float64 `json:"edt_006"`
	Edt007 float64 `json:"edt_007"`
	Edt008 float64 `json:"edt_008"`
	Edt009 float64 `json:"edt_009"`
	Edt103 float64 `json:"edt_103"`
	Edt104 float64 `json:"edt_104"`
	Edt105 float64 `json:"edt_105"`
	Edt106 float64 `json:"edt_106"`
	Edt107 float64 `json:"edt_107"`
	Edt108 float64 `json:"edt_108"`
	Edt109 float64 `json:"edt_109"`
	Edt110 float64 `json:"edt_110"`
}

// EppSekcija represents a section in the EPP (Electronic Public Procurement) system
type EppSekcija struct {
	IDFepp        int            `db:"idfepp" gorm:"primaryKey"`
	Nivo          int            `db:"nivo"`
	Sekcija       string         `db:"sekcija"`
	Izvor         string         `db:"izvor"`
	Naziv         string         `db:"naziv"`
	Akt1          bool           `db:"akt1"`
	Akt2          bool           `db:"akt2"`
	Akt3          bool           `db:"akt3"`
	Akt4          bool           `db:"akt4"`
	KprPoc        string         `db:"kprpoc"`
	KprDodatnaPDV string         `db:"krpdodatnap"`
	KprPDV        string         `db:"kprpdv"`
	XDatUnosa     sql.NullTime   `db:"xdatunosa" format:"datetime"`
	XDatIzmene    sql.NullTime   `db:"xdatizmene" format:"datetime"`
	XOpUnos       sql.NullString `db:"xopunos"`
	XOpIzmene     sql.NullString `db:"xopizmene"`
}

// EppEvidencija represents an EPP evidence record
type EppEvidencija struct {
	FsepID     int            `db:"fsepid" gorm:"primaryKey"`
	God        int            `db:"god"`
	Kar        int            `db:"kar"`
	Polje      string         `db:"polje"`
	Opis       string         `db:"opis"`
	Osn1       float64        `db:"osn1"`
	Pdv1       float64        `db:"pdv1"`
	Osn2       float64        `db:"osn2"`
	Pdv2       float64        `db:"pdv2"`
	Oddat      time.Time      `db:"oddat" format:"date"`
	Dodat      time.Time      `db:"dodat" format:"date"`
	Nipo       string         `db:"nipo"`
	XDatUnosa  sql.NullTime   `db:"xdatunosa" format:"datetime"`
	XDatIzmene sql.NullTime   `db:"xdatizmene" format:"datetime"`
	XOpUnos    sql.NullString `db:"xopunos"`
	XOpIzmene  sql.NullString `db:"xopizmene"`
}

// EppSefKpr represents EPP SEF KPR (received invoice) data
type EppSefKpr struct {
	IDFeppSef      int            `db:"idfeppsef" gorm:"primaryKey"`
	God            int            `db:"god"`
	Kar            int            `db:"kar"`
	RedBroj        int            `db:"redbroj"`
	DokumentTip    string         `db:"dokumenttip"`
	BrDokumenta    string         `db:"brdokumenta"`
	DatumDokumenta time.Time      `db:"datumdokumenta" format:"date"`
	DatumLicnog    time.Time      `db:"datumlicnog" format:"date"`
	Iznos          float64        `db:"iznos"`
	PDV            float64        `db:"pdv"`
	Konto          string         `db:"konto"`
	Status         string         `db:"status"`
	XDatUnosa      sql.NullTime   `db:"xdatunosa" format:"datetime"`
	XDatIzmene     sql.NullTime   `db:"xdatizmene" format:"datetime"`
	XOpUnos        sql.NullString `db:"xopunos"`
	XOpIzmene      sql.NullString `db:"xopizmene"`
}

// Bils represents the "baza.bils" table (Balance Sheets).
type Bils struct {
	BilsID     int            `db:"bilsid" addupdate:"true"`
	God        int            `db:"god" addupdate:"true"`
	Kar        int            `db:"kar" addupdate:"true"`
	AOP        int            `db:"aop" form:"aop" addupdate:"true"`
	NazP       string         `db:"nazp" form:"nazp" addupdate:"true"`
	Konta      string         `db:"konta" form:"konta" addupdate:"true"`
	Rbr        int64          `db:"rbr" form:"rbr" addupdate:"true"`
	TGod       float64        `db:"tgod" addupdate:"true"`
	PGod       float64        `db:"pgod" addupdate:"true"`
	Grac       string         `db:"grac" form:"grac" addupdate:"true"`
	NiPo       int16          `db:"nipo" form:"nipo" addupdate:"true"`
	TGodH      int64          `db:"tgodh" addupdate:"true"`
	PGodH      int64          `db:"pgodh" addupdate:"true"`
	Pozic1     int16          `db:"pozic_1" form:"pozic_1" addupdate:"true"`
	Pozic2     int16          `db:"pozic_2" form:"pozic_2" addupdate:"true"`
	Pozic3     int16          `db:"pozic_3" form:"pozic_3" addupdate:"true"`
	Pozic4     int16          `db:"pozic_4" form:"pozic_4" addupdate:"true"`
	Pozic5     int16          `db:"pozic_5" form:"pozic_5" addupdate:"true"`
	Pozic6     int16          `db:"pozic_6" form:"pozic_6" addupdate:"true"`
	Pozic7     int16          `db:"pozic_7" form:"pozic_7" addupdate:"true"`
	Pozic8     int16          `db:"pozic_8" form:"pozic_8" addupdate:"true"`
	Pozic9     int16          `db:"pozic_9" form:"pozic_9" addupdate:"true"`
	Pozic10    int16          `db:"pozic_10" form:"pozic_10" addupdate:"true"`
	Pozic11    int16          `db:"pozic_11" form:"pozic_11" addupdate:"true"`
	Pozic12    int16          `db:"pozic_12" form:"pozic_12" addupdate:"true"`
	Odm        int16          `db:"odm" addupdate:"true"`
	Dom        int16          `db:"dom" addupdate:"true"`
	XDatUnosa  sql.NullTime   `db:"xdatunosa" addupdate:"true" format:"datetime"`
	XDatIzmene sql.NullTime   `db:"xdatizmene" addupdate:"true" format:"datetime"`
	XOpUnos    sql.NullString `db:"xopunos" addupdate:"true"`
	XOpIzmene  sql.NullString `db:"xopizmene" addupdate:"true"`
	PGodPS     float64        `db:"pgodps" addupdate:"true"`
	PGodHPS    int64          `db:"pgodhps" addupdate:"true"`
	TGodPS     float64        `db:"tgodps" addupdate:"true"`
	TGodHPS    int64          `db:"tgodhps" addupdate:"true"`
	Napomena   string         `db:"napomena" addupdate:"true"`
	Skraceni   int16          `db:"skraceni" form:"skraceni" addupdate:"true"`
	Naziv      string         `db:"naziv" addupdate:"false"`
	Vkonta     string         `db:"vkonta" addupdate:"false"`
}

// Bilu represents the "baza.bilu" table (Income Statements).
type Bilu struct {
	BiluID     int            `db:"biluid" addupdate:"true" `
	God        int            `db:"god" addupdate:"true"`
	Kar        int            `db:"kar" addupdate:"true"`
	AOP        int            `db:"aop" form:"aop" addupdate:"true"`
	NazP       string         `db:"nazp" form:"nazp" addupdate:"true"`
	Konta      string         `db:"konta" form:"konta" addupdate:"true"`
	Rbr        int64          `db:"rbr" form:"rbr" addupdate:"true"`
	TGod       float64        `db:"tgod" addupdate:"true"`
	PGod       float64        `db:"pgod" addupdate:"true"`
	Grac       string         `db:"grac" addupdate:"true"`
	NiPo       int16          `db:"nipo" form:"nipo" addupdate:"true"`
	Pozic1     int16          `db:"pozic_1" form:"pozic_1" addupdate:"true"`
	Pozic2     int16          `db:"pozic_2" form:"pozic_2" addupdate:"true"`
	Pozic3     int16          `db:"pozic_3" form:"pozic_3" addupdate:"true"`
	Pozic4     int16          `db:"pozic_4" form:"pozic_4" addupdate:"true"`
	Pozic5     int16          `db:"pozic_5" form:"pozic_5" addupdate:"true"`
	Pozic6     int16          `db:"pozic_6" form:"pozic_6" addupdate:"true"`
	Pozic7     int16          `db:"pozic_7" form:"pozic_7" addupdate:"true"`
	Pozic8     int16          `db:"pozic_8" form:"pozic_8" addupdate:"true"`
	Pozic9     int16          `db:"pozic_9" form:"pozic_9" addupdate:"true"`
	Pozic10    int16          `db:"pozic_10" form:"pozic_10" addupdate:"true"`
	Pozic11    int16          `db:"pozic_11" form:"pozic_11" addupdate:"true"`
	Pozic12    int16          `db:"pozic_12" form:"pozic_12" addupdate:"true"`
	TGodH      int            `db:"tgodh" addupdate:"true"`
	PGodH      int            `db:"pgodh" addupdate:"true"`
	Odm        int16          `db:"odm" addupdate:"true"`
	Dom        int16          `db:"dom" addupdate:"true"`
	XDatUnosa  sql.NullTime   `db:"xdatunosa" addupdate:"true" format:"datetime"`
	XDatIzmene sql.NullTime   `db:"xdatizmene" addupdate:"true" format:"datetime"`
	XOpUnos    sql.NullString `db:"xopunos" addupdate:"true"`
	XOpIzmene  sql.NullString `db:"xopizmene" addupdate:"true"`
	Napomena   string         `db:"napomena" addupdate:"true"`
	Skraceni   int16          `db:"skraceni" form:"skraceni" addupdate:"true"`
	Naziv      string         `db:"naziv" addupdate:"false"`
	Vkonta     string         `db:"vkonta" addupdate:"false"`
}

type BilansiTotals struct {
	TekGod     string
	TekGodH    string
	PrethGod   string
	PrethGodH  string
	PocStanje  string
	PocStanjeH string
}

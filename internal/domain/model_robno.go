package domain

import (
	"database/sql"
	"time"
)

// TIPKNPISMA Model (tip knjižnog pisma)
type Tipknpisma struct {
	Tipknjid   int64          `json:"tipknjid" db:"tipknjid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Sifrazlog  int16          `json:"sifrazlog" db:"sifrazlog" form:"sifrazlog"`
	Opis       string         `json:"opis" db:"opis" form:"opis"`
	XOpUnos    sql.NullString `json:"xop_unos" db:"xopunos"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
}

// MAGACINI Model
type Magacini struct {
	MagaciniID int          `json:"magaciniid" db:"magaciniid"`
	Mag        int          `json:"mag" db:"mag"`
	Opis       string       `json:"opis" db:"opis"`
	Tipmag     string       `json:"tipmag" db:"tipmag"`
	Adresa     string       `json:"adresa" db:"adresa"`
	Pobro      int          `json:"pobro" db:"pobro"`
	Mesto      string       `json:"mesto" db:"mesto"`
	Nadmag     int          `json:"nadmag" db:"nadmag"`
	Magosoba   string       `json:"magosoba" db:"magosoba"`
	God        int          `json:"god" db:"god"`
	Kar        int          `json:"kar" db:"kar"`
	Tel        string       `json:"tel" db:"tel"`
	Fax        string       `json:"fax" db:"fax"`
	Tipzal     int          `json:"tipzal" db:"tipzal"`
	Tipcene    int          `json:"tipcene" db:"tipcene"`
	Nacvodzal  int          `json:"nacvodzal" db:"nacvodzal"`
	Analiza    int          `json:"analiza" db:"analiza"`
	Email      string       `json:"email" db:"email"`
	Tipart     string       `json:"tipart" db:"tipart"`
	XDatUnosa  sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene  string       `json:"xop_izmene" db:"xopizmene"`
}

// FISP Model
type Fisp struct {
	IDPartneri   int          `json:"idpartneri" db:"idpartneri"`
	FispID       int          `json:"fispid" db:"fispid"`
	God          int          `json:"god" db:"god"`
	Kar          int          `json:"kar" db:"kar"`
	Konto        string       `json:"konto" db:"konto"`
	Sifra        string       `json:"sifra" db:"sifra"`
	Vk           int          `json:"vk" db:"vk"`
	Mag          int          `json:"mag" db:"mag"`
	Naziv        string       `json:"naziv" db:"naziv"`
	Adresa       string       `json:"adresa" db:"adresa"`
	Mesto        string       `json:"mesto" db:"mesto"`
	Pobro        int          `json:"pobro" db:"pobro"`
	XDatUnosa    sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XOpUnos      string       `json:"xop_unos" db:"xopunos"`
	XDatIzmene   sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpIzmene    string       `json:"xop_izmene" db:"xopizmene"`
	PIB          string       `json:"pib" db:"pib"`
	JIB          string       `json:"jib" db:"jib"`
	KontaktOsb   string       `json:"kontaktosb" db:"kontaktosb"`
	Ziro         string       `json:"ziro" db:"ziro"`
	TipPDV       int          `json:"tippdv" db:"tippdv"`
	MI           int64        `json:"mi" db:"mi"`
	Ter          int          `json:"ter" db:"ter"`
	Flg          string       `json:"flg" db:"flg"`
	GLN          int64        `json:"gln" db:"gln"`
	Email        string       `json:"email" db:"email"`
	PartnerNaziv string       `json:"partnernaziv" db:"partnernaziv"`
}

// ARTIKLI Model
type Artikli struct {
	ArtikliID    int          `json:"artikliid" db:"artikliid"`
	Sifra        string       `json:"sifra" db:"sifra"`
	Naziv        string       `json:"naziv" db:"naziv"`
	Tarifa       int          `json:"tarifa" db:"tarifa"`
	Tip          string       `json:"tip" db:"tip"`
	Filter       string       `json:"filter" db:"filter"`
	God          int          `json:"god" db:"god"`
	Kar          int          `json:"kar" db:"kar"`
	XDatUnosa    sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XOpUnos      string       `json:"xop_unos" db:"xopunos"`
	XDatIzmene   sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpIzmene    string       `json:"xop_izmene" db:"xopizmene"`
	PorezID      int          `json:"porezid" db:"porezid"`
	JedmereID    int          `json:"jedmereid" db:"jedmereid"`
	RobneGrupeID int          `json:"robnegrupeid" db:"robnegrupeid"`
	Barkod       string       `json:"barkod" db:"barkod"`
	ZalMinus     bool         `json:"zalminus" db:"zalminus"`
	FiskalnoIme  string       `json:"fiskalnoime" db:"fiskalnoime"`
}

// JEDMERE Model
type Jedmere struct {
	JedmereID      int            `json:"jedmereid" db:"jedmereid"`
	God            int            `json:"god" db:"god"`
	Kar            int            `json:"kar" db:"kar"`
	JM             string         `json:"jm" db:"jm" form:"jm"`
	XOpUnos        sql.NullString `json:"xop_unos" db:"xopunos"`
	XOpIzmene      sql.NullString `json:"xop_izmene" db:"xopizmene"`
	XDatUnosa      sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene     sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	Opis           string         `json:"opis" db:"opis" form:"opis"`
	BrDecimala     int            `json:"brdecimala" db:"brdecimala" form:"brdecimala"`
	ImaDuzinu      bool           `json:"ima_duzinu" db:"ima_duzinu" form:"ima_duzinu"`
	ImaSirinu      bool           `json:"ima_sirinu" db:"ima_sirinu" form:"ima_sirinu"`
	ImaKomade      bool           `json:"ima_komade" db:"ima_komade" form:"ima_komade"`
	KoristeSpecTez bool           `json:"koristi_spectez" db:"koristi_spectez" form:"koristi_spectez"`
}

// RGRU Model (Robne Grupe)
type Rgru struct {
	RgruID     int          `json:"rgruid" db:"rgruid"`
	God        int          `json:"god" db:"god"`
	Kar        int          `json:"kar" db:"kar"`
	Gru        int          `json:"gru" db:"gru" form:"gru"`
	Naziv      string       `json:"naziv" db:"naziv" form:"naziv"`
	XOpUnos    string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene  string       `json:"xop_izmene" db:"xopizmene"`
	XDatUnosa  sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
}

// RPGRU Model (Pod-Grupe Robnih Grupa)
type Rpgru struct {
	RpgruID    int          `json:"rpgruid" db:"rpgruid"`
	God        int          `json:"god" db:"god"`
	Kar        int          `json:"kar" db:"kar"`
	Gru        int          `json:"gru" db:"gru" form:"gru"`
	Pgru       int          `json:"pgru" db:"pgru" form:"pgru"`
	Naziv      string       `json:"naziv" db:"naziv" form:"naziv"`
	XOpUnos    string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene  string       `json:"xop_izmene" db:"xopizmene"`
	XDatIzmene sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XDatUnosa  sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	RgruID     int          `json:"rgruid" db:"rgruid"`
}

// RPOR Model
type Rpor struct {
	Rporid     int            `json:"rporid" db:"rporid"`
	PP         float64        `json:"pp" db:"pp" form:"pp"`
	PT         string         `json:"pt" db:"pt" form:"pt"`
	Datum      time.Time      `json:"datum" db:"datum" form:"datum"`
	Tip        int            `json:"tip" db:"tip" form:"tip"`
	XOpUnos    sql.NullString `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	Po         int            `json:"po" db:"po" form:"po"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	Slovo      string         `json:"slovo" db:"slovo" form:"slovo"`
}

// RPRO Model (Robni promet stavke / Robne promene)
type Rpro struct {
	RproID     int            `json:"rproid" db:"rproid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Nalog      int            `json:"nalog" db:"nalog"`
	Vrd        int            `json:"vrd" db:"vrd"`
	Dokum      int            `json:"dokum" db:"dokum"`
	Dadok      sql.NullTime   `json:"dadok" db:"dadok"`
	Iznos      float64        `json:"iznos" db:"iznos"`
	Kolic      float64        `json:"kolic" db:"kolic"`
	Brst       int            `json:"brst" db:"brst"`
	Cena       float64        `json:"cena" db:"cena"`
	Pcena      float64        `json:"pcena" db:"pcena"`
	Fcena      float64        `json:"fcena" db:"fcena"`
	Mcena      float64        `json:"mcena" db:"mcena"`
	Mcenap     float64        `json:"mcenap" db:"mcenap"`
	Ncena      float64        `json:"ncena" db:"ncena"`
	Rbr        int            `json:"rbr" db:"rbr"`
	Naz1       string         `json:"naz1" db:"naz1"`
	JM         string         `json:"jm" db:"jm"`
	Konto      string         `json:"konto" db:"konto"`
	Po         int16          `json:"po" db:"po"`
	Sifra      int            `json:"sifra" db:"sifra"`
	Fkto       string         `json:"fkto" db:"fkto"`
	Fana       string         `json:"fana" db:"fana"`
	Rab        float32        `json:"rab" db:"rab"`
	Mag        int16          `json:"mag" db:"mag"`
	Pkto       string         `json:"pkto" db:"pkto"`
	Pana       string         `json:"pana" db:"pana"`
	Ztro       float64        `json:"ztro" db:"ztro"`
	Ztrin      float64        `json:"ztrin" db:"ztrin"`
	Tz         string         `json:"tz" db:"tz"`
	Tzi        string         `json:"tzi" db:"tzi"`
	Vma        float32        `json:"vma" db:"vma"`
	Mma        float32        `json:"mma" db:"mma"`
	Vra        float64        `json:"vra" db:"vra"`
	Mra        float64        `json:"mra" db:"mra"`
	Model      string         `json:"model" db:"model"`
	Iakc       float64        `json:"iakc" db:"iakc"`
	Pakc       float32        `json:"pakc" db:"pakc"`
	Itaksa     float64        `json:"itaksa" db:"itaksa"`
	Ptaksa     float32        `json:"ptaksa" db:"ptaksa"`
	Dani       int16          `json:"dani" db:"dani"`
	Ztrof      float64        `json:"ztrof" db:"ztrof"`
	RdokID     int            `json:"rdokid" db:"rdokid"`
	Ststatus   string         `json:"ststatus" db:"ststatus"`
	RnalID     int            `json:"rnalid" db:"rnalid"`
	MagaciniID int            `json:"magaciniid" db:"magaciniid"`
	XDatUnosa  time.Time      `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	Otk        string         `json:"otk" db:"otk"`
	Serija     string         `json:"serija" db:"serija"`
	Roktr      int            `json:"roktr" db:"roktr"`
	RsifID     int            `json:"rsifid" db:"rsifid"`
	Cenaval    float64        `json:"cenaval" db:"cenaval"`
	RproID1    int            `json:"rproid1" db:"rproid1"`
	Kolic1     float64        `json:"kolic1" db:"kolic1"`
	Pdvpct     float32        `json:"pdvpct" db:"pdvpct"`
	Vpcena     float64        `json:"vpcena" db:"vpcena"`
	PinarID    int            `json:"pinarid" db:"pinarid"`
	Konprosnc  float64        `json:"konprosnc" db:"konprosnc"`
	Cenaold    float64        `json:"cenaold" db:"cenaold"`
	Taxcat     string         `json:"taxcat" db:"taxcat"`
	Taxexreco  string         `json:"taxexreco" db:"taxexreco"`
	Sifartkup  string         `json:"sifartkup" db:"sifartkup"`
}

// MAGKONTO Model
type Magkonto struct {
	MagaciniID  int          `json:"magaciniid" db:"magaciniid"`
	Mag         int          `json:"mag" db:"mag"`
	God         int          `json:"god" db:"god"`
	Kar         int          `json:"kar" db:"kar"`
	XDatUnosa   sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene  sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos     string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene   string       `json:"xop_izmene" db:"xopizmene"`
	Konto       string       `json:"konto" db:"konto" form:"konto"`
	IDFkpl      int          `json:"idfkpl" db:"idfkpl"`
	VKonta      int          `json:"vkonta" db:"vkonta" form:"vkonta"`
	KontoPrih   string       `json:"kontoprih" db:"kontoprih" form:"kontoprih"`
	KontoTroska string       `json:"kontotroska" db:"kontotroska" form:"kontotroska"`
	KontoRuc    string       `json:"kontoruc" db:"kontoruc" form:"kontoruc"`
	KontoRab    string       `json:"kontorab" db:"kontorab" form:"kontorab"`
}

// KOMERCIJALISTI Model
type Komercijalisti struct {
	KomID        int          `json:"komid" db:"komid"`
	God          int          `json:"god" db:"god"`
	Kar          int          `json:"kar" db:"kar"`
	Sifkom       int          `json:"sifkom" db:"sifkom" form:"sifkom"`
	SifNadred    int          `json:"sifnadred" db:"sifnadred" form:"sifnadred"`
	ImePrezime   string       `json:"imeprezime" db:"imeprezime" form:"imeprezime"`
	Adresa       string       `json:"adresa" db:"adresa" form:"adresa"`
	Mesto        string       `json:"mesto" db:"mesto" form:"mesto"`
	TelPosao     string       `json:"telposao" db:"telposao" form:"telposao"`
	TelMob       string       `json:"telmob" db:"telmob" form:"telmob"`
	TotProd      float64      `json:"totprod" db:"totprod"`
	TotProfit    float64      `json:"totprofit" db:"totprofit"`
	ZadDatProd   time.Time    `json:"zaddatprod" db:"zaddatprod"`
	TotNaplaceno float64      `json:"totnaplaceno" db:"totnaplaceno"`
	LoginName    string       `json:"loginname" db:"loginname" form:"loginname"`
	XDatUnosa    sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene   sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos      string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene    string       `json:"xop_izmene" db:"xopizmene"`
}

// RSIF Model (Artikli sa detaljima - Articles with details)
type Rsif struct {
	RsifID       int          `json:"rsifid" db:"rsifid"`
	God          int          `json:"god" db:"god"`
	Kar          int          `json:"kar" db:"kar"`
	Konto        string       `json:"konto" db:"konto"`
	Barkod       string       `json:"barkod" db:"barkod"`
	Sifra        int          `json:"sifra" db:"sifra"`
	Pro          string       `json:"pro" db:"pro"`
	Naziv        string       `json:"naziv" db:"naziv"`
	Po           int16        `json:"po" db:"po"`
	JM           string       `json:"jm" db:"jm"`
	Pa           int          `json:"pa" db:"pa"`
	Gru          int          `json:"gru" db:"gru"`
	Model        string       `json:"model" db:"model"`
	Miza         float64      `json:"miza" db:"miza"`
	Maza         float64      `json:"maza" db:"maza"`
	Kdob         string       `json:"kdob" db:"kdob"`
	Pakov        float32      `json:"pakov" db:"pakov"`
	Dani         int16        `json:"dani" db:"dani"`
	Serbr        string       `json:"serbr" db:"serbr"`
	XDatUnosa    sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene   sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos      string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene    string       `json:"xop_izmene" db:"xopizmene"`
	SifraDobaV   string       `json:"sifradobav" db:"sifradobav"`
	Pgru         int          `json:"pgru" db:"pgru"`
	RgruID       int          `json:"rgruid" db:"rgruid"`
	RpgruID      int          `json:"rpgruid" db:"rpgruid"`
	Tip          string       `json:"tip" db:"tip"`
	ArtRoba      bool         `json:"artroba" db:"artroba"`
	ArtProizv    bool         `json:"artproizv" db:"artproizv"`
	ArtSir       bool         `json:"artsir" db:"artsir"`
	ArtUsl       bool         `json:"artusl" db:"artusl"`
	ArtAmb       bool         `json:"artamb" db:"artamb"`
	ArtPolup     bool         `json:"artpolup" db:"artpolup"`
	TezJM        float64      `json:"tezjm" db:"tezjm"`
	Kvalitet     string       `json:"kvalitet" db:"kvalitet"`
	TrPakov      int          `json:"trpakov" db:"trpakov"`
	Kubikaza     float64      `json:"kubikaza" db:"kubikaza"`
	Koeftez      float64      `json:"koeftez" db:"koeftez"`
	PerRekontr   int16        `json:"perrekontr" db:"perrekontr"`
	ProizSifra   string       `json:"proizsifra" db:"proizsifra"`
	KomercOpis   string       `json:"komercopis" db:"komercopis"`
	DodatnaNaziv string       `json:"dodatnaziv" db:"dodatnaziv"`
	FSimE        string       `json:"fsime" db:"fsime"`
	ZemljaProizv string       `json:"zemljaproizv" db:"zemljaproizv"`
	TarifnaOzn   int64        `json:"tarifnaozn" db:"tarifnaozn"`
	BojeID       int          `json:"bojeid" db:"bojeid"`
	DebljinEID   int          `json:"debljineid" db:"debljineid"`
	ObliciID     int          `json:"obliciid" db:"obliciid"`
	MetaliID     int          `json:"metaliid" db:"metaliid"`
	ModMetalaID  int          `json:"modmetalaid" db:"modmetalaid"`
	Duzina       float64      `json:"duzina" db:"duzina"`
	Sirina       float64      `json:"sirina" db:"sirina"`
	RazVsir      float32      `json:"razvsir" db:"razvsir"`
	AmbSif       int          `json:"ambsif" db:"ambsif"`
	Lokacija     string       `json:"lokacija" db:"lokacija"`
	RokTra       string       `json:"roktra" db:"roktra"`
	ZemljaUvoza  string       `json:"zemljauvoza" db:"zemljauvoza"`
	Uvoznik      string       `json:"uvoznik" db:"uvoznik"`
	GodUvoza     int          `json:"goduvoza" db:"goduvoza"`
	GodProiz     int          `json:"godproiz" db:"godproiz"`
	AkciznaKat   string       `json:"akciznakat" db:"akciznakat"`
	KontoNaziv   string       `json:"kontonaziv" db:"kontonaziv"`
}

type RobnoStanjeDto struct {
	Magacin      int     `json:"magacin" db:"magacin"`
	Konto        string  `json:"konto" db:"konto"`
	SifraArtikla string  `json:"sifra_artikla" db:"sifra_artikla"`
	NazivArtikla string  `json:"naziv_artikla" db:"naziv_artikla"`
	Cena         float64 `json:"cena" db:"cena"`
	ReportTip    string  `json:"report_tip" db:"reporttip"`
}

type RobnoStanjaParams struct {
	Magacin      int     `json:"magacin" db:"magacin"`
	Konto        string  `json:"konto" db:"konto"`
	SifraArtikla string  `json:"sifra_artikla" db:"sifra_artikla"`
	NazivArtikla string  `json:"naziv_artikla" db:"naziv_artikla"`
	Cena         float64 `json:"cena" db:"cena"`
	ReportTip    string  `json:"report_tip" db:"reporttip"`
}
type RobnoKarticaParams struct {
	Magacin       int
	Konto         string
	Nalozi        string
	OdDanal       string
	DoDanal       string
	OdIznosa      float64
	DoIznosa      float64
	Cena          float64
	CbxBrojNaloga bool
	CbxDatum      bool
	CbxIznos      bool
	ReportTip     string
	SearchText    string
}
type RobnoKomPodaciParams struct {
	OdArtikla    string
	DoArtikla    string
	OdGrupe      string
	DoGrupe      string
	OdSifreKupca string
	DoSifreKupca string
	OdDatuma     string
	DoDatuma     string
	TipIzvestaja string // 1 = po artiklima, 2 = po kupcima
	Trziste      string // 1 = domaće, 2 = izvoz, 3 = ukupno
	ReportTip    string
	SearchText   string
}

// TrzisteName returns the human-readable name for the Trziste code.
func (p RobnoKomPodaciParams) TrzisteName() string {
	switch p.Trziste {
	case "1":
		return "DOMAĆE"
	case "2":
		return "IZVOZ"
	case "3":
		return "UKUPNO"
	default:
		return "-"
	}
}

type RobnoKomPodaciDto struct {
	Sifra         string  `json:"sifra" db:"sifra"`
	Naziv         string  `json:"naziv" db:"naziv"`
	Jm            string  `json:"jm" db:"jm"`
	Grupa         string  `json:"grupa" db:"gru"`
	NazivGrupe    string  `json:"naziv_grupe" db:"nazivgrupe"`
	Kolic         float64 `json:"kolic" db:"kolic"`
	XIznos        float64 `json:"xiznos" db:"xiznos"`
	XRab          float64 `json:"xrab" db:"xrab"`
	XUgrabat      float64 `json:"xugrabat" db:"xugrabat"`
	XKasa         float64 `json:"xkasa" db:"xkasa"`
	NabIznos      float64 `json:"nabiznos" db:"nabiznos"`
	TotKolic      float64 `json:"totkolic" db:"totkolic"`
	TotalNetvalue float64 `json:"total_netvalue" db:"total_netvalue"`
	Kupac         string  `json:"kupac" db:"kupac"`
	Mi            int     `json:"mi" db:"mi"`
	NazivKupca    string  `json:"naziv_kupca" db:"nazivkupca"`
	NazivMesta    string  `json:"naziv_mesta" db:"nazivmesta"`
}

type KrajPopisType1Row struct {
	RedBr          int     `json:"redbr" db:"redbr"`
	Konto          string  `json:"konto" db:"konto"`
	Sifra          int     `json:"sifra" db:"sifra"`
	Naziv          string  `json:"naziv" db:"naziv"`
	JM             string  `json:"jm" db:"jm"`
	Grupa          int     `json:"grupa" db:"grupa"`
	NazivGrupe     string  `json:"naziv_grupe" db:"nazivgrupe"`
	Cena           float64 `json:"cena" db:"cena"`
	Kolicina1      float64 `json:"kolicina1" db:"kolicina1"`
	Kolicina2      float64 `json:"kolicina2" db:"kolicina2"`
	Kolicina3      float64 `json:"kolicina3" db:"kolicina3"`
	UkupnaKolicina float64 `json:"ukupna_kolicina" db:"ukupnakolicina"`
	StanjeZaliha   float64 `json:"stanje_zaliha" db:"stanjezaliha"`
}

type KrajPopisType2Row struct {
	RedBr          int     `json:"redbr" db:"redbr"`
	Konto          string  `json:"konto" db:"konto"`
	Sifra          int     `json:"sifra" db:"sifra"`
	Naziv          string  `json:"naziv" db:"naziv"`
	JM             string  `json:"jm" db:"jm"`
	Grupa          int     `json:"grupa" db:"grupa"`
	NazivGrupe     string  `json:"naziv_grupe" db:"nazivgrupe"`
	OTK            string  `json:"otk" db:"otk"`
	Serija         string  `json:"serija" db:"serija"`
	Rok            string  `json:"rok" db:"rok"`
	Kolicina1      float64 `json:"kolicina1" db:"kolicina1"`
	Kolicina2      float64 `json:"kolicina2" db:"kolicina2"`
	Kolicina3      float64 `json:"kolicina3" db:"kolicina3"`
	UkupnaKolicina float64 `json:"ukupna_kolicina" db:"ukupnakolicina"`
	StanjeZaliha   float64 `json:"stanje_zaliha" db:"stanjezaliha"`
}

type KrajPopisData struct {
	TipZal int                 `json:"tipzal"`
	Type1  []KrajPopisType1Row `json:"type1"`
	Type2  []KrajPopisType2Row `json:"type2"`
}

// The tab-2 report has sixteen source/calculation columns.
type KrajVisakManjakRow struct {
	RedBr                 int     `json:"redbr" db:"redbr"`
	Konto                 string  `json:"konto" db:"konto"`
	Sifra                 int     `json:"sifra" db:"sifra"`
	Naziv                 string  `json:"naziv" db:"naziv"`
	JM                    string  `json:"jm" db:"jm"`
	Cena                  float64 `json:"cena" db:"cena"`
	KolicinaPopisa        float64 `json:"kolicina_popisa" db:"kolicinapopisa"`
	IznosPopisa           float64 `json:"iznos_popisa" db:"iznospopisa"`
	StanjeKnjigovodstveno float64 `json:"stanje_knjigovodstveno" db:"stanjeknjigovodstveno"`
	IznosKnjigovodstveno  float64 `json:"iznos_knjigovodstveno" db:"iznosknjigovodstveno"`
	Visak                 float64 `json:"visak" db:"visak"`
	IznosViska            float64 `json:"iznos_viska" db:"iznosviska"`
	Manjak                float64 `json:"manjak" db:"manjak"`
	IznosManjka           float64 `json:"iznos_manjka" db:"iznosmanjka"`
	FinansijskiVisak      float64 `json:"finansijski_visak" db:"finansijskivisak"`
	FinansijskiManjak     float64 `json:"finansijski_manjak" db:"finansijskimanjak"`
}

type KrajPoslovneGodineParams struct {
	Magacin     int    `json:"magacin"`
	OdKonta     string `json:"odkonta"`
	DoKonta     string `json:"dokonta"`
	OdSifre     int    `json:"odsifre"`
	DoSifre     int    `json:"dosifre"`
	Cena        int    `json:"cena"` // 1=CENA, 2=PROSNC, 3=NCENA, 4=VPCENA
	ObradiNule  bool   `json:"obradinule"`
	Tip         string `json:"tip"`
	TipZal      int    `json:"tipzal"`
	Nalog       int    `json:"nalog"`
	Dokum       int    `json:"dokum"`
	Vrd         int    `json:"vrd"`
	NovaGod     int    `json:"novagod"`
	StaraGod    int    `json:"staragod"`
	OrgJed      int    `json:"orgjed"`
	MestoTroska int    `json:"mestotroska"`
	DanObrade   string `json:"danobrade"`
	Opis        string `json:"opis"`
	DanNalog    string `json:"dannalog"`
	DanDokum    string `json:"dandokum"`
}

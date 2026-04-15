package domain

import "database/sql"

// MAGACINI Model
type Magacini struct {
	MagaciniID int            `json:"magaciniid" db:"magaciniid"`
	Mag        int            `json:"mag" db:"mag"`
	Opis       string         `json:"opis" db:"opis"`
	Tipmag     string         `json:"tipmag" db:"tipmag"`
	Adresa     string         `json:"adresa" db:"adresa"`
	Pobro      int            `json:"pobro" db:"pobro"`
	Mesto      string         `json:"mesto" db:"mesto"`
	Nadmag     int            `json:"nadmag" db:"nadmag"`
	Magosoba   string         `json:"magosoba" db:"magosoba"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Tel        string         `json:"tel" db:"tel"`
	Fax        string         `json:"fax" db:"fax"`
	Tipzal     int            `json:"tipzal" db:"tipzal"`
	Tipcene    int            `json:"tipcene" db:"tipcene"`
	Nacvodzal  int            `json:"nacvodzal" db:"nacvodzal"`
	Analiza    int            `json:"analiza" db:"analiza"`
	Email      string         `json:"email" db:"email"`
	Tipart     string         `json:"tipart" db:"tipart"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
}

// FISP Model
type Fisp struct {
	IDPartneri int            `json:"idpartneri" db:"idpartneri"`
	FispID     int            `json:"fispid" db:"fispid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Konto      string         `json:"konto" db:"konto"`
	Sifra      string         `json:"sifra" db:"sifra"`
	Vk         int            `json:"vk" db:"vk"`
	Mag        int            `json:"mag" db:"mag"`
	Naziv      string         `json:"naziv" db:"naziv"`
	Adresa     string         `json:"adresa" db:"adresa"`
	Mesto      string         `json:"mesto" db:"mesto"`
	Pobro      int            `json:"pobro" db:"pobro"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
	PIB        string         `json:"pib" db:"pib"`
	JIB        string         `json:"jib" db:"jib"`
	KontaktOsb string         `json:"kontaktosb" db:"kontaktosb"`
	Ziro       string         `json:"ziro" db:"ziro"`
	TipPDV     int            `json:"tippdv" db:"tippdv"`
	MI         int64          `json:"mi" db:"mi"`
	Ter        int            `json:"ter" db:"ter"`
	Flg        string         `json:"flg" db:"flg"`
	GLN        int64          `json:"gln" db:"gln"`
	Email      string         `json:"email" db:"email"`
}

// Jedmere Model
type Jedmere struct {
	JedmereID       int            `json:"jedmereid" db:"jedmereid"`
	God             int            `json:"god" db:"god"`
	Kar             int            `json:"kar" db:"kar"`
	JM              string         `json:"jm" db:"jm"`
	XOpUnos         string         `json:"xopunos" db:"xopunos"`
	XOpIzmene       sql.NullString `json:"xopizmene" db:"xopizmene"`
	XDatUnosa       sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene      sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	Opis            string         `json:"opis" db:"opis"`
	BrDecimala      int            `json:"brdecimala" db:"brdecimala"`
	Ima_Duzinu      bool           `json:"ima_duzinu" db:"ima_duzinu"`
	Ima_Sirinu      bool           `json:"ima_sirinu" db:"ima_sirinu"`
	Ima_Komade      bool           `json:"ima_komade" db:"ima_komade"`
	Koristi_SpecTez bool           `json:"koristi_spectez" db:"koristi_spectez"`
}

// Kepu Model
type Kepu struct {
	KepuID     int            `json:"kepuid" db:"kepuid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	MagaciniID int            `json:"magaciniid" db:"magaciniid"`
	DatKnj     sql.NullTime   `json:"datknj" db:"datknj"`
	Opis       string         `json:"opis" db:"opis"`
	DocDte     sql.NullTime   `json:"docdte" db:"docdte"`
	RdokID     int            `json:"rdokid" db:"rdokid"`
	Zaduz      float64        `json:"zaduz" db:"zaduz"`
	Razduz     float64        `json:"razduz" db:"razduz"`
	UplRac     float64        `json:"uplrac" db:"uplrac"`
	Napomena   string         `json:"napomena" db:"napomena"`
	RnalID     int            `json:"rnalid" db:"rnalid"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	IDPartneri int            `json:"idpartneri" db:"idpartneri"`
	IDTipDok   int            `json:"idtipdok" db:"idtipdok"`
	Dokum      string         `json:"dokum" db:"dokum"`
	Konto      string         `json:"konto" db:"konto"`
	Sifra      string         `json:"sifra" db:"sifra"`
	DokVrstaID int            `json:"dokvrstaid" db:"dokvrstaid"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
	IDFPro     int64          `json:"idfpro" db:"idfpro"`
	IDKir      sql.NullInt64  `json:"idkir" db:"idkir"`
	Kes        float64        `json:"kes" db:"kes"`
	Cek        float64        `json:"cek" db:"cek"`
	Kartica    float64        `json:"kartica" db:"kartica"`
	Racuni     float64        `json:"racuni" db:"racuni"`
	Pocek      float64        `json:"pocek" db:"pocek"`
}

// Magkonto Model
type Magkonto struct {
	MagaciniID int `json:"magaciniid" db:"magaciniid"`
	Mag        int `json:"mag" db:"mag"`
	God        int `json:"god" db:"god"`
	Kar        int `json:"kar" db:"kar"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	Konto  string `json:"konto" db:"konto"`
	IDFkpl int    `json:"idfkpl" db:"idfkpl"`
	VKonta int    `json:"vkonta" db:"vkonta"`

	KontoPrih   string `json:"kontoprih" db:"kontoprih"`
	KontoTroska string `json:"kontotroska" db:"kontotroska"`
	KontoRuc    string `json:"kontoruc" db:"kontoruc"`
	KontoRab    string `json:"kontorab" db:"kontorab"`
}

// Maguser Model
type Maguser struct {
	MaguserID  int            `json:"maguserid" db:"maguserid"`
	Mag        int            `json:"mag" db:"mag"`
	IDUser     string         `json:"iduser" db:"iduser"`
	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
}

// Porkat Model
type Porkat struct {
	PorkatID     int          `json:"porkatid" db:"porkatid"`
	ReasonID     int64        `json:"reasonid" db:"reasonid"`
	KeyKat       string       `json:"keykat" db:"keykat"`
	Law          string       `json:"law" db:"law"`
	Article      string       `json:"article" db:"article"`
	Paragraph    string       `json:"paragraph" db:"paragraph"`
	Point        string       `json:"point" db:"point"`
	Subpoint     string       `json:"subpoint" db:"subpoint"`
	Text         string       `json:"text" db:"text"`
	FreeformNote string       `json:"freeformnote" db:"freeformnote"`
	ActiveFrom   sql.NullTime `json:"activefrom" db:"activefrom"`
	ActiveTo     sql.NullTime `json:"activeto" db:"activeto"`
	Category     string       `json:"category" db:"category"`
	Obveznik     bool         `json:"obveznik" db:"obveznik"`
}

// Rcene Model
type Rcene struct {
	God   int `json:"god" db:"god"`
	Kar   int `json:"kar" db:"kar"`
	Sifra int `json:"sifra" db:"sifra"`

	Vma  float32 `json:"vma" db:"vma"`
	Mma  float32 `json:"mma" db:"mma"`
	Vra  float64 `json:"vra" db:"vra"`
	Mra  float64 `json:"mra" db:"mra"`
	Ztro float64 `json:"ztro" db:"ztro"`

	Model string `json:"model" db:"model"`

	Cena   float64 `json:"cena" db:"cena"`
	PCena  float64 `json:"pcena" db:"pcena"`
	FCena  float64 `json:"fcena" db:"fcena"`
	MCena  float64 `json:"mcena" db:"mcena"`
	MCenaP float64 `json:"mcenap" db:"mcenap"`
	NCena  float64 `json:"ncena" db:"ncena"`

	Ztrin float64 `json:"ztrin" db:"ztrin"`
	Rab   float32 `json:"rab" db:"rab"`

	Tz  string `json:"tz" db:"tz"`
	Tzi string `json:"tzi" db:"tzi"`

	IAkc   float64 `json:"iakc" db:"iakc"`
	PAkc   float32 `json:"pakc" db:"pakc"`
	ITaksa float64 `json:"itaksa" db:"itaksa"`
	PTaksa float32 `json:"ptaksa" db:"ptaksa"`

	URaz   float64 `json:"uraz" db:"uraz"`
	PRCena float64 `json:"prcena" db:"prcena"`

	RceneID int `json:"rceneid" db:"rceneid"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	RSifID sql.NullInt64 `json:"rsifid" db:"rsifid"`

	CenaVal1 float64 `json:"cenaval1" db:"cenaval1"`
	CenaVal2 float64 `json:"cenaval2" db:"cenaval2"`
	CenaVal3 float64 `json:"cenaval3" db:"cenaval3"`
	CenaVal4 float64 `json:"cenaval4" db:"cenaval4"`
}

// Rcenovnik Model
type Rcenovnik struct {
	RcenovnikID int `json:"rcenovnikid" db:"rcenovnikid"`
	God         int `json:"god" db:"god"`
	Kar         int `json:"kar" db:"kar"`

	DatCen sql.NullTime `json:"datcen" db:"datcen"`

	Sifra int `json:"sifra" db:"sifra"`

	VPCena float64 `json:"vpcena" db:"vpcena"`
	MPCena float64 `json:"mpcena" db:"mpcena"`
	NCena  float64 `json:"ncena" db:"ncena"`
	Rabat  float32 `json:"rabat" db:"rabat"`

	RSifID int `json:"rsifid" db:"rsifid"`

	VaziDo string `json:"vazido" db:"vazido"`

	XOpUnos   string       `json:"xopunos" db:"xopunos"`
	XDatUnosa sql.NullTime `json:"xdatunosa" db:"xdatunosa"`

	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`

	Broj           sql.NullInt64 `json:"broj" db:"broj"`
	RcenovnikHdrID sql.NullInt64 `json:"rcenovnikhdrid" db:"rcenovnikhdrid"`
}

// RcenovnikHdr Model
type RcenovnikHdr struct {
	RcenovnikHdrID int `json:"rcenovnikhdrid" db:"rcenovnikhdrid"`
	Broj           int `json:"broj" db:"broj"`

	DatCen sql.NullTime `json:"datcen" db:"datcen"`

	VaziDo string `json:"vazido" db:"vazido"`

	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`

	KlasaCenID sql.NullInt64 `json:"klasacenid" db:"klasacenid"`

	Tip    string `json:"tip" db:"tip"`
	SifVal int    `json:"sifval" db:"sifval"`
}

// Rsif Model
type Rsif struct {
	God    int     `json:"god" db:"god"`
	Kar    int     `json:"kar" db:"kar"`
	Konto  string  `json:"konto" db:"konto"`
	Barkod string  `json:"barkod" db:"barkod"`
	Sifra  int     `json:"sifra" db:"sifra"`
	Pro    string  `json:"pro" db:"pro"`
	Naziv  string  `json:"naziv" db:"naziv"`
	Po     int     `json:"po" db:"po"`
	JM     string  `json:"jm" db:"jm"`
	Pa     int     `json:"pa" db:"pa"`
	Gru    int     `json:"gru" db:"gru"`
	Model  string  `json:"model" db:"model"`
	Miza   float64 `json:"miza" db:"miza"`
	Maza   float64 `json:"maza" db:"maza"`
	KDob   string  `json:"kdob" db:"kdob"`
	Pakov  float32 `json:"pakov" db:"pakov"`
	Dani   int     `json:"dani" db:"dani"`
	SerBr  string  `json:"serbr" db:"serbr"`

	RSifID int `json:"rsifid" db:"rsifid"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	SifraDobav string `json:"sifradobav" db:"sifradobav"`
	PGru       int    `json:"pgru" db:"pgru"`
	RGruID     int    `json:"rgruid" db:"rgruid"`
	RPGruID    int    `json:"rpgruid" db:"rpgruid"`
	Tip        string `json:"tip" db:"tip"`

	ArtRoba   bool `json:"artroba" db:"artroba"`
	ArtProizv bool `json:"artproizv" db:"artproizv"`
	ArtSir    bool `json:"artsir" db:"artsir"`
	ArtUsl    bool `json:"artusl" db:"artusl"`
	ArtAmb    bool `json:"artamb" db:"artamb"`
	ArtPolup  bool `json:"artpolup" db:"artpolup"`

	TezJM      float64 `json:"tezjm" db:"tezjm"`
	Kvalitet   string  `json:"kvalitet" db:"kvalitet"`
	TrPakov    int     `json:"trpakov" db:"trpakov"`
	Kubikaza   float64 `json:"kubikaza" db:"kubikaza"`
	KoefTez    float64 `json:"koeftez" db:"koeftez"`
	PerRekontr int     `json:"perrekontr" db:"perrekontr"`

	ProizSifra   string `json:"proizsifra" db:"proizsifra"`
	KomercOpis   string `json:"komercopis" db:"komercopis"`
	DodatNaziv   string `json:"dodatnaziv" db:"dodatnaziv"`
	FSIme        string `json:"fsime" db:"fsime"`
	ZemljaProizv string `json:"zemljaproizv" db:"zemljaproizv"`

	TarifnaOzn int64 `json:"tarifnaozn" db:"tarifnaozn"`

	BojeID      int `json:"bojeid" db:"bojeid"`
	DebljineID  int `json:"debljineid" db:"debljineid"`
	ObliciID    int `json:"obliciid" db:"obliciid"`
	MetaliID    int `json:"metaliid" db:"metaliid"`
	ModMetalaID int `json:"modmetalaid" db:"modmetalaid"`

	Duzina  float64 `json:"duzina" db:"duzina"`
	Sirina  float64 `json:"sirina" db:"sirina"`
	RazvSir float32 `json:"razvsir" db:"razvsir"`

	AmbSif   int    `json:"ambsif" db:"ambsif"`
	Lokacija string `json:"lokacija" db:"lokacija"`
	RokTra   string `json:"roktra" db:"roktra"`

	ZemljaUvoza string `json:"zemljauvoza" db:"zemljauvoza"`
	Uvoznik     string `json:"uvoznik" db:"uvoznik"`

	GodUvoza int `json:"goduvoza" db:"goduvoza"`
	GodProiz int `json:"godproiz" db:"godproiz"`

	AkciznaKat string `json:"akciznakat" db:"akciznakat"`
}

// Rgru Model
type Rgru struct {
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	RgruID     int            `json:"rgruid" db:"rgruid"`
	Gru        int            `json:"gru" db:"gru"`
	Naziv      string         `json:"naziv" db:"naziv"`
	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
}

// Rpgru Model
type Rpgru struct {
	RpgruID int `json:"rpgruid" db:"rpgruid"`
	God     int `json:"god" db:"god"`
	Kar     int `json:"kar" db:"kar"`
	Gru     int `json:"gru" db:"gru"`
	PGru    int `json:"pgru" db:"pgru"`

	Naziv string `json:"naziv" db:"naziv"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`
	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`

	RGruID int `json:"rgruid" db:"rgruid"`
}

type Rpor struct {
	PP float32 `json:"pp" db:"pp"`
	PT string  `json:"pt" db:"pt"`

	Datum sql.NullTime `json:"datum" db:"datum"`

	Tip int `json:"tip" db:"tip"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	Po         int          `json:"po" db:"po"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	RporID int `json:"rporid" db:"rporid"`

	Slovo string `json:"slovo" db:"slovo"`
}

// Rdok Model
type Rdok struct {
	God   int `json:"god" db:"god"`
	Kar   int `json:"kar" db:"kar"`
	Nalog int `json:"nalog" db:"nalog"`

	Danal sql.NullTime `json:"danal" db:"danal"`
	Dokum int          `json:"dokum" db:"dokum"`
	Dadok sql.NullTime `json:"dadok" db:"dadok"`

	Vrd   int     `json:"vrd" db:"vrd"`
	Oper  string  `json:"oper" db:"oper"`
	Iznos float64 `json:"iznos" db:"iznos"`

	Rbr  int    `json:"rbr" db:"rbr"`
	Opis string `json:"opis" db:"opis"`

	Rok   int          `json:"rok" db:"rok"`
	DatIz sql.NullTime `json:"datiz" db:"datiz"`
	Brst  int          `json:"brst" db:"brst"`
	DatOb sql.NullTime `json:"datob" db:"datob"`

	FKto  string `json:"fkto" db:"fkto"`
	Fana  string `json:"fana" db:"fana"`
	DokIz string `json:"dokiz" db:"dokiz"`

	PorNapomena string       `json:"pornapomena" db:"pornapomena"`
	ProFak      string       `json:"profak" db:"profak"`
	DatPro      sql.NullTime `json:"datpro" db:"datpro"`

	Mi   int    `json:"mi" db:"mi"`
	Otp  string `json:"otp" db:"otp"`
	Kamt string `json:"kamt" db:"kamt"`
	Pla  string `json:"pla" db:"pla"`
	Foot string `json:"foot" db:"foot"`

	PKto string `json:"pkto" db:"pkto"`
	Pana string `json:"pana" db:"pana"`
	OpIz string `json:"opiz" db:"opiz"`

	Odp sql.NullTime `json:"odp" db:"odp"`
	Dop sql.NullTime `json:"dop" db:"dop"`

	Kom int `json:"kom" db:"kom"`

	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xopunos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`

	PlVred   float64 `json:"plvred" db:"plvred"`
	NabVred  float64 `json:"nabvred" db:"nabvred"`
	ProdVred float64 `json:"prodvred" db:"prodvred"`

	RdokID int `json:"rdokid" db:"rdokid"`

	RnalID     sql.NullInt64 `json:"rnalid" db:"rnalid"`
	MagaciniID sql.NullInt64 `json:"magaciniid" db:"magaciniid"`

	TipDok string        `json:"tipdok" db:"tipdok"`
	Mag    int           `json:"mag" db:"mag"`
	KomID  sql.NullInt64 `json:"komid" db:"komid"`

	PKase float32 `json:"pkase" db:"pkase"`
	VKase float64 `json:"vkase" db:"vkase"`

	Knjige1 string `json:"knjige_1" db:"knjige_1"`
	Knjige2 string `json:"knjige_2" db:"knjige_2"`
	Knjige3 string `json:"knjige_3" db:"knjige_3"`
	Knjige4 string `json:"knjige_4" db:"knjige_4"`

	Bi       int `json:"bi" db:"bi"`
	BrojRata int `json:"brojrata" db:"brojrata"`

	DatumPrRate sql.NullTime `json:"datumprrate" db:"datumprrate"`

	SifVal int     `json:"sifval" db:"sifval"`
	Kurs   float64 `json:"kurs" db:"kurs"`
	VaziZa int     `json:"vaziza" db:"vaziza"`

	RbrKepu int     `json:"rbrkepu" db:"rbrkepu"`
	VPorez  float64 `json:"vporez" db:"vporez"`

	MagID1  int `json:"magid1" db:"magid1"`
	VrdokID int `json:"vrdokid" db:"vrdokid"`

	Vozac      string `json:"vozac" db:"vozac"`
	BrVozila   string `json:"brvozila" db:"brvozila"`
	Magacioner string `json:"magacioner" db:"magacioner"`
	Primio     string `json:"primio" db:"primio"`
	Predao     string `json:"predao" db:"predao"`

	IDOrgJed  sql.NullInt64 `json:"idorgjed" db:"idorgjed"`
	MestoTrID sql.NullInt64 `json:"mestotrid" db:"mestotrid"`

	UGRabat   float32 `json:"ugrabat" db:"ugrabat"`
	VRUGRabat float64 `json:"vrugrabat" db:"vrugrabat"`

	ZiroRac   string `json:"zirorac" db:"zirorac"`
	PosUslNab string `json:"posuslnab" db:"posuslnab"`

	RcenovnikHdrID int           `json:"rcenovnikhdrid" db:"rcenovnikhdrid"`
	TipKnjID       sql.NullInt64 `json:"tipknjid" db:"tipknjid"`

	AvansID int `json:"avansid" db:"avansid"`

	BttoWght float64 `json:"bttowght" db:"bttowght"`
	NetWght  float64 `json:"netwght" db:"netwght"`

	TotBoxs  int `json:"totboxs" db:"totboxs"`
	BrPaleta int `json:"brpaleta" db:"brpaleta"`

	Paritet string `json:"paritet" db:"paritet"`

	IzjIzv  int `json:"izjizv" db:"izjizv"`
	SifBank int `json:"sifbank" db:"sifbank"`

	IDOrgJed1  int `json:"idorgjed1" db:"idorgjed1"`
	MestoTrID1 int `json:"mestotrid1" db:"mestotrid1"`

	FlgStariPDV bool `json:"flgstaripdv" db:"flgstaripdv"`

	TKonto string `json:"tkonto" db:"tkonto"`
	TSifra string `json:"tsifra" db:"tsifra"`

	EDokID  int `json:"edokid" db:"edokid"`
	PIDokID int `json:"pidokid" db:"pidokid"`

	LiSifra int `json:"li_sifra" db:"li_sifra"`
	ObSifra int `json:"ob_sifra" db:"ob_sifra"`

	Polje string `json:"polje" db:"polje"`

	InvoiceID      int64 `json:"invoiceid" db:"invoiceid"`
	SalesInvoiceID int64 `json:"salesinvoiceid" db:"salesinvoiceid"`

	StatusSalInv    string       `json:"status_salinv" db:"status_salinv"`
	DatumStatSalInv sql.NullTime `json:"datum_stat_salinv" db:"datum_stat_salinv"`

	Komentar     string `json:"komentar" db:"komentar"`
	CirInvoiceID string `json:"cirinvoiceid" db:"cirinvoiceid"`

	Demo bool `json:"demo" db:"demo"`

	KeyKat   string `json:"keykat" db:"keykat"`
	Category string `json:"category" db:"category"`

	BrUgov string `json:"brugov" db:"brugov"`

	IndividualVATID    int64  `json:"individualvatid" db:"individualvatid"`
	VATRecordingStatus string `json:"vatrecordingstatus" db:"vatrecordingstatus"`

	DatumStatIndVAT sql.NullTime `json:"datum_stat_indvat" db:"datum_stat_indvat"`

	BrTend string `json:"brtend" db:"brtend"`

	JavnaB   bool   `json:"javnab" db:"javnab"`
	JBKJSNJN string `json:"jbkjsnjn" db:"jbkjsnjn"`

	ObrnPDV bool `json:"obrnpdv" db:"obrnpdv"`

	Narudzb string `json:"narudzb" db:"narudzb"`

	KPPeriod bool `json:"kpperiod" db:"kpperiod"`

	OdDat sql.NullTime `json:"oddat" db:"oddat"`
	DoDat sql.NullTime `json:"dodat" db:"dodat"`

	DescriptionCode int    `json:"descriptioncode" db:"descriptioncode"`
	EppPolje        string `json:"epp_polje" db:"epp_polje"`
	FsEppID         int    `json:"fseppid" db:"fseppid"`

	Odbitni bool `json:"odbitni" db:"odbitni"`
}

// Rnal Model
type Rnal struct {
	Nalog int `json:"nalog" db:"nalog"`

	Danal sql.NullTime `json:"danal" db:"danal"`

	Opis string `json:"opis" db:"opis"`

	Dug float64 `json:"dug" db:"dug"`
	Pot float64 `json:"pot" db:"pot"`

	Rbr int `json:"rbr" db:"rbr"`

	DatOb sql.NullTime `json:"datob" db:"datob"`

	Oper string `json:"oper" db:"oper"`

	Brst int `json:"brst" db:"brst"`

	God int `json:"god" db:"god"`
	Kar int `json:"kar" db:"kar"`

	Brdo int `json:"brdo" db:"brdo"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	RnalID int `json:"rnalid" db:"rnalid"`

	TipDok string `json:"tipdok" db:"tipdok"`
	NalSts string `json:"nalsts" db:"nalsts"`

	IDTipDok int `json:"idtipdok" db:"idtipdok"`

	MagaciniID int `json:"magaciniid" db:"magaciniid"`
	Mag        int `json:"mag" db:"mag"`

	PiNalID int `json:"pinalid" db:"pinalid"`
}

// Rpro Model
type Rpro struct {
	Nalog int `json:"nalog" db:"nalog"`
	Vrd   int `json:"vrd" db:"vrd"`
	Dokum int `json:"dokum" db:"dokum"`

	Dadok sql.NullTime `json:"dadok" db:"dadok"`

	Iznos float64 `json:"iznos" db:"iznos"`
	Kolic float64 `json:"kolic" db:"kolic"`

	Brst int `json:"brst" db:"brst"`

	Cena   float64 `json:"cena" db:"cena"`
	PCena  float64 `json:"pcena" db:"pcena"`
	FCena  float64 `json:"fcena" db:"fcena"`
	MCena  float64 `json:"mcena" db:"mcena"`
	MCenaP float64 `json:"mcenap" db:"mcenap"`
	NCena  float64 `json:"ncena" db:"ncena"`

	Rbr int `json:"rbr" db:"rbr"`

	God int `json:"god" db:"god"`
	Kar int `json:"kar" db:"kar"`

	Naz1 string `json:"naz1" db:"naz1"`
	JM   string `json:"jm" db:"jm"`

	Konto string `json:"konto" db:"konto"`

	Po    int `json:"po" db:"po"`
	Sifra int `json:"sifra" db:"sifra"`

	FKto string `json:"fkto" db:"fkto"`
	Fana string `json:"fana" db:"fana"`

	Rab float32 `json:"rab" db:"rab"`

	Mag int `json:"mag" db:"mag"`

	PKto string `json:"pkto" db:"pkto"`
	Pana string `json:"pana" db:"pana"`

	Ztro  float64 `json:"ztro" db:"ztro"`
	Ztrin float64 `json:"ztrin" db:"ztrin"`

	Tz  string `json:"tz" db:"tz"`
	Tzi string `json:"tzi" db:"tzi"`

	Vma float32 `json:"vma" db:"vma"`
	Mma float32 `json:"mma" db:"mma"`
	Vra float64 `json:"vra" db:"vra"`
	Mra float64 `json:"mra" db:"mra"`

	Model string `json:"model" db:"model"`

	IAkc   float64 `json:"iakc" db:"iakc"`
	PAkc   float32 `json:"pakc" db:"pakc"`
	ITaksa float64 `json:"itaksa" db:"itaksa"`
	PTaksa float32 `json:"ptaksa" db:"ptaksa"`

	Dani int `json:"dani" db:"dani"`

	Ztrof float64 `json:"ztrof" db:"ztrof"`

	RproID int `json:"rproid" db:"rproid"`

	RdokID sql.NullInt64 `json:"rdokid" db:"rdokid"`

	StStatus string `json:"ststatus" db:"ststatus"`

	RnalID sql.NullInt64 `json:"rnalid" db:"rnalid"`

	MagaciniID int `json:"magaciniid" db:"magaciniid"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	Otk    string `json:"otk" db:"otk"`
	Serija string `json:"serija" db:"serija"`

	RokTr sql.NullInt64 `json:"roktr" db:"roktr"`

	RSifID sql.NullInt64 `json:"rsifid" db:"rsifid"`

	CenaVal float64 `json:"cenaval" db:"cenaval"`

	RproID1 int `json:"rproid1" db:"rproid1"`

	Kolic1 float64 `json:"kolic1" db:"kolic1"`

	PDVPct float32 `json:"pdvpct" db:"pdvpct"`

	VPCena float64 `json:"vpcena" db:"vpcena"`

	PinArID sql.NullInt64 `json:"pinarid" db:"pinarid"`

	KonProSnc float64 `json:"konprosnc" db:"konprosnc"`
	CenaOld   float64 `json:"cenaold" db:"cenaold"`

	TaxCat    string `json:"taxcat" db:"taxcat"`
	TaxExReco string `json:"taxexreco" db:"taxexreco"`
	SifArtKup string `json:"sifartkup" db:"sifartkup"`
}

// Rppo Model
type Rppo struct {
	God   int `json:"god" db:"god"`
	Kar   int `json:"kar" db:"kar"`
	Vrd   int `json:"vrd" db:"vrd"`
	Dokum int `json:"dokum" db:"dokum"`

	PPor float64 `json:"ppor" db:"ppor"`
	Po   int     `json:"po" db:"po"`

	Konto string `json:"konto" db:"konto"`
	Sifra string `json:"sifra" db:"sifra"`

	RporID int `json:"rporid" db:"rporid"`

	Dadok sql.NullTime `json:"dadok" db:"dadok"`

	Ter int `json:"ter" db:"ter"`

	Rac string `json:"rac" db:"rac"`

	Osn float64 `json:"osn" db:"osn"`
	Pdv float64 `json:"pdv" db:"pdv"`

	XDatUnosa  sql.NullTime `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime `json:"xdatizmene" db:"xdatizmene"`

	XOpUnos   string         `json:"xopunos" db:"xopunos"`
	XOpIzmene sql.NullString `json:"xopizmene" db:"xopizmene"`

	RppoID int `json:"rppoid" db:"rppoid"`

	RdokID sql.NullInt64 `json:"rdokid" db:"rdokid"`
	RztrID sql.NullInt64 `json:"rztrid" db:"rztrid"`

	EDokID int `json:"edokid" db:"edokid"`
}

// Rztr Model
type Rztr struct {
	Dadok sql.NullTime `json:"dadok" db:"dadok"`

	Dokum int `json:"dokum" db:"dokum"`

	Fana string `json:"fana" db:"fana"`
	FKto string `json:"fkto" db:"fkto"`

	God int `json:"god" db:"god"`

	Iznos float64 `json:"iznos" db:"iznos"`

	Kar int `json:"kar" db:"kar"`

	Kolic float64 `json:"kolic" db:"kolic"`

	Rok int `json:"rok" db:"rok"`

	TKonto string `json:"tkonto" db:"tkonto"`

	Vrd int `json:"vrd" db:"vrd"`

	Rac int `json:"rac" db:"rac"`

	Opis string `json:"opis" db:"opis"`

	IznFak float64 `json:"iznfak" db:"iznfak"`

	RztrID int `json:"rztrid" db:"rztrid"`

	RdokID sql.NullInt64 `json:"rdokid" db:"rdokid"`

	DokDob string `json:"dokdob" db:"dokdob"`

	EDokID int `json:"edokid" db:"edokid"`

	PorNapomena string `json:"pornapomena" db:"pornapomena"`

	FlgUvozPDV bool `json:"flguvozpdv" db:"flguvozpdv"`

	TSifra string `json:"tsifra" db:"tsifra"`

	Polje string `json:"polje" db:"polje"`
}

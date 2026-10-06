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

// Rnal represents the "baza.rnal" table (zaglavlje robnog naloga).
type Rnal struct {
	RnalID     int64          `json:"rnalid" db:"rnalid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Tipdok     string         `json:"tipdok" db:"tipdok" form:"tipdok"`
	IDTipdok   sql.NullInt64  `json:"idtipdok" db:"idtipdok" form:"idtipdok"`
	Nalog      int            `json:"nalog" db:"nalog" form:"nalog"`
	Rbr        int            `json:"rbr" db:"rbr"`
	Danal      sql.NullTime   `json:"danal" db:"danal" form:"danal" format:"date"`
	Datob      sql.NullTime   `json:"datob" db:"datob" form:"datob" format:"date"`
	Opis       string         `json:"opis" db:"opis" form:"opis"`
	Dug        float64        `json:"dug" db:"dug"`
	Pot        float64        `json:"pot" db:"pot"`
	Brdo       int            `json:"brdo" db:"brdo"`
	Brst       int            `json:"brst" db:"brst"`
	Oper       sql.NullString `json:"oper" db:"oper"`
	Nalsts     sql.NullString `json:"nalsts" db:"nalsts"`
	MagaciniID sql.NullInt64  `json:"magaciniid" db:"magaciniid" form:"magaciniid"`
	Mag        sql.NullInt64  `json:"mag" db:"mag"`
	PinalID    sql.NullInt64  `json:"pinalid" db:"pinalid"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa" format:"datetime"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene" format:"datetime"`
	XOpUnos    sql.NullString `json:"xopunos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
}

// RobnoDokumentaDto is one row of the grids of the "Robna dokumenta" option: a robni nalog (rnal)
type RobnoDokumentaDto struct {
	// Identifiers of the robni nalog (rnal) and of the robni dokument (rdok).
	God    int64 `json:"god" db:"god"`
	Kar    int64 `json:"kar" db:"kar"`
	RnalID int64 `json:"rnalid" db:"rnalid"`
	RdokID int64 `json:"rdokid" db:"rdokid"`

	// Header of the nalog and of the document: the vrsta naloga (rnal.tipdok / rdok.tipdok), the
	Tipdok     string        `json:"tipdok" db:"tipdok"`
	Nalog      int           `json:"nalog" db:"nalog"`
	MagaciniID sql.NullInt64 `json:"magaciniid" db:"magaciniid"`
	Mag        sql.NullInt64 `json:"mag" db:"mag"`
	Danal      sql.NullTime  `json:"danal" db:"danal"`
	Opis       string        `json:"opis" db:"opis"`
	Datob      sql.NullTime  `json:"datob" db:"datob"`
	Oper       string        `json:"oper" db:"oper"`

	// Document: the vrsta dokumenta (rdok.vrd), the broj dokumenta, its date and the date of its
	Vrd     sql.NullInt64 `json:"vrd" db:"vrd"`
	Dokum   sql.NullInt64 `json:"dokum" db:"dokum"`
	Dadok   sql.NullTime  `json:"dadok" db:"dadok"`
	Dop     sql.NullTime  `json:"dop" db:"dop"`
	Dokiz   string        `json:"dokiz" db:"dokiz"`
	Datiz   sql.NullTime  `json:"datiz" db:"datiz"`
	Polje   string        `json:"polje" db:"polje"`
	Rok     sql.NullInt64 `json:"rok" db:"rok"`
	Knjige1 string        `json:"knjige1" db:"knjige1"`

	// Totals of the nalog (rnal) and amount of the document (rdok.iznos).
	Brdo  int     `json:"brdo" db:"brdo"`
	Brst  int     `json:"brst" db:"brst"`
	Dug   float64 `json:"dug" db:"dug"`
	Pot   float64 `json:"pot" db:"pot"`
	Iznos float64 `json:"iznos" db:"iznos"`

	// Tab 3 - "Specifikacije dokumenta": the rows of the tab are the stavke (rpro) of the robni
	Rabat float64 `json:"rabat" db:"rabat"`
	Porez float64 `json:"porez" db:"porez"`
	Stopa float64 `json:"stopa" db:"stopa"`

	// Partner of the document (fkpl/partneri) and the broj avansnog računa it was created from.
	Fkto        string        `json:"fkto" db:"fkto"`
	Fana        string        `json:"fana" db:"fana"`
	Naziv       string        `json:"naziv" db:"naziv"`
	Pib         string        `json:"pib" db:"pib"`
	Jbkjs       string        `json:"jbkjs" db:"jbkjs"`
	Adresa      string        `json:"adresa" db:"adresa"`
	Mesto       string        `json:"mesto" db:"mesto"`
	Avansi      string        `json:"avansi" db:"avansi"`
	Pkto        string        `json:"pkto" db:"pkto"`
	Pana        string        `json:"pana" db:"pana"`
	Naziv1      string        `json:"naziv1" db:"naziv1"`
	Pornapomena string        `json:"pornapomena" db:"pornapomena"`
	Tkonto      string        `json:"tkonto" db:"tkonto"`

	// eFaktura (SEF) status of the document.
	StatusSalinv       string          `json:"statussalinv" db:"statussalinv"`
	DatumStatSalinv    sql.NullTime    `json:"datumstatsalinv" db:"datumstatsalinv"`
	Komentar           string          `json:"komentar" db:"komentar"`
	CirInvoiceID       string          `json:"cirinvoiceid" db:"cirinvoiceid"`
	VatRecordingStatus string          `json:"vatrecordingstatus" db:"vatrecordingstatus"`
	DatumStatIndVat    sql.NullTime    `json:"datumstatindvat" db:"datumstatindvat"`
	Valuta             string          `json:"valuta" db:"valuta"`
	Kurs               float64         `json:"kurs" db:"kurs"`
	SalesInvoiceID     sql.NullFloat64 `json:"salesinvoiceid" db:"salesinvoiceid"`
}

// RobnoDokumentaTotalsDto is the aggregate row behind the "Prikaz ukupne obrade" panel.
type RobnoDokumentaTotalsDto struct {
	UkNaloga     int64   `json:"uknaloga" db:"uknaloga"`
	UkDokumenata int64   `json:"ukdokumenata" db:"ukdokumenata"`
	UkStavki     int64   `json:"ukstavki" db:"ukstavki"`
	UkDuguje     float64 `json:"ukduguje" db:"ukduguje"`
	UkPotrazuje  float64 `json:"ukpotrazuje" db:"ukpotrazuje"`
	Duguje       float64 `json:"duguje" db:"duguje"`
	Potrazuje    float64 `json:"potrazuje" db:"potrazuje"`
	Saldo        float64 `json:"saldo" db:"saldo"`
}
type RobnoDokumentaTotal struct {
	UkNaloga     string
	UkDokumenata string
	UkStavki     string
	UkDuguje     string
	UkPotrazuje  string
	Duguje       string
	Potrazuje    string
	Saldo        string
}

// RobnoDokumentaParams is the single parameter structure of the "Robna dokumenta" option: it holds
type RobnoDokumentaParams struct {
	// Tab 1 - Unos dokumenta: the header of the robni nalog (vrsta naloga, vrsta dokumenta, broj
	// naloga, datum naloga, datum obrade, opis knjiženja and magacin) and the filters of its grid.
	Tipdok     string `json:"tipdok" form:"tipdok"`
	Vrd        string `json:"vrd" form:"vrd"`
	Nalog      string `json:"nalog" form:"nalog"`
	Danal      string `json:"danal" form:"danal"`
	Datob      string `json:"datob" form:"datob"`
	Opis       string `json:"opis" form:"opis"`
	MagaciniID int    `json:"magaciniid" form:"magaciniid"`

	// Tab 2 - Pregled dokumenta: GrupeDokumenata is the comma separated list of the document groups
	GrupeDokumenata string `json:"grupedokumenata" form:"grupedokumenata"`
	DatumStatusa    string `json:"datumstatusa" form:"datumstatusa"`

	// OdDanal/DoDanal is the range of the dates: the "Od/Do datuma naloga" of the Štampa sub-tab of
	OdDanal string `json:"oddanal" form:"oddanal"`
	DoDanal string `json:"dodanal" form:"dodanal"`

	// Tab 4 - Kontiranje dokumenata: the ranges of the broj naloga (OdNaloga/DoNaloga) and of the
	OdNaloga string `json:"odnaloga" form:"odnaloga"`
	DoNaloga string `json:"donaloga" form:"donaloga"`
	OdDokum  string `json:"oddokum" form:"oddokum"`
	DoDokum  string `json:"dodokum" form:"dodokum"`
	// Proknjizen is the state selected with the radio buttons of the last two sub-tabs of
	Proknjizen           string `json:"proknjizen" form:"proknjizen"`
	OznaciNeproknjizenim bool   `json:"oznacineproknjizenim" form:"oznacineproknjizenim"`

	// Tabs 7 and 8 - Prikaz naloga and Prikaz dokumenata u nalogu: OdVrd/DoVrd is the range of the
	OdVrd string `json:"odvrd" form:"odvrd"`
	DoVrd string `json:"dovrd" form:"dovrd"`

	// Tabs 7, 8 and 9 - the optional filters: every filter is applied only when its checkbox is
	ChkDatumNaloga bool   `json:"chkpodatumunaloga" form:"chkpodatumunaloga"`
	ChkDatumObrade bool   `json:"chkpodatumuobrade" form:"chkpodatumuobrade"`
	OdDatob        string `json:"oddatob" form:"oddatob"`
	DoDatob        string `json:"dodatob" form:"dodatob"`
	ChkOperator    bool   `json:"chkpooperateru" form:"chkpooperateru"`
	Oper           string `json:"oper" form:"oper"`

	// Tab 3 - Specifikacije dokumenta: the state of the print (the "Štampaj samo zbir" checkbox and
	StampajSamoZbir  bool   `json:"stampajsamozbir" form:"stampajsamozbir"`
	TipSpecifikacije string `json:"tipspecifikacije" form:"tipspecifikacije"`

	// SearchText is the text of the search box of the grid of the active tab.
	SearchText string `json:"searchText" form:"query"`
}

// PrikazUkupneObradeDto is one row of the grid of the "Prikaz ukupne obrade" tab: one magacin of the
type PrikazUkupneObradeDto struct {
	Mag          int     `db:"mag"`
	Opis         string  `db:"opis"`
	Mesto        string  `db:"mesto"`
	BrojNaloga   int     `db:"brojnaloga"`
	UkDokumenata int     `db:"ukdokumenata"`
	BrStavki     int     `db:"brstavki"`
	Duguje       float64 `db:"duguje"`
	Potrazuje    float64 `db:"potrazuje"`
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

// Drsta represents the baza.drsta table
type Drsta struct {
	God        int     `db:"god" json:"god"`
	Kar        int     `db:"kar" json:"kar"`
	Konto      string  `db:"konto" json:"konto"`
	Dug        float64 `db:"dug" json:"dug"`
	Pot        float64 `db:"pot" json:"pot"`
	Mdug1      float64 `db:"mdug_1" json:"mdug_1"`
	Mdug2      float64 `db:"mdug_2" json:"mdug_2"`
	Mdug3      float64 `db:"mdug_3" json:"mdug_3"`
	Mdug4      float64 `db:"mdug_4" json:"mdug_4"`
	Mdug5      float64 `db:"mdug_5" json:"mdug_5"`
	Mdug6      float64 `db:"mdug_6" json:"mdug_6"`
	Mdug7      float64 `db:"mdug_7" json:"mdug_7"`
	Mdug8      float64 `db:"mdug_8" json:"mdug_8"`
	Mdug9      float64 `db:"mdug_9" json:"mdug_9"`
	Mdug10     float64 `db:"mdug_10" json:"mdug_10"`
	Mdug11     float64 `db:"mdug_11" json:"mdug_11"`
	Mdug12     float64 `db:"mdug_12" json:"mdug_12"`
	Mpot1      float64 `db:"mpot_1" json:"mpot_1"`
	Mpot2      float64 `db:"mpot_2" json:"mpot_2"`
	Mpot3      float64 `db:"mpot_3" json:"mpot_3"`
	Mpot4      float64 `db:"mpot_4" json:"mpot_4"`
	Mpot5      float64 `db:"mpot_5" json:"mpot_5"`
	Mpot6      float64 `db:"mpot_6" json:"mpot_6"`
	Mpot7      float64 `db:"mpot_7" json:"mpot_7"`
	Mpot8      float64 `db:"mpot_8" json:"mpot_8"`
	Mpot9      float64 `db:"mpot_9" json:"mpot_9"`
	Mpot10     float64 `db:"mpot_10" json:"mpot_10"`
	Mpot11     float64 `db:"mpot_11" json:"mpot_11"`
	Mpot12     float64 `db:"mpot_12" json:"mpot_12"`
	Ulaz       float64 `db:"ulaz" json:"ulaz"`
	Izlaz      float64 `db:"izlaz" json:"izlaz"`
	Mul1       float64 `db:"mul_1" json:"mul_1"`
	Mul2       float64 `db:"mul_2" json:"mul_2"`
	Mul3       float64 `db:"mul_3" json:"mul_3"`
	Mul4       float64 `db:"mul_4" json:"mul_4"`
	Mul5       float64 `db:"mul_5" json:"mul_5"`
	Mul6       float64 `db:"mul_6" json:"mul_6"`
	Mul7       float64 `db:"mul_7" json:"mul_7"`
	Mul8       float64 `db:"mul_8" json:"mul_8"`
	Mul9       float64 `db:"mul_9" json:"mul_9"`
	Mul10      float64 `db:"mul_10" json:"mul_10"`
	Mul11      float64 `db:"mul_11" json:"mul_11"`
	Mul12      float64 `db:"mul_12" json:"mul_12"`
	Miz1       float64 `db:"miz_1" json:"miz_1"`
	Miz2       float64 `db:"miz_2" json:"miz_2"`
	Miz3       float64 `db:"miz_3" json:"miz_3"`
	Miz4       float64 `db:"miz_4" json:"miz_4"`
	Miz5       float64 `db:"miz_5" json:"miz_5"`
	Miz6       float64 `db:"miz_6" json:"miz_6"`
	Miz7       float64 `db:"miz_7" json:"miz_7"`
	Miz8       float64 `db:"miz_8" json:"miz_8"`
	Miz9       float64 `db:"miz_9" json:"miz_9"`
	Miz10      float64 `db:"miz_10" json:"miz_10"`
	Miz11      float64 `db:"miz_11" json:"miz_11"`
	Miz12      float64 `db:"miz_12" json:"miz_12"`
	Pos        float64 `db:"pos" json:"pos"`
	Pov        float64 `db:"pov" json:"pov"`
	Sifra      int     `db:"sifra" json:"sifra"`
	Nivel      float64 `db:"nivel" json:"nivel"`
	Cena       float64 `db:"cena" json:"cena"`
	Rez        float64 `db:"rez" json:"rez"`
	Rstaid     int     `db:"rstaid" json:"rstaid"`
	Rsifid     int     `db:"rsifid" json:"rsifid"`
	Magaciniid int     `db:"magaciniid" json:"magaciniid"`
	Mag        int16   `db:"mag" json:"mag"`
	Ncena      float64 `db:"ncena" json:"ncena"`
	Otk        string  `db:"otk" json:"otk"`
	Serija     string  `db:"serija" json:"serija"`
	Roktr      float64 `db:"roktr" json:"roktr"`
	Zstatus    string  `db:"zstatus" json:"zstatus"`
	Kolicner   float64 `db:"kolicner" json:"kolicner"`
	Prosnc     float64 `db:"prosnc" json:"prosnc"`
	Vpcena     float64 `db:"vpcena" json:"vpcena"`
}

type RobnoStanjeDto struct {
	Magacin      int     `json:"magacin" db:"magacin"`
	Konto        string  `json:"konto" db:"konto"`
	Sifra        string  `json:"sifra" db:"sifra"`
	SifraArtikla string  `json:"sifra_artikla" db:"sifra_artikla"`
	NazivArtikla string  `json:"naziv_artikla" db:"naziv_artikla"`
	Jm           string  `json:"jm" db:"jm"`
	Cena         float64 `json:"cena" db:"cena"`
	Prosnc       float64 `json:"prosnc" db:"prosnc"`
	Gru          int     `json:"gru" db:"gru"`
	KontoNaziv   string  `json:"kontonaziv" db:"kontonaziv"`
	MagacinNaziv string  `json:"magacinnaziv" db:"magacinnaziv"`
	Mesec        int     `json:"mesec" db:"mesec"`
	MDug1        float64 `json:"mdug_1" db:"mdug_1"`
	MDug2        float64 `json:"mdug_2" db:"mdug_2"`
	MDug3        float64 `json:"mdug_3" db:"mdug_3"`
	MDug4        float64 `json:"mdug_4" db:"mdug_4"`
	MDug5        float64 `json:"mdug_5" db:"mdug_5"`
	MDug6        float64 `json:"mdug_6" db:"mdug_6"`
	MDug7        float64 `json:"mdug_7" db:"mdug_7"`
	MDug8        float64 `json:"mdug_8" db:"mdug_8"`
	MDug9        float64 `json:"mdug_9" db:"mdug_9"`
	MDug10       float64 `json:"mdug_10" db:"mdug_10"`
	MDug11       float64 `json:"mdug_11" db:"mdug_11"`
	MDug12       float64 `json:"mdug_12" db:"mdug_12"`
	MPot1        float64 `json:"mpot_1" db:"mpot_1"`
	MPot2        float64 `json:"mpot_2" db:"mpot_2"`
	MPot3        float64 `json:"mpot_3" db:"mpot_3"`
	MPot4        float64 `json:"mpot_4" db:"mpot_4"`
	MPot5        float64 `json:"mpot_5" db:"mpot_5"`
	MPot6        float64 `json:"mpot_6" db:"mpot_6"`
	MPot7        float64 `json:"mpot_7" db:"mpot_7"`
	MPot8        float64 `json:"mpot_8" db:"mpot_8"`
	MPot9        float64 `json:"mpot_9" db:"mpot_9"`
	MPot10       float64 `json:"mpot_10" db:"mpot_10"`
	MPot11       float64 `json:"mpot_11" db:"mpot_11"`
	MPot12       float64 `json:"mpot_12" db:"mpot_12"`
	MUl1         float64 `json:"mul_1" db:"mul_1"`
	MUl2         float64 `json:"mul_2" db:"mul_2"`
	MUl3         float64 `json:"mul_3" db:"mul_3"`
	MUl4         float64 `json:"mul_4" db:"mul_4"`
	MUl5         float64 `json:"mul_5" db:"mul_5"`
	MUl6         float64 `json:"mul_6" db:"mul_6"`
	MUl7         float64 `json:"mul_7" db:"mul_7"`
	MUl8         float64 `json:"mul_8" db:"mul_8"`
	MUl9         float64 `json:"mul_9" db:"mul_9"`
	MUl10        float64 `json:"mul_10" db:"mul_10"`
	MUl11        float64 `json:"mul_11" db:"mul_11"`
	MUl12        float64 `json:"mul_12" db:"mul_12"`
	MIz1         float64 `json:"miz_1" db:"miz_1"`
	MIz2         float64 `json:"miz_2" db:"miz_2"`
	MIz3         float64 `json:"miz_3" db:"miz_3"`
	MIz4         float64 `json:"miz_4" db:"miz_4"`
	MIz5         float64 `json:"miz_5" db:"miz_5"`
	MIz6         float64 `json:"miz_6" db:"miz_6"`
	MIz7         float64 `json:"miz_7" db:"miz_7"`
	MIz8         float64 `json:"miz_8" db:"miz_8"`
	MIz9         float64 `json:"miz_9" db:"miz_9"`
	MIz10        float64 `json:"miz_10" db:"miz_10"`
	MIz11        float64 `json:"miz_11" db:"miz_11"`
	MIz12        float64 `json:"miz_12" db:"miz_12"`
	ReportTip    string  `json:"report_tip" db:"reporttip"`
	Ulaz         float64 `json:"ulaz" db:"ulaz"`
	Izlaz        float64 `json:"izlaz" db:"izlaz"`
	Stanje       float64 `json:"stanje" db:"stanje"`
	Duguje       float64 `json:"duguje" db:"duguje"`
	Potrazuje    float64 `json:"potrazuje" db:"potrazuje"`
	Saldo        float64 `json:"saldo" db:"saldo"`
}
type RobnoStanjaTotal struct {
	Ulaz            float64 `json:"ulaz" db:"ulaz"`
	Izlaz           float64 `json:"izlaz" db:"izlaz"`
	Stanje          float64 `json:"saldo" db:"stanje"`
	Duguje          float64 `json:"duguje" db:"duguje"`
	Potrazuje       float64 `json:"potrazuje" db:"potrazuje"`
	Saldo           float64 `json:"fin_saldo" db:"saldo"`
	Sifra           string  `json:"sifra" db:"sifra"`
	NazivArtikla    string  `json:"naziv_artikla" db:"naziv_artikla"`
	Jm              string  `json:"jm" db:"jm"`
	Cena            float64 `json:"cena" db:"cena"`
	KontoNaziv      string  `json:"kontonaziv" db:"kontonaziv"`
	MagacinNaziv    string  `json:"magacinnaziv" db:"magacinnaziv"`
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

type RobnoStanjaParams struct {
	MagaciniID             int     `json:"magaciniid" db:"magaciniid"`
	Magacin                int     `json:"magacin" db:"magacin"`
	Konto                  string  `json:"konto" db:"konto"`
	SifraArtikla           string  `json:"sifra_artikla" db:"sifra_artikla"`
	OdKonta                string  `json:"odkonta" db:"odkonta"`
	DoKonta                string  `json:"dokonta" db:"dokonta"`
	OdSifre                string  `json:"odsifre" db:"odsifre"`
	DoSifre                string  `json:"dosifre" db:"dosifre"`
	OdGrupe                string  `json:"odgrupe" db:"odgrupe"`
	DoGrupe                string  `json:"dogrupe" db:"dogrupe"`
	OdMeseca               string  `json:"odmeseca" db:"odmeseca"`
	DoMeseca               string  `json:"domeseca" db:"domeseca"`
	FinansijskiIznos       bool    `json:"finansijskiiznos" db:"finansijskiiznos"`
	ArtikliSaStanjem       bool    `json:"artiklisastanjem" db:"artiklisastanjem"`
	ArtikliBezStanja       bool    `json:"artiklibezstanja" db:"artiklibezstanja"`
	ProsecnaCenaStanje     bool    `json:"prosecnacenastanje" db:"prosecnacenastanje"`
	ProsecnaCenaUlaz       bool    `json:"prosecnacenaulaz" db:"prosecnacenaulaz"`
	ZaDobavljaca           bool    `json:"zadobavljaca" db:"zadobavljaca"`
	NovaStranaPoGrupi      bool    `json:"novastranapogrupi" db:"novastranapogrupi"`
	NacinSvodjenja         string  `json:"nacinsvodjenja" db:"nacinsvodjenja"`
	ObradiArtikleSaStanjem bool    `json:"obradiartiklesastanjem" db:"obradiartiklesastanjem"`
	VrstaNaloga            string  `json:"vrstanaloga" db:"vrstanaloga"`
	IdOrgJed               int     `json:"idorgjed" db:"idorgjed"`
	MestoTroskaID          int     `json:"mestotroskaid" db:"mestotroskaid"`
	BrojNaloga             string  `json:"brojnaloga" db:"brojnaloga"`
	DatumNaloga            string  `json:"datumnaloga" db:"datumnaloga"`
	DatumObradeNaloga      string  `json:"datumobradenaloga" db:"datumobradenaloga"`
	OpisKnjizenja          string  `json:"opisknjizenja" db:"opisknjizenja"`
	NazivArtikla           string  `json:"naziv_artikla" db:"naziv_artikla"`
	Cena                   float64 `json:"cena" db:"cena"`
	ReportTip              string  `json:"report_tip" db:"reporttip"`
	SearchText             string  `json:"searchtext" db:"searchtext"`
}

type RobnoKarticaDto struct {
	Magacin      int       `json:"magacin" db:"magacin"`
	Konto        string    `json:"konto" db:"konto"`
	Tipdok       string    `json:"tipdok" db:"tipdok"`
	Nalog        int       `json:"nalog" db:"nalog"`
	Danal        time.Time `json:"danal" db:"danal"`
	Vrd          string    `json:"vrd" db:"vrd"`
	Dokum        int       `json:"dokum" db:"dokum"`
	Dadok        time.Time `json:"dadok" db:"dadok"`
	Opis         string    `json:"opis" db:"opis"`
	Sifra        string    `json:"sifra" db:"sifra"`
	NazivArtikla string    `json:"naziv_artikla" db:"naziv_artikla"`
	Cena         float64   `json:"cena" db:"cena"`
	Fcena        float64   `json:"fcena" db:"fcena"`
	Ulaz         float64   `json:"ulaz" db:"ulaz"`
	Izlaz        float64   `json:"izlaz" db:"izlaz"`
	Iznos        float64   `json:"iznos" db:"iznos"`
	Duguje       float64   `json:"duguje" db:"duguje"`
	Potrazuje    float64   `json:"potrazuje" db:"potrazuje"`
	Saldo        float64   `json:"saldo" db:"saldo"`
	Stanje       float64   `json:"stanje" db:"stanje"`
	Fkto         string    `json:"fkto" db:"fkto"`
	Fana         string    `json:"fana" db:"fana"`
	Fkplnaz      string    `json:"fkplnaz" db:"fkplnaz"`
	Valuta       string    `json:"valuta" db:"valuta"`
	Kurs         float64   `json:"kurs" db:"kurs"`
	CenaVal      float64   `json:"cenaval" db:"cenaval"`
	Mesec        int       `json:"mesec" db:"mesec"`
	Dokiz        string    `json:"dokiz" db:"dokiz"`
	Dadokiz      time.Time `json:"dadokiz" db:"dadokiz"`
	Otk          string    `json:"otk" db:"otk"`
	Serija       string    `json:"serija" db:"serija"`
	Roktr        int       `json:"rok" db:"roktr"`
	Napomena     string    `json:"napomena" db:"napomena"`
}
type RobnoKarticaParams struct {
	Magacin           int
	Konto             string
	Nalozi            string
	OdDanal           string
	DoDanal           string
	OdDatumObrade     string
	DoDatumObrade     string
	SifVrsteDokumenta string
	BrojDokumenta     string
	OdDatumaDok       string
	DoDatumaDok       string
	OdIznosa          float64
	DoIznosa          float64
	OdSifre           string
	DoSifre           string
	Cena              float64
	CbxBrojNaloga     bool
	CbxDatumNaloga    bool
	CbxDatumObrade    bool
	CbxVrstaDokumenta bool
	CbxBrojDokumenta  bool
	CbxDatumDokumenta bool
	CbxIznos          bool
	CbxRPROID         bool
	StampajPoMesecima bool
	Sortiranje        string
	ReportTip         string
	SearchText        string
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

type RobnoPrometParams struct {
	OdMagacina     string
	DoMagacina     string
	OdSifreArtikla string
	DoSifreArtikla string
	OdGrupe        string
	DoGrupe        string
	OdDatuma       string
	DoDatuma       string

	// Tab 4 - Lager lista ulaz/izlaz RUC (sub-tab "Lager lista"): the price of the lager (TipCene:
	// "netofakturna", "prosecnanabavna" or "prodajna"), the date of the stanje (yyyy-mm-dd), the range
	// of the podgrupe and the options of the list: only the articles with a stanje <> 0, in the
	// azbučni red of their naziv, or by grupa and podgrupa.
	TipCene              string
	StanjeNaDan          string
	OdPodgrupe           string
	DoPodgrupe           string
	ZaliheOdNule         bool
	AzbucniRed           bool
	StampajGrupaPodgrupa bool

	// Tab 4 - sub-tab "Ulaz/izlaz za period": the promet of the period (UlazIzlaz: "ulaz", "izlaz" or
	// "ulazizlaz") and the sum of the selected magacini (one row per article instead of one row per
	// magacin and article). The period is OdDatuma - DoDatuma.
	UlazIzlaz    string
	ZbirMagacina bool

	// Tab 4 - sub-tab "RUC po magacinima": print only the total of every magacin (the checkbox "Štampaj
	// samo zbir po magacinima"). The period is OdDatuma - DoDatuma.
	StampajSamoZbir bool

	// Tab 4 - sub-tab "Izlazne fakture": include the stavke of the usluge (rsif.tip 'U').
	UkljuceneUsluge bool
}

// RobnoPrometDto is the single row type of the reports of the "Robno promet" option. Every report
// query selects only the columns of its own report (the fields that are not selected keep their
// zero value), so one type carries both the article rows of "Promet artikala po grupama" and the
// partner rows (konto + šifra of the stavka) of "Promet artikala po kupcima", "Nabavka po
// dobavljačima" and the two gradilišta reports (konto + šifra = pkto + pana of the document).
//
// The money columns are already the rounded per-stavka values summed in SQL (numeric), because the
// legacy procedures round (and tax) every stavka before summing it.
//
// SifraArt is the ARTICLE code (rpro.sifra) and Sifra the PARTNER code (rdok.fana), so the two never
// collide. Konto, Naziv and JM are shared on purpose: in the article report they hold the konto and
// the naziv/jm of the ARTICLE (rpro.konto, rsif.naziv, rsif.jm), in the two partner reports the
// ones of the PARTNER (rdok.fkto/fana -> fkpl -> partneri).
// The ClassRow of the rows of the lager list printed by grupe and podgrupe: the header of a magacin, of
// a robna grupa and of a robna podgrupa (Fields: the number and the naziv) and their totals (Fields:
// the number, the naziv, the stanje and the vrednost).
const (
	RobnoPrometLagerMagacin        = "lager-magacin"
	RobnoPrometLagerGrupa          = "lager-grupa"
	RobnoPrometLagerPodgrupa       = "lager-podgrupa"
	RobnoPrometLagerMagacinUkupno  = "lager-magacin-total"
	RobnoPrometLagerGrupaUkupno    = "lager-grupa-total"
	RobnoPrometLagerPodgrupaUkupno = "lager-podgrupa-total"
)

// The ClassRow of the rows of the print of the "RUC po magacinima" added by the service: the header of a
// magacin (Fields: the magacin and its naziv) and its total (Fields: the magacin and its naziv in the
// first two columns, then the totals in their columns).
const (
	RobnoPrometRucMagacin       = "ruc-magacin"
	RobnoPrometRucMagacinUkupno = "ruc-magacin-total"
)

type RobnoPrometDto struct {
	// Shared: the konto and the naziv/jm of the article (tab 1) or of the partner (tabs 2 and 3).
	Konto string `json:"konto" db:"konto"`
	Naziv string `json:"naziv" db:"naziv"`
	JM    string `json:"jm" db:"jm"`

	// Tab 1 - Promet artikala po grupama za period: one row per article.
	SifraArt int     `json:"sifra_art" db:"sifra_art"`
	Gru      int     `json:"gru" db:"gru"`
	Ulaz     float64 `json:"ulaz" db:"ulaz"`
	Izlaz    float64 `json:"izlaz" db:"izlaz"`
	Finulaz  float64 `json:"finulaz" db:"finulaz"`
	Finizlaz float64 `json:"finizlaz" db:"finizlaz"`

	// Tabs 2, 3, 4 and 5 - the partner of the staves: his šifra and the naziv/adresa/mesto of the
	// partner (tabs 2 and 3, rdok.fana -> fkpl) or of the gradilište (tabs 4 and 5, rdok.pana).
	Sifra  string `json:"sifra" db:"sifra"`
	Adresa string `json:"adresa" db:"adresa"`
	Mesto  string `json:"mesto" db:"mesto"`

	// Tab 2 - Promet artikala po kupcima za period. Rabat is shared with tab 3.
	Iznos     float64 `json:"iznos" db:"iznos"`
	Ugrabat   float64 `json:"ugrabat" db:"ugrabat"`
	Rabat     float64 `json:"rabat" db:"rabat"`
	Kasa      float64 `json:"kasa" db:"kasa"`
	Netrezpdv float64 `json:"netrezpdv" db:"netrezpdv"`
	Neto      float64 `json:"neto" db:"neto"`

	// Tab 3 - Nabavka po dobavljačima.
	Fakvred     float64 `json:"fakvred" db:"fakvred"`
	Netofakvred float64 `json:"netofakvred" db:"netofakvred"`
	Ztrovred    float64 `json:"ztrovred" db:"ztrovred"`
	Nabvred     float64 `json:"nabvred" db:"nabvred"`
	Ruc         float64 `json:"ruc" db:"ruc"`
	Procruc     float64 `json:"procruc" db:"procruc"`
	Vpvred      float64 `json:"vpvred" db:"vpvred"`

	// Tab 5 - Izveštaj zaduženja gradilišta VPC-NC: the VP amount (sum of kolic * cena), the NC
	// amount (sum of kolic * ncena) and Razlika = the difference of the two accumulated totals.
	Vpiznos float64 `json:"vpiznos" db:"vpiznos"`
	Nciznos float64 `json:"nciznos" db:"nciznos"`
	Razlika float64 `json:"razlika" db:"razlika"`

	// Tab 4 - Lager lista: one stavka of the promet of an article (the legacy ROB_QRY_LAGER: the magacin,
	// the vrsta and the kodknj of its document, the količina and the prices of the stavka) or one stanje
	// of an article in a magacin (rsta: Ulaz, Izlaz and Cena), with the podgrupa and the serijski broj of
	// the article (SifraArt, Naziv, JM and Gru are the ones of tab 1).
	Mag        int     `json:"mag" db:"mag"`
	MagaciniID int     `json:"magaciniid" db:"magaciniid"`
	Pgru       int     `json:"pgru" db:"pgru"`
	Serbr      string  `json:"serbr" db:"serbr"`
	Kodknj     string  `json:"kodknj" db:"kodknj"`
	Vrd        int     `json:"vrd" db:"vrd"`
	Kolic      float64 `json:"kolic" db:"kolic"`
	Fcena      float64 `json:"fcena" db:"fcena"`
	Ncena      float64 `json:"ncena" db:"ncena"`
	Cena       float64 `json:"cena" db:"cena"`
	// The names of the magacin, of the robna grupa and of the robna podgrupa of the article (the
	// headers and the totals of the lager list printed by grupe and podgrupe).
	MagNaziv  string `json:"magnaziv" db:"magnaziv"`
	GruNaziv  string `json:"grunaziv" db:"grunaziv"`
	PgruNaziv string `json:"pgrunaziv" db:"pgrunaziv"`

	// Tab 4, sub-tab "Ulaz/izlaz za period": the quantities and the fakturna, rabat and magacinska
	// values before the period (do), in the period (Ulaz, Izlaz, Fakvrednost, Rabat, Magvrednost) and up
	// to the end of the period (nadan), the stanje of the article (rsta) and the data of its producer.
	Koldodat         float64 `json:"koldodat" db:"koldodat"`
	Kolnadan         float64 `json:"kolnadan" db:"kolnadan"`
	Stanje           float64 `json:"stanje" db:"stanje"`
	Fakvreddo        float64 `json:"fakvreddo" db:"fakvreddo"`
	Rabatdo          float64 `json:"rabatdo" db:"rabatdo"`
	Magvreddo        float64 `json:"magvreddo" db:"magvreddo"`
	Fakvrednost      float64 `json:"fakvrednost" db:"fakvrednost"`
	Magvrednost      float64 `json:"magvrednost" db:"magvrednost"`
	Fakvrednostnadan float64 `json:"fakvrednostnadan" db:"fakvrednostnadan"`
	Rabatnadan       float64 `json:"rabatnadan" db:"rabatnadan"`
	Magvrednostnadan float64 `json:"magvrednostnadan" db:"magvrednostnadan"`
	Pro              string  `json:"pro" db:"pro"`
	Proizsifra       string  `json:"proizsifra" db:"proizsifra"`
	Komercopis       string  `json:"komercopis" db:"komercopis"`

	// Tab 4, sub-tab "RUC po magacinima": one stavka of a faktura with the rabat of the stavka, the id
	// of the article and the way the magacin keeps its stock (magacini.nacvodzal: 3 = the RUC is the
	// fakturna cena less the cena, otherwise the cena less the nabavna cena).
	Rab       float64 `json:"rab" db:"rab"`
	RsifID    int     `json:"rsifid" db:"rsifid"`
	Nacvodzal int     `json:"nacvodzal" db:"nacvodzal"`

	// Tab 4, sub-tab "Izlazne fakture": the document of the stavka (rdokid, tipdok-nalog, datum naloga,
	// broj, izvorni dokument, datum dokumenta and rok), its kupac (fkto, fana) and its komercijalista
	// (rdok.kom and the ime i prezime).
	RdokID int64        `json:"rdokid" db:"rdokid"`
	Brnal  string       `json:"brnal" db:"brnal"`
	Danal  sql.NullTime `json:"danal" db:"danal"`
	Dokum  string       `json:"dokum" db:"dokum"`
	Dokiz  string       `json:"dokiz" db:"dokiz"`
	Dadok  sql.NullTime `json:"dadok" db:"dadok"`
	Rok    int          `json:"rok" db:"rok"`
	Fkto   string       `json:"fkto" db:"fkto"`
	Fana   string       `json:"fana" db:"fana"`
	Kom    int          `json:"kom" db:"kom"`
	Komerc string       `json:"komerc" db:"komerc"`
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

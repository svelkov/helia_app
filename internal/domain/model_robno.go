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

// RobnoDokumentaDto is one row of the "Unos dokumenta" grid: one robni nalog (rnal) of the
// selected vrsta naloga, with the totals the legacy grid shows per nalog.
type RobnoDokumentaDto struct {
	RnalID     int64         `json:"rnalid" db:"rnalid"`
	Tipdok     string        `json:"tipdok" db:"tipdok"`
	Nalog      int           `json:"nalog" db:"nalog"`
	MagaciniID sql.NullInt64 `json:"magaciniid" db:"magaciniid"`
	Mag        sql.NullInt64 `json:"mag" db:"mag"`
	Danal      sql.NullTime  `json:"danal" db:"danal"`
	Opis       string        `json:"opis" db:"opis"`
	Datob      sql.NullTime  `json:"datob" db:"datob"`
	Brdo       int           `json:"brdo" db:"brdo"`
	Brst       int           `json:"brst" db:"brst"`
	Dug        float64       `json:"dug" db:"dug"`
	Pot        float64       `json:"pot" db:"pot"`
}

// RobnoDokumentaTotalsDto is the aggregate row behind the "Prikaz ukupne obrade" panel.
type RobnoDokumentaTotalsDto struct {
	UkNaloga     int     `json:"uknaloga" db:"uknaloga"`
	UkDokumenata int     `json:"ukdokumenata" db:"ukdokumenata"`
	UkStavki     int     `json:"ukstavki" db:"ukstavki"`
	UkDuguje     float64 `json:"ukduguje" db:"ukduguje"`
	UkPotrazuje  float64 `json:"ukpotrazuje" db:"ukpotrazuje"`
}

// RobnoDokumentaTotal is the "Prikaz ukupne obrade" of the "Robna dokumenta" option.
type RobnoDokumentaTotal struct {
	UkNaloga     string
	UkDokumenata string
	UkStavki     string
	Duguje       string
	Potrazuje    string
	Saldo        string
}

// RobnoDokumentaParams holds the header of a robni nalog (form of "Unos dokumenta") and the
// filters of the grid of that tab.
type RobnoDokumentaParams struct {
	Tipdok     string `json:"tipdok" form:"tipdok"`
	Vrd        string `json:"vrd" form:"vrd"`
	Nalog      string `json:"nalog" form:"nalog"`
	Danal      string `json:"danal" form:"danal"`
	Datob      string `json:"datob" form:"datob"`
	Opis       string `json:"opis" form:"opis"`
	MagaciniID int    `json:"magaciniid" form:"magaciniid"`
	SearchText string `json:"searchText" form:"query"`
}

// PregledDokumentaParams holds the parameters of the "Pregled dokumenta" tab (the sub-tabs "Štampa"
// and "eFaktura"): the robne dokumente (rdok) of the naloga of the current period filtered by
// magacin, vrsta dokumenta, grupe dokumenata and the range of the dates.
type PregledDokumentaParams struct {
	MagaciniID int    `json:"magaciniid" form:"magaciniid"`
	Vrd        string `json:"vrd" form:"vrd"`
	// OdDanal/DoDanal are the "Od/Do datuma naloga" of the Štampa sub-tab (rnal.danal) and the
	// "Od/Do datuma" of the eFaktura sub-tab (rdok.dadok, the date of the document).
	OdDanal string `json:"oddanal" form:"oddanal"`
	DoDanal string `json:"dodanal" form:"dodanal"`
	// GrupeDokumenata is a comma separated list of the document groups (dokvrsta.grpdok) used by the
	// eFaktura sub-tab.
	GrupeDokumenata string `json:"grupedokumenata" form:"grupedokumenata"`
	// DatumStatusa is the date used by the "Ažuriranje statusa eFaktura" action of the eFaktura tab.
	DatumStatusa string `json:"datumstatusa" form:"datumstatusa"`
	SearchText   string `json:"searchText" form:"query"`
}

// KontiranjeDokumentaParams holds the parameters of the "Kontiranje dokumenata" tab (the sub-tabs
// "Knjiženje dokumenata", "Pregled proknjiženih / neproknjiženih dokumenata" and "Pregled
// proknjiženih / neproknjiženih dokumenata po magacinima"): the robni dokumenti (rdok) of the
// current period filtered by magacin, vrsta naloga, vrsta dokumenta and the ranges of the broj
// naloga, the broj dokumenta and the datum naloga.
type KontiranjeDokumentaParams struct {
	MagaciniID int `json:"magaciniid" form:"magaciniid"`
	// Tipdok is the vrsta naloga za knjiženje (rdok.tipdok) and Vrd the vrsta dokumenta (rdok.vrd).
	Tipdok string `json:"tipdok" form:"tipdok"`
	Vrd    string `json:"vrd" form:"vrd"`
	// OdNaloga/DoNaloga, OdDokum/DoDokum and OdDanal/DoDanal are the ranges of the selection of the
	// documentation ("Od/Do broja naloga", "Od/Do broja dokumenta" and "Od/Do dat. naloga").
	OdNaloga string `json:"odnaloga" form:"odnaloga"`
	DoNaloga string `json:"donaloga" form:"donaloga"`
	OdDokum  string `json:"oddokum" form:"oddokum"`
	DoDokum  string `json:"dodokum" form:"dodokum"`
	OdDanal  string `json:"oddanal" form:"oddanal"`
	DoDanal  string `json:"dodanal" form:"dodanal"`
	// Proknjizen is the state selected with the radio buttons of the last two sub-tabs
	// (rdok.knjige_1 = 'D' proknjižen, anything else neproknjižen).
	Proknjizen string `json:"proknjizen" form:"proknjizen"`
	// OznaciNeproknjizenim is the checkbox "Označi prikazana dokumenta kao neproknjižena..." of the
	// second sub-tab.
	OznaciNeproknjizenim bool   `json:"oznacineproknjizenim" form:"oznacineproknjizenim"`
	SearchText           string `json:"searchText" form:"query"`
}

// PrikazNalogaParams holds the parameters of the "Prikaz naloga" tab: the robni nalozi (rnal) of
// the current period filtered by magacin, the range of the vrste naloga and the range of the broj
// naloga. The filters "Po datumu naloga", "Po datumu obrade" and "Po operateru" are applied only
// when their checkbox is checked (the fields are disabled until then, like the legacy screen).
type PrikazNalogaParams struct {
	MagaciniID int `json:"magaciniid" form:"magaciniid"`
	// OdVrd/DoVrd is the range of the vrste naloga (rnal.tipdok).
	OdVrd string `json:"odvrd" form:"odvrd"`
	DoVrd string `json:"dovrd" form:"dovrd"`
	// OdNaloga/DoNaloga is the range of the broj naloga (rnal.nalog).
	OdNaloga string `json:"odnaloga" form:"odnaloga"`
	DoNaloga string `json:"donaloga" form:"donaloga"`

	// Po datumu naloga (rnal.danal).
	ChkDatumNaloga bool   `json:"chkpodatumunaloga" form:"chkpodatumunaloga"`
	OdDanal        string `json:"oddanal" form:"oddanal"`
	DoDanal        string `json:"dodanal" form:"dodanal"`
	// Po datumu obrade (rnal.datob).
	ChkDatumObrade bool   `json:"chkpodatumuobrade" form:"chkpodatumuobrade"`
	OdDatob        string `json:"oddatob" form:"oddatob"`
	DoDatob        string `json:"dodatob" form:"dodatob"`
	// Po operateru (rnal.oper).
	ChkOperator bool   `json:"chkpooperateru" form:"chkpooperateru"`
	Oper        string `json:"oper" form:"oper"`

	SearchText string `json:"searchText" form:"query"`
}

// PrikazNalogaDto is one row of the grid of the "Prikaz naloga" tab: the robni nalog (rnal) with
// the totals of the documents it was saved with.
type PrikazNalogaDto struct {
	RnalID     int64         `db:"rnalid"`
	Tipdok     string        `db:"tipdok"`
	Nalog      int           `db:"nalog"`
	Danal      sql.NullTime  `db:"danal"`
	Datob      sql.NullTime  `db:"datob"`
	Brdo       int           `db:"brdo"`
	Brst       int           `db:"brst"`
	Dug        float64       `db:"dug"`
	Pot        float64       `db:"pot"`
	Oper       string        `db:"oper"`
	MagaciniID sql.NullInt64 `db:"magaciniid"`
}

// PrikazDokumenataUNaloguDto is one row of the grid of the "Prikaz dokumenata u nalogu" tab: a robni
// dokument (rdok) of the selected robni nalozi with the data of its nalog (the legacy screen shows
// the documents of a nalog with their dates, the number of items and their amount).
type PrikazDokumenataUNaloguDto struct {
	Tipdok string       `db:"tipdok"`
	Nalog  int          `db:"nalog"`
	Danal  sql.NullTime `db:"danal"`
	// Vrd is the vrsta dokumenta (rdok.vrd) and Dokum the broj dokumenta of the document.
	Vrd   sql.NullInt64 `db:"vrd"`
	Dokum sql.NullInt64 `db:"dokum"`
	Dadok sql.NullTime  `db:"dadok"`
	Brst  int           `db:"brst"`
	Iznos float64       `db:"iznos"`
	Datob sql.NullTime  `db:"datob"`
	Oper  string        `db:"oper"`
	// MagaciniID is not shown in the grid: it is selected so that the row carries the magacin of the
	// document (the same row type is used for the print of the tab).
	MagaciniID sql.NullInt64 `db:"magaciniid"`
}

// PrikazDokumenataPooperateruDto is one row of the grid of the "Prikaz dokumenata po operateru" tab:
// a robni dokument (rdok) of the selected nalozi with its operater (the legacy screen shows the same
// documents as the "Prikaz dokumenata u nalogu" tab, grouped by the operater).
type PrikazDokumenataPooperateruDto struct {
	Tipdok string       `db:"tipdok"`
	Nalog  int          `db:"nalog"`
	Danal  sql.NullTime `db:"danal"`
	// Vrd is the vrsta dokumenta (rdok.vrd) and Dokum the broj dokumenta of the document.
	Vrd   sql.NullInt64 `db:"vrd"`
	Dokum sql.NullInt64 `db:"dokum"`
	Dadok sql.NullTime  `db:"dadok"`
	Brst  int           `db:"brst"`
	Iznos float64       `db:"iznos"`
	Datob sql.NullTime  `db:"datob"`
	Oper  string        `db:"oper"`
	// MagaciniID is not shown in the grid: it is selected so that the row carries the magacin of the
	// document (the same row type is used for the print of the tab).
	MagaciniID sql.NullInt64 `db:"magaciniid"`
}

// KontiranjeDokumentaDto is one row of the grids of the "Kontiranje dokumenata" tab: the data of a
// robni dokument (rdok) shown by the legacy screen (the three sub-tabs select different columns and
// filters of the same row).
type KontiranjeDokumentaDto struct {
	Tipdok string       `db:"tipdok"`
	Nalog  int          `db:"nalog"`
	Danal  sql.NullTime `db:"danal"`
	// Vrd is the vrsta dokumenta (dokvrsta.vrd) of the document.
	Vrd   sql.NullInt64 `db:"vrd"`
	Dokum sql.NullInt64 `db:"dokum"`
	Dadok sql.NullTime  `db:"dadok"`
	Iznos float64       `db:"iznos"`
	Opis  string        `db:"opis"`
	// Polje is the "POPDV polje" of the legacy grid (rdok.polje).
	Polje string `db:"polje"`
	// Rok is the "Rok plaćanja" (in days) of the document.
	Rok        sql.NullInt64 `db:"rok"`
	MagaciniID sql.NullInt64 `db:"magaciniid"`
	// Knjige1 is the flag of the posting of the document (rdok.knjige_1: 'D' = proknjižen).
	Knjige1 string `db:"knjige1"`
}

// PrikazUkupneObradeDto is one row of the grid of the "Prikaz ukupne obrade" tab: the totals of the
// robni nalozi (rnal) of one magacin of the current period. Every magacin of the period is shown,
// also the ones without a nalog (their totals are 0), like the legacy screen lists them.
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

// PregledStampaDto is one row of the grid of the "Štampa" sub-tab of "Pregled dokumenta": a robni
// dokument (rdok) with the header of its nalog, the partner and the data of the document.
type PregledStampaDto struct {
	Nalog       int           `db:"nalog"`
	Danal       sql.NullTime  `db:"danal"`
	Datob       sql.NullTime  `db:"datob"`
	Dokum       sql.NullInt64 `db:"dokum"`
	Dadok       sql.NullTime  `db:"dadok"`
	Dop         sql.NullTime  `db:"dop"`
	Datiz       sql.NullTime  `db:"datiz"`
	Dokiz       string        `db:"dokiz"`
	Iznos       float64       `db:"iznos"`
	Fkto        string        `db:"fkto"`
	Fana        string        `db:"fana"`
	Naziv       string        `db:"naziv"`
	Pib         string        `db:"pib"`
	Jbkjs       string        `db:"jbkjs"`
	Adresa      string        `db:"adresa"`
	Mesto       string        `db:"mesto"`
	AvansDokum  sql.NullInt64 `db:"avansdokum"`
	Pornapomena string        `db:"pornapomena"`
	Tkonto      string        `db:"tkonto"`
}

// PregledEFakturaDto is one row of the grid of the "eFaktura" sub-tab of "Pregled dokumenta": a
// robni dokument (rdok) with its status in the eFaktura (SEF) system.
type PregledEFakturaDto struct {
	StatusSalinv       string          `db:"statussalinv"`
	DatumStatSalinv    sql.NullTime    `db:"datumstatsalinv"`
	Komentar           string          `db:"komentar"`
	CirInvoiceID       string          `db:"cirinvoiceid"`
	VatRecordingStatus string          `db:"vatrecordingstatus"`
	DatumStatIndVat    sql.NullTime    `db:"datumstatindvat"`
	MagaciniID         sql.NullInt64   `db:"magaciniid"`
	Nalog              int             `db:"nalog"`
	Danal              sql.NullTime    `db:"danal"`
	Dokum              sql.NullInt64   `db:"dokum"`
	Dadok              sql.NullTime    `db:"dadok"`
	Fkto               string          `db:"fkto"`
	Fana               string          `db:"fana"`
	Naziv              string          `db:"naziv"`
	Iznos              float64         `db:"iznos"`
	Dop                sql.NullTime    `db:"dop"`
	Dokiz              string          `db:"dokiz"`
	Valuta             string          `db:"valuta"`
	Kurs               float64         `db:"kurs"`
	RdokID             int64           `db:"rdokid"`
	SalesInvoiceID     sql.NullFloat64 `db:"salesinvoiceid"`
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

package domain

import "database/sql"

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
	IDPartneri int          `json:"idpartneri" db:"idpartneri"`
	FispID     int          `json:"fispid" db:"fispid"`
	God        int          `json:"god" db:"god"`
	Kar        int          `json:"kar" db:"kar"`
	Konto      string       `json:"konto" db:"konto"`
	Sifra      string       `json:"sifra" db:"sifra"`
	Vk         int          `json:"vk" db:"vk"`
	Mag        int          `json:"mag" db:"mag"`
	Naziv      string       `json:"naziv" db:"naziv"`
	Adresa     string       `json:"adresa" db:"adresa"`
	Mesto      string       `json:"mesto" db:"mesto"`
	Pobro      int          `json:"pobro" db:"pobro"`
	XDatUnosa  sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XOpUnos    string       `json:"xop_unos" db:"xopunos"`
	XDatIzmene sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpIzmene  string       `json:"xop_izmene" db:"xopizmene"`
	PIB        string       `json:"pib" db:"pib"`
	JIB        string       `json:"jib" db:"jib"`
	KontaktOsb string       `json:"kontaktosb" db:"kontaktosb"`
	Ziro       string       `json:"ziro" db:"ziro"`
	TipPDV     int          `json:"tippdv" db:"tippdv"`
	MI         int64        `json:"mi" db:"mi"`
	Ter        int          `json:"ter" db:"ter"`
	Flg        string       `json:"flg" db:"flg"`
	GLN        int64        `json:"gln" db:"gln"`
	Email      string       `json:"email" db:"email"`
}

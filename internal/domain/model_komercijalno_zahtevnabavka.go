package domain

import (
	"database/sql"
	"time"
)

// The "Zahtev za nabavku" screen of the komercijalno (the legacy pr_NarudzbenicaPROIZV): a document of edok (not of a
// nalog) with its stavke in epro, the vrste dokumenata of the modul PRO and the group INR. The header and the stavke
// are entered on one screen, like the fakture of the robno.

// The statuses of a zahtev za nabavku (edok.doksts).
const (
	KomZahtevNepotvrdjen = "10"
	KomZahtevPotvrdjen   = "40"
)

// KomZahtevRow is a zahtev za nabavku as edok keeps it, with the texts of its šifre (KomercijalnoRepository).
type KomZahtevRow struct {
	RdokID          int64
	Vrd             int
	VrstaOpis       string
	Dokum           int
	Dadok           sql.NullTime
	MagaciniID      int
	Mag             int
	MagacinOpis     string
	KomID           sql.NullInt64
	KomSifra        int
	KomIme          string
	Orgjed          sql.NullInt64
	OrgjedText      string
	MestoTroska     sql.NullInt64
	MestoTroskaText string
	Foot            string
	Doksts          string
	Opunos          string
	Knjige1         string
	Knjige4         string
}

// KomZahtevUpis are the columns of edok the save of the header writes (the legacy SnimiEDOK).
type KomZahtevUpis struct {
	RdokID      int64
	Vrd         int
	Dokum       int
	Dadok       time.Time
	MagaciniID  int
	Mag         int
	KomID       sql.NullInt64
	Orgjed      sql.NullInt64
	MestoTroska sql.NullInt64
	Foot        string
}

// KomZahtevStavkaRow is a stavka of a zahtev za nabavku (epro with the naziv and the JM of its artikal, the legacy
// PR_QRY_EPRO).
type KomZahtevStavkaRow struct {
	RproID    int64
	RdokID    int64
	Rbr       int
	Sifra     int
	Naziv     string
	JM        string
	Kolicina  float64
	Cena      float64
	DokumVrd  int
	DokumBroj int
}

// KomZahtevStavkaUpis are the columns of epro the save of a stavka writes (the legacy AddEPRO).
type KomZahtevStavkaUpis struct {
	RproID     int64
	RdokID     int64
	Rbr        int
	Sifra      int
	RsifID     int64
	Po         int
	Cena       float64
	Kolicina   float64
	Mag        int
	MagaciniID int
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
}

// KomArtikal is the artikal of a stavka (rsif) with its nabavna cena (rcene.ncena, 0 without it).
type KomArtikal struct {
	RsifID int64   `json:"-"`
	Sifra  int     `json:"sifra"`
	Naziv  string  `json:"naziv"`
	JM     string  `json:"jm"`
	Po     int     `json:"-"`
	NCena  float64 `json:"ncena"`
}

// KomKomercijalista is a komercijalista of the period: its id, its šifra and its ime i prezime.
type KomKomercijalista struct {
	KomID int64
	Sifra int
	Ime   string
}

// KomZahtevPrikazRow is a row of the grid of the tab "Pregled zahteva za nabavku" (the legacy PuniHDRZAHTEV).
type KomZahtevPrikazRow struct {
	RdokID   int64
	Magacin  string
	Vrd      int
	Dokum    int
	Dadok    sql.NullTime
	Foot     string
	Referent string
	OJ       string
}

// KomZahtevStampa is what the print of a zahtev za nabavku shows (the legacy PR_RPT_ZAHTEV_NABAVKA): the firma, the
// header of the document and its stavke.
type KomZahtevStampa struct {
	Firma         RobnoStampaFakturaFirmaDto
	Dokum         int
	Datum         string
	Magacin       string
	MagacinAdresa string
	OJ            string
	Nepotvrdjen   bool
	Stavke        []KomZahtevStavkaRow
}

// KomZahtevHeaderView is the header of the screen: the saved zahtev or a new one (RdokID 0) with its defaults.
type KomZahtevHeaderView struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Datum           time.Time
	MagaciniID      int
	KomSifra        int
	KomIme          string
	Opunos          string
	Orgjed          sql.NullInt64
	OrgjedText      string
	MestoTroska     sql.NullInt64
	MestoTroskaText string
	Tekst           string
	Potvrdjen       bool
	// Snimljen: the header is saved and locked, the stavke are open (after "Sačuvaj").
	Snimljen bool
}

// KomZahtevDokument is the content of the tab "Zahtev za nabavku": the header with the lists of its combos, the
// next redni broj of a stavka and a message of the screen.
type KomZahtevDokument struct {
	Header   KomZahtevHeaderView
	Vrste    []ComboItem
	Magacini []ComboItem
	NoviRbr  int
	Poruka   string
}

// KomZahtevParams are the fields of the header the screen sends ("Sačuvaj", the legacy BTN_Snimi1).
type KomZahtevParams struct {
	RdokID      int64  `form:"rdokid"`
	Vrd         string `form:"vrd"`
	MagaciniID  string `form:"magaciniid"`
	BrojDokum   string `form:"brdok"`
	Datum       string `form:"datum"`
	KomSifra    string `form:"komsifra"`
	Orgjed      string `form:"idorgjed"`
	MestoTroska string `form:"mestotrid"`
	Tekst       string `form:"foot"`
}

// KomZahtevStavkaParams are the fields of a stavka ("Sačuvaj" of the stavka, the legacy BTN_DODAJDET; a rproid of a
// saved stavka is its change).
type KomZahtevStavkaParams struct {
	RdokID   int64  `form:"rdokid"`
	RproID   int64  `form:"rproid"`
	Sifra    string `form:"sifraartikla"`
	Kolicina string `form:"kolicina"`
}

// KomZahtevStavkaForma is a saved stavka for the form of the stavke ("Izmeni" of its row).
type KomZahtevStavkaForma struct {
	RproID   int64   `json:"rproid"`
	Rbr      int     `json:"rbr"`
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	JM       string  `json:"jm"`
	Kolicina float64 `json:"kolicina"`
	NCena    float64 `json:"ncena"`
}

// KomZahtevPrikazParams are the filters of the tab "Pregled zahteva za nabavku": the vrsta of the first tab, the
// magacin, the status (nepotvrđeni or potvrđeni) and all the magacini of the user.
type KomZahtevPrikazParams struct {
	Vrd         int
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	// Tekst is the text of the search of the main screen: a part of the broj, the napomena, the referent or the OJ.
	Tekst string
}

// KomPregledUI are the ids and the urls of the main screen of a document of the komercijalno (opened from the menu,
// like the knjige of the finansijsko): the filters, the search with "Dodaj" and "Štampa", the list of the documents
// with "Izmeni" and "Briši" of the nepotvrđeni ones, the stavke of the selected one, and the entry dialog of a
// document that "Dodaj" and "Izmeni" open. The ids of the filters are the same on every screen (one of them is shown at
// a time): pvrd, pmagaciniid, pdoksts, psvimagacini and the search query.
type KomPregledUI struct {
	// URL is the main screen (the tab of the title bar), ListaURL the list of the documents with the filters,
	// DialogURL the entry dialog (GET id: a saved document; pvrd and pmagaciniid: a new one).
	URL       string
	ListaURL  string
	DialogURL string
	// RootID is the main screen, ListaID the container of the list, DetaljiID the container of the stavke of the
	// selected document, InfoID the container of the messages and the questions of the main screen.
	RootID    string
	ListaID   string
	DetaljiID string
	InfoID    string
	// RdokID is the hidden id of the selected document "Štampa" sends, SearchID the search input.
	RdokID   string
	SearchID string
	// Tabovi are the tabs of the title bar of a screen with more main screens (each tab opens its own; nil: one tab
	// of URL), BezStatusa hides the filter of the nepotvrđeni and potvrđeni documents.
	Tabovi     []TabItem
	BezStatusa bool
}

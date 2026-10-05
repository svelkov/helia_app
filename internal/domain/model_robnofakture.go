package domain

import (
	"database/sql"
	"time"
)

// The view structs of the "Fakture veleprodaje" screen (RobnoFakture / RobnoFaktureDialog in
// frontend/templates/robno/robnadokumenta.templ). They live here because both the templates and the
// handler of the screen need them: the handler fills them and the templates lay them out, and the
// template package cannot import the handler package (the handler imports the templates).

// RobnoFaktureHeader holds the values of the header of the faktura (the first window). The values are
// the ones the handler reads from the form (all strings, the handler formats the numbers), so the
// template only lays them out.
type RobnoFaktureHeader struct {
	// Snimljen tells whether the header is saved: while it is false the header is editable and the
	// stavke control is disabled, afterwards the other way round.
	Snimljen bool

	// Nalog (the read only strip at the top of both windows).
	VrNal, Nalog, DatumNaloga, UkDokum, UkStavki, Duguje, Potrazuje, IznosDokum, Magacin string

	// Otpremnica / Račun and "Štampaj detalje u stavkama".
	Otpremnica     bool
	StampajDetalje bool
}
type RdokDto struct {
	Vrd             string
	BrojDokum       string    `db:"dokum" form:"dokum"`
	DatumDokum      time.Time `db:"dadok" form:"dadok"`
	IzvorniDokum    string    `db:"dokiz" form:"dokiz"`
	IzvorniDokDatum time.Time `db:"datiz" form:"datiz"`
	Opis            string    `db:"opis" form:"opis"`
	ZaPeriod        bool      `form:"za_period"`
	PeriodOd        time.Time `db:"oddat1" form:"oddat1"`
	PeriodDo        time.Time `db:"dodat1" form:"dodat1"`
	DatumSmanjenja  time.Time `db:"datsmanj" form:"datsmanj"`
	// Podaci o kupcu.
	KupacKonto         string `db:"fkto" form:"fkto"`
	KupacSifra         string `db:"fana" form:"fana"`
	KupacNaziv         string
	KupacAdresa        string
	KupacPostanskiBroj string
	KupacMesto         string
	KupacPib           string
	SistemPdv          string
	// Podaci o načinu otpreme, rokovima, valuti i avansima.
	Valuta           int       `db:"sifval" form:"sifval"`
	Kurs             float64   `db:"kurs" form:"kurs"`
	VaziZa           int       `db:"vaziza" form:"vaziza"`
	BrojRata         int       `db:"brojrate" form:"brojrate"`
	DatumPrveRate    time.Time `db:"datumprrate" form:"datumprrate"`
	IznosFakture     float64   `db:"iznfaktrate" form:"iznfaktrate"`
	IznosPrveRate    float64   `db:"prvarata" form:"prvarata"`
	UsloviPlacanja   string    `db:"pla" form:"pla"`
	UsloviKamate     string    `db:"kam" form:"kam"`
	Rok              int       `db:"rok" form:"rok"`
	Dospeva          time.Time `db:"dospeva" form:"dospeva"`
	BrojUgovora      string    `db:"ugovor" form:"ugovor"`
	Narudzbenica     string    `db:"narudzb" form:"narudzb"`
	ZirpRacun        string    `db:"ziro" form:"ziro"`
	NacinOtpreme     string    `db:"otp" form:"otp"`
	Vozac            string    `db:"vozac" form:"vozac"`
	RegBr            string    `db:"brvozila" form:"brvozila"`
	BrutoTezina      float64   `db:"bttowght" form:"bttowght"`
	NetoTezina       float64   `db:"netwght" form:"netwght"`
	BrojKutija       int       `db:"totboxs" form:"totboxs"`
	BrojPaleta       int       `db:"brpaleta" form:"brpaleta"`
	Paritet          string    `db:"paritet" form:"paritet"`
	IzvozIzjava      string    `db:"izjizv" form:"izjizv"`
	BankaIzvozUplatu string    `db:"sifbank" form:"sifbank"`
	Popdv            string    `db:"popdv" form:"popdv"`
	// Podaci o porezom oslobođenju, nastanku PDV obaveze.
	PorKategorija   string `db:"porkat" form:"porkat"`
	OznakaKat       string `db:"oznpk" form:"oznpk"`
	NapomenaPoresko string
	NastanakPDV     string        `db:"nastanak" form:"nastanak"`
	Orgejed         sql.NullInt64 `db:"idorgjed" form:"idorgjed"`
	MestoTroska     sql.NullInt64 `db:"mtroska" form:"mtroska"`
	MestoIsporuke   sql.NullInt64 `db:"fispid" form:"fispid"`
	Komercijalista  sql.NullInt64 `db:"komid" form:"komid"`
	JavnaNabavka    bool
	JBKJNosioc      string `db:"jbkjsnjn" form:"jbkjsnjn"`
	BrojTendera     string `db:"brtend" form:"brtend"`
}

// The combo boxes of the header: the values are read from the database by the handler and passed
type RobnoFaktureCombos struct {
	CbxMi             []ComboItem
	CbxKomercijalista []ComboItem
	CbxPorKategorija  []ComboItem
	CbxOrgjed         []ComboItem
	CbxMtroska        []ComboItem
	CbxNastanakPDV    []ComboItem
	CbxMestoIsporuke  []ComboItem
	CbxValuta         []ComboItem
	CbxBankaIzvoz     []ComboItem
}

// Finansijsko stanje kupca (realizovano / planirano / reprogramirano and the saldo columns).
type RobnoFaktureFinStanjeKupca struct {
	UkRealiz     float64 `db:"ukrealiz"`
	Placeno      float64 `db:"placeno"`
	Saldo        float64 `db:"saldo"`
	Dospelo      float64 `db:"dospelo"`
	Nedospelo    float64 `db:"nedospelo"`
	UkRealizRob  float64 `db:"ukrealizrob"`
	DospeloRob   float64 `db:"dospelorob"`
	NedospeloRob float64 `db:"nedospeloro"`
	UkDospelo    float64 `db:"ukdospelo"`
	UkNedospelo  float64 `db:"uknedospelo"`
	UkSaldo      float64 `db:"uksaldo"`
}

// RobnoFaktureHeaderView is the whole header of the first window of the "Fakture veleprodaje"
// screen, the one RobnoFaktureHeaderFields lays out. It groups the four parts the screen renders the
// header from: the state of the two collapsible controls (RobnoFaktureHeader), the row of the robni
// dokument the header form is bound to (RdokDto), the combo boxes of its fields
// (RobnoFaktureCombos) and the finansijsko stanje of the kupac shown next to them
// (RobnoFaktureFinStanjeKupca). The template reads every field of the header (e.g. header.KupacNaziv,
// header.Vrd or header.CbxPorKategorija) from this one value, so the header form has a single
// parameter and its fields cannot be mixed with the ones of another part of the screen.
type RobnoFaktureHeaderView struct {
	RobnoFaktureHeader
	RdokDto
	RobnoFaktureCombos
	RobnoFaktureFinStanjeKupca
}

// RobnoFaktureStavka holds the values of the entry form of one stavka (the second window). The first
// group of fields is read only: it repeats the kupac and the podaci o dokumentu of the header.
type RobnoFaktureStavka struct {
	KontoMagacina, KontoKupca, Vrd, Brdok, DatumDokumenta string

	SifraArtikla, SifraArtiklaNaziv  string
	JM, Stanje                       string
	Tarifa, Dani                     string
	Kolicina, Iznos                  string
	ProdajnaCena, MagacinskaCena     string
	Rabat, IznosRabata               string
	NabavnaCena, NetoProdajnaCena    string
	Deklaracije, BarKod, Serija, Rok string

	// "Zadnja cena po kojoj je kupac kupovao ovaj artikal".
	ZadnjaCena, ZadnjiRabat, ZadnjiDatum string
	KoristiPoslednjiRabat                bool
}

// RobnoFaktureButtons groups the buttons of the screen so that the signature of the template stays
// short: the header and the stavke have their own save/delete buttons.
type RobnoFaktureButtons struct {
	Save   Button // "Sačuvaj" the header of the faktura
	Modify Button // "Izmeni" the header (unlocks the header, locks the stavke)
	New    Button // "Novi dok."
	Delete Button // "Briši" the document
	Back   Button // "Nazad"

	SaveStavka   Button
	ModifyStavka Button
	DeleteStavka Button
}

// RobnoFaktureStripField is one cell of the two line read only strip of the "Fakture veleprodaje"
type RobnoFaktureStripField struct {
	ID         string
	Label      string
	Value      string
	WidthClass string
	AlignClass string
}

// RobnoFaktureUI carries the urls and the element ids of the "Fakture veleprodaje" screen. They are
// declared in the handler of the screen (which cannot be imported by the templates) and passed to the
// templates through this struct, so every url and id is written in one place only.
type RobnoFaktureUI struct {
	// HeaderPanelID and StavkePanelID are the two collapsible controls (also used by the script that
	// switches the editable control).
	HeaderPanelID string
	StavkePanelID string

	// ContentID is the container the tab navigation swaps: the target of the buttons that reload the whole screen.
	ContentID string

	// DialogID is the id of the dialog of the screen and DialogStagingID the element of the "Unos
	// dokumenta" tab the dialog is rendered into (see RobnoDokumentaMain).
	DialogID        string
	DialogStagingID string

	// ArtikalSearchURL and PartnerSearchURL are the search popups of the stavka and of the kupac,
	// DokumentSearchURL the one of the broj dokumenta (all of them are opened with the "..." buttons of
	// the form).
	ArtikalSearchURL  string
	PartnerSearchURL  string
	DokumentSearchURL string
}

// RobnoStampaFakturaView is the printed invoice of the "Fakture veleprodaje" screen (the legacy
// WinDev report ROB_RPT_STAMPA_FAKTURA), the model of the template RobnoStampaFaktura in
// frontend/templates/reports/robno/robnadokumenta.templ. The izdavalac (the firm) comes from the
// ReportParameters of the print; this view holds the rest of the document. Every value is already
// formatted by the handler (amounts, dates, quantities), so the template only lays them out. A block
// of the legacy report is printed only when it has data: the avansi, the za naplatu po danima row,
// the avansni PDV, the ambalaža, the poresko oslobođenje, the avans totals, the rate and the
// obaveštenje o umanjenju.
type RobnoStampaFakturaView struct {
	// Izdavalac, as the legacy report composes it: the address (with the address of the magacin), the
	// status of the PDV obveznik, the rešenje APR, the e-mail, PIB, matični broj, šifra delatnosti,
	// telefon, BPG, the tekući računi (ITEM_ZIRO) and the banka (ITEM_Banka). Every value already
	// carries its caption (e.g. "PIB : 123"), like the legacy items.
	FirmaAdresa   string
	FirmaPib      string
	FirmaMbr      string
	FirmaSifdel   string
	FirmaTel      string
	FirmaObveznik string
	FirmaApr      string
	FirmaEmail    string
	FirmaBpg      string
	FirmaZiro     string
	FirmaBanka    string

	// Kupac (ITEM_SIFRAKUPCA .. ITEM_TELKUPCA, ITEM_PAK) and the mesto isporuke (ITEM_MISP, ITEM_MISP1).
	KupacKonto     string
	KupacSifra     string
	KupacNaziv     string
	KupacAdresa    string
	KupacMesto     string
	KupacPak       string
	KupacPib       string
	KupacMbr       string
	KupacTelefon   string
	MestoIsporuke  string
	MestoIsporuke2 string

	// The left block under the kupac: the otpremnica (rdok.dokiz) and the narudžbenica, uslovi and rok plaćanja (in days), the
	// komercijalista (šifra and naziv), the vozač and the registarski broj vozila.
	Otpremnica         string
	Narudzbenica       string
	UsloviPlacanja     string
	RokPlacanja        string
	Komercijalista     string
	KomercijalistaOpis string
	Vozac              string
	RegBrojVozila      string

	// The right block: the title of the document (ITEM_DOKNASLOV, e.g. "RAČUN - OTPREMNICA br."), its
	// broj, the datum prometa dobara, the valuta (the date of the dospeće), the datum dokumenta, the
	// nalog and the mesto izdavanja računa.
	Naslov         string
	BrojDokumenta  string
	DatumPrometa   string
	Valuta         string
	DatumDokumenta string
	Nalog          string
	MestoIzdavanja string

	// The stavke of the document (the body of the report).
	Stavke []RobnoStampaFakturaStavka

	// ITERATION_AVANS: the avansni računi closed with this document.
	Avansi []RobnoStampaFakturaAvans

	// ITERATION_PDV, _SVEGA, _RABAT and _ZANAPLATU: the osnovica and the PDV per poreska stopa and the
	// totals of the document.
	PdvPoStopama []RobnoStampaFakturaPdv
	Svega        string
	Rabat        string
	ZaNaplatu    string

	// ITERATION_ZANAPL_PODANIMA: the amount to pay per date of dospeće, when the stavke have their own
	// rok (rpro.dani): the date of the document plus the days of the stavka.
	ZaNaplatuPoDanima []RobnoStampaFakturaZaNaplatuDan

	// ITERATION_AVPDV: the osnovica and the PDV of the avansi per poreska stopa.
	AvansniPdv []RobnoStampaFakturaPdv

	// ITERATION_AVANSTOT and _AVANSZBIR: the total of the avansi with PDV, the rest to pay and the
	// total with PDV of an avansni račun.
	UkupnoAvans    string
	OstaloZaUplatu string
	UkupnoSaPdv    string

	// ITERATION_RATEHDR/_RATEDET/_RATETOT: the rate of the document and their totals.
	Rate           []RobnoStampaFakturaRata
	RateUkOsnovica string
	RateUkPdv      string
	RateUkSaPdv    string

	// ITERATION_AMBHEDER/_AMBBODY: the evidentna ambalaža.
	Ambalaza []RobnoStampaFakturaAmbalaza

	// ITERATION_PORNAP: the član of the Zakon o PDV the document is exempt by (printed when set) and
	// the amount the exemption applies to.
	OslobodjenjeClan  string
	OslobodjenjeIznos string
	// PoreskaNapomena is the free text poreska napomena of the document (rdok.pornapomena).
	PoreskaNapomena string

	// ITERATION_UMANJENJE: the obaveštenje o umanjenju odbitka prethodnog PDV-a (printed when
	// Umanjenje is set) and the član of the Zakon o PDV it refers to.
	Umanjenje     bool
	UmanjenjeClan string

	// ITERATION_ENDDOC: the napomena of the document (ITEM_ITEM2) and the shipping data.
	Napomena    string
	BrutoTezina string
	NetoTezina  string
	BrojKutija  string
	BrojPaleta  string
}

// RobnoStampaFakturaStavka is one stavka of the printed invoice.
type RobnoStampaFakturaStavka struct {
	Rbr, Sifra, Naziv, Jm, Kolicina, Cena, ProcRabata, Rabat, ProcPdv, Pdv, Iznos string
}

// RobnoStampaFakturaAvans is one avansni račun closed with the printed invoice.
type RobnoStampaFakturaAvans struct {
	Broj, Datum, Osnovica, Stopa, Pdv string
}

// RobnoStampaFakturaZaNaplatuDan is the amount to pay (with PDV) of the stavke that fall due on a date.
type RobnoStampaFakturaZaNaplatuDan struct {
	Datum, Iznos string
}

// RobnoStampaFakturaPdv is the osnovica and the PDV of one poreska stopa.
type RobnoStampaFakturaPdv struct {
	Stopa, Osnovica, Pdv string
}

// RobnoStampaFakturaRata is one rata of the printed invoice.
type RobnoStampaFakturaRata struct {
	Rbr, Datum, Procenat, Osnovica, Pdv, UkSaPdv string
}

// RobnoStampaFakturaAmbalaza is one stavka of the evidentna ambalaža.
type RobnoStampaFakturaAmbalaza struct {
	Rbr, Sifra, Naziv, Jm, Kolicina, Cena, Iznos, Stopa string
}

// RobnoStampaFakturaParams is the selection of the print of the fakture (RobnoStampaFaktura): the
// parameters of the source query of the legacy report ROB_RPT_STAMPA_FAKTURA (ipTipDok, ipGrupa,
// ipOdNaloga/ipDoNaloga, ipODDokumenta/ipDoDokumenta, ipMag and ipodVrd/ipDoVrd). Every parameter
// left empty (0 for the magacin) does not restrict the selection, except the vrsta naloga.
type RobnoStampaFakturaParams struct {
	Tipdok string `json:"tipdok" form:"tipdok"`
	// GrupeDokumenata is the comma separated list of the groups of the vrste dokumenta
	// (dokvrsta.grpdok) of the print.
	GrupeDokumenata string `json:"grupedokumenata" form:"grupedokumenata"`
	OdNaloga        string `json:"odnaloga" form:"odnaloga"`
	DoNaloga        string `json:"donaloga" form:"donaloga"`
	OdDokum         string `json:"oddokum" form:"oddokum"`
	DoDokum         string `json:"dodokum" form:"dodokum"`
	MagaciniID      int    `json:"magaciniid" form:"magaciniid"`
	OdVrd           string `json:"odvrd" form:"odvrd"`
	DoVrd           string `json:"dovrd" form:"dovrd"`

	// The selection of the "Štampa" sub-tab of "Pregled dokumenta": one robni dokument (the selected
	// row of the grid, RdokID) or the documents of the filter of the sub-tab (the vrsta dokumenta and
	// the range of the datum naloga).
	RdokID  int64  `json:"rdokid" form:"rdokid"`
	Vrd     string `json:"vrd" form:"vrd"`
	OdDanal string `json:"oddanal" form:"oddanal"`
	DoDanal string `json:"dodanal" form:"dodanal"`
}

// RobnoStampaFakturaRowDto is one row of the source query of the print of the fakture: one stavka
// (rpro) with the header of its robni dokument (rdok) and the data joined to it (the vrsta
// dokumenta, the kupac, the mesto isporuke, the komercijalista, the banka and the poreska
// kategorija of the oslobođenje).
type RobnoStampaFakturaRowDto struct {
	// Header of the document.
	RdokID      int64        `db:"rdokid"`
	Tipdok      string       `db:"tipdok"`
	Nalog       int64        `db:"nalog"`
	Vrd         int64        `db:"vrd"`
	Dokum       int64        `db:"dokum"`
	Dadok       sql.NullTime `db:"dadok"`
	Datiz       sql.NullTime `db:"datiz"`
	Rok         int64        `db:"rok"`
	Pla         string       `db:"pla"`
	Narudzb     string       `db:"narudzb"`
	Vozac       string       `db:"vozac"`
	Brvozila    string       `db:"brvozila"`
	Foot        string       `db:"foot"`
	Pornapomena string       `db:"pornapomena"`
	Dokiz       string       `db:"dokiz"`
	Zirorac     string       `db:"zirorac"`
	// Profak is the comma separated list of the brojevi of the avansni računi (vrd 172) closed with the
	// document and Posuslnab the tekući račun printed on the document ("", "-1" or "-": the računi of
	// the firm).
	Profak    string  `db:"profak"`
	Posuslnab string  `db:"posuslnab"`
	Bttowght  float64 `db:"bttowght"`
	Netwght   float64 `db:"netwght"`
	Totboxs   int64   `db:"totboxs"`
	Brpaleta  int64   `db:"brpaleta"`
	DokNaslov string  `db:"doknaslov"`
	Grpdok    string  `db:"grpdok"`

	// Kupac (fkpl / partneri), mesto isporuke (fisp), komercijalista and banka.
	KupacKonto        string `db:"kupackonto"`
	KupacSifra        string `db:"kupacsifra"`
	KupacNaziv        string `db:"kupacnaziv"`
	KupacAdresa       string `db:"kupacadresa"`
	KupacMesto        string `db:"kupacmesto"`
	KupacPak          string `db:"kupacpak"`
	KupacPib          string `db:"kupacpib"`
	KupacMbr          string `db:"kupacmbr"`
	KupacTelefon      string `db:"kupactelefon"`
	KomercijalistaSif string `db:"komsifra"`
	KomercijalistaNaz string `db:"komnaziv"`
	Banka             string `db:"banka"`
	OslobodjenjeClan  string `db:"oslobodjenjeclan"`

	// Stavka: the poreska stopa is the stopa (rpor.pp) of the poreska oznaka of the stavka (rpro.po) in
	// force on the date of the document (rpro.pdvpct, never filled, only when the oznaka has none).
	Rbr   int64   `db:"rbr"`
	Sifra int64   `db:"sifra"`
	Naz1  string  `db:"naz1"`
	Jm    string  `db:"jm"`
	Po    int64   `db:"po"`
	Kolic float64 `db:"kolic"`
	Cena  float64 `db:"cena"`
	Rab   float64 `db:"rab"`
	Iznos float64 `db:"iznos"`
	Stopa float64 `db:"stopa"`
	Dani  int64   `db:"dani"`
}

// RobnoStampaFakturaAvansDto is one stavka (rpro) of an avansni račun (vrd 172) closed with a
// faktura: its broj and date, its iznos (the osnovica) and the poreska stopa of its poreska oznaka in
// force on its date.
type RobnoStampaFakturaAvansDto struct {
	Dokum int64        `db:"dokum"`
	Dadok sql.NullTime `db:"dadok"`
	Iznos float64      `db:"iznos"`
	Stopa float64      `db:"stopa"`
}

// RobnoStampaFakturaRataDto is one rata of a faktura (faktrate).
type RobnoStampaFakturaRataDto struct {
	RdokID   int64        `db:"rdokid"`
	Rbrrate  int64        `db:"rbrrate"`
	Datrate  sql.NullTime `db:"datrate"`
	Procenat float64      `db:"procenat"`
	Osnovica float64      `db:"osnovica"`
	Pdv      float64      `db:"pdv"`
}

// RobnoStampaFakturaFirmaDto is the izdavalac of the fakture (fvr) as the print shows it.
type RobnoStampaFakturaFirmaDto struct {
	Naziv  string `db:"naziv"`
	Adresa string `db:"adresa"`
	Pobro  string `db:"pobro"`
	Mesto  string `db:"mesto"`
	Pib    string `db:"pib"`
	Matbr  string `db:"matbr"`
	Sifdel string `db:"sifdel"`
	Tel    string `db:"tel"`
	Email  string `db:"email"`
	Apr    string `db:"apr"`
	Bpg    string `db:"bpg"`
	Tekrac string `db:"tekrac"`
	Obv    bool   `db:"obv"`
	// Brobvpdv is the broj of the PDV obveznik and Banke the tekući računi of the firm (banke without
	// the flag nafakne), "brrac - banka" separated by "; ".
	Brobvpdv string `db:"brobvpdv"`
	Banke    string `db:"banke"`
	Logo     []byte `db:"logo"`
}

package domain

import (
	"database/sql"
	"strings"
	"time"
)

// The view structs of the "Fakture veleprodaje" screen (RobnoFakture / RobnoFaktureDialog in
// frontend/templates/robno/robnadokumentaunos.templ). They live here because both the templates and the
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

	// DialogID is the id of the dialog of the screen (RobnoFaktureDialog).
	DialogID string

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
	// Razduzenje is the razduženje za obaveze (the group RCO, the legacy ROB_RPT_STAMPA_RAZDUZENJA_CO):
	// the način razduženja instead of the rok and the uslovi plaćanja and the stanje of the account of the
	// kupac after the razduženje: its fakture (fpro.kat 1 and 2), its uplate (3 and 4) and the saldo.
	Razduzenje                               bool
	StanjeFakture, StanjeUplate, StanjeSaldo string
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

	// The captions of the valuta and of ZA NAPLATU when the print has its own (the knjižno odobrenje:
	// "Datum smanjenja" and "Za povraćaj"; empty: the captions of the faktura), and the napomena of the
	// knjižno odobrenje on the izmena of the poreska osnovica (printed when set).
	ValutaNaslov     string
	ZaNaplatuNaslov  string
	NapomenaOsnovica bool

	// The avansni račun: the datum of the avansna uplata (rdok.datiz) and the vezni dokument
	// (rdok.dokiz) with its opis (rdok.opis).
	DatumAvansneUplate string
	VezniDokument      string
	VezniDokumentOpis  string

	// UkupnoAvans is printed first on the profaktura (AVANS), FirmaBankeNazivi are the banks of the
	// tekući računi of the firm (ITEM_Banka of the profaktura).
	FirmaBankeNazivi string

	// The knjižno pismo: its tip (sifrazlog-opis), "Odobravamo/Zadužujemo Vas kako sledi", the napomena
	// on the izmena of the poreska osnovica, the naknada in words and, when the document has a ugovoreni
	// rabat or a kasa, the neto with the ugovoreni rabat and the kasa.
	TipKnjiznogPisma string
	OdobravaZaduzuje string
	NapomenaPorez    string
	Slovima          string
	Neto             string
	UgovoreniRabat   string
	Kasa             string

	// Skladište (the maloprodajni račun): the opis, the telefon and the e-mail of the magacin.
	MagacinOpis  string
	MagacinTel   string
	MagacinEmail string

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
	// Barkod and PdvPoJm (the PDV of one unit) are printed by the knjižno odobrenje/zaduženje.
	Barkod, PdvPoJm string
	// Konto and Analitika are the konto and the analitička šifra of the stavka (printed by the zaduženje
	// CO, the poziv na broj of the payment).
	Konto, Analitika string
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
	// Iznos is the osnovica with the PDV (printed by the maloprodajni račun).
	Iznos string
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
	RdokID int64        `db:"rdokid"`
	Tipdok string       `db:"tipdok"`
	Nalog  int64        `db:"nalog"`
	Vrd    int64        `db:"vrd"`
	Dokum  int64        `db:"dokum"`
	Dadok  sql.NullTime `db:"dadok"`
	Datiz  sql.NullTime `db:"datiz"`
	Rok    int64        `db:"rok"`
	Pla    string       `db:"pla"`
	// Mcenap is the maloprodajna cena with the PDV of the stavka (rpro.mcenap), the cena of the
	// maloprodajni račun.
	Mcenap float64 `db:"mcenap"`
	// Opis is the opis of the document (rdok.opis): the avansni račun prints it with its vezni dokument.
	Opis string `db:"opis"`
	// The knjižno odobrenje/zaduženje: the barkod of the artikal, the datum of the smanjenje
	// (rdok.datpro) and the predznak of the vrsta dokumenta ("-": the odobrenje).
	Barkod   string       `db:"barkod"`
	Datpro   sql.NullTime `db:"datpro"`
	Predznak string       `db:"predznak"`
	// The faktura-otpremnica: the magacin of the document with its opis (Skladište) and whether it is a
	// konačni račun (the document closes avansi, avansfakt).
	Mag     int64  `db:"mag"`
	MagOpis string `db:"magopis"`
	Konacni bool   `db:"konacni"`
	// The profaktura: the cena of the stavka (rpro.cena), its tax category (rpro.taxcat), the model of
	// the artikal ("D": the rabat per unit and the akciza), the taksa, the akciza and the marža of the
	// cene of the artikal (rcene) and the decimals of the količina of its jedinica mere (jedmere).
	Rcena      float64 `db:"rcena"`
	Taxcat     string  `db:"taxcat"`
	Model      string  `db:"model"`
	Itaksa     float64 `db:"itaksa"`
	Pakc       float64 `db:"pakc"`
	Iakc       float64 `db:"iakc"`
	Vma        float64 `db:"vma"`
	Brdecimala int64   `db:"brdecimala"`
	// The mesto isporuke of the document (fisp of the kupac and of rdok.mi): naziv, adresa, poštanski
	// broj, mesto and GLN.
	MispNaziv   string `db:"mispnaziv"`
	MispAdresa  string `db:"mispadresa"`
	MispPobro   int64  `db:"misppobro"`
	MispMesto   string `db:"mispmesto"`
	MispGln     string `db:"mispgln"`
	Narudzb     string `db:"narudzb"`
	Vozac       string `db:"vozac"`
	Brvozila    string `db:"brvozila"`
	Foot        string `db:"foot"`
	Pornapomena string `db:"pornapomena"`
	Dokiz       string `db:"dokiz"`
	Zirorac     string `db:"zirorac"`
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
	// The zaduženje CO: the konto and the analitička šifra of the stavka (rpro.fkto, rpro.fana) and the
	// naziv of the artikal (rsif.naziv; the stavka without an artikal, šifra 0, prints rpro.naz1).
	StavkaFkto string `db:"stavkafkto"`
	StavkaFana string `db:"stavkafana"`
	RsifNaziv  string `db:"rsifnaziv"`
	// The razduženje CO: the fakture and the uplate of the account of the kupac (fpro), read with the
	// same row.
	StanjeFakture float64 `db:"stanjefakture"`
	StanjeUplate  float64 `db:"stanjeuplate"`
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
	Regbr  string `db:"regbr"`
	Obv    bool   `db:"obv"`
	// Brobvpdv is the broj of the PDV obveznik and Banke the tekući računi of the firm (banke without
	// the flag nafakne), "brrac - banka" separated by "; ".
	Brobvpdv string `db:"brobvpdv"`
	Banke    string `db:"banke"`
	// BankeRacuni are the tekući računi of the firm alone and BankeNazivi the names of their banks (the
	// profaktura prints them apart).
	BankeRacuni string `db:"bankeracuni"`
	BankeNazivi string `db:"bankenazivi"`
	Logo        []byte `db:"logo"`
}

// RobnoStampaFakturaIzvozView is one printed izvozna faktura (the faktura in a foreign valuta, the
// legacy PR_RPT_FAKTURA_OTPIZV, printed in English): the izdavalac, the kupac with the destination,
// the data of the shipment, the stavke at the cena in the valuta of the document, the totals in the
// valuta and in RSD, the notes, the weights, the izjava of the exporter and the bank instructions.
// Every value is already formatted.
type RobnoStampaFakturaIzvozView struct {
	// Izdavalac (ITEM_Adresa .. ITEM_ZIRO): every value carries its caption, like the legacy items.
	FirmaAdresa string
	FirmaPib    string
	FirmaApr    string
	FirmaMbr    string
	FirmaSifdel string
	FirmaTel    string
	FirmaZiro   string
	// FirmaObveznik is the status of the PDV obveznik (the profaktura).
	FirmaObveznik string

	// Kupac (ITEM_SIFRAKUPCA .. ITEM_TELKUPCA) and the destination (ITEM_MISP, ITEM_MISP1).
	KupacSifra    string
	KupacNaziv    string
	KupacAdresa   string
	KupacMesto    string
	KupacPib      string // the PIB line by the PDV status of the partner (PIB, JBKJS, JMBG, BPG, INDEX)
	KupacMbr      string
	KupacTelefon  string
	Destinacija   string
	Destinacija2  string
	Otpremnica    string // the despatch note: broj and datum of the otpremnica (edok)
	UslovPlacanja string
	Vozac         string
	BrojVozila    string

	// The document: "INVOICE No.: 136-43", the invoice date and the place of issue.
	BrojDokumenta  string
	DatumFakture   string
	MestoIzdavanja string
	// The profaktura: the due date (datum dokumenta + rok) and the sales person (komercijalista).
	DatumDospeca   string
	Komercijalista string

	// Valuta is the oznaka of the valuta of the document (e.g. "EUR").
	Valuta string

	Stavke []RobnoStampaFakturaIzvozStavka

	// The totals in the valuta: TOTAL, DISCOUNT, the SUBTOTAL with the guarantee discount and the
	// discount for payment in advance (only when the document has one of them) and the TOTAL FOR
	// PAYMENT; UkupnoRsd is the total for payment in RSD (at the kurs of the document).
	Svega       string
	Rabat       string
	ImaPopuste  bool
	Neto        string
	UgRabatProc string
	UgRabat     string
	KasaProc    string
	Kasa        string
	ZaNaplatu   string
	UkupnoRsd   string

	// The notes (rdok.foot), the delivery terms (paritet), the weights and the izjava of the exporter.
	Napomena    string
	Paritet     string
	BrutoTezina string
	NetoTezina  string
	Izjava      string

	// The bank instructions (bnkizv of rdok.sifbank): 57A the beneficiary's bank and its SWIFT, 59 the
	// account and the beneficiary, 54A the correspondent bank.
	BankaNaziv        string
	BankaSwift        string
	BankaRacun        string
	BankaKorisnik     string
	BankaKorespondent string
}

// RobnoStampaFakturaIzvozStavka is one stavka of the printed izvozna faktura.
type RobnoStampaFakturaIzvozStavka struct {
	Rbr, Sifra, Naziv, Zemlja, Pcn, Jm, Kolicina, Cena, Rabat, Iznos string
	// Popust is the discount of the stavka (the profaktura prints it).
	Popust string
}

// RobnoStampaFakturaIzvozRowDto is one row of the query of the print of the izvozna faktura (the
// legacy PR_QRY_FAKTURE): one stavka (rpro) with the header of its robni dokument (rdok), the kupac,
// the destination, the otpremnica, the valuta and the bank of the document.
type RobnoStampaFakturaIzvozRowDto struct {
	RdokID    int64        `db:"rdokid"`
	Vrd       int64        `db:"vrd"`
	Dokum     int64        `db:"dokum"`
	Dadok     sql.NullTime `db:"dadok"`
	Fkto      string       `db:"fkto"`
	Fana      string       `db:"fana"`
	Pla       string       `db:"pla"`
	Vozac     string       `db:"vozac"`
	Brvozila  string       `db:"brvozila"`
	Foot      string       `db:"foot"`
	Kurs      float64      `db:"kurs"`
	Pkase     float64      `db:"pkase"`
	Ugrabat   float64      `db:"ugrabat"`
	Zirorac   string       `db:"zirorac"`
	Posuslnab string       `db:"posuslnab"`
	Bttowght  float64      `db:"bttowght"`
	Netwght   float64      `db:"netwght"`
	Paritet   string       `db:"paritet"`
	Izjizv    int64        `db:"izjizv"`
	Rok       int64        `db:"rok"`
	Dokiz     string       `db:"dokiz"`
	KomSifra  int64        `db:"komsifra"`
	KomNaziv  string       `db:"komnaziv"`
	Valuta    string       `db:"valuta"`
	Brotp     int64        `db:"brotp"`
	Datotp    sql.NullTime `db:"datotp"`

	KupacNaziv     string `db:"kupacnaziv"`
	KupacAdresa    string `db:"kupacadresa"`
	KupacPobro     int64  `db:"kupacpobro"`
	KupacMesto     string `db:"kupacmesto"`
	KupacTipPdv    int64  `db:"kupactippdv"`
	KupacTer       int64  `db:"kupacter"`
	KupacPib       string `db:"kupacpib"`
	KupacBudzetski bool   `db:"kupacbudzetski"`
	KupacJbkjs     string `db:"kupacjbkjs"`
	KupacJmbg      string `db:"kupacjmbg"`
	KupacBpg       string `db:"kupacbpg"`
	KupacIndex     string `db:"kupacindex"`
	KupacMbr       string `db:"kupacmbr"`
	KupacTelefon   string `db:"kupactelefon"`

	MispNaziv  string `db:"mispnaziv"`
	MispAdresa string `db:"mispadresa"`
	MispPobro  int64  `db:"misppobro"`
	MispMesto  string `db:"mispmesto"`
	MispGln    string `db:"mispgln"`

	BankaNaziv        string `db:"bankanaziv"`
	BankaSwift        string `db:"bankaswift"`
	BankaRacun        string `db:"bankaracun"`
	BankaKorisnik     string `db:"bankakorisnik"`
	BankaKorespondent string `db:"bankakorespondent"`

	Rbr        int64   `db:"rbr"`
	Sifra      int64   `db:"sifra"`
	Naziv      string  `db:"naziv"`
	Komercopis string  `db:"komercopis"`
	Zemlja     string  `db:"zemlja"`
	Pcn        string  `db:"pcn"`
	Jm         string  `db:"jm"`
	Kolic      float64 `db:"kolic"`
	Cenaval    float64 `db:"cenaval"`
	Rab        float64 `db:"rab"`
}

// RobnoStampaKnjiznoPismoRowDto is one row of the query of the print of the knjižno pismo (the legacy
// QRY_RPRO_ZADOK): one stavka (rpro) of the knjižno pismo with the header of its robni dokument (rdok),
// the kupac, the komercijalista, the tip of the knjižno pismo, the vezni dokument (rdok.vrdokid), the
// mesto isporuke, the magacin and the stavka it corrects (the stavka with rpro.rproid1 = its rproid).
type RobnoStampaKnjiznoPismoRowDto struct {
	RdokID      int64        `db:"rdokid"`
	Tipdok      string       `db:"tipdok"`
	Nalog       int64        `db:"nalog"`
	Vrd         int64        `db:"vrd"`
	Dokum       int64        `db:"dokum"`
	Dadok       sql.NullTime `db:"dadok"`
	Fkto        string       `db:"fkto"`
	Fana        string       `db:"fana"`
	Kom         int64        `db:"kom"`
	KomNaziv    string       `db:"komnaziv"`
	Pornapomena string       `db:"pornapomena"`
	Foot        string       `db:"foot"`
	Ugrabat     float64      `db:"ugrabat"`
	Pkase       float64      `db:"pkase"`
	VrdokID     int64        `db:"vrdokid"`
	TipKnjPisma string       `db:"tipknjpisma"`
	// The finansijsko knjižno pismo (without stavke): the oznaka of the vrsta dokumenta (KNO the
	// odobrenje, KNZ the zaduženje), the iznos (the poreska osnovica) and the PDV of the document.
	Dokozn string  `db:"dokozn"`
	Iznos  float64 `db:"iznos"`
	Vporez float64 `db:"vporez"`
	// DomacaValuta is the domestic valuta of the firm (fvr.sifval, the legacy nFVRSIFVAL): the naknada
	// of the finansijsko knjižno pismo is given in words only in dinari (941; rdok.pkase holds its stopa).
	DomacaValuta int64        `db:"domacavaluta"`
	VezniVrd     int64        `db:"veznivrd"`
	VezniDokum   int64        `db:"veznidokum"`
	VezniDadok   sql.NullTime `db:"veznidadok"`
	MagMesto     string       `db:"magmesto"`
	MagAdresa    string       `db:"magadresa"`
	// The račun of the interna zaključnica (ROB_RPT_STAMPA_INTRACFKT): the rok plaćanja, the uslovi
	// plaćanja and the otpremnica of the document.
	Rok   int64  `db:"rok"`
	Pla   string `db:"pla"`
	Dokiz string `db:"dokiz"`

	KupacNaziv     string `db:"kupacnaziv"`
	KupacAdresa    string `db:"kupacadresa"`
	KupacPobro     int64  `db:"kupacpobro"`
	KupacMesto     string `db:"kupacmesto"`
	KupacTipPdv    int64  `db:"kupactippdv"`
	KupacPib       string `db:"kupacpib"`
	KupacBudzetski bool   `db:"kupacbudzetski"`
	KupacJbkjs     string `db:"kupacjbkjs"`
	KupacJmbg      string `db:"kupacjmbg"`
	KupacBpg       string `db:"kupacbpg"`
	KupacIndex     string `db:"kupacindex"`
	KupacTelefon   string `db:"kupactelefon"`

	MispNaziv  string `db:"mispnaziv"`
	MispAdresa string `db:"mispadresa"`
	MispPobro  int64  `db:"misppobro"`
	MispMesto  string `db:"mispmesto"`
	MispGln    string `db:"mispgln"`

	Rbr   int64   `db:"rbr"`
	Sifra int64   `db:"sifra"`
	Naziv string  `db:"naziv"`
	Jm    string  `db:"jm"`
	Kolic float64 `db:"kolic"`
	Fcena float64 `db:"fcena"`
	Rab   float64 `db:"rab"`
	Stopa float64 `db:"stopa"`
	// The stavka the knjižno pismo corrects (rpro.rproid1 = the rproid of this stavka), when there is one.
	IspravkaFound bool    `db:"ispravkafound"`
	IspravkaKolic float64 `db:"ispravkakolic"`
	IspravkaFcena float64 `db:"ispravkafcena"`
	IspravkaRab   float64 `db:"ispravkarab"`
}

// RobnoStampaKnjiznoPismoFakturaDto is one stavka of the faktura a knjižno pismo corrects (its
// rdok.vrdokid) with the ugovoreni rabat and the kasa of the faktura and the poreska stopa of the stavka.
type RobnoStampaKnjiznoPismoFakturaDto struct {
	RdokID  int64   `db:"rdokid"`
	Kolic   float64 `db:"kolic"`
	Fcena   float64 `db:"fcena"`
	Rab     float64 `db:"rab"`
	Stopa   float64 `db:"stopa"`
	Ugrabat float64 `db:"ugrabat"`
	Pkase   float64 `db:"pkase"`
}

// RobnoStampaInterniPrenosView is one printed interni prenos proizvodnje (the group PPR of the vrste
// dokumenta, the legacy ROB_RPT_INTRAC_PROIZV): the header (the opis of the vrsta dokumenta with the
// broj, the date, the nalog, the objekat the goods come from and the one they go to), the stavke on two
// lines (the cene, the vrednosti), the totals of the document and the porezi per tarifa. Every value is
// already formatted.
type RobnoStampaInterniPrenosView struct {
	RdokID             int64  `json:"-"`
	Naslov             string // the opis of the vrsta dokumenta, e.g. "INT.RAC.PROIZ.:"
	BrojDokumenta      string // vrsta-broj
	DatumDokumenta     string
	Nalog              string // tipdok/nalog
	MagacinIzlaz       string // mag-opis of the magacin the goods come from
	MagacinIzlazAdresa string // its adresa and mesto
	ObjekatPrijem      string // the objekat the goods go to: rdok.pkto rdok.pana
	ObjekatPrijemNaziv string // its naziv (fkpl)
	// Interna is the interna zaključnica (the group IRT, the legacy ROB_RPT_STAMPA_INTRAC): the VP cena
	// and the veleprodajni rabat, the mesto of the firm without the godina.
	Interna bool

	Stavke []RobnoStampaInterniPrenosStavka
	Ukupno RobnoStampaInterniPrenosStavka // the totals of the vrednosti

	Porezi       []RobnoStampaInterniPrenosPorez
	UkupnoPorezi RobnoStampaInterniPrenosPorez
}

// RobnoStampaInterniPrenosStavka is one stavka of the printed interni prenos proizvodnje: the first line
// holds the cene (the proizvodna cena, the % rabata, the nabavna cena, the % marže, the maloprodajna
// cena bez poreza, the taksa, the poreska tarifa and stopa and the maloprodajna cena sa porezom), the
// second line the vrednosti.
type RobnoStampaInterniPrenosStavka struct {
	Rbr, Sifra, Naziv, Jm, Kolicina                                        string
	Cena, Rabat, NabavnaCena, Marza, McBezPoreza, Taksa, Tarifa, Stopa, Mc string

	Vrednost, RabatIznos, NabavnaVrednost, MarzaVrednost, McBezPorezaVrednost string
	TaksaVrednost, PorezVrednost, McVrednost                                  string
	// Barkod is the barkod of the artikal (the interna zaključnica prints it under the naziv).
	Barkod string
}

// RobnoStampaInterniPrenosPorez is one porez of the printed interni prenos proizvodnje (the legacy
// ITERATION_PORBODY): the šifra of the porez, the tarifa, the poreska osnovica, the stopa, the porez and
// the maloprodajni iznos.
type RobnoStampaInterniPrenosPorez struct {
	Sp, Tarifa, Osnovica, Stopa, Pdv, MaloprodajniIznos string
}

// RobnoStampaNivelacijaMPView is one printed nivelacija maloprodaje (the legacy
// RPT_NIVELACIJA_MALOPRODAJE): the izdavalac, the broj and the date of the nivelacija, the prodavnica
// (rdok.pkto and rdok.pana with its naziv), the stavke on two lines (the cene and the porezi per unit,
// the vrednosti), their totals and the porezi per nova poreska oznaka. Every value is already formatted.
type RobnoStampaNivelacijaMPView struct {
	FirmaNaziv, FirmaAdresa, FirmaPib, FirmaMbr, FirmaRegbr string

	BrojDokumenta   string // vrsta-broj
	DatumNivelacije string
	DatumStampe     string
	Prodavnica      string // pkto-pana
	ProdavnicaNaziv string

	Stavke []RobnoStampaNivelacijaMPStavka
	Ukupno RobnoStampaNivelacijaMPStavka // the totals of the vrednosti
	Porezi []RobnoStampaNivelacijaMPPorez
}

// RobnoStampaNivelacijaMPStavka is one stavka of the printed nivelacija maloprodaje: the stara and the
// nova MP cena, their razlika, the stara and the nova poreska oznaka with the stopa and the tarifa, the
// stari and the novi porez per unit, their razlika and the RUC per unit, each with its vrednost.
type RobnoStampaNivelacijaMPStavka struct {
	Rbr, Naziv, Kolicina string

	StaraCena, StaraVrednost, NovaCena, NovaVrednost, Razlika, RazlikaVrednost string

	StaraOznaka, StaraStopa, StariPorez, StariPorezVrednost string
	NovaOznaka, NovaStopa, NoviPorez, NoviPorezVrednost     string

	RazlikaPoreza, RazlikaPorezaVrednost, Ruc, RucVrednost string
}

// RobnoStampaNivelacijaMPPorez is the total of the printed nivelacija maloprodaje of one nova poreska
// oznaka: the stara and the nova osnovica (the vrednosti at the MP cene), the stopa of the porez out of
// the MP cena and the stari, the novi porez and their razlika.
type RobnoStampaNivelacijaMPPorez struct {
	Sp, Tarifa, StaraOsnovica, NovaOsnovica, Stopa, StariPorez, NoviPorez, RazlikaPoreza string
}

// RobnoStampaInternaZakljucnicaRacunView is one printed račun - otpremnica of an interna zaključnica (the
// legacy ROB_RPT_STAMPA_INTRACFKT): the header, the PDV per stopa and the totals of a faktura (Faktura)
// and the stavke and the porezi of the interna zaključnica (Prenos).
type RobnoStampaInternaZakljucnicaRacunView struct {
	Faktura RobnoStampaFakturaView
	Prenos  RobnoStampaInterniPrenosView
}

// RobnoStampaInterniPrenosRowDto is one row of the query of the print of the interni prenos proizvodnje
// (the legacy ROB_QRY_IRTSTAMPA): one stavka (rpro) with the header of its robni dokument (rdok), the
// artikal (rsif), its cene (rcene), the magacin and the objekat the goods go to.
type RobnoStampaInterniPrenosRowDto struct {
	RdokID       int64        `db:"rdokid"`
	Vrd          int64        `db:"vrd"`
	Dokum        int64        `db:"dokum"`
	Dadok        sql.NullTime `db:"dadok"`
	Tipdok       string       `db:"tipdok"`
	Nalog        int64        `db:"nalog"`
	VrdOpis      string       `db:"vrdopis"`
	MagOznaka    string       `db:"magoznaka"`
	MagAdresa    string       `db:"magadresa"`
	Pkto         string       `db:"pkto"`
	Pana         string       `db:"pana"`
	PrijemNaziv  string       `db:"prijemnaziv"`
	PrijemNadjen bool         `db:"prijemnadjen"`

	Rbr    int64   `db:"rbr"`
	Sifra  int64   `db:"sifra"`
	Naziv  string  `db:"naziv"`
	Jm     string  `db:"jm"`
	Model  string  `db:"model"`
	Kolic  float64 `db:"kolic"`
	Cena   float64 `db:"cena"`
	Mcenap float64 `db:"mcenap"`
	// Dani is the nova poreska oznaka of a stavka of the nivelacija maloprodaje (rpro.dani).
	Dani        int64        `db:"dani"`
	Vra         float64      `db:"vra"`
	Po          int64        `db:"po"`
	StavkaDadok sql.NullTime `db:"stavkadadok"`
	Itaksa      float64      `db:"itaksa"`
	Pakc        float64      `db:"pakc"`
	Iakc        float64      `db:"iakc"`
	Vma         float64      `db:"vma"`
	Mma         float64      `db:"mma"`
	// The interna zaključnica: the barkod and the grupa of the artikal, the otk, the serija and the rok
	// trajanja of the stavka, the tip zaliha of the magacin and the captions of the otk, the serija and the
	// rok trajanja of the firm (rvr).
	Barkod    string `db:"barkod"`
	Gru       int64  `db:"gru"`
	Otk       string `db:"otk"`
	Serija    string `db:"serija"`
	Roktr     int64  `db:"roktr"`
	Tipzal    int64  `db:"tipzal"`
	OtkLbl    string `db:"otklbl"`
	SerijaLbl string `db:"serijalbl"`
	RokLbl    string `db:"roklbl"`
}

// RobnoStampaPoreskaStopaDto is one poreska stopa (rpor): the poreska oznaka, the tip of the porez, the
// tarifa, the stopa and the date from which it applies.
type RobnoStampaPoreskaStopaDto struct {
	Po    int64        `db:"po"`
	Tip   int64        `db:"tip"`
	Pt    string       `db:"pt"`
	Pp    float64      `db:"pp"`
	Datum sql.NullTime `db:"datum"`
}

// RobnoStampaPopisView is the printed popis (vrsta dokumenta 101), the model of the template
// RobnoStampaPopis in frontend/templates/reports/robno/robnadokumenta.templ: the header of the
// document (its broj and date, the vrsta dokumenta, the nalog, the magacin, the organizaciona jedinica
// and the mesto troška and the ones the goods go to) and its stavke with the totals of the nabavna
// vrednost and of the iznos. Every value is already formatted.
type RobnoStampaPopisView struct {
	BrojDokumenta     string // vrsta-broj, e.g. "101-1"
	DatumDokumenta    string
	VrstaDokumenta    string // vrsta-opis, e.g. "101-POPIS"
	Nalog             string // tipdok-nalog, e.g. "00-1"
	Magacin           string // mag-opis, e.g. "1-veleprodaja"
	Oj                string
	MestoTroska       string
	OjPrijem          string // the organizaciona jedinica the goods go to
	MestoTroskaPrijem string // the mesto troška the goods go to
	// Grupa is the group of the vrsta dokumenta (dokvrsta.grpdok): the opšti dokumenti print their
	// stavke by it (FIN only the iznos, KOL only the količina, the others like the popis).
	Grupa string
	// IzvorniDokument is the izvorni dokument of an opšti dokument (rdok.dokiz).
	IzvorniDokument string
	// MagacinPrijem is the magacin the goods go to (the prenosnica, rdok.magid1: mag-opis).
	MagacinPrijem string
	// The zaduženje sitnog inventara: the konto of the radnik (rdok.pkto with the naziv of its sintetički
	// konto) and the radnik (rdok.pana with the naziv of its analitički konto), and the total of the cene.
	KontoRadnika string
	SifraRadnika string
	UkupnoCena   string
	// The zaduženje gradilišta: the gradilište (rdok.pkto rdok.pana) and its naziv (fkpl).
	Gradiliste      string
	GradilisteNaziv string
	// The nivelacija cena: its napomena (rdok.opis), the totals of the stara and of the nova vrednost
	// and of the nivelacija, and, when the print has more nalozi, the nalog of the document and (on the
	// last document of a nalog) the total of the nivelacija of the nalog.
	Napomena            string
	UkupnoStaraVrednost string
	UkupnoNovaVrednost  string
	UkupnoRazlika       string
	NalogNivelacije     string
	UkupnoZaNalog       string
	// RazlikaIznos is the total of the nivelacija of the document as a number (for the total of the
	// nalog).
	RazlikaIznos float64 `json:"-"`

	Stavke                []RobnoStampaPopisStavka
	UkupnoNabavnaVrednost string
	UkupnoIznos           string
	UkupnoKolicina        string
}

// RobnoStampaPopisStavka is one stavka of the printed popis.
type RobnoStampaPopisStavka struct {
	Rbr, Konto, Sifra, Naziv, Jm, Kolicina, NabavnaCena, NabavnaVrednost, Cena, Iznos string
	// Otk, Serija and Rok (the rok trajanja) are printed by the popis tekuće godine.
	Otk, Serija, Rok string
	// The nivelacija cena: the stara cena and vrednost, the procenat of the change and the vrednost of
	// the nivelacija (the nova cena and vrednost are Cena and Iznos).
	StaraCena, StaraVrednost, Procenat, Razlika string
}

// RobnoStampaPopisRowDto is one row of the query of the print of the popis: one stavka (rpro) with the
// header of its robni dokument (rdok) and the names joined to it.
type RobnoStampaPopisRowDto struct {
	RdokID int64  `db:"rdokid"`
	Grpdok string `db:"grpdok"`
	Dokiz  string `db:"dokiz"`
	// The prenosnica: the magacin the goods go to (rdok.magid1, mag-opis), the tip zaliha of the magacin of
	// the document (magacini.tipzal: 1 without serije), the serija and the rok trajanja of the stavka
	// (rpro.serija, rpro.roktr, a WinDev integer date), the naziv, the komercijalni opis, the
	// proizvođač and its šifra of the artikal (rsif) and the decimals of its jedinica mere (jedmere).
	MagPrijem       string  `db:"magprijem"`
	Tipzal          int64   `db:"tipzal"`
	Serija          string  `db:"serija"`
	Otk             string  `db:"otk"`
	Nalog1          int64   `db:"nalog1"`
	KontoRadnik     string  `db:"kontoradnik"`
	SifraRadnik     string  `db:"sifraradnik"`
	Gradiliste      string  `db:"gradiliste"`
	GradilisteNaziv string  `db:"gradilistenaziv"`
	RnalID          int64   `db:"rnalid"`
	Fcena           float64 `db:"fcena"`
	// Vma is the procenat of the nivelacija cena (rpro.vma).
	Vma               float64      `db:"vma"`
	Opis              string       `db:"opis"`
	Roktr             int64        `db:"roktr"`
	RsifNaziv         string       `db:"rsifnaziv"`
	Komercopis        string       `db:"komercopis"`
	Pro               string       `db:"pro"`
	Proizsifra        string       `db:"proizsifra"`
	RsifJm            string       `db:"rsifjm"`
	Brdecimala        int64        `db:"brdecimala"`
	Tipdok            string       `db:"tipdok"`
	Nalog             int64        `db:"nalog"`
	Vrd               int64        `db:"vrd"`
	VrdOpis           string       `db:"vrdopis"`
	Dokum             int64        `db:"dokum"`
	Dadok             sql.NullTime `db:"dadok"`
	Mag               int64        `db:"mag"`
	MagNaziv          string       `db:"magnaziv"`
	Oj                string       `db:"oj"`
	MestoTroska       string       `db:"mt"`
	OjPrijem          string       `db:"ojprijem"`
	MestoTroskaPrijem string       `db:"mtprijem"`
	Rbr               int64        `db:"rbr"`
	Konto             string       `db:"konto"`
	Sifra             int64        `db:"sifra"`
	Naz1              string       `db:"naz1"`
	Jm                string       `db:"jm"`
	Kolic             float64      `db:"kolic"`
	Ncena             float64      `db:"ncena"`
	Cena              float64      `db:"cena"`
	Iznos             float64      `db:"iznos"`
}

// RobnoStampaDokumentDto is what the print of the "Štampa" sub-tab of "Pregled dokumenta" needs to
// choose the print of a robni dokument, like the legacy print button (the SWITCH on the group of the
// vrsta dokumenta, DOKVRSTA.GRPDOK): the vrsta and its group, the oznaka (DOKOZN) and the suffix of
// the custom report of the vrsta (DODOZNFAK, the legacy report "<report>_<dodoznfak>"), the valuta of
// the document and the domestic valuta of the firm (FVR.SIFVAL, the legacy nFVRSIFVAL), the knjige of
// the document (kontiran when it starts with "D") and whether the document has stavke (rpro).
type RobnoStampaDokumentDto struct {
	RdokID       int64  `db:"rdokid"`
	Vrd          int64  `db:"vrd"`
	Grpdok       string `db:"grpdok"`
	Dokozn       string `db:"dokozn"`
	Dodoznfak    string `db:"dodoznfak"`
	Sifval       int64  `db:"sifval"`
	DomacaValuta int64  `db:"domacavaluta"`
	Knjige       string `db:"knjige"`
	ImaStavke    bool   `db:"imastavke"`
}

// Devizni reports whether the document is in a foreign valuta (the legacy RDOK.SIFVAL > 0 AND
// RDOK.SIFVAL <> nFVRSIFVAL).
func (d RobnoStampaDokumentDto) Devizni() bool {
	return d.Sifval > 0 && d.Sifval != d.DomacaValuta
}

// Kontiran reports whether the document is kontiran (the legacy RDOK.KNJIGE[1] = "D", the column
// rdok.knjige_1).
func (d RobnoStampaDokumentDto) Kontiran() bool {
	return strings.HasPrefix(strings.ToUpper(d.Knjige), "D")
}

// RobnoStampaKalkulacijaView is one printed kalkulacija veleprodaje (prijemni list, the group PLT of the
// vrste dokumenta, the legacy RPT_ROB_KALKULACIJA): the header of the document with the dobavljač, its
// stavke (two lines per artikal: the cene and the vrednosti) with the totals, the porezi of the
// document of the dobavljač by poreska tarifa (a domestic document) and the valuta with the kurs (a
// document in a foreign valuta, Devizni). Every value is already formatted.
type RobnoStampaKalkulacijaView struct {
	BrojDokumenta string // vrsta-broj, e.g. "110-285"
	Devizni       bool   // the document is in a foreign valuta (the print of the foreign dobavljači)

	KontoDobavljaca string
	SifraDobavljaca string
	DobavljacNaziv  string
	DobavljacAdresa string
	DobavljacMesto  string // poštanski broj and mesto
	DobavljacPib    string

	Magacin                  string // mag-opis, e.g. "1-VELEPRODAJA"
	Nalog                    string // tipdok/nalog, e.g. "20/285"
	DatumDokumenta           string
	NacinDopremanja          string
	DokumentDobavljaca       string
	DatumDokumentaDobavljaca string
	NacinPlacanja            string
	RokPlacanja              string
	Napomena                 string

	Valuta string // sifval and oznaka, e.g. "978 EUR"
	Oznaka string // oznaka of the valuta, e.g. "EUR"
	Kurs   string

	Stavke []RobnoStampaKalkulacijaStavka
	Ukupno RobnoStampaKalkulacijaStavka // the totals of the vrednosti

	Porezi       []RobnoStampaKalkulacijaPorez
	UkupnoPorezi RobnoStampaKalkulacijaPorez

	// Maloprodaja is the kalkulacija maloprodaje (the group KAL, the legacy RPT_ROB_KALKULACIJA_MP): the
	// prodavnica instead of the magacin, the data of the firm in the header, the stavke at the
	// maloprodajne cene with the PDV, the porezi of the document of the dobavljač of rppo (Porezi) and
	// the porezi of the kalkulacija (PoreziKalkulacije).
	Maloprodaja       bool
	Prodavnica        string // pkto-pana naziv
	FirmaMesto        string // adresa,mesto
	FirmaPib          string // PIB, šifra delatnosti and matični broj
	FirmaTel          string
	VremeStampe       string
	PoreziKalkulacije []RobnoStampaKalkulacijaPorez
	UkupnoPoreziKalk  RobnoStampaKalkulacijaPorez
}

// RobnoStampaKalkulacijaStavka is one artikal of the printed kalkulacija: the first line holds the
// cene (and the oznake of the troškova), the second line the vrednosti.
type RobnoStampaKalkulacijaStavka struct {
	Rbr, Sifra, Naziv, Jm, Kolicina string

	FakturnaCena, Rabat, NetoCena string
	OznakaTroska, ZavisniTrosak   string
	ZavisniTrosakDobavljaca       string
	OznakaInternog, InterniTrosak string
	NabavnaCena, Marza, VpCena    string

	FakturnaVrednost, VrednostRabata, NetoVrednost   string
	VrednostTroska, VrednostTroskaDobavljaca         string
	VrednostInternog, NabavnaVrednost, VrednostMarze string
	VpVrednost                                       string

	// The kalkulacija maloprodaje: the stopa of the PDV, the maloprodajna cena with the PDV, the PDV and
	// the maloprodajna vrednost with the PDV (Marza is the % of the MP marže, VrednostTroska all the
	// zavisni troškovi of the stavka).
	StopaPdv, MpCena, VrednostPdv, MpVrednost string
}

// RobnoStampaKalkulacijaPorez is one poreska tarifa of the document of the dobavljač ("Po dokumentu
// dobavljača"): the iznos of the document, the osnovica, the obračunat porez, the tarifa and the stopa.
type RobnoStampaKalkulacijaPorez struct {
	IznosDokumenta, Osnovica, Porez, Tarifa, Stopa string
}

// RobnoStampaKalkulacijaRowDto is one row of the query of the print of the kalkulacija: one stavka
// (rpro) with the header of its robni dokument (rdok), the dobavljač, the valuta and the poreska tarifa
// of the stavka.
type RobnoStampaKalkulacijaRowDto struct {
	RdokID        int64        `db:"rdokid"`
	Tipdok        string       `db:"tipdok"`
	Nalog         int64        `db:"nalog"`
	Vrd           int64        `db:"vrd"`
	Dokum         int64        `db:"dokum"`
	Dadok         sql.NullTime `db:"dadok"`
	Datiz         sql.NullTime `db:"datiz"`
	Dokiz         string       `db:"dokiz"`
	Otp           string       `db:"otp"`
	Pla           string       `db:"pla"`
	Rok           int64        `db:"rok"`
	Foot          string       `db:"foot"`
	Mag           int64        `db:"mag"`
	MagNaziv      string       `db:"magnaziv"`
	Fkto          string       `db:"fkto"`
	Fana          string       `db:"fana"`
	PartnerNaziv  string       `db:"partnernaziv"`
	PartnerAdresa string       `db:"partneradresa"`
	PartnerMesto  string       `db:"partnermesto"`
	PartnerPib    string       `db:"partnerpib"`
	Sifval        int64        `db:"sifval"`
	DomacaValuta  int64        `db:"domacavaluta"`
	ValutaOznaka  string       `db:"valutaoznaka"`
	Kurs          float64      `db:"kurs"`

	Rbr    int64   `db:"rbr"`
	Sifra  int64   `db:"sifra"`
	Naz1   string  `db:"naz1"`
	Jm     string  `db:"jm"`
	Kolic  float64 `db:"kolic"`
	Fcena  float64 `db:"fcena"`
	Rab    float64 `db:"rab"`
	Tz     string  `db:"tz"`
	Ztro   float64 `db:"ztro"`
	Ztrof  float64 `db:"ztrof"`
	Tzi    string  `db:"tzi"`
	Ztrin  float64 `db:"ztrin"`
	Ncena  float64 `db:"ncena"`
	Vpcena float64 `db:"vpcena"`
	Po     int64   `db:"po"`
	Tarifa string  `db:"tarifa"`
	Stopa  float64 `db:"stopa"`

	// The kalkulacija maloprodaje: the maloprodajna cena with the PDV (rpro.mcenap) and the prodavnica
	// (rdok.pkto and rdok.pana with the naziv of its fkpl).
	Mcenap            float64 `db:"mcenap"`
	Pkto              string  `db:"pkto"`
	Pana              string  `db:"pana"`
	ProdavnicaNaziv   string  `db:"prodavnicanaziv"`
	ProdavnicaNadjena bool    `db:"prodavnicanadjena"`
	// A porez of the document of the dobavljač (rppo, read with the same row): its osnovica, porez and
	// stopa.
	Osn  float64 `db:"osn"`
	Ppor float64 `db:"ppor"`
	Pdv  float64 `db:"pdv"`
}

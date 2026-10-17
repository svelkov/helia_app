package domain

import (
	"database/sql"
	"time"
)

// The "Otpremnice" screen of the komercijalno (the legacy pr_Otpremnica, "Otpremnica robe/gotovih proizvoda"): an
// otpremnica of a potvrđen nalog za isporuku (vrdokid) in edok of the vrste of the robne fakture (dokvrsta of the moduli
// PRO, SVI and ROB of the oznake FPR, FAK, PRE, FZR and FRP) with its stavke in epro (copied from the nalog; rproid1 is
// the stavka of the nalog). Its statuses: 10 new, 40 potvrđena: its robna faktura (rdok.edokid) is posted on a robni
// nalog. Until then the broj otpremnice and the broj fakture are temporary (from 990001 up).

// The statuses of an otpremnica (edok.doksts) and the first temporary broj (the legacy NoviBrojEDOK: 990001).
const (
	KomOtpNova             = "10"
	KomOtpPotvrdjena       = "40"
	KomOtpPrivremeniBroj   = 990000
	KomOtpPrviPrivremeni   = 990001
	KomOtpNastanakDefault  = "35"
	KomOtpNastanakNeObvez  = "0"
	KomOtpTeritorijaStrani = 3
)

// KomOtpVrsta is a vrsta dokumenta of an otpremnica (dokvrsta): the vrd with its opis, its oznaka (FPR, FAK, ...), its
// predznak ("-" a storno), its knjiženje (D dugovno, P potražno), whether it needs the POPDV polje with its default.
type KomOtpVrsta struct {
	DokvrstaID int64  `db:"dokvrstaid"`
	Vrd        int    `db:"vrd"`
	Opis       string `db:"opis"`
	Dokozn     string `db:"dokozn"`
	Predznak   string `db:"predznak"`
	Kodknj     string `db:"kodknj"`
	FlgPopdv   bool   `db:"flgpopdv"`
	Polje      string `db:"polje"`
}

// Storno tells whether the vrsta is a storno one (predznak "-").
func (v KomOtpVrsta) Storno() bool { return v.Predznak == "-" }

// KomOtpremnicaRow is an otpremnica of the period with the texts of its šifre, its magacin and its nalog.
type KomOtpremnicaRow struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Dadok           sql.NullTime
	Brotp           int
	Datotp          sql.NullTime
	Datiz           sql.NullTime
	MagaciniID      int
	Mag             int
	MagacinText     string
	MagacinAdresa   string
	MagacinMesto    string
	Tipmag          string
	Tipzal          int
	Nacvodzal       int
	Tipdok          string
	TipdokText      string
	Orgjed          sql.NullInt64
	OrgjedText      string
	MestoTroska     sql.NullInt64
	MestoTroskaText string
	Fkto            string
	Fana            string
	KomID           sql.NullInt64
	KomIme          string
	Rok             int
	Pla             string
	Kamt            string
	Mi              int
	MestoIsporuke   string
	Otp             string
	Vozac           string
	Brvozila        string
	Bttowght        float64
	Netwght         float64
	Totboxs         int
	Brpaleta        int
	Foot            string
	Napomena        string
	Opis            string
	SifVal          int
	OznVal          string
	Kurs            float64
	VaziZa          int
	Keykat          string
	Category        string
	Pornapomena     string
	Brugov          string
	Polje           string
	Posuslnab       string
	Ugrabat         float64
	Pkase           float64
	Doksts          string
	Knjige1         string
	VrdokID         int64
	NalogDokum      int
	NalogDadok      sql.NullTime
}

// KomOtpNalog is a potvrđen nalog za isporuku an otpremnica is made of (the legacy BTN_TraziDOK with PodaciHDR): what the
// header shows of it.
type KomOtpNalog struct {
	RdokID          int64        `json:"rdokid"`
	Dokum           int          `json:"dokum"`
	Dadok           sql.NullTime `json:"-"`
	Datum           string       `json:"datum"`
	MagaciniID      int          `json:"-"`
	Doksts          string       `json:"-"`
	Fkto            string       `json:"fkto"`
	Fana            string       `json:"fana"`
	TipdokText      string       `json:"tipdok"`
	OrgjedText      string       `json:"oj"`
	MestoTroskaText string       `json:"mt"`
	Posuslnab       string       `json:"posuslnab"`
	SifVal          int          `json:"sifval"`
	OznVal          string       `json:"oznval"`
	Kurs            float64      `json:"kurs"`
	VaziZa          int          `json:"vaziza"`
	Foot            string       `json:"foot"`
	Napomena        string       `json:"napomena"`
	Vozac           string       `json:"vozac"`
	Brvozila        string       `json:"brvozila"`
	Otp             string       `json:"otp"`
	Bttowght        float64      `json:"bttowght"`
	Netwght         float64      `json:"netwght"`
	Totboxs         int          `json:"totboxs"`
	Brpaleta        int          `json:"brpaleta"`
	Keykat          string       `json:"keykat"`
	Category        string       `json:"category"`
	Pornapomena     string       `json:"pornapomena"`
	Brugov          string       `json:"brugov"`
	Kupac           KomProKupac  `json:"kupac"`
	Ino             bool         `json:"ino"`
}

// KomOtpStavkaRow is a stavka of an otpremnica (or of its nalog) with what its artikal has: the naziv, the JM with its
// decimals and whether it has a length (jedmere.ima_duzinu, ima_sirinu; JmPostoji false without the JM), the komercijalni
// opis, the proizvođač, the tip and the težina.
type KomOtpStavkaRow struct {
	RproID     int64   `db:"rproid"`
	RproID1    int64   `db:"rproid1"`
	Rbr        int     `db:"rbr"`
	Brst       int     `db:"brst"`
	Konto      string  `db:"konto"`
	Sifra      int     `db:"sifra"`
	RsifID     int64   `db:"rsifid"`
	Naziv      string  `db:"naziv"`
	JM         string  `db:"jm"`
	Komercopis string  `db:"komercopis"`
	Pro        string  `db:"pro"`
	Proizsifra string  `db:"proizsifra"`
	Tip        string  `db:"tip"`
	Otk        string  `db:"otk"`
	Serija     string  `db:"serija"`
	Roktr      int64   `db:"roktr"`
	Trazkolic  float64 `db:"trazkolic"`
	Kolic      float64 `db:"kolic"`
	Kolic1     float64 `db:"kolic1"`
	Cena       float64 `db:"cena"`
	Ncena      float64 `db:"ncena"`
	Pcena      float64 `db:"pcena"`
	Fcena      float64 `db:"fcena"`
	CenaVal    float64 `db:"cenaval"`
	Mcenap     float64 `db:"mcenap"`
	Rab        float64 `db:"rab"`
	Po         int     `db:"po"`
	Tezjm      float64 `db:"tezjm"`
	Decimale   int     `db:"decimale"`
	JmPostoji  bool    `db:"jmpostoji"`
	ImaDuzinu  bool    `db:"imaduzinu"`
	ImaSirinu  bool    `db:"imasirinu"`
}

// KomOtpUpis are the columns of a new otpremnica (the legacy SnimiEDOK; the others are the ones of its nalog) and the
// ones a saved one changes.
type KomOtpUpis struct {
	RdokID      int64
	NalogID     int64
	Vrd         int
	Brotp       int
	Dokum       int
	Datotp      time.Time
	Dadok       time.Time
	Otp         string
	Vozac       string
	Brvozila    string
	Bttowght    float64
	Netwght     float64
	Totboxs     int
	Brpaleta    int
	Foot        string
	Napomena    string
	SifVal      int
	Kurs        float64
	VaziZa      int
	Posuslnab   string
	Polje       string
	Keykat      string
	Category    string
	Pornapomena string
	Brugov      string
}

// KomOtpParams are the fields of the header the screen sends ("Sačuvaj" and "Potvrda dok."): the nalog of a new
// otpremnica, its vrsta and magacin, the broj otpremnice and fakture (typed with "Dozvoli izmenu br. OTP i br. Fakture",
// dozvolibroj), the datumi, the kupac's napomena, the poreska kategorija, the valuta, the transport data, the nastanak PDV
// obaveze and the opis of the faktura.
type KomOtpParams struct {
	RdokID      int64  `form:"rdokid"`
	MagaciniID  string `form:"magaciniid"`
	Vrd         string `form:"vrd"`
	NalogID     string `form:"nalogid"`
	DozvoliBroj string `form:"dozvolibroj"`
	Brotp       string `form:"brotp"`
	Dokum       string `form:"brfakture"`
	Datotp      string `form:"datotp"`
	Datum       string `form:"datum"`
	Foot        string `form:"foot"`
	Napomena    string `form:"napomena"`
	Keykat      string `form:"keykat"`
	Category    string `form:"oznakakat"`
	Pornapomena string `form:"napomporeskoosl"`
	Nastanak    string `form:"nastanak"`
	Posuslnab   string `form:"posuslnab"`
	Vozac       string `form:"vozac"`
	Otp         string `form:"otp"`
	SifVal      string `form:"sifval"`
	Kurs        string `form:"kurs"`
	VaziZa      string `form:"vaziza"`
	Polje       string `form:"popdv"`
	Brugov      string `form:"brugov"`
	OpisDok     string `form:"opisdok"`
	Brvozila    string `form:"brvozila"`
	Bttowght    string `form:"bttowght"`
	Netwght     string `form:"netwght"`
	Totboxs     string `form:"totboxs"`
	Brpaleta    string `form:"brpaleta"`
}

// KomOtpStavkaParams are the fields of a changed stavka: the isporučena količina and, with "Dozvoli izmenu OTK, Seriju,
// Rok" (dozvolilot), its lot.
type KomOtpStavkaParams struct {
	RdokID     int64  `form:"rdokid"`
	RproID     int64  `form:"rproid"`
	Kolicina   string `form:"kolicina"`
	DozvoliLot string `form:"dozvolilot"`
	Otk        string `form:"otk"`
	Serija     string `form:"serija"`
	Rok        string `form:"roktr"`
}

// KomOtpStavkaForma is a saved stavka for the form of the stavke ("Izmeni" of its row).
type KomOtpStavkaForma struct {
	RproID    int64   `json:"rproid"`
	Rbr       int     `json:"rbr"`
	Sifra     int     `json:"sifra"`
	Naziv     string  `json:"naziv"`
	JM        string  `json:"jm"`
	Otk       string  `json:"otk"`
	Serija    string  `json:"serija"`
	Rok       string  `json:"rok"`
	Trazkolic float64 `json:"trazkolic"`
	Kolicina  float64 `json:"kolicina"`
}

// KomOtpHeaderView is the header of the entry dialog: the otpremnica with the datumi of its date fields, its kupac, its
// vrsta and whether the kupac is a foreign one (the valuta can change), the firm a PDV obveznik, the header saved.
type KomOtpHeaderView struct {
	KomOtpremnicaRow
	Datum       time.Time
	DatumOtp    time.Time
	Kupac       KomProKupac
	Vrsta       KomOtpVrsta
	Ino         bool
	PdvObveznik bool
	Snimljen    bool
}

// KomOtpDokument is the content of the entry dialog: the header with the lists of its combos (the vrste, the magacini,
// the žiro računi, the nastanak PDV obaveze with the chosen one), the valuta of the firm and a message.
type KomOtpDokument struct {
	Header      KomOtpHeaderView
	Vrste       []ComboItem
	Magacini    []ComboItem
	ZiroRacuni  []ComboItem
	Nastanci    []ComboItem
	Nastanak    string
	SifValFirme int
	Poruka      string
	// Eotp is the e-otpremnica of the "eOtpremnice" screen (nil on the "Otpremnice" screen): its controls follow the
	// header.
	Eotp *KomEotpView
}

// KomOtpPrikazRow is a row of the list of the main screen (the legacy PR_QRY_EDOK of "Pregled otpremnica").
type KomOtpPrikazRow struct {
	RdokID        int64        `db:"rdokid"`
	Magacin       string       `db:"magacin"`
	Vrd           int          `db:"vrd"`
	VrstaOpis     string       `db:"vrstaopis"`
	Dokum         int          `db:"dokum"`
	Dadok         sql.NullTime `db:"dadok"`
	Brotp         int          `db:"brotp"`
	NalogDokum    int          `db:"nalogdokum"`
	Fkto          string       `db:"fkto"`
	Fana          string       `db:"fana"`
	Kupac         string       `db:"kupac"`
	Doksts        string       `db:"doksts"`
	Rok           int          `db:"rok"`
	Datiz         sql.NullTime `db:"datiz"`
	Opis          string       `db:"opis"`
	MestoIsporuke string       `db:"mestoisporuke"`
	Kamt          string       `db:"kamt"`
}

// KomOtpPrikazParams are the filters of the list of the main screen.
type KomOtpPrikazParams struct {
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	Tekst       string
}

// KomOtpRnalUpis is a robni nalog the potvrda of an otpremnica creates (the legacy SnimiRDOK "Aut. kreiran nalog").
type KomOtpRnalUpis struct {
	Tipdok     string
	Danal      time.Time
	MagaciniID int
	Mag        int
}

// KomOtpRdokUpis are the columns of the robna faktura of an otpremnica the potvrda chooses (the legacy SnimiRDOK; the
// others are the ones of the otpremnica and of the nalog): its broj, the broj otpremnice (dokiz), the opis and the
// nastanak PDV obaveze.
type KomOtpRdokUpis struct {
	Dokum           int
	Dokiz           int
	Opis            string
	DescriptionCode int
}

// KomOtpRproUpis is a stavka of the robna faktura of an otpremnica (the legacy SnimiRDOK "Novi slog u RPRO").
type KomOtpRproUpis struct {
	RnalID     int64
	RdokID     int64
	Sifra      int
	RsifID     int64
	Konto      string
	JM         string
	Ncena      float64
	Kolic      float64
	Kolic1     float64
	Iznos      float64
	Cena       float64
	Pcena      float64
	Fcena      float64
	CenaVal    float64
	Mcenap     float64
	Rab        float64
	Po         int
	Brst       int
	Rbr        int
	Otk        string
	Serija     string
	Roktr      int64
	Taxcat     string
	Taxexreco  string
	MagaciniID int
	Mag        int
}

// KomOtpRceneUpis is the cena of an artikal the potvrda adds when it has none (the legacy RCENE of SnimiRDOK).
type KomOtpRceneUpis struct {
	RsifID int64
	Sifra  int
	Ncena  float64
	Rab    float64
	Cena   float64
	Fcena  float64
}

// KomOtpStampa is what the print of an otpremnica shows (the legacy PR_RPT_OTPREMNICA): the firm with its tekući računi,
// the kupac, the header data, the stavke with količina other than 0 and the totals; Ino prints it in English at the
// cene in the valuta (ITERATION_PAGEHDRIZV and ITERATION_KURS).
type KomOtpStampa struct {
	Firma          RobnoStampaFakturaFirmaDto
	Kupac          KomPonudaStampaDobavljac
	TekuciRacun    string
	Storno         bool
	Ino            bool
	OznVal         string
	SifVal         int
	Kurs           float64
	VaziZa         int
	Brotp          int
	Datotp         sql.NullTime
	NalogDokum     int
	Pla            string
	Rok            int
	Komercijalista string
	Vozac          string
	Brvozila       string
	Magacin        string
	OJ             string
	MestoIzdavanja string
	Bttowght       float64
	Netwght        float64
	Totboxs        int
	Brpaleta       int
	Napomena       string
	Pornapomena    string
	Nepotvrdjen    bool
	Stavke         []KomOtpStampaStavka
	Svega          float64
	UkupnoRSD      float64
}

// KomOtpStampaStavka is a printed stavka of an otpremnica.
type KomOtpStampaStavka struct {
	Rbr      int
	Sifra    int
	Naziv    string
	JM       string
	Otk      string
	Serija   string
	Rok      string
	Kolicina float64
	Decimale int
	Cena     float64
	ProcRab  float64
	Rabat    float64
	Iznos    float64
}

// KomOtpUI are the urls of the entry dialog of an otpremnica the shared RobnoFaktureUI does not have (JSON): the nalog za
// isporuku of a new one and the valuta of a šifra with its kurs.
type KomOtpUI struct {
	NalogURL  string
	ValutaURL string
	// Eotp are the urls and the ids of the controls of the e-otpremnica (the "eOtpremnice" screen).
	Eotp KomEotpUI
}

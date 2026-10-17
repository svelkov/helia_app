package domain

import (
	"database/sql"
	"time"
)

// The "Profakture" screen of the komercijalno (the legacy pr_Profakture with its stavke pr_Profakture_DETALJI): a
// document of edok (not of a nalog) of the vrste of the modul PRO and the group PRF (the predračun) for one kupac (fkpl
// konto and šifra) in a magacin of the tipovi G and V, with its stavke in epro: the artikal with its lot (OTK, serija,
// rok trajanja), the količina (trazkolic), the prodajna cena in RSD (fcena) and in the valuta of the document (cenaval),
// the planska cena (cena) and the rabat. Its statuses: 10 new, 20 confirmed ("Potvrda dok."), 40 the nalog (a
// porudžbenica kupca, the group NRK) formed from it ("Form. nalog"); a profaktura with the status 40 cannot change.

// The statuses of a profaktura.
const (
	KomProfakturaNova       = "10"
	KomProfakturaPotvrdjena = "20"
	KomProfakturaNalog      = "40"
)

// KomProfakturaRow is a profaktura as edok keeps it, with the texts of its šifre, its magacin (tip, zalihe, cene, mesto)
// and the broj of the nalog formed from it (KomercijalnoRepository).
type KomProfakturaRow struct {
	RdokID          int64
	Vrd             int
	Tipdok          string
	Dokum           int
	Dadok           sql.NullTime
	Rok             int
	MagaciniID      int
	Mag             int
	MagacinText     string
	MagacinAdresa   string
	MagacinMesto    string
	Tipmag          string
	Tipzal          int
	Tipcene         int
	Fkto            string
	Fana            string
	KomID           sql.NullInt64
	KomSifra        int
	KomIme          string
	Orgjed          sql.NullInt64
	OrgjedText      string
	MestoTroska     sql.NullInt64
	MestoTroskaText string
	Foot            string
	Doksts          string
	Knjige1         string
	Knjige4         string
	SifVal          int
	OznVal          string
	Kurs            float64
	VaziZa          int
	Otp             string
	Pla             string
	Posuslnab       string
	Mi              int
	Ugrabat         float64
	Vrugrabat       float64
	Pkase           float64
	Vkase           float64
	Iznos           float64
	Netwght         float64
	CenovnikID      int64
	Paritet         string
	Izjizv          int
	Keykat          string
	Category        string
	Pornapomena     string
	Kamt            string
	Brugov          string
	Vozac           string
	Brvozila        string
	NalogDokum      int
}

// KomProfakturaUpis are the columns of edok the save of the header writes (the legacy SnimiEDOK, ScreenToFile).
type KomProfakturaUpis struct {
	RdokID      int64
	Vrd         int
	Tipdok      string
	Dokum       int
	Dadok       time.Time
	Rok         int
	MagaciniID  int
	Mag         int
	Fkto        string
	Fana        string
	KomID       sql.NullInt64
	Orgjed      sql.NullInt64
	MestoTroska sql.NullInt64
	Foot        string
	SifVal      int
	Kurs        float64
	VaziZa      int
	Otp         string
	Pla         string
	Posuslnab   string
	Mi          int
	Ugrabat     float64
	Pkase       float64
	CenovnikID  int64
	Paritet     string
	Izjizv      int
	Keykat      string
	Category    string
	Pornapomena string
	Kamt        string
	Brugov      string
	// Opis is the opis of a new document ("Profaktura br: <broj>").
	Opis string
}

// KomProKupac is the kupac of a konto and šifra (fkpl with its partner) as the header shows it, with its ugovoreni
// rabat and rok plaćanja (partneri.ugrabat, partneri.proval; the legacy exit of EDT_FANA).
type KomProKupac struct {
	KomDobavljacPodaci
	Ugrabat float64 `json:"ugrabat"`
	Rok     int     `json:"rok"`
	// Ter is the teritorija of the partner (3: a foreign one, the legacy PARTNERI.TER).
	Ter int `json:"ter"`
}

// KomProMestoIsporuke is a mesto isporuke of a kupac (fisp; the legacy BTN_Sifra3).
type KomProMestoIsporuke struct {
	Naziv  string `json:"naziv"`
	Adresa string `json:"adresa"`
	Pobro  string `json:"pobro"`
	Mesto  string `json:"mesto"`
	Pib    string `json:"pib"`
	Jib    string `json:"jib"`
	Gln    string `json:"gln"`
}

// KomProStavkaRow is a stavka of a profaktura (epro with the naziv, the JM, the komercijalni opis, the proizvođač, the
// težina of its artikal and the decimals of its JM).
type KomProStavkaRow struct {
	RproID     int64
	Rbr        int
	Konto      string
	Sifra      int
	RsifID     int64
	Naziv      string
	JM         string
	Komercopis string
	Pro        string
	Proizsifra string
	Otk        string
	Serija     string
	Roktr      int64
	Kolicina   float64
	Cena       float64
	Fcena      float64
	CenaVal    float64
	Rabat      float64
	Po         int
	Tezjm      float64
	Decimale   int
	// Isporuceno is the isporučena količina (kolic; Kolicina is the tražena one) and Kubikaza the zapremina of one unit
	// of the artikal (rsif.kubikaza).
	Isporuceno float64
	Kubikaza   float64
}

// KomProStavkaUpis are the columns of epro the save of a new stavka writes (the legacy AddEPRO).
type KomProStavkaUpis struct {
	RdokID     int64
	Rbr        int
	Sifra      int
	RsifID     int64
	Po         int
	Konto      string
	Mag        int
	MagaciniID int
	Kolicina   float64
	Cena       float64
	Fcena      float64
	CenaVal    float64
	Rabat      float64
	Otk        string
	Serija     string
	Roktr      int64
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
	// Isporuceno is the isporučena količina (kolic): the količina for a profaktura, 0 for a nalog za isporuku (its
	// otpremnica writes it).
	Isporuceno float64
}

// KomProStavkaIzmena are the columns of epro the change of a stavka writes (the legacy exit of a changed row of
// TABLE_PR_QRY_EPRO): the količina, the prodajna cena in RSD and in the valuta and the rabat.
type KomProStavkaIzmena struct {
	RdokID   int64
	RproID   int64
	Kolicina float64
	Fcena    float64
	CenaVal  float64
	Rabat    float64
}

// KomProLot is a lot (drsta) of an artikal in a magacin with its stanje (ulaz - izlaz - rezervisano).
type KomProLot struct {
	RstaID int64   `json:"rstaid"`
	Konto  string  `json:"konto"`
	Otk    string  `json:"otk"`
	Serija string  `json:"serija"`
	Roktr  int64   `json:"roktr"`
	Rok    string  `json:"rok"`
	Stanje float64 `json:"stanje"`
}

// KomProArtikal is the artikal of a stavka (the legacy BTN_Sifra and the exit of EDT_SIFRA): its lots in the magacin
// and its prodajna cena with the rabat by the tip cene of the magacin (rcene.cena, or the cenovnik of the profaktura).
type KomProArtikal struct {
	RsifID int64       `json:"-"`
	Sifra  int         `json:"sifra"`
	Naziv  string      `json:"naziv"`
	JM     string      `json:"jm"`
	Po     int         `json:"-"`
	Cena   float64     `json:"cena"`
	Rabat  float64     `json:"rabat"`
	Lotovi []KomProLot `json:"lotovi"`
}

// KomProPrikazRow is a row of the list of the main screen (the legacy PR_QRY_EDOK of the tab "Pregled profaktura").
type KomProPrikazRow struct {
	RdokID     int64
	Magacin    string
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
	Fkto       string
	Fana       string
	Kupac      string
	Opis       string
	Rok        int
	Otp        string
	Doksts     string
	NalogDokum int
}

// KomProPrikazParams are the filters of the list of the main screen: the magacin (or all the magacini of the user) and
// the status (the nepotvrđene, 10, else the potvrđene: 20 and 40), with the text of the search.
type KomProPrikazParams struct {
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	Tekst       string
}

// KomProHeaderView is the header of the entry dialog: the saved profaktura or a new one (RdokID 0) with its defaults.
type KomProHeaderView struct {
	KomProfakturaRow
	Datum         time.Time
	Kupac         KomDobavljacPodaci
	MestoIsporuke KomProMestoIsporuke
	CenovnikText  string
	CenovnikOpis  string
	Ino           bool
	ImaStavke     bool
	PdvObveznik   bool
	Snimljen      bool
}

// KomProDokument is the content of the entry dialog: the header with the lists of its combos, the konta of its magacin,
// the next redni broj of a stavka and a message.
type KomProDokument struct {
	Header     KomProHeaderView
	Vrste      []ComboItem
	Tipdok     []ComboItem
	Magacini   []ComboItem
	Konta      []ComboItem
	ZiroRacuni []ComboItem
	// UsloviPlacanja and NaciniIsporuke are the lists of the combos of the uslovi plaćanja and the načini isporuke.
	UsloviPlacanja []ComboItem
	NaciniIsporuke []ComboItem
	NoviRbr        int
	Poruka         string
	SifValFirme    int
	// BrojLabel and DatumLabel are the labels of the broj and the datum of the document ("Broj profakture").
	BrojLabel  string
	DatumLabel string
}

// KomProParams are the fields of the header the screen sends ("Sačuvaj", the legacy BTN_Snimi1).
type KomProParams struct {
	RdokID      int64  `form:"rdokid"`
	Vrd         string `form:"vrd"`
	Tipdok      string `form:"tipdok"`
	MagaciniID  string `form:"magaciniid"`
	Datum       string `form:"datum"`
	Rok         string `form:"rok"`
	Fkto        string `form:"fkto"`
	Fana        string `form:"fana"`
	KomSifra    string `form:"komsifra"`
	Orgjed      string `form:"idorgjed"`
	MestoTroska string `form:"mestotrid"`
	Foot        string `form:"foot"`
	SifVal      string `form:"sifval"`
	Kurs        string `form:"kurs"`
	VaziZa      string `form:"vaziza"`
	Otp         string `form:"otp"`
	Pla         string `form:"pla"`
	Posuslnab   string `form:"posuslnab"`
	Mi          string `form:"mi"`
	Ugrabat     string `form:"ugrabat"`
	Pkase       string `form:"pkase"`
	CenovnikID  string `form:"cenovnikid"`
	Paritet     string `form:"paritet"`
	Izjizv      string `form:"izjizv"`
	Keykat      string `form:"keykat"`
	Category    string `form:"oznakakat"`
	Pornapomena string `form:"napomporeskoosl"`
	Kamt        string `form:"kamt"`
	Brugov      string `form:"brugov"`
}

// KomProStavkaParams are the fields of a stavka ("Dodaj" of the stavka, the legacy BTN_DODAJDET; a rproid of a saved
// stavka changes its količina, cena and rabat).
type KomProStavkaParams struct {
	RdokID   int64  `form:"rdokid"`
	RproID   int64  `form:"rproid"`
	Konto    string `form:"konto"`
	Sifra    string `form:"sifraartikla"`
	Otk      string `form:"otk"`
	Serija   string `form:"serija"`
	Rok      string `form:"roktr"`
	Kolicina string `form:"kolicina"`
	Cena     string `form:"cena"`
	Rabat    string `form:"rabat"`
}

// KomProStavkaForma is a saved stavka for the form of the stavke ("Izmeni" of its row): the cena in the valuta of the
// profaktura.
type KomProStavkaForma struct {
	RproID   int64   `json:"rproid"`
	Rbr      int     `json:"rbr"`
	Konto    string  `json:"konto"`
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	JM       string  `json:"jm"`
	Otk      string  `json:"otk"`
	Serija   string  `json:"serija"`
	Rok      string  `json:"rok"`
	Kolicina float64 `json:"kolicina"`
	Cena     float64 `json:"cena"`
	Rabat    float64 `json:"rabat"`
}

// KomProStampa is what the print of a profaktura shows (the legacy PR_RPT_PROFAKTURA, PR_RPT_PROFAKTURAIZV for a
// foreign valuta): the firma, the kupac with the mesto isporuke, the header, the stavke and the totals.
type KomProStampa struct {
	Firma          RobnoStampaFakturaFirmaDto
	ZiroRacuni     []string
	Kupac          KomPonudaStampaDobavljac
	MestoIsporuke  string
	MestoIsporuke2 string
	Ino            bool
	OznVal         string
	Dokum          int
	Datum          sql.NullTime
	Valuta         sql.NullTime
	MestoIzdavanja string
	Pla            string
	Otp            string
	Rok            int
	Komercijalista string
	Vozac          string
	Brvozila       string
	Paritet        string
	Izjizv         bool
	Napomena       string
	Pornapomena    string
	NetoTezina     float64
	PdvObveznik    bool
	Stavke         []KomProStampaStavka
	Svega          float64
	Rabat          float64
	Neto           float64
	NetoPrikaz     bool
	UgrabatPct     float64
	Ugrabat        float64
	KasaPct        float64
	Kasa           float64
	PdvPoStopama   []KomProStampaPdv
	ZaNaplatu      float64
	Slovima        string
}

// KomProStampaStavka is a printed stavka: the cena is the prodajna cena (in the valuta for a foreign one).
type KomProStampaStavka struct {
	Rbr      int
	Sifra    int
	Naziv    string
	JM       string
	Kolicina float64
	Decimale int
	Cena     float64
	ProcRab  float64
	Rabat    float64
	Stopa    float64
	PdvJM    float64
	Iznos    float64
}

// KomProStampaPdv is the poreska osnovica and the PDV of a stopa of the print.
type KomProStampaPdv struct {
	Stopa    float64
	Osnovica float64
	Pdv      float64
}

// KomProButtons are the buttons of the entry dialog the shared RobnoFaktureButtons do not have: the potvrda, "Form.
// nalog" and the print.
type KomProButtons struct {
	Potvrda Button
	Nalog   Button
	Stampa  Button
}

// KomProUI are the urls of the entry dialog the shared RobnoFaktureUI does not have.
type KomProUI struct {
	// KupacURL returns the kupac of a konto and šifra, MestoURL a mesto isporuke, ValutaURL the valuta of a šifra with
	// its kurs, CenovnikURL the text of a cenovnik (JSON).
	KupacURL    string
	MestoURL    string
	ValutaURL   string
	CenovnikURL string
}

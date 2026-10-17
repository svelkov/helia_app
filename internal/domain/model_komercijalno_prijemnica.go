package domain

import (
	"database/sql"
	"time"
)

// The "Prijemnica materijala po porudžbenici" screen of the komercijalno (the legacy pr_PrijemnicaSIROVINE): a document
// of edok (not of a nalog) of the vrste of the group ULS (the magacini of the tipovi P, S, L and D) or PLT (the magacini
// of the tip V) for a potvrđena porudžbenica dobavljaču (vrdokid), whose dobavljač, org. jedinica and mesto troška it
// takes, with its stavke in epro: taken from the stavke of the porudžbenica or entered by hand, each one with the
// konto of the magacin, the količina na dokumentu (trazkolic), the primljena količina (kolic), the serija and the OTK.
// Its broj (dokum) is the broj prijemnice (brotp), its datum (dadok) the datum prijemnice (datotp). The potvrda books
// the stavke into the stanje (the legacy AzurRSTA_ZN and AzurDRSTA_ZN). The pakovanja of the stavke are in kolpak.

// KomPrijemnicaAnalizirana is the status of a prijemnica confirmed in a magacin with the analiza (magacini.analiza 1:
// the stavke in the neraspoložive zalihe until the analiza); in a magacin without it the status is KomZahtevPotvrdjen.
const KomPrijemnicaAnalizirana = "30"

// KomPrijemnicaRow is a prijemnica as edok keeps it, with the texts of its magacin, org. jedinica and mesto troška, its
// vrsta (dokvrstaid, predznak) and the broj and datum of its porudžbenica (KomercijalnoRepository).
type KomPrijemnicaRow struct {
	RdokID          int64
	Vrd             int
	DokvrstaID      int64
	Predznak        string
	Dokum           int
	Dadok           sql.NullTime
	Datotp          sql.NullTime
	Tipdok          string
	MagaciniID      int
	Mag             int
	MagacinText     string
	MagacinAdresa   string
	Analiza         int
	Fkto            string
	Fana            string
	Orgjed          sql.NullInt64
	OrgjedText      string
	MestoTroska     sql.NullInt64
	MestoTroskaText string
	Doksts          string
	Knjige1         string
	VrdokID         int64
	VezaDokum       int
	VezaDatum       sql.NullTime
}

// KomPrijemnicaUpis are the columns of edok the save of the header writes (the legacy SnimiEDOK; a new prijemnica takes
// the rest from its porudžbenica).
type KomPrijemnicaUpis struct {
	RdokID     int64
	Vrd        int
	Dokum      int
	Datum      time.Time
	MagaciniID int
	Mag        int
	VrdokID    int64
}

// KomPrijVeza is the potvrđena porudžbenica dobavljaču a prijemnica is made for (the legacy BTN_TraziDOK and PodaciHDR):
// what the header shows of it.
type KomPrijVeza struct {
	RdokID      int64
	Dokum       int
	Dadok       sql.NullTime
	MagaciniID  int
	Orgjed      sql.NullInt64
	OrgjedText  string
	MestoTroska sql.NullInt64
	MtText      string
	Fkto        string
	Fana        string
}

// KomPrijStavkaRow is a stavka of a prijemnica (epro with the naziv and the JM of its artikal and the decimals of the
// JM).
type KomPrijStavkaRow struct {
	RproID    int64
	Rbr       int
	Konto     string
	Sifra     int
	Naziv     string
	JM        string
	TrazKolic float64
	Kolic     float64
	Serija    string
	Otk       string
	Roktr     int64
	Cena      float64
	Decimale  int
}

// KomPrijStavkaUpis are the columns of epro the save of a new stavka writes (the legacy AddEPRO for a stavka entered by
// hand, AddEPRONAR for a stavka of the porudžbenica with its cena and rabat).
type KomPrijStavkaUpis struct {
	RdokID     int64
	Rbr        int
	Sifra      int
	RsifID     int64
	Po         int
	Konto      string
	Mag        int
	MagaciniID int
	TrazKolic  float64
	Kolic      float64
	Otk        string
	Serija     string
	Fcena      float64
	Cena       float64
	Iznos      float64
	Rabat      float64
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
}

// KomPrijStavkaIzmena are the columns of epro the change of a stavka writes (the legacy exit of a changed row of
// TABLE_PR_QRY_EPRO1): the iznos is the cena of the stavka by the primljena količina.
type KomPrijStavkaIzmena struct {
	RdokID    int64
	RproID    int64
	Konto     string
	TrazKolic float64
	Kolic     float64
	Serija    string
	Otk       string
}

// KomPrijIzvor is a stavka of the porudžbenica a prijemnica takes (the legacy AddEPRONAR): its artikal, its tražena
// količina, its cena and rabat.
type KomPrijIzvor struct {
	RproID    int64
	Sifra     int
	Po        int
	RsifID    int64
	TrazKolic float64
	Cena      float64
	Rabat     float64
}

// KomPrijArtikal is the artikal of a stavka entered by hand (the legacy exit of EDT_SIFRA): with the OTK of its next
// prijemnica (the legacy VratiOTK).
type KomPrijArtikal struct {
	RsifID int64  `json:"-"`
	Sifra  int    `json:"sifra"`
	Naziv  string `json:"naziv"`
	JM     string `json:"jm"`
	Po     int    `json:"-"`
	Otk    string `json:"otk"`
}

// KomPrijPakovanjeRow is a stavka of a prijemnica with its pakovanje (kolpak; the legacy pr_Pakovanja "Količine po
// pakovanju"): the količina po pakovanju, the broj pakovanja and their product, the uskladištena količina.
type KomPrijPakovanjeRow struct {
	RproID       int64
	Rbr          int
	Sifra        int
	Naziv        string
	JM           string
	KolPoPak     float64
	BrPak        float64
	Uskladisteno float64
}

// KomPrijPrikazRow is a row of the list of the main screen (the legacy PR_QRY_EDOK of the tab "Pregled prijemnica").
type KomPrijPrikazRow struct {
	RdokID    int64
	Magacin   string
	Vrd       int
	Dokum     int
	Dadok     sql.NullTime
	Fkto      string
	Fana      string
	Dobavljac string
	Opis      string
	Rok       int
	Datiz     sql.NullTime
	Otp       string
	Kamt      string
}

// KomPrijPrikazParams are the filters of the list of the main screen: the magacin (or all the magacini of the user) and
// the status (the nepotvrđene, else the potvrđene: the statuses 30 and 40), with the text of the search.
type KomPrijPrikazParams struct {
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	Tekst       string
}

// KomPrijHeaderView is the header of the entry dialog: the saved prijemnica or a new one (RdokID 0) with its defaults.
type KomPrijHeaderView struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Datum           time.Time
	MagaciniID      int
	Fkto            string
	Fana            string
	Dobavljac       KomDobavljacPodaci
	OrgjedText      string
	MestoTroskaText string
	VrdokID         int64
	VezaDokum       int
	VezaDatum       sql.NullTime
	// Storno: the vrsta has the predznak "-" (the primljene količine are negative).
	Storno bool
	// Snimljen: the header is saved and locked, the stavke are open (after "Sačuvaj").
	Snimljen bool
}

// KomPrijDokument is the content of the entry dialog: the header with the lists of its combos, the stavke of its
// porudžbenica, the konta of its magacin, the next redni broj of a stavka, whether it has stavke and a message.
type KomPrijDokument struct {
	Header    KomPrijHeaderView
	Vrste     []ComboItem
	Magacini  []ComboItem
	Konta     []ComboItem
	Izvori    []KomPorStavkaRow
	NoviRbr   int
	ImaStavke bool
	Poruka    string
}

// KomPrijParams are the fields of the header the screen sends ("Sačuvaj", the legacy BTN_Snimi1).
type KomPrijParams struct {
	RdokID     int64  `form:"rdokid"`
	Vrd        string `form:"vrd"`
	MagaciniID string `form:"magaciniid"`
	Datum      string `form:"datum"`
	VrdokID    string `form:"vrdokid"`
}

// KomPrijStavkaParams are the fields of a stavka ("Sačuvaj" of the stavka, the legacy BTN_DODAJDET; a rproid of a saved
// stavka changes it, the legacy change of a row of the grid).
type KomPrijStavkaParams struct {
	RdokID    int64  `form:"rdokid"`
	RproID    int64  `form:"rproid"`
	Konto     string `form:"konto"`
	Sifra     string `form:"sifraartikla"`
	TrazKolic string `form:"trazkolic"`
	Kolic     string `form:"kolic"`
	Serija    string `form:"serija"`
	Otk       string `form:"otk"`
}

// KomPrijStavkaForma is a saved stavka for the form of the stavke ("Izmeni" of its row): the količine without the sign
// of a storno.
type KomPrijStavkaForma struct {
	RproID    int64   `json:"rproid"`
	Rbr       int     `json:"rbr"`
	Konto     string  `json:"konto"`
	Sifra     int     `json:"sifra"`
	Naziv     string  `json:"naziv"`
	JM        string  `json:"jm"`
	TrazKolic float64 `json:"trazkolic"`
	Kolic     float64 `json:"kolic"`
	Serija    string  `json:"serija"`
	Otk       string  `json:"otk"`
}

// KomPrijPakovanjeParams are the fields of the pakovanje of a stavka ("Sačuvaj" of the dialog "Pakovanje").
type KomPrijPakovanjeParams struct {
	RdokID   int64  `form:"rdokid"`
	RproID   int64  `form:"pakrproid"`
	KolPoPak string `form:"kolpopak"`
	BrPak    string `form:"brpak"`
}

// KomPrijStampa is what the print of a prijemnica shows (the legacy PR_RPT_PRIJEMNICA): the firma with its računi, the
// dobavljač, the header with the porudžbenica and the stavke; Nepotvrdjen marks a nepotvrđena one.
type KomPrijStampa struct {
	Firma         RobnoStampaFakturaFirmaDto
	ZiroRacuni    []string
	Dobavljac     KomPonudaStampaDobavljac
	Dokum         int
	Datum         sql.NullTime
	VezaDokum     int
	VezaDatum     sql.NullTime
	Magacin       string
	MagacinAdresa string
	OJ            string
	Nepotvrdjen   bool
	Stavke        []KomPrijStavkaRow
}

// KomPrijKarton is the karton identifikacije of a stavka of a prijemnica (the legacy RPT_KI): the dobavljač, the broj
// and the datum of the prijemnica (the datum prispeća), the artikal (materijal) with its serija, OTK and primljena
// količina.
type KomPrijKarton struct {
	Dobavljac string
	Dokum     int
	Datum     sql.NullTime
	Sifra     int
	Naziv     string
	JM        string
	Serija    string
	Otk       string
	Kolicina  float64
	Decimale  int
}

// KomPrijButtons are the buttons of the entry dialog the shared RobnoFaktureButtons do not have: the potvrda ("Potvrda
// dok."), the print ("Štampa"), the buttons of the stavke of the porudžbenica ("Dodaj", "Dodaj sve", "Vrati poslednju",
// "Vrati sve"), "Pakovanje" and "Karton identifikacije".
type KomPrijButtons struct {
	Potvrda        Button
	Stampa         Button
	Dodaj          Button
	DodajSve       Button
	VratiPoslednju Button
	VratiSve       Button
	Pakovanje      Button
	Karton         Button
}

// KomPrijUI are the ids and the urls of the entry dialog the shared RobnoFaktureUI does not have.
type KomPrijUI struct {
	// IzvorID is the hidden field of the selected stavka of the porudžbenica ("Dodaj"), KartonID the one of the selected
	// stavka of the prijemnica ("Karton identifikacije").
	IzvorID  string
	KartonID string
}

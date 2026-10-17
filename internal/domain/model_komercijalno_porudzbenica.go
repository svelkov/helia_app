package domain

import (
	"database/sql"
	"time"
)

// The "Porudžbenice dobavljačima" screen of the komercijalno (the legacy pr_PorudzbenicaDOBAVLJ with its stavke
// pr_PorudzbenicaDOBAVLJ_DETALJI): a document of edok (not of a nalog) of the vrste of the modul PRO and the group NRD
// for one dobavljač (fkpl konto and šifra), with its stavke in epro. It may be linked (vrdokid) to a potvrđen zahtev za
// ponudu, whose stavke it takes like the zahtev za ponudu takes the stavke of the zahtevi za nabavku: a taken stavka
// keeps the rproid of its source in rproid1 and the source is rezervisana (ststatus R) until the stavka is deleted.
// The cena of a stavka is entered in the valuta of the document (cenaval) and kept in RSD (cena, ncena) by the kurs.
// The statuses are the ones of the zahtev za nabavku (KomZahtevNepotvrdjen, KomZahtevPotvrdjen).

// KomPorudzbenicaRow is a porudžbenica as edok keeps it, with the texts of its šifre and the broj and datum of its
// zahtev za ponudu (KomercijalnoRepository).
type KomPorudzbenicaRow struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Dadok           sql.NullTime
	Rok             int
	MagaciniID      int
	Mag             int
	MagacinText     string
	MagacinAdresa   string
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
	VrdokID         int64
	VezaDokum       int
	VezaDatum       sql.NullTime
	Otp             string
	Pla             string
	Brugov          string
	Nacpak          string
	Posuslnab       string
	Pratdok         string
}

// KomPorudzbenicaUpis are the columns of edok the save of the header writes (the legacy SnimiEDOK).
type KomPorudzbenicaUpis struct {
	RdokID      int64
	Vrd         int
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
	VrdokID     int64
	Otp         string
	Pla         string
	Brugov      string
	Nacpak      string
	Posuslnab   string
	Pratdok     string
}

// KomPorVeza is the potvrđen zahtev za ponudu a new porudžbenica is linked to (the legacy BTN_TraziDOK): what the
// porudžbenica takes of it.
type KomPorVeza struct {
	RdokID      int64
	Dokum       int
	Dadok       sql.NullTime
	Rok         int
	Otp         string
	MagaciniID  int
	Orgjed      sql.NullInt64
	OrgjedText  string
	MestoTroska sql.NullInt64
	MtText      string
	Fkto        string
	Fana        string
}

// KomDobavljacPodaci is the dobavljač of a konto and šifra (fkpl with its partner) as the header shows it.
type KomDobavljacPodaci struct {
	Naziv     string `json:"naziv"`
	Adresa    string `json:"adresa"`
	Pobro     string `json:"pobro"`
	Mesto     string `json:"mesto"`
	Pib       string `json:"pib"`
	TipPdv    int    `json:"-"`
	SistemPdv string `json:"sistempdv"`
}

// KomPorStavkaRow is a stavka of a porudžbenica (epro with the naziv, the JM and the komercijalni opis of its
// artikal and the decimals of its JM).
type KomPorStavkaRow struct {
	RproID     int64
	Rbr        int
	Sifra      int
	Naziv      string
	JM         string
	Komercopis string
	Kolicina   float64
	Cena       float64
	CenaVal    float64
	Rabat      float64
	RproID1    int64
	Decimale   int
}

// KomPorStavkaUpis are the columns of epro the save of a stavka writes (the legacy AddEPRO and AddEPRONAR): RproID1 is
// the stavka of the zahtev za ponudu it was taken from (0 for a stavka entered by hand).
type KomPorStavkaUpis struct {
	RdokID     int64
	Rbr        int
	Sifra      int
	RsifID     int64
	Po         int
	Kolicina   float64
	Cena       float64
	CenaVal    float64
	Rabat      float64
	Mag        int
	MagaciniID int
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
	RproID1    int64
}

// KomPorIzvorRow is a slobodna stavka of the zahtev za ponudu of a porudžbenica (the legacy "Količine po ponudi").
type KomPorIzvorRow struct {
	RproID   int64
	Rbr      int
	Sifra    int
	Naziv    string
	JM       string
	Kolicina float64
}

// KomPorIzvor is a stavka of a zahtev za ponudu a porudžbenica takes (the legacy AddEPRONAR): its artikal, its tražena
// količina and its cena in the valuta.
type KomPorIzvor struct {
	RproID   int64
	Sifra    int
	Po       int
	RsifID   int64
	Kolicina float64
	CenaVal  float64
}

// KomPorPrikazRow is a row of the grid of the tab "Pregled porudžbenice" (the legacy PR_QRY_EDOK).
type KomPorPrikazRow struct {
	RdokID    int64
	Magacin   string
	Vrd       int
	Dokum     int
	Dadok     sql.NullTime
	Fkto      string
	Fana      string
	Dobavljac string
	Rok       int
	Otp       string
	Referent  string
}

// KomPorStampa is what the print of a porudžbenica shows (the legacy PR_RPT_PORUDZB_DOB, PR_RPT_PORUDZB_DOBINO for a
// foreign valuta): the firma with its računi, the dobavljač, the header, the uslovi nabavke and the stavke.
type KomPorStampa struct {
	Firma         RobnoStampaFakturaFirmaDto
	ZiroRacuni    []string
	Dobavljac     KomPonudaStampaDobavljac
	Ino           bool
	OznVal        string
	Dokum         int
	Datum         sql.NullTime
	Rok           int
	Napomena      string
	Otp           string
	Pla           string
	Brugov        string
	Nacpak        string
	Posuslnab     string
	Magacin       string
	MagacinAdresa string
	OJ            string
	Referent      string
	Stavke        []KomPorStampaStavka
	Ukupno        float64
}

// KomPorStampaStavka is a printed stavka: the opis is the naziv and/or the komercijalni opis of the artikal (the
// legacy CBOX_NAZIV and CBOX_KOMERCOPIS), the cena is the one of the valuta of the document.
type KomPorStampaStavka struct {
	Rbr      int
	Sifra    int
	Opis     string
	JM       string
	Kolicina float64
	Decimale int
	Cena     float64
	Rabat    float64
	Vrednost float64
}

// KomPorHeaderView is the header of the screen: the saved porudžbenica or a new one (RdokID 0) with its defaults.
type KomPorHeaderView struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Datum           time.Time
	Rok             int
	MagaciniID      int
	Fkto            string
	Fana            string
	Dobavljac       KomDobavljacPodaci
	KomSifra        int
	KomIme          string
	Orgjed          sql.NullInt64
	OrgjedText      string
	MestoTroska     sql.NullInt64
	MestoTroskaText string
	Tekst           string
	SifVal          int
	OznVal          string
	Kurs            float64
	VaziZa          int
	VrdokID         int64
	VezaDokum       int
	VezaDatum       sql.NullTime
	Otp             string
	Pla             string
	Brugov          string
	Nacpak          string
	Posuslnab       string
	Pratdok         string
	// Ino: the valuta is not the one of the firma (the legacy STC_IZVOZ "UVOZNI DOKUMENT - INO PARTNER").
	Ino bool
	// Snimljen: the header is saved and locked, the stavke are open (after "Sačuvaj").
	Snimljen bool
}

// KomPorDokument is the content of the tab "Porudžbenice dobavljačima": the header with the lists of its combos, the
// slobodne stavke of its zahtev za ponudu, the next redni broj of a stavka, whether it has stavke and a message.
type KomPorDokument struct {
	Header    KomPorHeaderView
	Vrste     []ComboItem
	Magacini  []ComboItem
	Izvori    []KomPorIzvorRow
	NoviRbr   int
	ImaStavke bool
	Poruka    string
	// SifValFirme is the valuta of the firma: another one is a foreign valuta (the INO document).
	SifValFirme int
	// Ukupno is the iznos of the porudžbenica: the total of the vrednost of its stavke (the podaci o dokumentu above
	// the stavke).
	Ukupno float64
}

// KomPorParams are the fields of the header the screen sends ("Sačuvaj", the legacy BTN_Snimi1).
type KomPorParams struct {
	RdokID      int64  `form:"rdokid"`
	Vrd         string `form:"vrd"`
	MagaciniID  string `form:"magaciniid"`
	Datum       string `form:"datum"`
	Rok         string `form:"rok"`
	VrdokID     string `form:"vrdokid"`
	Fkto        string `form:"fkto"`
	Fana        string `form:"fana"`
	KomSifra    string `form:"komsifra"`
	Orgjed      string `form:"idorgjed"`
	MestoTroska string `form:"mestotrid"`
	Tekst       string `form:"foot"`
	SifVal      string `form:"sifval"`
	Kurs        string `form:"kurs"`
	VaziZa      string `form:"vaziza"`
	Otp         string `form:"otp"`
	Pla         string `form:"pla"`
	Brugov      string `form:"brugov"`
	Nacpak      string `form:"nacpak"`
	Posuslnab   string `form:"posuslnab"`
	Pratdok     string `form:"pratdok"`
}

// KomPorStavkaParams are the fields of a stavka ("Sačuvaj" of the stavka, the legacy BTN_DODAJDET; a rproid of a saved
// stavka changes its količina, cena and rabat, the legacy change of a row of the grid).
type KomPorStavkaParams struct {
	RdokID   int64  `form:"rdokid"`
	RproID   int64  `form:"rproid"`
	Sifra    string `form:"sifraartikla"`
	Kolicina string `form:"kolicina"`
	Cena     string `form:"cena"`
	Rabat    string `form:"rabat"`
}

// KomPorStavkaForma is a saved stavka for the form of the stavke ("Izmeni" of its row): the cena in the valuta of the
// document.
type KomPorStavkaForma struct {
	RproID   int64   `json:"rproid"`
	Rbr      int     `json:"rbr"`
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	JM       string  `json:"jm"`
	Kolicina float64 `json:"kolicina"`
	Cena     float64 `json:"cena"`
	Rabat    float64 `json:"rabat"`
}

// KomPorButtons are the buttons of the "Porudžbenice dobavljačima" screen the shared RobnoFaktureButtons do not have:
// the potvrda ("Potvrda dok."), the print of the porudžbenica ("Štampa") and the buttons of the stavke of the zahtev
// za ponudu ("Dodaj", "Dodaj sve", "Vrati poslednju", "Vrati sve").
type KomPorButtons struct {
	Potvrda        Button
	Stampa         Button
	Dodaj          Button
	DodajSve       Button
	VratiPoslednju Button
	VratiSve       Button
}

// KomPorUI are the ids and the urls of the "Porudžbenice dobavljačima" screen the shared RobnoFaktureUI does not have.
type KomPorUI struct {
	// DobavljacURL returns the dobavljač of a konto and šifra, ValutaURL the valuta of a šifra with its kurs (JSON).
	DobavljacURL string
	ValutaURL    string
	// The checkboxes of the print: the naziv and the komercijalni opis of the artikal in the opis of a stavka.
	StampaNazivID      string
	StampaKomercopisID string
}

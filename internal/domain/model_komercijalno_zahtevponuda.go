package domain

import (
	"database/sql"
	"time"
)

// The "Zahtev za ponudu" screen of the komercijalno (the legacy pr_ZahtevPONUDA): a document of edok (not of a nalog)
// with its stavke in epro, the vrste dokumenata of the modul PRO and the group ZPN. Its stavke are taken mostly from the
// slobodne stavke (ststatus S) of the potvrđeni zahtevi za nabavku (the group INR): a taken stavka keeps the rproid of
// its source in rproid1 and the source becomes rezervisana (ststatus R) until the stavka is deleted. The dobavljači the
// zahtev is sent to are fkpl analitike, kept as "idfkpl;idfkpl;" in edok.vezniid. The statuses are the ones of the
// zahtev za nabavku (KomZahtevNepotvrdjen, KomZahtevPotvrdjen).

// KomPonudaVezniIDMax is the length of edok.vezniid (varchar(80)): the most "idfkpl;" of the dobavljači a zahtev za
// ponudu can keep.
const KomPonudaVezniIDMax = 80

// KomPonudaRow is a zahtev za ponudu as edok keeps it, with the texts of its šifre (KomercijalnoRepository).
type KomPonudaRow struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Dadok           sql.NullTime
	Rok             int
	MagaciniID      int
	Mag             int
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
	SifVal          int
	OznVal          string
	Kurs            float64
	VaziZa          int
	VezniID         string
}

// KomPonudaUpis are the columns of edok the save of the header writes (the legacy SnimiEDOK).
type KomPonudaUpis struct {
	RdokID      int64
	Vrd         int
	Dokum       int
	Dadok       time.Time
	Rok         int
	MagaciniID  int
	Mag         int
	KomID       sql.NullInt64
	Orgjed      sql.NullInt64
	MestoTroska sql.NullInt64
	Foot        string
	SifVal      int
	Kurs        float64
	VaziZa      int
	VezniID     string
}

// KomPonudaIzvorRow is a slobodna stavka of a potvrđen zahtev za nabavku the zahtev za ponudu can take (the grid of the
// first tab, the legacy PuniSTAVKE).
type KomPonudaIzvorRow struct {
	RproID   int64
	RdokID   int64
	God      int
	Dokum    int
	Dadok    sql.NullTime
	OJ       string
	Sifra    int
	Naziv    string
	JM       string
	Kolicina float64
}

// KomPonudaIzvor is a stavka of a zahtev za nabavku a zahtev za ponudu takes (the legacy AddEPRONAR): its artikal and
// its tražena količina.
type KomPonudaIzvor struct {
	RproID   int64
	Sifra    int
	Po       int
	RsifID   int64
	Kolicina float64
}

// KomPonudaStavkaUpis are the columns of epro the save of a stavka of a zahtev za ponudu writes (the legacy AddEPRO and
// AddEPRONAR): RproID1 is the stavka of the zahtev za nabavku it was taken from (0 for a stavka entered by hand).
type KomPonudaStavkaUpis struct {
	RdokID     int64
	Rbr        int
	Sifra      int
	RsifID     int64
	Po         int
	Kolicina   float64
	Mag        int
	MagaciniID int
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
	RproID1    int64
}

// KomDobavljac is a dobavljač of a zahtev za ponudu: an analitika of fkpl with the adresa and the mesto of its partner.
type KomDobavljac struct {
	IDFkpl     int64
	Konto      string
	Sifra      string
	Naziv      string
	Adresa     string
	Mesto      string
	IDPartneri sql.NullInt64
}

// KomValuta is the valuta of a document with its kurs on the datum of the document (the legacy exit of EDT_SIFVAL: the
// srednji kurs of the kursna lista, 1 without it).
type KomValuta struct {
	SifVal int     `json:"sifval"`
	OznVal string  `json:"oznval"`
	Kurs   float64 `json:"kurs"`
	VaziZa int     `json:"vaziza"`
	Found  bool    `json:"found"`
}

// KomPonudaPrikazRow is a row of the grid of the tab "Pregled ponuda" (the legacy PuniHDRPONUDE).
type KomPonudaPrikazRow struct {
	RdokID     int64
	Magacin    string
	Vrd        int
	Dokum      int
	Dadok      sql.NullTime
	Rok        int
	Dobavljaci string
	Referent   string
	Foot       string
}

// KomPonudaStampaDobavljac is the dobavljač a printed zahtev za ponudu is sent to (the legacy PR_RPT_ZAHTEV_PONUDA with
// ipIDFKPL and ipIDPARTNER).
type KomPonudaStampaDobavljac struct {
	Konto     string
	Sifra     string
	Naziv     string
	Adresa    string
	Mesto     string
	TipPdv    int
	Pib       string
	Budzetski bool
	Jbkjs     string
	Jmbg      string
	Bpg       string
	Index     string
	Telefon   string
}

// KomPonudaStampa is what the print of a zahtev za ponudu shows (the legacy PR_RPT_ZAHTEV_PONUDA): the firma with its
// računi, the dobavljač, the header of the document and its stavke.
type KomPonudaStampa struct {
	Firma       RobnoStampaFakturaFirmaDto
	ZiroRacuni  []string
	Dobavljac   KomPonudaStampaDobavljac
	Dokum       int
	Datum       sql.NullTime
	Rok         int
	Napomena    string
	Nepotvrdjen bool
	Stavke      []KomZahtevStavkaRow
}

// KomPonudaHeaderView is the header of the screen: the saved zahtev za ponudu or a new one (RdokID 0) with its defaults.
type KomPonudaHeaderView struct {
	RdokID          int64
	Vrd             int
	Dokum           int
	Datum           time.Time
	Rok             int
	MagaciniID      int
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
	Dobavljaci      []KomDobavljac
	Potvrdjen       bool
	// Snimljen: the header is saved and locked, the stavke are open (after "Sačuvaj").
	Snimljen bool
}

// KomPonudaDokument is the content of the tab "Ponude dobavljačima": the header with the lists of its combos, the
// slobodne stavke of the zahtevi za nabavku of the magacin, the next redni broj of a stavka and a message of the screen.
type KomPonudaDokument struct {
	Header   KomPonudaHeaderView
	Vrste    []ComboItem
	Magacini []ComboItem
	Izvori   []KomPonudaIzvorRow
	NoviRbr  int
	Poruka   string
}

// KomPonudaParams are the fields of the header the screen sends ("Sačuvaj", the legacy BTN_Sacuvaj) with the stavke of
// the zahtevi za nabavku chosen in the grid of the first tab (rproid separated by ",").
type KomPonudaParams struct {
	RdokID      int64  `form:"rdokid"`
	Vrd         string `form:"vrd"`
	MagaciniID  string `form:"magaciniid"`
	Datum       string `form:"datum"`
	Rok         string `form:"rok"`
	KomSifra    string `form:"komsifra"`
	Orgjed      string `form:"idorgjed"`
	MestoTroska string `form:"mestotrid"`
	Tekst       string `form:"foot"`
	SifVal      string `form:"sifval"`
	Kurs        string `form:"kurs"`
	VaziZa      string `form:"vaziza"`
	Dobavljaci  string `form:"dobavljaci"`
	Izabrane    string `form:"izabranestavke"`
}

// KomPonudaStavkaParams are the fields of a stavka entered by hand ("Sačuvaj" of the stavka, the legacy BTN_DODAJDET;
// a rproid of a saved stavka changes its količina).
type KomPonudaStavkaParams struct {
	RdokID   int64  `form:"rdokid"`
	RproID   int64  `form:"rproid"`
	Sifra    string `form:"sifraartikla"`
	Kolicina string `form:"kolicina"`
}

// KomPonudaUI are the ids and the urls of the "Zahtev za ponudu" screen the shared RobnoFaktureUI does not have.
type KomPonudaUI struct {
	// IzvoriURL returns the slobodne stavke of the zahtevi za nabavku of a magacin into IzvoriGridID; IzabraneID is the
	// hidden list of the chosen ones (rproid separated by ",").
	IzvoriURL    string
	IzvoriGridID string
	IzabraneID   string
	// ValutaURL returns the valuta of a šifra with its kurs on a datum (JSON).
	ValutaURL string
	// PotvrdaIDFkpl and StampaIDFkpl are the hidden idfkpl of the dobavljač chosen in the dialogs of the potvrda and of
	// the print.
	PotvrdaIDFkpl string
	StampaIDFkpl  string
}

// KomPonudaPrikazParams are the filters of the tab "Pregled ponuda": the vrsta of the first tab, the magacin, the
// status (nepotvrđeni or potvrđeni) and all the magacini of the user.
type KomPonudaPrikazParams struct {
	Vrd         int
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	// Tekst is the text of the search of the main screen (KomZahtevPrikazParams.Tekst).
	Tekst string
}

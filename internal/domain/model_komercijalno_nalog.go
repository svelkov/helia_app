package domain

import "database/sql"

// The statuses of a nalog za isporuku (edok.doksts; the legacy "Unet dokument, nepotvrdjen nalog", "Potvrdjen nalog,
// spreman za otpremu" and "Formirana otpremnica, faktura").
const (
	KomNalogNov        = "10"
	KomNalogPotvrdjen  = "20"
	KomNalogOtpremljen = "40"
)

// The "Nalozi za isporuku" screen of the komercijalno (the legacy pr_PorudzbenicaProdaja with its stavke
// pr_PorudzbenicaProdaja_DETALJI, the porudžbenice od kupaca): a document of edok of the vrste of the modul PRO and the
// group NRK in a magacin of the tipovi G, V and S, a variant of the profaktura (the same header, kupac, artikal and
// stavke) whose stavke reserve the zalihe of their lots (drsta.rez and rsta.rez). It may be made of a potvrđena
// profaktura (vrdokid), whose stavke it takes. Its statuses: 10 new, 20 confirmed (ready for the otprema), 40 the
// otpremnica and the faktura formed from it.

// KomNalogDodatno is what a nalog za isporuku keeps besides the fields of the profaktura: its profaktura (vrdokid with
// its broj and datum), the bank of the payment of an izvoz (sifbank, with the bnkizv of the šifra and its opis) and the
// broj of its otpremnica (brotp) and of its faktura (the rdok formed from it, rdok.edokid).
type KomNalogDodatno struct {
	VrdokID    int64
	ProDokum   int
	ProDadok   sql.NullTime
	SifBank    int
	BnkizvID   int64
	BnkDes     string
	Otpremnica int
	Faktura    int
}

// KomNalogProfaktura is a potvrđena profaktura a nalog is made of (the legacy BTN_TraziPROFAK and PodaciIzProfakture):
// what the header takes of it.
type KomNalogProfaktura struct {
	RdokID          int64        `json:"rdokid"`
	Dokum           int          `json:"dokum"`
	Dadok           sql.NullTime `json:"-"`
	Datum           string       `json:"datum"`
	MagaciniID      int          `json:"-"`
	Rok             int          `json:"rok"`
	Fkto            string       `json:"fkto"`
	Fana            string       `json:"fana"`
	Foot            string       `json:"foot"`
	SifVal          int          `json:"sifval"`
	OznVal          string       `json:"oznval"`
	VaziZa          int          `json:"vaziza"`
	Orgjed          int64        `json:"idorgjed"`
	OrgjedText      string       `json:"ojtext"`
	MestoTroska     int64        `json:"mestotrid"`
	MestoTroskaText string       `json:"mttext"`
	Ugrabat         float64      `json:"ugrabat"`
	Pkase           float64      `json:"pkase"`
	Posuslnab       string       `json:"posuslnab"`
	Keykat          string       `json:"keykat"`
	Category        string       `json:"category"`
}

// KomNalogSaldo is the finansijsko stanje of a kupac on a datum (the legacy SaldoKUPCA): the proknjižena realizacija
// (the fakture and the uplate of fpro), the neproknjižena one (the robne fakture not posted, with their PDV) and their
// sums.
type KomNalogSaldo struct {
	Realizacija    float64 `json:"realizacija"`
	Placeno        float64 `json:"placeno"`
	Saldo          float64 `json:"saldo"`
	Dospelo        float64 `json:"dospelo"`
	Nedospelo      float64 `json:"nedospelo"`
	RealizacijaRob float64 `json:"realizacijarob"`
	DospeloRob     float64 `json:"dospelorob"`
	NedospeloRob   float64 `json:"nedospelorob"`
	UkDospelo      float64 `json:"ukdospelo"`
	UkNedospelo    float64 `json:"uknedospelo"`
	UkSaldo        float64 `json:"uksaldo"`
}

// KomNalogFpro is a promet of a kupac in the finansijsko knjigovodstvo (fpro of the vrste 10, 30 and 80): its kategorija
// (1 and 2 the fakture, 3 and 4 the uplate), its iznos and its dospeće (the datum and the rok).
type KomNalogFpro struct {
	Kat   int          `db:"kat"`
	Iznos float64      `db:"iznos"`
	Dadok sql.NullTime `db:"dadok"`
	Rok   int          `db:"rok"`
}

// KomNalogRobnaStavka is a stavka of a robna faktura of a kupac not posted yet (rdok of the vrste FAK, FPR and KPK with
// its rpro): the document with its rok, ugovoreni rabat and kasa, the stavka with its količina, cena, rabat and the stopa
// PDV of its tarifa on its datum.
type KomNalogRobnaStavka struct {
	RdokID  int64        `db:"rdokid"`
	Dadok   sql.NullTime `db:"dadok"`
	Rok     int          `db:"rok"`
	Ugrabat float64      `db:"ugrabat"`
	Pkase   float64      `db:"pkase"`
	Kolic   float64      `db:"kolic"`
	Fcena   float64      `db:"fcena"`
	Rab     float64      `db:"rab"`
	Stopa   float64      `db:"stopa"`
}

// KomNalogLot is a lot (drsta) of an artikal in a konto of a magacin with its slobodna količina (ulaz - izlaz -
// rezervisano), the legacy "Skidaj zalihe po starosti".
type KomNalogLot struct {
	Konto    string  `db:"konto"`
	Otk      string  `db:"otk"`
	Serija   string  `db:"serija"`
	Roktr    int64   `db:"roktr"`
	Slobodno float64 `db:"slobodno"`
}

// KomNalogRezervacija is the reservation of the zalihe of a stavka of a nalog: the lot of the artikal in the magacin
// and the količina (negative: released).
type KomNalogRezervacija struct {
	MagaciniID int
	Mag        int
	RsifID     int64
	Sifra      int
	Konto      string
	Otk        string
	Serija     string
	Roktr      int64
	Cena       float64
	Kolicina   float64
}

// KomNalogPrikazRow is a row of the lists of the main screen (the legacy PR_QRY_EDOK of the tab "Pregled naloga za
// isporuku" and PuniTABLEHDR of the tab "Pregled delimično realizovanih naloga").
type KomNalogPrikazRow struct {
	RdokID     int64        `db:"rdokid"`
	Magacin    string       `db:"magacin"`
	Vrd        int          `db:"vrd"`
	Dokum      int          `db:"dokum"`
	Dadok      sql.NullTime `db:"dadok"`
	Otpremnica int          `db:"otpremnica"`
	Faktura    int          `db:"faktura"`
	Fkto       string       `db:"fkto"`
	Fana       string       `db:"fana"`
	Kupac      string       `db:"kupac"`
	Rok        int          `db:"rok"`
	Ugrabat    float64      `db:"ugrabat"`
	Doksts     string       `db:"doksts"`
}

// KomNalogPrikazParams are the filters of the lists of the main screen: the magacin (or all the magacini of the user),
// the status (the nepotvrđeni 10, else the potvrđeni 20 and 40; the delimično realizovani are 40 with a stavka not
// delivered fully) and the text of the search.
type KomNalogPrikazParams struct {
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	Delimicni   bool
	Tekst       string
}

// KomNalogDokument is the content of the entry dialog: the one of the profaktura with what the nalog adds.
type KomNalogDokument struct {
	KomProDokument
	Dodatno    KomNalogDodatno
	Saldo      KomNalogSaldo
	StatusOpis string
}

// KomNalogParams are the fields of the header the screen sends ("Sačuvaj"): the ones of the profaktura, the profaktura
// the nalog is made of and the bank of the payment of an izvoz.
type KomNalogParams struct {
	KomProParams
	ProfakturaID string `form:"profakturaid"`
	BnkizvID     string `form:"bnkizvid"`
}

// KomNalogStavkaParams are the fields of a stavka: the ones of the profaktura and "Skidaj zalihe po starosti".
type KomNalogStavkaParams struct {
	KomProStavkaParams
	PoStarosti string `form:"postarosti"`
}

// KomNalogStampa is what the print of a nalog shows (the legacy PR_RPT_NALOGPRODAJA): the firma, the kupac, the
// magacin, the org. jedinica, the tekući račun, the broj and the datum, the stavke, the klasa cena, the napomena and the
// težina and the zapremina; Nepotvrdjen marks a new one.
type KomNalogStampa struct {
	Firma         RobnoStampaFakturaFirmaDto
	Kupac         KomPonudaStampaDobavljac
	Magacin       string
	MagacinAdresa string
	OJ            string
	TekuciRacun   string
	Dokum         int
	Datum         sql.NullTime
	KlasaCena     string
	Napomena      string
	Tezina        float64
	Zapremina     float64
	Lot           bool
	Nepotvrdjen   bool
	Stavke        []KomNalogStampaStavka
}

// KomNalogStampaStavka is a printed stavka of a nalog.
type KomNalogStampaStavka struct {
	Rbr      int
	Sifra    int
	Opis     string
	Otk      string
	Serija   string
	Rok      string
	JM       string
	Kolicina float64
	Decimale int
}

// KomNalogUI are the urls of the entry dialog of a nalog the profaktura does not have (JSON): the profaktura a nalog is
// made of and the finansijsko stanje of the kupac.
type KomNalogUI struct {
	ProfakturaURL string
	SaldoURL      string
}

package domain

import (
	"database/sql"
	"time"
)

// ---------------------------------------------------------------------------------------------------
// Kalkulacije veleprodaje (the group PLT, the legacy rob_Kalkulacije_Veleprodaja with its stavke
// rob_Kalkulacije_Veleprodaja_Stavke and its zavisni troškovi rob_Zavisni_Troskovi): a robni dokument (rdok)
// of a dobavljač (rdok.fkto, rdok.fana) whose stavke enter the roba at its nabavna cena (the fakturna cena less
// the rabat with the zavisni troškovi) and set its veleprodajna and maloprodajna cena; the porez and the avans
// paid to the dobavljač (rppo, appo) and the zavisni troškovi of other dobavljači (rztr with its porezi in rppo)
// ---------------------------------------------------------------------------------------------------

// RobKalPorez is a row of the grids "Plaćen porez" (rppo) and "Plaćen avans" (appo) of a kalkulacija and of the
// grid "Plaćen porez" of a zavisni trošak (rppo of the rztr): the osnovica, the porez paid, the poreska oznaka
// with its stopa and its rpor row, and (of an avans) the broj and the datum of the avansni račun.
type RobKalPorez struct {
	ID       int64        `db:"id" json:"id"`
	Osnovica float64      `db:"osn" json:"osn"`
	Porez    float64      `db:"ppor" json:"ppor"`
	Po       int          `db:"po" json:"po"`
	Stopa    float64      `db:"pdv" json:"pdv"`
	RporID   int64        `db:"rporid" json:"rporid"`
	Rac      string       `db:"rac" json:"rac"`
	Datum    sql.NullTime `db:"dadok" json:"-"`
}

// RobKalPorezUnos is a row of a grid of the porezi as the form sends it (the texts of the fields) and as it is
// saved: the poreska oznaka with its stopa and its rpor row on the datum, the broj and the datum of an avans.
type RobKalPorezUnos struct {
	Osnovica, Porez, Stopa float64
	Po                     int
	RporID                 int64
	Rac                    string
	Datum                  sql.NullTime
}

// RobnoKalHeaderView is the header of a kalkulacija veleprodaje: the fields it shares with the faktura (the
// dobavljač is the kupac of the faktura: KupacKonto, KupacSifra, KupacNaziv..., the mesto nabavke its mesto
// isporuke, the valuta, the POPDV polje, the org. jedinica and the mesto troška), the opis of its vrsta, the
// EPP polje (the checkbox EPP and the polje of fsepp), "PDV se može odbiti" (rdok.odbitni), the veza
// porudžbenice (the narudžbenica of edok: rdok.edokid, with its broj and datum), the plaćen porez and the plaćen
// avans and whether it has zavisni troškovi.
type RobnoKalHeaderView struct {
	RobnoFaktureHeaderView
	VrstaOpis   string
	Epp         bool
	FseppID     sql.NullInt64
	FseppText   string
	Odbitni     bool
	VezaID      int64
	VezaBroj    string
	VezaDatum   time.Time
	Porezi      []RobKalPorez
	Avansi      []RobKalPorez
	ImaTroskove bool
}

// RobnoKalStavka is the entry of a stavka of a kalkulacija veleprodaje: the document, the konto magacina (the
// first konto of the magacin), the tip zaliha of the magacin (Tipzal 2: the OTK broj, the serija and the rok),
// the podaci o dokumentu it repeats (read only, with the valuta of the document: a devizni document has the
// fakturna cena in its valuta) and the procenat of the external zavisni troškovi of the document (the legacy
// ukZTRO, IzracunajEKSTZAVTROSAK).
type RobnoKalStavka struct {
	RdokID                    int64
	Konto                     string
	Tipzal                    int
	Vrd, Brdok, DatumDokument string
	Valuta                    string
	Kurs                      float64
	VaziZa                    int
	Devizni                   bool
	ProcenatTroskova          float64
}

// RobnoKalTrosak is the entry of a zavisni trošak of a kalkulacija (the legacy rob_Zavisni_Troskovi): the
// document, the default konto of the dobavljač (the konto of the dobavljači of the robno, rvr.fdob; after a save the
// konto of the last one entered) and the default datum of the document of the dobavljač (the datum dokumenta).
type RobnoKalTrosak struct {
	RdokID          int64
	KontoDobavljaca string
	DatumDokumenta  time.Time
}

// RobnoKalUI are the ids and the urls of the "Kalkulacije veleprodaje" screen besides RobnoFaktureUI: the collapsible
// control of the zavisni troškovi, the url of the stopa of a poreska oznaka on a datum (the grids of the porezi), of a
// saved zavisni trošak and of a saved stavka (for their forms).
type RobnoKalUI struct {
	TroskoviPanelID string
	StopaURL        string
	TrosakURL       string
	StavkaURL       string
}

// RobnoKalDokument is the document of the "Kalkulacije veleprodaje" screen: the header, the entry of a zavisni
// trošak and of a stavka, and the message of the screen (e.g. why the selected document cannot be changed).
type RobnoKalDokument struct {
	Header RobnoKalHeaderView
	Trosak RobnoKalTrosak
	Stavka RobnoKalStavka
	Poruka string
}

// RobnoKalParams are the values of the header of a kalkulacija veleprodaje (the legacy ValidacijaRDOK and
// NoviSLOGRDOK) by the names of the fields of the form, plus the nalog, the magacin and the document they belong
// to, the rows of the grids of the plaćen porez and avans (the fields of a row by their index) and Potvrda: the
// question of the validation (an izvorni dokument of the dobavljač that exists) is answered "Da".
type RobnoKalParams struct {
	RnalID int64 `form:"rnalid"`
	RdokID int64 `form:"rdokid"`
	// MagaciniID is the magacin of the nalog (set by the handler from the rnalid, not sent by the form).
	MagaciniID int `form:"-"`

	Vrd             string `form:"vrd"`
	BrojDokum       string `form:"brdok"`
	DatumDokum      string `form:"datumfakturisanja"`
	IzvorniDokum    string `form:"brdokorg"`
	IzvorniDokDatum string `form:"datumizvora"`
	Opis            string `form:"opis"`
	// The veza porudžbenice: the narudžbenica (edok.rdokid).
	Veza string `form:"edokid"`
	// The uslovi nabavke robe.
	NacinDopremanja string `form:"nacindopremanja"`
	UsloviPlacanja  string `form:"usloviplacanja"`
	Rok             string `form:"rok"`
	UsloviKamate    string `form:"uslovikamate"`
	Tekst           string `form:"foot"`
	NapomenaPoresko string `form:"napomporeskoosl"`
	// The dobavljač (rdok.fkto, rdok.fana), its mesto nabavke (fispid), the valuta.
	DobavljacKonto string `form:"kupackonto"`
	DobavljacSifra string `form:"kupacsifra"`
	MestoNabavke   string `form:"fispid"`
	Valuta         string `form:"idvalute"`
	Kurs           string `form:"kurs"`
	VaziZa         string `form:"vaziza"`
	Orgjed         string `form:"idorgjed"`
	MestoTroska    string `form:"mestotrid"`
	Epp            bool   `form:"epp"`
	Fsepp          string `form:"fseppid"`
	Popdv          string `form:"popdv"`
	Odbitni        bool   `form:"odbitni"`
	// The rows of the grid "Plaćen porez": osnovica, plaćen porez, poreska oznaka.
	PorezOsn  []string `form:"porezosn"`
	PorezPpor []string `form:"porezppor"`
	PorezPo   []string `form:"porezpo"`
	// The rows of the grid "Plaćen avans": osnovica, porez, poreska oznaka, broj and datum of the avansni račun.
	AvansOsn   []string `form:"avansosn"`
	AvansPpor  []string `form:"avansppor"`
	AvansPo    []string `form:"avanspo"`
	AvansRac   []string `form:"avansrac"`
	AvansDatum []string `form:"avansdatum"`
	Potvrda    bool     `form:"potvrda"`
}

// RobKalArtikal is the artikal of a stavka of a kalkulacija with what the form of the stavka shows of it (the legacy
// exit of EDT_SIFRA): the artikal (its konto when it has one), its tarifa with its stopa on the datum of the
// document, and from its stanje in the konto of the magacin (rsta) the stanje, the cene of the last kalkulacija
// (fakturna cena, rabat, nabavna cena, the marže, the veleprodajna and the maloprodajna cena) and the stara
// veleprodajna and maloprodajna cena.
type RobKalArtikal struct {
	Sifra       int     `json:"sifra"`
	Naziv       string  `json:"naziv"`
	Jm          string  `json:"jm"`
	Konto       string  `json:"konto"`
	Tarifa      int     `json:"tarifa"`
	Stopa       float64 `json:"stopa"`
	Stanje      float64 `json:"stanje"`
	Fcena       float64 `json:"fcena"`
	Rab         float64 `json:"rab"`
	Ncena       float64 `json:"ncena"`
	Vma         float64 `json:"vma"`
	Cena        float64 `json:"cena"`
	Mma         float64 `json:"mma"`
	Mcenap      float64 `json:"mcenap"`
	StaraCena   float64 `json:"staracena"`
	StaraMcenap float64 `json:"staramcenap"`
}

// RobKalStavkaParams are the values of a stavka of a kalkulacija by the names of the fields of the form (RproID 0:
// a new stavka): the konto, the artikal, the količina, the fakturna cena (in the valuta of a devizni document) and
// the rabat, the zavisni trošak and the interni zavisni trošak with their oznake ("-", "%", "d"), the veleprodajna
// marža and cena, the maloprodajna marža and cena with PDV, the lot; Potvrda: the question of the validation (no
// OTK broj in a magacin of the tip zaliha 2) is answered "Da".
type RobKalStavkaParams struct {
	RdokID  int64  `form:"rdokid"`
	RproID  int64  `form:"rproid"`
	Konto   string `form:"kontostavke"`
	Sifra   string `form:"sifraartikla"`
	Kolic   string `form:"kolicina"`
	Fcena   string `form:"fakturnacena"`
	Rabat   string `form:"rabatstavke"`
	Tz      string `form:"tz"`
	Ztro    string `form:"ztro"`
	Tzi     string `form:"tzi"`
	Ztrin   string `form:"ztrin"`
	Vma     string `form:"vma"`
	Cena    string `form:"vpcena"`
	Mma     string `form:"mma"`
	Mcenap  string `form:"mpcena"`
	Otk     string `form:"otk"`
	Serija  string `form:"serija"`
	Rok     string `form:"rokstavke"`
	Potvrda bool   `form:"potvrda"`
}

// RobKalStavkaForma is a saved stavka of a kalkulacija in the form of the stavke ("Izmeni", the legacy BTN_Izmeni):
// its values as the form shows them with the artikal, its tarifa and stopa and its stanje.
type RobKalStavkaForma struct {
	RproID   int64   `json:"rproid"`
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	Jm       string  `json:"jm"`
	Konto    string  `json:"konto"`
	Tarifa   int     `json:"tarifa"`
	Stopa    float64 `json:"stopa"`
	Stanje   float64 `json:"stanje"`
	Kolicina float64 `json:"kolicina"`
	Fcena    float64 `json:"fcena"`
	Cenaval  float64 `json:"cenaval"`
	Rab      float64 `json:"rab"`
	Ncena    float64 `json:"ncena"`
	Tz       string  `json:"tz"`
	Ztro     float64 `json:"ztro"`
	Tzi      string  `json:"tzi"`
	Ztrin    float64 `json:"ztrin"`
	Vma      float64 `json:"vma"`
	Cena     float64 `json:"cena"`
	Mma      float64 `json:"mma"`
	Mcenap   float64 `json:"mcenap"`
	Otk      string  `json:"otk"`
	Serija   string  `json:"serija"`
	Rok      string  `json:"rok"`
}

// RobKalStavkaDto is a saved stavka of a kalkulacija as its form and its taking back need it: the artikal, the
// konto, the količina, the iznos, the cene of the kalkulacija and the lot.
type RobKalStavkaDto struct {
	RproID   int64   `db:"rproid"`
	Rbr      int     `db:"rbr"`
	Sifra    int     `db:"sifra"`
	Konto    string  `db:"konto"`
	Kolicina float64 `db:"kolic"`
	Iznos    float64 `db:"iznos"`
	Cena     float64 `db:"cena"`
	Fcena    float64 `db:"fcena"`
	Cenaval  float64 `db:"cenaval"`
	Rab      float64 `db:"rab"`
	Ncena    float64 `db:"ncena"`
	Mcena    float64 `db:"mcena"`
	Mcenap   float64 `db:"mcenap"`
	Vma      float64 `db:"vma"`
	Mma      float64 `db:"mma"`
	Ztro     float64 `db:"ztro"`
	Ztrin    float64 `db:"ztrin"`
	Tz       string  `db:"tz"`
	Tzi      string  `db:"tzi"`
	Otk      string  `db:"otk"`
	Serija   string  `db:"serija"`
	Roktr    float64 `db:"roktr"`
	// Naz1, Jm and Po are the naziv, the JM and the poreska oznaka of the stavka (of the šifra 0 entered with it).
	Naz1 string `db:"naz1"`
	Jm   string `db:"jm"`
	Po   int    `db:"po"`
}

// RobKalRstaCene is the stanje of an artikal in a konto of a magacin with the cene of the kalkulacija.
type RobKalRstaCene struct {
	Stanje float64 `db:"stanje"`
	Cena   float64 `db:"cena"`
	Fcena  float64 `db:"fcena"`
	Rab    float64 `db:"rab"`
	Ncena  float64 `db:"ncena"`
	Vma    float64 `db:"vma"`
	Mma    float64 `db:"mma"`
	Mcenap float64 `db:"mcenap"`
}

// RobKalZalihaBrisanja is an artikal whose stanje would be negative after the deletion of a kalkulacija (the legacy
// check of BTN_Brisanje): its šifra, naziv and the stanje after the deletion.
type RobKalZalihaBrisanja struct {
	Sifra  int     `db:"sifra"`
	Naziv  string  `db:"naziv"`
	Stanje float64 `db:"stanje"`
}

// RobKalTrosakRow is a zavisni trošak of a kalkulacija (rztr) in its grid and in its form: the dobavljač with its
// naziv, the document of the dobavljač (broj, datum, rok), the opis, the napomena o porezu, "PDV plaćen pri uvozu",
// the POPDV polje, the fakturna vrednost robe, the iznos of the trošak and the konto troška with its naziv.
type RobKalTrosakRow struct {
	RztrID      int64        `db:"rztrid" json:"rztrid"`
	Fkto        string       `db:"fkto" json:"fkto"`
	Fana        string       `db:"fana" json:"fana"`
	Naziv       string       `db:"naziv" json:"naziv"`
	Dokdob      string       `db:"dokdob" json:"dokdob"`
	Dadok       sql.NullTime `db:"dadok" json:"-"`
	Rok         int          `db:"rok" json:"rok"`
	Opis        string       `db:"opis" json:"opis"`
	Pornapomena string       `db:"pornapomena" json:"pornapomena"`
	Flguvozpdv  bool         `db:"flguvozpdv" json:"flguvozpdv"`
	Polje       string       `db:"polje" json:"polje"`
	Iznfak      float64      `db:"iznfak" json:"iznfak"`
	Iznos       float64      `db:"iznos" json:"iznos"`
	Tkonto      string       `db:"tkonto" json:"tkonto"`
	TkontoNaziv string       `db:"tkontonaziv" json:"tkontonaziv"`
}

// RobKalTrosakForma is a saved zavisni trošak for its form ("Izmeni" of its row): the trošak with its datum as the
// form shows it, the data of the dobavljač and its plaćen porez.
type RobKalTrosakForma struct {
	RobKalTrosakRow
	Datum              string        `json:"dadok"`
	KupacAdresa        string        `json:"adresa"`
	KupacPostanskiBroj string        `json:"pobro"`
	KupacMesto         string        `json:"mesto"`
	KupacPib           string        `json:"pib"`
	SistemPdv          string        `json:"sistempdv"`
	Porezi             []RobKalPorez `json:"porezi"`
}

// RobKalTrosakParams are the values of the form of a zavisni trošak by the names of the fields of the form (RztrID
// 0: a new one) with the rows of its grid "Plaćen porez".
type RobKalTrosakParams struct {
	RdokID      int64    `form:"rdokid"`
	RztrID      int64    `form:"rztrid"`
	Konto       string   `form:"trkonto"`
	Sifra       string   `form:"trsifra"`
	Dokdob      string   `form:"trdokdob"`
	Dadok       string   `form:"trdadok"`
	Rok         string   `form:"trrok"`
	Opis        string   `form:"tropis"`
	Pornapomena string   `form:"trnapomena"`
	Flguvozpdv  bool     `form:"truvozpdv"`
	Polje       string   `form:"trpopdv"`
	Iznfak      string   `form:"triznfak"`
	Iznos       string   `form:"triznos"`
	Tkonto      string   `form:"trtkonto"`
	PorezOsn    []string `form:"trporezosn"`
	PorezPpor   []string `form:"trporezppor"`
	PorezPo     []string `form:"trporezpo"`
}

package domain

import "database/sql"

// ---------------------------------------------------------------------------------------------------
// Kalkulacije maloprodaje (the group KAL, the legacy rob_Kalkulacije_Maloprodaja with its stavke
// rob_Kalkulacije_Maloprodaja_Stavke): a robni dokument (rdok) of a dobavljač (rdok.fkto, rdok.fana) for a prodavnica
// (rdok.pkto, rdok.pana: the konto of its stanje) whose stavke enter the roba at the maloprodajna cena with the PDV:
// the fakturna cena less the rabat with the zavisni trošak is the nabavna cena, with the maloprodajna marža and the
// porez the MP cena; the porez paid to the dobavljač (rppo)
// ---------------------------------------------------------------------------------------------------

// RobnoKmpHeaderView is the header of a kalkulacija maloprodaje: the fields it shares with the faktura (the dobavljač
// is the kupac of the faktura: KupacKonto, KupacSifra, KupacNaziv..., the POPDV polje, the org. jedinica and the mesto
// troška), the opis of its vrsta, the prodavnica (the konto and the šifra of the kontni plan with its opis), the EPP
// polje, "PDV se može odbiti", the plaćen porez and whether the vrsta asks the POPDV polje (the legacy GR_POPDV).
type RobnoKmpHeaderView struct {
	RobnoFaktureHeaderView
	VrstaOpis       string
	ProdavnicaKonto string
	ProdavnicaSifra string
	ProdavnicaOpis  string
	Epp             bool
	FseppID         sql.NullInt64
	FseppText       string
	Odbitni         bool
	Porezi          []RobKalPorez
	TraziPopdv      bool
}

// RobnoKmpStavka is the entry of a stavka of a kalkulacija maloprodaje: the document, the konto and the šifra of the
// prodavnica (the konto of the stavke and of their stanje, read only), the podaci o dokumentu it repeats (read only)
// and whether the magacin is a magacin of the maloprodaja (magacini.tipmag "M": only it has stavke with an artikal,
// another magacin only the šifra 0 with its naziv).
type RobnoKmpStavka struct {
	RdokID                    int64
	Konto, Pana               string
	Vrd, Brdok, DatumDokument string
	Maloprodaja               bool
}

// RobnoKmpDokument is the document of the "Kalkulacije maloprodaje" screen: the header, the entry of a stavka and the
// message of the screen (e.g. why the selected document cannot be changed).
type RobnoKmpDokument struct {
	Header RobnoKmpHeaderView
	Stavka RobnoKmpStavka
	Poruka string
}

// RobnoKmpParams are the values of the header of a kalkulacija maloprodaje (the legacy ValidacijaRDOK and
// NoviSLOGRDOK) by the names of the fields of the form, plus the nalog, the magacin and the document they belong to,
// the rows of the grid of the plaćen porez (the fields of a row by their index) and Potvrda: the question of the
// validation (an izvorni dokument of the dobavljač that exists) is answered "Da".
type RobnoKmpParams struct {
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
	// The uslovi nabavke robe.
	UsloviPlacanja  string `form:"usloviplacanja"`
	Rok             string `form:"rok"`
	Tekst           string `form:"foot"`
	NapomenaPoresko string `form:"napomporeskoosl"`
	// The dobavljač (rdok.fkto, rdok.fana) and the prodavnica (rdok.pkto, rdok.pana).
	DobavljacKonto  string `form:"kupackonto"`
	DobavljacSifra  string `form:"kupacsifra"`
	ProdavnicaKonto string `form:"pkto"`
	ProdavnicaSifra string `form:"pana"`
	Orgjed          string `form:"idorgjed"`
	MestoTroska     string `form:"mestotrid"`
	Epp             bool   `form:"epp"`
	Fsepp           string `form:"fseppid"`
	Popdv           string `form:"popdv"`
	Odbitni         bool   `form:"odbitni"`
	// The rows of the grid "Plaćen porez": osnovica, plaćen porez, poreska oznaka.
	PorezOsn  []string `form:"porezosn"`
	PorezPpor []string `form:"porezppor"`
	PorezPo   []string `form:"porezpo"`
	Potvrda   bool     `form:"potvrda"`
}

// RobKmpArtikal is the artikal of a stavka of a kalkulacija maloprodaje with what the form shows of it (the legacy exit
// of EDT_SIFRA): the artikal with its naziv, JM, tarifa with its stopa on the datum of the document, its stanje in the
// konto of the prodavnica and its last maloprodajna cena with the PDV (rcene.mcenap, the poslednja cena); Poruka is the
// message of an artikal that cannot be entered (an artikal in a magacin that is not of the maloprodaja).
type RobKmpArtikal struct {
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	Jm       string  `json:"jm"`
	Tarifa   int     `json:"tarifa"`
	Stopa    float64 `json:"stopa"`
	Stanje   float64 `json:"stanje"`
	PoslCena float64 `json:"poslcena"`
	Poruka   string  `json:"poruka"`
}

// RobKmpStavkaParams are the values of a stavka of a kalkulacija maloprodaje by the names of the fields of the form
// (RproID 0: a new stavka): the artikal (the šifra 0 with its naziv, JM and tarifa), the količina, the fakturna cena,
// the rabat, the zavisni trošak with its oznaka ("-", "%", "d"), the maloprodajna marža and the MP cena with the PDV.
type RobKmpStavkaParams struct {
	RdokID int64  `form:"rdokid"`
	RproID int64  `form:"rproid"`
	Sifra  string `form:"sifraartikla"`
	Naziv  string `form:"nazivkmp"`
	Jm     string `form:"jm"`
	Tarifa string `form:"tarifa"`
	Kolic  string `form:"kolicina"`
	Fcena  string `form:"fakturnacena"`
	Rabat  string `form:"rabatstavke"`
	Tz     string `form:"tz"`
	Ztro   string `form:"ztro"`
	Mma    string `form:"mma"`
	Mcenap string `form:"mpcena"`
}

// RobKmpStavkaForma is a saved stavka of a kalkulacija maloprodaje in the form of the stavke ("Izmeni" of its row): its
// values as the form shows them, with the stopa of its tarifa, the stanje and the poslednja cena of the artikal.
type RobKmpStavkaForma struct {
	RproID   int64   `json:"rproid"`
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	Jm       string  `json:"jm"`
	Tarifa   int     `json:"tarifa"`
	Stopa    float64 `json:"stopa"`
	Stanje   float64 `json:"stanje"`
	PoslCena float64 `json:"poslcena"`
	Kolicina float64 `json:"kolicina"`
	Fcena    float64 `json:"fcena"`
	Rab      float64 `json:"rab"`
	Ncena    float64 `json:"ncena"`
	Tz       string  `json:"tz"`
	Ztro     float64 `json:"ztro"`
	Mma      float64 `json:"mma"`
	Mcenap   float64 `json:"mcenap"`
}

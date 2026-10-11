package domain

// The "Interni prenosi - dostavnice" screen (the legacy rob_Interni_racuni, the vrste dokumenata of the
// groups IRT and IRP): the header and the stavke of an interni prenos. They share most of their fields
// with the faktura veleprodaje (the same rdok and rpro), so the views embed the ones of the faktura and add
// only the fields of the interni prenos.

// RobnoInterniHeaderView is the header of an interni prenos: the fields it shares with the faktura (the
// strip of the nalog, the podaci o dokumentu, the kupac, the uslovi prodaje, the org. jedinica and the
// mesto troška) and the prodavnica (rdok.pkto and rdok.pana, the partner the goods are moved to, with its
// naziv, adresa, poštanski broj and mesto).
type RobnoInterniHeaderView struct {
	RobnoFaktureHeaderView
	ProdavnicaKonto, ProdavnicaSifra, ProdavnicaNaziv string
	ProdavnicaAdresa, ProdavnicaPostanskiBroj         string
	ProdavnicaMesto                                   string
	// StampajRacun is "Štampaj račun za kupca".
	StampajRacun bool
}

// RobnoInterniStavka is the entry form of one stavka of an interni prenos: the fields it shares with the
// stavka of the faktura and the ones of the maloprodaja (the prodajni objekat, the vrednosti, the marža and
// the maloprodajne cene) with the last maloprodajna cena of the artikal.
type RobnoInterniStavka struct {
	RobnoFaktureStavka
	ProdajniObjekat                          string
	Vrednost, NabavnaVrednost, MPVrednostPDV string
	MaloprodajnaMarza, MaloprodajnaCena      string
	MPCenaSaPorezom                          string
	// "Zadnja cena po kojoj je uziman ovaj artikal" (with ZadnjiRabat and ZadnjiDatum of the faktura).
	ZadnjaMPCena, ZadnjaMPCenaSaPorezom, ZadnjaMarza string
	// Grupa is the group of the vrsta of the document (IRT, IRP, IRS): the calculations of the form differ by it.
	Grupa string
}

// RobnoInterniDokument is the document of the "Interni prenosi - dostavnice" screen: the header, the entry
// of a stavka with the data of the document it repeats, and the message of the screen (e.g. why the
// selected document cannot be changed).
type RobnoInterniDokument struct {
	Header RobnoInterniHeaderView
	Stavka RobnoInterniStavka
	Poruka string
}

// RobnoInterniParams are the values of the header of an interni prenos (the legacy ValidacijaRDOK and
// SnimiRDOK of rob_Interni_racuni) by the names of the fields of the form, plus the nalog, the magacin and
// the document they belong to. The values are the texts of the form: the validation parses them itself, so
// a wrong value is a field error and not a binding error.
type RobnoInterniParams struct {
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
	// The prodavnica (rdok.pkto, rdok.pana) and the kupac (rdok.fkto, rdok.fana).
	ProdavnicaKonto string `form:"pkto"`
	ProdavnicaSifra string `form:"pana"`
	KupacKonto      string `form:"kupackonto"`
	KupacSifra      string `form:"kupacsifra"`
	Orgjed          string `form:"idorgjed"`
	MestoTroska     string `form:"mestotrid"`
	NapomenaPoresko string `form:"napomporeskoosl"`
	NacinOtpreme    string `form:"nacinotpreme"`
	UsloviPlacanja  string `form:"usloviplacanja"`
	Rok             string `form:"rok"`
	UsloviKamate    string `form:"uslovikamate"`
	Tekst           string `form:"foot"`
}

// RobInterniArtikal is the artikal of a stavka of an interni prenos with what the form of the stavka shows
// of it (the legacy exit of EDT_SIFRA and EDT_KONTO of rob_InterniRacuni_DET): the artikal, its tarifa with
// its stopa on the datum of the document, its model and its taksa (rcene.itaksa); its stanje and cene in the
// konto of the magacin (rsta: for an IRT the cena, the rabat, the nabavna cena, the marža and the
// maloprodajne cene; MaloprodajnaSaPorezom also for the model "D"); the last cena of the prodavnica
// (ZadnjaCena).
type RobInterniArtikal struct {
	Sifra  int     `json:"sifra"`
	Naziv  string  `json:"naziv"`
	Jm     string  `json:"jm"`
	Tarifa int     `json:"tarifa"`
	Stopa  float64 `json:"stopa"`
	Model  string  `json:"model"`
	Taksa  float64 `json:"taksa"`
	Stanje float64 `json:"stanje"`
	// Cene: set (Cene true) when the stanje of the artikal exists and the document is an IRT.
	Cene   bool    `json:"cene"`
	Cena   float64 `json:"cena"`
	Vra    float64 `json:"vra"`
	Ncena  float64 `json:"ncena"`
	Mma    float64 `json:"mma"`
	Mcena  float64 `json:"mcena"`
	Mcenap float64 `json:"mcenap"`
	// McenapModel: the maloprodajna cena with the porez of an artikal of the model "D" (rsta.mcenap).
	McenapModel bool `json:"mcenapmodel"`
	// Grupa is the group of the vrsta of the document (IRT, IRP, IRS); Rsta tells whether the artikal has its
	// stanje in the konto, with its cena, rabat (vma * 100 / (vma + 100)), marža and maloprodajna cena with the
	// porez as they are in rsta (the legacy exit of EDT_KONTO takes them).
	Grupa      string  `json:"grupa"`
	Rsta       bool    `json:"rsta"`
	RstaCena   float64 `json:"rstacena"`
	RstaVra    float64 `json:"rstavra"`
	RstaMma    float64 `json:"rstamma"`
	RstaMcenap float64 `json:"rstamcenap"`
	// The last cena of the prodavnica (ZadnjaCena): its maloprodajna cena, rabat, cena with the porez,
	// marža and datum; Poruka when the prodavnica has not taken the artikal.
	ZadnjaMcena  float64 `json:"zadnjamcena"`
	ZadnjiVra    float64 `json:"zadnjivra"`
	ZadnjaMcenap float64 `json:"zadnjamcenap"`
	ZadnjaMma    float64 `json:"zadnjamma"`
	ZadnjiDatum  string  `json:"zadnjidatum"`
	Poruka       string  `json:"poruka"`
}

// RobInterniStavkaParams are the values of the form of a stavka of an interni prenos (the legacy
// rob_InterniRacuni_DET) by the names of the fields of the form: the document and the stavka (rproid 0 for a
// new one), the konto and the šifra of the artikal, the količina, the cena (magacinska), the rabat (VRA), the
// nabavna cena, the marža (MMA), the maloprodajna cena and the one with the porez, the tarifa, the OTK broj,
// the serija and the rok.
type RobInterniStavkaParams struct {
	RdokID int64  `form:"rdokid"`
	RproID int64  `form:"rproid"`
	Konto  string `form:"kontostavke"`
	Sifra  string `form:"sifraartikla"`
	Kolic  string `form:"kolicina"`
	Cena   string `form:"magacinskacena"`
	Vra    string `form:"rabatstavke"`
	Ncena  string `form:"nabavnacena"`
	Mma    string `form:"maloprmarza"`
	Mcena  string `form:"maloprcena"`
	Mcenap string `form:"mpcenasaporezom"`
	Tarifa string `form:"tarifa"`
	Otk    string `form:"otk"`
	Serija string `form:"serija"`
	Rok    string `form:"rokstavke"`
}

// RobInterniStavkaForma is a saved stavka of an interni prenos in the form of the stavke (the legacy
// "Izmene"): its values as the form shows them, with the artikal, its stopa, model, taksa and stanje (the
// stanje with the količina of the stavka, like the legacy "Izmene", which takes the stavka back first).
type RobInterniStavkaForma struct {
	RproID int64   `json:"rproid"`
	Sifra  int     `json:"sifra"`
	Naziv  string  `json:"naziv"`
	Konto  string  `json:"konto"`
	Jm     string  `json:"jm"`
	Tarifa int     `json:"tarifa"`
	Stopa  float64 `json:"stopa"`
	Model  string  `json:"model"`
	Taksa  float64 `json:"taksa"`
	Stanje float64 `json:"stanje"`
	Kolic  float64 `json:"kolicina"`
	Cena   float64 `json:"cena"`
	Vra    float64 `json:"vra"`
	Ncena  float64 `json:"ncena"`
	Mma    float64 `json:"mma"`
	Mcena  float64 `json:"mcena"`
	Mcenap float64 `json:"mcenap"`
	Otk    string  `json:"otk"`
	Serija string  `json:"serija"`
	Rok    string  `json:"rok"`
}

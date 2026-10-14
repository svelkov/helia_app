package domain

// ---------------------------------------------------------------------------------------------------
// Porudžbenice (the group POR, the legacy rob_Porudzbenica with its stavke rob_PORUDZBENICE_DET): a robni
// dokument (rdok) of a dobavljač (rdok.fkto, rdok.fana) whose stavke are at the nabavna cena
// ---------------------------------------------------------------------------------------------------

// RobnoPorHeaderView is the header of a porudžbenica: the fields it shares with the faktura (the dobavljač
// is the kupac of the faktura: KupacKonto, KupacSifra, KupacNaziv...) and the opis of its vrsta dokumenta.
type RobnoPorHeaderView struct {
	RobnoFaktureHeaderView
	VrstaOpis string
}

// RobnoPorStavka is the entry form of one stavka of a porudžbenica: the fields it shares with the stavka of
// the faktura (konto, šifra, JM, stanje, tarifa, količina, nabavna cena, iznos) and the dobavljač of the
// document (read only: "konto šifra naziv ;adresa ;poštanski broj mesto", the legacy EDT_KUPAC).
type RobnoPorStavka struct {
	RobnoFaktureStavka
	Dobavljac string
	// MagaciniID is the magacin of the document: the konta of the stavke are its konta (magkonto).
	MagaciniID int
}

// RobnoPorDokument is the document of the "Porudžbenice" screen: the header, the entry of a stavka with the
// data of the document it repeats, and the message of the screen (e.g. why the selected document cannot be
// changed).
type RobnoPorDokument struct {
	Header RobnoPorHeaderView
	Stavka RobnoPorStavka
	Poruka string
}

// RobnoPorParams are the values of the header of a porudžbenica (the legacy ValidacijaRDOK and SnimiRDOK of
// rob_Porudzbenica) by the names of the fields of the form, plus the nalog, the magacin and the document they
// belong to. The values are the texts of the form: the validation parses them itself, so a wrong value is a
// field error and not a binding error.
type RobnoPorParams struct {
	RnalID int64 `form:"rnalid"`
	RdokID int64 `form:"rdokid"`
	// MagaciniID is the magacin of the nalog (set by the handler from the rnalid, not sent by the form).
	MagaciniID int `form:"-"`

	Vrd        string `form:"vrd"`
	BrojDokum  string `form:"brdok"`
	DatumDokum string `form:"datumfakturisanja"`
	// The dobavljač (rdok.fkto, rdok.fana).
	DobavljacKonto string `form:"kupackonto"`
	DobavljacSifra string `form:"kupacsifra"`
	// The uslovi nabavke robe: the način dopremanja (rdok.otp), the uslovi plaćanja (rdok.pla), the rok.
	NacinDopremanja string `form:"nacindopremanja"`
	UsloviPlacanja  string `form:"usloviplacanja"`
	Rok             string `form:"rok"`
	Orgjed          string `form:"idorgjed"`
	MestoTroska     string `form:"mestotrid"`
	Komercijalista  string `form:"komid"`
	Tekst           string `form:"foot"`
}

// RobPorArtikal is the artikal of a stavka of a porudžbenica with what the form of the stavka shows of it
// (the legacy exit of EDT_SIFRA and EDT_KONTO of rob_PORUDZBENICE_DET): the artikal, its tarifa with its
// stopa on the datum of the document, its stanje in the konto of the magacin (Rsta: it has one) and its
// nabavna cena.
type RobPorArtikal struct {
	Sifra  int     `json:"sifra"`
	Naziv  string  `json:"naziv"`
	Jm     string  `json:"jm"`
	Tarifa int     `json:"tarifa"`
	Stopa  float64 `json:"stopa"`
	Rsta   bool    `json:"rsta"`
	Stanje float64 `json:"stanje"`
	Ncena  float64 `json:"ncena"`
}

// RobPorStavkaParams are the values of a stavka of a porudžbenica by the names of the fields of the form
// (RproID 0: a new stavka).
type RobPorStavkaParams struct {
	RdokID int64  `form:"rdokid"`
	RproID int64  `form:"rproid"`
	Konto  string `form:"kontostavke"`
	Sifra  string `form:"sifraartikla"`
	Kolic  string `form:"kolicina"`
	Ncena  string `form:"nabavnacena"`
	Tarifa string `form:"tarifa"`
}

// RobPorStavkaForma is a saved stavka of a porudžbenica in the form of the stavke (the legacy "Izmene"): its
// values as the form shows them, with the artikal, its stopa and its stanje.
type RobPorStavkaForma struct {
	RproID int64   `json:"rproid"`
	Sifra  int     `json:"sifra"`
	Naziv  string  `json:"naziv"`
	Konto  string  `json:"konto"`
	Jm     string  `json:"jm"`
	Kolic  float64 `json:"kolicina"`
	Ncena  float64 `json:"ncena"`
	Tarifa int     `json:"tarifa"`
	Stopa  float64 `json:"stopa"`
	Stanje float64 `json:"stanje"`
}

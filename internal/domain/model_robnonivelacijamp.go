package domain

// ---------------------------------------------------------------------------------------------------
// Nivelacije maloprodaje (the group MNI, the legacy rob_Nivelacija_maloprodaje with its stavke
// rob_Nivelacija_Maloprodaje_DET): a robni dokument (rdok) of a prodavnica (rdok.pkto, rdok.pana, also
// saved as fkto, fana) whose stavke change the maloprodajna cena of an artikal given by its naziv only
// (the šifra 0, the konto 999999): the stara and the nova MP cena with PDV and the stara and the nova
// poreska tarifa
// ---------------------------------------------------------------------------------------------------

// RobnoMniHeaderView is the header of a nivelacija maloprodaje: the fields it shares with the faktura (the
// prodavnica is the kupac of the faktura: KupacKonto, KupacSifra, KupacNaziv...) and the opis of its vrsta
// dokumenta.
type RobnoMniHeaderView struct {
	RobnoFaktureHeaderView
	VrstaOpis string
}

// RobnoMniStavka is the entry form of one stavka of a nivelacija maloprodaje: the document it belongs to, the
// data of the document it repeats (read only, the legacy strip "Podaci o dokumentu" of the stavke) and the
// state of the form kept for the next stavka after a save (the legacy keeps its checkboxes and, with "Unesi
// procenat", the procenat).
type RobnoMniStavka struct {
	RdokID                    int64
	Vrd, Brdok, DatumDokument string
	UnesiProcenat             bool
	Procenat                  string
	PromeniTarifu             bool
}

// RobnoMniDokument is the document of the "Nivelacije maloprodaje" screen: the header, the entry of a stavka
// and the message of the screen (e.g. why the selected document cannot be changed).
type RobnoMniDokument struct {
	Header RobnoMniHeaderView
	Stavka RobnoMniStavka
	Poruka string
}

// RobnoMniParams are the values of the header of a nivelacija maloprodaje (the legacy ValidacijaRDOK and
// SnimiRDOK of rob_Nivelacija_maloprodaje) by the names of the fields of the form, plus the nalog, the
// magacin and the document they belong to. The values are the texts of the form: the validation parses
// them itself, so a wrong value is a field error and not a binding error.
type RobnoMniParams struct {
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
	// The prodavnica: the konto and the analitička šifra of the kontni plan (rdok.pkto, rdok.pana).
	ProdavnicaKonto string `form:"kupackonto"`
	ProdavnicaSifra string `form:"kupacsifra"`
	Orgjed          string `form:"idorgjed"`
	MestoTroska     string `form:"mestotrid"`
}

// RobMniStavkaParams are the values of a stavka of a nivelacija maloprodaje by the names of the fields of the
// form (RproID 0: a new stavka): the naziv and the JM of the artikal, the količina, the stara and the nova MP
// cena with PDV, the stara poreska tarifa and, with PromeniTarifu, the nova one (else the stara).
type RobMniStavkaParams struct {
	RdokID        int64  `form:"rdokid"`
	RproID        int64  `form:"rproid"`
	Naziv         string `form:"nazivmni"`
	Jm            string `form:"jm"`
	Kolic         string `form:"kolicina"`
	StaraCena     string `form:"staracena"`
	NovaCena      string `form:"novacena"`
	StaraTarifa   string `form:"staratarifa"`
	PromeniTarifu bool   `form:"promenitarifu"`
	NovaTarifa    string `form:"novatarifa"`
	// UnesiProcenat and Procenat are only kept for the next stavka (the nova cena already has the procenat).
	UnesiProcenat bool   `form:"unesiprocenat"`
	Procenat      string `form:"procenat"`
}

// RobMniStavkaForma is a saved stavka of a nivelacija maloprodaje in the form of the stavke ("Izmeni" of its
// row): its values as the form shows them, with the stope of the tarife on the datum of the document.
type RobMniStavkaForma struct {
	RproID      int64   `json:"rproid"`
	Naziv       string  `json:"naziv"`
	Jm          string  `json:"jm"`
	Kolicina    float64 `json:"kolicina"`
	StaraCena   float64 `json:"staracena"`
	NovaCena    float64 `json:"novacena"`
	StaraTarifa int     `json:"staratarifa"`
	StaraStopa  float64 `json:"starastopa"`
	NovaTarifa  int     `json:"novatarifa"`
	NovaStopa   float64 `json:"novastopa"`
}

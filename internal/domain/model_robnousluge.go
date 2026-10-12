package domain

// ---------------------------------------------------------------------------------------------------
// Fakture usluga (the groups PRE, FUR, FCO, RCO and KPF, the legacy rob_Fakture_Usluga): the header is the
// one of the faktura (RobnoFaktureHeaderView with Usluge), the stavke are the legacy rob_FAKTURE_Usluga_DET2
// ---------------------------------------------------------------------------------------------------

// RobnoUslugeStavka is the entry of a stavka of a faktura usluga: the one of the faktura (the data of the
// document it repeats, the konto, the poreske kategorije, the valuta) with what the stavke of the usluge
// show: the group of the vrsta (Grupa: a KPF has the šifra 0 only), the tip zaliha of the magacin (Tipzal 2:
// the OTK broj, the serija and the rok), the magacin of the veleprodajna (VpCena: no nabavna cena) or of the
// nabavna cena (NabCena: no magacinska cena).
type RobnoUslugeStavka struct {
	RobnoFaktureStavka
	Grupa           string
	Tipzal          int
	VpCena, NabCena bool
	Naslov          string
	OpisUsluge      string
	OpisDet         string
	// Avansni: the stavke of an avansni račun (the legacy rob_AvansniRacuni_DET: the šifra 0 with the opis,
	// the količina 1, the tarifa with its poreska kategorija, the cena and the iznos).
	Avansni bool
}

// RobnoUslugeDokument is the document of the "Fakture usluga" screen: the header, the entry of a stavka,
// the avansi, the žiro računi, the message of the screen and its title (the legacy STC_Static1: "Fakture
// tekst", "Knjižno odobrenje", "Knjižno zaduženje", "Zaduženje CO", "Razduženje CO").
type RobnoUslugeDokument struct {
	Header     RobnoFaktureHeaderView
	Stavka     RobnoUslugeStavka
	Avansi     []TableRow
	ZiroRacuni []ComboItem
	Poruka     string
	Naslov     string
}

// RobUslugaArtikal is the artikal of a stavka of a faktura usluga with what the form of the stavke shows of
// it (the legacy exit of EDT_SIFRA of rob_FAKTURE_Usluga_DET2): its naziv, JM, tarifa with its stopa, tip
// (U: a usluga, no zalihe), the konto of the stavka (the konto of the artikal, else the one of the magacin),
// the poreska kategorija of its tarifa, its stanje and cene in the konto (the nabavna cena prosnc, the cena
// vpcena or the cena of rcene, the magacinska cena), and, for the vrsta 280, the last cena to the kupac.
type RobUslugaArtikal struct {
	Sifra   int     `json:"sifra"`
	Naziv   string  `json:"naziv"`
	Jm      string  `json:"jm"`
	Tarifa  int     `json:"tarifa"`
	Stopa   float64 `json:"stopa"`
	Tip     string  `json:"tip"`
	Konto   string  `json:"konto"`
	Taxcat  string  `json:"taxcat"`
	Stanje  float64 `json:"stanje"`
	Ncena   float64 `json:"ncena"`
	Cena    float64 `json:"cena"`
	MagCena float64 `json:"magcena"`
}

// RobUslugaStopa is the poreska stopa of a tarifa of a stavka of a faktura usluga (the legacy exit of
// EDT_po) with the poreska kategorija of the tarifa (1: S20, 2: S10).
type RobUslugaStopa struct {
	Stopa  float64 `json:"stopa"`
	Taxcat string  `json:"taxcat"`
}

// RobUslugaStavkaParams are the values of the form of a stavka of a faktura usluga by the names of the
// fields of the form: the document and the stavka (rproid 0 for a new one), the konto, the šifra, the
// opis (naziv) and the JM of the usluga, the količina, the cena (and the cena as typed in the valuta), the
// rabat, the tarifa, the poreska kategorija with its osnov oslobođenja, the nabavna cena, the OTK broj,
// the serija, the rok and the dodatni opis usluge of the document.
type RobUslugaStavkaParams struct {
	RdokID    int64  `form:"rdokid"`
	RproID    int64  `form:"rproid"`
	Konto     string `form:"kontostavke"`
	Sifra     string `form:"sifraartikla"`
	Naziv     string `form:"opisusluge"`
	Jm        string `form:"jmusluge"`
	Kolic     string `form:"kolicina"`
	Cena      string `form:"cenausluge"`
	CenaVal   string `form:"cenaval"`
	Rab       string `form:"rabatusluge"`
	Tarifa    string `form:"tarifa"`
	Taxcat    string `form:"taxcat"`
	Taxexreco string `form:"taxexreco"`
	Ncena     string `form:"nabavnacena"`
	Otk       string `form:"otk"`
	Serija    string `form:"serija"`
	Rok       string `form:"rokstavke"`
	OpisDet   string `form:"opisdet"`
}

// RobUslugaStavkaForma is a saved stavka of a faktura usluga in the form of the stavke (the legacy
// "Izmene" of rob_FAKTURE_Usluga_DET2): its values as the form shows them, with the dodatni opis usluge of
// the document (dokusldet).
type RobUslugaStavkaForma struct {
	RproID    int64   `json:"rproid"`
	Sifra     int     `json:"sifra"`
	Naziv     string  `json:"naziv"`
	Konto     string  `json:"konto"`
	Jm        string  `json:"jm"`
	Tarifa    int     `json:"tarifa"`
	Stopa     float64 `json:"stopa"`
	Kolicina  float64 `json:"kolicina"`
	Cena      float64 `json:"cena"`
	Rabat     float64 `json:"rabat"`
	Ncena     float64 `json:"ncena"`
	Taxcat    string  `json:"taxcat"`
	Taxexreco string  `json:"taxexreco"`
	Otk       string  `json:"otk"`
	Serija    string  `json:"serija"`
	Rok       string  `json:"rok"`
	OpisDet   string  `json:"opisdet"`
}

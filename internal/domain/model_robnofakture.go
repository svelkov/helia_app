package domain

// The view structs of the "Fakture veleprodaje" screen (RobnoFakture / RobnoFaktureDialog in
// frontend/templates/robno/robnadokumenta.templ). They live here because both the templates and the
// handler of the screen need them: the handler fills them and the templates lay them out, and the
// template package cannot import the handler package (the handler imports the templates).

// RobnoFaktureHeader holds the values of the header of the faktura (the first window). The values are
// the ones the handler reads from the form (all strings, the handler formats the numbers), so the
// template only lays them out.
type RobnoFaktureHeader struct {
	// Snimljen tells whether the header is saved: while it is false the header is editable and the
	// stavke control is disabled, afterwards the other way round.
	Snimljen bool

	// Nalog (the read only strip at the top of both windows).
	VrNal, Nalog, DatumNaloga, UkDokum, UkStavki, Duguje, Potrazuje, IznosDokum, Magacin string

	// Otpremnica / Račun and "Štampaj detalje u stavkama".
	Otpremnica     bool
	StampajDetalje bool

	// Podaci o dokumentu.
	Vrd, Brdok, DatumFakturisanja, BrdokOrg, DatumIzvora, Opis string
	ZaPeriod                                                   bool
	PeriodOd, PeriodDo, DatumSmanjenja                         string

	// Podaci o kupcu.
	KupacKonto, KupacSifra, KupacNaziv, KupacAdresa string
	KupacPostanskiBroj, KupacMesto, KupacPib        string
	SistemPdv                                       string

	// Podaci o načinu otpreme, rokovima, valuti i avansima.
	Valuta, Kurs, VaziZa, BrojRata, DatumPrveRate, IznosFakture string
	UsloviPlacanja, Rok, Dospeva                                string
	BrojUgovora, Narudzbenica, IzvoRacun, NacinOtpreme          string
	RegBr, Vozac, Teret, BrutoTezina                            string
	BrojKutija, NetoTezina, BrojPaleta, IzvozIzjava             string
	BankaUplatu, Popdv, IzvozniDokument                         string

	// Podaci o porezom oslobođenju, nastanku PDV obaveze.
	PbrKategorija, NapomenaPoresko, RbrPdv     string
	MestoTroska, MestoIsporuke, Komercijalista string
	PostanskiBroj, IfBroj                      string
	JavnaNabavka                               bool
	BjkosNosilac, BrojTendera, Kst             string

	// Finansijsko stanje kupca (realizovano / planirano / reprogramirano and the saldo columns).
	Realizovana, Planirana, Reproknjizena, Ukupno string
	Saldo, Dospeo, ReproDospeo, SaldoDospeo       string
}

// RobnoFaktureStavka holds the values of the entry form of one stavka (the second window). The first
// group of fields is read only: it repeats the kupac and the podaci o dokumentu of the header.
type RobnoFaktureStavka struct {
	KontoMagacina, KontoKupca, Vrd, Brdok, DatumDokumenta string

	SifraArtikla, SifraArtiklaNaziv  string
	JM, Stanje                       string
	Tarifa, Dani                     string
	Kolicina, Iznos                  string
	ProdajnaCena, MagacinskaCena     string
	Rabat, IznosRabata               string
	NabavnaCena, NetoProdajnaCena    string
	Deklaracije, BarKod, Serija, Rok string

	// "Zadnja cena po kojoj je kupac kupovao ovaj artikal".
	ZadnjaCena, ZadnjiRabat, ZadnjiDatum string
	KoristiPoslednjiRabat                bool
}

// RobnoFaktureButtons groups the buttons of the screen so that the signature of the template stays
// short: the header and the stavke have their own save/delete buttons.
type RobnoFaktureButtons struct {
	Save   Button // "Sačuvaj" the header of the faktura
	Modify Button // "Izmeni" the header (unlocks the header, locks the stavke)
	New    Button // "Novi dok."
	Delete Button // "Briši" the document
	Back   Button // "Nazad"

	SaveStavka   Button
	ModifyStavka Button
	DeleteStavka Button
}

// RobnoFaktureStripField is one cell of the two line read only strip of the "Fakture veleprodaje"
// screen (the strip of the nalog above the header form, the same one both legacy windows show): the
// label is the first line of the strip (the title of the column) and the value the second one, like
// the two row grid header of the legacy window.
type RobnoFaktureStripField struct {
	ID    string
	Label string
	Value string

	// WidthClass is the Tailwind width of the column (e.g. "w-16"). The field without it takes the
	// rest of the strip (it must be the last one, e.g. the magacin).
	WidthClass string

	// AlignClass is "text-right" for the numeric fields (the amounts) and empty (left) for the others.
	AlignClass string
}

// RobnoFaktureUI carries the urls and the element ids of the "Fakture veleprodaje" screen. They are
// declared in the handler of the screen (which cannot be imported by the templates) and passed to the
// templates through this struct, so every url and id is written in one place only.
type RobnoFaktureUI struct {
	// HeaderPanelID and StavkePanelID are the two collapsible controls (also used by the script that
	// switches the editable control).
	HeaderPanelID string
	StavkePanelID string

	// ContentID is the container the tab navigation swaps: the target of the buttons that reload the
	// whole screen.
	ContentID string

	// DialogID is the id of the dialog of the screen and DialogStagingID the element of the "Unos
	// dokumenta" tab the dialog is rendered into (see RobnoDokumentaMain).
	DialogID        string
	DialogStagingID string

	// ArtikalSearchURL and PartnerSearchURL are the search popups of the stavka and of the kupac,
	// DokumentSearchURL the one of the broj dokumenta (all of them are opened with the "..." buttons of
	// the form).
	ArtikalSearchURL  string
	PartnerSearchURL  string
	DokumentSearchURL string
}

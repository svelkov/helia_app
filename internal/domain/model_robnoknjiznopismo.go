package domain

import "database/sql"

// ---------------------------------------------------------------------------------------------------
// Knjižna pisma (the groups KNO and KNZ, the legacy rob_KnjiznoPismoFIN with its stavke
// rob_KnjiznoPismoFINDET and the window rob_KnjPismoPR_Obracun): a robni dokument (rdok) of a kupac that
// corrects a faktura (the vezni dokument, rdok.vrdokid: its stavke are copied and corrected) or, without
// it, gives only the osnovica and the porez (a knjižno pismo finansijsko)
// ---------------------------------------------------------------------------------------------------

// RobKnpDokumentDto are the columns of a knjižno pismo (or of the faktura it corrects) the screen reads besides
// RobniDokumentRow: the ugovoreni rabat and the kasa (their % and vrednosti), the iznos and the porez, the
// vezni dokument and the tip of the knjižno pismo with its text.
type RobKnpDokumentDto struct {
	Ugrabat   float64       `db:"ugrabat"`
	Vrugrabat float64       `db:"vrugrabat"`
	Pkase     float64       `db:"pkase"`
	Vkase     float64       `db:"vkase"`
	Iznos     float64       `db:"iznos"`
	Vporez    float64       `db:"vporez"`
	VrdokID   int64         `db:"vrdokid"`
	TipknjID  sql.NullInt64 `db:"tipknjid"`
	TipText   string        `db:"tiptext"`
}

// RobKnpPreracunDto is what the legacy Preracun of a knjižno pismo reads in its transaction: the sums of its
// stavke (količina by the fakturna cena, by the cena and the rabat of the fakturna vrednost), their number, and
// the ugovoreni rabat, the kasa and the iznos of the document.
type RobKnpPreracunDto struct {
	Fakturna float64
	Cena     float64
	Rabat    float64
	Stavke   int
	Ugrabat  float64
	Pkase    float64
	Iznos    float64
}

// RobnoKnpVezni is the faktura a knjižno pismo corrects (the legacy BTN_TraziOTP): its broj and datum, its kupac,
// its org. jedinica and mesto troška, its ugovoreni rabat and kasa and its valuta (JSON for the screen; RdokID 0:
// no vezni dokument).
type RobnoKnpVezni struct {
	RdokID             int64   `json:"rdokid"`
	Dokum              int     `json:"dokum"`
	Datum              string  `json:"datum"`
	KupacKonto         string  `json:"kupackonto"`
	KupacSifra         string  `json:"kupacsifra"`
	KupacNaziv         string  `json:"kupacnaziv"`
	KupacAdresa        string  `json:"kupacadresa"`
	KupacPostanskiBroj string  `json:"kupacpostanskibroj"`
	KupacMesto         string  `json:"kupacmesto"`
	KupacPib           string  `json:"kupacpib"`
	SistemPdv          string  `json:"sistempdv"`
	IDOrgjed           int64   `json:"idorgjed"`
	OrgjedText         string  `json:"orgjedtext"`
	MestoTrID          int64   `json:"mestotrid"`
	MestoTrText        string  `json:"mestotrtext"`
	Ugrabat            float64 `json:"ugrabat"`
	Vrugrabat          float64 `json:"vrugrabat"`
	Pkase              float64 `json:"pkase"`
	Vkase              float64 `json:"vkase"`
	IDValute           int64   `json:"idvalute"`
	ValutaText         string  `json:"valutatext"`
	Kurs               float64 `json:"kurs"`
	Vaziza             int     `json:"vaziza"`
}

// RobnoKnpHeaderView is the header of a knjižno pismo: the fields it shares with the faktura, the opis of its
// vrsta dokumenta, its tip (tipknpisma), the vezni dokument (the faktura it corrects), the new ugovoreni rabat and
// kasa of a knjižno pismo with a vezni dokument and the osnovica and the porez of one without it.
type RobnoKnpHeaderView struct {
	RobnoFaktureHeaderView
	VrstaOpis  string
	TipKnjID   sql.NullInt64
	TipKnjText string
	Vezni      RobnoKnpVezni
	// The new ugovoreni rabat and kasa (rdok.ugrabat, rdok.pkase) with their vrednosti (rdok.vrugrabat,
	// rdok.vkase, the legacy Preracun).
	Ugrabat, Vrugrabat, Pkase, Vkase float64
	// A knjižno pismo without a vezni dokument: the osnovica (rdok.iznos), the % poreza (rdok.pkase), the vrednost
	// poreza (rdok.vporez) and the iznos dokumenta (the osnovica with the porez).
	Osnovica, ProcPoreza, VrednostPoreza, IznosDokumenta float64
}

// Vezan tells whether the knjižno pismo corrects a faktura (it has a vezni dokument).
func (h RobnoKnpHeaderView) Vezan() bool {
	return h.Vezni.RdokID > 0
}

// RobnoKnpStavke is the stavke control of a knjižno pismo: the document they belong to and the data of the
// document they repeat (read only: the kupac "konto šifra naziv ;adresa ;poštanski broj mesto", the vrsta, the
// broj, the datum and the iznos of the document).
type RobnoKnpStavke struct {
	RdokID                                            int64
	Kupac, Vrd, Brdok, DatumDokumenta, IznosDokumenta string
}

// RobnoKnpDokument is the document of the "Knjižna pisma" screen: the header, the data of the stavke and the
// message of the screen (e.g. why the selected document cannot be changed).
type RobnoKnpDokument struct {
	Header RobnoKnpHeaderView
	Stavke RobnoKnpStavke
	Poruka string
}

// RobnoKnpParams are the values of the header of a knjižno pismo (the legacy ValidacijaRDOK and SnimiRDOK of
// rob_KnjiznoPismoFIN): the fields it shares with the faktura, by the same names of the form, the tip of the
// knjižno pismo, the vezni dokument, the new ugovoreni rabat and kasa (with a vezni dokument) and the osnovica
// and the porez (without it).
type RobnoKnpParams struct {
	RobnoFaktureParams
	TipKnjID       string `form:"tipknjid"`
	VezniRdokID    string `form:"vrdokid"`
	Ugrabat        string `form:"ugrabat"`
	Pkase          string `form:"pkase"`
	Osnovica       string `form:"osnovica"`
	ProcPoreza     string `form:"procporeza"`
	VrednostPoreza string `form:"vrednostporeza"`
}

// RobKnpStavkaParams are the corrected values of a stavka of a knjižno pismo (the legacy row of
// TABLE_rpro1): the količina, the fakturna cena and the rabat.
type RobKnpStavkaParams struct {
	RdokID   int64  `form:"rdokid"`
	RproID   int64  `form:"rproid"`
	Kolicina string `form:"kolicina"`
	Cena     string `form:"prodajnacena"`
	Rabat    string `form:"rabatstavke"`
}

// RobKnpStavkaForma is a stavka of a knjižno pismo for its form (JSON): the artikal and the values that can be
// corrected.
type RobKnpStavkaForma struct {
	RproID   int64   `json:"rproid"`
	Sifra    int     `json:"sifra"`
	Naziv    string  `json:"naziv"`
	Kolicina float64 `json:"kolicina"`
	Cena     float64 `json:"cena"`
	Rabat    float64 `json:"rabat"`
}

// RobnoKnpButtons are the buttons of the "Knjižna pisma" screen: the ones of the faktura (the header, the stavka
// form), "Pregl. Vr. Fakt." of the header and of the stavke (the window of the obračun), "Dodaj sve" (all the
// stavke of the faktura) and "Vrati sve" (all the corrected stavke).
type RobnoKnpButtons struct {
	RobnoFaktureButtons
	Pregled, PregledStavke, DodajSve, VratiSve Button
}

// RobnoKnpObracunRed is one row of the window "Pregled vrednosti fakture po starim i novim uslovima prodaje"
// (the legacy rob_KnjPismoPR_Obracun): the value by the new conditions (the knjižno pismo), by the conditions of
// the faktura and their difference, each with its text (the % of the ugovoreni rabat and of the kasa differ).
type RobnoKnpObracunRed struct {
	OpisNovo, OpisFaktura, OpisRazlika string
	Novo, Faktura, Razlika             float64
}

package domain

import "database/sql"

// ---------------------------------------------------------------------------------------------------
// Opšti dokumenti (the groups OPD, POP and POT, the legacy rob_opsti_dokumenti_opd): a robni dokument with
// the protivstav konto and the org. jedinica and mesto troška of the document and of the robe; the header
// and the stavke on one screen
// ---------------------------------------------------------------------------------------------------

// RobnoOpstiHeaderView is the header of an opšti dokument: the header of the faktura (the document, its
// datumi, opis, org. jedinica and mesto troška, the strip of the nalog) with the protivstav konto and its
// analitička šifra (tkonto, tsifra) with its naziv, the org. jedinica and the mesto troška where the roba
// goes (idorgjed1, mestotrid1) and the opis of the vrsta.
type RobnoOpstiHeaderView struct {
	RobnoFaktureHeaderView
	OpisVrste                   string
	Tkonto, Tsifra, TkontoNaziv string
	Orgjed1                     sql.NullInt64
	Orgjed1Text                 string
	MestoTroska1                sql.NullInt64
	MestoTroska1Text            string
	// Fin: a finansijski opšti dokument (the group FIN, the legacy rob_opsti_dokumenti_FIN): the header has the
	// org. jedinica and the mesto troška of the document only, the stavke have the iznos only.
	Fin bool
	// Kol: a količinski opšti dokument (the group KOL, the legacy rob_opsti_dokumenti_KOL): the header of the
	// FIN, the stavke have the količina only (with the lot in a magacin of the tip zaliha 2).
	Kol bool
	// Mnv: a nivelacija (the group MNV, the legacy rob_opsti_dokumenti_MNV): the header without the org.
	// jedinica and the mesto troška, the stavke by the naziv (no artikal) with the stara and the nova cena.
	Mnv bool
	// Niv: a nivelacija veleprodaje (the group NIV, the legacy rob_Nivelacije_Veleprodaja): the header with the
	// org. jedinica and the mesto troška of the document, the stavke with the stara and the nova veleprodajna
	// cena of the artikal.
	Niv bool
}

// BezProtivstava tells whether the header has no protivstav konto and no org. jedinica and mesto troška of
// the roba: a FIN, a KOL, an MNV or a NIV.
func (h RobnoOpstiHeaderView) BezProtivstava() bool {
	return h.Fin || h.Kol || h.Mnv || h.Niv
}

// RobnoOpstiStavka is the entry of a stavka of an opšti dokument: the document, the konto (by default the
// first konto of the magacin), the tip zaliha of the magacin (Tipzal 2: the OTK broj, the serija and the rok),
// the label of the cena by the kind of the magacin (CenaNaziv: "Veleprodajna cena" or "Magacinska cena") and
// whether the cena is not entered (CenaZakljucana: a magacin of the nabavna cena, not of a nalog of the
// servis: the cena is the nabavna cena).
type RobnoOpstiStavka struct {
	RdokID         int64
	Konto          string
	Tipzal         int
	NabCena        bool
	CenaZakljucana bool
	CenaNaziv      string
	// Fin: the stavke of a finansijski opšti dokument (the konto, the artikal and the iznos).
	Fin bool
	// Kol: the stavke of a količinski opšti dokument (the konto, the artikal, the količina and the lot).
	Kol bool
	// Mnv: the stavke of a nivelacija (the naziv, the količina, the stara and the nova cena).
	Mnv bool
	// Niv: the stavke of a nivelacija veleprodaje (the artikal, the konto, the količina, the procenat, the stara
	// and the nova cena; the lot in a magacin of the tip zaliha 2).
	Niv bool
}

// RobnoOpstiDokument is the document of the "Opšti dokumenti" screen: the header, the entry of a stavka and
// the message of the screen (e.g. why the selected document cannot be changed).
type RobnoOpstiDokument struct {
	Header RobnoOpstiHeaderView
	Stavka RobnoOpstiStavka
	Poruka string
}

// RobnoOpstiParams are the fields of the header of an opšti dokument (the legacy rob_opsti_dokumenti_opd) by
// the names of the fields of the form; MagaciniID is the magacin of the nalog (set by the handler).
type RobnoOpstiParams struct {
	RnalID          int64  `form:"rnalid"`
	RdokID          int64  `form:"rdokid"`
	MagaciniID      int    `form:"-"`
	Vrd             string `form:"vrd"`
	BrojDokum       string `form:"brdok"`
	DatumDokum      string `form:"datumfakturisanja"`
	IzvorniDokum    string `form:"brdokorg"`
	IzvorniDokDatum string `form:"datumizvora"`
	Opis            string `form:"opis"`
	Tkonto          string `form:"tkonto"`
	Tsifra          string `form:"tsifra"`
	Orgjed          string `form:"idorgjed"`
	MestoTroska     string `form:"mestotrid"`
	Orgjed1         string `form:"idorgjed1"`
	MestoTroska1    string `form:"mestotrid1"`
}

// RobOpstiArtikal is the artikal of a stavka of an opšti dokument with what the form of the stavke shows of
// it (the legacy exit of EDT_SIFRA of rob_opsti_dokumenti_opd): its šifra, naziv, JM and konto, its cena, its
// nabavna cena and its stanje; Poruka is the message of an artikal that cannot be entered.
type RobOpstiArtikal struct {
	Sifra  int     `json:"sifra"`
	Naziv  string  `json:"naziv"`
	Jm     string  `json:"jm"`
	Konto  string  `json:"konto"`
	Cena   float64 `json:"cena"`
	Ncena  float64 `json:"ncena"`
	Stanje float64 `json:"stanje"`
	Poruka string  `json:"poruka"`
	// Kolicina and StaraCena are the količina and the stara (veleprodajna) cena of the stanje of the artikal of a
	// nivelacija veleprodaje (the group NIV).
	Kolicina  float64 `json:"kolicina"`
	StaraCena float64 `json:"staracena"`
}

// RobOpstiStavkaParams are the values of the form of a stavka of an opšti dokument by the names of the fields
// of the form: the document, the konto, the šifra, the količina, the cena, the nabavna cena, the OTK broj,
// the serija and the rok.
type RobOpstiStavkaParams struct {
	RdokID int64  `form:"rdokid"`
	Konto  string `form:"kontostavke"`
	Sifra  string `form:"sifraartikla"`
	Kolic  string `form:"kolicina"`
	Cena   string `form:"cenaopd"`
	Ncena  string `form:"nabavnacena"`
	Otk    string `form:"otk"`
	Serija string `form:"serija"`
	Rok    string `form:"rokstavke"`
	// Iznos is the iznos of a stavka of a finansijski opšti dokument (the group FIN).
	Iznos string `form:"iznosfin"`
	// Naziv, StaraCena and NovaCena are the naziv of the artikal and the cene of a stavka of a nivelacija (the
	// group MNV).
	Naziv     string `form:"nazivmnv"`
	StaraCena string `form:"staracena"`
	NovaCena  string `form:"novacena"`
	// Procenat is the procenat of the nivelacija veleprodaje (the group NIV; with StaraCena and NovaCena).
	Procenat string `form:"procenatniv"`
}

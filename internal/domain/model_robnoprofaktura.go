package domain

// ---------------------------------------------------------------------------------------------------
// Profakture (the group PRO, the legacy rob_ProFakture with its stavke rob_PROFAKTURE_DET): a robni
// dokument (rdok) of a kupac like the faktura veleprodaje, without the avansi, the rate and the obračun
// ---------------------------------------------------------------------------------------------------

// RobnoProHeaderView is the header of a profaktura: the fields it shares with the faktura, the opis of its
// vrsta dokumenta, the broj fiskalnog isečka (rdok.bi) and the domaća valuta of the firm (the paritet, the
// izvozna izjava and the banka za uplatu - izvoz are shown only for another valuta, the legacy GR_IZVOZ).
type RobnoProHeaderView struct {
	RobnoFaktureHeaderView
	VrstaOpis    string
	BrojIsecka   string
	DomacaValuta int
}

// RobnoProStavka is the entry form of one stavka of a profaktura: the fields it shares with the stavka of
// the faktura and the kupac of the document (read only: "konto šifra naziv ;adresa ;poštanski broj mesto",
// the legacy EDT_KUPAC).
type RobnoProStavka struct {
	RobnoFaktureStavka
	Kupac string
}

// RobnoProDokument is the document of the "Profakture" screen: the header, the entry of a stavka with the
// data of the document it repeats, and the message of the screen (e.g. why the selected document cannot be
// changed).
type RobnoProDokument struct {
	Header RobnoProHeaderView
	Stavka RobnoProStavka
	Poruka string
}

// RobnoProParams are the values of the header of a profaktura (the legacy ValidacijaRDOK and SnimiRDOK of
// rob_ProFakture): the fields it shares with the faktura, by the same names of the form, and the broj
// fiskalnog isečka.
type RobnoProParams struct {
	RobnoFaktureParams
	BrojIsecka string `form:"brojisecka"`
}

// RobProStavkaParams are the values of a stavka of a profaktura (the legacy Validacija, CreateRpro and
// EditRpro of rob_PROFAKTURE_DET): the fields it shares with the stavka of the faktura, and the naziv and the
// JM of the stavka (rpro.naz1, rpro.jm: typed for the šifra 0, else the ones of the artikal).
type RobProStavkaParams struct {
	RobStavkaParams
	Naziv string `form:"nazivstavke"`
	Jm    string `form:"jm"`
}

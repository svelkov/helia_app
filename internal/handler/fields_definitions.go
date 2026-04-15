package handler

import "helia/internal/domain"

func SetBankeFields() []domain.Fields {
	return []domain.Fields{
		{Name: "brrac", Label: "Broj Racuna", Width: "45", Sortable: true},
		{Name: "banka", Label: "Banka", Width: "60", Sortable: true},
		{Name: "konto", Label: "Konto", Width: "8", Sortable: true},
		{Name: "sifra", Label: "Sifra", Width: "8", Sortable: true},
		{Name: "bnkcod", Label: "Kod Banke", Width: "4", Sortable: true},
		{Name: "ebank", Label: "E-Banking", Width: "60", Sortable: true},
		{Name: "pocnazfajl", Label: "Poc Naz Fajla", Width: "10", Sortable: true},
		{Name: "tipdok", Label: "Tip Dokumenta", Width: "6", Sortable: true},
		{Name: "nafakne", Label: "Na Fakt Ne", Width: "1", Sortable: true},
	}
}

func SetBnkizvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifbank", Label: "Sifra Banke", Width: "6", Sortable: true},
		{Name: "bnkdes", Label: "Naziv Banke", Width: "80", Sortable: true},
		{Name: "swiftadr", Label: "Swift Adresa", Width: "80", Sortable: true},
		{Name: "brojrac", Label: "Broj Racuna", Width: "55", Sortable: true},
		{Name: "beneficiary", Label: "Beneficiary", Width: "70", Sortable: true},
		{Name: "corrbank", Label: "Korespodentna banka", Width: "90", Sortable: true},
		{Name: "tel", Label: "Telefon", Width: "25", Sortable: true},
		{Name: "fax", Label: "Fax", Width: "30", Sortable: true},
		{Name: "address", Label: "Adresa", Width: "40", Sortable: true},
		{Name: "komentar", Label: "Komentar", Width: "70", Sortable: true},
	}
}

func SetDokvrstaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vrd", Label: "Vrsta dokumenta", Type: "text", Width: "5", Sortable: true},
		{Name: "opis", Label: "Opis vrste dokumenta", Type: "text", Width: "40", Sortable: true},
		{Name: "kodknj", Label: "Nacin knjizenja (d,p,dp)", Type: "text", Width: "4", Sortable: true},
		{Name: "predznak", Label: "Predznak", Type: "text", Width: "3", Sortable: true},
		{Name: "grpdok", Label: "Grupa dokumenta", Type: "text", Width: "10", Sortable: true},
		{Name: "dokozn", Label: "Oznaka dokumenta", Type: "text", Width: "5", Sortable: true},
		{Name: "stornovrd", Label: "Storno vrsta dokumenta", Type: "text", Width: "6", Sortable: true},
		{Name: "modul", Label: "Modul", Type: "text", Width: "8", Sortable: true},
	}
}

func SetDrzaveFields() []domain.Fields {
	return []domain.Fields{
		{Name: "naziv", Label: "Naziv drzave", Width: "40", Sortable: true},
		{Name: "ozndrz", Label: "Oznaka drzave", Width: "6", Sortable: true},
	}
}
func SetFevpdvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vktip", Label: "Tip I/U", Width: "4", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "100", Sortable: true},
		{Name: "vkrbr", Label: "Vrsta r br", Width: "3", Sortable: true},
		{Name: "obrazac", Label: "Obrazac", Width: "6", Sortable: true},
	}
}

func SetMestotroskaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mtroska", Label: "Mesto troska", Width: "6", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "45", Sortable: true},
		{Name: "idorgjed", Label: "Org. jedinica", Width: "20", Sortable: true},
	}
}

func SetOrgjedFields() []domain.Fields {
	return []domain.Fields{
		{Name: "ojozn", Label: "Sifra Orgjed", Width: "6", Sortable: true},
		{Name: "naziv", Label: "Naziv Orgjed", Width: "45", Sortable: true},
	}
}

// key of the map must be the name of filed in the table in db (we need it for mapping)

func SetPopdvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "popdv", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10", Sortable: true},
		{Name: "opis", Label: "Opis", ValidationText: "Morate uneti opis dokumenta...", Width: "60", Sortable: true},
		{Name: "grpdok", Label: "Grupa dok", ValidationText: "Morate uneti grupu dokumenata...", Width: "20", Sortable: true},
		{Name: "grpvrd", Label: "Grp. Vrste Dok.", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20", Sortable: true},
		{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10", Sortable: true},
	}
}

func SetSifmestoFileds() []domain.Fields {
	return []domain.Fields{
		{Name: "sifm", Label: "Sifra Mesta", Width: "4", Sortable: true},
		{Name: "naziv", Label: "Naziv", Width: "80", Sortable: true},
		{Name: "ops", Label: "Opstina", Width: "8", Sortable: true},
		{Name: "pobro", Label: "Postanski broj", Width: "8", Sortable: true},
		{Name: "km", Label: "Km", Width: "6", Sortable: true},
	}
}

func SetSifopFields() []domain.Fields {
	return []domain.Fields{
		{Name: "ops", Label: "Sifra Opstine", Width: "6", Sortable: true},
		{Name: "naziv", Label: "Naziv Opstine", Width: "60", Sortable: true},
	}
}
func SetSifplizvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifplac", Label: "Sifra Placanja", ValidationText: "morate uneti sifru placanja...", Width: "6", Sortable: true},
		{Name: "oblik", Label: "Oblik", ValidationText: "morate uneti oblik placanja...", Width: "6", Sortable: true},
		{Name: "osnov", Label: "Osnov placanja", ValidationText: "morate uneti osnov placanja", Width: "4", Sortable: true},
		{Name: "opis", Label: "Opis", ValidationText: "morate uneti opis", Width: "80", Sortable: true},
		{Name: "konto", Label: "Konto", Width: "8", Sortable: true},
		{Name: "sifra", Label: "Sifra", Width: "8", Sortable: true},
	}
}
func SetTipdokFields() []domain.Fields {
	return []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10", Sortable: true},
		{Name: "opis", Label: "Opis", ValidationText: "Morate uneti opis dokumenta...", Width: "60", Sortable: true},
		{Name: "grpdok", Label: "Grupa dok", ValidationText: "Morate uneti grupu dokumenata...", Width: "20", Sortable: true},
		{Name: "grpvrd", Label: "Grp Vrste Dok", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20", Sortable: true},
		{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10", Sortable: true},
	}
}

func SetFvknjracFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vktip", Label: "Tip knjige (ulazni/izlazni racun)", Width: "2", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "50", Sortable: true},
		{Name: "vkrbr", Label: "Vrsta knjige racuna", Width: "2", Sortable: true},
		{Name: "konta", Label: "Konta", Width: "60", Sortable: true},
	}
}
func SetPartneriFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifra", Label: "Sifra", ValidationText: "morate uneti sifru partnera...", Width: "8"},
		{Name: "naziv", Label: "Naziv", ValidationText: "morate uneti naziv partnera..", Width: "60"},
		{Name: "adresa", Label: "Adresa", Width: "40"},
		{Name: "pobro", Label: "Postanski broj", Width: "8"},
		{Name: "mesto", Label: "Mesto", Width: "40"},
		{Name: "pib", Label: "PIB", Width: "12"},
		{Name: "jmbg", Label: "JMBG", Width: "15"},
		{Name: "bpg", Label: "BPG"},
		{Name: "index", Label: "Index"},
		{Name: "gln", Label: "GLN"},
		{Name: "jib", Label: "JIB"},
		{Name: "ziro", Label: "Ziro"},
		{Name: "matbr", Label: "Maticni Broj"},
		{Name: "konta", Label: "Konta"},
		{Name: "tippdv", Label: "Tip PDV"},
		{Name: "email", Label: "E-Mail"},
		{Name: "telefon", Label: "Telefon"},
		{Name: "kontaktosb", Label: "Kontakt osoba"},
		{Name: "budzetski", Label: "Budzetski"},
		{Name: "jbkjs", Label: "JBKJS"},
		{Name: "napomena", Label: "Naponema"},
		{Name: "idpartneri", Label: "ID Partneri"},
	}
}
func SetTipanalitikeFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifraanalitike", Label: "Sifra Analitike", Width: "4", Sortable: true},
		{Name: "naziv", Label: "Naziv Analitike", Width: "50", Sortable: true},
	}
}
func SetJedmereFields() []domain.Fields {
	return []domain.Fields{
		{Name: "jm", Label: "Jed mere", Width: "4", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "20", Sortable: true},
		{Name: "brdecimala", Label: "Broj decimala"},
		{Name: "ima_duzinu", Label: "Ima duzinu"},
		{Name: "ima_sirinu", Label: "Ima sirinu"},
		{Name: "ima_komade", Label: "Ima komade"},
		{Name: "koristi_spectez", Label: "Koristi spectez"},
	}
}
func SetRgruFields() []domain.Fields {
	return []domain.Fields{
		{Name: "gru", Label: "Grupa", Sortable: true},
		{Name: "naziv", Label: "Naziv", Sortable: true},
	}
}
func SetRpgruFields() []domain.Fields {
	return []domain.Fields{
		{Name: "gru", Label: "Grupa", Sortable: true},
		{Name: "pgru", Label: "Podgrupa", Width: "40", Sortable: true},
		{Name: "naziv", Label: "Naziv", Sortable: true},
	}
}

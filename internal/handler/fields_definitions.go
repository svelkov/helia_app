package handler

import "helia/internal/domain"

func SetBankeFields() []domain.Fields {
	return []domain.Fields{
		{Name: "brrac", Label: "Broj Racuna", ControlWidth: " w-60", MaxLength: "40", Sortable: true},
		{Name: "banka", Label: "Banka", ControlWidth: " w-96", MaxLength: "60", Sortable: true},
		{Name: "konto", Label: "Konto", ControlWidth: " w-16", MaxLength: "6", Sortable: true},
		{Name: "sifra", Label: "Sifra", ControlWidth: " w-16", MaxLength: "6", Sortable: true},
		{Name: "bnkcod", Label: "Kod Banke", ControlWidth: " w-12", MaxLength: "3", Sortable: true},
		{Name: "ebank", Label: "E-Banking", ControlWidth: " w-60", MaxLength: "60", Sortable: true},
		{Name: "pocnazfajl", Label: "Poc Naz Fajla", ControlWidth: " w-12", MaxLength: "8", Sortable: true},
		{Name: "tipdok", Label: "Tip Dokumenta", ControlWidth: " w-12", MaxLength: "4", Sortable: true},
		{Name: "nafakne", Label: "Na Fakt Ne", ControlWidth: " w-4", MaxLength: "2", Sortable: true},
	}
}

func SetBnkizvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifbank", Label: "Sifra Banke", ControlWidth: " w-10", MaxLength: "5", Width: "6", Sortable: true},
		{Name: "bnkdes", Label: "Naziv Banke", ControlWidth: " w-160", MaxLength: "80", Width: "80", Sortable: true},
		{Name: "swiftadr", Label: "Swift Adresa", ControlWidth: " w-160", MaxLength: "80", Width: "80", Sortable: true},
		{Name: "brojrac", Label: "Broj Racuna", ControlWidth: " w-120", MaxLength: "50", Width: "55", Sortable: true},
		{Name: "beneficiary", Label: "Beneficiary", ControlWidth: " w-160", MaxLength: "70", Width: "70", Sortable: true},
		{Name: "corrbank", Label: "Korespodentna banka", ControlWidth: " w-200", MaxLength: "90", Width: "90", Sortable: true},
		{Name: "tel", Label: "Telefon", ControlWidth: " w-48", MaxLength: "24", Width: "25", Sortable: true},
		{Name: "fax", Label: "Fax", ControlWidth: " w-60", MaxLength: "30", Width: "30", Sortable: true},
		{Name: "address", Label: "Adresa", ControlWidth: " w-80", MaxLength: "40", Width: "40", Sortable: true},
		{Name: "komentar", Label: "Komentar", ControlWidth: " w-160", MaxLength: "70", Width: "70", Sortable: true},
	}
}

func SetDokvrstaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vrd", Label: "Vrsta dokumenta", ControlWidth: " w-20", MaxLength: "10", Type: "text", Width: "5", Sortable: true},
		{Name: "opis", Label: "Opis vrste dokumenta", ControlWidth: " w-80", MaxLength: "40", Type: "text", Width: "40", Sortable: true},
		{Name: "kodknj", Label: "Nacin knjizenja (d,p,dp)", ControlWidth: " w-8", MaxLength: "2", Type: "text", Width: "4", Sortable: true},
		{Name: "predznak", Label: "Predznak", ControlWidth: " w-8", MaxLength: "1", Type: "text", Width: "3", Sortable: true},
		{Name: "grpdok", Label: "Grupa dokumenta", ControlWidth: " w-20", MaxLength: "10", Type: "text", Width: "10", Sortable: true},
		{Name: "dokozn", Label: "Oznaka dokumenta", ControlWidth: " w-10", MaxLength: "5", Type: "text", Width: "5", Sortable: true},
		{Name: "stornovrd", Label: "Storno vrsta dokumenta", ControlWidth: " w-20", MaxLength: "10", Type: "text", Width: "6", Sortable: true},
		{Name: "modul", Label: "Modul", ControlWidth: " w-8", MaxLength: "4", Type: "text", Width: "8", Sortable: true},
	}
}

func SetDrzaveFields() []domain.Fields {
	return []domain.Fields{
		{Name: "naziv", Label: "Naziv drzave", ControlWidth: " w-160", MaxLength: "80", Width: "40", Sortable: true},
		{Name: "ozndrz", Label: "Oznaka drzave", ControlWidth: " w-10", MaxLength: "5", Width: "6", Sortable: true},
	}
}
func SetFevpdvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vktip", Label: "Tip I/U", ControlWidth: " w-8", MaxLength: "1", Width: "4", Sortable: true},
		{Name: "opis", Label: "Opis", ControlWidth: " w-240", MaxLength: "500", Width: "100", Sortable: true},
		{Name: "vkrbr", Label: "Vrsta r br", ControlWidth: " w-20", MaxLength: "10", Width: "3", Sortable: true},
		{Name: "obrazac", Label: "Obrazac", ControlWidth: " w-8", MaxLength: "3", Width: "6", Sortable: true},
	}
}

func SetMestotroskaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mtroska", Label: "Mesto troska", ControlWidth: " w-12", MaxLength: "6", Width: "6", Sortable: true},
		{Name: "opis", Label: "Opis", ControlWidth: " w-120", MaxLength: "60", Width: "45", Sortable: true},
		{Name: "idorgjed", Label: "Org. jedinica", ControlWidth: " w-20", MaxLength: "10", Width: "20", Sortable: true},
	}
}

func SetOrgjedFields() []domain.Fields {
	return []domain.Fields{
		{Name: "ojozn", Label: "Sifra Orgjed", Width: "6", ControlWidth: " w-12", MaxLength: "5", Sortable: true},
		{Name: "naziv", Label: "Naziv Orgjed", Width: "45", ControlWidth: " w-96", MaxLength: "40", Sortable: true},
	}
}

// key of the map must be the name of filed in the table in db (we need it for mapping)

func SetPopdvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "popdv", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10", Sortable: true},
		{Name: "opis", Label: "Opis", ControlWidth: " w-240", MaxLength: "500", ValidationText: "Morate uneti opis dokumenta...", Width: "60", Sortable: true},
		{Name: "grpdok", Label: "Grupa dok", ValidationText: "Morate uneti grupu dokumenata...", Width: "20", Sortable: true},
		{Name: "grpvrd", Label: "Grp. Vrste Dok.", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20", Sortable: true},
		{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10", Sortable: true},
	}
}

func SetSifmestoFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifm", Label: "Sifra Mesta", ControlWidth: " w-20", MaxLength: "10", Width: "4", Sortable: true},
		{Name: "naziv", Label: "Naziv", ControlWidth: " w-160", MaxLength: "80", Width: "80", Sortable: true},
		{Name: "ops", Label: "Opstina", ControlWidth: " w-20", MaxLength: "10", Width: "8", Sortable: true},
		{Name: "pobro", Label: "Postanski broj", ControlWidth: " w-20", MaxLength: "10", Width: "8", Sortable: true},
		{Name: "km", Label: "Km", ControlWidth: " w-20", MaxLength: "10", Width: "6", Sortable: true},
	}
}

func SetSifopFields() []domain.Fields {
	return []domain.Fields{
		{Name: "ops", Label: "Sifra Opstine", MaxLength: "4", ControlWidth: " w-12", Sortable: true},
		{Name: "naziv", Label: "Naziv Opstine", MaxLength: "50", ControlWidth: " w-96", Sortable: true},
	}
}
func SetSifplizvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifplac", Label: "Sifra Placanja", ControlWidth: " w-10", MaxLength: "5", ValidationText: "morate uneti sifru placanja...", Width: "6", Sortable: true},
		{Name: "oblik", Label: "Oblik", ControlWidth: " w-10", MaxLength: "5", ValidationText: "morate uneti oblik placanja...", Width: "6", Sortable: true},
		{Name: "osnov", Label: "Osnov placanja", ControlWidth: " w-10", MaxLength: "5", ValidationText: "morate uneti osnov placanja", Width: "4", Sortable: true},
		{Name: "opis", Label: "Opis", ControlWidth: " w-160", MaxLength: "80", ValidationText: "morate uneti opis", Width: "80", Sortable: true},
		{Name: "konto", Label: "Konto", ControlWidth: " w-12", MaxLength: "6", Width: "8", Sortable: true},
		{Name: "sifra", Label: "Sifra", ControlWidth: " w-12", MaxLength: "6", Width: "8", Sortable: true},
	}
}
func SetTipdokFields() []domain.Fields {
	return []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", ControlWidth: " w-8", MaxLength: "4", ValidationText: "Morate uneti tip dokumenta...", Width: "10", Sortable: true},
		{Name: "opis", Label: "Opis", ControlWidth: " w-80", MaxLength: "40", ValidationText: "Morate uneti opis dokumenta...", Width: "60", Sortable: true},
		{Name: "grpdok", Label: "Grupa dok", ControlWidth: " w-20", MaxLength: "10", ValidationText: "Morate uneti grupu dokumenata...", Width: "20", Sortable: true},
		{Name: "grpvrd", Label: "Grp Vrste Dok", ControlWidth: " w-160", MaxLength: "80", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20", Sortable: true},
		{Name: "magacin", Label: "Magacin", ControlWidth: " w-120", MaxLength: "50", ValidationText: "", Width: "10", Sortable: true},
	}
}

func SetFvknjracFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vktip", Label: "Tip knjige (ulazni/izlazni racun)", ControlWidth: " w-8", MaxLength: "1", Width: "2", Sortable: true},
		{Name: "opis", Label: "Opis", ControlWidth: " w-120", MaxLength: "50", Width: "50", Sortable: true},
		{Name: "vkrbr", Label: "Vrsta knjige racuna", ControlWidth: " w-20", MaxLength: "10", Width: "2", Sortable: true},
		{Name: "konta", Label: "Konta", ControlWidth: " w-120", MaxLength: "60", Width: "60", Sortable: true},
	}
}
func SetPartneriFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifra", Label: "Sifra", ControlWidth: " w-12", MaxLength: "6", ValidationText: "morate uneti sifru partnera...", Width: "8"},
		{Name: "naziv", Label: "Naziv", ControlWidth: " w-240", MaxLength: "120", ValidationText: "morate uneti naziv partnera..", Width: "60"},
		{Name: "adresa", Label: "Adresa", ControlWidth: " w-120", MaxLength: "60", Width: "40"},
		{Name: "pobro", Label: "Postanski broj", ControlWidth: " w-20", MaxLength: "10", Width: "8"},
		{Name: "mesto", Label: "Mesto", ControlWidth: " w-120", MaxLength: "48", Width: "40"},
		{Name: "pib", Label: "PIB", ControlWidth: " w-24", MaxLength: "12", Width: "12"},
		{Name: "jmbg", Label: "JMBG", ControlWidth: " w-28", MaxLength: "13", Width: "15"},
		{Name: "bpg", Label: "BPG", ControlWidth: " w-24", MaxLength: "12"},
		{Name: "index", Label: "Index", ControlWidth: " w-28", MaxLength: "13"},
		{Name: "gln", Label: "GLN", ControlWidth: " w-40", MaxLength: "20"},
		{Name: "jib", Label: "JIB", ControlWidth: " w-28", MaxLength: "13"},
		{Name: "ziro", Label: "Ziro", ControlWidth: " w-120", MaxLength: "42"},
		{Name: "matbr", Label: "Maticni Broj", ControlWidth: " w-64", MaxLength: "32"},
		{Name: "konta", Label: "Konta", ControlWidth: " w-240", MaxLength: "120"},
		{Name: "tippdv", Label: "Tip PDV", ControlWidth: " w-10", MaxLength: "5"},
		{Name: "email", Label: "E-Mail", ControlWidth: " w-120", MaxLength: "60"},
		{Name: "telefon", Label: "Telefon", ControlWidth: " w-60", MaxLength: "30"},
		{Name: "kontaktosb", Label: "Kontakt osoba", ControlWidth: " w-120", MaxLength: "50"},
		{Name: "budzetski", Label: "Budzetski", ControlWidth: " w-4"},
		{Name: "jbkjs", Label: "JBKJS", ControlWidth: " w-10", MaxLength: "5"},
		{Name: "napomena", Label: "Naponema", ControlWidth: " w-240", MaxLength: "1024"},
		{Name: "idpartneri", Label: "ID Partneri", ControlWidth: " w-20", MaxLength: "10"},
	}
}
func SetTipanalitikeFields() []domain.Fields {
	return []domain.Fields{
		{Name: "tipanalitikeid", Label: "Sifra Analitike", ControlWidth: " w-20", MaxLength: "19", Width: "4", Sortable: true},
		{Name: "naziv", Label: "Naziv Analitike", ControlWidth: " w-120", MaxLength: "50", Width: "50", Sortable: true},
	}
}

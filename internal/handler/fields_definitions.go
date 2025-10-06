package handler

import "helia/internal/domain"

func SetBankeFields() []domain.Fields {
	return []domain.Fields{
		{Name: "brrac", Label: "Broj Racuna", Width: "45"},
		{Name: "banka", Label: "Banka", Width: "60"},
		{Name: "konto", Label: "Konto", Width: "8"},
		{Name: "sifra", Label: "Sifra", Width: "8"},
		{Name: "bnkcod", Label: "Kod Banke", Width: "4"},
		{Name: "ebank", Label: "E-Banking", Width: "60"},
		{Name: "pocnazfajl", Label: "Poc. Naz. Fajla", Width: "10"},
		{Name: "tipdok", Label: "Tip Dokumenta", Width: "6"},
		{Name: "nafakne", Label: "Na Fakt Ne", Width: "1"},
	}
}

func SetBnkizvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifbank", Label: "Sifra Banke", Width: "6"},
		{Name: "bnkdes", Label: "Naziv Banka", Width: "80"},
		{Name: "swiftadr", Label: "Swift Adresa", Width: "80"},
		{Name: "brojrac", Label: "Broj Racuna", Width: "55"},
		{Name: "beneficiary", Label: "Beneficiary", Width: "70"},
		{Name: "corrbank", Label: "Korespodentna banka", Width: "90"},
		{Name: "tel", Label: "Telefon", Width: "25"},
		{Name: "fax", Label: "Fax", Width: "30"},
		{Name: "address", Label: "Adresa", Width: "40"},
		{Name: "komentar", Label: "Komentar", Width: "70"},
	}
}

func SetDokvrstaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vrd", Label: "Vrsta dokumenta", Type: "text", Width: "5"},
		{Name: "opis", Label: "Opis vrste dokumenta", Type: "text", Width: "40"},
		{Name: "kodknj", Label: "Nacin knjizenja (d,p,dp)", Type: "text", Width: "4"},
		{Name: "predznak", Label: "Predznak", Type: "text", Width: "3"},
		{Name: "grpdok", Label: "Grupa dokumenta", Type: "text", Width: "10"},
		{Name: "dokozn", Label: "Oznaka dokumenta", Type: "text", Width: "5"},
		{Name: "stornovrd", Label: "Storno vrsta dokumenta", Type: "text", Width: "6"},
		{Name: "modul", Label: "Modul", Type: "text", Width: "8"},
	}
}

func SetDrzaveFields() []domain.Fields {
	return []domain.Fields{
		{Name: "naziv", Label: "Naziv drzave", Width: "40"},
		{Name: "ozndrz", Label: "Oznaka", Width: "6"},
	}
}
func SetFevpdvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vktip", Label: "Tip I/U", Width: "4"},
		{Name: "opis", Label: "Opis", Width: "100"},
		{Name: "vkrbr", Label: "Vrsta r.br.", Width: "3"},
		{Name: "obrazac", Label: "Obrazac", Width: "6"},
	}
}

func SetMestotroskaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mtroska", Label: "Mestotroska", Width: "6"},
		{Name: "opis", Label: "Opis", Width: "45"},
		{Name: "idorgjed", Label: "Org. jedinica", Width: "20"},
	}
}

func SetOrgjedFields() []domain.Fields {
	return []domain.Fields{
		{Name: "ojozn", Label: "Sifra Orgjed", Width: "6"},
		{Name: "naziv", Label: "Naziv Orgjed", Width: "45"},
	}
}

// key of the map must be the name of filed in the table in db (we need it for mapping)

func SetPopdvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "popdv", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10"},
		{Name: "opis", Label: "Opis", ValidationText: "Morate uneti opis dokumenta...", Width: "60"},
		{Name: "grpdok", Label: "Grupa Dok.", ValidationText: "Morate uneti grupu dokumenata...", Width: "20"},
		{Name: "grpvrd", Label: "Grp. Vrste Dok.", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20"},
		{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10"},
	}
}

func SetSifmestoFileds() []domain.Fields {
	return []domain.Fields{
		{Name: "sifm", Label: "Sifra Mesta", Width: "4"},
		{Name: "naziv", Label: "Naziv", Width: "80"},
		{Name: "ops", Label: "Opstina", Width: "8"},
		{Name: "pobro", Label: "Postanski broj", Width: "8"},
		{Name: "km", Label: "Km", Width: "6"},
	}
}

func SetSifopFields() []domain.Fields {
	return []domain.Fields{
		{Name: "ops", Label: "Sifra Opstine", Width: "6"},
		{Name: "naziv", Label: "Naziv Opstine", Width: "60"},
	}
}
func SetSifplizvFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifplac", Label: "Sifra Placanja", ValidationText: "morate uneti sifru placanja...", Width: "6"},
		{Name: "oblik", Label: "Oblik", ValidationText: "morate uneti oblik placanja...", Width: "6"},
		{Name: "osnov", Label: "Osnov placanja", ValidationText: "morate uneti osnov placanja", Width: "4"},
		{Name: "opis", Label: "Opis", ValidationText: "morate uneti opis", Width: "80"},
		{Name: "konto", Label: "Konto", Width: "8"},
		{Name: "sifra", Label: "Sifra", Width: "8"},
	}
}
func SetTipdokFields() []domain.Fields {
	return []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10"},
		{Name: "opis", Label: "Opis", ValidationText: "Morate uneti opis dokumenta...", Width: "60"},
		{Name: "grpdok", Label: "Grupa Dok.", ValidationText: "Morate uneti grupu dokumenata...", Width: "20"},
		{Name: "grpvrd", Label: "Grp. Vrste Dok.", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20"},
		{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10"},
	}
}

func SetFvknjracFields() []domain.Fields {
	return []domain.Fields{
		{Name: "vktip", Label: "Tip knjige (ulazni/izlazni racun)", Width: "2"},
		{Name: "opis", Label: "Opis", Width: "50"},
		{Name: "vkrbr", Label: "Vrsta knjige racuna", Width: "2"},
		{Name: "konta", Label: "Konta", Width: "60"},
	}
}

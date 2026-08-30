package handler

import (
	"helia/internal/domain"
)

// SetRgruFields returns the fields configuration for Rgru table
func SetRgruFields() []domain.Fields {
	return []domain.Fields{
		{Name: "gru", Label: "Grupa", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "55"},
	}
}

// SetRpgruFields returns the fields configuration for Rpgru table
func SetRpgruFields() []domain.Fields {
	return []domain.Fields{
		{Name: "gru", Label: "Grupa", Width: "10"},
		{Name: "pgru", Label: "Podgrupa", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "50"},
	}
}

// SetJedmereFields returns the fields configuration for Jedmere table
func SetJedmereFields() []domain.Fields {
	return []domain.Fields{
		{Name: "jm", Label: "Jedinica Mere", Width: "15"},
		{Name: "opis", Label: "Opis", Width: "50"},
		{Name: "brdecimala", Label: "Broj Decimala", Width: "10"},
		{Name: "imaduzinu", Label: "Ima Dužinu", Type: "checkbox", Width: "12"},
		{Name: "imasirinu", Label: "Ima Širinu", Type: "checkbox", Width: "12"},
		{Name: "imakomade", Label: "Ima Komade", Type: "checkbox", Width: "12"},
		{Name: "koristespectez", Label: "Koristi Spec. Težinu", Type: "checkbox", Width: "12"},
	}
}

// SetMagkontoFields returns the fields configuration for Magkonto table
func SetMagkontoFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "10"},
		{Name: "konto", Label: "Konto", Width: "12"},
		{Name: "vkonta", Label: "Vrsta Konta", Width: "12"},
		{Name: "kontoprih", Label: "Konto Prihoda", Width: "12"},
		{Name: "kontotroska", Label: "Konto Troška", Width: "12"},
		{Name: "kontoruc", Label: "Konto Računa", Width: "12"},
		{Name: "kontorab", Label: "Konto Rabata", Width: "12"},
	}
}

// SetMagaciniFields returns the fields configuration for Magacini table
func SetMagaciniFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8"},
		{Name: "opis", Label: "Opis", Width: "25"},
		{Name: "tipmag", Label: "Tip Magacina", Width: "12"},
		{Name: "adresa", Label: "Adresa", Width: "20"},
		{Name: "pobro", Label: "Pošt. Broj", Width: "10"},
		{Name: "mesto", Label: "Mesto", Width: "15"},
		{Name: "nadmag", Label: "Nadređeni Magacin", Width: "12"},
		{Name: "magosoba", Label: "Osoba Magacina", Width: "15"},
		{Name: "tel", Label: "Telefon", Width: "12"},
		{Name: "fax", Label: "Faks", Width: "12"},
		{Name: "tipzal", Label: "Tip Zalihе", Width: "10"},
		{Name: "tipcene", Label: "Tip Cena", Width: "10"},
		{Name: "nacvodzal", Label: "Način Vođenja Zalihе", Width: "12"},
		{Name: "analiza", Label: "Analiza", Width: "8"},
		{Name: "email", Label: "Email", Width: "20"},
		{Name: "tipart", Label: "Tip Artikla", Width: "12"},
	}
}

func SetRporFields() []domain.Fields {
	return []domain.Fields{
		{Name: "datum", Label: "Datum od kad važi stopa", Width: "18"},
		{Name: "pp", Label: "Šifra poreske tarife", Width: "15"},
		{Name: "pt", Label: "Poreska tarifa", Width: "15"},
		{Name: "tip", Label: "Oznaka za vrstu poreza", Width: "15"},
		{Name: "po", Label: "Poreska stopa", Width: "12"},
		{Name: "slovo", Label: "Slovo u MP", Width: "10"},
	}
}

// SetKomercijalistiFileds returns the fields configuration for Komercijalisti table
func SetKomercijalistiFileds() []domain.Fields {
	return []domain.Fields{
		{Name: "komid", Label: "ID", Width: "8"},
		{Name: "sifkom", Label: "Šifra", Width: "10"},
		{Name: "sifnadred", Label: "Nadređeni", Width: "12"},
		{Name: "imeprezime", Label: "Ime i Prezime", Width: "25"},
		{Name: "adresa", Label: "Adresa", Width: "20"},
		{Name: "mesto", Label: "Mesto", Width: "15"},
		{Name: "telposao", Label: "Telefon Posao", Width: "15"},
		{Name: "telmob", Label: "Telefon Mobilni", Width: "15"},
		{Name: "totprod", Label: "Ukupna Prodaja", Width: "15"},
		{Name: "totprofit", Label: "Ukupan Profit", Width: "15"},
		{Name: "zaddatprod", Label: "Poslednji Datum Prodaje", Width: "18"},
		{Name: "totnaplaceno", Label: "Ukupno Naplaćeno", Width: "15"},
		{Name: "loginname", Label: "Login Name", Width: "15"},
	}
}

// SetArtikliFields returns the fields configuration for Artikli/Rsif table
func SetArtikliFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rsifid", Label: "ID", Width: "8"},
		{Name: "sifra", Label: "Šifra artikla", Width: "12"},
		{Name: "naziv", Label: "Naziv artikla", Width: "30"},
		{Name: "komercopis", Label: "Komercijalni opis", Width: "35"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "pro", Label: "Proizvođač", Width: "20"},
		{Name: "barkod", Label: "Barkod", Width: "18"},
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "tip", Label: "Tip", Width: "6"},
		{Name: "model", Label: "Model", Width: "8"},
		{Name: "kvalitet", Label: "Kvalitet", Width: "12"},
		{Name: "serbr", Label: "Serijski broj", Width: "15"},
		{Name: "zemljaproizv", Label: "Zemlja proizvodnje", Width: "15"},
	}
}

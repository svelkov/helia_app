package handler

import (
	"helia/internal/domain"
)

// SetRgruFields returns the fields configuration for Rgru table
func SetRgruFields() []domain.Fields {
	return []domain.Fields{
		{Name: "gru", Label: "Grupa", ControlWidth: " w-20", MaxLength: "10", Width: "10"},
		{Name: "naziv", Label: "Naziv", ControlWidth: " w-120", MaxLength: "55", Width: "55"},
	}
}

// SetRpgruFields returns the fields configuration for Rpgru table
func SetRpgruFields() []domain.Fields {
	return []domain.Fields{
		{Name: "gru", Label: "Grupa", ControlWidth: " w-20", MaxLength: "10", Width: "10"},
		{Name: "pgru", Label: "Podgrupa", ControlWidth: " w-20", MaxLength: "10", Width: "10"},
		{Name: "naziv", Label: "Naziv", ControlWidth: " w-80", MaxLength: "40", Width: "50"},
	}
}

// SetJedmereFields returns the fields configuration for Jedmere table
func SetJedmereFields() []domain.Fields {
	return []domain.Fields{
		{Name: "jm", Label: "Jedinica Mere", ControlWidth: " w-16", MaxLength: "8", Width: "15"},
		{Name: "opis", Label: "Opis", ControlWidth: " w-120", MaxLength: "50", Width: "50"},
		{Name: "brdecimala", Label: "Broj Decimala", ControlWidth: " w-10", MaxLength: "5", Width: "10"},
		{Name: "imaduzinu", Label: "Ima Dužinu", ControlWidth: " w-4", Type: "checkbox", Width: "12"},
		{Name: "imasirinu", Label: "Ima Širinu", ControlWidth: " w-4", Type: "checkbox", Width: "12"},
		{Name: "imakomade", Label: "Ima Komade", ControlWidth: " w-4", Type: "checkbox", Width: "12"},
		{Name: "koristespectez", Label: "Koristi Spec. Težinu", ControlWidth: " w-4", Type: "checkbox", Width: "12"},
	}
}

// SetTipknpismaFields returns the fields configuration for Tipknpisma table.
func SetTipknpismaFields() []domain.Fields {
	return []domain.Fields{
		{Name: "sifrazlog", Label: "Šifra razloga", ControlWidth: " w-10", MaxLength: "5", Width: "15"},
		{Name: "opis", Label: "Opis", ControlWidth: " w-240", MaxLength: "120", Width: "50"},
	}
}

// SetMagkontoFields returns the fields configuration for Magkonto table
func SetMagkontoFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mag", Label: "Magacin", ControlWidth: " w-10", MaxLength: "5", Width: "10"},
		{Name: "konto", Label: "Konto", ControlWidth: " w-12", MaxLength: "6", Width: "12"},
		{Name: "vkonta", Label: "Vrsta Konta", ControlWidth: " w-10", MaxLength: "5", Width: "12"},
		{Name: "kontoprih", Label: "Konto Prihoda", ControlWidth: " w-12", MaxLength: "6", Width: "12"},
		{Name: "kontotroska", Label: "Konto Troška", ControlWidth: " w-12", MaxLength: "6", Width: "12"},
		{Name: "kontoruc", Label: "Konto Računa", ControlWidth: " w-12", MaxLength: "6", Width: "12"},
		{Name: "kontorab", Label: "Konto Rabata", ControlWidth: " w-12", MaxLength: "6", Width: "12"},
	}
}

// SetMagaciniFields returns the fields configuration for Magacini table
func SetMagaciniFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mag", Label: "Magacin", ControlWidth: " w-10", MaxLength: "5", Width: "8"},
		{Name: "opis", Label: "Opis", ControlWidth: " w-80", MaxLength: "40", Width: "25"},
		{Name: "tipmag", Label: "Tip Magacina", ControlWidth: " w-8", MaxLength: "2", Width: "12"},
		{Name: "adresa", Label: "Adresa", ControlWidth: " w-80", MaxLength: "40", Width: "20"},
		{Name: "pobro", Label: "Pošt. Broj", ControlWidth: " w-20", MaxLength: "10", Width: "10"},
		{Name: "mesto", Label: "Mesto", ControlWidth: " w-120", MaxLength: "50", Width: "15"},
		{Name: "nadmag", Label: "Nadređeni Magacin", ControlWidth: " w-20", MaxLength: "10", Width: "12"},
		{Name: "magosoba", Label: "Osoba Magacina", ControlWidth: " w-120", MaxLength: "50", Width: "15"},
		{Name: "tel", Label: "Telefon", ControlWidth: " w-80", MaxLength: "40", Width: "12"},
		{Name: "fax", Label: "Faks", ControlWidth: " w-80", MaxLength: "40", Width: "12"},
		{Name: "tipzal", Label: "Tip Zalihе", ControlWidth: " w-10", MaxLength: "5", Width: "10"},
		{Name: "tipcene", Label: "Tip Cena", ControlWidth: " w-10", MaxLength: "5", Width: "10"},
		{Name: "nacvodzal", Label: "Način Vođenja Zalihе", ControlWidth: " w-10", MaxLength: "5", Width: "12"},
		{Name: "analiza", Label: "Analiza", ControlWidth: " w-10", MaxLength: "5", Width: "8"},
		{Name: "email", Label: "Email", ControlWidth: " w-120", MaxLength: "60", Width: "20"},
		{Name: "tipart", Label: "Tip Artikla", ControlWidth: " w-40", MaxLength: "20", Width: "12"},
	}
}

func SetRporFields() []domain.Fields {
	return []domain.Fields{
		{Name: "datum", Label: "Datum od kad važi stopa", ControlWidth: " w-32", Width: "18"},
		{Name: "pp", Label: "Šifra poreske tarife", ControlWidth: " w-32", MaxLength: "15", Width: "15"},
		{Name: "pt", Label: "Poreska tarifa", ControlWidth: " w-32", MaxLength: "15", Width: "15"},
		{Name: "tip", Label: "Oznaka za vrstu poreza", ControlWidth: " w-20", MaxLength: "10", Width: "15"},
		{Name: "po", Label: "Poreska stopa", ControlWidth: " w-10", MaxLength: "5", Width: "12"},
		{Name: "slovo", Label: "Slovo u MP", ControlWidth: " w-8", MaxLength: "1", Width: "10"},
	}
}

// SetKomercijalistiFileds returns the fields configuration for Komercijalisti table
func SetKomercijalistiFileds() []domain.Fields {
	return []domain.Fields{
		{Name: "komid", Label: "ID", ControlWidth: " w-20", MaxLength: "10", Width: "8"},
		{Name: "sifkom", Label: "Šifra", ControlWidth: " w-20", MaxLength: "10", Width: "10"},
		{Name: "sifnadred", Label: "Nadređeni", ControlWidth: " w-20", MaxLength: "10", Width: "12"},
		{Name: "imeprezime", Label: "Ime i Prezime", ControlWidth: " w-60", MaxLength: "30", Width: "25"},
		{Name: "adresa", Label: "Adresa", ControlWidth: " w-120", MaxLength: "50", Width: "20"},
		{Name: "mesto", Label: "Mesto", ControlWidth: " w-120", MaxLength: "50", Width: "15"},
		{Name: "telposao", Label: "Telefon Posao", ControlWidth: " w-40", MaxLength: "20", Width: "15"},
		{Name: "telmob", Label: "Telefon Mobilni", ControlWidth: " w-40", MaxLength: "20", Width: "15"},
		{Name: "totprod", Label: "Ukupna Prodaja", ControlWidth: " w-32", MaxLength: "15", Width: "15"},
		{Name: "totprofit", Label: "Ukupan Profit", ControlWidth: " w-32", MaxLength: "15", Width: "15"},
		{Name: "zaddatprod", Label: "Poslednji Datum Prodaje", ControlWidth: " w-32", Width: "18"},
		{Name: "totnaplaceno", Label: "Ukupno Naplaćeno", ControlWidth: " w-32", MaxLength: "15", Width: "15"},
		{Name: "loginname", Label: "Login Name", ControlWidth: " w-32", MaxLength: "15", Width: "15"},
	}
}

// SetArtikliFields returns the fields configuration for Artikli/Rsif table
func SetArtikliFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rsifid", Label: "ID", ControlWidth: " w-20", MaxLength: "10", Width: "8"},
		{Name: "sifra", Label: "Šifra artikla", ControlWidth: " w-20", MaxLength: "10", Width: "12"},
		{Name: "naziv", Label: "Naziv artikla", ControlWidth: " w-240", MaxLength: "180", Width: "30"},
		{Name: "komercopis", Label: "Komercijalni opis", ControlWidth: " w-240", MaxLength: "500", Width: "35"},
		{Name: "jm", Label: "JM", ControlWidth: " w-16", MaxLength: "8", Width: "8"},
		{Name: "pro", Label: "Proizvođač", ControlWidth: " w-80", MaxLength: "35", Width: "20"},
		{Name: "barkod", Label: "Barkod", ControlWidth: " w-32", MaxLength: "15", Width: "18"},
		{Name: "konto", Label: "Konto", ControlWidth: " w-12", MaxLength: "6", Width: "10"},
		{Name: "tip", Label: "Tip", ControlWidth: " w-8", MaxLength: "1", Width: "6"},
		{Name: "model", Label: "Model", ControlWidth: " w-8", MaxLength: "1", Width: "8"},
		{Name: "kvalitet", Label: "Kvalitet", ControlWidth: " w-20", MaxLength: "10", Width: "12"},
		{Name: "serbr", Label: "Serijski broj", ControlWidth: " w-40", MaxLength: "20", Width: "15"},
		{Name: "zemljaproizv", Label: "Zemlja proizvodnje", ControlWidth: " w-60", MaxLength: "30", Width: "15"},
	}
}

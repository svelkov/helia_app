package handler

import (
	"helia/internal/domain"
	"helia/internal/service"
)

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

type PartneriHandler struct {
	Service *service.BaseService[domain.Partneri]
}

func NewPartneriHandler(service *service.BaseService[domain.Partneri]) *PartneriHandler {
	return &PartneriHandler{Service: service}
}

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var komProUpis = domain.KomProfakturaUpis{RdokID: 45, Vrd: 360, Tipdok: "43", Dokum: 7, Dadok: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Rok: 15,
	MagaciniID: 3, Mag: 1, Fkto: "2040", Fana: "00031", KomID: sql.NullInt64{Int64: 5, Valid: true}, Orgjed: sql.NullInt64{Int64: 2, Valid: true},
	MestoTroska: sql.NullInt64{Int64: 5, Valid: true}, Foot: "Hitno", SifVal: 941, Kurs: 1, VaziZa: 1, Otp: "KAMION", Pla: "VIRMAN",
	Posuslnab: "160-1  Banka", Mi: 2, Ugrabat: 3, Pkase: 2, Paritet: "FCO", Izjizv: 0, Keykat: "PDV-RS-24-1-2", Category: "S",
	Pornapomena: "član 24 stav 1", Kamt: "zakonska", Brugov: "U-12"}

var komProStavkaUpis = domain.KomProStavkaUpis{RdokID: 45, Rbr: 2, Sifra: 1001, RsifID: 44, Po: 2, Konto: "1320", Mag: 1, MagaciniID: 3,
	Kolicina: 3, Cena: 80, Fcena: 117.25, CenaVal: 1, Rabat: 5, Otk: "001", Serija: "S1", Roktr: 82000, Vrd: 360, Dokum: 7,
	Dadok: sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true}}

// komercijalnoProCases are the statements of the profakture of the komercijalno module.
var komercijalnoProCases = []goldenCase{
	{name: "vrste of the profakture",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProVrste(ctx)
			return err
		}},
	{name: "vrste of the nalozi of the profakture",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProNalogVrste(ctx)
			return err
		}},
	{name: "vrste naloga za knjiženje of the vrsta",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProTipdok(ctx, 360)
			return err
		}},
	{name: "profaktura not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomProfaktura(ctx, 45)
			if found {
				return errFound
			}
			return err
		}},
	{name: "magacin of the profaktura", respond: answer(0, "V", 2, 1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, _, err := kom(r).KomProMagacin(ctx, 3)
			return err
		}},
	{name: "add a profaktura", respond: answer(0, int64(45)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomProDodaj(ctx, tx, komProUpis)
			return err
		}},
	{name: "change a profaktura", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomProIzmeni(ctx, tx, komProUpis)
		}},
	{name: "totals of a profaktura", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomProUkupno(ctx, tx, 45, 240, 10.5, 4.2, 12.5)
		}},
	{name: "confirm a profaktura", respond: answer(1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProPotvrdi(ctx, 45)
			return err
		}},
	{name: "kupac of the konto and šifra", respond: answer(0, "Metal", "Glavna 1", "11000", "Beograd", "100", 1, 3.0, 15, 1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomProKupac(ctx, " 2040", "00031 ")
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "mesto isporuke of the kupac", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomProMestoIsporuke(ctx, "2040", "00031", 2)
			return err
		}},
	{name: "cenovnik", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, _, _, err := kom(r).KomProCenovnik(ctx, 9)
			return err
		}},
	{name: "stavke of a profaktura",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProStavke(ctx, 45)
			return err
		}},
	{name: "add a stavka of a profaktura", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomProStavkaDodaj(ctx, tx, komProStavkaUpis)
		}},
	{name: "change a stavka of a profaktura", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomProStavkaIzmeni(ctx, tx, domain.KomProStavkaIzmena{RdokID: 45, RproID: 61, Kolicina: 4, Fcena: 117.25, CenaVal: 1, Rabat: 5})
		}},
	{name: "lots of an artikal with stanje",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProLotovi(ctx, 3, 44, true)
			return err
		}},
	{name: "all the lots of an artikal",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProLotovi(ctx, 3, 44, false)
			return err
		}},
	{name: "cena of an artikal", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProCenaArtikla(ctx, 1001)
			return err
		}},
	{name: "cena of a cenovnik", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := kom(r).KomProCenaCenovnika(ctx, 9, 1001)
			return err
		}},
	{name: "planska cena", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomProPlanskaCena(ctx, 1001)
			return err
		}},
	{name: "add the nalog of a profaktura", respond: answer(0, int64(46)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomProNalogDodaj(ctx, tx, 45, 350, 3)
			return err
		}},
	{name: "copy the stavke into the nalog", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomProNalogStavke(ctx, tx, 45, 46, 350, 3)
		}},
	{name: "mark the nalog formed", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomProNalogFormiran(ctx, tx, 45)
			return err
		}},
	{name: "pregled of the nepotvrđene of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProPrikaz(ctx, domain.KomProPrikazParams{MagaciniID: 3, Doksts: domain.KomProfakturaNova})
			return err
		}},
	{name: "pregled of the potvrđene of all the magacini with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomProPrikaz(ctx, domain.KomProPrikazParams{Doksts: domain.KomProfakturaPotvrdjena, SviMagacini: true, Tekst: "metal"})
			return err
		}},
}

func TestKomercijalnoProGolden(t *testing.T) {
	runGolden(t, "komercijalno_profaktura.golden", komercijalnoProCases)
}

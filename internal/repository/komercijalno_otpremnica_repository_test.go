package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var komOtpDatum = time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

var komOtpUpis = domain.KomOtpUpis{RdokID: 60, NalogID: 50, Vrd: 188, Brotp: 990001, Dokum: 990001, Datotp: komOtpDatum, Dadok: komOtpDatum,
	Otp: "KAMION", Vozac: "Pera", Brvozila: "NS-123", Bttowght: 12.5, Netwght: 10, Totboxs: 2, Brpaleta: 1, Foot: "Hitno", Napomena: "Zvati",
	SifVal: 941, Kurs: 1, VaziZa: 1, Posuslnab: "160-1 Banka", Polje: "3.2", Keykat: "PDV-RS-24", Category: "S", Pornapomena: "",
	Brugov: "U-1"}

// komercijalnoOtpCases are the statements of the otpremnice of the komercijalno module.
var komercijalnoOtpCases = []goldenCase{
	{name: "vrste of the otpremnice",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpVrste(ctx)
			return err
		}},
	{name: "otpremnica not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomOtpremnica(ctx, 60)
			if found {
				return errFound
			}
			return err
		}},
	{name: "nalog of an otpremnica not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomOtpNalog(ctx, 50)
			if found {
				return errFound
			}
			return err
		}},
	{name: "stavke of the otpremnica",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpStavke(ctx, 60)
			return err
		}},
	{name: "temporary brojevi", respond: answer(0, 990000, 990004),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomOtpPrivremeniBrojevi(ctx, nil, 188, 30)
			return err
		}},
	{name: "broj otpremnice exists", respond: answer(0, 1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpBrotpPostoji(ctx, 188, 30, 15, 60)
			return err
		}},
	{name: "broj fakture exists", respond: answer(0, 0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpRdokPostoji(ctx, nil, 30, 188, 15)
			return err
		}},
	{name: "add an otpremnica", respond: answer(0, int64(60)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpDodaj(ctx, tx, komOtpUpis)
			return err
		}},
	{name: "change an otpremnica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpIzmeni(ctx, tx, komOtpUpis)
		}},
	{name: "copy the stavke of the nalog", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpStavkeKopiraj(ctx, tx, 60, 50, false, false)
		}},
	{name: "copy the stavke of the nalog to a storno at a new kurs", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpStavkeKopiraj(ctx, tx, 60, 50, true, true)
		}},
	{name: "prodajne cene at the kurs", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpStavkePreracunaj(ctx, tx, 60)
		}},
	{name: "prodajna vrednost", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpProdvred(ctx, tx, 60, 1234.5)
		}},
	{name: "change a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpStavkaIzmeni(ctx, tx, 60, 71, 3, "001", "S1", 82000)
		}},
	{name: "mark the nalog", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpNalogFormiran(ctx, tx, 50)
			return err
		}},
	{name: "set the nalog back", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpNalogVrati(ctx, tx, 50)
			return err
		}},
	{name: "pregled of the nepotvrđene of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpPrikaz(ctx, domain.KomOtpPrikazParams{MagaciniID: 3, Doksts: domain.KomOtpNova})
			return err
		}},
	{name: "pregled of the potvrđene of all the magacini with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpPrikaz(ctx, domain.KomOtpPrikazParams{Doksts: domain.KomOtpPotvrdjena, SviMagacini: true, Tekst: "metal"})
			return err
		}},
	{name: "closed year", respond: answer(0, false),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomOtpGodinaZatvorena(ctx)
			return err
		}},
	{name: "final brojevi", respond: answer(0, 15, 101), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := kom(r).KomOtpNoviBrojevi(ctx, tx, 188, 30)
			return err
		}},
	{name: "prosečna nabavna cena", respond: answer(0, 55.5), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpRstaProsnc(ctx, tx, 3, 1001)
			return err
		}},
	{name: "neto težina", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpNetoTezina(ctx, tx, 60, 10.5)
		}},
	{name: "robni nalog of the day", respond: answer(0, int64(7)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, _, err := kom(r).KomOtpRnal(ctx, tx, "43", sql.NullTime{Time: komOtpDatum, Valid: true}, 3)
			return err
		}},
	{name: "new robni nalog", respond: answer(0, int64(8)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpRnalDodaj(ctx, tx, domain.KomOtpRnalUpis{Tipdok: "43", Danal: komOtpDatum, MagaciniID: 3, Mag: 30})
			return err
		}},
	{name: "one more document of the robni nalog", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpRnalDokument(ctx, tx, 8)
		}},
	{name: "faktura of the otpremnica", respond: answer(0, int64(90)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpRdokDodaj(ctx, tx, 60, 8, domain.KomOtpRdokUpis{Dokum: 101, Dokiz: 15, Opis: "Dok.br:101", DescriptionCode: 35})
			return err
		}},
	{name: "stavka of the faktura", respond: answer(0, int64(91)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpRproDodaj(ctx, tx, domain.KomOtpRproUpis{RnalID: 8, RdokID: 90, Sifra: 1001, RsifID: 44, Konto: "1320", JM: "KG",
				Ncena: 55.5, Kolic: 3, Iznos: 166.5, Cena: 55.5, Pcena: 50, Fcena: 100, CenaVal: 100, Rab: 5, Po: 1, Brst: 1, Rbr: 1, Otk: "001",
				Serija: "S1", Roktr: 82000, Taxcat: "S20", Taxexreco: "-", MagaciniID: 3, Mag: 30})
			return err
		}},
	{name: "cena of an artikal without one", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpRceneDodaj(ctx, tx, domain.KomOtpRceneUpis{RsifID: 44, Sifra: 1001, Ncena: 55.5, Rab: 5, Cena: 80, Fcena: 100})
		}},
	{name: "broj stavki of the faktura", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpRdokBrst(ctx, tx, 90, 2)
		}},
	{name: "avansi of the nalog", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpAvansi(ctx, tx, 50, 90)
			return err
		}},
	{name: "potvrda", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomOtpPotvrdi(ctx, tx, 60, 101, 15, "Otpremnica br.:15")
			return err
		}},
	{name: "isporučene količine of the nalog", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomOtpNalogIsporuceno(ctx, tx, 60)
		}},
}

func TestKomercijalnoOtpGolden(t *testing.T) {
	runGolden(t, "komercijalno_otpremnica.golden", komercijalnoOtpCases)
}

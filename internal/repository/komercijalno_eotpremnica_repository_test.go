package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var komEotpUpis = domain.KomEotpUpis{
	KomEotpPodaci: domain.KomEotpPodaci{OpasneMat: true, TipEotp: domain.KomEotpEksterna, Jm1: "kg", Jm2: "m3", Bttovol: 2.5,
		Datpoc:    sql.NullTime{Time: time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC), Valid: true},
		StvDatotp: sql.NullTime{Time: time.Date(2026, 3, 10, 9, 30, 0, 0, time.UTC), Valid: true},
		BrNarudz:  "N-1", KontIme: "Ana", ReferKup: "Mika", NacOtp: domain.KomEotpPrevoznik, PrevKto: "4350", PrevAna: "12", VozPrez: "Perić",
		UlicaOtp: "Glavna", MestoOtp: "Novi Sad", UlicaOdr: "Bulevar", UlBrOdr: "5", MestoOdr: "Beograd", FisNap: "Rampa 2"},
	Mi: 3,
}

// komercijalnoEotpCases are the statements of the e-otpremnice of the komercijalno module.
var komercijalnoEotpCases = []goldenCase{
	{name: "e-otpremnica not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomEotp(ctx, 60)
			if found {
				return errFound
			}
			return err
		}},
	{name: "save the e-otpremnica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomEotpIzmeni(ctx, tx, 60, komEotpUpis)
		}},
	{name: "faze vožnje",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomEotpFaze(ctx, 60)
			return err
		}},
	{name: "opasne materije",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomEotpOpasne(ctx, 60)
			return err
		}},
	{name: "delete the faze and the opasne materije", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomEotpBrisi(ctx, tx, 60)
		}},
	{name: "add a faza vožnje", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomEotpFazaDodaj(ctx, tx, 60, domain.KomEotpFaza{Rbr: 1, NacOtp: 2, Naziv: "Prevoz DOO"})
		}},
	{name: "add an opasna materija", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomEotpOpasnaDodaj(ctx, tx, 60, domain.KomEotpOpasna{Rbr: 1, Naziv: "UN 1203", Opis: "Benzin", Komentar: "ADR"})
		}},
	{name: "partner not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomEotpPartner(ctx, " 2040 ", "15")
			if found {
				return errFound
			}
			return err
		}},
	{name: "pošiljalac not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomEotpPosiljalac(ctx, "FIRMA")
			return err
		}},
	{name: "adresa of the magacin not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomEotpMagacinAdresa(ctx, 3)
			if found {
				return errFound
			}
			return err
		}},
	{name: "mesto isporuke not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomEotpMestoIsporuke(ctx, "2040", "15", 3)
			if found {
				return errFound
			}
			return err
		}},
	{name: "nalog not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, found, err := kom(r).KomEotpNalog(ctx, 50)
			if found {
				return errFound
			}
			return err
		}},
}

func TestKomercijalnoEotpGolden(t *testing.T) {
	runGolden(t, "komercijalno_eotpremnica.golden", komercijalnoEotpCases)
}

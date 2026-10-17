package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// komercijalnoNalogCases are the statements of the nalozi za isporuku of the komercijalno module.
var komercijalnoNalogCases = []goldenCase{
	{name: "nalog not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomNalog(ctx, 46)
			if found {
				return errFound
			}
			return err
		}},
	{name: "dodatno of the nalog", respond: answer(0, int64(45), 7, time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), 2, int64(4), "Banka", 12, 31),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogDodatno(ctx, 46)
			return err
		}},
	{name: "dodatno of a nalog not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogDodatno(ctx, 46)
			return err
		}},
	{name: "save the profaktura and the banka of the nalog", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomNalogDodatnoSacuvaj(ctx, tx, 46, 45, 2)
		}},
	{name: "banka za uplatu", respond: answer(0, 2),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomNalogBanka(ctx, 4)
			return err
		}},
	{name: "copy the stavke of the profaktura", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomNalogKopirajStavke(ctx, tx, 45, 46, 350, 3, sql.NullTime{Time: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Valid: true}, 2)
		}},
	{name: "mark the profaktura of the nalog", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomNalogProfakturaVezi(ctx, tx, 45)
			return err
		}},
	{name: "set the profaktura of a deleted nalog back", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomNalogProfakturaVrati(ctx, tx, 45)
			return err
		}},
	{name: "lots by their age",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogLotoviStarost(ctx, 3, " 1320 ", 1001)
			return err
		}},
	{name: "promet of the kupac",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogFpro(ctx, "2040", "00031")
			return err
		}},
	{name: "neproknjižene fakture of the kupac",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogRobneStavke(ctx, "2040", "00031")
			return err
		}},
	{name: "pregled of the nepotvrđeni of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogPrikaz(ctx, domain.KomNalogPrikazParams{MagaciniID: 3, Doksts: domain.KomNalogNov})
			return err
		}},
	{name: "pregled of the potvrđeni of all the magacini with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogPrikaz(ctx, domain.KomNalogPrikazParams{Doksts: domain.KomNalogPotvrdjen, SviMagacini: true, Tekst: "metal"})
			return err
		}},
	{name: "pregled of the delimično realizovani of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomNalogPrikaz(ctx, domain.KomNalogPrikazParams{MagaciniID: 3, Delimicni: true})
			return err
		}},
}

func TestKomercijalnoNalogGolden(t *testing.T) {
	runGolden(t, "komercijalno_nalog.golden", komercijalnoNalogCases)
}

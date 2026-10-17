package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var komPrijUpis = domain.KomPrijemnicaUpis{RdokID: 35, Vrd: 198, Dokum: 4, Datum: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), MagaciniID: 3,
	Mag: 1, VrdokID: 25}

var komPrijStavkaUpis = domain.KomPrijStavkaUpis{RdokID: 35, Rbr: 2, Sifra: 1001, RsifID: 44, Po: 2, Konto: "1010", Mag: 1, MagaciniID: 3,
	TrazKolic: 3, Kolic: 0, Otk: "002", Fcena: 100, Cena: 95, Iznos: 285, Rabat: 5, Vrd: 198, Dokum: 4,
	Dadok: sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true}}

// komercijalnoPrijCases are the statements of the prijemnice materijala of the komercijalno module.
var komercijalnoPrijCases = []goldenCase{
	{name: "vrste of the prijemnice of a magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijVrste(ctx, 3)
			return err
		}},
	{name: "prijemnica not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomPrijemnica(ctx, 35)
			if found {
				return errFound
			}
			return err
		}},
	{name: "add a prijemnica of a porudžbenica", respond: answer(0, int64(35)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomPrijDodaj(ctx, tx, komPrijUpis)
			return err
		}},
	{name: "change the datum of a prijemnica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijIzmeni(ctx, tx, komPrijUpis)
		}},
	{name: "change a prijemnica of another period", inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijIzmeni(ctx, tx, komPrijUpis)
		}},
	{name: "confirm a prijemnica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomPrijPotvrdi(ctx, tx, 35, domain.KomZahtevPotvrdjen)
			return err
		}},
	{name: "delete a prijemnica with its stavke and pakovanja", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijBrisi(ctx, tx, 35)
		}},
	{name: "the potvrđena porudžbenica of a prijemnica", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomPrijVeza(ctx, 25)
			return err
		}},
	{name: "stavke of a prijemnica",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijStavke(ctx, 35)
			return err
		}},
	{name: "add a stavka of a prijemnica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijStavkaDodaj(ctx, tx, komPrijStavkaUpis)
		}},
	{name: "change a stavka of a prijemnica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijStavkaIzmeni(ctx, tx, domain.KomPrijStavkaIzmena{RdokID: 35, RproID: 51, Konto: "1010", TrazKolic: 3, Kolic: -2.5,
				Serija: "S-7", Otk: "003"})
		}},
	{name: "delete a stavka of a prijemnica with its pakovanje", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijStavkaBrisi(ctx, tx, 35, 51, 2)
		}},
	{name: "delete all the stavke of a prijemnica", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijStavkeBrisi(ctx, tx, 35)
		}},
	{name: "a stavka of the porudžbenica to take", respond: answer(0, int64(41), 1001, 2, int64(44), 3.0, 100.0, 5.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, found, err := kom(r).KomPrijIzvor(ctx, tx, 25, 41)
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "the artikal is in the prijemnica", respond: answer(0, 1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomPrijImaArtikal(ctx, tx, 35, 44)
			return err
		}},
	{name: "artikal of any tip", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomPrijArtikal(ctx, 1001)
			return err
		}},
	{name: "base of the OTK of an artikal", respond: answer(0, 4),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijOtkBroj(ctx, nil, 198, 1001)
			return err
		}},
	{name: "konta of a magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijKonta(ctx, 3)
			return err
		}},
	{name: "pakovanja of a prijemnica",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijPakovanja(ctx, 35)
			return err
		}},
	{name: "change the pakovanje of a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijPakovanjeSnimi(ctx, tx, 35, 51, 25, 4)
		}},
	{name: "a new pakovanje of a stavka", respond: answer(0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPrijPakovanjeSnimi(ctx, tx, 35, 51, 25, 4)
		}},
	{name: "pregled of the nepotvrđene of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijPrikaz(ctx, domain.KomPrijPrikazParams{MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen})
			return err
		}},
	{name: "pregled of the potvrđene of all the magacini with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPrijPrikaz(ctx, domain.KomPrijPrikazParams{Doksts: domain.KomZahtevPotvrdjen, SviMagacini: true, Tekst: "metal"})
			return err
		}},
}

func TestKomercijalnoPrijGolden(t *testing.T) {
	runGolden(t, "komercijalno_prijemnica.golden", komercijalnoPrijCases)
}

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var komPonudaUpis = domain.KomPonudaUpis{RdokID: 15, Vrd: 310, Dokum: 4, Dadok: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Rok: 10,
	MagaciniID: 3, Mag: 1, KomID: sql.NullInt64{Int64: 5, Valid: true}, Orgjed: sql.NullInt64{Int64: 2, Valid: true}, Foot: "Hitno",
	SifVal: 941, Kurs: 1, VaziZa: 1, VezniID: "101;102;"}

var komPonudaStavkaUpis = domain.KomPonudaStavkaUpis{RdokID: 15, Rbr: 3, Sifra: 1001, RsifID: 44, Po: 2, Kolicina: 3, Mag: 1, MagaciniID: 3,
	Vrd: 310, Dokum: 4, Dadok: sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true}, RproID1: 21}

// komercijalnoPonudaCases are the statements of the zahtevi za ponudu of the komercijalno module.
var komercijalnoPonudaCases = []goldenCase{
	{name: "vrste of the zahtevi za ponudu",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPonudaVrste(ctx)
			return err
		}},
	{name: "zahtev za ponudu with its texts", respond: answer(0, int64(15), 310, 4, nil, 10, 3, 1, int64(5), 7, "Ana Anić", int64(2), "1 - OJ",
		nil, "", "Hitno", "10", "ana", "N", "", 941, "RSD", 1.0, 1, "101;102;"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomPonuda(ctx, 15)
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "zahtev za ponudu not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomPonuda(ctx, 15)
			if found {
				return errFound
			}
			return err
		}},
	{name: "add a zahtev za ponudu", respond: answer(0, int64(15)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomPonudaDodaj(ctx, tx, komPonudaUpis)
			return err
		}},
	{name: "change a zahtev za ponudu", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaIzmeni(ctx, tx, komPonudaUpis)
		}},
	{name: "change a zahtev za ponudu of another period", inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaIzmeni(ctx, tx, komPonudaUpis)
		}},
	{name: "confirm a zahtev za ponudu for a dobavljač", respond: answer(1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			ok, err := kom(r).KomPonudaPotvrdi(ctx, 15, "4350", "00012")
			if err == nil && !ok {
				return ErrNotFound
			}
			return err
		}},
	{name: "delete a zahtev za ponudu: its sources slobodne again", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaBrisi(ctx, tx, 15)
		}},
	{name: "slobodne stavke of the potvrđeni zahtevi za nabavku of a magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPonudaIzvori(ctx, 3)
			return err
		}},
	{name: "a stavka of a zahtev za nabavku to take", respond: answer(0, int64(21), 1001, 2, int64(44), 3.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, found, err := kom(r).KomPonudaIzvor(ctx, tx, 21)
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "a stavka of a zahtev za nabavku already taken", respond: noRows, inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, found, err := kom(r).KomPonudaIzvor(ctx, tx, 21)
			if found {
				return errFound
			}
			return err
		}},
	{name: "reserve a stavka of a zahtev za nabavku", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaIzvorRezervisi(ctx, tx, 21)
		}},
	{name: "add a stavka of a zahtev za ponudu", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaStavkaDodaj(ctx, tx, komPonudaStavkaUpis)
		}},
	{name: "change the količina of a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaStavkaKolicina(ctx, tx, 15, 31, 4.5)
		}},
	{name: "delete a stavka: its source slobodna again", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPonudaStavkaBrisi(ctx, tx, 15, 31, 2)
		}},
	{name: "dobavljači of the ids",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomDobavljaci(ctx, []int64{101, 102})
			return err
		}},
	{name: "dobavljači without ids", wantErr: nil,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			if _, err := kom(r).KomDobavljaci(ctx, nil); err == nil {
				return errFound
			}
			return nil
		}},
	{name: "pregled of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPonudePrikaz(ctx, domain.KomPonudaPrikazParams{Vrd: 310, MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen})
			return err
		}},
	{name: "pregled of all the magacini of the user",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPonudePrikaz(ctx, domain.KomPonudaPrikazParams{Vrd: 310, MagaciniID: 3, Doksts: domain.KomZahtevPotvrdjen,
				SviMagacini: true})
			return err
		}},
	{name: "pregled with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPonudePrikaz(ctx, domain.KomPonudaPrikazParams{Vrd: 310, MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen,
				Tekst: "12"})
			return err
		}},
	{name: "dobavljač of the print", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomPonudaStampaDobavljac(ctx, 101)
			return err
		}},
}

func TestKomercijalnoPonudaGolden(t *testing.T) {
	runGolden(t, "komercijalno_ponuda.golden", komercijalnoPonudaCases)
}

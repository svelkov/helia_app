package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// kom is the repository of the komercijalno module over the database of a golden case.
func kom(r *RobnoRepository) *KomercijalnoRepository {
	return &KomercijalnoRepository{RobnoRepository: *r}
}

var komZahtevUpis = domain.KomZahtevUpis{RdokID: 9, Vrd: 410, Dokum: 12, Dadok: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), MagaciniID: 3,
	Mag: 1, KomID: sql.NullInt64{Int64: 5, Valid: true}, Orgjed: sql.NullInt64{Int64: 2, Valid: true}, Foot: "Hitno"}

var komZahtevStavkaUpis = domain.KomZahtevStavkaUpis{RproID: 21, RdokID: 9, Rbr: 2, Sifra: 1001, RsifID: 44, Po: 2, Cena: 12.5, Kolicina: 3,
	Mag: 1, MagaciniID: 3, Vrd: 410, Dokum: 12, Dadok: sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true}}

// komercijalnoCases are the statements of the zahtevi za nabavku of the komercijalno module.
var komercijalnoCases = []goldenCase{
	{name: "vrste of the zahtevi za nabavku",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahtevVrste(ctx)
			return err
		}},
	{name: "next broj of the vrsta", respond: answer(0, 13),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahtevNoviBroj(ctx, nil, 410)
			return err
		}},
	{name: "zahtev with its texts", respond: answer(0, int64(9), 410, "Zahtev", 12, nil, 3, 1, "Magacin sirovina", int64(5), 7, "Ana Anić",
		int64(2), "1 - OJ", nil, "", "Hitno", "10", "ana", "N", ""),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomZahtev(ctx, 9)
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "zahtev not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomZahtev(ctx, 9)
			if found {
				return errFound
			}
			return err
		}},
	{name: "add a zahtev", respond: answer(0, int64(9)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomZahtevDodaj(ctx, tx, komZahtevUpis)
			return err
		}},
	{name: "change a zahtev", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomZahtevIzmeni(ctx, tx, komZahtevUpis)
		}},
	{name: "change a zahtev of another period", respond: answer(0), inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomZahtevIzmeni(ctx, tx, komZahtevUpis)
		}},
	{name: "confirm a zahtev", respond: answer(1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahtevPotvrdi(ctx, 9)
			return err
		}},
	{name: "delete a zahtev with its stavke", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomZahtevBrisi(ctx, tx, 9)
		}},
	{name: "stavke of a zahtev",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahtevStavke(ctx, 9)
			return err
		}},
	{name: "stavka of a zahtev not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomZahtevStavka(ctx, 9, 21)
			if found {
				return errFound
			}
			return err
		}},
	{name: "next redni broj of a stavka", respond: answer(0, 3), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomZahtevNoviRbr(ctx, tx, 9)
			return err
		}},
	{name: "add a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomZahtevStavkaDodaj(ctx, tx, komZahtevStavkaUpis)
		}},
	{name: "change a stavka", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomZahtevStavkaIzmeni(ctx, tx, komZahtevStavkaUpis)
		}},
	{name: "delete a stavka and move the redni brojevi after it", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomZahtevStavkaBrisi(ctx, tx, 9, 21, 2)
		}},
	{name: "artikal with its nabavna cena", respond: answer(0, int64(44), 1001, "Vijak", "kom", 2, 12.5),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomArtikal(ctx, 1001)
			return err
		}},
	{name: "komercijalista by šifra", respond: answer(0, int64(5), 7, "Ana Anić"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomKomercijalista(ctx, 7)
			return err
		}},
	{name: "komercijalista of the user", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomKomercijalistaKorisnika(ctx, 4)
			return err
		}},
	{name: "magacin", respond: answer(0, 1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomMagacin(ctx, 3)
			return err
		}},
	{name: "pregled of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahteviPrikaz(ctx, domain.KomZahtevPrikazParams{Vrd: 410, MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen})
			return err
		}},
	{name: "pregled of all the magacini of the user",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahteviPrikaz(ctx, domain.KomZahtevPrikazParams{Vrd: 410, MagaciniID: 3, Doksts: domain.KomZahtevPotvrdjen,
				SviMagacini: true})
			return err
		}},
	{name: "pregled with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomZahteviPrikaz(ctx, domain.KomZahtevPrikazParams{Vrd: 410, MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen,
				Tekst: " hitno "})
			return err
		}},
	{name: "header of the print", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomZahtevStampaZaglavlje(ctx, 9)
			return err
		}},
	{name: "firma of the print", respond: noRows, wantErr: sql.ErrNoRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomFirma(ctx, "Firma")
			return err
		}},
}

func TestKomercijalnoGolden(t *testing.T) {
	runGolden(t, "komercijalno.golden", komercijalnoCases)
}

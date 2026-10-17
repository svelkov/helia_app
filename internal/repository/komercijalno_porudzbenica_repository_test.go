package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

var komPorUpis = domain.KomPorudzbenicaUpis{RdokID: 25, Vrd: 320, Dokum: 6, Dadok: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Rok: 15,
	MagaciniID: 3, Mag: 1, Fkto: "4330", Fana: "00012", KomID: sql.NullInt64{Int64: 5, Valid: true}, Orgjed: sql.NullInt64{Int64: 2, Valid: true},
	MestoTroska: sql.NullInt64{Int64: 5, Valid: true}, Foot: "Hitno", SifVal: 978, Kurs: 117.25, VaziZa: 1, VrdokID: 15, Otp: "Kamion",
	Pla: "VIRMANOM", Brugov: "U-12", Nacpak: "Palete", Posuslnab: "Bez avansa", Pratdok: "Atest"}

var komPorStavkaUpis = domain.KomPorStavkaUpis{RdokID: 25, Rbr: 2, Sifra: 1001, RsifID: 44, Po: 2, Kolicina: 3, Cena: 1172.5, CenaVal: 10,
	Rabat: 5, Mag: 1, MagaciniID: 3, Vrd: 320, Dokum: 6, Dadok: sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true},
	RproID1: 31}

// komercijalnoPorCases are the statements of the porudžbenice dobavljačima of the komercijalno module.
var komercijalnoPorCases = []goldenCase{
	{name: "vrste of the porudžbenice",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorVrste(ctx)
			return err
		}},
	{name: "next broj of the vrsta in the magacin", respond: answer(0, 7),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorNoviBroj(ctx, nil, 320, 1)
			return err
		}},
	{name: "porudžbenica not found", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomPorudzbenica(ctx, 25)
			if found {
				return errFound
			}
			return err
		}},
	{name: "add a porudžbenica", respond: answer(0, int64(25)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := kom(r).KomPorDodaj(ctx, tx, komPorUpis)
			return err
		}},
	{name: "change a porudžbenica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPorIzmeni(ctx, tx, komPorUpis)
		}},
	{name: "change a porudžbenica of another period", inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPorIzmeni(ctx, tx, komPorUpis)
		}},
	{name: "confirm a porudžbenica", respond: answer(1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorPotvrdi(ctx, 25)
			return err
		}},
	{name: "delete a porudžbenica: its sources slobodne again", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPorBrisi(ctx, tx, 25)
		}},
	{name: "the potvrđen zahtev za ponudu of a porudžbenica", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := kom(r).KomPorVeza(ctx, 15)
			return err
		}},
	{name: "dobavljač of the konto and šifra", respond: answer(0, "Metal", "Glavna 1", "11000", "Beograd", "100", 1),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, found, err := kom(r).KomPorDobavljac(ctx, " 4330", "00012 ")
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "analitika of the dobavljač", respond: noRows,
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorIDFkpl(ctx, "4330", "00012")
			return err
		}},
	{name: "stavke of a porudžbenica",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorStavke(ctx, 25)
			return err
		}},
	{name: "add a stavka of a porudžbenica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPorStavkaDodaj(ctx, tx, komPorStavkaUpis)
		}},
	{name: "change a stavka of a porudžbenica", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPorStavkaIzmeni(ctx, tx, 25, 41, 4, 1172.5, 10, 5)
		}},
	{name: "delete all the stavke: their sources slobodne again", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return kom(r).KomPorVratiSve(ctx, tx, 25)
		}},
	{name: "slobodne stavke of the zahtev za ponudu",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorIzvori(ctx, 15)
			return err
		}},
	{name: "a stavka of the zahtev za ponudu to take", respond: answer(0, int64(31), 1001, 2, int64(44), 3.0, 0.0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, found, err := kom(r).KomPorIzvor(ctx, tx, 15, 31)
			if !found {
				return ErrNotFound
			}
			return err
		}},
	{name: "pregled of one magacin",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorPrikaz(ctx, domain.KomPonudaPrikazParams{Vrd: 320, MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen})
			return err
		}},
	{name: "pregled with the search",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorPrikaz(ctx, domain.KomPonudaPrikazParams{Vrd: 320, MagaciniID: 3, Doksts: domain.KomZahtevNepotvrdjen, Tekst: "metal"})
			return err
		}},
	{name: "pregled of all the magacini of the user",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := kom(r).KomPorPrikaz(ctx, domain.KomPonudaPrikazParams{Vrd: 320, Doksts: domain.KomZahtevPotvrdjen, SviMagacini: true})
			return err
		}},
}

func TestKomercijalnoPorGolden(t *testing.T) {
	runGolden(t, "komercijalno_porudzbenica.golden", komercijalnoPorCases)
}

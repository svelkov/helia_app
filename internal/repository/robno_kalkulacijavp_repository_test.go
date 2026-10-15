package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// robnoKalkulacijaCases are the statements of the kalkulacije veleprodaje (the group PLT).
var robnoKalkulacijaCases = []goldenCase{
	{name: "rvr konto of the dobavljači", respond: answer(0, "4350"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RvrFdob(ctx)
			return err
		}},
	{name: "rdok columns of the kalkulacija",
		respond: answer(0, "8a.2", int64(5), "8a.2 - Nabavka", true, int64(77), int64(12), time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RdokKalkulacijaPodaci(ctx, 41)
			return err
		}},
	{name: "edok narudžbenica", respond: answer(0, int64(12), time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.EdokNarudzbenica(ctx, 77)
			return err
		}},
	postojiCase("rdok izvorni dokument of the dobavljač", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.RdokIzvorniDobavljacaPostoji(ctx, "4350", "15", 110, "R-12/26", 41)
	}),
	{name: "popdv polje of the ulazni računi", respond: answer(0, int64(3), "U", "1,2"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.PopdvPoljeUlaz(ctx, "8a.2")
			return err
		}},
	{name: "rpor row on the datum without a period", respond: answer(0, int64(9)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RporIDNaDan(ctx, 2, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
			return err
		}},
	{name: "fsepp polje", respond: answer(0, "8a.2"),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.FseppPolje(ctx, 5)
			return err
		}},
	{name: "rsif artikal of the kalkulacija by šifra", respond: answer(0, int64(31), int64(15), "Vijak", "kom", "1320", int64(2), 20.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RsifArtikalKalkulacije(ctx, false, "15", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
			return err
		}},
	{name: "rsif artikal of the kalkulacija by barkod", respond: answer(0, int64(31), int64(15), "Vijak", "kom", "", int64(2), 20.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RsifArtikalKalkulacije(ctx, true, "8600000000001", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
			return err
		}},
	// The plaćen porez and avans.
	{name: "rppo of the document",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RppoDokumenta(ctx, 41)
			return err
		}},
	{name: "rppo of the zavisni trošak",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RppoTroska(ctx, 8)
			return err
		}},
	{name: "appo of the document",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.AppoDokumenta(ctx, 41)
			return err
		}},
	{name: "rppo add of the document", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RppoDodaj(ctx, tx, KalPorezDokument{Vrd: 110, Dokum: 12, Dadok: sql.NullTime{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				Konto: "4350", Sifra: "15", Rac: "R-12/26", RdokID: 41}, domain.RobKalPorezUnos{Osnovica: 1000, Porez: 200, Stopa: 20, Po: 2, RporID: 9})
		}},
	{name: "rppo add of the zavisni trošak", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RppoDodaj(ctx, tx, KalPorezDokument{Vrd: 110, Dokum: 12, Konto: "4350", Sifra: "16", Rac: "12", RztrID: 8},
				domain.RobKalPorezUnos{Osnovica: 100, Porez: 20, Stopa: 20, Po: 2, RporID: 9})
		}},
	{name: "appo add", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AppoDodaj(ctx, tx, KalPorezDokument{Vrd: 110, Dokum: 12, Konto: "4350", Sifra: "15", RdokID: 41},
				domain.RobKalPorezUnos{Osnovica: 500, Porez: 100, Stopa: 20, Po: 2, RporID: 9, Rac: "AV-3",
					Datum: sql.NullTime{Time: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Valid: true}})
		}},
	{name: "rppo delete of the document", respond: answer(2), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RppoDokumentaBrisi(ctx, tx, 41)
		}},
	{name: "rppo delete of the zavisni trošak", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RppoTroskaBrisi(ctx, tx, 8)
		}},
	{name: "rppo delete of the zavisni troškovi of the document", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RppoTroskovaDokumentaBrisi(ctx, tx, 41)
		}},
	{name: "appo delete of the document", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.AppoDokumentaBrisi(ctx, tx, 41)
		}},
	{name: "rztr delete of the document", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RztrDokumentaBrisi(ctx, tx, 41)
		}},
	{name: "kepu delete of the document", respond: answer(0), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.KepuDokumentaBrisi(ctx, tx, 41)
		}},
	// The zavisni troškovi.
	{name: "rztr of the document",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RztrDokumenta(ctx, 41)
			return err
		}},
	{name: "rztr trošak of the document",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RztrTrosak(ctx, 8, 41)
			return err
		}},
	{name: "rztr add", respond: answer(1, int64(8)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RztrDodaj(ctx, tx, 41, 110, 12, RztrUnos{Fkto: "4350", Fana: "16", Dokdob: "T-5", Rok: 15, Opis: "Prevoz",
				Polje: "8a.2", Iznfak: 1000, Iznos: 50, Tkonto: "5120"})
			return err
		}},
	{name: "rztr update", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RztrIzmeni(ctx, tx, 8, 41, RztrUnos{Fkto: "4350", Fana: "16", Dokdob: "T-5", Flguvozpdv: true, Iznfak: 1000, Iznos: 60})
		}},
	{name: "rztr update of another document", respond: answer(0), inTx: true, wantErr: ErrNotFound,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RztrIzmeni(ctx, tx, 8, 42, RztrUnos{})
		}},
	{name: "rztr delete", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.RztrBrisi(ctx, tx, 8, 41)
		}},
	{name: "rztr procenat", respond: answer(0, 150.0, 1000.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RztrProcenat(ctx, 41)
			return err
		}},
	// The stanje and the stavke.
	{name: "rsta cene of the kalkulacija",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, err := r.RstaKalkulacijaCene(ctx, 3, " 1320 ", 15)
			return err
		}},
	{name: "rpro stavka of the kalkulacija for update",
		respond: answer(0, int64(77), int64(1), int64(15), "1320", 2.0, 240.0, 120.0, 100.0, 100.0, 5.0, 95.0, 110.0, 132.0, 26.3, 15.8,
			0.0, 0.0, "-", "-", "", "", 0.0, "", "kom", int64(2)), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			_, err := r.RproKalkulacija(ctx, tx, 77, 41)
			return err
		}},
	{name: "rpro stavka of the kalkulacija",
		respond: answer(0, int64(77), int64(1), int64(15), "1320", 2.0, 240.0, 120.0, 100.0, 100.0, 5.0, 95.0, 110.0, 132.0, 26.3, 15.8,
			0.0, 0.0, "-", "-", "", "", 0.0, "", "kom", int64(2)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproKalkulacija(ctx, nil, 77, 41)
			return err
		}},
	{name: "rcene maloprodajne cene", respond: answer(0, 125.0, 150.0),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, _, _, err := r.RceneMaloprodajne(ctx, 15)
			return err
		}},
	{name: "rpro zalihe after the deletion",
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			_, err := r.RproZaliheBrisanja(ctx, 41)
			return err
		}},
	{name: "drsta veleprodajna cena of the lot", respond: answer(1), inTx: true,
		run: func(ctx context.Context, r *RobnoRepository, tx db.Transaction) error {
			return r.DrstaLotPostaviVpcena(ctx, tx, "1320", 15, 1, "OTK-1", "S1", 46000, 120)
		}},
}

func TestRobnoKalkulacijaGolden(t *testing.T) {
	runGolden(t, "robno_kalkulacijavp.golden", robnoKalkulacijaCases)
}

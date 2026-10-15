package repository

import (
	"context"
	"testing"

	"helia/internal/infrastructure/db"
)

// postojiCase is a golden case of a check that answers one count.
func postojiCase(name string, check func(ctx context.Context, r *RobnoRepository) (bool, error)) goldenCase {
	return goldenCase{name: name, respond: answer(0, int64(1)),
		run: func(ctx context.Context, r *RobnoRepository, _ db.Transaction) error {
			ok, err := check(ctx, r)
			if err == nil && !ok {
				return errNotFoundInCase
			}
			return err
		}}
}

type notFoundInCase struct{}

func (notFoundInCase) Error() string { return "count 1 answered as not found" }

var errNotFoundInCase error = notFoundInCase{}

var robnoPostojiCases = []goldenCase{
	postojiCase("rsif", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.RsifPostoji(ctx, 15) }),
	postojiCase("fkpl analitika", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FkplAnalitikaPostoji(ctx, "2040", "15")
	}),
	postojiCase("fkpl analitički konto", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FkplAnalitickiKontoPostoji(ctx, "2040", "15")
	}),
	postojiCase("fkpl analitika trimmed", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FkplAnalitikaTrimPostoji(ctx, "2040", "")
	}),
	postojiCase("fkpl konto of the partner", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FkplKontoPartneraPostoji(ctx, "2040", 31)
	}),
	postojiCase("fkpl sintetika", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FkplSintetikaPostoji(ctx, "1320")
	}),
	postojiCase("avans closed", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.AvansZatvoren(ctx, 41) }),
	postojiCase("rdok broj", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.RdokBrojPostoji(ctx, 3, 89, 12) }),
	postojiCase("rdok broj of another nalog", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.RdokBrojDrugogNalogaPostoji(ctx, 3, 89, 12, 9, 41)
	}),
	postojiCase("rdok avansni of another kupac", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.RdokAvansniDrugogKupca(ctx, []int{4, 7}, "2040", "15")
	}),
	postojiCase("orgjed", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.OrgjedPostoji(ctx, 2) }),
	postojiCase("mestotr", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.MestotrPostoji(ctx, 5) }),
	postojiCase("mestotr of the orgjed", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.MestotrUOrgjed(ctx, 5, 2) }),
	postojiCase("komercijalista", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.KomercijalistaPostoji(ctx, 6) }),
	postojiCase("bnkizv", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.BnkizvPostoji(ctx, 8) }),
	postojiCase("fisp of the partner", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FispPostoji(ctx, 4, "2040", "15")
	}),
	postojiCase("porkat without a period", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.PorkatPostoji(ctx, 3) }),
	postojiCase("rpor without a period", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.RporPostoji(ctx, 4) }),
	postojiCase("magacin", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.MagacinPostoji(ctx, 3) }),
	postojiCase("magacin by the nabavna cena", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.MagacinPoNabavnoj(ctx, 3)
	}),
	postojiCase("magkonto", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.MagkontoPostoji(ctx, 3, "1320") }),
	postojiCase("dokvrsta prosečna", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.DokvrstaProsecna(ctx, 89) }),
	postojiCase("dokvrsta avansna of another year", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.DokvrstaAvansnaGodine(ctx, 2025, 172)
	}),
	postojiCase("fvr obveznik PDV", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.FvrObveznikPdv(ctx, "Firma d.o.o.")
	}),
	postojiCase("popdv polje", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.PopdvPoljePostoji(ctx, "3.2") }),
	postojiCase("popdv polje not of the izlazni računi", func(ctx context.Context, r *RobnoRepository) (bool, error) {
		return r.PopdvPoljeNeodgovarajuce(ctx, "3.2")
	}),
	postojiCase("valuta strana", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.ValutaStrana(ctx, 2) }),
	postojiCase("tipknpisma", func(ctx context.Context, r *RobnoRepository) (bool, error) { return r.TipknpismaPostoji(ctx, 4) }),
}

func TestRobnoPostojiGolden(t *testing.T) {
	runGolden(t, "robno_postoji.golden", robnoPostojiCases)
}

func TestRdokAvansniDrugogKupcaWithoutBrojevi(t *testing.T) {
	d, _ := newDrvDB(t, answer(0, int64(0)))
	if _, err := NewRobnoRepository(d).RdokAvansniDrugogKupca(robnoCtx(), nil, "2040", "15"); err == nil {
		t.Fatal("want an error without brojevi")
	}
}

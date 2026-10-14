package robno

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"helia/config"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestZzPorTmp(t *testing.T) {
	data, _ := os.ReadFile("../../../config/config.json")
	var cfg config.Config
	_ = json.Unmarshal(data, &cfg)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		cfg.DBConfig.DBHost, cfg.DBConfig.DBPort, cfg.DBConfig.DBUser, cfg.DBConfig.DBPassword, cfg.DBConfig.DBName, cfg.DBConfig.DBSearchPath)
	d := db.NewPostgresDB(sqlx.MustOpen("postgres", dsn))
	cs := commonsvc.NewCommonService(*repository.NewBaseRepository[domain.Magacini](d, "magacini"), *repository.NewBaseRepository[domain.Tipdok](d, "tipdok"),
		*repository.NewBaseRepository[domain.Dokvrsta](d, "dokvrsta"), *repository.NewBaseRepository[domain.Orgjed](d, "orgjed"), *repository.NewBaseRepository[domain.Mestotr](d, "mestotr"),
		*repository.NewBaseRepository[domain.Rgru](d, "rgru"), *repository.NewBaseRepository[domain.Rpgru](d, "rpgru"), *repository.NewBaseRepository[domain.Valute](d, "valute"),
		*repository.NewBaseRepository[domain.Komercijalisti](d, "komercijalisti"), *repository.NewBaseRepository[domain.Fisp](d, "fisp"), *repository.NewBaseRepository[domain.Banke](d, "banke"),
		*repository.NewBaseRepository[domain.Tipanalitike](d, "tipanalitike"), *repository.NewBaseRepository[domain.Jedmere](d, "jedmere"), *repository.NewBaseRepository[domain.Fvknjrac](d, "fvknjrac"),
		*repository.NewBaseRepository[domain.Fnal](d, "fnal"), *repository.NewBaseRepository[domain.RobnoDokumentaDto](d, "rnal"))
	s := NewRobnoDokumentaService(*repository.NewBaseRepository[domain.RobnoDokumentaDto](d, "rnal"), repository.BaseRepository[domain.RobnoDokumentaTotalsDto]{},
		*repository.NewBaseRepository[domain.Rnal](d, "rnal"), repository.BaseRepository[domain.PrikazUkupneObradeDto]{}, *repository.NewBaseRepository[domain.Tipdok](d, "tipdok"),
		*repository.NewBaseRepository[domain.Dokvrsta](d, "dokvrsta"), *repository.NewBaseRepository[domain.Magacini](d, "magacini"), *repository.NewBaseRepository[domain.Fvr](d, "fvr"), cs,
		repository.BaseRepository[domain.RobnoStampaFakturaRowDto]{}, repository.BaseRepository[domain.RobnoStampaFakturaAvansDto]{}, repository.BaseRepository[domain.RobnoStampaFakturaRataDto]{},
		repository.BaseRepository[domain.RobnoStampaFakturaFirmaDto]{}, repository.BaseRepository[domain.RobnoStampaPopisRowDto]{}, repository.BaseRepository[domain.RobnoStampaDokumentDto]{},
		repository.BaseRepository[domain.RobnoStampaKalkulacijaRowDto]{}, repository.BaseRepository[domain.RobnoStampaFakturaIzvozRowDto]{}, repository.BaseRepository[domain.RobnoStampaKnjiznoPismoRowDto]{},
		repository.BaseRepository[domain.RobnoStampaKnjiznoPismoFakturaDto]{}, repository.BaseRepository[domain.RobnoStampaInterniPrenosRowDto]{}, repository.BaseRepository[domain.RobnoStampaPoreskaStopaDto]{},
		RobnoDokumentaUnosRepos{
			Dokument: *repository.NewBaseRepository[domain.RobniDokumentRow](d, "rdok"), Kupac: *repository.NewBaseRepository[domain.RobKupacDto](d, "fkpl"),
			Avansi: *repository.NewBaseRepository[domain.RobAvansDto](d, "avansfakt"), Rate: *repository.NewBaseRepository[domain.RobRataDto](d, "faktrate"),
			Rpro: *repository.NewBaseRepository[domain.RobRproDto](d, "rpro"), Rcene: *repository.NewBaseRepository[domain.RobRceneDto](d, "rcene"),
			Rsif: *repository.NewBaseRepository[domain.RobRsifDto](d, "rsif"), Prikaz: *repository.NewBaseRepository[domain.RobDokumentPrikazRow](d, "rdok"),
		})
	r := s
	sess := &domain.UserSession{SelectedGod: 2026, SelectedKar: 47, Firma: "T", UserName: "test"}
	ctx := domain.SetSessionInStdContext(context.Background(), sess)

	// The validation of the header and of a stavka (committed data only).
	var oj, fkto, fana string
	_ = d.QueryRowContext(ctx, `select idorgjed::text from orgjed where god = 2026 and kar = 47 order by idorgjed limit 1`).Scan(&oj)
	_ = d.QueryRowContext(ctx, `select konto, sifra from fkpl where god = 2026 and kar = 47 and vkonta = 1 and konto like '43%' order by idfkpl limit 1`).Scan(&fkto, &fana)
	base := domain.RobnoPorParams{RnalID: 66521, MagaciniID: 323, Vrd: "320", BrojDokum: "1", DatumDokum: "2026-10-14",
		DobavljacKonto: fkto, DobavljacSifra: fana, Orgjed: oj, Rok: "15"}
	fmt.Printf("validate ok:      %+v\n", r.ValidateRobPor(ctx, base))
	bad := base
	bad.DatumDokum, bad.DobavljacSifra, bad.Orgjed, bad.BrojDokum, bad.Rok, bad.Komercijalista = "", "zzz", "-", "x", "a", "999999999"
	fmt.Printf("validate bad:     %+v\n", r.ValidateRobPor(ctx, bad))
	var postojeci int64
	_ = d.QueryRowContext(ctx, `select rdokid from rdok where rnalid = 66521 order by rdokid limit 1`).Scan(&postojeci)
	fmt.Printf("stavka bad:       %+v\n", r.ValidateRobPorStavka(ctx, domain.RobPorStavkaParams{RdokID: postojeci, Konto: "9999", Sifra: "0", Kolic: "0"}))
	por, err := r.GetRobPorDokument(ctx, domain.RobnoFaktureDokumentParams{RnalID: 66521, MagaciniID: 323, Vrd: 320})
	fmt.Printf("dokument: err=%v broj=%s datum=%s opis=%q pla=%q fkto=%q konto=%q dobavljac=%q\n", err, por.Header.BrojDokum,
		por.Header.DatumDokum.Format("2006-01-02"), por.Header.VrstaOpis, por.Header.UsloviPlacanja, por.Header.KupacKonto, por.Stavka.Konto, por.Stavka.Dobavljac)

	tx, err := repository.NewBaseRepository[domain.RobRproDto](d, "rpro").BeginTx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	show := func(label, q string, args ...any) {
		var out sql.NullString
		if err := tx.QueryRowContext(ctx, `select string_agg(row_to_json(x)::text, E'\n          ') from (`+q+`) x`, args...).Scan(&out); err != nil {
			fmt.Println(label, "ERR", err)
			return
		}
		fmt.Println(label, out.String)
	}
	// A vrsta of the group POR (posted on the duguje) and a porudžbenica of it in the nalog 66521.
	var dokvrstaID, rdokID int64
	if err := tx.QueryRowContext(ctx, `insert into dokvrsta overriding user value select (jsonb_populate_record(null::dokvrsta, to_jsonb(v) ||
			jsonb_build_object('vrd', 398, 'grpdok', 'POR', 'kodknj', 'D', 'predznak', '+'))).*
		from dokvrsta v where v.dokvrstaid = 14653 returning dokvrstaid`).Scan(&dokvrstaID); err != nil {
		t.Fatal("dokvrsta:", err)
	}
	if err := tx.QueryRowContext(ctx, `insert into rdok overriding user value select (jsonb_populate_record(null::rdok, to_jsonb(r) ||
			jsonb_build_object('vrd', 398, 'dokum', 1, 'iznos', 0, 'brst', 0, 'fkto', $1::text, 'fana', $2::text, 'vrdokid', null))).*
		from rdok r where r.rnalid = 66521 order by r.rdokid limit 1 returning rdokid`, fkto, fana).Scan(&rdokID); err != nil {
		t.Fatal("rdok:", err)
	}
	ctxQ := `select d.rdokid, coalesce(d.rnalid, 0)::bigint, coalesce(d.tipdok, ''), coalesce(d.nalog, 0), d.danal, coalesce(d.vrd, 0),
		coalesce(v.dokvrstaid, 0)::bigint, coalesce(d.dokum, 0), d.dadok, coalesce(d.fkto, ''), coalesce(d.fana, ''), coalesce(d.magaciniid, 0),
		coalesce(d.mag, 0), coalesce(m.tipzal, 0), coalesce(v.predznak, ''), coalesce(v.kodknj, ''), coalesce(m.nacvodzal, 0), coalesce(m.tipmag, '')
		from rdok d left join magacini m on m.magaciniid = d.magaciniid left join dokvrsta v on v.god = d.god and v.kar = d.kar and v.vrd = d.vrd
		where d.rdokid = $1`
	var k domain.RobStavkeKontekst
	if err := tx.QueryRowContext(ctx, ctxQ, rdokID).Scan(&k.RdokID, &k.RnalID, &k.Tipdok, &k.Nalog, &k.Danal, &k.Vrd, &k.DokvrstaID, &k.Dokum, &k.Dadok,
		&k.Fkto, &k.Fana, &k.MagaciniID, &k.Mag, &k.Tipzal, &k.Predznak, &k.Kodknj, &k.Nacvodzal, &k.Tipmag); err != nil {
		t.Fatal("kontekst:", err)
	}
	fmt.Printf("kontekst: rdok %d vrd %d dokvrsta %d kodknj %q fkto %s/%s\n", k.RdokID, k.Vrd, k.DokvrstaID, k.Kodknj, k.Fkto, k.Fana)
	var konto string
	var sifra int
	_ = tx.QueryRowContext(ctx, `select konto, sifra from rsta where god = 2026 and kar = 47 and magaciniid = 323 order by rstaid limit 1`).Scan(&konto, &sifra)
	rsta := func(label string) {
		show(label, `select konto, sifra, ulaz, izlaz, dug, pot, brst from rsta where god = 2026 and kar = 47 and magaciniid = 323 and konto = $1 and sifra = $2`, konto, sifra)
	}
	stanje := func(label string) {
		show(label+" rdok:", `select iznos, brst from rdok where rdokid = $1`, rdokID)
		show(label+" rnal:", `select brst, dug, pot from rnal where rnalid = 66521`)
		rsta(label + " rsta:")
	}
	stanje("before ")
	p := domain.RobPorStavkaParams{RdokID: rdokID, Konto: konto, Sifra: fmt.Sprint(sifra), Kolic: "4", Ncena: "12.5", Tarifa: "2"}
	rproID, err := r.robPorSnimiStavku(ctx, tx, k, p)
	if err != nil {
		t.Fatal("stavka:", err)
	}
	show("rpro:   ", `select rproid, vrd, rbr, brst, konto, sifra, kolic, cena, ncena, iznos, po, fkto, fana, naz1, jm from rpro where rproid = $1`, rproID)
	stanje("added  ")
	p.RproID, p.Kolic, p.Ncena = rproID, "6", "10"
	if _, err := r.robPorSnimiStavku(ctx, tx, k, p); err != nil {
		t.Fatal("izmena:", err)
	}
	show("rpro:   ", `select rproid, rbr, kolic, ncena, iznos from rpro where rproid = $1`, rproID)
	stanje("changed")
	if err := r.robPorBrisiStavku(ctx, tx, k, rproID); err != nil {
		t.Fatal("brisi:", err)
	}
	show("rpro count:", `select count(*) from rpro where rdokid = $1`, rdokID)
	stanje("deleted")
}

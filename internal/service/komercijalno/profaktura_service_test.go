package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

// fakeProRepo is an in-memory proRepository: the profakture by rdokid with their stavke, the artikli and the tipdok; it
// records what was written.
type fakeProRepo struct {
	profakture map[int64]domain.KomProfakturaRow
	stavke     map[int64][]domain.KomProStavkaRow
	artikli    map[int]domain.KomPrijArtikal
	tipdok     []domain.ComboItem
	nalogVrste []domain.ComboItem
	tipmag     string
	tipcene    int
	planska    float64
	noviRbr    int

	dodat      []domain.KomProfakturaUpis
	izmenjen   []domain.KomProfakturaUpis
	ukupno     [][4]float64
	potvrdjen  []int64
	stavkaNova []domain.KomProStavkaUpis
	stavkaIzm  []domain.KomProStavkaIzmena
	stavkaBris [][3]int64
	koraci     []string
	commits    int

	// lotovi and rste are the ids of the rows of drsta (by konto|sifra|mag|otk|serija|roktr) and of rsta (by
	// magaciniid|konto|sifra); rez records the changes of the stanje.
	lotovi map[string]int64
	rste   map[string]int64
	rez    []string
}

func (f *fakeProRepo) DrstaIDZaIzmenu(_ context.Context, _ db.Transaction, konto string, sifra, mag int, otk, serija string, roktr int64) (int64, error) {
	if id, ok := f.lotovi[fmt.Sprintf("%s|%d|%d|%s|%s|%d", konto, sifra, mag, otk, serija, roktr)]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("zaliha lota: %w", sql.ErrNoRows)
}
func (f *fakeProRepo) RstaIDZaIzmenu(_ context.Context, _ db.Transaction, magaciniID int, konto string, sifra int) (int64, error) {
	if id, ok := f.rste[fmt.Sprintf("%d|%s|%d", magaciniID, konto, sifra)]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("stanje: %w", sql.ErrNoRows)
}
func (f *fakeProRepo) StanjeDodaj(_ context.Context, _ db.Transaction, tabela repository.StanjeTabela, v map[string]any) error {
	f.rez = append(f.rez, fmt.Sprintf("novi %s %v/%v %v/%v/%v rsif %v rez %v", tabela, v["konto"], v["sifra"], v["otk"], v["serija"], v["roktr"],
		v["rsifid"], v["rez"]))
	return nil
}
func (f *fakeProRepo) StanjeIzmeni(_ context.Context, _ db.Transaction, tabela repository.StanjeTabela, id int64, izmene []repository.Izmena) error {
	for _, iz := range izmene {
		f.rez = append(f.rez, fmt.Sprintf("%s %d %s %+v nn=%v", tabela, id, iz.Kolona, iz.Vrednost, iz.NeNegativno))
	}
	return nil
}

func (f *fakeProRepo) BeginTx() (db.Transaction, error) {
	return &db.MockTransaction{CommitFunc: func() error { f.commits++; return nil }, RollbackFunc: func() error { return nil }}, nil
}
func (f *fakeProRepo) KomProVrste(context.Context) ([]domain.ComboItem, error) {
	return []domain.ComboItem{{Key: "360", Value: "360 - Predračun"}}, nil
}
func (f *fakeProRepo) KomProNalogVrste(context.Context) ([]domain.ComboItem, error) {
	return f.nalogVrste, nil
}
func (f *fakeProRepo) KomProTipdok(context.Context, int) ([]domain.ComboItem, error) {
	return f.tipdok, nil
}
func (f *fakeProRepo) KomProfaktura(_ context.Context, id int64) (domain.KomProfakturaRow, bool, error) {
	p, ok := f.profakture[id]
	return p, ok, nil
}
func (f *fakeProRepo) KomProMagacin(_ context.Context, id int) (string, int, int, bool, error) {
	return f.tipmag, 2, f.tipcene, id > 0, nil
}
func (f *fakeProRepo) KomProDodaj(_ context.Context, _ db.Transaction, p domain.KomProfakturaUpis) (int64, error) {
	f.dodat = append(f.dodat, p)
	return 99, nil
}
func (f *fakeProRepo) KomProIzmeni(_ context.Context, _ db.Transaction, p domain.KomProfakturaUpis) error {
	f.izmenjen = append(f.izmenjen, p)
	return nil
}
func (f *fakeProRepo) KomProUkupno(_ context.Context, _ db.Transaction, _ int64, iznos, vrugrabat, vkase, netwght float64) error {
	f.ukupno = append(f.ukupno, [4]float64{iznos, vrugrabat, vkase, netwght})
	return nil
}
func (f *fakeProRepo) KomProPotvrdi(_ context.Context, id int64) (bool, error) {
	f.potvrdjen = append(f.potvrdjen, id)
	return true, nil
}
func (f *fakeProRepo) KomProKupac(_ context.Context, konto, sifra string) (domain.KomProKupac, bool, error) {
	return domain.KomProKupac{KomDobavljacPodaci: domain.KomDobavljacPodaci{Naziv: "Kupac", TipPdv: 1}, Ugrabat: 3, Rok: 15},
		konto == "2040" && sifra == "00031", nil
}
func (f *fakeProRepo) KomProMestoIsporuke(context.Context, string, string, int) (domain.KomProMestoIsporuke, bool, error) {
	return domain.KomProMestoIsporuke{Naziv: "Magacin kupca", Adresa: "Put 1", Pobro: "21000", Mesto: "Novi Sad", Gln: "860"}, true, nil
}
func (f *fakeProRepo) KomProCenovnik(context.Context, int64) (string, string, int, sql.NullTime, bool, error) {
	return "VP", "R", 3, sql.NullTime{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Valid: true}, true, nil
}
func (f *fakeProRepo) KomProStavke(_ context.Context, id int64) ([]domain.KomProStavkaRow, error) {
	return f.stavke[id], nil
}
func (f *fakeProRepo) KomProStavkaDodaj(_ context.Context, _ db.Transaction, st domain.KomProStavkaUpis) error {
	f.stavkaNova = append(f.stavkaNova, st)
	return nil
}
func (f *fakeProRepo) KomProStavkaIzmeni(_ context.Context, _ db.Transaction, st domain.KomProStavkaIzmena) error {
	f.stavkaIzm = append(f.stavkaIzm, st)
	return nil
}
func (f *fakeProRepo) KomProLotovi(context.Context, int, int64, bool) ([]domain.KomProLot, error) {
	return []domain.KomProLot{{RstaID: 7, Otk: "001", Serija: "S1", Roktr: 82000, Stanje: 5}}, nil
}
func (f *fakeProRepo) KomProCenaArtikla(context.Context, int) (float64, error) { return 120, nil }
func (f *fakeProRepo) KomProCenaCenovnika(context.Context, int64, int) (float64, float64, bool, error) {
	return 110, 4, true, nil
}
func (f *fakeProRepo) KomProPlanskaCena(context.Context, int) (float64, bool, error) {
	return f.planska, true, nil
}
func (f *fakeProRepo) KomProNalogDodaj(_ context.Context, _ db.Transaction, _ int64, vrd, dokum int) (int64, error) {
	f.koraci = append(f.koraci, "nalog")
	return 46, nil
}
func (f *fakeProRepo) KomProNalogStavke(context.Context, db.Transaction, int64, int64, int, int) error {
	f.koraci = append(f.koraci, "stavke")
	return nil
}
func (f *fakeProRepo) KomProNalogFormiran(context.Context, db.Transaction, int64) (bool, error) {
	f.koraci = append(f.koraci, "status")
	return true, nil
}
func (f *fakeProRepo) KomProPrikaz(context.Context, domain.KomProPrikazParams) ([]domain.KomProPrikazRow, error) {
	return nil, nil
}
func (f *fakeProRepo) KomPrijArtikal(_ context.Context, sifra int) (domain.KomPrijArtikal, bool, error) {
	a, ok := f.artikli[sifra]
	return a, ok, nil
}
func (f *fakeProRepo) KomPrijKonta(context.Context, int) ([]domain.ComboItem, error) {
	return []domain.ComboItem{{Key: "1320", Value: "1320"}}, nil
}
func (f *fakeProRepo) KomPorNoviBroj(context.Context, db.Transaction, int, int) (int, error) {
	return 8, nil
}
func (f *fakeProRepo) KomPorIDFkpl(context.Context, string, string) (int64, error) { return 101, nil }
func (f *fakeProRepo) KomPonudaStampaDobavljac(context.Context, int64) (domain.KomPonudaStampaDobavljac, bool, error) {
	return domain.KomPonudaStampaDobavljac{Naziv: "Kupac"}, true, nil
}
func (f *fakeProRepo) KomZahtevNoviRbr(context.Context, db.Transaction, int64) (int, error) {
	return f.noviRbr, nil
}
func (f *fakeProRepo) KomZahtevStavkaBrisi(_ context.Context, _ db.Transaction, rdokID, rproID int64, rbr int) error {
	f.stavkaBris = append(f.stavkaBris, [3]int64{rdokID, rproID, int64(rbr)})
	return nil
}
func (f *fakeProRepo) KomZahtevBrisi(context.Context, db.Transaction, int64) error { return nil }
func (f *fakeProRepo) KomKomercijalista(_ context.Context, sifra int) (domain.KomKomercijalista, bool, error) {
	return domain.KomKomercijalista{KomID: 5, Sifra: 7, Ime: "Ana Anić"}, sifra == 7, nil
}
func (f *fakeProRepo) KomKomercijalistaKorisnika(context.Context, int64) (domain.KomKomercijalista, bool, error) {
	return domain.KomKomercijalista{KomID: 5, Sifra: 7, Ime: "Ana Anić"}, true, nil
}
func (f *fakeProRepo) KomMagacin(_ context.Context, id int) (int, bool, error) {
	return id * 10, id > 0, nil
}
func (f *fakeProRepo) KomFirma(context.Context, string) (domain.RobnoStampaFakturaFirmaDto, error) {
	return domain.RobnoStampaFakturaFirmaDto{Naziv: "Firma", Mesto: "Beograd", Obv: true}, nil
}
func (f *fakeProRepo) BankeZiroRacuni(context.Context) ([]string, error) {
	return []string{"160-1  Banka"}, nil
}
func (f *fakeProRepo) RporStopa(context.Context, db.Transaction, int, time.Time) (float64, error) {
	return 20, nil
}
func (f *fakeProRepo) NoviDokumentDefaults(context.Context, string) (*[]domain.SearchRowDto, error) {
	return &[]domain.SearchRowDto{{Key: "941", C3: "2040"}}, nil
}
func (f *fakeProRepo) Valuta(_ context.Context, sifval int) (string, int, bool, error) {
	switch sifval {
	case 941:
		return "RSD", 1, true, nil
	case 978:
		return "EUR", 1, true, nil
	}
	return "", 0, false, nil
}
func (f *fakeProRepo) SrednjiKurs(context.Context, int, time.Time) (float64, bool, error) {
	return 0, false, nil
}
func (f *fakeProRepo) OrgjedPostoji(_ context.Context, id int64) (bool, error)    { return id == 2, nil }
func (f *fakeProRepo) MestotrPostoji(_ context.Context, id int64) (bool, error)   { return id == 5, nil }
func (f *fakeProRepo) MestotrUOrgjed(context.Context, int64, int64) (bool, error) { return true, nil }

var proDatum = sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true}

func newFakeProRepo() *fakeProRepo {
	return &fakeProRepo{
		profakture: map[int64]domain.KomProfakturaRow{
			45: {RdokID: 45, Vrd: 360, Dokum: 7, Dadok: proDatum, Rok: 15, MagaciniID: 3, Mag: 30, Fkto: "2040", Fana: "00031", Doksts: "10", SifVal: 941,
				Kurs: 1, VaziZa: 1, Ugrabat: 10, Pkase: 2, Vrugrabat: 1, Vkase: 1, Tipzal: 1},
			46: {RdokID: 46, Vrd: 360, Dokum: 8, Dadok: proDatum, MagaciniID: 3, Mag: 30, Doksts: "20", SifVal: 978, OznVal: "EUR", Kurs: 117.25, VaziZa: 1},
			47: {RdokID: 47, Vrd: 360, Dokum: 9, MagaciniID: 3, Doksts: "40"},
		},
		stavke: map[int64][]domain.KomProStavkaRow{
			45: {{RproID: 61, Rbr: 1, Konto: "1320", Sifra: 1001, RsifID: 44, Naziv: "Brašno", JM: "kg", Otk: "001", Serija: "S1", Roktr: 82000,
				Kolicina: 10, Cena: 80, Fcena: 100, CenaVal: 100, Rabat: 5, Po: 2, Tezjm: 1},
				{RproID: 62, Rbr: 2, Konto: "1320", Sifra: 1002, RsifID: 45, Naziv: "Šećer", JM: "kg", Komercopis: "tip 400", Kolicina: 2, Cena: 50,
					Fcena: 60, CenaVal: 60, Po: 1, Tezjm: 0.5}},
		},
		artikli:    map[int]domain.KomPrijArtikal{1001: {RsifID: 44, Sifra: 1001, Naziv: "Brašno", JM: "kg", Po: 2}},
		tipdok:     []domain.ComboItem{{Key: "43", Value: "43 - Mešovito"}},
		nalogVrste: []domain.ComboItem{{Key: "350", Value: "350 - Porudžbenica kupca"}},
		tipmag:     "V",
		tipcene:    1,
		noviRbr:    3,
		lotovi:     map[string]int64{"1320|1001|30|001|S1|82000": 7},
		rste:       map[string]int64{"3|1320|1001": 9},
	}
}

// TestKomProUkupno checks the totals of the legacy Preracun.
func TestKomProUkupno(t *testing.T) {
	iznos, vrugrabat, vkase, netwght := komProUkupno(newFakeProRepo().stavke[45], 10, 2)
	// vrednost 1000 + 120 = 1120, rabat 50, ugovoreni rabat (1070) 10% = 107, kasa (963) 2% = 19.26, iznos 800 + 100.
	if iznos != 900 || vrugrabat != 107 || vkase != 19.26 || netwght != 11 {
		t.Errorf("ukupno = %v %v %v %v", iznos, vrugrabat, vkase, netwght)
	}
}

// TestKomProHelpers checks the cene in a foreign valuta and the WinDev dates.
func TestKomProHelpers(t *testing.T) {
	if fcena, cenaVal := komProCene(10.555, true, 117.25, 1); fcena != 1237.57 || cenaVal != 10.555 {
		t.Errorf("ino cene = %v %v", fcena, cenaVal)
	}
	if fcena, cenaVal := komProCene(100, false, 117.25, 1); fcena != 100 || cenaVal != 100 {
		t.Errorf("domaće cene = %v %v", fcena, cenaVal)
	}
	n := komWinDevDatum("2026-12-31")
	if got := komWinDevDatumText(n); got != "31.12.2026" {
		t.Errorf("rok = %d %q", n, got)
	}
	if komWinDevDatum("") != 0 || komWinDevDatumText(0) != "" {
		t.Errorf("an empty rok is not 0")
	}
}

// TestGetProfakturaNova checks the defaults of a new profaktura (the legacy BTN_NoviDOK and the initialization).
func TestGetProfakturaNova(t *testing.T) {
	s := NewProfakturaService(newFakeProRepo(), fakeMagacini{})
	danas := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return danas }
	dok, err := s.GetProfaktura(komCtx(), 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	h := dok.Header
	if h.Vrd != 360 || h.MagaciniID != 3 || h.Dokum != 8 || h.Tipdok != "43" || h.Fkto != "2040" || h.KomSifra != 7 || h.SifVal != 941 ||
		h.Pla != "VIRMAN" || h.Otp != "-" || h.Posuslnab != "160-1  Banka" || h.Category != "S" || !h.Datum.Equal(danas) {
		t.Errorf("header %+v", h)
	}
	if dok, _ := s.GetProfaktura(komCtx(), 47, 0, 0); dok.Poruka != komMsgProNalog || dok.Header.RdokID != 0 {
		t.Errorf("a profaktura with a nalog: %q %d", dok.Poruka, dok.Header.RdokID)
	}
}

// TestValidateProfaktura checks the legacy ValidacijaEDOK.
func TestValidateProfaktura(t *testing.T) {
	s := NewProfakturaService(newFakeProRepo(), fakeMagacini{})
	valid := domain.KomProParams{Vrd: "360", Tipdok: "43", MagaciniID: "3", Datum: "2026-03-09", Fkto: "2040", Fana: "00031", Orgjed: "2",
		MestoTroska: "5", SifVal: "941", Kurs: "1", VaziZa: "1"}
	for _, tt := range []struct {
		name   string
		change func(p *domain.KomProParams)
		want   []string
	}{
		{name: "valid", change: func(*domain.KomProParams) {}},
		{name: "without the vrsta naloga za knjiženje", change: func(p *domain.KomProParams) { p.Tipdok = "" }, want: []string{"tipdok"}},
		{name: "the vrsta naloga za knjiženje 00", change: func(p *domain.KomProParams) { p.Tipdok = "00" }, want: []string{"tipdok"}},
		{name: "a vrsta naloga za knjiženje of another vrsta", change: func(p *domain.KomProParams) { p.Tipdok = "11" }, want: []string{"tipdok"}},
		{name: "without the kupac, the OJ and the mesto troška", change: func(p *domain.KomProParams) { p.Fana, p.Orgjed, p.MestoTroska = "", "", "-" },
			want: []string{"fana", "idorgjed", "mestotrid"}},
		{name: "a foreign valuta without kurs and važi za", change: func(p *domain.KomProParams) { p.SifVal, p.Kurs, p.VaziZa = "978", "0", "0" },
			want: []string{"kurs", "vaziza"}},
		{name: "a profaktura with a nalog", change: func(p *domain.KomProParams) { p.RdokID = 47 }, want: []string{"datum"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := valid
			tt.change(&p)
			var got []string
			for _, e := range s.ValidateProfaktura(komCtx(), p) {
				got = append(got, e.Field)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestProfakturaSaveStavka checks the legacy AddEPRO (the planska cena of a magacin of the tip G, the cena of a foreign
// valuta) and the change of a stavka, each followed by the totals (Preracun).
func TestProfakturaSaveStavka(t *testing.T) {
	repo := newFakeProRepo()
	repo.tipmag, repo.planska = "G", 77
	repo.profakture[45] = func() domain.KomProfakturaRow { p := repo.profakture[45]; p.Tipmag = "G"; return p }()
	s := NewProfakturaService(repo, fakeMagacini{})
	if err := s.SaveStavka(komCtx(), domain.KomProStavkaParams{RdokID: 45, Sifra: "1001", Konto: "1320", Otk: "001", Serija: "S1", Rok: "2026-12-31",
		Kolicina: "3", Cena: "100", Rabat: "5"}); err != nil {
		t.Fatal(err)
	}
	want := domain.KomProStavkaUpis{RdokID: 45, Rbr: 3, Sifra: 1001, RsifID: 44, Po: 2, Konto: "1320", Mag: 30, MagaciniID: 3, Kolicina: 3, Cena: 77,
		Fcena: 100, CenaVal: 100, Rabat: 5, Otk: "001", Serija: "S1", Roktr: komWinDevDatum("2026-12-31"), Vrd: 360, Dokum: 7, Dadok: proDatum,
		Isporuceno: 3}
	if !reflect.DeepEqual(repo.stavkaNova, []domain.KomProStavkaUpis{want}) || len(repo.ukupno) != 1 {
		t.Errorf("stavka\n got %+v\nwant %+v (ukupno %v)", repo.stavkaNova, want, repo.ukupno)
	}
	if err := s.SaveStavka(komCtx(), domain.KomProStavkaParams{RdokID: 46, RproID: 71, Kolicina: "2", Cena: "10", Rabat: "1"}); err != nil {
		t.Fatal(err)
	}
	wantIzm := domain.KomProStavkaIzmena{RdokID: 46, RproID: 71, Kolicina: 2, Fcena: 1172.5, CenaVal: 10, Rabat: 1}
	if !reflect.DeepEqual(repo.stavkaIzm, []domain.KomProStavkaIzmena{wantIzm}) {
		t.Errorf("izmena %+v", repo.stavkaIzm)
	}
}

// TestProfakturaNalog checks the legacy FormirajNALOG: the status first (guard), then the nalog, its stavke and the
// rezervacija of every stavka (drsta of its lot, a new lot without one, and rsta) in one transaction; a profaktura with
// a nalog shows why.
func TestProfakturaNalog(t *testing.T) {
	repo := newFakeProRepo()
	s := NewProfakturaService(repo, fakeMagacini{})
	broj, poruka, err := s.FormirajNalog(komCtx(), 45)
	if err != nil || poruka != "" || broj != 8 {
		t.Fatalf("broj %d poruka %q err %v", broj, poruka, err)
	}
	if !reflect.DeepEqual(repo.koraci, []string{"status", "nalog", "stavke"}) || repo.commits != 1 {
		t.Errorf("koraci %v commits %d", repo.koraci, repo.commits)
	}
	wantRez := []string{"drsta 7 rez 10 nn=false", "rsta 9 rez 10 nn=false", "novi drsta 1320/1002 //0 rsif 45 rez 2"}
	if !reflect.DeepEqual(repo.rez, wantRez) {
		t.Errorf("rezervacija %q", repo.rez)
	}
	if _, poruka, _ := s.FormirajNalog(komCtx(), 47); poruka != komMsgProNalogPostoji {
		t.Errorf("with a nalog: %q", poruka)
	}
	repo.nalogVrste = nil
	if _, poruka, _ := s.FormirajNalog(komCtx(), 46); poruka != komMsgProNalogBezVrste {
		t.Errorf("without the vrsta: %q", poruka)
	}
}

// TestProfakturaPotvrda checks the legacy BTN_Potvrda.
func TestProfakturaPotvrda(t *testing.T) {
	s := NewProfakturaService(newFakeProRepo(), fakeMagacini{})
	for id, want := range map[int64]string{45: komMsgPotvrdjen, 46: komMsgBezStavkiPotvrda, 99: komMsgBezStavkiPotvrda} {
		if got, err := s.PotvrdiProfakturu(komCtx(), id); err != nil || got != want {
			t.Errorf("%d: %q (%v), want %q", id, got, err, want)
		}
	}
}

// TestKomProObracun checks the totals of the print: domestic with the PDV of the stopa, the ugovoreni rabat and the
// kasa; foreign without PDV at the cena in the valuta.
func TestKomProObracun(t *testing.T) {
	repo := newFakeProRepo()
	s := NewProfakturaService(repo, fakeMagacini{})
	st, poruka, err := s.GetStampa(komCtx(), 45)
	if err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	// Bruto 1120, rabat 50, neto 1070, ugovoreni rabat 107, kasa (963) 19.26; osnovice 950-95-17.10 = 837.90 and
	// 120-12-2.16 = 105.84, PDV 20% 167.58 + 21.17.
	if st.Svega != 1120 || st.Rabat != 50 || st.Ugrabat != 107 || math.Abs(st.Kasa-19.26) > 1e-9 || len(st.PdvPoStopama) != 1 {
		t.Fatalf("totals %+v", st)
	}
	if math.Abs(st.PdvPoStopama[0].Osnovica-943.74) > 1e-9 || math.Abs(st.PdvPoStopama[0].Pdv-188.75) > 1e-9 {
		t.Errorf("pdv %+v", st.PdvPoStopama)
	}
	if math.Abs(st.ZaNaplatu-(1120+188.75-50-107-19.26)) > 1e-9 || st.Slovima == "" || st.Stavke[1].Naziv != "Šećer tip 400" {
		t.Errorf("za naplatu %v slovima %q naziv %q", st.ZaNaplatu, st.Slovima, st.Stavke[1].Naziv)
	}
	repo.stavke[46] = repo.stavke[45]
	ino, _, err := s.GetStampa(komCtx(), 46)
	if err != nil || !ino.Ino || len(ino.PdvPoStopama) != 0 || ino.Slovima != "" || ino.ZaNaplatu != 1070 {
		t.Errorf("ino %+v %v", ino, err)
	}
}

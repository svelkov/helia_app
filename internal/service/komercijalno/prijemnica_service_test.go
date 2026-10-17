package komercijalno

import (
	"context"
	"database/sql"
	"math"
	"reflect"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// fakePrijRepo is an in-memory prijRepository: the prijemnice by rdokid with their stavke, the potvrđene porudžbenice
// with their stavke, the artikli, the konta of the magacini and the dobavljači; it records what was written.
type fakePrijRepo struct {
	prijemnice map[int64]domain.KomPrijemnicaRow
	stavke     map[int64][]domain.KomPrijStavkaRow
	veze       map[int64]domain.KomPrijVeza
	izvori     map[int64][]domain.KomPorStavkaRow
	artikli    map[int]domain.KomPrijArtikal
	noviRbr    int
	otkBroj    int
	vecIma     map[int64]bool

	dodat      []domain.KomPrijemnicaUpis
	izmenjen   []domain.KomPrijemnicaUpis
	potvrdjen  []string
	obrisan    []int64
	stavkaNova []domain.KomPrijStavkaUpis
	stavkaIzm  []domain.KomPrijStavkaIzmena
	stavkaBris [][3]int64
	sveBrisane []int64
	pakovanje  [][2]float64
	noviBroj   [][2]int
	commits    int
}

func (f *fakePrijRepo) BeginTx() (db.Transaction, error) {
	return &db.MockTransaction{CommitFunc: func() error { f.commits++; return nil }, RollbackFunc: func() error { return nil }}, nil
}
func (f *fakePrijRepo) KomPrijVrste(_ context.Context, magaciniID int) ([]domain.ComboItem, error) {
	if magaciniID == 4 {
		return []domain.ComboItem{{Key: "110", Value: "110 - Prijemni listovi"}}, nil
	}
	return []domain.ComboItem{{Key: "198", Value: "198 - Kalkulacija ulaza sirovine"}, {Key: "199", Value: "199 - Storno"}}, nil
}
func (f *fakePrijRepo) KomPorNoviBroj(_ context.Context, _ db.Transaction, vrd, mag int) (int, error) {
	f.noviBroj = append(f.noviBroj, [2]int{vrd, mag})
	return 5, nil
}
func (f *fakePrijRepo) KomPrijemnica(_ context.Context, id int64) (domain.KomPrijemnicaRow, bool, error) {
	p, ok := f.prijemnice[id]
	return p, ok, nil
}
func (f *fakePrijRepo) KomPrijDodaj(_ context.Context, _ db.Transaction, p domain.KomPrijemnicaUpis) (int64, error) {
	f.dodat = append(f.dodat, p)
	return 99, nil
}
func (f *fakePrijRepo) KomPrijIzmeni(_ context.Context, _ db.Transaction, p domain.KomPrijemnicaUpis) error {
	f.izmenjen = append(f.izmenjen, p)
	return nil
}
func (f *fakePrijRepo) KomPrijPotvrdi(_ context.Context, _ db.Transaction, _ int64, doksts string) (bool, error) {
	f.potvrdjen = append(f.potvrdjen, doksts)
	return true, nil
}
func (f *fakePrijRepo) KomPrijBrisi(_ context.Context, _ db.Transaction, id int64) error {
	f.obrisan = append(f.obrisan, id)
	return nil
}
func (f *fakePrijRepo) KomPrijVeza(_ context.Context, id int64) (domain.KomPrijVeza, bool, error) {
	v, ok := f.veze[id]
	return v, ok, nil
}
func (f *fakePrijRepo) KomPorDobavljac(_ context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error) {
	return domain.KomDobavljacPodaci{Naziv: "Metal", TipPdv: 1}, konto == "4330" && sifra == "00012", nil
}
func (f *fakePrijRepo) KomPorIDFkpl(context.Context, string, string) (int64, error) {
	return 101, nil
}
func (f *fakePrijRepo) KomPorStavke(_ context.Context, id int64) ([]domain.KomPorStavkaRow, error) {
	return f.izvori[id], nil
}
func (f *fakePrijRepo) KomPrijStavke(_ context.Context, id int64) ([]domain.KomPrijStavkaRow, error) {
	return f.stavke[id], nil
}
func (f *fakePrijRepo) KomPrijStavkaDodaj(_ context.Context, _ db.Transaction, st domain.KomPrijStavkaUpis) error {
	f.stavkaNova = append(f.stavkaNova, st)
	f.noviRbr++
	return nil
}
func (f *fakePrijRepo) KomPrijStavkaIzmeni(_ context.Context, _ db.Transaction, st domain.KomPrijStavkaIzmena) error {
	f.stavkaIzm = append(f.stavkaIzm, st)
	return nil
}
func (f *fakePrijRepo) KomPrijStavkaBrisi(_ context.Context, _ db.Transaction, rdokID, rproID int64, rbr int) error {
	f.stavkaBris = append(f.stavkaBris, [3]int64{rdokID, rproID, int64(rbr)})
	return nil
}
func (f *fakePrijRepo) KomPrijStavkeBrisi(_ context.Context, _ db.Transaction, id int64) error {
	f.sveBrisane = append(f.sveBrisane, id)
	return nil
}
func (f *fakePrijRepo) KomPrijIzvor(_ context.Context, _ db.Transaction, porID, rproID int64) (domain.KomPrijIzvor, bool, error) {
	for _, iz := range f.izvori[porID] {
		if iz.RproID == rproID {
			return domain.KomPrijIzvor{RproID: iz.RproID, Sifra: iz.Sifra, Po: 2, RsifID: int64(iz.Sifra), TrazKolic: iz.Kolicina, Cena: iz.Cena,
				Rabat: iz.Rabat}, true, nil
		}
	}
	return domain.KomPrijIzvor{}, false, nil
}
func (f *fakePrijRepo) KomPrijImaArtikal(_ context.Context, _ db.Transaction, _, rsifID int64) (bool, error) {
	return f.vecIma[rsifID], nil
}
func (f *fakePrijRepo) KomPrijArtikal(_ context.Context, sifra int) (domain.KomPrijArtikal, bool, error) {
	a, ok := f.artikli[sifra]
	return a, ok, nil
}
func (f *fakePrijRepo) KomPrijOtkBroj(context.Context, db.Transaction, int, int) (int, error) {
	return f.otkBroj, nil
}
func (f *fakePrijRepo) KomPrijKonta(_ context.Context, magaciniID int) ([]domain.ComboItem, error) {
	if magaciniID == 3 {
		return []domain.ComboItem{{Key: "1010", Value: "1010"}, {Key: "1011", Value: "1011"}}, nil
	}
	return nil, nil
}
func (f *fakePrijRepo) KomPrijPakovanja(context.Context, int64) ([]domain.KomPrijPakovanjeRow, error) {
	return nil, nil
}
func (f *fakePrijRepo) KomPrijPakovanjeSnimi(_ context.Context, _ db.Transaction, _, _ int64, kolPoPak, brPak float64) error {
	f.pakovanje = append(f.pakovanje, [2]float64{kolPoPak, brPak})
	return nil
}
func (f *fakePrijRepo) KomPrijPrikaz(context.Context, domain.KomPrijPrikazParams) ([]domain.KomPrijPrikazRow, error) {
	return nil, nil
}
func (f *fakePrijRepo) KomZahtevNoviRbr(context.Context, db.Transaction, int64) (int, error) {
	return f.noviRbr, nil
}
func (f *fakePrijRepo) KomMagacin(_ context.Context, id int) (int, bool, error) {
	return id * 10, id > 0, nil
}
func (f *fakePrijRepo) KomPonudaStampaDobavljac(context.Context, int64) (domain.KomPonudaStampaDobavljac, bool, error) {
	return domain.KomPonudaStampaDobavljac{Naziv: "Metal"}, true, nil
}
func (f *fakePrijRepo) KomFirma(context.Context, string) (domain.RobnoStampaFakturaFirmaDto, error) {
	return domain.RobnoStampaFakturaFirmaDto{Naziv: "Firma"}, nil
}
func (f *fakePrijRepo) BankeZiroRacuni(context.Context) ([]string, error) {
	return nil, nil
}

// fakeStanje records the postings of the potvrda (the legacy AzurRSTA_ZN and AzurDRSTA_ZN).
type fakeStanje struct {
	rsta, drsta []domain.RobPromenaStanja
	zalihe      []bool
}

func (f *fakeStanje) AzurirajRSTAZN(_ context.Context, _ db.Transaction, p domain.RobPromenaStanja, zaliha bool) error {
	f.rsta = append(f.rsta, p)
	f.zalihe = append(f.zalihe, zaliha)
	return nil
}
func (f *fakeStanje) AzurirajDRSTAZN(_ context.Context, _ db.Transaction, p domain.RobPromenaStanja, _ bool) error {
	f.drsta = append(f.drsta, p)
	return nil
}

var prijDatum = sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true}

func newFakePrijRepo() *fakePrijRepo {
	porDatum := sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true}
	return &fakePrijRepo{
		prijemnice: map[int64]domain.KomPrijemnicaRow{
			35: {RdokID: 35, Vrd: 198, DokvrstaID: 14536, Predznak: "+", Dokum: 4, Dadok: prijDatum, Datotp: prijDatum, MagaciniID: 3, Mag: 30,
				Fkto: "4330", Fana: "00012", Doksts: "10", VrdokID: 25, VezaDokum: 6, VezaDatum: porDatum},
			36: {RdokID: 36, Vrd: 199, Predznak: "-", Dokum: 5, Dadok: prijDatum, MagaciniID: 3, Mag: 30, Doksts: "10", VrdokID: 25},
			37: {RdokID: 37, Vrd: 198, Dokum: 6, MagaciniID: 3, Doksts: "40", VezaDatum: porDatum},
			38: {RdokID: 38, Vrd: 198, DokvrstaID: 14536, Predznak: "+", Dokum: 7, Dadok: prijDatum, Datotp: prijDatum, MagaciniID: 3, Mag: 30,
				Analiza: 1, Doksts: "10"},
		},
		stavke: map[int64][]domain.KomPrijStavkaRow{
			35: {{RproID: 51, Rbr: 1, Konto: "1010", Sifra: 1001, TrazKolic: 3, Kolic: 3, Serija: "S1", Otk: "001", Cena: 95},
				{RproID: 52, Rbr: 2, Konto: "1010", Sifra: 1002, TrazKolic: 2, Kolic: 2, Serija: "S2", Otk: "004", Cena: 10, Roktr: 20261231}},
			36: {{RproID: 61, Rbr: 1, Konto: "1010", Sifra: 1001, TrazKolic: 3, Kolic: 0, Serija: ""}},
			38: {{RproID: 71, Rbr: 1, Konto: "9999", Sifra: 1001, TrazKolic: 3, Kolic: 3, Serija: "S1"}},
		},
		veze: map[int64]domain.KomPrijVeza{25: {RdokID: 25, Dokum: 6, Dadok: porDatum, MagaciniID: 3, OrgjedText: "1 - Proizvodnja", Fkto: "4330",
			Fana: "00012"}},
		izvori: map[int64][]domain.KomPorStavkaRow{25: {
			{RproID: 41, Rbr: 1, Sifra: 1001, Kolicina: 3, Cena: 100, Rabat: 5},
			{RproID: 42, Rbr: 2, Sifra: 1002, Kolicina: 2, Cena: 10.55, Rabat: 10},
		}},
		artikli: map[int]domain.KomPrijArtikal{1001: {RsifID: 1001, Sifra: 1001, Naziv: "Brašno", JM: "kg", Po: 2}},
		noviRbr: 3,
		otkBroj: 4,
		vecIma:  map[int64]bool{},
	}
}

// TestGetPrijemnicaNova checks a new prijemnica of a porudžbenica (the legacy BTN_TraziDOK and PodaciHDR): its magacin,
// the first vrsta of the prijemnice of the magacin with its next broj, the dobavljač and the org. jedinica.
func TestGetPrijemnicaNova(t *testing.T) {
	repo := newFakePrijRepo()
	s := NewPrijemnicaService(repo, fakeMagacini{}, &fakeStanje{})
	danas := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return danas }
	dok, err := s.GetPrijemnica(komCtx(), 0, 0, 4, 25)
	if err != nil {
		t.Fatal(err)
	}
	want := domain.KomPrijHeaderView{Vrd: 198, Dokum: 5, Datum: danas, MagaciniID: 3, Fkto: "4330", Fana: "00012",
		Dobavljac: domain.KomDobavljacPodaci{Naziv: "Metal", TipPdv: 1, SistemPdv: "U sistemu PDV-a"}, OrgjedText: "1 - Proizvodnja",
		VrdokID: 25, VezaDokum: 6, VezaDatum: repo.veze[25].Dadok}
	if !reflect.DeepEqual(dok.Header, want) || !reflect.DeepEqual(repo.noviBroj, [][2]int{{198, 30}}) || len(dok.Konta) != 2 {
		t.Errorf("header\n got %+v\nwant %+v (novi broj %v, konta %v)", dok.Header, want, repo.noviBroj, dok.Konta)
	}
}

// TestValidatePrijemnica checks the legacy ValidacijaEDOK.
func TestValidatePrijemnica(t *testing.T) {
	s := NewPrijemnicaService(newFakePrijRepo(), fakeMagacini{}, &fakeStanje{})
	for _, tt := range []struct {
		name   string
		params domain.KomPrijParams
		want   []string
	}{
		{name: "a new one of a porudžbenica", params: domain.KomPrijParams{Vrd: "198", MagaciniID: "3", VrdokID: "25", Datum: "2026-03-09"}},
		{name: "without a porudžbenica", params: domain.KomPrijParams{Vrd: "198", MagaciniID: "3", Datum: "2026-03-09"},
			want: []string{"vrdokid"}},
		{name: "a porudžbenica of another magacin, a vrsta of another tip", params: domain.KomPrijParams{Vrd: "198", MagaciniID: "4", VrdokID: "25",
			Datum: "2026-03-09"}, want: []string{"vrd", "vrdokid"}},
		{name: "a datum before the porudžbenica", params: domain.KomPrijParams{Vrd: "198", MagaciniID: "3", VrdokID: "25", Datum: "2026-03-01"},
			want: []string{"datum"}},
		{name: "without a datum", params: domain.KomPrijParams{Vrd: "198", MagaciniID: "3", VrdokID: "25"}, want: []string{"datum"}},
		{name: "a confirmed one", params: domain.KomPrijParams{RdokID: 37, Datum: "2026-03-09"}, want: []string{"datum"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, e := range s.ValidatePrijemnica(komCtx(), tt.params) {
				got = append(got, e.Field)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSavePrijemnica checks a new prijemnica (the next broj of the vrsta in the magacin) and a saved one (its datum).
func TestSavePrijemnica(t *testing.T) {
	repo := newFakePrijRepo()
	s := NewPrijemnicaService(repo, fakeMagacini{}, &fakeStanje{})
	id, err := s.SavePrijemnica(komCtx(), domain.KomPrijParams{Vrd: "198", MagaciniID: "3", VrdokID: "25", Datum: "2026-03-09"})
	if err != nil || id != 99 {
		t.Fatalf("id %d, err %v", id, err)
	}
	want := domain.KomPrijemnicaUpis{Vrd: 198, Dokum: 5, Datum: prijDatum.Time, MagaciniID: 3, Mag: 30, VrdokID: 25}
	if !reflect.DeepEqual(repo.dodat, []domain.KomPrijemnicaUpis{want}) {
		t.Errorf("dodat %+v, want %+v", repo.dodat, want)
	}
	if _, err := s.SavePrijemnica(komCtx(), domain.KomPrijParams{RdokID: 35, Datum: "2026-03-10"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.izmenjen) != 1 || repo.izmenjen[0].RdokID != 35 || repo.commits != 2 {
		t.Errorf("izmenjen %+v, commits %d", repo.izmenjen, repo.commits)
	}
}

// TestPrijemnicaPreuzmi checks the legacy AddEPRONAR: the cena less the rabat, the OTK, the first konto of the magacin,
// and an artikal already in the prijemnica skipped with the message.
func TestPrijemnicaPreuzmi(t *testing.T) {
	repo := newFakePrijRepo()
	repo.vecIma[1001] = true
	s := NewPrijemnicaService(repo, fakeMagacini{}, &fakeStanje{})
	poruka, err := s.Preuzmi(komCtx(), 35, 0)
	if err != nil || poruka != komMsgPrijDupli {
		t.Fatalf("poruka %q, err %v", poruka, err)
	}
	want := domain.KomPrijStavkaUpis{RdokID: 35, Rbr: 3, Sifra: 1002, RsifID: 1002, Po: 2, Konto: "1010", Mag: 30, MagaciniID: 3, TrazKolic: 2,
		Otk: "005", Fcena: 10.55, Cena: 9.49, Iznos: 18.98, Rabat: 10, Vrd: 198, Dokum: 4, Dadok: prijDatum}
	if len(repo.stavkaNova) != 1 {
		t.Fatalf("stavke %+v", repo.stavkaNova)
	}
	got := repo.stavkaNova[0]
	if math.Abs(got.Cena-want.Cena) > 1e-9 || math.Abs(got.Iznos-want.Iznos) > 1e-9 {
		t.Errorf("cena %v iznos %v, want %v %v", got.Cena, got.Iznos, want.Cena, want.Iznos)
	}
	got.Cena, got.Iznos = want.Cena, want.Iznos
	if !reflect.DeepEqual(got, want) {
		t.Errorf("stavka\n got %+v\nwant %+v", got, want)
	}
	if poruka, _ := s.Preuzmi(komCtx(), 37, 0); poruka != komMsgPrijNijeIzmena {
		t.Errorf("confirmed: %q", poruka)
	}
}

// TestPrijemnicaStavka checks the legacy ValidacijaEPRO and AddEPRO: a storno has a negative primljena količina, a
// stavka without an OTK gets the one of VratiOTK.
func TestPrijemnicaStavka(t *testing.T) {
	s := NewPrijemnicaService(newFakePrijRepo(), fakeMagacini{}, &fakeStanje{})
	for _, tt := range []struct {
		name   string
		params domain.KomPrijStavkaParams
		want   []string
	}{
		{name: "valid", params: domain.KomPrijStavkaParams{RdokID: 35, Sifra: "1001", Konto: "1010", TrazKolic: "3", Kolic: "3", Serija: "S1"}},
		{name: "everything missing", params: domain.KomPrijStavkaParams{RdokID: 35, Sifra: "7", Konto: "9999"},
			want: []string{"sifraartikla", "konto", "trazkolic", "kolic", "serija"}},
		{name: "a stavka of another prijemnica", params: domain.KomPrijStavkaParams{RdokID: 35, RproID: 61, Konto: "1010", TrazKolic: "3", Kolic: "3",
			Serija: "S1"}, want: []string{"sifraartikla"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, e := range s.ValidateStavka(komCtx(), tt.params) {
				got = append(got, e.Field)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields = %v, want %v", got, tt.want)
			}
		})
	}

	repo := newFakePrijRepo()
	s = NewPrijemnicaService(repo, fakeMagacini{}, &fakeStanje{})
	if err := s.SaveStavka(komCtx(), domain.KomPrijStavkaParams{RdokID: 36, Sifra: "1001", Konto: "1010", TrazKolic: "3", Kolic: "2,5",
		Serija: " S1 "}); err != nil {
		t.Fatal(err)
	}
	if len(repo.stavkaNova) != 1 || repo.stavkaNova[0].Kolic != -2.5 || repo.stavkaNova[0].Otk != "005" || repo.stavkaNova[0].Serija != "S1" {
		t.Errorf("storno stavka %+v", repo.stavkaNova)
	}
	if err := s.SaveStavka(komCtx(), domain.KomPrijStavkaParams{RdokID: 35, RproID: 51, Konto: "1011", TrazKolic: "-3", Kolic: "3", Serija: "S9",
		Otk: "002"}); err != nil {
		t.Fatal(err)
	}
	wantIzm := domain.KomPrijStavkaIzmena{RdokID: 35, RproID: 51, Konto: "1011", TrazKolic: 3, Kolic: 3, Serija: "S9", Otk: "002"}
	if !reflect.DeepEqual(repo.stavkaIzm, []domain.KomPrijStavkaIzmena{wantIzm}) {
		t.Errorf("izmena %+v", repo.stavkaIzm)
	}
}

// TestPrijemnicaProveriPotvrdu checks the messages of the legacy BTN_NoviDOK1.
func TestPrijemnicaProveriPotvrdu(t *testing.T) {
	s := NewPrijemnicaService(newFakePrijRepo(), fakeMagacini{}, &fakeStanje{})
	for _, tt := range []struct {
		rdokID int64
		want   string
	}{
		{35, ""},
		{36, komMsgPrijNekompletne},
		{37, komMsgBezStavkiPotvrda},
		{38, komMsgPrijBezKonta},
		{99, komMsgBezStavkiPotvrda},
	} {
		if got, err := s.ProveriPotvrdu(komCtx(), tt.rdokID); err != nil || got != tt.want {
			t.Errorf("%d: %q (%v), want %q", tt.rdokID, got, err, tt.want)
		}
	}
}

// TestPotvrdiPrijemnicu checks the legacy PotvrdaDOK: the status 40 and the stavke booked into the zalihe (rsta at the
// cena on the datum, drsta by the lot).
func TestPotvrdiPrijemnicu(t *testing.T) {
	repo := newFakePrijRepo()
	stanje := &fakeStanje{}
	s := NewPrijemnicaService(repo, fakeMagacini{}, stanje)
	poruka, err := s.PotvrdiPrijemnicu(komCtx(), 35)
	if err != nil || poruka != komMsgPotvrdjen {
		t.Fatalf("poruka %q, err %v", poruka, err)
	}
	if !reflect.DeepEqual(repo.potvrdjen, []string{"40"}) || repo.commits != 1 || !reflect.DeepEqual(stanje.zalihe, []bool{true, true}) {
		t.Errorf("potvrdjen %v, commits %d, zalihe %v", repo.potvrdjen, repo.commits, stanje.zalihe)
	}
	wantRsta := domain.RobPromenaStanja{MagaciniID: 3, Konto: "1010", Sifra: 1002, Kolicina: 2, Iznos: 20, Cena: 10, Danal: prijDatum.Time,
		DokvrstaID: 14536}
	wantDrsta := domain.RobPromenaStanja{MagaciniID: 3, Konto: "1010", Sifra: 1002, Kolicina: 2, Iznos: 20, Danal: prijDatum.Time,
		DokvrstaID: 14536, Otk: "004", Serija: "S2", Roktr: 20261231}
	if len(stanje.rsta) != 2 || !reflect.DeepEqual(stanje.rsta[1], wantRsta) || !reflect.DeepEqual(stanje.drsta[1], wantDrsta) {
		t.Errorf("rsta %+v\ndrsta %+v", stanje.rsta, stanje.drsta)
	}
}

// TestPotvrdiPrijemnicuAnaliza checks a magacin with the analiza: the status 30 and the neraspoložive zalihe.
func TestPotvrdiPrijemnicuAnaliza(t *testing.T) {
	repo := newFakePrijRepo()
	repo.stavke[38][0].Konto = "1010"
	stanje := &fakeStanje{}
	s := NewPrijemnicaService(repo, fakeMagacini{}, stanje)
	if poruka, err := s.PotvrdiPrijemnicu(komCtx(), 38); err != nil || poruka != komMsgPotvrdjen {
		t.Fatalf("poruka %q, err %v", poruka, err)
	}
	if !reflect.DeepEqual(repo.potvrdjen, []string{"30"}) || !reflect.DeepEqual(stanje.zalihe, []bool{false}) {
		t.Errorf("potvrdjen %v, zalihe %v", repo.potvrdjen, stanje.zalihe)
	}
}

// TestPrijemnicaVratiPoslednju checks that "Vrati poslednju" deletes the stavka with the largest redni broj.
func TestPrijemnicaVratiPoslednju(t *testing.T) {
	repo := newFakePrijRepo()
	s := NewPrijemnicaService(repo, fakeMagacini{}, &fakeStanje{})
	if poruka, err := s.VratiPoslednju(komCtx(), 35); err != nil || poruka != "" {
		t.Fatalf("poruka %q, err %v", poruka, err)
	}
	if !reflect.DeepEqual(repo.stavkaBris, [][3]int64{{35, 52, 2}}) {
		t.Errorf("obrisano %v", repo.stavkaBris)
	}
}

// TestPrijemnicaPakovanjeKarton checks the pakovanje (not below 0) and the karton without a selected stavka.
func TestPrijemnicaPakovanjeKarton(t *testing.T) {
	repo := newFakePrijRepo()
	s := NewPrijemnicaService(repo, fakeMagacini{}, &fakeStanje{})
	if poruka, _ := s.SavePakovanje(komCtx(), domain.KomPrijPakovanjeParams{RdokID: 35, RproID: 51, KolPoPak: "-1", BrPak: "2"}); poruka != komMsgPrijPakovanje {
		t.Errorf("negative: %q", poruka)
	}
	if poruka, err := s.SavePakovanje(komCtx(), domain.KomPrijPakovanjeParams{RdokID: 35, RproID: 51, KolPoPak: "25", BrPak: "4"}); err != nil || poruka != "" {
		t.Errorf("save: %q %v", poruka, err)
	}
	if !reflect.DeepEqual(repo.pakovanje, [][2]float64{{25, 4}}) {
		t.Errorf("pakovanje %v", repo.pakovanje)
	}
	if _, poruka, _ := s.GetKarton(komCtx(), 35, 0); poruka != komMsgPrijKarton {
		t.Errorf("karton: %q", poruka)
	}
	k, poruka, err := s.GetKarton(komCtx(), 35, 52)
	if err != nil || poruka != "" || k.Dobavljac != "Metal" || k.Otk != "004" || k.Kolicina != 2 {
		t.Errorf("karton %+v %q %v", k, poruka, err)
	}
}

// TestKomPrijOtkCena checks the OTK of VratiOTK and the cena less the rabat of AddEPRONAR.
func TestKomPrijOtkCena(t *testing.T) {
	for broj, want := range map[int]string{0: "001", 9: "010", 99: "100", 120: "121"} {
		if got := komPrijOtk(broj); got != want {
			t.Errorf("otk(%d) = %q, want %q", broj, got, want)
		}
	}
	if got := komPrijCena(10.55, 10); math.Abs(got-9.49) > 1e-9 {
		t.Errorf("cena = %v, want 9.49", got)
	}
}

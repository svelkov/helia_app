package komercijalno

import (
	"context"
	"database/sql"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// fakePonudaRepo is an in-memory ponudaRepository: the zahtevi za ponudu by rdokid, their stavke, the slobodne stavke of
// the zahtevi za nabavku, the dobavljači, the artikli, the komercijalisti and the kursna lista; it records what was
// written.
type fakePonudaRepo struct {
	ponude         map[int64]domain.KomPonudaRow
	stavke         map[int64][]domain.KomZahtevStavkaRow
	izvori         map[int64]domain.KomPonudaIzvor
	dobavljaci     map[int64]domain.KomDobavljac
	artikli        map[int]domain.KomArtikal
	komercijalisti map[int]domain.KomKomercijalista
	kursevi        map[int]float64
	noviBroj       int
	noviRbr        int

	dodat         []domain.KomPonudaUpis
	izmenjen      []domain.KomPonudaUpis
	obrisan       []int64
	potvrdjen     []string
	stavkaNova    []domain.KomPonudaStavkaUpis
	kolicine      []float64
	rezervisan    []int64
	stavkaBris    [][3]int64
	commits       int
	izvoriMagacin []int
}

func (f *fakePonudaRepo) BeginTx() (db.Transaction, error) {
	return &db.MockTransaction{CommitFunc: func() error { f.commits++; return nil }, RollbackFunc: func() error { return nil }}, nil
}
func (f *fakePonudaRepo) KomPonudaVrste(context.Context) ([]domain.ComboItem, error) {
	return []domain.ComboItem{{Key: "310", Value: "310 - Zahtev za ponudu"}}, nil
}
func (f *fakePonudaRepo) KomZahtevNoviBroj(context.Context, db.Transaction, int) (int, error) {
	return f.noviBroj, nil
}
func (f *fakePonudaRepo) KomPonuda(_ context.Context, id int64) (domain.KomPonudaRow, bool, error) {
	z, ok := f.ponude[id]
	return z, ok, nil
}
func (f *fakePonudaRepo) KomPonudaDodaj(_ context.Context, _ db.Transaction, z domain.KomPonudaUpis) (int64, error) {
	f.dodat = append(f.dodat, z)
	return 99, nil
}
func (f *fakePonudaRepo) KomPonudaIzmeni(_ context.Context, _ db.Transaction, z domain.KomPonudaUpis) error {
	f.izmenjen = append(f.izmenjen, z)
	return nil
}
func (f *fakePonudaRepo) KomPonudaPotvrdi(_ context.Context, id int64, konto, sifra string) (bool, error) {
	f.potvrdjen = append(f.potvrdjen, konto+"/"+sifra)
	return true, nil
}
func (f *fakePonudaRepo) KomPonudaBrisi(_ context.Context, _ db.Transaction, id int64) error {
	f.obrisan = append(f.obrisan, id)
	return nil
}
func (f *fakePonudaRepo) KomPonudaIzvori(_ context.Context, magaciniID int) ([]domain.KomPonudaIzvorRow, error) {
	f.izvoriMagacin = append(f.izvoriMagacin, magaciniID)
	return []domain.KomPonudaIzvorRow{{RproID: 21, Sifra: 1001}}, nil
}
func (f *fakePonudaRepo) KomPonudaIzvor(_ context.Context, _ db.Transaction, id int64) (domain.KomPonudaIzvor, bool, error) {
	iz, ok := f.izvori[id]
	return iz, ok, nil
}
func (f *fakePonudaRepo) KomPonudaIzvorRezervisi(_ context.Context, _ db.Transaction, id int64) error {
	f.rezervisan = append(f.rezervisan, id)
	return nil
}
func (f *fakePonudaRepo) KomPonudaStavkaDodaj(_ context.Context, _ db.Transaction, st domain.KomPonudaStavkaUpis) error {
	f.stavkaNova = append(f.stavkaNova, st)
	f.noviRbr++
	return nil
}
func (f *fakePonudaRepo) KomPonudaStavkaKolicina(_ context.Context, _ db.Transaction, _, _ int64, kolicina float64) error {
	f.kolicine = append(f.kolicine, kolicina)
	return nil
}
func (f *fakePonudaRepo) KomPonudaStavkaBrisi(_ context.Context, _ db.Transaction, rdokID, rproID int64, rbr int) error {
	f.stavkaBris = append(f.stavkaBris, [3]int64{rdokID, rproID, int64(rbr)})
	return nil
}
func (f *fakePonudaRepo) KomZahtevStavke(_ context.Context, id int64) ([]domain.KomZahtevStavkaRow, error) {
	return f.stavke[id], nil
}
func (f *fakePonudaRepo) KomZahtevStavka(_ context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaRow, bool, error) {
	for _, st := range f.stavke[rdokID] {
		if st.RproID == rproID {
			return st, true, nil
		}
	}
	return domain.KomZahtevStavkaRow{}, false, nil
}
func (f *fakePonudaRepo) KomZahtevNoviRbr(context.Context, db.Transaction, int64) (int, error) {
	return f.noviRbr, nil
}
func (f *fakePonudaRepo) KomArtikal(_ context.Context, sifra int) (domain.KomArtikal, bool, error) {
	a, ok := f.artikli[sifra]
	return a, ok, nil
}
func (f *fakePonudaRepo) KomKomercijalista(_ context.Context, sifra int) (domain.KomKomercijalista, bool, error) {
	k, ok := f.komercijalisti[sifra]
	return k, ok, nil
}
func (f *fakePonudaRepo) KomKomercijalistaKorisnika(context.Context, int64) (domain.KomKomercijalista, bool, error) {
	return domain.KomKomercijalista{KomID: 5, Sifra: 7, Ime: "Ana Anić"}, true, nil
}
func (f *fakePonudaRepo) KomMagacin(_ context.Context, id int) (int, bool, error) {
	return id * 10, id > 0, nil
}
func (f *fakePonudaRepo) KomDobavljaci(_ context.Context, ids []int64) ([]domain.KomDobavljac, error) {
	var out []domain.KomDobavljac
	// The rows come in another order than the ids (the database gives no order).
	for i := len(ids) - 1; i >= 0; i-- {
		if d, ok := f.dobavljaci[ids[i]]; ok {
			out = append(out, d)
		}
	}
	return out, nil
}
func (f *fakePonudaRepo) KomPonudePrikaz(context.Context, domain.KomPonudaPrikazParams) ([]domain.KomPonudaPrikazRow, error) {
	return nil, nil
}
func (f *fakePonudaRepo) KomPonudaStampaDobavljac(_ context.Context, id int64) (domain.KomPonudaStampaDobavljac, bool, error) {
	d, ok := f.dobavljaci[id]
	return domain.KomPonudaStampaDobavljac{Konto: d.Konto, Sifra: d.Sifra, Naziv: d.Naziv}, ok, nil
}
func (f *fakePonudaRepo) KomFirma(context.Context, string) (domain.RobnoStampaFakturaFirmaDto, error) {
	return domain.RobnoStampaFakturaFirmaDto{Naziv: "Firma"}, nil
}
func (f *fakePonudaRepo) BankeZiroRacuni(context.Context) ([]string, error) {
	return []string{"160-1  Banka"}, nil
}
func (f *fakePonudaRepo) NoviDokumentDefaults(context.Context, string) (*[]domain.SearchRowDto, error) {
	return &[]domain.SearchRowDto{{Key: "941"}}, nil
}
func (f *fakePonudaRepo) Valuta(_ context.Context, sifval int) (string, int, bool, error) {
	switch sifval {
	case 941:
		return "RSD", 1, true, nil
	case 978:
		return "EUR", 1, true, nil
	}
	return "", 0, false, nil
}
func (f *fakePonudaRepo) SrednjiKurs(_ context.Context, sifval int, _ time.Time) (float64, bool, error) {
	k, ok := f.kursevi[sifval]
	return k, ok, nil
}
func (f *fakePonudaRepo) OrgjedPostoji(_ context.Context, id int64) (bool, error) {
	return id == 2, nil
}
func (f *fakePonudaRepo) MestotrPostoji(_ context.Context, id int64) (bool, error) {
	return id == 5 || id == 6, nil
}
func (f *fakePonudaRepo) MestotrUOrgjed(_ context.Context, mt, _ int64) (bool, error) {
	return mt == 5, nil
}

func newFakePonudaRepo() *fakePonudaRepo {
	dadok := sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true}
	return &fakePonudaRepo{
		ponude: map[int64]domain.KomPonudaRow{
			15: {RdokID: 15, Vrd: 310, Dokum: 4, MagaciniID: 3, Mag: 30, Doksts: "10", Dadok: dadok, Rok: 10, VezniID: "101;102;"},
			16: {RdokID: 16, Vrd: 310, Dokum: 5, MagaciniID: 3, Doksts: "40", VezniID: "101;"},
			17: {RdokID: 17, Vrd: 310, Dokum: 6, MagaciniID: 3, Doksts: "10", Knjige4: "S"},
			18: {RdokID: 18, Vrd: 310, Dokum: 7, MagaciniID: 3, Doksts: "10"},
		},
		stavke: map[int64][]domain.KomZahtevStavkaRow{
			15: {{RproID: 31, RdokID: 15, Rbr: 1, Sifra: 1001}, {RproID: 32, RdokID: 15, Rbr: 2, Sifra: 1002}},
			16: {{RproID: 41, RdokID: 16, Rbr: 1, Sifra: 1001}},
			18: {{RproID: 51, RdokID: 18, Rbr: 1, Sifra: 1001}},
		},
		izvori: map[int64]domain.KomPonudaIzvor{
			21: {RproID: 21, Sifra: 1001, Po: 2, RsifID: 44, Kolicina: 3},
			22: {RproID: 22, Sifra: 1002, Po: 2, RsifID: 45, Kolicina: 5},
		},
		dobavljaci: map[int64]domain.KomDobavljac{
			101: {IDFkpl: 101, Konto: "4350", Sifra: "00012", Naziv: "Metal"},
			102: {IDFkpl: 102, Konto: "4350", Sifra: "00013", Naziv: "Drvo"},
		},
		artikli:        map[int]domain.KomArtikal{1001: {RsifID: 44, Sifra: 1001, Naziv: "Vijak", JM: "kom", Po: 2}},
		komercijalisti: map[int]domain.KomKomercijalista{7: {KomID: 5, Sifra: 7, Ime: "Ana Anić"}},
		kursevi:        map[int]float64{978: 117.25},
		noviBroj:       8,
		noviRbr:        3,
	}
}

// TestGetPonudaNovi checks a new zahtev za ponudu (the legacy BTN_NoviDOK and NoviBrojEDOK): the first vrsta and
// magacin, the next broj, the datum of today, the komercijalista of the user, the valuta of the firma with the kurs 1 and
// the slobodne stavke of the zahtevi za nabavku of the magacin.
func TestGetPonudaNovi(t *testing.T) {
	repo := newFakePonudaRepo()
	s := NewZahtevPonudaService(repo, fakeMagacini{})
	danas := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return danas }
	dok, err := s.GetPonuda(komCtx(), 0, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := domain.KomPonudaHeaderView{Vrd: 310, Dokum: 8, Datum: danas, MagaciniID: 4, KomSifra: 7, KomIme: "Ana Anić", SifVal: 941, OznVal: "RSD",
		Kurs: 1, VaziZa: 1}
	if !reflect.DeepEqual(dok.Header, want) {
		t.Errorf("header\n got %+v\nwant %+v", dok.Header, want)
	}
	if len(dok.Izvori) != 1 || !reflect.DeepEqual(repo.izvoriMagacin, []int{4}) {
		t.Errorf("izvori %+v of the magacini %v", dok.Izvori, repo.izvoriMagacin)
	}

	t.Run("saved nepotvrđen: its dobavljači in the order of vezniid", func(t *testing.T) {
		dok, err := s.GetPonuda(komCtx(), 15, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		var nazivi []string
		for _, d := range dok.Header.Dobavljaci {
			nazivi = append(nazivi, d.Naziv)
		}
		if dok.Header.RdokID != 15 || dok.Header.Rok != 10 || !reflect.DeepEqual(nazivi, []string{"Metal", "Drvo"}) || dok.NoviRbr != 3 {
			t.Errorf("header %+v, nazivi %v, novi rbr %d", dok.Header, nazivi, dok.NoviRbr)
		}
	})
	t.Run("confirmed: a new one with the message", func(t *testing.T) {
		dok, err := s.GetPonuda(komCtx(), 16, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if dok.Header.RdokID != 0 || dok.Poruka != komMsgPonudaNijeIzmena {
			t.Errorf("rdokid %d, poruka %q", dok.Header.RdokID, dok.Poruka)
		}
	})
}

// TestValidatePonuda checks the legacy BTN_Sacuvaj and ValidacijaEDOK.
func TestValidatePonuda(t *testing.T) {
	s := NewZahtevPonudaService(newFakePonudaRepo(), fakeMagacini{})
	ok := domain.KomPonudaParams{Vrd: "310", MagaciniID: "3", Datum: "2026-03-02", Orgjed: "2", Dobavljaci: "101;102;", Izabrane: "21,22"}
	// Eight distinct ids of 10 digits: 88 characters of vezniid.
	var mnogo strings.Builder
	for i := range 8 {
		mnogo.WriteString(strconv.Itoa(1234567890 + i))
		mnogo.WriteString(";")
	}
	for _, tt := range []struct {
		name   string
		change func(p *domain.KomPonudaParams)
		want   []string
	}{
		{name: "valid new one", change: func(*domain.KomPonudaParams) {}},
		{name: "new one without a chosen stavka", change: func(p *domain.KomPonudaParams) { p.Izabrane = "" },
			want: []string{"magaciniid: " + komMsgPonudaBezIzvora}},
		{name: "no org. jedinica", change: func(p *domain.KomPonudaParams) { p.Orgjed = "-" },
			want: []string{"idorgjed: " + komMsgOrgjed}},
		{name: "mesto troška of another org. jedinica", change: func(p *domain.KomPonudaParams) { p.MestoTroska = "6" },
			want: []string{"mestotrid: " + komMsgMestoTroskaOJ}},
		{name: "no dobavljač", change: func(p *domain.KomPonudaParams) { p.Dobavljaci = "" },
			want: []string{"dobavljaci: " + komMsgBezDobavljaca}},
		{name: "unknown dobavljač", change: func(p *domain.KomPonudaParams) { p.Dobavljaci = "101;777;" },
			want: []string{"dobavljaci: " + komMsgDobavljacNema}},
		{name: "too many dobavljači for vezniid", change: func(p *domain.KomPonudaParams) { p.Dobavljaci = mnogo.String() + "101;" },
			want: []string{"dobavljaci: " + komMsgPrevisDobavljaca}},
		{name: "unknown komercijalista and no datum", change: func(p *domain.KomPonudaParams) { p.KomSifra = "8"; p.Datum = "" },
			want: []string{"datum: " + komMsgDatum, "komsifra: " + komMsgKomercijalista}},
		{name: "a magacin not of the user", change: func(p *domain.KomPonudaParams) { p.MagaciniID = "9" },
			want: []string{"magaciniid: " + komMsgMagacin}},
		{name: "saved confirmed one", change: func(p *domain.KomPonudaParams) { p.RdokID = 16; p.Izabrane = "" },
			want: []string{"brdok: " + komMsgPonudaNijeIzmena}},
		{name: "saved one needs no chosen stavka", change: func(p *domain.KomPonudaParams) { p.RdokID = 15; p.Izabrane = "" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ok
			tt.change(&p)
			if got := komPolja(s.ValidatePonuda(komCtx(), p)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("errors\n got %v\nwant %v", got, tt.want)
			}
		})
	}
}

// TestSavePonudaNova checks the save of a new zahtev za ponudu (the legacy SnimiEDOK and AddEPRONAR): the header with the
// oznaka of the magacin and the next broj, the dobavljači as vezniid, and the chosen stavke of the zahtevi za nabavku
// taken (the one already taken skipped) and reserved, in one transaction.
func TestSavePonudaNova(t *testing.T) {
	repo := newFakePonudaRepo()
	s := NewZahtevPonudaService(repo, fakeMagacini{})
	rdokID, err := s.SavePonuda(komCtx(), domain.KomPonudaParams{Vrd: "310", MagaciniID: "3", Datum: "2026-03-02", Rok: "10", KomSifra: "7",
		Orgjed: "2", SifVal: "978", Kurs: "117,25", VaziZa: "1", Dobavljaci: "102;101;102;", Izabrane: "21,23,22"})
	if err != nil {
		t.Fatal(err)
	}
	if rdokID != 99 || repo.commits != 1 || len(repo.dodat) != 1 {
		t.Fatalf("rdokid %d, commits %d, dodat %+v", rdokID, repo.commits, repo.dodat)
	}
	z := repo.dodat[0]
	if z.Mag != 30 || z.Dokum != 8 || z.Rok != 10 || z.VezniID != "102;101;" || z.SifVal != 978 || z.Kurs != 117.25 || !z.KomID.Valid {
		t.Errorf("header %+v", z)
	}
	if !reflect.DeepEqual(repo.rezervisan, []int64{21, 22}) || len(repo.stavkaNova) != 2 {
		t.Fatalf("rezervisani %v, stavke %+v", repo.rezervisan, repo.stavkaNova)
	}
	st := repo.stavkaNova[1]
	if st.RdokID != 99 || st.Rbr != 4 || st.RproID1 != 22 || st.Sifra != 1002 || st.Kolicina != 5 || st.Mag != 30 || st.Vrd != 310 || st.Dokum != 8 {
		t.Errorf("stavka %+v", st)
	}
}

// TestSavePonudaIzmena checks the save of a saved zahtev: its vrsta, broj and magacin stay, the header is changed.
func TestSavePonudaIzmena(t *testing.T) {
	repo := newFakePonudaRepo()
	s := NewZahtevPonudaService(repo, fakeMagacini{})
	if _, err := s.SavePonuda(komCtx(), domain.KomPonudaParams{RdokID: 15, Vrd: "999", MagaciniID: "4", Datum: "2026-03-05", Orgjed: "2",
		Dobavljaci: "101;"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.izmenjen) != 1 || len(repo.dodat) != 0 {
		t.Fatalf("izmenjen %+v, dodat %+v", repo.izmenjen, repo.dodat)
	}
	z := repo.izmenjen[0]
	if z.Vrd != 310 || z.Dokum != 4 || z.MagaciniID != 3 || z.VezniID != "101;" || z.Kurs != 1 || z.VaziZa != 1 {
		t.Errorf("header %+v", z)
	}
}

// TestPreuzmiStavke checks "+" of the stavke (the legacy BTN_DodajG).
func TestPreuzmiStavke(t *testing.T) {
	for _, tt := range []struct {
		name, izabrane string
		rdokID         int64
		want           string
		wantPreuzeto   int
	}{
		{name: "none chosen", rdokID: 15, want: komMsgNijednaIzabrana},
		{name: "confirmed zahtev", rdokID: 16, izabrane: "21", want: komMsgPonudaNijeIzmena},
		{name: "taken", rdokID: 15, izabrane: "21,22", wantPreuzeto: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakePonudaRepo()
			s := NewZahtevPonudaService(repo, fakeMagacini{})
			poruka, err := s.PreuzmiStavke(komCtx(), tt.rdokID, tt.izabrane)
			if err != nil {
				t.Fatal(err)
			}
			if poruka != tt.want || len(repo.stavkaNova) != tt.wantPreuzeto {
				t.Errorf("poruka %q, preuzeto %d", poruka, len(repo.stavkaNova))
			}
		})
	}
}

// TestDeletePonuda checks the legacy BTN_Brisanje: not a fiscalized one, the deletion in one transaction.
func TestDeletePonuda(t *testing.T) {
	repo := newFakePonudaRepo()
	s := NewZahtevPonudaService(repo, fakeMagacini{})
	if poruka, err := s.DeletePonuda(komCtx(), 17); err != nil || poruka != komMsgFiskalizovan {
		t.Errorf("fiscalized: %q %v", poruka, err)
	}
	if poruka, err := s.DeletePonuda(komCtx(), 15); err != nil || poruka != "" || !reflect.DeepEqual(repo.obrisan, []int64{15}) || repo.commits != 1 {
		t.Errorf("deleted: %q %v %v", poruka, err, repo.obrisan)
	}
}

// TestPotvrdiPonudu checks the potvrda (the legacy BTN_Potvrda and pr_PotvrdaPONUDE): the checks, the dobavljač that
// must be one of the zahtev and its konto and šifra as fkto and fana.
func TestPotvrdiPonudu(t *testing.T) {
	for _, tt := range []struct {
		name          string
		rdokID, fkpl  int64
		want          string
		wantPotvrdjen []string
	}{
		{name: "no stavke", rdokID: 17, fkpl: 101, want: komMsgBezStavkiPotvrda},
		{name: "already confirmed", rdokID: 16, fkpl: 101, want: komMsgVecPotvrdjen},
		{name: "no dobavljači", rdokID: 18, fkpl: 101, want: komMsgPotvrdaBezDob},
		{name: "a dobavljač not of the zahtev", rdokID: 15, fkpl: 777, want: komMsgPotvrdaDobavljac},
		{name: "confirmed", rdokID: 15, fkpl: 102, want: komMsgPotvrdjen, wantPotvrdjen: []string{"4350/00013"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakePonudaRepo()
			s := NewZahtevPonudaService(repo, fakeMagacini{})
			poruka, err := s.PotvrdiPonudu(komCtx(), tt.rdokID, tt.fkpl)
			if err != nil {
				t.Fatal(err)
			}
			if poruka != tt.want || !reflect.DeepEqual(repo.potvrdjen, tt.wantPotvrdjen) {
				t.Errorf("poruka %q, potvrđen %v", poruka, repo.potvrdjen)
			}
		})
	}
}

// TestPonudaStavka checks the stavke: a new one entered by hand (no stavka of a zahtev za nabavku), the change of a
// količina only, the validation and the deletion.
func TestPonudaStavka(t *testing.T) {
	repo := newFakePonudaRepo()
	s := NewZahtevPonudaService(repo, fakeMagacini{})
	if got := komPolja(s.ValidateStavka(komCtx(), domain.KomPonudaStavkaParams{RdokID: 15, Sifra: "9999", Kolicina: "0"})); !reflect.DeepEqual(got,
		[]string{"kolicina: " + komMsgKolicina, "sifraartikla: " + komMsgArtikal}) {
		t.Errorf("errors %v", got)
	}
	if got := s.ValidateStavka(komCtx(), domain.KomPonudaStavkaParams{RdokID: 15, RproID: 31, Sifra: "9999", Kolicina: "2"}); len(got) != 0 {
		t.Errorf("a change validates no šifra: %v", got)
	}
	if err := s.SaveStavka(komCtx(), domain.KomPonudaStavkaParams{RdokID: 15, Sifra: "1001", Kolicina: "4"}); err != nil {
		t.Fatal(err)
	}
	st := repo.stavkaNova[0]
	if st.RproID1 != 0 || st.Rbr != 3 || st.Sifra != 1001 || st.Kolicina != 4 || st.Mag != 30 || st.Dokum != 4 {
		t.Errorf("stavka %+v", st)
	}
	if err := s.SaveStavka(komCtx(), domain.KomPonudaStavkaParams{RdokID: 15, RproID: 31, Kolicina: "2,5"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.kolicine, []float64{2.5}) {
		t.Errorf("količine %v", repo.kolicine)
	}
	if err := s.DeleteStavka(komCtx(), 15, 32); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.stavkaBris, [][3]int64{{15, 32, 2}}) {
		t.Errorf("deleted %v", repo.stavkaBris)
	}
}

// TestGetValuta checks the kurs of a valuta (the legacy exit of EDT_SIFVAL): the srednji kurs on the datum, else 1.
func TestGetValuta(t *testing.T) {
	s := NewZahtevPonudaService(newFakePonudaRepo(), fakeMagacini{})
	for _, tt := range []struct {
		name, sifval, datum string
		want                domain.KomValuta
	}{
		{name: "kurs of the day", sifval: "978", datum: "2026-03-02", want: domain.KomValuta{SifVal: 978, OznVal: "EUR", Kurs: 117.25, VaziZa: 1, Found: true}},
		{name: "no kursna lista", sifval: "941", datum: "2026-03-02", want: domain.KomValuta{SifVal: 941, OznVal: "RSD", Kurs: 1, VaziZa: 1, Found: true}},
		{name: "unknown valuta", sifval: "5", datum: "", want: domain.KomValuta{SifVal: 5, Kurs: 1, VaziZa: 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetValuta(komCtx(), tt.sifval, tt.datum)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestGetStampaPonuda checks the print: only for a dobavljač of the zahtev, with the firm, its računi and the stavke.
func TestGetStampaPonuda(t *testing.T) {
	s := NewZahtevPonudaService(newFakePonudaRepo(), fakeMagacini{})
	if _, poruka, err := s.GetStampa(komCtx(), 15, 777); err != nil || poruka != komMsgPotvrdaDobavljac {
		t.Errorf("another dobavljač: %q %v", poruka, err)
	}
	if _, poruka, err := s.GetStampa(komCtx(), 17, 101); err != nil || poruka != komMsgBezStavkiStampa {
		t.Errorf("no stavke: %q %v", poruka, err)
	}
	stampa, poruka, err := s.GetStampa(komCtx(), 15, 101)
	if err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	if stampa.Dobavljac.Naziv != "Metal" || stampa.Firma.Naziv != "Firma" || len(stampa.ZiroRacuni) != 1 || len(stampa.Stavke) != 2 ||
		stampa.Rok != 10 || !stampa.Nepotvrdjen {
		t.Errorf("stampa %+v", stampa)
	}
}

// TestKomIDovi checks the lists of ids (vezniid, the chosen stavke) and the vezniid of the dobavljači.
func TestKomIDovi(t *testing.T) {
	for _, tt := range []struct {
		value  string
		want   []int64
		wantOK bool
	}{
		{value: "", wantOK: true},
		{value: "101;102;", want: []int64{101, 102}, wantOK: true},
		{value: "21, 22,21", want: []int64{21, 22}, wantOK: true},
		{value: "101;x;0;", want: []int64{101}, wantOK: false},
	} {
		got, ok := komIDovi(tt.value)
		if !reflect.DeepEqual(got, tt.want) || ok != tt.wantOK {
			t.Errorf("komIDovi(%q) = %v %v, want %v %v", tt.value, got, ok, tt.want, tt.wantOK)
		}
	}
	if got := komVezniID([]int64{101, 102}); got != "101;102;" {
		t.Errorf("vezniid %q", got)
	}
}

// TestPonudaPrikazTableRows checks the očekivani datum nabavke of the pregled: the datum plus the rok.
func TestPonudaPrikazTableRows(t *testing.T) {
	rows := PonudaPrikazTableRows([]domain.KomPonudaPrikazRow{{RdokID: 15, Dokum: 4, Rok: 10,
		Dadok: sql.NullTime{Time: time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC), Valid: true}}})
	if got := rows[0].Fields[5]; got != "04.04.2026" {
		t.Errorf("očekivani datum %q", got)
	}
}

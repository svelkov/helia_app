package komercijalno

import (
	"context"
	"database/sql"
	"reflect"
	"sort"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	commonsvc "helia/internal/service/common"
)

// fakeZahtevRepo is an in-memory zahtevRepository: the zahtevi by rdokid, their stavke, the artikli and the
// komercijalisti by šifra; it records what was written.
type fakeZahtevRepo struct {
	zahtevi        map[int64]domain.KomZahtevRow
	stavke         map[int64][]domain.KomZahtevStavkaRow
	artikli        map[int]domain.KomArtikal
	komercijalisti map[int]domain.KomKomercijalista
	korisnika      *domain.KomKomercijalista
	noviBroj       int
	noviRbr        int

	dodat      []domain.KomZahtevUpis
	izmenjen   []domain.KomZahtevUpis
	obrisan    []int64
	potvrdjen  []int64
	stavkaNova []domain.KomZahtevStavkaUpis
	stavkaIzm  []domain.KomZahtevStavkaUpis
	stavkaBris [][3]int64
	commits    int
}

func (f *fakeZahtevRepo) BeginTx() (db.Transaction, error) {
	return &db.MockTransaction{CommitFunc: func() error { f.commits++; return nil }, RollbackFunc: func() error { return nil }}, nil
}
func (f *fakeZahtevRepo) KomZahtevVrste(context.Context) ([]domain.ComboItem, error) {
	return []domain.ComboItem{{Key: "410", Value: "410 - Zahtev"}, {Key: "411", Value: "411 - Zahtev 2"}}, nil
}
func (f *fakeZahtevRepo) KomZahtevNoviBroj(context.Context, db.Transaction, int) (int, error) {
	return f.noviBroj, nil
}
func (f *fakeZahtevRepo) KomZahtev(_ context.Context, id int64) (domain.KomZahtevRow, bool, error) {
	z, ok := f.zahtevi[id]
	return z, ok, nil
}
func (f *fakeZahtevRepo) KomZahtevDodaj(_ context.Context, _ db.Transaction, z domain.KomZahtevUpis) (int64, error) {
	f.dodat = append(f.dodat, z)
	return 99, nil
}
func (f *fakeZahtevRepo) KomZahtevIzmeni(_ context.Context, _ db.Transaction, z domain.KomZahtevUpis) error {
	f.izmenjen = append(f.izmenjen, z)
	return nil
}
func (f *fakeZahtevRepo) KomZahtevPotvrdi(_ context.Context, id int64) (bool, error) {
	f.potvrdjen = append(f.potvrdjen, id)
	return true, nil
}
func (f *fakeZahtevRepo) KomZahtevBrisi(_ context.Context, _ db.Transaction, id int64) error {
	f.obrisan = append(f.obrisan, id)
	return nil
}
func (f *fakeZahtevRepo) KomZahtevStavke(_ context.Context, id int64) ([]domain.KomZahtevStavkaRow, error) {
	return f.stavke[id], nil
}
func (f *fakeZahtevRepo) KomZahtevStavka(_ context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaRow, bool, error) {
	for _, st := range f.stavke[rdokID] {
		if st.RproID == rproID {
			return st, true, nil
		}
	}
	return domain.KomZahtevStavkaRow{}, false, nil
}
func (f *fakeZahtevRepo) KomZahtevNoviRbr(context.Context, db.Transaction, int64) (int, error) {
	return f.noviRbr, nil
}
func (f *fakeZahtevRepo) KomZahtevStavkaDodaj(_ context.Context, _ db.Transaction, st domain.KomZahtevStavkaUpis) error {
	f.stavkaNova = append(f.stavkaNova, st)
	return nil
}
func (f *fakeZahtevRepo) KomZahtevStavkaIzmeni(_ context.Context, _ db.Transaction, st domain.KomZahtevStavkaUpis) error {
	f.stavkaIzm = append(f.stavkaIzm, st)
	return nil
}
func (f *fakeZahtevRepo) KomZahtevStavkaBrisi(_ context.Context, _ db.Transaction, rdokID, rproID int64, rbr int) error {
	f.stavkaBris = append(f.stavkaBris, [3]int64{rdokID, rproID, int64(rbr)})
	return nil
}
func (f *fakeZahtevRepo) KomArtikal(_ context.Context, sifra int) (domain.KomArtikal, bool, error) {
	a, ok := f.artikli[sifra]
	return a, ok, nil
}
func (f *fakeZahtevRepo) KomKomercijalista(_ context.Context, sifra int) (domain.KomKomercijalista, bool, error) {
	k, ok := f.komercijalisti[sifra]
	return k, ok, nil
}
func (f *fakeZahtevRepo) KomKomercijalistaKorisnika(context.Context, int64) (domain.KomKomercijalista, bool, error) {
	if f.korisnika == nil {
		return domain.KomKomercijalista{}, false, nil
	}
	return *f.korisnika, true, nil
}
func (f *fakeZahtevRepo) KomMagacin(_ context.Context, id int) (int, bool, error) {
	return id * 10, id > 0, nil
}
func (f *fakeZahtevRepo) KomZahteviPrikaz(context.Context, domain.KomZahtevPrikazParams) ([]domain.KomZahtevPrikazRow, error) {
	return nil, nil
}
func (f *fakeZahtevRepo) KomZahtevStampaZaglavlje(context.Context, int64) (domain.KomZahtevStampa, bool, error) {
	return domain.KomZahtevStampa{Dokum: 12}, true, nil
}
func (f *fakeZahtevRepo) KomFirma(context.Context, string) (domain.RobnoStampaFakturaFirmaDto, error) {
	return domain.RobnoStampaFakturaFirmaDto{Naziv: "Firma"}, nil
}
func (f *fakeZahtevRepo) OrgjedPostoji(_ context.Context, id int64) (bool, error) {
	return id == 2, nil
}
func (f *fakeZahtevRepo) MestotrPostoji(_ context.Context, id int64) (bool, error) {
	return id == 5 || id == 6, nil
}
func (f *fakeZahtevRepo) MestotrUOrgjed(_ context.Context, mt, _ int64) (bool, error) {
	return mt == 5, nil
}

// fakeMagacini are the magacini of the user: 3 and 4.
type fakeMagacini struct{}

func (fakeMagacini) GetMagacinKorisnikaComboValues(context.Context, ...commonsvc.ComboOption) ([]domain.ComboItem, error) {
	return []domain.ComboItem{{Key: "3", Value: "1 - Sirovine"}, {Key: "4", Value: "2 - Rezervni delovi"}}, nil
}

func komCtx() context.Context {
	return domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserID: 4, UserName: "ana", Firma: "Firma", SelectedGod: 2026, SelectedKar: 7})
}

func newFakeZahtevRepo() *fakeZahtevRepo {
	return &fakeZahtevRepo{
		zahtevi: map[int64]domain.KomZahtevRow{
			9:  {RdokID: 9, Vrd: 410, Dokum: 12, MagaciniID: 3, Mag: 30, Doksts: "10", Dadok: sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true}},
			10: {RdokID: 10, Vrd: 410, Dokum: 13, MagaciniID: 3, Doksts: "40"},
			11: {RdokID: 11, Vrd: 410, Dokum: 14, MagaciniID: 3, Doksts: "10", Knjige4: "S"},
			12: {RdokID: 12, Vrd: 410, Dokum: 15, MagaciniID: 3, Doksts: "10", Knjige1: "D"},
		},
		stavke: map[int64][]domain.KomZahtevStavkaRow{
			9:  {{RproID: 21, RdokID: 9, Rbr: 1, Sifra: 1001}, {RproID: 22, RdokID: 9, Rbr: 2, Sifra: 1002}},
			10: {{RproID: 31, RdokID: 10, Rbr: 1, Sifra: 1001}},
		},
		artikli:        map[int]domain.KomArtikal{1001: {RsifID: 44, Sifra: 1001, Naziv: "Vijak", JM: "kom", Po: 2, NCena: 12.5}},
		komercijalisti: map[int]domain.KomKomercijalista{7: {KomID: 5, Sifra: 7, Ime: "Ana Anić"}},
		noviBroj:       13,
		noviRbr:        3,
	}
}

func komPolja(errs []domain.FieldError) []string {
	var polja []string
	for _, e := range errs {
		polja = append(polja, e.Field+": "+e.ErrorMessage)
	}
	sort.Strings(polja)
	return polja
}

// TestGetZahtevNovi checks a new zahtev (the legacy BTN_NoviDOK): the given vrsta and magacin (else the first ones),
// the next broj, the datum of today, the user as the one who entered it and the komercijalista of the user.
func TestGetZahtevNovi(t *testing.T) {
	repo := newFakeZahtevRepo()
	repo.korisnika = &domain.KomKomercijalista{KomID: 5, Sifra: 7, Ime: "Ana Anić"}
	s := NewZahtevNabavkaService(repo, fakeMagacini{})
	danas := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return danas }
	for _, tt := range []struct {
		name          string
		vrd, magacin  int
		wantVrd, wMag int
	}{
		{name: "given vrsta and magacin", vrd: 411, magacin: 4, wantVrd: 411, wMag: 4},
		{name: "unknown ones: the first ones", vrd: 999, magacin: 9, wantVrd: 410, wMag: 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok, err := s.GetZahtev(komCtx(), 0, tt.vrd, tt.magacin)
			if err != nil {
				t.Fatal(err)
			}
			want := domain.KomZahtevHeaderView{Vrd: tt.wantVrd, Dokum: 13, Datum: danas, MagaciniID: tt.wMag, KomSifra: 7, KomIme: "Ana Anić", Opunos: "ana"}
			if !reflect.DeepEqual(dok.Header, want) {
				t.Errorf("header\n got %+v\nwant %+v", dok.Header, want)
			}
			if dok.NoviRbr != 1 || dok.Poruka != "" {
				t.Errorf("novi rbr %d, poruka %q", dok.NoviRbr, dok.Poruka)
			}
		})
	}
	t.Run("saved nepotvrđen", func(t *testing.T) {
		dok, err := s.GetZahtev(komCtx(), 9, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if dok.Header.RdokID != 9 || dok.Header.Dokum != 12 || dok.NoviRbr != 3 || dok.Poruka != "" {
			t.Errorf("got %+v, novi rbr %d, poruka %q", dok.Header, dok.NoviRbr, dok.Poruka)
		}
	})
	t.Run("potvrđen: a new one with the message", func(t *testing.T) {
		dok, err := s.GetZahtev(komCtx(), 10, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if dok.Header.RdokID != 0 || dok.Poruka != komMsgNijeIzmena || dok.Header.Vrd != 410 || dok.Header.MagaciniID != 3 {
			t.Errorf("got %+v, poruka %q", dok.Header, dok.Poruka)
		}
	})
}

// TestValidateZahtev checks the validation of the header (the legacy ValidacijaEDOK).
func TestValidateZahtev(t *testing.T) {
	s := NewZahtevNabavkaService(newFakeZahtevRepo(), fakeMagacini{})
	ok := domain.KomZahtevParams{Vrd: "410", MagaciniID: "3", Datum: "2026-03-02", KomSifra: "7", Orgjed: "2", MestoTroska: "5"}
	for _, tt := range []struct {
		name   string
		change func(p *domain.KomZahtevParams)
		want   []string
	}{
		{name: "valid", change: func(*domain.KomZahtevParams) {}},
		{name: "no org. jedinica, komercijalista or mesto troška", change: func(p *domain.KomZahtevParams) {
			p.KomSifra, p.Orgjed, p.MestoTroska = "", "-", "-"
		}},
		{name: "vrsta, magacin and datum", change: func(p *domain.KomZahtevParams) { p.Vrd, p.MagaciniID, p.Datum = "999", "9", "" },
			want: []string{"datum: " + komMsgDatum, "magaciniid: " + komMsgMagacin, "vrd: " + komMsgVrsta}},
		{name: "unknown komercijalista", change: func(p *domain.KomZahtevParams) { p.KomSifra = "8" },
			want: []string{"komsifra: " + komMsgKomercijalista}},
		{name: "mesto troška of another org. jedinica", change: func(p *domain.KomZahtevParams) { p.MestoTroska = "6" },
			want: []string{"mestotrid: " + komMsgMestoTroskaOJ}},
		{name: "a confirmed zahtev", change: func(p *domain.KomZahtevParams) { p.RdokID = 10 },
			want: []string{"brdok: " + komMsgNijeIzmena}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ok
			tt.change(&p)
			if got := komPolja(s.ValidateZahtev(komCtx(), p)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("errors %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSaveZahtev checks the save of the header (the legacy SnimiEDOK): a new zahtev with the next broj and the oznaka
// of its magacin, a saved one changed, both in a committed transaction.
func TestSaveZahtev(t *testing.T) {
	repo := newFakeZahtevRepo()
	s := NewZahtevNabavkaService(repo, fakeMagacini{})
	p := domain.KomZahtevParams{Vrd: "410", MagaciniID: "3", Datum: "2026-03-02", KomSifra: "7", Orgjed: "2", MestoTroska: "-", Tekst: " Hitno "}
	id, err := s.SaveZahtev(komCtx(), p)
	if err != nil || id != 99 {
		t.Fatalf("new: id %d, err %v", id, err)
	}
	want := domain.KomZahtevUpis{Vrd: 410, Dokum: 13, Dadok: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), MagaciniID: 3, Mag: 30,
		KomID: sql.NullInt64{Int64: 5, Valid: true}, Orgjed: sql.NullInt64{Int64: 2, Valid: true}, Foot: "Hitno"}
	if len(repo.dodat) != 1 || !reflect.DeepEqual(repo.dodat[0], want) {
		t.Errorf("added %+v, want %+v", repo.dodat, want)
	}
	p.RdokID, p.KomSifra = 9, ""
	if id, err := s.SaveZahtev(komCtx(), p); err != nil || id != 9 {
		t.Fatalf("change: id %d, err %v", id, err)
	}
	if len(repo.izmenjen) != 1 || repo.izmenjen[0].RdokID != 9 || repo.izmenjen[0].KomID.Valid {
		t.Errorf("changed %+v", repo.izmenjen)
	}
	if repo.commits != 2 {
		t.Errorf("%d commits, want 2", repo.commits)
	}
}

// TestDeleteZahtev checks the deletion (the legacy BTN_Brisanje): not a fiscalized or a posted zahtev.
func TestDeleteZahtev(t *testing.T) {
	for _, tt := range []struct {
		name    string
		id      int64
		poruka  string
		obrisan bool
	}{
		{name: "nepotvrđen", id: 9, obrisan: true},
		{name: "fiscalized", id: 11, poruka: komMsgFiskalizovan},
		{name: "posted", id: 12, poruka: komMsgProknjizen},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeZahtevRepo()
			poruka, err := NewZahtevNabavkaService(repo, fakeMagacini{}).DeleteZahtev(komCtx(), tt.id)
			if err != nil || poruka != tt.poruka || (len(repo.obrisan) == 1) != tt.obrisan {
				t.Errorf("poruka %q err %v obrisan %v", poruka, err, repo.obrisan)
			}
		})
	}
}

// TestPotvrdiZahtev checks the potvrda (the legacy BTN_Potvrda): a zahtev with stavke, not confirmed yet.
func TestPotvrdiZahtev(t *testing.T) {
	for _, tt := range []struct {
		name      string
		id        int64
		poruka    string
		potvrdjen bool
	}{
		{name: "nepotvrđen with stavke", id: 9, poruka: komMsgPotvrdjen, potvrdjen: true},
		{name: "without stavke", id: 11, poruka: komMsgBezStavkiPotvrda},
		{name: "already confirmed", id: 10, poruka: komMsgVecPotvrdjen},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeZahtevRepo()
			poruka, err := NewZahtevNabavkaService(repo, fakeMagacini{}).PotvrdiZahtev(komCtx(), tt.id)
			if err != nil || poruka != tt.poruka || (len(repo.potvrdjen) == 1) != tt.potvrdjen {
				t.Errorf("poruka %q err %v potvrđen %v", poruka, err, repo.potvrdjen)
			}
		})
	}
}

// TestStavke checks the validation (the legacy ValidacijaEPRO), the save (AddEPRO: the next redni broj, the nabavna
// cena of the artikal and the document of the zahtev; a change in its place) and the deletion of a stavka (the
// redni broj of the deleted one moves the ones after it).
func TestStavke(t *testing.T) {
	repo := newFakeZahtevRepo()
	s := NewZahtevNabavkaService(repo, fakeMagacini{})
	if got := komPolja(s.ValidateStavka(komCtx(), domain.KomZahtevStavkaParams{RdokID: 9, Sifra: "1001", Kolicina: "3"})); got != nil {
		t.Errorf("valid stavka: %v", got)
	}
	want := []string{"kolicina: " + komMsgKolicina, "sifraartikla: " + komMsgArtikal}
	if got := komPolja(s.ValidateStavka(komCtx(), domain.KomZahtevStavkaParams{RdokID: 9, Sifra: "5", Kolicina: "0"})); !reflect.DeepEqual(got, want) {
		t.Errorf("invalid stavka: %v, want %v", got, want)
	}
	if got := komPolja(s.ValidateStavka(komCtx(), domain.KomZahtevStavkaParams{RdokID: 10, Sifra: "1001", Kolicina: "1"})); len(got) != 1 {
		t.Errorf("stavka of a confirmed zahtev: %v", got)
	}
	if err := s.SaveStavka(komCtx(), domain.KomZahtevStavkaParams{RdokID: 9, Sifra: "1001", Kolicina: "3,5"}); err != nil {
		t.Fatal(err)
	}
	wantNova := domain.KomZahtevStavkaUpis{RdokID: 9, Rbr: 3, Sifra: 1001, RsifID: 44, Po: 2, Cena: 12.5, Kolicina: 3.5, Mag: 30, MagaciniID: 3,
		Vrd: 410, Dokum: 12, Dadok: repo.zahtevi[9].Dadok}
	if len(repo.stavkaNova) != 1 || !reflect.DeepEqual(repo.stavkaNova[0], wantNova) {
		t.Errorf("new stavka %+v, want %+v", repo.stavkaNova, wantNova)
	}
	if err := s.SaveStavka(komCtx(), domain.KomZahtevStavkaParams{RdokID: 9, RproID: 22, Sifra: "1001", Kolicina: "2"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.stavkaIzm) != 1 || repo.stavkaIzm[0].RproID != 22 || repo.stavkaIzm[0].Kolicina != 2 {
		t.Errorf("changed stavka %+v", repo.stavkaIzm)
	}
	if err := s.DeleteStavka(komCtx(), 9, 22); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.stavkaBris, [][3]int64{{9, 22, 2}}) {
		t.Errorf("deleted %v", repo.stavkaBris)
	}
}

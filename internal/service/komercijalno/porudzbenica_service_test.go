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

// fakePorRepo is an in-memory porRepository: the porudžbenice by rdokid, their stavke, the potvrđeni zahtevi za ponudu
// with their slobodne stavke, the dobavljači by "konto/šifra", the artikli and the komercijalisti; it records what was
// written.
type fakePorRepo struct {
	porudzbenice map[int64]domain.KomPorudzbenicaRow
	stavke       map[int64][]domain.KomPorStavkaRow
	veze         map[int64]domain.KomPorVeza
	izvori       map[int64][]domain.KomPorIzvor
	dobavljaci   map[string]domain.KomDobavljacPodaci
	artikli      map[int]domain.KomArtikal
	noviRbr      int

	dodat      []domain.KomPorudzbenicaUpis
	izmenjen   []domain.KomPorudzbenicaUpis
	obrisan    []int64
	potvrdjen  []int64
	stavkaNova []domain.KomPorStavkaUpis
	stavkaIzm  [][4]float64
	stavkaBris [][3]int64
	vratiSve   []int64
	rezervisan []int64
	noviBroj   [][2]int
	commits    int
}

func (f *fakePorRepo) BeginTx() (db.Transaction, error) {
	return &db.MockTransaction{CommitFunc: func() error { f.commits++; return nil }, RollbackFunc: func() error { return nil }}, nil
}
func (f *fakePorRepo) KomPorVrste(context.Context) ([]domain.ComboItem, error) {
	return []domain.ComboItem{{Key: "320", Value: "320 - Porudžbenica"}}, nil
}
func (f *fakePorRepo) KomPorNoviBroj(_ context.Context, _ db.Transaction, vrd, mag int) (int, error) {
	f.noviBroj = append(f.noviBroj, [2]int{vrd, mag})
	return 7, nil
}
func (f *fakePorRepo) KomPorudzbenica(_ context.Context, id int64) (domain.KomPorudzbenicaRow, bool, error) {
	p, ok := f.porudzbenice[id]
	return p, ok, nil
}
func (f *fakePorRepo) KomPorDodaj(_ context.Context, _ db.Transaction, p domain.KomPorudzbenicaUpis) (int64, error) {
	f.dodat = append(f.dodat, p)
	return 99, nil
}
func (f *fakePorRepo) KomPorIzmeni(_ context.Context, _ db.Transaction, p domain.KomPorudzbenicaUpis) error {
	f.izmenjen = append(f.izmenjen, p)
	return nil
}
func (f *fakePorRepo) KomPorPotvrdi(_ context.Context, id int64) (bool, error) {
	f.potvrdjen = append(f.potvrdjen, id)
	return true, nil
}
func (f *fakePorRepo) KomPorBrisi(_ context.Context, _ db.Transaction, id int64) error {
	f.obrisan = append(f.obrisan, id)
	return nil
}
func (f *fakePorRepo) KomPorVeza(_ context.Context, id int64) (domain.KomPorVeza, bool, error) {
	v, ok := f.veze[id]
	return v, ok, nil
}
func (f *fakePorRepo) KomPorDobavljac(_ context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error) {
	d, ok := f.dobavljaci[konto+"/"+sifra]
	return d, ok, nil
}
func (f *fakePorRepo) KomPorIDFkpl(_ context.Context, konto, sifra string) (int64, error) {
	if _, ok := f.dobavljaci[konto+"/"+sifra]; ok {
		return 101, nil
	}
	return 0, nil
}
func (f *fakePorRepo) KomPorStavke(_ context.Context, id int64) ([]domain.KomPorStavkaRow, error) {
	return f.stavke[id], nil
}
func (f *fakePorRepo) KomPorStavkaDodaj(_ context.Context, _ db.Transaction, st domain.KomPorStavkaUpis) error {
	f.stavkaNova = append(f.stavkaNova, st)
	f.noviRbr++
	return nil
}
func (f *fakePorRepo) KomPorStavkaIzmeni(_ context.Context, _ db.Transaction, _, _ int64, kolicina, cena, cenaVal, rabat float64) error {
	f.stavkaIzm = append(f.stavkaIzm, [4]float64{kolicina, cena, cenaVal, rabat})
	return nil
}
func (f *fakePorRepo) KomPorVratiSve(_ context.Context, _ db.Transaction, id int64) error {
	f.vratiSve = append(f.vratiSve, id)
	return nil
}
func (f *fakePorRepo) KomPorIzvori(_ context.Context, ponudaID int64) ([]domain.KomPorIzvorRow, error) {
	var out []domain.KomPorIzvorRow
	for _, iz := range f.izvori[ponudaID] {
		out = append(out, domain.KomPorIzvorRow{RproID: iz.RproID, Sifra: iz.Sifra, Kolicina: iz.Kolicina})
	}
	return out, nil
}
func (f *fakePorRepo) KomPorIzvor(_ context.Context, _ db.Transaction, ponudaID, rproID int64) (domain.KomPorIzvor, bool, error) {
	for _, iz := range f.izvori[ponudaID] {
		if iz.RproID == rproID {
			return iz, true, nil
		}
	}
	return domain.KomPorIzvor{}, false, nil
}
func (f *fakePorRepo) KomPorPrikaz(context.Context, domain.KomPonudaPrikazParams) ([]domain.KomPorPrikazRow, error) {
	return nil, nil
}
func (f *fakePorRepo) KomPonudaIzvorRezervisi(_ context.Context, _ db.Transaction, id int64) error {
	f.rezervisan = append(f.rezervisan, id)
	return nil
}
func (f *fakePorRepo) KomPonudaStavkaBrisi(_ context.Context, _ db.Transaction, rdokID, rproID int64, rbr int) error {
	f.stavkaBris = append(f.stavkaBris, [3]int64{rdokID, rproID, int64(rbr)})
	return nil
}
func (f *fakePorRepo) KomPonudaStampaDobavljac(context.Context, int64) (domain.KomPonudaStampaDobavljac, bool, error) {
	return domain.KomPonudaStampaDobavljac{Konto: "4330", Sifra: "00012", Naziv: "Metal"}, true, nil
}
func (f *fakePorRepo) KomZahtevNoviRbr(context.Context, db.Transaction, int64) (int, error) {
	return f.noviRbr, nil
}
func (f *fakePorRepo) KomArtikal(_ context.Context, sifra int) (domain.KomArtikal, bool, error) {
	a, ok := f.artikli[sifra]
	return a, ok, nil
}
func (f *fakePorRepo) KomKomercijalista(_ context.Context, sifra int) (domain.KomKomercijalista, bool, error) {
	return domain.KomKomercijalista{KomID: 5, Sifra: 7, Ime: "Ana Anić"}, sifra == 7, nil
}
func (f *fakePorRepo) KomKomercijalistaKorisnika(context.Context, int64) (domain.KomKomercijalista, bool, error) {
	return domain.KomKomercijalista{KomID: 5, Sifra: 7, Ime: "Ana Anić"}, true, nil
}
func (f *fakePorRepo) KomMagacin(_ context.Context, id int) (int, bool, error) {
	return id * 10, id > 0, nil
}
func (f *fakePorRepo) KomFirma(context.Context, string) (domain.RobnoStampaFakturaFirmaDto, error) {
	return domain.RobnoStampaFakturaFirmaDto{Naziv: "Firma"}, nil
}
func (f *fakePorRepo) BankeZiroRacuni(context.Context) ([]string, error) {
	return []string{"160-1  Banka"}, nil
}
func (f *fakePorRepo) RvrFdob(context.Context) (string, error) {
	return "4330", nil
}
func (f *fakePorRepo) NoviDokumentDefaults(context.Context, string) (*[]domain.SearchRowDto, error) {
	return &[]domain.SearchRowDto{{Key: "941"}}, nil
}
func (f *fakePorRepo) Valuta(_ context.Context, sifval int) (string, int, bool, error) {
	switch sifval {
	case 941:
		return "RSD", 1, true, nil
	case 978:
		return "EUR", 1, true, nil
	}
	return "", 0, false, nil
}
func (f *fakePorRepo) SrednjiKurs(context.Context, int, time.Time) (float64, bool, error) {
	return 0, false, nil
}
func (f *fakePorRepo) OrgjedPostoji(_ context.Context, id int64) (bool, error) {
	return id == 2, nil
}
func (f *fakePorRepo) MestotrPostoji(_ context.Context, id int64) (bool, error) {
	return id == 5 || id == 6, nil
}
func (f *fakePorRepo) MestotrUOrgjed(_ context.Context, mt, _ int64) (bool, error) {
	return mt == 5, nil
}

func newFakePorRepo() *fakePorRepo {
	dadok := sql.NullTime{Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Valid: true}
	return &fakePorRepo{
		porudzbenice: map[int64]domain.KomPorudzbenicaRow{
			25: {RdokID: 25, Vrd: 320, Dokum: 6, MagaciniID: 3, Mag: 30, Doksts: "10", Dadok: dadok, Fkto: "4330", Fana: "00012", SifVal: 978,
				OznVal: "EUR", Kurs: 117.25, VaziZa: 1, VrdokID: 15, KomIme: "Ana Anić", Posuslnab: "Avans 30%"},
			26: {RdokID: 26, Vrd: 320, Dokum: 7, MagaciniID: 3, Doksts: "40"},
			27: {RdokID: 27, Vrd: 320, Dokum: 8, MagaciniID: 3, Doksts: "10", SifVal: 941},
		},
		stavke: map[int64][]domain.KomPorStavkaRow{
			25: {{RproID: 41, Rbr: 1, Sifra: 1001, Naziv: "Vijak", Komercopis: "M8", JM: "kom", Kolicina: 3, Cena: 1172.5, CenaVal: 10, Rabat: 10},
				{RproID: 42, Rbr: 2, Sifra: 1002, Naziv: "Matica", JM: "kom", Kolicina: 1, Cena: 234.5, CenaVal: 2}},
			27: {{RproID: 51, Rbr: 1, Sifra: 1001, Cena: 0}},
		},
		veze: map[int64]domain.KomPorVeza{15: {RdokID: 15, Dokum: 4, Rok: 10, Otp: "Kamion", MagaciniID: 4, Orgjed: sql.NullInt64{Int64: 2, Valid: true},
			Fkto: "4330", Fana: "00012"}},
		izvori: map[int64][]domain.KomPorIzvor{15: {{RproID: 31, Sifra: 1001, Po: 2, RsifID: 44, Kolicina: 3, CenaVal: 0},
			{RproID: 32, Sifra: 1002, Po: 2, RsifID: 45, Kolicina: 5, CenaVal: 2}}},
		dobavljaci: map[string]domain.KomDobavljacPodaci{"4330/00012": {Naziv: "Metal", TipPdv: 1}},
		artikli:    map[int]domain.KomArtikal{1001: {RsifID: 44, Sifra: 1001, Naziv: "Vijak", JM: "kom", Po: 2, NCena: 12.5}},
		noviRbr:    3,
	}
}

// TestGetPorudzbenica checks a new porudžbenica (the legacy BTN_NoviDOK: the next broj of the vrsta in the magacin,
// the konto of the dobavljači, the valuta of the firma), a new one of a zahtev za ponudu (its data and magacin), a
// saved one (a foreign valuta, the slobodne stavke of its zahtev za ponudu) and a confirmed one (a new one with why).
func TestGetPorudzbenica(t *testing.T) {
	danas := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	t.Run("new", func(t *testing.T) {
		repo := newFakePorRepo()
		s := NewPorudzbenicaService(repo, fakeMagacini{})
		s.now = func() time.Time { return danas }
		dok, err := s.GetPorudzbenica(komCtx(), 0, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		want := domain.KomPorHeaderView{Vrd: 320, Dokum: 7, Datum: danas, MagaciniID: 3, Fkto: "4330", KomSifra: 7, KomIme: "Ana Anić", SifVal: 941,
			OznVal: "RSD", Kurs: 1, VaziZa: 1}
		if !reflect.DeepEqual(dok.Header, want) || !reflect.DeepEqual(repo.noviBroj, [][2]int{{320, 30}}) || dok.SifValFirme != 941 {
			t.Errorf("header\n got %+v\nwant %+v (novi broj %v)", dok.Header, want, repo.noviBroj)
		}
	})
	t.Run("new of a zahtev za ponudu", func(t *testing.T) {
		s := NewPorudzbenicaService(newFakePorRepo(), fakeMagacini{})
		dok, err := s.GetPorudzbenica(komCtx(), 0, 0, 3, 15)
		if err != nil {
			t.Fatal(err)
		}
		h := dok.Header
		if h.VrdokID != 15 || h.VezaDokum != 4 || h.Rok != 10 || h.Otp != "Kamion" || h.MagaciniID != 4 || h.Fana != "00012" ||
			h.Dobavljac.Naziv != "Metal" || h.Dobavljac.SistemPdv != "U sistemu PDV-a" || !h.Orgjed.Valid {
			t.Errorf("header %+v", h)
		}
	})
	t.Run("saved in a foreign valuta", func(t *testing.T) {
		s := NewPorudzbenicaService(newFakePorRepo(), fakeMagacini{})
		dok, err := s.GetPorudzbenica(komCtx(), 25, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !dok.Header.Ino || len(dok.Izvori) != 2 || !dok.ImaStavke || dok.NoviRbr != 3 || dok.Poruka != "" {
			t.Errorf("dok %+v", dok)
		}
	})
	t.Run("confirmed", func(t *testing.T) {
		s := NewPorudzbenicaService(newFakePorRepo(), fakeMagacini{})
		dok, err := s.GetPorudzbenica(komCtx(), 26, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if dok.Header.RdokID != 0 || dok.Poruka != komMsgPorNijeIzmena {
			t.Errorf("rdokid %d, poruka %q", dok.Header.RdokID, dok.Poruka)
		}
	})
}

// TestValidatePorudzbenica checks the legacy ValidacijaEDOK.
func TestValidatePorudzbenica(t *testing.T) {
	s := NewPorudzbenicaService(newFakePorRepo(), fakeMagacini{})
	ok := domain.KomPorParams{Vrd: "320", MagaciniID: "3", Datum: "2026-03-02", Fkto: "4330", Fana: "00012", Orgjed: "2", MestoTroska: "5",
		SifVal: "941", Kurs: "1", VaziZa: "1"}
	for _, tt := range []struct {
		name   string
		change func(p *domain.KomPorParams)
		want   []string
	}{
		{name: "valid", change: func(*domain.KomPorParams) {}},
		{name: "no konto", change: func(p *domain.KomPorParams) { p.Fkto = "" }, want: []string{"fkto: " + komMsgPorKonto}},
		{name: "no šifra", change: func(p *domain.KomPorParams) { p.Fana = " " }, want: []string{"fana: " + komMsgPorSifra}},
		{name: "unknown dobavljač", change: func(p *domain.KomPorParams) { p.Fana = "99999" }, want: []string{"fana: " + komMsgPorDobavljac}},
		{name: "no org. jedinica and no mesto troška (\"-\")", change: func(p *domain.KomPorParams) { p.Orgjed = "-"; p.MestoTroska = "-" }},
		{name: "mesto troška of another org. jedinica", change: func(p *domain.KomPorParams) { p.MestoTroska = "6" },
			want: []string{"mestotrid: " + komMsgMestoTroskaOJ}},
		{name: "unknown valuta", change: func(p *domain.KomPorParams) { p.SifVal = "5" }, want: []string{"valutaid: " + komMsgPorValuta}},
		{name: "foreign valuta without kurs", change: func(p *domain.KomPorParams) { p.SifVal = "978"; p.Kurs = "0" }, want: []string{"kurs: " + komMsgPorKurs}},
		{name: "unknown zahtev za ponudu", change: func(p *domain.KomPorParams) { p.VrdokID = "77" }, want: []string{"vezabroj: " + komMsgPorVeza}},
		{name: "saved confirmed one", change: func(p *domain.KomPorParams) { p.RdokID = 26 }, want: []string{"brdok: " + komMsgPorNijeIzmena}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ok
			tt.change(&p)
			if got := komPolja(s.ValidatePorudzbenica(komCtx(), p)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("errors\n got %v\nwant %v", got, tt.want)
			}
		})
	}
}

// TestSavePorudzbenica checks the save of a new porudžbenica (the next broj of the vrsta in its magacin, its zahtev za
// ponudu) and of a saved one (its vrsta, broj, magacin and zahtev za ponudu stay).
func TestSavePorudzbenica(t *testing.T) {
	repo := newFakePorRepo()
	s := NewPorudzbenicaService(repo, fakeMagacini{})
	params := domain.KomPorParams{Vrd: "320", MagaciniID: "3", Datum: "2026-03-02", Rok: "15", VrdokID: "15", Fkto: " 4330", Fana: "00012",
		KomSifra: "7", Orgjed: "2", MestoTroska: "5", SifVal: "978", Kurs: "117,25", VaziZa: "1", Otp: "Kamion", Pratdok: "Atest"}
	if id, err := s.SavePorudzbenica(komCtx(), params); err != nil || id != 99 {
		t.Fatalf("id %d, err %v", id, err)
	}
	p := repo.dodat[0]
	if p.Mag != 30 || p.Dokum != 7 || p.VrdokID != 15 || p.Fkto != "4330" || p.Kurs != 117.25 || !p.KomID.Valid || p.Pratdok != "Atest" {
		t.Errorf("new %+v", p)
	}
	params.RdokID, params.Vrd, params.MagaciniID, params.VrdokID = 25, "999", "4", "16"
	if _, err := s.SavePorudzbenica(komCtx(), params); err != nil {
		t.Fatal(err)
	}
	p = repo.izmenjen[0]
	if p.Vrd != 320 || p.Dokum != 6 || p.MagaciniID != 3 || p.VrdokID != 15 {
		t.Errorf("saved %+v", p)
	}
}

// TestPorudzbenicaPotvrdaBrisanje checks the potvrda (stavke with a cena, the legacy BTN_NoviDOK1) and the deletion
// (not a confirmed one, the legacy BTN_Brisanje).
func TestPorudzbenicaPotvrdaBrisanje(t *testing.T) {
	repo := newFakePorRepo()
	s := NewPorudzbenicaService(repo, fakeMagacini{})
	if poruka, _ := s.PotvrdiPorudzbenicu(komCtx(), 27); poruka != komMsgPorCenaNula {
		t.Errorf("cena 0: %q", poruka)
	}
	if poruka, _ := s.PotvrdiPorudzbenicu(komCtx(), 25); poruka != komMsgPotvrdjen || !reflect.DeepEqual(repo.potvrdjen, []int64{25}) {
		t.Errorf("confirmed: %q %v", poruka, repo.potvrdjen)
	}
	if poruka, _ := s.DeletePorudzbenica(komCtx(), 26); poruka != komMsgPorPotvrdjenBris {
		t.Errorf("delete confirmed: %q", poruka)
	}
	if poruka, err := s.DeletePorudzbenica(komCtx(), 25); err != nil || poruka != "" || !reflect.DeepEqual(repo.obrisan, []int64{25}) {
		t.Errorf("delete: %q %v %v", poruka, err, repo.obrisan)
	}
}

// TestPorudzbenicaStavke checks the stavke: a new one in a foreign valuta (the cena converted by the kurs, the legacy
// AddEPRO), the change of a stavka, the validation, the stavke taken from the zahtev za ponudu (rezervisane, rproid1)
// and the deletion of the last one.
func TestPorudzbenicaStavke(t *testing.T) {
	repo := newFakePorRepo()
	s := NewPorudzbenicaService(repo, fakeMagacini{})
	if got := komPolja(s.ValidateStavka(komCtx(), domain.KomPorStavkaParams{RdokID: 25, Sifra: "1001", Kolicina: "1", Cena: "0"})); !reflect.DeepEqual(got,
		[]string{"cena: " + komMsgPorCena}) {
		t.Errorf("errors %v", got)
	}
	if err := s.SaveStavka(komCtx(), domain.KomPorStavkaParams{RdokID: 25, Sifra: "1001", Kolicina: "2", Cena: "10", Rabat: "5"}); err != nil {
		t.Fatal(err)
	}
	st := repo.stavkaNova[0]
	if st.Cena != 1172.5 || st.CenaVal != 10 || st.Rabat != 5 || st.Rbr != 3 || st.RproID1 != 0 || st.Mag != 30 {
		t.Errorf("new stavka %+v", st)
	}
	if err := s.SaveStavka(komCtx(), domain.KomPorStavkaParams{RdokID: 25, RproID: 41, Kolicina: "4", Cena: "2", Rabat: "0"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.stavkaIzm, [][4]float64{{4, 234.5, 2, 0}}) {
		t.Errorf("changed %v", repo.stavkaIzm)
	}
	if poruka, err := s.Preuzmi(komCtx(), 25, 0); err != nil || poruka != "" {
		t.Fatalf("preuzmi: %q %v", poruka, err)
	}
	if !reflect.DeepEqual(repo.rezervisan, []int64{31, 32}) || repo.stavkaNova[2].RproID1 != 32 || repo.stavkaNova[2].Cena != 234.5 {
		t.Errorf("rezervisani %v, stavke %+v", repo.rezervisan, repo.stavkaNova)
	}
	if poruka, _ := s.Preuzmi(komCtx(), 27, 0); poruka != komMsgPorNemaZaDodavanje {
		t.Errorf("without a zahtev za ponudu: %q", poruka)
	}
	if poruka, err := s.VratiPoslednju(komCtx(), 25); err != nil || poruka != "" || !reflect.DeepEqual(repo.stavkaBris, [][3]int64{{25, 42, 2}}) {
		t.Errorf("vrati poslednju: %q %v %v", poruka, err, repo.stavkaBris)
	}
	if poruka, _ := s.VratiSve(komCtx(), 99); poruka != komMsgPorNemaZaBrisanje {
		t.Errorf("vrati sve without stavke: %q", poruka)
	}
}

// TestPorudzbenicaStampa checks the print: the opis by the checkboxes, the cena in the valuta, the vrednost less the
// rabat, the total and the posebni uslovi nabavke (instead of the tekući računi only for a domestic one).
func TestPorudzbenicaStampa(t *testing.T) {
	s := NewPorudzbenicaService(newFakePorRepo(), fakeMagacini{})
	stampa, poruka, err := s.GetStampa(komCtx(), 25, true, true)
	if err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	if !stampa.Ino || stampa.Stavke[0].Opis != "Vijak M8" || stampa.Stavke[0].Cena != 10 || stampa.Stavke[0].Vrednost != 27 ||
		stampa.Ukupno != 29 || stampa.Dobavljac.Naziv != "Metal" || !reflect.DeepEqual(stampa.ZiroRacuni, []string{"160-1  Banka"}) ||
		stampa.Referent != "(Ana Anić)" {
		t.Errorf("stampa %+v", stampa)
	}
	if stampa, _, _ := s.GetStampa(komCtx(), 25, false, true); stampa.Stavke[1].Opis != "" {
		t.Errorf("opis without the naziv: %q", stampa.Stavke[1].Opis)
	}
	if _, poruka, _ := s.GetStampa(komCtx(), 26, true, false); poruka != komMsgBezStavkiStampa {
		t.Errorf("without stavke: %q", poruka)
	}
}

// TestKomVrednost checks the vrednost of a stavka: the rabat rounded to 2 decimals (the legacy ITEM_VRED).
func TestKomVrednost(t *testing.T) {
	// 3 × 10,333 = 30,999; the rabat 7,5 % = 2,324925, rounded to 2,32.
	if got := komVrednost(3, 10.333, 7.5); math.Abs(got-28.679) > 1e-9 {
		t.Errorf("vrednost %v", got)
	}
}

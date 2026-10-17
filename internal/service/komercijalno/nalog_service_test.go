package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// fakeNalogRepo is an in-memory nalRepository: the profaktura fake with the nalozi by rdokid, what they have besides
// the profaktura, the lots by age, the promet of the kupac and the banke; it records what was written.
type fakeNalogRepo struct {
	*fakeProRepo
	nalozi     map[int64]domain.KomProfakturaRow
	dodatno    map[int64]domain.KomNalogDodatno
	starost    []domain.KomNalogLot
	fpro       []domain.KomNalogFpro
	robne      []domain.KomNalogRobnaStavka
	neObveznik bool

	sacuvano []string
	kopirano []string
	vezano   []int64
	vraceno  []int64
}

func (f *fakeNalogRepo) KomNalog(_ context.Context, id int64) (domain.KomProfakturaRow, bool, error) {
	n, ok := f.nalozi[id]
	return n, ok, nil
}
func (f *fakeNalogRepo) KomNalogDodatno(_ context.Context, id int64) (domain.KomNalogDodatno, error) {
	return f.dodatno[id], nil
}
func (f *fakeNalogRepo) KomNalogDodatnoSacuvaj(_ context.Context, _ db.Transaction, rdokID, vrdokID int64, sifbank int) error {
	f.sacuvano = append(f.sacuvano, fmtArgs(rdokID, vrdokID, sifbank))
	return nil
}
func (f *fakeNalogRepo) KomNalogBanka(_ context.Context, id int64) (int, bool, error) {
	return 2, id == 4, nil
}
func (f *fakeNalogRepo) KomNalogKopirajStavke(_ context.Context, _ db.Transaction, proID, nalogID int64, vrd, dokum int, dadok sql.NullTime, rbrOd int) error {
	f.kopirano = append(f.kopirano, fmtArgs(proID, nalogID, vrd, dokum, dadok.Time.Format("2006-01-02"), rbrOd))
	return nil
}
func (f *fakeNalogRepo) KomNalogProfakturaVezi(_ context.Context, _ db.Transaction, proID int64) (bool, error) {
	f.vezano = append(f.vezano, proID)
	return true, nil
}
func (f *fakeNalogRepo) KomNalogProfakturaVrati(_ context.Context, _ db.Transaction, proID int64) (bool, error) {
	f.vraceno = append(f.vraceno, proID)
	return true, nil
}
func (f *fakeNalogRepo) KomNalogLotoviStarost(context.Context, int, string, int) ([]domain.KomNalogLot, error) {
	return f.starost, nil
}
func (f *fakeNalogRepo) KomNalogFpro(context.Context, string, string) ([]domain.KomNalogFpro, error) {
	return f.fpro, nil
}
func (f *fakeNalogRepo) KomNalogRobneStavke(context.Context, string, string) ([]domain.KomNalogRobnaStavka, error) {
	return f.robne, nil
}
func (f *fakeNalogRepo) KomNalogPrikaz(context.Context, domain.KomNalogPrikazParams) ([]domain.KomNalogPrikazRow, error) {
	return nil, nil
}
func (f *fakeNalogRepo) KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error) {
	fi, err := f.fakeProRepo.KomFirma(ctx, firma)
	fi.Obv = !f.neObveznik
	return fi, err
}

// fmtArgs is a compact record of the arguments of a call.
func fmtArgs(v ...any) string {
	return fmt.Sprintln(v...)
}

func newFakeNalogRepo() *fakeNalogRepo {
	pro := newFakeProRepo()
	// 46 is a potvrđena profaktura of the magacin 3 (its stavke are the ones of 45).
	pro.stavke[46] = pro.stavke[45]
	pro.stavke[50] = pro.stavke[45]
	return &fakeNalogRepo{fakeProRepo: pro,
		nalozi: map[int64]domain.KomProfakturaRow{
			50: {RdokID: 50, Vrd: 350, Dokum: 3, Dadok: proDatum, MagaciniID: 3, Mag: 30, Fkto: "2040", Fana: "00031", Doksts: "10", SifVal: 941,
				Kurs: 1, VaziZa: 1, Tipzal: 2, Tipmag: "V", MagacinText: "30-Roba", OrgjedText: "1 - Prodaja", Posuslnab: "160-1", CenovnikID: 0},
			51: {RdokID: 51, Vrd: 350, Dokum: 4, MagaciniID: 3, Mag: 30, Doksts: "20"},
			52: {RdokID: 52, Vrd: 350, Dokum: 5, MagaciniID: 3, Mag: 30, Doksts: "40"},
		},
		dodatno: map[int64]domain.KomNalogDodatno{50: {VrdokID: 46}},
	}
}

// TestKomNalogSaldo checks the legacy SaldoKUPCA: the uplate close the dospelo first, the robne fakture with the rabat,
// the ugovoreni rabat, the kasa and the PDV split by their dospeće.
func TestKomNalogSaldo(t *testing.T) {
	danas := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	d := func(y, m, dan int) sql.NullTime {
		return sql.NullTime{Time: time.Date(y, time.Month(m), dan, 0, 0, 0, 0, time.UTC), Valid: true}
	}
	fpro := []domain.KomNalogFpro{
		{Kat: 1, Iznos: 1000, Dadok: d(2026, 2, 1), Rok: 30}, // dospela 3.3.
		{Kat: 2, Iznos: 500, Dadok: d(2026, 3, 5), Rok: 10},  // dospeva 15.3.
		{Kat: 3, Iznos: 1200, Dadok: d(2026, 3, 1)},
		{Kat: 5, Iznos: 99},
	}
	robne := []domain.KomNalogRobnaStavka{
		// 2 × 100 = 200, rabat 10% 20, ugovoreni 10% 18, kasa 2% 3.24, PDV 20% of 200 = 40: 198.76, dospela.
		{RdokID: 1, Dadok: d(2026, 3, 1), Rok: 5, Ugrabat: 10, Pkase: 2, Kolic: 2, Fcena: 100, Rab: 10, Stopa: 20},
		// 1 × 50 + PDV 10: 60, dospeva 20.3.
		{RdokID: 2, Dadok: d(2026, 3, 10), Rok: 10, Kolic: 1, Fcena: 50, Stopa: 20},
	}
	got := komNalogSaldo(fpro, robne, danas)
	want := domain.KomNalogSaldo{Realizacija: 1500, Placeno: 1200, Saldo: 300, Dospelo: 0, Nedospelo: 300, RealizacijaRob: 258.76,
		DospeloRob: 198.76, NedospeloRob: 60, UkDospelo: 198.76, UkNedospelo: 360, UkSaldo: 558.76}
	if !komBlizu(got, want) {
		t.Errorf("saldo\n got %+v\nwant %+v", got, want)
	}
	if got := komNalogSaldo([]domain.KomNalogFpro{{Kat: 1, Iznos: 1000, Dadok: d(2026, 2, 1)}, {Kat: 3, Iznos: 400}}, nil, danas); got.Dospelo != 600 ||
		got.Nedospelo != 0 {
		t.Errorf("dospelo %v nedospelo %v", got.Dospelo, got.Nedospelo)
	}
}

// komBlizu compares two salda to the cent.
func komBlizu(a, b domain.KomNalogSaldo) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	for i := 0; i < va.NumField(); i++ {
		if komRound2(va.Field(i).Float()) != komRound2(vb.Field(i).Float()) {
			return false
		}
	}
	return true
}

// TestKomNalogPoStarosti checks the split of a stavka over the lots by their age (the legacy CBOX_ZALSTAROST).
func TestKomNalogPoStarosti(t *testing.T) {
	st := domain.KomProStavkaUpis{Sifra: 1001, Konto: "1320", Otk: "x", Kolicina: 10}
	lotovi := []domain.KomNalogLot{{Konto: "1320", Otk: "001", Serija: "A", Roktr: 100, Slobodno: 4}, {Konto: "1320", Otk: "002", Serija: "B", Roktr: 200, Slobodno: 3}}
	got := komNalogPoStarosti(st, lotovi)
	want := []domain.KomProStavkaUpis{
		{Sifra: 1001, Konto: "1320", Otk: "001", Serija: "A", Roktr: 100, Kolicina: 4},
		{Sifra: 1001, Konto: "1320", Otk: "002", Serija: "B", Roktr: 200, Kolicina: 3},
		{Sifra: 1001, Konto: "1320", Kolicina: 3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("split\n got %+v\nwant %+v", got, want)
	}
	if got := komNalogPoStarosti(domain.KomProStavkaUpis{Kolicina: 2}, lotovi); len(got) != 1 || got[0].Otk != "001" || got[0].Kolicina != 2 {
		t.Errorf("one lot: %+v", got)
	}
	if got := komNalogPoStarosti(st, nil); len(got) != 1 || got[0].Otk != "" || got[0].Kolicina != 10 {
		t.Errorf("without lots: %+v", got)
	}
}

// TestKomRezervisi checks the reservation of a lot with a row of drsta and rsta, of a lot without one (a new row of
// drsta) and the release (no new row, rsta not below 0).
func TestKomRezervisi(t *testing.T) {
	repo := newFakeProRepo()
	ctx := komCtx()
	lot := domain.KomNalogRezervacija{MagaciniID: 3, Mag: 30, RsifID: 44, Sifra: 1001, Konto: " 1320 ", Otk: "001", Serija: "S1", Roktr: 82000, Cena: 80,
		Kolicina: 2}
	bezLota := domain.KomNalogRezervacija{MagaciniID: 3, Mag: 30, RsifID: 45, Sifra: 1002, Konto: "1320", Cena: 50, Kolicina: 3}
	for _, r := range []domain.KomNalogRezervacija{lot, bezLota, func() domain.KomNalogRezervacija { r := lot; r.Kolicina = -1; return r }(),
		func() domain.KomNalogRezervacija { r := bezLota; r.Kolicina = -3; return r }(), {Kolicina: 0}} {
		if err := komRezervisi(ctx, repo, nil, r); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"drsta 7 rez 2 nn=false", "rsta 9 rez 2 nn=false", "novi drsta 1320/1002 //0 rsif 45 rez 3", "drsta 7 rez -1 nn=false",
		"rsta 9 rez -1 nn=true"}
	if !reflect.DeepEqual(repo.rez, want) {
		t.Errorf("rez\n got %q\nwant %q", repo.rez, want)
	}
	if v := komNoviLot(bezLota, "1320"); v["mdug_12"] != 0 || v["miz_1"] != 0 || v["magaciniid"] != 3 || v["cena"] != 50.0 {
		t.Errorf("novi lot %v", v)
	}
}

// TestValidateNalog checks the legacy ValidacijaEDOK of a nalog: the kurs, the poreska kategorija of a firm that is not
// a PDV obveznik, the banka and the profaktura.
func TestValidateNalog(t *testing.T) {
	valid := domain.KomNalogParams{KomProParams: domain.KomProParams{Vrd: "350", Tipdok: "43", MagaciniID: "3", Datum: "2026-03-09", Fkto: "2040",
		Fana: "00031", Orgjed: "2", MestoTroska: "5", SifVal: "941", Kurs: "1", VaziZa: "1"}, ProfakturaID: "46", BnkizvID: "4"}
	for _, tt := range []struct {
		name   string
		repo   func(r *fakeNalogRepo)
		change func(p *domain.KomNalogParams)
		want   []string
	}{
		{name: "valid", change: func(*domain.KomNalogParams) {}},
		{name: "kurs 0 of the domicilna valuta", change: func(p *domain.KomNalogParams) { p.Kurs = "0" }, want: []string{"kurs"}},
		{name: "without the poreska kategorija of a firm not in PDV", repo: func(r *fakeNalogRepo) { r.neObveznik = true },
			change: func(*domain.KomNalogParams) {}, want: []string{"keykat"}},
		{name: "with the poreska kategorija of a firm not in PDV", repo: func(r *fakeNalogRepo) { r.neObveznik = true },
			change: func(p *domain.KomNalogParams) { p.Keykat = "PDV-RS-33" }},
		{name: "a banka that does not exist", change: func(p *domain.KomNalogParams) { p.BnkizvID = "5" }, want: []string{"bnkizvid"}},
		{name: "a new profaktura", change: func(p *domain.KomNalogParams) { p.ProfakturaID = "45" }, want: []string{"profakturaid"}},
		{name: "a profaktura of another magacin", change: func(p *domain.KomNalogParams) { p.MagaciniID = "4" }, want: []string{"profakturaid"}},
		{name: "a saved nalog keeps its profaktura", change: func(p *domain.KomNalogParams) { p.RdokID, p.ProfakturaID = 50, "45" }},
		{name: "a potvrđen nalog", change: func(p *domain.KomNalogParams) { p.RdokID = 51 }, want: []string{"datum"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeNalogRepo()
			if tt.repo != nil {
				tt.repo(repo)
			}
			s := NewNalogService(repo, fakeMagacini{})
			p := valid
			tt.change(&p)
			var got []string
			for _, e := range s.ValidateNalog(komCtx(), p) {
				got = append(got, e.Field)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSaveNalog checks the legacy SnimiEDOK of a new nalog made of a profaktura (KopirajStavkeizProfakture: the link,
// the stavke from the redni broj 1, the reservation of each of them) and of a saved one that has its profaktura.
func TestSaveNalog(t *testing.T) {
	repo := newFakeNalogRepo()
	s := NewNalogService(repo, fakeMagacini{})
	params := domain.KomNalogParams{KomProParams: domain.KomProParams{Vrd: "350", Tipdok: "43", MagaciniID: "3", Datum: "2026-03-10", Fkto: "2040",
		Fana: "00031", SifVal: "941", Kurs: "1", VaziZa: "1"}, ProfakturaID: "46", BnkizvID: "4"}
	id, err := s.SaveNalog(komCtx(), params)
	if err != nil || id != 99 {
		t.Fatalf("id %d err %v", id, err)
	}
	if len(repo.dodat) != 1 || repo.dodat[0].Opis != "Nalog za izdavanje robe br: 8" {
		t.Errorf("dodat %+v", repo.dodat)
	}
	if !reflect.DeepEqual(repo.sacuvano, []string{fmtArgs(int64(99), int64(46), 2)}) || !reflect.DeepEqual(repo.vezano, []int64{46}) ||
		!reflect.DeepEqual(repo.kopirano, []string{fmtArgs(int64(46), int64(99), 350, 8, "2026-03-10", 1)}) {
		t.Errorf("sačuvano %v vezano %v kopirano %v", repo.sacuvano, repo.vezano, repo.kopirano)
	}
	wantRez := []string{"drsta 7 rez 10 nn=false", "rsta 9 rez 10 nn=false", "novi drsta 1320/1002 //0 rsif 45 rez 2"}
	if !reflect.DeepEqual(repo.rez, wantRez) || repo.commits != 1 {
		t.Errorf("rez %q commits %d", repo.rez, repo.commits)
	}

	repo = newFakeNalogRepo()
	s = NewNalogService(repo, fakeMagacini{})
	params.RdokID, params.ProfakturaID, params.BnkizvID = 50, "45", ""
	if _, err := s.SaveNalog(komCtx(), params); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.sacuvano, []string{fmtArgs(int64(50), int64(46), 0)}) || len(repo.kopirano) != 0 || len(repo.vezano) != 0 ||
		len(repo.izmenjen) != 1 {
		t.Errorf("sačuvano %v kopirano %v vezano %v izmenjen %d", repo.sacuvano, repo.kopirano, repo.vezano, len(repo.izmenjen))
	}
}

// TestDeleteNalog checks the legacy BTN_Brisanje: only a new nalog, its zalihe released and its profaktura potvrđena
// again.
func TestDeleteNalog(t *testing.T) {
	repo := newFakeNalogRepo()
	s := NewNalogService(repo, fakeMagacini{})
	for id, want := range map[int64]string{51: komMsgNalPotvrdjen, 52: komMsgProNalog, 60: komMsgNalBrisanjeNema} {
		if poruka, err := s.DeleteNalog(komCtx(), id); err != nil || poruka != want {
			t.Errorf("%d: %q %v", id, poruka, err)
		}
	}
	if poruka, err := s.DeleteNalog(komCtx(), 50); err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	wantRez := []string{"drsta 7 rez -10 nn=false", "rsta 9 rez -10 nn=true"}
	if !reflect.DeepEqual(repo.rez, wantRez) || !reflect.DeepEqual(repo.vraceno, []int64{46}) || repo.commits != 1 {
		t.Errorf("rez %q vraćeno %v commits %d", repo.rez, repo.vraceno, repo.commits)
	}
}

// TestPotvrdiNalog checks the legacy BTN_Potvrda: the stavke, their cene and the status.
func TestPotvrdiNalog(t *testing.T) {
	repo := newFakeNalogRepo()
	s := NewNalogService(repo, fakeMagacini{})
	if poruka, err := s.PotvrdiNalog(komCtx(), 50); err != nil || poruka != komMsgPotvrdjen {
		t.Errorf("potvrda %q %v", poruka, err)
	}
	repo.stavke[51] = repo.stavke[45]
	if poruka, _ := s.ProveriPotvrdu(komCtx(), 51); poruka != komMsgVecPotvrdjen {
		t.Errorf("potvrđen %q", poruka)
	}
	repo.stavke[50] = []domain.KomProStavkaRow{{RproID: 1, Fcena: 10, Cena: 0}}
	if poruka, _ := s.ProveriPotvrdu(komCtx(), 50); poruka != komMsgNalCenaNula {
		t.Errorf("cena 0 %q", poruka)
	}
	repo.stavke[50] = nil
	if poruka, _ := s.ProveriPotvrdu(komCtx(), 50); poruka != komMsgBezStavkiPotvrda {
		t.Errorf("bez stavki %q", poruka)
	}
}

// TestSaveNalogStavka checks the legacy AddEPRO with "Skidaj zalihe po starosti" (one stavka per lot, the redni brojevi
// in order, each reserved), the change of a stavka (the reservation of the difference) and its delete (released).
func TestSaveNalogStavka(t *testing.T) {
	repo := newFakeNalogRepo()
	repo.starost = []domain.KomNalogLot{{Konto: "1320", Otk: "001", Serija: "S1", Roktr: 82000, Slobodno: 4}}
	s := NewNalogService(repo, fakeMagacini{})
	if err := s.SaveNalogStavka(komCtx(), domain.KomNalogStavkaParams{KomProStavkaParams: domain.KomProStavkaParams{RdokID: 50, Sifra: "1001",
		Konto: "1320", Kolicina: "6", Cena: "100", Rabat: "5"}, PoStarosti: "true"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.stavkaNova) != 2 || repo.stavkaNova[0].Rbr != 3 || repo.stavkaNova[0].Kolicina != 4 || repo.stavkaNova[0].Otk != "001" ||
		repo.stavkaNova[1].Rbr != 4 || repo.stavkaNova[1].Kolicina != 2 || repo.stavkaNova[1].Otk != "" || repo.stavkaNova[1].Cena != 100 {
		t.Errorf("stavke %+v", repo.stavkaNova)
	}
	wantRez := []string{"drsta 7 rez 4 nn=false", "rsta 9 rez 4 nn=false", "novi drsta 1320/1001 //0 rsif 44 rez 2", "rsta 9 rez 2 nn=false"}
	if !reflect.DeepEqual(repo.rez, wantRez) || len(repo.ukupno) != 1 {
		t.Errorf("rez %q ukupno %v", repo.rez, repo.ukupno)
	}

	repo.rez = nil
	if err := s.SaveNalogStavka(komCtx(), domain.KomNalogStavkaParams{KomProStavkaParams: domain.KomProStavkaParams{RdokID: 50, RproID: 61,
		Kolicina: "12", Cena: "100", Rabat: "5"}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.rez, []string{"drsta 7 rez 2 nn=false", "rsta 9 rez 2 nn=false"}) || len(repo.stavkaIzm) != 1 {
		t.Errorf("izmena rez %q izm %+v", repo.rez, repo.stavkaIzm)
	}

	repo.rez = nil
	if poruka, err := s.DeleteNalogStavka(komCtx(), 50, 61); err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	if !reflect.DeepEqual(repo.rez, []string{"drsta 7 rez -10 nn=false", "rsta 9 rez -10 nn=true"}) || len(repo.stavkaBris) != 1 {
		t.Errorf("brisanje rez %q bris %v", repo.rez, repo.stavkaBris)
	}
	if poruka, _ := s.DeleteNalogStavka(komCtx(), 51, 61); poruka != komMsgNalPotvrdjen {
		t.Errorf("potvrđen %q", poruka)
	}
}

// TestValidateNalogStavka checks the planska cena of a new stavka in a magacin of the tip G.
func TestValidateNalogStavka(t *testing.T) {
	repo := newFakeNalogRepo()
	n := repo.nalozi[50]
	n.Tipmag = "G"
	repo.nalozi[50] = n
	s := NewNalogService(repo, fakeMagacini{})
	params := domain.KomNalogStavkaParams{KomProStavkaParams: domain.KomProStavkaParams{RdokID: 50, Sifra: "1001", Kolicina: "1", Cena: "10"}}
	if errs := s.ValidateNalogStavka(komCtx(), params); len(errs) != 1 || errs[0].ErrorMessage != komMsgNalPlanskaCena {
		t.Errorf("bez planske cene %+v", errs)
	}
	repo.planska = 70
	if errs := s.ValidateNalogStavka(komCtx(), params); len(errs) != 0 {
		t.Errorf("sa planskom cenom %+v", errs)
	}
}

// TestNalogStampa checks the print: the opis with the lot, the težina and the zapremina, the status.
func TestNalogStampa(t *testing.T) {
	repo := newFakeNalogRepo()
	repo.stavke[50] = []domain.KomProStavkaRow{{RproID: 1, Sifra: 1001, Naziv: "Brašno", Serija: "S1", Roktr: komWinDevDatum("2026-12-31"),
		Kolicina: 4, Tezjm: 1.5, Kubikaza: 0.25}}
	s := NewNalogService(repo, fakeMagacini{})
	st, poruka, err := s.GetNalogStampa(komCtx(), 50)
	if err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	if st.Tezina != 6 || st.Zapremina != 1 || !st.Nepotvrdjen || !st.Lot || st.Dokum != 3 || st.Kupac.Naziv != "Kupac" ||
		st.Stavke[0].Opis != "Brašno\nSerija: S1 Rok trajanja: 31.12.2026" {
		t.Errorf("štampa %+v", st)
	}
	repo.stavke[51] = nil
	if _, poruka, _ := s.GetNalogStampa(komCtx(), 51); poruka != komMsgBezStavkiStampa {
		t.Errorf("bez stavki %q", poruka)
	}
}

// TestNalogPrikazRows checks the rows of the lists: "Izmeni" only of a new nalog, the status text, the red rows of the
// delimično realizovani with the vrednost of the isporučena količina.
func TestNalogPrikazRows(t *testing.T) {
	rows := NalogPrikazTableRows([]domain.KomNalogPrikazRow{{RdokID: 1, Doksts: "10"}, {RdokID: 2, Doksts: "40", Otpremnica: 7}}, true)
	if !rows[0].HasUpdate || rows[1].HasUpdate || rows[1].Fields[4] != "7" || rows[0].Fields[4] != "" || rows[1].Fields[11] != KomNalogStatusOpis("40") {
		t.Errorf("rows %+v", rows)
	}
	stavke := []domain.KomProStavkaRow{{RproID: 1, Kolicina: 5, Isporuceno: 4, Fcena: 10, Rabat: 10, Decimale: 2}, {RproID: 2, Kolicina: 1, Isporuceno: 1, Fcena: 5}}
	rows, ukupno := NalogPrikazStavkeTableRows(stavke, true)
	if ukupno != 41 || rows[0].ClassRow != common.ClassRowRazlika || rows[1].ClassRow != "" {
		t.Errorf("delimični %v %+v", ukupno, rows)
	}
	if _, ukupno := NalogPrikazStavkeTableRows(stavke, false); ukupno != 50 {
		t.Errorf("naručeno %v", ukupno)
	}
}

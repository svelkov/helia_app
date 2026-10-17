package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

// fakeOtpRepo is an in-memory otpRepository: the profaktura fake (the valuta, the kupac, the firma, the stanje of the
// reservations) with the otpremnice, their nalozi and stavke, the lots and the vrste; it records what was written.
type fakeOtpRepo struct {
	*fakeProRepo
	vrste      []domain.KomOtpVrsta
	otpremnice map[int64]domain.KomOtpremnicaRow
	nalozi     map[int64]domain.KomOtpNalog
	otpStavke  map[int64][]domain.KomOtpStavkaRow
	zalihe     map[string]repository.DrstaLot
	popdv      map[string]repository.PopdvUlaz
	brotpIma   bool
	rdokIma    bool
	zatvorena  bool
	rnalIma    bool
	prosnc     float64

	zapis []string
}

func (f *fakeOtpRepo) log(v ...any) { f.zapis = append(f.zapis, strings.TrimSpace(fmt.Sprintln(v...))) }

func (f *fakeOtpRepo) KomOtpVrste(context.Context) ([]domain.KomOtpVrsta, error) { return f.vrste, nil }
func (f *fakeOtpRepo) KomOtpremnica(_ context.Context, id int64) (domain.KomOtpremnicaRow, bool, error) {
	o, ok := f.otpremnice[id]
	return o, ok, nil
}
func (f *fakeOtpRepo) KomOtpNalog(_ context.Context, id int64) (domain.KomOtpNalog, bool, error) {
	n, ok := f.nalozi[id]
	return n, ok, nil
}
func (f *fakeOtpRepo) KomOtpStavke(_ context.Context, id int64) ([]domain.KomOtpStavkaRow, error) {
	return f.otpStavke[id], nil
}
func (f *fakeOtpRepo) KomOtpPrivremeniBrojevi(context.Context, db.Transaction, int, int) (int, int, error) {
	return 990002, 990003, nil
}
func (f *fakeOtpRepo) KomOtpBrotpPostoji(context.Context, int, int, int, int64) (bool, error) {
	return f.brotpIma, nil
}
func (f *fakeOtpRepo) KomOtpRdokPostoji(context.Context, db.Transaction, int, int, int) (bool, error) {
	return f.rdokIma, nil
}
func (f *fakeOtpRepo) KomOtpDodaj(_ context.Context, _ db.Transaction, p domain.KomOtpUpis) (int64, error) {
	f.log("dodaj", p.NalogID, p.Vrd, p.Brotp, p.Dokum, p.Napomena)
	return 61, nil
}
func (f *fakeOtpRepo) KomOtpIzmeni(_ context.Context, _ db.Transaction, p domain.KomOtpUpis) error {
	f.log("izmeni", p.RdokID, p.Brotp, p.Dokum, p.Kurs)
	return nil
}
func (f *fakeOtpRepo) KomOtpStavkeKopiraj(_ context.Context, _ db.Transaction, otpID, nalogID int64, storno, preracun bool) error {
	f.log("kopiraj", otpID, nalogID, storno, preracun)
	return nil
}
func (f *fakeOtpRepo) KomOtpStavkePreracunaj(_ context.Context, _ db.Transaction, otpID int64) error {
	f.log("preracunaj", otpID)
	return nil
}
func (f *fakeOtpRepo) KomOtpProdvred(_ context.Context, _ db.Transaction, otpID int64, prodvred float64) error {
	f.log("prodvred", otpID, prodvred)
	return nil
}
func (f *fakeOtpRepo) KomOtpStavkaIzmeni(_ context.Context, _ db.Transaction, otpID, rproID int64, kolic float64, otk, serija string, roktr int64) error {
	f.log("stavka", otpID, rproID, kolic, otk, serija, roktr)
	return nil
}
func (f *fakeOtpRepo) KomOtpNalogFormiran(_ context.Context, _ db.Transaction, nalogID int64) (bool, error) {
	f.log("nalog 40", nalogID)
	return true, nil
}
func (f *fakeOtpRepo) KomOtpNalogVrati(_ context.Context, _ db.Transaction, nalogID int64) (bool, error) {
	f.log("nalog 20", nalogID)
	return true, nil
}
func (f *fakeOtpRepo) KomOtpPrikaz(context.Context, domain.KomOtpPrikazParams) ([]domain.KomOtpPrikazRow, error) {
	return nil, nil
}
func (f *fakeOtpRepo) KomOtpGodinaZatvorena(context.Context) (bool, error) { return f.zatvorena, nil }
func (f *fakeOtpRepo) KomOtpNoviBrojevi(context.Context, db.Transaction, int, int) (int, int, error) {
	return 15, 101, nil
}
func (f *fakeOtpRepo) KomOtpRstaProsnc(context.Context, db.Transaction, int, int) (float64, error) {
	return f.prosnc, nil
}
func (f *fakeOtpRepo) KomOtpNetoTezina(_ context.Context, _ db.Transaction, otpID int64, tezina float64) error {
	f.log("tezina", otpID, tezina)
	return nil
}
func (f *fakeOtpRepo) KomOtpRnal(context.Context, db.Transaction, string, sql.NullTime, int) (int64, bool, error) {
	return 7, f.rnalIma, nil
}
func (f *fakeOtpRepo) KomOtpRnalDodaj(_ context.Context, _ db.Transaction, n domain.KomOtpRnalUpis) (int64, error) {
	f.log("rnal", n.Tipdok, n.MagaciniID, n.Mag)
	return 8, nil
}
func (f *fakeOtpRepo) KomOtpRnalDokument(_ context.Context, _ db.Transaction, rnalID int64) error {
	f.log("brdo", rnalID)
	return nil
}
func (f *fakeOtpRepo) KomOtpRdokDodaj(_ context.Context, _ db.Transaction, otpID, rnalID int64, d domain.KomOtpRdokUpis) (int64, error) {
	f.log("rdok", otpID, rnalID, d.Dokum, d.Dokiz, d.Opis, d.DescriptionCode)
	return 90, nil
}
func (f *fakeOtpRepo) KomOtpRproDodaj(_ context.Context, _ db.Transaction, p domain.KomOtpRproUpis) (int64, error) {
	f.log("rpro", p.RnalID, p.RdokID, p.Sifra, p.Kolic, p.Cena, p.Ncena, p.Iznos, p.Taxcat)
	return 91, nil
}
func (f *fakeOtpRepo) KomOtpRceneDodaj(_ context.Context, _ db.Transaction, c domain.KomOtpRceneUpis) error {
	f.log("rcene", c.Sifra, c.Ncena, c.Cena, c.Fcena)
	return nil
}
func (f *fakeOtpRepo) KomOtpRdokBrst(_ context.Context, _ db.Transaction, rdokID int64, brst int) error {
	f.log("brst", rdokID, brst)
	return nil
}
func (f *fakeOtpRepo) KomOtpAvansi(_ context.Context, _ db.Transaction, nalogID, rdokID int64) (int64, error) {
	f.log("avansi", nalogID, rdokID)
	return 0, nil
}
func (f *fakeOtpRepo) KomOtpPotvrdi(_ context.Context, _ db.Transaction, otpID int64, dokum, brotp int, opis string) (bool, error) {
	f.log("potvrdi", otpID, dokum, brotp, opis)
	return true, nil
}
func (f *fakeOtpRepo) KomOtpNalogIsporuceno(_ context.Context, _ db.Transaction, otpID int64) error {
	f.log("isporuceno", otpID)
	return nil
}
func (f *fakeOtpRepo) DrstaLotStanje(_ context.Context, konto string, sifra, mag int, otk, serija string, roktr int64) (repository.DrstaLot, bool, error) {
	z, ok := f.zalihe[fmt.Sprintf("%s|%d|%d|%s", konto, sifra, mag, otk)]
	return z, ok, nil
}
func (f *fakeOtpRepo) PopdvPoljeUlaz(_ context.Context, polje string) (repository.PopdvUlaz, bool, error) {
	p, ok := f.popdv[polje]
	return p, ok, nil
}
func (f *fakeOtpRepo) KomProKupac(_ context.Context, konto, sifra string) (domain.KomProKupac, bool, error) {
	k, found, err := f.fakeProRepo.KomProKupac(context.Background(), konto, sifra)
	if sifra == "INO" {
		k.Ter, found = 3, true
	}
	return k, found, err
}

// fakeOtpRobno records the posting of the robno service.
type fakeOtpRobno struct {
	repo *fakeOtpRepo
}

func (r fakeOtpRobno) AzurirajRSTA(_ context.Context, _ db.Transaction, p domain.RobPromenaStanja) error {
	r.repo.log("rsta", p.Konto, p.Sifra, p.Kolicina, p.Iznos, p.DokvrstaID, p.Tipdok)
	return nil
}
func (r fakeOtpRobno) AzurirajDRSTA(_ context.Context, _ db.Transaction, p domain.RobPromenaStanja) error {
	r.repo.log("drsta", p.Konto, p.Sifra, p.Kolicina, p.Otk, p.Danal.Format("2006-01-02"))
	return nil
}
func (r fakeOtpRobno) AzurirajRNAL(_ context.Context, _ db.Transaction, rnalID int64, dug, pot bool, iznos float64) error {
	r.repo.log("rnal iznos", rnalID, dug, pot, iznos)
	return nil
}
func (r fakeOtpRobno) ObracunFakture(_ context.Context, rdokID int64) error {
	r.repo.log("obracun", rdokID)
	return nil
}
func (r fakeOtpRobno) IzracunajNCENA(context.Context, time.Time, time.Time, int, int) (float64, error) {
	return 44.444, nil
}

var otpDatum = sql.NullTime{Time: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Valid: true}

func newFakeOtpRepo() *fakeOtpRepo {
	pro := newFakeProRepo()
	return &fakeOtpRepo{fakeProRepo: pro,
		vrste: []domain.KomOtpVrsta{{DokvrstaID: 5, Vrd: 188, Opis: "FAKTURA(GOT.PROIZ.)", Dokozn: "FPR", Predznak: "+", Kodknj: "P", FlgPopdv: true,
			Polje: "3.2"}, {DokvrstaID: 6, Vrd: 189, Opis: "STORNO", Dokozn: "FPR", Predznak: "-", Kodknj: "P"},
			{DokvrstaID: 7, Vrd: 130, Opis: "FAKTURA(TRG.ROBA)", Dokozn: "FAK", Predznak: "+", Kodknj: "P"}},
		otpremnice: map[int64]domain.KomOtpremnicaRow{
			60: {RdokID: 60, Vrd: 188, Dokum: 990001, Dadok: otpDatum, Brotp: 990001, Datotp: otpDatum, MagaciniID: 3, Mag: 30, Tipzal: 2, Nacvodzal: 3,
				Tipdok: "43", Fkto: "2040", Fana: "00031", Doksts: "10", SifVal: 941, Kurs: 1, VaziZa: 1, Category: "S", VrdokID: 50, Ugrabat: 10,
				NalogDokum: 3, NalogDadok: otpDatum, Posuslnab: "-"},
			62: {RdokID: 62, Vrd: 188, MagaciniID: 3, Mag: 30, Doksts: "40", VrdokID: 50},
		},
		nalozi: map[int64]domain.KomOtpNalog{
			50: {RdokID: 50, Dokum: 3, Dadok: otpDatum, MagaciniID: 3, Doksts: "20", Fkto: "2040", Fana: "00031", SifVal: 941, Kurs: 1, VaziZa: 1},
			51: {RdokID: 51, Dokum: 4, Dadok: otpDatum, MagaciniID: 4, Doksts: "20"},
		},
		otpStavke: map[int64][]domain.KomOtpStavkaRow{
			50: {{RproID: 71, Sifra: 1001, Cena: 80, Fcena: 100}},
			60: {
				{RproID: 81, RproID1: 71, Rbr: 1, Konto: "1320", Sifra: 1001, RsifID: 44, Naziv: "Brašno", JM: "kg", Otk: "001", Serija: "S1",
					Roktr: 82000, Trazkolic: 10, Kolic: 6, Cena: 80, Fcena: 100, CenaVal: 100, Rab: 5, Po: 1, Tezjm: 1, JmPostoji: true, Decimale: 2},
				{RproID: 82, RproID1: 72, Rbr: 2, Konto: "1320", Sifra: 1002, RsifID: 45, Naziv: "Šećer", JM: "kg", Trazkolic: 2, Kolic: 0, Cena: 50,
					Fcena: 60, CenaVal: 60, Po: 2, JmPostoji: true},
			},
		},
		zalihe: map[string]repository.DrstaLot{"1320|1001|30|001": {Ulaz: 20, Izlaz: 5}},
		popdv:  map[string]repository.PopdvUlaz{"3.2": {Tip: 3, Vktip: "I"}, "8.1": {Tip: 3, Vktip: "U"}},
		prosnc: 55.555,
	}
}

func newOtpService(repo *fakeOtpRepo) *OtpremnicaResource {
	return NewOtpremnicaService(repo, fakeMagacini{}, fakeOtpRobno{repo: repo})
}

// TestKomOtpProdvred checks the legacy PreracunFAKTURAPROIZVGLOBEDOK.
func TestKomOtpProdvred(t *testing.T) {
	stavke := []domain.KomOtpStavkaRow{{Trazkolic: 10, Fcena: 100, Rab: 5, Po: 1}, {Trazkolic: 2, Fcena: 60, Po: 2}}
	// iznos 1120, rabat 50, ugovoreni rabat 10% of 1070 = 107, kasa 0; PDV: (950 - 95) 20% = 171 + (120 - 12) 10% = 10.8.
	if got := komOtpProdvred(stavke, 10, 0, "", map[int]float64{1: 20, 2: 10}); got != 1144.8 {
		t.Errorf("prodvred = %v", got)
	}
	if got := komOtpProdvred(stavke, 10, 0, "član 24", map[int]float64{1: 20, 2: 10}); got != 963 {
		t.Errorf("prodvred bez PDV = %v", got)
	}
}

// TestKomOtpRpro checks the stavka of the faktura by the way the magacin keeps its zalihe, the storno and the poreska
// kategorija.
func TestKomOtpRpro(t *testing.T) {
	st := domain.KomOtpStavkaRow{Sifra: 1, Kolic: 3, Kolic1: 1, Cena: 80, Fcena: 100, Po: 1}
	o := domain.KomOtpremnicaRow{Nacvodzal: 3, Category: "S"}
	if r := komOtpRpro(o, st, false, 55.5, false); r.Cena != 55.5 || r.Ncena != 55.5 || r.Iznos != 166.5 || r.Kolic != 3 || r.Taxcat != "S20" ||
		r.Taxexreco != "-" {
		t.Errorf("nabavna %+v", r)
	}
	o.Nacvodzal, o.Category = 1, "AE"
	if r := komOtpRpro(o, st, true, 0, false); r.Cena != 100 || r.Iznos != -300 || r.Kolic != -3 || r.Kolic1 != -1 || r.Taxcat != "AE" {
		t.Errorf("veleprodajna storno %+v", r)
	}
	o.Nacvodzal = 2
	if r := komOtpRpro(o, st, false, 0, false); r.Cena != 80 || r.Iznos != 0 {
		t.Errorf("planska %+v", r)
	}
}

// TestGetOtpremnica checks a new otpremnica (the temporary brojevi, the defaults of a PDV obveznik) and a saved one of a
// domestic kupac (the valuta of the firm).
func TestGetOtpremnica(t *testing.T) {
	repo := newFakeOtpRepo()
	s := newOtpService(repo)
	dok, err := s.GetOtpremnica(komCtx(), 0, 188, 3)
	if err != nil {
		t.Fatal(err)
	}
	h := dok.Header
	if h.Brotp != 990002 || h.Dokum != 990003 || h.Mag != 30 || h.Category != "S" || h.Polje != "3.2" || dok.Nastanak != "35" || h.Posuslnab != "-" {
		t.Errorf("new %+v nastanak %q", h, dok.Nastanak)
	}
	o := repo.otpremnice[60]
	o.SifVal, o.Kurs = 978, 117
	repo.otpremnice[60] = o
	if dok, err = s.GetOtpremnica(komCtx(), 60, 0, 0); err != nil || !dok.Header.Snimljen || dok.Header.SifVal != 941 || dok.Header.Kurs != 1 {
		t.Errorf("saved %+v %v", dok.Header, err)
	}
	if dok, _ = s.GetOtpremnica(komCtx(), 62, 0, 0); dok.Poruka != komMsgOtpPotvrdjena || dok.Header.RdokID != 0 {
		t.Errorf("potvrđena %+v", dok)
	}
}

// TestValidateOtpremnica checks the legacy ValidacijaEDOK and ValidacijaEPRO.
func TestValidateOtpremnica(t *testing.T) {
	valid := domain.KomOtpParams{MagaciniID: "3", Vrd: "188", NalogID: "50", Datotp: "2026-03-10", Datum: "2026-03-10", Kurs: "1", VaziZa: "1",
		Category: "S", Nastanak: "35", Polje: "3.2"}
	for _, tt := range []struct {
		name   string
		repo   func(r *fakeOtpRepo)
		change func(p *domain.KomOtpParams)
		want   []string
	}{
		{name: "valid", change: func(*domain.KomOtpParams) {}},
		{name: "without the nalog", change: func(p *domain.KomOtpParams) { p.NalogID = "" }, want: []string{"nalogid"}},
		{name: "a nalog of another magacin", change: func(p *domain.KomOtpParams) { p.NalogID = "51" }, want: []string{"nalogid"}},
		{name: "a nalog with a cena 0", repo: func(r *fakeOtpRepo) { r.otpStavke[50] = []domain.KomOtpStavkaRow{{Fcena: 0, Cena: 5}} },
			change: func(*domain.KomOtpParams) {}, want: []string{"nalogid"}},
		{name: "a faktura of the roba from a magacin of the proizvodi", repo: func(r *fakeOtpRepo) { r.tipmag = "G" },
			change: func(p *domain.KomOtpParams) { p.Vrd = "130" }, want: []string{"vrd"}},
		{name: "a faktura of the proizvodi from a magacin of the roba", change: func(*domain.KomOtpParams) {}, repo: func(r *fakeOtpRepo) { r.tipmag = "V" },
			want: []string{"vrd"}},
		{name: "typed brojevi that exist", repo: func(r *fakeOtpRepo) { r.brotpIma, r.rdokIma = true, true },
			change: func(p *domain.KomOtpParams) { p.DozvoliBroj, p.Brotp, p.Dokum = "true", "15", "101" }, want: []string{"brotp", "brfakture"}},
		{name: "typed brojevi missing", change: func(p *domain.KomOtpParams) { p.DozvoliBroj = "true" }, want: []string{"brotp", "brfakture"}},
		{name: "datum otpremnice before the nalog", change: func(p *domain.KomOtpParams) { p.Datotp = "2026-03-09" }, want: []string{"datotp"}},
		{name: "datum of another year", change: func(p *domain.KomOtpParams) { p.Datum = "2025-12-31" }, want: []string{"datum"}},
		{name: "datum before the otpremnica", change: func(p *domain.KomOtpParams) { p.Datotp, p.Datum = "2026-03-12", "2026-03-11" },
			want: []string{"datum"}},
		{name: "POPDV polje of the ulazni računi", change: func(p *domain.KomOtpParams) { p.Polje = "8.1" }, want: []string{"popdv"}},
		{name: "POPDV polje missing", change: func(p *domain.KomOtpParams) { p.Polje = "" }, want: []string{"popdv"}},
		{name: "kurs, važi za and oznaka", change: func(p *domain.KomOtpParams) { p.Kurs, p.VaziZa, p.Category = "0", "0", "" },
			want: []string{"kurs", "vaziza", "oznakakat"}},
		{name: "a potvrđena one", change: func(p *domain.KomOtpParams) { p.RdokID = 62 }, want: []string{"datum"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeOtpRepo()
			repo.tipmag = "G"
			if tt.repo != nil {
				tt.repo(repo)
			}
			p := valid
			tt.change(&p)
			var got []string
			for _, e := range newOtpService(repo).ValidateOtpremnica(komCtx(), p) {
				got = append(got, e.Field)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSaveOtpremnica checks a new otpremnica (the nalog marked, the temporary brojevi, the stavke of the nalog, the
// prodajna vrednost) and a saved one with a new kurs and typed brojevi.
func TestSaveOtpremnica(t *testing.T) {
	repo := newFakeOtpRepo()
	repo.otpremnice[61] = domain.KomOtpremnicaRow{RdokID: 61, Dadok: otpDatum}
	s := newOtpService(repo)
	id, err := s.SaveOtpremnica(komCtx(), domain.KomOtpParams{MagaciniID: "3", Vrd: "189", NalogID: "50", Datotp: "2026-03-10",
		Datum: "2026-03-10", Kurs: "1", VaziZa: "1", Napomena: "Zvati"})
	if err != nil || id != 61 {
		t.Fatalf("id %d err %v", id, err)
	}
	want := []string{"nalog 40 50", "dodaj 50 189 990002 990003 Zvati", "kopiraj 61 50 true false", "prodvred 61 0"}
	if !reflect.DeepEqual(repo.zapis, want) || repo.commits != 1 {
		t.Errorf("new\n got %q\nwant %q", repo.zapis, want)
	}
	repo.zapis = nil
	if _, err := s.SaveOtpremnica(komCtx(), domain.KomOtpParams{RdokID: 60, Datotp: "2026-03-10", Datum: "2026-03-10", Kurs: "117.5",
		VaziZa: "1", DozvoliBroj: "true", Brotp: "15", Dokum: "101"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(repo.zapis[0], "izmeni 60 15 101 117.5") || repo.zapis[1] != "preracunaj 60" || repo.zapis[2] != "kopiraj 60 50 false true" {
		t.Errorf("saved %q", repo.zapis)
	}
}

// TestDeleteOtpremnica checks the legacy BTN_Brisanje: only a new otpremnica, its nalog potvrđen again.
func TestDeleteOtpremnica(t *testing.T) {
	repo := newFakeOtpRepo()
	s := newOtpService(repo)
	if poruka, _ := s.DeleteOtpremnica(komCtx(), 62); poruka != komMsgOtpPotvrdjenaBrisi {
		t.Errorf("potvrđena %q", poruka)
	}
	if poruka, err := s.DeleteOtpremnica(komCtx(), 60); err != nil || poruka != "" || !reflect.DeepEqual(repo.zapis, []string{"nalog 20 50"}) {
		t.Errorf("%q %v %q", poruka, err, repo.zapis)
	}
}

// TestOtpStavka checks the isporučena količina of a stavka (above the tražena, negative, the storno) and the move of the
// reservation to a new lot.
func TestOtpStavka(t *testing.T) {
	repo := newFakeOtpRepo()
	s := newOtpService(repo)
	for _, tt := range []struct {
		vrd      int
		kolicina string
		want     string
	}{
		{188, "6", ""}, {188, "11", komMsgOtpIsporucenaVeca}, {188, "-1", komMsgOtpIsporucenaNegativ},
		{189, "3", komMsgOtpStornoNegativna}, {189, "-1", ""},
	} {
		o := repo.otpremnice[60]
		o.Vrd = tt.vrd
		repo.otpremnice[60] = o
		if tt.vrd == 189 {
			repo.otpStavke[60][0].Trazkolic = -10
		}
		var got string
		for _, e := range s.ValidateStavka(komCtx(), domain.KomOtpStavkaParams{RdokID: 60, RproID: 81, Kolicina: tt.kolicina}) {
			got = e.ErrorMessage
		}
		if got != tt.want {
			t.Errorf("vrd %d kolicina %s: %q", tt.vrd, tt.kolicina, got)
		}
	}
	repo = newFakeOtpRepo()
	s = newOtpService(repo)
	if err := s.SaveStavka(komCtx(), domain.KomOtpStavkaParams{RdokID: 60, RproID: 81, Kolicina: "5", DozvoliLot: "true", Otk: "002",
		Serija: "S2", Rok: "2026-12-31"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(repo.zapis[0], "stavka 60 81 5 002 S2") {
		t.Errorf("stavka %q", repo.zapis)
	}
	want := []string{"drsta 7 rez -10 nn=false", "novi drsta 1320/1001 002/S2/" + fmt.Sprint(komWinDevDatum("2026-12-31")) + " rsif 44 rez 10"}
	if !reflect.DeepEqual(repo.rez, want) {
		t.Errorf("rezervacija\n got %q\nwant %q", repo.rez, want)
	}
}

// TestProveriPotvrduOtpremnice checks the legacy BTN_POTVRDADOK: the zatvorena godina, the nastanak, the OTK question,
// the zalihe of the lots, a storno without ulaz, the brojevi.
func TestProveriPotvrduOtpremnice(t *testing.T) {
	params := domain.KomOtpParams{RdokID: 60, Nastanak: "35"}
	for _, tt := range []struct {
		name    string
		repo    func(r *fakeOtpRepo)
		params  func(p *domain.KomOtpParams)
		poruka  string
		pitanja int
	}{
		{name: "ok"},
		{name: "zatvorena godina", repo: func(r *fakeOtpRepo) { r.zatvorena = true }, poruka: komMsgOtpGodinaZatvorena},
		{name: "nastanak not chosen", params: func(p *domain.KomOtpParams) { p.Nastanak = "999" }, poruka: komMsgOtpNastanak},
		{name: "without OTK of a JM with a length", repo: func(r *fakeOtpRepo) {
			r.otpStavke[60][0].Otk, r.otpStavke[60][0].ImaDuzinu = "", true
			r.zalihe["1320|1001|30|"] = repository.DrstaLot{Ulaz: 10}
		}, pitanja: 1},
		{name: "not enough zaliha", repo: func(r *fakeOtpRepo) { r.zalihe["1320|1001|30|001"] = repository.DrstaLot{Ulaz: 5, Izlaz: 1} },
			poruka: komMsgOtpZalihe},
		{name: "cena 0", repo: func(r *fakeOtpRepo) { r.otpStavke[60][1].Fcena = 0 }, poruka: komMsgOtpCenaNula},
		{name: "nothing delivered", repo: func(r *fakeOtpRepo) { r.otpStavke[60][0].Kolic = 0 }, poruka: komMsgOtpNulaKolicina},
		{name: "a storno without ulaz", repo: func(r *fakeOtpRepo) {
			o := r.otpremnice[60]
			o.Vrd = 189
			r.otpremnice[60] = o
			r.zalihe["1320|1001|30|001"] = repository.DrstaLot{}
		}, poruka: komMsgOtpStorno},
		{name: "typed broj fakture exists", repo: func(r *fakeOtpRepo) {
			o := r.otpremnice[60]
			o.Dokum = 101
			r.otpremnice[60] = o
			r.rdokIma = true
		}, poruka: komMsgOtpFakturaPostoji},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeOtpRepo()
			if tt.repo != nil {
				tt.repo(repo)
			}
			p := params
			if tt.params != nil {
				tt.params(&p)
			}
			r, err := newOtpService(repo).ProveriPotvrdu(komCtx(), p)
			if err != nil || r.Poruka != tt.poruka || len(r.BezOtk) != tt.pitanja {
				t.Errorf("provera %+v err %v", r, err)
			}
		})
	}
}

// TestPotvrdiOtpremnicu checks the legacy SnimiRDOK: the brojevi, the robni nalog, the faktura with its stavke (only the
// delivered ones) and cene, the posting, the reservations of all the stavke released, the avansi, the otpremnica and the
// nalog, then the obračun.
func TestPotvrdiOtpremnicu(t *testing.T) {
	repo := newFakeOtpRepo()
	r, err := newOtpService(repo).PotvrdiOtpremnicu(komCtx(), domain.KomOtpParams{RdokID: 60, Nastanak: "35", OpisDok: "Za Marka"})
	if err != nil || r.Poruka != komMsgPotvrdjen || !r.Potvrdjena {
		t.Fatalf("%+v %v", r, err)
	}
	want := []string{
		"tezina 60 6",
		"rnal 43 3 30",
		"brdo 8",
		"rdok 60 8 101 15 Za Marka Dok.br:101 35",
		"rpro 8 90 1001 6 55.56 55.56 333.36 S20",
		"rcene 1001 55.56 80 100",
		"rsta 1320 1001 6 333.36 5 43",
		"drsta 1320 1001 6 001 2026-03-10",
		"rnal iznos 8 false true 333.36",
		"brst 90 1",
		"avansi 50 90",
		"potvrdi 60 101 15 Otpremnica br.:15",
		"isporuceno 60",
		"obracun 90",
	}
	if !reflect.DeepEqual(repo.zapis, want) || repo.commits != 1 {
		t.Errorf("potvrda\n got %q\nwant %q", repo.zapis, want)
	}
	wantRez := []string{"drsta 7 rez -10 nn=false", "rsta 9 rez -10 nn=true"}
	if !reflect.DeepEqual(repo.rez, wantRez) {
		t.Errorf("rezervacije %q", repo.rez)
	}
}

// TestOtpStampa checks the print: only the delivered stavke, the naziv with the lot and the komercijalni opis, the cene in
// the valuta of a foreign one, the tekući računi of the firm without the žiro račun.
func TestOtpStampa(t *testing.T) {
	repo := newFakeOtpRepo()
	repo.otpStavke[60][0].Komercopis = "tip 400"
	s := newOtpService(repo)
	st, poruka, err := s.GetStampa(komCtx(), 60, true)
	if err != nil || poruka != "" {
		t.Fatalf("%q %v", poruka, err)
	}
	if len(st.Stavke) != 1 || st.Stavke[0].Naziv != "Brašno\nSerija: S1 Rok trajanja: "+komWinDevDatumText(82000)+" tip 400" || st.Svega != 600 ||
		st.TekuciRacun != "160-1  Banka" || !st.Nepotvrdjen || st.Ino {
		t.Errorf("domaća %+v", st)
	}
	o := repo.otpremnice[60]
	o.SifVal, o.Kurs, o.Posuslnab = 978, 117, "160-2"
	repo.otpremnice[60] = o
	repo.otpStavke[60][0].CenaVal = 0.85
	if st, _, _ = s.GetStampa(komCtx(), 60, false); !st.Ino || st.Stavke[0].Cena != 0.85 || st.Svega != 5.1 || st.UkupnoRSD != 600 || st.TekuciRacun != "160-2" {
		t.Errorf("izvozna %+v", st)
	}
}

package komercijalno

import (
	"context"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// fakeEotpRepo is the repository of the e-otpremnica of the tests: the partneri by "konto-šifra", the mesta isporuke by
// MI and the kupac and the mesto isporuke of the nalog 50.
type fakeEotpRepo struct {
	partneri map[string]domain.KomEotpPartner
	mesta    map[int]domain.KomEotpAdresa
}

func (f *fakeEotpRepo) KomEotp(context.Context, int64) (domain.KomEotpPodaci, bool, error) {
	return domain.KomEotpPodaci{}, false, nil
}
func (f *fakeEotpRepo) KomEotpIzmeni(context.Context, db.Transaction, int64, domain.KomEotpUpis) error {
	return nil
}
func (f *fakeEotpRepo) KomEotpFaze(context.Context, int64) ([]domain.KomEotpFaza, error) {
	return nil, nil
}
func (f *fakeEotpRepo) KomEotpOpasne(context.Context, int64) ([]domain.KomEotpOpasna, error) {
	return nil, nil
}
func (f *fakeEotpRepo) KomEotpBrisi(context.Context, db.Transaction, int64) error { return nil }
func (f *fakeEotpRepo) KomEotpFazaDodaj(context.Context, db.Transaction, int64, domain.KomEotpFaza) error {
	return nil
}
func (f *fakeEotpRepo) KomEotpOpasnaDodaj(context.Context, db.Transaction, int64, domain.KomEotpOpasna) error {
	return nil
}
func (f *fakeEotpRepo) KomEotpPartner(_ context.Context, konto, sifra string) (domain.KomEotpPartner, bool, error) {
	p, ok := f.partneri[konto+"-"+sifra]
	return p, ok, nil
}
func (f *fakeEotpRepo) KomEotpPosiljalac(context.Context, string) (domain.KomEotpPartner, error) {
	return domain.KomEotpPartner{Naziv: "FIRMA"}, nil
}
func (f *fakeEotpRepo) KomEotpMagacinAdresa(context.Context, int) (domain.KomEotpAdresa, bool, error) {
	return domain.KomEotpAdresa{}, false, nil
}
func (f *fakeEotpRepo) KomEotpMestoIsporuke(_ context.Context, _, _ string, mi int) (domain.KomEotpAdresa, bool, error) {
	a, ok := f.mesta[mi]
	return a, ok, nil
}
func (f *fakeEotpRepo) KomEotpNalog(_ context.Context, nalogID int64) (string, string, int, bool, error) {
	if nalogID != 50 {
		return "", "", 0, false, nil
	}
	return "2040", "15", 3, true, nil
}

func newFakeEotpRepo() *fakeEotpRepo {
	return &fakeEotpRepo{
		partneri: map[string]domain.KomEotpPartner{"2040-15": {Naziv: "KUPAC DOO"}, "4350-12": {Naziv: "PREVOZ DOO"}},
		mesta:    map[int]domain.KomEotpAdresa{3: {Ulica: "Bulevar 5", Mesto: "Beograd", Pobro: "11000"}},
	}
}

// TestKomEotpDatumVreme checks the datumi i vremena of the form: a datetime-local field, the Serbian form, none, a wrong
// one.
func TestKomEotpDatumVreme(t *testing.T) {
	for _, tt := range []struct {
		value     string
		want      time.Time
		valid, ok bool
	}{
		{value: "2026-03-10T09:30", want: time.Date(2026, 3, 10, 9, 30, 0, 0, time.UTC), valid: true, ok: true},
		{value: "10.03.2026 09:30", want: time.Date(2026, 3, 10, 9, 30, 0, 0, time.UTC), valid: true, ok: true},
		{value: " ", ok: true},
		{value: "sutra"},
	} {
		t.Run(tt.value, func(t *testing.T) {
			got, ok := komEotpDatumVreme(tt.value)
			if ok != tt.ok || got.Valid != tt.valid || (tt.valid && !got.Time.Equal(tt.want)) {
				t.Errorf("komEotpDatumVreme(%q) = %v, %v", tt.value, got, ok)
			}
		})
	}
}

// TestKomEotpRedovi checks the rows of the grids: the empty ones left out, the redni brojevi again, the first row that is
// not complete.
func TestKomEotpRedovi(t *testing.T) {
	faze, i := komEotpFaze(domain.KomEotpParams{FazaNacOtp: []string{"2", "", "4"}, FazaNaziv: []string{"Prevoz DOO", "", " Pera Perić "}})
	if i != -1 || len(faze) != 2 || faze[1] != (domain.KomEotpFaza{Rbr: 2, NacOtp: 4, Naziv: "Pera Perić"}) {
		t.Errorf("faze = %+v, %d", faze, i)
	}
	if _, i := komEotpFaze(domain.KomEotpParams{FazaNacOtp: []string{"9"}, FazaNaziv: []string{"X"}}); i != 0 {
		t.Errorf("a wrong način otpreme gave %d", i)
	}
	opasne, i := komEotpOpasne(domain.KomEotpParams{OpmaNaziv: []string{"", "UN 1203"}, OpmaOpis: []string{"", "Benzin"}, OpmaKoment: []string{"", ""}})
	if i != -1 || len(opasne) != 1 || opasne[0] != (domain.KomEotpOpasna{Rbr: 1, Naziv: "UN 1203", Opis: "Benzin"}) {
		t.Errorf("opasne = %+v, %d", opasne, i)
	}
	if _, i := komEotpOpasne(domain.KomEotpParams{OpmaNaziv: []string{""}, OpmaOpis: []string{"Benzin"}}); i != 0 {
		t.Errorf("an opasna materija without a naziv gave %d", i)
	}
}

// TestKomEotpUpis checks what the save writes: the jedinica mere "-" is none, the prevoznik only for its način otpreme,
// the lično podaci only for theirs, the roba treće strane and the opasne materije only with their checkbox.
func TestKomEotpUpis(t *testing.T) {
	e := domain.KomEotpParams{TipEotp: "Ext", Jm1: "-", Jm2: "m3", NacOtp: "1", PrevKto: "4350", PrevAna: "12", LicIme: "Pera",
		Roba3StrTxt: "Roba firme X", OpmaNaziv: []string{"UN 1203"}, Datpoc: "2026-03-10T08:00", Mi: "3", FazaNacOtp: []string{"1"},
		FazaNaziv: []string{"Sopstveni"}}
	u := komEotpUpis(e)
	if u.Jm1 != "" || u.Jm2 != "m3" || u.PrevKto != "" || u.LicIme != "" || u.Roba3StrTxt != "" || len(u.Opasne) != 0 || !u.Datpoc.Valid ||
		u.Mi != 3 || len(u.Faze) != 1 {
		t.Errorf("upis = %+v", u)
	}
	e.NacOtp, e.Roba3Str, e.OpasneMat = "2", "true", "true"
	u = komEotpUpis(e)
	if u.PrevKto != "4350" || u.PrevAna != "12" || u.Roba3StrTxt != "Roba firme X" || len(u.Opasne) != 1 {
		t.Errorf("upis of the prevoznik = %+v", u)
	}
	e.NacOtp = "4"
	if u = komEotpUpis(e); u.LicIme != "Pera" || u.PrevKto != "" {
		t.Errorf("upis of the lično preuzimanje = %+v", u)
	}
}

// TestEotpPrevoznik checks the prevoznik of the način otpreme: the firm, the one of the kontni plan, the primalac, none.
func TestEotpPrevoznik(t *testing.T) {
	s := NewEotpremnicaService(nil, newFakeEotpRepo())
	e := &domain.KomEotpView{Posiljalac: domain.KomEotpPartner{Naziv: "FIRMA"}, Primalac: domain.KomEotpPartner{Naziv: "KUPAC DOO"}}
	e.PrevKto, e.PrevAna = "4350", "12"
	for _, tt := range []struct {
		nacin int
		want  string
	}{{1, "FIRMA"}, {2, "PREVOZ DOO"}, {3, "KUPAC DOO"}, {4, ""}, {0, ""}} {
		e.NacOtp = tt.nacin
		p, err := s.prevoznik(context.Background(), e)
		if err != nil || p.Naziv != tt.want {
			t.Errorf("način %d: prevoznik %q, %v", tt.nacin, p.Naziv, err)
		}
	}
}

// TestGetEotpNalog checks what a new e-otpremnica takes of its nalog: the primalac and the adresa odredišta of its mesto
// isporuke; nothing for another nalog.
func TestGetEotpNalog(t *testing.T) {
	s := NewEotpremnicaService(nil, newFakeEotpRepo())
	n, found, err := s.GetEotpNalog(context.Background(), 50)
	if err != nil || !found || n.Primalac.Naziv != "KUPAC DOO" || n.Mi != 3 || n.Odrediste.Ulica != "Bulevar 5" || n.Fkto != "2040" {
		t.Errorf("nalog = %+v, %v, %v", n, found, err)
	}
	if _, found, err := s.GetEotpNalog(context.Background(), 51); found || err != nil {
		t.Errorf("another nalog: %v, %v", found, err)
	}
}

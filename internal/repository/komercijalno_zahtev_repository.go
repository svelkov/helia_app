package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// KomercijalnoRepository holds the statements of the komercijalno module: the zahtevi za nabavku (edok of the vrste
// of the modul PRO and the group INR) with their stavke (epro). It uses the statement helpers of RobnoRepository on
// the same database, and its lookups of the šifarnici (e.g. OrgjedPostoji). Every statement on a table with god/kar is
// restricted to the period of the session in ctx.
type KomercijalnoRepository struct {
	RobnoRepository
}

// NewKomercijalnoRepository returns the repository of the komercijalno module.
func NewKomercijalnoRepository(database db.Database) *KomercijalnoRepository {
	return &KomercijalnoRepository{RobnoRepository: RobnoRepository{db: database}}
}

// The modul and the group of the vrste dokumenata of the zahtevi za nabavku (the legacy PR_QRY_DOKVRSTA with gsDOKPRO
// and "INR").
const (
	komZahtevModul = "PRO"
	komZahtevGrupa = "INR"
)

// komZahtevTipoviArtikla are the tipovi (rsif.tip) of the artikli of a stavka of a zahtev za nabavku, the same as the
// search "artikli-proizvodnja" (the legacy pr_Trazi_GPROIZVOD1).
var komZahtevTipoviArtikla = []string{"S", "A", "R", "D", "K"}

// KomZahtevVrste returns the vrste dokumenata of the zahtevi za nabavku of the period (dokvrsta of the modul PRO and
// the group INR) as "vrd - opis" keyed by the vrd, in the order of the vrd.
func (r *KomercijalnoRepository) KomZahtevVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return r.komVrste(ctx, komZahtevGrupa)
}

// komVrste returns the vrste dokumenata of the modul PRO of a group of the period as "vrd - opis" keyed by the vrd, in
// the order of the vrd (the legacy PR_QRY_DOKVRSTA).
func (r *KomercijalnoRepository) komVrste(ctx context.Context, grupa string) ([]domain.ComboItem, error) {
	b := querybuilder.New(`select coalesce(v.vrd, 0)::text as key, coalesce(v.vrd, 0)::text || ' - ' || coalesce(v.opis, '') as value
		from dokvrsta v`).
		Where("upper(trim(coalesce(v.modul, ''))) = ?", komZahtevModul).
		Where("upper(trim(coalesce(v.grpdok, ''))) = ?", grupa)
	if err := inPeriod(ctx, b, "v."); err != nil {
		return nil, err
	}
	b.OrderBy("v.vrd")
	var rows []domain.ComboItem
	if err := r.selectRows(ctx, "vrste zahteva za nabavku", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomZahtevNoviBroj returns the next broj of a zahtev za nabavku of the vrsta in the period (the legacy NoviBrojEDOK:
// the largest dokum of the vrsta + 1).
func (r *KomercijalnoRepository) KomZahtevNoviBroj(ctx context.Context, tx db.Transaction, vrd int) (int, error) {
	b := querybuilder.New(`select coalesce(max(dokum), 0)::int + 1 from edok`).Where("vrd = ?", vrd)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var broj int
	if err := r.row(ctx, tx, "novi broj zahteva za nabavku", b, &broj); err != nil {
		return 0, err
	}
	return broj, nil
}

// KomZahtev returns a zahtev za nabavku of the period with the texts of its vrsta, magacin, komercijalista, org.
// jedinica and mesto troška; found is false without it.
func (r *KomercijalnoRepository) KomZahtev(ctx context.Context, rdokID int64) (domain.KomZahtevRow, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.vrd, 0)::int, coalesce(v.opis, ''), coalesce(e.dokum, 0)::int, e.dadok,
			coalesce(e.magaciniid, 0)::int, coalesce(e.mag, 0)::int, coalesce(m.opis, ''),
			e.komid::bigint, coalesce(k.sifkom, 0)::int, coalesce(k.imeprezime, ''),
			e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''),
			coalesce(e.foot, ''), coalesce(e.doksts, ''), coalesce(e.xopunos, ''), coalesce(e.knjige_1, ''), coalesce(e.knjige_4, '')
		from edok e`).
		AddJoin("left join dokvrsta v on v.vrd = e.vrd and v.god = e.god and v.kar = e.kar").
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		Where("e.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomZahtevRow{}, false, err
	}
	var z domain.KomZahtevRow
	err := r.row(ctx, nil, "zahtev za nabavku", b, &z.RdokID, &z.Vrd, &z.VrstaOpis, &z.Dokum, &z.Dadok, &z.MagaciniID, &z.Mag,
		&z.MagacinOpis, &z.KomID, &z.KomSifra, &z.KomIme, &z.Orgjed, &z.OrgjedText, &z.MestoTroska, &z.MestoTroskaText, &z.Foot,
		&z.Doksts, &z.Opunos, &z.Knjige1, &z.Knjige4)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomZahtevRow{}, false, nil
	}
	if err != nil {
		return domain.KomZahtevRow{}, false, err
	}
	return z, true, nil
}

// KomZahtevDodaj adds a zahtev za nabavku (the legacy SnimiEDOK of a new document): not of a nalog, nepotvrđen
// (doksts 10), not posted (knjige N), its opis "Zahtev za nabavku br: <broj>", the JM of the ambalaža empty (jm1
// and jm2 are not null without a default), with the audit columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomZahtevDodaj(ctx context.Context, tx db.Transaction, z domain.KomZahtevUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana,
			pla, otp, foot, brst, idorgjed, mestotrid, komid, doksts, knjige_1, knjige_2, jm1, jm2, xdatunosa, xopunos)
		values (?, ?, 0, ?, ?, ?, ?, ?, 0, 0, ?, 0, '', '', '', '', ?, 0, ?, ?, ?, ?, 'N', 'N', '', '', now(), ?)`,
		s.God, s.Kar, z.Vrd, z.Dokum, z.Dadok, z.Mag, z.MagaciniID, fmt.Sprintf("Zahtev za nabavku br: %d", z.Dokum), z.Foot,
		z.Orgjed, z.MestoTroska, z.KomID, domain.KomZahtevNepotvrdjen, s.UserName).
		Returning("rdokid")
	var rdokID int64
	if err := r.row(ctx, tx, "novi zahtev za nabavku", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomZahtevIzmeni saves the header of a zahtev za nabavku of the period (the legacy SnimiEDOK of a saved document):
// its datum, komercijalista, org. jedinica, mesto troška and tekst, with the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomZahtevIzmeni(ctx context.Context, tx db.Transaction, z domain.KomZahtevUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("dadok = ?", z.Dadok).
		Set("komid = ?", z.KomID).
		Set("idorgjed = ?", z.Orgjed).
		Set("mestotrid = ?", z.MestoTroska).
		Set("foot = ?", z.Foot).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", z.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "zahtev za nabavku", b)
}

// KomZahtevPotvrdi confirms a nepotvrđen zahtev za nabavku of the period (the legacy BTN_Potvrda: doksts 40). It
// returns false when the zahtev is not nepotvrđen any more (e.g. confirmed meanwhile).
func (r *KomercijalnoRepository) KomZahtevPotvrdi(ctx context.Context, rdokID int64) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	b := querybuilder.NewUpdate("edok").
		Set("doksts = ?", domain.KomZahtevPotvrdjen).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", rdokID).
		Where("coalesce(doksts, '') = ?", domain.KomZahtevNepotvrdjen)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	n, err := r.exec(ctx, nil, "potvrda zahteva za nabavku", b)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomZahtevBrisi deletes a zahtev za nabavku of the period with its stavke (the legacy BTN_Brisanje); ErrNotFound
// without it.
func (r *KomercijalnoRepository) KomZahtevBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	stavke := querybuilder.New(`delete from epro`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, stavke, ""); err != nil {
		return err
	}
	if _, err := r.exec(ctx, tx, "stavke zahteva za nabavku", stavke); err != nil {
		return err
	}
	b := querybuilder.New(`delete from edok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "zahtev za nabavku", b)
}

// KomZahtevStavke returns the stavke of a zahtev za nabavku of the period in the order of their redni broj, with the
// naziv and the JM of their artikal (the legacy PR_QRY_EPRO: epro joined with rsif by rsifid).
func (r *KomercijalnoRepository) KomZahtevStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rdokid, 0)::bigint as rdokid, coalesce(p.rbr, 0)::int as rbr,
			coalesce(p.sifra, 0)::int as sifra, coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm,
			coalesce(p.trazkolic, 0)::float8 as kolicina, coalesce(p.cena, 0)::float8 as cena,
			coalesce(p.vrd, 0)::int as vrd, coalesce(p.dokum, 0)::int as dokum
		from epro p`).
		AddJoin("join rsif s on s.rsifid = p.rsifid").
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []komZahtevStavkaDto
	if err := r.selectRows(ctx, "stavke zahteva za nabavku", &rows, b); err != nil {
		return nil, err
	}
	stavke := make([]domain.KomZahtevStavkaRow, len(rows))
	for i, row := range rows {
		stavke[i] = domain.KomZahtevStavkaRow{RproID: row.RproID, RdokID: row.RdokID, Rbr: row.Rbr, Sifra: row.Sifra, Naziv: row.Naziv,
			JM: row.JM, Kolicina: row.Kolicina, Cena: row.Cena, DokumVrd: row.Vrd, DokumBroj: row.Dokum}
	}
	return stavke, nil
}

// komZahtevStavkaDto is a row of KomZahtevStavke.
type komZahtevStavkaDto struct {
	RproID   int64   `db:"rproid"`
	RdokID   int64   `db:"rdokid"`
	Rbr      int     `db:"rbr"`
	Sifra    int     `db:"sifra"`
	Naziv    string  `db:"naziv"`
	JM       string  `db:"jm"`
	Kolicina float64 `db:"kolicina"`
	Cena     float64 `db:"cena"`
	Vrd      int     `db:"vrd"`
	Dokum    int     `db:"dokum"`
}

// KomZahtevStavka returns a stavka of a zahtev za nabavku of the period; found is false without it.
func (r *KomercijalnoRepository) KomZahtevStavka(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaRow, bool, error) {
	b := querybuilder.New(`select p.rproid::bigint, coalesce(p.rdokid, 0)::bigint, coalesce(p.rbr, 0)::int, coalesce(p.sifra, 0)::int,
			coalesce(s.naziv, ''), coalesce(s.jm, ''), coalesce(p.trazkolic, 0)::float8, coalesce(p.cena, 0)::float8
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rproid = ?", rproID).
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return domain.KomZahtevStavkaRow{}, false, err
	}
	var st domain.KomZahtevStavkaRow
	err := r.row(ctx, nil, "stavka zahteva za nabavku", b, &st.RproID, &st.RdokID, &st.Rbr, &st.Sifra, &st.Naziv, &st.JM, &st.Kolicina, &st.Cena)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomZahtevStavkaRow{}, false, nil
	}
	if err != nil {
		return domain.KomZahtevStavkaRow{}, false, err
	}
	return st, true, nil
}

// KomZahtevNoviRbr returns the redni broj of a new stavka of a zahtev za nabavku (the legacy NoviRBR: the largest
// rbr of its stavke + 1).
func (r *KomercijalnoRepository) KomZahtevNoviRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error) {
	b := querybuilder.New(`select coalesce(max(rbr), 0)::int + 1 from epro`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var rbr int
	if err := r.row(ctx, tx, "redni broj stavke zahteva za nabavku", b, &rbr); err != nil {
		return 0, err
	}
	return rbr, nil
}

// KomZahtevStavkaDodaj adds a stavka of a zahtev za nabavku (the legacy AddEPRO): the artikal with its nabavna cena
// as the cena and the ncena, the tražena količina, the magacin and the vrsta, broj and datum of the document, a
// slobodna stavka (ststatus S), with the audit columns.
func (r *KomercijalnoRepository) KomZahtevStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomZahtevStavkaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, cena, ncena, mag, magaciniid, trazkolic, ststatus,
			vrd, dokum, dadok, xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'S', ?, ?, ?, now(), ?)`,
		s.God, s.Kar, st.RdokID, st.Sifra, st.Po, st.RsifID, st.Rbr, st.Cena, st.Cena, st.Mag, st.MagaciniID, st.Kolicina,
		st.Vrd, st.Dokum, st.Dadok, s.UserName)
	_, err = r.exec(ctx, tx, "nova stavka zahteva za nabavku", b)
	return err
}

// KomZahtevStavkaIzmeni saves a changed stavka of a zahtev za nabavku of the period (the artikal with its nabavna
// cena and the tražena količina, the legacy change of the količina in the grid), with the audit columns;
// ErrNotFound without it.
func (r *KomercijalnoRepository) KomZahtevStavkaIzmeni(ctx context.Context, tx db.Transaction, st domain.KomZahtevStavkaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("sifra = ?", st.Sifra).
		Set("po = ?", st.Po).
		Set("rsifid = ?", st.RsifID).
		Set("cena = ?", st.Cena).
		Set("ncena = ?", st.Cena).
		Set("trazkolic = ?", st.Kolicina).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", st.RproID).
		Where("rdokid = ?", st.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavka zahteva za nabavku", b)
}

// KomZahtevStavkaBrisi deletes a stavka of a zahtev za nabavku of the period and moves the redni brojevi of the
// stavke after it one back (the legacy BTN_BrisiDET); ErrNotFound without it.
func (r *KomercijalnoRepository) KomZahtevStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error {
	b := querybuilder.New(`delete from epro`).Where("rproid = ?", rproID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	if err := r.execOne(ctx, tx, "stavka zahteva za nabavku", b); err != nil {
		return err
	}
	pomeri := querybuilder.NewUpdate("epro").Set("rbr = rbr - 1").Where("rdokid = ?", rdokID).Where("rbr > ?", rbr)
	if err := inPeriod(ctx, pomeri, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "redni brojevi stavki zahteva za nabavku", pomeri)
	return err
}

// KomArtikal returns the artikal of the šifra of the period (rsif; the first by its id) of the tipovi S, A, R, D and
// K with its nabavna cena (the first rcene of the šifra, 0 without it), the legacy exit of EDT_SIFRA; found is false
// without it (an artikal of another tip too).
func (r *KomercijalnoRepository) KomArtikal(ctx context.Context, sifra int) (domain.KomArtikal, bool, error) {
	b := querybuilder.New(`select r.rsifid::bigint, coalesce(r.sifra, 0)::int, coalesce(r.naziv, ''), coalesce(r.jm, ''), coalesce(r.po, 0)::int,
			coalesce((select c.ncena from rcene c where c.god = r.god and c.kar = r.kar and c.sifra = r.sifra order by c.rceneid limit 1), 0)::float8
		from rsif r`).
		Where("r.sifra = ?", sifra).
		Where("upper(trim(coalesce(r.tip, ''))) in (?, ?, ?, ?, ?)", querybuilder.Args(komZahtevTipoviArtikla)...)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return domain.KomArtikal{}, false, err
	}
	b.OrderBy("r.rsifid").Limit(1)
	var a domain.KomArtikal
	err := r.row(ctx, nil, "artikal", b, &a.RsifID, &a.Sifra, &a.Naziv, &a.JM, &a.Po, &a.NCena)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomArtikal{}, false, nil
	}
	if err != nil {
		return domain.KomArtikal{}, false, err
	}
	return a, true, nil
}

// KomKomercijalista returns the komercijalista of the šifra of the period; found is false without it.
func (r *KomercijalnoRepository) KomKomercijalista(ctx context.Context, sifra int) (domain.KomKomercijalista, bool, error) {
	b := querybuilder.New(`select k.komid::bigint, coalesce(k.sifkom, 0)::int, coalesce(k.imeprezime, '') from komercijalisti k`).
		Where("k.sifkom = ?", sifra)
	return r.komKomercijalista(ctx, b)
}

// KomKomercijalistaKorisnika returns the komercijalista of the period whose ime i prezime is the name of the user of
// the session (appusers.name, the legacy gpwGetUserInfo(gpwInfoName)); found is false without it.
func (r *KomercijalnoRepository) KomKomercijalistaKorisnika(ctx context.Context, userID int64) (domain.KomKomercijalista, bool, error) {
	b := querybuilder.New(`select k.komid::bigint, coalesce(k.sifkom, 0)::int, coalesce(k.imeprezime, '') from komercijalisti k`).
		AddJoin("join appusers u on upper(trim(u.name)) = upper(trim(k.imeprezime))").
		Where("u.id = ?", userID).
		Where("trim(coalesce(u.name, '')) <> ''")
	return r.komKomercijalista(ctx, b)
}

func (r *KomercijalnoRepository) komKomercijalista(ctx context.Context, b *querybuilder.Builder) (domain.KomKomercijalista, bool, error) {
	if err := inPeriod(ctx, b, "k."); err != nil {
		return domain.KomKomercijalista{}, false, err
	}
	b.OrderBy("k.komid").Limit(1)
	var k domain.KomKomercijalista
	err := r.row(ctx, nil, "komercijalista", b, &k.KomID, &k.Sifra, &k.Ime)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomKomercijalista{}, false, nil
	}
	if err != nil {
		return domain.KomKomercijalista{}, false, err
	}
	return k, true, nil
}

// KomMagacin returns the oznaka (mag) of a magacin of the period; found is false without it.
func (r *KomercijalnoRepository) KomMagacin(ctx context.Context, magaciniID int) (int, bool, error) {
	b := querybuilder.New(`select coalesce(mag, 0)::int from magacini`).Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	var mag int
	err := r.row(ctx, nil, "magacin", b, &mag)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return mag, true, nil
}

// komPretraga restricts the list of a main screen of the komercijalno to the documents whose columns (written in code,
// never from the request) contain the text of the search, ignoring the case; nothing without a text.
func komPretraga(b *querybuilder.Builder, tekst string, kolone ...string) {
	tekst = strings.TrimSpace(tekst)
	if tekst == "" || len(kolone) == 0 {
		return
	}
	uslovi := make([]string, len(kolone))
	args := make([]any, len(kolone))
	for i, k := range kolone {
		uslovi[i] = "coalesce(" + k + ", '') ilike ?"
		args[i] = "%" + tekst + "%"
	}
	b.Where("("+strings.Join(uslovi, " or ")+")", args...)
}

// KomZahteviPrikaz returns the zahtevi za nabavku of the tab "Pregled zahteva za nabavku" (the legacy PuniHDRZAHTEV):
// the ones of the vrsta and the status in the magacini of the user of the session (maguser), of one magacin or of all
// of them (sviMagacini), the last ones first, with the "mag-opis" of the magacin, the "šifra-ime" of the komercijalista
// and the "oznaka-naziv" of the org. jedinica.
func (r *KomercijalnoRepository) KomZahteviPrikaz(ctx context.Context, p domain.KomZahtevPrikazParams) ([]domain.KomZahtevPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid,
			coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin, coalesce(e.vrd, 0)::int as vrd,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, coalesce(e.foot, '') as foot,
			coalesce(k.sifkom::text || '-' || coalesce(k.imeprezime, ''), '') as referent,
			coalesce(o.ojozn::text || '-' || coalesce(o.naziv, ''), '') as oj
		from edok e`).
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		Where("e.vrd = ?", p.Vrd).
		Where("coalesce(e.doksts, '') = ?", p.Doksts)
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.foot", "k.imeprezime", "o.naziv")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok desc nulls last, e.dokum desc")
	var rows []komZahtevPrikazDto
	if err := r.selectRows(ctx, "pregled zahteva za nabavku", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomZahtevPrikazRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomZahtevPrikazRow{RdokID: row.RdokID, Magacin: row.Magacin, Vrd: row.Vrd, Dokum: row.Dokum, Dadok: row.Dadok,
			Foot: row.Foot, Referent: row.Referent, OJ: row.OJ}
	}
	return out, nil
}

// komZahtevPrikazDto is a row of KomZahteviPrikaz.
type komZahtevPrikazDto struct {
	RdokID   int64        `db:"rdokid"`
	Magacin  string       `db:"magacin"`
	Vrd      int          `db:"vrd"`
	Dokum    int          `db:"dokum"`
	Dadok    sql.NullTime `db:"dadok"`
	Foot     string       `db:"foot"`
	Referent string       `db:"referent"`
	OJ       string       `db:"oj"`
}

// KomZahtevStampaZaglavlje returns what the print of a zahtev za nabavku shows of its header (the legacy
// PR_RPT_ZAHTEV_NABAVKA): its broj and datum, its magacin ("mag-opis") with the adresa of the magacin, its org.
// jedinica ("oznaka-naziv") and whether it is nepotvrđen; found is false without it.
func (r *KomercijalnoRepository) KomZahtevStampaZaglavlje(ctx context.Context, rdokID int64) (domain.KomZahtevStampa, bool, error) {
	b := querybuilder.New(`select coalesce(e.dokum, 0)::int, coalesce(to_char(e.dadok, 'DD.MM.YYYY'), ''),
			coalesce(e.mag::text || '-' || m.opis, ''), coalesce(m.adresa, ''),
			coalesce(o.ojozn::text || '-' || coalesce(o.naziv, ''), ''), coalesce(e.doksts, '') = ?
		from edok e`, domain.KomZahtevNepotvrdjen).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		Where("e.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomZahtevStampa{}, false, err
	}
	var z domain.KomZahtevStampa
	err := r.row(ctx, nil, "zaglavlje zahteva za nabavku", b, &z.Dokum, &z.Datum, &z.Magacin, &z.MagacinAdresa, &z.OJ, &z.Nepotvrdjen)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomZahtevStampa{}, false, nil
	}
	if err != nil {
		return domain.KomZahtevStampa{}, false, err
	}
	return z, true, nil
}

// KomFirma returns the firm of the session (fvr of the period with the naziv of the firm of the session, like the
// prints of the robno) for the print: its naziv, adresa, PDV, APR, PIB, matični broj, šifra delatnosti, telefon and
// logo.
func (r *KomercijalnoRepository) KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error) {
	b := querybuilder.New(`select coalesce(f.naziv, ''), coalesce(f.adresa, ''), coalesce(nullif(f.pobro, 0)::text, ''), coalesce(f.mesto, ''),
			coalesce(f.pib, ''), coalesce(f.matbr, ''), coalesce(f.sifdel, ''), coalesce(f.tel, ''), coalesce(f.apr, ''),
			coalesce(f.obv, false), coalesce(f.brobvpdv, ''), f.logofirme
		from fvr f`).
		Where("f.naziv = ?", firma)
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.RobnoStampaFakturaFirmaDto{}, err
	}
	b.Limit(1)
	var f domain.RobnoStampaFakturaFirmaDto
	if err := r.row(ctx, nil, "firma", b, &f.Naziv, &f.Adresa, &f.Pobro, &f.Mesto, &f.Pib, &f.Matbr, &f.Sifdel, &f.Tel, &f.Apr, &f.Obv,
		&f.Brobvpdv, &f.Logo); err != nil {
		return domain.RobnoStampaFakturaFirmaDto{}, err
	}
	return f, nil
}

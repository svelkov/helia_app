package robno

import (
	"context"
	"fmt"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

type KrajPoslovneGodineService interface {
	GetMagacini(context.Context) ([]domain.ComboItem, error)
	GetPopis(context.Context, domain.KrajPoslovneGodineParams) (domain.KrajPopisData, error)
	GetVisakManjak(context.Context, domain.KrajPoslovneGodineParams) ([]domain.KrajVisakManjakRow, error)
	PrepisStanja(context.Context, domain.KrajPoslovneGodineParams, int) error
	ValidateRSTA(domain.KrajPoslovneGodineParams) []domain.FieldError
	ValidateObrada(domain.KrajPoslovneGodineParams) []domain.FieldError
	GetType1Fields() []domain.Fields
	GetType2Fields() []domain.Fields
	GetVisakManjakFields() []domain.Fields
}

type KrajPoslovneGodineResource struct {
	database    db.Database
	mag         *repository.BaseRepository[domain.Magacini]
	type1Fields []domain.Fields
	type2Fields []domain.Fields
	visakFields []domain.Fields
}

func NewKrajPoslovneGodineService(database db.Database, mag *repository.BaseRepository[domain.Magacini]) *KrajPoslovneGodineResource {
	s := &KrajPoslovneGodineResource{database: database, mag: mag}
	s.type1Fields = []domain.Fields{
		{Name: "redbr", Label: "Red.br.", Field: "redbr"}, {Name: "konto", Label: "Konto", Field: "konto"}, {Name: "sifra", Label: "Šifra", Field: "sifra"}, {Name: "naziv", Label: "Naziv", Field: "naziv"}, {Name: "jm", Label: "JM", Field: "jm"}, {Name: "grupa", Label: "Grupa", Field: "grupa"}, {Name: "nazivgrupe", Label: "Naziv grupe", Field: "nazivgrupe"}, {Name: "cena", Label: "Cena", Field: "cena", TextAlign: "right"}, {Name: "kolicina1", Label: "Količina 1", Field: "kolicina1", TextAlign: "right"}, {Name: "kolicina2", Label: "Količina 2", Field: "kolicina2", TextAlign: "right"}, {Name: "kolicina3", Label: "Količina 3", Field: "kolicina3", TextAlign: "right"}, {Name: "ukupnakolicina", Label: "Ukupna količina", Field: "ukupnakolicina", TextAlign: "right"}, {Name: "stanjezaliha", Label: "Stanje zaliha", Field: "stanjezaliha", TextAlign: "right"},
	}
	s.type2Fields = []domain.Fields{
		{Name: "redbr", Label: "Red.br.", Field: "redbr"}, {Name: "konto", Label: "Konto", Field: "konto"}, {Name: "sifra", Label: "Šifra", Field: "sifra"}, {Name: "naziv", Label: "Naziv", Field: "naziv"}, {Name: "jm", Label: "JM", Field: "jm"}, {Name: "grupa", Label: "Grupa", Field: "grupa"}, {Name: "nazivgrupe", Label: "Naziv grupe", Field: "nazivgrupe"}, {Name: "otk", Label: "OTK", Field: "otk"}, {Name: "serija", Label: "Serija", Field: "serija"}, {Name: "rok", Label: "Rok", Field: "rok"}, {Name: "kolicina1", Label: "Količina 1", Field: "kolicina1", TextAlign: "right"}, {Name: "kolicina2", Label: "Količina 2", Field: "kolicina2", TextAlign: "right"}, {Name: "kolicina3", Label: "Količina 3", Field: "kolicina3", TextAlign: "right"}, {Name: "ukupnakolicina", Label: "Ukupna količina", Field: "ukupnakolicina", TextAlign: "right"}, {Name: "stanjezaliha", Label: "Stanje zaliha", Field: "stanjezaliha", TextAlign: "right"},
	}
	s.visakFields = []domain.Fields{
		{Name: "redbr", Label: "Red.br.", Field: "redbr"}, {Name: "sifra", Label: "Šifra", Field: "sifra"}, {Name: "konto", Label: "Konto", Field: "konto"}, {Name: "naziv", Label: "Naziv", Field: "naziv"}, {Name: "jm", Label: "JM", Field: "jm"}, {Name: "cena", Label: "Cena", Field: "cena", TextAlign: "right"}, {Name: "kolicinapopisa", Label: "Popisano stanje", Field: "kolicinapopisa", TextAlign: "right"}, {Name: "iznospopisa", Label: "Iznos popisa", Field: "iznospopisa", TextAlign: "right"}, {Name: "stanjeknjigovodstveno", Label: "Knjigovodstveno stanje", Field: "stanjeknjigovodstveno", TextAlign: "right"}, {Name: "iznosknjigovodstveno", Label: "Knjigovodstveni iznos", Field: "iznosknjigovodstveno", TextAlign: "right"}, {Name: "visak", Label: "Višak", Field: "visak", TextAlign: "right"}, {Name: "iznosviska", Label: "Iznos viška", Field: "iznosviska", TextAlign: "right"}, {Name: "manjak", Label: "Manjak", Field: "manjak", TextAlign: "right"}, {Name: "iznosmanjka", Label: "Iznos manjka", Field: "iznosmanjka", TextAlign: "right"}, {Name: "finansijskivisak", Label: "Finansijski višak", Field: "finansijskivisak", TextAlign: "right"}, {Name: "finansijskimanjak", Label: "Finansijski manjak", Field: "finansijskimanjak", TextAlign: "right"},
	}
	return s
}

func (s *KrajPoslovneGodineResource) GetType1Fields() []domain.Fields       { return s.type1Fields }
func (s *KrajPoslovneGodineResource) GetType2Fields() []domain.Fields       { return s.type2Fields }
func (s *KrajPoslovneGodineResource) GetVisakManjakFields() []domain.Fields { return s.visakFields }

func (s *KrajPoslovneGodineResource) ValidateRSTA(p domain.KrajPoslovneGodineParams) []domain.FieldError {
	var errors []domain.FieldError
	if p.Magacin <= 0 {
		errors = append(errors, domain.FieldError{Field: "magacin", ErrorMessage: "obavezan podatak"})
	}
	if p.TipZal != 1 && p.TipZal != 2 {
		errors = append(errors, domain.FieldError{Field: "tipzal", ErrorMessage: "dozvoljene vrednosti su 1 ili 2"})
	}
	if p.Cena < 1 || p.Cena > 4 {
		errors = append(errors, domain.FieldError{Field: "cena", ErrorMessage: "dozvoljene vrednosti su 1-4"})
	}
	if p.NovaGod <= 0 {
		errors = append(errors, domain.FieldError{Field: "novagod", ErrorMessage: "obavezan podatak"})
	}
	if p.OdSifre != 0 && p.DoSifre != 0 && p.OdSifre > p.DoSifre {
		errors = append(errors, domain.FieldError{Field: "odsifre", ErrorMessage: "neispravan opseg"})
	}
	if p.OdKonta != "" && p.DoKonta != "" && p.OdKonta > p.DoKonta {
		errors = append(errors, domain.FieldError{Field: "odkonta", ErrorMessage: "neispravan opseg"})
	}
	return errors
}

func (s *KrajPoslovneGodineResource) ValidateObrada(p domain.KrajPoslovneGodineParams) []domain.FieldError {
	var errors []domain.FieldError
	if p.Magacin <= 0 {
		errors = append(errors, domain.FieldError{Field: "magacin", ErrorMessage: "obavezan podatak"})
	}
	if p.NovaGod <= 0 {
		errors = append(errors, domain.FieldError{Field: "novagod", ErrorMessage: "obavezan podatak"})
	}
	if p.Nalog <= 0 {
		errors = append(errors, domain.FieldError{Field: "nalog", ErrorMessage: "obavezan podatak"})
	}
	if p.Dokum <= 0 {
		errors = append(errors, domain.FieldError{Field: "dokum", ErrorMessage: "obavezan podatak"})
	}
	if p.Vrd <= 0 {
		errors = append(errors, domain.FieldError{Field: "vrd", ErrorMessage: "obavezan podatak"})
	}
	return errors
}

func (s *KrajPoslovneGodineResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	user := domain.GetSessionFromStdContext(ctx)
	if user == nil {
		return nil, fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder("SELECT magaciniid, mag, opis FROM magacini", true)
	god, kar := s.mag.GetHasGodHasKar()
	if god {
		qb.AddEqual("god", user.SelectedGod)
	}
	if kar {
		qb.AddEqual("kar", user.SelectedKar)
	}
	qb.AddOrderBy("mag")
	query, args := qb.Build()
	rows, err := s.mag.GetAllCustom(ctx, query, "", args, "", "")
	if err != nil {
		return nil, err
	}
	result := make([]domain.ComboItem, 0, len(*rows))
	for _, row := range *rows {
		result = append(result, domain.ComboItem{Key: fmt.Sprint(row.Mag), Value: fmt.Sprintf("%d - %s", row.Mag, row.Opis)})
	}
	return result, nil
}

func (s *KrajPoslovneGodineResource) GetPopis(ctx context.Context, p domain.KrajPoslovneGodineParams) (domain.KrajPopisData, error) {
	user := domain.GetSessionFromStdContext(ctx)
	if user == nil {
		return domain.KrajPopisData{}, fmt.Errorf("no user session found")
	}
	tipzal := p.TipZal
	if tipzal == 0 {
		if err := s.database.GetContext(ctx, &tipzal, `SELECT tipzal FROM magacini WHERE mag=$1 AND god=$2 AND kar=$3`, p.Magacin, user.SelectedGod, user.SelectedKar); err != nil {
			return domain.KrajPopisData{}, fmt.Errorf("get TIPZAL: %w", err)
		}
	}
	if tipzal == 1 {
		return s.getPopisType1(ctx, p, user.SelectedGod, user.SelectedKar)
	}
	if tipzal == 2 {
		return s.getPopisType2(ctx, p, user.SelectedGod, user.SelectedKar)
	}
	return domain.KrajPopisData{}, fmt.Errorf("unsupported TIPZAL %d; expected 1 or 2", tipzal)
}

func (s *KrajPoslovneGodineResource) getPopisType1(ctx context.Context, p domain.KrajPoslovneGodineParams, god, kar int) (domain.KrajPopisData, error) {
	if err := s.requireTables(ctx, "rsta", "rcene", "rsif", "rgru"); err != nil {
		return domain.KrajPopisData{}, err
	}
	order, err := popisOrder(p.Tip)
	if err != nil {
		return domain.KrajPopisData{}, err
	}
	query := fmt.Sprintf(`SELECT row_number() over (ORDER BY %s)::int redbr,s.konto,s.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,coalesce(a.gru,0) grupa,coalesce(g.naziv,'') nazivgrupe,
	 CASE WHEN coalesce(m.nacvodzal,0)=3 THEN coalesce(max(s.prosnc),0) ELSE coalesce(max(c.cena),0) END cena,
	 0::numeric kolicina1,0::numeric kolicina2,0::numeric kolicina3,0::numeric ukupnakolicina,
	 coalesce(sum(s.ulaz-s.izlaz),0) stanjezaliha
	 FROM rsta s LEFT JOIN rcene c ON c.sifra=s.sifra AND c.mag=s.mag AND c.god=s.god AND c.kar=s.kar
 LEFT JOIN rsif a ON a.sifra=s.sifra AND a.god=s.god AND a.kar=s.kar LEFT JOIN rgru g ON g.gru=a.gru AND g.god=s.god AND g.kar=s.kar
 LEFT JOIN magacini m ON m.mag=s.mag AND m.god=s.god AND m.kar=s.kar WHERE s.god=$1 AND s.kar=$2 AND ($3=0 OR s.mag=$3) AND ($4='' OR s.konto>=$4) AND ($5='' OR s.konto<=$5) AND ($6=0 OR s.sifra>=$6) AND ($7=0 OR s.sifra<=$7)
	 GROUP BY s.konto,s.sifra,a.naziv,a.jm,a.gru,g.naziv,m.nacvodzal HAVING ($8 OR sum(s.ulaz-s.izlaz)<>0) ORDER BY %s`, order, order)
	var rows []domain.KrajPopisType1Row
	err = s.database.SelectContext(ctx, &rows, query, god, kar, p.Magacin, p.OdKonta, p.DoKonta, p.OdSifre, p.DoSifre, p.ObradiNule)
	return domain.KrajPopisData{TipZal: 1, Type1: rows}, err
}

func (s *KrajPoslovneGodineResource) getPopisType2(ctx context.Context, p domain.KrajPoslovneGodineParams, god, kar int) (domain.KrajPopisData, error) {
	if err := s.requireTables(ctx, "drsta", "rsta", "rsif", "rgru"); err != nil {
		return domain.KrajPopisData{}, err
	}
	order, err := popisOrder(p.Tip)
	if err != nil {
		return domain.KrajPopisData{}, err
	}
	order = strings.ReplaceAll(order, "s.", "d.")
	query := fmt.Sprintf(`SELECT row_number() over (ORDER BY %s)::int redbr,d.konto,d.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,coalesce(a.gru,0) grupa,coalesce(g.naziv,'') nazivgrupe,
	 coalesce(d.otk,'') otk,coalesce(d.serija,'') serija,coalesce(d.roktr,'') rok,0::numeric kolicina1,0::numeric kolicina2,0::numeric kolicina3,0::numeric ukupnakolicina,
	 coalesce((SELECT sum(s.ulaz-s.izlaz) FROM rsta s WHERE s.god=$1 AND s.kar=$2 AND s.mag=d.mag AND s.sifra=d.sifra AND s.konto=d.konto),0) stanjezaliha
 FROM drsta d LEFT JOIN rsif a ON a.sifra=d.sifra AND a.god=d.god AND a.kar=d.kar LEFT JOIN rgru g ON g.gru=a.gru AND g.god=d.god AND g.kar=d.kar WHERE d.god=$1 AND d.kar=$2 AND ($3=0 OR d.mag=$3) AND ($4='' OR d.konto>=$4) AND ($5='' OR d.konto<=$5) AND ($6=0 OR d.sifra>=$6) AND ($7=0 OR d.sifra<=$7) ORDER BY %s`, order, order)
	var rows []domain.KrajPopisType2Row
	err = s.database.SelectContext(ctx, &rows, query, god, kar, p.Magacin, p.OdKonta, p.DoKonta, p.OdSifre, p.DoSifre)
	return domain.KrajPopisData{TipZal: 2, Type2: rows}, err
}

func popisOrder(tip string) (string, error) {
	switch tip {
	case "", "sifra":
		return "s.sifra", nil
	case "naziv":
		return "coalesce(a.naziv,''),s.sifra", nil
	case "grupa":
		return "coalesce(a.gru,0),s.sifra", nil
	case "grupa-naziv":
		return "coalesce(a.gru,0),coalesce(a.naziv,''),s.sifra", nil
	default:
		return "", fmt.Errorf("unknown popis sort %q", tip)
	}
}

func (s *KrajPoslovneGodineResource) GetVisakManjak(ctx context.Context, p domain.KrajPoslovneGodineParams) ([]domain.KrajVisakManjakRow, error) {
	user := domain.GetSessionFromStdContext(ctx)
	if user == nil {
		return nil, fmt.Errorf("no user session found")
	}
	if err := s.requireTables(ctx, "rsta", "rpro", "rdok", "rsif", "rgru"); err != nil {
		return nil, err
	}
	query := `WITH book AS (SELECT s.konto,s.sifra,sum(s.ulaz-s.izlaz) stanje,sum(s.dug-s.pot) iznos,max(s.cena) cena FROM rsta s WHERE s.god=$1 AND s.kar=$2 AND ($7=0 OR s.mag=$7) GROUP BY s.konto,s.sifra), popis AS (SELECT p.konto,p.sifra,sum(p.kolic) kolicina,sum(p.iznos) iznos FROM rpro p JOIN rdok d ON d.rdokid=p.rdokid WHERE p.god=$3 AND p.kar=$2 AND p.nalog=$4 AND p.dokum=$5 AND p.vrd=$6 AND ($7=0 OR p.mag=$7) GROUP BY p.konto,p.sifra), data_rows AS (SELECT b.konto,b.sifra,coalesce(a.naziv,'') naziv,coalesce(a.jm,'') jm,b.cena,b.iznos,b.stanje,coalesce(p.kolicina,0) kolicina,coalesce(p.iznos,0) popis_iznos,coalesce(p.kolicina,0)-b.stanje razlika FROM book b LEFT JOIN popis p ON p.konto=b.konto AND p.sifra=b.sifra LEFT JOIN rsif a ON a.sifra=b.sifra AND a.god=$1 AND a.kar=$2) SELECT row_number() over (ORDER BY sifra)::int redbr,sifra,konto,naziv,jm,cena,kolicina kolicinapopisa,popis_iznos iznospopisa,stanje stanjeknjigovodstveno,iznos iznosknjigovodstveno,case when razlika>0 then razlika else 0 end visak,case when razlika>0 then (popis_iznos-iznos) else 0 end iznosviska,case when razlika<0 then -razlika else 0 end manjak,case when razlika<0 then (iznos-popis_iznos) else 0 end iznosmanjka,case when razlika>0 and popis_iznos-iznos>0 then popis_iznos-iznos else 0 end finansijskivisak,case when razlika<0 and iznos-popis_iznos>0 then iznos-popis_iznos else 0 end finansijskimanjak FROM data_rows WHERE razlika<>0 ORDER BY sifra`
	var rows []domain.KrajVisakManjakRow
	err := s.database.SelectContext(ctx, &rows, query, user.SelectedGod, user.SelectedKar, p.NovaGod, p.Nalog, p.Dokum, p.Vrd, p.Magacin)
	return rows, err
}

func (s *KrajPoslovneGodineResource) PrepisStanja(ctx context.Context, p domain.KrajPoslovneGodineParams, targetYear int) error {
	if targetYear <= 0 {
		return fmt.Errorf("invalid target year")
	}
	price, err := prepisPriceColumn(p.Cena)
	if err != nil {
		return err
	}
	tx, err := s.database.Beginx()
	if err != nil {
		return fmt.Errorf("begin prepis transaction: %w", err)
	}
	defer tx.Rollback()
	for _, table := range []string{"rsta", "drsta", "rcene", "rdok", "rnal", "sr", "rpro", "rsal", "rsin", "dokvrsta"} {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists); err != nil {
			return fmt.Errorf("verify prepis table %s: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("prepis schema unavailable: required table %s does not exist", strings.ToUpper(table))
		}
	}
	return fmt.Errorf("prepis schema unavailable for price %s: WinDev column mapping is not present; no rows were written", price)
}
func prepisPriceColumn(choice int) (string, error) {
	switch choice {
	case 1:
		return "CENA", nil
	case 2:
		return "PROSNC", nil
	case 3:
		return "NCENA", nil
	case 4:
		return "VPCENA", nil
	default:
		return "", fmt.Errorf("invalid prepis price option %d; expected 1-4", choice)
	}
}
func (s *KrajPoslovneGodineResource) requireTables(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		var exists bool
		if err := s.database.GetContext(ctx, &exists, `SELECT to_regclass($1) IS NOT NULL`, table); err != nil {
			return fmt.Errorf("verify %s schema: %w", strings.ToUpper(table), err)
		}
		if !exists {
			return fmt.Errorf("schema unavailable: required table %s does not exist", strings.ToUpper(table))
		}
	}
	return nil
}

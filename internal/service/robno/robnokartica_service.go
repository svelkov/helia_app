package robno

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

// RobnoKarticaService exposes the two inventory-card views used by the Robno module.
type RobnoKarticaService interface {
	GetKarticaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams, printType string) error
	GetKarticaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams, printType string) error
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	ValidacijaSubsintetickogKonta(params domain.RobnoKarticaParams) []domain.FieldError
	ValidacijaKarticaArtikla(params domain.RobnoKarticaParams) []domain.FieldError
	GetKarticaArtiklaTableFields() []domain.Fields
	GetKarticaArtiklaStampaTableFields() []domain.Fields
	GetKarticaSubsintetickogKontaTableFields() []domain.Fields
	GetKarticaSubsintetickogKontaStampaTableFields() []domain.Fields
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// RobnoKarticaResource reuses the established Promet queries while keeping the
// Robno handler independent from the larger Finance service interface.
type RobnoKarticaResource struct {
	robnoKarticaRepo                      *repository.BaseRepository[domain.RobnoKarticaDto]
	rproRepo                              *repository.BaseRepository[domain.Rpro]
	magRepo                               *repository.BaseRepository[domain.Magacini]
	fvrRepo                               *repository.BaseRepository[domain.Fvr]
	commonSvc                             commonsvc.CommonService
	karticaArtiklaTableFields             []domain.Fields
	karticaArtiklaStampaTableFields       []domain.Fields
	subsintetickaKarticaKontaTableFields  []domain.Fields
	subsintetickaKarticaStampaTableFields []domain.Fields
}

func NewRobnoKarticaService(robnoKarticaRepo *repository.BaseRepository[domain.RobnoKarticaDto], rproRepo *repository.BaseRepository[domain.Rpro], magRepo *repository.BaseRepository[domain.Magacini], fvrRepo *repository.BaseRepository[domain.Fvr], commonSvc commonsvc.CommonService) *RobnoKarticaResource {
	rs := &RobnoKarticaResource{
		robnoKarticaRepo: robnoKarticaRepo,
		rproRepo:         rproRepo,
		magRepo:          magRepo,
		fvrRepo:          fvrRepo,
		commonSvc:        commonSvc,
	}
	rs.setTableFields()
	return rs
}

func (s *RobnoKarticaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}
func (s *RobnoKarticaResource) GetKarticaArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	if printType == common.TipStampePrint {
		tbl.Headers = s.karticaArtiklaStampaTableFields
	} else {
		if printType == common.TipStampePrint {
			tbl.Headers = s.karticaArtiklaStampaTableFields
		} else {
			tbl.Headers = s.karticaArtiklaTableFields
		}
	}
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		SELECT
			rnal.tipdok,
			rnal.nalog,
			rnal.danal,
			CONCAT(rdok.vrd, '-', dokvrsta.opis) as vrd,
			rdok.dokum,
			rdok.dadok,
			rpro.konto,
			rpro.sifra,
			rpro.fcena,
			COALESCE(rsif.naziv, '') AS naziv_artikla,
			CASE WHEN upper(dokvrsta.kodknj) = 'D' AND dokvrsta.grpdok <> 'NIV' THEN rpro.kolic ELSE 0 END AS ulaz,
			CASE WHEN upper(dokvrsta.kodknj) = 'P' AND dokvrsta.grpdok <> 'NIV' THEN rpro.kolic ELSE 0 END AS izlaz,
			CASE WHEN upper(dokvrsta.kodknj) = 'D' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END AS duguje,
			CASE WHEN upper(dokvrsta.kodknj) = 'P' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END AS potrazuje,
			CASE WHEN dokvrsta.grpdok = 'NIV' THEN 0 ELSE rpro.iznos END AS iznos,
			rdok.fkto,
			rdok.fana,
			COALESCE(fkpl.naziv, '') AS fkplnaz,
			COALESCE(CAST(rdok.sifval AS text) || '-' || valute.naziv, CAST(rdok.sifval AS text), '') AS valuta,
			rdok.kurs,
			rpro.cenaval,
			EXTRACT(MONTH FROM rnal.danal)::int AS mesec,
			rdok.dokiz,
			rdok.datiz AS dadokiz,
			rpro.otk,
			rpro.serija,
			rpro.roktr
		FROM rnal`, true)
	qb.AddJoin("inner join rdok ON rdok.rnalid = rnal.rnalid")
	qb.AddJoin("inner join rpro ON rpro.rdokid = rdok.rdokid")
	qb.AddJoin("inner join rsif ON rsif.rsifid = rpro.rsifid")
	qb.AddJoin("inner join dokvrsta ON rdok.god = dokvrsta.god AND rdok.kar = dokvrsta.kar AND rdok.vrd = dokvrsta.vrd")
	qb.AddJoin("left join fkpl ON fkpl.god = rdok.god AND fkpl.kar = rdok.kar AND fkpl.vkonta = 1 AND fkpl.konto = rpro.fkto AND fkpl.sifra = rpro.fana")
	qb.AddJoin("left join valute ON valute.sifval = rdok.sifval")
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	qb.AddEqual("rnal.magaciniid", params.Magacin)
	qb.AddCondition("rpro.sifra", params.OdSifre, ">=")
	qb.AddCondition("rpro.sifra", params.DoSifre, "<=")
	if params.CbxBrojNaloga {
		values := make([]any, 0)
		for _, part := range strings.Split(params.Nalozi, ",") {
			nalog, parseErr := strconv.Atoi(strings.TrimSpace(part))
			if parseErr != nil {
				return fmt.Errorf("invalid nalog list: %w", parseErr)
			}
			values = append(values, nalog)
		}
		qb.AddIn("rnal.nalog", values)
	}
	if params.CbxDatumNaloga {
		qb.AddCondition("rnal.danal", params.OdDanal, ">=")
		qb.AddCondition("rnal.danal", params.DoDanal, "<=")
	}
	if params.CbxVrstaDokumenta {
		values := make([]any, 0)
		for _, part := range strings.Split(params.SifVrsteDokumenta, ",") {
			vrd, parseErr := strconv.Atoi(strings.TrimSpace(part))
			if parseErr != nil {
				return fmt.Errorf("invalid document type list: %w", parseErr)
			}
			values = append(values, vrd)
		}
		qb.AddIn("rdok.vrd", values)
	}
	if params.CbxBrojDokumenta {
		values := make([]any, 0)
		for _, part := range strings.Split(params.BrojDokumenta, ",") {
			dokum, parseErr := strconv.Atoi(strings.TrimSpace(part))
			if parseErr != nil {
				return fmt.Errorf("invalid document number list: %w", parseErr)
			}
			values = append(values, dokum)
		}
		qb.AddIn("rpro.dokum", values)
	}
	if params.CbxDatumDokumenta {
		qb.AddCondition("rdok.dadok", params.OdDatumaDok, ">=")
		qb.AddCondition("rdok.dadok", params.DoDatumaDok, "<=")
	}
	if params.CbxIznos {
		qb.AddCondition("ABS(rpro.iznos)", params.OdIznosa, ">=")
		qb.AddCondition("ABS(rpro.iznos)", params.DoIznosa, "<=")
	}
	orderPrefix := "rpro.konto, rpro.sifra"
	if printType == common.TipStampePrint {
		orderPrefix = "rpro.sifra, rpro.konto"
	}
	if params.CbxRPROID {
		qb.AddOrderBy(orderPrefix + ", rpro.rproid")
	} else if params.Sortiranje == "dokument" {
		qb.AddOrderBy(orderPrefix + ", rpro.dadok, rpro.rproid")
	} else {
		qb.AddOrderBy(orderPrefix + ", rnal.danal, rpro.rproid")
	}

	if !getTotalRecords && printType == common.TipStampePreview {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoKarticaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	lastSifra := ""
	stanje, saldo := 0.0, 0.0
	printGroups := printType == common.TipStampePrint
	monthlyGroups := printGroups && params.StampajPoMesecima
	currentMonth := 0
	articleUlaz, articleIzlaz, articleDug, articlePot := 0.0, 0.0, 0.0, 0.0
	monthUlaz, monthIzlaz, monthDug, monthPot := 0.0, 0.0, 0.0, 0.0
	emitTotal := func(label string, ulaz, izlaz, dug, pot, saldoValue float64) {
		fieldCount := 28
		ulazIndex, izlazIndex, stanjeIndex := 9, 10, 11
		dugIndex, potIndex, saldoIndex := 13, 14, 15
		if printGroups {
			fieldCount = 13
			ulazIndex, izlazIndex, stanjeIndex = 6, 7, 8
			dugIndex, potIndex, saldoIndex = 10, 11, 12
		}
		fields := make([]string, fieldCount)
		fields[0] = label
		fields[ulazIndex] = common.FormatNumberWithSystemLocale(ulaz, 2)
		fields[izlazIndex] = common.FormatNumberWithSystemLocale(izlaz, 2)
		fields[stanjeIndex] = common.FormatNumberWithSystemLocale(ulaz-izlaz, 2)
		fields[dugIndex] = common.FormatNumberWithSystemLocale(dug, 2)
		fields[potIndex] = common.FormatNumberWithSystemLocale(pot, 2)
		fields[saldoIndex] = common.FormatNumberWithSystemLocale(saldoValue, 2)
		tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: "group-total", Fields: fields})
	}
	tbl.Rows = make([]domain.TableRow, 0, len(*entities))
	for _, entity := range *entities {
		if entity.Sifra != lastSifra {
			if printGroups && lastSifra != "" {
				if monthlyGroups {
					emitTotal(fmt.Sprintf("Promet za mesec %02d", currentMonth), monthUlaz, monthIzlaz, monthDug, monthPot, monthDug-monthPot)
				}
				emitTotal(fmt.Sprintf("Promet artikla: %s - %s", lastSifra, entity.NazivArtikla), articleUlaz, articleIzlaz, articleDug, articlePot, articleDug-articlePot)
			}
			lastSifra = entity.Sifra
			stanje, saldo = 0.0, 0.0
			articleUlaz, articleIzlaz, articleDug, articlePot = 0, 0, 0, 0
			monthUlaz, monthIzlaz, monthDug, monthPot = 0, 0, 0, 0
			currentMonth = 0
			if printGroups {
				tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: "group-header", Fields: []string{fmt.Sprintf("Artikal: %s - %s", entity.Sifra, entity.NazivArtikla)}})
			}
		}
		if monthlyGroups && currentMonth != entity.Mesec {
			if currentMonth != 0 {
				emitTotal(fmt.Sprintf("Promet za mesec %02d", currentMonth), monthUlaz, monthIzlaz, monthDug, monthPot, monthDug-monthPot)
			}
			currentMonth = entity.Mesec
			monthUlaz, monthIzlaz, monthDug, monthPot = 0, 0, 0, 0
			tbl.Rows = append(tbl.Rows, domain.TableRow{ClassRow: "subgroup-header", Fields: []string{fmt.Sprintf("Mesec: %02d", entity.Mesec)}})
		}
		stanje += entity.Ulaz - entity.Izlaz
		saldo += entity.Duguje - entity.Potrazuje
		articleUlaz += entity.Ulaz
		articleIzlaz += entity.Izlaz
		articleDug += entity.Duguje
		articlePot += entity.Potrazuje
		monthUlaz += entity.Ulaz
		monthIzlaz += entity.Izlaz
		monthDug += entity.Duguje
		monthPot += entity.Potrazuje
		entity.Stanje = stanje
		entity.Saldo = saldo

		fields := []string{
			fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
			entity.Danal.Format(common.DateLayout),
			entity.Vrd,
			fmt.Sprintf("%d", entity.Dokum),
			entity.Dadok.Format(common.DateLayout),
			entity.Konto,
			entity.Sifra,
			common.FormatNumberWithSystemLocale(entity.Cena, 2),
			common.FormatNumberWithSystemLocale(entity.Ulaz, 2),
			common.FormatNumberWithSystemLocale(entity.Izlaz, 2),
			common.FormatNumberWithSystemLocale(entity.Stanje, 2),
			common.FormatNumberWithSystemLocale(entity.Iznos, 2),
			common.FormatNumberWithSystemLocale(entity.Duguje, 2),
			common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
			common.FormatNumberWithSystemLocale(entity.Saldo, 2),
			entity.Fkto,
			entity.Fana,
			entity.Fkplnaz,
			entity.Valuta,
			common.FormatNumberWithSystemLocale(entity.Kurs, 4),
			common.FormatNumberWithSystemLocale(entity.CenaVal, 2),
			fmt.Sprintf("%d", entity.Mesec),
			entity.Dokiz,
			entity.Dadokiz.Format(common.DateLayout),
			entity.Otk,
			entity.Serija,
			fmt.Sprintf("%d", entity.Roktr),
		}
		if printType == common.TipStampePrint {
			fields = []string{
				fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				entity.Vrd,
				fmt.Sprintf("%d", entity.Dokum),
				entity.Dadok.Format(common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Cena, 2),
				common.FormatNumberWithSystemLocale(entity.Ulaz, 2),
				common.FormatNumberWithSystemLocale(entity.Izlaz, 2),
				common.FormatNumberWithSystemLocale(entity.Stanje, 2),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Saldo, 2),
			}
		}
		if printType == common.TipStampePrint {
			fields = []string{
				fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				entity.Vrd,
				fmt.Sprintf("%d", entity.Dokum),
				entity.Dadok.Format(common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Cena, 2),
				common.FormatNumberWithSystemLocale(entity.Ulaz, 2),
				common.FormatNumberWithSystemLocale(entity.Izlaz, 2),
				common.FormatNumberWithSystemLocale(entity.Stanje, 2),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(entity.Saldo, 2),
			}
		}

		tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false})
	}
	if printGroups && lastSifra != "" {
		if monthlyGroups {
			emitTotal(fmt.Sprintf("Promet za mesec %02d", currentMonth), monthUlaz, monthIzlaz, monthDug, monthPot, monthDug-monthPot)
		}
		emitTotal(fmt.Sprintf("Promet artikla: %s", lastSifra), articleUlaz, articleIzlaz, articleDug, articlePot, articleDug-articlePot)
	}
	return nil
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoKarticaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New("user session not found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()

	tbl.Headers = s.subsintetickaKarticaKontaTableFields
	qb := common.NewQueryBuilder(`
		SELECT
			rnal.tipdok,
			rnal.nalog,
			rnal.danal,
			rnal.opis,
			sum(case
				when upper(dokvrsta.kodknj) = 'D'
				then case when dokvrsta.predznak = '-'
							then 0 - rpro.iznos
							else rpro.iznos
					end
				else 0
				end) as duguje,

			sum(case
				when upper(dokvrsta.kodknj) = 'P'
				then case when dokvrsta.predznak = '-'
							then 0 - rpro.iznos
							else rpro.iznos
					end
				else 0
				end) as potrazuje
		FROM rnal`, true)
	qb.AddJoin(" inner join rdok on rdok.rnalid = rnal.rnalid")
	qb.AddJoin(" inner join rpro on rdok.rdokid = rpro.rdokid")
	qb.AddJoin(" inner join dokvrsta on rdok.god = dokvrsta.god and rdok.kar = dokvrsta.kar and rdok.vrd = dokvrsta.vrd")
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	qb.AddEqual("rpro.konto", params.Konto)
	qb.AddEqual("rnal.magaciniid", params.Magacin)
	if params.CbxBrojNaloga {
		values := make([]any, 0)
		for _, part := range strings.Split(params.Nalozi, ",") {
			nalog, parseErr := strconv.Atoi(strings.TrimSpace(part))
			if parseErr != nil {
				return fmt.Errorf("invalid nalog list: %w", parseErr)
			}
			values = append(values, nalog)
		}
		qb.AddIn("rnal.nalog", values)
	}
	if params.CbxDatumNaloga {
		qb.AddCondition("rnal.danal", params.OdDanal, ">=")
		qb.AddCondition("rnal.danal", params.DoDanal, "<=")
	}
	if params.CbxIznos {
		start := qb.GetArgsCount() + 1
		qb.AddCustomCondition(fmt.Sprintf("ABS(rpro.iznos) BETWEEN $%d AND $%d", start, start+1))
		qb.AddArgs(params.OdIznosa, params.DoIznosa)
	}
	qb.AddGroupBy("rnal.tipdok, rnal.nalog, rnal.danal, rnal.opis, rpro.konto")
	qb.AddOrderBy("rpro.konto, rnal.danal, rnal.nalog")
	if !getTotalRecords && printType == common.TipStampePreview {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnoKarticaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("load subsintetic card records: %w", err)
	}
	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = make([]domain.TableRow, 0, len(*entities))
	for _, entity := range *entities {
		fields := []string{
			fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
			entity.Danal.Format(common.DateLayout),
			entity.Opis,
			common.FormatNumberWithSystemLocale(entity.Duguje, 2),
			common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
			common.FormatNumberWithSystemLocale(entity.Duguje-entity.Potrazuje, 2),
		}
		tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
	return nil
}

func (s *RobnoKarticaResource) ValidacijaKarticaArtikla(params domain.RobnoKarticaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	if params.CbxDatumNaloga {
		if params.OdDanal == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatumanaloga", ErrorMessage: "obavezan podatak"})
		}
		if params.DoDanal == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumanaloga", ErrorMessage: "obavezan podatak"})
		}
		if params.OdDanal > params.DoDanal {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumanaloga", ErrorMessage: "Neispravan opseg datauma naloga"})
		}
	}
	if params.CbxDatumObrade {
		if params.OdDatumObrade == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatumaobrade", ErrorMessage: "obavezan podatak"})
		}
		if params.DoDatumObrade == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumaobrade", ErrorMessage: "obavezan podatak"})
		}
		if params.OdDatumObrade > params.DoDatumObrade {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumaobrade", ErrorMessage: "Neispravan opseg datauma obrade"})
		}
	}
	if params.CbxDatumDokumenta {
		if params.OdDatumaDok == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatumadok", ErrorMessage: "obavezan podatak"})
		}
		if params.DoDatumaDok == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumadok", ErrorMessage: "obavezan podatak"})
		}
		if params.OdDatumaDok > params.DoDatumaDok {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumadok", ErrorMessage: "Neispravan opseg datuma dokumenta"})
		}
	}
	if params.CbxIznos {
		if params.OdIznosa > params.DoIznosa {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "doiznosa", ErrorMessage: "Neispravan opseg iznosa"})
		}
	}
	if params.OdSifre == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "odsifreartikla", ErrorMessage: "obavezan podatak"})
	}
	if params.DoSifre == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "dosifreartikla", ErrorMessage: "obavezan podatak"})
	}
	return fieldErrors
}
func (s *RobnoKarticaResource) ValidacijaSubsintetickogKonta(params domain.RobnoKarticaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	if params.Konto == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "obavezan podatak"})
	}
	if params.Magacin <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "magacin", ErrorMessage: "obavezan podatak"})
	}
	if params.CbxDatumNaloga {
		if params.OdDanal == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatumanaloga", ErrorMessage: "obavezan podatak"})
		}
		if params.DoDanal == "" {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumanaloga", ErrorMessage: "obavezan podatak"})
		}
	}
	if params.OdDanal != "" && params.DoDanal != "" {
		odDanal, errOd := time.Parse("2006-01-02", params.OdDanal)
		doDanal, errDo := time.Parse("2006-01-02", params.DoDanal)
		if errOd != nil || errDo != nil {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatumanaloga", ErrorMessage: "Neispravan format datuma"})
		} else if odDanal.After(doDanal) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "oddatumanaloga", ErrorMessage: "Neispravan opseg datuma naloga"})
		}
		if odDanal.After(doDanal) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "dodatumanaloga", ErrorMessage: "Neispravan opseg datuma naloga"})
		}
	}
	if params.CbxBrojNaloga && params.Nalozi == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "brojnaloga", ErrorMessage: "obavezan podatak"})
	}
	if params.CbxIznos {
		if params.OdIznosa == 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "odiznosa", ErrorMessage: "obavezan podatak"})
		}
		if params.DoIznosa == 0 {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "doiznosa", ErrorMessage: "obavezan podatak"})
		}
		if params.OdIznosa > params.DoIznosa {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: "odiznosa", ErrorMessage: "Neispravan opseg iznosa"})
		}
	}
	return fieldErrors
}

func (s *RobnoKarticaResource) GetKarticaArtiklaTableFields() []domain.Fields {
	return s.karticaArtiklaTableFields
}
func (s *RobnoKarticaResource) GetKarticaArtiklaStampaTableFields() []domain.Fields {
	return s.karticaArtiklaStampaTableFields
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKontaTableFields() []domain.Fields {
	return s.subsintetickaKarticaKontaTableFields
}

func (s *RobnoKarticaResource) GetKarticaSubsintetickogKontaStampaTableFields() []domain.Fields {
	return s.subsintetickaKarticaStampaTableFields
}

// GetMagacinComboValues returns the magacini of the current period (CommonService).
func (s *RobnoKarticaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinComboValues(ctx)
}
func (s *RobnoKarticaResource) setTableFields() {
	// Initialize the table fields for the RobnoKarticaResource
	s.karticaArtiklaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "5", Field: "nalog", SkipInSearch: true, TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "5", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "5", Field: "vrd", SkipInSearch: true, TextAlign: "left"},
		{Name: "brdo", Label: "Broj Dokumenta", Width: "5", Field: "brdo", SkipInSearch: true, TextAlign: "left"},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "5", Field: "dadok", SkipInSearch: true, TextAlign: "center"},
		{Name: "konto", Label: "Konto", Width: "5", Field: "konto", SkipInSearch: true, TextAlign: "left"},
		{Name: "sifra", Label: "Sifra", Width: "5", Field: "sifra", SkipInSearch: true, TextAlign: "left"},
		{Name: "cena", Label: "Kolicina", Width: "5", Field: "kolicina", SkipInSearch: true, TextAlign: "right"},
		{Name: "ulaz", Label: "Ulaz", Width: "5", Field: "ulaz", SkipInSearch: true, TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "5", Field: "izlaz", SkipInSearch: true, TextAlign: "right"},
		{Name: "stanje", Label: "Stanje", Width: "5", Field: "stanje", SkipInSearch: true, TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "5", Field: "iznos", SkipInSearch: true, TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "5", Field: "duguje", SkipInSearch: true, TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "5", Field: "potrazuje", SkipInSearch: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "5", Field: "saldo", SkipInSearch: true, TextAlign: "right"},
		{Name: "fkto", Label: "Fkto", Width: "5", Field: "fkto", SkipInSearch: true, TextAlign: "left"},
		{Name: "fana", Label: "Napomena", Width: "5", Field: "fana", SkipInSearch: true, TextAlign: "left"},
		{Name: "fkplnaz", Label: "Naziv konta", Width: "5", Field: "fkplnaz", SkipInSearch: true, TextAlign: "left"},
		{Name: "valuta", Label: "Valuta", Width: "5", Field: "valuta", SkipInSearch: true, TextAlign: "left"},
		{Name: "kurs", Label: "Kurs", Width: "5", Field: "kurs", SkipInSearch: true, TextAlign: "right"},
		{Name: "cenaval", Label: "Cena u valuti", Width: "5", Field: "cenaval", SkipInSearch: true, TextAlign: "right"},
		{Name: "mesec", Label: "Mesec", Width: "5", Field: "mesec", SkipInSearch: true, TextAlign: "right"},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "5", Field: "dokiz", SkipInSearch: true, TextAlign: "left"},
		{Name: "dadokiz", Label: "Datum izvornog dokumenta", Width: "5", Field: "dadokiz", SkipInSearch: true, TextAlign: "left"},
		{Name: "otk", Label: "OTK", Width: "5", Field: "otk", SkipInSearch: true, TextAlign: "left"},
		{Name: "serija", Label: "Serija", Width: "5", Field: "serija", SkipInSearch: true, TextAlign: "left"},
		{Name: "rok", Label: "Rok", Width: "5", Field: "rok", SkipInSearch: true, TextAlign: "left"},
	}
	s.karticaArtiklaStampaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "5", Field: "nalog", SkipInSearch: true, TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "5", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta Dokumenta", Width: "5", Field: "vrd", SkipInSearch: true, TextAlign: "left"},
		{Name: "brdo", Label: "Broj Dokumenta", Width: "5", Field: "brdo", SkipInSearch: true, TextAlign: "left"},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "5", Field: "dadok", SkipInSearch: true, TextAlign: "center"},
		{Name: "cena", Label: "Kolicina", Width: "5", Field: "kolicina", SkipInSearch: true, TextAlign: "right"},
		{Name: "ulaz", Label: "Ulaz", Width: "5", Field: "ulaz", SkipInSearch: true, TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "5", Field: "izlaz", SkipInSearch: true, TextAlign: "right"},
		{Name: "stanje", Label: "Stanje", Width: "5", Field: "stanje", SkipInSearch: true, TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "5", Field: "iznos", SkipInSearch: true, TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "5", Field: "duguje", SkipInSearch: true, TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "5", Field: "potrazuje", SkipInSearch: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "5", Field: "saldo", SkipInSearch: true, TextAlign: "right"},
	}
	s.subsintetickaKarticaKontaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Nalog", Width: "5", Field: "nalog", SkipInSearch: false, TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "5", Field: "danal", SkipInSearch: false, TextAlign: "center"},
		{Name: "opis", Label: "Opis", Width: "5", Field: "opis", SkipInSearch: false, TextAlign: "left"},
		{Name: "duguje", Label: "Duguje", Width: "5", Field: "duguje", SkipInSearch: false, TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "5", Field: "potrazuje", SkipInSearch: false, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "5", Field: "saldo", SkipInSearch: false, TextAlign: "right"},
	}
	s.subsintetickaKarticaStampaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj Naloga", Width: "8", Field: "nalog", SkipInSearch: true, TextAlign: "left"},
		{Name: "danal", Label: "Datum naloga", Width: "10", Field: "danal", SkipInSearch: true, TextAlign: "center"},
		{Name: "opis", Label: "Opis", Width: "30", Field: "opis", SkipInSearch: true, TextAlign: "left"},
		{Name: "duguje", Label: "Iznos Duguje", Width: "12", Field: "duguje", SkipInSearch: true, TextAlign: "right"},
		{Name: "potrazuje", Label: "Iznos Potrazuje", Width: "12", Field: "potrazuje", SkipInSearch: true, TextAlign: "right"},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "saldo", SkipInSearch: true, TextAlign: "right"},
		{Name: "znak", Label: "D/P", Width: "3", Field: "znak", SkipInSearch: true, TextAlign: "center"},
	}
}

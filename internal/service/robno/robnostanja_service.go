package robno

import (
	"context"
	"fmt"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
)

type RobnoStanjaService interface {
	GetStanjePojedinacnogArtikla(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams, string) error
	GetUkupnaObrada(context.Context, *domain.RobnoStanjaTotal, domain.RobnoStanjaParams) error
	GetStanjaViseArtikala(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaViseArtikalaSifra(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaViseArtikalaGrupa(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetStanjaSubsintetickogKonta(context.Context, *domain.TableData, domain.RobnoStanjaParams) error
	GetSvodjenjeZaliha(context.Context, *domain.TableData, bool, int, int, domain.RobnoStanjaParams) error
	GetFvrData(context.Context) (domain.Fvr, error)
	GetMagacinComboValues(context.Context) ([]domain.ComboItem, error)
	GetTipDokComboValues(context.Context) ([]domain.ComboItem, error)
	GetOjComboValues(context.Context) ([]domain.ComboItem, error)
	GetMestoTroskaComboValues(context.Context, int) ([]domain.ComboItem, error)
	SetDefaultTableData(tbl *domain.TableData)
	GetPojedinacnogArtiklaTableFields() []domain.Fields
	GetViseArtikalaTableFields() []domain.Fields
	GetSubsintetiskogKontaTableFields() []domain.Fields
	GetSvodjenjeZalihaTableFields() []domain.Fields
}

type RobnoStanjaResource struct {
	robnostanjeRepo                repository.BaseRepository[domain.RobnoStanjeDto]
	rproRepo                       repository.BaseRepository[domain.Rpro]
	drstaRepo                      repository.BaseRepository[domain.Drsta]
	magRepo                        repository.BaseRepository[domain.Magacini]
	tipdokRepo                     repository.BaseRepository[domain.Tipdok]
	ojRepo                         repository.BaseRepository[domain.Orgjed]
	mestoTroskaRepo                repository.BaseRepository[domain.Mestotr]
	fvrRepo                        repository.BaseRepository[domain.Fvr]
	pojedinacniArtikalTableFields  []domain.Fields
	viseArtikalaTableFields        []domain.Fields
	subsintetickogKontaTableFields []domain.Fields
	svodjenjeZalihaTableFields     []domain.Fields
}

func NewRobnoStanjaService(robnostanjeRepo repository.BaseRepository[domain.RobnoStanjeDto], rproRepo repository.BaseRepository[domain.Rpro], drstaRepo repository.BaseRepository[domain.Drsta], magRepo repository.BaseRepository[domain.Magacini], tipdokRepo repository.BaseRepository[domain.Tipdok], ojRepo repository.BaseRepository[domain.Orgjed], mestoTroskaRepo repository.BaseRepository[domain.Mestotr], fvrRepo repository.BaseRepository[domain.Fvr]) *RobnoStanjaResource {
	rs := &RobnoStanjaResource{
		robnostanjeRepo: robnostanjeRepo,
		rproRepo:        rproRepo,
		drstaRepo:       drstaRepo,
		magRepo:         magRepo,
		tipdokRepo:      tipdokRepo,
		ojRepo:          ojRepo,
		mestoTroskaRepo: mestoTroskaRepo,
		fvrRepo:         fvrRepo,
	}
	rs.setTableFileds()
	return rs
}

func (s *RobnoStanjaResource) GetStanjePojedinacnogArtikla(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, params domain.RobnoStanjaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pojedinacniArtikalTableFields

	hasGod, hasKar := s.drstaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select rsif.sifra, rsif.jm, drsta.cena,
		mdug_1, mdug_2, mdug_3, mdug_4, mdug_5, mdug_6, mdug_7, mdug_8, mdug_9, mdug_10, mdug_11, mdug_12,
		mpot_1, mpot_2, mpot_3, mpot_4, mpot_5, mpot_6, mpot_7, mpot_8, mpot_9, mpot_10, mpot_11, mpot_12, 
		mul_1, mul_2, mul_3, mul_4, mul_5, mul_6, mul_7, mul_8, mul_9, mul_10, mul_11, mul_12, 
		miz_1, miz_2, miz_3, miz_4, miz_5, miz_6, miz_7, miz_8, miz_9, miz_10, miz_11, miz_12 from drsta`, true)
	qb.AddJoin(" inner join rsif on rsif.rsifid = drsta.rsifid")
	if hasGod {
		qb.AddEqual("drsta.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("drsta.kar", userSession.SelectedKar)
	}
	if params.Magacin != 0 {
		qb.AddEqual("drsta.magaciniid", params.Magacin)
	}
	if params.Konto != "" {
		qb.AddEqual("drsta.konto", params.Konto)
	}
	if params.SifraArtikla != "" {
		qb.AddEqual("rsif.sifra", params.SifraArtikla)
	}
	qb.AddOrderBy("rsif.sifra")
	if !getTotalRecords && printType == common.TipStampePreview {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	if !getTotalRecords && printType == common.TipStampePreview {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	months := common.GetMontshName()
	rowsPerArticle := len(months)
	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities)*rowsPerArticle, pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		ulaz := []float64{entity.MUl1, entity.MUl2, entity.MUl3, entity.MUl4, entity.MUl5, entity.MUl6, entity.MUl7, entity.MUl8, entity.MUl9, entity.MUl10, entity.MUl11, entity.MUl12}
		izlaz := []float64{entity.MIz1, entity.MIz2, entity.MIz3, entity.MIz4, entity.MIz5, entity.MIz6, entity.MIz7, entity.MIz8, entity.MIz9, entity.MIz10, entity.MIz11, entity.MIz12}
		duguje := []float64{entity.MDug1, entity.MDug2, entity.MDug3, entity.MDug4, entity.MDug5, entity.MDug6, entity.MDug7, entity.MDug8, entity.MDug9, entity.MDug10, entity.MDug11, entity.MDug12}
		potrazuje := []float64{entity.MPot1, entity.MPot2, entity.MPot3, entity.MPot4, entity.MPot5, entity.MPot6, entity.MPot7, entity.MPot8, entity.MPot9, entity.MPot10, entity.MPot11, entity.MPot12}
		for i, month := range months {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					month,
					common.FormatNumberWithSystemLocale(ulaz[i], 3),
					common.FormatNumberWithSystemLocale(izlaz[i], 3),
					common.FormatNumberWithSystemLocale(duguje[i], 2),
					common.FormatNumberWithSystemLocale(potrazuje[i], 2),
				},
				HasUpdate: false,
				HasDelete: false,
			})
		}
	}
	return nil
}
func (s *RobnoStanjaResource) GetUkupnaObrada(ctx context.Context, total *domain.RobnoStanjaTotal, params domain.RobnoStanjaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}

	hasGod, hasKar := s.rproRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`select drsta.ulaz, drsta.izlaz, drsta.dug as duguje, drsta.pot as potrazuje,
		rsif.sifra, coalesce(rsif.naziv, '') as naziv_artikla, coalesce(rsif.jm, '') as jm,
		coalesce(drsta.cena, 0) as cena, coalesce(fkpl.naziv, '') as kontonaziv, coalesce(magacini.opis, '') as magacinnaziv
		from drsta`, true)
	qb.AddJoin("inner join rsif on rsif.rsifid = drsta.rsifid")
	qb.AddJoin("left join magacini on magacini.magaciniid = drsta.magaciniid")
	qb.AddJoin("left join fkpl on fkpl.god = drsta.god and fkpl.kar = drsta.kar and fkpl.vkonta = 2 and fkpl.konto = drsta.konto")
	if hasGod {
		qb.AddEqual("drsta.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("drsta.kar", userSession.SelectedKar)
	}
	if params.Magacin != 0 {
		qb.AddEqual("drsta.magaciniid", params.Magacin)
	}
	if params.Konto != "" {
		qb.AddEqual("drsta.konto", params.Konto)
	}
	if params.SifraArtikla != "" {
		qb.AddEqual("rsif.sifra", params.SifraArtikla)
	}
	qb.AddOrderBy("rsif.sifra")
	sqlQuery, args := qb.Build()
	// Execute the query and populate the total object
	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if entities == nil || len(*entities) == 0 {
		return nil
	}
	entity := (*entities)[0]
	total.Ulaz = entity.Ulaz
	total.Izlaz = entity.Izlaz
	total.Stanje = entity.Ulaz - entity.Izlaz
	total.Duguje = entity.Duguje
	total.Potrazuje = entity.Potrazuje
	total.Saldo = entity.Duguje - entity.Potrazuje
	total.Sifra = entity.Sifra
	total.NazivArtikla = entity.NazivArtikla
	total.Jm = entity.Jm
	total.Cena = entity.Cena
	total.KontoNaziv = entity.KontoNaziv
	total.MagacinNaziv = entity.MagacinNaziv
	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikala(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikala

	return nil
}
func (s *RobnoStanjaResource) GetStanjaSubsintetickogKonta(ctx context.Context, tbl *domain.TableData, params domain.RobnoStanjaParams) error {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	// Create first query (opening balance - tipdok = '00')
	qb1 := common.NewQueryBuilder(`SELECT 0 as mesec,
   			CASE WHEN upper(dokvrsta.kodknj) = 'D' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END AS duguje,
			CASE WHEN upper(dokvrsta.kodknj) = 'P' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END AS potrazuje,
			CASE WHEN dokvrsta.grpdok = 'NIV' THEN 0 ELSE rpro.iznos END AS iznos,
    FROM rpro `, true)
	qb1.AddJoin(" inner join rnal on rnal.rnaild = rpro.rnalid")
	qb1.AddJoin(" inner join dokvrsta on dokvrsta.dokvrstaid = rnal.dokvrstaid")
	// Add conditions to first query
	hasGod, hasKar := s.robnostanjeRepo.GetHasGodHasKar()
	if hasGod {
		qb1.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb1.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	qb1.AddEqual("rnal.tipdok", "00")
	qb1.AddLike("rpro.konto", params.Konto)
	qb1.AddEqual("rnal.magaciniid", params.MagaciniID)
	qb1.AddGroupBy("EXTRACT(MONTH FROM rnal.danal)")
	qb1.AddOrderBy("mesec ASC")
	// Create second query (monthly data - tipdok != '00')
	qb2 := common.NewQueryBuilder(`SELECT EXTRACT(MONTH FROM rnal.danal) as mesec,
			CASE WHEN upper(dokvrsta.kodknj) = 'D' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END AS duguje,
			CASE WHEN upper(dokvrsta.kodknj) = 'P' THEN
				CASE WHEN dokvrsta.grpdok = 'NIV' THEN rpro.kolic * (rpro.cena - rpro.fcena) ELSE rpro.iznos END
			ELSE 0 END AS potrazuje,
			CASE WHEN dokvrsta.grpdok = 'NIV' THEN 0 ELSE rpro.iznos END AS iznos,
    FROM rpro `, true)
	qb2.AddJoin(" inner join rnal on rnal.rnaild = rpro.rnalid")
	qb2.AddJoin(" inner join dokvrsta on dokvrsta.dokvrstaid = rnal.dokvrstaid")
	// Add same base conditions
	if hasGod {
		qb2.AddEqual("rpro.god", userSession.SelectedGod)
	}
	if hasKar {
		qb2.AddEqual("rpro.kar", userSession.SelectedKar)
	}
	qb2.AddEqual("rnal.magaciniid", params.MagaciniID)
	qb2.AddEqual("rpro.konto", params.Konto)
	qb2.AddCondition("rnal.tipdok", "00", "!=")
	qb2.AddGroupBy("EXTRACT(MONTH FROM rnal.danal)")
	// Create UNION
	uqb := common.NewUnionQueryBuilder("UNION ALL")
	uqb.AddQuery(qb1)
	uqb.AddQuery(qb2)
	uqb.AddOrderBy("mesec")

	sqlQuery, args := uqb.Build()

	entities, err := s.robnostanjeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Create template data with opening balance and all 12 months initialized to 0
	templateData := make([]domain.SaldaDto, 13) // 0 (opening balance) + 12 months

	// Merge actual data into template - update the corresponding month rows with real values
	kumulSaldo := 0.0
	if entities != nil {
		for _, entity := range *entities {
			if entity.Mesec >= 0 && entity.Mesec <= 12 {
				// Monthly data
				saldoDto := domain.SaldaDto{
					Mesec:     entity.Mesec,
					Duguje:    entity.Duguje,
					Potrazuje: entity.Potrazuje,
					Saldo:     entity.Duguje - entity.Potrazuje,
				}
				kumulSaldo = kumulSaldo + (entity.Duguje - entity.Potrazuje)
				templateData[entity.Mesec] = saldoDto
				templateData[entity.Mesec].SaldoKumul = kumulSaldo
			}
		}
	}
	monthNames := common.GetMontshName()
	for i, salda := range templateData {
		monthName := monthNames[i]
		if monthName == "" {
			monthName = fmt.Sprintf("Mesec %d", salda.Mesec)
		}

		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				monthName,
				common.FormatNumberWithSystemLocale(salda.Duguje, 2),
				common.FormatNumberWithSystemLocale(salda.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(salda.Saldo, 2),
				common.FormatNumberWithSystemLocale(salda.SaldoKumul, 2),
			},
		})
	}
	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikalaSifra(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikalaSifra

	return nil
}
func (s *RobnoStanjaResource) GetStanjaViseArtikalaGrupa(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetStanjaViseArtikalaGrupa

	return nil
}

func (s *RobnoStanjaResource) GetSvodjenjeZaliha(ctx context.Context, tbl *domain.TableData, total bool, size, page int, params domain.RobnoStanjaParams) error {
	// TODO: Implement GetSvodjenjeZaliha

	return nil
}
func (s *RobnoStanjaResource) GetMagacinComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.magRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select magaciniid, mag, opis from magacini", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("mag")
	sqlQuery, args := qb.Build()
	entites, err := s.magRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := make([]domain.ComboItem, len(*entites))
	for i, entity := range *entites {
		comboItems[i] = domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.MagaciniID),
			Value: fmt.Sprintf("%d - %s", entity.Mag, entity.Opis),
		}
	}
	return comboItems, nil
}
func (s *RobnoStanjaResource) GetTipDokComboValues(ctx context.Context) ([]domain.ComboItem, error) {

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select idtipdok, tipdok, opis from tipdok", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("tipdok")
	sqlQuery, args := qb.Build()
	entites, err := s.tipdokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := make([]domain.ComboItem, len(*entites))
	for i, entity := range *entites {
		comboItems[i] = domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.IDTipDok),
			Value: fmt.Sprintf("%s - %s", entity.TipDok, entity.Opis),
		}
	}
	return comboItems, nil
}

func (s *RobnoStanjaResource) GetOjComboValues(ctx context.Context) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.ojRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select idorgjed, ojozn, naziv from orgjed", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddOrderBy("ojozn")
	sqlQuery, args := qb.Build()
	entites, err := s.ojRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found

	for _, entity := range *entites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.IDOrgjed),
			Value: fmt.Sprintf("%s - %s", entity.OjOzn, entity.Naziv),
		})
	}
	return comboItems, nil
}

func (s *RobnoStanjaResource) GetMestoTroskaComboValues(ctx context.Context, idOrgjed int) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	hasGod, haskar := s.mestoTroskaRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(" select mestotrid, mtroska, opis from mestotr", true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("idorgjed", idOrgjed)
	qb.AddOrderBy("mtroska")
	sqlQuery, args := qb.Build()
	entites, err := s.mestoTroskaRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	comboItems := []domain.ComboItem{}
	comboItems = append(comboItems, domain.ComboItem{Key: "-", Value: "-"}) // Default option when no records are found

	for _, entity := range *entites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", entity.MestoTrID),
			Value: fmt.Sprintf("%s - %s", entity.Mtroska, entity.Opis),
		})
	}
	return comboItems, nil
}

func (s *RobnoStanjaResource) GetPojedinacnogArtiklaTableFields() []domain.Fields {
	return s.pojedinacniArtikalTableFields
}
func (s *RobnoStanjaResource) GetViseArtikalaTableFields() []domain.Fields {
	return s.viseArtikalaTableFields
}
func (s *RobnoStanjaResource) GetSubsintetiskogKontaTableFields() []domain.Fields {
	return s.subsintetickogKontaTableFields
}
func (s *RobnoStanjaResource) GetSvodjenjeZalihaTableFields() []domain.Fields {
	return s.svodjenjeZalihaTableFields
}

// GetFvrData retrieves company (fvr) data for the current session.
func (s *RobnoStanjaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, &s.fvrRepo)
}

func (s *RobnoStanjaResource) SetDefaultTableData(tbl *domain.TableData) {
	tbl.TableID = "saldapojedinacnihkonta-table"
	tbl.SearchEnabled = false
	tbl.ShowPagination = false

	// Opening balance row
	fields := []string{"Početno stanje", "", "", "", ""}
	tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
	tbl.Rows = append(tbl.Rows, tblRow)

	months := common.GetMontshName()
	// Add month rows
	for i := 0; i <= 11; i++ {
		// Create a NEW slice for each row instead of reusing the same slice
		monthFields := []string{months[i], "", "", "", ""}
		tblRow := domain.TableRow{Fields: monthFields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
}

func (s *RobnoStanjaResource) setTableFileds() {
	s.pojedinacniArtikalTableFields = []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "12", TextAlign: "left"},
		{Name: "ulaz", Label: "Ulaz", Width: "14", TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "14", TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "14", TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "14", TextAlign: "right"},
	}
	s.viseArtikalaTableFields = []domain.Fields{}
	s.subsintetickogKontaTableFields = []domain.Fields{}
	s.svodjenjeZalihaTableFields = []domain.Fields{}
}

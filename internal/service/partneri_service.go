package service

import (
	"bytes"
	"context"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/validation"
	"html"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

type PartneriService interface {
	Service[domain.Partneri]
	GetCompanyByPIB(pib string) (*domain.CompanyInfo, error)
	GetCompanyAccountsByPIB(tbl *domain.TableData, pib string) error
	GetLastPartneriID(ctx context.Context) (int64, error)
	ValidacijaPartneri(ctx context.Context, entity *domain.Partneri) ([]domain.FieldError, error)
	GetTekuciRacuni(ctx context.Context, id int64, tbl *domain.TableData) error
	GetTekuciRacuniTableFields() []domain.Fields
	GetPartneriTableFields() []domain.Fields
	PrepareInsertUpdateFields(partner *domain.Partneri) []domain.Fields
	DeleteTekRacuniForPartner(ctx context.Context, partneriID int64) error
	CreateWithTekRacuni(ctx context.Context, partner *domain.Partneri, tekRacuniList []domain.TekRacuni) (int64, error)
	UpdateWithTekRacuni(ctx context.Context, partner *domain.Partneri, id int64, tekRacuniList []domain.TekRacuni) error
}

type PartneriResource struct {
	service                 *BaseService[domain.Partneri]
	validator               *validation.RuleBasedValidator[domain.Partneri]
	partneriRepo            *repository.BaseRepository[domain.Partneri]
	tekracuniRepo           *repository.BaseRepository[domain.TekRacuni]
	partneriTableFields     []domain.Fields
	tekuciRacuniTableFields []domain.Fields
}

func NewPartneriService(service *BaseService[domain.Partneri], validator *validation.RuleBasedValidator[domain.Partneri], partneriRepo *repository.BaseRepository[domain.Partneri], tekracuniRepo *repository.BaseRepository[domain.TekRacuni]) *PartneriResource {
	rs := &PartneriResource{
		service:       service,
		validator:     validator,
		partneriRepo:  partneriRepo,
		tekracuniRepo: tekracuniRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *PartneriResource) GetTekuciRacuniTableFields() []domain.Fields {
	return s.tekuciRacuniTableFields
}
func (s *PartneriResource) GetPartneriTableFields() []domain.Fields {
	return s.partneriTableFields
}

func (s *PartneriResource) GetLastPartneriID(ctx context.Context) (int64, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return 0, fmt.Errorf("user session not found")
	}
	qb := common.NewQueryBuilder(" SELECT MAX(partneri.sifra::numeric) as sifra FROM partneri", true)
	hasGod, hasKar := s.partneriRepo.GetHasGodHasKar()

	if hasGod {
		qb.AddCondition("partneri.god", session.SelectedGod, "=")
	}
	if hasKar {
		qb.AddCondition("partneri.kar", session.SelectedKar, "=")
	}
	sqlQuery, args := qb.Build()
	entities, err := s.partneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return 0, err
	}
	if len(*entities) == 0 {
		return 0, nil
	}
	lastID, err := strconv.ParseInt((*entities)[0].Sifra, 10, 64)
	if err != nil {
		return 0, err
	}
	return lastID + 1, nil
}

// Create implements NalogService.
func (s *PartneriResource) Create(ctx context.Context, partner *domain.Partneri, idField string, tableFields []domain.Fields) ([]domain.FieldError, int64, error) {
	lastInsertedID, err := s.partneriRepo.Create(ctx, partner, idField, tableFields)
	return []domain.FieldError{}, lastInsertedID, err
}

func (s *PartneriResource) Delete(ctx context.Context, idField string, id int64) error {
	return s.service.Delete(ctx, idField, id)
}

// GetAll implements NalogService.
func (s *PartneriResource) GetAll(ctx context.Context, page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Partneri, error) {
	return s.service.GetAll(ctx, page, offset, tableFields, idField, searchParams...)
}

// GetAllCustom implements NalogService.
func (s *PartneriResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Partneri, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements NalogService.
func (s *PartneriResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Partneri, error) {
	return s.service.GetByID(ctx, idField, idValue)
}

func (s *PartneriResource) GetTekuciRacuni(ctx context.Context, id int64, tbl *domain.TableData) error {

	qb := common.NewQueryBuilder("SELECT t.tekrac, b.banka FROM tekracuni as t", true)
	qb.AddJoin(" left join banke as b on b.idbanke = t.idbanke")
	qb.AddCondition("t.idpartneri", id, "=")

	sqlQuery, args := qb.Build()
	entities, err := s.tekracuniRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if len(*entities) == 0 {
		return nil
	}
	for i, entity := range *entities {
		fields := []string{
			fmt.Sprintf("%d", i+1),
			entity.TekRac,
			entity.BankeDes.String,
		}
		tableRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tableRow)
	}
	return nil
}

func (s *PartneriResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.fieldCache == nil {
		s.service.fieldCache = make(map[string]reflect.StructField)
	}
	return s.service.fieldCache
}

// GetTotalRecords implements NalogService.
func (s *PartneriResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchParams ...string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchParams...)
}

// GetTotalRecordsCustom implements NalogService.
func (s *PartneriResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements NalogService.
func (s *PartneriResource) MapEntityToValues(entity *domain.Partneri, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

// Update .
func (s *PartneriResource) Update(ctx context.Context, entity *domain.Partneri, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	return s.service.Update(ctx, entity, idField, idValue, tableFields)
}

// DeleteTekRacuniForPartner deletes all tekracuni for a partner (used when updating)
func (s *PartneriResource) DeleteTekRacuniForPartner(ctx context.Context, partneriID int64) error {
	// Get all tekracuni for this partner
	qb := common.NewQueryBuilder("SELECT * FROM tekracuni", true)
	qb.AddCondition("idpartneri", partneriID, "=")

	sqlQuery, args := qb.Build()
	entities, err := s.tekracuniRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("error fetching tekracuni for deletion: %w", err)
	}

	// Delete each tekracuni
	for _, tr := range *entities {
		err := s.tekracuniRepo.Delete(ctx, "tekracuniid", int64(tr.TekRacuniID))
		if err != nil {
			return fmt.Errorf("error deleting tekracuni: %w", err)
		}
	}

	return nil
}

// CreateWithTekRacuni creates a partneri and its tekracuni in one transaction
func (s *PartneriResource) CreateWithTekRacuni(ctx context.Context, partner *domain.Partneri, tekRacuniList []domain.TekRacuni) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("user session not found")
	}

	// Start a transaction
	tx, err := s.partneriRepo.BeginTx()
	if err != nil {
		return 0, fmt.Errorf("error beginning transaction: %w", err)
	}

	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Create the partner first within the transaction
	newID, err := s.partneriRepo.CreateWithTx(ctx, tx, partner, common.IDpartneri, s.PrepareInsertUpdateFields(partner))
	if err != nil {
		return 0, fmt.Errorf("error creating partner: %w", err)
	}

	// If tekracuni exist, save them within the same transaction
	if len(tekRacuniList) > 0 {
		for _, tr := range tekRacuniList {
			tr.IDPartneri = int(newID)
			tr.God = userSession.SelectedGod
			tr.Kar = userSession.SelectedKar
			// Create the tekracuni record within the transaction
			_, err := s.tekracuniRepo.CreateWithTx(ctx, tx, &tr, common.IDtekracuni, []domain.Fields{
				{Name: "idpartneri", Value: fmt.Sprintf("%d", tr.IDPartneri)},
				{Name: "tekrac", Value: tr.TekRac},
			})
			if err != nil {
				return 0, fmt.Errorf("error saving tekracuni: %w", err)
			}
		}
	}
	// Commit the transaction if everything succeeded
	err = tx.Commit()
	if err != nil {
		return newID, fmt.Errorf("error committing transaction: %w", err)
	}

	return newID, nil
}

// UpdateWithTekRacuni updates a partneri and its tekracuni in one transaction
func (s *PartneriResource) UpdateWithTekRacuni(ctx context.Context, partner *domain.Partneri, id int64, tekRacuniList []domain.TekRacuni) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	// Start a transaction
	tx, err := s.partneriRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update the partner within the transaction
	err = s.partneriRepo.UpdateWithTx(ctx, tx, partner, common.IDpartneri, id, s.PrepareInsertUpdateFields(partner))
	if err != nil {
		return fmt.Errorf("error updating partner: %w", err)
	}

	// Delete each tekracuni within the transaction
	err = s.tekracuniRepo.DeleteWithTx(ctx, tx, common.IDpartneri, int64(id))
	if err != nil {
		return fmt.Errorf("error deleting tekracuni: %w", err)

	}
	// Save new tekracuni within the transaction
	if len(tekRacuniList) > 0 {
		for _, tr := range tekRacuniList {
			tr.IDPartneri = int(id)
			tr.God = userSession.SelectedGod
			tr.Kar = userSession.SelectedKar

			// Create the tekracuni record within the transaction
			_, err := s.tekracuniRepo.CreateWithTx(ctx, tx, &tr, common.IDtekracuni, []domain.Fields{
				{Name: "idpartneri", Value: fmt.Sprintf("%d", tr.IDPartneri)},
				{Name: "tekrac", Value: tr.TekRac},
			})

			if err != nil {
				return fmt.Errorf("error saving tekracuni: %w", err)
			}
		}
	}

	// Commit the transaction if everything succeeded
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *PartneriResource) ValidacijaPartneri(ctx context.Context, entity *domain.Partneri) ([]domain.FieldError, error) {
	var fieldErrors []domain.FieldError

	// Trim spaces from fields
	pib := strings.TrimSpace(entity.PIB)
	bpg := strings.TrimSpace(entity.BPG)
	jmbg := strings.TrimSpace(entity.JMBG)
	index := strings.TrimSpace(entity.Index)
	//konta := strings.TrimSpace(entity.Konta)

	// Check Sifra - obavezna
	if strings.TrimSpace(entity.Sifra) == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "sifra",
			ErrorMessage: "Morate uneti analitiku sifru partnera!!!",
		})
	}
	// Always check for duplicate sifra (validation applies to both create and update)
	exists, err := s.ExistsByField(ctx, "sifra", strings.TrimSpace(entity.Sifra))
	if err != nil {
		return fieldErrors, err
	}
	if exists {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "sifra",
			ErrorMessage: "Partner sa ovom šifrom već postoji!!!",
		})
	}

	// Check Naziv - obavezna
	if strings.TrimSpace(entity.Naziv) == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "naziv",
			ErrorMessage: "Morate uneti naziv partnera!!!",
		})
	}

	// Check Mesto - obavezna
	if strings.TrimSpace(entity.Mesto) == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "mesto",
			ErrorMessage: "Morate uneti naziv mesta!!!",
		})
	}

	// Check COMBO_TIP (TipPDV) - obavezna
	if entity.TipPDV <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "tippdv",
			ErrorMessage: "Morate izabrati u koji sistem PDV-a pripada tip partnera!!!",
		})
	}

	// PIB validation for types 1 and 2 (U sistemu PDV-a, Van sistema PDV-a)
	if (entity.TipPDV == 1 || entity.TipPDV == 2) && pib != "" {
		// Check PIB length
		if len(pib) != 9 {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "pib",
				ErrorMessage: "Vas PIB nema 9 cifara ili ima više od 9 cifara, upišite novi",
			})
		} else {
			// Validate PIB checksum
			if !ValidatePIBChecksum(pib) {
				fieldErrors = append(fieldErrors, domain.FieldError{
					Field:        "pib",
					ErrorMessage: "Uneti PIB nije korektan!!!",
				})
			}
		}
	}

	// Check for duplicate PIB (types 1 and 2)
	if (entity.TipPDV == 1 || entity.TipPDV == 2) && pib != "" {
		// For now, this is a placeholder for database check logic
		exists, err := s.ExistsByField(ctx, "pib", pib)
		if err != nil {
			return fieldErrors, err
		}
		if exists {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "pib",
				ErrorMessage: "Partner sa izabranim PIB-om već postoji!",
			})
		}
	}

	// Check for duplicate BPG (type 3 - Registrovani poljoprivrednik)
	if entity.TipPDV == 3 && bpg != "" {
		exists, err := s.ExistsByField(ctx, "bpg", bpg)
		if err != nil {
			return fieldErrors, err
		}
		if exists {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "bpg",
				ErrorMessage: "Partner sa ovakvim BPG-om već postoji!",
			})
		}
	}

	// Check for duplicate JMBG (type 4 - Fizičko lice)
	if entity.TipPDV == 4 && jmbg != "" {
		// Placeholder for database check
		exists, err := s.ExistsByField(ctx, "jmbg", jmbg)
		if err != nil {
			return fieldErrors, err
		}
		if exists {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "jmbg",
				ErrorMessage: "Partner sa ovakvim JMBG-om već postoji!",
			})
		}
	}

	// Student validation (type 5)
	if entity.TipPDV == 5 {
		ok1 := true
		ok2 := true

		// Check JMBG for students
		if jmbg != "" {
			// Placeholder for database check
			exists, err := s.ExistsByField(ctx, "jmbg", jmbg)
			if err != nil {
				return fieldErrors, err
			}
			ok1 = exists
		}

		// Check Index for students
		if index != "" {
			// Placeholder for database check
			exists, err := s.ExistsByField(ctx, "index", index)
			if err != nil {
				return fieldErrors, err
			}
			ok2 = exists
		}

		if !ok1 || !ok2 {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "jmbg",
				ErrorMessage: "Student sa ovakvim podacima već postoji!!!",
			})
		}
	}

	// Budzetski validation - obavezna JBKJS
	if entity.Budzetski {
		jbkjs := strings.TrimSpace(entity.JBKJS)
		if jbkjs == "" || len(jbkjs) < 5 {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "jbkjs",
				ErrorMessage: "Morate uneti JBKJS od 5 cifara!!!",
			})
		}
	}

	return fieldErrors, nil
}

func (s *PartneriResource) ExistsByField(ctx context.Context, filedName string, filedValue interface{}) (bool, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return false, fmt.Errorf("neautorizovan pristup")
	}
	qb := common.NewQueryBuilder("SELECT sifra FROM partneri", true)
	hasGod, hasKar := s.partneriRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddCondition("partneri.god", session.SelectedGod, "=")
	}
	if hasKar {
		qb.AddCondition("partneri.kar", session.SelectedKar, "=")
	}
	qb.AddCondition(fmt.Sprintf("partneri.%s", filedName), filedValue, "=")

	sqlQuery, args := qb.Build()

	entites, err := s.partneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return false, err
	}
	if len(*entites) > 0 {
		return true, nil
	}
	return false, nil
}

// ValidatePIBChecksum validates PIB checksum using ISO 7064 MOD(11,10)
// PIB format: 9 digits
// Algorithm:
// 1. First digit + 10 = sum
// 2. sum mod 10 = remainder (if 0, then 10)
// 3. remainder * 2 = product
// 4. product mod 11 = new remainder
// 5. Repeat for each digit
// 6. Last digit is (11 - final remainder) mod 10
func ValidatePIBChecksum(pib string) bool {
	if len(pib) != 9 {
		return false
	}

	// Check if all characters are digits
	for _, ch := range pib {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	ostatak := 10
	for i := 0; i < len(pib)-1; i++ {
		cifra := int(pib[i] - '0')
		zbir := cifra + ostatak
		medjuostatak := zbir % 10
		if medjuostatak == 0 {
			medjuostatak = 10
		}
		ostatak = (medjuostatak * 2) % 11
	}

	kontrol := (11 - ostatak) % 10
	lastDigit := int(pib[len(pib)-1] - '0')

	return lastDigit == kontrol
}
func (s *PartneriResource) GetCompanyByPIB(pib string) (*domain.CompanyInfo, error) {
	xmlBody := fmt.Sprintf(`
	<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema">
		<SOAP-ENV:Header>
			<AuthenticationHeader xmlns="http://communicationoffice.nbs.rs">
				<UserName>lavkompjuteri</UserName>
				<Password>lav231055</Password>
				<LicenceID>3c6ea969-57f7-4c75-b2a8-576f8eccf578</LicenceID>
			</AuthenticationHeader>
		</SOAP-ENV:Header>
		<SOAP-ENV:Body>
			<GetCompany xmlns="http://communicationoffice.nbs.rs">
				<taxIdentificationNumber xsi:type="xsd:int">%s</taxIdentificationNumber>
			</GetCompany>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`, pib)

	req, err := http.NewRequest("POST",
		"https://webservices.nbs.rs/CommunicationOfficeService1_0/CoreXmlService.asmx",
		bytes.NewBufferString(xmlBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("greška pri slanju zahteva: %v", err)
	}
	defer resp.Body.Close()

	bodyStr, err := readResponseBodyToUTF8(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("NBS greška (%d): %s", resp.StatusCode, bodyStr)
	}

	resultXML := common.ExtractInnerXML(bodyStr, "<GetCompanyResult>", "</GetCompanyResult>")
	if resultXML == "" {
		return nil, fmt.Errorf("nije pronađen GetCompanyResult u odgovoru")
	}

	innerXML := html.UnescapeString(resultXML)

	info := domain.CompanyInfo{
		Naziv:  common.ExtractTag(innerXML, "ShortName"),
		Naziv1: common.ExtractTag(innerXML, "Name"),
		Mesto:  common.ExtractTag(innerXML, "City"),
		PBroj:  common.ExtractTag(innerXML, "PostalCode"),
		Adresa: common.ExtractTag(innerXML, "Address"),
		MatBr:  common.ExtractTag(innerXML, "NationalIdentificationNumber"),
	}

	// Trim svih polja
	info.Naziv = strings.TrimSpace(info.Naziv)
	info.Naziv1 = strings.TrimSpace(info.Naziv1)
	info.Mesto = strings.TrimSpace(info.Mesto)
	info.PBroj = strings.TrimSpace(info.PBroj)
	info.Adresa = strings.TrimSpace(info.Adresa)
	info.MatBr = strings.TrimSpace(info.MatBr)

	return &info, nil
}

// ---- Dohvatanje računa ----
func (s *PartneriResource) GetCompanyAccountsByPIB(tbl *domain.TableData, pib string) error {
	xmlBody := fmt.Sprintf(`
	<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema">
		<SOAP-ENV:Header>
			<AuthenticationHeader xmlns="http://communicationoffice.nbs.rs">
				<UserName>lavkompjuteri</UserName>
				<Password>lav231055</Password>
				<LicenceID>3c6ea969-57f7-4c75-b2a8-576f8eccf578</LicenceID>
			</AuthenticationHeader>
		</SOAP-ENV:Header>
		<SOAP-ENV:Body>
			<GetCompanyAccount xmlns="http://communicationoffice.nbs.rs">
				<taxIdentificationNumber xsi:type="xsd:int">%s</taxIdentificationNumber>
			</GetCompanyAccount>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`, pib)

	req, err := http.NewRequest("POST",
		"https://webservices.nbs.rs/CommunicationOfficeService1_0/CompanyAccountXmlService.asmx",
		bytes.NewBufferString(xmlBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("greška pri slanju zahteva: %v", err)
	}
	defer resp.Body.Close()

	bodyStr, err := readResponseBodyToUTF8(resp)
	if err != nil {
		return err
	}

	resultXML := common.ExtractInnerXML(bodyStr, "<GetCompanyAccountResult>", "</GetCompanyAccountResult>")
	if resultXML == "" {
		return fmt.Errorf("nije pronađen GetCompanyAccountResult")
	}
	innerXML := html.UnescapeString(resultXML)

	var racuni []domain.CompanyAccount
	for {
		frag := common.ExtractInnerXML(innerXML, "<CompanyAccount>", "</CompanyAccount>")
		if frag == "" {
			break
		}

		acc := domain.CompanyAccount{
			BankCode:      common.ExtractTag(frag, "BankCode"),
			AccountNumber: common.ExtractTag(frag, "AccountNumber"),
			ControlNumber: common.ExtractTag(frag, "ControlNumber"),
			BankName:      common.ExtractTag(frag, "BankName"),
		}

		if acc.BankCode != "" && acc.AccountNumber != "" && acc.ControlNumber != "" {
			acc.FullAccount = fmt.Sprintf("%s-%s-%s", acc.BankCode, acc.AccountNumber, acc.ControlNumber)
			racuni = append(racuni, acc)
		}

		pos := strings.Index(innerXML, "</CompanyAccount>") + len("</CompanyAccount>")
		if pos < len(innerXML) {
			innerXML = innerXML[pos:]
		} else {
			break
		}
	}
	for i, rac := range racuni {
		fields := []string{fmt.Sprintf("%d", i+1), rac.FullAccount, rac.BankName}
		row := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: true}
		tbl.Rows = append(tbl.Rows, row)
	}
	return nil
}

// ---- Pomoćna funkcija: čitanje i konverzija u UTF-8 ----
func readResponseBodyToUTF8(resp *http.Response) (string, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("greška pri čitanju odgovora: %v", err)
	}

	// automatsko detektovanje i konverzija u UTF-8
	r, err := charset.NewReader(bytes.NewReader(raw), resp.Header.Get("Content-Type"))
	if err != nil {
		// fallback na windows-1250 ako detekcija nije uspela
		r, err = charset.NewReaderLabel("windows-1250", bytes.NewReader(raw))
		if err != nil {
			// fallback: ISO-8859-2
			r, err = charset.NewReaderLabel("iso-8859-2", bytes.NewReader(raw))
			if err != nil {
				// ako baš ništa ne uspe, vrati original kao UTF-8
				return string(raw), nil
			}
		}
	}

	conv, err := io.ReadAll(r)
	if err != nil {
		return string(raw), nil
	}

	return string(conv), nil
}

func (s *PartneriResource) setServiceFieldValues() {
	s.partneriTableFields = []domain.Fields{
		{Name: "sifra", Label: "Sifra", ValidationText: "morate uneti sifru partnera...", Width: "8"},
		{Name: "naziv", Label: "Naziv", ValidationText: "morate uneti naziv partnera..", Width: "60"},
		{Name: "adresa", Label: "Adresa", Width: "40"},
		{Name: "pobro", Label: "Postanski broj", Width: "8"},
		{Name: "mesto", Label: "Mesto", Width: "40"},
		{Name: "pib", Label: "PIB", Width: "12"},
		{Name: "jmbg", Label: "JMBG", Width: "15"},
		{Name: "bpg", Label: "BPG"},
		{Name: "index", Label: "Index"},
		{Name: "gln", Label: "GLN"},
		{Name: "jib", Label: "JIB"},
		{Name: "ziro", Label: "Ziro"},
		{Name: "matbr", Label: "Maticni Broj"},
		{Name: "konta", Label: "Konta"},
		{Name: "tippdv", Label: "Tip PDV"},
		{Name: "email", Label: "E-Mail"},
		{Name: "telefon", Label: "Telefon"},
		{Name: "kontaktosb", Label: "Kontakt osoba"},
		{Name: "budzetski", Label: "Budzetski"},
		{Name: "jbkjs", Label: "JBKJS"},
		{Name: "napomena", Label: "Naponema"},
		{Name: "idpartneri", Label: "ID Partneri"},
	}

	s.tekuciRacuniTableFields = []domain.Fields{
		{Name: "redbroj", Label: "Redni broj"},
		{Name: "brojracuna", Label: "Broj računa"},
		{Name: "banka", Label: "Banka"},
	}
}

func (s *PartneriResource) PrepareInsertUpdateFields(partner *domain.Partneri) []domain.Fields {
	var fields []domain.Fields

	partneriType := reflect.TypeOf(domain.Partneri{})
	partneriValue := reflect.ValueOf(*partner)

	for i := 0; i < partneriType.NumField(); i++ {
		field := partneriType.Field(i)
		fieldValue := partneriValue.Field(i)

		// Check if field has 'form' tag
		formTag, ok := field.Tag.Lookup("form")
		if !ok || formTag == "" || formTag == "-" {
			continue
		}

		// Convert field value to string based on its type
		var valueStr string
		switch v := fieldValue.Interface().(type) {
		case string:
			valueStr = v
		case int, int32, int64:
			valueStr = fmt.Sprintf("%v", v)
		case bool:
			if v {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		case float32, float64:
			valueStr = fmt.Sprintf("%f", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		// Create domain.Fields entry with form tag as label and actual value
		domainField := domain.Fields{
			Name:  strings.ToLower(field.Name),
			Label: formTag,
			Value: valueStr,
		}

		fields = append(fields, domainField)
	}

	return fields
}

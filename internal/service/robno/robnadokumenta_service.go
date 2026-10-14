package robno

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
)

// RobnoDokumentaService defines the operations of the "Robna dokumenta" option
// (unos / pregled / specifikacije / kontiranje / prepis robnih dokumenata i prikazi naloga).
//
// Implemented so far: the shared helpers, the combos and the tab "Unos dokumenta".
// The queries of the remaining tabs are added as the single tabs are implemented, e.g.:
//
//	GetPregledDokumenata(context.Context, *domain.TableData, domain.RobnoDokumentaParams) error
type RobnoDokumentaService interface {
	GetFvrData(context.Context) (domain.Fvr, error)
	GetMagacinComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error)
	GetTipdokComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error)
	GetVrstaDokumentaComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error)

	// Tab 1 - Unos dokumenta
	GetUnosComboValues(ctx context.Context, tipdok string, moduli, tipoviMagacina []string) (magValues, vrstaDokumentaValues []domain.ComboItem, err error)
	GetUnosDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error
	GetUnosDokumentaTotal(ctx context.Context, total *domain.RobnoDokumentaTotal) error
	GetNextNalog(ctx context.Context, tipdok string) (int, error)
	GetByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Rnal, error)
	GetByTipdokNalogMagacin(ctx context.Context, tipdok string, nalog, magaciniID int) (domain.Rnal, domain.Magacini, error)
	ValidateRobnoUnosNaloga(ctx context.Context, params domain.RobnoDokumentaParams) []domain.FieldError
	// The entry of the documents of the nalog (robnadokumenta_unos_service.go).
	ValidateRobFakture(ctx context.Context, params domain.RobnoFaktureParams) []domain.FieldError
	GetRobFaktureFlags(ctx context.Context, vrd int) (pdvObveznik, storno bool, err error)
	GetRobFaktureDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoFaktureDokument, error)
	// The functions of the entry every document can use (the legacy procedures of the windows).
	GetRobniDokument(ctx context.Context, magaciniID, vrd, dokum int) (domain.RobniDokumentRow, bool, error)
	GetRobniDokumentByID(ctx context.Context, rdokID int64) (domain.RobniDokumentRow, bool, error)
	GetRobnoUnosNalog(ctx context.Context, rnalID int64) (domain.Rnal, bool, error)
	GetNoviBrojDokumenta(ctx context.Context, magaciniID, vrd int) (int, error)
	GetRobKupac(ctx context.Context, konto, sifra string) (domain.RobKupacDto, bool, error)
	GetRobSaldoKupca(ctx context.Context, konto, sifra string, datum time.Time) (domain.RobnoFaktureFinStanjeKupca, error)
	GetRobAvansiDokumenta(ctx context.Context, rdokID int64) ([]domain.RobAvansDto, error)
	GetRobZiroRacuni(ctx context.Context) ([]domain.ComboItem, error)
	SaveRobFaktura(ctx context.Context, params domain.RobnoFaktureParams) (int64, error)
	SaveRobniDokument(ctx context.Context, dok domain.RobniDokumentSnimanje) (int64, error)
	ObracunFakture(ctx context.Context, rdokID int64) error
	PreracunRataDokumenta(ctx context.Context, rdokID int64) error
	// The stavke of the documents.
	GetRobStavkeDokumenta(ctx context.Context, rdokID int64) ([]domain.RobStavkaRow, error)
	GetRobDokumentaPrikaz(ctx context.Context, magaciniID int, vrste []int, params domain.RobDokumentaPrikazParams) ([]domain.RobDokumentPrikazRow, int, error)
	GetRobStavkeKontekst(ctx context.Context, rdokID int64) (domain.RobStavkeKontekst, bool, error)
	GetRobArtikalStanje(ctx context.Context, magaciniID int, konto string, sifra int) (domain.RobArtikalStanjeDto, bool, error)
	GetRobKontaMagacina(ctx context.Context, magaciniID int) ([]string, error)
	GetRobZadnjaCena(ctx context.Context, rdokID int64, sifra int) (domain.RobZadnjaCenaDto, bool, error)
	ValidateRobStavka(ctx context.Context, params domain.RobStavkaParams) []domain.FieldError
	SaveRobStavka(ctx context.Context, params domain.RobStavkaParams) (int64, error)
	GetRobStavkaArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobStavkaArtikal, bool, error)
	GetRobStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobStavkaForma, bool, error)
	DeleteRobStavka(ctx context.Context, rproID int64) (int64, error)
	CreateRobnoNalog(ctx context.Context, params domain.RobnoDokumentaParams) (int64, error)
	UpdateRobnoNalog(ctx context.Context, rnalID int64, params domain.RobnoDokumentaParams) error

	// Tab 2 - Pregled dokumenta (sub-tabs "Štampa" and "eFaktura")
	// GetDokumentaPreview is the list of the robni dokumenti the "Obrada" of the "Štampa" sub-tab shows,
	// the same for every vrsta dokumenta (the print of a document is chosen by its vrsta).
	GetDokumentaPreview(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error
	GetPregledEFaktura(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error

	// Tab 3 - Specifikacije dokumenta (the grid is the one of the "Štampa" sub-tab of "Pregled
	// dokumenta"; the tab adds the ranges of the vrste naloga, of the broj naloga and of the broj
	// dokumenta and the state of the print).
	GetSpecifikacijeDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Tab 4 - Kontiranje dokumenata (sub-tabs "Knjiženje dokumenata", "Pregled proknjiženih /
	// neproknjiženih dokumenata" and "Pregled proknjiženih / neproknjiženih dokumenata po
	// magacinima").
	GetKontiranjeKnjizenje(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error
	GetKontiranjePregled(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error
	GetKontiranjePoMagacinima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error

	// Tab 6 - Prikaz ukupne obrade
	GetPrikazUkupneObrade(ctx context.Context, tbl *domain.TableData, getTotRecords bool, currentPage, pageSize int, printType string) error

	// Tab 7 - Prikaz naloga
	GetPrikazNaloga(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Tab 8 - Prikaz dokumenata u nalogu (the parameters of the tab are the same selection of nalozi
	// as the "Prikaz naloga" tab, its rows are the robni dokumenti of those nalozi).
	GetPrikazDokumenataUNalogu(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Tab 9 - Prikaz dokumenata po operateru (the robni dokumenti of tab 8 grouped by the operater of
	// the document, with the number of his documents and of their stavke and his duguje/potražuje).
	GetPrikazDokumenataPoOperateru(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error

	// Štampa fakture: the fakture of the selection ready to print (the report RobnoStampaFaktura, the
	// legacy ROB_RPT_STAMPA_FAKTURA) and the izdavalac (fvr) of the print.
	GetStampaFaktura(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	// Štampa popisa (vrsta dokumenta 101, the report RobnoStampaPopis) and the vrsta dokumenta of a robni
	// dokument (the print of the selected document is chosen by its vrsta).
	GetStampaPopis(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaFakturaMP(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaOpstiDokument(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaFakturaUsluge(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaZaduzenjeCO(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaRazduzenjeCO(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaFakturaAvansni(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaFakturaIzvoz(ctx context.Context, params domain.RobnoStampaFakturaParams, withKomercOpis bool) ([]domain.RobnoStampaFakturaIzvozView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaFakturaKnjizno(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaFakturaOtpremnica(ctx context.Context, params domain.RobnoStampaFakturaParams, tipFakture int) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaProfakturaIzvoz(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaIzvozView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaProfaktura(ctx context.Context, params domain.RobnoStampaFakturaParams, tipProfakture int) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaPrenosnica(ctx context.Context, params domain.RobnoStampaFakturaParams, opcije RobnoStampaPrenosnicaOpcije) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaKnjiznoPismo(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaKnjiznoPismoFin(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaFakturaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaPopisTekGod(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaZaduzenjeSI(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaZaduzenjeGradilista(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaTrebovanje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaNivelacija(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaPopisView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaInterniPrenosProizvodnje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaInterniPrenosView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaInternaZakljucnica(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaInterniPrenosView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaInternaZakljucnicaRacun(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaInternaZakljucnicaRacunView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaNivelacijaMaloprodaje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaNivelacijaMPView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaKalkulacija(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaKalkulacijaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaKalkulacijaMaloprodaje(ctx context.Context, params domain.RobnoStampaFakturaParams) ([]domain.RobnoStampaKalkulacijaView, domain.RobnoStampaFakturaFirmaDto, error)
	GetStampaDokument(ctx context.Context, rdokID int64, vrd string) (domain.RobnoStampaDokumentDto, error)

	GetUnosDokumentaTableFields() []domain.Fields
	GetPregledDokumentaTableFields() []domain.Fields
	GetDokumentaPreviewTableFields() []domain.Fields
	GetPregledEFakturaTableFields() []domain.Fields
	GetSpecifikacijeDokumentaTableFields() []domain.Fields
	GetKontiranjeKnjizenjeTableFields() []domain.Fields
	GetKontiranjePregledTableFields() []domain.Fields
	GetKontiranjePoMagacinimaTableFields() []domain.Fields
	GetPrepisDokumentaTableFields() []domain.Fields
	GetPrikazUkupneObradeTableFields() []domain.Fields
	GetPrikazNalogaTableFields() []domain.Fields
	GetPrikazDokumenataUNaloguTableFields() []domain.Fields
	GetPrikazDokumenataPooperateruTableFields() []domain.Fields
	GetGrupaDokumenta(ctx context.Context, vrd int) (string, error)
	GetFaktureStavkeTableFields() []domain.Fields
	GetFaktureAvansiTableFields() []domain.Fields
	GetFaktureDokumentaTableFields() []domain.Fields
	GetFaktureDokumentStavkeTableFields() []domain.Fields
	GetInterniStavkeTableFields() []domain.Fields
	GetRobInterniDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoInterniDokument, error)
	ValidateRobInterni(ctx context.Context, params domain.RobnoInterniParams) []domain.FieldError
	SaveRobInterni(ctx context.Context, params domain.RobnoInterniParams) (int64, error)
	DeleteRobFaktura(ctx context.Context, rdokID int64) (string, error)
	DeleteRobInterni(ctx context.Context, rdokID int64) (string, error)
	// Fakture usluga (the legacy rob_Fakture_Usluga and rob_FAKTURE_Usluga_DET2).
	GetRobUslugeDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoUslugeDokument, error)
	ValidateRobUsluge(ctx context.Context, params domain.RobnoFaktureParams) []domain.FieldError
	SaveRobUsluge(ctx context.Context, params domain.RobnoFaktureParams) (int64, error)
	DeleteRobFakturaUsluge(ctx context.Context, rdokID int64) (string, error)
	GetRobUslugaArtikal(ctx context.Context, rdokID int64, sifraIliBarkod string) (domain.RobUslugaArtikal, bool, error)
	GetRobUslugaStopa(ctx context.Context, rdokID int64, tarifa int) (domain.RobUslugaStopa, error)
	ValidateRobUslugaStavka(ctx context.Context, params domain.RobUslugaStavkaParams) []domain.FieldError
	SaveRobUslugaStavka(ctx context.Context, params domain.RobUslugaStavkaParams) (int64, error)
	DeleteRobUslugaStavka(ctx context.Context, rproID int64) (int64, error)
	GetRobUslugaStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobUslugaStavkaForma, bool, error)
	GetUslugeStavkeTableFields() []domain.Fields
	GetAvansniStavkeTableFields() []domain.Fields
	GetVrsteGrupeDokumenta(ctx context.Context, vrd, vrdOd, vrdDo int) ([]int, error)
	// Opšti dokumenti (the legacy rob_opsti_dokumenti_opd).
	GetRobOpstiDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoOpstiDokument, error)
	ValidateRobOpsti(ctx context.Context, params domain.RobnoOpstiParams) []domain.FieldError
	SaveRobOpsti(ctx context.Context, params domain.RobnoOpstiParams) (int64, error)
	GetRobOpstiArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto, otk, serija, rok string) (domain.RobOpstiArtikal, bool, error)
	RobPreKontoMagacina(ctx context.Context, magaciniID int) (konto, naziv string, err error)
	ValidateRobOpstiStavka(ctx context.Context, params domain.RobOpstiStavkaParams) []domain.FieldError
	SaveRobOpstiStavka(ctx context.Context, params domain.RobOpstiStavkaParams) (int64, error)
	DeleteRobOpstiStavka(ctx context.Context, rproID int64) (int64, error)
	GetOpstiStavkeTableFields() []domain.Fields
	GetOpstiFinStavkeTableFields() []domain.Fields
	GetOpstiKolStavkeTableFields() []domain.Fields
	GetOpstiMnvStavkeTableFields() []domain.Fields
	GetOpstiNivStavkeTableFields(lot bool) []domain.Fields
	GetOpstiPreStavkeTableFields() []domain.Fields
	GetOpstiGrdStavkeTableFields() []domain.Fields
	GetOpstiOtrStavkeTableFields() []domain.Fields
	GetRobInterniArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobInterniArtikal, bool, error)
	ValidateRobInterniStavka(ctx context.Context, params domain.RobInterniStavkaParams) []domain.FieldError
	SaveRobInterniStavka(ctx context.Context, params domain.RobInterniStavkaParams) (int64, error)
	DeleteRobInterniStavka(ctx context.Context, rproID int64) (int64, error)
	GetRobInterniStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobInterniStavkaForma, bool, error)
	// The porudžbenice (the group POR, the legacy rob_Porudzbenica and rob_PORUDZBENICE_DET).
	GetRobPorDokument(ctx context.Context, params domain.RobnoFaktureDokumentParams) (domain.RobnoPorDokument, error)
	ValidateRobPor(ctx context.Context, params domain.RobnoPorParams) []domain.FieldError
	SaveRobPor(ctx context.Context, params domain.RobnoPorParams) (int64, error)
	DeleteRobPor(ctx context.Context, rdokID int64) (string, error)
	GetRobPorArtikal(ctx context.Context, rdokID int64, sifraIliBarkod, konto string) (domain.RobPorArtikal, bool, error)
	ValidateRobPorStavka(ctx context.Context, params domain.RobPorStavkaParams) []domain.FieldError
	SaveRobPorStavka(ctx context.Context, params domain.RobPorStavkaParams) (int64, error)
	DeleteRobPorStavka(ctx context.Context, rproID int64) (int64, error)
	GetRobPorStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.RobPorStavkaForma, bool, error)
	GetPorStavkeTableFields() []domain.Fields
	GetInterniDokumentaTableFields() []domain.Fields
	GetInterniDokumentStavkeTableFields() []domain.Fields
}

type RobnoDokumentaResource struct {
	robnaDokRepo           repository.BaseRepository[domain.RobnoDokumentaDto]
	rnalTotalsRepo         repository.BaseRepository[domain.RobnoDokumentaTotalsDto]
	rnalHeaderRepo         repository.BaseRepository[domain.Rnal]
	prikazUkupneObradeRepo repository.BaseRepository[domain.PrikazUkupneObradeDto]
	tipdokRepo             repository.BaseRepository[domain.Tipdok]
	dokvrstaRepo           repository.BaseRepository[domain.Dokvrsta]
	magRepo                repository.BaseRepository[domain.Magacini]
	fvrRepo                repository.BaseRepository[domain.Fvr]
	commonSvc              commonsvc.CommonService

	// The queries of the štampa fakture (GetStampaFaktura): the stavke with the header of their
	// document, the avansi closed on the fakture, their rate and the izdavalac.
	stampaFakturaRepo             repository.BaseRepository[domain.RobnoStampaFakturaRowDto]
	stampaFakturaAvansRepo        repository.BaseRepository[domain.RobnoStampaFakturaAvansDto]
	stampaFakturaRateRepo         repository.BaseRepository[domain.RobnoStampaFakturaRataDto]
	stampaFakturaFirmaRepo        repository.BaseRepository[domain.RobnoStampaFakturaFirmaDto]
	stampaPopisRepo               repository.BaseRepository[domain.RobnoStampaPopisRowDto]
	stampaDokumentRepo            repository.BaseRepository[domain.RobnoStampaDokumentDto]
	stampaKalkulacijaRepo         repository.BaseRepository[domain.RobnoStampaKalkulacijaRowDto]
	stampaFakturaIzvozRepo        repository.BaseRepository[domain.RobnoStampaFakturaIzvozRowDto]
	stampaKnjiznoPismoRepo        repository.BaseRepository[domain.RobnoStampaKnjiznoPismoRowDto]
	stampaKnjiznoPismoFakturaRepo repository.BaseRepository[domain.RobnoStampaKnjiznoPismoFakturaDto]
	stampaInterniPrenosRepo       repository.BaseRepository[domain.RobnoStampaInterniPrenosRowDto]
	stampaPoreskaStopaRepo        repository.BaseRepository[domain.RobnoStampaPoreskaStopaDto]

	// The queries of the entry of the documents (robnadokumenta_unos_service.go).
	unos RobnoDokumentaUnosRepos

	// TODO: add the repositories needed by the remaining tabs (rdok - robni dokument,
	// rpro - robni promet, rsif - artikli, fkpl, ...). All the grids of the option share the row type
	// domain.RobnoDokumentaDto, so the repositories differ in the table of their queries only.

	unosDokumentaTableFields               []domain.Fields
	pregledDokumentaTableFields            []domain.Fields
	pregledStampaTableFields               []domain.Fields
	pregledEFakturaTableFields             []domain.Fields
	specifikacijeDokumentaTableFields      []domain.Fields
	kontiranjeKnjizenjeTableFields         []domain.Fields
	kontiranjePregledTableFields           []domain.Fields
	kontiranjePoMagacinimaTableFields      []domain.Fields
	prepisDokumentaTableFields             []domain.Fields
	prikazUkupneObradeTableFields          []domain.Fields
	prikazNalogaTableFields                []domain.Fields
	prikazNalogaPrintTableFields           []domain.Fields
	prikazDokumenataUNaloguTableFields     []domain.Fields
	prikazDokumenataPooperateruTableFields []domain.Fields
	faktureStavkeTableFields               []domain.Fields
	uslugeStavkeTableFields                []domain.Fields
	opstiStavkeTableFields                 []domain.Fields
	opstiFinStavkeTableFields              []domain.Fields
	opstiKolStavkeTableFields              []domain.Fields
	opstiMnvStavkeTableFields              []domain.Fields
	opstiNivStavkeTableFields              []domain.Fields
	opstiPreStavkeTableFields              []domain.Fields
	opstiGrdStavkeTableFields              []domain.Fields
	opstiOtrStavkeTableFields              []domain.Fields
	avansniStavkeTableFields               []domain.Fields
	faktureAvansiTableFields               []domain.Fields
	faktureDokumentaTableFields            []domain.Fields
	faktureDokumentStavkeTableFields       []domain.Fields
	interniStavkeTableFields               []domain.Fields
	interniDokumentaTableFields            []domain.Fields
	interniDokumentStavkeTableFields       []domain.Fields
	porStavkeTableFields                   []domain.Fields
}

// NewRobnoDokumentaService creates the service of the "Robna dokumenta" option.
func NewRobnoDokumentaService(
	robnaDokRepo repository.BaseRepository[domain.RobnoDokumentaDto],
	rnalTotalsRepo repository.BaseRepository[domain.RobnoDokumentaTotalsDto],
	rnalHeaderRepo repository.BaseRepository[domain.Rnal],
	prikazUkupneObradeRepo repository.BaseRepository[domain.PrikazUkupneObradeDto],
	tipdokRepo repository.BaseRepository[domain.Tipdok],
	dokvrstaRepo repository.BaseRepository[domain.Dokvrsta],
	magRepo repository.BaseRepository[domain.Magacini],
	fvrRepo repository.BaseRepository[domain.Fvr],
	commonSvc commonsvc.CommonService,
	stampaFakturaRepo repository.BaseRepository[domain.RobnoStampaFakturaRowDto],
	stampaFakturaAvansRepo repository.BaseRepository[domain.RobnoStampaFakturaAvansDto],
	stampaFakturaRateRepo repository.BaseRepository[domain.RobnoStampaFakturaRataDto],
	stampaFakturaFirmaRepo repository.BaseRepository[domain.RobnoStampaFakturaFirmaDto],
	stampaPopisRepo repository.BaseRepository[domain.RobnoStampaPopisRowDto],
	stampaDokumentRepo repository.BaseRepository[domain.RobnoStampaDokumentDto],
	stampaKalkulacijaRepo repository.BaseRepository[domain.RobnoStampaKalkulacijaRowDto],
	stampaFakturaIzvozRepo repository.BaseRepository[domain.RobnoStampaFakturaIzvozRowDto],
	stampaKnjiznoPismoRepo repository.BaseRepository[domain.RobnoStampaKnjiznoPismoRowDto],
	stampaKnjiznoPismoFakturaRepo repository.BaseRepository[domain.RobnoStampaKnjiznoPismoFakturaDto],
	stampaInterniPrenosRepo repository.BaseRepository[domain.RobnoStampaInterniPrenosRowDto],
	stampaPoreskaStopaRepo repository.BaseRepository[domain.RobnoStampaPoreskaStopaDto],
	unos RobnoDokumentaUnosRepos,
) *RobnoDokumentaResource {
	s := &RobnoDokumentaResource{
		robnaDokRepo:                  robnaDokRepo,
		rnalTotalsRepo:                rnalTotalsRepo,
		rnalHeaderRepo:                rnalHeaderRepo,
		prikazUkupneObradeRepo:        prikazUkupneObradeRepo,
		tipdokRepo:                    tipdokRepo,
		dokvrstaRepo:                  dokvrstaRepo,
		magRepo:                       magRepo,
		fvrRepo:                       fvrRepo,
		commonSvc:                     commonSvc,
		stampaFakturaRepo:             stampaFakturaRepo,
		stampaFakturaAvansRepo:        stampaFakturaAvansRepo,
		stampaFakturaRateRepo:         stampaFakturaRateRepo,
		stampaFakturaFirmaRepo:        stampaFakturaFirmaRepo,
		stampaPopisRepo:               stampaPopisRepo,
		stampaDokumentRepo:            stampaDokumentRepo,
		stampaKalkulacijaRepo:         stampaKalkulacijaRepo,
		stampaFakturaIzvozRepo:        stampaFakturaIzvozRepo,
		stampaKnjiznoPismoRepo:        stampaKnjiznoPismoRepo,
		stampaKnjiznoPismoFakturaRepo: stampaKnjiznoPismoFakturaRepo,
		stampaInterniPrenosRepo:       stampaInterniPrenosRepo,
		stampaPoreskaStopaRepo:        stampaPoreskaStopaRepo,
		unos:                          unos,
	}
	s.setTableFields()
	return s
}

// GetFvrData retrieves company (fvr) data of the current session (report headers).
func (s *RobnoDokumentaResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, &s.fvrRepo)
}

// GetMagacinComboValues returns the magacini of the current period keyed by magaciniid, filtered by the
// options (CommonService).
func (s *RobnoDokumentaResource) GetMagacinComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error) {
	return s.commonSvc.GetMagacinComboValues(ctx, opts...)
}

// GetTipdokComboValues returns the vrste naloga of the current period keyed by the tipdok code,
// filtered by the options (e.g. commonsvc.WithGrupeDokumenata).
func (s *RobnoDokumentaResource) GetTipdokComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error) {
	return s.commonSvc.GetTipdokComboValues(ctx, opts...)
}

// GetVrstaDokumentaComboValues returns the vrste dokumenata of the current period, filtered by the
// options (CommonService).
func (s *RobnoDokumentaResource) GetVrstaDokumentaComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error) {
	return s.commonSvc.GetVrstaDokumentaComboValues(ctx, opts...)
}

// GetUnosComboValues returns the magacini and the vrste dokumenata of the unos dokumenata for the vrsta
// naloga tipdok, like the legacy PostaviComboBoxes (run on the selection of the vrsta naloga):
//   - the magacini are the ones the user may work with (maguser) of the tipoviMagacina (magacini.tipmag,
//     the configuration grupe_dokumenata.tipmag_rob, the legacy "V;M;D"), limited to the magacini of the
//     vrsta naloga (tipdok.magacin, e.g. "1,") when it has them;
//   - the vrste dokumenata are the ones of the moduli (dokvrsta.modul, the configuration
//     grupe_dokumenata.dok_rob, the legacy gsDOKROB) that belong to the vrsta naloga: the ones whose
//     dokvrsta.grpnal has the vrsta naloga (e.g. "70", or a list "70,71,"); when no vrsta dokumenta has
//     it, all of them, limited to the vrste dokumenata of the vrsta naloga (tipdok.grpvrd, e.g.
//     "188,189,") when it has them.
//
// Without a vrsta naloga (or with an unknown one) the lists are not limited by it.
func (s *RobnoDokumentaResource) GetUnosComboValues(ctx context.Context, tipdok string, moduli, tipoviMagacina []string) (magValues, vrstaDokumentaValues []domain.ComboItem, err error) {
	var listMag, listVrd []int
	if tipdok = strings.TrimSpace(tipdok); tipdok != "" {
		if t, err := s.commonSvc.GetTipdokByCode(ctx, tipdok); err == nil {
			listMag = robnoDokumentaLista(t.Magacin)
			listVrd = robnoDokumentaLista(t.GrpVrd)
		}
		// The vrste dokumenata of the vrsta naloga (dokvrsta.grpnal); without them the lists above.
		grpnal, err := s.robnoDokumentaVrdGrpnal(ctx, tipdok)
		if err != nil {
			return nil, nil, err
		}
		if len(grpnal) > 0 {
			listVrd = grpnal
		}
	}
	magValues, err = s.commonSvc.GetMagacinKorisnikaComboValues(ctx,
		commonsvc.WithIn("magacini.tipmag", tipoviMagacina...),
		commonsvc.WithIn("magacini.mag", listMag...))
	if err != nil {
		return nil, nil, err
	}
	vrstaDokumentaValues, err = s.commonSvc.GetVrstaDokumentaComboValues(ctx,
		commonsvc.WithIn("modul", moduli...),
		commonsvc.WithIn("vrd", listVrd...))
	if err != nil {
		return nil, nil, err
	}
	return magValues, vrstaDokumentaValues, nil
}

// robnoDokumentaVrdGrpnal returns the vrste dokumenata (vrd) of the period that belong to the vrsta naloga
// tipdok: the ones whose dokvrsta.grpnal has it (one vrsta naloga, e.g. "70", or a list of them, "70,71,").
func (s *RobnoDokumentaResource) robnoDokumentaVrdGrpnal(ctx context.Context, tipdok string) ([]int, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	entities, err := s.dokvrstaRepo.GetAllCustom(ctx, `select vrd from dokvrsta d where d.god = $1 and d.kar = $2
		and exists (select 1 from regexp_split_to_table(coalesce(d.grpnal, ''), '[,;]') g where trim(g) = $3)
		order by vrd`, "", []any{userSession.SelectedGod, userSession.SelectedKar, strings.TrimSpace(tipdok)}, "", "")
	if err != nil || entities == nil {
		return nil, err
	}
	vrd := make([]int, 0, len(*entities))
	for _, d := range *entities {
		vrd = append(vrd, d.Vrd)
	}
	return vrd, nil
}

// robnoDokumentaLista returns the numbers of a comma separated list of a vrsta naloga (tipdok.grpvrd,
// tipdok.magacin, e.g. "188,189,"); the parts that are not numbers are skipped.
func robnoDokumentaLista(lista string) []int {
	var out []int
	for _, part := range strings.FieldsFunc(lista, func(r rune) bool { return r == ',' || r == ';' }) {
		if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// GetUnosDokumenta returns the robni nalozi (rnal) of the selected vrsta naloga for the grid: with a
// selected vrsta dokumenta the nalozi that have a document (rdok) of it.
func (s *RobnoDokumentaResource) GetUnosDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.unosDokumentaTableFields

	qb := common.NewQueryBuilder(`select rnalid, tipdok, nalog, magaciniid, mag, danal, opis, datob, brdo, brst, dug, pot from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	qb.AddCondition("rnal.tipdok", params.Tipdok, "=")
	if params.MagaciniID != 0 {
		qb.AddEqual("rnal.magaciniid", params.MagaciniID)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rnal.tipdok", "rnal.nalog", "rnal.opis", "rnal.oper"}, params.SearchText)
	}
	// The selected vrsta dokumenta: the nalozi that have a document of it (the last condition: it is
	// numbered from the arguments so far).
	if vrd, err := strconv.Atoi(strings.TrimSpace(params.Vrd)); err == nil && vrd > 0 {
		qb.AddCustomCondition(fmt.Sprintf("exists (select 1 from rdok d where d.rnalid = rnal.rnalid and d.vrd = $%d)", qb.GetArgsCount()+1), vrd)
	}
	qb.AddOrderBy("rnal.nalog desc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				strings.TrimSpace(entity.Tipdok),
				fmt.Sprintf("%d", entity.Nalog),
				magacinLabel(entity.Mag, entity.MagaciniID),
				nullDateLabel(entity.Danal),
				entity.Opis,
				nullDateLabel(entity.Datob),
				fmt.Sprintf("%d", entity.Brdo),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetUnosDokumentaTotal fills the "Prikaz ukupne obrade" panel of the tab.
func (s *RobnoDokumentaResource) GetUnosDokumentaTotal(ctx context.Context, total *domain.RobnoDokumentaTotal) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder(`select count(*) as uknaloga,
		coalesce(sum(brdo), 0)::int as ukdokumenata,
		coalesce(sum(brst), 0)::int as ukstavki,
		coalesce(sum(dug), 0) as ukduguje,
		coalesce(sum(pot), 0) as ukpotrazuje
		from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	sqlQuery, args := qb.Build()
	entities, err := s.rnalTotalsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	var row domain.RobnoDokumentaTotalsDto
	if len(*entities) > 0 {
		row = (*entities)[0]
	}
	total.UkNaloga = common.FormatNumberWithSystemLocale(row.UkNaloga, 0)
	total.UkDokumenata = common.FormatNumberWithSystemLocale(row.UkDokumenata, 0)
	total.UkStavki = common.FormatNumberWithSystemLocale(row.UkStavki, 0)
	total.Duguje = common.FormatNumberWithSystemLocale(row.UkDuguje, 2)
	total.Potrazuje = common.FormatNumberWithSystemLocale(row.UkPotrazuje, 2)
	total.Saldo = common.FormatNumberWithSystemLocale(row.UkDuguje-row.UkPotrazuje, 2)
	return nil
}

// GetNextNalog returns the next broj naloga of the given vrsta naloga (CommonService).
func (s *RobnoDokumentaResource) GetNextNalog(ctx context.Context, tipdok string) (int, error) {
	nextNalog, err := s.commonSvc.GetNextRnalNalog(ctx, tipdok)
	if err != nil {
		return 0, err
	}
	return int(nextNalog), nil
}

// GetByTipdokNalogMagacin returns the header of the robni nalog of a vrsta naloga, broj naloga and
// magacin (the legacy key GODKARNALOGTIPDOKMAG: the same broj naloga in another magacin is another
// nalog) and the magacin. The nalog is empty (RnalID 0) when it does not exist, and both are empty when
// the magacin does not exist (the validation reports it).
func (s *RobnoDokumentaResource) GetByTipdokNalogMagacin(ctx context.Context, tipdok string, nalog, magaciniID int) (domain.Rnal, domain.Magacini, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Rnal{}, domain.Magacini{}, fmt.Errorf("no user session found")
	}
	if magaciniID <= 0 {
		return domain.Rnal{}, domain.Magacini{}, nil
	}
	magacin, err := s.magacinByID(ctx, magaciniID)
	if err != nil {
		return domain.Rnal{}, domain.Magacini{}, nil
	}
	qb := common.NewQueryBuilder(`select rnalid, tipdok, idtipdok, nalog, danal, datob, opis, magaciniid, mag, dug, pot, brdo, brst, oper from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	qb.AddCondition("rnal.tipdok", strings.TrimSpace(tipdok), "=")
	qb.AddEqual("rnal.nalog", nalog)
	qb.AddEqual("rnal.magaciniid", magacin.MagaciniID)
	sqlQuery, args := qb.Build()
	entities, err := s.rnalHeaderRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Rnal{}, magacin, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Rnal{}, magacin, nil
	}
	return (*entities)[0], magacin, nil
}

// GetByTipdokNalog returns the header of the robni nalog of a vrsta naloga and broj naloga.
func (s *RobnoDokumentaResource) GetByTipdokNalog(ctx context.Context, tipdok string, nalog int) (domain.Rnal, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Rnal{}, fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder(`select rnalid, tipdok, idtipdok, nalog, danal, datob, opis, magaciniid, mag, dug, pot, brdo, brst, oper from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	qb.AddCondition("rnal.tipdok", tipdok, "=")
	qb.AddEqual("rnal.nalog", nalog)
	sqlQuery, args := qb.Build()
	entities, err := s.rnalHeaderRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Rnal{}, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Rnal{}, nil
	}
	return (*entities)[0], nil
}

// ValidateRobnoUnosNaloga validates the header of the "Unos dokumenta" form before the robni nalog is
// opened or continued ("Snimi nalog"), like the legacy validation of the robni nalozi and the checks of
// "Snimi nalog":
//   - the period of the knjiženje (fvr of god, kar and firma) must exist and its business year must not
//     be closed (fvr.godzatv); without them nothing else is checked;
//   - the vrsta naloga (tipdok) must exist and the broj naloga must be a whole number;
//   - the datum obrade and the datum naloga must be valid dates of the business year;
//   - the magacin must be given and exist, the vrsta dokumenta must be given and exist; a vrsta of the
//     group IRS is worked only from a magacin of the proizvodnja (tipmag P or L);
//   - a financial nalog (fnal) of the same vrsta and broj naloga fixes the datum naloga (the nalog is
//     transferred into it automatically), and so does a robni nalog (rnal) of the same vrsta and broj
//     naloga in another magacin.
func (s *RobnoDokumentaResource) ValidateRobnoUnosNaloga(ctx context.Context, params domain.RobnoDokumentaParams) []domain.FieldError {
	fieldErrors := []domain.FieldError{}
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("nalog", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	god := userSession.SelectedGod

	// The period of the knjiženje and its business year.
	fvr, found, err := s.periodKnjizenja(ctx)
	if err != nil {
		add("nalog", common.ErrMsgReadData)
		return fieldErrors
	}
	if !found {
		add("nalog", "Ne postoji otvoren period za knjiženje! Morate otvoriti period da biste mogli da nastavite knjiženje.")
		return fieldErrors
	}
	if fvr.GodZatv {
		add("nalog", "Poslovna godina je zatvorena, knjiženje nije dozvoljeno.")
		return fieldErrors
	}

	// The vrsta naloga.
	tipdok := strings.TrimSpace(params.Tipdok)
	tipdokOK := false
	if tipdok == "" {
		add("tipdok", common.ErrMsgObavezanPodatak)
	} else if _, err := s.commonSvc.GetTipdokByCode(ctx, tipdok); err != nil {
		add("tipdok", "Nepostojeća vrsta naloga.")
	} else {
		tipdokOK = true
	}

	// The broj naloga.
	nalog, nalogErr := strconv.Atoi(strings.TrimSpace(params.Nalog))
	if strings.TrimSpace(params.Nalog) == "" {
		add("nalog", common.ErrMsgObavezanPodatak)
	} else if nalogErr != nil {
		add("nalog", "Broj naloga mora biti ceo broj.")
	}
	nalogOK := strings.TrimSpace(params.Nalog) != "" && nalogErr == nil

	// The datum obrade and the datum naloga: valid dates of the business year.
	datob, err := common.StringToTimeWithLayout(params.Datob, common.HtmlLayout)
	if err != nil {
		add("datob", "Morate uneti korektan datum obrade naloga.")
	} else if datob.Year() != god {
		add("datob", fmt.Sprintf("Nekorektan datum obrade, godina mora biti jednaka poslovnoj godini %d.", god))
	}
	danal, err := common.StringToTimeWithLayout(params.Danal, common.HtmlLayout)
	danalOK := err == nil && danal.Year() == god
	if err != nil {
		add("danal", "Morate uneti datum naloga.")
	} else if danal.Year() != god {
		add("danal", fmt.Sprintf("Nekorektan datum naloga, godina mora biti jednaka poslovnoj godini %d.", god))
	}

	// The magacin.
	var magacin domain.Magacini
	magacinOK := false
	if params.MagaciniID <= 0 {
		add("magaciniid", "Morate izabrati magacin.")
	} else if magacin, err = s.magacinByID(ctx, params.MagaciniID); err != nil {
		add("magaciniid", "Nepostojeći magacin.")
	} else {
		magacinOK = true
	}

	// The vrsta dokumenta; IRS only from a magacin of the proizvodnja.
	if strings.TrimSpace(params.Vrd) == "" {
		add("vrd", "Ne postoji definisana vrsta dokumenta za izabranu vrstu naloga! Unesite vrste dokumenata za izabranu vrstu naloga.")
	} else if vrd, err := strconv.Atoi(strings.TrimSpace(params.Vrd)); err != nil {
		add("vrd", "Nepostojeća vrsta dokumenta.")
	} else if grpdok, err := s.GetGrupaDokumenta(ctx, vrd); err != nil {
		add("vrd", common.ErrMsgReadData)
	} else if grpdok == "" {
		add("vrd", "Nepostojeća vrsta dokumenta.")
	} else if strings.EqualFold(strings.TrimSpace(grpdok), "IRS") && magacinOK && !robnoMagacinProizvodnje(magacin) {
		add("vrd", "Ova vrsta dokumenta radi se samo iz magacina proizvodnje!")
	}

	if !tipdokOK || !nalogOK || !danalOK {
		return fieldErrors
	}
	// A financial nalog of the same vrsta and broj naloga fixes the datum naloga.
	fnal, found, err := s.commonSvc.GetFnalByTipdokNalog(ctx, tipdok, nalog)
	if err != nil {
		add("danal", common.ErrMsgReadData)
		return fieldErrors
	}
	if found && !sameDate(fnal.Danal, danal) {
		add("danal", fmt.Sprintf("Za izabranu vrstu naloga postoji otvoren nalog u finansijskom knjigovodstvu. "+
			"Zbog automatskog prenošenja naloga datum naloga mora biti: %s.", fnal.Danal.Format(common.DateLayout)))
	}
	// A robni nalog of the same vrsta and broj naloga in another magacin fixes the datum naloga.
	if magacinOK {
		nalozi, err := s.rnalNalozi(ctx, tipdok, nalog)
		if err != nil {
			add("danal", common.ErrMsgReadData)
			return fieldErrors
		}
		for _, r := range nalozi {
			if r.Mag.Valid && int(r.Mag.Int64) != magacin.Mag && r.Danal.Valid && !sameDate(r.Danal.Time, danal) {
				add("danal", fmt.Sprintf("Za vrstu naloga %s postoji otvoren nalog sa istim brojem u magacinu %d. "+
					"Datum naloga mora imati vrednost: %s.", tipdok, r.Mag.Int64, r.Danal.Time.Format(common.DateLayout)))
				break
			}
		}
	}
	return fieldErrors
}

// periodKnjizenja returns the period of the knjiženje of the session (fvr of god, kar and firma) with
// the flag of the closed business year (found false when the period does not exist).
func (s *RobnoDokumentaResource) periodKnjizenja(ctx context.Context) (domain.Fvr, bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Fvr{}, false, fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder("select god, kar, naziv, coalesce(godzatv, false) as godzatv from fvr", true)
	qb.AddEqual("god", userSession.SelectedGod)
	qb.AddEqual("kar", userSession.SelectedKar)
	qb.AddEqual("naziv", userSession.Firma)
	sqlQuery, args := qb.Build()
	entities, err := s.fvrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Fvr{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Fvr{}, false, nil
	}
	return (*entities)[0], true, nil
}

// rnalNalozi returns the robni nalozi (rnal) of the vrsta naloga and broj naloga in the current period,
// one per magacin.
func (s *RobnoDokumentaResource) rnalNalozi(ctx context.Context, tipdok string, nalog int) ([]domain.Rnal, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	qb := common.NewQueryBuilder(`select rnalid, tipdok, nalog, danal, magaciniid, mag from rnal`, true)
	qb.AddEqual("rnal.god", userSession.SelectedGod)
	qb.AddEqual("rnal.kar", userSession.SelectedKar)
	qb.AddEqual("rnal.tipdok", tipdok)
	qb.AddEqual("rnal.nalog", nalog)
	sqlQuery, args := qb.Build()
	entities, err := s.rnalHeaderRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, err
	}
	if entities == nil {
		return nil, nil
	}
	return *entities, nil
}

// robnoMagacinProizvodnje tells if a magacin is a magacin of the proizvodnja (tipmag P or L).
func robnoMagacinProizvodnje(m domain.Magacini) bool {
	tip := strings.ToUpper(strings.TrimSpace(m.Tipmag))
	return tip == "P" || tip == "L"
}

// sameDate tells if two times are the same calendar day.
func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// CreateUnosDokumenta inserts the header of a new robni nalog (rnal).
func (s *RobnoDokumentaResource) CreateRobnoNalog(ctx context.Context, params domain.RobnoDokumentaParams) (int64, error) {
	fields, err := s.unosDokumentaFields(ctx, params, true)
	if err != nil {
		return 0, err
	}
	// The entity is needed for the detection of the god/kar columns of rnal; the values are taken
	// from the field list above.
	return s.rnalHeaderRepo.Create(ctx, &domain.Rnal{}, common.IDrnal, fields)
}

// UpdateUnosDokumenta saves the header of an existing robni nalog (rnal).
func (s *RobnoDokumentaResource) UpdateRobnoNalog(ctx context.Context, rnalID int64, params domain.RobnoDokumentaParams) error {
	fields, err := s.unosDokumentaFields(ctx, params, false)
	if err != nil {
		return err
	}
	return s.rnalHeaderRepo.Update(ctx, &domain.Rnal{}, common.IDrnal, rnalID, fields)
}

// unosDokumentaFields returns the columns of the header of a robni nalog (TODO: rbr/nalsts).
func (s *RobnoDokumentaResource) unosDokumentaFields(ctx context.Context, params domain.RobnoDokumentaParams, insert bool) ([]domain.Fields, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("no user session found")
	}
	tipdok := strings.TrimSpace(params.Tipdok)
	idTipdok, err := s.commonSvc.GetTipdokIDByCode(ctx, tipdok)
	if err != nil {
		return nil, err
	}
	mag := 0
	if params.MagaciniID > 0 {
		magacin, err := s.magacinByID(ctx, params.MagaciniID)
		if err != nil {
			return nil, err
		}
		mag = magacin.Mag
	}
	fields := []domain.Fields{
		{Name: "tipdok", Value: tipdok},
		{Name: "idtipdok", Value: fmt.Sprintf("%d", idTipdok)},
		{Name: "nalog", Value: strconv.Itoa(common.StringToInt(params.Nalog))},
		{Name: "danal", Value: params.Danal},
		{Name: "datob", Value: params.Datob},
		{Name: "opis", Value: params.Opis},
		{Name: "magaciniid", Value: strconv.Itoa(params.MagaciniID)},
		{Name: "mag", Value: strconv.Itoa(mag)},
	}
	if !insert {
		return fields, nil
	}
	return append(fields,
		domain.Fields{Name: "rbr", Value: "0"},
		domain.Fields{Name: "dug", Value: "0"},
		domain.Fields{Name: "pot", Value: "0"},
		domain.Fields{Name: "brdo", Value: "0"},
		domain.Fields{Name: "brst", Value: "0"},
		domain.Fields{Name: "oper", Value: userSession.UserName},
		domain.Fields{Name: "nalsts", Value: ""},
		// pinalid links a robni nalog to the financial nalog it was transferred to; a new nalog has
		// no link yet, so NULL is written (the column default 0 does not exist in table pinal).
		domain.Fields{Name: "pinalid", IsNull: true},
	), nil
}

// magacinByID returns the magacin of the current period with the given magaciniid.
func (s *RobnoDokumentaResource) magacinByID(ctx context.Context, magaciniID int) (domain.Magacini, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Magacini{}, fmt.Errorf("no user session found")
	}
	hasGod, hasKar := s.magRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("select magaciniid, mag, opis, coalesce(tipmag, '') as tipmag from magacini", true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("magaciniid", magaciniID)
	sqlQuery, args := qb.Build()
	entities, err := s.magRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Magacini{}, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.Magacini{}, fmt.Errorf("%s: magaciniid=%d", common.ErrMsgNotFound, magaciniID)
	}
	return (*entities)[0], nil
}

// Tab 2 - Pregled dokumenta (sub-tabs "Štampa" and "eFaktura")
// GetDokumentaPreview fills the grid of the "Štampa" sub-tab of "Pregled dokumenta" (its "Obrada"): the
// robni dokumenti of the vrsta dokumenta, of the magacin and of the range of the datum naloga of the
// selection, the same list for every vrsta dokumenta. It is the Go translation of the WinDev procedure
// Obrada of the screen: for every document the nalog, its dates and the dospeće (datum dokumenta + rok),
// the izvorni dokument, the iznos, the kupac (fkto, fana) with the naziv of fkpl and the pib, jbkjs,
// adresa and mesto of the partner, the avansni računi closed with the document (avansfakt, as
// "god-kar-mag-vrd-dokum;"), the poreska napomena, the second konto (pkto, pana) with its naziv, the
// valuta ("sifval-naziv") and the kurs. The lookups of the legacy loop are joins of one query; the
// documents are ordered like the legacy (datum naloga, vrsta naloga, nalog, datum and broj dokumenta,
// all descending). The id of a row is the rdokid of the document: the print of the sub-tab prints the
// selected document with the print of its vrsta dokumenta.
func (s *RobnoDokumentaResource) GetDokumentaPreview(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pregledStampaTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rdok.rdokid,
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.datob,
			rdok.dokum,
			rdok.dadok,
			rdok.rok,
			rdok.datiz,
			coalesce(rdok.dokiz, '') as dokiz,
			coalesce(rdok.iznos, 0) as iznos,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(kupac.naziv, '') as naziv,
			coalesce(p.pib, '') as pib,
			coalesce(p.jbkjs, '') as jbkjs,
			coalesce(p.adresa, '') as adresa,
			coalesce(p.mesto, '') as mesto,
			coalesce((select string_agg(av.god || '-' || av.kar || '-' || av.mag || '-' || av.vrd || '-' || av.dokum || ';', '' order by av.dokum)
				from avansfakt af
				inner join rdok av on av.rdokid = af.avansid
				where af.rdokid = rdok.rdokid), '') as avansi,
			coalesce(rdok.pornapomena, '') as pornapomena,
			coalesce(rdok.pkto, '') as pkto,
			coalesce(rdok.pana, '') as pana,
			coalesce(drugi.naziv, '') as naziv1,
			case when valute.sifval is null then coalesce(rdok.sifval, 0)::text
				else rdok.sifval || '-' || coalesce(valute.naziv, '') end as valuta,
			coalesce(rdok.kurs, 0) as kurs,
			coalesce(rdok.tkonto, '') as tkonto
		from rdok`, true)
	// The naziv of the kupac (fkto, fana) and of the second konto (pkto, pana): the first fkpl of the
	// konto and šifra, so that a konto and šifra with more fkpl rows does not repeat the document.
	qb.AddJoin(`left join lateral (select f.naziv, f.idpartneri from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.fkto and f.sifra = rdok.fana
			and coalesce(rdok.fkto, '') <> '' and coalesce(rdok.fana, '') <> ''
		order by f.idfkpl limit 1) kupac on true`)
	qb.AddJoin("left join partneri p on p.idpartneri = kupac.idpartneri")
	qb.AddJoin(`left join lateral (select f.naziv from fkpl f
		where f.god = rdok.god and f.kar = rdok.kar and f.konto = rdok.pkto and f.sifra = rdok.pana
			and coalesce(rdok.pkto, '') <> '' and coalesce(rdok.pana, '') <> ''
		order by f.idfkpl limit 1) drugi on true`)
	qb.AddJoin("left join valute on valute.sifval = rdok.sifval")
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	if params.Vrd != "" {
		qb.AddCondition("rdok.vrd", params.Vrd, "=")
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.OdDanal != "" {
		qb.AddCondition("rdok.danal", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.danal", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rdok.dokiz", "rdok.fkto", "rdok.fana", "kupac.naziv", "p.pib", "p.mesto"}, params.SearchText)
	}
	qb.AddOrderBy("rdok.danal desc, rdok.tipdok desc, rdok.nalog desc, rdok.dadok desc, rdok.dokum desc")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			// The id of the row is the rdokid of the document: the print of the sub-tab prints the
			// selected document.
			ID: fmt.Sprintf("%d", entity.RdokID),
			Fields: []string{
				fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
				common.FormatNullTime(entity.Danal, common.DateLayout),
				common.FormatNullTime(entity.Datob, common.DateLayout),
				fmt.Sprintf("%d", entity.Dokum.Int64),
				common.FormatNullTime(entity.Dadok, common.DateLayout),
				common.AddDaysToNullTime(entity.Dadok, int(entity.Rok.Int64), common.DateLayout),
				common.FormatNullTime(entity.Datiz, common.DateLayout),
				entity.Dokiz,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				entity.Pib,
				entity.Jbkjs,
				entity.Adresa,
				entity.Mesto,
				entity.Avansi,
				entity.Pornapomena,
				entity.Pkto,
				entity.Pana,
				entity.Naziv1,
				entity.Valuta,
				common.FormatNumberWithSystemLocale(entity.Kurs, 4),
				entity.Tkonto,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// Tab 3 - Specifikacije dokumenta
// GetSpecifikacijeDokumenta fills the rows of the "Specifikacije dokumenta" tab: one row per stavka
func (s *RobnoDokumentaResource) GetSpecifikacijeDokumenta(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	isPrint := printType == common.TipStampePrint
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.specifikacijeDokumentaTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`
		select
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.vrd,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.rok, 0) as rok,
			case when coalesce(rdok.sifval, 0) = 0 then ''
				else coalesce(rdok.sifval::text || ' - ' || valute.naziv, rdok.sifval::text) end as valuta,
			coalesce(rpro.fkto, '') as fkto,
			coalesce(rpro.fana, '') as fana,
			coalesce(fkpl.naziv, '') as naziv,
			coalesce(rpro.iznos, 0) as iznos,
			round(coalesce(rpro.iznos, 0) * coalesce(rpro.rab, 0) / 100, 2) as rabat,
			round((coalesce(rpro.iznos, 0) - round(coalesce(rpro.iznos, 0) * coalesce(rpro.rab, 0) / 100, 2))
				* coalesce(ps.pp, 0) / 100, 2) as porez,
			coalesce(ps.pp, 0) as stopa,
			rdok.rnalid
		from rdok`, true)
	qb.AddJoin("inner join rpro on rpro.rdokid = rdok.rdokid")
	qb.AddJoin("inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rpro.fkto and fkpl.sifra = rpro.fana")
	qb.AddJoin("left join valute on valute.sifval = rdok.sifval")
	// The poreska stopa of the stavka (rpor.pp): the stopa of its "po" that was in force on the
	// datum of the document, like the other robno reports read it.
	qb.AddJoin(`left join lateral (select r.pp from rpor r
		where r.po = rpro.po and r.datum <= rdok.dadok
		order by r.datum desc limit 1) ps on true`)
	// The groups of the vrste dokumenta the report covers (the legacy FAK, PRE, DIR, ARU and FZR).
	qb.AddIn("dokvrsta.grpdok", []any{"FAK", "PRE", "DIR", "ARU", "FZR"})
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.Vrd != "" {
		qb.AddCondition("rpro.vrd", params.Vrd, "=")
	}
	if params.OdVrd != "" {
		qb.AddCondition("rdok.tipdok", params.OdVrd, ">=")
	}
	if params.DoVrd != "" {
		qb.AddCondition("rdok.tipdok", params.DoVrd, "<=")
	}
	if params.OdDanal != "" {
		qb.AddCondition("rdok.dadok", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.dadok", params.DoDanal, "<=")
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	addNumberCondition(qb, "rdok.dokum", params.OdDokum, ">=")
	addNumberCondition(qb, "rdok.dokum", params.DoDokum, "<=")
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rpro.fkto", "rpro.fana", "fkpl.naziv"}, params.SearchText)
	}
	// The rows of a nalog and of a document stay together (the bands of the report are the nalog and
	// the document), like the legacy ORDER BY RNALID, RDOKID.
	qb.AddOrderBy("rdok.rnalid, rdok.rdokid, rpro.rproid")
	if !getTotalRecords && !isPrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && !isPrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// The columns of the amounts of the report: the label of the bands is printed over the columns
	// before them and every amount in its own column.
	amountColumn := 0
	for i, header := range tbl.Headers {
		if header.Name == "iznos" {
			amountColumn = i
			break
		}
	}
	type amounts struct {
		iznos     float64
		rabat     float64
		porez     float64
		zanaplatu float64
	}
	type stopaAmounts struct {
		osnovica float64
		pdv      float64
	}
	bandRow := func(classRow, label string, sums amounts) domain.TableRow {
		cells := make([]string, len(tbl.Headers))
		cells[amountColumn] = label
		for i, value := range []float64{sums.iznos, sums.rabat, sums.porez, sums.zanaplatu} {
			if column := amountColumn + i; column < len(cells) {
				cells[column] = common.FormatNumberWithSystemLocale(value, 2)
			}
		}
		return domain.TableRow{ClassRow: classRow, Fields: cells}
	}

	rbr := 0
	currentNalog := int64(-1)
	nalogSums := amounts{}
	reportSums := amounts{}
	stopaSums := map[float64]*stopaAmounts{}
	var stopaOrder []float64
	flushNalog := func() {
		if currentNalog == -1 {
			return
		}
		tbl.Rows = append(tbl.Rows, bandRow("nalog-total", i18n.GetInstance().Label("Ukupno za nalog")+":", nalogSums))
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		if isPrint && entity.RnalID != currentNalog {
			flushNalog()
			currentNalog = entity.RnalID
			nalogSums = amounts{}
		}
		rbr++
		amountsOfStavka := amounts{
			iznos:     entity.Iznos,
			rabat:     entity.Rabat,
			porez:     entity.Porez,
			zanaplatu: entity.Iznos - entity.Rabat + entity.Porez,
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", rbr),
				entity.Tipdok,
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Vrd),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				nullInt64Label(entity.Rok),
				entity.Valuta,
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(amountsOfStavka.iznos, 2),
				common.FormatNumberWithSystemLocale(amountsOfStavka.rabat, 2),
				common.FormatNumberWithSystemLocale(amountsOfStavka.porez, 2),
				common.FormatNumberWithSystemLocale(amountsOfStavka.zanaplatu, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
		if !isPrint {
			continue
		}
		nalogSums.iznos += amountsOfStavka.iznos
		nalogSums.rabat += amountsOfStavka.rabat
		nalogSums.porez += amountsOfStavka.porez
		nalogSums.zanaplatu += amountsOfStavka.zanaplatu
		reportSums.iznos += amountsOfStavka.iznos
		reportSums.rabat += amountsOfStavka.rabat
		reportSums.porez += amountsOfStavka.porez
		reportSums.zanaplatu += amountsOfStavka.zanaplatu
		if _, found := stopaSums[entity.Stopa]; !found {
			stopaSums[entity.Stopa] = &stopaAmounts{}
			stopaOrder = append(stopaOrder, entity.Stopa)
		}
		stopaSums[entity.Stopa].osnovica += amountsOfStavka.iznos - amountsOfStavka.rabat
		stopaSums[entity.Stopa].pdv += amountsOfStavka.porez
	}
	if !isPrint {
		return nil
	}
	flushNalog()
	tbl.Rows = append(tbl.Rows, bandRow("report-total", i18n.GetInstance().Label("Ukupno za izvestaj")+":", reportSums))

	// The "Ambalaza" summary: the PDV of the selection per poreska stopa (the columns of the legacy
	// report are Stopa, Osnovica and PDV) and the total of the two amount columns.
	sort.Float64s(stopaOrder)
	totalOsnovica, totalPdv := 0.0, 0.0
	for _, stopa := range stopaOrder {
		sums := stopaSums[stopa]
		totalOsnovica += sums.osnovica
		totalPdv += sums.pdv
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			ClassRow: "ambalaza",
			Fields: []string{
				fmt.Sprintf("%s%%", common.FormatNumberWithSystemLocale(stopa, 2)),
				common.FormatNumberWithSystemLocale(sums.osnovica, 2),
				common.FormatNumberWithSystemLocale(sums.pdv, 2),
			},
		})
	}
	tbl.Rows = append(tbl.Rows, domain.TableRow{
		ClassRow: "ambalaza-total",
		Fields: []string{
			i18n.GetInstance().Label("Ukupno") + ":",
			common.FormatNumberWithSystemLocale(totalOsnovica, 2),
			common.FormatNumberWithSystemLocale(totalPdv, 2),
		},
	})
	return nil
}

// GetPregledEFaktura fills the grid of the "eFaktura" sub-tab (status of the documents).
func (s *RobnoDokumentaResource) GetPregledEFaktura(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.pregledEFakturaTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			coalesce(rdok.status_salinv, '') as statussalinv,
			rdok.datum_stat_salinv as datumstatsalinv,
			coalesce(rdok.komentar, '') as komentar,
			coalesce(rdok.cirinvoiceid, '') as cirinvoiceid,
			coalesce(rdok.vatrecordingstatus, '') as vatrecordingstatus,
			rdok.datum_stat_indvat as datumstatindvat,
			rdok.magaciniid,
			rnal.nalog,
			rnal.danal,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.fkto, '') as fkto,
			coalesce(rdok.fana, '') as fana,
			coalesce(fkpl.naziv, '') as naziv,
			coalesce(rdok.iznos, 0) as iznos,
			rdok.dop,
			coalesce(rdok.dokiz, '') as dokiz,
			case when coalesce(rdok.sifval, 0) = 0 then ''
				else coalesce(rdok.sifval::text || ' - ' || valute.naziv, rdok.sifval::text) end as valuta,
			coalesce(rdok.kurs, 0) as kurs,
			rdok.rdokid,
			rdok.salesinvoiceid
		from rnal`, true)
	qb.AddJoin("inner join rdok on rdok.rnalid = rnal.rnalid")
	qb.AddJoin("inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd")
	qb.AddJoin("left join fkpl on fkpl.god = rdok.god and fkpl.kar = rdok.kar and fkpl.vkonta = 1 and fkpl.konto = rdok.fkto and fkpl.sifra = rdok.fana")
	qb.AddJoin("left join partneri p on p.idpartneri = fkpl.idpartneri")
	qb.AddJoin("left join valute on valute.sifval = rdok.sifval")
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if groups := grupeDokumenata(params.GrupeDokumenata); len(groups) > 0 {
		qb.AddIn("dokvrsta.grpdok", groups)
	}
	if params.OdDanal != "" {
		qb.AddCondition("rdok.dadok", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.dadok", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.dokum", "rdok.dokiz", "rdok.fkto", "rdok.fana", "fkpl.naziv", "rdok.status_salinv", "rdok.cirinvoiceid"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.nalog desc, rdok.dokum desc")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.StatusSalinv,
				nullDateLabel(entity.DatumStatSalinv),
				entity.Komentar,
				entity.CirInvoiceID,
				entity.VatRecordingStatus,
				nullDateLabel(entity.DatumStatIndVat),
				nullInt64Label(entity.MagaciniID),
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				entity.Fkto,
				entity.Fana,
				entity.Naziv,
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				nullDateLabel(entity.Dop),
				entity.Dokiz,
				entity.Valuta,
				common.FormatNumberWithSystemLocale(entity.Kurs, 2),
				fmt.Sprintf("%d", entity.RdokID),
				nullFloat64Label(entity.SalesInvoiceID),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

//
// Tab 4 - Kontiranje dokumenata (sub-tabs "Knjiženje dokumenata", "Pregled proknjiženih /
// neproknjiženih dokumenata" and "Pregled proknjiženih / neproknjiženih dokumenata po magacinima")
//

// GetKontiranjeKnjizenje fills the grid of the "Knjiženje dokumenata" sub-tab (TODO).
func (s *RobnoDokumentaResource) GetKontiranjeKnjizenje(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.kontiranjeKnjizenjeTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(kontiranjeSelectQuery, true)
	addKontiranjeConditions(qb, params, hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddOrderBy("rdok.nalog, rdok.dokum")
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	return s.setKontiranjeRows(tbl, entities, getTotalRecords, pageSize)
}

// GetKontiranjePregled fills the grid of the "Pregled proknjiženih / neproknjiženih" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePregled(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	tbl.Headers = s.kontiranjePregledTableFields
	return s.getKontiranjePregledList(ctx, tbl, getTotalRecords, currentPage, pageSize, params, "rdok.nalog, rdok.dokum")
}

// GetKontiranjePoMagacinima fills the same grid grouped by the magacin of the document.
func (s *RobnoDokumentaResource) GetKontiranjePoMagacinima(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams) error {
	tbl.Headers = s.kontiranjePoMagacinimaTableFields
	return s.getKontiranjePregledList(ctx, tbl, getTotalRecords, currentPage, pageSize, params, "rdok.magaciniid, rdok.nalog, rdok.dokum")
}

// getKontiranjePregledList runs the shared query of the two "Pregled ..." sub-tabs.
func (s *RobnoDokumentaResource) getKontiranjePregledList(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, orderBy string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(kontiranjeSelectQuery, true)
	addKontiranjeConditions(qb, params, hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	addKontiranjeStatusCondition(qb, params.Proknjizen)
	qb.AddOrderBy(orderBy)
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	return s.setKontiranjeRows(tbl, entities, getTotalRecords, pageSize)
}

// knjigeProknjizen is the value of rdok.knjige_1 of a robni dokument that is already posted (the
// legacy screens keep 'D' = da / proknjižen and 'N' = ne / neproknjižen in the first of the four
// knjige flags of the document).
const knjigeProknjizen = "D"

// kontiranjeSelectQuery is the row of the grids of the "Kontiranje dokumenata" tab: the data of a
// robni dokument (rdok) shown by the three sub-tabs.
const kontiranjeSelectQuery = `
	select
		rdok.tipdok,
		rdok.nalog,
		rdok.danal,
		rdok.vrd,
		rdok.dokum,
		rdok.dadok,
		coalesce(rdok.iznos, 0) as iznos,
		coalesce(rdok.opis, '') as opis,
		coalesce(rdok.polje, '') as polje,
		rdok.rok,
		rdok.magaciniid,
		coalesce(rdok.knjige_1, '') as knjige1
	from rdok`

// addKontiranjeConditions adds the filters of the "Kontiranje dokumenata" tab.
func addKontiranjeConditions(qb *common.QueryBuilder, params domain.RobnoDokumentaParams, hasGod, hasKar bool, god, kar int) {
	if hasGod {
		qb.AddEqual("rdok.god", god)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", kar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.Tipdok != "" {
		qb.AddEqual("rdok.tipdok", params.Tipdok)
	}
	if params.Vrd != "" {
		qb.AddCondition("rdok.vrd", params.Vrd, "=")
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	addNumberCondition(qb, "rdok.dokum", params.OdDokum, ">=")
	addNumberCondition(qb, "rdok.dokum", params.DoDokum, "<=")
	if params.OdDanal != "" {
		qb.AddCondition("rdok.danal", params.OdDanal, ">=")
	}
	if params.DoDanal != "" {
		qb.AddCondition("rdok.danal", params.DoDanal, "<=")
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.tipdok", "rdok.dokum", "rdok.opis", "rdok.dokiz", "rdok.fkto", "rdok.fana"}, params.SearchText)
	}
}

// addKontiranjeStatusCondition adds the state of the posting (rdok.knjige_1).
func addKontiranjeStatusCondition(qb *common.QueryBuilder, proknjizen string) {
	if proknjizen == knjigeProknjizen {
		qb.AddCustomCondition("rdok.knjige_1 = 'D'")
		return
	}
	qb.AddCustomCondition("coalesce(rdok.knjige_1, '') <> 'D'")
}

// addNumberCondition adds a condition of a whole number (validated ranges only).
func addNumberCondition(qb *common.QueryBuilder, field, value, operator string) {
	if value == "" {
		return
	}
	if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
		return
	}
	qb.AddCondition(field, value, operator)
}

// setKontiranjeRows sets the total records or the rows of the page of the tab.
func (s *RobnoDokumentaResource) setKontiranjeRows(tbl *domain.TableData, entities *[]domain.RobnoDokumentaDto, getTotalRecords bool, pageSize int) error {
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		cells := make([]string, 0, len(tbl.Headers))
		for _, header := range tbl.Headers {
			cells = append(cells, kontiranjeCell(entity, header.Name))
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: cells, HasUpdate: false, HasDelete: false})
	}
	return nil
}

// kontiranjeCell renders one cell of the grids of the "Kontiranje dokumenata" tab.
func kontiranjeCell(entity domain.RobnoDokumentaDto, field string) string {
	switch field {
	case "tipdok":
		return entity.Tipdok
	case "nalog":
		return fmt.Sprintf("%d", entity.Nalog)
	case "danal":
		return nullDateLabel(entity.Danal)
	case "vrd":
		return nullInt64Label(entity.Vrd)
	case "dokum":
		return nullInt64Label(entity.Dokum)
	case "dadok":
		return nullDateLabel(entity.Dadok)
	case "iznos":
		return common.FormatNumberWithSystemLocale(entity.Iznos, 2)
	case "opis":
		return entity.Opis
	case "polje":
		return entity.Polje
	case "rok":
		return nullInt64Label(entity.Rok)
	case "magaciniid":
		return nullInt64Label(entity.MagaciniID)
	}
	return ""
}

// grupeDokumenata splits the "Grupe dokumenata" filter into the values of an IN condition.
func grupeDokumenata(value string) []any {
	groups := []any{}
	for _, part := range strings.Split(value, ",") {
		group := strings.ToUpper(strings.TrimSpace(part))
		if group == "" {
			continue
		}
		groups = append(groups, group)
	}
	return groups
}

// nullInt64Label renders a nullable number of the grid ("" when there is none).
func nullInt64Label(value sql.NullInt64) string {
	if !value.Valid {
		return ""
	}
	return fmt.Sprintf("%d", value.Int64)
}

// nullFloat64Label renders a nullable decimal number of the grid ("" when none).
func nullFloat64Label(value sql.NullFloat64) string {
	if !value.Valid {
		return ""
	}
	return common.FormatNumberWithSystemLocale(value.Float64, 0)
}

// magacinLabel renders the "Magacin" cell of the grid.
func magacinLabel(mag, magaciniID sql.NullInt64) string {
	if mag.Valid {
		return fmt.Sprintf("%d", mag.Int64)
	}
	if magaciniID.Valid {
		return fmt.Sprintf("%d", magaciniID.Int64)
	}
	return ""
}

// nullDateLabel renders a nullable date of the grid (dd.MM.yyyy).
func nullDateLabel(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(common.DateLayout)
}

// GetPrikazUkupneObrade fills the grid of the "Prikaz ukupne obrade" tab (totals per magacin).
func (s *RobnoDokumentaResource) GetPrikazUkupneObrade(ctx context.Context, tbl *domain.TableData, getTotRecords bool, currentPage, pageSize int, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazUkupneObradeTableFields
	qb := common.NewQueryBuilder(`SELECT
			m.mag,
				COALESCE(m.opis, '') AS opis,
				COALESCE(m.mesto, '') AS mesto,
				COALESCE(n.brojnaloga, 0)::int AS brojnaloga,
				COALESCE(d.ukdokumenata, 0)::int AS ukdokumenata,
				COALESCE(s.brstavki, 0)::int AS brstavki,
				COALESCE(n.duguje, 0) AS duguje,
				COALESCE(n.potrazuje, 0) AS potrazuje
			FROM magacini m
			LEFT JOIN (
				SELECT
					r.magaciniid,
					COUNT(*) AS brojnaloga,
					SUM(r.dug) AS duguje,
					SUM(r.pot) AS potrazuje
				FROM rnal r
				GROUP BY r.magaciniid
			) n
				ON n.magaciniid = m.magaciniid
			LEFT JOIN (
				SELECT
					r.magaciniid,
					COUNT(DISTINCT d.rdokid) AS ukdokumenata
				FROM rnal r
				JOIN rdok d
					ON d.rnalid = r.rnalid
				GROUP BY r.magaciniid
			) d
				ON d.magaciniid = m.magaciniid
			LEFT JOIN (
				SELECT
					r.magaciniid,
					COUNT(*) AS brstavki
				FROM rnal r
				JOIN rdok d
					ON d.rnalid = r.rnalid
				JOIN rpro p
					ON p.rdokid = d.rdokid
				GROUP BY r.magaciniid
			) s
				ON s.magaciniid = m.magaciniid `, true)
	qb.AddEqual("m.god", userSession.SelectedGod)
	qb.AddEqual("m.kar", userSession.SelectedKar)
	if !getTotRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	qb.AddOrderBy("m.mag")
	sqlQuery, args := qb.Build()
	entities, err := s.prikazUkupneObradeRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	var ukupnoBrojNaloga, ukupnoDokumenata, ukupnoStavki int
	var ukupnoDuguje, ukupnoPotrazuje float64
	for _, entity := range *entities {
		ukupnoBrojNaloga += entity.BrojNaloga
		ukupnoDokumenata += entity.UkDokumenata
		ukupnoStavki += entity.BrStavki
		ukupnoDuguje += entity.Duguje
		ukupnoPotrazuje += entity.Potrazuje
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", entity.Mag),
				entity.Opis,
				entity.Mesto,
				fmt.Sprintf("%d", entity.BrojNaloga),
				fmt.Sprintf("%d", entity.UkDokumenata),
				fmt.Sprintf("%d", entity.BrStavki),
				common.FormatNumberWithSystemLocale(entity.Duguje, 2),
				common.FormatNumberWithSystemLocale(entity.Potrazuje, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	// The printed report closes with the "Ukupno" line; the print preview of the tab (TipStampePreview)
	// shows the rows of the magacini only.
	if printType == common.TipStampePrint && len(tbl.Rows) > 0 {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			IsGroupTotal: true,
			Fields: []string{
				i18n.GetInstance().Label("Ukupno"),
				"",
				"",
				fmt.Sprintf("%d", ukupnoBrojNaloga),
				fmt.Sprintf("%d", ukupnoDokumenata),
				fmt.Sprintf("%d", ukupnoStavki),
				common.FormatNumberWithSystemLocale(ukupnoDuguje, 2),
				common.FormatNumberWithSystemLocale(ukupnoPotrazuje, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// Tab 7 - Prikaz naloga
// GetPrikazNaloga fills the grid of the "Prikaz naloga" tab (robni nalozi).
func (s *RobnoDokumentaResource) GetPrikazNaloga(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazNalogaTableFields
	if printType == common.TipStampePrint {
		tbl.Headers = s.prikazNalogaPrintTableFields
	}
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rnal.rnalid,
			rnal.tipdok,
			rnal.nalog,
			rnal.danal,
			rnal.datob,
			coalesce(rnal.brdo, 0) as brdo,
			coalesce(rnal.brst, 0) as brst,
			coalesce(rnal.dug, 0) as dug,
			coalesce(rnal.pot, 0) as pot,
			coalesce(rnal.xopunos, '') as oper,
			rnal.magaciniid
		from rnal`, true)
	if hasGod {
		qb.AddEqual("rnal.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rnal.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rnal.magaciniid", params.MagaciniID)
	}
	if params.OdVrd != "" {
		qb.AddCondition("rnal.tipdok", params.OdVrd, ">=")
	}
	if params.DoVrd != "" {
		qb.AddCondition("rnal.tipdok", params.DoVrd, "<=")
	}
	addNumberCondition(qb, "rnal.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rnal.nalog", params.DoNaloga, "<=")
	// The three optional filters of the tab, each one applied only when its checkbox is on.
	if params.ChkDatumNaloga {
		if params.OdDanal != "" {
			qb.AddCondition("rnal.danal", params.OdDanal, ">=")
		}
		if params.DoDanal != "" {
			qb.AddCondition("rnal.danal", params.DoDanal, "<=")
		}
	}
	if params.ChkDatumObrade {
		if params.OdDatob != "" {
			qb.AddCondition("rnal.datob", params.OdDatob, ">=")
		}
		if params.DoDatob != "" {
			qb.AddCondition("rnal.datob", params.DoDatob, "<=")
		}
	}
	if params.ChkOperator && params.Oper != "" {
		qb.AddCustomSearchCondition([]string{"rnal.xopunos", "rnal.xopizmene"}, params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rnal.tipdok", "rnal.nalog", "rnal.xopunos", "rnal.opis"}, params.SearchText)
	}
	qb.AddOrderBy("rnal.tipdok, rnal.nalog desc")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		if printType == common.TipStampePrint {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
					fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
					nullDateLabel(entity.Danal),
					nullDateLabel(entity.Datob),
					fmt.Sprintf("%d", entity.Brdo),
					fmt.Sprintf("%d", entity.Brst),
					common.FormatNumberWithSystemLocale(entity.Dug, 2),
					common.FormatNumberWithSystemLocale(entity.Pot, 2),
					entity.Oper,
				},
				HasUpdate: false,
				HasDelete: false})
		} else {
			tbl.Rows = append(tbl.Rows, domain.TableRow{
				Fields: []string{
					fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
					"🔽",
					fmt.Sprintf("%s-%d", entity.Tipdok, entity.Nalog),
					nullDateLabel(entity.Danal),
					nullDateLabel(entity.Datob),
					fmt.Sprintf("%d", entity.Brdo),
					fmt.Sprintf("%d", entity.Brst),
					common.FormatNumberWithSystemLocale(entity.Dug, 2),
					common.FormatNumberWithSystemLocale(entity.Pot, 2),
					entity.Oper,
				},
				HasUpdate: false,
				HasDelete: false})
		}
	}
	return nil
}

// Tab 8 - Prikaz dokumenata u nalogu
// GetPrikazDokumenataUNalogu fills the grid of the "Prikaz dokumenata u nalogu" tab: the robni
func (s *RobnoDokumentaResource) GetPrikazDokumenataUNalogu(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazDokumenataUNaloguTableFields
	hasGod, hasKar := s.robnaDokRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`
		select
			rdok.tipdok,
			rdok.nalog,
			rdok.danal,
			rdok.vrd,
			rdok.dokum,
			rdok.dadok,
			coalesce(rdok.brst, 0) as brst,
			coalesce(rdok.iznos, 0) as iznos,
			rdok.datob,
			coalesce(nullif(rdok.xopunos, ''), rdok.xopunos, '') as oper,
			rdok.magaciniid
		from rdok`, true)
	if hasGod {
		qb.AddEqual("rdok.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("rdok.kar", userSession.SelectedKar)
	}
	if params.MagaciniID != 0 {
		qb.AddEqual("rdok.magaciniid", params.MagaciniID)
	}
	if params.OdVrd != "" {
		qb.AddCondition("rdok.tipdok", params.OdVrd, ">=")
	}
	if params.DoVrd != "" {
		qb.AddCondition("rdok.tipdok", params.DoVrd, "<=")
	}
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	if params.ChkDatumNaloga {
		if params.OdDanal != "" {
			qb.AddCondition("rdok.danal", params.OdDanal, ">=")
		}
		if params.DoDanal != "" {
			qb.AddCondition("rdok.danal", params.DoDanal, "<=")
		}
	}
	if params.ChkDatumObrade {
		if params.OdDatob != "" {
			qb.AddCondition("rdok.datob", params.OdDatob, ">=")
		}
		if params.DoDatob != "" {
			qb.AddCondition("rdok.datob", params.DoDatob, "<=")
		}
	}
	// The operater filter of the legacy screen looks in both the user of the insert (xopunos) and
	// the user of the last change (xopizmene) of the document.
	if params.ChkOperator && params.Oper != "" {
		qb.AddCustomSearchCondition([]string{"rdok.xopunos", "rdok.xopizmene"}, params.Oper)
	}
	if params.SearchText != "" {
		qb.AddCustomSearchCondition([]string{"rdok.tipdok", "rdok.nalog", "rdok.dokum", "rdok.opis", "rdok.xopunos", "rdok.xopizmene"}, params.SearchText)
	}
	qb.AddOrderBy("rdok.nalog, rdok.vrd, rdok.dokum")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for i, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				fmt.Sprintf("%d", (currentPage-1)*pageSize+i+1),
				fmt.Sprintf("%d", entity.Nalog),
				nullDateLabel(entity.Danal),
				nullInt64Label(entity.Vrd),
				nullInt64Label(entity.Dokum),
				nullDateLabel(entity.Dadok),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				nullDateLabel(entity.Datob),
				entity.Oper,
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// Tab 9 - Prikaz dokumenata po operateru
// GetPrikazDokumenataPoOperateru fills the grid of the "Prikaz dokumenata po operateru" tab: the
// robni dokumenti (rdok) of the selection grouped by the operater of the document (its xopizmene
// when it was changed, else its xopunos), with the number of his documents and of their stavke and
// the sums of duguje/potražuje of his documents by the konta 'D'/'P' (dokvrsta.kodknj), like the
// legacy procedure builds its per-operater table. With printType = TipStampePrint the whole result
// is returned unpaginated (the print of the tab).
func (s *RobnoDokumentaResource) GetPrikazDokumenataPoOperateru(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, currentPage, pageSize int, params domain.RobnoDokumentaParams, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	common.SetupTablePagination(tbl, currentPage, pageSize)
	tbl.Headers = s.prikazDokumenataPooperateruTableFields

	qb := common.NewQueryBuilder(`
		select
			coalesce(nullif(rdok.xopunos, ''), rdok.xopunos, '') as oper,
			count(*)::int as brdo,
			coalesce(sum(rdok.brst), 0)::int as brst,
			coalesce(sum(case when upper(coalesce(dokvrsta.kodknj, '')) = 'D' then coalesce(rdok.iznos, 0) else 0 end), 0) as dug,
			coalesce(sum(case when upper(coalesce(dokvrsta.kodknj, '')) = 'P' then coalesce(rdok.iznos, 0) else 0 end), 0) as pot
		from rdok
		left join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd`, true)
	// The period of the session must scope the documents. The row type of this option has no
	// god/kar db tags (GetHasGodHasKar reports false), so the two conditions are added explicitly.
	qb.AddEqual("rdok.god", userSession.SelectedGod)
	qb.AddEqual("rdok.kar", userSession.SelectedKar)
	addNumberCondition(qb, "rdok.nalog", params.OdNaloga, ">=")
	addNumberCondition(qb, "rdok.nalog", params.DoNaloga, "<=")
	if params.ChkDatumNaloga {
		if params.OdDanal != "" {
			qb.AddCondition("rdok.danal", params.OdDanal, ">=")
		}
		if params.DoDanal != "" {
			qb.AddCondition("rdok.danal", params.DoDanal, "<=")
		}
	}
	if params.ChkDatumObrade {
		if params.OdDatob != "" {
			qb.AddCondition("rdok.datob", params.OdDatob, ">=")
		}
		if params.DoDatob != "" {
			qb.AddCondition("rdok.datob", params.DoDatob, "<=")
		}
	}
	// The operater filter of the legacy screen looks in both the user of the insert (xopunos) and
	// the user of the last change (xopizmene) of the document.
	if params.ChkOperator && params.Oper != "" {
		qb.AddCustomSearchCondition([]string{"rdok.xopunos", "rdok.xopizmene"}, params.Oper)
	}
	qb.AddGroupBy("coalesce(nullif(rdok.xopunos, ''), rdok.xopunos, '')")
	qb.AddOrderBy("oper")
	if !getTotalRecords && printType != common.TipStampePrint {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.robnaDokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if getTotalRecords && printType != common.TipStampePrint {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}
	tbl.Rows = []domain.TableRow{}
	for _, entity := range *entities {
		tbl.Rows = append(tbl.Rows, domain.TableRow{
			Fields: []string{
				entity.Oper,
				fmt.Sprintf("%d", entity.Brdo),
				fmt.Sprintf("%d", entity.Brst),
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
			},
			HasUpdate: false,
			HasDelete: false,
		})
	}
	return nil
}

// GetUnosDokumentaTableFields returns the grid columns of the "Unos dokumenta" tab.
func (s *RobnoDokumentaResource) GetUnosDokumentaTableFields() []domain.Fields {
	return s.unosDokumentaTableFields
}

// GetPregledDokumentaTableFields returns the grid columns of the "Pregled dokumenta" tab.
func (s *RobnoDokumentaResource) GetPregledDokumentaTableFields() []domain.Fields {
	return s.pregledDokumentaTableFields
}

// GetDokumentaPreviewTableFields returns the grid columns of the "Štampa" sub-tab (the list of Obrada).
func (s *RobnoDokumentaResource) GetDokumentaPreviewTableFields() []domain.Fields {
	return s.pregledStampaTableFields
}

// GetPregledEFakturaTableFields returns the grid columns of the "eFaktura" sub-tab.
func (s *RobnoDokumentaResource) GetPregledEFakturaTableFields() []domain.Fields {
	return s.pregledEFakturaTableFields
}

// GetSpecifikacijeDokumentaTableFields returns the (still empty) grid columns of the tab.
func (s *RobnoDokumentaResource) GetSpecifikacijeDokumentaTableFields() []domain.Fields {
	return s.specifikacijeDokumentaTableFields
}

// GetKontiranjeKnjizenjeTableFields returns the grid columns of the "Knjiženje" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjeKnjizenjeTableFields() []domain.Fields {
	return s.kontiranjeKnjizenjeTableFields
}

// GetKontiranjePregledTableFields returns the grid columns of the "Pregled" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePregledTableFields() []domain.Fields {
	return s.kontiranjePregledTableFields
}

// GetKontiranjePoMagacinimaTableFields returns the columns of the "po magacinima" sub-tab.
func (s *RobnoDokumentaResource) GetKontiranjePoMagacinimaTableFields() []domain.Fields {
	return s.kontiranjePoMagacinimaTableFields
}

// GetPrepisDokumentaTableFields returns the (still empty) grid columns of the tab.
func (s *RobnoDokumentaResource) GetPrepisDokumentaTableFields() []domain.Fields {
	return s.prepisDokumentaTableFields
}

// GetPrikazUkupneObradeTableFields returns the columns of the "Prikaz ukupne obrade" tab.
func (s *RobnoDokumentaResource) GetPrikazUkupneObradeTableFields() []domain.Fields {
	return s.prikazUkupneObradeTableFields
}

// GetPrikazNalogaTableFields returns the grid columns of the "Prikaz naloga" tab.
func (s *RobnoDokumentaResource) GetPrikazNalogaTableFields() []domain.Fields {
	return s.prikazNalogaTableFields
}

// GetPrikazDokumenataUNaloguTableFields returns the columns of the "u nalogu" tab.
func (s *RobnoDokumentaResource) GetPrikazDokumenataUNaloguTableFields() []domain.Fields {
	return s.prikazDokumenataUNaloguTableFields
}

// GetPrikazDokumenataPooperateruTableFields returns the columns of the "po operateru" tab.
func (s *RobnoDokumentaResource) GetPrikazDokumenataPooperateruTableFields() []domain.Fields {
	return s.prikazDokumenataPooperateruTableFields
}

// GetFaktureStavkeTableFields returns the columns of the grid of the stavke of the "Fakture
// veleprodaje" screen.
func (s *RobnoDokumentaResource) GetFaktureStavkeTableFields() []domain.Fields {
	return s.faktureStavkeTableFields
}

// GetOpstiFinStavkeTableFields returns the columns of the grid of the stavke of the finansijski opšti dokumenti.
func (s *RobnoDokumentaResource) GetOpstiFinStavkeTableFields() []domain.Fields {
	return s.opstiFinStavkeTableFields
}

// GetOpstiKolStavkeTableFields returns the columns of the grid of the stavke of the količinski opšti dokumenti.
func (s *RobnoDokumentaResource) GetOpstiKolStavkeTableFields() []domain.Fields {
	return s.opstiKolStavkeTableFields
}

// GetOpstiMnvStavkeTableFields returns the columns of the grid of the stavke of the nivelacije (the group MNV).
func (s *RobnoDokumentaResource) GetOpstiMnvStavkeTableFields() []domain.Fields {
	return s.opstiMnvStavkeTableFields
}

// GetOpstiNivStavkeTableFields returns the columns of the grid of the stavke of the nivelacije veleprodaje (the
// group NIV); without lot the OTK broj, the serija and the rok are left out (the legacy columns of a magacin
// that is not of the tip zaliha 2).
func (s *RobnoDokumentaResource) GetOpstiNivStavkeTableFields(lot bool) []domain.Fields {
	if lot {
		return s.opstiNivStavkeTableFields
	}
	return s.opstiNivStavkeTableFields[:len(s.opstiNivStavkeTableFields)-3]
}

// GetOpstiStavkeTableFields returns the columns of the grid of the stavke of the opšti dokumenti.
func (s *RobnoDokumentaResource) GetOpstiStavkeTableFields() []domain.Fields {
	return s.opstiStavkeTableFields
}

// GetOpstiPreStavkeTableFields returns the columns of the grid of the stavke of a prenosnica (the legacy TABLE_rpro
// of rob_Prenosnice: the ones of the opšti dokument with the nabavna cena after the iznos).
func (s *RobnoDokumentaResource) GetOpstiPreStavkeTableFields() []domain.Fields {
	return s.opstiPreStavkeTableFields
}

// GetOpstiGrdStavkeTableFields returns the columns of the grid of the stavke of a zaduženje gradilišta (the legacy
// TABLE_rpro of rob_ZadauzGRADILISTA: redni broj, konto, šifra, naziv, JM, količina, cena, vrednost).
func (s *RobnoDokumentaResource) GetOpstiGrdStavkeTableFields() []domain.Fields {
	return s.opstiGrdStavkeTableFields
}

// GetOpstiOtrStavkeTableFields returns the columns of the grid of the stavke of a trebovanje (the legacy TABLE_rpro
// of rob_Trebovanja: the ones of the opšti dokument with the vrednost).
func (s *RobnoDokumentaResource) GetOpstiOtrStavkeTableFields() []domain.Fields {
	return s.opstiOtrStavkeTableFields
}

// GetAvansniStavkeTableFields returns the columns of the grid of the stavke of the avansni računi.
func (s *RobnoDokumentaResource) GetAvansniStavkeTableFields() []domain.Fields {
	return s.avansniStavkeTableFields
}

// GetUslugeStavkeTableFields returns the columns of the grid of the stavke of the fakture usluga.
func (s *RobnoDokumentaResource) GetUslugeStavkeTableFields() []domain.Fields {
	return s.uslugeStavkeTableFields
}

// GetFaktureAvansiTableFields returns the columns of the grid of the avansi of the "Fakture
// veleprodaje" screen.
func (s *RobnoDokumentaResource) GetFaktureAvansiTableFields() []domain.Fields {
	return s.faktureAvansiTableFields
}

// GetFaktureDokumentaTableFields returns the columns of the grid of the documents of the nalog of the
// "Prikaz faktura" tab of the "Fakture veleprodaje" screen.
func (s *RobnoDokumentaResource) GetFaktureDokumentaTableFields() []domain.Fields {
	return s.faktureDokumentaTableFields
}

// GetFaktureDokumentStavkeTableFields returns the columns of the grid of the stavke of the selected
// document of the "Prikaz faktura" tab.
func (s *RobnoDokumentaResource) GetFaktureDokumentStavkeTableFields() []domain.Fields {
	return s.faktureDokumentStavkeTableFields
}

// GetInterniStavkeTableFields returns the columns of the grid of the stavke of the "Interni prenosi -
// dostavnice" screen.
func (s *RobnoDokumentaResource) GetInterniStavkeTableFields() []domain.Fields {
	return s.interniStavkeTableFields
}

// GetInterniDokumentaTableFields returns the columns of the grid of the documents of the "Pregled internih
// prenosa" tab.
func (s *RobnoDokumentaResource) GetInterniDokumentaTableFields() []domain.Fields {
	return s.interniDokumentaTableFields
}

// GetInterniDokumentStavkeTableFields returns the columns of the grid of the stavke (DETALJI) of the
// selected document of the "Pregled internih prenosa" tab.
func (s *RobnoDokumentaResource) GetInterniDokumentStavkeTableFields() []domain.Fields {
	return s.interniDokumentStavkeTableFields
}

// GetPorStavkeTableFields returns the columns of the grid of the stavke of the "Porudžbenice" screen (also of
// the stavke of the selected document of the "Pregled porudžbenica" tab).
func (s *RobnoDokumentaResource) GetPorStavkeTableFields() []domain.Fields {
	return s.porStavkeTableFields
}

// setTableFields defines the grid columns of every tab of the "Robna dokumenta" option.
func (s *RobnoDokumentaResource) setTableFields() {
	// Tab 1 - "Unos dokumenta": one row per robni nalog (rnal), the same columns as the legacy grid.
	s.unosDokumentaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "8", Field: "rnal.tipdok", TextAlign: "center", Sortable: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", Field: "rnal.nalog", TextAlign: "right", Sortable: true},
		{Name: "mag", Label: "Magacin", Width: "7", Field: "rnal.mag", TextAlign: "center"},
		{Name: "danal", Label: "Datum naloga", Width: "9", Field: "rnal.danal", TextAlign: "center", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "22", Field: "rnal.opis", Sortable: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", Field: "rnal.datob", TextAlign: "center", Sortable: true},
		{Name: "brdo", Label: "Uk. broj dokumenata", Width: "10", Field: "rnal.brdo", TextAlign: "right", SkipInSearch: true},
		{Name: "brst", Label: "Br. stavki", Width: "7", Field: "rnal.brst", TextAlign: "right", SkipInSearch: true},
		{Name: "dug", Label: "Duguje", Width: "10", Field: "rnal.dug", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "pot", Label: "Potražuje", Width: "10", Field: "rnal.pot", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}

	// Tab 2 - "Pregled dokumenta": the grid is the same for the tab and for its first sub-tab
	// ("Štampa"), the eFaktura sub-tab has its own columns.
	s.pregledStampaTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj naloga", Width: "7", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", TextAlign: "center"},
		{Name: "datob", Label: "Datum obrade", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dop", Label: "Datum dospeća", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "datiz", Label: "Datum izv. dokumenta", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "9"},
		{Name: "iznos", Label: "Iznos", Width: "9", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "fkto", Label: "Konto", Width: "6"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "naziv", Label: "Naziv", Width: "20"},
		{Name: "pib", Label: "PIB", Width: "8"},
		{Name: "jbkjs", Label: "JBKJS", Width: "7", SkipInSearch: true},
		{Name: "adresa", Label: "Adresa", Width: "16", SkipInSearch: true},
		{Name: "mesto", Label: "Mesto", Width: "10"},
		{Name: "avansi", Label: "Avansi", Width: "12", SkipInSearch: true},
		{Name: "pornapomena", Label: "Poreska napomena", Width: "14", SkipInSearch: true},
		{Name: "pkto", Label: "Konto 2", Width: "6", SkipInSearch: true},
		{Name: "pana", Label: "Šifra 2", Width: "6", SkipInSearch: true},
		{Name: "naziv1", Label: "Naziv 2", Width: "16", SkipInSearch: true},
		{Name: "valuta", Label: "Valuta", Width: "8", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "tkonto", Label: "Konto prodavnice", Width: "8", SkipInSearch: true},
	}
	s.pregledDokumentaTableFields = s.pregledStampaTableFields

	s.pregledEFakturaTableFields = []domain.Fields{
		{Name: "statussalinv", Label: "Status", Width: "9"},
		{Name: "datumstatsalinv", Label: "Datum statusa", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "komentar", Label: "Komentar", Width: "12", SkipInSearch: true},
		{Name: "cirinvoiceid", Label: "CIR", Width: "9"},
		{Name: "vatrecordingstatus", Label: "Status IndVAT", Width: "9", SkipInSearch: true},
		{Name: "datumstatindvat", Label: "Datum statusa PE", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "magaciniid", Label: "Magacin", Width: "6", TextAlign: "center", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "7", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8", TextAlign: "center"},
		{Name: "fkto", Label: "Konto", Width: "6"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "naziv", Label: "Naziv", Width: "20"},
		{Name: "iznos", Label: "Iznos", Width: "9", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "dop", Label: "Datum dospeća", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "9"},
		{Name: "valuta", Label: "Valuta", Width: "8", SkipInSearch: true},
		{Name: "kurs", Label: "Kurs", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "rdokid", Label: "RDOKID", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "salesinvoiceid", Label: "Sales", Width: "10", TextAlign: "right", SkipInSearch: true},
	}

	// Tab 3 - "Specifikacije dokumenta": one row per stavka (rpro) of the robni dokumenti (rdok) of
	// the selection, with the columns of the legacy report RobSpecifikacijaFakture ("SPECIFIKACIJA
	// EKSTERNIH RACUNA") - the header of the document, its partner and the amounts of the stavka.
	s.specifikacijeDokumentaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "tipdok", Label: "Vrsta naloga", Width: "5", TextAlign: "center"},
		{Name: "nalog", Label: "Broj naloga", Width: "5", TextAlign: "right"},
		{Name: "danal", Label: "Datum naloga", Width: "7", TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "5", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokum.", Width: "5", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "7", TextAlign: "center"},
		{Name: "rok", Label: "Rok", Width: "4", TextAlign: "center", SkipInSearch: true},
		{Name: "valuta", Label: "Valuta", Width: "5", SkipInSearch: true},
		{Name: "fkto", Label: "Kupac", Width: "5"},
		{Name: "fana", Label: "Šifra", Width: "5"},
		{Name: "naziv", Label: "Naziv kupca", Width: "16"},
		{Name: "iznos", Label: "Iznos", Width: "8", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "rabat", Label: "Rabat", Width: "7", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "porez", Label: "Porez", Width: "7", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "zanaplatu", Label: "Za naplatu", Width: "9", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}

	// Tab 4 - "Kontiranje dokumenata": one row per robni dokument (rdok) with the data the legacy
	// screen shows before the posting, with the vrsta naloga and the datum naloga of the document.
	s.kontiranjeKnjizenjeTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "8", TextAlign: "center", Sortable: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", TextAlign: "right", Sortable: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos dokumenta", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "20"},
		{Name: "polje", Label: "POPDV polje", Width: "8", SkipInSearch: true},
		{Name: "rok", Label: "Rok plaćanja", Width: "7", TextAlign: "center", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", SkipInSearch: true},
	}

	// The two "Pregled ..." sub-tabs of the tab show the datum naloga next to the broj naloga and
	// the vrsta dokumenta; the "po magacinima" one adds the vrsta naloga (its rows are grouped by
	// the magacin of the document).
	s.kontiranjePregledTableFields = []domain.Fields{
		{Name: "nalog", Label: "Broj naloga", Width: "8", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos dokumenta", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "20"},
		{Name: "polje", Label: "POPDV polje", Width: "8", SkipInSearch: true},
		{Name: "rok", Label: "Rok plaćanja", Width: "7", TextAlign: "center", SkipInSearch: true},
	}
	s.kontiranjePoMagacinimaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "8", TextAlign: "center", Sortable: true},
		{Name: "nalog", Label: "Broj naloga", Width: "8", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center"},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos dokumenta", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "opis", Label: "Opis", Width: "20"},
		{Name: "polje", Label: "POPDV polje", Width: "8", SkipInSearch: true},
		{Name: "rok", Label: "Rok plaćanja", Width: "7", TextAlign: "center", SkipInSearch: true},
	}

	// Tab 5 - "Prepis dokumenta"
	// TODO: columns of the document rewrite (prepis).
	s.prepisDokumentaTableFields = []domain.Fields{}

	// Tab 6 - "Prikaz ukupne obrade": one row per magacin of the current period with the totals of
	// its robni nalozi (rnal), like the legacy screen shows them.
	s.prikazUkupneObradeTableFields = []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8", TextAlign: "center", Sortable: true},
		{Name: "opis", Label: "Opis magacina", Width: "20", Sortable: true},
		{Name: "mesto", Label: "Mesto", Width: "12", Sortable: true},
		{Name: "brojnaloga", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true, IncludeInTotals: true},
		{Name: "ukdokumenata", Label: "Uk. broj dokumenata", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "brstavki", Label: "Br. stavki", Width: "8", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "duguje", Label: "Duguje", Width: "12", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "potrazuje", Label: "Potražuje", Width: "12", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}

	// Tab 7 - "Prikaz naloga": one row per robni nalog (rnal) with the totals it was saved with.
	s.prikazNalogaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "detalji", Label: "Detalji", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "brdo", Label: "Broj dokumenata", Width: "9", TextAlign: "right", SkipInSearch: true},
		{Name: "brst", Label: "Broj stavki", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "dug", Label: "Duguje", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "pot", Label: "Potražuje", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "10"},
	}

	// The printed "Prikaz naloga" report has no "Detalji" column (the expand arrow is a screen-only
	// affordance): the same columns without it.
	s.prikazNalogaPrintTableFields = make([]domain.Fields, 0, len(s.prikazNalogaTableFields))
	for _, field := range s.prikazNalogaTableFields {
		if field.Name == "detalji" {
			continue
		}
		s.prikazNalogaPrintTableFields = append(s.prikazNalogaPrintTableFields, field)
	}

	// Tab 8 - "Prikaz dokumenata u nalogu": one row per robni dokument (rdok) of the selected
	// nalozi, with the data of the document and of its nalog (the columns of the legacy screen).
	s.prikazDokumenataUNaloguTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "9", TextAlign: "right", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "9", TextAlign: "center", Sortable: true},
		{Name: "vrd", Label: "Vrsta dokumenta", Width: "8", TextAlign: "center"},
		{Name: "dokum", Label: "Dokument", Width: "9", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9", TextAlign: "center"},
		{Name: "brst", Label: "Broj stavki", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "datob", Label: "Datum obrade", Width: "9", TextAlign: "center", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "10"},
	}

	// Tab 9 - "Prikaz dokumenata po operateru": one row per operater (the operater of a document is
	// its xopizmene when it was changed, else its xopunos) with the number of his robni dokumenti and
	// of their stavke and the sums of duguje/potražuje of his documents (the legacy procedure groups
	// the documents by the operater and builds this table).
	s.prikazDokumenataPooperateruTableFields = []domain.Fields{
		{Name: "oper", Label: "Operater", Width: "16"},
		{Name: "nbrd", Label: "Broj dokumenata", Width: "10", TextAlign: "right"},
		{Name: "nbrst", Label: "Broj stavki", Width: "10", TextAlign: "right"},
		{Name: "xdug", Label: "Duguje", Width: "12", TextAlign: "right"},
		{Name: "xpot", Label: "Potražuje", Width: "12", TextAlign: "right"},
	}

	// "Fakture veleprodaje" screen: the grid of the stavke of the faktura and the grid of the avansi
	// that can be closed with it.
	s.faktureStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "24"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "magacinskacena", Label: "Magacinska cena", Width: "9", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", IncludeInTotals: true},
		{Name: "prodajnacena", Label: "Prodajna cena", Width: "9", TextAlign: "right"},
		{Name: "rabat", Label: "Rabat", Width: "6", TextAlign: "right"},
	}
	// The columns of the legacy TABLE_rpro of rob_opsti_dokumenti_FIN.
	s.opstiFinStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "40"},
		{Name: "jm", Label: "Jedinica mere", Width: "8", TextAlign: "center"},
		{Name: "iznos", Label: "Iznos", Width: "12", TextAlign: "right", IncludeInTotals: true},
	}
	// The columns of the legacy TABLE_rpro of rob_opsti_dokumenti_KOL (with the lot of the stavka).
	s.opstiKolStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "32"},
		{Name: "jm", Label: "Jedinica mere", Width: "8", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "10", TextAlign: "right"},
		{Name: "otk", Label: "OTK broj", Width: "8"},
		{Name: "serija", Label: "Serija", Width: "8"},
		{Name: "rok", Label: "Rok", Width: "8", TextAlign: "center"},
	}
	// The columns of the legacy TABLE_rpro of rob_opsti_dokumenti_MNV.
	s.opstiMnvStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "40"},
		{Name: "kolicina", Label: "Količina", Width: "10", TextAlign: "right"},
		{Name: "staracena", Label: "Stara cena", Width: "10", TextAlign: "right"},
		{Name: "novacena", Label: "Nova cena", Width: "10", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "12", TextAlign: "right", IncludeInTotals: true},
	}
	// The columns of the legacy TABLE_RPRO of rob_Nivelacije_Veleprodaja (the last three of the lot).
	s.opstiNivStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "7", SkipInSearch: true},
		{Name: "naziv", Label: "Opis artikla", Width: "26"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "staracena", Label: "Stara cena", Width: "8", TextAlign: "right"},
		{Name: "procenat", Label: "Procenat", Width: "6", TextAlign: "right"},
		{Name: "novacena", Label: "Nova cena", Width: "8", TextAlign: "right"},
		{Name: "razlika", Label: "Razlika", Width: "9", TextAlign: "right", IncludeInTotals: true},
		{Name: "otk", Label: "OTK", Width: "7"},
		{Name: "serija", Label: "Serija", Width: "7"},
		{Name: "rok", Label: "Rok trajanja", Width: "7", TextAlign: "center"},
	}
	// The columns of the legacy TABLE_rpro of rob_opsti_dokumenti_opd.
	s.opstiStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "26"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "cena", Label: "Cena", Width: "9", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", IncludeInTotals: true},
		{Name: "otk", Label: "OTK", Width: "8"},
		{Name: "serija", Label: "Serija", Width: "8"},
		{Name: "rok", Label: "Rok", Width: "8", TextAlign: "center"},
	}
	s.opstiGrdStavkeTableFields = append([]domain.Fields{}, s.opstiStavkeTableFields[:8]...)
	s.opstiGrdStavkeTableFields[7].Label = "Vrednost"
	s.opstiOtrStavkeTableFields = append([]domain.Fields{}, s.opstiStavkeTableFields...)
	s.opstiOtrStavkeTableFields[7].Label = "Vrednost"
	s.opstiPreStavkeTableFields = append(append(append([]domain.Fields{}, s.opstiStavkeTableFields[:8]...),
		domain.Fields{Name: "nabavnacena", Label: "Nabavna cena", Width: "9", TextAlign: "right"}), s.opstiStavkeTableFields[8:]...)
	// The columns of the legacy TABLE_rpro of rob_FAKTURE_Usluga_DET2 (the stopa is the porez of the tarifa).
	// The columns of the legacy TABLE_rpro of rob_AvansniRacuni_DET.
	s.avansniStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "naziv", Label: "Opis robe ili usluge", Width: "40"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "cena", Label: "Cena", Width: "10", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", IncludeInTotals: true},
		{Name: "po", Label: "Tarifa", Width: "5", TextAlign: "right"},
		{Name: "taxcat", Label: "Por. kat.", Width: "6", TextAlign: "center"},
		{Name: "taxexreco", Label: "Osnov osl.", Width: "8"},
	}
	s.uslugeStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto usluge", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra usluge", Width: "7", SkipInSearch: true},
		{Name: "naziv", Label: "Opis usluge", Width: "26"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "cena", Label: "Cena", Width: "9", TextAlign: "right"},
		{Name: "rabat", Label: "Rabat", Width: "6", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", IncludeInTotals: true},
		{Name: "porez", Label: "Porez", Width: "5", TextAlign: "right"},
		{Name: "po", Label: "PO", Width: "4", TextAlign: "right"},
		{Name: "taxcat", Label: "Por. kat.", Width: "6", TextAlign: "center"},
		{Name: "taxexreco", Label: "Osnov osl.", Width: "8"},
	}
	s.faktureAvansiTableFields = []domain.Fields{
		// The columns of the legacy TABLE_AVANS (COL_TRAZI, COL_BRAVANS, COL_DATAVANS, COL_IZNAVANS,
		// COL_OSTATAK, COL_IZNFAKT; the hidden AVANSID is the key of the row).
		{Name: "trazi", Label: "Traži", Width: "4", TextAlign: "center", SkipInSearch: true},
		{Name: "bravans", Label: "Broj avansa", Width: "10", TextAlign: "right", SkipInSearch: true},
		{Name: "datavans", Label: "Datum avansa", Width: "10", TextAlign: "center", SkipInSearch: true},
		{Name: "iznavans", Label: "Iznos avansa", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "ostatak", Label: "Ostatak avansa", Width: "12", TextAlign: "right", SkipInSearch: true},
		{Name: "iznfakt", Label: "Iznos koji je zatv. na fakt.", Width: "12", TextAlign: "right", SkipInSearch: true},
	}
	// The "Prikaz faktura" tab: the documents of the nalog and the stavke of the selected document,
	// like the two grids of the legacy window.
	s.faktureDokumentaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta naloga", Width: "5", TextAlign: "center", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "danal", Label: "Datum naloga", Width: "7", TextAlign: "center", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "6", TextAlign: "right", Sortable: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "7", TextAlign: "center", Sortable: true},
		{Name: "iznos", Label: "Iznos", Width: "8", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "fkto", Label: "Konto partnera", Width: "5"},
		{Name: "fana", Label: "Šifra", Width: "5"},
		{Name: "naziv", Label: "Naziv partnera", Width: "18"},
		{Name: "opis", Label: "Opis", Width: "10"},
		{Name: "rok", Label: "Rok plaćanja", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "datiz", Label: "Datum izvornog dokum.", Width: "7", TextAlign: "center", SkipInSearch: true},
		{Name: "datob", Label: "Datum obrade", Width: "7", TextAlign: "center", SkipInSearch: true},
		{Name: "mag", Label: "Magacin", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "dokiz", Label: "Izvorni dokument", Width: "7", SkipInSearch: true},
		{Name: "mi", Label: "Mesto isporuke", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "kamt", Label: "Uslovi kamate", Width: "8", SkipInSearch: true},
		{Name: "otp", Label: "Način otpreme", Width: "8", SkipInSearch: true},
		{Name: "kom", Label: "Komercijalista", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "oper", Label: "Operater", Width: "6", SkipInSearch: true},
	}
	s.faktureDokumentStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "10", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "35"},
		{Name: "jm", Label: "JM", Width: "8", TextAlign: "center", SkipInSearch: true},
		{Name: "kolic", Label: "Količina", Width: "9", TextAlign: "right", SkipInSearch: true},
		{Name: "cena", Label: "Cena", Width: "9", TextAlign: "right", SkipInSearch: true},
		{Name: "rab", Label: "Rabat", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}
	// The "Interni prenosi - dostavnice" screen (the legacy rob_Interni_racuni): the stavke of the document
	// and the two grids of the "Pregled internih prenosa" tab, the columns of the legacy grids.
	s.interniStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "R. br.", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "22"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "veleprcena", Label: "Velepr. cena", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "veleprrabat", Label: "Velepr. rabat", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "nabavnacena", Label: "Nabavna cena", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "maloprmarza", Label: "Malopr. marža", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "maloprcena", Label: "Malopr. cena", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "maloprcenapor", Label: "Malopr. cena sa porezom", Width: "9", TextAlign: "right", SkipInSearch: true},
	}
	s.interniDokumentaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "4", TextAlign: "right", SkipInSearch: true},
		{Name: "tipdok", Label: "Vrsta naloga", Width: "5", TextAlign: "center", SkipInSearch: true},
		{Name: "nalog", Label: "Broj naloga", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "dokum", Label: "Broj dokumenta", Width: "6", TextAlign: "right", Sortable: true},
		{Name: "dadok", Label: "Datum dokumenta", Width: "7", TextAlign: "center", Sortable: true},
		{Name: "iznos", Label: "Iznos", Width: "8", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
		{Name: "fkto", Label: "Konto dobavljača", Width: "6"},
		{Name: "fana", Label: "Analitička šifra", Width: "6"},
		{Name: "naziv", Label: "Naziv partnera", Width: "20"},
		{Name: "opis", Label: "Opis", Width: "12"},
	}
	// The "Porudžbenice" screen (the legacy rob_Porudzbenica): the columns of the legacy grid of the stavke.
	s.porStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Red. broj", Width: "5", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "8", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "35"},
		{Name: "jm", Label: "JM", Width: "6", TextAlign: "center", SkipInSearch: true},
		{Name: "kolic", Label: "Količina", Width: "9", TextAlign: "right", SkipInSearch: true},
		{Name: "ncena", Label: "Nabavna cena", Width: "10", TextAlign: "right", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", SkipInSearch: true, IncludeInTotals: true},
	}
	s.interniDokumentStavkeTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8"},
		{Name: "naziv", Label: "Naziv artikla", Width: "24"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center", SkipInSearch: true},
		{Name: "kolic", Label: "Količina", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "planskacena", Label: "Planska cena", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "maloprcena", Label: "Malopr. cena", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "veleprcena", Label: "Velepr. cena", Width: "7", TextAlign: "right", SkipInSearch: true},
		{Name: "veleprrabat", Label: "Velepr. rabat", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "maloprmarza", Label: "Malopr. marža", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "maloprcenapor", Label: "Malopr. cena sa porezom", Width: "8", TextAlign: "right", SkipInSearch: true},
	}
}

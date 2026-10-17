package domain

import "database/sql"

// The "eOtpremnice" screen of the komercijalno (the legacy pr_eOTPREMNICA with its internal window IW_Otpremnice): the
// otpremnica of the "Otpremnice" screen (the same edok and epro, the same vrste) with the data of the e-otpremnica in
// collapsible controls: the osnovni podaci with the pošiljalac and the primalac (DET1), the način otpreme with the
// prevoznik, the vozač and the faze vožnje (DET2), the adresa otpreme and the adresa odredišta (DET3), the stavke (DET4)
// and the opasne materije, the roba treće strane and the napomena (DET5). The columns are of migrations/003.

// The tipovi of an e-otpremnica (edok.tipeotp, the legacy COMBO_TIPEOTP).
const (
	KomEotpEksterna = "Ext"
	KomEotpInterna  = "Int"
)

// The načini otpreme (edok.nacotp, the legacy COMBO_NACINOTP): 1-3 a prevoz (the prevoznik is the firm, a prevoznik of
// the kontni plan or the primalac), 4-5 lično (the ime, the prezime and the lična karta of the person).
const (
	KomEotpSopstveniPrevoz  = 1
	KomEotpPrevoznik        = 2
	KomEotpPrevozPrimaoca   = 3
	KomEotpLicnoPreuzimanje = 4
	KomEotpLicnaDostava     = 5
)

// KomEotpPodaci are the data of the e-otpremnica in edok (the ones the otpremnica already has, the vozač, the registarski
// broj, the broj ugovora, the bruto težina, the broj paketa and the mesto isporuke, stay in KomOtpremnicaRow).
type KomEotpPodaci struct {
	OpasneMat   bool
	Roba3Str    bool
	TipEotp     string
	Jm1         string
	Jm2         string
	Bttovol     float64
	Datpoc      sql.NullTime
	Datzav      sql.NullTime
	StvDatotp   sql.NullTime
	PovratRobe  bool
	BrNarudz    string
	BrOkvSpor   string
	KontIme     string
	KontTel     string
	KontMail    string
	ReferKup    string
	TelKup      string
	MailKup     string
	NacOtp      int
	LicIme      string
	LicPrez     string
	LicBrLk     string
	MestoUtov   string
	MestoIsp    string
	PrevKto     string
	PrevAna     string
	VozPrez     string
	VozMail     string
	VozTel      string
	VozDozv     string
	UlicaOtp    string
	UlBrOtp     string
	MestoOtp    string
	PosBrOtp    string
	SifObjOtp   string
	UlicaOdr    string
	UlBrOdr     string
	MestoOdr    string
	PosBrOdr    string
	SifObjOdr   string
	FisNap      string
	Roba3StrTxt string
}

// KomEotpFaza is a faza vožnje of an otpremnica (enot, the legacy grid TABLE_PR_QRY_ENOT): its način otpreme and the
// naziv of the prevoznik or the ime i prezime.
type KomEotpFaza struct {
	Rbr    int    `json:"rbr"`
	NacOtp int    `json:"nacotp"`
	Naziv  string `json:"naziv"`
}

// KomEotpOpasna is an opasna materija of an otpremnica (eopma, the legacy grid TABLE_PR_QRY_EOPMA).
type KomEotpOpasna struct {
	Rbr      int    `json:"rbr"`
	Naziv    string `json:"naziv"`
	Opis     string `json:"opis"`
	Komentar string `json:"komentar"`
}

// KomEotpPartner is what the e-otpremnica shows of a partner (the pošiljalac, the primalac, the prevoznik; the legacy
// PrikaziKupca, PrikaziPrevoznika): the naziv, the adresa, the mesto (with the poštanski broj), the PIB, the matični broj
// and the JBKJS.
type KomEotpPartner struct {
	Naziv  string `json:"naziv"`
	Adresa string `json:"adresa"`
	Mesto  string `json:"mesto"`
	Pib    string `json:"pib"`
	Matbr  string `json:"matbr"`
	Jbkjs  string `json:"jbkjs"`
}

// KomEotpAdresa is an adresa of the e-otpremnica: of the magacin (the adresa otpreme) or of a mesto isporuke (the adresa
// odredišta, the legacy PrikaziMI).
type KomEotpAdresa struct {
	Ulica string `json:"ulica"`
	Mesto string `json:"mesto"`
	Pobro string `json:"pobro"`
}

// KomEotpView is the e-otpremnica of the entry dialog: its data, its faze vožnje and opasne materije, the pošiljalac (the
// firm), the primalac (the kupac of the otpremnica) and the prevoznik of the način otpreme 2.
type KomEotpView struct {
	KomEotpPodaci
	Faze       []KomEotpFaza
	Opasne     []KomEotpOpasna
	Posiljalac KomEotpPartner
	Primalac   KomEotpPartner
	Prevoznik  KomEotpPartner
}

// KomEotpParams are the fields of the controls of the e-otpremnica the "Sačuvaj" of the header sends with it (the
// vozač, the registarski broj, the broj ugovora, the bruto težina, the broj paketa and the mesto isporuke are fields of
// KomOtpParams). The faze vožnje and the opasne materije are the rows of their grids, every column a list.
type KomEotpParams struct {
	OpasneMat   string   `form:"opasnemat"`
	Roba3Str    string   `form:"roba3str"`
	TipEotp     string   `form:"tipeotp"`
	Jm1         string   `form:"jm1"`
	Jm2         string   `form:"jm2"`
	Bttovol     string   `form:"bttovol"`
	Datpoc      string   `form:"datpoc"`
	Datzav      string   `form:"datzav"`
	StvDatotp   string   `form:"stvdatotp"`
	PovratRobe  string   `form:"povratrobe"`
	BrNarudz    string   `form:"brnarudz"`
	BrOkvSpor   string   `form:"brokvspor"`
	KontIme     string   `form:"kontime"`
	KontTel     string   `form:"konttel"`
	KontMail    string   `form:"kontmail"`
	ReferKup    string   `form:"referkup"`
	TelKup      string   `form:"telkup"`
	MailKup     string   `form:"mailkup"`
	NacOtp      string   `form:"nacotp"`
	LicIme      string   `form:"licime"`
	LicPrez     string   `form:"licprez"`
	LicBrLk     string   `form:"licbrlk"`
	MestoUtov   string   `form:"mestoutov"`
	MestoIsp    string   `form:"mestoisp"`
	PrevKto     string   `form:"prevkto"`
	PrevAna     string   `form:"prevana"`
	VozPrez     string   `form:"vozprez"`
	VozMail     string   `form:"vozmail"`
	VozTel      string   `form:"voztel"`
	VozDozv     string   `form:"vozdozv"`
	UlicaOtp    string   `form:"ulicaotp"`
	UlBrOtp     string   `form:"ulbrotp"`
	MestoOtp    string   `form:"mestootp"`
	PosBrOtp    string   `form:"posbrotp"`
	SifObjOtp   string   `form:"sifobjotp"`
	Mi          string   `form:"eotpmi"`
	UlicaOdr    string   `form:"ulicaodr"`
	UlBrOdr     string   `form:"ulbrodr"`
	MestoOdr    string   `form:"mestoodr"`
	PosBrOdr    string   `form:"posbrodr"`
	SifObjOdr   string   `form:"sifobjodr"`
	FisNap      string   `form:"fisnap"`
	Roba3StrTxt string   `form:"roba3strtxt"`
	FazaNacOtp  []string `form:"fazanacotp"`
	FazaNaziv   []string `form:"fazanaziv"`
	OpmaNaziv   []string `form:"opmanaziv"`
	OpmaOpis    []string `form:"opmaopis"`
	OpmaKoment  []string `form:"opmakomentar"`
}

// KomEotpUpis is what the save of the header writes of the e-otpremnica: its data, its mesto isporuke, its faze vožnje
// and its opasne materije.
type KomEotpUpis struct {
	KomEotpPodaci
	Mi     int
	Faze   []KomEotpFaza
	Opasne []KomEotpOpasna
}

// KomEotpNalogPodaci is what a new e-otpremnica takes of its nalog za isporuku when it is chosen: the primalac (its
// kupac) and the adresa odredišta of its mesto isporuke.
type KomEotpNalogPodaci struct {
	Primalac  KomEotpPartner `json:"primalac"`
	Mi        int            `json:"mi"`
	Odrediste KomEotpAdresa  `json:"odrediste"`
	Fkto      string         `json:"fkto"`
	Fana      string         `json:"fana"`
}

// KomEotpUI are the urls of the JSON of the controls of the e-otpremnica and the ids of the controls (DET1, DET2, DET3
// and DET5; the stavke are the control of the otpremnica).
type KomEotpUI struct {
	NalogURL    string
	PartnerURL  string
	MestoURL    string
	Det1PanelID string
	Det2PanelID string
	Det3PanelID string
	Det5PanelID string
	// Koraci: the controls as the steps of a stepper (the test screen "Unos u koracima") instead of collapsible
	// controls.
	Koraci bool
}

package domain

import (
	"database/sql"
	"time"
)

type Fvr struct {
	God     int    `json:"god" db:"god"`
	Kar     int    `json:"kar" db:"kar"`
	Naziv   string `json:"naziv" db:"naziv"`
	Adresa  string `json:"adresa" db:"adresa"`
	GodZatv bool   `json:"god_zatv" db:"godzatv"`
}

type Firma struct {
	Firme []FvrFirma `json:"firme"`
}
type FvrFirma struct {
	IDFirma int      `json:"idfirma" db:"idfvr"`
	Godine  []Godina `json:"god" db:"god"`
	Naziv   string   `json:"naziv" db:"naziv"`
	Adresa  string   `json:"adresa" db:"adresa"`
}
type Godina struct {
	God int   `json:"god" db:"god"`
	Kar []int `json:"kar" db:"kar"`
}

// Drzave Model
type Drzave struct {
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	IDDrzave   int            `json:"id_drzave" db:"iddrzave"`
	Naziv      string         `json:"naziv" db:"naziv" form:"naziv"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpuNos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	OznDrz     string         `json:"ozn_drz" db:"ozndrz" form:"ozndrz"`
}

// Banke Model
type Banke struct {
	IDBanKe    int            `json:"idbanke" db:"idbanke"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	BrRac      string         `json:"brrac" db:"brrac"`
	Banka      string         `json:"banka" db:"banka"`
	Konto      string         `json:"konto" db:"konto"`
	Sifra      string         `json:"sifra" db:"sifra"`
	BnkCod     string         `json:"bnkcod" db:"bnkcod"`
	EBank      string         `json:"ebank" db:"ebank"`
	PocNazFajl string         `json:"pocnazfajl" db:"pocnazfajl"`
	TipDok     string         `json:"tipdok" db:"tipdok"`
	NaFakNe    bool           `json:"nafakne" db:"nafakne"`
	XDatUnosa  sql.NullTime   `json:"xdatunosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	XOpuNos    string         `json:"xopunos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
}

// Sifop Model
type Sifop struct {
	IDSifop    int            `json:"id_sifop" db:"idsifop"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Ops        string         `json:"ops" db:"ops"`
	Naziv      string         `json:"naziv" db:"naziv"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpuNos    sql.NullString `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
}

// Partneri Model
type Partneri struct {
	IDPartneri  int            `json:"id_partneri" db:"idpartneri"`
	God         int            `json:"god" db:"god"`
	Kar         int            `json:"kar" db:"kar"`
	Naziv       string         `json:"naziv" db:"naziv" form:"naziv"`
	Mesto       string         `json:"mesto" db:"mesto" form:"mesto"`
	PoBro       int            `json:"pobro" db:"pobro" form:"pobro"`
	Adresa      string         `json:"adresa" db:"adresa" form:"adresa"`
	Ziro        string         `json:"ziro" db:"ziro"`
	TipPDV      int            `json:"tip_pdv" db:"tippdv" form:"tippdv"`
	Mar         float32        `json:"mar" db:"mar"`
	Ter         int            `json:"ter" db:"ter" form:"ter"`
	PIB         string         `json:"pib" db:"pib" form:"pib"`
	Email       string         `json:"email" db:"email" form:"email"`
	Website     string         `json:"website" db:"website" form:"website"`
	Telefon     string         `json:"telefon" db:"telefon" form:"telefon"`
	StariIDPart int64          `json:"stari_id_part" db:"stariidpart" form:"stariidpart"`
	KontaktOsb  string         `json:"kontakt_osb" db:"kontaktosb" form:"kontaktosb"`
	Konta       string         `json:"konta" db:"konta" form:"konta"`
	MatBr       string         `json:"matbr" db:"matbr" form:"matbr"`
	PartnFlg    int            `json:"partn_flg" db:"partnflg" form:"partnflg"`
	ProVal      int            `json:"pro_val" db:"proval" form:"proval"`
	ProKam      float64        `json:"pro_kam" db:"prokam" form:"prokam"`
	NabVal      int            `json:"nab_val" db:"nabval" form:"nabval"`
	XDatUnosa   sql.NullTime   `json:"xdat_unosa" db:"xdatunosa" `
	XDatIzmene  sql.NullTime   `json:"xdat_izmene" db:"xdatizmene" `
	XOpuNos     string         `json:"xop_unos" db:"xopunos" `
	XOpIzmene   sql.NullString `json:"xop_izmene" db:"xopizmene"`
	IDSifOp     int            `json:"id_sifop" db:"idsifop"`
	KrLimit     float64        `json:"kr_limit" db:"krlimit" form:"krlimit"`
	KrLimitDosp float64        `json:"kr_limit_dosp" db:"krlimitdosp" form:"krlimitdosp"`
	Sifra       string         `json:"sifra" db:"sifra" form:"sifra"`
	JMBG        string         `json:"jmbg" db:"jmbg" form:"jmbg"`
	JIB         string         `json:"jib" db:"jib" form:"jib"`
	UGRabat     float32        `json:"ug_rabat" db:"ugrabat" form:"ugrabat"`
	Naziv1      string         `json:"naziv1" db:"naziv1" form:"naziv1"`
	BPG         string         `json:"bpg" db:"bpg" form:"bpg"`
	Index       string         `json:"index" db:"index" form:"index"`
	Kom         int            `json:"kom" db:"kom" form:"kom"`
	KomID       int            `json:"kom_id" db:"komid"`
	Napomena    string         `json:"napomena" db:"napomena" form:"napomena"`
	Pak         int64          `json:"pak" db:"pak" form:"pak"`
	Budzetski   bool           `json:"budzetski" db:"budzetski" form:"budzetski"`
	JBKJS       string         `json:"jbkjs" db:"jbkjs" form:"jbkjs"`
	GLN         int64          `json:"gln" db:"gln" form:"gln"`
}

// TekRacuni Model - Current accounts for partners
type TekRacuni struct {
	IDPartneri  int            `json:"id_partneri" db:"idpartneri" addupdate:"true"`
	TekRacuniID int            `json:"tek_racuni_id" db:"tekracuniid" addupdate:"true"`
	God         int            `json:"god" db:"god" addupdate:"true"`
	Kar         int            `json:"kar" db:"kar" addupdate:"true"`
	BnkCod      string         `json:"bnk_cod" db:"bnkcod" addupdate:"true"`
	TekRac      string         `json:"tek_rac" db:"tekrac" addupdate:"true"`
	XOpuNos     string         `json:"xop_unos" db:"xopunos" addupdate:"true"`
	XDatUnosa   sql.NullTime   `json:"xdat_unosa" db:"xdatunosa" addupdate:"true"`
	XOpIzmene   sql.NullString `json:"xop_izmene" db:"xopizmene" addupdate:"true"`
	XDatIzmene  sql.NullTime   `json:"xdat_izmene" db:"xdatizmene" addupdate:"true"`
	IDBanke     int            `json:"id_banke" db:"idbanke" addupdate:"true"`
	BankeDes    sql.NullString `json:"banka" db:"banka" addupdate:"false"`
}

// Tipdok Model
type Tipdok struct {
	IDTipDok   int            `json:"id_tip_dok" db:"idtipdok"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	TipDok     string         `json:"tip_dok" db:"tipdok" form:"tipdok"`
	Opis       string         `json:"opis" db:"opis" form:"opis"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpuNos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	GrpDok     string         `json:"grp_dok" db:"grpdok" form:"grpdok"`
	GrpVrd     string         `json:"grp_vrd" db:"grpvrd" form:"grpvrd"`
	Magacin    string         `json:"magacin" db:"magacin" form:"magacin"`
}

// Dokvrsta Model
type Dokvrsta struct {
	DokVrstaID int            `json:"dok_vrsta_id" db:"dokvrstaid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Opis       string         `json:"opis" db:"opis" form:"opis"`
	Predznak   string         `json:"predznak" db:"predznak" form:"predznak"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpuNos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	Vrd        int            `json:"vrd" db:"vrd" form:"vrd"`
	KodKnj     string         `json:"kod_knj" db:"kodknj" form:"kodknj"`
	GrpDok     string         `json:"grp_dok" db:"grpdok" form:"grpdok"`
	StornoVrd  int            `json:"storno_vrd" db:"stornovrd" form:"stornovrd"`
	Modul      string         `json:"modul" db:"modul" form:"modul"`
	DokOzn     string         `json:"dok_ozn" db:"dokozn" form:"dokozn"`
	FlgKepu    bool           `json:"flg_kepu" db:"flgkepu"`
	GrpNal     string         `json:"grp_nal" db:"grpnal" form:"grpnal"`
	DodOznFak  string         `json:"dod_ozn_fak" db:"dodoznfak"`
	DodOznKon  string         `json:"dod_ozn_kon" db:"dodoznkon"`
	Polje      string         `json:"polje" db:"polje"`
	FlgPopDV   bool           `json:"flg_pop_dv" db:"flgpopdv"`
}

// Popdv Model
type Popdv struct {
	PopdvID        int            `json:"popdv_id" db:"popdvid"`
	God            int            `json:"god" db:"god"`
	Kar            int            `json:"kar" db:"kar"`
	Tip            string         `json:"tip" db:"tip"`
	Deo            int            `json:"deo" db:"deo"`
	Polje          string         `json:"polje" db:"polje"`
	Opis           string         `json:"opis" db:"opis"`
	Naknada        float64        `json:"naknada" db:"naknada"`
	Osn1           float64        `json:"osn1" db:"osn1"`
	Pdv1           float64        `json:"pdv1" db:"pdv1"`
	Osn2           float64        `json:"osn2" db:"osn2"`
	Pdv2           float64        `json:"pdv2" db:"pdv2"`
	Iznos          float64        `json:"iznos" db:"iznos"`
	Poljvred       float64        `json:"poljvred" db:"poljvred"`
	Poljpdv        float64        `json:"poljpdv" db:"poljpdv"`
	Nipo           int            `json:"nipo" db:"nipo"`
	PppdvPolje     int            `json:"pppdv_polje" db:"pppdvpolje"`
	Pozic          string         `json:"pozic" db:"pozic"`
	XDatUnosa      sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene     sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpuNos        sql.NullString `json:"xop_unos" db:"xopunos"`
	XOpIzmene      sql.NullString `json:"xop_izmene" db:"xopizmene"`
	Npredznak      int            `json:"npredznak" db:"npredznak"`
	Aktosn1        bool           `json:"aktosn1" db:"aktosn1"`
	Aktpdv1        bool           `json:"aktpdv1" db:"aktpdv1"`
	Aktosn2        bool           `json:"aktosn2" db:"aktosn2"`
	Aktpdv2        bool           `json:"aktpdv2" db:"aktpdv2"`
	Oddat          time.Time      `json:"oddat" db:"oddat"`
	Dodat          time.Time      `json:"dodat" db:"dodat"`
	FvepdvID       int            `json:"fvepdv_id" db:"fvepdvid"`
	VkrbrFvknjrac  string         `json:"vkrbr_fvknjrac" db:"vkrbr_fvknjrac"`
	TerPartneri    string         `json:"ter_partneri" db:"terpartneri"`
	TippdvPartneri string         `json:"tip_pdv_partneri" db:"tippdvpartneri"`
	Prioritet      int            `json:"prioritet" db:"prioritet"`
	Povpolje1      string         `json:"povpolje1" db:"povpolje1"`
	LabelaAktosn1  string         `json:"labela_aktosn1" db:"labelaaktosn1"`
	LabelaAktpdv1  string         `json:"labela_aktpdv1" db:"labelaaktpdv1"`
	LabelaAktosn2  string         `json:"labela_aktosn2" db:"labelaaktosn2"`
	LabelaAktpdv2  string         `json:"labela_aktpdv2" db:"labelaaktpdv2"`
	Povpolje2      string         `json:"povpolje2" db:"povpolje2"`
	KirkprPolje1O  string         `json:"kirkpr_polje1o" db:"kirkprpolje1o"`
	KirkprPolje1P  string         `json:"kirkpr_polje1p" db:"kirkprpolje1p"`
	KirkprPolje2O  string         `json:"kirkpr_polje2o" db:"kirkprpolje2o"`
	KirkprPolje2P  string         `json:"kirkpr_polje2p" db:"kirkprpolje2p"`
}

// ORGJED Model
type Orgjed struct {
	IDOrgjed   int            `json:"idorgjed" db:"idorgjed"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Naziv      string         `json:"naziv" db:"naziv" form:"naziv"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    sql.NullString `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	OjOzn      string         `json:"ojozn" db:"ojozn" form:"ojozn"`
}

// MESTOTR Model
type Mestotr struct {
	MestoTrID  int            `json:"mestotrid" db:"mestotrid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Mtroska    string         `json:"mtroska" db:"mtroska"`
	Opis       string         `json:"opis" db:"opis"`
	IDOrgjed   int            `json:"idorgjed" db:"idorgjed"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
}

// SIFMESTO Model
type Sifmesto struct {
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	SifM       int            `json:"sifm" db:"sifm"`
	Naziv      string         `json:"naziv" db:"naziv"`
	Ops        string         `json:"ops" db:"ops"`
	Pobro      int            `json:"pobro" db:"pobro"`
	IDDrzave   string         `json:"iddrzave" db:"iddrzave"`
	Km         int            `json:"km" db:"km"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
}

// SIFPLIZV Model
type Sifplizv struct {
	SifPlizvID int            `json:"sifplizvid" db:"sifplizvid"`
	SifPlac    int            `json:"sifplac" db:"sifplac"`
	Oblik      int            `json:"oblik" db:"oblik"`
	Osnov      int            `json:"osnov" db:"osnov"`
	Opis       string         `json:"opis" db:"opis"`
	Konto      string         `json:"konto" db:"konto"`
	Sifra      string         `json:"sifra" db:"sifra"`
	XOpUnos    string         `json:"xop_unos" db:"xopunos"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
}

// FVKNJRAC Model
type Fvknjrac struct {
	IDFvknjrac int            `json:"idfvknjrac" db:"idfvknjrac"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	VkTip      string         `json:"vktip" db:"vktip"`
	VkRbr      int            `json:"vkrbr" db:"vkrbr"`
	Opis       string         `json:"opis" db:"opis"`
	Konta      string         `json:"konta" db:"konta"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    string         `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
}

// VALUTE Model
type Valute struct {
	IDValute   int            `json:"idvalute" db:"idvalute"`
	Sifval     int            `json:"sifval" db:"sifval"`
	Oznval     sql.NullString `json:"oznval" db:"oznval"`
	Naziv      sql.NullString `json:"naziv" db:"naziv"`
	Drzava     sql.NullString `json:"drzava" db:"drzava"`
	Vaziza     int            `json:"vaziza" db:"vaziza"`
	XDatUnosa  sql.NullTime   `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene sql.NullTime   `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos    sql.NullString `json:"xop_unos" db:"xopunos"`
	XOpIzmene  sql.NullString `json:"xop_izmene" db:"xopizmene"`
}

// KOMERCIJALISTI Model
type Komercijalisti struct {
	God          int          `json:"god" db:"god"`
	Kar          int          `json:"kar" db:"kar"`
	KomID        int          `json:"komid" db:"komid"`
	Sifkom       int          `json:"sifkom" db:"sifkom"`
	Sifnadred    int          `json:"sifnadred" db:"sifnadred"`
	Imeprezime   string       `json:"imeprezime" db:"imeprezime"`
	Adresa       string       `json:"adresa" db:"adresa"`
	Mesto        string       `json:"mesto" db:"mesto"`
	Telposao     string       `json:"telposao" db:"telposao"`
	Telmob       string       `json:"telmob" db:"telmob"`
	Totprod      float64      `json:"totprod" db:"totprod"`
	Totprofit    float64      `json:"totprofit" db:"totprofit"`
	Zaddatprod   time.Time    `json:"zaddatprod" db:"zaddatprod"`
	Totnaplaceno float64      `json:"totnaplaceno" db:"totnaplaceno"`
	Loginname    string       `json:"loginname" db:"loginname"`
	XDatUnosa    sql.NullTime `json:"xdat_unosa" db:"xdatunosa"`
	XDatIzmene   sql.NullTime `json:"xdat_izmene" db:"xdatizmene"`
	XOpUnos      string       `json:"xop_unos" db:"xopunos"`
	XOpIzmene    string       `json:"xop_izmene" db:"xopizmene"`
}

type Bnkizv struct {
	Bnkizvid    int            `json:"bnkizvid" db:"bnkizvid"`
	God         int            `json:"god" db:"god"`
	Kar         int            `json:"kar" db:"kar"`
	Sifbank     int            `json:"sifbank" db:"sifbank"`
	Bnkdes      string         `json:"bnkdes" db:"bnkdes"`
	Swiftadr    string         `json:"swiftadr" db:"swiftadr"`
	Brojrac     string         `json:"brojrac" db:"brojrac"`
	Beneficiary string         `json:"beneficiary" db:"beneficiary"`
	Corrbank    string         `json:"corrbank" db:"corrbank"`
	Tel         string         `json:"tel" db:"tel"`
	Fax         string         `json:"fax" db:"fax"`
	Address     string         `json:"address" db:"address"`
	Komentar    string         `json:"komentar" db:"komentar"`
	Xopunos     sql.NullString `json:"xopunos" db:"xopunos"`
	Xdatunosa   time.Time      `json:"xdatunosa" db:"xdatunosa"`
	Xopizmene   sql.NullString `json:"xopizmene" db:"xopizmene"`
	Xdatizmene  sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
}

type Fvepdv struct {
	Fvepdvid   int            `json:"fvepdvid" db:"fvepdvid"`
	God        int            `json:"god" db:"god"`
	Kar        int            `json:"kar" db:"kar"`
	Vktip      string         `json:"vktip" db:"vktip"`
	Vkrbr      int            `json:"vkrbr" db:"vkrbr"`
	Opis       string         `json:"opis" db:"opis"`
	Obrazac    string         `json:"obrazac" db:"obrazac"`
	Xdatunosa  time.Time      `json:"xdatunosa" db:"xdatunosa"`
	Xdatizmene sql.NullTime   `json:"xdatizmene" db:"xdatizmene"`
	Xopunos    string         `json:"xopunos" db:"xopunos"`
	Xopizmene  sql.NullString `json:"xopizmene" db:"xopizmene"`
}

type CompanyInfo struct {
	Naziv  string
	Naziv1 string
	Mesto  string
	PBroj  string
	Adresa string
	MatBr  string
}

type CompanyAccount struct {
	BankCode      string
	AccountNumber string
	ControlNumber string
	BankName      string
	FullAccount   string
}

package domain

import (
	"fmt"
	"time"
)

// tipovi kamate
type Kam struct {
	Idkam      int64      `json:"idkam" db:"idkam"`
	God        int        `json:"god" db:"god"`
	Kar        int        `json:"kar" db:"kar"`
	Tipkam     int        `json:"tipkam" db:"tipkam" form:"tipkam"`
	Opis       *string    `json:"opis,omitempty" db:"opis" form:"opis"`
	Model      *string    `json:"model,omitempty" db:"model" form:"model"`
	Xopunos    *string    `json:"xopunos,omitempty" db:"xopunos"`
	Xopizmene  *string    `json:"xopizmene,omitempty" db:"xopizmene"`
	Xdatunosa  *time.Time `json:"xdatunosa,omitempty" db:"xdatunosa"`
	Xdatizmene *time.Time `json:"xdatizmene,omitempty" db:"xdatizmene"`
}

// kamatne stope
type Tkam struct {
	Idtkam           int64      `json:"idtkam" db:"idtkam"`
	God              int        `json:"god" db:"god"`
	Kar              int        `json:"kar" db:"kar"`
	Tipkam           int        `json:"tipkam" db:"tipkam" form:"tipkam"`
	Odd              *time.Time `json:"odd,omitempty" db:"odd" form:"odd"`
	Dod              *time.Time `json:"dod,omitempty" db:"dod" form:"dod"`
	Kst              float64    `json:"kst" db:"kst" form:"kst"`
	Reast            float64    `json:"reast" db:"reast"`
	Revst            float64    `json:"revst" db:"revst"`
	Xopunos          *string    `json:"xopunos,omitempty" db:"xopunos"`
	Xopizmene        *string    `json:"xopizmene,omitempty" db:"xopizmene"`
	Xdatunosa        *time.Time `json:"xdatunosa,omitempty" db:"xdatunosa"`
	Xdatizmene       *time.Time `json:"xdatizmene,omitempty" db:"xdatizmene"`
	Idkam            *int       `json:"idkam,omitempty" db:"idkam"` // Foreign key to baza.kam
	Opis             *string    `json:"opis,omitempty" db:"opis"`
	Model            *string    `json:"model,omitempty" db:"model"`
	TipKamateOptions []ComboItem
}

type Kali struct {
	Idpartneri *int       `json:"idpartneri,omitempty" db:"idpartneri"`
	Idkali     int64      `json:"idkali" db:"idkali"`
	God        int        `json:"god" db:"god"`
	Kar        int        `json:"kar" db:"kar"`
	Brkaml     int        `json:"brkaml" db:"brkaml"`
	Datlis     *time.Time `json:"datlis,omitempty" db:"datlis"`
	Abr        int        `json:"abr" db:"abr"`
	Rbr        float64    `json:"rbr" db:"rbr"`
	Dug        float64    `json:"dug" db:"dug"`
	Pot        float64    `json:"pot" db:"pot"`
	Konto      *string    `json:"konto,omitempty" db:"konto"`
	Sifra      *string    `json:"sifra,omitempty" db:"sifra"`
	Liststs    *string    `json:"liststs,omitempty" db:"liststs"`
	Xopunos    *string    `json:"xopunos,omitempty" db:"xopunos"`
	Xopizmene  *string    `json:"xopizmene,omitempty" db:"xopizmene"`
	Xdatunosa  *time.Time `json:"xdatunosa,omitempty" db:"xdatunosa"`
	Xdatizmene *time.Time `json:"xdatizmene,omitempty" db:"xdatizmene"`
	Idtkam     *int       `json:"idtkam,omitempty" db:"idtkam"`
}

// Use a temporary DTO with string fields for dates to avoid parsing errors during binding
type TkamDTO struct {
	ID    int64   `form:"id"`
	IDkam int     `form:"idkam"`
	Odd   string  `form:"odd"`
	Dod   string  `form:"dod"`
	Kst   float64 `form:"kst"`
}

type KamateParameters struct {
	Konto           string `form:"konto"`
	OdSifre         string `form:"odsifre"`
	DoSifre         string `form:"dosifre"`
	OdDatuma        string `form:"oddatuma"`
	DoDatuma        string `form:"dodatuma"`
	PrikazOtvorene  bool   `form:"prikaz_otvorene"`
	PrikazZatvorene bool   `form:"prikaz_zatvorene"`
}

// UnmarshalText handles parsing time.Time from HTML form date input (YYYY-MM-DD format)
// This custom method intercepts form binding for time.Time fields
func unmarshalDateFromForm(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	// Try parsing with HTML date format (2006-01-02)
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		// Try parsing with ISO 8601 datetime format as fallback
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", value)
		if err != nil {
			return nil, fmt.Errorf("cannot parse date %q: %w", value, err)
		}
	}
	return &t, nil
}

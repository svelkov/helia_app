package finansijsko

import (
	"helia/internal/domain"
	"helia/pkg/utils"
	"time"
)

// config for fileds

var BrojNaloga = domain.InputFieldConfig{
	ID:         "brojnaloga",
	LabelText:  "Broj Naloga",
	FieldType:  "text",
	HasLabel:   true,
	Disabled:   false,
	ClassInput: utils.ClassInputTextEnabled,
	ClassLabel: utils.ClassLabel + " w-20",
	HxTarget:   "brojnaloga",
	MinLength:  "1",
	MaxLength:  "10",
	TabIndex:   "1"}
var DatumNaloga = domain.InputFieldConfig{
	ID:         "datumnaloga",
	Value:      time.Now().Format("2006-01-02"),
	LabelText:  "Datum Naloga",
	FieldType:  "date",
	HasLabel:   true,
	Disabled:   false,
	ClassInput: utils.ClassInputTextEnabled,
	ClassLabel: utils.ClassLabel + " pl-4",
	TabIndex:   "2"}

var DatumObrade = domain.InputFieldConfig{
	ID:         "datumobrade",
	Value:      time.Now().Format("2006-01-02"),
	LabelText:  "Datum Obrade",
	FieldType:  "date",
	HasLabel:   true,
	Disabled:   false,
	ClassInput: utils.ClassInputTextEnabled,
	ClassLabel: utils.ClassLabel + " pl-4",
	TabIndex:   "3"}

package validation

// TipanalitikeValidator implements validation for Tipanalitike entities.
type TipanalitikeValidator struct{}

// NewTipanalitikeValidator creates a new instance of TipanalitikeValidator.
func NewTipanalitikeValidator() *TipanalitikeValidator {
	return &TipanalitikeValidator{}
}

func TipanalitikeValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Sifraanalitike",
			Message: "Morate uneti sifru tipa analitike!!!",
			Check: func(value any) bool {
				num, ok := value.(int)
				return ok && num > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv tipa analitike!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

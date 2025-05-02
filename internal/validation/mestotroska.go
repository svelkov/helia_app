package validation

// PartneriValidator implements validation for Partneri entities.
type MestotroskaValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewMestotroskaValidator() *DokvrstaValidator {
	return &DokvrstaValidator{}
}

func MestotroskaValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Mtroska",
			Message: "Morate uneti mestotroska!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Opis",
			Message: "Morate uneti opis mesta troska!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

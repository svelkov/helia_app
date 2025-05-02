package validation

// PartneriValidator implements validation for Partneri entities.
type SifmestoValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewSifmestoValidator() *DokvrstaValidator {
	return &DokvrstaValidator{}
}

func SifmestoValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "SifM",
			Message: "Morate uneti sifru mesta...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv mesta...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

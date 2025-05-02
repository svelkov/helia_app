package validation

// PartneriValidator implements validation for Partneri entities.
type OpstineValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewOpstineValidator() *DokvrstaValidator {
	return &DokvrstaValidator{}
}

func OpstineValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Ops",
			Message: "Morate uneti sifru opstine...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv opstine...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

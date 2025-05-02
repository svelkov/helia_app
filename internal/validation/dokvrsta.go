package validation

// PartneriValidator implements validation for Partneri entities.
type DokvrstaValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewDokvrstaValidator() *DokvrstaValidator {
	return &DokvrstaValidator{}
}

func DokvrstaValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Vrd",
			Message: "Morate uneti vrstu dokumenta...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Opis",
			Message: "Morate uneti opis dokumenta...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

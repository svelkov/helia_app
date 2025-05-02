package validation

// PartneriValidator implements validation for Partneri entities.
type TipdokValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewTipdokValidator() *TipdokValidator {
	return &TipdokValidator{}
}

func TipdokValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "TipDok",
			Message: "Morate uneti vrstu naloga...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Opis",
			Message: "Morate uneti opis naloga...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "GrpDok",
			Message: "Morate uneti grupu dokumenata...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
}

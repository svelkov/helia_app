package validation

// PartneriValidator implements validation for Partneri entities.
type BankeValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewBankeValidator() *BankeValidator {
	return &BankeValidator{}
}

func BankeValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "BrRac",
			Message: "Morate uneti broj racuna banke...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Banka",
			Message: "Morate uneti naziv banke...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

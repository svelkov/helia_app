package validation

// PartneriValidator implements validation for Partneri entities.
type DrzavaValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewDrzavaValidator() *BankeValidator {
	return &BankeValidator{}
}

func DrzavaValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv drzave!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "OznDrz",
			Message: "Morate uneti oznaku drzave!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

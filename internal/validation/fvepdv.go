package validation

// PartneriValidator implements validation for Partneri entities.
type FvepdvValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewFvepdvValidator() *FvepdvValidator {
	return &FvepdvValidator{}
}

func FvepdvValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Vktip",
			Message: "Morate uneti tip evidencije pdv U/I...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Vkrbr",
			Message: "Morate uneti vrste r.br. evidencije pdv...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Opis",
			Message: "Morate uneti opis...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

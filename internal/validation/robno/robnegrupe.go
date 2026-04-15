package robno

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type RgruValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewRgruValidator() *RgruValidator {
	return &RgruValidator{}
}

func RgruValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Gru",
			Message: "Morate uneti grupu!!!",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv robne grupe!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

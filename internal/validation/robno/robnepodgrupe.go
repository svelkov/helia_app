package robno

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type RpgruValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewRpgruValidator() *RgruValidator {
	return &RgruValidator{}
}

func RpgruValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Gru",
			Message: "Morate izabrati grupu!!!",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "PGru",
			Message: "Morate uneti podgrupu!!!",
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

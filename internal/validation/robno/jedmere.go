package robno

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type JedmereValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewJedmereValidator() *JedmereValidator {
	return &JedmereValidator{}
}

func JedmereValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Jm",
			Message: "Morate uneti jedinicu mere!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv jedinice mere!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

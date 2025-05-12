package finnasijsko

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type FkplValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewFkplValidator() *FkplValidator {
	return &FkplValidator{}
}

func FkplValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Konto",
			Message: "Morate uneti konto...",
			Check: func(value any) bool {
				val, ok := value.(string)
				return ok && len(val) > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv konta...",
			Check: func(value any) bool {
				val, ok := value.(string)
				return ok && len(val) > 0
			},
		},
	}
}

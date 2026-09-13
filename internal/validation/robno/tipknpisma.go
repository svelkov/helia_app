package robno

import "helia/internal/validation"

// TipknpismaValidationRules defines validation rules for Tipknpisma entities.
func TipknpismaValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Opis",
			Message: "Opis je obavezan",
			Check: func(value any) bool {
				val, ok := value.(string)
				return ok && len(val) > 0
			},
		},
	}
}

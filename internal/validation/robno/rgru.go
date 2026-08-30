package robno

import "helia/internal/validation"

// RgruValidationRules defines validation rules for Rgru entities
func RgruValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Naziv",
			Message: "Naziv robne grupe je obavezan",
			Check: func(value any) bool {
				val, ok := value.(string)
				return ok && len(val) > 0
			},
		},
	}
}

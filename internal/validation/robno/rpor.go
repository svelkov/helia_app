package robno

import "helia/internal/validation"

// RporValidationRules defines validation rules for Rpor entities
func RporValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Datum",
			Message: "Datum je obavezan",
			Check: func(value any) bool {
				// Check if the date is not null
				return value != nil
			},
		},
	}
}

package robno

import "helia/internal/validation"

// RpgruValidationRules defines validation rules for Rpgru entities
func RpgruValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Naziv",
			Message: "Naziv robne podgrupe je obavezan",
			Check: func(value any) bool {
				val, ok := value.(string)
				return ok && len(val) > 0
			},
		},
	}
}

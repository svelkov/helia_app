package robno

import "helia/internal/validation"

// JedmereValidationRules defines validation rules for Jedmere entities
func JedmereValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "JM",
			Message: "Jedinica mere je obavezna",
			Check: func(value any) bool {
				val, ok := value.(string)
				return ok && len(val) > 0
			},
		},
	}
}

package finnasijsko

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type PrometValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewPrometValidator() *PrometValidator {
	return &PrometValidator{}
}

func PrometValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{

		// {
		// 	Field:   "Danal",
		// 	Message: "Morate uneti naziv banke...",
		// 	Check: func(value any) bool {
		// 		date, ok := value.(time.time)
		// 		return ok && len(str) > 0
		// 	},
		// },
	}
	//TODO SV add all vlidation rules!
}

package finnasijsko

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type SaldaValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewSaldaValidator() *SaldaValidator {
	return &SaldaValidator{}
}

func SaldaValidationRules() []validation.ValidationRule {
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

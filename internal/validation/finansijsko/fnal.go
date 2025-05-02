package finnasijsko

import "helia/internal/validation"

// PartneriValidator implements validation for Partneri entities.
type FnalValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewFnalValidator() *FnalValidator {
	return &FnalValidator{}
}

func FnalValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{
			Field:   "Nalog",
			Message: "Morate uneti broj racuna banke...",
			Check: func(value any) bool {
				val, ok := value.(int64)
				return ok && val > 0
			},
		},
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

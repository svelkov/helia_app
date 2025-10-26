package finnasijsko

import (
	"helia/global"
	"helia/internal/validation"
	"time"
)

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
			Message: "Morate uneti broj naloga...",
			Check: func(value any) bool {
				val, ok := value.(int64)
				return ok && val > 0
			},
		},
		{
			Field:   "Danal",
			Message: "Nekorektan datum...",
			Check: func(value any) bool {
				date, ok := value.(time.Time)
				return ok && date.Year() == global.GetGnGod()
			},
		},
		{
			Field:   "Datob",
			Message: "Nekorektan datum...",
			Check: func(value any) bool {
				date, ok := value.(time.Time)
				return ok && date.Year() == global.GetGnGod()
			},
		},
	}
}

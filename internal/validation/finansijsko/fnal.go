package finnasijsko

import (
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
				// Note: Year validation should be done at service level with context-aware god value
				return ok && !date.IsZero()
			},
		},
		{
			Field:   "Datob",
			Message: "Nekorektan datum...",
			Check: func(value any) bool {
				date, ok := value.(time.Time)
				// Note: Year validation should be done at service level with context-aware god value
				return ok && !date.IsZero()
			},
		},
	}
}

// FnalValidationRulesWithGod provides validation rules with business year context
func FnalValidationRulesWithGod(god int) []validation.ValidationRule {
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
				return ok && date.Year() == god
			},
		},
		{
			Field:   "Datob",
			Message: "Nekorektan datum...",
			Check: func(value any) bool {
				date, ok := value.(time.Time)
				return ok && date.Year() == god
			},
		},
	}
}

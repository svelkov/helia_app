package finnasijsko

import (
	"fmt"
	"helia/internal/domain"
	"helia/internal/validation"
	"time"
)

// FnalValidator implements validation for Fnal entities combining generic rules with entity-level validation
type FnalValidator struct {
	*validation.RuleBasedValidator[domain.Fnal]
}

// NewFnalValidator creates a new instance of FnalValidator with generic field validation rules.
func NewFnalValidator() *FnalValidator {
	return &FnalValidator{
		RuleBasedValidator: validation.NewRuleBasedValidator[domain.Fnal](FnalValidationRules()),
	}
}

// Validate performs both generic field validation and entity-level validation
func (v *FnalValidator) Validate(entity *domain.Fnal) ([]domain.FieldError, error) {
	// First, apply generic field validation rules
	fieldErrors, err := v.RuleBasedValidator.Validate(entity)
	if err != nil {
		return fieldErrors, err
	}

	// Then, apply entity-level validation for year matching
	entityErrors := v.validateEntityRules(entity)
	fieldErrors = append(fieldErrors, entityErrors...)

	return fieldErrors, nil
}

// validateEntityRules performs entity-level validation that requires access to multiple fields
func (v *FnalValidator) validateEntityRules(fnal *domain.Fnal) []domain.FieldError {
	var fieldErrors []domain.FieldError

	// Validate that Danal year matches God
	if !fnal.Danal.IsZero() && fnal.Danal.Year() != fnal.God {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "danal",
			ErrorMessage: fmt.Sprintf("Godina datuma naloga mora biti %d", fnal.God),
		})
	}

	// Validate that Datob year matches God
	if !fnal.Datob.IsZero() && fnal.Datob.Year() != fnal.God {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "datob",
			ErrorMessage: fmt.Sprintf("Godina datuma obrade mora biti %d", fnal.God),
		})
	}

	return fieldErrors
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
				return ok && !date.IsZero()
			},
		},
		{
			Field:   "Datob",
			Message: "Nekorektan datum...",
			Check: func(value any) bool {
				date, ok := value.(time.Time)
				return ok && !date.IsZero()
			},
		},
	}
}

package validation

import (
	"fmt"
	"helia/internal/domain"
	"reflect"
	"strings"
)

// Validator is the interface that all validators must implement
type Validator[T any] interface {
	Validate(entity *T) ([]domain.FieldError, error)
}

// ValidationRule defines a single validation rule
type ValidationRule struct {
	Field   string
	Message string
	Check   func(value any) bool
}

type RuleBasedValidator[T any] struct {
	Rules []ValidationRule // A slice of rules for the entity
}

// NewRuleBasedValidator creates a new validator with the provided rules.
func NewRuleBasedValidator[T any](rules []ValidationRule) *RuleBasedValidator[T] {
	return &RuleBasedValidator[T]{Rules: rules}
}

// Validate applies all rules to the entity and returns an error if any validation fails.
func (v *RuleBasedValidator[T]) Validate(entity *T) ([]domain.FieldError, error) {
	val := reflect.ValueOf(entity).Elem() // Reflect over the entity
	fieldErrors := []domain.FieldError{}
	for _, rule := range v.Rules {
		field := val.FieldByName(rule.Field)
		if !field.IsValid() {
			return fieldErrors, fmt.Errorf("field %s does not exist in the entity", rule.Field)
		}

		// Check the validation rule
		if !rule.Check(field.Interface()) {
			fieldErrors = append(fieldErrors, domain.FieldError{Field: strings.ToLower(rule.Field), ErrorMessage: rule.Message})

		}
	}

	return fieldErrors, nil
}

// EmptyValidator is a generic validator with no validation rules (pass-through)
type EmptyValidator[T any] struct{}

// NewValidator creates a new empty validator (no validation rules)
func NewValidator[T any]() *EmptyValidator[T] {
	return &EmptyValidator[T]{}
}

// Validate always returns nil (no validation errors)
func (v *EmptyValidator[T]) Validate(entity *T) ([]domain.FieldError, error) {
	return []domain.FieldError{}, nil
}

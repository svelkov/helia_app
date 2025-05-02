package validation

// PartneriValidator implements validation for Partneri entities.
type BnkizvValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewBnkizvValidator() *BnkizvValidator {
	return &BnkizvValidator{}
}

func BnkizvValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "Sifbank",
			Message: "Morate uneti difru banke...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0 && val < 1000
			},
		},
		{
			Field:   "Bnkdes",
			Message: "Morate uneti naziv banke...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Swiftadr",
			Message: "Morate uneti naziv swift adresu...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Brojrac",
			Message: "Morate uneti broj racuna banke...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
}

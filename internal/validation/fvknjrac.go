package validation

// PartneriValidator implements validation for Partneri entities.
type FvknjracValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewFvknjracValidator() *FvknjracValidator {
	return &FvknjracValidator{}
}

func FvknjracValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "VkTip",
			Message: "Morate uneti tip knjige ulaznih/izlaznih racuna...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "VkRbr",
			Message: "Morate uneti vrstu knjige ulaznih/izlaznih racuna...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Opis",
			Message: "Morate uneti opis knjige ulaznih/izlaznih racuna...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

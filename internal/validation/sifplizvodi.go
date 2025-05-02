package validation

// PartneriValidator implements validation for Partneri entities.
type SifplizvValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewSifplizvValidator() *SifplizvValidator {
	return &SifplizvValidator{}
}

func SifplizvValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "SifPlac",
			Message: "Morate uneti sifru placanja...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Oblik",
			Message: "Morate uneti oblik placanja...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Osnov",
			Message: "Morate uneti osnov placanja...",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && val > 0
			},
		},
		{
			Field:   "Opis",
			Message: "Morate uneti opis placanja...",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
}

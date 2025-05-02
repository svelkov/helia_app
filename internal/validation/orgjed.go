package validation

// PartneriValidator implements validation for Partneri entities.
type OrgjedValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewOrgjedValidator() *DokvrstaValidator {
	return &DokvrstaValidator{}
}

func OrgjedValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "OjOzn",
			Message: "Morate uneti oznaku organizacione jedinice!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "Naziv",
			Message: "Morate uneti naziv organizacione jedinice!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
	}
	//TODO SV add all vlidation rules!
}

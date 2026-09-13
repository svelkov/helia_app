package robno

import "helia/internal/validation"

func FispValidationRules() []validation.ValidationRule {
	return []validation.ValidationRule{
		{Field: "konto", Message: "Konto je obavezan", Check: func(value any) bool { v, ok := value.(string); return ok && v != "" }},
		{Field: "sifra", Message: "Šifra partnera je obavezna", Check: func(value any) bool { v, ok := value.(string); return ok && v != "" }},
		{Field: "naziv", Message: "Naziv mesta isporuke je obavezan", Check: func(value any) bool { v, ok := value.(string); return ok && v != "" }},
	}
}

package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	instance *Service
	once     sync.Once
)

type Service struct {
	translations map[string]map[string]interface{}
	fallbackLang string
	mu           sync.RWMutex
}

// Init initializes the translation service (call this once at app startup)
func Init(translationsPath string, languages []string, fallbackLang string) error {
	var initErr error
	once.Do(func() {
		instance = &Service{
			translations: make(map[string]map[string]interface{}),
			fallbackLang: fallbackLang,
		}
		initErr = instance.loadTranslations(translationsPath, languages, fallbackLang)
	})
	return initErr
}

// GetInstance returns the singleton instance
func GetInstance() *Service {
	return instance
}

// LoadTranslations loads translation files from a directory
func (s *Service) loadTranslations(dir string, languages []string, fallbackLang string) error {

	for _, lang := range languages {
		filename := fmt.Sprintf("%s/%s.json", dir, lang)
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}

		var translations map[string]interface{}
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("error parsing %s: %w", filename, err)
		}

		s.translations[lang] = translations
	}

	return nil
}

// T returns translation for a key
func (s *Service) T(lang, key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try requested language
	if trans, ok := s.translations[lang]; ok {
		if val, exists := trans[key]; exists {
			return val.(string)
		}
	}

	// Fallback to default language
	if trans, ok := s.translations[s.fallbackLang]; ok {
		if val, exists := trans[key]; exists {
			return val.(string)
		}
	}

	// Return key if not found
	return key
}

// Convenience methods
func (s *Service) Menu(lang, menuName string) string {
	return s.T(lang, "menu."+menuName)
}

func (s *Service) Button(lang, buttonName string) string {
	return s.T(lang, "button."+buttonName)
}

func (s *Service) Label(lang, labelName string) string {
	return s.T(lang, "label."+labelName)
}

func (s *Service) Message(lang, messageName string) string {
	return s.T(lang, "message."+messageName)
}
func (s *Service) Validation(lang, validationName string) string {
	return s.T(lang, "validation."+validationName)
}

func (s *Service) Form(lang, formName string) string {
	return s.T(lang, "form."+formName)
}

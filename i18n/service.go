package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	instance        *Service
	once            sync.Once
	reg             = regexp.MustCompile(`_+`)
	regSpecialChars = regexp.MustCompile(`[/\-,()%:?\s]+`)
)

type Service struct {
	translations map[string]map[string]interface{}
	fallbackLang string
	currentLang  string
	mu           sync.RWMutex
}

// Init initializes the translation service (call this once at app startup)
func Init(translationsPath string, languages []string, fallbackLang string) error {
	var initErr error
	once.Do(func() {
		instance = &Service{
			translations: make(map[string]map[string]interface{}),
			fallbackLang: fallbackLang,
			currentLang:  fallbackLang, // Set initial current language
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

// SetLanguage changes the current language at runtime
func (s *Service) SetLanguage(lang string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.translations[lang]; !exists {
		return fmt.Errorf("language %s not loaded", lang)
	}

	s.currentLang = lang
	return nil
}

// GetCurrentLanguage returns the current active language
func (s *Service) GetCurrentLanguage() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLang
}

// T returns translation for a key
func (s *Service) T(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key = toSnakeCase(key)
	// Try requested language
	if trans, ok := s.translations[s.currentLang]; ok {
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

// TWithParams returns translation with dynamic parameter replacement
// Example: TWithParams("label.dospeva", map[string]string{"days": "15"})
// Translation: "Dospeva {{days}} dana" -> "Dospeva 15 dana"
func (s *Service) TWithParams(key string, params map[string]string) string {
	translation := s.T(key)

	for k, v := range params {
		placeholder := fmt.Sprintf("{{%s}}", k)
		translation = strings.ReplaceAll(translation, placeholder, v)
	}

	return translation
}

// Convenience methods
func (s *Service) Menu(menuName string) string {
	return s.T("menu." + menuName)
}

func (s *Service) Title(labelName string) string {
	return s.T("title." + labelName)
}
func (s *Service) TitleWithParams(labelName string, params map[string]string) string {
	return s.TWithParams("title."+labelName, params)
}
func (s *Service) Text(textName string) string {
	return s.T("text." + textName)
}

func (s *Service) TextWithParams(textName string, params map[string]string) string {
	return s.TWithParams("text."+textName, params)
}

func (s *Service) Button(buttonName string) string {
	return s.T("button." + buttonName)
}
func (s *Service) ButtonWithParams(buttonName string, params map[string]string) string {
	return s.TWithParams("button."+buttonName, params)
}
func (s *Service) Label(labelName string) string {
	return s.T("label." + labelName)
}

func (s *Service) LabelWithParams(labelName string, params map[string]string) string {
	return s.TWithParams("label."+labelName, params)
}

func (s *Service) Placeholder(placeholder string) string {
	return s.T("placeholder." + placeholder)
}

func (s *Service) Message(messageName string) string {
	return s.T("message." + messageName)
}

func (s *Service) MessageWithParams(messageName string, params map[string]string) string {
	return s.TWithParams("message."+messageName, params)
}

func (s *Service) Validation(validationName string) string {
	return s.T("validation." + validationName)
}

func (s *Service) Form(formName string) string {
	return s.T("form." + formName)
}

func (s *Service) FormWithParams(formName string, params map[string]string) string {
	return s.TWithParams("form."+formName, params)
}

// toSnakeCase safely converts text to snake_case for translation keys
func toSnakeCase(s string) string {
	// Convert to lowercase
	result := strings.ToLower(s)

	// Replace Cyrillic characters
	cyrillic := map[string]string{
		"č": "c", "ž": "z", "š": "s", "đ": "d", "ć": "c",
	}
	for char, replacement := range cyrillic {
		result = strings.ReplaceAll(result, char, replacement)
	}

	// Replace "/" with space
	result = strings.ReplaceAll(result, "/", " ")

	// Replace special characters and spaces with underscores
	result = regSpecialChars.ReplaceAllString(result, "_")

	// Handle special case for " i " (Serbian "and")
	result = strings.ReplaceAll(result, " i ", "_")

	// Remove multiple underscores
	result = reg.ReplaceAllString(result, "_")

	// Trim underscores from start and end
	result = strings.Trim(result, "_")

	return result
}

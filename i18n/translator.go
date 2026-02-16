package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Translator handles translations
type Translator struct {
	translations map[string]map[string]interface{}
	mu           sync.RWMutex
	fallback     string
}

// NewTranslator creates a new translator instance
func NewTranslator(fallbackLang string) *Translator {
	return &Translator{
		translations: make(map[string]map[string]interface{}),
		fallback:     fallbackLang,
	}
}

// LoadTranslations loads translation files from a directory
func (t *Translator) LoadTranslations(dir string, langs []string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, lang := range langs {
		filename := fmt.Sprintf("%s/%s.json", dir, lang)
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}

		var translations map[string]interface{}
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("error parsing %s: %w", filename, err)
		}

		t.translations[lang] = translations
	}

	return nil
}

// Get retrieves a translation using dot notation (e.g., "message.description")
func (t *Translator) Get(lang, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try requested language
	if val := t.getValue(lang, key); val != "" {
		return val
	}

	// Fallback to default language
	if lang != t.fallback {
		if val := t.getValue(t.fallback, key); val != "" {
			return val
		}
	}

	// Return key if not found
	return key
}

// getValue retrieves nested value using dot notation
func (t *Translator) getValue(lang, key string) string {
	langMap, ok := t.translations[lang]
	if !ok {
		return ""
	}

	keys := strings.Split(key, ".")
	var current interface{} = langMap

	for _, k := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[k]
		default:
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return ""
}

// GetWithParams retrieves translation and replaces {{param}} placeholders
func (t *Translator) GetWithParams(lang, key string, params map[string]string) string {
	translation := t.Get(lang, key)

	for k, v := range params {
		placeholder := fmt.Sprintf("{{%s}}", k)
		translation = strings.ReplaceAll(translation, placeholder, v)
	}

	return translation
}

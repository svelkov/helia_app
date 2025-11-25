package i18n

// Helper functions that use current app language

func T(key string, lang string) string {
	return GetInstance().T(lang, key)
}

func Menu(menuName string, lang string) string {
	return GetInstance().Menu(lang, menuName)
}

func Button(buttonName string, lang string) string {
	return GetInstance().Button(lang, buttonName)
}

func Label(labelName string, lang string) string {
	return GetInstance().Label(lang, labelName)
}

func Message(messageName string, lang string) string {
	return GetInstance().Message(lang, messageName)
}

func Form(formName string, lang string) string {
	return GetInstance().Form(lang, formName)
}

func Validation(validationName string, lang string) string {
	return GetInstance().Validation(lang, validationName)
}

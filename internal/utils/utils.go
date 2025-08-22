package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetValueFromMap performs a case-insensitive search for a key in a map.
func GetValueFromMap(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}

	lowerKey := strings.ToLower(key)
	if v, ok := m[lowerKey]; ok {
		return v
	}

	upperKey := strings.ToUpper(key)
	if v, ok := m[upperKey]; ok {
		return v
	}

	caser := cases.Title(language.English, cases.Compact)
	titleKey := caser.String(key)
	if v, ok := m[titleKey]; ok {
		return v
	}

	return ""
}


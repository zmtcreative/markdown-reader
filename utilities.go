package main

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetValueFromMap(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	key = strings.ToLower(key)
	if v, ok := m[key]; ok {
		return v
	}
	key = strings.ToUpper(key)
	if v, ok := m[key]; ok {
		return v
	}
	caser := cases.Title(language.English, cases.Compact)
	key = caser.String(key)
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}


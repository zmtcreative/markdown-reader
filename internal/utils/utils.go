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

// NormalizeMapKeys takes a map[string]any and converts all keys to lowercase recursively,
// returning a new map with normalized keys and the original values. It handles nested maps
// by recursively normalizing their keys as well.
func NormalizeMapKeys(input map[string]any) (map[string]any, error) {
	if input == nil {
		return nil, nil
	}

	normalized := make(map[string]any, len(input))
	for key, value := range input {
		lowerKey := strings.ToLower(key)

		// Check if the value is itself a map[string]any and recursively normalize it
		if nestedMap, ok := value.(map[string]any); ok {
			normalizedNested, err := NormalizeMapKeys(nestedMap)
			if err != nil {
				return nil, err
			}
			normalized[lowerKey] = normalizedNested
		} else {
			// For non-map values, just copy the value as-is
			normalized[lowerKey] = value
		}
	}

	return normalized, nil
}

// Generic helper function to retrieve a value from a map[string]any with a specific key.
// Usage: getValue[YourType](yourMap, "yourKey")
//   if title, ok := getValue[string](docFrontmatter, "title"); ok {
//       fmDocTitle = title
//   }
//   if date, ok := getValue[time.Time](docFrontmatter, "date"); ok {
//       fmDocDate = date.String()
//   }
func GetValue[T any](m map[string]any, key string) (T, bool) {
    var zero T
    if m == nil {
        return zero, false
    }
    if val, ok := m[key].(T); ok {
        return val, true
    }
    return zero, false
}




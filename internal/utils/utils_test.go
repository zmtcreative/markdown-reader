package utils

import (
	"testing"
)

func TestGetValueFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]string
		key      string
		expected string
	}{
		{
			name:     "exact case match",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "lowercase key in mixed case map",
			m:        map[string]string{"Key": "value", "Another": "test"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "uppercase key in mixed case map",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "KEY",
			expected: "value",
		},
		{
			name:     "title case key in lowercase map",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "Key",
			expected: "value",
		},
		{
			name:     "mixed case variations",
			m:        map[string]string{"Testkey": "testvalue", "other": "othervalue"},
			key:      "testkey",
			expected: "testvalue",
		},
		{
			name:     "key not found",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "notfound",
			expected: "",
		},
		{
			name:     "empty map",
			m:        map[string]string{},
			key:      "key",
			expected: "",
		},
		{
			name:     "empty key",
			m:        map[string]string{"key": "value", "": "empty"},
			key:      "",
			expected: "empty",
		},
		{
			name:     "special characters in key",
			m:        map[string]string{"key-with-dashes": "value", "key_with_underscores": "undervalue"},
			key:      "KEY-WITH-DASHES",
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValueFromMap(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("GetValueFromMap() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetValueFromMapConsistency(t *testing.T) {
	// Test that the function consistently returns the same value for the same input
	m := map[string]string{"key": "value1", "KEY": "value2", "Key": "value3"}
	key := "key"

	first := GetValueFromMap(m, key)
	for i := 0; i < 10; i++ {
		result := GetValueFromMap(m, key)
		if result != first {
			t.Errorf("GetValueFromMap() returned inconsistent result on iteration %d: got %q, want %q", i, result, first)
		}
	}
}

func BenchmarkGetValueFromMap(b *testing.B) {
	m := map[string]string{
		"key1":     "value1",
		"key2":     "value2",
		"KEY3":     "value3",
		"Key4":     "value4",
		"TestKey5": "value5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetValueFromMap(m, "testkey5")
	}
}

package app

import (
	"strings"
	"testing"
)

func TestFontManager_IsMonospaceFont(t *testing.T) {
	// Create a config manager for the test
	cm := NewConfigManager()
	fm := NewFontManager(cm)

	testCases := []struct {
		name     string
		fontName string
		expected bool
	}{
		{"Consolas is monospace", "Consolas", true},
		{"Courier New is monospace", "Courier New", true},
		{"Monaco is monospace", "Monaco", true},
		{"Arial is not monospace", "Arial", false},
		{"Times New Roman is not monospace", "Times New Roman", false},
		{"Fira Code is monospace", "Fira Code", true},
		{"JetBrains Mono is monospace", "JetBrains Mono", true},
		{"Non-existent font", "NonExistentFont", false},
		{"Empty string", "", false},
		{"Font with fallback - Consolas", "Consolas, monospace", true},
		{"Font with fallback - Arial", "Arial, sans-serif", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fm.IsMonospaceFont(tc.fontName)
			if result != tc.expected {
				t.Errorf("IsMonospaceFont(%q) = %v; expected %v", tc.fontName, result, tc.expected)
			}
		})
	}
}

func TestFontManager_GetMonospaceFonts(t *testing.T) {
	// Create a config manager for the test
	cm := NewConfigManager()
	fm := NewFontManager(cm)
	monospaceFonts := fm.GetMonospaceFonts()

	// Should contain at least some fallback fonts
	if len(monospaceFonts) == 0 {
		t.Error("GetMonospaceFonts() returned empty list, expected at least fallback fonts")
	}

	// Test that at least our fallback fonts are present
	foundConsolas := false
	foundCourier := false

	for _, font := range monospaceFonts {
		if strings.Contains(strings.ToLower(font), "consolas") {
			foundConsolas = true
		}
		if strings.Contains(strings.ToLower(font), "courier new") {
			foundCourier = true
		}
	}

	if !foundConsolas {
		t.Log("Consolas not found in font list (may be expected on non-Windows systems)")
	}
	if !foundCourier {
		t.Log("Courier New not found in font list (may be expected on some systems)")
	}

	// Should contain common monospace fonts
	expectedFonts := []string{"Consolas", "Courier New", "Monaco"}
	for _, expected := range expectedFonts {
		found := false
		for _, font := range monospaceFonts {
			if fm.GetFontFamilyName(font) == expected {
				found = true
				break
			}
		}
		if !found {
			// This is acceptable since system fonts may not be available,
			// but we should have fallbacks
			t.Logf("Expected monospace font %s not found in system fonts (this is OK if fallbacks are present)", expected)
		}
	}
}

func TestConfigManager_MonospaceFontGettersAndSetters(t *testing.T) {
	cm := NewConfigManager()

	// Test default values
	defaultFontFamily := cm.GetFontFamilyMono()
	if defaultFontFamily == "" {
		t.Error("Default monospace font family should not be empty")
	}

	defaultFontSize := cm.GetFontSizeMono()
	if defaultFontSize <= 0 {
		t.Error("Default monospace font size should be greater than 0")
	}

	// Test advanced detection default
	defaultAdvancedDetection := cm.GetUseAdvancedFontDetection()
	if !defaultAdvancedDetection {
		t.Error("Default advanced font detection should be enabled")
	}

	// Test setters
	testFontFamily := "JetBrains Mono, Consolas, monospace"
	cm.SetFontFamilyMono(testFontFamily)
	if cm.GetFontFamilyMono() != testFontFamily {
		t.Errorf("SetFontFamilyMono() failed: expected %s, got %s", testFontFamily, cm.GetFontFamilyMono())
	}

	testFontSize := 15.5
	cm.SetFontSizeMono(testFontSize)
	if cm.GetFontSizeMono() != testFontSize {
		t.Errorf("SetFontSizeMono() failed: expected %f, got %f", testFontSize, cm.GetFontSizeMono())
	}

	// Test advanced detection setter
	cm.SetUseAdvancedFontDetection(false)
	if cm.GetUseAdvancedFontDetection() {
		t.Error("SetUseAdvancedFontDetection(false) failed")
	}

	cm.SetUseAdvancedFontDetection(true)
	if !cm.GetUseAdvancedFontDetection() {
		t.Error("SetUseAdvancedFontDetection(true) failed")
	}
}

func TestFontManager_DetectionModes(t *testing.T) {
	cm := NewConfigManager()
	fm := NewFontManager(cm)

	// Test basic detection
	cm.SetUseAdvancedFontDetection(false)
	basicFonts := fm.GetMonospaceFonts()

	// Test advanced detection
	cm.SetUseAdvancedFontDetection(true)
	advancedFonts := fm.GetMonospaceFonts()

	// Both should return fonts
	if len(basicFonts) == 0 {
		t.Error("Basic detection should return at least fallback fonts")
	}

	if len(advancedFonts) == 0 {
		t.Error("Advanced detection should return at least fallback fonts")
	}

	// Test that both detection methods exist independently
	basicOnly := fm.BasicMonospaceDetection()
	advancedOnly := fm.AdvancedMonospaceDetection()

	if len(basicOnly) == 0 {
		t.Error("BasicMonospaceDetection should return fonts")
	}

	if len(advancedOnly) == 0 {
		t.Error("AdvancedMonospaceDetection should return fonts")
	}
}

func TestFontManager_AdvancedNameDetection(t *testing.T) {
	cm := NewConfigManager()
	fm := NewFontManager(cm)

	testCases := []struct {
		name     string
		fontName string
		expected bool
	}{
		{"Source Code Pro detected", "Source Code Pro", true},
		{"JetBrains Mono detected", "JetBrains Mono", true},
		{"Something Mono detected", "Something Mono", true},
		{"Terminal Font detected", "Terminal Font", true},
		{"Fixed Width detected", "Fixed Width", true},
		{"Arial Text should not be detected", "Arial Text", false},
		{"Times Display should not be detected", "Times Display", false},
		{"Comic Sans should not be detected", "Comic Sans", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fm.detectMonospaceByName(tc.fontName)
			if result != tc.expected {
				t.Errorf("detectMonospaceByName(%q) = %v; expected %v", tc.fontName, result, tc.expected)
			}
		})
	}
}

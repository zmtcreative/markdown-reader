package app

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestConfigManager(t *testing.T, args ...string) *ConfigManager {
	t.Helper()

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	if len(args) == 0 {
		os.Args = []string{"md-reader"}
	} else {
		os.Args = args
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	return NewConfigManager()
}

func TestNewConfigManager(t *testing.T) {
	cm := newTestConfigManager(t)
	if cm == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	// Test that defaults are set correctly
	config := cm.GetConfig()
	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Test Application defaults
	if !config.Application.UseInlineHTML {
		t.Error("Default UseInlineHTML should be true")
	}
	if !config.Application.UseSanitize {
		t.Error("Default UseSanitize should be true")
	}
	if !config.Application.UseStripH1 {
		t.Error("Default StripH1 should be true")
	}
	if !config.Application.UseFrontmatterTitle {
		t.Error("Default UseFrontmatterTitle should be true")
	}

	// Test Markdown defaults
	if !config.Markdown.UseGFM {
		t.Error("Default UseGFM should be true")
	}
	if !config.Markdown.UseEmoji {
		t.Error("Default UseEmoji should be true")
	}
	if !config.Markdown.UseMermaid {
		t.Error("Default UseMermaid should be true")
	}

	// Test Alert Callouts defaults
	if !config.AlertCallouts.UseAlertCallouts {
		t.Error("Default UseAlertCallouts should be true")
	}
	if config.AlertCallouts.AlertCalloutStyle != "GFMPlus" {
		t.Errorf("Default AlertCalloutStyle = %q, want %q", config.AlertCallouts.AlertCalloutStyle, "GFMPlus")
	}
}

func TestConfigManagerGettersAndSetters(t *testing.T) {
	cm := newTestConfigManager(t)

	// Test initial values
	if !cm.UseInlineHTML() {
		t.Error("UseInlineHTML() should initially be true")
	}
	if !cm.UseSanitize() {
		t.Error("UseSanitize() should initially be true")
	}
	if !cm.UseGFM() {
		t.Error("UseGFM() should initially be true")
	}

	// Create new config and set it
	newConfig := &Config{
		Application: ApplicationOptions{
			UseInlineHTML:  false,
			UseSanitize:    false,
			UseStripH1:     false,
			UseFrontmatterTitle: false,
		},
		Markdown: MarkdownOptions{
			UseGFM:          false,
			UseEmoji:        false,
			UseMermaid:      false,
			UseFigure:       false,
			UseAnchor:       false,
			UseFences:       false,
			UseSections:     false,
			UseHighlighting: false,
			UseFancyLists:   false,
			UseAttributes:   false,
			UseTypographic:  false,
		},
		AlertCallouts: AlertCalloutOptions{
			UseAlertCallouts:  false,
			AlertCalloutStyle: "GFMStrict",
		},
	}

	cm.SetConfig(newConfig)

	// Test that values changed
	if cm.UseInlineHTML() {
		t.Error("UseInlineHTML() should be false after SetConfig")
	}
	if cm.UseSanitize() {
		t.Error("UseSanitize() should be false after SetConfig")
	}
	if cm.UseGFM() {
		t.Error("UseGFM() should be false after SetConfig")
	}
	if cm.AlertCalloutStyle() != "GFMStrict" {
		t.Errorf("AlertCalloutStyle() = %q, want %q", cm.AlertCalloutStyle(), "GFMStrict")
	}
}

func TestValidateAlertCalloutStyle(t *testing.T) {
	cm := newTestConfigManager(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid GFMStrict",
			input:    "GFMStrict",
			expected: "GFMStrict",
		},
		{
			name:     "valid GFMWithAliases",
			input:    "GFMWithAliases",
			expected: "GFMWithAliases",
		},
		{
			name:     "valid GFMPlus",
			input:    "GFMPlus",
			expected: "GFMPlus",
		},
		{
			name:     "valid Obsidian",
			input:    "Obsidian",
			expected: "Obsidian",
		},
		{
			name:     "invalid style",
			input:    "InvalidStyle",
			expected: "GFMPlus", // default
		},
		{
			name:     "empty string",
			input:    "",
			expected: "GFMPlus", // default
		},
		{
			name:     "case sensitive",
			input:    "gfmstrict", // lowercase
			expected: "GFMPlus",   // default because case matters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.ValidateAlertCalloutStyle(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateAlertCalloutStyle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetApplicationConfig(t *testing.T) {
	cm := newTestConfigManager(t)

	useInlineHTML, useSanitize := cm.GetApplicationConfig()
	if !useInlineHTML {
		t.Error("GetApplicationConfig() useInlineHTML should be true")
	}
	if !useSanitize {
		t.Error("GetApplicationConfig() useSanitize should be true")
	}

	// Change config and test again
	newConfig := &Config{
		Application: ApplicationOptions{
			UseInlineHTML: false,
			UseSanitize:   false,
		},
	}
	cm.SetConfig(newConfig)

	useInlineHTML, useSanitize = cm.GetApplicationConfig()
	if useInlineHTML {
		t.Error("GetApplicationConfig() useInlineHTML should be false after change")
	}
	if useSanitize {
		t.Error("GetApplicationConfig() useSanitize should be false after change")
	}
}

func TestGetAlertCalloutConfig(t *testing.T) {
	cm := newTestConfigManager(t)

	style := cm.GetAlertCalloutConfig()
	if style != "GFMPlus" {
		t.Errorf("GetAlertCalloutConfig() = %q, want %q", style, "GFMPlus")
	}

	// Change config and test again
	newConfig := &Config{
		AlertCallouts: AlertCalloutOptions{
			AlertCalloutStyle: "GFMStrict",
		},
	}
	cm.SetConfig(newConfig)

	style = cm.GetAlertCalloutConfig()
	if style != "GFMStrict" {
		t.Errorf("GetAlertCalloutConfig() = %q, want %q", style, "GFMStrict")
	}
}

func TestApplyCliOverrides(t *testing.T) {
	cm := newTestConfigManager(t)

	// Initial values should be defaults (true)
	if !cm.UseInlineHTML() || !cm.UseSanitize() || !cm.UseStripH1() {
		t.Fatal("Initial configuration should have all true values")
	}

	// Apply CLI overrides
	allowInlineHTML := false
	sanitizeHTML := false
	stripH1 := false

	cm.ApplyCliOverrides(&allowInlineHTML, &sanitizeHTML, &stripH1)

	// Test that overrides were applied
	if cm.UseInlineHTML() {
		t.Error("UseInlineHTML() should be false after CLI override")
	}
	if cm.UseSanitize() {
		t.Error("UseSanitize() should be false after CLI override")
	}
	if cm.UseStripH1() {
		t.Error("UseStripH1() should be false after CLI override")
	}

	// Test with nil values (should not change)
	cm.ApplyCliOverrides(nil, nil, nil)

	// Values should remain false
	if cm.UseInlineHTML() || cm.UseSanitize() || cm.UseStripH1() {
		t.Error("Values should remain false when nil overrides are applied")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	cm := newTestConfigManager(t, "test-md-reader")

	// Test that we can set and get configuration
	newConfig := &Config{
		Application: ApplicationOptions{
			UseInlineHTML:  false,
			UseSanitize:    false,
			UseStripH1:     false,
			UseFrontmatterTitle: false,
		},
		Markdown: MarkdownOptions{
			UseGFM:   false,
			UseEmoji: false,
		},
		AlertCallouts: AlertCalloutOptions{
			UseAlertCallouts:  false,
			AlertCalloutStyle: "GFMStrict",
		},
	}

	cm.SetConfig(newConfig)

	// Verify that the configuration was set correctly
	if cm.UseInlineHTML() {
		t.Error("Config should have UseInlineHTML = false")
	}
	if cm.UseSanitize() {
		t.Error("Config should have UseSanitize = false")
	}
	if cm.AlertCalloutStyle() != "GFMStrict" {
		t.Errorf("Config AlertCalloutStyle = %q, want %q", cm.AlertCalloutStyle(), "GFMStrict")
	}

	// Test that SaveConfig doesn't return an error (even if it creates files in user directories)
	err := cm.SaveConfig()
	if err != nil {
		t.Errorf("SaveConfig() error = %v", err)
	}

	if _, err := os.Stat(cm.configPath); err != nil {
		t.Fatalf("SaveConfig() did not create config file %q: %v", cm.configPath, err)
	}

	wantConfigPath := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "test-md-reader", "test-md-reader.json")
	if cm.configPath != wantConfigPath {
		t.Fatalf("configPath = %q, want %q", cm.configPath, wantConfigPath)
	}
}

func TestGetAppNameFromExecutable(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple name",
			args:     []string{"md-reader"},
			expected: "md-reader",
		},
		{
			name:     "with extension",
			args:     []string{"md-reader.exe"},
			expected: "md-reader",
		},
		{
			name:     "full path",
			args:     []string{"/usr/local/bin/md-reader"},
			expected: "md-reader",
		},
		{
			name:     "windows path",
			args:     []string{"C:\\Program Files\\md-reader\\md-reader.exe"},
			expected: "md-reader",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "md-reader", // fallback
		},
		{
			name:     "different name",
			args:     []string{"markdown-viewer.exe"},
			expected: "markdown-viewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			result := getAppNameFromExecutable()
			if result != tt.expected {
				t.Errorf("getAppNameFromExecutable() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAlertCalloutStylesConstant(t *testing.T) {
	// Test that the AlertCalloutStyles map is properly defined
	expectedStyles := map[string]string{
		"GFMStrict":      "Strict GFM Alerts (Standard 5 Alert Types)",
		"GFMWithAliases": "GFM Alerts + Aliases (GFM + Aliases [e.g., notes = note])",
		"GFMPlus":        "GFM Alerts Plus (GFM + Some Obsidian-Style Callouts)",
		"Obsidian":       "Obsidian-Style (Obsidian Icons and Callout Names)",
	}

	for key, expectedDesc := range expectedStyles {
		if desc, exists := AlertCalloutStyles[key]; !exists {
			t.Errorf("AlertCalloutStyles missing key: %s", key)
		} else if desc != expectedDesc {
			t.Errorf("AlertCalloutStyles[%s] = %q, want %q", key, desc, expectedDesc)
		}
	}

	if len(AlertCalloutStyles) != len(expectedStyles) {
		t.Errorf("AlertCalloutStyles length = %d, want %d", len(AlertCalloutStyles), len(expectedStyles))
	}
}

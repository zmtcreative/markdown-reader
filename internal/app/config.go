package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	AllowInlineHTML    bool   `mapstructure:"allow_inline_html" json:"allow_inline_html"`
	SanitizeHTML       bool   `mapstructure:"sanitize_html" json:"sanitize_html"`
	StripH1            bool   `mapstructure:"strip_h1" json:"strip_h1"`
	AlertCalloutStyle  string `mapstructure:"alert_callout_style" json:"alert_callout_style"`
}

// ConfigManager handles configuration loading and saving
type ConfigManager struct {
	config     *Config
	configPath string
	viper      *viper.Viper
}

// AlertCalloutStyles defines the available alert callout styles
// (Note to self: Do NOT change "GFMStrict" if you can help it -- this is being used as the fallback default)
var AlertCalloutStyles = []string{"GFMStrict", "GFMWithAliases", "GFMPlus", "Obsidian"}

// getAppNameFromExecutable extracts the application name from the executable path
// without the file extension
func getAppNameFromExecutable() string {
	if len(os.Args) == 0 {
		return "md-reader" // fallback
	}

	// Get the base name from the executable path
	execPath := os.Args[0]
	baseName := filepath.Base(execPath)

	// Remove the file extension
	ext := filepath.Ext(baseName)
	if ext != "" {
		baseName = strings.TrimSuffix(baseName, ext)
	}

	// Use fallback if empty or invalid
	if baseName == "" || baseName == "." {
		return "md-reader"
	}

	return baseName
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	v := viper.New()

	// Get appName
	appName := getAppNameFromExecutable()

	// Set configuration file name and type
	v.SetConfigName(appName)
	v.SetConfigType("json")

	// Set default values
	v.SetDefault("allow_inline_html", true)
	v.SetDefault("sanitize_html", true)
	v.SetDefault("strip_h1", true)
	v.SetDefault("alert_callout_style", "GFMPlus")

	// Get configuration directory
	configDir := getConfigDir(appName)
	v.AddConfigPath(configDir)

	configPath := filepath.Join(configDir, appName+".json")

	cm := &ConfigManager{
		config:     &Config{},
		configPath: configPath,
		viper:      v,
	}

	// Load existing configuration
	cm.loadConfig()

	return cm
}

// getConfigDir returns the appropriate configuration directory for the OS
func getConfigDir(appName string) string {
	var configDir string

	if configPath := os.Getenv("XDG_CONFIG_HOME"); configPath != "" {
		// Linux XDG standard
		configDir = filepath.Join(configPath, appName)
	} else if homeDir, err := os.UserHomeDir(); err == nil {
		// Windows, macOS, and Linux fallback
		switch {
		case os.Getenv("OS") == "Windows_NT" || filepath.Separator == '\\':
			configDir = filepath.Join(homeDir, "AppData", "Roaming", appName)
		case os.Getenv("HOME") != "":
			configDir = filepath.Join(homeDir, ".config", appName)
		default:
			configDir = filepath.Join(homeDir, ".config", appName)
		}
	} else {
		// Fallback to current directory
		configDir = "."
	}

	// Ensure directory exists
	os.MkdirAll(configDir, 0755)

	return configDir
}

// loadConfig loads the configuration from file
func (cm *ConfigManager) loadConfig() {
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			fmt.Printf("Config file not found, using defaults: %s\n", cm.configPath)
		} else {
			// Config file found but another error occurred
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	// Unmarshal configuration into struct
	if err := cm.viper.Unmarshal(cm.config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		// Use defaults if unmarshal fails
		cm.config = &Config{
			AllowInlineHTML:   true,
			SanitizeHTML:      true,
			StripH1:           true,
			AlertCalloutStyle: "GFMStrict",
		}
	}
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig updates the configuration
func (cm *ConfigManager) SetConfig(newConfig *Config) {
	cm.config = newConfig
}

// SaveConfig saves the current configuration to file
func (cm *ConfigManager) SaveConfig() error {
	// Update viper with current config values
	cm.viper.Set("allow_inline_html", cm.config.AllowInlineHTML)
	cm.viper.Set("sanitize_html", cm.config.SanitizeHTML)
	cm.viper.Set("strip_h1", cm.config.StripH1)
	cm.viper.Set("alert_callout_style", cm.config.AlertCalloutStyle)

	// Write configuration to file
	if err := cm.viper.WriteConfigAs(cm.configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ApplyCliOverrides applies command-line argument overrides to the configuration
func (cm *ConfigManager) ApplyCliOverrides(allowInlineHTML, sanitizeHTML, stripH1 *bool) {
	if allowInlineHTML != nil {
		cm.config.AllowInlineHTML = *allowInlineHTML
	}
	if sanitizeHTML != nil {
		cm.config.SanitizeHTML = *sanitizeHTML
	}
	if stripH1 != nil {
		cm.config.StripH1 = *stripH1
	}
}

// ValidateAlertCalloutStyle validates and returns a valid alert callout style
func (cm *ConfigManager) ValidateAlertCalloutStyle(style string) string {
	for _, validStyle := range AlertCalloutStyles {
		if style == validStyle {
			return style
		}
	}
	// Return default if invalid
	return "GFMStrict"
}

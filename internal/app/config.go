package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the tabs in the settings dialog
type Config struct {
	Application        ApplicationOptions   `mapstructure:"application" json:"application"`         // Application
	Markdown           MarkdownOptions      `mapstructure:"markdown" json:"markdown"`               // Markdown Features
	AlertCallouts      AlertCalloutOptions  `mapstructure:"alert_callouts" json:"alert_callouts"`   // Alert Callouts
}

// Application-Specific Settings
type ApplicationOptions struct {
	UseInlineHTML       bool    `mapstructure:"use_inline_html" json:"use_inline_html"`             // Inline HTML support (allow inline HTML in Markdown)
	UseSanitize         bool    `mapstructure:"use_sanitize_html" json:"use_sanitize_html"`         // HTML Sanitization (remove unsafe elements and links)
	StripH1             bool    `mapstructure:"strip_h1" json:"strip_h1"`                           // Strip First H1 and Use as Title
	// UseFrontmatter      bool   `mapstructure:"use_frontmatter" json:"use_frontmatter"`             // Parse Frontmatter
	UseFrontmatterTitle bool    `mapstructure:"use_frontmatter_title" json:"use_frontmatter_title"` // Always Use Frontmatter Title if Available
	FontFamily          string  `mapstructure:"font_family" json:"font_family"`                     // Selected font family
	FontSize            float64 `mapstructure:"font_size" json:"font_size"`                         // Selected font size in pixels
	FontFamilyMono      string  `mapstructure:"font_family_mono" json:"font_family_mono"`           // Selected monospace font family
	FontSizeMono        float64 `mapstructure:"font_size_mono" json:"font_size_mono"`               // Selected monospace font size in pixels
	UseAdvancedFontDetection bool `mapstructure:"use_advanced_font_detection" json:"use_advanced_font_detection"` // Use advanced monospace font detection
}

// Markdown-Specific Settings
type MarkdownOptions struct {
	UseGFM             bool   `mapstructure:"use_gfm" json:"use_gfm"`                         // GitHub Flavored Markdown and PHP Markdown Extensions
	UseEmoji           bool   `mapstructure:"use_emoji" json:"use_emoji"`                     // Emoji Support
	UseMermaid         bool   `mapstructure:"use_mermaid" json:"use_mermaid"`                 // Mermaid Diagrams Support
	UseFigure          bool   `mapstructure:"use_figure" json:"use_figure"`                   // Image Figure Wrapping Support
	UseAnchor          bool   `mapstructure:"use_anchor" json:"use_anchor"`                   // Anchor Links on Headings
	UseFences          bool   `mapstructure:"use_fences" json:"use_fences"`                   // Fenced DIVs
	UseSections        bool   `mapstructure:"use_sections" json:"use_sections"`               // Wrap Headings in SECTION Elements
	UseHighlighting    bool   `mapstructure:"use_highlighting" json:"use_highlighting"`       // Fenced Code Highlighting
	UseFancyLists      bool   `mapstructure:"use_fancylists" json:"use_fancylists"`           // Allow Pandoc-Style Fancy Lists
	UseAttributes      bool   `mapstructure:"use_attributes" json:"use_attributes"`           // Allow Custom Attributes (using '{.myclass}' syntax)
	UseTypographic     bool   `mapstructure:"use_typographic" json:"use_typographic"`         // Typographic Extensions to Use Fancy Quotes
}

// Alert Callouts Settings
type AlertCalloutOptions struct {
	UseAlertCallouts   bool   `mapstructure:"use_alertcallouts" json:"use_alertcallouts"`     // GitHub and/or Obsidian Alert/Callouts
	AlertCalloutStyle  string `mapstructure:"alertcallout_style" json:"alertcallout_style"`   // Select Alert Callout Style
}

// ConfigManager handles configuration loading and saving
type ConfigManager struct {
	config     *Config
	configPath string
	viper      *viper.Viper
}

// AlertCalloutStyles defines the available alert callout styles
// (Note to self: Do NOT change "GFMStrict" if you can help it -- this is being used as the fallback default)
var AlertCalloutStyles = map[string]string{
	"GFMStrict":     "GFM Alerts (Standard 5 Alert Types)",
	"GFMWithAliases": "GFM Alerts (GFM + Aliases [e.g., notes = note])",
	"GFMPlus":       "GFM Alerts Plus (GFM + Some Obsidian-Style Callouts)",
	"Obsidian":      "Obsidian-Style (Obsidian Icons and Callout Names)",
}

// getAppNameFromExecutable extracts the application name from the executable path
// without the file extension
func getAppNameFromExecutable() string {
	if len(os.Args) == 0 {
		return "md-reader" // fallback
	}

	// Get the base name from the executable path
	execPath := os.Args[0]

	// Handle both Windows and Unix path separators
	var baseName string
	if strings.Contains(execPath, "\\") {
		// Windows path - handle manually to work cross-platform
		parts := strings.Split(execPath, "\\")
		baseName = parts[len(parts)-1]
	} else {
		// Unix path or no path separators
		baseName = filepath.Base(execPath)
	}

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

	// Set default values for Application section
	v.SetDefault("application.use_inline_html", true)
	v.SetDefault("application.use_sanitize_html", true)
	v.SetDefault("application.strip_h1", true)
	v.SetDefault("application.use_frontmatter_title", true)
	v.SetDefault("application.font_family", "Verdana, Arial, Helvetica, Tahoma, Geneva, sans-serif")
	v.SetDefault("application.font_size", 16.0)
	v.SetDefault("application.font_family_mono", "Consolas, Monaco, DejaVu Sans Mono, Liberation Mono, Courier New, Courier, monospace")
	v.SetDefault("application.font_size_mono", 14.0)
	v.SetDefault("application.use_advanced_font_detection", true)

	// Set default values for Markdown section
	v.SetDefault("markdown.use_gfm", true)
	v.SetDefault("markdown.use_emoji", true)
	v.SetDefault("markdown.use_mermaid", true)
	v.SetDefault("markdown.use_figure", true)
	v.SetDefault("markdown.use_anchor", true)
	v.SetDefault("markdown.use_fences", true)
	v.SetDefault("markdown.use_sections", true)
	v.SetDefault("markdown.use_highlighting", true)
	v.SetDefault("markdown.use_fancylists", true)
	v.SetDefault("markdown.use_attributes", true)
	v.SetDefault("markdown.use_typographic", true)

	// Set default values for Alert Callouts section
	v.SetDefault("alert_callouts.use_alertcallouts", true)
	v.SetDefault("alert_callouts.alertcallout_style", "GFMPlus")

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
			Application: ApplicationOptions{
				UseInlineHTML:  true,
				UseSanitize:    true,
				StripH1:        true,
				UseFrontmatterTitle: true,
				FontFamily:     "Verdana, Arial, Helvetica, Tahoma, Geneva, sans-serif",
				FontSize:       16.0,
				FontFamilyMono: "Consolas, Monaco, DejaVu Sans Mono, Liberation Mono, Courier New, Courier, monospace",
				FontSizeMono:   14.0,
				UseAdvancedFontDetection: true,
			},
			Markdown: MarkdownOptions{
				UseGFM:          true,
				UseEmoji:        true,
				UseMermaid:      true,
				UseFigure:       true,
				UseAnchor:       true,
				UseFences:       true,
				UseSections:     true,
				UseHighlighting: true,
				UseFancyLists:   true,
				UseAttributes:   true,
				UseTypographic:  true,
			},
			AlertCallouts: AlertCalloutOptions{
				UseAlertCallouts:  true,
				AlertCalloutStyle: "GFMPlus",
			},
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
	// Update viper with current config values for Application section
	cm.viper.Set("application.use_inline_html", cm.config.Application.UseInlineHTML)
	cm.viper.Set("application.use_sanitize_html", cm.config.Application.UseSanitize)
	cm.viper.Set("application.strip_h1", cm.config.Application.StripH1)
	cm.viper.Set("application.use_frontmatter_title", cm.config.Application.UseFrontmatterTitle)
	cm.viper.Set("application.font_family", cm.config.Application.FontFamily)
	cm.viper.Set("application.font_size", cm.config.Application.FontSize)
	cm.viper.Set("application.font_family_mono", cm.config.Application.FontFamilyMono)
	cm.viper.Set("application.font_size_mono", cm.config.Application.FontSizeMono)
	cm.viper.Set("application.use_advanced_font_detection", cm.config.Application.UseAdvancedFontDetection)

	// Update viper with current config values for Markdown section
	cm.viper.Set("markdown.use_gfm", cm.config.Markdown.UseGFM)
	cm.viper.Set("markdown.use_emoji", cm.config.Markdown.UseEmoji)
	cm.viper.Set("markdown.use_mermaid", cm.config.Markdown.UseMermaid)
	cm.viper.Set("markdown.use_figure", cm.config.Markdown.UseFigure)
	cm.viper.Set("markdown.use_anchor", cm.config.Markdown.UseAnchor)
	cm.viper.Set("markdown.use_fences", cm.config.Markdown.UseFences)
	cm.viper.Set("markdown.use_sections", cm.config.Markdown.UseSections)
	cm.viper.Set("markdown.use_highlighting", cm.config.Markdown.UseHighlighting)
	cm.viper.Set("markdown.use_fancylists", cm.config.Markdown.UseFancyLists)
	cm.viper.Set("markdown.use_attributes", cm.config.Markdown.UseAttributes)
	cm.viper.Set("markdown.use_typographic", cm.config.Markdown.UseTypographic)

	// Update viper with current config values for Alert Callouts section
	cm.viper.Set("alert_callouts.use_alertcallouts", cm.config.AlertCallouts.UseAlertCallouts)
	cm.viper.Set("alert_callouts.alertcallout_style", cm.config.AlertCallouts.AlertCalloutStyle)

	// Write configuration to file
	if err := cm.viper.WriteConfigAs(cm.configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ApplyCliOverrides applies command-line argument overrides to the configuration
func (cm *ConfigManager) ApplyCliOverrides(allowInlineHTML, sanitizeHTML, stripH1 *bool) {
	if allowInlineHTML != nil {
		cm.config.Application.UseInlineHTML = *allowInlineHTML
	}
	if sanitizeHTML != nil {
		cm.config.Application.UseSanitize = *sanitizeHTML
	}
	if stripH1 != nil {
		cm.config.Application.StripH1 = *stripH1
	}
}

// ValidateAlertCalloutStyle validates and returns a valid alert callout style
func (cm *ConfigManager) ValidateAlertCalloutStyle(style string) string {
	if _, exists := AlertCalloutStyles[style]; exists {
		return style
	}
	// Return default if invalid
	return "GFMPlus"
}

// GetApplicationConfig returns application configuration for markdown processing
func (cm *ConfigManager) GetApplicationConfig() (useInlineHTML, useSanitize bool) {
	return cm.config.Application.UseInlineHTML, cm.config.Application.UseSanitize
}

// GetAlertCalloutConfig returns the alert callout style configuration
func (cm *ConfigManager) GetAlertCalloutConfig() string {
	return cm.config.AlertCallouts.AlertCalloutStyle
}

// Application-specific configuration getters
func (cm *ConfigManager) UseInlineHTML() bool {
	return cm.config.Application.UseInlineHTML
}

func (cm *ConfigManager) UseSanitize() bool {
	return cm.config.Application.UseSanitize
}

func (cm *ConfigManager) StripH1() bool {
	return cm.config.Application.StripH1
}

func (cm *ConfigManager) UseFrontmatterTitle() bool {
	return cm.config.Application.UseFrontmatterTitle
}

// Markdown-specific configuration getters
func (cm *ConfigManager) UseGFM() bool {
	return cm.config.Markdown.UseGFM
}

func (cm *ConfigManager) UseEmoji() bool {
	return cm.config.Markdown.UseEmoji
}

func (cm *ConfigManager) UseMermaid() bool {
	return cm.config.Markdown.UseMermaid
}

func (cm *ConfigManager) UseFigure() bool {
	return cm.config.Markdown.UseFigure
}

func (cm *ConfigManager) UseAnchor() bool {
	return cm.config.Markdown.UseAnchor
}

func (cm *ConfigManager) UseFences() bool {
	return cm.config.Markdown.UseFences
}

func (cm *ConfigManager) UseSections() bool {
	return cm.config.Markdown.UseSections
}

func (cm *ConfigManager) UseHighlighting() bool {
	return cm.config.Markdown.UseHighlighting
}

func (cm *ConfigManager) UseFancyLists() bool {
	return cm.config.Markdown.UseFancyLists
}

func (cm *ConfigManager) UseAttributes() bool {
	return cm.config.Markdown.UseAttributes
}

func (cm *ConfigManager) UseTypographic() bool {
	return cm.config.Markdown.UseTypographic
}

// Alert callouts configuration getters
func (cm *ConfigManager) UseAlertCallouts() bool {
	return cm.config.AlertCallouts.UseAlertCallouts
}

func (cm *ConfigManager) AlertCalloutStyle() string {
	return cm.config.AlertCallouts.AlertCalloutStyle
}

// Font configuration getters and setters
func (cm *ConfigManager) GetFontFamily() string {
	return cm.config.Application.FontFamily
}

func (cm *ConfigManager) SetFontFamily(fontFamily string) {
	cm.config.Application.FontFamily = fontFamily
}

func (cm *ConfigManager) GetFontSize() float64 {
	return cm.config.Application.FontSize
}

func (cm *ConfigManager) SetFontSize(fontSize float64) {
	cm.config.Application.FontSize = fontSize
}

func (cm *ConfigManager) GetFontFamilyMono() string {
	return cm.config.Application.FontFamilyMono
}

func (cm *ConfigManager) SetFontFamilyMono(fontFamily string) {
	cm.config.Application.FontFamilyMono = fontFamily
}

func (cm *ConfigManager) GetFontSizeMono() float64 {
	return cm.config.Application.FontSizeMono
}

func (cm *ConfigManager) SetFontSizeMono(fontSize float64) {
	cm.config.Application.FontSizeMono = fontSize
}

func (cm *ConfigManager) GetUseAdvancedFontDetection() bool {
	return cm.config.Application.UseAdvancedFontDetection
}

func (cm *ConfigManager) SetUseAdvancedFontDetection(useAdvanced bool) {
	cm.config.Application.UseAdvancedFontDetection = useAdvanced
}

package app

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SettingsWindowManager handles the settings functionality
type SettingsWindowManager struct {
	ctx            context.Context
	configManager  *ConfigManager
}

// NewSettingsWindowManager creates a new settings window manager
func NewSettingsWindowManager(ctx context.Context, configManager *ConfigManager) *SettingsWindowManager {
	return &SettingsWindowManager{
		ctx:           ctx,
		configManager: configManager,
	}
}

// OpenSettingsWindow opens the settings in a new window
func (swm *SettingsWindowManager) OpenSettingsWindow() error {
	// For Wails v2, we'll use a different approach - open in a new browser window
	// or use the existing modal but make it resizable and larger

	// Emit event to open settings modal
	runtime.EventsEmit(swm.ctx, "show-settings-window")
	return nil
}

// GetSettings returns the current configuration
func (swm *SettingsWindowManager) GetSettings() *Config {
	return swm.configManager.GetConfig()
}

// SaveSettings saves the provided settings configuration
func (swm *SettingsWindowManager) SaveSettings(settings *Config) error {
	// Validate the alert callout style
	settings.AlertCalloutStyle = swm.configManager.ValidateAlertCalloutStyle(settings.AlertCalloutStyle)

	// Update the configuration
	swm.configManager.SetConfig(settings)

	// Save to file
	if err := swm.configManager.SaveConfig(); err != nil {
		return err
	}

	// Emit event to main window that settings have been updated
	runtime.EventsEmit(swm.ctx, "settings-updated", settings)

	return nil
}

// GetAlertCalloutStyles returns available alert callout styles
func (swm *SettingsWindowManager) GetAlertCalloutStyles() []string {
	return AlertCalloutStyles
}

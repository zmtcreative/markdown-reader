package app

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ThemeManager handles theme-related functionality
type ThemeManager struct {
    ctx   context.Context
    theme string
}

// NewThemeManager creates a new ThemeManager
func NewThemeManager(ctx context.Context) *ThemeManager {
    return &ThemeManager{
        ctx:   ctx,
        theme: "light",
    }
}

// GetTheme returns the current theme
func (tm *ThemeManager) GetTheme() string {
    return tm.theme
}

// SetTheme sets the theme and emits an event to the frontend
func (tm *ThemeManager) SetTheme(theme string) {
    tm.theme = theme
    runtime.EventsEmit(tm.ctx, "theme:changed", theme)
}
//go:build !windows

package app

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// getWindowsFontsFromRegistry is not available on non-Windows platforms
func (fm *FontManager) getWindowsFontsFromRegistry() []string {
	// Return empty slice on non-Windows platforms
	return []string{}
}

// getLinuxSystemFonts attempts to scan common Linux font directories
func (fm *FontManager) getLinuxSystemFonts() []string {
	fonts := make([]string, 0)
	fontMap := make(map[string]bool)

	// Common Linux font directories
	fontDirs := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		"/usr/share/fonts/truetype",
		"/usr/share/fonts/opentype",
		"/usr/share/fonts/TTF",
		"/usr/share/fonts/OTF",
	}

	// Add user font directories if HOME is set
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		userFontDirs := []string{
			filepath.Join(homeDir, ".fonts"),
			filepath.Join(homeDir, ".local/share/fonts"),
		}
		fontDirs = append(fontDirs, userFontDirs...)
	}

	// Scan directories for font files
	for _, dir := range fontDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue scanning
			}

			if d.IsDir() {
				return nil
			}

			// Check for font file extensions
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".ttf" || ext == ".otf" || ext == ".woff" || ext == ".woff2" {
				// Extract potential font name from filename
				basename := filepath.Base(path)
				fontName := strings.TrimSuffix(basename, ext)

				// Clean up the font name
				fontName = strings.ReplaceAll(fontName, "_", " ")
				fontName = strings.ReplaceAll(fontName, "-", " ")
				fontName = cases.Title(language.English, cases.Compact).String(strings.ToLower(fontName))

				// Remove common style suffixes
				styleSuffixes := []string{
					" Regular", " Bold", " Italic", " Light", " Medium", " Heavy",
					" Black", " Thin", " Normal", " Roman", "regular", "bold",
				}
				for _, suffix := range styleSuffixes {
					fontName = strings.TrimSuffix(fontName, suffix)
				}

				if fontName != "" && !fontMap[fontName] {
					fontMap[fontName] = true
					fonts = append(fonts, fontName)
				}
			}
			return nil
		})

		if err != nil {
			// If walking fails, continue with next directory
			continue
		}
	}

	return fonts
}

//go:build windows

package app

import (
	"strings"
	"syscall"
	"unsafe"
)

// getWindowsFontsFromRegistry reads fonts from Windows registry
func (fm *FontManager) getWindowsFontsFromRegistry() []string {
	var fonts []string

	// Define Windows API constants and functions
	const (
		HKEY_LOCAL_MACHINE = 0x80000002
		KEY_READ           = 0x20019
	)

	// Load required DLLs
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	regOpenKeyEx := advapi32.NewProc("RegOpenKeyExW")
	regEnumValue := advapi32.NewProc("RegEnumValueW")
	regCloseKey := advapi32.NewProc("RegCloseKey")

	// Convert registry path to UTF-16
	keyPath, _ := syscall.UTF16PtrFromString("SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Fonts")

	var hKey syscall.Handle
	ret, _, _ := regOpenKeyEx.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(keyPath)),
		0,
		KEY_READ,
		uintptr(unsafe.Pointer(&hKey)),
	)

	if ret != 0 {
		// Registry method failed, return empty slice
		return fonts
	}
	defer regCloseKey.Call(uintptr(hKey))

	// Enumerate registry values
	index := uint32(0)
	for {
		nameBuffer := make([]uint16, 256)
		nameSize := uint32(len(nameBuffer))

		ret, _, _ := regEnumValue.Call(
			uintptr(hKey),
			uintptr(index),
			uintptr(unsafe.Pointer(&nameBuffer[0])),
			uintptr(unsafe.Pointer(&nameSize)),
			0, 0, 0, 0,
		)

		if ret != 0 {
			break // No more entries
		}

		fontName := syscall.UTF16ToString(nameBuffer[:nameSize])

		// Clean up font name (remove file extensions and extra info)
		fontName = strings.TrimSpace(fontName)
		if strings.Contains(fontName, "(") {
			fontName = strings.TrimSpace(strings.Split(fontName, "(")[0])
		}

		if fontName != "" && !strings.Contains(fontName, ".") {
			fonts = append(fonts, fontName)
		}

		index++
	}

	return fonts
}

// getLinuxSystemFonts is not needed on Windows
func (fm *FontManager) getLinuxSystemFonts() []string {
	return []string{}
}

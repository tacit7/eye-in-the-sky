package config

import "os"

// FontMode represents whether to use Nerd Fonts or plain ASCII
type FontMode int

const (
	FontModeNerd FontMode = iota
	FontModePlain
)

// UseNerdFonts is the global font mode setting
var UseNerdFonts = DetectNerdFonts()

// DetectNerdFonts checks if Nerd Fonts should be enabled
// Option 1: Env var toggle (preferred)
// Returns true (Nerd Fonts) by default unless NERD_FONTS=0
func DetectNerdFonts() bool {
	if os.Getenv("NERD_FONTS") == "0" {
		return false
	}
	// Option 2: Config file or CLI flag could also be added later
	return true
}

// GetFontMode returns the current font mode
func GetFontMode() FontMode {
	if UseNerdFonts {
		return FontModeNerd
	}
	return FontModePlain
}

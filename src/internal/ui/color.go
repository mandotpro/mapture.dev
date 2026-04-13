package ui

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ColorMode controls whether terminal styling should be emitted.
type ColorMode string

const (
	// ColorAuto enables ANSI styling only when the target writer is a terminal.
	ColorAuto ColorMode = "auto"
	// ColorAlways forces ANSI styling even for non-terminal writers.
	ColorAlways ColorMode = "always"
	// ColorNever disables ANSI styling entirely.
	ColorNever ColorMode = "never"
)

var isTerminal = term.IsTerminal

// ParseColorMode validates a user-supplied color mode.
func ParseColorMode(value string) (ColorMode, error) {
	switch ColorMode(value) {
	case ColorAuto, ColorAlways, ColorNever:
		return ColorMode(value), nil
	default:
		return ColorAuto, fmt.Errorf("unsupported color mode %q (expected auto, always, or never)", value)
	}
}

// ColorEnabled reports whether styling should be emitted for the writer.
func ColorEnabled(w io.Writer, mode ColorMode) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	}

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	return isTerminal(int(file.Fd()))
}

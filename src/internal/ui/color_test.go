package ui

import (
	"bytes"
	"os"
	"testing"
)

func TestParseColorMode(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"auto", "always", "never"} {
		got, err := ParseColorMode(value)
		if err != nil {
			t.Fatalf("ParseColorMode(%q) returned error: %v", value, err)
		}
		if got != ColorMode(value) {
			t.Fatalf("ParseColorMode(%q) = %q", value, got)
		}
	}

	if _, err := ParseColorMode("loud"); err == nil {
		t.Fatal("expected invalid color mode to fail")
	}
}

func TestColorEnabledModes(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	var buffer bytes.Buffer

	if !ColorEnabled(&buffer, ColorAlways) {
		t.Fatal("expected always mode to enable color")
	}
	if ColorEnabled(&buffer, ColorNever) {
		t.Fatal("expected never mode to disable color")
	}
	if ColorEnabled(&buffer, ColorAuto) {
		t.Fatal("expected auto mode to disable color for non-file writer")
	}
}

func TestColorEnabledRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	originalIsTerminal := isTerminal
	isTerminal = func(int) bool { return true }
	defer func() { isTerminal = originalIsTerminal }()

	if ColorEnabled(os.Stdout, ColorAuto) {
		t.Fatal("expected NO_COLOR to disable color")
	}
}

func TestColorEnabledAutoUsesTerminalDetection(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	originalIsTerminal := isTerminal
	isTerminal = func(int) bool { return true }
	defer func() { isTerminal = originalIsTerminal }()

	if !ColorEnabled(os.Stdout, ColorAuto) {
		t.Fatal("expected auto mode to use terminal detection")
	}
}

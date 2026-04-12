// Package ui provides shared CLI reporting primitives for rich and
// plain-text terminal output.
package ui

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Console renders shared human-facing terminal output.
type Console struct {
	color  bool
	styles consoleStyles
}

type consoleStyles struct {
	brand   lipgloss.Style
	strong  lipgloss.Style
	muted   lipgloss.Style
	accent  lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
	code    lipgloss.Style
}

// NewConsole creates a shared terminal renderer for human-facing command output.
func NewConsole(primary io.Writer, peers ...io.Writer) *Console {
	color := SupportsColor(primary)
	for _, peer := range peers {
		color = color && SupportsColor(peer)
	}

	base := lipgloss.NewStyle()
	if !color {
		return &Console{
			color: color,
			styles: consoleStyles{
				brand:   base.Bold(true),
				strong:  base.Bold(true),
				muted:   base,
				accent:  base.Bold(true),
				success: base.Bold(true),
				warning: base.Bold(true),
				error:   base.Bold(true),
				code:    base.Bold(true),
			},
		}
	}

	return &Console{
		color: color,
		styles: consoleStyles{
			brand:   lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true),
			strong:  lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
			muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
			accent:  lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true),
			success: lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
			warning: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
			error:   lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
			code:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		},
	}
}

// SupportsColor reports whether styled terminal output should be used.
func SupportsColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("CI") != "" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

// ColorEnabled reports whether ANSI styling is active for this console.
func (c *Console) ColorEnabled() bool {
	return c != nil && c.color
}

// Brand renders branded product text.
func (c *Console) Brand(text string) string {
	return c.styles.brand.Render(text)
}

// Strong renders high-emphasis text.
func (c *Console) Strong(text string) string {
	return c.styles.strong.Render(text)
}

// Muted renders low-emphasis metadata text.
func (c *Console) Muted(text string) string {
	return c.styles.muted.Render(text)
}

// Accent renders informational highlight text.
func (c *Console) Accent(text string) string {
	return c.styles.accent.Render(text)
}

// Success renders success state text.
func (c *Console) Success(text string) string {
	return c.styles.success.Render(text)
}

// Warning renders warning state text.
func (c *Console) Warning(text string) string {
	return c.styles.warning.Render(text)
}

// Error renders error state text.
func (c *Console) Error(text string) string {
	return c.styles.error.Render(text)
}

// Code renders command or identifier text with emphasis.
func (c *Console) Code(text string) string {
	return c.styles.code.Render(text)
}

// Join combines metadata fragments with a shared separator.
func (c *Console) Join(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) == 0 {
		return ""
	}
	separator := " • "
	if c.ColorEnabled() {
		separator = c.Muted(separator)
	}
	return strings.Join(filtered, separator)
}

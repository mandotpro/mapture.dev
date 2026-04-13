package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type iconSet struct {
	stage   string
	success string
	warn    string
	err     string
	info    string
}

type styles struct {
	brand   textStyle
	stage   textStyle
	success textStyle
	warning textStyle
	error   textStyle
	path    textStyle
	muted   textStyle
	summary textStyle
	code    textStyle
	accent  textStyle
	strong  textStyle
}

type textStyle struct {
	start string
	end   string
}

func ansiStyle(codes ...string) textStyle {
	if len(codes) == 0 {
		return textStyle{}
	}
	return textStyle{
		start: "\x1b[" + strings.Join(codes, ";") + "m",
		end:   "\x1b[0m",
	}
}

func (s textStyle) Render(text string) string {
	if s.start == "" || text == "" {
		return text
	}
	return s.start + text + s.end
}

// Console renders consistent human-facing CLI output.
type Console struct {
	out    io.Writer
	color  bool
	icons  iconSet
	styles styles
}

// NewConsole creates a console that respects the requested color mode.
func NewConsole(out io.Writer, mode ColorMode) *Console {
	color := ColorEnabled(out, mode)
	icons, styles := buildTheme(color)
	return &Console{
		out:    out,
		color:  color,
		icons:  icons,
		styles: styles,
	}
}

func buildTheme(color bool) (iconSet, styles) {
	icons := iconSet{
		stage:   "›",
		success: "✓",
		warn:    "!",
		err:     "x",
		info:    "·",
	}
	if !color {
		return iconSet{
				stage:   "[..]",
				success: "[ok]",
				warn:    "[!]",
				err:     "[x]",
				info:    " - ",
			}, styles{
				brand:   textStyle{},
				stage:   textStyle{},
				success: textStyle{},
				warning: textStyle{},
				error:   textStyle{},
				path:    textStyle{},
				muted:   textStyle{},
				summary: textStyle{},
				code:    textStyle{},
				accent:  textStyle{},
				strong:  textStyle{},
			}
	}

	return icons, styles{
		brand:   ansiStyle("1", "38;5;87"),
		stage:   ansiStyle("1", "38;5;69"),
		success: ansiStyle("1", "38;5;42"),
		warning: ansiStyle("1", "38;5;214"),
		error:   ansiStyle("1", "38;5;203"),
		path:    ansiStyle("4", "38;5;111"),
		muted:   ansiStyle("38;5;244"),
		summary: ansiStyle("1"),
		code:    ansiStyle("1", "38;5;252"),
		accent:  ansiStyle("1", "38;5;81"),
		strong:  ansiStyle("1", "38;5;255"),
	}
}

// Brand renders branded product text.
func (c *Console) Brand(text string) string { return c.styles.brand.Render(text) }

// Strong renders high-emphasis text.
func (c *Console) Strong(text string) string { return c.styles.strong.Render(text) }

// Accent renders informational highlight text.
func (c *Console) Accent(text string) string { return c.styles.accent.Render(text) }

// Muted renders low-emphasis metadata text.
func (c *Console) Muted(text string) string { return c.styles.muted.Render(text) }

// Code renders command or identifier text with emphasis.
func (c *Console) Code(text string) string { return c.styles.code.Render(text) }

// Path renders a filesystem path consistently across CLI output.
func (c *Console) Path(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return c.styles.path.Render(filepath.ToSlash(text))
}

// Join combines metadata fragments with a shared separator.
func (c *Console) Join(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, " • ")
}

// Header renders the branded CLI header block.
func (c *Console) Header(title string, details ...string) string {
	lines := []string{fmt.Sprintf("%s - %s", c.Brand("mapture.dev"), c.Strong(title))}
	for _, detail := range details {
		if strings.TrimSpace(detail) == "" {
			continue
		}
		lines = append(lines, c.Muted(detail))
	}
	return strings.Join(lines, "\n")
}

// Stage prints an informational stage line.
func (c *Console) Stage(label string, details string) error {
	return c.writeStatus(c.styles.stage.Render(c.icons.stage), label, details)
}

// Success prints a success line.
func (c *Console) Success(label string, details string) error {
	return c.writeStatus(c.styles.success.Render(c.icons.success), label, details)
}

// Warning prints a warning line.
func (c *Console) Warning(label string, details string) error {
	return c.writeStatus(c.styles.warning.Render(c.icons.warn), label, details)
}

// Error prints an error line.
func (c *Console) Error(label string, details string) error {
	return c.writeStatus(c.styles.error.Render(c.icons.err), label, details)
}

// Info prints a low-severity informational line.
func (c *Console) Info(label string, details string) error {
	return c.writeStatus(c.styles.accent.Render(c.icons.info), label, details)
}

// Printf writes formatted text to the configured output writer.
func (c *Console) Printf(format string, args ...any) error {
	if c.out == nil {
		return nil
	}
	_, err := fmt.Fprintf(c.out, format, args...)
	return err
}

// Println writes a full line to the configured output writer.
func (c *Console) Println(text string) error {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return c.Printf("%s", text)
}

func (c *Console) writeStatus(prefix string, label string, details string) error {
	text := fmt.Sprintf("%s %s", prefix, label)
	if details != "" {
		text += " " + c.Muted(details)
	}
	return c.Println(text)
}

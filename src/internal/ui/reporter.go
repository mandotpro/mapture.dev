// Package ui provides shared CLI reporting primitives for rich and
// plain-text terminal output.
package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

// Reporter renders consistent CLI output for commands.
type Reporter struct {
	out    io.Writer
	color  bool
	icons  iconSet
	styles styles
}

// NewReporter creates a reporter that auto-detects whether rich styling
// should be enabled.
func NewReporter(out, errOut io.Writer, mode ...ColorMode) *Reporter {
	selectedMode := ColorAuto
	if len(mode) > 0 {
		selectedMode = mode[0]
	}
	color := ColorEnabled(out, selectedMode) && ColorEnabled(errOut, selectedMode)
	icons, styles := buildTheme(color)
	return &Reporter{
		out:    out,
		color:  color,
		icons:  icons,
		styles: styles,
	}
}

// Stage prints a step header.
func (r *Reporter) Stage(label string, details string) error {
	text := fmt.Sprintf("%s %s", r.styles.stage.Render(r.icons.stage), label)
	if details != "" {
		text += " " + r.styles.muted.Render(details)
	}
	return r.writeLine(r.out, text)
}

// Success prints a success line.
func (r *Reporter) Success(label string, details string) error {
	text := fmt.Sprintf("%s %s", r.styles.success.Render(r.icons.success), label)
	if details != "" {
		text += " " + r.styles.muted.Render(details)
	}
	return r.writeLine(r.out, text)
}

// Diagnostics prints grouped warnings and errors.
func (r *Reporter) Diagnostics(diagnostics []validator.Diagnostic) error {
	if len(diagnostics) == 0 {
		return nil
	}

	errors := make([]validator.Diagnostic, 0)
	warnings := make([]validator.Diagnostic, 0)
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case "error":
			errors = append(errors, diagnostic)
		case "warning":
			warnings = append(warnings, diagnostic)
		}
	}

	if len(errors) > 0 {
		if err := r.writeLine(r.out, fmt.Sprintf("%s %s", r.styles.error.Render(r.icons.err), r.styles.error.Render("Errors"))); err != nil {
			return err
		}
		for _, diagnostic := range errors {
			if err := r.writeDiagnostic(r.out, diagnostic); err != nil {
				return err
			}
		}
	}
	if len(warnings) > 0 {
		if err := r.writeLine(r.out, fmt.Sprintf("%s %s", r.styles.warning.Render(r.icons.warn), r.styles.warning.Render("Warnings"))); err != nil {
			return err
		}
		for _, diagnostic := range warnings {
			if err := r.writeDiagnostic(r.out, diagnostic); err != nil {
				return err
			}
		}
	}
	return nil
}

// Summary prints the final result line.
func (r *Reporter) Summary(ok bool, errors int, warnings int, blocks int, nodes int, edges int) error {
	label := "Validation Failed"
	icon := r.styles.error.Render(r.icons.err)
	if ok {
		label = "Validation Succeeded"
		icon = r.styles.success.Render(r.icons.success)
	}

	text := fmt.Sprintf(
		"%s %s: %d error(s), %d warning(s), %d block(s), %d node(s), %d edge(s)",
		icon,
		r.styles.summary.Render(label),
		errors,
		warnings,
		blocks,
		nodes,
		edges,
	)
	return r.writeLine(r.out, text)
}

func (r *Reporter) writeDiagnostic(w io.Writer, diagnostic validator.Diagnostic) error {
	location := ""
	if diagnostic.File != "" {
		location = r.styles.path.Render(filepath.ToSlash(diagnostic.File))
		if diagnostic.Line > 0 {
			location += r.styles.muted.Render(fmt.Sprintf(":%d", diagnostic.Line))
		}
	}
	prefix := fmt.Sprintf("  %s layer %d %s", r.icons.info, diagnostic.Layer, r.styles.code.Render(diagnostic.Code))
	if location != "" {
		prefix += " " + location
	}
	return r.writeLine(w, prefix+" "+diagnostic.Message)
}

func (r *Reporter) writeLine(w io.Writer, text string) error {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	_, err := io.WriteString(w, text)
	return err
}

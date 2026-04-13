package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

func TestReporterDiagnosticsPlainText(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	reporter := NewReporter(&stdout, &stderr)
	reporter.color = false
	reporter.icons = iconSet{
		stage:   "[..]",
		success: "[ok]",
		warn:    "[!]",
		err:     "[x]",
		info:    " - ",
	}
	_, reporter.styles = buildTheme(false)

	diagnostics := []validator.Diagnostic{
		{Severity: "error", Layer: 4, Code: "unknown_domain", File: "src/app.go", Line: 3, Message: `unknown domain "missing"`},
		{Severity: "warning", Layer: 6, Code: "orphaned_node", File: "src/app.go", Line: 8, Message: `node "service:demo" has no edges`},
	}

	if err := reporter.Diagnostics(diagnostics); err != nil {
		t.Fatalf("Diagnostics returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"Errors", "Warnings", "src/app.go:3", "unknown_domain", "orphaned_node"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

func TestReporterSummaryPlainText(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	reporter := NewReporter(&stdout, &stderr)
	reporter.color = false
	reporter.icons = iconSet{
		stage:   "[..]",
		success: "[ok]",
		warn:    "[!]",
		err:     "[x]",
		info:    " - ",
	}
	_, reporter.styles = buildTheme(false)

	if err := reporter.Summary(true, 0, 1, 5, 4, 4); err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Validation Succeeded: 0 error(s), 1 warning(s), 5 block(s), 4 node(s), 4 edge(s)") {
		t.Fatalf("unexpected summary output: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/config"
)

func TestLoadProjectSuccess(t *testing.T) {
	t.Parallel()

	configPath, cfg, cat, err := loadProject("../../examples/demo")
	if err != nil {
		t.Fatalf("loadProject returned error: %v", err)
	}
	if !strings.HasSuffix(configPath, "examples/demo/mapture.yaml") {
		t.Fatalf("unexpected config path: %s", configPath)
	}
	if cfg.Catalog.Dir != "./architecture" {
		t.Fatalf("unexpected catalog dir: %s", cfg.Catalog.Dir)
	}
	if len(cat.Teams) != 2 || len(cat.Domains) != 2 || len(cat.Events) != 1 {
		t.Fatalf("unexpected catalog sizes: teams=%d domains=%d events=%d", len(cat.Teams), len(cat.Domains), len(cat.Events))
	}
}

func TestLoadProjectRejectsBrokenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "bad config role", path: "../../examples/invalid/bad-config-role", wantErr: "random"},
		{name: "duplicate team", path: "../../examples/invalid/duplicate-team", wantErr: "duplicate team id"},
		{name: "unknown domain owner", path: "../../examples/invalid/unknown-domain-owner", wantErr: "unknown team"},
		{name: "invalid event status", path: "../../examples/invalid/invalid-event-status", wantErr: "random"},
		{name: "missing teams file", path: "../../examples/invalid/missing-teams-file", wantErr: "read"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, err := loadProject(tc.path)
			if err == nil {
				t.Fatalf("expected error for %s", tc.path)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q in error, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateProjectRejectsValidationFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "unknown comment domain", path: "../../examples/invalid/comment-unknown-domain-ref", wantErr: "unknown domain"},
		{name: "event domain mismatch", path: "../../examples/invalid/comment-event-domain-mismatch", wantErr: "event_domain_mismatch"},
		{name: "unknown node target", path: "../../examples/invalid/comment-unknown-node-target", wantErr: "unknown_node_target"},
		{name: "missing owner", path: "../../examples/invalid/comment-missing-owner", wantErr: "arch.owner"},
		{name: "bad event role", path: "../../examples/invalid/comment-bad-event-role", wantErr: "event.role"},
		{name: "unknown key", path: "../../examples/invalid/comment-unknown-key", wantErr: "arch.foobar"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			configPath, cfg, cat, err := loadProject(tc.path)
			if err != nil {
				t.Fatalf("loadProject returned error: %v", err)
			}

			_, _, err = validateProject(filepath.Dir(configPath), cfg, cat)
			if err == nil {
				t.Fatalf("expected validateProject error for %s", tc.path)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q in error, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateProjectSuccess(t *testing.T) {
	t.Parallel()

	configPath, _, cat, err := loadProject("../../examples/demo")
	if err != nil {
		t.Fatalf("loadProject returned error: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}

	blocks, result, err := validateProject("../../examples/demo", cfg, cat)
	if err != nil {
		t.Fatalf("validateProject returned error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected scanned blocks")
	}
	if len(result.Graph.Nodes) == 0 || len(result.Graph.Edges) == 0 {
		t.Fatalf("expected non-empty graph, got %#v", result.Graph)
	}
}

func TestRunValidateOutputsRichSummary(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runValidate("../../examples/demo", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runValidate returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"Resolving project",
		"Loading config",
		"Loading catalogs",
		"Scanning sources",
		"Building graph",
		"Validation Succeeded:",
		"Validation complete",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunValidateOutputsDiagnosticsOnFailure(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runValidate("../../examples/invalid/comment-unknown-node-target", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected runValidate error")
	}

	errorOutput := stderr.String()
	if errorOutput != "" {
		t.Fatalf("expected empty stderr, got %q", errorOutput)
	}
	errorOutput = stdout.String()
	for _, want := range []string{
		"Errors",
		"unknown_node_target",
		"Validation Failed:",
	} {
		if !strings.Contains(errorOutput, want) {
			t.Fatalf("expected %q in stderr, got %q", want, errorOutput)
		}
	}
}

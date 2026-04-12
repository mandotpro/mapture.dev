package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/config"
	exportercanonical "github.com/mandotpro/mapture.dev/src/internal/exporter/canonical"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
	"github.com/mandotpro/mapture.dev/src/internal/updater"
)

func TestLoadProjectSuccess(t *testing.T) {
	t.Parallel()

	configPath, _, cat, err := loadProject("../../examples/demo")
	if err != nil {
		t.Fatalf("loadProject returned error: %v", err)
	}
	if !strings.HasSuffix(configPath, "examples/demo/mapture.yaml") {
		t.Fatalf("unexpected config path: %s", configPath)
	}
	if len(cat.Teams) != 2 || len(cat.Domains) != 2 {
		t.Fatalf("unexpected catalog sizes: teams=%d domains=%d", len(cat.Teams), len(cat.Domains))
	}
}

func TestResolveVersionPrefersInjectedReleaseVersion(t *testing.T) {
	t.Parallel()

	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "v0.9.9",
		},
	}

	got := resolveVersion("v1.2.3", info)
	if got != "v1.2.3" {
		t.Fatalf("expected injected version, got %q", got)
	}
}

func TestResolveVersionKeepsInjectedDevVersion(t *testing.T) {
	t.Parallel()

	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "v0.0.0-20260411-abcdef",
		},
	}

	got := resolveVersion("0.0.0-dev", info)
	if got != "0.0.0-dev" {
		t.Fatalf("expected injected dev version, got %q", got)
	}
}

func TestResolveVersionUsesModuleVersionForSourceInstalls(t *testing.T) {
	t.Parallel()

	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "v0.0.0-20260411-abcdef",
		},
	}

	got := resolveVersion("", info)
	if got != "v0.0.0-20260411-abcdef" {
		t.Fatalf("expected build info version, got %q", got)
	}
}

func TestResolveVersionFallsBackToRevisionWhenVersionMissing(t *testing.T) {
	t.Parallel()

	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "(devel)",
		},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890abcdef"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	got := resolveVersion("", info)
	if got != "0.0.0-dev+dirty.1234567" {
		t.Fatalf("expected dirty revision fallback, got %q", got)
	}
}

func TestUpdateCommandPassesThroughChannel(t *testing.T) {
	t.Parallel()

	original := runUpdateCmd
	defer func() {
		runUpdateCmd = original
	}()

	called := false
	runUpdateCmd = func(_ context.Context, opts updater.Options) error {
		called = true
		if opts.RequestedChannel != updater.ChannelCanary {
			t.Fatalf("RequestedChannel = %q, want %q", opts.RequestedChannel, updater.ChannelCanary)
		}
		return nil
	}

	cmd := newUpdateCmd()
	cmd.SetArgs([]string{"--channel", "canary"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !called {
		t.Fatal("expected updater to be called")
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

			_, _, err = validateProject(filepath.Dir(configPath), cfg, cat, nil)
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

	blocks, result, err := validateProject("../../examples/demo", cfg, cat, nil)
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

func TestValidateProjectScopedSuccessWithBoundaryNodes(t *testing.T) {
	t.Parallel()

	configPath, cfg, cat, err := loadProject("../../examples/ecommerce")
	if err != nil {
		t.Fatalf("loadProject returned error: %v", err)
	}

	_, result, err := validateProject(filepath.Dir(configPath), cfg, cat, []string{"./src/php/orders"})
	if err != nil {
		t.Fatalf("validateProject returned error: %v", err)
	}

	foundBoundary := false
	for _, node := range result.Graph.Nodes {
		if node.ID == "api:payment-api" && node.File == "" && strings.Contains(node.Summary, "out-of-scope") {
			foundBoundary = true
		}
	}
	if !foundBoundary {
		t.Fatalf("expected scoped graph to synthesize api:payment-api boundary node, got %#v", result.Graph.Nodes)
	}
}

func TestScanCommandRespectsScope(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	cmd := newScanCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"../../examples/ecommerce", "--scope", "./src/php/orders"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "src/php/orders/CheckoutService.php") {
		t.Fatalf("expected scoped scan output to include php orders file, got %q", output)
	}
	if strings.Contains(output, "src/ts/shipping/ShippingService.ts") {
		t.Fatalf("expected scoped scan output to exclude unrelated files, got %q", output)
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
		"Loading catalog metadata",
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

func TestGraphCommandWritesMermaidToStdout(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	commandStdout = &stdout
	t.Cleanup(func() { commandStdout = os.Stdout })

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"../../examples/demo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"flowchart LR", "Checkout Service", "|calls|", "Order Placed"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

func TestGraphCommandWritesMermaidFile(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "ecommerce.mmd")

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"../../examples/ecommerce", "-o", outputPath, "--domain", "billing"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	output := string(data)
	for _, want := range []string{"flowchart LR", "Billing", "Payment Service"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in file output, got %q", want, output)
		}
	}
}

func TestExportJSONCommandWritesCanonicalDocument(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	cmd := newExportJSONCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"../../examples/demo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.Bytes()
	if err := schema.ValidateJSON(schema.CanonicalDefinition, "stdout.json", output); err != nil {
		t.Fatalf("canonical schema validation failed: %v", err)
	}

	var doc exportercanonical.Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if doc.SchemaVersion != exportercanonical.SchemaVersion {
		t.Fatalf("unexpected schema version: %d", doc.SchemaVersion)
	}
	if doc.Source.ConfigPath == "" || doc.Source.ProjectRoot == "" {
		t.Fatalf("expected source metadata, got %+v", doc.Source)
	}
	if len(doc.Graph.Nodes) == 0 || len(doc.Catalog.Teams) == 0 || len(doc.Validation.Diagnostics) != 0 {
		t.Fatalf("unexpected export payload: %+v", doc)
	}
}

func TestReportServeErrorIncludesPortBusyHint(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	err := fmt.Errorf("listen %s: %w", "127.0.0.1:8765", syscall.EADDRINUSE)

	reportServeError(&stderr, "127.0.0.1:8765", err)

	output := stderr.String()
	for _, want := range []string{
		"mapture serve: listen 127.0.0.1:8765",
		"already in use",
		"Ctrl-Z",
		"kill %<job>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

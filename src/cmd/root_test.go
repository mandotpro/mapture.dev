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
	exporterjgf "github.com/mandotpro/mapture.dev/src/internal/exporter/jgf"
	exportervis "github.com/mandotpro/mapture.dev/src/internal/exporter/visualization"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
	"github.com/mandotpro/mapture.dev/src/internal/ui"
	"github.com/mandotpro/mapture.dev/src/internal/updater"
)

func resetRootFlags() {
	colorModeFlag = string(ui.ColorAuto)
	noColorFlag = false
	versionFlag = false
	if flag := rootCmd.PersistentFlags().Lookup("color"); flag != nil {
		_ = flag.Value.Set(string(ui.ColorAuto))
		flag.Changed = false
	}
	if flag := rootCmd.PersistentFlags().Lookup("no-color"); flag != nil {
		_ = flag.Value.Set("false")
		flag.Changed = false
	}
	if flag := rootCmd.Flags().Lookup("version"); flag != nil {
		_ = flag.Value.Set("false")
		flag.Changed = false
	}
}

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

func TestVersionCommandShowsBrandedRuntimeInfo(t *testing.T) {
	originalInspect := inspectRuntime
	originalCheck := checkVersionStatus
	defer func() {
		inspectRuntime = originalInspect
		checkVersionStatus = originalCheck
	}()

	inspectRuntime = func(_ string, _ *debug.BuildInfo) (updater.RuntimeDetails, error) {
		return updater.RuntimeDetails{
			Version:        "0.0.0-canary.20260412+sha.1bd3598",
			Channel:        updater.ChannelCanary,
			InstallMethod:  "homebrew",
			ExecutablePath: "/opt/homebrew/bin/mapture",
		}, nil
	}
	checkVersionStatus = func(_ context.Context, _ string, _ *debug.BuildInfo) (updater.VersionStatus, error) {
		return updater.VersionStatus{
			UpdateAvailable:  true,
			LatestForChannel: "0.0.0-canary.20260413+sha.abcdef0",
		}, nil
	}

	var stdout bytes.Buffer
	cmd := newVersionCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"mapture.dev - 0.0.0-canary.20260412+sha.1bd3598",
		"canary",
		"homebrew",
		"/opt/homebrew/bin/mapture",
		"Repo-native architecture mapping that stays close to the code.",
		"Update available",
		"Run: mapture update",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

func TestRootHelpShowsBrandedHeader(t *testing.T) {
	originalInspect := inspectRuntime
	originalCheck := checkVersionStatus
	previousOut := rootCmd.OutOrStdout()
	previousErr := rootCmd.ErrOrStderr()
	defer func() {
		inspectRuntime = originalInspect
		checkVersionStatus = originalCheck
		rootCmd.SetOut(previousOut)
		rootCmd.SetErr(previousErr)
		rootCmd.SetArgs(nil)
		resetRootFlags()
	}()

	inspectRuntime = func(_ string, _ *debug.BuildInfo) (updater.RuntimeDetails, error) {
		return updater.RuntimeDetails{
			Version:        "v0.3.0",
			Channel:        updater.ChannelStable,
			InstallMethod:  "direct binary",
			ExecutablePath: "/usr/local/bin/mapture",
		}, nil
	}
	checkVersionStatus = func(_ context.Context, _ string, _ *debug.BuildInfo) (updater.VersionStatus, error) {
		return updater.VersionStatus{}, nil
	}

	var stdout bytes.Buffer
	resetRootFlags()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"mapture.dev - v0.3.0",
		"stable",
		"direct binary",
		"Usage:",
		"Available Commands:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

func TestRootHelpHonorsForcedColor(t *testing.T) {
	var stdout bytes.Buffer
	previousOut := rootCmd.OutOrStdout()
	previousErr := rootCmd.ErrOrStderr()
	defer func() {
		rootCmd.SetOut(previousOut)
		rootCmd.SetErr(previousErr)
		rootCmd.SetArgs(nil)
		resetRootFlags()
	}()

	resetRootFlags()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)
	rootCmd.SetArgs([]string{"--help", "--color=always"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("expected ANSI output, got %q", stdout.String())
	}
}

func TestRootHelpHonorsNoColor(t *testing.T) {
	var stdout bytes.Buffer
	previousOut := rootCmd.OutOrStdout()
	previousErr := rootCmd.ErrOrStderr()
	defer func() {
		rootCmd.SetOut(previousOut)
		rootCmd.SetErr(previousErr)
		rootCmd.SetArgs(nil)
		resetRootFlags()
	}()

	resetRootFlags()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)
	rootCmd.SetArgs([]string{"--help", "--no-color"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("expected plain output, got %q", stdout.String())
	}
}

func TestRootHelpRejectsConflictingColorFlags(t *testing.T) {
	defer func() {
		rootCmd.SetArgs(nil)
		resetRootFlags()
	}()

	resetRootFlags()
	rootCmd.SetArgs([]string{"--color=always", "--no-color"})
	err := rootCmd.ParseFlags([]string{"--color=always", "--no-color"})
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}
	_, err = selectedColorMode(rootCmd)
	if err == nil {
		t.Fatal("expected conflicting color flags to fail")
	}
	if !strings.Contains(err.Error(), "--color and --no-color cannot be used together") {
		t.Fatalf("unexpected error: %v", err)
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

func TestExportJSONGraphCommandWritesJGFDocument(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	cmd := newExportJSONGraphCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"../../examples/demo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.Bytes()
	if err := schema.ValidateJSON(schema.JSONGraphDefinition, "stdout.json", output); err != nil {
		t.Fatalf("json graph schema validation failed: %v", err)
	}

	var doc exporterjgf.Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if doc.Graph.Metadata.Mapture.SchemaVersion != exporterjgf.SchemaVersion {
		t.Fatalf("unexpected schema version: %d", doc.Graph.Metadata.Mapture.SchemaVersion)
	}
	if doc.Graph.Metadata.Mapture.Source.ConfigPath == "" || doc.Graph.Metadata.Mapture.Source.ProjectRoot == "" {
		t.Fatalf("expected source metadata, got %+v", doc.Graph.Metadata.Mapture.Source)
	}
	if len(doc.Graph.Nodes) == 0 || len(doc.Graph.Metadata.Mapture.Catalog.Teams) == 0 || len(doc.Graph.Metadata.Mapture.Validation.Diagnostics) != 0 {
		t.Fatalf("unexpected export payload: %+v", doc)
	}
}

func TestExportJSONVisualizationCommandWritesVisualizationDocument(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	cmd := newExportJSONVisualizationCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"../../examples/demo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.Bytes()
	if err := schema.ValidateJSON(schema.VisualizationDefinition, "stdout.json", output); err != nil {
		t.Fatalf("visualization schema validation failed: %v", err)
	}

	var doc exportervis.Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if doc.SchemaVersion != exportervis.SchemaVersion {
		t.Fatalf("unexpected schema version: %d", doc.SchemaVersion)
	}
	if doc.Source.ConfigPath == "" || doc.Source.ProjectRoot == "" {
		t.Fatalf("expected source metadata, got %+v", doc.Source)
	}
	if len(doc.Graph.Nodes) == 0 || len(doc.Catalog.Teams) == 0 {
		t.Fatalf("unexpected export payload: %+v", doc)
	}
}

func TestExportHTMLCommandWritesStaticBundle(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "bundle")
	cmd := newExportHTMLCmd()
	cmd.SetArgs([]string{"../../examples/demo", "-o", outputDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	for _, name := range []string{"index.html", "app.js", "styles.css", "data.json"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Fatalf("expected %s in bundle: %v", name, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "data.json"))
	if err != nil {
		t.Fatalf("ReadFile(data.json): %v", err)
	}
	if err := schema.ValidateJSON(schema.VisualizationDefinition, "data.json", data); err != nil {
		t.Fatalf("visualization schema validation failed: %v", err)
	}
}

func TestRootHelpListsNewExportCommandsOnly(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"export-json-graph", "export-json-visualisation", "export-html", "export-ai"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in help output, got %q", want, output)
		}
	}
	for _, unwanted := range []string{"\ngraph ", "\nexport-json "} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("did not expect %q in help output, got %q", unwanted, output)
		}
	}
}

func TestReportServeErrorIncludesPortBusyHint(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	err := fmt.Errorf("listen %s: %w", "127.0.0.1:8765", syscall.EADDRINUSE)

	reportServeError(&stderr, "127.0.0.1:8765", err, ui.ColorNever)

	output := stderr.String()
	for _, want := range []string{
		"Serve failed",
		"already in use",
		"Ctrl-Z",
		"kill %<job>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

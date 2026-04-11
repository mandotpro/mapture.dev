// Package cmd wires up the Mapture CLI.
//
// Command surface matches the planned CLI. v0.1 implementations are
// stubs that print a TODO banner; real logic lands in src/internal/*.
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/mandotpro/mapture.dev/src/internal/bootstrap"
	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	exportermermaid "github.com/mandotpro/mapture.dev/src/internal/exporter/mermaid"
	"github.com/mandotpro/mapture.dev/src/internal/projectscope"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/server"
	"github.com/mandotpro/mapture.dev/src/internal/ui"
	"github.com/mandotpro/mapture.dev/src/internal/updater"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags. When a binary is
// installed directly with `go install module/path@version`, no project
// ldflags are applied, so we fall back to Go build metadata.
var version string

var (
	commandStdout io.Writer = os.Stdout
	commandStderr io.Writer = os.Stderr
	runUpdateCmd            = updater.Run
)

var rootCmd = &cobra.Command{
	Use:           "mapture",
	Short:         "Repo-native architecture graph tool",
	Long:          "Mapture turns catalog YAML and structured code comments into validated architecture graphs, diagrams, and AI-ready bundles.",
	Version:       version,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	version = resolveVersion(version, readBuildInfo())
	rootCmd.Version = version
	rootCmd.AddCommand(
		newInitCmd(),
		newValidateCmd(),
		newScanCmd(),
		newGraphCmd(),
		newServeCmd(),
		newUpdateCmd(),
		newExportHTMLCmd(),
		newExportAICmd(),
	)
}

var readBuildInfo = func() *debug.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return info
}

func resolveVersion(injected string, info *debug.BuildInfo) string {
	const devVersion = "0.0.0-dev"

	if injected != "" {
		return injected
	}
	if info == nil {
		return devVersion
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	var revision string
	var modified bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	if revision == "" {
		return devVersion
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if modified {
		return fmt.Sprintf("%s+dirty.%s", devVersion, revision)
	}
	return fmt.Sprintf("%s+sha.%s", devVersion, revision)
}

// todo is a placeholder body used while v0.1 commands are scaffolded.
// Each caller should be replaced by a real implementation in the
// matching src/internal/* package.
func todo(name string) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		if _, err := fmt.Fprintf(os.Stderr, "mapture %s: not implemented yet (target=%s)\n", name, path); err != nil {
			return err
		}
		return nil
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "Bootstrap a starter mapture.yaml for the target repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return bootstrap.Run(path, os.Stdin, os.Stdout, os.Stderr)
		},
	}
}

func newValidateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate config, catalogs, comments, and graph references",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			scopes, err := cmd.Flags().GetStringSlice("scope")
			if err != nil {
				return err
			}
			return runValidateWithScopes(target, scopes, commandStdout, commandStderr)
		},
	}
	bindScopeFlag(c)
	return c
}

func runValidate(target string, stdout, stderr io.Writer) error {
	return runValidateWithScopes(target, nil, stdout, stderr)
}

func runValidateWithScopes(target string, scopes []string, stdout, stderr io.Writer) error {
	reporter := ui.NewReporter(stdout, stderr)

	if err := reporter.Stage("Resolving project", target); err != nil {
		return err
	}
	configPath, err := config.Discover(target)
	if err != nil {
		return err
	}
	if err := reporter.Success("Loaded config path", filepath.Clean(configPath)); err != nil {
		return err
	}

	if err := reporter.Stage("Loading config", filepath.Clean(configPath)); err != nil {
		return err
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := reporter.Success("Config ready", fmt.Sprintf("include=%d exclude=%d", len(cfg.Scan.Include), len(cfg.Scan.Exclude))); err != nil {
		return err
	}

	if err := reporter.Stage("Loading catalog metadata", filepath.Clean(configPath)); err != nil {
		return err
	}
	c, err := catalog.Load(configPath, cfg)
	if err != nil {
		return err
	}
	if err := reporter.Success("Catalog metadata loaded", fmt.Sprintf("teams=%d domains=%d", len(c.Teams), len(c.Domains))); err != nil {
		return err
	}

	projectRoot := filepath.Dir(configPath)
	scoped, err := projectscope.Apply(projectRoot, cfg, scopes)
	if err != nil {
		return err
	}
	if err := reporter.Stage("Scanning sources", filepath.Clean(projectRoot)); err != nil {
		return err
	}
	blocks, err := scanner.Scan(projectRoot, scoped.Config)
	if err != nil {
		return err
	}
	if err := reporter.Success("Source scan complete", fmt.Sprintf("blocks=%d", len(blocks))); err != nil {
		return err
	}

	if err := reporter.Stage("Building graph", "layers 4-6"); err != nil {
		return err
	}
	result, err := validator.Build(cfg, c, blocks, validator.BuildOptions{
		SourceRoot: projectRoot,
		Scoped:     scoped.Scoped,
	})
	if result != nil {
		if diagErr := reporter.Diagnostics(result.Diagnostics); diagErr != nil {
			return diagErr
		}
		if summaryErr := reporter.Summary(err == nil, countErrors(result.Diagnostics), countWarnings(result.Diagnostics), len(blocks), len(result.Graph.Nodes), len(result.Graph.Edges)); summaryErr != nil {
			return summaryErr
		}
	}
	if err != nil {
		return err
	}

	return reporter.Success("Validation complete", fmt.Sprintf("config=%s", filepath.Clean(configPath)))
}

func validateProject(root string, cfg *config.Config, c *catalog.Catalog, scopes []string) ([]scanner.RawBlock, *validator.Result, error) {
	scoped, err := projectscope.Apply(root, cfg, scopes)
	if err != nil {
		return nil, nil, err
	}

	blocks, err := scanner.Scan(root, scoped.Config)
	if err != nil {
		return nil, nil, err
	}

	result, err := validator.Build(cfg, c, blocks, validator.BuildOptions{
		SourceRoot: root,
		Scoped:     scoped.Scoped,
	})
	if err != nil {
		return blocks, result, err
	}

	return blocks, result, nil
}

func loadProject(target string) (string, *config.Config, *catalog.Catalog, error) {
	configPath, err := config.Discover(target)
	if err != nil {
		return "", nil, nil, err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return "", nil, nil, err
	}

	c, err := catalog.Load(configPath, cfg)
	if err != nil {
		return "", nil, nil, err
	}

	return configPath, cfg, c, nil
}

func countWarnings(diagnostics []validator.Diagnostic) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "warning" {
			count++
		}
	}
	return count
}

func countErrors(diagnostics []validator.Diagnostic) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			count++
		}
	}
	return count
}

func newScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan [path]",
		Short: "Parse comments and emit normalized graph JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			scopes, err := cmd.Flags().GetStringSlice("scope")
			if err != nil {
				return err
			}

			configPath, cfg, _, err := loadProject(target)
			if err != nil {
				return err
			}

			scoped, err := projectscope.Apply(filepath.Dir(configPath), cfg, scopes)
			if err != nil {
				return err
			}

			blocks, err := scanner.Scan(filepath.Dir(configPath), scoped.Config)
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(blocks)
		},
	}
	bindScopeFlag(c)
	return c
}

func newGraphCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "graph [path]",
		Short: "Produce graph-oriented exports (JSON, Mermaid)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			scopes, err := cmd.Flags().GetStringSlice("scope")
			if err != nil {
				return err
			}

			configPath, cfg, c, err := loadProject(target)
			if err != nil {
				return err
			}
			blocks, result, err := validateProject(filepath.Dir(configPath), cfg, c, scopes)
			_ = blocks
			if err != nil {
				return err
			}

			domains, err := cmd.Flags().GetStringSlice("domain")
			if err != nil {
				return err
			}
			teams, err := cmd.Flags().GetStringSlice("team")
			if err != nil {
				return err
			}
			nodeTypes, err := cmd.Flags().GetStringSlice("type")
			if err != nil {
				return err
			}
			outputPath, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			rendered, err := exportermermaid.Render(&result.Graph, exportermermaid.Options{
				Domains:   domains,
				Teams:     teams,
				NodeTypes: nodeTypes,
			})
			if err != nil {
				return err
			}

			if outputPath == "" {
				_, err = io.WriteString(commandStdout, rendered)
				return err
			}

			return os.WriteFile(outputPath, []byte(rendered), 0o644)
		},
	}
	c.Flags().StringP("output", "o", "", "write Mermaid output to file")
	c.Flags().StringSlice("domain", nil, "include only nodes in the given domain ids")
	c.Flags().StringSlice("team", nil, "include only nodes owned by the given team ids")
	c.Flags().StringSlice("type", nil, "include only nodes of the given node types")
	bindScopeFlag(c)
	return c
}

func newServeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "serve [path]",
		Short: "Start the local interactive explorer UI",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}

			addr, err := cmd.Flags().GetString("addr")
			if err != nil {
				return err
			}
			noWatch, err := cmd.Flags().GetBool("no-watch")
			if err != nil {
				return err
			}
			open, err := cmd.Flags().GetBool("open")
			if err != nil {
				return err
			}
			scopes, err := cmd.Flags().GetStringSlice("scope")
			if err != nil {
				return err
			}

			configPath, err := config.Discover(target)
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), serveSignals()...)
			defer stop()

			opts := server.Options{
				ConfigPath: configPath,
				Addr:       addr,
				Scopes:     scopes,
				Watch:      !noWatch,
				OnReady: func(url string) {
					writeCommandf(commandStdout, "mapture serve: listening on %s (config=%s)\n", url, configPath)
					if open {
						if err := openBrowser(url); err != nil {
							writeCommandf(commandStderr, "mapture serve: could not open browser: %v\n", err)
						}
					}
				},
			}
			if err := server.Serve(ctx, opts); err != nil {
				reportServeError(commandStderr, addr, err)
				return err
			}
			return nil
		},
	}
	c.Flags().String("addr", server.DefaultAddr, "listen address")
	c.Flags().Bool("no-watch", false, "disable filesystem watching and live reload")
	c.Flags().Bool("open", false, "open the explorer in the default browser on start")
	bindScopeFlag(c)
	return c
}

func newUpdateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "update",
		Short: "Upgrade the current mapture installation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			channelValue, err := cmd.Flags().GetString("channel")
			if err != nil {
				return err
			}

			return runUpdateCmd(cmd.Context(), updater.Options{
				RequestedChannel: updater.Channel(channelValue),
				CurrentVersion:   version,
				BuildInfo:        readBuildInfo(),
				Stdout:           commandStdout,
				Stderr:           commandStderr,
			})
		},
	}
	c.Flags().String("channel", "", "update channel override: stable or canary (default: detect from current install)")
	return c
}

func bindScopeFlag(c *cobra.Command) {
	c.Flags().StringSlice("scope", nil, "narrow scanning to one or more project-relative files or directories")
}

func writeCommandf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, format, args...)
}

func reportServeError(w io.Writer, addr string, err error) {
	if err == nil {
		return
	}
	writeCommandf(w, "mapture serve: %v\n", err)
	if errors.Is(err, syscall.EADDRINUSE) {
		writeCommandf(w, "mapture serve: %s is already in use; if you suspended a previous server with Ctrl-Z, run `jobs`, `fg`, or `kill %%<job>` and try again\n", addr)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func newExportHTMLCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "export-html [path]",
		Short: "Write a self-contained HTML architecture report",
		Args:  cobra.MaximumNArgs(1),
		RunE:  todo("export-html"),
	}
	c.Flags().StringP("output", "o", "architecture-report.html", "output file")
	return c
}

func newExportAICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export-ai [path]",
		Short: "Write an AI-ready bundle under .mapture/ai/",
		Args:  cobra.MaximumNArgs(1),
		RunE:  todo("export-ai"),
	}
}

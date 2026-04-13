// Package cmd wires up the Mapture CLI.
//
// Command surface matches the planned CLI. v0.1 implementations are
// stubs that print a TODO banner; real logic lands in src/internal/*.
package cmd

import (
	"context"
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
	"strings"
	"syscall"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/bootstrap"
	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	exportercanonical "github.com/mandotpro/mapture.dev/src/internal/exporter/canonical"
	exporterhtml "github.com/mandotpro/mapture.dev/src/internal/exporter/html"
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
	commandStdout      io.Writer = os.Stdout
	commandStderr      io.Writer = os.Stderr
	runUpdateCmd                 = updater.Run
	inspectRuntime               = updater.Inspect
	checkVersionStatus           = updater.CheckVersion
	colorModeFlag                = string(ui.ColorAuto)
	noColorFlag        bool
	versionFlag        bool
)

const (
	productSiteURL   = "https://mapture.dev"
	productGitHubURL = "https://mapture.dev/github"
)

var rootCmd = &cobra.Command{
	Use:           "mapture",
	Short:         "Repo-native architecture mapping that stays close to the code.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if versionFlag {
			return renderVersionInfo(cmd.OutOrStdout(), readBuildInfo(), currentColorMode(cmd))
		}
		return renderRootHelp(cmd)
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	version = resolveVersion(version, readBuildInfo())
	rootCmd.PersistentFlags().StringVar(&colorModeFlag, "color", string(ui.ColorAuto), "color output: auto, always, never")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable color output (alias for --color=never)")
	rootCmd.AddCommand(
		newInitCmd(),
		newValidateCmd(),
		newScanCmd(),
		newGraphCmd(),
		newServeCmd(),
		newExportJSONCmd(),
		newUpdateCmd(),
		newVersionCmd(),
		newExportHTMLCmd(),
		newExportAICmd(),
	)
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "show version information")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		_, err := selectedColorMode(cmd)
		return err
	}
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		if cmd == rootCmd {
			if err := renderRootHelp(cmd); err != nil {
				_ = ui.NewConsole(cmd.ErrOrStderr(), ui.ColorNever).Error("Could not render help", err.Error())
			}
			return
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
	})
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return renderRootHelp(rootCmd)
			}
			target, _, err := rootCmd.Find(args)
			if err != nil {
				return err
			}
			return target.Help()
		},
	})
}

func cliBrandBlock(console *ui.Console, versionText string, details ...string) string {
	lines := []string{
		console.Header(displayProductVersion(versionText)),
		console.Muted("Repo-native architecture mapping that stays close to the code."),
		console.Join("Site: "+console.Accent(productSiteURL), "GitHub: "+console.Accent(productGitHubURL)),
	}
	for _, detail := range details {
		if strings.TrimSpace(detail) == "" {
			continue
		}
		lines = append(lines, detail)
	}
	return strings.Join(lines, "\n")
}

func selectedColorMode(cmd *cobra.Command) (ui.ColorMode, error) {
	colorValue := colorModeFlag
	noColor := noColorFlag

	if cmd != nil {
		if flag := cmd.Flag("color"); flag != nil && flag.Value != nil {
			colorValue = flag.Value.String()
		}
		if flag := cmd.Flag("no-color"); flag != nil && flag.Value != nil {
			noColor = flag.Value.String() == "true"
			colorFlag := cmd.Flag("color")
			if flag.Changed && colorFlag != nil && colorFlag.Changed {
				return ui.ColorAuto, fmt.Errorf("--color and --no-color cannot be used together")
			}
		}
	}

	if noColor {
		return ui.ColorNever, nil
	}

	return ui.ParseColorMode(colorValue)
}

func currentColorMode(cmd *cobra.Command) ui.ColorMode {
	mode, err := selectedColorMode(cmd)
	if err != nil {
		return ui.ColorAuto
	}
	return mode
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
	return func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		console := ui.NewConsole(commandStderr, currentColorMode(cmd))
		return console.Warning(
			fmt.Sprintf("%s not implemented yet", console.Code("mapture "+name)),
			fmt.Sprintf("target=%s", console.Path(path)),
		)
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "Bootstrap a starter mapture.yaml for the target repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return bootstrap.Run(path, os.Stdin, os.Stdout, os.Stderr, currentColorMode(cmd))
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
			return runValidateWithScopes(target, scopes, commandStdout, commandStderr, currentColorMode(cmd))
		},
	}
	bindScopeFlag(c)
	return c
}

func runValidate(target string, stdout, stderr io.Writer) error {
	return runValidateWithScopes(target, nil, stdout, stderr, ui.ColorAuto)
}

func runValidateWithScopes(target string, scopes []string, stdout, stderr io.Writer, colorMode ui.ColorMode) error {
	reporter := ui.NewReporter(stdout, stderr, colorMode)

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
			fromPath, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}
			if fromPath != "" && len(scopes) > 0 {
				return fmt.Errorf("--scope cannot be used with --from")
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), serveSignals()...)
			defer stop()

			configPath := ""
			readySource := fromPath
			if fromPath == "" {
				configPath, err = config.Discover(target)
				if err != nil {
					return err
				}
				readySource = configPath
			}

			opts := server.Options{
				ConfigPath:  configPath,
				FromPath:    fromPath,
				Addr:        addr,
				Scopes:      scopes,
				ToolVersion: version,
				Watch:       !noWatch && fromPath == "",
				OnReady: func(url string) {
					label := "config"
					if fromPath != "" {
						label = "export"
					}
					console := ui.NewConsole(commandStdout, currentColorMode(cmd))
					_ = console.Success("Explorer listening", console.Join(console.Accent(url), fmt.Sprintf("%s=%s", label, console.Path(readySource))))
					if open {
						if err := openBrowser(url); err != nil {
							_ = ui.NewConsole(commandStderr, currentColorMode(cmd)).Warning("Could not open browser", err.Error())
						}
					}
				},
			}
			if err := server.Serve(ctx, opts); err != nil {
				reportServeError(commandStderr, addr, err, currentColorMode(cmd))
				return err
			}
			return nil
		},
	}
	c.Flags().String("addr", server.DefaultAddr, "listen address")
	c.Flags().String("from", "", "serve a canonical export JSON file instead of scanning a repository")
	c.Flags().Bool("no-watch", false, "disable filesystem watching and live reload")
	c.Flags().Bool("open", false, "open the explorer in the default browser on start")
	bindScopeFlag(c)
	return c
}

func newExportJSONCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "export-json [path]",
		Short: "Write the canonical Mapture JSON export",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			outputPath, err := cmd.Flags().GetString("out")
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

			doc, buildErr := exportercanonical.BuildProject(configPath, exportercanonical.ProjectOptions{
				Scopes:      scopes,
				ToolVersion: version,
				Mode:        exportercanonical.ModeStatic,
			})
			if doc == nil {
				return buildErr
			}

			writer := cmd.OutOrStdout()
			if outputPath != "" {
				file, err := os.Create(outputPath)
				if err != nil {
					return err
				}
				defer func() { _ = file.Close() }()
				writer = file
			}

			encoder := json.NewEncoder(writer)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(doc); err != nil {
				return err
			}
			return buildErr
		},
	}
	c.Flags().StringP("out", "o", "", "write canonical JSON export to file")
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
				ColorMode:        currentColorMode(cmd),
			})
		},
	}
	c.Flags().String("channel", "", "update channel override: stable or canary (default: detect from current install)")
	return c
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the current Mapture version and release channel",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderVersionInfo(cmd.OutOrStdout(), readBuildInfo(), currentColorMode(cmd))
		},
	}
}

func bindScopeFlag(c *cobra.Command) {
	c.Flags().StringSlice("scope", nil, "narrow scanning to one or more project-relative files or directories")
}

func renderRootHelp(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	if err := writeProductHeader(out, readBuildInfo(), true, currentColorMode(cmd)); err != nil {
		return err
	}

	clone := *cmd
	clone.Short = ""
	clone.Long = ""
	_, err := io.WriteString(out, clone.UsageString())
	return err
}

func renderVersionInfo(out io.Writer, info *debug.BuildInfo, colorMode ui.ColorMode) error {
	return writeProductHeader(out, info, true, colorMode)
}

func writeProductHeader(out io.Writer, info *debug.BuildInfo, includeFreshness bool, colorMode ui.ColorMode) error {
	console := ui.NewConsole(out, colorMode)

	runtimeInfo, err := inspectRuntime(version, info)
	if err != nil {
		return err
	}

	metaLine := console.Join(string(runtimeInfo.Channel), runtimeInfo.InstallMethod, console.Path(runtimeInfo.ExecutablePath))
	details := []string{}
	if metaLine != "" {
		details = append(details, metaLine)
	}
	if err := console.Println(cliBrandBlock(console, runtimeInfo.Version, details...)); err != nil {
		return err
	}

	if includeFreshness {
		status, err := bestEffortVersionStatus(info)
		if err == nil && status.UpdateAvailable {
			if err := console.Warning("Update available", console.Strong(status.LatestForChannel)); err != nil {
				return err
			}
			if err := console.Println(console.Accent("Run: mapture update")); err != nil {
				return err
			}
		}
	}

	return console.Println("")
}

func bestEffortVersionStatus(info *debug.BuildInfo) (updater.VersionStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return checkVersionStatus(ctx, version, info)
}

func displayProductVersion(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func reportServeError(w io.Writer, addr string, err error, colorMode ui.ColorMode) {
	if err == nil {
		return
	}
	console := ui.NewConsole(w, colorMode)
	_ = console.Error("Serve failed", err.Error())
	if errors.Is(err, syscall.EADDRINUSE) {
		_ = console.Warning(
			"Address already in use",
			fmt.Sprintf("%s — if you suspended a previous server with Ctrl-Z, run `jobs`, `fg`, or `kill %%<job>` and try again", console.Code(addr)),
		)
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
		Short: "Write a static explorer bundle with data.json",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			outputDir, err := cmd.Flags().GetString("output")
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

			doc, buildErr := exportercanonical.BuildProject(configPath, exportercanonical.ProjectOptions{
				Scopes:      scopes,
				ToolVersion: version,
				Mode:        exportercanonical.ModeStatic,
				SourceLabel: "static build",
			})
			if doc == nil {
				return buildErr
			}
			if err := exporterhtml.WriteBundle(outputDir, doc); err != nil {
				return err
			}
			return buildErr
		},
	}
	c.Flags().StringP("output", "o", "mapture-explorer", "output directory")
	bindScopeFlag(c)
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

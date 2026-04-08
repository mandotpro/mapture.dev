// Package cmd wires up the Mapture CLI.
//
// Command surface matches the planned CLI. v0.1 implementations are
// stubs that print a TODO banner; real logic lands in src/internal/*.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/angelmanchev/mapture/src/internal/bootstrap"
	"github.com/angelmanchev/mapture/src/internal/catalog"
	"github.com/angelmanchev/mapture/src/internal/config"
	"github.com/angelmanchev/mapture/src/internal/scanner"
	"github.com/angelmanchev/mapture/src/internal/validator"
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags.
var version = "0.0.0-dev"

var rootCmd = &cobra.Command{
	Use:          "mapture",
	Short:        "Repo-native architecture graph tool",
	Long:         "Mapture turns catalog YAML and structured code comments into validated architecture graphs, diagrams, and AI-ready bundles.",
	Version:      version,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(
		newInitCmd(),
		newValidateCmd(),
		newScanCmd(),
		newGraphCmd(),
		newServeCmd(),
		newExportHTMLCmd(),
		newExportAICmd(),
	)
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
		Short: "Bootstrap mapture.yaml and architecture/ catalog files",
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
	return &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate config, catalogs, comments, and graph references",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}

			configPath, cfg, c, err := loadProject(target)
			if err != nil {
				return err
			}
			blocks, result, err := validateProject(filepath.Dir(configPath), cfg, c)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(
				os.Stdout,
				"mapture validate: OK (config=%s teams=%d domains=%d events=%d blocks=%d nodes=%d edges=%d warnings=%d)\n",
				filepath.Clean(configPath),
				len(c.Teams),
				len(c.Domains),
				len(c.Events),
				len(blocks),
				len(result.Graph.Nodes),
				len(result.Graph.Edges),
				countWarnings(result.Diagnostics),
			); err != nil {
				return err
			}
			return nil
		},
	}
}

func validateProject(root string, cfg *config.Config, c *catalog.Catalog) ([]scanner.RawBlock, *validator.Result, error) {
	blocks, err := scanner.Scan(root, cfg)
	if err != nil {
		return nil, nil, err
	}

	result, err := validator.Build(cfg, c, blocks)
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

	catalogDir, err := cfg.CatalogDir(configPath)
	if err != nil {
		return "", nil, nil, err
	}

	c, err := catalog.Load(catalogDir)
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

func newScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan [path]",
		Short: "Parse comments and emit normalized graph JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}

			configPath, cfg, _, err := loadProject(target)
			if err != nil {
				return err
			}

			blocks, err := scanner.Scan(filepath.Dir(configPath), cfg)
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(blocks)
		},
	}
}

func newGraphCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "graph [path]",
		Short: "Produce graph-oriented exports (JSON, Mermaid)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  todo("graph"),
	}
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve [path]",
		Short: "Start the local interactive explorer UI",
		Args:  cobra.MaximumNArgs(1),
		RunE:  todo("serve"),
	}
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

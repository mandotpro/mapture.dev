// Package cmd wires up the Mapture CLI.
//
// Command surface matches the planned CLI. v0.1 implementations are
// stubs that print a TODO banner; real logic lands in src/internal/*.
package cmd

import (
	"fmt"
	"os"

	"github.com/angelmanchev/mapture/src/internal/bootstrap"
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
		fmt.Fprintf(os.Stderr, "mapture %s: not implemented yet (target=%s)\n", name, path)
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
		RunE:  todo("validate"),
	}
}

func newScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan [path]",
		Short: "Parse comments and emit normalized graph JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE:  todo("scan"),
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

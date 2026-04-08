// build-web bundles the Mapture Explorer frontend into web/dist/.
//
// Contributors never need a separate Node workflow: this is a plain Go
// program that drives esbuild's Go API to transpile TypeScript sources
// under web/src/, copies the vendored Cytoscape.js distribution, and
// writes a deterministic set of files to web/dist/ that both the HTML
// exporter and the local server embed via //go:embed.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fail(err)
	}

	webDir := filepath.Join(repoRoot, "web")
	srcDir := filepath.Join(webDir, "src")
	vendorDir := filepath.Join(webDir, "vendor")
	distDir := filepath.Join(webDir, "dist")

	if err := os.RemoveAll(distDir); err != nil {
		fail(fmt.Errorf("clean dist: %w", err))
	}
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		fail(fmt.Errorf("create dist: %w", err))
	}

	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{filepath.Join(srcDir, "main.ts")},
		Bundle:            true,
		Format:            api.FormatIIFE,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Target:            api.ES2020,
		Outfile:           filepath.Join(distDir, "app.js"),
		Write:             true,
		LogLevel:          api.LogLevelWarning,
		Loader:            map[string]api.Loader{".ts": api.LoaderTS},
	})
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "esbuild error: %s\n", e.Text)
		}
		os.Exit(1)
	}

	copies := []struct{ src, dst string }{
		{filepath.Join(srcDir, "index.html"), filepath.Join(distDir, "index.html")},
		{filepath.Join(srcDir, "styles.css"), filepath.Join(distDir, "styles.css")},
		{filepath.Join(vendorDir, "cytoscape.min.js"), filepath.Join(distDir, "cytoscape.min.js")},
	}
	for _, c := range copies {
		if err := copyFile(c.src, c.dst); err != nil {
			fail(fmt.Errorf("copy %s: %w", filepath.Base(c.src), err))
		}
	}

	fmt.Printf("build-web: wrote %s\n", relPath(repoRoot, distDir))
}

func findRepoRoot() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate go.mod from %s", start)
		}
		dir = parent
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func relPath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "build-web: %v\n", err)
	os.Exit(1)
}

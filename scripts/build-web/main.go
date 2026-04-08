// build-web installs and builds the embedded frontend bundle under
// src/internal/webui/dist.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fail(err)
	}

	frontendDir := filepath.Join(repoRoot, "src", "internal", "webui", "frontend")
	distDir := filepath.Join(repoRoot, "src", "internal", "webui", "dist")

	lockfilePath := filepath.Join(frontendDir, "package-lock.json")
	if _, err := os.Stat(lockfilePath); err == nil {
		run(frontendDir, "npm", "ci")
	} else {
		run(frontendDir, "npm", "install")
	}
	run(frontendDir, "npm", "run", "build")

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

func relPath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

func run(workdir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = workdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fail(fmt.Errorf("%s %v: %w", name, args, err))
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "build-web: %v\n", err)
	os.Exit(1)
}

package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectIncludeDirsFallsBackToSrc(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	options, defaults, err := detectIncludeDirs(root)
	if err != nil {
		t.Fatalf("detectIncludeDirs returned error: %v", err)
	}

	if len(options) == 0 || options[0] != "./src" {
		t.Fatalf("expected ./src fallback option, got %v", options)
	}
	if len(defaults) != 1 || defaults[0] != "./src" {
		t.Fatalf("expected ./src fallback default, got %v", defaults)
	}
}

func TestDetectLanguagesFindsGoAndTypeScript(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "client.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatalf("write ts file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "legacy.php"), []byte("<?php\n"), 0o644); err != nil {
		t.Fatalf("write php file: %v", err)
	}

	langs, err := detectLanguages(root, []string{"./src"})
	if err != nil {
		t.Fatalf("detectLanguages returned error: %v", err)
	}

	if !langs["go"] {
		t.Fatalf("expected go to be detected")
	}
	if !langs["typescript"] {
		t.Fatalf("expected typescript to be detected")
	}
	if langs["php"] {
		t.Fatalf("did not expect php outside selected includes to be detected")
	}
}

func TestMergePathsParsesCommaSeparatedInput(t *testing.T) {
	t.Parallel()

	paths := mergePaths(nil, "./src, cmd, ../shared, ./src")

	expected := []string{"./src", "./cmd", "../shared"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %v", len(expected), paths)
	}
	for i := range expected {
		if paths[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, paths)
		}
	}
}

func TestWriteScaffoldSkipsExistingFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "architecture"), 0o755); err != nil {
		t.Fatalf("mkdir architecture: %v", err)
	}
	existing := filepath.Join(root, "mapture.yaml")
	if err := os.WriteFile(existing, []byte("version: 1\n"), 0o644); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	config := initConfig{
		Includes:               []string{"./src"},
		Excludes:               []string{"./.git"},
		LanguageEnabled:        map[string]bool{"go": true, "php": false, "typescript": false, "javascript": false},
		FailOnUnknownDomain:    true,
		FailOnUnknownTeam:      true,
		FailOnUnknownEvent:     true,
		FailOnUnknownNode:      true,
		WarnOnDeprecatedEvents: true,
		WarnOnOrphanedNodes:    false,
	}

	result, err := writeScaffold(root, config, true)
	if err != nil {
		t.Fatalf("writeScaffold returned error: %v", err)
	}

	if len(result.Skipped) != 1 || result.Skipped[0] != "mapture.yaml" {
		t.Fatalf("expected mapture.yaml to be skipped, got %+v", result)
	}
	if len(result.Created) != 3 {
		t.Fatalf("expected three created catalog files, got %+v", result)
	}
}

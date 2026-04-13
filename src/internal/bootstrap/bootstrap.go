// Package bootstrap scaffolds repository config and starter catalogs for
// `mapture init`.
package bootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/mandotpro/mapture.dev/src/internal/ui"
)

var supportedLanguages = []languageOption{
	{Label: "Go", Key: "go", Extensions: []string{".go"}},
	{Label: "PHP", Key: "php", Extensions: []string{".php"}},
	{Label: "TypeScript", Key: "typescript", Extensions: []string{".ts", ".tsx"}},
	{Label: "JavaScript", Key: "javascript", Extensions: []string{".js", ".jsx"}},
}

var preferredIncludeDirs = []string{"src", "cmd", "pkg", "internal", "services", "app", "lib"}
var wellKnownExcludeDirs = []string{".git", "vendor", "node_modules", "dist", "build"}

// Run executes the interactive init flow for the target repository.
func Run(target string, stdin, stdout, stderr *os.File, colorMode ui.ColorMode) error {
	root, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve target path: %w", err)
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	state, err := detectProject(root)
	if err != nil {
		return err
	}

	stdio := survey.WithStdio(stdin, stdout, stderr)
	console := ui.NewConsole(stdout, colorMode)

	skipExisting, err := resolveExistingFiles(state.ExistingFiles, stdio)
	if err != nil {
		return err
	}

	config, err := promptConfig(root, state, stdio, console)
	if err != nil {
		return err
	}

	result, err := writeScaffold(root, config, skipExisting)
	if err != nil {
		return err
	}

	return printSummary(console, root, result)
}

type languageOption struct {
	Label      string
	Key        string
	Extensions []string
}

type detectedState struct {
	IncludeSuggestions []string
	DefaultIncludes    []string
	DefaultExcludes    []string
	ExistingFiles      []string
}

type initConfig struct {
	Includes               []string
	Excludes               []string
	LanguageEnabled        map[string]bool
	FailOnUnknownDomain    bool
	FailOnUnknownTeam      bool
	FailOnUnknownNode      bool
	WarnOnOrphanedNodes    bool
	WarnOnDeprecatedEvents bool
}

type writeResult struct {
	Created []string
	Skipped []string
}

func detectProject(root string) (detectedState, error) {
	includeOptions, defaultIncludes, err := detectIncludeDirs(root)
	if err != nil {
		return detectedState{}, err
	}

	return detectedState{
		IncludeSuggestions: includeOptions,
		DefaultIncludes:    defaultIncludes,
		DefaultExcludes:    defaultExcludeDirs(),
		ExistingFiles:      existingScaffoldFiles(root),
	}, nil
}

func detectIncludeDirs(root string) ([]string, []string, error) {
	options := make([]string, 0, len(preferredIncludeDirs)+1)
	defaults := make([]string, 0, len(preferredIncludeDirs))

	rootHasSource, err := hasSupportedFile(root)
	if err != nil {
		return nil, nil, err
	}
	if rootHasSource {
		options = append(options, ".")
		defaults = append(defaults, ".")
	}

	for _, dir := range preferredIncludeDirs {
		fullPath := filepath.Join(root, dir)
		info, err := os.Stat(fullPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, nil, fmt.Errorf("inspect %s: %w", fullPath, err)
		}
		if info.IsDir() {
			rel := "./" + dir
			options = append(options, rel)
			defaults = append(defaults, rel)
		}
	}

	if len(options) == 0 {
		options = []string{"./src", "./cmd", "./pkg"}
		defaults = []string{"./src"}
	}

	return options, defaults, nil
}

func hasSupportedFile(root string) (bool, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", root, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		for _, lang := range supportedLanguages {
			for _, supportedExt := range lang.Extensions {
				if ext == supportedExt {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func detectLanguages(root string, includes []string) (map[string]bool, error) {
	detected := make(map[string]bool, len(supportedLanguages))
	excluded := make(map[string]struct{}, len(wellKnownExcludeDirs)+1)
	for _, dir := range wellKnownExcludeDirs {
		excluded[dir] = struct{}{}
	}
	excluded["architecture"] = struct{}{}

	for _, include := range includes {
		scanRoot := include
		if include == "." {
			scanRoot = root
		} else if !filepath.IsAbs(include) {
			scanRoot = filepath.Join(root, include)
		}

		info, err := os.Stat(scanRoot)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("inspect include %s: %w", scanRoot, err)
		}

		if !info.IsDir() {
			detectLanguageForFile(scanRoot, detected)
			continue
		}

		err = filepath.WalkDir(scanRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				if path == scanRoot {
					return nil
				}
				if _, skip := excluded[d.Name()]; skip {
					return filepath.SkipDir
				}
				return nil
			}

			detectLanguageForFile(path, detected)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("detect languages in %s: %w", include, err)
		}
	}

	return detected, nil
}

func detectLanguageForFile(path string, detected map[string]bool) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, lang := range supportedLanguages {
		for _, candidate := range lang.Extensions {
			if ext == candidate {
				detected[lang.Key] = true
			}
		}
	}
}

func existingScaffoldFiles(root string) []string {
	candidates := []string{"mapture.yaml"}

	existing := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		fullPath := filepath.Join(root, candidate)
		if _, err := os.Stat(fullPath); err == nil {
			existing = append(existing, candidate)
		}
	}
	sort.Strings(existing)
	return existing
}

func resolveExistingFiles(existing []string, stdio survey.AskOpt) (bool, error) {
	if len(existing) == 0 {
		return false, nil
	}

	message := "Starter files already exist:\n"
	for _, path := range existing {
		message += "  - " + path + "\n"
	}
	message += "\nChoose how init should proceed."

	var action string
	prompt := &survey.Select{
		Message: message,
		Options: []string{
			"Merge by keeping existing files and creating only missing files",
			"Abort and inspect the existing scaffold manually",
		},
		Default: "Merge by keeping existing files and creating only missing files",
	}
	if err := survey.AskOne(prompt, &action, stdio); err != nil {
		return false, fmt.Errorf("prompt for existing files: %w", err)
	}

	if strings.HasPrefix(action, "Abort") {
		return false, errors.New("init aborted because scaffold files already exist")
	}

	return true, nil
}

func promptConfig(root string, state detectedState, stdio survey.AskOpt, console *ui.Console) (initConfig, error) {
	defaultIncludes := strings.Join(state.DefaultIncludes, ", ")
	includePrompt := "Source directories to scan (comma-separated)"
	if len(state.IncludeSuggestions) > 0 {
		includePrompt += "\nSuggestions: " + strings.Join(state.IncludeSuggestions, ", ")
	}

	var includeInput string
	if err := survey.AskOne(
		&survey.Input{
			Message: includePrompt,
			Default: defaultIncludes,
		},
		&includeInput,
		survey.WithValidator(pathListRequired),
		stdio,
	); err != nil {
		return initConfig{}, fmt.Errorf("prompt for scan.include: %w", err)
	}

	includes := mergePaths(nil, includeInput)
	detectedLangs, err := detectLanguages(root, includes)
	if err != nil {
		return initConfig{}, err
	}
	if err := printDetectedLanguages(console, includes, detectedLangs); err != nil {
		return initConfig{}, fmt.Errorf("report detected languages: %w", err)
	}

	defaultExcludes := strings.Join(state.DefaultExcludes, ", ")
	excludePrompt := "Exclude directories or globs (comma-separated)"
	if len(state.DefaultExcludes) > 0 {
		excludePrompt += "\nSuggestions: " + strings.Join(state.DefaultExcludes, ", ")
	}

	var excludeInput string
	if err := survey.AskOne(
		&survey.Input{
			Message: excludePrompt,
			Default: defaultExcludes,
		},
		&excludeInput,
		stdio,
	); err != nil {
		return initConfig{}, fmt.Errorf("prompt for scan.exclude: %w", err)
	}

	languageOptions := make([]string, 0, len(supportedLanguages))
	languageDefaults := make([]string, 0, len(supportedLanguages))
	for _, option := range supportedLanguages {
		languageOptions = append(languageOptions, option.Label)
		if detectedLangs[option.Key] {
			languageDefaults = append(languageDefaults, option.Label)
		}
	}

	var selectedLanguages []string
	if err := survey.AskOne(
		&survey.MultiSelect{
			Message: "Confirm which languages Mapture should scan:",
			Options: languageOptions,
			Default: languageDefaults,
			Description: func(value string, _ int) string {
				for _, option := range supportedLanguages {
					if option.Label == value && detectedLangs[option.Key] {
						return "Detected in the target tree"
					}
				}
				return ""
			},
		},
		&selectedLanguages,
		survey.WithValidator(survey.Required),
		stdio,
	); err != nil {
		return initConfig{}, fmt.Errorf("prompt for languages: %w", err)
	}

	config := initConfig{
		Includes:               includes,
		Excludes:               mergePaths(nil, excludeInput),
		LanguageEnabled:        make(map[string]bool, len(supportedLanguages)),
		FailOnUnknownDomain:    true,
		FailOnUnknownTeam:      true,
		FailOnUnknownNode:      true,
		WarnOnDeprecatedEvents: true,
	}

	selectedSet := make(map[string]struct{}, len(selectedLanguages))
	for _, label := range selectedLanguages {
		selectedSet[label] = struct{}{}
	}
	for _, option := range supportedLanguages {
		_, enabled := selectedSet[option.Label]
		config.LanguageEnabled[option.Key] = enabled
	}

	validationPrompts := []struct {
		message string
		target  *bool
	}{
		{message: "Fail when a comment references an unknown team?", target: &config.FailOnUnknownTeam},
		{message: "Fail when a comment references an unknown domain?", target: &config.FailOnUnknownDomain},
		{message: "Fail when a relation targets an unknown node?", target: &config.FailOnUnknownNode},
		{message: "Warn when declared nodes have no edges?", target: &config.WarnOnOrphanedNodes},
		{message: "Warn when deprecated events are referenced?", target: &config.WarnOnDeprecatedEvents},
	}

	for _, prompt := range validationPrompts {
		if err := survey.AskOne(
			&survey.Confirm{
				Message: prompt.message,
				Default: *prompt.target,
			},
			prompt.target,
			stdio,
		); err != nil {
			return initConfig{}, fmt.Errorf("prompt for validation settings: %w", err)
		}
	}

	return config, nil
}

func mergePaths(selected []string, extra string) []string {
	seen := make(map[string]struct{}, len(selected))
	merged := make([]string, 0, len(selected)+2)

	for _, path := range selected {
		normalized := normalizePath(path)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		merged = append(merged, normalized)
	}

	for _, raw := range strings.Split(extra, ",") {
		normalized := normalizePath(raw)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		merged = append(merged, normalized)
	}

	return merged
}

func pathListRequired(value any) error {
	text, _ := value.(string)
	if len(mergePaths(nil, text)) == 0 {
		return errors.New("enter at least one path")
	}
	return nil
}

func printDetectedLanguages(console *ui.Console, includes []string, detected map[string]bool) error {
	found := make([]string, 0, len(supportedLanguages))
	for _, option := range supportedLanguages {
		if detected[option.Key] {
			found = append(found, option.Label)
		}
	}

	if len(found) == 0 {
		if err := console.Warning(
			"No supported languages detected",
			fmt.Sprintf("%s — you can still enable them manually in the next step", strings.Join(includes, ", ")),
		); err != nil {
			return err
		}
		return console.Println("")
	}

	if err := console.Success(
		"Detected languages",
		fmt.Sprintf("%s → %s", strings.Join(includes, ", "), strings.Join(found, ", ")),
	); err != nil {
		return err
	}
	return console.Println("")
}

func normalizePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if trimmed == "." || trimmed == "./" {
		return "."
	}
	if strings.HasPrefix(trimmed, "./") {
		return "./" + strings.TrimPrefix(filepath.Clean(trimmed), "./")
	}
	if strings.HasPrefix(trimmed, "../") {
		return filepath.Clean(trimmed)
	}
	if strings.HasPrefix(trimmed, "/") {
		return filepath.Clean(trimmed)
	}
	return "./" + filepath.Clean(trimmed)
}

func defaultExcludeDirs() []string {
	defaults := make([]string, 0, len(wellKnownExcludeDirs))
	for _, dir := range wellKnownExcludeDirs {
		defaults = append(defaults, "./"+dir)
	}
	return defaults
}

func writeScaffold(root string, config initConfig, skipExisting bool) (writeResult, error) {
	files := map[string]string{
		"mapture.yaml": renderConfig(config),
	}

	result := writeResult{}
	for relativePath, content := range files {
		fullPath := filepath.Join(root, relativePath)
		if _, err := os.Stat(fullPath); err == nil && skipExisting {
			result.Skipped = append(result.Skipped, relativePath)
			continue
		}

		if err := writeFile(fullPath, content, skipExisting); err != nil {
			return writeResult{}, err
		}
		result.Created = append(result.Created, relativePath)
	}

	sort.Strings(result.Created)
	sort.Strings(result.Skipped)
	return result, nil
}

func writeFile(path string, content string, skipExisting bool) (err error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) && skipExisting {
			return nil
		}
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("refusing to overwrite existing file %s", path)
		}
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", path, closeErr)
		}
	}()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func renderConfig(config initConfig) string {
	var b bytes.Buffer

	b.WriteString("# Mapture repository config.\n\n")
	b.WriteString("version: 1\n\n")
	b.WriteString("# Teams and domains live here by default so a fresh setup starts with one file.\n")
	b.WriteString("# If you later want split catalogs, uncomment the legacy section below.\n")
	b.WriteString("# catalog:\n")
	b.WriteString("#   dir: ./architecture\n\n")
	b.WriteString("# Optional shared tag vocabulary for nodes, events, teams, and domains.\n")
	b.WriteString("# tags:\n")
	b.WriteString("#   - critical-path\n")
	b.WriteString("#   - customer-facing\n\n")
	b.WriteString("# Optional direct-only categorical facets for nodes and events.\n")
	b.WriteString("# facets:\n")
	b.WriteString("#   event.type:\n")
	b.WriteString("#     label: Event Type\n")
	b.WriteString("#     values:\n")
	b.WriteString("#       - sync\n")
	b.WriteString("#       - async\n")
	b.WriteString("#       - queue\n")
	b.WriteString("#       - event-bus\n")
	b.WriteString("#   db.type:\n")
	b.WriteString("#     label: Database Type\n")
	b.WriteString("#     values:\n")
	b.WriteString("#       - tenant\n")
	b.WriteString("#       - shared\n\n")
	b.WriteString("teams:\n")
	b.WriteString("  - id: team-commerce\n")
	b.WriteString("    name: Commerce Team\n")
	b.WriteString("    email: commerce@example.com\n\n")
	b.WriteString("  - id: team-billing\n")
	b.WriteString("    name: Billing Team\n")
	b.WriteString("    email: billing@example.com\n\n")
	b.WriteString("domains:\n")
	b.WriteString("  - id: orders\n")
	b.WriteString("    name: Orders\n")
	b.WriteString("    ownerTeams: [team-commerce]\n")
	b.WriteString("    description: Handles the full lifecycle of customer orders.\n\n")
	b.WriteString("  - id: billing\n")
	b.WriteString("    name: Billing\n")
	b.WriteString("    ownerTeams: [team-billing]\n")
	b.WriteString("    description: Manages payment capture and invoicing.\n\n")
	b.WriteString("scan:\n")
	b.WriteString("  # Directories Mapture scans for @arch.* and @event.* tags.\n")
	b.WriteString("  include:\n")
	for _, path := range config.Includes {
		b.WriteString("    - " + path + "\n")
	}
	b.WriteString("  # Directories or globs Mapture skips during scanning.\n")
	b.WriteString("  exclude:\n")
	for _, path := range config.Excludes {
		b.WriteString("    - " + path + "\n")
	}
	b.WriteString("\n")
	b.WriteString("languages:\n")
	for _, option := range supportedLanguages {
		_, _ = fmt.Fprintf(&b, "  %s: %t\n", option.Key, config.LanguageEnabled[option.Key])
	}
	b.WriteString("\n")
	b.WriteString("comments:\n")
	b.WriteString("  # v0.1 supports flat @key value tags only.\n")
	b.WriteString("  style: tags\n\n")
	b.WriteString("ui:\n")
	b.WriteString("  # Default explorer layout on first load.\n")
	b.WriteString("  defaultLayout: elk-horizontal\n\n")
	b.WriteString("validation:\n")
	_, _ = fmt.Fprintf(&b, "  failOnUnknownDomain: %t\n", config.FailOnUnknownDomain)
	_, _ = fmt.Fprintf(&b, "  failOnUnknownTeam: %t\n", config.FailOnUnknownTeam)
	_, _ = fmt.Fprintf(&b, "  failOnUnknownNode: %t\n", config.FailOnUnknownNode)
	b.WriteString("  # Uncomment roles below if every usage must carry event metadata.\n")
	b.WriteString("  # requireMetadataOn:\n")
	b.WriteString("  #   - trigger\n")
	b.WriteString("  #   - listener\n")
	_, _ = fmt.Fprintf(&b, "  warnOnOrphanedNodes: %t\n", config.WarnOnOrphanedNodes)
	_, _ = fmt.Fprintf(&b, "  warnOnDeprecatedEvents: %t\n", config.WarnOnDeprecatedEvents)

	return b.String()
}

func printSummary(console *ui.Console, root string, result writeResult) error {
	if err := console.Success("Initialized mapture scaffold", console.Path(root)); err != nil {
		return err
	}
	if len(result.Created) > 0 {
		if err := console.Stage("Created", ""); err != nil {
			return err
		}
		for _, path := range result.Created {
			if err := console.Println("  - " + console.Path(path)); err != nil {
				return err
			}
		}
	}
	if len(result.Skipped) > 0 {
		if err := console.Warning("Skipped existing", ""); err != nil {
			return err
		}
		for _, path := range result.Skipped {
			if err := console.Println("  - " + console.Path(path)); err != nil {
				return err
			}
		}
	}
	return nil
}

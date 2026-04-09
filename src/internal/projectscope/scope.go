// Package projectscope applies ad-hoc CLI scan scopes on top of discovered project config.
package projectscope

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mandotpro/mapture.dev/src/internal/config"
)

// Applied captures the effective scope-adjusted configuration.
type Applied struct {
	Config  *config.Config
	Scopes  []string
	Scoped  bool
	Display string
}

// Apply narrows cfg.Scan.Include to the provided project-relative scopes.
func Apply(root string, cfg *config.Config, scopes []string) (*Applied, error) {
	if cfg == nil {
		return nil, fmt.Errorf("scope config is nil")
	}
	if len(scopes) == 0 {
		return &Applied{
			Config:  cloneConfig(cfg),
			Scopes:  nil,
			Scoped:  false,
			Display: "",
		}, nil
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve scope root: %w", err)
	}

	normalizedScopes := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		normalized, err := normalizeScope(absRoot, scope)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		if err := ensureCoveredByIncludes(absRoot, cfg.Scan.Include, normalized); err != nil {
			return nil, err
		}
		seen[normalized] = struct{}{}
		normalizedScopes = append(normalizedScopes, normalized)
	}
	sort.Strings(normalizedScopes)

	cloned := cloneConfig(cfg)
	cloned.Scan.Include = append([]string(nil), normalizedScopes...)

	return &Applied{
		Config:  cloned,
		Scopes:  normalizedScopes,
		Scoped:  true,
		Display: strings.Join(normalizedScopes, ", "),
	}, nil
}

// SourceLabel formats a scoped source label for human-facing surfaces.
func SourceLabel(base string, scopes []string) string {
	if len(scopes) == 0 {
		return base
	}
	return fmt.Sprintf("%s (scoped: %s)", base, strings.Join(scopes, ", "))
}

func cloneConfig(cfg *config.Config) *config.Config {
	cloned := *cfg
	cloned.Scan.Include = append([]string(nil), cfg.Scan.Include...)
	cloned.Scan.Exclude = append([]string(nil), cfg.Scan.Exclude...)
	cloned.Validation.RequireMetadataOn = append([]string(nil), cfg.Validation.RequireMetadataOn...)
	return &cloned
}

func normalizeScope(root string, scope string) (string, error) {
	trimmed := strings.TrimSpace(scope)
	if trimmed == "" {
		return "", fmt.Errorf("scope path cannot be empty")
	}

	candidate := trimmed
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(root, candidate)
	}
	candidate = filepath.Clean(candidate)

	info, err := os.Stat(candidate)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("scope path %s does not exist", trimmed)
		}
		return "", fmt.Errorf("inspect scope path %s: %w", trimmed, err)
	}
	_ = info

	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", fmt.Errorf("resolve scope path %s relative to project root: %w", trimmed, err)
	}
	if rel == "." {
		return ".", nil
	}

	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("scope path %s is outside project root", trimmed)
	}
	return "./" + rel, nil
}

func ensureCoveredByIncludes(root string, includes []string, scope string) error {
	scopeAbs := scope
	if !filepath.IsAbs(scopeAbs) {
		scopeAbs = filepath.Join(root, scope)
	}
	scopeAbs = filepath.Clean(scopeAbs)

	for _, include := range includes {
		includeAbs := include
		if !filepath.IsAbs(includeAbs) {
			includeAbs = filepath.Join(root, include)
		}
		includeAbs = filepath.Clean(includeAbs)
		if covers(includeAbs, scopeAbs) {
			return nil
		}
	}

	return fmt.Errorf("scope path %s is outside configured scan.include paths", scope)
}

func covers(includeAbs string, targetAbs string) bool {
	includeInfo, err := os.Stat(includeAbs)
	if err != nil {
		return false
	}
	if !includeInfo.IsDir() {
		return includeAbs == targetAbs
	}
	if includeAbs == targetAbs {
		return true
	}
	rel, err := filepath.Rel(includeAbs, targetAbs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(filepath.ToSlash(rel), "../")
}

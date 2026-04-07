// Package scanner walks configured source trees and extracts structured
// `@arch.*` and `@event.*` metadata from comment blocks.
package scanner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/angelmanchev/mapture/src/internal/config"
	"github.com/angelmanchev/mapture/src/internal/graph"
)

var (
	tagPattern     = regexp.MustCompile(`^@(arch|event)\.([a-z_]+)\s+(.+?)\s*$`)
	nodeIDPattern  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	dotIDPattern   = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)
	allowedArch    = map[string]struct{}{"node": {}, "name": {}, "domain": {}, "owner": {}, "description": {}, "version": {}, "tags": {}, "status": {}, "calls": {}, "depends_on": {}, "stores_in": {}, "reads_from": {}}
	allowedEvent   = map[string]struct{}{"id": {}, "role": {}, "domain": {}, "owner": {}, "phase": {}, "topic": {}, "version": {}, "notes": {}, "producer": {}, "consumer": {}}
	repeatableArch = map[string]struct{}{"calls": {}, "depends_on": {}, "stores_in": {}, "reads_from": {}}
	archStatuses   = map[string]struct{}{"active": {}, "deprecated": {}, "experimental": {}}
	eventRoles     = map[string]struct{}{"definition": {}, "trigger": {}, "listener": {}, "bridge-out": {}, "bridge-in": {}, "publisher": {}, "subscriber": {}}
	eventPhases    = map[string]struct{}{"pre-commit": {}, "post-commit": {}, "async": {}, "integration": {}}
)

var languageExtensions = map[string][]string{
	"php":        {".php"},
	"go":         {".go"},
	"typescript": {".ts", ".tsx"},
	"javascript": {".js", ".jsx"},
}

// TargetRef is a typed node reference extracted from `@arch.*` relation tags.
type TargetRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// RawBlock is the scanner output passed to later validation and graph phases.
type RawBlock struct {
	Kind      string                 `json:"kind"`
	File      string                 `json:"file"`
	Line      int                    `json:"line"`
	Fields    map[string]string      `json:"fields,omitempty"`
	Relations map[string][]TargetRef `json:"relations,omitempty"`
}

// Scan walks a configured repository tree and extracts valid raw blocks.
func Scan(root string, cfg *config.Config) ([]RawBlock, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve scan root: %w", err)
	}

	extensions := enabledExtensions(cfg)
	if len(extensions) == 0 {
		return nil, errors.New("scanner requires at least one enabled language")
	}

	matcher := newExcludeMatcher(cfg.Scan.Exclude)
	seen := make(map[string]struct{})
	blocks := make([]RawBlock, 0)

	for _, include := range cfg.Scan.Include {
		scanRoot := include
		if !filepath.IsAbs(scanRoot) {
			scanRoot = filepath.Join(root, include)
		}
		scanRoot = filepath.Clean(scanRoot)

		info, err := os.Stat(scanRoot)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("scan.include path %s does not exist", include)
			}
			return nil, fmt.Errorf("inspect include %s: %w", include, err)
		}

		if !info.IsDir() {
			fileBlocks, err := scanFile(root, scanRoot, extensions, matcher)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, fileBlocks...)
			continue
		}

		err = filepath.WalkDir(scanRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			rel, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("derive path relative to root: %w", err)
			}
			if rel == "." {
				rel = ""
			}

			if rel != "" && matcher.Match(rel) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				return nil
			}
			if _, ok := extensions[strings.ToLower(filepath.Ext(path))]; !ok {
				return nil
			}
			if _, ok := seen[path]; ok {
				return nil
			}
			seen[path] = struct{}{}

			fileBlocks, err := scanFile(root, path, extensions, matcher)
			if err != nil {
				return err
			}
			blocks = append(blocks, fileBlocks...)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", include, err)
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].File == blocks[j].File {
			if blocks[i].Line == blocks[j].Line {
				return blocks[i].Kind < blocks[j].Kind
			}
			return blocks[i].Line < blocks[j].Line
		}
		return blocks[i].File < blocks[j].File
	})

	return blocks, nil
}

// ParseError reports a malformed comment block.
type ParseError struct {
	File      string
	Line      int
	Namespace string
	Key       string
	Message   string
}

func (e *ParseError) Error() string {
	var b strings.Builder
	b.WriteString("layer 3 comment validation")
	if e.File != "" {
		_, _ = fmt.Fprintf(&b, " %s:%d", e.File, e.Line)
	}
	if e.Namespace != "" {
		_, _ = fmt.Fprintf(&b, " %s", e.Namespace)
	}
	if e.Key != "" {
		_, _ = fmt.Fprintf(&b, ".%s", e.Key)
	}
	_, _ = fmt.Fprintf(&b, ": %s", e.Message)
	return b.String()
}

type commentBlock struct {
	line  int
	lines []string
}

type blockAccumulator struct {
	file      string
	line      int
	namespace string
	fields    map[string]string
	relations map[string][]TargetRef
	seen      map[string]struct{}
}

type excludeMatcher struct {
	patterns []string
}

func newExcludeMatcher(patterns []string) excludeMatcher {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		value := normalizeMatchPath(pattern)
		if value == "" || value == "." {
			continue
		}
		normalized = append(normalized, value)
	}
	return excludeMatcher{patterns: normalized}
}

func (m excludeMatcher) Match(rel string) bool {
	if rel == "" {
		return false
	}
	rel = normalizeMatchPath(rel)
	for _, pattern := range m.patterns {
		if hasGlob(pattern) {
			if ok, _ := filepath.Match(pattern, rel); ok {
				return true
			}
			continue
		}
		if rel == pattern || strings.HasPrefix(rel, pattern+"/") {
			return true
		}
		for _, part := range strings.Split(rel, "/") {
			if part == pattern {
				return true
			}
		}
	}
	return false
}

func normalizeMatchPath(path string) string {
	path = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	path = strings.TrimPrefix(path, "./")
	if path == "." {
		return ""
	}
	return path
}

func hasGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

func enabledExtensions(cfg *config.Config) map[string]struct{} {
	enabled := make(map[string]struct{})
	if cfg.Languages.PHP {
		for _, ext := range languageExtensions["php"] {
			enabled[ext] = struct{}{}
		}
	}
	if cfg.Languages.Go {
		for _, ext := range languageExtensions["go"] {
			enabled[ext] = struct{}{}
		}
	}
	if cfg.Languages.TypeScript {
		for _, ext := range languageExtensions["typescript"] {
			enabled[ext] = struct{}{}
		}
	}
	if cfg.Languages.JavaScript {
		for _, ext := range languageExtensions["javascript"] {
			enabled[ext] = struct{}{}
		}
	}
	return enabled
}

func scanFile(root, path string, extensions map[string]struct{}, matcher excludeMatcher) ([]RawBlock, error) {
	if _, ok := extensions[strings.ToLower(filepath.Ext(path))]; !ok {
		return nil, nil
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, fmt.Errorf("derive relative path for %s: %w", path, err)
	}
	rel = filepath.ToSlash(rel)
	if matcher.Match(rel) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rel, err)
	}

	commentBlocks := extractCommentBlocks(string(data))
	blocks := make([]RawBlock, 0, len(commentBlocks))
	for _, block := range commentBlocks {
		parsed, err := parseBlock(rel, block)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, parsed...)
	}
	return blocks, nil
}

func extractCommentBlocks(src string) []commentBlock {
	lines := strings.Split(src, "\n")
	blocks := make([]commentBlock, 0)
	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])

		if strings.HasPrefix(trimmed, "//") {
			start := i + 1
			group := make([]string, 0, 4)
			for i < len(lines) {
				line := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(line, "//") {
					break
				}
				group = append(group, strings.TrimSpace(strings.TrimPrefix(line, "//")))
				i++
			}
			blocks = append(blocks, commentBlock{line: start, lines: group})
			continue
		}

		if strings.Contains(trimmed, "/**") {
			start := i + 1
			group := make([]string, 0, 4)
			line := lines[i]
			for {
				group = append(group, cleanBlockCommentLine(line))
				if strings.Contains(line, "*/") {
					i++
					break
				}
				i++
				if i >= len(lines) {
					break
				}
				line = lines[i]
			}
			blocks = append(blocks, commentBlock{line: start, lines: group})
			continue
		}

		i++
	}
	return blocks
}

func cleanBlockCommentLine(line string) string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.Replace(trimmed, "/**", "", 1)
	trimmed = strings.Replace(trimmed, "/*", "", 1)
	trimmed = strings.Replace(trimmed, "*/", "", 1)
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.TrimPrefix(trimmed, "*")
	return strings.TrimSpace(trimmed)
}

func parseBlock(file string, block commentBlock) ([]RawBlock, error) {
	accumulators := map[string]*blockAccumulator{}
	for offset, line := range block.lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "@") {
			continue
		}

		matches := tagPattern.FindStringSubmatch(line)
		if matches == nil {
			if strings.HasPrefix(line, "@arch.") || strings.HasPrefix(line, "@event.") {
				return nil, &ParseError{File: file, Line: block.line + offset, Message: fmt.Sprintf("invalid tag format %q", line)}
			}
			continue
		}

		namespace := matches[1]
		key := matches[2]
		value := strings.TrimSpace(matches[3])

		acc, ok := accumulators[namespace]
		if !ok {
			acc = &blockAccumulator{
				file:      file,
				line:      block.line,
				namespace: namespace,
				fields:    make(map[string]string),
				relations: make(map[string][]TargetRef),
				seen:      make(map[string]struct{}),
			}
			accumulators[namespace] = acc
		}

		if err := acc.add(key, value); err != nil {
			return nil, err
		}
	}

	blocks := make([]RawBlock, 0, len(accumulators))
	for _, namespace := range []string{"arch", "event"} {
		acc, ok := accumulators[namespace]
		if !ok {
			continue
		}
		raw, err := acc.finish()
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, raw)
	}
	return blocks, nil
}

func (a *blockAccumulator) add(key, value string) error {
	allowed := allowedEvent
	if a.namespace == "arch" {
		allowed = allowedArch
	}
	if _, ok := allowed[key]; !ok {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: key, Message: "unknown tag key"}
	}

	if a.namespace == "arch" {
		if _, ok := repeatableArch[key]; ok {
			target, err := parseTargetRef(value)
			if err != nil {
				return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: key, Message: err.Error()}
			}
			a.relations[key] = append(a.relations[key], target)
			return nil
		}
	}

	if _, exists := a.seen[key]; exists {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: key, Message: "duplicate tag in the same block"}
	}
	a.seen[key] = struct{}{}
	a.fields[key] = value
	return nil
}

func (a *blockAccumulator) finish() (RawBlock, error) {
	switch a.namespace {
	case "arch":
		if err := validateArchBlock(a); err != nil {
			return RawBlock{}, err
		}
	case "event":
		if err := validateEventBlock(a); err != nil {
			return RawBlock{}, err
		}
	default:
		return RawBlock{}, &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Message: "unknown namespace"}
	}

	fields := make(map[string]string, len(a.fields))
	for key, value := range a.fields {
		fields[key] = value
	}

	relations := make(map[string][]TargetRef, len(a.relations))
	for key, refs := range a.relations {
		clone := append([]TargetRef(nil), refs...)
		relations[key] = clone
	}

	return RawBlock{
		Kind:      a.namespace,
		File:      a.file,
		Line:      a.line,
		Fields:    fields,
		Relations: relations,
	}, nil
}

func validateArchBlock(a *blockAccumulator) error {
	if _, ok := a.fields["node"]; !ok {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "node", Message: "missing required tag"}
	}
	for _, key := range []string{"name", "domain", "owner"} {
		if strings.TrimSpace(a.fields[key]) == "" {
			return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: key, Message: "missing required tag"}
		}
	}

	target, err := parseTargetRef(a.fields["node"])
	if err != nil {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "node", Message: err.Error()}
	}
	if _, ok := map[string]struct{}{
		graph.NodeService:  {},
		graph.NodeAPI:      {},
		graph.NodeDatabase: {},
		graph.NodeEvent:    {},
	}[target.Type]; !ok {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "node", Message: fmt.Sprintf("unsupported node type %q", target.Type)}
	}
	if _, ok := archStatuses[a.fields["status"]]; a.fields["status"] != "" && !ok {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "status", Message: fmt.Sprintf("unsupported status %q", a.fields["status"])}
	}
	return nil
}

func validateEventBlock(a *blockAccumulator) error {
	for _, key := range []string{"id", "role", "domain"} {
		if strings.TrimSpace(a.fields[key]) == "" {
			return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: key, Message: "missing required tag"}
		}
	}
	if !dotIDPattern.MatchString(a.fields["id"]) {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "id", Message: fmt.Sprintf("invalid event id %q", a.fields["id"])}
	}

	role := a.fields["role"]
	if _, ok := eventRoles[role]; !ok {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "role", Message: fmt.Sprintf("unsupported event role %q", role)}
	}
	if phase := a.fields["phase"]; phase != "" {
		if _, ok := eventPhases[phase]; !ok {
			return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "phase", Message: fmt.Sprintf("unsupported event phase %q", phase)}
		}
	}

	if role == "trigger" && strings.TrimSpace(a.fields["producer"]) == "" {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "producer", Message: "missing required tag for trigger role"}
	}
	if role == "listener" && strings.TrimSpace(a.fields["consumer"]) == "" {
		return &ParseError{File: a.file, Line: a.line, Namespace: a.namespace, Key: "consumer", Message: "missing required tag for listener role"}
	}
	return nil
}

func parseTargetRef(value string) (TargetRef, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return TargetRef{}, fmt.Errorf("expected <type> <id>, got %q", value)
	}
	if !nodeIDPattern.MatchString(parts[1]) {
		return TargetRef{}, fmt.Errorf("invalid target id %q", parts[1])
	}
	switch parts[0] {
	case graph.NodeService, graph.NodeAPI, graph.NodeDatabase, graph.NodeEvent:
	default:
		return TargetRef{}, fmt.Errorf("unsupported node type %q", parts[0])
	}
	return TargetRef{Type: parts[0], ID: parts[1]}, nil
}

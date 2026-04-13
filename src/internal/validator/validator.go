// Package validator turns scanned raw blocks plus catalog metadata into
// a normalized graph and enforces validation layers 4-6.
package validator

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
)

const (
	severityError   = "error"
	severityWarning = "warning"
)

// Diagnostic is a machine-readable validation issue.
type Diagnostic struct {
	Severity string `json:"severity"`
	Layer    int    `json:"layer"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// Result is the output of the validator pipeline.
type Result struct {
	Graph       graph.Graph  `json:"graph"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// BuildOptions control metadata attached to the normalized graph.
type BuildOptions struct {
	SourceRoot     string
	GeneratedAt    time.Time
	ScannerVersion string
	Scoped         bool
}

type eventModel struct {
	Node                graph.Node
	Deprecated          bool
	hasDefinition       bool
	hasProducerMetadata bool
}

type eventModelSet struct {
	byID             map[string]eventModel
	aliases          map[string]string
	archByLocation   map[string]scanner.RawBlock
	pairedByLocation map[string]string
}

// ValidationError reports one or more validation errors.
type ValidationError struct {
	Result *Result
}

func (e *ValidationError) Error() string {
	if e == nil || e.Result == nil {
		return "validation failed"
	}

	lines := make([]string, 0, len(e.Result.Diagnostics))
	for _, diagnostic := range e.Result.Diagnostics {
		if diagnostic.Severity != severityError {
			continue
		}

		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "layer %d %s", diagnostic.Layer, diagnostic.Code)
		if diagnostic.File != "" {
			_, _ = fmt.Fprintf(&b, " %s", diagnostic.File)
			if diagnostic.Line > 0 {
				_, _ = fmt.Fprintf(&b, ":%d", diagnostic.Line)
			}
		}
		_, _ = fmt.Fprintf(&b, ": %s", diagnostic.Message)
		lines = append(lines, b.String())
	}

	if len(lines) == 0 {
		return "validation failed"
	}
	return strings.Join(lines, "\n")
}

// Build validates scanned blocks against catalog/config state and returns
// the normalized graph plus any warnings.
func Build(cfg *config.Config, cat *catalog.Catalog, blocks []scanner.RawBlock, opts ...BuildOptions) (*Result, error) {
	result := &Result{}
	if cfg == nil {
		return result, fmt.Errorf("config is nil")
	}
	if cat == nil {
		return result, fmt.Errorf("catalog is nil")
	}

	builder := graph.NewBuilder()
	buildOptions := BuildOptions{}
	if len(opts) > 0 {
		buildOptions = opts[0]
	}
	eventModels := collectEventModels(blocks, result)
	fileNodes := make(map[string][]graph.Node)
	requiredNodeRefs := collectScopedRefs(blocks, eventModels.aliases)
	allowedTags := allowedTagSet(cfg.Tags)

	validateCatalogTags(result, cat, allowedTags)
	validateBlockTags(result, blocks, allowedTags)
	validateBlockFacets(result, blocks, cfg.Facets)

	for _, event := range eventModels.byID {
		node := applyEffectiveTags(event.Node, cat)
		if err := builder.AddNode(node); err != nil {
			addError(result, 6, "duplicate_node_id", err.Error(), node.File, node.Line)
		}
	}

	for _, block := range blocks {
		if block.Kind != "arch" {
			continue
		}
		if _, paired := eventModels.pairedByLocation[blockLocationKey(block.File, block.Line)]; paired {
			continue
		}
		node, err := buildArchNode(block, eventModels.aliases)
		if err != nil {
			addError(result, 6, "invalid_node", err.Error(), block.File, block.Line)
			continue
		}

		checkDomain(result, cfg.Validation.FailOnUnknownDomain, block.File, block.Line, node.Domain, cat)
		checkOwner(result, cfg.Validation.FailOnUnknownTeam, block.File, block.Line, node.Owner, cat)
		node = applyEffectiveTags(node, cat)

		if err := builder.AddNode(node); err != nil {
			addError(result, 6, "duplicate_node_id", err.Error(), block.File, block.Line)
			continue
		}
		fileNodes[node.File] = append(fileNodes[node.File], node)
	}

	for file := range fileNodes {
		sort.Slice(fileNodes[file], func(i, j int) bool {
			return fileNodes[file][i].Line < fileNodes[file][j].Line
		})
	}

	for _, block := range blocks {
		switch block.Kind {
		case "arch":
			if _, paired := eventModels.pairedByLocation[blockLocationKey(block.File, block.Line)]; paired {
				continue
			}
			sourceNode, err := buildArchNode(block, eventModels.aliases)
			if err != nil {
				continue
			}
			for relation, targets := range block.Relations {
				for _, target := range targets {
					targetRef := nodeRef{Type: target.Type, ID: target.ID}
					if targetRef.Type == graph.NodeEvent {
						targetRef.ID = canonicalEventID(targetRef.ID, eventModels.aliases)
					}
					targetNodeID := nodeIDForRef(targetRef)
					if buildOptions.Scoped {
						requiredNodeRefs[targetNodeID] = targetRef
					}
					builder.AddEdge(graph.Edge{
						From: sourceNode.ID,
						To:   targetNodeID,
						Type: relation,
					})
				}
			}
		case "event":
			eventID := block.Fields["id"]
			model := eventModels.byID[eventID]
			validateEventBlock(result, cfg, cat, block, model, eventModels.archByLocation[blockLocationKey(block.File, block.Line)])

			relation, ok := eventRelation(block.Fields["role"])
			if !ok {
				continue
			}

			source, found := nearestFileNode(fileNodes[block.File], block.Line)
			if !found {
				addWarning(result, 5, "unattached_event", "event block could not be attached to a nearby architecture node", block.File, block.Line)
				continue
			}

			from := source.ID
			to := eventNodeID(eventID)
			if relation.FromEvent {
				from, to = to, from
			}
			if buildOptions.Scoped {
				requiredNodeRefs[eventNodeID(eventID)] = nodeRef{
					Type: graph.NodeEvent,
					ID:   eventID,
				}
			}

			builder.AddEdge(graph.Edge{
				From: from,
				To:   to,
				Type: relation.EdgeType,
			})
		}
	}

	if buildOptions.Scoped {
		synthesizeScopedBoundaryNodes(builder, requiredNodeRefs)
	}

	result.Graph = builder.Build(graph.NewMetadata(buildOptions.SourceRoot, buildOptions.GeneratedAt, buildOptions.ScannerVersion))

	if cfg.Validation.WarnOnOrphanedNodes {
		reportOrphanedNodes(result, result.Graph)
	}
	validateEdgeTargets(result, cfg, result.Graph)
	sortDiagnostics(result.Diagnostics)

	if hasErrors(result.Diagnostics) {
		return result, &ValidationError{Result: result}
	}

	return result, nil
}

func buildArchNode(block scanner.RawBlock, eventAliases map[string]string) (graph.Node, error) {
	ref, err := parseNodeRef(block.Fields["node"])
	if err != nil {
		return graph.Node{}, err
	}
	if ref.Type == graph.NodeEvent {
		ref.ID = canonicalEventID(ref.ID, eventAliases)
	}

	return graph.Node{
		ID:      nodeIDForRef(ref),
		Type:    ref.Type,
		Name:    block.Fields["name"],
		Domain:  block.Fields["domain"],
		Owner:   block.Fields["owner"],
		File:    block.File,
		Line:    block.Line,
		Summary: block.Fields["description"],
		Tags:    parseTagList(block.Fields["tags"]),
		Facets:  extractFacetAssignments(block.Fields),
	}, nil
}

type nodeRef struct {
	Type string
	ID   string
}

func parseNodeRef(value string) (nodeRef, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return nodeRef{}, fmt.Errorf("expected <type> <id>, got %q", value)
	}
	return nodeRef{Type: parts[0], ID: parts[1]}, nil
}

func eventNodeID(eventID string) string {
	return graph.NodeEvent + ":" + eventID
}

func nodeIDForRef(ref nodeRef) string {
	if ref.Type == graph.NodeEvent {
		return eventNodeID(ref.ID)
	}
	return fmt.Sprintf("%s:%s", ref.Type, ref.ID)
}

func blockLocationKey(file string, line int) string {
	return fmt.Sprintf("%s:%d", file, line)
}

func canonicalEventID(id string, aliases map[string]string) string {
	if canonical, ok := aliases[id]; ok {
		return canonical
	}
	return id
}

func collectEventModels(blocks []scanner.RawBlock, result *Result) eventModelSet {
	models := eventModelSet{
		byID:             make(map[string]eventModel),
		aliases:          make(map[string]string),
		archByLocation:   collectEventArchBlocks(blocks),
		pairedByLocation: make(map[string]string),
	}

	for _, block := range blocks {
		if block.Kind != "event" {
			continue
		}

		eventID := block.Fields["id"]
		if eventID == "" {
			continue
		}

		model := ensureEventModel(models.byID[eventID], eventID)
		location := blockLocationKey(block.File, block.Line)
		if archBlock, ok := models.archByLocation[location]; ok {
			model = attachPairedArchMetadata(model, eventID, archBlock, location, &models, result)
		}
		model.Node.Tags = mergeTags(model.Node.Tags, parseTagList(block.Fields["tags"]))
		model.Node.Facets = mergeFacetAssignments(
			model.Node.Facets,
			extractFacetAssignments(block.Fields),
			block.File,
			block.Line,
			result,
		)
		model = applyEventRoleMetadata(model, block)
		models.byID[eventID] = model
	}

	return models
}

func collectEventArchBlocks(blocks []scanner.RawBlock) map[string]scanner.RawBlock {
	archByLocation := make(map[string]scanner.RawBlock)
	for _, block := range blocks {
		if block.Kind != "arch" {
			continue
		}
		ref, err := parseNodeRef(block.Fields["node"])
		if err != nil || ref.Type != graph.NodeEvent {
			continue
		}
		archByLocation[blockLocationKey(block.File, block.Line)] = block
	}
	return archByLocation
}

func ensureEventModel(model eventModel, eventID string) eventModel {
	if model.Node.ID != "" {
		return model
	}
	model.Node = graph.Node{
		ID:   eventNodeID(eventID),
		Type: graph.NodeEvent,
		Name: fallbackNodeName(eventID),
	}
	return model
}

func attachPairedArchMetadata(model eventModel, eventID string, archBlock scanner.RawBlock, location string, models *eventModelSet, result *Result) eventModel {
	ref, err := parseNodeRef(archBlock.Fields["node"])
	if err == nil && ref.Type == graph.NodeEvent {
		models.aliases[ref.ID] = eventID
		models.pairedByLocation[location] = eventID
	}
	if archBlock.Fields["name"] != "" {
		model.Node.Name = archBlock.Fields["name"]
	}
	if archBlock.Fields["description"] != "" {
		model.Node.Summary = archBlock.Fields["description"]
	}
	if archBlock.Fields["status"] == "deprecated" {
		model.Deprecated = true
	}
	model.Node.Tags = mergeTags(model.Node.Tags, parseTagList(archBlock.Fields["tags"]))
	model.Node.Facets = mergeFacetAssignments(
		model.Node.Facets,
		extractFacetAssignments(archBlock.Fields),
		archBlock.File,
		archBlock.Line,
		result,
	)
	if model.Node.File == "" {
		model.Node.File = archBlock.File
		model.Node.Line = archBlock.Line
	}
	return model
}

func applyEventRoleMetadata(model eventModel, block scanner.RawBlock) eventModel {
	switch block.Fields["role"] {
	case "definition":
		model.hasDefinition = true
		model.Node.Domain = block.Fields["domain"]
		if owner := block.Fields["owner"]; owner != "" {
			model.Node.Owner = owner
		}
		if notes := block.Fields["notes"]; notes != "" && model.Node.Summary == "" {
			model.Node.Summary = notes
		}
		model.Node.File = block.File
		model.Node.Line = block.Line
	case "trigger", "publisher", "bridge-out":
		if !model.hasDefinition && !model.hasProducerMetadata {
			applyEventFallbackMetadata(&model, block)
		}
		model.hasProducerMetadata = true
	default:
		if model.Node.Domain == "" || model.Node.Owner == "" || model.Node.File == "" || model.Node.Summary == "" {
			applyEventFallbackMetadata(&model, block)
		}
	}
	return model
}

func applyEventFallbackMetadata(model *eventModel, block scanner.RawBlock) {
	if model.Node.Domain == "" {
		model.Node.Domain = block.Fields["domain"]
	}
	if model.Node.Owner == "" {
		model.Node.Owner = block.Fields["owner"]
	}
	if notes := block.Fields["notes"]; notes != "" && model.Node.Summary == "" {
		model.Node.Summary = notes
	}
	if model.Node.File == "" {
		model.Node.File = block.File
		model.Node.Line = block.Line
	}
}

func collectScopedRefs(blocks []scanner.RawBlock, eventAliases map[string]string) map[string]nodeRef {
	requiredNodeRefs := make(map[string]nodeRef)

	for _, block := range blocks {
		if block.Kind == "event" {
			eventID := block.Fields["id"]
			if eventID == "" {
				continue
			}
			requiredNodeRefs[eventNodeID(eventID)] = nodeRef{
				Type: graph.NodeEvent,
				ID:   eventID,
			}
		}

		for _, targets := range block.Relations {
			for _, target := range targets {
				ref := nodeRef{
					Type: target.Type,
					ID:   target.ID,
				}
				if ref.Type == graph.NodeEvent {
					ref.ID = canonicalEventID(ref.ID, eventAliases)
				}
				requiredNodeRefs[nodeIDForRef(ref)] = ref
			}
		}
	}

	return requiredNodeRefs
}

func checkDomain(result *Result, fail bool, file string, line int, domainID string, cat *catalog.Catalog) {
	if _, ok := cat.DomainsByID[domainID]; ok {
		return
	}
	if fail {
		addError(result, 4, "unknown_domain", fmt.Sprintf("unknown domain %q", domainID), file, line)
		return
	}
	addWarning(result, 4, "unknown_domain", fmt.Sprintf("unknown domain %q", domainID), file, line)
}

func checkOwner(result *Result, fail bool, file string, line int, ownerID string, cat *catalog.Catalog) {
	if _, ok := cat.TeamsByID[ownerID]; ok {
		return
	}
	if fail {
		addError(result, 4, "unknown_team", fmt.Sprintf("unknown team %q", ownerID), file, line)
		return
	}
	addWarning(result, 4, "unknown_team", fmt.Sprintf("unknown team %q", ownerID), file, line)
}

func validateCatalogTags(result *Result, cat *catalog.Catalog, allowed map[string]struct{}) {
	for _, team := range cat.Teams {
		for _, tag := range team.Tags {
			if _, ok := allowed[tag]; ok {
				continue
			}
			addError(result, 4, "unknown_tag", fmt.Sprintf("team %q references unknown tag %q", team.ID, tag), "", 0)
		}
	}
	for _, domain := range cat.Domains {
		for _, tag := range domain.Tags {
			if _, ok := allowed[tag]; ok {
				continue
			}
			addError(result, 4, "unknown_tag", fmt.Sprintf("domain %q references unknown tag %q", domain.ID, tag), "", 0)
		}
	}
}

func validateBlockTags(result *Result, blocks []scanner.RawBlock, allowed map[string]struct{}) {
	for _, block := range blocks {
		for _, tag := range parseTagList(block.Fields["tags"]) {
			if _, ok := allowed[tag]; ok {
				continue
			}
			addError(result, 4, "unknown_tag", fmt.Sprintf("unknown tag %q", tag), block.File, block.Line)
		}
	}
}

func validateBlockFacets(result *Result, blocks []scanner.RawBlock, definitions config.Facets) {
	for _, block := range blocks {
		for key, value := range extractFacetAssignments(block.Fields) {
			definition, ok := definitions[key]
			if !ok {
				addError(result, 4, "unknown_facet_key", fmt.Sprintf("unknown facet key %q", key), block.File, block.Line)
				continue
			}
			if containsFacetValue(definition.Values, value) {
				continue
			}
			addError(result, 4, "unknown_facet_value", fmt.Sprintf("facet %q does not allow value %q", key, value), block.File, block.Line)
		}
	}
}

func allowedTagSet(tags []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		set[tag] = struct{}{}
	}
	return set
}

func applyEffectiveTags(node graph.Node, cat *catalog.Catalog) graph.Node {
	effective := append([]string(nil), node.Tags...)
	if domain, ok := cat.DomainsByID[node.Domain]; ok {
		effective = mergeTags(effective, domain.Tags)
	}
	if team, ok := cat.TeamsByID[node.Owner]; ok {
		effective = mergeTags(effective, team.Tags)
	}
	node.Tags = normalizeTags(node.Tags)
	node.EffectiveTags = normalizeTags(effective)
	return node
}

func extractFacetAssignments(fields map[string]string) map[string]string {
	if len(fields) == 0 {
		return nil
	}

	assignments := make(map[string]string)
	for key, value := range fields {
		if !strings.Contains(key, ".") {
			continue
		}
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized == "" {
			continue
		}
		assignments[key] = normalized
	}
	if len(assignments) == 0 {
		return nil
	}
	return assignments
}

func mergeFacetAssignments(base map[string]string, extra map[string]string, file string, line int, result *Result) map[string]string {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}

	merged := make(map[string]string, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		if existing, ok := merged[key]; ok && existing != value {
			addError(result, 4, "conflicting_facet_value", fmt.Sprintf("facet %q is already set to %q and cannot also be %q", key, existing, value), file, line)
			continue
		}
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func containsFacetValue(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func parseTagList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(strings.ToLower(part))
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	return normalizeTags(tags)
}

func mergeTags(base []string, extra []string) []string {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}
	combined := append([]string(nil), base...)
	combined = append(combined, extra...)
	return normalizeTags(combined)
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	sort.Strings(normalized)
	return normalized
}

func validateEventBlock(result *Result, cfg *config.Config, cat *catalog.Catalog, block scanner.RawBlock, model eventModel, pairedArch scanner.RawBlock) {
	checkDomain(result, cfg.Validation.FailOnUnknownDomain, block.File, block.Line, block.Fields["domain"], cat)
	if owner := block.Fields["owner"]; owner != "" {
		checkOwner(result, cfg.Validation.FailOnUnknownTeam, block.File, block.Line, owner, cat)
	}

	if block.Fields["role"] == "definition" && pairedArch.Kind == "arch" {
		if domain := pairedArch.Fields["domain"]; domain != "" && domain != block.Fields["domain"] {
			addError(result, 4, "event_domain_mismatch", fmt.Sprintf("event %q belongs to domain %q, not %q", block.Fields["id"], domain, block.Fields["domain"]), block.File, block.Line)
		}
		if owner := block.Fields["owner"]; owner != "" {
			if archOwner := pairedArch.Fields["owner"]; archOwner != "" && archOwner != owner {
				addError(result, 4, "event_owner_mismatch", fmt.Sprintf("event %q owner should be %q, not %q", block.Fields["id"], archOwner, owner), block.File, block.Line)
			}
		}
	}

	if cfg.Validation.WarnOnDeprecatedEvents && model.Deprecated {
		addWarning(result, 4, "deprecated_event", fmt.Sprintf("event %q is deprecated", block.Fields["id"]), block.File, block.Line)
	}
}

type eventRelationSpec struct {
	EdgeType  string
	FromEvent bool
}

func eventRelation(role string) (eventRelationSpec, bool) {
	switch role {
	case "trigger", "bridge-out", "publisher":
		return eventRelationSpec{EdgeType: graph.EdgeEmits, FromEvent: false}, true
	case "listener", "bridge-in", "subscriber":
		return eventRelationSpec{EdgeType: graph.EdgeConsumes, FromEvent: true}, true
	default:
		return eventRelationSpec{}, false
	}
}

func nearestFileNode(nodes []graph.Node, line int) (graph.Node, bool) {
	var match graph.Node
	found := false
	for _, node := range nodes {
		if node.Line > line {
			break
		}
		match = node
		found = true
	}
	return match, found
}

func reportOrphanedNodes(result *Result, g graph.Graph) {
	linked := make(map[string]struct{}, len(g.Nodes))
	for _, edge := range g.Edges {
		linked[edge.From] = struct{}{}
		linked[edge.To] = struct{}{}
	}
	for _, node := range g.Nodes {
		if _, ok := linked[node.ID]; ok {
			continue
		}
		addWarning(result, 6, "orphaned_node", fmt.Sprintf("node %q has no edges", node.ID), node.File, node.Line)
	}
}

func validateEdgeTargets(result *Result, cfg *config.Config, g graph.Graph) {
	nodeIDs := make(map[string]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		nodeIDs[node.ID] = struct{}{}
	}

	for _, edge := range g.Edges {
		if _, ok := nodeIDs[edge.To]; ok {
			continue
		}
		if cfg.Validation.FailOnUnknownNode {
			addError(result, 6, "unknown_node_target", fmt.Sprintf("edge target %q does not exist", edge.To), "", 0)
			continue
		}
		addWarning(result, 6, "unknown_node_target", fmt.Sprintf("edge target %q does not exist", edge.To), "", 0)
	}
}

func synthesizeScopedBoundaryNodes(builder *graph.Builder, refs map[string]nodeRef) {
	for nodeID, ref := range refs {
		if builder.HasNode(nodeID) {
			continue
		}

		node := graph.Node{
			ID:      nodeID,
			Type:    ref.Type,
			Name:    fallbackNodeName(ref.ID),
			Summary: "Inferred out-of-scope boundary from scoped scan.",
		}
		_ = builder.AddNode(node)
	}
}

func fallbackNodeName(id string) string {
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	if len(parts) == 0 {
		return id
	}
	return strings.Join(parts, " ")
}

func addError(result *Result, layer int, code string, message string, file string, line int) {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{
		Severity: severityError,
		Layer:    layer,
		Code:     code,
		Message:  message,
		File:     file,
		Line:     line,
	})
}

func addWarning(result *Result, layer int, code string, message string, file string, line int) {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{
		Severity: severityWarning,
		Layer:    layer,
		Code:     code,
		Message:  message,
		File:     file,
		Line:     line,
	})
}

func hasErrors(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severityError {
			return true
		}
	}
	return false
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Severity == diagnostics[j].Severity {
			if diagnostics[i].Layer == diagnostics[j].Layer {
				if diagnostics[i].File == diagnostics[j].File {
					if diagnostics[i].Line == diagnostics[j].Line {
						return diagnostics[i].Code < diagnostics[j].Code
					}
					return diagnostics[i].Line < diagnostics[j].Line
				}
				return diagnostics[i].File < diagnostics[j].File
			}
			return diagnostics[i].Layer < diagnostics[j].Layer
		}
		return diagnostics[i].Severity < diagnostics[j].Severity
	})
}

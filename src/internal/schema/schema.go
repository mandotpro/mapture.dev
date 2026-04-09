// Package schema embeds the CUE definitions used to validate YAML inputs.
package schema

import (
	"embed"
	"fmt"
	"sort"
	"strings"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	cuejson "cuelang.org/go/encoding/json"
	cueyaml "cuelang.org/go/encoding/yaml"
)

//go:embed *.cue
var schemaFS embed.FS

// Definition names an embedded top-level CUE definition.
type Definition string

// Embedded schema entrypoints used by the YAML decoders.
const (
	ConfigDefinition   Definition = "Config"
	TeamsDefinition    Definition = "TeamsFile"
	DomainsDefinition  Definition = "DomainsFile"
	EventsDefinition   Definition = "EventsFile"
	GraphDefinition    Definition = "Graph"
	ExplorerDefinition Definition = "ExplorerPayload"
)

var (
	registryOnce sync.Once
	registryInst *registry
	registryErr  error
)

type registry struct {
	ctx  *cue.Context
	root cue.Value
}

// DecodeYAML validates a YAML payload against an embedded CUE definition and
// decodes the validated value into out.
func DecodeYAML(def Definition, filename string, src []byte, out any) error {
	r, err := defaultRegistry()
	if err != nil {
		return err
	}
	return r.decodeYAML(def, filename, src, out)
}

// ValidateJSON validates a JSON payload against an embedded CUE definition.
func ValidateJSON(def Definition, filename string, src []byte) error {
	r, err := defaultRegistry()
	if err != nil {
		return err
	}
	return r.validateJSON(def, filename, src)
}

func defaultRegistry() (*registry, error) {
	registryOnce.Do(func() {
		ctx := cuecontext.New()
		root, err := compileSchema(ctx)
		if err != nil {
			registryErr = err
			return
		}

		registryInst = &registry{
			ctx:  ctx,
			root: root,
		}
	})

	return registryInst, registryErr
}

func compileSchema(ctx *cue.Context) (cue.Value, error) {
	entries, err := schemaFS.ReadDir(".")
	if err != nil {
		return cue.Value{}, fmt.Errorf("read embedded schema files: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cue") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	overlay := make(map[string]load.Source, len(names))
	for _, name := range names {
		data, err := schemaFS.ReadFile(name)
		if err != nil {
			return cue.Value{}, fmt.Errorf("read embedded schema %s: %w", name, err)
		}
		overlay["/schema/"+name] = load.FromBytes(data)
	}

	instances := load.Instances([]string{"."}, &load.Config{
		Dir:     "/schema",
		Module:  "schema",
		Overlay: overlay,
	})
	if len(instances) != 1 {
		return cue.Value{}, fmt.Errorf("load embedded schema package: expected 1 instance, got %d", len(instances))
	}
	if instances[0].Err != nil {
		return cue.Value{}, fmt.Errorf("load embedded schema package: %w", instances[0].Err)
	}

	root := ctx.BuildInstance(instances[0])
	if err := root.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("compile embedded schema: %w", err)
	}

	return root, nil
}

func (r *registry) decodeYAML(def Definition, filename string, src []byte, out any) error {
	file, err := cueyaml.Extract(filename, src)
	if err != nil {
		return formatError(filename, err)
	}

	data := r.ctx.BuildFile(file)
	if err := data.Err(); err != nil {
		return formatError(filename, err)
	}

	schemaValue := r.root.LookupPath(cue.MakePath(cue.Def(string(def))))
	if err := schemaValue.Err(); err != nil {
		return fmt.Errorf("lookup schema %s: %w", def, err)
	}

	validated := schemaValue.Unify(data).Eval()
	if err := validated.Validate(cue.Final(), cue.Concrete(true)); err != nil {
		return formatError(filename, err)
	}

	if defaulted, ok := validated.Default(); ok {
		validated = defaulted
	}
	if err := validated.Decode(out); err != nil {
		return formatError(filename, err)
	}

	return nil
}

func (r *registry) validateJSON(def Definition, filename string, src []byte) error {
	expr, err := cuejson.Extract(filename, src)
	if err != nil {
		return formatError(filename, err)
	}
	data := r.ctx.BuildExpr(expr)
	if err := data.Err(); err != nil {
		return formatError(filename, err)
	}

	schemaValue := r.root.LookupPath(cue.MakePath(cue.Def(string(def))))
	if err := schemaValue.Err(); err != nil {
		return fmt.Errorf("lookup schema %s: %w", def, err)
	}

	validated := schemaValue.Unify(data).Eval()
	if err := validated.Validate(cue.Final(), cue.Concrete(true)); err != nil {
		return formatError(filename, err)
	}

	return nil
}

func formatError(filename string, err error) error {
	details := strings.TrimSpace(cueerrors.Details(err, nil))
	if details == "" {
		return fmt.Errorf("%s: %w", filename, err)
	}
	return fmt.Errorf("%s:\n%s", filename, details)
}

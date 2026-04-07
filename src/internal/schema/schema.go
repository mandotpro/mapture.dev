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
	cueyaml "cuelang.org/go/encoding/yaml"
)

//go:embed *.cue
var schemaFS embed.FS

type Definition string

const (
	ConfigDefinition  Definition = "Config"
	TeamsDefinition   Definition = "TeamsFile"
	DomainsDefinition Definition = "DomainsFile"
	EventsDefinition  Definition = "EventsFile"
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

	root := ctx.CompileString("{}")
	if err := root.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("compile empty schema root: %w", err)
	}

	for _, name := range names {
		data, err := schemaFS.ReadFile(name)
		if err != nil {
			return cue.Value{}, fmt.Errorf("read embedded schema %s: %w", name, err)
		}

		value := ctx.CompileString(string(data), cue.Filename(name))
		if err := value.Err(); err != nil {
			return cue.Value{}, fmt.Errorf("compile embedded schema: %w", err)
		}
		root = root.Unify(value)
	}

	if err := root.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("unify embedded schema: %w", err)
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

func formatError(filename string, err error) error {
	details := strings.TrimSpace(cueerrors.Details(err, nil))
	if details == "" {
		return fmt.Errorf("%s: %w", filename, err)
	}
	return fmt.Errorf("%s:\n%s", filename, details)
}

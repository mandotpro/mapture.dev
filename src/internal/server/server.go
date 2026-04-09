// Package server hosts the `mapture serve` local explorer.
//
// It exposes a small JSON API over the scanner/validator pipeline,
// embeds a minimal HTML/JS UI, and (optionally) watches source files
// for changes, broadcasting reloads to any connected browser over
// Server-Sent Events.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/projectscope"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
	"github.com/mandotpro/mapture.dev/src/internal/webui"
)

// DefaultAddr is the address used when the caller does not override it.
const DefaultAddr = "127.0.0.1:8765"

// Options configure a running explorer.
type Options struct {
	// ConfigPath is the absolute path to the project's mapture.yaml.
	ConfigPath string
	// Addr is the listen address (e.g. "127.0.0.1:8765").
	Addr string
	// Scopes narrows scanning to one or more project-relative files/directories.
	Scopes []string
	// Watch enables fsnotify-based file watching + SSE reload.
	Watch bool
	// OnReady is invoked once the listener is bound, with the concrete
	// base URL. Used by tests and by the --open flag plumbing.
	OnReady func(url string)
}

// Serve boots the explorer and blocks until ctx is cancelled.
func Serve(ctx context.Context, opts Options) error {
	if opts.ConfigPath == "" {
		return errors.New("server: ConfigPath is required")
	}
	addr := opts.Addr
	if addr == "" {
		addr = DefaultAddr
	}

	srv, err := newServer(opts.ConfigPath, opts.Scopes)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	srv.register(mux)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	httpSrv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if opts.OnReady != nil {
		opts.OnReady("http://" + listener.Addr().String())
	}

	var watcherWG sync.WaitGroup
	watchCtx, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	if opts.Watch {
		watcherWG.Add(1)
		go func() {
			defer watcherWG.Done()
			if err := srv.watch(watchCtx); err != nil && !errors.Is(err, context.Canceled) {
				// Non-fatal: log to stderr-style via broadcaster? Just swallow —
				// live reload is best-effort.
				_ = err
			}
		}()
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := httpSrv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		cancelWatch()
		watcherWG.Wait()
		srv.broadcaster.close()
		return err
	}

	cancelWatch()
	watcherWG.Wait()
	// Close SSE subscriptions before shutting down the HTTP server so
	// long-lived /api/events connections do not block graceful exit.
	srv.broadcaster.close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdownErr := httpSrv.Shutdown(shutdownCtx)

	if shutdownErr != nil {
		return fmt.Errorf("shutdown: %w", shutdownErr)
	}
	return nil
}

type explorer struct {
	configPath  string
	scopes      []string
	uiHandler   http.Handler
	broadcaster *broadcaster
}

func newServer(configPath string, scopes []string) (*explorer, error) {
	return &explorer{
		configPath:  configPath,
		scopes:      append([]string(nil), scopes...),
		uiHandler:   http.FileServer(http.FS(webui.FS())),
		broadcaster: newBroadcaster(),
	}, nil
}

func (e *explorer) register(mux *http.ServeMux) {
	mux.HandleFunc("/api/explorer", e.handleExplorer)
	mux.HandleFunc("/api/graph", e.handleGraph)
	mux.HandleFunc("/api/catalog", e.handleCatalog)
	mux.HandleFunc("/api/validate", e.handleValidate)
	mux.HandleFunc("/api/events", e.handleEvents)
	mux.Handle("/", e.uiHandler)
}

// loadProject re-reads config/catalog on every request so edits to
// mapture.yaml or catalog files are picked up without a restart.
func (e *explorer) loadProject() (*config.Config, *catalog.Catalog, string, error) {
	cfg, err := config.Load(e.configPath)
	if err != nil {
		return nil, nil, "", err
	}
	catalogDir, err := cfg.CatalogDir(e.configPath)
	if err != nil {
		return nil, nil, "", err
	}
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, nil, "", err
	}
	root := filepath.Dir(e.configPath)
	scoped, err := projectscope.Apply(root, cfg, e.scopes)
	if err != nil {
		return nil, nil, "", err
	}
	return scoped.Config, cat, root, nil
}

func (e *explorer) buildResult() (*validator.Result, error) {
	cfg, cat, root, err := e.loadProject()
	if err != nil {
		return nil, err
	}
	blocks, err := scanner.Scan(root, cfg)
	if err != nil {
		return nil, err
	}
	result, err := validator.Build(cfg, cat, blocks, validator.BuildOptions{
		SourceRoot: root,
		Scoped:     len(e.scopes) > 0,
	})
	if err != nil {
		// Return the partial result so the UI can surface diagnostics.
		var vErr *validator.ValidationError
		if errors.As(err, &vErr) && vErr.Result != nil {
			return vErr.Result, nil
		}
		return nil, err
	}
	return result, nil
}

func (e *explorer) buildExplorerPayload() (*ExplorerPayload, error) {
	cfg, cat, root, err := e.loadProject()
	if err != nil {
		return nil, err
	}

	blocks, err := scanner.Scan(root, cfg)
	if err != nil {
		return nil, err
	}

	result, err := validator.Build(cfg, cat, blocks, validator.BuildOptions{
		SourceRoot: root,
		Scoped:     len(e.scopes) > 0,
	})
	if err != nil {
		var vErr *validator.ValidationError
		if !errors.As(err, &vErr) || vErr.Result == nil {
			return nil, err
		}
		result = vErr.Result
	}

	return &ExplorerPayload{
		SchemaVersion: explorerPayloadSchemaVersion,
		Graph:         result.Graph,
		Catalog: ExplorerCatalog{
			Teams:   cat.Teams,
			Domains: cat.Domains,
			Events:  cat.Events,
		},
		Validation: ExplorerValidation{
			Diagnostics: result.Diagnostics,
			Summary: ValidationSummary{
				Errors:   countDiagnostics(result.Diagnostics, "error"),
				Warnings: countDiagnostics(result.Diagnostics, "warning"),
				Nodes:    len(result.Graph.Nodes),
				Edges:    len(result.Graph.Edges),
			},
		},
		UI: cfg.UI,
		Meta: ExplorerMeta{
			ProjectID:   filepath.Dir(e.configPath),
			SourceLabel: projectscope.SourceLabel("live api", scopedIncludes(cfg, len(e.scopes) > 0)),
			Mode:        "live",
		},
	}, nil
}

func scopedIncludes(cfg *config.Config, scoped bool) []string {
	if !scoped {
		return nil
	}
	return append([]string(nil), cfg.Scan.Include...)
}

func (e *explorer) handleExplorer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload, err := e.buildExplorerPayload()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, payload)
}

func (e *explorer) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	result, err := e.buildResult()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, result.Graph)
}

func (e *explorer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	result, err := e.buildResult()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, result)
}

func (e *explorer) handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, cat, _, err := e.loadProject()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, ExplorerCatalog{
		Teams:   cat.Teams,
		Domains: cat.Domains,
		Events:  cat.Events,
	})
}

func (e *explorer) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := e.broadcaster.subscribe()
	defer e.broadcaster.unsubscribe(ch)

	// Prime the stream so clients know they're connected.
	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ctx := r.Context()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case msg, open := <-ch:
			if !open {
				return
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg.event, msg.data); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (e *explorer) watch(ctx context.Context) error {
	cfg, _, root, err := e.loadProject()
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	if err := addWatchRoots(watcher, root, cfg); err != nil {
		return err
	}
	// Always watch the config file itself.
	_ = watcher.Add(e.configPath)

	const debounce = 250 * time.Millisecond
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	fire := func() {
		e.broadcaster.publish(sseMessage{event: "graph", data: "reload"})
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Newly created directories should also be watched so deep
			// edits keep triggering reloads.
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = watcher.Add(event.Name)
				}
			}
			if timer == nil {
				timer = time.AfterFunc(debounce, fire)
			} else {
				timer.Reset(debounce)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			_ = err
		}
	}
}

func countDiagnostics(diagnostics []validator.Diagnostic, severity string) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severity {
			count++
		}
	}
	return count
}

func addWatchRoots(watcher *fsnotify.Watcher, root string, cfg *config.Config) error {
	includes := cfg.Scan.Include
	if len(includes) == 0 {
		includes = []string{"."}
	}
	for _, include := range includes {
		base := include
		if !filepath.IsAbs(base) {
			base = filepath.Join(root, include)
		}
		base = filepath.Clean(base)
		if err := filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				// Non-fatal: skip unreachable entries.
				return nil
			}
			if d.IsDir() {
				return watcher.Add(path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("walk watch root %s: %w", include, err)
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

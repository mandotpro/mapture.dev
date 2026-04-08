package server

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

func ecommerceConfig(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	path, err := filepath.Abs(filepath.Join(filepath.Dir(file), "..", "..", "..", "examples", "ecommerce", "mapture.yaml"))
	if err != nil {
		t.Fatalf("resolve ecommerce config: %v", err)
	}
	return path
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func startTestServer(t *testing.T, watch bool) (baseURL string, cancel func()) {
	t.Helper()
	addr := freePort(t)

	ctx, ctxCancel := context.WithCancel(context.Background())
	ready := make(chan string, 1)
	done := make(chan error, 1)

	go func() {
		done <- Serve(ctx, Options{
			ConfigPath: ecommerceConfig(t),
			Addr:       addr,
			Watch:      watch,
			OnReady: func(url string) {
				select {
				case ready <- url:
				default:
				}
			},
		})
	}()

	select {
	case url := <-ready:
		return url, func() {
			ctxCancel()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("server returned error: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Error("server did not shut down within 5s")
			}
		}
	case err := <-done:
		t.Fatalf("server exited before ready: %v", err)
	case <-time.After(5 * time.Second):
		ctxCancel()
		t.Fatal("server did not become ready within 5s")
	}
	return "", func() {}
}

func TestServeGraphEndpointReturnsGraph(t *testing.T) {
	baseURL, stop := startTestServer(t, false)
	defer stop()

	resp, err := http.Get(baseURL + "/api/graph")
	if err != nil {
		t.Fatalf("GET /api/graph: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	var g graph.Graph
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		t.Fatalf("decode graph: %v", err)
	}
	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes in graph")
	}
}

func TestServeValidateEndpointReturnsResult(t *testing.T) {
	baseURL, stop := startTestServer(t, false)
	defer stop()

	resp, err := http.Get(baseURL + "/api/validate")
	if err != nil {
		t.Fatalf("GET /api/validate: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	var result validator.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if len(result.Graph.Nodes) == 0 {
		t.Fatal("expected graph in validator result")
	}
}

func TestServeCatalogEndpoint(t *testing.T) {
	baseURL, stop := startTestServer(t, false)
	defer stop()

	resp, err := http.Get(baseURL + "/api/catalog")
	if err != nil {
		t.Fatalf("GET /api/catalog: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	var payload struct {
		Teams   []map[string]any `json:"teams"`
		Domains []map[string]any `json:"domains"`
		Events  []map[string]any `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode catalog: %v", err)
	}
	if len(payload.Teams) == 0 || len(payload.Domains) == 0 {
		t.Fatalf("expected populated catalog, got %+v", payload)
	}
}

func TestServeIndexHTML(t *testing.T) {
	baseURL, stop := startTestServer(t, false)
	defer stop()

	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "Mapture Explorer") {
		t.Fatalf("index.html missing expected content: %s", body)
	}
}

func TestServeWatchBroadcastsOnFileChange(t *testing.T) {
	// Copy the ecommerce fixture into a tmp dir so we can mutate files
	// without touching the source tree.
	tmp := t.TempDir()
	if err := copyDir(t, filepath.Dir(ecommerceConfig(t)), tmp); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	configPath := filepath.Join(tmp, "mapture.yaml")

	addr := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, Options{
			ConfigPath: configPath,
			Addr:       addr,
			Watch:      true,
			OnReady: func(url string) {
				select {
				case ready <- url:
				default:
				}
			},
		})
	}()

	var baseURL string
	select {
	case baseURL = <-ready:
	case err := <-done:
		t.Fatalf("server exited before ready: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("server not ready")
	}

	// Open the SSE stream.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/events", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	reader := bufio.NewReader(resp.Body)

	gotEvent := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "event: ") {
				select {
				case gotEvent <- strings.TrimSpace(strings.TrimPrefix(line, "event: ")):
				default:
				}
				return
			}
		}
	}()

	// Give fsnotify a beat to register its initial watchers.
	time.Sleep(150 * time.Millisecond)

	// Touch a source file to trigger a reload event.
	target := findScannableFile(t, filepath.Join(tmp, "src"))
	if err := appendToFile(target, "\n// touch\n"); err != nil {
		t.Fatalf("touch file: %v", err)
	}

	select {
	case name := <-gotEvent:
		if name != "graph" {
			t.Fatalf("unexpected event name: %q", name)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("did not receive graph reload event within 3s")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("server returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down")
	}
	wg.Wait()
}

func TestServeShutsDownCleanlyOnContextCancel(t *testing.T) {
	addr := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, Options{
			ConfigPath: ecommerceConfig(t),
			Addr:       addr,
			Watch:      false,
			OnReady: func(url string) {
				select {
				case ready <- url:
				default:
				}
			},
		})
	}()

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("not ready")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected clean shutdown, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within 5s of cancel")
	}
}

func findScannableFile(t *testing.T, root string) string {
	t.Helper()
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || found != "" || info.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".php", ".ts", ".tsx", ".js", ".jsx":
			found = path
		}
		return nil
	})
	if err != nil || found == "" {
		t.Fatalf("no scannable file under %s: %v", root, err)
	}
	return found
}

func appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(content)
	return err
}

func copyDir(t *testing.T, src, dst string) error {
	t.Helper()
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

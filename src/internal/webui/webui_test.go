package webui

import (
	"io/fs"
	"strings"
	"testing"
)

func TestDistBundleHasExpectedFiles(t *testing.T) {
	want := []string{
		"index.html",
		"app.js",
		"styles.css",
	}
	for _, name := range want {
		if _, err := ReadFile(name); err != nil {
			t.Errorf("missing %s in embedded bundle: %v (run `make web`)", name, err)
		}
	}
}

func TestIndexHTMLReferencesBundle(t *testing.T) {
	data, err := ReadFile("index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	body := string(data)
	for _, needle := range []string{"app.js", "styles.css", "mapture.dev explorer"} {
		if !strings.Contains(body, needle) {
			t.Errorf("index.html missing reference to %q", needle)
		}
	}
}

func TestAppJSContainsExpectedAPI(t *testing.T) {
	data, err := ReadFile("app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	body := string(data)
	for _, needle := range []string{
		"/api/export",
		"./data.json",
		"/api/events",
		"__MAPTURE_DATA__",
		"Attach JSON",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("app.js missing reference to %q", needle)
		}
	}
}

func TestFSWalkable(t *testing.T) {
	count := 0
	err := fs.WalkDir(FS(), ".", func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk fs: %v", err)
	}
	if count < 3 {
		t.Errorf("expected at least 3 files in bundle, got %d", count)
	}
}

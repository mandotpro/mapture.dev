// Package webui hosts the built frontend bundle for the Mapture
// Explorer. It is imported by `src/internal/server` and by the HTML
// exporter so both surfaces ship the same UI.
//
// The dist/ directory is produced by `make web` (see
// scripts/build-web/main.go) and committed so `go build` alone always
// produces a working binary.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// FS returns the dist/ subtree as an fs.FS suitable for http.FileServer
// or template.ParseFS.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// Only reachable if the embed directive is broken at build time.
		panic(err)
	}
	return sub
}

// ReadFile returns a single file from the bundled dist/ tree.
func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(FS(), name)
}

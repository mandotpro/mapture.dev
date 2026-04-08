// Package webui exposes the embedded frontend bundle used by the server.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// FS returns the embedded distribution directory.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// ReadFile reads a single file from the embedded bundle.
func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(FS(), name)
}

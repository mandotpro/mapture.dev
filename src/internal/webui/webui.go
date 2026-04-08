// Package webui exposes the embedded Svelte Flow static distribution directory built by SvelteKit to Go servers and exporters.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// FS returns the isolated embedded file system scoped exactly to the static web/dist output.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// ReadFile provides a shortcut to read raw bytes from bound static UI web assets safely.
func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(FS(), name)
}

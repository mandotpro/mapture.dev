// Mapture — repo-native architecture graph tool.
// See _docs/mapture-dev-prd-v1.md for the full product spec.
package main

import (
	"os"

	"github.com/angelmanchev/mapture/src/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

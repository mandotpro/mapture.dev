package main

import (
	"os"

	"github.com/mandotpro/mapture.dev/src/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

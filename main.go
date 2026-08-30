package main

import (
	"os"

	"github.com/koki-develop/blot/internal/cli"
)

// version is the released version, which goreleaser's default ldflags write
// here. Any other build leaves it empty and settles the version for itself.
var version string

func main() {
	if err := cli.Execute(version); err != nil {
		os.Exit(1)
	}
}

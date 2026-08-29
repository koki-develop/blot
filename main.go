package main

import (
	"os"

	"github.com/koki-develop/blot/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

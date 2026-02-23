package main

import (
	"os"

	"github.com/pfldy2850/claudectx/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

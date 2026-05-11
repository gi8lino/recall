package main

import (
	"fmt"
	"os"

	"github.com/containeroo/recall/internal/app"
)

var version = "dev"

// main starts recall and exits with a non-zero status on failure.
func main() {
	if err := app.Run(version, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

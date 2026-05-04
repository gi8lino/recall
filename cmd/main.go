package main

import (
	"fmt"
	"os"

	"github.com/containeroo/recall/internal/app"
)

// main starts recall and exits with a non-zero status on failure.
func main() {
	if err := app.Run(os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

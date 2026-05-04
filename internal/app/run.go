package app

import (
	"fmt"
	"io"
	"os"

	"github.com/containeroo/recall/internal/runner"
)

// Run creates the process runner and executes recall with the provided arguments.
func Run(stdout, stderr io.Writer, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	binary, err := os.Executable()
	if err != nil {
		binary = "recall"
	}

	runner := runner.NewRunner(
		stdout,
		stderr,
		home,
		binary,
	)

	return runner.Run(args)
}

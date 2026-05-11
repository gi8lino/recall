package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/containeroo/recall/internal/flags"
	"github.com/containeroo/recall/internal/runner"
	"github.com/containeroo/tinyflags"
)

// Run creates the process runner and executes recall with the provided arguments.
func Run(version string, stdout, stderr io.Writer, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	binary, err := os.Executable()
	if err != nil {
		binary = "recall"
	}

	r := runner.NewRunner(
		stdout,
		stderr,
		home,
		binary,
	)

	parsed, err := flags.Parse(args, version, &r)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(stdout, err) // nolint:errcheck
			return nil
		}

		if help, ok := tinyflags.HelpText(err); ok {
			return fmt.Errorf("%s\n%s", err, help)
		}

		return err
	}

	r.SetLogging(parsed.Verbose, parsed.Debug)

	return parsed.Runner.Run(context.Background())
}

package runner

import (
	"errors"
	"io"
	"os"

	"github.com/alecthomas/kong"

	"github.com/containeroo/recall/internal/logger"
)

const version = "dev"

// Runner holds process dependencies used by the application layer.
type Runner struct {
	// input reads command input for interactive commands.
	input io.Reader

	// output writes normal command output.
	output logger.Logger

	// logger writes warnings, informational messages, and debug messages.
	logger logger.Logger

	// home is the current user's home directory.
	home string

	// binary is the executable path used in generated shell snippets.
	binary string
}

// NewRunner creates a runner with the given process dependencies.
func NewRunner(stdout, stderr io.Writer, home, binary string) Runner {
	return Runner{
		input:  os.Stdin,
		output: logger.New(stdout),
		logger: logger.New(stderr),
		home:   home,
		binary: binary,
	}
}

// Run parses CLI arguments and executes the selected recall command.
func (r Runner) Run(args []string) error {
	var cli CLI

	parser, err := kong.New(
		&cli,
		kong.Name("recall"),
		kong.Description("A terminal command palette powered by shell history and pinned commands."),
		kong.Vars{
			"version": version,
		},
		kong.Writers(r.output.Writer(), r.logger.Writer()),
	)
	if err != nil {
		return err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return r.parseError(err)
	}

	r.logger = r.logger.
		WithVerbose(cli.Verbose).
		WithDebug(cli.Debug)

	return ctx.Run(r, &cli)
}

// parseError prints generated Kong usage for parse errors and returns the original error.
func (r Runner) parseError(err error) error {
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) && parseErr.Context != nil {
		_ = parseErr.Context.PrintUsage(true)
	}

	return err
}

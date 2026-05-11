package flags

import (
	"context"
	"os"

	"github.com/containeroo/recall/internal/runner"
	"github.com/containeroo/recall/internal/store"
	"github.com/containeroo/tinyflags"
)

// ParseResult contains one parsed command runner plus global logging flags.
type ParseResult struct {
	Runner  tinyflags.Runner
	Verbose bool
	Debug   bool
}

// Parse builds the recall tinyflags tree, parses args, and returns the selected command runner.
func Parse(args []string, version string, r *runner.Runner) (ParseResult, error) {
	app := tinyflags.NewCommand("recall", tinyflags.ContinueOnError)
	app.Version(version)
	app.RequireCommand()

	defaultDataDir := "~/.local/share/recall"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDataDir = store.DefaultDir(home)
	}

	globals := app.Globals()
	dataDir := globals.String("data-dir", defaultDataDir, "Directory used to store recall data.").
		Value()
	historyFiles := globals.StringSlice("history", nil, "History file to scan; may be provided multiple times.").
		Value()

	historyLimit := globals.Int("history-limit", 2000, "Maximum number of recent history entries to read; zero means unlimited.").Value()
	limit := globals.Int("limit", 20, "Maximum number of search results to print; zero means unlimited.").Short("n").Value()
	verbose := globals.Bool("verbose", false, "Print informational diagnostics.").Short("v").Value()
	debug := globals.Bool("debug", false, "Print debug diagnostics.").Value()

	currentConfig := func() runner.Config {
		return runner.Config{
			DataDir:      *dataDir,
			HistoryFiles: append([]string(nil), (*historyFiles)...),
			HistoryLimit: *historyLimit,
			Limit:        *limit,
		}
	}

	search := app.Command("search", "Search pinned commands and shell history.")
	search.Note("Positional arguments:\n  QUERY...  Search query terms.")

	searchHistoryOnly := search.Bool("history-only", false, "Search shell history commands only.").
		OneOfGroup("search-mode").
		Value()

	searchPinsOnly := search.Bool("pins-only", false, "Search pinned commands only.").
		OneOfGroup("search-mode").
		OneOfGroup("pins-only").
		HelpOneOfGroups("search-mode").
		Value()

	searchExcludePins := search.Bool("exclude-pins", false, "Exclude commands from history results when they are already pinned.").
		OneOfGroup("pins-only").
		Value()

	searchFormat := search.String("format", "plain", "Output format.").
		Choices("plain", "fzf").
		Value()

	search.Run(func(_ context.Context, pinsOnly, historyOnly, excludePins bool, format string) error {
		return runner.RunSearch(*r, currentConfig(), runner.SearchOptions{
			PinsOnly:    pinsOnly,
			HistoryOnly: historyOnly,
			ExcludePins: excludePins,
			Format:      format,
			Query:       append([]string(nil), search.Args()...),
		})
	}, searchPinsOnly, searchHistoryOnly, searchExcludePins, searchFormat)

	picker := app.Command("picker", "Open the interactive fzf picker.")
	pickerFormat := picker.String("format", "fzf", "Picker output format.").Value()
	pickerPinLimit := picker.Int("pin-limit", 0, "Maximum number of pinned commands to load; zero means unlimited.").Value()
	picker.Run(func(_ context.Context, format string, pinLimit int) error {
		return runner.RunPicker(*r, currentConfig(), runner.PickerOptions{
			Format:   format,
			PinLimit: pinLimit,
		})
	}, pickerFormat, pickerPinLimit)

	pin := app.Command("pin", "Pin a command.")
	pin.Note("Positional arguments:\n  COMMAND...  Command text to pin.")
	pinDescription := pin.String("description", "", "Optional description for the pinned command.").Short("d").Value()
	pin.Run(func(_ context.Context, description string) error {
		return runner.RunPin(*r, currentConfig(), runner.PinOptions{
			Description: description,
			Command:     append([]string(nil), pin.Args()...),
		})
	}, pinDescription)

	describe := app.Command("describe", "Update the description of a pinned command.")
	describe.Note("Positional arguments:\n  COMMAND...  Pinned command text to describe.")
	describeDescription := describe.String("description", "", "Description for the pinned command.").Short("d").Value()
	describe.Run(func(_ context.Context, description string) error {
		return runner.RunDescribe(*r, currentConfig(), runner.DescribeOptions{
			Description: description,
			Command:     append([]string(nil), describe.Args()...),
		})
	}, describeDescription)

	remove := app.Command("remove", "Remove a pinned command.")
	remove.Note("Positional arguments:\n  COMMAND...  Pinned command text to remove.")
	remove.Run(func(_ context.Context) error {
		return runner.RunRemove(*r, currentConfig(), runner.RemoveOptions{
			Command: append([]string(nil), remove.Args()...),
		})
	})

	list := app.Command("list", "List pinned commands.")
	list.Run(func(_ context.Context) error {
		return runner.RunList(*r, currentConfig())
	})

	initCmd := app.Command("init", "Print shell integration for bash or zsh.")
	initCmd.Note("Positional arguments:\n  SHELL  Shell to initialize (bash or zsh).")
	initNoBind := initCmd.Bool("no-bind", false, "Print shell integration without binding Ctrl+R.").Value()
	initCmd.RequirePositional(1)
	initCmd.Run(func(_ context.Context, noBind bool) error {
		args := append([]string(nil), initCmd.Args()...)
		shell := ""
		if len(args) > 0 {
			shell = args[0]
		}

		return runner.RunInit(*r, runner.InitOptions{
			NoBind: noBind,
			Shell:  shell,
		})
	}, initNoBind)

	commandRunner, err := app.ParseRunner(args)
	if err != nil {
		return ParseResult{}, err
	}

	return ParseResult{
		Runner:  commandRunner,
		Verbose: *verbose,
		Debug:   *debug,
	}, nil
}

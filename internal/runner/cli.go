package runner

import (
	"github.com/alecthomas/kong"

	"github.com/containeroo/recall/internal/store"
)

// CLI describes recall's command-line interface.
type CLI struct {
	// DataDir is the directory where recall stores its own data.
	DataDir string `name:"data-dir" help:"Directory used to store recall data."`

	// HistoryFiles contains optional custom history files to scan.
	HistoryFiles []string `name:"history" help:"History file to scan; may be provided multiple times."`

	// HistoryLimit is the maximum number of recent history entries to read, with zero meaning unlimited.
	HistoryLimit int `name:"history-limit" default:"20000" help:"Maximum number of recent history entries to read; zero means unlimited."`

	// Limit is the maximum number of search results to print, with zero meaning unlimited.
	Limit int `short:"n" default:"20" help:"Maximum number of search results to print; zero means unlimited."`

	// Verbose enables informational diagnostic output.
	Verbose bool `short:"v" help:"Print informational diagnostics."`

	// Debug enables debug diagnostic output.
	Debug bool `help:"Print debug diagnostics."`

	// Version prints the recall version and exits.
	Version kong.VersionFlag `name:"version" help:"Print recall version and exit."`

	// Search scans pinned commands and shell history.
	Search SearchCmd `cmd:"" help:"Search pinned commands and shell history."`

	// Picker opens an interactive fzf picker for pins and shell-provided history.
	Picker PickerCmd `cmd:"" help:"Open the interactive fzf picker."`

	// Pin stores a command in recall's own pinned-command store.
	Pin PinCmd `cmd:"" help:"Pin a command."`

	// Describe updates the description of a pinned command.
	Describe DescribeCmd `cmd:"" help:"Update the description of a pinned command."`

	// Remove deletes a pinned command.
	Remove RemoveCmd `cmd:"" help:"Remove a pinned command."`

	// List prints pinned commands.
	List ListCmd `cmd:"" help:"List pinned commands."`

	// Init prints shell integration.
	Init InitCmd `cmd:"" help:"Print shell integration for bash or zsh."`
}

// Config contains resolved runtime configuration shared by commands.
type Config struct {
	// DataDir is the directory where recall stores its own data.
	DataDir string

	// HistoryFiles contains optional custom history files to scan.
	HistoryFiles []string

	// HistoryLimit is the maximum number of recent history entries to read, with zero meaning unlimited.
	HistoryLimit int

	// Limit is the maximum number of search results to print, with zero meaning unlimited.
	Limit int
}

// Config resolves global CLI options into runtime configuration.
func (c *CLI) Config(home string) Config {
	cfg := Config{
		DataDir:      c.DataDir,
		HistoryFiles: c.HistoryFiles,
		HistoryLimit: c.HistoryLimit,
		Limit:        c.Limit,
	}

	if cfg.DataDir == "" {
		cfg.DataDir = store.DefaultDir(home)
	}

	if cfg.HistoryLimit < 0 {
		cfg.HistoryLimit = 20000
	}

	if cfg.Limit < 0 {
		cfg.Limit = 20
	}

	return cfg
}

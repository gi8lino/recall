package runner

import (
	"io"
	"os"
	"path/filepath"

	"github.com/containeroo/recall/internal/history"
	"github.com/containeroo/recall/internal/logger"
)

// Runner holds process dependencies used by the application layer.
type Runner struct {
	input  io.Reader     // input reads command input for interactive commands.
	output logger.Logger // output writes normal command output.
	logger logger.Logger // logger writes warnings, informational messages, and debug messages.
	home   string        // home is the current user's home directory.
	binary string        // binary is the executable path used in generated shell snippets.
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

// SetLogging updates verbose/debug logger behavior for future command execution.
func (r *Runner) SetLogging(verbose bool, debug bool) {
	r.logger = r.logger.
		WithVerbose(verbose).
		WithDebug(debug)
}

// historySources returns either custom history files or default bash/zsh sources.
func (r Runner) historySources(cfg Config) []history.Source {
	if len(cfg.HistoryFiles) == 0 {
		return history.DefaultSources(r.home)
	}

	sources := make([]history.Source, 0, len(cfg.HistoryFiles))
	for _, file := range cfg.HistoryFiles {
		sources = append(sources, history.Source{
			Path:  expandHome(file, r.home),
			Shell: "custom",
		})
	}

	return sources
}

// storePath returns the JSON store path inside the configured data directory.
func (r Runner) storePath(cfg Config) string {
	return filepath.Join(cfg.DataDir, "store.json")
}

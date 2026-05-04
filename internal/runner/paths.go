package runner

import (
	"path/filepath"
	"strings"

	"github.com/containeroo/recall/internal/history"
)

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

// expandHome replaces a leading ~/ with the current user's home directory.
func expandHome(path string, home string) string {
	if path == "~" {
		return home
	}

	if after, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(home, after)
	}

	return path
}

package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerHistorySources(t *testing.T) {
	t.Run("returns default sources when no custom history files are configured", func(t *testing.T) {
		home := t.TempDir()
		histfile := filepath.Join(home, ".cache", "zsh", "history")

		t.Setenv("HISTFILE", histfile)

		r, _, _ := newTestRunner(t, home)
		sources := r.historySources(Config{})

		require.Len(t, sources, 3)
		assert.Equal(t, history.Source{Path: histfile, Shell: "shell"}, sources[0])
		assert.Equal(t, history.Source{Path: filepath.Join(home, ".zsh_history"), Shell: "zsh"}, sources[1])
		assert.Equal(t, history.Source{Path: filepath.Join(home, ".bash_history"), Shell: "bash"}, sources[2])
	})

	t.Run("returns custom expanded history sources", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		sources := r.historySources(Config{
			HistoryFiles: []string{
				"~/.cache/zsh/history",
				"/tmp/bash_history",
			},
		})

		require.Len(t, sources, 2)
		assert.Equal(t, history.Source{Path: filepath.Join(home, ".cache", "zsh", "history"), Shell: "custom"}, sources[0])
		assert.Equal(t, history.Source{Path: "/tmp/bash_history", Shell: "custom"}, sources[1])
	})
}

func TestRunnerStorePath(t *testing.T) {
	t.Parallel()

	t.Run("returns store path inside data directory", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, _, _ := newTestRunner(t, home)

		path := r.storePath(Config{DataDir: dataDir})

		assert.Equal(t, filepath.Join(dataDir, "store.json"), path)
	})
}

func TestExpandHome(t *testing.T) {
	t.Parallel()

	t.Run("expands exact tilde", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "/home/user", expandHome("~", "/home/user"))
	})

	t.Run("expands leading tilde slash", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, filepath.Join("/home/user", ".zsh_history"), expandHome("~/.zsh_history", "/home/user"))
	})

	t.Run("leaves absolute path unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "/tmp/history", expandHome("/tmp/history", "/home/user"))
	})

	t.Run("leaves embedded tilde unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "prefix/~/history", expandHome("prefix/~/history", "/home/user"))
	})
}

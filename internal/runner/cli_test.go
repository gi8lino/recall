package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIConfig(t *testing.T) {
	t.Run("uses default data dir and keeps zero limits", func(t *testing.T) {
		home := t.TempDir()

		t.Setenv("XDG_DATA_HOME", "")

		cfg := (&CLI{}).Config(home)

		assert.Equal(t, store.DefaultDir(home), cfg.DataDir)
		assert.Empty(t, cfg.HistoryFiles)
		assert.Equal(t, 0, cfg.HistoryLimit)
		assert.Equal(t, 0, cfg.Limit)
	})

	t.Run("uses kong populated default limit values", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()

		cfg := (&CLI{
			HistoryLimit: 20000,
			Limit:        20,
		}).Config(home)

		assert.Equal(t, store.DefaultDir(home), cfg.DataDir)
		assert.Empty(t, cfg.HistoryFiles)
		assert.Equal(t, 20000, cfg.HistoryLimit)
		assert.Equal(t, 20, cfg.Limit)
	})

	t.Run("uses configured values", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := filepath.Join(home, "history")

		cfg := (&CLI{
			DataDir:      dataDir,
			HistoryFiles: []string{historyFile},
			HistoryLimit: 42,
			Limit:        7,
		}).Config(home)

		assert.Equal(t, dataDir, cfg.DataDir)
		require.Equal(t, []string{historyFile}, cfg.HistoryFiles)
		assert.Equal(t, 42, cfg.HistoryLimit)
		assert.Equal(t, 7, cfg.Limit)
	})

	t.Run("normalizes negative limits", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")

		cfg := (&CLI{
			DataDir:      dataDir,
			HistoryLimit: -1,
			Limit:        -1,
		}).Config(home)

		assert.Equal(t, dataDir, cfg.DataDir)
		assert.Equal(t, 20000, cfg.HistoryLimit)
		assert.Equal(t, 20, cfg.Limit)
	})
}

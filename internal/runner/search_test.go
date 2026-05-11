package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/command"
	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSearch(t *testing.T) {
	t.Parallel()

	t.Run("prints pins only plain output", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		err := RunSearch(r, newTestConfig(dataDir), SearchOptions{
			PinsOnly: true,
			Query:    []string{"pods"},
		})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())
	})

	t.Run("prints fzf output", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		err := RunSearch(r, newTestConfig(dataDir), SearchOptions{
			PinsOnly: true,
			Format:   "fzf",
		})

		require.NoError(t, err)
		assert.Equal(
			t,
			"pin\tkubectl get pods\tShow running pods\tcommand\ta3ViZWN0bCBnZXQgcG9kcw==\n",
			stdout.String(),
		)
	})

	t.Run("prints pins before history", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := writeRunnerFile(t, home, "history", "kubectl get pods\ndocker compose up\n")
		r, stdout, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		cfg := newTestConfig(dataDir)
		cfg.HistoryFiles = []string{historyFile}

		err := RunSearch(r, cfg, SearchOptions{})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\ndocker compose up\n", stdout.String())
	})

	t.Run("excludes pinned history commands when requested", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := writeRunnerFile(t, home, "history", "kubectl get pods\ndocker compose up\n")
		r, stdout, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		cfg := newTestConfig(dataDir)
		cfg.HistoryFiles = []string{historyFile}

		err := RunSearch(r, cfg, SearchOptions{
			ExcludePins: true,
		})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\ndocker compose up\n", stdout.String())
	})
}

func TestRunnerSearchCommands(t *testing.T) {
	t.Parallel()

	t.Run("loads pinned commands", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, _, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		items, err := r.searchCommands(newTestConfig(dataDir), true, false, false)

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "kubectl get pods", items[0].Text)
		assert.True(t, items[0].Pinned)
	})

	t.Run("loads history commands", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		historyFile := writeRunnerFile(t, home, "history", "echo one\necho two\n")
		r, _, _ := newTestRunner(t, home)

		cfg := newTestConfig(filepath.Join(home, "data"))
		cfg.HistoryFiles = []string{historyFile}

		items, err := r.searchCommands(cfg, false, true, false)

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "echo two", items[0].Text)
		assert.Equal(t, "echo one", items[1].Text)
	})

	t.Run("loads default commands with exclude pins enabled", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := writeRunnerFile(t, home, "history", "kubectl get pods\ndocker compose up\n")
		r, _, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		cfg := newTestConfig(dataDir)
		cfg.HistoryFiles = []string{historyFile}

		items, err := r.searchCommands(cfg, false, false, true)

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "kubectl get pods", items[0].Text)
		assert.True(t, items[0].Pinned)
		assert.Equal(t, "docker compose up", items[1].Text)
		assert.False(t, items[1].Pinned)
	})
}

func TestRunnerLoadPinnedCommands(t *testing.T) {
	t.Parallel()

	t.Run("loads pins from store", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, _, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("go test ./...", "Run tests"))

		items, err := r.loadPinnedCommands(newTestConfig(dataDir))

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "go test ./...", items[0].Text)
		assert.Equal(t, "Run tests", items[0].Description)
	})
}

func TestRunnerLoadHistoryCommands(t *testing.T) {
	t.Parallel()

	t.Run("loads history commands", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := writeRunnerFile(t, home, "history", "echo one\necho two\n")
		r, _, _ := newTestRunner(t, home)

		cfg := newTestConfig(dataDir)
		cfg.HistoryFiles = []string{historyFile}

		items, err := r.loadHistoryCommands(cfg, false)

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "echo two", items[0].Text)
		assert.Equal(t, "echo one", items[1].Text)
	})

	t.Run("excludes pinned history commands", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := writeRunnerFile(t, home, "history", "kubectl get pods\ndocker compose up\n")
		r, _, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		cfg := newTestConfig(dataDir)
		cfg.HistoryFiles = []string{historyFile}

		items, err := r.loadHistoryCommands(cfg, true)

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "docker compose up", items[0].Text)
	})
}

func TestRunnerLoadCommands(t *testing.T) {
	t.Parallel()

	t.Run("loads pins before history and excludes duplicate history pins", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := writeRunnerFile(t, home, "history", "kubectl get pods\ndocker compose up\n")
		r, _, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		cfg := newTestConfig(dataDir)
		cfg.HistoryFiles = []string{historyFile}

		items, err := r.loadCommands(cfg, true)

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "kubectl get pods", items[0].Text)
		assert.True(t, items[0].Pinned)
		assert.Equal(t, "docker compose up", items[1].Text)
		assert.False(t, items[1].Pinned)
	})
}

func TestExcludePinnedCommands(t *testing.T) {
	t.Parallel()

	t.Run("removes history commands that are pinned", func(t *testing.T) {
		t.Parallel()

		historyCommands := []command.Command{
			command.NewCommand("kubectl get pods", "history", false),
			command.NewCommand("docker compose up", "history", false),
		}

		pins := []command.Command{
			command.NewCommand("KUBECTL GET PODS", "pin", true),
		}

		filtered := excludePinnedCommands(historyCommands, pins)

		require.Len(t, filtered, 1)
		assert.Equal(t, "docker compose up", filtered[0].Text)
	})
}

func TestCommandKey(t *testing.T) {
	t.Parallel()

	t.Run("normalizes command key", func(t *testing.T) {
		t.Parallel()

		key := commandKey("  KUBECTL GET PODS  ")

		assert.Equal(t, "kubectl get pods", key)
	})
}

func TestRunnerPrintPlainResults(t *testing.T) {
	t.Parallel()

	t.Run("prints one command per line", func(t *testing.T) {
		t.Parallel()

		r, stdout, _ := newTestRunner(t, t.TempDir())

		r.printPlainResults([]command.Command{
			command.NewCommand("kubectl get pods", "pin", true),
			command.NewCommand("go test ./...", "history", false),
		})

		assert.Equal(t, "kubectl get pods\ngo test ./...\n", stdout.String())
	})
}

func TestRunnerPrintFZFResults(t *testing.T) {
	t.Parallel()

	t.Run("prints sanitized tab separated rows with encoded payload", func(t *testing.T) {
		t.Parallel()

		r, stdout, _ := newTestRunner(t, t.TempDir())

		r.printFZFResults([]command.Command{
			command.NewCommandWithDescription("kubectl\tget\r\npods", "Show\tpods", "pin", true),
		})

		assert.Equal(
			t,
			"pin\tkubectl get\\r\\npods\tShow pods\tcommand\ta3ViZWN0bAlnZXQNCnBvZHM=\n",
			stdout.String(),
		)
	})
}

func TestResultKind(t *testing.T) {
	t.Parallel()

	t.Run("returns pin for pinned command", func(t *testing.T) {
		t.Parallel()

		kind := resultKind(command.NewCommand("kubectl get pods", "pin", true))

		assert.Equal(t, "pin", kind)
	})

	t.Run("returns history for non pinned command", func(t *testing.T) {
		t.Parallel()

		kind := resultKind(command.NewCommand("kubectl get pods", "history", false))

		assert.Equal(t, "history", kind)
	})
}

func TestDisplayFZFField(t *testing.T) {
	t.Parallel()

	t.Run("replaces separators and trims whitespace", func(t *testing.T) {
		t.Parallel()

		value := displayFZFField("  kubectl\tget\r\npods  ")

		assert.Equal(t, "kubectl get\\r\\npods", value)
	})
}

func TestEncodeFZFPayload(t *testing.T) {
	t.Parallel()

	t.Run("encodes raw command text", func(t *testing.T) {
		t.Parallel()

		payload := encodeFZFPayload("kubectl get pods")

		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", payload)
	})
}

func writeRunnerFile(t testing.TB, dir string, name string, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	return path
}

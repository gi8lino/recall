package runner

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunner(t *testing.T) {
	t.Parallel()

	t.Run("creates runner with process dependencies", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, stdout, stderr := newTestRunner(t, home)

		r.output.Line("normal")
		r.logger.Warnf("warning")

		assert.Equal(t, "normal\n", stdout.String())
		assert.Equal(t, "warning: warning\n", stderr.String())
		assert.Equal(t, home, r.home)
		assert.Equal(t, filepath.Join(home, "bin", "recall"), r.binary)
	})
}

func TestRunnerRun(t *testing.T) {
	t.Parallel()

	t.Run("parses and runs command", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, stderr := newTestRunner(t, home)

		err := r.Run([]string{
			"--data-dir", dataDir,
			"pin",
			"--description", "Show running pods",
			"--",
			"kubectl", "get", "pods",
		})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())
		assert.Empty(t, stderr.String())
	})
}

func TestRunnerParseError(t *testing.T) {
	t.Parallel()

	t.Run("returns original non kong parse error", func(t *testing.T) {
		t.Parallel()

		r, _, _ := newTestRunner(t, t.TempDir())
		err := errors.New("boom")

		got := r.parseError(err)

		assert.Same(t, err, got)
	})
}

func newTestRunner(t testing.TB, home string) (Runner, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	r := NewRunner(
		&stdout,
		&stderr,
		home,
		filepath.Join(home, "bin", "recall"),
	)

	return r, &stdout, &stderr
}

func newTestCLI(dataDir string) *CLI {
	return &CLI{
		DataDir:      dataDir,
		HistoryLimit: 0,
		Limit:        0,
	}
}

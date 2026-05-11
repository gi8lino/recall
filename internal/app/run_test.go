package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Run("parses and runs command", func(t *testing.T) {
		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		binDir := filepath.Join(home, "bin")
		require.NoError(t, os.MkdirAll(binDir, 0o755))

		oldHome := os.Getenv("HOME")
		t.Setenv("HOME", home)

		execPath := filepath.Join(binDir, "recall")
		require.NoError(t, os.WriteFile(execPath, []byte("#!/bin/sh\n"), 0o755))

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		err := Run("dev", &stdout, &stderr, []string{
			"--data-dir", dataDir,
			"pin",
			"--description", "Show running pods",
			"--",
			"kubectl", "get", "pods",
		})

		_ = oldHome
		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())
		assert.Empty(t, stderr.String())
	})

	t.Run("rejects bare args without command", func(t *testing.T) {
		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		historyFile := filepath.Join(home, ".bash_history")
		require.NoError(t, os.WriteFile(historyFile, []byte("docker compose up\n"), 0o644))
		t.Setenv("HOME", home)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		err := Run("dev", &stdout, &stderr, []string{
			"--data-dir", dataDir,
			"--history", historyFile,
			"docker",
		})

		require.Error(t, err)
		assert.Empty(t, stdout.String())
		assert.Empty(t, stderr.String())
		assert.Contains(t, err.Error(), "Usage:")
		assert.Contains(t, err.Error(), "search")
		assert.Contains(t, err.Error(), "pin")
	})
}

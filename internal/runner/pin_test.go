package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPin(t *testing.T) {
	t.Parallel()

	t.Run("requires command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunPin(r, newTestConfig(filepath.Join(home, "data")), PinOptions{})

		require.Error(t, err)
		assert.ErrorContains(t, err, "pin requires command text")
	})

	t.Run("requires non empty command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunPin(r, newTestConfig(filepath.Join(home, "data")), PinOptions{Command: []string{"--"}})

		require.Error(t, err)
		assert.ErrorContains(t, err, "pin requires non-empty command text")
	})

	t.Run("stores command with description", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		err := RunPin(r, newTestConfig(dataDir), PinOptions{
			Description: "Show running pods",
			Command:     []string{"--", "kubectl", "get", "pods"},
		})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())

		pins, err := store.NewStore(filepath.Join(dataDir, "store.json")).Pins()
		require.NoError(t, err)
		require.Len(t, pins, 1)
		assert.Equal(t, "kubectl get pods", pins[0].Text)
		assert.Equal(t, "Show running pods", pins[0].Description)
	})
}

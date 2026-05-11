package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRemove(t *testing.T) {
	t.Parallel()

	t.Run("requires command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunRemove(r, newTestConfig(filepath.Join(home, "data")), RemoveOptions{})

		require.Error(t, err)
		assert.ErrorContains(t, err, "remove requires command text")
	})

	t.Run("requires non empty command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunRemove(r, newTestConfig(filepath.Join(home, "data")), RemoveOptions{Command: []string{"--"}})

		require.Error(t, err)
		assert.ErrorContains(t, err, "remove requires non-empty command text")
	})

	t.Run("removes existing pin", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		err := RunRemove(r, newTestConfig(dataDir), RemoveOptions{
			Command: []string{"--", "kubectl", "get", "pods"},
		})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())

		pins, err := st.Pins()
		require.NoError(t, err)
		assert.Empty(t, pins)
	})
}

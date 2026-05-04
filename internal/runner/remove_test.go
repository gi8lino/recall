package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveCmdRun(t *testing.T) {
	t.Parallel()

	t.Run("requires command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := (&RemoveCmd{}).Run(r, newTestCLI(filepath.Join(home, "data")))

		require.Error(t, err)
		assert.ErrorContains(t, err, "remove requires command text")
	})

	t.Run("requires non empty command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := (&RemoveCmd{Command: []string{"--"}}).Run(r, newTestCLI(filepath.Join(home, "data")))

		require.Error(t, err)
		assert.ErrorContains(t, err, "remove requires non-empty command text")
	})

	t.Run("removes existing pin", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		err := (&RemoveCmd{
			Command: []string{"--", "kubectl", "get", "pods"},
		}).Run(r, newTestCLI(dataDir))

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())

		pins, err := st.Pins()
		require.NoError(t, err)
		assert.Empty(t, pins)
	})
}

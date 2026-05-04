package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmdRun(t *testing.T) {
	t.Parallel()

	t.Run("prints no output when there are no pins", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		err := (&ListCmd{}).Run(r, newTestCLI(dataDir))

		require.NoError(t, err)
		assert.Empty(t, stdout.String())
	})

	t.Run("prints pins in priority order", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("go test ./...", ""))
		require.NoError(t, st.AddPin("kubectl get pods", "Show running pods"))

		err := (&ListCmd{}).Run(r, newTestCLI(dataDir))

		require.NoError(t, err)
		assert.Equal(t, "Show running pods\tkubectl get pods\ngo test ./...\n", stdout.String())
	})
}

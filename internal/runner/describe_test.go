package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeCmdRun(t *testing.T) {
	t.Parallel()

	t.Run("requires description", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := (&DescribeCmd{
			Command: []string{"kubectl", "get", "pods"},
		}).Run(r, newTestCLI(filepath.Join(home, "data")))

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires a non-empty description")
	})

	t.Run("requires command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := (&DescribeCmd{
			Description: "Show running pods",
		}).Run(r, newTestCLI(filepath.Join(home, "data")))

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires command text")
	})

	t.Run("requires non empty command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := (&DescribeCmd{
			Description: "Show running pods",
			Command:     []string{"--"},
		}).Run(r, newTestCLI(filepath.Join(home, "data")))

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires non-empty command text")
	})

	t.Run("updates existing pin description", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Old description"))

		err := (&DescribeCmd{
			Description: "Show running pods",
			Command:     []string{"--", "kubectl", "get", "pods"},
		}).Run(r, newTestCLI(dataDir))

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())

		pins, err := st.Pins()
		require.NoError(t, err)
		require.Len(t, pins, 1)
		assert.Equal(t, "Show running pods", pins[0].Description)
	})
}

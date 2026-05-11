package runner

import (
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDescribe(t *testing.T) {
	t.Parallel()

	t.Run("requires description", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunDescribe(r, newTestConfig(filepath.Join(home, "data")), DescribeOptions{
			Command: []string{"kubectl", "get", "pods"},
		})

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires a non-empty description")
	})

	t.Run("requires command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunDescribe(r, newTestConfig(filepath.Join(home, "data")), DescribeOptions{
			Description: "Show running pods",
		})

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires command text")
	})

	t.Run("requires non empty command text", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		r, _, _ := newTestRunner(t, home)

		err := RunDescribe(r, newTestConfig(filepath.Join(home, "data")), DescribeOptions{
			Description: "Show running pods",
			Command:     []string{"--"},
		})

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires non-empty command text")
	})

	t.Run("updates existing pin description", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)

		st := store.NewStore(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Old description"))

		err := RunDescribe(r, newTestConfig(dataDir), DescribeOptions{
			Description: "Show running pods",
			Command:     []string{"--", "kubectl", "get", "pods"},
		})

		require.NoError(t, err)
		assert.Equal(t, "kubectl get pods\n", stdout.String())

		pins, err := st.Pins()
		require.NoError(t, err)
		require.Len(t, pins, 1)
		assert.Equal(t, "Show running pods", pins[0].Description)
	})
}

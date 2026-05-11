package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandTextFromArgs(t *testing.T) {
	t.Parallel()

	t.Run("joins arguments", func(t *testing.T) {
		t.Parallel()

		text := commandTextFromArgs([]string{"kubectl", "get", "pods"})

		assert.Equal(t, "kubectl get pods", text)
	})

	t.Run("strips leading passthrough separator", func(t *testing.T) {
		t.Parallel()

		text := commandTextFromArgs([]string{"--", "kubectl", "get", "pods"})

		assert.Equal(t, "kubectl get pods", text)
	})

	t.Run("keeps separator inside command", func(t *testing.T) {
		t.Parallel()

		text := commandTextFromArgs([]string{"ssh", "host", "--", "ls", "-lah"})

		assert.Equal(t, "ssh host -- ls -lah", text)
	})

	t.Run("returns empty string for empty args", func(t *testing.T) {
		t.Parallel()

		text := commandTextFromArgs(nil)

		assert.Empty(t, text)
	})

	t.Run("returns empty string for separator only", func(t *testing.T) {
		t.Parallel()

		text := commandTextFromArgs([]string{"--"})

		assert.Empty(t, text)
	})
}

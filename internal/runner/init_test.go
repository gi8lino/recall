package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInit(t *testing.T) {
	t.Parallel()

	t.Run("prints zsh integration", func(t *testing.T) {
		t.Parallel()

		r, stdout, _ := newTestRunner(t, t.TempDir())

		err := RunInit(r, InitOptions{Shell: "zsh"})

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "__recall_ctrl_r")
		assert.Contains(t, stdout.String(), "bindkey '^R' __recall_ctrl_r")
	})

	t.Run("prints zsh integration without binding", func(t *testing.T) {
		t.Parallel()

		r, stdout, _ := newTestRunner(t, t.TempDir())

		err := RunInit(r, InitOptions{Shell: "zsh", NoBind: true})

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "zle -N __recall_ctrl_r")
		assert.Contains(t, stdout.String(), "# bindkey '^R' __recall_ctrl_r")
	})
}

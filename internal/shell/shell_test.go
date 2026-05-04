package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("prints zsh integration", func(t *testing.T) {
		t.Parallel()

		script, err := Init(InitRequest{
			Shell:  "zsh",
			Binary: "/tmp/recall",
		})

		require.NoError(t, err)
		assert.Contains(t, script, "__recall_history_for_fzf")
		assert.Contains(t, script, "zsh-history")
		assert.Contains(t, script, "RECALL_FZF_RETURN=record")
		assert.Contains(t, script, "recall-fzf")
		assert.Contains(t, script, "bindkey '^R' __recall_ctrl_r")
		assert.Contains(t, script, "RECALL_BIN='/tmp/recall'")
	})

	t.Run("prints zsh integration without binding", func(t *testing.T) {
		t.Parallel()

		script, err := Init(InitRequest{
			Shell:  "zsh",
			Binary: "recall",
			NoBind: true,
		})

		require.NoError(t, err)
		assert.Contains(t, script, "zle -N __recall_ctrl_r")
		assert.Contains(t, script, "# bindkey '^R' __recall_ctrl_r")
		assert.NotContains(t, script, "\nbindkey '^R' __recall_ctrl_r\n")
	})

	t.Run("prints bash integration", func(t *testing.T) {
		t.Parallel()

		script, err := Init(InitRequest{
			Shell:  "bash",
			Binary: "recall",
		})

		require.NoError(t, err)
		assert.Contains(t, script, "__recall_history_for_fzf")
		assert.Contains(t, script, "command\\t\\t")
		assert.Contains(t, script, "recall-fzf")
		assert.Contains(t, script, `bind -x '"\C-r": __recall_ctrl_r'`)
	})

	t.Run("prints bash integration without binding", func(t *testing.T) {
		t.Parallel()

		script, err := Init(InitRequest{
			Shell:  "bash",
			Binary: "recall",
			NoBind: true,
		})

		require.NoError(t, err)
		assert.Contains(t, script, "# bind -x")
		assert.NotContains(t, script, "\nbind -x")
	})

	t.Run("uses recall when binary is empty", func(t *testing.T) {
		t.Parallel()

		script, err := Init(InitRequest{Shell: "zsh"})

		require.NoError(t, err)
		assert.Contains(t, script, "RECALL_BIN='recall'")
	})

	t.Run("rejects unsupported shell", func(t *testing.T) {
		t.Parallel()

		script, err := Init(InitRequest{Shell: "fish"})

		require.Error(t, err)
		assert.Empty(t, script)
		assert.ErrorContains(t, err, `unsupported shell "fish"`)
	})
}

func TestShellQuote(t *testing.T) {
	t.Parallel()

	t.Run("quotes plain value", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "'recall'", shellQuote("recall"))
	})

	t.Run("quotes value with apostrophe", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, `'a'"'"'b'`, shellQuote("a'b"))
	})
}

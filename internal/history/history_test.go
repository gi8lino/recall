package history

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSources(t *testing.T) {
	t.Run("uses HISTFILE and common shell history files", func(t *testing.T) {
		home := t.TempDir()
		histfile := filepath.Join(home, ".cache", "zsh", "history")

		t.Setenv("HISTFILE", histfile)

		sources := DefaultSources(home)

		require.Len(t, sources, 3)
		assert.Equal(t, Source{Path: histfile, Shell: "shell"}, sources[0])
		assert.Equal(t, Source{Path: filepath.Join(home, ".zsh_history"), Shell: "zsh"}, sources[1])
		assert.Equal(t, Source{Path: filepath.Join(home, ".bash_history"), Shell: "bash"}, sources[2])
	})

	t.Run("expands leading tilde in HISTFILE", func(t *testing.T) {
		home := t.TempDir()

		t.Setenv("HISTFILE", "~/.cache/zsh/history")

		sources := DefaultSources(home)

		require.Len(t, sources, 3)
		assert.Equal(t, filepath.Join(home, ".cache", "zsh", "history"), sources[0].Path)
		assert.Equal(t, "shell", sources[0].Shell)
	})

	t.Run("removes empty HISTFILE source", func(t *testing.T) {
		home := t.TempDir()

		t.Setenv("HISTFILE", "")

		sources := DefaultSources(home)

		require.Len(t, sources, 2)
		assert.Equal(t, Source{Path: filepath.Join(home, ".zsh_history"), Shell: "zsh"}, sources[0])
		assert.Equal(t, Source{Path: filepath.Join(home, ".bash_history"), Shell: "bash"}, sources[1])
	})

	t.Run("removes duplicate HISTFILE source", func(t *testing.T) {
		home := t.TempDir()
		zshHistory := filepath.Join(home, ".zsh_history")

		t.Setenv("HISTFILE", zshHistory)

		sources := DefaultSources(home)

		require.Len(t, sources, 2)
		assert.Equal(t, Source{Path: zshHistory, Shell: "shell"}, sources[0])
		assert.Equal(t, Source{Path: filepath.Join(home, ".bash_history"), Shell: "bash"}, sources[1])
	})
}

func TestNewReader(t *testing.T) {
	t.Parallel()

	t.Run("stores sources and max lines", func(t *testing.T) {
		t.Parallel()

		sources := []Source{
			{Path: "/tmp/zsh_history", Shell: "zsh"},
			{Path: "/tmp/bash_history", Shell: "bash"},
		}

		reader := NewReader(sources, 42)

		assert.Equal(t, sources, reader.Sources)
		assert.Equal(t, 42, reader.MaxLines)
	})
}

func TestReaderRead(t *testing.T) {
	t.Parallel()

	t.Run("reads all configured sources in source order", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		zshPath := writeHistoryFile(t, dir, "zsh_history", ": 1:0;echo zsh one\n: 2:0;echo zsh two\n")
		bashPath := writeHistoryFile(t, dir, "bash_history", "echo bash one\necho bash two\n")

		reader := NewReader([]Source{
			{Path: zshPath, Shell: "zsh"},
			{Path: bashPath, Shell: "bash"},
		}, 0)

		commands, err := reader.Read()

		require.NoError(t, err)
		require.Len(t, commands, 4)

		assert.Equal(t, "echo zsh two", commands[0].Text)
		assert.Equal(t, "zsh", commands[0].Source)

		assert.Equal(t, "echo zsh one", commands[1].Text)
		assert.Equal(t, "zsh", commands[1].Source)

		assert.Equal(t, "echo bash two", commands[2].Text)
		assert.Equal(t, "bash", commands[2].Source)

		assert.Equal(t, "echo bash one", commands[3].Text)
		assert.Equal(t, "bash", commands[3].Source)
	})

	t.Run("ignores missing source files", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		existingPath := writeHistoryFile(t, dir, "history", "echo one\n")
		missingPath := filepath.Join(dir, "missing")

		reader := NewReader([]Source{
			{Path: missingPath, Shell: "missing"},
			{Path: existingPath, Shell: "custom"},
		}, 0)

		commands, err := reader.Read()

		require.NoError(t, err)
		require.Len(t, commands, 1)
		assert.Equal(t, "echo one", commands[0].Text)
		assert.Equal(t, "custom", commands[0].Source)
	})

	t.Run("returns readable commands and joined source errors", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		existingPath := writeHistoryFile(t, dir, "history", "echo one\n")
		badPath := dir

		reader := NewReader([]Source{
			{Path: badPath, Shell: "bad"},
			{Path: existingPath, Shell: "custom"},
		}, 0)

		commands, err := reader.Read()

		require.Error(t, err)
		require.Len(t, commands, 1)
		assert.Equal(t, "echo one", commands[0].Text)
	})
}

func TestReadSource(t *testing.T) {
	t.Parallel()

	t.Run("returns commands in reverse file order", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := writeHistoryFile(t, dir, "history", "echo one\necho two\necho three\n")

		commands, err := readSource(Source{Path: path, Shell: "custom"}, 0)

		require.NoError(t, err)
		require.Len(t, commands, 3)

		assert.Equal(t, "echo three", commands[0].Text)
		assert.Equal(t, "custom", commands[0].Source)
		assert.False(t, commands[0].Pinned)

		assert.Equal(t, "echo two", commands[1].Text)
		assert.Equal(t, "echo one", commands[2].Text)
	})

	t.Run("returns no commands for missing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "missing")

		commands, err := readSource(Source{Path: path, Shell: "custom"}, 0)

		require.NoError(t, err)
		assert.Empty(t, commands)
	})

	t.Run("applies max lines before reversing commands", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := writeHistoryFile(t, dir, "history", "echo one\necho two\necho three\n")

		commands, err := readSource(Source{Path: path, Shell: "custom"}, 2)

		require.NoError(t, err)
		require.Len(t, commands, 2)
		assert.Equal(t, "echo three", commands[0].Text)
		assert.Equal(t, "echo two", commands[1].Text)
	})

	t.Run("preserves multiline zsh commands as one entry", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := writeHistoryFile(t, dir, "zsh_history", ": 1700000000:0;git commit \\\n  -m 'message'\n: 1700000001:0;echo done\n")

		commands, err := readSource(Source{Path: path, Shell: "zsh"}, 3)

		require.NoError(t, err)
		require.Len(t, commands, 2)
		assert.Equal(t, "echo done", commands[0].Text)
		assert.Equal(t, "git commit \\\n  -m 'message'", commands[1].Text)
	})

	t.Run("detects zsh extended history through shell sources", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := writeHistoryFile(t, dir, "history", ": 1700000000:0;git commit \\\n  -m 'message'\n: 1700000001:0;echo done\n")

		commands, err := readSource(Source{Path: path, Shell: "shell"}, 3)

		require.NoError(t, err)
		require.Len(t, commands, 2)
		assert.Equal(t, "echo done", commands[0].Text)
		assert.Equal(t, "git commit \\\n  -m 'message'", commands[1].Text)
	})
}

func TestReadLines(t *testing.T) {
	t.Parallel()

	t.Run("reads all lines when max lines is zero", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "echo one\necho two\n")
		defer file.Close()

		lines, err := readLines(file, 0)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo one", "echo two"}, lines)
	})

	t.Run("reads recent lines when max lines is positive", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "echo one\necho two\necho three\n")
		defer file.Close() // nolint:errcheck

		lines, err := readLines(file, 2)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo two", "echo three"}, lines)
	})
}

func TestScanAllLines(t *testing.T) {
	t.Parallel()

	t.Run("parses all supported history lines", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "\n#123456\n: 1700000000:0;echo zsh\n  echo bash  \n")
		defer file.Close() // nolint:errcheck

		lines, err := scanAllLines(file)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo zsh", "echo bash"}, lines)
	})

	t.Run("supports long command lines", func(t *testing.T) {
		t.Parallel()

		longCommand := "echo " + string(make([]byte, 70*1024))
		file := openHistoryFile(t, longCommand+"\n")
		defer file.Close() // nolint:errcheck

		lines, err := scanAllLines(file)

		require.NoError(t, err)
		require.Len(t, lines, 1)
		assert.Len(t, lines[0], len(longCommand))
	})
}

func TestReadLastLines(t *testing.T) {
	t.Parallel()

	t.Run("returns recent lines in original order", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "echo one\necho two\necho three\necho four\n")
		defer file.Close()

		lines, err := readLastLines(file, 2)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo three", "echo four"}, lines)
	})

	t.Run("returns all lines when file has fewer lines than max", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "echo one\necho two\n")
		defer file.Close()

		lines, err := readLastLines(file, 5)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo one", "echo two"}, lines)
	})

	t.Run("returns empty lines for empty file", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "")
		defer file.Close()

		lines, err := readLastLines(file, 5)

		require.NoError(t, err)
		assert.Empty(t, lines)
	})

	t.Run("parses recent zsh and bash history lines", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "#123456\n: 1700000000:0;echo zsh\necho bash\n")
		defer file.Close()

		lines, err := readLastLines(file, 3)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo zsh", "echo bash"}, lines)
	})

	t.Run("handles file without trailing newline", func(t *testing.T) {
		t.Parallel()

		file := openHistoryFile(t, "echo one\necho two\necho three")
		defer file.Close()

		lines, err := readLastLines(file, 2)

		require.NoError(t, err)
		assert.Equal(t, []string{"echo two", "echo three"}, lines)
	})
}

func TestParseLine(t *testing.T) {
	t.Parallel()

	t.Run("returns empty string for blank line", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, parseLine("   "))
	})

	t.Run("skips bash timestamp line", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, parseLine("#1700000000"))
	})

	t.Run("extracts zsh extended history command", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "kubectl get pods", parseLine(": 1700000000:0;kubectl get pods"))
	})

	t.Run("trims plain command line", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "go test ./...", parseLine("  go test ./...  "))
	})
}

func TestParseZshExtendedLine(t *testing.T) {
	t.Parallel()

	t.Run("extracts command text", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "echo hello", parseZshExtendedLine(": 1700000000:0;echo hello"))
	})

	t.Run("trims extracted command text", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "echo hello", parseZshExtendedLine(": 1700000000:0;  echo hello  "))
	})

	t.Run("returns empty string for non zsh line", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, parseZshExtendedLine("echo hello"))
	})

	t.Run("returns empty string when semicolon is missing", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, parseZshExtendedLine(": 1700000000:0"))
	})

	t.Run("returns empty string when command is missing", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, parseZshExtendedLine(": 1700000000:0;"))
	})
}

func TestParseZshHistoryLines(t *testing.T) {
	t.Parallel()

	t.Run("preserves multiline commands and skips orphaned leading fragments", func(t *testing.T) {
		t.Parallel()

		lines := parseZshHistoryLines([]string{
			"  -m 'orphaned fragment'",
			": 1700000000:0;git commit \\",
			"  -m 'message'",
			": 1700000001:0;echo done",
		})

		assert.Equal(t, []string{"git commit \\\n  -m 'message'", "echo done"}, lines)
	})
}

func TestLooksLikeBashTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("returns true for numeric timestamp", func(t *testing.T) {
		t.Parallel()

		assert.True(t, looksLikeBashTimestamp("#1700000000"))
	})

	t.Run("returns false for empty hash", func(t *testing.T) {
		t.Parallel()

		assert.False(t, looksLikeBashTimestamp("#"))
	})

	t.Run("returns false for non hash line", func(t *testing.T) {
		t.Parallel()

		assert.False(t, looksLikeBashTimestamp("1700000000"))
	})

	t.Run("returns false for non numeric timestamp", func(t *testing.T) {
		t.Parallel()

		assert.False(t, looksLikeBashTimestamp("#1700000000abc"))
	})
}

func TestUniqueSources(t *testing.T) {
	t.Parallel()

	t.Run("removes empty paths", func(t *testing.T) {
		t.Parallel()

		sources := uniqueSources([]Source{
			{Path: "", Shell: "empty"},
			{Path: "/tmp/history", Shell: "custom"},
		})

		assert.Equal(t, []Source{
			{Path: "/tmp/history", Shell: "custom"},
		}, sources)
	})

	t.Run("removes duplicate paths and keeps first source", func(t *testing.T) {
		t.Parallel()

		sources := uniqueSources([]Source{
			{Path: "/tmp/history", Shell: "first"},
			{Path: "/tmp/history", Shell: "second"},
			{Path: "/tmp/other", Shell: "other"},
		})

		assert.Equal(t, []Source{
			{Path: "/tmp/history", Shell: "first"},
			{Path: "/tmp/other", Shell: "other"},
		}, sources)
	})
}

func TestExpandHome(t *testing.T) {
	t.Parallel()

	t.Run("expands exact tilde", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "/home/user", expandHome("~", "/home/user"))
	})

	t.Run("expands leading tilde slash", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, filepath.Join("/home/user", ".zsh_history"), expandHome("~/.zsh_history", "/home/user"))
	})

	t.Run("leaves absolute path unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "/tmp/history", expandHome("/tmp/history", "/home/user"))
	})

	t.Run("leaves embedded tilde unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "prefix/~/history", expandHome("prefix/~/history", "/home/user"))
	})
}

func writeHistoryFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	return path
}

func openHistoryFile(t *testing.T, content string) *os.File {
	t.Helper()

	path := writeHistoryFile(t, t.TempDir(), "history", content)

	file, err := os.Open(path)
	require.NoError(t, err)

	return file
}

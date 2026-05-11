package history

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/containeroo/recall/internal/command"
)

// Source describes one shell history file to scan.
type Source struct {
	Path  string // Path is the absolute path to the history file.
	Shell string // Shell identifies the shell format expected in the file.
}

// Reader reads commands from one or more shell history files.
type Reader struct {
	Sources  []Source // Sources contains the ordered list of history files to scan.
	MaxLines int      // MaxLines limits how many recent lines are read from each history file.
}

// DefaultSources returns the default shell history files.
func DefaultSources(home string) []Source {
	return uniqueSources([]Source{
		{
			Path:  expandHome(os.Getenv("HISTFILE"), home),
			Shell: "shell",
		},
		{
			Path:  filepath.Join(home, ".zsh_history"),
			Shell: "zsh",
		},
		{
			Path:  filepath.Join(home, ".bash_history"),
			Shell: "bash",
		},
	})
}

// NewReader creates a history reader for the given sources.
func NewReader(sources []Source, maxLines int) Reader {
	return Reader{
		Sources:  sources,
		MaxLines: maxLines,
	}
}

// Read returns commands from all configured history sources.
func (r Reader) Read() ([]command.Command, error) {
	var joined error

	commands := make([]command.Command, 0)
	for _, source := range r.Sources {
		items, err := readSource(source, r.MaxLines)
		if err != nil {
			joined = errors.Join(joined, err)
			continue
		}

		commands = append(commands, items...)
	}

	return commands, joined
}

// readSource reads commands from one shell history source.
func readSource(source Source, maxLines int) ([]command.Command, error) {
	file, err := os.Open(source.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint:errcheck

	rawLines, err := readRawLines(file, maxLines)
	if err != nil {
		return nil, err
	}

	lines := parseHistoryLines(rawLines)
	if source.Shell == "zsh" || (source.Shell == "shell" && containsZshExtendedLine(rawLines)) {
		lines = parseZshHistoryLines(rawLines)
	}

	commands := make([]command.Command, 0, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		commands = append(commands, command.NewCommand(lines[i], source.Shell, false))
	}

	return commands, nil
}

// readLines reads either all lines or the most recent lines from a file.
func readLines(file *os.File, maxLines int) ([]string, error) {
	if maxLines > 0 {
		return readLastLines(file, maxLines)
	}

	return scanAllLines(file)
}

// readRawLines reads raw lines from a history file without parsing shell syntax.
func readRawLines(file *os.File, maxLines int) ([]string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	if maxLines > 0 {
		return readLastRawLines(file, maxLines)
	}

	return scanAllRawLines(file)
}

// scanAllLines reads and parses every line from a history file.
func scanAllLines(file *os.File) ([]string, error) {
	rawLines, err := readRawLines(file, 0)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(rawLines))
	for _, rawLine := range rawLines {
		text := parseLine(rawLine)
		if text == "" {
			continue
		}

		lines = append(lines, text)
	}

	return lines, nil
}

// scanAllRawLines reads every raw line from a history file.
func scanAllRawLines(file *os.File) ([]string, error) {
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	rawLines := make([]string, 0)
	for scanner.Scan() {
		rawLines = append(rawLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rawLines, nil
}

// readLastLines reads and parses the most recent lines from a history file.
func readLastLines(file *os.File, maxLines int) ([]string, error) {
	rawLines, err := readLastRawLines(file, maxLines)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(rawLines))
	for _, rawLine := range rawLines {
		text := parseLine(rawLine)
		if text == "" {
			continue
		}

		lines = append(lines, text)
	}

	return lines, nil
}

// readLastRawLines reads the most recent raw lines from a history file.
func readLastRawLines(file *os.File, maxLines int) ([]string, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Size() == 0 {
		return nil, nil
	}

	const chunkSize int64 = 64 * 1024

	offset := stat.Size()
	data := make([]byte, 0)
	truncated := false

	for offset > 0 && bytes.Count(data, []byte{'\n'}) <= maxLines {
		readSize := min(chunkSize, offset)

		offset -= readSize
		chunk := make([]byte, readSize)

		if _, err := file.ReadAt(chunk, offset); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		data = append(chunk, data...)
		truncated = offset > 0
	}

	rawLines := strings.Split(string(data), "\n")
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	if truncated && len(rawLines) > 0 {
		rawLines = rawLines[1:]
	}

	if len(rawLines) > maxLines {
		rawLines = rawLines[len(rawLines)-maxLines:]
	}

	return rawLines, nil
}

// parseZshHistoryLines parses raw zsh extended-history lines into logical commands.
func parseZshHistoryLines(rawLines []string) []string {
	lines := make([]string, 0, len(rawLines))
	start := firstZshExtendedLine(rawLines)
	if start < 0 {
		return lines
	}

	rawLines = rawLines[start:]

	var current strings.Builder
	inEntry := false

	flushZshEntry := func() {
		if !inEntry {
			return
		}

		text := strings.TrimSpace(current.String())
		if text != "" {
			lines = append(lines, text)
		}

		current.Reset()
		inEntry = false
	}

	for _, rawLine := range rawLines {
		line := strings.TrimRight(rawLine, "\r")

		if zshCommand := parseZshExtendedLine(line); zshCommand != "" {
			flushZshEntry()
			current.WriteString(zshCommand)
			inEntry = true
			continue
		}

		if inEntry {
			current.WriteByte('\n')
			current.WriteString(line)
		}
	}

	flushZshEntry()

	return lines
}

// parseHistoryLines parses raw history file lines one line at a time.
func parseHistoryLines(rawLines []string) []string {
	lines := make([]string, 0, len(rawLines))
	for _, rawLine := range rawLines {
		text := parseLine(rawLine)
		if text == "" {
			continue
		}

		lines = append(lines, text)
	}

	return lines
}

// containsZshExtendedLine reports whether raw lines include zsh extended-history markers.
func containsZshExtendedLine(rawLines []string) bool {
	for _, line := range rawLines {
		if parseZshExtendedLine(line) != "" {
			return true
		}
	}

	return false
}

// firstZshExtendedLine returns the first raw line index that starts a zsh history entry.
func firstZshExtendedLine(rawLines []string) int {
	for i, line := range rawLines {
		if parseZshExtendedLine(line) != "" {
			return i
		}
	}

	return -1
}

// parseLine extracts command text from bash or zsh history lines.
func parseLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	if looksLikeBashTimestamp(line) {
		return ""
	}

	if commandText := parseZshExtendedLine(line); commandText != "" {
		return commandText
	}

	return line
}

// parseZshExtendedLine extracts command text from zsh extended-history format.
func parseZshExtendedLine(line string) string {
	if !strings.HasPrefix(line, ": ") {
		return ""
	}

	index := strings.Index(line, ";")
	if index < 0 || index+1 >= len(line) {
		return ""
	}

	return strings.TrimSpace(line[index+1:])
}

// looksLikeBashTimestamp reports whether a line is a bash history timestamp.
func looksLikeBashTimestamp(line string) bool {
	if len(line) < 2 || line[0] != '#' {
		return false
	}

	for _, r := range line[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// uniqueSources removes empty and duplicate history sources.
func uniqueSources(sources []Source) []Source {
	seen := make(map[string]struct{}, len(sources))
	result := make([]Source, 0, len(sources))

	for _, source := range sources {
		if source.Path == "" {
			continue
		}

		if _, ok := seen[source.Path]; ok {
			continue
		}

		seen[source.Path] = struct{}{}
		result = append(result, source)
	}

	return result
}

// expandHome replaces a leading ~/ with the current user's home directory.
func expandHome(path string, home string) string {
	if path == "~" {
		return home
	}

	if after, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(home, after)
	}

	return path
}

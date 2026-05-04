package command

import "strings"

// Command represents one executable shell command known to recall.
type Command struct {
	// Text is the command line exactly as recall should return it.
	Text string

	// Description explains why the command is useful.
	Description string

	// Source identifies where the command came from.
	Source string

	// Pinned marks commands explicitly saved by the user.
	Pinned bool
}

// New creates a normalized command value.
func New(text string, source string, pinned bool) Command {
	return NewWithDescription(text, "", source, pinned)
}

// NewWithDescription creates a normalized command value with an optional description.
func NewWithDescription(text string, description string, source string, pinned bool) Command {
	return Command{
		Text:        Normalize(text),
		Description: strings.TrimSpace(description),
		Source:      source,
		Pinned:      pinned,
	}
}

// Normalize trims surrounding whitespace from a command line.
func Normalize(text string) string {
	return strings.TrimSpace(text)
}

// IsZero reports whether the command contains no executable text.
func (c Command) IsZero() bool {
	return c.Text == ""
}

// Match reports whether the command or description contains all query terms.
func (c Command) Match(query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}

	text := strings.ToLower(c.Text)
	description := strings.ToLower(c.Description)

	for term := range strings.FieldsSeq(query) {
		if !strings.Contains(text, term) && !strings.Contains(description, term) {
			return false
		}
	}

	return true
}

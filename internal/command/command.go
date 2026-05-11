package command

import "strings"

// Command represents one executable shell command known to recall.
type Command struct {
	Text        string // Text is the command line exactly as recall should return it.
	Description string // Description explains why the command is useful.
	Source      string // Source identifies where the command came from.
	Pinned      bool   // Pinned marks commands explicitly saved by the user.
}

// NewCommand creates a normalized command value.
func NewCommand(text string, source string, pinned bool) Command {
	return NewCommandWithDescription(text, "", source, pinned)
}

// NewCommandWithDescription creates a normalized command value with an optional description.
func NewCommandWithDescription(text string, description string, source string, pinned bool) Command {
	return Command{
		Text:        strings.TrimSpace(text),
		Description: strings.TrimSpace(description),
		Source:      source,
		Pinned:      pinned,
	}
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

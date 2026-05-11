package runner

import (
	"encoding/base64"
	"strings"

	"github.com/containeroo/recall/internal/command"
	"github.com/containeroo/recall/internal/history"
	"github.com/containeroo/recall/internal/index"
	"github.com/containeroo/recall/internal/store"
)

// SearchOptions configures one search invocation.
type SearchOptions struct {
	PinsOnly    bool     // PinsOnly limits search results to pinned commands.
	HistoryOnly bool     // HistoryOnly limits search results to shell history commands.
	ExcludePins bool     // ExcludePins removes history commands that are already pinned.
	Format      string   // Format controls how search results are printed.
	Query       []string // Query contains the search terms used to match commands.
}

// RunSearch scans pinned commands and shell history for matching commands.
func RunSearch(r Runner, cfg Config, opts SearchOptions) error {
	items, err := r.searchCommands(cfg, opts.PinsOnly, opts.HistoryOnly, opts.ExcludePins)
	if err != nil {
		r.logger.Warnf("load commands: %v", err)
	}

	query := strings.Join(opts.Query, " ")
	results := index.NewIndex(items).Search(query, cfg.Limit)

	switch opts.Format {
	case "fzf":
		r.printFZFResults(results)
	default:
		r.printPlainResults(results)
	}

	return nil
}

// searchCommands returns pinned commands, history commands, or both.
func (r Runner) searchCommands(cfg Config, pinsOnly, historyOnly, excludePins bool) ([]command.Command, error) {
	switch {
	case pinsOnly:
		return r.loadPinnedCommands(cfg)

	case historyOnly:
		return r.loadHistoryCommands(cfg, excludePins)

	default:
		return r.loadCommands(cfg, excludePins)
	}
}

// loadPinnedCommands reads pinned commands from the recall store.
func (r Runner) loadPinnedCommands(cfg Config) ([]command.Command, error) {
	st := store.NewStore(r.storePath(cfg))
	return st.Pins()
}

// loadHistoryCommands reads shell history commands from configured history sources.
func (r Runner) loadHistoryCommands(cfg Config, excludePins bool) ([]command.Command, error) {
	historyCommands, historyErr := history.NewReader(r.historySources(cfg), cfg.HistoryLimit).Read()
	if !excludePins {
		return historyCommands, historyErr
	}

	pins, pinErr := r.loadPinnedCommands(cfg)
	filtered := excludePinnedCommands(historyCommands, pins)

	if pinErr != nil {
		return filtered, pinErr
	}

	return filtered, historyErr
}

// loadCommands reads pinned commands and shell history into one command list.
func (r Runner) loadCommands(cfg Config, excludePins bool) ([]command.Command, error) {
	pins, pinErr := r.loadPinnedCommands(cfg)
	historyCommands, historyErr := r.loadHistoryCommands(cfg, excludePins)

	items := make([]command.Command, 0, len(pins)+len(historyCommands))
	items = append(items, pins...)
	items = append(items, historyCommands...)

	if pinErr != nil {
		return items, pinErr
	}

	return items, historyErr
}

// excludePinnedCommands removes history commands that are already pinned.
func excludePinnedCommands(historyCommands []command.Command, pins []command.Command) []command.Command {
	pinned := make(map[string]struct{}, len(pins))
	for _, pin := range pins {
		key := commandKey(pin.Text)
		if key == "" {
			continue
		}

		pinned[key] = struct{}{}
	}

	filtered := make([]command.Command, 0, len(historyCommands))
	for _, item := range historyCommands {
		if _, ok := pinned[commandKey(item.Text)]; ok {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// commandKey returns the normalized deduplication key for a command.
func commandKey(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// printPlainResults prints one command per line.
func (r Runner) printPlainResults(results []command.Command) {
	for _, result := range results {
		r.output.Line(result.Text)
	}
}

// printFZFResults prints fzf rows with a display command and hidden payload.
func (r Runner) printFZFResults(results []command.Command) {
	for _, result := range results {
		r.output.Linef(
			"%s\t%s\t%s\t%s\t%s",
			resultKind(result),
			displayFZFField(result.Text),
			displayFZFField(result.Description),
			"command",
			encodeFZFPayload(result.Text),
		)
	}
}

// resultKind returns the stable machine-readable result kind.
func resultKind(result command.Command) string {
	if result.Pinned {
		return "pin"
	}

	return "history"
}

// displayFZFField normalizes text for one-line fzf display.
func displayFZFField(value string) string {
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.ReplaceAll(value, "\r", `\r`)
	value = strings.ReplaceAll(value, "\n", `\n`)

	return strings.TrimSpace(value)
}

// encodeFZFPayload encodes raw command text for hidden fzf transport.
func encodeFZFPayload(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

package runner

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/containeroo/recall/internal/command"
	"github.com/containeroo/recall/internal/store"
)

// PickerCmd contains options for the interactive picker candidate producer.
type PickerCmd struct {
	// Format controls how picker candidates are printed.
	Format string `name:"format" enum:"fzf,json" default:"fzf" help:"Picker output format."`

	// PinLimit limits how many pinned commands are loaded, with zero meaning unlimited.
	PinLimit int `name:"pin-limit" default:"0" help:"Maximum number of pinned commands to load; zero means unlimited."`
}

// pickerCandidate is one selectable picker candidate.
type pickerCandidate struct {
	// Kind is the visible source type, such as pin or history.
	Kind string

	// Text is the raw command text.
	Text string

	// Description explains why the command is useful.
	Description string

	// PayloadType describes how Payload should be interpreted.
	PayloadType string

	// Payload is the selected payload returned by the picker row.
	Payload string
}

// pickerOutput is the JSON representation of one picker candidate.
type pickerOutput struct {
	// Kind is the candidate source type.
	Kind string `json:"kind"`

	// Text is the raw command text.
	Text string `json:"text"`

	// Description explains why the command is useful.
	Description string `json:"description,omitempty"`

	// PayloadType describes how Payload should be interpreted.
	PayloadType string `json:"payload_type"`

	// Payload is the original selection payload.
	Payload string `json:"payload"`

	// CommandPayload is the base64-encoded raw command text.
	CommandPayload string `json:"command_payload"`

	// DescriptionPayload is the base64-encoded description.
	DescriptionPayload string `json:"description_payload,omitempty"`

	// Display is the formatted picker display row.
	Display string `json:"display"`
}

// pickerStyle contains fzf display customization.
type pickerStyle struct {
	// CommandWidth is the maximum visible command column width.
	CommandWidth int

	// DescriptionWidth is the maximum visible description column width.
	DescriptionWidth int

	// PinLabel is the visible label for pinned commands.
	PinLabel string

	// HistoryLabel is the visible label for history commands.
	HistoryLabel string

	// DescriptionColor is the ANSI color prefix for descriptions.
	DescriptionColor string

	// ResetColor resets ANSI formatting.
	ResetColor string
}

// Run prints picker candidates in the requested machine-readable format.
func (c *PickerCmd) Run(r Runner, cli *CLI) error {
	cfg := cli.Config(r.home)

	candidates, err := r.pickerCandidates(cfg, c.PinLimit)
	if err != nil {
		r.logger.Warnf("load picker candidates: %v", err)
	}

	style := defaultPickerStyle()

	switch c.Format {
	case "fzf":
		_, _ = r.output.Writer().Write(formatPickerRows(candidates, style))
	case "json":
		content, err := formatPickerJSON(candidates, style)
		if err != nil {
			return err
		}

		_, _ = r.output.Writer().Write(content)
	default:
		return fmt.Errorf("unsupported picker format %q", c.Format)
	}

	return nil
}

// pickerCandidates returns pins followed by shell-provided history candidates.
func (r Runner) pickerCandidates(cfg Config, pinLimit int) ([]pickerCandidate, error) {
	pins, pinErr := r.pickerPinCandidates(cfg, pinLimit)
	historyCandidates, historyErr := readPickerHistoryCandidates(r.input)

	candidates := make([]pickerCandidate, 0, len(pins)+len(historyCandidates))
	candidates = append(candidates, pins...)
	candidates = append(candidates, historyCandidates...)
	candidates = dedupePickerCandidates(candidates)

	if pinErr != nil {
		return candidates, pinErr
	}

	return candidates, historyErr
}

// pickerPinCandidates returns pinned commands as picker candidates.
func (r Runner) pickerPinCandidates(cfg Config, limit int) ([]pickerCandidate, error) {
	st := store.New(r.storePath(cfg))

	pins, err := st.Pins()
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(pins) > limit {
		pins = pins[:limit]
	}

	candidates := make([]pickerCandidate, 0, len(pins))
	for _, pin := range pins {
		if pin.IsZero() {
			continue
		}

		candidates = append(candidates, pickerCandidate{
			Kind:        "pin",
			Text:        pin.Text,
			Description: pin.Description,
			PayloadType: "command",
			Payload:     encodePickerPayload(pin.Text),
		})
	}

	return candidates, nil
}

// readPickerHistoryCandidates reads shell-provided history candidates from input.
func readPickerHistoryCandidates(input io.Reader) ([]pickerCandidate, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	records := splitPickerInputRecords(data)
	candidates := make([]pickerCandidate, 0, len(records))

	for _, record := range records {
		candidate, ok := parsePickerHistoryRecord(record)
		if !ok {
			continue
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// splitPickerInputRecords splits stdin records by NUL when present, otherwise by newline.
func splitPickerInputRecords(data []byte) []string {
	separator := byte('\n')
	if bytes.Contains(data, []byte{0}) {
		separator = 0
	}

	parts := bytes.Split(data, []byte{separator})
	records := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		records = append(records, string(part))
	}

	return records
}

// parsePickerHistoryRecord parses one shell-provided history record.
func parsePickerHistoryRecord(record string) (pickerCandidate, bool) {
	fields := strings.SplitN(record, "\t", 3)
	if len(fields) == 3 {
		payloadType := strings.TrimSpace(fields[0])
		payload := strings.TrimSpace(fields[1])
		text := command.Normalize(fields[2])

		if payloadType == "" || text == "" {
			return pickerCandidate{}, false
		}

		if payload == "" {
			payloadType = "command"
			payload = encodePickerPayload(text)
		}

		return pickerCandidate{
			Kind:        "history",
			Text:        text,
			PayloadType: payloadType,
			Payload:     payload,
		}, true
	}

	text := command.Normalize(record)
	if text == "" {
		return pickerCandidate{}, false
	}

	return pickerCandidate{
		Kind:        "history",
		Text:        text,
		PayloadType: "command",
		Payload:     encodePickerPayload(text),
	}, true
}

// dedupePickerCandidates removes duplicate commands while preserving priority order.
func dedupePickerCandidates(candidates []pickerCandidate) []pickerCandidate {
	seen := make(map[string]struct{}, len(candidates))
	result := make([]pickerCandidate, 0, len(candidates))

	for _, candidate := range candidates {
		key := commandKey(candidate.Text)
		if key == "" {
			continue
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, candidate)
	}

	return result
}

// defaultPickerStyle returns picker display settings from environment variables.
func defaultPickerStyle() pickerStyle {
	return pickerStyle{
		CommandWidth:     envInt("RECALL_FZF_COMMAND_WIDTH", 70),
		DescriptionWidth: envInt("RECALL_FZF_DESCRIPTION_WIDTH", 45),
		PinLabel:         envString("RECALL_FZF_PIN_LABEL", "\033[38;5;208mpin\033[0m"),
		HistoryLabel:     envString("RECALL_FZF_HISTORY_LABEL", "\033[38;5;245mhistory\033[0m"),
		DescriptionColor: envString("RECALL_FZF_DESCRIPTION_COLOR", "\033[38;5;110m"),
		ResetColor:       "\033[0m",
	}
}

// formatPickerRows formats candidates as NUL-separated fzf rows.
func formatPickerRows(candidates []pickerCandidate, style pickerStyle) []byte {
	var buf bytes.Buffer

	for _, candidate := range candidates {
		output := newPickerOutput(candidate, style)

		fmt.Fprintf(
			&buf,
			"%s\t%s\t%s\t%s\t%s\t%s%c",
			output.PayloadType,
			output.Payload,
			output.CommandPayload,
			output.Kind,
			output.DescriptionPayload,
			output.Display,
			0,
		)
	}

	return buf.Bytes()
}

// formatPickerJSON formats candidates as pretty JSON.
func formatPickerJSON(candidates []pickerCandidate, style pickerStyle) ([]byte, error) {
	output := make([]pickerOutput, 0, len(candidates))
	for _, candidate := range candidates {
		output = append(output, newPickerOutput(candidate, style))
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return nil, fmt.Errorf("encode picker json: %w", err)
	}

	return buf.Bytes(), nil
}

// newPickerOutput builds the transport representation for one candidate.
func newPickerOutput(candidate pickerCandidate, style pickerStyle) pickerOutput {
	return pickerOutput{
		Kind:               candidate.Kind,
		Text:               candidate.Text,
		Description:        candidate.Description,
		PayloadType:        candidate.PayloadType,
		Payload:            candidate.Payload,
		CommandPayload:     encodePickerPayload(candidate.Text),
		DescriptionPayload: encodePickerPayload(candidate.Description),
		Display:            formatPickerDisplay(candidate, style),
	}
}

// formatPickerDisplay formats one visible fzf row.
func formatPickerDisplay(candidate pickerCandidate, style pickerStyle) string {
	lines := strings.Split(displayPickerCommand(candidate.Text), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	first := fmt.Sprintf(
		"%s  %s  %s",
		pickerLabel(candidate.Kind, style),
		fitPickerField(lines[0], style.CommandWidth),
		colorPickerDescription(fitPickerField(displayPickerField(candidate.Description), style.DescriptionWidth), style),
	)

	if len(lines) == 1 {
		return first
	}

	var display strings.Builder
	display.WriteString(first)

	indent := strings.Repeat(" ", 9)
	for _, line := range lines[1:] {
		display.WriteByte('\n')
		display.WriteString(indent)
		display.WriteString(fitPickerField(line, style.CommandWidth))
	}

	return display.String()
}

// pickerLabel returns the visible source label for one candidate.
func pickerLabel(kind string, style pickerStyle) string {
	switch kind {
	case "pin":
		return style.PinLabel + "    "
	case "history":
		return style.HistoryLabel
	default:
		return fmt.Sprintf("%-7s", kind)
	}
}

// displayPickerCommand normalizes command text for fzf display while preserving newlines.
func displayPickerCommand(value string) string {
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.ReplaceAll(value, "\r", `\r`)

	return strings.TrimSpace(value)
}

// displayPickerField normalizes one non-command field for fzf display.
func displayPickerField(value string) string {
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.ReplaceAll(value, "\r", `\r`)
	value = strings.ReplaceAll(value, "\n", `\n`)

	return strings.TrimSpace(value)
}

// fitPickerField truncates or pads one visible fzf field.
func fitPickerField(value string, width int) string {
	if width <= 0 {
		return value
	}

	if len(value) > width {
		if width > 3 {
			return value[:width-3] + "..."
		}

		return value[:width]
	}

	return fmt.Sprintf("%-*s", width, value)
}

// colorPickerDescription applies the configured description color.
func colorPickerDescription(description string, style pickerStyle) string {
	if description == "" {
		return ""
	}

	return style.DescriptionColor + description + style.ResetColor
}

// encodePickerPayload encodes raw text for fzf transport.
func encodePickerPayload(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

// envString returns an environment value or fallback.
func envString(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

// envInt returns an integer environment value or fallback.
func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

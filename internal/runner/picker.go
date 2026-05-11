package runner

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/containeroo/recall/internal/store"
)

// PickerOptions configures one picker invocation.
type PickerOptions struct {
	Format   string // Format controls how picker candidates are printed.
	PinLimit int    // PinLimit limits how many pinned commands are loaded, with zero meaning unlimited.
}

// pickerCandidate is one selectable picker candidate.
type pickerCandidate struct {
	Kind        string // Kind is the visible source type, such as pin or history.
	Text        string // Text is the raw command text.
	Description string // Description explains why the command is useful.
	PayloadType string // PayloadType describes how Payload should be interpreted.
	Payload     string // Payload is the selected payload returned by the picker row.
}

// pickerOutput is the JSON representation of one picker candidate.
type pickerOutput struct {
	Kind               string `json:"kind"`                          // Kind is the candidate source type.
	Text               string `json:"text"`                          // Text is the raw command text.
	Description        string `json:"description,omitempty"`         // Description explains why the command is useful.
	PayloadType        string `json:"payload_type"`                  // PayloadType describes how Payload should be interpreted.
	Payload            string `json:"payload"`                       // Payload is the original selection payload.
	CommandPayload     string `json:"command_payload"`               // CommandPayload is the base64-encoded raw command text.
	DescriptionPayload string `json:"description_payload,omitempty"` // DescriptionPayload is the base64-encoded description.
	Display            string `json:"display"`                       // Display is the formatted picker display row.
}

// pickerStyle contains fzf display customization.
type pickerStyle struct {
	CommandWidth     int    // CommandWidth is the maximum visible command column width.
	DescriptionWidth int    // DescriptionWidth is the maximum visible description column width.
	PinLabel         string // PinLabel is the visible label for pinned commands.
	HistoryLabel     string // HistoryLabel is the visible label for history commands.
	DescriptionColor string // DescriptionColor is the ANSI color prefix for descriptions.
	ResetColor       string // ResetColor resets ANSI formatting.
}

// RunPicker prints picker candidates in the requested machine-readable format.
func RunPicker(r Runner, cfg Config, opts PickerOptions) error {
	candidates, err := r.pickerCandidates(cfg, opts.PinLimit)
	if err != nil {
		r.logger.Warnf("load picker candidates: %v", err)
	}

	style := defaultPickerStyle()

	switch opts.Format {
	case "fzf":
		_, _ = r.output.Writer().Write(formatPickerRows(candidates, style))
	case "json":
		content, err := formatPickerJSON(candidates, style)
		if err != nil {
			return err
		}

		_, _ = r.output.Writer().Write(content)
	default:
		return fmt.Errorf("unsupported picker format %q", opts.Format)
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
	st := store.NewStore(r.storePath(cfg))

	pins, err := st.Pins()
	if err != nil {
		return nil, err
	}

	if exceedsLimit(pins, limit) {
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
		text := strings.TrimSpace(fields[2])

		if payloadType == "" || text == "" {
			return pickerCandidate{}, false
		}

		if payloadType == "command" || payload == "" {
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

	text := strings.TrimSpace(record)
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
		CommandWidth:     envInt(EnvCommandWidth, defaultCommandWidth),
		DescriptionWidth: envInt(EnvDescriptionWidth, defaultDescriptionWidth),
		PinLabel:         envString(EnvPinLabel, defaultPinLabel),
		HistoryLabel:     envString(EnvHistoryLabel, defaultHistoryLabel),
		DescriptionColor: envString(EnvDescriptionColor, defaultDescriptionColor),
		ResetColor:       defaultResetColor,
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

package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containeroo/recall/internal/command"
)

// Store reads and writes recall's persistent data file.
type Store struct {
	Path string // Path is the JSON file path used for persistent storage.
}

// Data is the on-disk JSON format for recall data.
type Data struct {
	Pins []Pin `json:"pins"` // Pins contains user-pinned commands in priority order.
}

// Pin is the on-disk JSON format for one pinned command.
type Pin struct {
	Command     string `json:"command"`               // Command is the executable command text.
	Description string `json:"description,omitempty"` // Description explains why the command is useful.
}

// MatchesCommand reports whether pin matches command text after trimming whitespace.
func (p Pin) MatchesCommand(text string) bool {
	commandText := strings.TrimSpace(p.Command)
	text = strings.TrimSpace(text)

	if commandText == "" || text == "" {
		return false
	}

	return strings.EqualFold(commandText, text)
}

// DefaultDir returns recall's default XDG-compatible data directory.
func DefaultDir(home string) string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "recall")
	}

	return filepath.Join(home, ".local", "share", "recall")
}

// NewStore creates a persistent store backed by path.
func NewStore(path string) Store {
	return Store{
		Path: path,
	}
}

// Load reads recall data from disk.
func (s Store) Load() (Data, error) {
	content, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return Data{}, nil
	}

	if err != nil {
		return Data{}, err
	}

	if len(bytes.TrimSpace(content)) == 0 {
		return Data{}, nil
	}

	return decodeData(content)
}

// Save writes recall data to disk.
func (s Store) Save(data Data) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(normalizeData(data), "", "  ")
	if err != nil {
		return err
	}

	content = append(content, '\n')
	return os.WriteFile(s.Path, content, 0o644)
}

// AddPin stores a command as the highest-priority pinned command.
func (s Store) AddPin(text string, description string) error {
	text = strings.TrimSpace(text)
	description = strings.TrimSpace(description)

	if text == "" {
		return nil
	}

	data, err := s.Load()
	if err != nil {
		return err
	}

	pin := Pin{
		Command:     text,
		Description: description,
	}

	pins := make([]Pin, 0, len(data.Pins)+1)

	for _, existing := range data.Pins {
		existing = normalizePin(existing)
		if existing.Command == "" {
			continue
		}

		if existing.MatchesCommand(text) {
			if pin.Description == "" {
				pin.Description = existing.Description
			}

			continue
		}

		pins = append(pins, existing)
	}

	data.Pins = append([]Pin{pin}, pins...)
	return s.Save(data)
}

// DescribePin updates the description for an existing pinned command.
func (s Store) DescribePin(text string, description string) error {
	text = strings.TrimSpace(text)
	description = strings.TrimSpace(description)

	if text == "" {
		return errors.New("describe requires non-empty command text")
	}

	data, err := s.Load()
	if err != nil {
		return err
	}

	pins := make([]Pin, 0, len(data.Pins))
	updated := false

	for _, pin := range data.Pins {
		pin = normalizePin(pin)
		if pin.Command == "" {
			continue
		}

		if !updated && pin.MatchesCommand(text) {
			pin.Description = description
			updated = true
		}

		pins = append(pins, pin)
	}

	if !updated {
		return fmt.Errorf("command is not pinned: %s", text)
	}

	data.Pins = pins
	return s.Save(data)
}

// RemovePin removes an existing pinned command.
func (s Store) RemovePin(text string) error {
	text = strings.TrimSpace(text)

	if text == "" {
		return errors.New("remove requires non-empty command text")
	}

	data, err := s.Load()
	if err != nil {
		return err
	}

	pins := make([]Pin, 0, len(data.Pins))
	removed := false

	for _, pin := range data.Pins {
		pin = normalizePin(pin)
		if pin.Command == "" {
			continue
		}

		if pin.MatchesCommand(text) {
			removed = true
			continue
		}

		pins = append(pins, pin)
	}

	if !removed {
		return fmt.Errorf("command is not pinned: %s", text)
	}

	data.Pins = pins
	return s.Save(data)
}

// Pins returns pinned commands in priority order.
func (s Store) Pins() ([]command.Command, error) {
	data, err := s.Load()
	if err != nil {
		return nil, err
	}

	pins := make([]command.Command, 0, len(data.Pins))
	for _, pin := range data.Pins {
		pin = normalizePin(pin)

		item := command.NewCommandWithDescription(pin.Command, pin.Description, "pin", true)
		if item.IsZero() {
			continue
		}

		pins = append(pins, item)
	}

	return pins, nil
}

// decodeData decodes the current store format.
func decodeData(content []byte) (Data, error) {
	var data Data
	if err := json.Unmarshal(content, &data); err != nil {
		return Data{}, fmt.Errorf("decode store data: %w", err)
	}

	return normalizeData(data), nil
}

// normalizeData trims all persisted pin values.
func normalizeData(data Data) Data {
	pins := make([]Pin, 0, len(data.Pins))
	for _, pin := range data.Pins {
		pin = normalizePin(pin)
		if pin.Command == "" {
			continue
		}

		pins = append(pins, pin)
	}

	data.Pins = pins
	return data
}

// normalizePin trims one persisted pin value.
func normalizePin(pin Pin) Pin {
	return Pin{
		Command:     strings.TrimSpace(pin.Command),
		Description: strings.TrimSpace(pin.Description),
	}
}

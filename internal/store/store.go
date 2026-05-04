package store

import (
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
	// Path is the JSON file path used for persistent storage.
	Path string
}

// Data is the on-disk JSON format for recall data.
type Data struct {
	// Pins contains user-pinned commands in priority order.
	Pins []Pin `json:"pins"`
}

// Pin is the on-disk JSON format for one pinned command.
type Pin struct {
	// Command is the executable command text.
	Command string `json:"command"`

	// Description explains why the command is useful.
	Description string `json:"description,omitempty"`
}

// legacyData is the old on-disk JSON format used before pin descriptions.
type legacyData struct {
	// Pins contains user-pinned command text in priority order.
	Pins []string `json:"pins"`
}

// DefaultDir returns recall's default XDG-compatible data directory.
func DefaultDir(home string) string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "recall")
	}

	return filepath.Join(home, ".local", "share", "recall")
}

// New creates a persistent store backed by path.
func New(path string) Store {
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

	if len(strings.TrimSpace(string(content))) == 0 {
		return Data{}, nil
	}

	data, err := decodeData(content)
	if err != nil {
		return Data{}, err
	}

	return data, nil
}

// Save writes recall data to disk.
func (s Store) Save(data Data) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	content = append(content, '\n')
	return os.WriteFile(s.Path, content, 0o644)
}

// AddPin stores a command as the highest-priority pinned command.
func (s Store) AddPin(text string, description string) error {
	text = command.Normalize(text)
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
		if !strings.EqualFold(existing.Command, text) {
			continue
		}

		if description == "" {
			pin.Description = existing.Description
		}

		break
	}

	pins = append(pins, pin)

	for _, existing := range data.Pins {
		if strings.EqualFold(existing.Command, text) {
			continue
		}

		pins = append(pins, normalizePin(existing))
	}

	data.Pins = pins
	return s.Save(data)
}

// DescribePin updates the description for an existing pinned command.
func (s Store) DescribePin(text string, description string) error {
	text = command.Normalize(text)
	description = strings.TrimSpace(description)

	if text == "" {
		return fmt.Errorf("describe requires non-empty command text")
	}

	data, err := s.Load()
	if err != nil {
		return err
	}

	for i, pin := range data.Pins {
		if !strings.EqualFold(pin.Command, text) {
			continue
		}

		data.Pins[i] = normalizePin(Pin{
			Command:     pin.Command,
			Description: description,
		})

		return s.Save(data)
	}

	return fmt.Errorf("command is not pinned: %s", text)
}

// RemovePin removes an existing pinned command.
func (s Store) RemovePin(text string) error {
	text = command.Normalize(text)
	if text == "" {
		return fmt.Errorf("remove requires non-empty command text")
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

		if strings.EqualFold(pin.Command, text) {
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

		item := command.NewWithDescription(pin.Command, pin.Description, "pin", true)
		if item.IsZero() {
			continue
		}

		pins = append(pins, item)
	}

	return pins, nil
}

// decodeData decodes the current store format or migrates the legacy format.
func decodeData(content []byte) (Data, error) {
	var data Data
	if err := json.Unmarshal(content, &data); err == nil {
		return normalizeData(data), nil
	}

	var legacy legacyData
	if err := json.Unmarshal(content, &legacy); err != nil {
		return Data{}, err
	}

	pins := make([]Pin, 0, len(legacy.Pins))
	for _, text := range legacy.Pins {
		text = command.Normalize(text)
		if text == "" {
			continue
		}

		pins = append(pins, Pin{
			Command: text,
		})
	}

	return Data{
		Pins: pins,
	}, nil
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
		Command:     command.Normalize(pin.Command),
		Description: strings.TrimSpace(pin.Description),
	}
}

package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/containeroo/recall/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPinMatchesCommand verifies case-insensitive command matching.
func TestPinMatchesCommand(t *testing.T) {
	t.Parallel()

	t.Run("matches trimmed command text", func(t *testing.T) {
		t.Parallel()

		pin := Pin{
			Command: "  kubectl get pods  ",
		}

		assert.True(t, pin.MatchesCommand("kubectl get pods"))
	})

	t.Run("matches command text case insensitively", func(t *testing.T) {
		t.Parallel()

		pin := Pin{
			Command: "kubectl get pods",
		}

		assert.True(t, pin.MatchesCommand("KUBECTL GET PODS"))
	})

	t.Run("rejects empty pin command", func(t *testing.T) {
		t.Parallel()

		pin := Pin{
			Command: "   ",
		}

		assert.False(t, pin.MatchesCommand("kubectl get pods"))
	})

	t.Run("rejects empty command text", func(t *testing.T) {
		t.Parallel()

		pin := Pin{
			Command: "kubectl get pods",
		}

		assert.False(t, pin.MatchesCommand("   "))
	})
}

// TestDefaultDir verifies default data directory selection.
func TestDefaultDir(t *testing.T) {
	t.Run("uses xdg data home when set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")

		assert.Equal(t, filepath.Join("/tmp/xdg-data", "recall"), DefaultDir("/home/user"))
	})

	t.Run("uses local share directory when xdg data home is empty", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")

		assert.Equal(t, filepath.Join("/home/user", ".local", "share", "recall"), DefaultDir("/home/user"))
	})
}

// TestNewStore verifies store construction.
func TestNewStore(t *testing.T) {
	t.Parallel()

	t.Run("stores path", func(t *testing.T) {
		t.Parallel()

		st := NewStore("/tmp/recall/store.json")

		assert.Equal(t, "/tmp/recall/store.json", st.Path)
	})
}

// TestLoad verifies loading persisted store data.
func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("returns empty data when file is missing", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		data, err := st.Load()

		require.NoError(t, err)
		assert.Equal(t, Data{}, data)
	})

	t.Run("returns empty data when file is blank", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)
		writeStoreContent(t, st.Path, "  \n\t  ")

		data, err := st.Load()

		require.NoError(t, err)
		assert.Equal(t, Data{}, data)
	})

	t.Run("loads and normalizes current data", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)
		writeStoreContent(t, st.Path, `{
			"pins": [
				{
					"command": "  kubectl get pods  ",
					"description": "  Show pods  "
				},
				{
					"command": "   ",
					"description": "ignored"
				}
			]
		}`)

		data, err := st.Load()

		require.NoError(t, err)
		assert.Equal(t, Data{
			Pins: []Pin{
				{
					Command:     "kubectl get pods",
					Description: "Show pods",
				},
			},
		}, data)
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)
		writeStoreContent(t, st.Path, "{")

		data, err := st.Load()

		require.Error(t, err)
		assert.Equal(t, Data{}, data)
		assert.ErrorContains(t, err, "decode store data")
	})
}

// TestSave verifies writing persisted store data.
func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("creates parent directory and writes normalized json", func(t *testing.T) {
		t.Parallel()

		st := NewStore(filepath.Join(t.TempDir(), "nested", "store.json"))

		err := st.Save(Data{
			Pins: []Pin{
				{
					Command:     "  kubectl get pods  ",
					Description: "  Show pods  ",
				},
				{
					Command:     "   ",
					Description: "ignored",
				},
			},
		})

		require.NoError(t, err)

		content, err := os.ReadFile(st.Path)
		require.NoError(t, err)

		expected := `{
  "pins": [
    {
      "command": "kubectl get pods",
      "description": "Show pods"
    }
  ]
}
`

		assert.Equal(t, expected, string(content))
	})

	t.Run("omits empty descriptions", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		err := st.Save(Data{
			Pins: []Pin{
				{
					Command: "kubectl get pods",
				},
			},
		})

		require.NoError(t, err)

		content, err := os.ReadFile(st.Path)
		require.NoError(t, err)
		assert.NotContains(t, string(content), "description")
	})
}

// TestAddPin verifies adding and promoting pinned commands.
func TestAddPin(t *testing.T) {
	t.Parallel()

	t.Run("stores new pin as highest priority", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("go test ./...", "Run tests"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		data := readStoreData(t, st.Path)

		assert.Equal(t, Data{
			Pins: []Pin{
				{
					Command:     "kubectl get pods",
					Description: "Show pods",
				},
				{
					Command:     "go test ./...",
					Description: "Run tests",
				},
			},
		}, data)
	})

	t.Run("promotes existing pin and preserves description when empty", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("go test ./...", "Run tests"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.AddPin("go test ./...", ""))

		data := readStoreData(t, st.Path)

		require.Len(t, data.Pins, 2)
		assert.Equal(t, "go test ./...", data.Pins[0].Command)
		assert.Equal(t, "Run tests", data.Pins[0].Description)
		assert.Equal(t, "kubectl get pods", data.Pins[1].Command)
	})

	t.Run("promotes existing pin and updates description when provided", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("go test ./...", "Run tests"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.AddPin("go test ./...", "Run all tests"))

		data := readStoreData(t, st.Path)

		require.Len(t, data.Pins, 2)
		assert.Equal(t, "go test ./...", data.Pins[0].Command)
		assert.Equal(t, "Run all tests", data.Pins[0].Description)
		assert.Equal(t, "kubectl get pods", data.Pins[1].Command)
	})

	t.Run("matches existing pin case insensitively", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.AddPin("KUBECTL GET PODS", ""))

		data := readStoreData(t, st.Path)

		require.Len(t, data.Pins, 1)
		assert.Equal(t, "KUBECTL GET PODS", data.Pins[0].Command)
		assert.Equal(t, "Show pods", data.Pins[0].Description)
	})

	t.Run("ignores empty command", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("   ", "Empty"))

		data, err := st.Load()

		require.NoError(t, err)
		assert.Empty(t, data.Pins)
	})
}

// TestDescribePin verifies updating pinned command descriptions.
func TestDescribePin(t *testing.T) {
	t.Parallel()

	t.Run("updates existing pin description", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.DescribePin("kubectl get pods", "  List pods  "))

		data := readStoreData(t, st.Path)

		require.Len(t, data.Pins, 1)
		assert.Equal(t, "kubectl get pods", data.Pins[0].Command)
		assert.Equal(t, "List pods", data.Pins[0].Description)
	})

	t.Run("matches existing pin case insensitively", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.DescribePin("KUBECTL GET PODS", "List pods"))

		data := readStoreData(t, st.Path)

		require.Len(t, data.Pins, 1)
		assert.Equal(t, "kubectl get pods", data.Pins[0].Command)
		assert.Equal(t, "List pods", data.Pins[0].Description)
	})

	t.Run("rejects empty command", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		err := st.DescribePin("   ", "Description")

		require.Error(t, err)
		assert.ErrorContains(t, err, "describe requires non-empty command text")
	})

	t.Run("rejects missing pin", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		err := st.DescribePin("kubectl get pods", "Show pods")

		require.Error(t, err)
		assert.ErrorContains(t, err, "command is not pinned: kubectl get pods")
	})
}

// TestRemovePin verifies removing pinned commands.
func TestRemovePin(t *testing.T) {
	t.Parallel()

	t.Run("removes existing pin", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("go test ./...", "Run tests"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.RemovePin("kubectl get pods"))

		data := readStoreData(t, st.Path)

		assert.Equal(t, Data{
			Pins: []Pin{
				{
					Command:     "go test ./...",
					Description: "Run tests",
				},
			},
		}, data)
	})

	t.Run("matches existing pin case insensitively", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))
		require.NoError(t, st.RemovePin("KUBECTL GET PODS"))

		data := readStoreData(t, st.Path)

		assert.Empty(t, data.Pins)
	})

	t.Run("rejects empty command", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		err := st.RemovePin("   ")

		require.Error(t, err)
		assert.ErrorContains(t, err, "remove requires non-empty command text")
	})

	t.Run("rejects missing pin", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)

		err := st.RemovePin("kubectl get pods")

		require.Error(t, err)
		assert.ErrorContains(t, err, "command is not pinned: kubectl get pods")
	})
}

// TestPins verifies converting persisted pins into command values.
func TestPins(t *testing.T) {
	t.Parallel()

	t.Run("returns pinned commands in priority order", func(t *testing.T) {
		t.Parallel()

		st := newTestStore(t)
		writeStoreContent(t, st.Path, `{
			"pins": [
				{
					"command": "  kubectl get pods  ",
					"description": "  Show pods  "
				},
				{
					"command": "go test ./...",
					"description": "Run tests"
				},
				{
					"command": "   ",
					"description": "ignored"
				}
			]
		}`)

		pins, err := st.Pins()

		require.NoError(t, err)
		assert.Equal(t, []command.Command{
			command.NewCommandWithDescription("kubectl get pods", "Show pods", "pin", true),
			command.NewCommandWithDescription("go test ./...", "Run tests", "pin", true),
		}, pins)
	})
}

// TestDecodeData verifies decoding the current store format.
func TestDecodeData(t *testing.T) {
	t.Parallel()

	t.Run("decodes current store format", func(t *testing.T) {
		t.Parallel()

		data, err := decodeData([]byte(`{
			"pins": [
				{
					"command": "  kubectl get pods  ",
					"description": "  Show pods  "
				}
			]
		}`))

		require.NoError(t, err)
		assert.Equal(t, Data{
			Pins: []Pin{
				{
					Command:     "kubectl get pods",
					Description: "Show pods",
				},
			},
		}, data)
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		t.Parallel()

		data, err := decodeData([]byte("{"))

		require.Error(t, err)
		assert.Equal(t, Data{}, data)
		assert.ErrorContains(t, err, "decode store data")
	})

	t.Run("rejects legacy pin list format", func(t *testing.T) {
		t.Parallel()

		data, err := decodeData([]byte(`{"pins":["kubectl get pods"]}`))

		require.Error(t, err)
		assert.Equal(t, Data{}, data)
		assert.ErrorContains(t, err, "decode store data")
	})
}

// TestNormalizeData verifies trimming and filtering store data.
func TestNormalizeData(t *testing.T) {
	t.Parallel()

	t.Run("trims pins and removes empty commands", func(t *testing.T) {
		t.Parallel()

		data := normalizeData(Data{
			Pins: []Pin{
				{
					Command:     "  kubectl get pods  ",
					Description: "  Show pods  ",
				},
				{
					Command:     "   ",
					Description: "ignored",
				},
				{
					Command:     "go test ./...",
					Description: "  ",
				},
			},
		})

		assert.Equal(t, Data{
			Pins: []Pin{
				{
					Command:     "kubectl get pods",
					Description: "Show pods",
				},
				{
					Command: "go test ./...",
				},
			},
		}, data)
	})
}

// TestNormalizePin verifies trimming one persisted pin.
func TestNormalizePin(t *testing.T) {
	t.Parallel()

	t.Run("trims command and description", func(t *testing.T) {
		t.Parallel()

		pin := normalizePin(Pin{
			Command:     "  kubectl get pods  ",
			Description: "  Show pods  ",
		})

		assert.Equal(t, Pin{
			Command:     "kubectl get pods",
			Description: "Show pods",
		}, pin)
	})
}

// newTestStore creates a store backed by a temporary file.
func newTestStore(t *testing.T) Store {
	t.Helper()

	return NewStore(filepath.Join(t.TempDir(), "store.json"))
}

// writeStoreContent writes raw store content for tests.
func writeStoreContent(t *testing.T, path string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// readStoreData reads and decodes raw store data for tests.
func readStoreData(t *testing.T, path string) Data {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	var data Data
	require.NoError(t, json.Unmarshal(content, &data))

	return data
}

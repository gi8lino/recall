package runner

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/containeroo/recall/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPickerCmdRun(t *testing.T) {
	t.Run("prints fzf picker rows", func(t *testing.T) {
		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)
		r.input = strings.NewReader("zsh-history\t10\tgo test ./...\x00")

		t.Setenv("RECALL_FZF_PIN_LABEL", "pin")
		t.Setenv("RECALL_FZF_HISTORY_LABEL", "history")
		t.Setenv("RECALL_FZF_COMMAND_WIDTH", "20")
		t.Setenv("RECALL_FZF_DESCRIPTION_WIDTH", "20")

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		err := (&PickerCmd{Format: "fzf"}).Run(r, newTestCLI(dataDir))

		require.NoError(t, err)

		output := bytes.TrimSuffix(stdout.Bytes(), []byte{0})
		rows := bytes.Split(output, []byte{0})

		require.Len(t, rows, 2)

		pinFields := strings.SplitN(string(rows[0]), "\t", 6)
		require.Len(t, pinFields, 6)
		assert.Equal(t, "command", pinFields[0])
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", pinFields[1])
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", pinFields[2])
		assert.Equal(t, "pin", pinFields[3])
		assert.Equal(t, "U2hvdyBwb2Rz", pinFields[4])
		assert.Contains(t, pinFields[5], "pin")
		assert.Contains(t, pinFields[5], "kubectl get pods")
		assert.Contains(t, pinFields[5], "Show pods")

		historyFields := strings.SplitN(string(rows[1]), "\t", 6)
		require.Len(t, historyFields, 6)
		assert.Equal(t, "zsh-history", historyFields[0])
		assert.Equal(t, "10", historyFields[1])
		assert.Equal(t, "Z28gdGVzdCAuLy4uLg==", historyFields[2])
		assert.Equal(t, "history", historyFields[3])
		assert.Empty(t, historyFields[4])
		assert.Contains(t, historyFields[5], "history")
		assert.Contains(t, historyFields[5], "go test ./...")
	})

	t.Run("prints json picker rows", func(t *testing.T) {
		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, stdout, _ := newTestRunner(t, home)
		r.input = strings.NewReader("zsh-history\t10\tgo test ./...\x00")

		t.Setenv("RECALL_FZF_PIN_LABEL", "pin")
		t.Setenv("RECALL_FZF_HISTORY_LABEL", "history")
		t.Setenv("RECALL_FZF_COMMAND_WIDTH", "20")
		t.Setenv("RECALL_FZF_DESCRIPTION_WIDTH", "20")

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		err := (&PickerCmd{Format: "json"}).Run(r, newTestCLI(dataDir))

		require.NoError(t, err)

		var output []pickerOutput
		require.NoError(t, json.Unmarshal(stdout.Bytes(), &output))
		require.Len(t, output, 2)

		assert.Equal(t, "pin", output[0].Kind)
		assert.Equal(t, "kubectl get pods", output[0].Text)
		assert.Equal(t, "Show pods", output[0].Description)
		assert.Equal(t, "command", output[0].PayloadType)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", output[0].Payload)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", output[0].CommandPayload)
		assert.Equal(t, "U2hvdyBwb2Rz", output[0].DescriptionPayload)
		assert.Contains(t, output[0].Display, "kubectl get pods")

		assert.Equal(t, "history", output[1].Kind)
		assert.Equal(t, "go test ./...", output[1].Text)
		assert.Empty(t, output[1].Description)
		assert.Equal(t, "zsh-history", output[1].PayloadType)
		assert.Equal(t, "10", output[1].Payload)
		assert.Equal(t, "Z28gdGVzdCAuLy4uLg==", output[1].CommandPayload)
		assert.Empty(t, output[1].DescriptionPayload)
		assert.Contains(t, output[1].Display, "go test ./...")
	})

	t.Run("rejects unsupported format", func(t *testing.T) {
		t.Parallel()

		r, _, _ := newTestRunner(t, t.TempDir())

		err := (&PickerCmd{Format: "yaml"}).Run(r, newTestCLI(filepath.Join(t.TempDir(), "data")))

		require.Error(t, err)
		assert.ErrorContains(t, err, `unsupported picker format "yaml"`)
	})
}

func TestRunnerPickerCandidates(t *testing.T) {
	t.Parallel()

	t.Run("returns pins before history and dedupes", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, _, _ := newTestRunner(t, home)
		r.input = strings.NewReader(
			"zsh-history\t20\tkubectl get pods\x00" +
				"zsh-history\t19\tgo test ./...\x00",
		)

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		candidates, err := r.pickerCandidates(newTestCLIConfig(dataDir), 0)

		require.NoError(t, err)
		require.Len(t, candidates, 2)

		assert.Equal(t, "pin", candidates[0].Kind)
		assert.Equal(t, "kubectl get pods", candidates[0].Text)
		assert.Equal(t, "command", candidates[0].PayloadType)

		assert.Equal(t, "history", candidates[1].Kind)
		assert.Equal(t, "go test ./...", candidates[1].Text)
		assert.Equal(t, "zsh-history", candidates[1].PayloadType)
		assert.Equal(t, "19", candidates[1].Payload)
	})
}

func TestRunnerPickerPinCandidates(t *testing.T) {
	t.Parallel()

	t.Run("returns encoded pin candidates", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		dataDir := filepath.Join(home, "data")
		r, _, _ := newTestRunner(t, home)

		st := store.New(filepath.Join(dataDir, "store.json"))
		require.NoError(t, st.AddPin("go test ./...", "Run tests"))
		require.NoError(t, st.AddPin("kubectl get pods", "Show pods"))

		candidates, err := r.pickerPinCandidates(newTestCLIConfig(dataDir), 1)

		require.NoError(t, err)
		require.Len(t, candidates, 1)
		assert.Equal(t, "pin", candidates[0].Kind)
		assert.Equal(t, "kubectl get pods", candidates[0].Text)
		assert.Equal(t, "Show pods", candidates[0].Description)
		assert.Equal(t, "command", candidates[0].PayloadType)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", candidates[0].Payload)
	})
}

func TestReadPickerHistoryCandidates(t *testing.T) {
	t.Parallel()

	t.Run("reads nul separated history records", func(t *testing.T) {
		t.Parallel()

		input := strings.NewReader(
			"zsh-history\t42\tfind .\ncat file\x00" +
				"command\t\tgo test ./...\x00",
		)

		candidates, err := readPickerHistoryCandidates(input)

		require.NoError(t, err)
		require.Len(t, candidates, 2)

		assert.Equal(t, "find .\ncat file", candidates[0].Text)
		assert.Equal(t, "zsh-history", candidates[0].PayloadType)
		assert.Equal(t, "42", candidates[0].Payload)

		assert.Equal(t, "go test ./...", candidates[1].Text)
		assert.Equal(t, "command", candidates[1].PayloadType)
		assert.Equal(t, "Z28gdGVzdCAuLy4uLg==", candidates[1].Payload)
	})

	t.Run("returns empty candidates for empty input", func(t *testing.T) {
		t.Parallel()

		candidates, err := readPickerHistoryCandidates(strings.NewReader(""))

		require.NoError(t, err)
		assert.Empty(t, candidates)
	})
}

func TestSplitPickerInputRecords(t *testing.T) {
	t.Parallel()

	t.Run("splits nul separated records", func(t *testing.T) {
		t.Parallel()

		records := splitPickerInputRecords([]byte("one\x00two\x00"))

		assert.Equal(t, []string{"one", "two"}, records)
	})

	t.Run("splits newline separated records", func(t *testing.T) {
		t.Parallel()

		records := splitPickerInputRecords([]byte("one\ntwo\n"))

		assert.Equal(t, []string{"one", "two"}, records)
	})
}

func TestParsePickerHistoryRecord(t *testing.T) {
	t.Parallel()

	t.Run("parses zsh history record", func(t *testing.T) {
		t.Parallel()

		candidate, ok := parsePickerHistoryRecord("zsh-history\t123\tkubectl get pods")

		require.True(t, ok)
		assert.Equal(t, "history", candidate.Kind)
		assert.Equal(t, "kubectl get pods", candidate.Text)
		assert.Equal(t, "zsh-history", candidate.PayloadType)
		assert.Equal(t, "123", candidate.Payload)
	})

	t.Run("encodes command record with empty payload", func(t *testing.T) {
		t.Parallel()

		candidate, ok := parsePickerHistoryRecord("command\t\tkubectl get pods")

		require.True(t, ok)
		assert.Equal(t, "history", candidate.Kind)
		assert.Equal(t, "kubectl get pods", candidate.Text)
		assert.Equal(t, "command", candidate.PayloadType)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", candidate.Payload)
	})

	t.Run("parses plain command record", func(t *testing.T) {
		t.Parallel()

		candidate, ok := parsePickerHistoryRecord("kubectl get pods")

		require.True(t, ok)
		assert.Equal(t, "history", candidate.Kind)
		assert.Equal(t, "kubectl get pods", candidate.Text)
		assert.Equal(t, "command", candidate.PayloadType)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", candidate.Payload)
	})

	t.Run("rejects empty command", func(t *testing.T) {
		t.Parallel()

		candidate, ok := parsePickerHistoryRecord("zsh-history\t123\t   ")

		assert.False(t, ok)
		assert.Equal(t, pickerCandidate{}, candidate)
	})
}

func TestDedupePickerCandidates(t *testing.T) {
	t.Parallel()

	t.Run("preserves first duplicate", func(t *testing.T) {
		t.Parallel()

		candidates := []pickerCandidate{
			{Kind: "pin", Text: "kubectl get pods", Description: "Pinned"},
			{Kind: "history", Text: "KUBECTL GET PODS"},
			{Kind: "history", Text: "go test ./..."},
		}

		result := dedupePickerCandidates(candidates)

		require.Len(t, result, 2)
		assert.Equal(t, "pin", result[0].Kind)
		assert.Equal(t, "kubectl get pods", result[0].Text)
		assert.Equal(t, "go test ./...", result[1].Text)
	})
}

func TestDefaultPickerStyle(t *testing.T) {
	t.Run("reads style from environment", func(t *testing.T) {
		t.Setenv("RECALL_FZF_COMMAND_WIDTH", "12")
		t.Setenv("RECALL_FZF_DESCRIPTION_WIDTH", "8")
		t.Setenv("RECALL_FZF_PIN_LABEL", "PIN")
		t.Setenv("RECALL_FZF_HISTORY_LABEL", "HISTORY")
		t.Setenv("RECALL_FZF_DESCRIPTION_COLOR", "<d>")

		style := defaultPickerStyle()

		assert.Equal(t, 12, style.CommandWidth)
		assert.Equal(t, 8, style.DescriptionWidth)
		assert.Equal(t, "PIN", style.PinLabel)
		assert.Equal(t, "HISTORY", style.HistoryLabel)
		assert.Equal(t, "<d>", style.DescriptionColor)
		assert.Equal(t, "\033[0m", style.ResetColor)
	})
}

func TestFormatPickerRows(t *testing.T) {
	t.Parallel()

	t.Run("formats nul separated rows", func(t *testing.T) {
		t.Parallel()

		rows := formatPickerRows([]pickerCandidate{
			{
				Kind:        "pin",
				Text:        "kubectl get pods",
				Description: "Show running pods",
				PayloadType: "command",
				Payload:     "payload",
			},
		}, testPickerStyle())

		expected := []byte(
			"command\t" +
				"payload\t" +
				"a3ViZWN0bCBnZXQgcG9kcw==\t" +
				"pin\t" +
				"U2hvdyBydW5uaW5nIHBvZHM=\t" +
				"pin      kubec...  Show..." +
				"\x00",
		)

		assert.Equal(t, expected, rows)
	})
}

func TestFormatPickerJSON(t *testing.T) {
	t.Parallel()

	t.Run("formats pretty json rows", func(t *testing.T) {
		t.Parallel()

		content, err := formatPickerJSON([]pickerCandidate{
			{
				Kind:        "pin",
				Text:        "kubectl get pods",
				Description: "Show running pods",
				PayloadType: "command",
				Payload:     "payload",
			},
		}, testPickerStyle())

		require.NoError(t, err)
		assert.Contains(t, string(content), "\n  {")
		assert.Contains(t, string(content), `"kind": "pin"`)

		var output []pickerOutput
		require.NoError(t, json.Unmarshal(content, &output))
		require.Len(t, output, 1)
		assert.Equal(t, "pin", output[0].Kind)
		assert.Equal(t, "kubectl get pods", output[0].Text)
		assert.Equal(t, "Show running pods", output[0].Description)
		assert.Equal(t, "command", output[0].PayloadType)
		assert.Equal(t, "payload", output[0].Payload)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", output[0].CommandPayload)
		assert.Equal(t, "U2hvdyBydW5uaW5nIHBvZHM=", output[0].DescriptionPayload)
		assert.Equal(t, "pin      kubec...  Show...", output[0].Display)
	})
}

func TestNewPickerOutput(t *testing.T) {
	t.Parallel()

	t.Run("builds transport output", func(t *testing.T) {
		t.Parallel()

		output := newPickerOutput(pickerCandidate{
			Kind:        "pin",
			Text:        "kubectl get pods",
			Description: "Show running pods",
			PayloadType: "command",
			Payload:     "payload",
		}, testPickerStyle())

		assert.Equal(t, "pin", output.Kind)
		assert.Equal(t, "kubectl get pods", output.Text)
		assert.Equal(t, "Show running pods", output.Description)
		assert.Equal(t, "command", output.PayloadType)
		assert.Equal(t, "payload", output.Payload)
		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", output.CommandPayload)
		assert.Equal(t, "U2hvdyBydW5uaW5nIHBvZHM=", output.DescriptionPayload)
		assert.Equal(t, "pin      kubec...  Show...", output.Display)
	})
}

func TestFormatPickerDisplay(t *testing.T) {
	t.Parallel()

	t.Run("formats multiline command display", func(t *testing.T) {
		t.Parallel()

		display := formatPickerDisplay(pickerCandidate{
			Kind: "history",
			Text: "find .\nprintf test",
		}, testPickerStyle())

		assert.Contains(t, display, "history  find .")
		assert.Contains(t, display, "\n         print...")
	})
}

func TestPickerLabel(t *testing.T) {
	t.Parallel()

	t.Run("returns pin label", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "pin    ", pickerLabel("pin", testPickerStyle()))
	})

	t.Run("returns history label", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "history", pickerLabel("history", testPickerStyle()))
	})

	t.Run("returns padded unknown label", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "custom ", pickerLabel("custom", testPickerStyle()))
	})
}

func TestDisplayPickerCommand(t *testing.T) {
	t.Parallel()

	t.Run("preserves newlines and normalizes tabs", func(t *testing.T) {
		t.Parallel()

		value := displayPickerCommand("  echo\tone\necho two\r  ")

		assert.Equal(t, "echo one\necho two\\r", value)
	})
}

func TestDisplayPickerField(t *testing.T) {
	t.Parallel()

	t.Run("normalizes field separators", func(t *testing.T) {
		t.Parallel()

		value := displayPickerField("  one\ttwo\r\nthree  ")

		assert.Equal(t, "one two\\r\\nthree", value)
	})
}

func TestFitPickerField(t *testing.T) {
	t.Parallel()

	t.Run("pads short field", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "go  ", fitPickerField("go", 4))
	})

	t.Run("truncates long field", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "kubec...", fitPickerField("kubectl get pods", 8))
	})

	t.Run("keeps field when width is zero", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "kubectl", fitPickerField("kubectl", 0))
	})
}

func TestColorPickerDescription(t *testing.T) {
	t.Parallel()

	t.Run("colors non empty description", func(t *testing.T) {
		t.Parallel()

		style := testPickerStyle()
		style.DescriptionColor = "<d>"
		style.ResetColor = "</d>"

		assert.Equal(t, "<d>Show pods</d>", colorPickerDescription("Show pods", style))
	})

	t.Run("keeps empty description empty", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, colorPickerDescription("", testPickerStyle()))
	})
}

func TestEncodePickerPayload(t *testing.T) {
	t.Parallel()

	t.Run("encodes raw command", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "a3ViZWN0bCBnZXQgcG9kcw==", encodePickerPayload("kubectl get pods"))
	})
}

func TestEnvString(t *testing.T) {
	t.Run("returns environment value", func(t *testing.T) {
		t.Setenv("RECALL_TEST_VALUE", "configured")

		assert.Equal(t, "configured", envString("RECALL_TEST_VALUE", "fallback"))
	})

	t.Run("returns fallback for empty value", func(t *testing.T) {
		t.Setenv("RECALL_TEST_VALUE", "")

		assert.Equal(t, "fallback", envString("RECALL_TEST_VALUE", "fallback"))
	})
}

func TestEnvInt(t *testing.T) {
	t.Run("returns environment integer", func(t *testing.T) {
		t.Setenv("RECALL_TEST_INT", "42")

		assert.Equal(t, 42, envInt("RECALL_TEST_INT", 7))
	})

	t.Run("returns fallback for invalid integer", func(t *testing.T) {
		t.Setenv("RECALL_TEST_INT", "nope")

		assert.Equal(t, 7, envInt("RECALL_TEST_INT", 7))
	})

	t.Run("returns fallback for empty value", func(t *testing.T) {
		t.Setenv("RECALL_TEST_INT", "")

		assert.Equal(t, 7, envInt("RECALL_TEST_INT", 7))
	})
}

func testPickerStyle() pickerStyle {
	return pickerStyle{
		CommandWidth:     8,
		DescriptionWidth: 7,
		PinLabel:         "pin",
		HistoryLabel:     "history",
		DescriptionColor: "",
		ResetColor:       "",
	}
}

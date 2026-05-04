package logger

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("uses provided writer", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Print("hello")

		assert.Equal(t, "hello", buf.String())
	})

	t.Run("uses discard writer when nil", func(t *testing.T) {
		t.Parallel()

		log := New(nil)

		require.NotNil(t, log.Writer())
		log.Print("hello")
	})
}

func TestLoggerWriter(t *testing.T) {
	t.Parallel()

	t.Run("returns underlying writer", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)

		assert.Same(t, &buf, log.Writer())
	})
}

func TestLoggerWithVerbose(t *testing.T) {
	t.Parallel()

	t.Run("enables info output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithVerbose(true)
		log.Infof("hello %s", "world")

		assert.Equal(t, "info: hello world\n", buf.String())
	})

	t.Run("disables info output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithVerbose(false)
		log.Infof("hello %s", "world")

		assert.Empty(t, buf.String())
	})

	t.Run("does not mutate original logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		verboseLog := log.WithVerbose(true)

		log.Infof("hidden")
		require.Empty(t, buf.String())

		verboseLog.Infof("shown")
		assert.Equal(t, "info: shown\n", buf.String())
	})
}

func TestLoggerWithDebug(t *testing.T) {
	t.Parallel()

	t.Run("enables debug output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithDebug(true)
		log.Debugf("value=%d", 42)

		assert.Equal(t, "debug: value=42\n", buf.String())
	})

	t.Run("debug implies verbose output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithDebug(true)
		log.Infof("hello")

		assert.Equal(t, "info: hello\n", buf.String())
	})

	t.Run("disables debug output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithDebug(false)
		log.Debugf("hidden")

		assert.Empty(t, buf.String())
	})

	t.Run("does not mutate original logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		debugLog := log.WithDebug(true)

		log.Debugf("hidden")
		require.Empty(t, buf.String())

		debugLog.Debugf("shown")
		assert.Equal(t, "debug: shown\n", buf.String())
	})
}

func TestLoggerPrint(t *testing.T) {
	t.Parallel()

	t.Run("writes text without newline", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Print("hello")

		assert.Equal(t, "hello", buf.String())
	})
}

func TestLoggerPrintf(t *testing.T) {
	t.Parallel()

	t.Run("writes formatted text without newline", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Printf("hello %s %d", "world", 42)

		assert.Equal(t, "hello world 42", buf.String())
	})
}

func TestLoggerLine(t *testing.T) {
	t.Parallel()

	t.Run("writes text with newline", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Line("hello")

		assert.Equal(t, "hello\n", buf.String())
	})
}

func TestLoggerLinef(t *testing.T) {
	t.Parallel()

	t.Run("writes formatted text with newline", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Linef("hello %s %d", "world", 42)

		assert.Equal(t, "hello world 42\n", buf.String())
	})
}

func TestLoggerOutputf(t *testing.T) {
	t.Parallel()

	t.Run("writes formatted text with newline", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Outputf("hello %s", "world")

		assert.Equal(t, "hello world\n", buf.String())
	})
}

func TestLoggerWarnf(t *testing.T) {
	t.Parallel()

	t.Run("writes warning message", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Warnf("hello %s", "world")

		assert.Equal(t, "warning: hello world\n", buf.String())
	})
}

func TestLoggerInfof(t *testing.T) {
	t.Parallel()

	t.Run("does not write when verbose is disabled", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Infof("hello")

		assert.Empty(t, buf.String())
	})

	t.Run("writes info message when verbose is enabled", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithVerbose(true)
		log.Infof("hello %s", "world")

		assert.Equal(t, "info: hello world\n", buf.String())
	})
}

func TestLoggerDebugf(t *testing.T) {
	t.Parallel()

	t.Run("does not write when debug is disabled", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.Debugf("hello")

		assert.Empty(t, buf.String())
	})

	t.Run("writes debug message when debug is enabled", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf).WithDebug(true)
		log.Debugf("hello %s", "world")

		assert.Equal(t, "debug: hello world\n", buf.String())
	})
}

func TestLoggerPrintfHelper(t *testing.T) {
	t.Parallel()

	t.Run("writes prefixed formatted message with newline", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		log.printf("custom", "hello %s", "world")

		assert.Equal(t, "custom: hello world\n", buf.String())
	})
}

func TestLoggerWriterIntegration(t *testing.T) {
	t.Parallel()

	t.Run("returned writer can be written to directly", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log := New(&buf)
		_, err := io.WriteString(log.Writer(), "direct")

		require.NoError(t, err)
		assert.Equal(t, "direct", buf.String())
	})
}

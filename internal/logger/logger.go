package logger

import (
	"fmt"
	"io"
)

// Logger writes user-facing messages to an output stream.
type Logger struct {
	// writer receives formatted messages.
	writer io.Writer

	// verbose enables informational messages.
	verbose bool

	// debug enables debug messages and implies verbose output.
	debug bool
}

// New creates a logger backed by the given writer.
func New(writer io.Writer) Logger {
	if writer == nil {
		writer = io.Discard
	}

	return Logger{
		writer: writer,
	}
}

// Writer returns the underlying writer.
func (l Logger) Writer() io.Writer {
	return l.writer
}

// WithVerbose returns a logger configured to print informational messages.
func (l Logger) WithVerbose(verbose bool) Logger {
	l.verbose = verbose
	return l
}

// WithDebug returns a logger configured to print debug messages.
func (l Logger) WithDebug(debug bool) Logger {
	l.debug = debug
	if debug {
		l.verbose = true
	}

	return l
}

// Print writes text without adding a trailing newline.
func (l Logger) Print(text string) {
	_, _ = io.WriteString(l.writer, text)
}

// Printf writes formatted text without adding a trailing newline.
func (l Logger) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(l.writer, format, args...)
}

// Line writes text with a trailing newline.
func (l Logger) Line(text string) {
	_, _ = fmt.Fprintln(l.writer, text)
}

// Linef writes formatted text with a trailing newline.
func (l Logger) Linef(format string, args ...any) {
	_, _ = fmt.Fprintf(l.writer, format+"\n", args...)
}

// Outputf writes formatted text with a trailing newline.
func (l Logger) Outputf(format string, args ...any) {
	l.Linef(format, args...)
}

// Warnf prints a warning message.
func (l Logger) Warnf(format string, args ...any) {
	l.printf("warning", format, args...)
}

// Infof prints an informational message when verbose output is enabled.
func (l Logger) Infof(format string, args ...any) {
	if !l.verbose {
		return
	}

	l.printf("info", format, args...)
}

// Debugf prints a debug message when debug output is enabled.
func (l Logger) Debugf(format string, args ...any) {
	if !l.debug {
		return
	}

	l.printf("debug", format, args...)
}

// printf writes a prefixed message with a trailing newline.
func (l Logger) printf(prefix, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(l.writer, "%s: %s\n", prefix, message)
}

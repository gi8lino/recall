package runner

import (
	"path/filepath"
	"strings"
)

// expandHome replaces a leading ~/ with the current user's home directory.
func expandHome(path string, home string) string {
	if path == "~" {
		return home
	}

	if after, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(home, after)
	}

	return path
}

// commandTextFromArgs returns normalized command text from passthrough CLI args.
func commandTextFromArgs(args []string) string {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	return strings.TrimSpace(strings.Join(args, " "))
}

// exceedsLimit reports whether a slice of items exceeds a limit.
func exceedsLimit[T any](items []T, limit int) bool {
	return limit > 0 && len(items) > limit
}

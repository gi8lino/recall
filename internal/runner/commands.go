package runner

import (
	"strings"

	"github.com/containeroo/recall/internal/command"
)

// commandTextFromArgs returns normalized command text from passthrough CLI args.
func commandTextFromArgs(args []string) string {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	return command.Normalize(strings.Join(args, " "))
}

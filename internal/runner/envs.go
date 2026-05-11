package runner

import (
	"os"
	"strconv"
)

const (
	EnvCommandWidth     = "RECALL_FZF_COMMAND_WIDTH"
	EnvDescriptionWidth = "RECALL_FZF_DESCRIPTION_WIDTH"
	EnvPinLabel         = "RECALL_FZF_PIN_LABEL"
	EnvHistoryLabel     = "RECALL_FZF_HISTORY_LABEL"
	EnvDescriptionColor = "RECALL_FZF_DESCRIPTION_COLOR"
)

const (
	defaultCommandWidth     = 70
	defaultDescriptionWidth = 45
	defaultPinLabel         = "\033[38;5;208mpin\033[0m"
	defaultHistoryLabel     = "\033[38;5;245mhistory\033[0m"
	defaultDescriptionColor = "\033[38;5;110m"
	defaultResetColor       = "\033[0m"
)

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

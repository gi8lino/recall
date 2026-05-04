package shell

import (
	"fmt"
	"strings"
)

// InitRequest describes the shell integration to generate.
type InitRequest struct {
	// Shell is the target shell name.
	Shell string

	// Binary is the recall executable path to call from shell functions.
	Binary string

	// NoBind disables automatic Ctrl+R key binding generation.
	NoBind bool
}

// Init returns shell integration code for the requested shell.
func Init(req InitRequest) (string, error) {
	binary := req.Binary
	if binary == "" {
		binary = "recall"
	}

	switch strings.ToLower(req.Shell) {
	case "bash":
		return bashInit(binary, req.NoBind), nil
	case "zsh":
		return zshInit(binary, req.NoBind), nil
	default:
		return "", fmt.Errorf("unsupported shell %q; expected bash or zsh", req.Shell)
	}
}

// bashInit returns bash helper functions and optionally binds Ctrl+R.
func bashInit(binary string, noBind bool) string {
	bin := shellQuote(binary)
	binding := `bind -x '"\C-r": __recall_ctrl_r'
`

	if noBind {
		binding = `# To bind manually:
# bind -x '"\C-r": __recall_ctrl_r'
`
	}

	return fmt.Sprintf(`# recall shell integration for bash
__recall_history_for_fzf() {
  local limit
  limit="${RECALL_HISTORY_LIMIT:-5000}"

  if [[ "$limit" -gt 0 ]]; then
    history "$limit"
  else
    history
  fi | awk '
    {
      lines[NR] = $0
    }
    END {
      for (i = NR; i >= 1; i--) {
        line = lines[i]
        sub(/^ *[0-9]+[* ]*/, "", line)
        printf "command\t\t%%s%%c", line, 0
      }
    }
  '
}

__recall_ctrl_r() {
  local selected

  selected=$(__recall_history_for_fzf | RECALL_BIN=%[1]s recall-fzf "$READLINE_LINE") || return

  if [[ -n "$selected" ]]; then
    READLINE_LINE="$selected"
    READLINE_POINT=${#READLINE_LINE}
  fi
}

recall_pin_last() {
  local last
  last=$(history 1 | sed 's/^ *[0-9]\+[* ]*//')
  if [[ -n "$last" ]]; then
    %[1]s pin "$last"
  fi
}

%[2]s`, bin, binding)
}

// zshInit returns zsh helper functions and optionally binds Ctrl+R.
func zshInit(binary string, noBind bool) string {
	bin := shellQuote(binary)
	binding := `zle -N __recall_ctrl_r
bindkey '^R' __recall_ctrl_r
`

	if noBind {
		binding = `zle -N __recall_ctrl_r
# To bind manually:
# bindkey '^R' __recall_ctrl_r
`
	}

	return fmt.Sprintf(`# recall shell integration for zsh
__recall_base64_decode() {
  if printf '%%s' "$1" | base64 --decode 2>/dev/null; then
    return
  fi

  if printf '%%s' "$1" | base64 -d 2>/dev/null; then
    return
  fi

  printf '%%s' "$1" | base64 -D
}

__recall_history_for_fzf() {
  emulate -L zsh
  zmodload zsh/parameter 2>/dev/null || return

  local limit count event command
  local -a events

  limit="${RECALL_HISTORY_LIMIT:-5000}"
  count=0

  events=("${(@k)history}")
  events=("${(@On)events}")

  for event in "${events[@]}"; do
    command="${history[$event]}"
    [[ -n "$command" ]] || continue

    printf 'zsh-history\t%%s\t%%s\0' "$event" "$command"

    (( count++ ))
    if (( limit > 0 && count >= limit )); then
      break
    fi
  done
}

__recall_ctrl_r() {
  local selected payload_type payload command

  selected=$(__recall_history_for_fzf | RECALL_BIN=%[1]s RECALL_FZF_RETURN=record recall-fzf "$BUFFER") || {
    zle reset-prompt
    zle redisplay
    return
  }

  IFS=$'\t' read -r payload_type payload <<< "$selected"

  case "$payload_type" in
    zsh-history)
      command="${history[$payload]}"
      ;;
    command)
      command="$(__recall_base64_decode "$payload")"
      ;;
    *)
      command=""
      ;;
  esac

  if [[ -n "$command" ]]; then
    BUFFER="$command"
    CURSOR=${#BUFFER}
  fi

  zle reset-prompt
  zle redisplay
}

recall_pin_last() {
  local last
  last=$(fc -ln -1)
  if [[ -n "$last" ]]; then
    %[1]s pin "$last"
  fi
}

%[2]s`, bin, binding)
}

// shellQuote returns a POSIX-safe single-quoted shell string.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

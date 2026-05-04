# Recall

Recall is a small terminal command palette powered by shell history and pinned commands.

It lets you search history, pin useful commands, add descriptions, and use a fast `Ctrl+R` picker backed by `fzf`.

## Features

- Search Bash and Zsh history files
- Use in-memory shell history for interactive `Ctrl+R`
- Pin commands in Recall's own store
- Add, update, list, and remove pins
- Pin history entries directly from the picker
- Edit pin descriptions from the picker
- Show pins before history in the picker
- Preserve multiline Zsh history commands
- Use `fzf-tmux` inside tmux
- Keep normal output on `stdout`
- Keep diagnostics on `stderr`

## Installation

```sh
git clone https://github.com/containeroo/recall.git
cd recall
go build -o recall ./cmd
```

```sh
mkdir -p ~/.local/bin
mv recall ~/.local/bin/recall
chmod +x scripts/recall-fzf
cp scripts/recall-fzf ~/.local/bin/recall-fzf
```

Make sure `~/.local/bin` is on your `PATH`.

## Requirements

The interactive picker needs:

- `fzf`
- optionally `fzf-tmux`

## Usage

```sh
recall <command> [flags]
```

```sh
recall --help
recall search --help
recall picker --help
recall pin --help
recall describe --help
recall remove --help
recall list --help
recall init --help
```

## Search

Search pins and history files:

```sh
recall search docker
recall search kubectl pods
```

Search only pins:

```sh
recall search --pins-only kubectl
```

Search only history files:

```sh
recall search --history-only kubectl
```

Limit results:

```sh
recall --limit 5 search docker
```

Search preserves priority order:

```text
pins first
recent history after pins
duplicates removed
```

### Search fzf format

`recall search --format fzf` is a low-level tab format for scripts.

```sh
recall search --format fzf
```

It prints:

```text
kind<TAB>command<TAB>description<TAB>payload_type<TAB>payload
```

For the interactive picker, use `recall picker --format fzf` through `recall-fzf`.

## Picker

`recall picker --format fzf` produces NUL-separated rows for `fzf`.

It reads shell-provided history from `stdin`, loads pins, dedupes candidates, formats multiline display rows, and prints:

```text
payload_type<TAB>payload<TAB>command_payload<TAB>kind<TAB>description_payload<TAB>display<NUL>
```

`recall-fzf` owns the actual `fzf` UI.

Default picker keys:

```text
Enter   insert selected command
Ctrl-P  pin selected command
Ctrl-O  edit selected pin description
Ctrl-X  remove selected pin
```

Editing a description opens a small prompt. In tmux, Recall uses a popup window. If the pin already has a description, the existing text is prefilled.

## Pin commands

```sh
recall pin go test ./...
```

With description:

```sh
recall pin -d "Run all Go tests" -- go test ./...
```

Use `--` when the pinned command contains flags:

```sh
recall pin -d "Show pods" -- kubectl get pods -n default
```

The `--` separator is not stored.

## Describe pins

```sh
recall describe -d "Show running Kubernetes pods" -- kubectl get pods
```

Descriptions apply to pins only.

## Remove pins

```sh
recall remove -- kubectl get pods
```

## List pins

```sh
recall list
```

Pins are printed newest first.

Pins with descriptions are printed as:

```text
description<TAB>command
```

## Shell integration

Zsh:

```sh
eval "$(recall init zsh)"
```

Bash:

```sh
eval "$(recall init bash)"
```

Without binding `Ctrl+R`:

```sh
eval "$(recall init zsh --no-bind)"
bindkey '^R' __recall_ctrl_r
```

```sh
eval "$(recall init bash --no-bind)"
bind -x '"\C-r": __recall_ctrl_r'
```

## Zsh setup

```sh
source <(fzf --zsh)

export PATH="$PATH:$HOME/.local/bin"
export RECALL_HISTORY_LIMIT=5000

eval "$(recall init zsh)"
```

Zsh integration passes history event IDs to the picker. After selection, Recall restores history commands from Zsh's `$history` array, so multiline commands stay intact.

## Bash setup

```sh
export PATH="$PATH:$HOME/.local/bin"
export RECALL_HISTORY_LIMIT=5000

eval "$(recall init bash)"
```

Bash multiline history support is best-effort.

## Global flags

### `--data-dir`

```sh
recall --data-dir ~/.local/share/recall search docker
```

Default:

```text
$XDG_DATA_HOME/recall
```

Fallback:

```text
~/.local/share/recall
```

### `--history`

```sh
recall --history ~/.zsh_history search docker
recall --history ~/.zsh_history --history ~/.bash_history search docker
```

Defaults:

```text
$HISTFILE
~/.zsh_history
~/.bash_history
```

### `--history-limit`

Limits file-history search:

```sh
recall --history-limit 20000 search docker
```

Use `0` for all file history:

```sh
recall --history-limit 0 search docker
```

Interactive `Ctrl+R` uses in-memory shell history instead.

### `--limit`, `-n`

```sh
recall --limit 10 search git
recall -n 10 search git
```

Use `0` for no result cap.

## Environment variables

### `RECALL_BIN`

Path to the Recall binary used by `recall-fzf`.

```sh
export RECALL_BIN="$HOME/.local/bin/recall"
```

### `RECALL_FZF_LIMIT`

Maximum pins loaded by `recall-fzf`.

```sh
export RECALL_FZF_LIMIT=0
```

Use `0` for no pin cap.

### `RECALL_HISTORY_LIMIT`

Maximum in-memory history entries sent to the picker.

```sh
export RECALL_HISTORY_LIMIT=5000
```

### `RECALL_FZF_TMUX_SIZE`

Size used by `fzf-tmux`.

```sh
export RECALL_FZF_TMUX_SIZE="80%,70%"
```

### `RECALL_FZF_HEADER`

Custom picker header.

```sh
export RECALL_FZF_HEADER="source   command   description"
```

### `RECALL_FZF_KEY_HELP`

Custom picker footer/help text.

```sh
export RECALL_FZF_KEY_HELP="ctrl-p: pin | ctrl-o: edit description | ctrl-x: remove pin | enter: insert"
```

### `RECALL_FZF_PIN_KEY`

Key used to pin the selected command.

```sh
export RECALL_FZF_PIN_KEY="ctrl-p"
```

### `RECALL_FZF_DESCRIBE_KEY`

Key used to edit the selected pin description.

```sh
export RECALL_FZF_DESCRIBE_KEY="ctrl-o"
```

### `RECALL_FZF_REMOVE_KEY`

Key used to remove the selected pin.

```sh
export RECALL_FZF_REMOVE_KEY="ctrl-x"
```

### `RECALL_FZF_PIN_LABEL`

Display label for pinned commands.

```sh
export RECALL_FZF_PIN_LABEL=$'\033[38;5;208mpin\033[0m'
```

### `RECALL_FZF_HISTORY_LABEL`

Display label for history commands.

```sh
export RECALL_FZF_HISTORY_LABEL=$'\033[38;5;245mhistory\033[0m'
```

### `RECALL_FZF_DESCRIPTION_COLOR`

ANSI color prefix for the description column.

```sh
export RECALL_FZF_DESCRIPTION_COLOR=$'\033[38;5;110m'
```

### `RECALL_FZF_SOURCE_WIDTH`

Width of the source column.

```sh
export RECALL_FZF_SOURCE_WIDTH=7
```

### `RECALL_FZF_COMMAND_WIDTH`

Width of the command column.

```sh
export RECALL_FZF_COMMAND_WIDTH=70
```

### `RECALL_FZF_DESCRIPTION_WIDTH`

Width of the description column.

```sh
export RECALL_FZF_DESCRIPTION_WIDTH=45
```

### `RECALL_FZF_DESCRIPTION_POPUP_WIDTH`

Width of the tmux description popup.

```sh
export RECALL_FZF_DESCRIPTION_POPUP_WIDTH="75%"
```

### `RECALL_FZF_DESCRIPTION_POPUP_HEIGHT`

Height of the tmux description popup.

```sh
export RECALL_FZF_DESCRIPTION_POPUP_HEIGHT="40%"
```

## Data storage

Recall stores pins in:

```text
~/.local/share/recall/store.json
```

Example:

```json
{
  "pins": [
    {
      "command": "go test ./...",
      "description": "Run all Go tests"
    }
  ]
}
```

Recall never writes to `.bash_history` or `.zsh_history`.

## Design

```text
recall search
  non-interactive search over pins and history files

recall picker --format fzf
  produces fzf-ready rows from pins and shell-provided history

recall-fzf
  owns fzf/fzf-tmux UI options and picker key actions

shell integration
  sends in-memory history and inserts the selected command
```

## Development

Run tests:

```sh
go test ./...
bash scripts/recall-fzf_test.sh
```

Build:

```sh
go build -o recall ./cmd
```

Format:

```sh
go fmt ./...
```

Install locally:

```sh
go build -o recall ./cmd
mv recall ~/.local/bin/recall
chmod +x scripts/recall-fzf
cp scripts/recall-fzf ~/.local/bin/recall-fzf
```

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

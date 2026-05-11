package runner

import (
	"errors"

	"github.com/containeroo/recall/internal/store"
)

// PinOptions configures one pin invocation.
type PinOptions struct {
	Description string   // Description explains why the command is useful.
	Command     []string // Command contains the command text to pin.
}

// RunPin stores a command in recall's own pinned-command store.
func RunPin(r Runner, cfg Config, opts PinOptions) error {
	if len(opts.Command) == 0 {
		return errors.New("pin requires command text")
	}

	text := commandTextFromArgs(opts.Command)
	if text == "" {
		return errors.New("pin requires non-empty command text")
	}

	st := store.NewStore(r.storePath(cfg))
	if err := st.AddPin(text, opts.Description); err != nil {
		return err
	}

	r.output.Line(text)
	return nil
}

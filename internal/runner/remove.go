package runner

import (
	"errors"

	"github.com/containeroo/recall/internal/store"
)

// RemoveOptions configures one remove invocation.
type RemoveOptions struct {
	Command []string // Command contains the pinned command text to remove.
}

// RunRemove removes an existing pinned command.
func RunRemove(r Runner, cfg Config, opts RemoveOptions) error {
	if len(opts.Command) == 0 {
		return errors.New("remove requires command text")
	}

	text := commandTextFromArgs(opts.Command)
	if text == "" {
		return errors.New("remove requires non-empty command text")
	}

	st := store.NewStore(r.storePath(cfg))
	if err := st.RemovePin(text); err != nil {
		return err
	}

	r.output.Line(text)
	return nil
}

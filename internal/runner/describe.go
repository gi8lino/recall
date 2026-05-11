package runner

import (
	"errors"
	"strings"

	"github.com/containeroo/recall/internal/store"
)

// DescribeOptions configures one describe invocation.
type DescribeOptions struct {
	Description string   // Description explains why the pinned command is useful.
	Command     []string // Command contains the pinned command text to update.
}

// RunDescribe updates the description of an existing pinned command.
func RunDescribe(r Runner, cfg Config, opts DescribeOptions) error {
	if strings.TrimSpace(opts.Description) == "" {
		return errors.New("describe requires a non-empty description")
	}

	if len(opts.Command) == 0 {
		return errors.New("describe requires command text")
	}

	text := commandTextFromArgs(opts.Command)
	if text == "" {
		return errors.New("describe requires non-empty command text")
	}

	st := store.NewStore(r.storePath(cfg))
	if err := st.DescribePin(text, opts.Description); err != nil {
		return err
	}

	r.output.Line(text)
	return nil
}

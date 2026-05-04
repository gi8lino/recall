package runner

import (
	"fmt"

	"github.com/containeroo/recall/internal/store"
)

// PinCmd contains options for the pin command.
type PinCmd struct {
	// Description explains why the command is useful.
	Description string `short:"d" name:"description" help:"Optional description for the pinned command."`

	// Command contains the command text to pin.
	Command []string `arg:"" name:"command" passthrough:"" help:"Command text to pin."`
}

// Run stores a command in recall's own pinned-command store.
func (c *PinCmd) Run(r Runner, cli *CLI) error {
	if len(c.Command) == 0 {
		return fmt.Errorf("pin requires command text")
	}

	cfg := cli.Config(r.home)
	text := commandTextFromArgs(c.Command)
	if text == "" {
		return fmt.Errorf("pin requires non-empty command text")
	}

	st := store.New(r.storePath(cfg))
	if err := st.AddPin(text, c.Description); err != nil {
		return err
	}

	r.output.Line(text)
	return nil
}

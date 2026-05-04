package runner

import (
	"fmt"

	"github.com/containeroo/recall/internal/store"
)

// RemoveCmd contains options for the remove command.
type RemoveCmd struct {
	// Command contains the pinned command text to remove.
	Command []string `arg:"" name:"command" passthrough:"" help:"Pinned command text to remove."`
}

// Run removes an existing pinned command.
func (c *RemoveCmd) Run(r Runner, cli *CLI) error {
	if len(c.Command) == 0 {
		return fmt.Errorf("remove requires command text")
	}

	cfg := cli.Config(r.home)
	text := commandTextFromArgs(c.Command)
	if text == "" {
		return fmt.Errorf("remove requires non-empty command text")
	}

	st := store.New(r.storePath(cfg))
	if err := st.RemovePin(text); err != nil {
		return err
	}

	r.output.Line(text)
	return nil
}

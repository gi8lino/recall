package runner

import (
	"fmt"
	"strings"

	"github.com/containeroo/recall/internal/store"
)

// DescribeCmd contains options for the describe command.
type DescribeCmd struct {
	// Description explains why the pinned command is useful.
	Description string `short:"d" name:"description" help:"Description for the pinned command."`

	// Command contains the pinned command text to update.
	Command []string `arg:"" name:"command" passthrough:"" help:"Pinned command text to describe."`
}

// Run updates the description of an existing pinned command.
func (c *DescribeCmd) Run(r Runner, cli *CLI) error {
	if strings.TrimSpace(c.Description) == "" {
		return fmt.Errorf("describe requires a non-empty description")
	}

	if len(c.Command) == 0 {
		return fmt.Errorf("describe requires command text")
	}

	cfg := cli.Config(r.home)
	text := commandTextFromArgs(c.Command)
	if text == "" {
		return fmt.Errorf("describe requires non-empty command text")
	}

	st := store.New(r.storePath(cfg))
	if err := st.DescribePin(text, c.Description); err != nil {
		return err
	}

	r.output.Line(text)
	return nil
}

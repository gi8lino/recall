package runner

import "github.com/containeroo/recall/internal/store"

// ListCmd contains options for the list command.
type ListCmd struct{}

// Run prints all pinned commands in priority order.
func (c *ListCmd) Run(r Runner, cli *CLI) error {
	cfg := cli.Config(r.home)
	st := store.New(r.storePath(cfg))

	pins, err := st.Pins()
	if err != nil {
		return err
	}

	for _, pin := range pins {
		if pin.Description == "" {
			r.output.Line(pin.Text)
			continue
		}

		r.output.Linef("%s\t%s", pin.Description, pin.Text)
	}

	return nil
}

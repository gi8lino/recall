package runner

import "github.com/containeroo/recall/internal/store"

// RunList prints all pinned commands in priority order.
func RunList(r Runner, cfg Config) error {
	st := store.NewStore(r.storePath(cfg))

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

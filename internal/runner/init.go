package runner

import "github.com/containeroo/recall/internal/shell"

// InitCmd contains options for the init command.
type InitCmd struct {
	// NoBind prints shell integration without binding Ctrl+R.
	NoBind bool `name:"no-bind" help:"Print shell integration without binding Ctrl+R."`

	// Shell is the shell integration to generate.
	Shell string `arg:"" enum:"bash,zsh" help:"Shell to initialize."`
}

// Run prints shell integration for bash or zsh.
func (c *InitCmd) Run(r Runner) error {
	script, err := shell.Init(shell.InitRequest{
		Shell:  c.Shell,
		Binary: r.binary,
		NoBind: c.NoBind,
	})
	if err != nil {
		return err
	}

	r.output.Print(script)
	return nil
}

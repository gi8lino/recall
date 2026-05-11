package runner

import "github.com/containeroo/recall/internal/shell"

// InitOptions configures one init invocation.
type InitOptions struct {
	NoBind bool   // NoBind prints shell integration without binding Ctrl+R.
	Shell  string // Shell is the shell integration to generate.
}

// RunInit prints shell integration for bash or zsh.
func RunInit(r Runner, opts InitOptions) error {
	script, err := shell.Init(shell.InitRequest{
		Shell:  opts.Shell,
		Binary: r.binary,
		NoBind: opts.NoBind,
	})
	if err != nil {
		return err
	}

	r.output.Print(script)
	return nil
}

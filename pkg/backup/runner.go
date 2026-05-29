package backup

import (
	"context"
	"io"
	"os/exec"
)

type Command struct {
	Name   string
	Args   []string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
}

type Runner interface {
	Run(ctx context.Context, cmd Command) error
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, cmd Command) error {
	// #nosec G204 -- command names and args are built from trusted config.
	c := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	c.Env = append(c.Environ(), cmd.Env...)
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	return c.Run()
}

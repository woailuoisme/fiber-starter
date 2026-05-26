package prompts

import (
	"io"
	"os"

	"lfiber/internal/console/ui"

	"github.com/charmbracelet/huh"
)

type DestructiveOptions struct {
	Force         bool
	NoInteraction bool
	Message       string
}

func ConfirmDestructive(in io.Reader, out io.Writer, opts DestructiveOptions) (bool, error) {
	if opts.Force {
		return true, nil
	}

	isCI := os.Getenv("CI") != ""
	isNonInteractive := opts.NoInteraction || isCI
	isPipeOutput := !ui.IsTerminal(in) || !ui.IsTerminal(out)

	if isNonInteractive || isPipeOutput {
		return false, nil
	}

	confirmed := false
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirm destructive operation").
				Description(opts.Message).
				Affirmative("Yes, continue").
				Negative("No, cancel").
				Value(&confirmed),
		),
	).WithInput(in).WithOutput(out).Run()
	if err != nil {
		return false, err
	}
	return confirmed, nil
}

func Cancelled(out io.Writer) {
	ui.Warning(out, "Operation cancelled")
}

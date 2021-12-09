package internal

import (
	"os/exec"

	"github.com/docker/cli/cli"
)

func ExitToStatus(err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		err = cli.StatusError{
			StatusCode: exitErr.ExitCode(),
		}
	}
	return err
}

package system

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"unicode"

	"github.com/docker/cli/cli"
)

type Cmd struct {
	arg    []string
	env    []string
	name   string
	stderr io.Writer
	stdin  io.Reader
	stdout io.Writer
}

func Command(name string) *Cmd {
	return &Cmd{
		name: name,
	}
}

func (cmd *Cmd) debug(args []string) {
	for index, arg := range args {
		for _, r := range arg {
			if unicode.IsSpace(r) {
				args[index] = fmt.Sprintf("'%s'", arg)

				break
			}
		}
	}

	fmt.Println("$", strings.Join(args, " "))
}

func (cmd *Cmd) Arg(arg ...string) {
	cmd.arg = append(cmd.arg, arg...)
}

func (cmd *Cmd) Env(env ...string) {
	cmd.env = append(cmd.env, env...)
}

func (cmd *Cmd) Err(debug bool) (string, error) {
	var stderr bytes.Buffer

	command := exec.Command(cmd.name, cmd.arg...)

	command.Env = cmd.env

	command.Stdin = cmd.stdin
	command.Stdout = cmd.stdout
	if cmd.stderr != nil {
		command.Stderr = io.MultiWriter(cmd.stderr, &stderr)
	} else {
		command.Stderr = &stderr
	}

	if debug {
		cmd.debug(command.Args)
	}

	err := command.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = cli.StatusError{
				StatusCode: exitErr.ExitCode(),
			}
		}
	}

	return stderr.String(), err
}

func (cmd *Cmd) Flag(key string, value interface{}) {
	if !reflect.ValueOf(value).IsZero() {
		switch len(key) {
		case 1:
			key = fmt.Sprintf("-%s", key)
		default:
			key = fmt.Sprintf("--%s", key)
		}

		switch reflect.TypeOf(value).Kind() {
		case reflect.Bool:
			if reflect.ValueOf(value).Bool() {
				cmd.arg = append(cmd.arg, key)
			}
		case reflect.Slice:
			slice := reflect.ValueOf(value)
			for index := 0; index < slice.Len(); index++ {
				cmd.arg = append(cmd.arg, key, fmt.Sprintf("%v", slice.Index(index)))
			}
		default:
			cmd.arg = append(cmd.arg, key, fmt.Sprintf("%v", value))
		}
	}
}

func (cmd *Cmd) Out(debug bool) (string, error) {
	var stdout bytes.Buffer

	command := exec.Command(cmd.name, cmd.arg...)

	command.Env = cmd.env

	command.Stdin = cmd.stdin
	if cmd.stdout != nil {
		command.Stdout = io.MultiWriter(cmd.stdout, &stdout)
	} else {
		command.Stdout = &stdout
	}
	command.Stderr = cmd.stderr

	if debug {
		cmd.debug(command.Args)
	}

	err := command.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = cli.StatusError{
				StatusCode: exitErr.ExitCode(),
			}
		}
	}

	return stdout.String(), err
}

func (cmd *Cmd) Run(debug bool) error {
	command := exec.Command(cmd.name, cmd.arg...)

	command.Env = cmd.env

	command.Stdin = cmd.stdin
	command.Stdout = cmd.stdout
	command.Stderr = cmd.stderr

	if debug {
		cmd.debug(command.Args)
	}

	err := command.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = cli.StatusError{
				StatusCode: exitErr.ExitCode(),
			}
		}
	}

	return err
}

func (cmd *Cmd) Std(in io.Reader, out, err io.Writer) {
	cmd.stdin = in
	cmd.stdout = out
	cmd.stderr = err
}

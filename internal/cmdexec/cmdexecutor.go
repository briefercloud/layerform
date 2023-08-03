package cmdexec

import (
	"io"
	"os/exec"
)

type CommandExecutor interface {
	SetStdin(stdin io.Reader)
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
	SetDir(dir string)
	Run(name string, args ...string) (int, error)
}

type OSCommandExecutor struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	dir    string
}

var _ CommandExecutor = &OSCommandExecutor{}

func (c *OSCommandExecutor) SetStdin(stdin io.Reader) {
	c.Stdin = stdin
}

func (c *OSCommandExecutor) SetStdout(stdout io.Writer) {
	c.Stdout = stdout
}

func (c *OSCommandExecutor) SetStderr(stderr io.Writer) {
	c.Stderr = stderr
}

func (c *OSCommandExecutor) SetDir(dir string) {
	c.dir = dir
}

func (c *OSCommandExecutor) Run(name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)

	cmd.Dir = c.dir
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr

	err := cmd.Run()

	return cmd.ProcessState.ExitCode(), err
}

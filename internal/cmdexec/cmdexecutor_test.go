package cmdexec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOSCommandExecutor_Works(t *testing.T) {
	stdin := bytes.NewReader([]byte("hello\nworld"))
	stdout := &bytes.Buffer{}

	executor := &OSCommandExecutor{Stdin: stdin, Stdout: stdout}

	n, err := executor.Run("grep", "hello")
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	assert.Equal(t, "hello\n", stdout.String())
}

func TestOSCommandExecutor_Run(t *testing.T) {
	executor := &OSCommandExecutor{}

	tests := []struct {
		name     string
		args     []string
		exitCode int
		err      bool
	}{
		{
			name:     "Successful execution",
			args:     []string{"ls", "-l"},
			exitCode: 0,
			err:      false,
		},
		{
			name:     "Failed execution",
			args:     []string{"false"},
			exitCode: 1,
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exitCode, err := executor.Run(test.args[0], test.args[1:]...)

			assert.Equal(t, test.exitCode, exitCode)
			if test.err {
				assert.Error(t, err)
			}
		})
	}
}

func TestOSCommandExecutor_SetDir(t *testing.T) {
	executor := &OSCommandExecutor{}
	dir := "/path/to/directory"

	executor.SetDir(dir)

	assert.Equal(t, dir, executor.dir)
}

func TestOSCommandExecutor_SetStdin(t *testing.T) {
	executor := &OSCommandExecutor{}
	stdin := bytes.NewReader([]byte("test input"))

	executor.SetStdin(stdin)

	assert.Equal(t, stdin, executor.Stdin)
}

func TestOSCommandExecutor_SetStdout(t *testing.T) {
	executor := &OSCommandExecutor{}
	stdout := &bytes.Buffer{}

	executor.SetStdout(stdout)

	assert.Equal(t, stdout, executor.Stdout)
}

func TestOSCommandExecutor_SetStderr(t *testing.T) {
	executor := &OSCommandExecutor{}
	stderr := &bytes.Buffer{}

	executor.SetStderr(stderr)

	assert.Equal(t, stderr, executor.Stderr)
}

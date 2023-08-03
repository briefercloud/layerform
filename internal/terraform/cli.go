package terraform

import (
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/cmdexec"
)

type terraformCLI struct {
	exec cmdexec.CommandExecutor
}

var _ Client = &terraformCLI{}

func NewCLI(exec cmdexec.CommandExecutor) *terraformCLI {
	return &terraformCLI{exec}
}

func (t *terraformCLI) Init(dir string) error {
	t.exec.SetDir(dir)
	_, err := t.exec.Run("terraform", "init")
	return err
}

func (t *terraformCLI) Apply(dir string, state *State) (*State, error) {
	if state != nil {
		err := os.WriteFile(path.Join(dir, "terraform.tfstate"), state.Bytes, 0644)
		if err != nil {
			return nil, err
		}
	}

	t.exec.SetDir(dir)
	_, err := t.exec.Run("terraform", "apply")
	if err != nil {
		return nil, err
	}

	nextState, err := TFStateFromFile(path.Join(dir, "terraform.tfstate"))

	return nextState, errors.Wrap(err, "fail to parse state out of teraform.tfstate")
}

func (t *terraformCLI) Destroy(dir string, state *State, target ...string) (*State, error) {
	if state != nil {
		err := os.WriteFile(path.Join(dir, "terraform.tfstate"), state.Bytes, 0644)
		if err != nil {
			return nil, err
		}
	}

	t.exec.SetDir(dir)
	args := []string{"destroy"}
	if len(target) > 0 {
		for _, t := range target {
			args = append(args, "-target", t)
		}
	}

	_, err := t.exec.Run("terraform", args...)
	if err != nil {
		return nil, err
	}

	nextState, err := TFStateFromFile(path.Join(dir, "terraform.tfstate"))

	return nextState, errors.Wrap(err, "fail to parse state out of teraform.tfstate")
}

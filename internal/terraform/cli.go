package terraform

import (
	"os"
	"path"
)

type terraformCLI struct {
	exec CommandExecutor
}

var _ Client = &terraformCLI{}

func NewCLI(exec CommandExecutor) *terraformCLI {
	return &terraformCLI{exec}
}

func (t *terraformCLI) Init(dir string) error {
	t.exec.SetDir(dir)
	_, err := t.exec.Run("terraform", "init")
	return err
}

func (t *terraformCLI) Apply(dir string, state []byte) ([]byte, error) {
	if state != nil {
		err := os.WriteFile(path.Join(dir, "terraform.tfstate"), state, 0644)
		if err != nil {
			return nil, err
		}
	}

	t.exec.SetDir(dir)
	_, err := t.exec.Run("terraform", "apply")
	if err != nil {
		return nil, err
	}

	state, err = os.ReadFile(path.Join(dir, "terraform.tfstate"))
	if err != nil {
		return nil, err
	}

	return state, err
}

func (t *terraformCLI) Destroy(dir string, state []byte) ([]byte, error) {
	if state != nil {
		err := os.WriteFile(path.Join(dir, "terraform.tfstate"), state, 0644)
		if err != nil {
			return nil, err
		}
	}

	t.exec.SetDir(dir)
	_, err := t.exec.Run("terraform", "destroy")
	if err != nil {
		return nil, err
	}

	state, err = os.ReadFile(path.Join(dir, "terraform.tfstate"))
	if err != nil {
		return nil, err
	}

	return state, err
}

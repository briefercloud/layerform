package tfclient

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func New(dir string, tfpath string) (*tfexec.Terraform, error) {
	tf, err := tfexec.NewTerraform(dir, tfpath)
	if err != nil {
		return nil, err
	}

	logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
	if logLevel != hclog.NoLevel && logLevel != hclog.Off {
		tf.SetStdout(os.Stdout)
		tf.SetStderr(os.Stderr)
	}

	return tf, nil
}

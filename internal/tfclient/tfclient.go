package tfclient

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

type client struct {
	tf *tfexec.Terraform
}

func New(dir string, tfpath string) (*client, error) {
	tf, err := tfexec.NewTerraform(dir, tfpath)
	if err != nil {
		return nil, err
	}

	logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
	if logLevel != hclog.NoLevel && logLevel != hclog.Off {
		tf.SetStdout(os.Stdout)
		tf.SetStderr(os.Stderr)
	}

	return &client{tf}, nil
}

func (c *client) ShowStateFile(ctx context.Context, statePath string) (*tfjson.State, error) {
	hclog.FromContext(ctx).Debug("Running terraform show")

	return c.tf.ShowStateFile(ctx, statePath)
}

func (c *client) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	hclog.FromContext(ctx).Debug("Running terraform destroy")

	return c.tf.Destroy(ctx, opts...)
}

func (c *client) Init(ctx context.Context) error {
	hclog.FromContext(ctx).Debug("Running terraform init")

	return c.tf.Init(ctx)
}

func (c *client) Output(ctx context.Context) (map[string]tfexec.OutputMeta, error) {
	hclog.FromContext(ctx).Debug("Running terraform output")

	return c.tf.Output(ctx)
}

func (c *client) StateMv(ctx context.Context, source, destination string, opts ...tfexec.StateMvCmdOption) error {
	hclog.FromContext(ctx).Debug("Running terraform state mv")

	return c.tf.StateMv(ctx, source, destination, opts...)
}

func (c *client) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	hclog.FromContext(ctx).Debug("Running terraform apply")

	return c.tf.Apply(ctx, opts...)
}

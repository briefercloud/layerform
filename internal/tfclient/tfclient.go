package tfclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
)

type client struct {
	dir string
	tf  *tfexec.Terraform
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

	return &client{dir, tf}, nil
}

func (c *client) ShowStateFile(ctx context.Context, statePath string) (*tfjson.State, error) {
	hclog.FromContext(ctx).Debug("Running terraform show")

	return c.tf.ShowStateFile(ctx, statePath)
}

func (c *client) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	hclog.FromContext(ctx).Debug("Running terraform destroy")

	return c.tf.Destroy(ctx, opts...)
}

func (c *client) Init(ctx context.Context, cacheKey []byte) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Running terraform init")

	if cacheKey == nil {
		return c.tf.Init(ctx)
	}

	hexCacheKey := fmt.Sprintf("%x", cacheKey)

	homedir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "fail to get user home dir")
	}

	cacheBaseFolder := path.Join(homedir, ".layerform", ".cache", "init", hexCacheKey)

	cacheTerraformFolder := path.Join(cacheBaseFolder, ".terraform")
	localTerraformFolder := path.Join(c.dir, ".terraform")

	cacheLockFile := path.Join(cacheBaseFolder, ".terraform.lock.hcl")
	localLockFile := path.Join(c.dir, ".terraform.lock.hcl")
	_, err = os.Stat(cacheTerraformFolder)
	if err == nil {
		logger.Debug("Restoring .terraform from cache")
		if err := copyDir(cacheTerraformFolder, localTerraformFolder); err != nil {
			return err
		}

		logger.Debug("Restoring .terraform.lock.hcl from cache")
		err := copyFile(cacheLockFile, localLockFile)
		if err != nil {
			return err
		}

		return nil
	} else if os.IsNotExist(err) {
		if err := c.tf.Init(ctx); err != nil {
			return err
		}

		logger.Debug("Caching .terraform")
		if err := copyDir(localTerraformFolder, cacheTerraformFolder); err != nil {
			return errors.Wrap(err, "fail to update cache")
		}

		logger.Debug("Caching .terraform.lock.hcl")
		err := copyFile(path.Join(c.dir, ".terraform.lock.hcl"), path.Join(cacheBaseFolder, ".terraform.lock.hcl"))
		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.Wrap(err, "fail to check if .terraform is cached")
	}
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

func (c *client) Validate(ctx context.Context) (*tfjson.ValidateOutput, error) {
	hclog.FromContext(ctx).Debug("Running terraform validate")

	return c.tf.Validate(ctx)
}

func copyFile(sourcePath, destPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the file content
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Get the source file's permissions
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	permissions := sourceInfo.Mode()

	// Apply the same permissions to the destination file
	if err := destFile.Chmod(permissions); err != nil {
		return err
	}

	return nil
}

func copyDir(sourcePath, destPath string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", sourcePath)
	}

	if err := os.MkdirAll(destPath, sourceInfo.Mode()); err != nil {
		return err
	}

	dir, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		sourceFile := path.Join(sourcePath, file.Name())
		destFile := path.Join(destPath, file.Name())

		if file.IsDir() {
			if err := copyDir(sourceFile, destFile); err != nil {
				return err
			}
		} else {
			if err := copyFile(sourceFile, destFile); err != nil {
				return err
			}
		}
	}

	return nil
}

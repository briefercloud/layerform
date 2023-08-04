package command

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/pathutils"
	"github.com/ergomake/layerform/internal/states"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

type launchCommand struct {
	layers layers.Backend
	states states.Backend
}

var _ cli.Command = &launchCommand{}

func NewLaunch(layers layers.Backend, states states.Backend) *launchCommand {
	return &launchCommand{layers, states}
}

func (c *launchCommand) Help() string {
	return "launch help"
}

func (c *launchCommand) Synopsis() string {
	return "launch synopsis"
}

func (c *launchCommand) Run(args []string) int {
	ctx := context.Background()

	layerName := args[0]

	stateName := ""
	if len(args) > 1 {
		stateName = args[1]
	} else {
		stateName = shortuuid.New()
	}

	i := install.NewInstaller()
	i.SetLogger(log.Default())
	tfpath, err := i.Ensure(ctx, []src.Source{
		&fs.Version{
			Product:     product.Terraform,
			Constraints: version.MustConstraints(version.NewConstraint(">=1.1.0")),
		},
	})
	if err != nil {
		fmt.Println("fail to ensure terraform", err)
		return 1
	}

	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Println("fail to create work directory", err)
		return 1
	}
	fmt.Println(workdir)
	// defer os.RemoveAll(workdir)

	nextState, err := c.spawnLayer(ctx, layerName, stateName, workdir, tfpath)
	if err != nil {
		fmt.Println("fail to spawn layer", err)
		return 1
	}

	err = c.states.SaveState(layerName, stateName, nextState)
	if err != nil {
		fmt.Println("fail to save state", err)
		return 1
	}

	return 0
}

func (c *launchCommand) spawnLayer(ctx context.Context, layerName, stateName, workdir, tfpath string) ([]byte, error) {
	var inner func(layerName, stateName string) ([]byte, error)
	inner = func(layerName, stateName string) ([]byte, error) {
		layer, err := c.layers.GetLayer(layerName)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get layer")
		}

		if layer == nil {
			return nil, errors.New("layer not found")
		}

		layerWorkdir := path.Join(workdir, layer.Name)
		err = os.Mkdir(layerWorkdir, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "fail to create sub work directory for layer")
		}

		layerWorkdir, err = c.writeLayerToWorkdir(layerWorkdir, layer)
		if err != nil {
			return nil, errors.Wrap(err, "fail to write layer to workdir")
		}

		state, err := c.states.GetState(layerName, stateName)
		if err == nil {
			err := os.WriteFile(path.Join(layerWorkdir, "terraform.tfstate"), state.Bytes, 0644)
			if err != nil {
				return nil, errors.Wrap(err, "fail to write layer state to layer work dir")
			}
		} else if !errors.Is(err, states.ErrStateNotFound) {
			return nil, errors.Wrap(err, "fail to get state")
		}

		tf, err := tfexec.NewTerraform(layerWorkdir, tfpath)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get terraform client")
		}

		tf.SetStdout(os.Stdout)
		tf.SetStderr(os.Stderr)

		err = tf.Init(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to terraform init")
		}

		err = tf.Apply(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to terraform apply")
		}

		nextState, err := tf.StatePull(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to pull state")
		}

		return []byte(nextState), nil
	}

	return inner(layerName, stateName)
}

func (c *launchCommand) writeLayerToWorkdir(layerWorkdir string, layer *model.Layer) (string, error) {
	var inner func(*model.Layer) ([]string, error)
	inner = func(layer *model.Layer) ([]string, error) {
		fpaths := make([]string, 0)
		for _, dep := range layer.Dependencies {
			layer, err := c.layers.GetLayer(dep)
			if err != nil {
				return nil, errors.Wrap(err, "fail to get layer")
			}

			depPaths, err := inner(layer)
			if err != nil {
				return nil, errors.Wrap(err, "fail to write layer to workdir")
			}

			fpaths = append(fpaths, depPaths...)
		}

		for _, f := range layer.Files {
			fpaths = append(fpaths, f.Path)
			fpath := path.Join(layerWorkdir, f.Path)

			err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
			if err != nil {
				return fpaths, errors.Wrap(err, "fail to MkdirAll")
			}

			err = os.WriteFile(fpath, f.Content, 0644)
			if err != nil {
				return fpaths, errors.Wrap(err, "fail to write layer file")
			}
		}

		return fpaths, nil
	}

	paths, err := inner(layer)
	if err != nil {
		return "", errors.Wrap(err, "fail to write layer to workdir")
	}

	return path.Join(layerWorkdir, pathutils.FindCommonParentPath(paths)), nil
}

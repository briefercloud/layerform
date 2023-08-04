package command

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/lithammer/shortuuid/v3"
	"github.com/magodo/tfmerge/tfmerge"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
)

type spawnCommand2 struct {
	layersBackend layers.Backend
}

var _ cli.Command = &spawnCommand2{}

func NewSpawn2(layersBackend layers.Backend) *spawnCommand2 {
	return &spawnCommand2{layersBackend}
}

func (c *spawnCommand2) Help() string {
	return "spawn help"
}

func (c *spawnCommand2) Synopsis() string {
	return "spawn synopsis"
}

func (c *spawnCommand2) Run(args []string) int {
	layerName := args[0]
	stateName := ""
	if len(args) > 1 {
		stateName = args[1]
	} else {
		stateName = shortuuid.New()
	}
	fmt.Println(stateName)

	layer, err := c.layersBackend.GetLayer(layerName)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to get layer %s", layerName))
		return 1
	}

	if layer == nil {
		fmt.Printf("ERROR: Layer \"%s\" not found\n", layerName)
		return 1
	}

	ctx := context.Background()

	i := install.NewInstaller()
	tfpath, err := i.Ensure(ctx, []src.Source{
		&fs.Version{
			Product:     product.Terraform,
			Constraints: version.MustConstraints(version.NewConstraint(">=1.1.0")),
		},
	})
	if err != nil {
		fmt.Printf("ERROR: Fail to initialize terraform client\n%v", err)
		return 1
	}

	state, err := c.spawnRecursively(ctx, layer, tfpath)
	if err != nil {
		fmt.Printf("ERROR: Fail to spawn layer %s\n%v", layer.Name, err)
		return 1
	}

	fmt.Println("resulting state")
	fmt.Println(string(state))

	return 0
}

func mergeState2(ctx context.Context, baseState []byte, otherState []byte, tfpath string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "fail to create a temporary directory to merge terraform states")
	}

	otherStateFilepath := path.Join(dir, "otherstate.tfstate")
	err = os.WriteFile(otherStateFilepath, otherState, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "fail to write otherstate to temporary directory")
	}

	tf, err := tfexec.NewTerraform(dir, tfpath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to initialize terraform client")
	}

	return tfmerge.Merge(ctx, tf, baseState, otherStateFilepath)
}

func (c *spawnCommand2) spawnRecursively(ctx context.Context, layer *model.Layer, tfpath string) ([]byte, error) {
	var state []byte
	for _, depName := range layer.Dependencies {
		layer, err := c.layersBackend.GetLayer(depName)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get layer")
		}

		if layer == nil {
			return nil, errors.Wrapf(err, "could not find layer %s", depName)
		}

		depState, err := c.spawnRecursively(ctx, layer, tfpath)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to spawn layer %s", depName)
		}

		if state == nil {
			state = depState
		} else {
			state, err = mergeState2(ctx, depState, state, tfpath)
			if err != nil {
				return nil, errors.Wrap(err, "fail to merge states")
			}
		}
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "fail to create an empty directory to spawn layer from")
	}

	layerDir, err := c.materializeLayerWithDeps(layer, dir)
	if err != nil {
		return nil, errors.Wrap(err, "fail to write layer to a temporary directory")
	}

	stateFilepath := path.Join(layerDir, "terraform.tfstate")
	if state != nil {
		err = os.WriteFile(stateFilepath, state, 0644)
		if err != nil {
			return nil, errors.Wrap(err, "fail to write layer state to the layer temporary directory")
		}
	}

	tf, err := tfexec.NewTerraform(layerDir, tfpath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to initialize terraform client")
	}

	err = tf.Init(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to terraform init")
	}

	err = tf.Apply(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to terraform apply")
	}

	newState, err := os.ReadFile(stateFilepath)
	return newState, errors.Wrap(err, "fail to read new state after apply")
}

func (c *spawnCommand2) materializeLayerWithDeps(layer *model.Layer, dir string) (string, error) {
	deps, err := c.layersBackend.ResolveDependencies(layer)
	if err != nil {
		return "", errors.Wrapf(err, "fail to resolve dependencies of \"%s\"", layer.Name)
	}

	for _, d := range deps {
		_, err = c.materializeLayerWithDeps(d, dir)
		if err != nil {
			return "", errors.Wrapf(err, "fail to materialize layer dependencies of \"%s\"", layer.Name)
		}
	}

	layerDir, err := materializeLayerToDisk(layer, dir)
	return layerDir, errors.Wrapf(err, "fail to materialize layers \"%s\" to disk", layer.Name)
}

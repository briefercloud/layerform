package command

import (
	"fmt"
	"os"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/state"
	"github.com/ergomake/layerform/internal/terraform"
)

type killCommand struct {
	layersBackend   layers.Backend
	stateBackend    state.Backend
	terraformClient terraform.Client
}

var _ cli.Command = &killCommand{}

func NewKill(
	layersBackend layers.Backend,
	stateBackend state.Backend,
	terraformClient terraform.Client,
) *killCommand {
	return &killCommand{layersBackend, stateBackend, terraformClient}
}

func (c *killCommand) Help() string {
	return "kill help"
}

func (c *killCommand) Synopsis() string {
	return "kill synopsis"
}

func (c *killCommand) Run(args []string) int {
	layerName := args[0]

	instance := ""
	if len(args) > 1 {
		instance = args[1]
	} else {
		instance = shortuuid.New()
	}

	layer, err := c.layersBackend.GetLayer(layerName)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to get layer %s", layerName))
		return 1
	}

	if layer == nil {
		fmt.Printf("ERROR: Layer \"%s\" not found\n", layerName)
		return 1
	}

	err = c.kill(layer, instance)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to kill \"%s\" of layer \"%s\"\n", layerName, instance))
		return 1
	}

	return 0
}

func (c *killCommand) materializeLayerWithDeps(layer *model.Layer, dir string) (string, error) {
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

func (c *killCommand) kill(layer *model.Layer, instance string) error {
	dir, err := os.MkdirTemp("", fmt.Sprintf("layerform_%s", layer.Name))
	if err != nil {
		return errors.Wrapf(err, "fail to create a directory to materialize layers")
	}
	defer os.RemoveAll(dir)

	layerState, err := c.stateBackend.GetLayerState(layer, instance)
	if err != nil {
		return errors.Wrap(err, "fail to get layer state")
	}

	layersDir, err := c.materializeLayerWithDeps(layer, dir)
	if err != nil {
		return errors.Wrapf(err, "fail to materialize layer \"%s\" with dependencies", layer.Name)
	}

	deps, err := c.layersBackend.ResolveDependencies(layer)
	if err != nil {
		return errors.Wrapf(err, "fail to resolve dependencies of layer \"%s\"", layer.Name)
	}

	resources := map[string]struct{}{}
	for _, dep := range deps {
		depState, err := c.stateBackend.GetLayerState(dep, "default")
		if err != nil {
			return errors.Wrapf(err, "fail to get \"%s\" state of layer \"%s\"", "default", dep.Name)
		}

		diff := depState.Terraform().ResourceDiff(layerState.Terraform())
		if err != nil {
			return errors.Wrapf(
				err,
				"fail to compute resource diff between state \"%s\" of layer \"%s\" and state \"%s\" of layer \"%s\"", "default",
				dep.Name,
				instance,
				layer.Name,
			)
		}

		for _, res := range diff {
			resources[res.Address()] = struct{}{}
		}
	}

	err = c.terraformClient.Init(layersDir)
	if err != nil {
		return errors.Wrap(err, "fail to initialize terraform")
	}

	targets := []string{}
	for t := range resources {
		targets = append(targets, t)
	}

	fmt.Println("TARGETS", targets)

	_, err = c.terraformClient.Destroy(layersDir, layerState.Terraform(), targets...)
	if err != nil {
		return errors.Wrap(err, "fail to apply terraform")
	}

	err = c.stateBackend.RemoveLayerState(layer, instance)
	return errors.Wrapf(err, "fail to remove state \"%s\" of layer \"%s\"", instance, layer.Name)
}

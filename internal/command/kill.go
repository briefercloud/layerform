package command

import (
	"fmt"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

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

	state, err := c.stateBackend.GetLayerState(layer, instance)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to get layer %s %s state", layerName, instance))
		return 1
	}

	if state == nil {
		fmt.Printf("Instance %s of layer %s not found", instance, layer.Name)
		return 1
	}

	tmpDir, layerDir, err := materializeLayerToDisk(layer)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to materialize layer %s to disk", layerName))
		return 1
	}

	fmt.Println(tmpDir, layerDir)

	err = c.terraformClient.Init(layerDir)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrap(err, "fail to terraform init"))
		return 1
	}

	state, err = c.terraformClient.Destroy(layerDir, state)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrap(err, "fail to terraform init"))
		return 1
	}

	err = c.stateBackend.SaveLayerState(layer, instance, state)
	if err != nil {
		fmt.Printf("%v\n", err)
		return 1
	}

	return 0
}

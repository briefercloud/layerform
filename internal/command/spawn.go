package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/pathutils"
	"github.com/ergomake/layerform/internal/state"
	"github.com/ergomake/layerform/internal/terraform"
)

type spawnCommand struct {
	layersBackend   layers.Backend
	stateBackend    state.Backend
	terraformClient terraform.Client
}

var _ cli.Command = &spawnCommand{}

func NewSpawn(
	layersBackend layers.Backend,
	stateBackend state.Backend,
	terraformClient terraform.Client,
) *spawnCommand {
	return &spawnCommand{layersBackend, stateBackend, terraformClient}
}

func (c *spawnCommand) Help() string {
	return "spawn help"
}

func (c *spawnCommand) Synopsis() string {
	return "spawn synopsis"
}

func (c *spawnCommand) Run(args []string) int {
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

	// baseLayers, err := c.layersBackend.ResolveDependencies(layer)
	// if err != nil {
	//   fmt.Printf("ERROR: Fail to resolve \"%s\" dependencies\n", layer.Name)
	//   return 1
	// }

	// var baseState []byte
	// for _, bl := range baseLayers {
	// }

	err = c.spawn(layer, instance)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to spawn %s of layer %s", instance, layerName))
		return 1
	}

	return 0
}

func materializeLayerToDisk(layer *model.Layer, dir string) (string, error) {
	layerDir := dir

	if len(layer.Files) == 0 {
		return layerDir, nil
	}

	for _, file := range layer.Files {
		filePath := filepath.Join(dir, file.Path)

		// Ensure the parent directory exists.
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			os.RemoveAll(dir)
			return "", err
		}

		// Write the content to the file.
		if err := os.WriteFile(filePath, file.Content, 0644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}

	if len(layer.Files) > 0 {
		layerFilePaths := []string{}
		for _, f := range layer.Files {
			layerFilePaths = append(layerFilePaths, f.Path)
		}

		commonParent := pathutils.FindCommonParentPath(layerFilePaths)
		layerDir = path.Join(layerDir, commonParent)
	}

	return layerDir, nil
}

func mergeState(a *state.State, b *state.State) *state.State {
	// TODO: actually merge states? is that possible?
	if b == nil {
		return a
	}

	return b
}

func (c *spawnCommand) spawnRecursively(layer *model.Layer, instance string, dir string) (*state.State, error) {
	deps, err := c.layersBackend.ResolveDependencies(layer)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to resolve dependencies of layer \"%s\"", layer.Name)
	}

	var baseState *state.State
	for _, dep := range deps {
		state, err := c.spawnRecursively(dep, "default", dir)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to spawn dependecy \"%s\" of layer \"%s\"", dep.Name, layer.Name)
		}

		baseState = mergeState(baseState, state)
	}

	layerState, err := c.stateBackend.GetLayerState(layer, instance)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get layer %s %s state", layer.Name, instance)
	}

	layerState = mergeState(baseState, layerState)

	layerDir, err := materializeLayerToDisk(layer, dir)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to materialize layer %s to disk", layer.Name)
	}

	fmt.Printf("#######################################\n")
	fmt.Printf("[INFO]: Spawning \"%s\" of layer \"%s\"\n", instance, layer.Name)
	fmt.Printf("#######################################\n")

	err = c.terraformClient.Init(layerDir)
	if err != nil {
		return nil, errors.Wrap(err, "fail to terraform init")
	}

	tfState, err := c.terraformClient.Apply(layerDir, layerState.Terraform())
	if err != nil {
		return nil, errors.Wrap(err, "fail to terraform apply")
	}

	layerState = state.NewState(tfState)
	err = c.stateBackend.SaveLayerState(layer, instance, layerState)

	return layerState, errors.Wrapf(err, "fail to save state %s of layer %s", instance, layer.Name)
}

func (c *spawnCommand) spawn(layer *model.Layer, instance string) error {
	dir, err := os.MkdirTemp("", fmt.Sprintf("layerform_%s", layer.Name))
	if err != nil {
		return errors.Wrap(err, "fail to create a directory to materialize layers to")
	}
	defer os.RemoveAll(dir)

	state, err := c.spawnRecursively(layer, instance, dir)
	if err != nil {
		return errors.Wrapf(err, "fail to spawn layer \"%s\"", layer.Name)
	}

	err = c.stateBackend.SaveLayerState(layer, instance, state)
	return errors.Wrapf(err, "fail to save state \"%s\" of layer \"%s\"", instance, layer.Name)
}

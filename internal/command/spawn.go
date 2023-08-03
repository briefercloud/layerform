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

	state, err := c.stateBackend.GetLayerState(layer, instance)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to get layer %s %s state", layerName, instance))
		return 1
	}

	tmpDir, layerDir, err := materializeLayerToDisk(layer)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to materialize layer %s to disk", layerName))
		return 1
	}
	defer os.RemoveAll(tmpDir)

	err = c.terraformClient.Init(layerDir)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrap(err, "fail to terraform init"))
		return 1
	}

	state, err = c.terraformClient.Apply(layerDir, state)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrap(err, "fail to terraform apply"))
		return 1
	}

	err = c.stateBackend.SaveLayerState(layer, instance, state)
	if err != nil {
		fmt.Printf("%v\n", err)
		return 1
	}

	return 0
}

func materializeLayerToDisk(layer *model.Layer) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("layerform_%s", layer.Name))
	if err != nil {
		return "", "", err
	}

	layerDir := tmpDir

	if len(layer.Files) == 0 {
		return tmpDir, layerDir, nil
	}

	for _, file := range layer.Files {
		filePath := filepath.Join(tmpDir, file.Path)

		// Ensure the parent directory exists.
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			os.RemoveAll(tmpDir)
			return "", "", err
		}

		// Write the content to the file.
		if err := os.WriteFile(filePath, file.Content, 0644); err != nil {
			os.RemoveAll(tmpDir)
			return "", "", err
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

	return tmpDir, layerDir, nil
}

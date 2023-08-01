package command

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
  "github.com/lithammer/shortuuid/v3"

	"github.com/ergomake/layerform/client"
	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/pathutils"
)

type spawnCommand struct {
	layerformClient client.Client
}

var _ cli.Command = &spawnCommand{}

func NewSpawn(layerformClient client.Client) *spawnCommand {
	return &spawnCommand{layerformClient}
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

	layer, err := c.layerformClient.GetLayer(layerName)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to get layer %s", layerName))
		return 1
	}

	if layer == nil {
		fmt.Printf("ERROR: Layer \"%s\" not found\n", layerName)
		return 1
	}

	state, err := c.layerformClient.GetLayerState(layer, instance)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to get layer %s %s state", layerName, instance))
		return 1
	}

	tmpDir, layerDir, err := materializeLayerToDisk(layer, state)
	if err != nil {
		fmt.Printf("%v\n", errors.Wrapf(err, "fail to materialize layer %s to disk", layerName))
		return 1
	}

	fmt.Println(tmpDir, layerDir)

	cmd := exec.Command("terraform", "init")
	cmd.Dir = layerDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
		return cmd.ProcessState.ExitCode()
	}

	cmd = exec.Command("terraform", "apply")
	cmd.Dir = layerDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		fmt.Printf("%v\n", err)
		return cmd.ProcessState.ExitCode()
	}

  if cmd.ProcessState.ExitCode() != 0 {
    return cmd.ProcessState.ExitCode()
  }

  state, err = os.ReadFile(path.Join(layerDir, "terraform.tfstate"))
  if err != nil {
		fmt.Printf("%v\n", err)
    return 1
  }

  err = c.layerformClient.SaveLayerState(layer, instance, state)
  if err != nil {
		fmt.Printf("%v\n", err)
    return 1
  }

	return 0
}

func materializeLayerToDisk(layer *model.Layer, state []byte) (string, string, error) {
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

	if state != nil {
		err := os.WriteFile(path.Join(layerDir, "terraform.tfstate"), state, 0644)
		if err != nil {
			os.RemoveAll(tmpDir)
			return "", "", err
		}
	}

	return tmpDir, layerDir, nil
}

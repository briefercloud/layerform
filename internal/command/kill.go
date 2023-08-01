package command

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/client"
)

type killCommand struct {
	layerformClient client.Client
}

var _ cli.Command = &killCommand{}

func NewKill(layerformClient client.Client) *killCommand {
	return &killCommand{layerformClient}
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

  if state == nil {
		fmt.Printf("Instance %s of layer %s not found", instance, layer.Name)
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

	cmd = exec.Command("terraform", "destroy")
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


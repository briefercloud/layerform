package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/cmdexec"
	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/state"
	"github.com/ergomake/layerform/internal/terraform"
)

func main() {
	layerfile, err := layerfile.FromFile("layerform.json")
	if err != nil {
		panic(errors.Wrap(err, "fail to load layerform.json"))
	}

	layerslist, err := layerfile.ToLayers()
	if err != nil {
		panic(errors.Wrap(err, "fail to import layers defined at layerform.json"))
	}

	stateBackend, err := state.NewFileBackend("layerform.lfstate")
	if err != nil {
		panic(errors.Wrap(err, "fail to initialize a state backend backed by file"))
	}

	layersBackend := layers.NewInMemoryBackend(layerslist)

	terraformClient := terraform.NewCLI(&cmdexec.OSCommandExecutor{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	// TODO: fix hardcoded version
	c := cli.NewCLI("layerform", "0.0.1")

	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"spawn": func() (cli.Command, error) {
			return command.NewSpawn(layersBackend, stateBackend, terraformClient), nil
		},
		"kill": func() (cli.Command, error) {
			return command.NewKill(layersBackend, stateBackend, terraformClient), nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

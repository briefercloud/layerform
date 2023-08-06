package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
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

	layersBackend := layers.NewInMemoryBackend(layerslist)
	statesBackend := layerstate.NewFileBackend("layerform.lfstate")

	// TODO: fix hardcoded version
	c := cli.NewCLI("layerform", "0.0.1")

	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"spawn": func() (cli.Command, error) {
			return command.NewLaunch(layersBackend, statesBackend), nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

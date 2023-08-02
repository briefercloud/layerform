package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/client"
	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/commandexecutor"
	"github.com/ergomake/layerform/internal/terraform"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrap(err, "fail to fetch user home directory"))
	}

	apiClient, err := client.NewFileClient(filepath.Join(homeDir, ".layerform.state.json"))
	if err != nil {
		panic(errors.Wrap(err, "fail to create Layerform API Client"))
	}

	cmdExecutor := &commandexecutor.OSCommandExecutor{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	terraformClient := terraform.NewCLI(cmdExecutor)

	// TODO: fix hardcoded version
	c := cli.NewCLI("layerform", "0.0.1")

	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		// TODO: this command will most likely be replaced with a proper terraform provider
		"import": func() (cli.Command, error) {
			return command.NewImport(apiClient), nil
		},
		"spawn": func() (cli.Command, error) {
			return command.NewSpawn(apiClient, terraformClient), nil
		},
		"kill": func() (cli.Command, error) {
			return command.NewKill(apiClient, terraformClient), nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

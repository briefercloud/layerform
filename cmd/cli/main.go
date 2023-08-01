package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ergomake/layerform/client"
	"github.com/ergomake/layerform/internal/command"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
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

	// TODO: fix hardcoded version
	c := cli.NewCLI("layerform", "0.0.1")

	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		// TODO: this command will most likely be replaced with a proper terraform provider
		"import": func() (cli.Command, error) {
			return command.NewImport(apiClient), nil
		},
		"spawn": func() (cli.Command, error) {
			return command.NewSpawn(apiClient), nil
		},
		"kill": func() (cli.Command, error) {
			return command.NewKill(apiClient), nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)

	// cli := cli.New(apiClient)

	// layerName := "eks"
	// err = cli.Spawn(layerName, "")
	// if err != nil {
	// 	panic(errors.Wrapf(err, "fail to spawn layer %s", layerName))
	// }
}

// func runTerraformApply(layer *model.Layer) error {
// 	layerDir, err := materializeLayerToDisk(layer)
// 	if err != nil {
// 		return err
// 	}
// 	defer os.RemoveAll(layerDir)
//
// 	if len(layer.Files) > 0 {
//     layerFilePaths := []string{}
//     for _, f := range layer.Files {
//       layerFilePaths = append(layerFilePaths, f.Path)
//     }
//
// 		commonParent := pathutils.FindCommonParentPath(layerFilePaths)
// 		layerDir = path.Join(layerDir, commonParent)
// 	}
//
// 	// Get the current working directory.
// 	cwd, err := os.Getwd()
// 	if err != nil {
// 		return err
// 	}
//
// 	err = os.Chdir(layerDir)
// 	if err != nil {
// 		return err
// 	}
// 	defer os.Chdir(cwd)
//
// 	cmd := exec.Command("terraform", "init")
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	err = cmd.Run()
// 	if err != nil {
// 		return err
// 	}
//
// 	cmd = exec.Command("terraform", "apply")
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	err = cmd.Run()
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func materializeLayerToDisk(layer *model.Layer) (string, error) {
// 	tempDir, err := ioutil.TempDir("", fmt.Sprintf("layerform_%s", layer.Name))
// 	if err != nil {
// 		return "", err
// 	}
//
// 	if len(layer.Files) == 0 {
// 		return tempDir, nil
// 	}
//
// 	for _, file := range layer.Files {
// 		filePath := filepath.Join(tempDir, file.Path)
//
// 		// Ensure the parent directory exists.
// 		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
// 			os.RemoveAll(tempDir)
// 			return "", err
// 		}
//
// 		// Write the content to the file.
// 		if err := ioutil.WriteFile(filePath, file.Content, 0644); err != nil {
// 			os.RemoveAll(tempDir)
// 			return "", err
// 		}
// 	}
//
// 	return tempDir, nil
// }

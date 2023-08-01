package command

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/ergomake/layerform/client"
	"github.com/ergomake/layerform/internal/data/model"
	"github.com/mitchellh/cli"
)

type layerItem struct {
	Name         string   `json:"name"`
	Files        []string `json:"files"`
	Dependencies []string `json:"dependencies"`
}
type layers []layerItem

type importCommand struct {
	layerformClient client.Client
}

var _ cli.Command = &importCommand{}

func NewImport(layerformClient client.Client) *importCommand {
	return &importCommand{layerformClient}
}

func (c *importCommand) Help() string {
	return "import help"
}

func (c *importCommand) Synopsis() string {
	return "import synopsis"
}

func (c *importCommand) Run(args []string) int {
	configPath := "layerform.json"
	if len(args) > 0 {
		configPath = args[0]
	}

	bs, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("ERROR: Could not read %s.\n\t%v.", configPath, err)
		return 1
	}

	var ls layers
	err = json.Unmarshal(bs, &ls)
	if err != nil {
		fmt.Printf("ERROR: Fail to unmarshal %s.\n\t%v.\n", configPath, err)
		return 1
	}

	configDir := path.Dir(configPath)
	modelLayers := []*model.Layer{}
	for _, l := range ls {
		files := make([]model.LayerFile, len(l.Files))
		for i, f := range l.Files {
			fPath := path.Join(configDir, f)
			content, err := os.ReadFile(fPath)
			if err != nil {
				fmt.Printf("ERROR: Could not read %s.\n\t%v.\n", fPath, err)
				return 1
			}

			files[i] = model.LayerFile{
				Path:    f,
				Content: content,
			}
		}

		layer := &model.Layer{
			Name:         l.Name,
			Files:        files,
			Dependencies: l.Dependencies,
		}
		modelLayers = append(modelLayers, layer)
	}

	for _, ml := range modelLayers {
		_, err := c.layerformClient.CreateLayer(ml)
		if err != nil {
			fmt.Printf("ERROR: Could not create layer %s.\n\t%v.\n", ml.Name, err)
			return 1
		}
	}

	return 0
}

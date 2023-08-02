package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ergomake/layerform/internal/data/model"
	clientMock "github.com/ergomake/layerform/mocks/client"
	tfMock "github.com/ergomake/layerform/mocks/internal_/terraform"
)

func TestCommandSpawn_Run(t *testing.T) {
	layerName := "eks"
	instanceName := "instance"

	layer := &model.Layer{
		Name: layerName,
		Files: []model.LayerFile{
			{
				Path:    "layers/main.tf",
				Content: []byte("layers/main.tf mock content"),
			},
		},
		Dependencies: []string{},
	}

	state := []byte("current state")

	layerClient := clientMock.NewClient(t)
	layerClient.EXPECT().GetLayer(layerName).Return(layer, nil)
	layerClient.EXPECT().GetLayerState(layer, instanceName).Return(state, nil)
	layerClient.EXPECT().SaveLayerState(layer, instanceName, []byte("next state")).Return(nil)

	tfClient := tfMock.NewClient(t)
	tfClient.EXPECT().Init(mock.Anything).Return(nil)
	tfClient.EXPECT().Apply(mock.Anything, state).Return([]byte("next state"), nil)

	spawn := NewSpawn(layerClient, tfClient)

	exit := spawn.Run([]string{layerName, instanceName})
	assert.Equal(t, 0, exit)
}

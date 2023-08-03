package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ergomake/layerform/internal/data/model"
	layersMock "github.com/ergomake/layerform/mocks/internal_/layers"
	stateMock "github.com/ergomake/layerform/mocks/internal_/state"
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

	layersBackend := layersMock.NewBackend(t)
	layersBackend.EXPECT().GetLayer(layerName).Return(layer, nil)

	stateBackend := stateMock.NewBackend(t)
	stateBackend.EXPECT().GetLayerState(layer, instanceName).Return(state, nil)
	stateBackend.EXPECT().SaveLayerState(layer, instanceName, []byte("next state")).Return(nil)

	tfClient := tfMock.NewClient(t)
	tfClient.EXPECT().Init(mock.Anything).Return(nil)
	tfClient.EXPECT().Apply(mock.Anything, state).Return([]byte("next state"), nil)

	spawn := NewSpawn(layersBackend, stateBackend, tfClient)

	exit := spawn.Run([]string{layerName, instanceName})
	assert.Equal(t, 0, exit)
}

package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/state"
	"github.com/ergomake/layerform/internal/terraform"
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

	currState := state.NewState(&terraform.State{Bytes: []byte("current state")})

	layersBackend := layersMock.NewBackend(t)
	layersBackend.EXPECT().GetLayer(layerName).Return(layer, nil)
	layersBackend.EXPECT().ResolveDependencies(layer).Return([]*model.Layer{}, nil)

	nextState := state.NewState(&terraform.State{Bytes: []byte("next state")})

	stateBackend := stateMock.NewBackend(t)
	stateBackend.EXPECT().GetLayerState(layer, instanceName).Return(currState, nil)
	stateBackend.EXPECT().SaveLayerState(layer, instanceName, nextState).Return(nil)

	tfClient := tfMock.NewClient(t)
	tfClient.EXPECT().Init(mock.Anything).Return(nil)
	tfClient.EXPECT().Apply(mock.Anything, currState.Terraform()).Return(nextState.Terraform(), nil)

	spawn := NewSpawn(layersBackend, stateBackend, tfClient)

	exit := spawn.Run([]string{layerName, instanceName})
	assert.Equal(t, 0, exit)
}

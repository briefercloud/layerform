package layers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/layerform/internal/data/model"
)

func TestLayers_InMemoryBackend(t *testing.T) {
	layers := []*model.Layer{
		{Name: "layer1"},
		{Name: "layer2"},
	}
	stateBackend := NewInMemoryBackend(layers)

	layer1, err := stateBackend.GetLayer("layer1")
	require.NoError(t, err)
	assert.Equal(t, layers[0], layer1)

	layer2, err := stateBackend.GetLayer("layer2")
	require.NoError(t, err)
	assert.Equal(t, layers[1], layer2)

	layer3, err := stateBackend.GetLayer("layer3")
	require.NoError(t, err)
	assert.Nil(t, layer3)
}

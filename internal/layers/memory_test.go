package layers

import (
	"context"
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

	layer1, err := stateBackend.GetLayer(context.Background(), "layer1")
	require.NoError(t, err)
	assert.Equal(t, layers[0], layer1)

	layer2, err := stateBackend.GetLayer(context.Background(), "layer2")
	require.NoError(t, err)
	assert.Equal(t, layers[1], layer2)

	layer3, err := stateBackend.GetLayer(context.Background(), "layer3")
	require.NoError(t, err)
	assert.Nil(t, layer3)
}

func TestInMemoryBackend_ResolveDependencies(t *testing.T) {
	layer1 := &model.Layer{Name: "layer1", Dependencies: []string{"layer2"}}
	layer2 := &model.Layer{Name: "layer2", Dependencies: []string{"layer3"}}
	layer3 := &model.Layer{Name: "layer3", Dependencies: []string{"layer4"}}

	stateBackend := NewInMemoryBackend([]*model.Layer{layer1, layer2, layer3})

	t.Run("single dependency", func(t *testing.T) {
		dependencies, err := stateBackend.ResolveDependencies(context.Background(), layer1)
		require.NoError(t, err)
		assert.Len(t, dependencies, 1)
		assert.Equal(t, layer2, dependencies[0])
	})

	t.Run("dependency not found", func(t *testing.T) {
		_, err := stateBackend.ResolveDependencies(context.Background(), layer3)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		layer4 := &model.Layer{Name: "layer4", Dependencies: []string{"layer2", "layer3"}}
		dependencies, err := stateBackend.ResolveDependencies(context.Background(), layer4)
		require.NoError(t, err)
		assert.Len(t, dependencies, 2)
		assert.Equal(t, layer2, dependencies[0])
		assert.Equal(t, layer3, dependencies[1])
	})
}

package layers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/layerform/pkg/data"
)

func setup(ls []*data.Layer) *fileLikeBackend {
	layers := make(map[string]*data.Layer)
	for _, layer := range ls {
		layers[layer.Name] = layer
	}

	return &fileLikeBackend{
		data: &fileLikeModel{
			Version: bloblayersVersion,
			Layers:  layers,
		},
		storage: nil,
	}
}

func TestFileLikeBackendGetLayer(t *testing.T) {
	layers := []*data.Layer{
		{Name: "layer1"},
		{Name: "layer2"},
	}
	stateBackend := setup(layers)

	layer1, err := stateBackend.GetLayer(context.Background(), "layer1")
	require.NoError(t, err)
	assert.Equal(t, layers[0], layer1)

	layer2, err := stateBackend.GetLayer(context.Background(), "layer2")
	require.NoError(t, err)
	assert.Equal(t, layers[1], layer2)

	_, err = stateBackend.GetLayer(context.Background(), "layer3")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFileLikeBackend_ResolveDependencies(t *testing.T) {
	layer1 := &data.Layer{Name: "layer1", Dependencies: []string{"layer2"}}
	layer2 := &data.Layer{Name: "layer2", Dependencies: []string{"layer3"}}
	layer3 := &data.Layer{Name: "layer3", Dependencies: []string{"layer4"}}

	stateBackend := setup([]*data.Layer{layer1, layer2, layer3})

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
		layer4 := &data.Layer{Name: "layer4", Dependencies: []string{"layer2", "layer3"}}
		dependencies, err := stateBackend.ResolveDependencies(context.Background(), layer4)
		require.NoError(t, err)
		assert.Len(t, dependencies, 2)
		assert.Equal(t, layer2, dependencies[0])
		assert.Equal(t, layer3, dependencies[1])
	})
}

func TestFileLikeBackend_ListLayers(t *testing.T) {
	layer1 := &data.Layer{Name: "layer1"}
	layer2 := &data.Layer{Name: "layer2"}
	layer3 := &data.Layer{Name: "layer3"}

	stateBackend := setup([]*data.Layer{layer1, layer2, layer3})

	t.Run("list all layers", func(t *testing.T) {
		list, err := stateBackend.ListLayers(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 3)
		assert.Contains(t, list, layer1)
		assert.Contains(t, list, layer2)
		assert.Contains(t, list, layer3)
	})

	t.Run("empty list", func(t *testing.T) {
		emptyBackend := setup([]*data.Layer{})
		list, err := emptyBackend.ListLayers(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})
}

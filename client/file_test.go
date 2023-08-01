package client

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/layerform/internal/data/model"
)

func TestFile_NewFileClient(t *testing.T) {
	// Test the case when the file does not exist
	t.Run("when the file does not exists", func(t *testing.T) {
		filePath := "test_file_not_exist.json"
		t.Cleanup(func() {
			os.Remove(filePath)
		})

		client, err := NewFileClient(filePath)
		assert.FileExists(t, filePath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Empty(t, client.content.Layers)
		assert.Empty(t, client.content.State)
	})

	t.Run("when the file exists with valid JSON data", func(t *testing.T) {
		existingFilePath := "test_existing_file.json"
		t.Cleanup(func() {
			os.Remove(existingFilePath)
		})

		existingData := &fileContent{
			Layers: map[string]*model.Layer{
				"test_layer": {Name: "test_layer"},
			},
			State: make(map[string]map[string][]byte),
		}
		existingDataJSON, _ := json.Marshal(existingData)
		err := os.WriteFile(existingFilePath, existingDataJSON, 0644)
		assert.NoError(t, err)

		client, err := NewFileClient(existingFilePath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, existingData, client.content)
	})

	t.Run("errors when file exists with invalid data", func(t *testing.T) {
		existingFilePath := "test_existing_file.json"
		t.Cleanup(func() {
			os.Remove(existingFilePath)
		})

		err := os.WriteFile(existingFilePath, []byte("invalid data"), 0644)
		assert.NoError(t, err)

		_, err = NewFileClient(existingFilePath)
		assert.Error(t, err)
	})
}

func TestFile_ReadLayer(t *testing.T) {
	// Create a fileClient with some test data
	filePath := "test_read_layer.json"
	defer os.Remove(filePath)

	testData := fileContent{
		Layers: map[string]*model.Layer{
			"test_layer":    {Name: "test_layer"},
			"another_layer": {Name: "another_layer"},
		},
	}

	dataJSON, _ := json.Marshal(testData)
	err := os.WriteFile(filePath, dataJSON, 0644)
	assert.NoError(t, err)

	client, _ := NewFileClient(filePath)

	expectedLayer := testData.Layers["test_layer"]
	layer, err := client.GetLayer("test_layer")
	assert.NoError(t, err)
	assert.Equal(t, expectedLayer, layer)

	// Test reading a non-existing layer
	layer, err = client.GetLayer("a_third_layer")
	assert.NoError(t, err)
	assert.Nil(t, layer)
}

func TestFile_CreateLayer(t *testing.T) {
	// Create a fileClient
	filePath := "test_create_layer.json"
	defer os.Remove(filePath)

	client, err := NewFileClient(filePath)
	require.NoError(t, err)

	// Test creating a new layer
	newLayer := &model.Layer{
		Name: "new_layer",
		Files: []model.LayerFile{
			{Path: "path", Content: make([]byte, 0)},
		},
	}
	createdLayer, err := client.CreateLayer(newLayer)
	require.NoError(t, err)
	require.NotNil(t, createdLayer)
	assert.Equal(t, newLayer, client.content.Layers[createdLayer.Name])

	// Test creating a new layer with an existing name
	existingLayerName := &model.Layer{
		Name: createdLayer.Name,
	}
	_, err = client.CreateLayer(existingLayerName)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLayerAlreadyExists)
}

func TestFile_GetLayerState(t *testing.T) {
	// Create a fileClient with some test data
	filePath := "test_get_layer_state.json"
	defer os.Remove(filePath)

	testData := fileContent{
		Layers: map[string]*model.Layer{
			"test_layer": {Name: "test_layer"},
			"another_layer": {Name: "another_layer"},
		},
		State: map[string]map[string][]byte{
			"test_layer": {
				"instance1": []byte("state_data_instance1"),
				"instance2": []byte("state_data_instance2"),
			},
		},
	}

	dataJSON, _ := json.Marshal(testData)
	err := os.WriteFile(filePath, dataJSON, 0644)
	require.NoError(t, err)

	client, _ := NewFileClient(filePath)

	// Test fetching a layer state for an existing layer and instance
	expectedState := []byte("state_data_instance1")
	state, err := client.GetLayerState(testData.Layers["test_layer"], "instance1")
	require.NoError(t, err)
	assert.Equal(t, expectedState, state)

	// Test fetching a layer state for an existing layer but non-existing instance
	state, err = client.GetLayerState(testData.Layers["test_layer"], "non_existing_instance")
	require.NoError(t, err)
	assert.Nil(t, state)
}

func TestFile_SaveLayerState(t *testing.T) {
	// Create a fileClient with some test data
	filePath := "test_save_layer_state.json"
	defer os.Remove(filePath)

	testData := fileContent{
		Layers: map[string]*model.Layer{
			"test_layer": {Name: "test_layer"},
			"another_layer": {Name: "another_layer"},
		},
		State: map[string]map[string][]byte{
			"test_layer": {
				"instance1": []byte("state_data_instance1"),
				"instance2": []byte("state_data_instance2"),
			},
		},
	}

	dataJSON, _ := json.Marshal(testData)
	err := os.WriteFile(filePath, dataJSON, 0644)
	require.NoError(t, err)

	client, _ := NewFileClient(filePath)

	// Test saving a layer state for an existing layer and instance
	err = client.SaveLayerState(testData.Layers["test_layer"], "new_instance", []byte("new_state_data"))
	require.NoError(t, err)
	expectedState := []byte("new_state_data")
	state, err := client.GetLayerState(testData.Layers["test_layer"], "new_instance")
	require.NoError(t, err)
	assert.Equal(t, expectedState, state)

	// Test saving a layer state for an existing layer but non-existing instance
	err = client.SaveLayerState(testData.Layers["test_layer"], "non_existing_instance", []byte("state_data"))
	require.NoError(t, err)
	expectedState = []byte("state_data")
	state, err = client.GetLayerState(testData.Layers["test_layer"], "non_existing_instance")
	require.NoError(t, err)
	assert.Equal(t, expectedState, state)
}

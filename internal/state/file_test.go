package state

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/layerform/internal/data/model"
)

func TestState_FileBackend_NewFileBackend(t *testing.T) {
	// Test the case when the file does not exist
	t.Run("when the file does not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := path.Join(tmpDir, "test_file_not_exist.json")

		client, err := NewFileBackend(filePath)
		assert.FileExists(t, filePath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Empty(t, client.state.State)
	})

	t.Run("when the file exists with valid JSON data", func(t *testing.T) {
		tmpDir := t.TempDir()
		existingFilePath := path.Join(tmpDir, "test_existing_file.json")

		existingData := &fileContent{
			State: map[string]map[string][]byte{
				"eks": {
					"default": []byte("test state"),
				},
			},
		}
		existingDataJSON, _ := json.Marshal(existingData)
		err := os.WriteFile(existingFilePath, existingDataJSON, 0644)
		assert.NoError(t, err)

		client, err := NewFileBackend(existingFilePath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, existingData, client.state)
	})

	t.Run("errors when file exists with invalid data", func(t *testing.T) {
		tmpDir := t.TempDir()
		existingFilePath := path.Join(tmpDir, "test_existing_file.json")

		err := os.WriteFile(existingFilePath, []byte("invalid data"), 0644)
		assert.NoError(t, err)

		_, err = NewFileBackend(existingFilePath)
		assert.Error(t, err)
	})
}

func TestState_FileBackend_GetLayerState(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "test_get_layer_state.json")

	testData := fileContent{
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

	client, _ := NewFileBackend(filePath)

	layer := &model.Layer{
		Name: "test_layer",
	}

	// Test fetching a layer state for an existing layer and instance
	expectedState := []byte("state_data_instance1")
	state, err := client.GetLayerState(layer, "instance1")
	require.NoError(t, err)
	assert.Equal(t, expectedState, state)

	// Test fetching a layer state for an existing layer but non-existing instance
	state, err = client.GetLayerState(layer, "non_existing_instance")
	require.NoError(t, err)
	assert.Nil(t, state)
}

func TestState_FileBackend_SaveLayerState(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "test_save_layer_state.json")

	testData := fileContent{
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

	client, _ := NewFileBackend(filePath)

	layer := &model.Layer{
		Name: "test_layer",
	}

	// Test saving a layer state for an existing layer and instance
	err = client.SaveLayerState(layer, "new_instance", []byte("new_state_data"))
	require.NoError(t, err)
	expectedState := []byte("new_state_data")
	state, err := client.GetLayerState(layer, "new_instance")
	require.NoError(t, err)
	assert.Equal(t, expectedState, state)

	// Test saving a layer state for an existing layer but non-existing instance
	err = client.SaveLayerState(layer, "non_existing_instance", []byte("state_data"))
	require.NoError(t, err)
	expectedState = []byte("state_data")
	state, err = client.GetLayerState(layer, "non_existing_instance")
	require.NoError(t, err)
	assert.Equal(t, expectedState, state)
}

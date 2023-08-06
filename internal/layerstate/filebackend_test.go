package layerstate

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBackend_GetState(t *testing.T) {
	t.Run("state found", func(t *testing.T) {
		tempDir := t.TempDir()

		state := &State{
			LayerName: "testLayer",
			StateName: "testState",
			Bytes:     []byte("state1"),
		}

		fb := NewFileBackend(tempDir + "/testfile.json")

		// adds new state
		err := fb.SaveState(state.LayerName, state.StateName, state.Bytes)
		require.NoError(t, err)

		result, err := fb.GetState(state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, state, result)

		// updates existing state
		state.Bytes = []byte("state2")
		err = fb.SaveState(state.LayerName, state.StateName, state.Bytes)
		require.NoError(t, err)

		result, err = fb.GetState(state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, "state2", string(result.Bytes))

		// preserves existing state
		state2 := &State{
			LayerName: "testLayer2",
			StateName: "testState2",
			Bytes:     []byte("state3"),
		}
		err = fb.SaveState(state2.LayerName, state2.StateName, state2.Bytes)
		require.NoError(t, err)

		result, err = fb.GetState(state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, state, result)
	})

	t.Run("state not found", func(t *testing.T) {
		tempDir := t.TempDir()

		fb := NewFileBackend(tempDir + "/testfile.json")
		_, err := fb.GetState("nonExistentLayer", "nonExistentState")
		assert.ErrorIs(t, err, ErrStateNotFound)
	})

	t.Run("fail to read file", func(t *testing.T) {
		tempDir := t.TempDir()

		// write invalid json to the file to force parse failure
		err := os.WriteFile(tempDir+"/testfile.json", []byte("this is not valid json"), 0644)
		require.NoError(t, err)

		fb := NewFileBackend(tempDir + "/testfile.json")

		_, err = fb.GetState("not-importannt", "not-important")
		assert.Error(t, err)
	})
}

func TestFileBackend_SaveState(t *testing.T) {
	t.Run("adds or update state correctly", func(t *testing.T) {
		tempDir := t.TempDir()

		fb := NewFileBackend(tempDir + "/testfile.json")

		err := fb.SaveState("layer1", "state1", []byte("data1"))
		require.NoError(t, err)

		data, err := os.ReadFile(tempDir + "/testfile.json")
		require.NoError(t, err)

		var fstate filestate
		err = json.Unmarshal(data, &fstate)
		require.NoError(t, err)

		assert.Len(t, fstate.States, 1)
		assert.Equal(t, "layer1", fstate.States[0].LayerName)
		assert.Equal(t, "state1", fstate.States[0].StateName)
		assert.Equal(t, []byte("data1"), fstate.States[0].Bytes)
	})

	t.Run("fail to read file", func(t *testing.T) {
		tempDir := t.TempDir()

		// write invalid json to the file to force parse failure
		err := os.WriteFile(tempDir+"/testfile.json", []byte("this is not valid json"), 0644)
		require.NoError(t, err)

		fb := NewFileBackend(tempDir + "/testfile.json")

		err = fb.SaveState("not-importannt", "not-important", []byte("not-important"))
		assert.Error(t, err)
	})
}

func TestFileBackend_readFile(t *testing.T) {
	t.Run("fail to read file", func(t *testing.T) {
		tempDir := t.TempDir()

		err := os.WriteFile(tempDir+"/testfile.json", []byte(`{"states": []}`), 0644)
		require.NoError(t, err)

		// change permissions of testfile.json to make it unreadable
		err = os.Chmod(tempDir+"/testfile.json", 0000)
		require.NoError(t, err)

		fb := NewFileBackend(tempDir + "/testfile.json")

		_, err = fb.readFile()
		assert.Error(t, err)
	})

	t.Run("fail to read file", func(t *testing.T) {
		tempDir := t.TempDir()

		// write invalid json to the file to force parse failure
		err := os.WriteFile(tempDir+"/testfile.json", []byte("this is not valid json"), 0644)
		require.NoError(t, err)

		fb := NewFileBackend(tempDir + "/testfile.json")

		_, err = fb.readFile()
		assert.Error(t, err)
	})
}

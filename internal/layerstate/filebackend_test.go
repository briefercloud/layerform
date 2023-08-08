package layerstate

import (
	"context"
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
			LayerName:         "testLayer",
			StateName:         "testState",
			DependenciesState: map[string]string{"base": "testBaseStae"},
			Bytes:             []byte("state1"),
		}

		fb := NewFileBackend(tempDir + "/testfile.json")

		// adds new state
		err := fb.SaveState(context.Background(), state)
		require.NoError(t, err)

		result, err := fb.GetState(context.Background(), state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, state, result)

		// updates existing state
		state.Bytes = []byte("state2")
		err = fb.SaveState(context.Background(), state)
		require.NoError(t, err)

		result, err = fb.GetState(context.Background(), state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, "state2", string(result.Bytes))

		// preserves existing state
		state2 := &State{
			LayerName: "testLayer2",
			StateName: "testState2",
			Bytes:     []byte("state3"),
		}
		err = fb.SaveState(context.Background(), state2)
		require.NoError(t, err)

		result, err = fb.GetState(context.Background(), state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, state, result)
	})

	t.Run("state not found", func(t *testing.T) {
		tempDir := t.TempDir()

		fb := NewFileBackend(tempDir + "/testfile.json")
		_, err := fb.GetState(context.Background(), "nonExistentLayer", "nonExistentState")
		assert.ErrorIs(t, err, ErrStateNotFound)
	})

	t.Run("fail to read file", func(t *testing.T) {
		tempDir := t.TempDir()

		// write invalid json to the file to force parse failure
		err := os.WriteFile(tempDir+"/testfile.json", []byte("this is not valid json"), 0644)
		require.NoError(t, err)

		fb := NewFileBackend(tempDir + "/testfile.json")

		_, err = fb.GetState(context.Background(), "not-importannt", "not-important")
		assert.Error(t, err)
	})
}

func TestFileBackend_SaveState(t *testing.T) {
	t.Run("adds or update state correctly", func(t *testing.T) {
		tempDir := t.TempDir()

		fb := NewFileBackend(tempDir + "/testfile.json")

		state := &State{
			LayerName: "layer1",
			StateName: "state1",
			Bytes:     []byte("data1"),
		}
		err := fb.SaveState(context.Background(), state)
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

		state := &State{
			LayerName: "not-importannt",
			StateName: "not-important",
			Bytes:     []byte("not-important"),
		}
		err = fb.SaveState(context.Background(), state)
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

		_, err = fb.readFile(context.Background())
		assert.Error(t, err)
	})

	t.Run("fail to read file", func(t *testing.T) {
		tempDir := t.TempDir()

		// write invalid json to the file to force parse failure
		err := os.WriteFile(tempDir+"/testfile.json", []byte("this is not valid json"), 0644)
		require.NoError(t, err)

		fb := NewFileBackend(tempDir + "/testfile.json")

		_, err = fb.readFile(context.Background())
		assert.Error(t, err)
	})
}

func TestFileBackend_DeleteState(t *testing.T) {
	tempDir := t.TempDir()

	state1 := &State{
		LayerName: "layer1",
		StateName: "state1",
		Bytes:     []byte("data1"),
	}

	state2 := &State{
		LayerName: "layer2",
		StateName: "state2",
		Bytes:     []byte("data2"),
	}

	state3 := &State{
		LayerName: "layer1",
		StateName: "state3",
		Bytes:     []byte("data3"),
	}

	fb := NewFileBackend(tempDir + "/testfile.json")

	err := fb.SaveState(context.Background(), state1)
	require.NoError(t, err)

	err = fb.SaveState(context.Background(), state2)
	require.NoError(t, err)

	err = fb.SaveState(context.Background(), state3)
	require.NoError(t, err)

	t.Run("delete existing state", func(t *testing.T) {
		err := fb.DeleteState(context.Background(), state1.LayerName, state1.StateName)
		require.NoError(t, err)

		states, err := fb.ListStatesByLayer(context.Background(), state1.LayerName)
		require.NoError(t, err)
		assert.Len(t, states, 1)
		assert.Equal(t, state3, states[0])
	})

	t.Run("delete non-existent state", func(t *testing.T) {
		err := fb.DeleteState(context.Background(), "nonExistentLayer", "nonExistentState")
		assert.NoError(t, err)
	})
}

func TestFileBackend_ListStatesByLayer(t *testing.T) {
	tempDir := t.TempDir()

	state1 := &State{
		LayerName: "layer1",
		StateName: "state1",
		Bytes:     []byte("data1"),
	}

	state2 := &State{
		LayerName: "layer2",
		StateName: "state2",
		Bytes:     []byte("data2"),
	}

	state3 := &State{
		LayerName: "layer1",
		StateName: "state3",
		Bytes:     []byte("data3"),
	}

	fb := NewFileBackend(tempDir + "/testfile.json")

	err := fb.SaveState(context.Background(), state1)
	require.NoError(t, err)

	err = fb.SaveState(context.Background(), state2)
	require.NoError(t, err)

	err = fb.SaveState(context.Background(), state3)
	require.NoError(t, err)

	t.Run("list states for existing layer", func(t *testing.T) {
		states, err := fb.ListStatesByLayer(context.Background(), state1.LayerName)
		require.NoError(t, err)
		assert.Len(t, states, 2)
		assert.Contains(t, states, state1)
		assert.Contains(t, states, state3)
	})

	t.Run("list states for non-existent layer", func(t *testing.T) {
		states, err := fb.ListStatesByLayer(context.Background(), "nonExistentLayer")
		require.NoError(t, err)
		assert.Empty(t, states)
	})
}

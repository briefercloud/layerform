package layerstate

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	storageMock "github.com/ergomake/layerform/mocks/internal_/storage"
)

func TestFileLikeBackend_GetState(t *testing.T) {
	t.Run("state found", func(t *testing.T) {
		state := &State{
			LayerName:         "layer1",
			StateName:         "state1",
			DependenciesState: map[string]string{"base": "testBaseStae"},
			Bytes:             []byte("state1"),
		}

		fb := &fileLikeBackend{
			model: &fileLikeModel{
				Version: 0,
				States:  []*State{state},
			},
		}

		result, err := fb.GetState(context.Background(), state.LayerName, state.StateName)
		require.NoError(t, err)
		assert.Equal(t, state, result)
	})

	t.Run("state not found", func(t *testing.T) {
		fb := fileLikeBackend{
			model: &fileLikeModel{
				States: []*State{
					{LayerName: "layer1", StateName: "state1"},
				},
			},
		}

		_, err := fb.GetState(context.Background(), "layer2", "state2")
		assert.ErrorIs(t, err, ErrStateNotFound)
	})
}

func TestFileLikeBackend_SaveState(t *testing.T) {
	t.Run("adds or update state correctly", func(t *testing.T) {
		state := &State{
			LayerName: "layer1",
			StateName: "state1",
			Bytes:     []byte("data1"),
		}
		storage := storageMock.NewFileLike(t)
		storage.EXPECT().Save(
			context.Background(),
			&fileLikeModel{States: []*State{state}},
		).Return(nil)

		fb := &fileLikeBackend{
			model:   &fileLikeModel{},
			storage: storage,
		}

		err := fb.SaveState(context.Background(), state)
		require.NoError(t, err)

		assert.Len(t, fb.model.States, 1)
		assert.Equal(t, "layer1", fb.model.States[0].LayerName)
		assert.Equal(t, "state1", fb.model.States[0].StateName)
		assert.Equal(t, []byte("data1"), fb.model.States[0].Bytes)
	})

	t.Run("fails when fails to save fileLike", func(t *testing.T) {
		expectedErr := errors.New("rip")

		state := &State{
			LayerName: "layer1",
			StateName: "state1",
			Bytes:     []byte("data1"),
		}
		storage := storageMock.NewFileLike(t)
		storage.EXPECT().Save(
			context.Background(),
			&fileLikeModel{States: []*State{state}},
		).Return(expectedErr)

		fb := &fileLikeBackend{
			model:   &fileLikeModel{},
			storage: storage,
		}

		err := fb.SaveState(context.Background(), state)
		assert.ErrorIs(t, err, expectedErr)

		assert.Len(t, fb.model.States, 1)
		assert.Equal(t, "layer1", fb.model.States[0].LayerName)
		assert.Equal(t, "state1", fb.model.States[0].StateName)
		assert.Equal(t, []byte("data1"), fb.model.States[0].Bytes)
	})
}

func TestFileLikeBackend_DeleteState(t *testing.T) {
	setup := func() *fileLikeBackend {
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

		return &fileLikeBackend{
			model: &fileLikeModel{
				Version: 0,
				States:  []*State{state1, state2},
			},
		}
	}

	t.Run("delete existing state", func(t *testing.T) {
		flb := setup()
		state2 := flb.model.States[1]

		storage := storageMock.NewFileLike(t)
		storage.EXPECT().
			Save(
				mock.Anything,
				&fileLikeModel{States: []*State{flb.model.States[1]}},
			).
			Return(nil)

		flb.storage = storage

		err := flb.DeleteState(context.Background(), "layer1", "state1")
		require.NoError(t, err)

		assert.Len(t, flb.model.States, 1)
		assert.Equal(t, state2, flb.model.States[0])
	})

	t.Run("delete non-existent state", func(t *testing.T) {
		flb := setup()

		storage := storageMock.NewFileLike(t)
		storage.EXPECT().
			Save(mock.Anything, flb.model).
			Return(nil)

		flb.storage = storage

		err := flb.DeleteState(context.Background(), "nonExistentLayer", "nonExistentState")
		assert.NoError(t, err)

		assert.Len(t, flb.model.States, 2)
	})
}

func TestFileLikeBackend_ListStatesByLayer(t *testing.T) {
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

	flb := fileLikeBackend{
		model: &fileLikeModel{
			Version: 0,
			States:  []*State{state1, state2, state3},
		},
	}

	t.Run("list states for existing layer", func(t *testing.T) {
		states, err := flb.ListStatesByLayer(context.Background(), state1.LayerName)
		require.NoError(t, err)
		assert.Len(t, states, 2)
		assert.Contains(t, states, state1)
		assert.Contains(t, states, state3)
	})

	t.Run("list states for non-existent layer", func(t *testing.T) {
		states, err := flb.ListStatesByLayer(context.Background(), "nonExistentLayer")
		require.NoError(t, err)
		assert.Empty(t, states)
	})
}

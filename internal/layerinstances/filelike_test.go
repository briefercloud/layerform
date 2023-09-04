package layerinstances

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	storageMock "github.com/ergomake/layerform/mocks/internal_/storage"
	"github.com/ergomake/layerform/pkg/data"
)

func TestFileLikeModelUnmarshalJSON(t *testing.T) {
	t.Run("support v0", func(t *testing.T) {
		v0 := &fileLikeModelV0{
			Version: 0,
			States: []*data.InstanceV0{{
				LayerSHA:          []byte("layerSHA"),
				LayerName:         "layer1",
				StateName:         "instance1",
				DependenciesState: map[string]string{"layer2": "instance2"},
				Bytes:             []byte("some bytes"),
				Status:            data.InstanceStatusAlive,
			}},
		}

		v0b, err := json.Marshal(v0)
		require.NoError(t, err)

		var flm fileLikeModel
		err = json.Unmarshal(v0b, &flm)
		require.NoError(t, err)

		expected := fileLikeModel{
			Version: CURRENT_FILE_LIKE_MODEL_VERSION,
			Instances: []*data.Instance{{
				DefinitionSHA:        []byte("layerSHA"),
				DefinitionName:       "layer1",
				InstanceName:         "instance1",
				DependenciesInstance: map[string]string{"layer2": "instance2"},
				Bytes:                []byte("some bytes"),
				Status:               data.InstanceStatusAlive,
				Version:              data.CURRENT_INSTANCE_VERSION,
			}},
		}

		assert.Equal(t, expected, flm)
	})
}

func TestFileLikeBackend_GetInstance(t *testing.T) {
	t.Run("instance found", func(t *testing.T) {
		instance := &data.Instance{
			DefinitionName:       "layer1",
			InstanceName:         "instance1",
			DependenciesInstance: map[string]string{"base": "testBaseStae"},
			Bytes:                []byte("instance1"),
		}

		fb := &fileLikeBackend{
			model: &fileLikeModel{
				Version:   CURRENT_FILE_LIKE_MODEL_VERSION,
				Instances: []*data.Instance{instance},
			},
		}

		result, err := fb.GetInstance(context.Background(), instance.DefinitionName, instance.InstanceName)
		require.NoError(t, err)
		assert.Equal(t, instance, result)
	})

	t.Run("instance not found", func(t *testing.T) {
		fb := fileLikeBackend{
			model: &fileLikeModel{
				Instances: []*data.Instance{
					{DefinitionName: "layer1", InstanceName: "instance1"},
				},
			},
		}

		_, err := fb.GetInstance(context.Background(), "layer2", "instance2")
		assert.ErrorIs(t, err, ErrInstanceNotFound)
	})
}

func TestFileLikeBackend_SaveInstance(t *testing.T) {
	t.Run("adds or update instance correctly", func(t *testing.T) {
		instance := &data.Instance{
			DefinitionName: "layer1",
			InstanceName:   "instance1",
			Bytes:          []byte("data1"),
		}
		storage := storageMock.NewFileLike(t)
		storage.EXPECT().Save(
			context.Background(),
			&fileLikeModel{Version: CURRENT_FILE_LIKE_MODEL_VERSION, Instances: []*data.Instance{instance}},
		).Return(nil)

		fb := &fileLikeBackend{
			model:   &fileLikeModel{Version: CURRENT_FILE_LIKE_MODEL_VERSION},
			storage: storage,
		}

		err := fb.SaveInstance(context.Background(), instance)
		require.NoError(t, err)

		assert.Len(t, fb.model.Instances, 1)
		assert.Equal(t, "layer1", fb.model.Instances[0].DefinitionName)
		assert.Equal(t, "instance1", fb.model.Instances[0].InstanceName)
		assert.Equal(t, []byte("data1"), fb.model.Instances[0].Bytes)
	})

	t.Run("fails when fails to save fileLike", func(t *testing.T) {
		expectedErr := errors.New("rip")

		instance := &data.Instance{
			DefinitionName: "layer1",
			InstanceName:   "instance1",
			Bytes:          []byte("data1"),
		}
		storage := storageMock.NewFileLike(t)
		storage.EXPECT().Save(
			context.Background(),
			&fileLikeModel{Version: CURRENT_FILE_LIKE_MODEL_VERSION, Instances: []*data.Instance{instance}},
		).Return(expectedErr)

		fb := &fileLikeBackend{
			model:   &fileLikeModel{Version: CURRENT_FILE_LIKE_MODEL_VERSION},
			storage: storage,
		}

		err := fb.SaveInstance(context.Background(), instance)
		assert.ErrorIs(t, err, expectedErr)

		assert.Len(t, fb.model.Instances, 1)
		assert.Equal(t, "layer1", fb.model.Instances[0].DefinitionName)
		assert.Equal(t, "instance1", fb.model.Instances[0].InstanceName)
		assert.Equal(t, []byte("data1"), fb.model.Instances[0].Bytes)
	})
}

func TestFileLikeBackend_DeleteInstance(t *testing.T) {
	setup := func() *fileLikeBackend {
		instance1 := &data.Instance{
			DefinitionName: "layer1",
			InstanceName:   "instance1",
			Bytes:          []byte("data1"),
		}

		instance2 := &data.Instance{
			DefinitionName: "layer2",
			InstanceName:   "instance2",
			Bytes:          []byte("data2"),
		}

		return &fileLikeBackend{
			model: &fileLikeModel{
				Version:   CURRENT_FILE_LIKE_MODEL_VERSION,
				Instances: []*data.Instance{instance1, instance2},
			},
		}
	}

	t.Run("delete existing instance", func(t *testing.T) {
		flb := setup()
		instance2 := flb.model.Instances[1]

		storage := storageMock.NewFileLike(t)
		storage.EXPECT().
			Save(
				mock.Anything,
				&fileLikeModel{Version: CURRENT_FILE_LIKE_MODEL_VERSION, Instances: []*data.Instance{flb.model.Instances[1]}},
			).
			Return(nil)

		flb.storage = storage

		err := flb.DeleteInstance(context.Background(), "layer1", "instance1")
		require.NoError(t, err)

		assert.Len(t, flb.model.Instances, 1)
		assert.Equal(t, instance2, flb.model.Instances[0])
	})

	t.Run("delete non-existent instance", func(t *testing.T) {
		flb := setup()

		storage := storageMock.NewFileLike(t)
		storage.EXPECT().
			Save(mock.Anything, flb.model).
			Return(nil)

		flb.storage = storage

		err := flb.DeleteInstance(context.Background(), "nonExistentLayer", "nonExistentInstance")
		assert.NoError(t, err)

		assert.Len(t, flb.model.Instances, 2)
	})
}

func TestFileLikeBackend_ListInstancesByLayer(t *testing.T) {
	instance1 := &data.Instance{
		DefinitionName: "layer1",
		InstanceName:   "instance1",
		Bytes:          []byte("data1"),
	}

	instance2 := &data.Instance{
		DefinitionName: "layer2",
		InstanceName:   "instance2",
		Bytes:          []byte("data2"),
	}

	instance3 := &data.Instance{
		DefinitionName: "layer1",
		InstanceName:   "instance3",
		Bytes:          []byte("data3"),
	}

	flb := fileLikeBackend{
		model: &fileLikeModel{
			Version:   CURRENT_FILE_LIKE_MODEL_VERSION,
			Instances: []*data.Instance{instance1, instance2, instance3},
		},
	}

	t.Run("list instances for existing layer", func(t *testing.T) {
		instances, err := flb.ListInstancesByLayer(context.Background(), instance1.DefinitionName)
		require.NoError(t, err)
		assert.Len(t, instances, 2)
		assert.Contains(t, instances, instance1)
		assert.Contains(t, instances, instance3)
	})

	t.Run("list instances for non-existent layer", func(t *testing.T) {
		instances, err := flb.ListInstancesByLayer(context.Background(), "nonExistentLayer")
		require.NoError(t, err)
		assert.Empty(t, instances)
	})
}

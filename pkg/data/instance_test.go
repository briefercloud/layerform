package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSON(t *testing.T) {
	t.Run("support v0", func(t *testing.T) {
		v0 := &InstanceV0{
			LayerSHA:          []byte("layerSHA"),
			LayerName:         "layer1",
			StateName:         "instance1",
			DependenciesState: map[string]string{"layer0": "instance1"},
			Bytes:             []byte("some bytes"),
			Status:            InstanceStatusAlive,
		}

		v0b, err := json.Marshal(v0)
		require.NoError(t, err)

		var instance Instance
		err = json.Unmarshal(v0b, &instance)
		require.NoError(t, err)

		expected := Instance{
			DefinitionSHA:        []byte("layerSHA"),
			DefinitionName:       "layer1",
			InstanceName:         "instance1",
			DependenciesInstance: map[string]string{"layer0": "instance1"},
			Bytes:                []byte("some bytes"),
			Status:               InstanceStatusAlive,
			Version:              CURRENT_INSTANCE_VERSION,
		}
		assert.Equal(t, expected, instance)
	})
}

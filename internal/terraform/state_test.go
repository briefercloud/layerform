package terraform

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_ResourceDiff(t *testing.T) {
	a := &State{
		Resources: []StateResource{
			{Module: "module1", Type: "type1", Name: "resource1"},
			{Module: "module2", Type: "type2", Name: "resource2"},
		},
	}

	b := &State{
		Resources: []StateResource{
			{Module: "module2", Type: "type2", Name: "resource2"},
			{Module: "module3", Type: "type3", Name: "resource3"},
		},
	}

	diff := a.ResourceDiff(b)
	assert.Len(t, diff, 1)
	assert.Equal(t, StateResource{Module: "module3", Type: "type3", Name: "resource3"}, diff[0])
}

func TestStateResource_Equal(t *testing.T) {
	res1 := StateResource{Module: "module1", Type: "type1", Name: "resource1"}
	res2 := StateResource{Module: "module1", Type: "type1", Name: "resource1"}
	res3 := StateResource{Module: "module1", Type: "type1", Name: "resource2"}

	assert.True(t, res1.Equal(&res2))
	assert.False(t, res1.Equal(&res3))
}

func TestStateResource_Address(t *testing.T) {
	res := StateResource{Module: "module1", Type: "type1", Name: "resource1"}
	assert.Equal(t, "module1.type1.resource1", res.Address())
}

func TestTFStateFromBytes(t *testing.T) {
	data := []byte(`{"resources": [{"module": "module1", "type": "type1", "name": "resource1"}]}`)

	state, err := TFStateFromBytes(data)
	require.NoError(t, err)
	assert.Len(t, state.Resources, 1)
	assert.Equal(t, StateResource{Module: "module1", Type: "type1", Name: "resource1"}, state.Resources[0])
	assert.Equal(t, data, state.Bytes)
}

func TestTFStateFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_tf_state.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	data := []byte(`{"resources": [{"module": "module1", "type": "type1", "name": "resource1"}]}`)
	err = os.WriteFile(tmpFile.Name(), data, 0644)
	require.NoError(t, err)

	state, err := TFStateFromFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Len(t, state.Resources, 1)
	assert.Equal(t, StateResource{Module: "module1", Type: "type1", Name: "resource1"}, state.Resources[0])
	assert.Equal(t, data, state.Bytes)
}

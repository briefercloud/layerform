package state

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
)

const version = 1

type fileContent struct {
	Version uint                         `json:"version"`
	State   map[string]map[string][]byte `json:"state"`
}

type fileBackend struct {
	state    *fileContent
	filePath string
	lock     sync.Mutex
}

var _ Backend = &fileBackend{}

// NewFileBackend creates a StateBackend backed by a file
func NewFileBackend(filePath string) (*fileBackend, error) {
	fileContent := &fileContent{
		Version: version,
		State:   make(map[string]map[string][]byte),
	}

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		data, err := json.Marshal(fileContent)
		if err != nil {
			return nil, errors.Wrap(err, "fail to marshal state to file")
		}

		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to write state to %s", filePath)
		}
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to read state out of %s", filePath)
		}

		err = json.Unmarshal(data, fileContent)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to parse state out of %s", filePath)
		}
	}

	return &fileBackend{
		state:    fileContent,
		filePath: filePath,
	}, nil
}

// GetLayerState fetches a layer state for a instance, return nil when instance has no state
func (c *fileBackend) GetLayerState(layer *model.Layer, instance string) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	layerState := c.state.State[layer.Name]
	if layerState == nil {
		return nil, nil
	}

	state := layerState[instance]
	return state, nil
}

// SaveLayerState saves a layer state for a instance
func (c *fileBackend) SaveLayerState(layer *model.Layer, instance string, state []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	layerState := c.state.State[layer.Name]
	if layerState == nil {
		layerState = make(map[string][]byte)
	}

	layerState[instance] = state

	c.state.State[layer.Name] = layerState

	data, err := json.Marshal(c.state)
	if err != nil {
		return errors.Wrap(err, "fail to marshal state to json")
	}

	err = os.WriteFile(c.filePath, data, 0644)
	return errors.Wrapf(err, "fail to write state to file %s", c.filePath)
}

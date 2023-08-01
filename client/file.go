package client

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
)

type fileContent struct {
	Layers map[string]*model.Layer      `json:"layers"`
	State  map[string]map[string][]byte `json:"state"`
}

type fileClient struct {
	content  *fileContent
	filePath string
	lock     sync.Mutex
}

var _ Client = &fileClient{}

// NewFileClient creates a LayerformClient backed by a file
func NewFileClient(filePath string) (*fileClient, error) {
	fileContent := &fileContent{
		Layers: make(map[string]*model.Layer),
		State:  make(map[string]map[string][]byte),
	}

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		data, err := json.Marshal(fileContent)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return nil, err
		}
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, fileContent)
		if err != nil {
			return nil, err
		}
	}

	return &fileClient{
		content:  fileContent,
		filePath: filePath,
	}, nil
}

// ReadLayer fetches a layer from the file backend
func (c *fileClient) GetLayer(name string) (*model.Layer, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.content.Layers[name], nil
}

// CreateLayer creates a layer in the file backend
func (c *fileClient) CreateLayer(layer *model.Layer) (*model.Layer, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.content.Layers[layer.Name]; ok {
		return nil, errors.Wrapf(ErrLayerAlreadyExists, "layer name %s already exists", layer.Name)
	}

	c.content.Layers[layer.Name] = layer

  err := c.commit()

  return layer, err
}

// GetLayerState fetches a layer state for a instance, return nil when instance has no state
func (c *fileClient) GetLayerState(layer *model.Layer, instance string) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	layerState := c.content.State[layer.Name]
	if layerState == nil {
		return nil, nil
	}

	state := layerState[instance]
	return state, nil
}

// SaveLayerState saves a layer state for a instance
func (c *fileClient) SaveLayerState(layer *model.Layer, instance string, state []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	layerState := c.content.State[layer.Name]
	if layerState == nil {
		layerState = make(map[string][]byte)
	}

	layerState[instance] = state

	c.content.State[layer.Name] = layerState

	return c.commit()
}

func (c *fileClient) commit() error {
	data, err := json.Marshal(c.content)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.filePath, data, 0644)
	return err
}

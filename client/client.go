package client

import (
	"errors"

	"github.com/ergomake/layerform/internal/data/model"
)

var ErrLayerAlreadyExists = errors.New("layer already exists")

type Client interface {
	CreateLayer(*model.Layer) (*model.Layer, error)
	GetLayer(name string) (*model.Layer, error)
	GetLayerState(layer *model.Layer, instance string) ([]byte, error)
	SaveLayerState(layer *model.Layer, instance string, state []byte) error
}

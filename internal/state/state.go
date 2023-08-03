package state

import "github.com/ergomake/layerform/internal/data/model"

type Backend interface {
	GetLayerState(layer *model.Layer, instance string) ([]byte, error)
	SaveLayerState(layer *model.Layer, instance string, state []byte) error
}

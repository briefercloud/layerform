package layers

import "github.com/ergomake/layerform/internal/data/model"

type Backend interface {
	GetLayer(name string) (*model.Layer, error)
	ResolveDependencies(*model.Layer) ([]*model.Layer, error)
}

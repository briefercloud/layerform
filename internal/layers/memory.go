package layers

import "github.com/ergomake/layerform/internal/data/model"

type inMemoryBackend struct {
	layers map[string]*model.Layer
}

var _ Backend = &inMemoryBackend{}

func NewInMemoryBackend(list []*model.Layer) *inMemoryBackend {
	layers := make(map[string]*model.Layer)
	for _, l := range list {
		layers[l.Name] = l
	}

	return &inMemoryBackend{layers}
}

func (mb *inMemoryBackend) GetLayer(name string) (*model.Layer, error) {
	return mb.layers[name], nil

}

package layers

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
)

var ErrNotFound = errors.New("layer not found")

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

func (mb *inMemoryBackend) GetLayer(ctx context.Context, name string) (*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Getting layer", "layer", name)
	return mb.layers[name], nil
}

func (mb *inMemoryBackend) ResolveDependencies(ctx context.Context, layer *model.Layer) ([]*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Resolving layer dependencies", "layer", layer.Name)
	layers := make([]*model.Layer, len(layer.Dependencies))
	for i, d := range layer.Dependencies {
		depLayer, err := mb.GetLayer(ctx, d)
		if err != nil {
			// this never happens for inMemoryBackend btw
			return nil, errors.Wrapf(err, "fail to get dependency \"%s\" of layer \"%s\"", d, layer.Name)
		}

		if depLayer == nil {
			return nil, errors.Wrapf(ErrNotFound, "dependency \"%s\" of layer \"%s\" not found", d, layer.Name)
		}

		layers[i] = depLayer
	}

	return layers, nil
}

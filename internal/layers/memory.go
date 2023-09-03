package layers

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
)

type inMemoryBackend struct {
	layers map[string]*model.Layer
}

var _ Backend = &inMemoryBackend{}

func NewInMemoryBackend(layersArr []*model.Layer) *inMemoryBackend {
	layers := map[string]*model.Layer{}
	for _, l := range layersArr {
		layers[l.Name] = l
	}

	return &inMemoryBackend{layers}
}

func (imb *inMemoryBackend) GetLayer(ctx context.Context, name string) (*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Getting layer", "layer", name)

	return imb.layers[name], nil
}

func (imb *inMemoryBackend) ResolveDependencies(ctx context.Context, layer *model.Layer) ([]*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Resolving layer dependencies", "layer", layer.Name)
	layers := make([]*model.Layer, len(layer.Dependencies))
	for i, d := range layer.Dependencies {
		depLayer, err := imb.GetLayer(ctx, d)
		if err != nil {
			// this never happens for in memory backend btw
			return nil, errors.Wrapf(err, "fail to get dependency \"%s\" of layer \"%s\"", d, layer.Name)
		}

		if depLayer == nil {
			return nil, errors.Wrapf(ErrNotFound, "dependency \"%s\" of layer \"%s\" not found", d, layer.Name)
		}

		layers[i] = depLayer
	}

	return layers, nil
}

func (imb *inMemoryBackend) ListLayers(ctx context.Context) ([]*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Listing layers")
	layers := make([]*model.Layer, 0)
	for _, l := range imb.layers {
		layers = append(layers, l)
	}

	return layers, nil
}

func (imb *inMemoryBackend) UpdateLayers(ctx context.Context, layers []*model.Layer) error {
	hclog.FromContext(ctx).Debug("Updating layers")

	imb.layers = make(map[string]*model.Layer)
	for _, l := range layers {
		imb.layers[l.Name] = l
	}

	return nil
}

func (imb *inMemoryBackend) Location(ctx context.Context) (string, error) {
	return "memory", nil
}

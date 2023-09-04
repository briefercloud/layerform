package layerdefinitions

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/storage"
	"github.com/ergomake/layerform/pkg/data"
)

const bloblayersVersion = 0

type fileLikeModel struct {
	Version uint                        `json:"version"`
	Layers  map[string]*data.Definition `json:"layers"`
}

type fileLikeBackend struct {
	data    *fileLikeModel
	storage storage.FileLike
}

var _ Backend = &fileLikeBackend{}

func NewFileLikeBackend(ctx context.Context, storage storage.FileLike) (*fileLikeBackend, error) {
	filelayers := fileLikeModel{
		Version: bloblayersVersion,
	}
	err := storage.Load(ctx, &filelayers)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read file")
	}

	return &fileLikeBackend{data: &filelayers, storage: storage}, nil
}

func (flb *fileLikeBackend) GetLayer(ctx context.Context, name string) (*data.Definition, error) {
	hclog.FromContext(ctx).Debug("Getting layer", "layer", name)

	layer, ok := flb.data.Layers[name]
	if !ok {
		return nil, errors.Wrapf(ErrNotFound, "fail to get layer %s", name)
	}

	return layer, nil
}

func (flb *fileLikeBackend) ResolveDependencies(ctx context.Context, layer *data.Definition) ([]*data.Definition, error) {
	hclog.FromContext(ctx).Debug("Resolving layer dependencies", "layer", layer.Name)
	layers := make([]*data.Definition, len(layer.Dependencies))
	for i, d := range layer.Dependencies {
		depLayer, err := flb.GetLayer(ctx, d)
		if err != nil {
			// this never happens for fileBackend btw
			return nil, errors.Wrapf(err, "fail to get dependency \"%s\" of layer \"%s\"", d, layer.Name)
		}

		if depLayer == nil {
			return nil, errors.Wrapf(ErrNotFound, "dependency \"%s\" of layer \"%s\" not found", d, layer.Name)
		}

		layers[i] = depLayer
	}

	return layers, nil
}

func (flb *fileLikeBackend) ListLayers(ctx context.Context) ([]*data.Definition, error) {
	hclog.FromContext(ctx).Debug("Listing layers")
	layers := make([]*data.Definition, 0)
	for _, l := range flb.data.Layers {
		layers = append(layers, l)
	}

	return layers, nil
}

func (flb *fileLikeBackend) UpdateLayers(ctx context.Context, layers []*data.Definition) error {
	hclog.FromContext(ctx).Debug("Updating layers")

	flb.data.Layers = make(map[string]*data.Definition)
	for _, l := range layers {
		flb.data.Layers[l.Name] = l
	}

	return flb.storage.Save(ctx, flb.data)
}

func (flb *fileLikeBackend) Location(ctx context.Context) (string, error) {
	return flb.storage.Path(ctx)
}

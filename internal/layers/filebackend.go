package layers

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
)

var ErrNotFound = errors.New("layer not found")

const filelayersVersion = 0

type filelayers struct {
	Version uint                    `json:"version"`
	Layers  map[string]*model.Layer `json:"layers"`
}

type filebackend struct {
	fpath string

	filelayers *filelayers
}

var _ Backend = &filebackend{}

func readFile(ctx context.Context, fpath string) (*filelayers, error) {
	hclog.FromContext(ctx).Debug("Reading layers file", "path", fpath)

	raw, err := os.ReadFile(fpath)
	if errors.Is(err, os.ErrNotExist) {
		return &filelayers{Version: filelayersVersion}, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "fail to read %s", fpath)
	}

	var fstate filelayers
	err = json.Unmarshal(raw, &fstate)

	return &fstate, errors.Wrapf(err, "fail to parse layers out of %s", fpath)
}

func NewFileBackend(ctx context.Context, fpath string) (*filebackend, error) {
	filelayers, err := readFile(ctx, fpath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read file")
	}

	return &filebackend{fpath: fpath, filelayers: filelayers}, nil
}

func (fb *filebackend) GetLayer(ctx context.Context, name string) (*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Getting layer", "layer", name)

	layer, ok := fb.filelayers.Layers[name]
	if !ok {
		return nil, errors.Wrapf(ErrNotFound, "fail to get layer %s", name)
	}

	return layer, nil
}

func (fb *filebackend) ResolveDependencies(ctx context.Context, layer *model.Layer) ([]*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Resolving layer dependencies", "layer", layer.Name)
	layers := make([]*model.Layer, len(layer.Dependencies))
	for i, d := range layer.Dependencies {
		depLayer, err := fb.GetLayer(ctx, d)
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

func (fb *filebackend) ListLayers(ctx context.Context) ([]*model.Layer, error) {
	hclog.FromContext(ctx).Debug("Listing layers")
	layers := make([]*model.Layer, 0)
	for _, l := range fb.filelayers.Layers {
		layers = append(layers, l)
	}

	return layers, nil
}

func (fb *filebackend) UpdateLayers(ctx context.Context, layers []*model.Layer) error {
	hclog.FromContext(ctx).Debug("Updating layers")

	fb.filelayers.Layers = make(map[string]*model.Layer)
	for _, l := range layers {
		fb.filelayers.Layers[l.Name] = l
	}

	return fb.writeFile(ctx)
}

func (fb *filebackend) writeFile(ctx context.Context) error {
	hclog.FromContext(ctx).Debug("Writting layers to file", "path", fb.fpath)

	data, err := json.Marshal(fb.filelayers)
	if err != nil {
		return errors.Wrap(err, "fail to marshal filelayers")
	}

	err = os.WriteFile(fb.fpath, data, 0644)
	return errors.Wrap(err, "fail to write file")
}

package layers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
)

var ErrNotFound = errors.New("layer not found")

type Backend interface {
	ListLayers(ctx context.Context) ([]*data.Layer, error)
	GetLayer(ctx context.Context, name string) (*data.Layer, error)
	ResolveDependencies(ctx context.Context, layer *data.Layer) ([]*data.Layer, error)
	UpdateLayers(ctx context.Context, layers []*data.Layer) error
	Location(ctx context.Context) (string, error)
}

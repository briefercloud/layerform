package layers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
)

var ErrNotFound = errors.New("layer not found")

type Backend interface {
	ListLayers(ctx context.Context) ([]*model.Layer, error)
	GetLayer(ctx context.Context, name string) (*model.Layer, error)
	ResolveDependencies(ctx context.Context, layer *model.Layer) ([]*model.Layer, error)
	UpdateLayers(ctx context.Context, layers []*model.Layer) error
	Location(ctx context.Context) (string, error)
}

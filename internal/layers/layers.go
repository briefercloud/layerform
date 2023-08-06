package layers

import (
	"context"

	"github.com/ergomake/layerform/internal/data/model"
)

type Backend interface {
	GetLayer(ctx context.Context, name string) (*model.Layer, error)
	ResolveDependencies(ctx context.Context, layer *model.Layer) ([]*model.Layer, error)
}

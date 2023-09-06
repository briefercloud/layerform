package layerdefinitions

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
)

var ErrNotFound = errors.New("layer not found")

type Backend interface {
	ListLayers(ctx context.Context) ([]*data.LayerDefinition, error)
	GetLayer(ctx context.Context, name string) (*data.LayerDefinition, error)
	ResolveDependencies(ctx context.Context, layer *data.LayerDefinition) ([]*data.LayerDefinition, error)
	UpdateLayers(ctx context.Context, layers []*data.LayerDefinition) error
	Location(ctx context.Context) (string, error)
}

package layerdefinitions

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
)

var ErrNotFound = errors.New("layer not found")

type Backend interface {
	ListLayers(ctx context.Context) ([]*data.Definition, error)
	GetLayer(ctx context.Context, name string) (*data.Definition, error)
	ResolveDependencies(ctx context.Context, layer *data.Definition) ([]*data.Definition, error)
	UpdateLayers(ctx context.Context, layers []*data.Definition) error
	Location(ctx context.Context) (string, error)
}

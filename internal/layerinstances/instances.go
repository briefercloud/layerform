package layerinstances

import (
	"context"
	"errors"

	"github.com/ergomake/layerform/pkg/data"
)

var ErrInstanceNotFound = errors.New("instance not found")

type Backend interface {
	GetInstance(ctx context.Context, layerName, instanceName string) (*data.LayerInstance, error)
	ListInstancesByLayer(ctx context.Context, layerName string) ([]*data.LayerInstance, error)
	SaveInstance(ctx context.Context, instance *data.LayerInstance) error
	DeleteInstance(ctx context.Context, layerName, instanceName string) error
	ListInstances(ctx context.Context) ([]*data.LayerInstance, error)
}

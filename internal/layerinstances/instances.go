package layerinstances

import (
	"context"
	"errors"

	"github.com/ergomake/layerform/pkg/data"
)

var ErrInstanceNotFound = errors.New("instance not found")

type Backend interface {
	GetInstance(ctx context.Context, layerName, instanceName string) (*data.Instance, error)
	ListInstancesByLayer(ctx context.Context, layerName string) ([]*data.Instance, error)
	SaveInstance(ctx context.Context, instance *data.Instance) error
	DeleteInstance(ctx context.Context, layerName, instanceName string) error
	ListInstances(ctx context.Context) ([]*data.Instance, error)
}

package envvars

import (
	"context"

	"github.com/ergomake/layerform/pkg/data"
)

type Backend interface {
	ListVariables(ctx context.Context) ([]*data.EnvVar, error)
	SaveVariable(ctx context.Context, variable *data.EnvVar) error
}

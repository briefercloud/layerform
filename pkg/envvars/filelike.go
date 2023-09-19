package envvars

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/storage"
	"github.com/ergomake/layerform/pkg/data"
)

type fileLikeBackend struct {
	variables []*data.EnvVar
	storage   storage.FileLike
}

var _ Backend = &fileLikeBackend{}

func NewFileLikeBackend(ctx context.Context, storage storage.FileLike) (*fileLikeBackend, error) {
	variables := make([]*data.EnvVar, 0)
	err := storage.Load(ctx, &variables)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read file")
	}

	return &fileLikeBackend{variables, storage}, nil
}

func (flb *fileLikeBackend) ListVariables(ctx context.Context) ([]*data.EnvVar, error) {
	return flb.variables, nil
}

func (flb *fileLikeBackend) SaveVariable(ctx context.Context, variable *data.EnvVar) error {
	for i, v := range flb.variables {
		if v.Name == variable.Name {
			flb.variables[i] = variable
			return flb.storage.Save(ctx, flb.variables)
		}
	}

	flb.variables = append(flb.variables, variable)
	return flb.storage.Save(ctx, flb.variables)
}

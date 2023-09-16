package command

import (
	"context"

	"github.com/ergomake/layerform/pkg/data"
	"github.com/ergomake/layerform/pkg/envvars"
)

type setenvCommand struct {
	backend envvars.Backend
}

func NewSetEnv(backend envvars.Backend) *setenvCommand {
	return &setenvCommand{backend}
}

func (c *setenvCommand) Run(ctx context.Context, variable *data.EnvVar) error {
	return c.backend.SaveVariable(ctx, variable)
}

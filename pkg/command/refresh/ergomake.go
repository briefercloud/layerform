package refresh

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/layerinstances"
)

type ergomakeRefreshCommand struct {
	baseURL          string
	instancesBackend layerinstances.Backend
}

var _ Refresh = &ergomakeRefreshCommand{}

func NewErgomake(baseURL string) *ergomakeRefreshCommand {
	instancesBackend := layerinstances.NewErgomake(baseURL)

	return &ergomakeRefreshCommand{baseURL, instancesBackend}
}

func (e *ergomakeRefreshCommand) Run(
	ctx context.Context,
	definitionName, instanceName string,
	vars []string,
) error {
	return errors.New("not implemented yet")
}

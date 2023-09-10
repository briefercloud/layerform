package spawn

import (
	"context"
)

type Spawn interface {
	Run(
		ctx context.Context,
		definitionName, instanceName string,
		dependenciesInstance map[string]string,
		vars []string,
	) error
}

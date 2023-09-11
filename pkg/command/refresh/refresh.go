package refresh

import (
	"context"
)

type Refresh interface {
	Run(
		ctx context.Context,
		definitionName, instanceName string,
		vars []string,
	) error
}

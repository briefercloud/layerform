package kill

import (
	"context"
)

type Kill interface {
	Run(ctx context.Context, definitionName, instanceName string, autoApprove bool, vars []string, force bool) error
}

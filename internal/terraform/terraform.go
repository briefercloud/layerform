package terraform

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
)

func GetTFPath(ctx context.Context) (string, error) {
	logger := hclog.FromContext(ctx)

	logger.Debug("Geting tfpath")

	i := install.NewInstaller()
	i.SetLogger(logger.StandardLogger(&hclog.StandardLoggerOptions{
		ForceLevel: hclog.Debug,
	}))
	tfpath, err := i.Ensure(ctx, []src.Source{
		&fs.Version{
			Product:     product.Terraform,
			Constraints: version.MustConstraints(version.NewConstraint("<=1.5.5")),
		},
	})

	return tfpath, err
}

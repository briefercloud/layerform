package cli

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	configureCmd.Flags().String("file", "layerform.json", "the configuration file with layer definitions")
	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "transforms layer definition configurations in an actual layer definition file",
	Long: `Transforms layer definition configurations in an actual layer definition file.

This command is temporary. It will eventually be replaced by a Terraform provider.

Here's an example layer definition configurations:

{
  "layers": [
    {
      "name": "eks",
      "files": [
        "layers/eks.tf",
        "layers/eks/main.tf",
        "layers/eks/output.tf"
      ]
    },
    {
      "name"  : "kibana",
      "files": [
        "layers/kibana.tf",
        "layers/kibana/main.tf",
        "layers/kibana/output.tf",
        "layers/kibana/variables.tf"
      ],
      "dependencies": [
        "eks"
      ]
    }
  ]
}
`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		logger := hclog.Default()
		logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
		if logLevel != hclog.NoLevel {
			logger.SetLevel(logLevel)
		}
		ctx := hclog.WithContext(context.Background(), logger)

		cfg, err := lfconfig.Load("")
		if err != nil {
			return errors.Wrap(err, "fail to load config")
		}

		fpath, err := cmd.Flags().GetString("file")
		if err != nil {
			return errors.Wrap(err, "fail to get --file flag, this is a bug in layerform")
		}

		layerfile, err := layerfile.FromFile(fpath)
		if err != nil {
			return errors.Wrap(err, "fail to read layerform layers definitions from file")
		}

		layers, err := layerfile.ToLayers()
		if err != nil {
			return errors.Wrap(err, "fail to load layers from layerform layers definitions file")
		}

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend")
		}

		err = layersBackend.UpdateLayers(ctx, layers)
		return errors.Wrap(err, "fail to update layers")
	},
}

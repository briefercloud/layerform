package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
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

		s := spinner.New(
			spinner.CharSets[14],
			60*time.Millisecond,
			spinner.WithWriter(os.Stdout),
			spinner.WithSuffix(
				fmt.Sprintf(
					" Loading layer definitions from \"%s\"\n",
					fpath,
				),
			),
		)
		s.Start()

		layerfile, err := layerfile.FromFile(fpath)
		if err != nil {
			return errors.Wrap(err, "fail to read layerform layers definitions from file")
		}

		layers, err := layerfile.ToLayers()
		if err != nil {
			return errors.Wrap(err, "fail to load layers from layerform layers definitions file")
		}

		if len(layers) == 0 {
			s.Stop()
			return errors.Errorf("No layers are defined at \"%s\"\n", fpath)
		}

		s.FinalMSG = fmt.Sprintf(
			"✓ %d %s loaded from \"%s\"\n",
			len(layers),
			pluralize("layer", len(layers)),
			fpath,
		)
		s.Stop()

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend")
		}

		s = spinner.New(
			spinner.CharSets[14],
			60*time.Millisecond,
			spinner.WithWriter(os.Stdout),
			spinner.WithSuffix(" Saving layer definitions\n"),
		)
		s.Start()

		location, err := layersBackend.Location(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend location")
		}

		err = layersBackend.UpdateLayers(ctx, layers)
		if err != nil {
			return errors.Wrap(err, "fail to update layers")
		}

		s.FinalMSG = fmt.Sprintf(
			"✓ %d %s saved to \"%s\"\n",
			len(layers),
			pluralize("layer", len(layers)),
			location,
		)
		s.Stop()

		return nil
	},
}

func pluralize(s string, n int) string {
	if n == 1 {
		return s
	}

	return s + "s"
}

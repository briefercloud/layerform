package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/lfconfig"
	"github.com/ergomake/layerform/pkg/command"
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
}`,
	Run: func(cmd *cobra.Command, _ []string) {
		logger := hclog.Default()
		logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
		if logLevel != hclog.NoLevel {
			logger.SetLevel(logLevel)
		}
		ctx := hclog.WithContext(context.Background(), logger)

		cfg, err := lfconfig.Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to load config"))
			os.Exit(1)
			return
		}

		fpath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --file flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}

		layersBackend, err := cfg.GetDefinitionsBackend(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get layers backend"))
			os.Exit(1)
			return
		}

		instancesBackend, err := cfg.GetInstancesBackend(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get instance backend"))
			os.Exit(1)
			return
		}

		configure := command.NewConfigure(layersBackend, instancesBackend)

		err = configure.Run(ctx, fpath)
		if err != nil {
			if errors.Is(err, layerfile.ErrInvalidDefinitionName) {
				fmt.Fprintln(
					os.Stderr,
					"Name must start and end with an alphanumeric character and can include dashes and underscores in between.",
				)
			}

			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

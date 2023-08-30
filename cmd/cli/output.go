package cli

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	rootCmd.AddCommand(outputCmd)
}

var outputCmd = &cobra.Command{
	Use:   "output <layer> <instance>",
	Args:  cobra.MinimumNArgs(2),
	Short: "reads all output variables from the provided layer instance",
	Long:  `The output command reads all output variables from the given layer instance and prints them as json to standard output.`,
	RunE: func(_ *cobra.Command, args []string) error {
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

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend")
		}

		statesBackend, err := cfg.GetStateBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get state backend")
		}

		layerName := args[0]
		stateName := args[1]

		output := command.NewOutput(layersBackend, statesBackend)

		return output.Run(ctx, layerName, stateName)
	},
}

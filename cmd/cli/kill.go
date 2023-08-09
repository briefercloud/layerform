package cli

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/lithammer/shortuuid/v3"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	// TODO: :bike: fill usage of --var flag for the kill command
	killCmd.Flags().StringArray("var", []string{}, "usage of var flag")

	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use: "kill",
	// TODO: :bike: fill short description of kill command
	Short: "kill short help text",
	// TODO: :bike: fill long description of kill command
	Long: "kill long help text",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		vars, err := cmd.Flags().GetStringArray("var")
		if err != nil {
			return errors.Wrap(err, "fail to get --var flag, this is a bug in layerform")
		}

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend")
		}

		statesBackend := cfg.GetStateBackend()

		layerName := args[0]
		stateName := shortuuid.New()
		if len(args) > 1 {
			stateName = args[1]
		}

		kill := command.NewKill(layersBackend, statesBackend)

		return kill.Run(ctx, layerName, stateName, vars)
	},
}

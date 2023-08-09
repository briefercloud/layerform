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
	// TODO: :bike: fill usage of --base flag of spawn command
	spawnCmd.Flags().StringToString("base", map[string]string{}, "usage of base flag")
	// TODO: :bike: fill usage of --var flag
	spawnCmd.Flags().StringArray("var", []string{}, "usage of var flag")
	rootCmd.AddCommand(spawnCmd)
}

var spawnCmd = &cobra.Command{
	Use: "spawn [layer] <name>",
	// TODO: :bike: fill short description of spawn command
	Short: "spawn short help text",
	// TODO: :bike: fill long description of spawn command
	Long: "spawn long help text",
	Args: cobra.MinimumNArgs(1),
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

		dependenciesState, err := cmd.Flags().GetStringToString("base")
		if err != nil {
			return errors.Wrap(err, "fail to get --base flag, this is a bug in layerform")
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
		stateName := shortuuid.New()
		if len(args) > 1 {
			stateName = args[1]
		}

		spawn := command.NewSpawn(layersBackend, statesBackend)

		return spawn.Run(ctx, layerName, stateName, dependenciesState, vars)
	},
}

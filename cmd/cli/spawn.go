package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/lithammer/shortuuid/v3"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	spawnCmd.Flags().StringToString("base", map[string]string{}, "a map of underlying layers and their IDs to place the layer on top of")
	spawnCmd.Flags().StringArray("var", []string{}, "a map of variables for the layer's Terraform files. I.e. 'foo=bar,baz=qux'")
	rootCmd.AddCommand(spawnCmd)
}

var spawnCmd = &cobra.Command{
	Use:   "spawn <layer> [desired_id]",
	Short: "creates a layer instance",
	Long: `The spawn command creates a layer instance.

Whenever a desired ID is not provided, Layerform will generate a random UUID for the layer instance.

If an instance with the same ID already exists for the layer definition, Layerform will return an error.
    `,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		vars, err := cmd.Flags().GetStringArray("var")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --var flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}

		dependenciesState, err := cmd.Flags().GetStringToString("base")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --base flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get layers backend"))
			os.Exit(1)
			return
		}

		statesBackend, err := cfg.GetStateBackend(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get state backend"))
			os.Exit(1)
			return
		}

		layerName := args[0]
		stateName := shortuuid.New()
		if len(args) > 1 {
			stateName = args[1]
		}

		spawn := command.NewSpawn(layersBackend, statesBackend)

		err = spawn.Run(ctx, layerName, stateName, dependenciesState, vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

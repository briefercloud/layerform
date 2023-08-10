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
	killCmd.Flags().StringArray("var", []string{}, "a map of variables for the layer's Terraform files. I.e. 'foo=bar,baz=qux'")

	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "destroys a layer instance",
	Long: `The kill command destroys a layer instance.

Please notice that the kill command cannot destroy a layer instance which has dependants. To delete a layer instance with dependants, you must first delete all of its dependants.
    `,
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

		statesBackend, err := cfg.GetStateBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get state backend")
		}

		layerName := args[0]
		stateName := shortuuid.New()
		if len(args) > 1 {
			stateName = args[1]
		}

		kill := command.NewKill(layersBackend, statesBackend)

		return kill.Run(ctx, layerName, stateName, vars)
	},
}

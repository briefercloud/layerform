package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/lfconfig"
	"github.com/ergomake/layerform/pkg/command"
)

func init() {
	outputCmd.Flags().String("template", "", "path to a mustache template file to render the output into")
	rootCmd.AddCommand(outputCmd)
}

var outputCmd = &cobra.Command{
	Use:   "output <layer> <instance>",
	Args:  cobra.MinimumNArgs(2),
	Short: "Reads all output variables from the provided layer instance",
	Long:  `The output command reads all output variables from the given layer instance and prints them as json to standard output.`,
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

		layerName := args[0]
		instanceName := args[1]

		output := command.NewOutput(layersBackend, instancesBackend)

		template, err := cmd.Flags().GetString("template")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --template flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}

		err = output.Run(ctx, layerName, instanceName, template)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", output.Run(ctx, layerName, instanceName, template))
			os.Exit(1)
		}
	},
}

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	killCmd.Flags().StringArray("var", []string{}, "a map of variables for the layer's Terraform files. I.e. 'foo=bar,baz=qux'")
	killCmd.Flags().Bool("auto-approve", false, "skip interactive approval before killing layer instance")

	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use:   "kill <layer> <instance>",
	Args:  cobra.MinimumNArgs(2),
	Short: "destroys a layer instance",
	Long: `The kill command destroys a layer instance.

Please notice that the kill command cannot destroy a layer instance which has dependants. To delete a layer instance with dependants, you must first delete all of its dependants.
    `,
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
		kill, err := cfg.GetKillCommand(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get kill command"))
			os.Exit(1)
		}

		layerName := args[0]
		instanceName := args[1]

		err = kill.Run(ctx, layerName, instanceName, false, vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

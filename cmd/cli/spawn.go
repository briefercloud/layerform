package cli

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/go-hclog"
	"github.com/lithammer/shortuuid/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	spawnCmd.Flags().StringToString("base", map[string]string{}, "a map of underlying layers and their IDs to place the layer on top of")
	spawnCmd.Flags().StringArray("var", []string{}, "a map of variables for the layer's Terraform files. I.e. 'foo=bar,baz=qux'")
	rootCmd.AddCommand(spawnCmd)
}

var alphanumericRegex = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9_-]*[A-Za-z0-9]$")

var spawnCmd = &cobra.Command{
	Use:   "spawn <layer> [desired_id]",
	Short: "Creates a layer instance",
	Long: `The spawn command creates a layer instance.

Whenever a desired ID is not provided, Layerform will generate a random UUID for the layer instance.

If an instance with the same ID already exists for the layer definition, Layerform will return an error.`,
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

		dependenciesInstance, err := cmd.Flags().GetStringToString("base")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --base flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}

		spawn, err := cfg.GetSpawnCommand(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get spawn command"))
			os.Exit(1)
		}

		layerName := args[0]
		instanceName := shortuuid.New()
		if len(args) > 1 {
			instanceName = args[1]
		}

		if !alphanumericRegex.MatchString(instanceName) {
			fmt.Fprintf(os.Stderr, "Invalid name: %s\n", instanceName)
			fmt.Fprintln(os.Stderr, "Name must start and end with an alphanumeric character and can include dashes and underscores in between.")
			os.Exit(1)
		}

		err = spawn.Run(ctx, layerName, instanceName, dependenciesInstance, vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

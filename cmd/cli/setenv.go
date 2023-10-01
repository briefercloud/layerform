package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
	"github.com/ergomake/layerform/pkg/data"
)

func init() {
	rootCmd.AddCommand(setEnvCmd)
}

var setEnvCmd = &cobra.Command{
	Use:   "set-env <VAR_NAME> <value>",
	Short: "set an environment variable to be used when spawning a layer",
	Long: `The set-env command sets an environment variable to be used when spawning layers.

These are often used for configuring the providers, for instance, you can use this command to set AWS credentials.
Environment variables can also be used to set values for the variables in your layers. The environment variables must be in the format TF_VAR_name.`,
	Example: `# Set value for a variable in your layer
layerform set-env TF_VAR_foo bar`,
	Args: cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
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

		varName := args[0]
		varValue := args[1]

		envvarsBackend, err := cfg.GetEnvVarsBackend(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get environment variables backend"))
			os.Exit(1)
		}

		err = envvarsBackend.SaveVariable(ctx, &data.EnvVar{Name: varName, Value: varValue})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to save environment variable"))
			os.Exit(1)
		}
	},
}

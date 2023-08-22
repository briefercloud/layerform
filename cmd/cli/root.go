package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/telemetry"
)

var rootCmd = &cobra.Command{
	Use:   "layerform",
	Short: "Layerform helps engineers create their own staging environments using plain Terraform files.",
	Long: `Layerform helps engineers create their own staging environments using plain Terraform files.

Please read our documentation at https://docs.layerform.dev for more information.
`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		telemetry.Push(
			telemetry.EventRunCommand,
			map[string]interface{}{"command": cmd.CalledAs()},
		)
	},
}

func Execute() {
	telemetry.Init()
	defer telemetry.Close()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

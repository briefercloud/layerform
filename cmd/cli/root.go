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

Please read our documentation at https://docs.layerform.dev for more information.`,
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func Execute() {
	telemetry.Init()
	telemetry.RegisterCommand()
	defer telemetry.Close()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

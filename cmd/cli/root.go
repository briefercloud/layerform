package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "layerform",
	Short: "Layerform helps engineers create their own staging environments using plain Terraform files.",
	Long: `Layerform helps engineers create their own staging environments using plain Terraform files.

Please read our documentation at https://docs.layerform.dev for more information.
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

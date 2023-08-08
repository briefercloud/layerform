package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "layerform",
	// TODO: :bike: fill short description of layerform
	Short: "layerform short help text",
	// TODO: :bike: fill long description of layerform
	Long: "layerform long help text",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

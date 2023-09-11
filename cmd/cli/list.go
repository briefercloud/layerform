package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List layerform resources",
	Long: `List layerform resources.

Prints a table of the most important information about the specified resource.`,

	Example: `# List all layer definitions
layerform list definitions

# List all layer instances
layerform list instances`,
}

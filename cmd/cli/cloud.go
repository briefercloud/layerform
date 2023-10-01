package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cloudCmd)
}

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Modify layerform cloud entities",
	Long: `Modify layerform cloud entities using subcomands like "layerform cloud create-user"

This command only works if the current context is of type "cloud"`,
	Example: `# Create a new cloud user
layerform cloud create-user --name "John Doe" --email john@doe.com`,
}

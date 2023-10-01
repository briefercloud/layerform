package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Modify layerform config file",
	Long:  `Modify layerform config file using subcomands like "layerform config set-context"`,
	Example: `# Set a context entry in config
layerform config set-context example --type=local --dir=~/.layerform/contexts/example`,
}

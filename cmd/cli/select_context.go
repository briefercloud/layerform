package cli

import (
	"fmt"
	"os"

	"github.com/ergomake/layerform/internal/lfconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configSelectContextCmd)
}

var configSelectContextCmd = &cobra.Command{
	Use:   "select-context <name>",
	Short: "Select a context entry from layerform config file",
	Long: `Select a context entry from layerform config file.
	
  Selecting a name that does not exist will return error.`,
	Example: `# Select a context
layerform config select-context local-example`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := lfconfig.Load("")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to open config file"))
			os.Exit(1)
		}

		_, ok := cfg.Contexts[name]
		if !ok {
			fmt.Fprintf(
				os.Stderr,
				"context %s does not exist\n",
				name,
			)
			os.Exit(1)
		}

		cfg.CurrentContext = name

		err = cfg.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to save config file"))
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Context \"%s\" selected.\n", name)
	},
	SilenceErrors: true,
}

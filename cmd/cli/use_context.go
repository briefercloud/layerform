package cli

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	configCmd.AddCommand(configUseContextCmd)
}

var configUseContextCmd = &cobra.Command{
	Use:   "use-context <name>",
	Short: "Use a context entry from layerform config file",
	Long: `Use a context entry from layerform config file.
	
  Using a name that does not exist will return error.`,
	Example: `# Use a context
layerform config use-context local-example`,
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
				"no context exists with the name \"%s\".\n",
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

		fmt.Fprintf(os.Stdout, "Switched to context \"%s\".\n", name)
	},
	SilenceErrors: true,
}

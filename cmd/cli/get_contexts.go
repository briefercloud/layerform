package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	configCmd.AddCommand(configGetContextsCmd)
}

var configGetContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "Display contexts from layerform config file",
	Long:  `Display contexts from layerform config file`,
	Run: func(_ *cobra.Command, _ []string) {
		cfg, err := lfconfig.Load("")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stdout, "No contexts configure, configure contexts using the set-context command.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "CURRENT\tNAME\tTYPE\tLOCATION")
		for name, ctx := range cfg.Contexts {
			isCurrent := name == cfg.CurrentContext
			current := ""
			if isCurrent {
				current = "*"
			}

			fmt.Fprintln(w, strings.Join([]string{current, name, ctx.Type, ctx.Location()}, "\t"))
		}
		err = w.Flush()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to print output"))
			os.Exit(1)
		}

	},
	SilenceErrors: true,
}

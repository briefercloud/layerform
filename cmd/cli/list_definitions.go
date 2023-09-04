package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
	"github.com/ergomake/layerform/pkg/data"
)

func init() {
	listCmd.AddCommand(listDefinitionsCmd)
}

var listDefinitionsCmd = &cobra.Command{
	Use:   "definitions",
	Short: "List layers definitions",
	Long: `List layers definitions.

Prints a table of the most important information about layer definitions.`,

	Run: func(_ *cobra.Command, _ []string) {
		logger := hclog.Default()
		logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
		if logLevel != hclog.NoLevel {
			logger.SetLevel(logLevel)
		}
		ctx := hclog.WithContext(context.Background(), logger)

		cfg, err := lfconfig.Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to load config"))
			os.Exit(1)
			return
		}

		layersBackend, err := cfg.GetDefinitionsBackend(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get layers backend"))
			os.Exit(1)
			return
		}

		layers, err := layersBackend.ListLayers(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to list layer definitions"))
			os.Exit(1)
			return
		}

		if len(layers) == 0 {
			fmt.Fprintln(
				os.Stdout,
				"No layer definitions configured, provision layers by running \"layerform configure\"",
			)
			return
		}

		sortLayersByDepth(layers)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(w, "NAME\tDEPENDENCIES")
		for _, layer := range layers {
			deps := strings.Join(layer.Dependencies, ",")
			fmt.Fprintln(w, layer.Name+"\t"+deps)
		}
		err = w.Flush()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to print output"))
			os.Exit(1)
		}
	},
}

func sortLayersByDepth(layers []*data.Definition) {
	byName := make(map[string]*data.Definition)
	for _, l := range layers {
		byName[l.Name] = l
	}

	sort.SliceStable(layers, func(x, y int) bool {
		lx := layers[x]
		ly := layers[y]

		return computeDepth(lx, byName, 0) < computeDepth(ly, byName, 0)
	})
}

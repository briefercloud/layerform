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

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	listCmd.AddCommand(listDefinitionsCmd)
}

var listDefinitionsCmd = &cobra.Command{
	Use:   "definitions",
	Short: "List layers definitions",
	Long: `List layers definitions.

Prints a table of the most important information about layer definitions.`,

	RunE: func(_ *cobra.Command, _ []string) error {
		logger := hclog.Default()
		logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
		if logLevel != hclog.NoLevel {
			logger.SetLevel(logLevel)
		}
		ctx := hclog.WithContext(context.Background(), logger)

		cfg, err := lfconfig.Load("")
		if err != nil {
			return errors.Wrap(err, "fail to load config")
		}

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend")
		}

		layers, err := layersBackend.ListLayers(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to list layer definitions")
		}

		if len(layers) == 0 {
			_, err := fmt.Println("No layer definitions are configured, provision layers by running \"layerform configure\"")
			return errors.Wrap(err, "fail to print output")
		}

		sortLayersByDepth(layers)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(w, "NAME\tDEPENDENCIES")
		for _, layer := range layers {
			deps := strings.Join(layer.Dependencies, ",")
			fmt.Fprintln(w, layer.Name+"\t"+deps)
		}
		err = w.Flush()

		return errors.Wrap(err, "fail to print output")
	},
}

func sortLayersByDepth(layers []*model.Layer) {
	byName := make(map[string]*model.Layer)
	for _, l := range layers {
		byName[l.Name] = l
	}

	sort.SliceStable(layers, func(x, y int) bool {
		lx := layers[x]
		ly := layers[y]

		return computeDepth(lx, byName, 0) < computeDepth(ly, byName, 0)
	})
}

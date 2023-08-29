package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layerstate"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	listCmd.AddCommand(listInstancesCmd)
}

var listInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List layers instances",
	Long: `List layers instances.

Prints a table of the most important information about layer instances.`,

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

		instancesBackend, err := cfg.GetStateBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers instances backend")
		}

		instances, err := instancesBackend.ListStates(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to list layer instances")
		}

		if len(instances) == 0 {
			_, err := fmt.Println("No layer instances are spawned, spawn layers by running \"layerform spawn\"")
			return errors.Wrap(err, "fail to print output")
		}

		layers, err := layersBackend.ListLayers(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to list layer definitions")
		}

		layersByName := make(map[string]*model.Layer)
		for _, l := range layers {
			layersByName[l.Name] = l
		}

		sortInstancesByDepth(instances, layersByName)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(w, "INSTANCE NAME\tLAYER NAME\tDEPENDENCIES")
		for _, instance := range instances {
			layer := layersByName[instance.LayerName]
			deps := ""
			for i, dep := range layer.Dependencies {
				if i > 0 {
					deps += ","
				}

				depInstName := instance.GetDependencyStateName(dep)
				deps += dep + "=" + depInstName
			}

			fmt.Fprintln(w, instance.StateName+"\t"+instance.LayerName+"\t"+deps)
		}
		err = w.Flush()

		return errors.Wrap(err, "fail to print output")
	},
}

func computeDepth(layer *model.Layer, layers map[string]*model.Layer, level int) int {
	depth := level
	for _, d := range layer.Dependencies {
		dDepth := computeDepth(layers[d], layers, level+1)
		if dDepth > depth {
			depth = dDepth
		}
	}

	return depth
}

func sortInstancesByDepth(instances []*layerstate.State, layers map[string]*model.Layer) {
	sort.SliceStable(instances, func(x, y int) bool {
		instX := instances[x]
		layerX := layers[instX.LayerName]
		depthX := computeDepth(layerX, layers, 0)

		instY := instances[y]
		layerY := layers[instY.LayerName]
		depthY := computeDepth(layerY, layers, 0)

		return depthX < depthY
	})
}

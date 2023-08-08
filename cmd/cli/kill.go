package cli

import (
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
)

func init() {
	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use: "kill",
	// TODO: :bike: fill short description of kill command
	Short: "kill short help text",
	// TODO: :bike: fill long description of kill command
	Long: "kill long help text",
	RunE: func(_ *cobra.Command, args []string) error {
		layerfile, err := layerfile.FromFile("layerform.json")
		if err != nil {
			return errors.Wrap(err, "fail to load layerform.json")
		}

		layerslist, err := layerfile.ToLayers()
		if err != nil {
			return errors.Wrap(err, "fail to import layers defined at layerform.json")
		}

		layersBackend := layers.NewInMemoryBackend(layerslist)
		statesBackend := layerstate.NewFileBackend("layerform.lfstate")

		kill := command.NewKill(layersBackend, statesBackend)

		return kill.Run(args)
	},
}

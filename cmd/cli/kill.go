package cli

import (
	"github.com/lithammer/shortuuid/v3"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/command"
	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
)

func init() {
	// TODO: :bike: fill usage of --var flag for the kill command
	killCmd.Flags().StringArray("var", []string{}, "usage of var flag")

	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use: "kill",
	// TODO: :bike: fill short description of kill command
	Short: "kill short help text",
	// TODO: :bike: fill long description of kill command
	Long: "kill long help text",
	RunE: func(cmd *cobra.Command, args []string) error {
		vars, err := cmd.Flags().GetStringArray("var")
		if err != nil {
			return errors.Wrap(err, "fail to get --var flag, this is a bug in layerform")
		}

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

		layerName := args[0]
		stateName := shortuuid.New()
		if len(args) > 1 {
			stateName = args[1]
		}

		kill := command.NewKill(layersBackend, statesBackend)

		return kill.Run(layerName, stateName, vars)
	},
}

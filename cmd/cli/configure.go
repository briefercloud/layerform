package cli

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	// TODO: :bike: fill usage of --file flag
	configureCmd.Flags().String("file", "layerform.json", "usage of file flag")
	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use: "configure",
	// TODO: :bike: fill short description of configure command
	Short: "configure short help text",
	// TODO: :bike: fill long description of configure command
	Long: "configure long help text",
	RunE: func(cmd *cobra.Command, _ []string) error {
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

		fpath, err := cmd.Flags().GetString("file")
		if err != nil {
			return errors.Wrap(err, "fail to get --file flag, this is a bug in layerform")
		}

		layerfile, err := layerfile.FromFile(fpath)
		if err != nil {
			return errors.Wrap(err, "fail to read layerform layers definitions from file")
		}

		layers, err := layerfile.ToLayers()
		if err != nil {
			return errors.Wrap(err, "fail to load layers from layerform layers definitions file")
		}

		layersBackend, err := cfg.GetLayersBackend(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get layers backend")
		}

		err = layersBackend.UpdateLayers(ctx, layers)
		return errors.Wrap(err, "fail to update layers")
	},
}

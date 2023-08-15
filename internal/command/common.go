package command

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/pathutils"
	"github.com/ergomake/layerform/internal/tags"
)

func writeLayerToWorkdir(
	ctx context.Context,
	layersBackend layers.Backend,
	layerWorkdir string,
	layer *model.Layer,
	stateByLayer map[string]string,
) (string, error) {
	logger := hclog.FromContext(ctx).With("layer", layer.Name, "layerWorkdir", layerWorkdir)
	logger.Debug("Writting layer to workdir")

	var inner func(*model.Layer) ([]string, error)
	inner = func(layer *model.Layer) ([]string, error) {
		fpaths := make([]string, 0)
		for _, dep := range layer.Dependencies {
			logger.Debug("Writting dependency to workdir", "dependency", dep)

			layer, err := layersBackend.GetLayer(ctx, dep)
			if err != nil {
				return nil, errors.Wrap(err, "fail to get layer")
			}

			depPaths, err := inner(layer)
			if err != nil {
				return nil, errors.Wrap(err, "fail to write layer to workdir")
			}

			fpaths = append(fpaths, depPaths...)
		}

		stateName := stateByLayer[layer.Name]

		for _, f := range layer.Files {
			fpaths = append(fpaths, f.Path)
			fpath := path.Join(layerWorkdir, f.Path)

			err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
			if err != nil {
				return fpaths, errors.Wrap(err, "fail to MkdirAll")
			}

			err = os.WriteFile(fpath, f.Content, 0644)
			if err != nil {
				return fpaths, errors.Wrap(err, "fail to write layer file")
			}

			err = tags.AddTagsToFile(
				fpath,
				map[string]string{
					"layerform_layer_name":     layer.Name,
					"layerform_layer_instance": stateName,
				},
			)
			if err != nil {
				return fpaths, errors.Wrap(err, "fail to add tags")
			}
		}

		return fpaths, nil
	}

	paths, err := inner(layer)
	if err != nil {
		return "", errors.Wrap(err, "fail to write layer to workdir")
	}

	return path.Join(layerWorkdir, pathutils.FindCommonParentPath(paths)), nil
}

func getTFState(ctx context.Context, statePath string, tfpath string) (*tfjson.State, error) {
	hclog.FromContext(ctx).Debug("Getting terraform state", "path", statePath)
	dir := filepath.Dir(statePath)
	tf, err := tfexec.NewTerraform(dir, tfpath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create terraform client")
	}

	return tf.ShowStateFile(ctx, statePath)
}

func getStateModuleAddresses(module *tfjson.StateModule) []string {
	addresses := make([]string, 0)
	for _, res := range module.Resources {
		addresses = append(addresses, res.Address)
	}

	for _, child := range module.ChildModules {
		addresses = append(addresses, getStateModuleAddresses(child)...)
	}

	return addresses
}

func findTFVarFiles() ([]string, error) {
	var matchingFiles []string

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "fail to get current working directory")
	}

	filepath.WalkDir(cwd, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "fail to walk current working directory")
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		if filename == "terraform.tfvars" ||
			filename == "terraform.tfvars.json" ||
			strings.HasSuffix(filename, ".auto.tfvars") ||
			strings.HasSuffix(filename, ".auto.tfvars.json") {
			matchingFiles = append(matchingFiles, path)
		}

		return nil
	})

	return matchingFiles, nil
}

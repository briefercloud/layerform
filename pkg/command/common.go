package command

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/pathutils"
	"github.com/ergomake/layerform/internal/tags"
	"github.com/ergomake/layerform/internal/tfclient"
	"github.com/ergomake/layerform/pkg/data"
	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

func WriteLayerToWorkdir(
	ctx context.Context,
	definitionsBackend layerdefinitions.Backend,
	layerWorkdir string,
	layer *data.LayerDefinition,
	instanceByLayer map[string]string,
) (string, error) {
	logger := hclog.FromContext(ctx).With("layer", layer.Name, "layerWorkdir", layerWorkdir)
	logger.Debug("Writting layer to workdir")

	var inner func(*data.LayerDefinition) ([]string, error)
	inner = func(layer *data.LayerDefinition) ([]string, error) {
		fpaths := make([]string, 0)
		for _, dep := range layer.Dependencies {
			logger.Debug("Writting dependency to workdir", "dependency", dep)

			layer, err := definitionsBackend.GetLayer(ctx, dep)
			if err != nil {
				return nil, errors.Wrap(err, "fail to get layer")
			}

			depPaths, err := inner(layer)
			if err != nil {
				return nil, errors.Wrap(err, "fail to write layer to workdir")
			}

			fpaths = append(fpaths, depPaths...)
		}

		instanceName := instanceByLayer[layer.Name]

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
					"layerform_layer_instance": instanceName,
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

	commonParentPath := pathutils.FindCommonParentPath(paths)
	dir := path.Join(layerWorkdir, commonParentPath)

	err = writeLFVars(dir, instanceByLayer)
	if err != nil {
		return "", errors.Wrap(err, "fail to write layerform specific variable definitions")
	}

	return dir, nil
}

func writeLFVars(dir string, instanceByDefinition map[string]string) error {
	definitions := ""
	for def := range instanceByDefinition {
		definitions += def + "= string\n"
	}

	defaults := ""
	for def, inst := range instanceByDefinition {
		defaults += fmt.Sprintf("%s=\"%s\"\n", def, inst)
	}

	lfVars := fmt.Sprintf(`variable "lf_names" {
  type = object({
    %s
  })
  default = {
    %s
  }
}`, definitions, defaults)

	suffix := time.Now().Unix()
	fname := fmt.Sprintf("lf_names-%d.tf", suffix)
	return os.WriteFile(path.Join(dir, fname), []byte(lfVars), 0644)
}

func GetTFState(ctx context.Context, statePath string, tfpath string) (*tfjson.State, error) {
	hclog.FromContext(ctx).Debug("Getting terraform state", "path", statePath)
	dir := filepath.Dir(statePath)
	tf, err := tfclient.New(dir, tfpath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create terraform client")
	}

	return tf.ShowStateFile(ctx, statePath)
}

func GetStateModuleAddresses(module *tfjson.StateModule) []string {
	addresses := make([]string, 0)
	for _, res := range module.Resources {
		addresses = append(addresses, res.Address)
	}

	for _, child := range module.ChildModules {
		addresses = append(addresses, GetStateModuleAddresses(child)...)
	}

	return addresses
}

func FindTFVarFiles() ([]string, error) {
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

func ComputeInstanceByLayer(
	ctx context.Context,
	definitionsBackend layerdefinitions.Backend,
	instancesBackend layerinstances.Backend,
	layer *data.LayerDefinition,
	instance *data.LayerInstance,
) (map[string]string, error) {
	instanceByLayer := map[string]string{}
	instanceByLayer[layer.Name] = instance.InstanceName
	for _, dep := range layer.Dependencies {
		depLayer, err := definitionsBackend.GetLayer(ctx, dep)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get layer")
		}

		depInstanceName := instance.GetDependencyInstanceName(dep)

		depInstance, err := instancesBackend.GetInstance(ctx, dep, depInstanceName)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get instance")
		}

		depDepInstances, err := ComputeInstanceByLayer(ctx, definitionsBackend, instancesBackend, depLayer, depInstance)
		if err != nil {
			return nil, errors.Wrap(err, "fail to compute instance by layer")
		}

		for k, v := range depDepInstances {
			instanceByLayer[k] = v
		}

		instanceByLayer[dep] = depInstanceName
	}

	return instanceByLayer, nil
}

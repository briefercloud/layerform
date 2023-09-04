package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"

	"github.com/cbroglie/mustache"

	"github.com/ergomake/layerform/internal/layerdefinitions"
	"github.com/ergomake/layerform/internal/layerinstances"
	"github.com/ergomake/layerform/internal/terraform"
	"github.com/ergomake/layerform/internal/tfclient"
)

type outputCommand struct {
	definitionsBackend layerdefinitions.Backend
	instancesBackend   layerinstances.Backend
}

func NewOutput(definitionsBackend layerdefinitions.Backend, instancesBackend layerinstances.Backend) *outputCommand {
	return &outputCommand{definitionsBackend, instancesBackend}
}

func (c *outputCommand) Run(ctx context.Context, layerName, instanceName, template string) error {
	logger := hclog.FromContext(ctx)

	layer, err := c.definitionsBackend.GetLayer(ctx, layerName)
	if err != nil {
		return errors.Wrap(err, "fail to get layer")
	}

	if layer == nil {
		return errors.New("layer not found")
	}

	instance, err := c.instancesBackend.GetInstance(ctx, layer.Name, instanceName)
	if err != nil {
		if errors.Is(err, layerinstances.ErrInstanceNotFound) {
			return errors.Errorf(
				"instance %s not found for layer %s\n",
				instanceName,
				layer.Name,
			)
		}

		return errors.Wrap(err, "fail to get layer instance")
	}

	tfpath, err := terraform.GetTFPath(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to get terraform path")
	}
	logger.Debug("Using terraform from", "tfpath", tfpath)
	logger.Debug("Found terraform installation", "tfpath", tfpath)

	logger.Debug("Creating a temporary work directory")
	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.Wrap(err, "fail to create work directory")
	}
	defer os.RemoveAll(workdir)

	layerDir := path.Join(workdir, layerName)

	instanceByLayer, err := computeInstanceByLayer(ctx, c.definitionsBackend, c.instancesBackend, layer, instance)
	if err != nil {
		return errors.Wrap(err, "fail to compute instance by layer instance")
	}

	layerWorkdir, err := writeLayerToWorkdir(ctx, c.definitionsBackend, layerDir, layer, instanceByLayer)
	if err != nil {
		return errors.Wrap(err, "fail to write layer to work directory")
	}

	statePath := path.Join(layerWorkdir, "terraform.tfstate")
	err = os.WriteFile(statePath, instance.Bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "fail to write terraform state to work directory")
	}

	tf, err := tfclient.New(layerWorkdir, tfpath)
	if err != nil {
		return errors.Wrap(err, "fail to get terraform client")
	}

	output, err := tf.Output(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to terraform output")
	}

	if template != "" {
		context, err := prepareOutputForTemplate(output)
		if err != nil {
			return errors.Wrap(err, "fail to prepare output for template")
		}

		mustache.AllowMissingVariables = false
		result, err := mustache.RenderFile(template, context)
		if err != nil {
			return errors.Wrapf(err, "fail to render template %s", template)
		}

		fmt.Fprint(os.Stdout, result)
		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(output)
	return errors.Wrap(err, "fail to encode output to json")
}

func prepareOutputForTemplate(output map[string]tfexec.OutputMeta) (map[string]interface{}, error) {
	var context map[string]interface{}
	bs, err := json.Marshal(output)
	if err != nil {
		return nil, errors.Wrap(err, "fail to marshal output to json")
	}

	err = json.Unmarshal(bs, &context)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal output from context")
	}

	return context, nil
}

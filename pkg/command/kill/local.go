package kill

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/chelnak/ysmrr"
	"github.com/chelnak/ysmrr/pkg/animations"
	"github.com/chelnak/ysmrr/pkg/colors"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/terraform"
	"github.com/ergomake/layerform/internal/tfclient"
	"github.com/ergomake/layerform/pkg/command"
	"github.com/ergomake/layerform/pkg/data"
	"github.com/ergomake/layerform/pkg/envvars"
	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type localKillCommand struct {
	definitionsBackend layerdefinitions.Backend
	instancesBackend   layerinstances.Backend
	envVarsBackend     envvars.Backend
}

var _ Kill = &localKillCommand{}

func NewLocal(
	definitionsBackend layerdefinitions.Backend,
	instancesBackend layerinstances.Backend,
	envVarsBackend envvars.Backend,
) *localKillCommand {
	return &localKillCommand{definitionsBackend, instancesBackend, envVarsBackend}
}

func (c *localKillCommand) Run(
	ctx context.Context,
	layerName, instanceName string,
	autoApprove bool,
	vars []string,
) error {
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
				"instance %s not found for layer %s",
				instanceName,
				layer.Name,
			)
		}

		return errors.Wrap(err, "fail to get layer instance")
	}

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()
	s := sm.AddSpinner(
		fmt.Sprintf(
			"Preparing to kill instance \"%s\" of layer \"%s\"",
			instanceName,
			layerName,
		),
	)

	hasDependants, err := HasDependants(
		ctx,
		c.instancesBackend,
		c.definitionsBackend,
		layerName,
		instanceName,
	)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to check if layer has dependants")
	}

	if hasDependants {
		s.Error()
		sm.Stop()
		return errors.New("can't kill this layer because other layers depend on it")
	}

	envVars, err := c.envVarsBackend.ListVariables(ctx)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to list environment variables")
	}

	for _, envVar := range envVars {
		err := os.Setenv(envVar.Name, envVar.Value)
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrapf(err, "fail to set %s environment variable", envVar.Name)
		}
	}

	tfpath, err := terraform.GetTFPath(ctx)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get terraform path")
	}
	logger.Debug("Using terraform from", "tfpath", tfpath)
	logger.Debug("Found terraform installation", "tfpath", tfpath)

	logger.Debug("Creating a temporary work directory")
	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to create work directory")
	}
	defer os.RemoveAll(workdir)

	layerDir := path.Join(workdir, layerName)
	layerAddrs, layerDir, err := c.getLayerAddresses(ctx, layer, instance, layerDir, tfpath)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get layer addresses")
	}

	layerAddrsMap := make(map[string]struct{})
	for _, addr := range layerAddrs {
		layerAddrsMap[addr] = struct{}{}
	}

	for _, dep := range layer.Dependencies {
		depLayer, err := c.definitionsBackend.GetLayer(ctx, dep)
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to get dependency layer")
		}

		if depLayer == nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "dependency layer not found")
		}

		depInstance, err := c.instancesBackend.GetInstance(ctx, depLayer.Name, instance.GetDependencyInstanceName(dep))
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to get dependency instance")
		}

		depDir := path.Join(workdir, dep)
		depAddrs, _, err := c.getLayerAddresses(ctx, depLayer, depInstance, depDir, tfpath)
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to get dependency layer addresses")
		}

		for _, addr := range depAddrs {
			delete(layerAddrsMap, addr)
		}
	}

	tf, err := tfclient.New(layerDir, tfpath)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get terraform client")
	}

	logger.Debug("Looking for variable definitions in .tfvars files")
	varFiles, err := command.FindTFVarFiles()
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to find .tfvars files")
	}
	logger.Debug(fmt.Sprintf("Found %d var files", len(varFiles)), "varFiles", varFiles)

	destroyOptions := make([]tfexec.DestroyOption, 0)
	for _, vf := range varFiles {
		destroyOptions = append(destroyOptions, tfexec.VarFile(vf))
	}
	for _, v := range vars {
		destroyOptions = append(destroyOptions, tfexec.Var(v))
	}

	for addr := range layerAddrsMap {
		destroyOptions = append(destroyOptions, tfexec.Target(addr))
	}
	logger.Debug(
		"Running terraform destroy targetting layer specific addresses",
		"layer", layer.Name, "instance", instanceName, "targets", destroyOptions,
	)

	s.Complete()
	sm.Stop()

	if !autoApprove {
		var answer string
		fmt.Print("Are you sure? This can't be undone. [yes/no]: ")
		_, err = fmt.Scan(&answer)
		if err != nil {
			return errors.Wrap(err, "fail to read asnwer")
		}

		if strings.ToLower(strings.TrimSpace(answer)) != "yes" {
			return nil
		}
	}

	sm = ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()

	s = sm.AddSpinner(
		fmt.Sprintf(
			"Killing instance \"%s\" of layer \"%s\"",
			instanceName,
			layerName,
		),
	)

	err = tf.Destroy(ctx, destroyOptions...)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to terraform destroy")
	}

	err = c.instancesBackend.DeleteInstance(ctx, layerName, instanceName)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to delete instance")
	}

	s.Complete()
	sm.Stop()

	return nil
}

func (c *localKillCommand) getLayerAddresses(
	ctx context.Context,
	layer *data.LayerDefinition,
	instance *data.LayerInstance,
	layerDir, tfpath string,
) ([]string, string, error) {
	logger := hclog.FromContext(ctx)
	logger.Debug("Getting layer addresses", "layer", layer.Name, "instance", instance.InstanceName)

	instanceByLayer, err := command.ComputeInstanceByLayer(ctx, c.definitionsBackend, c.instancesBackend, layer, instance)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to compute instance by layer instance")
	}

	layerWorkdir, err := command.WriteLayerToWorkdir(ctx, c.definitionsBackend, layerDir, layer, instanceByLayer)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to write layer to work directory")
	}

	statePath := path.Join(layerWorkdir, "terraform.tfstate")
	err = os.WriteFile(statePath, instance.Bytes, 0644)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to write terraform state to work directory")
	}

	tf, err := tfclient.New(layerWorkdir, tfpath)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to get terraform client")
	}

	err = tf.Init(ctx, layer.SHA)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to terraform init")
	}

	tfState, err := command.GetTFState(ctx, statePath, tfpath)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to get terraform state")
	}

	addresses := command.GetStateModuleAddresses(tfState.Values.RootModule)

	return addresses, layerWorkdir, nil
}

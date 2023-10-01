package refresh

import (
	"context"
	"fmt"
	"os"
	"path"

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

type localRefreshCommand struct {
	definitionsBackend layerdefinitions.Backend
	instancesBackend   layerinstances.Backend
	envVarsBackend     envvars.Backend
}

var _ Refresh = &localRefreshCommand{}

func NewLocal(
	definitionsBackend layerdefinitions.Backend,
	instancesBackend layerinstances.Backend,
	envVarsBackend envvars.Backend,
) *localRefreshCommand {
	return &localRefreshCommand{definitionsBackend, instancesBackend, envVarsBackend}
}

func (c *localRefreshCommand) Run(
	ctx context.Context,
	definitionName, instanceName string,
	vars []string,
) error {
	logger := hclog.FromContext(ctx)

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()
	s := sm.AddSpinner(
		fmt.Sprintf(
			"Preparing to refresh instance \"%s\" of layer \"%s\"",
			instanceName,
			definitionName,
		),
	)

	definition, err := c.definitionsBackend.GetLayer(ctx, definitionName)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get layer definition")
	}

	instance, err := c.instancesBackend.GetInstance(ctx, definition.Name, instanceName)
	if err != nil {
		s.Error()
		sm.Stop()

		if errors.Is(err, layerinstances.ErrInstanceNotFound) {
			return errors.Errorf(
				"instance %s not found for layer %s",
				instanceName,
				definition.Name,
			)
		}

		return errors.Wrap(err, "fail to get layer instance")
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

	layerDir := path.Join(workdir, definitionName)

	instanceByLayer, err := command.ComputeInstanceByLayer(
		ctx,
		c.definitionsBackend,
		c.instancesBackend,
		definition,
		instance,
	)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to compute instance by layer instance")
	}

	layerWorkdir, err := command.WriteLayerToWorkdir(ctx, c.definitionsBackend, layerDir, definition, instanceByLayer)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to write layer to work directory")
	}

	statePath := path.Join(layerWorkdir, "terraform.tfstate")
	err = os.WriteFile(statePath, instance.Bytes, 0644)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to write terraform state to work directory")
	}

	tf, err := tfclient.New(layerWorkdir, tfpath)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get terraform client")
	}

	err = tf.Init(ctx, definition.SHA)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to terraform init")
	}

	logger.Debug("Looking for variable definitions in .tfvars files")
	varFiles, err := command.FindTFVarFiles()
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to find .tfvars files")
	}
	logger.Debug(fmt.Sprintf("Found %d var files", len(varFiles)), "varFiles", varFiles)

	applyOptions := []tfexec.ApplyOption{}
	for _, vf := range varFiles {
		applyOptions = append(applyOptions, tfexec.VarFile(vf))
	}
	for _, v := range vars {
		applyOptions = append(applyOptions, tfexec.Var(v))
	}

	s.Complete()

	s = sm.AddSpinner(
		fmt.Sprintf(
			"Refreshing instance \"%s\" of layer \"%s\"",
			instanceName,
			definitionName,
		),
	)

	err = tf.Apply(ctx, applyOptions...)
	if err != nil {
		originalErr := err

		nextStateBytes, err := os.ReadFile(statePath)
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to read next state")
		}

		// if this refresh attempt generated state, we should save it as faulty
		// so user can fix it later
		if len(nextStateBytes) > 0 {
			instance.Bytes = nextStateBytes
			instance.Status = data.LayerInstanceStatusFaulty
			err = c.instancesBackend.SaveInstance(ctx, instance)
			if err != nil {
				s.Error()
				sm.Stop()
				return errors.Wrap(err, "fail to save instance of failed instance")
			}
		}

		s.Error()
		sm.Stop()
		return errors.Wrap(originalErr, "fail to terraform apply")
	}

	nextStateBytes, err := os.ReadFile(statePath)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to read next state")
	}

	instance.Bytes = nextStateBytes
	instance.Status = data.LayerInstanceStatusAlive
	err = c.instancesBackend.SaveInstance(ctx, instance)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to save instance")
	}

	s.Complete()
	sm.Stop()

	return nil
}

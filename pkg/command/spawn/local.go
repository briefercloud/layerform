package spawn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/chelnak/ysmrr"
	"github.com/chelnak/ysmrr/pkg/animations"
	"github.com/chelnak/ysmrr/pkg/colors"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/terraform"
	"github.com/ergomake/layerform/internal/tfclient"
	"github.com/ergomake/layerform/pkg/command"
	"github.com/ergomake/layerform/pkg/data"
	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type localSpawnCommand struct {
	definitionsBackend layerdefinitions.Backend
	instancesBackend   layerinstances.Backend
}

var _ Spawn = &localSpawnCommand{}

func NewLocal(definitionsBackend layerdefinitions.Backend, instancesBackend layerinstances.Backend) *localSpawnCommand {
	return &localSpawnCommand{definitionsBackend, instancesBackend}
}

func (c *localSpawnCommand) Run(
	ctx context.Context,
	layerName, instanceName string,
	dependenciesInstance map[string]string,
	vars []string,
) error {
	logger := hclog.FromContext(ctx)

	tfpath, err := terraform.GetTFPath(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to get terraform path")
	}
	logger.Debug("Using terraform from", "tfpath", tfpath)

	logger.Debug("Creating a temporary work directory")
	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.Wrap(err, "fail to create work directory")
	}
	defer os.RemoveAll(workdir)

	_, err = c.instancesBackend.GetInstance(ctx, layerName, instanceName)
	if err == nil {
		return errors.Errorf("layer %s already spawned with name %s", layerName, instanceName)
	}
	if !errors.Is(err, layerinstances.ErrInstanceNotFound) {
		return errors.Wrap(err, "fail to get instance")
	}

	err = c.spawnLayer(ctx, layerName, instanceName, workdir, tfpath, dependenciesInstance, vars)
	if err != nil {
		return errors.Wrap(err, "fail to spawn layer")
	}

	return nil
}

func getStateDiff(a *tfjson.State, b *tfjson.State) []string {
	aAddrs := command.GetStateModuleAddresses(a.Values.RootModule)
	resourceMap := make(map[string]struct{})
	for _, addr := range aAddrs {
		resourceMap[addr] = struct{}{}
	}

	diff := make([]string, 0)
	for _, addr := range command.GetStateModuleAddresses(b.Values.RootModule) {
		if _, found := resourceMap[addr]; !found {
			diff = append(diff, addr)
		}
	}

	return diff
}

func mergeTFState(ctx context.Context, tfpath, basePath, dest string, states ...string) error {
	hclog.FromContext(ctx).Debug("Merging terraform state", "base", basePath, "dest", dest, "states", states)
	dir := filepath.Dir(basePath)

	err := copyFile(basePath, dest)
	if err != nil {
		return errors.Wrap(err, "fail to copy file")
	}

	aState, err := command.GetTFState(ctx, basePath, tfpath)
	if err != nil {
		return errors.Wrap(err, "fail to get base tf state")
	}

	addedAddress := make(map[string]struct{})
	for _, bPath := range states {
		bState, err := command.GetTFState(ctx, bPath, tfpath)
		if err != nil {
			return errors.Wrap(err, "fail to get tf state")
		}

		diff := getStateDiff(aState, bState)

		tf, err := tfclient.New(dir, tfpath)
		if err != nil {
			return errors.Wrap(err, "fail to create terraform client")
		}

		for _, item := range diff {
			if _, ok := addedAddress[item]; ok {
				continue
			}

			//lint:ignore SA1019 tfexec.State is deprecated but the workaround does not support our use case
			err = tf.StateMv(ctx, item, item, tfexec.State(bPath), tfexec.StateOut(dest))
			if err != nil {
				return errors.Wrapf(err, "fail to move state %s out of %s to %s", item, bPath, dest)
			}
			addedAddress[item] = struct{}{}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrapf(err, "fail to read %s", src)
	}

	if err := os.WriteFile(dst, b, 0644); err != nil {
		return errors.Wrapf(err, "fail to write to %s", dst)
	}

	return nil
}

func (c *localSpawnCommand) spawnLayer(
	ctx context.Context,
	layerName, instanceName, workdir, tfpath string,
	dependenciesInstance map[string]string,
	vars []string,
) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Start spawning layer")

	visited := make(map[string]string)

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()

	var inner func(layerName, instanceName, layerWorkdir string) (string, error)
	inner = func(layerName, instanceName, layerWorkdir string) (string, error) {
		logger = logger.With("layer", layerName, "instance", instanceName, "layerWorkdir", layerWorkdir)
		logger.Debug("Spawning layer")

		if st, ok := visited[layerName]; ok {
			logger.Debug("Layer already spawned before")
			return st, nil
		}

		err := os.Mkdir(layerWorkdir, os.ModePerm)
		if err != nil {
			return "", errors.Wrap(err, "fail to create sub work directory for layer")
		}

		layer, err := c.definitionsBackend.GetLayer(ctx, layerName)
		if err != nil {
			return "", errors.Wrap(err, "fail to get layer")
		}

		if layer == nil {
			return "", errors.New("layer not found")
		}

		thisLayerDepInstances := map[string]string{}
		for _, dep := range layer.Dependencies {
			thisLayerDepInstances[dep] = dependenciesInstance[dep]
			if thisLayerDepInstances[dep] == "" {
				thisLayerDepInstances[dep] = "default"
			}
		}

		instanceByLayer := map[string]string{}
		instanceByLayer[layer.Name] = instanceName
		for k, v := range thisLayerDepInstances {
			instanceByLayer[k] = v
		}

		s := sm.AddSpinner(fmt.Sprintf("Preparing instance \"%s\" of layer \"%s\"", instanceName, layerName))

		layerWorkdir, err = command.WriteLayerToWorkdir(ctx, c.definitionsBackend, layerWorkdir, layer, instanceByLayer)
		if err != nil {
			return "", errors.Wrap(err, "fail to write layer to workdir")
		}

		tf, err := tfclient.New(layerWorkdir, tfpath)
		if err != nil {
			return "", errors.Wrap(err, "fail to get terraform client")
		}

		err = tf.Init(ctx, layer.SHA)
		if err != nil {
			s.Error()
			return "", errors.Wrap(err, "fail to terraform init")
		}

		statePath := path.Join(layerWorkdir, "terraform.tfstate")
		err = os.WriteFile(statePath, []byte{}, 0644)
		if err != nil {
			s.Error()
			return "", errors.Wrap(err, "fail to create empty terraform state")
		}

		depStates := []string{}
		for _, dep := range layer.Dependencies {
			layerWorkdir := path.Join(workdir, dep)

			depInstanceName := dependenciesInstance[dep]
			if depInstanceName == "" {
				depInstanceName = data.DEFAULT_LAYER_INSTANCE_NAME
			} else {
				thisLayerDepInstances[dep] = depInstanceName
			}

			depState, err := inner(dep, depInstanceName, layerWorkdir)
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to launch dependency layer")
			}

			depStates = append(depStates, depState)
		}

		instance, err := c.instancesBackend.GetInstance(ctx, layerName, instanceName)
		if err == nil {
			err := os.WriteFile(statePath, instance.Bytes, 0644)
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to write layer instance to layer work dir")
			}

			depStates = append(depStates, statePath)
		}

		if err != nil && !errors.Is(err, layerinstances.ErrInstanceNotFound) {
			s.Error()
			return "", errors.Wrap(err, "fail to get layer instance")
		}

		if len(depStates) > 1 {
			destFile, err := os.CreateTemp("", "")
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to create temp file to use as output of merged state")
			}
			defer destFile.Close()
			defer os.Remove(destFile.Name())

			base := depStates[0]
			rest := depStates[1:]
			err = mergeTFState(ctx, tfpath, base, destFile.Name(), rest...)
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to merge states")
			}

			err = copyFile(destFile.Name(), statePath)
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to copy merged state into state path")
			}
		} else if len(depStates) > 0 {
			err = copyFile(depStates[0], statePath)
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to copy base state into state path")
			}
		}

		s.Complete()

		logger.Debug("Looking for variable definitions in .tfvars files")
		varFiles, err := command.FindTFVarFiles()
		if err != nil {
			return "", errors.Wrap(err, "fail to find .tfvars files")
		}
		logger.Debug(fmt.Sprintf("Found %d var files", len(varFiles)), "varFiles", varFiles)

		applyOptions := []tfexec.ApplyOption{}
		for _, vf := range varFiles {
			applyOptions = append(applyOptions, tfexec.VarFile(vf))
		}
		for _, v := range vars {
			applyOptions = append(applyOptions, tfexec.Var(v))
		}

		verb := "Spawning"
		if instance != nil {
			verb = "Refreshing"
		}
		s = sm.AddSpinner(fmt.Sprintf("%s instance \"%s\" of layer \"%s\"", verb, instanceName, layerName))

		var nextStateBytes []byte
		if instance == nil || !bytes.Equal(layer.SHA, instance.DefinitionSHA) {
			logger.Debug("Running terraform apply")
			err = tf.Apply(ctx, applyOptions...)
			if err != nil {
				s.Error()

				originalErr := err

				nextStateBytes, err = os.ReadFile(statePath)
				if err != nil {
					return "", errors.Wrap(err, "fail to read next state")
				}

				// if this spawn attempt generated state, we should save it as faulty
				// so user can fix it
				if len(nextStateBytes) > 0 {
					instance = &data.LayerInstance{
						DefinitionSHA:        layer.SHA,
						DefinitionName:       layerName,
						InstanceName:         instanceName,
						DependenciesInstance: thisLayerDepInstances,
						Bytes:                nextStateBytes,
						Status:               data.LayerInstanceStatusFaulty,
						Version:              data.CURRENT_INSTANCE_VERSION,
					}
					err = c.instancesBackend.SaveInstance(ctx, instance)
					if err != nil {
						return "", errors.Wrap(err, "fail to save instance of failed instance")
					}
				}

				return "", errors.Wrap(originalErr, "fail to terraform apply")
			}

			nextStateBytes, err = os.ReadFile(statePath)
			if err != nil {
				s.Error()
				return "", errors.Wrap(err, "fail to read next state")
			}

		} else {
			nextStateBytes = instance.Bytes
		}

		instance = &data.LayerInstance{
			DefinitionSHA:        layer.SHA,
			DefinitionName:       layerName,
			InstanceName:         instanceName,
			DependenciesInstance: thisLayerDepInstances,
			Bytes:                nextStateBytes,
			Status:               data.LayerInstanceStatusAlive,
			Version:              data.CURRENT_INSTANCE_VERSION,
		}
		err = c.instancesBackend.SaveInstance(ctx, instance)
		if err != nil {
			s.Error()
			return "", errors.Wrap(err, "fail to save instance")
		}

		s.Complete()
		visited[layerName] = statePath
		return visited[layerName], nil
	}

	layerWorkdir := path.Join(workdir, layerName)
	_, err := inner(layerName, instanceName, layerWorkdir)

	sm.Stop()
	return err
}

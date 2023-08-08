package command

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
)

type spawnCommand struct {
	layersBackend layers.Backend
	statesBackend layerstate.Backend
}

func NewSpawn(layersBackend layers.Backend, statesBackend layerstate.Backend) *spawnCommand {
	return &spawnCommand{layersBackend, statesBackend}
}

func (c *spawnCommand) Run(layerName, stateName string, dependenciesState map[string]string) error {
	logger := hclog.Default()
	logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
	if logLevel != hclog.NoLevel {
		logger.SetLevel(logLevel)
	}
	ctx := hclog.WithContext(context.Background(), logger)

	logger.Debug("Finding terraform installation")
	i := install.NewInstaller()
	i.SetLogger(logger.StandardLogger(&hclog.StandardLoggerOptions{
		ForceLevel: hclog.Debug,
	}))
	tfpath, err := i.Ensure(ctx, []src.Source{
		&fs.Version{
			Product:     product.Terraform,
			Constraints: version.MustConstraints(version.NewConstraint(">=1.1.0")),
		},
	})
	if err != nil {
		return errors.Wrap(err, "fail to ensure terraform")
	}
	logger.Debug("Found terraform installation", "tfpath", tfpath)

	logger.Debug("Creating a temporary work directory")
	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.Wrap(err, "fail to create work directory")
	}
	defer os.RemoveAll(workdir)

	err = c.spawnLayer(ctx, layerName, stateName, workdir, tfpath, dependenciesState)
	if err != nil {
		return errors.Wrap(err, "fail to spawn layer")
	}

	return nil
}

func getStateDiff(a *tfjson.State, b *tfjson.State) []string {
	aAddrs := getStateModuleAddresses(a.Values.RootModule)
	resourceMap := make(map[string]struct{})
	for _, addr := range aAddrs {
		resourceMap[addr] = struct{}{}
	}

	diff := make([]string, 0)
	for _, addr := range getStateModuleAddresses(b.Values.RootModule) {
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

	aState, err := getTFState(ctx, basePath, tfpath)
	if err != nil {
		return errors.Wrap(err, "fail to get tf state")
	}

	addedAddress := make(map[string]struct{})
	for _, bPath := range states {
		bState, err := getTFState(ctx, bPath, tfpath)
		if err != nil {
			return errors.Wrap(err, "fail to get tf state")
		}

		diff := getStateDiff(aState, bState)

		tf, err := tfexec.NewTerraform(dir, tfpath)
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

func (c *spawnCommand) spawnLayer(
	ctx context.Context,
	layerName, stateName, workdir, tfpath string,
	dependenciesState map[string]string,
) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Start spawning layer")

	visited := make(map[string]string)

	var inner func(layerName, stateName, layerWorkdir string) (string, error)
	inner = func(layerName, stateName, layerWorkdir string) (string, error) {
		logger = logger.With("layer", layerName, "state", stateName, "layerWorkdir", layerWorkdir)
		logger.Debug("Spawning layer")

		thisLayerDepStates := map[string]string{}

		if st, ok := visited[layerName]; ok {
			logger.Debug("Layer already spawned before")
			return st, nil
		}

		err := os.Mkdir(layerWorkdir, os.ModePerm)
		if err != nil {
			return "", errors.Wrap(err, "fail to create sub work directory for layer")
		}

		layer, err := c.layersBackend.GetLayer(ctx, layerName)
		if err != nil {
			return "", errors.Wrap(err, "fail to get layer")
		}

		if layer == nil {
			return "", errors.New("layer not found")
		}

		layerWorkdir, err = writeLayerToWorkdir(ctx, c.layersBackend, layerWorkdir, layer)
		if err != nil {
			return "", errors.Wrap(err, "fail to write layer to workdir")
		}

		tf, err := tfexec.NewTerraform(layerWorkdir, tfpath)
		if err != nil {
			return "", errors.Wrap(err, "fail to get terraform client")
		}

		logger.Debug("Running terraform init")
		err = tf.Init(ctx)
		if err != nil {
			return "", errors.Wrap(err, "fail to terraform init")
		}

		statePath := path.Join(layerWorkdir, "terraform.tfstate")
		err = os.WriteFile(statePath, []byte{}, 0644)
		if err != nil {
			return "", errors.Wrap(err, "fail to create empty terraform state")
		}

		depStates := []string{}
		for _, dep := range layer.Dependencies {
			layerWorkdir := path.Join(workdir, dep)

			depStateName := dependenciesState[dep]
			if depStateName == "" {
				depStateName = layerstate.DEFAULT_LAYER_STATE_NAME
			} else {
				thisLayerDepStates[dep] = depStateName
			}

			depState, err := inner(dep, depStateName, layerWorkdir)
			if err != nil {
				return "", errors.Wrap(err, "fail to launch dependency layer")
			}

			depStates = append(depStates, depState)
		}

		state, err := c.statesBackend.GetState(ctx, layerName, stateName)
		if err == nil {
			err := os.WriteFile(statePath, state.Bytes, 0644)
			if err != nil {
				return "", errors.Wrap(err, "fail to write layer state to layer work dir")
			}

			depStates = append(depStates, statePath)
		}

		if err != nil && !errors.Is(err, layerstate.ErrStateNotFound) {
			return "", errors.Wrap(err, "fail to get layer state")
		}

		if len(depStates) > 1 {
			destFile, err := os.CreateTemp("", "")
			if err != nil {
				return "", errors.Wrap(err, "fail to create temp file to use as output of merged state")
			}
			defer destFile.Close()
			defer os.Remove(destFile.Name())

			base := depStates[0]
			rest := depStates[1:]
			err = mergeTFState(ctx, tfpath, base, destFile.Name(), rest...)
			if err != nil {
				return "", errors.Wrap(err, "fail to merge states")
			}

			err = copyFile(destFile.Name(), statePath)
			if err != nil {
				return "", errors.Wrap(err, "fail to copy merged state into state path")
			}
		} else if len(depStates) > 0 {
			err = copyFile(depStates[0], statePath)
			if err != nil {
				return "", errors.Wrap(err, "fail to copy base state into state path")
			}
		}

		logger.Debug("Running terraform apply")
		err = tf.Apply(ctx)
		if err != nil {
			return "", errors.Wrap(err, "fail to terraform apply")
		}

		nextStateBytes, err := os.ReadFile(statePath)
		if err != nil {
			return "", errors.Wrap(err, "fail to read next state")
		}

		state = &layerstate.State{
			LayerName:         layerName,
			StateName:         stateName,
			DependenciesState: thisLayerDepStates,
			Bytes:             nextStateBytes,
		}
		err = c.statesBackend.SaveState(ctx, state)
		if err != nil {
			return "", errors.Wrap(err, "fail to save state")
		}

		visited[layerName] = statePath
		return visited[layerName], nil
	}

	layerWorkdir := path.Join(workdir, layerName)
	_, err := inner(layerName, stateName, layerWorkdir)
	return err
}

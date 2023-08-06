package command

import (
	"context"
	"fmt"
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
	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
	"github.com/ergomake/layerform/internal/pathutils"
)

type launchCommand struct {
	layersBackend layers.Backend
	statesBackend layerstate.Backend
}

var _ cli.Command = &launchCommand{}

func NewLaunch(layersBackend layers.Backend, statesBackend layerstate.Backend) *launchCommand {
	return &launchCommand{layersBackend, statesBackend}
}

func (c *launchCommand) Help() string {
	return "launch help"
}

func (c *launchCommand) Synopsis() string {
	return "launch synopsis"
}

func (c *launchCommand) Run(args []string) int {
	layerName := args[0]

	stateName := ""
	if len(args) > 1 {
		stateName = args[1]
	} else {
		stateName = shortuuid.New()
	}

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
		fmt.Println("fail to ensure terraform", err)
		return 1
	}
	logger.Debug("Found terraform installation", "tfpath", tfpath)

	logger.Debug("Creating a temporary work directory")
	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Println("fail to create work directory", err)
		return 1
	}
	defer os.RemoveAll(workdir)

	err = c.spawnLayer(ctx, layerName, stateName, workdir, tfpath)
	if err != nil {
		fmt.Println("fail to spawn layer", err)
		return 1
	}

	return 0
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

func (c *launchCommand) spawnLayer(ctx context.Context, layerName, stateName, workdir, tfpath string) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Start spawning layer")

	visited := make(map[string]string)

	var inner func(layerName, stateName, layerWorkdir string) (string, error)
	inner = func(layerName, stateName, layerWorkdir string) (string, error) {
		logger = logger.With("layer", layerName, "state", stateName, "layerWorkdir", layerWorkdir)

		logger.Debug("Spawning layer")
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

		layerWorkdir, err = c.writeLayerToWorkdir(ctx, layerWorkdir, layer)
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

			depState, err := inner(dep, "default", layerWorkdir)
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

		nextState, err := os.ReadFile(statePath)
		if err != nil {
			return "", errors.Wrap(err, "fail to read next state")
		}

		err = c.statesBackend.SaveState(ctx, layerName, stateName, nextState)
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

func (c *launchCommand) writeLayerToWorkdir(ctx context.Context, layerWorkdir string, layer *model.Layer) (string, error) {
	logger := hclog.FromContext(ctx).With("layer", layer.Name, "layerWorkdir", layerWorkdir)
	logger.Debug("Writting layer to workdir")

	var inner func(*model.Layer) ([]string, error)
	inner = func(layer *model.Layer) ([]string, error) {
		fpaths := make([]string, 0)
		for _, dep := range layer.Dependencies {
			logger.Debug("Writting dependency to workdir", "dependency", dep)

			layer, err := c.layersBackend.GetLayer(ctx, dep)
			if err != nil {
				return nil, errors.Wrap(err, "fail to get layer")
			}

			depPaths, err := inner(layer)
			if err != nil {
				return nil, errors.Wrap(err, "fail to write layer to workdir")
			}

			fpaths = append(fpaths, depPaths...)
		}

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
		}

		return fpaths, nil
	}

	paths, err := inner(layer)
	if err != nil {
		return "", errors.Wrap(err, "fail to write layer to workdir")
	}

	return path.Join(layerWorkdir, pathutils.FindCommonParentPath(paths)), nil
}

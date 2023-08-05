package command

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

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
	"github.com/ergomake/layerform/internal/pathutils"
	"github.com/ergomake/layerform/internal/states"
)

type launchCommand struct {
	layers layers.Backend
	states states.Backend
}

var _ cli.Command = &launchCommand{}

func NewLaunch(layers layers.Backend, states states.Backend) *launchCommand {
	return &launchCommand{layers, states}
}

func (c *launchCommand) Help() string {
	return "launch help"
}

func (c *launchCommand) Synopsis() string {
	return "launch synopsis"
}

func (c *launchCommand) Run(args []string) int {
	ctx := context.Background()

	layerName := args[0]

	stateName := ""
	if len(args) > 1 {
		stateName = args[1]
	} else {
		stateName = shortuuid.New()
	}

	i := install.NewInstaller()
	i.SetLogger(log.Default())
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

	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Println("fail to create work directory", err)
		return 1
	}
	fmt.Println(workdir)
	// defer os.RemoveAll(workdir)

	err = c.spawnLayer(ctx, layerName, stateName, workdir, tfpath)
	if err != nil {
		fmt.Println("fail to spawn layer", err)
		return 1
	}

	return 0
}

func GetTFState(ctx context.Context, statePath string, tfpath string) (*tfjson.State, error) {
	dir := filepath.Dir(statePath)
	tf, err := tfexec.NewTerraform(dir, tfpath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create terraform client")
	}

	return tf.ShowStateFile(ctx, statePath)
}

func Addresses(module *tfjson.StateModule) []string {
	addresses := make([]string, 0)
	for _, res := range module.Resources {
		addresses = append(addresses, res.Address)
	}

	for _, child := range module.ChildModules {
		addresses = append(addresses, Addresses(child)...)
	}

	return addresses
}

func StateDiff(a *tfjson.State, b *tfjson.State) []string {
	aAddrs := Addresses(a.Values.RootModule)
	resourceMap := make(map[string]struct{})
	for _, addr := range aAddrs {
		resourceMap[addr] = struct{}{}
	}

	diff := make([]string, 0)
	for _, addr := range Addresses(b.Values.RootModule) {
		if _, found := resourceMap[addr]; !found {
			diff = append(diff, addr)
		}
	}

	return diff
}

func MergeState(ctx context.Context, tfpath string, basePath string, dest string, states ...string) error {
	dir := filepath.Dir(basePath)

	err := copyFile(basePath, dest)
	if err != nil {
		return errors.Wrap(err, "fail to copy file")
	}

	aState, err := GetTFState(ctx, basePath, tfpath)
	if err != nil {
		return errors.Wrap(err, "fail to get tf state")
	}

	addedAddress := make(map[string]struct{})
	for _, bPath := range states {
		fmt.Printf("merging state %s into %s\n", bPath, basePath)
		bState, err := GetTFState(ctx, bPath, tfpath)
		if err != nil {
			return errors.Wrap(err, "fail to get tf state")
		}

		diff := StateDiff(aState, bState)
		fmt.Printf("diff %+v\n", diff)

		tf, err := tfexec.NewTerraform(dir, tfpath)
		if err != nil {
			return errors.Wrap(err, "fail to create terraform client")
		}

		for _, item := range diff {
			if _, ok := addedAddress[item]; ok {
				continue
			}

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
	type statet struct {
		path string
	}

	visited := make(map[string]string)

	var inner func(layerName, stateName, layerWorkdir string) (string, error)
	inner = func(layerName, stateName, layerWorkdir string) (string, error) {
		fmt.Printf("Spawning layer %s, state %s\n", layerName, stateName)
		if st, ok := visited[layerName]; ok {
			fmt.Printf("Layer %s was cached\n", layerName)
			return st, nil
		}

		err := os.Mkdir(layerWorkdir, os.ModePerm)
		if err != nil {
			return "", errors.Wrap(err, "fail to create sub work directory for layer")
		}

		layer, err := c.layers.GetLayer(layerName)
		if err != nil {
			return "", errors.Wrap(err, "fail to get layer")
		}

		if layer == nil {
			return "", errors.New("layer not found")
		}

		layerWorkdir, err = c.writeLayerToWorkdir(layerWorkdir, layer)
		if err != nil {
			return "", errors.Wrap(err, "fail to write layer to workdir")
		}

		tf, err := tfexec.NewTerraform(layerWorkdir, tfpath)
		if err != nil {
			return "", errors.Wrap(err, "fail to get terraform client")
		}

		fmt.Printf("Initting layer %s\n", layerName)
		err = tf.Init(ctx)
		if err != nil {
			return "", errors.Wrap(err, "fail to terraform init")
		}
		fmt.Printf("Layer %s initted\n", layerName)

		statePath := path.Join(layerWorkdir, "terraform.tfstate")
		err = os.WriteFile(statePath, []byte{}, 0644)
		if err != nil {
			return "", errors.Wrap(err, "fail to create empty terraform state")
		}

		fmt.Printf("Spawnning %d dependencies of layer %s first\n", len(layer.Dependencies), layerName)
		depStates := []string{}
		for _, dep := range layer.Dependencies {
			layerWorkdir := path.Join(workdir, dep)

			depState, err := inner(dep, "default", layerWorkdir)
			if err != nil {
				return "", errors.Wrap(err, "fail to launch dependency layer")
			}

			depStates = append(depStates, depState)
		}
		fmt.Printf("dependencies of layer %s spawned\n", layerName)

		state, err := c.states.GetState(layerName, stateName)
		if err == nil {
			err := os.WriteFile(statePath, state.Bytes, 0644)
			if err != nil {
				return "", errors.Wrap(err, "fail to write layer state to layer work dir")
			}

			fmt.Printf("Layer %s, state %s already exists, appending state to merge\n", layerName, stateName)
			depStates = append(depStates, statePath)
		}

		if err != nil && !errors.Is(err, states.ErrStateNotFound) {
			return "", errors.Wrap(err, "fail to get layer state")
		}

		if len(depStates) > 1 {
			fmt.Printf("Layer %s has more than 1 dependency state, merging states\n", layerName)
			fmt.Printf("%+v\n", depStates)
			destFile, err := os.CreateTemp("", "")
			if err != nil {
				return "", errors.Wrap(err, "fail to create temp file to use as output of merged state")
			}
			defer destFile.Close()
			defer os.Remove(destFile.Name())

			base := depStates[0]
			rest := depStates[1:]
			err = MergeState(ctx, tfpath, base, destFile.Name(), rest...)
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

		fmt.Printf("Applying layer %s, state %s\n", layerName, stateName)
		err = tf.Apply(ctx)
		if err != nil {
			return "", errors.Wrap(err, "fail to terraform apply")
		}
		fmt.Printf("Layer %s, state %s applied\n", layerName, stateName)

		nextState, err := os.ReadFile(statePath)
		if err != nil {
			return "", errors.Wrap(err, "fail to read next state")
		}

		err = c.states.SaveState(layerName, stateName, nextState)
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

func (c *launchCommand) writeLayerToWorkdir(layerWorkdir string, layer *model.Layer) (string, error) {
	var inner func(*model.Layer) ([]string, error)
	inner = func(layer *model.Layer) ([]string, error) {
		fpaths := make([]string, 0)
		for _, dep := range layer.Dependencies {
			layer, err := c.layers.GetLayer(dep)
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

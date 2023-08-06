package command

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
)

type killCommand struct {
	layersBackend layers.Backend
	statesBackend layerstate.Backend
}

var _ cli.Command = &killCommand{}

func NewKill(layersBackend layers.Backend, statesBackend layerstate.Backend) *killCommand {
	return &killCommand{layersBackend, statesBackend}
}

func (c *killCommand) Help() string {
	return "kill help"
}

func (c *killCommand) Synopsis() string {
	return "kill synopsis"
}

func (c *killCommand) Run(args []string) int {
	layerName := args[0]
	stateName := "default"
	if len(args) > 1 {
		stateName = args[1]
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
	fmt.Println(workdir)
	// defer os.RemoveAll(workdir)

	layer, err := c.layersBackend.GetLayer(ctx, layerName)
	if err != nil {
		fmt.Println("Fail to get layer", err)
		return 1
	}

	if layer == nil {
		fmt.Println("Layer not found")
		return 1
	}

	layerDir := path.Join(workdir, layerName)
	layerAddrs, layerDir, err := c.getLayerAddresses(ctx, layer, layerDir, tfpath, stateName)
	if err != nil {
		fmt.Println("Fail to get layer addresses", err)
		return 1
	}

	layerAddrsMap := make(map[string]struct{})
	for _, addr := range layerAddrs {
		layerAddrsMap[addr] = struct{}{}
	}

	for _, dep := range layer.Dependencies {
		layer, err := c.layersBackend.GetLayer(ctx, dep)
		if err != nil {
			fmt.Println("Fail to get dependency layer", err)
			return 1
		}

		if layer == nil {
			fmt.Println("Dependency layer not found", err)
			return 1
		}

		depDir := path.Join(workdir, dep)
		depAddrs, _, err := c.getLayerAddresses(ctx, layer, depDir, tfpath, "default")
		if err != nil {
			fmt.Println("Fail to get dependency layer addresses", err)
			return 1
		}

		for _, addr := range depAddrs {
			delete(layerAddrsMap, addr)
		}
	}

	tf, err := tfexec.NewTerraform(layerDir, tfpath)
	if err != nil {
		fmt.Println("Fail to get terraform client", err)
		return 1
	}

	targets := make([]tfexec.DestroyOption, 0)
	for addr := range layerAddrsMap {
		targets = append(targets, tfexec.Target(addr))
	}
	logger.Debug(
		"Running terraform destroy targetting layer specific addresses",
		"layer", layer.Name, "state", stateName, "targets", targets,
	)

	err = tf.Destroy(ctx, targets...)
	if err != nil {
		fmt.Println("Fail to terraform destroy", err)
		return 1
	}

	err = c.statesBackend.DeleteState(ctx, layerName, stateName)
	if err != nil {
		fmt.Println("Fail to delete state", err)
		return 1
	}

	return 0
}

func (c *killCommand) getLayerAddresses(
	ctx context.Context,
	layer *model.Layer,
	layerDir, tfpath, stateName string,
) ([]string, string, error) {
	logger := hclog.FromContext(ctx)
	logger.Debug("Getting layer addresses", "layer", layer.Name, "state", stateName)

	layerWorkdir, err := writeLayerToWorkdir(ctx, c.layersBackend, layerDir, layer)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to write layer to work directory")
	}

	state, err := c.statesBackend.GetState(ctx, layer.Name, stateName)
	if err != nil {
		if errors.Is(err, layerstate.ErrStateNotFound) {
			return nil, "", errors.Errorf(
				"State %s not found for layer %s\n",
				stateName,
				layer.Name,
			)
		}

		return nil, "", errors.Wrap(err, "fail to get layer state")
	}

	statePath := path.Join(layerWorkdir, "terraform.tfstate")
	err = os.WriteFile(statePath, state.Bytes, 0644)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to write terraform state to work directory")
	}

	tf, err := tfexec.NewTerraform(layerWorkdir, tfpath)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to get terraform client")
	}

	logger.Debug("Running terraform init", "layer", layer.Name, "state", stateName)
	err = tf.Init(ctx)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to terraform init")
	}

	tfState, err := getTFState(ctx, statePath, tfpath)
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to get terraform state")
	}

	addresses := getStateModuleAddresses(tfState.Values.RootModule)

	return addresses, layerWorkdir, nil
}

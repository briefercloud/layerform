package lfconfig

import (
	"context"
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
)

type configFile struct {
	CurrentContext string                   `yaml:"currentContext"`
	Contexts       map[string]configContext `yaml:"contexts"`
}

type configContext struct {
	Type string `yaml:"type"`
	Dir  string `yaml:"dir,omitempty"`
}

func getDefaultPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "fail to get user home dir")
	}

	return path.Join(homedir, ".layerform", "config"), nil
}

type config struct {
	*configFile
	path string
}

func Load(path string) (*config, error) {
	if path == "" {
		p, err := getDefaultPath()
		if err != nil {
			return nil, errors.Wrap(err, "fail to get default path")
		}

		path = p
	}

	var cfg configFile

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read config file")
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode config content")
	}

	if _, ok := cfg.Contexts[cfg.CurrentContext]; !ok {
		return nil, errors.Errorf("context %s not found", cfg.CurrentContext)
	}

	return &config{configFile: &cfg, path: path}, nil
}

func (c *config) getDir() string {
	currentContext := c.Contexts[c.CurrentContext]
	dir := currentContext.Dir
	if !path.IsAbs(dir) {
		dir = path.Join(path.Dir(c.path), dir)
	}

	return dir
}

func (c *config) GetStateBackend() layerstate.Backend {
	return layerstate.NewFileBackend(path.Join(c.getDir(), "layerform.lftstate"))
}

func (c *config) GetLayersBackend(ctx context.Context) (layers.Backend, error) {
	return layers.NewFileBackend(ctx, path.Join(c.getDir(), "layerform.definitions.json"))
}

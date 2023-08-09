package lfconfig

import (
	"context"
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/ergomake/layerform/internal/layers"
	"github.com/ergomake/layerform/internal/layerstate"
	"github.com/ergomake/layerform/internal/storage"
)

type configFile struct {
	CurrentContext string                   `yaml:"currentContext"`
	Contexts       map[string]configContext `yaml:"contexts"`
}

type configContext struct {
	Type   string `yaml:"type"`
	Dir    string `yaml:"dir,omitempty"`
	Bucket string `yaml:"bucket,omitempty"`
	Region string `yaml:"region,omitempty"`
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

func (c *config) getCurrent() configContext {
	return c.Contexts[c.CurrentContext]
}

func (c *config) getDir() string {
	dir := c.getCurrent().Dir
	if !path.IsAbs(dir) {
		dir = path.Join(path.Dir(c.path), dir)
	}

	return dir
}

const stateFileName = "layerform.lfstate"

func (c *config) GetStateBackend(ctx context.Context) (layerstate.Backend, error) {
	current := c.getCurrent()
	var blob storage.FileLike
	switch current.Type {
	case "local":
		blob = storage.NewFileStorage(path.Join(c.getDir(), stateFileName))
	case "s3":
		b, err := storage.NewS3Backend(current.Bucket, stateFileName, current.Region)
		if err != nil {
			return nil, errors.Wrap(err, "fail to initialize s3 backend")
		}
		blob = b
	}

	return layerstate.NewFileLikeBackend(ctx, blob)
}

const definitionsFileName = "layerform.definitions.json"

func (c *config) GetLayersBackend(ctx context.Context) (layers.Backend, error) {
	current := c.getCurrent()
	var blob storage.FileLike
	switch current.Type {
	case "local":
		blob = storage.NewFileStorage(path.Join(c.getDir(), definitionsFileName))
	case "s3":
		b, err := storage.NewS3Backend(current.Bucket, definitionsFileName, current.Region)
		if err != nil {
			return nil, errors.Wrap(err, "fail to initialize s3 backend")
		}
		blob = b
	}

	return layers.NewFileLikeBackend(ctx, blob)
}

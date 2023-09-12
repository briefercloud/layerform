package lfconfig

import (
	"context"
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/ergomake/layerform/internal/storage"
	"github.com/ergomake/layerform/pkg/command/kill"
	"github.com/ergomake/layerform/pkg/command/refresh"
	"github.com/ergomake/layerform/pkg/command/spawn"
	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type configFile struct {
	CurrentContext string                   `yaml:"currentContext"`
	Contexts       map[string]ConfigContext `yaml:"contexts"`
}

type ConfigContext struct {
	Type     string `yaml:"type"`
	Dir      string `yaml:"dir,omitempty"`
	Bucket   string `yaml:"bucket,omitempty"`
	Region   string `yaml:"region,omitempty"`
	URL      string `yaml:"url,omitempty"`
	Email    string `yaml:"email,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func getDefaultPaths() ([]string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return []string{}, errors.Wrap(err, "fail to get user home dir")
	}

	return []string{
		path.Join(homedir, ".layerform", "config"),
		path.Join(homedir, ".layerform", "configurations.yaml"),
		path.Join(homedir, ".layerform", "configurations.yml"),
		path.Join(homedir, ".layerform", "configuration.yaml"),
		path.Join(homedir, ".layerform", "configuration.yml"),
		path.Join(homedir, ".layerform", "config.yaml"),
		path.Join(homedir, ".layerform", "config.yml"),
	}, nil
}

type config struct {
	*configFile
	path string
}

func Init(name string, ctx ConfigContext, path string) (*config, error) {
	ctxs := map[string]ConfigContext{}
	ctxs[name] = ctx
	cfgFile := &configFile{
		CurrentContext: name,
		Contexts:       ctxs,
	}
	if path == "" {
		paths, err := getDefaultPaths()
		if err != nil {
			return nil, err
		}
		path = paths[0]
	}

	return &config{
		cfgFile,
		path,
	}, nil
}

func Load(path string) (*config, error) {
	paths := []string{path}
	if path == "" {
		ps, err := getDefaultPaths()
		if err != nil {
			return nil, errors.Wrap(err, "fail to get default path")
		}

		paths = ps
	}

	var cfg configFile
	var err error

	for _, path := range paths {
		data, e := os.ReadFile(path)
		if e != nil {
			if err == nil {
				err = errors.Wrap(e, "fail to read config file")
			}
			continue
		}

		err = yaml.Unmarshal(data, &cfg)
		if e != nil {
			if err == nil {
				err = errors.Wrap(e, "fail to decode config content")
			}
			continue
		}

		if _, ok := cfg.Contexts[cfg.CurrentContext]; !ok {
			if err == nil {
				err = errors.Errorf("context %s not found", cfg.CurrentContext)
			}
			continue
		}

		return &config{configFile: &cfg, path: path}, nil
	}

	return nil, err
}

func (cfg *config) Save() error {
	data, err := yaml.Marshal(cfg.configFile)
	if err != nil {
		return errors.Wrap(err, "fail to encode config file to yaml")
	}

	err = os.MkdirAll(path.Dir(cfg.path), 0755)
	if err != nil {
		return errors.Wrap(err, "fail to create config file dir")
	}

	err = os.WriteFile(cfg.path, data, 0644)
	return errors.Wrap(err, "fail to write config file")
}

func (c *config) GetCurrent() ConfigContext {
	return c.Contexts[c.CurrentContext]
}

func (c *config) getDir() string {
	dir := c.GetCurrent().Dir
	if !path.IsAbs(dir) {
		dir = path.Join(path.Dir(c.path), dir)
	}

	return dir
}

const stateFileName = "layerform.lfstate"

func (c *config) GetInstancesBackend(ctx context.Context) (layerinstances.Backend, error) {
	current := c.GetCurrent()
	var blob storage.FileLike
	switch current.Type {
	case "local":
		blob = storage.NewFileStorage(path.Join(c.getDir(), stateFileName))
	case "cloud":
		return layerinstances.NewCloud(current.URL), nil
	case "s3":
		b, err := storage.NewS3Backend(current.Bucket, stateFileName, current.Region)
		if err != nil {
			return nil, errors.Wrap(err, "fail to initialize s3 backend")
		}
		blob = b
	}

	return layerinstances.NewFileLikeBackend(ctx, blob)
}

const definitionsFileName = "layerform.definitions.json"

func (c *config) GetDefinitionsBackend(ctx context.Context) (layerdefinitions.Backend, error) {
	current := c.GetCurrent()
	var blob storage.FileLike
	switch current.Type {
	case "cloud":
		return layerdefinitions.NewCloud(current.URL), nil
	case "local":
		blob = storage.NewFileStorage(path.Join(c.getDir(), definitionsFileName))
	case "s3":
		b, err := storage.NewS3Backend(current.Bucket, definitionsFileName, current.Region)
		if err != nil {
			return nil, errors.Wrap(err, "fail to initialize s3 backend")
		}
		blob = b
	}

	return layerdefinitions.NewFileLikeBackend(ctx, blob)
}

func (c *config) GetSpawnCommand(ctx context.Context) (spawn.Spawn, error) {
	current := c.GetCurrent()

	switch current.Type {
	case "cloud":
		return spawn.NewCloud(current.URL), nil
	case "s3":
		fallthrough
	case "local":
		layersBackend, err := c.GetDefinitionsBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get layers backend")
		}

		instancesBackend, err := c.GetInstancesBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get instance backend")
		}

		return spawn.NewLocal(layersBackend, instancesBackend), nil
	}

	return nil, errors.Errorf("fail to get spawn command unexpected context type %s", current.Type)
}

func (c *config) GetKillCommand(ctx context.Context) (kill.Kill, error) {
	current := c.GetCurrent()

	switch current.Type {
	case "cloud":
		return kill.NewCloud(current.URL), nil
	case "s3":
		fallthrough
	case "local":
		layersBackend, err := c.GetDefinitionsBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get layers backend")
		}

		instancesBackend, err := c.GetInstancesBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get instance backend")
		}

		return kill.NewLocal(layersBackend, instancesBackend), nil
	}

	return nil, errors.Errorf("fail to get kill command unexpected context type %s", current.Type)
}

func (c *config) GetRefreshCommand(ctx context.Context) (refresh.Refresh, error) {
	current := c.GetCurrent()

	switch current.Type {
	case "cloud":
		return refresh.NewCloud(current.URL), nil
	case "s3":
		fallthrough
	case "local":
		layersBackend, err := c.GetDefinitionsBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get layers backend")
		}

		instancesBackend, err := c.GetInstancesBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get instance backend")
		}

		return refresh.NewLocal(layersBackend, instancesBackend), nil
	}

	return nil, errors.Errorf("fail to get spawn command unexpected context type %s", current.Type)
}

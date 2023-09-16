package lfconfig

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/ergomake/layerform/internal/cloud"
	"github.com/ergomake/layerform/internal/storage"
	"github.com/ergomake/layerform/pkg/command/kill"
	"github.com/ergomake/layerform/pkg/command/refresh"
	"github.com/ergomake/layerform/pkg/command/spawn"
	"github.com/ergomake/layerform/pkg/envvars"
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

func (cfg *ConfigContext) Location() string {
	switch cfg.Type {
	case "local":
		return fmt.Sprintf("dir://%s", cfg.Dir)
	case "s3":
		return fmt.Sprintf("s3://%s", cfg.Bucket)
	case "cloud":
		return cfg.URL
	}

	panic("unreachable")
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
	url := strings.TrimSpace(os.Getenv("LF_CLOUD_URL"))
	email := strings.TrimSpace(os.Getenv("LF_CLOUD_EMAIL"))
	password := strings.TrimSpace(os.Getenv("LF_CLOUD_PASSWORD"))

	if url != "" && email != "" && password != "" {
		return ConfigContext{
			Type:     "cloud",
			URL:      url,
			Email:    email,
			Password: password,
		}
	}

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
		cloudClient, err := c.GetCloudClient(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get cloud client")
		}

		return layerinstances.NewCloud(cloudClient), nil
	case "s3":
		b, err := storage.NewS3Backend(current.Bucket, stateFileName, current.Region)
		if err != nil {
			return nil, errors.Wrap(err, "fail to initialize s3 backend")
		}
		blob = b
	}

	return layerinstances.NewFileLikeBackend(ctx, blob)
}

func (c *config) GetCloudClient(ctx context.Context) (*cloud.HTTPClient, error) {
	current := c.GetCurrent()
	return cloud.NewHTTPClient(ctx, current.URL, current.Email, current.Password)
}

const definitionsFileName = "layerform.definitions.json"

func (c *config) GetDefinitionsBackend(ctx context.Context) (layerdefinitions.Backend, error) {
	current := c.GetCurrent()
	var blob storage.FileLike
	switch current.Type {
	case "cloud":
		cloudClient, err := c.GetCloudClient(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get cloud client")
		}

		return layerdefinitions.NewCloud(cloudClient), nil
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
		cloudClient, err := c.GetCloudClient(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get cloud client")
		}

		return spawn.NewCloud(cloudClient), nil
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

		envVarsBackend, err := c.GetEnvVarsBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get env vars backend")
		}

		return spawn.NewLocal(layersBackend, instancesBackend, envVarsBackend), nil
	}

	return nil, errors.Errorf("fail to get spawn command unexpected context type %s", current.Type)
}

func (c *config) GetKillCommand(ctx context.Context) (kill.Kill, error) {
	current := c.GetCurrent()

	switch current.Type {
	case "cloud":
		cloudClient, err := c.GetCloudClient(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get cloud client")
		}

		return kill.NewCloud(cloudClient), nil
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

		envVarsBackend, err := c.GetEnvVarsBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get env vars backend")
		}

		return kill.NewLocal(layersBackend, instancesBackend, envVarsBackend), nil
	}

	return nil, errors.Errorf("fail to get kill command unexpected context type %s", current.Type)
}

func (c *config) GetRefreshCommand(ctx context.Context) (refresh.Refresh, error) {
	current := c.GetCurrent()

	switch current.Type {
	case "cloud":
		cloudClient, err := c.GetCloudClient(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get cloud client")
		}

		return refresh.NewCloud(cloudClient), nil
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

		envVarsBackend, err := c.GetEnvVarsBackend(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get env vars backend")
		}

		return refresh.NewLocal(layersBackend, instancesBackend, envVarsBackend), nil
	}

	return nil, errors.Errorf("fail to get spawn command unexpected context type %s", current.Type)
}

const envVarsFileName = "layerform.env"

func (c *config) GetEnvVarsBackend(ctx context.Context) (envvars.Backend, error) {
	current := c.GetCurrent()

	switch current.Type {
	case "cloud":
		cloudClient, err := c.GetCloudClient(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get cloud client")
		}

		return envvars.NewCloud(cloudClient), nil
	case "s3":
		s3, err := storage.NewS3Backend(current.Bucket, envVarsFileName, current.Region)
		if err != nil {
			return nil, errors.Wrap(err, "fail to initialize s3 backend")
		}

		return envvars.NewFileLikeBackend(ctx, s3)
	case "local":
		fileStorage := storage.NewFileStorage(path.Join(c.getDir(), envVarsFileName))
		return envvars.NewFileLikeBackend(ctx, fileStorage)
	}

	return nil, errors.Errorf("fail to get set-env command unexpected context type %s", current.Type)
}

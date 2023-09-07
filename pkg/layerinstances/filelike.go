package layerinstances

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/storage"
	"github.com/ergomake/layerform/pkg/data"
)

type version struct {
	Version uint `json:"version"`
}

const CURRENT_FILE_LIKE_MODEL_VERSION = 1

type fileLikeModelV0 struct {
	Version uint                    `json:"version"`
	States  []*data.LayerInstanceV0 `json:"states"`
}
type fileLikeModel struct {
	Version   uint                  `json:"version"`
	Instances []*data.LayerInstance `json:"instances"`
}

func (f *fileLikeModel) UnmarshalJSON(b []byte) error {
	f.Version = CURRENT_FILE_LIKE_MODEL_VERSION

	var v version
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if v.Version == CURRENT_FILE_LIKE_MODEL_VERSION {
		// need a type alias to avoid infinite recursion
		type alias fileLikeModel
		var tmp alias

		err := json.Unmarshal(b, &tmp)
		if err != nil {
			return err
		}

		*f = fileLikeModel(tmp)
		return nil
	}

	if v.Version > CURRENT_FILE_LIKE_MODEL_VERSION {
		return errors.New("instances file was created using a newer version of layerform")
	}

	if v.Version == 0 {
		var v0 fileLikeModelV0
		err := json.Unmarshal(b, &v0)
		if err != nil {
			return err
		}

		f.Instances = make([]*data.LayerInstance, len(v0.States))
		for i, s := range v0.States {
			f.Instances[i] = s.ToLayerInstance()
		}

		return nil
	}

	return errors.Errorf("got unexpected version %d of instances file", v.Version)
}

type fileLikeBackend struct {
	model   *fileLikeModel
	storage storage.FileLike
}

var _ Backend = &fileLikeBackend{}

func NewFileLikeBackend(ctx context.Context, storage storage.FileLike) (*fileLikeBackend, error) {
	finstance := fileLikeModel{
		Version: CURRENT_FILE_LIKE_MODEL_VERSION,
	}

	err := storage.Load(ctx, &finstance)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read file")
	}

	return &fileLikeBackend{model: &finstance, storage: storage}, nil
}

func (flb *fileLikeBackend) GetInstance(ctx context.Context, layerName, instanceName string) (*data.LayerInstance, error) {
	hclog.FromContext(ctx).Debug("Getting layer instance", "layer", layerName, "instance", instanceName)

	for _, instance := range flb.model.Instances {
		if instance.DefinitionName == layerName && instance.InstanceName == instanceName {
			return instance, nil
		}
	}

	return nil, errors.Wrapf(ErrInstanceNotFound, "instance %s for layer %s not found", instanceName, layerName)
}

func (flb *fileLikeBackend) SaveInstance(ctx context.Context, instance *data.LayerInstance) error {
	hclog.FromContext(ctx).Debug("Saving layer instance", "layer", instance.DefinitionName, "instance", instance.InstanceName)

	nextInstances := []*data.LayerInstance{}
	for _, s := range flb.model.Instances {
		if s.DefinitionName != instance.DefinitionName || s.InstanceName != instance.InstanceName {
			nextInstances = append(nextInstances, s)
		}
	}

	nextInstances = append(nextInstances, instance)

	flb.model.Instances = nextInstances

	return flb.storage.Save(ctx, flb.model)
}

func (flb *fileLikeBackend) DeleteInstance(ctx context.Context, layerName, instanceName string) error {
	hclog.FromContext(ctx).Debug("Deleting layer instance", "layer", layerName, "instance", instanceName)

	nextInstances := []*data.LayerInstance{}
	for _, s := range flb.model.Instances {
		if s.DefinitionName != layerName || s.InstanceName != instanceName {
			nextInstances = append(nextInstances, s)
		}
	}

	flb.model.Instances = nextInstances

	return flb.storage.Save(ctx, flb.model)
}

func (flb *fileLikeBackend) ListInstancesByLayer(ctx context.Context, layerName string) ([]*data.LayerInstance, error) {
	hclog.FromContext(ctx).Debug("Listing instances by layer", "layer", layerName)

	result := make([]*data.LayerInstance, 0)
	for _, s := range flb.model.Instances {
		if s.DefinitionName == layerName {
			result = append(result, s)
		}
	}

	return result, nil
}

func (flb *fileLikeBackend) ListInstances(ctx context.Context) ([]*data.LayerInstance, error) {
	hclog.FromContext(ctx).Debug("Listing all layers instances")

	return flb.model.Instances, nil
}

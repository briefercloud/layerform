package data

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type LayerInstanceStatus string

const (
	LayerInstanceStatusAlive  LayerInstanceStatus = LayerInstanceStatus("alive")
	LayerInstanceStatusFaulty LayerInstanceStatus = LayerInstanceStatus("faulty")
)

const DEFAULT_LAYER_INSTANCE_NAME = "default"

type version struct {
	Version uint `json:"version"`
}

const CURRENT_INSTANCE_VERSION = 1

type LayerInstanceV0 struct {
	LayerSHA          []byte              `json:"layerSHA"`
	LayerName         string              `json:"layerName"`
	StateName         string              `json:"stateName"`
	DependenciesState map[string]string   `json:"dependenciesState"`
	Bytes             []byte              `json:"bytes"`
	Status            LayerInstanceStatus `json:"status"`
}

func (v0 *LayerInstanceV0) ToLayerInstance() *LayerInstance {
	return &LayerInstance{
		DefinitionSHA:        v0.LayerSHA,
		DefinitionName:       v0.LayerName,
		InstanceName:         v0.StateName,
		DependenciesInstance: v0.DependenciesState,
		Bytes:                v0.Bytes,
		Status:               v0.Status,
		Version:              CURRENT_INSTANCE_VERSION,
	}
}

type LayerInstance struct {
	DefinitionSHA        []byte              `json:"definitionSHA"`
	DefinitionName       string              `json:"definitionName"`
	InstanceName         string              `json:"instanceName"`
	DependenciesInstance map[string]string   `json:"dependenciesInstance"`
	Bytes                []byte              `json:"bytes"`
	Status               LayerInstanceStatus `json:"status"`
	Version              uint                `json:"version"`
}

func (i *LayerInstance) UnmarshalJSON(b []byte) error {
	i.Version = CURRENT_INSTANCE_VERSION

	var v version
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if v.Version == CURRENT_INSTANCE_VERSION {
		// need a type alias to avoid infinite recursion
		type alias LayerInstance
		var tmp alias

		err := json.Unmarshal(b, &tmp)
		if err != nil {
			return err
		}

		*i = LayerInstance(tmp)
		return nil
	}

	if v.Version > CURRENT_INSTANCE_VERSION {
		return errors.New("layer instance was created using a newer version of layerform")
	}

	if v.Version == 0 {
		var v0 LayerInstanceV0
		err := json.Unmarshal(b, &v0)
		if err != nil {
			return err
		}

		*i = *v0.ToLayerInstance()
		return nil
	}

	return errors.Errorf("got unexpected version %d of layer instance", v.Version)
}

func (s *LayerInstance) GetDependencyInstanceName(dep string) string {
	if name, ok := s.DependenciesInstance[dep]; ok {
		return name
	}

	return DEFAULT_LAYER_INSTANCE_NAME
}

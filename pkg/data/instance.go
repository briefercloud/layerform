package data

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type InstanceStatus string

const (
	InstanceStatusAlive  InstanceStatus = InstanceStatus("alive")
	InstanceStatusFaulty InstanceStatus = InstanceStatus("faulty")
)

const DEFAULT_LAYER_INSTANCE_NAME = "default"

type version struct {
	Version uint `json:"version"`
}

const CURRENT_INSTANCE_VERSION = 1

type InstanceV0 struct {
	LayerSHA          []byte            `json:"layerSHA"`
	LayerName         string            `json:"layerName"`
	StateName         string            `json:"stateName"`
	DependenciesState map[string]string `json:"dependenciesState"`
	Bytes             []byte            `json:"bytes"`
	Status            InstanceStatus    `json:"status"`
}

func (v0 *InstanceV0) ToInstance() *Instance {
	return &Instance{
		DefinitionSHA:        v0.LayerSHA,
		DefinitionName:       v0.LayerName,
		InstanceName:         v0.StateName,
		DependenciesInstance: v0.DependenciesState,
		Bytes:                v0.Bytes,
		Status:               v0.Status,
		Version:              CURRENT_INSTANCE_VERSION,
	}
}

type Instance struct {
	DefinitionSHA        []byte            `json:"definitionSHA"`
	DefinitionName       string            `json:"definitionName"`
	InstanceName         string            `json:"instanceName"`
	DependenciesInstance map[string]string `json:"dependenciesInstance"`
	Bytes                []byte            `json:"bytes"`
	Status               InstanceStatus    `json:"status"`
	Version              uint              `json:"version"`
}

func (i *Instance) UnmarshalJSON(b []byte) error {
	i.Version = CURRENT_INSTANCE_VERSION

	var v version
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if v.Version == CURRENT_INSTANCE_VERSION {
		return json.Unmarshal(b, i)
	}

	if v.Version > CURRENT_INSTANCE_VERSION {
		return errors.New("layer instance was created using a newer version of layerform")
	}

	if v.Version == 0 {
		var v0 InstanceV0
		err := json.Unmarshal(b, &v0)
		if err != nil {
			return err
		}

		i.DefinitionSHA = v0.LayerSHA
		i.DefinitionName = v0.LayerName
		i.InstanceName = v0.StateName
		i.DependenciesInstance = v0.DependenciesState
		i.Bytes = v0.Bytes
		i.Status = v0.Status
		return nil
	}

	return errors.Errorf("got unexpected version %d of layer instance", v.Version)
}

func (s *Instance) GetDependencyInstanceName(dep string) string {
	if name, ok := s.DependenciesInstance[dep]; ok {
		return name
	}

	return DEFAULT_LAYER_INSTANCE_NAME
}

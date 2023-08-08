package layerstate

import (
	"context"
	"errors"
)

type State struct {
	LayerName         string            `json:"layerName"`
	StateName         string            `json:"stateName"`
	DependenciesState map[string]string `json:"dependenciesState"`
	Bytes             []byte            `json:"bytes"`
}

const DEFAULT_LAYER_STATE_NAME = "default"

func (s *State) GetDependencyStateName(dep string) string {
	if name, ok := s.DependenciesState[dep]; ok {
		return name
	}

	return DEFAULT_LAYER_STATE_NAME
}

var ErrStateNotFound = errors.New("state not found")

type Backend interface {
	GetState(ctx context.Context, layerName, stateName string) (*State, error)
	ListStatesByLayer(ctx context.Context, layerName string) ([]*State, error)
	SaveState(ctx context.Context, state *State) error
	DeleteState(ctx context.Context, layerName, stateName string) error
}

package layerstate

import (
	"context"
	"errors"
)

type State struct {
	LayerName string `json:"layerName"`
	StateName string `json:"stateName"`
	Bytes     []byte `json:"bytes"`
}

var ErrStateNotFound = errors.New("state not found")

type Backend interface {
	GetState(ctx context.Context, layerName, stateName string) (*State, error)
	SaveState(ctx context.Context, layerName, stateName string, bytes []byte) error
}

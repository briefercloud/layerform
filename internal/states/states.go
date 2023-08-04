package states

import "errors"

type State struct {
	LayerName string `json:"layerName"`
	StateName string `json:"stateName"`
	Bytes     []byte `json:"bytes"`
}

var ErrStateNotFound = errors.New("state not found")

type Backend interface {
	GetState(layerName, stateName string) (*State, error)
	SaveState(layerName, stateName string, bytes []byte) error
}

package state

import (
	"github.com/ergomake/layerform/internal/data/model"
	"github.com/ergomake/layerform/internal/terraform"
)

type State struct {
	*terraform.State
}

func NewState(tfState *terraform.State) *State {
	return &State{tfState}
}

func (s *State) Terraform() *terraform.State {
	if s == nil {
		return nil
	}

	return s.State
}

type Backend interface {
	GetLayerState(layer *model.Layer, instance string) (*State, error)
	SaveLayerState(layer *model.Layer, instance string, state *State) error
	RemoveLayerState(layer *model.Layer, instance string) error
}

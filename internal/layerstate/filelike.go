package layerstate

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/storage"
)

const fileLikeModelVersion = 0

type fileLikeModel struct {
	Version uint     `json:"version"`
	States  []*State `json:"states"`
}

type fileLikeBackend struct {
	model   *fileLikeModel
	storage storage.FileLike
}

var _ Backend = &fileLikeBackend{}

func NewFileLikeBackend(ctx context.Context, storage storage.FileLike) (*fileLikeBackend, error) {
	fstate := fileLikeModel{
		Version: fileLikeModelVersion,
	}

	err := storage.Load(ctx, &fstate)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read file")
	}

	return &fileLikeBackend{model: &fstate, storage: storage}, nil
}

func (flb *fileLikeBackend) GetState(ctx context.Context, layerName, stateName string) (*State, error) {
	hclog.FromContext(ctx).Debug("Getting layer state", "layer", layerName, "state", stateName)

	for _, state := range flb.model.States {
		if state.LayerName == layerName && state.StateName == stateName {
			return state, nil
		}
	}

	return nil, errors.Wrapf(ErrStateNotFound, "state %s for layer %s not found", stateName, layerName)
}

func (flb *fileLikeBackend) SaveState(ctx context.Context, state *State) error {
	hclog.FromContext(ctx).Debug("Saving layer state", "layer", state.LayerName, "state", state.StateName)

	nextStates := []*State{}
	for _, s := range flb.model.States {
		if s.LayerName != state.LayerName || s.StateName != state.StateName {
			nextStates = append(nextStates, s)
		}
	}

	nextStates = append(nextStates, state)

	flb.model.States = nextStates

	return flb.storage.Save(ctx, flb.model)
}

func (flb *fileLikeBackend) DeleteState(ctx context.Context, layerName, stateName string) error {
	hclog.FromContext(ctx).Debug("Deleting layer state", "layer", layerName, "state", stateName)

	nextStates := []*State{}
	for _, s := range flb.model.States {
		if s.LayerName != layerName || s.StateName != stateName {
			nextStates = append(nextStates, s)
		}
	}

	flb.model.States = nextStates

	return flb.storage.Save(ctx, flb.model)
}

func (flb *fileLikeBackend) ListStatesByLayer(ctx context.Context, layerName string) ([]*State, error) {
	hclog.FromContext(ctx).Debug("Listing states by layer", "layer", layerName)

	result := make([]*State, 0)
	for _, s := range flb.model.States {
		if s.LayerName == layerName {
			result = append(result, s)
		}
	}

	return result, nil
}

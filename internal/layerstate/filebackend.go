package layerstate

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

const fileStateVersion = 0

type filestate struct {
	Version uint     `json:"version"`
	States  []*State `json:"states"`
}

type filebackend struct {
	fpath string
}

var _ Backend = &filebackend{}

func NewFileBackend(fpath string) *filebackend {
	return &filebackend{fpath}
}

func (fb *filebackend) readFile(ctx context.Context) (*filestate, error) {
	hclog.FromContext(ctx).Debug("Reading state file", "path", fb.fpath)

	raw, err := os.ReadFile(fb.fpath)
	if errors.Is(err, os.ErrNotExist) {
		return &filestate{Version: fileStateVersion}, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "fail to read %s", fb.fpath)
	}

	var fstate filestate
	err = json.Unmarshal(raw, &fstate)

	return &fstate, errors.Wrapf(err, "fail to parse state out of %s", fb.fpath)
}

func (fb *filebackend) writeFile(ctx context.Context, fstate *filestate) error {
	hclog.FromContext(ctx).Debug("Writting state to file", "path", fb.fpath)

	data, err := json.Marshal(fstate)
	if err != nil {
		return errors.Wrap(err, "fail to marshal file state")
	}

	err = os.WriteFile(fb.fpath, data, 0644)
	return errors.Wrap(err, "fail to write file")
}

func (fb *filebackend) GetState(ctx context.Context, layerName, stateName string) (*State, error) {
	hclog.FromContext(ctx).Debug("Getting layer state", "layer", layerName, "state", stateName)

	fstate, err := fb.readFile(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to read file")
	}

	for _, state := range fstate.States {
		if state.LayerName == layerName && state.StateName == stateName {
			return state, nil
		}
	}

	return nil, errors.Wrapf(ErrStateNotFound, "state %s for layer %s not found", stateName, layerName)
}

func (fb *filebackend) SaveState(ctx context.Context, state *State) error {
	hclog.FromContext(ctx).Debug("Saving layer state", "layer", state.LayerName, "state", state.StateName)

	fstate, err := fb.readFile(ctx)
	if err != nil {
		return errors.Wrapf(err, "fail to read file")
	}

	nextStates := []*State{}
	for _, s := range fstate.States {
		if s.LayerName != state.LayerName || s.StateName != state.StateName {
			nextStates = append(nextStates, s)
		}
	}

	nextStates = append(nextStates, state)

	fstate.States = nextStates

	return fb.writeFile(ctx, fstate)
}

func (fb *filebackend) DeleteState(ctx context.Context, layerName, stateName string) error {
	hclog.FromContext(ctx).Debug("Deleting layer state", "layer", layerName, "state", stateName)

	fstate, err := fb.readFile(ctx)
	if err != nil {
		return errors.Wrapf(err, "fail to read file")
	}

	nextStates := []*State{}
	for _, s := range fstate.States {
		if s.LayerName != layerName || s.StateName != stateName {
			nextStates = append(nextStates, s)
		}
	}

	fstate.States = nextStates

	return fb.writeFile(ctx, fstate)
}

func (fb *filebackend) ListStatesByLayer(ctx context.Context, layerName string) ([]*State, error) {
	hclog.FromContext(ctx).Debug("Listing states by layer", "layer", layerName)

	fstate, err := fb.readFile(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to read file")
	}

	result := make([]*State, 0)
	for _, s := range fstate.States {
		if s.LayerName == layerName {
			result = append(result, s)
		}
	}

	return result, nil
}

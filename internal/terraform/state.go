package terraform

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type State struct {
	Resources []StateResource `json:"resources"`
	Bytes     []byte          `json:"bytes"`
}

// ResourceDiff returns the resources in b which are not present in a
func (state *State) ResourceDiff(b *State) []StateResource {
	resourceMap := make(map[StateResource]struct{})
	for _, aRes := range state.Resources {
		resourceMap[aRes] = struct{}{}
	}

	diff := make([]StateResource, 0)
	for _, bRes := range b.Resources {
		if _, found := resourceMap[bRes]; !found {
			diff = append(diff, bRes)
		}
	}

	return diff
}

type StateResource struct {
	Module string `json:"module,omitempty"`
	Type   string `json:"type"`
	Name   string `json:"name"`
}

func (res *StateResource) Equal(b *StateResource) bool {
	return res.Module == b.Module && res.Type == b.Type && res.Name == b.Name
}

func (res *StateResource) Address() string {
	return strings.Join([]string{res.Module, res.Type, res.Name}, ".")
}

func TFStateFromBytes(data []byte) (*State, error) {
	state := &State{Bytes: data}
	err := json.Unmarshal(data, state)

	return state, errors.Wrap(err, "fail to unmarshal data into state")
}

func TFStateFromFile(filePath string) (*State, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to read file %s", filePath)
	}

	return TFStateFromBytes(data)
}

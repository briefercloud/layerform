package layerdefinitions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
)

type ergomake struct {
	baseURL string
}

var _ Backend = &ergomake{}

func NewErgomake(baseURL string) *ergomake {
	return &ergomake{baseURL}
}

func (e *ergomake) Location(ctx context.Context) (string, error) {
	return e.baseURL, nil
}

func (e *ergomake) GetLayer(ctx context.Context, name string) (*data.LayerDefinition, error) {
	url := fmt.Sprintf("%s/v1/definitions/%s", e.baseURL, name)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	var layer data.LayerDefinition
	err = json.NewDecoder(resp.Body).Decode(&layer)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode layer JSON response")
	}

	return &layer, nil
}

func (e *ergomake) ListLayers(ctx context.Context) ([]*data.LayerDefinition, error) {
	url := fmt.Sprintf("%s/v1/definitions", e.baseURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	var layers []*data.LayerDefinition
	err = json.NewDecoder(resp.Body).Decode(&layers)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode layers JSON response")
	}

	return layers, nil
}

func (e *ergomake) ResolveDependencies(ctx context.Context, layer *data.LayerDefinition) ([]*data.LayerDefinition, error) {
	var resolvedLayers []*data.LayerDefinition

	for _, dependencyName := range layer.Dependencies {
		dependencyLayer, err := e.GetLayer(ctx, dependencyName)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to resolve dependency %s", dependencyName)
		}

		resolvedLayers = append(resolvedLayers, dependencyLayer)
	}

	return resolvedLayers, nil
}

func (e *ergomake) UpdateLayers(ctx context.Context, layers []*data.LayerDefinition) error {
	dataBytes, err := json.Marshal(layers)
	if err != nil {
		return errors.Wrap(err, "fail to marshal layers to json")
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/v1/configure", e.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		return errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	return nil
}

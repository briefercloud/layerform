package layerdefinitions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/cloud"
	"github.com/ergomake/layerform/pkg/data"
)

type cloudBackend struct {
	client *cloud.HTTPClient
}

var _ Backend = &cloudBackend{}

func NewCloud(client *cloud.HTTPClient) *cloudBackend {
	return &cloudBackend{client}
}

func (e *cloudBackend) Location(ctx context.Context) (string, error) {
	return e.client.BaseURL, nil
}

func (e *cloudBackend) GetLayer(ctx context.Context, name string) (*data.LayerDefinition, error) {
	url := fmt.Sprintf("/v1/definitions/%s", name)

	req, err := e.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request to cloud backend")
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fail to perform http request to cloud backend")
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

func (e *cloudBackend) ListLayers(ctx context.Context) ([]*data.LayerDefinition, error) {
	url := "/v1/definitions"

	req, err := e.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request to cloud backend")
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fail to perform http request to cloud backend")
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

func (e *cloudBackend) ResolveDependencies(ctx context.Context, layer *data.LayerDefinition) ([]*data.LayerDefinition, error) {
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

func (e *cloudBackend) UpdateLayers(ctx context.Context, layers []*data.LayerDefinition) error {
	dataBytes, err := json.Marshal(layers)
	if err != nil {
		return errors.Wrap(err, "fail to marshal layers to json")
	}

	url := "/v1/configure"

	req, err := e.client.NewRequest(ctx, "POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		return errors.Wrap(err, "fail to create http request to cloud backend")
	}

	req.SetHeader("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "fail to perform http request to cloud backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	return nil
}

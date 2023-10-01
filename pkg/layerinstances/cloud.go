package layerinstances

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

func (e *cloudBackend) DeleteInstance(ctx context.Context, layerName string, instanceName string) error {
	url := fmt.Sprintf("/v1/definitions/%s/instances/%s", layerName, instanceName)
	req, err := e.client.NewRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return errors.Wrap(err, "fail to create http request to cloud backend")
	}

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

func (e *cloudBackend) GetInstance(ctx context.Context, definitionName string, instanceName string) (*data.LayerInstance, error) {
	url := fmt.Sprintf("/v1/definitions/%s/instances/%s", definitionName, instanceName)
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
		if resp.StatusCode == http.StatusNotFound {
			err := errors.Wrapf(ErrInstanceNotFound, "fail to get instance %s of definition %s", instanceName, definitionName)
			return nil, err
		}

		return nil, errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	var instance data.LayerInstance
	err = json.NewDecoder(resp.Body).Decode(&instance)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode instance JSON response")
	}

	return &instance, nil
}

func (e *cloudBackend) ListInstances(ctx context.Context) ([]*data.LayerInstance, error) {
	url := "/v1/instances"
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

	var instances []*data.LayerInstance
	err = json.NewDecoder(resp.Body).Decode(&instances)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode instances JSON response")
	}

	return instances, nil
}

func (e *cloudBackend) ListInstancesByLayer(ctx context.Context, layerName string) ([]*data.LayerInstance, error) {
	url := fmt.Sprintf("/v1/definitions/%s/instances", layerName)
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

	var instances []*data.LayerInstance
	err = json.NewDecoder(resp.Body).Decode(&instances)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode instances JSON response")
	}

	return instances, nil
}

func (e *cloudBackend) SaveInstance(ctx context.Context, instance *data.LayerInstance) error {
	url := "/v1/instances"
	dataBytes, err := json.Marshal(instance)
	if err != nil {
		return errors.Wrap(err, "fail to marshal instance to json")
	}

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

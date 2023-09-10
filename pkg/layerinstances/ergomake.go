package layerinstances

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

func (e *ergomake) DeleteInstance(ctx context.Context, layerName string, instanceName string) error {
	url := fmt.Sprintf("%s/v1/definitions/%s/instances/%s", e.baseURL, layerName, instanceName)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	return nil
}

func (e *ergomake) GetInstance(ctx context.Context, definitionName string, instanceName string) (*data.LayerInstance, error) {
	url := fmt.Sprintf("%s/v1/definitions/%s/instances/%s", e.baseURL, definitionName, instanceName)
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

func (e *ergomake) ListInstances(ctx context.Context) ([]*data.LayerInstance, error) {
	url := fmt.Sprintf("%s/v1/instances", e.baseURL)
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

	var instances []*data.LayerInstance
	err = json.NewDecoder(resp.Body).Decode(&instances)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode instances JSON response")
	}

	return instances, nil
}

func (e *ergomake) ListInstancesByLayer(ctx context.Context, layerName string) ([]*data.LayerInstance, error) {
	url := fmt.Sprintf("%s/v1/definitions/%s/instances", e.baseURL, layerName)
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

	var instances []*data.LayerInstance
	err = json.NewDecoder(resp.Body).Decode(&instances)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode instances JSON response")
	}

	return instances, nil
}

func (e *ergomake) SaveInstance(ctx context.Context, instance *data.LayerInstance) error {
	url := fmt.Sprintf("%s/v1/instances", e.baseURL)
	dataBytes, err := json.Marshal(instance)
	if err != nil {
		return errors.Wrap(err, "fail to marshal instance to json")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		return errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	return nil
}

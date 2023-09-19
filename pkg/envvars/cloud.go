package envvars

import (
	"bytes"
	"context"
	"encoding/json"
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

func (cb *cloudBackend) ListVariables(ctx context.Context) ([]*data.EnvVar, error) {
	url := "/v1/env-vars"
	req, err := cb.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request to cloud backend")
	}

	req.SetHeader("Content-Type", "application/json")
	resp, err := cb.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fail to perform http request to cloud backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	var variables []*data.EnvVar
	err = json.NewDecoder(resp.Body).Decode(&variables)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode variables JSON response")
	}

	return variables, nil
}

func (cb *cloudBackend) SaveVariable(ctx context.Context, variable *data.EnvVar) error {
	url := "/v1/env-vars"
	dataBytes, err := json.Marshal(variable)
	if err != nil {
		return errors.Wrap(err, "fail to marshal env var to json")
	}

	req, err := cb.client.NewRequest(ctx, "POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		return errors.Wrap(err, "fail to create http request to cloud backend")
	}

	req.SetHeader("Content-Type", "application/json")
	resp, err := cb.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "fail to perform http request to cloud backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	return nil
}

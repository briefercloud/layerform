package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type httpRequest struct {
	req *http.Request
}

func (r *httpRequest) SetHeader(key, value string) {
	r.req.Header.Set(key, value)
}

var ErrInvalidCreds = errors.New("invalid credentials")

type HTTPClient struct {
	token   string
	BaseURL string
}

func NewHTTPClient(ctx context.Context, baseURL, email, password string) (*HTTPClient, error) {
	url := fmt.Sprintf("%s/v1/auth/signin", baseURL)
	data, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return nil, errors.Wrap(err, "fail to encode auth payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request to cloud backend")
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fail to perform http request to cloud backend")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
			return nil, errors.Wrapf(ErrInvalidCreds, "status code %d", res.StatusCode)
		}

		return nil, errors.Errorf("HTTP request to %s failed with status code %d", url, res.StatusCode)
	}

	var body struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode auth JSON response")
	}

	return &HTTPClient{
		token:   body.Token,
		BaseURL: baseURL,
	}, nil
}

func (c *HTTPClient) NewRequest(ctx context.Context, method, path string, body io.Reader) (*httpRequest, error) {
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.BaseURL, path), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	return &httpRequest{req}, nil
}

func (c *HTTPClient) Do(req *httpRequest) (*http.Response, error) {
	return http.DefaultClient.Do(req.req)
}

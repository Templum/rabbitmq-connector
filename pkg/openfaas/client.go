package openfaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	internal "github.com/Templum/rabbitmq-connector/pkg/types"
	external "github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
)

// Client is used for interacting with Open FaaS
type Client struct {
	Client      *http.Client
	credentials *internal.Credentials
	url         string
}

// InvokeSync TODO:
func (c *Client) InvokeSync(ctx context.Context, name string, payload []byte) ([]byte, error) { // TODO: either reuse provided payload or make it pasable
	functionURL := fmt.Sprintf("%s/function/%s", c.url, name)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, functionURL, bytes.NewReader(payload))
	if c.credentials != nil {
		req.SetBasicAuth(c.credentials.User, c.credentials.Password)
	}

	if req.Body != nil {
		defer req.Body.Close()
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to invoke function %s", name)
	}

	var output []byte

	if res.Body != nil {
		defer res.Body.Close()
		output, _ = ioutil.ReadAll(res.Body)
	}

	switch res.StatusCode {
	case 200:
		return output, nil
	case 401:
		return nil, errors.New("OpenFaaS Credentials are invalid")
	case 404:
		return nil, errors.New(fmt.Sprintf("Function %s is not deployed", name))
	default:
		return nil, errors.New(fmt.Sprintf("Received unexpected Status Code %d", res.StatusCode))
	}
}

// InvokeAsync TODO:
func (c *Client) InvokeAsync(ctx context.Context, name string, payload []byte) (bool, error) { // TODO: either reuse provided payload or make it pasable
	functionURL := fmt.Sprintf("%s/async-function/%s", c.url, name)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, functionURL, bytes.NewReader(payload))
	if c.credentials != nil {
		req.SetBasicAuth(c.credentials.User, c.credentials.Password)
	}

	if req.Body != nil {
		defer req.Body.Close()
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return false, errors.Wrapf(err, "unable to invoke function %s", name)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case 202:
		return true, nil
	case 401:
		return false, errors.New("OpenFaaS Credentials are invalid")
	case 404:
		return false, errors.New(fmt.Sprintf("Function %s is not deployed", name))
	default:
		return false, errors.New(fmt.Sprintf("Received unexpected Status Code %d", res.StatusCode))
	}
}

// HasNamespaceSupport TODO:
func (c *Client) HasNamespaceSupport(ctx context.Context) (bool, error) {
	getNamespaces := fmt.Sprintf("%s/system/namespaces", c.url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, getNamespaces, nil)
	if c.credentials != nil {
		req.SetBasicAuth(c.credentials.User, c.credentials.Password)
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return false, errors.Wrapf(err, "unable to determine namespace support")
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case 200:
		return true, nil
	case 401:
		return false, errors.New("OpenFaaS Credentials are invalid")
	default:
		return false, nil
	}
}

// GetFunctions TODO:
func (c *Client) GetFunctions(ctx context.Context, namespace string) ([]external.FunctionStatus, error) {
	getFunctions := fmt.Sprintf("%s/system/functions", c.url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, getFunctions, nil)
	if c.credentials != nil {
		req.SetBasicAuth(c.credentials.User, c.credentials.Password)
	}

	if len(namespace) > 0 {
		q := req.URL.Query()
		q.Add("namespace", namespace)
		req.URL.RawQuery = q.Encode()
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "unable to obtain functions")
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resp, _ := ioutil.ReadAll(res.Body)
	functions := []external.FunctionStatus{}
	err = json.Unmarshal(resp, &functions)

	if err != nil {
		if res.StatusCode == 401 {
			return nil, errors.New("OpenFaaS Credentials are invalid")
		}
		return nil, errors.Wrapf(err, "Received during obtaining function along with status code %d", res.StatusCode)
	}

	return functions, nil
}

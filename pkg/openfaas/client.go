package openfaas

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	types "github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/pkg/errors"
)

// Client is used for interacting with Open FaaS
type Client struct {
	Client      *http.Client
	credentials *types.Credentials
	url         string
}

// Invoke TODO:
func (c *Client) Invoke(ctx context.Context, name string, payload []byte) ([]byte, error) { // TODO: either reuse provided payload or make it pasable
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

	if res.Body != nil {
		defer res.Body.Close()
		output, _ := ioutil.ReadAll(res.Body)

		switch res.StatusCode {
		case 200:
			return output, nil
		case 401:
			return nil, errors.New("OpenFaaS Credentials are invalid")
		case 404:
			return nil, errors.New(fmt.Sprintf("Function %s is not deployed", name))
		default:
			return nil, errors.New(fmt.Sprintf("Recieved unexpected Status Code %d", res.StatusCode))
		}
	}

	return nil, nil
}

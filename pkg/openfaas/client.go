package openfaas

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
)

// Invoker defines interfaces that invoke deployed OpenFaaS Functions.
type Invoker interface {
	InvokeSync(ctx context.Context, name string, payload []byte) ([]byte, error)
	InvokeAsync(ctx context.Context, name string, payload []byte) (bool, error)
}

// NamespaceFetcher defines interfaces to explore namespaces of an OpenFaaS installation.
type NamespaceFetcher interface {
	HasNamespaceSupport(ctx context.Context) (bool, error)
	GetNamespaces(ctx context.Context) ([]string, error)
}

// FunctionFetcher defines interface to explore deployed function of an OpenFaaS installation.
type FunctionFetcher interface {
	GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error)
}

// FunctionCrawler defines interfaces required to crawl OpenFaaS for functions
type FunctionCrawler interface {
	NamespaceFetcher
	FunctionFetcher
}

// Client is used for interacting with Open FaaS
type Client struct {
	client      *http.Client
	credentials *auth.BasicAuthCredentials
	url         string
}

// NewClient creates a new instance of an OpenFaaS Client using
// the provided informations
func NewClient(client *http.Client, creds *auth.BasicAuthCredentials, gatewayURL string) *Client {
	return &Client{
		client:      client,
		credentials: creds,
		url:         gatewayURL,
	}
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

	res, err := c.client.Do(req)
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

	res, err := c.client.Do(req)
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

	res, err := c.client.Do(req)
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

// GetNamespaces TODO:
func (c *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	getNamespaces := fmt.Sprintf("%s/system/namespaces", c.url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, getNamespaces, nil)
	if c.credentials != nil {
		req.SetBasicAuth(c.credentials.User, c.credentials.Password)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "unable to obtain namespaces")
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resp, _ := ioutil.ReadAll(res.Body)
	var namespaces []string
	err = json.Unmarshal(resp, &namespaces)

	if err != nil {
		if res.StatusCode == 401 {
			return nil, errors.New("OpenFaaS Credentials are invalid")
		}
		return namespaces, nil
	}

	return namespaces, nil
}

// GetFunctions TODO:
func (c *Client) GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {
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

	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "unable to obtain functions")
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resp, _ := ioutil.ReadAll(res.Body)
	functions := []types.FunctionStatus{}
	err = json.Unmarshal(resp, &functions)

	if err != nil {
		if res.StatusCode == 401 {
			return nil, errors.New("OpenFaaS Credentials are invalid")
		}
		return nil, errors.New(fmt.Sprintf("Received unexpected Status Code %d", res.StatusCode))
	}

	return functions, nil
}

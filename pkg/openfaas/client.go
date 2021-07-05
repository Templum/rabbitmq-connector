/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package openfaas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	internal "github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/valyala/fasthttp"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
)

// Invoker defines interfaces that invoke deployed OpenFaaS Functions.
type Invoker interface {
	InvokeSync(ctx context.Context, name string, invocation *internal.OpenFaaSInvocation) ([]byte, error)
	InvokeAsync(ctx context.Context, name string, invocation *internal.OpenFaaSInvocation) (bool, error)
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
	Invoker
}

// Client is used for interacting with Open FaaS
type Client struct {
	client      *fasthttp.Client
	credentials *auth.BasicAuthCredentials
	url         string
}

// NewClient creates a new instance of an OpenFaaS Client using
// the provided information
func NewClient(client *fasthttp.Client, creds *auth.BasicAuthCredentials, gatewayURL string) *Client {
	return &Client{
		client:      client,
		credentials: creds,
		url:         gatewayURL,
	}
}

// InvokeSync calls a given function in a synchronous way waiting for the response using the provided payload while considering the provided context
func (c *Client) InvokeSync(ctx context.Context, name string, invocation *internal.OpenFaaSInvocation) ([]byte, error) {
	functionURL := fmt.Sprintf("%s/function/%s", c.url, name)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(functionURL)
	if invocation.Message != nil {
		req.SetBody(*invocation.Message)
	} else {
		req.SetBody(nil)
	}

	req.Header.Set("Content-Type", invocation.ContentType)
	req.Header.Set("Content-Encoding", invocation.ContentEncoding)
	req.Header.SetUserAgent("OpenFaaS - Rabbit MQ Connector")
	if c.credentials != nil {
		credentials := c.credentials.User + ":" + c.credentials.Password
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	}

	err := c.client.Do(req, resp)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to invoke function %s", name)
	}

	switch resp.StatusCode() {
	case fasthttp.StatusOK:
		return resp.Body(), nil
	case fasthttp.StatusUnauthorized:
		return nil, errors.New("OpenFaaS Credentials are invalid")
	case fasthttp.StatusNotFound:
		return nil, errors.New(fmt.Sprintf("Function %s is not deployed", name))
	default:
		return nil, errors.New(fmt.Sprintf("Received unexpected Status Code %d", resp.StatusCode()))
	}
}

// InvokeAsync calls a given function in a asynchronous way waiting for the response using the provided payload while considering the provided context
func (c *Client) InvokeAsync(ctx context.Context, name string, invocation *internal.OpenFaaSInvocation) (bool, error) {
	functionURL := fmt.Sprintf("%s/async-function/%s", c.url, name)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(functionURL)
	if invocation.Message != nil {
		req.SetBody(*invocation.Message)
	} else {
		req.SetBody(nil)
	}

	req.Header.Set("Content-Type", invocation.ContentType)
	req.Header.Set("Content-Encoding", invocation.ContentEncoding)
	req.Header.SetUserAgent("OpenFaaS - Rabbit MQ Connector")
	if c.credentials != nil {
		credentials := c.credentials.User + ":" + c.credentials.Password
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	}

	err := c.client.Do(req, resp)
	if err != nil {
		return false, errors.Wrapf(err, "unable to invoke function %s", name)
	}

	switch resp.StatusCode() {
	case fasthttp.StatusAccepted:
		return true, nil
	case fasthttp.StatusUnauthorized:
		return false, errors.New("OpenFaaS Credentials are invalid")
	case fasthttp.StatusNotFound:
		return false, errors.New(fmt.Sprintf("Function %s is not deployed", name))
	default:
		return false, errors.New(fmt.Sprintf("Received unexpected Status Code %d", resp.StatusCode()))
	}
}

// HasNamespaceSupport Checks if the version of OpenFaaS does support Namespace
func (c *Client) HasNamespaceSupport(ctx context.Context) (bool, error) {
	getNamespaces := fmt.Sprintf("%s/system/namespaces", c.url)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(getNamespaces)

	req.Header.SetUserAgent("OpenFaaS - Rabbit MQ Connector")
	if c.credentials != nil {
		credentials := c.credentials.User + ":" + c.credentials.Password
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	}

	err := c.client.Do(req, resp)
	if err != nil {
		return false, errors.Wrapf(err, "unable to determine namespace support")
	}

	switch resp.StatusCode() {
	case fasthttp.StatusOK:
		var namespaces []string
		_ = json.Unmarshal(resp.Body(), &namespaces)
		// Swarm edition of OF does not support namespaces and is simply returning empty array
		return len(namespaces) > 0, nil
	case fasthttp.StatusUnauthorized:
		return false, errors.New("OpenFaaS Credentials are invalid")
	default:
		log.Println(fmt.Sprintf("Received unexpected Status Code %d while fetching namespaces", resp.StatusCode()))
		return false, nil
	}
}

// GetNamespaces returns all namespaces where Functions are deployed on
func (c *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	getNamespaces := fmt.Sprintf("%s/system/namespaces", c.url)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(getNamespaces)

	req.Header.SetUserAgent("OpenFaaS - Rabbit MQ Connector")
	if c.credentials != nil {
		credentials := c.credentials.User + ":" + c.credentials.Password
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	}

	err := c.client.Do(req, resp)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch namespaces")
	}

	switch resp.StatusCode() {
	case fasthttp.StatusOK:
		var namespaces []string
		_ = json.Unmarshal(resp.Body(), &namespaces)
		// Swarm edition of OF does not support namespaces and is simply returning empty array
		return namespaces, nil
	case fasthttp.StatusUnauthorized:
		return nil, errors.New("OpenFaaS Credentials are invalid")
	default:
		log.Println(fmt.Sprintf("Received unexpected Status Code %d while fetching namespaces", resp.StatusCode()))
		return nil, nil
	}
}

// GetFunctions returns a list of all functions in the given namespace or in the default namespace
func (c *Client) GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {
	getFunctions := fmt.Sprintf("%s/system/functions", c.url)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(getFunctions)

	req.Header.SetUserAgent("OpenFaaS - Rabbit MQ Connector")
	if c.credentials != nil {
		credentials := c.credentials.User + ":" + c.credentials.Password
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	}

	if len(namespace) > 0 {
		req.URI().QueryArgs().Add("namespace", namespace)
	}

	err := c.client.Do(req, resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to obtain functions")
	}

	switch resp.StatusCode() {
	case fasthttp.StatusOK:
		var functions []types.FunctionStatus
		_ = json.Unmarshal(resp.Body(), &functions)
		// Swarm edition of OF does not support namespaces and is simply returning empty array
		return functions, nil
	case fasthttp.StatusUnauthorized:
		return nil, errors.New("OpenFaaS Credentials are invalid")
	default:
		return nil, errors.New(fmt.Sprintf("Received unexpected Status Code %d", resp.StatusCode()))
	}
}

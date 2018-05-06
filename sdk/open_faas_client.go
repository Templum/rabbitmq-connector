// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas/gateway/requests"
)

// OpenFaaSClient is an abstraction which handles
// the interaction with OpenFaaS gateway.
type OpenFaaSClient struct {
	url        string
	httpClient *http.Client
}

func NewOpenFaaSClient(url string, client *http.Client) OpenFaaSClient {
	return OpenFaaSClient{
		url:        url,
		httpClient: client,
	}
}

// buildUrl returns either the fetch path or the invoke path
// based on the provided input
func buildUrl(baseUrl string, function string) string {
	if len(function) == 0 {
		return fmt.Sprintf("%s/system/functions", baseUrl)
	} else {
		return fmt.Sprintf("%s/function/%s", baseUrl, function)
	}
}

// FetchFunctions queries the OpenFaaS gateway for all available functions
// and returns either the list of the functions or an error.
func (client *OpenFaaSClient) FetchFunctions() (*[]requests.Function, error) {
	req, _ := http.NewRequest(http.MethodGet, buildUrl(client.url, ""), nil)
	req.Close = true

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var functions []requests.Function
	if err := json.Unmarshal(bytesOut, &functions); err != nil {
		return nil, err
	} else {
		return &functions, nil
	}
}

// InvokeFunction invokes the provided function and returns either the result
// of the invocation or an error.
func (client *OpenFaaSClient) InvokeFunction(function string, message []byte) (*[]byte, error) {
	req, _ := http.NewRequest(http.MethodPost, buildUrl(client.url, function), bytes.NewReader(message))
	req.Close = true
	defer req.Body.Close()

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	} else {
		return &bytesOut, nil
	}

}

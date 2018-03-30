// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package openfaas

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

type Invoker struct {
	PrintResponse bool
	Client        *http.Client
	GatewayURL    string
}

func (i *Invoker) Invoke(lookupTable *TopicLookupTable, topic string, message *[]byte) {
	if len(*message) > 0 {

		matchedFunctions := lookupTable.Match(topic)
		for _, matchedFunction := range matchedFunctions {

			log.Printf("Invoke function: %s", matchedFunction)

			gatewayURL := fmt.Sprintf("%s/function/%s", i.GatewayURL, matchedFunction)
			reader := bytes.NewReader(*message)

			body, statusCode, err := invokeFunction(i.Client, gatewayURL, reader)

			if err != nil {
				log.Printf("Unable to invoke from %s, error: %s\n", matchedFunction, err)
				return
			}

			if body != nil && i.PrintResponse {
				log.Printf("Response [%d] from %s %s", statusCode, matchedFunction, string(*body))

			} else {
				log.Printf("Response [%d] from %s", statusCode, matchedFunction)
			}
		}
	}
}

func invokeFunction(c *http.Client, gatewayURL string, reader io.Reader) (*[]byte, int, error) {

	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL, reader)
	defer httpReq.Body.Close()

	var body *[]byte
	res, err := c.Do(httpReq)

	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error reading body")
			return nil, http.StatusServiceUnavailable, err

		}
		body = &bytesOut
	}

	return body, res.StatusCode, err
}

func MakeClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     120 * time.Millisecond,
		},
	}
}

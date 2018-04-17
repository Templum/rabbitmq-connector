// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas/gateway/requests"
)

// FunctionLookupBuilder builds a list of OpenFaaS functions
type FunctionLookupBuilder struct {
	GatewayURL string
	Client     *http.Client
}

// Build compiles a map of topic names and functions that have
// advertised to receive messages on said topic
func (s *FunctionLookupBuilder) Build() (map[string][]string, error) {
	var err error
	serviceMap := make(map[string][]string)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/system/functions", s.GatewayURL), nil)
	res, reqErr := s.Client.Do(req)

	if reqErr != nil {
		return serviceMap, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, _ := ioutil.ReadAll(res.Body)

	functions := []requests.Function{}
	marshalErr := json.Unmarshal(bytesOut, &functions)

	if marshalErr != nil {
		return serviceMap, marshalErr
	}

	for _, function := range functions {
		if *function.Labels != nil {
			labels := *function.Labels

			if topic, pass := labels["topic"]; pass {

				if serviceMap[topic] == nil {
					serviceMap[topic] = []string{}
				}
				serviceMap[topic] = append(serviceMap[topic], function.Name)
			}
		}
	}

	return serviceMap, err
}

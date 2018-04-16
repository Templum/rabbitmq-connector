// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package openfaas

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

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/system/functions", s.GatewayURL),  nil)
	req.Close = true
	res, err := s.Client.Do(req)

	if err != nil {
		return serviceMap, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return serviceMap, err
	}

	functions := []requests.Function{}
	err = json.Unmarshal(bytesOut, &functions)

	if err != nil {
		return serviceMap, err
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

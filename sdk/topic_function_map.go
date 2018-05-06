// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package sdk

import (
	"github.com/openfaas/faas/gateway/requests"
	"sync"
	"strings"
)

// TopicFunctionMap is an structure which contains an mapping between
// functions an topics.
type TopicFunctionMap struct {
	lookupTable *map[string][]string
	mutex       sync.Mutex
}

func NewTopicFunctionMap() TopicFunctionMap {
	mapping := make(map[string][]string)

	return TopicFunctionMap{
		lookupTable: &mapping,
		mutex:       sync.Mutex{},
	}
}

// Match returns all function names which are listening
// on the provided topic
func (t *TopicFunctionMap) Match(topic string) []string {
	var values []string

	t.mutex.Lock()

	for key, val := range *t.lookupTable {
		if key == topic {
			values = val
			break
		}
	}

	t.mutex.Unlock()

	return values
}

// Sync takes a list of functions and uses it for syncing the lookupTable
func (t *TopicFunctionMap) Sync(functions *[]requests.Function) {
	mapping := make(map[string][]string)

	for _, function := range *functions {
		if *function.Labels != nil {
			labels := *function.Labels
			if topics, pass := labels["topic"]; pass {
				for _, topic := range strings.Split(topics, ","){
					if len(topic) > 0{
						if mapping[topic] == nil {
							mapping[topic] = []string{}
						}
						mapping[topic] = append(mapping[topic], function.Name)
					}
				}
			}
		}
	}

	t.mutex.Lock()

	t.lookupTable = &mapping

	t.mutex.Unlock()
}

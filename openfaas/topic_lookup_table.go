// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package openfaas

import "sync"

func CreateEmptyLookupMap() TopicLookupTable {
	lookup := make(map[string][]string)
	return TopicLookupTable{
		lookup: &lookup,
		lock:   sync.Mutex{},
	}
}

type TopicLookupTable struct {
	lookup *map[string][]string
	lock   sync.Mutex
}

func (t *TopicLookupTable) Match(topicName string) []string {
	t.lock.Lock()

	var values []string

	for key, val := range *t.lookup {
		if key == topicName {
			values = val
			break
		}
	}

	t.lock.Unlock()

	return values
}

func (t *TopicLookupTable) Sync(updated *map[string][]string) {
	t.lock.Lock()

	t.lookup = updated

	t.lock.Unlock()
}

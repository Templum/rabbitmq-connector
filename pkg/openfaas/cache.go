package openfaas

import "sync"

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// TopicFunctionCache TODO:
type TopicFunctionCache struct {
	topicMap *map[string][]string
	lock     sync.RWMutex
}

// Includes TODO:
func (m *TopicFunctionCache) Includes(name string) []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var functions []string
	for topic, function := range *m.topicMap {
		if topic == name {
			functions = function
			break
		}
	}

	return functions
}

// Refresh TODO:
func (m *TopicFunctionCache) Refresh(update *map[string][]string) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	m.topicMap = update
}

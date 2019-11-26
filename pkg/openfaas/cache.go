package openfaas

import (
	"log"
	"sync"
)

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// TopicMap defines a interface for a topic map
type TopicMap interface {
	GetCachedValues(name string) []string
	Refresh(update map[string][]string)
}

// TopicFunctionCache TODO:
type TopicFunctionCache struct {
	topicMap map[string][]string
	lock     sync.RWMutex
}

// NewTopicFunctionCache return a new instance
func NewTopicFunctionCache() *TopicFunctionCache {
	return &TopicFunctionCache{
		topicMap: make(map[string][]string),
		lock:     sync.RWMutex{},
	}
}

// GetCachedValues TODO:
func (m *TopicFunctionCache) GetCachedValues(name string) []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var functions []string
	for topic, function := range m.topicMap {
		if topic == name {
			functions = function
			break
		}
	}

	return functions
}

// Refresh TODO:
func (m *TopicFunctionCache) Refresh(update map[string][]string) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	log.Printf("Update cache with %d entries", len(update))
	m.topicMap = update
}

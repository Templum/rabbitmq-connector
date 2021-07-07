/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package openfaas

import (
	"log"
	"sync"
)

// TopicMap defines a interface for a topic map
type TopicMap interface {
	GetCachedValues(name string) []string
	Refresh(update map[string][]string)
}

// TopicFunctionCache contains a map of of topics to functions
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

// GetCachedValues reads the cached functions for a given topic
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

// Refresh updates the existing cache with new values while syncing ensuring no read conflicts
func (m *TopicFunctionCache) Refresh(update map[string][]string) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	log.Printf("Update cache with %d entries", len(update))
	m.topicMap = update
}

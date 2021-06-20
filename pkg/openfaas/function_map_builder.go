/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package openfaas

import "strings"

// TopicMapBuilder defines an interface that allows to build a TopicMap
type TopicMapBuilder interface {
	Append(topic string, function string)
	Build() map[string][]string
}

// FunctionMapBuilder convenient construct to build a map
// of function <=> topic
type FunctionMapBuilder struct {
	target map[string][]string
}

// NewFunctionMapBuilder returns a new instance with an empty build target
func NewFunctionMapBuilder() *FunctionMapBuilder {
	return &FunctionMapBuilder{
		target: make(map[string][]string),
	}
}

// Append the provided function to the specified topic
func (b *FunctionMapBuilder) Append(topic string, function string) {
	key := strings.TrimSpace(topic)

	if len(key) == 0 {
		println("Topic was empty after trimming will ignore provided functions")
		return
	}

	if b.target[key] == nil {
		b.target[key] = []string{}
	}

	b.target[key] = append(b.target[key], function)
}

// Build returns a map containing values based on previous Append calls
func (b *FunctionMapBuilder) Build() map[string][]string {
	return b.target
}

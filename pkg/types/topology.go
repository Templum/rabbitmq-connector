/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package types

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type Topology []struct {
	Name        string   `json:"name"`
	Topics      []string `json:"topics"`
	Declare     bool     `json:"declare"`
	Type        string   `json:"type,omitempty"`
	Durable     bool     `json:"durable,omitempty"`
	AutoDeleted bool     `json:"auto-deleted,omitempty"`
}

type Exchange struct {
	Name        string
	Topics      []string
	Declare     bool
	Type        string
	Durable     bool
	AutoDeleted bool
}

func (e *Exchange) EnsureCorrectType() {
	switch strings.ToLower(e.Type) {
	case "direct":
		e.Type = "direct"
	case "topic":
		e.Type = "topic"
	default:
		e.Type = "direct"
	}
}

func ReadTopologyFromFile(path string) (Topology, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return Topology{}, err
	}

	var out Topology
	err = yaml.Unmarshal(yamlFile, &out)
	if err != nil {
		return Topology{}, err
	}

	return out, nil
}

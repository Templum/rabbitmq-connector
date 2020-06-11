/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package types

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
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

	for _, exchange := range out {
		switch strings.ToLower(exchange.Type) {
		case "direct":
			exchange.Type = "direct"
			break
		case "topic":
			exchange.Type = "topic"
			break
		default:
			exchange.Type = "direct"
			break
		}
	}

	return out, nil
}

/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package types

import (
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// Topology definition
type Topology []struct {
	Name        string   `json:"name"`
	Topics      []string `json:"topics"`
	Declare     bool     `json:"declare"`
	Type        string   `json:"type,omitempty"`
	Durable     bool     `json:"durable,omitempty"`
	AutoDeleted bool     `json:"auto-deleted,omitempty"`
}

// Exchange Definition of a RabbitMQ Exchange
type Exchange struct {
	Name        string
	Topics      []string
	Declare     bool
	Type        string
	Durable     bool
	AutoDeleted bool
}

// EnsureCorrectType is responsible to make sure that the read-in type is one of the allowed
// which right now is direct or topic. If it is not a valid type, will default to direct.
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

// ReadTopologyFromFile reads a topology file in yaml format from the specified path.
// Further it parses the file and returns it already in the Topology struct format.
func ReadTopologyFromFile(fs afero.Fs, path string) (Topology, error) {
	yamlFile, err := afero.ReadFile(fs, path)
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

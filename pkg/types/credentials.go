package types

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Credentials is a container that contains information relevant for auth
type Credentials struct {
	User     string
	Password string
}

// NewCredentials creates a new instance using the provided information
func NewCredentials(user string, password string) *Credentials {
	return &Credentials{
		User:     user,
		Password: password,
	}
}

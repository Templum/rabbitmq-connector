package types

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Invoker is the Interface used by the OpenFaaS Connector SDK to perform invocations
// of Lambdas based on a provided topic and message
type Invoker interface {
	Invoke(topic string, message []byte)
}

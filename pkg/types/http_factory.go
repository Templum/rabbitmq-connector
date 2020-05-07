package types

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// MakeHTTPClient generates an HTTP Client setting basic properties including timeouts
func MakeHTTPClient(insecure bool, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			MaxIdleConns:        512,
			MaxIdleConnsPerHost: 512,
			IdleConnTimeout:     120 * time.Millisecond,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
		},
	}
}

/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package types

import (
	"crypto/tls"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

// MakeHTTPClient generates an HTTP Client setting basic properties including timeouts
func MakeHTTPClient(insecure bool, timeout time.Duration) *fasthttp.Client {
	client := fasthttp.Client{
		Name: "Main_Client",

		NoDefaultUserAgentHeader: false,

		Dial: fasthttpproxy.FasthttpProxyHTTPDialer(),

		ReadTimeout:  timeout,
		WriteTimeout: timeout,

		MaxIdleConnDuration: 5 * time.Second,
		/* #nosec G402 as default is false*/
		TLSConfig: &tls.Config{InsecureSkipVerify: insecure},

		MaxConnsPerHost: 256,
	}

	return &client
}

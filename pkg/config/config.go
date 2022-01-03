/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	internal "github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/openfaas/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
	"github.com/spf13/afero"
)

// Controller is the config needed for the connector
type Controller struct {
	GatewayURL          string
	RabbitConnectionURL string
	RabbitSanitizedURL  string

	IsTLSEnabled bool
	TLSConfig    *tls.Config

	Topology internal.Topology

	TopicRefreshTime   time.Duration
	BasicAuth          *auth.BasicAuthCredentials
	InsecureSkipVerify bool
	MaxClientsPerHost  int
}

// NewConfig reads the connector config from environment variables and further validates them,
// in some cases it will leverage default values.
func NewConfig(fs afero.Fs) (*Controller, error) {
	gatewayURL, err := getOpenFaaSUrl()
	if err != nil {
		return nil, err
	}

	var rabbitURL, sanitizedURL string
	var tlsConfig *tls.Config = nil

	if readFromEnv(envUseTLS, "false") == "true" {
		rabbitURL, sanitizedURL, err = getRabbitMQConnectionURL(true)

		if cfg, confErr := generateTlsConfig(fs); confErr == nil {
			tlsConfig = cfg
		} else {
			return nil, confErr
		}

	} else {
		rabbitURL, sanitizedURL, err = getRabbitMQConnectionURL(false)
	}

	if err != nil {
		return nil, err
	}

	skipVerify, err := strconv.ParseBool(readFromEnv(envSkipVerify, "false"))
	if err != nil {
		skipVerify = false
	}

	topology, err := getTopology(fs)
	if err != nil {
		return nil, err
	}

	maxClients, err := getMaxClients()
	if err != nil {
		maxClients = 256
	}

	return &Controller{
		GatewayURL: gatewayURL,
		BasicAuth:  types.GetCredentials(),

		TLSConfig: tlsConfig,

		RabbitConnectionURL: rabbitURL,
		RabbitSanitizedURL:  sanitizedURL,

		Topology: topology,

		TopicRefreshTime:   getRefreshTime(),
		InsecureSkipVerify: skipVerify,
		MaxClientsPerHost:  maxClients,
	}, nil
}

const (
	envFaaSGwURL         = "OPEN_FAAS_GW_URL"
	envSkipVerify        = "INSECURE_SKIP_VERIFY"
	envMaxClientsPerHost = "MAX_CLIENT_PER_HOST"

	envUseTLS           = "TLS_ENABLED"
	envPathToCACert     = "TLS_CA_CERT_PATH"
	envPathToServerCert = "TLS_SERVER_CERT_PATH"
	envPathToServerKey  = "TLS_SERVER_KEY_PATH"

	envRabbitUser  = "RMQ_USER"
	envRabbitPass  = "RMQ_PASS"
	envRabbitHost  = "RMQ_HOST"
	envRabbitPort  = "RMQ_PORT"
	envRabbitVHost = "RMQ_VHOST"

	envPathToTopology = "PATH_TO_TOPOLOGY"
	envRefreshTime    = "TOPIC_MAP_REFRESH_TIME"
)

func getMaxClients() (int, error) {
	return strconv.Atoi(readFromEnv(envMaxClientsPerHost, "256"))
}

func getOpenFaaSUrl() (string, error) {
	url := readFromEnv(envFaaSGwURL, "http://gateway:8080")
	if !(strings.HasPrefix(url, "http://")) && !(strings.HasPrefix(url, "https://")) {
		message := fmt.Sprintf("Provided url %s does not include the protocol http / https", url)
		return "", errors.New(message)
	}
	return url, nil
}

func generateTlsConfig(fs afero.Fs) (*tls.Config, error) {
	caCertPath := readFromEnv(envPathToCACert, "")
	if exists, err := afero.Exists(fs, caCertPath); !exists {
		return nil, fmt.Errorf("Ca Cert at %s does not exist or is not accessible %s", caCertPath, err)
	}

	serverCertPath := readFromEnv(envPathToServerCert, "")
	if exists, err := afero.Exists(fs, serverCertPath); !exists {
		return nil, fmt.Errorf("Server Cert at %s does not exist or is not accessible %s", serverCertPath, err)
	}

	serverKeyPath := readFromEnv(envPathToServerKey, "")
	if exists, err := afero.Exists(fs, serverKeyPath); !exists {
		return nil, fmt.Errorf("Server Key at %s does not exist or is not accessible %s", serverKeyPath, err)
	}

	// At this point we know every required file is present and accessible
	cfg := new(tls.Config)
	cfg.RootCAs = x509.NewCertPool()

	if ca, err := afero.ReadFile(fs, caCertPath); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	} else {
		return nil, err
	}

	if cert, err := afero.ReadFile(fs, serverCertPath); err == nil {
		if key, err := afero.ReadFile(fs, serverKeyPath); err == nil {
			if cert, err := tls.X509KeyPair(cert, key); err == nil {
				cfg.Certificates = append(cfg.Certificates, cert)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	return cfg, nil
}

// getRabbitMQConnectionURL returns the fully build url and the sanitized version for usage in logging
func getRabbitMQConnectionURL(isTls bool) (string, string, error) {
	user := readFromEnv(envRabbitUser, "")
	pass := readFromEnv(envRabbitPass, "")
	host := readFromEnv(envRabbitHost, "localhost")
	port := readFromEnv(envRabbitPort, "5672")
	vhost := readFromEnv(envRabbitVHost, "")
	protocol := "amqp"
	if isTls {
		protocol = "amqps"
	}

	parsedPort, err := strconv.Atoi(port)

	if err != nil {
		message := fmt.Sprintf("Provided port %s is not a valid port", port)
		return "", "", errors.New(message)
	}

	if parsedPort <= 0 || parsedPort > 65535 {
		message := fmt.Sprintf("Provided port %s is outside of the allowed port range", port)
		return "", "", errors.New(message)
	}

	if user == "" && pass == "" {
		return fmt.Sprintf("%s://%s:%s/%s", protocol, host, port, vhost), fmt.Sprintf("%s://%s:%s/%s", protocol, host, port, vhost), nil
	}

	return fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, user, pass, host, port, vhost), fmt.Sprintf("%s://%s:%s/%s", protocol, host, port, vhost), nil
}

func getTopology(fs afero.Fs) (internal.Topology, error) {
	path := readFromEnv(envPathToTopology, ".")

	if info, err := fs.Stat(path); os.IsNotExist(err) || !strings.HasSuffix(info.Name(), "yaml") {
		return internal.Topology{}, errors.New("provided topology is either non existing or does not end with .yaml")
	}

	return internal.ReadTopologyFromFile(fs, path)
}

func getRefreshTime() time.Duration {
	refreshTime, err := time.ParseDuration(readFromEnv(envRefreshTime, "30s"))
	if err != nil {
		log.Println("Provided Topicmap Refresh Time was not a valid Duration, like 30s or 60ms. Falling back to 30s")
		refreshTime, _ = time.ParseDuration("30s")
	}

	return refreshTime
}

// Helper Functions
func readFromEnv(env string, fallback string) string {
	if val, exists := os.LookupEnv(env); exists {
		return val
	}

	return fallback
}

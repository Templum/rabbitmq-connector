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
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	internal "github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/openfaas/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
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
func NewConfig() (*Controller, error) {
	gatewayURL, err := getOpenFaaSUrl()
	if err != nil {
		return nil, err
	}

	var rabbitURL, sanitizedURL string
	var tlsConfig *tls.Config = nil

	if readFromEnv(envUseTLS, "false") == "true" {
		rabbitURL, err = getRabbitMQTLSConnectionURL()
		sanitizedURL = getSanitizedRabbitMQURL(true)

		if cfg, confErr := generateTlsConfig(); confErr == nil {
			tlsConfig = cfg
		} else {
			return nil, confErr
		}

	} else {
		rabbitURL, err = getRabbitMQConnectionURL()
		sanitizedURL = getSanitizedRabbitMQURL(false)
	}

	if err != nil {
		return nil, err
	}

	skipVerify, err := strconv.ParseBool(readFromEnv(envSkipVerify, "false"))
	if err != nil {
		skipVerify = false
	}

	topology, err := getTopology()
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
	envPathToClientCert = "TLS_CLIENT_CERT_PATH"
	envPathToClientKey  = "TLS_CLIENT_KEY_PATH"

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

func getRabbitMQTLSConnectionURL() (string, error) {
	host := readFromEnv(envRabbitHost, "localhost")
	port := readFromEnv(envRabbitPort, "5672")
	vhost := readFromEnv(envRabbitVHost, "")

	parsedPort, err := strconv.Atoi(port)

	if err != nil {
		message := fmt.Sprintf("Provided port %s is not a valid port", port)
		return "", errors.New(message)
	}

	if parsedPort <= 0 || parsedPort > 65535 {
		message := fmt.Sprintf("Provided port %s is outside of the allowed port range", port)
		return "", errors.New(message)
	}

	return fmt.Sprintf("amqps://%s:%s/%s", host, port, vhost), nil
}

func generateTlsConfig() (*tls.Config, error) {
	caCertPath := readFromEnv(envPathToCACert, "")
	if caCertPath == "" {
		return nil, errors.New("no path to CA cert was provided")
	}
	if exists, err := doesFileExist(caCertPath); !exists {
		return nil, err
	}

	clientCertPath := readFromEnv(envPathToClientCert, "")
	if clientCertPath == "" {
		return nil, errors.New("no path to Client cert was provided")
	}
	if exists, err := doesFileExist(clientCertPath); !exists {
		return nil, err
	}

	clientKeyPath := readFromEnv(envPathToClientKey, "")
	if clientKeyPath == "" {
		return nil, errors.New("no path to Client key was provided")
	}
	if exists, err := doesFileExist(clientKeyPath); !exists {
		return nil, err
	}

	// At this point we know every required file is present and accessible
	cfg := new(tls.Config)
	cfg.RootCAs = x509.NewCertPool()

	if ca, err := ioutil.ReadFile(caCertPath); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	} else {
		return nil, err
	}

	if cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath); err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	} else {
		return nil, err
	}

	return cfg, nil
}

func getRabbitMQConnectionURL() (string, error) {
	user := readFromEnv(envRabbitUser, "user")
	pass := readFromEnv(envRabbitPass, "pass")
	host := readFromEnv(envRabbitHost, "localhost")
	port := readFromEnv(envRabbitPort, "5672")
	vhost := readFromEnv(envRabbitVHost, "")

	parsedPort, err := strconv.Atoi(port)

	if err != nil {
		message := fmt.Sprintf("Provided port %s is not a valid port", port)
		return "", errors.New(message)
	}

	if parsedPort <= 0 || parsedPort > 65535 {
		message := fmt.Sprintf("Provided port %s is outside of the allowed port range", port)
		return "", errors.New(message)
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, pass, host, port, vhost), nil
}

func getSanitizedRabbitMQURL(isTls bool) string {
	host := readFromEnv(envRabbitHost, "localhost")
	port := readFromEnv(envRabbitPort, "5672")
	vhost := readFromEnv(envRabbitVHost, "")

	if isTls {
		return fmt.Sprintf("amqps://%s:%s/%s", host, port, vhost)
	}

	return fmt.Sprintf("amqp://%s:%s/%s", host, port, vhost)
}

func getTopology() (internal.Topology, error) {
	path := readFromEnv(envPathToTopology, ".")

	if info, err := os.Stat(path); os.IsNotExist(err) || !strings.HasSuffix(info.Name(), "yaml") {
		return internal.Topology{}, errors.New("provided topology is either non existing or does not end with .yaml")
	}

	return internal.ReadTopologyFromFile(path)
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

func doesFileExist(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return false, fmt.Errorf("%s is a directory and not a file", path)
		}
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

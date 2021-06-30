/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func createTestCertBundle() ([]byte, []byte, []byte, error) {
	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1503),
		Subject: pkix.Name{
			Organization:  []string{"Opensource, INC."},
			Country:       []string{"DE"},
			Province:      []string{""},
			Locality:      []string{"Mannheim"},
			StreetAddress: []string{"P3"},
			PostalCode:    []string{"68161"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	caCert := new(bytes.Buffer)
	_ = pem.Encode(caCert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1503),
		Subject: pkix.Name{
			Organization:  []string{"Opensource, INC."},
			Country:       []string{"DE"},
			Province:      []string{""},
			Locality:      []string{"Mannheim"},
			StreetAddress: []string{"P3"},
			PostalCode:    []string{"68161"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	clientCert := new(bytes.Buffer)
	_ = pem.Encode(clientCert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	clientKey := new(bytes.Buffer)
	_ = pem.Encode(clientKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	return caCert.Bytes(), clientCert.Bytes(), clientKey.Bytes(), nil
}

func TestNewConfig(t *testing.T) {
	testFS := afero.NewMemMapFs()

	// Creating relevant structure
	_ = testFS.MkdirAll("config", 0755)
	_ = afero.WriteFile(testFS, "config/topology.yaml", []byte(`- name: AEx
  topics: [Foo, Bar]
  declare: true
  type: "direct"
  durable: false
  auto-deleted: false
- name: BEx
  topics: [Dead, Beef]
  declare: true`), 0644)

	_ = afero.WriteFile(testFS, "config/not-topology.yaml", []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: rabbitmq-connector-configmap
data:
  OPEN_FAAS_GW_URL: "http://gateway.openfaas:8080"
  RMQ_TOPICS: "account,billing,support"
  RMQ_HOST: "replace_me"
  RMQ_PORT: "replace_me"
  RMQ_USER: "replace_me"
  RMQ_PASS: "replace_me"
  REQ_TIMEOUT: "30s"
  TOPIC_MAP_REFRESH_TIME: "30s"`), 0644)

	pathToExampleToplogy := path.Join("config", "topology.yaml")

	t.Run("With invalid Gateway Url", func(t *testing.T) {
		os.Setenv("OPEN_FAAS_GW_URL", "gateway:8080")
		defer os.Unsetenv("OPEN_FAAS_GW_URL")

		var err error

		_, err = NewConfig(testFS)

		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "does not include the protocol http / https", "Did not throw correct error")

		os.Setenv("OPEN_FAAS_GW_URL", "tcp://gateway:8080")
		_, err = NewConfig(testFS)
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "does not include the protocol http / https", "Did not throw correct error")
	})

	t.Run("With invalid Rabbit MQ Port", func(t *testing.T) {
		os.Setenv("RMQ_PORT", "is_string")
		defer os.Unsetenv("RMQ_PORT")

		var err error

		_, err = NewConfig(testFS)
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "is not a valid port", "Did not throw correct error")

		os.Setenv("RMQ_PORT", "-1")
		_, err = NewConfig(testFS)
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "is outside of the allowed port range", "Did not throw correct error")

		os.Setenv("RMQ_PORT", "65536")
		_, err = NewConfig(testFS)
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "is outside of the allowed port range", "Did not throw correct error")
	})

	t.Run("With invalid RefreshTime", func(t *testing.T) {
		os.Setenv("TOPIC_MAP_REFRESH_TIME", "is_string")
		defer os.Unsetenv("TOPIC_MAP_REFRESH_TIME")

		var duration time.Duration

		duration = getRefreshTime()
		assert.Equal(t, duration, 30*time.Second, "Should fallback to 30s")

		os.Setenv("TOPIC_MAP_REFRESH_TIME", "66,31h")
		duration = getRefreshTime()
		assert.Equal(t, duration, 30*time.Second, "Should fallback to 30s")
	})

	t.Run("With invalid SkipVerify", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		os.Setenv("INSECURE_SKIP_VERIFY", "is_string")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")
		defer os.Unsetenv("INSECURE_SKIP_VERIFY")

		config, err := NewConfig(testFS)

		assert.Nil(t, err, "Should not throw")
		assert.False(t, config.InsecureSkipVerify, "Expected default value")
	})

	t.Run("With invalid max clients", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		os.Setenv("MAX_CLIENT_PER_HOST", "fifty")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")
		defer os.Unsetenv("MAX_CLIENT_PER_HOST")

		config, err := NewConfig(testFS)

		assert.Nil(t, err, "Should not throw")
		assert.Equal(t, config.MaxClientsPerHost, 256, "Expected default value")
	})

	t.Run("With non existing Topology", func(t *testing.T) {
		_, err := NewConfig(testFS)
		assert.Error(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "provided topology is either non existing or does not end with .yaml")
	})

	t.Run("With invalid Topology", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", "config/not-topology.yaml")
		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		_, err := NewConfig(testFS)
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), " cannot unmarshal")
	})

	t.Run("Default Config", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		config, err := NewConfig(testFS)

		assert.Nil(t, err, "Should not throw")
		assert.Nil(t, config.TLSConfig, "Should not have a TLS config")
		assert.Equal(t, config.GatewayURL, "http://gateway:8080", "Expected default value")
		assert.Equal(t, config.RabbitConnectionURL, "amqp://user:pass@localhost:5672/", "Expected default value")
		assert.NotContains(t, config.RabbitSanitizedURL, "user:pass", "Expected credentials not to be present")
		assert.Equal(t, config.RabbitSanitizedURL, "amqp://localhost:5672/", "Expected default value")
		assert.Equal(t, config.TopicRefreshTime, 30*time.Second, "Expected default value")
		assert.False(t, config.InsecureSkipVerify, "Expected default value")
		assert.Equal(t, config.MaxClientsPerHost, 256, "Expected default value")
	})

	t.Run("Override Config", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		os.Setenv("RMQ_HOST", "rabbit")
		os.Setenv("RMQ_PORT", "1337")
		os.Setenv("RMQ_USER", "username")
		os.Setenv("RMQ_PASS", "password")
		os.Setenv("RMQ_VHOST", "other")
		os.Setenv("OPEN_FAAS_GW_URL", "https://gateway")
		os.Setenv("TOPIC_MAP_REFRESH_TIME", "40s")
		os.Setenv("INSECURE_SKIP_VERIFY", "true")
		os.Setenv("MAX_CLIENT_PER_HOST", "512")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")
		defer os.Unsetenv("RMQ_HOST")
		defer os.Unsetenv("RMQ_PORT")
		defer os.Unsetenv("RMQ_USER")
		defer os.Unsetenv("RMQ_PASS")
		defer os.Unsetenv("RMQ_VHOST")
		defer os.Unsetenv("OPEN_FAAS_GW_URL")
		defer os.Unsetenv("TOPIC_MAP_REFRESH_TIME")
		defer os.Unsetenv("INSECURE_SKIP_VERIFY")
		defer os.Unsetenv("MAX_CLIENT_PER_HOST")

		config, err := NewConfig(testFS)

		assert.Nil(t, err, "Should not throw")
		assert.Nil(t, config.TLSConfig, "Should not have a TLS config")
		assert.Equal(t, config.GatewayURL, "https://gateway", "Expected override value")
		assert.Equal(t, config.RabbitConnectionURL, "amqp://username:password@rabbit:1337/other", "Expected override value")
		assert.NotContains(t, config.RabbitSanitizedURL, "username:password", "Expected credentials not to be present")
		assert.Equal(t, config.RabbitSanitizedURL, "amqp://rabbit:1337/other", "Expected override value")
		assert.Equal(t, config.TopicRefreshTime, 40*time.Second, "Expected override value")
		assert.True(t, config.InsecureSkipVerify, "Expected override value")
		assert.Equal(t, config.MaxClientsPerHost, 512, "Expected override value")
	})

	// TLS Specific Setup Code

	tlsTestFS := afero.NewMemMapFs()

	// Creating relevant structure
	_ = tlsTestFS.MkdirAll("config", 0755)
	_ = afero.WriteFile(tlsTestFS, "config/topology.yaml", []byte(`- name: AEx
  topics: [Foo, Bar]
  declare: true
  type: "direct"
  durable: false
  auto-deleted: false
- name: BEx
  topics: [Dead, Beef]
  declare: true`), 0644)

	caCert, clientCert, clientKey, err := createTestCertBundle()
	if err != nil {
		t.Fatalf("createTestCertBundle failed with %s", err)
	}

	_ = afero.WriteFile(tlsTestFS, "config/ca.pem", caCert, 0644)
	_ = afero.WriteFile(tlsTestFS, "config/client.pem", clientCert, 0644)
	_ = afero.WriteFile(tlsTestFS, "config/client.key", clientKey, 0644)

	pathToCACert := path.Join("config", "ca.pem")
	pathToServerCert := path.Join("config", "client.pem")
	pathToServerKey := path.Join("config", "client.key")

	t.Run("TLS based Config", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)

		os.Setenv("TLS_ENABLED", "true")
		os.Setenv("TLS_CA_CERT_PATH", pathToCACert)
		os.Setenv("TLS_SERVER_CERT_PATH", pathToServerCert)
		os.Setenv("TLS_SERVER_KEY_PATH", pathToServerKey)

		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		defer os.Unsetenv("TLS_ENABLED")
		defer os.Unsetenv("TLS_CA_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_KEY_PATH")

		config, err := NewConfig(tlsTestFS)

		assert.Nil(t, err, "Should not throw")

		assert.Equal(t, config.RabbitConnectionURL, "amqps://localhost:5672/", "Expected default value")
		assert.NotContains(t, config.RabbitSanitizedURL, "user:pass", "Expected credentials not to be present")
		assert.Equal(t, config.RabbitSanitizedURL, "amqps://localhost:5672/", "Expected default value")

		assert.Len(t, config.TLSConfig.Certificates, 1, "Should only have the server cert in the chain")
	})

	t.Run("TLS config without a ca at target path", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)

		os.Setenv("TLS_ENABLED", "true")
		os.Setenv("TLS_CA_CERT_PATH", "config/notca.pem")
		os.Setenv("TLS_SERVER_CERT_PATH", pathToServerCert)
		os.Setenv("TLS_SERVER_KEY_PATH", pathToServerKey)

		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		defer os.Unsetenv("TLS_ENABLED")
		defer os.Unsetenv("TLS_CA_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_KEY_PATH")

		config, err := NewConfig(tlsTestFS)

		assert.Nil(t, config, "Should return not config")
		assert.Error(t, err, "should throw")
		assert.Contains(t, err.Error(), "Ca Cert at config/notca.pem", "Message should point to CA cert")
	})

	t.Run("TLS config without a server cert at target path", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)

		os.Setenv("TLS_ENABLED", "true")
		os.Setenv("TLS_CA_CERT_PATH", pathToCACert)
		os.Setenv("TLS_SERVER_CERT_PATH", "config/notserver.pem")
		os.Setenv("TLS_SERVER_KEY_PATH", pathToServerKey)

		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		defer os.Unsetenv("TLS_ENABLED")
		defer os.Unsetenv("TLS_CA_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_KEY_PATH")

		config, err := NewConfig(tlsTestFS)

		assert.Nil(t, config, "Should return not config")
		assert.Error(t, err, "should throw")
		assert.Contains(t, err.Error(), "Server Cert at config/notserver.pem", "Message should point to Server cert")
	})

	t.Run("TLS config without a server cert at target path", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)

		os.Setenv("TLS_ENABLED", "true")
		os.Setenv("TLS_CA_CERT_PATH", pathToCACert)
		os.Setenv("TLS_SERVER_CERT_PATH", pathToServerKey)
		os.Setenv("TLS_SERVER_KEY_PATH", "config/notserver.key")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		defer os.Unsetenv("TLS_ENABLED")
		defer os.Unsetenv("TLS_CA_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_CERT_PATH")
		defer os.Unsetenv("TLS_SERVER_KEY_PATH")

		config, err := NewConfig(tlsTestFS)

		assert.Nil(t, config, "Should return not config")
		assert.Error(t, err, "should throw")
		assert.Contains(t, err.Error(), "Server Key at config/notserver.key", "Message should point to Server key")
	})
}

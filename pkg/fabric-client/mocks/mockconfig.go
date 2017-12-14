/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// MockConfig ...
type MockConfig struct {
	tlsEnabled       bool
	mutualTLSEnabled bool
	errorCase        bool
}

// NewMockConfig ...
func NewMockConfig() config.Config {
	return &MockConfig{}
}

// NewMockConfigCustomized ...
func NewMockConfigCustomized(tlsEnabled, mutualTLSEnabled, errorCase bool) config.Config {
	return &MockConfig{tlsEnabled: tlsEnabled, mutualTLSEnabled: mutualTLSEnabled, errorCase: errorCase}
}

// Client ...
func (c *MockConfig) Client() (*config.ClientConfig, error) {
	if c.mutualTLSEnabled {
		mutualTLSCerts := config.MutualTLSConfig{
			Client: struct {
				KeyPem   string
				Keyfile  string
				CertPem  string
				Certfile string
			}{KeyPem: "", Keyfile: "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem", CertPem: "", Certfile: "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"},
		}

		return &config.ClientConfig{TLSCerts: mutualTLSCerts}, nil
	}

	return &config.ClientConfig{}, nil
}

// CAConfig not implemented
func (c *MockConfig) CAConfig(org string) (*config.CAConfig, error) {
	return nil, nil
}

//CAServerCertPems Read configuration option for the server certificate embedded pems
func (c *MockConfig) CAServerCertPems(org string) ([]string, error) {
	return nil, nil
}

//CAServerCertPaths Read configuration option for the server certificate files
func (c *MockConfig) CAServerCertPaths(org string) ([]string, error) {
	return nil, nil
}

//CAClientKeyPem Read configuration option for the fabric CA client key from a string
func (c *MockConfig) CAClientKeyPem(org string) (string, error) {
	return "", nil
}

//CAClientKeyPath Read configuration option for the fabric CA client key file
func (c *MockConfig) CAClientKeyPath(org string) (string, error) {
	return "", nil
}

//CAClientCertPem Read configuration option for the fabric CA client cert from a string
func (c *MockConfig) CAClientCertPem(org string) (string, error) {
	return "", nil
}

//CAClientCertPath Read configuration option for the fabric CA client cert file
func (c *MockConfig) CAClientCertPath(org string) (string, error) {
	return "", nil
}

//TimeoutOrDefault not implemented
func (c *MockConfig) TimeoutOrDefault(arg config.TimeoutType) time.Duration {
	return time.Second * 10
}

// PeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) PeersConfig(org string) ([]config.PeerConfig, error) {
	return nil, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *MockConfig) PeerConfig(org string, name string) (*config.PeerConfig, error) {
	return nil, nil
}

// TLSCACertPool ...
func (c *MockConfig) TLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	if c.errorCase {
		return nil, errors.New("just to test error scenario")
	}
	return nil, nil
}

// SetTLSCACertPool ...
func (c *MockConfig) SetTLSCACertPool(pool *x509.CertPool) {
}

// TcertBatchSize ...
func (c *MockConfig) TcertBatchSize() int {
	return 0
}

// SecurityAlgorithm ...
func (c *MockConfig) SecurityAlgorithm() string {
	return "SHA2"
}

// SecurityLevel ...
func (c *MockConfig) SecurityLevel() int {
	return 256

}

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *MockConfig) SecurityProviderLibPath() string {
	return ""
}

// OrderersConfig returns a list of defined orderers
func (c *MockConfig) OrderersConfig() ([]config.OrdererConfig, error) {
	return nil, nil
}

// RandomOrdererConfig not implemented
func (c *MockConfig) RandomOrdererConfig() (*config.OrdererConfig, error) {
	return nil, nil
}

// OrdererConfig not implemented
func (c *MockConfig) OrdererConfig(name string) (*config.OrdererConfig, error) {
	return nil, nil
}

// MspID not implemented
func (c *MockConfig) MspID(org string) (string, error) {
	return "", nil
}

// PeerMspID not implemented
func (c *MockConfig) PeerMspID(name string) (string, error) {
	return "", nil
}

// KeyStorePath ...
func (c *MockConfig) KeyStorePath() string {
	return ""
}

// CAKeyStorePath not implemented
func (c *MockConfig) CAKeyStorePath() string {
	return ""
}

// CryptoConfigPath ...
func (c *MockConfig) CryptoConfigPath() string {
	return ""
}

// NetworkConfig not implemented
func (c *MockConfig) NetworkConfig() (*config.NetworkConfig, error) {
	return nil, nil
}

// ChannelConfig returns the channel configuration
func (c *MockConfig) ChannelConfig(name string) (*config.ChannelConfig, error) {
	return nil, nil
}

// ChannelPeers returns the channel peers configuration
func (c *MockConfig) ChannelPeers(name string) ([]config.ChannelPeer, error) {
	return nil, nil
}

// ChannelOrderers returns a list of channel orderers
func (c *MockConfig) ChannelOrderers(name string) ([]config.OrdererConfig, error) {
	return nil, nil
}

// NetworkPeers returns the mock network peers configuration
func (c *MockConfig) NetworkPeers() ([]config.NetworkPeer, error) {
	return nil, nil
}

// Ephemeral flag
func (c *MockConfig) Ephemeral() bool {
	return false
}

// SecurityProvider ...
func (c *MockConfig) SecurityProvider() string {
	return "SW"
}

// SecurityProviderLabel ...
func (c *MockConfig) SecurityProviderLabel() string {
	return ""
}

//SecurityProviderPin ...
func (c *MockConfig) SecurityProviderPin() string {
	return ""
}

//SoftVerify flag
func (c *MockConfig) SoftVerify() bool {
	return false
}

// IsSecurityEnabled ...
func (c *MockConfig) IsSecurityEnabled() bool {
	return false
}

// TLSClientCerts ...
func (c *MockConfig) TLSClientCerts() ([]tls.Certificate, error) {
	return nil, nil
}

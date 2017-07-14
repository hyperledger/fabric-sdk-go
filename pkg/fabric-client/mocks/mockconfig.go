/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/x509"
	"time"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	"fmt"

	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	"github.com/spf13/viper"
)

// MockConfig ...
type MockConfig struct {
	tlsEnabled bool
	errorCase  bool
}

// NewMockConfig ...
func NewMockConfig() config.Config {
	return &MockConfig{}
}

// NewMockConfigCustomized ...
func NewMockConfigCustomized(tlsEnabled bool, errorCase bool) config.Config {
	return &MockConfig{tlsEnabled: tlsEnabled, errorCase: errorCase}
}

// CAConfig not implemented
func (c *MockConfig) CAConfig(org string) (*config.CAConfig, error) {
	return nil, nil
}

//CAServerCertFiles Read configuration option for the server certificate files
func (c *MockConfig) CAServerCertFiles(org string) ([]string, error) {
	return nil, nil
}

//CAClientKeyFile Read configuration option for the fabric CA client key file
func (c *MockConfig) CAClientKeyFile(org string) (string, error) {
	return "", nil
}

//CAClientCertFile Read configuration option for the fabric CA client cert file
func (c *MockConfig) CAClientCertFile(org string) (string, error) {
	return "", nil
}

//TimeoutOrDefault not implemented
func (c *MockConfig) TimeoutOrDefault(config.ConnectionType) time.Duration {
	return 0
}

// FabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func (c *MockConfig) FabricClientViper() *viper.Viper {
	return nil
}

// PeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) PeersConfig(org string) ([]config.PeerConfig, error) {
	return nil, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *MockConfig) PeerConfig(org string, name string) (*config.PeerConfig, error) {
	return nil, nil
}

// IsTLSEnabled ...
func (c *MockConfig) IsTLSEnabled() bool {
	return c.tlsEnabled
}

// TLSCACertPool ...
func (c *MockConfig) TLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	if c.errorCase {
		return nil, fmt.Errorf("just to test error scenario")
	}
	return nil, nil
}

// SetTLSCACertPool ...
func (c *MockConfig) SetTLSCACertPool(pool *x509.CertPool) {
}

// IsSecurityEnabled ...
func (c *MockConfig) IsSecurityEnabled() bool {
	return false
}

// TcertBatchSize ...
func (c *MockConfig) TcertBatchSize() int {
	return 0
}

// SecurityAlgorithm ...
func (c *MockConfig) SecurityAlgorithm() string {
	return ""
}

// SecurityLevel ...
func (c *MockConfig) SecurityLevel() int {
	return 0

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

// MspID ...
func (c *MockConfig) MspID(org string) (string, error) {
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

// CSPConfig ...
func (c *MockConfig) CSPConfig() *bccspFactory.FactoryOpts {
	return nil
}

// NetworkConfig not implemented
func (c *MockConfig) NetworkConfig() (*config.NetworkConfig, error) {
	return nil, nil
}

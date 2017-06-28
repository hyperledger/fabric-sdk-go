/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/x509"

	api "github.com/hyperledger/fabric-sdk-go/api"

	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	"github.com/spf13/viper"
)

// MockConfig ...
type MockConfig struct {
}

// NewMockConfig ...
func NewMockConfig() api.Config {
	return &MockConfig{}
}

// GetCAConfig not implemented
func (c *MockConfig) GetCAConfig(org string) (*api.CAConfig, error) {
	return nil, nil
}

//GetCAServerCertFiles Read configuration option for the server certificate files
func (c *MockConfig) GetCAServerCertFiles(org string) ([]string, error) {
	return nil, nil
}

//GetCAClientKeyFile Read configuration option for the fabric CA client key file
func (c *MockConfig) GetCAClientKeyFile(org string) (string, error) {
	return "", nil
}

//GetCAClientCertFile Read configuration option for the fabric CA client cert file
func (c *MockConfig) GetCAClientCertFile(org string) (string, error) {
	return "", nil
}

// GetFabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func (c *MockConfig) GetFabricClientViper() *viper.Viper {
	return nil
}

// GetPeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) GetPeersConfig(org string) ([]api.PeerConfig, error) {
	return nil, nil
}

// IsTLSEnabled ...
func (c *MockConfig) IsTLSEnabled() bool {
	return false
}

// GetTLSCACertPool ...
func (c *MockConfig) GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	return nil, nil
}

// GetTLSCACertPoolFromRoots ...
func (c *MockConfig) GetTLSCACertPoolFromRoots(ordererRootCAs [][]byte) (*x509.CertPool, error) {
	return nil, nil
}

// IsSecurityEnabled ...
func (c *MockConfig) IsSecurityEnabled() bool {
	return false
}

// TcertBatchSize ...
func (c *MockConfig) TcertBatchSize() int {
	return 0
}

// GetSecurityAlgorithm ...
func (c *MockConfig) GetSecurityAlgorithm() string {
	return ""
}

// GetSecurityLevel ...
func (c *MockConfig) GetSecurityLevel() int {
	return 0

}

// GetRandomOrdererConfig not implemented
func (c *MockConfig) GetRandomOrdererConfig() (*api.OrdererConfig, error) {
	return nil, nil
}

// GetOrdererConfig not implemented
func (c *MockConfig) GetOrdererConfig(name string) (*api.OrdererConfig, error) {
	return nil, nil
}

// GetMspID ...
func (c *MockConfig) GetMspID(org string) (string, error) {
	return "", nil
}

// GetKeyStorePath ...
func (c *MockConfig) GetKeyStorePath() string {
	return ""
}

// GetCAKeyStorePath not implemented
func (c *MockConfig) GetCAKeyStorePath() string {
	return ""
}

// GetCryptoConfigPath ...
func (c *MockConfig) GetCryptoConfigPath() string {
	return ""
}

// GetCSPConfig ...
func (c *MockConfig) GetCSPConfig() *bccspFactory.FactoryOpts {
	return nil
}

// GetNetworkConfig not implemented
func (c *MockConfig) GetNetworkConfig() (*api.NetworkConfig, error) {
	return nil, nil
}

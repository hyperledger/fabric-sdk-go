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

//GetServerURL Read configuration option for the fabric CA server URL
func (c *MockConfig) GetServerURL() string {
	return ""
}

//GetServerCertFiles Read configuration option for the server certificate files
func (c *MockConfig) GetServerCertFiles() []string {
	return nil
}

//GetFabricCAClientKeyFile Read configuration option for the fabric CA client key file
func (c *MockConfig) GetFabricCAClientKeyFile() string {
	return ""
}

//GetFabricCAClientCertFile Read configuration option for the fabric CA client cert file
func (c *MockConfig) GetFabricCAClientCertFile() string {
	return ""
}

//GetFabricCATLSEnabledFlag Read configuration option for the fabric CA TLS flag
func (c *MockConfig) GetFabricCATLSEnabledFlag() bool {
	return false
}

// GetFabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func (c *MockConfig) GetFabricClientViper() *viper.Viper {
	return nil
}

// GetPeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) GetPeersConfig() ([]api.PeerConfig, error) {
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

// GetOrdererHost ...
func (c *MockConfig) GetOrdererHost() string {
	return ""
}

// GetOrdererPort ...
func (c *MockConfig) GetOrdererPort() string {
	return ""
}

// GetOrdererTLSServerHostOverride ...
func (c *MockConfig) GetOrdererTLSServerHostOverride() string {
	return ""
}

// GetOrdererTLSCertificate ...
func (c *MockConfig) GetOrdererTLSCertificate() string {
	return ""
}

// GetFabricCAID ...
func (c *MockConfig) GetFabricCAID() string {
	return ""
}

//GetFabricCAName Read the fabric CA name
func (c *MockConfig) GetFabricCAName() string {
	return ""
}

// GetKeyStorePath ...
func (c *MockConfig) GetKeyStorePath() string {
	return ""
}

// GetFabricCAHomeDir ...
func (c *MockConfig) GetFabricCAHomeDir() string {
	return ""
}

// GetFabricCAMspDir ...
func (c *MockConfig) GetFabricCAMspDir() string {
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

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	bccspFactory "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/spf13/viper"
)

// MockConfig ...
type MockConfig struct {
	CAServerURL string
}

// NewMockConfig ...
func NewMockConfig(CAServerURL string) apiconfig.Config {
	return &MockConfig{CAServerURL: CAServerURL}
}

// Client returns the Client config
func (c *MockConfig) Client() (*apiconfig.ClientConfig, error) {
	return nil, nil
}

// CAConfig return ca configuration
func (c *MockConfig) CAConfig(org string) (*apiconfig.CAConfig, error) {
	return &apiconfig.CAConfig{Url: c.CAServerURL, CaName: "test", TlsCACerts: apiconfig.MutualTLSConfig{}}, nil
}

// CAServerCertFiles Read configuration option for the server certificate files
func (c *MockConfig) CAServerCertFiles(org string) ([]string, error) {
	return nil, nil
}

// CAClientKeyFile Read configuration option for the fabric CA client key file
func (c *MockConfig) CAClientKeyFile(org string) (string, error) {
	return "", nil
}

// CAClientCertFile Read configuration option for the fabric CA client cert file
func (c *MockConfig) CAClientCertFile(org string) (string, error) {
	return "", nil
}

// FabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func (c *MockConfig) FabricClientViper() *viper.Viper {
	return nil
}

//TimeoutOrDefault not implemented
func (c *MockConfig) TimeoutOrDefault(apiconfig.ConnectionType) time.Duration {
	return 0
}

// PeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) PeersConfig(org string) ([]apiconfig.PeerConfig, error) {
	return nil, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *MockConfig) PeerConfig(org string, name string) (*apiconfig.PeerConfig, error) {
	return nil, nil
}

// IsTLSEnabled ...
func (c *MockConfig) IsTLSEnabled() bool {
	return false
}

// TLSCACertPool ...
func (c *MockConfig) TLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
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
	return ""
}

// SecurityLevel ...
func (c *MockConfig) SecurityLevel() int {
	return 0

}

// OrderersConfig returns a list of defined orderers
func (c *MockConfig) OrderersConfig() ([]apiconfig.OrdererConfig, error) {
	return nil, nil
}

// RandomOrdererConfig not implemented
func (c *MockConfig) RandomOrdererConfig() (*apiconfig.OrdererConfig, error) {
	return nil, nil
}

// OrdererConfig not implemented
func (c *MockConfig) OrdererConfig(name string) (*apiconfig.OrdererConfig, error) {
	return nil, nil
}

// MspID ...
func (c *MockConfig) MspID(org string) (string, error) {
	return "", nil
}

// KeyStorePath ...
func (c *MockConfig) KeyStorePath() string {
	return "/tmp/msp"
}

// CAKeyStorePath ...
func (c *MockConfig) CAKeyStorePath() string {
	return "/tmp/msp"
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
func (c *MockConfig) NetworkConfig() (*apiconfig.NetworkConfig, error) {
	return nil, nil
}

//SecurityProvider provider SW or PKCS11
func (c *MockConfig) SecurityProvider() string {
	return "SW"
}

//Ephemeral flag
func (c *MockConfig) Ephemeral() bool {
	return false
}

//SoftVerify flag
func (c *MockConfig) SoftVerify() bool {
	return true
}

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *MockConfig) SecurityProviderLibPath() string {
	return ""
}

//SecurityProviderPin will be set only if provider is PKCS11
func (c *MockConfig) SecurityProviderPin() string {
	return ""
}

//SecurityProviderLabel will be set only if provider is PKCS11
func (c *MockConfig) SecurityProviderLabel() string {
	return ""
}

// IsSecurityEnabled ...
func (c *MockConfig) IsSecurityEnabled() bool {
	return false
}

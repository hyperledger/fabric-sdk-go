/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mockmsp

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

// MockConfig ...
type MockConfig struct {
	CAServerURL string
}

// NewMockConfig ...
func NewMockConfig(CAServerURL string) core.Config {
	return &MockConfig{CAServerURL: CAServerURL}
}

// Client returns the Client config
func (c *MockConfig) Client() (*core.ClientConfig, error) {
	return nil, nil
}

// CAConfig return ca configuration
func (c *MockConfig) CAConfig(org string) (*core.CAConfig, error) {
	return &core.CAConfig{
		URL:        c.CAServerURL,
		CAName:     "test",
		TLSCACerts: core.MutualTLSConfig{},
		Registrar: core.EnrollCredentials{
			EnrollID:     "admin",
			EnrollSecret: "adminpw",
		},
	}, nil
}

//CAServerCertPems Read configuration option for the server certificate embedded pems
func (c *MockConfig) CAServerCertPems(org string) ([]string, error) {
	return nil, nil
}

// CAServerCertPaths Read configuration option for the server certificate files
func (c *MockConfig) CAServerCertPaths(org string) ([]string, error) {
	return nil, nil
}

//CAClientKeyPem Read configuration option for the fabric CA client key from a string
func (c *MockConfig) CAClientKeyPem(org string) (string, error) {
	return "", nil
}

// CAClientKeyPath Read configuration option for the fabric CA client key file
func (c *MockConfig) CAClientKeyPath(org string) (string, error) {
	return "", nil
}

//CAClientCertPem Read configuration option for the fabric CA client cert from a string
func (c *MockConfig) CAClientCertPem(org string) (string, error) {
	return "", nil
}

// CAClientCertPath Read configuration option for the fabric CA client cert file
func (c *MockConfig) CAClientCertPath(org string) (string, error) {
	return "", nil
}

//TimeoutOrDefault not implemented
func (c *MockConfig) TimeoutOrDefault(core.TimeoutType) time.Duration {
	return 0
}

//Timeout not implemented
func (c *MockConfig) Timeout(core.TimeoutType) time.Duration {
	return 0
}

// NetworkPeers returns the mock network peers configuration
func (c *MockConfig) NetworkPeers() ([]core.NetworkPeer, error) {
	return nil, nil
}

// PeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) PeersConfig(org string) ([]core.PeerConfig, error) {
	return nil, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *MockConfig) PeerConfig(org string, name string) (*core.PeerConfig, error) {
	return nil, nil
}

// PeerConfigByURL retrieves PeerConfig by URL
func (c *MockConfig) PeerConfigByURL(url string) (*core.PeerConfig, error) {
	return nil, nil
}

// ChannelOrderers returns a list of channel orderers
func (c *MockConfig) ChannelOrderers(name string) ([]core.OrdererConfig, error) {
	return nil, nil
}

// TLSCACertPool ...
func (c *MockConfig) TLSCACertPool(cert ...*x509.Certificate) (*x509.CertPool, error) {
	return nil, nil
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

// OrderersConfig returns a list of defined orderers
func (c *MockConfig) OrderersConfig() ([]core.OrdererConfig, error) {
	return nil, nil
}

// RandomOrdererConfig not implemented
func (c *MockConfig) RandomOrdererConfig() (*core.OrdererConfig, error) {
	return nil, nil
}

// OrdererConfig not implemented
func (c *MockConfig) OrdererConfig(name string) (*core.OrdererConfig, error) {
	return nil, nil
}

// MSPID ...
func (c *MockConfig) MSPID(org string) (string, error) {
	return "", nil
}

// PeerMSPID not implemented
func (c *MockConfig) PeerMSPID(name string) (string, error) {
	return "", nil
}

// CredentialStorePath ...
func (c *MockConfig) CredentialStorePath() string {
	return "/tmp/userstore"
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

// NetworkConfig not implemented
func (c *MockConfig) NetworkConfig() (*core.NetworkConfig, error) {
	return nil, nil
}

// ChannelConfig returns the channel configuration
func (c *MockConfig) ChannelConfig(name string) (*core.ChannelConfig, error) {
	return nil, nil
}

// ChannelPeers returns the channel peers configuration
func (c *MockConfig) ChannelPeers(name string) ([]core.ChannelPeer, error) {
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

// TLSClientCerts ...
func (c *MockConfig) TLSClientCerts() ([]tls.Certificate, error) {
	return nil, nil
}

// EventServiceType returns the type of event service client to use
func (c *MockConfig) EventServiceType() core.EventServiceType {
	return core.DeliverEventServiceType
}

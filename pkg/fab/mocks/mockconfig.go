/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	config "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"

	"github.com/pkg/errors"
)

// MockConfig ...
type MockConfig struct {
	tlsEnabled             bool
	mutualTLSEnabled       bool
	errorCase              bool
	customNetworkPeerCfg   []config.NetworkPeer
	customPeerCfg          *config.PeerConfig
	customOrdererCfg       *config.OrdererConfig
	customRandomOrdererCfg *config.OrdererConfig
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
	clientConfig := config.ClientConfig{}

	clientConfig.CredentialStore = config.CredentialStoreType{
		Path: "/tmp/fabsdkgo_test/store",
	}

	if c.mutualTLSEnabled {
		mutualTLSCerts := config.MutualTLSConfig{

			Client: config.TLSKeyPair{
				Key: endpoint.TLSConfig{
					Path: "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem",
					Pem:  "",
				},
				Cert: endpoint.TLSConfig{
					Path: "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem",
					Pem:  "",
				},
			},
		}
		clientConfig.TLSCerts = mutualTLSCerts
	}

	return &clientConfig, nil
}

// CAConfig not implemented
func (c *MockConfig) CAConfig(org string) (*config.CAConfig, error) {
	caConfig := config.CAConfig{
		CAName: "org1",
	}

	return &caConfig, nil
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
	return time.Second * 5
}

//Timeout not implemented
func (c *MockConfig) Timeout(arg config.TimeoutType) time.Duration {
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

// PeerConfigByURL retrieves PeerConfig by URL
func (c *MockConfig) PeerConfigByURL(url string) (*config.PeerConfig, error) {
	if url == "invalid" {
		return nil, errors.New("no orderer")
	}
	if c.customPeerCfg != nil {
		return c.customPeerCfg, nil
	}
	cfg := config.PeerConfig{
		URL: "example.com",
	}
	return &cfg, nil
}

// TLSCACertPool ...
func (c *MockConfig) TLSCACertPool(cert ...*x509.Certificate) (*x509.CertPool, error) {
	if c.errorCase {
		return nil, errors.New("just to test error scenario")
	}
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

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *MockConfig) SecurityProviderLibPath() string {
	return ""
}

// OrderersConfig returns a list of defined orderers
func (c *MockConfig) OrderersConfig() ([]config.OrdererConfig, error) {
	oConfig, err := c.OrdererConfig("")

	return []config.OrdererConfig{*oConfig}, err
}

// RandomOrdererConfig not implemented
func (c *MockConfig) RandomOrdererConfig() (*config.OrdererConfig, error) {
	if c.customRandomOrdererCfg != nil {
		return c.customRandomOrdererCfg, nil
	}
	return nil, nil
}

//SetCustomNetworkPeerCfg sets custom orderer config for unit-tests
func (c *MockConfig) SetCustomNetworkPeerCfg(customNetworkPeerCfg []config.NetworkPeer) {
	c.customNetworkPeerCfg = customNetworkPeerCfg
}

//SetCustomPeerCfg sets custom orderer config for unit-tests
func (c *MockConfig) SetCustomPeerCfg(customPeerCfg *config.PeerConfig) {
	c.customPeerCfg = customPeerCfg
}

//SetCustomOrdererCfg sets custom orderer config for unit-tests
func (c *MockConfig) SetCustomOrdererCfg(customOrdererCfg *config.OrdererConfig) {
	c.customOrdererCfg = customOrdererCfg
}

//SetCustomRandomOrdererCfg sets custom random orderer config for unit-tests
func (c *MockConfig) SetCustomRandomOrdererCfg(customRandomOrdererCfg *config.OrdererConfig) {
	c.customRandomOrdererCfg = customRandomOrdererCfg
}

// OrdererConfig not implemented
func (c *MockConfig) OrdererConfig(name string) (*config.OrdererConfig, error) {
	if name == "Invalid" {
		return nil, errors.New("no orderer")
	}
	if c.customOrdererCfg != nil {
		return c.customOrdererCfg, nil
	}
	oConfig := config.OrdererConfig{
		URL: "example.com",
	}

	return &oConfig, nil
}

// MSPID not implemented
func (c *MockConfig) MSPID(org string) (string, error) {
	return "", nil
}

// PeerMSPID not implemented
func (c *MockConfig) PeerMSPID(name string) (string, error) {
	return "", nil
}

// KeyStorePath ...
func (c *MockConfig) KeyStorePath() string {
	return "/tmp/fabsdkgo_test"
}

// CredentialStorePath ...
func (c *MockConfig) CredentialStorePath() string {
	return "/tmp/userstore"
}

// CAKeyStorePath not implemented
func (c *MockConfig) CAKeyStorePath() string {
	return "/tmp/fabsdkgo_test"
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
	if name == "Invalid" {
		return nil, errors.New("no orderer")
	}

	oConfig, err := c.OrdererConfig("")

	return []config.OrdererConfig{*oConfig}, err
}

// NetworkPeers returns the mock network peers configuration
func (c *MockConfig) NetworkPeers() ([]config.NetworkPeer, error) {
	if c.customNetworkPeerCfg != nil {
		return c.customNetworkPeerCfg, nil
	}
	return nil, errors.New("no config")
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

// EventServiceType returns the type of event service client to use
func (c *MockConfig) EventServiceType() config.EventServiceType {
	return config.DeliverEventServiceType
}

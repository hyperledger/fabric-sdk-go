/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/tls"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/pkg/errors"
)

// MockConfig ...
type MockConfig struct {
	tlsEnabled             bool
	mutualTLSEnabled       bool
	errorCase              bool
	customNetworkPeerCfg   []fab.NetworkPeer
	customPeerCfg          *fab.PeerConfig
	customOrdererCfg       *fab.OrdererConfig
	customRandomOrdererCfg *fab.OrdererConfig
	CustomTLSCACertPool    commtls.CertPool
	chConfig               map[string]*fab.ChannelEndpointConfig
}

func getConfigPath() string {
	return filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata")
}

// NewMockCryptoConfig ...
func NewMockCryptoConfig() core.CryptoSuiteConfig {
	return &MockConfig{}
}

// NewMockEndpointConfig ...
func NewMockEndpointConfig() fab.EndpointConfig {
	return &MockConfig{}
}

// NewMockIdentityConfig ...
func NewMockIdentityConfig() msp.IdentityConfig {
	return &MockConfig{}
}

// NewMockCryptoConfigCustomized ...
func NewMockCryptoConfigCustomized(tlsEnabled, mutualTLSEnabled, errorCase bool) core.CryptoSuiteConfig {
	return &MockConfig{tlsEnabled: tlsEnabled, mutualTLSEnabled: mutualTLSEnabled, errorCase: errorCase}
}

// NewMockEndpointConfigCustomized ...
func NewMockEndpointConfigCustomized(tlsEnabled, mutualTLSEnabled, errorCase bool) fab.EndpointConfig {
	return &MockConfig{tlsEnabled: tlsEnabled, mutualTLSEnabled: mutualTLSEnabled, errorCase: errorCase}
}

// NewMockIdentityConfigCustomized ...
func NewMockIdentityConfigCustomized(tlsEnabled, mutualTLSEnabled, errorCase bool) msp.IdentityConfig {
	return &MockConfig{tlsEnabled: tlsEnabled, mutualTLSEnabled: mutualTLSEnabled, errorCase: errorCase}
}

// Client ...
func (c *MockConfig) Client() *msp.ClientConfig {
	clientConfig := msp.ClientConfig{}

	clientConfig.CredentialStore = msp.CredentialStoreType{
		Path: "/tmp/fabsdkgo_test/store",
	}

	if c.mutualTLSEnabled {
		key := endpoint.TLSConfig{Path: filepath.Join(getConfigPath(), "certs", "client_sdk_go-key.pem")}
		cert := endpoint.TLSConfig{Path: filepath.Join(getConfigPath(), "certs", "client_sdk_go.pem")}

		err := key.LoadBytes()
		if err != nil {
			panic(err)
		}

		err = cert.LoadBytes()
		if err != nil {
			panic(err)
		}

		clientConfig.TLSKey = key.Bytes()
		clientConfig.TLSCert = cert.Bytes()
	}

	return &clientConfig
}

// CAConfig not implemented
func (c *MockConfig) CAConfig(org string) (*msp.CAConfig, bool) {
	caConfig := msp.CAConfig{
		ID: "org1",
	}

	return &caConfig, true
}

//CAServerCerts Read configuration option for the server certificates for given org
func (c *MockConfig) CAServerCerts(org string) ([][]byte, bool) {
	return nil, false
}

//CAClientKey Read configuration option for the fabric CA client key for given org
func (c *MockConfig) CAClientKey(org string) ([]byte, bool) {
	return nil, false
}

//CAClientCert Read configuration option for the fabric CA client cert for given org
func (c *MockConfig) CAClientCert(org string) ([]byte, bool) {
	return nil, false
}

//Timeout not implemented
func (c *MockConfig) Timeout(arg fab.TimeoutType) time.Duration {
	return time.Second * 10
}

// PeersConfig Retrieves the fabric peers from the config file provided
func (c *MockConfig) PeersConfig(org string) ([]fab.PeerConfig, bool) {
	return nil, false
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *MockConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, bool) {

	if nameOrURL == "invalid" || nameOrURL == "missing" {
		return nil, false
	}
	if c.customPeerCfg != nil {
		return c.customPeerCfg, true
	}
	cfg := fab.PeerConfig{
		URL: "example.com",
	}
	return &cfg, true
}

// TLSCACertPool ...
func (c *MockConfig) TLSCACertPool() commtls.CertPool {
	if c.errorCase {
		return &mockfab.MockCertPool{Err: errors.New("just to test error scenario")}
	} else if c.CustomTLSCACertPool != nil {
		return c.CustomTLSCACertPool
	}
	return &mockfab.MockCertPool{CertPool: x509.NewCertPool()}
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
func (c *MockConfig) OrderersConfig() []fab.OrdererConfig {
	oConfig, _, _ := c.OrdererConfig("")
	return []fab.OrdererConfig{*oConfig}
}

//SetCustomNetworkPeerCfg sets custom orderer config for unit-tests
func (c *MockConfig) SetCustomNetworkPeerCfg(customNetworkPeerCfg []fab.NetworkPeer) {
	c.customNetworkPeerCfg = customNetworkPeerCfg
}

//SetCustomPeerCfg sets custom orderer config for unit-tests
func (c *MockConfig) SetCustomPeerCfg(customPeerCfg *fab.PeerConfig) {
	c.customPeerCfg = customPeerCfg
}

//SetCustomOrdererCfg sets custom orderer config for unit-tests
func (c *MockConfig) SetCustomOrdererCfg(customOrdererCfg *fab.OrdererConfig) {
	c.customOrdererCfg = customOrdererCfg
}

//SetCustomRandomOrdererCfg sets custom random orderer config for unit-tests
func (c *MockConfig) SetCustomRandomOrdererCfg(customRandomOrdererCfg *fab.OrdererConfig) {
	c.customRandomOrdererCfg = customRandomOrdererCfg
}

// OrdererConfig not implemented
func (c *MockConfig) OrdererConfig(name string) (*fab.OrdererConfig, bool, bool) {
	if name == "Invalid" {
		return nil, false, false
	}
	if c.customOrdererCfg != nil {
		return c.customOrdererCfg, true, false
	}
	oConfig := fab.OrdererConfig{
		URL: "example.com",
	}

	return &oConfig, true, false
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
func (c *MockConfig) NetworkConfig() *fab.NetworkConfig {
	return nil
}

// ChannelConfig returns the channel configuration
func (c *MockConfig) ChannelConfig(channelID string) *fab.ChannelEndpointConfig {
	if c.chConfig != nil {
		config, ok := c.chConfig[channelID]
		if ok {
			return config
		}
	}

	return &fab.ChannelEndpointConfig{
		Policies: fab.ChannelPolicies{
			QueryChannelConfig: fab.QueryChannelConfigPolicy{},
			Discovery:          fab.DiscoveryPolicy{},
			Selection:          fab.SelectionPolicy{},
			EventService:       fab.EventServicePolicy{},
		},
	}
}

// SetCustomChannelConfig sets the config for the given channel
func (c *MockConfig) SetCustomChannelConfig(channelID string, config *fab.ChannelEndpointConfig) {
	if c.chConfig == nil {
		c.chConfig = make(map[string]*fab.ChannelEndpointConfig)
	}
	c.chConfig[channelID] = config
}

// ChannelPeers returns the channel peers configuration
func (c *MockConfig) ChannelPeers(name string) []fab.ChannelPeer {

	if name == "noChannelPeers" {
		return nil
	}

	peerChCfg := fab.PeerChannelConfig{EndorsingPeer: true, ChaincodeQuery: true, LedgerQuery: true, EventSource: true}
	if name == "noEndpoints" {
		peerChCfg = fab.PeerChannelConfig{EndorsingPeer: false, ChaincodeQuery: false, LedgerQuery: false, EventSource: false}
	}

	mockPeer := fab.ChannelPeer{PeerChannelConfig: peerChCfg, NetworkPeer: fab.NetworkPeer{PeerConfig: fab.PeerConfig{URL: "example.com"}}}
	return []fab.ChannelPeer{mockPeer}
}

// ChannelOrderers returns a list of channel orderers
func (c *MockConfig) ChannelOrderers(name string) []fab.OrdererConfig {
	if name == "Invalid" {
		return nil
	}

	oConfig, _, _ := c.OrdererConfig("")

	return []fab.OrdererConfig{*oConfig}
}

// NetworkPeers returns the mock network peers configuration
func (c *MockConfig) NetworkPeers() []fab.NetworkPeer {
	return c.customNetworkPeerCfg
}

// SecurityProvider ...
func (c *MockConfig) SecurityProvider() string {
	return "sw"
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
func (c *MockConfig) TLSClientCerts() []tls.Certificate {
	return nil
}

// Lookup gets the Value from config file by Key
func (c *MockConfig) Lookup(key string) (interface{}, bool) {
	if key == "invalid" {
		return nil, false
	}
	value, ok := c.Lookup(key)
	if !ok {
		return nil, false
	}
	return value, true
}

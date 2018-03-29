/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var endpointConfig *EndpointConfig
var cryptoConfig *CryptoSuiteConfig
var identityConfig *IdentityConfig

const (
	org0                            = "org0"
	org1                            = "org1"
	configTestFilePath              = "testdata/config_test.yaml"
	configTestTemplateFilePath      = "testdata/config_test_template.yaml"
	configEmptyTestFilePath         = "testdata/empty.yaml"
	configPemTestFilePath           = "testdata/config_test_pem.yaml"
	configEmbeddedUsersTestFilePath = "testdata/config_test_embedded_pems.yaml"
	configType                      = "yaml"
	defaultConfigPath               = "testdata/template"
)

func TestCAConfig(t *testing.T) {
	//Test config
	vConfig := viper.New()
	vConfig.SetConfigFile(configTestFilePath)
	vConfig.ReadInConfig()
	vc := vConfig.ConfigFileUsed()

	if vc == "" {
		t.Fatalf("Failed to load config file")
	}

	//Test network config version
	if vConfig.GetString("version") != "1.0.0" {
		t.Fatalf("Incorrect network version")
	}

	//Test client organization
	if vConfig.GetString("client.organization") != org1 {
		t.Fatalf("Incorrect Client organization")
	}

	//Test Crypto config path
	crossCheckWithViperConfig(endpointConfig.backend.getString("client.cryptoconfig.path"), endpointConfig.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing CA Client File Location
	certfile, err := identityConfig.CAClientCertPath(org1)

	if certfile == "" || err != nil {
		t.Fatalf("CA Cert file location read failed %s", err)
	}

	//Testing CA Key File Location
	keyFile, err := identityConfig.CAClientKeyPath(org1)

	if keyFile == "" || err != nil {
		t.Fatal("CA Key file location read failed")
	}

	//Testing CA Server Cert Files
	sCertFiles, err := identityConfig.CAServerCertPaths(org1)

	if sCertFiles == nil || len(sCertFiles) == 0 || err != nil {
		t.Fatal("Getting CA server cert files failed")
	}

	//Testing MSPID
	mspID, err := endpointConfig.MSPID(org1)
	if mspID != "Org1MSP" || err != nil {
		t.Fatal("Get MSP ID failed")
	}

	//Testing CAConfig
	caConfig, err := identityConfig.CAConfig(org1)
	if caConfig == nil || err != nil {
		t.Fatal("Get CA Config failed")
	}

	// Test User Store Path
	if vConfig.GetString("client.credentialStore.path") != identityConfig.CredentialStorePath() {
		t.Fatalf("Incorrect User Store path")
	}

	// Test CA KeyStore Path
	if vConfig.GetString("client.credentialStore.cryptoStore.path") != identityConfig.CAKeyStorePath() {
		t.Fatalf("Incorrect CA keystore path")
	}

	// Test KeyStore Path
	if path.Join(vConfig.GetString("client.credentialStore.cryptoStore.path"), "keystore") != cryptoConfig.KeyStorePath() {
		t.Fatalf("Incorrect keystore path ")
	}

	// Test BCCSP security is enabled
	if vConfig.GetBool("client.BCCSP.security.enabled") != cryptoConfig.IsSecurityEnabled() {
		t.Fatalf("Incorrect BCCSP Security enabled flag")
	}

	// Test SecurityAlgorithm
	if vConfig.GetString("client.BCCSP.security.hashAlgorithm") != cryptoConfig.SecurityAlgorithm() {
		t.Fatalf("Incorrect BCCSP Security Hash algorithm")
	}

	// Test Security Level
	if vConfig.GetInt("client.BCCSP.security.level") != cryptoConfig.SecurityLevel() {
		t.Fatalf("Incorrect BCCSP Security Level")
	}

	// Test SecurityProvider provider
	if vConfig.GetString("client.BCCSP.security.default.provider") != cryptoConfig.SecurityProvider() {
		t.Fatalf("Incorrect BCCSP SecurityProvider provider")
	}

	// Test Ephemeral flag
	if vConfig.GetBool("client.BCCSP.security.ephemeral") != cryptoConfig.Ephemeral() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}

	// Test SoftVerify flag
	if vConfig.GetBool("client.BCCSP.security.softVerify") != cryptoConfig.SoftVerify() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}

	// Test SecurityProviderPin
	if vConfig.GetString("client.BCCSP.security.pin") != cryptoConfig.SecurityProviderPin() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}

	// Test SecurityProviderPin
	if vConfig.GetString("client.BCCSP.security.label") != cryptoConfig.SecurityProviderLabel() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}

	// test Client
	c, err := identityConfig.Client()
	if err != nil {
		t.Fatalf("Received error when fetching Client info, error is %s", err)
	}
	if c == nil {
		t.Fatal("Received empty client when fetching Client info")
	}

	// testing empty OrgMSP
	mspID, err = endpointConfig.MSPID("dummyorg1")
	if err == nil {
		t.Fatal("Get MSP ID did not fail for dummyorg1")
	}
}

func TestCAConfigFailsByNetworkConfig(t *testing.T) {

	//Tamper 'client.network' value and use a new config to avoid conflicting with other tests

	configBackend, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config: %v", err)
	}

	_, endpointCfg, identityCfg, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatalf("Unexpected error reading config: %v", err)
	}

	sampleEndpointConfig := endpointCfg.(*EndpointConfig)
	sampleEndpointConfig.networkConfigCached = false

	sampleEndpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client", "INVALID")
	sampleEndpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("peers", "INVALID")
	sampleEndpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("organizations", "INVALID")
	sampleEndpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("orderers", "INVALID")
	sampleEndpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("channels", "INVALID")

	sampleIdentityConfig := identityCfg.(*IdentityConfig)
	sampleIdentityConfig.endpointConfig = sampleEndpointConfig

	_, err = sampleEndpointConfig.NetworkConfig()
	if err == nil {
		t.Fatal("Network config load supposed to fail")
	}

	//Test CA client cert file failure scenario
	certfile, err := sampleIdentityConfig.CAClientCertPath("peerorg1")
	if certfile != "" || err == nil {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, err := sampleIdentityConfig.CAClientKeyPath("peerorg1")
	if keyFile != "" || err == nil {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	sCertFiles, err := sampleIdentityConfig.CAServerCertPaths("peerorg1")
	if len(sCertFiles) > 0 || err == nil {
		t.Fatal("Getting CA server cert files supposed to fail")
	}

	//Testing MSPID failure scenario
	mspID, err := sampleEndpointConfig.MSPID("peerorg1")
	if mspID != "" || err == nil {
		t.Fatal("Get MSP ID supposed to fail")
	}

	//Testing CAConfig failure scenario
	caConfig, err := sampleIdentityConfig.CAConfig("peerorg1")
	if caConfig != nil || err == nil {
		t.Fatal("Get CA Config supposed to fail")
	}

	//Testing RandomOrdererConfig failure scenario
	oConfig, err := sampleEndpointConfig.RandomOrdererConfig()
	if oConfig != nil || err == nil {
		t.Fatal("Testing get RandomOrdererConfig supposed to fail")
	}

	//Testing RandomOrdererConfig failure scenario
	oConfig, err = sampleEndpointConfig.OrdererConfig("peerorg1")
	if oConfig != nil || err == nil {
		t.Fatal("Testing get OrdererConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfigs, err := sampleEndpointConfig.PeersConfig("peerorg1")
	if pConfigs != nil || err == nil {
		t.Fatal("Testing PeersConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfig, err := sampleEndpointConfig.PeerConfig("peerorg1", "peer1")
	if pConfig != nil || err == nil {
		t.Fatal("Testing PeerConfig supposed to fail")
	}

	//Testing ChannelConfig failure scenario
	chConfig, err := sampleEndpointConfig.ChannelConfig("invalid")
	if chConfig != nil || err == nil {
		t.Fatal("Testing ChannelConfig supposed to fail")
	}

	//Testing ChannelPeers failure scenario
	cpConfigs, err := sampleEndpointConfig.ChannelPeers("invalid")
	if cpConfigs != nil || err == nil {
		t.Fatal("Testing ChannelPeeers supposed to fail")
	}

	//Testing ChannelOrderers failure scenario
	coConfigs, err := sampleEndpointConfig.ChannelOrderers("invalid")
	if coConfigs != nil || err == nil {
		t.Fatal("Testing ChannelOrderers supposed to fail")
	}

	// test empty network objects
	sampleEndpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("organizations", nil)
	_, err = sampleEndpointConfig.NetworkConfig()
	if err == nil {
		t.Fatalf("Organizations were empty, it should return an error")
	}
}

func TestTLSCAConfig(t *testing.T) {
	//Test TLSCA Cert Pool (Positive test case)

	certFile, _ := identityConfig.CAClientCertPath(org1)
	certConfig := endpoint.TLSConfig{Path: certFile}

	cert, err := certConfig.TLSCert()
	if err != nil {
		t.Fatalf("Failed to get TLS CA Cert, reason: %v", err)
	}

	_, err = endpointConfig.TLSCACertPool(cert)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	//Try again with same cert
	_, err = endpointConfig.TLSCACertPool(cert)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	assert.False(t, len(endpointConfig.tlsCerts) > 1, "number of certs in cert list shouldn't accept duplicates")

	//Test TLSCA Cert Pool (Negative test case)

	badCertConfig := endpoint.TLSConfig{Path: "some random invalid path"}

	badCert, err := badCertConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

	_, err = endpointConfig.TLSCACertPool(badCert)

	keyFile, _ := identityConfig.CAClientKeyPath(org1)

	keyConfig := endpoint.TLSConfig{Path: keyFile}

	key, err := keyConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}

	_, err = endpointConfig.TLSCACertPool(key)
}

func TestTLSCAConfigFromPems(t *testing.T) {
	configBackend, err := FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	_, _, c, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatal(err)
	}

	//Test TLSCA Cert Pool (Positive test case)

	certPem, _ := c.CAClientCertPem(org1)
	certConfig := endpoint.TLSConfig{Pem: certPem}

	cert, err := certConfig.TLSCert()

	if err != nil {
		t.Fatalf("TLS CA cert parse failed, reason: %v", err)
	}

	_, err = endpointConfig.TLSCACertPool(cert)

	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}
	//Test TLSCA Cert Pool (Negative test case)

	badCertConfig := endpoint.TLSConfig{Pem: "some random invalid pem"}

	badCert, err := badCertConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert parse was supposed to fail")
	}

	_, err = endpointConfig.TLSCACertPool(badCert)

	keyPem, _ := identityConfig.CAClientKeyPem(org1)

	keyConfig := endpoint.TLSConfig{Pem: keyPem}

	key, err := keyConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}

	_, err = endpointConfig.TLSCACertPool(key)
}

func TestTimeouts(t *testing.T) {
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.peer.timeout.connection", "2s")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.peer.timeout.response", "6s")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.eventService.timeout.connection", "2m")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.eventService.timeout.registrationResponse", "2h")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.orderer.timeout.connection", "2ms")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.global.timeout.query", "7h")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.global.timeout.execute", "8h")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.global.timeout.resmgmt", "118s")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.global.cache.connectionIdle", "1m")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.global.cache.eventServiceIdle", "2m")
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.orderer.timeout.response", "6s")

	t1 := endpointConfig.TimeoutOrDefault(fab.EndorserConnection)
	if t1 != time.Second*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.EventHubConnection)
	if t1 != time.Minute*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.EventReg)
	if t1 != time.Hour*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.Query)
	if t1 != time.Hour*7 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.Execute)
	if t1 != time.Hour*8 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.OrdererConnection)
	if t1 != time.Millisecond*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.OrdererResponse)
	if t1 != time.Second*6 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.ConnectionIdle)
	if t1 != time.Minute*1 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.EventServiceIdle)
	if t1 != time.Minute*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.PeerResponse)
	if t1 != time.Second*6 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = endpointConfig.TimeoutOrDefault(fab.ResMgmt)
	if t1 != time.Second*118 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}

	// Test default
	endpointConfig.backend.coreBackend.(*defConfigBackend).configViper.Set("client.orderer.timeout.connection", "")
	t1 = endpointConfig.TimeoutOrDefault(fab.OrdererConnection)
	if t1 != time.Second*5 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}

}

func TestOrdererConfig(t *testing.T) {
	oConfig, err := endpointConfig.RandomOrdererConfig()

	if oConfig == nil || err != nil {
		t.Fatal("Testing get RandomOrdererConfig failed")
	}

	oConfig, err = endpointConfig.OrdererConfig("invalid")

	if oConfig != nil || err == nil {
		t.Fatal("Testing non-existing OrdererConfig failed")
	}

	orderers, err := endpointConfig.OrderersConfig()
	if err != nil {
		t.Fatal(err)
	}

	if orderers[0].TLSCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TLSCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if len(orderers[0].TLSCACerts.Pem) == 0 {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func TestChannelOrderers(t *testing.T) {
	orderers, err := endpointConfig.ChannelOrderers("mychannel")
	if orderers == nil || err != nil {
		t.Fatal("Testing ChannelOrderers failed")
	}

	if len(orderers) != 1 {
		t.Fatalf("Expecting one channel orderer got %d", len(orderers))
	}

	if orderers[0].TLSCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TLSCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if len(orderers[0].TLSCACerts.Pem) == 0 {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func testCommonConfigPeerByURL(t *testing.T, expectedConfigURL string, fetchedConfigURL string) {
	expectedConfig, err := endpointConfig.peerConfig(expectedConfigURL)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err := endpointConfig.PeerConfigByURL(fetchedConfigURL)

	if fetchedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) || fetchedConfig.TLSCACerts.Pem != expectedConfig.TLSCACerts.Pem {
		t.Fatalf("Expected Config and fetched config differ")
	}

	if fetchedConfig.URL != expectedConfig.URL || fetchedConfig.EventURL != expectedConfig.EventURL || fetchedConfig.GRPCOptions["ssl-target-name-override"] != expectedConfig.GRPCOptions["ssl-target-name-override"] {
		t.Fatalf("Expected Config and fetched config differ")
	}
}

func TestPeerConfigByUrl_directMatching(t *testing.T) {
	testCommonConfigPeerByURL(t, "peer0.org1.example.com", "peer0.org1.example.com:7051")
}

func TestPeerConfigByUrl_entityMatchers(t *testing.T) {
	testCommonConfigPeerByURL(t, "peer0.org1.example.com", "peer1.org1.example.com:7051")
}

func testCommonConfigOrderer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *fab.OrdererConfig, fetchedConfig *fab.OrdererConfig) {

	expectedConfig, err := endpointConfig.OrdererConfig(expectedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err = endpointConfig.OrdererConfig(fetchedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if expectedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}
	if fetchedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) || fetchedConfig.TLSCACerts.Pem != expectedConfig.TLSCACerts.Pem {
		t.Fatalf("Expected Config and fetched config differ")
	}

	return expectedConfig, fetchedConfig
}

func TestOrdererWithSubstitutedConfig_WithADifferentSubstituteUrl(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example2.com")

	if fetchedConfig.URL == "orderer.example2.com:7050" || fetchedConfig.URL == expectedConfig.URL {
		t.Fatalf("Expected Config should have url that is given in urlSubstitutionExp of match pattern")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatalf("Config should have got localhost as its ssl-target-name-override url as per the matched config")
	}
}

func TestOrdererWithSubstitutedConfig_WithEmptySubstituteUrl(t *testing.T) {
	_, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example3.com")

	if fetchedConfig.URL != "orderer.example3.com:7050" {
		t.Fatalf("Fetched Config should have the same url")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "orderer.example3.com" {
		t.Fatalf("Fetched config should have the same ssl-target-name-override as its hostname")
	}
}

func TestOrdererWithSubstitutedConfig_WithSubstituteUrlExpression(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example4.com:7050")

	if fetchedConfig.URL != expectedConfig.URL {
		t.Fatalf("fetched Config url should be same as expected config url as given in the substituteexp in yaml file")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "orderer.example.com" {
		t.Fatalf("Fetched config should have the ssl-target-name-override as per sslTargetOverrideUrlSubstitutionExp in yaml file")
	}
}

func TestPeersConfig(t *testing.T) {
	pc, err := endpointConfig.PeersConfig(org0)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.URL == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventURL == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}

	pc, err = endpointConfig.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.URL == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventURL == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}
}

func TestPeerConfig(t *testing.T) {
	pc, err := endpointConfig.PeerConfig(org1, "peer0.org1.example.com")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pc.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if pc.TLSCACerts.Path != "" {
		if !filepath.IsAbs(pc.TLSCACerts.Path) {
			t.Fatalf("Expected cert path to be absolute")
		}
	} else if len(pc.TLSCACerts.Pem) == 0 {
		t.Fatalf("Peer %s must have at least a TlsCACerts.Path or TlsCACerts.Pem set", "peer0")
	}
	if len(pc.GRPCOptions) == 0 || pc.GRPCOptions["ssl-target-name-override"] != "peer0.org1.example.com" {
		t.Fatalf("Peer %s must have grpcOptions set in config_test.yaml", "peer0")
	}
}

func testCommonConfigPeer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *fab.PeerConfig, fetchedConfig *fab.PeerConfig) {

	expectedConfig, err := endpointConfig.peerConfig(expectedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err = endpointConfig.peerConfig(fetchedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if expectedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}
	if fetchedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if fetchedConfig.TLSCACerts.Path != expectedConfig.TLSCACerts.Path || len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) {
		t.Fatalf("Expected Config and fetched config differ")
	}

	return expectedConfig, fetchedConfig
}

func TestPeerWithSubstitutedConfig_WithADifferentSubstituteUrl(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer3.org1.example5.com")

	if fetchedConfig.URL == "peer3.org1.example5.com:7051" || fetchedConfig.URL == expectedConfig.URL {
		t.Fatalf("Expected Config should have url that is given in urlSubstitutionExp of match pattern")
	}

	if fetchedConfig.EventURL == "peer3.org1.example5.com:7053" || fetchedConfig.EventURL == expectedConfig.EventURL {
		t.Fatalf("Expected Config should have event url that is given in eventUrlSubstitutionExp of match pattern")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatalf("Config should have got localhost as its ssl-target-name-override url as per the matched config")
	}
}

func TestPeerWithSubstitutedConfig_WithEmptySubstituteUrl(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer4.org1.example3.com")

	if fetchedConfig.URL != "peer4.org1.example3.com:7051" {
		t.Fatalf("Fetched Config should have the same url")
	}

	if fetchedConfig.EventURL != "peer4.org1.example3.com:7053" {
		t.Fatalf("Fetched Config should have the same event url")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "peer4.org1.example3.com" {
		t.Fatalf("Fetched config should have the same ssl-target-name-override as its hostname")
	}
}

func TestPeerWithSubstitutedConfig_WithSubstituteUrlExpression(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer5.example4.com:1234")

	if fetchedConfig.URL != "peer5.org1.example.com:1234" {
		t.Fatalf("fetched Config url should change to include org1 as given in the substituteexp in yaml file")
	}

	if fetchedConfig.EventURL != "peer5.org1.example.com:7053" {
		t.Fatalf("fetched Config event url should change to include org1 as given in the eventsubstituteexp in yaml file")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "peer5.org1.example.com" {
		t.Fatalf("Fetched config should have the ssl-target-name-override as per sslTargetOverrideUrlSubstitutionExp in yaml file")
	}
}

func TestPeerWithSubstitutedConfig_WithMultipleMatchings(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org2.example.com", "peer2.example2.com:1234")

	//Both 2nd and 5th entityMatchers match, however we are only taking 2nd one as its the first one to match
	if fetchedConfig.URL == "peer0.org2.example.com:7051" {
		t.Fatalf("fetched Config url should be matched with the first suitable matcher")
	}

	if fetchedConfig.EventURL != "localhost:7053" {
		t.Fatalf("fetched Config event url should have the config from first suitable matcher")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatalf("Fetched config should have the ssl-target-name-override as per first suitable matcher in yaml file")
	}
}

func TestPeerNotInOrgConfig(t *testing.T) {
	_, err := endpointConfig.PeerConfig(org1, "peer1.org0.example.com")
	if err == nil {
		t.Fatalf("Fetching peer config not for an unassigned org should fail")
	}
}

func TestFromRawSuccess(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)

	// test init config from bytes
	_, err = FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromReaderSuccess(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	buf := bytes.NewBuffer(cBytes)

	// test init config from bytes
	_, err = FromReader(buf, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromFileEmptyFilename(t *testing.T) {
	_, err := FromFile("")()
	if err == nil {
		t.Fatalf("Expected error when passing empty string to FromFile")
	}
}

func loadConfigBytesFromFile(t *testing.T, filePath string) ([]byte, error) {
	// read test config file into bytes array
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to read config file. Error: %s", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Failed to read config file stat. Error: %s", err)
	}
	s := fi.Size()
	cBytes := make([]byte, s, s)
	n, err := f.Read(cBytes)
	if err != nil {
		t.Fatalf("Failed to read test config for bytes array testing. Error: %s", err)
	}
	if n == 0 {
		t.Fatalf("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestInitConfigSuccess(t *testing.T) {
	//Test init config
	//...Positive case
	_, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Failed to initialize config. Error: %s", err)
	}
}

func TestInitConfigWithCmdRoot(t *testing.T) {
	TestInitConfigSuccess(t)
	fileLoc := configTestFilePath
	cmdRoot := "fabric_sdk"
	var logger = logging.NewLogger("fabsdk/core")
	logger.Infof("fileLoc is %s", fileLoc)

	logger.Infof("fileLoc right before calling InitConfigWithCmdRoot is %s", fileLoc)

	configBackend, err := FromFile(fileLoc, WithEnvPrefix(cmdRoot))()
	if err != nil {
		t.Fatalf("Failed to initialize config backend with cmd root. Error: %s", err)
	}

	configProvider, _, _, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatalf("Failed to initialize config with cmd root. Error: %s", err)
	}

	config := configProvider.(*CryptoSuiteConfig)

	//Test if Viper is initialized after calling init config
	if config.backend.getString("client.BCCSP.security.hashAlgorithm") != cryptoConfig.SecurityAlgorithm() {
		t.Fatal("Config initialized with incorrect viper configuration")
	}

}

func TestInitConfigPanic(t *testing.T) {

	os.Setenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL", "INVALID")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Init config with cmdroot was supposed to panic")
		} else {
			//Setting it back during panic so as not to fail other tests
			os.Unsetenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL")
		}
	}()

	backend, err := FromFile(configTestFilePath)()
	assert.Nil(t, err, "not supposed to get error")
	FromBackend(backend)()
}

func TestInitConfigInvalidLocation(t *testing.T) {
	//...Negative case
	_, err := FromFile("invalid file location")()
	if err == nil {
		t.Fatalf("Config file initialization is supposed to fail. Error: %s", err)
	}
}

// Test case to create a new viper instance to prevent conflict with existing
// viper instances in applications that use the SDK
func TestMultipleVipers(t *testing.T) {
	viper.SetConfigFile("./test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		t.Log(err.Error())
	}
	testValue1 := viper.GetString("test.testkey")
	// Read initial value from test.yaml
	if testValue1 != "testvalue" {
		t.Fatalf("Expected testValue before config initialization got: %s", testValue1)
	}
	// initialize go sdk
	configBackend, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Log(err.Error())
	}

	configProvider, _, _, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatal(err)
	}

	config := configProvider.(*CryptoSuiteConfig)

	// Make sure initial value is unaffected
	testValue2 := viper.GetString("test.testkey")
	if testValue2 != "testvalue" {
		t.Fatalf("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3 := config.backend.getBool("client.BCCSP.security.softVerify")
	if testValue3 != true {
		t.Fatalf("Expected existing config value to remain unchanged")
	}
}

func TestEnvironmentVariablesDefaultCmdRoot(t *testing.T) {
	testValue := endpointConfig.backend.getString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("FABRIC_SDK_ENV_TEST", "123")
	defer os.Unsetenv("FABRIC_SDK_ENV_TEST")

	if err != nil {
		t.Log(err.Error())
	}

	testValue = endpointConfig.backend.getString("env.test")
	if testValue != "123" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestEnvironmentVariablesSpecificCmdRoot(t *testing.T) {
	testValue := endpointConfig.backend.getString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("TEST_ROOT_ENV_TEST", "456")
	defer os.Unsetenv("TEST_ROOT_ENV_TEST")

	if err != nil {
		t.Log(err.Error())
	}

	configBackend, err := FromFile(configTestFilePath, WithEnvPrefix("test_root"))()
	if err != nil {
		t.Log(err.Error())
	}

	value, _ := configBackend.Lookup("env.test")
	if value != "456" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestNetworkConfig(t *testing.T) {
	conf, err := endpointConfig.NetworkConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(conf.Orderers) == 0 {
		t.Fatal("Expected orderers to be set")
	}
	if len(conf.Organizations) == 0 {
		t.Fatal("Expected atleast one organisation to be set")
	}
	// viper map keys are lowercase
	if len(conf.Organizations[strings.ToLower(org1)].Peers) == 0 {
		t.Fatalf("Expected org %s to be present in network configuration and peers to be set", org1)
	}
}

func TestMain(m *testing.M) {
	setUp(m)
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setUp(m *testing.M) {
	// do any test setup here...
	var err error
	configBackend, err := FromFile(configTestFilePath)()
	if err != nil {
		fmt.Println(err.Error())
	}

	cryptoSuiteCfg, endpointCfg, identityCfg, err := FromBackend(configBackend)()
	if err != nil {
		fmt.Println(err.Error())
	}
	endpointConfig = endpointCfg.(*EndpointConfig)
	cryptoConfig = cryptoSuiteCfg.(*CryptoSuiteConfig)
	identityConfig = identityCfg.(*IdentityConfig)
}

func teardown() {
	// do any teadown activities here ..
	endpointConfig = nil
}

func crossCheckWithViperConfig(expected string, actual string, message string, t *testing.T) {
	expected = pathvar.Subst(expected)
	if actual != expected {
		t.Fatalf(message)
	}
}

func TestSystemCertPoolDisabled(t *testing.T) {

	// get a config file with pool disabled
	configBackend, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	_, configProvider, _, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatal(err)
	}

	certPool, err := configProvider.TLSCACertPool()
	if err != nil {
		t.Fatal("not supposed to get error")
	}
	// cert pool should be empty
	if len(certPool.Subjects()) > 0 {
		t.Fatal("Expecting empty tls cert pool due to disabled system cert pool")
	}
}

func TestSystemCertPoolEnabled(t *testing.T) {

	// get a config file with pool enabled
	configBackend, err := FromFile(configPemTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	_, configProvider, _, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatal(err)
	}

	certPool, err := configProvider.TLSCACertPool()
	if err != nil {
		t.Fatal("not supposed to get error")
	}

	if len(certPool.Subjects()) == 0 {
		t.Fatal("System Cert Pool not loaded even though it is enabled")
	}

	// Org2 'mychannel' peer is missing cert + pem (it should not fail when systemCertPool enabled)
	_, err = configProvider.ChannelPeers("mychannel")
	if err != nil {
		t.Fatalf("Should have skipped verifying ca cert + pem: %s", err)
	}

}

func TestInitConfigFromRawWithPem(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config from bytes
	backend, err := FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	_, epConfig, idConfig, err := FromBackend(backend)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	o, err := epConfig.OrderersConfig()
	if err != nil {
		t.Fatalf("Failed to load orderers from config. Error: %s", err)
	}

	if o == nil || len(o) == 0 {
		t.Fatalf("orderer cannot be nil or empty")
	}

	oPem := `-----BEGIN CERTIFICATE-----
MIICNjCCAdygAwIBAgIRAILSPmMB3BzoLIQGsFxwZr8wCgYIKoZIzj0EAwIwbDEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xFDASBgNVBAoTC2V4YW1wbGUuY29tMRowGAYDVQQDExF0bHNjYS5l
eGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBaMGwxCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRQwEgYDVQQKEwtleGFtcGxlLmNvbTEaMBgGA1UEAxMRdGxzY2EuZXhh
bXBsZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQfgKb4db53odNzdMXn
P5FZTZTFztOO1yLvCHDofSNfTPq/guw+YYk7ZNmhlhj8JHFG6dTybc9Qb/HOh9hh
gYpXo18wXTAOBgNVHQ8BAf8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMB
Af8EBTADAQH/MCkGA1UdDgQiBCBxaEP3nVHQx4r7tC+WO//vrPRM1t86SKN0s6XB
8LWbHTAKBggqhkjOPQQDAgNIADBFAiEA96HXwCsuMr7tti8lpcv1oVnXg0FlTxR/
SQtE5YgdxkUCIHReNWh/pluHTxeGu2jNCH1eh6o2ajSGeeizoapvdJbN
-----END CERTIFICATE-----`
	loadedOPem := strings.TrimSpace(o[0].TLSCACerts.Pem) // viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedOPem != oPem {
		t.Fatalf("Orderer Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", oPem, loadedOPem)
	}

	pc, err := endpointConfig.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pc == nil || len(pc) == 0 {
		t.Fatalf("peers list of %s cannot be nil or empty", org1)
	}
	peer0 := "peer0.org1.example.com"
	p0, err := epConfig.PeerConfig(org1, peer0)
	if err != nil {
		t.Fatalf("Failed to load %s of %s from the config. Error: %s", peer0, org1, err)
	}
	if p0 == nil {
		t.Fatalf("%s of %s cannot be nil", peer0, org1)
	}
	pPem := `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`

	loadedPPem := strings.TrimSpace(p0.TLSCACerts.Pem) // viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedPPem != pPem {
		t.Fatalf("%s Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", peer0, pPem, loadedPPem)
	}

	// get CA Server cert pems (embedded) for org1
	certs, err := idConfig.CAServerCertPems("org1")
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPems from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty PEM certs for CAServerCertPems")
	}

	// get the client cert pem (embedded) for org1
	idConfig.CAClientCertPem("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPem from config. Error: %s", err)
	}

	// get CA Server certs paths for org1
	certs, err = idConfig.CAServerCertPaths("org1")
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPaths from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty cert file paths for CAServerCertPaths")
	}

	// get the client cert path for org1
	idConfig.CAClientCertPath("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPath from config. Error: %s", err)
	}

	// get the client key pem (embedded) for org1
	idConfig.CAClientKeyPem("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPem from config. Error: %s", err)
	}

	// get the client key file path for org1
	idConfig.CAClientKeyPath("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPath from config. Error: %s", err)
	}
}

func TestLoadConfigWithEmbeddedUsersWithPems(t *testing.T) {
	// get a config file with embedded users
	configBackend, err := FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	_, c, _, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatal(err)
	}

	conf, err := c.NetworkConfig()

	if err != nil {
		t.Fatal(err)
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Cert.Pem == "" {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Key.Pem == "" {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Key.Pem != "" {
		t.Fatal("Mistakenly found an embedded key for user NonExistentEmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Cert.Pem != "" {
		t.Fatal("Mistakenly found an embedded cert for user NonExistentEmbeddedUser")
	}
}

func TestLoadConfigWithEmbeddedUsersWithPaths(t *testing.T) {
	// get a config file with embedded users
	configBackend, err := FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	_, c, _, err := FromBackend(configBackend)()
	if err != nil {
		t.Fatal(err)
	}

	conf, err := c.NetworkConfig()

	if err != nil {
		t.Fatal(err)
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Cert.Path == "" {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Key.Path == "" {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Key.Path != "" {
		t.Fatal("Mistakenly found an embedded key for user NonExistentEmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Cert.Path != "" {
		t.Fatal("Mistakenly found an embedded cert for user NonExistentEmbeddedUser")
	}
}

func TestInitConfigFromRawWrongType(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config with empty type
	backend, err := FromRaw(cBytes, "")()
	if err == nil {
		t.Fatalf("Expected error when initializing config with wrong config type but got no error.")
	}

	// test init config with wrong type
	backend, err = FromRaw(cBytes, "json")()
	if err != nil {
		t.Fatalf("Failed to initialize config backend from bytes array. Error: %s", err)
	}

	_, c, _, err := FromBackend(backend)()
	if err != nil {
		t.Fatalf("Failed to initialize config from backend. Error: %s", err)
	}

	o, err := c.OrderersConfig()
	if len(o) > 0 {
		t.Fatalf("Expected to get an empty list of orderers for wrong config type")
	}

	np, err := c.NetworkPeers()
	if len(np) > 0 {
		t.Fatalf("Expected to get an empty list of peers for wrong config type")
	}
}

func TestTLSClientCertsFromFiles(t *testing.T) {
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	certs, err := endpointConfig.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsFromFilesIncorrectPaths(t *testing.T) {
	// incorrect paths to files
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	_, err := endpointConfig.TLSClientCerts()
	if err == nil {
		t.Fatalf("Expected error but got no errors instead")
	}

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("Expected no such file or directory error")
	}
}

func TestTLSClientCertsFromPem(t *testing.T) {
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = ""

	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := endpointConfig.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertFromPemAndKeyFromFile(t *testing.T) {
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem"

	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	certs, err := endpointConfig.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertFromFileAndKeyFromPem(t *testing.T) {
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = ""

	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""

	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := endpointConfig.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsPemBeforeFiles(t *testing.T) {
	// files have incorrect paths, but pems are loaded first
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"

	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := endpointConfig.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsNoCerts(t *testing.T) {
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	certs, err := endpointConfig.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only empty tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if !reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is not equal to empty cert")
	}
}

func TestNewGoodOpt(t *testing.T) {
	_, err := FromFile("../../../test/fixtures/config/config_test.yaml", goodOpt())()
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	if err != nil || len(cBytes) == 0 {
		t.Fatalf("Unexpected error from loadConfigBytesFromFile")
	}

	buf := bytes.NewBuffer(cBytes)

	_, err = FromReader(buf, configType, goodOpt())()
	if err != nil {
		t.Fatalf("Unexpected error from FromReader: %v", err)
	}

	_, err = FromRaw(cBytes, configType, goodOpt())()
	if err != nil {
		t.Fatalf("Unexpected error from FromRaw %v", err)
	}

	err = os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	/*
		_, err = FromDefaultPath(goodOpt())
		if err != nil {
			t.Fatalf("Unexpected error from FromRaw: %v", err)
		}
	*/
}

func goodOpt() Option {
	return func(opts *options) error {
		return nil
	}
}

func TestNewBadOpt(t *testing.T) {
	_, err := FromFile("../../../test/fixtures/config/config_test.yaml", badOpt())()
	if err == nil {
		t.Fatalf("Expected error from FromFile")
	}

	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	if err != nil || len(cBytes) == 0 {
		t.Fatalf("Unexpected error from loadConfigBytesFromFile")
	}

	buf := bytes.NewBuffer(cBytes)

	_, err = FromReader(buf, configType, badOpt())()
	if err == nil {
		t.Fatalf("Expected error from FromReader")
	}

	_, err = FromRaw(cBytes, configType, badOpt())()
	if err == nil {
		t.Fatalf("Expected error from FromRaw")
	}

	err = os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	/*
		_, err = FromDefaultPath(badOpt())
		if err == nil {
			t.Fatalf("Expected error from FromRaw")
		}
	*/
}

func badOpt() Option {
	return func(opts *options) error {
		return errors.New("Bad Opt")
	}
}

func TestConfigBackend_Lookup(t *testing.T) {
	configBackend, err := FromFile(configTestTemplateFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config backend: %v", err)
	}

	value, ok := configBackend.Lookup("name")
	if !ok {
		t.Fatal(err)
	}

	name := value.(string)
	if name != "global-trade-network" {
		t.Fatal("Expected Name to be global-trade-network")
	}

	value, ok = configBackend.Lookup("description")
	if !ok {
		t.Fatal(err)
	}
	description := value.(string)
	if description == "" {
		t.Fatal("Expected non empty description")
	}

	value, ok = configBackend.Lookup("x-type")
	if !ok {
		t.Fatal(err)
	}
	xType := value.(string)
	if xType != "h1fv1" {
		t.Fatal("Expected x-type to be h1fv1")
	}

	value, ok = configBackend.Lookup("channels.mychannel.chaincodes")
	if !ok {
		t.Fatal(err)
	}
	chaincodes := value.([]interface{})
	if len(chaincodes) != 2 {
		t.Fatal("Expected only 2 chaincodes")
	}

}

/*
func TestDefaultConfigFromFile(t *testing.T) {
	c, err := FromFile(configEmptyTestFilePath, WithTemplatePath(defaultConfigPath))

	if err != nil {
		t.Fatalf("Unexpected error from FromFile: %s", err)
	}

	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	if n.Description != "hello" {
		t.Fatalf("Incorrect Network name from default config. Got %s", n.Description)
	}
}

func TestDefaultConfigFromRaw(t *testing.T) {
	cBytes, err := loadConfigBytesFromFile(t, configEmptyTestFilePath)
	c, err := FromRaw(cBytes, configType, WithTemplatePath(defaultConfigPath))

	if err != nil {
		t.Fatalf("Unexpected error from FromFile: %s", err)
	}

	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	if n.Description != "hello" {
		t.Fatalf("Incorrect Network name from default config. Got %s", n.Description)
	}
}
*/

/*
func TestFromDefaultPathSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathCustomPrefixSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK2_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK2_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath(WithEnvPrefix("FABRIC_SDK2"))
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathCustomPathSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK2_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK2_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath(WithTemplatePath(defaultConfigPath))
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathEmptyFailure(t *testing.T) {
	os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err := FromDefaultPath()
	if err == nil {
		t.Fatalf("Expected failure from unset FABRIC_SDK_CONFIG_PATH")
	}
}

func TestFromDefaultPathFailure(t *testing.T) {
	err := os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath+"/bad")
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath()
	if err == nil {
		t.Fatalf("Expected failure from bad default path")
	}
}
*/

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

const (
	org2                         = "org2"
	org1                         = "org1"
	configTestFile               = "config_test.yaml"
	certPath                     = "${FABRIC_SDK_GO_PROJECT_PATH}/pkg/core/config/testdata/certs/client_sdk_go.pem"
	keyPath                      = "${FABRIC_SDK_GO_PROJECT_PATH}/pkg/core/config/testdata/certs/client_sdk_go-key.pem"
	configPemTestFile            = "config_test_pem.yaml"
	configEmbeddedUsersTestFile  = "config_test_embedded_pems.yaml"
	configTestEntityMatchersFile = "config_test_entity_matchers.yaml"
	configType                   = "yaml"
	orgChannelID                 = "orgchannel"
)

var configBackend core.ConfigBackend

func TestMain(m *testing.M) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	cfgBackend, err := config.FromFile(configPath)()
	if err != nil {
		panic(fmt.Sprintf("Unexpected error reading config: %s", err))
	}
	if len(cfgBackend) != 1 {
		panic(fmt.Sprintf("expected 1 backend but got %d", len(cfgBackend)))
	}
	configBackend = cfgBackend[0]
	r := m.Run()
	os.Exit(r)
}

func getCustomBackend() *mocks.MockConfigBackend {
	backendMap := make(map[string]interface{})
	backendMap["client"], _ = configBackend.Lookup("client")
	backendMap["certificateAuthorities"], _ = configBackend.Lookup("certificateAuthorities")
	backendMap["entityMatchers"], _ = configBackend.Lookup("entityMatchers")
	backendMap["peers"], _ = configBackend.Lookup("peers")
	backendMap["organizations"], _ = configBackend.Lookup("organizations")
	backendMap["orderers"], _ = configBackend.Lookup("orderers")
	backendMap["channels"], _ = configBackend.Lookup("channels")
	return &mocks.MockConfigBackend{KeyValueMap: backendMap}
}

func TestCAConfigFailsByNetworkConfig(t *testing.T) {

	customBackend := getCustomBackend()
	customBackend.KeyValueMap["client"], _ = configBackend.Lookup("client")
	customBackend.KeyValueMap["certificateAuthorities"], _ = configBackend.Lookup("certificateAuthorities")
	customBackend.KeyValueMap["entityMatchers"], _ = configBackend.Lookup("entityMatchers")
	customBackend.KeyValueMap["peers"], _ = configBackend.Lookup("peers")
	customBackend.KeyValueMap["organizations"], _ = configBackend.Lookup("organizations")
	customBackend.KeyValueMap["orderers"], _ = configBackend.Lookup("orderers")
	customBackend.KeyValueMap["channels"], _ = configBackend.Lookup("channels")

	endpointCfg, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatalf("Unexpected error initializing endpoint config: %s", err)
	}

	sampleEndpointConfig := endpointCfg.(*EndpointConfig)
	customBackend.KeyValueMap["channels"] = "INVALID"
	err = sampleEndpointConfig.ResetNetworkConfig()
	assert.NotNil(t, err)

	netConfig := sampleEndpointConfig.NetworkConfig()
	if netConfig != nil {
		t.Fatal("Network config load supposed to fail")
	}

	//Testing MSPID failure scenario
	mspID, ok := comm.MSPID(sampleEndpointConfig, "peerorg1")
	if mspID != "" || ok {
		t.Fatal("Get MSP ID supposed to fail")
	}

	customBackend.KeyValueMap["channels"], _ = configBackend.Lookup("channels")
	err = sampleEndpointConfig.ResetNetworkConfig()
	if err != nil {
		t.Fatalf("failed to reset network config, cause:%s", err)
	}
	//Testing OrdererConfig failure scenario
	oConfig, ok, ignoreOrderer := sampleEndpointConfig.OrdererConfig("peerorg1")
	if oConfig != nil || ok || ignoreOrderer {
		t.Fatal("Testing get OrdererConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfigs, ok := sampleEndpointConfig.PeersConfig("peerorg1")
	if pConfigs != nil || ok {
		t.Fatal("Testing PeersConfig supposed to fail")
	}

	checkCAConfigFailsByNetworkConfig(sampleEndpointConfig, t)
}

func checkCAConfigFailsByNetworkConfig(sampleEndpointConfig *EndpointConfig, t *testing.T) {
	//Testing ChannelPeers failure scenario
	cpConfigs := sampleEndpointConfig.ChannelPeers("invalid")
	if len(cpConfigs) > 0 {
		t.Fatal("Testing ChannelPeeers supposed to fail")
	}
	//Testing ChannelOrderers failure scenario
	coConfigs := sampleEndpointConfig.ChannelOrderers("invalid")
	if len(coConfigs) > 0 {
		t.Fatal("Testing ChannelOrderers supposed to fail")
	}
}

func TestTimeouts(t *testing.T) {
	customBackend := getCustomBackend()
	customBackend.KeyValueMap["client.peer.timeout.connection"] = "12s"
	customBackend.KeyValueMap["client.peer.timeout.response"] = "6s"
	customBackend.KeyValueMap["client.peer.timeout.discovery.greylistExpiry"] = "5m"
	customBackend.KeyValueMap["client.eventService.timeout.connection"] = "2m"
	customBackend.KeyValueMap["client.eventService.timeout.registrationResponse"] = "2h"
	customBackend.KeyValueMap["client.orderer.timeout.connection"] = "2ms"
	customBackend.KeyValueMap["client.orderer.timeout.response"] = "6s"
	customBackend.KeyValueMap["client.discovery.timeout.connection"] = "20s"
	customBackend.KeyValueMap["client.discovery.timeout.response"] = "20s"
	customBackend.KeyValueMap["client.global.timeout.query"] = "7h"
	customBackend.KeyValueMap["client.global.timeout.execute"] = "8h"
	customBackend.KeyValueMap["client.global.timeout.resmgmt"] = "118s"
	customBackend.KeyValueMap["client.global.cache.connectionIdle"] = "1m"
	customBackend.KeyValueMap["client.global.cache.eventServiceIdle"] = "2m"
	customBackend.KeyValueMap["client.global.cache.channelConfig"] = "3m"
	customBackend.KeyValueMap["client.global.cache.channelMembership"] = "4m"
	customBackend.KeyValueMap["client.global.cache.discovery"] = "15s"
	customBackend.KeyValueMap["client.global.cache.selection"] = "15m"

	endpointConfig, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	errStr := "%s timeout not read correctly. Got: %s"
	t1 := endpointConfig.Timeout(fab.PeerConnection)
	if t1 != time.Second*12 {
		t.Fatalf(errStr, "PeerConnection", t1)
	}
	t1 = endpointConfig.Timeout(fab.PeerResponse)
	if t1 != time.Second*6 {
		t.Fatalf(errStr, "PeerResponse", t1)
	}
	t1 = endpointConfig.Timeout(fab.DiscoveryGreylistExpiry)
	if t1 != time.Minute*5 {
		t.Fatalf(errStr, "DiscoveryGreylistExpiry", t1)
	}
	t1 = endpointConfig.Timeout(fab.EventReg)
	if t1 != time.Hour*2 {
		t.Fatalf(errStr, "EventReg", t1)
	}
	t1 = endpointConfig.Timeout(fab.OrdererConnection)
	if t1 != time.Millisecond*2 {
		t.Fatalf(errStr, "OrdererConnection", t1)
	}
	checkTimeouts(endpointConfig, t, errStr)
}

func TestEventServiceConfig(t *testing.T) {
	customBackend := getCustomBackend()
	customBackend.KeyValueMap["client.eventService.type"] = "deliver"
	customBackend.KeyValueMap["client.eventService.blockHeightLagThreshold"] = "4"
	customBackend.KeyValueMap["client.eventService.reconnectBlockHeightLagThreshold"] = "7"
	customBackend.KeyValueMap["client.eventService.peerMonitorPeriod"] = "7s"
	customBackend.KeyValueMap["client.eventService.resolverStrategy"] = "Balanced"
	customBackend.KeyValueMap["client.eventService.balancer"] = "RoundRobin"
}

func checkTimeouts(endpointConfig fab.EndpointConfig, t *testing.T, errStr string) {
	t1 := endpointConfig.Timeout(fab.OrdererResponse)
	assert.Equal(t, time.Second*6, t1, "OrdererResponse")
	t1 = endpointConfig.Timeout(fab.Query)
	assert.Equal(t, time.Hour*7, t1, "Query")
	t1 = endpointConfig.Timeout(fab.Execute)
	assert.Equal(t, time.Hour*8, t1, "Execute")
	t1 = endpointConfig.Timeout(fab.ResMgmt)
	assert.Equal(t, time.Second*118, t1, "ResMgmt")
	t1 = endpointConfig.Timeout(fab.ConnectionIdle)
	assert.Equal(t, time.Minute, t1, "ConnectionIdle")
	t1 = endpointConfig.Timeout(fab.EventServiceIdle)
	assert.Equal(t, time.Minute*2, t1, "EventServiceIdle")
	t1 = endpointConfig.Timeout(fab.ChannelConfigRefresh)
	assert.Equal(t, time.Minute*3, t1, "ChannelConfigRefresh")
	t1 = endpointConfig.Timeout(fab.ChannelMembershipRefresh)
	assert.Equal(t, time.Minute*4, t1, "ChannelMembershipRefresh")
	t1 = endpointConfig.Timeout(fab.DiscoveryServiceRefresh)
	assert.Equal(t, time.Second*15, t1, "DiscoveryServiceRefresh")
	t1 = endpointConfig.Timeout(fab.SelectionServiceRefresh)
	assert.Equal(t, time.Minute*15, t1, "SelectionServiceRefresh")
	t1 = endpointConfig.Timeout(fab.DiscoveryConnection)
	assert.Equal(t, time.Second*20, t1, "DiscoveryConnection")
	t1 = endpointConfig.Timeout(fab.DiscoveryResponse)
	assert.Equal(t, time.Second*20, t1, "DiscoveryResponse")
}

func TestDefaultTimeouts(t *testing.T) {
	customBackend := getCustomBackend()
	customBackend.KeyValueMap["client.peer.timeout.connection"] = ""
	customBackend.KeyValueMap["client.peer.timeout.response"] = ""
	customBackend.KeyValueMap["client.peer.timeout.discovery.greylistExpiry"] = ""
	customBackend.KeyValueMap["client.eventService.timeout.connection"] = ""
	customBackend.KeyValueMap["client.eventService.timeout.registrationResponse"] = ""
	customBackend.KeyValueMap["client.orderer.timeout.connection"] = ""
	customBackend.KeyValueMap["client.orderer.timeout.response"] = ""
	customBackend.KeyValueMap["client.global.timeout.query"] = ""
	customBackend.KeyValueMap["client.global.timeout.execute"] = ""
	customBackend.KeyValueMap["client.global.timeout.resmgmt"] = ""
	customBackend.KeyValueMap["client.global.cache.connectionIdle"] = ""
	customBackend.KeyValueMap["client.global.cache.eventServiceIdle"] = ""
	customBackend.KeyValueMap["client.global.cache.channelConfig"] = ""
	customBackend.KeyValueMap["client.global.cache.channelMembership"] = ""
	customBackend.KeyValueMap["client.global.cache.discovery"] = ""
	customBackend.KeyValueMap["client.global.cache.selection"] = ""

	endpointConfig, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	errStr := "%s default timeout not read correctly. Got: %s"
	t1 := endpointConfig.Timeout(fab.PeerConnection)
	if t1 != defaultPeerConnectionTimeout {
		t.Fatalf(errStr, "PeerConnection", t1)
	}
	t1 = endpointConfig.Timeout(fab.PeerResponse)
	if t1 != defaultPeerResponseTimeout {
		t.Fatalf(errStr, "PeerResponse", t1)
	}
	t1 = endpointConfig.Timeout(fab.DiscoveryGreylistExpiry)
	if t1 != defaultDiscoveryGreylistExpiryTimeout {
		t.Fatalf(errStr, "DiscoveryGreylistExpiry", t1)
	}
	t1 = endpointConfig.Timeout(fab.EventReg)
	if t1 != defaultEventRegTimeout {
		t.Fatalf(errStr, "EventReg", t1)
	}
	t1 = endpointConfig.Timeout(fab.OrdererConnection)
	if t1 != defaultOrdererConnectionTimeout {
		t.Fatalf(errStr, "OrdererConnection", t1)
	}
	t1 = endpointConfig.Timeout(fab.DiscoveryServiceRefresh)
	if t1 != defaultDiscoveryRefreshInterval {
		t.Fatalf(errStr, "DiscoveryRefreshInterval", t1)
	}
	t1 = endpointConfig.Timeout(fab.SelectionServiceRefresh)
	if t1 != defaultSelectionRefreshInterval {
		t.Fatalf(errStr, "SelectionRefreshInterval", t1)
	}
	checkDefaultTimeout(endpointConfig, t, errStr)
}

func checkDefaultTimeout(endpointConfig fab.EndpointConfig, t *testing.T, errStr string) {
	t1 := endpointConfig.Timeout(fab.OrdererResponse)
	if t1 != defaultOrdererResponseTimeout {
		t.Fatalf(errStr, "OrdererResponse", t1)
	}
	t1 = endpointConfig.Timeout(fab.Query)
	if t1 != defaultQueryTimeout {
		t.Fatalf(errStr, "Query", t1)
	}
	t1 = endpointConfig.Timeout(fab.Execute)
	if t1 != defaultExecuteTimeout {
		t.Fatalf(errStr, "Execute", t1)
	}
	t1 = endpointConfig.Timeout(fab.ResMgmt)
	if t1 != defaultResMgmtTimeout {
		t.Fatalf(errStr, "ResMgmt", t1)
	}
	t1 = endpointConfig.Timeout(fab.ConnectionIdle)
	if t1 != defaultConnIdleInterval {
		t.Fatalf(errStr, "ConnectionIdle", t1)
	}
	t1 = endpointConfig.Timeout(fab.EventServiceIdle)
	if t1 != defaultEventServiceIdleInterval {
		t.Fatalf(errStr, "EventServiceIdle", t1)
	}
	t1 = endpointConfig.Timeout(fab.ChannelConfigRefresh)
	if t1 != defaultChannelConfigRefreshInterval {
		t.Fatalf(errStr, "ChannelConfigRefresh", t1)
	}
	t1 = endpointConfig.Timeout(fab.ChannelMembershipRefresh)
	if t1 != defaultChannelMemshpRefreshInterval {
		t.Fatalf(errStr, "ChannelMembershipRefresh", t1)
	}
}

func TestOrdererConfig(t *testing.T) {
	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	oConfig, _, _ := endpointConfig.OrdererConfig("invalid")
	if oConfig != nil {
		t.Fatal("Testing non-existing OrdererConfig failed")
	}

	orderers := endpointConfig.OrderersConfig()
	if orderers[0].TLSCACert == nil {
		t.Fatalf("Orderer %+v must have TLS CA Cert", orderers[0])
	}
}

func TestChannelOrderers(t *testing.T) {
	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	orderers := endpointConfig.ChannelOrderers("mychannel")
	if len(orderers) != 1 {
		t.Fatalf("Expecting one channel orderer got %d", len(orderers))
	}

	if orderers[0].TLSCACert == nil {
		t.Fatalf("Orderer %v must have TLS CA CERT", orderers[0])
	}
}

func TestPeerConfigByUrl_directMatching(t *testing.T) {
	testCommonConfigPeerByURL(t, "peer0.org1.example.com", "peer0.org1.example.com:7051")
}

func TestPeerConfigByUrl_entityMatchers(t *testing.T) {
	testCommonConfigPeerByURL(t, "peer0.org1.example.com", "peer1.org1.example.com:7051")
}

func testCommonConfigPeerByURL(t *testing.T, expectedConfigURL string, fetchedConfigURL string) {
	config1, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	endpointConfig := config1.(*EndpointConfig)

	expectedConfig, ok := endpointConfig.PeerConfig(expectedConfigURL)
	assert.True(t, ok, "getting peerconfig supposed to be successful")

	fetchedConfig, ok := endpointConfig.PeerConfig(fetchedConfigURL)
	assert.True(t, ok, "getting peerconfig supposed to be successful")

	if fetchedConfig.URL == "" {
		t.Fatal("Url value for the host is empty")
	}

	if len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) || fetchedConfig.TLSCACert != expectedConfig.TLSCACert {
		t.Fatal("Expected Config and fetched config differ")
	}

	if fetchedConfig.URL != expectedConfig.URL || fetchedConfig.GRPCOptions["ssl-target-name-override"] != expectedConfig.GRPCOptions["ssl-target-name-override"] {
		t.Fatal("Expected Config and fetched config differ")
	}
}

func testCommonConfigOrderer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *fab.OrdererConfig, fetchedConfig *fab.OrdererConfig) {

	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	expectedConfig, ok, ignoreOrderer := endpointConfig.OrdererConfig(expectedConfigHost)
	assert.True(t, ok)
	assert.False(t, ignoreOrderer)

	fetchedConfig, ok, ignoreOrderer = endpointConfig.OrdererConfig(fetchedConfigHost)
	assert.True(t, ok)
	assert.False(t, ignoreOrderer)

	if expectedConfig.URL == "" {
		t.Fatal("Url value for the host is empty")
	}
	if fetchedConfig.URL == "" {
		t.Fatal("Url value for the host is empty")
	}

	if len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) || fetchedConfig.TLSCACert != expectedConfig.TLSCACert {
		t.Fatal("Expected Config and fetched config differ")
	}

	return expectedConfig, fetchedConfig
}

func TestOrdererWithSubstitutedConfig_WithADifferentSubstituteUrl(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example2.com")

	if fetchedConfig.URL == "orderer.example2.com:7050" || fetchedConfig.URL == expectedConfig.URL {
		t.Fatal("Expected Config should have url that is given in urlSubstitutionExp of match pattern")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatal("Config should have got localhost as its ssl-target-name-override url as per the matched config")
	}
}

func TestOrdererWithSubstitutedConfig_WithEmptySubstituteUrl(t *testing.T) {
	_, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example3.com")

	if fetchedConfig.URL != "orderer.example.com:7050" {
		t.Fatal("Fetched Config should have the same url")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "orderer.example.com" {
		t.Fatal("Fetched config should have the same ssl-target-name-override as its hostname")
	}
}

func TestOrdererWithSubstitutedConfig_WithSubstituteUrlExpression(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example4.com:7050")

	if fetchedConfig.URL != expectedConfig.URL {
		t.Fatal("fetched Config url should be same as expected config url as given in the substituteexp in yaml file")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "orderer.example.com" {
		t.Fatal("Fetched config should have the ssl-target-name-override as per sslTargetOverrideUrlSubstitutionExp in yaml file")
	}
}

func TestChannelConfigName_directMatching(t *testing.T) {
	testMatchingConfigChannel(t, "ch1", "sampleachannel")
	testMatchingConfigChannel(t, "ch1", "samplebchannel")
	testMatchingConfigChannel(t, "ch1", "samplecchannel")
	testMatchingConfigChannel(t, "ch1", "samplechannel")

	testNonMatchingConfigChannel(t, "ch1", "mychannel")
	testNonMatchingConfigChannel(t, "ch1", "orgchannel")
}

func testMatchingConfigChannel(t *testing.T, expectedConfigName string, fetchedConfigName string) (expectedConfig *fab.ChannelEndpointConfig, fetchedConfig *fab.ChannelEndpointConfig) {
	e, f := testCommonConfigChannel(t, expectedConfigName, fetchedConfigName)
	if !deepEquals(e, f) {
		t.Fatalf("'expectedConfig' should be the same as 'fetchedConfig' for %s and %s.", expectedConfigName, fetchedConfigName)
	}

	return e, f
}

func testNonMatchingConfigChannel(t *testing.T, expectedConfigName string, fetchedConfigName string) (expectedConfig *fab.ChannelEndpointConfig, fetchedConfig *fab.ChannelEndpointConfig) {
	e, f := testCommonConfigChannel(t, expectedConfigName, fetchedConfigName)
	if deepEquals(e, f) {
		t.Fatalf("'expectedConfig' should be different than 'fetchedConfig' for %s and %s but got same config.", expectedConfigName, fetchedConfigName)
	}

	return e, f
}

func testCommonConfigChannel(t *testing.T, expectedConfigName string, fetchedConfigName string) (expectedConfig *fab.ChannelEndpointConfig, fetchedConfig *fab.ChannelEndpointConfig) {
	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	expectedConfig = endpointConfig.ChannelConfig(expectedConfigName)
	fetchedConfig = endpointConfig.ChannelConfig(fetchedConfigName)

	return
}

func deepEquals(n, n2 interface{}) bool {
	return reflect.DeepEqual(n, n2)
}

func TestPeersConfig(t *testing.T) {

	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	pc, ok := endpointConfig.PeersConfig(org2)
	assert.True(t, ok)

	for _, value := range pc {
		if value.URL == "" {
			t.Fatal("Url value for the host is empty")
		}
	}

	pc, ok = endpointConfig.PeersConfig(org1)
	assert.True(t, ok)

	for _, value := range pc {
		if value.URL == "" {
			t.Fatal("Url value for the host is empty")
		}
	}
}

func testCommonConfigPeer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *fab.PeerConfig, fetchedConfig *fab.PeerConfig) {

	config1, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	endpointConfig := config1.(*EndpointConfig)

	expectedConfig, ok := endpointConfig.PeerConfig(expectedConfigHost)
	assert.True(t, ok, "getting peerconfig supposed to be successful")

	fetchedConfig, ok = endpointConfig.PeerConfig(fetchedConfigHost)
	assert.True(t, ok, "getting peerconfig supposed to be successful")

	if expectedConfig.URL == "" {
		t.Fatal("Url value for the host is empty")
	}
	if fetchedConfig.URL == "" {
		t.Fatal("Url value for the host is empty")
	}

	if fetchedConfig.TLSCACert != expectedConfig.TLSCACert || len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) {
		t.Fatal("Expected Config and fetched config differ")
	}

	return expectedConfig, fetchedConfig
}

func TestPeerWithSubstitutedConfig_WithADifferentSubstituteUrl(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer3.org1.example5.com")

	if fetchedConfig.URL == "peer3.org1.example5.com:7051" || fetchedConfig.URL == expectedConfig.URL {
		t.Fatal("Expected Config should have url that is given in urlSubstitutionExp of match pattern")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatal("Config should have got localhost as its ssl-target-name-override url as per the matched config")
	}
}

func TestPeerWithSubstitutedConfig_WithEmptySubstituteUrl(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer4.org1.example3.com")

	if fetchedConfig.URL != "peer0.org1.example.com:7051" {
		t.Fatal("Fetched Config should have the same url")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "peer0.org1.example.com" {
		t.Fatal("Fetched config should have the same ssl-target-name-override as its hostname")
	}
}

func TestPeerWithSubstitutedConfig_WithSubstituteUrlExpression(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer5.example4.com:1234")

	if fetchedConfig.URL != "peer5.org1.example.com:1234" {
		t.Fatal("fetched Config url should change to include org1 as given in the substituteexp in yaml file")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "peer5.org1.example.com" {
		t.Fatal("Fetched config should have the ssl-target-name-override as per sslTargetOverrideUrlSubstitutionExp in yaml file")
	}
}

func TestPeerWithSubstitutedConfig_WithMultipleMatchings(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org2.example.com", "peer2.example2.com:1234")

	//Both 2nd and 5th entityMatchers match, however we are only taking 2nd one as its the first one to match
	if fetchedConfig.URL == "peer0.org2.example.com:7051" {
		t.Fatal("fetched Config url should be matched with the first suitable matcher")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatal("Fetched config should have the ssl-target-name-override as per first suitable matcher in yaml file")
	}
}

func TestNetworkConfig(t *testing.T) {
	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	conf := endpointConfig.NetworkConfig()
	assert.NotNil(t, conf)

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

func TestSystemCertPoolDisabled(t *testing.T) {

	// get a config file with pool enabled
	customBackend := getCustomBackend()
	customBackend.KeyValueMap["client.tlsCerts.systemCertPool"] = false
	// get a config file with pool disabled
	endpointConfig, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	_, err = endpointConfig.TLSCACertPool().Get()
	if err != nil {
		t.Fatal("not supposed to get error")
	}
}

func TestInitConfigFromRawWithPem(t *testing.T) {
	// get a config byte for testing
	configPath := filepath.Join(getConfigPath(), configPemTestFile)
	cBytes, err := loadConfigBytesFromFile(t, configPath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config from bytes
	backend, err := config.FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	config1, err := ConfigFromBackend(backend...)
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	endpointConfig := config1.(*EndpointConfig)

	o := endpointConfig.OrderersConfig()
	if len(o) == 0 {
		t.Fatal("orderer cannot be nil or empty")
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

	oCert, err := tlsCertByBytes([]byte(oPem))
	if err != nil {
		t.Fatal("failed to cert from pem bytes")
	}

	if !reflect.DeepEqual(oCert.RawSubject, o[0].TLSCACert.RawSubject) {
		t.Fatal("certs supposed to match")
	}

	pc, ok := endpointConfig.PeersConfig(org1)
	if !ok {
		t.Fatal("unexpected error while getting peerConfig")
	}
	if len(pc) == 0 {
		t.Fatalf("peers list of %s cannot be nil or empty", org1)
	}
	checkPem(endpointConfig, t)

}

func checkPem(endpointConfig *EndpointConfig, t *testing.T) {
	peer0 := "peer0.org1.example.com"
	p0, ok := endpointConfig.PeerConfig(peer0)
	if !ok {
		t.Fatalf("Failed to load %s of %s from the config.", peer0, org1)
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

	oCert, err := tlsCertByBytes([]byte(pPem))
	if err != nil {
		t.Fatal("failed to cert from pem bytes")
	}

	if !reflect.DeepEqual(oCert.RawSubject, p0.TLSCACert.RawSubject) {
		t.Fatal("certs supposed to match")
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
	cBytes := make([]byte, s)
	n, err := f.Read(cBytes)
	if err != nil {
		t.Fatalf("Failed to read test config for bytes array testing. Error: %s", err)
	}
	if n == 0 {
		t.Fatal("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestLoadConfigWithEmbeddedUsersWithPems(t *testing.T) {
	// get a config file with embedded users
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	configBackend1, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig, err := ConfigFromBackend(configBackend1...)
	if err != nil {
		t.Fatal(err)
	}

	conf := endpointConfig.NetworkConfig()
	assert.NotNil(t, conf)

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Cert == nil {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Key == nil {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUser")
	}

}

func TestLoadConfigWithEmbeddedUsersWithPaths(t *testing.T) {
	// get a config file with embedded users
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	configBackend1, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig, err := ConfigFromBackend(configBackend1...)
	if err != nil {
		t.Fatal(err)
	}

	conf := endpointConfig.NetworkConfig()

	assert.NotNil(t, conf)

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Cert == nil {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Key == nil {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUserWithPaths")
	}

}

func TestInitConfigFromRawWrongType(t *testing.T) {
	// get a config byte for testing
	configPath := filepath.Join(getConfigPath(), configPemTestFile)
	cBytes, err := loadConfigBytesFromFile(t, configPath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config with empty type
	_, err = config.FromRaw(cBytes, "")()
	if err == nil {
		t.Fatal("Expected error when initializing config with wrong config type but got no error.")
	}

	// test init config with wrong type
	_, err = config.FromRaw(cBytes, "json")()
	if err == nil {
		t.Fatal("FromRaw didn't fail when config type is wrong")
	}

}

func TestTLSClientCertsFromFiles(t *testing.T) {

	clientTLSOverride := endpoint.MutualTLSConfig{}
	clientTLSOverride.Client.Cert.Path = pathvar.Subst(certPath)
	clientTLSOverride.Client.Key.Path = pathvar.Subst(keyPath)
	clientTLSOverride.Client.Cert.Pem = ""
	clientTLSOverride.Client.Key.Pem = ""

	backends, err := overrideClientTLSInBackend(configBackend, &clientTLSOverride)
	assert.Nil(t, err)

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err)

	certs := config.TLSClientCerts()
	assert.Equal(t, 1, len(certs), "Expected only one tls cert struct")

	if reflect.DeepEqual(certs[0], tls.Certificate{}) {
		t.Fatal("Actual cert is empty")
	}
}

func TestTLSClientCertsFromFilesIncorrectPaths(t *testing.T) {

	configEntity := endpointConfigEntity{}
	testlookup := lookup.New(configBackend)
	testlookup.UnmarshalKey("client", &configEntity.Client)

	//Set client tls paths to empty strings
	configEntity.Client.TLSCerts.Client.Cert.Path = filepath.Clean("/pkg/config/testdata/certs/client_sdk_go.pem")
	configEntity.Client.TLSCerts.Client.Key.Path = filepath.Clean("/pkg/config/testdata/certs/client_sdk_go-key.pem")
	configEntity.Client.TLSCerts.Client.Cert.Pem = ""
	configEntity.Client.TLSCerts.Client.Key.Pem = ""

	//Create backend override
	configBackendOverride := &mocks.MockConfigBackend{}
	configBackendOverride.KeyValueMap = make(map[string]interface{})
	configBackendOverride.KeyValueMap["client"] = configEntity.Client

	_, err := ConfigFromBackend(configBackendOverride, configBackend)
	if err == nil || !strings.Contains(err.Error(), "failed to load client key: failed to load pem bytes from path") {
		t.Fatal(err)
	}

}

func TestTLSClientCertsFromPem(t *testing.T) {

	clientTLSOverride := endpoint.MutualTLSConfig{}

	clientTLSOverride.Client.Cert.Path = ""
	clientTLSOverride.Client.Key.Path = ""

	clientTLSOverride.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
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

	clientTLSOverride.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	backends, err := overrideClientTLSInBackend(configBackend, &clientTLSOverride)
	assert.Nil(t, err)

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err)

	certs := config.TLSClientCerts()
	assert.Equal(t, 1, len(certs), "Expected only one tls cert struct")

	if reflect.DeepEqual(certs[0], tls.Certificate{}) {
		t.Fatal("Actual cert is empty")
	}
}

func TestTLSClientCertFromPemAndKeyFromFile(t *testing.T) {

	clientTLSOverride := endpoint.MutualTLSConfig{}

	clientTLSOverride.Client.Cert.Path = ""
	clientTLSOverride.Client.Key.Path = pathvar.Subst(keyPath)

	clientTLSOverride.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
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

	clientTLSOverride.Client.Key.Pem = ""

	backends, err := overrideClientTLSInBackend(configBackend, &clientTLSOverride)
	assert.Nil(t, err)

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err)

	certs := config.TLSClientCerts()
	assert.Equal(t, 1, len(certs), "Expected only one tls cert struct")

	if reflect.DeepEqual(certs[0], tls.Certificate{}) {
		t.Fatal("Actual cert is empty")
	}
}

func TestTLSClientCertFromFileAndKeyFromPem(t *testing.T) {

	clientTLSOverride := endpoint.MutualTLSConfig{}
	clientTLSOverride.Client.Cert.Path = pathvar.Subst(certPath)
	clientTLSOverride.Client.Key.Path = ""
	clientTLSOverride.Client.Cert.Pem = ""
	clientTLSOverride.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	backends, err := overrideClientTLSInBackend(configBackend, &clientTLSOverride)
	assert.Nil(t, err)

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err)

	certs := config.TLSClientCerts()
	assert.Equal(t, 1, len(certs), "Expected only one tls cert struct")

	if reflect.DeepEqual(certs[0], tls.Certificate{}) {
		t.Fatal("Actual cert is empty")
	}
}

func TestTLSClientCertsPemBeforeFiles(t *testing.T) {

	clientTLSOverride := endpoint.MutualTLSConfig{}
	// files have incorrect paths, but pems are loaded first
	clientTLSOverride.Client.Cert.Path = filepath.Clean("/pkg/config/testdata/certs/client_sdk_go.pem")
	clientTLSOverride.Client.Key.Path = filepath.Clean("/pkg/config/testdata/certs/client_sdk_go-key.pem")

	clientTLSOverride.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
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

	clientTLSOverride.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	backends, err := overrideClientTLSInBackend(configBackend, &clientTLSOverride)
	assert.Nil(t, err)

	config, err := ConfigFromBackend(backends...)
	if err != nil {
		t.Fatal(err)
	}

	certs := config.TLSClientCerts()
	if len(certs) != 1 {
		t.Fatal("Expected only one tls cert struct")
	}

	if reflect.DeepEqual(certs[0], tls.Certificate{}) {
		t.Fatal("Actual cert is empty")
	}
}

func TestTLSClientCertsNoCerts(t *testing.T) {

	configEntity := endpointConfigEntity{}
	testlookup := lookup.New(configBackend)
	testlookup.UnmarshalKey("client", &configEntity.Client)

	//Set client tls paths to empty strings
	configEntity.Client.TLSCerts.Client.Cert.Path = ""
	configEntity.Client.TLSCerts.Client.Key.Path = ""
	configEntity.Client.TLSCerts.Client.Cert.Pem = ""
	configEntity.Client.TLSCerts.Client.Key.Pem = ""

	//Create backend override
	configBackendOverride := &mocks.MockConfigBackend{}
	configBackendOverride.KeyValueMap = make(map[string]interface{})
	configBackendOverride.KeyValueMap["client"] = configEntity.Client

	config, err := ConfigFromBackend(configBackendOverride, configBackend)
	if err != nil {
		t.Fatal(err)
	}

	certs := config.TLSClientCerts()
	if len(certs) != 1 {
		t.Fatal("Expected only empty tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if !reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatal("Actual cert is not equal to empty cert")
	}
}

func TestPeerChannelConfig(t *testing.T) {
	//get custom backend and tamper orgchannel values for test
	backend := getCustomBackend()
	tamperPeerChannelConfig(backend)

	//get endpoint config
	config, err := ConfigFromBackend(backend)
	if err != nil {
		t.Fatal(err)
	}

	//get network config
	networkConfig := config.NetworkConfig()
	assert.NotNil(t, networkConfig)
	//Test if channels config are working as expected, with time values parsed properly
	assert.True(t, len(networkConfig.Channels) == 3)

	channelConfig, ok := networkConfig.Channels["mychannel"]
	require.True(t, ok)

	assert.True(t, len(channelConfig.Peers) == 1)

	qccPolicies := channelConfig.Policies.QueryChannelConfig
	assert.True(t, qccPolicies.MinResponses == 1)
	assert.True(t, qccPolicies.MaxTargets == 1)
	assert.True(t, qccPolicies.RetryOpts.MaxBackoff.String() == (5*time.Second).String())
	assert.True(t, qccPolicies.RetryOpts.InitialBackoff.String() == (500*time.Millisecond).String())
	assert.True(t, qccPolicies.RetryOpts.BackoffFactor == 2.0)

	eventPolicies := channelConfig.Policies.EventService
	assert.Equalf(t, fab.MinBlockHeightStrategy, eventPolicies.ResolverStrategy, "Unexpected value for ResolverStrategy")
	assert.Equal(t, fab.RoundRobin, eventPolicies.Balancer, "Unexpected value for Balancer")
	assert.Equal(t, 4, eventPolicies.BlockHeightLagThreshold, "Unexpected value for BlockHeightLagThreshold")
	assert.Equal(t, 8, eventPolicies.ReconnectBlockHeightLagThreshold, "Unexpected value for ReconnectBlockHeightLagThreshold")
	assert.Equal(t, 6*time.Second, eventPolicies.PeerMonitorPeriod, "Unexpected value for PeerMonitorPeriod")

	//Test if custom hook for (default=true) func is working
	assert.True(t, len(networkConfig.Channels[orgChannelID].Peers) == 3)
	//test orgchannel peer1 (EndorsingPeer should be true as set, remaining should be default = true)
	orgChannelPeer1 := networkConfig.Channels[orgChannelID].Peers["peer0.org1.example.com"]
	assert.True(t, orgChannelPeer1.EndorsingPeer)
	assert.True(t, orgChannelPeer1.LedgerQuery)
	assert.True(t, orgChannelPeer1.EventSource)
	assert.True(t, orgChannelPeer1.ChaincodeQuery)

	//test orgchannel peer2 (EndorsingPeer should be false as set, remaining should be default = true)
	orgChannelPeer2 := networkConfig.Channels[orgChannelID].Peers["peer0.org2.example.com"]
	assert.False(t, orgChannelPeer2.EndorsingPeer)
	assert.True(t, orgChannelPeer2.LedgerQuery)
	assert.True(t, orgChannelPeer2.EventSource)
	assert.True(t, orgChannelPeer2.ChaincodeQuery)

	//test orgchannel peer3 (All should be true)
	orgChannelPeer3 := networkConfig.Channels[orgChannelID].Peers["peer0.org3.example.com"]
	assert.True(t, orgChannelPeer3.EndorsingPeer)
	assert.True(t, orgChannelPeer3.LedgerQuery)
	assert.True(t, orgChannelPeer3.EventSource)
	assert.True(t, orgChannelPeer3.ChaincodeQuery)
}

func TestEndpointConfigWithMultipleBackends(t *testing.T) {

	configPath := filepath.Join(getConfigPath(), configTestEntityMatchersFile)
	sampleViper := newViper(configPath)

	var backends []core.ConfigBackend
	backendMap := make(map[string]interface{})
	backendMap["client"] = sampleViper.Get("client")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["channels"] = sampleViper.Get("channels")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["entityMatchers"] = sampleViper.Get("entityMatchers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["organizations"] = sampleViper.Get("organizations")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["orderers"] = sampleViper.Get("orderers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["peers"] = sampleViper.Get("peers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	//create endpointConfig with all 7 backends having 7 different entities
	endpointConfig, err := ConfigFromBackend(backends...)

	assert.Nil(t, err, "ConfigFromBackend should have been successful for multiple backends")
	assert.NotNil(t, endpointConfig, "Invalid endpoint config from multiple backends")

	//Get network Config
	networkConfig := endpointConfig.NetworkConfig()
	assert.NotNil(t, networkConfig, "Invalid networkConfig")

	//Channel
	assert.Equal(t, len(networkConfig.Channels), 5)
	assert.Equal(t, len(networkConfig.Channels["mychannel"].Peers), 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String(), (5 * time.Second).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String(), (500 * time.Millisecond).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor, 2.0)

	//Organizations
	assert.Equal(t, len(networkConfig.Organizations), 4)
	assert.Equal(t, networkConfig.Organizations["org1"].MSPID, "Org1MSP")

	//Orderer
	assert.Equal(t, len(networkConfig.Orderers), 2)
	assert.Equal(t, networkConfig.Orderers["local.orderer.example.com"].URL, "orderer.example.com:7050")
	assert.Equal(t, networkConfig.Orderers["orderer1.example.com"].URL, "orderer1.example.com:7050")

	//Peer
	assert.Equal(t, len(networkConfig.Peers), 3)
	assert.Equal(t, networkConfig.Peers["local.peer0.org1.example.com"].URL, "peer0.org1.example.com:7051")
	assert.Equal(t, networkConfig.Peers["peer0.org3.example.com"].URL, "peer0.org3.example.com:7051")

	//EntityMatchers
	endpointConfigImpl := endpointConfig.(*EndpointConfig)
	assert.Equal(t, len(endpointConfigImpl.entityMatchers.matchers), 4)
	assert.Equal(t, len(endpointConfigImpl.entityMatchers.matchers["peer"]), 10)
	assert.Equal(t, endpointConfigImpl.entityMatchers.matchers["peer"][0].MappedHost, "local.peer0.org1.example.com")
	assert.Equal(t, len(endpointConfigImpl.entityMatchers.matchers["orderer"]), 6)
	assert.Equal(t, endpointConfigImpl.entityMatchers.matchers["orderer"][0].MappedHost, "local.orderer.example.com")
	assert.Equal(t, len(endpointConfigImpl.entityMatchers.matchers["certificateauthority"]), 3)
	assert.Equal(t, endpointConfigImpl.entityMatchers.matchers["certificateauthority"][0].MappedHost, "local.ca.org1.example.com")
	assert.Equal(t, len(endpointConfigImpl.entityMatchers.matchers["channel"]), 1)
	assert.Equal(t, endpointConfigImpl.entityMatchers.matchers["channel"][0].MappedName, "ch1")

}

func TestCAConfig(t *testing.T) {
	//Test config
	configPath := filepath.Join(getConfigPath(), configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	endpointConfig, err := ConfigFromBackend(backend...)
	if err != nil {
		t.Fatal("Failed to get identity config")
	}

	//Test Crypto config path
	val, _ := backend[0].Lookup("client.cryptoconfig.path")
	assert.True(t, pathvar.Subst(val.(string)) == endpointConfig.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing MSPID
	mspID, ok := comm.MSPID(endpointConfig, org1)
	assert.True(t, ok, "Get MSP ID failed")
	assert.True(t, mspID == "Org1MSP", "Get MSP ID failed")

	// testing empty OrgMSP
	_, ok = comm.MSPID(endpointConfig, "dummyorg1")
	assert.False(t, ok, "Get MSP ID did not fail for dummyorg1")

}

func TestNetworkPeers(t *testing.T) {

	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	testNetworkPeers(t, endpointConfig)
}

func TestNetworkPeersWithEntityMatchers(t *testing.T) {

	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}
	testNetworkPeers(t, endpointConfig)

}

func testNetworkPeers(t *testing.T, endpointConfig fab.EndpointConfig) {
	networkPeers := endpointConfig.NetworkPeers()
	assert.NotNil(t, networkPeers, "supposed to get valid network peers")
	assert.Equal(t, 2, len(networkPeers), "supposed to get 2 network peers")
	assert.NotEmpty(t, networkPeers[0].MSPID)
	assert.NotEmpty(t, networkPeers[1].MSPID)
	assert.NotEmpty(t, networkPeers[0].PeerConfig)
	assert.NotEmpty(t, networkPeers[1].PeerConfig)

	//cross check with peer config for org1
	peerConfigOrg1, ok := endpointConfig.PeersConfig("org1")
	assert.True(t, ok, "not suppopsed to get false")

	//cross check with peer config for org2
	peerConfigOrg2, ok := endpointConfig.PeersConfig("org2")
	assert.True(t, ok, "not suppopsed to get false")

	if networkPeers[0].MSPID == "Org1MSP" {
		assert.Equal(t, peerConfigOrg1[0], networkPeers[0].PeerConfig)
		assert.Equal(t, peerConfigOrg2[0], networkPeers[1].PeerConfig)

	} else if networkPeers[0].MSPID == "Org2MSP" {
		assert.Equal(t, peerConfigOrg2[0], networkPeers[0].PeerConfig)
		assert.Equal(t, peerConfigOrg1[0], networkPeers[1].PeerConfig)
	} else {
		t.Fatal("invalid MSPID found")
	}
}

func tamperPeerChannelConfig(backend *mocks.MockConfigBackend) {
	channelsMap := backend.KeyValueMap["channels"]
	orgChannel := map[string]interface{}{
		"orderers": []string{"orderer.example.com"},
		"peers": map[string]interface{}{
			"peer0.org1.example.com": map[string]interface{}{"endorsingpeer": true},
			"peer0.org2.example.com": map[string]interface{}{"endorsingpeer": false},
			"peer0.org3.example.com": nil,
		},
	}
	(channelsMap.(map[string]interface{}))[orgChannelID] = orgChannel
}

func getMatcherConfig() core.ConfigBackend {
	configPath := filepath.Join(getConfigPath(), configTestEntityMatchersFile)
	cfgBackend, err := config.FromFile(configPath)()
	if err != nil {
		panic(fmt.Sprintf("Unexpected error reading config: %s", err))
	}
	if len(cfgBackend) != 1 {
		panic(fmt.Sprintf("expected 1 backend but got %d", len(cfgBackend)))
	}
	return cfgBackend[0]
}

func newViper(path string) *viper.Viper {
	myViper := viper.New()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	myViper.SetConfigFile(path)
	err := myViper.MergeInConfig()
	if err != nil {
		panic(err)
	}
	return myViper
}

func overrideClientTLSInBackend(backend core.ConfigBackend, tlsCerts *endpoint.MutualTLSConfig) ([]core.ConfigBackend, error) {
	endpointEntity := endpointConfigEntity{}
	err := lookup.New(backend).UnmarshalKey("client", &endpointEntity.Client)
	if err != nil {
		return nil, err
	}
	endpointEntity.Client.TLSCerts.Client = tlsCerts.Client

	backendOverride := mocks.MockConfigBackend{}
	backendOverride.KeyValueMap = make(map[string]interface{})
	backendOverride.KeyValueMap["client"] = endpointEntity.Client

	return []core.ConfigBackend{&backendOverride, backend}, nil
}

func tlsCertByBytes(bytes []byte) (*x509.Certificate, error) {

	block, _ := pem.Decode(bytes)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		return pub, nil
	}

	//no cert found and there is no error
	return nil, errors.New("empty byte")
}

func TestEntityMatchers(t *testing.T) {

	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	endpointConfigImpl := endpointConfig.(*EndpointConfig)
	assert.Equal(t, 10, len(endpointConfigImpl.peerMatchers), "preloading matchers isn't working as expected")
	assert.Equal(t, 6, len(endpointConfigImpl.ordererMatchers), "preloading matchers isn't working as expected")
	assert.Equal(t, 1, len(endpointConfigImpl.channelMatchers), "preloading matchers isn't working as expected")

	peerConfig, ok := endpointConfig.PeerConfig("xyz.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.NotNil(t, peerConfig, "supposed to find peer config")

	ordererConfig, ok, ignoreOrderer := endpointConfig.OrdererConfig("xyz.org1.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "supposed to not ignore orderer")
	assert.NotNil(t, ordererConfig, "supposed to find orderer config")

	channelConfig := endpointConfig.ChannelConfig("samplexyzchannel")
	assert.NotNil(t, channelConfig, "supposed to find channel config")
}

func TestDefaultGRPCOpts(t *testing.T) {

	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	peerConfig, ok := endpointConfig.PeerConfig("xyz.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.NotNil(t, peerConfig, "supposed to find peer config")
	assert.NotEmpty(t, peerConfig.GRPCOptions)
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))
	assert.Equal(t, "0s", peerConfig.GRPCOptions["keep-alive-time"])
	assert.Equal(t, "peer0.org1.example.com", peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, "20s", peerConfig.GRPCOptions["keep-alive-timeout"])
	assert.Equal(t, false, peerConfig.GRPCOptions["keep-alive-permit"])
	assert.Equal(t, false, peerConfig.GRPCOptions["fail-fast"])
	assert.Equal(t, false, peerConfig.GRPCOptions["allow-insecure"])

	//make sure map has all the expected grpc opts keys
	_, ok = peerConfig.GRPCOptions["keep-alive-time"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["ssl-target-name-override"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["keep-alive-timeout"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["keep-alive-permit"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["fail-fast"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["allow-insecure"]
	assert.True(t, ok)

	ordererConfig, ok, ignoreOrderer := endpointConfig.OrdererConfig("xyz.org1.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer)
	assert.NotNil(t, ordererConfig, "supposed to find orderer config")
	assert.NotEmpty(t, ordererConfig.GRPCOptions)
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))
	assert.Equal(t, "0s", ordererConfig.GRPCOptions["keep-alive-time"])
	assert.Equal(t, "orderer.example.com", ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, "20s", ordererConfig.GRPCOptions["keep-alive-timeout"])
	assert.Equal(t, false, ordererConfig.GRPCOptions["keep-alive-permit"])
	assert.Equal(t, false, ordererConfig.GRPCOptions["fail-fast"])
	assert.Equal(t, false, ordererConfig.GRPCOptions["allow-insecure"])

	//make sure map has all the expected grpc opts keys
	_, ok = ordererConfig.GRPCOptions["keep-alive-time"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["ssl-target-name-override"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["keep-alive-timeout"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["keep-alive-permit"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["fail-fast"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["allow-insecure"]
	assert.True(t, ok)
}

func TestSetDefault(t *testing.T) {
	dataMap := make(map[string]interface{})
	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	defaultVal1 := true
	defaultVal2 := false
	key3Val := true

	setDefault(dataMap, key1, defaultVal1)
	assert.Equal(t, defaultVal1, dataMap[key1])

	setDefault(dataMap, key2, defaultVal2)
	assert.Equal(t, defaultVal2, dataMap[key2])

	// setDefault makes no effects
	dataMap[key3] = key3Val
	setDefault(dataMap, key3, !key3Val)
	assert.Equal(t, key3Val, dataMap[key3])
}

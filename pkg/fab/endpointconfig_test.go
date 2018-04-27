/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"testing"

	"os"

	"fmt"

	"time"

	"path/filepath"

	"strings"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	org0                             = "org0"
	org1                             = "org1"
	configTestFilePath               = "../core/config/testdata/config_test.yaml"
	certPath                         = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/client/client_fabric_client.pem"
	keyPath                          = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/client/client_fabric_client-key.pem"
	configPemTestFilePath            = "../core/config/testdata/config_test_pem.yaml"
	configEmbeddedUsersTestFilePath  = "../core/config/testdata/config_test_embedded_pems.yaml"
	configTestEntityMatchersFilePath = "../core/config/testdata/config_test_entity_matchers.yaml"
	configType                       = "yaml"
	orgChannelID                     = "orgchannel"
)

var configBackend core.ConfigBackend

func TestMain(m *testing.M) {
	cfgBackend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		panic(fmt.Sprintf("Unexpected error reading config: %v", err))
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
		t.Fatalf("Unexpected error initializing endpoint config: %v", err)
	}

	sampleEndpointConfig := endpointCfg.(*EndpointConfig)
	sampleEndpointConfig.networkConfigCached = false

	customBackend.KeyValueMap["channels"] = "INVALID"
	_, err = sampleEndpointConfig.NetworkConfig()
	if err == nil {
		t.Fatal("Network config load supposed to fail")
	}

	//Testing MSPID failure scenario
	mspID, err := sampleEndpointConfig.MSPID("peerorg1")
	if mspID != "" || err == nil {
		t.Fatal("Get MSP ID supposed to fail")
	}

	//Testing OrdererConfig failure scenario
	oConfig, err := sampleEndpointConfig.OrdererConfig("peerorg1")
	if oConfig != nil || err == nil {
		t.Fatal("Testing get OrdererConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfigs, err := sampleEndpointConfig.PeersConfig("peerorg1")
	if pConfigs != nil || err == nil {
		t.Fatal("Testing PeersConfig supposed to fail")
	}

	checkCAConfigFailsByNetworkConfig(sampleEndpointConfig, t)
}

func checkCAConfigFailsByNetworkConfig(sampleEndpointConfig *EndpointConfig, t *testing.T) {
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

	endpointConfig, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	errStr := "%s timeout not read correctly. Got: %s"
	t1 := endpointConfig.Timeout(fab.EndorserConnection)
	if t1 != time.Second*12 {
		t.Fatalf(errStr, "EndorserConnection", t1)
	}
	t1 = endpointConfig.Timeout(fab.PeerResponse)
	if t1 != time.Second*6 {
		t.Fatalf(errStr, "PeerResponse", t1)
	}
	t1 = endpointConfig.Timeout(fab.DiscoveryGreylistExpiry)
	if t1 != time.Minute*5 {
		t.Fatalf(errStr, "DiscoveryGreylistExpiry", t1)
	}
	t1 = endpointConfig.Timeout(fab.EventHubConnection)
	if t1 != time.Minute*2 {
		t.Fatalf(errStr, "EventHubConnection", t1)
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

	endpointConfig, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	errStr := "%s default timeout not read correctly. Got: %s"
	t1 := endpointConfig.Timeout(fab.EndorserConnection)
	if t1 != defaultEndorserConnectionTimeout {
		t.Fatalf(errStr, "EndorserConnection", t1)
	}
	t1 = endpointConfig.Timeout(fab.PeerResponse)
	if t1 != defaultPeerResponseTimeout {
		t.Fatalf(errStr, "PeerResponse", t1)
	}
	t1 = endpointConfig.Timeout(fab.DiscoveryGreylistExpiry)
	if t1 != defaultDiscoveryGreylistExpiryTimeout {
		t.Fatalf(errStr, "DiscoveryGreylistExpiry", t1)
	}
	t1 = endpointConfig.Timeout(fab.EventHubConnection)
	if t1 != defaultEventHubConnectionTimeout {
		t.Fatalf(errStr, "EventHubConnection", t1)
	}
	t1 = endpointConfig.Timeout(fab.EventReg)
	if t1 != defaultEventRegTimeout {
		t.Fatalf(errStr, "EventReg", t1)
	}
	t1 = endpointConfig.Timeout(fab.OrdererConnection)
	if t1 != defaultOrdererConnectionTimeout {
		t.Fatalf(errStr, "OrdererConnection", t1)
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

	oConfig, _ := endpointConfig.OrdererConfig("invalid")
	if oConfig != nil {
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
	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

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

	expectedConfig, err := endpointConfig.PeerConfig(expectedConfigURL)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err := endpointConfig.PeerConfig(fetchedConfigURL)
	if err != nil {
		t.Fatalf(err.Error())
	}

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

func testCommonConfigOrderer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *fab.OrdererConfig, fetchedConfig *fab.OrdererConfig) {

	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	expectedConfig, err = endpointConfig.OrdererConfig(expectedConfigHost)
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

func TestChannelConfigName_directMatching(t *testing.T) {
	testMatchingConfigChannel(t, "ch1", "sampleachannel")
	testMatchingConfigChannel(t, "ch1", "samplebchannel")
	testMatchingConfigChannel(t, "ch1", "samplecchannel")
	testMatchingConfigChannel(t, "ch1", "samplechannel")

	testNonMatchingConfigChannel(t, "ch1", "mychannel")
	testNonMatchingConfigChannel(t, "ch1", "orgchannel")
}

func testMatchingConfigChannel(t *testing.T, expectedConfigName string, fetchedConfigName string) (expectedConfig *fab.ChannelNetworkConfig, fetchedConfig *fab.ChannelNetworkConfig) {
	e, f := testCommonConfigChannel(t, expectedConfigName, fetchedConfigName)
	if !deepEquals(e, f) {
		t.Fatalf("'expectedConfig' should be the same as 'fetchedConfig' for %s and %s.", expectedConfigName, fetchedConfigName)
	}

	return e, f
}

func testNonMatchingConfigChannel(t *testing.T, expectedConfigName string, fetchedConfigName string) (expectedConfig *fab.ChannelNetworkConfig, fetchedConfig *fab.ChannelNetworkConfig) {
	e, f := testCommonConfigChannel(t, expectedConfigName, fetchedConfigName)
	if deepEquals(e, f) {
		t.Fatalf("'expectedConfig' should be different than 'fetchedConfig' for %s and %s but got same config.", expectedConfigName, fetchedConfigName)
	}

	return e, f
}

func testCommonConfigChannel(t *testing.T, expectedConfigName string, fetchedConfigName string) (expectedConfig *fab.ChannelNetworkConfig, fetchedConfig *fab.ChannelNetworkConfig) {
	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	expectedConfig, err = endpointConfig.ChannelConfig(expectedConfigName)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err = endpointConfig.ChannelConfig(fetchedConfigName)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return expectedConfig, fetchedConfig
}

func deepEquals(n, n2 interface{}) bool {
	return reflect.DeepEqual(n, n2)
}

func TestPeersConfig(t *testing.T) {

	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

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

func testCommonConfigPeer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *fab.PeerConfig, fetchedConfig *fab.PeerConfig) {

	config1, err := ConfigFromBackend(getMatcherConfig())
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	endpointConfig := config1.(*EndpointConfig)

	expectedConfig, err = endpointConfig.PeerConfig(expectedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err = endpointConfig.PeerConfig(fetchedConfigHost)
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

func TestNetworkConfig(t *testing.T) {
	endpointConfig, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

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

func TestSystemCertPoolDisabled(t *testing.T) {

	// get a config file with pool enabled
	customBackend := getCustomBackend()
	customBackend.KeyValueMap["client.tlsCerts.systemCertPool"] = false
	// get a config file with pool disabled
	endpointConfig, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatal("Failed to get endpoint config from backend")
	}

	_, err = endpointConfig.TLSCACertPool()
	if err != nil {
		t.Fatal("not supposed to get error")
	}
}

func TestInitConfigFromRawWithPem(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
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

	o, err := endpointConfig.OrderersConfig()
	if err != nil {
		t.Fatalf("Failed to load orderers from config. Error: %s", err)
	}

	if len(o) == 0 {
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
	if len(pc) == 0 {
		t.Fatalf("peers list of %s cannot be nil or empty", org1)
	}
	checkPem(endpointConfig, t)

}

func checkPem(endpointConfig *EndpointConfig, t *testing.T) {
	peer0 := "peer0.org1.example.com"
	p0, err := endpointConfig.PeerConfig(peer0)
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
	loadedPPem := strings.TrimSpace(p0.TLSCACerts.Pem)
	// viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedPPem != pPem {
		t.Fatalf("%s Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", peer0, pPem, loadedPPem)
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
		t.Fatalf("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestLoadConfigWithEmbeddedUsersWithPems(t *testing.T) {
	// get a config file with embedded users
	configBackend1, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig, err := ConfigFromBackend(configBackend1...)
	if err != nil {
		t.Fatal(err)
	}

	conf, err := endpointConfig.NetworkConfig()

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
	configBackend1, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig, err := ConfigFromBackend(configBackend1...)
	if err != nil {
		t.Fatal(err)
	}

	conf, err := endpointConfig.NetworkConfig()

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
	_, err = config.FromRaw(cBytes, "")()
	if err == nil {
		t.Fatalf("Expected error when initializing config with wrong config type but got no error.")
	}

	// test init config with wrong type
	_, err = config.FromRaw(cBytes, "json")()
	if err == nil {
		t.Fatalf("FromRaw didn't fail when config type is wrong")
	}

}

func TestTLSClientCertsFromFiles(t *testing.T) {
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = pathvar.Subst(certPath)
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = pathvar.Subst(keyPath)
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
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
	// incorrect paths to files
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	_, err = endpointConfig.TLSClientCerts()
	if err == nil {
		t.Fatalf("Expected error but got no errors instead")
	}

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("Expected no such file or directory error")
	}
}

func TestTLSClientCertsFromPem(t *testing.T) {
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
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
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = pathvar.Subst(keyPath)

	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkegAwIBAgIUBzAG7MTjO4n9GFkYTkJBnvCInRIwCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzA3MTkxOTUyMDBaFw0xODA3MTkxOTUyMDBaMGoxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMRYwFAYDVQQDDA1mYWJyaWNfY2xpZW50MHYwEAYH
KoZIzj0CAQYFK4EEACIDYgAEyW+qHu26Zp7icI2DGkF+w9mENLyx5kVirEEp+u+M
UCeTfKzBwAPw17aSDCiObrpaLdIyecRZKYpCxnfPurKEKfKXebZDKmQdGpxaFKbX
aJvC44EbrOq5x218RqnCDeqAo4GKMIGHMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUE
DDAKBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBRBA9pDyeovnjWP
uvftCfEagM/wKjAfBgNVHSMEGDAWgBQUcJ+Hm9wjfMO4jh0E7LBIXBATDzASBgNV
HREECzAJggd0ZXN0aW5nMAoGCCqGSM49BAMEA4GLADCBhwJCATMHAs0T6yZFDByA
XNzhG5LwkITa+GcMJNR9qXlFBG18P+LM/2cdT6Y2+Fz9ZEvGjYMC+c+yg4nyRwu3
rIYog3WBAkECntF217dk3VCZHXfl+rik6wm+ijzYk+k336UERiSJRu09YHHEh7x6
NRCHI3uXUJ5/3zDZM3qtV8UYHou4KDS35Q==
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
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Path = pathvar.Subst(certPath)
	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Path = ""

	endpointConfig.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""

	endpointConfig.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDAeWRhdAl+olgpLiI9mXHwcgJ1g4NNgPrYFSkkukISeAGfvK348izwG
0Aub948H5IygBwYFK4EEACKhZANiAATJb6oe7bpmnuJwjYMaQX7D2YQ0vLHmRWKs
QSn674xQJ5N8rMHAA/DXtpIMKI5uulot0jJ5xFkpikLGd8+6soQp8pd5tkMqZB0a
nFoUptdom8LjgRus6rnHbXxGqcIN6oA=
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
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
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
	config, err := ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig := config.(*EndpointConfig)
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
	networkConfig, err := config.NetworkConfig()
	if err != nil {
		t.Fatal(err)
	}

	//Test if channels config are working as expected, with time values parsed properly
	assert.True(t, len(networkConfig.Channels) == 3)
	assert.True(t, len(networkConfig.Channels["mychannel"].Peers) == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String() == (5*time.Second).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String() == (500*time.Millisecond).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor == 2.0)

	//Test if custom hook for (default=true) func is working
	assert.True(t, len(networkConfig.Channels[orgChannelID].Peers) == 2)
	//test orgchannel peer1 (EndorsingPeer should be true as set, remaining should be default = true)
	orgChannelPeer1 := networkConfig.Channels[orgChannelID].Peers["peer0.org1.example.com"]
	assert.True(t, orgChannelPeer1.EndorsingPeer)
	assert.True(t, orgChannelPeer1.LedgerQuery)
	assert.True(t, orgChannelPeer1.EventSource)
	assert.True(t, orgChannelPeer1.ChaincodeQuery)

	//test orgchannel peer1 (EndorsingPeer should be false as set, remaining should be default = true)
	orgChannelPeer2 := networkConfig.Channels[orgChannelID].Peers["peer0.org2.example.com"]
	assert.False(t, orgChannelPeer2.EndorsingPeer)
	assert.True(t, orgChannelPeer2.LedgerQuery)
	assert.True(t, orgChannelPeer2.EventSource)
	assert.True(t, orgChannelPeer2.ChaincodeQuery)

}

func TestEndpointConfigWithMultipleBackends(t *testing.T) {

	sampleViper := newViper(configTestEntityMatchersFilePath)

	var backends []core.ConfigBackend
	backendMap := make(map[string]interface{})
	backendMap["client"] = sampleViper.Get("client")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["channels"] = sampleViper.Get("channels")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["certificateAuthorities"] = sampleViper.Get("certificateAuthorities")
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
	networkConfig, err := endpointConfig.NetworkConfig()
	assert.Nil(t, err, "failed to get network config")
	assert.NotNil(t, networkConfig, "Invalid networkConfig")

	//Client
	assert.True(t, networkConfig.Client.Organization == "org1")

	//Channel
	assert.Equal(t, len(networkConfig.Channels), 3)
	assert.Equal(t, len(networkConfig.Channels["mychannel"].Peers), 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String(), (5 * time.Second).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String(), (500 * time.Millisecond).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor, 2.0)

	//CertificateAuthorities
	assert.Equal(t, len(networkConfig.CertificateAuthorities), 2)
	assert.Equal(t, networkConfig.CertificateAuthorities["local.ca.org1.example.com"].URL, "https://ca.org1.example.com:7054")
	assert.Equal(t, networkConfig.CertificateAuthorities["local.ca.org2.example.com"].URL, "https://ca.org2.example.com:8054")

	//EntityMatchers
	assert.Equal(t, len(networkConfig.EntityMatchers), 4)
	assert.Equal(t, len(networkConfig.EntityMatchers["peer"]), 8)
	assert.Equal(t, networkConfig.EntityMatchers["peer"][0].MappedHost, "local.peer0.org1.example.com")
	assert.Equal(t, len(networkConfig.EntityMatchers["orderer"]), 4)
	assert.Equal(t, networkConfig.EntityMatchers["orderer"][0].MappedHost, "local.orderer.example.com")
	assert.Equal(t, len(networkConfig.EntityMatchers["certificateauthority"]), 2)
	assert.Equal(t, networkConfig.EntityMatchers["certificateauthority"][0].MappedHost, "local.ca.org1.example.com")
	assert.Equal(t, len(networkConfig.EntityMatchers["channel"]), 1)
	assert.Equal(t, networkConfig.EntityMatchers["channel"][0].MappedName, "ch1")

	//Organizations
	assert.Equal(t, len(networkConfig.Organizations), 3)
	assert.Equal(t, networkConfig.Organizations["org1"].MSPID, "Org1MSP")

	//Orderer
	assert.Equal(t, len(networkConfig.Orderers), 1)
	assert.Equal(t, networkConfig.Orderers["local.orderer.example.com"].URL, "orderer.example.com:7050")

	//Peer
	assert.Equal(t, len(networkConfig.Peers), 2)
	assert.Equal(t, networkConfig.Peers["local.peer0.org1.example.com"].URL, "peer0.org1.example.com:7051")
	assert.Equal(t, networkConfig.Peers["local.peer0.org1.example.com"].EventURL, "peer0.org1.example.com:7053")

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
	networkPeers, err := endpointConfig.NetworkPeers()

	assert.Nil(t, err, "not suppopsed to get error")
	assert.NotNil(t, networkPeers, "supposed to get valid network peers")
	assert.Equal(t, 2, len(networkPeers), "supposed to get 2 network peers")
	assert.NotEmpty(t, networkPeers[0].MSPID)
	assert.NotEmpty(t, networkPeers[1].MSPID)
	assert.NotEmpty(t, networkPeers[0].PeerConfig)
	assert.NotEmpty(t, networkPeers[1].PeerConfig)

	//cross check with peer config for org1
	peerConfigOrg1, err := endpointConfig.PeersConfig("org1")
	assert.Nil(t, err, "not suppopsed to get error")

	//cross check with peer config for org2
	peerConfigOrg2, err := endpointConfig.PeersConfig("org2")
	assert.Nil(t, err, "not suppopsed to get error")

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
		},
	}
	(channelsMap.(map[string]interface{}))[orgChannelID] = orgChannel
}

func getMatcherConfig() core.ConfigBackend {
	cfgBackend, err := config.FromFile(configTestEntityMatchersFilePath)()
	if err != nil {
		panic(fmt.Sprintf("Unexpected error reading config: %v", err))
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

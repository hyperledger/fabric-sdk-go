/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
)

const (
	sampleMatchersOverrideAll      = "matchers_sample1.yaml"
	sampleMatchersOverridePartial  = "matchers_sample2.yaml"
	sampleMatchersURLMapping       = "matchers_sample3.yaml"
	sampleMatchersHostNameOverride = "matchers_sample4.yaml"
	sampleMatchersDefaultConfigs   = "matchers_sample5.yaml"
	sampleMatchersIgnoreEndpoint   = "matchers_sample6.yaml"
	matchersDir                    = "matcher-samples"

	actualPeerURL                 = "peer0.org1.example.com:7051"
	actualPeerHostNameOverride    = "peer0.org1.example.com"
	actualOrdererURL              = "orderer.example.com:7050"
	actualOrdererHostNameOverride = "orderer.example.com"

	overridedPeerURL                 = "peer0.org1.example.com:8888"
	overridedPeerHostNameOverride    = "peer0.org1.override.com"
	overridedOrdererURL              = "orderer.example.com:8888"
	overridedOrdererHostNameOverride = "orderer.override.com"

	testChannelID = "matcherchannel"
)

func getConfigPath() string {
	return filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata")
}

//TestAllOptionsOverride
//Scenario: Actual peer/orderer config are overridden using entity matchers.
// Endpoint config matches given URL/name with all available entity matcher patterns first to get the
//overrided values, rest of the options are fetched from mapped host.
//If no entity matcher provided, then it falls back to exact match in endpoint configuration.
func TestAllOptionsOverride(t *testing.T) {
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersOverrideAll)
	configPath := filepath.Join(getConfigPath(), configTestFile)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("peer0.org1.example.com:7051")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//PeerConfig Search Based on Name configured in config
	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok, ignoreOrderer := config.OrdererConfig("orderer.example.com:7051")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//OrdererConfig Search Based on Name configured in config
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//Channel Endpoint Config Search Based on Name configured in config
	channelConfig := config.ChannelConfig("testXYZchannel")
	assert.Equal(t, 1, len(channelConfig.Orderers))
	assert.Equal(t, 1, len(channelConfig.Peers))
}

//TestPartialOptionsOverride
//Scenario: Entity matchers overriding only few options (URLs) in Peer/Orderer config. Options which are not overridden
// are fetched from mapped host entity.
func TestPartialOptionsOverride(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersOverridePartial)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("peer0.org1.example.com:7051")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, actualPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//PeerConfig Search Based on Name configured in config
	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, actualPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok, ignoreOrderer := config.OrdererConfig("orderer.example.com:7051")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, actualOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//OrdererConfig Search Based on Name configured in config
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, actualOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

}

//TestOnlyHostNameOptionsOverride
//Scenario: Entity matchers overriding only few options(hostname overrides) in Peer/Orderer config. Options which are not overridden
// are fetched from mapped host entity.
func TestOnlyHostNameOptionsOverride(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersHostNameOverride)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("peer0.org1.example.com:7051")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, actualPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//PeerConfig Search Based on Name configured in config
	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, actualPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok, ignoreOrderer := config.OrdererConfig("orderer.example.com:7051")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, actualOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//OrdererConfig Search Based on Name configured in config
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, actualOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

}

//TestURLMapping
//Scenario:  A URL based entity pattern is used to return config entities with customized URLs, host overrides etc
func TestURLMapping(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersURLMapping)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("my.org.exampleX.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok, ignoreOrderer := config.OrdererConfig("my.org.exampleX.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//PeerConfig Search Based on URL configured in config (using $ in entity matchers)
	peerConfig, ok = config.PeerConfig("peer0.exampleY.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	peerConfig, ok = config.PeerConfig("sample-org0peer1.demo.example.com:4321")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, "xsample-org0peer1.demo.example.com:4321", peerConfig.URL)
	assert.Equal(t, "xsample-org0peer1.demo.example.com", peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config (using $ in entity matchers)
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.exampleY.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

}

func TestPeerMatchersWithDefaults(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersDefaultConfigs)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on matching URL,
	peerConfig, ok := config.PeerConfig("XYZ.org.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peer0.org1.override.com")

	//PeerConfig Search Based on matching URL, using regex replace pattern (unknown pattern)
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	peerConfig, ok = config.PeerConfig("peerABC.replace.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, "peerABC.org1.example.com:1234", peerConfig.URL)
	assert.Equal(t, "peerABC.org1.override.com", peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peerABC.org1.override.com")

	//PeerConfig Search Based on matching URL, where mapped host is missing
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	peerConfig, ok = config.PeerConfig("peer0.missing.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peer0.org1.override.com")

	//PeerConfig Search Based on matching URL, where non existing mapped host is used
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	peerConfig, ok = config.PeerConfig("peer0.random.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peer0.org1.override.com")
}

func TestOrdererMatchersWithDefaults(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersDefaultConfigs)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//OrdererConfig Search Based on matching URL,
	ordererConfig, ok, ignoreOrderer := config.OrdererConfig("XYZ.org.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "orderer.override.com")

	//PeerConfig Search Based on matching URL, using regex replace pattern (unknown pattern)
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("ordABC.replace.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, "ordABC.example.com:1234", ordererConfig.URL)
	assert.Equal(t, "ordABC.override.com", ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "ordABC.override.com")

	//PeerConfig Search Based on matching URL, where mapped host is missing
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("ordABC.missing.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "orderer.override.com")

	//PeerConfig Search Based on matching URL, where non existing mapped host is used
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("ordABC.random.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "orderer.override.com")
}

func verifyGRPCOpts(t *testing.T, grpcOpts map[string]interface{}, expectedHost string) {
	assert.Equal(t, 6, len(grpcOpts))
	assert.Equal(t, "1s", grpcOpts["keep-alive-time"])
	assert.Equal(t, expectedHost, grpcOpts["ssl-target-name-override"])
	assert.Equal(t, "21s", grpcOpts["keep-alive-timeout"])
	assert.Equal(t, true, grpcOpts["keep-alive-permit"])
	assert.Equal(t, true, grpcOpts["fail-fast"])
	assert.Equal(t, true, grpcOpts["allow-insecure"])

	//make sure map has all the expected grpc opts keys
	_, ok := grpcOpts["keep-alive-time"]
	assert.True(t, ok)
	_, ok = grpcOpts["ssl-target-name-override"]
	assert.True(t, ok)
	_, ok = grpcOpts["keep-alive-timeout"]
	assert.True(t, ok)
	_, ok = grpcOpts["keep-alive-permit"]
	assert.True(t, ok)
	_, ok = grpcOpts["fail-fast"]
	assert.True(t, ok)
	_, ok = grpcOpts["allow-insecure"]
	assert.True(t, ok)
}

func getBackendsFromFiles(files ...string) ([]core.ConfigBackend, error) {

	backends := make([]core.ConfigBackend, len(files))
	for i, file := range files {
		backend, err := config.FromFile(file)()
		if err != nil {
			return nil, err
		}
		backends[i] = backend[0]
	}
	return backends, nil
}

//TestDefaultPeerForNonExistingURL tests default peerConfig result for search by URL scenario
func TestDefaultPeerForNonExistingURL(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersDefaultConfigs)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on unmatched URL, default peerconfig with searchkey as URL should be returned
	peerConfig, ok := config.PeerConfig("ABC.XYZ:2222")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, "ABC.XYZ:2222", peerConfig.URL)
	assert.Equal(t, nil, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)

	assert.Equal(t, "1s", peerConfig.GRPCOptions["keep-alive-time"])
	assert.Equal(t, nil, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, "21s", peerConfig.GRPCOptions["keep-alive-timeout"])
	assert.Equal(t, true, peerConfig.GRPCOptions["keep-alive-permit"])
	assert.Equal(t, true, peerConfig.GRPCOptions["fail-fast"])
	assert.Equal(t, true, peerConfig.GRPCOptions["allow-insecure"])

	//make sure map has all the expected grpc opts keys
	_, ok = peerConfig.GRPCOptions["keep-alive-time"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["ssl-target-name-override"]
	assert.False(t, ok)
	_, ok = peerConfig.GRPCOptions["keep-alive-timeout"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["keep-alive-permit"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["fail-fast"]
	assert.True(t, ok)
	_, ok = peerConfig.GRPCOptions["allow-insecure"]
	assert.True(t, ok)

}

//TestDefaultOrdererForNonExistingURL tests default ordererConfig result for search by URL scenario
func TestDefaultOrdererForNonExistingURL(t *testing.T) {
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersDefaultConfigs)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on unmatched URL, default peerconfig with searchkey as URL should be returned
	ordererConfig, ok, ignoreOrderer := config.OrdererConfig("ABC.XYZ:2222")
	assert.True(t, ok, "supposed to find peer config")
	assert.False(t, ignoreOrderer, "orderer must not be ignored")
	assert.Equal(t, "ABC.XYZ:2222", ordererConfig.URL)
	assert.Equal(t, nil, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)

	assert.Equal(t, "1s", ordererConfig.GRPCOptions["keep-alive-time"])
	assert.Equal(t, nil, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, "21s", ordererConfig.GRPCOptions["keep-alive-timeout"])
	assert.Equal(t, true, ordererConfig.GRPCOptions["keep-alive-permit"])
	assert.Equal(t, true, ordererConfig.GRPCOptions["fail-fast"])
	assert.Equal(t, true, ordererConfig.GRPCOptions["allow-insecure"])

	//make sure map has all the expected grpc opts keys
	_, ok = ordererConfig.GRPCOptions["keep-alive-time"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["ssl-target-name-override"]
	assert.False(t, ok)
	_, ok = ordererConfig.GRPCOptions["keep-alive-timeout"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["keep-alive-permit"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["fail-fast"]
	assert.True(t, ok)
	_, ok = ordererConfig.GRPCOptions["allow-insecure"]
	assert.True(t, ok)

}

//TestMatchersIgnoreEndpoint tests entity matcher ignore endpoint feature
// If marked as `IgnoreEndpoint: true` then config for,
//				peer excluded in org peers
//				peer excluded in network peers
//				peer excluded in peer search by name
//				peer excluded in peer search by URL
//				orderer excluded in all orderers list
//				orderer excluded in orderer search by name
//				orderer excluded in orderer search by URL
//				peer/orderer excluded in networkconfig
func TestMatchersIgnoreEndpoint(t *testing.T) {

	//prepare backends for test
	configPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersIgnoreEndpoint)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	//get config from backend
	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//Test if orderer excluded in channel orderers
	testIgnoreEndpointChannelOrderers(t, config)

	//Test if peer excluded in channel peers
	testIgnoreEndpointChannelPeers(t, config)

	//Test if orderer/peer excluded in channel config
	testIgnoreEndpointChannelConfig(t, config)

	//Test if peer excluded in org peers
	testIgnoreEndpointOrgPeers(t, config)

	//Test if peer excluded in network peers
	testIgnoreEndpointNetworkPeers(t, config)

	//Test if peer excluded in peer search by URL
	testIgnoreEndpointPeerSearch(t, config)

	//Test if orderer excluded in all orderers
	testIgnoreEndpointAllOrderers(t, config)

	//Test if orderer excluded in orderer search by name/URL
	testIgnoreEndpointOrdererSearch(t, config)

	//test NetworkConfig
	testIgnoreEndpointNetworkConfig(t, config)
}

func testIgnoreEndpointChannelOrderers(t *testing.T, config fab.EndpointConfig) {
	orderers := config.ChannelOrderers(testChannelID)
	assert.Equal(t, 1, len(orderers))
	checkOrdererConfigExcluded(orderers, "orderer.exclude.example.com", t)
}

func testIgnoreEndpointChannelPeers(t *testing.T, config fab.EndpointConfig) {
	channelPeers := config.ChannelPeers(testChannelID)
	assert.Equal(t, 2, len(channelPeers))
	checkChannelPeerExcluded(channelPeers, "peer1.org", t)
}

func testIgnoreEndpointChannelConfig(t *testing.T, config fab.EndpointConfig) {
	chNwConfig := config.ChannelConfig(testChannelID)
	assert.NotNil(t, chNwConfig)
	assert.NotEmpty(t, chNwConfig.Peers)
	_, ok := chNwConfig.Peers["peer1.org1.example.com"]
	assert.False(t, ok, "should not have excluded peer's entry in channel network config")
	_, ok = chNwConfig.Peers["peer1.org2.example.com"]
	assert.False(t, ok, "should not have excluded peer's entry in channel network config")
	assert.NotEmpty(t, chNwConfig.Orderers)
	assert.Equal(t, 1, len(chNwConfig.Orderers))
	assert.NotEqual(t, "orderer.exclude.example.com", chNwConfig.Orderers[0])
}

func testIgnoreEndpointOrgPeers(t *testing.T, config fab.EndpointConfig) {
	// test org 1 peers
	orgPeers, ok := config.PeersConfig("org1")
	assert.True(t, ok)
	assert.NotEmpty(t, orgPeers)
	checkPeerConfigExcluded(orgPeers, "peer1.org1", t)

	// test org 2 peers
	orgPeers, ok = config.PeersConfig("org2")
	assert.True(t, ok)
	assert.NotEmpty(t, orgPeers)
	checkPeerConfigExcluded(orgPeers, "peer1.org2", t)
}

func testIgnoreEndpointNetworkPeers(t *testing.T, config fab.EndpointConfig) {
	nwPeers := config.NetworkPeers()
	assert.NotEmpty(t, nwPeers)
	checkNetworkPeerExcluded(nwPeers, "peer1.org", t)
}

func testIgnoreEndpointPeerSearch(t *testing.T, config fab.EndpointConfig) {

	//Test if peer excluded in peer search by URL
	peerConfig, ok := config.PeerConfig("peer1.org1.example.com:7151")
	assert.False(t, ok)
	assert.Nil(t, peerConfig)

	peerConfig, ok = config.PeerConfig("peer1.org2.example.com:8051")
	assert.False(t, ok)
	assert.Nil(t, peerConfig)

	//arbitrary URLs
	peerConfig, ok = config.PeerConfig("peer1.org1.example.com:1234")
	assert.False(t, ok)
	assert.Nil(t, peerConfig)

	peerConfig, ok = config.PeerConfig("peer1.org2.example.com:4321")
	assert.False(t, ok)
	assert.Nil(t, peerConfig)

	//Test if peer excluded in peer search by name
	peerConfig, ok = config.PeerConfig("peer1.org1.example.com")
	assert.False(t, ok)
	assert.Nil(t, peerConfig)

	peerConfig, ok = config.PeerConfig("peer1.org2.example.com")
	assert.False(t, ok)
	assert.Nil(t, peerConfig)

	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok)
	assert.NotNil(t, peerConfig)

	peerConfig, ok = config.PeerConfig("peer0.org2.example.com")
	assert.True(t, ok)
	assert.NotNil(t, peerConfig)
}

func testIgnoreEndpointAllOrderers(t *testing.T, config fab.EndpointConfig) {
	ordererConfigs := config.OrderersConfig()
	assert.NotEmpty(t, ordererConfigs)
	checkOrdererConfigExcluded(ordererConfigs, "orderer.exclude.", t)
}

func testIgnoreEndpointOrdererSearch(t *testing.T, config fab.EndpointConfig) {

	//Test if orderer excluded in orderer search by name

	ordererConfig, ok, ignoreOrderer:= config.OrdererConfig("orderer.exclude.example.com")
	assert.False(t, ok)
	assert.True(t, ignoreOrderer, "orderer must be excluded")
	assert.Nil(t, ordererConfig)

	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.exclude.example.com")
	assert.False(t, ok)
	assert.True(t, ignoreOrderer, "orderer must be excluded")
	assert.Nil(t, ordererConfig)

	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok)
	assert.False(t, ignoreOrderer, "orderer must not be excluded")
	assert.NotNil(t, ordererConfig)

	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok)
	assert.False(t, ignoreOrderer, "orderer must not be excluded")
	assert.NotNil(t, ordererConfig)

	//Test if orderer excluded in orderer search by URL

	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.exclude.example.com:8050")
	assert.False(t, ok)
	assert.True(t, ignoreOrderer, "orderer must be excluded")
	assert.Nil(t, ordererConfig)

	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.exclude.example.com:8050")
	assert.False(t, ok)
	assert.True(t, ignoreOrderer, "orderer must be excluded")
	assert.Nil(t, ordererConfig)

	//arbitrary URLs
	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.exclude.example.com:1234")
	assert.False(t, ok)
	assert.True(t, ignoreOrderer, "orderer must be excluded")
	assert.Nil(t, ordererConfig)

	ordererConfig, ok, ignoreOrderer = config.OrdererConfig("orderer.exclude.example.com:4321")
	assert.False(t, ok)
	assert.True(t, ignoreOrderer, "orderer must be excluded")
	assert.Nil(t, ordererConfig)

}

func testIgnoreEndpointNetworkConfig(t *testing.T, config fab.EndpointConfig) {
	networkConfig := config.NetworkConfig()
	assert.NotNil(t, networkConfig)
	assert.Equal(t, 2, len(networkConfig.Peers))
	assert.Equal(t, 1, len(networkConfig.Orderers))
	_, ok := networkConfig.Peers["peer1.org1.example.com"]
	assert.False(t, ok)
	_, ok = networkConfig.Peers["peer1.org2.example.com"]
	assert.False(t, ok)
	_, ok = networkConfig.Peers["peer0.org1.example.com"]
	assert.True(t, ok)
	_, ok = networkConfig.Peers["peer0.org2.example.com"]
	assert.True(t, ok)
	_, ok = networkConfig.Orderers["orderer.exclude.example.com"]
	assert.False(t, ok)
	_, ok = networkConfig.Orderers["orderer.example.com"]
	assert.True(t, ok)
}

func checkOrdererConfigExcluded(ordererConfigs []fab.OrdererConfig, excluded string, t *testing.T) {
	for _, v := range ordererConfigs {
		assert.False(t, strings.Contains(strings.ToLower(v.URL), strings.ToLower(excluded)), "'%s' supposed to be excluded", v.URL)
	}
}

func checkChannelPeerExcluded(peerConfigs []fab.ChannelPeer, excluded string, t *testing.T) {
	for _, v := range peerConfigs {
		assert.False(t, strings.Contains(strings.ToLower(v.URL), strings.ToLower(excluded)), "'%s' supposed to be excluded", v.URL)
	}
}

func checkPeerConfigExcluded(peerConfigs []fab.PeerConfig, excluded string, t *testing.T) {
	for _, v := range peerConfigs {
		assert.False(t, strings.Contains(strings.ToLower(v.URL), strings.ToLower(excluded)), "'%s' supposed to be excluded", v.URL)
	}
}

func checkNetworkPeerExcluded(peerConfigs []fab.NetworkPeer, excluded string, t *testing.T) {
	for _, v := range peerConfigs {
		assert.False(t, strings.Contains(strings.ToLower(v.URL), strings.ToLower(excluded)), "'%s' supposed to be excluded", v.URL)
	}
}

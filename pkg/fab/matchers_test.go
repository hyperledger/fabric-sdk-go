/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/stretchr/testify/assert"
)

const (
	sampleMatchersOverrideAll      = "../core/config/testdata/matcher-samples/matchers_sample1.yaml"
	sampleMatchersOverridePartial  = "../core/config/testdata/matcher-samples/matchers_sample2.yaml"
	sampleMatchersURLMapping       = "../core/config/testdata/matcher-samples/matchers_sample3.yaml"
	sampleMatchersHostNameOverride = "../core/config/testdata/matcher-samples/matchers_sample4.yaml"
	sampleMatchersDefaultConfigs   = "../core/config/testdata/matcher-samples/matchers_sample5.yaml"

	actualPeerURL                 = "peer0.org1.example.com:7051"
	actualPeerEventURL            = "peer0.org1.example.com:7053"
	actualPeerHostNameOverride    = "peer0.org1.example.com"
	actualOrdererURL              = "orderer.example.com:7050"
	actualOrdererHostNameOverride = "orderer.example.com"

	overridedPeerURL                 = "peer0.org1.example.com:8888"
	overridedPeerEventURL            = "peer0.org1.example.com:9999"
	overridedPeerHostNameOverride    = "peer0.org1.override.com"
	overridedOrdererURL              = "orderer.example.com:8888"
	overridedOrdererHostNameOverride = "orderer.override.com"
)

//TestAllOptionsOverride
//Scenario: Actual peer/orderer config are overridden using entity matchers.
// Endpoint config matches given URL/name with all available entity matcher patterns first to get the
//overrided values, rest of the options are fetched from mapped host.
//If no entity matcher provided, then it falls back to exact match in endpoint configuration.
func TestAllOptionsOverride(t *testing.T) {
	backends, err := getBackendsFromFiles(sampleMatchersOverrideAll, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("peer0.org1.example.com:7051")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//PeerConfig Search Based on Name configured in config
	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok := config.OrdererConfig("orderer.example.com:7051")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//OrdererConfig Search Based on Name configured in config
	ordererConfig, ok = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//Channel Endpoint Config Search Based on Name configured in config
	channelConfig, ok := config.ChannelConfig("testXYZchannel")
	assert.True(t, ok, "supposed to find channel config")
	assert.Equal(t, 1, len(channelConfig.Orderers))
	assert.Equal(t, 1, len(channelConfig.Peers))
}

//TestPartialOptionsOverride
//Scenario: Entity matchers overriding only few options (URLs) in Peer/Orderer config. Options which are not overridden
// are fetched from mapped host entity.
func TestPartialOptionsOverride(t *testing.T) {
	backends, err := getBackendsFromFiles(sampleMatchersOverridePartial, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("peer0.org1.example.com:7051")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, actualPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, actualPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//PeerConfig Search Based on Name configured in config
	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, actualPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, actualPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok := config.OrdererConfig("orderer.example.com:7051")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, actualOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//OrdererConfig Search Based on Name configured in config
	ordererConfig, ok = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, actualOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

}

//TestOnlyHostNameOptionsOverride
//Scenario: Entity matchers overriding only few options(hostname overrides) in Peer/Orderer config. Options which are not overridden
// are fetched from mapped host entity.
func TestOnlyHostNameOptionsOverride(t *testing.T) {
	backends, err := getBackendsFromFiles(sampleMatchersHostNameOverride, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("peer0.org1.example.com:7051")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, actualPeerURL, peerConfig.URL)
	assert.Equal(t, actualPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//PeerConfig Search Based on Name configured in config
	peerConfig, ok = config.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, actualPeerURL, peerConfig.URL)
	assert.Equal(t, actualPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok := config.OrdererConfig("orderer.example.com:7051")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, actualOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//OrdererConfig Search Based on Name configured in config
	ordererConfig, ok = config.OrdererConfig("orderer.example.com")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, actualOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

}

//TestURLMapping
//Scenario:  A URL based entity pattern is used to return config entities with customized URLs, host overrides etc
func TestURLMapping(t *testing.T) {
	backends, err := getBackendsFromFiles(sampleMatchersURLMapping, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on URL configured in config
	peerConfig, ok := config.PeerConfig("my.org.exampleX.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config
	ordererConfig, ok := config.OrdererConfig("my.org.exampleX.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

	//PeerConfig Search Based on URL configured in config (using $ in entity matchers)
	peerConfig, ok = config.PeerConfig("peer0.exampleY.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(peerConfig.GRPCOptions))

	//OrdererConfig Search Based on URL configured in config (using $ in entity matchers)
	ordererConfig, ok = config.OrdererConfig("orderer.exampleY.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.Equal(t, 6, len(ordererConfig.GRPCOptions))

}

func TestPeerMatchersWithDefaults(t *testing.T) {
	backends, err := getBackendsFromFiles(sampleMatchersDefaultConfigs, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//PeerConfig Search Based on matching URL,
	peerConfig, ok := config.PeerConfig("XYZ.org.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peer0.org1.override.com")

	//PeerConfig Search Based on matching URL, using regex replace pattern (unknown pattern)
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	peerConfig, ok = config.PeerConfig("peerABC.replace.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, "peerABC.org1.example.com:1234", peerConfig.URL)
	assert.Equal(t, "peerABC.org1.example.com:1234", peerConfig.EventURL)
	assert.Equal(t, "peerABC.org1.override.com", peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peerABC.org1.override.com")

	//PeerConfig Search Based on matching URL, where mapped host is missing
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	peerConfig, ok = config.PeerConfig("peer0.missing.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peer0.org1.override.com")

	//PeerConfig Search Based on matching URL, where non existing mapped host is used
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	peerConfig, ok = config.PeerConfig("peer0.random.example.com:1234")
	assert.True(t, ok, "supposed to find peer config")
	assert.Equal(t, overridedPeerURL, peerConfig.URL)
	assert.Equal(t, overridedPeerEventURL, peerConfig.EventURL)
	assert.Equal(t, overridedPeerHostNameOverride, peerConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, peerConfig.TLSCACert)
	verifyGRPCOpts(t, peerConfig.GRPCOptions, "peer0.org1.override.com")
}

func TestOrdererMatchersWithDefaults(t *testing.T) {
	backends, err := getBackendsFromFiles(sampleMatchersDefaultConfigs, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	//OrdererConfig Search Based on matching URL,
	ordererConfig, ok := config.OrdererConfig("XYZ.org.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "orderer.override.com")

	//PeerConfig Search Based on matching URL, using regex replace pattern (unknown pattern)
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	ordererConfig, ok = config.OrdererConfig("ordABC.replace.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, "ordABC.example.com:1234", ordererConfig.URL)
	assert.Equal(t, "ordABC.override.com", ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "ordABC.override.com")

	//PeerConfig Search Based on matching URL, where mapped host is missing
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	ordererConfig, ok = config.OrdererConfig("ordABC.missing.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
	assert.Equal(t, overridedOrdererURL, ordererConfig.URL)
	assert.Equal(t, overridedOrdererHostNameOverride, ordererConfig.GRPCOptions["ssl-target-name-override"])
	assert.NotNil(t, ordererConfig.TLSCACert)
	verifyGRPCOpts(t, ordererConfig.GRPCOptions, "orderer.override.com")

	//PeerConfig Search Based on matching URL, where non existing mapped host is used
	//it shouldn't fail since default peer config will be picked for unknown mapped host
	ordererConfig, ok = config.OrdererConfig("ordABC.random.example.com:1234")
	assert.True(t, ok, "supposed to find orderer config")
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

//TODO to be enabled once default search result for peerConfig search by URL is available
//func TestDefaultPeerForNonExistingURL(t *testing.T) {
//	backends, err := getBackendsFromFiles(sampleMatchersDefaultConfigs, configTestFilePath)
//	assert.Nil(t, err, "not supposed to get error")
//	assert.Equal(t, 2, len(backends))
//
//	config, err := ConfigFromBackend(backends...)
//	assert.Nil(t, err, "not supposed to get error")
//	assert.NotNil(t, config)
//
//	//PeerConfig Search Based on unmatched URL, default peerconfig with searchkey as URL should be returned
//	peerConfig, ok := config.PeerConfig("ABC.XYZ:2222")
//	assert.True(t, ok, "supposed to find peer config")
//	assert.Equal(t, "ABC.XYZ:2222", peerConfig.URL)
//	assert.Equal(t, "", peerConfig.EventURL)
//	assert.Equal(t, nil, peerConfig.GRPCOptions["ssl-target-name-override"])
//	assert.NotNil(t, peerConfig.TLSCACert)
//
//	assert.Equal(t, "1s", peerConfig.GRPCOptions["keep-alive-time"])
//	assert.Equal(t, nil, peerConfig.GRPCOptions["ssl-target-name-override"])
//	assert.Equal(t, "21s", peerConfig.GRPCOptions["keep-alive-timeout"])
//	assert.Equal(t, true, peerConfig.GRPCOptions["keep-alive-permit"])
//	assert.Equal(t, true, peerConfig.GRPCOptions["fail-fast"])
//	assert.Equal(t, true, peerConfig.GRPCOptions["allow-insecure"])
//
//	//make sure map has all the expected grpc opts keys
//	_, ok = peerConfig.GRPCOptions["keep-alive-time"]
//	assert.True(t, ok)
//	_, ok = peerConfig.GRPCOptions["ssl-target-name-override"]
//	assert.False(t, ok)
//	_, ok = peerConfig.GRPCOptions["keep-alive-timeout"]
//	assert.True(t, ok)
//	_, ok = peerConfig.GRPCOptions["keep-alive-permit"]
//	assert.True(t, ok)
//	_, ok = peerConfig.GRPCOptions["fail-fast"]
//	assert.True(t, ok)
//	_, ok = peerConfig.GRPCOptions["allow-insecure"]
//	assert.True(t, ok)
//
//}

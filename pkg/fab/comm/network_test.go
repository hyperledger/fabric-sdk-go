/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/stretchr/testify/assert"
)

const configTestFilePath = "../../core/config/testdata/config_test.yaml"
const entityMatcherTestFilePath = "../../core/config/testdata/config_test_entity_matchers.yaml"
const localOverrideEntityMatcher = "../../../test/fixtures/config/overrides/local_entity_matchers.yaml"

func TestNetworkPeerConfigFromURL(t *testing.T) {
	configBackend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config backend: %s", err)
	}

	sampleConfig, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Unexpected error reading config: %s", err)
	}

	_, err = NetworkPeerConfig(sampleConfig, "invalid")
	assert.NotNil(t, err, "invalid url should return err")

	np, err := NetworkPeerConfig(sampleConfig, "peer0.org2.example.com:8051")
	assert.Nil(t, err, "valid url should not return err")
	assert.Equal(t, "peer0.org2.example.com:8051", np.URL, "wrong URL")
	assert.Equal(t, "Org2MSP", np.MSPID, "wrong MSP")

	np, err = NetworkPeerConfig(sampleConfig, "peer0.org1.example.com:7051")
	assert.Nil(t, err, "valid url should not return err")
	assert.Equal(t, "peer0.org1.example.com:7051", np.URL, "wrong URL")
	assert.Equal(t, "Org1MSP", np.MSPID, "wrong MSP")
}

func TestSearchPeerConfigFromURL(t *testing.T) {
	configBackend1, err := config.FromFile(localOverrideEntityMatcher)()
	if err != nil {
		t.Fatalf("Unexpected error reading config backend: %s", err)
	}

	configBackend2, err := config.FromFile(entityMatcherTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config backend: %s", err)
	}

	//override entitymatcher
	backends := append([]core.ConfigBackend{}, configBackend1...)
	backends = append(backends, configBackend2...)

	sampleConfig, err := fabImpl.ConfigFromBackend(backends...)
	if err != nil {
		t.Fatalf("Unexpected error reading config: %s", err)
	}

	peer0Org1, ok := sampleConfig.PeerConfig("peer0.org1.example.com")
	assert.True(t, ok, "peerconfig search was expected to be successful")

	//Positive scenario,
	// peerconfig should be found using matched URL
	testURL := "localhost:7051"
	peerConfig, err := SearchPeerConfigFromURL(sampleConfig, testURL)
	assert.Nil(t, err, "supposed to get no error")
	assert.NotNil(t, peerConfig, "supposed to get valid peerConfig by url :%s", testURL)
	assert.Equal(t, testURL, peerConfig.URL)
	assert.Nil(t, err, "supposed to get no error")
	assert.Equal(t, peer0Org1.EventURL, peerConfig.EventURL)

	// peerconfig should be found using actual URL
	testURL2 := "peer0.org1.example.com:7051"
	peerConfig, err = SearchPeerConfigFromURL(sampleConfig, testURL2)

	assert.Nil(t, err, "supposed to get no error")
	assert.NotNil(t, peerConfig, "supposed to get valid peerConfig by url :%s", testURL2)
	assert.Equal(t, testURL, peerConfig.URL)
	assert.Nil(t, err, "supposed to get no error")
	assert.Equal(t, peer0Org1.EventURL, peerConfig.EventURL)

}

func TestMSPID(t *testing.T) {
	configBackend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config backend: %s", err)
	}

	sampleConfig, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Unexpected error reading config: %s", err)
	}

	mspID, ok := MSPID(sampleConfig, "invalid")
	assert.False(t, ok, "supposed to fail for invalid org name")
	assert.Empty(t, mspID, "supposed to get valid MSP ID")

	mspID, ok = MSPID(sampleConfig, "org1")
	assert.True(t, ok, "supposed to pass with valid org name")
	assert.NotEmpty(t, mspID, "supposed to get valid MSP ID")
	assert.Equal(t, "Org1MSP", mspID, "supposed to get valid MSP ID")
}

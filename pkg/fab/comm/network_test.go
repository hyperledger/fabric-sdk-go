/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/stretchr/testify/assert"
)

const configTestFilePath = "../../core/config/testdata/config_test.yaml"

func TestNetworkPeerConfigFromURL(t *testing.T) {
	configBackend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config backend: %v", err)
	}

	sampleConfig, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("Unexpected error reading config: %v", err)
	}

	_, err = NetworkPeerConfigFromURL(sampleConfig, "invalid")
	assert.NotNil(t, err, "invalid url should return err")

	np, err := NetworkPeerConfigFromURL(sampleConfig, "peer0.org2.example.com:8051")
	assert.Nil(t, err, "valid url should not return err")
	assert.Equal(t, "peer0.org2.example.com:8051", np.URL, "wrong URL")
	assert.Equal(t, "Org2MSP", np.MSPID, "wrong MSP")

	np, err = NetworkPeerConfigFromURL(sampleConfig, "peer0.org1.example.com:7051")
	assert.Nil(t, err, "valid url should not return err")
	assert.Equal(t, "peer0.org1.example.com:7051", np.URL, "wrong URL")
	assert.Equal(t, "Org1MSP", np.MSPID, "wrong MSP")
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"
)

const configFile = "config_test.yaml"

func TestStaticDiscovery(t *testing.T) {

	ctx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("user1", "Org1MSP"))
	discoveryService, err := NewService(ctx.EndpointConfig(), ctx.InfraProvider(), "mychannel")
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	peers, err := discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("Failed to get peers from discovery service: %s", err)
	}

	// One peer is configured for "mychannel"
	expectedNumOfPeeers := 1
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}

}

func TestStaticDiscoveryWhenChannelIsEmpty(t *testing.T) {

	ctx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("user1", "Org1MSP"))
	_, err := NewService(ctx.EndpointConfig(), ctx.InfraProvider(), "")
	assert.Error(t, err, "expecting error when channel ID is empty")
}

func TestStaticLocalDiscovery(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	configBackend, err := config.FromFile(configPath)()
	assert.NoError(t, err)

	config1, err := fabImpl.ConfigFromBackend(configBackend...)
	assert.NoError(t, err)

	discoveryProvider, err := NewLocalProvider(config1)
	assert.NoError(t, err)

	clientCtx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("user1", "Org1MSP"))
	discoveryProvider.Initialize(clientCtx)

	discoveryService, err := discoveryProvider.CreateLocalDiscoveryService(clientCtx.Identifier().MSPID)
	assert.NoError(t, err)

	peers, err := discoveryService.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(peers))
}

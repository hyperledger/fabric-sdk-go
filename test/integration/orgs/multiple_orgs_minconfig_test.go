// +build devstable

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bootStrapCC = "btspExampleCC"

//TestOrgsEndToEndWithBootstrapConfigs does the same as TestOrgsEndToEnd with the difference of loading
// minimal configs instead of the normal config_test.yaml configs and with the help of discovery service to discover
// other peers not in the config (example org1 has 2 peers and only peer0 is defined in the bootstrap configs)
func TestOrgsEndToEndWithBootstrapConfigs(t *testing.T) {
	configPath := "../../fixtures/config/config_test_multiorg_bootstrap.yaml"
	sdk, err := fabsdk.New(config.FromFile(configPath),
		fabsdk.WithServicePkg(&DynamicDiscoveryProviderFactory{}),
	)
	if err != nil {
		require.NoError(t, err, "Failed to create new SDK")
	}
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//prepare contexts
	mc := multiorgContext{
		ordererClientContext:   sdk.Context(fabsdk.WithUser(ordererAdminUser), fabsdk.WithOrg(ordererOrgName)),
		org1AdminClientContext: sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1)),
		org2AdminClientContext: sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2)),
		ccName:                 bootStrapCC,
	}

	// create channel and join orderer/orgs peers to it if was not done already
	setupClientContextsAndChannel(t, sdk, &mc)

	// wait some time to allow the gossip to propagate the peers discovery
	time.Sleep(10 * time.Second)

	testDynamicDiscovery(t, sdk)

	// now run the same test as multiple_orgs_test.go to make sure it works with bootstrap config..

	// Load specific targets for move funds test
	loadOrgPeers(t, sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1)))

	expectedValue := testWithOrg1(t, sdk, &mc)
	expectedValue = testWithOrg2(t, expectedValue, mc.ccName)
	verifyWithOrg1(t, sdk, expectedValue, mc.ccName)
}

func testDynamicDiscovery(t *testing.T, sdk *fabsdk.FabricSDK) {
	// example discovering the peers from the bootstap peer
	// there should be three peers returned from discovery:
	// 1 org1 anchor peer (peer0.org1.example.com)
	// 1 discovered peer (not in config: peer1.org1.example.com)
	// 1 org2 anchor peer (peer0.org2.example.com)
	peersList := discoverPeers(t, sdk)
	assert.Equal(t, 3, len(peersList), "Expected exactly 3 peers as per %s's channel and %s's org configs", channelID, org2)
}

func discoverPeers(t *testing.T, sdk *fabsdk.FabricSDK) []fab.Peer {
	// any user from the network can access the discovery service, user org2User is selected for the test.
	chProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))
	chCtx, err := chProvider()
	require.NoError(t, err, "Error creating channel context")

	chCtx.ChannelService()
	peers, err := chCtx.DiscoveryService().GetPeers()
	require.NoErrorf(t, err, "Error getting peers for channel [%s]", channelID)
	require.NotEmptyf(t, peers, "No peers were found for channel [%s]", channelID)

	t.Logf("Peers of channel [%s]:", channelID)
	for i, p := range peers {
		t.Logf("%d- [%s] - MSP [%s]", i, p.URL(), p.MSPID())
	}
	return peers
}

// DynamicDiscoveryProviderFactory is configured with dynamic (endorser) selection provider
type DynamicDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateDiscoveryProvider returns a new dynamic discovery provider
func (f *DynamicDiscoveryProviderFactory) CreateDiscoveryProvider(config fab.EndpointConfig) (fab.DiscoveryProvider, error) {
	return dynamicdiscovery.New(config), nil
}

// CreateLocalDiscoveryProvider returns a new local dynamic discovery provider
func (f *DynamicDiscoveryProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return dynamicdiscovery.New(config), nil
}

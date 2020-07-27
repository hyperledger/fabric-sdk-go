// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	bootStrapCC    = "example_cc_fabtest_btsp"
	configFilename = "config_e2e_multiorg_bootstrap.yaml"
)

//TestOrgsEndToEndWithBootstrapConfigs does the same as TestOrgsEndToEnd with the difference of loading
// minimal configs instead of the normal config_e2e.yaml configs and with the help of discovery service to discover
// other peers not in the config (example org1 has 2 peers and only peer0 is defined in the bootstrap configs)
func TestOrgsEndToEndWithBootstrapConfigs(t *testing.T) {

	//load config provider
	configProvider := config.FromFile(integration.GetConfigPath(configFilename))

	//if local test, add entity matchers to override URLs to localhost
	if integration.IsLocal() {
		configProvider = integration.AddLocalEntityMapping(configProvider)
	}

	sdk, err := fabsdk.New(configProvider,
		fabsdk.WithServicePkg(&DynamicDiscoveryProviderFactory{}),
		fabsdk.WithProviderOpts(
			dynamicdiscovery.WithRefreshInterval(500*time.Millisecond),
		),
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
		ccName:                 bootStrapCC + metadata.TestRunID,
		ccVersion:              "0",
		channelID:              channelID,
	}

	// create channel and join orderer/orgs peers to it if was not done already
	setupClientContextsAndChannel(t, sdk, &mc)

	org1Peers, err := integration.DiscoverLocalPeers(mc.org1AdminClientContext, 2)
	require.NoError(t, err)
	_, err = integration.DiscoverLocalPeers(mc.org2AdminClientContext, 2)
	require.NoError(t, err)

	joined, err := integration.IsJoinedChannel(channelID, mc.org1ResMgmt, org1Peers[0])
	require.NoError(t, err)
	if !joined {
		createAndJoinChannel(t, &mc)
	}

	testDynamicDiscovery(t, sdk, &mc)

	// now run the same test as multiple_orgs_test.go to make sure it works with bootstrap config..

	// Load specific targets for move funds test
	loadOrgPeers(t, sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1)))

	expectedValue := testWithOrg1(t, sdk, &mc)
	expectedValue = testWithOrg2(t, expectedValue, mc.ccName, channelID)
	verifyWithOrg1(t, sdk, expectedValue, mc.ccName, channelID)
}

func testDynamicDiscovery(t *testing.T, sdk *fabsdk.FabricSDK, mc *multiorgContext) {
	_, err := integration.DiscoverLocalPeers(mc.org1AdminClientContext, 2)
	require.NoError(t, err)
	_, err = integration.DiscoverLocalPeers(mc.org2AdminClientContext, 2)
	require.NoError(t, err)

	// example discovering the peers from the bootstap peer
	// there should be three peers returned from discovery:
	// 1 org1 anchor peer (peer0.org1.example.com)
	// 1 discovered peer (not in config: peer1.org1.example.com)
	// 1 org2 anchor peer (peer0.org2.example.com)
	// 1 discovered peer (not in config: peer1.org2.example.com)
	expectedPeers := 4
	peersList := discoverPeers(t, sdk, expectedPeers)
	assert.Equal(t, expectedPeers, len(peersList), "Expected exactly %d peers as per %s's channel and %s's org configs", expectedPeers, channelID, org2)
}

func discoverPeers(t *testing.T, sdk *fabsdk.FabricSDK, expectedPeers int) []fab.Peer {
	// any user from the network can access the discovery service, user org1User is selected for the test.
	chProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))
	chCtx, err := chProvider()
	require.NoError(t, err, "Error creating channel context")

	chCtx.ChannelService()
	discovery, err := chCtx.ChannelService().Discovery()
	require.NoErrorf(t, err, "Error getting discovery service for channel [%s]", channelID)

	discoveredPeers, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			peers, err := discovery.GetPeers()
			require.NoErrorf(t, err, "Error getting peers for channel [%s]", channelID)

			t.Logf("Peers of channel [%s]:", channelID)
			for i, p := range peers {
				t.Logf("%d- [%s] - MSP [%s]", i, p.URL(), p.MSPID())
			}
			if len(peers) < expectedPeers {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), "didn't get expected number of peers", nil)
			}
			return peers, nil
		},
	)

	require.NoError(t, err)
	return discoveredPeers.([]fab.Peer)
}

// DynamicDiscoveryProviderFactory is configured with dynamic (endorser) selection provider
type DynamicDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateLocalDiscoveryProvider returns a new local dynamic discovery provider
func (f *DynamicDiscoveryProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return dynamicdiscovery.NewLocalProvider(config), nil
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *DynamicDiscoveryProviderFactory) CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config, opts...)
	if err != nil {
		return nil, err
	}
	return &channelProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*dynamicdiscovery.ChannelService),
	}, nil
}

type channelProvider struct {
	fab.ChannelProvider
	services map[string]*dynamicdiscovery.ChannelService
}

type initializer interface {
	Initialize(providers context.Providers) error
}

// Initialize sets the provider context
func (cp *channelProvider) Initialize(providers context.Providers) error {
	init, ok := cp.ChannelProvider.(initializer)
	if ok {
		init.Initialize(providers)
	}
	return nil
}

type channelService struct {
	fab.ChannelService
	discovery fab.DiscoveryService
}

type closable interface {
	Close()
}

// Close frees resources and caches.
func (cp *channelProvider) Close() {
	if c, ok := cp.ChannelProvider.(closable); ok {
		c.Close()
	}
	for _, discovery := range cp.services {
		discovery.Close()
	}
}

// ChannelService creates a ChannelService for an identity
func (cp *channelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	chService, err := cp.ChannelProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, err
	}

	membership, err := chService.Membership()
	if err != nil {
		return nil, err
	}

	discovery, ok := cp.services[channelID]
	if !ok {
		discovery, err = dynamicdiscovery.NewChannelService(ctx, membership, channelID)
		if err != nil {
			return nil, err
		}
		cp.services[channelID] = discovery
	}

	return &channelService{
		ChannelService: chService,
		discovery:      discovery,
	}, nil
}

func (cs *channelService) Discovery() (fab.DiscoveryService, error) {
	return cs.discovery, nil
}

// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package provider

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDynamicDiscovery(t *testing.T) {
	testSetup := mainTestSetup

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithServicePkg(&dynamicDiscoveryProviderFactory{}))
	require.NoError(t, err, "Failed to create new SDK")
	defer sdk.Close()

	err = testSetup.Initialize(sdk)
	require.NoError(t, err, "Failed to initialize test setup")

	chProvider := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	chCtx, err := chProvider()
	require.NoError(t, err, "Error creating channel context")

	discoveryService, err := chCtx.ChannelService().Discovery()
	require.NoError(t, err, "Error creating discovery service")

	peers, err := discoveryService.GetPeers()
	require.NoErrorf(t, err, "Error getting peers for channel [%s]", testSetup.ChannelID)
	require.NotEmptyf(t, peers, "No peers were found for channel [%s]", testSetup.ChannelID)

	t.Logf("Peers of channel [%s]:", testSetup.ChannelID)
	for _, p := range peers {
		t.Logf("- [%s] - MSP [%s]", p.URL(), p.MSPID())
	}

	p0 := peers[0]
	require.NotEmpty(t, p0.Properties())
	require.Less(t, uint64(0), p0.Properties()[fab.PropertyLedgerHeight])
}

func TestDynamicLocalDiscovery(t *testing.T) {
	testSetup := mainTestSetup

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithServicePkg(&dynamicDiscoveryProviderFactory{}))
	require.NoError(t, err, "Failed to create new SDK")
	defer sdk.Close()

	err = testSetup.Initialize(sdk)
	require.NoError(t, err, "Failed to initialize test setup")

	// By default, query for local peers (outside of a channel) requires admin privileges.
	// To bypass this restriction, set peer.discovery.orgMembersAllowedAccess=true in core.yaml.
	ctxProvider := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	locCtx, err := contextImpl.NewLocal(ctxProvider)
	require.NoError(t, err, "Error creating local context")

	peers, err := locCtx.LocalDiscoveryService().GetPeers()
	require.NoErrorf(t, err, "Error getting local peers for MSP [%s]", locCtx.Identifier().MSPID)
	require.NotEmptyf(t, peers, "No local peers were found for MSP [%s]", locCtx.Identifier().MSPID)

	t.Logf("Local peers for MSP [%s]:", locCtx.Identifier().MSPID)
	for _, p := range peers {
		t.Logf("- [%s] - MSP [%s]", p.URL(), p.MSPID())
	}
}

type dynamicDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

type channelProvider struct {
	fab.ChannelProvider
	services map[string]*dynamicdiscovery.ChannelService
}

type channelService struct {
	fab.ChannelService
	discovery fab.DiscoveryService
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *dynamicDiscoveryProviderFactory) CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config, opts...)
	if err != nil {
		return nil, err
	}
	return &channelProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*dynamicdiscovery.ChannelService),
	}, nil
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

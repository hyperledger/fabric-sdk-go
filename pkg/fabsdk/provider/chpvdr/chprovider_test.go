// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"errors"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/fabricselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClientContext struct {
	context.Providers
	msp.SigningIdentity
}

func TestBasicValidChannel(t *testing.T) {
	ctx := mocks.NewMockProviderContext()

	user := mspmocks.NewMockSigningIdentity("user", "user")

	clientCtx := &mockClientContext{
		Providers:       ctx,
		SigningIdentity: user,
	}

	cp, err := New(clientCtx.EndpointConfig())
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %s", err)
	}

	err = cp.Initialize(ctx)
	assert.NoError(t, err)

	testChannelCfg := mocks.NewMockChannelCfg("testchannel")
	testChannelCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true

	SetChannelConfig(chconfig.NewChannelCfg(""), chconfig.NewChannelCfg("mychannel"), testChannelCfg)

	// System channel
	channelService, err := cp.ChannelService(clientCtx, "")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %s", err)
	}

	channelService, err = cp.ChannelService(clientCtx, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %s", err)
	}

	m, err := channelService.Membership()
	assert.Nil(t, err)
	assert.NotNil(t, m)

	chConfig, err := channelService.Config()
	assert.Nil(t, err)
	assert.NotNil(t, chConfig)

	channelConfig, err := channelService.ChannelConfig()
	assert.Nil(t, err)
	assert.NotNil(t, channelConfig)
	assert.NotEmptyf(t, channelConfig.ID(), "Got empty channel ID from channel config")

	eventService, err := channelService.EventService()
	require.NoError(t, err)
	require.NotNil(t, eventService)

	discovery, err := channelService.Discovery()
	require.NoError(t, err)
	require.NotNil(t, discovery)
	_, ok := discovery.(*staticdiscovery.DiscoveryService)
	assert.Truef(t, ok, "Expecting discovery to be Static")

	selection, err := channelService.Selection()
	require.NoError(t, err)
	require.NotNil(t, selection)
	_, ok = selection.(*dynamicselection.SelectionService)
	assert.Truef(t, ok, "Expecting selection to be Dynamic")

	// testchannel has v1_2 capabilities
	channelService, err = cp.ChannelService(clientCtx, "testchannel")
	require.NoError(t, err)
	require.NotNil(t, channelService)
	discovery, err = channelService.Discovery()
	require.NoError(t, err)
	require.NotNil(t, discovery)
	_, ok = discovery.(*dynamicdiscovery.ChannelService)
	assert.Truef(t, ok, "Expecting discovery to be Dynamic for v1_2")

	selection, err = channelService.Selection()
	require.NoError(t, err)
	require.NotNil(t, selection)
	_, ok = selection.(*fabricselection.Service)
	assert.Truef(t, ok, "Expecting selection to be Fabric for v1_2")
}

func TestCloseContext(t *testing.T) {
	testChannelCfg := mocks.NewMockChannelCfg("testchannel")
	testChannelCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true

	SetChannelConfig(chconfig.NewChannelCfg(""), testChannelCfg)

	discClient := clientmocks.NewMockDiscoveryClient()
	dynamicdiscovery.SetClientProvider(func(ctx context.Client) (dynamicdiscovery.DiscoveryClient, error) {
		return discClient, nil
	})

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	channelProvider := getChannelProvider(t, mocks.NewMockProviderContext(),
		dynamicdiscovery.WithRefreshInterval(5*time.Millisecond),
	)
	defer channelProvider.Close()

	clientCtxt1 := newMockClientContext("user1", "org")
	clientCtxt2 := newMockClientContext("user2", "org")

	channelService1, err := channelProvider.ChannelService(clientCtxt1, "testchannel")
	require.NoError(t, err)
	channelService2, err := channelProvider.ChannelService(clientCtxt2, "testchannel")
	require.NoError(t, err)

	discovery1, err := channelService1.Discovery()
	require.NoError(t, err)
	require.NotNil(t, discovery1)

	discovery2, err := channelService2.Discovery()
	require.NoError(t, err)
	require.NotNil(t, discovery2)

	_, err = discovery1.GetPeers()
	require.NoError(t, err)

	_, err = discovery2.GetPeers()
	require.NoError(t, err)

	channelProvider.CloseContext(clientCtxt1)

	// Subsequent calls on the old discovery should fail since the service is closed
	_, err = discovery1.GetPeers()
	require.Error(t, err)
	assert.Equal(t, "Discovery client has been closed", err.Error())

	// Calls should still succeed on the second discovery since it wasn't closed
	_, err = discovery2.GetPeers()
	require.NoError(t, err)
}

func newMockClientContext(userID, mspID string) fab.ClientContext {
	user := mspmocks.NewMockSigningIdentity(userID, mspID)
	return &mockClientContext{
		Providers:       mocks.NewMockProviderContext(),
		SigningIdentity: user,
	}
}

func TestDiscoveryAccessDenied(t *testing.T) {
	discClient, channelService := setupDiscovery(t, func(discClient *clientmocks.MockDiscoveryClient) {
		dynamicdiscovery.SetClientProvider(func(ctx context.Client) (dynamicdiscovery.DiscoveryClient, error) {
			return discClient, nil
		})
	})

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			Error: errors.New("access denied"),
		},
	)

	discovery, err := channelService.Discovery()
	require.NoError(t, err)
	require.NotNil(t, discovery)
	_, ok := discovery.(*dynamicdiscovery.ChannelService)
	assert.Truef(t, ok, "Expecting discovery to be Dynamic for v1_2")

	_, err = discovery.GetPeers()
	require.Error(t, err)
	assert.Equal(t, "access denied", err.Error())

	time.Sleep(50 * time.Millisecond)

	// Subsequent calls should fail since the service is closed
	_, err = discovery.GetPeers()
	require.Error(t, err)
	assert.Equal(t, "Discovery client has been closed due to error: access denied", err.Error())
}

func TestSelectionAccessDenied(t *testing.T) {
	discClient, channelService := setupDiscovery(t, func(discClient *clientmocks.MockDiscoveryClient) {
		fabricselection.SetClientProvider(func(ctx context.Client) (fabricselection.DiscoveryClient, error) {
			return discClient, nil
		})
	})

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			Error: errors.New("access denied"),
		},
	)

	selection, err := channelService.Selection()
	require.NoError(t, err)
	require.NotNil(t, selection)
	_, ok := selection.(*fabricselection.Service)
	assert.Truef(t, ok, "Expecting selection to be Fabric for v1_2")

	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.Error(t, err)
	assert.Equal(t, "error getting channel response for channel [testchannel]: access denied", err.Error())

	time.Sleep(50 * time.Millisecond)

	// Subsequent calls should fail since the service is closed
	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.Error(t, err)
	assert.Equal(t, "Selection service has been closed due to error: access denied", err.Error())
}

func setupDiscovery(t *testing.T, preInit func(discClient *clientmocks.MockDiscoveryClient)) (*clientmocks.MockDiscoveryClient, fab.ChannelService) {
	ctx := mocks.NewMockProviderContext()

	user := mspmocks.NewMockSigningIdentity("user", "user")

	clientCtx := &mockClientContext{
		Providers:       ctx,
		SigningIdentity: user,
	}

	discClient := clientmocks.NewMockDiscoveryClient()

	preInit(discClient)

	cp, err := New(clientCtx.EndpointConfig())
	require.NoError(t, err)

	err = cp.Initialize(ctx)
	assert.NoError(t, err)

	testChannelCfg := mocks.NewMockChannelCfg("testchannel")
	testChannelCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true

	SetChannelConfig(chconfig.NewChannelCfg(""), testChannelCfg)

	channelService, err := cp.ChannelService(clientCtx, "testchannel")
	require.NoError(t, err)

	return discClient, channelService
}

func getChannelProvider(t *testing.T, providers context.Providers, opts ...options.Opt) *ChannelProvider {
	cp, err := New(providers.EndpointConfig(), opts...)
	require.NoError(t, err)

	err = cp.Initialize(providers)
	assert.NoError(t, err)

	return cp
}

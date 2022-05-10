//go:build testing
// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/fabricselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
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

	eventService, err := channelService.EventService(client.WithBlockEvents(), deliverclient.WithSeekType("from"), deliverclient.WithBlockNum(10), deliverclient.WithChaincodeID("testChaincode"))
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

	discClient := discovery.NewMockDiscoveryClient()
	dynamicdiscovery.SetClientProvider(func(ctx context.Client) (discovery.Client, error) {
		return discClient, nil
	})

	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
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
	var channelProvider *ChannelProvider
	var disc fab.DiscoveryService
	var mutex sync.RWMutex

	testChannelCfg := mocks.NewMockChannelCfg("testchannel")
	testChannelCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true

	SetChannelConfig(chconfig.NewChannelCfg(""), testChannelCfg)

	discClient := discovery.NewMockDiscoveryClient()
	dynamicdiscovery.SetClientProvider(func(ctx context.Client) (discovery.Client, error) {
		return discClient, nil
	})

	newDiscovery := func(userID, mspID string) fab.DiscoveryService {
		channelService, err := channelProvider.ChannelService(newMockClientContext(userID, mspID), "testchannel")
		require.NoError(t, err)

		d, err := channelService.Discovery()
		require.NoError(t, err)
		require.NotNil(t, d)

		mutex.Lock()
		defer mutex.Unlock()
		disc = d
		return d
	}

	getDiscovery := func() fab.DiscoveryService {
		mutex.RLock()
		defer mutex.RUnlock()
		return disc
	}

	errHandler := func(ctxt fab.ClientContext, channelID string, err error) {
		if derr, ok := errors.Cause(err).(dynamicdiscovery.DiscoveryError); ok && derr.IsAccessDenied() {
			// Spawn a new Go routine or else we'll hit a deadlock when closing the context
			go func() {
				channelProvider.CloseContext(ctxt)

				// Reset the error
				discClient.SetResponses(
					&discovery.MockDiscoverEndpointResponse{
						PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
					},
				)

				// Replace Discovery with a new one using different credentials
				newDiscovery("user2", "org1")
			}()
		}
	}

	channelProvider = getChannelProvider(t, mocks.NewMockProviderContext(),
		dynamicdiscovery.WithErrorHandler(errHandler),
		dynamicdiscovery.WithRefreshInterval(5*time.Millisecond),
	)
	defer channelProvider.Close()

	discoveryService := newDiscovery("user1", "org1")

	// First set a successful response
	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	_, err := discoveryService.GetPeers()
	require.NoError(t, err)

	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			Error: errors.New("access denied"),
		},
	)

	time.Sleep(10 * time.Millisecond)

	// Subsequent calls on the old discovery should fail since the service is closed
	_, err = discoveryService.GetPeers()
	require.Error(t, err)
	assert.Equal(t, "Discovery client has been closed", err.Error())

	time.Sleep(10 * time.Millisecond)

	// Subsequent calls should succeed since the error handler should have replaced the discovery service
	discoveryService = getDiscovery()
	_, err = discoveryService.GetPeers()
	require.NoError(t, err)

	// Set a transient error
	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			Error: errors.New("some transient error"),
		},
	)

	// Wait for the cache to refresh
	time.Sleep(10 * time.Millisecond)

	// Calls should still succeed since the error handler ignores transient errors
	_, err = discoveryService.GetPeers()
	require.NoError(t, err)
}

func TestSelectionAccessDenied(t *testing.T) {
	var channelProvider *ChannelProvider
	var sel fab.SelectionService
	var mutex sync.RWMutex

	testChannelCfg := mocks.NewMockChannelCfg("testchannel")
	testChannelCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true

	SetChannelConfig(chconfig.NewChannelCfg(""), testChannelCfg)

	discClient := discovery.NewMockDiscoveryClient()
	dynamicdiscovery.SetClientProvider(func(ctx context.Client) (discovery.Client, error) {
		return discClient, nil
	})
	fabricselection.SetClientProvider(func(ctx context.Client) (fabricselection.DiscoveryClient, error) {
		logger.Infof("Returning mock discovery client")
		return discClient, nil
	})

	newSelection := func(userID, mspID string) fab.SelectionService {
		channelService, err := channelProvider.ChannelService(newMockClientContext(userID, mspID), "testchannel")
		require.NoError(t, err)

		s, err := channelService.Selection()
		require.NoError(t, err)
		require.NotNil(t, s)

		mutex.Lock()
		defer mutex.Unlock()
		sel = s
		return s
	}

	getSelection := func() fab.SelectionService {
		mutex.RLock()
		defer mutex.RUnlock()
		return sel
	}

	errHandler := func(ctxt fab.ClientContext, channelID string, err error) {
		if derr, ok := errors.Cause(err).(fabricselection.DiscoveryError); ok && derr.IsAccessDenied() {
			// Spawn a new Go routine or else we'll hit a deadlock when closing the context
			go func() {
				channelProvider.CloseContext(ctxt)

				// Reset the error
				discClient.SetResponses(
					&discovery.MockDiscoverEndpointResponse{
						PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
					},
				)

				// Replace Selection with a new one using different credentials
				newSelection("user2", "org1")
			}()
		}
	}

	channelProvider = getChannelProvider(t, mocks.NewMockProviderContext(),
		dynamicdiscovery.WithErrorHandler(errHandler),
		fabricselection.WithRefreshInterval(3*time.Millisecond),
	)
	defer channelProvider.Close()

	// First set a successful response
	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	selection := newSelection("user1", "org1")

	_, err := selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.NoError(t, err)

	// Now set an error response
	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			Error: errors.New("access denied"),
		},
	)

	// Wait for the cache to refresh
	time.Sleep(10 * time.Millisecond)

	// The old selection service should be closed
	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	assert.EqualError(t, err, "Selection service has been closed")

	// The selection service should have been replaced with a good one
	_, err = getSelection().GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.NoError(t, err)
}

func TestInitButNotConnected(t *testing.T) {
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

	channelService, err := cp.ChannelService(clientCtx, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %s", err)
	}

	eventService, err := channelService.EventService()
	require.NoError(t, err)
	require.NotNil(t, eventService)

	reg, ch, err := eventService.RegisterFilteredBlockEvent()
	require.Error(t, err)
	require.Nil(t, reg)
	require.Nil(t, ch)
}

func getChannelProvider(t *testing.T, providers context.Providers, opts ...options.Opt) *ChannelProvider {
	cp, err := New(providers.EndpointConfig(), opts...)
	require.NoError(t, err)

	err = cp.Initialize(providers)
	assert.NoError(t, err)

	return cp
}

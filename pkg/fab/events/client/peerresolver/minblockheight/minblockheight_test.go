/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package minblockheight

import (
	"math"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testChannel = "testchannel"
	org1MSP     = "Org1MSP"
	p1          = clientmocks.NewMockPeer("peer1", "peer1.example.com:7051", 100)
	p2          = clientmocks.NewMockPeer("peer2", "peer2.example.com:7051", 110)
	p3          = clientmocks.NewMockPeer("peer3", "peer3.example.com:7051", 111)
	peers       = []fab.Peer{p1, p2, p3}
)

func TestFilter(t *testing.T) {
	dispatcher := &clientmocks.MockDispatcher{}
	ctx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("test", org1MSP))

	resolver := New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(-1))
	filteredPeers := resolver.Filter(peers)
	assert.Equal(t, 3, len(filteredPeers))

	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(0))
	filteredPeers = resolver.Filter(peers)
	assert.Equal(t, 1, len(filteredPeers))

	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(5))
	filteredPeers = resolver.Filter(peers)
	assert.Equal(t, 2, len(filteredPeers))

	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(20))
	filteredPeers = resolver.Filter(peers)
	assert.Equal(t, 3, len(filteredPeers))

	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(20))
	resolver.minBlockHeight = 105
	filteredPeers = resolver.Filter(peers)
	assert.Equal(t, 2, len(filteredPeers))

	dispatcher = &clientmocks.MockDispatcher{LastBlock: math.MaxUint64}
	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(20))
	resolver.minBlockHeight = 105
	filteredPeers = resolver.Filter(peers)
	assert.Equal(t, 2, len(filteredPeers))

	dispatcher = &clientmocks.MockDispatcher{LastBlock: 109}
	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(20))
	resolver.minBlockHeight = 105
	filteredPeers = resolver.Filter(peers)
	assert.Equalf(t, 2, len(filteredPeers), "expecting 2 peers to be returned since minBlockHeight is 105")

	dispatcher = &clientmocks.MockDispatcher{LastBlock: 111}
	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(20))
	resolver.minBlockHeight = 112
	filteredPeers = resolver.Filter(peers)
	assert.Equalf(t, 1, len(filteredPeers), "expecting 1 peer to be returned since minBlockHeight was just 1 under the last block received")

	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(20))
	resolver.minBlockHeight = 113
	filteredPeers = resolver.Filter(peers)
	assert.Equalf(t, 3, len(filteredPeers), "Expected all peers to be returned since min block height is unrealistic")
}

func TestResolve(t *testing.T) {
	dispatcher := &clientmocks.MockDispatcher{}
	ctx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("test", org1MSP))

	config := &mocks.MockConfig{}
	config.SetCustomChannelConfig(testChannel, &fab.ChannelEndpointConfig{
		Policies: fab.ChannelPolicies{
			EventService: fab.EventServicePolicy{
				Balancer: fab.RoundRobin,
			},
		},
	})
	ctx.SetEndpointConfig(config)

	resolver := New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(0))
	peer, err := resolver.Resolve(peers)
	require.NoError(t, err)
	assert.Equal(t, p3.URL(), peer.URL())

	resolver = New(dispatcher, ctx, testChannel, WithBlockHeightLagThreshold(-1))

	chosenPeers := make(map[string]struct{})
	for i := 0; i < len(peers); i++ {
		peer, err := resolver.Resolve(peers)
		require.NoError(t, err)
		chosenPeers[peer.URL()] = struct{}{}
	}
	assert.Equalf(t, 3, len(chosenPeers), "expecting all 3 peers to have been chosen")
}

func TestShouldDisconnect(t *testing.T) {
	dispatcher := &clientmocks.MockDispatcher{LastBlock: 100}
	ctx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("test", org1MSP))

	resolver := New(dispatcher, ctx, testChannel, WithReconnectBlockHeightThreshold(120))
	disconnect := resolver.ShouldDisconnect(peers, p1)
	assert.Falsef(t, disconnect, "expecting peer NOT to be disconnected since the reconnectBlockHeightThreshold is greater than the maximum block height")

	resolver = New(dispatcher, ctx, testChannel, WithReconnectBlockHeightThreshold(5))
	disconnect = resolver.ShouldDisconnect(peers, p1)
	assert.Truef(t, disconnect, "expecting peer to be disconnected since its block height is lagging more than 5 blocks behind")

	dispatcher = &clientmocks.MockDispatcher{LastBlock: 98}
	resolver = New(dispatcher, ctx, testChannel, WithReconnectBlockHeightThreshold(5))
	disconnect = resolver.ShouldDisconnect(peers, p1)
	assert.Falsef(t, disconnect, "expecting peer NOT to be disconnected since the last block received is less than the block height of the peer")

	dispatcher = &clientmocks.MockDispatcher{LastBlock: 110}
	resolver = New(dispatcher, ctx, testChannel, WithReconnectBlockHeightThreshold(5))
	disconnect = resolver.ShouldDisconnect(peers, p3)
	assert.Falsef(t, disconnect, "expecting peer NOT to be disconnected since the peer's block height is under the reconnectBlockHeightThreshold")
}

func TestOpts(t *testing.T) {
	channelID := "testchannel"

	config := &mocks.MockConfig{}
	context := mocks.NewMockContext(
		mockmsp.NewMockSigningIdentity("user1", "Org1MSP"),
	)
	context.SetEndpointConfig(config)

	t.Run("Default", func(t *testing.T) {
		config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
			Policies: fab.ChannelPolicies{
				EventService: fab.EventServicePolicy{},
			},
		})

		params := defaultParams(context, channelID)
		require.NotNil(t, params)
		require.NotNil(t, params.loadBalancePolicy)
		assert.Equal(t, defaultBlockHeightLagThreshold, params.blockHeightLagThreshold)
		assert.Equal(t, defaultReconnectBlockHeightLagThreshold, params.reconnectBlockHeightLagThreshold)
	})

	t.Run("ResolveLatest", func(t *testing.T) {
		config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
			Policies: fab.ChannelPolicies{
				EventService: fab.EventServicePolicy{
					MinBlockHeightResolverMode:       fab.ResolveLatest,
					ReconnectBlockHeightLagThreshold: 9,
				},
			},
		})

		params := defaultParams(context, channelID)
		require.NotNil(t, params)
		require.NotNil(t, params.loadBalancePolicy)
		assert.Equal(t, 0, params.blockHeightLagThreshold)
		assert.Equal(t, 9, params.reconnectBlockHeightLagThreshold)
	})

}

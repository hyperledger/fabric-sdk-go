/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package balanced

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	channelID = "testchannel"
	org1MSP   = "Org1MSP"
	p1        = clientmocks.NewMockPeer("peer1", "peer1.example.com:7051", 100)
	p2        = clientmocks.NewMockPeer("peer2", "peer2.example.com:7051", 110)
	p3        = clientmocks.NewMockPeer("peer3", "peer3.example.com:7051", 111)
	peers     = []fab.Peer{p1, p2, p3}
)

func TestResolve(t *testing.T) {
	dispatcher := &clientmocks.MockDispatcher{}
	ctx := mocks.NewMockContext(mockmsp.NewMockSigningIdentity("test", org1MSP))
	config := &mocks.MockConfig{}
	config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
		Policies: fab.ChannelPolicies{
			EventService: fab.EventServicePolicy{
				Balancer: fab.RoundRobin,
			},
		},
	})
	ctx.SetEndpointConfig(config)
	ctx.SetEndpointConfig(config)

	resolver := New(dispatcher, ctx, channelID)

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

	resolver := New(dispatcher, ctx, channelID)
	disconnect := resolver.ShouldDisconnect(peers, p1)
	assert.Falsef(t, disconnect, "expecting peer NOT to be disconnected")
}

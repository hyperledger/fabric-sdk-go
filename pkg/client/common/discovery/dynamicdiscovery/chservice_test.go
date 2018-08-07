/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"testing"
	"time"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	peer1MSP2 = "peer1.org2.com:9999"
)

func TestDiscoveryService(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	config := &config{
		EndpointConfig: mocks.NewMockEndpointConfig(),
		peers: []pfab.ChannelPeer{
			{
				NetworkPeer: pfab.NetworkPeer{
					PeerConfig: pfab.PeerConfig{
						URL: peer1MSP1,
					},
					MSPID: mspID1,
				},
			},
			{
				NetworkPeer: pfab.NetworkPeer{
					PeerConfig: pfab.PeerConfig{
						URL: peer1MSP2,
					},
					MSPID: mspID2,
				},
			},
		},
	}
	ctx.SetEndpointConfig(config)

	discClient := clientmocks.NewMockDiscoveryClient()
	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	clientProvider = func(ctx contextAPI.Client) (discoveryClient, error) {
		return discClient, nil
	}

	service, err := NewChannelService(
		ctx, mocks.NewMockMembership(), ch,
		WithRefreshInterval(500*time.Millisecond),
		WithResponseTimeout(2*time.Second),
	)
	require.NoError(t, err)
	defer service.Close()

	peers, err := service.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(peers))

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     peer1MSP1,
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(1 * time.Second)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equalf(t, 1, len(peers), "Expected 1 peer")

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     peer1MSP1,
					LedgerHeight: 5,
				},
				{
					MSPID:        mspID2,
					Endpoint:     peer1MSP2,
					LedgerHeight: 15,
				},
			},
		},
	)

	time.Sleep(1 * time.Second)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equalf(t, 2, len(peers), "Expected 2 peers")

	filteredService := discovery.NewDiscoveryFilterService(service, &blockHeightFilter{minBlockHeight: 10})
	peers, err = filteredService.GetPeers()
	require.NoError(t, err)
	require.Equalf(t, 1, len(peers), "expecting discovery filter to return only one peer")
}

func TestDiscoveryServiceWithNewOrgJoined(t *testing.T) {

	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))

	config := &config{
		EndpointConfig: mocks.NewMockEndpointConfig(),
		peers: []pfab.ChannelPeer{
			{
				NetworkPeer: pfab.NetworkPeer{
					PeerConfig: pfab.PeerConfig{
						URL: peer1MSP1,
					},
					MSPID: mspID1,
				},
			},
			{
				NetworkPeer: pfab.NetworkPeer{
					PeerConfig: pfab.PeerConfig{
						URL: peer1MSP2,
					},
					MSPID: mspID2,
				},
			},
		},
	}
	ctx.SetEndpointConfig(config)

	discClient := clientmocks.NewMockDiscoveryClient()
	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	clientProvider = func(ctx contextAPI.Client) (discoveryClient, error) {
		return discClient, nil
	}

	service, err := NewChannelService(
		ctx,
		mocks.NewMockMembershipWithMSPFilter([]string{mspID2}),
		ch,
		WithRefreshInterval(500*time.Millisecond),
		WithResponseTimeout(2*time.Second),
	)
	require.NoError(t, err)
	defer service.Close()

	peers, err := service.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(peers))

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     peer1MSP1,
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(1 * time.Second)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equalf(t, 1, len(peers), "Expected 1 peer")

	discClient.SetResponses(
		&clientmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     peer1MSP1,
					LedgerHeight: 5,
				},
				{
					MSPID:        mspID2,
					Endpoint:     peer1MSP2,
					LedgerHeight: 15,
				},
			},
		},
	)

	time.Sleep(1 * time.Second)

	//one of the peer for MSPID2 should be filtered out since it is not yet being updated by memebership cache (ContainsMSP returns false)
	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equalf(t, 1, len(peers), "Expected 1 peer among 2 been discovered, since one of them belong to new org with pending membership update")

	filteredService := discovery.NewDiscoveryFilterService(service, &blockHeightFilter{minBlockHeight: 10})
	peers, err = filteredService.GetPeers()
	require.NoError(t, err)
	require.Equalf(t, 0, len(peers), "expecting discovery filter to return only one peer")

}

func TestPickRandomNPeerConfigs(t *testing.T) {
	counter := 20
	allChPeers := createNChannelPeers(counter)

	result := pickRandomNPeerConfigs(allChPeers, 4)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 4, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 1)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 19)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 19, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 20)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 20, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 21)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 20, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 24)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 20, len(result))
	verifyDuplicates(t, result)

	counter = 7
	allChPeers = createNChannelPeers(counter)

	result = pickRandomNPeerConfigs(allChPeers, 6)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 6, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 7)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 7, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 8)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 7, len(result))
	verifyDuplicates(t, result)

	counter = 2
	allChPeers = createNChannelPeers(counter)

	result = pickRandomNPeerConfigs(allChPeers, 2)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 2, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 24)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 2, len(result))
	verifyDuplicates(t, result)

	counter = 1
	allChPeers = createNChannelPeers(counter)

	result = pickRandomNPeerConfigs(allChPeers, 1)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 2)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

	result = pickRandomNPeerConfigs(allChPeers, 24)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

}

func createNChannelPeers(n int) []pfab.ChannelPeer {
	allChPeers := make([]pfab.ChannelPeer, n)
	for i := 0; i < n; i++ {
		allChPeers[i] = pfab.ChannelPeer{
			NetworkPeer: pfab.NetworkPeer{
				PeerConfig: pfab.PeerConfig{URL: fmt.Sprintf("URL-%d", i)},
			},
		}
	}
	return allChPeers
}

func verifyDuplicates(t *testing.T, chPeers []pfab.PeerConfig) {
	seen := make(map[string]bool)
	for _, v := range chPeers {
		if seen[v.URL] {
			t.Fatalf("found duplicate channel peer: %s", v.URL)
		}
		seen[v.URL] = true
	}
}

type blockHeightFilter struct {
	minBlockHeight uint64
}

func (f *blockHeightFilter) Accept(peer pfab.Peer) bool {
	if p, ok := peer.(pfab.PeerState); ok {
		return p.BlockHeight() >= f.minBlockHeight
	}
	panic("expecting peer to have state")
}

// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/pkg/errors"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabDiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	discClient := fabDiscovery.NewMockDiscoveryClient()
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
			Target:        peer1MSP1,
		},
	)

	SetClientProvider(func(ctx contextAPI.Client) (fabDiscovery.Client, error) {
		return discClient, nil
	})

	var service *ChannelService
	service, err := NewChannelService(
		ctx, mocks.NewMockMembership(), ch,
		WithRefreshInterval(10*time.Millisecond),
		WithResponseTimeout(100*time.Millisecond),
		WithErrorHandler(
			func(ctx fab.ClientContext, channelID string, err error) {
				derr, ok := errors.Cause(err).(DiscoveryError)

				if ok {
					//peer1MSP1 or peer1MSP2, depending on request
					assert.NotEmpty(t, derr.Target())
					assert.NotEmpty(t, derr.Error())

					if derr.IsAccessDenied() {
						service.Close()
					}
				}
			},
		),
	)
	require.NoError(t, err)
	defer service.Close()

	peers, err := service.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(peers))

	chaincodes := []*gossip.Chaincode{
		{
			Name:    "cc1",
			Version: "v1",
		},
	}

	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     peer1MSP1,
					LedgerHeight: 5,
					Chaincodes:   chaincodes,
					LeftChannel:  false,
				},
			},
			Target: peer1MSP2,
		},
	)

	time.Sleep(20 * time.Millisecond)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equalf(t, 1, len(peers), "Expected 1 peer")

	peer := peers[0]
	require.NotEmpty(t, peer.Properties())
	require.Equal(t, uint64(5), peer.Properties()[fab.PropertyLedgerHeight])
	require.Equal(t, false, peer.Properties()[fab.PropertyLeftChannel])
	require.Equal(t, chaincodes, peer.Properties()[fab.PropertyChaincodes])

	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
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
			Target: peer1MSP1,
		},
	)

	time.Sleep(20 * time.Millisecond)

	peers, err = service.GetPeers()
	require.NoError(t, err)
	assert.Equalf(t, 2, len(peers), "Expected 2 peers")

	filteredService := discovery.NewDiscoveryFilterService(service, &blockHeightFilter{minBlockHeight: 10})
	peers, err = filteredService.GetPeers()
	require.NoError(t, err)
	require.Equalf(t, 1, len(peers), "expecting discovery filter to return only one peer")

	// Non-fatal error
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			Error:  errors.New("some transient error"),
			Target: peer1MSP1,
		},
	)

	time.Sleep(50 * time.Millisecond)

	// GetPeers should return the cached response
	peers, err = service.GetPeers()
	require.NoError(t, err)
	assert.Equalf(t, 2, len(peers), "Expected 2 peers")

	// Fatal error (access denied can be due a user being revoked)
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			Error:  errors.New(AccessDenied),
			Target: peer1MSP1,
		},
	)

	time.Sleep(50 * time.Millisecond)

	// The discovery service should have been closed
	_, err = service.GetPeers()
	require.Error(t, err)
	assert.Equal(t, "Discovery client has been closed", err.Error())

	ctx = mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	ctx.SetEndpointConfig(mocks.NewMockEndpointConfig())

	service, err = NewChannelService(ctx, mocks.NewMockMembership(), "noChannelPeers")
	require.NoError(t, err)

	_, err = service.GetPeers()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no channel peers configured for channel [noChannelPeers]")
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

	discClient := fabDiscovery.NewMockDiscoveryClient()
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	SetClientProvider(func(ctx contextAPI.Client) (fabDiscovery.Client, error) {
		return discClient, nil
	})

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
		&fabDiscovery.MockDiscoverEndpointResponse{
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
		&fabDiscovery.MockDiscoverEndpointResponse{
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

type blockHeightFilter struct {
	minBlockHeight uint64
}

func (f *blockHeightFilter) Accept(peer pfab.Peer) bool {
	if p, ok := peer.(pfab.PeerState); ok {
		return p.BlockHeight() >= f.minBlockHeight
	}
	panic("expecting peer to have state")
}

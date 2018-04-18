/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"testing"
	"time"

	dyndiscmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery/mocks"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
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

	discClient := dyndiscmocks.NewMockDiscoveryClient()
	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	clientProvider = func(ctx contextAPI.Client) (discoveryClient, error) {
		return discClient, nil
	}

	membershipService := newChannelService(
		options{
			refreshInterval: 500 * time.Millisecond,
			responseTimeout: 2 * time.Second,
		},
	)
	defer membershipService.Close()

	chCtx := mocks.NewMockChannelContext(ctx, ch)
	err := membershipService.Initialize(chCtx)
	assert.NoError(t, err)
	// Initialize again should produce no error
	err = membershipService.Initialize(chCtx)
	assert.NoError(t, err)

	peers, err := membershipService.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(peers))

	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
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

	peers, err = membershipService.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(peers))

	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     peer1MSP1,
					LedgerHeight: 5,
				},
				{
					MSPID:        mspID2,
					Endpoint:     peer1MSP2,
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(1 * time.Second)

	peers, err = membershipService.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(peers))
}

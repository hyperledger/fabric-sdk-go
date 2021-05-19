// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabDiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDiscoveryServiceEndpointsEvaluation(t *testing.T) {

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
		},
	}
	ctx.SetEndpointConfig(config)

	discClient := fabDiscovery.NewMockDiscoveryClient()

	SetClientProvider(func(ctx contextAPI.Client) (fabDiscovery.Client, error) {
		return discClient, nil
	})

	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

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
				{
					MSPID:        mspID1,
					Endpoint:     "missing",
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(1 * time.Second)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equalf(t, 2, len(peers), "expected 2 peers")
}

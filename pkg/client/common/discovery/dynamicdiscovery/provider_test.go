/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"testing"
	"time"

	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

const (
	ch = "orgchannel"

	mspID1 = "Org1MSP"
	mspID2 = "Org2MSP"

	peer1MSP1 = "peer1.org1.com:9999"
)

func TestDiscoveryProvider(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	config := &config{
		EndpointConfig: mocks.NewMockEndpointConfig(),
		peers: []pfab.ChannelPeer{
			{
				NetworkPeer: pfab.NetworkPeer{
					PeerConfig: pfab.PeerConfig{
						URL: peer1MSP1,
					},
				},
			},
		},
	}
	ctx.SetEndpointConfig(config)

	p := New(config, WithRefreshInterval(30*time.Second), WithResponseTimeout(10*time.Second))
	defer p.Close()

	service, err := p.CreateDiscoveryService(ch)
	assert.NoError(t, err)

	chCtx := mocks.NewMockChannelContext(ctx, ch)

	err = service.(*channelService).Initialize(chCtx)
	assert.NoError(t, err)

	localService, err := p.CreateLocalDiscoveryService()
	assert.NoError(t, err)

	localCtx := mocks.NewMockLocalContext(ctx, nil)
	err = localService.(*LocalService).Initialize(localCtx)
	assert.NoError(t, err)
}

type config struct {
	pfab.EndpointConfig
	peers []pfab.ChannelPeer
}

func (c *config) ChannelPeers(name string) ([]pfab.ChannelPeer, error) {
	return c.peers, nil
}

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

func TestLocalProvider(t *testing.T) {
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

	p := NewLocalProvider(config, WithRefreshInterval(30*time.Second), WithResponseTimeout(10*time.Second))
	defer p.Close()

	localService1, err := p.CreateLocalDiscoveryService(mspID1)
	assert.NoError(t, err)

	localCtx := mocks.NewMockLocalContext(ctx, nil)
	err = localService1.(*LocalService).Initialize(localCtx)
	assert.NoError(t, err)

	localService2, err := p.CreateLocalDiscoveryService(mspID1)
	assert.NoError(t, err)
	assert.Equal(t, localService1, localService2)

	localService2, err = p.CreateLocalDiscoveryService(mspID2)
	assert.NoError(t, err)
	assert.NotEqual(t, localService1, localService2)
}

type config struct {
	pfab.EndpointConfig
	peers []pfab.ChannelPeer
}

func (c *config) ChannelPeers(name string) ([]pfab.ChannelPeer, bool) {
	if len(c.peers) == 0 {
		return nil, false
	}
	return c.peers, true
}

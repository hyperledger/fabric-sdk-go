// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

const (
	ch = "orgchannel"

	mspID1 = "Org1MSP"
	mspID2 = "Org2MSP"

	peer1MSP1 = "peer1.org1.com:9999"
	peer1MSP2 = "peer1.org2.com:9999"
)

func TestLocalProvider(t *testing.T) {
	config := &mocks.MockConfig{}
	peer1Org1 := pfab.NetworkPeer{
		PeerConfig: pfab.PeerConfig{
			URL: peer1MSP1,
		},
		MSPID: mspID1,
	}
	peer1Org2 := pfab.NetworkPeer{
		PeerConfig: pfab.PeerConfig{
			URL: peer1MSP2,
		},
		MSPID: mspID2,
	}
	config.SetCustomNetworkPeerCfg([]pfab.NetworkPeer{peer1Org1, peer1Org2})

	discClient := discovery.NewMockDiscoveryClient()
	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	SetClientProvider(func(ctx contextAPI.Client) (discovery.Client, error) {
		return discClient, nil
	})

	ctx1 := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	ctx1.SetEndpointConfig(config)
	localCtx1 := mocks.NewMockLocalContext(ctx1, nil)

	ctx2 := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID2))
	ctx2.SetEndpointConfig(config)
	localCtx2 := mocks.NewMockLocalContext(ctx2, nil)

	p := NewLocalProvider(config, WithRefreshInterval(30*time.Second), WithResponseTimeout(10*time.Second))
	defer p.Close()

	localService1, err := p.CreateLocalDiscoveryService(mspID1)
	assert.NoError(t, err)

	err = localService1.(*LocalService).Initialize(localCtx1)
	assert.NoError(t, err)

	localService2, err := p.CreateLocalDiscoveryService(mspID1)
	assert.NoError(t, err)
	assert.Equal(t, localService1, localService2)

	localService2, err = p.CreateLocalDiscoveryService(mspID2)
	assert.NoError(t, err)
	assert.NotEqual(t, localService1, localService2)

	err = localService2.(*LocalService).Initialize(localCtx2)
	assert.NoError(t, err)

	_, err = localService1.GetPeers()
	assert.NoError(t, err)

	_, err = localService2.GetPeers()
	assert.NoError(t, err)

	p.CloseContext(localCtx1)

	_, err = localService1.GetPeers()
	assert.EqualError(t, err, "Discovery client has been closed")

	_, err = localService2.GetPeers()
	assert.NoError(t, err)

}

type config struct {
	pfab.EndpointConfig
	peers []pfab.ChannelPeer
}

func (c *config) ChannelPeers(name string) []pfab.ChannelPeer {
	return c.peers
}

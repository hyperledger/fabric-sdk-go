/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
)

const (
	url1 = "p1.test.com:9051"
	url2 = "p2.test.com:9051"
	url3 = "p3.test.com:9051"
)

var p1 = fabmocks.NewMockPeer("p1", url1)
var p2 = fabmocks.NewMockPeer("p2", url2)
var p3 = fabmocks.NewMockPeer("p3", url3)

var peers = []fab.Peer{p1, p2, p3}

func TestEndpoint(t *testing.T) {
	expectedAllowInsecure := true
	expectedFailfast := true
	expectedKeepAliveTime := time.Second
	expectedKeepAliveTimeout := time.Second
	expectedKeepAlivePermit := true
	expectedNumOpts := 6

	config := fabmocks.NewMockEndpointConfig()
	peer := fabmocks.NewMockPeer("p1", "localhost:7051")
	peerConfig := &fab.PeerConfig{
		GRPCOptions: make(map[string]interface{}),
	}
	peerConfig.GRPCOptions["allow-insecure"] = expectedAllowInsecure
	peerConfig.GRPCOptions["fail-fast"] = expectedFailfast
	peerConfig.GRPCOptions["keep-alive-time"] = expectedKeepAliveTime
	peerConfig.GRPCOptions["keep-alive-timeout"] = expectedKeepAliveTimeout
	peerConfig.GRPCOptions["keep-alive-permit"] = expectedKeepAlivePermit

	endpoint := FromPeerConfig(config, peer, peerConfig)

	opts := endpoint.Opts()
	if len(opts) != expectedNumOpts {
		t.Fatalf("expecting number of options returned to be %d but got %d", expectedNumOpts, len(opts))
	}
}

func TestDiscoveryProvider(t *testing.T) {
	ctx := newMockContext()

	expectedNumPeers := len(peers)

	discoveryService, err := NewEndpointDiscoveryWrapper(ctx, "testchannel", clientmocks.NewDiscoveryService(peers...))
	require.NoError(t, err, "error creating discovery wrapper")

	peers, err = discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("error getting peers: %s", err)
	}
	if len(peers) != expectedNumPeers {
		t.Fatalf("expecting %d peers but got %d", expectedNumPeers, len(peers))
	}
}

func TestDiscoveryProviderWithTargetFilter(t *testing.T) {
	ctx := newMockContext()

	expectedNumPeers := len(peers) - 1

	discoveryService, err := NewEndpointDiscoveryWrapper(ctx, "testchannel", clientmocks.NewDiscoveryService(peers...), WithTargetFilter(newMockFilter(p3)))
	assert.NoError(t, err, "error creating discovery wrapper")

	peers, err = discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("error getting peers: %s", err)
	}
	if len(peers) != expectedNumPeers {
		t.Fatalf("expecting %d peers but got %d", expectedNumPeers, len(peers))
	}
}

func TestDiscoveryProviderWithEventSource(t *testing.T) {
	ctx := newMockContext()

	chPeer2 := fab.ChannelPeer{}
	chPeer2.URL = p2.URL()
	chPeer2.EventSource = false
	ctx.SetEndpointConfig(newMockConfig(chPeer2))

	expectedNumPeers := len(peers) - 1

	discoveryService, err := NewEndpointDiscoveryWrapper(ctx, "testchannel", clientmocks.NewDiscoveryService(peers...))
	assert.NoError(t, err, "error creating discovery wrapper")

	peers, err = discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("error getting peers: %s", err)
	}
	if len(peers) != expectedNumPeers {
		t.Fatalf("expecting %d peers but got %d", expectedNumPeers, len(peers))
	}
}

type mockConfig struct {
	fab.EndpointConfig
	channelPeers []fab.ChannelPeer
}

func newMockConfig(channelPeers ...fab.ChannelPeer) *mockConfig {
	return &mockConfig{
		EndpointConfig: fabmocks.NewMockEndpointConfig(),
		channelPeers:   channelPeers,
	}
}

func (c *mockConfig) ChannelPeers(name string) []fab.ChannelPeer {
	test.Logf("mockConfig.ChannelPeers [%#v]", c.channelPeers)
	return c.channelPeers
}

func newMockContext() *fabmocks.MockContext {
	ctx := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)

	chPeer := fab.ChannelPeer{}
	chPeer.URL = p2.URL()
	chPeer.EventSource = true

	ctx.SetEndpointConfig(newMockConfig(chPeer))
	return ctx
}

type mockFilter struct {
	excludePeers []fab.Peer
}

func newMockFilter(excludePeers ...fab.Peer) *mockFilter {
	return &mockFilter{excludePeers: excludePeers}
}

func (f *mockFilter) Accept(peer fab.Peer) bool {
	for _, p := range f.excludePeers {
		if p.URL() == peer.URL() {
			return false
		}
	}
	return true
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	url1 = "p1.test.com:9051"
	url2 = "p2.test.com:9051"
	url3 = "p3.test.com:9051"
)

var p1 = fabmocks.NewMockPeer("p1", url1)
var p2 = fabmocks.NewMockPeer("p2", url2)
var p3 = fabmocks.NewMockPeer("p3", url3)

var pc1 = fab.PeerConfig{URL: url1}
var pc2 = fab.PeerConfig{URL: url2}
var pc3 = fab.PeerConfig{URL: url3}

var peers = []fab.Peer{p1, p2, p3}
var peerConfigs = []fab.PeerConfig{pc1, pc2, pc3}

func TestEndpoint(t *testing.T) {
	expectedEventURL := "localhost:7053"
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
		EventURL:    "localhost:7053",
	}
	peerConfig.GRPCOptions["allow-insecure"] = expectedAllowInsecure
	peerConfig.GRPCOptions["fail-fast"] = expectedFailfast
	peerConfig.GRPCOptions["keep-alive-time"] = expectedKeepAliveTime
	peerConfig.GRPCOptions["keep-alive-timeout"] = expectedKeepAliveTimeout
	peerConfig.GRPCOptions["keep-alive-permit"] = expectedKeepAlivePermit

	endpoint, err := FromPeerConfig(config, peer, peerConfig)
	if err != nil {
		t.Fatalf("unexpected error from peer config: %s", err)
	}

	if endpoint.EventURL() != expectedEventURL {
		t.Fatalf("expecting eventURL %s but got %s", expectedEventURL, endpoint.EventURL())
	}

	opts := endpoint.Opts()
	if len(opts) != expectedNumOpts {
		t.Fatalf("expecting number of options returned to be %d but got %d", expectedNumOpts, len(opts))
	}
}

func TestDiscoveryProvider(t *testing.T) {
	ctx := newMockContext()

	expectedNumPeers := len(peers)

	discoveryProvider := NewDiscoveryProvider(ctx)

	discoveryService, err := discoveryProvider.CreateDiscoveryService("testchannel")
	if err != nil {
		t.Fatalf("error creating discovery service: %s", err)
	}
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

	discoveryProvider := NewDiscoveryProvider(ctx, WithTargetFilter(newMockFilter(p3)))

	discoveryService, err := discoveryProvider.CreateDiscoveryService("testchannel")
	if err != nil {
		t.Fatalf("error creating discovery service: %s", err)
	}
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

	discoveryProvider := NewDiscoveryProvider(ctx)

	discoveryService, err := discoveryProvider.CreateDiscoveryService("testchannel")
	if err != nil {
		t.Fatalf("error creating discovery service: %s", err)
	}
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

func (c *mockConfig) PeerConfigByURL(url string) (*fab.PeerConfig, error) {
	for _, pc := range peerConfigs {
		if pc.URL == url {
			return &pc, nil
		}
	}
	return nil, nil
}

func (c *mockConfig) ChannelPeers(name string) ([]fab.ChannelPeer, error) {
	fmt.Printf("mockConfig.ChannelPeers - returning %#v", c.channelPeers)
	return c.channelPeers, nil
}

func newMockContext() *fabmocks.MockContext {
	discoveryProvider := fabmocks.NewMockDiscoveryProvider(nil, peers)

	ctx := fabmocks.NewMockContextWithCustomDiscovery(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		discoveryProvider,
	)
	ctx.SetEndpointConfig(newMockConfig())
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

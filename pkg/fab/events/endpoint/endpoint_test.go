/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/mocks"
)

func TestEndpoint(t *testing.T) {
	expectedEventURL := "localhost:7053"
	expectedAllowInsecure := true
	expectedFailfast := true
	expectedKeepAliveTime := time.Second
	expectedKeepAliveTimeout := time.Second
	expectedKeepAlivePermit := true
	expectedNumOpts := 6

	config := fabmocks.NewMockConfig()
	peer := fabmocks.NewMockPeer("p1", "localhost:7051")
	peerConfig := &core.PeerConfig{
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
	if endpoint.AllowInsecure != expectedAllowInsecure {
		t.Fatalf("expecting allowInsecure %t but got %t", expectedAllowInsecure, endpoint.AllowInsecure)
	}
	if endpoint.FailFast != expectedFailfast {
		t.Fatalf("expecting failFast %t but got %t", expectedFailfast, endpoint.FailFast)
	}
	if endpoint.KeepAliveParams.Time != expectedKeepAliveTime {
		t.Fatalf("expecting keepAliveParams.Time %s but got %s", expectedKeepAliveTime, endpoint.KeepAliveParams.Time)
	}
	if endpoint.KeepAliveParams.Timeout != expectedKeepAliveTimeout {
		t.Fatalf("expecting keepAliveParams.Timeout %s but got %s", expectedKeepAliveTimeout, endpoint.KeepAliveParams.Timeout)
	}
	if endpoint.KeepAliveParams.PermitWithoutStream != expectedKeepAlivePermit {
		t.Fatalf("expecting keepAliveParams.PermitWithoutStream %t but got %t", expectedKeepAlivePermit, endpoint.KeepAliveParams.PermitWithoutStream)
	}

	opts := endpoint.Opts()
	if len(opts) != expectedNumOpts {
		t.Fatalf("expecting number of options returned to be %d but got %d", expectedNumOpts, len(opts))
	}
}

func TestDiscoveryProvider(t *testing.T) {
	ctx := newMockContext()
	discoveryProvider := NewDiscoveryProvider(ctx)

	discoveryService, err := discoveryProvider.CreateDiscoveryService("testchannel")
	if err != nil {
		t.Fatalf("error creating discovery service: %s", err)
	}
	_, err = discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("error getting peers: %s", err)
	}

}

func TestDiscoveryProviderWithTargetFilter(t *testing.T) {
	ctx := newMockContext()

	var numTimesCalled int
	expectedNumTimesCalled := 1

	discoveryProvider := NewDiscoveryProvider(ctx, WithTargetFilter(newMockFilter(&numTimesCalled)))

	discoveryService, err := discoveryProvider.CreateDiscoveryService("testchannel")
	if err != nil {
		t.Fatalf("error creating discovery service: %s", err)
	}
	_, err = discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("error getting peers: %s", err)
	}
	if numTimesCalled != expectedNumTimesCalled {
		t.Fatalf("expecting target filter to be called %d time(s) but was called %d time(s)", expectedNumTimesCalled, numTimesCalled)
	}
}

type mockConfig struct {
	core.Config
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		Config: fabmocks.NewMockConfig(),
	}
}

func (c *mockConfig) PeerConfigByURL(url string) (*core.PeerConfig, error) {
	return &core.PeerConfig{}, nil
}

func newMockContext() *fabmocks.MockContext {
	ctx := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)
	ctx.SetConfig(newMockConfig())
	return ctx
}

type mockFilter struct {
	numTimesCalled *int
}

func newMockFilter(numTimesCalled *int) *mockFilter {
	return &mockFilter{numTimesCalled: numTimesCalled}
}

func (f *mockFilter) Accept(peer fab.Peer) bool {
	*f.numTimesCalled++
	return true
}

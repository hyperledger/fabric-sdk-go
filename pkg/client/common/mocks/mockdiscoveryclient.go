/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqcontext "context"
	"sync"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	gprotoext "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/gossip/protoext"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/gossip"
)

// MockDiscoveryClient implements a mock Discover service
type MockDiscoveryClient struct {
	resp []fabdiscovery.Response
	lock sync.RWMutex
}

// MockResponseBuilder builds a mock discovery response
type MockResponseBuilder interface {
	Build() fabdiscovery.Response
}

// NewMockDiscoveryClient returns a new mock Discover service
func NewMockDiscoveryClient() *MockDiscoveryClient {
	return &MockDiscoveryClient{}
}

// Send sends a Discovery request
func (m *MockDiscoveryClient) Send(ctx reqcontext.Context, req *discclient.Request, targets ...fab.PeerConfig) ([]fabdiscovery.Response, error) {
	return m.responses(), nil
}

// SetResponses sets the responses that the mock client should return from the Send function
func (m *MockDiscoveryClient) SetResponses(responses ...MockResponseBuilder) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.resp = nil

	for _, resp := range responses {
		m.resp = append(m.resp, resp.Build())
	}
}

func (m *MockDiscoveryClient) responses() []fabdiscovery.Response {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.resp
}

type mockDiscoverResponse struct {
	discclient.Response
	target string
}

func (r *mockDiscoverResponse) Target() string {
	return r.target
}

type response struct {
	peers []*discclient.Peer
	err   error
}

func (r *response) ForChannel(string) discclient.ChannelResponse {
	return &channelResponse{
		peers: r.peers,
		err:   r.err,
	}
}

func (r *response) ForLocal() discclient.LocalResponse {
	return &localResponse{
		peers: r.peers,
		err:   r.err,
	}
}

type channelResponse struct {
	peers discclient.Endorsers
	err   error
}

// Config returns a response for a config query, or error if something went wrong
func (cr *channelResponse) Config() (*discovery.ConfigResult, error) {
	panic("not implemented")
}

// Peers returns a response for a peer membership query, or error if something went wrong
func (cr *channelResponse) Peers(invocationChain ...*discovery.ChaincodeCall) ([]*discclient.Peer, error) {
	return cr.peers, cr.err
}

// Endorsers returns the response for an endorser query
func (cr *channelResponse) Endorsers(invocationChain discclient.InvocationChain, f discclient.Filter) (discclient.Endorsers, error) {
	if cr.err != nil {
		return nil, cr.err
	}
	return f.Filter(cr.peers), nil
}

type localResponse struct {
	peers []*discclient.Peer
	err   error
}

// Peers returns a response for a peer membership query, or error if something went wrong
func (cr *localResponse) Peers() ([]*discclient.Peer, error) {
	return cr.peers, cr.err
}

// MockDiscoverEndpointResponse contains a mock response for the discover client
type MockDiscoverEndpointResponse struct {
	Target        string
	PeerEndpoints []*discmocks.MockDiscoveryPeerEndpoint
	Error         error
}

// Build builds a mock discovery response
func (b *MockDiscoverEndpointResponse) Build() fabdiscovery.Response {
	var peers discclient.Endorsers
	for _, endpoint := range b.PeerEndpoints {
		peer := &discclient.Peer{
			MSPID:            endpoint.MSPID,
			AliveMessage:     newAliveMessage(endpoint),
			StateInfoMessage: newStateInfoMessage(endpoint),
		}
		peers = append(peers, peer)
	}
	return &mockDiscoverResponse{
		Response: &response{
			peers: peers,
			err:   b.Error,
		},
		target: b.Target,
	}
}

func newAliveMessage(endpoint *discmocks.MockDiscoveryPeerEndpoint) *gprotoext.SignedGossipMessage {
	return &gprotoext.SignedGossipMessage{
		GossipMessage: &gossip.GossipMessage{
			Content: &gossip.GossipMessage_AliveMsg{
				AliveMsg: &gossip.AliveMessage{
					Membership: &gossip.Member{
						Endpoint: endpoint.Endpoint,
					},
				},
			},
		},
	}
}

func newStateInfoMessage(endpoint *discmocks.MockDiscoveryPeerEndpoint) *gprotoext.SignedGossipMessage {
	return &gprotoext.SignedGossipMessage{
		GossipMessage: &gossip.GossipMessage{
			Content: &gossip.GossipMessage_StateInfo{
				StateInfo: &gossip.StateInfo{
					Properties: &gossip.Properties{
						LedgerHeight: endpoint.LedgerHeight,
					},
				},
			},
		},
	}
}

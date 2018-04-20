/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqcontext "context"
	"sync"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/gossip"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
)

// MockDiscoveryClient implements a mock Discover service
type MockDiscoveryClient struct {
	resp []fabdiscovery.Response
	lock sync.RWMutex
}

// MockDiscoverEndpointResponse contains a mock response for the discover client
type MockDiscoverEndpointResponse struct {
	Target        string
	PeerEndpoints []*discmocks.MockDiscoveryPeerEndpoint
	Error         error
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
func (m *MockDiscoveryClient) SetResponses(responses ...*MockDiscoverEndpointResponse) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.resp = nil

	for _, resp := range responses {
		var peers []*discclient.Peer
		for _, endpoint := range resp.PeerEndpoints {
			peer := &discclient.Peer{
				MSPID:            endpoint.MSPID,
				AliveMessage:     newAliveMessage(endpoint),
				StateInfoMessage: newStateInfoMessage(endpoint),
			}
			peers = append(peers, peer)
		}
		m.resp = append(m.resp, &mockDiscoverResponse{
			Response: &response{peers: peers}, target: resp.Target, err: resp.Error,
		})
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
	err    error
}

func (r *mockDiscoverResponse) Target() string {
	return r.target
}

func (r *mockDiscoverResponse) Error() error {
	return r.err
}

type response struct {
	peers []*discclient.Peer
}

func (r *response) ForChannel(string) discclient.ChannelResponse {
	return &channelResponse{
		peers: r.peers,
	}
}

func (r *response) ForLocal() discclient.LocalResponse {
	return &localResponse{
		peers: r.peers,
	}
}

type channelResponse struct {
	peers []*discclient.Peer
}

// Config returns a response for a config query, or error if something went wrong
func (cr *channelResponse) Config() (*discovery.ConfigResult, error) {
	panic("not implemented")
}

// Peers returns a response for a peer membership query, or error if something went wrong
func (cr *channelResponse) Peers() ([]*discclient.Peer, error) {
	return cr.peers, nil
}

// Endorsers returns the response for an endorser query
func (cr *channelResponse) Endorsers(cc string, ps discclient.PrioritySelector, ef discclient.ExclusionFilter) (discclient.Endorsers, error) {
	panic("not implemented")
}

type localResponse struct {
	peers []*discclient.Peer
}

// Peers returns a response for a peer membership query, or error if something went wrong
func (cr *localResponse) Peers() ([]*discclient.Peer, error) {
	return cr.peers, nil
}

func newAliveMessage(endpoint *discmocks.MockDiscoveryPeerEndpoint) *gossip.SignedGossipMessage {
	return &gossip.SignedGossipMessage{
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

func newStateInfoMessage(endpoint *discmocks.MockDiscoveryPeerEndpoint) *gossip.SignedGossipMessage {
	return &gossip.SignedGossipMessage{
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

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"context"
	"sync"

	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric-protos-go/gossip"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	gprotoext "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/gossip/protoext"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/pkg/errors"
)

// MockDiscoveryClient implements a mock Discover service
type MockDiscoveryClient struct {
	resp []Response
	lock sync.RWMutex
}

// MockResponseBuilder builds a mock discovery response
type MockResponseBuilder interface {
	Build() Response
}

// NewMockDiscoveryClient returns a new mock Discover service
func NewMockDiscoveryClient() *MockDiscoveryClient {
	return &MockDiscoveryClient{}
}

// Send sends a Discovery request
func (m *MockDiscoveryClient) Send(ctx context.Context, req *Request, targets ...fab.PeerConfig) (<-chan Response, error) {
	respCh := make(chan Response, len(targets))

	for _, r := range m.responses() {
		respCh <- r
	}

	close(respCh)

	return respCh, nil
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

func (m *MockDiscoveryClient) responses() []Response {
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

type fakeResponse struct {
	peers        []*discclient.Peer
	err          error
	accessDenied bool
}

func (r *fakeResponse) ForChannel(string) discclient.ChannelResponse {
	chanResp := &channelResponse{
		peers: r.peers,
		err:   r.err,
	}

	//usually "access denied" is a successful response.
	//The problem is that fakeResponse struct contains only peers arr and not message with content, so need to add bool if access denied
	//see origins of https://github.com/hyperledger/fabric-sdk-go/pull/62
	if r.accessDenied {
		chanResp.err = errors.New("access denied")
	}

	return chanResp
}

func (r *fakeResponse) ForLocal() discclient.LocalResponse {
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

	for _, call := range invocationChain {
		if call.Name == "notInstalledToAnyPeer" {
			return nil, errors.New("no endorsement combination can be satisfied")
		}
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
	PeerEndpoints []*mocks.MockDiscoveryPeerEndpoint
	Error         error
	//todo it's really hard to make an expectation in tests, maybe we need to switch to mocks because current thing is stub
	AccessDenied bool
}

// Build builds a mock discovery response
func (b *MockDiscoverEndpointResponse) Build() Response {
	var peers discclient.Endorsers
	for _, endpoint := range b.PeerEndpoints {
		peer := &discclient.Peer{
			MSPID:            endpoint.MSPID,
			AliveMessage:     newAliveMessage(endpoint),
			StateInfoMessage: newStateInfoMessage(endpoint),
		}
		peers = append(peers, peer)
	}

	disResp := &fakeResponse{
		peers: peers,
		err:   b.Error,
	}

	if b.AccessDenied {
		disResp.accessDenied = true
	}

	return &mockDiscoverResponse{
		Response: disResp,
		target:   b.Target,
		err:      b.Error,
	}
}

func newAliveMessage(endpoint *mocks.MockDiscoveryPeerEndpoint) *gprotoext.SignedGossipMessage {
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

func newStateInfoMessage(endpoint *mocks.MockDiscoveryPeerEndpoint) *gprotoext.SignedGossipMessage {
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

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/pkg/errors"
)

// MockDiscoveryServer is a mock Discovery server
type MockDiscoveryServer struct {
	localPeersByOrg map[string]*discovery.Peers
	peersByOrg      map[string]*discovery.Peers
}

// MockDiscoveryServerOpt is an option for the MockDiscoveryServer
type MockDiscoveryServerOpt func(s *MockDiscoveryServer)

// WithPeers adds a set of mock peers to the MockDiscoveryServer
func WithPeers(peers ...*MockDiscoveryPeerEndpoint) MockDiscoveryServerOpt {
	return func(s *MockDiscoveryServer) {
		s.peersByOrg = asPeersByOrg(peers)
	}
}

// WithLocalPeers adds a set of mock peers to the MockDiscoveryServer
func WithLocalPeers(peers ...*MockDiscoveryPeerEndpoint) MockDiscoveryServerOpt {
	return func(s *MockDiscoveryServer) {
		s.localPeersByOrg = asPeersByOrg(peers)
	}
}

// NewServer returns a new MockDiscoveryServer
func NewServer(opts ...MockDiscoveryServerOpt) *MockDiscoveryServer {
	s := &MockDiscoveryServer{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Discover Processes the given Discovery request and returns a mock response
func (s *MockDiscoveryServer) Discover(ctx context.Context, request *discovery.SignedRequest) (*discovery.Response, error) {
	if request == nil {
		return nil, errors.New("nil request")
	}

	req := &discovery.Request{}
	err := proto.Unmarshal(request.Payload, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing request")
	}
	if req.Authentication == nil {
		return nil, errors.New("access denied, no authentication info in request")
	}
	if len(req.Authentication.ClientIdentity) == 0 {
		return nil, errors.New("access denied, client identity wasn't supplied")
	}

	var results []*discovery.QueryResult
	for _, q := range req.Queries {
		result := s.processQuery(q)
		if result != nil {
			results = append(results, result)
		}
	}
	return &discovery.Response{
		Results: results,
	}, nil
}

func (s *MockDiscoveryServer) processQuery(q *discovery.Query) *discovery.QueryResult {
	if q.Channel == "" {
		if query := q.GetLocalPeers(); query != nil {
			return s.getLocalPeerQueryResult()
		}
	} else {
		if query := q.GetPeerQuery(); query != nil {
			return s.getPeerQueryResult()
		}
		if query := q.GetConfigQuery(); query != nil {
			return s.getConfigQueryResult()
		}
		if query := q.GetCcQuery(); query != nil {
			return s.getCCQueryResult()
		}
	}
	return nil
}

func (s *MockDiscoveryServer) getLocalPeerQueryResult() *discovery.QueryResult {
	if s.localPeersByOrg != nil {
		return &discovery.QueryResult{
			Result: &discovery.QueryResult_Members{
				Members: &discovery.PeerMembershipResult{
					PeersByOrg: s.localPeersByOrg,
				},
			},
		}
	}
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "no peers",
			},
		},
	}
}

func (s *MockDiscoveryServer) getPeerQueryResult() *discovery.QueryResult {
	if s.peersByOrg != nil {
		return &discovery.QueryResult{
			Result: &discovery.QueryResult_Members{
				Members: &discovery.PeerMembershipResult{
					PeersByOrg: s.peersByOrg,
				},
			},
		}
	}
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "no peers",
			},
		},
	}
}

func (s *MockDiscoveryServer) getConfigQueryResult() *discovery.QueryResult {
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "not implemented",
			},
		},
	}
}

func (s *MockDiscoveryServer) getCCQueryResult() *discovery.QueryResult {
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "not implemented",
			},
		},
	}
}

func asDiscoveryPeer(p *MockDiscoveryPeerEndpoint) *discovery.Peer {
	memInfoMsg := &gossip.GossipMessage{
		Content: &gossip.GossipMessage_AliveMsg{
			AliveMsg: &gossip.AliveMessage{
				Membership: &gossip.Member{
					Endpoint: p.Endpoint,
				},
				Timestamp: &gossip.PeerTime{
					SeqNum: uint64(1000),
					IncNum: uint64(time.Now().UnixNano()),
				},
			},
		},
	}
	memInfoPayload, err := proto.Marshal(memInfoMsg)
	if err != nil {
		panic(err.Error())
	}

	stateInfoMsg := &gossip.GossipMessage{
		Content: &gossip.GossipMessage_StateInfo{
			StateInfo: &gossip.StateInfo{
				Properties: &gossip.Properties{
					LedgerHeight: p.LedgerHeight,
					Chaincodes:   p.Chaincodes,
					LeftChannel:  p.LeftChannel,
				},
				Timestamp: &gossip.PeerTime{
					SeqNum: uint64(1000),
					IncNum: uint64(time.Now().UnixNano()),
				},
			},
		},
	}
	stateInfoPayload, err := proto.Marshal(stateInfoMsg)
	if err != nil {
		panic(err.Error())
	}

	return &discovery.Peer{
		MembershipInfo: &gossip.Envelope{
			Payload: memInfoPayload,
		},
		StateInfo: &gossip.Envelope{
			Payload: stateInfoPayload,
		},
	}
}

// MockDiscoveryPeerEndpoint contains information about a Discover peer endpoint
type MockDiscoveryPeerEndpoint struct {
	MSPID        string
	Endpoint     string
	LedgerHeight uint64
	Chaincodes   []*gossip.Chaincode
	LeftChannel  bool
}

func asPeersByOrg(peers []*MockDiscoveryPeerEndpoint) map[string]*discovery.Peers {
	peersByOrg := make(map[string]*discovery.Peers)
	for _, p := range peers {
		peers, ok := peersByOrg[p.MSPID]
		if !ok {
			peers = &discovery.Peers{}
			peersByOrg[p.MSPID] = peers
		}

		peers.Peers = append(peers.Peers, asDiscoveryPeer(p))
	}
	return peersByOrg
}

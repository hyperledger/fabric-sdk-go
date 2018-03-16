/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockDiscoveryProvider mocks out the discovery provider
type MockDiscoveryProvider struct {
	peers []fab.Peer
}

// NewDiscoveryProvider returns a new MockDiscoveryProvider
func NewDiscoveryProvider(peers ...fab.Peer) fab.DiscoveryProvider {
	return &MockDiscoveryProvider{
		peers: peers,
	}
}

// CreateDiscoveryService returns a new MockDiscoveryService
func (p *MockDiscoveryProvider) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {
	return &MockDiscoveryService{
		peers: p.peers,
	}, nil
}

// MockDiscoveryService is a mock discovery service used for event endpoint discovery
type MockDiscoveryService struct {
	peers []fab.Peer
}

// GetPeers returns a list of discovered peers
func (s *MockDiscoveryService) GetPeers() ([]fab.Peer, error) {
	return s.peers, nil
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

/**
 * Mock Discovery Provider is used to mock peers on the network
 */

// MockStaticDiscoveryProvider implements mock discovery provider
type MockStaticDiscoveryProvider struct {
	Error error
	Peers []fab.Peer
}

// MockStaticDiscoveryService implements mock discovery service
type MockStaticDiscoveryService struct {
	Error error
	Peers []fab.Peer
}

// NewMockDiscoveryProvider returns mock discovery provider
func NewMockDiscoveryProvider(err error, peers []fab.Peer) (*MockStaticDiscoveryProvider, error) {
	return &MockStaticDiscoveryProvider{Error: err, Peers: peers}, nil
}

// CreateLocalDiscoveryService return discovery service for specific channel
func (dp *MockStaticDiscoveryProvider) CreateLocalDiscoveryService(mspID string) (fab.DiscoveryService, error) {
	return &MockStaticDiscoveryService{Error: dp.Error, Peers: dp.Peers}, nil
}

// NewMockDiscoveryService returns mock discovery service
func NewMockDiscoveryService(err error, peers ...fab.Peer) *MockStaticDiscoveryService {
	return &MockStaticDiscoveryService{Error: err, Peers: peers}
}

// GetPeers is used to discover eligible peers for chaincode
func (ds *MockStaticDiscoveryService) GetPeers() ([]fab.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	if ds.Peers == nil {
		mockPeer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
		peers := make([]fab.Peer, 0)
		peers = append(peers, &mockPeer)
		ds.Peers = peers
	}

	return ds.Peers, nil

}

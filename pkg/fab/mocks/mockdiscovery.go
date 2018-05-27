/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

/**
 * Mock Discovery Provider is used to mock peers on the network
 */

// MockStaticDiscoveryProvider implements mock discovery provider
type MockStaticDiscoveryProvider struct {
	Error                  error
	Peers                  []fab.Peer
	customDiscoveryService fab.DiscoveryService
}

// MockStaticDiscoveryService implements mock discovery service
type MockStaticDiscoveryService struct {
	Error error
	Peers []fab.Peer
}

// NewMockDiscoveryProvider returns mock discovery provider
func NewMockDiscoveryProvider(err error, peers []fab.Peer) *MockStaticDiscoveryProvider {
	return &MockStaticDiscoveryProvider{Error: err, Peers: peers}
}

// CreateLocalDiscoveryService return local discovery service
func (dp *MockStaticDiscoveryProvider) CreateLocalDiscoveryService(mspID string) (fab.DiscoveryService, error) {

	if dp.customDiscoveryService != nil {
		return dp.customDiscoveryService, nil
	}

	return NewMockDiscoveryService(dp.Error, dp.Peers...), nil
}

//SetCustomDiscoveryService sets custom discoveryService
func (dp *MockStaticDiscoveryProvider) SetCustomDiscoveryService(customDiscoveryService fab.DiscoveryService) {
	dp.customDiscoveryService = customDiscoveryService
}

//NewMockDiscoveryService returns a new MockStaticDiscoveryService
func NewMockDiscoveryService(err error, peers ...fab.Peer) *MockStaticDiscoveryService {
	return &MockStaticDiscoveryService{Error: err, Peers: peers}
}

// GetPeers is used to discover eligible peers for chaincode
func (ds *MockStaticDiscoveryService) GetPeers() ([]fab.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	if ds.Peers == nil {
		mockPeer := MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
		peers := make([]fab.Peer, 0)
		peers = append(peers, &mockPeer)
		ds.Peers = peers
	}

	return ds.Peers, nil

}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockSelectionService implements mock selection service
type MockSelectionService struct {
	Error error
	Peers []fab.Peer
}

// NewMockSelectionService returns mock selection service
func NewMockSelectionService(err error, peers ...fab.Peer) *MockSelectionService {
	return &MockSelectionService{Error: err, Peers: peers}
}

// GetEndorsersForChaincode mockcore retrieving endorsing peers
func (ds *MockSelectionService) GetEndorsersForChaincode(chaincodes []*fab.ChaincodeCall, opts ...options.Opt) ([]fab.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	if ds.Peers == nil {
		mockPeer := NewMockPeer("Peer1", "http://peer1.com")
		peers := make([]fab.Peer, 0)
		peers = append(peers, mockPeer)
		ds.Peers = peers
	}

	return ds.Peers, nil

}

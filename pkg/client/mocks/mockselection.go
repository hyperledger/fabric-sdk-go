/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

// MockSelectionProvider implements mock selection provider
type MockSelectionProvider struct {
	Error error
	Peers []fab.Peer
}

// MockSelectionService implements mock selection service
type MockSelectionService struct {
	Error     error
	Peers     []fab.Peer
	SelectAll bool
}

// NewMockSelectionProvider returns mock selection provider
func NewMockSelectionProvider(err error, peers []fab.Peer) (*MockSelectionProvider, error) {
	return &MockSelectionProvider{Error: err, Peers: peers}, nil
}

// NewSelectionService returns mock selection service
func (dp *MockSelectionProvider) NewSelectionService(channelID string) (*MockSelectionService, error) {
	return &MockSelectionService{Error: dp.Error, Peers: dp.Peers}, nil
}

// GetEndorsersForChaincode mocks retrieving endorsing peers
func (ds *MockSelectionService) GetEndorsersForChaincode(channelPeers []fab.Peer,
	chaincodeIDs ...string) ([]fab.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	if ds.SelectAll {
		return channelPeers, nil
	}

	if ds.Peers == nil {
		mockPeer := mocks.NewMockPeer("Peer1", "http://peer1.com")
		peers := make([]fab.Peer, 0)
		peers = append(peers, mockPeer)
		ds.Peers = peers
	}

	return ds.Peers, nil

}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

// MockSelectionProvider implements mock selection provider
type MockSelectionProvider struct {
	Error error
	Peers []apifabclient.Peer
}

// MockSelectionService implements mock selection service
type MockSelectionService struct {
	Error error
	Peers []apifabclient.Peer
}

// NewMockSelectionProvider returns mock selection provider
func NewMockSelectionProvider(err error, peers []apifabclient.Peer) (*MockSelectionProvider, error) {
	return &MockSelectionProvider{Error: err, Peers: peers}, nil
}

// NewSelectionService returns mock selection service
func (dp *MockSelectionProvider) NewSelectionService(channelID string) (apifabclient.SelectionService, error) {
	return &MockSelectionService{Error: dp.Error, Peers: dp.Peers}, nil
}

// GetEndorsersForChaincode mocks retrieving endorsing peers
func (ds *MockSelectionService) GetEndorsersForChaincode(channelPeers []apifabclient.Peer,
	chaincodeIDs ...string) ([]apifabclient.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	if ds.Peers == nil {
		mockPeer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
		peers := make([]apifabclient.Peer, 0)
		peers = append(peers, &mockPeer)
		ds.Peers = peers
	}

	return ds.Peers, nil

}

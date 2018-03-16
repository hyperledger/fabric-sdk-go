/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

// MockSelectionProvider implements mock selection provider
type MockSelectionProvider struct {
	Error error
	Peers []fab.Peer
}

// MockSelectionService implements mock selection service
type MockSelectionService struct {
	Error          error
	Peers          []fab.Peer
	ChannelContext context.Channel
}

// NewMockSelectionProvider returns mock selection provider
func NewMockSelectionProvider(err error, peers []fab.Peer) (*MockSelectionProvider, error) {
	return &MockSelectionProvider{Error: err, Peers: peers}, nil
}

// CreateSelectionService returns mock selection service
func (dp *MockSelectionProvider) CreateSelectionService(channelID string) (*MockSelectionService, error) {
	return &MockSelectionService{Error: dp.Error, Peers: dp.Peers}, nil
}

// GetEndorsersForChaincode mocks retrieving endorsing peers
func (ds *MockSelectionService) GetEndorsersForChaincode(chaincodeIDs []string, opts ...options.Opt) ([]fab.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	params := selectopts.NewParams(opts)

	var peers []fab.Peer
	if ds.ChannelContext != nil {
		var err error
		peers, err = ds.ChannelContext.DiscoveryService().GetPeers()
		if err != nil {
			return nil, err
		}
	} else if ds.Peers == nil {
		mockPeer := mocks.NewMockPeer("Peer1", "http://peer1.com")
		peers = append(peers, mockPeer)
	}

	if params.PeerFilter != nil {
		for _, p := range ds.Peers {
			if params.PeerFilter(p) {
				peers = append(peers, p)
			}
		}
	} else {
		peers = ds.Peers
	}

	return peers, nil

}

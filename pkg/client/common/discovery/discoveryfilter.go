/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// filterService implements discovery service
type filterService struct {
	discoveryService fab.DiscoveryService
	targetFilter     fab.TargetFilter
}

// NewDiscoveryFilterService return discovery service with filter
func NewDiscoveryFilterService(discoveryService fab.DiscoveryService, targetFilter fab.TargetFilter) fab.DiscoveryService {
	return &filterService{discoveryService: discoveryService, targetFilter: targetFilter}
}

// GetPeers is used to get peers
func (fs *filterService) GetPeers() ([]fab.Peer, error) {
	peers, err := fs.discoveryService.GetPeers()
	if err != nil {
		return nil, err
	}
	targets := filterTargets(peers, fs.targetFilter)
	return targets, nil
}

// filterTargets is helper method to filter peers
func filterTargets(peers []fab.Peer, filter fab.TargetFilter) []fab.Peer {

	if filter == nil {
		return peers
	}

	filteredPeers := []fab.Peer{}
	for _, peer := range peers {
		if filter.Accept(peer) {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	return filteredPeers
}

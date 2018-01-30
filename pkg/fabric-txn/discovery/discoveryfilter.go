/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// filterService implements discovery service
type filterService struct {
	discoveryService apifabclient.DiscoveryService
	targetFilter     apifabclient.TargetFilter
}

// NewDiscoveryFilterService return discovery service with filter
func NewDiscoveryFilterService(discoveryService apifabclient.DiscoveryService, targetFilter apifabclient.TargetFilter) apifabclient.DiscoveryService {
	return &filterService{discoveryService: discoveryService, targetFilter: targetFilter}
}

// GetPeers is used to get peers
func (fs *filterService) GetPeers() ([]apifabclient.Peer, error) {
	peers, err := fs.discoveryService.GetPeers()
	if err != nil {
		return nil, err
	}
	targets := filterTargets(peers, fs.targetFilter)
	return targets, nil
}

// filterTargets is helper method to filter peers
func filterTargets(peers []apifabclient.Peer, filter apifabclient.TargetFilter) []apifabclient.Peer {

	if filter == nil {
		return peers
	}

	filteredPeers := []apifabclient.Peer{}
	for _, peer := range peers {
		if filter.Accept(peer) {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	return filteredPeers
}

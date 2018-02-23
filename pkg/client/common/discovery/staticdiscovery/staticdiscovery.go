/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/pkg/errors"
)

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config core.Config
}

// discoveryService implements discovery service
type discoveryService struct {
	config core.Config
	peers  []fab.Peer
}

// NewDiscoveryProvider returns discovery provider
func NewDiscoveryProvider(config core.Config) (*DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config}, nil
}

// NewDiscoveryService return discovery service for specific channel
func (dp *DiscoveryProvider) NewDiscoveryService(channelID string) (fab.DiscoveryService, error) {

	peers := []fab.Peer{}

	if channelID != "" {

		// Use configured channel peers
		chPeers, err := dp.config.ChannelPeers(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
		}

		for _, p := range chPeers {

			newPeer, err := peer.New(dp.config, peer.FromPeerConfig(&p.NetworkPeer))
			if err != nil || newPeer == nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}

			peers = append(peers, newPeer)
		}

	} else { // channel id is empty, return all configured peers

		netPeers, err := dp.config.NetworkPeers()
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for network peers")
		}

		for _, p := range netPeers {
			newPeer, err := peer.New(dp.config, peer.FromPeerConfig(&p))
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
			}

			peers = append(peers, newPeer)
		}
	}

	return &discoveryService{config: dp.config, peers: peers}, nil
}

// GetPeers is used to get peers
func (ds *discoveryService) GetPeers() ([]fab.Peer, error) {

	return ds.peers, nil
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config apiconfig.Config
}

// discoveryService implements discovery service
type discoveryService struct {
	config apiconfig.Config
	peers  []apifabclient.Peer
}

// NewDiscoveryProvider returns discovery provider
func NewDiscoveryProvider(config apiconfig.Config) (*DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config}, nil
}

// NewDiscoveryService return discovery service for specific channel
func (dp *DiscoveryProvider) NewDiscoveryService(channelID string) (apifabclient.DiscoveryService, error) {

	peers := []apifabclient.Peer{}

	if channelID != "" {

		// Use configured channel peers
		chPeers, err := dp.config.ChannelPeers(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
		}

		for _, p := range chPeers {

			newPeer, err := peer.NewPeerFromConfig(&p.NetworkPeer, dp.config)
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
			newPeer, err := peer.NewPeerFromConfig(&p, dp.config)
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
			}

			peers = append(peers, newPeer)
		}
	}

	return &discoveryService{config: dp.config, peers: peers}, nil
}

// GetPeers is used to get peers
func (ds *discoveryService) GetPeers() ([]apifabclient.Peer, error) {

	return ds.peers, nil
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/pkg/errors"
)

type peerCreator interface {
	CreatePeerFromConfig(peerCfg *fab.NetworkPeer) (fab.Peer, error)
}

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config  fab.EndpointConfig
	fabPvdr peerCreator
}

// New returns discovery provider
func New(config fab.EndpointConfig) (*DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config}, nil
}

// Initialize initializes the DiscoveryProvider
func (dp *DiscoveryProvider) Initialize(fabPvdr contextAPI.Providers) error {
	dp.fabPvdr = fabPvdr.InfraProvider()
	return nil
}

// CreateDiscoveryService return discovery service for specific channel
func (dp *DiscoveryProvider) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {
	if channelID == "" {
		return nil, errors.New("channel ID must be provided")
	}

	// Use configured channel peers
	chPeers, err := dp.config.ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	peers := []fab.Peer{}
	for _, p := range chPeers {
		newPeer, err := dp.fabPvdr.CreatePeerFromConfig(&p.NetworkPeer)
		if err != nil || newPeer == nil {
			return nil, errors.WithMessage(err, "NewPeer failed")
		}

		peers = append(peers, newPeer)
	}

	return &discoveryService{config: dp.config, peers: peers}, nil
}

// CreateLocalDiscoveryService return a local discovery service
func (dp *DiscoveryProvider) CreateLocalDiscoveryService() (fab.DiscoveryService, error) {
	peers := []fab.Peer{}

	netPeers, err := dp.config.NetworkPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for network peers")
	}

	for _, p := range netPeers {
		newPeer, err := dp.fabPvdr.CreatePeerFromConfig(&p)
		if err != nil {
			return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
		}

		peers = append(peers, newPeer)
	}

	return &localDiscoveryService{config: dp.config, peers: peers}, nil
}

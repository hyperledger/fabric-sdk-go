/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

type peerCreator interface {
	CreatePeerFromConfig(peerCfg *fab.NetworkPeer) (fab.Peer, error)
}

/**
 * Discovery Provider is used to discover peers on the network
 */

// LocalProvider implements discovery provider
type LocalProvider struct {
	config  fab.EndpointConfig
	fabPvdr peerCreator
}

// NewLocalProvider returns discovery provider
func NewLocalProvider(config fab.EndpointConfig) (*LocalProvider, error) {
	return &LocalProvider{config: config}, nil
}

// Initialize initializes the DiscoveryProvider
func (dp *LocalProvider) Initialize(fabPvdr contextAPI.Providers) error {
	dp.fabPvdr = fabPvdr.InfraProvider()
	return nil
}

// CreateLocalDiscoveryService return a local discovery service
func (dp *LocalProvider) CreateLocalDiscoveryService(mspID string) (fab.DiscoveryService, error) {
	peers := []fab.Peer{}
	netPeers := dp.config.NetworkPeers()

	logger.Debugf("Found %d peers", len(netPeers))

	for _, p := range netPeers {
		newPeer, err := dp.fabPvdr.CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: p.PeerConfig, MSPID: p.MSPID})
		if err != nil {
			return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
		}
		if newPeer.MSPID() == mspID {
			logger.Debugf("Adding local peer [%s] for MSP [%s]", newPeer.URL(), mspID)
			peers = append(peers, newPeer)
		}
	}

	return &localDiscoveryService{config: dp.config, peers: peers}, nil
}

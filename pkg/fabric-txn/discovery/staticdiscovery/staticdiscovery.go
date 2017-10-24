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

	peerConfig, err := dp.config.ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	peers := []apifabclient.Peer{}

	for _, p := range peerConfig {

		serverHostOverride := ""
		if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}

		newPeer, err := peer.NewPeerTLSFromCert(p.URL, p.TLSCACerts.Path, serverHostOverride, dp.config)
		if err != nil || newPeer == nil {
			return nil, errors.WithMessage(err, "NewPeer failed")
		}

		newPeer.SetMSPID(p.MspID)

		peers = append(peers, newPeer)
	}

	return &discoveryService{config: dp.config, peers: peers}, nil
}

// GetPeers is used to get peers (discovery service is channel based)
func (ds *discoveryService) GetPeers() ([]apifabclient.Peer, error) {

	return ds.peers, nil
}

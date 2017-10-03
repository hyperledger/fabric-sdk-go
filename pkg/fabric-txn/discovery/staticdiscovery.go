/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

/**
 * Discovery Provider is used to discover peers on the network
 */

// StaticDiscoveryProvider implements discovery provider
type StaticDiscoveryProvider struct {
	config apiconfig.Config
}

// StaticDiscoveryService implements discovery service
type StaticDiscoveryService struct {
	config  apiconfig.Config
	channel apifabclient.Channel
	peers   []apifabclient.Peer
}

// NewDiscoveryProvider returns discovery provider
func NewDiscoveryProvider(config apiconfig.Config) (*StaticDiscoveryProvider, error) {
	return &StaticDiscoveryProvider{config: config}, nil
}

// NewDiscoveryService return discovery service for specific channel
func (dp *StaticDiscoveryProvider) NewDiscoveryService(channel apifabclient.Channel) (apifabclient.DiscoveryService, error) {

	peerConfig, err := dp.config.ChannelPeers(channel.Name())
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	peers := []apifabclient.Peer{}

	for _, p := range peerConfig {

		serverHostOverride := ""
		if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}
		peer, err := peer.NewPeerTLSFromCert(p.URL, p.TLSCACerts.Path, serverHostOverride, dp.config)
		if err != nil {
			return nil, errors.WithMessage(err, "NewPeer failed")
		}
		peers = append(peers, peer)
	}

	return &StaticDiscoveryService{channel: channel, config: dp.config, peers: peers}, nil
}

// GetPeers is used to discover eligible peers for chaincode
func (ds *StaticDiscoveryService) GetPeers(chaincodeID string) ([]apifabclient.Peer, error) {
	// TODO: Incorporate CC policy here
	return ds.peers, nil
}
